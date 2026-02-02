# 🔄 DEVOPS-001: Observability Setup

**Priority**: P0 | **Status**: ✅ Implemented  | **Story Points**: 8
**Linear URL**: https://linear.app/bvlucena/issue/BVL-233/devops-001-observability-setup

---

## 📋 User Story

**As a** DevOps Engineer  
**I want to** set up comprehensive observability (metrics, logs, traces)  
**So that** we can monitor platform health and debug issues quickly

---

## 🎯 Acceptance Criteria

- [ ] Prometheus scraping all components
- [ ] Grafana dashboards for key workflows
- [ ] Alerts configured for critical metrics
- [ ] OpenTelemetry tracing enabled
- [ ] Log aggregation with structured logging
- [ ] SLO/SLI dashboards tracking 99.9% availability

---

## 🏗️ Architecture

```
┌────────────────────────────────────────────────────────────────┐
│                   OBSERVABILITY STACK                          │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│  METRICS (Prometheus)                                          │
│  ├─ Builder Service (:8080/metrics)                            │
│  ├─ Kaniko Jobs (sidecar metrics-pusher)                       │
│  ├─ RabbitMQ (:15692/metrics)                                  │
│  ├─ Knative Serving (autoscaler, activator)                    │
│  └─ Node Exporter (system metrics)                             │
│                                                                │
│  LOGS (Loki)                                                   │
│  ├─ Builder Service (JSON structured)                          │
│  ├─ Kaniko Jobs (build logs)                                   │
│  ├─ Function pods (application logs)                           │
│  └─ Kubernetes events                                          │
│                                                                │
│  TRACES (Tempo)                                                │
│  ├─ OpenTelemetry SDK (Go)                                     │
│  ├─ Trace context propagation                                  │
│  ├─ Span creation (build pipeline stages)                      │
│  └─ Exemplars (link metrics → traces)                          │
│                                                                │
│  DASHBOARDS (Grafana)                                          │
│  ├─ Platform Overview                                          │
│  ├─ Build Pipeline                                             │
│  ├─ Function Performance                                       │
│  ├─ Cost Analysis                                              │
│  └─ SLO/SLI Tracking                                           │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

---

## 📋 Setup Steps

### 1. Deploy Prometheus Operator

```bash
# Add Prometheus Operator via Helm
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace prometheus \
  --create-namespace \
  --values deploy/monitoring/prometheus-values.yaml
```

### 2. Configure ServiceMonitor

```yaml
# deploy/templates/servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: knative-lambda-builder
  namespace: knative-lambda
spec:
  selector:
    matchLabels:
      app: knative-lambda-builder
  endpoints:
  - port: http
    path: /metrics
    interval: 30s
```

### 3. Deploy Grafana Dashboards

```bash
# Import dashboards as ConfigMaps
kubectl create configmap grafana-dashboard-knative-lambda \
  --from-file=dashboards/knative-lambda-overview.json \
  --namespace=prometheus \
  --dry-run=client -o yaml | kubectl apply -f -
```

### 4. Configure Alertmanager

```yaml
# deploy/monitoring/alertmanager-config.yaml
route:
  receiver: 'slack-knative-lambda'
  group_by: ['alertname', 'cluster', 'service']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 4h

receivers:
- name: 'slack-knative-lambda'
  slack_configs:
  - api_url: '${SLACK_WEBHOOK_URL}'
    channel: '#knative-lambda-alerts'
    title: '{{ .GroupLabels.alertname }}'
    text: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'
```

---

## 📊 Key Metrics

### Golden Signals

```promql
# 1. Latency - Build duration p95
histogram_quantile(0.95, 
  rate(build_duration_seconds_bucket[5m])
)

# 2. Traffic - Builds per second
rate(builds_total[5m])

# 3. Errors - Build failure rate
rate(build_failures_total[5m]) / rate(builds_total[5m])

# 4. Saturation - Kaniko job utilization
kaniko_jobs_running / kaniko_jobs_limit
```

### Business Metrics

```promql
# Build success rate
sum(rate(builds_total{status="success"}[5m])) / 
sum(rate(builds_total[5m])) * 100

# Cost per build
(sum(rate(kaniko_cpu_seconds_total[1h])) * 0.04) / 
sum(rate(builds_total[1h]))

# Cold start p95
histogram_quantile(0.95, 
  rate(function_cold_start_seconds_bucket[5m])
)
```

---

## 🚨 Critical Alerts

```yaml
groups:
- name: knative-lambda
  rules:
  - alert: BuildFailureRateHigh
    expr: | rate(build_failures_total[5m]) / rate(builds_total[5m]) > 0.10
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: "Build failure rate {{ $value }}% (threshold: 10%)"
      
  - alert: BuilderServiceDown
    expr: | up{job="knative-lambda-builder"} == 0
    for: 2m
    labels:
      severity: critical
    annotations:
      summary: "Builder service is down"
```

---

## 💡 Pro Tips

- Use Prometheus recording rules for expensive queries
- Enable Grafana alerting for anomaly detection
- Use trace exemplars to jump from metrics → traces
- Set up retention policies (metrics: 30d, logs: 7d, traces: 3d)
- Monitor cardinality to avoid high-cardinality issues

---

## 📈 Performance Requirements

- **Metrics Collection**: < 30 seconds latency
- **Query Response Time**: < 3 seconds (95th percentile)
- **Log Ingestion**: < 5 seconds delay
- **Trace Collection**: < 100ms overhead
- **Dashboard Load Time**: < 2 seconds

---

## 📚 Related Documentation

- [DEVOPS-002: GitOps Deployment](DEVOPS-002-gitops-deployment.md)
- [DEVOPS-003: Multi-Environment Management](DEVOPS-003-multi-environment.md)
- Prometheus Documentation: https://prometheus.io/docs/
- Grafana Documentation: https://grafana.com/docs/

---

**Last Updated**: October 29, 2025  
**Owner**: DevOps Team  
**Status**: Production Ready
