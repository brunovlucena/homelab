# 🚨 Runbook: PostgreSQL Service Down

## Alert Information

**Alert Name:** `PostgreSQLDown`  
**Severity:** Critical  
**Component:** postgres  
**Service:** postgres-postgresql  
**Threshold:** PostgreSQL instance unreachable for > 1 minute

## Symptom

PostgreSQL database is completely unavailable - applications cannot connect or execute queries.

## Impact

- **User Impact:** CRITICAL - Complete application outage
- **Business Impact:** CRITICAL - All database-dependent services down
- **Data Impact:** CRITICAL - No read/write operations possible

## Diagnosis

### 1. Check Pod Status

```bash
# Check if pod is running
kubectl get pods -n postgres

# Expected: postgres-postgresql-0 in Running state with 1/1 READY
# Problem states: CrashLoopBackOff, Pending, Error, OOMKilled
```

### 2. Check Recent Events

```bash
# Check for issues
kubectl get events -n postgres --sort-by='.lastTimestamp' | head -20

# Look for:
# - ImagePullBackOff
# - Failed scheduling
# - Liveness/Readiness probe failures
# - OOMKilled events
# - Volume mount failures
```

### 3. Check Pod Logs

```bash
# View recent logs
kubectl logs -n postgres postgres-postgresql-0 --tail=100

# Check for errors
kubectl logs -n postgres postgres-postgresql-0 --tail=500 | grep -i "error\|fatal\|panic"

# Check previous container logs (if pod restarted)
kubectl logs -n postgres postgres-postgresql-0 --previous
```

### 4. Check Service Endpoint

```bash
# Verify service exists
kubectl get svc -n postgres postgres-postgresql

# Check endpoints
kubectl get endpoints -n postgres postgres-postgresql

# Should show pod IP - empty means no healthy pods
```

### 5. Check PostgreSQL Process

```bash
# If pod is running, check PostgreSQL status
kubectl exec -n postgres postgres-postgresql-0 -- pg_isready -U postgres

# Check if PostgreSQL process is running
kubectl exec -n postgres postgres-postgresql-0 -- ps aux | grep postgres
```

### 6. Check Storage

```bash
# Check PVC status
kubectl get pvc -n postgres

# Check if volume is bound
kubectl describe pvc -n postgres data-postgres-postgresql-0

# Check disk space
kubectl exec -n postgres postgres-postgresql-0 -- df -h /var/lib/postgresql/data
```

## Resolution Steps

### Step 1: Quick Checks

#### Issue: Pod Not Running
**Cause:** Pod crashed or failed to start  
**Fix:**
```bash
# Check pod description for errors
kubectl describe pod -n postgres postgres-postgresql-0

# Check resource availability
kubectl describe node | grep -A 5 "Allocated resources"

# If pod is in CrashLoopBackOff, check logs
kubectl logs -n postgres postgres-postgresql-0 --previous
```

#### Issue: Storage Mount Failure
**Cause:** PVC not bound or volume unavailable  
**Fix:**
```bash
# Check PVC status
kubectl get pvc -n postgres
kubectl describe pvc -n postgres data-postgres-postgresql-0

# If PVC pending, check storage class
kubectl get sc

# If needed, delete PVC and recreate (⚠️ DATA LOSS)
# kubectl delete pvc -n postgres data-postgres-postgresql-0
# kubectl delete pod -n postgres postgres-postgresql-0
```

#### Issue: OOMKilled
**Cause:** PostgreSQL exceeded memory limits  
**Fix:**
```bash
# Check if pod was OOMKilled
kubectl get pod -n postgres postgres-postgresql-0 -o jsonpath='{.status.containerStatuses[0].lastState}'

# Increase memory limits
kubectl edit helmrelease -n postgres postgres
# Update:
#   resources:
#     limits:
#       memory: "4Gi"  # Increased from 2Gi
#     requests:
#       memory: "1Gi"  # Increased from 512Mi

# Apply changes
flux reconcile helmrelease postgres -n postgres
```

### Step 2: Restart PostgreSQL

#### Simple Restart

```bash
# Delete pod (StatefulSet will recreate it)
kubectl delete pod -n postgres postgres-postgresql-0

# Monitor restart
kubectl get pod -n postgres postgres-postgresql-0 -w

# Watch events
kubectl get events -n postgres --watch
```

#### Helm Release Reconciliation

```bash
# Force reconcile HelmRelease
flux reconcile helmrelease postgres -n postgres

# Check reconciliation status
flux get helmreleases -n postgres

# If reconciliation failed, check logs
kubectl logs -n postgres deployment/helm-controller -f
```

### Step 3: Database Recovery

#### Issue: Corrupted Data Directory
**Cause:** Disk corruption or improper shutdown  
**Fix:**
```bash
# Check PostgreSQL logs for corruption messages
kubectl logs -n postgres postgres-postgresql-0 --tail=200 | grep -i "corrupt\|invalid"

# If corruption detected, attempt recovery
kubectl exec -n postgres postgres-postgresql-0 -- pg_resetwal /var/lib/postgresql/data/pgdata

# If severe corruption, restore from backup
# 1. Stop PostgreSQL
kubectl scale statefulset -n postgres postgres-postgresql --replicas=0

# 2. Delete PVC (⚠️ DATA LOSS)
kubectl delete pvc -n postgres data-postgres-postgresql-0

# 3. Recreate and restore from backup
kubectl scale statefulset -n postgres postgres-postgresql --replicas=1
kubectl wait --for=condition=ready pod -n postgres postgres-postgresql-0 --timeout=300s

# 4. Copy backup and restore
kubectl cp ./postgres-backup.sql postgres/postgres-postgresql-0:/tmp/restore.sql
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres < /tmp/restore.sql
```

#### Issue: Lock Files Preventing Startup
**Cause:** Stale lock files from improper shutdown  
**Fix:**
```bash
# Remove lock files
kubectl exec -n postgres postgres-postgresql-0 -- rm -f /var/lib/postgresql/data/postmaster.pid
kubectl exec -n postgres postgres-postgresql-0 -- rm -f /tmp/.s.PGSQL.5432*

# Restart pod
kubectl delete pod -n postgres postgres-postgresql-0
```

### Step 4: Configuration Issues

#### Issue: Invalid Configuration
**Cause:** Invalid postgresql.conf settings  
**Fix:**
```bash
# Check HelmRelease values
kubectl get helmrelease -n postgres postgres -o yaml

# Check ConfigMap if exists
kubectl get cm -n postgres

# Fix configuration in HelmRelease
kubectl edit helmrelease -n postgres postgres

# Reconcile
flux reconcile helmrelease postgres -n postgres
```

#### Issue: Authentication Failure
**Cause:** pg_hba.conf misconfiguration  
**Fix:**
```bash
# Check authentication config
kubectl exec -n postgres postgres-postgresql-0 -- cat /var/lib/postgresql/data/pg_hba.conf

# For trust method (internal only), should have:
# host all all all trust

# Update in HelmRelease if needed
kubectl edit helmrelease -n postgres postgres
```

### Step 5: Network Issues

#### Issue: Service Not Routing
**Cause:** Service misconfigured or endpoints missing  
**Fix:**
```bash
# Check service
kubectl get svc -n postgres postgres-postgresql -o yaml

# Check endpoints
kubectl get endpoints -n postgres postgres-postgresql

# If endpoints missing, pod may not be ready
kubectl get pod -n postgres postgres-postgresql-0 -o yaml | grep -A 10 readinessProbe

# Test connectivity from another pod
kubectl run postgres-test --image=postgres:16 --rm -it --restart=Never -- \
  psql postgresql://postgres@postgres-postgresql.postgres.svc.cluster.local:5432/postgres -c "SELECT 1;"
```

### Step 6: Resource Exhaustion

#### Issue: No CPU/Memory Available
**Cause:** Node resources exhausted  
**Fix:**
```bash
# Check node resources
kubectl describe node | grep -A 5 "Allocated resources"

# Check pod requests
kubectl get pod -n postgres postgres-postgresql-0 -o jsonpath='{.spec.containers[0].resources}'

# If node full, consider:
# 1. Reducing other workloads
# 2. Adding nodes to cluster
# 3. Reducing PostgreSQL resource requests
```

## Verification

### 1. Check Pod Health

```bash
# Pod should be Running with 1/1 READY
kubectl get pods -n postgres

# Check pod age (should be recent if restarted)
kubectl get pod -n postgres postgres-postgresql-0 -o jsonpath='{.status.startTime}'
```

### 2. Test PostgreSQL Connectivity

```bash
# Test pg_isready
kubectl exec -n postgres postgres-postgresql-0 -- pg_isready -U postgres

# Test simple query
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "SELECT 1;"

# Check database list
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "\l"
```

### 3. Check Database Operations

```bash
# Check active connections
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT count(*) FROM pg_stat_activity;"

# Check if databases are accessible
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c "SELECT version();"

# Verify write operations
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "CREATE TABLE test_write (id serial, created_at timestamp DEFAULT now());"
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -d bruno_site -c \
  "DROP TABLE test_write;"
```

### 4. Test from Application

```bash
# Test from homepage API (if applicable)
kubectl exec -n homepage deployment/homepage-api -- nc -zv postgres-postgresql.postgres.svc.cluster.local 5432

# Check application logs for successful connections
kubectl logs -n homepage deployment/homepage-api --tail=50 | grep -i postgres
```

### 5. Check Metrics

```bash
# Check if metrics are being collected
kubectl port-forward -n postgres postgres-postgresql-0 9187:9187 &
curl http://localhost:9187/metrics | grep pg_up

# Should return: pg_up 1
```

### 6. Monitor Stability

```bash
# Watch pod for stability (should stay Running)
kubectl get pod -n postgres postgres-postgresql-0 -w

# Check for new errors in logs
kubectl logs -n postgres postgres-postgresql-0 -f

# No new error/fatal/panic messages should appear
```

## Prevention

### 1. Enable Persistence

```yaml
# In HelmRelease values
primary:
  persistence:
    enabled: true
    size: 20Gi
    storageClass: local-path  # or your storage class
```

### 2. Configure Proper Resources

```yaml
# In HelmRelease values
primary:
  resources:
    limits:
      memory: "2Gi"
      cpu: "1000m"
    requests:
      memory: "512Mi"
      cpu: "200m"
```

### 3. Set Up Health Checks

```yaml
# In HelmRelease values
primary:
  livenessProbe:
    enabled: true
    initialDelaySeconds: 30
    periodSeconds: 10
    timeoutSeconds: 5
    failureThreshold: 6
  
  readinessProbe:
    enabled: true
    initialDelaySeconds: 5
    periodSeconds: 10
    timeoutSeconds: 5
    failureThreshold: 6
```

### 4. Configure Monitoring

```yaml
# Prometheus alert rules
- alert: PostgreSQLDown
  expr: pg_up == 0
  for: 1m
  labels:
    severity: critical
  annotations:
    summary: "PostgreSQL is down"
    description: "PostgreSQL instance {{ $labels.instance }} has been down for more than 1 minute"

- alert: PostgreSQLRestartingOften
  expr: changes(pg_postmaster_start_time_seconds[1h]) > 2
  labels:
    severity: warning
  annotations:
    summary: "PostgreSQL restarting frequently"
```

### 5. Regular Backups

```bash
# Set up automated backup CronJob
apiVersion: batch/v1
kind: CronJob
metadata:
  name: postgres-backup
  namespace: postgres
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: postgres:16
            command:
            - /bin/sh
            - -c
            - |
              pg_dump -h postgres-postgresql -U postgres -d bruno_site > /backup/bruno_site-$(date +%Y%m%d).sql
              # Upload to S3 or persistent storage
          restartPolicy: OnFailure
```

### 6. Implement High Availability

For production, consider:
- **Replication:** Set up streaming replication with standby
- **Connection Pooling:** Use PgBouncer or PgPool
- **Load Balancing:** Distribute read queries to replicas
- **Automated Failover:** Use Patroni or Stolon for HA

### 7. Capacity Planning

```bash
# Monitor growth trends
# - Database size
# - Connection count
# - Query volume
# - Storage usage

# Set alerts for capacity thresholds:
# - Storage > 70% full
# - Connections > 80% of max
# - Memory usage > 80%
```

## Common Failure Scenarios

### Scenario 1: Pod Stuck in Pending

**Symptoms:** Pod shows Pending state  
**Causes:**
- No nodes with sufficient resources
- PVC cannot be bound
- Node selector/affinity not matching

**Resolution:**
```bash
kubectl describe pod -n postgres postgres-postgresql-0
# Check "Events" section for exact reason
```

### Scenario 2: CrashLoopBackOff

**Symptoms:** Pod continuously restarting  
**Causes:**
- Application error on startup
- Invalid configuration
- Resource limits too low
- Corrupted data

**Resolution:**
```bash
kubectl logs -n postgres postgres-postgresql-0 --previous
# Fix underlying issue then restart
```

### Scenario 3: Liveness Probe Failure

**Symptoms:** Pod killed and restarted by kubelet  
**Causes:**
- PostgreSQL slow to respond
- Resource contention
- Long-running queries blocking

**Resolution:**
```bash
# Increase probe timeout/threshold
kubectl edit helmrelease -n postgres postgres
```

## Escalation

If PostgreSQL remains down after applying fixes:

1. ✅ Check cluster-wide issues (etcd, API server, nodes)
2. 📊 Review storage system health
3. 🔍 Analyze PostgreSQL crash dumps if available
4. 💾 Consider restoring from backup to new PVC
5. 📞 Contact database team for advanced troubleshooting
6. 🆘 Escalate to platform team if cluster issue

## Additional Resources

- [PostgreSQL Logs Analysis](https://www.postgresql.org/docs/current/logfile-analysis.html)
- [PostgreSQL Crash Recovery](https://www.postgresql.org/docs/current/crash-recovery.html)
- [Kubernetes StatefulSet Debugging](https://kubernetes.io/docs/tasks/debug/debug-application/)
- [Bitnami PostgreSQL Chart](https://github.com/bitnami/charts/tree/main/bitnami/postgresql)

---

**Last Updated:** 2025-10-15  
**Version:** 1.0  
**Owner:** Infrastructure Team

