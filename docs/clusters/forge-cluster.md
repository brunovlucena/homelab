# Forge Cluster 🤖

> **Part of**: [Homelab Documentation](../README.md) → Clusters  
> **Last Updated**: November 7, 2025

---

## Overview

**Platform**: k3s (lightweight Kubernetes)  
**Hardware**: NVIDIA GPU Server (x86_64)  
**Architecture**: x86_64  
**Purpose**: AI training & inference, GPU-accelerated ML workloads 🤖  
**Nodes**: 8 (1 control-plane + 7 GPU workers)  
**Network**: 10.248.0.0/16

---

## Node Architecture (GPU-Optimized)

| Node | Role | GPUs | Workloads |
|------|------|------|-----------|
| 1 | control-plane | - | k3s control plane |
| 2 | platform | - | Flux, monitoring, logging |
| 3 | training-primary | 2× A100 | PyTorch training, fine-tuning |
| 4 | training-secondary | 2× A100 | Distributed training |
| 5 | inference | 2× A100 | VLLM (Llama 3.1 70B), Ollama |
| 6 | ml-platform | - | Flyte, JupyterHub, MLflow |
| 7 | data-ml | - | MinIO, model registry, datasets |
| 8 | observability | - | NVIDIA DCGM Exporter, metrics |

---

## Key Components

### VLLM (High-Performance LLM Inference)

**Purpose**: Serve large language models with OpenAI-compatible API

**Deployment**:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: vllm
  namespace: ml-inference
  annotations:
    multicluster.linkerd.io/export: "true"
spec:
  replicas: 1
  template:
    spec:
      nodeSelector:
        role: inference
        gpu-type: nvidia
      containers:
      - name: vllm
        image: vllm/vllm-openai:latest
        args:
        - "--model=meta-llama/Meta-Llama-3.1-70B-Instruct"
        - "--tensor-parallel-size=2"
        - "--max-model-len=4096"
        resources:
          limits:
            nvidia.com/gpu: "2"
        ports:
        - containerPort: 8000
          name: http
```

**Access from Studio**:
```python
import openai

client = openai.OpenAI(
    api_key="EMPTY",
    base_url="http://vllm.ml-inference.svc.forge.remote:8000/v1"
)

response = client.chat.completions.create(
    model="meta-llama/Meta-Llama-3.1-70B-Instruct",
    messages=[{"role": "user", "content": "Hello!"}]
)
```

---

### Ollama (Small Language Models)

**Purpose**: Lightweight LLM serving for faster inference

**Models**: Llama 2 7B, Mistral 7B, CodeLlama

**Deployment**:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ollama
  namespace: ml-inference
spec:
  template:
    spec:
      nodeSelector:
        role: inference
      containers:
      - name: ollama
        image: ollama/ollama:latest
        ports:
        - containerPort: 11434
```

---

### PyTorch (Model Training)

**Purpose**: Train and fine-tune machine learning models

**Example Training Job**:
```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: training-job
  namespace: ml-training
spec:
  template:
    spec:
      nodeSelector:
        role: training
        gpu-type: nvidia
      containers:
      - name: trainer
        image: pytorch/pytorch:latest
        command:
        - python
        - train.py
        resources:
          limits:
            nvidia.com/gpu: "2"
        env:
        - name: CUDA_VISIBLE_DEVICES
          value: "0,1"
```

---

### Flyte (ML Workflow Orchestration)

**Purpose**: Orchestrate complex ML pipelines

**Access**: http://flyte.ml-platform.svc.forge.remote:30081

**Example Workflow**:
```python
from flytekit import task, workflow

@task
def preprocess_data(data_path: str) -> str:
    # Data preprocessing
    return processed_path

@task(gpu="2")
def train_model(data_path: str) -> str:
    # Model training
    return model_path

@workflow
def training_pipeline(data_path: str) -> str:
    processed = preprocess_data(data_path=data_path)
    model = train_model(data_path=processed)
    return model
```

---

### JupyterHub (Interactive Notebooks)

**Purpose**: Interactive development for data scientists

**Access**: http://jupyterhub.ml-platform.svc.forge.remote:30102

**GPU Access**:
```python
# In Jupyter notebook
import torch

# Check GPU availability
print(torch.cuda.is_available())
print(torch.cuda.device_count())
print(torch.cuda.get_device_name(0))
```

---

### MinIO (Object Storage)

**Purpose**: Store models, datasets, and artifacts

**Access**: http://minio.data-ml.svc.forge.remote:30063

**Usage**:
```python
from minio import Minio

client = Minio(
    "minio.data-ml.svc.forge.remote:30063",
    access_key="minioadmin",
    secret_key="minioadmin",
    secure=False
)

# Upload model
client.fput_object(
    "models",
    "llama-2-7b-fine-tuned.bin",
    "/tmp/model.bin"
)
```

---

## Cross-Cluster Integration

Forge services are consumed by AI agents running on Studio cluster:

### Service Export

```yaml
# Export VLLM for cross-cluster access
apiVersion: v1
kind: Service
metadata:
  name: vllm
  namespace: ml-inference
  annotations:
    multicluster.linkerd.io/export: "true"
spec:
  ports:
  - port: 8000
    targetPort: 8000
  selector:
    app: vllm
```

### Service Pattern

```
Studio (AI Agents) → Linkerd → Forge (GPU Services)

service.namespace.svc.forge.remote:port
```

---

## GPU Management

### GPU Labels

```yaml
nodeSelector:
  gpu-type: nvidia
  nvidia.com/gpu: "true"
  nvidia.com/gpu.product: NVIDIA-A100-SXM4-40GB
```

### GPU Resource Requests

```yaml
resources:
  limits:
    nvidia.com/gpu: "2"  # Request 2 GPUs
  requests:
    cpu: "8000m"
    memory: "32Gi"
```

### NVIDIA Device Plugin

Automatically labels nodes with GPU capabilities:

```bash
# Check GPU nodes
kubectl --context=forge get nodes -l nvidia.com/gpu=true

# Check GPU allocations
kubectl --context=forge describe node forge-worker-3
```

---

## Use Cases

### 1. LLM Inference (VLLM)

Serve large language models:

```bash
# Deploy VLLM
kubectl --context=forge apply -f ml-inference/vllm/

# Test from Studio
curl http://vllm.ml-inference.svc.forge.remote:8000/v1/models
```

### 2. Model Training

Train custom models:

```bash
# Submit training job
kubectl --context=forge apply -f training-job.yaml

# Monitor
kubectl --context=forge logs -f job/training-job -n ml-training
```

### 3. ML Pipeline Orchestration

Run complex workflows:

```bash
# Submit Flyte workflow
flytectl --config ~/.flyte/config.yaml create execution \
  --project homelab \
  --domain production \
  --workflow training_pipeline
```

### 4. Interactive Development

Develop models interactively:

```bash
# Access JupyterHub
open http://forge.cluster:30102

# Start GPU-enabled notebook
# Select kernel with GPU access
```

---

## Resource Limits

### Per GPU Node

- **CPU**: 64 cores (AMD EPYC or Intel Xeon)
- **Memory**: 256GB RAM
- **GPU**: 2× NVIDIA A100 (40GB each)
- **Disk**: 2TB NVMe SSD
- **Network**: 10Gbps

### Cluster Total

- **CPU**: 384 cores
- **Memory**: 1.5TB
- **GPU**: 8× A100 (320GB total)
- **Disk**: 14TB
- **Power**: ~3kW

---

## Best Practices

### DO

- ✅ Use tensor parallelism for large models
- ✅ Monitor GPU utilization (DCGM)
- ✅ Use model registry (MinIO)
- ✅ Set resource limits
- ✅ Use job queues (Flyte)

### DON'T

- ❌ Don't overallocate GPUs
- ❌ Don't run non-ML workloads
- ❌ Don't ignore temperature monitoring
- ❌ Don't skip model versioning

---

## Monitoring

### GPU Metrics (DCGM Exporter)

```bash
# Check GPU utilization
kubectl --context=forge port-forward -n gpu-operator \
  svc/dcgm-exporter 9400:9400

# Query metrics
curl http://localhost:9400/metrics | grep gpu_utilization
```

### Grafana Dashboards

- **NVIDIA DCGM Dashboard**: GPU utilization, temperature, power
- **VLLM Dashboard**: Request rate, latency, throughput
- **Training Dashboard**: Job status, resource usage

---

## Cost Optimization

### GPU Sharing

Multiple small jobs can share GPUs:

```yaml
resources:
  limits:
    nvidia.com/gpu: "1"  # Share GPU
  requests:
    cpu: "4000m"
    memory: "16Gi"
```

### Time-zone Based Scheduling

Utilize GPUs 24/7 across time zones:

```yaml
# Phase 2: Schedule training jobs across regions
# Brazil (daytime) → Training
# US/Europe (nighttime for Brazil) → Inference
```

---

## Troubleshooting

### GPU Not Available

```bash
# Check NVIDIA driver
kubectl --context=forge exec -it gpu-pod -- nvidia-smi

# Check device plugin
kubectl --context=forge get pods -n gpu-operator
```

### Out of GPU Memory

```bash
# Check GPU memory
kubectl --context=forge exec -it vllm-pod -- nvidia-smi

# Reduce model size or batch size
# Use model quantization (4-bit, 8-bit)
```

### Slow Inference

```bash
# Check GPU utilization
nvidia-smi dmon -s u

# If low utilization:
# - Increase batch size
# - Use tensor parallelism
# - Optimize model configuration
```

---

## Related Documentation

- [Studio Cluster](studio-cluster.md) - AI agents consuming Forge services
- [AI Agent Architecture](../architecture/ai-agent-architecture.md)
- [AI Components](../architecture/ai-components.md)

---

**Last Updated**: November 7, 2025  
**Maintained by**: SRE Team (Bruno Lucena)

