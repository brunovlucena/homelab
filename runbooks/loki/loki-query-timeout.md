# 🚨 Runbook: Loki Query Timeout

## Alert Information

**Alert Name:** `LokiQueryTimeout`  
**Severity:** Warning  
**Component:** Loki  
**Service:** Query Processing

## Symptom

Loki queries are timing out before completion. Users cannot retrieve logs for debugging.

## Impact

- **User Impact:** HIGH - Unable to access logs for troubleshooting
- **Business Impact:** MODERATE - Reduced operational efficiency
- **Data Impact:** NONE - No data loss, just access issues

## Diagnosis

### 1. Identify Timed-Out Queries

```bash
# Port forward to Loki
kubectl port-forward -n loki svc/loki-gateway 3100:80

# Check for timeout errors in logs
kubectl logs -n loki -l app.kubernetes.io/component=read --tail=200 | grep -i "timeout\|deadline exceeded"
```

### 2. Check Query Performance Metrics

```bash
# Check query duration metrics
curl http://localhost:3100/metrics | grep loki_query_duration_seconds

# Check for slow queries
curl http://localhost:3100/metrics | grep loki_logql_querystats_
```

### 3. Test Query Performance

```bash
# Try a simple query
time curl -G -s "http://localhost:3100/loki/api/v1/query" \
  --data-urlencode 'query={namespace="loki"}' \
  --data-urlencode 'limit=10'

# Try a range query (typically slower)
time curl -G -s "http://localhost:3100/loki/api/v1/query_range" \
  --data-urlencode 'query={namespace="loki"}' \
  --data-urlencode 'start=now-24h' \
  --data-urlencode 'end=now' \
  --data-urlencode 'limit=100'
```

### 4. Check Resource Constraints

```bash
# Check if read pods are resource-constrained
kubectl top pods -n loki -l app.kubernetes.io/component=read

# Check for throttling
kubectl describe pod -n loki -l app.kubernetes.io/component=read | grep -A 5 "Limits\|Requests"
```

### 5. Check Storage Backend Performance

```bash
# Test MinIO response time
kubectl exec -n loki <minio-pod> -- sh -c 'time wget -O /dev/null http://localhost:9000/minio/health/live'

# Check MinIO metrics
kubectl port-forward -n loki svc/loki-minio 9000:9000
curl http://localhost:9000/minio/v2/metrics/cluster
```

## Resolution Steps

### Step 1: Identify query pattern causing timeout

```bash
# Check recent queries in logs
kubectl logs -n loki -l app.kubernetes.io/component=read --tail=300 | grep "query=" | tail -20

# Look for patterns:
# - Large time ranges (>7 days)
# - Queries without label filters ({})
# - Regex-heavy queries
# - Queries on high-cardinality labels
```

### Step 2: Common Issues and Fixes

#### Issue: Query timeout configuration too low
**Cause:** Default timeout too short for valid queries  
**Fix:**
```bash
# Check current timeout
kubectl get helmrelease -n loki loki -o yaml | grep query_timeout

# Increase query timeout
kubectl edit helmrelease -n loki loki
# Add or update:
# loki:
#   limits_config:
#     query_timeout: 10m  # Increase from default 5m

# Wait for reconciliation
flux reconcile helmrelease loki -n loki
```

#### Issue: Expensive queries overwhelming system
**Cause:** Queries scanning too much data  
**Fix:**
```bash
# Set stricter query limits
kubectl edit helmrelease -n loki loki
# Add:
# loki:
#   limits_config:
#     max_query_length: 168h  # Limit to 7 days
#     max_query_lookback: 168h
#     max_entries_limit_per_query: 5000  # Limit results
#     max_streams_matcher_per_query: 1000
#     max_chunks_per_query: 1000000  # Reduce from 2M

# Educate users on efficient queries (see Query Optimization below)
```

#### Issue: Insufficient read replicas
**Cause:** Too many concurrent queries for available resources  
**Fix:**
```bash
# Scale up read replicas
kubectl scale deployment -n loki loki-read --replicas=3

# Or edit HelmRelease for permanent change
kubectl edit helmrelease -n loki loki
# Update:
# read:
#   replicas: 3

# Verify scaling
kubectl get pods -n loki -l app.kubernetes.io/component=read
```

#### Issue: Storage backend slow
**Cause:** MinIO disk I/O bottleneck  
**Fix:**
```bash
# Test disk performance
kubectl exec -n loki <minio-pod> -- sh -c 'dd if=/dev/zero of=/export/test bs=1M count=100 conv=fdatasync'

# Check if PVC is using slow storage class
kubectl get pvc -n loki -o yaml | grep storageClassName

# Consider migrating to faster storage or adding cache

# Enable query result caching
kubectl edit helmrelease -n loki loki
# Add:
# queryFrontend:
#   replicas: 1
# resultsCache:
#   enabled: true
#   backend: inmemory
#   inmemory:
#     max_size_mb: 500
```

#### Issue: Too many chunks to scan
**Cause:** Poor data indexing or large time range  
**Fix:**
```bash
# Check chunk count per query
curl http://localhost:3100/metrics | grep loki_chunk_store_chunks_queried

# Improve indexing (requires re-ingestion)
# Ensure using TSDB index (already configured)
kubectl get helmrelease -n loki loki -o yaml | grep -A 5 "schemaConfig"

# Enable bloom filters for faster filtering
kubectl edit helmrelease -n loki loki
# Add:
# loki:
#   bloomBuild:
#     enabled: true
#   bloomGateway:
#     enabled: true
```

#### Issue: Memory pressure causing slow queries
**Cause:** Insufficient memory for query processing  
**Fix:**
```bash
# Check memory usage
kubectl top pod -n loki -l app.kubernetes.io/component=read

# Increase memory limits
kubectl edit helmrelease -n loki loki
# Update:
# read:
#   resources:
#     limits:
#       memory: 3Gi
#     requests:
#       memory: 2Gi
```

#### Issue: Network latency between components
**Cause:** Communication delays in distributed setup  
**Fix:**
```bash
# Test inter-pod connectivity
kubectl exec -it -n loki <loki-read-pod> -- sh -c 'time nc -zv loki-backend 3100'

# Check if pods are on same node (reduces latency)
kubectl get pods -n loki -o wide

# Consider pod affinity rules for co-location
kubectl edit helmrelease -n loki loki
# Add under read:
# affinity:
#   podAntiAffinity:
#     preferredDuringSchedulingIgnoredDuringExecution:
#     - weight: 100
#       podAffinityTerm:
#         topologyKey: kubernetes.io/hostname
```

### Step 3: Implement query optimization

```bash
# Enable query statistics for monitoring
kubectl edit helmrelease -n loki loki
# Add:
# loki:
#   querier:
#     query_stats_enabled: true
```

## Verification

1. Test query performance:
```bash
# Port forward
kubectl port-forward -n loki svc/loki-gateway 3100:80

# Test simple query
time curl -G -s "http://localhost:3100/loki/api/v1/query" \
  --data-urlencode 'query={namespace="loki"}' \
  --data-urlencode 'limit=10'
# Should complete in <5s

# Test range query
time curl -G -s "http://localhost:3100/loki/api/v1/query_range" \
  --data-urlencode 'query={namespace="loki"}' \
  --data-urlencode 'start=now-1h' \
  --data-urlencode 'end=now'
# Should complete in <30s
```

2. Verify no timeout errors:
```bash
kubectl logs -n loki -l app.kubernetes.io/component=read --tail=100 | grep -c "timeout"
# Should be 0 or very low
```

3. Check query metrics:
```bash
# Check P99 query latency
curl http://localhost:3100/metrics | grep loki_query_duration_seconds | grep quantile
```

4. Test from Grafana:
```bash
# Access Grafana Explore
# Run test queries with various time ranges
# Verify completion times are acceptable
```

## Query Optimization Guide

### Efficient Query Examples

```logql
# ✅ EXCELLENT: Specific labels + line filter + small time range
{namespace="production", app="api"} |= "error" [5m]

# ✅ GOOD: Specific labels + structured filter
{namespace="production"} | json | level="error" [1h]

# ✅ GOOD: Metric aggregation with specific labels
rate({namespace="production", app="api"}[5m])

# ⚠️ ACCEPTABLE: Broader labels but filtered
{namespace="production"} |= "timeout" [1h]

# ❌ BAD: No label filters (scans everything!)
{} |= "error"

# ❌ BAD: Regex on labels (slow!)
{namespace=~"prod.*"}

# ❌ BAD: Very large time range
{namespace="production"}[30d]

# ❌ BAD: Complex regex in line filter
{namespace="production"} |~ "error.*timeout.*database"
```

### Query Best Practices

1. **Always use label selectors** - Narrow down to specific services
2. **Keep time ranges small** - Start with 1h, expand if needed
3. **Use line filters early** - Filter before parsing
4. **Limit results** - Use `limit` parameter or `| limit 100`
5. **Use metric queries for aggregations** - More efficient than log queries
6. **Avoid regex when possible** - Use simple substring match `|=` instead of `|~`

## Prevention

1. **Set appropriate query limits**
   - Current: 5m timeout (consider increasing to 10m)
   - Limit max query length to 7 days
   - Limit results per query

2. **Scale read path appropriately**
   - Current: 2 read replicas
   - Consider 3+ for high query load
   - Use HPA for automatic scaling

3. **Implement caching**
   - Enable query frontend
   - Cache query results
   - Use split-by-interval

4. **Monitor query patterns**
   - Track slow queries
   - Alert on timeout rate >5%
   - Educate users on efficient queries

5. **Optimize storage**
   - Use fast storage for MinIO
   - Enable bloom filters
   - Implement tiered storage

6. **Regular maintenance**
   - Compact old chunks
   - Monitor chunk count
   - Adjust retention policies

## Related Alerts

- `LokiReadPathSlow`
- `LokiDown`
- `LokiHighMemory`
- `LokiStorageIssues`

## Escalation

If query timeouts persist:
1. Review overall system capacity
2. Consider implementing query frontend with caching
3. Evaluate data retention and volume
4. Consider architectural changes (query sharding)

## Additional Resources

- [Loki Query Performance](https://grafana.com/docs/loki/latest/operations/query-performance/)
- [LogQL Query Optimization](https://grafana.com/docs/loki/latest/logql/query_examples/)
- [Loki Query Limits](https://grafana.com/docs/loki/latest/configuration/#limits_config)
- [Query Frontend](https://grafana.com/docs/loki/latest/fundamentals/architecture/components/#query-frontend)

