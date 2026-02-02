# Flyte Sandbox - ML Training Pipeline

> **Purpose**: Kubernetes-native ML training pipeline automation for AI agents  
> **Platform**: Flyte on Forge cluster  
> **Last Updated**: 2025-12-23

---

## Overview

Flyte Sandbox provides a reusable, Kubernetes-native ML training pipeline for fine-tuning AI agent models. It uses **Flyte** (not GitHub Actions) for workflow orchestration, leveraging the Forge cluster's GPU resources.

### Why Flyte?

**Flyte is the industry standard for Kubernetes-native ML pipelines** (2025):

- ✅ **Kubernetes-native**: Built specifically for K8s
- ✅ **GPU support**: Native GPU scheduling on Forge cluster
- ✅ **ML-optimized**: Experiment tracking, versioning, reproducibility
- ✅ **Scalable**: Auto-scales with Kubernetes
- ✅ **Cost-efficient**: Uses existing K8s infrastructure

---

## Project Structure

```
flyte-sandbox/
├── workflows/
│   └── agent_training.py      # Main Flyte workflow
├── scripts/
│   ├── prepare_dataset.py     # Dataset preparation from RUNBOOK.md
│   ├── train.py               # LoRA fine-tuning script
│   └── evaluate.py            # Model evaluation script
├── config/
│   └── training_config.yaml   # Training configuration
├── docs/
│   └── MLOPS_PIPELINE.md     # Full documentation
├── Dockerfile                 # Container image for Flyte tasks
├── requirements.txt           # Python dependencies
└── README.md                  # This file
```

---

## Quick Start

### 1. Install Flyte CLI

```bash
# macOS
brew install flytectl

# Or via pip
pip install flytectl flytekit
```

### 2. Configure Flyte

```bash
# Create config
mkdir -p ~/.flyte
cat > ~/.flyte/config.yaml <<EOF
admin:
  endpoint: flyte.ml-platform.svc.forge.remote:81
  insecure: true
  project: homelab
  domain: production
EOF

# Validate
flytectl config validate
```

### 3. Build and Register Workflow

```bash
cd flyte-sandbox

# Build training image
docker build -t ghcr.io/brunovlucena/flyte-sandbox-training:latest .

# Push image
docker push ghcr.io/brunovlucena/flyte-sandbox-training:latest

# Register workflow
pyflyte register workflows/agent_training.py \
  --project homelab \
  --domain production \
  --image ghcr.io/brunovlucena/flyte-sandbox-training:latest
```

### 4. Trigger Training

```bash
# Manual trigger
flytectl create execution \
  --project homelab \
  --domain production \
  --workflow agent_training_pipeline \
  --inputs '{"agent_name": "agent-sre", "iters": 1000}'

# Monitor execution
flytectl get execution <exec-id> --project homelab --domain production
```

---

## Workflow Steps

1. **Prepare Dataset**: RUNBOOK.md → training JSONL
2. **Convert Model**: HuggingFace → MLX format
3. **Train**: LoRA fine-tuning
4. **Evaluate**: Test set evaluation
5. **Register**: MLflow model registry
6. **Store**: MinIO artifact storage

---

## Multi-Agent Support

The same workflow works for all agents:

```bash
# agent-sre
flytectl create execution ... --inputs '{"agent_name": "agent-sre"}'

# agent-bruno
flytectl create execution ... --inputs '{"agent_name": "agent-bruno"}'

# agent-auditor
flytectl create execution ... --inputs '{"agent_name": "agent-auditor"}'
```

---

## Monitoring

- **Flyte UI**: `http://flyte.ml-platform.svc.forge.remote:81`
- **MLflow UI**: `http://mlflow.ml-platform.svc.forge.remote:5000`
- **MinIO**: `http://minio.data-ml.svc.forge.remote:30063`

---

## Documentation

- **Full Guide**: `docs/MLOPS_PIPELINE.md`
- **Forge Cluster**: `../../homelab/docs/clusters/forge-cluster.md`

---

## Architecture

```
GitHub (Code) 
    ↓
GitHub Actions (Trigger)
    ↓
Flyte (Forge Cluster - K8s)
    ↓
Training Pipeline
    ├─ Dataset Prep
    ├─ Model Conversion
    ├─ LoRA Training
    ├─ Evaluation
    ├─ MLflow Registry
    └─ MinIO Storage
```

---

## Requirements

- Flyte installed on Forge cluster ✅
- MLflow installed on Forge cluster
- MinIO configured on Forge cluster ✅
- Access to Forge cluster from Studio

---

**Last Updated**: 2025-12-23  
**Status**: ✅ Ready for Use

