# ⚠️ Runbook: High Connection Pool Usage

## Alert Information
**Alert Name:** `BrunoSiteHighConnectionPoolUsage`  
**Severity:** Warning  

## Symptom
PostgreSQL connection pool usage is above 80%.

## Diagnosis

```bash
# Check active connections
kubectl exec -n postgres <pod> -- psql -U postgres -c "SELECT count(*), state FROM pg_stat_activity GROUP BY state;"

# Check connection limit
kubectl exec -n postgres <pod> -- psql -U postgres -c "SHOW max_connections;"
```

## Resolution
1. Kill idle connections
2. Implement connection pooling (PgBouncer)
3. Increase max_connections if needed
4. Fix application connection leaks
