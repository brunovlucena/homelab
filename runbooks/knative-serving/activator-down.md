# 🚨 Runbook: Knative Activator Down

## Alert Information

**Alert Name:** `KnativeActivatorDown`  
**Severity:** Critical  
**Component:** knative-serving / activator  
**Impact:** Cannot scale services from zero

## Symptom

The Knative Activator component is unavailable or not responding. Services scaled to zero cannot receive traffic and will not scale up.

## Impact

- **User Impact:** CRITICAL - Services at zero replicas cannot handle traffic
- **Business Impact:** HIGH - Cold starts failing, user requests timing out
- **Data Impact:** LOW - No data loss, but service access blocked

## Diagnosis

### 1. Check Activator Pod Status

```bash
kubectl get pods -n knative-serving -l app=activator
kubectl get pods -n knative-serving -l app=activator -o wide
```

**Expected Output:**
```
NAME                         READY   STATUS    RESTARTS   AGE
activator-xxxxxxxxxx-xxxxx   1/1     Running   0          24h
```

### 2. Check Activator Deployment

```bash
kubectl describe deployment -n knative-serving activator
kubectl get deployment -n knative-serving activator -o yaml
```

### 3. Check Activator Logs

```bash
# Recent logs
kubectl logs -n knative-serving -l app=activator --tail=100

# Previous container logs (if crashed)
kubectl logs -n knative-serving -l app=activator --tail=100 --previous
```

### 4. Check Activator Service

```bash
kubectl get svc -n knative-serving activator-service
kubectl describe svc -n knative-serving activator-service
```

### 5. Check Resource Usage

```bash
kubectl top pods -n knative-serving -l app=activator
```

### 6. Check Endpoints

```bash
kubectl get endpoints -n knative-serving activator-service
```

### 7. Check Recent Events

```bash
kubectl get events -n knative-serving --field-selector involvedObject.name=activator --sort-by='.lastTimestamp'
```

## Resolution Steps

### Step 1: Identify Root Cause

Check pod status for common issues:

```bash
# Get pod details
kubectl describe pod -n knative-serving -l app=activator

# Common indicators:
# - CrashLoopBackOff: Application error
# - ImagePullBackOff: Image issue
# - OOMKilled: Out of memory
# - Pending: Resource constraints
```

### Step 2: Common Issues and Fixes

#### Issue: Pod CrashLoopBackOff
**Cause:** Application error or misconfiguration  
**Fix:**
```bash
# Check logs for error messages
kubectl logs -n knative-serving -l app=activator --tail=200

# Check previous container logs
kubectl logs -n knative-serving -l app=activator --tail=200 --previous

# Check for configuration issues
kubectl get configmap -n knative-serving config-network -o yaml
kubectl get configmap -n knative-serving config-observability -o yaml

# Restart activator
kubectl rollout restart deployment -n knative-serving activator
```

#### Issue: Pod OOMKilled
**Cause:** Insufficient memory allocation  
**Fix:**
```bash
# Check current resource limits
kubectl get deployment -n knative-serving activator -o yaml | grep -A 10 resources

# Increase memory limits via KnativeServing CR or directly
kubectl patch deployment -n knative-serving activator -p '{"spec":{"template":{"spec":{"containers":[{"name":"activator","resources":{"limits":{"memory":"1Gi"},"requests":{"memory":"512Mi"}}}]}}}}'

# Wait for rollout
kubectl rollout status deployment -n knative-serving activator
```

#### Issue: ImagePullBackOff
**Cause:** Cannot pull activator image  
**Fix:**
```bash
# Check image details
kubectl describe pod -n knative-serving -l app=activator | grep -A 5 "Image:"

# Check image pull secrets
kubectl get serviceaccount -n knative-serving default -o yaml

# Force image pull by restarting
kubectl rollout restart deployment -n knative-serving activator
```

#### Issue: No Endpoints Available
**Cause:** Service not routing to pods  
**Fix:**
```bash
# Check service endpoints
kubectl get endpoints -n knative-serving activator-service

# Check service selector matches pods
kubectl get svc -n knative-serving activator-service -o yaml | grep -A 3 selector
kubectl get pods -n knative-serving -l app=activator --show-labels

# Verify pod is ready
kubectl get pods -n knative-serving -l app=activator -o jsonpath='{.items[*].status.conditions[?(@.type=="Ready")].status}'
```

#### Issue: High CPU/Memory Usage
**Cause:** Activator overloaded with traffic  
**Fix:**
```bash
# Check current resource usage
kubectl top pods -n knative-serving -l app=activator

# Scale activator horizontally
kubectl scale deployment -n knative-serving activator --replicas=2

# Or patch for permanent scaling
kubectl patch deployment -n knative-serving activator -p '{"spec":{"replicas":2}}'

# Monitor new pods
kubectl get pods -n knative-serving -l app=activator -w
```

### Step 3: Restart Activator

If no specific issue identified:

```bash
# Restart activator deployment
kubectl rollout restart deployment -n knative-serving activator

# Watch rollout progress
kubectl rollout status deployment -n knative-serving activator

# Verify new pod is running
kubectl get pods -n knative-serving -l app=activator
```

### Step 4: Check Networking

Verify activator can communicate with other components:

```bash
# Test activator service endpoint from another pod
kubectl run test-pod --rm -it --image=curlimages/curl --restart=Never -- \
  curl -v activator-service.knative-serving.svc.cluster.local:9090/health

# Check if activator can reach autoscaler
kubectl logs -n knative-serving -l app=activator --tail=50 | grep autoscaler
```

### Step 5: Verify Webhook Configuration

Activator needs webhook to function:

```bash
# Check webhook is running
kubectl get pods -n knative-serving -l app=webhook

# Check webhook configuration
kubectl get mutatingwebhookconfiguration | grep knative

# Restart webhook if needed
kubectl rollout restart deployment -n knative-serving webhook
```

## Verification

### 1. Check Activator is Running

```bash
kubectl get pods -n knative-serving -l app=activator
# Should show Running status with 1/1 READY
```

### 2. Check Activator Logs

```bash
kubectl logs -n knative-serving -l app=activator --tail=50
# Should show healthy operation, no errors
```

### 3. Test Scale-from-Zero

```bash
# Find or create a service at zero
kubectl get ksvc -A

# If needed, create test service
cat <<EOF | kubectl apply -f -
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: activator-test
  namespace: default
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/min-scale: "0"
    spec:
      containers:
      - image: gcr.io/knative-samples/helloworld-go
        env:
        - name: TARGET
          value: "Activator Test"
EOF

# Wait for service to scale to zero
sleep 60

# Verify at zero
kubectl get pods -n default -l serving.knative.dev/service=activator-test

# Test cold start via activator
URL=$(kubectl get ksvc activator-test -n default -o jsonpath='{.status.url}')
curl -v $URL

# Verify pod scaled up
kubectl get pods -n default -l serving.knative.dev/service=activator-test

# Cleanup
kubectl delete ksvc activator-test -n default
```

### 4. Check Activator Metrics

```bash
# Check activator endpoint
kubectl port-forward -n knative-serving svc/activator-service 9090:9090 &
curl http://localhost:9090/metrics | grep activator
kill %1
```

### 5. Verify Service Endpoints

```bash
kubectl get endpoints -n knative-serving activator-service
# Should show pod IP addresses
```

## Prevention

### 1. Resource Management

Ensure adequate resources:

```yaml
# In KnativeServing CR or deployment
resources:
  requests:
    cpu: 200m
    memory: 256Mi
  limits:
    cpu: 1000m
    memory: 512Mi
```

### 2. Horizontal Scaling

For high traffic environments:

```bash
# Scale activator to multiple replicas
kubectl scale deployment -n knative-serving activator --replicas=2
```

### 3. Monitoring Setup

Key metrics to monitor:
- Activator pod availability
- Request latency through activator
- Activator CPU/memory usage
- Cold start success rate
- Scale-from-zero failures

### 4. Pod Disruption Budget

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: activator-pdb
  namespace: knative-serving
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app: activator
```

### 5. Health Checks

Ensure proper liveness and readiness probes:

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8012
  initialDelaySeconds: 30
  periodSeconds: 10
readinessProbe:
  httpGet:
    path: /ready
    port: 8012
  initialDelaySeconds: 10
  periodSeconds: 3
```

## Performance Tips

1. **Multiple Replicas**: Run 2+ replicas in production for HA
2. **Resource Allocation**: Provide adequate CPU/memory based on traffic
3. **Network Policies**: Ensure activator can communicate with services
4. **Timeout Configuration**: Tune activator timeout settings for your workloads
5. **Queue Proxy**: Configure queue proxy settings appropriately

## Related Alerts

- `KnativeServingDown`
- `KnativeScalingIssues`
- `KnativeHighLatency`
- `KnativeAutoscalerDown`

## Escalation

If activator cannot be restored within 10 minutes:

1. ✅ Verify all resolution steps completed
2. 🔍 Check cluster node health
3. 📊 Review recent cluster changes
4. 🔄 Consider scaling activator horizontally
5. 📞 Escalate to platform team
6. 🆘 Page on-call engineer if critical services affected

## Additional Resources

- [Knative Activator Documentation](https://knative.dev/docs/serving/architecture/#activator)
- [Knative Scaling Documentation](https://knative.dev/docs/serving/autoscaling/)
- [Knative Troubleshooting](https://knative.dev/docs/serving/troubleshooting/)

## Quick Commands Reference

```bash
# Check activator status
kubectl get pods -n knative-serving -l app=activator

# View activator logs
kubectl logs -n knative-serving -l app=activator --tail=100

# Restart activator
kubectl rollout restart deployment -n knative-serving activator

# Scale activator
kubectl scale deployment -n knative-serving activator --replicas=2

# Check activator service
kubectl get svc -n knative-serving activator-service

# Test activator health
kubectl port-forward -n knative-serving svc/activator-service 9090:9090
curl http://localhost:9090/health

# Check resource usage
kubectl top pods -n knative-serving -l app=activator
```

---

**Last Updated:** 2025-10-15  
**Version:** 1.0  
**Owner:** Platform Team

