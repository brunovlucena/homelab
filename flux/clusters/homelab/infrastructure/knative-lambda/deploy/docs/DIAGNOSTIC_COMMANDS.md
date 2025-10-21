# Knative Lambda Scaling Diagnostics

## 🔍 **Immediate Diagnostic Commands**

### **1. Check Service Status**
```bash
# Check the specific service
kubectl get kservice lambda-0307ea43639b461-0197ad6c10b973b -n knative-lambda-prd -o yaml

# Check autoscaler for this service
kubectl get pa -n knative-lambda-prd | grep lambda-0307ea43639b461-0197ad6c10b973b

# Check pod status and health
kubectl get pods -n knative-lambda-prd | grep lambda-0307ea43639b461-0197ad6c10b973b
kubectl describe pod -n knative-lambda-prd <pod-name>
```

### **2. Check Health Probes**
```bash
# Check if health endpoint is responding
kubectl port-forward -n knative-lambda-prd svc/lambda-0307ea43639b461-0197ad6c10b973b 8081:80
curl http://localhost:8081/health

# Check probe logs
kubectl logs -n knative-lambda-prd <pod-name> -c lambda | grep -i health
kubectl logs -n knative-lambda-prd <pod-name> -c queue-proxy | grep -i health
```

### **3. Check Knative-Serving Components**
```bash
# Check autoscaler status
kubectl get pods -n knative-serving | grep autoscaler
kubectl logs -n knative-serving deployment/autoscaler -c autoscaler --tail=100 | grep -i "lambda-0307ea43639b461-0197ad6c10b973b"

# Check activator status
kubectl get pods -n knative-serving | grep activator
kubectl logs -n knative-serving deployment/activator -c activator --tail=100

# Check controller status
kubectl get pods -n knative-serving | grep controller
kubectl logs -n knative-serving deployment/controller -c controller --tail=100
```

### **4. Check Scaling Metrics**
```bash
# Check autoscaler metrics
kubectl port-forward -n knative-serving svc/autoscaler 9090:9090
curl http://localhost:9090/metrics | grep -E "(autoscaler_actual_pods|autoscaler_requested_pods|autoscaler_work_queue_depth)"

# Check queue-proxy metrics
kubectl port-forward -n knative-lambda-prd svc/lambda-0307ea43639b461-0197ad6c10b973b 9091:9091
curl http://localhost:9091/metrics | grep -E "(queue_depth|concurrency|request_count)"
```

### **5. Check Events and Alerts**
```bash
# Check for scaling events
kubectl get events -n knative-lambda-prd --sort-by='.lastTimestamp' | grep -i scale

# Check for probe failures
kubectl get events -n knative-lambda-prd --sort-by='.lastTimestamp' | grep -i probe

# Check Prometheus alerts
kubectl get prometheusrules -n knative-lambda-prd -o yaml | grep -A 10 -B 10 "scaling"
```

## 🚨 **Common Issues and Solutions**

### **Issue 1: Health Probes Failing**
**Symptoms:**
- Pods stuck in `Running` but not `Ready`
- Scaling decisions delayed

**Solution:**
```bash
# Check if health endpoint exists
kubectl exec -n knative-lambda-prd <pod-name> -c lambda -- curl -f http://localhost:8081/health

# If failing, check application logs
kubectl logs -n knative-lambda-prd <pod-name> -c lambda
```

### **Issue 2: Autoscaler Queue Depth High**
**Symptoms:**
- `autoscaler_work_queue_depth > 50`
- Scaling decisions delayed

**Solution:**
```bash
# Check autoscaler resources
kubectl top pods -n knative-serving | grep autoscaler

# Restart autoscaler if needed
kubectl rollout restart deployment/autoscaler -n knative-serving
```

### **Issue 3: Activator Issues**
**Symptoms:**
- High activator error rates
- Requests not reaching services

**Solution:**
```bash
# Check activator health
kubectl get pods -n knative-serving | grep activator
kubectl logs -n knative-serving deployment/activator -c activator --tail=50

# Check activator metrics
kubectl port-forward -n knative-serving svc/activator 9090:9090
curl http://localhost:9090/metrics | grep activator_request_count
```

### **Issue 4: Service Not Scaling Down**
**Symptoms:**
- Pods remain running despite no traffic
- High resource usage

**Solution:**
```bash
# Force scale down (emergency)
kubectl patch kservice lambda-0307ea43639b461-0197ad6c10b973b \
  -n knative-lambda-prd \
  --type='merge' \
  -p='{"spec":{"template":{"metadata":{"annotations":{"autoscaling.knative.dev/minScale":"0"}}}}}'

# Check scaling configuration
kubectl get kservice lambda-0307ea43639b461-0197ad6c10b973b \
  -n knative-lambda-prd \
  -o jsonpath='{.spec.template.metadata.annotations}' | jq
```

## 📊 **Monitoring Queries**

### **Prometheus Queries for Scaling Issues**

```promql
# Check if service is receiving traffic
rate(http_requests_total{namespace="knative-lambda-prd", service="lambda-0307ea43639b461-0197ad6c10b973b"}[5m])

# Check autoscaler work queue depth
autoscaler_work_queue_depth{namespace="knative-lambda-prd"}

# Check scaling decision latency
histogram_quantile(0.95, rate(autoscaler_scaling_decision_latency_seconds_bucket[5m]))

# Check pod readiness
kube_pod_status_ready{namespace="knative-lambda-prd", pod=~"lambda-0307ea43639b461-0197ad6c10b973b.*"}

# Check health probe failures
rate(kube_pod_status_ready{namespace="knative-lambda-prd"}[5m]) == 0
```

## 🔧 **Emergency Fixes**

### **Force Scale Down**
```bash
# Emergency scale to 0
kubectl scale kservice lambda-0307ea43639b461-0197ad6c10b973b \
  -n knative-lambda-prd \
  --replicas=0
```

### **Restart Knative Components**
```bash
# Restart autoscaler
kubectl rollout restart deployment/autoscaler -n knative-serving

# Restart activator
kubectl rollout restart deployment/activator -n knative-serving

# Restart controller
kubectl rollout restart deployment/controller -n knative-serving
```

### **Check and Fix Health Endpoints**
```bash
# Test health endpoint directly
kubectl exec -n knative-lambda-prd <pod-name> -c lambda -- wget -qO- http://localhost:8081/health

# If failing, check if health endpoint is implemented
kubectl exec -n knative-lambda-prd <pod-name> -c lambda -- find /app -name "*.js" -o -name "*.py" -o -name "*.go" | xargs grep -l "health"
``` 