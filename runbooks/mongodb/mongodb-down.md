# 🚨 Runbook: MongoDB Service Down

## Alert Information

**Alert Name:** `MongoDBDown`  
**Severity:** Critical  
**Component:** mongodb  
**Service:** mongodb  
**Threshold:** MongoDB instance unavailable for > 1 minute

## Symptom

MongoDB service is completely unavailable - applications cannot connect or perform database operations.

## Impact

- **User Impact:** CRITICAL - Complete application failure for services using MongoDB
- **Business Impact:** CRITICAL - No data persistence, application errors
- **Data Impact:** CRITICAL - No read/write operations possible

## Diagnosis

### 1. Check Pod Status

```bash
# Check if MongoDB pod is running
kubectl get pods -n mongodb

# Check pod details
kubectl describe pod -n mongodb mongodb-0

# Check recent events
kubectl get events -n mongodb --sort-by='.lastTimestamp' | head -20
```

**Expected Output:**
- Pod should be in `Running` state with `1/1` ready
- No recent `CrashLoopBackOff` or `Error` events

### 2. Check Service Connectivity

```bash
# Check service endpoint
kubectl get svc -n mongodb

# Check endpoints
kubectl get endpoints -n mongodb mongodb

# Test connection from within cluster
kubectl run mongodb-test --image=mongo:7 --rm -it --restart=Never -- \
  mongosh mongodb://mongodb.mongodb.svc.cluster.local:27017/test --eval 'db.runCommand({ping: 1})'
```

### 3. Check MongoDB Logs

```bash
# View recent logs
kubectl logs -n mongodb mongodb-0 --tail=100

# Check for errors
kubectl logs -n mongodb mongodb-0 --tail=500 | grep -i "error\|fatal\|exception"

# Check previous container logs (if restarted)
kubectl logs -n mongodb mongodb-0 --previous
```

**Common Error Patterns:**
- `WiredTiger error`: Storage engine issues
- `out of memory`: OOMKilled
- `No space left on device`: Disk full
- `connection refused`: Port binding issues

### 4. Check Resource Usage

```bash
# Check memory and CPU
kubectl top pod -n mongodb

# Check resource limits
kubectl get pod -n mongodb mongodb-0 -o jsonpath='{.spec.containers[0].resources}'

# Check for OOMKills
kubectl get pods -n mongodb -o json | jq -r '.items[] | select(.status.containerStatuses[].lastState.terminated.reason == "OOMKilled") | .metadata.name'
```

### 5. Check Storage

```bash
# Check PVC status
kubectl get pvc -n mongodb

# Check disk usage
kubectl exec -n mongodb mongodb-0 -- df -h /bitnami/mongodb

# Check PV status
kubectl get pv | grep mongodb
```

## Resolution

### Option 1: Quick Restart (First Response)

```bash
# Delete pod to trigger restart
kubectl delete pod -n mongodb mongodb-0

# Watch pod come back up
kubectl get pods -n mongodb -w

# Wait for pod to be ready
kubectl wait --for=condition=ready pod/mongodb-0 -n mongodb --timeout=5m

# Verify connectivity
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.runCommand({ping: 1})'
```

**Expected Time:** 2-3 minutes

### Option 2: Pod is CrashLooping

```bash
# Check current status
kubectl get pod -n mongodb mongodb-0

# If CrashLoopBackOff, check logs for root cause
kubectl logs -n mongodb mongodb-0 --tail=200

# Common fixes:

# A. If disk full
kubectl exec -n mongodb mongodb-0 -- df -h /bitnami/mongodb
# See mongodb-storage-full.md runbook

# B. If OOMKilled
kubectl edit helmrelease -n mongodb mongodb
# Increase memory limits (see mongodb-high-memory.md)

# C. If permission issues
kubectl exec -n mongodb mongodb-0 -- ls -la /bitnami/mongodb
kubectl exec -n mongodb mongodb-0 -- chown -R mongodb:mongodb /bitnami/mongodb
```

### Option 3: Storage/PVC Issues

```bash
# Check PVC binding
kubectl get pvc -n mongodb
# If PVC in Pending state:

# Check PV availability
kubectl get pv

# Check storage class
kubectl get storageclass

# Delete and recreate PVC (DATA LOSS!)
# Only if PVC corrupted and backup exists
kubectl delete pvc -n mongodb data-mongodb-0
kubectl delete pod -n mongodb mongodb-0
# StatefulSet will recreate both
```

### Option 4: Complete Redeployment

```bash
# Reconcile Flux HelmRelease
flux reconcile helmrelease -n mongodb mongodb

# If Flux reconciliation fails
flux get helmreleases -n mongodb
flux logs --level=error

# Force Helm release upgrade
helm upgrade -n mongodb mongodb bitnami/mongodb --reuse-values

# Last resort: Delete and recreate (DATA LOSS!)
# Only with backup!
kubectl delete helmrelease -n mongodb mongodb
kubectl delete statefulset -n mongodb mongodb
kubectl delete pvc -n mongodb data-mongodb-0
flux reconcile kustomization -n flux-system infrastructure
```

## Post-Resolution Verification

### 1. Health Check

```bash
# Verify pod is running
kubectl get pods -n mongodb

# Check MongoDB status
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.serverStatus().ok'

# Should return: 1
```

### 2. Connectivity Test

```bash
# Test from agent-bruno pod
kubectl exec -n bruno deployment/agent-bruno -- nc -zv mongodb.mongodb.svc.cluster.local 27017

# Should return: Connection to mongodb.mongodb.svc.cluster.local 27017 port [tcp/*] succeeded!
```

### 3. Data Integrity

```bash
# List databases
kubectl exec -n mongodb mongodb-0 -- mongosh --eval 'db.adminCommand({listDatabases: 1})'

# Check collections in main database
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval 'db.getCollectionNames()'

# Verify document count
kubectl exec -n mongodb mongodb-0 -- mongosh myapp --eval 'db.users.countDocuments()'
```

### 4. Metrics Check

```bash
# Verify Prometheus metrics endpoint
kubectl exec -n mongodb mongodb-0 -- curl -s localhost:9216/metrics | head -20

# Check ServiceMonitor
kubectl get servicemonitor -n mongodb

# Query Prometheus for MongoDB metrics
# up{job="mongodb"} should be 1
```

## Root Cause Analysis

### Common Causes

| Cause | Indicator | Prevention |
|-------|-----------|------------|
| OOM Kill | `OOMKilled` in pod status | Increase memory limits, optimize queries |
| Disk Full | `No space left on device` | Monitor storage, implement retention |
| Configuration Error | Pod failing to start | Validate Helm values, test in staging |
| Network Issue | Cannot bind port | Check network policies, service config |
| Data Corruption | WiredTiger errors | Enable journaling, regular backups |
| Resource Starvation | Node pressure | Set appropriate resource requests/limits |

### Investigation Commands

```bash
# Check MongoDB server logs for startup issues
kubectl logs -n mongodb mongodb-0 | grep -A5 -B5 "STORAGE\|CONTROL\|NETWORK"

# Check system logs
kubectl exec -n mongodb mongodb-0 -- dmesg | tail -50

# Check file system
kubectl exec -n mongodb mongodb-0 -- mongosh admin --eval 'db.runCommand({dbStats: 1})'

# Check WiredTiger status
kubectl exec -n mongodb mongodb-0 -- ls -la /bitnami/mongodb/data
```

## Prevention

### 1. Monitoring

Set up alerts for:
```yaml
# Example Prometheus alerts
- alert: MongoDBDown
  expr: up{job="mongodb"} == 0
  for: 1m
  
- alert: MongoDBHighMemory
  expr: mongodb_memory_resident_bytes / mongodb_memory_limit_bytes > 0.8
  for: 5m

- alert: MongoDBStorageUsage
  expr: mongodb_storage_usage_bytes / mongodb_storage_limit_bytes > 0.85
  for: 5m

- alert: MongoDBRestartCount
  expr: rate(kube_pod_container_status_restarts_total{namespace="mongodb"}[15m]) > 0
  for: 5m
```

### 2. Resource Management

```yaml
# Recommended resource configuration
resources:
  requests:
    cpu: 500m
    memory: 1Gi
  limits:
    cpu: 2000m
    memory: 4Gi

# Persistent storage
persistence:
  size: 50Gi  # Adjust based on data growth
```

### 3. Regular Backups

```bash
# Schedule regular backups
kubectl create cronjob mongodb-backup -n mongodb \
  --image=mongo:7 \
  --schedule="0 2 * * *" \
  -- mongodump --host=mongodb.mongodb.svc.cluster.local --out=/backup
```

### 4. Health Checks

```yaml
# Liveness probe
livenessProbe:
  exec:
    command:
      - mongosh
      - --eval
      - db.adminCommand('ping')
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5

# Readiness probe
readinessProbe:
  exec:
    command:
      - mongosh
      - --eval
      - db.adminCommand('ping')
  initialDelaySeconds: 5
  periodSeconds: 10
```

## Escalation

**Escalate if:**
- Service down for > 15 minutes
- Data corruption detected
- Cannot restore from backup
- Multiple restart attempts failed

**Escalation Path:**
1. **Immediate**: Alert on-call SRE
2. **15 minutes**: Engage database team
3. **30 minutes**: Engage platform team
4. **1 hour**: Consider failover strategy

**Communication Template:**
```
INCIDENT: MongoDB Complete Outage
STATUS: [Investigating/Mitigating/Resolved]
START TIME: [timestamp]
IMPACT: All applications using MongoDB unable to persist/read data
ACTIONS TAKEN:
- [List actions]
NEXT STEPS:
- [List next steps]
ETA: [estimated resolution time]
```

## Related Runbooks

- [MongoDB High Memory](./mongodb-high-memory.md)
- [MongoDB Storage Full](./mongodb-storage-full.md)
- [MongoDB Slow Queries](./mongodb-slow-queries.md)
- [Agent Bruno MongoDB Connection Issues](../agent-bruno/mongodb-connection-issues.md)

## Additional Resources

- [MongoDB Troubleshooting Guide](https://docs.mongodb.com/manual/faq/diagnostics/)
- [WiredTiger Storage Engine](https://docs.mongodb.com/manual/core/wiredtiger/)
- [MongoDB Operations Best Practices](https://docs.mongodb.com/manual/administration/production-notes/)

---

**Last Updated**: 2025-10-15  
**Version**: 1.0  
**Maintainer**: Homelab Platform Team

