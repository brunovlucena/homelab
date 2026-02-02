# ✅ Cloudflare Dashboard Setup - Complete

## What Was Done

### 1. ✅ Updated Kubernetes Secret
- **Secret**: `cloudflare` in namespace `cloudflare-tunnel`
- **Key**: `cloudflare-api-token`
- **Value**: Updated with token from `~/.zshrc` (`CLOUDFLARE_API_TOKEN`)
- **Command used**:
  ```bash
  kubectl create secret generic cloudflare -n cloudflare-tunnel \
    --from-literal=cloudflare-api-token='H0wJl0PtiuKTb6IexMiQ4zS2nn__HtHBKlF2S-gJ' \
    --dry-run=client -o yaml | kubectl apply -f -
  ```

### 2. ✅ Updated Cloudflare Exporter HelmRelease
- **File**: `flux/infrastructure/cloudflare-exporter/helmrelease.yaml`
- **Changes**:
  - Added `CF_API_TOKEN` to env array with token value
  - Configured `SCRAPE_INTERVAL` to 60 seconds
  - Token is now properly configured for the exporter

### 3. ✅ Updated Grafana Dashboard ConfigMap
- **File**: `flux/infrastructure/prometheus-operator/k8s/dashboards/homepage-metrics-dashboard-configmap.yaml`
- **Changes**:
  - Regenerated ConfigMap with updated dashboard JSON
  - Added new "Cloudflare Performance (vs Origin)" section with 5 panels:
    1. Request Rate: Cloudflare vs Origin
    2. Estimated Cache Hit Ratio
    3. Origin Load Reduction
    4. Cache Hit vs Miss Breakdown
    5. Bandwidth: Cloudflare vs Origin

### 4. ✅ Updated Dashboard JSON
- **File**: `flux/infrastructure/prometheus-operator/k8s/dashboards/homepage-metrics-dashboard.json`
- **Changes**: Added Cloudflare comparison panels (already done in previous step)

## Next Steps

### 1. Commit and Push Changes
```bash
cd /Users/brunolucena/workspace/bruno/repos/homelab
git add flux/infrastructure/cloudflare-exporter/helmrelease.yaml
git add flux/infrastructure/prometheus-operator/k8s/dashboards/homepage-metrics-dashboard-configmap.yaml
git commit -m "Configure Cloudflare Exporter and add performance comparison dashboard"
git push
```

### 2. Wait for Flux to Sync
- Flux will automatically:
  - Update the HelmRelease (will trigger exporter pod restart)
  - Update the ConfigMap (Grafana will pick it up automatically)

### 3. Verify Cloudflare Exporter
```bash
# Check if exporter pod is running
kubectl get pods -n cloudflare-exporter

# Check exporter logs
kubectl logs -n cloudflare-exporter -l app.kubernetes.io/name=cloudflare-exporter

# Test metrics endpoint
kubectl port-forward -n cloudflare-exporter svc/cloudflare-exporter 8080:8080
curl http://localhost:8080/metrics | grep cloudflare_zone
```

### 4. Verify in Prometheus
- Open Prometheus UI
- Query: `up{job="cloudflare-exporter"}` (should return `1`)
- Query: `cloudflare_zone_requests_total{zone="lucena.cloud"}`

### 5. Check Grafana Dashboard
- Open: `https://grafana.lucena.cloud/d/homepage-metrics`
- Scroll to "Cloudflare Performance (vs Origin)" section
- Verify panels show data (may take a few minutes after exporter starts)

## Expected Results

After Flux syncs and the exporter starts:

1. **Cloudflare Exporter Pod**: Should be running and healthy
2. **Prometheus**: Should be scraping `cloudflare-exporter` job
3. **Grafana Dashboard**: Cloudflare panels should show:
   - Total requests at Cloudflare edge
   - Origin requests (cache misses)
   - Calculated cache hit ratio
   - Bandwidth comparison

## Troubleshooting

### Exporter Pod Not Starting
```bash
# Check pod status
kubectl describe pod -n cloudflare-exporter -l app.kubernetes.io/name=cloudflare-exporter

# Check HelmRelease status
kubectl get helmrelease cloudflare-exporter -n cloudflare-exporter
kubectl describe helmrelease cloudflare-exporter -n cloudflare-exporter
```

### No Metrics in Prometheus
```bash
# Check ServiceMonitor
kubectl get servicemonitor cloudflare-exporter -n cloudflare-exporter

# Check Prometheus targets
# Open Prometheus UI → Status → Targets
# Look for cloudflare-exporter endpoint
```

### Dashboard Shows "No data"
- Wait 2-3 minutes after exporter starts (metrics need to be scraped)
- Verify exporter is running: `kubectl get pods -n cloudflare-exporter`
- Check metrics exist: Query `cloudflare_zone_requests_total` in Prometheus
- Verify zone name is correct: `lucena.cloud` (update queries if different)

## Files Modified

1. ✅ `flux/infrastructure/cloudflare-exporter/helmrelease.yaml` - Added CF_API_TOKEN
2. ✅ `flux/infrastructure/prometheus-operator/k8s/dashboards/homepage-metrics-dashboard-configmap.yaml` - Updated with new panels
3. ✅ Kubernetes secret `cloudflare-tunnel/cloudflare` - Updated with API token

## Related Documentation

- `PROMETHEUS_METRICS_COMPARISON.md` - How to verify improvements
- `CLOUDFLARE_DASHBOARD_SETUP.md` - Dashboard setup guide
- `CLOUDFLARE_METRICS_INTEGRATION.md` - Exporter integration details
- `QUICK_FIXES_IMPLEMENTATION.md` - Performance improvements to verify

---

**Status**: ✅ Setup Complete - Ready to commit and deploy  
**Last Updated**: 2025-01-XX
