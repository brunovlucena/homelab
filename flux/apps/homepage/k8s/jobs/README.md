# 📸 Grafana Screenshots Job

Kubernetes Job to capture Grafana dashboard screenshots and upload them to MinIO for blog posts.

## Quick Start

```bash
# From homepage directory
make capture-grafana-screenshots

# Or directly with kubectl
kubectl apply -f k8s/jobs/grafana-screenshots-job.yaml
kubectl logs -f job/grafana-screenshots -n homepage
```

## What It Does

1. **Syncs Credentials**: Fetches MinIO and Grafana credentials from Kubernetes secrets
2. **Captures Screenshots**: Uses Grafana rendering API to capture dashboard images
3. **Uploads to MinIO**: Pushes PNG files to `homepage-blog/images/graphs/` bucket

## Captured Dashboards

| Dashboard | Output File | Description |
|-----------|-------------|-------------|
| `demo-notifi` | `demo-notifi-lifecycle-events.png` | Demo-Notifi event metrics |
| `k6-knative-lambda` | `k6-load-testing.png` | K6 load testing results |
| `knative-lambda-metrics` | `knative-lambda-metrics.png` | Knative Lambda metrics |
| `8b7a...` | `kubernetes-networking-pods.png` | Pod network traffic distribution |
| `200a...` | `kubernetes-compute-node-pods.png` | Node resource distribution |

## Configuration

### Time Range

Edit the `capture.sh` script in the ConfigMap to adjust the time range:

```bash
TIME_FROM="now-6h"  # Start time
TIME_TO="now"       # End time
```

### Dashboard Selection

Add or remove dashboards in the `DASHBOARDS` array:

```bash
declare -A DASHBOARDS=(
  ["dashboard-uid"]="output-filename"
)
```

### Image Size

Adjust width and height:

```bash
WIDTH=1800
HEIGHT=1200
```

## Usage

### Run the Job

```bash
# Using Makefile (recommended)
make capture-grafana-screenshots

# Using kubectl directly
kubectl apply -f k8s/jobs/grafana-screenshots-job.yaml
```

### Check Status

```bash
# Using Makefile
make grafana-screenshots-status

# Using kubectl
kubectl get job grafana-screenshots -n homepage
kubectl get pods -l job-name=grafana-screenshots -n homepage
```

### View Logs

```bash
# Follow logs (recommended)
kubectl logs -f job/grafana-screenshots -n homepage

# View all containers
kubectl logs job/grafana-screenshots -n homepage --all-containers=true

# View specific container
kubectl logs job/grafana-screenshots -n homepage -c capture-screenshots
kubectl logs job/grafana-screenshots -n homepage -c upload-to-minio
```

### Clean Up

```bash
# Using Makefile
make grafana-screenshots-cleanup

# Using kubectl
kubectl delete job grafana-screenshots -n homepage
```

## Troubleshooting

### Job Fails to Start

```bash
# Check pod status
kubectl get pods -l job-name=grafana-screenshots -n homepage

# Describe pod for events
kubectl describe pod -l job-name=grafana-screenshots -n homepage
```

### Grafana Authentication Issues

```bash
# Verify Grafana secret exists
kubectl get secret grafana-admin -n monitoring

# Check Grafana password
kubectl get secret grafana-admin -n monitoring -o jsonpath='{.data.GF_SECURITY_ADMIN_PASSWORD}' | base64 -d
```

### MinIO Upload Issues

```bash
# Verify MinIO credentials
kubectl get secret minio-credentials -n minio

# Test MinIO connectivity
kubectl run -it --rm minio-test --image=minio/mc:latest --restart=Never -- \
  mc alias set test http://minio.minio.svc.cluster.local:9000 minioadmin <password>
```

### Screenshots Not Generated

1. **Check Grafana is accessible**:
   ```bash
   kubectl exec -it deployment/grafana -n monitoring -- \
     curl -s http://localhost:3000/api/health
   ```

2. **Verify dashboard UIDs**:
   ```bash
   kubectl exec -it deployment/grafana -n monitoring -- \
     curl -s -u admin:<password> http://localhost:3000/api/search
   ```

3. **Check render API**:
   ```bash
   # Test render endpoint
   curl -u admin:<password> \
     "http://grafana.monitoring.svc.cluster.local:3000/render/d/demo-notifi?width=800&height=600" \
     -o test.png
   ```

## Access Screenshots

### Via MinIO Console

1. Open MinIO console: `http://minio.minio.svc.cluster.local:9000`
2. Navigate to `homepage-blog` bucket
3. Browse `images/graphs/` directory

### Via CDN URL

Screenshots are publicly accessible at:
```
http://minio.minio.svc.cluster.local:9000/homepage-blog/images/graphs/<filename>
```

### In Blog Posts

Reference screenshots in Markdown:
```markdown
![Description](./graphs/<filename>?v=0.1.21)
```

## Advanced Usage

### Custom Time Range

Run job with custom time range by modifying the ConfigMap before applying:

```yaml
# Edit k8s/jobs/grafana-screenshots-job.yaml
data:
  capture.sh: |
    TIME_FROM="2025-12-17T10:00:00Z"
    TIME_TO="2025-12-17T16:00:00Z"
```

### Capture Specific Panel

Add panel-specific capture to the script:

```bash
# Capture single panel
render_url="${GRAFANA_URL}/render/d-solo/${uid}?panelId=2&width=1200&height=600"
```

### High Resolution Export

Increase DPI for print-quality screenshots:

```bash
WIDTH=3600   # Double resolution
HEIGHT=2400  # Double resolution
```

## Integration with Blog Workflow

1. **Capture screenshots**: `make capture-grafana-screenshots`
2. **Verify upload**: Check MinIO bucket
3. **Reference in blog**: Update markdown with image paths
4. **Deploy blog**: `make deploy`

## Related Jobs

- **`blog-graphs-generate-upload-job.yaml`**: Generate graphs from Python scripts
- **`blog-images-upload-job.yaml`**: Upload static images to MinIO
- **`three-scales-framework-upload-job.yaml`**: Upload three-scales-framework.png diagram to MinIO (see below)

### Three Scales Framework Image Upload

Uploads the `three-scales-framework.png` diagram used in the "understanding-vs-knowledge" blog post.

**Usage:**
```bash
# Automated upload script
cd k8s/jobs
./upload-three-scales-framework.sh

# Or manual:
# 1. Create ConfigMap
kubectl create configmap three-scales-framework-image \
  --from-file=three-scales-framework.png=../../storage/homepage-blog/images/graphs/three-scales-framework.png \
  --namespace=homepage --dry-run=client -o yaml | kubectl apply -f -

# 2. Apply job
kubectl apply -f three-scales-framework-upload-job.yaml

# 3. Check logs
kubectl logs -f job/three-scales-framework-upload -n homepage
```

**Image Location:**
- Source: `storage/homepage-blog/images/graphs/three-scales-framework.png`
- MinIO: `homepage-blog/images/graphs/three-scales-framework.png`
- Blog reference: `/storage/homepage-blog/images/graphs/three-scales-framework.png`

## Architecture

```
┌─────────────────┐
│  Init Container │  Sync MinIO + Grafana credentials
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Capture         │  curl → Grafana render API → PNG files
│ Container       │  Output: /output/*.png
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Upload          │  MinIO mc → homepage-blog/images/graphs/
│ Container       │  Set public read policy
└─────────────────┘
```

## Security

- **Secrets**: Automatically synced from minio and monitoring namespaces
- **RBAC**: Minimal permissions (read secrets, create in homepage namespace)
- **Network**: Internal cluster communication only
- **Cleanup**: TTL set to 300s (5 minutes) after completion

## Performance

- **Execution Time**: ~30-60 seconds per dashboard
- **Image Size**: ~200-500 KB per PNG (compressed)
- **Resource Usage**: 
  - CPU: ~100m (minimal)
  - Memory: ~256Mi (curl + mc client)

## Monitoring

View job metrics in Grafana:
- Dashboard: `Kubernetes / Jobs`
- Namespace: `homepage`
- Job Name: `grafana-screenshots`
