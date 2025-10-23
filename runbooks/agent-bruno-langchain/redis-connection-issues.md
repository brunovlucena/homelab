# 🚨 Runbook: Agent Bruno Redis Connection Issues

## Alert Information

**Alert Name:** `AgentBrunoRedisConnectionFailure`  
**Severity:** High  
**Component:** agent-bruno  
**Service:** redis-connection

## Symptom

Agent Bruno cannot connect to Redis, causing session memory failures. Recent conversation context is unavailable.

## Impact

- **User Impact:** MEDIUM - No session memory, conversations lose context
- **Business Impact:** LOW - AI still works but without recent conversation history
- **Data Impact:** MEDIUM - Session data lost, persistent data in MongoDB preserved

## Diagnosis

### 1. Check Agent Bruno Logs

```bash
kubectl logs -n bruno -l app=agent-bruno --tail=100 | grep -i redis
```

### 2. Check Redis Pod Status

```bash
kubectl get pods -n redis
kubectl describe pod -n redis -l app.kubernetes.io/name=redis
```

### 3. Check Redis Service

```bash
kubectl get svc -n redis
kubectl get endpoints -n redis redis-master
```

### 4. Test Redis Connectivity

```bash
# From agent-bruno pod
kubectl exec -it -n bruno deployment/agent-bruno -- sh -c 'nc -zv redis-master.redis.svc.cluster.local 6379'

# Or test with Python
kubectl exec -it -n bruno deployment/agent-bruno -- python3 -c "
import redis
try:
    r = redis.Redis(host='redis-master.redis.svc.cluster.local', port=6379, socket_timeout=5)
    print('PING:', r.ping())
    print('INFO:', r.info('server')['redis_version'])
except Exception as e:
    print('ERROR:', str(e))
"
```

### 5. Check Redis Health

```bash
# Connect to Redis directly
kubectl exec -it -n redis statefulset/redis-master -- redis-cli ping
kubectl exec -it -n redis statefulset/redis-master -- redis-cli info server
```

## Resolution Steps

### Step 1: Verify Redis is running

```bash
kubectl get pods -n redis -l app.kubernetes.io/name=redis
```

### Step 2: Check Redis Logs

```bash
kubectl logs -n redis -l app.kubernetes.io/name=redis --tail=100
```

### Step 3: Common Issues and Fixes

#### Issue: Redis Pod Not Running
**Cause:** Redis crashed or failed to start  
**Fix:**
```bash
# Check why Redis failed
kubectl describe pod -n redis -l app.kubernetes.io/name=redis

# Restart Redis StatefulSet
kubectl rollout restart statefulset -n redis redis-master

# Force reconcile Flux HelmRelease
flux reconcile helmrelease redis -n redis
```

#### Issue: Network Policy Blocking
**Cause:** Network policies preventing connection  
**Fix:**
```bash
# Check network policies
kubectl get networkpolicies -n redis
kubectl get networkpolicies -n bruno

# Test connectivity without network policies
kubectl exec -it -n bruno deployment/agent-bruno -- ping redis-master.redis.svc.cluster.local
```

#### Issue: Wrong Redis URL
**Cause:** Incorrect REDIS_URL environment variable  
**Fix:**
```bash
# Check current configuration
kubectl get deployment -n bruno agent-bruno -o jsonpath='{.spec.template.spec.containers[0].env[?(@.name=="REDIS_URL")].value}'

# Should be: redis://redis-master.redis.svc.cluster.local:6379

# If incorrect, update deployment
kubectl edit deployment -n bruno agent-bruno
# Update REDIS_URL environment variable
```

#### Issue: Redis Out of Memory
**Cause:** Redis maxmemory exceeded  
**Fix:**
```bash
# Check Redis memory usage
kubectl exec -it -n redis statefulset/redis-master -- redis-cli info memory

# Check maxmemory configuration
kubectl exec -it -n redis statefulset/redis-master -- redis-cli config get maxmemory

# Clear old sessions if needed (be careful!)
kubectl exec -it -n redis statefulset/redis-master -- redis-cli
# Inside redis-cli:
# SCAN 0 MATCH bruno:session:* COUNT 100
# DEL bruno:session:old-ip-address

# Or increase maxmemory in Redis HelmRelease
kubectl edit helmrelease -n redis redis
# Update maxmemory value
```

#### Issue: DNS Resolution Failure
**Cause:** DNS not resolving redis service  
**Fix:**
```bash
# Test DNS from agent-bruno pod
kubectl exec -it -n bruno deployment/agent-bruno -- nslookup redis-master.redis.svc.cluster.local

# Check CoreDNS
kubectl get pods -n kube-system -l k8s-app=kube-dns
kubectl logs -n kube-system -l k8s-app=kube-dns --tail=50
```

#### Issue: Redis Authentication Failed
**Cause:** Password mismatch (if auth enabled)  
**Fix:**
```bash
# Check if Redis requires password
kubectl exec -it -n redis statefulset/redis-master -- redis-cli config get requirepass

# Get Redis password from secret
kubectl get secret -n redis redis -o jsonpath='{.data.redis-password}' | base64 -d

# Update agent-bruno deployment with password
kubectl edit deployment -n bruno agent-bruno
# Update REDIS_URL: redis://:password@redis-master.redis.svc.cluster.local:6379
```

### Step 4: Restart Agent Bruno

```bash
kubectl rollout restart deployment/agent-bruno -n bruno
kubectl rollout status deployment/agent-bruno -n bruno
```

## Verification

1. Check Redis connectivity from agent-bruno:
```bash
kubectl exec -it -n bruno deployment/agent-bruno -- python3 -c "
import redis
r = redis.Redis(host='redis-master.redis.svc.cluster.local', port=6379)
print('Connected:', r.ping())
print('Set test:', r.setex('test:key', 10, 'test-value'))
print('Get test:', r.get('test:key'))
print('Keys:', r.keys('bruno:session:*')[:5])
"
```

2. Test session storage via API:
```bash
kubectl port-forward -n bruno svc/agent-bruno-service 8080:8080

# Send a chat message (creates session)
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -H "X-Forwarded-For: 192.168.1.100" \
  -d '{"message": "Hello"}'

# Check memory stats
curl http://localhost:8080/memory/192.168.1.100
```

3. Verify Redis has session data:
```bash
kubectl exec -it -n redis statefulset/redis-master -- redis-cli
# Inside redis-cli:
# KEYS bruno:session:*
# LRANGE bruno:session:192.168.1.100 0 -1
```

4. Check Agent Bruno metrics:
```bash
kubectl port-forward -n bruno svc/agent-bruno-service 8080:8080
curl http://localhost:8080/metrics | grep bruno_memory_operations
```

## Prevention

1. Monitor Redis health and availability
2. Set up Redis persistence and backups
3. Configure proper resource limits for Redis
4. Implement connection pooling
5. Set up Redis Sentinel for high availability
6. Monitor Redis memory usage
7. Configure appropriate maxmemory-policy
8. Test Redis connectivity in deployment pipeline

## Performance Tips

1. **Session TTL**: Adjust `SESSION_TTL` environment variable (default: 86400s / 24h)
2. **Memory Limits**: Monitor and adjust Redis maxmemory
3. **Connection Pooling**: Redis client uses connection pooling by default
4. **Eviction Policy**: Redis configured with `allkeys-lru` eviction

## Related Alerts

- `AgentBrunoAPIDown`
- `RedisDown`
- `RedisHighMemory`
- `AgentBrunoMemoryOperationsFailed`

## Escalation

If unable to resolve within 30 minutes:
1. Check Redis StatefulSet configuration
2. Verify Redis HelmRelease values
3. Check cluster networking and DNS
4. Review Redis logs for errors
5. Contact infrastructure team

## Additional Resources

- [Redis Documentation](https://redis.io/docs/)
- [Redis Python Client](https://redis-py.readthedocs.io/)
- [Agent Bruno Memory System](../../../flux/clusters/homelab/infrastructure/agent-bruno/README.md#-memory-system)
- [Kubernetes DNS Troubleshooting](https://kubernetes.io/docs/tasks/administer-cluster/dns-debugging-resolution/)

---

**Last Updated:** 2025-10-15  
**Version:** 1.0

