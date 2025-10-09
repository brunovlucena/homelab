#!/bin/bash

# 🔧 Cloudflare Tunnel Token Regeneration Script
# This script helps regenerate and validate Cloudflare tunnel tokens

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
CLOUDFLARE_API_TOKEN=""
TUNNEL_NAME=""
CLOUDFLARE_ZONE_ID=""
DOMAIN=""

# Functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check dependencies
check_dependencies() {
    log_info "Checking dependencies..."
    
    if ! command -v curl &> /dev/null; then
        log_error "curl is required but not installed"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        log_error "jq is required but not installed"
        exit 1
    fi
    
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is required but not installed"
        exit 1
    fi
    
    log_success "Dependencies check completed"
}

# Get configuration from user
get_configuration() {
    log_info "Getting configuration..."
    
    if [ -z "$CLOUDFLARE_API_TOKEN" ]; then
        read -p "Enter your Cloudflare API token: " CLOUDFLARE_API_TOKEN
    fi
    
    if [ -z "$TUNNEL_NAME" ]; then
        read -p "Enter your tunnel name: " TUNNEL_NAME
    fi
    
    if [ -z "$CLOUDFLARE_ZONE_ID" ]; then
        read -p "Enter your Cloudflare Zone ID: " CLOUDFLARE_ZONE_ID
    fi
    
    if [ -z "$DOMAIN" ]; then
        read -p "Enter your domain (e.g., example.com): " DOMAIN
    fi
    
    log_success "Configuration collected"
}

# Test Cloudflare API connection
test_cloudflare_connection() {
    log_info "Testing Cloudflare API connection..."
    
    response=$(curl -s -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
        "https://api.cloudflare.com/client/v4/user/tokens/verify")
    
    if echo "$response" | jq -e '.success' | grep -q true; then
        log_success "Cloudflare API connection successful"
    else
        log_error "Failed to connect to Cloudflare API"
        log_error "Response: $response"
        exit 1
    fi
}

# Get account ID
get_account_id() {
    log_info "Getting account ID..."
    
    response=$(curl -s -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
        "https://api.cloudflare.com/client/v4/accounts")
    
    ACCOUNT_ID=$(echo "$response" | jq -r '.result[0].id')
    
    if [ "$ACCOUNT_ID" = "null" ] || [ -z "$ACCOUNT_ID" ]; then
        log_error "Failed to get account ID"
        exit 1
    fi
    
    log_success "Account ID: $ACCOUNT_ID"
}

# Delete existing tunnel (if exists)
delete_existing_tunnel() {
    log_info "Checking for existing tunnel..."
    
    response=$(curl -s -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
        "https://api.cloudflare.com/client/v4/accounts/$ACCOUNT_ID/cfd_tunnel?name=$TUNNEL_NAME")
    
    tunnel_id=$(echo "$response" | jq -r '.result[0].id // empty')
    
    if [ -n "$tunnel_id" ] && [ "$tunnel_id" != "null" ]; then
        log_warning "Found existing tunnel with ID: $tunnel_id"
        read -p "Do you want to delete the existing tunnel? (y/N): " -n 1 -r
        echo
        
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            log_info "Deleting existing tunnel..."
            
            delete_response=$(curl -s -X DELETE \
                -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
                "https://api.cloudflare.com/client/v4/accounts/$ACCOUNT_ID/cfd_tunnel/$tunnel_id")
            
            if echo "$delete_response" | jq -e '.success' | grep -q true; then
                log_success "Existing tunnel deleted"
            else
                log_error "Failed to delete existing tunnel"
                log_error "Response: $delete_response"
                exit 1
            fi
        else
            log_info "Keeping existing tunnel"
        fi
    else
        log_info "No existing tunnel found"
    fi
}

# Create new tunnel
create_tunnel() {
    log_info "Creating new tunnel..."
    
    response=$(curl -s -X POST \
        -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{\"name\": \"$TUNNEL_NAME\"}" \
        "https://api.cloudflare.com/client/v4/accounts/$ACCOUNT_ID/cfd_tunnel")
    
    if echo "$response" | jq -e '.success' | grep -q true; then
        TUNNEL_ID=$(echo "$response" | jq -r '.result.id')
        TUNNEL_TOKEN=$(echo "$response" | jq -r '.result.token')
        log_success "Tunnel created with ID: $TUNNEL_ID"
    else
        log_error "Failed to create tunnel"
        log_error "Response: $response"
        exit 1
    fi
}

# Create tunnel configuration
create_tunnel_config() {
    log_info "Creating tunnel configuration..."
    
    config_json=$(cat << EOF
{
  "tunnel": "$TUNNEL_ID",
  "credentials-file": "/etc/cloudflared/$TUNNEL_ID.json",
  "ingress": [
    {
      "hostname": "$DOMAIN",
      "service": "http://localhost:8080"
    },
    {
      "hostname": "api.$DOMAIN",
      "service": "http://localhost:8081"
    },
    {
      "service": "http_status:404"
    }
  ]
}
EOF
)
    
    # Create config file
    echo "$config_json" > "config.yaml"
    log_success "Tunnel configuration created: config.yaml"
}

# Create DNS records
create_dns_records() {
    log_info "Creating DNS records..."
    
    # Create CNAME for main domain
    curl -s -X POST "https://api.cloudflare.com/client/v4/zones/$CLOUDFLARE_ZONE_ID/dns_records" \
        -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"type\": \"CNAME\",
            \"name\": \"@\",
            \"content\": \"$TUNNEL_ID.cfargotunnel.com\",
            \"proxied\": true
        }" > /dev/null
    
    # Create CNAME for API subdomain
    curl -s -X POST "https://api.cloudflare.com/client/v4/zones/$CLOUDFLARE_ZONE_ID/dns_records" \
        -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"type\": \"CNAME\",
            \"name\": \"api\",
            \"content\": \"$TUNNEL_ID.cfargotunnel.com\",
            \"proxied\": true
        }" > /dev/null
    
    log_success "DNS records created"
}

# Update Kubernetes secret
update_k8s_secret() {
    log_info "Updating Kubernetes secret..."
    
    # Create new sealed secret (you'll need to seal it manually)
    cat > cloudflare-tunnel-credentials.yaml << EOF
apiVersion: v1
kind: Secret
metadata:
  name: cloudflare-tunnel-credentials
  namespace: cloudflare-tunnel
type: Opaque
data:
  CLOUDFLARE_TOKEN: $(echo -n "$TUNNEL_TOKEN" | base64)
EOF
    
    log_warning "Manual step required:"
    log_warning "1. Seal the secret using kubeseal"
    log_warning "2. Update the sealed-secrets/cloudflare.yaml file"
    log_warning "3. Apply the changes to your cluster"
    
    log_info "Secret template created: cloudflare-tunnel-credentials.yaml"
    log_info "Tunnel Token: $TUNNEL_TOKEN"
}

# Test tunnel connection
test_tunnel() {
    log_info "Testing tunnel connection..."
    
    if command -v cloudflared &> /dev/null; then
        log_info "Testing tunnel with cloudflared CLI..."
        timeout 10 cloudflared tunnel --token "$TUNNEL_TOKEN" run --no-autoupdate || true
    else
        log_warning "cloudflared CLI not found, skipping tunnel test"
    fi
}

# Generate summary
generate_summary() {
    log_info "Generating summary..."
    
    cat > tunnel-summary.md << EOF
# Cloudflare Tunnel Setup Summary

## Tunnel Information
- **Tunnel Name**: $TUNNEL_NAME
- **Tunnel ID**: $TUNNEL_ID
- **Domain**: $DOMAIN
- **Account ID**: $ACCOUNT_ID

## Files Created
- \`config.yaml\` - Tunnel configuration
- \`cloudflare-tunnel-credentials.yaml\` - Kubernetes secret template

## Next Steps
1. Seal the secret: \`kubeseal -f cloudflare-tunnel-credentials.yaml -w sealed-secret.yaml\`
2. Update \`sealed-secrets/cloudflare.yaml\` with the new sealed secret
3. Apply changes: \`kubectl apply -k .\`
4. Monitor tunnel status: \`kubectl logs -n cloudflare-tunnel deployment/cloudflared\`

## Troubleshooting
- Check tunnel logs: \`kubectl logs -n cloudflare-tunnel deployment/cloudflared\`
- Verify DNS records in Cloudflare dashboard
- Test tunnel connectivity: \`cloudflared tunnel --token $TUNNEL_TOKEN run\`
EOF
    
    log_success "Summary created: tunnel-summary.md"
}

# Main execution
main() {
    echo "🔧 Cloudflare Tunnel Token Regeneration Script"
    echo "=============================================="
    echo ""
    
    check_dependencies
    get_configuration
    test_cloudflare_connection
    get_account_id
    delete_existing_tunnel
    create_tunnel
    create_tunnel_config
    create_dns_records
    update_k8s_secret
    test_tunnel
    generate_summary
    
    echo ""
    echo "🎉 Tunnel regeneration completed!"
    echo ""
    echo "📋 Next steps:"
    echo "1. Seal the secret using kubeseal"
    echo "2. Update your sealed-secrets/cloudflare.yaml"
    echo "3. Apply the changes to your cluster"
    echo "4. Monitor the tunnel logs"
    echo ""
    echo "📚 Summary: ./tunnel-summary.md"
}

# Run main function
main "$@"

