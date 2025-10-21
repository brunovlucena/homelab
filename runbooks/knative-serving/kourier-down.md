# 🚨 Runbook: Knative Kourier Ingress Down

## Alert Information

**Alert Name:** `KnativeKourierDown`  
**Severity:** Critical  
**Component:** knative-serving / kourier (3scale-kourier-gateway)  
**Impact:** External traffic cannot reach Knative services

## Symptom

The Kourier ingress gateway is unavailable or not responding. External traffic cannot reach any Knative services. All service URLs return connection errors or timeouts.

## Impact

- **User Impact:** CRITICAL - All Knative services unreachable from outside cluster
- **Business Impact:** CRITICAL - Complete external access outage
- **Data Impact:** LOW - No data loss, services run internally

## Diagnosis

### 1. Check Kourier Pod Status

```bash
kubectl get pods -n knative-serving -l app=3scale-kourier-gateway
kubectl get pods -n knative-serving -l app=3scale-kourier-gateway -o wide
```

**Expected Output:**
```
NAME                                      READY   STATUS    RESTARTS   AGE
3scale-kourier-gateway-xxxxxxxxxx-xxxxx   1/1     Running   0          24h
```

### 2. Check Kourier Deployment

```bash
kubectl describe deployment -n knative-serving 3scale-kourier-gateway
kubectl get deployment -n knative-serving 3scale-kourier-gateway -o yaml
```

### 3. Check Kourier Service

```bash
# Internal service
kubectl get svc -n knative-serving kourier
kubectl describe svc -n knative-serving kourier

# External service (LoadBalancer or NodePort)
kubectl get svc -n knative-serving kourier-external
kubectl describe svc -n knative-serving kourier-external
```

### 4. Check Kourier Logs

```bash
# Recent logs
kubectl logs -n knative-serving -l app=3scale-kourier-gateway --tail=100

# Previous container logs (if crashed)
kubectl logs -n knative-serving -l app=3scale-kourier-gateway --tail=100 --previous
```

### 5. Check Envoy Configuration

```bash
# Kourier uses Envoy under the hood
kubectl logs -n knative-serving -l app=3scale-kourier-gateway --tail=100 | grep -i envoy
kubectl logs -n knative-serving -l app=3scale-kourier-gateway --tail=100 | grep -i error
```

### 6. Check Resource Usage

```bash
kubectl top pods -n knative-serving -l app=3scale-kourier-gateway
```

### 7. Check Network Configuration

```bash
# Check network config
kubectl get configmap -n knative-serving config-network -o yaml

# Check Kourier ingress class
kubectl get configmap -n knative-serving config-network -o yaml | grep ingress-class
```

### 8. Test Connectivity

```bash
# Test from within cluster
kubectl run test-kourier --rm -it --image=curlimages/curl --restart=Never -- \
  curl -v http://kourier.knative-serving.svc.cluster.local

# Test external endpoint
EXTERNAL_IP=$(kubectl get svc -n knative-serving kourier-external -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
curl -v http://$EXTERNAL_IP
```

### 9. Check Recent Events

```bash
kubectl get events -n knative-serving --field-selector involvedObject.name=3scale-kourier-gateway --sort-by='.lastTimestamp'
```

## Resolution Steps

### Step 1: Identify Root Cause

Check pod status for common issues:

```bash
# Get pod details
kubectl describe pod -n knative-serving -l app=3scale-kourier-gateway

# Common indicators:
# - CrashLoopBackOff: Configuration error
# - OOMKilled: Out of memory
# - Running but no traffic: Service/networking issue
```

### Step 2: Common Issues and Fixes

#### Issue: Pod CrashLoopBackOff
**Cause:** Configuration error or resource issue  
**Fix:**
```bash
# Check logs for errors
kubectl logs -n knative-serving -l app=3scale-kourier-gateway --tail=200

# Common errors:
# - Envoy config errors
# - Port binding issues
# - Certificate errors

# Check network configuration
kubectl get configmap -n knative-serving config-network -o yaml

# Restart Kourier
kubectl rollout restart deployment -n knative-serving 3scale-kourier-gateway
```

#### Issue: External Service Not Accessible
**Cause:** LoadBalancer or NodePort service issue  
**Fix:**
```bash
# Check external service
kubectl get svc -n knative-serving kourier-external
kubectl describe svc -n knative-serving kourier-external

# Check service type
kubectl get svc -n knative-serving kourier-external -o jsonpath='{.spec.type}'

# If LoadBalancer pending, check cloud provider
kubectl describe svc -n knative-serving kourier-external | grep -A 10 Events

# For NodePort, verify ports
kubectl get svc -n knative-serving kourier-external -o jsonpath='{.spec.ports[*].nodePort}'

# Test specific node
NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')
NODE_PORT=$(kubectl get svc -n knative-serving kourier-external -o jsonpath='{.spec.ports[0].nodePort}')
curl -v http://$NODE_IP:$NODE_PORT
```

#### Issue: Service Endpoints Empty
**Cause:** Pods not ready or selector mismatch  
**Fix:**
```bash
# Check service endpoints
kubectl get endpoints -n knative-serving kourier
kubectl get endpoints -n knative-serving kourier-external

# If empty, check pod labels
kubectl get pods -n knative-serving -l app=3scale-kourier-gateway --show-labels

# Check service selector
kubectl get svc -n knative-serving kourier -o yaml | grep -A 3 selector

# Verify pod is ready
kubectl get pods -n knative-serving -l app=3scale-kourier-gateway -o jsonpath='{.items[*].status.conditions[?(@.type=="Ready")].status}'

# Restart if needed
kubectl rollout restart deployment -n knative-serving 3scale-kourier-gateway
```

#### Issue: Pod OOMKilled
**Cause:** Insufficient memory for traffic volume  
**Fix:**
```bash
# Check current resource limits
kubectl get deployment -n knative-serving 3scale-kourier-gateway -o yaml | grep -A 10 resources

# Increase memory limits
kubectl patch deployment -n knative-serving 3scale-kourier-gateway -p '{"spec":{"template":{"spec":{"containers":[{"name":"kourier-gateway","resources":{"limits":{"memory":"1Gi"},"requests":{"memory":"512Mi"}}}]}}}}'

# Wait for rollout
kubectl rollout status deployment -n knative-serving 3scale-kourier-gateway
```

#### Issue: High Connection Count/Latency
**Cause:** Kourier overloaded with traffic  
**Fix:**
```bash
# Check current resource usage
kubectl top pods -n knative-serving -l app=3scale-kourier-gateway

# Scale Kourier horizontally
kubectl scale deployment -n knative-serving 3scale-kourier-gateway --replicas=2

# Monitor new pods
kubectl get pods -n knative-serving -l app=3scale-kourier-gateway -w

# Check logs for errors
kubectl logs -n knative-serving -l app=3scale-kourier-gateway --tail=100
```

#### Issue: Route Configuration Issues
**Cause:** Envoy routes not configured correctly  
**Fix:**
```bash
# Check Knative routes
kubectl get route -A

# Check specific route
kubectl describe route <route-name> -n <namespace>

# Check Kourier is receiving route updates
kubectl logs -n knative-serving -l app=3scale-kourier-control --tail=100

# Restart Kourier control plane
kubectl rollout restart deployment -n knative-serving 3scale-kourier-control

# Restart gateway
kubectl rollout restart deployment -n knative-serving 3scale-kourier-gateway
```

### Step 3: Restart Kourier

If no specific issue identified:

```bash
# Restart Kourier gateway
kubectl rollout restart deployment -n knative-serving 3scale-kourier-gateway

# Watch rollout progress
kubectl rollout status deployment -n knative-serving 3scale-kourier-gateway

# Verify new pod is running
kubectl get pods -n knative-serving -l app=3scale-kourier-gateway
```

### Step 4: Verify Network Configuration

```bash
# Check network config is set to Kourier
kubectl get configmap -n knative-serving config-network -o yaml | grep ingress-class
# Should show: ingress-class: kourier.ingress.networking.knative.dev

# If missing, set it
kubectl patch configmap -n knative-serving config-network -p '{"data":{"ingress-class":"kourier.ingress.networking.knative.dev"}}'
```

### Step 5: Check Kourier Control Plane

```bash
# Kourier has a control plane component
kubectl get pods -n knative-serving -l app=3scale-kourier-control
kubectl logs -n knative-serving -l app=3scale-kourier-control --tail=100

# Restart if needed
kubectl rollout restart deployment -n knative-serving 3scale-kourier-control
```

## Verification

### 1. Check Kourier is Running

```bash
kubectl get pods -n knative-serving -l app=3scale-kourier-gateway
# Should show Running status with 1/1 READY
```

### 2. Check Kourier Services

```bash
kubectl get svc -n knative-serving kourier
kubectl get svc -n knative-serving kourier-external
# Both should have endpoints
```

### 3. Check Service Endpoints

```bash
kubectl get endpoints -n knative-serving kourier
kubectl get endpoints -n knative-serving kourier-external
# Should show pod IPs
```

### 4. Test Internal Connectivity

```bash
# Test from within cluster
kubectl run test-curl --rm -it --image=curlimages/curl --restart=Never -- \
  curl -v http://kourier.knative-serving.svc.cluster.local
```

### 5. Test External Connectivity

```bash
# Get external IP or NodePort
kubectl get svc -n knative-serving kourier-external

# For LoadBalancer
EXTERNAL_IP=$(kubectl get svc -n knative-serving kourier-external -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
curl -v http://$EXTERNAL_IP

# For NodePort
NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')
NODE_PORT=$(kubectl get svc -n knative-serving kourier-external -o jsonpath='{.spec.ports[0].nodePort}')
curl -v http://$NODE_IP:$NODE_PORT
```

### 6. Test with Knative Service

```bash
# Create test service
cat <<EOF | kubectl apply -f -
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: kourier-test
  namespace: default
spec:
  template:
    spec:
      containers:
      - image: gcr.io/knative-samples/helloworld-go
        env:
        - name: TARGET
          value: "Kourier Test"
EOF

# Wait for service to be ready
kubectl wait --for=condition=Ready ksvc/kourier-test -n default --timeout=60s

# Get service URL
URL=$(kubectl get ksvc kourier-test -n default -o jsonpath='{.status.url}')
echo "Service URL: $URL"

# Test service (may need to resolve DNS or use --resolve)
curl $URL

# Or test with explicit host header
EXTERNAL_IP=$(kubectl get svc -n knative-serving kourier-external -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
curl -H "Host: kourier-test.default.example.com" http://$EXTERNAL_IP

# Cleanup
kubectl delete ksvc kourier-test -n default
```

### 7. Check Kourier Logs

```bash
kubectl logs -n knative-serving -l app=3scale-kourier-gateway --tail=50
# Should show access logs with successful requests
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
    cpu: 2000m
    memory: 1Gi
```

### 2. Horizontal Scaling

For high traffic environments:

```bash
# Scale Kourier to multiple replicas
kubectl scale deployment -n knative-serving 3scale-kourier-gateway --replicas=2
```

### 3. Monitoring Setup

Key metrics to monitor:
- Kourier pod availability
- Request throughput
- Request latency (p50, p95, p99)
- Error rate (4xx, 5xx)
- Connection count
- CPU/memory usage

### 4. LoadBalancer Health

Monitor external service:

```bash
# Check LoadBalancer status
kubectl get svc -n knative-serving kourier-external -w
```

### 5. Pod Disruption Budget

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: kourier-gateway-pdb
  namespace: knative-serving
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app: 3scale-kourier-gateway
```

## Performance Tips

1. **Multiple Replicas**: Run 2+ replicas for HA and load distribution
2. **Resource Allocation**: Provide adequate CPU/memory for expected traffic
3. **Envoy Tuning**: Configure Envoy settings for your workload
4. **Connection Limits**: Set appropriate connection and timeout limits
5. **DNS Configuration**: Ensure proper DNS resolution for service URLs

## Related Alerts

- `KnativeServingDown`
- `KnativeHighLatency`
- `LoadBalancerDown`

## Escalation

If Kourier cannot be restored within 10 minutes:

1. ✅ Verify all resolution steps completed
2. 🔍 Check LoadBalancer or ingress infrastructure
3. 📊 Review network policies and firewall rules
4. 🔄 Consider emergency traffic rerouting
5. 📞 Escalate to platform team
6. 🆘 Page on-call engineer for critical service access

## Additional Resources

- [Kourier Documentation](https://github.com/knative-extensions/net-kourier)
- [Knative Networking](https://knative.dev/docs/serving/networking/)
- [Envoy Proxy Documentation](https://www.envoyproxy.io/docs)

## Quick Commands Reference

```bash
# Check Kourier status
kubectl get pods -n knative-serving -l app=3scale-kourier-gateway

# View Kourier logs
kubectl logs -n knative-serving -l app=3scale-kourier-gateway --tail=100

# Restart Kourier
kubectl rollout restart deployment -n knative-serving 3scale-kourier-gateway

# Check Kourier services
kubectl get svc -n knative-serving kourier
kubectl get svc -n knative-serving kourier-external

# Scale Kourier
kubectl scale deployment -n knative-serving 3scale-kourier-gateway --replicas=2

# Test connectivity
EXTERNAL_IP=$(kubectl get svc -n knative-serving kourier-external -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
curl -v http://$EXTERNAL_IP

# Check resource usage
kubectl top pods -n knative-serving -l app=3scale-kourier-gateway

# Check network config
kubectl get configmap -n knative-serving config-network -o yaml | grep ingress-class
```

---

**Last Updated:** 2025-10-15  
**Version:** 1.0  
**Owner:** Platform Team

