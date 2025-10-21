# ⚠️ Runbook: High CPU Usage

## Alert Information

**Alert Name:** `BrunoSiteHighCPUUsage`  
**Severity:** Warning  
**Component:** bruno-site  
**Service:** api

## Symptom

Bruno Site API CPU usage is above 80% for the last 5 minutes.

## Diagnosis

```bash
# Check CPU usage
kubectl top pods -n homepage -l app.kubernetes.io/component=api

# Check for CPU-intensive operations in logs
kubectl logs -n homepage -l app.kubernetes.io/component=api --tail=200
```

## Resolution Steps

### 1. Identify CPU-intensive operations

Check for:
- Complex computations
- JSON parsing/serialization
- Regular expressions
- Image processing

### 2. Scale horizontally

```bash
kubectl scale deployment -n homepage homepage-api --replicas=3
```

### 3. Increase CPU limits if needed

```bash
kubectl edit deployment -n homepage homepage-api
# Increase resources.limits.cpu
```

## Prevention

1. Optimize algorithms
2. Use caching
3. Implement rate limiting
4. Profile the application

