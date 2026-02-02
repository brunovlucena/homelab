# 🧪 Air Cluster - Minimal Test Cluster

Minimal Kubernetes cluster for testing Agent-Reasoning service.

## Overview

The Air cluster is a minimal test environment with only the essential components needed to run and test Agent-Reasoning:

- **1 Node**: Single control-plane node (no workers)
- **Network**: 10.248.0.0/16 (pods), 10.98.0.0/16 (services)
- **Purpose**: Testing Agent-Reasoning locally without full production stack

## Components

### Core Infrastructure (01-core)
- `cert-manager`: TLS certificate management
- `knative-operator`: Knative Serving/Eventing operators
- `sealed-secrets`: Secret management

### Observability (02-observability)
- `prometheus-operator`: Metrics collection (Agent-Reasoning exposes Prometheus metrics)

### Knative Dependencies (03-knative-deps)
- `rabbitmq-operator`: RabbitMQ operator
- `rabbitmq-broker`: CloudEvents broker (Agent-Reasoning uses CloudEvents)
- `knative-instances`: Knative Serving/Eventing instances

### Applications (07-apps)
- `agent-reasoning`: The TRM reasoning service

## What's NOT Included (vs. Pro)

- ❌ Linkerd (service mesh)
- ❌ Flagger (progressive delivery)
- ❌ Loki, Tempo, Alloy (full observability stack)
- ❌ Testing/CI components
- ❌ Multiple worker nodes
- ❌ GPU support (uses CPU only)
- ❌ Full data stack (PostgreSQL, MongoDB, etc.)

## Port Mappings

| Service | Container Port | Host Port |
|---------|---------------|-----------|
| Metrics Server | 30001 | 34001 |
| Knative Serving | 30130 | 34130 |
| Knative Eventing | 30131 | 34131 |
| RabbitMQ Broker | 30133 | 34133 |
| Prometheus | 30041 | 34041 |
| Agent-Reasoning | 30150 | 34150 |
| HTTP | 80 | 8480 |
| HTTPS | 443 | 8443 |

## Creating the Cluster

### Prerequisites

- Docker running
- Kind installed
- kubectl configured

### Create Cluster

```bash
# Create the kind cluster
kind create cluster --config flux/clusters/air/kind.yaml --name air

# Wait for cluster to be ready
kubectl wait --for=condition=Ready nodes --all --timeout=300s
```

### Deploy with Flux

```bash
# Bootstrap Flux (if not already done)
flux bootstrap github \
  --owner=your-org \
  --repository=homelab \
  --branch=main \
  --path=./flux/clusters/air

# Or apply manually
kubectl apply -k flux/clusters/air
```

### Deploy Agent-Reasoning

```bash
# Build and load image into kind
docker build -t localhost:5000/agent-reasoning:latest \
  -f flux/ai/agent-reasoning/src/reasoning/Dockerfile \
  flux/ai/agent-reasoning/

kind load docker-image localhost:5000/agent-reasoning:latest --name air

# Deploy Agent-Reasoning
kubectl apply -k flux/ai/agent-reasoning/k8s/kustomize/air
```

## Testing

### Check Service Status

```bash
# Check Knative service
kubectl get ksvc -n ai-agents agent-reasoning

# Check pods
kubectl get pods -n ai-agents -l app=agent-reasoning

# Check service URL
kubectl get ksvc agent-reasoning -n ai-agents -o jsonpath='{.status.url}'
```

### Test Endpoints

```bash
# Get service URL
SERVICE_URL=$(kubectl get ksvc agent-reasoning -n ai-agents -o jsonpath='{.status.url}')

# Test health
curl $SERVICE_URL/health

# Test reasoning
curl -X POST $SERVICE_URL/reason \
  -H "Content-Type: application/json" \
  -d '{
    "question": "How should I optimize my Kubernetes cluster?",
    "context": {"nodes": 10},
    "max_steps": 6,
    "task_type": "optimization"
  }'
```

### Port Forward (Alternative)

```bash
# Port forward to localhost
kubectl port-forward -n ai-agents svc/agent-reasoning 8080:80

# Test locally
curl http://localhost:8080/health
```

## Resource Requirements

- **CPU**: 1-2 cores
- **Memory**: 4-8GB
- **Storage**: 20-50GB
- **No GPU**: Uses CPU for TRM (slower but works for testing)

## Cleanup

```bash
# Delete cluster
kind delete cluster --name air
```

## Differences from Pro Cluster

| Feature | Pro | Air |
|---------|-----|-----|
| Nodes | 3 (1 CP + 2 workers) | 1 (CP only) |
| Service Mesh | Linkerd | None |
| Observability | Full stack | Prometheus only |
| GPU Support | Yes | No (CPU only) |
| Data Services | Full stack | None |
| Testing/CI | Included | Excluded |

## Troubleshooting

### Service Not Starting

```bash
# Check logs
kubectl logs -n ai-agents -l app=agent-reasoning

# Check events
kubectl get events -n ai-agents --sort-by='.lastTimestamp'
```

### Image Pull Errors

```bash
# Ensure image is loaded
kind load docker-image localhost:5000/agent-reasoning:latest --name air

# Check image
docker images | grep agent-reasoning
```

### Knative Not Ready

```bash
# Check Knative components
kubectl get pods -n knative-serving
kubectl get pods -n knative-eventing
```

## Next Steps

1. ✅ Cluster created
2. ✅ Components deployed
3. ✅ Agent-Reasoning deployed
4. ⬜ Test reasoning endpoints
5. ⬜ Integrate with other agents
6. ⬜ Monitor metrics in Prometheus


