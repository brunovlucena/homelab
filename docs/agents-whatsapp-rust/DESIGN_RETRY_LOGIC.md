# Retry Logic with Exponential Backoff Design
## For agents-whatsapp-rust

> **Status**: Design Document  
> **Priority**: P0  
> **Date**: January 2025

---

## Executive Summary

Retry logic with exponential backoff handles transient failures gracefully, preventing message loss while avoiding overwhelming failing services. This design implements a comprehensive retry strategy for all external operations.

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Service Operation                        │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Operation (MongoDB, Redis, Broker, etc.)           │  │
│  └───────────────────┬──────────────────────────────────┘  │
│                      │                                       │
│         ┌────────────▼────────────┐                         │
│         │   Retry Manager         │                         │
│         │  ┌────────────────────┐ │                         │
│         │  │ Retry Strategy     │ │                         │
│         │  │ - Max Retries     │ │                         │
│         │  │ - Backoff Calc    │ │                         │
│         │  │ - Jitter          │ │                         │
│         │  └────────────────────┘ │                         │
│         │  ┌────────────────────┐ │                         │
│         │  │ Error Classifier  │ │                         │
│         │  │ - Transient       │ │                         │
│         │  │ - Permanent      │ │                         │
│         │  └────────────────────┘ │                         │
│         └────────────┬────────────┘                         │
│                      │                                       │
│         ┌────────────▼────────────┐                         │
│         │   Retry Decision        │                         │
│         │  - Retry?               │                         │
│         │  - Backoff Duration?    │                         │
│         │  - Send to DLQ?         │                         │
│         └─────────────────────────┘                         │
└─────────────────────────────────────────────────────────────┘
```

---

## Retry Strategy Matrix

| Operation | Max Retries | Initial Backoff | Max Backoff | Jitter | DLQ After |
|-----------|-------------|-----------------|-------------|--------|-----------|
| MongoDB Write | 5 | 100ms | 30s | ±50ms | Yes |
| MongoDB Read | 3 | 50ms | 10s | ±25ms | No |
| Redis Pub/Sub | 5 | 50ms | 5s | ±25ms | Yes |
| Redis Get/Set | 3 | 50ms | 5s | ±25ms | No |
| Broker Publish | 5 | 100ms | 30s | ±50ms | Yes |
| HTTP Requests | 3 | 200ms | 10s | ±100ms | No |

---

## Error Classification

### Transient Errors (Retry)
- **Network Errors**: Connection timeout, DNS failure
- **Rate Limiting**: 429 Too Many Requests
- **Service Unavailable**: 503 Service Unavailable
- **Timeout**: Request timeout
- **Temporary Failures**: 502 Bad Gateway, 504 Gateway Timeout

### Permanent Errors (No Retry, Send to DLQ)
- **Authentication**: 401 Unauthorized
- **Authorization**: 403 Forbidden
- **Not Found**: 404 Not Found
- **Validation**: 400 Bad Request (malformed data)
- **Conflict**: 409 Conflict (idempotency violation)

---

## Implementation

### 1. Retry Manager

```rust
pub struct RetryManager {
    max_retries: u32,
    initial_backoff_ms: u64,
    max_backoff_ms: u64,
    backoff_multiplier: f64,
    jitter_enabled: bool,
}

impl RetryManager {
    pub fn new(
        max_retries: u32,
        initial_backoff_ms: u64,
        max_backoff_ms: u64,
        backoff_multiplier: f64,
    ) -> Self {
        Self {
            max_retries,
            initial_backoff_ms,
            max_backoff_ms,
            backoff_multiplier,
            jitter_enabled: true,
        }
    }

    pub async fn execute<F, Fut, T, E>(
        &self,
        operation: F,
        error_classifier: impl Fn(&E) -> ErrorType,
    ) -> Result<T, RetryError<E>>
    where
        F: Fn() -> Fut,
        Fut: Future<Output = Result<T, E>>,
    {
        let mut retry_count = 0;
        let mut backoff_ms = self.initial_backoff_ms;

        loop {
            match operation().await {
                Ok(result) => return Ok(result),
                Err(e) => {
                    let error_type = error_classifier(&e);
                    
                    match error_type {
                        ErrorType::Permanent => {
                            return Err(RetryError::PermanentError(e));
                        }
                        ErrorType::Transient if retry_count >= self.max_retries => {
                            return Err(RetryError::MaxRetriesExceeded {
                                error: e,
                                retry_count,
                            });
                        }
                        ErrorType::Transient => {
                            retry_count += 1;
                            let delay = self.calculate_backoff(retry_count);
                            tokio::time::sleep(Duration::from_millis(delay)).await;
                            backoff_ms = self.next_backoff(backoff_ms);
                        }
                    }
                }
            }
        }
    }

    fn calculate_backoff(&self, retry_count: u32) -> u64 {
        let base_delay = (self.initial_backoff_ms as f64)
            * self.backoff_multiplier.powi(retry_count as i32);
        
        let delay_ms = base_delay.min(self.max_backoff_ms as f64) as u64;

        if self.jitter_enabled {
            // Add jitter: ±20% of delay
            use rand::Rng;
            let mut rng = rand::thread_rng();
            let jitter_range = (delay_ms as f64 * 0.2) as u64;
            let jitter = rng.gen_range(0..=jitter_range * 2);
            delay_ms.saturating_sub(jitter_range).saturating_add(jitter)
        } else {
            delay_ms
        }
    }

    fn next_backoff(&self, current: u64) -> u64 {
        (current as f64 * self.backoff_multiplier)
            .min(self.max_backoff_ms as f64) as u64
    }
}
```

### 2. Error Types

```rust
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum ErrorType {
    Transient,
    Permanent,
}

#[derive(Debug)]
pub enum RetryError<E> {
    PermanentError(E),
    MaxRetriesExceeded {
        error: E,
        retry_count: u32,
    },
}
```

### 3. Error Classifiers

```rust
// MongoDB Error Classifier
pub fn classify_mongodb_error(error: &mongodb::error::Error) -> ErrorType {
    match error.kind.as_ref() {
        mongodb::error::ErrorKind::Command(_) => ErrorType::Permanent,
        mongodb::error::ErrorKind::InvalidArgument(_) => ErrorType::Permanent,
        mongodb::error::ErrorKind::Authentication(_) => ErrorType::Permanent,
        mongodb::error::ErrorKind::ConnectionPoolTimeout => ErrorType::Transient,
        mongodb::error::ErrorKind::Io(_) => ErrorType::Transient,
        mongodb::error::ErrorKind::Internal(_) => ErrorType::Transient,
        _ => ErrorType::Transient,
    }
}

// HTTP Error Classifier
pub fn classify_http_error(status: u16) -> ErrorType {
    match status {
        400 | 401 | 403 | 404 | 409 => ErrorType::Permanent,
        429 | 500 | 502 | 503 | 504 => ErrorType::Transient,
        _ => ErrorType::Transient,
    }
}

// Redis Error Classifier
pub fn classify_redis_error(error: &redis::RedisError) -> ErrorType {
    match error.kind() {
        redis::ErrorKind::TypeError => ErrorType::Permanent,
        redis::ErrorKind::AuthenticationFailed => ErrorType::Permanent,
        redis::ErrorKind::IoError => ErrorType::Transient,
        redis::ErrorKind::ExtensionError => ErrorType::Transient,
        _ => ErrorType::Transient,
    }
}
```

---

## Usage Examples

### MongoDB Operations

```rust
impl AppState {
    pub async fn store_message_with_retry(
        &self,
        message: &StoredMessage,
    ) -> AppResult<()> {
        let retry_manager = RetryManager::new(5, 100, 30000, 2.0);
        
        retry_manager
            .execute(
                || async {
                    let collection = self.db.collection::<Document>("messages");
                    let doc = mongodb::bson::to_document(message)?;
                    collection.insert_one(doc, None).await?;
                    Ok::<(), mongodb::error::Error>(())
                },
                classify_mongodb_error,
            )
            .await
            .map_err(|e| match e {
                RetryError::PermanentError(e) => {
                    // Send to DLQ
                    self.send_to_dlq(message, e.to_string()).await?;
                    AppError::Internal("Permanent MongoDB error, sent to DLQ".to_string())
                }
                RetryError::MaxRetriesExceeded { error, retry_count } => {
                    // Send to DLQ
                    self.send_to_dlq(message, error.to_string()).await?;
                    AppError::Internal(format!(
                        "MongoDB operation failed after {} retries, sent to DLQ",
                        retry_count
                    ))
                }
            })
    }
}
```

### Broker Operations

```rust
impl AppState {
    pub async fn publish_to_broker_with_retry(
        &self,
        event: &cloudevents::Event,
    ) -> AppResult<()> {
        let retry_manager = RetryManager::new(5, 100, 30000, 2.0);
        
        retry_manager
            .execute(
                || async {
                    let client = reqwest::Client::new();
                    let response = client
                        .post(&self.config.broker_url)
                        .json(event)
                        .send()
                        .await?;
                    
                    if response.status().is_success() {
                        Ok(())
                    } else {
                        Err(reqwest::Error::from(response.error_for_status().err().unwrap()))
                    }
                },
                |e: &reqwest::Error| {
                    if let Some(status) = e.status() {
                        classify_http_error(status.as_u16())
                    } else {
                        ErrorType::Transient
                    }
                },
            )
            .await
            .map_err(|e| match e {
                RetryError::PermanentError(e) => {
                    // Send to DLQ
                    self.send_to_dlq_from_event(event, e.to_string()).await?;
                    AppError::Internal("Permanent broker error, sent to DLQ".to_string())
                }
                RetryError::MaxRetriesExceeded { error, retry_count } => {
                    // Send to DLQ
                    self.send_to_dlq_from_event(event, error.to_string()).await?;
                    AppError::Internal(format!(
                        "Broker publish failed after {} retries, sent to DLQ",
                        retry_count
                    ))
                }
            })
    }
}
```

---

## Backoff Calculation

### Exponential Backoff Formula

```
delay(n) = min(initial_backoff * multiplier^n, max_backoff) + jitter
```

### Example Progression

| Retry | Base Delay | With Jitter (±20%) | Actual Range |
|-------|------------|-------------------|--------------|
| 1 | 100ms | ±20ms | 80-120ms |
| 2 | 200ms | ±40ms | 160-240ms |
| 3 | 400ms | ±80ms | 320-480ms |
| 4 | 800ms | ±160ms | 640-960ms |
| 5 | 1600ms | ±320ms | 1280-1920ms |

### Jitter Benefits

- **Prevents Thundering Herd**: Staggers retry attempts
- **Reduces Contention**: Avoids synchronized retries
- **Improves Success Rate**: Better distribution of retry times

---

## Metrics & Observability

### Metrics to Track

```rust
pub struct RetryMetrics {
    // Retry counts
    retries_total: Counter,
    retries_by_operation: CounterVec, // operation="mongodb", "redis", "broker"
    
    // Success/failure
    retry_success: Counter,
    retry_failure: Counter,
    
    // Timing
    retry_duration: Histogram,
    backoff_duration: Histogram,
    
    // Error classification
    transient_errors: Counter,
    permanent_errors: Counter,
}
```

### Logging

```rust
// Retry attempt
info!(
    "Retrying operation: operation={}, attempt={}/{}, backoff={}ms",
    operation_name, retry_count, max_retries, backoff_ms
);

// Max retries exceeded
warn!(
    "Max retries exceeded: operation={}, error={}, sending to DLQ",
    operation_name, error
);

// Permanent error
warn!(
    "Permanent error, no retry: operation={}, error={}, sending to DLQ",
    operation_name, error
);
```

---

## Configuration

### Environment Variables

```bash
# Retry Configuration
RETRY_MAX_ATTEMPTS=5
RETRY_INITIAL_BACKOFF_MS=100
RETRY_MAX_BACKOFF_MS=30000
RETRY_BACKOFF_MULTIPLIER=2.0
RETRY_JITTER_ENABLED=true

# Operation-specific overrides
MONGODB_RETRY_MAX_ATTEMPTS=5
MONGODB_RETRY_INITIAL_BACKOFF_MS=100
REDIS_RETRY_MAX_ATTEMPTS=3
REDIS_RETRY_INITIAL_BACKOFF_MS=50
BROKER_RETRY_MAX_ATTEMPTS=5
BROKER_RETRY_INITIAL_BACKOFF_MS=100
```

---

## Testing Strategy

### Unit Tests

1. **Backoff Calculation**
   - Verify exponential progression
   - Verify max backoff cap
   - Verify jitter application

2. **Error Classification**
   - Transient errors → retry
   - Permanent errors → no retry
   - Max retries → DLQ

3. **Retry Logic**
   - Retries on transient errors
   - Stops on permanent errors
   - Stops after max retries

### Integration Tests

1. **MongoDB Retry**
   - Simulate transient failures
   - Verify retry behavior
   - Verify DLQ on max retries

2. **Broker Retry**
   - Simulate network failures
   - Verify exponential backoff
   - Verify DLQ fallback

---

## Implementation Plan

### Phase 1: Core Retry Manager (Week 1)
- [ ] Implement `RetryManager`
- [ ] Add backoff calculation
- [ ] Add jitter support
- [ ] Add unit tests

### Phase 2: Error Classification (Week 1)
- [ ] Implement error classifiers
- [ ] Add MongoDB classifier
- [ ] Add Redis classifier
- [ ] Add HTTP classifier

### Phase 3: Integration (Week 1)
- [ ] Integrate with MongoDB operations
- [ ] Integrate with Redis operations
- [ ] Integrate with Broker operations
- [ ] Add DLQ fallback

### Phase 4: Metrics & Observability (Week 2)
- [ ] Add Prometheus metrics
- [ ] Add logging
- [ ] Create Grafana dashboards
- [ ] Add alerts

---

## Success Criteria

✅ **Transient failures are retried**
- Network errors retry automatically
- Service unavailable errors retry
- Rate limiting errors retry

✅ **Permanent errors skip retries**
- Authentication errors → DLQ immediately
- Validation errors → DLQ immediately
- No wasted retries on permanent errors

✅ **Exponential backoff prevents overload**
- Backoff increases exponentially
- Jitter prevents thundering herd
- Max backoff prevents excessive delays

✅ **DLQ fallback works**
- Failed operations after max retries → DLQ
- Permanent errors → DLQ
- DLQ entries can be replayed

---

## References

- [Exponential Backoff](https://en.wikipedia.org/wiki/Exponential_backoff)
- [Retry Pattern](https://docs.microsoft.com/en-us/azure/architecture/patterns/retry)
- [Jitter in Retry Logic](https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/)

---

**End of Design Document**
