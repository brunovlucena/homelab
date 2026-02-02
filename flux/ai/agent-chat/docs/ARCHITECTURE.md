# 📱 AgentChat - Private WhatsApp for AI Agents

**Version**: 1.0.0  
**Last Updated**: December 10, 2025  
**Status**: Architecture Design

---

## 🎯 Overview

AgentChat is a private WhatsApp-like messaging infrastructure where **AI agents serve as intelligent assistants**. Users interact with their personal agent-assistant via an iOS app, while the agents have powerful capabilities including voice cloning, media generation, and location-based social alerts.

### Key Features

| Feature | Description |
|---------|-------------|
| 🗣️ **Voice Recording & Cloning** | Record user voice, create voice doubles for agent responses |
| 🖼️ **Image Generation** | Generate images on behalf of users via AI models |
| 🎬 **Video Generation** | Create videos using AI generation capabilities |
| 📍 **Location-Based Alerts** | Notify contacts when users are nearby |
| 🤖 **Agent Assistants** | AI agents as personal assistants in conversations |
| 🎛️ **Command & Control Center** | Admin dashboard for managing agents and users |

---

## 🏗️ System Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           AGENTCHAT PLATFORM                                     │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  ┌─────────────────────────────────────────────────────────────────────────┐    │
│  │  📱 CLIENT LAYER                                                         │    │
│  │                                                                          │    │
│  │  ┌──────────────┐    ┌──────────────┐    ┌──────────────────────────┐   │    │
│  │  │  iOS App     │    │  Web C&C     │    │  Agent Web Interface    │   │    │
│  │  │  (Swift)     │    │  (Next.js)   │    │  (Next.js - existing)   │   │    │
│  │  │              │    │              │    │                          │   │    │
│  │  │  • Chat UI   │    │  • Dashboard │    │  • Agent monitoring     │   │    │
│  │  │  • Voice     │    │  • User mgmt │    │  • Event feeds          │   │    │
│  │  │  • Media     │    │  • Agents    │    │  • Chat interface       │   │    │
│  │  │  • Location  │    │  • Analytics │    │                          │   │    │
│  │  └──────────────┘    └──────────────┘    └──────────────────────────┘   │    │
│  └────────────────────────────┬─────────────────────────────────────────────┘    │
│                               │                                                  │
│                               │ WebSocket / HTTP / CloudEvents                   │
│                               ▼                                                  │
│  ┌─────────────────────────────────────────────────────────────────────────┐    │
│  │  🚪 GATEWAY LAYER                                                        │    │
│  │                                                                          │    │
│  │  ┌──────────────────────────────────────────────────────────────────┐   │    │
│  │  │  Messaging Gateway (Knative Service)                              │   │    │
│  │  │  ├─ WebSocket handler (real-time chat)                           │   │    │
│  │  │  ├─ REST API (chat history, user management)                      │   │    │
│  │  │  ├─ CloudEvents ingress (agent communication)                     │   │    │
│  │  │  ├─ Authentication (JWT + device tokens)                          │   │    │
│  │  │  └─ Rate limiting & abuse prevention                              │   │    │
│  │  └──────────────────────────────────────────────────────────────────┘   │    │
│  └────────────────────────────┬─────────────────────────────────────────────┘    │
│                               │                                                  │
│                               │ CloudEvents (RabbitMQ)                           │
│                               ▼                                                  │
│  ┌─────────────────────────────────────────────────────────────────────────┐    │
│  │  🤖 AGENT LAYER (LambdaAgents)                                          │    │
│  │                                                                          │    │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌─────────────┐  │    │
│  │  │ Messaging    │  │ Voice Agent  │  │ Media Agent  │  │ Location    │  │    │
│  │  │ Hub          │  │              │  │              │  │ Agent       │  │    │
│  │  │              │  │              │  │              │  │             │  │    │
│  │  │ • Message    │  │ • Record     │  │ • Image gen  │  │ • Track     │  │    │
│  │  │   routing    │  │ • Clone      │  │ • Video gen  │  │   location  │  │    │
│  │  │ • History    │  │ • TTS        │  │ • Transform  │  │ • Proximity │  │    │
│  │  │ • Presence   │  │ • STT        │  │ • Filters    │  │   alerts    │  │    │
│  │  │ • Typing     │  │ • Voice ID   │  │              │  │ • Contacts  │  │    │
│  │  └──────────────┘  └──────────────┘  └──────────────┘  └─────────────┘  │    │
│  │                                                                          │    │
│  │  ┌──────────────────────────────────────────────────────────────────┐   │    │
│  │  │  Agent-Assistant (Per User) - The main chat companion             │   │    │
│  │  │  ├─ Personal AI assistant with user context                      │   │    │
│  │  │  ├─ Coordinates with Voice, Media, Location agents               │   │    │
│  │  │  ├─ Learns user preferences and communication style              │   │    │
│  │  │  └─ Can act on behalf of user (with consent)                     │   │    │
│  │  └──────────────────────────────────────────────────────────────────┘   │    │
│  └────────────────────────────┬─────────────────────────────────────────────┘    │
│                               │                                                  │
│                               │ CloudEvents                                      │
│                               ▼                                                  │
│  ┌─────────────────────────────────────────────────────────────────────────┐    │
│  │  🧠 AI SERVICES LAYER                                                    │    │
│  │                                                                          │    │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌─────────────┐  │    │
│  │  │ Ollama       │  │ Voice Clone  │  │ Stable       │  │ Video Gen   │  │    │
│  │  │ (LLM)        │  │ (XTTS/RVC)   │  │ Diffusion    │  │ (optional)  │  │    │
│  │  │              │  │              │  │              │  │             │  │    │
│  │  │ llama3.2:3b  │  │ Voice        │  │ Image        │  │ Stable      │  │    │
│  │  │ or Claude    │  │ synthesis    │  │ generation   │  │ Video       │  │    │
│  │  └──────────────┘  └──────────────┘  └──────────────┘  └─────────────┘  │    │
│  └────────────────────────────┬─────────────────────────────────────────────┘    │
│                               │                                                  │
│                               ▼                                                  │
│  ┌─────────────────────────────────────────────────────────────────────────┐    │
│  │  💾 DATA LAYER                                                           │    │
│  │                                                                          │    │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌─────────────┐  │    │
│  │  │ PostgreSQL   │  │ Redis        │  │ MinIO        │  │ RabbitMQ    │  │    │
│  │  │              │  │              │  │              │  │             │  │    │
│  │  │ • Users      │  │ • Sessions   │  │ • Media      │  │ • Events    │  │    │
│  │  │ • Messages   │  │ • Presence   │  │ • Voices     │  │ • Queues    │  │    │
│  │  │ • Contacts   │  │ • Cache      │  │ • Images     │  │ • DLQ       │  │    │
│  │  │ • Locations  │  │ • Pub/Sub    │  │ • Videos     │  │             │  │    │
│  │  └──────────────┘  └──────────────┘  └──────────────┘  └─────────────┘  │    │
│  └─────────────────────────────────────────────────────────────────────────┘    │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## 📊 Data Flow Diagrams

### 1. User Message Flow

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                        USER MESSAGE FLOW                                         │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  User (iOS App)                                                                  │
│       │                                                                          │
│       │ 1. Send message (text/voice/image)                                      │
│       ▼                                                                          │
│  ┌─────────────────────────────────────────────────────────────────────────┐    │
│  │  Messaging Gateway                                                       │    │
│  │  ├─ 2. Authenticate user (JWT)                                          │    │
│  │  ├─ 3. Validate message format                                          │    │
│  │  ├─ 4. Store message in PostgreSQL                                      │    │
│  │  ├─ 5. Emit CloudEvent: io.agentchat.message.sent                       │    │
│  │  └─ 6. Broadcast to recipient via WebSocket                             │    │
│  └────────────────────┬────────────────────────────────────────────────────┘    │
│                       │                                                          │
│                       │ CloudEvent: io.agentchat.message.sent                    │
│                       ▼                                                          │
│  ┌─────────────────────────────────────────────────────────────────────────┐    │
│  │  Agent-Assistant (User's Personal Agent)                                 │    │
│  │  ├─ 7. Receive message event                                            │    │
│  │  ├─ 8. Analyze intent (LLM)                                             │    │
│  │  ├─ 9. Determine required actions                                        │    │
│  │  │     ├─ Voice? → Forward to Voice Agent                               │    │
│  │  │     ├─ Image? → Forward to Media Agent                               │    │
│  │  │     └─ Location? → Check Location Agent                              │    │
│  │  ├─ 10. Generate response (LLM)                                         │    │
│  │  └─ 11. Emit CloudEvent: io.agentchat.message.response                  │    │
│  └────────────────────┬────────────────────────────────────────────────────┘    │
│                       │                                                          │
│                       │ CloudEvent: io.agentchat.message.response                │
│                       ▼                                                          │
│  ┌─────────────────────────────────────────────────────────────────────────┐    │
│  │  Messaging Gateway                                                       │    │
│  │  ├─ 12. Receive response event                                          │    │
│  │  ├─ 13. Store agent response                                            │    │
│  │  └─ 14. Send to user via WebSocket                                      │    │
│  └─────────────────────────────────────────────────────────────────────────┘    │
│                       │                                                          │
│                       ▼                                                          │
│       User receives agent response                                               │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 2. Voice Cloning Flow

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                        VOICE CLONING FLOW                                        │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  User (iOS App)                                                                  │
│       │                                                                          │
│       │ 1. Record voice sample (30s-3min)                                       │
│       ▼                                                                          │
│  ┌─────────────────────────────────────────────────────────────────────────┐    │
│  │  Messaging Gateway                                                       │    │
│  │  ├─ 2. Upload audio to MinIO                                            │    │
│  │  └─ 3. Emit CloudEvent: io.agentchat.voice.sample.uploaded              │    │
│  └────────────────────┬────────────────────────────────────────────────────┘    │
│                       │                                                          │
│                       │ CloudEvent                                               │
│                       ▼                                                          │
│  ┌─────────────────────────────────────────────────────────────────────────┐    │
│  │  Voice Agent (LambdaAgent)                                               │    │
│  │  ├─ 4. Fetch audio from MinIO                                           │    │
│  │  ├─ 5. Process with XTTS/RVC model                                      │    │
│  │  │     ├─ Extract voice characteristics                                 │    │
│  │  │     ├─ Create voice embedding                                        │    │
│  │  │     └─ Store voice model                                             │    │
│  │  ├─ 6. Generate test audio with cloned voice                            │    │
│  │  └─ 7. Emit CloudEvent: io.agentchat.voice.clone.ready                  │    │
│  └────────────────────┬────────────────────────────────────────────────────┘    │
│                       │                                                          │
│                       ▼                                                          │
│       Agent-Assistant can now speak with user's voice clone                      │
│       (for sending voice messages on behalf of user)                             │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 3. Location-Based Alert Flow

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                     LOCATION-BASED ALERT FLOW                                    │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  User A (Traveling)                     User B (Contact)                         │
│       │                                      ▲                                   │
│       │ 1. Location update                   │ 8. Notification                   │
│       ▼                                      │                                   │
│  ┌─────────────────────────────────────────────────────────────────────────┐    │
│  │  Messaging Gateway                                                       │    │
│  │  ├─ 2. Store location in Redis (ephemeral)                              │    │
│  │  └─ 3. Emit CloudEvent: io.agentchat.location.updated                   │    │
│  └────────────────────┬────────────────────────────────────────────────────┘    │
│                       │                                                          │
│                       │ CloudEvent                                               │
│                       ▼                                                          │
│  ┌─────────────────────────────────────────────────────────────────────────┐    │
│  │  Location Agent (LambdaAgent)                                            │    │
│  │  ├─ 4. Receive location event                                           │    │
│  │  ├─ 5. Query contacts' locations from Redis                             │    │
│  │  ├─ 6. Calculate proximity (configurable radius)                        │    │
│  │  ├─ 7. If within radius AND contact has alerts enabled:                 │    │
│  │  │     └─ Emit CloudEvent: io.agentchat.location.proximity.alert        │    │
│  │  └─ 8. Store proximity history                                          │    │
│  └────────────────────┬────────────────────────────────────────────────────┘    │
│                       │                                                          │
│                       │ CloudEvent: io.agentchat.location.proximity.alert        │
│                       ▼                                                          │
│  ┌─────────────────────────────────────────────────────────────────────────┐    │
│  │  User B's Agent-Assistant                                                │    │
│  │  ├─ 9. Receive proximity alert                                          │    │
│  │  ├─ 10. Generate friendly notification                                  │    │
│  │  │      "Hey! Bruno is in your area (São Paulo) - want to meet up?"     │    │
│  │  └─ 11. Send notification to User B                                     │    │
│  └─────────────────────────────────────────────────────────────────────────┘    │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## 📦 CloudEvents Specification

### Event Types

| Event Type | Description | Producer | Consumer |
|------------|-------------|----------|----------|
| **Messaging** | | | |
| `io.agentchat.message.sent` | User sent a message | Gateway | Agent-Assistant |
| `io.agentchat.message.response` | Agent response | Agent-Assistant | Gateway |
| `io.agentchat.message.delivered` | Message delivered | Gateway | Sender |
| `io.agentchat.message.read` | Message was read | Gateway | Sender |
| `io.agentchat.typing.started` | User started typing | Gateway | Recipients |
| `io.agentchat.typing.stopped` | User stopped typing | Gateway | Recipients |
| **Voice** | | | |
| `io.agentchat.voice.sample.uploaded` | Voice sample uploaded | Gateway | Voice Agent |
| `io.agentchat.voice.clone.ready` | Voice clone ready | Voice Agent | Agent-Assistant |
| `io.agentchat.voice.message.request` | Request voice message | Agent-Assistant | Voice Agent |
| `io.agentchat.voice.message.generated` | Voice message ready | Voice Agent | Gateway |
| `io.agentchat.voice.transcription.request` | Transcribe audio | Gateway | Voice Agent |
| `io.agentchat.voice.transcription.completed` | Transcription ready | Voice Agent | Gateway |
| **Media** | | | |
| `io.agentchat.media.image.request` | Generate image | Agent-Assistant | Media Agent |
| `io.agentchat.media.image.generated` | Image ready | Media Agent | Gateway |
| `io.agentchat.media.video.request` | Generate video | Agent-Assistant | Media Agent |
| `io.agentchat.media.video.generated` | Video ready | Media Agent | Gateway |
| **Location** | | | |
| `io.agentchat.location.updated` | User location update | Gateway | Location Agent |
| `io.agentchat.location.proximity.alert` | Contact nearby | Location Agent | Agent-Assistant |
| `io.agentchat.location.geofence.enter` | Entered geofence | Location Agent | Agent-Assistant |
| `io.agentchat.location.geofence.exit` | Exited geofence | Location Agent | Agent-Assistant |
| **Admin** | | | |
| `io.agentchat.admin.user.created` | New user registered | Gateway | C&C |
| `io.agentchat.admin.agent.deployed` | Agent deployed | Operator | C&C |
| `io.agentchat.admin.alert.raised` | System alert | Any | C&C |

### Event Data Schema (Example)

```json
{
  "specversion": "1.0",
  "id": "msg-123e4567-e89b-12d3-a456-426614174000",
  "source": "/agentchat/gateway/production",
  "type": "io.agentchat.message.sent",
  "subject": "users/user-123/chats/chat-456",
  "time": "2025-12-10T10:30:00Z",
  "datacontenttype": "application/json",
  "data": {
    "messageId": "msg-abc123",
    "senderId": "user-123",
    "chatId": "chat-456",
    "content": {
      "type": "text",
      "text": "Hey! Can you generate an image of a sunset?",
      "attachments": []
    },
    "metadata": {
      "deviceId": "ios-device-xyz",
      "appVersion": "1.0.0",
      "location": {
        "lat": -23.5505,
        "lng": -46.6333,
        "city": "São Paulo"
      }
    }
  }
}
```

---

## 🤖 LambdaAgent Definitions

### 1. Agent-Assistant (Template per User)

Each user gets their own Agent-Assistant instance with personalized configuration:

```yaml
apiVersion: lambda.knative.io/v1alpha1
kind: LambdaAgent
metadata:
  name: agent-assistant-{user-id}
  namespace: agent-chat
spec:
  image:
    repository: localhost:5001/agent-chat/assistant
    tag: "v1.0.0"
    port: 8080
  
  ai:
    provider: ollama
    endpoint: "http://ollama-native.ollama.svc.cluster.local:11434"
    model: "llama3.2:3b"
    maxTokens: 4096
    temperature: "0.7"
  
  behavior:
    maxContextMessages: 50
    emitEvents: true
    systemPrompt: |
      You are {user_name}'s personal AI assistant in AgentChat.
      
      YOUR CAPABILITIES:
      - Have natural conversations
      - Generate images on user's behalf (via Media Agent)
      - Send voice messages in user's cloned voice (via Voice Agent)
      - Alert contacts when user is nearby (via Location Agent)
      
      USER PREFERENCES:
      {user_preferences}
      
      COMMUNICATION STYLE:
      {user_style}
      
      IMPORTANT RULES:
      - Always be helpful and friendly
      - Respect user privacy settings
      - Ask for confirmation before sending messages on user's behalf
      - Keep conversations contextual and personal
  
  eventing:
    enabled: true
    eventSource: "/agent-chat/assistant/{user-id}"
    intents:
      - io.agentchat.message.response
      - io.agentchat.voice.message.request
      - io.agentchat.media.image.request
    subscriptions:
      - eventType: io.agentchat.message.sent
      - eventType: io.agentchat.voice.clone.ready
      - eventType: io.agentchat.location.proximity.alert
```

### 2. Messaging Hub (Singleton)

Handles all message routing and delivery:

```yaml
apiVersion: lambda.knative.io/v1alpha1
kind: LambdaAgent
metadata:
  name: messaging-hub
  namespace: agent-chat
spec:
  image:
    repository: localhost:5001/agent-chat/messaging-hub
    tag: "v1.0.0"
  
  scaling:
    minReplicas: 2
    maxReplicas: 10
    targetConcurrency: 100
  
  eventing:
    subscriptions:
      - eventType: io.agentchat.message.*
      - eventType: io.agentchat.typing.*
```

### 3. Voice Agent

Handles voice cloning, TTS, and STT:

```yaml
apiVersion: lambda.knative.io/v1alpha1
kind: LambdaAgent
metadata:
  name: voice-agent
  namespace: agent-chat
spec:
  image:
    repository: localhost:5001/agent-chat/voice-agent
    tag: "v1.0.0"
  
  resources:
    requests:
      cpu: "500m"
      memory: "2Gi"
    limits:
      cpu: "2000m"
      memory: "4Gi"
  
  env:
    - name: XTTS_MODEL_PATH
      value: "/models/xtts"
    - name: WHISPER_MODEL
      value: "base"
```

### 4. Media Agent

Handles image and video generation:

```yaml
apiVersion: lambda.knative.io/v1alpha1
kind: LambdaAgent
metadata:
  name: media-agent
  namespace: agent-chat
spec:
  image:
    repository: localhost:5001/agent-chat/media-agent
    tag: "v1.0.0"
  
  resources:
    requests:
      cpu: "1000m"
      memory: "4Gi"
    limits:
      cpu: "4000m"
      memory: "8Gi"
  
  env:
    - name: STABLE_DIFFUSION_URL
      value: "http://stable-diffusion.ai-services:7860"
```

### 5. Location Agent

Handles location tracking and proximity alerts:

```yaml
apiVersion: lambda.knative.io/v1alpha1
kind: LambdaAgent
metadata:
  name: location-agent
  namespace: agent-chat
spec:
  image:
    repository: localhost:5001/agent-chat/location-agent
    tag: "v1.0.0"
  
  env:
    - name: REDIS_URL
      value: "redis://redis.agent-chat:6379"
    - name: DEFAULT_PROXIMITY_RADIUS_KM
      value: "5"
```

---

## 🎛️ Command & Control Center

The C&C Dashboard provides administrative control over the entire AgentChat platform:

### Features

| Feature | Description |
|---------|-------------|
| 📊 **Dashboard** | Real-time metrics, user activity, agent health |
| 👥 **User Management** | Create, disable, configure users |
| 🤖 **Agent Management** | Deploy, scale, configure agents |
| 💬 **Chat Monitoring** | View conversations (with privacy controls) |
| 🔔 **Alerts** | System alerts, abuse detection |
| 📈 **Analytics** | Usage patterns, popular features |
| 🔐 **Security** | API keys, permissions, audit logs |

### Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                    COMMAND & CONTROL CENTER                                      │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  ┌────────────────────────────────────────────────────────────────────────────┐ │
│  │  Frontend (Next.js 14 + Tailwind CSS)                                      │ │
│  │  ├─ Dashboard.tsx          - Overview metrics & charts                     │ │
│  │  ├─ UserManagement.tsx     - User CRUD operations                          │ │
│  │  ├─ AgentMonitor.tsx       - Agent health & logs                           │ │
│  │  ├─ ChatViewer.tsx         - Conversation inspector                        │ │
│  │  ├─ AlertCenter.tsx        - Alert management                              │ │
│  │  ├─ Analytics.tsx          - Usage analytics                               │ │
│  │  └─ Settings.tsx           - System configuration                          │ │
│  └────────────────────────────────────────────────────────────────────────────┘ │
│                                      │                                           │
│                                      │ API Calls                                 │
│                                      ▼                                           │
│  ┌────────────────────────────────────────────────────────────────────────────┐ │
│  │  Backend API (Next.js API Routes + tRPC)                                   │ │
│  │  ├─ /api/users/*           - User management                               │ │
│  │  ├─ /api/agents/*          - Agent CRUD                                    │ │
│  │  ├─ /api/chats/*           - Chat history & monitoring                     │ │
│  │  ├─ /api/alerts/*          - Alert management                              │ │
│  │  ├─ /api/analytics/*       - Analytics queries                             │ │
│  │  └─ /api/cloudevents/*     - CloudEvents ingress                           │ │
│  └────────────────────────────────────────────────────────────────────────────┘ │
│                                      │                                           │
│                    ┌─────────────────┴─────────────────┐                        │
│                    ▼                                   ▼                        │
│  ┌─────────────────────────────────┐  ┌─────────────────────────────────────┐  │
│  │  Kubernetes API                 │  │  Data Services                       │  │
│  │  ├─ LambdaAgent CR management   │  │  ├─ PostgreSQL (users, chats)       │  │
│  │  ├─ Pod metrics                 │  │  ├─ Redis (sessions, cache)         │  │
│  │  └─ Log aggregation             │  │  └─ Prometheus (metrics)            │  │
│  └─────────────────────────────────┘  └─────────────────────────────────────┘  │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## 📱 iOS App Architecture

### Technology Stack

| Component | Technology |
|-----------|------------|
| Language | Swift 5.9 |
| UI Framework | SwiftUI |
| Architecture | MVVM + Clean Architecture |
| Networking | URLSession + Combine |
| WebSocket | Starscream |
| Local Storage | SwiftData (iOS 17+) |
| Push Notifications | APNs |
| Location | CoreLocation |
| Audio/Video | AVFoundation |

### App Structure

```
AgentChat/
├── App/
│   ├── AgentChatApp.swift
│   └── AppDelegate.swift
├── Core/
│   ├── Network/
│   │   ├── APIClient.swift
│   │   ├── WebSocketManager.swift
│   │   └── CloudEventsClient.swift
│   ├── Storage/
│   │   ├── UserDefaults+Extensions.swift
│   │   └── SwiftDataModels.swift
│   ├── Location/
│   │   └── LocationManager.swift
│   └── Audio/
│       ├── AudioRecorder.swift
│       └── AudioPlayer.swift
├── Features/
│   ├── Chat/
│   │   ├── Views/
│   │   │   ├── ChatListView.swift
│   │   │   ├── ChatDetailView.swift
│   │   │   └── MessageBubble.swift
│   │   ├── ViewModels/
│   │   │   ├── ChatListViewModel.swift
│   │   │   └── ChatDetailViewModel.swift
│   │   └── Models/
│   │       ├── Chat.swift
│   │       └── Message.swift
│   ├── Voice/
│   │   ├── Views/
│   │   │   ├── VoiceRecorderView.swift
│   │   │   └── VoiceCloneSetupView.swift
│   │   └── ViewModels/
│   │       └── VoiceViewModel.swift
│   ├── Media/
│   │   ├── Views/
│   │   │   ├── ImageGeneratorView.swift
│   │   │   └── MediaGalleryView.swift
│   │   └── ViewModels/
│   │       └── MediaViewModel.swift
│   ├── Location/
│   │   ├── Views/
│   │   │   ├── LocationSettingsView.swift
│   │   │   └── NearbyContactsView.swift
│   │   └── ViewModels/
│   │       └── LocationViewModel.swift
│   └── Settings/
│       ├── Views/
│       │   ├── SettingsView.swift
│       │   ├── AgentSettingsView.swift
│       │   └── PrivacySettingsView.swift
│       └── ViewModels/
│           └── SettingsViewModel.swift
└── Resources/
    ├── Assets.xcassets
    └── Info.plist
```

### Key Features Implementation

#### 1. Chat Interface
```swift
struct ChatDetailView: View {
    @StateObject var viewModel: ChatDetailViewModel
    
    var body: some View {
        VStack {
            // Messages list
            ScrollView {
                LazyVStack {
                    ForEach(viewModel.messages) { message in
                        MessageBubble(message: message)
                    }
                }
            }
            
            // Input area
            HStack {
                Button(action: viewModel.startVoiceRecording) {
                    Image(systemName: "mic.fill")
                }
                
                TextField("Message...", text: $viewModel.inputText)
                    .textFieldStyle(.roundedBorder)
                
                Button(action: viewModel.sendMessage) {
                    Image(systemName: "paperplane.fill")
                }
            }
            .padding()
        }
    }
}
```

#### 2. Voice Recording
```swift
class VoiceViewModel: ObservableObject {
    @Published var isRecording = false
    @Published var voiceCloneReady = false
    
    private let audioRecorder = AudioRecorder()
    private let apiClient: APIClient
    
    func recordVoiceSample() async throws {
        isRecording = true
        let audioData = try await audioRecorder.record(duration: 30)
        
        // Upload to backend
        try await apiClient.uploadVoiceSample(audioData)
        
        isRecording = false
    }
}
```

#### 3. Location Tracking
```swift
class LocationViewModel: ObservableObject {
    @Published var nearbyContacts: [Contact] = []
    @Published var locationEnabled = false
    
    private let locationManager = LocationManager()
    private let webSocket: WebSocketManager
    
    func startLocationUpdates() {
        locationManager.startUpdating { [weak self] location in
            self?.webSocket.send(
                CloudEvent(type: "io.agentchat.location.updated",
                          data: LocationData(lat: location.latitude,
                                           lng: location.longitude))
            )
        }
    }
}
```

---

## 🔐 Security Considerations

| Concern | Solution |
|---------|----------|
| Authentication | JWT tokens + Device registration |
| End-to-End Encryption | Signal Protocol for messages |
| Voice Data | Encrypted at rest, user consent required |
| Location Privacy | Opt-in, configurable sharing radius |
| API Security | Rate limiting, API keys, HTTPS |
| Admin Access | RBAC, audit logging |

---

## 📊 Observability

### Metrics

| Metric | Description |
|--------|-------------|
| `agentchat_messages_total` | Total messages sent |
| `agentchat_active_users` | Currently active users |
| `agentchat_voice_clones_total` | Total voice clones created |
| `agentchat_images_generated_total` | Images generated |
| `agentchat_location_alerts_total` | Proximity alerts sent |
| `agentchat_agent_response_time_seconds` | Agent response latency |

### Grafana Dashboard

Pre-built dashboard with panels for:
- Message throughput
- Active users over time
- Agent health status
- Feature usage breakdown
- Error rates

---

## 🚀 Deployment Strategy

1. **Development (pro cluster)**: Single replicas, debug logging
2. **Production (studio cluster)**: HA replicas, canary deployments
3. **iOS App**: TestFlight → App Store

---

## 📚 References

- [Existing Agent-Webinterface](../agent-webinterface/)
- [LambdaAgent CRD](../../infrastructure/knative-lambda-operator/k8s/base/crd-lambdaagent.yaml)
- [Command Center Example](../agent-pos-edge/k8s/kustomize/base/lambdaagent-command-center.yaml)
- [CloudEvents Specification](../../infrastructure/knative-lambda-operator/docs/04-architecture/CLOUDEVENTS_SPECIFICATION.md)

---

**Maintainer**: Bruno Lucena  
**Review Cycle**: Monthly
