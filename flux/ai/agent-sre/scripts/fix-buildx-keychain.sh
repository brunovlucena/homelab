#!/bin/bash
# Fix buildx keychain issue by recreating builder without credential helper

set -e

BUILDER_NAME="${1:-homelab}"

echo "🔧 Fixing buildx keychain issue..."
echo "   Builder: $BUILDER_NAME"
echo ""

# Remove existing builder if it exists
if docker buildx inspect "$BUILDER_NAME" >/dev/null 2>&1; then
    echo "🗑️  Removing existing builder..."
    docker buildx rm "$BUILDER_NAME" || true
fi

# Create new builder without credential helper
echo "🔨 Creating new builder without credential helper..."
docker buildx create \
    --name "$BUILDER_NAME" \
    --driver docker-container \
    --driver-opt network=host \
    --use

# Bootstrap the builder
echo "🚀 Bootstrapping builder..."
docker buildx inspect "$BUILDER_NAME" --bootstrap

# Configure builder to not use credential helper
echo "⚙️  Configuring builder..."
BUILDER_CONTAINER="buildx_buildkit_${BUILDER_NAME}0"

if docker ps -a --format '{{.Names}}' | grep -q "^${BUILDER_CONTAINER}$"; then
    echo "   Updating builder container config..."
    docker exec "$BUILDER_CONTAINER" sh -c 'echo "{}" > /root/.docker/config.json' || true
fi

echo ""
echo "✅ Builder recreated and configured!"
echo ""
echo "📋 Try building again:"
echo "   make build-local"

