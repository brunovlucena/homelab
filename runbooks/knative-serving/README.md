# 📚 Knative Serving Runbooks

Comprehensive operational runbooks for troubleshooting and resolving Knative Serving issues in the homelab Kubernetes cluster.

## Overview

Knative Serving enables serverless workloads on Kubernetes with automatic scaling (including scale-to-zero):
- **Namespace**: knative-serving
- **Architecture**: Microservices-based with controller, autoscaler, activator, and webhooks
- **Ingress**: Kourier (lightweight ingress for Knative)
- **Autoscaling**: Scale-to-zero enabled with 30s grace period
- **Metrics Backend**: Prometheus

## Quick Reference

| Alert | Severity | Impact | Runbook |
|-------|----------|--------|---------|
| KnativeServingDown | Critical | Serverless workloads unavailable | [serving-down.md](./serving-down.md) |
| KnativeActivatorDown | Critical | Cannot scale from zero | [activator-down.md](./activator-down.md) |
| KnativeAutoscalerDown | Critical | No autoscaling | [autoscaler-down.md](./autoscaler-down.md) |
| KnativeControllerDown | Critical | Cannot manage services | [controller-down.md](./controller-down.md) |
| KnativeWebhookDown | Critical | Cannot create/update services | [webhook-down.md](./webhook-down.md) |
| KnativeKourierDown | Critical | Ingress unavailable | [kourier-down.md](./kourier-down.md) |
| KnativeScalingIssues | Warning | Services not scaling properly | [scaling-issues.md](./scaling-issues.md) |
| KnativeHighLatency | Warning | Slow request processing | [high-latency.md](./high-latency.md) |

## Runbooks

### 🚨 Critical Issues

#### [Knative Serving Down](./serving-down.md)
Complete Knative Serving outage - all serverless workloads unavailable.

**Quick Check:**
```bash
kubectl get pods -n knative-serving
```

**Quick Fix:**
```bash
# Restart all components
kubectl rollout restart deployment -n knative-serving
```

---

#### [Activator Down](./activator-down.md)
Activator unavailable - cannot scale services from zero.

**Quick Check:**
```bash
kubectl get pods -n knative-serving -l app=activator
kubectl logs -n knative-serving -l app=activator --tail=50
```

**Quick Fix:**
```bash
# Restart activator
kubectl rollout restart deployment -n knative-serving activator
```

---

#### [Autoscaler Down](./autoscaler-down.md)
Autoscaler unavailable - services not scaling automatically.

**Quick Check:**
```bash
kubectl get pods -n knative-serving -l app=autoscaler
kubectl logs -n knative-serving -l app=autoscaler --tail=50
```

**Quick Fix:**
```bash
# Restart autoscaler
kubectl rollout restart deployment -n knative-serving autoscaler
```

---

#### [Controller Down](./controller-down.md)
Controller unavailable - cannot create or update Knative services.

**Quick Check:**
```bash
kubectl get pods -n knative-serving -l app=controller
kubectl logs -n knative-serving -l app=controller --tail=50
```

**Quick Fix:**
```bash
# Restart controller
kubectl rollout restart deployment -n knative-serving controller
```

---

#### [Webhook Down](./webhook-down.md)
Webhook unavailable - cannot validate/mutate Knative resources.

**Quick Check:**
```bash
kubectl get pods -n knative-serving -l app=webhook
kubectl logs -n knative-serving -l app=webhook --tail=50
```

**Quick Fix:**
```bash
# Restart webhook
kubectl rollout restart deployment -n knative-serving webhook
```

---

#### [Kourier Ingress Down](./kourier-down.md)
Kourier ingress unavailable - external traffic cannot reach services.

**Quick Check:**
```bash
kubectl get pods -n knative-serving -l app=3scale-kourier-gateway
kubectl logs -n knative-serving -l app=3scale-kourier-gateway --tail=50
```

**Quick Fix:**
```bash
# Restart Kourier
kubectl rollout restart deployment -n knative-serving 3scale-kourier-gateway
```

---

### ⚠️ Warning Issues

#### [Scaling Issues](./scaling-issues.md)
Services not scaling up/down properly, stuck at wrong replica count.

**Quick Check:**
```bash
kubectl get ksvc -A
kubectl get pods -A -l serving.knative.dev/service
kubectl get pa -A  # PodAutoscalers
```

**Quick Fix:**
```bash
# Check autoscaler logs
kubectl logs -n knative-serving -l app=autoscaler --tail=100
# Force service reconciliation
kubectl annotate ksvc <service-name> -n <namespace> reconcile=$(date +%s) --overwrite
```

---

#### [High Latency](./high-latency.md)
Requests experiencing high latency through Knative.

**Quick Check:**
```bash
kubectl logs -n knative-serving -l app=activator --tail=100
kubectl top pods -n knative-serving
```

**Quick Fix:**
```bash
# Check activator and gateway resources
kubectl top pods -n knative-serving -l app=activator
kubectl scale deployment -n knative-serving activator --replicas=2
```

---

## Common Troubleshooting Commands

### Check Overall Health
```bash
# All pods status
kubectl get pods -n knative-serving

# Resource usage
kubectl top pods -n knative-serving

# Recent events
kubectl get events -n knative-serving --sort-by='.lastTimestamp' | head -20
```

### Check Knative Services
```bash
# List all Knative services
kubectl get ksvc -A

# Get service details
kubectl describe ksvc <service-name> -n <namespace>

# Check service routes
kubectl get route -A

# Check service revisions
kubectl get revision -A

# Check service configurations
kubectl get configuration -A
```

### Check Autoscaling
```bash
# Check PodAutoscalers
kubectl get pa -A

# Describe PodAutoscaler
kubectl describe pa <service-name> -n <namespace>

# Check autoscaler config
kubectl get configmap -n knative-serving config-autoscaler -o yaml
```

### Check Component Logs
```bash
# Controller logs
kubectl logs -n knative-serving -l app=controller --tail=100

# Autoscaler logs
kubectl logs -n knative-serving -l app=autoscaler --tail=100

# Activator logs
kubectl logs -n knative-serving -l app=activator --tail=100

# Webhook logs
kubectl logs -n knative-serving -l app=webhook --tail=100

# Kourier logs
kubectl logs -n knative-serving -l app=3scale-kourier-gateway --tail=100
```

### Check Configuration
```bash
# List all Knative configs
kubectl get configmap -n knative-serving | grep config-

# Autoscaler config
kubectl get configmap -n knative-serving config-autoscaler -o yaml

# Network config
kubectl get configmap -n knative-serving config-network -o yaml

# Features config
kubectl get configmap -n knative-serving config-features -o yaml

# Observability config
kubectl get configmap -n knative-serving config-observability -o yaml
```

### Test Service Deployment
```bash
# Deploy test service
cat <<EOF | kubectl apply -f -
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: hello
  namespace: default
spec:
  template:
    spec:
      containers:
      - image: gcr.io/knative-samples/helloworld-go
        env:
        - name: TARGET
          value: "World"
EOF

# Check service status
kubectl get ksvc hello -n default

# Get service URL
kubectl get ksvc hello -n default -o jsonpath='{.status.url}'

# Test service
curl $(kubectl get ksvc hello -n default -o jsonpath='{.status.url}')

# Cleanup
kubectl delete ksvc hello -n default
```

## Architecture

```
┌─────────────────────────────────────────────┐
│           External Traffic                  │
└────────────────┬────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────┐
│    Kourier Ingress (3scale-kourier)         │
│    - Routes external traffic                │
└────────────────┬────────────────────────────┘
                 │
        ┌────────┴────────┐
        │                 │
        ▼                 ▼
┌──────────────┐   ┌──────────────┐
│  Activator   │   │   Service    │
│  (if scaled  │   │    Pods      │
│   to zero)   │   │ (if running) │
└──────┬───────┘   └──────────────┘
       │
       │ Metrics
       ▼
┌──────────────┐
│  Autoscaler  │◄── Prometheus Metrics
│  - Scales    │
│    pods      │
└──────┬───────┘
       │
       │ Updates
       ▼
┌──────────────┐
│  Controller  │◄── Manages Knative Resources
│  - Revisions │
│  - Routes    │
└──────────────┘
       ▲
       │ Validates
       │
┌──────────────┐
│   Webhook    │◄── Admission Control
└──────────────┘
```

## Configuration

**Location**: `flux/clusters/homelab/infrastructure/knative-serving/knativeserving.yaml`

**Key Settings**:
- Ingress: Kourier
- Scale-to-zero: Enabled
- Scale-to-zero grace period: 30s
- Initial scale: 1
- Container concurrency target: 100
- RPS target: 200
- Max scale-up rate: 1000
- Max scale-down rate: 2.0
- Metrics backend: Prometheus

## Autoscaling Configuration

### Scale-to-Zero Settings
```yaml
enable-scale-to-zero: "true"
allow-zero-initial-scale: "true"
scale-to-zero-grace-period: "30s"
scale-to-zero-pod-retention-period: "0s"
```

### Concurrency Settings
```yaml
container-concurrency-target-default: "100"
container-concurrency-target-percentage: "70"
requests-per-second-target-default: "200"
```

### Scale Rate Limits
```yaml
max-scale-up-rate: "1000.0"
max-scale-down-rate: "2.0"
initial-scale: "1"
```

## Performance Tips

1. **Cold Start Optimization**: Keep frequently-used services from scaling to zero
2. **Concurrency Tuning**: Adjust `containerConcurrency` based on workload
3. **Resource Limits**: Set appropriate CPU/memory for predictable scaling
4. **Readiness Probes**: Ensure fast startup for quick scaling
5. **Min/Max Scale**: Set bounds to prevent over-scaling

## Escalation Matrix

| Issue | First Response | Escalation Time | Escalate To |
|-------|---------------|-----------------|-------------|
| Complete outage | Restart all components | 15 minutes | Platform team |
| Activator down | Restart activator | 10 minutes | Platform team |
| Scaling issues | Check autoscaler logs | 30 minutes | Platform team |
| High latency | Scale activator | 20 minutes | Performance team |
| Controller down | Restart controller | 15 minutes | Platform team |

## Related Documentation

- [Knative Serving Configuration](../../../flux/clusters/homelab/infrastructure/knative-serving/knativeserving.yaml)
- [Knative Operator Runbooks](../knative-operator/README.md)
- [Architecture Overview](../../../ARCHITECTURE.md)
- [Knative Official Docs](https://knative.dev/docs/serving/)

## Monitoring & Alerts

Knative Serving exports metrics to Prometheus. Key metrics:
- Request latency (activator, gateway)
- Autoscaler decisions
- Revision scale events
- Queue proxy metrics

Access Grafana dashboards for Knative Serving visualization.

## Support

For issues not covered by these runbooks:
1. Check component logs
2. Review KnativeServing CR status: `kubectl describe knativeserving -n knative-serving`
3. Consult [Knative documentation](https://knative.dev/docs/serving/)
4. Check [GitHub issues](https://github.com/knative/serving/issues)

---

**Last Updated**: 2025-10-15  
**Knative Serving Version**: Managed by Operator 1.16.3  
**Maintainer**: Homelab Platform Team

