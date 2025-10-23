# 🏗️ Architecture Documentation

## System Overview

The LanceDB Serverless system provides a scalable, cost-efficient vector database solution built on Kubernetes using modern cloud-native technologies.

```
┌──────────────────────────────────────────────────────────────────┐
│                         External Client                          │
│                    (HTTP/REST API Requests)                      │
└────────────────────────────┬─────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Knative Serving                             │
│  ┌────────────────────────────────────────────────────────┐     │
│  │  Knative Service: lancedb-serverless                   │     │
│  │  - Auto-scaling (0-10 replicas)                        │     │
│  │  - Scale-to-zero after 60s                             │     │
│  │  - Concurrency: 10 requests per pod                    │     │
│  └────────────────────────────────────────────────────────┘     │
└────────────────────────────┬─────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Application Pod                               │
│  ┌──────────────────────────────────────────────────────┐       │
│  │  Flask Application (Python)                          │       │
│  │  ┌────────────────────────────────────────────┐      │       │
│  │  │  REST API Endpoints:                       │      │       │
│  │  │  - POST /tables     (create/update)        │      │       │
│  │  │  - GET  /tables     (list)                 │      │       │
│  │  │  - POST /search     (vector search)        │      │       │
│  │  │  - DELETE /tables/<name>                   │      │       │
│  │  │  - GET  /health     (health check)         │      │       │
│  │  │  - GET  /metrics    (Prometheus)           │      │       │
│  │  └────────────────────────────────────────────┘      │       │
│  │                                                        │       │
│  │  ┌────────────────────────────────────────────┐      │       │
│  │  │  LanceDB SDK                                │      │       │
│  │  │  - Vector storage & retrieval               │      │       │
│  │  │  - Hybrid search (vector + filters)         │      │       │
│  │  │  - S3-compatible storage backend            │      │       │
│  │  └────────────────────────────────────────────┘      │       │
│  └──────────────────────────────────────────────────────┘       │
│                             │                                     │
│                             │ S3 API                              │
│                             ▼                                     │
│  ┌──────────────────────────────────────────────────────┐       │
│  │  AWS SDK (boto3)                                     │       │
│  │  - S3-compatible client                              │       │
│  │  - MinIO endpoint configuration                      │       │
│  └──────────────────────────────────────────────────────┘       │
└────────────────────────────┬─────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                      MinIO Storage                               │
│  ┌──────────────────────────────────────────────────────┐       │
│  │  Bucket: lancedb                                     │       │
│  │  ┌────────────────────────────────────────────┐     │       │
│  │  │  LanceDB Tables:                           │     │       │
│  │  │  - table1/                                 │     │       │
│  │  │    ├── data.lance                          │     │       │
│  │  │    └── metadata.json                       │     │       │
│  │  │  - table2/                                 │     │       │
│  │  │    ├── data.lance                          │     │       │
│  │  │    └── metadata.json                       │     │       │
│  │  └────────────────────────────────────────────┘     │       │
│  │                                                       │       │
│  │  Service: minio-service.minio.svc.cluster.local:9000│       │
│  │  Persistent Volume: 50Gi                             │       │
│  └──────────────────────────────────────────────────────┘       │
└─────────────────────────────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Observability Stack                           │
│  ┌──────────────────────────────────────────────────────┐       │
│  │  Prometheus (Metrics)                                │       │
│  │  - Request count, duration, errors                   │       │
│  │  - Table counts, search operations                   │       │
│  │  - ServiceMonitor scrapes /metrics every 30s         │       │
│  └──────────────────────────────────────────────────────┘       │
│                                                                   │
│  ┌──────────────────────────────────────────────────────┐       │
│  │  Grafana (Visualization)                             │       │
│  │  - Dashboard for LanceDB metrics                     │       │
│  │  - Alerting for service health                       │       │
│  └──────────────────────────────────────────────────────┘       │
└─────────────────────────────────────────────────────────────────┘
```

## Component Details

### 1. Knative Service

**Role:** Serverless compute platform providing auto-scaling and scale-to-zero capabilities.

**Key Features:**
- **Auto-scaling**: Automatically scales pods from 0 to 10 based on request load
- **Scale-to-zero**: Reduces to zero pods after 60 seconds of inactivity
- **Cold start**: ~2-5 seconds to spin up a new pod
- **Concurrency control**: Each pod handles up to 10 concurrent requests
- **Traffic management**: Knative routes requests to healthy pods

**Configuration:**
```yaml
autoscaling.knative.dev/min-scale: "0"
autoscaling.knative.dev/max-scale: "10"
autoscaling.knative.dev/target: "100"
autoscaling.knative.dev/scale-to-zero-pod-retention-period: "60s"
```

### 2. Application Container

**Base Image:** `python:3.11-slim`

**Components:**
- **Flask**: Lightweight web framework for REST API
- **LanceDB SDK**: Vector database library
- **Prometheus Client**: Metrics exposition
- **boto3**: AWS S3-compatible client

**Resource Limits:**
```yaml
requests:
  cpu: 100m
  memory: 256Mi
limits:
  cpu: 1000m
  memory: 1Gi
```

### 3. LanceDB Engine

**Type:** Embedded vector database (runs in-process)

**Storage Format:**
- Uses Apache Arrow/Lance columnar format
- Optimized for vector similarity search
- Supports both vector and scalar data
- ACID transactions

**Capabilities:**
- Vector similarity search (ANN)
- Hybrid search (vector + filters)
- Reranking
- Multi-modal data support

### 4. MinIO Storage

**Role:** S3-compatible object storage backend

**Configuration:**
- **Endpoint**: `http://minio-service.minio.svc.cluster.local:9000`
- **Bucket**: `lancedb`
- **Region**: `us-east-1` (required by AWS SDK)

**Storage Structure:**
```
s3://lancedb/
├── table1/
│   ├── data.lance          # Vector and scalar data
│   ├── metadata.json       # Table metadata
│   └── _indices/           # Vector indices
├── table2/
│   └── ...
```

### 5. Observability

**Metrics Exposed:**
- `lancedb_requests_total{method, endpoint, status}` - Total requests
- `lancedb_request_duration_seconds{method, endpoint}` - Request latency
- `lancedb_errors_total{error_type}` - Error counts
- `lancedb_tables_total` - Number of tables
- `lancedb_search_operations_total{table}` - Search operations per table

**Monitoring:**
- Prometheus scrapes metrics every 30 seconds
- ServiceMonitor auto-discovers the service
- Grafana dashboards for visualization

## Data Flow

### Creating a Table

```
1. Client sends POST /tables with data
2. Flask receives request
3. LanceDB SDK processes data
4. Data written to MinIO via S3 API
5. MinIO persists to disk (PVC)
6. Response sent to client
7. Prometheus metrics updated
```

### Vector Search

```
1. Client sends POST /search with query vector
2. Flask receives request
3. LanceDB SDK loads table from MinIO
4. Vector similarity search performed in-memory
5. Results filtered (if filter expression provided)
6. Top-K results returned to client
7. Metrics updated (search_operations_total++)
```

### Auto-scaling Behavior

```
No traffic → Scale to 0 (after 60s)
  ↓
First request → Knative spins up pod (cold start: ~2-5s)
  ↓
Sustained load → Scale up to max 10 pods
  ↓
Load decreases → Pods terminate after idle period
```

## Security

### Authentication
- ⚠️ **Current**: No authentication (suitable for internal cluster use)
- **Recommended**: Add API key or JWT authentication for production

### Network Security
- Service runs within Kubernetes cluster
- MinIO credentials stored in Kubernetes Secret
- Network policies can be added to restrict traffic

### Data Security
- Data encrypted at rest (if MinIO encryption enabled)
- Data encrypted in transit (HTTPS via Knative if configured)
- MinIO access controlled via IAM policies

## Scalability

### Horizontal Scaling
- Knative auto-scales from 0 to 10 pods
- Each pod handles 10 concurrent requests
- **Total capacity**: 100 concurrent requests at max scale

### Vertical Scaling
- CPU/Memory limits can be increased in `service.yaml`
- Recommended for larger datasets or higher throughput

### Storage Scaling
- MinIO PVC can be expanded
- Multiple MinIO instances can be deployed (distributed mode)
- LanceDB tables are independent and can be distributed

## Performance Characteristics

### Latency
- **Cold start**: 2-5 seconds
- **Warm request**: < 100ms (simple operations)
- **Vector search**: 10-500ms (depends on table size and vector dimensions)

### Throughput
- **Per pod**: ~100 requests/second (simple operations)
- **At max scale (10 pods)**: ~1000 requests/second

### Storage
- **Compression**: LanceDB uses columnar compression
- **Typical ratio**: 3-5x compression vs. raw data
- **Index overhead**: ~10-20% of data size

## Fault Tolerance

### Pod Failures
- Knative automatically restarts failed pods
- Liveness/Readiness probes detect unhealthy pods
- Traffic rerouted to healthy pods

### Storage Failures
- MinIO data persisted on PVC
- PVC backed by underlying storage (local/NFS/cloud)
- **Recommendation**: Enable MinIO erasure coding for production

### Network Failures
- Automatic retries in AWS SDK (boto3)
- Exponential backoff for transient errors
- Circuit breaker pattern can be added

## Limitations

1. **Scale-to-zero cold start**: 2-5 second latency on first request
2. **In-memory processing**: Large tables may exceed memory limits
3. **Single-region MinIO**: No built-in geo-distribution
4. **No built-in auth**: Must add authentication layer for public access

## Future Enhancements

1. **Caching**: Add Redis for frequently accessed tables
2. **Authentication**: JWT or API key middleware
3. **Rate limiting**: Prevent abuse and ensure fair usage
4. **Multi-region**: Deploy MinIO in distributed mode
5. **Batch processing**: Support bulk insert/search operations
6. **Streaming**: WebSocket support for real-time updates
7. **Advanced indexing**: HNSW, IVF-PQ indices for faster search

## References

- [LanceDB Architecture](https://lancedb.com/docs/concepts/architecture/)
- [Knative Serving](https://knative.dev/docs/serving/)
- [MinIO S3 Compatibility](https://min.io/docs/minio/linux/index.html)
- [Prometheus Python Client](https://github.com/prometheus/client_python)

