# Circuit Breakers Design
## For agents-whatsapp-rust

> **Status**: Design Document  
> **Priority**: P0  
> **Date**: January 2025

---

## Executive Summary

Circuit breakers prevent cascading failures by stopping requests to failing services, allowing them to recover. This design implements circuit breakers for MongoDB, Redis, and Knative Broker to ensure system resilience.

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Service Layer                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │  Messaging   │  │ Agent Gateway│  │   Storage    │     │
│  │   Service    │  │              │  │   Service    │     │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘     │
│         │                 │                 │              │
│         └─────────────────┼─────────────────┘              │
│                           │                                 │
│         ┌─────────────────▼─────────────────┐              │
│         │    Circuit Breaker Layer          │              │
│         │  ┌──────────┐  ┌──────────┐      │              │
│         │  │ MongoDB  │  │  Redis   │      │              │
│         │  │ Circuit  │  │ Circuit  │      │              │
│         │  └──────────┘  └──────────┘      │              │
│         │  ┌──────────┐                    │              │
│         │  │ Broker  │                    │              │
│         │  │ Circuit │                    │              │
│         │  └──────────┘                    │              │
│         └──────────────────────────────────┘              │
│                           │                                 │
│         ┌─────────────────▼─────────────────┐              │
│         │      External Dependencies        │              │
│         │  ┌──────────┐  ┌──────────┐      │              │
│         │  │ MongoDB  │  │  Redis   │      │              │
│         │  └──────────┘  └──────────┘      │              │
│         │  ┌──────────┐                    │              │
│         │  │ Broker   │                    │              │
│         │  └──────────┘                    │              │
│         └──────────────────────────────────┘              │
└─────────────────────────────────────────────────────────────┘
```

---

## Circuit Breaker States

### 1. Closed (Normal Operation)
- **State**: Requests flow through normally
- **Behavior**: All requests pass through
- **Failure Tracking**: Count failures
- **Transition**: Move to Open after threshold

### 2. Open (Failing)
- **State**: Requests are immediately rejected
- **Behavior**: Fast-fail without calling dependency
- **Recovery**: Wait for timeout, then move to Half-Open
- **Transition**: After recovery timeout expires

### 3. Half-Open (Testing Recovery)
- **State**: Allow limited requests to test recovery
- **Behavior**: Allow 1 request, monitor result
- **Success**: Move to Closed
- **Failure**: Move back to Open

---

## Configuration

### MongoDB Circuit Breaker

```rust
pub struct MongoDBCircuitBreaker {
    failure_threshold: usize,      // 5 failures
    recovery_timeout: Duration,     // 30 seconds
    half_open_max_requests: usize,  // 1 request
    state: Arc<RwLock<CircuitState>>,
    failure_count: Arc<AtomicUsize>,
    last_failure_time: Arc<RwLock<Option<Instant>>>,
}

impl MongoDBCircuitBreaker {
    pub fn new() -> Self {
        Self {
            failure_threshold: 5,
            recovery_timeout: Duration::from_secs(30),
            half_open_max_requests: 1,
            state: Arc::new(RwLock::new(CircuitState::Closed)),
            failure_count: Arc::new(AtomicUsize::new(0)),
            last_failure_time: Arc::new(RwLock::new(None)),
        }
    }
}
```

### Redis Circuit Breaker

```rust
pub struct RedisCircuitBreaker {
    failure_threshold: usize,      // 5 failures
    recovery_timeout: Duration,     // 10 seconds (Redis recovers faster)
    half_open_max_requests: usize,  // 1 request
    state: Arc<RwLock<CircuitState>>,
    failure_count: Arc<AtomicUsize>,
    last_failure_time: Arc<RwLock<Option<Instant>>>,
}
```

### Broker Circuit Breaker

```rust
pub struct BrokerCircuitBreaker {
    failure_threshold: usize,      // 3 failures (broker is critical)
    recovery_timeout: Duration,     // 60 seconds
    half_open_max_requests: usize,  // 1 request
    state: Arc<RwLock<CircuitState>>,
    failure_count: Arc<AtomicUsize>,
    last_failure_time: Arc<RwLock<Option<Instant>>>,
}
```

---

## Implementation Strategy

### 1. Circuit Breaker Trait

```rust
#[async_trait]
pub trait CircuitBreaker {
    async fn call<F, Fut, T, E>(&self, f: F) -> Result<T, CircuitBreakerError<E>>
    where
        F: FnOnce() -> Fut,
        Fut: Future<Output = Result<T, E>>;
    
    fn is_open(&self) -> bool;
    fn is_closed(&self) -> bool;
    fn is_half_open(&self) -> bool;
    
    async fn record_success(&self);
    async fn record_failure(&self);
}
```

### 2. State Management

```rust
enum CircuitState {
    Closed {
        failure_count: usize,
    },
    Open {
        opened_at: Instant,
    },
    HalfOpen {
        test_request_count: usize,
    },
}
```

### 3. Error Types

```rust
#[derive(Debug)]
pub enum CircuitBreakerError<E> {
    CircuitOpen {
        opened_at: Instant,
        recovery_timeout: Duration,
    },
    ServiceError(E),
}
```

---

## Usage Patterns

### MongoDB Operations

```rust
impl AppState {
    pub async fn store_message(
        &self,
        message: &StoredMessage,
    ) -> AppResult<()> {
        self.mongodb_circuit.call(|| async {
            let collection = self.db.collection::<Document>("messages");
            let doc = mongodb::bson::to_document(message)?;
            collection.insert_one(doc, None).await?;
            Ok(())
        }).await
        .map_err(|e| match e {
            CircuitBreakerError::CircuitOpen { .. } => {
                // Fallback: Store in Redis temporarily
                self.fallback_store_redis(message).await?;
                AppError::Internal("MongoDB circuit open, using Redis fallback".to_string())
            }
            CircuitBreakerError::ServiceError(e) => e.into(),
        })
    }
}
```

### Redis Operations

```rust
impl AppState {
    pub async fn publish_to_redis(
        &self,
        channel: &str,
        message: &str,
    ) -> AppResult<()> {
        self.redis_circuit.call(|| async {
            let mut conn = self.redis.get_async_connection().await?;
            redis::cmd("PUBLISH")
                .arg(channel)
                .arg(message)
                .query_async(&mut conn)
                .await?;
            Ok(())
        }).await
        .map_err(|e| match e {
            CircuitBreakerError::CircuitOpen { .. } => {
                // Fallback: Queue in MongoDB for later processing
                self.fallback_queue_mongodb(channel, message).await?;
                AppError::Internal("Redis circuit open, using MongoDB fallback".to_string())
            }
            CircuitBreakerError::ServiceError(e) => e.into(),
        })
    }
}
```

### Broker Operations

```rust
impl AppState {
    pub async fn publish_to_broker(
        &self,
        event: &cloudevents::Event,
    ) -> AppResult<()> {
        self.broker_circuit.call(|| async {
            let client = reqwest::Client::new();
            client
                .post(&self.config.broker_url)
                .json(event)
                .send()
                .await?;
            Ok(())
        }).await
        .map_err(|e| match e {
            CircuitBreakerError::CircuitOpen { .. } => {
                // Fallback: Send to DLQ immediately
                self.send_to_dlq(event).await?;
                AppError::Internal("Broker circuit open, sent to DLQ".to_string())
            }
            CircuitBreakerError::ServiceError(e) => e.into(),
        })
    }
}
```

---

## Fallback Strategies

### MongoDB Fallback
- **Primary**: MongoDB
- **Fallback**: Redis (temporary storage)
- **Recovery**: Sync from Redis to MongoDB when circuit closes

### Redis Fallback
- **Primary**: Redis Pub/Sub
- **Fallback**: MongoDB queue
- **Recovery**: Process MongoDB queue when circuit closes

### Broker Fallback
- **Primary**: Knative Broker
- **Fallback**: DLQ (immediate)
- **Recovery**: Retry from DLQ when circuit closes

---

## Metrics & Observability

### Metrics to Track

```rust
pub struct CircuitBreakerMetrics {
    // State transitions
    transitions_to_open: Counter,
    transitions_to_closed: Counter,
    transitions_to_half_open: Counter,
    
    // Request counts
    requests_allowed: Counter,
    requests_rejected: Counter,
    
    // Timing
    time_in_open_state: Histogram,
    time_in_half_open_state: Histogram,
    
    // Current state
    current_state: Gauge, // 0=Closed, 1=Open, 2=HalfOpen
}
```

### Logging

```rust
// State transitions
info!(
    "Circuit breaker state transition: {} -> {}",
    old_state, new_state
);

// Rejections
warn!(
    "Circuit breaker rejected request: service={}, state={}",
    service_name, state
);
```

---

## Testing Strategy

### Unit Tests

1. **State Transitions**
   - Closed → Open (after threshold)
   - Open → Half-Open (after timeout)
   - Half-Open → Closed (on success)
   - Half-Open → Open (on failure)

2. **Failure Counting**
   - Count increments on failure
   - Count resets on success
   - Count resets on state transition

3. **Timeout Handling**
   - Recovery timeout respected
   - State transitions at correct times

### Integration Tests

1. **MongoDB Circuit Breaker**
   - Simulate MongoDB failures
   - Verify fallback to Redis
   - Verify recovery sync

2. **Redis Circuit Breaker**
   - Simulate Redis failures
   - Verify fallback to MongoDB
   - Verify recovery processing

3. **Broker Circuit Breaker**
   - Simulate Broker failures
   - Verify DLQ fallback
   - Verify retry from DLQ

---

## Configuration Options

### Environment Variables

```bash
# MongoDB Circuit Breaker
MONGODB_CIRCUIT_FAILURE_THRESHOLD=5
MONGODB_CIRCUIT_RECOVERY_TIMEOUT=30s
MONGODB_CIRCUIT_HALF_OPEN_MAX_REQUESTS=1

# Redis Circuit Breaker
REDIS_CIRCUIT_FAILURE_THRESHOLD=5
REDIS_CIRCUIT_RECOVERY_TIMEOUT=10s
REDIS_CIRCUIT_HALF_OPEN_MAX_REQUESTS=1

# Broker Circuit Breaker
BROKER_CIRCUIT_FAILURE_THRESHOLD=3
BROKER_CIRCUIT_RECOVERY_TIMEOUT=60s
BROKER_CIRCUIT_HALF_OPEN_MAX_REQUESTS=1
```

---

## Implementation Plan

### Phase 1: Core Circuit Breaker (Week 1)
- [ ] Implement `CircuitBreaker` trait
- [ ] Implement state management
- [ ] Add unit tests

### Phase 2: MongoDB Circuit Breaker (Week 1)
- [ ] Implement MongoDB circuit breaker
- [ ] Add fallback to Redis
- [ ] Add recovery sync
- [ ] Add integration tests

### Phase 3: Redis Circuit Breaker (Week 1)
- [ ] Implement Redis circuit breaker
- [ ] Add fallback to MongoDB
- [ ] Add recovery processing
- [ ] Add integration tests

### Phase 4: Broker Circuit Breaker (Week 1)
- [ ] Implement Broker circuit breaker
- [ ] Integrate with DLQ
- [ ] Add retry from DLQ
- [ ] Add integration tests

### Phase 5: Metrics & Observability (Week 2)
- [ ] Add Prometheus metrics
- [ ] Add logging
- [ ] Create Grafana dashboards
- [ ] Add alerts

---

## Success Criteria

✅ **Circuit breakers prevent cascading failures**
- Services fail fast when dependencies are down
- No resource exhaustion from retrying failed requests

✅ **Fallback strategies work**
- MongoDB failures → Redis fallback
- Redis failures → MongoDB fallback
- Broker failures → DLQ fallback

✅ **Automatic recovery**
- Circuits close automatically when dependencies recover
- Data syncs correctly after recovery

✅ **Observability**
- Metrics track circuit breaker state
- Alerts fire when circuits open
- Dashboards show circuit breaker health

---

## References

- [Circuit Breaker Pattern](https://martinfowler.com/bliki/CircuitBreaker.html)
- [Resilience4j Circuit Breaker](https://resilience4j.readme.io/docs/circuitbreaker)
- [Hystrix Circuit Breaker](https://github.com/Netflix/Hystrix/wiki/How-it-Works#CircuitBreaker)

---

**End of Design Document**
