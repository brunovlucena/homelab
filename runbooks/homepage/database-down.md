# 🚨 Runbook: Bruno Site Database Down

## Alert Information

**Alert Name:** `BrunoSiteDatabaseDown`  
**Severity:** Critical  
**Component:** bruno-site  
**Service:** database

## Symptom

PostgreSQL database has been down for more than 1 minute. All data operations are failing.

## Impact

- **User Impact:** SEVERE - No data can be retrieved or saved
- **Business Impact:** CRITICAL - Complete functionality loss
- **Data Impact:** POTENTIAL - Risk of data loss if not recovered

## Diagnosis

### 1. Check PostgreSQL Pod Status

```bash
kubectl get pods -n postgres
kubectl describe pod -n postgres -l app.kubernetes.io/name=postgresql
```

### 2. Check PostgreSQL Logs

```bash
kubectl logs -n postgres -l app.kubernetes.io/name=postgresql --tail=100
```

### 3. Check PVC Status

```bash
kubectl get pvc -n postgres
kubectl describe pvc -n postgres
```

### 4. Test Database Connectivity

```bash
kubectl exec -it -n postgres <postgres-pod-name> -- psql -U postgres -c "SELECT 1;"
```

## Resolution Steps

### Step 1: Check if PostgreSQL pod is running

```bash
kubectl get pods -n postgres -l app.kubernetes.io/name=postgresql
```

### Step 2: Check PostgreSQL logs for errors

```bash
kubectl logs -n postgres -l app.kubernetes.io/name=postgresql --tail=100 | grep -i error
```

### Step 3: Common Issues and Fixes

#### Issue: Pod not running
**Cause:** Pod crashed or failed to start  
**Fix:**
```bash
# Restart PostgreSQL
kubectl rollout restart statefulset -n postgres postgres-postgresql
```

#### Issue: Disk full
**Cause:** PVC storage exhausted  
**Fix:**
```bash
# Check PVC usage
kubectl exec -n postgres <pod-name> -- df -h

# Expand PVC if needed
kubectl edit pvc -n postgres data-postgres-postgresql-0
# Increase spec.resources.requests.storage
```

#### Issue: Connection refused
**Cause:** PostgreSQL not accepting connections  
**Fix:**
```bash
# Check pg_hba.conf
kubectl exec -n postgres <pod-name> -- cat /var/lib/postgresql/data/pg_hba.conf

# Check listen_addresses
kubectl exec -n postgres <pod-name> -- psql -U postgres -c "SHOW listen_addresses;"
```

### Step 4: Verify database connectivity from API pod

```bash
kubectl exec -it -n homepage <homepage-api-pod> -- sh -c 'nc -zv $POSTGRES_HOST $POSTGRES_PORT'
```

## Verification

1. Check PostgreSQL is running:
```bash
kubectl get pods -n postgres -l app.kubernetes.io/name=postgresql
```

2. Test database connection:
```bash
kubectl exec -n postgres <pod-name> -- psql -U postgres -c "SELECT version();"
```

3. Verify API can connect:
```bash
kubectl logs -n homepage -l app.kubernetes.io/component=api --tail=20 | grep -i postgres
```

## Prevention

1. Set up automated backups
2. Monitor disk usage
3. Implement connection pooling
4. Use persistent volumes with adequate size
5. Enable PostgreSQL replication for HA

## Related Alerts

- `BrunoSiteDatabaseConnectionFailure`
- `BrunoSiteDatabaseSlowQueries`
- `BrunoSiteExperienceDatabaseUnavailable`

## Escalation

Critical database issues require immediate attention. If unable to resolve within 15 minutes, escalate to database administrator.

## Additional Resources

- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Kubernetes StatefulSets](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/)

