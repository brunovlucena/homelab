# 🚀 Quick Reference Guide

## Architecture Summary

```
Homepage API (Go) → Jamie MCP Server (Python) → Agent-SRE MCP Server (Python) → Agent-SRE Agent Service (Python)
    Port 8080            Port 30121                    Port 30120                     Port 8080
```

## 📍 Component Locations

| Component | Location | Language |
|-----------|----------|----------|
| Homepage API | `/infrastructure/homepage/api/` | Go |
| Jamie MCP Server | `/infrastructure/jamie/src/mcp-server/` | Python |
| Jamie Slack Bot | `/infrastructure/jamie/src/slack-bot/` | Python |
| Agent-SRE MCP Server | `/infrastructure/agent-sre/deployments/mcp-server/` | Python |
| Agent-SRE Agent Service | `/infrastructure/agent-sre/deployments/agent/` | Python |

## 🔌 API Endpoints

### Homepage API (Port 8080)

```bash
# Main chatbot endpoint
POST /api/v1/jamie/chat
{
  "message": "Your question here"
}

# Check service golden signals
POST /api/v1/jamie/golden-signals
{
  "service_name": "homepage",
  "namespace": "default"
}

# Execute PromQL query
POST /api/v1/jamie/prometheus/query
{
  "query": "up{job=\"homepage\"}"
}

# Get pod logs
POST /api/v1/jamie/pod-logs
{
  "pod_name": "homepage-xyz",
  "namespace": "default",
  "tail_lines": 100
}

# Analyze logs
POST /api/v1/jamie/analyze-logs
{
  "logs": "your log data here",
  "context": "optional context"
}

# Health check
GET /api/v1/jamie/health
GET /api/v1/jamie/ready
```

### Jamie MCP Server (Port 30121)

```bash
# REST API (for Homepage)
POST /api/chat
POST /api/golden-signals
POST /api/prometheus/query
POST /api/pod-logs
POST /api/analyze-logs

# MCP Protocol (for Cursor IDE)
POST /mcp
GET  /mcp

# Health
GET /health
GET /ready
```

### Agent-SRE MCP Server (Port 30120)

```bash
# MCP Protocol only
POST /mcp
GET  /mcp

# Health
GET /health
GET /ready
```

## 🧪 Testing Commands

### Test Full Chain (Homepage → Jamie → Agent-SRE)

```bash
# Test from Homepage API
curl -X POST http://localhost:8080/api/v1/jamie/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Check the golden signals for homepage"}'
```

### Test Jamie Directly

```bash
# Test REST API
curl -X POST http://localhost:30121/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "What are best practices for monitoring?"}'

# Test MCP endpoint
curl -X POST http://localhost:30121/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "ask_jamie",
      "arguments": {
        "question": "How do I check service health?"
      }
    }
  }'
```

### Test Agent-SRE Directly

```bash
# Test MCP endpoint
curl -X POST http://localhost:30120/mcp \
  -H "Content-Type: application/json" \
  -d '{
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
  }'
```

## 🐛 Debugging

### Check Service Status

```bash
# Homepage
kubectl get pods -n homepage
kubectl logs -n homepage -l app=bruno-site-api

# Jamie
kubectl get pods -n jamie
kubectl logs -n jamie -l app=jamie-mcp-server

# Agent-SRE
kubectl get pods -n agent-sre
kubectl logs -n agent-sre -l app=sre-agent-mcp-server
kubectl logs -n agent-sre -l app=sre-agent
```

### Check Service Connectivity

```bash
# From Homepage to Jamie
kubectl exec -n homepage <homepage-pod> -- \
  curl http://jamie-mcp-server-service.jamie.svc.cluster.local:30121/health

# From Jamie to Agent-SRE
kubectl exec -n jamie <jamie-pod> -- \
  curl http://sre-agent-mcp-server-service.agent-sre.svc.cluster.local:30120/health

# From Agent-SRE MCP to Agent Service
kubectl exec -n agent-sre <mcp-server-pod> -- \
  curl http://sre-agent-service:8080/health
```

### Check Ollama Connection

```bash
# Test Ollama API
curl http://192.168.0.16:11434/api/tags

# Check if model is available
curl http://192.168.0.16:11434/api/tags | jq '.models[] | select(.name | contains("llama3.2"))'
```

## 🔧 Configuration

### Environment Variables

#### Homepage API
```bash
JAMIE_URL=http://jamie-mcp-server-service.jamie.svc.cluster.local:30121
```

#### Jamie MCP Server
```bash
AGENT_SRE_URL=http://sre-agent-mcp-server-service.agent-sre.svc.cluster.local:30120
OLLAMA_URL=http://192.168.0.16:11434
MODEL_NAME=llama3.2:3b
MCP_HOST=0.0.0.0
MCP_PORT=30121
LOGFIRE_TOKEN_JAMIE_MCP=your-token-here  # Optional
```

#### Agent-SRE MCP Server
```bash
AGENT_SERVICE_URL=http://sre-agent-service:8080
MCP_HOST=0.0.0.0
MCP_PORT=30120
```

## 📦 Build & Deploy

### Homepage API

```bash
cd /infrastructure/homepage
make build-api
make deploy
```

### Jamie

```bash
cd /infrastructure/jamie
make build      # Build both slack-bot and mcp-server
make push       # Push to registry
make deploy     # Deploy to cluster
```

### Agent-SRE

```bash
cd /infrastructure/agent-sre
make build      # Build both agent and mcp-server
make push       # Push to registry
make deploy     # Deploy to cluster
```

## 🔍 Monitoring

### Metrics

```bash
# Homepage metrics
curl http://localhost:8080/metrics

# Jamie metrics (if exposed)
curl http://localhost:30121/metrics

# Agent-SRE metrics (if exposed)
curl http://localhost:30120/metrics
```

### Logfire

Check traces at: https://logfire.pydantic.dev

**Tokens needed:**
- `LOGFIRE_TOKEN_JAMIE_MCP` - Jamie MCP Server
- `LOGFIRE_TOKEN_AGENT_SRE` - Agent-SRE

## 🎯 Common Use Cases

### 1. User asks question on Homepage

```
User → Frontend → Homepage API (/api/v1/jamie/chat)
     → Jamie (AI processes with Ollama)
     → Agent-SRE (executes if needed)
     → Response back to user
```

### 2. Check service health

```bash
curl -X POST http://localhost:8080/api/v1/jamie/golden-signals \
  -d '{"service_name": "homepage", "namespace": "default"}'
```

### 3. Query metrics

```bash
curl -X POST http://localhost:8080/api/v1/jamie/prometheus/query \
  -d '{"query": "rate(http_requests_total[5m])"}'
```

### 4. Get pod logs

```bash
curl -X POST http://localhost:8080/api/v1/jamie/pod-logs \
  -d '{"pod_name": "homepage-xyz", "namespace": "default", "tail_lines": 100}'
```

### 5. Analyze logs

```bash
curl -X POST http://localhost:8080/api/v1/jamie/analyze-logs \
  -d '{
    "logs": "ERROR: Connection timeout\nERROR: Database unavailable",
    "context": "Production API logs"
  }'
```

## 🚨 Troubleshooting Checklist

- [ ] All pods are running
- [ ] Services are accessible (health endpoints return 200)
- [ ] Network connectivity between services works
- [ ] Ollama is accessible from Jamie
- [ ] Prometheus is accessible from Agent-SRE
- [ ] Kubernetes API is accessible from Agent-SRE
- [ ] Environment variables are set correctly
- [ ] Secrets are created (for Jamie Slack bot)
- [ ] Logfire tokens are valid (optional)

## 📚 Documentation

- **Architecture:** `ARCHITECTURE.md` - Detailed system architecture
- **Jamie README:** `/infrastructure/jamie/README.md`
- **Agent-SRE README:** `/infrastructure/agent-sre/README.md`
- **Homepage README:** `/infrastructure/homepage/README.md`

## 🆘 Support

Check logs and traces:
1. Kubernetes logs: `kubectl logs -n <namespace> <pod>`
2. Logfire traces: https://logfire.pydantic.dev
3. Prometheus: http://192.168.0.16:30090
4. Grafana: http://192.168.0.16:30091

---

**Quick Tip:** Start with testing health endpoints, then move up the chain:
```bash
# 1. Test Agent-SRE
curl http://localhost:30120/health

# 2. Test Jamie
curl http://localhost:30121/health

# 3. Test Homepage API
curl http://localhost:8080/api/v1/jamie/health

# 4. Test full integration
curl -X POST http://localhost:8080/api/v1/jamie/chat -d '{"message": "hello"}'
```

