# Cloudflare Metrics Integration with Prometheus

## Executive Summary

As your Senior SRE, I recommend **using the LabLabs Cloudflare Exporter** to bring Cloudflare HTTP traffic analytics into Prometheus. This is the most straightforward and reliable solution that exposes exactly the metrics you're seeing in the Cloudflare dashboard.

## Problem Statement

You want to track Cloudflare HTTP traffic metrics for `lucena.cloud` in Prometheus, specifically:
- Total HTTP requests
- Cached requests
- Uncached requests
- Bandwidth usage
- Traffic by country/region
- Request trends over time

## Options Analysis

### Option 1: LabLabs Cloudflare Exporter ✅ **RECOMMENDED**

**What it does:**
- Queries Cloudflare's REST and GraphQL APIs to fetch analytics data
- Exposes metrics in Prometheus format
- Provides per-zone metrics including HTTP requests, bandwidth, threats, and geographic distribution

**Pros:**
- ✅ Provides **exact metrics** from Cloudflare dashboard (HTTP requests, cached/uncached, bandwidth, by country)
- ✅ Well-maintained, production-ready exporter
- ✅ Supports API tokens (secure, least privilege)
- ✅ Works with free Cloudflare plan
- ✅ Helm chart available for easy deployment
- ✅ Can filter by specific zones
- ✅ Provides worker metrics if you use Cloudflare Workers

**Cons:**
- ⚠️ Requires Cloudflare API token with `Zone/Analytics:Read` permission
- ⚠️ Rate-limited by Cloudflare API (scrape interval should be ≥60s)

**Metrics Provided:**
- `cloudflare_zone_requests_total` - Total requests per zone
- `cloudflare_zone_bandwidth_total` - Total bandwidth per zone (bytes)
- `cloudflare_zone_threats_total` - Threats blocked per zone
- `cloudflare_zone_requests_cached` - Cached requests (if available)
- `cloudflare_zone_colocation_*` - Metrics by colocation (country/region)
- Worker metrics (CPU time, duration, errors, requests)

### Option 2: Cloudflare Tunnel Metrics (Existing) ❌ **NOT SUITABLE**

**What it does:**
- The existing `cloudflare-tunnel` ServiceMonitor scrapes metrics from `cloudflared` pods
- These metrics expose tunnel connection health and performance

**Pros:**
- ✅ Already deployed and working
- ✅ No additional infrastructure needed

**Cons:**
- ❌ **Does NOT provide HTTP traffic analytics** (requests, cached/uncached, bandwidth)
- ❌ Only shows tunnel connection metrics (connections, latency, errors)
- ❌ Cannot track the metrics shown in Cloudflare dashboard

**Conclusion:** Keep the existing tunnel monitoring for infrastructure health, but use Option 1 for business/application metrics.

### Option 3: Kubernetes Metrics (Not Applicable)

Pulling metrics from Kubernetes would only show:
- Pod metrics (CPU, memory, requests)
- Service metrics (internal traffic)
- Ingress metrics (if using Ingress controller)

This would **NOT** include:
- Cloudflare edge analytics
- Cached vs uncached requests
- Geographic distribution
- External traffic before it hits your cluster

## Recommendation: LabLabs Cloudflare Exporter

### Why This Approach?

1. **Matches Your Requirements:** Provides exactly the metrics shown in your Cloudflare dashboard
2. **Production-Ready:** Battle-tested exporter used by many organizations
3. **Maintainable:** Well-documented, actively maintained
4. **Secure:** Uses API tokens (not global API keys)
5. **GitOps-Friendly:** Helm chart integrates cleanly with your Flux setup

### Implementation Steps

1. **Create Cloudflare API Token:**
   - Go to: https://dash.cloudflare.com/profile/api-tokens
   - Create token with `Zone/Analytics:Read` permission
   - Store in Kubernetes secret (can reuse existing `cloudflare` secret)

2. **Deploy Cloudflare Exporter:**
   - Add HelmRepository for LabLabs charts
   - Create HelmRelease with proper configuration
   - Expose metrics endpoint via Service

3. **Configure Prometheus Scraping:**
   - Create ServiceMonitor for the exporter
   - Metrics will automatically appear in Prometheus

4. **Visualize in Grafana:**
   - Use pre-built dashboard (ID: 13133) or create custom dashboards
   - Query metrics using PromQL

### Required Cloudflare API Token Permissions

Create an API token with:
- **Permission:** `Zone/Analytics:Read`
- **Zone Resources:** Include - Specific zone (`lucena.cloud`)
- **Account Resources:** (Not required for zone-level metrics)

### Architecture

```
Cloudflare API (REST/GraphQL)
    ↓
Cloudflare Exporter (Pod)
    ↓ (scrapes every 60s)
Prometheus (ServiceMonitor)
    ↓
Grafana (Dashboards)
```

### Scrape Interval Considerations

- **Recommended:** 60-120 seconds
- **Why:** Cloudflare API rate limits + analytics data granularity
- **Note:** Analytics data is aggregated, so frequent scraping doesn't provide additional resolution

## Implementation Files

The implementation includes:
- `cloudflare-exporter/namespace.yaml` - Namespace for the exporter
- `cloudflare-exporter/helmrelease.yaml` - Flux HelmRelease configuration
- `cloudflare-exporter/kustomization.yaml` - Kustomize resources
- `prometheus-operator/k8s/servicemonitors/cloudflare-exporter-servicemonitor.yaml` - ServiceMonitor for Prometheus scraping

## Next Steps

1. **Create Cloudflare API Token:**
   - Go to: https://dash.cloudflare.com/profile/api-tokens
   - Click "Create Token"
   - Use template or manually set:
     - **Permissions:** `Zone` > `Analytics` > `Read`
     - **Zone Resources:** Include > Specific zone > `lucena.cloud`
   - Copy the generated token

2. **Configure the API Token:**
   
   **Option A: Update HelmRelease directly (Quick)**
   - Edit `cloudflare-exporter/helmrelease.yaml`
   - Set `env[0].value` to your API token (base64 encode if needed)
   
   **Option B: Use existing secret (Recommended)**
   - Ensure `cloudflare-tunnel/cloudflare` secret exists with `cloudflare-api-token` key
   - If using a different secret, update the HelmRelease values accordingly
   - You can create/update the secret:
     ```bash
     kubectl create secret generic cloudflare -n cloudflare-tunnel \
       --from-literal=cloudflare-api-token='YOUR_API_TOKEN' \
       --dry-run=client -o yaml | kubectl apply -f -
     ```

3. **Deploy via Flux (GitOps):**
   - Commit and push changes
   - Flux will automatically deploy the exporter
   - Monitor with: `kubectl get helmrelease cloudflare-exporter -n cloudflare-exporter`

4. **Verify Deployment:**
   ```bash
   # Check pod is running
   kubectl get pods -n cloudflare-exporter
   
   # Check metrics endpoint
   kubectl port-forward -n cloudflare-exporter svc/cloudflare-exporter 8080:8080
   curl http://localhost:8080/metrics | grep cloudflare
   ```

5. **Verify in Prometheus:**
   - Open Prometheus UI
   - Query: `up{job="cloudflare-exporter"}`
   - Should show `1` if scraping successfully
   - Query: `cloudflare_zone_requests_total` to see HTTP request metrics

6. **Visualize in Grafana:**
   - Import dashboard ID: `13133` (Cloudflare Zone Analytics)
   - Or create custom dashboards using the metrics
   - Example queries:
     - Total requests: `sum(cloudflare_zone_requests_total)`
     - Requests by zone: `cloudflare_zone_requests_total{zone="lucena.cloud"}`

## References

- [LabLabs Cloudflare Exporter GitHub](https://github.com/lablabs/cloudflare-exporter)
- [Grafana Dashboard (ID: 13133)](https://grafana.com/grafana/dashboards/13133)
- [Cloudflare API Tokens Documentation](https://developers.cloudflare.com/fundamentals/api/get-started/create-token/)
