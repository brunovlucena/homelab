# 🚨 Runbook: MongoDB High Connections

## Alert Information

**Alert Name:** `MongoDBHighConnections`  
**Severity:** Warning  
**Component:** mongodb  
**Service:** mongodb  
**Threshold:** Active connections > 80% of maxIncomingConnections

## Symptom

MongoDB connection count is approaching or has exceeded configured limits, potentially causing connection exhaustion and application connection failures.

## Impact

- **User Impact:** MEDIUM-HIGH - Application unable to connect, timeouts, errors
- **Business Impact:** HIGH - Service degradation or outage for new requests
- **Data Impact:** LOW - No data loss, but reduced availability

## Diagnosis

### 1. Check Current Connection Status

```bash
# Get connection statistics
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.serverStatus().connections'

# Output shows:
# - current: Active connections
# - available: Available connection slots
# - totalCreated: Total connections created since startup
```

### 2. Check Connection Limit

```bash
# Check max connections setting
kubectl exec -n mongodb mongodb-0 -- mongosh admin --eval \
  'db.runCommand({getParameter: 1, maxIncomingConnections: 1})'

# Default is usually 65536, but can be lower based on system resources

# Calculate connection usage percentage
kubectl exec -n mongodb mongodb-0 -- mongosh --eval \
  'var conn = db.serverStatus().connections;
   var maxConn = db.adminCommand({getParameter: 1, maxIncomingConnections: 1}).maxIncomingConnections;
   print("Current: " + conn.current);
   print("Available: " + conn.available);
   print("Max: " + maxConn);
   print("Usage: " + ((conn.current / maxConn) * 100).toFixed(2) + "%")'
```

### 3. Identify Connection Sources

```bash
# List current connections by client
kubectl exec -n mongodb mongodb-0 -- mongosh --eval \
  'db.currentOp(true).inprog.forEach(function(op) {
     if (op.client) print(op.client);
   })' | sort | uniq -c | sort -rn

# Get detailed connection info
kubectl exec -n mongodb mongodb-0 -- mongosh admin --eval \
  'db.aggregate([
     {$currentOp: {allUsers: true, idleConnections: true}},
     {$group: {_id: "$client", count: {$sum: 1}}},
     {$sort: {count: -1}},
     {$limit: 10}
   ])'
```

### 4. Check for Connection Leaks

```bash
# Check idle connections (potential leaks)
kubectl exec -n mongodb mongodb-0 -- mongosh admin --eval \
  'db.aggregate([
     {$currentOp: {allUsers: true, idleConnections: true}},
     {$match: {active: false}},
     {$group: {
       _id: "$client",
       count: {$sum: 1},
       avgIdleTime: {$avg: {$subtract: [new Date(), "$connectionStarted"]}}
     }},
     {$sort: {count: -1}},
     {$limit: 10}
   ])'

# Check connection duration
kubectl exec -n mongodb mongodb-0 -- mongosh admin --eval \
  'db.aggregate([
     {$currentOp: {allUsers: true, idleConnections: true}},
     {$project: {
       client: 1,
       duration: {$divide: [
         {$subtract: [new Date(), "$connectionStarted"]},
         1000
       ]}
     }},
     {$group: {
       _id: "$client",
       avgDuration: {$avg: "$duration"},
       maxDuration: {$max: "$duration"}
     }},
     {$sort: {avgDuration: -1}},
     {$limit: 10}
   ])'
```

### 5. Check Application Connection Pools

```bash
# Check agent-bruno connections
kubectl logs -n bruno deployment/agent-bruno --tail=100 | grep -i "mongo\|connection"

# Check if applications are properly closing connections
kubectl logs -n bruno deployment/agent-bruno --tail=500 | grep -i "pool\|timeout\|connection"

# Check application connection pool settings
kubectl get deployment -n bruno agent-bruno -o yaml | grep -A5 -i mongo
```

### 6. Check System Resources

```bash
# Check if connection limits are due to file descriptor limits
kubectl exec -n mongodb mongodb-0 -- sh -c 'ulimit -n'

# Check current file descriptor usage
kubectl exec -n mongodb mongodb-0 -- sh -c 'ls -1 /proc/self/fd | wc -l'

# Check system-wide connection count
kubectl exec -n mongodb mongodb-0 -- netstat -an | grep :27017 | wc -l
```

## Resolution

### Option 1: Restart Leaking Applications

```bash
# Identify the application with most connections
kubectl exec -n mongodb mongodb-0 -- mongosh --eval \
  'db.currentOp(true).inprog.forEach(function(op) {
     if (op.client) print(op.client);
   })' | sort | uniq -c | sort -rn | head -5

# Restart the application with most connections
# Example: agent-bruno has many connections
kubectl rollout restart deployment -n bruno agent-bruno

# Monitor connection count decrease
watch -n 5 'kubectl exec -n mongodb mongodb-0 -- mongosh --eval "db.serverStatus().connections.current"'
```

**Expected Time:** 2-3 minutes

### Option 2: Kill Idle Connections

```bash
# Kill connections idle for > 30 minutes
kubectl exec -n mongodb mongodb-0 -- mongosh admin --eval '
  db.aggregate([
    {$currentOp: {allUsers: true, idleConnections: true}},
    {$match: {
      active: false,
      connectionStarted: {$lt: new Date(Date.now() - 30*60*1000)}
    }},
    {$project: {opid: 1}}
  ]).forEach(function(op) {
    print("Killing connection: " + op.opid);
    db.killOp(op.opid);
  })'

# Kill connections from a specific client (use with caution)
kubectl exec -n mongodb mongodb-0 -- mongosh admin --eval '
  var targetClient = "10.244.0.50:12345";
  db.aggregate([
    {$currentOp: {allUsers: true, idleConnections: true}},
    {$match: {client: {$regex: targetClient}}},
    {$project: {opid: 1}}
  ]).forEach(function(op) {
    print("Killing connection: " + op.opid);
    db.killOp(op.opid);
  })'
```

### Option 3: Increase Connection Limit

```bash
# Increase max connections (temporary - resets on restart)
kubectl exec -n mongodb mongodb-0 -- mongosh admin --eval \
  'db.runCommand({setParameter: 1, maxIncomingConnections: 100000})'

# Permanent increase via Helm values
kubectl edit helmrelease -n mongodb mongodb

# Add under values:
# configuration: |
#   net:
#     maxIncomingConnections: 100000

# Reconcile
flux reconcile helmrelease -n mongodb mongodb

# Verify new limit
kubectl exec -n mongodb mongodb-0 -- mongosh admin --eval \
  'db.runCommand({getParameter: 1, maxIncomingConnections: 1})'
```

**Note:** Increasing limits is not a permanent solution if there's a connection leak

### Option 4: Configure Application Connection Pools

```bash
# Update application connection string parameters
# Example for agent-bruno

# Get current configuration
kubectl get deployment -n bruno agent-bruno -o yaml > /tmp/agent-bruno.yaml

# Update MongoDB connection string to include pool settings:
# mongodb://mongodb.mongodb.svc.cluster.local:27017/myapp?
#   maxPoolSize=50&           # Max connections per host
#   minPoolSize=10&           # Min connections to maintain
#   maxIdleTimeMS=60000&      # Close idle connections after 60s
#   waitQueueTimeoutMS=5000&  # Timeout waiting for connection
#   serverSelectionTimeoutMS=5000

# Apply updated configuration
kubectl apply -f /tmp/agent-bruno.yaml

# Restart to apply new connection string
kubectl rollout restart deployment -n bruno agent-bruno
```

### Option 5: Implement Connection Pooling Middleware

```yaml
# Example: Update application environment variables
apiVersion: apps/v1
kind: Deployment
metadata:
  name: agent-bruno
  namespace: bruno
spec:
  template:
    spec:
      containers:
      - name: app
        env:
        - name: MONGODB_URI
          value: "mongodb://mongodb.mongodb.svc.cluster.local:27017/myapp"
        - name: MONGODB_MAX_POOL_SIZE
          value: "50"
        - name: MONGODB_MIN_POOL_SIZE
          value: "10"
        - name: MONGODB_MAX_IDLE_TIME_MS
          value: "60000"
        - name: MONGODB_CONNECTION_TIMEOUT_MS
          value: "5000"
```

### Option 6: Scale MongoDB (Advanced)

```bash
# If standalone MongoDB cannot handle connection load,
# consider implementing a replica set for connection distribution

# This requires architecture change - escalate to database team
# See: https://docs.mongodb.com/manual/tutorial/deploy-replica-set/
```

## Post-Resolution Verification

### 1. Verify Connection Count

```bash
# Monitor connection count for 5 minutes
watch -n 10 'kubectl exec -n mongodb mongodb-0 -- mongosh --eval "
  var conn = db.serverStatus().connections;
  print(\"Current: \" + conn.current);
  print(\"Available: \" + conn.available);
  print(\"Usage: \" + ((conn.current / (conn.current + conn.available)) * 100).toFixed(2) + \"%\");
"'
```

### 2. Verify Application Health

```bash
# Check application logs for connection errors
kubectl logs -n bruno deployment/agent-bruno --tail=100 | grep -i "connection\|pool\|timeout"

# Test application connectivity
kubectl exec -n bruno deployment/agent-bruno -- curl -s http://localhost:8080/health

# Verify no connection timeout errors
kubectl get events -n bruno | grep -i "connection\|timeout"
```

### 3. Verify Connection Pool Efficiency

```bash
# Check connection creation rate (should be low if pooling works)
kubectl exec -n mongodb mongodb-0 -- mongosh --eval '
  var conn = db.serverStatus().connections;
  print("Total created since startup: " + conn.totalCreated);
  print("Current: " + conn.current);
  print("Churn rate: " + (conn.totalCreated / conn.current).toFixed(2));
'
# Low churn rate (<10) is good, indicates stable pool
```

### 4. Load Test (Optional)

```bash
# Test connection handling under load
kubectl run mongo-loadtest --image=mongo:7 --rm -it --restart=Never -- \
  bash -c 'for i in {1..100}; do
    mongosh mongodb://mongodb.mongodb.svc.cluster.local:27017/test \
      --eval "db.runCommand({ping: 1})" &
  done; wait'

# Monitor connections during test
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.serverStatus().connections'
```

## Root Cause Analysis

### Common Causes

| Cause | Indicator | Solution |
|-------|-----------|----------|
| Connection Leak | Steady increase over time | Fix application code, restart apps |
| No Connection Pooling | High totalCreated count | Implement connection pooling |
| Large Connection Pool | Many connections from single app | Reduce maxPoolSize |
| Not Closing Connections | Many idle connections | Fix application code |
| Multiple App Instances | Proportional to instance count | Reduce pool size per instance |
| Connection Timeout Too High | Long-lived idle connections | Reduce maxIdleTimeMS |
| Insufficient Limit | Usage at 100% | Increase maxIncomingConnections |

### Investigation Commands

```bash
# Calculate connection growth rate
kubectl exec -n mongodb mongodb-0 -- mongosh --eval '
  var start = db.serverStatus().connections.current;
  print("Starting connections: " + start);
  print("Waiting 60 seconds...");
' && sleep 60 && kubectl exec -n mongodb mongodb-0 -- mongosh --eval '
  var end = db.serverStatus().connections.current;
  print("Ending connections: " + end);
  print("Growth rate: " + (end - start) + " connections/minute");
'

# Analyze connection patterns by time
kubectl logs -n mongodb mongodb-0 --tail=1000 | \
  grep "connection accepted" | \
  awk '{print $1, $2}' | \
  uniq -c

# Check for connection storms
kubectl logs -n mongodb mongodb-0 --since=10m | \
  grep -c "connection accepted"
```

## Prevention

### 1. Optimal Connection Pool Configuration

```javascript
// Application-side best practices

// Node.js (Mongoose)
mongoose.connect('mongodb://mongodb:27017/myapp', {
  maxPoolSize: 50,           // Max connections
  minPoolSize: 10,           // Min connections to keep open
  maxIdleTimeMS: 60000,      // Close after 60s idle
  socketTimeoutMS: 45000,    // Socket timeout
  serverSelectionTimeoutMS: 5000,
  family: 4                  // Use IPv4
});

// Python (PyMongo)
client = MongoClient(
    'mongodb://mongodb:27017/',
    maxPoolSize=50,
    minPoolSize=10,
    maxIdleTimeMS=60000,
    serverSelectionTimeoutMS=5000
)

// Go (Official Driver)
clientOpts := options.Client().
    ApplyURI("mongodb://mongodb:27017").
    SetMaxPoolSize(50).
    SetMinPoolSize(10).
    SetMaxConnIdleTime(60 * time.Second)
```

### 2. Connection Pool Sizing Formula

```
Optimal Pool Size = (concurrent_requests * avg_query_time_ms) / 1000

Example:
- 100 concurrent requests
- 50ms average query time
- Optimal pool size = (100 * 50) / 1000 = 5 connections

Add buffer: 5 * 2 = 10 connections per instance

For 5 application instances: 10 connections each = 50 total
Set maxPoolSize = 15 per instance (50% buffer)
```

### 3. Monitoring and Alerts

```yaml
# Prometheus alerts
groups:
  - name: mongodb-connections
    rules:
      - alert: MongoDBHighConnections
        expr: |
          mongodb_connections{state="current"} /
          (mongodb_connections{state="current"} + mongodb_connections{state="available"}) > 0.8
        for: 5m
        annotations:
          summary: "MongoDB using > 80% of available connections"
          
      - alert: MongoDBConnectionChurn
        expr: rate(mongodb_connections{state="totalCreated"}[10m]) > 10
        for: 10m
        annotations:
          summary: "High connection creation rate indicates poor pooling"
          
      - alert: MongoDBConnectionLeak
        expr: |
          deriv(mongodb_connections{state="current"}[30m]) > 1
        for: 1h
        annotations:
          summary: "MongoDB connections steadily increasing"
```

### 4. Application Health Checks

```yaml
# Implement proper connection lifecycle management

# Python example
class MongoConnectionManager:
    def __init__(self):
        self.client = None
        self.max_retries = 3
        
    def connect(self):
        for attempt in range(self.max_retries):
            try:
                self.client = MongoClient(
                    uri,
                    maxPoolSize=50,
                    minPoolSize=10,
                    maxIdleTimeMS=60000
                )
                # Test connection
                self.client.admin.command('ping')
                return True
            except Exception as e:
                logger.error(f"Connection attempt {attempt + 1} failed: {e}")
                time.sleep(2 ** attempt)
        return False
    
    def close(self):
        if self.client:
            self.client.close()
            self.client = None

# Ensure proper cleanup on shutdown
import atexit
mongo_manager = MongoConnectionManager()
atexit.register(mongo_manager.close)
```

### 5. Regular Connection Audits

```bash
# Weekly connection audit script
cat << 'EOF' > /tmp/connection-audit.sh
#!/bin/bash
echo "=== MongoDB Connection Audit ==="
echo "Date: $(date)"
echo

echo "1. Current Connections:"
kubectl exec -n mongodb mongodb-0 -- mongosh --quiet --eval 'db.serverStatus().connections'

echo -e "\n2. Connections by Client (Top 10):"
kubectl exec -n mongodb mongodb-0 -- mongosh --quiet admin --eval '
  db.aggregate([
    {$currentOp: {allUsers: true, idleConnections: true}},
    {$group: {_id: "$client", count: {$sum: 1}}},
    {$sort: {count: -1}},
    {$limit: 10}
  ]).forEach(function(doc) {
    print(doc._id + ": " + doc.count);
  })'

echo -e "\n3. Idle Connections (> 10 minutes):"
kubectl exec -n mongodb mongodb-0 -- mongosh --quiet admin --eval '
  db.aggregate([
    {$currentOp: {allUsers: true, idleConnections: true}},
    {$match: {
      active: false,
      connectionStarted: {$lt: new Date(Date.now() - 10*60*1000)}
    }},
    {$count: "idle_connections"}
  ]).forEach(printjson)'

echo -e "\n4. Connection Churn:"
kubectl exec -n mongodb mongodb-0 -- mongosh --quiet --eval '
  var conn = db.serverStatus().connections;
  print("Total created: " + conn.totalCreated);
  print("Current: " + conn.current);
  print("Churn ratio: " + (conn.totalCreated / conn.current).toFixed(2));
'
EOF

chmod +x /tmp/connection-audit.sh
# Run weekly via cron or manually
```

## Escalation

**Escalate if:**
- Connections remain high after application restarts
- Connection leak identified but cannot locate source
- Need to implement replica set for connection distribution
- Requires application code changes

**Escalation Path:**
1. **15 minutes**: Restart applications with high connection count
2. **30 minutes**: Engage application development team
3. **1 hour**: Engage database architecture team
4. **2 hours**: Consider emergency capacity increase

**Communication Template:**
```
ISSUE: MongoDB High Connection Count
SEVERITY: Warning
CURRENT CONNECTIONS: [X] / [Y] ([Z]%)
TOP CONSUMERS:
- Application: [name], Connections: [count]
- Application: [name], Connections: [count]
SYMPTOMS: [Connection exhaustion/Timeouts/etc]
ACTIONS TAKEN:
- Restarted applications: [list]
- Increased connection limit: [old] -> [new]
- Optimized pool settings: [details]
ROOT CAUSE: [If identified]
STATUS: [Mitigated/Ongoing]
NEXT STEPS: [Additional actions needed]
```

## Related Runbooks

- [MongoDB Service Down](./mongodb-down.md)
- [MongoDB Slow Queries](./mongodb-slow-queries.md)
- [Agent Bruno MongoDB Connection Issues](../agent-bruno/mongodb-connection-issues.md)

## Additional Resources

- [MongoDB Connection String Options](https://docs.mongodb.com/manual/reference/connection-string/)
- [Connection Pool Monitoring](https://docs.mongodb.com/manual/reference/command/connPoolStats/)
- [Production Notes](https://docs.mongodb.com/manual/administration/production-notes/#connection-pools)

---

**Last Updated**: 2025-10-15  
**Version**: 1.0  
**Maintainer**: Homelab Platform Team

