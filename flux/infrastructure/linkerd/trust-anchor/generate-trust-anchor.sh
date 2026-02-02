#!/bin/bash
set -euo pipefail

# 🔐 Generate Linkerd Trust Anchor CA
# This script generates the shared trust anchor CA certificate and key
# that will be reused across all clusters for multicluster setup

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CA_CRT="$SCRIPT_DIR/ca.crt"
CA_KEY="$SCRIPT_DIR/ca.key"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔐 Generating Linkerd shared trust anchor"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Check if step-cli is installed
if ! command -v step >/dev/null 2>&1; then
  echo "❌ step-cli not found. Install via: brew install step"
  exit 1
fi

# Check if certificates already exist
if [ -f "$CA_CRT" ] && [ -f "$CA_KEY" ]; then
  echo "⚠️  Trust anchor already exists at:"
  echo "   $CA_CRT"
  echo "   $CA_KEY"
  read -p "Regenerate? (y/N): " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "✅ Keeping existing trust anchor"
    exit 0
  fi
  echo "🧹 Removing old trust anchor files..."
  rm -f "$CA_CRT" "$CA_KEY"
fi

echo "🔧 Generating shared Linkerd trust anchor..."
step certificate create root.linkerd.cluster.local \
  "$CA_CRT" "$CA_KEY" \
  --profile root-ca \
  --no-password \
  --insecure \
  --not-after 87600h

echo ""
echo "✅ Trust anchor generated successfully!"
echo "   Certificate: $CA_CRT"
echo "   Private Key: $CA_KEY"
echo ""
echo "💾 Commit these files to share with other clusters:"
echo "   git add $CA_CRT $CA_KEY"
echo "   git commit -m \"chore(linkerd): generate shared trust anchor for multicluster\""
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

