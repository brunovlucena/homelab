# ⚠️ Runbook: Redis High Memory Usage

## Alert Information
**Alert Name:** `BrunoSiteRedisHighMemoryUsage`  
**Severity:** Warning  

## Symptom
Redis memory usage is above 80% of max memory.

## Diagnosis
```bash
kubectl exec -it -n redis redis-master-0 -- redis-cli INFO memory
kubectl exec -it -n redis redis-master-0 -- redis-cli INFO keyspace
```

## Resolution
```bash
# Check for large keys
kubectl exec -it -n redis redis-master-0 -- redis-cli --bigkeys

# Flush old data if safe
kubectl exec -it -n redis redis-master-0 -- redis-cli FLUSHDB

# Increase maxmemory
kubectl edit statefulset -n redis redis-master
```

## Prevention
- Configure maxmemory-policy (e.g., allkeys-lru)
- Set appropriate TTLs
- Monitor key growth
