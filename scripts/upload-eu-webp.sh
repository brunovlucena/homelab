#!/bin/bash

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🚀 Uploading eu.webp to MinIO...${NC}"

# Check if MinIO is ready
echo -e "${YELLOW}⏳ Waiting for MinIO to be ready...${NC}"
until kubectl get pods -n minio -l app=minio --field-selector=status.phase=Running | grep -q minio; do
    echo "MinIO not ready yet, waiting..."
    sleep 5
done

echo -e "${GREEN}✅ MinIO is ready!${NC}"

# Get MinIO credentials from secret
MINIO_ROOT_USER=$(kubectl get secret -n minio minio-secret -o jsonpath='{.data.MINIO_ROOT_USER}' | base64 -d)
MINIO_ROOT_PASSWORD=$(kubectl get secret -n minio minio-secret -o jsonpath='{.data.MINIO_ROOT_PASSWORD}' | base64 -d)

# Path to the eu.webp file
EU_WEBP_PATH="../flux/clusters/homelab/infrastructure/homepage/frontend/public/assets/eu.webp"

# Check if file exists
if [[ ! -f "$EU_WEBP_PATH" ]]; then
    echo -e "${RED}❌ Error: eu.webp file not found at $EU_WEBP_PATH${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Found eu.webp file${NC}"

# Create temporary configmap with the file
kubectl create configmap eu-webp-temp --from-file=eu.webp="$EU_WEBP_PATH" --namespace=minio --dry-run=client -o yaml | kubectl apply -f -

# Run upload in a pod with the configmap mounted
kubectl run minio-uploader --rm -i --restart=Never --namespace=minio \
    --image=minio/mc:latest \
    --env="MINIO_ROOT_USER=$MINIO_ROOT_USER" \
    --env="MINIO_ROOT_PASSWORD=$MINIO_ROOT_PASSWORD" \
    --overrides='{
        "spec": {
            "containers": [{
                "name": "uploader",
                "image": "minio/mc:latest",
                "command": ["/bin/sh", "-c"],
                "args": ["mc alias set myminio http://minio-service:9000 $MINIO_ROOT_USER $MINIO_ROOT_PASSWORD && mc mb myminio/homepage-assets --ignore-existing && mc anonymous set public myminio/homepage-assets && mc cp /assets/eu.webp myminio/homepage-assets/eu.webp && echo Upload completed"],
                "env": [
                    {"name": "MINIO_ROOT_USER", "value": "'"$MINIO_ROOT_USER"'"},
                    {"name": "MINIO_ROOT_PASSWORD", "value": "'"$MINIO_ROOT_PASSWORD"'"}
                ],
                "volumeMounts": [{
                    "name": "eu-webp",
                    "mountPath": "/assets",
                    "readOnly": true
                }]
            }],
            "volumes": [{
                "name": "eu-webp",
                "configMap": {
                    "name": "eu-webp-temp"
                }
            }]
        }
    }'

# Clean up temporary configmap
kubectl delete configmap eu-webp-temp --namespace=minio --ignore-not-found=true

echo -e "${GREEN}🎉 Upload completed successfully!${NC}"
