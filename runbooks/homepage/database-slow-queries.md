# ⚠️ Runbook: Database Slow Queries

## Alert Information
**Alert Name:** `BrunoSiteDatabaseSlowQueries`  
**Severity:** Warning  

## Symptom
PostgreSQL database shows inefficient query patterns with low tuple return ratio.

## Diagnosis
```bash
# Check slow queries
kubectl exec -n postgres <pod> -- psql -U postgres -c "SELECT query, calls, mean_exec_time FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 20;"

# Check missing indexes
kubectl exec -n postgres <pod> -- psql -U postgres -d bruno_site -c "SELECT schemaname, tablename, attname, n_distinct, correlation FROM pg_stats WHERE schemaname = 'public' ORDER BY abs(correlation) DESC;"
```

## Resolution
1. Add missing indexes
2. Optimize queries with EXPLAIN ANALYZE
3. Update table statistics: `ANALYZE;`
4. Consider query rewriting

## Prevention
- Regular VACUUM and ANALYZE
- Monitor pg_stat_statements
- Use connection pooling
