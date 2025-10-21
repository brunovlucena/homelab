# ⚠️ Runbook: Knative Operator High Memory Usage

## Alert Information

**Alert Name:** `KnativeOperatorHighMemory`  
**Severity:** Warning  
**Component:** knative-operator  
**Service:** knative-operator

## Symptom

Knative Operator is experiencing high memory usage or has been OOMKilled (Out of Memory killed).

## Impact

- **User Impact:** LOW to MEDIUM - Existing services continue running normally
- **Business Impact:** MEDIUM - Operator may become unstable or crash, preventing component updates
- **Data Impact:** LOW - No data loss, operational stability concern

## Diagnosis

### 1. Check Current Memory Usage

```bash
# Check pod memory usage
kubectl top pod -n knative-operator -l app=knative-operator

# Get memory limits and requests
kubectl get pod -n knative-operator -l app=knative-operator -o jsonpath='{.items[0].spec.containers[0].resources}'

# Check if pod was OOMKilled
kubectl get pod -n knative-operator -l app=knative-operator -o jsonpath='{.items[0].status.containerStatuses[0].lastState.terminated.reason}'
```

**Expected Normal Values:**
- Memory usage: < 80% of limit
- Current limit: 512Mi
- Current request: 128Mi

### 2. Check for OOMKills in History

```bash
# Check recent pod restarts
kubectl get pod -n knative-operator -l app=knative-operator -o jsonpath='{.items[0].status.containerStatuses[0].restartCount}'

# Check restart reasons
kubectl describe pod -n knative-operator -l app=knative-operator | grep -A 10 "Last State"

# Check events for OOMKilled
kubectl get events -n knative-operator --sort-by='.lastTimestamp' | grep -i "oom"
```

### 3. Check Operator Logs for Memory Issues

```bash
# Check logs for memory warnings
kubectl logs -n knative-operator -l app=knative-operator --tail=200 | grep -i "memory\|oom\|allocation"

# Check previous logs if pod restarted
kubectl logs -n knative-operator -l app=knative-operator --previous --tail=200 | grep -i "memory\|oom"
```

### 4. Check Number of Managed Resources

```bash
# Count KnativeServing resources
kubectl get knativeserving -A | wc -l

# Count KnativeEventing resources
kubectl get knativeeventing -A | wc -l

# Check all Knative CRDs
kubectl get crd | grep knative
```

### 5. Check Node Memory Pressure

```bash
# Get node status
kubectl describe node $(kubectl get pod -n knative-operator -l app=knative-operator -o jsonpath='{.items[0].spec.nodeName}')

# Check for memory pressure
kubectl get nodes -o custom-columns=NAME:.metadata.name,MEMORY_PRESSURE:.status.conditions[?(@.type==\"MemoryPressure\")].status
```

### 6. Analyze Memory Trends

```bash
# Get historical memory usage (if metrics-server available)
kubectl top pod -n knative-operator -l app=knative-operator --use-protocol-buffers

# Check Prometheus metrics (if available)
# Memory usage query: container_memory_usage_bytes{namespace="knative-operator", pod=~"knative-operator.*"}
```

## Resolution Steps

### Step 1: Identify Memory Leak or Spike

#### Check if Memory is Growing Over Time

```bash
# Monitor memory usage in real-time
watch -n 5 'kubectl top pod -n knative-operator -l app=knative-operator'

# Check heap profile if available
kubectl logs -n knative-operator -l app=knative-operator --tail=500 | grep -i "heap\|gc"
```

#### Common Causes:
1. **Memory Leak** - Memory grows continuously without release
2. **Resource Spike** - Temporary spike due to many reconciliations
3. **Configuration Issue** - Too many watches or large objects
4. **Insufficient Limits** - Memory limits too low for workload

### Step 2: Immediate Mitigation

#### Option A: Increase Memory Limits (Temporary Fix)

```bash
# Check current HelmRelease
kubectl get helmrelease -n knative-operator knative-operator -o yaml

# Edit HelmRelease to increase memory
kubectl edit helmrelease knative-operator -n knative-operator

# Update the memory values:
# values:
#   resources:
#     limits:
#       memory: 1Gi  # Increase from 512Mi
#     requests:
#       memory: 256Mi  # Increase from 128Mi

# Reconcile to apply changes
flux reconcile helmrelease knative-operator -n knative-operator

# Wait for rollout
kubectl rollout status deployment -n knative-operator knative-operator
```

#### Option B: Restart Operator to Clear Memory

⚠️ **Warning:** This will temporarily interrupt operator reconciliation!

```bash
# Restart deployment
kubectl rollout restart deployment -n knative-operator knative-operator

# Monitor restart
kubectl get pods -n knative-operator -l app=knative-operator -w

# Verify memory after restart
sleep 30
kubectl top pod -n knative-operator -l app=knative-operator
```

### Step 3: Long-Term Solutions

#### Solution 1: Optimize Operator Configuration

```bash
# Check for excessive watches or reconciliations
kubectl logs -n knative-operator -l app=knative-operator --tail=500 | grep -i "reconcile\|watch"

# Review KnativeServing configuration for optimization
kubectl get knativeserving -A -o yaml > /tmp/knative-serving-config.yaml

# Look for excessive replicas or resources
cat /tmp/knative-serving-config.yaml | grep -A 5 "replicas:"
```

#### Solution 2: Reduce Operator Scope

If managing too many resources:

```bash
# Check number of namespaces with Knative resources
kubectl get knativeserving -A
kubectl get knativeeventing -A

# Consider splitting operator instances if managing many clusters
```

#### Solution 3: Update Operator Version

Check if newer version has memory improvements:

```bash
# Check current version
kubectl get deployment -n knative-operator knative-operator -o jsonpath='{.spec.template.spec.containers[0].image}'

# Check available versions
helm search repo knative-operator --versions

# Update HelmRelease to newer version
kubectl edit helmrelease knative-operator -n knative-operator
# Update chart version

# Reconcile
flux reconcile helmrelease knative-operator -n knative-operator
```

#### Solution 4: Enable Memory Profiling

For debugging memory leaks:

```bash
# Add profiling flags to operator
kubectl edit deployment -n knative-operator knative-operator

# Add to container args:
# args:
#   - --pprof-addr=:6060
#   - --enable-profiling=true

# Port-forward to access pprof
kubectl port-forward -n knative-operator deployment/knative-operator 6060:6060

# Analyze memory profile (in another terminal)
go tool pprof http://localhost:6060/debug/pprof/heap
```

### Step 4: Update Configuration in Git

Update the permanent configuration:

```bash
# Edit HelmRelease values in Flux repository
# File: flux/clusters/homelab/infrastructure/knative-operator/helmrelease.yaml

# Example update:
cat <<EOF
spec:
  values:
    resources:
      requests:
        memory: "256Mi"
        cpu: "100m"
      limits:
        memory: "1Gi"
        cpu: "1000m"
EOF

# Commit and push changes
git add flux/clusters/homelab/infrastructure/knative-operator/helmrelease.yaml
git commit -m "feat: increase knative-operator memory limits"
git push

# Reconcile Flux
flux reconcile source git flux-system -n flux-system
flux reconcile helmrelease knative-operator -n knative-operator
```

## Verification

### 1. Check Memory Usage Stabilized

```bash
# Monitor memory for 5 minutes
watch -n 10 'kubectl top pod -n knative-operator -l app=knative-operator'

# Memory should be:
# - Below 80% of new limit
# - Not continuously growing
# - Stable over time
```

### 2. Check No Recent OOMKills

```bash
# Verify no restart due to OOM
kubectl get pod -n knative-operator -l app=knative-operator -o jsonpath='{.items[0].status.containerStatuses[0].restartCount}'

# Should not increase after changes
kubectl get events -n knative-operator --sort-by='.lastTimestamp' | grep -i "oom"
```

### 3. Check Operator Functionality

```bash
# Verify operator is functioning
kubectl logs -n knative-operator -l app=knative-operator --tail=50

# Check managed resources are healthy
kubectl get knativeserving -A
kubectl describe knativeserving knative-serving -n knative-serving

# Verify Ready status
kubectl get knativeserving knative-serving -n knative-serving -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}'
```

### 4. Test Reconciliation

```bash
# Trigger reconciliation
kubectl annotate knativeserving knative-serving -n knative-serving test-reconcile=$(date +%s) --overwrite

# Monitor operator logs
kubectl logs -n knative-operator -l app=knative-operator -f

# Verify memory doesn't spike excessively
kubectl top pod -n knative-operator -l app=knative-operator
```

### 5. Check Prometheus Metrics

```bash
# Query memory metrics (if Prometheus available)
# container_memory_usage_bytes{namespace="knative-operator"}
# container_memory_working_set_bytes{namespace="knative-operator"}
# rate(container_memory_usage_bytes{namespace="knative-operator"}[5m])
```

## Prevention

### 1. Set Appropriate Resource Limits

```yaml
# Recommended resource configuration
resources:
  requests:
    memory: "256Mi"  # Enough for normal operation
    cpu: "100m"
  limits:
    memory: "1Gi"    # Room for growth
    cpu: "1000m"
```

### 2. Implement Memory Monitoring

Create Prometheus alerts:

```yaml
# Alert on high memory usage
- alert: KnativeOperatorHighMemory
  expr: |
    container_memory_usage_bytes{namespace="knative-operator", container="knative-operator"} 
    / 
    container_spec_memory_limit_bytes{namespace="knative-operator", container="knative-operator"} 
    > 0.8
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Knative Operator memory usage above 80%"

# Alert on OOMKills
- alert: KnativeOperatorOOMKilled
  expr: |
    increase(kube_pod_container_status_restarts_total{namespace="knative-operator", container="knative-operator"}[10m]) > 0
    and
    kube_pod_container_status_last_terminated_reason{namespace="knative-operator", container="knative-operator", reason="OOMKilled"} == 1
  labels:
    severity: critical
  annotations:
    summary: "Knative Operator was OOMKilled"
```

### 3. Regular Memory Analysis

```bash
# Schedule regular memory checks
kubectl top pod -n knative-operator -l app=knative-operator

# Review operator logs for memory warnings
kubectl logs -n knative-operator -l app=knative-operator --tail=200 | grep -i "memory"
```

### 4. Keep Operator Updated

```bash
# Check for updates quarterly
helm search repo knative-operator --versions

# Review release notes for memory improvements
# https://github.com/knative/operator/releases
```

### 5. Optimize Managed Resources

```bash
# Regularly review managed resources
kubectl get knativeserving -A -o yaml | grep -A 10 "resources:"

# Ensure configurations are not excessive
kubectl get knativeeventing -A -o yaml | grep -A 10 "replicas:"
```

## Performance Tips

1. **Memory Headroom**: Always set limits 2-3x higher than typical usage
2. **Gradual Scaling**: Increase memory in increments (512Mi → 1Gi → 2Gi)
3. **Monitor Trends**: Use Prometheus to track memory growth over time
4. **Resource Cleanup**: Ensure unused Knative resources are deleted
5. **Version Updates**: Keep operator up-to-date for performance improvements

## Common Memory Usage Patterns

### Normal Usage
```
Startup: 100-150Mi
Idle: 150-200Mi
Active Reconciliation: 200-300Mi
Peak: 300-400Mi
```

### Concerning Patterns
```
Continuous Growth: Memory increases without plateau
Spike Recovery: Memory doesn't decrease after reconciliation
Sawtooth Pattern: Rapid increases followed by OOMKills
```

## Related Alerts

- `KnativeOperatorDown` - May trigger if OOMKilled
- `KnativeOperatorCrashLoop` - May occur due to repeated OOMKills
- `KnativeOperatorReconciliationFailed` - Can be caused by memory issues
- `NodeMemoryPressure` - May affect operator scheduling

## Escalation

If memory issues persist after optimization:

1. ✅ Verify all resolution steps completed
2. 📊 Collect memory profiles and metrics
3. 🔍 Analyze operator logs for patterns
4. 📈 Review historical memory trends
5. 🔄 Consider operator version upgrade
6. 📞 Contact platform team with diagnostics
7. 🐛 File bug report with Knative project if suspected memory leak

**Escalation Criteria:**
- Memory usage exceeds limits despite increases
- Repeated OOMKills (> 3 in 1 hour)
- Memory grows continuously without stabilizing
- Impact on managed Knative services

## Additional Resources

- [Knative Operator Performance](https://knative.dev/docs/install/operator/knative-with-operators/)
- [Kubernetes Memory Management](https://kubernetes.io/docs/tasks/configure-pod-container/assign-memory-resource/)
- [Memory Profiling with pprof](https://github.com/google/pprof/blob/master/doc/README.md)
- [Go Memory Management](https://go.dev/doc/diagnostics#profiling)

## Quick Commands Reference

```bash
# Check memory usage
kubectl top pod -n knative-operator -l app=knative-operator

# Check for OOMKills
kubectl get pod -n knative-operator -l app=knative-operator -o jsonpath='{.items[0].status.containerStatuses[0].lastState.terminated.reason}'

# View memory limits
kubectl get pod -n knative-operator -l app=knative-operator -o jsonpath='{.items[0].spec.containers[0].resources.limits.memory}'

# Increase memory limits
kubectl edit helmrelease knative-operator -n knative-operator

# Restart operator
kubectl rollout restart deployment -n knative-operator knative-operator

# Monitor memory
watch -n 5 'kubectl top pod -n knative-operator -l app=knative-operator'
```

---

**Last Updated:** 2025-10-15  
**Version:** 1.0  
**Owner:** Platform Team

