# 🏗️ Speech Coach Agent Architecture

## Overview

The Speech Coach Agent is a personal AI agent designed to help autistic children develop speech skills through interactive games and exercises. It follows the homelab agent patterns and integrates with the AgentApp mobile framework.

## Architecture Components

### 1. Backend Agent (`agent-speech-coach`)

**Location**: `/flux/ai/agent-speech-coach/`

- **Type**: LambdaAgent (Knative service)
- **Technology**: Python 3.11 + FastAPI
- **AI Provider**: Ollama (SLM)
- **Database**: MongoDB (for progress tracking)
- **Protocol**: CloudEvents 1.0

**Key Features**:
- Exercise management (word repetition, phrase completion, story telling, etc.)
- Progress tracking and analytics
- Game session management
- Personalized coaching responses
- Integration with homelab mobile-api

### 2. Mobile API Integration

**Location**: `/flux/services/homelab-services/mobile-api/`

The mobile-api acts as a router/gateway for mobile apps:
- Routes `speech-coach-agent` messages to the agent
- Handles CloudEvents protocol
- Provides service discovery

### 3. iOS Mobile App

**Location**: `/flux/ai/agent-speech-coach/ios-app/`

- **Framework**: SwiftUI + AgentApp patterns
- **Features**:
  - Face recognition (AVFoundation)
  - Speech recognition (Speech framework)
  - Chat interface with agent
  - Progress tracking
  - Customizable themes

## Data Flow

```
┌─────────────────────────────────┐  ┌─────────────────────────────────┐
│   📱 iOS App (SpeechCoach)     │  │   🍓 Raspberry Pi Client        │
│  • Face Recognition             │  │  • Web Interface               │
│  • Speech Recognition           │  │  • Camera Support              │
│  • Chat UI                      │  │  • Microphone                  │
└──────────────┬──────────────────┘  └──────────────┬──────────────────┘
               │                                     │
               │ CloudEvents (HTTPS)                 │ CloudEvents (HTTPS)
               │ agentId: "speech-coach-agent"       │ agentId: "speech-coach-agent"
               └──────────────┬──────────────────────┘
                              ▼
               ┌─────────────────────────────────┐
               │  🌐 Mobile API (Router)        │
               │  • Routes to speech-coach       │
               │  • Service discovery            │
               │  • CloudEvents gateway          │
               └──────────────┬──────────────────┘
                              │
                              │ CloudEvents
                              │ (via Knative Eventing)
                              ▼
               ┌─────────────────────────────────┐
               │  🖥️ Studio Cluster             │
               │                                 │
               │  ┌───────────────────────────┐ │
               │  │ 🤖 Speech Coach Agent     │ │
               │  │ • Exercise logic          │ │
               │  │ • Progress tracking       │ │
               │  │ • LLM coaching responses  │ │
               │  │ • MongoDB storage         │ │
               │  └───────────────────────────┘ │
               └─────────────────────────────────┘
```

### Clientes Disponíveis

1. **iOS App** (Smartphone)
   - Interface nativa SwiftUI
   - Reconhecimento facial via AVFoundation
   - Reconhecimento de voz via Speech framework
   - Conecta via VPN ou Cloudflare Tunnel

2. **Raspberry Pi Client** (Web Interface)
   - Interface web acessível via navegador
   - Suporte para câmera USB
   - Microfone USB ou GPIO
   - Executa localmente no Raspberry Pi
   - Conecta ao agente no servidor studio

## Deployment

### Prerequisites

1. MongoDB (for progress tracking)
2. Ollama (for SLM inference)
3. Knative Lambda controller (for LambdaAgent)
4. RabbitMQ (for eventing)

### Steps

1. **Build and push Docker image**:
   ```bash
   cd flux/ai/agent-speech-coach
   make build
   docker push ghcr.io/brunovlucena/speech-coach-agent:latest
   ```

2. **Deploy to cluster**:
   ```bash
   kubectl apply -k k8s/kustomize/studio/
   ```

3. **Verify deployment**:
   ```bash
   kubectl get lambdaagent -n agent-speech-coach
   kubectl get pods -n agent-speech-coach
   ```

## Configuration

### Environment Variables

- `MONGODB_URL`: MongoDB connection string
- `MONGODB_DATABASE`: Database name (default: `speech_coach_db`)
- `OLLAMA_URL`: Ollama endpoint (default: `http://ollama-native.ollama.svc.cluster.local:11434`)
- `OLLAMA_MODEL`: Model to use (default: `llama3.2:3b`)
- `EVENT_SOURCE`: CloudEvent source (default: `/agent-speech-coach/games`)

### Secrets

Create a secret for database credentials:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: speech-coach-db-credentials
  namespace: agent-speech-coach
type: Opaque
stringData:
  url: "mongodb://..."
```

## Mobile App Setup

1. **Open Xcode project**:
   ```bash
   cd ios-app/SpeechCoach
   open SpeechCoach.xcodeproj
   ```

2. **Configure agent URL**:
   - Update `Agent.swift` with your homelab mobile-api URL
   - Or use VPN to access cluster-local services

3. **Build and run**:
   - Select target device
   - Press Cmd+R

## Privacy & Security

- **Face Recognition**: Runs entirely on-device using AVFoundation
- **Speech Recognition**: Uses iOS Speech framework (on-device)
- **Data Storage**: All progress data stored in your homelab MongoDB
- **Communication**: Encrypted via HTTPS/mTLS
- **No External Services**: Everything runs in your homelab

## Future Enhancements

- [ ] Offline mode with local SLM
- [ ] More exercise types
- [ ] Parent/guardian dashboard
- [ ] Progress reports and analytics
- [ ] Multi-language support
- [ ] Custom exercise builder
