#!/bin/bash

# 🚀 FORCE DATABASE MIGRATION SCRIPT
# This script triggers an immediate database migration

set -e

echo "🚀 FORCING DATABASE MIGRATION..."
echo "⏰ Timestamp: $(date)"

# Check if we're in the right directory
if [ ! -f "chart/values.yaml" ]; then
    echo "❌ Error: Please run this script from the homepage directory"
    echo "   Current directory: $(pwd)"
    echo "   Expected: .../homepage/"
    exit 1
fi

# Get the namespace (default to bruno)
NAMESPACE=${1:-bruno}
echo "📦 Using namespace: $NAMESPACE"

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "❌ Error: kubectl is not installed or not in PATH"
    exit 1
fi

# Check if we can connect to the cluster
if ! kubectl cluster-info &> /dev/null; then
    echo "❌ Error: Cannot connect to Kubernetes cluster"
    exit 1
fi

echo "✅ Kubernetes cluster connection verified"

# Create a unique job name with timestamp
TIMESTAMP=$(date +"%Y%m%d-%H%M%S")
JOB_NAME="homepage-bruno-site-db-migrate-force-${TIMESTAMP}"

echo "🔧 Creating migration job: $JOB_NAME"

# Create the job from the cronjob template
kubectl create job "$JOB_NAME" \
    --from=cronjob/homepage-bruno-site-db-migrate \
    -n "$NAMESPACE" \
    --dry-run=client -o yaml | \
kubectl apply -f -

echo "✅ Migration job created: $JOB_NAME"

# Wait for the job to complete
echo "⏳ Waiting for migration to complete..."
kubectl wait --for=condition=complete job/"$JOB_NAME" -n "$NAMESPACE" --timeout=300s

# Get the job status
JOB_STATUS=$(kubectl get job "$JOB_NAME" -n "$NAMESPACE" -o jsonpath='{.status.conditions[0].type}')
if [ "$JOB_STATUS" = "Complete" ]; then
    echo "✅ Migration completed successfully!"
    
    # Show the logs
    echo "📋 Migration logs:"
    kubectl logs job/"$JOB_NAME" -n "$NAMESPACE"
    
    # Clean up the job
    echo "🧹 Cleaning up job..."
    kubectl delete job "$JOB_NAME" -n "$NAMESPACE"
    
    echo "✅ Force migration completed successfully!"
else
    echo "❌ Migration failed!"
    echo "📋 Job status: $JOB_STATUS"
    
    # Show the logs for debugging
    echo "📋 Migration logs:"
    kubectl logs job/"$JOB_NAME" -n "$NAMESPACE"
    
    echo "🔍 Job details:"
    kubectl describe job "$JOB_NAME" -n "$NAMESPACE"
    
    exit 1
fi
