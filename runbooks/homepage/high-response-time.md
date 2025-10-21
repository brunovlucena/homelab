# ⚠️ Runbook: High Response Time

## Alert Information

**Alert Name:** `BrunoSiteHighResponseTime`  
**Severity:** Warning  
**Component:** bruno-site  
**Service:** api

## Symptom

Bruno Site API 95th percentile response time is greater than 1 second for the last 5 minutes.

## Diagnosis

```bash
# Check current response times
kubectl logs -n homepage -l app.kubernetes.io/component=api --tail=100 | grep "duration"

# Check resource usage
kubectl top pods -n homepage

# Check database query performance
kubectl exec -n postgres <pod> -- psql -U postgres -c "SELECT query, calls, mean_exec_time, max_exec_time FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 10;"
```

## Resolution Steps

### 1. Identify Slow Endpoints

Check OpenTelemetry traces in Grafana to identify slow endpoints.

### 2. Check for N+1 Queries

Review database queries for inefficient patterns.

### 3. Scale if Needed

```bash
kubectl scale deployment -n homepage homepage-api --replicas=3
```

## Prevention

1. Add database indexes
2. Implement caching
3. Optimize queries
4. Use connection pooling
5. Enable query result caching

