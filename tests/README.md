# Homelab Testing Suite

## 📋 Overview

This directory contains comprehensive test coverage for the homelab infrastructure, including:

- **Unit Tests**: Go unit tests for Pulumi infrastructure code
- **Integration Tests**: Shell script integration tests using BATS
- **Functional Tests**: End-to-end testing of cluster provisioning
- **Contract Tests**: Validation of infrastructure contracts and outputs

## 🗂️ Directory Structure

```
tests/
├── README.md                  # This file
├── pulumi/                    # Go unit tests for Pulumi code
│   ├── main_test.go          # Tests for main.go
│   └── fixtures/             # Test fixtures and mock data
├── scripts/                   # Shell script tests (BATS)
│   ├── setup-local-registry.bats
│   ├── bootstrap-github-app.bats
│   └── helpers/              # Test helper functions
├── integration/               # Integration tests
│   └── cluster-provisioning.bats
└── fixtures/                  # Shared test fixtures
    ├── kind-configs/         # Sample Kind configurations
    └── mock-clusters/        # Mock cluster data
```

## 🚀 Running Tests

### Prerequisites

```bash
# Install BATS (Bash Automated Testing System)
brew install bats-core

# Install BATS helper libraries
brew tap kaos/shell
brew install bats-assert bats-support

# Or install manually
git clone https://github.com/bats-core/bats-support test/test_helper/bats-support
git clone https://github.com/bats-core/bats-assert test/test_helper/bats-assert
```

### Run All Tests

```bash
# From repository root
make test

# Or manually
cd tests && ./run-all-tests.sh
```

### Run Specific Test Suites

```bash
# Go unit tests
cd tests/pulumi && go test -v ./...

# Shell script tests
cd tests/scripts && bats setup-local-registry.bats
cd tests/scripts && bats bootstrap-github-app.bats

# Integration tests
cd tests/integration && bats cluster-provisioning.bats
```

### Run with Coverage

```bash
# Go tests with coverage
cd tests/pulumi && go test -v -cover -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# BATS with verbose output
cd tests/scripts && bats -t setup-local-registry.bats
```

## 📊 Test Coverage Goals

| Component | Target Coverage | Status |
|-----------|----------------|--------|
| pulumi/main.go | 80%+ | ✅ |
| scripts/mac/setup-local-registry.sh | 90%+ | ✅ |
| scripts/bootstrap-github-app.sh | 90%+ | ✅ |
| Integration flows | 100% | ✅ |

## 🧪 Test Categories

### Unit Tests (pulumi/)
- Test individual functions in isolation
- Mock external dependencies
- Fast execution (< 1 second per test)
- No external dependencies required

### Integration Tests (scripts/)
- Test scripts with real Docker/Kind
- Validate command outputs
- Test error handling and edge cases
- Require Docker running

### Functional Tests (integration/)
- Test complete cluster provisioning flow
- Validate end-to-end scenarios
- Test with real clusters (cleanup required)
- Longer execution time (minutes)

## 🛠️ Writing New Tests

### Go Unit Tests

```go
func TestMyFunction(t *testing.T) {
    // Arrange
    input := "test-value"
    expected := "expected-output"
    
    // Act
    result := MyFunction(input)
    
    // Assert
    if result != expected {
        t.Errorf("Expected %s, got %s", expected, result)
    }
}
```

### BATS Tests

```bash
@test "script succeeds with valid input" {
    run ./script.sh valid-arg
    
    assert_success
    assert_output --partial "Expected output"
}

@test "script fails with invalid input" {
    run ./script.sh invalid-arg
    
    assert_failure
    assert_output --partial "Error message"
}
```

## 🔍 Debugging Failed Tests

### Enable Verbose Output

```bash
# Go tests
go test -v -run TestSpecificTest

# BATS tests
bats -t test-file.bats
```

### Use Test Fixtures

Test fixtures are located in `tests/fixtures/` and provide:
- Sample Kind configurations
- Mock cluster data
- Test environment variables
- Reusable test data

### Check Test Logs

```bash
# View test execution logs
cat tests/logs/test-run-$(date +%Y%m%d).log

# View specific test output
bats -o tap test-file.bats > test-output.tap
```

## 📈 Continuous Integration

Tests run automatically on:
- Every pull request
- Commits to main branch
- Nightly builds (integration tests)

CI Configuration: `.github/workflows/test.yml`

## 🤝 Contributing

When adding new features:

1. **Write tests first** (TDD approach)
2. **Ensure tests pass locally**
3. **Add test documentation**
4. **Update this README** if needed

### Code Review Checklist

- [ ] Tests cover happy path
- [ ] Tests cover error cases
- [ ] Tests are independent (no order dependency)
- [ ] Test names are descriptive
- [ ] Mocks/fixtures are properly cleaned up
- [ ] Tests run in CI pipeline

## 📚 Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [BATS Documentation](https://bats-core.readthedocs.io/)
- [Pulumi Testing Guide](https://www.pulumi.com/docs/guides/testing/)
- [Shell Script Testing Best Practices](https://github.com/bats-core/bats-core#best-practices)

