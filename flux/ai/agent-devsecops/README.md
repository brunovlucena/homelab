# 🔐 Agent DevSecOps

Security-focused AI agent for vulnerability scanning, policy enforcement, and secure rollout coordination.

## Overview

Agent DevSecOps provides automated security operations with three RBAC levels:

| Level | Service Account | Capabilities | Use Case |
|-------|-----------------|--------------|----------|
| **readonly** | `devsecops-readonly-sa` | View all resources, cannot modify | Default scanning/auditing |
| **operator** | `devsecops-operator-sa` | Create proposals, limited writes | Automated security ops |
| **admin** | `devsecops-admin-sa` | Full cluster access | Emergency response only |

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                      AGENT DEVSECOPS                                 │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐             │
│  │   SCANNER   │───▶│  PROPOSALS  │───▶│   HUMAN     │             │
│  │             │    │   (CRD)     │    │  APPROVAL   │             │
│  │ • CVE scan  │    │             │    │             │             │
│  │ • Compliance│    │ • Patch     │    │ • Review    │             │
│  │ • Policies  │    │ • Suspend   │    │ • Approve   │             │
│  │ • Secrets   │    │ • Rotate    │    │ • Reject    │             │
│  └─────────────┘    └─────────────┘    └─────────────┘             │
│                                                                     │
│  RBAC LEVELS:                                                       │
│  ┌────────────────────────────────────────────────────────────┐    │
│  │ readonly ──▶ operator ──▶ admin (emergency only)           │    │
│  │    ↑            ↑            ↑                             │    │
│  │  default   automated    requires human confirmation        │    │
│  └────────────────────────────────────────────────────────────┘    │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

## Quick Start

```bash
# Build and deploy
make build
make deploy-pro

# Check status
make status

# View RBAC level
make rbac-status

# List security proposals
make proposals
```

## RBAC Levels

### 1. Readonly (Default)

Safest mode - can only observe and report.

```bash
make rbac-readonly
```

Capabilities:
- ✅ View all resources across cluster
- ✅ Generate vulnerability reports
- ✅ Log security findings
- ❌ Cannot modify any resources
- ❌ Cannot create proposals

### 2. Operator (Recommended for Automation)

Limited write access for automated security operations.

```bash
make rbac-operator
```

Capabilities:
- ✅ Everything in readonly
- ✅ Create SecurityProposal CRDs
- ✅ Create/update NetworkPolicies
- ✅ Rotate secrets
- ✅ Suspend Flux reconciliation (emergency)
- ❌ Cannot delete core resources
- ❌ Cannot escalate privileges

### 3. Admin (Emergency Only)

Full cluster access - use with extreme caution.

```bash
make rbac-admin  # Requires confirmation
```

Capabilities:
- ⚠️ Full cluster access
- ⚠️ Should only be used during incidents
- ⚠️ All actions are audit logged
- ⚠️ Requires manual confirmation to activate

## Security Proposals

Instead of directly modifying resources, the agent creates `SecurityProposal` CRDs that require human approval:

```yaml
apiVersion: devsecops.homelab.io/v1alpha1
kind: SecurityProposal
metadata:
  name: proposal-20241210123456
spec:
  action: suspend
  targetResource:
    kind: HelmRelease
    name: vulnerable-app
    namespace: production
  severity: critical
  reason: "CVE-2024-1234 found in container image"
  cveIds:
    - CVE-2024-1234
  autoApprove: false
```

### Managing Proposals

```bash
# List all proposals
make proposals

# List pending proposals
make proposals-pending

# Approve a proposal
make proposals-approve PROPOSAL=proposal-20241210123456

# Reject via kubectl
kubectl patch securityproposal proposal-20241210123456 \
  -n agent-devsecops \
  --type=merge -p '{"status":{"phase":"Rejected"}}'
```

## Events

### Events Emitted

| Event Type | Description |
|------------|-------------|
| `security.scan.started` | Vulnerability scan initiated |
| `security.scan.completed` | Scan finished with results |
| `security.vuln.found` | Vulnerability discovered |
| `security.proposal.created` | New security proposal |
| `security.alert.raised` | Security alert generated |

### Events Consumed

| Event Type | Description |
|------------|-------------|
| `io.homelab.build.completed` | Triggers image scan |
| `io.homelab.deploy.requested` | Security gate check |
| `io.homelab.alert.security` | Security alert handling |
| `io.homelab.schedule.daily` | Daily compliance check |
| `io.homelab.schedule.weekly` | Weekly vulnerability scan |

## CRDs

| CRD | Purpose |
|-----|---------|
| `SecurityProposal` | Proposed changes requiring approval |
| `VulnerabilityReport` | CVE scan results |
| `ComplianceCheck` | Policy compliance status |
| `SecretRotation` | Automated secret rotation tracking |

## Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `RBAC_LEVEL` | `readonly` | Active RBAC level |
| `APPROVAL_REQUIRED` | `true` | Require human approval |
| `SCANNER_MODE` | `continuous` | Scan mode |
| `LOG_LEVEL` | `info` | Logging verbosity |

## Development

```bash
# Run locally
cd src/scanner
python main.py

# Run tests
make test

# Build image
make build-local

# Push to GHCR
make build-push
```

## Security Considerations

1. **Principle of Least Privilege**: Always use the lowest RBAC level needed
2. **Audit Logging**: All actions are logged for forensic analysis
3. **Human in the Loop**: Critical changes require explicit approval
4. **No Auto-Delete**: Agent cannot delete core resources even in admin mode
5. **Secret Handling**: Secrets are only rotated, never exposed in logs

## Roadmap

- [ ] Integration with Falco for runtime security
- [ ] SBOM generation and tracking
- [ ] Integration with external vulnerability databases
- [ ] Policy-as-code with OPA/Gatekeeper
- [ ] Automated incident response playbooks
