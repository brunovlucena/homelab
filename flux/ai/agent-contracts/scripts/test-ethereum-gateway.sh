#!/bin/bash
#
# Test script for Cloudflare Ethereum Gateway
# Verifies connectivity and JSON-RPC functionality
#
# Usage:
#   ./scripts/test-ethereum-gateway.sh [gateway-url]
#
# Default gateway: https://ethereum.lucena.cloud

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default gateway URL (Cloudflare format: /v1/mainnet)
GATEWAY_URL="${1:-https://ethereum.lucena.cloud/v1/mainnet}"

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}🧪 Testing Cloudflare Ethereum Gateway${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}Gateway URL: ${GATEWAY_URL}${NC}"
echo ""

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Function to test JSON-RPC method
test_rpc_method() {
    local method=$1
    local params=$2
    local description=$3
    
    echo -e "${BLUE}Testing: ${description}${NC}"
    echo -e "  Method: ${method}"
    echo -e "  Params: ${params}"
    
    local response=$(curl -s -X POST "${GATEWAY_URL}" \
        -H "Content-Type: application/json" \
        -d "{
            \"jsonrpc\": \"2.0\",
            \"method\": \"${method}\",
            \"params\": ${params},
            \"id\": 1
        }" 2>&1)
    
    if echo "$response" | grep -q '"error"'; then
        echo -e "  ${RED}✗ FAILED${NC}"
        echo -e "  ${RED}Error: $(echo "$response" | grep -o '"message":"[^"]*"' | cut -d'"' -f4)${NC}"
        ((TESTS_FAILED++))
        return 1
    elif echo "$response" | grep -q '"result"'; then
        echo -e "  ${GREEN}✓ PASSED${NC}"
        local result=$(echo "$response" | grep -o '"result":"[^"]*"' | cut -d'"' -f4 || echo "$response" | grep -o '"result":[^,}]*')
        echo -e "  ${GREEN}Result: ${result}${NC}"
        ((TESTS_PASSED++))
        return 0
    else
        echo -e "  ${RED}✗ FAILED - Invalid response${NC}"
        echo -e "  ${RED}Response: ${response}${NC}"
        ((TESTS_FAILED++))
        return 1
    fi
    echo ""
}

# Test 1: Basic connectivity (JSON-RPC endpoints typically return 405 for GET, which is fine)
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}Test 1: Basic Connectivity${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "${GATEWAY_URL}" || echo "000")
if [[ "$HTTP_CODE" =~ ^(200|405|400|401)$ ]]; then
    echo -e "${GREEN}✓ Gateway is reachable (HTTP ${HTTP_CODE})${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${YELLOW}⚠ Gateway returned HTTP ${HTTP_CODE} (might be normal for JSON-RPC)${NC}"
    echo -e "${YELLOW}  Proceeding with JSON-RPC tests...${NC}"
fi
echo ""

# Test 2: eth_blockNumber - Get current block number
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}Test 2: eth_blockNumber${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
test_rpc_method "eth_blockNumber" "[]" "Get current block number"
echo ""

# Test 3: eth_getCode - Get contract bytecode (used by contract_fetcher)
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}Test 3: eth_getCode (Contract Bytecode)${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
# Using USDC contract address as test
USDC_ADDRESS="0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"
test_rpc_method "eth_getCode" "[\"${USDC_ADDRESS}\", \"latest\"]" "Get bytecode for USDC contract (used by contract_fetcher)"
echo ""

# Test 4: eth_getBalance - Get account balance
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}Test 4: eth_getBalance${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
# Using Vitalik's address as test
VITALIK_ADDRESS="0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045"
test_rpc_method "eth_getBalance" "[\"${VITALIK_ADDRESS}\", \"latest\"]" "Get balance for test address"
echo ""

# Test 5: net_version - Get network ID
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}Test 5: net_version${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
test_rpc_method "net_version" "[]" "Get network ID (should be 1 for mainnet)"
echo ""

# Summary
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}📊 Test Summary${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}Passed: ${TESTS_PASSED}${NC}"
echo -e "${RED}Failed: ${TESTS_FAILED}${NC}"
echo ""

TOTAL_TESTS=$((TESTS_PASSED + TESTS_FAILED))
if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All tests passed! Gateway is ready to use.${NC}"
    echo ""
    echo -e "${YELLOW}Next steps:${NC}"
    echo -e "  1. Ensure DNS for ethereum.lucena.cloud is properly configured in Cloudflare"
    echo -e "  2. Update Kubernetes secret with: kubectl patch secret agent-contracts-secrets -n ai --type='json' -p='[{\"op\": \"replace\", \"path\": \"/data/ETHEREUM_RPC_URL\", \"value\": \"'$(echo -n "${GATEWAY_URL}" | base64)'\"}]'"
    echo -e "  3. Restart contract-fetcher pods to pick up the new RPC URL"
    echo -e "  4. Monitor logs: kubectl logs -f -n ai -l app.kubernetes.io/component=fetcher"
    exit 0
else
    echo -e "${RED}❌ Some tests failed. Please check the gateway configuration.${NC}"
    exit 1
fi
