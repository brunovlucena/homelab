# TRM Fine-Tuning Deployment Guide

## Quick Start

```bash
# 1. Run quick start script
./scripts/quick-start.sh

# 2. Register workflow with Flyte
pyflyte register src/flyte_workflow.py \
  --project homelab \
  --domain production \
  --image localhost:5001/trm-finetune:latest

# 3. Verify scheduled workflow
flytectl get launch-plan \
  --project homelab \
  --domain production \
  monthly_trm_finetuning
```

## Architecture Integration

### With Homelab Infrastructure

```
┌─────────────────────────────────────────────────────────┐
│                    HOMELAB INFRASTRUCTURE                 │
├─────────────────────────────────────────────────────────┤
│                                                           │
│  ┌───────────────────────────────────────────────────┐  │
│  │         FORGE CLUSTER (GPU Nodes)                 │  │
│  │                                                     │  │
│  │  ┌──────────────┐  ┌──────────────┐             │  │
│  │  │    Flyte     │  │   TRM Train   │             │  │
│  │  │  Workflows   │→ │   Pipeline    │             │  │
│  │  └──────────────┘  └──────────────┘             │  │
│  │         │                  │                       │  │
│  │         └──────────┬────────┘                       │  │
│  │                    ▼                                 │  │
│  │         ┌──────────────────┐                        │  │
│  │         │  MinIO Storage    │                        │  │
│  │         │  (Model Artifacts)│                        │  │
│  │         └──────────────────┘                        │  │
│  └───────────────────────────────────────────────────┘  │
│                                                           │
│  ┌───────────────────────────────────────────────────┐  │
│  │         STUDIO CLUSTER (AI Agents)                │  │
│  │                                                     │  │
│  │  ┌──────────────┐  ┌──────────────┐             │  │
│  │  │  Ollama      │  │  VLLM        │             │  │
│  │  │  (TRM Model) │  │  (TRM Model)  │             │  │
│  │  └──────────────┘  └──────────────┘             │  │
│  │         │                  │                       │  │
│  │         └──────────┬────────┘                       │  │
│  │                    ▼                                 │  │
│  │         ┌──────────────────┐                        │  │
│  │         │  AI Agents       │                        │  │
│  │         │  (Use TRM)       │                        │  │
│  │         └──────────────────┘                        │  │
│  └───────────────────────────────────────────────────┘  │
│                                                           │
│  ┌───────────────────────────────────────────────────┐  │
│  │         MONITORING (Prometheus + Loki)             │  │
│  │                                                     │  │
│  │  ┌──────────────┐  ┌──────────────┐             │  │
│  │  │ Prometheus   │  │    Loki      │             │  │
│  │  │ (Metrics)    │  │  (Logs)      │             │  │
│  │  └──────┬───────┘  └──────┬───────┘             │  │
│  │         │                  │                       │  │
│  │         └──────────┬───────┘                       │  │
│  │                    ▼                                 │  │
│  │         ┌──────────────────┐                        │  │
│  │         │  Data Collector  │                        │  │
│  │         │  (Last 30 days)  │                        │  │
│  │         └──────────────────┘                        │  │
│  └───────────────────────────────────────────────────┘  │
│                                                           │
└─────────────────────────────────────────────────────────┘
```

## Data Flow

1. **Collection Phase** (Every 30 days):
   - Scans notifi-services repository
   - Queries Prometheus for metrics (last 30 days)
   - Queries Loki for logs (last 30 days)
   - Queries Tempo for distributed traces (last 30 days)
   - Formats data for TRM training

2. **Training Phase**:
   - Fine-tunes TRM 7M model on collected data
   - Saves checkpoints during training
   - Evaluates on test set

3. **Deployment Phase**:
   - Uploads model to MinIO
   - Updates Ollama model registry
   - Updates VLLM configuration
   - Agents can now use fine-tuned model

## Configuration

### Environment Variables

Set in `k8s/kustomize/base/configmap.yaml`:

```yaml
data:
  NOTIFI_SERVICES_PATH: "/workspace/notifi/repos/notifi-services"
  PROMETHEUS_URL: "http://prometheus.monitoring.svc:9090"
  LOKI_URL: "http://loki.monitoring.svc:3100"
  DATA_DAYS: "30"
  EPOCHS: "50000"
  LR: "1e-4"
```

### Training Parameters

Adjust in `k8s/kustomize/pro/patch-configmap.yaml` for production:

```yaml
data:
  EPOCHS: "100000"  # More epochs
  GLOBAL_BATCH_SIZE: "256"  # Larger batches
```

## Monitoring

### Flyte Dashboard

- **URL**: `http://flyte.ml-platform.svc.forge.remote:81`
- **Project**: `homelab`
- **Domain**: `production`
- **Workflow**: `trm_finetuning_workflow`

### Check Execution Status

```bash
# List recent executions
flytectl get execution \
  --project homelab \
  --domain production \
  --limit 10

# Get execution details
flytectl get execution \
  --project homelab \
  --domain production \
  <execution-id>

# View logs
flytectl get execution-logs \
  --project homelab \
  --domain production \
  <execution-id>
```

## Troubleshooting

### Data Collection Issues

**Problem**: Cannot access Prometheus/Loki

**Solution**:
```bash
# Verify services are accessible
kubectl get svc -n monitoring prometheus loki
kubectl get svc -n tempo tempo

# Test connectivity
kubectl run -it --rm test-pod --image=curlimages/curl --restart=Never -- \
  curl http://prometheus.monitoring.svc:9090/api/v1/status/config

# Test Tempo
kubectl run -it --rm test-pod --image=curlimages/curl --restart=Never -- \
  curl http://tempo.tempo.svc:3200/api/search?tags=service.name=agent-sre
```

### Training Issues

**Problem**: GPU not available

**Solution**:
```bash
# Check GPU nodes
kubectl get nodes -l accelerator=nvidia-tesla-v100

# Check GPU resources
kubectl describe node <gpu-node-name> | grep nvidia.com/gpu
```

**Problem**: Out of memory

**Solution**: Increase memory request in `flyte_workflow.py`:
```python
@task(
    requests=Resources(cpu="8", mem="64Gi", gpu="1"),  # Increase mem
    ...
)
```

### Model Deployment Issues

**Problem**: Cannot upload to MinIO

**Solution**:
```bash
# Verify MinIO credentials
kubectl get secret -n ml-platform trm-finetune-secrets

# Test MinIO access
kubectl run -it --rm test-minio --image=minio/mc --restart=Never -- \
  mc alias set local http://minio.ml-platform.svc:9000 <access-key> <secret-key>
```

## Integration with Agents

After model is trained and deployed:

### 1. Update Ollama

```bash
# Import model
kubectl exec -n ollama deployment/ollama -- \
  ollama import /path/to/model

# Verify
kubectl exec -n ollama deployment/ollama -- \
  ollama list | grep trm-homelab
```

### 2. Update Agent Configuration

```yaml
# In agent deployment
env:
  - name: TRM_MODEL_NAME
    value: "trm-homelab"
  - name: OLLAMA_URL
    value: "http://ollama.ollama.svc:11434"
```

### 3. Use in Agent Code

```python
# In agent code
import httpx

async def use_trm_model(problem: str):
    response = await httpx.post(
        "http://ollama.ollama.svc:11434/api/generate",
        json={
            "model": "trm-homelab",
            "prompt": problem,
            "stream": False
        }
    )
    return response.json()["response"]
```

## Schedule Customization

To change the schedule (e.g., every 15 days instead of 30):

1. Edit `src/flyte_workflow.py`:
```python
@workflow(
    schedule=CronSchedule(
        schedule="0 2 */15 * *",  # Every 15 days at 2 AM
        kickoff_time_input_arg="trigger_time",
    )
)
```

2. Update data collection days:
```python
days: int = 15,  # Collect last 15 days
```

3. Re-register workflow:
```bash
pyflyte register src/flyte_workflow.py \
  --project homelab \
  --domain production \
  --image localhost:5001/trm-finetune:latest
```

## Cost Optimization

- **GPU Usage**: Training runs only during scheduled time (1st of month)
- **Data Collection**: Efficient queries with time ranges
- **Model Storage**: Old models can be archived after N versions
- **Batch Size**: Adjust `GLOBAL_BATCH_SIZE` based on GPU memory

## Next Steps

1. ✅ Deploy pipeline
2. ✅ Register workflow
3. ✅ Verify scheduled execution
4. ⏳ Monitor first training run
5. ⏳ Evaluate model performance
6. ⏳ Integrate with agents
7. ⏳ Monitor agent performance with fine-tuned model

