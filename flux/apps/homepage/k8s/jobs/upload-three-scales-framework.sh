#!/bin/bash
# Upload three-scales-framework.png to MinIO using Kubernetes job
#
# Usage: ./upload-three-scales-framework.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NAMESPACE="homepage"
CONFIGMAP_NAME="three-scales-framework-image"
JOB_NAME="three-scales-framework-upload"
IMAGE_PATH="${SCRIPT_DIR}/../../storage/homepage-blog/images/graphs/three-scales-framework.png"

echo "🚀 Uploading three-scales-framework.png to MinIO"

# Check if image exists
if [ ! -f "$IMAGE_PATH" ]; then
    echo "❌ Image not found at: $IMAGE_PATH"
    exit 1
fi

echo "📋 Image found: $IMAGE_PATH ($(du -h "$IMAGE_PATH" | cut -f1))"

# Create ConfigMap from image file
echo "📦 Creating ConfigMap $CONFIGMAP_NAME..."
kubectl create configmap "$CONFIGMAP_NAME" \
  --from-file=three-scales-framework.png="$IMAGE_PATH" \
  --namespace="$NAMESPACE" \
  --dry-run=client -o yaml | kubectl apply -f -

echo "✅ ConfigMap created/updated"

# Apply job
echo "🚀 Applying upload job..."
kubectl apply -f "${SCRIPT_DIR}/three-scales-framework-upload-job.yaml"

# Wait for job to be created
echo "⏳ Waiting for job to start..."
kubectl wait --for=condition=Ready --timeout=30s job/$JOB_NAME -n "$NAMESPACE" || true

# Show logs
echo "📋 Job logs:"
kubectl logs -f job/$JOB_NAME -n "$NAMESPACE"

# Check job status
echo ""
echo "📊 Job status:"
kubectl get job/$JOB_NAME -n "$NAMESPACE"

