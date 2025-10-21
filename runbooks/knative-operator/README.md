# 📚 Knative Operator Runbooks

Comprehensive operational runbooks for troubleshooting and resolving Knative Operator issues in the homelab Kubernetes cluster.

## Overview

Knative Operator manages the lifecycle of Knative components (Serving, Eventing) in the cluster:
- **Version**: 1.16.3
- **Namespace**: knative-operator
- **Resource Allocation**:
  - Requests: 100m CPU, 128Mi memory
  - Limits: 500m CPU, 512Mi memory
- **Purpose**: Manages installation and upgrades of Knative components

## Quick Reference

| Alert | Severity | Impact | Runbook |
|-------|----------|--------|---------|
| KnativeOperatorDown | Critical | Cannot manage Knative components | [operator-down.md](./operator-down.md) |
| KnativeOperatorHighMemory | Warning | Memory pressure/OOMKills | [operator-high-memory.md](./operator-high-memory.md) |
| KnativeOperatorReconciliationFailed | Warning | Components not updating | [operator-reconciliation-failed.md](./operator-reconciliation-failed.md) |
| KnativeOperatorCrashLoop | Critical | Operator continuously crashing | [operator-crash-loop.md](./operator-crash-loop.md) |

## Runbooks

### 🚨 Critical Issues

#### [Knative Operator Down](./operator-down.md)
Complete operator outage - cannot manage Knative components.

**Quick Check:**
```bash
kubectl get pods -n knative-operator
```

**Quick Fix:**
```bash
# Restart operator
kubectl rollout restart deployment -n knative-operator knative-operator
```

---

#### [Operator Crash Loop](./operator-crash-loop.md)
Operator pod continuously crashing and restarting.

**Quick Check:**
```bash
kubectl get pods -n knative-operator
kubectl logs -n knative-operator -l app=knative-operator --tail=50
```

**Quick Fix:**
```bash
# Check for configuration issues
kubectl describe deployment -n knative-operator knative-operator
kubectl logs -n knative-operator -l app=knative-operator --previous
```

---

### ⚠️ Warning Issues

#### [High Memory Usage](./operator-high-memory.md)
Operator experiencing memory pressure or OOMKills.

**Quick Check:**
```bash
kubectl top pods -n knative-operator
kubectl get pods -n knative-operator -o json | jq -r '.items[] | select(.status.containerStatuses[].lastState.terminated.reason == "OOMKilled") | .metadata.name'
```

**Quick Fix:**
```bash
# Increase memory limits
kubectl edit helmrelease -n knative-operator knative-operator
# Update memory limits
```

---

#### [Reconciliation Failed](./operator-reconciliation-failed.md)
Operator failing to reconcile Knative components.

**Quick Check:**
```bash
kubectl get knativeserving -A
kubectl get knativeeventing -A
kubectl describe knativeserving knative-serving -n knative-serving
```

**Quick Fix:**
```bash
# Force reconciliation
kubectl annotate knativeserving knative-serving -n knative-serving reconcile=$(date +%s) --overwrite
```

---

## Common Troubleshooting Commands

### Check Overall Health
```bash
# All pods status
kubectl get pods -n knative-operator

# Resource usage
kubectl top pods -n knative-operator

# Recent events
kubectl get events -n knative-operator --sort-by='.lastTimestamp' | head -20
```

### Check Managed Components
```bash
# Check KnativeServing resources
kubectl get knativeserving -A

# Check KnativeEventing resources
kubectl get knativeeventing -A

# Describe Knative Serving
kubectl describe knativeserving knative-serving -n knative-serving
```

### Check Operator Logs
```bash
# View operator logs
kubectl logs -n knative-operator -l app=knative-operator --tail=100

# Follow logs
kubectl logs -n knative-operator -l app=knative-operator -f

# Check previous logs if crashed
kubectl logs -n knative-operator -l app=knative-operator --previous
```

### Check Operator Configuration
```bash
# Check HelmRelease
kubectl get helmrelease -n knative-operator knative-operator -o yaml

# Check deployment
kubectl get deployment -n knative-operator -o yaml

# Check operator version
kubectl get deployment -n knative-operator knative-operator -o jsonpath='{.spec.template.spec.containers[0].image}'
```

### Force Reconciliation
```bash
# Reconcile HelmRelease
flux reconcile helmrelease knative-operator -n knative-operator

# Force operator restart
kubectl rollout restart deployment -n knative-operator knative-operator
```

## Architecture

```
┌─────────────────────┐
│  Knative Operator   │
│   (Deployment)      │
└──────────┬──────────┘
           │
           │ Manages
           ▼
┌─────────────────────┐
│  KnativeServing CR  │
│ (knative-serving ns)│
└──────────┬──────────┘
           │
           │ Creates/Manages
           ▼
┌─────────────────────┐
│  Knative Serving    │
│  - Activator        │
│  - Autoscaler       │
│  - Controller       │
│  - Webhook          │
│  - Kourier Ingress  │
└─────────────────────┘
```

## Configuration

**Location**: `flux/clusters/homelab/infrastructure/knative-operator/helmrelease.yaml`

**Key Settings**:
- Chart: knative-operator
- Version: 1.16.3
- Resources:
  - CPU: 100m (request) → 500m (limit)
  - Memory: 128Mi (request) → 512Mi (limit)

## CRD Management

The Knative Operator manages these Custom Resource Definitions:

### KnativeServing
Manages Knative Serving installation:
```bash
kubectl get knativeserving -A
kubectl describe knativeserving knative-serving -n knative-serving
```

### KnativeEventing
Manages Knative Eventing installation:
```bash
kubectl get knativeeventing -A
kubectl describe knativeeventing knative-eventing -n knative-eventing
```

## Escalation Matrix

| Issue | First Response | Escalation Time | Escalate To |
|-------|---------------|-----------------|-------------|
| Operator down | Restart deployment | 15 minutes | Platform team |
| Crash loop | Check logs, config | 15 minutes | Platform team |
| High memory | Increase limits | 30 minutes | Capacity planning |
| Reconciliation failure | Check CRD status | 30 minutes | Platform team |

## Related Documentation

- [Knative Operator Configuration](../../../flux/clusters/homelab/infrastructure/knative-operator/helmrelease.yaml)
- [Knative Serving Runbooks](../knative-serving/README.md)
- [Architecture Overview](../../../ARCHITECTURE.md)
- [Knative Official Docs](https://knative.dev/docs/install/operator/knative-with-operators/)

## Monitoring & Alerts

Knative Operator metrics should be scraped by Prometheus. Key metrics to monitor:
- Pod restarts
- Memory usage
- CPU usage
- Reconciliation status

## Support

For issues not covered by these runbooks:
1. Check operator logs: `kubectl logs -n knative-operator -l app=knative-operator`
2. Review recent changes: `flux get helmreleases -n knative-operator`
3. Consult [Knative documentation](https://knative.dev/docs/)
4. Check [GitHub issues](https://github.com/knative/operator/issues)

---

**Last Updated**: 2025-10-15  
**Knative Operator Version**: 1.16.3  
**Maintainer**: Homelab Platform Team

