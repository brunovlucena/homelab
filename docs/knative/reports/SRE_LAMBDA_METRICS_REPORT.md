# 🔍 SRE Lambda Metrics Report

**Generated:** 2025-12-09 17:33 UTC  
**Operator Version:** v1.5.10  
**Report Type:** Lambda Functions Health & Performance Analysis

---

## 📊 Executive Summary

| Metric | Value | Status |
|--------|-------|--------|
| **Total Lambda Functions** | 10 (Ready) | ✅ Healthy |
| **Operator Workqueue Depth** | 0 | ✅ Healthy |
| **Total Operator Errors** | 35 | ⚠️ Monitor |
| **P95 Build Duration** | 54.7s | ✅ Normal |
| **P95 Reconcile Duration** | 45.8ms | ✅ Excellent |

---

## 🏃 Lambda Functions Status

### knative-lambda Namespace

| Function | Invocations (Success) | Invocations (Error) | Error Rate | Cold Starts | Avg Duration |
|----------|----------------------|---------------------|------------|-------------|--------------|
| **k6-lambda-1** (Node.js) | 300 | 0 | 0% | 2 | 0.41ms |
| **k6-lambda-3** (Node.js) | 240 | 0 | 0% | 2 | 0.33ms |
| **k6-lambda-5** (Node.js) | 240 | 0 | 0% | 2 | 0.18ms |
| **k6-lambda-7** (Node.js) | 0 | 0 | N/A | 0 | N/A |
| **notifi-parser-0197ad6c** | 0 | 14,280 | 100% | 2 | 1.61ms |

### Throughput (Current - 5m Rate)

| Function | Requests/sec |
|----------|-------------|
| k6-lambda-1 | 2.53 req/s |
| k6-lambda-5 | 1.89 req/s |
| k6-lambda-3 | 1.67 req/s |
| k6-lambda-7 | 0.93 req/s |
| notifi-parser-0197ad6c | 0 req/s |

### P95 Response Times

| Function | P95 Duration |
|----------|-------------|
| k6-lambda-1 | 4.98ms |
| notifi-parser-0197ad6c | 4.91ms |
| k6-lambda-3 | 4.79ms |
| k6-lambda-7 | 4.77ms |
| k6-lambda-5 | 4.76ms |

---

## 🔧 Operator Metrics

### Reconciliation Summary

| Metric | Value |
|--------|-------|
| **Total Reconciliations** | 123 |
| **Ready Phase** | 46 |
| **Building Phase** | 8 |
| **Deploying Phase** | 27 |
| **Other** | 35 |
| **P95 Reconcile Duration** | 45.8ms |

### Lambda Functions by Namespace

| Namespace | Ready | Building | Deploying | Pending |
|-----------|-------|----------|-----------|---------|
| knative-lambda | 10 | 0 | 0 | 0 |
| redteam-test | 2 | 1 | 2 | 3 |

### Build Performance

| Metric | Value |
|--------|-------|
| **P95 Build Duration** | 54.7 seconds |
| **Total Operator Errors** | 35 |

---

## 📝 Log Analysis

### Recent Operator Activity

```
✅ Reconciling eventing infrastructure - contract-fetcher (agent-contracts)
✅ Reconciling Deploying phase - contract-fetcher
✅ Kubernetes security context warnings (non-critical)
```

### Lambda Function Logs

```
✅ k6-lambda-2: All POST / HTTP/1.1 returning 200 OK
✅ No 500 errors detected in Python lambdas
✅ Health endpoints responding correctly
```

### ⚠️ Errors Detected

| Error Type | Count | Impact | Namespace |
|------------|-------|--------|-----------|
| Read-only filesystem (Git clone) | 3 | **Medium** | redteam-test |

**Error Details:**
```
Failed to create build context: failed to get source code: 
failed to create temp directory: mkdir /tmp/lambda-git-clone-*: read-only file system
```

**Affected Functions:**
- redteam-blue001-ssrf-k8s
- redteam-blue001-ssrf-metadata
- redteam-blue005-path-traversal

**Root Cause:** Operator running with read-only root filesystem security context, preventing Git clone operations to /tmp.

**Recommendation:** Add writable emptyDir volume mounted at /tmp in operator deployment.

---

## 📈 Grafana Dashboards

### Available Dashboards

| Dashboard | URL | Tags |
|-----------|-----|------|
| **Knative Lambda Metrics** | `/d/knative-lambda-metrics` | knative, lambda, serverless, functions |
| **Knative Lambda Operator** | `/d/knative-lambda-operator` | knative, lambda, operator |

### Dashboard Panels (Knative Lambda Metrics)

| Panel | Type | Description |
|-------|------|-------------|
| Total Invocations | stat | Overall function invocations |
| Total Errors | stat | Cumulative error count |
| Total Cold Starts | stat | Cold start counter |
| Error Rate | stat | Current error percentage |
| Invocation Rate by Function | timeseries | Rate of invocations |
| Error Rate by Function | timeseries | Errors over time |
| Function Duration (P50/P95/P99) | timeseries | Latency percentiles |
| Cold Starts Rate | timeseries | Cold start frequency |
| CPU Usage by Pod | timeseries | Resource consumption |
| Memory Usage by Pod | timeseries | Memory utilization |
| Running Pods by Deployment | timeseries | Autoscaling metrics |

---

## 🎯 Key Findings

### ✅ Healthy

1. **Lambda Functions**: All 10 functions in knative-lambda are Ready
2. **Operator Queue**: Workqueue depth is 0 (no backlog)
3. **Reconciliation**: P95 at 45.8ms (excellent performance)
4. **Python Runtime**: Fix for CloudEvent response working - no more 500 errors
5. **Node.js Lambdas**: 0% error rate, healthy invocations

### ⚠️ Warnings

1. **notifi-parser-0197ad6c**: 14,280 errors recorded (100% error rate)
   - This is expected behavior for DLQ testing
   
2. **Build Errors in redteam-test**: 3 functions failing due to read-only filesystem
   - Action Required: Update operator deployment to add writable /tmp volume

3. **k6-lambda-7**: No invocations recorded - may need investigation

### 📊 Performance Metrics

| Category | Status | Notes |
|----------|--------|-------|
| Invocation Latency | ✅ Excellent | P95 < 5ms for all functions |
| Build Duration | ✅ Good | P95 at 54.7s |
| Reconcile Speed | ✅ Excellent | P95 at 45.8ms |
| Error Rate | ⚠️ Mixed | 0% for k6 lambdas, 100% for notifi (expected) |

---

## 🔗 Quick Links

- **Grafana Lambda Dashboard**: [Knative Lambda Metrics](http://grafana.homelab/d/knative-lambda-metrics)
- **Grafana Operator Dashboard**: [Knative Lambda Operator](http://grafana.homelab/d/knative-lambda-operator)
- **Prometheus Datasource**: `prometheus`
- **Loki Datasource**: `loki`

---

## 📋 Recommendations

### Immediate Actions

1. **Fix read-only filesystem issue** for Git source lambdas:
   ```yaml
   # Add to operator deployment
   volumes:
   - name: tmp
     emptyDir: {}
   volumeMounts:
   - name: tmp
     mountPath: /tmp
   ```

2. **Investigate k6-lambda-7**: Check if function is being invoked correctly

### Monitoring Improvements

1. Set up alerts for:
   - Error rate > 5% for any lambda
   - Build duration > 120s
   - Workqueue depth > 10

2. Add dashboard variable for filtering by runtime (Python/Node.js/Go)

---

*Report generated by SRE Automation using Grafana MCP*
