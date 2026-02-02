#!/bin/bash
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Agent Best Practices Compliance Checker
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#
# Checks all agents for compliance with best practices:
# - VERSION file exists
# - version-bump target in Makefile
# - Image tags in kustomizations
# - Standardized Makefile structure
#
# Usage: ./scripts/check-compliance.sh [agent-name]
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
AI_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
AGENT_DIR="${1:-}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
TOTAL_AGENTS=0
COMPLIANT_AGENTS=0
ISSUES=0

check_agent() {
    local agent_dir="$1"
    local agent_name=$(basename "$agent_dir")
    local has_issues=false
    
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}Checking: ${agent_name}${NC}"
    echo ""
    
    # Check 1: VERSION file exists
    if [ ! -f "$agent_dir/VERSION" ]; then
        echo -e "  ${RED}❌ Missing VERSION file${NC}"
        has_issues=true
        ((ISSUES++))
    else
        local version=$(cat "$agent_dir/VERSION" 2>/dev/null || echo "")
        if [ -z "$version" ]; then
            echo -e "  ${RED}❌ VERSION file is empty${NC}"
            has_issues=true
            ((ISSUES++))
        else
            echo -e "  ${GREEN}✅ VERSION file: $version${NC}"
        fi
    fi
    
    # Check 2: Makefile exists
    if [ ! -f "$agent_dir/Makefile" ]; then
        echo -e "  ${RED}❌ Missing Makefile${NC}"
        has_issues=true
        ((ISSUES++))
    else
        echo -e "  ${GREEN}✅ Makefile exists${NC}"
        
        # Check 3: version-bump target exists
        if ! grep -q "version-bump:" "$agent_dir/Makefile"; then
            echo -e "  ${RED}❌ Missing version-bump target${NC}"
            has_issues=true
            ((ISSUES++))
        else
            echo -e "  ${GREEN}✅ version-bump target found${NC}"
        fi
        
        # Check 4: release-patch/minor/major targets exist
        local has_release_targets=true
        if ! grep -q "release-patch:" "$agent_dir/Makefile"; then
            echo -e "  ${YELLOW}⚠️  Missing release-patch target${NC}"
            has_release_targets=false
        fi
        if ! grep -q "release-minor:" "$agent_dir/Makefile"; then
            echo -e "  ${YELLOW}⚠️  Missing release-minor target${NC}"
            has_release_targets=false
        fi
        if ! grep -q "release-major:" "$agent_dir/Makefile"; then
            echo -e "  ${YELLOW}⚠️  Missing release-major target${NC}"
            has_release_targets=false
        fi
        if [ "$has_release_targets" = true ]; then
            echo -e "  ${GREEN}✅ release-patch/minor/major targets found${NC}"
        fi
    fi
    
    # Check 5: Kustomization overlays have image tags
    local kustomize_dir="$agent_dir/k8s/kustomize"
    if [ -d "$kustomize_dir" ]; then
        local has_tags=false
        for overlay in "$kustomize_dir"/pro "$kustomize_dir"/studio; do
            if [ -f "$overlay/kustomization.yaml" ]; then
                if grep -q "path: /spec/image/tag" "$overlay/kustomization.yaml" || \
                   grep -q "newTag:" "$overlay/kustomization.yaml"; then
                    has_tags=true
                    break
                fi
            fi
        done
        
        if [ "$has_tags" = false ]; then
            echo -e "  ${RED}❌ No image tags in kustomization overlays${NC}"
            has_issues=true
            ((ISSUES++))
        else
            echo -e "  ${GREEN}✅ Image tags found in kustomizations${NC}"
        fi
    else
        echo -e "  ${YELLOW}⚠️  No k8s/kustomize directory found${NC}"
    fi
    
    # Check 6: VERSION_FILE variable in Makefile
    if [ -f "$agent_dir/Makefile" ]; then
        if ! grep -q "VERSION_FILE" "$agent_dir/Makefile"; then
            echo -e "  ${YELLOW}⚠️  Makefile doesn't use VERSION_FILE variable${NC}"
        else
            echo -e "  ${GREEN}✅ Makefile uses VERSION_FILE variable${NC}"
        fi
    fi
    
    # Summary for this agent
    if [ "$has_issues" = false ]; then
        echo -e "\n  ${GREEN}✅ ${agent_name} is COMPLIANT${NC}"
        ((COMPLIANT_AGENTS++))
    else
        echo -e "\n  ${RED}❌ ${agent_name} has ISSUES${NC}"
    fi
    
    echo ""
    ((TOTAL_AGENTS++))
}

# Main execution
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Agent Best Practices Compliance Checker${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

if [ -n "$AGENT_DIR" ]; then
    # Check specific agent
    if [ -d "$AI_DIR/$AGENT_DIR" ]; then
        check_agent "$AI_DIR/$AGENT_DIR"
    else
        echo -e "${RED}Error: Agent directory not found: $AGENT_DIR${NC}"
        exit 1
    fi
else
    # Check all agents
    for agent_dir in "$AI_DIR"/agent-*; do
        if [ -d "$agent_dir" ]; then
            check_agent "$agent_dir"
        fi
    done
fi

# Final summary
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Summary${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "Total agents checked: ${TOTAL_AGENTS}"
echo -e "Compliant agents: ${GREEN}${COMPLIANT_AGENTS}${NC}"
echo -e "Issues found: ${RED}${ISSUES}${NC}"

if [ $ISSUES -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✅ All agents are compliant!${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}❌ Some agents need fixes. See details above.${NC}"
    exit 1
fi
