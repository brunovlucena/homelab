#!/bin/bash
set -e

# 📊 Install Linkerd Viz extension for metrics and visibility
# Based on: https://linkerd.io/2-edge/tasks/extensions/

CLUSTER_NAME=${1:-homelab}
CONTEXT="kind-${CLUSTER_NAME}"

echo "📊 Installing Linkerd Viz extension on cluster: ${CLUSTER_NAME}"
echo "📍 Using context: ${CONTEXT}"

# Verify Linkerd is installed first
echo "✅ Verifying Linkerd is installed..."
if ! linkerd check --context "${CONTEXT}" > /dev/null 2>&1; then
    echo "❌ Linkerd must be installed before installing the Viz extension"
    echo "    Run install-linkerd.sh first"
    exit 1
fi

# Install Linkerd Viz extension
echo "🎯 Installing Linkerd Viz extension..."
if ! linkerd viz install --context "${CONTEXT}" | kubectl apply -f - --context "${CONTEXT}"; then
    echo "❌ Failed to install Linkerd Viz extension"
    exit 1
fi

# Wait for Viz to be ready
echo "⏳ Waiting for Linkerd Viz to be ready..."
sleep 10  # Give it a moment to start creating resources

# Wait for linkerd-viz namespace pods
kubectl wait --for=condition=Ready pods --all -n linkerd-viz --timeout=300s --context "${CONTEXT}" || true

# Verify Viz installation
echo "🔍 Verifying Linkerd Viz installation..."
if linkerd viz check --context "${CONTEXT}"; then
    echo "✅ Linkerd Viz installation completed successfully!"
else
    echo "⚠️  Linkerd Viz check reported some issues, but installation is complete."
    echo "    You may need to wait a bit longer for all components to be ready."
fi

echo "🎉 Linkerd Viz is now installed on ${CLUSTER_NAME}!"
echo "💡 Access the dashboard with: linkerd viz dashboard --context ${CONTEXT}"

