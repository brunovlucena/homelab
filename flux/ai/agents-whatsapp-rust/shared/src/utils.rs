use uuid::Uuid;

/// Generate a new UUID v4 for idempotency keys
pub fn generate_idempotency_key() -> String {
    Uuid::new_v4().to_string()
}

/// Generate a new message ID
pub fn generate_message_id() -> String {
    format!("msg_{}", Uuid::new_v4())
}

/// Generate a new conversation ID
pub fn generate_conversation_id() -> String {
    format!("conv_{}", Uuid::new_v4())
}

/// Generate a new session ID
pub fn generate_session_id() -> String {
    format!("sess_{}", Uuid::new_v4())
}
