# 🚨 Runbook: MongoDB Storage Full

## Alert Information

**Alert Name:** `MongoDBStorageFull`  
**Severity:** Critical  
**Component:** mongodb  
**Service:** mongodb  
**Threshold:** Storage usage > 85% of persistent volume capacity

## Symptom

MongoDB persistent volume is running out of disk space, preventing write operations and potentially causing service failures.

## Impact

- **User Impact:** CRITICAL - Cannot write new data, application errors
- **Business Impact:** CRITICAL - Data persistence failure, potential data loss
- **Data Impact:** CRITICAL - Risk of corruption if storage completely exhausted

## Diagnosis

### 1. Check Current Disk Usage

```bash
# Check disk usage inside MongoDB pod
kubectl exec -n mongodb mongodb-0 -- df -h /bitnami/mongodb

# Detailed breakdown
kubectl exec -n mongodb mongodb-0 -- du -sh /bitnami/mongodb/*

# Check PVC status
kubectl get pvc -n mongodb

# Check PV details
kubectl get pv | grep mongodb
kubectl describe pv $(kubectl get pvc -n mongodb data-mongodb-0 -o jsonpath='{.spec.volumeName}')
```

### 2. Identify Space Consumers

```bash
# List all databases and their sizes
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.adminCommand({listDatabases: 1})'

# Detailed size breakdown per database
kubectl exec -n mongodb mongodb-0 -- mongosh --eval '
  db.adminCommand({listDatabases: 1}).databases.forEach(function(database) {
    var sizeGB = database.sizeOnDisk / 1024 / 1024 / 1024;
    print(database.name + ": " + sizeGB.toFixed(2) + " GB");
  });
'

# Check collection sizes in specific database
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval '
  db.getCollectionNames().forEach(function(collName) {
    var stats = db[collName].stats();
    var sizeMB = stats.size / 1024 / 1024;
    var storageMB = stats.storageSize / 1024 / 1024;
    print(collName + ":");
    print("  Data size: " + sizeMB.toFixed(2) + " MB");
    print("  Storage size: " + storageMB.toFixed(2) + " MB");
    print("  Index size: " + (stats.totalIndexSize / 1024 / 1024).toFixed(2) + " MB");
    print("  Documents: " + stats.count);
  });
'
```

### 3. Check for WiredTiger Bloat

```bash
# Check WiredTiger cache and storage
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.serverStatus().wiredTiger.cache' | jq

# Check journal files
kubectl exec -n mongodb mongodb-0 -- du -sh /bitnami/mongodb/data/journal

# Check diagnostic data directory
kubectl exec -n mongodb mongodb-0 -- du -sh /bitnami/mongodb/data/diagnostic.data

# Check temp files
kubectl exec -n mongodb mongodb-0 -- find /bitnami/mongodb/data -name '*.tmp' -ls
```

### 4. Check Backup Files

```bash
# Look for backup files or dumps
kubectl exec -n mongodb mongodb-0 -- find /bitnami/mongodb -name '*dump*' -o -name '*backup*' -ls

# Check for old log files
kubectl exec -n mongodb mongodb-0 -- find /bitnami/mongodb -name '*.log*' -ls
```

### 5. Check for Orphaned Files

```bash
# Check for files not owned by MongoDB
kubectl exec -n mongodb mongodb-0 -- find /bitnami/mongodb/data -type f ! -name '*.wt' ! -name '*.log' -ls

# Check for large unexpected files
kubectl exec -n mongodb mongodb-0 -- find /bitnami/mongodb -type f -size +100M -exec ls -lh {} \;
```

## Resolution

### Option 1: Emergency - Expand PVC (Immediate)

```bash
# Check if storage class supports volume expansion
kubectl get storageclass -o jsonpath='{.items[*].allowVolumeExpansion}'

# If true, expand the PVC
kubectl edit pvc -n mongodb data-mongodb-0

# Update the size:
# spec:
#   resources:
#     requests:
#       storage: 50Gi  # Increase from current (e.g., 20Gi -> 50Gi)

# Save and verify
kubectl get pvc -n mongodb -w

# Check if pod needs restart for resize to take effect
kubectl get events -n mongodb | grep -i resize

# If needed, restart pod
kubectl delete pod -n mongodb mongodb-0

# Verify new size
kubectl exec -n mongodb mongodb-0 -- df -h /bitnami/mongodb
```

**Expected Time:** 5-15 minutes depending on storage provisioner

**Note:** This is the fastest solution but doesn't address root cause

### Option 2: Delete Old/Unnecessary Data

```bash
# Identify candidates for deletion

# 1. Check for old log/audit collections
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval '
  db.getCollectionNames().filter(name => 
    name.includes("log") || name.includes("audit") || name.includes("event")
  )'

# 2. Check document counts and dates
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval '
  db.logs.find().sort({createdAt: 1}).limit(5).pretty()
  db.logs.find().sort({createdAt: -1}).limit(5).pretty()
'

# 3. Delete old data (example: logs older than 90 days)
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval '
  var cutoffDate = new Date(Date.now() - 90*24*60*60*1000);
  var result = db.logs.deleteMany({createdAt: {$lt: cutoffDate}});
  print("Deleted " + result.deletedCount + " documents");
'

# 4. Drop entire old collections if not needed
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval '
  db.old_collection.drop()
'

# 5. Drop old databases if not needed
kubectl exec -n mongodb mongodb-0 -- mongosh --eval '
  db.getSiblingDB("old_database").dropDatabase()
'
```

**Expected Time:** Minutes to hours depending on data volume

### Option 3: Compact Collections to Reclaim Space

```bash
# Compact reclaims fragmented space but requires exclusive lock
# This blocks ALL operations during compaction

# Check collection stats before
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval 'db.large_collection.stats()' | grep -E 'size|storageSize'

# Compact collection (use during maintenance window)
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval '
  db.runCommand({compact: "large_collection", force: true})
'

# Check stats after
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval 'db.large_collection.stats()' | grep -E 'size|storageSize'

# Compact all collections in database (very disruptive)
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval '
  db.getCollectionNames().forEach(function(collName) {
    if (collName != "system.profile") {
      print("Compacting " + collName);
      db.runCommand({compact: collName, force: true});
    }
  });
'
```

**Expected Time:** Minutes to hours per collection  
**Warning:** Blocks all operations, use during maintenance window

### Option 4: Drop Unused Indexes

```bash
# Identify unused indexes
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval '
  db.getCollectionNames().forEach(function(collName) {
    print("\n=== " + collName + " ===");
    db[collName].aggregate([{$indexStats: {}}]).forEach(function(indexStat) {
      print(indexStat.name + " - Ops: " + indexStat.accesses.ops);
      if (indexStat.accesses.ops === 0) {
        print("  ^ UNUSED INDEX");
      }
    });
  });
'

# Drop unused indexes (except _id)
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval '
  db.users.dropIndex("unused_index_name")
'

# Check index sizes before and after
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval '
  db.users.stats().indexSizes
'
```

**Expected Time:** Seconds to minutes  
**Space Saved:** Can be significant for large collections

### Option 5: Archive Data to External Storage

```bash
# Export old data to dump files for archival

# 1. Create export directory
kubectl exec -n mongodb mongodb-0 -- mkdir -p /tmp/archive

# 2. Export old data
kubectl exec -n mongodb mongodb-0 -- mongodump \
  --db=myapp \
  --collection=logs \
  --query='{"createdAt": {"$lt": {"$date": "2024-01-01T00:00:00Z"}}}' \
  --out=/tmp/archive

# 3. Copy to local storage
kubectl cp mongodb/mongodb-0:/tmp/archive ./mongodb-archive-$(date +%Y%m%d)

# 4. Compress archive
tar -czf mongodb-archive-$(date +%Y%m%d).tar.gz ./mongodb-archive-$(date +%Y%m%d)

# 5. Upload to object storage (if available)
# aws s3 cp mongodb-archive-$(date +%Y%m%d).tar.gz s3://backups/mongodb/

# 6. Delete archived data from MongoDB
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval '
  var cutoffDate = new Date("2024-01-01");
  var result = db.logs.deleteMany({createdAt: {$lt: cutoffDate}});
  print("Deleted " + result.deletedCount + " documents");
'

# 7. Clean up temporary files
kubectl exec -n mongodb mongodb-0 -- rm -rf /tmp/archive
```

**Expected Time:** Hours depending on data size  
**Note:** Preserves data while freeing space

### Option 6: Implement Data Retention Policy

```bash
# Set up TTL indexes for automatic data expiration

# Example: Automatically delete logs older than 90 days
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval '
  db.logs.createIndex(
    {createdAt: 1},
    {expireAfterSeconds: 7776000}  // 90 days
  )
'

# Verify TTL index created
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval '
  db.logs.getIndexes().filter(idx => idx.expireAfterSeconds)
'

# TTL monitor runs every 60 seconds and deletes expired documents
# Check TTL monitor statistics
kubectl exec -n mongodb mongodb-0 -- mongosh --eval '
  db.serverStatus().metrics.ttl
'
```

**Expected Time:** Immediate setup, ongoing automatic cleanup

### Option 7: Clean Up Journal and Diagnostic Data

```bash
# Check journal size
kubectl exec -n mongodb mongodb-0 -- du -sh /bitnami/mongodb/data/journal

# Journal can be large, but is necessary - only clean if MongoDB is stopped

# Clean old diagnostic data (safe to delete)
kubectl exec -n mongodb mongodb-0 -- rm -rf /bitnami/mongodb/data/diagnostic.data/*

# Check space freed
kubectl exec -n mongodb mongodb-0 -- df -h /bitnami/mongodb
```

## Post-Resolution Verification

### 1. Verify Disk Space Available

```bash
# Check current usage
kubectl exec -n mongodb mongodb-0 -- df -h /bitnami/mongodb

# Should show < 80% usage

# Monitor over time
watch -n 60 'kubectl exec -n mongodb mongodb-0 -- df -h /bitnami/mongodb'
```

### 2. Verify MongoDB Operations

```bash
# Test write operation
kubectl exec -n mongodb mongodb-0 -- mongosh test --eval '
  db.test_write.insert({timestamp: new Date(), test: "write_check"})
'

# Verify write succeeded
kubectl exec -n mongodb mongodb-0 -- mongosh test --eval '
  db.test_write.findOne({test: "write_check"})
'

# Clean up test data
kubectl exec -n mongodb mongodb-0 -- mongosh test --eval '
  db.test_write.deleteMany({test: "write_check"})
'
```

### 3. Verify Database Sizes

```bash
# Check that databases show expected sizes after cleanup
kubectl exec -n mongodb mongodb-0 -- mongosh --eval '
  db.adminCommand({listDatabases: 1}).databases.forEach(function(db) {
    print(db.name + ": " + (db.sizeOnDisk / 1024 / 1024 / 1024).toFixed(2) + " GB");
  });
'
```

### 4. Check Application Health

```bash
# Verify applications can connect and write
kubectl logs -n bruno deployment/agent-bruno --tail=50 | grep -i "mongodb\|error"

# Test from application
kubectl exec -n bruno deployment/agent-bruno -- curl -s http://localhost:8080/health
```

## Root Cause Analysis

### Common Causes

| Cause | Indicator | Prevention |
|-------|-----------|------------|
| No Data Retention Policy | Old data never deleted | Implement TTL indexes |
| Rapid Data Growth | Large increase in short time | Monitor growth rate, capacity planning |
| Insufficient Initial Size | PVC too small for workload | Right-size based on growth projections |
| Unnecessary Indexes | Many large indexes | Regular index review and cleanup |
| Large Documents | avgObjSize > 1MB | Schema optimization, use GridFS |
| Log/Audit Data | Accumulating logs | Implement log rotation or TTL |
| Backup Files Left Behind | Dump files in data directory | Clean up after backups |
| WiredTiger Bloat | High storageSize vs size | Regular compact operations |

### Investigation Commands

```bash
# Analyze storage growth over time (requires historical data)
# Check recent database size changes

# Calculate growth rate per database
kubectl exec -n mongodb mongodb-0 -- mongosh --eval '
  db.adminCommand({listDatabases: 1}).databases.forEach(function(database) {
    if (database.name !== "admin" && database.name !== "local") {
      var db = db.getSiblingDB(database.name);
      var collections = db.getCollectionNames();
      var totalDocs = 0;
      collections.forEach(function(coll) {
        totalDocs += db[coll].countDocuments();
      });
      print("\n" + database.name + ":");
      print("  Size: " + (database.sizeOnDisk / 1024 / 1024 / 1024).toFixed(2) + " GB");
      print("  Collections: " + collections.length);
      print("  Documents: " + totalDocs);
    }
  });
'
```

## Prevention

### 1. Implement Data Lifecycle Management

```javascript
// TTL indexes for automatic expiration
db.logs.createIndex(
  {createdAt: 1},
  {expireAfterSeconds: 7776000}  // 90 days
);

db.sessions.createIndex(
  {lastAccess: 1},
  {expireAfterSeconds: 3600}  // 1 hour
);

db.events.createIndex(
  {timestamp: 1},
  {expireAfterSeconds: 2592000}  // 30 days
);
```

### 2. Capacity Planning

```yaml
# Calculate required storage:
# Base formula: (current_size / days_of_data) * retention_days * growth_factor * 1.5

# Example calculation:
# Current: 20GB for 30 days of data
# Retention: 90 days
# Growth factor: 1.2 (20% year-over-year growth)
# Safety margin: 1.5x

# Required: (20GB / 30) * 90 * 1.2 * 1.5 = 108GB
# Round up: 120GB PVC

persistence:
  size: 120Gi  # Based on capacity planning
```

### 3. Monitoring and Alerts

```yaml
# Prometheus alerts
groups:
  - name: mongodb-storage
    rules:
      - alert: MongoDBStorageUsageHigh
        expr: |
          (
            kubelet_volume_stats_used_bytes{persistentvolumeclaim="data-mongodb-0"} /
            kubelet_volume_stats_capacity_bytes{persistentvolumeclaim="data-mongodb-0"}
          ) > 0.85
        for: 5m
        annotations:
          summary: "MongoDB storage usage > 85%"
          
      - alert: MongoDBStorageGrowthHigh
        expr: |
          predict_linear(
            kubelet_volume_stats_used_bytes{persistentvolumeclaim="data-mongodb-0"}[7d],
            7*24*3600
          ) > kubelet_volume_stats_capacity_bytes{persistentvolumeclaim="data-mongodb-0"}
        for: 1h
        annotations:
          summary: "MongoDB storage will be full in 7 days at current growth rate"
          
      - alert: MongoDBDatabaseSizeAnomal y
        expr: |
          rate(mongodb_db_data_size_bytes[1h]) > 
          avg_over_time(rate(mongodb_db_data_size_bytes[7d])[1h:]) * 2
        for: 2h
        annotations:
          summary: "MongoDB database growing 2x faster than normal"
```

### 4. Regular Maintenance Schedule

```bash
# Weekly maintenance script
cat << 'EOF' > /tmp/storage-maintenance.sh
#!/bin/bash
echo "=== MongoDB Storage Maintenance ==="
echo "Date: $(date)"

echo -e "\n1. Current Disk Usage:"
kubectl exec -n mongodb mongodb-0 -- df -h /bitnami/mongodb

echo -e "\n2. Database Sizes:"
kubectl exec -n mongodb mongodb-0 -- mongosh --quiet --eval '
  db.adminCommand({listDatabases: 1}).databases.forEach(function(db) {
    print(db.name + ": " + (db.sizeOnDisk / 1024 / 1024 / 1024).toFixed(2) + " GB");
  });
'

echo -e "\n3. Top 5 Largest Collections:"
kubectl exec -n mongodb mongodb-0 -- mongosh --quiet myapp --eval '
  var collections = [];
  db.getCollectionNames().forEach(function(name) {
    var size = db[name].stats().storageSize;
    collections.push({name: name, size: size});
  });
  collections.sort((a, b) => b.size - a.size).slice(0, 5).forEach(function(c) {
    print(c.name + ": " + (c.size / 1024 / 1024).toFixed(2) + " MB");
  });
'

echo -e "\n4. Storage Growth (last 7 days):"
# Requires metrics history
# Query Prometheus for growth rate

echo -e "\n5. TTL Index Status:"
kubectl exec -n mongodb mongodb-0 -- mongosh --quiet myapp --eval '
  db.getCollectionNames().forEach(function(coll) {
    var ttlIndexes = db[coll].getIndexes().filter(idx => idx.expireAfterSeconds);
    if (ttlIndexes.length > 0) {
      print(coll + " has TTL indexes:");
      ttlIndexes.forEach(idx => print("  - " + idx.name + ": " + idx.expireAfterSeconds + "s"));
    }
  });
'
EOF

chmod +x /tmp/storage-maintenance.sh
# Run weekly via cron
```

### 5. Schema Optimization

```javascript
// Best practices to minimize storage usage

// 1. Use appropriate data types
// BAD: storing numbers as strings
{price: "19.99"}
// GOOD:
{price: 19.99}

// 2. Avoid storing computed values
// BAD: storing what can be calculated
{quantity: 10, price: 5, total: 50}
// GOOD:
{quantity: 10, price: 5}

// 3. Use shorter field names for large collections
// BAD:
{customerIdentifier: "123", orderTimestamp: new Date()}
// GOOD:
{cid: "123", ts: new Date()}

// 4. Normalize large embedded arrays
// BAD: {user: {...}, orders: [...10000 orders...]}
// GOOD: Split into users and orders collections

// 5. Use GridFS for large files
// Don't store files > 16MB as documents
db.fs.files.insert({...})
```

## Escalation

**Escalate if:**
- Cannot expand PVC (storage limit reached)
- Data deletion requires business approval
- Storage fills up faster than cleanup
- Need to migrate to larger cluster

**Escalation Path:**
1. **Immediate**: Implement Option 1 (expand PVC)
2. **30 minutes**: Engage database team for data cleanup
3. **1 hour**: Engage capacity planning for long-term solution
4. **4 hours**: Consider data archival or migration

**Communication Template:**
```
INCIDENT: MongoDB Storage Critically Low
SEVERITY: Critical
CURRENT USAGE: [X]GB / [Y]GB ([Z]%)
GROWTH RATE: [rate]GB/day
TIME TO FULL: [estimated time]
LARGEST DATABASES:
- [db1]: [size]GB
- [db2]: [size]GB
ACTIONS TAKEN:
- PVC expanded: [yes/no] [old_size] -> [new_size]
- Data deleted: [yes/no] [amount]
- Compaction run: [yes/no]
STATUS: [Space available/Still critical]
ROOT CAUSE: [If identified]
LONG-TERM SOLUTION: [Plan]
```

## Related Runbooks

- [MongoDB Service Down](./mongodb-down.md)
- [MongoDB High Memory](./mongodb-high-memory.md)
- [MongoDB Slow Queries](./mongodb-slow-queries.md)

## Additional Resources

- [MongoDB Storage](https://docs.mongodb.com/manual/faq/storage/)
- [WiredTiger Storage Engine](https://docs.mongodb.com/manual/core/wiredtiger/)
- [Compact Command](https://docs.mongodb.com/manual/reference/command/compact/)
- [TTL Indexes](https://docs.mongodb.com/manual/core/index-ttl/)

---

**Last Updated**: 2025-10-15  
**Version**: 1.0  
**Maintainer**: Homelab Platform Team

