# 🚨 Runbook: Health Check Failures

## Alert Information
**Alert Name:** `BrunoSiteHealthCheckFailures`  
**Severity:** Critical  

## Symptom
Bruno Site API health checks are failing.

## Diagnosis

```bash
# Test health endpoint manually
kubectl port-forward -n homepage svc/homepage-api 8080:8080
curl -v http://localhost:8080/health

# Check pod logs
kubectl logs -n homepage -l app.kubernetes.io/component=api --tail=100 | grep -i health

# Check pod status
kubectl get pods -n homepage -l app.kubernetes.io/component=api
```

## Resolution

### 1. Check Health Endpoint Implementation

Verify the /health endpoint is:
- Responding within timeout
- Checking dependencies correctly
- Not too strict (causing false positives)

### 2. Check Dependencies

```bash
# Check database
kubectl get pods -n postgres

# Check Redis
kubectl get pods -n redis
```

### 3. Adjust Health Check if Needed

```bash
kubectl edit deployment -n homepage homepage-api
# Adjust: initialDelaySeconds, periodSeconds, timeoutSeconds, failureThreshold
```

## Prevention
1. Implement proper health check logic
2. Don't make health checks too sensitive
3. Allow adequate startup time
4. Monitor health check response times
