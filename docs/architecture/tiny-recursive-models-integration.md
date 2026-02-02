# 🧠 TinyRecursiveModels Integration Guide

> **Part of**: [AI Agent Architecture](ai-agent-architecture.md)  
> **Related**: [Agent Orchestration](agent-orchestration.md) | [AI Components](ai-components.md)  
> **Last Updated**: January 2025

---

## Overview

This guide explains how to integrate [TinyRecursiveModels (TRM)](https://github.com/SamsungSAILMontreal/TinyRecursiveModels) into your homelab agents. TRM is a recursive reasoning model that achieves impressive results on complex reasoning tasks (ARC-AGI, Sudoku, Maze solving) using only 7M parameters.

### Why TRM?

- **Parameter Efficient**: 7M parameters vs. billions in LLMs
- **Recursive Reasoning**: Iteratively improves answers over K steps
- **Cost Effective**: Can run on consumer GPUs (1 L40S for Sudoku)
- **Specialized**: Excellent for structured reasoning tasks (puzzles, logic, planning)

### Use Cases in Homelab

1. **Complex Problem Solving**: When agents need to reason through multi-step problems
2. **Planning Tasks**: Infrastructure planning, resource optimization
3. **Logic Puzzles**: Troubleshooting complex issues, dependency resolution
4. **Iterative Refinement**: Improving answers over multiple reasoning cycles

---

## Architecture Integration

### Option 1: Standalone Reasoning Service (Recommended)

Deploy TRM as a dedicated Knative service that other agents can call:

```
┌─────────────────────────────────────────────────────────┐
│              Agent Architecture with TRM                │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌──────────────┐         ┌──────────────┐             │
│  │ Agent-Bruno  │────────▶│ Agent-Reason │             │
│  │  (Chatbot)   │         │   (TRM)      │             │
│  └──────────────┘         └──────┬───────┘             │
│         │                        │                      │
│  ┌──────┴──────┐         ┌───────┴───────┐             │
│  │   Ollama    │         │  TRM Model    │             │
│  │   (LLM)     │         │  (7M params)  │             │
│  └─────────────┘         └───────────────┘             │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

**Benefits**:
- Isolated resource usage
- Can scale independently
- Easy to update/version
- Other agents can use it via HTTP/CloudEvents

### Option 2: Embedded in Existing Agents

Integrate TRM directly into agents that need reasoning capabilities:

- `agent-auditor`: For complex SRE reasoning tasks
- `agent-jamie`: For data science problem solving
- `agent-bruno`: For complex user queries requiring reasoning

**Benefits**:
- Lower latency (no network call)
- Simpler deployment
- Better for high-frequency use

---

## Implementation Guide

### Step 1: Clone and Setup TRM

```bash
cd /Users/brunolucena/workspace/bruno/repos/homelab/flux/ai
git clone https://github.com/SamsungSAILMontreal/TinyRecursiveModels.git trm
cd trm

# Install dependencies
pip install --upgrade pip wheel setuptools
pip install --pre --upgrade torch torchvision torchaudio --index-url https://download.pytorch.org/whl/nightly/cu126
pip install -r requirements.txt
```

### Step 2: Train or Download Pre-trained Models

For production use, you'll need to:

1. **Train on your domain** (recommended for homelab-specific tasks)
2. **Use pre-trained models** (if available for your use case)
3. **Fine-tune** existing models on homelab data

Example training command (Sudoku):
```bash
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

### Step 3: Create Agent-Reasoning Service

See `agent-reasoning/` directory for a complete implementation.

Key components:
- **FastAPI service** exposing TRM inference
- **Model loader** for TRM checkpoints
- **CloudEvents integration** for event-driven reasoning
- **Prometheus metrics** for observability

### Step 4: Integrate with Existing Agents

#### Example: Using from Agent-Bruno

```python
# In agent-bruno/src/chatbot/handler.py

from shared.reasoning import ReasoningClient

class ChatBot:
    def __init__(self, ...):
        # ... existing init ...
        self.reasoning_client = ReasoningClient(
            base_url=os.getenv("REASONING_SERVICE_URL", 
                             "http://agent-reasoning.ai-agents.svc.cluster.local:8080")
        )
    
    async def chat(self, message: str, ...):
        # Detect if message requires complex reasoning
        if self._needs_reasoning(message):
            # Use TRM for complex reasoning
            reasoning_result = await self.reasoning_client.reason(
                question=message,
                context=self._build_context(),
                max_steps=6
            )
            return reasoning_result.answer
        
        # Otherwise use normal LLM path
        return await self._query_ollama(...)
    
    def _needs_reasoning(self, message: str) -> bool:
        """Detect if message requires complex reasoning."""
        reasoning_keywords = [
            "solve", "plan", "optimize", "figure out", 
            "calculate", "determine", "analyze"
        ]
        return any(kw in message.lower() for kw in reasoning_keywords)
```

---

## API Design

### Reasoning Service Endpoints

#### POST /reason

Process a reasoning task with TRM.

**Request**:
```json
{
  "question": "How should I optimize my Kubernetes cluster for cost?",
  "context": {
    "current_nodes": 10,
    "workloads": [...],
    "constraints": [...]
  },
  "max_steps": 6,
  "task_type": "planning"
}
```

**Response**:
```json
{
  "answer": "Based on recursive reasoning...",
  "steps": 6,
  "confidence": 0.87,
  "reasoning_trace": [...],
  "duration_ms": 1234.5
}
```

#### POST /events

Receive CloudEvents for reasoning tasks.

**Event Types**:
- `io.homelab.reasoning.requested`: Agent requests reasoning
- `io.homelab.reasoning.completed`: Reasoning task completed
- `io.homelab.reasoning.failed`: Reasoning task failed

---

## Deployment

### Kubernetes Deployment

Deploy as Knative service (scale-to-zero when idle):

```yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: agent-reasoning
  namespace: ai-agents
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/minScale: "0"
        autoscaling.knative.dev/maxScale: "10"
    spec:
      containers:
      - image: your-registry/agent-reasoning:latest
        resources:
          requests:
            memory: "4Gi"
            cpu: "2"
            nvidia.com/gpu: "1"  # GPU required for TRM
          limits:
            memory: "8Gi"
            cpu: "4"
            nvidia.com/gpu: "1"
        env:
        - name: MODEL_PATH
          value: "/models/trm-checkpoint.pth"
        - name: DEVICE
          value: "cuda"
```

### Resource Requirements

- **Minimum**: 1 GPU (L40S or similar, 48GB RAM)
- **Recommended**: 1-4 GPUs for production
- **Memory**: 4-8GB per instance
- **Storage**: ~500MB for model checkpoints

---

## Use Cases

### 1. Infrastructure Planning

**Scenario**: User asks "How should I deploy this application across my clusters?"

**Flow**:
1. Agent-Bruno receives question
2. Detects complex reasoning need
3. Calls Agent-Reasoning with:
   - Question: Deployment planning
   - Context: Current cluster state, app requirements
   - Task type: "planning"
4. TRM recursively reasons through:
   - Resource requirements
   - Network topology
   - Cost optimization
   - High availability
5. Returns structured plan

### 2. Troubleshooting

**Scenario**: "Why is my service slow?"

**Flow**:
1. Agent-Auditor receives alert
2. Calls Agent-Reasoning with:
   - Question: Root cause analysis
   - Context: Metrics, logs, topology
   - Task type: "troubleshooting"
3. TRM reasons through:
   - Dependency chains
   - Bottleneck identification
   - Solution prioritization
4. Returns diagnosis and recommendations

### 3. Optimization Tasks

**Scenario**: "Optimize my Kubernetes resource allocation"

**Flow**:
1. Agent-Jamie receives optimization request
2. Calls Agent-Reasoning with:
   - Question: Resource optimization
   - Context: Current allocations, workloads
   - Task type: "optimization"
3. TRM iteratively improves solution:
   - Step 1: Initial allocation
   - Step 2: Identify waste
   - Step 3: Rebalance
   - Step 4: Validate constraints
   - Step 5: Final optimization
4. Returns optimized configuration

---

## Monitoring

### Metrics

Key metrics to track:

- `agent_reasoning_requests_total{task_type, status}` - Total requests
- `agent_reasoning_duration_seconds{task_type}` - Processing time
- `agent_reasoning_steps{task_type}` - Reasoning steps used
- `agent_reasoning_confidence{task_type}` - Answer confidence
- `agent_reasoning_gpu_utilization` - GPU usage

### Observability

- **Tracing**: OpenTelemetry spans for reasoning steps
- **Logging**: Structured logs with reasoning traces
- **Alerting**: Alert on high failure rate or low confidence

---

## Best Practices

### 1. Task Detection

Only use TRM for tasks that benefit from recursive reasoning:
- ✅ Multi-step problems
- ✅ Planning tasks
- ✅ Optimization problems
- ✅ Logic puzzles
- ❌ Simple Q&A
- ❌ Fact retrieval
- ❌ Single-step operations

### 2. Context Building

Provide rich context to TRM:
- Current state
- Constraints
- Goals
- Historical data

### 3. Step Limits

Configure appropriate `max_steps`:
- Simple tasks: 3-4 steps
- Medium complexity: 6-8 steps
- Complex problems: 10-12 steps

### 4. Fallback Strategy

Always have a fallback:
```python
try:
    result = await reasoning_client.reason(...)
except ReasoningError:
    # Fallback to LLM or simple response
    result = await self._query_ollama(...)
```

---

## Training Custom Models

### For Homelab-Specific Tasks

1. **Collect training data**:
   - Infrastructure planning scenarios
   - Troubleshooting cases
   - Optimization problems

2. **Format data**:
   - Convert to TRM's puzzle format
   - Create embeddings
   - Augment with variations

3. **Train model**:
   ```bash
   python pretrain.py \
     arch=trm \
     data_paths="[data/homelab-reasoning]" \
     arch.H_cycles=3 arch.L_cycles=6 \
     +run_name=homelab_reasoning
   ```

4. **Evaluate**:
   - Test on held-out scenarios
   - Measure accuracy vs. LLM baseline
   - Monitor inference time

---

## Cost Analysis

### TRM vs. LLM

| Metric | TRM (7M) | LLM (70B) |
|--------|----------|-----------|
| Parameters | 7M | 70B |
| GPU Memory | ~500MB | ~140GB |
| Inference Time | 50-200ms | 1-5s |
| Cost per Request | $0.001 | $0.01-0.05 |
| Reasoning Quality | High (specialized) | High (general) |

**When to use TRM**:
- Structured reasoning tasks
- High-frequency requests
- Cost-sensitive applications
- Edge deployment

**When to use LLM**:
- General knowledge
- Creative tasks
- Natural language understanding
- Complex multi-modal tasks

---

## Troubleshooting

### Common Issues

1. **Out of Memory**
   - Reduce batch size
   - Use gradient checkpointing
   - Decrease model size

2. **Slow Inference**
   - Use GPU acceleration
   - Reduce max_steps
   - Optimize model loading

3. **Low Accuracy**
   - Train on domain-specific data
   - Increase training epochs
   - Tune hyperparameters

---

## References

- [TinyRecursiveModels GitHub](https://github.com/SamsungSAILMontreal/TinyRecursiveModels)
- [Paper: Less is More: Recursive Reasoning with Tiny Networks](https://arxiv.org/abs/2510.04871)
- [Homelab AI Architecture](ai-agent-architecture.md)

---

## Next Steps

1. ✅ Review this guide
2. ⬜ Clone TRM repository
3. ⬜ Deploy Agent-Reasoning service
4. ⬜ Integrate with existing agents
5. ⬜ Train domain-specific models
6. ⬜ Monitor and optimize

