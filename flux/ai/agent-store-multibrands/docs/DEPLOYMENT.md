# 🚀 Deployment Guide

## Prerequisites

1. **Kubernetes Cluster** with:
   - Knative Serving installed
   - Knative Eventing installed
   - knative-lambda-operator deployed

2. **Infrastructure**:
   - RabbitMQ cluster for eventing
   - Ollama for LLM inference (or OpenAI API key)
   - Alloy/Grafana for observability

3. **External Services**:
   - WhatsApp Business API account
   - Meta Developer account for webhook configuration

## Quick Start

### 1. Create Namespace and Secrets

```bash
# Apply base configuration
kubectl apply -k k8s/kustomize/base/namespace.yaml
kubectl apply -k k8s/kustomize/base/rbac.yaml

# Edit secrets with real credentials
kubectl edit secret whatsapp-credentials -n agent-store-multibrands
```

### 2. Build Container Images

```bash
# Build all images
make build

# Or build individually
docker build -t ghcr.io/brunovlucena/agent-store-multibrands/whatsapp_gateway:v0.1.0 \
  -f src/whatsapp_gateway/Dockerfile src/
```

### 3. Push to Registry

```bash
# Login to GitHub Container Registry
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Push all images
make push
```

### 4. Deploy to Cluster

```bash
# For studio/dev environment
kubectl apply -k k8s/kustomize/studio/

# For production
kubectl apply -k k8s/kustomize/pro/

# Check deployment status
kubectl get lambdaagents -n agent-store-multibrands
```

### 5. Configure WhatsApp Webhook

1. Go to Meta Developer Console
2. Navigate to your WhatsApp Business App
3. Configure webhook URL:
   ```
   https://your-domain.com/webhook
   ```
4. Set verify token (from secret)
5. Subscribe to `messages` webhook field

## Configuration

### Environment Variables

#### WhatsApp Gateway
| Variable | Description | Required |
|----------|-------------|----------|
| `WHATSAPP_PHONE_NUMBER_ID` | Meta phone number ID | Yes |
| `WHATSAPP_ACCESS_TOKEN` | Meta access token | Yes |
| `WHATSAPP_VERIFY_TOKEN` | Webhook verify token | Yes |
| `WHATSAPP_APP_SECRET` | App secret for signature verification | Yes |

#### AI Sellers
| Variable | Description | Default |
|----------|-------------|---------|
| `OLLAMA_URL` | Ollama API endpoint | `http://ollama-native.ollama.svc.cluster.local:11434` |
| `OLLAMA_MODEL` | LLM model to use | `llama3.2:3b` |
| `AGENT_BRAND` | Brand this seller handles | Required |

### Scaling Configuration

Edit the LambdaAgent manifests to adjust scaling:

```yaml
scaling:
  minReplicas: 0      # Scale to zero when idle
  maxReplicas: 10     # Maximum instances
  targetConcurrency: 5  # Requests per instance
  scaleToZeroGracePeriod: 120s  # Wait before scaling down
```

## Testing

### Run Smoke Tests

```bash
# Deploy k6 tests
kubectl apply -k k8s/tests/

# Watch test progress
kubectl logs -f -l app=k6 -n agent-store-multibrands
```

### Manual Testing

```bash
# Port forward to WhatsApp Gateway
kubectl port-forward -n agent-store-multibrands svc/whatsapp-gateway 8081:80

# Test health
curl http://localhost:8081/health

# Test sending message (manual)
curl -X POST http://localhost:8081/send \
  -H "Content-Type: application/json" \
  -d '{"phone": "5511999999999", "message": "Hello!"}'
```

### Test AI Seller

```bash
# Port forward to AI seller
kubectl port-forward -n agent-store-multibrands svc/ai-seller-fashion 8082:80

# Test chat
curl -X POST http://localhost:8082/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Quero ver vestidos",
    "customer_id": "test-123",
    "customer_phone": "5511999999999"
  }'
```

## Monitoring

### View Metrics

```bash
# Port forward to get metrics
kubectl port-forward -n agent-store-multibrands svc/whatsapp-gateway 8081:80
curl http://localhost:8081/metrics
```

### View Logs

```bash
# All agents
kubectl logs -n agent-store-multibrands -l app.kubernetes.io/part-of=agent-store-multibrands -f

# Specific agent
kubectl logs -n agent-store-multibrands -l app.kubernetes.io/name=ai-seller-fashion -f
```

### Grafana Dashboard

Import the dashboard from `docs/dashboards/store-overview.json` in your Grafana instance.

## Troubleshooting

### Common Issues

#### 1. Agents Not Starting
```bash
# Check LambdaAgent status
kubectl describe lambdaagent whatsapp-gateway -n agent-store-multibrands

# Check Knative service
kubectl get ksvc -n agent-store-multibrands
```

#### 2. Events Not Flowing
```bash
# Check broker status
kubectl get broker -n agent-store-multibrands

# Check triggers
kubectl get triggers -n agent-store-multibrands
```

#### 3. LLM Timeouts
```bash
# Check Ollama availability
kubectl get pods -n ollama

# Test Ollama directly
kubectl port-forward -n ollama svc/ollama 11434:11434
curl http://localhost:11434/api/tags
```

#### 4. WhatsApp Webhook Failing
- Verify webhook URL is publicly accessible
- Check signature verification is passing
- Ensure access token is valid

## Upgrading

### Rolling Update

```bash
# Update image tag
kubectl set image lambdaagent/ai-seller-fashion \
  container=ghcr.io/brunovlucena/agent-store-multibrands/ai_seller:v0.2.0 \
  -n agent-store-multibrands
```

### Full Redeploy

```bash
# Update VERSION file
echo "0.2.0" > VERSION

# Rebuild and push
make build push

# Apply updated manifests
kubectl apply -k k8s/kustomize/pro/
```

## Backup & Recovery

### Order Data
Currently orders are stored in-memory. For production, integrate with:
- PostgreSQL for persistent storage
- Redis for caching
- Event sourcing for audit trail

### Conversation History
Conversations are ephemeral. For persistence:
- Store in Redis with TTL
- Archive to object storage for analytics
