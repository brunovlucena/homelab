# 🚨 Runbook: Loki Ingestion Errors

## Alert Information

**Alert Name:** `LokiIngestionErrors`  
**Severity:** Warning  
**Component:** Loki  
**Service:** Log Ingestion

## Symptom

Loki is rejecting or dropping log entries. Some logs are not being stored.

## Impact

- **User Impact:** MODERATE - Partial log loss
- **Business Impact:** MODERATE - Incomplete observability
- **Data Impact:** MODERATE - Some logs permanently lost

## Diagnosis

### 1. Check Distributor Metrics

```bash
# Port forward to Loki
kubectl port-forward -n loki svc/loki-gateway 3100:80

# Check distributor errors
curl http://localhost:3100/metrics | grep -i "loki_distributor_bytes_received\|loki_distributor_lines_received"

# Check for rate limit errors
curl http://localhost:3100/metrics | grep -i "loki_discarded"
```

### 2. Check Write Pod Logs

```bash
kubectl logs -n loki -l app.kubernetes.io/component=write --tail=200 | grep -i "error\|reject\|discard\|limit"
```

### 3. Check Backend Pod Logs

```bash
kubectl logs -n loki -l app.kubernetes.io/component=backend --tail=100 | grep -i "error\|limit"
```

### 4. Check Alloy (Log Collector) Status

```bash
# Check Alloy pods
kubectl get pods -n alloy

# Check Alloy logs for Loki push errors
kubectl logs -n alloy -l app.kubernetes.io/name=alloy --tail=100 | grep -i "loki\|error\|failed"
```

### 5. Check for Rate Limiting

```bash
# Check ingestion rate metrics
curl http://localhost:3100/metrics | grep loki_ingester_streams_created_total
curl http://localhost:3100/metrics | grep loki_distributor_ingester_append_failures_total
```

## Resolution Steps

### Step 1: Identify error types

```bash
# Check for common error patterns
kubectl logs -n loki -l app.kubernetes.io/component=write --tail=300 | \
  grep -i "error" | sort | uniq -c | sort -rn | head -10
```

### Step 2: Common Issues and Fixes

#### Issue: Rate limiting errors
**Cause:** Ingestion rate exceeds configured limits  
**Symptoms:** `rate limit exceeded` or `too many streams`  
**Fix:**
```bash
# Check current limits
kubectl get helmrelease -n loki loki -o yaml | grep -A 20 "limits_config"

# Increase rate limits
kubectl edit helmrelease -n loki loki
# Update:
# loki:
#   limits_config:
#     ingestion_rate_mb: 10  # Increase from default 4MB
#     ingestion_burst_size_mb: 20  # Increase from default 6MB
#     per_stream_rate_limit: 5MB  # Increase per-stream limit
#     per_stream_rate_limit_burst: 10MB

# Wait for reconciliation
flux reconcile helmrelease loki -n loki
```

#### Issue: Too many streams error
**Cause:** Too many unique label combinations  
**Symptoms:** `maximum number of streams exceeded`  
**Fix:**
```bash
# Check stream count
curl http://localhost:3100/metrics | grep loki_ingester_streams

# Increase stream limit
kubectl edit helmrelease -n loki loki
# Update:
# loki:
#   limits_config:
#     max_streams_per_user: 0  # 0 = unlimited (use with caution)
#     # Or set a higher limit: 10000

# Review log labels to reduce cardinality
# Check Alloy configuration to use fewer labels
kubectl edit configmap -n alloy alloy-config
# Remove high-cardinality labels like pod_name, container_id
```

#### Issue: Out of order entries
**Cause:** Logs arriving with timestamps older than last write  
**Symptoms:** `entry out of order` or `entry too far behind`  
**Fix:**
```bash
# Check current out-of-order window
kubectl get helmrelease -n loki loki -o yaml | grep reject_old_samples

# Increase acceptance window
kubectl edit helmrelease -n loki loki
# Update:
# loki:
#   limits_config:
#     reject_old_samples: false  # Accept old samples
#     reject_old_samples_max_age: 168h  # Accept samples up to 7 days old
#     creation_grace_period: 10m  # Accept future timestamps up to 10m

# Fix timestamp issues at source (Alloy)
kubectl edit configmap -n alloy alloy-config
# Ensure proper timestamp extraction
```

#### Issue: Invalid log format
**Cause:** Logs don't meet Loki requirements  
**Symptoms:** `invalid log line` or `parse error`  
**Fix:**
```bash
# Check what's being rejected
kubectl logs -n loki -l app.kubernetes.io/component=write --tail=100 | grep "invalid"

# Common issues:
# - Missing timestamp
# - Invalid JSON
# - Malformed labels

# Fix at Alloy level
kubectl edit configmap -n alloy alloy-config
# Ensure proper log parsing and validation
```

#### Issue: Storage backend errors
**Cause:** MinIO rejecting writes  
**Symptoms:** `S3 error` or `storage error`  
**Fix:**
```bash
# Check MinIO health
kubectl get pods -n loki -l app.kubernetes.io/name=minio

# Check MinIO logs
kubectl logs -n loki -l app.kubernetes.io/name=minio --tail=100

# Check disk space
kubectl exec -n loki <minio-pod> -- df -h /export

# Restart MinIO if needed
kubectl rollout restart statefulset -n loki loki-minio
```

#### Issue: Memory pressure causing drops
**Cause:** Insufficient memory for buffering  
**Fix:**
```bash
# Check memory usage
kubectl top pod -n loki -l app.kubernetes.io/component=write

# Increase write pod memory
kubectl edit helmrelease -n loki loki
# Update:
# write:
#   replicas: 2
#   resources:
#     limits:
#       memory: 2Gi
#     requests:
#       memory: 1Gi
```

### Step 3: Verify ingestion is working

```bash
# Send test logs
kubectl run test-logger --image=busybox --restart=Never -- sh -c "
  for i in {1..10}; do 
    echo \"Test log \$i at \$(date)\"
    sleep 1
  done
"

# Wait 10 seconds, then query
sleep 10
curl -G -s "http://localhost:3100/loki/api/v1/query" \
  --data-urlencode 'query={namespace="default",pod="test-logger"}' | jq .

# Cleanup
kubectl delete pod test-logger
```

## Verification

1. Check ingestion metrics:
```bash
# Port forward
kubectl port-forward -n loki svc/loki-gateway 3100:80

# Check ingestion rate
curl http://localhost:3100/metrics | grep loki_distributor_bytes_received_total

# Check for discarded samples
curl http://localhost:3100/metrics | grep loki_discarded_samples_total
```

2. Monitor error rate:
```bash
# Should see no or very few errors
kubectl logs -n loki -l app.kubernetes.io/component=write --tail=100 | grep -c "error"
```

3. Verify from Alloy:
```bash
# Check Alloy successfully sending logs
kubectl logs -n alloy -l app.kubernetes.io/name=alloy --tail=50 | grep -i "loki"
# Should see successful batches sent
```

4. Test end-to-end:
```bash
# Generate a known log
kubectl run test-marker --image=busybox --restart=Never -- sh -c "echo 'UNIQUE_TEST_MARKER_$(date +%s)'"

# Wait and query for it
sleep 15
kubectl port-forward -n loki svc/loki-gateway 3100:80
curl -G -s "http://localhost:3100/loki/api/v1/query" \
  --data-urlencode 'query={namespace="default"} |= "UNIQUE_TEST_MARKER"' | jq .

kubectl delete pod test-marker
```

## Prevention

1. **Configure appropriate rate limits**
   - Set ingestion_rate_mb based on actual load
   - Add 50% headroom for spikes
   - Monitor ingestion patterns

2. **Reduce label cardinality**
   - Avoid high-cardinality labels (IDs, timestamps)
   - Use <10 labels per stream
   - Index only queryable fields

3. **Configure proper retention**
   - Current: 744h (31 days)
   - Balance between cost and requirements
   - Implement tiered storage

4. **Monitor ingestion health**
   - Alert on ingestion error rate >5%
   - Track discarded samples
   - Monitor stream creation rate

5. **Implement buffering at source**
   - Configure Alloy with retry logic
   - Use persistent queue in Alloy
   - Handle backpressure gracefully

6. **Right-size resources**
   - Monitor actual resource usage
   - Scale write replicas based on load
   - Add memory headroom for spikes

## Best Practices for Labels

```yaml
# ✅ GOOD: Low cardinality labels
labels:
  namespace: "production"
  app: "api"
  environment: "prod"
  cluster: "homelab"

# ❌ BAD: High cardinality labels  
labels:
  pod_name: "api-xyz-123"  # Changes frequently
  request_id: "uuid-here"  # Unique per request
  timestamp: "1234567890"  # Always unique
  container_id: "docker://abc123"  # Unique per container
```

## Related Alerts

- `LokiDown`
- `LokiWritePathDown`
- `LokiHighMemory`
- `LokiStorageIssues`

## Escalation

If ingestion errors persist:
1. Review log source configurations (Alloy)
2. Analyze log label cardinality
3. Consider scaling write path
4. Review storage backend performance

## Additional Resources

- [Loki Ingestion Limits](https://grafana.com/docs/loki/latest/configuration/#limits_config)
- [Best Practices for Labels](https://grafana.com/docs/loki/latest/getting-started/labels/)
- [Troubleshooting Ingestion](https://grafana.com/docs/loki/latest/operations/troubleshooting/)
- [Alloy Configuration](https://grafana.com/docs/alloy/latest/)

