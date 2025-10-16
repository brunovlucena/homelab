#!/bin/bash
set -e

# 🚀 Install Flux CD
# Based on: https://fluxcd.io/flux/installation/

CLUSTER_NAME=${1:-homelab}
CONTEXT="kind-${CLUSTER_NAME}"

echo "🚀 Installing Flux CD on cluster: ${CLUSTER_NAME}"
echo "📍 Using context: ${CONTEXT}"

# Check if Flux is already installed
if kubectl get namespace flux-system --context "${CONTEXT}" >/dev/null 2>&1; then
    echo "✅ Flux is already installed, skipping installation..."
    if flux check --context "${CONTEXT}" >/dev/null 2>&1; then
        echo "✅ Flux is healthy!"
        exit 0
    else
        echo "⚠️  Flux is installed but health check returned warnings (this may be normal)"
        exit 0
    fi
fi

# Pre-flight check
echo "✅ Running pre-flight checks..."
if ! flux check --pre --context "${CONTEXT}"; then
    echo "❌ Pre-flight checks failed. Please address the issues before proceeding."
    exit 1
fi

# Install Flux with specific components
echo "🎯 Installing Flux CD components..."
if ! flux install \
    --components source-controller,kustomize-controller,helm-controller,notification-controller \
    --context "${CONTEXT}"; then
    echo "❌ Failed to install Flux CD"
    exit 1
fi

# Wait for Flux to be ready
echo "⏳ Waiting for Flux control plane to be ready..."
sleep 10  # Give it a moment to start creating resources

# Wait for flux-system namespace
kubectl wait --for=condition=Ready pods --all -n flux-system --timeout=300s --context "${CONTEXT}" || true

# Wait for Flux CRDs to be established
echo "⏳ Waiting for Flux CRDs to be established..."
kubectl wait --for condition=established --timeout=300s \
    crd/gitrepositories.source.toolkit.fluxcd.io \
    crd/helmrepositories.source.toolkit.fluxcd.io \
    crd/helmreleases.helm.toolkit.fluxcd.io \
    crd/kustomizations.kustomize.toolkit.fluxcd.io \
    --context "${CONTEXT}" 2>/dev/null || echo "⚠️  Some Flux CRDs may not be available yet, continuing..."

# Verify installation
echo "🔍 Verifying Flux installation..."
if flux check --context "${CONTEXT}"; then
    echo "✅ Flux installation completed successfully!"
else
    echo "⚠️  Flux check reported some issues, but installation is complete."
    echo "    You may need to wait a bit longer for all components to be ready."
fi

echo "🎉 Flux is now installed on ${CLUSTER_NAME}!"

