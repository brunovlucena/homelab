#!/bin/bash

# 🔧 Cloudflare Tunnel Route Updater
# Automatically updates Cloudflare tunnel published application routes with current service IPs

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
TUNNEL_NAME="homelab"
CLOUDFLARE_API_TOKEN="${CLOUDFLARE_API_TOKEN:-}"
CLOUDFLARE_ACCOUNT_ID="${CLOUDFLARE_ACCOUNT_ID:-}"

# Function to print colored output
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

# Check if required tools are installed
check_dependencies() {
    print_status "Checking dependencies..."
    
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl is not installed or not in PATH"
        exit 1
    fi
    
    if ! command -v curl &> /dev/null; then
        print_error "curl is not installed or not in PATH"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        print_error "jq is not installed or not in PATH"
        exit 1
    fi
    
    print_success "All dependencies are available"
}

# Check if environment variables are set
check_env() {
    if [[ -z "${CLOUDFLARE_API_TOKEN}" ]]; then
        print_error "CLOUDFLARE_API_TOKEN environment variable is not set"
        exit 1
    fi
    
    if [[ -z "${CLOUDFLARE_ACCOUNT_ID}" ]]; then
        print_error "CLOUDFLARE_ACCOUNT_ID environment variable is not set"
        exit 1
    fi
    
    print_success "Environment variables are set"
}

# Get current service IPs from Kubernetes
get_service_ips() {
    print_status "Getting current service IPs from Kubernetes..."
    
    # Get homepage-frontend service IP
    HOMEPAGE_IP=$(kubectl get svc homepage-frontend -n homepage -o jsonpath='{.spec.clusterIP}')
    HOMEPAGE_PORT=$(kubectl get svc homepage-frontend -n homepage -o jsonpath='{.spec.ports[0].port}')
    
    # Get grafana service IP
    GRAFANA_IP=$(kubectl get svc prometheus-operator-grafana -n prometheus -o jsonpath='{.spec.clusterIP}')
    GRAFANA_PORT=$(kubectl get svc prometheus-operator-grafana -n prometheus -o jsonpath='{.spec.ports[0].port}')
    
    # Get alertmanager service IP
    ALERTMANAGER_IP=$(kubectl get svc prometheus-operator-kube-p-alertmanager -n prometheus -o jsonpath='{.spec.clusterIP}')
    ALERTMANAGER_PORT=$(kubectl get svc prometheus-operator-kube-p-alertmanager -n prometheus -o jsonpath='{.spec.ports[0].port}')
    
    print_success "Service IPs retrieved:"
    echo "  📱 Homepage: ${HOMEPAGE_IP}:${HOMEPAGE_PORT}"
    echo "  📊 Grafana: ${GRAFANA_IP}:${GRAFANA_PORT}"
    echo "  🚨 Alertmanager: ${ALERTMANAGER_IP}:${ALERTMANAGER_PORT}"
}

# Get tunnel ID
get_tunnel_id() {
    print_status "Getting tunnel ID for '${TUNNEL_NAME}'..."
    
    TUNNEL_ID=$(curl -s -X GET \
        "https://api.cloudflare.com/client/v4/accounts/${CLOUDFLARE_ACCOUNT_ID}/cfd_tunnel" \
        -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
        -H "Content-Type: application/json" | \
        jq -r ".result[] | select(.name == \"${TUNNEL_NAME}\") | .id")
    
    if [[ -z "${TUNNEL_ID}" || "${TUNNEL_ID}" == "null" ]]; then
        print_error "Tunnel '${TUNNEL_NAME}' not found"
        exit 1
    fi
    
    print_success "Tunnel ID: ${TUNNEL_ID}"
}

# Get current tunnel configuration
get_tunnel_config() {
    print_status "Getting current tunnel configuration..."
    
    TUNNEL_CONFIG=$(curl -s -X GET \
        "https://api.cloudflare.com/client/v4/accounts/${CLOUDFLARE_ACCOUNT_ID}/cfd_tunnel/${TUNNEL_ID}/configurations" \
        -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
        -H "Content-Type: application/json")
    
    print_success "Tunnel configuration retrieved"
}

# Update tunnel configuration
update_tunnel_config() {
    print_status "Updating tunnel configuration with new service IPs..."
    
    # Create new configuration JSON
    NEW_CONFIG=$(cat <<EOF
{
  "config": {
    "ingress": [
      {
        "hostname": "lucena.cloud",
        "service": "http://${HOMEPAGE_IP}:${HOMEPAGE_PORT}"
      },
      {
        "hostname": "grafana.lucena.cloud",
        "service": "http://${GRAFANA_IP}:${GRAFANA_PORT}"
      },
      {
        "hostname": "alertmanager.lucena.cloud",
        "service": "http://${ALERTMANAGER_IP}:${ALERTMANAGER_PORT}"
      },
      {
        "hostname": "k8s-api.lucena.cloud",
        "service": "https://10.96.0.1:443"
      },
      {
        "service": "http_status:404"
      }
    ]
  }
}
EOF
)
    
    # Update the tunnel configuration
    RESPONSE=$(curl -s -X PUT \
        "https://api.cloudflare.com/client/v4/accounts/${CLOUDFLARE_ACCOUNT_ID}/cfd_tunnel/${TUNNEL_ID}/configurations" \
        -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
        -H "Content-Type: application/json" \
        -d "${NEW_CONFIG}")
    
    # Check if update was successful
    SUCCESS=$(echo "${RESPONSE}" | jq -r '.success')
    
    if [[ "${SUCCESS}" == "true" ]]; then
        print_success "Tunnel configuration updated successfully!"
        echo "Updated routes:"
        echo "  🌐 lucena.cloud → http://${HOMEPAGE_IP}:${HOMEPAGE_PORT}"
        echo "  📊 grafana.lucena.cloud → http://${GRAFANA_IP}:${GRAFANA_PORT}"
        echo "  🚨 alertmanager.lucena.cloud → http://${ALERTMANAGER_IP}:${ALERTMANAGER_PORT}"
    else
        print_error "Failed to update tunnel configuration"
        echo "${RESPONSE}" | jq -r '.errors[]?.message // "Unknown error"'
        exit 1
    fi
}

# Main execution
main() {
    echo "🚀 Cloudflare Tunnel Route Updater"
    echo "=================================="
    
    check_dependencies
    check_env
    get_service_ips
    get_tunnel_id
    get_tunnel_config
    update_tunnel_config
    
    print_success "All done! 🎉"
}

# Run main function
main "$@"
