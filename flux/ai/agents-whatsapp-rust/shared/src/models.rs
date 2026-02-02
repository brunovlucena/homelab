use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};

/// Message types supported by the platform
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "lowercase")]
pub enum MessageType {
    Text,
    Image,
    Video,
    Audio,
    Document,
    Location,
    System,
}

/// Message status
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "lowercase")]
pub enum MessageStatus {
    Sent,
    Delivered,
    Read,
}

/// WebSocket message types
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
#[serde(rename_all = "snake_case")]
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

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AuthPayload {
    pub user_id: String,
    pub auth_token: String,
    pub device_id: String,
    pub platform: String,
    pub app_version: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AuthSuccessPayload {
    pub session_id: String,
    pub server_time: i64,
    pub unread_count: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MessagePayload {
    pub conversation_id: String,
    pub receiver_id: String,
    pub content: String, // E2EE encrypted
    #[serde(rename = "type")]
    pub message_type: MessageType,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub media_url: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub reply_to_message_id: Option<String>,
    pub timestamp: i64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MessageAckPayload {
    pub message_id: String,
    pub sequence_number: u64,
    pub status: MessageStatus,
    pub timestamp: i64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MessagesPayload {
    pub conversation_id: String,
    pub messages: Vec<StoredMessage>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MigrationPayload {
    pub new_endpoint: String,
    pub session_token: String,
}

/// Stored message in MongoDB
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StoredMessage {
    #[serde(rename = "_id", skip_serializing_if = "Option::is_none")]
    pub message_id: Option<String>,
    pub idempotency_key: String,
    pub conversation_id: String,
    pub sequence_number: u64,
    pub sender_id: String,
    pub receiver_id: String,
    #[serde(rename = "type")]
    pub message_type: MessageType,
    pub content: String, // E2EE encrypted
    #[serde(skip_serializing_if = "Option::is_none")]
    pub media_url: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub reply_to_message_id: Option<String>,
    pub timestamp: DateTime<Utc>,
    pub status: MessageStatus,
    pub created_at: DateTime<Utc>,
}

/// User model
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct User {
    #[serde(rename = "_id", skip_serializing_if = "Option::is_none")]
    pub user_id: Option<String>,
    pub phone: Option<String>,
    pub email: Option<String>,
    pub name: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub avatar_url: Option<String>,
    pub created_at: DateTime<Utc>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub last_seen: Option<DateTime<Utc>>,
    pub status: UserStatus,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "lowercase")]
pub enum UserStatus {
    Online,
    Offline,
}

/// Conversation model
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Conversation {
    #[serde(rename = "_id", skip_serializing_if = "Option::is_none")]
    pub conversation_id: Option<String>,
    pub user_id: String,
    pub agent_id: String,
    #[serde(rename = "type")]
    pub conversation_type: ConversationType,
    pub participants: Vec<String>,
    pub last_sequence_number: u64,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub last_message_at: Option<DateTime<Utc>>,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "lowercase")]
pub enum ConversationType {
    #[serde(rename = "1:1")]
    OneToOne,
    Group,
}

/// Idempotency key record
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct IdempotencyKey {
    #[serde(rename = "_id")]
    pub idempotency_key: String,
    pub message_id: String,
    pub created_at: DateTime<Utc>,
}

/// Sequence number record
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SequenceNumber {
    #[serde(rename = "_id")]
    pub conversation_id: String,
    pub last_sequence_number: u64,
    pub updated_at: DateTime<Utc>,
}

/// Connection registry entry in Redis
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ConnectionRegistry {
    pub instance_id: String,
    pub connection_id: String,
    pub timestamp: i64,
}

/// Presence status
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Presence {
    pub status: UserStatus,
    pub last_seen: i64,
}

/// CloudEvent for messaging
#[derive(Debug, Clone, Serialize, Deserialize)]
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

/// CloudEvent for agent response
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AgentResponseEvent {
    pub idempotency_key: String,
    pub conversation_id: String,
    pub user_id: String,
    pub agent_id: String,
    pub response: String, // E2EE encrypted
    pub timestamp: DateTime<Utc>,
}

/// Dead Letter Queue entry
#[derive(Debug, Clone, Serialize, Deserialize)]
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

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "snake_case")]
pub enum DLQErrorType {
    BrokerPublishFailed,
    StorageFailed,
    ValidationFailed,
    Timeout,
    ServiceUnavailable,
    NetworkError,
    Unknown,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "snake_case")]
pub enum DLQStatus {
    Pending,
    Retrying,
    Failed,
    Resolved,
}
