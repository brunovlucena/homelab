# Cloudflare Exporter - Next Steps & Verification

## ✅ Commit Completed

All changes have been committed:
```
3e9bf96c feat: Add Cloudflare Exporter for Prometheus metrics integration
```

## 🔍 Step 1: Verify Cloudflare API Token Secret

The exporter requires a Cloudflare API token in the `cloudflare-tunnel` namespace.

### Check if secret exists:
```bash
kubectl get secret cloudflare -n cloudflare-tunnel
```

### Check if it has the required key:
```bash
kubectl get secret cloudflare -n cloudflare-tunnel -o jsonpath='{.data.cloudflare-api-token}' && echo
```

If the secret doesn't exist or doesn't have `cloudflare-api-token` key, you need to create/update it:

### Create/Update Secret:

**Option A: If you have an existing API token:**
```bash
kubectl create secret generic cloudflare -n cloudflare-tunnel \
  --from-literal=cloudflare-api-token='YOUR_API_TOKEN' \
  --dry-run=client -o yaml | kubectl apply -f -
```

**Option B: Create a new API token:**
1. Go to: https://dash.cloudflare.com/profile/api-tokens
2. Click "Create Token"
3. Use "Edit Cloudflare Workers" template or create custom:
   - **Permissions:**
     - `Zone` > `Analytics` > `Read`
   - **Zone Resources:**
     - Include > Specific zone > `lucena.cloud`
4. Copy the token and run:
```bash
kubectl create secret generic cloudflare -n cloudflare-tunnel \
  --from-literal=cloudflare-api-token='YOUR_NEW_TOKEN' \
  --dry-run=client -o yaml | kubectl apply -f -
```

## 📝 Step 2: Review HelmRelease Configuration

**⚠️ Important Note:** The LabLabs Cloudflare Exporter Helm chart expects environment variables in array format. The current configuration uses `valuesFrom` with `targetPath: cfApiToken`, which may not work if the chart doesn't support this path.

### If deployment fails, you may need to use env array format:

Edit `flux/infrastructure/cloudflare-exporter/helmrelease.yaml`:

```yaml
values:
  env:
    - name: CF_API_TOKEN
      valueFrom:
        secretKeyRef:
          name: cloudflare
          namespace: cloudflare-tunnel
          key: cloudflare-api-token
    - name: SCRAPE_INTERVAL
      value: "60"
```

Or if you prefer to set the token directly (less secure, but works):
```yaml
valuesFrom:
  - kind: Secret
    name: cloudflare
    namespace: cloudflare-tunnel
    valuesKey: cloudflare-api-token
    targetPath: env[0].value
    optional: false
values:
  env:
    - name: CF_API_TOKEN
      value: ""  # Injected via valuesFrom
    - name: SCRAPE_INTERVAL
      value: "60"
```

## 🚀 Step 3: Monitor Deployment

### Check HelmRepository:
```bash
kubectl get helmrepository lablabs-cloudflare-exporter -n flux-system
```

### Check HelmRelease status:
```bash
kubectl get helmrelease cloudflare-exporter -n cloudflare-exporter
kubectl describe helmrelease cloudflare-exporter -n cloudflare-exporter
```

### Check if namespace exists:
```bash
kubectl get namespace cloudflare-exporter
```

### Watch for pod creation:
```bash
kubectl get pods -n cloudflare-exporter -w
```

### Check pod logs (once running):
```bash
kubectl logs -n cloudflare-exporter -l app.kubernetes.io/name=cloudflare-exporter --tail=50
```

## 🔍 Step 4: Verify Metrics Endpoint

### Port-forward to test metrics:
```bash
kubectl port-forward -n cloudflare-exporter svc/cloudflare-exporter 8080:8080
```

### In another terminal, test metrics:
```bash
curl http://localhost:8080/metrics | grep cloudflare
```

Expected output should include metrics like:
- `cloudflare_zone_requests_total`
- `cloudflare_zone_bandwidth_total`
- `cloudflare_zone_threats_total`

## 📊 Step 5: Verify Prometheus Scraping

### Check ServiceMonitor:
```bash
kubectl get servicemonitor cloudflare-exporter -n cloudflare-exporter
```

### Check Prometheus targets:
1. Port-forward to Prometheus:
```bash
kubectl port-forward -n prometheus svc/kube-prometheus-stack-prometheus 9090:9090
```

2. Open browser: http://localhost:9090
3. Go to: Status > Targets
4. Look for `cloudflare-exporter` job
5. Status should be "UP"

### Query metrics in Prometheus:
```promql
# Check if exporter is up
up{job="cloudflare-exporter"}

# Check for Cloudflare metrics
cloudflare_zone_requests_total

# Filter by zone
cloudflare_zone_requests_total{zone="lucena.cloud"}
```

## 📈 Step 6: Import Grafana Dashboard

1. Open Grafana
2. Go to: Dashboards > Import
3. Enter dashboard ID: `13133`
4. Select Prometheus datasource
5. Import

Or use the dashboard URL:
https://grafana.com/grafana/dashboards/13133

## 🐛 Troubleshooting

### Pod not starting:
- Check logs: `kubectl logs -n cloudflare-exporter -l app.kubernetes.io/name=cloudflare-exporter`
- Check secret exists: `kubectl get secret cloudflare -n cloudflare-tunnel`
- Verify API token permissions

### No metrics in Prometheus:
- Check ServiceMonitor: `kubectl describe servicemonitor cloudflare-exporter -n cloudflare-exporter`
- Check Prometheus targets page
- Verify service exists: `kubectl get svc -n cloudflare-exporter`
- Check pod is running: `kubectl get pods -n cloudflare-exporter`

### Authentication errors:
- Verify API token is valid
- Check token has `Zone/Analytics:Read` permission
- Check token is for the correct zone (`lucena.cloud`)

### HelmRelease not deploying:
- Check HelmRepository is ready: `kubectl get helmrepository -n flux-system`
- Check HelmRelease events: `kubectl describe helmrelease cloudflare-exporter -n cloudflare-exporter`
- Check Flux logs: `kubectl logs -n flux-system -l app=helm-controller`
