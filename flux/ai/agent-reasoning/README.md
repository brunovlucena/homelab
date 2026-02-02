# 🧠 Agent-Reasoning

**TinyRecursiveModels (TRM) Service for Homelab Agents**

A Knative service that provides recursive reasoning capabilities using TinyRecursiveModels (7M parameters). This service enables homelab agents to solve complex reasoning tasks efficiently without requiring large LLMs.

## 🎯 Overview

Agent-Reasoning provides:
- **Recursive Reasoning**: Iteratively improves answers over K steps
- **Parameter Efficient**: Only 7M parameters (vs. billions in LLMs)
- **Cost Effective**: Runs on consumer GPUs (1 L40S)
- **Specialized**: Excellent for structured reasoning (planning, optimization, logic)

## 📋 Quick Start

```bash
# Install dependencies
make install

# Run locally (requires GPU)
make run-dev

# Test the reasoning service
curl -X POST http://localhost:8080/reason \
  -H "Content-Type: application/json" \
  -d '{
    "question": "How should I optimize my Kubernetes cluster?",
    "context": {"nodes": 10, "workloads": []},
    "max_steps": 6
  }'
```

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────┐
│              Agent-Reasoning Service                    │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌──────────────┐         ┌──────────────┐             │
│  │   FastAPI    │────────▶│  TRM Handler │             │
│  │   Endpoints  │         │              │             │
│  └──────────────┘         └──────┬───────┘             │
│         │                        │                      │
│  ┌──────┴──────┐         ┌───────┴───────┐             │
│  │ CloudEvents │         │  TRM Model    │             │
│  │  Publisher  │         │  (7M params) │             │
│  └─────────────┘         └───────────────┘             │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

## 📡 CloudEvents Integration

### Events Emitted

| Event Type | Trigger | Purpose |
|------------|---------|---------|
| `io.homelab.reasoning.completed` | Reasoning task completed | Notify requesting agent |
| `io.homelab.reasoning.failed` | Reasoning task failed | Error notification |

### Events Received

| Event Type | Source | Effect |
|------------|--------|--------|
| `io.homelab.reasoning.requested` | Any agent | Process reasoning task |

## 📁 Project Structure

```
agent-reasoning/
├── src/
│   ├── reasoning/
│   │   ├── __init__.py
│   │   ├── main.py          # FastAPI entry point
│   │   ├── handler.py       # TRM inference handler
│   │   ├── model_loader.py  # TRM model loading
│   │   └── Dockerfile
│   ├── shared/
│   │   ├── __init__.py
│   │   ├── types.py         # Shared types
│   │   ├── metrics.py       # Prometheus metrics
│   │   └── events.py        # CloudEvents
│   └── requirements.txt
├── k8s/
│   └── kustomize/
│       ├── base/
│       │   ├── kustomization.yaml
│       │   ├── namespace.yaml
│       │   ├── service.yaml
│       │   └── configmap.yaml
│       ├── studio/
│       └── pro/
├── tests/
│   ├── unit/
│   └── conftest.py
├── Makefile
└── README.md
```

## 🔧 Configuration

| Environment Variable | Description | Default |
|---------------------|-------------|---------|
| `MODEL_PATH` | Path to TRM checkpoint | `/models/trm-checkpoint.pth` |
| `DEVICE` | Device for inference (cuda/cpu) | `cuda` |
| `MAX_STEPS` | Maximum reasoning steps | `6` |
| `H_CYCLES` | High-level cycles | `3` |
| `L_CYCLES` | Low-level cycles | `6` |
| `EMIT_EVENTS` | Enable CloudEvent emission | `true` |
| `KNATIVE_BROKER_URL` | Broker ingress URL | Auto-detected |

## 🚀 Deployment

### Prerequisites

- GPU node available (L40S or similar)
- TRM model checkpoint trained/available
- Knative Serving installed

### Deploy to Homelab

```bash
# Build image
make build

# Push to registry
make push

# Deploy to Kubernetes
make deploy-studio
```

## 📊 API Endpoints

### POST /reason

Process a reasoning task with TRM.

**Request**:
```json
{
  "question": "How should I optimize my Kubernetes cluster?",
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

### POST /events

Receive CloudEvents for reasoning tasks.

### GET /health

Health check endpoint.

### GET /metrics

Prometheus metrics endpoint.

## 📈 Monitoring

Metrics exposed:
- `agent_reasoning_requests_total{task_type, status}` - Total requests
- `agent_reasoning_duration_seconds{task_type}` - Processing time
- `agent_reasoning_steps{task_type}` - Reasoning steps used
- `agent_reasoning_confidence{task_type}` - Answer confidence
- `agent_reasoning_gpu_utilization` - GPU usage

## 🔒 Security

- Input validation and sanitization
- Rate limiting per client
- Resource limits (GPU, memory)
- No external API calls

## 📄 License

MIT

