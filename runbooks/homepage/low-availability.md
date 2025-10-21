# 🚨 Runbook: Low Availability

## Alert Information
**Alert Name:** `BrunoSiteLowAvailability`  
**Severity:** Critical  

## Symptom
Bruno Site API availability is below 99% threshold for the last hour.

## Diagnosis

```bash
# Check error rate
# In Prometheus: rate(http_requests_total{job="homepage-api",code=~"2.."}[1h]) / rate(http_requests_total{job="homepage-api"}[1h])

# Check recent errors
kubectl logs -n homepage -l app.kubernetes.io/component=api --tail=500 | grep -E ' [45][0-9][0-9] '
```

## Resolution
1. Identify error causes (database, dependencies, application errors)
2. Fix underlying issues
3. Scale if needed
4. Review recent deployments
