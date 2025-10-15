# 🚨 Runbook: Loki Storage Issues

## Alert Information

**Alert Name:** `LokiStorageIssues`  
**Severity:** Critical  
**Component:** Loki  
**Service:** Storage Backend (MinIO)

## Symptom

Loki is experiencing issues with the storage backend (MinIO/S3). Reads or writes may be failing.

## Impact

- **User Impact:** HIGH - Log queries failing or incomplete
- **Business Impact:** CRITICAL - Potential data loss
- **Data Impact:** CRITICAL - Risk of permanent log loss

## Diagnosis

### 1. Check MinIO Pod Status

```bash
kubectl get pods -n loki -l app.kubernetes.io/name=minio
kubectl describe pod -n loki -l app.kubernetes.io/name=minio
```

### 2. Check MinIO Logs

```bash
kubectl logs -n loki -l app.kubernetes.io/name=minio --tail=100
```

### 3. Check Storage PVC Status

```bash
kubectl get pvc -n loki
kubectl describe pvc -n loki data-loki-minio-0
```

### 4. Check Disk Usage

```bash
kubectl exec -n loki <minio-pod> -- df -h /export
kubectl exec -n loki <minio-pod> -- du -sh /export/*
```

### 5. Test MinIO Connectivity

```bash
# From Loki write pod
kubectl exec -it -n loki <loki-write-pod> -- sh -c 'wget -O- http://loki-minio:9000/minio/health/live'

# Port forward and test locally
kubectl port-forward -n loki svc/loki-minio 9000:9000
curl http://localhost:9000/minio/health/live
```

### 6. Check Loki Storage Errors

```bash
# Check all Loki components for storage errors
kubectl logs -n loki -l app.kubernetes.io/name=loki --tail=200 | grep -i "s3\|storage\|minio"
```

## Resolution Steps

### Step 1: Identify storage issue type

```bash
# Check MinIO status
MINIO_STATUS=$(kubectl get pods -n loki -l app.kubernetes.io/name=minio -o jsonpath='{.items[0].status.phase}')
echo "MinIO Status: $MINIO_STATUS"

# Check disk space
kubectl exec -n loki <minio-pod> -- df -h /export | grep -v "Filesystem"
```

### Step 2: Common Issues and Fixes

#### Issue: MinIO pod not running
**Cause:** Pod crashed or failed to start  
**Fix:**
```bash
# Check why pod is not running
kubectl describe pod -n loki <minio-pod> | grep -A 10 "Events:"

# Check previous logs if pod restarted
kubectl logs -n loki <minio-pod> --previous

# Restart MinIO StatefulSet
kubectl rollout restart statefulset -n loki loki-minio

# Wait for MinIO to be ready
kubectl wait --for=condition=ready pod -n loki -l app.kubernetes.io/name=minio --timeout=300s

# After MinIO is ready, restart Loki components
kubectl rollout restart statefulset -n loki loki-backend
kubectl rollout restart deployment -n loki loki-write
kubectl rollout restart deployment -n loki loki-read
```

#### Issue: Disk full
**Cause:** Storage capacity exhausted  
**Fix:**
```bash
# Check current usage
kubectl exec -n loki <minio-pod> -- df -h /export

# Option 1: Increase PVC size (if storage class supports it)
kubectl edit pvc -n loki data-loki-minio-0
# Change: spec.resources.requests.storage: 50Gi -> 100Gi

# Option 2: Reduce Loki retention period
kubectl edit helmrelease -n loki loki
# Update:
# loki:
#   limits_config:
#     retention_period: 336h  # Reduce from 744h (31d) to 336h (14d)

# Force compaction and cleanup
kubectl exec -n loki <minio-pod> -- mc admin heal -r loki-minio/loki

# Option 3: Manually clean old data (DANGEROUS - backup first!)
kubectl exec -it -n loki <minio-pod> -- sh
# cd /export/loki
# ls -lt  # Check oldest chunks
# rm -rf <old-chunks>  # Remove carefully!
```

#### Issue: PVC not mounting
**Cause:** Storage class issue or PVC corruption  
**Fix:**
```bash
# Check PVC status
kubectl get pvc -n loki data-loki-minio-0 -o yaml

# Check storage class
kubectl get sc

# Check PV bound to PVC
kubectl get pv | grep loki

# If PVC stuck, try deleting and recreating MinIO
# CAUTION: This will lose all data unless backed up!
kubectl delete statefulset -n loki loki-minio
flux reconcile helmrelease loki -n loki
```

#### Issue: S3 authentication errors
**Cause:** Invalid or missing credentials  
**Fix:**
```bash
# Check secret exists
kubectl get secret -n loki loki-minio-secret

# Verify secret has required keys
kubectl get secret -n loki loki-minio-secret -o jsonpath='{.data}' | jq

# Recreate secret if needed (check scripts/create-secrets.sh)
# Then force Flux reconciliation
flux reconcile helmrelease loki -n loki
```

#### Issue: Network connectivity to MinIO
**Cause:** Service or network policy blocking access  
**Fix:**
```bash
# Check MinIO service
kubectl get svc -n loki loki-minio
kubectl describe svc -n loki loki-minio

# Test connectivity from write pod
kubectl exec -it -n loki <loki-write-pod> -- sh -c 'nc -zv loki-minio 9000'
kubectl exec -it -n loki <loki-write-pod> -- sh -c 'nslookup loki-minio'

# Check network policies
kubectl get networkpolicies -n loki

# If using service mesh, check policies
kubectl get servers -n loki
kubectl get httproutes -n loki
```

#### Issue: MinIO corruption or data inconsistency
**Cause:** Unclean shutdown or disk issues  
**Fix:**
```bash
# Run MinIO heal command
kubectl exec -n loki <minio-pod> -- mc admin heal -r local/loki

# Check MinIO consistency
kubectl exec -n loki <minio-pod> -- mc admin info local

# If severe corruption, may need to restore from backup
# (Ensure backups are configured!)
```

#### Issue: MinIO out of memory
**Cause:** Insufficient memory allocation  
**Fix:**
```bash
# Check current memory usage
kubectl top pod -n loki -l app.kubernetes.io/name=minio

# Increase memory limits
kubectl edit helmrelease -n loki loki
# Update:
# minio:
#   resources:
#     limits:
#       memory: 2Gi
#     requests:
#       memory: 1Gi
```

### Step 3: Verify storage is accessible

```bash
# Port forward to MinIO console
kubectl port-forward -n loki svc/loki-minio 9001:9001

# Access MinIO console at http://localhost:9001
# Username: root-user
# Password: (from secret)

# Or use CLI
kubectl exec -n loki <minio-pod> -- mc ls local/loki
```

## Verification

1. Check MinIO is healthy:
```bash
kubectl get pods -n loki -l app.kubernetes.io/name=minio

# Check health endpoint
kubectl port-forward -n loki svc/loki-minio 9000:9000
curl http://localhost:9000/minio/health/live
curl http://localhost:9000/minio/health/ready
```

2. Verify Loki can write to storage:
```bash
# Send test logs
kubectl run test-logger --image=busybox --restart=Never -- sh -c "echo 'Storage test log'"

# Wait a bit
sleep 30

# Verify it was written to MinIO
kubectl exec -n loki <minio-pod> -- mc ls local/loki --recursive | grep -c "."
# Should show files

kubectl delete pod test-logger
```

3. Verify Loki can read from storage:
```bash
# Port forward to Loki
kubectl port-forward -n loki svc/loki-gateway 3100:80

# Query historical logs
curl -G -s "http://localhost:3100/loki/api/v1/query_range" \
  --data-urlencode 'query={namespace="loki"}' \
  --data-urlencode 'start=now-1h' \
  --data-urlencode 'end=now' | jq .
```

4. Check storage metrics:
```bash
# Check Loki storage metrics
curl http://localhost:3100/metrics | grep loki_chunk_store
```

## Prevention

1. **Monitor storage capacity**
   - Alert at 70% disk usage
   - Set up automated cleanup
   - Implement retention policies

2. **Regular backups**
   - Backup MinIO data regularly
   - Test restore procedures
   - Document backup locations

3. **Configure appropriate retention**
   - Current: 744h (31 days)
   - Balance cost vs requirements
   - Implement tiered storage

4. **Set resource limits appropriately**
   - Monitor actual MinIO usage
   - Set limits with headroom
   - Use storage class with expansion support

5. **Implement high availability**
   - Consider MinIO distributed mode
   - Use external S3-compatible storage
   - Implement disaster recovery plan

6. **Regular health checks**
   - Monitor MinIO metrics
   - Check disk I/O performance
   - Verify backup integrity

## Backup Strategy

```bash
# Manual backup of MinIO data
kubectl exec -n loki <minio-pod> -- mc mirror local/loki /backup/loki-$(date +%Y%m%d)

# Or using volume snapshots (if supported)
kubectl create volumesnapshot loki-backup-$(date +%Y%m%d) \
  --source-kind=PersistentVolumeClaim \
  --source-name=data-loki-minio-0 \
  -n loki
```

## Related Alerts

- `LokiDown`
- `LokiWritePathDown`
- `LokiIngestionErrors`
- `LokiReadPathSlow`

## Escalation

If storage issues persist:
1. Check node disk health
2. Review storage class configuration
3. Consider migrating to external S3
4. Contact platform team for infrastructure issues

## Additional Resources

- [MinIO Administration](https://min.io/docs/minio/linux/administration/minio-console.html)
- [Loki Storage Configuration](https://grafana.com/docs/loki/latest/storage/)
- [Kubernetes Persistent Volumes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/)
- [MinIO on Kubernetes](https://min.io/docs/minio/kubernetes/upstream/)

