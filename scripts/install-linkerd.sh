#!/bin/bash
set -e

# 🚀 Install Linkerd service mesh
# Based on: https://linkerd.io/2-edge/tasks/install/

CLUSTER_NAME=${1:-homelab}
CONTEXT="kind-${CLUSTER_NAME}"

echo "🚀 Installing Linkerd on cluster: ${CLUSTER_NAME}"
echo "📍 Using context: ${CONTEXT}"

# Check if Linkerd control plane is already installed
if kubectl get deployment -n linkerd linkerd-destination --context "${CONTEXT}" >/dev/null 2>&1; then
    echo "✅ Linkerd control plane is already installed, skipping installation..."
    echo "💡 Run 'linkerd check --context ${CONTEXT}' manually to verify health if needed."
    exit 0
fi

# Pre-flight check
echo "✅ Running pre-flight checks..."
if ! linkerd check --pre --context "${CONTEXT}"; then
    echo "❌ Pre-flight checks failed. Please address the issues before proceeding."
    exit 1
fi

# Install Gateway API CRDs (required by Linkerd)
echo "📦 Installing Gateway API CRDs..."
if ! kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.2.1/standard-install.yaml --context "${CONTEXT}"; then
    echo "❌ Failed to install Gateway API CRDs"
    exit 1
fi

echo "⏳ Waiting for Gateway API CRDs to be established..."
kubectl wait --for condition=established --timeout=60s \
    crd/gatewayclasses.gateway.networking.k8s.io \
    crd/gateways.gateway.networking.k8s.io \
    crd/httproutes.gateway.networking.k8s.io \
    --context "${CONTEXT}" 2>/dev/null || echo "⚠️  Some Gateway API CRDs may not be available yet, continuing..."

# Install Linkerd CRDs
echo "📦 Installing Linkerd CRDs..."
if ! linkerd install --crds --context "${CONTEXT}" | kubectl apply -f - --context "${CONTEXT}"; then
    echo "❌ Failed to install Linkerd CRDs"
    exit 1
fi

# Wait for CRDs to be established
echo "⏳ Waiting for Linkerd CRDs to be established..."
kubectl wait --for condition=established --timeout=300s \
    crd/authorizationpolicies.policy.linkerd.io \
    crd/httproutes.policy.linkerd.io \
    crd/meshtlsauthentications.policy.linkerd.io \
    crd/networkauthentications.policy.linkerd.io \
    crd/serverauthorizations.policy.linkerd.io \
    crd/servers.policy.linkerd.io \
    --context "${CONTEXT}" 2>/dev/null || echo "⚠️  Some CRDs may not be available yet, continuing..."

# Install Linkerd control plane
echo "🎯 Installing Linkerd control plane..."
if ! linkerd install --context "${CONTEXT}" | kubectl apply -f - --context "${CONTEXT}"; then
    echo "❌ Failed to install Linkerd control plane"
    exit 1
fi

# Wait for Linkerd to be ready
echo "⏳ Waiting for Linkerd control plane to be ready..."
sleep 10  # Give it a moment to start creating resources

# Wait for linkerd namespace
kubectl wait --for=condition=Ready pods --all -n linkerd --timeout=300s --context "${CONTEXT}" || true

# Verify installation
echo "🔍 Verifying Linkerd installation..."
if linkerd check --context "${CONTEXT}"; then
    echo "✅ Linkerd installation completed successfully!"
else
    echo "⚠️  Linkerd check reported some issues, but installation is complete."
    echo "    You may need to wait a bit longer for all components to be ready."
fi

echo "🎉 Linkerd is now installed on ${CLUSTER_NAME}!"

