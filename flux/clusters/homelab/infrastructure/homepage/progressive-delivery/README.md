# 🚀 Homepage Progressive Delivery

This directory contains configurations for progressive delivery (canary deployments and A/B testing) for the homepage application.

## 📁 Files

- **`canary-api.yaml`** - Automated canary deployment for API backend
- **`canary-frontend.yaml`** - Automated canary deployment for React frontend
- **`trafficsplit-manual-example.yaml`** - Manual traffic splitting examples
- **`httproute-ab-test-example.yaml`** - A/B testing with header/cookie routing

## 🎯 Quick Start

### 1. Install Flagger (if not already installed)

```bash
# Flagger should be installed via Flux
kubectl get pods -n flagger-system

# If not installed, commit and push the flagger directory
cd /Users/brunolucena/workspace/bruno/repos/homelab
git add flux/clusters/homelab/infrastructure/flagger
git commit -m "feat: add Flagger for progressive delivery"
git push
```

### 2. Enable Canary Deployment for API

To enable automated canary deployments for the homepage API:

```bash
# Apply the canary configuration
kubectl apply -f canary-api.yaml

# Watch the canary resource
kubectl -n homepage get canary homepage-api --watch
```

### 3. Deploy a New Version

Simply update your deployment (via Flux or manually):

```bash
# Update image tag in your Helm values or deployment
# For example, in chart/values.yaml:
# api:
#   image:
#     tag: v1.1.0  # New version

# Commit and push
git add .
git commit -m "feat: deploy homepage-api v1.1.0"
git push

# Flagger will automatically:
# 1. Create a canary deployment
# 2. Gradually shift traffic (10% → 20% → 30% → 40% → 50%)
# 3. Monitor metrics at each step
# 4. Promote or rollback based on metrics
```

### 4. Monitor the Rollout

```bash
# Watch canary progress
kubectl -n homepage get canary homepage-api --watch

# Check events
kubectl -n homepage describe canary homepage-api

# View traffic split
kubectl -n homepage get trafficsplit

# Monitor with Linkerd
linkerd viz stat deploy/homepage-api-primary deploy/homepage-api-canary -n homepage

# View live traffic
linkerd viz tap deploy/homepage-api-canary -n homepage
```

## 📊 How Canary Deployment Works

### Automatic Process

1. **Detection** (0s)
   - You push new deployment to Git
   - Flux applies the change
   - Flagger detects the new version

2. **Initialization** (0-30s)
   - Flagger creates canary deployment
   - Runs pre-rollout checks
   - Initializes metrics baseline

3. **Progressive Rollout** (5-10 minutes)
   ```
   Primary: 100% → 90% → 80% → 70% → 60% → 50% → 0%
   Canary:    0% → 10% → 20% → 30% → 40% → 50% → 100%
   ```
   - Each step lasts 1 minute
   - Metrics are evaluated at each step
   - If metrics fail, automatic rollback

4. **Promotion** (30s)
   - Canary becomes primary
   - Old primary is scaled down
   - Rollout complete

### Metrics Evaluation

At each step, Flagger checks:

✅ **Success Rate** >= 99%
- HTTP 2xx, 3xx responses / Total requests

✅ **Latency P99** <= 500ms
- 99th percentile response time

✅ **Error Rate** <= 1%
- HTTP 5xx responses / Total requests

If any metric fails:
- ❌ Automatic rollback
- 🔔 Alert triggered
- 📝 Event logged

## 🧪 A/B Testing

For feature testing with specific users:

### Example 1: Header-Based Routing

```bash
# Beta users get new version
curl -H "X-Beta-User: true" https://bruno.lucena.dev/api/health

# Normal users get stable version
curl https://bruno.lucena.dev/api/health
```

### Example 2: Cookie-Based Routing

```javascript
// In your frontend code
// Enable canary for this user
document.cookie = "canary=enabled; path=/; max-age=86400";

// Make requests - they'll hit the canary
fetch('/api/data');
```

### Apply A/B Testing

```bash
# Apply the HTTPRoute for A/B testing
kubectl apply -f httproute-ab-test-example.yaml

# Test it
curl -H "X-Beta-User: true" http://homepage-api.homepage.svc:8080/health
```

## 🎛️ Configuration

### Adjust Rollout Speed

Edit `canary-api.yaml`:

```yaml
analysis:
  interval: 30s      # How often to evaluate metrics (30s = faster, 2m = slower)
  threshold: 5       # Number of successful checks needed (higher = safer)
  stepWeight: 10     # Traffic increment % (10 = gradual, 25 = aggressive)
  maxWeight: 50      # Max canary traffic (50 = safe, 100 = full canary)
```

**Profiles:**

| Profile | Interval | Threshold | StepWeight | Duration |
|---------|----------|-----------|------------|----------|
| **Conservative** | 2m | 10 | 5 | ~40 min |
| **Balanced** (default) | 1m | 5 | 10 | ~10 min |
| **Aggressive** | 30s | 3 | 20 | ~5 min |
| **YOLO** | 30s | 1 | 50 | ~2 min |

### Add Custom Metrics

Edit `canary-api.yaml` to add custom Prometheus queries:

```yaml
metrics:
- name: custom-business-metric
  templateRef:
    name: custom-template
    namespace: flagger-system
  thresholdRange:
    min: 95
  interval: 1m
```

Then create the template:

```yaml
apiVersion: flagger.app/v1beta1
kind: MetricTemplate
metadata:
  name: custom-template
  namespace: flagger-system
spec:
  provider:
    type: prometheus
    address: http://prometheus-kube-prometheus-prometheus.prometheus.svc:9090
  query: |
    your_custom_prometheus_query{namespace="{{ namespace }}"}
```

## 🔧 Troubleshooting

### Canary Stuck in "Progressing"

```bash
# Check Flagger logs
kubectl -n flagger-system logs deployment/flagger -f

# Check if metrics are available
kubectl -n homepage port-forward svc/prometheus-kube-prometheus-prometheus 9090:9090
# Open http://localhost:9090 and query:
# rate(http_requests_total{namespace="homepage"}[5m])
```

### Metrics Not Working

```bash
# Verify your app exports Prometheus metrics
kubectl -n homepage port-forward deploy/homepage-api 8080:8080
curl http://localhost:8080/metrics

# Check if Prometheus scrapes your app
# In Prometheus UI, check Targets page
```

### Manual Rollback

```bash
# If automatic rollback isn't working, manually rollback:
kubectl -n homepage set image deployment/homepage-api api=ghcr.io/brunovlucena/bruno-site-api:v1.0.0

# Or delete the canary to stop the rollout
kubectl -n homepage delete canary homepage-api
```

## 📈 Monitoring

### Grafana Dashboard

Create a dashboard with these queries:

```promql
# Traffic distribution
sum(rate(http_requests_total{namespace="homepage",deployment=~"homepage-api-primary"}[5m]))
sum(rate(http_requests_total{namespace="homepage",deployment=~"homepage-api-canary"}[5m]))

# Success rate comparison
sum(rate(http_requests_total{namespace="homepage",deployment="homepage-api-primary",status!~"5.*"}[5m]))
/
sum(rate(http_requests_total{namespace="homepage",deployment="homepage-api-primary"}[5m]))

# Latency comparison
histogram_quantile(0.99, rate(http_request_duration_seconds_bucket{deployment="homepage-api-primary"}[5m]))
histogram_quantile(0.99, rate(http_request_duration_seconds_bucket{deployment="homepage-api-canary"}[5m]))
```

## 🎓 Best Practices

1. **Start Conservative** - Use slower intervals and higher thresholds
2. **Monitor Business Metrics** - Not just technical metrics
3. **Test in Staging First** - Validate canary config before production
4. **Use Feature Flags** - Combine with application-level feature toggles
5. **Alert on Rollbacks** - Know when automatic rollbacks happen
6. **Document Rollout Policies** - Team should understand the process

## 📚 Resources

- [Main Guide](/Users/brunolucena/workspace/bruno/docs/CANARY_AND_AB_TESTING_GUIDE.md)
- [Flagger Docs](https://docs.flagger.app/)
- [Linkerd Traffic Split](https://linkerd.io/2/features/traffic-split/)
- [Gateway API HTTPRoute](https://gateway-api.sigs.k8s.io/api-types/httproute/)

