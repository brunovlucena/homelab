# 🔧 AI Components

> **Part of**: [AI Agent Architecture](ai-agent-architecture.md)  
> **Related**: [Agent Orchestration](agent-orchestration.md) | [Forge Cluster](../clusters/forge-cluster.md) | [Studio Cluster](../clusters/studio-cluster.md)  
> **Last Updated**: November 7, 2025

---

## Overview

This document provides detailed technical specifications for all AI components in the homelab architecture:

- [Small Language Models (SLMs)](#small-language-models-slms)
- [Knowledge Graph](#knowledge-graph)
- [Large Language Model (LLM)](#large-language-model-llm)
- [Workflow Orchestration](#workflow-orchestration)
- [Security & Authentication](#security--authentication)

---

## Small Language Models (SLMs)

**Purpose**: Specialized, lightweight models for specific tasks (code generation, data extraction, classification)

### Deployment

**Location**: [Forge Cluster](../clusters/forge-cluster.md) (GPU nodes)

**Configuration**:
```yaml
Technology: Ollama
Models:
  - Llama 3 (8B)
  - CodeLlama (7B-13B)
  - Mistral (7B)
Resources:
  gpu: 1 per model
  memory: 16-32GB RAM
Service: ollama.ml-inference.svc.forge.remote:11434
```

### Access Pattern

```python
# From Studio agent
import ollama

# Fast, specialized inference
response = ollama.generate(
    model="codellama",
    prompt="Generate Kubernetes manifest for nginx",
    endpoint="http://ollama.ml-inference.svc.forge.remote:11434"
)
```

### Use Cases

| Task | Model | Latency | Accuracy |
|------|-------|---------|----------|
| Code generation | CodeLlama | <100ms | 85% |
| Classification | Llama 3 | <50ms | 92% |
| Data extraction | Mistral | <100ms | 88% |
| Sentiment analysis | Llama 3 | <50ms | 90% |
| Entity recognition | Mistral | <75ms | 87% |

### Performance Characteristics

- **Throughput**: 200-500 tokens/second
- **Concurrency**: 4-8 parallel requests per GPU
- **Memory**: 8-16GB per model instance
- **Cold start**: 2-5 seconds
- **Response time**: 50-200ms (typical)

---

## Knowledge Graph

**Purpose**: Centralized knowledge storage, relationships, and context for AI agents

### Deployment

**Location**: [Studio Cluster](../clusters/studio-cluster.md) (data nodes)

**Configuration**:
```yaml
Technology: LanceDB (vector database)
Storage: MinIO (object storage for embeddings)
Indexing: FAISS/HNSW (similarity search)
Service: lancedb.ml-storage.svc.cluster.local:8000
```

### Schema

```yaml
Collections:
  - homelab-docs:
      type: documentation
      embeddings: all-MiniLM-L6-v2
      chunks: 512 tokens
      metadata: [file, section, date, author]
      size: ~10k documents
  
  - incident-history:
      type: operational
      embeddings: all-MiniLM-L6-v2
      metadata: [severity, cluster, service, resolution]
      retention: 365 days
      size: ~1k incidents
  
  - code-snippets:
      type: technical
      embeddings: code-search-net
      metadata: [language, framework, tested]
      validation: automated testing
      size: ~5k snippets
  
  - team-knowledge:
      type: collaborative
      embeddings: all-MiniLM-L6-v2
      metadata: [team, project, stakeholders]
      access: RBAC-controlled
      size: ~2k entries
```

### RAG Pipeline

```python
# Retrieve relevant context from Knowledge Graph
from lancedb import LanceDB

db = LanceDB("lancedb.ml-storage.svc.cluster.local:8000")
table = db.open_table("homelab-docs")

# Semantic search
results = table.search("How to deploy to Studio cluster?") \
    .limit(5) \
    .to_list()

# Augment prompt with context
context = "\n".join([r["text"] for r in results])
augmented_prompt = f"Context:\n{context}\n\nQuestion: {user_query}"
```

### Embedding Strategy

```python
# Document ingestion pipeline
from sentence_transformers import SentenceTransformer

model = SentenceTransformer('all-MiniLM-L6-v2')

def embed_document(doc: str) -> dict:
    # 1. Chunk document
    chunks = chunk_text(doc, max_tokens=512, overlap=50)
    
    # 2. Generate embeddings
    embeddings = model.encode(chunks)
    
    # 3. Store with metadata
    return {
        "text": chunks,
        "embeddings": embeddings,
        "metadata": extract_metadata(doc)
    }
```

### Performance Characteristics

- **Search latency**: 10-50ms
- **Embedding latency**: 5-20ms per chunk
- **Storage**: 384 dimensions (all-MiniLM-L6-v2)
- **Index**: HNSW with M=16, ef_construction=200
- **Recall@5**: >95%

---

## Large Language Model (LLM)

**Purpose**: Complex reasoning, code generation, analysis for difficult tasks

### Deployment

**Location**: [Forge Cluster](../clusters/forge-cluster.md) (inference nodes)

**Configuration**:
```yaml
Technology: VLLM (vLLM inference engine)
Model: Meta-Llama-3.1-70B-Instruct
Tensor Parallelism: 2 GPUs
Max Model Length: 8192 tokens
GPU Memory: 80GB (2× A100 40GB)
Service: vllm.ml-inference.svc.forge.remote:8000
```

### API Access

**OpenAI-Compatible API**:

```python
import openai

client = openai.OpenAI(
    api_key="EMPTY",
    base_url="http://vllm.ml-inference.svc.forge.remote:8000/v1"
)

response = client.chat.completions.create(
    model="meta-llama/Meta-Llama-3.1-70B-Instruct",
    messages=[
        {"role": "system", "content": "You are an expert SRE."},
        {"role": "user", "content": augmented_prompt}
    ],
    temperature=0.7,
    max_tokens=2000
)
```

### Performance Characteristics

- **Throughput**: 20-30 tokens/second (with tensor parallelism)
- **Latency**: 1-3 seconds (first token), 30-50ms/token (subsequent)
- **Batch size**: 32-64 requests
- **Context window**: 8192 tokens
- **GPU utilization**: 80-95%
- **Memory per request**: ~2GB

### Model Selection Logic

```python
def select_model(query: str, complexity: str) -> str:
    """
    Choose between SLM and LLM based on task complexity
    """
    if complexity == "low":
        # Quick tasks: classification, simple extraction
        return "ollama/llama3:8b"
    
    elif complexity == "medium":
        # Code generation, analysis
        return "ollama/codellama:13b"
    
    elif complexity == "high":
        # Complex reasoning, multi-step planning
        return "vllm/llama-3.1-70b"
    
    else:
        # Default to medium
        return "ollama/codellama:13b"
```

### Cost Optimization

```yaml
Strategy:
  - SLM for 80% of tasks (fast, cheap)
  - LLM for 20% of tasks (powerful, expensive)

Estimated Costs:
  SLM: $0.001 per 1k tokens
  LLM: $0.01 per 1k tokens
  
Average Cost per Request:
  SLM: $0.0001 (100 tokens)
  LLM: $0.002 (200 tokens)
  
Savings: 10x cost reduction vs LLM-only
```

---

## Workflow Orchestration

**Purpose**: Complex, multi-step AI/ML pipelines and task graphs

### Deployment

**Location**: [Forge Cluster](../clusters/forge-cluster.md) (ml-platform nodes)

**Configuration**:
```yaml
Technology: Flyte (workflow orchestration)
Service: flyte.flyte.svc.forge.remote:81
Components:
  - FlyteAdmin: Control plane
  - FlytePropeller: Workflow executor
  - FlyteConsole: UI dashboard
Storage: MinIO (artifact storage)
```

### Pipeline Example

```python
from flytekit import task, workflow
import pandas as pd

@task
def embed_documents(docs: list[str]) -> list[float]:
    """Embed documents using SLM"""
    ollama = OllamaClient("ollama.ml-inference.svc.forge.remote")
    return ollama.embed(docs)

@task
def index_knowledge_graph(embeddings: list[float]):
    """Index in LanceDB"""
    lancedb = LanceDBClient("lancedb.ml-storage.svc.cluster.local")
    lancedb.insert(embeddings)

@task
def train_classifier(data: pd.DataFrame):
    """Fine-tune SLM on GPU"""
    return train_on_forge(data)

@workflow
def knowledge_pipeline():
    """End-to-end knowledge ingestion pipeline"""
    docs = extract_docs()
    embeddings = embed_documents(docs)
    index_knowledge_graph(embeddings)
```

### Use Cases

| Pipeline | Tasks | Duration | Frequency |
|----------|-------|----------|-----------|
| Knowledge Ingestion | 5 tasks | 10-20 min | Daily |
| Model Fine-tuning | 8 tasks | 2-4 hours | Weekly |
| Batch Inference | 3 tasks | 5-15 min | Hourly |
| Data Processing | 6 tasks | 15-30 min | Daily |

---

## Security & Authentication

### External Secrets Operator Integration

**Purpose**: Secret management for AI agents via GitHub repository secrets

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: agent-bruno-api-key
  namespace: ai-agents
spec:
  secretStoreRef:
    name: github-backend
    kind: ClusterSecretStore
  target:
    name: agent-bruno-api-key
    creationPolicy: Owner
  data:
    - secretKey: api-key
      remoteRef:
        key: AGENT_BRUNO_API_KEY
```

### RBAC Configuration

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: agent-bruno-role
  namespace: ai-agents
rules:
- apiGroups: [""]
  resources: ["pods", "services"]
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["get"]
- apiGroups: ["batch"]
  resources: ["jobs"]
  verbs: ["create", "get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: agent-bruno-binding
  namespace: ai-agents
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: agent-bruno-role
subjects:
- kind: ServiceAccount
  name: agent-bruno-sa
  namespace: ai-agents
```

### mTLS via Linkerd

**Automatic encryption** for all agent-to-service communication:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: ollama
  namespace: ml-inference
  annotations:
    # Linkerd automatically enables mTLS
    linkerd.io/inject: enabled
spec:
  ports:
  - port: 11434
    targetPort: 11434
  selector:
    app: ollama
```

### Access Control Matrix

| Component | Agent Access | User Access | External Access |
|-----------|--------------|-------------|-----------------|
| Ollama SLM | ✅ (all agents) | ❌ | ❌ |
| VLLM LLM | ✅ (all agents) | ❌ | ❌ |
| Knowledge Graph | ✅ (RBAC) | ✅ (read-only) | ❌ |
| Flyte Workflows | ✅ (approved) | ✅ (UI) | ❌ |
| MCP Server | ✅ (all agents) | ✅ (all users) | ❌ |

---

## Related Documentation

- [🤖 AI Architecture Overview](ai-agent-architecture.md)
- [🎯 Agent Orchestration](agent-orchestration.md)
- [🌐 AI Connectivity](ai-connectivity.md)
- [📊 MCP Observability](mcp-observability.md)
- [⚙️ Forge Cluster](../clusters/forge-cluster.md)
- [🎯 Studio Cluster](../clusters/studio-cluster.md)

---

**Last Updated**: November 7, 2025  
**Maintained by**: SRE Team (Bruno Lucena)

