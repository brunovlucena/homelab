# Agent WhatsApp Rust - Requirements

> **Version**: 1.0.0  
> **Status**: Production-Ready  
> **Language**: Rust

## Overview

Complete requirements for a production-ready messaging platform built in Rust. All critical architectural issues have been addressed from the start.

## Business Requirements

### BR-1: User Registration & Authentication

- Phone number or email-based registration
- SMS/Email verification
- Secure authentication (JWT tokens)
- Profile creation and management
- Account recovery options

### BR-2: Multi-Agent Chat Interface

- Separate chat threads per agent
- Agent list/discovery interface
- Switch between agent conversations
- Agent status indicators (online/offline)
- Agent descriptions and capabilities
- **Group Chats** (Future): Support group conversations with multiple participants (up to 100)

### BR-3: Real-Time Messaging

- Message delivery < 1 second
- **Message Ordering**: Messages delivered in correct order (per conversation)
- **Gap Detection**: Client detects and requests missing messages
- Typing indicators
- Read receipts (delivered/read)
- Message status indicators
- Offline message queuing (with ordering preserved)

### BR-4: Message Types Support

- Text messages
- Images (with preview)
- Videos
- Audio messages
- Documents/PDFs
- Location sharing
- Voice messages
- Emoji and reactions
- Reply/forward messages

### BR-5: Agent Discovery

- Browse available agents
- Search agents by name/capability
- View agent descriptions
- See agent examples/demos
- Agent categories/tags

### BR-6: Conversation Management

- Conversation list with previews
- Unread message counts
- Conversation search
- Archive conversations
- Delete conversations
- Pin important conversations

### BR-7: Response Time

- Simple queries: < 5 seconds
- Complex queries: < 30 seconds
- Long-running tasks: Progress updates every 30 seconds
- Typing indicators during processing
- Timeout handling with user notification

## Technical Requirements

### TR-1: Application Architecture

#### TR-1.1: Client Applications

- **Mobile Apps**:
  - iOS (Swift/SwiftUI) - Native for best performance
  - Android (Kotlin/Jetpack Compose) - Native for best performance
  - Cross-platform option: React Native or Flutter (faster development)
- **Web Application**:
  - **Tauri** (Rust backend + web frontend) - Excellent performance, small bundle size
  - Or: React/Next.js or Vue.js
  - Progressive Web App (PWA) support
  - Responsive design (mobile-first)

#### TR-1.2: Backend Services (All in Rust)

- **Messaging Service**: Real-time message handling (Rust/Tokio/Axum)
- **Agent Gateway**: Route messages to agents (Rust/Tokio)
- **User Service**: Authentication and user management (Rust/Axum)
- **Media Service**: File upload/download (Rust, parallel processing)
- **Presence Service**: Online/offline status management (Rust/Tokio)
- **Notification Service**: Push notifications (Rust/Tokio, APNs/FCM)
- **Message Storage Service**: Consumes from queue, stores in MongoDB (Rust/Tokio)

### TR-2: Real-Time Communication

#### TR-2.1: WebSocket Protocol

- **WebSocket Server**: WebSocket over TLS (WSS only)
- **Horizontal Scalability**: Multiple Messaging Service instances can run simultaneously
- **Connection Management**:
  - **Connection Registry**: Redis-based registry mapping `user_id` → `service_instance_id`
  - **Session Affinity**: Ingress controller uses session affinity (cookie-based) to route WebSocket connections
  - **Connection Migration**: Graceful connection handoff when instances scale down
  - **Heartbeat**: Ping/pong every 5 seconds to detect dead connections
  - Client sends heartbeat every 5 seconds
  - Server responds with heartbeat_ack
  - If no heartbeat for 60 seconds: Connection considered dead
  - Presence Service tracks heartbeats and updates online/offline status
  - **Reconnection**: Client automatically reconnects with exponential backoff
- **Message Routing**:
  - **Redis Pub/Sub**: Cross-instance message routing via Redis channels
  - **Channel Pattern**: `user:{user_id}` for user-specific messages
  - **Instance Discovery**: Each instance subscribes to relevant user channels
- **State Management**:
  - **Stateless Services**: Messaging Service instances are stateless
  - **Connection State in Redis**: Active connections tracked in Redis (not in-memory)
  - **Message Queuing**: Offline messages stored in MongoDB, delivered on reconnection
- **Multi-Device Support**: Same user can have multiple active connections (different devices)
- **Zero-Downtime Scaling**: Connections maintained during deployments via graceful shutdown

#### TR-2.2: Message Protocol

- **Message Format**: JSON (encrypted payload for E2EE)
- **Message Types**: text, image, video, audio, document, location, system
- **Message Structure**:
  ```json
  {
    "id": "msg-uuid",
    "idempotency_key": "idemp-key-uuid",  // REQUIRED
    "conversation_id": "conv-uuid",
    "from": "user-id|agent-id",
    "to": "user-id|agent-id",
    "type": "text",
    "content": "encrypted-payload",  // E2EE encrypted
    "timestamp": "2025-01-27T10:00:00Z",
    "status": "sent|delivered|read",
    "sequence_number": 123,  // REQUIRED: For ordering
    "metadata": {}
  }
  ```
- **Message Ordering Requirements** (WhatsApp Pattern):
  - **Sequence Numbers**: Every message MUST have sequence_number per conversation
  - **Ordering Guarantee**: Messages delivered in sequence_number order
  - **Gap Detection**: Client detects missing sequence numbers and requests retransmission
  - **Out-of-Order Handling**: Server buffers out-of-order messages until gap filled
  - **Per-Conversation Ordering**: Sequence numbers scoped to conversation_id
- **Idempotency Requirements**:
  - **Idempotency Key**: Every message MUST include unique idempotency key
  - **Deduplication**: Check idempotency key before processing
  - **Storage**: Idempotency keys stored in MongoDB with TTL (24 hours)
  - **Exactly-Once Delivery**: Guarantee message processed exactly once
  - **Retry Handling**: Retries use same idempotency key

### TR-3: Agent Integration Architecture

#### TR-3.1: CloudEvents Protocol

- Incoming user messages → CloudEvent (`messaging.message.received`)
- Agent responses → CloudEvent (`agent.response`)
- Event routing via Knative Broker
- Event schema compliance (v1.0)
- **Integration Pattern**: The Messaging Service acts as a bridge:
  - **WebSocket** (real-time) for client ↔ messaging service
  - **CloudEvents** (async) for messaging service ↔ agents
  - **Redis Pub/Sub** (real-time) for cross-instance routing

#### TR-3.2: Agent Routing

- Explicit agent selection (user selects agent)
- Intent-based routing (optional, for auto-suggestions)
- Multi-agent support (user can chat with multiple agents)
- Agent availability checking

#### TR-3.3: Service Discovery

- **Kubernetes Native**: DNS-based service discovery
  - Services registered automatically via Kubernetes Service objects
  - DNS: `messaging-service.homelab-services.svc.cluster.local`
  - Health checks via Kubernetes liveness/readiness probes
  - Auto-updates when pods scale up/down
  - Built-in service discovery and load balancing
- **Agent Discovery**:
  - Query Kubernetes API for agent services (async, cached)
  - Cache agent metadata in Redis (name, description, capabilities, avatar)
  - Background refresh every 5 minutes (not per-message)
  - Health check integration
  - Agent availability status
- **Rust Implementation**: 
  - `k8s-openapi` for Kubernetes API
  - `tokio` for async operations
  - Redis caching for fast lookups

### TR-4: Infrastructure Requirements

#### TR-4.1: Deployment Platform

- Messaging Service: Knative Service (always-on for WebSocket)
- Agent Gateway: Knative Lambda function
- User Service: Knative Service
- Media Service: Knative Service
- Auto-scaling based on load
- Resource limits per service

#### TR-4.2: Storage

- **MongoDB**: Primary database for ALL data (users, conversations, messages, metadata)
  - **Replica Set**: Minimum 3 nodes for high availability
  - **Sharding**: For horizontal scalability (if needed)
  - **Collections**:
    - `users`: User accounts, profiles, authentication data
    - `conversations`: Conversation metadata, participants, settings
    - `messages`: Message history (time-series optimized with TTL indexes)
    - `media_metadata`: Media file references and metadata
    - `sessions`: User session data (with TTL indexes)
    - `idempotency_keys`: Message deduplication tracking
    - `sequence_numbers`: Per-conversation sequence number tracking
  - **Indexes**: Optimized for query patterns (user_id, conversation_id, timestamp, sequence_number)
  - **Data retention**: 90 days (configurable via TTL indexes)
- **Redis**: Real-time state, connection registry, caching, pub/sub
  - **Connection Registry**: Maps user_id → messaging_service_instance
  - **Pub/Sub Channels**: For cross-instance message routing
  - **Session Cache**: Fast session lookups
  - **Agent Registry Cache**: Cached agent metadata (TTL: 5 minutes)
  - **Intent Classification Cache**: Cached intent results (TTL: 1 hour)
- **MinIO/S3**: Media storage (images, videos, documents)
  - **Buckets**: `messaging-media`, `messaging-documents`
  - **Lifecycle Policies**: Auto-delete after 90 days
  - **Signed URLs**: Pre-signed URLs for direct client upload (bypasses servers)
  - **CDN Integration**: Optional CDN for fast global media delivery
  - **Direct Upload**: Clients upload directly to MinIO/S3 (not through chat servers)

#### TR-4.3: Message Queue & Ordering

- **Knative Broker**: Message queue (RabbitMQ/NATS backend)
  - **Purpose**: Decouples message writing from delivery
  - **Pattern**: Immediate ACK to client, async publish to queue
  - **Benefits**: Fast response, handles traffic spikes, reliable storage
- **Message Storage Service**: Consumes from queue, writes to MongoDB
  - **Async Processing**: Non-blocking message persistence
  - **Retry Logic**: Exponential backoff for failed writes
  - **Dead Letter Queue**: Failed messages after max retries
- **Message Ordering**: Per-conversation message ordering (sequence numbers)
- **Out-of-Order Buffering**: Buffer messages until sequence gaps filled
- **Message Persistence**: All messages persisted in MongoDB
- **Priority Queues**: For urgent messages (typing indicators, read receipts)
- **Offline Queue**: 
  - Messages stored in MongoDB when user offline
  - Redis inbox (`inbox:{user_id}`) for fast lookup
  - Messages delivered on reconnect (ordered by sequence_number)
- **Sequence Number Generation**: Atomic sequence number generation per conversation

### TR-5: Security Requirements

#### TR-5.1: Authentication & Authorization

- JWT tokens for API authentication
- Refresh token rotation
- WebSocket authentication (token in connection)
- OAuth2 support (optional, for SSO)
- Rate limiting per user

#### TR-5.2: Data Privacy

- **End-to-End Encryption (E2EE)**: **MANDATORY** for all messages
  - **Key Exchange**: Double Ratchet protocol (Signal Protocol) or similar
  - **Key Management**: Client-side key generation and storage
  - **Server Role**: Server cannot decrypt messages (only routes encrypted payloads)
  - **Key Rotation**: Automatic key rotation per conversation
  - **Forward Secrecy**: New keys for each message exchange
  - **Key Backup**: Optional encrypted key backup (user-controlled)
- **Encryption at Rest**:
  - **MongoDB Encryption**: Encrypt sensitive fields (user data, conversation metadata)
  - **Media Encryption**: Encrypt media files in MinIO before storage
  - **Key Management Service (KMS)**: Use Kubernetes Secrets or external KMS for encryption keys
  - **Key Rotation**: Automatic key rotation every 90 days
- **Encryption in Transit**:
  - **TLS 1.3**: All connections use TLS 1.3
  - **mTLS**: Mutual TLS for internal service communication (via Linkerd)
  - **Certificate Management**: Automatic certificate rotation
- **PII Protection**:
  - **Logging**: No PII in logs (user IDs hashed with SHA-256)
  - **Error Messages**: No user data in error messages
  - **Audit Logging**: All data access logged (encrypted audit trail)
- **GDPR Compliance**:
  - **Data Deletion**: Complete data deletion on user request (including backups)
  - **Right to Access**: Export all user data in machine-readable format
  - **Privacy Policy**: Clear privacy policy and consent management
  - **Data Portability**: Export conversations in standard format

#### TR-5.3: Secrets Management

- Use Kubernetes Secrets (encrypted at rest)
- Sealed Secrets for GitOps
- Token rotation support
- No secrets in code or logs

### TR-6: Idempotency & Deduplication Requirements

#### TR-6.1: Message Idempotency

- **Idempotency Keys**: Every message MUST include unique idempotency_key
- **Client-Generated**: Idempotency keys generated by client (UUID v4)
- **Storage**: Idempotency keys stored in MongoDB `idempotency_keys` collection
- **TTL**: Idempotency keys expire after 24 hours (automatic cleanup)
- **Index**: Unique index on `idempotency_key` for fast lookups
- **Defense in Depth**: Multiple idempotency checks:
  1. Messaging Service checks before storing message
  2. Agent Gateway checks before routing to agent
  3. Agent checks before processing (optional)

#### TR-6.2: Deduplication Strategy

- **Check Before Processing**: Query MongoDB before any processing
- **Atomic Operations**: Use MongoDB transactions for idempotency key storage
- **Retry Handling**: Retries use same idempotency key (no duplicates)
- **CloudEvents Deduplication**: CloudEvents include idempotency_key in event data
- **Response Deduplication**: Agent responses also include idempotency keys

#### TR-6.3: Message Ordering Implementation

- **Sequence Number Generation**:
  - Atomic operation: MongoDB `findAndModify` on `sequence_numbers` collection
  - Per-conversation sequence counter: `{conversation_id, last_sequence_number}`
  - Increment and return new sequence number (atomic, no race conditions)
  - **Rust Implementation**: `mongodb` async driver with find_one_and_update
  
- **Ordering Guarantees**:
  - Messages stored with sequence_number
  - Messages retrieved ordered by sequence_number
  - Index: `{conversation_id: 1, sequence_number: 1}` for fast ordered queries
  
- **Out-of-Order Handling**:
  - Server buffers out-of-order messages (in-memory, per conversation)
  - When gap filled: Deliver buffered messages in order
  - Buffer timeout: 5 seconds (then deliver what we have, request missing)
  
- **Gap Detection**:
  - Client tracks last received sequence_number per conversation
  - Client detects gaps (missing sequence numbers)
  - Client requests retransmission: `{"type": "retransmit", "conversation_id": "...", "from_sequence": 100, "to_sequence": 105}`
  - Server delivers missing messages from MongoDB
  
- **Retransmission**:
  - Client sends retransmission request
  - Server queries MongoDB: `messages.find({conversation_id, sequence_number: {$gte: from, $lte: to}})`
  - Server delivers messages in order
  - Client fills gaps and displays messages

### TR-7: Performance Requirements

#### TR-7.1: Latency

- WebSocket message delivery: < 100ms
- Agent routing: < 50ms
- Message persistence: < 200ms
- End-to-end (user → agent → user): < 30s (depends on agent)

#### TR-7.2: Throughput

- 10,000+ concurrent WebSocket connections per instance (Rust/Tokio advantage)
- 100,000+ messages/minute (with horizontal scaling)
- Auto-scale based on connection count and queue depth
- **Rust Benefits**: 
  - Tokio async runtime handles millions of concurrent tasks
  - Lower memory footprint = more connections per instance
  - Zero-cost abstractions = higher throughput

#### TR-7.3: Reliability

- 99.9% uptime
- **Zero Disconnections**: Users never experience unexpected disconnections
- **Message Delivery Guarantee**: Exactly-once delivery (with idempotency)
- **Automatic Failover**: Seamless failover with connection migration
- **Offline Message Sync**: All messages delivered when user reconnects
- **Graceful Degradation**: System continues operating during partial failures

## Implementation Phases

### Phase 1: MVP Backend (Weeks 1-4)

**Goal**: Basic messaging infrastructure in Rust

**Deliverables**:
- [ ] User service (Rust/Axum, registration, authentication)
- [ ] Messaging service (Rust/Tokio/Axum, WebSocket server)
- [ ] Basic message routing to agent-bruno
- [ ] Message storage (MongoDB only, with idempotency support)
- [ ] Sequence number generation (per-conversation ordering)
- [ ] Basic observability (logs, metrics)

**Success Criteria**:
- Users can register and authenticate
- WebSocket connections work (Rust/Tokio)
- Messages can be sent and received
- Agent responses delivered
- Messages delivered in order (sequence numbers)

### Phase 2: Mobile/Web Clients (Weeks 5-9)

**Goal**: Client applications

**Deliverables**:
- [ ] Web app (Tauri with Rust backend, or React/Next.js)
- [ ] iOS app (Swift/SwiftUI) - or React Native
- [ ] Android app (Kotlin/Jetpack Compose) - or React Native
- [ ] Basic chat UI
- [ ] Authentication flow
- [ ] Message list and chat view
- [ ] Sequence number handling (gap detection, retransmission)

**Success Criteria**:
- Users can use web or mobile app
- Chat interface works
- Real-time messaging functional
- Messages displayed in correct order

### Phase 3: Multi-Agent Support (Weeks 8-10)

**Goal**: Support multiple agents

**Deliverables**:
- [ ] Agent discovery service
- [ ] Agent selection UI
- [ ] Multi-agent conversation management
- [ ] Agent list/discovery interface
- [ ] Agent status indicators

**Success Criteria**:
- Users can select and chat with different agents
- Multiple conversations work simultaneously
- Agent discovery functional

### Phase 4: Rich Features (Weeks 11-14)

**Goal**: Enhanced messaging features

**Deliverables**:
- [ ] Media support (images, videos, documents)
- [ ] Voice messages
- [ ] Location sharing
- [ ] Read receipts
- [ ] Typing indicators
- [ ] Message reactions
- [ ] Push notifications

**Success Criteria**:
- All message types supported
- Rich media works
- Real-time features functional

### Phase 5: Production Ready (Weeks 15-18)

**Goal**: Production hardening

**Deliverables**:
- [ ] Security hardening (encryption, PII protection)
- [ ] Performance optimization
- [ ] Comprehensive testing
- [ ] Documentation
- [ ] Deployment automation

**Success Criteria**:
- Security audit passed
- Performance targets met
- Full documentation
- Production deployment

## Key Architectural Decisions

1. ✅ **Rust Language**: All backend services in Rust (Tokio for parallelism)
2. ✅ **MongoDB Only**: Single database for all data (no PostgreSQL)
3. ✅ **Horizontal Scalability**: Stateless messaging service with Redis Pub/Sub
4. ✅ **Zero Disconnections**: Connection migration and graceful shutdown
5. ✅ **Idempotency**: Exactly-once delivery with idempotency keys
6. ✅ **Message Ordering**: Per-conversation sequence numbers (WhatsApp pattern)
7. ✅ **E2EE Mandatory**: End-to-end encryption required for all messages
8. ✅ **Gap Detection**: Client detects missing messages and requests retransmission
9. ✅ **Message Queue**: Knative Broker decouples writing from delivery
10. ✅ **Presence Service**: Dedicated service for online/offline status
11. ✅ **Notification Service**: Push notifications for offline users (APNs/FCM)
12. ✅ **Service Discovery**: Kubernetes native DNS-based discovery
13. ✅ **Message Inbox**: Redis inbox for fast undelivered message lookup
14. ✅ **CDN Integration**: Fast global media delivery
15. ✅ **Signed URLs**: Direct client upload to MinIO/S3 (bypasses servers)

## Status

🟢 **Ready for Implementation** - All critical architectural issues resolved
