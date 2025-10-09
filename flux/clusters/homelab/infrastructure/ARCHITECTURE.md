# 🏗️ Infrastructure Architecture

## Overview

This document describes the architecture of the SRE infrastructure, focusing on the integration between Homepage, Jamie (AI-powered SRE assistant), and Agent-SRE (SRE operations backend).

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                    🌐 Homepage (Go API)                         │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  /api/v1/jamie/chat                                     │   │
│  │  /api/v1/jamie/golden-signals                          │   │
│  │  /api/v1/jamie/prometheus/query                        │   │
│  │  /api/v1/jamie/pod-logs                                │   │
│  │  /api/v1/jamie/analyze-logs                            │   │
│  └─────────────────────────────────────────────────────────┘   │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            │ HTTP REST API
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│              🤖 Jamie MCP Server (Python)                       │
│                     Port: 30121                                 │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  REST API Endpoints (for Homepage)                      │   │
│  │  • POST /api/chat                                       │   │
│  │  • POST /api/golden-signals                             │   │
│  │  • POST /api/prometheus/query                           │   │
│  │  • POST /api/pod-logs                                   │   │
│  │  • POST /api/analyze-logs                               │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  MCP Protocol Endpoints (for Cursor IDE)                │   │
│  │  • POST /mcp (JSON-RPC 2.0)                             │   │
│  │  • GET  /mcp (Server info)                              │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  🧠 AI Integration                                      │   │
│  │  • Ollama: http://192.168.0.16:11434                    │   │
│  │  • Model: llama3.2:3b                                   │   │
│  └─────────────────────────────────────────────────────────┘   │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            │ MCP JSON-RPC 2.0
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│           🔧 Agent-SRE MCP Server (Python)                      │
│                     Port: 30120                                 │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  MCP Tools (exposed via JSON-RPC)                       │   │
│  │  • sre_chat                                             │   │
│  │  • check_golden_signals                                 │   │
│  │  • query_prometheus                                     │   │
│  │  • get_pod_logs                                         │   │
│  │  • analyze_logs                                         │   │
│  │  • incident_response                                    │   │
│  │  • monitoring_advice                                    │   │
│  └─────────────────────────────────────────────────────────┘   │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            │ HTTP API
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│          🎯 Agent-SRE Agent Service (Python)                    │
│                     Port: 8080                                  │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  Core SRE Operations                                    │   │
│  │  • Kubernetes API integration                           │   │
│  │  • Prometheus queries                                   │   │
│  │  • Log analysis with AI                                 │   │
│  │  • Incident investigation                               │   │
│  │  • Performance monitoring                               │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

## Component Details

### 1. Homepage API (Go)

**Location:** `/infrastructure/homepage/api/`

**Purpose:** Main website API and frontend backend

**Key Features:**
- Provides REST API for the Homepage frontend
- Proxies chatbot requests to Jamie
- Manages projects, skills, experiences, and content
- Integrates with MinIO for asset storage
- Cloudflare CDN integration

**Environment Variables:**
```bash
JAMIE_URL=http://jamie-mcp-server-service.jamie.svc.cluster.local:30121
```

**API Endpoints:**
```
POST /api/v1/jamie/chat                  # Main chatbot endpoint
POST /api/v1/jamie/golden-signals        # Check service health
POST /api/v1/jamie/prometheus/query      # PromQL queries
POST /api/v1/jamie/pod-logs              # Get pod logs
POST /api/v1/jamie/analyze-logs          # AI log analysis
GET  /api/v1/jamie/health                # Health check
GET  /api/v1/jamie/ready                 # Readiness check
```

**Example Request:**
```bash
curl -X POST http://localhost:8080/api/v1/jamie/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Check the golden signals for homepage service"}'
```

**Example Response:**
```json
{
  "response": "🤖 Jamie: Let me check the golden signals for the homepage service...",
  "timestamp": "2025-10-09T12:00:00Z"
}
```

---

### 2. Jamie MCP Server (Python)

**Location:** `/infrastructure/jamie/src/mcp-server/`

**Purpose:** AI-powered SRE assistant with dual interfaces (REST + MCP)

**Key Features:**
- **REST API** for Homepage integration (simple HTTP endpoints)
- **MCP Protocol** for Cursor IDE integration (JSON-RPC 2.0)
- AI-powered responses using Ollama (llama3.2:3b)
- Forwards SRE operations to Agent-SRE MCP Server
- Logfire instrumentation for observability

**Environment Variables:**
```bash
AGENT_SRE_URL=http://sre-agent-mcp-server-service.agent-sre.svc.cluster.local:30120
OLLAMA_URL=http://192.168.0.16:11434
MODEL_NAME=llama3.2:3b
MCP_HOST=0.0.0.0
MCP_PORT=30121
LOGFIRE_TOKEN_JAMIE_MCP=your-token-here  # Optional
```

**REST API Endpoints (for Homepage):**
```
POST /api/chat                           # AI chat interface
POST /api/golden-signals                 # Check golden signals
POST /api/prometheus/query               # Execute PromQL
POST /api/pod-logs                       # Get pod logs
POST /api/analyze-logs                   # Analyze logs with AI
```

**MCP Endpoints (for Cursor IDE):**
```
POST /mcp                                # MCP JSON-RPC endpoint
GET  /mcp                                # Server information
```

**MCP Tools Available:**
- `ask_jamie` - Main AI interaction
- `check_golden_signals` - Monitor service health
- `query_prometheus` - Execute PromQL queries
- `get_pod_logs` - Retrieve pod logs
- `analyze_logs` - AI-powered log analysis
- `sre_chat` - General SRE consultation

**Example REST Request:**
```bash
curl -X POST http://localhost:30121/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "What are the best practices for monitoring microservices?"}'
```

**Example MCP Request (JSON-RPC 2.0):**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "ask_jamie",
    "arguments": {
      "question": "How do I check service health?"
    }
  }
}
```

---

### 3. Agent-SRE MCP Server (Python)

**Location:** `/infrastructure/agent-sre/deployments/mcp-server/`

**Purpose:** Thin MCP protocol layer that forwards requests to Agent-SRE Agent Service

**Key Features:**
- Implements MCP JSON-RPC 2.0 protocol
- Exposes SRE tools via MCP
- Forwards execution to Agent Service
- No AI integration (stateless proxy)

**Environment Variables:**
```bash
AGENT_SERVICE_URL=http://sre-agent-service:8080
MCP_HOST=0.0.0.0
MCP_PORT=30120
```

**MCP Tools:**
```
sre_chat              # General SRE chat
check_golden_signals  # Check service golden signals
query_prometheus      # Execute PromQL queries
get_pod_logs          # Get Kubernetes pod logs
analyze_logs          # Analyze logs
incident_response     # Incident guidance
monitoring_advice     # Monitoring advice
health_check          # Check agent health
```

**Example MCP Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "check_golden_signals",
    "arguments": {
      "service_name": "homepage",
      "namespace": "default"
    }
  }
}
```

---

### 4. Agent-SRE Agent Service (Python)

**Location:** `/infrastructure/agent-sre/deployments/agent/`

**Purpose:** Core SRE operations engine

**Key Features:**
- Kubernetes API integration
- Prometheus query execution
- AI-powered log analysis
- Incident investigation
- Performance monitoring
- Golden signals tracking

**HTTP Endpoints:**
```
POST /chat                               # SRE chat
POST /analyze-logs                       # Log analysis
POST /incident-response                  # Incident guidance
POST /monitoring-advice                  # Monitoring advice
POST /golden-signals                     # Golden signals
POST /prometheus/query                   # Prometheus query
POST /kubernetes/logs                    # Pod logs
GET  /health                             # Health check
```

---

## Data Flow

### Homepage Chatbot Request Flow

```
1. User sends message to Homepage frontend
   ↓
2. Frontend calls: POST /api/v1/jamie/chat
   {
     "message": "Check homepage service health"
   }
   ↓
3. Homepage Go API proxies to Jamie MCP Server
   ↓
4. Jamie MCP Server:
   - Receives REST request
   - Uses Ollama AI to process message
   - Determines which tool to use
   - Calls Agent-SRE via MCP protocol
   ↓
5. Agent-SRE MCP Server:
   - Receives MCP JSON-RPC request
   - Forwards to Agent Service
   ↓
6. Agent-SRE Agent Service:
   - Executes Kubernetes/Prometheus queries
   - Returns results
   ↓
7. Response flows back through the chain
   ↓
8. User receives AI-powered response with actual data
```

### Example: Golden Signals Check

```bash
# 1. Homepage API call
curl -X POST http://homepage-api/api/v1/jamie/chat \
  -d '{"message": "check golden signals for homepage"}'

# 2. Jamie determines to use check_golden_signals tool

# 3. Jamie → Agent-SRE MCP call
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "check_golden_signals",
    "arguments": {
      "service_name": "homepage",
      "namespace": "default"
    }
  }
}

# 4. Agent-SRE queries Prometheus and returns:
{
  "status": "healthy",
  "latency": "45ms (p95)",
  "traffic": "150 req/min",
  "errors": "0.1%",
  "saturation": "CPU: 15%, Memory: 25%"
}

# 5. Jamie formats response with AI
"🤖 Jamie: The homepage service is healthy! Here's what I found:
- Latency: 45ms (p95) ✅
- Traffic: 150 requests/min 📈
- Error rate: 0.1% ✅
- Resource usage: CPU 15%, Memory 25% 💚
Everything looks good!"
```

---

## Integration Patterns

### 1. Homepage → Jamie (REST)

**Pattern:** Simple HTTP REST API  
**Protocol:** JSON over HTTP  
**Use Case:** Homepage chatbot, frontend integrations

```go
// Go code in homepage API
response, err := http.Post(
    jamieURL + "/api/chat",
    "application/json",
    bytes.NewBuffer(jsonData),
)
```

### 2. Jamie → Agent-SRE (MCP)

**Pattern:** MCP JSON-RPC 2.0  
**Protocol:** JSON-RPC over HTTP  
**Use Case:** Tool execution, infrastructure operations

```python
# Python code in Jamie
payload = {
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
        "name": "check_golden_signals",
        "arguments": {"service_name": "homepage"}
    }
}
response = await session.post(agent_sre_url + "/mcp", json=payload)
```

### 3. Agent-SRE MCP → Agent Service

**Pattern:** Internal HTTP API  
**Protocol:** JSON over HTTP  
**Use Case:** Actual SRE operations execution

```python
# Python code in Agent-SRE MCP Server
response = await session.post(
    agent_service_url + "/golden-signals",
    json={"service_name": "homepage", "namespace": "default"}
)
```

---

## Deployment Configuration

### Homepage API

**Kubernetes Service:**
```yaml
apiVersion: v1
kind: Service
metadata:
  name: bruno-site-api-service
  namespace: homepage
spec:
  ports:
  - port: 8080
    targetPort: 8080
```

**Environment:**
```yaml
env:
- name: JAMIE_URL
  value: http://jamie-mcp-server-service.jamie.svc.cluster.local:30121
```

### Jamie MCP Server

**Kubernetes Service:**
```yaml
apiVersion: v1
kind: Service
metadata:
  name: jamie-mcp-server-service
  namespace: jamie
spec:
  type: NodePort
  ports:
  - port: 30121
    targetPort: 30121
    nodePort: 30121
```

**Environment:**
```yaml
env:
- name: AGENT_SRE_URL
  value: http://sre-agent-mcp-server-service.agent-sre.svc.cluster.local:30120
- name: OLLAMA_URL
  value: http://192.168.0.16:11434
- name: MODEL_NAME
  value: llama3.2:3b
```

### Agent-SRE MCP Server

**Kubernetes Service:**
```yaml
apiVersion: v1
kind: Service
metadata:
  name: sre-agent-mcp-server-service
  namespace: agent-sre
spec:
  type: NodePort
  ports:
  - port: 30120
    targetPort: 30120
    nodePort: 30120
```

**Environment:**
```yaml
env:
- name: AGENT_SERVICE_URL
  value: http://sre-agent-service:8080
```

---

## Testing

### Test Homepage → Jamie Integration

```bash
# Test chat endpoint
curl -X POST http://localhost:8080/api/v1/jamie/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello Jamie!"}'

# Test golden signals
curl -X POST http://localhost:8080/api/v1/jamie/golden-signals \
  -H "Content-Type: application/json" \
  -d '{"service_name": "homepage", "namespace": "default"}'
```

### Test Jamie → Agent-SRE Integration

```bash
# Test Jamie MCP endpoint directly
curl -X POST http://localhost:30121/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "check_golden_signals",
      "arguments": {"service_name": "homepage"}
    }
  }'
```

### Test Agent-SRE MCP Server

```bash
# Test Agent-SRE MCP endpoint
curl -X POST http://localhost:30120/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "health_check",
      "arguments": {}
    }
  }'
```

---

## Observability

### Logfire Integration

All components are instrumented with Logfire for distributed tracing:

```python
# Example instrumentation
@logfire.instrument("rest_chat")
async def handle_rest_chat(self, request: Request) -> Response:
    # Automatically traces request/response
    pass
```

**Logfire Tokens:**
- Jamie MCP Server: `LOGFIRE_TOKEN_JAMIE_MCP`
- Agent-SRE: `LOGFIRE_TOKEN_AGENT_SRE`

### Metrics

**Prometheus Metrics:**
- Request counts: `http_requests_total`
- Request duration: `http_request_duration_seconds`
- Error rates: `http_requests_errors_total`

**ServiceMonitors:**
- Homepage API: Port 8080
- Jamie MCP Server: Port 30121
- Agent-SRE MCP Server: Port 30120

---

## Troubleshooting

### Homepage can't reach Jamie

```bash
# Check Jamie service
kubectl get svc -n jamie

# Check Jamie pods
kubectl get pods -n jamie

# Test Jamie endpoint
curl http://jamie-mcp-server-service.jamie.svc.cluster.local:30121/health
```

### Jamie can't reach Agent-SRE

```bash
# Check Agent-SRE service
kubectl get svc -n agent-sre

# Check Agent-SRE pods
kubectl get pods -n agent-sre

# Test Agent-SRE endpoint
curl http://sre-agent-mcp-server-service.agent-sre.svc.cluster.local:30120/health
```

### AI responses not working

```bash
# Check Ollama connection
curl http://192.168.0.16:11434/api/tags

# Check Jamie logs for Ollama errors
kubectl logs -n jamie -l app=jamie-mcp-server | grep -i ollama
```

---

## Security Considerations

1. **Internal Services:** Agent-SRE and Jamie communicate via internal Kubernetes services
2. **External Access:** Only Homepage API is exposed publicly
3. **Authentication:** Add authentication middleware if needed
4. **Rate Limiting:** Implement rate limiting for public endpoints
5. **Network Policies:** Use Kubernetes Network Policies to restrict traffic

---

## Future Enhancements

1. **Caching Layer:** Add Redis caching for frequent queries
2. **Streaming Responses:** Support SSE for real-time AI responses
3. **Multi-Model Support:** Support different AI models based on query type
4. **Enhanced Context:** Pass conversation history for better AI responses
5. **Authorization:** Add RBAC for different user roles
6. **WebSocket Support:** Real-time bidirectional communication

---

## Summary

The architecture follows a clean separation of concerns:

1. **Homepage API (Go)** - Public-facing REST API
2. **Jamie MCP Server (Python)** - AI-powered interface layer (REST + MCP)
3. **Agent-SRE MCP Server (Python)** - MCP protocol layer
4. **Agent-SRE Agent Service (Python)** - Core SRE operations

This design provides:
- ✅ Simple REST API for Homepage integration
- ✅ MCP protocol for IDE (Cursor) integration
- ✅ AI-powered responses via Ollama
- ✅ Clean separation of concerns
- ✅ Easy to test and debug
- ✅ Observable via Logfire
- ✅ Scalable and maintainable

---

**Author:** Bruno Lucena  
**Date:** October 9, 2025  
**Version:** 1.0

