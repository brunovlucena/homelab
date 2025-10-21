# 🚨 Runbook: MongoDB Replication Lag

## Alert Information

**Alert Name:** `MongoDBReplicationLag`  
**Severity:** Warning  
**Component:** mongodb  
**Service:** mongodb  
**Threshold:** Replication lag > 60 seconds (if replica set enabled)

## Symptom

Secondary MongoDB replica members are falling behind the primary, causing stale reads and potential data consistency issues.

## Impact

- **User Impact:** MEDIUM - Stale data on reads from secondaries
- **Business Impact:** MEDIUM - Data inconsistency, potential failover issues
- **Data Impact:** MEDIUM - Temporary data inconsistency across replicas

## Note

⚠️ **Current Configuration**: The homelab MongoDB deployment runs in **standalone mode** (single instance, no replication). This runbook is included for reference in case the architecture is upgraded to a replica set in the future.

If you're seeing replication-related alerts with a standalone instance, verify the alerting configuration.

## Diagnosis

### 1. Verify Replication Status

```bash
# Check if running as replica set
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'rs.status()' 2>&1 | head -10

# If standalone, you'll see: "MongoServerError: not running with --replSet"
# If replica set, you'll see detailed status

# For replica set, check replication info
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'rs.printReplicationInfo()'

# Check secondary lag
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'rs.printSecondaryReplicationInfo()'
```

### 2. Check Replica Set Members

```bash
# List all replica set members
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'rs.status().members' | jq

# Check member states
kubectl exec -n mongodb mongodb-0 -- mongosh --eval '
  rs.status().members.forEach(function(member) {
    print(member.name + " - State: " + member.stateStr + ", Health: " + member.health);
  })'

# Identify primary and secondaries
kubectl exec -n mongodb mongodb-0 -- mongosh --eval '
  var status = rs.status();
  var primary = status.members.find(m => m.state === 1);
  var secondaries = status.members.filter(m => m.state === 2);
  print("Primary: " + primary.name);
  print("Secondaries: " + secondaries.length);
  secondaries.forEach(s => print("  - " + s.name + " (lag: " + (status.date - s.optimeDate)/1000 + "s)"));
'
```

### 3. Check Oplog Status

```bash
# Check oplog size and usage
kubectl exec -n mongodb mongodb-0 -- mongosh --eval '
  var oplog = db.getSiblingDB("local").oplog.rs.stats();
  print("Oplog size: " + (oplog.maxSize / 1024 / 1024).toFixed(2) + " MB");
  print("Oplog used: " + (oplog.size / 1024 / 1024).toFixed(2) + " MB");
  print("Usage: " + ((oplog.size / oplog.maxSize) * 100).toFixed(2) + "%");
'

# Check oplog time range
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'rs.printReplicationInfo()'

# Output shows:
# - configured oplog size
# - log length start to end
# - oplog first event time
# - oplog last event time
# - now
```

### 4. Check Replication Lag Details

```bash
# Detailed lag information per secondary
kubectl exec -n mongodb mongodb-0 -- mongosh --eval '
  var status = rs.status();
  var primary = status.members.find(m => m.state === 1);
  
  status.members.filter(m => m.state === 2).forEach(function(secondary) {
    var lag = (primary.optime.ts.t - secondary.optime.ts.t);
    print("\nSecondary: " + secondary.name);
    print("  State: " + secondary.stateStr);
    print("  Lag: " + lag + " seconds");
    print("  Last heartbeat: " + secondary.lastHeartbeat);
    print("  Ping: " + secondary.pingMs + "ms");
  });
'
```

### 5. Check Write Load on Primary

```bash
# Check write operations per second
kubectl exec -n mongodb mongodb-0 -- mongosh --eval '
  var before = db.serverStatus().opcounters;
  print("Waiting 10 seconds...");
' && sleep 10 && kubectl exec -n mongodb mongodb-0 -- mongosh --eval '
  var after = db.serverStatus().opcounters;
  print("Inserts/sec: " + ((after.insert - before.insert) / 10));
  print("Updates/sec: " + ((after.update - before.update) / 10));
  print("Deletes/sec: " + ((after.delete - before.delete) / 10));
'

# Check replication queue
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.serverStatus().metrics.repl'
```

### 6. Check Network and Resource Issues

```bash
# Check network latency between members
# From primary to secondaries
kubectl exec -n mongodb mongodb-0 -- mongosh --eval '
  rs.status().members.filter(m => m.state === 2).forEach(function(member) {
    print(member.name + " - Ping: " + member.pingMs + "ms");
  });
'

# Check secondary resource usage
kubectl top pods -n mongodb

# Check for slow disk I/O
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.serverStatus().wiredTiger'
```

## Resolution

### Option 1: Increase Oplog Size

```bash
# Check current oplog size
kubectl exec -n mongodb mongodb-0 -- mongosh --eval '
  db.getSiblingDB("local").oplog.rs.stats().maxSize / 1024 / 1024 / 1024
'

# Resize oplog (requires primary, can be done online in MongoDB 4.0+)
kubectl exec -n mongodb mongodb-0 -- mongosh admin --eval '
  db.adminCommand({replSetResizeOplog: 1, size: 10240})  // 10GB
'

# Verify new size
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'rs.printReplicationInfo()'
```

**Expected Time:** Immediate, but secondary catch-up depends on lag

### Option 2: Resync Lagging Secondary

```bash
# Identify the lagging secondary
# Example: mongodb-2 is lagging

# Option A: Force resync (data wipe and full sync)
kubectl exec -n mongodb mongodb-2 -- mongosh admin --eval '
  db.shutdownServer()
'

# Remove data directory
kubectl exec -n mongodb mongodb-2 -- rm -rf /bitnami/mongodb/data/*

# Restart - will automatically resync from primary
kubectl delete pod -n mongodb mongodb-2

# Monitor resync progress
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'rs.printSecondaryReplicationInfo()'

# Option B: Roll back and resync using replSetSyncFrom
kubectl exec -n mongodb mongodb-2 -- mongosh admin --eval '
  rs.syncFrom("mongodb-0.mongodb.mongodb.svc.cluster.local:27017")
'
```

**Expected Time:** Minutes to hours depending on data size

### Option 3: Optimize Write Load

```bash
# If high write load is causing lag:

# 1. Use write concern "majority" instead of all secondaries
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval '
  db.users.insert({name: "test"}, {writeConcern: {w: "majority"}})
'

# 2. Batch writes instead of individual operations
# Application code change required

# 3. Use bulk operations
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval '
  var bulk = db.users.initializeUnorderedBulkOp();
  bulk.insert({name: "user1"});
  bulk.insert({name: "user2"});
  bulk.execute();
'

# 4. Temporarily stop writes to allow catch-up (emergency only)
kubectl exec -n mongodb mongodb-0 -- mongosh admin --eval '
  db.fsyncLock()  // Blocks writes
'
# Wait for secondaries to catch up, then:
kubectl exec -n mongodb mongodb-0 -- mongosh admin --eval '
  db.fsyncUnlock()
'
```

### Option 4: Increase Secondary Resources

```bash
# If secondary is resource-constrained

# Check current resources
kubectl get pod -n mongodb mongodb-2 -o json | jq '.spec.containers[0].resources'

# Increase CPU/Memory for secondary
kubectl edit statefulset -n mongodb mongodb

# Update resources for the container:
# resources:
#   limits:
#     cpu: 2000m
#     memory: 4Gi
#   requests:
#     cpu: 1000m
#     memory: 2Gi

# Rolling update will apply changes
kubectl rollout status statefulset -n mongodb mongodb
```

### Option 5: Add Read Preference to Reduce Secondary Load

```bash
# Configure applications to prefer primary for reads
# This reduces load on lagging secondaries

# Connection string update:
# mongodb://mongodb-0:27017,mongodb-1:27017,mongodb-2:27017/myapp?
#   replicaSet=rs0&
#   readPreference=primaryPreferred

# Or use readConcern for specific queries:
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval '
  db.users.find().readPref("primaryPreferred")
'
```

### Option 6: Check and Fix Network Issues

```bash
# Test network connectivity between members
kubectl exec -n mongodb mongodb-0 -- nc -zv mongodb-1.mongodb.mongodb.svc.cluster.local 27017
kubectl exec -n mongodb mongodb-0 -- nc -zv mongodb-2.mongodb.mongodb.svc.cluster.local 27017

# Check for network policies blocking replication
kubectl get networkpolicies -n mongodb

# Verify DNS resolution
kubectl exec -n mongodb mongodb-0 -- nslookup mongodb-1.mongodb.mongodb.svc.cluster.local
kubectl exec -n mongodb mongodb-0 -- nslookup mongodb-2.mongodb.mongodb.svc.cluster.local
```

## Post-Resolution Verification

### 1. Verify Replication Lag Resolved

```bash
# Check current lag
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'rs.printSecondaryReplicationInfo()'

# Should show lag < 10 seconds for all secondaries

# Monitor lag over time
for i in {1..6}; do
  echo "=== Check $i ==="
  kubectl exec -n mongodb mongodb-0 -- mongosh --eval '
    var status = rs.status();
    var primary = status.members.find(m => m.state === 1);
    status.members.filter(m => m.state === 2).forEach(function(s) {
      var lag = (status.date - s.optimeDate) / 1000;
      print(s.name + " lag: " + lag + "s");
    });
  '
  sleep 10
done
```

### 2. Verify All Members Healthy

```bash
# Check replica set health
kubectl exec -n mongodb mongodb-0 -- mongosh --eval '
  rs.status().members.forEach(function(m) {
    print(m.name + " - State: " + m.stateStr + ", Health: " + m.health);
  });
'

# All members should show:
# - State: PRIMARY or SECONDARY
# - Health: 1
```

### 3. Verify Oplog Coverage

```bash
# Ensure oplog covers sufficient time window
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'rs.printReplicationInfo()'

# Oplog should cover at least several hours (ideally 24h+)
# "hours of oplog" should be > 24
```

### 4. Test Read Operations on Secondaries

```bash
# Connect to secondary and test read
kubectl exec -n mongodb mongodb-1 -- mongosh myapp --eval '
  db.getMongo().setReadPref("secondary");
  db.users.findOne();
'

# Should return data without errors
```

## Root Cause Analysis

### Common Causes

| Cause | Indicator | Solution |
|-------|-----------|----------|
| High Write Load | High insert/update/delete rate | Batch operations, optimize writes |
| Small Oplog | Oplog hours < 12 | Increase oplog size |
| Slow Secondary | High CPU/memory on secondary | Increase resources |
| Network Issues | High ping times between members | Fix network, check policies |
| Initial Sync | Member shows STARTUP2 state | Wait for sync to complete |
| Inefficient Indexes | Slow index builds on secondary | Build indexes with background:true |
| Secondary Disk I/O | High disk latency | Use faster storage, check disk |

### Investigation Commands

```bash
# Detailed replication metrics
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.serverStatus().metrics.repl' | jq

# Check oplog entries per second
kubectl exec -n mongodb mongodb-0 -- mongosh --eval '
  var before = db.getSiblingDB("local").oplog.rs.count();
  print("Waiting 10 seconds...");
' && sleep 10 && kubectl exec -n mongodb mongodb-0 -- mongosh --eval '
  var after = db.getSiblingDB("local").oplog.rs.count();
  print("Oplog entries/sec: " + ((after - before) / 10));
'

# Analyze oplog content
kubectl exec -n mongodb mongodb-0 -- mongosh local --eval '
  db.oplog.rs.find().sort({$natural: -1}).limit(10).forEach(printjson)
'
```

## Prevention

### 1. Proper Oplog Sizing

```yaml
# Formula: Oplog size should accommodate peak write load for 24+ hours
# Recommended: 5-10% of total storage

# For Helm configuration:
replicaSetConfigurationSettings:
  enabled: true
  configuration: |
    replication:
      oplogSizeMB: 10240  # 10GB for moderate write load
```

### 2. Write Concern Configuration

```javascript
// Use appropriate write concerns in application

// For critical data
db.users.insert(
  {name: "critical"},
  {writeConcern: {w: "majority", j: true, wtimeout: 5000}}
);

// For less critical data (better performance)
db.logs.insert(
  {message: "log entry"},
  {writeConcern: {w: 1, j: false}}
);
```

### 3. Read Preference Strategy

```javascript
// Distribute read load appropriately

// Critical reads - always from primary
db.orders.find().readPref("primary");

// Analytics/reporting - can use secondaries
db.analytics.find().readPref("secondary");

// Balance between performance and consistency
db.users.find().readPref("primaryPreferred");
```

### 4. Monitoring and Alerts

```yaml
# Prometheus alerts for replication
groups:
  - name: mongodb-replication
    rules:
      - alert: MongoDBReplicationLag
        expr: |
          mongodb_replset_member_replication_lag_seconds > 60
        for: 5m
        annotations:
          summary: "MongoDB replication lag > 60 seconds"
          
      - alert: MongoDBReplicationHeartbeatFailed
        expr: |
          mongodb_replset_member_health == 0
        for: 1m
        annotations:
          summary: "MongoDB replica member health check failed"
          
      - alert: MongoDBOplogWindowLow
        expr: |
          mongodb_replset_oplog_head_timestamp - mongodb_replset_oplog_tail_timestamp < 43200
        for: 10m
        annotations:
          summary: "MongoDB oplog window < 12 hours"
```

### 5. Regular Maintenance

```bash
# Weekly replication health check
cat << 'EOF' > /tmp/replication-check.sh
#!/bin/bash
echo "=== MongoDB Replication Health Check ==="
echo "Date: $(date)"
echo

echo "1. Replica Set Status:"
kubectl exec -n mongodb mongodb-0 -- mongosh --quiet --eval '
  rs.status().members.forEach(function(m) {
    print(m.name + " - " + m.stateStr + " (health: " + m.health + ")");
  });
'

echo -e "\n2. Replication Lag:"
kubectl exec -n mongodb mongodb-0 -- mongosh --quiet --eval 'rs.printSecondaryReplicationInfo()'

echo -e "\n3. Oplog Status:"
kubectl exec -n mongodb mongodb-0 -- mongosh --quiet --eval 'rs.printReplicationInfo()'

echo -e "\n4. Write Load (last 10s):"
kubectl exec -n mongodb mongodb-0 -- mongosh --quiet --eval '
  var before = db.serverStatus().opcounters;
' && sleep 10 && kubectl exec -n mongodb mongodb-0 -- mongosh --quiet --eval '
  var after = db.serverStatus().opcounters;
  print("Inserts/sec: " + ((after.insert - before.insert) / 10));
  print("Updates/sec: " + ((after.update - before.update) / 10));
  print("Deletes/sec: " + ((after.delete - before.delete) / 10));
'
EOF

chmod +x /tmp/replication-check.sh
```

## Escalation

**Escalate if:**
- Replication lag continues to grow
- Secondary cannot catch up after several hours
- Data inconsistency detected
- Need to add new replica members

**Escalation Path:**
1. **30 minutes**: Engage database team
2. **1 hour**: Consider emergency oplog resize
3. **2 hours**: Evaluate replica set architecture
4. **4 hours**: Consider temporary removal of lagging secondary

**Communication Template:**
```
ISSUE: MongoDB Replication Lag
SEVERITY: Warning
PRIMARY: [hostname]
LAGGING SECONDARY: [hostname]
CURRENT LAG: [X] seconds
OPLOG WINDOW: [Y] hours
SYMPTOMS: [Stale reads/Performance degradation/etc]
ACTIONS TAKEN:
- Checked oplog size: [size]
- Increased resources: [yes/no]
- Attempted resync: [yes/no]
ROOT CAUSE: [If identified]
STATUS: [Improving/Not improving]
ESTIMATED CATCH-UP TIME: [time]
```

## Related Runbooks

- [MongoDB Service Down](./mongodb-down.md)
- [MongoDB High Memory](./mongodb-high-memory.md)
- [MongoDB Storage Full](./mongodb-storage-full.md)

## Additional Resources

- [MongoDB Replication](https://docs.mongodb.com/manual/replication/)
- [Oplog Sizing](https://docs.mongodb.com/manual/core/replica-set-oplog/)
- [Replication Lag](https://docs.mongodb.com/manual/tutorial/troubleshoot-replica-sets/#check-the-replication-lag)
- [Read Preference](https://docs.mongodb.com/manual/core/read-preference/)

---

**Last Updated**: 2025-10-15  
**Version**: 1.0  
**Maintainer**: Homelab Platform Team

