# 📱 AgentChat - Private WhatsApp for AI Agents

[![Version](https://img.shields.io/badge/version-1.0.0-blue.svg)](https://github.com/brunovlucena/homelab)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

---

## 🎯 Overview

**AgentChat** is a private WhatsApp-like messaging infrastructure where AI agents serve as intelligent personal assistants. Built on the Knative Lambda Operator, each user gets a dedicated agent-assistant with powerful capabilities.

### Key Features

| Feature | Description |
|---------|-------------|
| 🗣️ **Voice Cloning** | Record your voice, create a digital twin for agent responses |
| 🖼️ **Image Generation** | Generate images on your behalf via AI (Stable Diffusion) |
| 🎬 **Video Generation** | Create short videos using AI models |
| 📍 **Location Alerts** | Get notified when contacts are nearby |
| 🤖 **Personal AI Assistant** | Each user gets a dedicated agent assistant |
| 🎛️ **Command Center** | Admin dashboard for platform management |

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           AGENTCHAT PLATFORM                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  📱 CLIENTS                        🚪 GATEWAY                                │
│  ┌──────────┐  ┌──────────┐       ┌────────────────────┐                    │
│  │ iOS App  │  │ Web C&C  │ ───── │ Messaging Gateway  │                    │
│  └──────────┘  └──────────┘       └─────────┬──────────┘                    │
│                                              │ CloudEvents                   │
│  🤖 LAMBDA AGENTS (Knative)                  ▼                               │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────────┐  │
│  │Messaging │ │  Voice   │ │  Media   │ │ Location │ │ Agent-Assistant  │  │
│  │   Hub    │ │  Agent   │ │  Agent   │ │  Agent   │ │ (per user)       │  │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────────────┘  │
│                                              │                               │
│  💾 DATA                                     ▼                               │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐                        │
│  │PostgreSQL│ │  Redis   │ │  MinIO   │ │ RabbitMQ │                        │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘                        │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 📂 Project Structure

```
agent-chat/
├── docs/
│   └── ARCHITECTURE.md      # Detailed system design
├── k8s/
│   ├── kustomize/
│   │   ├── base/
│   │   │   ├── namespace.yaml
│   │   │   ├── rbac.yaml
│   │   │   ├── configmap-system.yaml
│   │   │   ├── secrets.yaml
│   │   │   ├── lambdaagent-messaging-hub.yaml
│   │   │   ├── lambdaagent-voice.yaml
│   │   │   ├── lambdaagent-media.yaml
│   │   │   ├── lambdaagent-location.yaml
│   │   │   ├── lambdaagent-command-center.yaml
│   │   │   ├── lambdaagent-assistant-template.yaml
│   │   │   └── kustomization.yaml
│   │   ├── pro/             # Development overlay
│   │   └── studio/          # Production overlay
│   └── tests/               # k6 load tests
├── src/
│   ├── messaging-hub/       # Message routing agent
│   ├── voice-agent/         # Voice cloning & TTS
│   ├── media-agent/         # Image/video generation
│   ├── location-agent/      # Proximity tracking
│   └── shared/              # Common utilities
├── web-command-center/      # Admin dashboard (Next.js)
├── ios-client/              # iOS app documentation
├── Makefile
├── VERSION
└── README.md
```

---

## 🚀 Quick Start

### Prerequisites

- Kubernetes cluster with Knative Lambda Operator
- PostgreSQL, Redis, MinIO, RabbitMQ
- Ollama (or other LLM provider)

### Deploy to Development

```bash
# Deploy base resources
make deploy-dev

# Or using kustomize directly
kubectl apply -k k8s/kustomize/pro/
```

### Deploy to Production

```bash
# Deploy with production settings
make deploy-prod

# Or using kustomize
kubectl apply -k k8s/kustomize/studio/
```

### Verify Deployment

```bash
# Check agents
kubectl get lambdaagents -n agent-chat

# Check services
kubectl get ksvc -n agent-chat

# View logs
kubectl logs -n agent-chat -l app.kubernetes.io/part-of=agent-chat -f
```

---

## 🤖 LambdaAgents

| Agent | Role | Description |
|-------|------|-------------|
| `messaging-hub` | Core | Central message routing and WebSocket management |
| `voice-agent` | Capability | Voice cloning, TTS, STT using XTTS/Whisper |
| `media-agent` | Capability | Image/video generation via Stable Diffusion |
| `location-agent` | Capability | Location tracking and proximity alerts |
| `command-center` | Orchestrator | Admin dashboard backend, user management |
| `agent-assistant-{user}` | Assistant | Per-user personal AI assistant (dynamically deployed) |

---

## 📊 CloudEvents

All communication uses CloudEvents v1.0:

| Event Type | Producer | Consumer |
|------------|----------|----------|
| `io.agentchat.message.sent` | Gateway | Agent-Assistant |
| `io.agentchat.message.response` | Agent-Assistant | Gateway |
| `io.agentchat.voice.sample.uploaded` | Gateway | Voice Agent |
| `io.agentchat.voice.clone.ready` | Voice Agent | Agent-Assistant |
| `io.agentchat.media.image.request` | Agent-Assistant | Media Agent |
| `io.agentchat.media.image.generated` | Media Agent | Gateway |
| `io.agentchat.location.updated` | Gateway | Location Agent |
| `io.agentchat.location.proximity.alert` | Location Agent | Agent-Assistant |

---

## 🎛️ Command Center

Web-based admin dashboard for platform management:

```bash
cd web-command-center
npm install
npm run dev  # http://localhost:3001
```

Features:
- 📊 Real-time dashboard with metrics
- 👥 User management
- 🤖 Agent monitoring and deployment
- 💬 Conversation inspector
- 🔔 Alert management
- ⚙️ System settings

---

## 📱 iOS Client

Native iOS app for end users:

- **SwiftUI** for modern UI
- **WebSocket** for real-time messaging
- **CoreLocation** for proximity features
- **AVFoundation** for voice recording

See [ios-client/README.md](ios-client/README.md) for details.

---

## 🔐 Security

| Feature | Implementation |
|---------|----------------|
| Authentication | JWT + Device tokens |
| Transport | TLS 1.3 |
| Messages | End-to-end encryption (Signal Protocol) |
| Voice Data | Encrypted at rest, user consent required |
| Location | Opt-in, configurable sharing radius |
| Admin | RBAC, audit logging |

---

## 📈 Observability

- **Metrics**: Prometheus + Grafana dashboards
- **Tracing**: OpenTelemetry → Alloy → Tempo
- **Logging**: JSON logs → Loki
- **Alerts**: PrometheusRule → Alertmanager

---

## 🔧 Configuration

Environment variables in ConfigMap:

```yaml
VOICE_CLONING_ENABLED: "true"
IMAGE_GENERATION_ENABLED: "true"
LOCATION_TRACKING_ENABLED: "true"
DEFAULT_PROXIMITY_RADIUS_KM: "5"
LLM_MODEL: "llama3.2:3b"
```

---

## 📚 Documentation

- [Architecture Design](docs/ARCHITECTURE.md)
- [CloudEvents Specification](../knative-lambda-operator/docs/04-architecture/CLOUDEVENTS_SPECIFICATION.md)
- [LambdaAgent CRD](../knative-lambda-operator/k8s/base/crd-lambdaagent.yaml)

---

## 🤝 Contributing

1. Fork the repository
2. Create feature branch
3. Implement changes with tests
4. Submit pull request

---

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

---

**Maintainer**: Bruno Lucena  
**Version**: 1.0.0
