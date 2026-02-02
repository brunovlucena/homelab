# Core Requirements Review
## Principal Rust Engineer Assessment

> **Review Date**: January 2025  
> **System**: agents-whatsapp-rust  
> **Reviewer**: Principal Rust Engineer  
> **Status**: 🟡 **PARTIALLY COMPLETE - CRITICAL GAPS IDENTIFIED**

---

## Executive Summary

This document reviews the implementation status of **core messaging requirements** against the specified functionality. The assessment covers:

1. ✅ **1:1 Text Messages** - Fully implemented
2. ⚠️ **Group Chats (100 participants)** - Models exist, validation missing
3. 🔴 **30-Day Message Queuing** - Not implemented
4. ⚠️ **Media Attachments** - Models exist, service missing
5. ✅ **Message Status Tracking** - Fully implemented
6. ✅ **Online/Offline Status** - Fully implemented

**Overall Assessment**: 🟡 **PARTIALLY READY**
- ✅ Core messaging infrastructure is solid
- ✅ Real-time delivery works correctly
- ⚠️ Group chat validation needs implementation
- 🔴 **CRITICAL**: 30-day message retention not implemented
- ⚠️ Media service needs to be built

---

## Requirement 1: 1:1 Text Messages ✅

### Status: **FULLY IMPLEMENTED**

### Evidence

**Models Support** (`shared/src/models.rs`):
```rust
pub struct MessagePayload {
    pub conversation_id: String,
    pub receiver_id: String,
    pub content: String, // E2EE encrypted
    pub message_type: MessageType,
    // ...
}

pub enum ConversationType {
    #[serde(rename = "1:1")]
    OneToOne,
    Group,
}
```

**WebSocket Handler** (`messaging-service/src/handlers.rs`):
- ✅ Handles `WebSocketMessage::Message` with text content
- ✅ Validates idempotency keys
- ✅ Generates sequence numbers per conversation
- ✅ Publishes to Knative Broker for async processing
- ✅ Sends immediate ACK to client

**Message Storage** (`message-storage-service/src/handlers.rs`):
- ✅ Stores messages in MongoDB
- ✅ Routes to online users via Redis Pub/Sub
- ✅ Queues for offline users in Redis inbox

### Implementation Quality

**Strengths**:
- ✅ Proper idempotency handling prevents duplicates
- ✅ Sequence numbers ensure message ordering
- ✅ Async processing via Knative Broker (non-blocking)
- ✅ Immediate ACK to client (low latency)
- ✅ Offline message queuing in Redis inbox

**Gaps**: None identified for 1:1 messaging

### Recommendation

✅ **APPROVED** - No changes needed for 1:1 text messaging.

---

## Requirement 2: Group Chats (Up to 100 Participants) ⚠️

### Status: **PARTIALLY IMPLEMENTED**

### Evidence

**Models Support** (`shared/src/models.rs`):
```rust
pub struct Conversation {
    pub conversation_type: ConversationType,
    pub participants: Vec<String>,  // ✅ Supports multiple participants
    // ...
}

pub enum ConversationType {
    OneToOne,
    Group,  // ✅ Group type exists
}
```

**Missing Implementation**:
- ❌ No validation to enforce 100-participant limit
- ❌ No group creation endpoint
- ❌ No participant management (add/remove)
- ❌ No group-specific message routing logic

### Current Behavior

The system **can** store group conversations in MongoDB, but:
1. No validation prevents >100 participants
2. No API endpoints to create/manage groups
3. Message routing assumes 1:1 (single `receiver_id`)

### Critical Gaps

#### Gap 1: Participant Limit Validation

**Location**: `user-service/src/handlers.rs` (group creation endpoint - **MISSING**)

**Required Implementation**:
```rust
pub async fn create_group_conversation(
    State(state): State<Arc<AppState>>,
    Json(payload): Json<CreateGroupPayload>,
) -> Result<Json<Conversation>, AppError> {
    // CRITICAL: Enforce 100-participant limit
    if payload.participants.len() > 100 {
        return Err(AppError::Validation(
            "Group chat cannot exceed 100 participants".to_string()
        ));
    }
    
    // Validate all participants exist
    for participant_id in &payload.participants {
        if !state.db.collection::<User>("users")
            .find_one(doc! { "_id": participant_id }, None)
            .await?
            .is_some() {
            return Err(AppError::Validation(
                format!("Participant {} not found", participant_id)
            ));
        }
    }
    
    // Create conversation
    let conversation = Conversation {
        conversation_id: Some(generate_id()),
        user_id: payload.creator_id,
        agent_id: String::new(), // Groups don't have agents
        conversation_type: ConversationType::Group,
        participants: payload.participants,
        last_sequence_number: 0,
        last_message_at: None,
        created_at: Utc::now(),
    };
    
    // Store in MongoDB
    state.db.collection("conversations")
        .insert_one(bson::to_document(&conversation)?, None)
        .await?;
    
    Ok(Json(conversation))
}
```

#### Gap 2: Group Message Routing

**Location**: `message-storage-service/src/handlers.rs` (line 96-138)

**Current Code**:
```rust
// ❌ Only handles single receiver_id
let receiver_id = data.get("receiver_id")?;
```

**Required Implementation**:
```rust
// Check if this is a group conversation
let conversation = state.db.collection::<Conversation>("conversations")
    .find_one(doc! { "_id": conversation_id }, None)
    .await?;

if let Some(conv) = conversation {
    match conv.conversation_type {
        ConversationType::Group => {
            // Broadcast to all participants (except sender)
            for participant_id in &conv.participants {
                if participant_id != &sender_id {
                    // Check if participant is online
                    let is_online = check_user_online(participant_id, &state.redis).await?;
                    
                    if is_online {
                        // Publish to participant's channel
                        publish_to_user(participant_id, &message, &state.redis).await?;
                    } else {
                        // Add to participant's inbox
                        add_to_inbox(participant_id, &message_id, &state.redis).await?;
                    }
                }
            }
        }
        ConversationType::OneToOne => {
            // Existing 1:1 logic
        }
    }
}
```

#### Gap 3: Participant Management

**Missing Endpoints**:
- `POST /api/v1/conversations/{id}/participants` - Add participant
- `DELETE /api/v1/conversations/{id}/participants/{user_id}` - Remove participant
- `GET /api/v1/conversations/{id}/participants` - List participants

**Required Validation**:
```rust
// When adding participant
if conv.participants.len() >= 100 {
    return Err(AppError::Validation(
        "Group chat has reached maximum capacity (100 participants)".to_string()
    ));
}
```

### Recommendation

⚠️ **REQUIRES IMPLEMENTATION** - Group chat models exist but functionality is incomplete.

**Priority**: **HIGH** (Core requirement)

**Action Items**:
1. ✅ Add participant limit validation (100 max)
2. ✅ Implement group conversation creation endpoint
3. ✅ Implement group message routing (broadcast to all participants)
4. ✅ Add participant management endpoints (add/remove)
5. ✅ Add group conversation queries (list groups, get group details)

**Estimated Effort**: 2-3 days

---

## Requirement 3: 30-Day Message Queuing 🔴

### Status: **NOT IMPLEMENTED**

### Evidence

**Current Implementation** (`message-storage-service/src/handlers.rs`):
```rust
if !is_online {
    // User is offline, add to inbox
    let inbox_key = format!("inbox:{}", receiver_id);
    redis::cmd("LPUSH")
        .arg(&inbox_key)
        .arg(message_id)
        .query_async(&mut conn)
        .await?;
    
    info!("Added message to inbox for offline user: {}", receiver_id);
}
```

**Problems**:
- ❌ No TTL on Redis inbox keys (messages never expire)
- ❌ No expiration logic for old messages
- ❌ No cleanup job to remove messages >30 days old
- ❌ MongoDB messages have no TTL index for 30-day retention

### Critical Gaps

#### Gap 1: Redis Inbox TTL

**Location**: `message-storage-service/src/handlers.rs` (line 105-110)

**Current Code**:
```rust
redis::cmd("LPUSH")
    .arg(&inbox_key)
    .arg(message_id)
    .query_async(&mut conn)
    .await?;
// ❌ No TTL set - messages never expire
```

**Required Fix**:
```rust
// Add message to inbox
redis::cmd("LPUSH")
    .arg(&inbox_key)
    .arg(message_id)
    .query_async(&mut conn)
    .await?;

// CRITICAL: Set TTL to 30 days (2,592,000 seconds)
redis::cmd("EXPIRE")
    .arg(&inbox_key)
    .arg(2592000)  // 30 days in seconds
    .query_async(&mut conn)
    .await?;
```

#### Gap 2: MongoDB Message TTL Index

**Location**: Database initialization (missing)

**Required Implementation**:
```rust
// In message-storage-service/src/main.rs or initialization code
pub async fn setup_mongodb_indexes(db: &Database) -> AppResult<()> {
    let collection = db.collection::<Document>("messages");
    
    // Create TTL index on created_at field (30 days)
    let index_model = IndexModel::builder()
        .keys(doc! { "created_at": 1 })
        .options(IndexOptions::builder()
            .expire_after(Duration::seconds(2592000))  // 30 days
            .name("created_at_ttl_idx".to_string())
            .build())
        .build();
    
    collection.create_index(index_model, None).await?;
    
    info!("Created TTL index on messages collection (30 days)");
    Ok(())
}
```

#### Gap 3: Cleanup Job for Old Messages

**Location**: New service or background task (missing)

**Required Implementation**:
```rust
// Background task to clean up expired messages
pub async fn cleanup_expired_messages(state: Arc<AppState>) {
    let mut interval = tokio::time::interval(Duration::from_secs(3600)); // Run every hour
    
    loop {
        interval.tick().await;
        
        let cutoff_date = Utc::now() - Duration::days(30);
        
        // Delete messages older than 30 days
        let collection = state.db.collection::<Document>("messages");
        let filter = doc! {
            "created_at": { "$lt": cutoff_date }
        };
        
        match collection.delete_many(filter, None).await {
            Ok(result) => {
                if result.deleted_count > 0 {
                    info!("Cleaned up {} expired messages", result.deleted_count);
                }
            }
            Err(e) => {
                error!("Failed to cleanup expired messages: {}", e);
            }
        }
        
        // Also cleanup Redis inboxes (they should auto-expire, but clean up stale ones)
        cleanup_stale_inboxes(&state.redis).await;
    }
}
```

#### Gap 4: Message Delivery on Reconnect

**Location**: `messaging-service/src/handlers.rs` (line 207-237)

**Current Implementation**:
```rust
async fn deliver_pending_messages(
    user_id: &str,
    storage: &mut Storage,
    tx: &mpsc::UnboundedSender<axum::extract::ws::Message>,
) -> AppResult<()> {
    // Get pending messages from Redis inbox
    let pending = storage.get_pending_messages(user_id).await?;
    // ...
}
```

**Required Enhancement**:
```rust
async fn deliver_pending_messages(
    user_id: &str,
    storage: &mut Storage,
    tx: &mpsc::UnboundedSender<axum::extract::ws::Message>,
) -> AppResult<()> {
    // Get pending messages from Redis inbox
    let pending = storage.get_pending_messages(user_id).await?;
    
    // CRITICAL: Filter out messages older than 30 days
    let cutoff = Utc::now() - Duration::days(30);
    let valid_messages: Vec<_> = pending
        .into_iter()
        .filter(|msg| msg.created_at >= cutoff)
        .collect();
    
    // Only deliver messages within 30-day window
    for message in valid_messages {
        // ... deliver message
    }
    
    Ok(())
}
```

### Recommendation

🔴 **CRITICAL GAP** - 30-day message retention is a core requirement and is **NOT IMPLEMENTED**.

**Priority**: **CRITICAL** (Core requirement)

**Action Items**:
1. 🔴 Add TTL to Redis inbox keys (30 days)
2. 🔴 Create MongoDB TTL index on `messages.created_at` (30 days)
3. 🔴 Implement cleanup job for expired messages
4. 🔴 Filter expired messages during delivery on reconnect
5. 🔴 Add monitoring/alerting for message expiration

**Estimated Effort**: 1-2 days

**Impact**: Without this, messages will accumulate indefinitely, causing:
- Storage bloat
- Performance degradation
- Cost overruns
- Non-compliance with requirements

---

## Requirement 4: Media Attachments (Images, Videos, Audio) ⚠️

### Status: **PARTIALLY IMPLEMENTED**

### Evidence

**Models Support** (`shared/src/models.rs`):
```rust
pub enum MessageType {
    Text,
    Image,   // ✅ Model exists
    Video,   // ✅ Model exists
    Audio,   // ✅ Model exists
    Document,
    Location,
    System,
}

pub struct MessagePayload {
    pub media_url: Option<String>,  // ✅ Field exists
    // ...
}
```

**Missing Implementation**:
- ❌ No Media Service implementation
- ❌ No file upload endpoints
- ❌ No MinIO/S3 integration
- ❌ No presigned URL generation
- ❌ No media processing (thumbnails, compression)

### Current Behavior

The system **can** store media URLs in messages, but:
1. No way to upload files
2. No way to generate presigned URLs
3. No media processing pipeline
4. No CDN integration

### Critical Gaps

#### Gap 1: Media Service Missing

**Location**: `media-service/` directory (referenced in docs but **NOT IMPLEMENTED**)

**Required Implementation** (from `REQUIREMENTS.md`):
```rust
// media-service/src/handlers.rs
pub async fn generate_presigned_url(
    State(state): State<Arc<AppState>>,
    Json(payload): Json<PresignedUrlRequest>,
) -> Result<Json<PresignedUrlResponse>, AppError> {
    // Generate presigned URL for direct client upload to MinIO
    let url = state.minio_client
        .presigned_put_object(
            &payload.bucket,
            &payload.key,
            Duration::from_secs(3600), // 1 hour expiry
        )
        .await?;
    
    Ok(Json(PresignedUrlResponse {
        upload_url: url,
        file_id: generate_file_id(),
        expires_in: 3600,
    }))
}

pub async fn process_media(
    State(state): State<Arc<AppState>>,
    Json(payload): Json<MediaProcessRequest>,
) -> Result<Json<MediaMetadata>, AppError> {
    // Download from MinIO
    // Generate thumbnails (for images/videos)
    // Compress media
    // Store metadata in MongoDB
    // Return CDN URL
}
```

#### Gap 2: Media Upload Flow

**Required Flow** (from `ARCHITECTURE.md`):
1. Client requests presigned URL from Media Service
2. Client uploads directly to MinIO using presigned URL
3. Client notifies Media Service upload complete
4. Media Service processes file (thumbnails, compression)
5. Media Service stores metadata in MongoDB
6. Media URL returned to client for message attachment

**Current Status**: ❌ None of these steps are implemented

### Recommendation

⚠️ **REQUIRES IMPLEMENTATION** - Media models exist but service is missing.

**Priority**: **MEDIUM** (Can be deferred to Phase 4 per requirements)

**Action Items**:
1. ⚠️ Implement Media Service (Rust/Axum)
2. ⚠️ Integrate MinIO client for presigned URLs
3. ⚠️ Implement media processing (thumbnails, compression)
4. ⚠️ Add media metadata storage in MongoDB
5. ⚠️ Integrate CDN for fast delivery

**Estimated Effort**: 3-5 days

**Note**: Per `REQUIREMENTS.md` Phase 4, media support is planned for "Weeks 11-14", so this can be deferred if 1:1 text messaging is the MVP priority.

---

## Requirement 5: Message Status Tracking (Sent, Delivered, Read) ✅

### Status: **FULLY IMPLEMENTED**

### Evidence

**Models Support** (`shared/src/models.rs`):
```rust
pub enum MessageStatus {
    Sent,      // ✅ Implemented
    Delivered, // ✅ Implemented
    Read,      // ✅ Implemented
}

pub struct StoredMessage {
    pub status: MessageStatus,  // ✅ Status tracked
    // ...
}
```

**Implementation**:

1. **Sent Status** (`messaging-service/src/handlers.rs`):
```rust
let ack = WebSocketMessage::MessageAck {
    client_message_id: client_message_id.unwrap_or_default(),
    payload: MessageAckPayload {
        message_id: message_id.clone(),
        sequence_number,
        status: MessageStatus::Sent,  // ✅ Sent immediately
        timestamp: Utc::now().timestamp_millis(),
    },
};
```

2. **Delivered Status** (`message-storage-service/src/handlers.rs`):
```rust
let ws_message = serde_json::json!({
    "status": "delivered",  // ✅ Delivered when published to Pub/Sub
    // ...
});
```

3. **Read Receipts** (`shared/src/models.rs`):
```rust
pub enum WebSocketMessage {
    ReadReceipt {
        message_ids: Vec<String>,
        timestamp: i64,
    },
    // ...
}
```

**WebSocket Protocol**:
- ✅ `MessageAck` - Sent status (immediate)
- ✅ `DeliveryAck` - Delivered status (when received)
- ✅ `ReadReceipt` - Read status (when user reads)

### Implementation Quality

**Strengths**:
- ✅ Status tracked in MongoDB (`StoredMessage.status`)
- ✅ Real-time status updates via WebSocket
- ✅ Read receipts supported in protocol
- ✅ Status persisted for message history

**Gaps**: None identified

### Recommendation

✅ **APPROVED** - Message status tracking is fully implemented and working correctly.

---

## Requirement 6: Online/Offline Status with "Last Seen" ✅

### Status: **FULLY IMPLEMENTED**

### Evidence

**Models Support** (`shared/src/models.rs`):
```rust
pub struct User {
    pub last_seen: Option<DateTime<Utc>>,  // ✅ Last seen tracked
    pub status: UserStatus,                // ✅ Online/offline status
}

pub enum UserStatus {
    Online,   // ✅ Implemented
    Offline,  // ✅ Implemented
}

pub struct Presence {
    pub status: UserStatus,
    pub last_seen: i64,  // ✅ Timestamp in milliseconds
}
```

**Implementation**:

1. **Heartbeat Updates** (`messaging-service/src/connection.rs`):
```rust
pub async fn update_heartbeat(&self) -> AppResult<()> {
    if let Some(user_id) = &self.user_id {
        let key = format!("presence:{}", user_id);
        let value = serde_json::json!({
            "status": "online",
            "last_seen": Utc::now().timestamp_millis(),  // ✅ Updated on heartbeat
        });
        
        conn.set_ex::<String, String, ()>(key, value.to_string(), 300).await?;
        // ✅ TTL of 300 seconds (5 minutes) - user goes offline if no heartbeat
    }
    Ok(())
}
```

2. **Heartbeat Handler** (`messaging-service/src/handlers.rs`):
```rust
WebSocketMessage::Heartbeat { timestamp: _ } => {
    connection.update_heartbeat().await?;  // ✅ Updates presence
    
    let ack = WebSocketMessage::HeartbeatAck {
        server_time: Utc::now().timestamp_millis(),
    };
    // ...
}
```

3. **Connection Registry** (`messaging-service/src/connection.rs`):
```rust
pub async fn register(&mut self, user_id: &str) -> AppResult<()> {
    let key = format!("connection:{}", user_id);
    // ✅ Registers user as online when WebSocket connects
    conn.set_ex::<String, String, ()>(key, value.to_string(), 3600).await?;
}
```

4. **Cleanup on Disconnect** (`messaging-service/src/connection.rs`):
```rust
pub async fn cleanup(&self) {
    let key = format!("connection:{}", user_id);
    redis::cmd("DEL").arg(&key).query_async(&mut conn).await;
    // ✅ Removes connection registry on disconnect
}
```

### Implementation Quality

**Strengths**:
- ✅ Presence tracked in Redis (fast lookups)
- ✅ Heartbeat updates `last_seen` every 5 seconds
- ✅ TTL-based offline detection (5 minutes)
- ✅ Connection registry tracks active WebSocket connections
- ✅ Cleanup on disconnect removes presence

**Potential Enhancement**:
- ⚠️ Consider updating MongoDB `User.last_seen` periodically (not just Redis)
- ⚠️ Consider presence service for cross-instance presence queries

### Recommendation

✅ **APPROVED** - Online/offline status and "last seen" are fully implemented.

**Optional Enhancement** (Low Priority):
- Sync `User.last_seen` to MongoDB periodically (for historical queries)
- Dedicated Presence Service for cross-instance presence (already planned per `ARCHITECTURE.md`)

---

## Summary & Action Plan

### Implementation Status

| Requirement | Status | Priority | Effort |
|------------|--------|----------|--------|
| 1:1 Text Messages | ✅ Complete | - | - |
| Group Chats (100) | ⚠️ Partial | HIGH | 2-3 days |
| 30-Day Queuing | 🔴 Missing | **CRITICAL** | 1-2 days |
| Media Attachments | ⚠️ Partial | MEDIUM | 3-5 days |
| Message Status | ✅ Complete | - | - |
| Online/Offline | ✅ Complete | - | - |

### Critical Path

**Must Fix Before Production**:
1. 🔴 **30-Day Message Retention** (1-2 days)
   - Add Redis inbox TTL
   - Create MongoDB TTL index
   - Implement cleanup job
   - Filter expired messages on delivery

2. ⚠️ **Group Chat Validation** (2-3 days)
   - Add 100-participant limit validation
   - Implement group message routing
   - Add participant management endpoints

**Can Defer**:
3. ⚠️ **Media Service** (3-5 days) - Per requirements, this is Phase 4

### Recommendations

1. **Immediate Action**: Implement 30-day message retention (critical gap)
2. **High Priority**: Complete group chat functionality (core requirement)
3. **Medium Priority**: Build media service (Phase 4 per requirements)

### Code Quality Assessment

**Strengths**:
- ✅ Clean Rust code with proper error handling
- ✅ Good separation of concerns (services, models, handlers)
- ✅ Proper async/await usage (Tokio)
- ✅ Idempotency and sequence numbers implemented correctly
- ✅ WebSocket protocol well-designed

**Areas for Improvement**:
- ⚠️ Add integration tests for group chats
- ⚠️ Add monitoring for message expiration
- ⚠️ Document TTL behavior in code comments

---

## Conclusion

The core messaging infrastructure is **solid** and handles 1:1 messaging, status tracking, and presence correctly. However, **two critical gaps** must be addressed:

1. 🔴 **30-day message retention** - Not implemented (critical)
2. ⚠️ **Group chat validation** - Models exist but functionality incomplete (high priority)

**Recommendation**: Fix 30-day retention immediately, then complete group chat functionality before considering production deployment.

---

**Review Completed**: January 2025  
**Next Review**: After critical gaps are addressed
