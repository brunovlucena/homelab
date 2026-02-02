# Cloudflare Exporter - Quick Verification Checklist

## ✅ Deployment Status

After committing, Flux will automatically deploy. Verify with:

```bash
# 1. Check HelmRelease status
kubectl get helmrelease cloudflare-exporter -n cloudflare-exporter

# 2. Check pod status
kubectl get pods -n cloudflare-exporter

# 3. Check service
kubectl get svc -n cloudflare-exporter

# 4. Check logs (if pod is running)
kubectl logs -n cloudflare-exporter -l app.kubernetes.io/name=cloudflare-exporter --tail=50
```

## 🔍 Quick Metrics Test

```bash
# Port-forward to exporter
kubectl port-forward -n cloudflare-exporter svc/cloudflare-exporter 8080:8080

# In another terminal, test metrics
curl -s http://localhost:8080/metrics | grep cloudflare_zone | head -10
```

## 📊 Verify in Prometheus

```bash
# Port-forward to Prometheus
kubectl port-forward -n prometheus svc/kube-prometheus-stack-prometheus 9090:9090
```

Then open: http://localhost:9090

1. Go to: **Status > Targets**
2. Find `cloudflare-exporter` job - should show **UP**
3. Go to: **Graph**
4. Query: `up{job="cloudflare-exporter"}` - should return `1`
5. Query: `cloudflare_zone_requests_total` - should show metrics

## 📈 Import Grafana Dashboard

1. Open Grafana
2. Dashboards > Import
3. Dashboard ID: `13133`
4. Select Prometheus datasource
5. Import

## ⚠️ Security Note

The API token is currently hardcoded in the HelmRelease. For production, consider:
- Using External Secrets Operator
- Using Sealed Secrets
- Or using valuesFrom with proper secret management
