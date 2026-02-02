#!/bin/bash
# Setup and run three-scales-framework image generation and upload
#
# This script follows DevOps best practices:
# 1. Creates ConfigMap with Python script (not binary image - avoids size limits)
# 2. Applies the generate+upload job
# 3. Monitors job progress
#
# Usage: ./setup-three-scales-upload.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NAMESPACE="homepage"
CONFIGMAP_NAME="three-scales-generator"
JOB_NAME="three-scales-generate-upload"
SCRIPT_PATH="${SCRIPT_DIR}/../../src/frontend/public/blog-posts/generate-three-scales-diagram.py"

echo "🚀 Setting up three-scales-framework image generation and upload"
echo ""

# Check if script exists
if [ ! -f "$SCRIPT_PATH" ]; then
    echo "❌ Generation script not found at: $SCRIPT_PATH"
    exit 1
fi

echo "📋 Found generation script: $SCRIPT_PATH"

# Create ConfigMap with Python script and requirements
echo "📦 Creating ConfigMap $CONFIGMAP_NAME..."
kubectl create configmap "$CONFIGMAP_NAME" \
  --from-file=generate-three-scales-diagram.py="$SCRIPT_PATH" \
  --from-literal=requirements.txt="matplotlib
numpy" \
  --namespace="$NAMESPACE" \
  --dry-run=client -o yaml | kubectl apply -f -

echo "✅ ConfigMap created/updated"
echo ""

# Apply job
echo "🚀 Applying generate+upload job..."
kubectl apply -f "${SCRIPT_DIR}/three-scales-generate-upload-job.yaml"

echo "✅ Job applied"
echo ""

# Wait for job to start
echo "⏳ Waiting for job to start..."
sleep 3

# Show job status
echo "📊 Job status:"
kubectl get job/$JOB_NAME -n "$NAMESPACE" || true
echo ""

# Show pod status
echo "📦 Pod status:"
kubectl get pods -l job-name=$JOB_NAME -n "$NAMESPACE" || true
echo ""

# Wait for pod to be ready
POD_NAME=$(kubectl get pods -l job-name=$JOB_NAME -n "$NAMESPACE" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")

if [ -n "$POD_NAME" ]; then
    echo "⏳ Waiting for pod $POD_NAME to be ready..."
    kubectl wait --for=condition=Ready --timeout=60s pod/$POD_NAME -n "$NAMESPACE" || true
    echo ""
    
    # Show logs from generation container
    echo "📋 Generation logs (init container):"
    kubectl logs $POD_NAME -n "$NAMESPACE" -c generate-image --tail=50 || echo "Container not ready yet"
    echo ""
    
    # Follow upload logs
    echo "📋 Upload logs (main container):"
    echo "   (Press Ctrl+C to stop following logs, job will continue)"
    echo ""
    kubectl logs -f $POD_NAME -n "$NAMESPACE" -c upload-to-minio || {
        echo ""
        echo "⚠️  Upload container not ready yet, showing recent logs:"
        kubectl logs $POD_NAME -n "$NAMESPACE" -c upload-to-minio --tail=20 || true
    }
    
    echo ""
    echo "📊 Final job status:"
    kubectl get job/$JOB_NAME -n "$NAMESPACE"
    echo ""
    
    # Check if job succeeded
    JOB_STATUS=$(kubectl get job/$JOB_NAME -n "$NAMESPACE" -o jsonpath='{.status.conditions[?(@.type=="Complete")].status}' 2>/dev/null || echo "")
    if [ "$JOB_STATUS" = "True" ]; then
        echo "✅ Job completed successfully!"
        echo ""
        echo "🌐 Image should now be available at:"
        echo "   http://minio.minio.svc.cluster.local:9000/homepage-blog/images/graphs/three-scales-framework.png"
        echo "   Blog path: /storage/homepage-blog/images/graphs/three-scales-framework.png"
    else
        echo "⚠️  Job may still be running or failed. Check logs above."
    fi
else
    echo "⚠️  Pod not found yet. Check job status:"
    kubectl get job/$JOB_NAME -n "$NAMESPACE"
    echo ""
    echo "To view logs later:"
    echo "  kubectl logs -f job/$JOB_NAME -n $NAMESPACE"
fi

