#!/bin/bash

# Test all Cloudflare Tunnel Ingress configurations for command centers
# Note: set -e is disabled to allow HTTP tests to continue even if some fail
set +e

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🧪 Testing All Command Center Tunnel Links"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test results
PASSED=0
FAILED=0
WARNINGS=0

# Function to test a tunnel ingress configuration
test_tunnel_ingress() {
    local file=$1
    local name=$(basename $(dirname $(dirname $(dirname $file))))
    
    echo -e "${BLUE}Testing: ${name}${NC}"
    echo "  File: $file"
    
    # Check if file exists
    if [ ! -f "$file" ]; then
        echo -e "  ${RED}❌ FAILED: File not found${NC}"
        ((FAILED++))
        return 1
    fi
    
    # Extract configuration values
    local hostname=$(grep "^  hostname:" "$file" | sed 's/.*hostname: *//' | tr -d ' "')
    local service_name=$(grep "^    name:" "$file" | sed 's/.*name: *//' | tr -d ' "')
    local service_namespace=$(grep "^    namespace:" "$file" | sed 's/.*namespace: *//' | tr -d ' "')
    local service_port=$(grep "^    port:" "$file" | sed 's/.*port: *//' | tr -d ' "')
    
    # Extract namespace from metadata if not in service spec
    if [ -z "$service_namespace" ]; then
        service_namespace=$(grep -A 3 "metadata:" "$file" | grep "namespace:" | sed 's/.*namespace: *//' | tr -d ' "')
    fi
    
    echo "  Hostname: $hostname"
    echo "  Service: $service_name"
    echo "  Namespace: $service_namespace"
    echo "  Port: $service_port"
    
    # Validate YAML syntax
    if ! command -v yq &> /dev/null && ! command -v kubectl &> /dev/null; then
        echo -e "  ${YELLOW}⚠️  WARNING: yq or kubectl not available, skipping YAML validation${NC}"
        ((WARNINGS++))
    else
        # Try to validate with kubectl if available
        if command -v kubectl &> /dev/null; then
            if kubectl --dry-run=client apply -f "$file" &> /dev/null; then
                echo -e "  ${GREEN}✓ YAML syntax valid${NC}"
            else
                echo -e "  ${RED}❌ YAML syntax invalid${NC}"
                kubectl --dry-run=client apply -f "$file" 2>&1 | head -5
                ((FAILED++))
                return 1
            fi
        fi
    fi
    
    # Check if service exists (if kubectl is available and connected to cluster)
    if command -v kubectl &> /dev/null && kubectl cluster-info &> /dev/null 2>&1; then
        if kubectl get service "$service_name" -n "$service_namespace" &> /dev/null; then
            echo -e "  ${GREEN}✓ Service exists in cluster${NC}"
            
            # Check if service port matches
            local actual_port=$(kubectl get service "$service_name" -n "$service_namespace" -o jsonpath='{.spec.ports[0].port}' 2>/dev/null || echo "")
            if [ -n "$actual_port" ] && [ "$actual_port" = "$service_port" ]; then
                echo -e "  ${GREEN}✓ Service port matches ($service_port)${NC}"
            elif [ -n "$actual_port" ]; then
                echo -e "  ${YELLOW}⚠️  WARNING: Service port mismatch (expected: $service_port, actual: $actual_port)${NC}"
                ((WARNINGS++))
            fi
            
            # Check tunnel ingress CRD if deployed
            local ingress_name=$(grep -A 3 "metadata:" "$file" | grep "name:" | sed 's/.*name: *//' | tr -d ' "')
            if kubectl get cloudflaretunnelingress "$ingress_name" -n "$service_namespace" &> /dev/null 2>&1; then
                echo -e "  ${GREEN}✓ Tunnel ingress CRD exists${NC}"
                local phase=$(kubectl get cloudflaretunnelingress "$ingress_name" -n "$service_namespace" -o jsonpath='{.status.phase}' 2>/dev/null || echo "")
                if [ -n "$phase" ]; then
                    if [ "$phase" = "Ready" ]; then
                        echo -e "  ${GREEN}✓ Tunnel status: Ready${NC}"
                    else
                        echo -e "  ${YELLOW}⚠️  Tunnel status: $phase${NC}"
                        ((WARNINGS++))
                    fi
                fi
            else
                echo -e "  ${YELLOW}⚠️  Tunnel ingress CRD not found (may not be deployed yet)${NC}"
                ((WARNINGS++))
            fi
        else
            echo -e "  ${YELLOW}⚠️  Service not found in cluster (may not be deployed yet)${NC}"
            ((WARNINGS++))
        fi
    else
        echo -e "  ${YELLOW}⚠️  Cannot connect to cluster, skipping runtime checks${NC}"
        ((WARNINGS++))
    fi
    
    # Validate hostname format
    if [[ "$hostname" =~ ^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
        echo -e "  ${GREEN}✓ Hostname format valid${NC}"
    else
        echo -e "  ${RED}❌ Hostname format invalid: $hostname${NC}"
        ((FAILED++))
        return 1
    fi
    
    # Check if port is valid
    if [ -n "$service_port" ] && [ "$service_port" -ge 1 ] && [ "$service_port" -le 65535 ] 2>/dev/null; then
        echo -e "  ${GREEN}✓ Port valid ($service_port)${NC}"
    elif [ -n "$service_port" ]; then
        echo -e "  ${RED}❌ Port invalid: $service_port${NC}"
        ((FAILED++))
        return 1
    fi
    
    # Test actual HTTP connectivity
    local url="https://${hostname}"
    echo -e "  ${BLUE}Testing HTTP connectivity to ${url}...${NC}"
    
    local http_code=$(curl -s -o /dev/null -w "%{http_code}" --max-time 10 --connect-timeout 5 -L "${url}" 2>&1 | tail -1)
    local curl_exit=$?
    local curl_error=$(curl -s -o /dev/null -w "%{http_code}" --max-time 10 --connect-timeout 5 -L "${url}" 2>&1 | grep -v "^[0-9]")
    
    # If curl failed, http_code might be empty or contain error text
    if [ -z "$http_code" ] || [ "$curl_exit" -ne 0 ] || ! [[ "$http_code" =~ ^[0-9]+$ ]]; then
        echo -e "  ${RED}❌ HTTP connectivity FAILED (Connection error)${NC}"
        if [ -n "$curl_error" ]; then
            echo -e "    Error: ${curl_error}${NC}"
        fi
        ((FAILED++))
        return 1
    fi
    
    if [ "$http_code" = "200" ] || [ "$http_code" = "301" ] || [ "$http_code" = "302" ] || [ "$http_code" = "307" ] || [ "$http_code" = "308" ]; then
        echo -e "  ${GREEN}✓ HTTP connectivity OK (Status: ${http_code})${NC}"
    elif [ "$http_code" = "000" ]; then
        echo -e "  ${RED}❌ HTTP connectivity FAILED (Connection timeout/refused)${NC}"
        ((FAILED++))
        return 1
    elif [ "$http_code" -ge 400 ] && [ "$http_code" -lt 500 ]; then
        echo -e "  ${RED}❌ HTTP connectivity FAILED (Client error: ${http_code})${NC}"
        ((FAILED++))
        return 1
    elif [ "$http_code" -ge 500 ]; then
        echo -e "  ${YELLOW}⚠️  HTTP connectivity WARNING (Server error: ${http_code})${NC}"
        ((WARNINGS++))
    else
        echo -e "  ${YELLOW}⚠️  HTTP connectivity WARNING (Status: ${http_code})${NC}"
        ((WARNINGS++))
    fi
    
    echo ""
    ((PASSED++))
}

# Find all tunnel ingress files
BASE_DIR="/Users/brunolucena/workspace/bruno/repos/homelab/flux/ai"

echo "📋 Found tunnel ingress configurations:"
echo ""

# Test each command center tunnel
test_tunnel_ingress "$BASE_DIR/agent-chat/web-command-center/k8s/kustomize/base/cloudflare-tunnel-ingress.yaml"
test_tunnel_ingress "$BASE_DIR/agent-restaurant/web/k8s/kustomize/base/cloudflare-tunnel-ingress.yaml"
test_tunnel_ingress "$BASE_DIR/agent-pos-edge/web-mcdonalds/k8s/kustomize/base/cloudflare-tunnel-ingress.yaml"
test_tunnel_ingress "$BASE_DIR/agent-pos-edge/web-gas-station/k8s/kustomize/base/cloudflare-tunnel-ingress.yaml"
test_tunnel_ingress "$BASE_DIR/agent-store-multibrands/web-command-center/k8s/kustomize/base/cloudflare-tunnel-ingress.yaml"

# Summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 Test Summary"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}✓ Passed: $PASSED${NC}"
echo -e "${YELLOW}⚠️  Warnings: $WARNINGS${NC}"
echo -e "${RED}❌ Failed: $FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All tunnel configurations are valid!${NC}"
    echo ""
    echo "🌐 Command Center URLs:"
    echo "  • Chat:           https://chat.lucena.cloud"
    echo "  • Restaurant:     https://restaurant.lucena.cloud"
    echo "  • McDonald's:     https://mcdonalds.lucena.cloud"
    echo "  • Gas Station:    https://gasstation.lucena.cloud"
    echo "  • Store:          https://store.lucena.cloud"
    exit 0
else
    echo -e "${RED}❌ Some configurations failed validation${NC}"
    exit 1
fi
