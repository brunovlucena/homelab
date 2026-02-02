# Data Model Documentation
## agents-whatsapp-rust

> **Last Updated**: January 2025  
> **Source**: `shared/src/models.rs`

---

## Overview

The `agents-whatsapp-rust` system uses a comprehensive data model designed for real-time messaging with support for 1:1 conversations, group chats (up to 100 participants), media attachments, and end-to-end encryption. The data model is implemented in Rust using `serde` for serialization/deserialization and stored primarily in MongoDB with Redis for real-time state management.

### Storage Architecture

- **MongoDB**: Primary database for all persistent data (single source of truth)
- **Redis**: Real-time state, connection registry, caching, and pub/sub
- **MinIO/S3**: Media storage (images, videos, documents)

---

## Core Entities

### 1. User Model

**Location**: `shared/src/models.rs`

```rust
pub struct User {
    #[serde(rename = "_id", skip_serializing_if = "Option::is_none")]
    pub user_id: Option<String>,        // MongoDB _id
    pub phone: Option<String>,         // Unique index
    pub email: Option<String>,          // Unique index
    pub name: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub avatar_url: Option<String>,
    pub created_at: DateTime<Utc>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub last_seen: Option<DateTime<Utc>>,
    pub status: UserStatus,             // Online/Offline
}
```

**UserStatus Enum**:
```rust
pub enum UserStatus {
    Online,
    Offline,
}
```

**MongoDB Collection**: `users`

**Indexes**:
- `{user_id: 1}` (unique)
- `{phone: 1}` (unique)
- `{email: 1}` (unique)

**Redis Keys**:
- `presence:{user_id}` - Presence status (TTL: 5 minutes)
- `connection:{user_id}` - WebSocket connection registry (TTL: 1 hour)

---

### 2. Conversation Model

**Location**: `shared/src/models.rs`

```rust
pub struct Conversation {
    #[serde(rename = "_id", skip_serializing_if = "Option::is_none")]
    pub conversation_id: Option<String>,
    pub user_id: String,
    pub agent_id: String,
    #[serde(rename = "type")]
    pub conversation_type: ConversationType,  // OneToOne or Group
    pub participants: Vec<String>,           // For group chats (max 100)
    pub last_sequence_number: u64,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub last_message_at: Option<DateTime<Utc>>,
    pub created_at: DateTime<Utc>,
}
```

**ConversationType Enum**:
```rust
pub enum ConversationType {
    #[serde(rename = "1:1")]
    OneToOne,
    Group,
}
```

**MongoDB Collection**: `conversations`

**Indexes**:
- `{conversation_id: 1}` (unique)
- `{user_id: 1, agent_id: 1}` (compound)

**Constraints**:
- Group conversations: Maximum 100 participants
- Participants array includes all user IDs in the conversation

---

### 3. Message Models

#### MessagePayload (WebSocket/API)

Used for incoming messages via WebSocket or REST API.

```rust
pub struct MessagePayload {
    pub conversation_id: String,
    pub receiver_id: String,
    pub content: String,                // E2EE encrypted
    #[serde(rename = "type")]
    pub message_type: MessageType,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub media_url: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub reply_to_message_id: Option<String>,
    pub timestamp: i64,
}
```

#### StoredMessage (MongoDB)

Persistent message storage in MongoDB.

```rust
pub struct StoredMessage {
    #[serde(rename = "_id", skip_serializing_if = "Option::is_none")]
    pub message_id: Option<String>,
    pub idempotency_key: String,       // Unique index for deduplication
    pub conversation_id: String,
    pub sequence_number: u64,          // For ordering within conversation
    pub sender_id: String,
    pub receiver_id: String,
    #[serde(rename = "type")]
    pub message_type: MessageType,
    pub content: String,                // E2EE encrypted
    #[serde(skip_serializing_if = "Option::is_none")]
    pub media_url: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub reply_to_message_id: Option<String>,
    pub timestamp: DateTime<Utc>,
    pub status: MessageStatus,         // Sent, Delivered, Read
    pub created_at: DateTime<Utc>,
}
```

**MessageType Enum**:
```rust
pub enum MessageType {
    Text,
    Image,
    Video,
    Audio,
    Document,
    Location,
    System,
}
```

**MessageStatus Enum**:
```rust
pub enum MessageStatus {
    Sent,
    Delivered,
    Read,
}
```

**MongoDB Collection**: `messages`

**Indexes**:
- `{message_id: 1}` (unique)
- `{idempotency_key: 1}` (unique)
- `{conversation_id: 1, sequence_number: 1}` (compound, for ordering)
- `{conversation_id: 1, timestamp: -1}` (for message history queries)
- `{created_at: 1}` (TTL index: 90 days expiration)

**Data Retention**: 90 days (configurable via TTL index)

---

### 4. IdempotencyKey Model

Prevents duplicate message processing.

```rust
pub struct IdempotencyKey {
    #[serde(rename = "_id")]
    pub idempotency_key: String,
    pub message_id: String,
    pub created_at: DateTime<Utc>,
}
```

**MongoDB Collection**: `idempotency_keys`

**Indexes**:
- `{idempotency_key: 1}` (unique)

**TTL Index**: `{created_at: 1}` expires after 24 hours

**Purpose**: Ensures idempotent message processing - duplicate messages with the same idempotency key are rejected.

---

### 5. SequenceNumber Model

Tracks per-conversation sequence numbers for message ordering.

```rust
pub struct SequenceNumber {
    #[serde(rename = "_id")]
    pub conversation_id: String,
    pub last_sequence_number: u64,
    pub updated_at: DateTime<Utc>,
}
```

**MongoDB Collection**: `sequence_numbers`

**Indexes**:
- `{conversation_id: 1}` (unique, for atomic increments)

**Purpose**: Ensures messages within a conversation are delivered in order. Sequence numbers are atomically incremented per conversation.

---

### 6. DeadLetterQueueEntry Model

Tracks failed messages for retry processing.

```rust
pub struct DeadLetterQueueEntry {
    #[serde(rename = "_id", skip_serializing_if = "Option::is_none")]
    pub id: Option<String>,
    pub message: StoredMessage,
    pub error: String,
    pub error_type: DLQErrorType,
    pub retry_count: u32,
    pub max_retries: u32,
    pub next_retry_at: Option<DateTime<Utc>>,
    pub created_at: DateTime<Utc>,
    pub last_retry_at: Option<DateTime<Utc>>,
    pub status: DLQStatus,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub resolved_at: Option<DateTime<Utc>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub resolved_reason: Option<String>,
}
```

**DLQErrorType Enum**:
```rust
pub enum DLQErrorType {
    BrokerPublishFailed,
    StorageFailed,
    ValidationFailed,
    Timeout,
    ServiceUnavailable,
    NetworkError,
    Unknown,
}
```

**DLQStatus Enum**:
```rust
pub enum DLQStatus {
    Pending,
    Retrying,
    Failed,
    Resolved,
}
```

**MongoDB Collection**: `dead_letter_queue`

**Indexes**:
- `{status: 1, next_retry_at: 1}` (for retry scheduling)
- `{created_at: -1}` (for querying recent failures)
- `{message.message_id: 1}` (for message lookup)
- `{expires_at: 1}` (TTL index for auto-cleanup)

---

## WebSocket Protocol Models

The system uses a tagged enum for WebSocket message types with `serde(tag = "type")`.

### WebSocketMessage Enum

```rust
pub enum WebSocketMessage {
    // Authentication
    Auth {
        payload: AuthPayload,
    },
    AuthSuccess {
        payload: AuthSuccessPayload,
    },
    AuthError {
        error: String,
    },

    // Messages
    Message {
        #[serde(skip_serializing_if = "Option::is_none")]
        client_message_id: Option<String>,
        idempotency_key: String,
        payload: MessagePayload,
    },
    MessageAck {
        client_message_id: String,
        payload: MessageAckPayload,
    },

    // Delivery receipts
    DeliveryAck {
        message_id: String,
        timestamp: i64,
    },
    ReadReceipt {
        message_ids: Vec<String>,
        timestamp: i64,
    },

    // Heartbeat
    Heartbeat {
        timestamp: i64,
    },
    HeartbeatAck {
        server_time: i64,
    },

    // Retransmission
    Retransmit {
        conversation_id: String,
        from_sequence: u64,
        to_sequence: u64,
    },
    Messages {
        payload: MessagesPayload,
    },

    // Connection migration
    Migration {
        payload: MigrationPayload,
    },
}
```

### Supporting WebSocket Payloads

**AuthPayload**:
```rust
pub struct AuthPayload {
    pub user_id: String,
    pub auth_token: String,
    pub device_id: String,
    pub platform: String,
    pub app_version: String,
}
```

**AuthSuccessPayload**:
```rust
pub struct AuthSuccessPayload {
    pub session_id: String,
    pub server_time: i64,
    pub unread_count: u64,
}
```

**MessageAckPayload**:
```rust
pub struct MessageAckPayload {
    pub message_id: String,
    pub sequence_number: u64,
    pub status: MessageStatus,
    pub timestamp: i64,
}
```

**MessagesPayload**:
```rust
pub struct MessagesPayload {
    pub conversation_id: String,
    pub messages: Vec<StoredMessage>,
}
```

**MigrationPayload**:
```rust
pub struct MigrationPayload {
    pub new_endpoint: String,
    pub session_token: String,
}
```

---

## Event Models

### MessageReceivedEvent (CloudEvent)

Used for async message processing via Knative Broker.

```rust
pub struct MessageReceivedEvent {
    pub message_id: String,
    pub idempotency_key: String,
    pub conversation_id: String,
    pub sender_id: String,
    pub receiver_id: String,
    pub sequence_number: u64,
    pub message_type: MessageType,
    pub content: String,
    pub timestamp: DateTime<Utc>,
}
```

### AgentResponseEvent (CloudEvent)

Used for agent responses to user messages.

```rust
pub struct AgentResponseEvent {
    pub idempotency_key: String,
    pub conversation_id: String,
    pub user_id: String,
    pub agent_id: String,
    pub response: String,  // E2EE encrypted
    pub timestamp: DateTime<Utc>,
}
```

---

## Redis Data Structures

### Connection Registry

**Key Pattern**: `connection:{user_id}`

**Value**:
```rust
pub struct ConnectionRegistry {
    pub instance_id: String,
    pub connection_id: String,
    pub timestamp: i64,
}
```

**TTL**: 1 hour (refreshed on heartbeat)

**Purpose**: Maps user_id to the messaging service instance handling their WebSocket connection.

---

### Presence Status

**Key Pattern**: `presence:{user_id}`

**Value**:
```rust
pub struct Presence {
    pub status: UserStatus,
    pub last_seen: i64,  // Timestamp in milliseconds
}
```

**TTL**: 5 minutes (refreshed on heartbeat)

**Purpose**: Fast lookup of user online/offline status and last seen timestamp.

---

### Message Inbox (Temporary)

**Key Pattern**: `inbox:{user_id}`

**Value**: Redis List of undelivered message IDs

**TTL**: 30 days (should be implemented, currently missing per CORE_REQUIREMENTS_REVIEW.md)

**Purpose**: Fast lookup for undelivered messages when user comes online. Messages are removed from inbox after delivery.

**Note**: ⚠️ **Gap Identified**: TTL on inbox keys is not currently implemented (see CORE_REQUIREMENTS_REVIEW.md).

---

### Pub/Sub Channels

**Channel Patterns**:
- `user:{user_id}` - User-specific messages
- `presence:{user_id}` - Presence updates
- `notification:{user_id}` - Push notifications

**Purpose**: Cross-instance message routing. Instances subscribe to channels for active connections.

---

### Caches

**Key Patterns**:
- `agent:{agent_id}` - Agent registry cache (TTL: 5 minutes)
- `intent:{message_hash}` - Intent classification cache (TTL: 1 hour)
- `session:{user_id}` - Session cache (TTL: 1 hour)

---

## MongoDB Collections Summary

| Collection | Purpose | Key Indexes | TTL Index |
|------------|---------|-------------|-----------|
| `users` | User profiles | user_id (unique), phone (unique), email (unique) | - |
| `conversations` | Conversation metadata | conversation_id (unique), user_id+agent_id | - |
| `messages` | Message history | message_id (unique), idempotency_key (unique), conversation_id+sequence_number (compound) | created_at (90 days) |
| `idempotency_keys` | Deduplication | idempotency_key (unique) | created_at (24 hours) |
| `sequence_numbers` | Per-conversation sequence tracking | conversation_id (unique) | - |
| `dead_letter_queue` | Failed message retry queue | status+next_retry_at, message.message_id | expires_at |

---

## Design Characteristics

### 1. End-to-End Encryption (E2EE)
- Message `content` field is encrypted end-to-end
- Only clients can decrypt message content
- Server stores encrypted content only

### 2. Idempotency
- Every message includes an `idempotency_key`
- Duplicate messages with the same key are rejected
- Prevents duplicate processing on retries

### 3. Message Ordering
- `sequence_number` ensures ordering within conversations
- Sequence numbers are atomically incremented per conversation
- Messages delivered in sequence order

### 4. Status Tracking
- Three-state status: `Sent` → `Delivered` → `Read`
- Real-time status updates via WebSocket
- Status persisted in MongoDB for history

### 5. Group Chat Support
- Models support group conversations (up to 100 participants)
- `participants` array tracks all group members
- ⚠️ **Gap Identified**: 100-participant limit validation not enforced (see CORE_REQUIREMENTS_REVIEW.md)

### 6. Media Support
- Models include `media_url` and `MessageType` enum for media
- Supports: Image, Video, Audio, Document, Location
- ⚠️ **Gap Identified**: Media service not implemented (models exist but no upload/processing)

### 7. Data Retention
- Messages: 90 days (TTL index on `created_at`)
- Idempotency keys: 24 hours (TTL index)
- Redis inbox: 30 days (⚠️ **Gap**: TTL not implemented)

### 8. Offline Message Queuing
- Messages for offline users stored in Redis inbox
- Delivered when user comes online
- Messages also stored in MongoDB for persistence

---

## Known Gaps & Issues

Based on `CORE_REQUIREMENTS_REVIEW.md`:

### 🔴 Critical Gaps

1. **30-Day Message Retention**
   - Redis inbox TTL not implemented
   - Messages accumulate indefinitely in Redis
   - **Impact**: Storage bloat, performance degradation

### ⚠️ High Priority Gaps

2. **Group Chat Validation**
   - 100-participant limit not enforced
   - Group message routing incomplete
   - Participant management endpoints missing

3. **Media Service**
   - Models exist but service not implemented
   - No file upload endpoints
   - No MinIO/S3 integration
   - **Note**: Per requirements, this is Phase 4 (can be deferred)

---

## Serialization Details

All models use `serde` for serialization/deserialization:

- **Enum naming**: `snake_case` for WebSocket messages, `lowercase` for status/type enums
- **Field renaming**: `_id` for MongoDB document IDs
- **Optional fields**: `skip_serializing_if = "Option::is_none"` to omit null fields
- **Tagged enums**: `#[serde(tag = "type")]` for WebSocket message discrimination

---

## References

- **Source Code**: `flux/ai/agents-whatsapp-rust/shared/src/models.rs`
- **Architecture**: `docs/agents-whatsapp-rust/ARCHITECTURE.md`
- **Requirements Review**: `docs/agents-whatsapp-rust/CORE_REQUIREMENTS_REVIEW.md`
- **Requirements**: `docs/agents-whatsapp-rust/REQUIREMENTS.md`

---

**Document Version**: 1.0  
**Last Updated**: January 2025
