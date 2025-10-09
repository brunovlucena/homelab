#!/bin/bash

# 🧪 Run All Tests for Chatbot Agent-SRE Integration

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PROJECT_ROOT="/Users/brunolucena/workspace/bruno/repos/homelab/flux/clusters/homelab/infrastructure/homepage"

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

# Test counters
TOTAL_SUITES=0
PASSED_SUITES=0
FAILED_SUITES=0

run_test_suite() {
    local name=$1
    local command=$2
    
    ((TOTAL_SUITES++))
    print_header "$name"
    
    if eval "$command"; then
        print_success "$name passed"
        ((PASSED_SUITES++))
        return 0
    else
        print_failure "$name failed"
        ((FAILED_SUITES++))
        return 1
    fi
}

main() {
    print_header "🧪 Running All Tests - Chatbot Agent-SRE Integration"
    
    # Check dependencies
    print_info "Checking dependencies..."
    if ! command -v go &> /dev/null; then
        print_failure "Go not installed"
        exit 1
    fi
    print_success "Go found: $(go version)"
    
    if ! command -v npm &> /dev/null; then
        print_failure "npm not installed"
        exit 1
    fi
    print_success "npm found: $(npm --version)"
    
    if ! command -v curl &> /dev/null; then
        print_failure "curl not installed"
        exit 1
    fi
    print_success "curl found"
    
    # Backend Unit Tests
    run_test_suite "Backend Unit Tests (Go)" \
        "cd ${PROJECT_ROOT}/api && go test -v ./handlers/ -run TestAgentSRE" || true
    
    # Integration Tests (if services are running)
    print_header "Integration Tests"
    print_info "Checking if services are running..."
    
    if curl -s --max-time 5 http://localhost:8080/health > /dev/null 2>&1; then
        print_success "Homepage API is running"
        
        run_test_suite "Agent-SRE Integration Test" \
            "cd ${PROJECT_ROOT}/tests/integration && ./test-agent-sre-integration.sh" || true
        
        run_test_suite "MCP Connection Test" \
            "cd ${PROJECT_ROOT}/tests/integration && ./test-mcp-connection.sh" || true
    else
        print_info "Homepage API not running - skipping integration tests"
        print_info "To run integration tests, start the services with: docker-compose up -d"
    fi
    
    # Summary
    print_header "📊 Test Summary"
    echo "Total Test Suites: $TOTAL_SUITES"
    echo -e "${GREEN}Passed: $PASSED_SUITES${NC}"
    echo -e "${RED}Failed: $FAILED_SUITES${NC}"
    echo ""
    
    if [ $FAILED_SUITES -eq 0 ]; then
        echo -e "${GREEN}✓ All tests passed!${NC}"
        exit 0
    else
        echo -e "${RED}✗ Some tests failed${NC}"
        exit 1
    fi
}

# Frontend tests note
print_header "📝 Frontend Tests"
print_info "Frontend tests require Jest to be installed"
print_info "To run frontend tests manually:"
echo ""
echo "  cd ${PROJECT_ROOT}/frontend"
echo "  npm install"
echo "  npm test"
echo ""

# Run all tests
main "$@"

