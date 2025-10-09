#!/bin/bash

# 🔐 ONE SCRIPT TO RULE THEM ALL - Master Secret Generator
# This script generates ALL sealed secrets for the homelab infrastructure
# Run this once and you're done!

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SEALED_SECRETS_DIR="$REPO_ROOT/flux/clusters/homelab/infrastructure/sealed-secrets"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Stats tracking
CREATED_COUNT=0
FAILED_COUNT=0
SKIPPED_COUNT=0

# Arrays to track results
declare -a CREATED_SECRETS=()
declare -a FAILED_SECRETS=()
declare -a SKIPPED_SECRETS=()

# Functions
log_header() {
    echo ""
    echo -e "${BOLD}${MAGENTA}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BOLD}${MAGENTA}$1${NC}"
    echo -e "${BOLD}${MAGENTA}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

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

log_step() {
    echo -e "${CYAN}🔧 $1${NC}"
}

# Banner
show_banner() {
    echo -e "${BOLD}${MAGENTA}"
    cat << 'EOF'
╔═══════════════════════════════════════════════════════════════╗
║                                                               ║
║   🔐 ONE SCRIPT TO RULE THEM ALL 🔐                          ║
║                                                               ║
║   Master Secret Generator for Homelab Infrastructure         ║
║                                                               ║
╚═══════════════════════════════════════════════════════════════╝
EOF
    echo -e "${NC}"
}

# Check all dependencies
check_dependencies() {
    log_header "🔍 CHECKING DEPENDENCIES"
    
    local all_good=true
    
    # Check kubectl
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed"
        all_good=false
    else
        log_success "kubectl found"
    fi
    
    # Check kubeseal
    if ! command -v kubeseal &> /dev/null; then
        log_error "kubeseal is not installed"
        echo "Install with: brew install kubeseal"
        all_good=false
    else
        log_success "kubeseal found"
    fi
    
    # Check cluster connection
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Cannot connect to Kubernetes cluster"
        all_good=false
    else
        log_success "Connected to Kubernetes cluster"
    fi
    
    # Check sealed-secrets controller
    if ! kubectl get deployment sealed-secrets -n flux-system &> /dev/null 2>&1; then
        log_warning "Sealed Secrets controller not found in flux-system namespace"
        log_info "Will attempt to use it anyway..."
    else
        log_success "Sealed Secrets controller is running"
    fi
    
    if [ "$all_good" = false ]; then
        log_error "Please fix the above issues before continuing"
        exit 1
    fi
    
    echo ""
    log_success "All dependencies satisfied!"
}

# Removed interactive menu - script now generates all secrets automatically

# Create Loki MinIO Secret
create_loki_minio_secret() {
    log_header "🔐 Creating Loki MinIO Sealed Secret"
    
    local namespace="loki"
    local secret_name="loki-minio-secret"
    local output_file="$REPO_ROOT/flux/clusters/homelab/infrastructure/loki/loki-minio-secret-sealed.yaml"
    
    echo ""
    # Read from environment variables first
    local minio_user="${LOKI_MINIO_USER:-loki-user}"
    local minio_password="${LOKI_MINIO_PASSWORD}"
    
    if [ -z "$minio_password" ]; then
        log_error "LOKI_MINIO_PASSWORD environment variable is not set"
        log_info "Set it in ~/.zshrc: export LOKI_MINIO_PASSWORD='your-password'"
        FAILED_COUNT=$((FAILED_COUNT + 1))
        FAILED_SECRETS+=("Loki MinIO")
        return 1
    fi
    
    log_success "Using credentials from environment variables"
    
    log_step "Creating namespace if needed..."
    kubectl create namespace ${namespace} --dry-run=client -o yaml | kubectl apply -f - &> /dev/null
    
    log_step "Fetching sealed-secrets certificate..."
    local cert_file="/tmp/sealed-secrets-cert.pem"
    kubectl get secret -n flux-system -l sealedsecrets.bitnami.com/sealed-secrets-key=active -o jsonpath='{.items[0].data.tls\.crt}' | base64 -d > "${cert_file}" 2>/dev/null || true
    
    log_step "Generating sealed secret..."
    kubectl create secret generic ${secret_name} \
      --from-literal=root-user="${minio_user}" \
      --from-literal=root-password="${minio_password}" \
      --namespace="${namespace}" \
      --dry-run=client -o yaml | \
      kubeseal --format=yaml --cert="${cert_file}" > "${output_file}"
    
    log_success "Created: ${output_file}"
    CREATED_COUNT=$((CREATED_COUNT + 1))
    CREATED_SECRETS+=("Loki MinIO")
}

# Create SRE Agent Secret
create_sre_agent_secret() {
    log_header "🔐 Creating SRE Agent Sealed Secret"
    
    local namespace="agent-sre"
    local secret_name="agent-sre-secret"
    local output_file="$REPO_ROOT/flux/clusters/homelab/infrastructure/agent-sre/agent-sre-secret-sealed.yaml"
    
    echo ""
    # Read from environment variable
    local sre_api_key="${SRE_API_KEY}"
    
    if [ -z "$sre_api_key" ]; then
        log_error "SRE_API_KEY environment variable is not set"
        log_info "Set it in ~/.zshrc: export SRE_API_KEY='your-api-key'"
        FAILED_COUNT=$((FAILED_COUNT + 1))
        FAILED_SECRETS+=("SRE Agent")
        return 1
    fi
    
    log_success "Using credentials from environment variables"
    
    log_step "Creating namespace if needed..."
    kubectl create namespace ${namespace} --dry-run=client -o yaml | kubectl apply -f - &> /dev/null
    
    log_step "Fetching sealed-secrets certificate..."
    local cert_file="/tmp/sealed-secrets-cert.pem"
    kubectl get secret -n flux-system -l sealedsecrets.bitnami.com/sealed-secrets-key=active -o jsonpath='{.items[0].data.tls\.crt}' | base64 -d > "${cert_file}" 2>/dev/null || true
    
    log_step "Generating sealed secret..."
    kubectl create secret generic ${secret_name} \
      --from-literal=sre-api-key="${sre_api_key}" \
      --namespace="${namespace}" \
      --dry-run=client -o yaml | \
      kubeseal --format=yaml --cert="${cert_file}" > "${output_file}"
    
    log_success "Created: ${output_file}"
    CREATED_COUNT=$((CREATED_COUNT + 1))
    CREATED_SECRETS+=("SRE Agent")
}

# Create Grafana MCP Secret
create_grafana_mcp_secret() {
    log_header "🔐 Creating Grafana MCP Sealed Secret"
    
    local namespace="grafana-mcp"
    local secret_name="grafana-mcp-secrets"
    local output_file="$REPO_ROOT/flux/clusters/homelab/infrastructure/grafana-mcp/grafana-mcp-secret-sealed.yaml"
    
    echo ""
    # Read from environment variable
    local grafana_api_key="${GRAFANA_API_KEY}"
    
    if [ -z "$grafana_api_key" ]; then
        log_error "GRAFANA_API_KEY environment variable is not set"
        log_info "Set it in ~/.zshrc: export GRAFANA_API_KEY='your-api-key'"
        FAILED_COUNT=$((FAILED_COUNT + 1))
        FAILED_SECRETS+=("Grafana MCP")
        return 1
    fi
    
    log_success "Using credentials from environment variables"
    
    log_step "Creating namespace if needed..."
    kubectl create namespace ${namespace} --dry-run=client -o yaml | kubectl apply -f - &> /dev/null
    
    log_step "Fetching sealed-secrets certificate..."
    local cert_file="/tmp/sealed-secrets-cert.pem"
    kubectl get secret -n flux-system -l sealedsecrets.bitnami.com/sealed-secrets-key=active -o jsonpath='{.items[0].data.tls\.crt}' | base64 -d > "${cert_file}" 2>/dev/null || true
    
    log_step "Generating sealed secret..."
    kubectl create secret generic ${secret_name} \
      --from-literal=GRAFANA_API_KEY="${grafana_api_key}" \
      --namespace="${namespace}" \
      --dry-run=client -o yaml | \
      kubeseal --format=yaml --cert="${cert_file}" > "${output_file}"
    
    log_success "Created: ${output_file}"
    CREATED_COUNT=$((CREATED_COUNT + 1))
    CREATED_SECRETS+=("Grafana MCP")
}

# Create Cloudflare Tunnel Secret
create_cloudflare_tunnel_secret() {
    log_header "🔐 Creating Cloudflare Tunnel Sealed Secret"
    
    local namespace="default"
    local secret_name="cloudflare-tunnel-secret"
    local output_file="$REPO_ROOT/flux/clusters/homelab/infrastructure/cloudflare-tunnel/cloudflare-tunnel-secret-sealed.yaml"
    
    echo ""
    # Read from environment variables
    local cf_api_token="${CLOUDFLARE_API_TOKEN}"
    local cf_account_id="${CLOUDFLARE_ACCOUNT_ID}"
    
    if [ -z "$cf_api_token" ]; then
        log_error "CLOUDFLARE_API_TOKEN environment variable is not set"
        log_info "Set it in ~/.zshrc: export CLOUDFLARE_API_TOKEN='your-api-token'"
        FAILED_COUNT=$((FAILED_COUNT + 1))
        FAILED_SECRETS+=("Cloudflare Tunnel")
        return 1
    fi
    
    if [ -z "$cf_account_id" ]; then
        log_error "CLOUDFLARE_ACCOUNT_ID environment variable is not set"
        log_info "Set it in ~/.zshrc: export CLOUDFLARE_ACCOUNT_ID='your-account-id'"
        FAILED_COUNT=$((FAILED_COUNT + 1))
        FAILED_SECRETS+=("Cloudflare Tunnel")
        return 1
    fi
    
    log_success "Using credentials from environment variables"
    
    log_step "Fetching sealed-secrets certificate..."
    local cert_file="/tmp/sealed-secrets-cert.pem"
    kubectl get secret -n flux-system -l sealedsecrets.bitnami.com/sealed-secrets-key=active -o jsonpath='{.items[0].data.tls\.crt}' | base64 -d > "${cert_file}" 2>/dev/null || true
    
    log_step "Generating sealed secret..."
    kubectl create secret generic ${secret_name} \
      --from-literal=api-token="${cf_api_token}" \
      --from-literal=account-id="${cf_account_id}" \
      --namespace="${namespace}" \
      --dry-run=client -o yaml | \
      kubeseal --format=yaml --cert="${cert_file}" > "${output_file}"
    
    log_success "Created: ${output_file}"
    CREATED_COUNT=$((CREATED_COUNT + 1))
    CREATED_SECRETS+=("Cloudflare Tunnel")
}

# Create Homepage Cloudflare Secret
create_homepage_cloudflare_secret() {
    log_header "🔐 Creating Homepage Cloudflare Sealed Secret"
    
    local namespace="bruno"
    local secret_name="bruno-site-cloudflare-secret"
    local output_file="$REPO_ROOT/flux/clusters/homelab/infrastructure/homepage/k8s/bruno-site-cloudflare-secret-sealed.yaml"
    
    echo ""
    # Read from environment variables
    local cf_zone_id="${CLOUDFLARE_ZONE_ID}"
    local cf_api_token="${CLOUDFLARE_API_TOKEN}"
    local cf_domain="${CLOUDFLARE_DOMAIN:-lucena.cloud}"
    local cf_cache_ttl="${CLOUDFLARE_CACHE_TTL:-86400}"
    
    if [ -z "$cf_zone_id" ]; then
        log_error "CLOUDFLARE_ZONE_ID environment variable is not set"
        log_info "Set it in ~/.zshrc: export CLOUDFLARE_ZONE_ID='your-zone-id'"
        FAILED_COUNT=$((FAILED_COUNT + 1))
        FAILED_SECRETS+=("Homepage Cloudflare")
        return 1
    fi
    
    if [ -z "$cf_api_token" ]; then
        log_error "CLOUDFLARE_API_TOKEN environment variable is not set"
        log_info "Set it in ~/.zshrc: export CLOUDFLARE_API_TOKEN='your-api-token'"
        FAILED_COUNT=$((FAILED_COUNT + 1))
        FAILED_SECRETS+=("Homepage Cloudflare")
        return 1
    fi
    
    log_success "Using credentials from environment variables"
    
    log_step "Creating namespace if needed..."
    kubectl create namespace ${namespace} --dry-run=client -o yaml | kubectl apply -f - &> /dev/null
    
    log_step "Fetching sealed-secrets certificate..."
    local cert_file="/tmp/sealed-secrets-cert.pem"
    kubectl get secret -n flux-system -l sealedsecrets.bitnami.com/sealed-secrets-key=active -o jsonpath='{.items[0].data.tls\.crt}' | base64 -d > "${cert_file}" 2>/dev/null || true
    
    log_step "Generating sealed secret..."
    kubectl create secret generic ${secret_name} \
      --from-literal=zone-id="${cf_zone_id}" \
      --from-literal=api-token="${cf_api_token}" \
      --from-literal=domain="${cf_domain}" \
      --from-literal=enabled="true" \
      --from-literal=cache-ttl="${cf_cache_ttl}" \
      --namespace="${namespace}" \
      --dry-run=client -o yaml | \
      kubeseal --format=yaml --cert="${cert_file}" > "${output_file}"
    
    log_success "Created: ${output_file}"
    CREATED_COUNT=$((CREATED_COUNT + 1))
    CREATED_SECRETS+=("Homepage Cloudflare")
}

# Create Homepage MinIO Secret
create_homepage_minio_secret() {
    log_header "🔐 Creating Homepage MinIO Sealed Secret"
    
    local namespace="bruno"
    local secret_name="bruno-site-minio-secret"
    local output_file="$REPO_ROOT/flux/clusters/homelab/infrastructure/homepage/k8s/bruno-site-minio-secret-sealed.yaml"
    
    echo ""
    # Read from environment variables
    local minio_access_key="${MINIO_ACCESS_KEY:-minioadmin}"
    local minio_secret_key="${MINIO_SECRET_KEY:-minioadmin}"
    
    log_success "Using credentials from environment variables (or defaults)"
    
    log_step "Creating namespace if needed..."
    kubectl create namespace ${namespace} --dry-run=client -o yaml | kubectl apply -f - &> /dev/null
    
    log_step "Fetching sealed-secrets certificate..."
    local cert_file="/tmp/sealed-secrets-cert.pem"
    kubectl get secret -n flux-system -l sealedsecrets.bitnami.com/sealed-secrets-key=active -o jsonpath='{.items[0].data.tls\.crt}' | base64 -d > "${cert_file}" 2>/dev/null || true
    
    log_step "Generating sealed secret..."
    kubectl create secret generic ${secret_name} \
      --from-literal=accessKey="${minio_access_key}" \
      --from-literal=secretKey="${minio_secret_key}" \
      --namespace="${namespace}" \
      --dry-run=client -o yaml | \
      kubeseal --format=yaml --cert="${cert_file}" > "${output_file}"
    
    log_success "Created: ${output_file}"
    CREATED_COUNT=$((CREATED_COUNT + 1))
    CREATED_SECRETS+=("Homepage MinIO")
}

# Create Jamie Slack Bot Secret
create_jamie_slack_secret() {
    log_header "🤖 Creating Jamie Slack Bot Sealed Secret"
    
    local namespace="jamie"
    local secret_name="jamie-slack-secrets"
    local output_file="$REPO_ROOT/flux/clusters/homelab/infrastructure/jamie/jamie-slack-secrets-sealed.yaml"
    
    echo ""
    # Read from environment variables
    local slack_bot_token="${SLACK_BOT_JAMIE_OAUTH_TOKEN}"
    local slack_app_token="${SLACK_APP_JAMIE_APP_TOKEN}"
    local slack_signing_secret="${SLACK_SIGNING_SECRET}"
    
    if [ -z "$slack_bot_token" ]; then
        log_error "SLACK_BOT_JAMIE_OAUTH_TOKEN environment variable is not set"
        log_info "Set it in ~/.zshrc: export SLACK_BOT_JAMIE_OAUTH_TOKEN='xoxb-...'"
        FAILED_COUNT=$((FAILED_COUNT + 1))
        FAILED_SECRETS+=("Jamie Slack Bot")
        return 1
    fi
    
    if [ -z "$slack_app_token" ]; then
        log_error "SLACK_APP_JAMIE_APP_TOKEN environment variable is not set"
        log_info "Set it in ~/.zshrc: export SLACK_APP_JAMIE_APP_TOKEN='xapp-...'"
        FAILED_COUNT=$((FAILED_COUNT + 1))
        FAILED_SECRETS+=("Jamie Slack Bot")
        return 1
    fi
    
    if [ -z "$slack_signing_secret" ]; then
        log_error "SLACK_SIGNING_SECRET environment variable is not set"
        log_info "Set it in ~/.zshrc: export SLACK_SIGNING_SECRET='your-signing-secret'"
        FAILED_COUNT=$((FAILED_COUNT + 1))
        FAILED_SECRETS+=("Jamie Slack Bot")
        return 1
    fi
    
    log_success "Using credentials from environment variables"
    
    log_step "Creating namespace if needed..."
    kubectl create namespace ${namespace} --dry-run=client -o yaml | kubectl apply -f - &> /dev/null
    
    log_step "Fetching sealed-secrets certificate..."
    local cert_file="/tmp/sealed-secrets-cert.pem"
    kubectl get secret -n flux-system -l sealedsecrets.bitnami.com/sealed-secrets-key=active -o jsonpath='{.items[0].data.tls\.crt}' | base64 -d > "${cert_file}" 2>/dev/null || true
    
    log_step "Generating sealed secret..."
    kubectl create secret generic ${secret_name} \
      --from-literal=SLACK_BOT_TOKEN="${slack_bot_token}" \
      --from-literal=SLACK_APP_TOKEN="${slack_app_token}" \
      --from-literal=SLACK_SIGNING_SECRET="${slack_signing_secret}" \
      --namespace="${namespace}" \
      --dry-run=client -o yaml | \
      kubeseal --format=yaml --cert="${cert_file}" > "${output_file}"
    
    log_success "Created: ${output_file}"
    CREATED_COUNT=$((CREATED_COUNT + 1))
    CREATED_SECRETS+=("Jamie Slack Bot")
}

# Generate summary
generate_summary() {
    log_header "📊 GENERATION SUMMARY"
    
    echo ""
    echo -e "${BOLD}Statistics:${NC}"
    echo -e "  ${GREEN}✅ Created: ${CREATED_COUNT}${NC}"
    echo -e "  ${RED}❌ Failed: ${FAILED_COUNT}${NC}"
    echo -e "  ${YELLOW}⏭️  Skipped: ${SKIPPED_COUNT}${NC}"
    
    if [ ${#CREATED_SECRETS[@]} -gt 0 ]; then
        echo ""
        echo -e "${BOLD}${GREEN}Successfully Created:${NC}"
        for secret in "${CREATED_SECRETS[@]}"; do
            echo -e "  ${GREEN}✅${NC} $secret"
        done
    fi
    
    if [ ${#FAILED_SECRETS[@]} -gt 0 ]; then
        echo ""
        echo -e "${BOLD}${RED}Failed:${NC}"
        for secret in "${FAILED_SECRETS[@]}"; do
            echo -e "  ${RED}❌${NC} $secret"
        done
    fi
    
    if [ ${#SKIPPED_SECRETS[@]} -gt 0 ]; then
        echo ""
        echo -e "${BOLD}${YELLOW}Skipped:${NC}"
        for secret in "${SKIPPED_SECRETS[@]}"; do
            echo -e "  ${YELLOW}⏭️${NC} $secret"
        done
    fi
    
    echo ""
    log_header "📋 NEXT STEPS"
    echo ""
    echo "1. Review the generated sealed secret files"
    echo "2. Commit the sealed secrets to Git (they're encrypted!)"
    echo "3. Push to your repository"
    echo "4. Flux will automatically reconcile and create the secrets"
    echo ""
    echo "To verify secrets were created:"
    echo "  kubectl get secrets -A | grep -E '(loki|agent-sre|grafana-mcp|cloudflare|bruno|jamie)'"
    echo ""
    echo "To view Flux reconciliation:"
    echo "  flux logs --all-namespaces --follow"
    echo ""
    
    if [ $CREATED_COUNT -gt 0 ]; then
        log_success "🎉 Secret generation completed!"
    fi
}

# Generate all secrets
generate_all_secrets() {
    log_header "🚀 GENERATING ALL SECRETS"
    echo ""
    log_info "This will create all sealed secrets for your homelab"
    echo ""
    
    # Create each secret, but don't stop on failure
    create_loki_minio_secret || true
    echo ""
    
    create_sre_agent_secret || true
    echo ""
    
    create_grafana_mcp_secret || true
    echo ""
    
    create_cloudflare_tunnel_secret || true
    echo ""
    
    create_homepage_cloudflare_secret || true
    echo ""
    
    create_homepage_minio_secret || true
    echo ""
    
    create_jamie_slack_secret || true
    echo ""
}

# Removed main_menu - script now runs non-interactively

# Main execution
main() {
    show_banner
    check_dependencies
    generate_all_secrets
    generate_summary
    
    echo ""
    echo -e "${BOLD}${MAGENTA}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BOLD}${MAGENTA}            🔐 ONE SCRIPT TO RULE THEM ALL 🔐                  ${NC}"
    echo -e "${BOLD}${MAGENTA}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
}

# Run main function
main "$@"

