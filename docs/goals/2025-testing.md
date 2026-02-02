# 🧪 Testing & Quality Goals 2025

## Overview

Testing coverage and quality assurance targets for all homelab components.

---

## 📊 Current State

### Unit Test Coverage

| Component | Test Files | Coverage | Status |
|-----------|-----------|----------|--------|
| agent-bruno | 1 | ~30% | ⚠️ |
| agent-redteam | 1 | ~40% | ⚠️ |
| agent-blueteam | 1 | ~35% | ⚠️ |
| agent-contracts | 4 | ~45% | ⚠️ |
| agent-medical | 3 | ~50% | ✅ |
| agent-store-multibrands | 2 | ~20% | ⚠️ |
| agent-tools | 0 | 0% | ❌ |
| agent-restaurant | 0 | 0% | ❌ |
| agent-pos-edge | 0 | 0% | ❌ |
| agent-chat | 0 | 0% | ❌ |
| agent-rpg | 0 | 0% | ❌ |
| agent-devsecops | 0 | 0% | ❌ |

### K6 Test Coverage

| Component | Smoke | Load | Stress | E2E |
|-----------|-------|------|--------|-----|
| knative-lambda-operator | ✅ | ✅ | ✅ | ✅ |
| agent-bruno | ✅ | ✅ | ❌ | ❌ |
| agent-redteam | ✅ | ✅ | ✅ | ✅ |
| agent-blueteam | ✅ | ❌ | ❌ | ✅ |
| agent-contracts | ✅ | ❌ | ❌ | ✅ |
| agent-restaurant | ✅ | ✅ | ❌ | ✅ |
| agent-pos-edge | ✅ | ✅ | ❌ | ❌ |
| agent-store-multibrands | ✅ | ❌ | ❌ | ❌ |
| agent-chat | ❌ | ✅ | ✅ | ❌ |
| demo-mag7-battle | ✅ | ❌ | ❌ | ✅ |

---

## 🎯 2025 Targets

### Unit Testing

| Target | Q1 | Q2 | Q3 | Q4 |
|--------|----|----|----|----|
| Overall coverage | 40% | 60% | 75% | 80% |
| Agents with tests | 6/12 | 10/12 | 12/12 | 12/12 |
| CI enforcement | Soft | Soft | Hard | Hard |

**Actions:**
- [ ] Add pytest to all agent CI/CD pipelines
- [ ] Implement coverage gates (fail < 60%)
- [ ] Create test fixtures library
- [ ] Add mutation testing for critical paths

### K6 Load Testing

| Target | Q1 | Q2 | Q3 | Q4 |
|--------|----|----|----|----|
| Smoke tests | 100% | 100% | 100% | 100% |
| Load tests | 50% | 75% | 90% | 100% |
| Stress tests | 25% | 50% | 75% | 100% |
| E2E tests | 40% | 60% | 80% | 100% |

**Actions:**
- [ ] Create K6 test templates
- [ ] Implement scheduled test runs
- [ ] Add K6 results to Grafana
- [ ] Create performance baselines

### Integration Testing

| Target | Current | Goal |
|--------|---------|------|
| API integration tests | Partial | 90% |
| CloudEvent flow tests | Minimal | 80% |
| Database tests | None | 70% |
| External API mocks | Partial | 100% |

---

## 📋 Test Standards

### Unit Test Requirements

```python
# Required for all Python agents
- pytest >= 8.0
- pytest-asyncio (for async code)
- pytest-cov (coverage reports)
- pytest-mock (mocking)
- hypothesis (property testing)
```

### K6 Test Structure

```
k8s/tests/
├── k6-smoke.yaml          # Basic health checks
├── k6-load.yaml           # Normal load simulation
├── k6-stress.yaml         # Breaking point tests
├── k6-e2e.yaml            # End-to-end flows
└── k6-sre-metrics.yaml    # SRE metric validation
```

### Test Naming Convention

```
test_<component>_<function>_<scenario>.py
k6-<type>-<description>.yaml
```

---

## 🔄 CI/CD Integration

### Pipeline Test Stages

```yaml
stages:
  - lint           # Ruff, MyPy
  - unit-tests     # pytest --cov
  - build          # Docker build
  - smoke-tests    # K6 smoke
  - integration    # API tests
  - deploy-staging # Staging deploy
  - e2e-tests      # Full flow tests
  - deploy-prod    # Production deploy
```

### Quality Gates

| Gate | Threshold | Action |
|------|-----------|--------|
| Unit coverage | >= 60% | Block merge |
| Linting | 0 errors | Block merge |
| Smoke tests | 100% pass | Block deploy |
| Load tests | P95 < SLO | Warn |
| Security scan | 0 critical | Block deploy |

---

## 📊 Testing Dashboard Requirements

### Required Metrics

- [ ] Test coverage trend per component
- [ ] Test execution time trends
- [ ] Flaky test identification
- [ ] K6 performance trends
- [ ] Error rate by test type

### Alerting

- [ ] Coverage drop > 5%
- [ ] P95 latency increase > 20%
- [ ] Test failure rate > 5%
- [ ] Flaky test threshold

---

## 🗓️ Milestones

| Milestone | Date | Description |
|-----------|------|-------------|
| Test templates | Jan 31 | Create reusable test templates |
| 50% coverage | Mar 31 | Achieve 50% unit test coverage |
| K6 automation | Jun 30 | Automated K6 runs in CI |
| 80% coverage | Sep 30 | Achieve 80% unit test coverage |
| Full E2E | Dec 31 | Complete E2E test suite |

---

## 📝 Test Documentation

### Required per Component

- [ ] Test README explaining test structure
- [ ] Fixtures documentation
- [ ] Mock setup instructions
- [ ] Local test run guide

### Shared Resources

- [ ] Testing best practices guide
- [ ] K6 scripting cookbook
- [ ] CI/CD testing patterns
- [ ] Performance testing guide
