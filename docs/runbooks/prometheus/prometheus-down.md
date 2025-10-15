# 🚨 Runbook: Prometheus Down

## Alert Information

**Alert Name:** `PrometheusDown`  
**Severity:** Critical  
**Component:** prometheus  
**Service:** prometheus-server

## Symptom

Prometheus server has been down for more than 1 minute. All metrics collection and alerting is unavailable.

## Impact

- **User Impact:** HIGH - No metrics dashboards, no alerting
- **Business Impact:** CRITICAL - Complete observability blind spot
- **Data Impact:** MEDIUM - Metrics gap during downtime (cannot be recovered)

## Diagnosis

### 1. Check Pod Status

```bash
kubectl get pods -n prometheus -l app.kubernetes.io/name=prometheus
kubectl describe pod -n prometheus -l app.kubernetes.io/name=prometheus
```

### 2. Check Pod Logs

```bash
kubectl logs -n prometheus -l app.kubernetes.io/name=prometheus --tail=100
kubectl logs -n prometheus -l app.kubernetes.io/name=prometheus --previous  # If pod restarted
```

### 3. Check Service and Endpoints

```bash
kubectl get svc -n prometheus prometheus-kube-prometheus-prometheus
kubectl get endpoints -n prometheus prometheus-kube-prometheus-prometheus
```

### 4. Check Events

```bash
kubectl get events -n prometheus --sort-by='.lastTimestamp' | head -20
```

### 5. Check Resource Limits

```bash
kubectl top pods -n prometheus -l app.kubernetes.io/name=prometheus
```

### 6. Check StatefulSet Status

```bash
kubectl get statefulset -n prometheus
kubectl describe statefulset -n prometheus prometheus-prometheus-kube-prometheus-prometheus
```

### 7. Check PVC Status

```bash
kubectl get pvc -n prometheus
kubectl describe pvc -n prometheus
```

## Resolution Steps

### Step 1: Check if pods are running

```bash
POD_STATUS=$(kubectl get pods -n prometheus -l app.kubernetes.io/name=prometheus -o jsonpath='{.items[0].status.phase}')
echo "Pod Status: $POD_STATUS"
```

### Step 2: If pods are not running, check why

```bash
# Check for ImagePullBackOff
kubectl describe pod -n prometheus -l app.kubernetes.io/name=prometheus | grep -A 10 "Events:"

# Check for CrashLoopBackOff
kubectl logs -n prometheus -l app.kubernetes.io/name=prometheus --previous
```

### Step 3: Common Issues and Fixes

#### Issue: ImagePullBackOff
**Cause:** Cannot pull container image  
**Fix:**
```bash
# Verify image exists
kubectl get statefulset -n prometheus prometheus-prometheus-kube-prometheus-prometheus \
  -o jsonpath='{.spec.template.spec.containers[0].image}'

# Check image pull secrets
kubectl get secrets -n prometheus
```

#### Issue: CrashLoopBackOff
**Cause:** Application failing to start  
**Fix:**
```bash
# Check application logs for startup errors
kubectl logs -n prometheus -l app.kubernetes.io/name=prometheus --tail=100

# Common causes:
# - TSDB corruption
# - Storage issues
# - Configuration errors
# - Out of disk space
# - WAL replay failures
```

#### Issue: TSDB Corruption
**Cause:** Corrupted time-series database  
**Fix:**
```bash
# Check TSDB status in logs
kubectl logs -n prometheus -l app.kubernetes.io/name=prometheus | grep -i "tsdb\|wal\|corruption"

# If corrupted, you may need to delete and recreate (DATA LOSS)
# kubectl delete pod -n prometheus -l app.kubernetes.io/name=prometheus
```

#### Issue: OOMKilled
**Cause:** Out of memory  
**Fix:**
```bash
# Check memory usage patterns
kubectl top pods -n prometheus -l app.kubernetes.io/name=prometheus

# Increase memory limits via Helm values
# resources:
#   limits:
#     memory: 4Gi  # Increase from current
#   requests:
#     memory: 2Gi
```

#### Issue: Storage Full
**Cause:** PVC out of space  
**Fix:**
```bash
# Check PVC usage
kubectl exec -n prometheus prometheus-prometheus-kube-prometheus-prometheus-0 -- df -h /prometheus

# If full, expand PVC (if storage class allows)
kubectl edit pvc -n prometheus prometheus-prometheus-kube-prometheus-prometheus-db-prometheus-prometheus-kube-prometheus-prometheus-0

# Or reduce retention period in Helm values:
# retention: 7d  # Reduce from current
```

#### Issue: WAL Replay Timeout
**Cause:** Taking too long to replay Write-Ahead Log  
**Fix:**
```bash
# Check WAL status
kubectl logs -n prometheus -l app.kubernetes.io/name=prometheus | grep -i "wal"

# Increase startup timeout if needed
kubectl edit statefulset -n prometheus prometheus-prometheus-kube-prometheus-prometheus
# Update: livenessProbe.initialDelaySeconds to higher value
```

### Step 4: Restart the StatefulSet if needed

```bash
kubectl rollout restart statefulset/prometheus-prometheus-kube-prometheus-prometheus -n prometheus
kubectl rollout status statefulset/prometheus-prometheus-kube-prometheus-prometheus -n prometheus
```

### Step 5: Force reconcile Flux/Helm if needed

```bash
flux reconcile helmrelease kube-prometheus-stack -n prometheus
```

## Verification

1. Check that pods are running:
```bash
kubectl get pods -n prometheus -l app.kubernetes.io/name=prometheus
```

2. Verify Prometheus UI is accessible:
```bash
kubectl port-forward -n prometheus svc/prometheus-kube-prometheus-prometheus 9090:9090
# Open http://localhost:9090
```

3. Test basic queries:
```bash
# Check if metrics are being scraped
# Query: up
# Should show targets with value 1
```

4. Verify TSDB health:
```bash
kubectl exec -n prometheus prometheus-prometheus-kube-prometheus-prometheus-0 -- \
  promtool check config /etc/prometheus/config_out/prometheus.env.yaml
```

5. Check scrape targets:
```bash
# Open Prometheus UI -> Status -> Targets
# Verify all targets are UP
```

## Prevention

1. Set appropriate resource requests and limits
2. Monitor disk usage and set up alerts
3. Configure retention period appropriately (balance storage vs history)
4. Regular TSDB health checks
5. Implement PVC auto-expansion if supported
6. Monitor WAL size and replay times
7. Use remote write for long-term storage
8. Regular backups of Prometheus data

## Related Alerts

- `PrometheusTargetDown`
- `PrometheusHighMemoryUsage`
- `PrometheusStorageFull`
- `PrometheusTSDBCompactionsFailing`
- `PrometheusRuleEvaluationFailures`
- `PrometheusAlertmanagerDown`

## Escalation

If the issue persists after following these steps:
1. Check for cluster-wide issues (node failures, network issues)
2. Review recent Helm chart updates
3. Check Prometheus configuration for errors
4. Review scrape configs for problematic targets
5. Contact on-call engineer

## Additional Resources

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Prometheus Troubleshooting Guide](https://prometheus.io/docs/prometheus/latest/troubleshooting/)
- [TSDB Documentation](https://prometheus.io/docs/prometheus/latest/storage/)
- [Kube-Prometheus-Stack Docs](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)

---

**Last Updated:** 2025-10-15  
**Version:** 1.0

