# ⚡ LambdaFunctions Versions Dashboard - Implementation Summary

**Date:** December 10, 2025  
**Status:** ✅ Complete

---

## 📋 Overview

A comprehensive dashboard and scanning system for tracking LambdaFunction image versions and identifying outdated images. The system uses `agent-devsecops` to scan LambdaFunctions and expose Prometheus metrics, which are then visualized in a Grafana dashboard.

---

## 🎯 Features

### 1. **Image Scanning** (`agent-devsecops`)
- Scans all LambdaFunction resources in the cluster
- Extracts image URIs from `.status.buildStatus.imageURI` or `.spec.source.image`
- Parses semantic versions from image tags
- Compares versions against minimum expected versions
- Identifies outdated images

### 2. **Prometheus Metrics**
The following metrics are exposed by `agent-devsecops`:

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `devsecops_lambdafunction_image_info` | Gauge | LambdaFunction image version information | `function`, `namespace`, `image_uri`, `tag`, `version`, `registry` |
| `devsecops_lambdafunction_outdated_total` | Gauge | Number of LambdaFunctions with outdated images | `namespace`, `registry` |
| `devsecops_lambdafunction_scans_total` | Counter | Total number of LambdaFunction scans performed | `namespace`, `status` |
| `devsecops_lambdafunction_scan_errors_total` | Counter | Total number of scan errors | `namespace`, `error_type` |

### 3. **Grafana Dashboard**
- **Total LambdaFunctions**: Count of all scanned functions
- **Outdated Images**: Count of functions with outdated images (highlighted in red)
- **Up-to-date Images**: Count of functions with current images
- **Functions Ready**: Count from `knative_lambda_operator_lambdafunctions_total`
- **Version Details Table**: Shows all functions with their image URIs, tags, versions, and status
- **Function Activity**: Time series showing invocations and duration from `knative-lambda-operator` metrics

---

## 📁 Files Created/Modified

### New Files

1. **`flux/ai/agent-devsecops/src/scanner/image_scanner.py`**
   - Image scanning logic
   - Version extraction and comparison
   - Outdated detection

2. **`flux/ai/agent-devsecops/src/scanner/metrics_exporter.py`**
   - Prometheus metrics exporter
   - Metrics server on port 9090 (configurable via `METRICS_PORT`)

3. **`flux/infrastructure/prometheus-operator/k8s/dashboards/lambdafunctions-versions-dashboard.json`**
   - Grafana dashboard JSON definition

4. **`flux/infrastructure/prometheus-operator/k8s/dashboards/lambdafunctions-versions-dashboard-configmap.yaml`**
   - Kubernetes ConfigMap for automatic Grafana dashboard discovery

### Modified Files

1. **`flux/ai/agent-devsecops/src/scanner/handler.py`**
   - Added `handle_scan_lambdafunctions()` method
   - Integrated image scanner and metrics exporter
   - Added event handler for `io.homelab.scan.lambdafunctions`

2. **`flux/ai/agent-devsecops/src/scanner/main.py`**
   - Added Prometheus metrics server startup
   - Added `/scan/lambdafunctions` API endpoint

3. **`flux/ai/agent-devsecops/src/requirements.txt`**
   - Added `prometheus-client>=0.19.0`
   - Added `packaging>=23.0` (for version comparison)

---

## 🚀 Usage

### 1. **Trigger a Scan**

#### Via CloudEvent
```bash
# Send a CloudEvent to trigger a scan
curl -X POST http://agent-devsecops:8080/ \
  -H "Content-Type: application/json" \
  -H "Ce-Type: io.homelab.scan.lambdafunctions" \
  -H "Ce-Source: manual-trigger" \
  -H "Ce-Id: scan-$(date +%s)" \
  -d '{"namespace": "knative-lambda"}'  # Optional: filter by namespace
```

#### Via API Endpoint
```bash
# Direct API call
curl -X POST http://agent-devsecops:8080/scan/lambdafunctions \
  -H "Content-Type: application/json" \
  -d '{"namespace": "knative-lambda"}'  # Optional
```

### 2. **View Metrics**

Metrics are exposed on port 9090 (default):
```bash
curl http://agent-devsecops:9090/metrics
```

### 3. **Access Dashboard**

1. Deploy the ConfigMap:
   ```bash
   kubectl apply -f flux/infrastructure/prometheus-operator/k8s/dashboards/lambdafunctions-versions-dashboard-configmap.yaml
   ```

2. Access Grafana and navigate to the **"⚡ LambdaFunctions Versions - QA Dashboard"**

---

## ⚙️ Configuration

### Minimum Version Thresholds

Edit `image_scanner.py` to adjust minimum expected versions:

```python
MIN_VERSIONS = {
    "ghcr.io/brunovlucena": "1.0.0",
    "localhost:5001": "0.1.0",
    "knative-lambdas": "0.1.0",
}
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `METRICS_PORT` | `9090` | Port for Prometheus metrics server |
| `PORT` | `8080` | Port for main Flask application |
| `LOG_LEVEL` | `INFO` | Logging verbosity |

---

## 🔍 How It Works

1. **Scan Trigger**: A CloudEvent or API call triggers `handle_scan_lambdafunctions()`
2. **Kubernetes Query**: The handler queries all LambdaFunction CRDs using the Kubernetes API
3. **Image Analysis**: For each LambdaFunction:
   - Extracts image URI from status or spec
   - Parses tag and version from image URI
   - Compares version against minimum expected version
   - Determines if image is outdated
4. **Metrics Export**: Results are exported as Prometheus metrics
5. **Dashboard Visualization**: Grafana queries Prometheus and displays the data

---

## 📊 Dashboard Panels

### Overview Stats
- **Total LambdaFunctions**: Total count of scanned functions
- **Outdated Images**: Functions with outdated images (red if > 0)
- **Up-to-date Images**: Functions with current images
- **Functions Ready**: Count from knative-lambda-operator metrics

### Version Details Table
- Function name
- Namespace
- Image URI
- Tag
- Version (highlighted in red if outdated)
- Registry
- Status (✅ Up-to-date / ⚠️ Outdated / ❌ Unknown / ⏳ No Image)

### Activity Metrics
- **Function Invocations (Rate)**: Success and error rates
- **Function Duration (P95)**: P95 latency percentiles

---

## 🔄 Integration with knative-lambda-operator Metrics

The dashboard combines metrics from two sources:

1. **agent-devsecops metrics** (`devsecops_lambdafunction_*`):
   - Image version information
   - Outdated image detection

2. **knative-lambda-operator metrics** (`knative_lambda_*`):
   - Function invocations (`knative_lambda_function_invocations_total`)
   - Function duration (`knative_lambda_function_duration_seconds`)
   - Function status (`knative_lambda_operator_lambdafunctions_total`)

---

## 🛠️ Troubleshooting

### No Metrics Appearing

1. **Check if scan was triggered**:
   ```bash
   kubectl logs -n agent-devsecops deployment/agent-devsecops | grep "LambdaFunction scan"
   ```

2. **Verify metrics endpoint**:
   ```bash
   kubectl port-forward -n agent-devsecops deployment/agent-devsecops 9090:9090
   curl http://localhost:9090/metrics | grep devsecops_lambdafunction
   ```

3. **Check Prometheus scraping**:
   - Ensure `agent-devsecops` has a ServiceMonitor or PodMonitor
   - Verify Prometheus is scraping the metrics endpoint

### Dashboard Not Showing Data

1. **Verify ConfigMap is deployed**:
   ```bash
   kubectl get configmap -n prometheus lambdafunctions-versions-dashboard
   ```

2. **Check Grafana dashboard discovery**:
   - Ensure Grafana is configured to discover dashboards from ConfigMaps
   - Check Grafana logs for dashboard loading errors

3. **Verify Prometheus queries**:
   - Test queries directly in Grafana Explore
   - Check if metrics exist: `devsecops_lambdafunction_image_info`

---

## 📝 Next Steps

1. **Automated Scanning**: Set up a CronJob or scheduled CloudEvent to trigger scans periodically
2. **Alerting**: Create Prometheus alerts for outdated images
3. **Version Policies**: Add support for per-namespace or per-function version policies
4. **Image Registry Integration**: Query registry APIs to check for newer available versions
5. **SBOM Integration**: Correlate image versions with SBOM data for vulnerability tracking

---

## 🎉 Summary

✅ Image scanning functionality added to `agent-devsecops`  
✅ Prometheus metrics exporter implemented  
✅ LambdaFunctions versions dashboard created  
✅ ConfigMap for automatic Grafana discovery created  
✅ API endpoint for manual scan triggering added  
✅ Integration with knative-lambda-operator metrics  

The system is ready to track LambdaFunction versions and highlight outdated images in red on the Grafana dashboard!
