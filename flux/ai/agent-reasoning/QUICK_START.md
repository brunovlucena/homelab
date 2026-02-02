# Quick Start: Using TinyRecursiveModels in Your Homelab

This guide will help you quickly integrate TinyRecursiveModels (TRM) into your homelab agents.

## Prerequisites

- Kubernetes cluster with GPU nodes (L40S or similar)
- Knative Serving installed
- Access to homelab cluster

## Step 1: Clone TRM Repository

```bash
cd /Users/brunolucena/workspace/bruno/repos/homelab/flux/ai
git clone https://github.com/SamsungSAILMontreal/TinyRecursiveModels.git trm
cd trm

# Install dependencies
pip install --upgrade pip wheel setuptools
pip install --pre --upgrade torch torchvision torchaudio --index-url https://download.pytorch.org/whl/nightly/cu126
pip install -r requirements.txt
```

## Step 2: Train or Download Model

### Option A: Use Pre-trained Model (if available)

```bash
# Download checkpoint to models directory
mkdir -p /models
# Place checkpoint.pth in /models/trm-checkpoint.pth
```

### Option B: Train Your Own Model

For a simple test, train on Sudoku:

```bash
# Prepare dataset
python dataset/build_sudoku_dataset.py \
  --output-dir data/sudoku-extreme-1k-aug-1000 \
  --subsample-size 1000 \
  --num-aug 1000

# Train model
run_name="pretrain_mlp_t_sudoku"
python pretrain.py \
  arch=trm \
  data_paths="[data/sudoku-extreme-1k-aug-1000]" \
  evaluators="[]" \
  epochs=50000 eval_interval=5000 \
  lr=1e-4 puzzle_emb_lr=1e-4 weight_decay=1.0 puzzle_emb_weight_decay=1.0 \
  arch.mlp_t=True arch.pos_encodings=none \
  arch.L_layers=2 \
  arch.H_cycles=3 arch.L_cycles=6 \
  +run_name=${run_name} ema=True
```

## Step 3: Deploy Agent-Reasoning Service

```bash
cd /Users/brunolucena/workspace/bruno/repos/homelab/flux/ai/agent-reasoning

# Build Docker image
make build

# Push to registry (update REGISTRY in Makefile)
make push

# Deploy to Kubernetes
make deploy-studio
```

## Step 4: Integrate with Existing Agent

### Add to Agent-Bruno

1. **Update requirements.txt**:
```bash
# Add to agent-bruno/src/requirements.txt
# Agent-Reasoning client
-e ../../shared-lib/agent_reasoning
```

2. **Update handler.py**:
```python
# Add import
from agent_reasoning import ReasoningClient, TaskType

# In ChatBot.__init__
self.reasoning_client = ReasoningClient(
    base_url=os.getenv(
        "REASONING_SERVICE_URL",
        "http://agent-reasoning.ai-agents.svc.cluster.local:8080"
    )
)

# In chat method, add reasoning check
if self._needs_reasoning(message):
    result = await self.reasoning_client.reason(
        question=message,
        context=self._build_context(),
        max_steps=6,
        task_type=self._detect_task_type(message),
    )
    return result.answer
```

3. **Update ConfigMap**:
```yaml
# k8s/kustomize/base/configmap.yaml
data:
  REASONING_ENABLED: "true"
  REASONING_SERVICE_URL: "http://agent-reasoning.ai-agents.svc.cluster.local:8080"
```

## Step 5: Test Integration

```bash
# Test reasoning service directly
curl -X POST http://agent-reasoning.ai-agents.svc.cluster.local:8080/reason \
  -H "Content-Type: application/json" \
  -d '{
    "question": "How should I optimize my Kubernetes cluster?",
    "context": {"nodes": 10},
    "max_steps": 6,
    "task_type": "optimization"
  }'

# Test via Agent-Bruno
curl -X POST http://agent-bruno.ai-agents.svc.cluster.local:8080/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "How should I optimize my Kubernetes cluster?"
  }'
```

## Step 6: Monitor

```bash
# Check reasoning service health
kubectl get pods -n ai-agents -l app=agent-reasoning

# View logs
kubectl logs -n ai-agents -l app=agent-reasoning -f

# Check metrics
curl http://agent-reasoning.ai-agents.svc.cluster.local:8080/metrics
```

## Troubleshooting

### Model Not Loading

- Check GPU availability: `kubectl get nodes -o json | jq '.items[].status.capacity."nvidia.com/gpu"'`
- Verify model path: Check `MODEL_PATH` environment variable
- Check logs: `kubectl logs -n ai-agents -l app=agent-reasoning`

### Service Not Responding

- Check service: `kubectl get svc -n ai-agents agent-reasoning`
- Check pods: `kubectl get pods -n ai-agents -l app=agent-reasoning`
- Check readiness: `curl http://agent-reasoning.ai-agents.svc.cluster.local:8080/ready`

### Low Accuracy

- Train on domain-specific data
- Increase training epochs
- Tune hyperparameters (H_cycles, L_cycles)

## Next Steps

1. Train domain-specific models for your use cases
2. Integrate with more agents (agent-auditor, agent-jamie)
3. Set up monitoring and alerting
4. Optimize for your specific workloads

For detailed information, see:
- [Integration Guide](../docs/architecture/tiny-recursive-models-integration.md)
- [Integration Example](INTEGRATION_EXAMPLE.md)
- [Agent-Reasoning README](README.md)

