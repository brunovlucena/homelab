# 🏗️ SRE Infrastructure Architecture

## Overview

This document describes the MCP-based architecture for the SRE observability infrastructure, enabling intelligent query and analysis of metrics, logs, and traces.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                                                             │
│                      🌐 Communication Flow                                  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

Slack User
    │
    │ (Slack API)
    ▼
┌─────────────────────────────────────────┐
│     jamie-sre-chatbot (Slack Bot)       │
│     Port: 8080                          │
│                                         │
│  🧠 LLM Brain (Ollama llama3.2:3b)     │
│  📊 LangChain Agent with Tools          │
│                                         │
│  Capabilities:                          │
│  • Chat interface via Slack             │
│  • Tool calling via LangChain           │
│  • Context management                   │
│  • REST API endpoints                   │
└─────────────────────────────────────────┘
    │
    │ (Can use MCP or direct API)
    │
    ├─────────────────────┬──────────────────────────┐
    │                     │                          │
    │ (MCP Protocol)      │ (Direct REST API)        │
    ▼                     ▼                          │
┌─────────────────┐  ┌──────────────────┐          │
│ agent-sre-mcp-  │  │   agent-sre      │◄─────────┘
│    server       │  │   API            │
│  Port: 30120    │  │   Port: 8080     │
│                 │  │                  │
│ 🔧 MCP Tools:   │  │ 🛠️  API Routes:  │
│ • query_metrics │  │ • /api/query_    │
│ • query_logs    │  │   metrics        │
│ • query_traces  │  │ • /api/query_    │
│ • query_grafana │  │   logs           │
│ • sre_chat      │  │ • /api/query_    │
│ • analyze_logs  │  │   traces         │
│ • incident_resp │  │ • /api/query_    │
│ • golden_       │  │   grafana        │
│   signals       │  │ • /chat          │
└─────────────────┘  │ • /analyze-logs  │
         │           │ • /incident-     │
         │           │   response       │
         │           │ • /monitoring-   │
         │           │   advice         │
         │           └──────────────────┘
         │                     │
         │ (Proxy to API)      │ (Direct queries)
         └─────────┬───────────┘
                   │
                   ▼
         ┌─────────────────────┐
         │  Observability Stack│
         └─────────────────────┘
                   │
    ┌──────────────┼──────────────┬──────────────┐
    │              │              │              │
    ▼              ▼              ▼              ▼
┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐
│Prometheus│  │  Loki   │  │  Tempo  │  │ Grafana │
│ :9090    │  │ :3100   │  │ :3100   │  │ :3000   │
└─────────┘  └─────────┘  └─────────┘  └─────────┘
```

## Components

### 1. agent-sre (Main API)

**Location**: `@agent-sre/deployments/agent/`

**Purpose**: Core SRE agent with LangGraph and Ollama integration

**Features**:
- 🧠 LLM-powered analysis using Ollama
- 📊 Direct queries to observability stack
- 🔄 LangGraph state management
- 🎯 Multiple task types (logs, incident, monitoring, performance)

**API Endpoints**:
- `POST /chat` - General SRE chat
- `POST /analyze-logs` - Log analysis
- `POST /incident-response` - Incident guidance
- `POST /monitoring-advice` - Monitoring recommendations
- `POST /api/query_metrics` - Query Prometheus
- `POST /api/query_logs` - Query Loki
- `POST /api/query_traces` - Query Tempo
- `POST /api/query_grafana` - Query Grafana
- `POST /webhook/alert` - Alertmanager webhook
- `GET /health` - Health check
- `GET /ready` - Readiness probe

**Environment Variables**:
```bash
OLLAMA_URL=http://192.168.0.16:11434
MODEL_NAME=llama3.2:3b
SERVICE_NAME=sre-agent
PROMETHEUS_URL=http://prometheus.prometheus:9090
LOKI_URL=http://loki.loki:3100
TEMPO_URL=http://tempo.tempo:3100
GRAFANA_URL=https://grafana.lucena.cloud
GRAFANA_API_KEY=glsa_...
LOGFIRE_TOKEN_SRE_AGENT=...
LANGSMITH_API_KEY=...
```

### 2. agent-sre-mcp-server (MCP Protocol Layer)

**Location**: `@agent-sre/deployments/mcp-server/`

**Purpose**: Thin MCP protocol layer that forwards requests to agent-sre API

**Features**:
- 🔌 MCP JSON-RPC 2.0 protocol
- 🔧 Exposes SRE tools via MCP
- 🔄 Proxies requests to agent-sre API
- 📡 SSE support for real-time communication

**MCP Tools**:
- `sre_chat` - General SRE consultation
- `analyze_logs` - Log analysis
- `incident_response` - Incident guidance
- `monitoring_advice` - Monitoring recommendations
- `check_golden_signals` - Golden signals monitoring
- `query_prometheus` - Legacy PromQL queries
- **`query_metrics`** - 📊 NEW: Query Prometheus via agent-sre API
- **`query_logs`** - 📝 NEW: Query Loki via agent-sre API
- **`query_traces`** - 🔍 NEW: Query Tempo via agent-sre API
- **`query_grafana`** - 📈 NEW: Query Grafana via agent-sre API
- `get_pod_logs` - Kubernetes pod logs
- `query_grafana_mcp` - Advanced Grafana queries
- `query_prometheus_mcp` - Advanced Prometheus queries

**Environment Variables**:
```bash
AGENT_SERVICE_URL=http://sre-agent-service:8080
MCP_HOST=0.0.0.0
MCP_PORT=30120
```

### 3. jamie-sre-chatbot (Slack Interface)

**Location**: `@jamie/src/slack-bot/`

**Purpose**: Slack bot with LLM brain and intelligent tool calling

**Features**:
- 🤖 Slack integration (mentions, DMs, slash commands)
- 🧠 LLM brain using Ollama
- 🔧 LangChain agent with tool calling
- 📊 REST API for programmatic access
- 🎯 Context-aware conversations
- 🔀 Can use MCP or direct API

**LangChain Tools** (Call agent-sre directly via API):
- `check_golden_signals` - Monitor golden signals
- `query_prometheus` - Execute PromQL queries
- `query_grafana` - Query Grafana
- `analyze_logs` - Analyze log data
- `incident_response` - Get incident guidance
- `monitoring_advice` - Get monitoring advice
- `get_agent_status` - Check agent health

**Slack Commands**:
- `/jamie-help` - Show help message
- `/jamie-status` - Check Agent-SRE status
- `@Jamie <question>` - Mention in channel
- Direct messages - Private chat

**REST API Endpoints**:
- `POST /api/chat` - Chat with Jamie
- `POST /api/golden-signals` - Check golden signals
- `POST /api/prometheus/query` - Query Prometheus
- `POST /api/grafana/query` - Query Grafana
- `POST /api/analyze-logs` - Analyze logs
- `POST /api/incident-response` - Get incident guidance
- `POST /api/monitoring-advice` - Get monitoring advice
- `GET /health` - Health check
- `GET /ready` - Readiness probe

**Environment Variables**:
```bash
SLACK_BOT_TOKEN=xoxb-...
SLACK_SIGNING_SECRET=...
SLACK_APP_TOKEN=xapp-...
AGENT_SRE_URL=http://sre-agent-service.agent-sre:8080
AGENT_SRE_MCP_URL=http://sre-agent-mcp-server-service.agent-sre:30120
OLLAMA_URL=http://192.168.0.16:11434
MODEL_NAME=llama3.2:3b
API_HOST=0.0.0.0
API_PORT=8080
```

## Communication Patterns

### Pattern 1: Slack → Jamie → Agent-SRE API (Direct)

```
User: "Check the golden signals for homepage"
  │
  ▼
Jamie Slack Bot
  │ (Interprets using LLM)
  │ (Selects tool: check_golden_signals)
  │
  ▼ (REST API call)
Agent-SRE API
  │ (Queries Prometheus)
  │
  ▼
Prometheus
  │
  ▼ (Returns metrics)
Agent-SRE API
  │
  ▼ (Formats response)
Jamie Slack Bot
  │
  ▼ (Sends to Slack)
User: "✅ Homepage: Latency 45ms, Traffic 120req/min, Errors 0.1%, Saturation 35%"
```

### Pattern 2: Slack → Jamie → MCP Server → Agent-SRE API

```
User: "Query logs: {app='api'} |= 'error'"
  │
  ▼
Jamie Slack Bot (via MCP client)
  │ (MCP JSON-RPC 2.0 call)
  │
  ▼ (MCP Protocol)
Agent-SRE MCP Server
  │ (Proxies to API)
  │
  ▼ (REST API call)
Agent-SRE API
  │ (Queries Loki)
  │
  ▼
Loki
  │
  ▼ (Returns logs)
Agent-SRE API
  │
  ▼ (Returns JSON)
Agent-SRE MCP Server
  │ (MCP response)
  │
  ▼
Jamie Slack Bot
  │ (Formats for Slack)
  │
  ▼
User: "📝 Found 23 error logs in the last hour..."
```

### Pattern 3: Direct API Access

```
External System
  │
  ▼ (HTTP POST)
Agent-SRE API
  │ (Queries observability)
  │
  ▼
Prometheus/Loki/Tempo/Grafana
  │
  ▼ (Returns data)
Agent-SRE API
  │
  ▼ (JSON response)
External System
```

## Data Flow

### Query Metrics Flow

```
1. User Request
   "What's the CPU usage for the homepage service?"

2. Jamie LLM Brain
   - Interprets intent: "Query metrics"
   - Extracts parameters: service="homepage", metric="CPU"
   - Generates PromQL: rate(container_cpu_usage_seconds_total{service="homepage"}[5m])

3. Tool Selection
   - Selects: query_prometheus or query_metrics

4. Execution Path A (Direct API):
   jamie → agent-sre API → Prometheus → Response

5. Execution Path B (MCP):
   jamie → agent-sre-mcp-server → agent-sre API → Prometheus → Response

6. Response Formatting
   - Jamie formats response for Slack
   - Adds context and recommendations
   - Sends to user
```

## Deployment

### Kubernetes Resources

All services are deployed in the `agent-sre` namespace:

```yaml
# agent-sre namespace
apiVersion: v1
kind: Namespace
metadata:
  name: agent-sre

# Services
- sre-agent-service:8080 (agent-sre API)
- sre-agent-mcp-server-service:30120 (MCP server)
- jamie-slack-bot-service:8080 (Jamie bot)

# NodePort Services (for external access)
- sre-agent-nodeport:31081
- sre-agent-mcp-server-nodeport:31120
- jamie-slack-bot-nodeport:31082
```

### Build & Deploy

```bash
# Build all images
cd @agent-sre
make build-all

# Push to registry
make push-all

# Deploy to Kubernetes
make deploy-all

# Or deploy individually
make deploy-agent
make deploy-mcp
cd @jamie
make deploy
```

## Monitoring

### Health Checks

All services expose health and readiness endpoints:

```bash
# agent-sre API
curl http://sre-agent-service:8080/health
curl http://sre-agent-service:8080/ready

# agent-sre-mcp-server
curl http://sre-agent-mcp-server-service:30120/health
curl http://sre-agent-mcp-server-service:30120/ready

# jamie-slack-bot
curl http://jamie-slack-bot-service:8080/health
curl http://jamie-slack-bot-service:8080/ready
```

### Observability

All services are instrumented with:
- **Logfire**: Distributed tracing
- **LangSmith**: LLM observability
- **Prometheus**: Metrics (via ServiceMonitor)
- **Structured Logging**: JSON logs

## Security

### Secrets Management

Secrets are stored in Kubernetes secrets:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: agent-sre-secrets
  namespace: agent-sre
type: Opaque
data:
  GRAFANA_API_KEY: base64-encoded
  LOGFIRE_TOKEN_SRE_AGENT: base64-encoded
  LANGSMITH_API_KEY: base64-encoded

---
apiVersion: v1
kind: Secret
metadata:
  name: jamie-secrets
  namespace: jamie
type: Opaque
data:
  SLACK_BOT_TOKEN: base64-encoded
  SLACK_SIGNING_SECRET: base64-encoded
  SLACK_APP_TOKEN: base64-encoded
```

### Network Policies

Services are protected with NetworkPolicies:
- Only Alertmanager can send webhooks to agent-sre
- Only jamie can call agent-sre-mcp-server
- Only authenticated requests to observability stack

## Testing

### Test MCP Server

```bash
# List tools
curl -X POST http://localhost:30120/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}'

# Call query_metrics tool
curl -X POST http://localhost:30120/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc":"2.0",
    "id":2,
    "method":"tools/call",
    "params":{
      "name":"query_metrics",
      "arguments":{"query":"up"}
    }
  }'
```

### Test Jamie Slack Bot

```bash
# Chat API
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message":"Check the golden signals for homepage"}'

# Golden signals API
curl -X POST http://localhost:8080/api/golden-signals \
  -H "Content-Type: application/json" \
  -d '{"service_name":"homepage","namespace":"default"}'
```

### Test Agent-SRE API

```bash
# Query metrics
curl -X POST http://localhost:8080/api/query_metrics \
  -H "Content-Type: application/json" \
  -d '{"query":"up"}'

# Query logs
curl -X POST http://localhost:8080/api/query_logs \
  -H "Content-Type: application/json" \
  -d '{"query":"{app=\"api\"} |= \"error\"","limit":100}'

# Query traces
curl -X POST http://localhost:8080/api/query_traces \
  -H "Content-Type: application/json" \
  -d '{"trace_id":"abc123def456"}'
```

## Troubleshooting

### MCP Server Not Responding

```bash
# Check pod status
kubectl get pods -n agent-sre -l app=sre-agent-mcp-server

# Check logs
kubectl logs -n agent-sre -l app=sre-agent-mcp-server -f

# Check agent-sre connectivity
kubectl exec -it -n agent-sre <mcp-server-pod> -- curl http://sre-agent-service:8080/health
```

### Jamie Not Responding in Slack

```bash
# Check pod status
kubectl get pods -n jamie -l app=jamie-slack-bot

# Check logs
kubectl logs -n jamie -l app=jamie-slack-bot -f

# Test Slack connectivity
kubectl exec -it -n jamie <jamie-pod> -- curl https://slack.com/api/api.test
```

### Agent-SRE API Errors

```bash
# Check pod status
kubectl get pods -n agent-sre -l app=sre-agent

# Check logs
kubectl logs -n agent-sre -l app=sre-agent -f

# Check observability stack connectivity
kubectl exec -it -n agent-sre <agent-pod> -- curl http://prometheus.prometheus:9090/-/healthy
kubectl exec -it -n agent-sre <agent-pod> -- curl http://loki.loki:3100/ready
```

## Future Enhancements

- [ ] RAG integration for runbook search
- [ ] Automated remediation actions
- [ ] Multi-cluster support
- [ ] Cost tracking and optimization
- [ ] Feedback loop for continuous learning
- [ ] More sophisticated golden signals analysis
- [ ] Predictive alerting using ML
- [ ] Integration with more observability tools

## References

- [MCP Protocol Specification](https://spec.modelcontextprotocol.io/)
- [LangChain Documentation](https://python.langchain.com/)
- [LangGraph Documentation](https://langchain-ai.github.io/langgraph/)
- [Ollama Documentation](https://ollama.ai/docs)
- [Prometheus Query API](https://prometheus.io/docs/prometheus/latest/querying/api/)
- [Loki Query API](https://grafana.com/docs/loki/latest/api/)
- [Tempo API](https://grafana.com/docs/tempo/latest/api_docs/)

---

**Built with ❤️ for SRE automation**

