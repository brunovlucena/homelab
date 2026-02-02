use crate::AppState;
use axum::{extract::State, Json};
use chrono::Utc;
use cloudevents::{AttributesReader, Event, EventBuilder};
use mongodb::bson::doc;
use shared::dlq::DeadLetterQueue;
use shared::models::{DLQErrorType, MessageStatus, MessageType, StoredMessage};
use tracing::{error, info};

pub async fn handle_event(
    State(state): State<std::sync::Arc<AppState>>,
    Json(event): Json<Event>,
) -> std::result::Result<&'static str, shared::errors::AppError> {
    // Only process messaging.message.received events
    if event.ty() != "messaging.message.received" {
        return Ok("OK");
    }

    // Extract message data
    let data = event
        .data()
        .and_then(|d| match d {
            cloudevents::Data::Json(v) => Some(v),
            _ => None,
        })
        .ok_or_else(|| shared::errors::AppError::Validation("Missing event data".to_string()))?;

    let idempotency_key = data
        .get("idempotency_key")
        .and_then(|v| v.as_str())
        .ok_or_else(|| {
            shared::errors::AppError::Validation("Missing idempotency_key".to_string())
        })?;

    // Check idempotency (defense in depth)
    let collection = state
        .db
        .collection::<mongodb::bson::Document>("idempotency_keys");
    let filter = doc! { "_id": idempotency_key };
    let existing = collection.find_one(filter, None).await?;

    if existing.is_some() {
        info!(
            "Duplicate message detected (idempotency key: {})",
            idempotency_key
        );
        return Ok("OK");
    }

    let conversation_id = data
        .get("conversation_id")
        .and_then(|v| v.as_str())
        .ok_or_else(|| {
            shared::errors::AppError::Validation("Missing conversation_id".to_string())
        })?;

    // Determine agent from conversation context
    let conv_collection = state
        .db
        .collection::<mongodb::bson::Document>("conversations");
    let conv_filter = doc! { "_id": conversation_id };

    let agent_id = if let Some(conv) = conv_collection.find_one(conv_filter, None).await? {
        // Use agent_id from existing conversation
        conv.get_str("agent_id")
            .map(String::from)
            .unwrap_or_else(|_| "agent-bruno".to_string())
    } else {
        // New conversation - determine agent based on message content/intent
        // For now, default to agent-bruno, but this can be enhanced with:
        // - Content analysis (keywords, intent detection)
        // - User preferences
        // - Conversation type
        // - Historical routing patterns
        "agent-bruno".to_string()
    };

    // Create agent.message CloudEvent
    let agent_event = cloudevents::EventBuilderV10::new()
        .id(uuid::Uuid::new_v4().to_string())
        .source(state.config.broker_url.clone())
        .ty("agent.message")
        .data("application/json", data.clone())
        .build()
        .map_err(|e| {
            shared::errors::AppError::Internal(format!("Failed to build CloudEvent: {}", e))
        })?;

    // Publish to Knative Broker with retry and DLQ
    let broker_url = state.config.broker_url.clone();
    let event_data_clone = data.clone();
    let conversation_id_clone = conversation_id.to_string();

    tokio::spawn(async move {
        if let Err(e) = publish_agent_event_with_retry(
            &broker_url,
            &agent_event,
            &event_data_clone,
            &conversation_id_clone,
            &state,
        )
        .await
        {
            error!("Failed to publish agent message after retries: {}", e);
        } else {
            info!("Routed message to agent: {}", agent_id);
        }
    });

    Ok("OK")
}

async fn publish_agent_event_with_retry(
    broker_url: &str,
    event: &cloudevents::Event,
    event_data: &serde_json::Value,
    conversation_id: &str,
    state: &std::sync::Arc<AppState>,
) -> Result<(), shared::errors::AppError> {
    let max_retries = 5;
    let mut retry_count = 0;
    let mut backoff_ms = 100;

    // Create a StoredMessage-like structure for DLQ (if needed)
    let message_for_dlq = StoredMessage {
        message_id: event_data
            .get("message_id")
            .and_then(|v| v.as_str())
            .map(String::from),
        idempotency_key: event_data
            .get("idempotency_key")
            .and_then(|v| v.as_str())
            .unwrap_or("")
            .to_string(),
        conversation_id: conversation_id.to_string(),
        sequence_number: event_data
            .get("sequence_number")
            .and_then(|v| v.as_u64())
            .unwrap_or(0),
        sender_id: event_data
            .get("sender_id")
            .and_then(|v| v.as_str())
            .unwrap_or("")
            .to_string(),
        receiver_id: event_data
            .get("receiver_id")
            .and_then(|v| v.as_str())
            .unwrap_or("")
            .to_string(),
        message_type: MessageType::Text,
        content: event_data
            .get("content")
            .and_then(|v| v.as_str())
            .unwrap_or("")
            .to_string(),
        media_url: None,
        reply_to_message_id: None,
        timestamp: Utc::now(),
        status: MessageStatus::Sent,
        created_at: Utc::now(),
    };

    loop {
        let client = reqwest::Client::builder()
            .timeout(std::time::Duration::from_secs(10))
            .build()
            .map_err(|e| {
                shared::errors::AppError::Internal(format!("Failed to create HTTP client: {}", e))
            })?;

        match client.post(broker_url).json(event).send().await {
            Ok(response) if response.status().is_success() => {
                return Ok(());
            }
            Ok(response) => {
                let error_msg = format!("Broker returned error status: {}", response.status());
                if retry_count >= max_retries {
                    let db = state.db.clone();
                    let dlq = DeadLetterQueue::new(db);
                    let _ = dlq
                        .add(
                            message_for_dlq.clone(),
                            error_msg.clone(),
                            DLQErrorType::BrokerPublishFailed,
                            retry_count,
                        )
                        .await;
                    return Err(shared::errors::AppError::Internal(error_msg));
                }
                retry_count += 1;
            }
            Err(e) => {
                let error_msg = format!("Failed to publish to broker: {}", e);
                if retry_count >= max_retries {
                    let db = state.db.clone();
                    let dlq = DeadLetterQueue::new(db);
                    let _ = dlq
                        .add(
                            message_for_dlq.clone(),
                            error_msg.clone(),
                            DLQErrorType::BrokerPublishFailed,
                            retry_count,
                        )
                        .await;
                    return Err(shared::errors::AppError::Internal(error_msg));
                }
                retry_count += 1;
            }
        }

        // Exponential backoff with jitter
        let jitter = rand::random::<u64>() % 50;
        let delay_ms = backoff_ms + jitter;
        tokio::time::sleep(std::time::Duration::from_millis(delay_ms)).await;

        backoff_ms = (backoff_ms as f64 * 2.0).min(30000.0) as u64;
    }
}
