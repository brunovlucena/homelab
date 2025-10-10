# 🤖 Homepage Chatbot ↔️ Jamie Integration

## Overview

The Homepage chatbot is now connected to **Jamie**, the AI-powered SRE assistant. Jamie provides intelligent responses by combining:
- 🧠 **Ollama AI** - Bruno's fine-tuned SRE model
- 🔧 **Agent-SRE MCP** - Real-time infrastructure monitoring and operations
- 📊 **Golden Signals** - Service health metrics (latency, traffic, errors, saturation)
- ☸️ **Kubernetes Operations** - Pod logs, deployments, and cluster management
- 📈 **Prometheus/Grafana** - Metrics queries and dashboard access

## Architecture

```
┌─────────────────┐      ┌─────────────────┐      ┌─────────────────┐
│   Homepage      │      │   Homepage      │      │   Jamie         │
│   Frontend      │─────▶│   API (Go)      │─────▶│   Slack Bot     │
│   (React)       │      │                 │      │   REST API      │
└─────────────────┘      └─────────────────┘      └─────────────────┘
                                                            │
                                                            ├─────▶ Ollama AI
                                                            │       (bruno-sre)
                                                            │
                                                            └─────▶ Agent-SRE
                                                                    MCP Tools
```

## Components

### 1. Frontend Chatbot (`/frontend/src/services/chatbot.ts`)

**Endpoint**: `/api/v1/jamie/chat`

**Request Format**:
```json
{
  "message": "Your question here",
  "timestamp": "2025-10-10T12:00:00Z"
}
```

**Response Format**:
```json
{
  "response": "Jamie's response",
  "timestamp": "2025-10-10T12:00:01Z",
  "sources": ["Jamie"]
}
```

### 2. Homepage API Proxy (`/api/handlers/jamie.go`)

The Go API acts as a secure proxy between the frontend and Jamie:

**Features**:
- ✅ Request forwarding to Jamie service
- ✅ Header preservation and client IP tracking
- ✅ Error handling with graceful fallback
- ✅ Health checks via `/health` and `/ready` endpoints
- ✅ 60-second timeout for complex AI queries

**Available Endpoints**:
```
POST /api/v1/jamie/chat              - Main chatbot endpoint
POST /api/v1/jamie/golden-signals    - Check service health metrics
POST /api/v1/jamie/prometheus/query  - Execute PromQL queries
POST /api/v1/jamie/pod-logs          - Get Kubernetes pod logs
POST /api/v1/jamie/analyze-logs      - AI-powered log analysis
GET  /api/v1/jamie/health            - Jamie service health
GET  /api/v1/jamie/ready             - Jamie service readiness
```

### 3. Jamie Slack Bot REST API

**Service URL**: `http://jamie-slack-bot-service.jamie.svc.cluster.local:8080`

Jamie processes messages by:
1. Analyzing the question type
2. Routing to appropriate service (Ollama AI or Agent-SRE MCP)
3. Gathering context and data from infrastructure
4. Synthesizing a comprehensive response
5. Returning formatted answer to homepage

## Configuration

### Environment Variables

**Homepage API** (`chart/values.yaml`):
```yaml
jamie:
  enabled: true
  url: "http://jamie-slack-bot-service.jamie.svc.cluster.local:8080"
```

**Docker Compose** (`docker-compose.yml`):
```yaml
environment:
  - JAMIE_URL=http://host.docker.internal:8081
```

### Kubernetes Deployment

The Helm chart automatically configures the connection:

```yaml
# chart/templates/api-deployment.yaml
{{- if .Values.jamie.enabled }}
- name: JAMIE_URL
  value: "{{ .Values.jamie.url }}"
{{- end }}
```

## Usage Examples

### Simple Question
```
User: "How do I check if my services are healthy?"
Jamie: "You can check service health using golden signals..."
```

### Infrastructure Query
```
User: "Show me the error rate for the API"
Jamie: "📊 The API error rate is 0.1%..."
```

### SRE Advice
```
User: "What are best practices for alerting?"
Jamie: "🤖 Here are key alerting best practices..."
```

## Testing

### 1. Check Service Health

```bash
# Test Jamie directly
curl http://jamie-slack-bot-service.jamie.svc.cluster.local:8080/health

# Test via Homepage API
curl http://localhost:8080/api/v1/jamie/health
```

### 2. Send Test Chat Message

```bash
# Via Homepage API
curl -X POST http://localhost:8080/api/v1/jamie/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Hello Jamie, how are you?"
  }'
```

### 3. Check Frontend Integration

1. Open homepage at `http://localhost:3000`
2. Click the chatbot icon (bottom right)
3. Send a message: "Hello Jamie"
4. Check browser console for connection logs

## Troubleshooting

### Frontend Not Connecting

Check browser console:
```javascript
🤖 [ChatbotService] Initializing Jamie connection...
🤖 [ChatbotService] Jamie URL: /api/v1/jamie
```

### API Proxy Issues

Check API logs:
```bash
kubectl logs -n bruno -l app.kubernetes.io/name=homepage-api
```

Look for:
```
🤖 [JamieHandler] Proxying request to Jamie
```

### Jamie Service Down

Check Jamie status:
```bash
kubectl get pods -n jamie
kubectl logs -n jamie -l app=jamie-slack-bot
```

Common issues:
- Ollama server unreachable (192.168.0.16:11434)
- Agent-SRE service unavailable
- Slack token issues (for Slack Bot only, doesn't affect REST API)

### Timeout Errors

If queries timeout:
1. Check if Ollama is responding (may be slow for complex queries)
2. Verify Agent-SRE connectivity
3. Consider increasing timeout in `chatbot.ts` (currently 60s)

## Benefits

### For Users
- 🎯 **Natural Language** - Ask questions in plain English
- ⚡ **Real-Time Data** - Get live metrics and infrastructure status
- 🧠 **AI-Powered** - Intelligent responses with context awareness
- 🔍 **Comprehensive** - Access to logs, metrics, and documentation

### For Developers
- 🔒 **Secure** - Proxied through Go API with proper authentication
- 📊 **Observable** - Full request/response logging
- 🛡️ **Resilient** - Graceful error handling and timeouts
- 🔧 **Extensible** - Easy to add new Jamie capabilities

## Future Enhancements

- [ ] Add conversation history/context
- [ ] Implement streaming responses for long queries
- [ ] Add voice input support
- [ ] Integrate with more MCP tools
- [ ] Add response caching for common questions
- [ ] Implement rate limiting per user
- [ ] Add analytics dashboard for chat usage

## Related Documentation

- [Jamie README](../agent-jamie/README.md) - Jamie architecture and features
- [Agent-SRE](../agent-sre/README.md) - Backend SRE agent with MCP tools
- [Homepage API](./api/README.md) - Go API documentation
- [Frontend](./frontend/README.md) - React frontend documentation

---

**Status**: ✅ Connected and Ready

**Last Updated**: 2025-10-10

Made with ❤️ by Bruno Lucena

