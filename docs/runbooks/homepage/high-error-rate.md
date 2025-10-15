# ⚠️ Runbook: Bruno Site API High Error Rate

## Alert Information

**Alert Name:** `BrunoSiteHighErrorRate`  
**Severity:** Critical  
**Component:** bruno-site  
**Service:** api

## Symptom

Bruno Site API error rate is above 5% for the last 5 minutes (5xx errors).

## Impact

- **User Impact:** HIGH - Many requests failing
- **Business Impact:** HIGH - Poor user experience
- **Data Impact:** POTENTIAL - Some operations may fail

## Diagnosis

```bash
# Check error rate in Prometheus
# Query: rate(http_requests_total{job="bruno-site-api",code=~"5.."}[5m])

# Check recent logs for errors
kubectl logs -n homepage -l app.kubernetes.io/component=api --tail=200 | grep -i error

# Check API pod status
kubectl get pods -n homepage -l app.kubernetes.io/component=api
```

## Resolution Steps

### 1. Identify Error Types

```bash
# Check specific error codes
kubectl logs -n homepage -l app.kubernetes.io/component=api --tail=500 | grep " 5[0-9][0-9] "
```

### 2. Common Causes

- **500 Internal Server Error:** Application bugs
- **502 Bad Gateway:** Upstream service issues
- **503 Service Unavailable:** Resource exhaustion
- **504 Gateway Timeout:** Slow backend services

### 3. Check Dependencies

```bash
# Check database connection
kubectl exec -n homepage <api-pod> -- nc -zv postgres-postgresql.postgres.svc.cluster.local 5432

# Check Redis connection
kubectl exec -n homepage <api-pod> -- nc -zv redis-master.redis.svc.cluster.local 6379
```

## Verification

Monitor error rate returns to normal (<1%):
```bash
# In Prometheus: rate(http_requests_total{job="homepage-api",code=~"5.."}[5m])
```

## Related Alerts

- `BrunoSiteAPIDown`
- `BrunoSiteDatabaseDown`
- `BrunoSiteHighResponseTime`

