# ⚠️ Runbook: Knative Scaling Issues

## Alert Information

**Alert Name:** `KnativeScalingIssues`  
**Severity:** Warning  
**Component:** knative-serving / autoscaler  
**Impact:** Services not scaling properly

## Symptom

Knative services are not scaling up or down properly. Services may be:
- Stuck at wrong replica count
- Not scaling up under load
- Not scaling down when idle
- Not scaling to zero when expected
- Scaling too aggressively or too slowly

## Impact

- **User Impact:** MEDIUM - Possible performance degradation or resource waste
- **Business Impact:** MEDIUM - Poor resource utilization, potential SLA breaches
- **Data Impact:** LOW - No data loss

## Diagnosis

### 1. Check Service Status

```bash
# List all Knative services
kubectl get ksvc -A

# Check specific service
kubectl get ksvc <service-name> -n <namespace>
kubectl describe ksvc <service-name> -n <namespace>
```

### 2. Check Current Pod Count

```bash
# Check pods for a service
kubectl get pods -n <namespace> -l serving.knative.dev/service=<service-name>

# Watch pods in real-time
kubectl get pods -n <namespace> -l serving.knative.dev/service=<service-name> -w
```

### 3. Check PodAutoscaler (PA) Status

```bash
# List PodAutoscalers
kubectl get pa -A

# Check specific PA
kubectl get pa <service-name> -n <namespace>
kubectl describe pa <service-name> -n <namespace>
```

### 4. Check Autoscaler Logs

```bash
# View autoscaler decisions
kubectl logs -n knative-serving -l app=autoscaler --tail=100

# Filter for specific service
kubectl logs -n knative-serving -l app=autoscaler --tail=200 | grep <service-name>
```

### 5. Check Service Configuration

```bash
# Check autoscaling annotations
kubectl get ksvc <service-name> -n <namespace> -o yaml | grep -A 10 annotations

# Key annotations to check:
# - autoscaling.knative.dev/min-scale
# - autoscaling.knative.dev/max-scale
# - autoscaling.knative.dev/target
# - autoscaling.knative.dev/metric
# - autoscaling.knative.dev/class
```

### 6. Check Autoscaler Configuration

```bash
# Check global autoscaler config
kubectl get configmap -n knative-serving config-autoscaler -o yaml
```

### 7. Check Metrics

```bash
# Check if metrics are being collected
kubectl logs -n knative-serving -l app=autoscaler --tail=100 | grep -i metric

# Check queue proxy metrics (if using Prometheus)
kubectl port-forward -n <namespace> <pod-name> 9090:9090 &
curl http://localhost:9090/metrics | grep queue_proxy
kill %1
```

### 8. Check Activator Status

```bash
# For scale-from-zero issues
kubectl get pods -n knative-serving -l app=activator
kubectl logs -n knative-serving -l app=activator --tail=100
```

## Resolution Steps

### Step 1: Identify Scaling Issue Type

Determine what type of scaling issue:

```bash
# Check current vs desired replicas
kubectl get pa <service-name> -n <namespace> -o yaml | grep -E "desiredScale|actualScale"

# Check service status
kubectl get ksvc <service-name> -n <namespace> -o jsonpath='{.status.conditions[?(@.type=="Ready")]}'
```

### Step 2: Common Issues and Fixes

#### Issue: Service Not Scaling Up Under Load
**Cause:** Target concurrency/RPS too high or metrics not collected  
**Fix:**
```bash
# Check current traffic
kubectl logs -n <namespace> -l serving.knative.dev/service=<service-name> --tail=100

# Check PA status for metrics
kubectl describe pa <service-name> -n <namespace>

# Lower the target concurrency (will scale up sooner)
kubectl patch ksvc <service-name> -n <namespace> --type merge -p '{
  "spec": {
    "template": {
      "metadata": {
        "annotations": {
          "autoscaling.knative.dev/target": "50"
        }
      }
    }
  }
}'

# Or set target RPS
kubectl patch ksvc <service-name> -n <namespace> --type merge -p '{
  "spec": {
    "template": {
      "metadata": {
        "annotations": {
          "autoscaling.knative.dev/metric": "rps",
          "autoscaling.knative.dev/target": "100"
        }
      }
    }
  }
}'

# Check autoscaler logs
kubectl logs -n knative-serving -l app=autoscaler --tail=100 | grep <service-name>
```

#### Issue: Service Not Scaling Down
**Cause:** Traffic still present, scale-down rate limited, or min-scale set  
**Fix:**
```bash
# Check min-scale annotation
kubectl get ksvc <service-name> -n <namespace> -o jsonpath='{.spec.template.metadata.annotations.autoscaling\.knative\.dev/min-scale}'

# If min-scale is set, remove or adjust it
kubectl patch ksvc <service-name> -n <namespace> --type json -p '[
  {"op": "remove", "path": "/spec/template/metadata/annotations/autoscaling.knative.dev~1min-scale"}
]'

# Or set to 0 for scale-to-zero
kubectl annotate ksvc <service-name> -n <namespace> autoscaling.knative.dev/min-scale="0" --overwrite

# Check if traffic has stopped
kubectl logs -n <namespace> -l serving.knative.dev/service=<service-name> --tail=50

# Check stable window (default 60s)
kubectl get configmap -n knative-serving config-autoscaler -o yaml | grep stable-window

# Force immediate reconciliation
kubectl annotate pa <service-name> -n <namespace> reconcile=$(date +%s) --overwrite
```

#### Issue: Service Not Scaling to Zero
**Cause:** Min-scale annotation, traffic present, or scale-to-zero disabled  
**Fix:**
```bash
# Check if scale-to-zero is enabled globally
kubectl get configmap -n knative-serving config-autoscaler -o yaml | grep enable-scale-to-zero

# Check min-scale annotation
kubectl get ksvc <service-name> -n <namespace> -o yaml | grep min-scale

# Remove min-scale if present
kubectl annotate ksvc <service-name> -n <namespace> autoscaling.knative.dev/min-scale- --overwrite

# Ensure scale-to-zero is allowed
kubectl annotate ksvc <service-name> -n <namespace> autoscaling.knative.dev/min-scale="0" --overwrite

# Check grace period
kubectl get configmap -n knative-serving config-autoscaler -o yaml | grep scale-to-zero-grace-period

# Wait for grace period (default 30s) plus stable window (default 60s)
# Then check if scaled to zero
sleep 90
kubectl get pods -n <namespace> -l serving.knative.dev/service=<service-name>
```

#### Issue: Service Scaling Too Aggressively
**Cause:** Target concurrency/RPS too low, panic mode triggered  
**Fix:**
```bash
# Increase target concurrency
kubectl annotate ksvc <service-name> -n <namespace> autoscaling.knative.dev/target="200" --overwrite

# Increase target RPS
kubectl annotate ksvc <service-name> -n <namespace> \
  autoscaling.knative.dev/metric="rps" \
  autoscaling.knative.dev/target="500" \
  --overwrite

# Set max scale limit
kubectl annotate ksvc <service-name> -n <namespace> autoscaling.knative.dev/max-scale="10" --overwrite

# Adjust panic window (% of stable window)
kubectl get configmap -n knative-serving config-autoscaler -o yaml | grep panic-window-percentage
```

#### Issue: Service Scaling Too Slowly
**Cause:** Target concurrency/RPS too high, scale-up rate limited  
**Fix:**
```bash
# Decrease target concurrency for faster scale-up
kubectl annotate ksvc <service-name> -n <namespace> autoscaling.knative.dev/target="50" --overwrite

# Check scale-up rate limit
kubectl get configmap -n knative-serving config-autoscaler -o yaml | grep max-scale-up-rate

# Set initial scale to handle immediate traffic
kubectl annotate ksvc <service-name> -n <namespace> autoscaling.knative.dev/initial-scale="2" --overwrite

# Set min scale to keep pods ready
kubectl annotate ksvc <service-name> -n <namespace> autoscaling.knative.dev/min-scale="1" --overwrite
```

#### Issue: Metrics Not Being Collected
**Cause:** Queue proxy not reporting metrics, Prometheus issues  
**Fix:**
```bash
# Check queue proxy logs
kubectl logs -n <namespace> <pod-name> -c queue-proxy --tail=100

# Check metrics endpoint
kubectl port-forward -n <namespace> <pod-name> 9090:9090 &
curl http://localhost:9090/metrics | grep queue_proxy
kill %1

# Check autoscaler can reach metrics
kubectl logs -n knative-serving -l app=autoscaler --tail=100 | grep -i "metric"

# Check metrics backend configuration
kubectl get configmap -n knative-serving config-observability -o yaml

# Restart autoscaler
kubectl rollout restart deployment -n knative-serving autoscaler
```

#### Issue: PodAutoscaler in Error State
**Cause:** Invalid configuration or metrics unavailable  
**Fix:**
```bash
# Check PA status
kubectl describe pa <service-name> -n <namespace>

# Look for error conditions
kubectl get pa <service-name> -n <namespace> -o yaml | grep -A 10 conditions

# Force PA reconciliation
kubectl annotate pa <service-name> -n <namespace> reconcile=$(date +%s) --overwrite

# If PA is stuck, delete it (will be recreated)
kubectl delete pa <service-name> -n <namespace>

# Wait for recreation
sleep 10
kubectl get pa <service-name> -n <namespace>
```

### Step 3: Restart Autoscaler

If autoscaler is not making correct decisions:

```bash
# Restart autoscaler
kubectl rollout restart deployment -n knative-serving autoscaler

# Watch autoscaler logs
kubectl logs -n knative-serving -l app=autoscaler -f
```

### Step 4: Force Service Reconciliation

```bash
# Force service to reconcile
kubectl annotate ksvc <service-name> -n <namespace> reconcile=$(date +%s) --overwrite

# Force PA to reconcile
kubectl annotate pa <service-name> -n <namespace> reconcile=$(date +%s) --overwrite
```

### Step 5: Manual Scaling (Temporary)

If autoscaling is broken, manually scale as workaround:

```bash
# Disable autoscaling and set fixed replicas
kubectl patch ksvc <service-name> -n <namespace> --type merge -p '{
  "spec": {
    "template": {
      "metadata": {
        "annotations": {
          "autoscaling.knative.dev/class": "kpa.autoscaling.knative.dev",
          "autoscaling.knative.dev/min-scale": "3",
          "autoscaling.knative.dev/max-scale": "3"
        }
      }
    }
  }
}'

# This keeps service at 3 replicas
# Remove when autoscaling is fixed
```

## Verification

### 1. Check Service is Scaling

```bash
# Generate load
URL=$(kubectl get ksvc <service-name> -n <namespace> -o jsonpath='{.status.url}')

# Install hey if needed: https://github.com/rakyll/hey
hey -z 30s -c 50 $URL

# Watch pods scale up
kubectl get pods -n <namespace> -l serving.knative.dev/service=<service-name> -w

# Should see pods increase under load
```

### 2. Check Scale Down

```bash
# Stop traffic
# Wait for stable window + grace period (default 60s + 30s)
sleep 90

# Check pods scaled down
kubectl get pods -n <namespace> -l serving.knative.dev/service=<service-name>

# Should see fewer pods or zero (if scale-to-zero enabled)
```

### 3. Check PodAutoscaler Metrics

```bash
# Check PA has metrics
kubectl describe pa <service-name> -n <namespace> | grep -A 10 Status

# Should show current metrics
```

### 4. Check Autoscaler Logs

```bash
# Should show scaling decisions
kubectl logs -n knative-serving -l app=autoscaler --tail=100 | grep <service-name>
```

## Prevention

### 1. Proper Annotation Configuration

```yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: example
spec:
  template:
    metadata:
      annotations:
        # Autoscaling class
        autoscaling.knative.dev/class: "kpa.autoscaling.knative.dev"
        
        # Metric type: concurrency or rps
        autoscaling.knative.dev/metric: "concurrency"
        
        # Target value
        autoscaling.knative.dev/target: "100"
        
        # Min replicas (0 for scale-to-zero)
        autoscaling.knative.dev/min-scale: "1"
        
        # Max replicas
        autoscaling.knative.dev/max-scale: "10"
        
        # Initial scale
        autoscaling.knative.dev/initial-scale: "1"
        
        # Scale-to-zero pod retention period
        autoscaling.knative.dev/scale-to-zero-pod-retention-period: "0s"
```

### 2. Monitoring Setup

Key metrics to monitor:
- Current replica count vs target
- Request concurrency per pod
- Requests per second
- Autoscaler decisions (scale up/down events)
- Scale-to-zero events
- Cold start latency

### 3. Load Testing

Regularly test autoscaling behavior:

```bash
# Test scale-up
hey -z 60s -c 100 $SERVICE_URL

# Test scale-down
# Wait after traffic stops

# Test scale-to-zero
# Ensure no traffic for grace period
```

### 4. Tune Autoscaler Config

Global autoscaler configuration:

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

## Performance Tips

1. **Right-size Targets**: Set appropriate concurrency/RPS targets based on testing
2. **Min Scale**: Use min-scale > 0 for frequently accessed services to avoid cold starts
3. **Max Scale**: Set max-scale to prevent runaway scaling
4. **Initial Scale**: Set initial-scale for predictable startup
5. **Resource Limits**: Set proper CPU/memory limits for predictable scaling
6. **Readiness Probes**: Fast readiness probes enable quick scaling

## Related Alerts

- `KnativeAutoscalerDown`
- `KnativeHighLatency`
- `KnativeServingDown`
- `KnativeActivatorDown`

## Escalation

If scaling issues cannot be resolved within 30 minutes:

1. ✅ Verify all resolution steps completed
2. 🔍 Check autoscaler and metrics collection health
3. 📊 Review service configuration and annotations
4. 🔄 Consider temporary manual scaling
5. 📞 Escalate to platform team
6. 🆘 Page on-call engineer if SLA at risk

## Additional Resources

- [Knative Autoscaling](https://knative.dev/docs/serving/autoscaling/)
- [Autoscaling Concepts](https://knative.dev/docs/serving/autoscaling/autoscaling-concepts/)
- [Autoscaling Configuration](https://knative.dev/docs/serving/autoscaling/autoscale-go/)
- [Knative Metrics](https://knative.dev/docs/serving/observability/metrics/)

## Quick Commands Reference

```bash
# Check service status
kubectl get ksvc <service-name> -n <namespace>

# Check pods
kubectl get pods -n <namespace> -l serving.knative.dev/service=<service-name>

# Check PodAutoscaler
kubectl get pa <service-name> -n <namespace>
kubectl describe pa <service-name> -n <namespace>

# Check autoscaler logs
kubectl logs -n knative-serving -l app=autoscaler --tail=100 | grep <service-name>

# Set target concurrency
kubectl annotate ksvc <service-name> -n <namespace> autoscaling.knative.dev/target="100" --overwrite

# Set min/max scale
kubectl annotate ksvc <service-name> -n <namespace> \
  autoscaling.knative.dev/min-scale="1" \
  autoscaling.knative.dev/max-scale="10" \
  --overwrite

# Enable scale-to-zero
kubectl annotate ksvc <service-name> -n <namespace> autoscaling.knative.dev/min-scale="0" --overwrite

# Force reconciliation
kubectl annotate ksvc <service-name> -n <namespace> reconcile=$(date +%s) --overwrite

# Generate test load
hey -z 30s -c 50 $(kubectl get ksvc <service-name> -n <namespace> -o jsonpath='{.status.url}')
```

---

**Last Updated:** 2025-10-15  
**Version:** 1.0  
**Owner:** Platform Team

