# MLOps Training Pipeline Automation

> **Purpose**: Automated ML training pipeline for agent-sre and future agents  
> **Platform**: Flyte on Forge cluster (Kubernetes-native)  
> **Last Updated**: 2025-12-23

---

## Overview

The homelab uses **Flyte** (not GitHub Actions) for ML training pipeline automation. Flyte is a Kubernetes-native workflow orchestration platform specifically designed for ML/data pipelines.

### Why Flyte Over GitHub Actions?

| Feature | Flyte | GitHub Actions |
|---------|-------|----------------|
| **Kubernetes Native** | ✅ Built for K8s | ❌ External CI/CD |
| **GPU Support** | ✅ Native GPU scheduling | ❌ Limited GPU access |
| **Resource Management** | ✅ K8s resource quotas | ⚠️ Runner-based limits |
| **ML-Optimized** | ✅ Experiment tracking, versioning | ❌ Generic CI/CD |
| **Scalability** | ✅ Auto-scales with K8s | ⚠️ Runner capacity limits |
| **Cost Efficiency** | ✅ Use existing K8s cluster | ❌ Separate runner costs |
| **Reproducibility** | ✅ Versioned workflows | ⚠️ Less structured |
| **MLflow Integration** | ✅ Native support | ❌ Manual integration |

**Conclusion**: For ML training on Kubernetes, **Flyte is the industry standard** (2025).

---

## Architecture

### Training Pipeline Components

```
┌─────────────────────────────────────────────────────────────┐
│                    Flyte Workflow Engine                     │
│                  (Forge Cluster - K8s)                     │
└─────────────────────────────────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
        ▼                   ▼                   ▼
┌──────────────┐   ┌──────────────┐   ┌──────────────┐
│  Dataset     │   │   Model      │   │  Training   │
│  Preparation │   │   Conversion │   │  (LoRA)     │
└──────────────┘   └──────────────┘   └──────────────┘
        │                   │                   │
        └───────────────────┼───────────────────┘
                            │
                            ▼
                    ┌──────────────┐
                    │  Evaluation  │
                    └──────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
        ▼                   ▼                   ▼
┌──────────────┐   ┌──────────────┐   ┌──────────────┐
│   MLflow    │   │    MinIO     │   │  Git Version │
│  Registry   │   │   Storage    │   │   Update     │
└──────────────┘   └──────────────┘   └──────────────┘
```

### Infrastructure Components

1. **Flyte** (Forge cluster)
   - Workflow orchestration
   - Task scheduling
   - Resource management
   - Service: `flyte.ml-platform.svc.forge.remote:81`

2. **MLflow** (Forge cluster)
   - Experiment tracking
   - Model registry
   - Model versioning
   - Service: `mlflow.ml-platform.svc.forge.remote:5000`

3. **MinIO** (Forge cluster)
   - Model artifact storage
   - Long-term model storage
   - Service: `minio.data-ml.svc.forge.remote:30063`

4. **GitHub** (External)
   - Code repository
   - Model version tracking (MODEL_VERSION file)
   - GitOps integration

---

## Flyte Workflow

### Workflow Definition

The training pipeline is defined in `training/flyte_workflow.py`:

```python
@workflow
def agent_training_pipeline(
    agent_name: str = "agent-sre",
    model_name: str = "google/functiongemma-270m-it",
    runbook_path: str = "docs/RUNBOOK.md",
    learning_rate: float = 1e-4,
    batch_size: int = 4,
    iters: int = 1000,
) -> Dict[str, str]:
    """Complete training pipeline."""
    # 1. Prepare dataset
    dataset = prepare_dataset(...)
    
    # 2. Convert model
    mlx_model = convert_model_to_mlx(...)
    
    # 3. Train
    training_result = train_model_lora(...)
    
    # 4. Evaluate
    metrics = evaluate_model(...)
    
    # 5. Register in MLflow
    mlflow_uri = register_model_mlflow(...)
    
    # 6. Store in MinIO
    minio_path = store_model_minio(...)
    
    # 7. Update version
    git_commit = update_model_version(...)
    
    return {...}
```

### Workflow Tasks

1. **`prepare_dataset`**: Converts RUNBOOK.md → training JSONL
2. **`convert_model_to_mlx`**: HuggingFace → MLX format
3. **`train_model_lora`**: LoRA fine-tuning on M3 Ultra
4. **`evaluate_model`**: Test set evaluation
5. **`register_model_mlflow`**: MLflow model registry
6. **`store_model_minio`**: Long-term artifact storage
7. **`update_model_version`**: Git version update

---

## Deployment

### Prerequisites

1. **Flyte installed** on Forge cluster ✅ (already deployed)
2. **MLflow installed** on Forge cluster (verify)
3. **MinIO configured** on Forge cluster ✅ (already deployed)
4. **Flyte CLI** installed locally

### Install Flyte CLI

```bash
# macOS
brew install flytectl

# Or via pip
pip install flytekit flytectl
```

### Configure Flyte

```bash
# Create Flyte config
mkdir -p ~/.flyte
cat > ~/.flyte/config.yaml <<EOF
admin:
  endpoint: flyte.ml-platform.svc.forge.remote:81
  insecure: true
  project: homelab
  domain: production
EOF
```

### Deploy Workflow

```bash
cd flux/ai/agent-sre

# Package and register workflow
flytectl register files \
  --project homelab \
  --domain production \
  --archive flyte-package.tar.gz \
  --version $(git rev-parse --short HEAD)

# Or use pyflyte (Python SDK)
pyflyte register training/flyte_workflow.py \
  --project homelab \
  --domain production \
  --image ghcr.io/brunovlucena/agent-sre-training:latest
```

### Trigger Training

```bash
# Manual trigger
flytectl create execution \
  --project homelab \
  --domain production \
  --workflow agent_training_pipeline \
  --inputs '{"agent_name": "agent-sre", "iters": 1000}'

# Or via Python
from flytekit.remote import FlyteRemote
remote = FlyteRemote.from_config("homelab", "production")
execution = remote.execute(
    remote.fetch_workflow("homelab", "production", "agent_training_pipeline"),
    inputs={"agent_name": "agent-sre", "iters": 1000}
)
```

---

## Automated Triggers

### 1. Scheduled Retraining

Flyte supports scheduled workflows:

```python
from flytekit import LaunchPlan, CronSchedule

# Weekly retraining
weekly_retraining = LaunchPlan(
    name="weekly_retraining",
    workflow=agent_training_pipeline,
    schedule=CronSchedule(
        schedule="0 2 * * 0",  # Every Sunday at 2 AM
        kickoff_time_input_arg="scheduled_time",
    ),
    default_inputs={
        "agent_name": "agent-sre",
        "iters": 1000,
    },
)
```

### 2. GitOps Trigger (GitHub Actions)

GitHub Actions can trigger Flyte workflows:

```yaml
# .github/workflows/train-agent-sre.yml
name: Train Agent-SRE Model

on:
  schedule:
    - cron: '0 2 * * 0'  # Weekly
  workflow_dispatch:
    inputs:
      agent_name:
        default: agent-sre
      force_retrain:
        type: boolean
        default: false

jobs:
  trigger-training:
    runs-on: ubuntu-latest
    steps:
      - name: Trigger Flyte Workflow
        run: |
          flytectl create execution \
            --project homelab \
            --domain production \
            --workflow agent_training_pipeline \
            --inputs "{\"agent_name\": \"${{ inputs.agent_name }}\"}"
```

### 3. Data Drift Trigger

Monitor RUNBOOK.md changes and trigger retraining:

```yaml
# .github/workflows/retrain-on-runbook-change.yml
on:
  push:
    paths:
      - 'flux/ai/agent-sre/docs/RUNBOOK.md'

jobs:
  retrain:
    runs-on: ubuntu-latest
    steps:
      - name: Trigger Retraining
        run: |
          flytectl create execution \
            --project homelab \
            --domain production \
            --workflow agent_training_pipeline \
            --inputs '{"agent_name": "agent-sre"}'
```

### 4. Performance Degradation Trigger

Monitor model performance and auto-retrain:

```python
# Flyte task to monitor model performance
@task
def check_model_performance(agent_name: str) -> bool:
    """Check if model needs retraining."""
    import mlflow
    
    mlflow.set_tracking_uri("http://mlflow.ml-platform.svc.forge.remote:5000")
    
    # Get latest model metrics
    client = mlflow.tracking.MlflowClient()
    model_name = f"{agent_name}-functiongemma"
    
    latest_version = client.get_latest_versions(model_name, stages=["Production"])[0]
    metrics = client.get_run(latest_version.run_id).data.metrics
    
    # If accuracy < threshold, trigger retraining
    if metrics.get("accuracy", 1.0) < 0.80:
        return True  # Needs retraining
    
    return False
```

---

## Model Versioning Strategy

### Version Format

```
{agent_name}-functiongemma-{date}-{commit}
Example: agent-sre-functiongemma-20251223-a1b2c3d4
```

### Version Components

1. **Service Version** (`VERSION` file)
   - Application code version
   - Format: `MAJOR.MINOR.PATCH`

2. **Model Version** (`MODEL_VERSION` file)
   - Fine-tuned model version
   - Format: `v{YYYYMMDD}-{commit-sha}`
   - Updated by Flyte workflow

3. **MLflow Version**
   - Automatic versioning in MLflow registry
   - Stages: `Staging`, `Production`, `Archived`

### Version Update Flow

```
Flyte Workflow
    │
    ├─ Train Model
    ├─ Evaluate (accuracy > threshold)
    ├─ Register in MLflow (auto-versioned)
    ├─ Store in MinIO
    └─ Update MODEL_VERSION file
        │
        └─ Git Commit (via Flyte task)
            │
            └─ Flux Reconciliation
                │
                └─ Agent Deployment (new model version)
```

---

## Monitoring & Observability

### Flyte Dashboard

Access Flyte UI:
```bash
kubectl port-forward -n flyte svc/flyteconsole 8088:8088
# Open http://localhost:8088
```

### MLflow Tracking

Access MLflow UI:
```bash
kubectl port-forward -n ml-platform svc/mlflow 5000:5000
# Open http://localhost:5000
```

### Metrics to Monitor

1. **Training Metrics**:
   - Training loss
   - Validation loss
   - Training time
   - GPU utilization

2. **Model Metrics**:
   - Test accuracy
   - Command generation accuracy
   - Inference latency

3. **Pipeline Metrics**:
   - Workflow success rate
   - Task failure rate
   - Average pipeline duration
   - Resource utilization

### Prometheus Integration

Flyte exposes Prometheus metrics:

```promql
# Training pipeline success rate
rate(flyte_workflow_executions_total{status="success"}[5m])

# Average training duration
histogram_quantile(0.95, flyte_task_duration_seconds_bucket{task="train_model_lora"})

# Model accuracy over time
mlflow_model_accuracy{agent="agent-sre"}
```

---

## Multi-Agent Support

### Reusable Workflow Pattern

The same Flyte workflow supports multiple agents:

```python
# Train agent-sre
flytectl create execution \
  --workflow agent_training_pipeline \
  --inputs '{"agent_name": "agent-sre", "runbook_path": "flux/ai/agent-sre/docs/RUNBOOK.md"}'

# Train agent-bruno
flytectl create execution \
  --workflow agent_training_pipeline \
  --inputs '{"agent_name": "agent-bruno", "runbook_path": "flux/ai/agent-bruno/docs/RUNBOOK.md"}'

# Train agent-auditor
flytectl create execution \
  --workflow agent_training_pipeline \
  --inputs '{"agent_name": "agent-auditor", "runbook_path": "flux/ai/agent-auditor/docs/RUNBOOK.md"}'
```

### Agent-Specific Configuration

Each agent can have custom training parameters:

```yaml
# flux/ai/agent-sre/training/config.yaml
training:
  learning_rate: 1e-4
  batch_size: 4
  iters: 1000
  lora_layers: 16
  lora_rank: 8

# flux/ai/agent-bruno/training/config.yaml
training:
  learning_rate: 2e-4
  batch_size: 8
  iters: 2000
  lora_layers: 32
  lora_rank: 16
```

---

## Best Practices (2025)

### 1. Infrastructure as Code

- **Flyte workflows** defined in Python (versioned in Git)
- **Training configs** in YAML (per-agent)
- **Model versions** tracked in Git (MODEL_VERSION file)

### 2. Experiment Tracking

- **MLflow** for all experiments
- **Versioned datasets** (track RUNBOOK.md version)
- **Reproducible runs** (seed, environment, dependencies)

### 3. Resource Management

- **GPU scheduling** via Flyte (Forge cluster)
- **Resource quotas** per agent
- **Cost tracking** (monitor GPU hours)

### 4. Automated Testing

- **Unit tests** for training scripts
- **Integration tests** for Flyte workflows
- **Model validation** before registration

### 5. Continuous Monitoring

- **Model performance** tracking
- **Data drift** detection
- **Auto-retraining** triggers

---

## Comparison: Flyte vs Alternatives

### Flyte vs Kubeflow

| Feature | Flyte | Kubeflow |
|---------|-------|----------|
| **Complexity** | ✅ Simpler | ❌ More complex |
| **K8s Native** | ✅ Yes | ✅ Yes |
| **ML Focus** | ✅ Strong | ✅ Strong |
| **Community** | ✅ Growing | ✅ Established |
| **For Homelab** | ✅ **Better fit** | ⚠️ Overkill |

### Flyte vs Argo Workflows

| Feature | Flyte | Argo Workflows |
|---------|-------|----------------|
| **ML Features** | ✅ Built-in | ❌ Generic |
| **Experiment Tracking** | ✅ Native | ❌ Manual |
| **Model Registry** | ✅ MLflow integration | ❌ Manual |
| **For ML Training** | ✅ **Purpose-built** | ⚠️ Generic workflow |

### Flyte vs Prefect

| Feature | Flyte | Prefect |
|---------|-------|---------|
| **K8s Native** | ✅ Yes | ⚠️ Optional |
| **ML Focus** | ✅ Yes | ⚠️ Generic |
| **Deployment** | ✅ K8s-first | ⚠️ Cloud-first |
| **For Homelab** | ✅ **Better fit** | ⚠️ Cloud-oriented |

**Conclusion**: **Flyte is the best choice** for Kubernetes-native ML training pipelines in 2025.

---

## Quick Start

### 1. Verify Flyte Access

```bash
# Check Flyte connection
flytectl config validate

# List workflows
flytectl get workflows --project homelab --domain production
```

### 2. Register Workflow

```bash
cd flux/ai/agent-sre

# Build training image
docker build -t ghcr.io/brunovlucena/agent-sre-training:latest \
  -f training/Dockerfile training/

# Register workflow
pyflyte register training/flyte_workflow.py \
  --project homelab \
  --domain production \
  --image ghcr.io/brunovlucena/agent-sre-training:latest
```

### 3. Trigger Training

```bash
# Manual trigger
flytectl create execution \
  --project homelab \
  --domain production \
  --workflow agent_training_pipeline \
  --inputs '{"agent_name": "agent-sre", "iters": 1000}'

# Monitor execution
flytectl get execution <execution-id> --project homelab --domain production
```

### 4. View Results

```bash
# MLflow UI
kubectl port-forward -n ml-platform svc/mlflow 5000:5000
# Open http://localhost:5000

# Flyte UI
kubectl port-forward -n flyte svc/flyteconsole 8088:8088
# Open http://localhost:8088
```

---

## Troubleshooting

### Issue: Flyte Connection Failed

```bash
# Check Flyte service
kubectl get svc -n flyte flyte

# Verify network connectivity
curl http://flyte.ml-platform.svc.forge.remote:81/healthz
```

### Issue: GPU Not Available

```bash
# Check GPU nodes
kubectl get nodes -l nvidia.com/gpu=true --context=forge

# Check GPU resources
kubectl describe node <gpu-node> --context=forge | grep nvidia
```

### Issue: MLflow Connection Failed

```bash
# Check MLflow service
kubectl get svc -n ml-platform mlflow

# Verify MLflow is accessible
curl http://mlflow.ml-platform.svc.forge.remote:5000/health
```

---

## Future Enhancements

1. **Hyperparameter Tuning**: Integrate Optuna for auto-tuning
2. **A/B Testing**: Deploy multiple model versions
3. **Auto-Retraining**: Trigger on performance degradation
4. **Multi-Cluster**: Train on Forge, deploy to Studio
5. **Cost Optimization**: GPU sharing, spot instances

---

## References

- **Flyte Documentation**: https://docs.flyte.org/
- **MLflow Documentation**: https://mlflow.org/docs/latest/index.html
- **Forge Cluster**: `docs/clusters/forge-cluster.md`
- **ML Ops Best Practices**: Industry standards (2025)

---

**Last Updated**: 2025-12-23  
**Platform**: Flyte on Forge Cluster (Kubernetes)  
**Status**: ✅ Ready for Implementation

