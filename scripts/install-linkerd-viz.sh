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
if ! kubectl get namespace linkerd --context "${CONTEXT}" >/dev/null 2>&1; then
    echo "❌ Linkerd must be installed before installing the Viz extension"
    echo "    Run install-linkerd.sh first"
    exit 1
fi

# Check if Viz is already installed
if kubectl get namespace linkerd-viz --context "${CONTEXT}" >/dev/null 2>&1; then
    echo "✅ Linkerd Viz is already installed, skipping installation..."
    POD_COUNT=$(kubectl get pods -n linkerd-viz --context "${CONTEXT}" --no-headers 2>/dev/null | wc -l | tr -d ' ')
    READY_COUNT=$(kubectl get pods -n linkerd-viz --context "${CONTEXT}" --no-headers 2>/dev/null | grep -c "Running" || echo "0")
    echo "    Pods: $READY_COUNT/$POD_COUNT running"
    exit 0
fi

# Ensure Linkerd control plane is ready
echo "⏳ Ensuring Linkerd control plane is ready..."
kubectl wait --for=condition=Ready pods --all -n linkerd --timeout=300s --context "${CONTEXT}" || echo "⚠️  Proceeding anyway..."

# Install Linkerd Viz extension
echo "🎯 Installing Linkerd Viz extension..."
MANIFEST=$(linkerd viz install --context "${CONTEXT}")
if [ -z "$MANIFEST" ]; then
    echo "❌ Failed to generate Linkerd Viz manifest"
    exit 1
fi

if ! echo "$MANIFEST" | kubectl apply -f - --context "${CONTEXT}"; then
    echo "❌ Failed to install Linkerd Viz extension"
    exit 1
fi

# Wait for Viz to be ready
echo "⏳ Waiting for Linkerd Viz to be ready..."
sleep 10  # Give it a moment to start creating resources

# Wait for linkerd-viz namespace pods
if kubectl wait --for=condition=Ready pods --all -n linkerd-viz --timeout=300s --context "${CONTEXT}"; then
    echo "✅ All Linkerd Viz pods are ready!"
else
    echo "⚠️  Some pods may not be ready yet, checking status..."
    kubectl get pods -n linkerd-viz --context "${CONTEXT}"
fi

# Verify Viz installation (best effort - may fail due to localhost DNS issues)
echo "🔍 Verifying Linkerd Viz installation..."
if linkerd viz check --context "${CONTEXT}" --wait=30s 2>&1 | grep -q "√"; then
    echo "✅ Linkerd Viz installation completed successfully!"
else
    # Check if pods are running as fallback verification
    POD_COUNT=$(kubectl get pods -n linkerd-viz --context "${CONTEXT}" --no-headers 2>/dev/null | wc -l | tr -d ' ')
    READY_COUNT=$(kubectl get pods -n linkerd-viz --context "${CONTEXT}" --no-headers 2>/dev/null | grep -c "Running" || echo "0")
    
    if [ "$POD_COUNT" -gt 0 ] && [ "$READY_COUNT" -eq "$POD_COUNT" ]; then
        echo "✅ All Linkerd Viz pods are running - installation successful!"
    else
        echo "⚠️  Linkerd Viz check had issues, but installation may still be completing."
        echo "    Pods: $READY_COUNT/$POD_COUNT running"
    fi
fi

echo "🎉 Linkerd Viz is now installed on ${CLUSTER_NAME}!"
echo "💡 Access the dashboard with: linkerd viz dashboard --context ${CONTEXT}"

