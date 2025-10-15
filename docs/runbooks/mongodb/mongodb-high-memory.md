# 🚨 Runbook: MongoDB High Memory Usage

## Alert Information

**Alert Name:** `MongoDBHighMemory`  
**Severity:** Warning  
**Component:** mongodb  
**Service:** mongodb  
**Threshold:** Memory usage > 80% of limit or OOMKilled

## Symptom

MongoDB memory usage is approaching or has exceeded configured limits, causing memory pressure, performance degradation, or OOM kills.

## Impact

- **User Impact:** MEDIUM - Slow database operations, potential service interruptions
- **Business Impact:** MEDIUM - Degraded application performance, possible downtime
- **Data Impact:** HIGH - Risk of OOM kills causing data loss or corruption

## Diagnosis

### 1. Check Current Memory Usage

```bash
# Check pod memory usage
kubectl top pod -n mongodb

# Check memory statistics
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.serverStatus().mem'

# Output shows:
# - resident: Physical RAM used (RSS)
# - virtual: Virtual memory
# - mapped: Memory-mapped files
```

### 2. Check WiredTiger Cache

```bash
# WiredTiger cache statistics
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.serverStatus().wiredTiger.cache' | jq

# Key metrics:
# - bytes currently in the cache
# - maximum bytes configured
# - pages evicted from cache
# - pages read into cache
```

### 3. Check for OOMKills

```bash
# Check if pod was OOMKilled recently
kubectl get pods -n mongodb -o json | \
  jq -r '.items[] | select(.status.containerStatuses[].lastState.terminated.reason == "OOMKilled") | .metadata.name'

# Check restart count
kubectl get pods -n mongodb -o jsonpath='{.items[0].status.containerStatuses[0].restartCount}'

# View OOMKill events
kubectl get events -n mongodb | grep -i oom
```

### 4. Check Resource Limits

```bash
# Current resource configuration
kubectl get pod -n mongodb mongodb-0 -o json | jq '.spec.containers[0].resources'

# Compare with usage
kubectl describe pod -n mongodb mongodb-0 | grep -A3 "Limits:\|Requests:"
```

### 5. Analyze Memory Usage Patterns

```bash
# Check database sizes
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.adminCommand({listDatabases: 1})' | jq '.databases'

# Check collection stats
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval 'db.stats()' | jq

# Check index sizes
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval 'db.users.stats()' | jq '.indexSizes'

# Memory usage by operation
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.serverStatus().globalLock'
```

### 6. Check for Memory Leaks

```bash
# Check connection count (potential leak)
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.serverStatus().connections'

# Check cursors (potential leak)
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.serverStatus().metrics.cursor'

# Check active operations
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.currentOp()'
```

## Resolution

### Option 1: Immediate Mitigation - Restart Pod

```bash
# If MongoDB is unresponsive due to memory pressure
kubectl delete pod -n mongodb mongodb-0

# Watch pod restart
kubectl get pods -n mongodb -w

# Verify it's back and responsive
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.runCommand({ping: 1})'
```

**Expected Time:** 2-3 minutes  
**Note:** Temporary fix, does not address root cause

### Option 2: Increase Memory Limits

```bash
# Edit HelmRelease to increase memory
kubectl edit helmrelease -n mongodb mongodb

# Update the values section:
# resources:
#   limits:
#     memory: 4Gi  # Increase from current (e.g., 2Gi -> 4Gi)
#   requests:
#     memory: 2Gi  # Also increase requests

# Save and exit, Flux will reconcile
flux reconcile helmrelease -n mongodb mongodb

# Monitor the rollout
kubectl rollout status statefulset -n mongodb mongodb

# Verify new limits
kubectl get pod -n mongodb mongodb-0 -o jsonpath='{.spec.containers[0].resources.limits.memory}'
```

**Expected Time:** 5-10 minutes

### Option 3: Optimize WiredTiger Cache Size

```bash
# MongoDB's WiredTiger engine uses 50% of RAM by default
# Can be adjusted via configuration

kubectl edit helmrelease -n mongodb mongodb

# Add storage engine configuration:
# mongodbExtraFlags:
#   - "--wiredTigerCacheSizeGB=1.5"  # Adjust based on available memory

# Reconcile
flux reconcile helmrelease -n mongodb mongodb

# Verify new cache size
kubectl exec -n mongodb mongodb-0 -- mongosh --eval \
  'db.serverStatus().wiredTiger.cache["maximum bytes configured"]'
```

### Option 4: Identify and Optimize Heavy Queries

```bash
# Enable profiling to capture slow queries
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.setProfilingLevel(1, {slowms: 100})'

# Check profiling data
kubectl exec -n mongodb mongodb-0 -- mongosh --eval \
  'db.system.profile.find().sort({ts: -1}).limit(10).pretty()'

# Look for queries with high:
# - docsExamined (indicates missing index)
# - keysExamined (indicates inefficient index)
# - millis (execution time)

# Create missing indexes
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval \
  'db.collection.createIndex({field: 1})'
```

### Option 5: Clean Up Unused Data

```bash
# Identify large collections
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval \
  'db.getCollectionNames().forEach(function(c) {
     var stats = db[c].stats();
     print(c + ": " + (stats.size / 1024 / 1024).toFixed(2) + " MB");
   })'

# Archive or delete old data
# Example: Remove documents older than 90 days
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval \
  'db.logs.deleteMany({createdAt: {$lt: new Date(Date.now() - 90*24*60*60*1000)}})'

# Compact collection to reclaim space
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval 'db.runCommand({compact: "logs"})'

# Drop unused indexes
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval 'db.logs.dropIndex("old_index_name")'
```

### Option 6: Reduce Connection Pool Size

```bash
# Check connection count
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.serverStatus().connections'

# If high, check which clients are holding connections
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.currentOp(true).inprog.forEach(function(op) { if(op.client) print(op.client); })'

# Update application connection strings to reduce pool size
# Example connection string parameter: maxPoolSize=20
```

## Post-Resolution Verification

### 1. Verify Memory Stability

```bash
# Monitor memory usage over time
watch -n 5 'kubectl top pod -n mongodb'

# Check memory metrics
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.serverStatus().mem'

# Verify no OOMKills in past hour
kubectl get events -n mongodb --sort-by='.lastTimestamp' | grep -i oom
```

### 2. Check Performance

```bash
# Verify query performance
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval \
  'db.users.find({email: "test@example.com"}).explain("executionStats")'

# Check cache hit ratio (should be > 90%)
kubectl exec -n mongodb mongodb-0 -- mongosh --eval \
  'var stats = db.serverStatus().wiredTiger.cache;
   var reads = stats["pages read into cache"];
   var requests = stats["pages requested from the cache"];
   print("Cache hit ratio: " + (((requests - reads) / requests) * 100).toFixed(2) + "%")'
```

### 3. Verify WiredTiger Cache

```bash
# Check cache configuration
kubectl exec -n mongodb mongodb-0 -- mongosh --eval \
  'print("Max cache size: " + db.serverStatus().wiredTiger.cache["maximum bytes configured"] / 1024 / 1024 / 1024 + " GB")'

# Check cache utilization
kubectl exec -n mongodb mongodb-0 -- mongosh --eval \
  'print("Current cache size: " + db.serverStatus().wiredTiger.cache["bytes currently in the cache"] / 1024 / 1024 / 1024 + " GB")'
```

## Root Cause Analysis

### Common Causes

| Cause | Indicator | Solution |
|-------|-----------|----------|
| Insufficient Memory Limits | Memory usage at 100% | Increase memory limits |
| Large Working Set | High cache evictions | Increase cache size or optimize queries |
| Memory Leak | Steady increase over time | Restart pod, update MongoDB version |
| Too Many Connections | High connection count | Reduce application connection pools |
| Inefficient Queries | High docsExamined | Add missing indexes |
| Large Documents | High avgObjSize | Normalize schema, use GridFS for large files |
| Unused Indexes | Many indexes per collection | Drop unused indexes |

### Investigation Commands

```bash
# Memory growth over time (requires historical data)
kubectl logs -n mongodb mongodb-0 | grep -i "memory\|cache"

# Check for memory fragmentation
kubectl exec -n mongodb mongodb-0 -- mongosh --eval \
  'db.serverStatus().tcmalloc.pageheap'

# Detailed memory breakdown
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.serverStatus().wiredTiger.cache' | \
  jq '{
    max_bytes: .["maximum bytes configured"],
    current_bytes: .["bytes currently in the cache"],
    dirty_bytes: .["tracked dirty bytes in the cache"],
    pages_evicted: .["pages evicted by application threads"]
  }'
```

## Prevention

### 1. Right-Size Resources

```yaml
# Recommended configuration for production
resources:
  requests:
    cpu: 500m
    memory: 2Gi      # 2x working set size
  limits:
    cpu: 2000m
    memory: 4Gi      # 4x working set size

# WiredTiger cache sizing
# Rule of thumb: (RAM - 1GB) * 0.5
# For 4GB limit: ~1.5GB cache
mongodbExtraFlags:
  - "--wiredTigerCacheSizeGB=1.5"
```

### 2. Query Optimization

```bash
# Regularly review slow queries
kubectl exec -n mongodb mongodb-0 -- mongosh --eval \
  'db.setProfilingLevel(1, {slowms: 100})'

# Weekly index review
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval \
  'db.getCollectionNames().forEach(function(c) {
     print("\n" + c + " indexes:");
     db[c].getIndexes().forEach(function(idx) {
       print("  - " + idx.name);
     });
   })'
```

### 3. Connection Management

```yaml
# Application configuration
# Use connection pooling with limits
mongodb:
  connection_string: "mongodb://mongodb.mongodb.svc.cluster.local:27017/myapp?maxPoolSize=50&minPoolSize=10&maxIdleTimeMS=30000"
```

### 4. Monitoring and Alerts

```yaml
# Prometheus alerts
groups:
  - name: mongodb-memory
    rules:
      - alert: MongoDBHighMemory
        expr: |
          mongodb_memory_resident_bytes / 
          (kube_pod_container_resource_limits{resource="memory", namespace="mongodb"} * 1024 * 1024) > 0.8
        for: 5m
        annotations:
          summary: "MongoDB memory usage > 80%"
          
      - alert: MongoDBHighCacheEviction
        expr: rate(mongodb_wiredtiger_cache_pages_evicted_total[5m]) > 100
        for: 10m
        annotations:
          summary: "High WiredTiger cache eviction rate"
          
      - alert: MongoDBOOMKilled
        expr: |
          rate(kube_pod_container_status_restarts_total{namespace="mongodb"}[15m]) > 0
          and
          kube_pod_container_status_last_terminated_reason{namespace="mongodb", reason="OOMKilled"} == 1
        annotations:
          summary: "MongoDB pod was OOMKilled"
```

### 5. Regular Maintenance

```bash
# Weekly tasks
# 1. Check database growth
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.adminCommand({listDatabases: 1})'

# 2. Review collection sizes
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval \
  'db.getCollectionNames().forEach(function(c) {
     print(c + ": " + db[c].stats().size);
   })'

# 3. Compact collections if needed (requires downtime)
# kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval 'db.runCommand({compact: "large_collection"})'

# 4. Review and optimize indexes
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval \
  'db.users.aggregate([{$indexStats: {}}])'
```

## Escalation

**Escalate if:**
- Memory usage continues to grow after optimization
- Multiple OOMKills within 1 hour
- Cannot identify memory leak source
- Requires significant schema changes

**Escalation Path:**
1. **30 minutes**: Engage database team
2. **1 hour**: Engage capacity planning team
3. **2 hours**: Consider architectural changes

**Communication Template:**
```
ISSUE: MongoDB High Memory Usage
SEVERITY: Warning
CURRENT MEMORY: [X]Gi / [Y]Gi ([Z]%)
SYMPTOMS: [OOMKills/Performance degradation/etc]
ACTIONS TAKEN:
- [List mitigation steps]
ROOT CAUSE: [If identified]
RECOMMENDATION: [Increase resources/Optimize queries/etc]
```

## Related Runbooks

- [MongoDB Service Down](./mongodb-down.md)
- [MongoDB Slow Queries](./mongodb-slow-queries.md)
- [MongoDB Storage Full](./mongodb-storage-full.md)

## Additional Resources

- [MongoDB Memory Usage](https://docs.mongodb.com/manual/faq/diagnostics/#memory-usage)
- [WiredTiger Cache Tuning](https://docs.mongodb.com/manual/core/wiredtiger/#memory-use)
- [MongoDB Performance Best Practices](https://docs.mongodb.com/manual/administration/analyzing-mongodb-performance/)

---

**Last Updated**: 2025-10-15  
**Version**: 1.0  
**Maintainer**: Homelab Platform Team

