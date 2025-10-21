# 🚨 Runbook: Loki High Memory Usage

## Alert Information

**Alert Name:** `LokiHighMemory`  
**Severity:** Warning  
**Component:** Loki  
**Service:** All Components

## Symptom

Loki pods are using high memory, approaching or hitting configured limits. May be causing OOMKills.

## Impact

- **User Impact:** MODERATE - Performance degradation or service interruption
- **Business Impact:** MODERATE - Reduced query performance and potential log loss
- **Data Impact:** POTENTIAL - Risk during OOMKill events

## Diagnosis

### 1. Check Memory Usage by Component

```bash
# Check all Loki pods memory usage
kubectl top pods -n loki

# Check specifically by component
kubectl top pods -n loki -l app.kubernetes.io/component=backend
kubectl top pods -n loki -l app.kubernetes.io/component=read
kubectl top pods -n loki -l app.kubernetes.io/component=write
```

### 2. Check for OOMKilled Pods

```bash
# Check for recent OOM kills
kubectl get pods -n loki -o json | jq -r '.items[] | select(.status.containerStatuses[].lastState.terminated.reason == "OOMKilled") | .metadata.name'

# Check pod events for OOM
kubectl get events -n loki --sort-by='.lastTimestamp' | grep -i "oom"

# Describe pods showing OOM issues
kubectl describe pod -n loki <pod-name> | grep -A 10 "Last State"
```

### 3. Check Memory Metrics from Loki

```bash
# Port forward to Loki
kubectl port-forward -n loki svc/loki-gateway 3100:80

# Check memory metrics
curl http://localhost:3100/metrics | grep "process_resident_memory_bytes\|go_memstats_alloc_bytes"
```

### 4. Check Current Memory Limits

```bash
# Check configured limits
kubectl get pods -n loki -o json | jq -r '.items[] | "\(.metadata.name): \(.spec.containers[].resources.limits.memory)"'

# Check HelmRelease configuration
kubectl get helmrelease -n loki loki -o yaml | grep -A 20 "resources:"
```

### 5. Analyze Memory Growth Pattern

```bash
# Check logs for memory-related warnings
kubectl logs -n loki -l app.kubernetes.io/name=loki --tail=500 | grep -i "memory\|oom\|gc"
```

## Resolution Steps

### Step 1: Identify which component has high memory

```bash
# Get memory usage sorted
kubectl top pods -n loki --sort-by=memory
```

### Step 2: Common Issues and Fixes

#### Issue: Read pods high memory (query load)
**Cause:** Large queries or many concurrent queries  
**Fix:**
```bash
# Check query activity
kubectl logs -n loki -l app.kubernetes.io/component=read --tail=100 | grep -i "query"

# Solution 1: Increase memory limits
kubectl edit helmrelease -n loki loki
# Update:
# read:
#   resources:
#     limits:
#       memory: 2Gi  # Increase from default
#     requests:
#       memory: 1Gi

# Solution 2: Scale read replicas horizontally
kubectl scale deployment -n loki loki-read --replicas=3

# Solution 3: Implement query limits
kubectl edit helmrelease -n loki loki
# Add:
# loki:
#   limits_config:
#     max_query_length: 721h  # 30 days max
#     max_query_lookback: 744h
#     max_chunks_per_query: 2000000  # Reduce if needed
#     max_entries_limit_per_query: 5000  # Limit results
```

#### Issue: Write pods high memory (ingestion load)
**Cause:** High ingestion rate or large batches  
**Fix:**
```bash
# Check ingestion rate
curl http://localhost:3100/metrics | grep loki_distributor_bytes_received_total

# Solution 1: Increase memory limits
kubectl edit helmrelease -n loki loki
# Update:
# write:
#   resources:
#     limits:
#       memory: 2Gi
#     requests:
#       memory: 1Gi

# Solution 2: Scale write replicas
kubectl scale deployment -n loki loki-write --replicas=3

# Solution 3: Reduce batch size at source (Alloy)
kubectl edit configmap -n alloy alloy-config
# Reduce batch_size and batch_wait parameters
```

#### Issue: Backend pods high memory
**Cause:** Compaction or index operations  
**Fix:**
```bash
# Check backend logs for compaction activity
kubectl logs -n loki -l app.kubernetes.io/component=backend --tail=100 | grep -i "compact"

# Increase backend memory
kubectl edit helmrelease -n loki loki
# Update:
# backend:
#   resources:
#     limits:
#       memory: 3Gi
#     requests:
#       memory: 2Gi

# Adjust compaction settings
kubectl edit helmrelease -n loki loki
# Add:
# loki:
#   compactor:
#     retention_enabled: true
#     compaction_interval: 10m
#     working_directory: /tmp/loki-compactor
```

#### Issue: High cardinality causing memory issues
**Cause:** Too many unique label combinations  
**Fix:**
```bash
# Check stream count
curl http://localhost:3100/metrics | grep loki_ingester_streams

# Check series metrics
curl http://localhost:3100/loki/api/v1/label

# Solution: Reduce label cardinality at source
kubectl edit configmap -n alloy alloy-config
# Remove high-cardinality labels:
# - pod_name
# - container_id  
# - request_id
# Keep only: namespace, app, component, environment

# Set stream limit
kubectl edit helmrelease -n loki loki
# Add:
# loki:
#   limits_config:
#     max_streams_per_user: 10000  # Set reasonable limit
```

#### Issue: Memory leak in Loki
**Cause:** Known bug or memory not being released  
**Fix:**
```bash
# Temporary: Restart affected pods
kubectl rollout restart deployment -n loki loki-<component>

# Or restart all components
kubectl rollout restart statefulset -n loki loki-backend
kubectl rollout restart deployment -n loki loki-write
kubectl rollout restart deployment -n loki loki-read

# Check Loki version
kubectl get helmrelease -n loki loki -o jsonpath='{.spec.chart.spec.version}'

# Consider upgrading if on old version
kubectl edit helmrelease -n loki loki
# Update: spec.chart.spec.version: "6.x.x"  # Latest stable
```

#### Issue: MinIO consuming high memory
**Cause:** High I/O load or large file operations  
**Fix:**
```bash
# Check MinIO memory usage
kubectl top pod -n loki -l app.kubernetes.io/name=minio

# Increase MinIO memory
kubectl edit helmrelease -n loki loki
# Update:
# minio:
#   resources:
#     limits:
#       memory: 2Gi
#     requests:
#       memory: 1Gi
```

### Step 3: Implement memory monitoring

```bash
# Add resource metrics collection if not already present
# Should be automatic with metrics-server

# Query Prometheus for memory trends
kubectl port-forward -n prometheus svc/prometheus-kube-prometheus-prometheus 9090:9090
# Query: container_memory_usage_bytes{namespace="loki"}
```

### Step 4: Restart pods if OOMKilled

```bash
# If pods are in CrashLoopBackOff due to OOM
# First increase limits, then restart
kubectl delete pod -n loki <oomkilled-pod>
```

## Verification

1. Check memory usage is stable:
```bash
# Monitor memory over time
watch kubectl top pods -n loki

# Should show stable or decreasing memory usage
```

2. Verify no OOM kills occurring:
```bash
kubectl get events -n loki --watch | grep -i oom
# Should see no new OOM events
```

3. Check all pods are running:
```bash
kubectl get pods -n loki
# All pods should be Running with reasonable restart count
```

4. Verify Loki is functioning:
```bash
# Port forward
kubectl port-forward -n loki svc/loki-gateway 3100:80

# Test query
curl -G -s "http://localhost:3100/loki/api/v1/query" \
  --data-urlencode 'query={namespace="loki"}' | jq .
```

5. Monitor Go GC metrics:
```bash
# Check garbage collection frequency
curl http://localhost:3100/metrics | grep go_gc_duration_seconds
```

## Prevention

1. **Set appropriate memory limits**
   - Monitor actual usage patterns
   - Set limits 20-30% above typical usage
   - Use requests to guarantee minimum

2. **Right-size components**
   - Read: 1-2Gi per replica
   - Write: 1-2Gi per replica  
   - Backend: 2-3Gi per replica
   - MinIO: 1-2Gi

3. **Implement horizontal scaling**
   - Use HPA based on memory metrics
   - Scale read/write independently
   - Configure proper replica counts

4. **Reduce label cardinality**
   - Use <10 labels per stream
   - Avoid high-cardinality values
   - Index labels intelligently

5. **Configure query limits**
   - Limit query time ranges
   - Limit results per query
   - Set query timeout

6. **Monitor and alert**
   - Alert at 80% memory usage
   - Track memory growth trends
   - Monitor OOMKill events

7. **Regular maintenance**
   - Enable compaction
   - Set retention policies
   - Clean up unused streams

## Memory Sizing Guidelines

```yaml
# Recommended memory settings by ingestion rate

# Low volume (< 1GB/day)
backend:
  resources:
    limits:
      memory: 1Gi
read:
  resources:
    limits:
      memory: 1Gi
write:
  resources:
    limits:
      memory: 1Gi

# Medium volume (1-10GB/day) - CURRENT RECOMMENDATION
backend:
  resources:
    limits:
      memory: 2Gi
read:
  resources:
    limits:
      memory: 2Gi
write:
  resources:
    limits:
      memory: 2Gi

# High volume (> 10GB/day)
backend:
  resources:
    limits:
      memory: 4Gi
    replicas: 2
read:
  resources:
    limits:
      memory: 4Gi
    replicas: 3
write:
  resources:
    limits:
      memory: 4Gi
    replicas: 3
```

## Related Alerts

- `LokiDown`
- `LokiReadPathSlow`
- `LokiWritePathDown`
- `LokiIngestionErrors`

## Escalation

If memory issues persist:
1. Review log volume and patterns
2. Consider implementing log sampling
3. Evaluate alternative storage backends
4. Review Loki architecture for scale

## Additional Resources

- [Loki Scaling](https://grafana.com/docs/loki/latest/operations/scalability/)
- [Loki Resource Sizing](https://grafana.com/docs/loki/latest/operations/recording-rules/)
- [Kubernetes Resource Management](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)
- [Go Memory Management](https://go.dev/doc/gc-guide)

