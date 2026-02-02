# Alertmanager PagerDuty Setup

This directory contains the Alertmanager configuration for PagerDuty integration.

## Files

- `alertmanagerconfig-pagerduty.yaml` - AlertmanagerConfig CRD for PagerDuty integration
- `prometheusrules/grafana.yaml` - PrometheusRule for Grafana health monitoring

## Setup Instructions

### 1. Create PagerDuty Service Key Secret

You need to create a Kubernetes secret with your PagerDuty service key:

```bash
kubectl create secret generic alertmanager-pagerduty \
  -n prometheus \
  --from-literal=service-key='YOUR_PAGERDUTY_SERVICE_KEY'
```

**Or use SealedSecret (recommended for GitOps):**

1. Get your PagerDuty service key from PagerDuty dashboard
2. Create a SealedSecret using kubeseal:
   ```bash
   echo -n 'YOUR_PAGERDUTY_SERVICE_KEY' | kubeseal \
     --raw \
     --from-file=/dev/stdin \
     --namespace prometheus \
     --name alertmanager-pagerduty
   ```
3. Add the sealed secret to your GitOps repository

### 2. Apply AlertmanagerConfig

```bash
kubectl apply -f k8s/alertmanagerconfig-pagerduty.yaml
```

### 3. Apply Grafana PrometheusRule

```bash
kubectl apply -f k8s/prometheusrules/grafana.yaml
```

### 4. Verify Configuration

```bash
# Check AlertmanagerConfig
kubectl get alertmanagerconfig -n prometheus

# Check PrometheusRule
kubectl get prometheusrule -n prometheus grafana-rules

# Check Alertmanager configuration
kubectl get secret -n prometheus alertmanager-kube-prometheus-stack-alertmanager-generated -o jsonpath='{.data.alertmanager\.yaml}' | base64 -d
```

## Testing

### Test Alert

You can test the PagerDuty integration by triggering a test alert:

```bash
# Create a test alert
kubectl run test-alert --image=curlimages/curl --rm -it --restart=Never -- \
  curl -X POST http://kube-prometheus-stack-alertmanager.prometheus.svc.cluster.local:9093/api/v1/alerts \
  -H "Content-Type: application/json" \
  -d '[{
    "labels": {
      "alertname": "TestAlert",
      "severity": "critical",
      "service": "grafana"
    },
    "annotations": {
      "summary": "Test alert for PagerDuty",
      "description": "This is a test alert to verify PagerDuty integration"
    }
  }]'
```

## Troubleshooting

### Alertmanager not receiving alerts

1. Check if PrometheusRule is loaded:
   ```bash
   kubectl get prometheusrule -n prometheus
   ```

2. Check Prometheus targets:
   ```bash
   kubectl port-forward -n prometheus svc/kube-prometheus-stack-prometheus 9090:9090
   # Then visit http://localhost:9090/targets
   ```

3. Check Alertmanager logs:
   ```bash
   kubectl logs -n prometheus -l app.kubernetes.io/name=alertmanager
   ```

### PagerDuty not receiving alerts

1. Verify secret exists:
   ```bash
   kubectl get secret -n prometheus alertmanager-pagerduty
   ```

2. Check AlertmanagerConfig:
   ```bash
   kubectl get alertmanagerconfig -n prometheus pagerduty-config -o yaml
   ```

3. Check Alertmanager configuration:
   ```bash
   kubectl get secret -n prometheus alertmanager-kube-prometheus-stack-alertmanager-generated -o jsonpath='{.data.alertmanager\.yaml}' | base64 -d | grep -A 10 pagerduty
   ```

## Current Status

- ✅ PrometheusRule for Grafana created
- ✅ AlertmanagerConfig for PagerDuty created
- ⚠️ PagerDuty secret needs to be created (see step 1 above)
- ⚠️ Alertmanager HelmRelease updated (needs Flux sync)
