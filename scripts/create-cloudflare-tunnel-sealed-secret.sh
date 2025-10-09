#!/bin/bash

# 🔐 Create Cloudflare Tunnel API Token Sealed Secret
# This script creates a sealed secret for Cloudflare Tunnel API credentials

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🔐 Creating Cloudflare Tunnel Sealed Secret${NC}"
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

# Prompt for Cloudflare credentials
echo -e "${YELLOW}📝 Enter Cloudflare Tunnel credentials:${NC}"
read -sp "🔑 Cloudflare API Token: " CLOUDFLARE_API_TOKEN
echo ""
if [ -z "$CLOUDFLARE_API_TOKEN" ]; then
    echo -e "${RED}❌ API Token is required${NC}"
    exit 1
fi

read -p "📋 Cloudflare Account ID: " CLOUDFLARE_ACCOUNT_ID
if [ -z "$CLOUDFLARE_ACCOUNT_ID" ]; then
    echo -e "${RED}❌ Account ID is required${NC}"
    exit 1
fi

# Namespace
NAMESPACE=${NAMESPACE:-default}

echo ""
echo -e "${YELLOW}📦 Creating sealed secret for namespace: ${NAMESPACE}${NC}"

# Create the secret and seal it
kubectl create secret generic cloudflare-tunnel-secret \
  --from-literal=api-token="${CLOUDFLARE_API_TOKEN}" \
  --from-literal=account-id="${CLOUDFLARE_ACCOUNT_ID}" \
  --namespace="${NAMESPACE}" \
  --dry-run=client -o yaml | \
  kubeseal --format=yaml > ../flux/clusters/homelab/infrastructure/cloudflare-tunnel/cloudflare-tunnel-secret-sealed.yaml

echo ""
echo -e "${GREEN}✅ Sealed secret created: ../flux/clusters/homelab/infrastructure/cloudflare-tunnel/cloudflare-tunnel-secret-sealed.yaml${NC}"
echo ""
echo -e "${YELLOW}📋 Next steps:${NC}"
echo "1. Review the sealed secret file"
echo "2. Apply the sealed secret:"
echo "   kubectl apply -f ../flux/clusters/homelab/infrastructure/cloudflare-tunnel/cloudflare-tunnel-secret-sealed.yaml"
echo "3. Update update-tunnel.py to read from environment variables"
echo ""
echo -e "${GREEN}🎉 Done!${NC}"

