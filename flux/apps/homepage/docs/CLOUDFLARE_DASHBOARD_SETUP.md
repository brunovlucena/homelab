# Cloudflare Performance Dashboard Setup

## ✅ What Was Done

1. **Added Cloudflare Performance Section to Dashboard**
   - New section "Cloudflare Performance (vs Origin)" added to Homepage Metrics dashboard
   - 5 new panels comparing Cloudflare edge metrics with origin server metrics

2. **Created Comparison Documentation**
   - `PROMETHEUS_METRICS_COMPARISON.md` - Comprehensive guide on how to verify Cloudflare improvements using Prometheus

## 📊 New Dashboard Panels

The following panels were added to the Homepage Metrics dashboard:

### 1. Request Rate: Cloudflare vs Origin
- **Query**: Compares `cloudflare_zone_requests_total` with `http_requests_total{job="homepage-api"}`
- **Purpose**: Shows total requests at Cloudflare edge vs requests reaching origin
- **Insight**: Lower origin requests = better cache hit ratio

### 2. Estimated Cache Hit Ratio
- **Query**: `(1 - Origin Requests / Cloudflare Total) * 100`
- **Purpose**: Calculated cache hit percentage
- **Target**: >85% (green threshold)
- **Visual**: Stat panel with color coding (red <70%, yellow 70-85%, green >85%)

### 3. Origin Load Reduction
- **Query**: Percentage of requests served from cache
- **Purpose**: Shows how much load is being offloaded from origin
- **Insight**: Higher percentage = more effective caching

### 4. Cache Hit vs Miss Breakdown
- **Query**: Stacked view of cached requests (estimated) vs cache misses
- **Purpose**: Visual breakdown of cache effectiveness
- **Visual**: Stacked area chart

### 5. Bandwidth: Cloudflare vs Origin
- **Query**: Compares `cloudflare_zone_bandwidth_total` with `http_response_size_bytes_sum`
- **Purpose**: Shows bandwidth usage at edge vs origin
- **Insight**: Lower origin bandwidth = more cache hits

## 🔧 Next Steps

### 1. Update ConfigMap (Required)

The dashboard JSON file has been updated, but the ConfigMap needs to be regenerated:

```bash
cd /Users/brunolucena/workspace/bruno/repos/homelab/flux/infrastructure/prometheus-operator/k8s/dashboards

# Regenerate ConfigMap from JSON
cat homepage-metrics-dashboard-configmap.yaml | head -12 > temp.yaml
cat homepage-metrics-dashboard.json | sed 's/^/    /' >> temp.yaml
mv temp.yaml homepage-metrics-dashboard-configmap.yaml
```

Or manually update the ConfigMap by:
1. Opening `homepage-metrics-dashboard-configmap.yaml`
2. Replacing the JSON content in the `data.homepage-metrics-dashboard.json` field with the updated JSON from `homepage-metrics-dashboard.json`

### 2. Configure Cloudflare Exporter (If Not Done)

The dashboard panels will show "No data" until the Cloudflare Exporter is configured:

1. **Create Cloudflare API Token:**
   - Go to: https://dash.cloudflare.com/profile/api-tokens
   - Create token with `Zone/Analytics:Read` permission for `lucena.cloud`

2. **Create Kubernetes Secret:**
   ```bash
   kubectl create secret generic cloudflare -n cloudflare-tunnel \
     --from-literal=cloudflare-api-token='YOUR_API_TOKEN' \
     --dry-run=client -o yaml | kubectl apply -f -
   ```

3. **Verify Exporter is Running:**
   ```bash
   kubectl get pods -n cloudflare-exporter
   kubectl get servicemonitor cloudflare-exporter -n cloudflare-exporter
   ```

4. **Check Metrics in Prometheus:**
   ```promql
   up{job="cloudflare-exporter"}
   cloudflare_zone_requests_total{zone="lucena.cloud"}
   ```

### 3. Deploy Changes

After updating the ConfigMap:

```bash
# Commit and push changes
git add .
git commit -m "Add Cloudflare performance comparison panels to homepage dashboard"
git push

# Flux will automatically sync the ConfigMap
# Verify in Grafana after a few minutes
```

### 4. Verify Dashboard

1. Open Grafana: `https://grafana.lucena.cloud/d/homepage-metrics`
2. Scroll to "Cloudflare Performance (vs Origin)" section
3. Verify panels show data (if exporter is configured)
4. If panels show "No data":
   - Check Cloudflare Exporter pod status
   - Verify ServiceMonitor is configured
   - Check Prometheus targets
   - Verify API token has correct permissions

## 📈 Expected Results

After implementing Cloudflare improvements from `QUICK_FIXES_IMPLEMENTATION.md`:

- **Cache Hit Ratio**: Should increase from ~70% to >85%
- **Origin Request Rate**: Should decrease by 30-40%
- **Origin Load Reduction**: Should show 70-85% of requests served from cache
- **Bandwidth**: Origin bandwidth should be significantly lower than Cloudflare total

## 🔍 Troubleshooting

### Panels Show "No data"

1. **Check Cloudflare Exporter:**
   ```bash
   kubectl get pods -n cloudflare-exporter
   kubectl logs -n cloudflare-exporter deployment/cloudflare-exporter
   ```

2. **Check ServiceMonitor:**
   ```bash
   kubectl get servicemonitor cloudflare-exporter -n cloudflare-exporter -o yaml
   ```

3. **Check Prometheus Targets:**
   - Open Prometheus UI
   - Go to Status → Targets
   - Look for `cloudflare-exporter` endpoint
   - Should show "UP" status

4. **Verify Metrics Exist:**
   ```promql
   # In Prometheus query interface
   up{job="cloudflare-exporter"}
   cloudflare_zone_requests_total
   ```

### Incorrect Zone Name

If metrics don't show up, verify the zone name in queries:
- Current zone: `lucena.cloud`
- To find your zone ID: Cloudflare Dashboard → Domain → Overview → Zone ID (right sidebar)

Update queries in dashboard if zone name is different.

### API Token Issues

If exporter logs show authentication errors:
1. Verify token has `Zone/Analytics:Read` permission
2. Verify token is for correct zone
3. Check secret exists: `kubectl get secret cloudflare -n cloudflare-tunnel`
4. Verify secret key name matches HelmRelease configuration

## 📚 Related Documentation

- `PROMETHEUS_METRICS_COMPARISON.md` - Detailed comparison guide
- `CLOUDFLARE_METRICS_INTEGRATION.md` - Cloudflare Exporter setup
- `QUICK_FIXES_IMPLEMENTATION.md` - Performance improvements to verify
- `HOMEPAGE_DASHBOARDS.md` - Dashboard documentation

---

**Last Updated**: 2025-01-XX  
**Status**: Dashboard panels added, ConfigMap update pending
