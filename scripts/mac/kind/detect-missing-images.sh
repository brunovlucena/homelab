#!/bin/bash
set -euo pipefail

# 🔍 Detect missing images in local registry
# This script compares images referenced in Flux manifests with images in the local registry

REGISTRY_PORT="${REGISTRY_PORT:-5001}"
REGISTRY_URL="localhost:${REGISTRY_PORT}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HOMELAB_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔍 Detecting missing images in local registry"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Check if registry is running
if ! docker ps --format '{{.Names}}' | grep -q "^kind-registry$"; then
    echo -e "${RED}❌ Error: kind-registry is not running${NC}"
    echo "   Run: docker ps | grep registry"
    exit 1
fi

echo -e "${GREEN}✅ Registry is running on ${REGISTRY_URL}${NC}"
echo ""

# Get images from registry
echo "📦 Fetching images from registry..."
REGISTRY_IMAGES=$(curl -s "http://${REGISTRY_URL}/v2/_catalog" | jq -r '.repositories[]' 2>/dev/null || echo "")

if [ -z "$REGISTRY_IMAGES" ]; then
    echo -e "${RED}❌ Error: Failed to fetch images from registry${NC}"
    echo "   Check if registry is accessible: curl http://${REGISTRY_URL}/v2/_catalog"
    exit 1
fi

REGISTRY_COUNT=$(echo "$REGISTRY_IMAGES" | wc -l | tr -d ' ')
echo -e "   Found ${GREEN}${REGISTRY_COUNT}${NC} images in registry"
echo ""

# Extract expected images from Flux manifests
echo "🔎 Scanning Flux manifests for localhost:5001 references..."
EXPECTED_IMAGES=$(cd "$HOMELAB_ROOT" && grep -r "repository: localhost:5001/" flux/ 2>/dev/null | \
    grep -v ".git" | \
    sed 's/.*repository: localhost:5001\///' | \
    sed 's/"$//' | \
    awk '{print $1}' | \
    sort -u || echo "")

if [ -z "$EXPECTED_IMAGES" ]; then
    echo -e "${YELLOW}⚠️  Warning: No images found in Flux manifests${NC}"
    echo ""
else
    EXPECTED_COUNT=$(echo "$EXPECTED_IMAGES" | wc -l | tr -d ' ')
    echo -e "   Found ${BLUE}${EXPECTED_COUNT}${NC} unique image references in Flux"
    echo ""
fi

# Extract images from Jobs/CronJobs
echo "🔎 Scanning Jobs and CronJobs..."
JOB_IMAGES=$(cd "$HOMELAB_ROOT" && grep -r "image: localhost:5001/" flux/ 2>/dev/null | \
    grep -E "(job|cronjob)\.yaml:" | \
    sed 's/.*image: localhost:5001\///' | \
    sed 's/:.*$//' | \
    sort -u || echo "")

if [ -n "$JOB_IMAGES" ]; then
    JOB_COUNT=$(echo "$JOB_IMAGES" | wc -l | tr -d ' ')
    echo -e "   Found ${BLUE}${JOB_COUNT}${NC} images in Jobs/CronJobs"
    echo ""
fi

# Combine all expected images
ALL_EXPECTED=$(echo -e "${EXPECTED_IMAGES}\n${JOB_IMAGES}" | sed '/^$/d' | sort -u)
ALL_EXPECTED_COUNT=$(echo "$ALL_EXPECTED" | wc -l | tr -d ' ')

# Known Pulumi-managed images (not in Flux manifests)
PULUMI_IMAGES="coredns
helm-controller
kustomize-controller
notification-controller
source-controller
linkerd-controller
controller
linkerd-metrics-api
linkerd-proxy
linkerd-proxy-init
linkerd-tap
linkerd-web
kube-state-metrics"

# Utility and networking images
UTILITY_IMAGES="alpine
busybox
curl
cloudflared"

# Combine with Pulumi and utility images
COMPLETE_EXPECTED=$(echo -e "${ALL_EXPECTED}\n${PULUMI_IMAGES}\n${UTILITY_IMAGES}" | sed '/^$/d' | sort -u)
COMPLETE_COUNT=$(echo "$COMPLETE_EXPECTED" | wc -l | tr -d ' ')

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 Analysis Results"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Expected images (total):  ${COMPLETE_COUNT}"
echo "  • From Flux manifests:  ${ALL_EXPECTED_COUNT}"
echo "  • From Pulumi:          $(echo "$PULUMI_IMAGES" | wc -l | tr -d ' ')"
echo "  • Utility images:       $(echo "$UTILITY_IMAGES" | wc -l | tr -d ' ')"
echo "Registry images:          ${REGISTRY_COUNT}"
echo ""

# Find missing images
MISSING_IMAGES=""
MISSING_COUNT=0

while IFS= read -r expected_img; do
    [ -z "$expected_img" ] && continue
    if ! echo "$REGISTRY_IMAGES" | grep -q "^${expected_img}$"; then
        MISSING_IMAGES="${MISSING_IMAGES}${expected_img}\n"
        MISSING_COUNT=$((MISSING_COUNT + 1))
    fi
done <<< "$COMPLETE_EXPECTED"

# Find unused images
UNUSED_IMAGES=""
UNUSED_COUNT=0

while IFS= read -r registry_img; do
    [ -z "$registry_img" ] && continue
    if ! echo "$COMPLETE_EXPECTED" | grep -q "^${registry_img}$"; then
        UNUSED_IMAGES="${UNUSED_IMAGES}${registry_img}\n"
        UNUSED_COUNT=$((UNUSED_COUNT + 1))
    fi
done <<< "$REGISTRY_IMAGES"

# Report results
if [ $MISSING_COUNT -eq 0 ]; then
    echo -e "${GREEN}✅ All expected images are present in registry!${NC}"
    echo ""
else
    echo -e "${RED}❌ Missing ${MISSING_COUNT} image(s) in registry:${NC}"
    echo ""
    echo -e "$MISSING_IMAGES" | sed '/^$/d' | while read -r img; do
        # Find where it's used
        USAGE=$(cd "$HOMELAB_ROOT" && grep -r "localhost:5001/${img}" flux/ 2>/dev/null | head -1 | cut -d':' -f1 || echo "Unknown")
        echo -e "  ${RED}•${NC} ${img}"
        echo -e "    ${BLUE}→${NC} Used in: $(basename "$USAGE")"
    done
    echo ""
fi

if [ $UNUSED_COUNT -gt 0 ]; then
    echo -e "${YELLOW}⚠️  Found ${UNUSED_COUNT} unused image(s) in registry:${NC}"
    echo ""
    echo -e "$UNUSED_IMAGES" | sed '/^$/d' | while read -r img; do
        echo -e "  ${YELLOW}•${NC} ${img}"
    done
    echo ""
    echo "Note: These might be old versions or test images. Consider cleaning them up."
    echo ""
fi

# Check if cluster is running
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔍 Cluster Status Check"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

CLUSTER_RUNNING=false
if kind get clusters 2>/dev/null | grep -q "studio"; then
    CLUSTER_RUNNING=true
    echo -e "${GREEN}✅ Kind cluster 'studio' is running${NC}"
    
    # Check if pods can pull images
    if kubectl config current-context 2>/dev/null | grep -q "kind-studio"; then
        echo ""
        echo "📊 Checking pod image pull status..."
        
        PENDING_PODS=$(kubectl get pods -A --field-selector=status.phase=Pending -o json 2>/dev/null | \
            jq -r '.items[] | select(.status.conditions[]?.reason == "ImagePullBackOff" or .status.conditions[]?.reason == "ErrImagePull") | .metadata.name' 2>/dev/null || echo "")
        
        if [ -n "$PENDING_PODS" ]; then
            echo -e "${RED}❌ Found pods with image pull issues:${NC}"
            echo "$PENDING_PODS" | while read -r pod; do
                [ -z "$pod" ] && continue
                echo -e "  ${RED}•${NC} ${pod}"
            done
            echo ""
            echo -e "${YELLOW}💡 Run prewarm-images.sh to add missing images${NC}"
        else
            echo -e "${GREEN}✅ No image pull issues detected${NC}"
        fi
    fi
else
    echo -e "${YELLOW}⚠️  No Kind cluster found${NC}"
    echo "   Run: make up ENV=studio"
fi

echo ""

# Summary and recommendations
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📋 Summary"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

if [ $MISSING_COUNT -eq 0 ] && [ $UNUSED_COUNT -eq 0 ]; then
    echo -e "${GREEN}✅ Registry is healthy and up-to-date!${NC}"
    echo ""
    echo "Registry size: $(docker system df -v | grep kind-registry -A 1 | tail -1 | awk '{print $3}')"
    echo "Total images:  ${REGISTRY_COUNT}"
    exit 0
else
    if [ $MISSING_COUNT -gt 0 ]; then
        echo -e "${RED}❌ Action required: ${MISSING_COUNT} missing image(s)${NC}"
        echo ""
        echo "To fix:"
        echo "  1. Run: bash scripts/mac/prewarm-images.sh"
        echo "  2. Or pull manually:"
        echo ""
        echo -e "$MISSING_IMAGES" | sed '/^$/d' | while read -r img; do
            echo "     docker pull <source-image> && docker tag <source-image> localhost:5001/${img} && docker push localhost:5001/${img}"
        done
        echo ""
    fi
    
    if [ $UNUSED_COUNT -gt 0 ]; then
        echo -e "${YELLOW}⚠️  Optional: ${UNUSED_COUNT} unused image(s) taking space${NC}"
        echo ""
        echo "To clean up unused images:"
        echo "  bash scripts/mac/kind/cleanup-registry.sh"
        echo ""
    fi
    
    exit 1
fi

