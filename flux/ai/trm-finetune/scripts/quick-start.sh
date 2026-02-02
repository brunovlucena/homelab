#!/bin/bash
set -euo pipefail

# Quick start script for TRM fine-tuning pipeline

echo "🚀 TRM Fine-Tuning Pipeline - Quick Start"
echo "=========================================="

# Check prerequisites
echo "📋 Checking prerequisites..."

command -v docker >/dev/null 2>&1 || { echo "❌ Docker not found"; exit 1; }
command -v kubectl >/dev/null 2>&1 || { echo "❌ kubectl not found"; exit 1; }
command -v flytectl >/dev/null 2>&1 || { echo "⚠️  flytectl not found (optional for manual triggers)"; }

echo "✅ Prerequisites check complete"

# Build Docker image
echo ""
echo "🔨 Building Docker image..."
docker build -t localhost:5001/trm-finetune:latest .

if [ $? -eq 0 ]; then
    echo "✅ Docker image built successfully"
else
    echo "❌ Docker build failed"
    exit 1
fi

# Push to registry (if local registry)
echo ""
echo "📤 Pushing to registry..."
docker push localhost:5001/trm-finetune:latest || echo "⚠️  Push failed (may need to configure registry)"

# Deploy Kubernetes resources
echo ""
echo "☸️  Deploying to Kubernetes..."
kubectl apply -k k8s/kustomize/base/

if [ $? -eq 0 ]; then
    echo "✅ Kubernetes resources deployed"
else
    echo "❌ Deployment failed"
    exit 1
fi

# Verify deployment
echo ""
echo "🔍 Verifying deployment..."
kubectl get configmap -n ml-platform trm-finetune-config
kubectl get secret -n ml-platform trm-finetune-secrets

echo ""
echo "🎉 Setup complete!"
echo ""
echo "Next steps:"
echo "1. Register workflow with Flyte:"
echo "   pyflyte register src/flyte_workflow.py \\"
echo "     --project homelab \\"
echo "     --domain production \\"
echo "     --image localhost:5001/trm-finetune:latest"
echo ""
echo "2. Trigger manual run:"
echo "   flytectl create execution \\"
echo "     --project homelab \\"
echo "     --domain production \\"
echo "     --workflow trm_finetuning_workflow"
echo ""
echo "3. Check scheduled workflow:"
echo "   flytectl get launch-plan \\"
echo "     --project homelab \\"
echo "     --domain production \\"
echo "     monthly_trm_finetuning"


