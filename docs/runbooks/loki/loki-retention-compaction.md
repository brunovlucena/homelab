# 🚨 Runbook: Loki Retention and Compaction Issues

## Alert Information

**Alert Name:** `LokiRetentionIssues` / `LokiCompactionFailing`  
**Severity:** Warning  
**Component:** Loki  
**Service:** Data Management

## Symptom

Loki is not properly cleaning up old logs or compacting chunks. Storage usage growing unexpectedly.

## Impact

- **User Impact:** LOW - No immediate user-facing issues
- **Business Impact:** MODERATE - Rising storage costs
- **Data Impact:** NONE - Data not being lost, just not cleaned

## Diagnosis

### 1. Check Current Storage Usage

```bash
# Check MinIO disk usage
kubectl exec -n loki <minio-pod> -- df -h /export

# Check PVC usage
kubectl get pvc -n loki
kubectl describe pvc -n loki data-loki-minio-0
```

### 2. Check Retention Configuration

```bash
# Check configured retention period
kubectl get helmrelease -n loki loki -o yaml | grep -A 5 "retention_period"

# Current configuration: 744h (31 days)
```

### 3. Check Compactor Status

```bash
# Check compactor logs (runs on backend component)
kubectl logs -n loki -l app.kubernetes.io/component=backend --tail=200 | grep -i "compact\|retention"

# Check for compaction errors
kubectl logs -n loki -l app.kubernetes.io/component=backend --tail=500 | grep -i "error" | grep -i "compact"
```

### 4. Check Compaction Metrics

```bash
# Port forward to Loki
kubectl port-forward -n loki svc/loki-gateway 3100:80

# Check compaction metrics
curl http://localhost:3100/metrics | grep "loki_compactor"
```

### 5. List Objects in Storage

```bash
# Check how many objects are in MinIO
kubectl exec -n loki <minio-pod> -- mc ls --recursive local/loki | wc -l

# Check oldest data
kubectl exec -n loki <minio-pod> -- mc ls --recursive local/loki | head -20
```

## Resolution Steps

### Step 1: Verify retention is enabled

```bash
# Check if retention is configured
kubectl get helmrelease -n loki loki -o yaml | grep -B 5 -A 10 "retention"

# Should see:
# loki:
#   limits_config:
#     retention_period: 744h
```

### Step 2: Common Issues and Fixes

#### Issue: Retention not enabled or not working
**Cause:** Compactor not configured properly  
**Fix:**
```bash
# Enable retention and compaction
kubectl edit helmrelease -n loki loki
# Add or verify:
# loki:
#   limits_config:
#     retention_enabled: true
#     retention_period: 744h  # 31 days
#   compactor:
#     retention_enabled: true
#     delete_request_store: s3
#   storage_config:
#     boltdb_shipper:
#       active_index_directory: /var/loki/index
#       cache_location: /var/loki/cache

# Force reconciliation
flux reconcile helmrelease loki -n loki

# Verify backend pods restart with new config
kubectl rollout status statefulset -n loki loki-backend
```

#### Issue: Compaction running but not cleaning up
**Cause:** Data not old enough or compactor interval too long  
**Fix:**
```bash
# Check compactor logs for activity
kubectl logs -n loki -l app.kubernetes.io/component=backend --tail=500 | grep -i "compact"

# Adjust compaction interval
kubectl edit helmrelease -n loki loki
# Add:
# loki:
#   compactor:
#     working_directory: /tmp/loki-compactor
#     compaction_interval: 10m  # Run more frequently
#     retention_enabled: true
#     retention_delete_delay: 2h  # Delay before deletion
#     retention_delete_worker_count: 150

# Wait for reconciliation
flux reconcile helmrelease loki -n loki
```

#### Issue: Compactor failing due to S3 errors
**Cause:** Storage backend issues  
**Fix:**
```bash
# Check MinIO health
kubectl get pods -n loki -l app.kubernetes.io/name=minio

# Check compactor can access S3
kubectl logs -n loki -l app.kubernetes.io/component=backend --tail=100 | grep -i "s3\|minio"

# Test S3 connectivity
kubectl exec -it -n loki <loki-backend-pod> -- sh -c 'wget -O- http://loki-minio:9000/minio/health/live'

# Restart backend if needed
kubectl rollout restart statefulset -n loki loki-backend
```

#### Issue: Disk full preventing compaction
**Cause:** Not enough space for compaction working directory  
**Fix:**
```bash
# Check disk usage on backend pod
kubectl exec -n loki <loki-backend-pod> -- df -h

# Increase ephemeral storage if needed
kubectl edit helmrelease -n loki loki
# Update:
# backend:
#   resources:
#     limits:
#       ephemeral-storage: 10Gi

# Or expand MinIO PVC
kubectl edit pvc -n loki data-loki-minio-0
# Increase: spec.resources.requests.storage: 100Gi
```

#### Issue: Retention period too long
**Cause:** Keeping logs longer than necessary  
**Fix:**
```bash
# Reduce retention period
kubectl edit helmrelease -n loki loki
# Update:
# loki:
#   limits_config:
#     retention_period: 336h  # 14 days instead of 31

# Logs older than 14 days will be deleted on next compaction cycle
```

#### Issue: Per-tenant retention not working
**Cause:** Multi-tenancy not configured or limits per tenant  
**Fix:**
```bash
# For single-tenant mode (current setup with auth_enabled: false)
# Retention applies globally

# If needed, enable per-stream retention
kubectl edit helmrelease -n loki loki
# Add:
# loki:
#   limits_config:
#     retention_period: 744h  # Default
#     retention_stream:
#     - selector: '{namespace="production"}'
#       priority: 1
#       period: 2160h  # 90 days for production
#     - selector: '{namespace="development"}'
#       priority: 2
#       period: 168h  # 7 days for dev
```

### Step 3: Manually trigger compaction (if needed)

```bash
# Compaction runs automatically, but you can force by restarting backend
kubectl rollout restart statefulset -n loki loki-backend

# Monitor compaction progress
kubectl logs -n loki -l app.kubernetes.io/component=backend --follow | grep -i compact
```

### Step 4: Verify old data is being deleted

```bash
# Check oldest data in MinIO
kubectl exec -n loki <minio-pod> -- mc ls --recursive local/loki | head -30

# Check dates of oldest chunks
# Should not see data older than retention period (744h = 31 days)

# Check compaction metrics
curl http://localhost:3100/metrics | grep loki_compactor_oldest_pending_delete_request_age_seconds
```

## Verification

1. Check storage usage trend:
```bash
# Monitor over time
watch kubectl exec -n loki <minio-pod> -- df -h /export

# Should see usage stabilize or decrease after retention kicks in
```

2. Verify retention is working:
```bash
# Check compactor logs for successful deletion
kubectl logs -n loki -l app.kubernetes.io/component=backend --tail=200 | grep -i "deleted\|removed"

# Check metrics
curl http://localhost:3100/metrics | grep loki_compactor_blocks_marked_for_deletion_total
```

3. Verify no compaction errors:
```bash
kubectl logs -n loki -l app.kubernetes.io/component=backend --tail=500 | grep -i "error" | grep -i "compact"
# Should be empty or minimal
```

4. Check data age distribution:
```bash
# List all chunks and check dates
kubectl exec -n loki <minio-pod> -- mc ls --recursive local/loki | \
  awk '{print $3}' | sort | uniq -c

# Oldest data should be within retention period
```

## Manual Data Cleanup (Emergency Only)

⚠️ **WARNING**: Only use if automatic retention is failing and disk is critically full

```bash
# 1. Backup data first!
kubectl exec -n loki <minio-pod> -- mc mirror local/loki /backup/loki-$(date +%Y%m%d)

# 2. Identify old data to delete
kubectl exec -n loki <minio-pod> -- mc ls --recursive local/loki | grep "2024-01" | head

# 3. Carefully delete specific old chunks
kubectl exec -n loki <minio-pod> -- mc rm --recursive --force local/loki/fake/index/...

# 4. Restart Loki components to resync
kubectl rollout restart statefulset -n loki loki-backend
kubectl rollout restart deployment -n loki loki-write
kubectl rollout restart deployment -n loki loki-read

# 5. Verify system health
kubectl get pods -n loki
```

## Prevention

1. **Monitor storage growth**
   - Set alerts at 70% capacity
   - Track growth rate
   - Project future needs

2. **Right-size retention period**
   - Balance compliance vs cost
   - Current: 31 days (744h)
   - Consider: 14 days for dev, 90 days for prod

3. **Enable compaction monitoring**
   - Alert on compaction failures
   - Track compaction duration
   - Monitor deleted chunks

4. **Implement tiered retention**
   - Keep recent logs hot
   - Move older logs to cold storage
   - Delete very old logs

5. **Regular capacity planning**
   - Review log volume trends
   - Plan storage expansion
   - Optimize log collection

6. **Configure appropriate limits**
   - Set ingestion rate limits
   - Limit per-stream data
   - Reduce label cardinality

## Retention Best Practices

```yaml
# Recommended retention configuration

# Development/Testing
loki:
  limits_config:
    retention_period: 168h  # 7 days

# Production (current)
loki:
  limits_config:
    retention_period: 744h  # 31 days
    retention_enabled: true

# Compliance/Audit
loki:
  limits_config:
    retention_period: 2160h  # 90 days

# Long-term archival
# Consider exporting to external object storage
# Use Loki export API or backup MinIO data
```

## Storage Optimization Tips

1. **Reduce log volume at source**
   - Filter out debug logs in production
   - Sample high-volume logs
   - Aggregate similar events

2. **Optimize label usage**
   - Use <10 labels per stream
   - Avoid high-cardinality labels
   - Use structured metadata for extra fields

3. **Enable compression**
   - Loki automatically compresses chunks
   - Verify compression is working
   - Monitor compression ratio

4. **Implement sampling**
   - Sample verbose logs
   - Keep all errors and warnings
   - Use probabilistic sampling

## Related Alerts

- `LokiStorageIssues`
- `LokiHighMemory`
- `LokiDown`

## Escalation

If retention/compaction issues persist:
1. Review log volume and growth patterns
2. Consider increasing storage capacity
3. Evaluate log sampling strategies
4. Review backup and archival policies

## Additional Resources

- [Loki Retention](https://grafana.com/docs/loki/latest/operations/storage/retention/)
- [Loki Compactor](https://grafana.com/docs/loki/latest/operations/storage/retention/#compactor)
- [Storage Configuration](https://grafana.com/docs/loki/latest/storage/)
- [Loki Limits Configuration](https://grafana.com/docs/loki/latest/configuration/#limits_config)

