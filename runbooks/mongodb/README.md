# 📚 MongoDB Runbooks

Comprehensive operational runbooks for troubleshooting and resolving MongoDB issues in the homelab Kubernetes cluster.

## Overview

MongoDB is our NoSQL document database running in standalone mode:
- **Architecture**: Standalone (single instance)
- **Namespace**: mongodb
- **Storage**: 20Gi persistent volume
- **Port**: 27017
- **Authentication**: Disabled (internal cluster use only)
- **Metrics**: Enabled with Prometheus ServiceMonitor
- **Resource Allocation**:
  - Requests: 200m CPU, 512Mi memory
  - Limits: 1000m CPU, 2Gi memory

## Quick Reference

| Alert | Severity | Impact | Runbook |
|-------|----------|--------|---------|
| MongoDBDown | Critical | Complete database outage | [mongodb-down.md](./mongodb-down.md) |
| MongoDBHighMemory | Warning | Memory pressure/OOMKills | [mongodb-high-memory.md](./mongodb-high-memory.md) |
| MongoDBSlowQueries | Warning | Slow database operations | [mongodb-slow-queries.md](./mongodb-slow-queries.md) |
| MongoDBHighConnections | Warning | Connection pool exhaustion | [mongodb-high-connections.md](./mongodb-high-connections.md) |
| MongoDBReplicationLag | Warning | Replication issues (if enabled) | [mongodb-replication-lag.md](./mongodb-replication-lag.md) |
| MongoDBStorageFull | Critical | Disk space exhausted | [mongodb-storage-full.md](./mongodb-storage-full.md) |

## Runbooks

### 🚨 Critical Issues

#### [MongoDB Service Down](./mongodb-down.md)
Complete MongoDB outage - database unavailable.

**Quick Check:**
```bash
kubectl get pods -n mongodb
```

**Quick Fix:**
```bash
# Restart MongoDB
kubectl rollout restart statefulset -n mongodb mongodb
```

---

#### [Storage Full](./mongodb-storage-full.md)
MongoDB storage volume full - cannot write new data.

**Quick Check:**
```bash
kubectl exec -n mongodb mongodb-0 -- df -h /bitnami/mongodb
```

**Quick Fix:**
```bash
# Increase PVC size or clean old data
kubectl edit pvc -n mongodb data-mongodb-0
```

---

### ⚠️ Warning Issues

#### [High Memory Usage](./mongodb-high-memory.md)
MongoDB experiencing memory pressure or OOMKills.

**Quick Check:**
```bash
kubectl top pods -n mongodb
kubectl get pods -n mongodb -o json | jq -r '.items[] | select(.status.containerStatuses[].lastState.terminated.reason == "OOMKilled") | .metadata.name'
```

**Quick Fix:**
```bash
# Increase memory limits
kubectl edit helmrelease -n mongodb mongodb
# Update memory limits
```

---

#### [Slow Queries](./mongodb-slow-queries.md)
Database queries taking longer than expected.

**Quick Check:**
```bash
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.currentOp({"secs_running": {$gte: 3}})'
```

**Quick Fix:**
```bash
# Check for missing indexes
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.getProfilingStatus()'
```

---

#### [High Connections](./mongodb-high-connections.md)
MongoDB connection count approaching limits.

**Quick Check:**
```bash
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.serverStatus().connections'
```

**Quick Fix:**
```bash
# Increase max connections or fix connection leaks
kubectl exec -n mongodb mongodb-0 -- mongosh admin --eval 'db.runCommand({setParameter: 1, maxIncomingConnections: 2000})'
```

---

## Common Troubleshooting Commands

### Check Overall Health
```bash
# All pods status
kubectl get pods -n mongodb

# Resource usage
kubectl top pods -n mongodb

# Recent events
kubectl get events -n mongodb --sort-by='.lastTimestamp' | head -20
```

### Check MongoDB Status
```bash
# Connect to MongoDB shell
kubectl exec -it -n mongodb mongodb-0 -- mongosh

# Check server status
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.serverStatus()'

# Check database stats
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.stats()'

# Check connection info
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.serverStatus().connections'
```

### Check Storage
```bash
# Check disk usage
kubectl exec -n mongodb mongodb-0 -- df -h /bitnami/mongodb

# Check PVC status
kubectl get pvc -n mongodb

# Check database sizes
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.adminCommand({listDatabases: 1})'
```

### Check Performance
```bash
# Check current operations
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.currentOp()'

# Check slow queries (> 3 seconds)
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.currentOp({"secs_running": {$gte: 3}})'

# Check replication lag (if replica set)
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'rs.status()'

# Check locks
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.serverStatus().locks'
```

### Check Logs
```bash
# View MongoDB logs
kubectl logs -n mongodb mongodb-0 --tail=100

# Follow logs
kubectl logs -n mongodb mongodb-0 -f

# Check for errors
kubectl logs -n mongodb mongodb-0 --tail=500 | grep -i "error\|warn\|exception"
```

### Test Connectivity
```bash
# Test from agent-bruno
kubectl exec -n bruno deployment/agent-bruno -- nc -zv mongodb.mongodb.svc.cluster.local 27017

# Test from another pod
kubectl run mongodb-test --image=mongo:7 --rm -it --restart=Never -- mongosh mongodb://mongodb.mongodb.svc.cluster.local:27017/test --eval 'db.runCommand({ping: 1})'
```

## Architecture

```
┌─────────────────┐
│  Applications   │
│  - agent-bruno  │
│  - agent-sre    │
└────────┬────────┘
         │
         │ mongodb://mongodb.mongodb.svc.cluster.local:27017
         ▼
┌─────────────────┐
│  MongoDB Service│
│   (ClusterIP)   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  MongoDB Pod    │
│  mongodb-0      │
│  (StatefulSet)  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Persistent Vol │
│     (20Gi)      │
│  /bitnami/mongo │
└─────────────────┘
         │
         ▼
┌─────────────────┐
│  Prometheus     │◄── Metrics Exporter
└─────────────────┘
```

## Configuration

**Location**: `flux/clusters/homelab/infrastructure/mongodb/helmrelease.yaml`

**Key Settings**:
- Chart: mongodb (Bitnami)
- Version: >= 16.5.45
- Architecture: Standalone
- Auth: Disabled (internal service)
- Persistence: 20Gi
- Metrics: Enabled with ServiceMonitor

## Database Management

### Create Database
```bash
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'use myapp; db.createCollection("users")'
```

### List Databases
```bash
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'show dbs'
```

### Create Index
```bash
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval 'db.users.createIndex({email: 1}, {unique: true})'
```

### Check Indexes
```bash
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval 'db.users.getIndexes()'
```

### Backup Database
```bash
# Dump database
kubectl exec -n mongodb mongodb-0 -- mongodump --db=myapp --out=/tmp/backup

# Copy to local
kubectl cp mongodb/mongodb-0:/tmp/backup ./mongodb-backup-$(date +%Y%m%d)
```

### Restore Database
```bash
# Copy backup to pod
kubectl cp ./mongodb-backup mongodb/mongodb-0:/tmp/restore

# Restore
kubectl exec -n mongodb mongodb-0 -- mongorestore --db=myapp /tmp/restore/myapp
```

## Performance Tuning

### Connection Pool Settings
For application clients:
```javascript
// Recommended connection string parameters
mongodb://mongodb.mongodb.svc.cluster.local:27017/myapp?maxPoolSize=50&minPoolSize=10&maxIdleTimeMS=30000
```

### Query Optimization
```bash
# Enable profiling for slow queries (> 100ms)
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.setProfilingLevel(1, {slowms: 100})'

# Check profiling data
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.system.profile.find().limit(5).sort({ts: -1}).pretty()'

# Analyze query performance
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval 'db.users.find({email: "test@example.com"}).explain("executionStats")'
```

### Memory Management
```bash
# Check WiredTiger cache size
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.serverStatus().wiredTiger.cache'

# Check memory usage
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.serverStatus().mem'
```

## Monitoring

### Key Metrics to Monitor
- Connection count and utilization
- Query execution time
- Memory usage (WiredTiger cache)
- Disk I/O and storage usage
- Replication lag (if applicable)
- Lock contention

### Prometheus Queries
```promql
# MongoDB up status
up{job="mongodb"}

# Connection count
mongodb_connections{state="current"}

# Operation latency
mongodb_op_latencies_latency_total

# Memory usage
mongodb_memory{type="resident"}

# Query rate
rate(mongodb_op_counters_total[5m])
```

## Escalation Matrix

| Issue | First Response | Escalation Time | Escalate To |
|-------|---------------|-----------------|-------------|
| Complete outage | Restart pod | 15 minutes | Database team |
| Storage full | Expand PVC/cleanup | 30 minutes | Storage admin |
| High memory | Increase limits | 30 minutes | Capacity planning |
| Slow queries | Check indexes | 1 hour | Database team |
| Connection issues | Check app pools | 30 minutes | Application team |

## Related Documentation

- [MongoDB Configuration](../../../flux/clusters/homelab/infrastructure/mongodb/helmrelease.yaml)
- [Agent Bruno MongoDB Connection](../agent-bruno/mongodb-connection-issues.md)
- [Architecture Overview](../../../ARCHITECTURE.md)
- [MongoDB Official Docs](https://docs.mongodb.com/)

## Best Practices

1. **Indexes**: Create indexes for frequently queried fields
2. **Connection Pooling**: Use connection pools in applications
3. **Query Optimization**: Use explain() to analyze queries
4. **Monitoring**: Set up alerts for connection count and slow queries
5. **Backups**: Schedule regular backups
6. **Resource Limits**: Monitor and adjust CPU/memory as needed

## Support

For issues not covered by these runbooks:
1. Check MongoDB logs: `kubectl logs -n mongodb mongodb-0`
2. Review HelmRelease: `flux get helmreleases -n mongodb`
3. Consult [MongoDB documentation](https://docs.mongodb.com/)
4. Check [MongoDB Community](https://www.mongodb.com/community/forums/)

---

**Last Updated**: 2025-10-15  
**MongoDB Version**: 7.x (managed by Helm chart >= 16.5.45)  
**Maintainer**: Homelab Platform Team

