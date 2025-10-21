# ⚠️ Runbook: Database Connection Issues

## Alert Information

**Alert Name:** `BrunoSiteDatabaseConnectionFailure`  
**Severity:** Critical  
**Component:** bruno-site  
**Service:** database

## Symptom

PostgreSQL database has more than 10 deadlocks in the last 5 minutes.

## Diagnosis

```bash
# Check deadlocks
kubectl exec -n postgres <pod> -- psql -U postgres -c "SELECT * FROM pg_stat_database WHERE datname='bruno_site';"

# Check active connections
kubectl exec -n postgres <pod> -- psql -U postgres -c "SELECT count(*) FROM pg_stat_activity;"

# Check for long-running queries
kubectl exec -n postgres <pod> -- psql -U postgres -c "SELECT pid, now() - pg_stat_activity.query_start AS duration, query FROM pg_stat_activity WHERE state = 'active' ORDER BY duration DESC;"
```

## Resolution Steps

### 1. Kill long-running queries if needed

```bash
kubectl exec -n postgres <pod> -- psql -U postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE state = 'active' AND now() - pg_stat_activity.query_start > interval '5 minutes';"
```

### 2. Check connection pool settings

Review `max_connections` and connection pooling configuration.

## Prevention

1. Implement proper connection pooling
2. Use prepared statements
3. Add proper indexes
4. Monitor query performance

