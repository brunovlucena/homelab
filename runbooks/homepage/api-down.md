# 🚨 Runbook: Bruno Site API Down

## Alert Information

**Alert Name:** `BrunoSiteAPIDown`  
**Severity:** Critical  
**Component:** bruno-site  
**Service:** api

## Symptom

The Bruno Site API has been down for more than 1 minute. All site functionality is affected.

## Impact

- **User Impact:** SEVERE - Complete site outage
- **Business Impact:** HIGH - No visitors can access the homepage or any API endpoints
- **Data Impact:** NONE - No data loss expected

## Diagnosis

### 1. Check Pod Status

```bash
kubectl get pods -n homepage -l app.kubernetes.io/component=api
kubectl describe pod -n homepage -l app.kubernetes.io/component=api
```

### 2. Check Pod Logs

```bash
kubectl logs -n homepage -l app.kubernetes.io/component=api --tail=100
kubectl logs -n homepage -l app.kubernetes.io/component=api --previous  # If pod restarted
```

### 3. Check Service and Endpoints

```bash
kubectl get svc -n homepage homepage-api
kubectl get endpoints -n homepage homepage-api
```

### 4. Check Events

```bash
kubectl get events -n homepage --sort-by='.lastTimestamp' | head -20
```

### 5. Check Resource Limits

```bash
kubectl top pods -n homepage
```

## Resolution Steps

### Step 1: Check if pods are running

```bash
POD_STATUS=$(kubectl get pods -n homepage -l app.kubernetes.io/component=api -o jsonpath='{.items[0].status.phase}')
echo "Pod Status: $POD_STATUS"
```

### Step 2: If pods are not running, check why

```bash
# Check for ImagePullBackOff
kubectl describe pod -n homepage -l app.kubernetes.io/component=api | grep -A 10 "Events:"

# Check for CrashLoopBackOff
kubectl logs -n homepage -l app.kubernetes.io/component=api --previous
```

### Step 3: Common Issues and Fixes

#### Issue: ImagePullBackOff
**Cause:** Cannot pull container image  
**Fix:**
```bash
# Verify image exists
kubectl get deployment -n homepage -o jsonpath='{.spec.template.spec.containers[0].image}'

# Check image pull secrets
kubectl get secrets -n homepage ghcr-secret
```

#### Issue: CrashLoopBackOff
**Cause:** Application failing to start  
**Fix:**
```bash
# Check application logs
kubectl logs -n homepage -l app.kubernetes.io/component=api --tail=50

# Common causes:
# - Database connection failure
# - Missing environment variables
# - Configuration errors
```

#### Issue: OOMKilled
**Cause:** Out of memory  
**Fix:**
```bash
# Increase memory limits
kubectl edit deployment -n homepage homepage-api
# Increase resources.limits.memory
```

### Step 4: Restart the deployment if needed

```bash
kubectl rollout restart deployment/homepage-api -n homepage
kubectl rollout status deployment/homepage-api -n homepage
```

### Step 5: Force reconcile Flux if needed

```bash
flux reconcile helmrelease homepage -n homepage
```

## Verification

1. Check that pods are running:
```bash
kubectl get pods -n homepage -l app.kubernetes.io/component=api
```

2. Verify the service is UP in Prometheus:
```bash
kubectl port-forward -n prometheus svc/prometheus-kube-prometheus-prometheus 9090:9090
# Open http://localhost:9090 and query: up{job="homepage-api"}
```

3. Test the API health endpoint:
```bash
kubectl port-forward -n homepage svc/homepage-api 8080:8080
curl http://localhost:8080/health
```

## Prevention

1. Set up proper resource requests and limits
2. Implement pod disruption budgets
3. Enable horizontal pod autoscaling
4. Monitor pod restart counts
5. Set up liveness and readiness probes correctly

## Related Alerts

- `BrunoSiteHighErrorRate`
- `BrunoSitePodCrashLooping`
- `BrunoSiteHealthCheckFailures`

## Escalation

If the issue persists after following these steps:
1. Check database connectivity
2. Check Redis connectivity
3. Review recent deployments
4. Contact on-call engineer

## Additional Resources

- [Homepage Architecture](../../../ARCHITECTURE.md)
- [Kubernetes Troubleshooting Guide](https://kubernetes.io/docs/tasks/debug/)
- [Homepage API Documentation](../../../flux/clusters/homelab/infrastructure/homepage/README.md)

