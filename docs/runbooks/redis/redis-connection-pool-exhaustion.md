# 🚨 Runbook: Redis Connection Pool Exhaustion

## Alert Information

**Alert Name:** `RedisConnectionPoolExhausted`  
**Severity:** High  
**Component:** redis  
**Service:** redis-master

## Symptom

Applications cannot connect to Redis due to connection pool exhaustion. New connection attempts fail or timeout.

## Impact

- **User Impact:** HIGH - Service degradation or complete failure
- **Business Impact:** HIGH - Operations dependent on Redis failing
- **Data Impact:** LOW - No data loss, but operations blocked

## Diagnosis

### 1. Check Connected Clients

```bash
# Count current connections
kubectl exec -n redis redis-master-0 -- redis-cli client list | wc -l

# List all clients
kubectl exec -n redis redis-master-0 -- redis-cli client list

# Get client info
kubectl exec -n redis redis-master-0 -- redis-cli info clients
# Key metrics:
# - connected_clients: Current connections
# - blocked_clients: Clients waiting on blocking ops
# - client_recent_max_input_buffer: Largest input buffer
# - client_recent_max_output_buffer: Largest output buffer
```

### 2. Check Max Connections Limit

```bash
kubectl exec -n redis redis-master-0 -- redis-cli config get maxclients
# Default: 10000

# Check if limit is being hit
CURRENT=$(kubectl exec -n redis redis-master-0 -- redis-cli client list | wc -l)
MAX=$(kubectl exec -n redis redis-master-0 -- redis-cli config get maxclients | tail -1)
echo "Connections: $CURRENT / $MAX"
```

### 3. Analyze Client Distribution

```bash
# Group clients by source IP
kubectl exec -n redis redis-master-0 -- redis-cli client list | \
  awk '{print $2}' | cut -d= -f2 | cut -d: -f1 | \
  sort | uniq -c | sort -rn

# Group by client name
kubectl exec -n redis redis-master-0 -- redis-cli client list | \
  grep -o "name=[^ ]*" | cut -d= -f2 | \
  sort | uniq -c | sort -rn
```

### 4. Check for Idle Connections

```bash
# List idle connections > 5 minutes
kubectl exec -n redis redis-master-0 -- redis-cli client list | \
  awk '$0 ~ /idle=[3-9][0-9][0-9][0-9]|idle=[0-9][0-9][0-9][0-9][0-9]/ {print}' | \
  wc -l

# Show details of idle connections
kubectl exec -n redis redis-master-0 -- redis-cli client list | \
  grep "idle=" | awk '{print $2, $4}' | sort -t= -k2 -rn | head -20
```

### 5. Check Application Connection Pools

```bash
# Check agent-bruno logs for connection errors
kubectl logs -n bruno -l app=agent-bruno --tail=100 | grep -i "redis\|connection\|pool"

# Check homepage logs
kubectl logs -n homepage -l app=homepage-api --tail=100 | grep -i "redis\|connection\|pool"

# Look for timeout errors
kubectl logs -n bruno -l app=agent-bruno --tail=100 | grep -i "timeout\|refused"
```

### 6. Check for Connection Leaks

```bash
# Monitor connection count over time
for i in {1..10}; do
  COUNT=$(kubectl exec -n redis redis-master-0 -- redis-cli client list | wc -l)
  echo "$(date): $COUNT connections"
  sleep 5
done

# Should be relatively stable, not continuously increasing
```

## Resolution Steps

### Step 1: Immediate Actions

#### Kill Idle Connections

```bash
# Kill connections idle > 10 minutes (600 seconds)
kubectl exec -n redis redis-master-0 -- redis-cli client list | \
  awk '$0 ~ /idle=[6-9][0-9][0-9]|idle=[0-9][0-9][0-9][0-9]/ {
    match($0, /addr=([^ ]+)/, addr);
    system("redis-cli -h localhost client kill " addr[1])
  }'

# Or use timeout setting
kubectl exec -n redis redis-master-0 -- redis-cli config set timeout 300
# Closes connections idle for 5 minutes
```

#### Increase Max Connections (Temporary)

```bash
# Check current limit
kubectl exec -n redis redis-master-0 -- redis-cli config get maxclients

# Increase temporarily
kubectl exec -n redis redis-master-0 -- redis-cli config set maxclients 20000

# Make permanent in HelmRelease
kubectl edit helmrelease redis -n redis
# Add:
#   master:
#     configuration: |
#       maxclients 20000
```

### Step 2: Fix Application Connection Pools

#### Issue: Connection Pool Not Configured
**Cause:** Applications creating new connections instead of pooling  
**Fix:**

Update application code to use connection pooling:

```python
# ❌ Bad: New connection per request
def get_data():
    redis_client = redis.Redis(host='redis-master.redis.svc.cluster.local')
    return redis_client.get('key')

# ✅ Good: Use connection pool
import redis

# Create pool once (global or singleton)
redis_pool = redis.ConnectionPool(
    host='redis-master.redis.svc.cluster.local',
    port=6379,
    max_connections=50,  # Limit per instance
    socket_timeout=5,
    socket_connect_timeout=5,
    socket_keepalive=True,
    socket_keepalive_options={
        socket.TCP_KEEPIDLE: 60,
        socket.TCP_KEEPINTVL: 10,
        socket.TCP_KEEPCNT: 3
    }
)
redis_client = redis.Redis(connection_pool=redis_pool)

# Use the pooled client
def get_data():
    return redis_client.get('key')
```

#### Issue: Pool Size Too Large
**Cause:** Each application pod has huge pool, multiplied by replicas  
**Fix:**

```python
# Calculate: (replicas × max_connections) < Redis maxclients
# Example: 10 pods × 50 connections = 500 total

# ❌ Bad: Pool too large
pool = redis.ConnectionPool(max_connections=1000)  # Per pod!

# ✅ Good: Reasonable pool size
pool = redis.ConnectionPool(
    max_connections=20,  # 20 per pod
    # With 10 pods = 200 total connections
)
```

Check and adjust in application deployment:

```bash
# Check number of replicas
kubectl get deployment -n bruno agent-bruno -o jsonpath='{.spec.replicas}'
kubectl get deployment -n homepage homepage-api -o jsonpath='{.spec.replicas}'

# If replicas are high, reduce pool size per pod
# Or scale down replicas if over-provisioned
```

#### Issue: Connections Not Being Released
**Cause:** Application not properly closing connections  
**Fix:**

```python
# ❌ Bad: Connection not returned to pool
def process():
    client = redis.Redis(connection_pool=pool)
    # ... use client ...
    # Forgot to close/release!

# ✅ Good: Use context manager
def process():
    # Connection automatically returned to pool
    with redis.Redis(connection_pool=pool) as client:
        return client.get('key')

# ✅ Also good: Use module-level client
# Connection automatically managed by pool
redis_client = redis.Redis(connection_pool=pool)

def process():
    return redis_client.get('key')  # Connection reused
```

#### Issue: Blocking Operations Holding Connections
**Cause:** BLPOP/BRPOP holding connections for long time  
**Fix:**

```bash
# Check for blocked clients
kubectl exec -n redis redis-master-0 -- redis-cli info clients | grep blocked_clients

# List blocked clients
kubectl exec -n redis redis-master-0 -- redis-cli client list | grep "flags=b"

# Solutions:
# 1. Use separate pool for blocking operations
# 2. Set reasonable timeouts on BLPOP/BRPOP
# 3. Use pub/sub instead of blocking lists
```

```python
# ✅ Good: Separate pools
# Regular pool
regular_pool = redis.ConnectionPool(max_connections=20)
regular_client = redis.Redis(connection_pool=regular_pool)

# Dedicated pool for blocking ops
blocking_pool = redis.ConnectionPool(max_connections=5)
blocking_client = redis.Redis(connection_pool=blocking_pool)

# Use with timeout
blocking_client.blpop('queue', timeout=10)  # Not forever!
```

### Step 3: Optimize Redis Configuration

```bash
# Set connection timeout (auto-close idle connections)
kubectl exec -n redis redis-master-0 -- redis-cli config set timeout 300

# Increase TCP backlog
kubectl exec -n redis redis-master-0 -- redis-cli config set tcp-backlog 511

# Enable TCP keepalive
kubectl exec -n redis redis-master-0 -- redis-cli config set tcp-keepalive 300

# Make permanent
kubectl edit helmrelease redis -n redis
```

```yaml
master:
  configuration: |
    # Connection settings
    maxclients 20000
    timeout 300          # Close idle connections after 5 minutes
    tcp-backlog 511
    tcp-keepalive 300
```

### Step 4: Scale Application Pods

If too many replicas creating too many connections:

```bash
# Check current replicas
kubectl get deployment -n bruno agent-bruno

# Scale down if over-provisioned
kubectl scale deployment -n bruno agent-bruno --replicas=3

# Or use HPA to auto-scale
kubectl get hpa -n bruno
```

### Step 5: Restart Affected Applications

```bash
# Restart to clear connection leaks
kubectl rollout restart deployment/agent-bruno -n bruno
kubectl rollout restart deployment/homepage-api -n homepage

# Wait for rollout
kubectl rollout status deployment/agent-bruno -n bruno
```

## Verification

### 1. Check Connection Count Stabilized

```bash
# Monitor connections over time
watch -n 5 "kubectl exec -n redis redis-master-0 -- redis-cli client list | wc -l"

# Should be stable, not continuously growing
```

### 2. Verify No Connection Errors

```bash
# Check application logs
kubectl logs -n bruno -l app=agent-bruno --tail=50 | grep -i "error\|connection"
kubectl logs -n homepage -l app=homepage-api --tail=50 | grep -i "error\|connection"
```

### 3. Test Application Connectivity

```bash
# Test from agent-bruno
kubectl exec -n bruno deployment/agent-bruno -- python3 -c "
import redis
pool = redis.ConnectionPool(host='redis-master.redis.svc.cluster.local', max_connections=10)
client = redis.Redis(connection_pool=pool)
print('PING:', client.ping())
print('Pool:', pool)
"
```

### 4. Check Connection Distribution

```bash
# Verify reasonable distribution
kubectl exec -n redis redis-master-0 -- redis-cli client list | \
  awk '{print $2}' | cut -d= -f2 | cut -d: -f1 | \
  sort | uniq -c | sort -rn | head -10
```

### 5. Monitor Redis Metrics

```bash
kubectl exec -n redis redis-master-0 -- redis-cli info clients
kubectl exec -n redis redis-master-0 -- redis-cli info stats
```

## Prevention

### 1. Configure Proper Connection Pools

**Application Configuration:**

```python
# config.py
REDIS_POOL_CONFIG = {
    'host': 'redis-master.redis.svc.cluster.local',
    'port': 6379,
    'max_connections': 20,  # Per pod
    'socket_timeout': 5,
    'socket_connect_timeout': 5,
    'socket_keepalive': True,
    'retry_on_timeout': True,
    'health_check_interval': 30
}

# Initialize once
redis_pool = redis.ConnectionPool(**REDIS_POOL_CONFIG)
redis_client = redis.Redis(connection_pool=redis_pool)
```

### 2. Set Resource Limits

```yaml
# Calculate total connections
# Formula: (num_apps × replicas_per_app × connections_per_pod) + buffer

# Example:
# - agent-bruno: 3 replicas × 20 connections = 60
# - homepage-api: 5 replicas × 20 connections = 100
# - other apps: ~40
# Total: ~200 + 100 buffer = 300

master:
  configuration: |
    maxclients 10000  # Conservative limit
    timeout 300       # Auto-close idle connections
```

### 3. Implement Health Checks

```python
# Health check with connection pool verification
from flask import Flask
import redis

app = Flask(__name__)

@app.route('/health')
def health():
    try:
        # Check Redis connectivity
        redis_client.ping()
        
        # Check pool stats
        pool_info = redis_pool.get_connection('_')
        redis_pool.release(pool_info)
        
        return {'status': 'healthy', 'redis': 'ok'}, 200
    except Exception as e:
        return {'status': 'unhealthy', 'error': str(e)}, 503
```

### 4. Set Up Monitoring

```yaml
# Prometheus alerts
- alert: RedisHighConnections
  expr: redis_connected_clients > 1000
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Redis has too many connections"

- alert: RedisConnectionPoolExhausted
  expr: redis_connected_clients / redis_config_maxclients > 0.8
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "Redis connection pool near exhaustion"

- alert: RedisBlockedClients
  expr: redis_blocked_clients > 10
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Many Redis clients blocked"
```

### 5. Regular Connection Audits

```bash
#!/bin/bash
# redis-connection-audit.sh

echo "=== Redis Connection Audit ==="
echo "Date: $(date)"
echo ""

TOTAL=$(kubectl exec -n redis redis-master-0 -- redis-cli client list | wc -l)
MAX=$(kubectl exec -n redis redis-master-0 -- redis-cli config get maxclients | tail -1)
echo "Total Connections: $TOTAL / $MAX"

echo ""
echo "Connections by Source:"
kubectl exec -n redis redis-master-0 -- redis-cli client list | \
  awk '{print $2}' | cut -d= -f2 | cut -d: -f1 | \
  sort | uniq -c | sort -rn

echo ""
echo "Idle Connections (>5min):"
kubectl exec -n redis redis-master-0 -- redis-cli client list | \
  awk '$0 ~ /idle=[3-9][0-9][0-9]|idle=[0-9][0-9][0-9][0-9]/ {print}' | \
  wc -l

echo ""
echo "Blocked Clients:"
kubectl exec -n redis redis-master-0 -- redis-cli info clients | grep blocked_clients
```

### 6. Connection Pool Best Practices

- ✅ Use connection pooling in all applications
- ✅ Set reasonable pool sizes (10-50 per pod)
- ✅ Calculate total connections: replicas × pool_size
- ✅ Enable socket keepalive
- ✅ Set connection timeouts
- ✅ Implement proper error handling
- ✅ Monitor pool utilization
- ✅ Use separate pools for blocking operations
- ✅ Close connections properly (use context managers)
- ✅ Implement connection health checks

## Performance Tips

1. **Connection Reuse:** Always use connection pooling
2. **Proper Sizing:** Pool size = typical concurrent operations + buffer
3. **Timeouts:** Set reasonable timeouts (5-10s)
4. **Keepalive:** Enable TCP keepalive to detect dead connections
5. **Monitoring:** Track pool utilization metrics
6. **Graceful Degradation:** Handle connection errors gracefully

## Related Alerts

- `RedisHighConnections`
- `RedisDown`
- `ApplicationConnectionTimeout`
- `RedisBlockedClients`
- `RedisMaxClientsReached`

## Escalation

If connection pool issues persist:

1. ✅ Verify application connection pool configuration
2. 📊 Analyze connection patterns and growth
3. 🔍 Check for connection leaks in application code
4. 💾 Review application logs for patterns
5. 🔄 Consider Redis Cluster for horizontal scaling
6. 📞 Contact development team for code review
7. 🆘 Implement circuit breakers and backoff strategies

## Additional Resources

- [Redis Clients Documentation](https://redis.io/docs/clients/)
- [redis-py Connection Pooling](https://redis-py.readthedocs.io/en/stable/connections.html)
- [Connection Pool Best Practices](https://redis.io/docs/manual/patterns/connection-handling/)
- [Redis CLIENT command](https://redis.io/commands/client/)

---

**Last Updated:** 2025-10-15  
**Version:** 1.0  
**Owner:** Infrastructure Team

