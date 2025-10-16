#!/bin/bash
set -e

# 🚀 Install Telepresence Traffic Manager
# Based on: https://telepresence.io/docs/install/manager

CLUSTER_NAME=${1:-homelab}
CONTEXT="kind-${CLUSTER_NAME}"
NAMESPACE=${2:-ambassador}

echo "🚀 Installing Telepresence Traffic Manager on cluster: ${CLUSTER_NAME}"
echo "📍 Using context: ${CONTEXT}"
echo "📦 Installing into namespace: ${NAMESPACE}"

# Check if telepresence CLI is installed
if ! command -v telepresence &> /dev/null; then
    echo "❌ telepresence CLI is not installed. Please install it first:"
    echo "   macOS: brew install datawire/blackbird/telepresence"
    echo "   Linux: https://telepresence.io/docs/install/"
    exit 1
fi

echo "✅ Telepresence CLI version: $(telepresence version 2>&1 | grep 'Client' || telepresence version)"

# Check if Traffic Manager is already installed
if kubectl get namespace "${NAMESPACE}" --context "${CONTEXT}" >/dev/null 2>&1; then
    if kubectl get deployment traffic-manager -n "${NAMESPACE}" --context "${CONTEXT}" >/dev/null 2>&1; then
        echo "✅ Telepresence Traffic Manager is already installed in namespace ${NAMESPACE}"
        
        # Check if it's healthy
        if kubectl rollout status deployment/traffic-manager -n "${NAMESPACE}" --context "${CONTEXT}" --timeout=10s >/dev/null 2>&1; then
            echo "✅ Traffic Manager is healthy!"
            exit 0
        else
            echo "⚠️  Traffic Manager is installed but may not be fully ready yet"
            exit 0
        fi
    fi
fi

# Pre-flight check - verify cluster connection
echo "✅ Running pre-flight checks..."
if ! kubectl cluster-info --context "${CONTEXT}" >/dev/null 2>&1; then
    echo "❌ Cannot connect to cluster. Please check your kubeconfig and cluster status."
    exit 1
fi

# Create namespace if it doesn't exist
if ! kubectl get namespace "${NAMESPACE}" --context "${CONTEXT}" >/dev/null 2>&1; then
    echo "📦 Creating namespace ${NAMESPACE}..."
    kubectl create namespace "${NAMESPACE}" --context "${CONTEXT}"
fi

# Install Telepresence Traffic Manager
echo "🎯 Installing Telepresence Traffic Manager..."
if ! telepresence helm install \
    --namespace "${NAMESPACE}" \
    --context "${CONTEXT}"; then
    echo "❌ Failed to install Telepresence Traffic Manager"
    exit 1
fi

# Wait for Traffic Manager to be ready
echo "⏳ Waiting for Traffic Manager to be ready..."
sleep 5  # Give it a moment to start creating resources

# Wait for the traffic-manager deployment
if kubectl wait --for=condition=Available deployment/traffic-manager \
    -n "${NAMESPACE}" \
    --timeout=300s \
    --context "${CONTEXT}" 2>/dev/null; then
    echo "✅ Traffic Manager deployment is ready!"
else
    echo "⚠️  Traffic Manager deployment may not be fully ready yet, continuing..."
fi

# Wait for all pods in the namespace
kubectl wait --for=condition=Ready pods --all \
    -n "${NAMESPACE}" \
    --timeout=300s \
    --context "${CONTEXT}" 2>/dev/null || echo "⚠️  Some pods may still be starting..."

# Verify installation
echo "🔍 Verifying Telepresence installation..."
if kubectl get deployment traffic-manager -n "${NAMESPACE}" --context "${CONTEXT}" >/dev/null 2>&1; then
    echo "✅ Telepresence Traffic Manager installation completed successfully!"
    
    # Show deployment status
    echo ""
    echo "📊 Traffic Manager status:"
    kubectl get deployment,pods -n "${NAMESPACE}" --context "${CONTEXT}" -l app.kubernetes.io/name=telepresence
else
    echo "❌ Failed to verify Telepresence installation"
    exit 1
fi

echo ""
echo "🎉 Telepresence is now installed on ${CLUSTER_NAME}!"
echo ""
echo "💡 To connect to the cluster, run:"
echo "   telepresence connect --context ${CONTEXT}"
echo ""
echo "💡 To uninstall, run:"
echo "   telepresence helm uninstall --namespace ${NAMESPACE} --context ${CONTEXT}"
echo ""

