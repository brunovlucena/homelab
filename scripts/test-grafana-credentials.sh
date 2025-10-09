#!/bin/bash

# 🔐 Grafana Credentials Test Script
# This script helps test the new Grafana admin credentials

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    print_error "kubectl is not installed or not in PATH"
    exit 1
fi

print_status "Testing Grafana sealed secrets implementation..."

# Check if the sealed secret exists
print_status "Checking if grafana-admin-secret sealed secret exists..."
if kubectl get sealedsecret grafana-admin-secret -n prometheus &> /dev/null; then
    print_success "✅ Sealed secret 'grafana-admin-secret' exists in prometheus namespace"
else
    print_error "❌ Sealed secret 'grafana-admin-secret' not found in prometheus namespace"
    exit 1
fi

# Check if the regular secret was created by the sealed secret controller
print_status "Checking if regular secret was created by sealed secret controller..."
if kubectl get secret grafana-admin-secret -n prometheus &> /dev/null; then
    print_success "✅ Regular secret 'grafana-admin-secret' exists (created by sealed secret controller)"
else
    print_warning "⚠️ Regular secret not yet created - sealed secret controller may still be processing"
fi

# Check Grafana pod status
print_status "Checking Grafana pod status..."
GRAFANA_POD=$(kubectl get pods -n prometheus -l app.kubernetes.io/name=grafana -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")
if [ -n "$GRAFANA_POD" ]; then
    POD_STATUS=$(kubectl get pod $GRAFANA_POD -n prometheus -o jsonpath='{.status.phase}')
    if [ "$POD_STATUS" = "Running" ]; then
        print_success "✅ Grafana pod '$GRAFANA_POD' is running"
    else
        print_warning "⚠️ Grafana pod '$GRAFANA_POD' status: $POD_STATUS"
    fi
else
    print_error "❌ No Grafana pod found"
fi

# Display connection information
print_status "Grafana connection information:"
echo "  🌐 Grafana URL: http://192.168.0.12:31080 (NodePort)"
echo "  👤 Username: admin"
echo "  🔐 Password: ovNE2KNaYhuqF2LWqoKkqexNGKWGKusHSriqZEMPv7k="
echo ""
print_warning "⚠️ IMPORTANT: Change this password after first login for security!"

# Test if Grafana is accessible
print_status "Testing Grafana accessibility..."
if curl -s -f http://192.168.0.12:31080/api/health > /dev/null 2>&1; then
    print_success "✅ Grafana is accessible at http://192.168.0.12:31080"
else
    print_warning "⚠️ Grafana may not be accessible yet - check pod status and service"
fi

print_success "🎉 Grafana sealed secrets implementation test completed!"
