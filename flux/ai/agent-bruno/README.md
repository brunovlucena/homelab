# 🤖 Agent-Bruno

**AI-Powered Chatbot for Homelab Homepage**

A conversational AI assistant deployed as a serverless function on Knative, providing an interactive chatbot experience on the homelab homepage. **Now with CloudEvents integration for cross-agent communication!**

## 🎯 Overview

Agent-Bruno is a lightweight chatbot that:
- **Answers questions** about the homelab infrastructure
- **Provides assistance** with common tasks
- **Integrates** with the homepage frontend
- **Uses local LLM** (Ollama) for privacy-first AI
- **Communicates with other agents** via CloudEvents (NEW!)

## 📋 Quick Start

```bash
# Install dependencies
make install

# Run locally
make run-dev

# Test the chatbot
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "What can you help me with?"}'
```

## 🏗️ Architecture

```
┌──────────────────────────────────────────────────────────────────────────┐
│                            HOMELAB CLUSTER                                │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                           │
│   ┌─────────────┐         ┌─────────────┐         ┌─────────────┐        │
│   │  Homepage   │────────▶│ Agent-Bruno │────────▶│   Ollama    │        │
│   │  Frontend   │         │  (Chatbot)  │         │   (LLM)     │        │
│   └─────────────┘         └──────┬──────┘         └─────────────┘        │
│         │                        │                                        │
│         │              ┌─────────┴─────────┐                              │
│         ▼              ▼                   ▼                              │
│   ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                      │
│   │ CloudFlare  │  │  RabbitMQ   │  │ Prometheus  │                      │
│   │   Tunnel    │  │(CloudEvents)│  │  (Metrics)  │                      │
│   └─────────────┘  └──────┬──────┘  └─────────────┘                      │
│                           │                                               │
│              ┌────────────┴────────────┐                                  │
│              ▼                         ▼                                  │
│   ┌──────────────────┐    ┌──────────────────┐                           │
│   │ Agent-Contracts  │    │  Alertmanager    │                           │
│   │ (vuln scanning)  │    │   (alerts)       │                           │
│   └──────────────────┘    └──────────────────┘                           │
│                                                                           │
└──────────────────────────────────────────────────────────────────────────┘
```

## 📡 CloudEvents Integration

Agent-Bruno participates in the homelab's event-driven architecture, enabling real-time awareness of security events and cross-agent communication.

### Events Emitted

| Event Type | Trigger | Purpose |
|------------|---------|---------|
| `io.homelab.chat.message` | Every chat | Analytics and logging |
| `io.homelab.chat.intent.security` | User asks about security | Cross-agent awareness |
| `io.homelab.chat.intent.status` | User asks about service status | Monitoring integration |

### Events Received

| Event Type | Source | Effect |
|------------|--------|--------|
| `io.homelab.vuln.found` | agent-contracts | Chatbot aware of vulnerabilities |
| `io.homelab.exploit.validated` | agent-contracts | Critical security notification |
| `io.homelab.alert.fired` | alertmanager | System alerts available to users |

### How It Works

1. **User asks about security** → Intent detected → Event emitted → Other agents notified
2. **Agent-contracts finds vulnerability** → Event sent → Bruno stores notification → User informed in next chat
3. **Critical exploit validated** → Bruno receives event → Can proactively warn users

## 📁 Project Structure

```
agent-bruno/
├── src/
│   ├── chatbot/
│   │   ├── __init__.py
│   │   ├── main.py          # FastAPI entry point
│   │   ├── handler.py       # Chat handler with Ollama
│   │   └── Dockerfile
│   ├── shared/
│   │   ├── __init__.py
│   │   ├── types.py         # Shared types
│   │   └── metrics.py       # Prometheus metrics
│   └── requirements.txt
├── k8s/
│   └── kustomize/
│       ├── base/
│       │   ├── kustomization.yaml
│       │   ├── namespace.yaml
│       │   ├── service.yaml
│       │   └── configmap.yaml
│       ├── studio/
│       └── pro/
├── tests/
│   ├── unit/
│   └── conftest.py
├── Makefile
└── README.md
```

## 🔧 Configuration

| Environment Variable | Description | Default |
|---------------------|-------------|---------|
| `OLLAMA_URL` | Ollama LLM endpoint | `http://ollama.ai-inference.svc.cluster.local:11434` |
| `OLLAMA_MODEL` | Default model for chat | `llama3.2:3b` |
| `SYSTEM_PROMPT` | System prompt for personality | See config |
| `MAX_CONTEXT_LENGTH` | Max conversation context | `4096` |
| `CORS_ORIGINS` | Allowed CORS origins | `*` |
| `EMIT_EVENTS` | Enable CloudEvent emission | `true` |
| `KNATIVE_BROKER_URL` | Broker ingress URL | `http://agent-bruno-broker-ingress.agent-bruno.svc.cluster.local` |

## 🚀 Deployment

### Prerequisites

- Ollama deployed in `ai-inference` namespace
- Knative Serving installed
- Homepage frontend deployed

### Deploy to Homelab

```bash
# Build image
make build

# Push to registry
make push

# Deploy to Kubernetes
make deploy-studio
```

## 📊 API Endpoints

### POST /chat
Send a message and get a response.

```json
{
  "message": "Hello, how can you help?",
  "conversation_id": "optional-uuid"
}
```

Response:
```json
{
  "response": "Hello! I'm Agent-Bruno, your homelab assistant...",
  "conversation_id": "uuid",
  "tokens_used": 128,
  "model": "llama3.2:3b",
  "duration_ms": 1234.5
}
```

### POST /events
Receive CloudEvents from Knative triggers (internal use).

```bash
# Example: Sending a test event
curl -X POST http://agent-bruno/events \
  -H "Content-Type: application/cloudevents+json" \
  -d '{
    "specversion": "1.0",
    "type": "io.homelab.vuln.found",
    "source": "/test",
    "data": {"chain": "ethereum", "address": "0x123...", "max_severity": "high"}
  }'
```

### GET /notifications
Get recent notifications from other agents.

```json
{
  "count": 2,
  "notifications": [
    {
      "type": "vulnerability",
      "severity": "critical",
      "chain": "ethereum",
      "message": "🔴 CRITICAL vulnerability found...",
      "timestamp": "2024-01-15T10:30:00Z"
    }
  ]
}
```

### DELETE /notifications
Clear stored notifications.

### GET /health
Health check endpoint.

### GET /metrics
Prometheus metrics endpoint.

## 📈 Monitoring

Metrics exposed:
- `agent_bruno_messages_total{status}` - Total messages processed
- `agent_bruno_response_duration_seconds` - Response latency
- `agent_bruno_tokens_used_total{model}` - LLM tokens consumed
- `agent_bruno_active_conversations` - Active conversation count
- `agent_bruno_events_emitted_total{event_type, status}` - CloudEvents sent
- `agent_bruno_events_received_total{event_type, status}` - CloudEvents received

## 🔒 Security

- No external API calls (fully local with Ollama)
- Rate limiting per conversation
- Conversation context automatically pruned
- No PII stored or logged

## 📄 License

MIT
