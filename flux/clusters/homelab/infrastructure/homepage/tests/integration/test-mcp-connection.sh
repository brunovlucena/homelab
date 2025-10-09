#!/bin/bash

# 🔌 MCP Connection Test
# Tests the connection between homepage API, agent-sre, and MCP server

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
AGENT_DIRECT_URL="${AGENT_DIRECT_URL:-http://localhost:31081}"
API_ENDPOINT="${API_BASE_URL}/api/v1/agent-sre"

print_header() {
    echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_failure() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

main() {
    print_header "🔌 MCP Connection Test Suite"
    
    # Test 1: Direct Agent Access
    print_header "Test 1: Direct Agent Access"
    print_info "Testing direct connection to agent-sre..."
    
    RESPONSE=$(curl -s -w "\n%{http_code}" --max-time 10 "${AGENT_DIRECT_URL}/health" 2>&1)
    HTTP_CODE=$(echo "$RESPONSE" | tail -n 1)
    
    if [ "$HTTP_CODE" == "200" ]; then
        print_success "Direct agent access working"
    else
        print_warning "Direct agent not accessible (HTTP $HTTP_CODE)"
        print_info "This is expected in production where NodePort is not exposed"
    fi
    
    # Test 2: API Proxy to Agent
    print_header "Test 2: API Proxy to Agent"
    print_info "Testing homepage API proxy to agent-sre..."
    
    RESPONSE=$(curl -s -w "\n%{http_code}" --max-time 10 "${API_ENDPOINT}/health" 2>&1)
    HTTP_CODE=$(echo "$RESPONSE" | tail -n 1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    
    if [ "$HTTP_CODE" == "200" ]; then
        print_success "API proxy to agent working"
        print_info "Response: $BODY"
    else
        print_failure "API proxy to agent failed (HTTP $HTTP_CODE)"
        print_info "Response: $BODY"
        exit 1
    fi
    
    # Test 3: Agent Status with MCP Info
    print_header "Test 3: Agent Status & MCP Server Info"
    print_info "Getting agent status to check MCP server connection..."
    
    RESPONSE=$(curl -s -w "\n%{http_code}" --max-time 10 "${API_ENDPOINT}/status" 2>&1)
    HTTP_CODE=$(echo "$RESPONSE" | tail -n 1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    
    if [ "$HTTP_CODE" == "200" ]; then
        print_success "Status endpoint working"
        echo ""
        echo "Full Status:"
        echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
        echo ""
        
        # Parse MCP server info
        MCP_STATUS=$(echo "$BODY" | jq -r '.mcp_server.status' 2>/dev/null || echo "not_found")
        MCP_URL=$(echo "$BODY" | jq -r '.mcp_server.url' 2>/dev/null || echo "not_found")
        
        if [ "$MCP_STATUS" != "not_found" ] && [ "$MCP_STATUS" != "null" ]; then
            if [ "$MCP_STATUS" == "healthy" ]; then
                print_success "MCP Server is healthy"
                print_info "MCP URL: $MCP_URL"
            else
                print_warning "MCP Server status: $MCP_STATUS"
                print_info "MCP URL: $MCP_URL"
            fi
        else
            print_warning "MCP Server info not available in status"
        fi
    else
        print_failure "Status endpoint failed (HTTP $HTTP_CODE)"
    fi
    
    # Test 4: MCP Chat Endpoint
    print_header "Test 4: MCP Chat Functionality"
    print_info "Testing chat via MCP protocol..."
    
    PAYLOAD='{"message": "Test MCP connection - what is Kubernetes?", "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"}'
    RESPONSE=$(curl -s -w "\n%{http_code}" --max-time 30 \
        -X POST "${API_ENDPOINT}/mcp/chat" \
        -H "Content-Type: application/json" \
        -d "$PAYLOAD" 2>&1)
    HTTP_CODE=$(echo "$RESPONSE" | tail -n 1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    
    if [ "$HTTP_CODE" == "200" ]; then
        print_success "MCP chat endpoint working"
        
        RESPONSE_TEXT=$(echo "$BODY" | jq -r '.response' 2>/dev/null || echo "")
        if [ -n "$RESPONSE_TEXT" ] && [ "$RESPONSE_TEXT" != "null" ]; then
            echo ""
            print_info "Question: Test MCP connection - what is Kubernetes?"
            echo -e "${GREEN}Response:${NC} ${RESPONSE_TEXT:0:200}..."
            echo ""
        fi
        
        # Check response metadata
        MODEL=$(echo "$BODY" | jq -r '.model' 2>/dev/null || echo "")
        SOURCES=$(echo "$BODY" | jq -r '.sources[]' 2>/dev/null || echo "")
        
        if [ -n "$MODEL" ] && [ "$MODEL" != "null" ]; then
            print_info "Model used: $MODEL"
        fi
        if [ -n "$SOURCES" ]; then
            print_info "Sources: $SOURCES"
        fi
    elif [ "$HTTP_CODE" == "502" ]; then
        print_warning "MCP endpoint returned 502 - MCP server may be unavailable"
        print_info "Response: $BODY"
    elif [ "$HTTP_CODE" == "503" ]; then
        print_warning "MCP endpoint returned 503 - Service temporarily unavailable"
        print_info "Response: $BODY"
    else
        print_failure "MCP chat endpoint failed (HTTP $HTTP_CODE)"
        print_info "Response: $BODY"
    fi
    
    # Test 5: Direct vs MCP Comparison
    print_header "Test 5: Direct vs MCP Mode Comparison"
    
    # Test direct mode
    print_info "Testing direct mode (no MCP)..."
    PAYLOAD='{"message": "What is a Kubernetes pod?", "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"}'
    RESPONSE_DIRECT=$(curl -s -w "\n%{http_code}" --max-time 30 \
        -X POST "${API_ENDPOINT}/chat" \
        -H "Content-Type: application/json" \
        -d "$PAYLOAD" 2>&1)
    HTTP_CODE_DIRECT=$(echo "$RESPONSE_DIRECT" | tail -n 1)
    
    # Test MCP mode
    print_info "Testing MCP mode..."
    RESPONSE_MCP=$(curl -s -w "\n%{http_code}" --max-time 30 \
        -X POST "${API_ENDPOINT}/mcp/chat" \
        -H "Content-Type: application/json" \
        -d "$PAYLOAD" 2>&1)
    HTTP_CODE_MCP=$(echo "$RESPONSE_MCP" | tail -n 1)
    
    echo ""
    if [ "$HTTP_CODE_DIRECT" == "200" ] && [ "$HTTP_CODE_MCP" == "200" ]; then
        print_success "Both direct and MCP modes working"
        print_info "Direct mode: ✓"
        print_info "MCP mode: ✓"
    elif [ "$HTTP_CODE_DIRECT" == "200" ]; then
        print_warning "Only direct mode working"
        print_info "Direct mode: ✓"
        print_info "MCP mode: ✗ (fallback available)"
    elif [ "$HTTP_CODE_MCP" == "200" ]; then
        print_success "MCP mode working"
        print_info "Direct mode: ✗"
        print_info "MCP mode: ✓"
    else
        print_failure "Neither mode working properly"
    fi
    
    # Test 6: Connection Chain Verification
    print_header "Test 6: Connection Chain Verification"
    print_info "Verifying the complete connection chain..."
    echo ""
    
    echo "Connection Chain:"
    echo "  1. Frontend → Homepage API"
    echo "  2. Homepage API → Agent-SRE Service"
    echo "  3. Agent-SRE → MCP Server (optional)"
    echo "  4. MCP Server → Ollama/LLM"
    echo ""
    
    # Check each link
    print_info "Checking Homepage API..."
    API_HEALTH=$(curl -s --max-time 5 "${API_BASE_URL}/health" 2>&1)
    if echo "$API_HEALTH" | grep -q "healthy\|ok"; then
        print_success "Homepage API: ✓"
    else
        print_warning "Homepage API: ✗"
    fi
    
    print_info "Checking Agent-SRE (via proxy)..."
    AGENT_HEALTH=$(curl -s --max-time 5 "${API_ENDPOINT}/health" 2>&1)
    if echo "$AGENT_HEALTH" | grep -q "healthy\|ok"; then
        print_success "Agent-SRE (via proxy): ✓"
    else
        print_failure "Agent-SRE (via proxy): ✗"
    fi
    
    print_info "Checking MCP Server (via status)..."
    STATUS_RESPONSE=$(curl -s --max-time 5 "${API_ENDPOINT}/status" 2>&1)
    MCP_STATUS=$(echo "$STATUS_RESPONSE" | jq -r '.mcp_server.status' 2>/dev/null || echo "unknown")
    if [ "$MCP_STATUS" == "healthy" ]; then
        print_success "MCP Server: ✓"
    elif [ "$MCP_STATUS" == "unknown" ]; then
        print_warning "MCP Server: ? (status unknown)"
    else
        print_warning "MCP Server: ✗ (status: $MCP_STATUS)"
    fi
    
    # Summary
    print_header "📊 MCP Connection Summary"
    echo "API Endpoint: ${API_ENDPOINT}"
    echo "Direct Agent: ${AGENT_DIRECT_URL}"
    echo ""
    
    if [ "$HTTP_CODE" == "200" ] && [ "$HTTP_CODE_MCP" == "200" ]; then
        echo -e "${GREEN}✓ MCP connection fully functional${NC}"
        echo "  - API proxy working"
        echo "  - Agent-SRE accessible"
        echo "  - MCP mode operational"
        exit 0
    elif [ "$HTTP_CODE" == "200" ]; then
        echo -e "${YELLOW}⚠ Partial functionality${NC}"
        echo "  - API proxy working"
        echo "  - Agent-SRE accessible"
        echo "  - MCP mode unavailable (direct mode available)"
        exit 0
    else
        echo -e "${RED}✗ Connection issues detected${NC}"
        echo "  Please check agent-sre deployment"
        exit 1
    fi
}

# Check dependencies
if ! command -v curl &> /dev/null; then
    echo -e "${RED}Error: curl is required${NC}"
    exit 1
fi

if ! command -v jq &> /dev/null; then
    echo -e "${YELLOW}Warning: jq not installed. JSON parsing limited${NC}"
fi

# Run tests
main "$@"

