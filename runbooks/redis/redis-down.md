# 🚨 Runbook: Redis Service Down

## Alert Information

**Alert Name:** `RedisDown`  
**Severity:** Critical  
**Component:** redis  
**Service:** redis-master

## Symptom

Redis service is completely unavailable. All applications depending on Redis are experiencing failures.

## Impact

- **User Impact:** HIGH - Session management, caching, and real-time features unavailable
- **Business Impact:** HIGH - Multiple services affected, degraded user experience
- **Data Impact:** MEDIUM - In-memory data lost if persistence not configured properly

## Diagnosis

### 1. Check Redis Pod Status

```bash
kubectl get pods -n redis
kubectl get pods -n redis -l app.kubernetes.io/name=redis -o wide
```

**Expected Output:**
```
NAME             READY   STATUS    RESTARTS   AGE
redis-master-0   1/1     Running   0          24h
```

### 2. Check Redis Service and Endpoints

```bash
kubectl get svc -n redis
kubectl get endpoints -n redis redis-master
```

### 3. Check Redis StatefulSet

```bash
kubectl get statefulset -n redis
kubectl describe statefulset redis-master -n redis
```

### 4. Check Recent Events

```bash
kubectl get events -n redis --sort-by='.lastTimestamp' | tail -20
```

### 5. Check Redis Logs

```bash
kubectl logs -n redis statefulset/redis-master --tail=100
kubectl logs -n redis statefulset/redis-master --previous  # If pod restarted
```

### 6. Check Resource Usage

```bash
kubectl top pod -n redis
kubectl describe node $(kubectl get pod -n redis redis-master-0 -o jsonpath='{.spec.nodeName}')
```

## Resolution Steps

### Step 1: Identify the Root Cause

Check the pod status and logs to understand why Redis is down:

```bash
# Get detailed pod status
kubectl describe pod -n redis redis-master-0

# Check for OOMKilled
kubectl get pod -n redis redis-master-0 -o jsonpath='{.status.containerStatuses[0].lastState.terminated.reason}'

# Check for CrashLoopBackOff
kubectl get pod -n redis redis-master-0 -o jsonpath='{.status.containerStatuses[0].state.waiting.reason}'
```

### Step 2: Common Issues and Fixes

#### Issue: Pod OOMKilled (Out of Memory)
**Cause:** Redis exceeded memory limits  
**Fix:**
```bash
# Check current memory limits
kubectl get pod -n redis redis-master-0 -o jsonpath='{.spec.containers[0].resources.limits.memory}'

# Check Redis memory usage from logs
kubectl logs -n redis redis-master-0 --tail=200 | grep -i "memory\|oom"

# Increase memory limits in HelmRelease
kubectl edit helmrelease redis -n redis
# Update:
#   resources:
#     limits:
#       memory: 512Mi  # Increase as needed
#     requests:
#       memory: 256Mi

# Or use Flux to update
flux reconcile helmrelease redis -n redis
```

#### Issue: Persistent Volume Issues
**Cause:** PVC not bound or disk full  
**Fix:**
```bash
# Check PVC status
kubectl get pvc -n redis
kubectl describe pvc -n redis redis-data-redis-master-0

# Check PV status
kubectl get pv | grep redis

# If PVC pending, check storage class
kubectl get storageclass

# Check disk usage
kubectl exec -n redis redis-master-0 -- df -h /data

# If disk full, may need to clean old RDB/AOF files or increase PVC size
```

#### Issue: Configuration Error
**Cause:** Invalid Redis configuration  
**Fix:**
```bash
# Check Redis config
kubectl exec -n redis redis-master-0 -- redis-cli config get '*'

# Check ConfigMap
kubectl get configmap -n redis
kubectl describe configmap -n redis redis-configuration

# Validate redis.conf syntax
kubectl exec -n redis redis-master-0 -- cat /etc/redis/redis.conf

# Rollback to previous working version
flux get helmreleases -n redis
flux suspend helmrelease redis -n redis
flux resume helmrelease redis -n redis
```

#### Issue: Node Issues
**Cause:** Node where Redis is running has problems  
**Fix:**
```bash
# Check node status
kubectl get nodes
kubectl describe node $(kubectl get pod -n redis redis-master-0 -o jsonpath='{.spec.nodeName}')

# If node NotReady, cordon and drain
kubectl cordon <node-name>
kubectl delete pod -n redis redis-master-0  # Forces rescheduling

# StatefulSet will recreate pod on healthy node
```

#### Issue: Image Pull Error
**Cause:** Cannot pull Redis image  
**Fix:**
```bash
# Check image pull status
kubectl describe pod -n redis redis-master-0 | grep -A 10 "Events:"

# Verify image exists
kubectl get pod -n redis redis-master-0 -o jsonpath='{.spec.containers[0].image}'

# Check image pull secrets
kubectl get secrets -n redis | grep docker

# Manual pull test
kubectl run test --image=redis:7.2 --rm -it --restart=Never -n redis -- redis-cli --version
```

#### Issue: Network Policy Blocking
**Cause:** Network policies preventing Redis from starting  
**Fix:**
```bash
# Check network policies
kubectl get networkpolicies -n redis
kubectl describe networkpolicies -n redis

# Temporarily disable to test
kubectl delete networkpolicy -n redis <policy-name>

# Test connectivity
kubectl run test -n redis --image=redis:7.2 --rm -it --restart=Never -- redis-cli -h redis-master ping
```

### Step 3: Force Redis Restart

If no clear issue found, force restart:

```bash
# Delete the pod (StatefulSet will recreate)
kubectl delete pod -n redis redis-master-0

# Wait for pod to come up
kubectl wait --for=condition=ready pod -n redis redis-master-0 --timeout=5m

# Check pod status
kubectl get pod -n redis redis-master-0
```

### Step 4: Rollout Restart StatefulSet

```bash
kubectl rollout restart statefulset redis-master -n redis
kubectl rollout status statefulset redis-master -n redis
```

### Step 5: Force Flux Reconciliation

```bash
# Reconcile Flux HelmRelease
flux reconcile helmrelease redis -n redis --force

# Check reconciliation status
flux get helmreleases -n redis
```

### Step 6: Emergency Recovery - Reinstall Redis

⚠️ **Warning:** This will cause data loss if persistence is not properly configured!

```bash
# Suspend Flux HelmRelease
flux suspend helmrelease redis -n redis

# Delete StatefulSet (keeps PVC)
kubectl delete statefulset redis-master -n redis --cascade=orphan

# Resume Flux to recreate
flux resume helmrelease redis -n redis
flux reconcile helmrelease redis -n redis

# Monitor recreation
watch kubectl get pods -n redis
```

## Verification

### 1. Check Redis Pod is Running

```bash
kubectl get pod -n redis redis-master-0
# Should show: Running and 1/1 READY
```

### 2. Test Redis Connectivity

```bash
# Test PING
kubectl exec -n redis redis-master-0 -- redis-cli ping
# Should return: PONG

# Test INFO
kubectl exec -n redis redis-master-0 -- redis-cli info server
```

### 3. Test from Client Pods

```bash
# Test from agent-bruno
kubectl exec -n bruno deployment/agent-bruno -- sh -c 'nc -zv redis-master.redis.svc.cluster.local 6379'

# Test from homepage
kubectl exec -n homepage deployment/homepage-api -- sh -c 'nc -zv redis-master.redis.svc.cluster.local 6379'
```

### 4. Check Redis Performance

```bash
# Check latency
kubectl exec -n redis redis-master-0 -- redis-cli --latency -h localhost -p 6379

# Check memory usage
kubectl exec -n redis redis-master-0 -- redis-cli info memory

# Check connected clients
kubectl exec -n redis redis-master-0 -- redis-cli client list
```

### 5. Verify Data Persistence

```bash
# Check if AOF/RDB enabled
kubectl exec -n redis redis-master-0 -- redis-cli config get save
kubectl exec -n redis redis-master-0 -- redis-cli config get appendonly

# Check last save time
kubectl exec -n redis redis-master-0 -- redis-cli lastsave

# List persistence files
kubectl exec -n redis redis-master-0 -- ls -lh /data/
```

### 6. Monitor Application Recovery

```bash
# Check application logs
kubectl logs -n bruno -l app=agent-bruno --tail=50 | grep -i redis
kubectl logs -n homepage -l app=homepage-api --tail=50 | grep -i redis

# Check application metrics
kubectl port-forward -n bruno svc/agent-bruno-service 8080:8080
curl http://localhost:8080/metrics | grep redis
```

## Prevention

### 1. Resource Management

```yaml
# In Redis HelmRelease, configure proper resources
master:
  resources:
    requests:
      memory: "256Mi"
      cpu: "100m"
    limits:
      memory: "512Mi"
      cpu: "500m"
```

### 2. Enable Persistence

```yaml
master:
  persistence:
    enabled: true
    size: 8Gi
    storageClass: "standard"
```

### 3. Configure High Availability

```yaml
# Enable Redis Sentinel
sentinel:
  enabled: true
  replicas: 3
replica:
  replicaCount: 2
```

### 4. Set Up Monitoring

- Monitor Redis pod availability
- Alert on Redis high memory usage
- Alert on Redis slow operations
- Monitor persistence health

### 5. Regular Backups

```bash
# Schedule automated RDB backups
kubectl exec -n redis redis-master-0 -- redis-cli bgsave

# Export data regularly
kubectl exec -n redis redis-master-0 -- redis-cli --rdb /data/backup-$(date +%Y%m%d).rdb
```

### 6. Resource Quotas

```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: redis-quota
  namespace: redis
spec:
  hard:
    requests.memory: "1Gi"
    limits.memory: "2Gi"
```

### 7. Pod Disruption Budget

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: redis-pdb
  namespace: redis
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: redis
```

## Performance Tips

1. **Memory Management:** Configure maxmemory and eviction policies
2. **Persistence:** Choose between RDB (snapshots) and AOF (append-only file)
3. **Connection Pooling:** Use connection pools in client applications
4. **Key Naming:** Use consistent, structured key naming conventions
5. **TTL Strategy:** Set appropriate TTLs on temporary keys
6. **Monitoring:** Use Redis INFO command for health checks

## Related Alerts

- `RedisHighMemory`
- `RedisSlowOperations`
- `RedisPersistenceIssues`
- `RedisReplicationLag`
- `AgentBrunoRedisConnectionFailure`
- `HomepageRedisDown`

## Escalation

If Redis cannot be restored within 15 minutes:

1. ✅ Check all resolution steps above
2. 🔍 Review Redis StatefulSet YAML configuration
3. 📊 Analyze node health and cluster capacity
4. 💾 Verify PVC and storage backend health
5. 🔄 Consider emergency failover to replica (if HA enabled)
6. 📞 Contact infrastructure team
7. 🆘 Page on-call engineer for critical production impact

## Additional Resources

- [Redis Documentation](https://redis.io/docs/)
- [Redis Troubleshooting Guide](https://redis.io/docs/management/optimization/)
- [Kubernetes StatefulSet Concepts](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/)
- [Redis Persistence Documentation](https://redis.io/docs/management/persistence/)
- [Redis Sentinel Documentation](https://redis.io/docs/management/sentinel/)

## Quick Commands Reference

```bash
# Health check
kubectl exec -n redis redis-master-0 -- redis-cli ping

# Get info
kubectl exec -n redis redis-master-0 -- redis-cli info

# Check memory
kubectl exec -n redis redis-master-0 -- redis-cli info memory

# Check clients
kubectl exec -n redis redis-master-0 -- redis-cli client list

# Force save
kubectl exec -n redis redis-master-0 -- redis-cli save

# Check persistence
kubectl exec -n redis redis-master-0 -- ls -lh /data/

# Restart pod
kubectl delete pod -n redis redis-master-0

# Force reconcile
flux reconcile helmrelease redis -n redis
```

---

**Last Updated:** 2025-10-15  
**Version:** 1.0  
**Owner:** Infrastructure Team

