#!/bin/bash
set -e

echo "🔍 Diagnosing build environment..."
echo ""

# Check Docker
echo "1. Checking Docker installation..."
if command -v docker &> /dev/null; then
    echo "   ✅ Docker installed: $(docker --version)"
else
    echo "   ❌ Docker not found"
    exit 1
fi

# Check Docker daemon
echo ""
echo "2. Checking Docker daemon..."
if docker info &> /dev/null; then
    echo "   ✅ Docker daemon is running"
    echo "   Docker context: $(docker context show)"
else
    echo "   ❌ Docker daemon is not running or not accessible"
    exit 1
fi

# Check buildx
echo ""
echo "3. Checking Docker buildx..."
if docker buildx version &> /dev/null; then
    echo "   ✅ Buildx installed: $(docker buildx version)"
else
    echo "   ❌ Buildx not found"
    exit 1
fi

# Check buildx builder
echo ""
echo "4. Checking buildx builder 'homelab'..."
if docker buildx inspect homelab &> /dev/null; then
    echo "   ✅ Buildx builder 'homelab' exists"
    docker buildx inspect homelab --bootstrap 2>&1 | head -20
else
    echo "   ❌ Buildx builder 'homelab' does not exist"
    echo "   Try running: make ensure-buildx-builder"
    exit 1
fi

# Check local registry
echo ""
echo "5. Checking local registry at localhost:5001..."
if curl -s http://localhost:5001/v2/ &> /dev/null; then
    echo "   ✅ Local registry is accessible"
    echo "   Registry catalog:"
    curl -s http://localhost:5001/v2/_catalog 2>/dev/null | jq '.' || echo "   (empty or not json)"
else
    echo "   ❌ Local registry at localhost:5001 is not accessible"
    echo "   Try running: make registry (from homelab root) or check if local-registry container is running"
    echo "   Check with: docker ps | grep local-registry"
    exit 1
fi

# Check platform
echo ""
echo "6. Checking platform configuration..."
ARCH=$(uname -m)
echo "   Host architecture: $ARCH"
case "$ARCH" in
    arm64|aarch64)
        echo "   ✅ Expected platform: linux/arm64"
        ;;
    x86_64)
        echo "   ⚠️  Host is x86_64, but Makefile uses linux/arm64"
        ;;
    *)
        echo "   ⚠️  Unknown architecture: $ARCH"
        ;;
esac

# Check network connectivity for buildx
echo ""
echo "7. Checking buildx builder network configuration..."
BUILDER_NETWORK=$(docker buildx inspect homelab 2>/dev/null | grep -i "network" || echo "not found")
echo "   Builder network config: $BUILDER_NETWORK"

# Check if builder can access registry
echo ""
echo "8. Testing buildx builder connectivity to registry..."
docker buildx inspect homelab --bootstrap &> /dev/null
if docker buildx ls | grep -q "homelab.*running"; then
    echo "   ✅ Builder is running"
else
    echo "   ⚠️  Builder exists but may not be running"
fi

echo ""
echo "✅ All checks passed! Build environment looks good."
echo ""
echo "If build still fails, check:"
echo "  1. Docker buildx logs: docker buildx inspect homelab --bootstrap"
echo "  2. Try building with verbose output: DOCKER_BUILDKIT=1 docker buildx build --builder homelab --progress=plain ..."
echo "  3. Check if you can manually push to registry: docker tag alpine:latest localhost:5001/test:latest && docker push localhost:5001/test:latest"

