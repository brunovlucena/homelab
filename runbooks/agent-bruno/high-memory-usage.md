# 🚨 Runbook: Agent Bruno High Memory Usage

## Alert Information

**Alert Name:** `AgentBrunoHighMemoryUsage`  
**Severity:** Warning  
**Component:** agent-bruno  
**Service:** resource-management

## Symptom

Agent Bruno pod is consuming excessive memory, approaching or exceeding its resource limits (current limit: 1Gi).

## Impact

- **User Impact:** LOW to MEDIUM - Potential slowdowns, possible OOMKills
- **Business Impact:** LOW - May cause pod restarts affecting availability
- **Data Impact:** NONE - No data loss expected

## Diagnosis

### 1. Check Current Memory Usage

```bash
# Check pod memory usage
kubectl top pods -n bruno -l app=agent-bruno

# Get detailed resource metrics
kubectl describe pod -n bruno -l app=agent-bruno | grep -A 10 "Limits:\|Requests:"
```

### 2. Check for OOMKills

```bash
# Check if pod was OOMKilled
kubectl get pods -n bruno -l app=agent-bruno -o jsonpath='{.items[*].status.containerStatuses[*].lastState.terminated.reason}'

# Check events for OOMKilled
kubectl get events -n bruno --sort-by='.lastTimestamp' | grep -i oom
```

### 3. Check Memory Trends

```bash
# View memory metrics in Prometheus
kubectl port-forward -n prometheus svc/prometheus-kube-prometheus-prometheus 9090:9090

# Query in Prometheus:
# container_memory_usage_bytes{namespace="bruno", pod=~"agent-bruno.*"}
# container_memory_working_set_bytes{namespace="bruno", pod=~"agent-bruno.*"}
```

### 4. Check Application Logs

```bash
kubectl logs -n bruno -l app=agent-bruno --tail=200 | grep -i "memory\|leak\|cache"
```

### 5. Profile Memory Usage

```bash
# Get memory stats from within the pod
kubectl exec -it -n bruno deployment/agent-bruno -- python3 -c "
import psutil
import os

process = psutil.Process(os.getpid())
mem_info = process.memory_info()
print(f'RSS: {mem_info.rss / 1024 / 1024:.2f} MB')
print(f'VMS: {mem_info.vms / 1024 / 1024:.2f} MB')
print(f'Shared: {mem_info.shared / 1024 / 1024:.2f} MB')

# System memory
vm = psutil.virtual_memory()
print(f'Total: {vm.total / 1024 / 1024:.2f} MB')
print(f'Available: {vm.available / 1024 / 1024:.2f} MB')
print(f'Used: {vm.used / 1024 / 1024:.2f} MB')
print(f'Percent: {vm.percent}%')
"
```

## Resolution Steps

### Step 1: Identify Memory Consumer

Common memory consumers in Agent Bruno:
1. **Redis connection pool** - Multiple connections
2. **MongoDB connection pool** - Multiple connections
3. **LLM context caching** - Large conversation contexts
4. **Session data** - Many active sessions
5. **Memory leaks** - Unclosed connections or cached data

### Step 2: Common Issues and Fixes

#### Issue: Too Many Active Sessions
**Cause:** Many concurrent users with large conversation contexts  
**Fix:**
```bash
# Check active sessions in Redis
kubectl exec -it -n redis statefulset/redis-master -- redis-cli
# DBSIZE
# KEYS bruno:session:*

# Check session count via metrics
kubectl port-forward -n bruno svc/agent-bruno-service 8080:8080
curl http://localhost:8080/metrics | grep bruno_active_sessions

# Reduce SESSION_TTL to expire sessions faster
kubectl edit deployment -n bruno agent-bruno
# Change SESSION_TTL from 86400 (24h) to 3600 (1h)
```

#### Issue: Connection Pool Exhaustion
**Cause:** Too many connections to Redis/MongoDB  
**Fix:**
```bash
# Check connection stats
kubectl exec -it -n bruno deployment/agent-bruno -- python3 -c "
import redis
from pymongo import MongoClient

# Redis connections
r = redis.Redis(host='redis-master.redis.svc.cluster.local')
info = r.info('clients')
print('Redis connected clients:', info['connected_clients'])

# MongoDB connections
client = MongoClient('mongodb://mongodb.mongodb.svc.cluster.local:27017')
server_status = client.admin.command('serverStatus')
print('MongoDB current connections:', server_status['connections']['current'])
"

# Configure connection pool limits in application code
# Or restart pods to clear stale connections
kubectl rollout restart deployment/agent-bruno -n bruno
```

#### Issue: Memory Leak in Application
**Cause:** Unclosed resources or circular references  
**Fix:**
```bash
# Enable memory profiling (requires code changes)
# Add memory_profiler to dependencies

# Check for Python memory leaks
kubectl exec -it -n bruno deployment/agent-bruno -- python3 -c "
import gc
gc.collect()
print('Garbage collected:', gc.collect())
print('Garbage stats:', gc.get_stats())
"

# Review application code for:
# - Unclosed file handles
# - Unclosed database connections
# - Large cached objects
# - Circular references
```

#### Issue: Large Conversation Contexts
**Cause:** Storing too much conversation history in memory  
**Fix:**
```bash
# Limit conversation context window in code
# Only keep last N messages in memory
# Store rest in MongoDB

# Clear old sessions from Redis
kubectl exec -it -n redis statefulset/redis-master -- redis-cli
# SCAN 0 MATCH bruno:session:* COUNT 1000
# Check TTL and manually expire old ones if needed
```

#### Issue: Insufficient Memory Limits
**Cause:** Application legitimately needs more memory  
**Fix:**
```bash
# Increase memory limits
kubectl edit deployment -n bruno agent-bruno

# Update resources:
#   limits:
#     memory: 2Gi  # Increase from 1Gi
#   requests:
#     memory: 512Mi  # Increase from 256Mi

# Monitor after change
kubectl top pods -n bruno -l app=agent-bruno
```

### Step 3: Scale Horizontally if Needed

```bash
# Check current HPA status
kubectl get hpa -n bruno agent-bruno-hpa

# Manually scale up if needed
kubectl scale deployment agent-bruno -n bruno --replicas=3

# Or adjust HPA
kubectl edit hpa -n bruno agent-bruno-hpa
# Update minReplicas to 2
```

### Step 4: Restart Pods to Clear Memory

```bash
# Restart deployment to clear memory
kubectl rollout restart deployment/agent-bruno -n bruno

# Monitor restart
kubectl rollout status deployment/agent-bruno -n bruno

# Check memory after restart
kubectl top pods -n bruno -l app=agent-bruno
```

## Verification

1. Check memory usage is back to normal:
```bash
kubectl top pods -n bruno -l app=agent-bruno
```

2. Verify no OOMKills:
```bash
kubectl get events -n bruno --sort-by='.lastTimestamp' | grep agent-bruno | grep -i oom
```

3. Monitor memory trends:
```bash
# Check metrics over time
kubectl port-forward -n bruno svc/agent-bruno-service 8080:8080
curl http://localhost:8080/metrics | grep process_resident_memory_bytes
```

4. Test functionality:
```bash
# Ensure agent is still working
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello"}'
```

## Prevention

1. **Resource Right-Sizing**
   - Monitor actual memory usage over time
   - Set appropriate requests and limits
   - Leave headroom for spikes (e.g., 70% utilization)

2. **Memory Management Best Practices**
   - Implement connection pooling with limits
   - Close connections explicitly
   - Use context managers for resources
   - Implement memory caching with size limits

3. **Session Management**
   - Set reasonable session TTL
   - Limit conversation context window
   - Implement session cleanup job
   - Monitor active session count

4. **Horizontal Scaling**
   - Use HPA to distribute load
   - Set appropriate scaling thresholds
   - Monitor scaling metrics

5. **Monitoring and Alerting**
   - Set up memory usage alerts (80%, 90%)
   - Monitor OOMKill events
   - Track memory trends over time
   - Alert on memory growth rate

6. **Code Optimization**
   - Profile memory usage in development
   - Use generators for large datasets
   - Implement streaming for large responses
   - Avoid caching large objects unnecessarily

## Memory Optimization Tips

1. **Python Memory Management**
   ```python
   # Use generators instead of lists
   def get_conversations():
       for conv in db.conversations.find():
           yield conv
   
   # Clear caches periodically
   import gc
   gc.collect()
   
   # Use __slots__ for classes
   class Message:
       __slots__ = ['ip', 'message', 'response', 'timestamp']
   ```

2. **Redis Memory Optimization**
   - Use appropriate data structures (Lists vs Hashes)
   - Set TTL on all session keys
   - Configure maxmemory-policy (allkeys-lru)
   - Monitor Redis memory usage

3. **MongoDB Memory Optimization**
   - Use projection to limit fields returned
   - Implement pagination for large queries
   - Close cursors explicitly
   - Use connection pooling

## Related Alerts

- `AgentBrunoOOMKilled`
- `AgentBrunoPodCrashLooping`
- `AgentBrunoAPIDown`
- `AgentBrunoHighCPUUsage`

## Escalation

If memory issues persist:
1. Review application code for memory leaks
2. Profile memory usage in development
3. Consider architectural changes (stateless design)
4. Evaluate alternative LLM implementations
5. Contact development team

## Additional Resources

- [Kubernetes Resource Management](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)
- [Python Memory Profiling](https://docs.python.org/3/library/tracemalloc.html)
- [Redis Memory Optimization](https://redis.io/docs/management/optimization/memory-optimization/)
- [MongoDB Memory Usage](https://www.mongodb.com/docs/manual/faq/diagnostics/#memory-usage)

---

**Last Updated:** 2025-10-15  
**Version:** 1.0

