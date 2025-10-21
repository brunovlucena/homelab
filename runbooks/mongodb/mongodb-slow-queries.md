# 🚨 Runbook: MongoDB Slow Queries

## Alert Information

**Alert Name:** `MongoDBSlowQueries`  
**Severity:** Warning  
**Component:** mongodb  
**Service:** mongodb  
**Threshold:** Queries taking > 3 seconds or high query latency

## Symptom

Database queries are executing slowly, causing increased application response times and potential timeouts.

## Impact

- **User Impact:** MEDIUM - Slow page loads, timeouts, poor user experience
- **Business Impact:** MEDIUM - Reduced throughput, customer dissatisfaction
- **Data Impact:** LOW - No data loss, but reduced data access performance

## Diagnosis

### 1. Identify Slow Queries

```bash
# Check currently running slow operations (> 3 seconds)
kubectl exec -n mongodb mongodb-0 -- mongosh --eval \
  'db.currentOp({"secs_running": {$gte: 3}})'

# Enable query profiling if not already enabled
kubectl exec -n mongodb mongodb-0 -- mongosh --eval \
  'db.setProfilingLevel(1, {slowms: 100})'

# Check profiling status
kubectl exec -n mongodb mongodb-0 -- mongosh --eval \
  'db.getProfilingStatus()'
```

### 2. Review Profiled Slow Queries

```bash
# Get top 10 slowest queries
kubectl exec -n mongodb mongodb-0 -- mongosh --eval \
  'db.system.profile.find().sort({millis: -1}).limit(10).pretty()'

# Analyze specific query patterns
kubectl exec -n mongodb mongodb-0 -- mongosh --eval \
  'db.system.profile.find({millis: {$gt: 1000}}).sort({ts: -1}).limit(20).forEach(function(doc) {
     print("Duration: " + doc.millis + "ms");
     print("Operation: " + doc.op);
     print("Namespace: " + doc.ns);
     if (doc.command) print("Command: " + JSON.stringify(doc.command));
     print("---");
   })'
```

### 3. Check Query Execution Plans

```bash
# Explain a specific slow query
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval \
  'db.users.find({email: "test@example.com"}).explain("executionStats")'

# Key metrics to check:
# - totalDocsExamined: Documents scanned
# - totalKeysExamined: Index keys scanned
# - executionTimeMillis: Execution time
# - stage: "COLLSCAN" indicates no index used (bad)
# - stage: "IXSCAN" indicates index used (good)
```

### 4. Check for Missing Indexes

```bash
# List all indexes for a collection
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval \
  'db.users.getIndexes()'

# Check index usage statistics
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval \
  'db.users.aggregate([{$indexStats: {}}])'

# Look for collections without indexes
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval \
  'db.getCollectionNames().forEach(function(coll) {
     var indexes = db[coll].getIndexes();
     if (indexes.length <= 1) {  // Only _id index
       print(coll + " has no custom indexes");
     }
   })'
```

### 5. Check Resource Utilization

```bash
# Check CPU and memory
kubectl top pod -n mongodb

# Check WiredTiger cache efficiency
kubectl exec -n mongodb mongodb-0 -- mongosh --eval \
  'var cache = db.serverStatus().wiredTiger.cache;
   var reads = cache["pages read into cache"];
   var requests = cache["pages requested from the cache"];
   print("Cache hit ratio: " + (((requests - reads) / requests) * 100).toFixed(2) + "%")'

# Check for lock contention
kubectl exec -n mongodb mongodb-0 -- mongosh --eval \
  'db.serverStatus().locks'
```

### 6. Check Connection and Operation Counts

```bash
# Check active connections
kubectl exec -n mongodb mongodb-0 -- mongosh --eval \
  'db.serverStatus().connections'

# Check operation counts
kubectl exec -n mongodb mongodb-0 -- mongosh --eval \
  'db.serverStatus().opcounters'

# Check global lock statistics
kubectl exec -n mongodb mongodb-0 -- mongosh --eval \
  'db.serverStatus().globalLock'
```

## Resolution

### Option 1: Create Missing Indexes

```bash
# Identify queries that need indexes from profiler
kubectl exec -n mongodb mongodb-0 -- mongosh --eval \
  'db.system.profile.find(
     {millis: {$gt: 1000}, "planSummary": /COLLSCAN/}
   ).limit(10).pretty()'

# Create index for frequently queried fields
# Single field index
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval \
  'db.users.createIndex({email: 1})'

# Compound index for multi-field queries
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval \
  'db.orders.createIndex({userId: 1, createdAt: -1})'

# Unique index
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval \
  'db.users.createIndex({username: 1}, {unique: true})'

# Background index creation (non-blocking)
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval \
  'db.large_collection.createIndex({field: 1}, {background: true})'

# Verify index creation
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval \
  'db.users.getIndexes()'
```

**Expected Time:** Seconds to minutes depending on collection size

### Option 2: Optimize Existing Queries

```bash
# Example: Limit result sets
# BAD: db.users.find()
# GOOD: db.users.find().limit(100)

# Use projection to return only needed fields
# BAD: db.users.find({email: "test@example.com"})
# GOOD: db.users.find({email: "test@example.com"}, {name: 1, email: 1})

# Use covered queries (query only uses index)
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval \
  'db.users.find({email: "test@example.com"}, {email: 1, _id: 0}).hint({email: 1})'

# Batch operations instead of individual operations
# Use bulkWrite instead of multiple insertOne/updateOne calls
```

### Option 3: Analyze and Rebuild Indexes

```bash
# Check index sizes
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval \
  'db.users.stats().indexSizes'

# Drop unused indexes (check indexStats first!)
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval \
  'db.users.dropIndex("old_unused_index")'

# Rebuild fragmented indexes (requires write lock)
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval \
  'db.users.reIndex()'
```

### Option 4: Increase WiredTiger Cache

```bash
# Check current cache configuration
kubectl exec -n mongodb mongodb-0 -- mongosh --eval \
  'db.serverStatus().wiredTiger.cache["maximum bytes configured"] / 1024 / 1024 / 1024'

# If cache hit ratio is low, increase cache size
kubectl edit helmrelease -n mongodb mongodb

# Add or modify:
# mongodbExtraFlags:
#   - "--wiredTigerCacheSizeGB=2.0"  # Increase from current

# Reconcile
flux reconcile helmrelease -n mongodb mongodb
```

### Option 5: Kill Long-Running Queries

```bash
# Identify long-running operations
kubectl exec -n mongodb mongodb-0 -- mongosh --eval \
  'db.currentOp({"secs_running": {$gte: 10}}).inprog.forEach(function(op) {
     print("OpId: " + op.opid + ", Duration: " + op.secs_running + "s");
     print("Query: " + JSON.stringify(op.query || op.command));
   })'

# Kill a specific operation by opid
kubectl exec -n mongodb mongodb-0 -- mongosh --eval \
  'db.killOp(12345)'  # Replace with actual opid

# Kill all operations running > 30 seconds (use with caution!)
kubectl exec -n mongodb mongodb-0 -- mongosh --eval \
  'db.currentOp({"secs_running": {$gte: 30}}).inprog.forEach(function(op) {
     if (op.opid) {
       print("Killing opid: " + op.opid);
       db.killOp(op.opid);
     }
   })'
```

### Option 6: Schema Optimization

```bash
# Check document sizes
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval \
  'db.users.stats().avgObjSize'

# If documents are large:
# 1. Move large fields to separate collections (normalization)
# 2. Use GridFS for binary data
# 3. Avoid deeply nested documents

# Example: Split large document
# Before: {_id, name, email, large_array: [...]}
# After:
#   users: {_id, name, email}
#   user_data: {_id, user_id, data_chunk}
```

## Post-Resolution Verification

### 1. Verify Query Performance

```bash
# Re-run slow queries with explain
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval \
  'db.users.find({email: "test@example.com"}).explain("executionStats")'

# Check that:
# - stage is "IXSCAN" (using index)
# - totalDocsExamined is close to nReturned
# - executionTimeMillis is acceptable (< 100ms)
```

### 2. Monitor Profiler

```bash
# Check for new slow queries
kubectl exec -n mongodb mongodb-0 -- mongosh --eval \
  'db.system.profile.find({ts: {$gt: new Date(Date.now() - 600000)}}).sort({millis: -1}).limit(10)'

# If no slow queries, disable profiling level 1 (keep level 0)
kubectl exec -n mongodb mongodb-0 -- mongosh --eval \
  'db.setProfilingLevel(0)'
```

### 3. Check Index Usage

```bash
# Verify indexes are being used
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval \
  'db.users.aggregate([{$indexStats: {}}])'

# Look for:
# - accesses.ops > 0 (index is being used)
# - accesses.since (when index was last accessed)
```

### 4. Verify Application Performance

```bash
# Check application logs for improved response times
kubectl logs -n bruno deployment/agent-bruno --tail=100 | grep "mongodb"

# Test from application
kubectl exec -n bruno deployment/agent-bruno -- curl -s http://localhost:8080/health
```

## Root Cause Analysis

### Common Causes

| Cause | Indicator | Solution |
|-------|-----------|----------|
| Missing Indexes | COLLSCAN in explain plan | Create appropriate indexes |
| Inefficient Queries | High docsExamined vs nReturned | Optimize query, add indexes |
| Large Result Sets | Returning thousands of docs | Add limits, pagination |
| Poor Index Selection | Wrong index used | Use hint() or refine index |
| Low Cache Hit Ratio | High page reads | Increase WiredTiger cache |
| Lock Contention | High lock wait times | Optimize writes, use bulk ops |
| Large Documents | avgObjSize > 1MB | Normalize schema, use GridFS |
| Unindexed Sorting | Sort without index | Create index on sort field |

### Analysis Commands

```bash
# Detailed query analysis
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval \
  'db.users.find({field: "value"}).explain("allPlansExecution")'

# Check for index intersection (multiple indexes used)
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval \
  'db.users.find({field1: "val1", field2: "val2"}).explain("executionStats")'

# Analyze collection statistics
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval \
  'db.users.stats()'
```

## Prevention

### 1. Index Strategy

```javascript
// Best practices for indexes:
// 1. Index frequently queried fields
db.users.createIndex({email: 1})

// 2. Compound indexes for multi-field queries
// Order matters: equality -> sort -> range
db.orders.createIndex({userId: 1, status: 1, createdAt: -1})

// 3. Partial indexes for subsets
db.orders.createIndex(
  {createdAt: -1},
  {partialFilterExpression: {status: "pending"}}
)

// 4. Text indexes for full-text search
db.articles.createIndex({title: "text", content: "text"})

// 5. TTL indexes for automatic expiration
db.sessions.createIndex(
  {createdAt: 1},
  {expireAfterSeconds: 86400}
)
```

### 2. Query Optimization Guidelines

```javascript
// Use projection to limit fields
db.users.find({email: "test@example.com"}, {name: 1, email: 1})

// Use limit for large result sets
db.users.find({status: "active"}).limit(100)

// Avoid regex without anchors
// BAD: db.users.find({name: /john/i})
// GOOD: db.users.find({name: /^john/i})  // Anchored to start

// Use aggregation pipeline efficiently
db.orders.aggregate([
  {$match: {status: "pending"}},  // Filter early
  {$sort: {createdAt: -1}},       // Use index for sort
  {$limit: 100}                   // Limit early
])
```

### 3. Regular Index Maintenance

```bash
# Weekly index review script
cat << 'EOF' > /tmp/index-review.js
db.getCollectionNames().forEach(function(collName) {
  print("\n=== " + collName + " ===");
  
  // Index stats
  db[collName].aggregate([{$indexStats: {}}]).forEach(function(stat) {
    print("Index: " + stat.name);
    print("  Ops: " + stat.accesses.ops);
    print("  Since: " + stat.accesses.since);
  });
  
  // Collection stats
  var stats = db[collName].stats();
  print("Documents: " + stats.count);
  print("Avg size: " + (stats.avgObjSize / 1024).toFixed(2) + " KB");
  print("Total index size: " + (stats.totalIndexSize / 1024 / 1024).toFixed(2) + " MB");
});
EOF

kubectl cp /tmp/index-review.js mongodb/mongodb-0:/tmp/
kubectl exec -n mongodb mongodb-0 -- mongosh myapp /tmp/index-review.js
```

### 4. Monitoring and Alerts

```yaml
# Prometheus alerts
groups:
  - name: mongodb-performance
    rules:
      - alert: MongoDBSlowQueries
        expr: |
          rate(mongodb_op_latencies_latency_total{type="commands"}[5m]) / 
          rate(mongodb_op_latencies_ops_total{type="commands"}[5m]) > 1000
        for: 10m
        annotations:
          summary: "Average MongoDB query latency > 1000ms"
          
      - alert: MongoDBHighScanRatio
        expr: |
          rate(mongodb_metrics_query_executor_scanned_total[5m]) /
          rate(mongodb_metrics_query_executor_returned_total[5m]) > 100
        for: 15m
        annotations:
          summary: "MongoDB scanning 100x more documents than returning"
          
      - alert: MongoDBLowCacheHitRatio
        expr: |
          (
            rate(mongodb_wiredtiger_cache_pages_requested_total[5m]) -
            rate(mongodb_wiredtiger_cache_pages_read_total[5m])
          ) / 
          rate(mongodb_wiredtiger_cache_pages_requested_total[5m]) < 0.9
        for: 10m
        annotations:
          summary: "MongoDB cache hit ratio < 90%"
```

### 5. Application Best Practices

```yaml
# Connection string with optimal settings
mongodb://mongodb.mongodb.svc.cluster.local:27017/myapp?
  maxPoolSize=50&
  minPoolSize=10&
  maxIdleTimeMS=30000&
  serverSelectionTimeoutMS=5000&
  socketTimeoutMS=10000

# Application code patterns:
# 1. Use connection pooling
# 2. Implement query timeouts
# 3. Use bulk operations for multiple writes
# 4. Implement pagination for large result sets
# 5. Cache frequently accessed data in Redis
```

## Escalation

**Escalate if:**
- Queries remain slow after index optimization
- System-wide performance degradation
- Cannot identify root cause
- Requires schema redesign

**Escalation Path:**
1. **30 minutes**: Engage database team
2. **1 hour**: Engage application developers
3. **2 hours**: Consider architectural review

**Communication Template:**
```
ISSUE: MongoDB Slow Query Performance
SEVERITY: Warning
SYMPTOMS: Queries taking [X]ms (threshold: [Y]ms)
AFFECTED QUERIES: [List of slow queries]
ACTIONS TAKEN:
- Analyzed execution plans
- Created indexes on [fields]
- Optimized [specific queries]
CURRENT STATUS: [Improving/Not resolved]
NEXT STEPS: [Additional optimization/Escalation]
```

## Related Runbooks

- [MongoDB High Memory](./mongodb-high-memory.md)
- [MongoDB High Connections](./mongodb-high-connections.md)
- [MongoDB Service Down](./mongodb-down.md)

## Additional Resources

- [MongoDB Query Optimization](https://docs.mongodb.com/manual/core/query-optimization/)
- [MongoDB Indexing Strategies](https://docs.mongodb.com/manual/applications/indexes/)
- [Database Profiler](https://docs.mongodb.com/manual/tutorial/manage-the-database-profiler/)
- [Explain Results](https://docs.mongodb.com/manual/reference/explain-results/)

---

**Last Updated**: 2025-10-15  
**Version**: 1.0  
**Maintainer**: Homelab Platform Team

