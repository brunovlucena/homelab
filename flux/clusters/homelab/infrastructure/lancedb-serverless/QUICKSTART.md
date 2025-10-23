# 🚀 Quick Start Guide

Get started with LanceDB Serverless in 5 minutes!

## Prerequisites

- Kubernetes cluster with Knative Serving installed
- MinIO running in the `minio` namespace
- Docker for building images
- `kubectl` configured

## 🔥 Quick Deploy (3 Steps)

### Step 1: Build and Push Image

```bash
make build-push
```

> **Note:** Update the `IMAGE_NAME` in the Makefile to match your container registry.

### Step 2: Create MinIO Bucket

```bash
make create-bucket
```

This creates a bucket named `lancedb` in MinIO and sets it to public access.

### Step 3: Deploy to Kubernetes

```bash
make deploy
```

## ✅ Verify Deployment

Check the status:

```bash
make status
```

You should see output like:

```
=== Knative Service ===
NAME                  URL                                              READY
lancedb-serverless   http://lancedb-serverless.lancedb-serverless...  True

=== Pods ===
NAME                                              READY   STATUS
lancedb-serverless-00001-deployment-xxx-xxx       2/2     Running
```

## 🧪 Test the API

Run the automated tests:

```bash
make test
```

Or manually test endpoints:

```bash
# Get the service URL
SERVICE_URL=$(kubectl get ksvc lancedb-serverless -n lancedb-serverless -o jsonpath='{.status.url}')

# Health check
curl $SERVICE_URL/health

# List tables
curl $SERVICE_URL/tables

# Create a table
curl -X POST $SERVICE_URL/tables \
  -H "Content-Type: application/json" \
  -d '{
    "name": "demo",
    "data": [
      {"id": 1, "vector": [0.1, 0.2, 0.3], "text": "Hello LanceDB"}
    ]
  }'

# Search
curl -X POST $SERVICE_URL/search \
  -H "Content-Type: application/json" \
  -d '{
    "table": "demo",
    "query_vector": [0.1, 0.2, 0.3],
    "limit": 5
  }'
```

## 📊 View Logs

```bash
make logs
```

## 🔌 Port Forward (Optional)

For local testing:

```bash
make port-forward
```

Then access at `http://localhost:8080`

## 🐍 Python Client Example

```python
import requests

SERVICE_URL = "http://lancedb-serverless.lancedb-serverless.svc.cluster.local"

# Create a table
response = requests.post(f"{SERVICE_URL}/tables", json={
    "name": "vectors",
    "data": [
        {"id": 1, "vector": [0.1, 0.2, 0.3], "metadata": "doc1"},
        {"id": 2, "vector": [0.4, 0.5, 0.6], "metadata": "doc2"}
    ]
})

print(response.json())

# Search
response = requests.post(f"{SERVICE_URL}/search", json={
    "table": "vectors",
    "query_vector": [0.1, 0.2, 0.3],
    "limit": 5
})

print(response.json())
```

See `example-usage.py` for more comprehensive examples.

## 🔧 Configuration

### Update MinIO Credentials

Edit `secret.yaml` and update with your MinIO credentials:

```yaml
stringData:
  AWS_ACCESS_KEY_ID: "your-access-key"
  AWS_SECRET_ACCESS_KEY: "your-secret-key"
```

### Change Auto-scaling Settings

Edit `service.yaml`:

```yaml
annotations:
  autoscaling.knative.dev/min-scale: "0"      # Min replicas
  autoscaling.knative.dev/max-scale: "10"     # Max replicas
  autoscaling.knative.dev/target: "100"       # Concurrency target
  autoscaling.knative.dev/scale-to-zero-pod-retention-period: "60s"
```

## 🗑️ Cleanup

To remove everything:

```bash
make delete
```

## 📚 Next Steps

- Read the full [README.md](./README.md) for detailed documentation
- Explore [example-usage.py](./example-usage.py) for code examples
- Check the [API documentation](#) by visiting the service root URL
- Set up monitoring with the included ServiceMonitor

## ⚠️ Troubleshooting

### Service not scaling up

```bash
kubectl logs -n knative-serving -l app=controller
```

### MinIO connection issues

```bash
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -n lancedb-serverless -- \
  curl http://minio-service.minio.svc.cluster.local:9000
```

### View detailed service status

```bash
make describe
```

## 🎉 You're All Set!

Your serverless LanceDB is now running and ready to handle vector search workloads with automatic scaling!

