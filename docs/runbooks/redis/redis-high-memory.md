# 🚨 Runbook: Redis High Memory Usage

## Alert Information

**Alert Name:** `RedisHighMemory`  
**Severity:** Warning  
**Component:** redis  
**Service:** redis-master  
**Threshold:** Memory usage > 80% of maxmemory

## Symptom

Redis memory usage is approaching or has exceeded configured limits, potentially causing evictions or OOM kills.

## Impact

- **User Impact:** MEDIUM - Potential cache misses, session loss, degraded performance
- **Business Impact:** MEDIUM - Increased latency, possible service degradation
- **Data Impact:** HIGH - Risk of data eviction or loss

## Diagnosis

### 1. Check Current Memory Usage

```bash
# Get memory info
kubectl exec -n redis redis-master-0 -- redis-cli info memory

# Key metrics to check:
# - used_memory_human: Total memory used
# - used_memory_rss_human: RSS memory (actual RAM)
# - maxmemory_human: Maximum memory limit
# - mem_fragmentation_ratio: Fragmentation ratio
# - evicted_keys: Number of keys evicted
```

### 2. Check Memory Statistics

```bash
# Detailed memory stats
kubectl exec -n redis redis-master-0 -- redis-cli memory stats

# Memory usage by key pattern
kubectl exec -n redis redis-master-0 -- redis-cli --bigkeys

# Sample memory usage
kubectl exec -n redis redis-master-0 -- redis-cli memory doctor
```

### 3. Check Eviction Policy

```bash
kubectl exec -n redis redis-master-0 -- redis-cli config get maxmemory-policy
kubectl exec -n redis redis-master-0 -- redis-cli config get maxmemory
```

### 4. Check Key Distribution

```bash
# Count keys
kubectl exec -n redis redis-master-0 -- redis-cli dbsize

# Sample keys by pattern
kubectl exec -n redis redis-master-0 -- redis-cli --scan --pattern "*" | head -20

# Check keys by namespace
kubectl exec -n redis redis-master-0 -- redis-cli --scan --pattern "bruno:*" | wc -l
kubectl exec -n redis redis-master-0 -- redis-cli --scan --pattern "homepage:*" | wc -l
```

### 5. Analyze Memory Usage by Type

```bash
# Memory usage breakdown
kubectl exec -n redis redis-master-0 -- redis-cli memory usage <key>

# Example for session keys
for key in $(kubectl exec -n redis redis-master-0 -- redis-cli --scan --pattern "bruno:session:*" | head -10); do
  echo "Key: $key"
  kubectl exec -n redis redis-master-0 -- redis-cli memory usage $key
done
```

### 6. Check Pod Resource Limits

```bash
# Check configured limits
kubectl get pod -n redis redis-master-0 -o jsonpath='{.spec.containers[0].resources}'

# Check actual usage
kubectl top pod -n redis redis-master-0
```

## Resolution Steps

### Step 1: Immediate Actions

#### Quick Win: Flush Expired Keys

```bash
# Force active expire (careful in production!)
kubectl exec -n redis redis-master-0 -- redis-cli --scan --pattern "*" | \
  xargs -L 1 kubectl exec -n redis redis-master-0 -- redis-cli ttl | \
  grep -E '^-1$' -B 1
```

#### Remove Old/Unused Keys

```bash
# Identify keys without TTL
kubectl exec -n redis redis-master-0 -- redis-cli --scan | \
  while read key; do
    ttl=$(kubectl exec -n redis redis-master-0 -- redis-cli ttl "$key")
    if [ "$ttl" -eq -1 ]; then
      echo "No TTL: $key"
    fi
  done | head -20

# Add TTL to keys missing it (example: session keys)
kubectl exec -n redis redis-master-0 -- redis-cli --scan --pattern "bruno:session:*" | \
  while read key; do
    kubectl exec -n redis redis-master-0 -- redis-cli expire "$key" 86400
  done
```

### Step 2: Optimize Memory Usage

#### Issue: Large Keys Consuming Memory
**Cause:** Individual keys storing too much data  
**Fix:**
```bash
# Find biggest keys
kubectl exec -n redis redis-master-0 -- redis-cli --bigkeys

# Analyze specific large keys
kubectl exec -n redis redis-master-0 -- redis-cli memory usage <large-key>

# Check if key can be split or compressed
kubectl exec -n redis redis-master-0 -- redis-cli debug object <large-key>

# Consider application changes to:
# 1. Split large keys into smaller chunks
# 2. Use hash structures instead of strings
# 3. Implement compression
# 4. Set appropriate TTLs
```

#### Issue: Too Many Small Keys
**Cause:** Excessive number of keys with poor TTL management  
**Fix:**
```bash
# Count total keys
kubectl exec -n redis redis-master-0 -- redis-cli dbsize

# Analyze key patterns
kubectl exec -n redis redis-master-0 -- redis-cli --scan --pattern "*:*" | \
  cut -d: -f1 | sort | uniq -c | sort -rn | head -10

# Set appropriate TTLs for each namespace
# Example: Set 1-hour TTL for cache keys
kubectl exec -n redis redis-master-0 -- redis-cli --scan --pattern "cache:*" | \
  while read key; do
    kubectl exec -n redis redis-master-0 -- redis-cli expire "$key" 3600
  done
```

#### Issue: Memory Fragmentation
**Cause:** High mem_fragmentation_ratio (> 1.5)  
**Fix:**
```bash
# Check fragmentation ratio
kubectl exec -n redis redis-master-0 -- redis-cli info memory | grep mem_fragmentation_ratio

# If > 1.5, consider restarting Redis to defragment
# But first, ensure persistence is enabled!

# Check persistence
kubectl exec -n redis redis-master-0 -- redis-cli config get save
kubectl exec -n redis redis-master-0 -- redis-cli config get appendonly

# Force save before restart
kubectl exec -n redis redis-master-0 -- redis-cli bgsave

# Restart to defragment
kubectl delete pod -n redis redis-master-0

# Or enable active defragmentation (Redis 4.0+)
kubectl exec -n redis redis-master-0 -- redis-cli config set activedefrag yes
```

### Step 3: Adjust Configuration

#### Update Eviction Policy

```bash
# Check current policy
kubectl exec -n redis redis-master-0 -- redis-cli config get maxmemory-policy

# Set appropriate policy based on use case
# - allkeys-lru: Evict any key, LRU (good for cache)
# - volatile-lru: Evict keys with TTL, LRU (good for mixed workload)
# - allkeys-lfu: Evict any key, LFU (best for most cache scenarios)
# - volatile-ttl: Evict keys with shortest TTL

kubectl exec -n redis redis-master-0 -- redis-cli config set maxmemory-policy allkeys-lfu

# Make permanent by updating HelmRelease
kubectl edit helmrelease redis -n redis
# Add:
#   master:
#     configuration: |
#       maxmemory-policy allkeys-lfu
```

#### Increase maxmemory

```bash
# Check current maxmemory
kubectl exec -n redis redis-master-0 -- redis-cli config get maxmemory

# Increase maxmemory (temporary)
kubectl exec -n redis redis-master-0 -- redis-cli config set maxmemory 512mb

# Make permanent in HelmRelease
kubectl edit helmrelease redis -n redis
# Update:
#   master:
#     configuration: |
#       maxmemory 512mb
```

### Step 4: Increase Pod Resources

```bash
# Edit HelmRelease to increase memory limits
kubectl edit helmrelease redis -n redis

# Update resources section:
#   master:
#     resources:
#       limits:
#         memory: "1Gi"  # Increased from 512Mi
#       requests:
#         memory: "512Mi"  # Increased from 256Mi

# Apply changes
flux reconcile helmrelease redis -n redis

# Monitor restart
kubectl rollout status statefulset redis-master -n redis
```

### Step 5: Clean Up Specific Patterns

#### Remove Old Sessions

```bash
# Find old session keys (example: agent-bruno)
kubectl exec -n redis redis-master-0 -- redis-cli --scan --pattern "bruno:session:*"

# Delete specific old sessions
kubectl exec -n redis redis-master-0 -- redis-cli del "bruno:session:old-ip"

# Or set shorter TTL on all sessions
kubectl exec -n redis redis-master-0 -- redis-cli --scan --pattern "bruno:session:*" | \
  while read key; do
    kubectl exec -n redis redis-master-0 -- redis-cli expire "$key" 43200  # 12 hours
  done
```

#### Remove Stale Cache

```bash
# Remove cache older than certain time
kubectl exec -n redis redis-master-0 -- redis-cli --scan --pattern "cache:*" | \
  while read key; do
    ttl=$(kubectl exec -n redis redis-master-0 -- redis-cli ttl "$key")
    if [ "$ttl" -eq -1 ] || [ "$ttl" -gt 3600 ]; then
      kubectl exec -n redis redis-master-0 -- redis-cli del "$key"
    fi
  done
```

## Verification

### 1. Check Memory Usage Decreased

```bash
kubectl exec -n redis redis-master-0 -- redis-cli info memory | grep used_memory_human
kubectl exec -n redis redis-master-0 -- redis-cli info memory | grep maxmemory_human
```

### 2. Verify Eviction Policy

```bash
kubectl exec -n redis redis-master-0 -- redis-cli config get maxmemory-policy
kubectl exec -n redis redis-master-0 -- redis-cli info stats | grep evicted_keys
```

### 3. Monitor Key Count

```bash
kubectl exec -n redis redis-master-0 -- redis-cli dbsize
```

### 4. Check Application Performance

```bash
# Test from client applications
kubectl exec -n bruno deployment/agent-bruno -- python3 -c "
import redis
r = redis.Redis(host='redis-master.redis.svc.cluster.local', port=6379)
print('PING:', r.ping())
print('Memory:', r.info('memory')['used_memory_human'])
"
```

### 5. Verify No OOM Kills

```bash
kubectl get pod -n redis redis-master-0 -o jsonpath='{.status.containerStatuses[0].lastState}'
kubectl describe pod -n redis redis-master-0 | grep -i "oom"
```

## Prevention

### 1. Implement TTL Strategy

All keys should have appropriate TTLs:

```python
# In application code
# Sessions: 24 hours
redis.setex('bruno:session:user1', 86400, data)

# Cache: 1 hour
redis.setex('cache:api:result', 3600, data)

# Temporary data: 5 minutes
redis.setex('temp:job:123', 300, data)
```

### 2. Configure Proper Resources

```yaml
# Redis HelmRelease values
master:
  resources:
    limits:
      memory: "1Gi"
      cpu: "500m"
    requests:
      memory: "512Mi"
      cpu: "100m"
  
  configuration: |
    maxmemory 800mb  # Leave 20% headroom
    maxmemory-policy allkeys-lfu
```

### 3. Set Up Monitoring

```yaml
# Prometheus rules for Redis memory
- alert: RedisHighMemory
  expr: |
    redis_memory_used_bytes / redis_memory_max_bytes > 0.8
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Redis memory usage high"

- alert: RedisMemoryCritical
  expr: |
    redis_memory_used_bytes / redis_memory_max_bytes > 0.95
  for: 2m
  labels:
    severity: critical
  annotations:
    summary: "Redis memory usage critical"
```

### 4. Regular Cleanup Jobs

```yaml
# Kubernetes CronJob for Redis cleanup
apiVersion: batch/v1
kind: CronJob
metadata:
  name: redis-cleanup
  namespace: redis
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: cleanup
            image: redis:7.2
            command:
            - /bin/sh
            - -c
            - |
              # Connect to Redis and cleanup
              redis-cli -h redis-master scan --pattern "temp:*" | xargs redis-cli del
              redis-cli -h redis-master bgsave
          restartPolicy: OnFailure
```

### 5. Application Best Practices

- ✅ Always set TTL on keys
- ✅ Use appropriate data structures (hashes for objects)
- ✅ Implement key compression for large values
- ✅ Use connection pooling
- ✅ Monitor application Redis usage
- ✅ Implement cache warming strategies
- ✅ Use Redis namespaces for different data types

### 6. Key Naming Convention

```
<app>:<type>:<identifier>
Examples:
- bruno:session:192.168.1.1
- homepage:cache:projects:list
- api:rate-limit:user:123
```

### 7. Enable Persistence

```yaml
master:
  persistence:
    enabled: true
    size: 8Gi
  
  configuration: |
    # RDB snapshots
    save 900 1      # Save after 900s if 1 key changed
    save 300 10     # Save after 300s if 10 keys changed
    save 60 10000   # Save after 60s if 10000 keys changed
    
    # AOF (more durable but slower)
    appendonly yes
    appendfsync everysec
```

## Performance Tips

1. **Use Pipelining:** Batch multiple commands together
2. **Use Hashes:** More memory efficient than individual keys
3. **Avoid Large Keys:** Split into smaller chunks
4. **Use Lazy Expiration:** Redis will lazily delete expired keys
5. **Monitor Fragmentation:** Restart periodically if high fragmentation
6. **Use Memory-Efficient Commands:** SCAN instead of KEYS

## Memory Analysis Script

```bash
#!/bin/bash
# redis-memory-analysis.sh

NAMESPACE="redis"
POD="redis-master-0"

echo "=== Redis Memory Analysis ==="
echo ""

# Basic memory info
echo "1. Memory Overview:"
kubectl exec -n $NAMESPACE $POD -- redis-cli info memory | grep -E "used_memory_human|maxmemory_human|mem_fragmentation_ratio|evicted_keys"

echo ""
echo "2. Key Count:"
kubectl exec -n $NAMESPACE $POD -- redis-cli dbsize

echo ""
echo "3. Biggest Keys:"
kubectl exec -n $NAMESPACE $POD -- redis-cli --bigkeys | tail -20

echo ""
echo "4. Key Patterns:"
kubectl exec -n $NAMESPACE $POD -- redis-cli --scan --pattern "*:*" | \
  cut -d: -f1-2 | sort | uniq -c | sort -rn | head -10

echo ""
echo "5. Keys Without TTL:"
kubectl exec -n $NAMESPACE $POD -- redis-cli --scan | \
  while read key; do
    ttl=$(kubectl exec -n $NAMESPACE $POD -- redis-cli ttl "$key")
    if [ "$ttl" -eq -1 ]; then
      echo "No TTL: $key"
    fi
  done | head -10

echo ""
echo "6. Eviction Stats:"
kubectl exec -n $NAMESPACE $POD -- redis-cli info stats | grep evicted
```

## Related Alerts

- `RedisDown`
- `RedisSlowOperations`
- `RedisEvictionRate`
- `RedisFragmentation`
- `RedisPodOOMKilled`

## Escalation

If memory issues persist after applying fixes:

1. ✅ Review application code for memory leaks
2. 📊 Analyze key growth patterns over time
3. 🔍 Check for keys without proper TTL
4. 💾 Consider scaling horizontally (Redis Cluster)
5. 🔄 Evaluate Redis Sentinel for HA
6. 📞 Contact development team for application optimization
7. 🆘 Consider migrating to Redis Cluster for partitioning

## Additional Resources

- [Redis Memory Optimization](https://redis.io/docs/management/optimization/memory-optimization/)
- [Redis Eviction Policies](https://redis.io/docs/reference/eviction/)
- [Redis Memory Analysis Tools](https://redis.io/docs/management/optimization/)
- [Redis Best Practices](https://redis.io/docs/manual/patterns/)

---

**Last Updated:** 2025-10-15  
**Version:** 1.0  
**Owner:** Infrastructure Team

