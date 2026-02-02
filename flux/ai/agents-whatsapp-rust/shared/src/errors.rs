use axum::{
    http::StatusCode,
    response::{IntoResponse, Response},
};
use thiserror::Error;

#[derive(Error, Debug)]
pub enum AppError {
    #[error("Database error: {0}")]
    Database(#[from] mongodb::error::Error),

    #[error("Redis error: {0}")]
    Redis(#[from] redis::RedisError),

    #[error("Serialization error: {0}")]
    Serialization(#[from] serde_json::Error),

    #[error("BSON deserialization error: {0}")]
    BsonDeserialization(#[from] mongodb::bson::de::Error),

    #[error("BSON serialization error: {0}")]
    BsonSerialization(#[from] mongodb::bson::ser::Error),

    #[error("WebSocket error: {0}")]
    WebSocket(String),

    #[error("Authentication error: {0}")]
    Authentication(String),

    #[error("Validation error: {0}")]
    Validation(String),

    #[error("Idempotency key already exists")]
    DuplicateIdempotencyKey,

    #[error("Message not found")]
    MessageNotFound,

    #[error("Conversation not found")]
    ConversationNotFound,

    #[error("User not found")]
    UserNotFound,

    #[error("Internal error: {0}")]
    Internal(String),
}

pub type AppResult<T> = Result<T, AppError>;

impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        let status = match self {
            AppError::Authentication(_) => StatusCode::UNAUTHORIZED,
            AppError::Validation(_) => StatusCode::BAD_REQUEST,
            AppError::UserNotFound | AppError::MessageNotFound | AppError::ConversationNotFound => {
                StatusCode::NOT_FOUND
            }
            AppError::DuplicateIdempotencyKey => StatusCode::CONFLICT,
            _ => StatusCode::INTERNAL_SERVER_ERROR,
        };
        (status, self.to_string()).into_response()
    }
}
