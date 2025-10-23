# 🚀 Serverless LanceDB with MinIO and Knative

This implementation adapts the [LanceDB serverless example](https://github.com/lancedb/lancedb/blob/main/docs/src/examples/serverless_lancedb_with_s3_and_lambda.md) to use **MinIO** for S3-compatible object storage and **Knative** for serverless compute.

## 📋 Overview

This serverless application provides a REST API for vector search and storage using:
- **LanceDB**: AI-native vector database
- **MinIO**: S3-compatible object storage
- **Knative**: Serverless platform with auto-scaling

### Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       ▼
┌─────────────────────┐
│ Knative Service     │
│ (lancedb-serverless)│
└──────┬──────────────┘
       │
       ├──────────┐
       ▼          ▼
┌──────────┐  ┌──────────┐
│  MinIO   │  │ LanceDB  │
│ Storage  │  │ Engine   │
└──────────┘  └──────────┘
```

## ✨ Features

- **Auto-scaling**: Scales to zero when idle, saves resources
- **S3-Compatible**: Uses MinIO for persistent vector storage
- **REST API**: Simple HTTP interface for vector operations
- **Vector Search**: Hybrid search with filters and reranking
- **Observability**: Prometheus metrics and structured logging

## 🔧 Prerequisites

Ensure these are deployed in your cluster:
- MinIO (`minio` namespace)
- Knative Serving
- Prometheus Operator (for metrics)

## 📦 Deployment

### 1. Build and Push the Image

```bash
make build-push
```

### 2. Create MinIO Bucket

Create a bucket named `lancedb` in MinIO:

```bash
# Port-forward to MinIO
kubectl port-forward -n minio svc/minio-service 9000:9000

# Using mc (MinIO Client)
mc alias set local http://localhost:9000 <MINIO_ROOT_USER> <MINIO_ROOT_PASSWORD>
mc mb local/lancedb
mc policy set public local/lancedb
```

### 3. Deploy to Kubernetes

```bash
kubectl apply -f namespace.yaml
kubectl apply -f secret.yaml
kubectl apply -f service.yaml
```

Or use Kustomize:

```bash
kubectl apply -k .
```

## 🔑 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `AWS_ENDPOINT` | MinIO endpoint URL | `http://minio-service.minio.svc.cluster.local:9000` |
| `AWS_DEFAULT_REGION` | AWS region (required by boto3) | `us-east-1` |
| `AWS_ACCESS_KEY_ID` | MinIO access key | From secret |
| `AWS_SECRET_ACCESS_KEY` | MinIO secret key | From secret |
| `LANCEDB_BUCKET` | S3 bucket name | `lancedb` |
| `LOG_LEVEL` | Logging level | `INFO` |

### Creating Secrets

```bash
kubectl create secret generic lancedb-serverless-secret \
  -n lancedb-serverless \
  --from-literal=AWS_ACCESS_KEY_ID=<your-minio-user> \
  --from-literal=AWS_SECRET_ACCESS_KEY=<your-minio-password>
```

## 🔌 API Usage

### Get Service URL

```bash
kubectl get ksvc lancedb-serverless -n lancedb-serverless
```

### Create a Table

```bash
curl -X POST http://<service-url>/tables \
  -H "Content-Type: application/json" \
  -d '{
    "name": "documents",
    "data": [
      {"id": 1, "vector": [0.1, 0.2, 0.3], "text": "Hello world"},
      {"id": 2, "vector": [0.4, 0.5, 0.6], "text": "Goodbye world"}
    ]
  }'
```

### Vector Search

```bash
curl -X POST http://<service-url>/search \
  -H "Content-Type: application/json" \
  -d '{
    "table": "documents",
    "query_vector": [0.1, 0.2, 0.3],
    "limit": 10
  }'
```

### List Tables

```bash
curl http://<service-url>/tables
```

### Health Check

```bash
curl http://<service-url>/health
```

## 📊 Monitoring

The service exposes Prometheus metrics at `/metrics`:

- `lancedb_requests_total` - Total number of requests
- `lancedb_request_duration_seconds` - Request duration histogram
- `lancedb_errors_total` - Total number of errors
- `lancedb_tables_total` - Number of tables
- `lancedb_search_operations_total` - Number of search operations

## 🔍 Troubleshooting

### Check Pod Status

```bash
kubectl get pods -n lancedb-serverless
kubectl logs -n lancedb-serverless -l serving.knative.dev/service=lancedb-serverless
```

### View Knative Service Status

```bash
kubectl get ksvc -n lancedb-serverless
kubectl describe ksvc lancedb-serverless -n lancedb-serverless
```

### Test MinIO Connectivity

```bash
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -n lancedb-serverless -- \
  curl http://minio-service.minio.svc.cluster.local:9000
```

### Common Issues

#### Service Not Scaling Up

Check Knative Serving logs:
```bash
kubectl logs -n knative-serving -l app=controller
```

#### MinIO Connection Failed

Verify MinIO is running:
```bash
kubectl get pods -n minio
```

Check network connectivity:
```bash
kubectl run -it --rm debug --image=nicolaka/netshoot --restart=Never -n lancedb-serverless -- \
  nc -zv minio-service.minio.svc.cluster.local 9000
```

#### Bucket Not Found

Create the bucket in MinIO:
```bash
mc mb local/lancedb
```

## 🎯 Example Use Cases

### Document Search

```python
import requests

url = "http://lancedb-serverless.lancedb-serverless.svc.cluster.local"

# Create a table with document embeddings
response = requests.post(f"{url}/tables", json={
    "name": "docs",
    "data": [
        {"id": 1, "vector": embedding1, "text": "Document 1", "category": "tech"},
        {"id": 2, "vector": embedding2, "text": "Document 2", "category": "science"}
    ]
})

# Search for similar documents
response = requests.post(f"{url}/search", json={
    "table": "docs",
    "query_vector": query_embedding,
    "limit": 5,
    "filter": "category = 'tech'"
})

results = response.json()
```

### Image Similarity Search

```python
# Store image embeddings
requests.post(f"{url}/tables", json={
    "name": "images",
    "data": [
        {"id": "img1", "vector": img_embedding1, "filename": "cat.jpg"},
        {"id": "img2", "vector": img_embedding2, "filename": "dog.jpg"}
    ]
})

# Find similar images
response = requests.post(f"{url}/search", json={
    "table": "images",
    "query_vector": query_img_embedding,
    "limit": 10
})
```

## 🔗 References

- [LanceDB Documentation](https://lancedb.com/docs/)
- [Knative Serving Documentation](https://knative.dev/docs/serving/)
- [MinIO Documentation](https://min.io/docs/)
- [Original LanceDB Serverless Example](https://github.com/lancedb/lancedb/blob/main/docs/src/examples/serverless_lancedb_with_s3_and_lambda.md)

## 📝 Notes

- The service scales to zero after 60 seconds of inactivity (configurable in `service.yaml`)
- Cold start time is typically 2-5 seconds
- For production use, consider:
  - Adding authentication/authorization
  - Implementing rate limiting
  - Using specific image tags instead of `latest`
  - Setting up proper backup strategies
  - Configuring resource limits based on workload

