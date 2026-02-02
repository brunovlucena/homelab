#!/bin/bash
# =============================================================================
# Run All Tests Script
# =============================================================================
# Executes all test suites in the homelab testing framework
#
# Usage:
#   ./run-all-tests.sh [options]
#
# Options:
#   -u, --unit          Run only unit tests
#   -i, --integration   Run only integration tests
#   -s, --scripts       Run only script tests
#   -v, --verbose       Verbose output
#   -c, --coverage      Generate coverage reports
#   -f, --fast          Skip slow integration tests
#   -h, --help          Show this help message
# =============================================================================

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default options
RUN_UNIT=true
RUN_INTEGRATION=true
RUN_SCRIPTS=true
VERBOSE=false
COVERAGE=false
FAST=false

# Test results
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_SKIPPED=0

# =============================================================================
# Helper Functions
# =============================================================================

print_header() {
    echo ""
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
    echo ""
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

show_help() {
    head -n 20 "$0" | grep '^#' | sed 's/^# //g' | sed 's/^#//g'
}

# =============================================================================
# Parse Command Line Arguments
# =============================================================================

while [[ $# -gt 0 ]]; do
    case $1 in
        -u|--unit)
            RUN_UNIT=true
            RUN_INTEGRATION=false
            RUN_SCRIPTS=false
            shift
            ;;
        -i|--integration)
            RUN_UNIT=false
            RUN_INTEGRATION=true
            RUN_SCRIPTS=false
            shift
            ;;
        -s|--scripts)
            RUN_UNIT=false
            RUN_INTEGRATION=false
            RUN_SCRIPTS=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -c|--coverage)
            COVERAGE=true
            shift
            ;;
        -f|--fast)
            FAST=true
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# =============================================================================
# Pre-flight Checks
# =============================================================================

print_header "🔍 Pre-flight Checks"

# Check if we're in the right directory
if [ ! -f "../Makefile" ]; then
    print_error "Must be run from the tests/ directory"
    exit 1
fi
print_success "Running from correct directory"

# Check for required tools
MISSING_TOOLS=()

if ! command -v go >/dev/null 2>&1; then
    MISSING_TOOLS+=("go")
fi

if ! command -v bats >/dev/null 2>&1; then
    MISSING_TOOLS+=("bats-core")
fi

if [ ${#MISSING_TOOLS[@]} -gt 0 ]; then
    print_error "Missing required tools: ${MISSING_TOOLS[*]}"
    print_info "Install with: brew install ${MISSING_TOOLS[*]}"
    exit 1
fi
print_success "All required tools available"

# Check for optional tools
if $RUN_INTEGRATION; then
    if ! command -v docker >/dev/null 2>&1 || ! docker info >/dev/null 2>&1; then
        print_warning "Docker not available - integration tests will be skipped"
        RUN_INTEGRATION=false
    else
        print_success "Docker available for integration tests"
    fi
    
    if ! command -v kind >/dev/null 2>&1; then
        print_warning "Kind not available - some integration tests will be skipped"
    else
        print_success "Kind available for integration tests"
    fi
fi

# =============================================================================
# Run Go Unit Tests
# =============================================================================

if $RUN_UNIT; then
    print_header "🧪 Running Go Unit Tests"
    
    cd pulumi
    
    if $COVERAGE; then
        print_info "Running with coverage..."
        if $VERBOSE; then
            go test -v -cover -coverprofile=coverage.out ./...
        else
            go test -cover -coverprofile=coverage.out ./...
        fi
        
        # Generate coverage report
        go tool cover -html=coverage.out -o coverage.html
        print_success "Coverage report generated: pulumi/coverage.html"
        
        # Show coverage summary
        COVERAGE_PCT=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
        print_info "Total coverage: $COVERAGE_PCT"
    else
        if $VERBOSE; then
            go test -v ./...
        else
            go test ./...
        fi
    fi
    
    GO_TEST_EXIT=$?
    
    if [ $GO_TEST_EXIT -eq 0 ]; then
        print_success "Go unit tests passed"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        print_error "Go unit tests failed"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    
    cd ..
fi

# =============================================================================
# Run Shell Script Tests
# =============================================================================

if $RUN_SCRIPTS; then
    print_header "📜 Running Shell Script Tests"
    
    BATS_OPTS=""
    if $VERBOSE; then
        BATS_OPTS="-t"
    fi
    
    # Test setup-local-registry.sh
    print_info "Testing setup-local-registry.sh..."
    if bats $BATS_OPTS scripts/setup-local-registry.bats; then
        print_success "setup-local-registry.sh tests passed"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        print_error "setup-local-registry.sh tests failed"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    
    # Test bootstrap-github-app.sh
    print_info "Testing bootstrap-github-app.sh..."
    if bats $BATS_OPTS scripts/bootstrap-github-app.bats; then
        print_success "bootstrap-github-app.sh tests passed"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        print_error "bootstrap-github-app.sh tests failed"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
fi

# =============================================================================
# Run Integration Tests
# =============================================================================

if $RUN_INTEGRATION; then
    print_header "🔗 Running Integration Tests"
    
    if $FAST; then
        print_warning "Skipping slow integration tests (--fast mode)"
        TESTS_SKIPPED=$((TESTS_SKIPPED + 1))
    else
        BATS_OPTS=""
        if $VERBOSE; then
            BATS_OPTS="-t"
        fi
        
        print_info "Running cluster provisioning tests..."
        if [ -f "integration/cluster-provisioning.bats" ]; then
            if bats $BATS_OPTS integration/cluster-provisioning.bats; then
                print_success "Integration tests passed"
                TESTS_PASSED=$((TESTS_PASSED + 1))
            else
                print_error "Integration tests failed"
                TESTS_FAILED=$((TESTS_FAILED + 1))
            fi
        else
            print_warning "Integration tests not found, skipping"
            TESTS_SKIPPED=$((TESTS_SKIPPED + 1))
        fi
    fi
fi

# =============================================================================
# Test Summary
# =============================================================================

print_header "📊 Test Summary"

echo ""
echo "Test Results:"
echo "  Passed:  $TESTS_PASSED"
echo "  Failed:  $TESTS_FAILED"
echo "  Skipped: $TESTS_SKIPPED"
echo ""

TOTAL_TESTS=$((TESTS_PASSED + TESTS_FAILED))
if [ $TOTAL_TESTS -gt 0 ]; then
    SUCCESS_RATE=$((TESTS_PASSED * 100 / TOTAL_TESTS))
    echo "Success Rate: ${SUCCESS_RATE}%"
    echo ""
fi

# Generate test report
TIMESTAMP=$(date +"%Y-%m-%d_%H-%M-%S")
REPORT_FILE="test-report-${TIMESTAMP}.txt"

cat > "$REPORT_FILE" <<EOF
Homelab Test Suite Report
Generated: $(date)

Test Configuration:
  Unit Tests:        $RUN_UNIT
  Script Tests:      $RUN_SCRIPTS
  Integration Tests: $RUN_INTEGRATION
  Verbose:           $VERBOSE
  Coverage:          $COVERAGE
  Fast Mode:         $FAST

Results:
  Passed:  $TESTS_PASSED
  Failed:  $TESTS_FAILED
  Skipped: $TESTS_SKIPPED
  Total:   $TOTAL_TESTS
EOF

if [ $TOTAL_TESTS -gt 0 ]; then
    echo "  Success Rate: ${SUCCESS_RATE}%" >> "$REPORT_FILE"
fi

print_success "Test report saved: $REPORT_FILE"

# =============================================================================
# Exit with appropriate code
# =============================================================================

if [ $TESTS_FAILED -eq 0 ]; then
    print_success "All tests passed! 🎉"
    exit 0
else
    print_error "Some tests failed"
    exit 1
fi

