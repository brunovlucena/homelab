# 🤖 SRE Agent with LangGraph

Production-ready SRE Agent with LangGraph state management and MCP Server capabilities for distributed AI-powered operations.

## 🚀 Quick Start

### Local Development

```bash
# Start all services with docker-compose
docker-compose up -d

# Or start individual services
cd deployments/agent && python agent.py
cd deployments/mcp-server && python mcp_server.py
```

### Production Deployment

```bash
# Build and push images
make build-agent
make build-mcp-server

# Deploy to Kubernetes
kubectl apply -f deployments/agent/k8s-agent.yaml
kubectl apply -f deployments/mcp-server/k8s-mcp-server.yaml
```

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      MCP Server (Port 30120)                │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  MCP Protocol Handler (JSON-RPC 2.0)                │   │
│  │  - initialize, tools/list, tools/call               │   │
│  └─────────────────────────────────────────────────────┘   │
│                           ↓                                 │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  Tool Forwarding Layer                              │   │
│  │  Maps MCP tools → Agent API endpoints              │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│                   SRE Agent (Port 8080)                     │
│  ┌─────────────────────────────────────────────────────┐   │
│  │         LangGraph State Management                  │   │
│  │  ┌──────────────────────────────────────────────┐  │   │
│  │  │  1. Analyze Node (LLM Analysis)              │  │   │
│  │  │  2. Generate Recommendations Node            │  │   │
│  │  │  3. Format Response Node                     │  │   │
│  │  └──────────────────────────────────────────────┘  │   │
│  │                     ↓                               │   │
│  │         Memory Checkpointer (Thread State)          │   │
│  └─────────────────────────────────────────────────────┘   │
│                           ↓                                 │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  Ollama LLM (ChatOllama)                            │   │
│  │  Model: bruno-sre:latest                            │   │
│  │  URL: http://192.168.0.3:11434                      │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## ✨ Features

### LangGraph State Management
- **State-based Workflows**: Structured agent execution with clear state transitions
- **Memory Checkpointer**: Conversation history and context preservation
- **Multi-Node Processing**: Analyze → Recommend → Format pipeline
- **Thread Management**: Isolated sessions with unique thread IDs

### MCP Server
- **Decoupled Architecture**: Thin protocol layer separate from agent logic
- **JSON-RPC 2.0**: Standard MCP protocol implementation
- **Tool Registry**: Expose agent capabilities as MCP tools
- **Health Monitoring**: Liveness and readiness probes

### SRE Capabilities
- **Log Analysis**: Parse and analyze logs with AI insights
- **Incident Response**: Guided incident management and remediation
- **Monitoring Advice**: Observability and alerting recommendations
- **Performance Analysis**: Bottleneck identification and optimization

## 📦 Components

### 1. SRE Agent (`deployments/agent/`)
- **core.py**: LangGraph agent implementation with state management
- **agent.py**: HTTP API server for direct agent access
- **Dockerfile**: Container image for agent service
- **k8s-agent.yaml**: Kubernetes deployment manifests

### 2. MCP Server (`deployments/mcp-server/`)
- **mcp_server.py**: MCP protocol handler and tool forwarding
- **core.py**: Shared logging and utilities
- **Dockerfile**: Container image for MCP server
- **k8s-mcp-server.yaml**: Kubernetes deployment manifests

## 🛠️ Technology Stack

**AI Framework:**
- LangGraph 0.2+ (State management)
- LangChain 0.3+ (LLM orchestration)
- ChatOllama (LLM interface)
- LangSmith (Tracing and monitoring)

**Observability:**
- Logfire (Distributed tracing)
- Prometheus (Metrics)
- Custom ServiceMonitors

**Infrastructure:**
- Kubernetes (Container orchestration)
- Docker (Containerization)
- Flux CD (GitOps deployment)

## 🔧 Configuration

### Environment Variables

**Agent Service:**
```bash
OLLAMA_URL=http://192.168.0.3:11434
MODEL_NAME=bruno-sre:latest
SERVICE_NAME=sre-agent
AGENT_HOST=0.0.0.0
AGENT_PORT=8080
LOGFIRE_TOKEN_SRE_AGENT=<token>
LANGSMITH_API_KEY=<key>
```

**MCP Server:**
```bash
MCP_HOST=0.0.0.0
MCP_PORT=30120
AGENT_SERVICE_URL=http://sre-agent-service:8080
```

## 📊 API Endpoints

### Agent Service (Port 8080)

**Health & Status:**
- `GET /health` - Liveness probe
- `GET /ready` - Readiness probe
- `GET /status` - Detailed status

**SRE Operations:**
- `POST /chat` - General SRE chat
- `POST /analyze-logs` - Log analysis
- `POST /incident-response` - Incident guidance
- `POST /monitoring-advice` - Monitoring recommendations

### MCP Server (Port 30120)

**MCP Protocol:**
- `GET /mcp` - Server info
- `POST /mcp` - JSON-RPC 2.0 requests
- `GET /sse` - Server-Sent Events

**Health:**
- `GET /health` - Liveness probe
- `GET /ready` - Readiness probe (checks agent connectivity)

## 🧪 Testing

### Test Agent Service

```bash
# Health check
curl http://localhost:8080/health

# Chat with agent
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "How do I monitor Kubernetes pods?", "thread_id": "test-1"}'

# Analyze logs
curl -X POST http://localhost:8080/analyze-logs \
  -H "Content-Type: application/json" \
  -d '{"logs": "ERROR: Connection timeout...", "thread_id": "log-1"}'
```

### Test MCP Server

```bash
# Initialize MCP session
curl -X POST http://localhost:30120/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2024-11-05",
      "capabilities": {},
      "clientInfo": {"name": "test", "version": "1.0"}
    }
  }'

# List available tools
curl -X POST http://localhost:30120/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc": "2.0", "id": 2, "method": "tools/list", "params": {}}'

# Call a tool
curl -X POST http://localhost:30120/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "sre_chat",
      "arguments": {"message": "How do I debug pod crashes?"}
    }
  }'
```

## 📈 Monitoring

### Kubernetes Resources

```bash
# Check agent status
kubectl get pods -n agent-sre -l app=sre-agent

# Check MCP server status
kubectl get pods -n agent-sre -l app=sre-agent-mcp-server

# View agent logs
kubectl logs -n agent-sre -l app=sre-agent --tail=100 -f

# View MCP server logs
kubectl logs -n agent-sre -l app=sre-agent-mcp-server --tail=100 -f
```

### Service Endpoints

```bash
# Agent via NodePort
curl http://<node-ip>:31150/health

# MCP Server via NodePort
curl http://<node-ip>:31160/health
```

## 🔐 Security

- **Sealed Secrets**: API keys and tokens managed via sealed-secrets
- **Network Policies**: Service-to-service communication restrictions
- **RBAC**: Kubernetes role-based access control
- **No Direct Exposure**: MCP server accessed through controlled endpoints

## 📝 Development

### Build Images Locally

```bash
# Build agent image
cd deployments/agent
docker build -t agent-sre:dev .

# Build MCP server image
cd deployments/mcp-server
docker build -t agent-sre-mcp-server:dev .
```

### Run with Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

## 🎯 Use Cases

1. **Automated Log Analysis**: Feed logs to the agent for AI-powered insights
2. **Incident Management**: Get step-by-step incident response guidance
3. **Monitoring Setup**: Receive recommendations for observability stack
4. **Performance Debugging**: Analyze system performance bottlenecks
5. **Knowledge Base**: Query SRE best practices and solutions

## 🤝 Integration

### Homepage Integration

The homepage API integrates with the agent via:
- Direct API calls to agent service
- MCP protocol for tool-based access
- Automatic fallback handling
- Health check monitoring

### Cursor IDE Integration

Configure MCP server in `mcp_config.json`:
```json
{
  "mcpServers": {
    "sre-agent": {
      "command": "mcp-remote",
      "args": ["http://127.0.0.1:31160/mcp/"],
      "env": {}
    }
  }
}
```

## 📚 Documentation

- **LangGraph**: https://langchain-ai.github.io/langgraph/
- **LangChain**: https://python.langchain.com/
- **MCP Protocol**: https://modelcontextprotocol.io/
- **Ollama**: https://ollama.ai/

## 🚀 Roadmap

- [ ] Add more SRE tools (kubectl, Prometheus queries, etc.)
- [ ] Implement streaming responses via SSE
- [ ] Add RAG for documentation lookup
- [ ] Integrate with Prometheus for real-time metrics
- [ ] Add custom model fine-tuning capabilities
- [ ] Multi-agent collaboration workflows

## 📞 Support

- **Issues**: Report bugs and feature requests via GitHub Issues
- **Logs**: Check Logfire dashboard for distributed traces
- **Metrics**: View Prometheus/Grafana for service metrics

---

**Version:** 0.2.0  
**Status:** ✅ Production Ready with LangGraph  
**Last Updated:** 2025-10-09
