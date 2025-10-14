#!/bin/bash

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ZSHRC_FILE="$HOME/.zshrc"
NAMESPACE_PROMETHEUS="prometheus"
NAMESPACE_AGENT_SRE="agent-sre"
NAMESPACE_JAMIE="jamie"
NAMESPACE_LOKI="loki"
NAMESPACE_BRUNO="bruno"
NAMESPACE_ALLOY="alloy"
NAMESPACE_MINIO="minio"
NAMESPACE_CLOUDFLARE_TUNNEL="cloudflare-tunnel"

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}❌ Error: kubectl is not installed${NC}"
    exit 1
fi

# Check if .zshrc exists
if [[ ! -f "${ZSHRC_FILE}" ]]; then
    echo -e "${RED}❌ Error: .zshrc file not found at ${ZSHRC_FILE}${NC}"
    exit 1
fi

echo -e "${BLUE}🔐 Creating Kubernetes Secrets from .zshrc environment variables${NC}"
echo ""

# ============================================================================
# Helper function to extract environment variable value
# ============================================================================
extract_env_value() {
    local var_name="$1"
    local value=$(grep "^export ${var_name}=" "${ZSHRC_FILE}" | sed 's/^export [^=]*=//g' | sed 's/"//g' | sed "s/'//g")
    echo "${value}"
}

# ============================================================================
# Helper function to ensure namespace exists
# ============================================================================
ensure_namespace() {
    local namespace="$1"
    if ! kubectl get namespace "${namespace}" &> /dev/null; then
        echo -e "${YELLOW}📦 Creating namespace: ${namespace}${NC}"
        kubectl create namespace "${namespace}"
        echo -e "${GREEN}✅ Namespace created: ${namespace}${NC}"
    else
        echo -e "${BLUE}ℹ️  Namespace already exists: ${namespace}${NC}"
    fi
}

# ============================================================================
# Create all required namespaces
# ============================================================================
echo -e "${BLUE}📦 Ensuring all required namespaces exist...${NC}"
ensure_namespace "${NAMESPACE_PROMETHEUS}"
ensure_namespace "${NAMESPACE_AGENT_SRE}"
ensure_namespace "${NAMESPACE_JAMIE}"
ensure_namespace "${NAMESPACE_LOKI}"
ensure_namespace "${NAMESPACE_BRUNO}"
ensure_namespace "${NAMESPACE_ALLOY}"
ensure_namespace "${NAMESPACE_MINIO}"
ensure_namespace "${NAMESPACE_CLOUDFLARE_TUNNEL}"
echo ""

# ============================================================================
# 1. Create ghcr-secret for jamie, bruno, and agent-sre namespaces
# ============================================================================
echo -e "${YELLOW}🔐 Creating ghcr-secret for multiple namespaces...${NC}"

# Extract GHCR credentials
GHCR_USERNAME=$(extract_env_value "GHCR_USERNAME")
GHCR_TOKEN=$(extract_env_value "GHCR_TOKEN")

# Create ghcr-secret for jamie namespace
echo "Creating ghcr-secret in jamie namespace..."
kubectl create secret docker-registry ghcr-secret \
  --namespace="${NAMESPACE_JAMIE}" \
  --docker-server=ghcr.io \
  --docker-username="${GHCR_USERNAME}" \
  --docker-password="${GHCR_TOKEN}" \
  --dry-run=client -o yaml | kubectl apply -f -
echo -e "${GREEN}✅ Secret created in cluster: ghcr-secret (${NAMESPACE_JAMIE})${NC}"

# Create ghcr-secret for bruno namespace
echo "Creating ghcr-secret in bruno namespace..."
kubectl create secret docker-registry ghcr-secret \
  --namespace="${NAMESPACE_BRUNO}" \
  --docker-server=ghcr.io \
  --docker-username="${GHCR_USERNAME}" \
  --docker-password="${GHCR_TOKEN}" \
  --dry-run=client -o yaml | kubectl apply -f -
echo -e "${GREEN}✅ Secret created in cluster: ghcr-secret (${NAMESPACE_BRUNO})${NC}"

# Create ghcr-secret for agent-sre namespace
echo "Creating ghcr-secret in agent-sre namespace..."
kubectl create secret docker-registry ghcr-secret \
  --namespace="${NAMESPACE_AGENT_SRE}" \
  --docker-server=ghcr.io \
  --docker-username="${GHCR_USERNAME}" \
  --docker-password="${GHCR_TOKEN}" \
  --dry-run=client -o yaml | kubectl apply -f -
echo -e "${GREEN}✅ Secret created in cluster: ghcr-secret (${NAMESPACE_AGENT_SRE})${NC}"

# ============================================================================
# 2. Create prometheus-secrets (consolidated: grafana, pagerduty, slack, strava)
# ============================================================================
echo -e "${YELLOW}📊 Creating prometheus-secrets...${NC}"

# Extract values
# GRAFANA_SERVICE_ACCOUNT_TOKEN=$(extract_env_value "GRAFANA_SERVICE_ACCOUNT_TOKEN")
GRAFANA_PASSWORD=$(extract_env_value "GRAFANA_PASSWORD")
PAGERDUTY_SERVICE_KEY=$(extract_env_value "PAGERDUTY_SERVICE_KEY")
PAGERDUTY_URL=$(extract_env_value "PAGERDUTY_URL")
SLACK_APP_JAMIE_APP_TOKEN=$(extract_env_value "SLACK_APP_JAMIE_APP_TOKEN")
SLACK_BOT_ALERTMANAGER_APP_TOKEN=$(extract_env_value "SLACK_BOT_ALERTMANAGER_APP_TOKEN")
SLACK_BOT_ALERTMANAGER_OAUTH_TOKEN=$(extract_env_value "SLACK_BOT_ALERTMANAGER_OAUTH_TOKEN")
SLACK_BOT_JAMIE_OAUTH_TOKEN=$(extract_env_value "SLACK_BOT_JAMIE_OAUTH_TOKEN")
SLACK_SIGNING_SECRET=$(extract_env_value "SLACK_SIGNING_SECRET")
SLACK_WEBHOOK_URL=$(extract_env_value "SLACK_WEBHOOK_URL")
STRAVA_ACCESS_TOKEN=$(extract_env_value "STRAVA_ACCESS_TOKEN")
STRAVA_CLIENT_ID=$(extract_env_value "STRAVA_CLIENT_ID")
STRAVA_CLIENT_SECRET=$(extract_env_value "STRAVA_CLIENT_SECRET")
STRAVA_REFRESH_TOKEN=$(extract_env_value "STRAVA_REFRESH_TOKEN")

# Create the secret in the cluster
kubectl create secret generic prometheus-secrets \
  --namespace="${NAMESPACE_PROMETHEUS}" \
  --from-literal=admin-username="admin" \
  --from-literal=admin-password="${GRAFANA_PASSWORD}" \
  --from-literal=GRAFANA_PASSWORD="${GRAFANA_PASSWORD}" \
  --from-literal=PAGERDUTY_SERVICE_KEY="${PAGERDUTY_SERVICE_KEY}" \
  --from-literal=PAGERDUTY_URL="${PAGERDUTY_URL}" \
  --from-literal=pagerduty-service-key="${PAGERDUTY_SERVICE_KEY}" \
  --from-literal=slack-webhook-url="${SLACK_WEBHOOK_URL}" \
  --from-literal=SLACK_APP_JAMIE_APP_TOKEN="${SLACK_APP_JAMIE_APP_TOKEN}" \
  --from-literal=SLACK_BOT_ALERTMANAGER_APP_TOKEN="${SLACK_BOT_ALERTMANAGER_APP_TOKEN}" \
  --from-literal=SLACK_BOT_ALERTMANAGER_OAUTH_TOKEN="${SLACK_BOT_ALERTMANAGER_OAUTH_TOKEN}" \
  --from-literal=SLACK_BOT_JAMIE_OAUTH_TOKEN="${SLACK_BOT_JAMIE_OAUTH_TOKEN}" \
  --from-literal=SLACK_SIGNING_SECRET="${SLACK_SIGNING_SECRET}" \
  --from-literal=SLACK_WEBHOOK_URL="${SLACK_WEBHOOK_URL}" \
  --from-literal=STRAVA_ACCESS_TOKEN="${STRAVA_ACCESS_TOKEN}" \
  --from-literal=STRAVA_CLIENT_ID="${STRAVA_CLIENT_ID}" \
  --from-literal=STRAVA_CLIENT_SECRET="${STRAVA_CLIENT_SECRET}" \
  --from-literal=STRAVA_REFRESH_TOKEN="${STRAVA_REFRESH_TOKEN}" \
  --dry-run=client -o yaml | kubectl apply -f -
echo -e "${GREEN}✅ Secret created in cluster: prometheus-secrets (${NAMESPACE_PROMETHEUS})${NC}"

# ============================================================================
# 3. Create agent-sre-secrets (consolidated: ai-secrets)
# ============================================================================
echo -e "${YELLOW}🤖 Creating agent-sre-secrets...${NC}"

# Extract values
GITHUB_PERSONAL_ACCESS_TOKEN=$(extract_env_value "GITHUB_PERSONAL_ACCESS_TOKEN")
GITHUB_TOKEN=$(extract_env_value "GITHUB_TOKEN")
GITHUB_TOKEN_BRUNO=$(extract_env_value "GITHUB_TOKEN_BRUNO")
GITHUB_TOKEN_NOTIFI=$(extract_env_value "GITHUB_TOKEN_NOTIFI")
GITHUB_USERNAME=$(extract_env_value "GITHUB_USERNAME")
GITHUB_USERNAME_NOTIFI=$(extract_env_value "GITHUB_USERNAME_NOTIFI")
HF_TOKEN=$(extract_env_value "HF_TOKEN")
HUGGINGFACE_HUB_TOKEN=$(extract_env_value "HUGGINGFACE_HUB_TOKEN")
LANGSMITH_API_KEY=$(extract_env_value "LANGSMITH_API_KEY")
LANGSMITH_ENDPOINT=$(extract_env_value "LANGSMITH_ENDPOINT")
LANGSMITH_OTEL_ENABLED=$(extract_env_value "LANGSMITH_OTEL_ENABLED")
LANGSMITH_PROJECT=$(extract_env_value "LANGSMITH_PROJECT")
LANGSMITH_TRACING=$(extract_env_value "LANGSMITH_TRACING")
LOGFIRE_TOKEN=$(extract_env_value "LOGFIRE_TOKEN")
PAGERDUTY_SERVICE_KEY=$(extract_env_value "PAGERDUTY_SERVICE_KEY")
PAGERDUTY_URL=$(extract_env_value "PAGERDUTY_URL")
TWINGATE_ACCESS_TOKEN=$(extract_env_value "TWINGATE_ACCESS_TOKEN")
TWINGATE_NETWORK=$(extract_env_value "TWINGATE_NETWORK")
TWINGATE_REFRESH_TOKEN=$(extract_env_value "TWINGATE_REFRESH_TOKEN")

# Create the secret in the cluster
kubectl create secret generic agent-sre-secrets \
  --namespace="${NAMESPACE_AGENT_SRE}" \
  --from-literal=GITHUB_PERSONAL_ACCESS_TOKEN="${GITHUB_PERSONAL_ACCESS_TOKEN}" \
  --from-literal=GITHUB_TOKEN="${GITHUB_TOKEN}" \
  --from-literal=GITHUB_TOKEN_BRUNO="${GITHUB_TOKEN_BRUNO}" \
  --from-literal=GITHUB_TOKEN_NOTIFI="${GITHUB_TOKEN_NOTIFI}" \
  --from-literal=GITHUB_USERNAME="${GITHUB_USERNAME}" \
  --from-literal=GITHUB_USERNAME_NOTIFI="${GITHUB_USERNAME_NOTIFI}" \
  --from-literal=HF_TOKEN="${HF_TOKEN}" \
  --from-literal=HUGGINGFACE_HUB_TOKEN="${HUGGINGFACE_HUB_TOKEN}" \
  --from-literal=LANGSMITH_API_KEY="${LANGSMITH_API_KEY}" \
  --from-literal=LANGSMITH_ENDPOINT="${LANGSMITH_ENDPOINT}" \
  --from-literal=LANGSMITH_OTEL_ENABLED="${LANGSMITH_OTEL_ENABLED}" \
  --from-literal=LANGSMITH_PROJECT="${LANGSMITH_PROJECT}" \
  --from-literal=LANGSMITH_TRACING="${LANGSMITH_TRACING}" \
  --from-literal=LOGFIRE_TOKEN="${LOGFIRE_TOKEN}" \
  --from-literal=PAGERDUTY_SERVICE_KEY="${PAGERDUTY_SERVICE_KEY}" \
  --from-literal=PAGERDUTY_URL="${PAGERDUTY_URL}" \
  --dry-run=client -o yaml | kubectl apply -f -
echo -e "${GREEN}✅ Secret created in cluster: agent-sre-secrets (${NAMESPACE_AGENT_SRE})${NC}"

# ============================================================================
# 4. Create jamie-secrets (consolidated: jamie-slack-secrets)
# ============================================================================
echo -e "${YELLOW}🎵 Creating jamie-secrets...${NC}"

# Extract values (LOGFIRE_TOKEN already extracted in agent-sre section)
SLACK_APP_JAMIE_APP_TOKEN=$(extract_env_value "SLACK_APP_JAMIE_APP_TOKEN")
SLACK_BOT_JAMIE_OAUTH_TOKEN=$(extract_env_value "SLACK_BOT_JAMIE_OAUTH_TOKEN")
SLACK_SIGNING_SECRET=$(extract_env_value "SLACK_SIGNING_SECRET")

# Create the secret in the cluster
kubectl create secret generic jamie-secrets \
  --namespace="${NAMESPACE_JAMIE}" \
  --from-literal=LOGFIRE_TOKEN="${LOGFIRE_TOKEN}" \
  --from-literal=SLACK_APP_TOKEN="${SLACK_APP_JAMIE_APP_TOKEN}" \
  --from-literal=SLACK_BOT_TOKEN="${SLACK_BOT_JAMIE_OAUTH_TOKEN}" \
  --from-literal=SLACK_SIGNING_SECRET="${SLACK_SIGNING_SECRET}" \
  --dry-run=client -o yaml | kubectl apply -f -
echo -e "${GREEN}✅ Secret created in cluster: jamie-secrets (${NAMESPACE_JAMIE})${NC}"

# ============================================================================
# 4.6. Create alloy-secrets (for Alloy OTLP → Logfire)
# ============================================================================
echo -e "${YELLOW}📊 Creating alloy-secrets...${NC}"

# Create the secret in the cluster
kubectl create secret generic alloy-secrets \
  --namespace="alloy" \
  --from-literal=LOGFIRE_TOKEN="${LOGFIRE_TOKEN}" \
  --dry-run=client -o yaml | kubectl apply -f -
echo -e "${GREEN}✅ Secret created in cluster: alloy-secrets (alloy)${NC}"

# ============================================================================
# 4.7. Create bruno-site-secret
# ============================================================================
echo -e "${YELLOW}🐘 Creating bruno-site-secret...${NC}"

# Extract values
# Use correct service names from their respective namespaces
POSTGRES_HOST="postgres-postgresql.postgres.svc.cluster.local"
POSTGRES_PORT="5432"
POSTGRES_DB=$(extract_env_value "POSTGRES_DB")
POSTGRES_USER="postgres"
POSTGRES_PASSWORD=$(extract_env_value "POSTGRES_PASSWORD")
REDIS_PASSWORD=$(extract_env_value "REDIS_PASSWORD")
HOMEPAGE_MINIO_ACCESS_KEY=$(extract_env_value "HOMEPAGE_MINIO_ACCESS_KEY")
HOMEPAGE_MINIO_SECRET_KEY=$(extract_env_value "HOMEPAGE_MINIO_SECRET_KEY")

# Create the secret in the cluster with individual fields (no hardcoded database-url)
# The application constructs the DATABASE_URL programmatically from these components
kubectl create secret generic bruno-site-secret \
  --namespace="${NAMESPACE_BRUNO}" \
  --from-literal=LOGFIRE_TOKEN="${LOGFIRE_TOKEN}" \
  --from-literal=POSTGRES_HOST="${POSTGRES_HOST}" \
  --from-literal=POSTGRES_PORT="${POSTGRES_PORT}" \
  --from-literal=POSTGRES_DB="${POSTGRES_DB}" \
  --from-literal=POSTGRES_USER="${POSTGRES_USER}" \
  --from-literal=POSTGRES_PASSWORD="${POSTGRES_PASSWORD}" \
  --from-literal=MINIO_ACCESS_KEY="${HOMEPAGE_MINIO_ACCESS_KEY}" \
  --from-literal=MINIO_SECRET_KEY="${HOMEPAGE_MINIO_SECRET_KEY}" \
  --from-literal=REDIS_PASSWORD="${REDIS_PASSWORD}" \
  --dry-run=client -o yaml | kubectl apply -f -
echo -e "${GREEN}✅ Secret created in cluster: bruno-site-secret (${NAMESPACE_BRUNO})${NC}"

# ============================================================================
# 5. Create loki-minio-secret
# ============================================================================
echo -e "${YELLOW}📝 Creating loki-minio-secret...${NC}"

# Extract values
LOKI_MINIO_ROOT_PASSWORD=$(extract_env_value "LOKI_MINIO_ROOT_PASSWORD")

# Create the secret in the cluster
kubectl create secret generic loki-minio-secret \
  --namespace="${NAMESPACE_LOKI}" \
  --from-literal=root-password="${LOKI_MINIO_ROOT_PASSWORD}" \
  --dry-run=client -o yaml | kubectl apply -f -
echo -e "${GREEN}✅ Secret created in cluster: loki-minio-secret (${NAMESPACE_LOKI})${NC}"

# ============================================================================
# 6. Create minio-secret
# ============================================================================
echo -e "${YELLOW}🗄️  Creating minio-secret...${NC}"

# Extract values
MINIO_ROOT_USER=$(extract_env_value "MINIO_ROOT_USER")
MINIO_ROOT_PASSWORD=$(extract_env_value "MINIO_ROOT_PASSWORD")

# Create the secret in the cluster
kubectl create secret generic minio-secret \
  --namespace="minio" \
  --from-literal=MINIO_ROOT_USER="${MINIO_ROOT_USER}" \
  --from-literal=MINIO_ROOT_PASSWORD="${MINIO_ROOT_PASSWORD}" \
  --dry-run=client -o yaml | kubectl apply -f -
echo -e "${GREEN}✅ Secret created in cluster: minio-secret (minio)${NC}"

# ============================================================================
# 7. Create cloudflare-tunnel-credentials
# ============================================================================
echo -e "${YELLOW}☁️  Creating cloudflare-tunnel-credentials...${NC}"

# Extract values
CLOUDFLARE_TOKEN=$(extract_env_value "CLOUDFLARE_TOKEN")

# Create the secret in the cluster
kubectl create secret generic cloudflare-tunnel-credentials \
  --namespace="cloudflare-tunnel" \
  --from-literal=CLOUDFLARE_TOKEN="${CLOUDFLARE_TOKEN}" \
  --from-literal=tunnel-token="${CLOUDFLARE_TOKEN}" \
  --dry-run=client -o yaml | kubectl apply -f -
echo -e "${GREEN}✅ Secret created in cluster: cloudflare-tunnel-credentials (cloudflare-tunnel)${NC}"

# ============================================================================
# Summary
# ============================================================================
echo ""
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✅ All secrets created successfully in the cluster!${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${YELLOW}📝 Created secrets in cluster:${NC}"
echo "  1. ghcr-secret (${NAMESPACE_JAMIE} namespace)"
echo "     └─ GitHub Container Registry credentials"
echo "  2. ghcr-secret (${NAMESPACE_BRUNO} namespace)"
echo "     └─ GitHub Container Registry credentials"
echo "  3. ghcr-secret (${NAMESPACE_AGENT_SRE} namespace)"
echo "     └─ GitHub Container Registry credentials"
echo "  4. prometheus-secrets (${NAMESPACE_PROMETHEUS} namespace)"
echo "     └─ Grafana, PagerDuty, Slack, Strava credentials"
echo "  5. agent-sre-secrets (${NAMESPACE_AGENT_SRE} namespace)"
echo "     └─ GitHub, HuggingFace, LangSmith, Logfire (LOGFIRE_TOKEN), PagerDuty, Twingate credentials"
echo "  6. jamie-secrets (${NAMESPACE_JAMIE} namespace)"
echo "     └─ Slack and Logfire (LOGFIRE_TOKEN) credentials"
echo "  7. bruno-site-secret (${NAMESPACE_BRUNO} namespace)"
echo "     └─ PostgreSQL, Redis, MinIO, Logfire, and OTEL endpoint credentials"
echo "  8. alloy-secrets (alloy namespace)"
echo "     └─ Logfire (LOGFIRE_TOKEN) for Alloy → Logfire forwarding"
echo "  9. loki-minio-secret (${NAMESPACE_LOKI} namespace)"
echo "     └─ Loki MinIO credentials"
echo "  10. minio-secret (minio namespace)"
echo "     └─ MinIO root credentials"
echo "  11. cloudflare-tunnel-credentials (cloudflare-tunnel namespace)"
echo "     └─ Cloudflare Tunnel token"
echo ""
