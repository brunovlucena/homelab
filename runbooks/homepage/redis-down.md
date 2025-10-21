# 🚨 Runbook: Bruno Site Redis Down

## Alert Information

**Alert Name:** `BrunoSiteRedisDown`  
**Severity:** Critical  
**Component:** bruno-site  
**Service:** redis

## Symptom

Redis cache has been down for more than 2 minutes. Session management and caching are affected.

## Impact

- **User Impact:** MODERATE - Sessions lost, slower performance
- **Business Impact:** MEDIUM - Degraded user experience
- **Data Impact:** LOW - Only cached data affected

## Diagnosis

### 1. Check Redis Pod Status

```bash
kubectl get pods -n redis
kubectl describe pod -n redis redis-master-0
```

### 2. Check Redis Logs

```bash
kubectl logs -n redis redis-master-0 --tail=100
```

### 3. Test Redis Connectivity

```bash
kubectl exec -it -n redis redis-master-0 -- redis-cli ping
```

## Resolution Steps

### Step 1: Check if Redis is running

```bash
kubectl get pods -n redis -l app.kubernetes.io/name=redis
```

### Step 2: Restart Redis if needed

```bash
kubectl rollout restart statefulset -n redis redis-master
kubectl rollout status statefulset -n redis redis-master
```

### Step 3: Verify API can connect

```bash
kubectl logs -n homepage -l app.kubernetes.io/component=api --tail=20 | grep -i redis
```

## Verification

1. Test Redis is responding:
```bash
kubectl exec -it -n redis redis-master-0 -- redis-cli ping
# Should return: PONG
```

2. Check Redis memory usage:
```bash
kubectl exec -it -n redis redis-master-0 -- redis-cli INFO memory
```

## Prevention

1. Monitor Redis memory usage
2. Configure maxmemory policies
3. Enable Redis persistence (RDB/AOF)
4. Set up Redis Sentinel for HA

## Related Alerts

- `BrunoSiteRedisHighMemoryUsage`
- `BrunoSiteRedisSlowOperations`

## Additional Resources

- [Redis Documentation](https://redis.io/documentation)

