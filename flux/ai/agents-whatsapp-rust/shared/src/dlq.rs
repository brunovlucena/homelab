use crate::errors::{AppError, AppResult};
use crate::models::{DLQErrorType, DLQStatus, DeadLetterQueueEntry, StoredMessage};
use chrono::{Duration, Utc};
use mongodb::bson::{doc, Document};
use mongodb::{Database, IndexModel};
use std::time::Duration as StdDuration;
use tracing::{info, warn};

/// Dead Letter Queue manager
pub struct DeadLetterQueue {
    db: Database,
    max_retries: u32,
    retry_backoff_base_ms: u64,
    retry_backoff_multiplier: f64,
    dlq_ttl_days: i64,
}

impl DeadLetterQueue {
    pub fn new(db: Database) -> Self {
        Self {
            db,
            max_retries: 5,
            retry_backoff_base_ms: 100,
            retry_backoff_multiplier: 2.0,
            dlq_ttl_days: 7,
        }
    }

    /// Add a failed message to DLQ
    pub async fn add(
        &self,
        message: StoredMessage,
        error: String,
        error_type: DLQErrorType,
        retry_count: u32,
    ) -> AppResult<String> {
        let dlq_id = uuid::Uuid::new_v4().to_string();

        let next_retry_at = if retry_count < self.max_retries {
            let backoff_ms = self.calculate_backoff(retry_count);
            Some(Utc::now() + Duration::milliseconds(backoff_ms as i64))
        } else {
            None
        };

        let entry = DeadLetterQueueEntry {
            id: Some(dlq_id.clone()),
            message,
            error: error.clone(),
            error_type,
            retry_count,
            max_retries: self.max_retries,
            next_retry_at,
            created_at: Utc::now(),
            last_retry_at: None,
            status: if retry_count < self.max_retries {
                DLQStatus::Pending
            } else {
                DLQStatus::Failed
            },
            resolved_at: None,
            resolved_reason: None,
        };

        let collection = self.db.collection::<Document>("dead_letter_queue");
        let mut doc = mongodb::bson::to_document(&entry)
            .map_err(|e| AppError::Internal(format!("Failed to serialize DLQ entry: {}", e)))?;

        // Add TTL index expiration (MongoDB TTL index will auto-delete)
        let expires_at = Utc::now() + Duration::days(self.dlq_ttl_days);
        doc.insert("expires_at", expires_at);

        collection.insert_one(doc, None).await?;

        info!(
            "Added message to DLQ: id={}, retry_count={}, error={}",
            dlq_id, retry_count, error
        );

        Ok(dlq_id)
    }

    /// Retry failed messages that are due for retry
    pub async fn retry_pending_messages(&self) -> AppResult<usize> {
        let collection = self.db.collection::<Document>("dead_letter_queue");

        let filter = doc! {
            "status": "pending",
            "next_retry_at": { "$lte": Utc::now() },
            "retry_count": { "$lt": self.max_retries },
        };

        let mut count = 0;
        let mut cursor = collection.find(filter, None).await?;

        while cursor.advance().await? {
            let raw_doc = cursor.current();
            let doc = mongodb::bson::from_slice(raw_doc.as_bytes())
                .map_err(|e| AppError::Internal(format!("Failed to parse document: {}", e)))?;

            if let Ok(entry) = mongodb::bson::from_document::<DeadLetterQueueEntry>(doc) {
                // Update status to retrying
                let id = entry
                    .id
                    .as_ref()
                    .ok_or_else(|| AppError::Internal("DLQ entry missing id".to_string()))?;
                collection
                    .update_one(
                        doc! { "_id": id },
                        doc! {
                            "$set": {
                                "status": "retrying",
                                "last_retry_at": Utc::now(),
                            }
                        },
                        None,
                    )
                    .await?;

                // Here you would trigger the retry logic
                // For now, we'll just log it
                info!(
                    "Retrying DLQ message: id={}, retry_count={}",
                    entry.id.as_ref().unwrap_or(&"unknown".to_string()),
                    entry.retry_count
                );

                count += 1;
            }
        }

        Ok(count)
    }

    /// Get DLQ entries by status
    pub async fn get_entries_by_status(
        &self,
        status: DLQStatus,
        limit: Option<i64>,
    ) -> AppResult<Vec<DeadLetterQueueEntry>> {
        let collection = self.db.collection::<Document>("dead_letter_queue");

        let status_str = match status {
            DLQStatus::Pending => "pending",
            DLQStatus::Retrying => "retrying",
            DLQStatus::Failed => "failed",
            DLQStatus::Resolved => "resolved",
        };
        let filter = doc! {
            "status": status_str,
        };

        let mut options = mongodb::options::FindOptions::default();
        if let Some(limit) = limit {
            options.limit = Some(limit);
        }
        options.sort = Some(doc! { "created_at": -1 });

        let mut entries = Vec::new();
        let mut cursor = collection.find(filter, Some(options)).await?;

        while cursor.advance().await? {
            let raw_doc = cursor.current();
            let doc = mongodb::bson::from_slice(raw_doc.as_bytes())
                .map_err(|e| AppError::Internal(format!("Failed to parse document: {}", e)))?;
            if let Ok(entry) = mongodb::bson::from_document::<DeadLetterQueueEntry>(doc) {
                entries.push(entry);
            }
        }

        Ok(entries)
    }

    /// Mark a DLQ entry as resolved
    pub async fn mark_resolved(&self, dlq_id: &str, reason: Option<String>) -> AppResult<()> {
        let collection = self.db.collection::<Document>("dead_letter_queue");

        collection
            .update_one(
                doc! { "_id": dlq_id },
                doc! {
                    "$set": {
                        "status": "resolved",
                        "resolved_at": Utc::now(),
                        "resolved_reason": reason,
                    }
                },
                None,
            )
            .await?;

        info!("Marked DLQ entry as resolved: id={}", dlq_id);
        Ok(())
    }

    /// Cleanup expired DLQ entries
    pub async fn cleanup_expired(&self) -> AppResult<usize> {
        let collection = self.db.collection::<Document>("dead_letter_queue");

        let filter = doc! {
            "expires_at": { "$lt": Utc::now() },
        };

        let result = collection.delete_many(filter, None).await?;
        let count = result.deleted_count as usize;

        if count > 0 {
            info!("Cleaned up {} expired DLQ entries", count);
        }

        Ok(count)
    }

    /// Calculate exponential backoff delay in milliseconds
    fn calculate_backoff(&self, retry_count: u32) -> u64 {
        let delay_ms = (self.retry_backoff_base_ms as f64)
            * self.retry_backoff_multiplier.powi(retry_count as i32);

        // Cap at 5 minutes
        delay_ms.min(300_000.0) as u64
    }

    /// Get DLQ statistics
    pub async fn get_statistics(&self) -> AppResult<DLQStatistics> {
        let collection = self.db.collection::<Document>("dead_letter_queue");

        let total = collection.count_documents(doc! {}, None).await? as usize;

        let pending = collection
            .count_documents(doc! { "status": "pending" }, None)
            .await? as usize;

        let failed = collection
            .count_documents(doc! { "status": "failed" }, None)
            .await? as usize;

        let retrying = collection
            .count_documents(doc! { "status": "retrying" }, None)
            .await? as usize;

        Ok(DLQStatistics {
            total,
            pending,
            failed,
            retrying,
        })
    }
}

#[derive(Debug, Clone)]
pub struct DLQStatistics {
    pub total: usize,
    pub pending: usize,
    pub failed: usize,
    pub retrying: usize,
}

/// Initialize DLQ collection with indexes
/// Retries with exponential backoff if MongoDB is not ready yet
pub async fn init_dlq_collection(db: &Database) -> AppResult<()> {
    let collection = db.collection::<Document>("dead_letter_queue");

    // Create indexes
    let indexes = vec![
        IndexModel::builder()
            .keys(doc! { "status": 1, "next_retry_at": 1 })
            .options(
                mongodb::options::IndexOptions::builder()
                    .name("status_next_retry_idx".to_string())
                    .build(),
            )
            .build(),
        IndexModel::builder()
            .keys(doc! { "expires_at": 1 })
            .options(
                mongodb::options::IndexOptions::builder()
                    .name("expires_at_idx".to_string())
                    .expire_after(Some(StdDuration::ZERO)) // TTL index - MongoDB will auto-delete expired docs
                    .build(),
            )
            .build(),
        IndexModel::builder()
            .keys(doc! { "created_at": -1 })
            .options(
                mongodb::options::IndexOptions::builder()
                    .name("created_at_idx".to_string())
                    .build(),
            )
            .build(),
        IndexModel::builder()
            .keys(doc! { "message.message_id": 1 })
            .options(
                mongodb::options::IndexOptions::builder()
                    .name("message_id_idx".to_string())
                    .build(),
            )
            .build(),
    ];

    // Retry logic with exponential backoff
    const MAX_RETRIES: u32 = 10;
    const INITIAL_DELAY_MS: u64 = 100;
    const MAX_DELAY_MS: u64 = 5000;

    for attempt in 0..MAX_RETRIES {
        // First, try a ping to ensure MongoDB is ready
        match db.run_command(doc! { "ping": 1 }, None).await {
            Ok(_) => {
                // MongoDB is ready, try to create indexes
                match collection.create_indexes(indexes.clone(), None).await {
                    Ok(_) => {
                        info!("Initialized DLQ collection with indexes");
                        return Ok(());
                    }
                    Err(e) => {
                        if attempt < MAX_RETRIES - 1 {
                            let delay_ms =
                                std::cmp::min(INITIAL_DELAY_MS * (1 << attempt), MAX_DELAY_MS);
                            warn!(
                                "Failed to create DLQ indexes (attempt {}/{}): {}. Retrying in {}ms",
                                attempt + 1,
                                MAX_RETRIES,
                                e,
                                delay_ms
                            );
                            tokio::time::sleep(tokio::time::Duration::from_millis(delay_ms)).await;
                        } else {
                            return Err(AppError::Internal(format!(
                                "Failed to create DLQ indexes after {} attempts: {}",
                                MAX_RETRIES, e
                            )));
                        }
                    }
                }
            }
            Err(e) => {
                if attempt < MAX_RETRIES - 1 {
                    let delay_ms = std::cmp::min(INITIAL_DELAY_MS * (1 << attempt), MAX_DELAY_MS);
                    warn!(
                        "MongoDB ping failed (attempt {}/{}): {}. Retrying in {}ms",
                        attempt + 1,
                        MAX_RETRIES,
                        e,
                        delay_ms
                    );
                    tokio::time::sleep(tokio::time::Duration::from_millis(delay_ms)).await;
                } else {
                    return Err(AppError::Internal(format!(
                        "MongoDB not ready after {} ping attempts: {}",
                        MAX_RETRIES, e
                    )));
                }
            }
        }
    }

    Err(AppError::Internal(
        "Failed to initialize DLQ collection: max retries exceeded".to_string(),
    ))
}
