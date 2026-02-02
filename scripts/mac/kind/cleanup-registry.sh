#!/bin/bash
set -uo pipefail

# 🧹 Clean up unused images from local registry
# This script removes images that are not referenced in Flux manifests

REGISTRY_PORT="${REGISTRY_PORT:-5001}"
REGISTRY_HOST="127.0.0.1"
REGISTRY_URL="http://${REGISTRY_HOST}:${REGISTRY_PORT}"
REGISTRY_NAME="kind-registry"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HOMELAB_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🧹 Cleaning up unused images from local registry"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Check if registry is running
if ! docker ps --format '{{.Names}}' | grep -q "^${REGISTRY_NAME}$"; then
    echo -e "${RED}❌ Error: ${REGISTRY_NAME} is not running${NC}"
    echo "   Run: docker ps | grep registry"
    exit 1
fi

echo -e "${GREEN}✅ Registry is running on ${REGISTRY_HOST}:${REGISTRY_PORT}${NC}"
echo ""

# Get images from registry
echo "📦 Fetching images from registry..."
REGISTRY_IMAGES=$(curl -s "${REGISTRY_URL}/v2/_catalog" | jq -r '.repositories[]' 2>/dev/null || echo "")

if [ -z "$REGISTRY_IMAGES" ]; then
    echo -e "${RED}❌ Error: Failed to fetch images from registry${NC}"
    exit 1
fi

REGISTRY_COUNT=$(echo "$REGISTRY_IMAGES" | wc -l | tr -d ' ')
echo -e "   Found ${GREEN}${REGISTRY_COUNT}${NC} images in registry"
echo ""

# Extract expected images from Flux manifests (same logic as detect-missing-images.sh)
echo "🔎 Scanning Flux manifests for localhost:5001 references..."
EXPECTED_IMAGES=$(cd "$HOMELAB_ROOT" && grep -r "repository: localhost:5001/" flux/ 2>/dev/null | \
    grep -v ".git" | \
    sed 's/.*repository: localhost:5001\///' | \
    sed 's/"$//' | \
    awk '{print $1}' | \
    sort -u || echo "")

# Extract images from Jobs/CronJobs
JOB_IMAGES=$(cd "$HOMELAB_ROOT" && grep -r "image: localhost:5001/" flux/ 2>/dev/null | \
    grep -E "(job|cronjob)\.yaml:" | \
    sed 's/.*image: localhost:5001\///' | \
    sed 's/:.*$//' | \
    sort -u || echo "")

# Combine all expected images
ALL_EXPECTED=$(echo -e "${EXPECTED_IMAGES}\n${JOB_IMAGES}" | sed '/^$/d' | sort -u)

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

# Find unused images
UNUSED_IMAGES=()
UNUSED_COUNT=0

while IFS= read -r registry_img; do
    [ -z "$registry_img" ] && continue
    if ! echo "$COMPLETE_EXPECTED" | grep -q "^${registry_img}$"; then
        UNUSED_IMAGES+=("$registry_img")
        UNUSED_COUNT=$((UNUSED_COUNT + 1))
    fi
done <<< "$REGISTRY_IMAGES"

if [ $UNUSED_COUNT -eq 0 ]; then
    echo -e "${GREEN}✅ No unused images found! Registry is clean.${NC}"
    echo ""
    exit 0
fi

echo -e "${YELLOW}⚠️  Found ${UNUSED_COUNT} unused image(s):${NC}"
echo ""
for img in "${UNUSED_IMAGES[@]}"; do
    echo -e "  ${YELLOW}•${NC} ${img}"
done
echo ""

# Ask for confirmation (skip if running non-interactively with AUTO_YES=1)
if [ "${AUTO_YES:-}" != "1" ]; then
    read -p "Delete these ${UNUSED_COUNT} unused image(s)? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Cancelled."
        exit 0
    fi
else
    echo "AUTO_YES=1 detected, proceeding with deletion..."
fi

echo ""
echo "🗑️  Deleting unused images..."
echo ""

DELETED_COUNT=0
FAILED_COUNT=0
NOT_IN_CATALOG_COUNT=0
EMPTY_REPO_COUNT=0

# Function to delete an image repository from registry
delete_image_repo() {
    local repo=$1
    local deleted=0
    local failed=0
    local not_found=0
    
    # First, check if repository exists in catalog at all
    local catalog_check=$(curl -sf "${REGISTRY_URL}/v2/_catalog" | jq -r ".repositories[] | select(. == \"${repo}\")" 2>/dev/null)
    
    if [ -z "$catalog_check" ]; then
        # Not in catalog - definitely doesn't exist
        echo -e "  ${GREEN}✅${NC} Confirmed: ${repo} does NOT exist in registry (already deleted or never existed)"
        return 2
    fi
    
    # Repository exists in catalog, get all tags
    local tags_response=$(curl -s "${REGISTRY_URL}/v2/${repo}/tags/list" 2>/dev/null)
    
    # Check if we got a valid response
    if ! echo "$tags_response" | jq -e '.name' >/dev/null 2>&1; then
        # Invalid response - try to delete anyway by attempting common tags
        echo -e "  ${YELLOW}⚠️${NC}  Invalid tags response for ${repo}, attempting deletion via common tags..."
    fi
    
    # Get tags from response
    local tags=$(echo "$tags_response" | jq -r '.tags[]?' 2>/dev/null || echo "")
    
    # Collect all unique digests we'll try to delete
    local all_digests=""
    
    # If we have tags, get their digests
    if [ -n "$tags" ] && [ "$tags" != "null" ] && [ -n "$(echo "$tags" | grep -v '^$')" ]; then
        while IFS= read -r tag; do
            [ -z "$tag" ] && continue
            [ "$tag" = "null" ] && continue
            
            # Get manifest digest for this tag (try v2 first, then v1)
            local manifest_digest=$(curl -sf -I \
                -H "Accept: application/vnd.docker.distribution.manifest.v2+json" \
                "${REGISTRY_URL}/v2/${repo}/manifests/${tag}" 2>/dev/null | \
                grep -i "Docker-Content-Digest:" | cut -d' ' -f2 | tr -d '\r')
            
            if [ -z "$manifest_digest" ]; then
                manifest_digest=$(curl -sf -I \
                    -H "Accept: application/vnd.docker.distribution.manifest.v1+json" \
                    "${REGISTRY_URL}/v2/${repo}/manifests/${tag}" 2>/dev/null | \
                    grep -i "Docker-Content-Digest:" | cut -d' ' -f2 | tr -d '\r')
            fi
            
            if [ -n "$manifest_digest" ]; then
                # Add to list if not already there
                if ! echo "$all_digests" | grep -q "$manifest_digest"; then
                    all_digests="${all_digests}${manifest_digest}\n"
                fi
            fi
        done <<< "$tags"
    fi
    
    # Also try common tag names as fallback
    for common_tag in "latest" "main" "master" "v1" "v1.0" "1.0.0"; do
        local manifest_digest=$(curl -sf -I \
            -H "Accept: application/vnd.docker.distribution.manifest.v2+json" \
            "${REGISTRY_URL}/v2/${repo}/manifests/${common_tag}" 2>/dev/null | \
            grep -i "Docker-Content-Digest:" | cut -d' ' -f2 | tr -d '\r')
        
        if [ -n "$manifest_digest" ]; then
            if ! echo "$all_digests" | grep -q "$manifest_digest"; then
                all_digests="${all_digests}${manifest_digest}\n"
            fi
        fi
    done
    
    # Now delete all unique digests
    if [ -n "$all_digests" ] && [ -n "$(echo "$all_digests" | grep -v '^$')" ]; then
        while IFS= read -r digest; do
            [ -z "$digest" ] && continue
            
            # Delete by digest
            if curl -sf -X DELETE "${REGISTRY_URL}/v2/${repo}/manifests/${digest}" >/dev/null 2>&1; then
                deleted=1
            else
                # Try with explicit accept header
                if curl -sf -X DELETE \
                    -H "Accept: application/vnd.docker.distribution.manifest.v2+json" \
                    "${REGISTRY_URL}/v2/${repo}/manifests/${digest}" >/dev/null 2>&1; then
                    deleted=1
                else
                    # Last attempt: try with v1 header
                    if curl -sf -X DELETE \
                        -H "Accept: application/vnd.docker.distribution.manifest.v1+json" \
                        "${REGISTRY_URL}/v2/${repo}/manifests/${digest}" >/dev/null 2>&1; then
                        deleted=1
                    else
                        failed=1
                    fi
                fi
            fi
        done <<< "$(echo -e "$all_digests" | grep -v '^$')"
    else
        # Repository exists in catalog but we couldn't find any manifests
        # This could mean:
        # 1. It's an empty repository (already cleaned but catalog entry remains)
        # 2. Tags exist but manifests are corrupted
        # 3. Registry API issue
        # Since it's in the catalog, we should report it as needing cleanup
        echo -e "  ${YELLOW}⚠️${NC}  ${repo} exists in catalog but NO MANIFESTS found (empty repo - needs garbage collection)"
        not_found=1
    fi
    
    if [ $deleted -eq 1 ]; then
        echo -e "  ${GREEN}✅${NC} Deleted: ${repo}"
        return 0
    elif [ $failed -eq 1 ]; then
        echo -e "  ${RED}❌${NC} Failed: ${repo} (could not delete manifests - registry may not support deletion)"
        return 1
    else
        # Not found - but we confirmed it's not in catalog or has no manifests
        return 2
    fi
}

# Delete each unused image
for img in "${UNUSED_IMAGES[@]}"; do
    delete_image_repo "$img"
    result=$?
    if [ $result -eq 0 ]; then
        DELETED_COUNT=$((DELETED_COUNT + 1))
    elif [ $result -eq 1 ]; then
        FAILED_COUNT=$((FAILED_COUNT + 1))
    elif [ $result -eq 2 ]; then
        # Check if it was "not in catalog" (2) or "empty repo" (also 2 but different message)
        # We can't distinguish easily, so we'll count them together and show in summary
        NOT_IN_CATALOG_COUNT=$((NOT_IN_CATALOG_COUNT + 1))
    else
        # Empty repo case (also returns 2 but with warning message)
        EMPTY_REPO_COUNT=$((EMPTY_REPO_COUNT + 1))
    fi
done

echo ""
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 Cleanup Summary"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "✅ Deleted:     ${DELETED_COUNT} (manifests successfully removed)"
echo "✅ Confirmed:   ${NOT_IN_CATALOG_COUNT} (do NOT exist in registry - verified)"
if [ $EMPTY_REPO_COUNT -gt 0 ]; then
    echo "⚠️  Empty repos: ${EMPTY_REPO_COUNT} (in catalog but no manifests - need garbage collection)"
fi
if [ $FAILED_COUNT -gt 0 ]; then
    echo "❌ Failed:      ${FAILED_COUNT} (exist but deletion failed)"
fi
echo ""

if [ $DELETED_COUNT -gt 0 ]; then
    echo -e "${GREEN}✅ Successfully deleted ${DELETED_COUNT} image(s)!${NC}"
    echo ""
    echo "💡 Run garbage collection to free disk space:"
    echo "   docker exec ${REGISTRY_NAME} registry garbage-collect /etc/docker/registry/config.yml"
    echo ""
fi

if [ $NOT_IN_CATALOG_COUNT -gt 0 ]; then
    echo -e "${GREEN}✅ Confirmed ${NOT_IN_CATALOG_COUNT} image(s) do NOT exist in registry (verified via catalog).${NC}"
    echo ""
fi

if [ $EMPTY_REPO_COUNT -gt 0 ]; then
    echo -e "${YELLOW}⚠️  ${EMPTY_REPO_COUNT} repository/repositories exist in catalog but have no manifests.${NC}"
    echo "   These are empty repository entries that need garbage collection to remove:"
    echo "   docker exec ${REGISTRY_NAME} registry garbage-collect /etc/docker/registry/config.yml"
    echo ""
fi

if [ $FAILED_COUNT -gt 0 ]; then
    echo -e "${RED}❌ ${FAILED_COUNT} image(s) FAILED to delete.${NC}"
    echo "   These images exist in the registry with manifests but deletion failed."
    echo "   Possible causes:"
    echo "   - Registry doesn't support deletion (check registry config for 'delete.enabled: true')"
    echo "   - Images are in use by running containers"
    echo "   - Permission issues"
    echo ""
    echo "   Try running garbage collection:"
    echo "   docker exec ${REGISTRY_NAME} registry garbage-collect /etc/docker/registry/config.yml"
    echo ""
fi

# Exit with error if there are failures
if [ $FAILED_COUNT -gt 0 ]; then
    exit 1
fi

exit 0
