#!/bin/bash

# Script to generate kubeconfig for GitHub Actions
# This creates a kubeconfig that points to your Cloudflare-exposed K8s API
# Uses the existing homepage service account

set -e

# Configuration
CLOUDFLARE_K8S_URL="https://k8s-api.lucena.cloud"  # Update with your actual Cloudflare URL
NAMESPACE="homepage"
SERVICE_ACCOUNT="homepage"
SECRET_NAME="homepage-github-token"
OUTPUT_FILE="github-kubeconfig.yaml"

echo "🔧 Generating kubeconfig for GitHub Actions using homepage service account..."
echo "📝 Using service account: $SERVICE_ACCOUNT in namespace: $NAMESPACE"

# Verify the service account exists
if ! kubectl get serviceaccount $SERVICE_ACCOUNT -n $NAMESPACE &>/dev/null; then
  echo "❌ Error: Service account $SERVICE_ACCOUNT not found in namespace $NAMESPACE"
  echo "   Make sure the homepage Helm chart is deployed first"
  exit 1
fi

# Verify the token secret exists
echo "🔍 Checking for service account token secret..."
if ! kubectl get secret $SECRET_NAME -n $NAMESPACE &>/dev/null; then
  echo "❌ Error: Secret $SECRET_NAME not found in namespace $NAMESPACE"
  echo "   The secret should be created by the homepage Helm chart"
  echo "   Make sure the homepage chart is deployed with serviceAccount.create=true"
  exit 1
fi

# Get the token
TOKEN=$(kubectl get secret $SECRET_NAME -n $NAMESPACE -o jsonpath='{.data.token}' | base64 -d)

# Get the CA certificate
CA_CERT=$(kubectl get secret $SECRET_NAME -n $NAMESPACE -o jsonpath='{.data.ca\.crt}')

# Create kubeconfig
cat > $OUTPUT_FILE << EOF
apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority-data: $CA_CERT
    server: $CLOUDFLARE_K8S_URL
  name: homelab-cluster
contexts:
- context:
    cluster: homelab-cluster
    user: $SERVICE_ACCOUNT
    namespace: $NAMESPACE
  name: homepage-github-context
current-context: homepage-github-context
users:
- name: $SERVICE_ACCOUNT
  user:
    token: $TOKEN
EOF

echo "✅ Kubeconfig generated: $OUTPUT_FILE"
echo "📋 Add this as a GitHub secret named 'KUBECONFIG':"
echo ""
cat $OUTPUT_FILE | base64 -w 0
echo ""
echo ""
echo "🔐 Security Note: This kubeconfig uses the $SERVICE_ACCOUNT service account in the $NAMESPACE namespace."
echo "📝 The kubeconfig is already configured to use $CLOUDFLARE_K8S_URL (no localhost references)"
