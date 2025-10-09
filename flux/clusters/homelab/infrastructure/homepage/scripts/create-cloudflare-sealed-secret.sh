#!/bin/bash

# Create Cloudflare Sealed Secret for Bruno Site
# This script helps you create a sealed secret for Cloudflare CDN configuration

set -e

echo "🔐 Creating Cloudflare Sealed Secret for Bruno Site"
echo "=================================================="

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl is not installed or not in PATH"
    exit 1
fi

# Check if kubeseal is available
if ! command -v kubeseal &> /dev/null; then
    echo "❌ kubeseal is not installed. Please install it first:"
    echo "   brew install kubeseal"
    echo "   # or"
    echo "   kubectl apply -f https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.24.5/controller.yaml"
    exit 1
fi

# Get Cloudflare credentials from user
echo ""
echo "📋 Please provide your Cloudflare credentials:"
echo ""

read -p "🌐 Cloudflare Zone ID: " CLOUDFLARE_ZONE_ID
if [ -z "$CLOUDFLARE_ZONE_ID" ]; then
    echo "❌ Zone ID is required"
    exit 1
fi

read -p "🔑 Cloudflare API Token: " CLOUDFLARE_API_TOKEN
if [ -z "$CLOUDFLARE_API_TOKEN" ]; then
    echo "❌ API Token is required"
    exit 1
fi

read -p "🌍 Domain (default: lucena.cloud): " CLOUDFLARE_DOMAIN
CLOUDFLARE_DOMAIN=${CLOUDFLARE_DOMAIN:-lucena.cloud}

read -p "⏰ Cache TTL in seconds (default: 86400): " CLOUDFLARE_CACHE_TTL
CLOUDFLARE_CACHE_TTL=${CLOUDFLARE_CACHE_TTL:-86400}

echo ""
echo "🔍 Verifying connection to Kubernetes cluster..."
if ! kubectl cluster-info &> /dev/null; then
    echo "❌ Cannot connect to Kubernetes cluster"
    exit 1
fi

echo "✅ Connected to Kubernetes cluster"

echo ""
echo "🔍 Checking if Sealed Secrets controller is running..."
if ! kubectl get deployment sealed-secrets-controller -n kube-system &> /dev/null; then
    echo "⚠️  Sealed Secrets controller not found. Installing..."
    kubectl apply -f https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.24.5/controller.yaml
    echo "⏳ Waiting for Sealed Secrets controller to be ready..."
    kubectl wait --for=condition=available --timeout=300s deployment/sealed-secrets-controller -n kube-system
fi

echo "✅ Sealed Secrets controller is ready"

echo ""
echo "🔐 Creating sealed secret..."

# Create the sealed secret
kubectl create secret generic bruno-site-cloudflare-secret \
  --from-literal=zone-id="$CLOUDFLARE_ZONE_ID" \
  --from-literal=api-token="$CLOUDFLARE_API_TOKEN" \
  --from-literal=domain="$CLOUDFLARE_DOMAIN" \
  --from-literal=enabled="true" \
  --from-literal=cache-ttl="$CLOUDFLARE_CACHE_TTL" \
  -n bruno \
  --dry-run=client -o yaml | kubeseal -o yaml > k8s/bruno-site-cloudflare-secret-sealed.yaml

echo "✅ Sealed secret created: k8s/bruno-site-cloudflare-secret-sealed.yaml"

echo ""
echo "🧹 Cleaning up temporary secret..."
kubectl delete secret bruno-site-cloudflare-secret -n bruno --ignore-not-found=true

echo ""
echo "🎉 Cloudflare Sealed Secret created successfully!"
echo ""
echo "📋 Next steps:"
echo "1. Review the sealed secret: cat k8s/bruno-site-cloudflare-secret-sealed.yaml"
echo "2. Commit the sealed secret to your repository"
echo "3. Deploy your application: make up"
echo ""
echo "🔍 To verify the secret:"
echo "   kubectl get secret bruno-site-cloudflare-secret -n bruno -o yaml"
echo ""
echo "🚀 Your Cloudflare CDN is now ready to use!"
