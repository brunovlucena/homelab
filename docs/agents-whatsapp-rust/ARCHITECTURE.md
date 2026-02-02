# Agent WhatsApp Rust - System Architecture

> **Version**: 1.0.0  
> **Status**: Production-Ready  
> **Language**: Rust (Tokio, Axum)

## Executive Summary

This document describes the complete architecture for a production-ready messaging platform built in Rust. All critical architectural issues have been addressed from the start:

- ✅ Horizontal scalability (stateless services)
- ✅ Zero disconnections (connection migration)
- ✅ Exactly-once delivery (idempotency)
- ✅ Message ordering (sequence numbers)
- ✅ E2EE mandatory
- ✅ MongoDB only (single source of truth)

## System Overview

```
┌───────────────────────────────────────────────────────────────┐
│                    Client Applications                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │  iOS App     │  │ Android App  │  │  Web App     │         │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘         │
└─────────┼─────────────────┼─────────────────┼─────────────────┘
          │                 │                 │
          │    WSS (TLS)    │    WSS (TLS)    │    WSS (TLS)
          │                 │                 │
┌─────────▼─────────────────▼─────────────────▼──────────────────┐
│              Ingress Controller (Session Affinity)             │
│              • Cookie-based session affinity                   │
│              • Routes WebSocket to same instance               │
│              • Health checks for graceful shutdown             │
└─────────┬─────────────────┬─────────────────┬──────────────────┘
          │                 │                 │
    ┌─────▼─────┐    ┌──────▼──────┐   ┌──────▼──────┐
    │ Messaging │    │ Messaging   │   │ Messaging   │
    │ Service   │    │ Service     │   │ Service     │
    │ Instance 1│    │ Instance 2  │   │ Instance 3  │
    │(Rust/Tokio)│   │(Rust/Tokio) │   │(Rust/Tokio) │
    │ Stateless │    │ Stateless   │   │ Stateless   │
    └─┬───┬───┬─┘    └─┬───┬───┬───┘   └─┬────┬───┬──┘
      │   │   │        │   │   │         │    │   │
      │   │   └────────┼───┼───┼─────────┘    │   │
      │   │            │   │   │              │   │
      │   │            │   │   │              │   │
      │   │    ┌───────▼───▼───▼────────┐     │   │
      │   │    │  Redis (Connection     │     │   │
      │   │    │  Registry + Pub/Sub)   │     │   │
      │   │    │  • connection:{user_id}│     │   │
      │   │    │  • Pub/Sub: user:{id}  │     │   │
      │   │    │  • Agent cache         │     │   │
      │   │    └────────────────────────┘     │   │
      │   │                                   │   │
      │   │    ┌──────────────────────────────▼───┐
      │   └───►│  MongoDB (All Data)              │
      │        │  • Users, Conversations, Messages│
      │        │  • Idempotency keys, Sequences   │
      │        └──────────────────────────────────┘
      │
      │    ┌───────────────────────────────────────┐
      └───►│  Knative Broker (Events)              │
           │  • messaging.message.received         │
           │  • agent.response                     │
           └───────────┬───────────────────────────┘
                       │
           ┌───────────▼───────────────────────────┐
           │  Agent Gateway (Rust/Tokio)           │
           │  • Idempotency check                  │
           │  • Sequence number generation         │
           │  • Agent routing (cached)             │
           │  • Connects to: MongoDB, Redis, Broker│
           └───────────┬───────────────────────────┘
                       │
           ┌───────────▼───────────────────────────┐
           │  AI Agents                            │
           │  • agent-bruno                        │
           │  • agent-auditor                      │
           │  • agent-medical                      │
           └───────────────────────────────────────┘
```

### Connection Clarification

**Important**: The diagram shows data flow, not direct connections. Here's what actually connects to what:

- **Messaging Service** connects to:
  - ✅ **Redis**: Connection registry, Pub/Sub subscriptions, message inbox
  - ✅ **MongoDB**: Idempotency checks, sequence number generation
  - ✅ **Knative Broker**: Publishes `messaging.message.received` events
  - ✅ **Knative Trigger**: Receives `agent.response` events (via HTTP POST from Trigger)

- **Message Storage Service** connects to:
  - ✅ **Knative Broker**: Consumes `messaging.message.received` events
  - ✅ **MongoDB**: Stores messages, updates conversation metadata
  - ✅ **Redis**: Updates message inbox for offline users, Pub/Sub for delivery

- **Agent Gateway** connects to:
  - ✅ **MongoDB**: Idempotency checks, sequence number generation
  - ✅ **Redis**: Agent registry cache, intent classification cache
  - ✅ **Knative Broker**: Subscribes to `messaging.message.received`, publishes `agent.message`

- **Presence Service** connects to:
  - ✅ **Redis**: Stores presence status, publishes presence updates
  - ✅ **MongoDB**: Updates last_seen timestamps

- **Notification Service** connects to:
  - ✅ **Redis**: Subscribes to `notification:{user_id}` Pub/Sub channels
  - ✅ **APNs/FCM**: Sends push notifications

- **Media Service** connects to:
  - ✅ **MongoDB**: Stores media metadata
  - ✅ **MinIO/S3**: Generates signed URLs, stores media files

- **Agents** connect to:
  - ✅ **Knative Broker**: Publish `agent.response` events

- **MongoDB** is storage only - it does NOT connect to Knative Broker
- **Knative Broker** is for events - services publish/consume from it (acts as message queue)
- **Knative Trigger** routes `agent.response` events to Messaging Service (HTTP POST)

## Core Principles

### 1. Stateless Services

**All services are stateless** - no in-memory connection state.

- Connection state stored in Redis
- Message state stored in MongoDB
- Any instance can handle any user (after reconnection)
- Enables true horizontal scaling

### 2. Horizontal Scalability

**Multiple instances of each service can run simultaneously.**

- **Messaging Service**: Handles WebSocket connections (10K+ per instance)
- **Agent Gateway**: Routes messages (parallel routing with Tokio)
- **User Service**: Handles authentication (stateless)
- **Media Service**: Processes files (parallel processing)

### 3. Zero Disconnections

**Users never experience unexpected disconnections.**

- **Connection Migration**: Graceful handoff during deployments
- **Session Affinity**: Ingress routes to same instance
- **Automatic Reconnection**: Client reconnects automatically
- **Message Sync**: Pending messages delivered on reconnect

### 4. Exactly-Once Delivery

**Every message processed exactly once.**

- **Idempotency Keys**: Client-generated UUIDs for all messages
- **Deduplication**: MongoDB-based idempotency checks
- **Defense in Depth**: Multiple idempotency checks
- **TTL Cleanup**: Idempotency keys expire after 24 hours

### 5. Message Ordering

**Messages delivered in correct order per conversation.**

- **Sequence Numbers**: Atomic per-conversation sequence numbers
- **Gap Detection**: Client detects missing sequence numbers
- **Retransmission**: Client requests missing messages
- **Out-of-Order Buffering**: Server buffers until gaps filled

### 6. E2EE Mandatory

**All messages encrypted end-to-end.**

- **Double Ratchet Protocol**: Signal Protocol or similar
- **Client-Side Keys**: Server cannot decrypt
- **Forward Secrecy**: Automatic key rotation
- **Key Management**: KMS integration for at-rest encryption

## Service Architecture

### Messaging Service (Rust/Tokio/Axum)

**Purpose**: Real-time WebSocket server for client communication.

**Technology**:
- **Rust**: Memory-safe, high-performance
- **Tokio**: Async runtime (millions of concurrent tasks)
- **Axum**: Modern web framework
- **tokio-tungstenite**: High-performance WebSocket

**Features**:
- WebSocket connection management
- Redis Pub/Sub subscription (cross-instance routing)
- MongoDB message storage
- Idempotency checking
- Sequence number handling
- Connection migration

**Scalability**:
- 10,000+ concurrent connections per instance
- Horizontal scaling (multiple instances)
- Stateless design (all state in Redis/MongoDB)

### Agent Gateway (Rust/Tokio)

**Purpose**: Intelligent routing of messages to AI agents.

**Technology**:
- **Rust/Tokio**: Parallel routing decisions
- **Axum**: HTTP framework
- **cloudevents-sdk-rust**: CloudEvents SDK

**Routing Strategies** (priority order):
1. **Explicit Agent**: User specifies agent
2. **Conversation Context**: Extract from conversation_id
3. **Intent Classification**: AI-powered (cached)
4. **Capability Matching**: Match message to agent capabilities
5. **Default Fallback**: agent-bruno

**Performance**:
- < 50ms routing time (P95)
- Redis caching for agent registry
- Intent classification cache (80%+ hit rate)

### User Service (Rust/Axum)

**Purpose**: Authentication and user management.

**Technology**:
- **Rust/Axum**: HTTP API
- **Tokio**: Async runtime
- **mongodb**: Async database driver

**Features**:
- User registration/authentication
- JWT token generation/validation
- Profile management
- Session management

### Media Service (Rust)

**Purpose**: File upload/download and processing.

**Technology**:
- **Rust/Axum**: HTTP endpoints
- **tokio::fs**: Async file I/O
- **image-rs**: Image processing
- **ffmpeg-next**: Video processing (optional)

**Features**:
- **Signed URLs**: Generate pre-signed URLs for direct upload to MinIO/S3 (bypasses chat servers)
- **Direct Upload**: Clients upload directly to MinIO/S3 (not through chat servers)
- **CDN Integration**: Media served via CDN for fast global delivery
- Image/video processing (thumbnails, compression)
- Thumbnail generation
- Virus scanning
- MinIO/S3 storage with lifecycle policies

**Media Upload Flow**:
1. Client requests upload URL from Media Service
2. Media Service generates signed URL (expires in 1 hour)
3. Client uploads directly to MinIO/S3 using signed URL
4. Media Service processes file (thumbnails, compression)
5. Media Service stores metadata in MongoDB
6. Media served via CDN for fast delivery

### Presence Service (Rust/Tokio)

**Purpose**: Dedicated service for managing online/offline status and "last seen".

**Technology**:
- **Rust/Tokio**: Async runtime for handling heartbeats
- **Axum**: HTTP endpoints
- **Redis**: Presence state storage

**Features**:
- **Heartbeat Management**: Receives heartbeats from users (every 5 seconds)
- **Online/Offline Status**: Tracks user presence in Redis
- **Last Seen**: Records last activity timestamp
- **Presence Publishing**: Publishes presence changes to Redis Pub/Sub
- **Status Updates**: Notifies interested subscribers when users come online/offline

**Why Separate Service**:
- Decouples presence logic from messaging
- Handles high-frequency heartbeats efficiently
- Scales independently from messaging service

### Message Storage Service (Rust/Tokio)

**Purpose**: Consumes messages from queue and stores in MongoDB (decouples writing from delivery).

**Technology**:
- **Rust/Tokio**: Async runtime for queue consumption
- **mongodb**: Async database driver
- **cloudevents-sdk-rust**: CloudEvents SDK

**Features**:
- **Queue Consumer**: Subscribes to Knative Broker (consumes `messaging.message.received` events)
- **Async Processing**: Non-blocking message persistence
- **Retry Logic**: Exponential backoff for failed writes
- **Dead Letter Queue**: Failed messages after max retries
- **Message History**: Queries message history from MongoDB
- **Retention Management**: Enforces 90-day retention policy

**Why Separate Service**:
- Decouples message writing from real-time delivery
- Handles high write throughput independently
- Can scale separately from Messaging Service
- Provides buffering during traffic spikes

**Flow**:
1. Consumes `messaging.message.received` events from Knative Broker
2. Stores message in MongoDB (with sequence_number)
3. Updates conversation metadata (last_message_at, last_sequence_number)
4. If user offline: Adds message ID to Redis inbox (`inbox:{user_id}`)
5. Publishes to Redis Pub/Sub for delivery (if user online)

### Notification Service (Rust/Tokio)

**Purpose**: Push notifications for offline users.

**Technology**:
- **Rust/Tokio**: Async runtime
- **APNs**: Apple Push Notification Service (iOS)
- **FCM**: Firebase Cloud Messaging (Android)

**Features**:
- **Offline Notifications**: Sends push notifications when users are offline
- **Message Preview**: Includes message preview in notification (encrypted)
- **Badge Counts**: Updates app badge with unread count
- **Silent Notifications**: For message sync without alerting user
- **Retry Logic**: Retries failed notifications with exponential backoff

**Flow**:
1. Message arrives for offline user
2. Message Storage Service stores message in MongoDB
3. Message Storage Service publishes to Redis Pub/Sub: `notification:{user_id}`
4. Notification Service receives event
5. Notification Service sends push notification via APNs/FCM
6. User's device receives notification and syncs messages on app open

## Data Flow

### User Message Flow

```
1. Client → Messaging Service (WebSocket, encrypted, idempotency_key)
2. Messaging Service:
   - Check idempotency (MongoDB)
   - Generate sequence number (atomic, MongoDB)
   - **Immediately ACK to client** (fast response, < 100ms)
   - Create CloudEvent (messaging.message.received)
3. CloudEvent → Knative Broker (async, non-blocking)
4. **Message Storage Service** (consumes from Broker):
   - Store message in MongoDB (with sequence_number)
   - Update conversation metadata
   - If user offline: Add to Redis inbox
   - Publish to Redis Pub/Sub for delivery
5. Agent Gateway (subscribes to Broker):
   - Check idempotency (defense in depth)
   - Route to agent (cached, parallel)
   - Create CloudEvent (agent.message)
6. Agent processes and responds
7. Agent → CloudEvent (agent.response)
8. Agent Gateway:
   - Check idempotency
   - Generate sequence number
   - Publish to Redis Pub/Sub (user:{user_id})
9. Messaging Service instance (with connection):
   - Receives from Redis Pub/Sub
   - Checks ordering (sequence number)
   - Delivers via WebSocket (encrypted, in order)
10. Client decrypts and displays
```

**Key Pattern**: Immediate ACK to client, async processing via queue (WhatsApp pattern)

### Agent Response Flow (Detailed)

**Complete flow from agent to client app:**

```
┌──────────┐
│  Agent   │
│(processes│
│ message) │
└────┬─────┘
     │
     │ 1. Agent generates response
     │    Creates CloudEvent: agent.response
     │    {
     │      "type": "agent.response",
     │      "idempotency_key": "...",
     │      "conversation_id": "...",
     │      "user_id": "...",
     │      "response": "encrypted-payload",
     │      ...
     │    }
     │
     ▼
┌──────────────────┐
│ Knative Broker   │
│ (Event Bus)      │
└────┬─────────────┘
     │
     │ 2. CloudEvent published to Broker
     │
     ▼
┌──────────────────┐
│ Knative Trigger  │
│ (routes to       │
│  Messaging Svc)  │
└────┬─────────────┘
     │
     │ 3. Trigger matches agent.response events
     │    Routes to Messaging Service
     │
     ▼
┌──────────────────────────────┐
│ Messaging Service Instance   │
│ (receives CloudEvent)        │
│                              │
│ 4. Check idempotency         │
│    (MongoDB idempotency_keys)│
│                              │
│ 5. Generate sequence number   │
│    (MongoDB atomic increment) │
│                              │
│ 6. Store message in MongoDB  │
│    (with sequence_number)     │
│                              │
│ 7. Check message ordering     │
│    (verify sequence_number)   │
│                              │
│ 8. Publish to Redis Pub/Sub  │
│    Channel: user:{user_id}   │
└────┬─────────────────────────┘
     │
     │ 9. Redis Pub/Sub delivers to
     │    instance with active connection
     │
     ▼
┌──────────────────────────────┐
│ Messaging Service Instance   │
│ (with WebSocket connection) │
│                              │
│ 10. Receives from Redis      │
│     (Tokio async subscriber) │
│                              │
│ 11. Verifies ordering        │
│     (sequence_number check)  │
│                              │
│ 12. Delivers via WebSocket   │
│     (encrypted, in order)    │
└────┬─────────────────────────┘
     │
     │ 13. WebSocket message
     │     (encrypted payload)
     │
     ▼
┌──────────┐
│  Client  │
│   App    │
│          │
│ 14. Decrypts message         │
│ 15. Verifies sequence_number │
│ 16. Displays in chat         │
└──────────┘
```

**Key Points:**

1. **Agent → Broker**: Agent publishes `agent.response` CloudEvent to Knative Broker
2. **Broker → Trigger**: Knative Trigger matches `agent.response` events and routes to Messaging Service
3. **Messaging Service Processing**:
   - Receives CloudEvent via HTTP POST (from Trigger)
   - Checks idempotency (MongoDB)
   - Generates sequence number (atomic, MongoDB)
   - Stores message (MongoDB)
   - Publishes to Redis Pub/Sub: `user:{user_id}`
4. **Redis Pub/Sub Routing**:
   - All Messaging Service instances subscribe to Redis channels for their active connections
   - Instance with active WebSocket connection receives message
   - If no active connection: Message stored in MongoDB, delivered on reconnect
5. **WebSocket Delivery**:
   - Instance delivers message via WebSocket (encrypted, in order)
   - Client decrypts and displays

**Why Redis Pub/Sub?**

- **Cross-Instance Routing**: Agent response arrives at any Messaging Service instance
- **Connection Discovery**: Need to find which instance has the user's WebSocket connection
- **Efficient**: Redis Pub/Sub is O(1) publish, O(n) subscribers
- **No Duplication**: Redis Pub/Sub guarantees delivery to subscribers

## Storage Architecture

### MongoDB (Single Source of Truth)

**All data stored in MongoDB** - no split databases.

**Collections**:

**Users Collection**:
```javascript
{
  user_id: "uuid",
  phone: "+1234567890",  // unique index
  email: "user@example.com",  // unique index
  name: "John Doe",
  avatar_url: "https://...",
  created_at: ISODate(),
  last_seen: ISODate(),
  status: "online|offline"
}
```
- Indexes: `{user_id: 1}` (unique), `{phone: 1}` (unique), `{email: 1}` (unique)

**Conversations Collection**:
```javascript
{
  conversation_id: "uuid",
  user_id: "uuid",
  agent_id: "agent-bruno",
  type: "1:1|group",
  participants: ["user_id1", "user_id2", ...],  // for group chats
  last_sequence_number: 123,
  last_message_at: ISODate(),
  created_at: ISODate()
}
```
- Indexes: `{conversation_id: 1}` (unique), `{user_id: 1, agent_id: 1}`

**Messages Collection**:
```javascript
{
  message_id: "uuid",
  idempotency_key: "uuid",  // unique index
  conversation_id: "uuid",
  sequence_number: 123,  // compound index with conversation_id
  sender_id: "user_id|agent_id",
  receiver_id: "user_id|agent_id",
  type: "text|image|video|audio|document|location",
  content: "encrypted-payload",  // E2EE encrypted
  media_url: "https://...",  // for media messages
  reply_to_message_id: "uuid",  // for reply messages
  timestamp: ISODate(),
  status: "sent|delivered|read",
  created_at: ISODate()
}
```
- Indexes: 
  - `{conversation_id: 1, sequence_number: 1}` (compound, for ordering)
  - `{message_id: 1}` (unique)
  - `{idempotency_key: 1}` (unique)
  - `{conversation_id: 1, timestamp: -1}` (for message history queries)
- TTL Index: `{created_at: 1}` expires after 90 days

**Idempotency Keys Collection**:
```javascript
{
  idempotency_key: "uuid",  // unique index
  message_id: "uuid",
  created_at: ISODate()
}
```
- Indexes: `{idempotency_key: 1}` (unique)
- TTL Index: `{created_at: 1}` expires after 24 hours

**Sequence Numbers Collection**:
```javascript
{
  conversation_id: "uuid",  // unique index
  last_sequence_number: 123,
  updated_at: ISODate()
}
```
- Indexes: `{conversation_id: 1}` (unique, for atomic increments)

**Groups Collection** (for future group chat support):
```javascript
{
  group_id: "uuid",
  name: "Team Chat",
  description: "...",
  participants: ["user_id1", "user_id2", ...],  // max 100
  admin_id: "user_id",
  created_at: ISODate()
}
```
- Indexes: `{group_id: 1}` (unique)

**Replica Set**: 3 nodes for HA

### Redis (Real-Time State)

**Connection Registry**:
- Key: `connection:{user_id}`
- Value: `{instance_id, connection_id, timestamp}`
- TTL: 1 hour (refreshed on heartbeat)

**User Presence**:
- Key: `presence:{user_id}`
- Value: `{status: "online|offline", last_seen: timestamp}`
- TTL: 5 minutes (refreshed on heartbeat)

**Message Inbox (Temporary)**:
- Key: `inbox:{user_id}`
- Value: List of undelivered message IDs
- TTL: 30 days (messages also stored in MongoDB)
- **Purpose**: Fast lookup for undelivered messages when user comes online
- **Cleanup**: Messages removed from inbox after delivery

**Pub/Sub Channels**:
- Pattern: `user:{user_id}` for user-specific messages
- Pattern: `presence:{user_id}` for presence updates
- Pattern: `notification:{user_id}` for push notifications
- Instances subscribe to channels for active connections

**Caches**:
- Agent registry: `agent:{agent_id}` (TTL: 5 minutes)
- Intent classification: `intent:{message_hash}` (TTL: 1 hour)
- Session cache: `session:{user_id}` (TTL: 1 hour)

### MinIO/S3 (Media Storage)

**Buckets**:
- `messaging-media`: Images, videos
- `messaging-documents`: PDFs, files

**Lifecycle Policies**: Auto-delete after 90 days

**CDN Integration**:
- **Purpose**: Fast global media delivery
- **Pattern**: Media files cached at edge locations worldwide
- **Benefits**: Reduces latency, offloads bandwidth from origin servers
- **Implementation**: CloudFlare CDN or similar (optional, can add later)

## Connection Management

### Connection Lifecycle

1. **Connect**:
   - Client establishes WebSocket (WSS)
   - Instance authenticates (JWT, async)
   - Instance registers in Redis (async)
   - Instance subscribes to Redis Pub/Sub (Tokio task)
   - Instance delivers pending messages (MongoDB, async, ordered)

2. **Active**:
   - **Heartbeat every 5 seconds** (ping/pong, Tokio timer)
     - Client sends: `{"type": "heartbeat", "timestamp": 1699189200000}`
     - Server responds: `{"type": "heartbeat_ack", "server_time": 1699189200123}`
     - If no heartbeat for 60 seconds: Connection considered dead, removed from registry
   - Connection registry refreshed (async Redis update)
   - Presence status updated in Redis (async)
   - Messages received via Redis Pub/Sub (Tokio async subscriber)
   - Messages delivered via WebSocket (tokio-tungstenite, async)
   - Parallel processing (multiple messages concurrently)

3. **Disconnect**:
   - Remove from Redis registry (async)
   - Unsubscribe from Pub/Sub (async)
   - Clean up Tokio tasks

4. **Reconnect**:
   - Client automatically reconnects
   - New instance registers connection (async)
   - Pending messages delivered (ordered by sequence_number)

### Connection Migration (Zero Disconnections)

**When pod terminates** (scale-down, deployment):

1. Pod receives SIGTERM
2. Pod stops accepting new connections
3. Pod notifies Ingress: "draining connections"
4. Ingress stops routing to this pod
5. Pod waits for active connections (up to 30 seconds)
6. If connections still active:
   - Notify clients: `{"type": "migration", "new_endpoint": "wss://..."}`
   - Clients reconnect to new endpoint
   - New instance registers connection
   - Pending messages delivered

**Result**: Zero disconnections during deployments.

## Message Ordering (WhatsApp Pattern)

### Sequence Number Generation

- **Atomic Operation**: MongoDB `findAndModify` on `sequence_numbers` collection
- **Per-Conversation**: Sequence numbers scoped to `conversation_id`
- **Monotonically Increasing**: Always increments, never decreases
- **Rust Implementation**: `mongodb` async driver with `find_one_and_update`

### Ordering Guarantees

- Messages stored with `sequence_number`
- Messages retrieved ordered by `sequence_number`
- Index: `{conversation_id: 1, sequence_number: 1}` for fast ordered queries

### Out-of-Order Handling

- Server buffers out-of-order messages (in-memory, per conversation)
- When gap filled: Deliver buffered messages in order
- Buffer timeout: 5 seconds (then deliver what we have, request missing)

### Gap Detection & Retransmission

- Client tracks last received `sequence_number` per conversation
- Client detects gaps (missing sequence numbers)
- Client requests retransmission: `{"type": "retransmit", "conversation_id": "...", "from_sequence": 100, "to_sequence": 105}`
- Server delivers missing messages from MongoDB (ordered)

## Idempotency & Deduplication

### Idempotency Keys

- **Format**: UUID v4 (client-generated)
- **Required**: Every message MUST include `idempotency_key`
- **Storage**: MongoDB `idempotency_keys` collection
- **TTL**: 24 hours (automatic cleanup)

### Deduplication Flow

1. **Message Received**: Client sends with `idempotency_key`
2. **Check MongoDB**: Query `idempotency_keys` collection (async)
3. **If Exists**: Return existing `message_id` (deduplication, no processing)
4. **If Not Exists**:
   - Store `idempotency_key` in MongoDB (atomic)
   - Process message
   - Store message in MongoDB

### Defense in Depth

Multiple idempotency checks:
1. **Messaging Service**: Checks before storing message
2. **Agent Gateway**: Checks before routing to agent
3. **Agent** (optional): Checks before processing

## Security Architecture

### End-to-End Encryption (E2EE)

**MANDATORY** for all messages.

- **Protocol**: Double Ratchet (Signal Protocol) or similar
- **Key Management**: Client-side key generation and storage
- **Server Role**: Cannot decrypt (only routes encrypted payloads)
- **Forward Secrecy**: Automatic key rotation
- **Key Backup**: Optional encrypted backup (user-controlled)

### Encryption at Rest

- **MongoDB**: Encrypt sensitive fields (user data, metadata)
- **Media**: Encrypt files in MinIO before storage
- **KMS**: Kubernetes Secrets or external KMS for encryption keys
- **Key Rotation**: Automatic rotation every 90 days

### Encryption in Transit

- **TLS 1.3**: All connections use TLS 1.3
- **mTLS**: Mutual TLS for internal service communication (Linkerd)
- **Certificate Management**: Automatic certificate rotation

### Authentication & Authorization

- **JWT Tokens**: API authentication
- **Refresh Tokens**: Token rotation
- **WebSocket Auth**: Token in connection handshake
- **Rate Limiting**: Per-user rate limits
- **RBAC**: Role-based access control (if needed)

## Performance Targets

### Latency

- WebSocket message delivery: < 100ms
- Agent routing: < 50ms (P95)
- Message persistence: < 200ms
- End-to-end (user → agent → user): < 30s (depends on agent)

### Throughput

- 10,000+ concurrent WebSocket connections per instance
- 100,000+ messages/minute (with horizontal scaling)
- Auto-scale based on connection count and queue depth

### Reliability

- 99.9% uptime
- Zero disconnections (connection migration)
- Exactly-once delivery (idempotency)
- Message ordering (sequence numbers)
- Automatic failover

## Service Discovery

**Problem**: Services need to discover each other dynamically in Kubernetes.

**Solution**: Kubernetes native service discovery.

**Kubernetes Service Discovery**:
- Services registered automatically via Kubernetes Service objects
- DNS-based discovery: `messaging-service.homelab-services.svc.cluster.local`
- Health checks via Kubernetes liveness/readiness probes
- Auto-updates when pods scale up/down

**Why Kubernetes Native**:
- Simple (no additional infrastructure)
- Works well for single-cluster deployments
- Automatic health checks and updates
- Built-in service discovery and load balancing

## Deployment Architecture

### Kubernetes + Knative

- **Messaging Service**: Knative Service (always-on for WebSocket)
- **Agent Gateway**: Knative Lambda function
- **User Service**: Knative Service
- **Media Service**: Knative Service
- **Presence Service**: Knative Service (handles heartbeats)
- **Notification Service**: Knative Service (push notifications)
- **Message Storage Service**: Knative Service (consumes from queue, stores in MongoDB)

### Auto-Scaling

- Scale based on:
  - WebSocket connection count
  - Message queue depth
  - CPU/Memory usage
- Min replicas: 2 (for HA)
- Max replicas: 10 (configurable)

### Health Checks

- **Liveness**: Service is running
- **Readiness**: Service is ready to accept connections
- **Startup**: Service has started successfully

## Observability

### Metrics (Prometheus)

- **Business**: Active users, conversations, messages
- **Technical**: Latency, errors, throughput
- **Operational**: Connection count, queue depth, cache hit rates

### Logging (Loki)

- Structured JSON logs
- No PII in logs (user IDs hashed)
- Log levels: INFO, WARN, ERROR, DEBUG

### Tracing (Tempo)

- OpenTelemetry instrumentation
- Trace ID propagation via CloudEvents
- 10% sampling (configurable)

### Alerting

- **Critical**: High error rate, service down, agent unavailable
- **Warning**: Queue depth, storage usage, connection count

## Disaster Recovery

### Backup Strategy

- **Frequency**: Hourly backups
- **RTO**: < 15 minutes
- **RPO**: < 5 minutes
- **Testing**: Monthly automated failover tests

### High Availability

- **MongoDB**: Replica set (3 nodes)
- **Redis**: Sentinel/cluster mode
- **Services**: Multiple replicas (min 2)
- **Ingress**: Multiple ingress controllers

## WebSocket Message Schemas

### Connection Establishment

**Client → Server (Auth)**:
```json
{
  "type": "auth",
  "payload": {
    "user_id": "123",
    "auth_token": "jwt_token_here",
    "device_id": "phone-abc",
    "platform": "ios|android|web",
    "app_version": "1.0.0"
  }
}
```

**Server → Client (Auth Success)**:
```json
{
  "type": "auth_success",
  "payload": {
    "session_id": "sess_xyz789",
    "server_time": 1699189200000,
    "unread_count": 47
  }
}
```

### Sending a Message

**Client → Server**:
```json
{
  "type": "message",
  "client_message_id": "client_abc123",
  "idempotency_key": "uuid-v4",
  "payload": {
    "conversation_id": "conv-uuid",
    "receiver_id": "456",
    "content": "encrypted-payload",
    "type": "text|image|video|audio|document|location",
    "media_url": null,
    "reply_to_message_id": null,
    "timestamp": 1699189200000
  }
}
```

**Server → Client (ACK)**:
```json
{
  "type": "message_ack",
  "client_message_id": "client_abc123",
  "payload": {
    "message_id": "server-msg-id",
    "sequence_number": 123,
    "status": "sent",
    "timestamp": 1699189200123
  }
}
```

### Receiving a Message

**Server → Client**:
```json
{
  "type": "message",
  "payload": {
    "message_id": "server-msg-id",
    "sequence_number": 124,
    "sender_id": "agent-bruno",
    "receiver_id": "123",
    "conversation_id": "conv-uuid",
    "content": "encrypted-payload",
    "type": "text",
    "media_url": null,
    "timestamp": 1699189200123,
    "status": "delivered"
  }
}
```

**Client → Server (Delivery ACK)**:
```json
{
  "type": "delivery_ack",
  "message_id": "server-msg-id",
  "timestamp": 1699189200456
}
```

### Read Receipt

**Client → Server (User opened chat)**:
```json
{
  "type": "read_receipt",
  "message_ids": [
    "msg-id-1",
    "msg-id-2",
    "msg-id-3"
  ],
  "timestamp": 1699189260000
}
```

**Server → Original Sender**:
```json
{
  "type": "read_receipt",
  "payload": {
    "message_ids": ["msg-id-1", "msg-id-2"],
    "read_by": "123",
    "timestamp": 1699189260000
  }
}
```

### Heartbeat

**Client → Server (every 5 seconds)**:
```json
{
  "type": "heartbeat",
  "timestamp": 1699189200000
}
```

**Server → Client**:
```json
{
  "type": "heartbeat_ack",
  "server_time": 1699189200123
}
```

### Retransmission Request

**Client → Server (Gap detected)**:
```json
{
  "type": "retransmit",
  "conversation_id": "conv-uuid",
  "from_sequence": 100,
  "to_sequence": 105
}
```

**Server → Client (Missing messages)**:
```json
{
  "type": "messages",
  "payload": {
    "conversation_id": "conv-uuid",
    "messages": [
      {
        "message_id": "msg-100",
        "sequence_number": 100,
        "content": "encrypted-payload",
        "timestamp": 1699189200000
      },
      // ... messages 101-105
    ]
  }
}
```

### Connection Migration

**Server → Client (During deployment)**:
```json
{
  "type": "migration",
  "payload": {
    "new_endpoint": "wss://messaging.example.com/ws",
    "session_token": "session-token-here"
  }
}
```

## Message Queue Pattern

**Why Message Queue?**:
- Decouples message writing from delivery
- Handles high write throughput
- Provides buffering during spikes
- Enables async processing

**Implementation**:
- **Knative Broker** acts as message queue (RabbitMQ/NATS backend)
- Messages published to Broker asynchronously
- Message Storage Service consumes from Broker
- Stores messages in MongoDB
- Delivers to online users via Redis Pub/Sub

**Flow**:
1. Messaging Service receives message
2. Immediately ACK to client (fast response)
3. Publish to Knative Broker (async, non-blocking)
4. Message Storage Service consumes from Broker
5. Stores in MongoDB
6. Publishes to Redis Pub/Sub for delivery

**Benefits**:
- Fast client response (< 100ms)
- Reliable storage (queue persistence)
- Handles traffic spikes (queue buffering)
- Decoupled architecture

## Summary

This architecture addresses all critical issues from the original design and includes WhatsApp system design patterns:

✅ **Horizontal Scalability**: Stateless services, Redis Pub/Sub  
✅ **Zero Disconnections**: Connection migration  
✅ **Exactly-Once Delivery**: Idempotency keys  
✅ **Message Ordering**: Sequence numbers (WhatsApp pattern)  
✅ **E2EE Mandatory**: End-to-end encryption  
✅ **MongoDB Only**: Single source of truth  
✅ **Message Queue**: Knative Broker for async processing  
✅ **Presence Service**: Dedicated service for online/offline status  
✅ **Notification Service**: Push notifications for offline users  
✅ **Service Discovery**: Kubernetes native + optional registry  
✅ **Message Inbox**: Redis inbox for fast undelivered message lookup  
✅ **CDN Integration**: Fast global media delivery  
✅ **WebSocket Schemas**: Complete message format specifications  
✅ **Data Models**: Detailed MongoDB collection schemas  
✅ **Production Ready**: All critical issues resolved

**Status**: 🟢 Ready for Implementation
