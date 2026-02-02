#!/bin/bash
# Script to run Agent-SRE LambdaFunction tests
# Creates test scenarios and monitors remediation

set -e

NAMESPACE="ai"
TEST_DIR="$(dirname "$0")"

echo "🧪 Agent-SRE LambdaFunction Test Suite"
echo "======================================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to print status
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check if namespace exists
if ! kubectl get namespace "$NAMESPACE" &>/dev/null; then
    print_error "Namespace $NAMESPACE does not exist"
    exit 1
fi

# Function to apply test manifests
apply_tests() {
    echo "📦 Applying test manifests..."
    kubectl apply -k "$TEST_DIR"
    print_status "Test manifests applied"
}

# Function to wait for test scenarios to trigger
wait_for_scenarios() {
    echo ""
    echo "⏳ Waiting for test scenarios to trigger..."
    sleep 10
    
    echo ""
    echo "📊 Test Pod Status:"
    kubectl get pods -n "$NAMESPACE" -l test.agent-sre.io/scenario
    
    echo ""
    echo "📊 Test Deployment Status:"
    kubectl get deployments -n "$NAMESPACE" -l test.agent-sre.io/scenario
}

# Function to check LambdaFunction status
check_lambdafunctions() {
    echo ""
    echo "🔧 LambdaFunction Status:"
    kubectl get lambdafunctions -n "$NAMESPACE" | grep -E "NAME|flux-reconcile|pod-restart|pod-check|check-pvc|scale-deployment" || echo "No LambdaFunctions found"
}

# Function to monitor agent-sre logs
monitor_agent_sre() {
    echo ""
    echo "📋 Monitoring Agent-SRE logs (last 20 lines)..."
    kubectl logs -n "$NAMESPACE" -l serving.knative.dev/service=agent-sre --tail=20 2>/dev/null || print_warning "Agent-SRE not running or scaled to zero"
}

# Function to clean up
cleanup() {
    echo ""
    echo "🧹 Cleaning up test resources..."
    kubectl delete -k "$TEST_DIR" --ignore-not-found=true
    print_status "Test resources cleaned up"
}

# Main menu
case "${1:-}" in
    apply)
        apply_tests
        wait_for_scenarios
        ;;
    status)
        wait_for_scenarios
        check_lambdafunctions
        monitor_agent_sre
        ;;
    monitor)
        echo "📊 Monitoring test scenarios (Ctrl+C to stop)..."
        while true; do
            clear
            echo "=== Agent-SRE Test Monitor ==="
            echo ""
            wait_for_scenarios
            check_lambdafunctions
            monitor_agent_sre
            sleep 5
        done
        ;;
    cleanup)
        cleanup
        ;;
    *)
        echo "Usage: $0 {apply|status|monitor|cleanup}"
        echo ""
        echo "Commands:"
        echo "  apply    - Apply test manifests"
        echo "  status   - Check test scenario status"
        echo "  monitor  - Continuously monitor test scenarios"
        echo "  cleanup  - Remove all test resources"
        exit 1
        ;;
esac

