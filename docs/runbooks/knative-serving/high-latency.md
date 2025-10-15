# ⚠️ Runbook: Knative High Latency

## Alert Information

**Alert Name:** `KnativeHighLatency`  
**Severity:** Warning  
**Component:** knative-serving  
**Impact:** Slow request processing through Knative

## Symptom

Requests through Knative Serving are experiencing high latency. Users report slow response times, timeouts, or degraded performance.

## Impact

- **User Impact:** MEDIUM - Slow response times, poor user experience
- **Business Impact:** MEDIUM - SLA violations possible, user dissatisfaction
- **Data Impact:** LOW - No data loss

## Diagnosis

### 1. Measure Current Latency

```bash
# Test service latency
URL=$(kubectl get ksvc <service-name> -n <namespace> -o jsonpath='{.status.url}')
time curl -s $URL

# Multiple requests to get average
for i in {1..10}; do
  time curl -s $URL
done

# Use hey for detailed metrics
hey -n 100 -c 10 $URL
```

### 2. Check Service Status

```bash
# Check service health
kubectl get ksvc <service-name> -n <namespace>
kubectl describe ksvc <service-name> -n <namespace>

# Check pods
kubectl get pods -n <namespace> -l serving.knative.dev/service=<service-name>
kubectl top pods -n <namespace> -l serving.knative.dev/service=<service-name>
```

### 3. Check Activator (Cold Start Latency)

```bash
# Check if service is at zero (will cause cold start)
kubectl get pods -n <namespace> -l serving.knative.dev/service=<service-name>

# Check activator logs
kubectl logs -n knative-serving -l app=activator --tail=100

# Check activator resource usage
kubectl top pods -n knative-serving -l app=activator
```

### 4. Check Kourier Gateway

```bash
# Check Kourier logs for latency
kubectl logs -n knative-serving -l app=3scale-kourier-gateway --tail=100

# Check Kourier resource usage
kubectl top pods -n knative-serving -l app=3scale-kourier-gateway

# Check Kourier connection count
kubectl logs -n knative-serving -l app=3scale-kourier-gateway --tail=100 | grep -i connection
```

### 5. Check Queue Proxy

```bash
# Queue proxy sits between ingress and application
kubectl logs -n <namespace> <pod-name> -c queue-proxy --tail=100

# Check queue proxy metrics
kubectl port-forward -n <namespace> <pod-name> 9090:9090 &
curl http://localhost:9090/metrics | grep queue_proxy
kill %1
```

### 6. Check Application Container

```bash
# Check application logs
kubectl logs -n <namespace> <pod-name> -c user-container --tail=100

# Check application resource usage
kubectl top pod <pod-name> -n <namespace> --containers
```

### 7. Check Network Latency

```bash
# Test latency from within cluster
kubectl run test-latency --rm -it --image=curlimages/curl --restart=Never -- sh -c "
  time curl -s http://<service-name>.<namespace>.svc.cluster.local
"

# Compare external vs internal latency
```

### 8. Check for Resource Constraints

```bash
# Check node resource usage
kubectl top nodes

# Check namespace resource usage
kubectl top pods -n <namespace>

# Check for resource limits
kubectl get pod <pod-name> -n <namespace> -o yaml | grep -A 10 resources
```

## Resolution Steps

### Step 1: Identify Latency Source

Determine where latency is coming from:

```bash
# Test end-to-end latency
time curl -w "@-" -o /dev/null -s $URL <<'EOF'
    time_namelookup:  %{time_namelookup}s\n
       time_connect:  %{time_connect}s\n
    time_appconnect:  %{time_appconnect}s\n
   time_pretransfer:  %{time_pretransfer}s\n
      time_redirect:  %{time_redirect}s\n
 time_starttransfer:  %{time_starttransfer}s\n
                    ----------\n
         time_total:  %{time_total}s\n
EOF

# High time_starttransfer = slow application
# High time_connect = network issues
# High time_total with low time_starttransfer = slow response generation
```

### Step 2: Common Issues and Fixes

#### Issue: Cold Start Latency
**Cause:** Service scaled to zero, activator must start pods  
**Fix:**
```bash
# Check if service is at zero
kubectl get pods -n <namespace> -l serving.knative.dev/service=<service-name>

# Prevent scale-to-zero by setting min-scale
kubectl annotate ksvc <service-name> -n <namespace> autoscaling.knative.dev/min-scale="1" --overwrite

# Or increase initial scale
kubectl annotate ksvc <service-name> -n <namespace> autoscaling.knative.dev/initial-scale="2" --overwrite

# Optimize container startup time
# - Use smaller images
# - Pre-warm dependencies
# - Optimize readiness probe

# Set faster readiness probe
kubectl patch ksvc <service-name> -n <namespace> --type merge -p '{
  "spec": {
    "template": {
      "spec": {
        "containers": [{
          "readinessProbe": {
            "initialDelaySeconds": 0,
            "periodSeconds": 1
          }
        }]
      }
    }
  }
}'
```

#### Issue: Activator Overloaded
**Cause:** Too much traffic through activator (scale-from-zero)  
**Fix:**
```bash
# Check activator resource usage
kubectl top pods -n knative-serving -l app=activator

# Scale activator horizontally
kubectl scale deployment -n knative-serving activator --replicas=2

# Increase activator resources
kubectl patch deployment -n knative-serving activator -p '{
  "spec": {
    "template": {
      "spec": {
        "containers": [{
          "name": "activator",
          "resources": {
            "limits": {
              "cpu": "1000m",
              "memory": "512Mi"
            },
            "requests": {
              "cpu": "300m",
              "memory": "256Mi"
            }
          }
        }]
      }
    }
  }
}'

# Set min-scale to avoid activator path
kubectl annotate ksvc <service-name> -n <namespace> autoscaling.knative.dev/min-scale="1" --overwrite
```

#### Issue: Kourier Gateway Overloaded
**Cause:** High traffic through Kourier ingress  
**Fix:**
```bash
# Check Kourier resource usage
kubectl top pods -n knative-serving -l app=3scale-kourier-gateway

# Scale Kourier horizontally
kubectl scale deployment -n knative-serving 3scale-kourier-gateway --replicas=2

# Increase Kourier resources
kubectl patch deployment -n knative-serving 3scale-kourier-gateway -p '{
  "spec": {
    "template": {
      "spec": {
        "containers": [{
          "name": "kourier-gateway",
          "resources": {
            "limits": {
              "cpu": "2000m",
              "memory": "1Gi"
            },
            "requests": {
              "cpu": "500m",
              "memory": "512Mi"
            }
          }
        }]
      }
    }
  }
}'

# Check Kourier logs for bottlenecks
kubectl logs -n knative-serving -l app=3scale-kourier-gateway --tail=200
```

#### Issue: Application Container Resource Starved
**Cause:** Insufficient CPU/memory for application  
**Fix:**
```bash
# Check current resource usage
kubectl top pod <pod-name> -n <namespace> --containers

# Check for CPU throttling
kubectl describe pod <pod-name> -n <namespace> | grep -i throttl

# Increase application resources
kubectl patch ksvc <service-name> -n <namespace> --type merge -p '{
  "spec": {
    "template": {
      "spec": {
        "containers": [{
          "resources": {
            "limits": {
              "cpu": "2000m",
              "memory": "1Gi"
            },
            "requests": {
              "cpu": "500m",
              "memory": "512Mi"
            }
          }
        }]
      }
    }
  }
}'

# Wait for rollout
kubectl wait --for=condition=Ready ksvc/<service-name> -n <namespace> --timeout=60s
```

#### Issue: Queue Proxy Overhead
**Cause:** Queue proxy adding latency between ingress and app  
**Fix:**
```bash
# Check queue proxy logs
kubectl logs -n <namespace> <pod-name> -c queue-proxy --tail=100

# Check queue proxy metrics
kubectl port-forward -n <namespace> <pod-name> 9090:9090 &
curl http://localhost:9090/metrics | grep -E "queue_proxy_request_duration|queue_proxy_response_time"
kill %1

# Increase queue proxy resources (advanced)
# This requires patching the KnativeServing CR
kubectl get knativeserving knative-serving -n knative-serving -o yaml
```

#### Issue: Too Many Concurrent Requests per Pod
**Cause:** Container concurrency too high, pods overloaded  
**Fix:**
```bash
# Check current concurrency
kubectl get ksvc <service-name> -n <namespace> -o yaml | grep containerConcurrency

# Lower container concurrency to scale out sooner
kubectl patch ksvc <service-name> -n <namespace> --type merge -p '{
  "spec": {
    "template": {
      "spec": {
        "containerConcurrency": 50
      }
    }
  }
}'

# Or use annotation for autoscaling target
kubectl annotate ksvc <service-name> -n <namespace> autoscaling.knative.dev/target="50" --overwrite

# Ensure max-scale is high enough
kubectl annotate ksvc <service-name> -n <namespace> autoscaling.knative.dev/max-scale="20" --overwrite
```

#### Issue: Slow Application Code
**Cause:** Application itself is slow  
**Fix:**
```bash
# Check application logs for slow operations
kubectl logs -n <namespace> <pod-name> -c user-container --tail=200

# Profile application
# - Add application profiling
# - Check database query times
# - Check external API calls
# - Review application metrics

# Ensure application is optimized
# - Database connection pooling
# - Caching
# - Async operations
# - Efficient algorithms

# Scale horizontally if application is CPU-bound
kubectl annotate ksvc <service-name> -n <namespace> \
  autoscaling.knative.dev/target="50" \
  autoscaling.knative.dev/max-scale="10" \
  --overwrite
```

#### Issue: Network Latency
**Cause:** Network issues between components  
**Fix:**
```bash
# Check network policies
kubectl get networkpolicy -A

# Test internal latency
kubectl run test-net --rm -it --image=nicolaka/netshoot --restart=Never -- sh -c "
  time curl -s http://<service-name>.<namespace>.svc.cluster.local
"

# Check DNS resolution time
kubectl run test-dns --rm -it --image=nicolaka/netshoot --restart=Never -- sh -c "
  time nslookup <service-name>.<namespace>.svc.cluster.local
"

# Check for CNI issues
kubectl get pods -n kube-system | grep -E "calico|cilium|flannel"

# Check node network saturation
kubectl top nodes
```

### Step 3: Scale Components

Scale relevant components based on bottleneck:

```bash
# Scale activator
kubectl scale deployment -n knative-serving activator --replicas=2

# Scale Kourier
kubectl scale deployment -n knative-serving 3scale-kourier-gateway --replicas=2

# Scale application
kubectl annotate ksvc <service-name> -n <namespace> autoscaling.knative.dev/min-scale="3" --overwrite
```

### Step 4: Optimize Autoscaling

```bash
# Set lower target for faster scale-up
kubectl annotate ksvc <service-name> -n <namespace> autoscaling.knative.dev/target="50" --overwrite

# Increase max scale
kubectl annotate ksvc <service-name> -n <namespace> autoscaling.knative.dev/max-scale="20" --overwrite

# Set min scale to avoid cold starts
kubectl annotate ksvc <service-name> -n <namespace> autoscaling.knative.dev/min-scale="2" --overwrite
```

## Verification

### 1. Measure Latency Improvement

```bash
# Test latency
time curl -s $URL

# Run load test
hey -n 1000 -c 50 $URL

# Check p50, p95, p99 latencies in output
```

### 2. Check Component Health

```bash
# Check all pods healthy
kubectl get pods -n knative-serving
kubectl get pods -n <namespace> -l serving.knative.dev/service=<service-name>

# Check resource usage
kubectl top pods -n knative-serving
kubectl top pods -n <namespace> -l serving.knative.dev/service=<service-name>
```

### 3. Monitor Over Time

```bash
# Continuous latency monitoring
while true; do
  time curl -s $URL > /dev/null
  sleep 5
done
```

### 4. Check Logs for Errors

```bash
# No errors in activator
kubectl logs -n knative-serving -l app=activator --tail=50

# No errors in Kourier
kubectl logs -n knative-serving -l app=3scale-kourier-gateway --tail=50

# No errors in application
kubectl logs -n <namespace> -l serving.knative.dev/service=<service-name> --tail=50
```

## Prevention

### 1. Right-Size Resources

```yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: example
spec:
  template:
    spec:
      containers:
      - image: example/app
        resources:
          requests:
            cpu: 500m
            memory: 512Mi
          limits:
            cpu: 2000m
            memory: 1Gi
```

### 2. Configure Autoscaling Appropriately

```yaml
metadata:
  annotations:
    autoscaling.knative.dev/target: "50"
    autoscaling.knative.dev/min-scale: "2"
    autoscaling.knative.dev/max-scale: "10"
spec:
  containerConcurrency: 50
```

### 3. Optimize Container Startup

- Use smaller base images
- Pre-warm dependencies at build time
- Optimize readiness probes
- Use liveness probes appropriately

### 4. Monitoring Setup

Key metrics to monitor:
- Request latency (p50, p95, p99)
- Cold start latency
- Request rate
- Pod count vs load
- Component resource usage (activator, Kourier)
- Queue proxy latency

### 5. Load Testing

Regularly test latency under load:

```bash
# Sustained load test
hey -z 5m -c 50 $URL

# Burst load test
hey -n 1000 -c 100 $URL

# Gradual ramp-up
for i in 10 20 50 100; do
  echo "Testing with $i concurrent connections"
  hey -n 1000 -c $i $URL
  sleep 30
done
```

## Performance Tips

1. **Avoid Scale-to-Zero**: For latency-sensitive services, set min-scale > 0
2. **Right-size Concurrency**: Set containerConcurrency based on testing
3. **Fast Startup**: Optimize container startup time
4. **Resource Limits**: Set appropriate CPU/memory limits
5. **Horizontal Scaling**: Scale out rather than up for consistent latency
6. **Connection Pooling**: Use connection pools in application
7. **Caching**: Cache frequently accessed data
8. **Async Operations**: Use async processing for long operations

## Related Alerts

- `KnativeActivatorDown`
- `KnativeKourierDown`
- `KnativeScalingIssues`
- `HighPodCPUUsage`

## Escalation

If latency cannot be reduced within 20 minutes:

1. ✅ Verify all resolution steps completed
2. 🔍 Profile application for bottlenecks
3. 📊 Review infrastructure capacity
4. 🔄 Consider architectural changes
5. 📞 Escalate to platform team
6. 🆘 Page on-call engineer if SLA breached

## Additional Resources

- [Knative Performance Tuning](https://knative.dev/docs/serving/autoscaling/)
- [Knative Observability](https://knative.dev/docs/serving/observability/)
- [Container Optimization](https://knative.dev/docs/serving/services/configure-resource-overrides/)

## Quick Commands Reference

```bash
# Test latency
time curl -s $(kubectl get ksvc <service-name> -n <namespace> -o jsonpath='{.status.url}')

# Load test
hey -n 1000 -c 50 $(kubectl get ksvc <service-name> -n <namespace> -o jsonpath='{.status.url}')

# Check component resources
kubectl top pods -n knative-serving
kubectl top pods -n <namespace> -l serving.knative.dev/service=<service-name>

# Scale components
kubectl scale deployment -n knative-serving activator --replicas=2
kubectl scale deployment -n knative-serving 3scale-kourier-gateway --replicas=2

# Prevent cold starts
kubectl annotate ksvc <service-name> -n <namespace> autoscaling.knative.dev/min-scale="1" --overwrite

# Increase resources
kubectl patch ksvc <service-name> -n <namespace> --type merge -p '{
  "spec": {
    "template": {
      "spec": {
        "containers": [{
          "resources": {
            "requests": {"cpu": "500m", "memory": "512Mi"},
            "limits": {"cpu": "2000m", "memory": "1Gi"}
          }
        }]
      }
    }
  }
}'

# Lower concurrency target
kubectl annotate ksvc <service-name> -n <namespace> autoscaling.knative.dev/target="50" --overwrite
```

---

**Last Updated:** 2025-10-15  
**Version:** 1.0  
**Owner:** Platform Team

