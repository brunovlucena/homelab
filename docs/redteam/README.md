# 🔴 Red Team - Knative Lambda Operator Exploits

**⚠️ WARNING: These are proof-of-concept exploits for educational and authorized testing purposes only.**

This directory contains proof-of-concept exploits for vulnerabilities discovered in the `knative-lambda-operator` project. These exploits are provided to:

1. Validate the existence of vulnerabilities
2. Test remediation effectiveness
3. Train security teams on attack patterns

---

## 📋 Vulnerability Summary

| ID | Severity | Title | Exploit |
|----|----------|-------|---------|
| BLUE-001 | 🔴 CRITICAL | SSRF via Go-Git Library | `exploits/blue-001-ssrf-git/` |
| BLUE-002 | 🔴 CRITICAL | Go Template Injection | `exploits/blue-002-template-injection/` |
| VULN-001 | 🔴 CRITICAL | Command Injection (Git) | `exploits/vuln-001-cmd-injection-git/` |
| VULN-002 | 🔴 CRITICAL | Command Injection (MinIO) | `exploits/vuln-002-cmd-injection-minio/` |
| VULN-003 | 🔴 CRITICAL | Arbitrary Code Execution | `exploits/vuln-003-inline-code-exec/` |
| BLUE-005 | 🟠 HIGH | Path Traversal | `exploits/blue-005-path-traversal/` |
| BLUE-006 | 🟠 HIGH | SA Token Exposure | `exploits/blue-006-sa-token-exposure/` |
| VULN-004 | 🟠 HIGH | RBAC Privilege Escalation | `exploits/vuln-004-rbac-escalation/` |
| VULN-013 | 🟡 MEDIUM | Receiver Mode Escalation | `exploits/vuln-013-receiver-escalation/` |

---

## 🚀 Quick Start

### Prerequisites

```bash
# Ensure you have access to a test Kubernetes cluster
kubectl cluster-info

# Verify knative-lambda-operator is installed
kubectl get crd lambdafunctions.lambda.knative.io
```

### Running Exploits

```bash
# Run all exploits in sequence (test environment only!)
make run-all

# Run specific exploit
make run EXPLOIT=blue-001-ssrf-git

# Clean up after testing
make cleanup
```

---

## 🎯 Exploit Categories

### Category 1: Server-Side Request Forgery (SSRF)
- **BLUE-001**: Exploits go-git library to make requests to internal services

### Category 2: Code Injection
- **BLUE-002**: Go template injection via handler field
- **VULN-001/002**: Shell command injection via CRD fields
- **VULN-003**: Arbitrary Python/Node.js code execution

### Category 3: Privilege Escalation
- **VULN-004**: RBAC exploitation for cluster-admin
- **VULN-013**: Receiver mode SA inheritance
- **BLUE-006**: Build job SA token theft

### Category 4: Path Traversal
- **BLUE-005**: Git path traversal to read arbitrary files

---

## 🛡️ Detection Signatures

After running exploits, check for these indicators:

```bash
# Check for suspicious LambdaFunction resources
kubectl get lambdafunctions -A -o json | jq '.items[] | select(.spec.source.git.url | test("169.254|metadata|kubernetes.default"))'

# Check for failed builds with suspicious commands
kubectl logs -n knative-lambda -l lambda.knative.io/build=true --tail=100

# Check audit logs for escalation attempts
kubectl logs -n kube-system -l app=kube-apiserver | grep -E "(clusterrole|secret)"
```

---

## ⚠️ Legal Disclaimer

These exploits are provided for **authorized security testing only**. Unauthorized use against systems you do not own or have explicit permission to test is **illegal**.

By using these exploits, you agree to:
1. Only test against systems you own or have written authorization to test
2. Not use these tools for malicious purposes
3. Report any new vulnerabilities responsibly

---

## 📁 Directory Structure

```
redteam/
├── README.md                           # This file
├── Makefile                            # Automation for running exploits
├── exploits/
│   ├── blue-001-ssrf-git/             # SSRF via go-git
│   ├── blue-002-template-injection/    # Go template injection
│   ├── blue-005-path-traversal/        # Path traversal
│   ├── blue-006-sa-token-exposure/     # SA token theft
│   ├── vuln-001-cmd-injection-git/     # Git command injection
│   ├── vuln-002-cmd-injection-minio/   # MinIO command injection
│   ├── vuln-003-inline-code-exec/      # Inline code execution
│   ├── vuln-004-rbac-escalation/       # RBAC privilege escalation
│   └── vuln-013-receiver-escalation/   # Receiver mode escalation
├── payloads/
│   ├── reverse-shell.py               # Reverse shell payload
│   ├── exfiltrator.py                 # Data exfiltration payload
│   └── persistence.py                 # Persistence payload
├── scripts/
│   ├── attacker-server.py             # C2 server for receiving data
│   ├── verify-exploit.sh              # Verify exploit success
│   └── cleanup.sh                     # Cleanup after testing
└── reports/
    └── .gitkeep                       # Exploit execution reports
```

---

## 📞 Contact

For responsible disclosure of new vulnerabilities, contact:
- **Security Team:** security@example.com
