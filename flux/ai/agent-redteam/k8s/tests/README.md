# 🔴 Agent-Redteam K6 Test Suite

Comprehensive k6 test suite for simulating attacks and vulnerability assessments via agent-redteam.

## 📋 Test Overview

| Test | Purpose | Scenario | Duration |
|------|---------|----------|----------|
| **k6-smoke.yaml** | Basic functionality validation | Quick health and catalog checks | ~30s |
| **k6-attack-sequential.yaml** | Sequential attack simulation | Attacker testing exploits one by one | ~10m |
| **k6-attack-parallel.yaml** | Parallel attack simulation | Multiple attackers launching exploits simultaneously | ~2m |
| **k6-vulnerability-assessment.yaml** | Full vulnerability assessment | Security team performing comprehensive assessment | ~15m |
| **k6-random-exploit-chaos.yaml** | Chaos testing with random exploits | Unpredictable random attack patterns | ~4m |

## 🎯 Test Scenarios

### 1. Smoke Test (`k6-smoke.yaml`)
**Purpose**: Quick validation of agent-redteam functionality

**Tests**:
- Health check endpoint
- Readiness check endpoint
- Catalog API (list exploits, categories, severities)

**Metrics**:
- `smoke_health_check_success`: Health check success rate
- `smoke_catalog_success`: Catalog API success rate
- `smoke_response_time_ms`: Response time trends

**Thresholds**:
- Health check success > 95%
- Catalog success > 80%
- P95 response time < 5s

---

### 2. Sequential Attack (`k6-attack-sequential.yaml`)
**Purpose**: Simulate a methodical attacker testing exploits sequentially

**Scenario**: Attacker tests exploits one by one in order of severity (Critical → High → Medium)

**Exploit Sequence**:
1. `vuln-001` - Command Injection Git (CRITICAL)
2. `vuln-002` - Command Injection MinIO (CRITICAL)
3. `vuln-003` - Inline Code Execution (CRITICAL)
4. `blue-001` - SSRF (CRITICAL)
5. `blue-002` - Template Injection (CRITICAL)
6. `blue-005` - Path Traversal (HIGH)
7. `blue-006` - SA Token Exposure (HIGH)
8. `vuln-004` - RBAC Escalation (HIGH)
9. `vuln-013` - Receiver SA Escalation (MEDIUM)

**Metrics**:
- `attack_exploit_success`: Success rate
- `attack_exploit_blocked`: Blocked rate
- `attack_exploit_failed`: Failure rate
- `attack_exploit_duration_ms`: Execution duration

**Thresholds**:
- At least some exploits succeed or are blocked
- P95 duration < 30s

---

### 3. Parallel Attack (`k6-attack-parallel.yaml`)
**Purpose**: Simulate coordinated attack with multiple exploits running simultaneously

**Scenario**: Multiple attackers launching critical exploits in parallel

**Load Pattern**:
- Ramp up: 0 → 5 concurrent attacks (10s)
- Sustain: 5 concurrent attacks (30s)
- Ramp down: 5 → 0 (10s)

**Exploits**: Critical exploits only (`vuln-001`, `vuln-002`, `vuln-003`, `blue-001`, `blue-002`)

**Metrics**:
- `parallel_exploit_success`: Success rate
- `parallel_exploit_blocked`: Blocked rate
- `concurrent_attacks_total`: Total concurrent attacks

**Thresholds**:
- P95 duration < 60s (allows longer for parallel load)
- HTTP failure rate < 20%

---

### 4. Vulnerability Assessment (`k6-vulnerability-assessment.yaml`)
**Purpose**: Simulate comprehensive security vulnerability assessment

**Scenario**: Security team performing full vulnerability assessment

**Phases**:
1. **Discovery**: Catalog, categories, severities
2. **Critical Assessment**: Test all critical exploits
3. **High Assessment**: Test all high severity exploits
4. **Medium Assessment**: Test all medium severity exploits
5. **Full Suite**: Execute complete test suite

**Metrics**:
- `assessment_phase_complete`: Phases completed
- `critical_exploits_tested`: Critical exploits tested
- `high_exploits_tested`: High exploits tested
- `medium_exploits_tested`: Medium exploits tested
- `vulnerabilities_found`: Vulnerabilities discovered
- `assessment_duration_ms`: Total assessment time

**Thresholds**:
- All 5 phases complete
- HTTP failure rate < 10%

---

### 5. Random Exploit Chaos (`k6-random-exploit-chaos.yaml`)
**Purpose**: Test system resilience under unpredictable random attack patterns

**Scenario**: Unpredictable attacker launching random exploits continuously

**Load Pattern**:
- Ramp up: 0.5 → 2 req/s (30s)
- Increase: 2 → 3 req/s (1m)
- Peak: 3 → 5 req/s (1m)
- Ramp down: 5 → 0 (1m)

**Behavior**:
- Randomly selects exploits from catalog
- Randomly uses severity filters
- Chaotic timing (0.5-3s between attacks)

**Metrics**:
- `chaos_random_exploit_executed`: Random exploits executed
- `chaos_random_exploit_success`: Success rate
- `chaos_random_exploit_blocked`: Blocked rate
- `chaos_intensity`: Total chaos intensity

**Thresholds**:
- At least 10 random exploits executed
- P95 duration < 60s

---

## 🚀 Running Tests

### Run All Tests
```bash
kubectl apply -k flux/ai/agent-redteam/k8s/tests/
```

### Run Specific Test
```bash
# Smoke test
kubectl apply -f flux/ai/agent-redteam/k8s/tests/k6-smoke.yaml

# Sequential attack
kubectl apply -f flux/ai/agent-redteam/k8s/tests/k6-attack-sequential.yaml

# Parallel attack
kubectl apply -f flux/ai/agent-redteam/k8s/tests/k6-attack-parallel.yaml

# Vulnerability assessment
kubectl apply -f flux/ai/agent-redteam/k8s/tests/k6-vulnerability-assessment.yaml

# Random chaos
kubectl apply -f flux/ai/agent-redteam/k8s/tests/k6-random-exploit-chaos.yaml
```

### Check Test Status
```bash
kubectl get testruns -n agent-redteam
```

### View Test Logs
```bash
# Get pod name
kubectl get pods -n agent-redteam -l k6_cr=agent-redteam-smoke

# View logs
kubectl logs -n agent-redteam -l k6_cr=agent-redteam-smoke --tail=100
```

### Cleanup Tests
```bash
kubectl delete testruns -n agent-redteam --all
```

---

## 📊 Metrics & Monitoring

All tests export metrics to Prometheus via remote write:
- **Endpoint**: `kube-prometheus-stack-prometheus.prometheus.svc.cluster.local:9090/api/v1/write`
- **Format**: Native histograms for trends

### Key Metrics

**Attack Metrics**:
- `attack_exploit_success` / `parallel_exploit_success` / `chaos_random_exploit_success`
- `attack_exploit_blocked` / `parallel_exploit_blocked` / `chaos_random_exploit_blocked`
- `attack_exploit_duration_ms` / `parallel_exploit_duration_ms` / `chaos_exploit_duration_ms`

**Assessment Metrics**:
- `assessment_phase_complete`
- `critical_exploits_tested` / `high_exploits_tested` / `medium_exploits_tested`
- `vulnerabilities_found`

**Chaos Metrics**:
- `chaos_random_exploit_executed`
- `chaos_intensity`

---

## 🔧 Configuration

### Environment Variables

All tests use these environment variables:
- `TARGET_URL`: Agent-redteam service URL (default: `http://agent-redteam.agent-redteam.svc.cluster.local`)
- `TARGET_NAMESPACE`: Target namespace for exploits (default: `redteam-test`)
- `K6_PROMETHEUS_RW_SERVER_URL`: Prometheus remote write endpoint
- `K6_PROMETHEUS_RW_TREND_AS_NATIVE_HISTOGRAM`: Enable native histograms

### Resource Limits

| Test | Memory | CPU |
|------|--------|-----|
| Smoke | 256Mi | 200m |
| Sequential Attack | 512Mi | 500m |
| Parallel Attack | 1Gi | 1000m |
| Vulnerability Assessment | 512Mi | 500m |
| Random Chaos | 512Mi | 500m |

---

## 🎯 Use Cases

### Security Testing
- **Sequential Attack**: Test defense mechanisms against methodical attackers
- **Parallel Attack**: Test system resilience under coordinated attacks
- **Vulnerability Assessment**: Comprehensive security evaluation

### Chaos Engineering
- **Random Chaos**: Test system behavior under unpredictable attack patterns
- **Load Testing**: Validate agent-redteam performance under attack load

### CI/CD Integration
- **Smoke Test**: Quick validation in CI pipelines
- **Full Assessment**: Comprehensive testing in staging environments

---

## ⚠️ Warnings

- **Dry-Run Mode**: Tests run in dry-run mode by default (studio environment)
- **Test Namespace**: All exploits target `redteam-test` namespace
- **Resource Cleanup**: Tests don't automatically cleanup - use `/cleanup` endpoint
- **Production**: Never run these tests against production clusters!

---

## 📝 Notes

- All exploits are executed via agent-redteam API
- Tests simulate real attack patterns for security validation
- Metrics are exported to Prometheus for monitoring and alerting
- Test results include detailed summaries with success/blocked/failed rates
