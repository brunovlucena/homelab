# LanceDB - AI-Native Multimodal Lakehouse

LanceDB is an AI-Native Multimodal Lakehouse designed for modern AI applications. It provides vector storage, hybrid search, and multimodal data support for AI workloads.

## 🚀 Features

- **Vector Storage**: Efficient storage and retrieval of embeddings
- **Hybrid Search**: Combine vector similarity with filters and reranking
- **Multimodal Data**: Support for text, images, audio, and video
- **Scalable**: From prototype to petabyte-scale production
- **AI Training**: Optimized dataloading for PyTorch and JAX

## 🔧 Configuration

### Service Endpoints

- **HTTP API**: `http://lancedb.lancedb.svc.cluster.local:8000`
- **Health Check**: `http://lancedb.lancedb.svc.cluster.local:8000/health`
- **Metrics**: `http://lancedb.lancedb.svc.cluster.local:8000/metrics`

### Resources

- **CPU Request**: 200m
- **CPU Limit**: 1000m
- **Memory Request**: 512Mi
- **Memory Limit**: 2Gi
- **Storage**: 50Gi

### Data Persistence

LanceDB uses a PersistentVolumeClaim (PVC) to store data at `/data/lancedb` within the container.

## 📊 Monitoring

LanceDB is monitored via Prometheus using a ServiceMonitor. Metrics are scraped from the `/metrics` endpoint every 30 seconds.

## 🔗 Usage

Connect to LanceDB from your applications using the Python SDK:

\`\`\`python
import lancedb

# Connect to LanceDB running in the cluster
db = lancedb.connect("http://lancedb.lancedb.svc.cluster.local:8000")

# Create a table
table = db.create_table("my_table", data=your_data)

# Perform hybrid search
results = (table.search("flying cars", query_type="hybrid")
    .where("date > '2025-01-01'")
    .limit(10)
    .to_pandas())
\`\`\`

## 📚 Documentation

- [LanceDB Documentation](https://lancedb.com/docs/)
- [LanceDB GitHub](https://github.com/lancedb/lancedb)
- [Quick Start Guide](https://lancedb.com/docs/quickstart/)

## 🔍 Troubleshooting

### Check Pod Status
\`\`\`bash
kubectl get pods -n lancedb
kubectl logs -n lancedb deployment/lancedb
\`\`\`

### Verify Service
\`\`\`bash
kubectl get svc -n lancedb
kubectl describe svc lancedb -n lancedb
\`\`\`

### Test Health Endpoint
\`\`\`bash
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- \
  curl http://lancedb.lancedb.svc.cluster.local:8000/health
\`\`\`

### Check PVC
\`\`\`bash
kubectl get pvc -n lancedb
kubectl describe pvc lancedb-pvc -n lancedb
\`\`\`

## ⚠️ Note

This deployment uses the `lancedb/lancedb:latest` Docker image. If this image is not available or you need a different configuration, you may need to:

1. Build your own LanceDB server image
2. Use LanceDB Cloud instead
3. Deploy LanceDB as a sidecar in your application pods

For production use, consider:
- Using specific image tags instead of `latest`
- Implementing backup strategies for the PVC
- Configuring resource limits based on your workload
- Setting up authentication and access controls

