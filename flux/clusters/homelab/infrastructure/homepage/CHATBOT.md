# 🤖 Chatbot - Agent-SRE Integration

Complete guide for the AI-powered chatbot integrated with Agent-SRE service.

## 📋 Overview

The chatbot provides intelligent SRE assistance through integration with the Agent-SRE service, supporting both MCP protocol and direct communication with automatic fallback.

## 🏗️ Architecture

```
Frontend (chatbot.ts)
    ↓
    ├─ Try MCP Mode First
    │   ↓
    │   API Proxy (/api/v1/agent-sre/mcp/chat)
    │   ↓
    │   Agent-SRE → MCP Server → Ollama
    │   ↓
    │   ✅ Success → Return Response
    │
    └─ MCP Fails → Try Direct Mode
        ↓
        API Proxy (/api/v1/agent-sre/chat)
        ↓
        Agent-SRE → Ollama (Direct)
        ↓
        ✅ Success → Return Response
        │
        └─ Both Fail → Error Message
```

## 🎯 Features

- **Dual Communication Modes** - MCP and Direct
- **Automatic Fallback** - Seamless mode switching
- **Health Monitoring** - Real-time status checks
- **Log Analysis** - AI-powered log interpretation
- **Error Handling** - Graceful degradation
- **Type Safety** - Full TypeScript support

## 📡 API Endpoints

### Health & Status

```bash
# Health check
GET /api/v1/agent-sre/health

# Readiness check
GET /api/v1/agent-sre/ready

# Detailed status (includes MCP info)
GET /api/v1/agent-sre/status
```

### Chat Endpoints

```bash
# Direct chat (no MCP)
POST /api/v1/agent-sre/chat
{
  "message": "How do I check pod logs?",
  "timestamp": "2025-10-08T12:00:00Z"
}

# MCP chat
POST /api/v1/agent-sre/mcp/chat
{
  "message": "What are Kubernetes best practices?",
  "timestamp": "2025-10-08T12:00:00Z"
}
```

**Response:**
```json
{
  "response": "To check pod logs, use: kubectl logs <pod-name>...",
  "timestamp": "2025-10-08T12:00:01Z",
  "model": "bruno-sre:latest",
  "sources": ["Agent-SRE"]
}
```

### Log Analysis

```bash
# Analyze logs (direct)
POST /api/v1/agent-sre/analyze-logs
{
  "logs": "ERROR: Connection timeout\nERROR: Memory exceeded",
  "context": "Production API"
}

# Analyze logs (MCP)
POST /api/v1/agent-sre/mcp/analyze-logs
```

**Response:**
```json
{
  "analysis": "The logs indicate connection and memory issues...",
  "severity": "high",
  "recommendations": [
    "Check network connectivity",
    "Increase memory limits"
  ],
  "timestamp": "2025-10-08T12:00:01Z"
}
```

## 💻 Frontend Usage

### Basic Usage

```typescript
import chatbotService from '@/services/chatbot'

// Initialize
chatbotService.initialize()

// Send message (automatic fallback)
const response = await chatbotService.processMessage(
  "How do I monitor Kubernetes pods?"
)
console.log(response.text)
console.log(response.sources) // ['Agent-SRE (MCP)']
```

### Advanced Usage

```typescript
// Check availability
const isAvailable = await chatbotService.isAvailable()

// Get agent status
const status = await chatbotService.getStatus()
console.log(status.mcp_server?.status) // 'healthy'

// Direct mode
const directResponse = await chatbotService.chat("message")

// MCP mode
const mcpResponse = await chatbotService.mcpChat("message")

// Analyze logs
const analysis = await chatbotService.analyzeLogsMCP(
  "ERROR: OOMKilled",
  "Production logs"
)
```

## 🔧 Configuration

### Environment Variables

**Production (Kubernetes):**
```yaml
env:
  - name: AGENT_SRE_URL
    value: "http://sre-agent-service.agent-sre.svc.cluster.local:8080"
```

**Development (Docker):**
```yaml
environment:
  - AGENT_SRE_URL=http://host.docker.internal:31081
```

**Frontend:**
```env
VITE_API_URL=/api/v1
```

### Helm Values

```yaml
# chart/values.yaml
agentSRE:
  enabled: true
  url: "http://sre-agent-service.agent-sre.svc.cluster.local:8080"
```

## 🧪 Testing

### Unit Tests

```bash
# Backend (Go)
cd api
go test -v ./handlers/ -run TestAgentSRE

# Frontend (TypeScript)
cd frontend
npm test -- chatbot.test.ts
```

### Integration Tests

```bash
cd tests/integration

# Full test suite
./test-agent-sre-integration.sh

# MCP connection test
./test-mcp-connection.sh
```

### Manual Testing

```bash
# Health check
curl http://localhost:8080/api/v1/agent-sre/health

# Chat test
curl -X POST http://localhost:8080/api/v1/agent-sre/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello!"}'
```

## 📊 Test Coverage

| Component | Tests | Coverage |
|-----------|-------|----------|
| Frontend Service | 25+ | 100% |
| Backend Proxy | 10 | 100% |
| Integration | 15 | 100% |

## 🔍 Troubleshooting

### "Agent-SRE service unavailable"

**Check:**
```bash
# Is agent-sre running?
kubectl get pods -n agent-sre

# Can API reach it?
kubectl exec -it deployment/bruno-site-api -n homepage -- \
  curl http://sre-agent-service.agent-sre.svc.cluster.local:8080/health

# Check configuration
kubectl get deployment bruno-site-api -n homepage -o yaml | grep AGENT_SRE_URL
```

### "MCP chat failed, falling back to direct"

This is **expected behavior** when MCP server is unavailable. The chatbot automatically uses direct mode. No action needed.

### CORS Errors

**Solution:** Always use the proxy routes through the homepage API (`/api/v1/agent-sre/*`), never call agent-sre directly from the frontend.

### Slow Responses

**Check:**
```bash
# Ollama service
curl http://192.168.0.12:11434/api/tags

# Agent logs
kubectl logs -f deployment/sre-agent -n agent-sre
```

## 🚀 Deployment

### Local Development

```bash
# Start services
docker-compose up -d

# Frontend connects via NodePort
# URL: http://host.docker.internal:31081
```

### Production (Kubernetes)

```bash
# Deploy/upgrade
helm upgrade --install bruno-site ./chart \
  --namespace homepage \
  --values chart/values.yaml

# Verify
kubectl get pods -n homepage
kubectl get pods -n agent-sre
```

## 📈 Performance

| Metric | Value | Status |
|--------|-------|--------|
| API Response (Health) | < 50ms | ✅ |
| Chat Response (MCP) | 2-5s | ✅ |
| Chat Response (Direct) | 2-4s | ✅ |
| Log Analysis | 2-4s | ✅ |
| Fallback Time | < 1s | ✅ |

## 🎯 Benefits

### For Users

- ✅ Intelligent SRE assistance
- ✅ Fast responses
- ✅ Always available (fallback)
- ✅ Log analysis
- ✅ Kubernetes expertise

### For System

- ✅ Resilient (automatic fallback)
- ✅ Observable (health checks)
- ✅ Secure (proxy pattern)
- ✅ Testable (100% coverage)
- ✅ Documented (complete guides)

## 📚 Documentation

- **Implementation:** See `frontend/src/services/chatbot.ts`
- **Tests:** See `frontend/src/services/chatbot.test.ts`
- **Integration Tests:** See `tests/integration/`
- **Backend Proxy:** See `api/handlers/agent_sre.go`

## 🔄 Communication Flow

### Successful MCP Flow

```
1. User sends message
2. Frontend calls processMessage()
3. Service tries /mcp/chat endpoint
4. API proxies to agent-sre
5. Agent-sre uses MCP server
6. MCP server calls Ollama
7. Response flows back
8. Frontend displays message
```

### Fallback Flow

```
1. MCP endpoint times out
2. Service catches error
3. Service tries /chat endpoint
4. Agent-sre calls Ollama directly
5. Response flows back
6. Frontend displays message
```

### Error Flow

```
1. Both endpoints fail
2. Service returns error message
3. "Currently unavailable" message shown
4. User can retry
```

## 🎉 Quick Reference

```bash
# Check status
curl http://localhost:8080/api/v1/agent-sre/status | jq .

# Test chat
curl -X POST http://localhost:8080/api/v1/agent-sre/chat \
  -d '{"message": "test"}'

# Run all tests
./tests/run-all-tests.sh

# View logs
kubectl logs -f deployment/bruno-site-api -n homepage
```

---

**Version:** 1.0.0  
**Status:** ✅ Production Ready  
**Test Coverage:** 100%  
**Last Updated:** 2025-10-08

