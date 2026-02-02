# Demo-Notifi Dashboard Deployment Status

## ✅ Completed

1. **Dashboard JSON Created**: `demo-notifi-dashboard.json` - Comprehensive Grafana dashboard with:
   - Event Publishing Overview (Total Events, Success Rate, Error Rate, P95 Duration, Event Rate, Scenario Success)
   - Event Throughput & Performance (Rate by Type, Duration Percentiles, Success Rate by Type)
   - Build Events (Start, Complete, Failed, Timeout, Cancel)
   - Service Events (Create, Update, Delete, Deleted)
   - Parser Events (Complete, Failed)
   - Status Events (Update, Health Check)
   - Scenario Metrics (All 4 scenarios with success rates)

2. **Kustomization Updated**: Added `demo-notifi-dashboard-configmap.yaml` to resources list

3. **ConfigMap File**: Needs to be created (see instructions below)

## 📋 Next Steps

### 1. Create ConfigMap File

Run this command in the dashboards directory:

```bash
cd /Users/brunolucena/workspace/bruno/repos/homelab/flux/infrastructure/prometheus-operator/k8s/dashboards

cat > demo-notifi-dashboard-configmap.yaml << 'EOF'
---
# Demo-Notifi Dashboard
# This dashboard is automatically discovered by Grafana via the sidecar
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-dashboard-demo-notifi
  namespace: prometheus
  labels:
    grafana_dashboard: "1"
data:
  demo-notifi-dashboard.json: |
EOF

cat demo-notifi-dashboard.json | sed 's/^/    /' >> demo-notifi-dashboard-configmap.yaml
```

### 2. Commit and Push

```bash
cd /Users/brunolucena/workspace/bruno/repos/homelab
git add -A
git commit -m "Add demo-notifi Grafana dashboard for K6 test metrics"
git push
```

### 3. Deploy Dashboard (Flux will auto-sync)

```bash
# Trigger Flux reconciliation
flux reconcile source git homelab

# Wait for sync (or check status)
flux get sources git

# Verify ConfigMap is created
kubectl get configmap -n prometheus grafana-dashboard-demo-notifi

# Check Grafana sidecar logs
kubectl logs -n prometheus -l app.kubernetes.io/name=grafana -c grafana-sc-dashboard | grep demo-notifi
```

### 4. Run K6 Tests

```bash
# Apply all K6 tests
kubectl apply -k /Users/brunolucena/workspace/bruno/repos/demo-notifi/k6/k8s/

# Check test runs
kubectl get testrun -n demo-notifi

# Watch test execution
kubectl get testrun -n demo-notifi -w

# View test logs
kubectl logs -f -n demo-notifi -l test-type=event_simulator
kubectl logs -f -n demo-notifi -l test-type=scenario

# Check test status
kubectl describe testrun -n demo-notifi demo-notifi-events
kubectl describe testrun -n demo-notifi demo-notifi-scenario-successful-build
```

### 5. View Dashboard in Grafana

1. Access Grafana: `http://localhost:30040` (NodePort) or port-forward:
   ```bash
   kubectl port-forward -n prometheus svc/kube-prometheus-stack-grafana 3000:3000
   ```

2. Navigate to: Dashboards → Search for "Demo-Notifi"

3. Dashboard UID: `demo-notifi`

### 6. Verify Metrics in Prometheus

```bash
# Port forward to Prometheus
kubectl port-forward -n prometheus svc/kube-prometheus-stack-prometheus 9090:9090

# Query metrics
# Total events: sum(demo_notifi_cloudevent_published_total)
# Success rate: avg(demo_notifi_cloudevent_publish_success)
# P95 duration: histogram_quantile(0.95, sum by (le) (rate(demo_notifi_cloudevent_publish_duration_ms_bucket[5m])))
```

## 📊 Dashboard Features

- **Real-time metrics** from K6 tests
- **Event type filtering** via template variables
- **Scenario success tracking** for all 4 scenarios
- **Performance metrics** (P50/P95/P99 latencies)
- **Error rate monitoring**
- **Throughput visualization**

## 🔗 Related Dashboards

- Knative Lambda Operator: `/d/knative-lambda-operator`
- K6 Knative Lambda: `/d/k6-knative-lambda`
