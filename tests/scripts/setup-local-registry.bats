#!/usr/bin/env bats
# =============================================================================
# Tests for setup-local-registry.sh
# =============================================================================
# Test coverage for local container registry setup script
# 
# Prerequisites:
#   - bats-core installed
#   - bats-support and bats-assert libraries
#   - Docker running
#   - kubectl installed
#
# Usage: bats setup-local-registry.bats
# =============================================================================

# Load test helpers
load '../helpers/test_helper'

# Setup function runs before each test
setup() {
    # Set test environment variables
    export REG_NAME="test-registry-$$"
    export REG_PORT="5999"
    export SCRIPT_PATH="../../scripts/mac/setup-local-registry.sh"
    
    # Ensure script exists and is executable
    [ -f "$SCRIPT_PATH" ]
    [ -x "$SCRIPT_PATH" ]
    
    # Clean up any existing test registry
    docker rm -f "$REG_NAME" 2>/dev/null || true
}

# Teardown function runs after each test
teardown() {
    # Clean up test registry
    docker rm -f "$REG_NAME" 2>/dev/null || true
    
    # Clean up any test clusters
    kind delete cluster --name "test-cluster-$$" 2>/dev/null || true
    
    return 0
}

# =============================================================================
# Basic Functionality Tests
# =============================================================================

@test "script exists and is executable" {
    [ -f "$SCRIPT_PATH" ]
    [ -x "$SCRIPT_PATH" ]
}

@test "script has correct shebang" {
    run head -n 1 "$SCRIPT_PATH"
    assert_output --partial "#!/bin/sh"
}

@test "script contains required configuration variables" {
    run grep -E "REG_NAME|REG_PORT|CLUSTER_NAME|CLUSTER_TYPE" "$SCRIPT_PATH"
    assert_success
}

@test "script validates cluster type parameter" {
    skip "Requires Kind cluster setup"
    
    run bash "$SCRIPT_PATH" test-cluster invalid-type
    assert_failure
    assert_output --partial "Invalid cluster type"
}

# =============================================================================
# Registry Container Tests
# =============================================================================

@test "creates registry container if not exists" {
    skip "Requires Docker"
    
    # Ensure registry doesn't exist
    docker rm -f "$REG_NAME" 2>/dev/null || true
    
    # Run script to create registry
    export REGISTRY_NAME="$REG_NAME"
    export REGISTRY_PORT="$REG_PORT"
    
    # Create minimal Kind cluster for test
    cat > /tmp/kind-config-$$.yaml <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: test-cluster-$$
EOF
    
    kind create cluster --config /tmp/kind-config-$$.yaml --wait 60s
    
    run bash "$SCRIPT_PATH" "test-cluster-$$" kind
    assert_success
    assert_output --partial "Registry container created successfully"
    
    # Verify container exists
    run docker inspect "$REG_NAME"
    assert_success
    
    # Cleanup
    kind delete cluster --name "test-cluster-$$"
    rm /tmp/kind-config-$$.yaml
}

@test "detects existing registry container" {
    skip "Requires Docker"
    
    # Create registry manually
    docker run -d --name "$REG_NAME" -p "127.0.0.1:$REG_PORT:5000" registry:2
    
    export REGISTRY_NAME="$REG_NAME"
    export REGISTRY_PORT="$REG_PORT"
    
    run bash "$SCRIPT_PATH" studio kind
    assert_success
    assert_output --partial "already running"
}

@test "registry container has correct port mapping" {
    skip "Requires Docker"
    
    export REGISTRY_NAME="$REG_NAME"
    export REGISTRY_PORT="$REG_PORT"
    
    run bash "$SCRIPT_PATH" studio kind
    assert_success
    
    # Check port mapping
    run docker port "$REG_NAME"
    assert_success
    assert_output --partial "127.0.0.1:$REG_PORT"
}

@test "registry container has restart policy" {
    skip "Requires Docker"
    
    export REGISTRY_NAME="$REG_NAME"
    export REGISTRY_PORT="$REG_PORT"
    
    bash "$SCRIPT_PATH" studio kind
    
    run docker inspect --format='{{.HostConfig.RestartPolicy.Name}}' "$REG_NAME"
    assert_success
    assert_output "always"
}

# =============================================================================
# Cluster Detection Tests
# =============================================================================

@test "detects Kind cluster type automatically" {
    skip "Requires Kind cluster"
    
    # Create test Kind cluster
    kind create cluster --name "test-cluster-$$"
    
    # Run without cluster type parameter
    run bash "$SCRIPT_PATH" "test-cluster-$$"
    assert_success
    assert_output --partial "Detected cluster type: kind"
    
    kind delete cluster --name "test-cluster-$$"
}

@test "fails gracefully when no cluster found" {
    run bash "$SCRIPT_PATH" non-existent-cluster-12345 kind
    assert_failure
}

@test "accepts cluster name from command line" {
    run grep -E 'CLUSTER_NAME="\${1:-""}' "$SCRIPT_PATH"
    assert_success
}

@test "accepts cluster type from command line" {
    run grep -E 'CLUSTER_TYPE="\${2:-""}' "$SCRIPT_PATH"
    assert_success
}

# =============================================================================
# Kind Cluster Configuration Tests
# =============================================================================

@test "configures Kind cluster nodes with containerd settings" {
    skip "Requires Kind cluster"
    
    kind create cluster --name "test-cluster-$$"
    
    export REGISTRY_NAME="$REG_NAME"
    export REGISTRY_PORT="$REG_PORT"
    
    run bash "$SCRIPT_PATH" "test-cluster-$$" kind
    assert_success
    assert_output --partial "Configuring Kind nodes"
}

@test "creates hosts.toml on Kind nodes" {
    skip "Requires Kind cluster"
    
    kind create cluster --name "test-cluster-$$"
    
    export REGISTRY_NAME="$REG_NAME"
    run bash "$SCRIPT_PATH" "test-cluster-$$" kind
    assert_success
    
    # Verify hosts.toml was created
    CONTROL_PLANE=$(kind get nodes --name "test-cluster-$$" | head -n 1)
    run docker exec "$CONTROL_PLANE" cat "/etc/containerd/certs.d/localhost:$REG_PORT/hosts.toml"
    assert_success
    assert_output --partial "http://$REG_NAME:5000"
    
    kind delete cluster --name "test-cluster-$$"
}

@test "connects registry to Kind network" {
    skip "Requires Kind cluster and Docker"
    
    kind create cluster --name "test-cluster-$$"
    
    export REGISTRY_NAME="$REG_NAME"
    export REGISTRY_PORT="$REG_PORT"
    
    run bash "$SCRIPT_PATH" "test-cluster-$$" kind
    assert_success
    assert_output --partial "Registry connected to cluster network"
    
    # Verify network connection
    run docker inspect --format='{{json .NetworkSettings.Networks.kind}}' "$REG_NAME"
    assert_success
    refute_output "null"
    
    kind delete cluster --name "test-cluster-$$"
}

@test "warns if containerd config_path not enabled" {
    skip "Requires Kind cluster without containerd config"
    
    # This would require a specially configured cluster
    run bash "$SCRIPT_PATH" "test-cluster-$$" kind
    assert_output --partial "Consider recreating the cluster"
}

# =============================================================================
# k3s Configuration Tests
# =============================================================================

@test "provides k3s configuration instructions" {
    run grep -A 10 'k3s cluster configuration' "$SCRIPT_PATH"
    assert_success
    assert_output --partial "registries.yaml"
}

@test "shows k3s restart commands" {
    run grep 'systemctl restart k3s' "$SCRIPT_PATH"
    assert_success
}

@test "skips network connection for k3s" {
    run grep 'Skipping network connection (not needed for k3s)' "$SCRIPT_PATH"
    assert_success
}

# =============================================================================
# ConfigMap Creation Tests
# =============================================================================

@test "creates ConfigMap in kube-public namespace" {
    skip "Requires Kind cluster and kubectl"
    
    kind create cluster --name "test-cluster-$$"
    
    export REGISTRY_NAME="$REG_NAME"
    run bash "$SCRIPT_PATH" "test-cluster-$$" kind
    assert_success
    
    # Verify ConfigMap exists
    run kubectl get configmap local-registry-hosting -n kube-public --context "kind-test-cluster-$$"
    assert_success
    
    kind delete cluster --name "test-cluster-$$"
}

@test "ConfigMap contains registry host information" {
    skip "Requires Kind cluster and kubectl"
    
    kind create cluster --name "test-cluster-$$"
    
    export REGISTRY_NAME="$REG_NAME"
    export REGISTRY_PORT="$REG_PORT"
    
    bash "$SCRIPT_PATH" "test-cluster-$$" kind
    
    run kubectl get configmap local-registry-hosting -n kube-public \
        --context "kind-test-cluster-$$" -o yaml
    assert_success
    assert_output --partial "localhost:$REG_PORT"
    
    kind delete cluster --name "test-cluster-$$"
}

@test "ConfigMap includes help URL" {
    run grep 'https://kind.sigs.k8s.io/docs/user/local-registry/' "$SCRIPT_PATH"
    assert_success
}

# =============================================================================
# Output and Messaging Tests
# =============================================================================

@test "displays colored output for info messages" {
    run grep "GREEN=" "$SCRIPT_PATH"
    assert_success
}

@test "displays colored output for warning messages" {
    run grep "YELLOW=" "$SCRIPT_PATH"
    assert_success
}

@test "displays colored output for error messages" {
    run grep "RED=" "$SCRIPT_PATH"
    assert_success
}

@test "provides usage examples in output" {
    run grep 'Usage Examples:' "$SCRIPT_PATH"
    assert_success
}

@test "shows docker tag command example" {
    run grep 'docker tag' "$SCRIPT_PATH"
    assert_success
}

@test "shows docker push command example" {
    run grep 'docker push' "$SCRIPT_PATH"
    assert_success
}

@test "shows kubectl run command example" {
    run grep 'kubectl run' "$SCRIPT_PATH"
    assert_success
}

@test "shows registry catalog command example" {
    run grep 'curl.*_catalog' "$SCRIPT_PATH"
    assert_success
}

# =============================================================================
# Registry Connectivity Tests
# =============================================================================

@test "tests registry connectivity after setup" {
    run grep 'Testing registry connectivity' "$SCRIPT_PATH"
    assert_success
}

@test "verifies registry API accessibility" {
    run grep 'curl.*v2/_catalog' "$SCRIPT_PATH"
    assert_success
}

# =============================================================================
# Environment Variable Tests
# =============================================================================

@test "respects REGISTRY_NAME environment variable" {
    run grep 'REGISTRY_NAME:-local-registry' "$SCRIPT_PATH"
    assert_success
}

@test "respects REGISTRY_PORT environment variable" {
    run grep 'REGISTRY_PORT:-5001' "$SCRIPT_PATH"
    assert_success
}

@test "shows environment variable override tips" {
    run grep 'Override registry name' "$SCRIPT_PATH"
    assert_success
    
    run grep 'Override registry port' "$SCRIPT_PATH"
    assert_success
}

# =============================================================================
# Error Handling Tests
# =============================================================================

@test "script uses 'set -o errexit' for error handling" {
    run head -n 25 "$SCRIPT_PATH"
    assert_output --partial "set -o errexit"
}

@test "has error function defined" {
    run grep -A 3 '^error()' "$SCRIPT_PATH"
    assert_success
    assert_output --partial "exit 1"
}

@test "has warn function defined" {
    run grep '^warn()' "$SCRIPT_PATH"
    assert_success
}

@test "has info function defined" {
    run grep '^info()' "$SCRIPT_PATH"
    assert_success
}

# =============================================================================
# Documentation Tests
# =============================================================================

@test "script contains usage documentation in header" {
    run head -n 30 "$SCRIPT_PATH"
    assert_output --partial "Usage:"
}

@test "script provides examples in header" {
    run head -n 30 "$SCRIPT_PATH"
    assert_output --partial "Examples:"
}

@test "script references official Kind documentation" {
    run head -n 30 "$SCRIPT_PATH"
    assert_output --partial "https://kind.sigs.k8s.io"
}

@test "script mentions all supported clusters" {
    run grep -i 'studio\|pro\|forge\|pi\|air' "$SCRIPT_PATH"
    assert_success
}

# =============================================================================
# Integration Tests
# =============================================================================

@test "end-to-end: setup registry for Kind cluster" {
    skip "Full integration test - requires Docker and Kind"
    
    # Create test cluster
    cat > /tmp/kind-config-$$.yaml <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: test-e2e-$$
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry]
    config_path = "/etc/containerd/certs.d"
EOF
    
    kind create cluster --config /tmp/kind-config-$$.yaml --wait 60s
    
    # Setup registry
    export REGISTRY_NAME="test-registry-$$"
    export REGISTRY_PORT="5999"
    
    run bash "$SCRIPT_PATH" "test-e2e-$$" kind
    assert_success
    
    # Verify registry works end-to-end
    docker pull busybox:latest
    docker tag busybox:latest "localhost:5999/busybox:test"
    docker push "localhost:5999/busybox:test"
    
    # Test in cluster
    kubectl run test-pod --image=localhost:5999/busybox:test \
        --context "kind-test-e2e-$$" --command -- sleep 3600
    
    kubectl wait --for=condition=ready pod/test-pod \
        --context "kind-test-e2e-$$" --timeout=60s
    
    # Cleanup
    kubectl delete pod test-pod --context "kind-test-e2e-$$"
    kind delete cluster --name "test-e2e-$$"
    docker rm -f "test-registry-$$"
    rm /tmp/kind-config-$$.yaml
}

@test "stress test: multiple clusters with same registry" {
    skip "Stress test - requires significant resources"
    
    # This test would create multiple clusters and verify they all work with the same registry
    # Implementation details would depend on available resources
}

