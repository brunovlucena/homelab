# 🚨 Runbook: Prometheus Storage Full

## Alert Information

**Alert Name:** `PrometheusStorageFull`  
**Severity:** Warning/Critical  
**Component:** prometheus  
**Service:** storage

## Symptom

Prometheus storage (PVC) usage has exceeded warning threshold (>80%) or critical threshold (>95%). Running out of storage will prevent new metrics from being stored.

## Impact

- **User Impact:** LOW (at warning) to CRITICAL (when full)
- **Business Impact:** HIGH - Risk of metrics loss
- **Data Impact:** CRITICAL - Cannot store new metrics when full

## Diagnosis

### 1. Check Storage Usage

```bash
# Check PVC usage
kubectl exec -n prometheus prometheus-prometheus-kube-prometheus-prometheus-0 -- \
  df -h /prometheus

# Check PVC size
kubectl get pvc -n prometheus
```

### 2. Check Storage Usage Over Time

```promql
# Current storage usage (bytes)
prometheus_tsdb_storage_blocks_bytes + prometheus_tsdb_head_chunks_storage_size_bytes

# Storage usage percentage
100 * (prometheus_tsdb_storage_blocks_bytes + prometheus_tsdb_head_chunks_storage_size_bytes) / 
node_filesystem_size_bytes{mountpoint="/prometheus"}

# Storage growth rate (per day)
rate(prometheus_tsdb_storage_blocks_bytes[24h]) * 86400
```

### 3. Check TSDB Stats

```bash
# Access Prometheus UI
kubectl port-forward -n prometheus svc/prometheus-kube-prometheus-prometheus 9090:9090

# Open http://localhost:9090/tsdb-status
# Check:
# - Number of series
# - Number of blocks
# - Block size breakdown
```

### 4. Check Retention Settings

```bash
# Check current retention
kubectl get prometheus -n prometheus prometheus-kube-prometheus-prometheus \
  -o jsonpath='{.spec.retention}'

# Check retention size (if set)
kubectl get prometheus -n prometheus prometheus-kube-prometheus-prometheus \
  -o jsonpath='{.spec.retentionSize}'
```

### 5. Check WAL Size

```bash
# Check WAL directory size
kubectl exec -n prometheus prometheus-prometheus-kube-prometheus-prometheus-0 -- \
  du -sh /prometheus/wal

# List WAL segments
kubectl exec -n prometheus prometheus-prometheus-kube-prometheus-prometheus-0 -- \
  ls -lh /prometheus/wal
```

## Resolution Steps

### Step 1: Assess severity and time to full

```promql
# Predict when storage will be full (hours)
(node_filesystem_avail_bytes{mountpoint="/prometheus"} / 
rate(prometheus_tsdb_storage_blocks_bytes[24h])) / 3600

# If less than 24 hours: URGENT
# If less than 7 days: Plan action
```

### Step 2: Common Issues and Fixes

#### Issue: Normal Growth - Need More Space
**Cause:** Storage legitimately needs expansion  
**Fix:**
```bash
# Option 1: Expand PVC (if storage class supports it)
kubectl edit pvc -n prometheus \
  prometheus-prometheus-kube-prometheus-prometheus-db-prometheus-prometheus-kube-prometheus-prometheus-0

# Change spec.resources.requests.storage to larger value
# Example: 50Gi -> 100Gi

# Check if resize is in progress
kubectl get pvc -n prometheus -w

# Option 2: If storage class doesn't support expansion
# You'll need to create new PVC and migrate data
# See "Data Migration" section below
```

#### Issue: Retention Too Long
**Cause:** Keeping data longer than necessary  
**Fix:**
```bash
# Reduce retention time via Helm values
# retention: 7d  # Reduce from current (e.g., 30d -> 7d)

# Apply changes
flux reconcile helmrelease kube-prometheus-stack -n prometheus

# Verify new retention
kubectl get prometheus -n prometheus prometheus-kube-prometheus-prometheus \
  -o jsonpath='{.spec.retention}'

# Wait for Prometheus to clean up old blocks
# Check storage after a few minutes
```

#### Issue: High Cardinality
**Cause:** Too many unique time series consuming space  
**Fix:**
```bash
# Identify high-cardinality metrics
# In Prometheus UI: topk(20, count by (__name__)({__name__=~".+"}))

# Options:
# 1. Drop high-cardinality metrics
# 2. Reduce scrape frequency
# 3. Use metric relabeling
# See high-memory-usage.md for detailed steps

# After reducing cardinality, old data will expire based on retention
```

#### Issue: Failed Compactions
**Cause:** TSDB compaction failures leaving orphaned blocks  
**Fix:**
```bash
# Check for compaction errors
kubectl logs -n prometheus -l app.kubernetes.io/name=prometheus \
  | grep -i "compaction\|compact"

# Check block statistics
kubectl exec -n prometheus prometheus-prometheus-kube-prometheus-prometheus-0 -- \
  promtool tsdb analyze /prometheus

# If compaction is stuck, restart Prometheus
kubectl delete pod -n prometheus -l app.kubernetes.io/name=prometheus
```

#### Issue: WAL Not Truncated
**Cause:** Write-Ahead Log growing without truncation  
**Fix:**
```bash
# Check WAL size
kubectl exec -n prometheus prometheus-prometheus-kube-prometheus-prometheus-0 -- \
  du -sh /prometheus/wal

# Normal WAL size: < 2GB
# If WAL is very large (>10GB), restart Prometheus
kubectl delete pod -n prometheus -l app.kubernetes.io/name=prometheus

# WAL will be replayed and truncated on restart
```

#### Issue: Orphaned Blocks
**Cause:** Old block directories not cleaned up  
**Fix:**
```bash
# List blocks
kubectl exec -n prometheus prometheus-prometheus-kube-prometheus-prometheus-0 -- \
  ls -lh /prometheus

# Check for blocks older than retention period
# Prometheus should auto-cleanup, but you can manually remove if needed

# DANGEROUS: Only do this if you're sure
# kubectl exec -n prometheus prometheus-prometheus-kube-prometheus-prometheus-0 -- \
#   rm -rf /prometheus/<block-id>
```

### Step 3: Immediate Space Relief (Emergency)

```bash
# If critically low on space and need immediate relief

# Option 1: Reduce retention drastically (temporary)
kubectl edit prometheus -n prometheus prometheus-kube-prometheus-prometheus
# Set retention: 1d (temporary)
# Change back after expanding storage

# Option 2: Delete oldest blocks manually (LAST RESORT)
kubectl exec -n prometheus prometheus-prometheus-kube-prometheus-prometheus-0 -- \
  bash -c 'cd /prometheus && ls -t | tail -5'
# Review block names and delete oldest if needed
```

### Step 4: Set Retention Size Limit

```bash
# Set size-based retention (preferred over time-based)
# Edit via Helm values:
# retentionSize: "45GB"  # Set to ~90% of PVC size

# This ensures Prometheus never fills the disk completely
flux reconcile helmrelease kube-prometheus-stack -n prometheus
```

### Step 5: Implement Remote Write (Long-term)

```yaml
# Configure remote write to offload long-term storage
# Helm values:
prometheus:
  prometheusSpec:
    remoteWrite:
      - url: "http://thanos-receive:19291/api/v1/receive"
        queueConfig:
          capacity: 10000
          maxShards: 50
    retention: 7d  # Keep only recent data locally
```

## Data Migration (If PVC Can't Expand)

```bash
# 1. Create new larger PVC
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: prometheus-new-storage
  namespace: prometheus
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 200Gi  # Larger size
  storageClassName: <same-as-current>
EOF

# 2. Scale down Prometheus
kubectl scale statefulset -n prometheus \
  prometheus-prometheus-kube-prometheus-prometheus --replicas=0

# 3. Copy data to new PVC
kubectl run -n prometheus pv-migration --image=busybox --restart=Never \
  --overrides='
  {
    "spec": {
      "containers": [{
        "name": "migration",
        "image": "busybox",
        "command": ["sh", "-c", "cp -a /old/* /new/"],
        "volumeMounts": [
          {"name": "old", "mountPath": "/old"},
          {"name": "new", "mountPath": "/new"}
        ]
      }],
      "volumes": [
        {"name": "old", "persistentVolumeClaim": {"claimName": "prometheus-old-storage"}},
        {"name": "new", "persistentVolumeClaim": {"claimName": "prometheus-new-storage"}}
      ]
    }
  }'

# 4. Wait for migration
kubectl wait --for=condition=completed pod/pv-migration -n prometheus --timeout=30m

# 5. Update StatefulSet to use new PVC
kubectl edit statefulset -n prometheus prometheus-prometheus-kube-prometheus-prometheus
# Update volumeClaimTemplates to reference new PVC

# 6. Scale up
kubectl scale statefulset -n prometheus \
  prometheus-prometheus-kube-prometheus-prometheus --replicas=1
```

## Verification

1. Check storage usage decreased:
```bash
kubectl exec -n prometheus prometheus-prometheus-kube-prometheus-prometheus-0 -- \
  df -h /prometheus
```

2. Verify retention settings:
```bash
kubectl get prometheus -n prometheus prometheus-kube-prometheus-prometheus \
  -o jsonpath='{.spec.retention}'
```

3. Monitor storage usage:
```promql
100 * (prometheus_tsdb_storage_blocks_bytes + prometheus_tsdb_head_chunks_storage_size_bytes) / 
node_filesystem_size_bytes{mountpoint="/prometheus"}
```

4. Check TSDB health:
```bash
kubectl exec -n prometheus prometheus-prometheus-kube-prometheus-prometheus-0 -- \
  promtool tsdb analyze /prometheus
```

5. Verify metrics still being collected:
```promql
rate(prometheus_tsdb_head_samples_appended_total[5m])
```

## Prevention

1. Set retentionSize to 90% of PVC size
2. Monitor storage usage and set alerts
3. Implement remote write for long-term storage
4. Regular cardinality audits
5. Appropriate retention period for use case
6. Use storage class that supports PVC expansion
7. Monitor storage growth rate
8. Plan capacity based on growth trends
9. Implement automated PVC expansion
10. Regular cleanup of unnecessary metrics

## Related Alerts

- `PrometheusStorageAlmostFull`
- `PrometheusTSDBCompactionsFailing`
- `PrometheusHighMemoryUsage`
- `PrometheusHighCardinality`

## Escalation

If the issue persists after following these steps:
1. Check for underlying storage infrastructure issues
2. Review historical storage growth patterns
3. Audit all metrics for necessity
4. Consider Prometheus federation/sharding
5. Contact storage team or on-call engineer

## Additional Resources

- [Prometheus Storage Documentation](https://prometheus.io/docs/prometheus/latest/storage/)
- [TSDB Format](https://prometheus.io/docs/prometheus/latest/storage/)
- [Retention Configuration](https://prometheus.io/docs/prometheus/latest/storage/#retention)
- [Remote Write](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#remote_write)
- [Kubernetes PVC Expansion](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#expanding-persistent-volumes-claims)

---

**Last Updated:** 2025-10-15  
**Version:** 1.0

