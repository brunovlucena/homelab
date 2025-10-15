# ⚠️ Runbook: Slow Database Queries

## Alert Information
**Alert Name:** `BrunoSiteSlowDatabaseQueries`  
**Severity:** Warning  

## Symptom
PostgreSQL database 95th percentile query performance is fetching more than 1000 tuples per second.

## Diagnosis

```bash
# Check slow queries
kubectl exec -n postgres <pod> -- psql -U postgres -c "SELECT query, calls, mean_exec_time, max_exec_time FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 10;"

# Check table statistics
kubectl exec -n postgres <pod> -- psql -U postgres -d bruno_site -c "SELECT relname, seq_scan, seq_tup_read, idx_scan, idx_tup_fetch FROM pg_stat_user_tables ORDER BY seq_scan DESC;"
```

## Resolution
1. Identify slow queries
2. Add missing indexes
3. Optimize query structure
4. Run VACUUM ANALYZE
