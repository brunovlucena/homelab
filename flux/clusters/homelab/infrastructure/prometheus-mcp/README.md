# Prometheus MCP Server - Knative Serving

This service provides a Model Context Protocol (MCP) server for interacting with Prometheus metrics using **Knative Serving** for serverless-style deployment.

## Features

- 🚀 **Knative Serving**: Auto-scales from 1-3 replicas based on load
- 📊 **Prometheus Integration**: Direct access to your Prometheus metrics
- 🔍 **PromQL Queries**: Execute instant and range queries
- 📈 **Metrics Discovery**: List all available metrics
- 🔄 **Auto-scaling**: Scales to zero when idle (configurable)

## Architecture

```
┌─────────────────┐
│  MCP Clients    │
│  (Cursor, CLI)  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Knative Service │
│ prometheus-mcp  │
│  (Auto-scales)  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Prometheus     │
│   Operator      │
└─────────────────┘
```

## Configuration

### Prometheus URL
The service connects to Prometheus at:
```
http://prometheus-kube-prometheus-prometheus.prometheus.svc.cluster.local:9090
```

### Auto-scaling Settings
- **Min Scale**: 1 replica (always ready)
- **Max Scale**: 3 replicas
- **Target Concurrency**: 10 requests per pod

### Resource Limits
- **CPU**: 100m (request) / 500m (limit)
- **Memory**: 128Mi (request) / 512Mi (limit)

## Deployment

This service is managed by Flux and will be automatically deployed when merged.

### Manual Testing
```bash
# Check service status
kubectl get ksvc -n prometheus-mcp

# Get service URL
kubectl get ksvc prometheus-mcp -n prometheus-mcp -o jsonpath='{.status.url}'

# Port forward for local testing
kubectl port-forward -n prometheus-mcp svc/prometheus-mcp 8080:80

# Test health endpoint
curl http://localhost:8080/health
```

## Available MCP Tools

1. **GetAvailableWorkspaces** - List Prometheus workspaces
2. **ExecuteQuery** - Run instant PromQL queries
3. **ExecuteRangeQuery** - Run time-range queries
4. **ListMetrics** - List all available metrics
5. **GetServerInfo** - Get Prometheus server information

## Environment Variables

| Variable | Value | Description |
|----------|-------|-------------|
| `PORT` | `8080` | Server listening port |
| `PROMETHEUS_URL` | (cluster URL) | Prometheus endpoint |
| `FASTMCP_LOG_LEVEL` | `INFO` | MCP framework log level |
| `LOG_LEVEL` | `INFO` | Application log level |

## Troubleshooting

### Service not accessible
```bash
# Check service status
kubectl get ksvc prometheus-mcp -n prometheus-mcp

# Check pod logs
kubectl logs -n prometheus-mcp -l app=prometheus-mcp

# Check Knative serving status
kubectl get pods -n knative-serving
```

### Prometheus connection issues
```bash
# Test Prometheus connectivity from within the cluster
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- \
  curl http://prometheus-kube-prometheus-prometheus.prometheus.svc.cluster.local:9090/api/v1/status/config
```

## Differences from Container Deployment

| Feature | Container | Knative Serving |
|---------|-----------|-----------------|
| Auto-scaling | Manual HPA | Built-in |
| Cold starts | None | ~2-5 seconds |
| Resource efficiency | Always running | Scales to min |
| Load balancing | Service | Istio/Kourier |
| Observability | Manual | Built-in metrics |

## Benefits of Knative Serving

- ✅ **Auto-scaling**: Automatically scales based on traffic
- ✅ **Resource Efficient**: Can scale to 1 replica when idle
- ✅ **Blue-Green Deployments**: Built-in traffic splitting
- ✅ **Revision Management**: Easy rollbacks
- ✅ **Observability**: Built-in metrics and tracing

## Monitoring

The service exposes metrics compatible with Prometheus:
- Request count
- Request duration
- Active connections
- Error rates

Access via:
```bash
kubectl port-forward -n prometheus-mcp svc/prometheus-mcp 8080:80
curl http://localhost:8080/metrics
```

