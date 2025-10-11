# 🤖 Agent Bruno - AI Assistant with Memory

Agent Bruno is an AI-powered assistant with deep knowledge of the homepage application and IP-based conversation memory.

## 🎯 Features

- **📚 Homepage Knowledge**: Deep understanding of the homepage application architecture, APIs, and components
- **🧠 IP-Based Memory**: Maintains conversation history per unique IP using Redis (sessions) and MongoDB (persistent)
- **🔌 MCP Protocol**: Full MCP server implementation for structured communication
- **⚡ FastAPI**: High-performance async API
- **📊 Observability**: Prometheus metrics and OpenTelemetry tracing
- **🎵 Production-Ready**: Kubernetes deployment with health checks and auto-scaling

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Agent Bruno                             │
│                                                             │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐   │
│  │  FastAPI     │    │    Memory    │    │  Knowledge   │   │
│  │  Server      │◄──►│   Manager    │◄──►│    Base      │   │
│  │              │    │              │    │              │   │
│  └──────┬───────┘    └──────┬───────┘    └──────────────┘   │
│         │                   │                               │
│         ▼                   ▼                               │
│  ┌──────────────┐    ┌──────────────┐                       │
│  │ MCP Server   │    │    Redis     │                       │
│  │              │    │   (Session)  │                       │
│  └──────────────┘    └──────────────┘                       │
│                             │                               │
│                             ▼                               │
│                      ┌──────────────┐                       │
│                      │   MongoDB    │                       │
│                      │  (Persistent)│                       │
│                      └──────────────┘                       │
└─────────────────────────────────────────────────────────────┘
```

## 🚀 Quick Start

### Local Development

```bash
# Install dependencies
uv sync

# Run locally
uvicorn agent_bruno.main:app --reload --port 8080

# Run tests
pytest -v
```

### Docker

```bash
# Build
docker build -t agent-bruno:latest .

# Run
docker run -p 8080:8080 \
  -e REDIS_URL=redis://localhost:6379 \
  -e MONGODB_URL=mongodb://localhost:27017 \
  agent-bruno:latest
```

### Kubernetes

```bash
# Deploy
kubectl apply -k /path/to/agent-bruno

# Check status
kubectl get pods -n agent-bruno
```

## 📡 API Endpoints

### Chat Endpoints

- `POST /chat` - Direct chat with agent
- `POST /mcp/chat` - MCP protocol chat

### Memory Endpoints

- `GET /memory/{ip}` - Get conversation history for IP
- `DELETE /memory/{ip}` - Clear memory for IP

### Health & Status

- `GET /health` - Health check
- `GET /ready` - Readiness check
- `GET /metrics` - Prometheus metrics

## 🧠 Memory System

### Session Memory (Redis)

- **TTL**: 24 hours
- **Purpose**: Recent conversation context
- **Key Format**: `bruno:session:{ip}`

### Persistent Memory (MongoDB)

- **Collection**: `conversations`
- **Purpose**: Long-term conversation history
- **Indexed**: IP address, timestamp

### Memory Structure

```json
{
  "ip": "192.168.1.100",
  "timestamp": "2025-10-11T12:00:00Z",
  "message": "How do I deploy the homepage?",
  "response": "To deploy the homepage...",
  "context": {
    "user_agent": "Mozilla/5.0...",
    "endpoint": "/chat"
  }
}
```

## 📚 Homepage Knowledge

Agent Bruno has deep knowledge of:

- Architecture and components
- API endpoints and handlers
- Database schema
- Deployment procedures
- Configuration options
- Security practices
- CI/CD workflows

## 🔧 Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | API server port |
| `REDIS_URL` | `redis://redis:6379` | Redis connection URL |
| `MONGODB_URL` | `mongodb://mongodb:27017` | MongoDB connection URL |
| `MONGODB_DB` | `agent_bruno` | MongoDB database name |
| `SESSION_TTL` | `86400` | Session TTL in seconds (24h) |
| `LOG_LEVEL` | `INFO` | Logging level |
| `OLLAMA_URL` | `http://192.168.0.16:11434` | Ollama server URL |

## 📊 Monitoring

### Metrics

- `bruno_requests_total` - Total requests
- `bruno_request_duration_seconds` - Request duration
- `bruno_memory_operations_total` - Memory operations
- `bruno_active_sessions` - Active sessions

### Health Checks

- **Liveness**: `/health` - Basic health check
- **Readiness**: `/ready` - Checks Redis and MongoDB connectivity

## 🔐 Security

- IP-based isolation (no cross-IP memory access)
- No authentication required (internal service)
- CORS configured for homepage frontend
- Rate limiting per IP (future)

## 📈 Scaling

- Stateless design (all state in Redis/MongoDB)
- Horizontal scaling supported
- HPA configured for auto-scaling

## 🛠️ Development

### Project Structure

```
agent-bruno/
├── src/
│   ├── agent/
│   │   ├── core.py           # Agent core logic
│   │   └── knowledge.py       # Homepage knowledge base
│   ├── memory/
│   │   ├── manager.py         # Memory management
│   │   ├── redis_store.py     # Redis session store
│   │   └── mongo_store.py     # MongoDB persistent store
│   ├── mcp/
│   │   └── server.py          # MCP server implementation
│   └── api/
│       ├── main.py            # FastAPI application
│       └── routes.py          # API routes
├── tests/
├── k8s/
│   ├── deployment.yaml
│   ├── service.yaml
│   └── kustomization.yaml
├── Dockerfile
└── pyproject.toml
```

### Running Tests

```bash
# All tests
pytest -v

# With coverage
pytest --cov=src --cov-report=html

# Specific test
pytest tests/test_memory.py -v
```

## 📞 Contact

- **GitHub**: [brunovlucena](https://github.com/brunovlucena)
- **LinkedIn**: [Bruno Lucena](https://www.linkedin.com/in/bvlucena)

---

**Version**: 0.1.0  
**Status**: 🚧 In Development  
**Last Updated**: 2025-10-11

