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
    clear 2>/dev/null || true
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

# Show menu
show_menu() {
    log_header "📋 SECRET GENERATION MENU"
    echo ""
    echo "Select which secrets to generate:"
    echo ""
    echo -e "${BOLD}1)${NC} Generate ALL secrets (recommended)"
    echo -e "${BOLD}2)${NC} Loki MinIO Secret"
    echo -e "${BOLD}3)${NC} SRE Agent API Secret"
    echo -e "${BOLD}4)${NC} Grafana MCP API Secret"
    echo -e "${BOLD}5)${NC} Cloudflare Tunnel Secret"
    echo -e "${BOLD}6)${NC} Homepage Cloudflare Secret"
    echo -e "${BOLD}7)${NC} Homepage MinIO Secret"
    echo -e "${BOLD}8)${NC} Jamie Slack Bot Secret"
    echo -e "${BOLD}9)${NC} Exit"
    echo ""
}

# Create Loki MinIO Secret
create_loki_minio_secret() {
    log_header "🔐 Creating Loki MinIO Sealed Secret"
    
    local namespace="loki"
    local secret_name="loki-minio-secret"
    local output_file="$REPO_ROOT/flux/clusters/homelab/infrastructure/loki/loki-minio-secret-sealed.yaml"
    
    echo ""
    log_step "Enter Loki MinIO credentials:"
    read -p "MinIO Root User [loki-user]: " minio_user
    minio_user=${minio_user:-loki-user}
    
    read -sp "MinIO Root Password: " minio_password
    echo ""
    
    if [ -z "$minio_password" ]; then
        log_error "Password is required"
        FAILED_COUNT=$((FAILED_COUNT + 1))
        FAILED_SECRETS+=("Loki MinIO")
        return 1
    fi
    
    log_step "Creating namespace if needed..."
    kubectl create namespace ${namespace} --dry-run=client -o yaml | kubectl apply -f - &> /dev/null
    
    log_step "Generating sealed secret..."
    kubectl create secret generic ${secret_name} \
      --from-literal=root-user="${minio_user}" \
      --from-literal=root-password="${minio_password}" \
      --namespace="${namespace}" \
      --dry-run=client -o yaml | \
      kubeseal --format=yaml \
        --controller-name=sealed-secrets \
        --controller-namespace=flux-system > "${output_file}"
    
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
    log_step "Enter SRE Agent API credentials:"
    read -sp "SRE API Key: " sre_api_key
    echo ""
    
    if [ -z "$sre_api_key" ]; then
        log_error "SRE API Key is required"
        FAILED_COUNT=$((FAILED_COUNT + 1))
        FAILED_SECRETS+=("SRE Agent")
        return 1
    fi
    
    log_step "Creating namespace if needed..."
    kubectl create namespace ${namespace} --dry-run=client -o yaml | kubectl apply -f - &> /dev/null
    
    log_step "Generating sealed secret..."
    kubectl create secret generic ${secret_name} \
      --from-literal=sre-api-key="${sre_api_key}" \
      --namespace="${namespace}" \
      --dry-run=client -o yaml | \
      kubeseal --format=yaml \
        --controller-name=sealed-secrets \
        --controller-namespace=flux-system > "${output_file}"
    
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
    log_step "Enter Grafana MCP API credentials:"
    read -sp "Grafana API Key: " grafana_api_key
    echo ""
    
    if [ -z "$grafana_api_key" ]; then
        log_error "Grafana API Key is required"
        FAILED_COUNT=$((FAILED_COUNT + 1))
        FAILED_SECRETS+=("Grafana MCP")
        return 1
    fi
    
    log_step "Creating namespace if needed..."
    kubectl create namespace ${namespace} --dry-run=client -o yaml | kubectl apply -f - &> /dev/null
    
    log_step "Generating sealed secret..."
    kubectl create secret generic ${secret_name} \
      --from-literal=GRAFANA_API_KEY="${grafana_api_key}" \
      --namespace="${namespace}" \
      --dry-run=client -o yaml | \
      kubeseal --format=yaml \
        --controller-name=sealed-secrets \
        --controller-namespace=flux-system > "${output_file}"
    
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
    log_step "Enter Cloudflare Tunnel credentials:"
    read -sp "Cloudflare API Token: " cf_api_token
    echo ""
    
    if [ -z "$cf_api_token" ]; then
        log_error "API Token is required"
        FAILED_COUNT=$((FAILED_COUNT + 1))
        FAILED_SECRETS+=("Cloudflare Tunnel")
        return 1
    fi
    
    read -p "Cloudflare Account ID: " cf_account_id
    
    if [ -z "$cf_account_id" ]; then
        log_error "Account ID is required"
        FAILED_COUNT=$((FAILED_COUNT + 1))
        FAILED_SECRETS+=("Cloudflare Tunnel")
        return 1
    fi
    
    log_step "Generating sealed secret..."
    kubectl create secret generic ${secret_name} \
      --from-literal=api-token="${cf_api_token}" \
      --from-literal=account-id="${cf_account_id}" \
      --namespace="${namespace}" \
      --dry-run=client -o yaml | \
      kubeseal --format=yaml \
        --controller-name=sealed-secrets \
        --controller-namespace=flux-system > "${output_file}"
    
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
    log_step "Enter Homepage Cloudflare credentials:"
    read -p "Cloudflare Zone ID: " cf_zone_id
    
    if [ -z "$cf_zone_id" ]; then
        log_error "Zone ID is required"
        FAILED_COUNT=$((FAILED_COUNT + 1))
        FAILED_SECRETS+=("Homepage Cloudflare")
        return 1
    fi
    
    read -sp "Cloudflare API Token: " cf_api_token
    echo ""
    
    if [ -z "$cf_api_token" ]; then
        log_error "API Token is required"
        FAILED_COUNT=$((FAILED_COUNT + 1))
        FAILED_SECRETS+=("Homepage Cloudflare")
        return 1
    fi
    
    read -p "Domain [lucena.cloud]: " cf_domain
    cf_domain=${cf_domain:-lucena.cloud}
    
    read -p "Cache TTL in seconds [86400]: " cf_cache_ttl
    cf_cache_ttl=${cf_cache_ttl:-86400}
    
    log_step "Creating namespace if needed..."
    kubectl create namespace ${namespace} --dry-run=client -o yaml | kubectl apply -f - &> /dev/null
    
    log_step "Generating sealed secret..."
    kubectl create secret generic ${secret_name} \
      --from-literal=zone-id="${cf_zone_id}" \
      --from-literal=api-token="${cf_api_token}" \
      --from-literal=domain="${cf_domain}" \
      --from-literal=enabled="true" \
      --from-literal=cache-ttl="${cf_cache_ttl}" \
      --namespace="${namespace}" \
      --dry-run=client -o yaml | \
      kubeseal --format=yaml \
        --controller-name=sealed-secrets \
        --controller-namespace=flux-system > "${output_file}"
    
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
    log_step "Enter Homepage MinIO credentials:"
    read -p "MinIO Access Key [minioadmin]: " minio_access_key
    minio_access_key=${minio_access_key:-minioadmin}
    
    read -sp "MinIO Secret Key [minioadmin]: " minio_secret_key
    echo ""
    minio_secret_key=${minio_secret_key:-minioadmin}
    
    log_step "Creating namespace if needed..."
    kubectl create namespace ${namespace} --dry-run=client -o yaml | kubectl apply -f - &> /dev/null
    
    log_step "Generating sealed secret..."
    kubectl create secret generic ${secret_name} \
      --from-literal=accessKey="${minio_access_key}" \
      --from-literal=secretKey="${minio_secret_key}" \
      --namespace="${namespace}" \
      --dry-run=client -o yaml | \
      kubeseal --format=yaml \
        --controller-name=sealed-secrets \
        --controller-namespace=flux-system > "${output_file}"
    
    log_success "Created: ${output_file}"
    CREATED_COUNT=$((CREATED_COUNT + 1))
    CREATED_SECRETS+=("Homepage MinIO")
}

# Create Jamie Slack Bot Secret
create_jamie_slack_secret() {
    log_header "🤖 Creating Jamie Slack Bot Sealed Secret"
    
    local namespace="homepage"
    local secret_name="jamie-slack-secrets"
    local output_file="$REPO_ROOT/flux/clusters/homelab/infrastructure/jamie/k8s/jamie-slack-secrets-sealed.yaml"
    
    echo ""
    log_step "Enter Jamie Slack Bot credentials:"
    log_info "Hint: Check ~/.zshrc for SLACK_BOT_JAMIE_OAUTH_TOKEN, SLACK_APP_JAMIE_APP_TOKEN, and SLACK_SIGNING_SECRET"
    echo ""
    
    read -sp "Slack Bot Token (xoxb-...): " slack_bot_token
    echo ""
    
    if [ -z "$slack_bot_token" ]; then
        log_error "Slack Bot Token is required"
        FAILED_COUNT=$((FAILED_COUNT + 1))
        FAILED_SECRETS+=("Jamie Slack Bot")
        return 1
    fi
    
    read -sp "Slack App Token (xapp-...): " slack_app_token
    echo ""
    
    if [ -z "$slack_app_token" ]; then
        log_error "Slack App Token is required"
        FAILED_COUNT=$((FAILED_COUNT + 1))
        FAILED_SECRETS+=("Jamie Slack Bot")
        return 1
    fi
    
    read -sp "Slack Signing Secret: " slack_signing_secret
    echo ""
    
    if [ -z "$slack_signing_secret" ]; then
        log_error "Slack Signing Secret is required"
        FAILED_COUNT=$((FAILED_COUNT + 1))
        FAILED_SECRETS+=("Jamie Slack Bot")
        return 1
    fi
    
    log_step "Creating namespace if needed..."
    kubectl create namespace ${namespace} --dry-run=client -o yaml | kubectl apply -f - &> /dev/null
    
    log_step "Generating sealed secret..."
    kubectl create secret generic ${secret_name} \
      --from-literal=SLACK_BOT_TOKEN="${slack_bot_token}" \
      --from-literal=SLACK_APP_TOKEN="${slack_app_token}" \
      --from-literal=SLACK_SIGNING_SECRET="${slack_signing_secret}" \
      --namespace="${namespace}" \
      --dry-run=client -o yaml | \
      kubeseal --format=yaml \
        --controller-name=sealed-secrets \
        --controller-namespace=flux-system > "${output_file}"
    
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

# Main menu loop
main_menu() {
    while true; do
        show_menu
        read -p "Select an option [1-9]: " choice
        
        case $choice in
            1)
                generate_all_secrets
                break
                ;;
            2)
                create_loki_minio_secret || true
                echo ""
                read -p "Press Enter to continue..."
                clear 2>/dev/null || true
                ;;
            3)
                create_sre_agent_secret || true
                echo ""
                read -p "Press Enter to continue..."
                clear 2>/dev/null || true
                ;;
            4)
                create_grafana_mcp_secret || true
                echo ""
                read -p "Press Enter to continue..."
                clear 2>/dev/null || true
                ;;
            5)
                create_cloudflare_tunnel_secret || true
                echo ""
                read -p "Press Enter to continue..."
                clear 2>/dev/null || true
                ;;
            6)
                create_homepage_cloudflare_secret || true
                echo ""
                read -p "Press Enter to continue..."
                clear 2>/dev/null || true
                ;;
            7)
                create_homepage_minio_secret || true
                echo ""
                read -p "Press Enter to continue..."
                clear 2>/dev/null || true
                ;;
            8)
                create_jamie_slack_secret || true
                echo ""
                read -p "Press Enter to continue..."
                clear 2>/dev/null || true
                ;;
            9)
                log_info "Exiting..."
                exit 0
                ;;
            *)
                log_error "Invalid option. Please select 1-9."
                sleep 2
                clear 2>/dev/null || true
                ;;
        esac
    done
}

# Main execution
main() {
    show_banner
    check_dependencies
    main_menu
    generate_summary
    
    echo ""
    echo -e "${BOLD}${MAGENTA}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BOLD}${MAGENTA}            🔐 ONE SCRIPT TO RULE THEM ALL 🔐                  ${NC}"
    echo -e "${BOLD}${MAGENTA}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
}

# Run main function
main "$@"

