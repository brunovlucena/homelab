# 🧪 QA Engineering - User Stories

This directory contains QA Engineering user stories and test specifications for the Knative Lambda project.

## 📋 User Stories

| Story | Title | Priority | Story Points | Status |
|-------|-------|----------|--------------|--------|
| **QA-001** | E2E CloudEvents Testing | P0 | 13 | ✅ Active |
| **QA-002** | Load and Performance Testing | P0 | 13 | ✅ Active |

---

## 🎯 QA-001: E2E CloudEvents Testing

**Purpose:** Validate complete CloudEvents processing pipeline end-to-end

### Test Files
- `tests/e2e/qa_001_e2e_cloudevents_test.py` - Main E2E test suite
- `tests/e2e/conftest.py` - Pytest fixtures and configuration
- `tests/e2e/pytest.ini` - Pytest configuration

### Test Coverage
- ✅ AC1: Build event complete flow
- ✅ AC2: Parser event autoscaling  
- ✅ AC3: Service deletion and cleanup
- ✅ AC4: Job lifecycle management
- ✅ AC5: Complete lifecycle tests

### Running Tests

```bash
# Run all E2E tests
make test-e2e ENV=dev

# Run specific story
./scripts/run-qa-e2e-tests.sh --story 001

# Run specific marker
./scripts/run-qa-e2e-tests.sh --marker build
./scripts/run-qa-e2e-tests.sh --marker parser
./scripts/run-qa-e2e-tests.sh --marker delete

# Run specific test
./scripts/run-qa-e2e-tests.sh --keyword "test_build_event_complete_flow"

# Verbose output
./scripts/run-qa-e2e-tests.sh --verbose
```

### Prerequisites
- Kubernetes cluster access
- Port-forward broker: `make pf-broker ENV=dev`
- Python 3.9+ with uv
- kubectl configured

---

## 🚀 QA-002: Load and Performance Testing

**Purpose:** Validate system performance under high load conditions

### Test Files
- `tests/load/python/qa_002_load_performance_test.py` - Python async load tests
- `tests/load/k6/qa_002_builder_load_test.js` - K6 build event load tests
- `tests/load/k6/qa_002_parser_load_test.js` - K6 parser event load tests

### Test Coverage
- ✅ AC1: Build event load testing (100-500 concurrent)
- ✅ AC2: Parser event load testing (500-1500 concurrent)
- ✅ AC3: HTTP direct load testing
- ✅ AC4: Stress testing
- ✅ AC5: Endurance testing

### Running Tests

```bash
# Run all load tests
make test-load ENV=dev

# Run Python async load tests
make test-load-python EVENT_TYPE=parser CONCURRENT=100 DURATION=300

# Run K6 load tests
make test-load-k6 EVENT_TYPE=parser

# Custom configuration
./scripts/run-qa-load-tests.sh \
  --type python \
  --event parser \
  --concurrent 200 \
  --duration 600 \
  --rampup 120
```

### Prerequisites
- Kubernetes cluster access
- Port-forward broker: `make pf-broker ENV=dev`
- Python 3.9+ with aiohttp
- k6 installed: `brew install k6`

### Load Test Scenarios

#### Python Async Tests
- Configurable concurrency (default: 100)
- Configurable duration (default: 300s)
- Rampup period (default: 60s)
- Event types: build, parser, delete

#### K6 Tests
- **Baseline:** 10 events/sec for 5 minutes
- **Ramp-up:** 10 → 200 events/sec over 10 minutes
- **Spike:** 500 events/sec burst for 2 minutes
- **Sustained:** 100 events/sec for 30 minutes
- **Parser specific:** Up to 1500 events/sec

---

## 📊 Performance Targets (SLOs)

| Metric | Target | Critical Threshold |
|--------|--------|-------------------|
| Event Processing Latency (p95) | < 3s | < 10s |
| Event Publishing Success Rate | > 99% | > 95% |
| Requests per Second | > 100 | > 50 |
| Build Job Creation Rate | > 50/sec | > 20/sec |
| Service Autoscaling Time | < 30s | < 60s |

---

## 🔍 Test Organization

### Naming Convention
All test files follow the pattern: `{team}_{number}_{description}_test.{ext}`

Examples:
- `qa_001_e2e_cloudevents_test.py` ← QA-001
- `qa_002_load_performance_test.py` ← QA-002
- `qa_002_builder_load_test.js` ← QA-002 (K6 variant)

This ensures **1:1 traceability** between:
- User story documentation (e.g., `QA-001-E2E-CloudEvents-Testing.md`)
- Test implementation (e.g., `qa_001_e2e_cloudevents_test.py`)
- Makefile targets (e.g., `make test-e2e`)
- CI/CD pipelines

### Test Markers
- `@pytest.mark.e2e` - End-to-end tests
- `@pytest.mark.build` - Build event tests
- `@pytest.mark.parser` - Parser event tests
- `@pytest.mark.delete` - Deletion tests
- `@pytest.mark.lifecycle` - Full lifecycle tests
- `@pytest.mark.slow` - Slow tests (> 5 minutes)

---

## 🚀 CI/CD Integration

Tests are integrated into the CI/CD pipeline:

- **Unit tests:** Run on every PR
- **Integration tests:** Run on every PR to develop/main
- **E2E tests:** Run nightly or on-demand
- **Load tests:** Run on-demand or weekly

---

## 📚 Related Documentation

- [Backend User Stories](../backend/README.md)
- [SRE User Stories](../sre/README.md)
- [DevOps User Stories](../devops/README.md)
- [Platform User Stories](../platform/README.md)

---

## 🛠️ Development

### Adding New Tests

1. Create user story document: `QA-XXX-Description.md`
2. Create test file: `qa_XXX_description_test.py`
3. Update this README
4. Add Makefile target if needed
5. Update CI/CD if needed

### Test Best Practices

- ✅ Follow naming convention for traceability
- ✅ Include story number in test file name
- ✅ Document acceptance criteria in tests
- ✅ Use appropriate markers
- ✅ Clean up resources after tests
- ✅ Make tests idempotent
- ✅ Handle timeouts gracefully
- ✅ Provide clear error messages

---

## 📞 Contact

For questions about QA tests:
- Check user story documentation first
- Review test implementation
- Contact team lead for clarification
