#!/bin/bash

# 🔐 Create SRE Agent API Key Sealed Secret
# This script creates a sealed secret for SRE Agent API key

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🔐 Creating SRE Agent Sealed Secret${NC}"
echo ""

# Check if kubeseal is installed
if ! command -v kubeseal &> /dev/null; then
    echo -e "${RED}❌ Error: kubeseal is not installed${NC}"
    echo "Please install kubeseal: https://github.com/bitnami-labs/sealed-secrets"
    exit 1
fi

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}❌ Error: kubectl is not installed${NC}"
    exit 1
fi

# Prompt for SRE Agent API key
echo -e "${YELLOW}📝 Enter SRE Agent API key:${NC}"
read -sp "🔑 SRE API Key: " SRE_API_KEY
echo ""
if [ -z "$SRE_API_KEY" ]; then
    echo -e "${RED}❌ SRE API Key is required${NC}"
    exit 1
fi

# Namespace
NAMESPACE="agent-sre"

echo ""
echo -e "${YELLOW}📦 Creating sealed secret for namespace: ${NAMESPACE}${NC}"

# Create namespace if it doesn't exist
kubectl create namespace ${NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -

# Create the secret and seal it
kubectl create secret generic agent-sre-secret \
  --from-literal=sre-api-key="${SRE_API_KEY}" \
  --namespace="${NAMESPACE}" \
  --dry-run=client -o yaml | \
  kubeseal --format=yaml > ../flux/clusters/homelab/infrastructure/agent-sre/agent-sre-secret-sealed.yaml

echo ""
echo -e "${GREEN}✅ Sealed secret created: ../flux/clusters/homelab/infrastructure/agent-sre/agent-sre-secret-sealed.yaml${NC}"
echo ""
echo -e "${YELLOW}📋 Next steps:${NC}"
echo "1. Review the sealed secret file"
echo "2. Apply the sealed secret:"
echo "   kubectl apply -f ../flux/clusters/homelab/infrastructure/agent-sre/agent-sre-secret-sealed.yaml"
echo "3. Update mcp_config.json or deployment to use the secret"
echo ""
echo -e "${GREEN}🎉 Done!${NC}"

