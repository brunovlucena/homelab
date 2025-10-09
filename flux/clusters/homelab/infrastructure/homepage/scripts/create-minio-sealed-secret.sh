#!/bin/bash

# 🔐 Create MinIO Sealed Secret for Bruno Site
# This script creates a sealed secret for MinIO credentials

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🔐 Creating MinIO Sealed Secret${NC}"
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

# Prompt for MinIO credentials
echo -e "${YELLOW}📝 Enter MinIO credentials:${NC}"
read -p "MinIO Access Key [minioadmin]: " MINIO_ACCESS_KEY
MINIO_ACCESS_KEY=${MINIO_ACCESS_KEY:-minioadmin}

read -sp "MinIO Secret Key [minioadmin]: " MINIO_SECRET_KEY
echo ""
MINIO_SECRET_KEY=${MINIO_SECRET_KEY:-minioadmin}

# Namespace
NAMESPACE=${NAMESPACE:-bruno}

echo ""
echo -e "${YELLOW}📦 Creating sealed secret for namespace: ${NAMESPACE}${NC}"

# Create the secret and seal it
kubectl create secret generic bruno-site-minio-secret \
  --from-literal=accessKey="${MINIO_ACCESS_KEY}" \
  --from-literal=secretKey="${MINIO_SECRET_KEY}" \
  --namespace="${NAMESPACE}" \
  --dry-run=client -o yaml | \
  kubeseal --format=yaml --cert=https://sealed-secrets.lucena.cloud/v1/cert.pem > ../k8s/bruno-site-minio-secret-sealed.yaml

echo ""
echo -e "${GREEN}✅ Sealed secret created: ../k8s/bruno-site-minio-secret-sealed.yaml${NC}"
echo ""
echo -e "${YELLOW}📋 Next steps:${NC}"
echo "1. Review the sealed secret file"
echo "2. Update k8s/secrets.yaml with the new sealed secret"
echo "3. Apply the sealed secret:"
echo "   kubectl apply -f ../k8s/bruno-site-minio-secret-sealed.yaml"
echo ""
echo -e "${GREEN}🎉 Done!${NC}"

