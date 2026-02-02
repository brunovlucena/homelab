#!/usr/bin/env bash
# =============================================================================
# Test Helper Functions
# =============================================================================
# Common helper functions and setup for BATS tests
#
# This file provides:
#   - Common test utilities
#   - Setup/teardown helpers
#   - Assertion helpers
#   - Mock functions
#   - Test fixtures
# =============================================================================

# Load BATS libraries if available
if [ -n "$BATS_TEST_DIRNAME" ]; then
    # Try to load bats-support and bats-assert
    if [ -f "/opt/homebrew/lib/bats-support/load.bash" ]; then
        load '/opt/homebrew/lib/bats-support/load.bash'
    elif [ -f "/usr/local/lib/bats-support/load.bash" ]; then
        load '/usr/local/lib/bats-support/load.bash'
    fi
    
    if [ -f "/opt/homebrew/lib/bats-assert/load.bash" ]; then
        load '/opt/homebrew/lib/bats-assert/load.bash'
    elif [ -f "/usr/local/lib/bats-assert/load.bash" ]; then
        load '/usr/local/lib/bats-assert/load.bash'
    fi
fi

# =============================================================================
# Color Output Functions
# =============================================================================

color_green() {
    printf '\033[0;32m%s\033[0m\n' "$1"
}

color_yellow() {
    printf '\033[1;33m%s\033[0m\n' "$1"
}

color_red() {
    printf '\033[0;31m%s\033[0m\n' "$1"
}

color_blue() {
    printf '\033[0;34m%s\033[0m\n' "$1"
}

# =============================================================================
# Test Environment Functions
# =============================================================================

# Check if running in CI environment
is_ci() {
    [ -n "$CI" ] || [ -n "$GITHUB_ACTIONS" ]
}

# Check if Docker is available
has_docker() {
    command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1
}

# Check if kubectl is available
has_kubectl() {
    command -v kubectl >/dev/null 2>&1
}

# Check if Kind is available
has_kind() {
    command -v kind >/dev/null 2>&1
}

# Check if jq is available
has_jq() {
    command -v jq >/dev/null 2>&1
}

# Skip test if requirement is not met
require_docker() {
    if ! has_docker; then
        skip "Docker is required but not available"
    fi
}

require_kubectl() {
    if ! has_kubectl; then
        skip "kubectl is required but not available"
    fi
}

require_kind() {
    if ! has_kind; then
        skip "Kind is required but not available"
    fi
}

require_jq() {
    if ! has_jq; then
        skip "jq is required but not available"
    fi
}

# =============================================================================
# Kind Cluster Helpers
# =============================================================================

# Create a temporary Kind cluster for testing
create_test_kind_cluster() {
    local cluster_name="${1:-test-cluster-$$}"
    local config_file="${2:-}"
    
    if [ -n "$config_file" ]; then
        kind create cluster --name "$cluster_name" --config "$config_file" --wait 60s
    else
        kind create cluster --name "$cluster_name" --wait 60s
    fi
    
    echo "$cluster_name"
}

# Delete test Kind cluster
delete_test_kind_cluster() {
    local cluster_name="$1"
    kind delete cluster --name "$cluster_name" 2>/dev/null || true
}

# Wait for Kind cluster to be ready
wait_for_kind_cluster() {
    local cluster_name="$1"
    local timeout="${2:-120}"
    local elapsed=0
    
    while [ $elapsed -lt $timeout ]; do
        if kubectl cluster-info --context "kind-$cluster_name" >/dev/null 2>&1; then
            return 0
        fi
        sleep 2
        elapsed=$((elapsed + 2))
    done
    
    return 1
}

# =============================================================================
# Docker Helpers
# =============================================================================

# Clean up Docker container
cleanup_docker_container() {
    local container_name="$1"
    docker rm -f "$container_name" 2>/dev/null || true
}

# Check if Docker container exists
docker_container_exists() {
    local container_name="$1"
    docker inspect "$container_name" >/dev/null 2>&1
}

# Check if Docker container is running
docker_container_running() {
    local container_name="$1"
    [ "$(docker inspect -f '{{.State.Running}}' "$container_name" 2>/dev/null)" = "true" ]
}

# Wait for Docker container to be ready
wait_for_container() {
    local container_name="$1"
    local timeout="${2:-60}"
    local elapsed=0
    
    while [ $elapsed -lt $timeout ]; do
        if docker_container_running "$container_name"; then
            return 0
        fi
        sleep 1
        elapsed=$((elapsed + 1))
    done
    
    return 1
}

# =============================================================================
# Kubernetes Helpers
# =============================================================================

# Wait for namespace to exist
wait_for_namespace() {
    local namespace="$1"
    local context="${2:-}"
    local timeout="${3:-60}"
    local elapsed=0
    
    local context_arg=""
    [ -n "$context" ] && context_arg="--context $context"
    
    while [ $elapsed -lt $timeout ]; do
        if kubectl get namespace "$namespace" $context_arg >/dev/null 2>&1; then
            return 0
        fi
        sleep 2
        elapsed=$((elapsed + 2))
    done
    
    return 1
}

# Wait for pod to be ready
wait_for_pod() {
    local pod_name="$1"
    local namespace="${2:-default}"
    local context="${3:-}"
    local timeout="${4:-120}"
    
    local context_arg=""
    [ -n "$context" ] && context_arg="--context $context"
    
    kubectl wait --for=condition=ready "pod/$pod_name" \
        -n "$namespace" $context_arg --timeout="${timeout}s"
}

# Wait for deployment to be ready
wait_for_deployment() {
    local deployment_name="$1"
    local namespace="${2:-default}"
    local context="${3:-}"
    local timeout="${4:-120}"
    
    local context_arg=""
    [ -n "$context" ] && context_arg="--context $context"
    
    kubectl wait --for=condition=available "deployment/$deployment_name" \
        -n "$namespace" $context_arg --timeout="${timeout}s"
}

# =============================================================================
# File and Directory Helpers
# =============================================================================

# Create temporary directory
create_temp_dir() {
    mktemp -d "${TMPDIR:-/tmp}/homelab-test.XXXXXX"
}

# Create temporary file
create_temp_file() {
    local suffix="${1:-}"
    mktemp "${TMPDIR:-/tmp}/homelab-test.XXXXXX${suffix}"
}

# Clean up temporary directory
cleanup_temp_dir() {
    local dir="$1"
    [ -d "$dir" ] && rm -rf "$dir"
}

# =============================================================================
# Assertion Helpers
# =============================================================================

# Assert file exists
assert_file_exists() {
    local file="$1"
    [ -f "$file" ] || {
        echo "File does not exist: $file"
        return 1
    }
}

# Assert directory exists
assert_directory_exists() {
    local dir="$1"
    [ -d "$dir" ] || {
        echo "Directory does not exist: $dir"
        return 1
    }
}

# Assert command succeeds
assert_command_success() {
    "$@" || {
        echo "Command failed: $*"
        return 1
    }
}

# Assert command fails
assert_command_failure() {
    if "$@"; then
        echo "Command succeeded but should have failed: $*"
        return 1
    fi
    return 0
}

# Assert string contains substring
assert_contains() {
    local string="$1"
    local substring="$2"
    
    if [[ "$string" != *"$substring"* ]]; then
        echo "String does not contain expected substring"
        echo "String: $string"
        echo "Expected substring: $substring"
        return 1
    fi
}

# Assert string does not contain substring
assert_not_contains() {
    local string="$1"
    local substring="$2"
    
    if [[ "$string" == *"$substring"* ]]; then
        echo "String contains unexpected substring"
        echo "String: $string"
        echo "Unexpected substring: $substring"
        return 1
    fi
}

# =============================================================================
# Mock Functions
# =============================================================================

# Mock kubectl command
mock_kubectl() {
    kubectl() {
        echo "MOCK: kubectl $*"
        return 0
    }
    export -f kubectl
}

# Mock kind command
mock_kind() {
    kind() {
        echo "MOCK: kind $*"
        return 0
    }
    export -f kind
}

# Mock docker command
mock_docker() {
    docker() {
        echo "MOCK: docker $*"
        return 0
    }
    export -f docker
}

# Restore real commands
restore_commands() {
    unset -f kubectl 2>/dev/null || true
    unset -f kind 2>/dev/null || true
    unset -f docker 2>/dev/null || true
}

# =============================================================================
# Test Fixtures
# =============================================================================

# Get path to test fixtures directory
get_fixtures_dir() {
    echo "${BATS_TEST_DIRNAME}/../fixtures"
}

# Get path to specific fixture file
get_fixture() {
    local fixture_name="$1"
    echo "$(get_fixtures_dir)/$fixture_name"
}

# Load fixture content
load_fixture() {
    local fixture_name="$1"
    cat "$(get_fixture "$fixture_name")"
}

# =============================================================================
# Logging Helpers
# =============================================================================

# Log test info message
log_info() {
    color_blue "[INFO] $*" >&3
}

# Log test warning message
log_warn() {
    color_yellow "[WARN] $*" >&3
}

# Log test error message
log_error() {
    color_red "[ERROR] $*" >&3
}

# =============================================================================
# Cleanup Helpers
# =============================================================================

# Register cleanup function
register_cleanup() {
    local cleanup_func="$1"
    CLEANUP_FUNCTIONS+=("$cleanup_func")
}

# Run all cleanup functions
run_cleanups() {
    for cleanup_func in "${CLEANUP_FUNCTIONS[@]:-}"; do
        $cleanup_func || true
    done
}

# Initialize cleanup array
CLEANUP_FUNCTIONS=()

# =============================================================================
# Test Data Generation
# =============================================================================

# Generate random string
random_string() {
    local length="${1:-8}"
    LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c "$length"
}

# Generate test cluster name
generate_test_cluster_name() {
    echo "test-cluster-$(random_string 6)"
}

# Generate test namespace name
generate_test_namespace() {
    echo "test-ns-$(random_string 6)"
}

# =============================================================================
# Export all functions
# =============================================================================

export -f color_green color_yellow color_red color_blue
export -f is_ci has_docker has_kubectl has_kind has_jq
export -f require_docker require_kubectl require_kind require_jq
export -f create_test_kind_cluster delete_test_kind_cluster wait_for_kind_cluster
export -f cleanup_docker_container docker_container_exists docker_container_running wait_for_container
export -f wait_for_namespace wait_for_pod wait_for_deployment
export -f create_temp_dir create_temp_file cleanup_temp_dir
export -f assert_file_exists assert_directory_exists assert_command_success assert_command_failure
export -f assert_contains assert_not_contains
export -f mock_kubectl mock_kind mock_docker restore_commands
export -f get_fixtures_dir get_fixture load_fixture
export -f log_info log_warn log_error
export -f register_cleanup run_cleanups
export -f random_string generate_test_cluster_name generate_test_namespace

