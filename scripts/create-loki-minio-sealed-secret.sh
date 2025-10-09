#!/bin/bash

# 🔐 Create Loki MinIO Sealed Secret
# This script creates a sealed secret for Loki MinIO credentials

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🔐 Creating Loki MinIO Sealed Secret${NC}"
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
echo -e "${YELLOW}📝 Enter Loki MinIO credentials:${NC}"
read -p "MinIO Root User [root-user]: " MINIO_ROOT_USER
MINIO_ROOT_USER=${MINIO_ROOT_USER:-root-user}

read -sp "MinIO Root Password: " MINIO_ROOT_PASSWORD
echo ""
if [ -z "$MINIO_ROOT_PASSWORD" ]; then
    echo -e "${RED}❌ Root Password is required${NC}"
    exit 1
fi

# Namespace
NAMESPACE="loki"

echo ""
echo -e "${YELLOW}📦 Creating sealed secret for namespace: ${NAMESPACE}${NC}"

# Create namespace if it doesn't exist
kubectl create namespace ${NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -

# Create the secret and seal it
kubectl create secret generic loki-minio-secret \
  --from-literal=root-user="${MINIO_ROOT_USER}" \
  --from-literal=root-password="${MINIO_ROOT_PASSWORD}" \
  --namespace="${NAMESPACE}" \
  --dry-run=client -o yaml | \
  kubeseal --format=yaml > ../flux/clusters/homelab/infrastructure/loki/loki-minio-secret-sealed.yaml

echo ""
echo -e "${GREEN}✅ Sealed secret created: ../flux/clusters/homelab/infrastructure/loki/loki-minio-secret-sealed.yaml${NC}"
echo ""
echo -e "${YELLOW}📋 Next steps:${NC}"
echo "1. Review the sealed secret file"
echo "2. Apply the sealed secret:"
echo "   kubectl apply -f ../flux/clusters/homelab/infrastructure/loki/loki-minio-secret-sealed.yaml"
echo "3. Update helmrelease.yaml to reference the sealed secret"
echo ""
echo -e "${GREEN}🎉 Done!${NC}"

