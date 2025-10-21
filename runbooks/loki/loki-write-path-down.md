# 🚨 Runbook: Loki Write Path Down

## Alert Information

**Alert Name:** `LokiWritePathDown`  
**Severity:** Critical  
**Component:** Loki  
**Service:** Log Ingestion

## Symptom

Loki write path is unavailable. New logs cannot be ingested but queries may still work.

## Impact

- **User Impact:** MODERATE - Existing logs still queryable
- **Business Impact:** HIGH - New logs not being stored
- **Data Impact:** CRITICAL - Active log loss if buffer exhausted

## Diagnosis

### 1. Check Write Pod Status

```bash
kubectl get pods -n loki -l app.kubernetes.io/component=write
kubectl describe pod -n loki -l app.kubernetes.io/component=write
```

### 2. Check Write Pod Logs

```bash
kubectl logs -n loki -l app.kubernetes.io/component=write --tail=100
kubectl logs -n loki -l app.kubernetes.io/component=write --previous  # If restarting
```

### 3. Check Storage Backend Connectivity

```bash
# Check MinIO status
kubectl get pods -n loki -l app.kubernetes.io/name=minio

# Test S3 connectivity from write pod
kubectl exec -it -n loki <loki-write-pod> -- wget -O- http://loki-minio:9000/minio/health/live
```

### 4. Check Write Path Metrics

```bash
# Port forward to Loki
kubectl port-forward -n loki svc/loki-gateway 3100:80

# Check write metrics
curl http://localhost:3100/metrics | grep -i "loki_distributor\|loki_ingester"
```

### 5. Check Events

```bash
kubectl get events -n loki --sort-by='.lastTimestamp' | grep write
```

## Resolution Steps

### Step 1: Verify write pod status

```bash
WRITE_STATUS=$(kubectl get pods -n loki -l app.kubernetes.io/component=write -o jsonpath='{.items[*].status.phase}')
echo "Write pods status: $WRITE_STATUS"
```

### Step 2: Check for specific errors

```bash
# Check for storage errors
kubectl logs -n loki -l app.kubernetes.io/component=write --tail=50 | grep -i "s3\|storage\|minio"

# Check for authentication errors
kubectl logs -n loki -l app.kubernetes.io/component=write --tail=50 | grep -i "auth\|access denied"

# Check for OOM errors
kubectl describe pod -n loki -l app.kubernetes.io/component=write | grep -i "oom"
```

### Step 3: Common Issues and Fixes

#### Issue: Cannot connect to MinIO
**Cause:** Storage backend unavailable or network issue  
**Fix:**
```bash
# Check MinIO health
kubectl get pods -n loki -l app.kubernetes.io/name=minio

# Test connectivity
kubectl exec -it -n loki <loki-write-pod> -- sh -c 'nc -zv loki-minio 9000'

# Restart MinIO if needed
kubectl rollout restart statefulset -n loki loki-minio
kubectl wait --for=condition=ready pod -n loki -l app.kubernetes.io/name=minio --timeout=300s
```

#### Issue: S3 authentication failure
**Cause:** Invalid credentials or secret not mounted  
**Fix:**
```bash
# Verify secret exists
kubectl get secret -n loki loki-minio-secret

# Check if secret is mounted in pod
kubectl describe pod -n loki <loki-write-pod> | grep -A 5 "Mounts:"

# Verify secret content (base64 encoded)
kubectl get secret -n loki loki-minio-secret -o jsonpath='{.data}'

# Force reconcile to reload secrets
flux reconcile helmrelease loki -n loki
```

#### Issue: Write path OOMKilled
**Cause:** Insufficient memory for ingestion rate  
**Fix:**
```bash
# Check memory usage
kubectl top pod -n loki -l app.kubernetes.io/component=write

# Increase memory limit
kubectl edit helmrelease -n loki loki
# Update: write.resources.limits.memory

# Or scale up write replicas
kubectl scale deployment -n loki loki-write --replicas=3
```

#### Issue: Disk full on MinIO
**Cause:** Storage exhausted  
**Fix:**
```bash
# Check MinIO disk usage
kubectl exec -n loki <minio-pod> -- df -h /export

# Check PVC size
kubectl get pvc -n loki

# Increase PVC size if supported by storage class
kubectl edit pvc -n loki data-loki-minio-0
# Increase spec.resources.requests.storage to 100Gi

# Or reduce retention period
kubectl edit helmrelease -n loki loki
# Update: loki.limits_config.retention_period to 336h (14 days)
```

### Step 4: Restart write pods

```bash
kubectl rollout restart deployment -n loki loki-write
kubectl rollout status deployment -n loki loki-write
```

### Step 5: Verify ingestion is working

```bash
# Check write pod logs for successful writes
kubectl logs -n loki -l app.kubernetes.io/component=write --tail=20 | grep -i "pushed\|success"
```

## Verification

1. Check write pods are running and ready:
```bash
kubectl get pods -n loki -l app.kubernetes.io/component=write
```

2. Test log ingestion:
```bash
# Send a test log
kubectl run test-logger --image=busybox --restart=Never -- sh -c "echo 'Test log entry'"

# Wait a few seconds, then query
kubectl port-forward -n loki svc/loki-gateway 3100:80
curl -G -s "http://localhost:3100/loki/api/v1/query" \
  --data-urlencode 'query={namespace="default"}' \
  --data-urlencode 'limit=10' | jq .

# Cleanup test pod
kubectl delete pod test-logger
```

3. Monitor ingestion rate:
```bash
# Check distributor metrics
curl http://localhost:3100/metrics | grep loki_distributor_bytes_received_total
```

4. Verify from log sources:
```bash
# Check Alloy is successfully sending logs
kubectl logs -n alloy -l app.kubernetes.io/name=alloy --tail=20 | grep -i loki
```

## Prevention

1. **Monitor write path health**
   - Alert on write pod restarts
   - Monitor ingestion rate and latency
   - Track failed write requests

2. **Ensure adequate storage**
   - Monitor MinIO disk usage
   - Set up alerts at 70% capacity
   - Implement log retention policies

3. **Right-size write replicas**
   - Scale based on ingestion rate
   - Currently: 2 write replicas
   - Consider 3+ for high availability

4. **Configure resource limits appropriately**
   - Monitor actual usage patterns
   - Set limits with 20-30% headroom
   - Use HPA for automatic scaling

5. **Implement circuit breakers**
   - Configure rate limits
   - Set up backpressure handling
   - Buffer logs at source (Alloy)

## Related Alerts

- `LokiDown`
- `LokiIngestionErrors`
- `LokiStorageIssues`
- `LokiHighMemory`

## Escalation

If write path issues persist:
1. Check MinIO data integrity and performance
2. Review ingestion rate spikes
3. Verify network policies allow Loki ingestion
4. Check for disk I/O bottlenecks

## Additional Resources

- [Loki Write Path Architecture](https://grafana.com/docs/loki/latest/fundamentals/architecture/components/#write-path)
- [Scaling Loki](https://grafana.com/docs/loki/latest/operations/scalability/)
- [Loki Limits Configuration](https://grafana.com/docs/loki/latest/configuration/#limits_config)

