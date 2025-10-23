# 🚨 Runbook: Agent Bruno API Down

## Alert Information

**Alert Name:** `AgentBrunoAPIDown`  
**Severity:** Critical  
**Component:** agent-bruno  
**Service:** api

## Symptom

The Agent Bruno API has been down for more than 1 minute. All AI assistant functionality is unavailable.

## Impact

- **User Impact:** HIGH - AI assistant unavailable, no chat functionality
- **Business Impact:** MEDIUM - Homepage assistance features disabled
- **Data Impact:** NONE - No data loss expected, memory preserved in Redis/MongoDB

## Diagnosis

### 1. Check Pod Status

```bash
kubectl get pods -n bruno -l app=agent-bruno
kubectl describe pod -n bruno -l app=agent-bruno
```

### 2. Check Pod Logs

```bash
kubectl logs -n bruno -l app=agent-bruno --tail=100
kubectl logs -n bruno -l app=agent-bruno --previous  # If pod restarted
```

### 3. Check Service and Endpoints

```bash
kubectl get svc -n bruno agent-bruno-service
kubectl get endpoints -n bruno agent-bruno-service
```

### 4. Check Events

```bash
kubectl get events -n bruno --sort-by='.lastTimestamp' | head -20
```

### 5. Check Resource Limits

```bash
kubectl top pods -n bruno -l app=agent-bruno
```

### 6. Check HPA Status

```bash
kubectl get hpa -n bruno agent-bruno-hpa
kubectl describe hpa -n bruno agent-bruno-hpa
```

## Resolution Steps

### Step 1: Check if pods are running

```bash
POD_STATUS=$(kubectl get pods -n bruno -l app=agent-bruno -o jsonpath='{.items[0].status.phase}')
echo "Pod Status: $POD_STATUS"
```

### Step 2: If pods are not running, check why

```bash
# Check for ImagePullBackOff
kubectl describe pod -n bruno -l app=agent-bruno | grep -A 10 "Events:"

# Check for CrashLoopBackOff
kubectl logs -n bruno -l app=agent-bruno --previous
```

### Step 3: Common Issues and Fixes

#### Issue: ImagePullBackOff
**Cause:** Cannot pull container image  
**Fix:**
```bash
# Verify image exists
kubectl get deployment -n bruno agent-bruno -o jsonpath='{.spec.template.spec.containers[0].image}'

# Check image pull secrets
kubectl get secrets -n bruno ghcr-secret
kubectl describe secret -n bruno ghcr-secret

# If secret is missing or expired, recreate it
# See scripts/create-secrets.sh for secret creation
```

#### Issue: CrashLoopBackOff
**Cause:** Application failing to start  
**Fix:**
```bash
# Check application logs for startup errors
kubectl logs -n bruno -l app=agent-bruno --tail=50

# Common causes:
# - Redis connection failure (see redis-connection-issues.md)
# - MongoDB connection failure (see mongodb-connection-issues.md)
# - Ollama connection failure (see ollama-connection-issues.md)
# - Missing environment variables
# - Configuration errors
```

#### Issue: OOMKilled
**Cause:** Out of memory  
**Fix:**
```bash
# Check memory usage patterns
kubectl top pods -n bruno -l app=agent-bruno

# Increase memory limits
kubectl edit deployment -n bruno agent-bruno
# Update resources.limits.memory (currently 1Gi)

# Or scale horizontally with HPA
kubectl get hpa -n bruno agent-bruno-hpa
```

#### Issue: Health Check Failures
**Cause:** Liveness/Readiness probes failing  
**Fix:**
```bash
# Check health endpoint directly
kubectl port-forward -n bruno svc/agent-bruno-service 8080:8080
curl http://localhost:8080/health
curl http://localhost:8080/ready

# Check probe configuration
kubectl describe pod -n bruno -l app=agent-bruno | grep -A 10 "Liveness\|Readiness"
```

### Step 4: Restart the deployment if needed

```bash
kubectl rollout restart deployment/agent-bruno -n bruno
kubectl rollout status deployment/agent-bruno -n bruno
```

### Step 5: Force reconcile Flux if needed

```bash
flux reconcile kustomization infrastructure -n flux-system
flux reconcile kustomization agent-bruno -n flux-system
```

## Verification

1. Check that pods are running:
```bash
kubectl get pods -n bruno -l app=agent-bruno
```

2. Verify the service is accessible:
```bash
kubectl port-forward -n bruno svc/agent-bruno-service 8080:8080
curl http://localhost:8080/health
```

3. Test the API endpoints:
```bash
# Health check
curl http://localhost:8080/health

# Readiness check
curl http://localhost:8080/ready

# Knowledge summary
curl http://localhost:8080/knowledge/summary

# Test chat
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello, are you working?"}'
```

4. Verify metrics are being collected:
```bash
curl http://localhost:8080/metrics | grep bruno_
```

## Prevention

1. Set up proper resource requests and limits
2. Implement pod disruption budgets
3. Monitor pod restart counts
4. Set up alerts for health check failures
5. Ensure Redis and MongoDB are highly available
6. Monitor memory usage trends
7. Test image pulls before deployment

## Related Alerts

- `AgentBrunoRedisConnectionFailure`
- `AgentBrunoMongoDBConnectionFailure`
- `AgentBrunoOllamaConnectionFailure`
- `AgentBrunoPodCrashLooping`
- `AgentBrunoHighMemoryUsage`
- `AgentBrunoHealthCheckFailures`

## Escalation

If the issue persists after following these steps:
1. Check Redis connectivity (see redis-connection-issues.md)
2. Check MongoDB connectivity (see mongodb-connection-issues.md)
3. Check Ollama server at 192.168.0.16:11434
4. Review recent deployments
5. Contact on-call engineer

## Additional Resources

- [Agent Bruno Architecture](../../../flux/clusters/homelab/infrastructure/agent-bruno/README.md)
- [Agent Bruno Implementation](../../../flux/clusters/homelab/infrastructure/agent-bruno/IMPLEMENTATION.md)
- [Kubernetes Troubleshooting Guide](https://kubernetes.io/docs/tasks/debug/)
- [FastAPI Health Checks](https://fastapi.tiangolo.com/)

---

**Last Updated:** 2025-10-15  
**Version:** 1.0

