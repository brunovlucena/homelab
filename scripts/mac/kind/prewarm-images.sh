#!/bin/bash
set -uo pipefail

# 🔥 Pre-warm local registry with required images
# This script pulls images from public registries and pushes them to the local Kind registry

REGISTRY_PORT="${REGISTRY_PORT:-5001}"
REGISTRY_NAME="kind-registry"
# NOTE: BuildKit containers do not resolve "localhost" reliably when running
# under Docker Desktop. Use the loopback IP explicitly to reach the local registry.
REGISTRY_HOST="127.0.0.1"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Track failed images
FAILED_IMAGES=()

# Auto-detect platform
ARCH=$(uname -m)
case "$ARCH" in
    x86_64)
        PLATFORM="linux/amd64"
        ;;
    arm64|aarch64)
        PLATFORM="linux/arm64"
        ;;
    *)
        echo "❌ Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔥 Pre-warming local registry with images"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "🖥️  Detected platform: $PLATFORM (arch: $ARCH)"
echo ""

# Check if local registry is running
if [ "$(docker inspect -f '{{.State.Running}}' "${REGISTRY_NAME}" 2>/dev/null || true)" != 'true' ]; then
    echo "❌ Local registry '${REGISTRY_NAME}' is not running!"
    echo "   Run: docker run -d --restart=always -p 127.0.0.1:${REGISTRY_PORT}:5000 --network bridge --name ${REGISTRY_NAME} registry:2"
    exit 1
fi

echo "✅ Local registry is running on ${REGISTRY_HOST}:${REGISTRY_PORT}"
echo ""

# Check if cluster is already running and detect missing images
if kind get clusters 2>/dev/null | grep -q "studio"; then
    echo "🔍 Detected running cluster - checking for missing images..."
    echo ""
    
    # Run detection script if available
    DETECT_SCRIPT="${SCRIPT_DIR}/detect-missing-images.sh"
    if [ -f "$DETECT_SCRIPT" ]; then
        if bash "$DETECT_SCRIPT"; then
            echo "✅ All images present. You can skip this script if desired."
            echo ""
            read -p "Continue with prewarm anyway? (y/N): " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                echo "Skipping prewarm."
                exit 0
            fi
            echo ""
        else
            echo "⚠️  Missing images detected. Proceeding with prewarm..."
            echo ""
        fi
    fi
fi

# Function to check if image exists in local registry (verifies manifest is actually pullable)
check_image_exists() {
    local repo=$1
    local tag_or_digest=$2
    local registry_url="http://${REGISTRY_HOST}:${REGISTRY_PORT}/v2/${repo}"
    
    # For digest SHA256, check manifest directly
    if [[ "$tag_or_digest" =~ ^sha256: ]]; then
        local digest="$tag_or_digest"
        # Verify manifest actually exists and is accessible
        if curl -sf -H "Accept: application/vnd.docker.distribution.manifest.v2+json" "${registry_url}/manifests/${digest}" >/dev/null 2>&1; then
            return 0
        fi
    else
        # For tags, verify the manifest is actually pullable (not just that tag exists in list)
        # This catches corrupted/broken manifests where tag exists but content is missing
        if curl -sf -H "Accept: application/vnd.docker.distribution.manifest.v2+json" "${registry_url}/manifests/${tag_or_digest}" >/dev/null 2>&1; then
            return 0
        fi
    fi
    
    return 1
}

# Function to pull, tag, and push an image (preserves multi-arch manifests)
prewarm_image() {
    local source_image=$1
    local target_name=${2:-$(basename $source_image)}
    local target_image_host="${REGISTRY_HOST}:${REGISTRY_PORT}/${target_name}"
    
    # Extract tag or digest from source image
    local tag_or_digest
    if [[ "$source_image" =~ @(sha256:[a-f0-9]{64})$ ]]; then
        # Image with digest
        tag_or_digest="${BASH_REMATCH[1]}"
    elif [[ "$source_image" =~ :([^@]+)$ ]]; then
        # Image with tag
        tag_or_digest="${BASH_REMATCH[1]}"
    else
        # Default to latest
        tag_or_digest="latest"
    fi
    
    # Extract repository name and the tag we push (from target_name).
    # We must verify using pushed_tag: we push as repo:pushed_tag, but tag_or_digest
    # comes from source (e.g. digest) and may differ — checking the wrong ref causes
    # "buildx imagetools reported success but image not found in registry".
    local repo_name="${target_name%%:*}"
    local pushed_tag
    if [[ "$target_name" == *:* ]]; then
        pushed_tag="${target_name#*:}"
    else
        pushed_tag="latest"
    fi
    
    echo "📦 Processing: $source_image"
    echo "   → ${target_image_host}"
    
    # Check if image already exists in local registry (idempotent check)
    if check_image_exists "$repo_name" "$pushed_tag"; then
        # Verify the image is actually pullable (not corrupted)
        echo "   🔍 Verifying image is not corrupted..."
        if docker pull "$target_image_host" >/dev/null 2>&1; then
            echo "   ℹ️  Image already exists in local registry and is valid, skipping"
            echo ""
            return 0
        else
            echo "   ⚠️  Image exists but is CORRUPTED - deleting and re-prewarming..."
            # Delete corrupted manifest/tag
            curl -sf -X DELETE "http://${REGISTRY_HOST}:${REGISTRY_PORT}/v2/${repo_name}/manifests/${pushed_tag}" >/dev/null 2>&1 || true
            # Also try to delete by digest if we can get it
            local manifest_digest=$(curl -sf -I -H "Accept: application/vnd.docker.distribution.manifest.v2+json" "http://${REGISTRY_HOST}:${REGISTRY_PORT}/v2/${repo_name}/manifests/${pushed_tag}" 2>/dev/null | grep -i "Docker-Content-Digest:" | cut -d' ' -f2 | tr -d '\r')
            if [ -n "$manifest_digest" ]; then
                curl -sf -X DELETE "http://${REGISTRY_HOST}:${REGISTRY_PORT}/v2/${repo_name}/manifests/${manifest_digest}" >/dev/null 2>&1 || true
            fi
            echo "   🔄 Re-prewarming corrupted image..."
            # Continue to prewarm below
        fi
    fi
    
    # Check if image already exists locally in Docker cache
    if docker image inspect "$source_image" >/dev/null 2>&1; then
        echo "   ℹ️  Image already exists locally, using cached version..."
        local tag_output
        tag_output=$(docker tag "$source_image" "$target_image_host" 2>&1)
        local tag_exit=$?
        
        if [ $tag_exit -ne 0 ]; then
            echo "   ❌ FAILED to tag image:"
            echo "      $(echo "$tag_output" | head -2 | sed 's/^/      /')"
            FAILED_IMAGES+=("$target_image_host (tag failed)")
            echo ""
            return 1
        fi
        
        local push_output
        push_output=$(docker push "$target_image_host" 2>&1)
        local push_exit=$?
        
        if [ $push_exit -eq 0 ]; then
            # Verify it was actually pushed and is pullable
            if check_image_exists "$repo_name" "$pushed_tag"; then
                # Final verification: actually try to pull it
                if docker pull "$target_image_host" >/dev/null 2>&1; then
                    echo "   ✅ Done (using local cache)"
                    echo ""
                    return 0
                else
                    echo "   ❌ FAILED: Image exists but is not pullable (corrupted)"
                    FAILED_IMAGES+=("$target_image_host (corrupted after push)")
                    echo ""
                    return 1
                fi
            else
                echo "   ❌ FAILED: Image push reported success but not found in registry"
                FAILED_IMAGES+=("$target_image_host (verification failed)")
                echo ""
                return 1
            fi
        else
            echo "   ❌ FAILED to push image to registry:"
            echo "      $(echo "$push_output" | grep -i "error\|failed" | head -2 | sed 's/^/      /' || echo "$push_output" | tail -2 | sed 's/^/      /')"
            FAILED_IMAGES+=("$target_image_host (push failed)")
            echo ""
            return 1
        fi
    fi
    
    # Try docker buildx imagetools first (preserves multi-arch manifest)
    # Only use buildx if image is already pulled locally (avoids rate limits)
    if docker image inspect "$source_image" >/dev/null 2>&1; then
        echo "   🔧 Using buildx imagetools to preserve multi-arch manifest (image already local)..."
        local buildx_output
        buildx_output=$(docker buildx imagetools create --tag "$target_image_host" "$source_image" 2>&1)
        local buildx_exit=$?
        
        if [ $buildx_exit -eq 0 ]; then
            # Verify it was actually created and is pullable
            if check_image_exists "$repo_name" "$pushed_tag"; then
                # Final verification: actually try to pull it
                if docker pull "$target_image_host" >/dev/null 2>&1; then
                    echo "   ✅ Done (multi-arch manifest preserved)"
                    echo ""
                    return 0
                else
                    echo "   ❌ FAILED: Image exists but is not pullable (corrupted)"
                    FAILED_IMAGES+=("$target_image_host (buildx corrupted)")
                    echo ""
                    return 1
                fi
            else
                echo "   ❌ FAILED: buildx imagetools reported success but image not found in registry"
                FAILED_IMAGES+=("$target_image_host (buildx verification failed)")
                echo ""
                return 1
            fi
        else
            echo "   ⚠️  buildx imagetools failed:"
            echo "      $(echo "$buildx_output" | head -3 | sed 's/^/      /')"
            echo "   🔄 Falling back to docker pull/tag/push..."
        fi
    else
        echo "   ℹ️  Image not in local cache, will pull first then use buildx..."
    fi
    
    # Fallback: traditional docker pull/tag/push (loses multi-arch manifest)
    # Only pull if image doesn't exist locally
    if ! docker image inspect "$source_image" >/dev/null 2>&1; then
        echo "   🔄 Pulling image from source..."
        local pull_output
        pull_output=$(docker pull "$source_image" 2>&1)
        local pull_exit=$?
        
        if [ $pull_exit -ne 0 ]; then
            echo "   ❌ FAILED to pull image:"
            echo "      $(echo "$pull_output" | grep -i "error\|failed\|rate limit" | head -2 | sed 's/^/      /' || echo "$pull_output" | tail -2 | sed 's/^/      /')"
            if echo "$pull_output" | grep -qi "rate limit\|429\|too many requests"; then
                echo "   💡 Rate limited! Try: docker login"
            fi
            FAILED_IMAGES+=("$target_image_host (pull failed)")
            echo ""
            return 1
        fi
        
        # Now that image is local, try buildx again
        echo "   🔧 Retrying buildx imagetools now that image is local..."
        local buildx_output
        buildx_output=$(docker buildx imagetools create --tag "$target_image_host" "$source_image" 2>&1)
        local buildx_exit=$?
        
        if [ $buildx_exit -eq 0 ]; then
            if check_image_exists "$repo_name" "$pushed_tag"; then
                if docker pull "$target_image_host" >/dev/null 2>&1; then
                    echo "   ✅ Done (multi-arch manifest preserved after pull)"
                    echo ""
                    return 0
                fi
            fi
        fi
        echo "   ⚠️  buildx still failed, using traditional push method..."
    fi
    
    local tag_output
    tag_output=$(docker tag "$source_image" "$target_image_host" 2>&1)
    local tag_exit=$?
    
    if [ $tag_exit -ne 0 ]; then
        echo "   ❌ FAILED to tag image:"
        echo "      $(echo "$tag_output" | head -2 | sed 's/^/      /')"
        FAILED_IMAGES+=("$target_image_host (tag failed)")
        echo ""
        return 1
    fi
    
    local push_output
    push_output=$(docker push "$target_image_host" 2>&1)
    local push_exit=$?
    
    if [ $push_exit -eq 0 ]; then
        # Verify it was actually pushed and is pullable
        if check_image_exists "$repo_name" "$pushed_tag"; then
            # Final verification: actually try to pull it
            if docker pull "$target_image_host" >/dev/null 2>&1; then
                echo "   ✅ Done (single-arch, multi-arch manifest lost)"
                echo ""
                return 0
            else
                echo "   ❌ FAILED: Image exists but is not pullable (corrupted)"
                FAILED_IMAGES+=("$target_image_host (corrupted after push)")
                echo ""
                return 1
            fi
        else
            echo "   ❌ FAILED: Image push reported success but not found in registry"
            FAILED_IMAGES+=("$target_image_host (verification failed)")
            echo ""
            return 1
        fi
    else
        echo "   ❌ FAILED to push image to registry:"
        echo "      $(echo "$push_output" | grep -i "error\|failed" | head -2 | sed 's/^/      /' || echo "$push_output" | tail -2 | sed 's/^/      /')"
        FAILED_IMAGES+=("$target_image_host (push failed)")
        echo ""
        return 1
    fi
}

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 🚀 GitHub Container Registry (High Priority)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
prewarm_image "docker.io/kindest/node:v1.34.0" "kindest-node:v1.34.0"
prewarm_image "ghcr.io/fluxcd/flux-cli:v2.4.0" "flux-cli:v2.4.0"
prewarm_image "ghcr.io/actions/gha-runner-scale-set-controller:0.13.0" "gha-runner-scale-set-controller:0.13.0"
prewarm_image "ghcr.io/actions/actions-runner:v2.329.0" "actions-runner:v2.329.0"
prewarm_image "ghcr.io/fluxcd/source-controller:v1.4.1" "source-controller:v1.4.1"
prewarm_image "ghcr.io/fluxcd/kustomize-controller:v1.4.0" "kustomize-controller:v1.4.0"
prewarm_image "ghcr.io/fluxcd/helm-controller:v1.1.0" "helm-controller:v1.1.0"
prewarm_image "ghcr.io/fluxcd/notification-controller:v1.4.0" "notification-controller:v1.4.0"
prewarm_image "ghcr.io/fluxcd/flagger:1.39.0" "flagger:1.39.0"
prewarm_image "ghcr.io/external-secrets/external-secrets:v0.11.0@sha256:776b383deb6f793c7161bee75d4e53e0d8bdaabbd4e5c3346929a83a35b46ed2" "external-secrets:v0.11.0"
prewarm_image "ghcr.io/telepresenceio/tel2:2.25.0" "datawire/tel2:2.25.0"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Kubernetes Core Components
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# prewarm_image "registry.k8s.io/coredns/coredns:v1.12.1" "coredns:v1.12.1"
# prewarm_image "registry.k8s.io/etcd:3.6.4-0" "etcd:3.6.4-0"
# prewarm_image "registry.k8s.io/kube-apiserver:v1.34.0" "kube-apiserver:v1.34.0"
# prewarm_image "registry.k8s.io/kube-controller-manager:v1.34.0" "kube-controller-manager:v1.34.0"
# prewarm_image "registry.k8s.io/kube-proxy:v1.34.0" "kube-proxy:v1.34.0"
# prewarm_image "registry.k8s.io/kube-scheduler:v1.34.0" "kube-scheduler:v1.34.0"
# prewarm_image "gcr.io/google_containers/pause:3.2" "pause:3.2"
prewarm_image "registry.k8s.io/kube-state-metrics/kube-state-metrics:v2.17.0" "kube-state-metrics:v2.17.0"
prewarm_image "registry.k8s.io/metrics-server/metrics-server:v0.7.2" "metrics-server:v0.7.2"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Linkerd Service Mesh
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
prewarm_image "cr.l5d.io/linkerd/controller:edge-25.11.1" "linkerd-controller:edge-25.11.1"
prewarm_image "cr.l5d.io/linkerd/controller:edge-25.11.1" "controller:edge-25.11.1"
prewarm_image "cr.l5d.io/linkerd/metrics-api:edge-25.11.1" "linkerd-metrics-api:edge-25.11.1"
prewarm_image "cr.l5d.io/linkerd/proxy:edge-25.11.1" "linkerd-proxy:edge-25.11.1"
prewarm_image "cr.l5d.io/linkerd/proxy-init:v2.4.3" "linkerd-proxy-init:v2.4.3"
prewarm_image "cr.l5d.io/linkerd/smi-adaptor:v0.2.7" "linkerd-smi-adaptor:v0.2.7"
prewarm_image "cr.l5d.io/linkerd/tap:edge-25.11.1" "linkerd-tap:edge-25.11.1"
prewarm_image "cr.l5d.io/linkerd/web:edge-25.11.1" "linkerd-web:edge-25.11.1"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Cert Manager
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
prewarm_image "quay.io/jetstack/cert-manager-cainjector:v1.18.2" "cert-manager-cainjector:v1.18.2"
prewarm_image "quay.io/jetstack/cert-manager-controller:v1.18.2" "cert-manager-controller:v1.18.2"
prewarm_image "quay.io/jetstack/cert-manager-webhook:v1.18.2" "cert-manager-webhook:v1.18.2"
prewarm_image "quay.io/jetstack/cert-manager-startupapicheck:v1.18.2" "cert-manager-startupapicheck:v1.18.2"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Sealed Secrets
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
prewarm_image "bitnami/sealed-secrets-controller:0.33.1" "sealed-secrets-controller:0.33.1"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Grafana Stack
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
prewarm_image "grafana/alloy:v1.7.5" "grafana-alloy:v1.7.5"
prewarm_image "grafana/grafana:12.2.1" "grafana:12.2.1"
prewarm_image "grafana/tempo:2.8.2" "tempo:2.8.2"
prewarm_image "grafana/loki:3.5.7" "loki:3.5.7"
prewarm_image "grafana/loki-canary:3.5.7" "loki-canary:3.5.7"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Prometheus Stack
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
prewarm_image "prom/prometheus:v3.7.3" "prometheus:v3.7.3"
prewarm_image "quay.io/prometheus/alertmanager:v0.29.0" "alertmanager:v0.29.0"
prewarm_image "quay.io/prometheus/node-exporter:v1.10.2" "node-exporter:v1.10.2"
prewarm_image "quay.io/prometheus-operator/prometheus-operator:v0.86.2" "prometheus-operator:v0.86.2"
prewarm_image "quay.io/prometheus-operator/prometheus-config-reloader:v0.86.2" "prometheus-config-reloader:v0.86.2"
prewarm_image "quay.io/prometheus-operator/prometheus-config-reloader:v0.81.0" "prometheus-config-reloader:v0.81.0"
prewarm_image "registry.k8s.io/ingress-nginx/kube-webhook-certgen:v1.6.4" "kube-webhook-certgen:v1.6.4"
prewarm_image "prom/pushgateway:v1.11.0" "pushgateway:v1.11.0"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Sidecars & Utilities
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
prewarm_image "kiwigrid/k8s-sidecar:1.30.10" "k8s-sidecar:1.30.10"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Knative Serverless Platform
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
prewarm_image "gcr.io/knative-releases/knative.dev/operator/cmd/operator:v1.16.3" "knative-operator:v1.16.3"
prewarm_image "gcr.io/knative-releases/knative.dev/operator/cmd/webhook:v1.16.3" "knative-operator-webhook:v1.16.3"
prewarm_image "gcr.io/knative-releases/knative.dev/serving/cmd/activator:v1.16.3" "knative-serving-activator:v1.16.3"
prewarm_image "gcr.io/knative-releases/knative.dev/serving/cmd/autoscaler:v1.16.3" "knative-serving-autoscaler:v1.16.3"
prewarm_image "gcr.io/knative-releases/knative.dev/serving/cmd/autoscaler-hpa:v1.16.3" "knative-serving-autoscaler-hpa:v1.16.3"
prewarm_image "gcr.io/knative-releases/knative.dev/serving/cmd/controller:v1.16.3" "knative-serving-controller:v1.16.3"
prewarm_image "gcr.io/knative-releases/knative.dev/serving/cmd/webhook:v1.16.3" "knative-serving-webhook:v1.16.3"
prewarm_image "gcr.io/knative-releases/knative.dev/serving/cmd/default-domain:v1.16.3" "knative-serving-cleanup:v1.16.3"
prewarm_image "gcr.io/knative-releases/knative.dev/eventing/cmd/controller:v1.16.5" "knative-eventing-controller:v1.16.5"
prewarm_image "gcr.io/knative-releases/knative.dev/eventing/cmd/webhook:v1.16.5" "knative-eventing-webhook:v1.16.5"
prewarm_image "gcr.io/knative-releases/knative.dev/eventing/cmd/in_memory/channel_controller:v1.16.5" "knative-eventing-imc-controller:v1.16.5"
prewarm_image "gcr.io/knative-releases/knative.dev/eventing/cmd/in_memory/channel_dispatcher:v1.16.5" "knative-eventing-imc-dispatcher:v1.16.5"
prewarm_image "gcr.io/knative-releases/knative.dev/eventing/cmd/broker/ingress:v1.16.5" "knative-eventing-mt-broker-ingress:v1.16.5"
prewarm_image "gcr.io/knative-releases/knative.dev/eventing/cmd/broker/filter:v1.16.5" "knative-eventing-mt-broker-filter:v1.16.5"
prewarm_image "gcr.io/knative-releases/knative.dev/eventing/cmd/mtchannel_broker:v1.16.5" "knative-eventing-mt-broker-controller:v1.16.5"
prewarm_image "gcr.io/knative-releases/knative.dev/eventing/cmd/jobsink:v1.16.5" "knative-eventing-job-sink:v1.16.5"
prewarm_image "gcr.io/knative-releases/knative.dev/net-kourier/cmd/kourier:v1.16.0" "knative-net-kourier-controller:v1.16.0"
prewarm_image "docker.io/envoyproxy/envoy:v1.31-latest" "envoyproxy-envoy:v1.31-latest"
prewarm_image "gcr.io/knative-releases/knative.dev/pkg/apiextensions/storageversion/cmd/migrate:v1.16.3" "knative-migrate:v1.16.3"
prewarm_image "gcr.io/knative-releases/knative.dev/pkg/apiextensions/storageversion/cmd/migrate@sha256:1145ac9e94eaf4a04b1f26d3f87dd9afd0c524a8e0382dbf38eeffb4c419513b" "knative-migrate-serving:sha256-1145ac9e"
prewarm_image "gcr.io/knative-releases/knative.dev/pkg/apiextensions/storageversion/cmd/migrate@sha256:c6baaa8b2a882ff9269ebfe14b54dcba29efc060dfafc06971001b89b084667b" "knative-migrate-eventing:sha256-c6baaa8b"
prewarm_image "gcr.io/knative-releases/knative.dev/serving/cmd/default-domain:v1.16.3" "knative-serving-default-domain:v1.16.3"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# RabbitMQ
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
prewarm_image "rabbitmq:3.12-management-alpine" "rabbitmq:3.12-management-alpine"
prewarm_image "rabbitmqoperator/cluster-operator:2.17.2" "rabbitmq-cluster-operator:2.17.2"
prewarm_image "rabbitmqoperator/messaging-topology-operator:1.17.0" "messaging-topology-operator:1.17.0"

# RabbitMQ Operator - Knative Eventing Images
prewarm_image "gcr.io/knative-releases/knative.dev/eventing-rabbitmq/cmd/controller/source:v1.17.6" "knative-rabbitmq-source-controller:v1.17.6"
prewarm_image "gcr.io/knative-releases/knative.dev/eventing-rabbitmq/cmd/receive_adapter:v1.17.6" "knative-rabbitmq-receive-adapter:v1.17.6"
prewarm_image "gcr.io/knative-releases/knative.dev/eventing-rabbitmq/cmd/webhook/source:v1.17.6" "knative-rabbitmq-source-webhook:v1.17.6"
prewarm_image "gcr.io/knative-releases/knative.dev/eventing-rabbitmq/cmd/controller/broker:v1.17.6" "knative-rabbitmq-broker-controller:v1.17.6"
prewarm_image "gcr.io/knative-releases/knative.dev/eventing-rabbitmq/cmd/ingress:v1.17.6" "knative-rabbitmq-broker-ingress:v1.17.6"
prewarm_image "gcr.io/knative-releases/knative.dev/eventing-rabbitmq/cmd/dispatcher:v1.17.6" "knative-rabbitmq-dispatcher:v1.17.6"
prewarm_image "gcr.io/knative-releases/knative.dev/eventing-rabbitmq/cmd/webhook/broker:v1.17.6" "knative-rabbitmq-broker-webhook:v1.17.6"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Knative Lambda
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Knative Lambda - Base images for Kaniko builds
prewarm_image "docker.io/library/node:22-alpine" "node:22-alpine"
prewarm_image "docker.io/library/node:20-alpine" "node:20-alpine"
prewarm_image "docker.io/library/python:3.11-alpine" "python:3.11-alpine"
prewarm_image "docker.io/library/python:3.11-slim" "python:3.11-slim"
prewarm_image "docker.io/library/golang:1.25-alpine" "golang:1.25-alpine"
prewarm_image "docker.io/library/alpine:3.19" "alpine:3.19"
prewarm_image "docker.io/alpine/helm:3.16.4" "alpine-helm:3.16.4"
prewarm_image "gcr.io/kaniko-project/executor:v1.19.2" "kaniko-executor:v1.19.2"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Additional Application & Example Images
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
prewarm_image "pulumi/pulumi-kubernetes-operator:v1.15.0" "pulumi-kubernetes-operator:v1.15.0"
prewarm_image "docker.io/pulumi/pulumi-kubernetes-operator:v2.3.0" "pulumi-kubernetes-operator:v2.3.0"

prewarm_image "docker.io/library/redis:7.2-alpine" "redis:7.2-alpine"
prewarm_image "docker.io/library/postgres:16-alpine" "postgres:16-alpine"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# MinIO Object Storage (for Knative Lambda sources)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
prewarm_image "minio/minio:RELEASE.2025-09-07T16-13-09Z-cpuv1" "minio:RELEASE.2025-09-07T16-13-09Z-cpuv1"
prewarm_image "minio/mc:RELEASE.2025-08-13T08-35-41Z-cpuv1" "mc:RELEASE.2025-08-13T08-35-41Z-cpuv1"
prewarm_image "busybox:1.36" "busybox:1.36"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# K6 Load Testing
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
prewarm_image "grafana/k6:1.0.0" "k6:1.0.0"
prewarm_image "grafana/k6:0.47.0" "k6:0.47.0"
prewarm_image "ghcr.io/grafana/k6-operator:0.0.19-starter" "k6-operator:0.0.19-starter"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Ollama LLM
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
prewarm_image "ollama/ollama:0.13.1" "ollama:0.13.1"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Flyte Workflow Platform
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
prewarm_image "cr.flyte.org/flyteorg/datacatalog-release:v1.16.3" "flyteorg-datacatalog-release:v1.16.3"
prewarm_image "cr.flyte.org/flyteorg/flyteadmin-release:v1.16.3" "flyteorg-flyteadmin-release:v1.16.3"
prewarm_image "cr.flyte.org/flyteorg/flyteconsole-release:v1.16.3" "flyteorg-flyteconsole-release:v1.16.3"
prewarm_image "cr.flyte.org/flyteorg/flytepropeller-release:v1.16.3" "flyteorg-flytepropeller-release:v1.16.3"
prewarm_image "ghcr.io/brunovlucena/flyte-sandbox-training:latest" "flyte-sandbox-training:latest"
prewarm_image "ghcr.io/brunovlucena/flyte-workflow-registry:latest" "flyte-workflow-registry:latest"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Kubernetes Utilities (for lambda-samples-init and k6 tests)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
prewarm_image "bitnami/kubectl:1.34.0" "kubectl:v1.34.0"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Custom Applications (brunovlucena)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
prewarm_image "ghcr.io/brunovlucena/agent-bruno:v1.5.1" "agent-bruno:v1.5.1"
prewarm_image "ghcr.io/brunovlucena/agent-sre:latest" "agent-sre:latest"
prewarm_image "ghcr.io/brunovlucena/ai_seller:0.3.0" "ai_seller:0.3.0"
prewarm_image "ghcr.io/brunovlucena/command-center:v1.2.0" "command-center:v1.2.0"
prewarm_image "ghcr.io/brunovlucena/agent-chat/command-center:1.2.0" "agent-chat/command-center:1.2.0"
prewarm_image "ghcr.io/brunovlucena/homepage-api:v0.1.37" "homepage-api:v0.1.37"
prewarm_image "ghcr.io/brunovlucena/homepage-frontend:v0.1.37" "homepage-frontend:v0.1.37"
prewarm_image "ghcr.io/brunovlucena/knative-lambda-operator:v1.13.10" "knative-lambda-operator:v1.13.10"
prewarm_image "ghcr.io/brunovlucena/prometheus-events:v0.1.0" "prometheus-events:v0.1.0"
prewarm_image "ghcr.io/brunovlucena/location-agent:v1.2.0" "location-agent:v1.2.0"
prewarm_image "ghcr.io/brunovlucena/media-agent:v1.2.0" "media-agent:v1.2.0"
prewarm_image "ghcr.io/brunovlucena/messaging-hub:v1.2.0" "messaging-hub:v1.2.0"
prewarm_image "ghcr.io/brunovlucena/order_processor:0.3.0" "order_processor:0.3.0"
prewarm_image "ghcr.io/brunovlucena/product_catalog:0.3.0" "product_catalog:0.3.0"
prewarm_image "ghcr.io/brunovlucena/sales_assistant:0.3.0" "sales_assistant:0.3.0"
prewarm_image "ghcr.io/brunovlucena/voice-agent:v1.2.0" "voice-agent:v1.2.0"
prewarm_image "ghcr.io/brunovlucena/whatsapp_gateway:0.3.0" "whatsapp_gateway:0.3.0"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Infrastructure Services
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
prewarm_image "cloudflare/cloudflared:2025.11.1" "cloudflared:2025.11.1"
prewarm_image "pihole/pihole:2025.11.0" "pihole:2025.11.0"
prewarm_image "mitmproxy/mitmproxy:11.1.0" "mitmproxy:11.1.0"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Report results
if [ ${#FAILED_IMAGES[@]} -eq 0 ]; then
    echo "✅ All images pre-warmed successfully!"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "📦 Total images: 94 production images"
    echo ""
    echo "Images available in localhost:${REGISTRY_PORT}:"
    echo "  • Kubernetes: coredns, etcd, kube-apiserver, kube-controller-manager, kube-proxy, kube-scheduler, kube-state-metrics, metrics-server"
    echo "  • Linkerd: controller, metrics-api, proxy, proxy-init, smi-adaptor, tap, web"
    echo "  • Flux: flux-cli, helm-controller, kustomize-controller, notification-controller, source-controller, flagger"
    echo "  • Cert Manager: cainjector, controller, webhook, startupapicheck"
    echo "  • Sealed Secrets: controller"
    echo "  • External Secrets: external-secrets"
    echo "  • Grafana: alloy, grafana, loki, loki-canary, tempo, prometheus"
    echo "  • Prometheus: prometheus, alertmanager, node-exporter, operator, config-reloader"
    echo "  • Knative: operator, serving, eventing, kourier, envoy"
    echo "  • Utilities: k8s-sidecar"
    echo "  • GitHub Actions: gha-runner-scale-set-controller, actions-runner"
    echo "  • RabbitMQ: rabbitmq, messaging-topology-operator, knative-eventing"
    echo "  • MinIO: minio, mc"
    echo "  • Knative Lambda: builder, sidecar, metrics-pusher, base images, kaniko"
    echo "  • Testing: k6"
    echo "  • Flyte: datacatalog, flyteadmin, flyteconsole, flytepropeller, flyte-sandbox-training, flyte-workflow-registry"
    echo "  • Custom Apps: agent-bruno, agent-sre, ai_seller, command-center, homepage-api, homepage-frontend, knative-lambda-operator, prometheus-events, location-agent, media-agent, messaging-hub, order_processor, product_catalog, sales_assistant, voice-agent, whatsapp_gateway"
    echo "  • Database: postgres:16-alpine, redis:7.2-alpine"
    echo ""
    exit 0
else
    echo "❌ PREWARM FAILED - ${#FAILED_IMAGES[@]} image(s) failed to prewarm!"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "Failed images:"
    for failed in "${FAILED_IMAGES[@]}"; do
        echo "  ❌ $failed"
    done
    echo ""
    echo "💡 Fix the errors above and run 'make prewarm-images' again"
    echo ""
    exit 1
fi
