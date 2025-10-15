# 🚨 Runbook: Agent Bruno High Response Time

## Alert Information

**Alert Name:** `AgentBrunoHighResponseTime`  
**Severity:** Warning  
**Component:** agent-bruno  
**Service:** performance

## Symptom

Agent Bruno is taking excessively long to respond to requests (>5 seconds for chat requests).

## Impact

- **User Impact:** MEDIUM - Poor user experience, slow responses
- **Business Impact:** LOW - Functionality works but degraded
- **Data Impact:** NONE - No data impact

## Diagnosis

### 1. Check Response Time Metrics

```bash
# Port forward to access metrics
kubectl port-forward -n bruno svc/agent-bruno-service 8080:8080

# Check response time histogram
curl http://localhost:8080/metrics | grep bruno_request_duration_seconds

# Check via Prometheus
kubectl port-forward -n prometheus svc/prometheus-kube-prometheus-prometheus 9090:9090
# Query: histogram_quantile(0.95, bruno_request_duration_seconds_bucket{namespace="bruno"})
```

### 2. Check Application Logs

```bash
# Look for slow operations
kubectl logs -n bruno -l app=agent-bruno --tail=200 | grep -E "slow|timeout|took"
```

### 3. Check Resource Usage

```bash
# CPU and Memory
kubectl top pods -n bruno -l app=agent-bruno

# Check if CPU throttling is occurring
kubectl describe pod -n bruno -l app=agent-bruno | grep -A 5 "cpu"
```

### 4. Test Response Times

```bash
kubectl port-forward -n bruno svc/agent-bruno-service 8080:8080

# Time a simple health check
time curl http://localhost:8080/health

# Time a knowledge query (should be fast)
time curl "http://localhost:8080/knowledge/summary"

# Time a chat request (typically slower due to LLM)
time curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "What are the API endpoints?"}'
```

### 5. Check Dependencies

```bash
# Redis latency
kubectl exec -it -n redis statefulset/redis-master -- redis-cli --latency

# MongoDB performance
kubectl exec -it -n mongodb statefulset/mongodb -- mongosh --eval "db.serverStatus().opcounters"
```

## Resolution Steps

### Step 1: Identify Performance Bottleneck

Check logs for timing information to identify slow components:
- Redis operations
- MongoDB queries
- Ollama LLM requests
- Knowledge base searches

### Step 2: Common Performance Issues and Fixes

#### Issue: Ollama LLM Response Slow
**Cause:** Large model, high load on Ollama server, or network latency  
**Symptoms:** Chat requests take >10 seconds

**Fix:**
```bash
# Check Ollama server status
curl http://192.168.0.16:11434/api/version

# Check Ollama server load
ssh user@192.168.0.16
top
nvidia-smi  # If using GPU

# Test Ollama response time directly
time curl http://192.168.0.16:11434/api/generate -d '{
  "model": "llama2",
  "prompt": "Hello",
  "stream": false
}'

# Solutions:
# 1. Use smaller/faster model
#    ollama pull llama2:7b  # Instead of llama2:13b
# 
# 2. Use quantized model
#    ollama pull llama2:7b-q4_0
#
# 3. Implement request timeout
#    kubectl edit deployment -n bruno agent-bruno
#    # Add OLLAMA_TIMEOUT environment variable
#
# 4. Implement streaming responses
#    # Update code to use streaming
#
# 5. Deploy Ollama in cluster (closer)
#    # Reduces network latency
```

See also: [ollama-connection-issues.md](./ollama-connection-issues.md)

#### Issue: MongoDB Slow Queries
**Cause:** Missing indexes, large result sets, inefficient queries  
**Symptoms:** Memory/history operations taking >1 second

**Fix:**
```bash
# Enable MongoDB profiling
kubectl exec -it -n mongodb statefulset/mongodb -- mongosh
# use agent_bruno
# db.setProfilingLevel(2)  # Profile all operations
# db.system.profile.find().sort({ts:-1}).limit(5)

# Check for missing indexes
# db.conversations.getIndexes()

# Create missing indexes
# db.conversations.createIndex({ip: 1})
# db.conversations.createIndex({timestamp: -1})
# db.conversations.createIndex({ip: 1, timestamp: -1})

# Check slow queries
# db.system.profile.find({millis: {$gt: 100}}).sort({ts:-1})

# Optimize queries:
# - Use projection to limit fields
# - Add pagination
# - Use proper indexes
# - Limit result set size
```

#### Issue: Redis Slow Operations
**Cause:** Redis overloaded, slow commands, network latency  
**Symptoms:** Session operations taking >100ms

**Fix:**
```bash
# Check Redis latency
kubectl exec -it -n redis statefulset/redis-master -- redis-cli --latency

# Check slow log
kubectl exec -it -n redis statefulset/redis-master -- redis-cli SLOWLOG GET 10

# Check Redis stats
kubectl exec -it -n redis statefulset/redis-master -- redis-cli INFO stats

# Solutions:
# 1. Avoid slow commands (KEYS)
#    Use SCAN instead of KEYS
#
# 2. Use pipelining for multiple operations
#
# 3. Reduce data size per key
#
# 4. Monitor and limit max connections
kubectl exec -it -n redis statefulset/redis-master -- redis-cli CONFIG GET maxclients
```

#### Issue: High CPU Usage
**Cause:** CPU throttling, insufficient CPU allocation  
**Symptoms:** General slowness, high CPU metrics

**Fix:**
```bash
# Check current CPU usage
kubectl top pods -n bruno -l app=agent-bruno

# Check CPU limits
kubectl describe pod -n bruno -l app=agent-bruno | grep -A 3 "Limits:"

# Increase CPU limits
kubectl edit deployment -n bruno agent-bruno

# Update resources:
spec:
  template:
    spec:
      containers:
      - name: agent-bruno
        resources:
          requests:
            cpu: 200m      # Increase from 100m
          limits:
            cpu: 2000m     # Increase from 1000m
```

#### Issue: Network Latency
**Cause:** High latency to external services (Ollama, MongoDB, Redis)  
**Symptoms:** Consistent baseline slowness

**Fix:**
```bash
# Test network latency from pod
kubectl exec -it -n bruno deployment/agent-bruno -- ping -c 10 redis-master.redis.svc.cluster.local
kubectl exec -it -n bruno deployment/agent-bruno -- ping -c 10 mongodb.mongodb.svc.cluster.local
kubectl exec -it -n bruno deployment/agent-bruno -- ping -c 10 192.168.0.16

# Check for network policies causing issues
kubectl get networkpolicies -n bruno

# Solutions:
# 1. Deploy services in same namespace
# 2. Use node affinity to co-locate pods
# 3. Optimize network paths
# 4. Check for CNI issues
```

#### Issue: Large Context/Memory Retrieval
**Cause:** Retrieving and processing large conversation histories  
**Symptoms:** Slower for returning users with long histories

**Fix:**
```bash
# Limit context window size in code
# Only retrieve last N messages instead of full history

# Implement pagination for history queries
# Use MongoDB projections to limit fields

# Cache frequently accessed data
# Implement in-memory LRU cache for recent conversations

# Example optimization:
# db.conversations.find({ip: "x"})
#   .sort({timestamp: -1})
#   .limit(10)  # Only last 10 messages
#   .project({message: 1, response: 1, timestamp: 1})  # Only needed fields
```

#### Issue: No Connection Pooling
**Cause:** Creating new connections for each request  
**Fix:**
```bash
# Verify connection pooling is enabled (should be default)
# Check application code for proper connection management

# Redis: redis-py uses connection pooling by default
# MongoDB: pymongo uses connection pooling by default

# Ensure not creating new clients per request
```

#### Issue: Not Enough Replicas
**Cause:** Single pod handling all requests  
**Fix:**
```bash
# Check current replicas
kubectl get deployment -n bruno agent-bruno

# Scale up
kubectl scale deployment agent-bruno -n bruno --replicas=3

# Or update HPA for automatic scaling
kubectl edit hpa -n bruno agent-bruno-hpa
# Update minReplicas and adjust metrics
```

### Step 3: Implement Caching

```python
# Add caching for knowledge base queries
from functools import lru_cache

@lru_cache(maxsize=128)
def search_knowledge(query: str):
    # Knowledge base search
    pass

# Add caching for frequently accessed data
import redis
r = redis.Redis(host='redis-master.redis.svc.cluster.local')
r.setex(f'cache:knowledge:{query}', 3600, result)  # Cache for 1 hour
```

### Step 4: Implement Timeouts

```python
# Add timeouts to external calls
import requests

# Ollama timeout
response = requests.post(
    ollama_url,
    json=payload,
    timeout=10  # 10 second timeout
)

# MongoDB timeout
client = MongoClient(
    mongodb_url,
    serverSelectionTimeoutMS=5000,
    socketTimeoutMS=5000
)
```

## Verification

1. Test response times after fixes:
```bash
kubectl port-forward -n bruno svc/agent-bruno-service 8080:8080

# Multiple chat requests
for i in {1..10}; do
  echo "Request $i:"
  time curl -X POST http://localhost:8080/chat \
    -H "Content-Type: application/json" \
    -d '{"message": "What are the API endpoints?"}' \
    2>&1 | grep real
done
```

2. Check metrics:
```bash
curl http://localhost:8080/metrics | grep bruno_request_duration_seconds
```

3. Monitor in Prometheus:
```promql
# 95th percentile response time
histogram_quantile(0.95, 
  rate(bruno_request_duration_seconds_bucket{namespace="bruno"}[5m])
)

# Average response time
rate(bruno_request_duration_seconds_sum{namespace="bruno"}[5m]) 
/ 
rate(bruno_request_duration_seconds_count{namespace="bruno"}[5m])
```

## Prevention

1. **Performance Monitoring**
   - Set up Grafana dashboard for response times
   - Alert on P95 response time > 5s
   - Track response time trends
   - Monitor external dependency latency

2. **Load Testing**
   - Regular load testing in staging
   - Simulate realistic workloads
   - Test with various conversation lengths
   - Test concurrent users

3. **Optimization Best Practices**
   - Implement caching strategically
   - Use connection pooling
   - Optimize database queries and indexes
   - Use appropriate data structures
   - Profile code regularly

4. **Resource Planning**
   - Right-size resource allocations
   - Use HPA for automatic scaling
   - Monitor resource utilization trends
   - Plan for peak loads

5. **Code Optimization**
   - Async operations where possible
   - Avoid blocking calls
   - Stream large responses
   - Implement pagination

## Performance Targets

- **Health Check:** < 100ms
- **Knowledge Search:** < 500ms
- **Memory Stats:** < 1s
- **Chat (without LLM):** < 1s
- **Chat (with LLM):** < 10s

## Load Testing

```bash
# Install k6 if not already installed
# brew install k6

# Create load test script
cat <<EOF > test-agent-bruno.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '30s', target: 10 },  // Ramp up to 10 users
    { duration: '1m', target: 10 },   // Stay at 10 users
    { duration: '30s', target: 0 },   // Ramp down
  ],
};

export default function () {
  let payload = JSON.stringify({
    message: 'What are the API endpoints?',
  });

  let params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  let res = http.post('http://localhost:8080/chat', payload, params);
  
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 10s': (r) => r.timings.duration < 10000,
  });

  sleep(1);
}
EOF

# Run load test
kubectl port-forward -n bruno svc/agent-bruno-service 8080:8080 &
k6 run test-agent-bruno.js
```

## Related Alerts

- `AgentBrunoAPIDown`
- `AgentBrunoHighCPUUsage`
- `AgentBrunoHighMemoryUsage`
- `OllamaServerSlow`
- `RedisHighLatency`
- `MongoDBSlowQueries`

## Escalation

If performance issues persist:
1. Review application architecture
2. Consider code profiling
3. Evaluate LLM alternatives
4. Check infrastructure capacity
5. Contact development team

## Additional Resources

- [FastAPI Performance](https://fastapi.tiangolo.com/deployment/)
- [Python Profiling](https://docs.python.org/3/library/profile.html)
- [Redis Performance](https://redis.io/docs/management/optimization/)
- [MongoDB Performance](https://www.mongodb.com/docs/manual/administration/analyzing-mongodb-performance/)
- [Ollama Performance](https://github.com/ollama/ollama/blob/main/docs/faq.md#performance)

---

**Last Updated:** 2025-10-15  
**Version:** 1.0

