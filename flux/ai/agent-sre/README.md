# 🤖 Agent-SRE

**AI-Powered SRE Health Report Generator for Homelab**

Automated SRE health report generation using fine-tuned FunctionGemma 270M model with MLX-LM framework, integrated with Prometheus metrics and observability stack.

## 🎯 Overview

This agent generates comprehensive SRE health reports by:
- Querying Prometheus record rules for pre-computed health metrics
- Analyzing Loki, Prometheus, and infrastructure health
- Generating structured reports using fine-tuned FunctionGemma 270M model
- Supporting MLX-LM framework for efficient inference on Apple Silicon
- Evaluating deepagents library for complex reasoning tasks

## 🏗️ Architecture

```
┌──────────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│  Prometheus  │───▶│   Metrics    │───▶│   SRE Agent  │───▶│    Report    │
│ Record Rules │    │  Collector   │    │  (Function   │    │  Generator   │
│              │    │              │    │   Gemma)     │    │              │
└──────────────┘    └──────────────┘    └──────────────┘    └──────────────┘
                                                                    │
                                                                    ▼
                                                            ┌──────────────┐
                                                            │   Grafana    │
                                                            │   Dashboard  │
                                                            └──────────────┘
```

## 📋 Quick Start

```bash
# Install dependencies
make install

# Run locally
make run-agent

# Generate health report
make generate-report COMPONENT=loki
```

## 🔧 Model Selection & Framework

### FunctionGemma 270M
- **Size**: 270M parameters (lightweight)
- **Purpose**: Function calling and structured output
- **Suitability**: ✅ Excellent for SRE report generation
- **MLX Support**: ✅ Available via `mlx-community/functiongemma-270m-it-bf16`

### MLX-LM Framework
- **Purpose**: Efficient training/inference on Apple Silicon
- **Integration**: Direct support for FunctionGemma models
- **Benefits**: Optimized for M1/M2/M3 chips

### EXO Framework
- **Status**: Researching integration
- **Purpose**: Fine-tuning pipeline optimization
- **Note**: May require custom integration

### DeepAgents Library (Langchain-AI)
- **Purpose**: Complex reasoning and multi-agent coordination
- **Evaluation**: Testing compatibility with FunctionGemma
- **GitHub**: https://github.com/langchain-ai/deepagents

## 📁 Project Structure

```
agent-sre/
├── docs/
│   ├── ARCHITECTURE.md       # System architecture
│   ├── MODEL_SELECTION.md    # Model comparison and selection
│   └── FINE_TUNING.md        # Fine-tuning guide
├── src/
│   ├── sre_agent/            # Main agent logic
│   ├── report_generator/     # Report generation
│   └── metrics_collector/    # Prometheus metrics collection
├── k8s/
│   └── kustomize/            # Kubernetes manifests
├── tests/
│   ├── unit/
│   └── integration/
├── training/
│   ├── data/                 # Fine-tuning datasets
│   └── scripts/              # Training scripts
├── Makefile
└── README.md
```

## 🚀 Deployment

### Prerequisites

- Prometheus with record rules deployed
- MLX-LM framework installed (for local development)
- Ollama or compatible LLM endpoint
- Kubernetes cluster with Knative Lambda

### Deploy to Homelab

```bash
# Build images
make build

# Push to registry
make push

# Deploy to Kubernetes
make deploy-pro
```

## 📊 Monitoring

The agent exposes metrics at `/metrics`:
- `sre_agent_reports_generated_total`
- `sre_agent_report_generation_duration_seconds`
- `sre_agent_metrics_query_errors_total`

## 🔬 Fine-Tuning

See [docs/FINE_TUNING.md](docs/FINE_TUNING.md) for detailed fine-tuning instructions using:
- MLX-LM framework
- EXO (if integrated)
- Custom SRE report datasets

## 📝 License

See LICENSE file.

