# 🚨 Runbook: Loki Read Path Slow

## Alert Information

**Alert Name:** `LokiReadPathSlow`  
**Severity:** Warning  
**Component:** Loki  
**Service:** Log Queries

## Symptom

Loki queries are taking longer than expected. Dashboard and log searches are slow.

## Impact

- **User Impact:** MODERATE - Slow troubleshooting and debugging
- **Business Impact:** MODERATE - Reduced operational efficiency
- **Data Impact:** NONE - No data loss

## Diagnosis

### 1. Check Read Pod Status

```bash
kubectl get pods -n loki -l app.kubernetes.io/component=read
kubectl top pods -n loki -l app.kubernetes.io/component=read
```

### 2. Check Query Latency Metrics

```bash
# Port forward to Loki
kubectl port-forward -n loki svc/loki-gateway 3100:80

# Check query metrics
curl http://localhost:3100/metrics | grep -i "loki_query\|loki_querier"
```

### 3. Check Read Pod Logs

```bash
kubectl logs -n loki -l app.kubernetes.io/component=read --tail=100 | grep -i "slow\|timeout\|error"
```

### 4. Test Query Performance

```bash
# Simple query
time curl -G -s "http://localhost:3100/loki/api/v1/query" \
  --data-urlencode 'query={namespace="loki"}'

# Range query (more expensive)
time curl -G -s "http://localhost:3100/loki/api/v1/query_range" \
  --data-urlencode 'query={namespace="loki"}' \
  --data-urlencode 'start=now-1h' \
  --data-urlencode 'end=now'
```

### 5. Check Storage Backend Performance

```bash
# Check MinIO response time
kubectl exec -n loki <minio-pod> -- sh -c 'time wget -O /dev/null http://localhost:9000/minio/health/live'

# Check MinIO logs for slow operations
kubectl logs -n loki -l app.kubernetes.io/name=minio --tail=100 | grep -i "slow"
```

## Resolution Steps

### Step 1: Identify query patterns

```bash
# Check recent queries in logs
kubectl logs -n loki -l app.kubernetes.io/component=read --tail=100 | grep -i "query"

# Look for expensive queries
kubectl logs -n loki -l app.kubernetes.io/component=read --tail=200 | grep -i "duration\|took"
```

### Step 2: Common Issues and Fixes

#### Issue: Large time range queries
**Cause:** Users querying too much data at once  
**Fix:**
```bash
# Check query limits
kubectl get helmrelease -n loki loki -o yaml | grep -A 10 "limits_config"

# Adjust query limits to prevent abuse
kubectl edit helmrelease -n loki loki
# Add or update:
# loki:
#   limits_config:
#     max_query_length: 721h  # 30 days
#     max_query_lookback: 744h  # 31 days
#     max_streams_per_user: 0  # unlimited
#     max_chunks_per_query: 2000000
```

#### Issue: Too many concurrent queries
**Cause:** Insufficient read replicas  
**Fix:**
```bash
# Scale up read replicas
kubectl scale deployment -n loki loki-read --replicas=3

# Monitor performance after scaling
kubectl top pods -n loki -l app.kubernetes.io/component=read
```

#### Issue: Storage backend slow
**Cause:** MinIO disk I/O bottleneck  
**Fix:**
```bash
# Check MinIO disk performance
kubectl exec -n loki <minio-pod> -- sh -c 'dd if=/dev/zero of=/export/testfile bs=1M count=100 oflag=direct'

# Check PVC performance class
kubectl get pvc -n loki -o yaml | grep storageClassName

# Consider migrating to faster storage class or adding caching
```

#### Issue: High memory usage causing GC pauses
**Cause:** Insufficient memory for query workload  
**Fix:**
```bash
# Check memory usage
kubectl top pod -n loki -l app.kubernetes.io/component=read

# Increase memory limits
kubectl edit helmrelease -n loki loki
# Update: read.resources.limits.memory: 2Gi

# Or enable query frontend caching
kubectl edit helmrelease -n loki loki
# Add:
# queryFrontend:
#   replicas: 1
#   cacheResults: true
```

#### Issue: Uncompressed or poorly indexed data
**Cause:** Data not properly chunked or indexed  
**Fix:**
```bash
# Verify schema configuration
kubectl get helmrelease -n loki loki -o yaml | grep -A 20 "schemaConfig"

# Ensure TSDB index is being used (more efficient)
# Current config uses: store: tsdb, schema: v13

# Consider adding bloom filters for better query performance
kubectl edit helmrelease -n loki loki
# Add:
# loki:
#   bloomBuild:
#     enabled: true
```

### Step 3: Optimize query patterns

```bash
# Educate users on efficient queries:
# GOOD: {namespace="loki"} |= "error" (specific namespace + filter)
# BAD: {} (queries all logs)
# BAD: {namespace=~".+"} (regex on label)

# Add query timeout to prevent runaway queries
kubectl edit helmrelease -n loki loki
# Add:
# loki:
#   limits_config:
#     query_timeout: 5m
```

### Step 4: Clear cache if corrupted

```bash
# Restart read pods to clear in-memory cache
kubectl rollout restart deployment -n loki loki-read
kubectl rollout status deployment -n loki loki-read
```

## Verification

1. Test query performance:
```bash
# Port forward
kubectl port-forward -n loki svc/loki-gateway 3100:80

# Run test query and measure time
time curl -G -s "http://localhost:3100/loki/api/v1/query" \
  --data-urlencode 'query={namespace="loki"}' \
  --data-urlencode 'limit=100' | jq .
```

2. Check query latency metrics:
```bash
curl http://localhost:3100/metrics | grep loki_query_duration_seconds
```

3. Verify from Grafana:
```bash
# Navigate to Grafana Explore
# Run a LogQL query: {namespace="loki"} | json
# Check query execution time in bottom right
```

4. Monitor resource usage:
```bash
kubectl top pods -n loki -l app.kubernetes.io/component=read
```

## Prevention

1. **Set appropriate query limits**
   - Limit time range per query
   - Limit number of streams returned
   - Set query timeout thresholds

2. **Scale read path appropriately**
   - Current: 2 read replicas
   - Consider 3+ for high query volume
   - Use HPA based on CPU/memory

3. **Implement caching**
   - Enable query result caching
   - Use query frontend component
   - Configure cache TTL appropriately

4. **Optimize storage**
   - Use fast storage class for MinIO
   - Implement tiered storage (hot/cold)
   - Enable compression

5. **Monitor query patterns**
   - Track slow queries
   - Alert on expensive query patterns
   - Educate users on efficient queries

6. **Regular maintenance**
   - Compact old chunks
   - Clean up deleted streams
   - Review and adjust retention policies

## Query Optimization Tips

### Efficient Query Examples
```logql
# ✅ GOOD: Specific label selectors
{namespace="production", app="api"} |= "error"

# ✅ GOOD: Use line filters early
{namespace="production"} |= "error" | json | level="error"

# ✅ GOOD: Use metric queries for aggregations
rate({namespace="production"}[5m])

# ❌ BAD: No label selectors
{} |= "error"

# ❌ BAD: Regex on labels
{namespace=~"prod.*"}

# ❌ BAD: Large time ranges without filters
{namespace="production"}[7d]
```

## Related Alerts

- `LokiDown`
- `LokiQueryTimeout`
- `LokiHighMemory`
- `LokiStorageIssues`

## Escalation

If query performance issues persist:
1. Review query patterns and user behavior
2. Consider implementing query frontend with caching
3. Evaluate storage backend performance
4. Consider upgrading to Loki 3.x for better performance

## Additional Resources

- [Loki Query Performance](https://grafana.com/docs/loki/latest/operations/query-performance/)
- [LogQL Query Optimization](https://grafana.com/docs/loki/latest/logql/query_examples/)
- [Loki Query Limits](https://grafana.com/docs/loki/latest/configuration/#limits_config)
- [Loki Caching](https://grafana.com/docs/loki/latest/operations/caching/)

