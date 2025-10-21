# 🚨 Runbook: Agent Bruno MongoDB Connection Issues

## Alert Information

**Alert Name:** `AgentBrunoMongoDBConnectionFailure`  
**Severity:** High  
**Component:** agent-bruno  
**Service:** mongodb-connection

## Symptom

Agent Bruno cannot connect to MongoDB, causing persistent memory storage failures. Long-term conversation history cannot be saved or retrieved.

## Impact

- **User Impact:** MEDIUM - No persistent conversation history
- **Business Impact:** LOW - AI still works with session memory from Redis
- **Data Impact:** HIGH - Cannot save conversations to persistent storage

## Diagnosis

### 1. Check Agent Bruno Logs

```bash
kubectl logs -n bruno -l app=agent-bruno --tail=100 | grep -i mongo
```

### 2. Check MongoDB Pod Status

```bash
kubectl get pods -n mongodb
kubectl describe pod -n mongodb -l app.kubernetes.io/name=mongodb
```

### 3. Check MongoDB Service

```bash
kubectl get svc -n mongodb
kubectl get endpoints -n mongodb mongodb
```

### 4. Test MongoDB Connectivity

```bash
# From agent-bruno pod
kubectl exec -it -n bruno deployment/agent-bruno -- sh -c 'nc -zv mongodb.mongodb.svc.cluster.local 27017'

# Or test with Python
kubectl exec -it -n bruno deployment/agent-bruno -- python3 -c "
from pymongo import MongoClient
try:
    client = MongoClient('mongodb://mongodb.mongodb.svc.cluster.local:27017', serverSelectionTimeoutMS=5000)
    print('Server Info:', client.server_info()['version'])
    print('Databases:', client.list_database_names())
except Exception as e:
    print('ERROR:', str(e))
"
```

### 5. Check MongoDB Health

```bash
# Connect to MongoDB directly
kubectl exec -it -n mongodb statefulset/mongodb -- mongosh --eval "db.adminCommand('ping')"
kubectl exec -it -n mongodb statefulset/mongodb -- mongosh --eval "db.serverStatus().connections"
```

## Resolution Steps

### Step 1: Verify MongoDB is running

```bash
kubectl get pods -n mongodb -l app.kubernetes.io/name=mongodb
```

### Step 2: Check MongoDB Logs

```bash
kubectl logs -n mongodb -l app.kubernetes.io/name=mongodb --tail=100
```

### Step 3: Common Issues and Fixes

#### Issue: MongoDB Pod Not Running
**Cause:** MongoDB crashed or failed to start  
**Fix:**
```bash
# Check why MongoDB failed
kubectl describe pod -n mongodb -l app.kubernetes.io/name=mongodb

# Check for disk issues
kubectl exec -n mongodb statefulset/mongodb -- df -h

# Restart MongoDB StatefulSet
kubectl rollout restart statefulset -n mongodb mongodb

# Force reconcile Flux HelmRelease
flux reconcile helmrelease mongodb -n mongodb
```

#### Issue: Network Policy Blocking
**Cause:** Network policies preventing connection  
**Fix:**
```bash
# Check network policies
kubectl get networkpolicies -n mongodb
kubectl get networkpolicies -n bruno

# Test connectivity
kubectl exec -it -n bruno deployment/agent-bruno -- ping mongodb.mongodb.svc.cluster.local
```

#### Issue: Wrong MongoDB URL
**Cause:** Incorrect MONGODB_URL environment variable  
**Fix:**
```bash
# Check current configuration
kubectl get deployment -n bruno agent-bruno -o jsonpath='{.spec.template.spec.containers[0].env[?(@.name=="MONGODB_URL")].value}'

# Should be: mongodb://mongodb.mongodb.svc.cluster.local:27017

# If incorrect, update deployment
kubectl edit deployment -n bruno agent-bruno
# Update MONGODB_URL and MONGODB_DB environment variables
```

#### Issue: Database Disk Full
**Cause:** PVC storage exhausted  
**Fix:**
```bash
# Check PVC usage
kubectl exec -n mongodb statefulset/mongodb -- df -h /bitnami/mongodb

# Check PVC
kubectl get pvc -n mongodb
kubectl describe pvc -n mongodb

# Expand PVC if supported
kubectl edit pvc -n mongodb datadir-mongodb-0
# Increase spec.resources.requests.storage

# Or clean old data
kubectl exec -it -n mongodb statefulset/mongodb -- mongosh
# Inside mongosh:
# use agent_bruno
# db.conversations.countDocuments()
# db.conversations.deleteMany({timestamp: {$lt: new Date(Date.now() - 90*24*60*60*1000)}})  # Delete older than 90 days
```

#### Issue: Too Many Connections
**Cause:** MongoDB connection limit reached  
**Fix:**
```bash
# Check current connections
kubectl exec -it -n mongodb statefulset/mongodb -- mongosh --eval "db.serverStatus().connections"

# Check max connections
kubectl exec -it -n mongodb statefulset/mongodb -- mongosh --eval "db.serverStatus().connections.totalCreated"

# Restart Agent Bruno to reset connections
kubectl rollout restart deployment/agent-bruno -n bruno

# Increase max connections in MongoDB
kubectl edit helmrelease -n mongodb mongodb
# Add configuration for maxIncomingConnections
```

#### Issue: DNS Resolution Failure
**Cause:** DNS not resolving mongodb service  
**Fix:**
```bash
# Test DNS from agent-bruno pod
kubectl exec -it -n bruno deployment/agent-bruno -- nslookup mongodb.mongodb.svc.cluster.local

# Check CoreDNS
kubectl get pods -n kube-system -l k8s-app=kube-dns
kubectl logs -n kube-system -l k8s-app=kube-dns --tail=50
```

#### Issue: MongoDB Authentication Failed
**Cause:** Credentials mismatch (if auth enabled)  
**Fix:**
```bash
# Check MongoDB auth configuration
kubectl exec -it -n mongodb statefulset/mongodb -- mongosh --eval "db.adminCommand({getParameter: 1, authenticationMechanisms: 1})"

# Get MongoDB credentials from secret
kubectl get secret -n mongodb mongodb -o jsonpath='{.data.mongodb-root-password}' | base64 -d

# Update agent-bruno deployment with credentials
kubectl edit deployment -n bruno agent-bruno
# Update MONGODB_URL with credentials if needed
```

#### Issue: Collection/Index Issues
**Cause:** Database corruption or missing indexes  
**Fix:**
```bash
# Connect to MongoDB
kubectl exec -it -n mongodb statefulset/mongodb -- mongosh

# Inside mongosh:
use agent_bruno
db.conversations.getIndexes()

# Create indexes if missing
db.conversations.createIndex({ip: 1})
db.conversations.createIndex({timestamp: -1})
db.conversations.createIndex({ip: 1, timestamp: -1})

# Check collection stats
db.conversations.stats()

# Repair database if needed
db.repairDatabase()
```

### Step 4: Restart Agent Bruno

```bash
kubectl rollout restart deployment/agent-bruno -n bruno
kubectl rollout status deployment/agent-bruno -n bruno
```

## Verification

1. Check MongoDB connectivity from agent-bruno:
```bash
kubectl exec -it -n bruno deployment/agent-bruno -- python3 -c "
from pymongo import MongoClient
import datetime

client = MongoClient('mongodb://mongodb.mongodb.svc.cluster.local:27017')
db = client['agent_bruno']

# Test write
result = db.conversations.insert_one({
    'ip': '192.168.1.100',
    'timestamp': datetime.datetime.utcnow(),
    'message': 'test',
    'response': 'test',
    'test': True
})
print('Insert ID:', result.inserted_id)

# Test read
doc = db.conversations.find_one({'_id': result.inserted_id})
print('Found document:', doc)

# Clean up test data
db.conversations.delete_one({'_id': result.inserted_id})
print('Test complete')
"
```

2. Test persistent memory via API:
```bash
kubectl port-forward -n bruno svc/agent-bruno-service 8080:8080

# Send chat messages to create history
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -H "X-Forwarded-For: 192.168.1.100" \
  -d '{"message": "First message"}'

curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -H "X-Forwarded-For: 192.168.1.100" \
  -d '{"message": "Second message"}'

# Get full history (from MongoDB)
curl http://localhost:8080/memory/192.168.1.100/history
```

3. Verify MongoDB has conversation data:
```bash
kubectl exec -it -n mongodb statefulset/mongodb -- mongosh
# Inside mongosh:
# use agent_bruno
# db.conversations.countDocuments()
# db.conversations.find({ip: "192.168.1.100"}).sort({timestamp: -1}).limit(5)
```

4. Check Agent Bruno metrics:
```bash
kubectl port-forward -n bruno svc/agent-bruno-service 8080:8080
curl http://localhost:8080/metrics | grep bruno_memory_operations
```

## Prevention

1. Monitor MongoDB health and availability
2. Set up MongoDB regular backups
3. Configure proper resource limits for MongoDB
4. Implement connection pooling (done by default)
5. Set up MongoDB replication for HA
6. Monitor MongoDB disk usage
7. Create and maintain proper indexes
8. Test MongoDB connectivity in deployment pipeline
9. Set up automated data retention policies

## Performance Tips

1. **Indexes**: Ensure proper indexes exist:
   - `{ip: 1}`
   - `{timestamp: -1}`
   - `{ip: 1, timestamp: -1}`

2. **Data Retention**: Implement automatic cleanup:
```javascript
// Delete conversations older than 180 days
db.conversations.deleteMany({
  timestamp: {$lt: new Date(Date.now() - 180*24*60*60*1000)}
})
```

3. **Connection Pooling**: PyMongo uses connection pooling by default
   - Max pool size: 100 (configurable)
   - Min pool size: 0

4. **Write Concern**: Default write concern is `w: 1` (acknowledgment from primary)

## Related Alerts

- `AgentBrunoAPIDown`
- `MongoDBDown`
- `MongoDBHighDiskUsage`
- `MongoDBSlowQueries`
- `AgentBrunoMemoryOperationsFailed`

## Escalation

If unable to resolve within 30 minutes:
1. Check MongoDB StatefulSet configuration
2. Verify MongoDB HelmRelease values
3. Check cluster networking and DNS
4. Review MongoDB logs for corruption
5. Check PVC status and storage class
6. Contact infrastructure team

## Additional Resources

- [MongoDB Documentation](https://www.mongodb.com/docs/)
- [PyMongo Documentation](https://pymongo.readthedocs.io/)
- [Agent Bruno Memory System](../../../flux/clusters/homelab/infrastructure/agent-bruno/README.md#-memory-system)
- [MongoDB Troubleshooting](https://www.mongodb.com/docs/manual/reference/program/mongod/#std-program-mongod)

---

**Last Updated:** 2025-10-15  
**Version:** 1.0

