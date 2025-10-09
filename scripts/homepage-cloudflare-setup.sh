#!/bin/bash

# 🛡️ Cloudflare Setup Script for Bruno Site
# This script helps automate Cloudflare configuration

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
DOMAIN=""
CLOUDFLARE_API_TOKEN=""
CLOUDFLARE_ZONE_ID=""
SERVER_IP=""

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
        log_warning "jq is not installed. Install it for better JSON parsing"
    fi
    
    log_success "Dependencies check completed"
}

# Get configuration from user
get_configuration() {
    log_info "Getting configuration..."
    
    if [ -z "$DOMAIN" ]; then
        read -p "Enter your domain (e.g., brunosite.com): " DOMAIN
    fi
    
    if [ -z "$SERVER_IP" ]; then
        read -p "Enter your server IP address: " SERVER_IP
    fi
    
    if [ -z "$CLOUDFLARE_API_TOKEN" ]; then
        read -p "Enter your Cloudflare API token: " CLOUDFLARE_API_TOKEN
    fi
    
    if [ -z "$CLOUDFLARE_ZONE_ID" ]; then
        read -p "Enter your Cloudflare Zone ID: " CLOUDFLARE_ZONE_ID
    fi
    
    log_success "Configuration collected"
}

# Validate configuration
validate_configuration() {
    log_info "Validating configuration..."
    
    if [ -z "$DOMAIN" ] || [ -z "$SERVER_IP" ] || [ -z "$CLOUDFLARE_API_TOKEN" ] || [ -z "$CLOUDFLARE_ZONE_ID" ]; then
        log_error "All configuration values are required"
        exit 1
    fi
    
    # Validate IP format
    if ! [[ $SERVER_IP =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        log_error "Invalid IP address format"
        exit 1
    fi
    
    log_success "Configuration validated"
}

# Test Cloudflare API connection
test_cloudflare_connection() {
    log_info "Testing Cloudflare API connection..."
    
    response=$(curl -s -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
        "https://api.cloudflare.com/client/v4/zones/$CLOUDFLARE_ZONE_ID")
    
    if echo "$response" | grep -q '"success":true'; then
        log_success "Cloudflare API connection successful"
    else
        log_error "Failed to connect to Cloudflare API"
        log_error "Response: $response"
        exit 1
    fi
}

# Create DNS records
create_dns_records() {
    log_info "Creating DNS records..."
    
    # Create A record for main domain
    curl -s -X POST "https://api.cloudflare.com/client/v4/zones/$CLOUDFLARE_ZONE_ID/dns_records" \
        -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"type\": \"A\",
            \"name\": \"@\",
            \"content\": \"$SERVER_IP\",
            \"proxied\": true
        }" > /dev/null
    
    # Create A record for API subdomain
    curl -s -X POST "https://api.cloudflare.com/client/v4/zones/$CLOUDFLARE_ZONE_ID/dns_records" \
        -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"type\": \"A\",
            \"name\": \"api\",
            \"content\": \"$SERVER_IP\",
            \"proxied\": true
        }" > /dev/null
    
    # Create CNAME record for www
    curl -s -X POST "https://api.cloudflare.com/client/v4/zones/$CLOUDFLARE_ZONE_ID/dns_records" \
        -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"type\": \"CNAME\",
            \"name\": \"www\",
            \"content\": \"$DOMAIN\",
            \"proxied\": true
        }" > /dev/null
    
    log_success "DNS records created"
}

# Configure SSL/TLS
configure_ssl() {
    log_info "Configuring SSL/TLS..."
    
    # Set SSL mode to Full (strict)
    curl -s -X PATCH "https://api.cloudflare.com/client/v4/zones/$CLOUDFLARE_ZONE_ID/settings/ssl" \
        -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"value": "full_strict"}' > /dev/null
    
    # Enable TLS 1.3
    curl -s -X PATCH "https://api.cloudflare.com/client/v4/zones/$CLOUDFLARE_ZONE_ID/settings/tls_1_3" \
        -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"value": "on"}' > /dev/null
    
    # Set minimum TLS version to 1.2
    curl -s -X PATCH "https://api.cloudflare.com/client/v4/zones/$CLOUDFLARE_ZONE_ID/settings/min_tls_version" \
        -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"value": "1.2"}' > /dev/null
    
    log_success "SSL/TLS configured"
}

# Configure security headers
configure_security_headers() {
    log_info "Configuring security headers..."
    
    # Enable security headers
    curl -s -X PATCH "https://api.cloudflare.com/client/v4/zones/$CLOUDFLARE_ZONE_ID/settings/security_header" \
        -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "value": {
                "strict_transport_security": {
                    "enabled": true,
                    "max_age": 31536000,
                    "include_subdomains": true,
                    "preload": true
                },
                "x_content_type_options": "nosniff",
                "x_frame_options": "DENY",
                "x_xss_protection": "1; mode=block",
                "referrer_policy": "strict-origin-when-cross-origin"
            }
        }' > /dev/null
    
    log_success "Security headers configured"
}

# Configure caching
configure_caching() {
    log_info "Configuring caching..."
    
    # Set browser cache TTL
    curl -s -X PATCH "https://api.cloudflare.com/client/v4/zones/$CLOUDFLARE_ZONE_ID/settings/browser_cache_ttl" \
        -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"value": 14400}' > /dev/null
    
    # Enable auto minify
    curl -s -X PATCH "https://api.cloudflare.com/client/v4/zones/$CLOUDFLARE_ZONE_ID/settings/minify" \
        -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"value": {"css": "on", "html": "on", "js": "on"}}' > /dev/null
    
    # Enable Brotli compression
    curl -s -X PATCH "https://api.cloudflare.com/client/v4/zones/$CLOUDFLARE_ZONE_ID/settings/brotli" \
        -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"value": "on"}' > /dev/null
    
    log_success "Caching configured"
}

# Test the setup
test_setup() {
    log_info "Testing setup..."
    
    # Test main domain
    if curl -s -I "https://$DOMAIN" | grep -q "HTTP/2 200\|HTTP/1.1 200"; then
        log_success "Main domain is accessible"
    else
        log_warning "Main domain may not be accessible yet (DNS propagation)"
    fi
    
    # Test API subdomain
    if curl -s -I "https://api.$DOMAIN" | grep -q "HTTP/2 200\|HTTP/1.1 200"; then
        log_success "API subdomain is accessible"
    else
        log_warning "API subdomain may not be accessible yet (DNS propagation)"
    fi
    
    # Check SSL certificate
    if curl -s -I "https://$DOMAIN" | grep -q "cloudflare"; then
        log_success "Cloudflare proxy is active"
    else
        log_warning "Cloudflare proxy may not be active yet"
    fi
}

# Generate environment file
generate_env_file() {
    log_info "Generating environment file..."
    
    cat > .env.cloudflare << EOF
# Cloudflare Configuration
CLOUDFLARE_DOMAIN=$DOMAIN
CLOUDFLARE_API_TOKEN=$CLOUDFLARE_API_TOKEN
CLOUDFLARE_ZONE_ID=$CLOUDFLARE_ZONE_ID
CLOUDFLARE_SERVER_IP=$SERVER_IP

# DNS Records
CLOUDFLARE_DNS_MAIN=$DOMAIN
CLOUDFLARE_DNS_API=api.$DOMAIN
CLOUDFLARE_DNS_WWW=www.$DOMAIN
EOF
    
    log_success "Environment file generated: .env.cloudflare"
}

# Main execution
main() {
    echo "🛡️  Cloudflare Setup Script for Bruno Site"
    echo "=============================================="
    echo ""
    
    check_dependencies
    get_configuration
    validate_configuration
    test_cloudflare_connection
    create_dns_records
    configure_ssl
    configure_security_headers
    configure_caching
    generate_env_file
    test_setup
    
    echo ""
    echo "🎉 Cloudflare setup completed!"
    echo ""
    echo "📋 Next steps:"
    echo "1. Wait for DNS propagation (up to 24 hours)"
    echo "2. Configure Page Rules in Cloudflare dashboard"
    echo "3. Set up monitoring and alerts"
    echo "4. Test your site thoroughly"
    echo ""
    echo "📚 Documentation: ./CLOUDFLARE_SETUP_GUIDE.md"
    echo "🔧 Environment file: .env.cloudflare"
}

# Run main function
main "$@"
