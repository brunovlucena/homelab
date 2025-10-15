# 🚨 Runbook: Knative Autoscaler Down

## Alert Information

**Alert Name:** `KnativeAutoscalerDown`  
**Severity:** Critical  
**Component:** knative-serving / autoscaler  
**Impact:** Services not scaling automatically

## Symptom

The Knative Autoscaler component is unavailable or not responding. Services cannot scale up or down based on traffic, including scale-to-zero functionality.

## Impact

- **User Impact:** HIGH - Services may not handle traffic spikes, poor performance
- **Business Impact:** HIGH - Fixed replica counts, inefficient resource usage
- **Data Impact:** LOW - No data loss, but service degradation possible

## Diagnosis

### 1. Check Autoscaler Pod Status

```bash
kubectl get pods -n knative-serving -l app=autoscaler
kubectl get pods -n knative-serving -l app=autoscaler -o wide
```

**Expected Output:**
```
NAME                          READY   STATUS    RESTARTS   AGE
autoscaler-xxxxxxxxxx-xxxxx   1/1     Running   0          24h
```

### 2. Check Autoscaler Deployment

```bash
kubectl describe deployment -n knative-serving autoscaler
kubectl get deployment -n knative-serving autoscaler -o yaml
```

### 3. Check Autoscaler Logs

```bash
# Recent logs
kubectl logs -n knative-serving -l app=autoscaler --tail=100

# Previous container logs (if crashed)
kubectl logs -n knative-serving -l app=autoscaler --tail=100 --previous
```

### 4. Check Autoscaler Configuration

```bash
kubectl get configmap -n knative-serving config-autoscaler -o yaml
```

### 5. Check PodAutoscalers (PA)

```bash
# List all PodAutoscalers
kubectl get pa -A

# Check specific PA status
kubectl describe pa <pa-name> -n <namespace>
```

### 6. Check Resource Usage

```bash
kubectl top pods -n knative-serving -l app=autoscaler
```

### 7. Check Recent Events

```bash
kubectl get events -n knative-serving --field-selector involvedObject.name=autoscaler --sort-by='.lastTimestamp'
```

### 8. Check Metrics Backend

```bash
# Autoscaler needs to scrape metrics
kubectl get configmap -n knative-serving config-observability -o yaml | grep metrics-backend

# Check if Prometheus is accessible
kubectl get svc -n prometheus prometheus-operated
```

## Resolution Steps

### Step 1: Identify Root Cause

Check pod status for common issues:

```bash
# Get pod details
kubectl describe pod -n knative-serving -l app=autoscaler

# Common indicators:
# - CrashLoopBackOff: Configuration or application error
# - OOMKilled: Out of memory
# - Error logs: Metrics backend issues
```

### Step 2: Common Issues and Fixes

#### Issue: Pod CrashLoopBackOff
**Cause:** Configuration error or metrics backend unavailable  
**Fix:**
```bash
# Check logs for specific errors
kubectl logs -n knative-serving -l app=autoscaler --tail=200

# Check autoscaler configuration
kubectl get configmap -n knative-serving config-autoscaler -o yaml

# Check if metrics backend is configured correctly
kubectl get configmap -n knative-serving config-observability -o yaml

# Verify Prometheus is accessible
kubectl get pods -n prometheus

# Restart autoscaler
kubectl rollout restart deployment -n knative-serving autoscaler
```

#### Issue: Configuration Error
**Cause:** Invalid autoscaler configuration  
**Fix:**
```bash
# Get current config
kubectl get configmap -n knative-serving config-autoscaler -o yaml

# Common configuration issues:
# - Invalid metric types
# - Wrong backend configuration
# - Invalid scale boundaries

# Reset to default if needed
kubectl get configmap -n knative-serving config-autoscaler -o yaml > /tmp/autoscaler-config-backup.yaml

# Edit configuration
kubectl edit configmap -n knative-serving config-autoscaler

# Key settings to verify:
# - enable-scale-to-zero: "true"
# - scale-to-zero-grace-period: "30s"
# - container-concurrency-target-default: "100"
# - requests-per-second-target-default: "200"

# Restart to apply changes
kubectl rollout restart deployment -n knative-serving autoscaler
```

#### Issue: Metrics Backend Unavailable
**Cause:** Cannot connect to Prometheus or metrics backend  
**Fix:**
```bash
# Check Prometheus is running
kubectl get pods -n prometheus

# Check if autoscaler can reach Prometheus
kubectl logs -n knative-serving -l app=autoscaler --tail=100 | grep -i prometheus

# Verify metrics backend configuration
kubectl get configmap -n knative-serving config-observability -o yaml

# Test Prometheus connectivity from autoscaler namespace
kubectl run test-metrics --rm -it --image=curlimages/curl --restart=Never -n knative-serving -- \
  curl -v http://prometheus-operated.prometheus.svc.cluster.local:9090/-/healthy
```

#### Issue: Pod OOMKilled
**Cause:** Insufficient memory allocation  
**Fix:**
```bash
# Check current resource limits
kubectl get deployment -n knative-serving autoscaler -o yaml | grep -A 10 resources

# Increase memory limits
kubectl patch deployment -n knative-serving autoscaler -p '{"spec":{"template":{"spec":{"containers":[{"name":"autoscaler","resources":{"limits":{"memory":"512Mi"},"requests":{"memory":"256Mi"}}}]}}}}'

# Wait for rollout
kubectl rollout status deployment -n knative-serving autoscaler
```

#### Issue: PodAutoscalers Not Working
**Cause:** Autoscaler not processing PA resources  
**Fix:**
```bash
# Check PA resources
kubectl get pa -A

# Check if PA has status
kubectl describe pa <pa-name> -n <namespace>

# Check autoscaler logs for PA processing
kubectl logs -n knative-serving -l app=autoscaler --tail=100 | grep -i "podautoscaler"

# Force PA reconciliation
kubectl annotate pa <pa-name> -n <namespace> reconcile=$(date +%s) --overwrite
```

### Step 3: Restart Autoscaler

If no specific issue identified:

```bash
# Restart autoscaler deployment
kubectl rollout restart deployment -n knative-serving autoscaler

# Watch rollout progress
kubectl rollout status deployment -n knative-serving autoscaler

# Verify new pod is running
kubectl get pods -n knative-serving -l app=autoscaler
```

### Step 4: Verify Metrics Collection

```bash
# Check autoscaler metrics endpoint
kubectl port-forward -n knative-serving svc/autoscaler 9090:9090 &
curl http://localhost:9090/metrics | grep autoscaler
kill %1

# Check if autoscaler is receiving service metrics
kubectl logs -n knative-serving -l app=autoscaler --tail=50 | grep -i metric
```

### Step 5: Check Webhook Integration

Autoscaler works with webhook for defaults:

```bash
# Check webhook is running
kubectl get pods -n knative-serving -l app=webhook

# Restart webhook if needed
kubectl rollout restart deployment -n knative-serving webhook
```

## Verification

### 1. Check Autoscaler is Running

```bash
kubectl get pods -n knative-serving -l app=autoscaler
# Should show Running status with 1/1 READY
```

### 2. Check Autoscaler Logs

```bash
kubectl logs -n knative-serving -l app=autoscaler --tail=50
# Should show autoscaling decisions, no errors
```

### 3. Test Autoscaling

```bash
# Create test service with scaling
cat <<EOF | kubectl apply -f -
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: autoscaler-test
  namespace: default
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/min-scale: "1"
        autoscaling.knative.dev/max-scale: "5"
        autoscaling.knative.dev/target: "10"
    spec:
      containers:
      - image: gcr.io/knative-samples/autoscale-go
        resources:
          requests:
            cpu: 100m
EOF

# Wait for service to be ready
kubectl wait --for=condition=Ready ksvc/autoscaler-test -n default --timeout=60s

# Get service URL
URL=$(kubectl get ksvc autoscaler-test -n default -o jsonpath='{.status.url}')

# Generate load to test scaling
hey -z 30s -c 50 $URL

# Watch pods scale up
kubectl get pods -n default -l serving.knative.dev/service=autoscaler-test -w

# Wait for scale down
sleep 60

# Verify scale down occurred
kubectl get pods -n default -l serving.knative.dev/service=autoscaler-test

# Cleanup
kubectl delete ksvc autoscaler-test -n default
```

### 4. Check PodAutoscaler Status

```bash
# List PodAutoscalers
kubectl get pa -A

# Check PA is reporting metrics
kubectl describe pa <pa-name> -n <namespace> | grep -A 10 Status
```

### 5. Verify Scale-to-Zero

```bash
# Find service with scale-to-zero enabled
kubectl get ksvc -A

# Check if services scale to zero after idle period
# (default 30s grace period)
```

## Prevention

### 1. Resource Management

Ensure adequate resources:

```yaml
# In KnativeServing CR or deployment
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 1000m
    memory: 256Mi
```

### 2. Valid Configuration

Maintain valid autoscaler configuration:

```yaml
# config-autoscaler ConfigMap
data:
  enable-scale-to-zero: "true"
  scale-to-zero-grace-period: "30s"
  stable-window: "60s"
  panic-window-percentage: "10.0"
  container-concurrency-target-default: "100"
  container-concurrency-target-percentage: "70"
  requests-per-second-target-default: "200"
  max-scale-up-rate: "1000.0"
  max-scale-down-rate: "2.0"
```

### 3. Monitoring Setup

Key metrics to monitor:
- Autoscaler pod availability
- Autoscaling decisions (scale up/down events)
- Metrics collection success rate
- PA resource status
- Scale-to-zero events

### 4. Metrics Backend Health

Ensure Prometheus is healthy and accessible:

```bash
# Monitor Prometheus
kubectl get pods -n prometheus

# Monitor metrics scraping
kubectl logs -n knative-serving -l app=autoscaler | grep -i error
```

### 5. Pod Disruption Budget

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: autoscaler-pdb
  namespace: knative-serving
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app: autoscaler
```

## Performance Tips

1. **Tune Metrics Windows**: Adjust stable-window and panic-window for your traffic patterns
2. **Configure Targets**: Set appropriate concurrency and RPS targets
3. **Scale Rate Limits**: Configure max-scale-up/down-rate to prevent thrashing
4. **Resource Allocation**: Provide adequate resources for high service count
5. **Metrics Backend**: Ensure Prometheus has sufficient resources

## Related Alerts

- `KnativeServingDown`
- `KnativeScalingIssues`
- `KnativeActivatorDown`
- `PrometheusDown`

## Escalation

If autoscaler cannot be restored within 15 minutes:

1. ✅ Verify all resolution steps completed
2. 🔍 Check Prometheus health and metrics collection
3. 📊 Review autoscaler configuration for errors
4. 🔄 Consider temporary manual scaling of affected services
5. 📞 Escalate to platform team
6. 🆘 Page on-call engineer if critical services affected

## Additional Resources

- [Knative Autoscaling Documentation](https://knative.dev/docs/serving/autoscaling/)
- [Autoscaler Configuration](https://knative.dev/docs/serving/autoscaling/autoscaling-concepts/)
- [Knative Troubleshooting](https://knative.dev/docs/serving/troubleshooting/)

## Quick Commands Reference

```bash
# Check autoscaler status
kubectl get pods -n knative-serving -l app=autoscaler

# View autoscaler logs
kubectl logs -n knative-serving -l app=autoscaler --tail=100

# Restart autoscaler
kubectl rollout restart deployment -n knative-serving autoscaler

# Check autoscaler config
kubectl get configmap -n knative-serving config-autoscaler -o yaml

# Check PodAutoscalers
kubectl get pa -A

# Check autoscaler metrics
kubectl port-forward -n knative-serving svc/autoscaler 9090:9090
curl http://localhost:9090/metrics

# Check resource usage
kubectl top pods -n knative-serving -l app=autoscaler
```

---

**Last Updated:** 2025-10-15  
**Version:** 1.0  
**Owner:** Platform Team

