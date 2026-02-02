# TRM Fine-Tuning Pipeline for Homelab

> **Purpose**: Fine-tune Tiny Recursive Models (TRM) on notifi-services code and observability data  
> **Schedule**: Automatic fine-tuning every 30 days  
> **Platform**: Flyte on Forge cluster (GPU-enabled)

---

## Overview

This pipeline fine-tunes TRM (Tiny Recursive Models) on:
1. **Notifi-services codebase** - C# files, configs, templates
2. **Observability data** - Prometheus metrics, Loki logs, and Tempo traces from last 30 days

The model learns to:
- Understand code structure and patterns
- Analyze metrics and logs
- Perform recursive reasoning on infrastructure data

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│              TRM Fine-Tuning Pipeline                        │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  1. Data Collection                                          │
│     ├─ Notifi-services code (C#, YAML, JSON, etc.)          │
│     └─ Observability (Prometheus + Loki, last 30 days)      │
│                                                               │
│  2. Data Formatting                                          │
│     └─ Convert to TRM training format (problem → solution)    │
│                                                               │
│  3. Model Training                                           │
│     ├─ Fine-tune TRM 7M model                               │
│     ├─ Recursive reasoning cycles                            │
│     └─ Save checkpoints                                      │
│                                                               │
│  4. Evaluation & Deployment                                  │
│     ├─ Evaluate on test set                                 │
│     ├─ Upload to MinIO                                       │
│     └─ Update Ollama/VLLM                                    │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

---

## Components

### 1. Data Collector (`src/data_collector.py`)

Collects training data from multiple sources:

- **NotifiServicesCollector**: Scans notifi-services repo for code files
- **ObservabilityCollector**: Queries Prometheus, Loki, and Tempo for metrics/logs/traces
- **DataCollector**: Orchestrates collection and formats for TRM

### 2. TRM Trainer (`src/trm_trainer.py`)

Fine-tunes TRM model:

- Prepares dataset in TRM format
- Runs training using TRM's `pretrain.py`
- Exports model for deployment

### 3. Flyte Workflow (`src/flyte_workflow.py`)

Orchestrates the pipeline:

- **`trm_finetuning_workflow`**: Main workflow
- **`scheduled_trm_finetuning`**: Scheduled (every 30 days)

---

## Setup

### Prerequisites

1. **Flyte** installed on Forge cluster ✅
2. **GPU nodes** available on Forge cluster ✅
3. **Prometheus & Loki** accessible ✅
4. **Notifi-services** repository accessible ✅
5. **MinIO** for model storage ✅

### Installation

```bash
cd flux/ai/trm-finetune

# Build Docker image
docker build -t localhost:5001/trm-finetune:latest .

# Push to registry
docker push localhost:5001/trm-finetune:latest
```

### Deploy to Kubernetes

```bash
# Apply base configuration
kubectl apply -k k8s/kustomize/base/

# Verify
kubectl get configmap -n ml-platform trm-finetune-config
```

---

## Usage

### Manual Trigger

```bash
# Using Flyte CLI
flytectl create execution \
  --project homelab \
  --domain production \
  --workflow trm_finetuning_workflow \
  --inputs '{
    "days": 30,
    "notifi_services_path": "/workspace/notifi/repos/notifi-services",
    "prometheus_url": "http://prometheus.monitoring.svc:9090",
    "loki_url": "http://loki.monitoring.svc:3100"
  }'
```

### Scheduled Execution

The workflow runs automatically on the **1st of every month at 2 AM**:

```python
# Already configured in flyte_workflow.py
@workflow(
    schedule=CronSchedule(
        schedule="0 2 1 * *",  # First day of every month at 2 AM
        kickoff_time_input_arg="trigger_time",
    )
)
def scheduled_trm_finetuning(...):
    ...
```

### Register Workflow

```bash
# Register with Flyte
pyflyte register src/flyte_workflow.py \
  --project homelab \
  --domain production \
  --image localhost:5001/trm-finetune:latest
```

---

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `NOTIFI_SERVICES_PATH` | `/workspace/notifi/repos/notifi-services` | Path to notifi-services repo |
| `PROMETHEUS_URL` | `http://prometheus.monitoring.svc:9090` | Prometheus API URL |
| `LOKI_URL` | `http://loki.monitoring.svc:3100` | Loki API URL |
| `TEMPO_URL` | `http://tempo.tempo.svc:3200` | Tempo API URL |
| `DATA_DAYS` | `30` | Days of observability data to collect |
| `EPOCHS` | `50000` | Training epochs |
| `LR` | `1e-4` | Learning rate |
| `L_LAYERS` | `2` | Number of layers |
| `H_CYCLES` | `3` | High-level cycles |
| `L_CYCLES` | `6` | Low-level cycles |

### Training Configuration

Edit `k8s/kustomize/base/configmap.yaml` to adjust training parameters.

---

## Data Format

Training examples are in JSONL format:

```json
{
  "problem": "Analyze Prometheus metric: up",
  "initial_answer": "",
  "solution": "Metric up indicates service availability...",
  "reasoning_steps": [
    "Step 1: Understand metric type",
    "Step 2: Analyze time series",
    "Step 3: Identify patterns",
    "Step 4: Generate insights"
  ],
  "metadata": {
    "source": "prometheus",
    "query": "up",
    "type": "metric_analysis",
    "timestamp": "2025-12-28T10:00:00"
  }
}
```

---

## Model Output

Trained models are saved to:
- **Local**: `/tmp/trm-model/export/`
- **MinIO**: `s3://trm-models/YYYYMMDD_HHMMSS/`

Model can be loaded for inference:

```python
from models.recursive_reasoning import TRMModel

model = TRMModel.load_from_checkpoint("path/to/model.ckpt")
result = model.reason(problem, max_iterations=10)
```

---

## Monitoring

### Flyte Dashboard

View workflow execution:
- **URL**: `http://flyte.ml-platform.svc.forge.remote:81`
- **Project**: `homelab`
- **Domain**: `production`

### Logs

```bash
# View workflow logs
flytectl get execution-logs \
  --project homelab \
  --domain production \
  <execution-id>

# Or via kubectl
kubectl logs -n ml-platform -l app=trm-finetune
```

---

## Troubleshooting

### Data Collection Fails

- **Check Prometheus/Loki URLs**: Verify services are accessible
- **Check notifi-services path**: Ensure repository is mounted
- **Check network**: Verify cluster DNS resolution

### Training Fails

- **Check GPU availability**: `kubectl get nodes -l accelerator=nvidia-tesla-v100`
- **Check CUDA**: Verify CUDA toolkit is installed
- **Check memory**: Increase `mem` resource request if OOM

### Model Export Fails

- **Check MinIO access**: Verify credentials in secrets
- **Check disk space**: Ensure sufficient storage

---

## Integration with Agents

After training, deploy model to agents:

### Option 1: Ollama

```bash
# Import model to Ollama
ollama import /path/to/model

# Use in agent
curl http://ollama.ollama.svc:11434/api/generate \
  -d '{
    "model": "trm-homelab",
    "prompt": "Analyze this metric..."
  }'
```

### Option 2: VLLM

```yaml
# Update VLLM deployment
apiVersion: v1
kind: ConfigMap
metadata:
  name: vllm-config
data:
  model_path: "s3://trm-models/latest/"
```

---

## References

- **TRM Repository**: https://github.com/SamsungSAILMontreal/TinyRecursiveModels
- **Flyte Documentation**: https://docs.flyte.org/
- **Prometheus API**: https://prometheus.io/docs/prometheus/latest/querying/api/
- **Loki API**: https://grafana.com/docs/loki/latest/api/

---

## License

MIT License - Same as TRM repository

