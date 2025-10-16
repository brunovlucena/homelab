#!/bin/bash
set -e

# 🚀 Create Kind Cluster
# Based on: https://kind.sigs.k8s.io/

CLUSTER_NAME=${1:-homelab}
CONTEXT="kind-${CLUSTER_NAME}"
CONFIG_FILE="../flux/clusters/${CLUSTER_NAME}/kind.yaml"

echo "🚀 Creating Kind cluster: ${CLUSTER_NAME}"
echo "📍 Using context: ${CONTEXT}"
echo "📋 Using config file: ${CONFIG_FILE}"

# Check if config file exists
if [ ! -f "${CONFIG_FILE}" ]; then
    echo "❌ Config file not found: ${CONFIG_FILE}"
    exit 1
fi

# Check if cluster already exists
if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
    echo "✅ Cluster '${CLUSTER_NAME}' already exists"
    
    # Verify cluster is accessible
    if kubectl cluster-info --context "${CONTEXT}" >/dev/null 2>&1; then
        echo "✅ Cluster is accessible and healthy!"
        
        # Export kubeconfig to ensure it's up to date
        echo "🔧 Exporting kubeconfig..."
        kind export kubeconfig --name "${CLUSTER_NAME}"
        
        # Show cluster info
        echo "📊 Cluster information:"
        kubectl cluster-info --context "${CONTEXT}"
        
        exit 0
    else
        echo "⚠️  Cluster exists but is not accessible. You may need to recreate it."
        exit 1
    fi
fi

# Create the cluster
echo "🏗️  Creating Kind cluster '${CLUSTER_NAME}'..."
if ! kind create cluster --name "${CLUSTER_NAME}" --config "${CONFIG_FILE}"; then
    echo "❌ Failed to create Kind cluster"
    exit 1
fi

# Export kubeconfig
echo "🔧 Exporting kubeconfig..."
if ! kind export kubeconfig --name "${CLUSTER_NAME}"; then
    echo "❌ Failed to export kubeconfig"
    exit 1
fi

# Verify cluster is accessible
echo "🔍 Verifying cluster is accessible..."
if ! kubectl cluster-info --context "${CONTEXT}"; then
    echo "❌ Failed to connect to cluster"
    exit 1
fi

# Wait for nodes to be ready
echo "⏳ Waiting for all nodes to be ready..."
if ! kubectl --context "${CONTEXT}" wait --for=condition=Ready nodes --all --timeout=300s; then
    echo "❌ Nodes failed to become ready within timeout"
    exit 1
fi

# Show node status
echo "📊 Cluster nodes:"
kubectl --context "${CONTEXT}" get nodes -o wide

echo "🎉 Kind cluster '${CLUSTER_NAME}' is ready!"
echo ""
echo "💡 Next steps:"
echo "   - Install Flux: ./install-flux.sh ${CLUSTER_NAME}"
echo "   - Create secrets: ./create-secrets.sh ${CONTEXT}"
echo "   - Install Linkerd: ./install-linkerd.sh ${CLUSTER_NAME}"
echo ""

