# 🚀 Flagger - Progressive Delivery

Flagger is a progressive delivery tool that automates the release process for applications running on Kubernetes.

## Features

- ✅ **Canary Deployments** - Gradually shift traffic to new versions
- ✅ **A/B Testing** - Route traffic based on headers/cookies
- ✅ **Blue/Green Deployments** - Switch between versions instantly
- ✅ **Automated Rollbacks** - Based on Prometheus metrics
- ✅ **Linkerd Integration** - Uses TrafficSplit for traffic management

## Architecture

```
Git Push → Flux → Deployment Updated → Flagger Detects Change
                                              ↓
                                       Creates Canary
                                              ↓
                                       Linkerd TrafficSplit
                                              ↓
                        ┌──────────────────────┴──────────────────────┐
                        ↓                                              ↓
                  Primary (90%)                                  Canary (10%)
                        ↓                                              ↓
                   Prometheus Metrics Analysis
                        ↓
            ┌───────────┴───────────┐
            ↓                       ↓
    Metrics Good                Metrics Bad
    Promote Canary             Rollback to Primary
```

## Installation

This Flagger installation is automatically managed by Flux. Simply commit these files to Git and Flux will install Flagger.

```bash
# Check installation status
kubectl -n flagger-system get pods
kubectl -n flagger-system get helmrelease flagger

# View logs
kubectl -n flagger-system logs deployment/flagger -f
```

## Configuration

### Metrics Server

Flagger uses Prometheus for metrics analysis:
- **Prometheus URL**: `http://prometheus-kube-prometheus-prometheus.prometheus.svc:9090`
- **Default Metrics**: Success rate, request duration, error rate

### Mesh Provider

- **Provider**: Linkerd
- **Traffic Management**: SMI TrafficSplit API
- **mTLS**: Enabled via Linkerd injection

## Usage

See the main guide: `/Users/brunolucena/workspace/bruno/docs/CANARY_AND_AB_TESTING_GUIDE.md`

### Quick Example

```yaml
apiVersion: flagger.app/v1beta1
kind: Canary
metadata:
  name: my-app
  namespace: default
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: my-app
  service:
    port: 8080
  analysis:
    interval: 30s
    threshold: 5
    maxWeight: 50
    stepWeight: 10
    metrics:
    - name: request-success-rate
      thresholdRange:
        min: 99
    - name: request-duration
      thresholdRange:
        max: 500
```

## Monitoring

```bash
# List all canaries
kubectl get canaries --all-namespaces

# Describe a specific canary
kubectl -n homepage describe canary homepage-api

# Watch canary progress
kubectl -n homepage get canary homepage-api --watch

# View traffic split
kubectl -n homepage get trafficsplit
```

## Troubleshooting

### Canary stuck in "Progressing"

```bash
# Check Flagger logs
kubectl -n flagger-system logs deployment/flagger -f

# Check metrics server connectivity
kubectl -n flagger-system exec deployment/flagger -- \
  wget -O- http://prometheus-kube-prometheus-prometheus.prometheus.svc:9090/api/v1/query?query=up
```

### Metrics not working

```bash
# Verify Prometheus has metrics for your app
kubectl -n prometheus port-forward svc/prometheus-kube-prometheus-prometheus 9090:9090

# Open browser to http://localhost:9090
# Query: rate(http_requests_total{namespace="homepage"}[5m])
```

## Resources

- [Flagger Documentation](https://docs.flagger.app/)
- [Linkerd + Flagger Tutorial](https://docs.flagger.app/tutorials/linkerd-progressive-delivery)
- [SMI TrafficSplit Spec](https://github.com/servicemeshinterface/smi-spec/blob/main/apis/traffic-split/v1alpha2/traffic-split.md)

