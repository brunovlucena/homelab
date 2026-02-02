# 🚀 SRE Improvements Proposal

**Generated:** 2025-12-09  
**Based on:** SRE_LAMBDA_METRICS_REPORT.md  
**Status:** Ready for Review

---

## 📋 Executive Summary

This document proposes improvements to the knative-lambda-operator based on the SRE Lambda Metrics Report findings. The changes focus on:

1. **Fixing critical bugs** - Read-only filesystem issue blocking Git source lambdas
2. **Improving observability** - Adding missing alert rules for proactive monitoring
3. **Enhancing reliability** - Better SLO-based alerting

---

## 🔴 Critical Issues Fixed

### 1. Read-Only Filesystem Issue

**Problem:** The operator deployment uses `readOnlyRootFilesystem: true` but needs `/tmp` for Git clone operations. This caused 3 lambda functions in `redteam-test` namespace to fail.

**Affected Functions:**
- `redteam-blue001-ssrf-k8s`
- `redteam-blue001-ssrf-metadata`  
- `redteam-blue005-path-traversal`

**Error:**
```
Failed to create build context: failed to get source code: 
failed to create temp directory: mkdir /tmp/lambda-git-clone-*: read-only file system
```

**Fix Applied:** Added emptyDir volume for `/tmp` in `k8s/base/deployment.yaml`:

```yaml
volumeMounts:
  - name: tmp
    mountPath: /tmp

volumes:
  - name: tmp
    emptyDir:
      sizeLimit: 256Mi  # Limit to prevent runaway disk usage
```

**Security Note:** This maintains security hardening (`readOnlyRootFilesystem: true`) while allowing necessary temporary file operations. The 256Mi size limit prevents disk exhaustion attacks.

---

## 🟡 New Alert Rules Added

The following alerts were added to `k8s/overlays/studio/alertrules.yaml` based on report recommendations:

### Alert: KnativeLambdaFunctionHighErrorRate
- **Threshold:** Error rate > 5% for any lambda function
- **For:** 5 minutes
- **Severity:** Warning
- **Purpose:** Detect unhealthy lambda functions early

### Alert: KnativeLambdaBuildDurationHigh
- **Threshold:** P95 build duration > 120 seconds
- **For:** 5 minutes  
- **Severity:** Warning
- **Purpose:** Detect slow builds that may indicate registry or resource issues

### Alert: KnativeLambdaWorkqueueDepthHigh
- **Threshold:** Workqueue depth > 10
- **For:** 5 minutes
- **Severity:** Warning
- **Purpose:** Detect operator falling behind on reconciliations

### Alert: KnativeLambdaHighColdStartRate
- **Threshold:** Cold start rate > 20%
- **For:** 10 minutes
- **Severity:** Info
- **Purpose:** Monitor autoscaling efficiency

### Alert: KnativeLambdaFunctionHighLatency
- **Threshold:** P95 latency > 5 seconds
- **For:** 5 minutes
- **Severity:** Warning
- **Purpose:** Detect performance degradation

---

## 📊 Current Metrics Coverage

| Metric Category | Status | Notes |
|-----------------|--------|-------|
| **Operator Reconciliation** | ✅ Covered | ReconcileTotal, ReconcileDuration |
| **Build Performance** | ✅ Covered | BuildDuration, BuildJobsActive |
| **Lambda Invocations** | ✅ Covered | FunctionInvocationsTotal |
| **Lambda Errors** | ✅ Covered | FunctionErrorsTotal |
| **Cold Starts** | ✅ Covered | FunctionColdStartsTotal |
| **Workqueue** | ✅ Covered | WorkQueueDepth, WorkQueueLatency |
| **Exemplars** | ✅ Covered | Trace ID linking |

---

## 🔮 Additional Recommendations

### Short-term (This Sprint)

1. **Investigate k6-lambda-7** - Zero invocations recorded, needs verification
   ```bash
   kubectl get lambdafunction k6-lambda-7 -n knative-lambda -o yaml
   kubectl logs -l app=k6-lambda-7 -n knative-lambda
   ```

2. **Add runtime filter to Grafana dashboards** - As recommended in report
   ```yaml
   # Dashboard variable
   - name: runtime
     type: query
     query: label_values(knative_lambda_function_invocations_total, runtime)
   ```

### Medium-term (Next 2 Sprints)

1. **Add DLQ monitoring alerts**
   ```yaml
   - alert: KnativeLambdaDLQAccumulating
     expr: |
       sum by (function) (
         rabbitmq_queue_messages{queue=~".*-dlq"}
       ) > 100
     for: 15m
   ```

2. **Add SLO-based recording rules** for burn rate alerting:
   ```yaml
   # Recording rule for error budget
   - record: knative_lambda:function_error_rate:5m
     expr: |
       sum by (function, namespace) (
         rate(knative_lambda_function_invocations_total{status="error"}[5m])
       ) / sum by (function, namespace) (
         rate(knative_lambda_function_invocations_total[5m])
       )
   ```

3. **Add resource usage alerts** for lambda pods:
   ```yaml
   - alert: KnativeLambdaPodMemoryHigh
     expr: |
       container_memory_working_set_bytes{pod=~".*-lambda-.*"} 
       / container_spec_memory_limit_bytes{pod=~".*-lambda-.*"} > 0.9
     for: 5m
   ```

### Long-term (Roadmap)

1. **Implement Grafana SLO feature** for lambda functions
2. **Add tracing-based latency SLIs** via Tempo
3. **Create runbooks** for all alert rules
4. **Add chaos engineering tests** for alert validation

---

## 📈 Dashboard Improvements

### Suggested Panel Additions

| Panel | Type | PromQL |
|-------|------|--------|
| Error Budget Remaining | Stat | `1 - (sum(increase(knative_lambda_function_errors_total[30d])) / (sum(increase(knative_lambda_function_invocations_total[30d])) * 0.001))` |
| Build Queue Depth | Gauge | `knative_lambda_operator_workqueue_depth` |
| Build Duration Heatmap | Heatmap | `sum(rate(knative_lambda_operator_build_duration_seconds_bucket[5m])) by (le)` |
| Reconcile Rate | Timeseries | `sum(rate(knative_lambda_operator_reconcile_total[5m])) by (phase)` |

---

## 🧪 Testing the Changes

### Verify /tmp Volume Fix

```bash
# After deploying, verify the volume is mounted
kubectl exec -n knative-lambda deployment/knative-lambda-operator -- df -h /tmp

# Test Git clone works
kubectl logs -n knative-lambda deployment/knative-lambda-operator | grep -i "git clone"

# Check redteam lambdas reconcile successfully
kubectl get lambdafunction -n redteam-test -w
```

### Verify Alert Rules

```bash
# Check PrometheusRules are loaded
kubectl get prometheusrules -n knative-lambda

# Verify in Prometheus UI
# Navigate to: http://prometheus:9090/rules
# Search for: KnativeLambda

# Test alert expression (should return results if condition met)
curl -g 'http://prometheus:9090/api/v1/query?query=knative_lambda_operator_workqueue_depth>10'
```

---

## 📝 Files Changed

| File | Change Type | Description |
|------|-------------|-------------|
| `k8s/base/deployment.yaml` | Modified | Added emptyDir volume for /tmp |
| `k8s/overlays/studio/alertrules.yaml` | Modified | Added 5 new alert rules |
| `docs/reports/SRE_IMPROVEMENTS_PROPOSAL.md` | Created | This document |

---

## ✅ Checklist

- [x] Fix read-only filesystem issue
- [x] Add error rate alert (> 5%)
- [x] Add build duration alert (> 120s)
- [x] Add workqueue depth alert (> 10)
- [x] Add cold start rate alert (> 20%)
- [x] Add function latency alert (P95 > 5s)
- [ ] Add DLQ accumulation alert (future)
- [ ] Add SLO recording rules (future)
- [ ] Create runbooks for alerts (future)

---

**Maintainer:** SRE Team  
**Review Requested:** Platform Team  
**Next Review:** After deployment verification
