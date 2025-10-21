# 🚨 Runbook: Pod Crash Looping

## Alert Information

**Alert Name:** `BrunoSitePodCrashLooping`  
**Severity:** Critical  
**Component:** bruno-site  
**Service:** kubernetes

## Symptom

Pod is crash looping with restarts in the last 15 minutes.

## Diagnosis

```bash
# Check pod status
kubectl get pods -n homepage -l app.kubernetes.io/component=api

# Check events
kubectl describe pod -n homepage <pod-name>

# Check logs from previous instance
kubectl logs -n homepage <pod-name> --previous

# Check current logs
kubectl logs -n homepage <pod-name> --tail=100
```

## Resolution Steps

### 1. Identify the crash cause

Common causes:
- Application startup failure
- Dependency unavailability (database, Redis)
- Configuration errors
- Out of memory (OOMKilled)
- Liveness probe failures

### 2. Fix based on cause

#### For OOMKilled:
```bash
kubectl edit deployment -n homepage homepage-api
# Increase memory limits
```

#### For configuration errors:
```bash
# Check ConfigMaps and Secrets
kubectl get configmap -n homepage
kubectl get secrets -n homepage

# Verify environment variables
kubectl get deployment -n homepage homepage-api -o yaml | grep -A 20 "env:"
```

#### For dependency issues:
```bash
# Check database is available
kubectl get pods -n postgres

# Check Redis is available
kubectl get pods -n redis
```

### 3. Rollback if recent deployment caused it

```bash
kubectl rollout undo deployment -n homepage homepage-api
```

## Verification

```bash
# Watch pod status
kubectl get pods -n homepage -l app.kubernetes.io/component=api -w

# Check restart count
kubectl get pods -n homepage -l app.kubernetes.io/component=api -o jsonpath='{.items[*].status.containerStatuses[*].restartCount}'
```

## Prevention

1. Test deployments in staging
2. Set appropriate resource limits
3. Implement health checks correctly
4. Use init containers for dependency checks
5. Implement graceful shutdown

