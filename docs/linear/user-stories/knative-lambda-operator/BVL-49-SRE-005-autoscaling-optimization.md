# ⚡ SRE-005: Auto-Scaling Optimization

**Status**: Done
**Linear URL**: https://linear.app/bvlucena/issue/BVL-223/sre-005-auto-scaling-optimization
**Priority**: P1
**Story Points**: 8  
**Linear URL**: https://linear.app/bvlucena/issue/BVL-171/sre-005-auto-scaling-optimization  
**Created**: 2025-10-29  
**Updated**: 2026-01-19  
**Project**: knative-lambda-operator

---

## 📋 User Story

**As an** SRE Engineer  
**I want** to tune Knative auto-scaling parameters  
**So that** functions scale efficiently based on load without waste

---


## 🎯 Acceptance Criteria

- [ ] [ ] Scale-up latency <30s (0→1 pod)
- [ ] [ ] Scale-down graceful (no request drops)
- [ ] [ ] CPU utilization 60-80% (efficient)
- [ ] [ ] Cold start <5s for 95% of requests
- [ ] [ ] No thrashing (rapid scale up/down)
- [ ] --

---


## 📊 Acceptance Criteria

- [ ] Scale-up latency <30s (0→1 pod)
- [ ] Scale-down graceful (no request drops)
- [ ] CPU utilization 60-80% (efficient)
- [ ] Cold start <5s for 95% of requests
- [ ] No thrashing (rapid scale up/down)

---

## ⚙️ Knative Auto-Scaling Configuration

### Default Settings

```yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: parser-${PARSER_ID}
spec:
  template:
    metadata:
      annotations:
        # Scale to zero settings
        autoscaling.knative.dev/scale-to-zero-pod-retention-period: "5m"
        autoscaling.knative.dev/scale-down-delay: "30s"
        
        # Scale up settings
        autoscaling.knative.dev/min-scale: "0"  # Allow scale to zero
        autoscaling.knative.dev/max-scale: "10"  # Max pods
        autoscaling.knative.dev/target: "10"  # Concurrent requests per pod
        autoscaling.knative.dev/metric: "concurrency"
        
        # Advanced tuning
        autoscaling.knative.dev/target-utilization-percentage: "70"
        autoscaling.knative.dev/panic-threshold-percentage: "200"  # 2x target
        autoscaling.knative.dev/panic-window-percentage: "10"  # 10% of stable window
```

### Optimization by Function Type

#### High Traffic Functions (>100 req/min)

```yaml
annotations:
  autoscaling.knative.dev/min-scale: "2"  # Keep 2 pods warm
  autoscaling.knative.dev/max-scale: "20"
  autoscaling.knative.dev/target: "50"  # Higher concurrency
  autoscaling.knative.dev/scale-down-delay: "10m"  # Slow scale down
```

#### Low Traffic Functions (<10 req/min)

```yaml
annotations:
  autoscaling.knative.dev/min-scale: "0"  # Aggressive scale to zero
  autoscaling.knative.dev/max-scale: "5"
  autoscaling.knative.dev/target: "10"
  autoscaling.knative.dev/scale-to-zero-pod-retention-period: "2m"
```

---

## 📊 Monitoring

### Key Metrics

```promql
# Scale-up latency
histogram_quantile(0.95, rate(knative_revision_scale_up_duration_seconds_bucket[5m]))

# Current replica count
sum(kube_pod_status_ready{namespace="knative-lambda"}) 
  by (serving_knative_dev_service)

# Request queue depth
queue_depth{namespace="knative-lambda"}

# CPU utilization
avg(rate(container_cpu_usage_seconds_total{namespace="knative-lambda"}[5m])) 
  by (pod)
```

---

