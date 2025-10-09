#!/bin/bash

# 🧪 Agent-SRE Integration Test Suite
# Tests the complete integration between homepage and agent-sre

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
API_ENDPOINT="${API_BASE_URL}/api/v1/agent-sre"
TIMEOUT=30

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Helper functions
print_header() {
    echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"
}

print_test() {
    echo -e "${YELLOW}▶ Test: $1${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
    ((TESTS_PASSED++))
}

print_failure() {
    echo -e "${RED}✗ $1${NC}"
    ((TESTS_FAILED++))
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

# Test function
run_test() {
    ((TESTS_RUN++))
    print_test "$1"
}

# Main test suite
main() {
    print_header "🤖 Agent-SRE Integration Test Suite"
    echo "Testing API: ${API_ENDPOINT}"
    echo "Timeout: ${TIMEOUT}s"
    echo ""

    # Test 1: Health Check
    run_test "Health Check"
    RESPONSE=$(curl -s -w "\n%{http_code}" --max-time $TIMEOUT "${API_ENDPOINT}/health")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n 1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    
    if [ "$HTTP_CODE" == "200" ]; then
        STATUS=$(echo "$BODY" | jq -r '.status' 2>/dev/null || echo "error")
        if [ "$STATUS" == "healthy" ] || [ "$STATUS" == "ok" ]; then
            print_success "Health check passed (HTTP $HTTP_CODE)"
            print_info "Response: $BODY"
        else
            print_failure "Health check returned unexpected status: $STATUS"
        fi
    else
        print_failure "Health check failed (HTTP $HTTP_CODE)"
        print_info "Response: $BODY"
    fi

    # Test 2: Ready Check
    run_test "Readiness Check"
    RESPONSE=$(curl -s -w "\n%{http_code}" --max-time $TIMEOUT "${API_ENDPOINT}/ready")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n 1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    
    if [ "$HTTP_CODE" == "200" ]; then
        print_success "Ready check passed (HTTP $HTTP_CODE)"
        print_info "Response: $BODY"
    else
        print_failure "Ready check failed (HTTP $HTTP_CODE)"
        print_info "Response: $BODY"
    fi

    # Test 3: Status Check
    run_test "Status Check (with MCP info)"
    RESPONSE=$(curl -s -w "\n%{http_code}" --max-time $TIMEOUT "${API_ENDPOINT}/status")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n 1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    
    if [ "$HTTP_CODE" == "200" ]; then
        print_success "Status check passed (HTTP $HTTP_CODE)"
        print_info "Response: $BODY"
        
        # Check for MCP server info
        MCP_STATUS=$(echo "$BODY" | jq -r '.mcp_server.status' 2>/dev/null || echo "not_found")
        if [ "$MCP_STATUS" != "not_found" ] && [ "$MCP_STATUS" != "null" ]; then
            print_info "MCP Server Status: $MCP_STATUS"
        fi
    else
        print_failure "Status check failed (HTTP $HTTP_CODE)"
        print_info "Response: $BODY"
    fi

    # Test 4: Direct Chat
    run_test "Direct Chat (no MCP)"
    PAYLOAD='{"message": "How do I check Kubernetes pod logs?", "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"}'
    RESPONSE=$(curl -s -w "\n%{http_code}" --max-time $TIMEOUT \
        -X POST "${API_ENDPOINT}/chat" \
        -H "Content-Type: application/json" \
        -d "$PAYLOAD")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n 1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    
    if [ "$HTTP_CODE" == "200" ]; then
        RESPONSE_TEXT=$(echo "$BODY" | jq -r '.response' 2>/dev/null || echo "")
        if [ -n "$RESPONSE_TEXT" ] && [ "$RESPONSE_TEXT" != "null" ]; then
            print_success "Direct chat passed (HTTP $HTTP_CODE)"
            print_info "Question: How do I check Kubernetes pod logs?"
            print_info "Answer: ${RESPONSE_TEXT:0:100}..."
        else
            print_failure "Direct chat returned empty response"
        fi
    else
        print_failure "Direct chat failed (HTTP $HTTP_CODE)"
        print_info "Response: $BODY"
    fi

    # Test 5: MCP Chat
    run_test "MCP Chat"
    PAYLOAD='{"message": "What are the best practices for monitoring in Kubernetes?", "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"}'
    RESPONSE=$(curl -s -w "\n%{http_code}" --max-time $TIMEOUT \
        -X POST "${API_ENDPOINT}/mcp/chat" \
        -H "Content-Type: application/json" \
        -d "$PAYLOAD")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n 1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    
    if [ "$HTTP_CODE" == "200" ]; then
        RESPONSE_TEXT=$(echo "$BODY" | jq -r '.response' 2>/dev/null || echo "")
        if [ -n "$RESPONSE_TEXT" ] && [ "$RESPONSE_TEXT" != "null" ]; then
            print_success "MCP chat passed (HTTP $HTTP_CODE)"
            print_info "Question: What are the best practices for monitoring?"
            print_info "Answer: ${RESPONSE_TEXT:0:100}..."
            
            # Check for sources
            SOURCES=$(echo "$BODY" | jq -r '.sources[]' 2>/dev/null || echo "")
            if [ -n "$SOURCES" ]; then
                print_info "Sources: $SOURCES"
            fi
        else
            print_failure "MCP chat returned empty response"
        fi
    elif [ "$HTTP_CODE" == "502" ] || [ "$HTTP_CODE" == "503" ]; then
        print_info "MCP server may be unavailable (HTTP $HTTP_CODE) - this is expected if MCP is not running"
    else
        print_failure "MCP chat failed (HTTP $HTTP_CODE)"
        print_info "Response: $BODY"
    fi

    # Test 6: Direct Log Analysis
    run_test "Direct Log Analysis"
    PAYLOAD='{"logs": "ERROR: Connection timeout to database\nERROR: Failed to authenticate user\nWARN: High memory usage detected", "context": "Production API Server"}'
    RESPONSE=$(curl -s -w "\n%{http_code}" --max-time $TIMEOUT \
        -X POST "${API_ENDPOINT}/analyze-logs" \
        -H "Content-Type: application/json" \
        -d "$PAYLOAD")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n 1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    
    if [ "$HTTP_CODE" == "200" ]; then
        ANALYSIS=$(echo "$BODY" | jq -r '.analysis' 2>/dev/null || echo "")
        SEVERITY=$(echo "$BODY" | jq -r '.severity' 2>/dev/null || echo "")
        if [ -n "$ANALYSIS" ] && [ "$ANALYSIS" != "null" ]; then
            print_success "Direct log analysis passed (HTTP $HTTP_CODE)"
            print_info "Severity: $SEVERITY"
            print_info "Analysis: ${ANALYSIS:0:100}..."
            
            # Check for recommendations
            RECOMMENDATIONS=$(echo "$BODY" | jq -r '.recommendations[]' 2>/dev/null || echo "")
            if [ -n "$RECOMMENDATIONS" ]; then
                print_info "Recommendations available"
            fi
        else
            print_failure "Direct log analysis returned empty response"
        fi
    else
        print_failure "Direct log analysis failed (HTTP $HTTP_CODE)"
        print_info "Response: $BODY"
    fi

    # Test 7: MCP Log Analysis
    run_test "MCP Log Analysis"
    PAYLOAD='{"logs": "ERROR: OOMKilled\nERROR: Container memory limit exceeded", "context": "Production Pod"}'
    RESPONSE=$(curl -s -w "\n%{http_code}" --max-time $TIMEOUT \
        -X POST "${API_ENDPOINT}/mcp/analyze-logs" \
        -H "Content-Type: application/json" \
        -d "$PAYLOAD")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n 1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    
    if [ "$HTTP_CODE" == "200" ]; then
        ANALYSIS=$(echo "$BODY" | jq -r '.analysis' 2>/dev/null || echo "")
        if [ -n "$ANALYSIS" ] && [ "$ANALYSIS" != "null" ]; then
            print_success "MCP log analysis passed (HTTP $HTTP_CODE)"
            print_info "Analysis: ${ANALYSIS:0:100}..."
        else
            print_failure "MCP log analysis returned empty response"
        fi
    elif [ "$HTTP_CODE" == "502" ] || [ "$HTTP_CODE" == "503" ]; then
        print_info "MCP server may be unavailable (HTTP $HTTP_CODE)"
    else
        print_failure "MCP log analysis failed (HTTP $HTTP_CODE)"
        print_info "Response: $BODY"
    fi

    # Test 8: Error Handling - Invalid JSON
    run_test "Error Handling - Invalid JSON"
    RESPONSE=$(curl -s -w "\n%{http_code}" --max-time $TIMEOUT \
        -X POST "${API_ENDPOINT}/chat" \
        -H "Content-Type: application/json" \
        -d 'invalid json')
    HTTP_CODE=$(echo "$RESPONSE" | tail -n 1)
    
    if [ "$HTTP_CODE" == "400" ] || [ "$HTTP_CODE" == "500" ]; then
        print_success "Invalid JSON handled correctly (HTTP $HTTP_CODE)"
    else
        print_failure "Invalid JSON not handled properly (HTTP $HTTP_CODE)"
    fi

    # Test 9: Error Handling - Empty Message
    run_test "Error Handling - Empty Message"
    PAYLOAD='{"message": "", "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"}'
    RESPONSE=$(curl -s -w "\n%{http_code}" --max-time $TIMEOUT \
        -X POST "${API_ENDPOINT}/chat" \
        -H "Content-Type: application/json" \
        -d "$PAYLOAD")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n 1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    
    # Agent might accept or reject empty messages
    if [ "$HTTP_CODE" == "200" ] || [ "$HTTP_CODE" == "400" ]; then
        print_success "Empty message handled (HTTP $HTTP_CODE)"
    else
        print_failure "Empty message not handled properly (HTTP $HTTP_CODE)"
    fi

    # Print Summary
    print_header "📊 Test Summary"
    echo "Total Tests Run: $TESTS_RUN"
    echo -e "${GREEN}Tests Passed: $TESTS_PASSED${NC}"
    echo -e "${RED}Tests Failed: $TESTS_FAILED${NC}"
    echo ""
    
    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "${GREEN}✓ All tests passed!${NC}"
        exit 0
    else
        echo -e "${RED}✗ Some tests failed${NC}"
        exit 1
    fi
}

# Check dependencies
check_dependencies() {
    if ! command -v curl &> /dev/null; then
        echo -e "${RED}Error: curl is required but not installed${NC}"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        echo -e "${YELLOW}Warning: jq is not installed. JSON parsing will be limited${NC}"
    fi
}

# Run the tests
check_dependencies
main "$@"

