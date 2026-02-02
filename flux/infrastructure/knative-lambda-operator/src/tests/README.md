# 🧪 Knative Lambda Testing Suite

This directory contains all tests for the knative-lambda project, organized by test type and domain.

## 📁 Directory Structure

```
tests/
├── unit/              # Unit tests organized by domain
│   ├── backend/       # Backend API tests (BACKEND-001 to BACKEND-012)
│   ├── security/      # Security pentesting tests (SEC-001 to SEC-010)
│   ├── sre/          # SRE reliability tests (SRE-001 to SRE-019)
│   ├── devops/       # DevOps automation tests (DEVOPS-001 to DEVOPS-008)
│   ├── observability/ # Observability tests
│   ├── resilience/    # Resilience and rate limiting tests
│   ├── errors/        # Error handling tests
│   └── handler/       # Event handler tests
├── integration/       # Integration tests requiring cluster access
├── e2e/              # End-to-end tests (Python-based)
├── load/             # Load and performance tests (K6 + Python)
└── testutils/        # Shared test utilities and helpers
```

## 🎯 Test Categories

### Unit Tests (`tests/unit/`)
Fast, isolated tests that don't require external dependencies. Run with:
```bash
make test-unit
```

#### By Domain:
- **Backend** (`backend/`): API handlers, CloudEvent processing, job lifecycle
- **Security** (`security/`): Pentesting scenarios, authentication, authorization
- **SRE** (`sre/`): Reliability, disaster recovery, observability
- **DevOps** (`devops/`): CI/CD, GitOps, infrastructure automation
- **Observability** (`observability/`): Metrics, tracing, logging
- **Resilience** (`resilience/`): Rate limiting, circuit breakers
- **Errors** (`errors/`): Error handling and recovery
- **Handler** (`handler/`): Event processing and routing

### Integration Tests (`tests/integration/`)
Tests that require cluster access or external services. Run with:
```bash
make test-integration
```

### E2E Tests (`tests/e2e/`)
End-to-end CloudEvent flow tests using Python. Run with:
```bash
make test-e2e
```

### Load Tests (`tests/load/`)
Performance and load tests using K6 and Python. Run with:
```bash
make test-load         # Run all load tests
make test-load-k6      # Run K6 tests only
make test-load-python  # Run Python async tests only
```

## 🚀 Running Tests

### Quick Test Commands

```bash
# Run all tests (lint + unit)
make test

# Run specific test categories
make test-unit           # All unit tests
make test-security       # Security pentesting tests
make test-backend        # Backend user story tests
make test-sre           # SRE reliability tests
make test-devops        # DevOps automation tests
make test-integration   # Integration tests

# Run QA tests
make test-e2e           # E2E CloudEvents tests
make test-load          # Load and performance tests
make test-qa-all        # All QA tests (E2E + Load)

# Run comprehensive test suite
make test-all-stories   # All user story tests
```

### Test Flags

```bash
# Run tests with verbose output
go test -v ./tests/unit/...

# Run tests with coverage
go test -cover ./tests/unit/...

# Run tests with race detection
go test -race ./tests/unit/...

# Run specific test
go test -v ./tests/unit/security -run TestSEC001

# Run tests in short mode (skip long-running tests)
go test -short ./tests/unit/...
```

## 📊 Test Coverage

Generate coverage reports:

```bash
# Generate coverage for all tests
make test-unit

# View coverage in browser
go tool cover -html=coverage.out
```

## 🛠️ Test Utilities

The `testutils` package provides common helpers:

```go
import "github.com/bruno/knative-lambda/tests/testutils"

// Setup test environment
testutils.SetupTestEnvironment(t)

// Get test context with timeout
ctx, cancel := testutils.GetTestContext(30 * time.Second)
defer cancel()

// Skip test in short mode
testutils.SkipIfShort(t, "requires cluster access")

// Mock CloudEvent data
data := testutils.MockCloudEventData("build")
```

## 🔍 Test Organization Guidelines

### Unit Tests
- **Fast**: Should complete in milliseconds
- **Isolated**: No external dependencies
- **Focused**: Test one thing at a time
- **Package**: Use `_test.go` suffix, same package as code

### Integration Tests
- **Realistic**: Use real services (Redis, RabbitMQ, K8s)
- **Tagged**: Use `//go:build integration` build tag
- **Cleanup**: Always clean up resources
- **Documentation**: Document setup requirements

### E2E Tests
- **Full Flow**: Test complete user scenarios
- **Cluster**: Require live cluster
- **Data**: Use realistic test data
- **Assertions**: Verify end-to-end behavior

### Load Tests
- **Performance**: Focus on throughput and latency
- **Scalability**: Test under various loads
- **Metrics**: Collect detailed performance metrics
- **Reports**: Generate visual reports

## 📝 Writing New Tests

### Unit Test Template

```go
package mypackage_test

import (
	"testing"
	"github.com/bruno/knative-lambda/tests/testutils"
)

func TestMyFeature(t *testing.T) {
	// Setup
	testutils.SetupTestEnvironment(t)
	defer testutils.CleanupTestEnvironment(t)
	
	// Test cases
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "valid input",
			input: "test",
			want:  "expected",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test implementation
		})
	}
}
```

### Integration Test Template

```go
//go:build integration

package integration_test

import (
	"testing"
	"github.com/bruno/knative-lambda/tests/testutils"
)

func TestIntegration(t *testing.T) {
	testutils.SkipIfShort(t, "integration test")
	
	// Setup cluster resources
	// Run test
	// Cleanup
}
```

## 🏷️ Test Conventions

- **Naming**: `Test<FeatureName>_<Scenario>`
- **Subtests**: Use `t.Run()` for table-driven tests
- **Helpers**: Mark helper functions with `t.Helper()`
- **Cleanup**: Use `t.Cleanup()` or defer for cleanup
- **Errors**: Check errors explicitly, don't ignore
- **Messages**: Provide clear failure messages

## 🐛 Debugging Tests

```bash
# Run single test with verbose output
go test -v ./tests/unit/security -run TestSEC001_AuthenticationBypass

# Run with debug logging
LOG_LEVEL=debug go test -v ./tests/unit/...

# Run with race detector
go test -race ./tests/unit/...

# Run with timeout
go test -timeout 5m ./tests/integration/...
```

## 📚 Additional Resources

- [Go Testing](https://golang.org/pkg/testing/)
- [Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [K6 Documentation](https://k6.io/docs/)
- [Pytest Documentation](https://docs.pytest.org/)

## 🤝 Contributing

When adding new tests:

1. Place in appropriate directory (`unit/`, `integration/`, `e2e/`, `load/`)
2. Follow naming conventions
3. Add test documentation
4. Update this README if adding new test categories
5. Ensure tests pass in CI/CD pipeline

