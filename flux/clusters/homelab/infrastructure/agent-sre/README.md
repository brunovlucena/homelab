# 🤖 Agent SRE - Site Reliability Engineering Agent

A comprehensive SRE (Site Reliability Engineering) agent service built with LangGraph and LangChain, designed to provide intelligent assistance for monitoring, incident response, log analysis, and system reliability tasks.

## 📋 Overview

The Agent SRE service consists of two main components:

1. **Agent Service** - The core SRE agent with LangGraph state management
2. **MCP Server** - A thin protocol layer for Model Context Protocol communication

Both services work together to provide a robust, scalable SRE assistant that can be integrated into various workflows and tools.

## 🏗️ Architecture

```
┌─────────────────┐
│   MCP Client    │  (Claude Desktop, IDEs, etc.)
└────────┬────────┘
         │ MCP Protocol
         ↓
┌─────────────────┐
│   MCP Server    │  (Port 30120)
│   Thin Layer    │
└────────┬────────┘
         │ HTTP API
         ↓
┌─────────────────┐
│  Agent Service  │  (Port 8080)
│   LangGraph     │
│   + Ollama      │
└─────────────────┘
```

## ✨ Features

- **🔍 Log Analysis**: Intelligent analysis of system logs with pattern detection and root cause identification
- **🚨 Incident Response**: Guided incident response with best practices and communication templates
- **📊 Monitoring Advice**: Recommendations for monitoring strategies, metrics, and alerting
- **💬 General SRE Chat**: Interactive consultation on SRE topics and best practices
- **🏥 Health Checks**: Built-in health and readiness probes for Kubernetes
- **📈 Observability**: Integrated with Logfire and LangSmith for tracing and monitoring
- **🔄 State Management**: LangGraph-based workflow for complex reasoning and decision-making

## 🚀 Quick Start

### Running Locally with Docker Compose

```bash
cd flux/clusters/homelab/infrastructure/agent-sre
docker-compose up -d
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `OLLAMA_URL` | URL to Ollama server | `http://192.168.0.3:11434` |
| `MODEL_NAME` | LLM model to use | `bruno-sre:latest` |
| `AGENT_HOST` | Agent service host | `0.0.0.0` |
| `AGENT_PORT` | Agent service port | `8080` |
| `MCP_HOST` | MCP server host | `0.0.0.0` |
| `MCP_PORT` | MCP server port | `30120` |
| `LOGFIRE_TOKEN_SRE_AGENT` | Logfire token (optional) | - |
| `LANGSMITH_API_KEY` | LangSmith API key (optional) | - |

## 📚 API Endpoints

### Agent Service (Port 8080)

#### Health & Status
- `GET /health` - Liveness probe
- `GET /ready` - Readiness probe
- `GET /status` - Detailed agent status

#### Direct Agent API
- `POST /chat` - General SRE chat
- `POST /analyze-logs` - Analyze logs
- `POST /incident-response` - Incident response guidance
- `POST /monitoring-advice` - Monitoring recommendations

#### MCP Forwarding
- `POST /mcp/chat` - Chat via MCP server
- `POST /mcp/analyze-logs` - Log analysis via MCP
- `POST /mcp/incident-response` - Incident response via MCP
- `POST /mcp/monitoring-advice` - Monitoring advice via MCP

### MCP Server (Port 30120)

- `POST /mcp` - MCP JSON-RPC 2.0 endpoint
- `GET /mcp` - Server information
- `GET /health` - Health check
- `GET /ready` - Readiness check
- `GET /sse` - Server-Sent Events for real-time updates

## 🧪 Testing

### Running Tests Locally

```bash
# Install dependencies with uv
uv sync --frozen --extra dev

# Run all tests
uv run pytest tests/ -v

# Run with coverage
uv run pytest tests/ --cov=deployments --cov-report=term-missing

# Run specific test file
uv run pytest tests/test_agent_core.py -v

# Run linting
uv run flake8 deployments/ --max-line-length=120
uv run black --check deployments/
uv run isort --check-only deployments/
```

### Test Structure

```
tests/
├── __init__.py
├── conftest.py              # Shared fixtures and configuration
├── test_agent_core.py       # Core LangGraph agent tests
├── test_agent_service.py    # HTTP API service tests
└── test_mcp_server.py       # MCP server protocol tests
```

## 🔄 CI/CD Pipeline

The Agent SRE service uses a two-stage CI/CD pipeline:

### Stage 1: Tests (agent-sre-tests.yml)
- ✅ Runs linting (flake8, black, isort)
- ✅ Runs unit tests with pytest
- ✅ Generates coverage reports
- ✅ Comments on PRs with test results

### Stage 2: Images (agent-sre-images.yml)
- ✅ Waits for tests to pass
- ✅ Builds Docker images
- ✅ Pushes to GitHub Container Registry
- ✅ Runs security scans with Trivy
- ✅ Supports multi-arch (amd64, arm64)

### Workflow Status

[![🧪 Agent SRE Tests](https://github.com/brunovlucena/homelab/actions/workflows/agent-sre-tests.yml/badge.svg)](https://github.com/brunovlucena/homelab/actions/workflows/agent-sre-tests.yml)
[![🤖 Agent SRE Images CI/CD](https://github.com/brunovlucena/homelab/actions/workflows/agent-sre-images.yml/badge.svg)](https://github.com/brunovlucena/homelab/actions/workflows/agent-sre-images.yml)

## 🐳 Docker Images

Images are automatically built and pushed to GitHub Container Registry:

- **Development**: `ghcr.io/brunovlucena/agent-sre:dev`
- **Production**: `ghcr.io/brunovlucena/agent-sre:latest`

## 📦 Dependencies

Main dependencies:
- `langchain >= 0.3.0` - LangChain framework
- `langchain-ollama >= 0.2.0` - Ollama integration
- `langgraph >= 0.2.0` - State management
- `langsmith >= 0.1.0` - Tracing and monitoring
- `logfire >= 0.1.0` - Observability
- `aiohttp >= 3.9.0` - Async HTTP server
- `pydantic >= 2.0.0` - Data validation

Dev dependencies:
- `pytest >= 7.0.0` - Testing framework
- `pytest-cov >= 4.0.0` - Coverage reporting
- `flake8 >= 6.0.0` - Linting
- `black >= 23.0.0` - Code formatting
- `isort >= 5.12.0` - Import sorting

## 🔧 Development

### Code Style

This project follows these code style guidelines:
- **Line length**: 120 characters
- **Python version**: 3.11+
- **Formatter**: Black
- **Linter**: Flake8
- **Import sorter**: isort

### Pre-commit Setup

```bash
# Format code
uv run black deployments/

# Sort imports
uv run isort deployments/

# Run linter
uv run flake8 deployments/ --max-line-length=120
```

## 🚢 Deployment

### Kubernetes

The service is deployed to Kubernetes using Kustomize:

```bash
# Deploy to Kubernetes
kubectl apply -k deployments/agent/

# Check deployment status
kubectl get pods -n agent-sre

# View logs
kubectl logs -n agent-sre -l app=sre-agent
```

### Configuration

See the Kubernetes manifests in:
- `deployments/agent/k8s-agent.yaml`
- `deployments/mcp-server/k8s-mcp-server.yaml`

## 📈 Monitoring & Observability

The service integrates with:
- **Logfire**: Distributed tracing and logging
- **LangSmith**: LLM call tracing and debugging
- **Prometheus**: Metrics collection (via ServiceMonitor)
- **Grafana**: Dashboards and visualization

## 🤝 Contributing

1. Create a feature branch
2. Make your changes
3. Run tests locally: `uv run pytest tests/ -v`
4. Ensure linting passes: `uv run flake8 deployments/`
5. Submit a pull request

The CI/CD pipeline will automatically:
- Run tests on your PR
- Comment with test results
- Build images if tests pass
- Deploy to dev environment

## 📄 License

This project is part of the homelab infrastructure.

## 🔗 Related Documentation

- [LangGraph Documentation](https://langchain-ai.github.io/langgraph/)
- [LangChain Documentation](https://python.langchain.com/)
- [Ollama Documentation](https://ollama.ai/docs)
- [MCP Protocol Specification](https://modelcontextprotocol.io/)

---

**Maintained by**: Bruno Lucena  
**Last Updated**: 2025-10-09
