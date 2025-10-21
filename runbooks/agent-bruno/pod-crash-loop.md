# 🚨 Runbook: Agent Bruno Pod Crash Loop

## Alert Information

**Alert Name:** `AgentBrunoPodCrashLooping`  
**Severity:** Critical  
**Component:** agent-bruno  
**Service:** pod-stability

## Symptom

Agent Bruno pod is repeatedly crashing and restarting (CrashLoopBackOff state).

## Impact

- **User Impact:** SEVERE - Service completely unavailable
- **Business Impact:** HIGH - AI assistant functionality down
- **Data Impact:** NONE - Data preserved in Redis/MongoDB

## Diagnosis

### 1. Check Pod Status

```bash
# Check pod status
kubectl get pods -n bruno -l app=agent-bruno

# Look for CrashLoopBackOff status
kubectl describe pod -n bruno -l app=agent-bruno
```

### 2. Check Recent Logs

```bash
# Current logs
kubectl logs -n bruno -l app=agent-bruno --tail=100

# Previous container logs (most important!)
kubectl logs -n bruno -l app=agent-bruno --previous
```

### 3. Check Events

```bash
# Recent events
kubectl get events -n bruno --sort-by='.lastTimestamp' | grep agent-bruno | head -20

# Detailed pod events
kubectl describe pod -n bruno -l app=agent-bruno | grep -A 20 "Events:"
```

### 4. Check Restart Count

```bash
# Get restart count
kubectl get pods -n bruno -l app=agent-bruno -o jsonpath='{.items[*].status.containerStatuses[*].restartCount}'

# Check restart reasons
kubectl get pods -n bruno -l app=agent-bruno -o jsonpath='{.items[*].status.containerStatuses[*].lastState.terminated.reason}'
```

## Resolution Steps

### Step 1: Identify Crash Cause

Check previous container logs for error messages:

```bash
kubectl logs -n bruno -l app=agent-bruno --previous | tail -50
```

### Step 2: Common Crash Causes and Fixes

#### Issue: Application Startup Failure
**Cause:** Python import errors, missing dependencies  
**Symptoms:**
```
ModuleNotFoundError: No module named 'fastapi'
ImportError: cannot import name 'X' from 'Y'
```

**Fix:**
```bash
# Check Dockerfile and ensure all dependencies are installed
# Rebuild image with correct dependencies
cd /Users/brunolucena/workspace/bruno/repos/homelab/flux/clusters/homelab/infrastructure/agent-bruno
make docker-build
make docker-push

# Update deployment to use new image
kubectl rollout restart deployment/agent-bruno -n bruno
```

#### Issue: Redis Connection Failure
**Cause:** Cannot connect to Redis on startup  
**Symptoms:**
```
redis.exceptions.ConnectionError: Error connecting to Redis
Connection refused: redis-master.redis.svc.cluster.local:6379
```

**Fix:**
```bash
# Check if Redis is running
kubectl get pods -n redis

# If Redis is down, restart it
flux reconcile helmrelease redis -n redis

# Wait for Redis to be ready
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=redis -n redis --timeout=300s

# Then restart Agent Bruno
kubectl rollout restart deployment/agent-bruno -n bruno
```

See also: [redis-connection-issues.md](./redis-connection-issues.md)

#### Issue: MongoDB Connection Failure
**Cause:** Cannot connect to MongoDB on startup  
**Symptoms:**
```
pymongo.errors.ServerSelectionTimeoutError: mongodb.mongodb.svc.cluster.local:27017
ConnectionError: Unable to connect to MongoDB
```

**Fix:**
```bash
# Check if MongoDB is running
kubectl get pods -n mongodb

# If MongoDB is down, restart it
flux reconcile helmrelease mongodb -n mongodb

# Wait for MongoDB to be ready
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=mongodb -n mongodb --timeout=300s

# Then restart Agent Bruno
kubectl rollout restart deployment/agent-bruno -n bruno
```

See also: [mongodb-connection-issues.md](./mongodb-connection-issues.md)

#### Issue: Configuration Error
**Cause:** Invalid environment variables or configuration  
**Symptoms:**
```
KeyError: 'REQUIRED_ENV_VAR'
ValueError: invalid literal for int()
```

**Fix:**
```bash
# Check current environment variables
kubectl get deployment -n bruno agent-bruno -o jsonpath='{.spec.template.spec.containers[0].env[*]}' | jq

# Verify all required variables are set:
kubectl get deployment -n bruno agent-bruno -o jsonpath='{.spec.template.spec.containers[0].env[?(@.name=="REDIS_URL")].value}'
kubectl get deployment -n bruno agent-bruno -o jsonpath='{.spec.template.spec.containers[0].env[?(@.name=="MONGODB_URL")].value}'
kubectl get deployment -n bruno agent-bruno -o jsonpath='{.spec.template.spec.containers[0].env[?(@.name=="OLLAMA_URL")].value}'

# Fix any incorrect values
kubectl edit deployment -n bruno agent-bruno
```

#### Issue: OOMKilled (Out of Memory)
**Cause:** Application exceeding memory limits  
**Symptoms:**
```
# Check lastState
kubectl get pods -n bruno -l app=agent-bruno -o jsonpath='{.items[*].status.containerStatuses[*].lastState.terminated.reason}'
# Output: OOMKilled
```

**Fix:**
```bash
# Increase memory limits
kubectl edit deployment -n bruno agent-bruno

# Update resources:
spec:
  template:
    spec:
      containers:
      - name: agent-bruno
        resources:
          requests:
            memory: 512Mi  # Increase from 256Mi
          limits:
            memory: 2Gi    # Increase from 1Gi
```

See also: [high-memory-usage.md](./high-memory-usage.md)

#### Issue: Liveness Probe Failure
**Cause:** Health endpoint not responding in time  
**Symptoms:**
```
Liveness probe failed: HTTP probe failed with statuscode: 500
Liveness probe failed: Get "http://10.244.0.1:8080/health": context deadline exceeded
```

**Fix:**
```bash
# Increase probe timing
kubectl edit deployment -n bruno agent-bruno

# Update livenessProbe:
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 60  # Increase from 30
  periodSeconds: 15        # Increase from 10
  timeoutSeconds: 10       # Increase from 5
  failureThreshold: 5      # Increase from 3
```

#### Issue: Permission Denied
**Cause:** Container user lacks required permissions  
**Symptoms:**
```
PermissionError: [Errno 13] Permission denied: '/path/to/file'
```

**Fix:**
```bash
# Check if running as non-root user
kubectl get deployment -n bruno agent-bruno -o jsonpath='{.spec.template.spec.securityContext}'

# Update Dockerfile to set correct permissions
# Or update deployment securityContext
kubectl edit deployment -n bruno agent-bruno
```

#### Issue: Port Already in Use
**Cause:** Port 8080 already bound  
**Symptoms:**
```
OSError: [Errno 98] Address already in use
```

**Fix:**
```bash
# Check if multiple processes trying to bind same port
# Usually indicates duplicate uvicorn processes
# Ensure proper process management in Dockerfile

# Check Dockerfile CMD/ENTRYPOINT
# Should be: CMD ["uvicorn", "src.main:app", "--host", "0.0.0.0", "--port", "8080"]
```

#### Issue: Ollama Connection Timeout on Startup
**Cause:** Cannot reach Ollama server during startup health check  
**Symptoms:**
```
requests.exceptions.Timeout: HTTPConnectionPool(host='192.168.0.16', port=11434)
Connection timeout to Ollama server
```

**Fix:**
```bash
# Check if startup health check requires Ollama
# Consider making Ollama check non-blocking on startup
# Or increase startup timeout

# Verify Ollama is accessible
curl http://192.168.0.16:11434/api/version

# If Ollama is required, ensure it's up first
```

See also: [ollama-connection-issues.md](./ollama-connection-issues.md)

### Step 3: Check Image

```bash
# Verify image exists and is pullable
kubectl describe pod -n bruno -l app=agent-bruno | grep "Image:"

# Check image pull status
kubectl describe pod -n bruno -l app=agent-bruno | grep -A 5 "Events:"

# If ImagePullBackOff, check image pull secret
kubectl get secret -n bruno ghcr-secret
kubectl describe secret -n bruno ghcr-secret
```

### Step 4: Temporary Debug Mode

```bash
# Deploy debug version that just sleeps (to inspect environment)
kubectl set image deployment/agent-bruno agent-bruno=busybox:latest -n bruno
kubectl set command deployment/agent-bruno agent-bruno -- sleep 3600 -n bruno

# Exec into pod to debug
kubectl exec -it -n bruno deployment/agent-bruno -- sh

# Check environment
env | grep -E "REDIS|MONGODB|OLLAMA"

# Test connectivity
ping redis-master.redis.svc.cluster.local
ping mongodb.mongodb.svc.cluster.local
ping 192.168.0.16

# Restore original image when done
kubectl set image deployment/agent-bruno agent-bruno=ghcr.io/brunovlucena/agent-bruno:latest -n bruno
```

## Verification

1. Check pod is running:
```bash
kubectl get pods -n bruno -l app=agent-bruno
# Status should be Running
```

2. Verify no recent restarts:
```bash
watch kubectl get pods -n bruno -l app=agent-bruno
# Restart count should stop increasing
```

3. Check logs are healthy:
```bash
kubectl logs -n bruno -l app=agent-bruno --tail=50
# Should show successful startup messages
```

4. Test endpoints:
```bash
kubectl port-forward -n bruno svc/agent-bruno-service 8080:8080
curl http://localhost:8080/health
curl http://localhost:8080/ready
```

## Prevention

1. **Startup Dependencies**
   - Implement retry logic for external dependencies
   - Use init containers to wait for dependencies
   - Make external connections non-blocking on startup

2. **Health Checks**
   - Set appropriate probe timing
   - Separate liveness from readiness checks
   - Don't include slow operations in health checks

3. **Resource Management**
   - Set appropriate memory limits
   - Monitor resource usage trends
   - Use HPA for scaling

4. **Configuration Validation**
   - Validate environment variables on startup
   - Provide clear error messages
   - Use configuration validation library

5. **Testing**
   - Test startup in local environment
   - Verify all dependencies are available
   - Test with realistic resource limits
   - Implement integration tests

6. **Monitoring**
   - Alert on high restart counts
   - Monitor crash patterns
   - Track error rates by type
   - Set up log aggregation

## Debug Checklist

- [ ] Check previous container logs
- [ ] Verify all environment variables are set
- [ ] Confirm Redis is running and accessible
- [ ] Confirm MongoDB is running and accessible
- [ ] Confirm Ollama server is accessible
- [ ] Check resource limits (memory, CPU)
- [ ] Verify image exists and is pullable
- [ ] Check health probe configuration
- [ ] Review recent code/config changes
- [ ] Check cluster events for issues

## Related Alerts

- `AgentBrunoAPIDown`
- `AgentBrunoRedisConnectionFailure`
- `AgentBrunoMongoDBConnectionFailure`
- `AgentBrunoOllamaConnectionFailure`
- `AgentBrunoOOMKilled`
- `AgentBrunoHighRestartRate`

## Escalation

If unable to resolve within 15 minutes:
1. Review recent deployments/changes
2. Check cluster-wide issues
3. Verify all dependencies (Redis, MongoDB, Ollama)
4. Review application code changes
5. Contact development team
6. Consider rollback to previous version

## Rollback Procedure

```bash
# Check deployment history
kubectl rollout history deployment/agent-bruno -n bruno

# Rollback to previous version
kubectl rollout undo deployment/agent-bruno -n bruno

# Or rollback to specific revision
kubectl rollout undo deployment/agent-bruno -n bruno --to-revision=2

# Monitor rollback
kubectl rollout status deployment/agent-bruno -n bruno
```

## Additional Resources

- [Kubernetes Pod Troubleshooting](https://kubernetes.io/docs/tasks/debug/debug-application/debug-pods/)
- [CrashLoopBackOff Solutions](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/)
- [Agent Bruno Architecture](../../../flux/clusters/homelab/infrastructure/agent-bruno/README.md)
- [FastAPI Startup Events](https://fastapi.tiangolo.com/advanced/events/)

---

**Last Updated:** 2025-10-15  
**Version:** 1.0

