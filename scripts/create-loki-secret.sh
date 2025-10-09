#!/bin/bash
set -e

# Script to create the loki-minio-secret SealedSecret
# Run from anywhere in the homelab repo

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SEALED_SECRETS_DIR="$REPO_ROOT/flux/clusters/homelab/infrastructure/sealed-secrets"

# Ensure the sealed-secrets directory exists
if [ ! -d "$SEALED_SECRETS_DIR" ]; then
    echo "Error: sealed-secrets directory not found at $SEALED_SECRETS_DIR"
    exit 1
fi

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}🔐 Creating Loki MinIO SealedSecret${NC}"
echo

# Check if kubeseal is installed
if ! command -v kubeseal &> /dev/null; then
    echo -e "${RED}❌ kubeseal is not installed${NC}"
    echo "Install it with: brew install kubeseal"
    exit 1
fi

# Check if kubectl is configured
if ! kubectl cluster-info &> /dev/null; then
    echo -e "${RED}❌ kubectl is not configured or cluster is not reachable${NC}"
    exit 1
fi

# Default password (same as used in notifi repo)
DEFAULT_PASSWORD="supersecretpassword"

# Prompt for password
echo -e "${YELLOW}Enter the MinIO root password${NC}"
echo -e "${YELLOW}(Press Enter to use default: supersecretpassword)${NC}"
read -s -p "Password: " PASSWORD
echo

if [ -z "$PASSWORD" ]; then
    PASSWORD="$DEFAULT_PASSWORD"
    echo -e "${GREEN}Using default password${NC}"
fi

# Create temporary files
TMP_SECRET="/tmp/loki-minio-secret-$$.yaml"
TMP_SEALED="/tmp/loki-minio-sealed-$$.yaml"

# Create the secret
echo -e "${YELLOW}📝 Creating secret...${NC}"
kubectl create secret generic loki-minio-secret \
  --from-literal=root-password="$PASSWORD" \
  --namespace=loki \
  --dry-run=client -o yaml > "$TMP_SECRET"

# Seal the secret
echo -e "${YELLOW}🔒 Sealing secret...${NC}"
kubeseal --format=yaml \
  --controller-name=sealed-secrets \
  --controller-namespace=flux-system \
  < "$TMP_SECRET" > "$TMP_SEALED"

# Move to final location
OUTPUT_FILE="$SEALED_SECRETS_DIR/loki-minio-secret.yaml"
mv "$TMP_SEALED" "$OUTPUT_FILE"
rm "$TMP_SECRET"

echo -e "${GREEN}✅ Created $OUTPUT_FILE${NC}"
echo
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Review the file: cat $OUTPUT_FILE"
echo "2. Add to kustomization.yaml: add '- loki-minio-secret.yaml' to the resources list"
echo "3. Commit and push the changes"
echo "4. Flux will automatically reconcile and create the secret"
echo
echo -e "${GREEN}Done!${NC}"

