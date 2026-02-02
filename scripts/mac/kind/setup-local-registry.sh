#!/bin/sh
# =============================================================================
# Local Container Registry Setup Script
# =============================================================================
# Based on official Kind documentation:
# https://kind.sigs.k8s.io/docs/user/local-registry/
#
# This script creates a local Docker registry and configures clusters
# (both Kind and k3s) to use it for testing container images.
#
# Usage:
#   ./scripts/mac/setup-local-registry.sh [cluster-name] [cluster-type]
#
# Examples:
#   ./scripts/mac/setup-local-registry.sh studio kind
#   ./scripts/mac/setup-local-registry.sh pro kind
#   ./scripts/mac/setup-local-registry.sh forge k3s
#   ./scripts/mac/setup-local-registry.sh pi k3s
#
# If no cluster type is provided, it will attempt to detect it.
# =============================================================================

set -o errexit

# Configuration (MUST match official Kind documentation)
# See: https://kind.sigs.k8s.io/docs/user/local-registry/
REG_NAME="${REGISTRY_NAME:-kind-registry}"
REG_PORT="${REGISTRY_PORT:-5001}"
CLUSTER_NAME="${1:-}"
CLUSTER_TYPE="${2:-}"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

info() {
  echo "${GREEN}[INFO]${NC} $1"
}

warn() {
  echo "${YELLOW}[WARN]${NC} $1"
}

error() {
  echo "${RED}[ERROR]${NC} $1"
  exit 1
}

# Detect cluster name if not provided
REGISTRY_ONLY=false
if [ -z "$CLUSTER_NAME" ]; then
  # Try to detect Kind clusters first
  CLUSTER_NAME=$(kind get clusters 2>/dev/null | head -n 1)
  if [ -n "$CLUSTER_NAME" ]; then
    CLUSTER_TYPE="kind"
    info "Using detected Kind cluster: $CLUSTER_NAME"
  else
    # No cluster found - just create the registry container
    # It will be connected to the cluster when the cluster is created
    info "No clusters found - setting up registry only (cluster will auto-connect on creation)"
    REGISTRY_ONLY=true
  fi
fi

# Detect cluster type if not provided (skip if registry-only mode)
if [ "$REGISTRY_ONLY" = false ]; then
  if [ -z "$CLUSTER_TYPE" ]; then
    # Check if it's a Kind cluster
    if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
      CLUSTER_TYPE="kind"
      info "Detected cluster type: kind"
    # Check if k3s is running
    elif kubectl config get-contexts | grep -q "$CLUSTER_NAME"; then
      CLUSTER_TYPE="k3s"
      info "Detected cluster type: k3s (assumed)"
    else
      error "Could not detect cluster type. Please specify: kind or k3s"
    fi
  fi

  # Validate cluster type
  if [ "$CLUSTER_TYPE" != "kind" ] && [ "$CLUSTER_TYPE" != "k3s" ]; then
    error "Invalid cluster type: $CLUSTER_TYPE (must be 'kind' or 'k3s')"
  fi

  info "Setting up local registry for cluster: $CLUSTER_NAME ($CLUSTER_TYPE)"
else
  info "Setting up local registry (no cluster configuration)"
fi

# =============================================================================
# Step 1: Create registry container unless it already exists
# =============================================================================
info "Step 1/5: Creating registry container..."

if [ "$(docker inspect -f '{{.State.Running}}' "${REG_NAME}" 2>/dev/null || true)" != 'true' ]; then
  info "Creating registry container '${REG_NAME}' on port ${REG_PORT}..."
  docker run \
    -d \
    --restart=always \
    -p "127.0.0.1:${REG_PORT}:5000" \
    --network bridge \
    --name "${REG_NAME}" \
    registry:2
  info "Registry container created successfully"
else
  info "Registry container '${REG_NAME}' already running"
fi

# =============================================================================
# Step 2: Configure cluster based on type (skip in registry-only mode)
# =============================================================================
if [ "$REGISTRY_ONLY" = false ]; then
  info "Step 2/5: Configuring cluster ($CLUSTER_TYPE)..."

  if [ "$CLUSTER_TYPE" = "kind" ]; then
  # Kind cluster configuration
  
  # Check if cluster was created with the correct containerd config
  CONTROL_PLANE=$(kind get nodes --name "${CLUSTER_NAME}" | head -n 1)
  if ! docker exec "${CONTROL_PLANE}" test -d "/etc/containerd/certs.d" 2>/dev/null; then
    warn "Cluster was not created with containerd config_path enabled"
    warn "This cluster may not work correctly with the local registry"
    warn "Consider recreating the cluster with the updated kind.yaml"
  fi
  
  # =============================================================================
  # Step 3: Add the registry config to all Kind nodes
  # =============================================================================
  info "Step 3/5: Configuring Kind nodes..."
  
  # Configure two registry endpoints:
  # 1. localhost:${REG_PORT} - for host-side access (docker push/pull)
  # 2. kind-registry:5000 - for pod-side access (Kubernetes image pulls)
  
  NODE_COUNT=0
  
  for node in $(kind get nodes --name "${CLUSTER_NAME}"); do
    info "Configuring node: $node"
    
    # Configure localhost:${REG_PORT} (host-side)
    REGISTRY_DIR_HOST="/etc/containerd/certs.d/localhost:${REG_PORT}"
    docker exec "${node}" mkdir -p "${REGISTRY_DIR_HOST}"
    cat <<EOF | docker exec -i "${node}" cp /dev/stdin "${REGISTRY_DIR_HOST}/hosts.toml"
server = "http://${REG_NAME}:5000"

[host."http://${REG_NAME}:5000"]
  capabilities = ["pull", "resolve", "push"]
  skip_verify = true
EOF
    
    # Configure kind-registry:5000 (pod-side, insecure HTTP)
    REGISTRY_DIR_POD="/etc/containerd/certs.d/kind-registry:5000"
    docker exec "${node}" mkdir -p "${REGISTRY_DIR_POD}"
    cat <<EOF | docker exec -i "${node}" cp /dev/stdin "${REGISTRY_DIR_POD}/hosts.toml"
server = "http://kind-registry:5000"

[host."http://kind-registry:5000"]
  capabilities = ["pull", "resolve", "push"]
  skip_verify = true
EOF
    
    NODE_COUNT=$((NODE_COUNT + 1))
  done
  
  info "Configured $NODE_COUNT nodes (both localhost:${REG_PORT} and kind-registry:5000)"
  
  # =============================================================================
  # Step 4: Connect the registry to the Kind network
  # =============================================================================
  info "Step 4/5: Connecting registry to Kind network..."
  
  if [ "$(docker inspect -f='{{json .NetworkSettings.Networks.kind}}' "${REG_NAME}")" = 'null' ]; then
    info "Connecting '${REG_NAME}' to 'kind' network..."
    docker network connect "kind" "${REG_NAME}"
    info "Registry connected to cluster network"
  else
    info "Registry already connected to cluster network"
  fi

elif [ "$CLUSTER_TYPE" = "k3s" ]; then
  # k3s cluster configuration
  info "Step 3/5: Configuring k3s cluster..."
  
  warn "k3s requires manual configuration of /etc/rancher/k3s/registries.yaml"
  warn "Please ensure the following is in /etc/rancher/k3s/registries.yaml on each node:"
  echo ""
  cat <<EOF
mirrors:
  "localhost:${REG_PORT}":
    endpoint:
      - "http://localhost:${REG_PORT}"

configs:
  "localhost:${REG_PORT}":
    tls:
      insecure_skip_verify: true
EOF
  echo ""
  warn "After updating registries.yaml, restart k3s:"
  warn "  sudo systemctl restart k3s        (control-plane)"
  warn "  sudo systemctl restart k3s-agent  (workers)"
  echo ""
  
    info "Step 4/5: Skipping network connection (not needed for k3s)..."
  fi

  # =============================================================================
  # Step 5: Document the local registry
  # =============================================================================
  info "Step 5/5: Creating registry documentation ConfigMap..."

  # Switch to the correct cluster context
  kubectl config use-context "kind-${CLUSTER_NAME}" >/dev/null 2>&1

  cat <<EOF | kubectl apply -f - >/dev/null
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-registry-hosting
  namespace: kube-public
data:
  localRegistryHosting.v1: |
    host: "localhost:${REG_PORT}"
    help: "https://kind.sigs.k8s.io/docs/user/local-registry/"
EOF

  info "ConfigMap created in kube-public namespace"
else
  # Registry-only mode: skip cluster configuration steps
  info "Step 2/5: Skipping cluster configuration (no cluster found)"
  info "Step 3/5: Skipping node configuration (no cluster found)"
  info "Step 4/5: Skipping network connection (no cluster found)"
  info "Step 5/5: Skipping ConfigMap creation (no cluster found)"
  info "Registry will be automatically connected when cluster is created"
fi

# =============================================================================
# Verification and Usage Instructions
# =============================================================================
echo ""
echo "${GREEN}✅ Local registry setup complete!${NC}"
echo ""
echo "Registry Details:"
echo "  Container Name: ${REG_NAME}"
echo "  Host Port:      localhost:${REG_PORT}"
echo "  Internal Port:  ${REG_NAME}:5000"
if [ "$REGISTRY_ONLY" = false ]; then
  echo "  Cluster:        ${CLUSTER_NAME} (${CLUSTER_TYPE})"
else
  echo "  Cluster:        None (registry-only mode)"
fi
echo ""
echo "Usage Examples:"
echo "  ${YELLOW}# Pull an image${NC}"
echo "  docker pull nginx:latest"
echo ""
echo "  ${YELLOW}# Tag for local registry${NC}"
echo "  docker tag nginx:latest localhost:${REG_PORT}/nginx:latest"
echo ""
echo "  ${YELLOW}# Push to local registry${NC}"
echo "  docker push localhost:${REG_PORT}/nginx:latest"
echo ""
echo "  ${YELLOW}# Use in Kubernetes${NC}"
echo "  kubectl run nginx --image=localhost:${REG_PORT}/nginx:latest"
echo ""
echo "  ${YELLOW}# List images in registry${NC}"
echo "  curl http://localhost:${REG_PORT}/v2/_catalog"
echo ""

# Test registry connectivity
info "Testing registry connectivity..."
if curl -s "http://localhost:${REG_PORT}/v2/_catalog" >/dev/null; then
  info "Registry is accessible at http://localhost:${REG_PORT}"
else
  warn "Could not connect to registry. Please check Docker is running."
fi

echo ""
if [ "$REGISTRY_ONLY" = false ]; then
  if [ "$CLUSTER_TYPE" = "kind" ]; then
    echo "For more information, see: https://kind.sigs.k8s.io/docs/user/local-registry/"
  else
    echo "For k3s registry configuration, see: https://docs.k3s.io/installation/private-registry"
  fi
else
  echo "For more information, see: https://kind.sigs.k8s.io/docs/user/local-registry/"
  echo ""
  echo "${YELLOW}ℹ️  Note:${NC} Registry is ready and will be automatically connected when cluster is created."
fi

# Environment variable tips
echo ""
echo "${YELLOW}💡 Tips:${NC}"
echo "  • Override registry name: ${GREEN}REGISTRY_NAME=my-registry${NC} ./scripts/mac/setup-local-registry.sh"
echo "  • Override registry port: ${GREEN}REGISTRY_PORT=5002${NC} ./scripts/mac/setup-local-registry.sh"
echo "  • Use with all clusters: The same registry works for Air, Pro, Studio, Forge, and Pi!"
if [ "$REGISTRY_ONLY" = true ]; then
  echo "  • Connect to cluster: The registry will auto-connect when you create a cluster with 'make up'"
fi
echo ""

