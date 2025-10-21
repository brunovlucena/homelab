# 🚨 Runbook: Redis Slow Operations

## Alert Information

**Alert Name:** `RedisSlowOperations`  
**Severity:** Warning  
**Component:** redis  
**Service:** redis-master  
**Threshold:** Command latency > 100ms

## Symptom

Redis commands are taking longer than expected to execute, causing application slowdowns and timeouts.

## Impact

- **User Impact:** HIGH - Slow page loads, timeouts, poor user experience
- **Business Impact:** MEDIUM - Degraded service performance, potential revenue impact
- **Data Impact:** LOW - Data integrity maintained but access is slow

## Diagnosis

### 1. Check Redis Latency

```bash
# Real-time latency monitoring (Ctrl+C to stop)
kubectl exec -n redis redis-master-0 -- redis-cli --latency

# Latency history
kubectl exec -n redis redis-master-0 -- redis-cli --latency-history

# Latency doctor diagnosis
kubectl exec -n redis redis-master-0 -- redis-cli --latency-doctor
```

### 2. Check Slow Log

```bash
# Get slow log entries (default threshold: 10ms)
kubectl exec -n redis redis-master-0 -- redis-cli slowlog get 20

# Check slow log configuration
kubectl exec -n redis redis-master-0 -- redis-cli config get slowlog-log-slower-than
kubectl exec -n redis redis-master-0 -- redis-cli config get slowlog-max-len
```

### 3. Monitor Real-Time Commands

```bash
# Watch commands in real-time (Ctrl+C to stop)
kubectl exec -n redis redis-master-0 -- redis-cli monitor
```

### 4. Check Redis Performance Stats

```bash
# Get comprehensive stats
kubectl exec -n redis redis-master-0 -- redis-cli info stats

# Key metrics to check:
# - instantaneous_ops_per_sec: Current ops/sec
# - total_commands_processed: Total commands
# - keyspace_hits/keyspace_misses: Cache hit ratio
# - evicted_keys: Number of evictions
# - blocked_clients: Clients waiting on blocking ops
```

### 5. Check Resource Usage

```bash
# Check CPU and memory
kubectl top pod -n redis redis-master-0

# Check disk I/O (if persistence enabled)
kubectl exec -n redis redis-master-0 -- iostat -x 1 5

# Check network
kubectl exec -n redis redis-master-0 -- netstat -s | grep -i tcp
```

### 6. Check Connected Clients

```bash
# List all connected clients
kubectl exec -n redis redis-master-0 -- redis-cli client list

# Count clients
kubectl exec -n redis redis-master-0 -- redis-cli client list | wc -l

# Check blocked clients
kubectl exec -n redis redis-master-0 -- redis-cli info clients | grep blocked_clients
```

### 7. Analyze Command Frequency

```bash
# Get command stats
kubectl exec -n redis redis-master-0 -- redis-cli info commandstats

# This shows:
# - calls: Number of times called
# - usec: Total time spent
# - usec_per_call: Average time per call
```

## Resolution Steps

### Step 1: Identify Slow Commands

#### Analyze Slow Log

```bash
# Get detailed slow log
kubectl exec -n redis redis-master-0 -- redis-cli slowlog get 50 | \
  grep -E "^\d+\)|\d+\) \"" | \
  paste -d " " - - - - - - | \
  awk '{print $4, $6, $8, $10, $12}' | \
  sort | uniq -c | sort -rn
```

#### Find Problematic Patterns

```bash
# Common slow operations:
# 1. KEYS * (scans entire keyspace)
# 2. Large list operations (LRANGE 0 -1)
# 3. Large set operations (SMEMBERS on huge sets)
# 4. SCAN with too many iterations
# 5. Large DEL operations

# Check for these in slow log
kubectl exec -n redis redis-master-0 -- redis-cli slowlog get 100 | grep -i "keys\|lrange\|smembers"
```

### Step 2: Fix Common Issues

#### Issue: KEYS Command Being Used
**Cause:** Application using KEYS instead of SCAN  
**Fix:**
```bash
# Identify clients using KEYS
kubectl exec -n redis redis-master-0 -- redis-cli client list | grep -i keys

# KEYS is O(N) - blocks Redis!
# Applications should use SCAN instead

# Example fix in application code:
# ❌ Bad:
# keys = redis.keys('pattern:*')

# ✅ Good:
# for key in redis.scan_iter('pattern:*'):
#     process(key)

# Temporarily rename KEYS command to prevent usage
kubectl exec -n redis redis-master-0 -- redis-cli config set rename-command KEYS ""
```

#### Issue: Large Collection Operations
**Cause:** Operating on large lists/sets/hashes  
**Fix:**
```bash
# Find large keys
kubectl exec -n redis redis-master-0 -- redis-cli --bigkeys

# Check specific large key
kubectl exec -n redis redis-master-0 -- redis-cli llen "large:list:key"
kubectl exec -n redis redis-master-0 -- redis-cli scard "large:set:key"

# Solutions:
# 1. Split large collections into smaller chunks
# 2. Use pagination (LRANGE start stop instead of 0 -1)
# 3. Use SSCAN instead of SMEMBERS for large sets
# 4. Implement lazy loading in application
```

#### Issue: High Memory Usage Causing Swapping
**Cause:** Redis swapping to disk due to memory pressure  
**Fix:**
```bash
# Check if swapping is happening
kubectl exec -n redis redis-master-0 -- redis-cli info stats | grep swap

# Check memory
kubectl exec -n redis redis-master-0 -- redis-cli info memory

# If memory high, see redis-high-memory.md runbook
# Quick fixes:
# 1. Delete unnecessary keys
# 2. Set eviction policy
# 3. Increase memory limits
```

#### Issue: Persistence Causing Delays
**Cause:** BGSAVE or AOF rewrite blocking operations  
**Fix:**
```bash
# Check if background save is running
kubectl exec -n redis redis-master-0 -- redis-cli info persistence | grep rdb_bgsave_in_progress
kubectl exec -n redis redis-master-0 -- redis-cli info persistence | grep aof_rewrite_in_progress

# Check last save time
kubectl exec -n redis redis-master-0 -- redis-cli lastsave

# Optimize persistence configuration
kubectl exec -n redis redis-master-0 -- redis-cli config get save
kubectl exec -n redis redis-master-0 -- redis-cli config get appendonly

# Reduce save frequency if needed
kubectl exec -n redis redis-master-0 -- redis-cli config set save "900 1 300 10"

# Or disable persistence temporarily (careful!)
kubectl exec -n redis redis-master-0 -- redis-cli config set save ""
```

#### Issue: Too Many Connections
**Cause:** Connection pool exhaustion or too many clients  
**Fix:**
```bash
# Check connection count
kubectl exec -n redis redis-master-0 -- redis-cli info clients

# Get max connections
kubectl exec -n redis redis-master-0 -- redis-cli config get maxclients

# Increase if needed
kubectl exec -n redis redis-master-0 -- redis-cli config set maxclients 10000

# Find clients with many connections
kubectl exec -n redis redis-master-0 -- redis-cli client list | \
  awk '{print $2}' | cut -d= -f2 | cut -d: -f1 | sort | uniq -c | sort -rn | head -10

# Kill idle clients (careful!)
kubectl exec -n redis redis-master-0 -- redis-cli client list | \
  grep "idle=300" | \
  awk '{print $2}' | cut -d= -f2 | \
  xargs -I {} kubectl exec -n redis redis-master-0 -- redis-cli client kill {}
```

#### Issue: Network Latency
**Cause:** High network latency between clients and Redis  
**Fix:**
```bash
# Test network latency from client pods
kubectl exec -n bruno deployment/agent-bruno -- sh -c 'time nc -zv redis-master.redis.svc.cluster.local 6379'

# Check if Redis and clients are on same node
kubectl get pod -n redis redis-master-0 -o jsonpath='{.spec.nodeName}'
kubectl get pod -n bruno -l app=agent-bruno -o jsonpath='{.items[0].spec.nodeName}'

# Use pipelining in application code to reduce round trips
# Check client timeout settings
```

#### Issue: CPU Throttling
**Cause:** Redis pod being CPU throttled  
**Fix:**
```bash
# Check CPU usage
kubectl top pod -n redis redis-master-0

# Check CPU limits
kubectl get pod -n redis redis-master-0 -o jsonpath='{.spec.containers[0].resources}'

# Increase CPU limits
kubectl edit helmrelease redis -n redis
# Update:
#   master:
#     resources:
#       limits:
#         cpu: "1000m"  # Increased from 500m
#       requests:
#         cpu: "200m"
```

### Step 3: Optimize Configuration

```bash
# Enable latency monitoring
kubectl exec -n redis redis-master-0 -- redis-cli config set latency-monitor-threshold 100

# Optimize slow log
kubectl exec -n redis redis-master-0 -- redis-cli config set slowlog-log-slower-than 10000  # 10ms
kubectl exec -n redis redis-master-0 -- redis-cli config set slowlog-max-len 500

# Optimize TCP settings
kubectl exec -n redis redis-master-0 -- redis-cli config set tcp-backlog 511
kubectl exec -n redis redis-master-0 -- redis-cli config set tcp-keepalive 300

# Disable problematic commands
kubectl exec -n redis redis-master-0 -- redis-cli config set rename-command KEYS ""
kubectl exec -n redis redis-master-0 -- redis-cli config set rename-command FLUSHDB ""
kubectl exec -n redis redis-master-0 -- redis-cli config set rename-command FLUSHALL ""
```

### Step 4: Application-Level Optimizations

Review application code for:

1. **Use Connection Pooling:**
```python
# ✅ Good: Use connection pool
import redis
pool = redis.ConnectionPool(
    host='redis-master.redis.svc.cluster.local',
    port=6379,
    max_connections=50,
    socket_timeout=5,
    socket_connect_timeout=5
)
redis_client = redis.Redis(connection_pool=pool)
```

2. **Use Pipelining:**
```python
# ❌ Bad: Multiple round trips
for i in range(1000):
    redis_client.set(f'key:{i}', f'value:{i}')

# ✅ Good: Use pipeline
pipe = redis_client.pipeline()
for i in range(1000):
    pipe.set(f'key:{i}', f'value:{i}')
pipe.execute()
```

3. **Use SCAN Instead of KEYS:**
```python
# ❌ Bad: Blocks Redis
keys = redis_client.keys('pattern:*')

# ✅ Good: Non-blocking
for key in redis_client.scan_iter('pattern:*', count=100):
    process(key)
```

4. **Optimize Data Structures:**
```python
# ❌ Bad: Large lists with full range
items = redis_client.lrange('list:key', 0, -1)

# ✅ Good: Paginate
items = redis_client.lrange('list:key', 0, 99)  # First 100 items
```

## Verification

### 1. Check Latency Improved

```bash
# Run latency test
kubectl exec -n redis redis-master-0 -- redis-cli --latency-history
# Should show < 10ms typically

# Check latency percentiles
kubectl exec -n redis redis-master-0 -- redis-cli --latency-dist
```

### 2. Verify Slow Log Empty or Minimal

```bash
kubectl exec -n redis redis-master-0 -- redis-cli slowlog len
# Should be low or zero

kubectl exec -n redis redis-master-0 -- redis-cli slowlog get 10
```

### 3. Check Operations Per Second

```bash
kubectl exec -n redis redis-master-0 -- redis-cli info stats | grep instantaneous_ops_per_sec
# Should be within normal range for your workload
```

### 4. Monitor Application Response Times

```bash
# Check application metrics
kubectl port-forward -n bruno svc/agent-bruno-service 8080:8080
curl http://localhost:8080/metrics | grep redis_operation_duration
```

### 5. Verify No Blocked Clients

```bash
kubectl exec -n redis redis-master-0 -- redis-cli info clients | grep blocked_clients
# Should be 0 or very low
```

## Prevention

### 1. Enable Comprehensive Monitoring

```yaml
# Prometheus alerts for Redis performance
- alert: RedisSlowOperations
  expr: redis_slowlog_length > 10
  for: 5m
  labels:
    severity: warning

- alert: RedisHighLatency
  expr: redis_latency_seconds > 0.1
  for: 5m
  labels:
    severity: warning

- alert: RedisHighOpsPerSec
  expr: rate(redis_commands_processed_total[1m]) > 10000
  for: 5m
  labels:
    severity: warning
```

### 2. Application Best Practices

- ✅ Always use connection pooling
- ✅ Use pipelining for bulk operations
- ✅ Use SCAN instead of KEYS
- ✅ Implement pagination for large collections
- ✅ Set appropriate timeouts
- ✅ Use appropriate data structures
- ✅ Monitor application Redis usage

### 3. Regular Performance Testing

```bash
# Benchmark Redis performance
kubectl exec -n redis redis-master-0 -- redis-benchmark -h localhost -p 6379 -n 100000 -c 50 -q
```

### 4. Optimize Redis Configuration

```yaml
# In HelmRelease values
master:
  configuration: |
    # Performance tuning
    tcp-backlog 511
    tcp-keepalive 300
    timeout 0
    
    # Slow log
    slowlog-log-slower-than 10000
    slowlog-max-len 500
    
    # Latency monitoring
    latency-monitor-threshold 100
    
    # Memory
    maxmemory-policy allkeys-lfu
    
    # Disable dangerous commands
    rename-command KEYS ""
    rename-command FLUSHDB ""
    rename-command FLUSHALL ""
```

### 5. Resource Planning

```yaml
master:
  resources:
    limits:
      cpu: "1000m"
      memory: "1Gi"
    requests:
      cpu: "200m"
      memory: "512Mi"
```

## Performance Benchmarking Script

```bash
#!/bin/bash
# redis-performance-test.sh

NAMESPACE="redis"
POD="redis-master-0"

echo "=== Redis Performance Benchmark ==="
echo ""

echo "1. Latency Test (30 seconds):"
kubectl exec -n $NAMESPACE $POD -- redis-cli --latency -i 1 | head -30

echo ""
echo "2. Operations Per Second:"
kubectl exec -n $NAMESPACE $POD -- redis-cli info stats | grep instantaneous_ops_per_sec

echo ""
echo "3. Slow Log Summary:"
kubectl exec -n $NAMESPACE $POD -- redis-cli slowlog get 20 | grep -E "^[0-9]+\)" | wc -l
echo "Slow operations found"

echo ""
echo "4. Command Stats (Top 10):"
kubectl exec -n $NAMESPACE $POD -- redis-cli info commandstats | \
  grep cmdstat | \
  sort -t= -k2 -rn | \
  head -10

echo ""
echo "5. Client Connections:"
kubectl exec -n $NAMESPACE $POD -- redis-cli client list | wc -l
echo "clients connected"

echo ""
echo "6. Full Benchmark:"
kubectl exec -n $NAMESPACE $POD -- redis-benchmark -h localhost -p 6379 -n 10000 -c 10 -q
```

## Related Alerts

- `RedisHighMemory`
- `RedisDown`
- `RedisConnectionPoolExhausted`
- `RedisHighCommandRate`
- `ApplicationTimeout`

## Escalation

If performance issues persist:

1. ✅ Review all resolution steps
2. 📊 Analyze slow log patterns
3. 🔍 Profile application Redis usage
4. 💾 Consider Redis Cluster for horizontal scaling
5. 🔄 Evaluate caching strategy
6. 📞 Contact development team for code review
7. 🆘 Consider migrating to managed Redis service

## Additional Resources

- [Redis Latency Troubleshooting](https://redis.io/docs/management/optimization/latency/)
- [Redis Performance Best Practices](https://redis.io/docs/management/optimization/)
- [Redis Slow Log](https://redis.io/commands/slowlog/)
- [Redis Pipelining](https://redis.io/docs/manual/pipelining/)

---

**Last Updated:** 2025-10-15  
**Version:** 1.0  
**Owner:** Infrastructure Team

