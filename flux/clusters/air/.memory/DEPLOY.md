# Deploying Air Cluster

Quick guide to deploy the minimal Air cluster for testing Agent-Reasoning.

## Step 1: Create Kind Cluster

```bash
cd /Users/brunolucena/workspace/bruno/repos/homelab

# Create cluster
kind create cluster --config flux/clusters/air/kind.yaml --name air

# Verify
kubectl cluster-info --context kind-air
```

## Step 2: Build and Load Agent-Reasoning Image

```bash
cd flux/ai/agent-reasoning

# Build image
docker build -t localhost:5000/agent-reasoning:latest \
  -f src/reasoning/Dockerfile .

# Load into kind
kind load docker-image localhost:5000/agent-reasoning:latest --name air
```

## Step 3: Deploy Infrastructure (Manual)

Since we're testing without Flux, deploy manually:

```bash
cd /Users/brunolucena/workspace/bruno/repos/homelab

# Deploy core infrastructure
kubectl apply -k flux/clusters/air/deploy/01-core

# Wait for cert-manager and knative-operator to be ready
kubectl wait --for=condition=Available deployment -n cert-manager --all --timeout=300s
kubectl wait --for=condition=Available deployment -n knative-operator --all --timeout=300s

# Deploy observability
kubectl apply -k flux/clusters/air/deploy/02-observability

# Deploy Knative dependencies
kubectl apply -k flux/clusters/air/deploy/03-knative-deps

# Wait for Knative to be ready
kubectl wait --for=condition=Ready knativeserving -n knative-serving --all --timeout=600s
kubectl wait --for=condition=Ready knativeeventing -n knative-eventing --all --timeout=600s

# Deploy Agent-Reasoning
kubectl apply -k flux/ai/agent-reasoning/k8s/kustomize/air
```

## Step 4: Verify Deployment

```bash
# Check Agent-Reasoning service
kubectl get ksvc -n ai-agents agent-reasoning

# Wait for service to be ready
kubectl wait --for=condition=Ready ksvc/agent-reasoning -n ai-agents --timeout=300s

# Get service URL
kubectl get ksvc agent-reasoning -n ai-agents -o jsonpath='{.status.url}'
```

## Step 5: Test

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

## Alternative: Port Forward

If service URL doesn't work, use port forward:

```bash
# Port forward
kubectl port-forward -n ai-agents svc/agent-reasoning 8080:80

# In another terminal
curl http://localhost:8080/health
```

## Cleanup

```bash
# Delete cluster
kind delete cluster --name air
```


