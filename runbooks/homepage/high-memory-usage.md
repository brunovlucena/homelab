# ⚠️ Runbook: High Memory Usage

## Alert Information

**Alert Name:** `BrunoSiteHighMemoryUsage`  
**Severity:** Warning  
**Component:** bruno-site  
**Service:** api

## Symptom

Bruno Site API memory usage is above 90% of the limit for the last 5 minutes.

## Diagnosis

```bash
# Check memory usage
kubectl top pods -n homepage -l app.kubernetes.io/component=api

# Check for memory leaks in application logs
kubectl logs -n homepage -l app.kubernetes.io/component=api --tail=100 | grep -i "memory\|oom"
```

## Resolution Steps

### 1. Restart pod to reclaim memory

```bash
kubectl rollout restart deployment -n homepage homepage-api
```

### 2. Increase memory limits

```bash
kubectl edit deployment -n homepage homepage-api
# Increase resources.limits.memory
```

### 3. Check for memory leaks

Review application code for:
- Unclosed connections
- Growing arrays/maps
- Goroutine leaks (if Go)
- Event listener leaks

## Prevention

1. Implement proper garbage collection
2. Close connections properly
3. Use memory profiling tools
4. Set appropriate memory limits
5. Monitor memory growth over time

