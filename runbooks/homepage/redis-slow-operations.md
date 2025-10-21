# ⚠️ Runbook: Redis Slow Operations

## Alert Information
**Alert Name:** `BrunoSiteRedisSlowOperations`  
**Severity:** Warning  

## Symptom
Redis operations are taking more than 0.1 seconds on average.

## Diagnosis

```bash
# Check Redis latency
kubectl exec -it -n redis redis-master-0 -- redis-cli --latency

# Check slow log
kubectl exec -it -n redis redis-master-0 -- redis-cli SLOWLOG GET 10

# Check for expensive commands
kubectl exec -it -n redis redis-master-0 -- redis-cli INFO commandstats
```

## Resolution
1. Identify slow commands (KEYS, SMEMBERS on large sets)
2. Replace O(N) commands with O(1) alternatives
3. Use pipelining for multiple commands
4. Consider Redis Cluster for scaling
