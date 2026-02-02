# 🔴 Agent-Redteam

**Security Testing Agent for Knative Lambda Operator Exploits**

⚠️ **WARNING: This agent is for AUTHORIZED TESTING ONLY.**  
Only run against test clusters you own or have explicit permission to test.

---

## 📋 Overview

Agent-Redteam is an automated security testing agent that executes proof-of-concept exploits against the Knative Lambda Operator. It integrates with the existing homelab AI agent ecosystem and provides:

- **Automated Exploit Execution**: Run individual exploits or full test suites
- **Mitigation Validation**: Verify that security controls are working
- **Metrics & Observability**: Prometheus metrics for tracking vulnerability status
- **CloudEvents Integration**: Emit events for cross-agent communication

---

## 🎯 Exploit Catalog

| ID | Severity | Category | Description |
|----|----------|----------|-------------|
| BLUE-001 | 🔴 CRITICAL | SSRF | Server-side request forgery via go-git library |
| BLUE-002 | 🔴 CRITICAL | Template Injection | Go template injection for RCE |
| VULN-001 | 🔴 CRITICAL | Command Injection | Shell injection via Git URL/ref fields |
| VULN-002 | 🔴 CRITICAL | Command Injection | Shell injection via MinIO fields |
| VULN-003 | 🔴 CRITICAL | Code Injection | Arbitrary inline code execution |
| BLUE-005 | 🟠 HIGH | Path Traversal | Read arbitrary files via git path traversal |
| BLUE-006 | 🟠 HIGH | Token Exposure | Service account token theft |
| VULN-004 | 🟠 HIGH | RBAC Escalation | Create cluster-admin via RBAC exploitation |
| VULN-013 | 🟡 MEDIUM | Receiver Escalation | SA inheritance via receiver mode |

---

## 🚀 Quick Start

### Prerequisites

```bash
# Ensure you have access to a test Kubernetes cluster
kubectl cluster-info

# Verify knative-lambda-operator is installed
kubectl get crd lambdafunctions.lambda.knative.io
```

### Local Development

```bash
# Install dependencies
make install-dev

# Run in dry-run mode (safe - no actual exploits executed)
make run-dev

# Test the catalog endpoint
curl http://localhost:8080/catalog | jq
```

### Running Exploits

```bash
# Run a single exploit (dry-run mode by default)
curl -X POST http://localhost:8080/exploit/run \
  -H "Content-Type: application/json" \
  -d '{"exploit_id": "vuln-001"}'

# Run only CRITICAL severity exploits
curl -X POST http://localhost:8080/test/run \
  -H "Content-Type: application/json" \
  -d '{"name": "critical-test", "severities": ["critical"]}'

# Run ALL exploits (use with caution!)
curl -X POST http://localhost:8080/test/run-all
```

---

## 📁 Project Structure

```
agent-redteam/
├── src/
│   ├── exploit_runner/
│   │   ├── __init__.py
│   │   ├── Dockerfile
│   │   ├── handler.py      # Main exploit runner logic
│   │   └── main.py         # FastAPI entry point
│   ├── shared/
│   │   ├── __init__.py
│   │   ├── types.py        # Type definitions
│   │   └── metrics.py      # Prometheus metrics
│   └── requirements.txt
├── k8s/
│   ├── kustomize/
│   │   ├── base/
│   │   │   ├── kustomization.yaml
│   │   │   ├── lambdaagent.yaml
│   │   │   ├── namespace.yaml
│   │   │   └── rbac.yaml
│   │   ├── studio/         # Dry-run mode (safe)
│   │   └── pro/            # Live mode (dangerous!)
│   └── tests/
│       ├── k6-smoke.yaml
│       └── kustomization.yaml
├── tests/
│   ├── unit/
│   │   └── test_exploit_runner.py
│   ├── conftest.py
│   └── requirements.txt
├── Makefile
├── README.md
└── VERSION
```

---

## 🔧 Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DRY_RUN` | `true` | When true, exploits are validated but not executed |
| `TARGET_NAMESPACE` | `redteam-test` | Namespace where exploits are deployed |
| `EXPLOITS_PATH` | `/app/exploits` | Path to exploit manifest files |
| `K8S_CONTEXT` | (none) | Kubernetes context to use |
| `K8S_TIMEOUT` | `60` | Timeout for kubectl operations |

### Deployment Modes

1. **Studio (Default)**: Dry-run mode enabled - exploits are validated but not executed
2. **Pro**: Live mode - exploits are actually executed ⚠️ USE WITH CAUTION

---

## 📊 API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/ready` | Readiness check |
| GET | `/metrics` | Prometheus metrics |
| GET | `/catalog` | List all exploits |
| GET | `/catalog/{id}` | Get exploit details |
| POST | `/exploit/run` | Run single exploit |
| POST | `/test/run` | Run test suite |
| POST | `/test/run-all` | Run ALL exploits |
| POST | `/cleanup` | Remove exploit resources |

---

## 📈 Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `agent_redteam_exploits_executed_total` | Counter | Total exploits executed |
| `agent_redteam_exploits_successful_total` | Counter | Successful exploits (vulnerabilities found) |
| `agent_redteam_exploits_blocked_total` | Counter | Blocked exploits (mitigations working) |
| `agent_redteam_vulnerabilities_found` | Gauge | Current count of exploitable vulnerabilities |
| `agent_redteam_exploit_duration_seconds` | Histogram | Time to execute exploits |

---

## 🔗 Integration

### CloudEvents

Agent-Redteam emits CloudEvents for cross-agent communication:

- `io.homelab.exploit.success`: Exploit succeeded (vulnerability found)
- `io.homelab.exploit.blocked`: Exploit blocked (mitigation working)
- `io.homelab.test.completed`: Test suite completed

### Cross-Agent Communication

Results are forwarded to `agent-bruno` for display in the homelab dashboard chatbot.

---

## ⚠️ Legal Disclaimer

These exploits are provided for **authorized security testing only**. Unauthorized use against systems you do not own or have explicit permission to test is **illegal**.

By using this agent, you agree to:
1. Only test against systems you own or have written authorization to test
2. Not use these tools for malicious purposes
3. Report any new vulnerabilities responsibly

---

## 📞 Contact

For responsible disclosure of new vulnerabilities, contact:
- **Security Team**: security@example.com
