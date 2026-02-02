#!/bin/bash
set -euo pipefail

# 🚀 Bootstrap Kind Cluster with Flux
# Usage: ./bootstrap-cluster.sh <cluster-name> <kind-config-path> <kube-context>

CLUSTER_NAME="${1:-}"
KIND_CONFIG="${2:-}"
KUBE_CONTEXT="${3:-}"

if [ -z "$CLUSTER_NAME" ] || [ -z "$KIND_CONFIG" ] || [ -z "$KUBE_CONTEXT" ]; then
    echo "❌ Usage: $0 <cluster-name> <kind-config-path> <kube-context>"
    exit 1
fi

nodes_ready() {
    local output
    if ! output=$(kubectl get nodes --context "$KUBE_CONTEXT" --no-headers 2>/dev/null) || [ -z "$output" ]; then
        return 1
    fi
    echo "$output" | awk 'NF && $2 != "Ready" {exit 1}' || return 1
    return 0
}

wait_for_nodes() {
    if nodes_ready; then
        echo "✅ Cluster nodes already ready"
        return
    fi
    kubectl wait --for=condition=ready nodes --all --timeout=300s --context "$KUBE_CONTEXT"
}

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🚀 Bootstrapping $CLUSTER_NAME cluster"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 1. Create local registry if it doesn't exist
echo "📦 Setting up local registry..."
if [ "$(docker inspect -f '{{.State.Running}}' kind-registry 2>/dev/null || true)" != 'true' ]; then
    docker run -d --restart=always -p "127.0.0.1:5001:5000" --network bridge --name kind-registry registry:2
    echo "✅ Registry created"
else
    echo "✅ Registry already running"
fi

# 2. Create Kind cluster (idempotent)
echo "🔧 Creating Kind cluster..."
if kind get clusters | grep -qw "$CLUSTER_NAME"; then
    # Check if cluster is actually running by trying to export kubeconfig
    if kind export kubeconfig --name "$CLUSTER_NAME" >/dev/null 2>&1; then
        echo "✅ Cluster already exists and is running"
    else
        echo "⚠️  Cluster exists but is not running, recreating..."
        kind delete cluster --name "$CLUSTER_NAME" 2>/dev/null || true
        kind create cluster --name "$CLUSTER_NAME" --config "$KIND_CONFIG"
    fi
else
    kind create cluster --name "$CLUSTER_NAME" --config "$KIND_CONFIG"
fi

# 2b. Ensure kube context alias exists for downstream tooling
echo "🧭 Ensuring kube context '$KUBE_CONTEXT' exists..."
if kubectl config get-contexts "$KUBE_CONTEXT" >/dev/null 2>&1; then
    echo "   • Context already configured"
else
    if ! kubectl config get-contexts "kind-$CLUSTER_NAME" >/dev/null 2>&1; then
        echo "   • Base context kind-$CLUSTER_NAME not found, exporting kubeconfig..."
        kind export kubeconfig --name "$CLUSTER_NAME" 2>/dev/null || {
            echo "   ⚠️  Failed to export kubeconfig, cluster may not be running"
            echo "   • Attempting to recreate cluster..."
            kind delete cluster --name "$CLUSTER_NAME" 2>/dev/null || true
            kind create cluster --name "$CLUSTER_NAME" --config "$KIND_CONFIG"
        }
    fi
    kubectl config set-context "$KUBE_CONTEXT" --cluster="kind-$CLUSTER_NAME" --user="kind-$CLUSTER_NAME"
    echo "   • Created kube context alias"
fi

# 3. Configure nodes to use local registry (always ensure all nodes are configured)
echo "🔗 Configuring registry for cluster nodes..."
REGISTRY_DIR="/etc/containerd/certs.d/localhost:5001"
wait_for_nodes
for node in $(kind get nodes --name "$CLUSTER_NAME"); do
    if docker exec "$node" test -f "${REGISTRY_DIR}/hosts.toml" 2>/dev/null; then
        echo "   • $node already configured"
    else
        docker exec "$node" mkdir -p "${REGISTRY_DIR}" 2>/dev/null || true
        docker exec "$node" sh -c "cat > ${REGISTRY_DIR}/hosts.toml <<EOF
[host.\"http://kind-registry:5000\"]
EOF" 2>/dev/null || true
        echo "   • Configured $node"
    fi
done

# 4. Connect registry to Kind network
echo "🌐 Ensuring registry is on kind network..."
if [ "$(docker inspect -f='{{json .NetworkSettings.Networks.kind}}' kind-registry)" = 'null' ]; then
    docker network connect kind kind-registry || true
    echo "   • Connected registry to kind network"
else
    echo "   • Registry already on kind network"
fi

# 5. Wait for cluster to be ready
echo "⏳ Waiting for cluster to be ready..."
wait_for_nodes

# 6. Restore SealedSecrets key (if backup exists)
echo "🔐 Restoring SealedSecrets key..."
"$(dirname "$0")/sealed-secrets-restore.sh" "$CLUSTER_NAME" 2>/dev/null || echo "ℹ️  No sealed-secrets backup found"

# 7. Hand off to Pulumi for in-cluster reconciliation
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Cluster substrate ready for $CLUSTER_NAME!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📋 Next steps:"
echo "   1. Run Pulumi to deploy Flux, secrets, and workloads declaratively."
echo "      e.g. (from repo root) -> cd pulumi && pulumi up"
echo "   2. Observe progress via 'make observe' once Pulumi completes."
echo ""
echo "ℹ️  Flux, SealedSecrets, External Secrets, and workloads are now"
echo "    managed directly by Pulumi (operator-friendly)."
echo ""

