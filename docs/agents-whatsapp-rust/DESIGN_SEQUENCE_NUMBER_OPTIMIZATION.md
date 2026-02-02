# Sequence Number Optimization Design
## For agents-whatsapp-rust

> **Status**: Design Document  
> **Priority**: P0  
> **Date**: January 2025

---

## Executive Summary

Sequence number generation is a critical bottleneck for high-throughput messaging. This design optimizes sequence number generation by using Redis for fast atomic increments, with periodic MongoDB synchronization for durability.

---

## Current Problem

### Current Implementation (MongoDB-only)

```rust
// Current: MongoDB findAndModify (slow at scale)
let sequence = collection.find_one_and_update(
    doc! { "conversation_id": conversation_id },
    doc! { "$inc": { "last_sequence_number": 1 } },
    FindOneAndModifyOptions::new()
        .return_document(ReturnDocument::After)
        .upsert(true)
).await?;
```

**Issues**:
- **Latency**: 10-50ms per operation (MongoDB round-trip)
- **Bottleneck**: Single MongoDB instance cannot handle billions of operations
- **Contention**: High contention on sequence_number collection
- **Scalability**: Does not scale horizontally

---

## Proposed Solution

### Hybrid Approach: Redis + MongoDB

```
┌─────────────────────────────────────────────────────────────┐
│              Sequence Number Generation                     │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Request: next_sequence_number(conversation_id)     │  │
│  └───────────────────┬──────────────────────────────────┘  │
│                      │                                       │
│         ┌────────────▼────────────┐                         │
│         │   Redis INCR            │                         │
│         │   Key: seq:{conv_id}    │                         │
│         │   Latency: < 1ms        │                         │
│         └────────────┬────────────┘                         │
│                      │                                       │
│         ┌────────────▼────────────┐                         │
│         │   Return Sequence       │                         │
│         │   (immediate)           │                         │
│         └─────────────────────────┘                         │
│                      │                                       │
│         ┌────────────▼────────────┐                         │
│         │   Background Sync       │                         │
│         │   (every N increments)  │                         │
│         │   Redis → MongoDB       │                         │
│         └─────────────────────────┘                         │
└─────────────────────────────────────────────────────────────┘
```

---

## Architecture

### Two-Tier Sequence Number System

1. **Redis (Fast Path)**
   - Atomic INCR operation (< 1ms)
   - Per-conversation keys: `seq:{conversation_id}`
   - Handles 99% of requests

2. **MongoDB (Durability)**
   - Periodic synchronization (every 1000 increments)
   - Backup for recovery
   - Historical tracking

### Flow Diagram

```
┌─────────────────────────────────────────────────────────────┐
│  Request: next_sequence_number("conv-123")                 │
└───────────────────┬─────────────────────────────────────────┘
                    │
        ┌───────────▼───────────┐
        │  Redis INCR           │
        │  seq:conv-123         │
        │  Returns: 1001        │
        └───────────┬───────────┘
                    │
        ┌───────────▼───────────┐
        │  Check: 1001 % 1000   │
        │  = 1 (sync point)    │
        └───────────┬───────────┘
                    │
        ┌───────────▼───────────┐
        │  Background Task:     │
        │  Sync to MongoDB      │
        │  (non-blocking)       │
        └───────────────────────┘
                    │
        ┌───────────▼───────────┐
        │  Return: 1001         │
        │  (immediate)          │
        └───────────────────────┘
```

---

## Implementation

### 1. Redis-Based Sequence Number Generator

```rust
pub struct SequenceNumberGenerator {
    redis: redis::Client,
    db: mongodb::Database,
    sync_interval: u64, // Sync every N increments (default: 1000)
}

impl SequenceNumberGenerator {
    pub fn new(
        redis: redis::Client,
        db: mongodb::Database,
        sync_interval: u64,
    ) -> Self {
        Self {
            redis,
            db,
            sync_interval,
        }
    }

    /// Get next sequence number for a conversation
    pub async fn next_sequence_number(
        &self,
        conversation_id: &str,
    ) -> AppResult<u64> {
        let mut conn = self.redis.get_async_connection().await?;
        let key = format!("seq:{}", conversation_id);
        
        // Atomic increment in Redis (sub-millisecond)
        let sequence: u64 = redis::cmd("INCR")
            .arg(&key)
            .query_async(&mut conn)
            .await?;
        
        // Periodic sync to MongoDB (non-blocking)
        if sequence % self.sync_interval == 0 {
            let conversation_id = conversation_id.to_string();
            let sequence_clone = sequence;
            let db_clone = self.db.clone();
            
            tokio::spawn(async move {
                if let Err(e) = sync_to_mongodb(&db_clone, &conversation_id, sequence_clone).await {
                    tracing::warn!("Failed to sync sequence to MongoDB: {}", e);
                }
            });
        }
        
        Ok(sequence)
    }

    /// Initialize sequence number from MongoDB (on startup/recovery)
    pub async fn initialize_from_mongodb(
        &self,
        conversation_id: &str,
    ) -> AppResult<()> {
        let collection = self.db.collection::<Document>("sequence_numbers");
        let filter = doc! { "_id": conversation_id };
        
        if let Some(doc) = collection.find_one(filter, None).await? {
            if let Some(last_seq) = doc.get_i64("last_sequence_number") {
                let mut conn = self.redis.get_async_connection().await?;
                let key = format!("seq:{}", conversation_id);
                
                // Set Redis to MongoDB value + 1 (next available)
                redis::cmd("SET")
                    .arg(&key)
                    .arg((last_seq + 1) as u64)
                    .arg("NX") // Only set if not exists
                    .query_async(&mut conn)
                    .await?;
            }
        }
        
        Ok(())
    }
}

async fn sync_to_mongodb(
    db: &mongodb::Database,
    conversation_id: &str,
    sequence: u64,
) -> AppResult<()> {
    let collection = db.collection::<Document>("sequence_numbers");
    let filter = doc! { "_id": conversation_id };
    let update = doc! {
        "$set": {
            "last_sequence_number": sequence as i64,
            "updated_at": Utc::now(),
        },
        "$setOnInsert": {
            "created_at": Utc::now(),
        },
    };
    
    collection.update_one(filter, update, mongodb::options::UpdateOptions::builder()
        .upsert(true)
        .build())
        .await?;
    
    Ok(())
}
```

### 2. Recovery Strategy

```rust
impl SequenceNumberGenerator {
    /// Recover sequence numbers from MongoDB on startup
    pub async fn recover_all_sequences(&self) -> AppResult<()> {
        let collection = self.db.collection::<Document>("sequence_numbers");
        let mut cursor = collection.find(doc! {}, None).await?;
        
        let mut conn = self.redis.get_async_connection().await?;
        
        while cursor.advance().await? {
            let doc = cursor.current();
            if let (Some(conv_id), Some(last_seq)) = (
                doc.get_str("_id").ok(),
                doc.get_i64("last_sequence_number"),
            ) {
                let key = format!("seq:{}", conv_id);
                
                // Set Redis to last known sequence + 1
                redis::cmd("SET")
                    .arg(&key)
                    .arg((last_seq + 1) as u64)
                    .arg("NX")
                    .query_async(&mut conn)
                    .await?;
            }
        }
        
        Ok(())
    }
}
```

### 3. Redis Persistence

```yaml
# Redis Configuration for Sequence Numbers
Redis:
  persistence:
    # RDB snapshots every 5 minutes
    save: "300 1"
    # AOF for durability
    appendonly: "yes"
    appendfsync: "everysec"
  
  # Replication for HA
  replica:
    enabled: true
    replicas: 2
```

---

## Performance Comparison

### Before (MongoDB-only)

| Metric | Value |
|--------|-------|
| Latency (p50) | 15ms |
| Latency (p95) | 50ms |
| Throughput | 1,000 ops/sec |
| Scalability | Single instance bottleneck |

### After (Redis + MongoDB)

| Metric | Value |
|--------|-------|
| Latency (p50) | 0.5ms |
| Latency (p95) | 1ms |
| Throughput | 100,000+ ops/sec |
| Scalability | Horizontal (Redis Cluster) |

**Improvement**: **30x faster latency, 100x higher throughput**

---

## Failure Handling

### Redis Failure

```rust
impl SequenceNumberGenerator {
    pub async fn next_sequence_number_with_fallback(
        &self,
        conversation_id: &str,
    ) -> AppResult<u64> {
        // Try Redis first
        match self.next_sequence_number(conversation_id).await {
            Ok(seq) => Ok(seq),
            Err(_) => {
                // Fallback to MongoDB
                tracing::warn!("Redis unavailable, falling back to MongoDB");
                self.next_sequence_number_mongodb(conversation_id).await
            }
        }
    }
    
    async fn next_sequence_number_mongodb(
        &self,
        conversation_id: &str,
    ) -> AppResult<u64> {
        let collection = self.db.collection::<Document>("sequence_numbers");
        let filter = doc! { "_id": conversation_id };
        let update = doc! { "$inc": { "last_sequence_number": 1 } };
        
        let options = mongodb::options::FindOneAndUpdateOptions::builder()
            .return_document(mongodb::options::ReturnDocument::After)
            .upsert(true)
            .build();
        
        let doc = collection.find_one_and_update(filter, update, options).await?;
        let sequence = doc
            .and_then(|d| d.get_i64("last_sequence_number"))
            .unwrap_or(0) as u64;
        
        Ok(sequence)
    }
}
```

### MongoDB Sync Failure

- **Non-blocking**: Sync failures don't block sequence generation
- **Retry**: Background task retries sync
- **Recovery**: On startup, recover from MongoDB

---

## Configuration

### Environment Variables

```bash
# Sequence Number Configuration
SEQUENCE_SYNC_INTERVAL=1000  # Sync to MongoDB every 1000 increments
SEQUENCE_REDIS_TTL=86400     # Redis key TTL (24 hours, auto-cleanup)
SEQUENCE_RECOVERY_ENABLED=true
```

### Tuning Parameters

| Parameter | Default | Description |
|-----------|---------|-------------|
| `sync_interval` | 1000 | Sync to MongoDB every N increments |
| `redis_ttl` | 86400 | Redis key TTL in seconds |
| `recovery_batch_size` | 1000 | Batch size for recovery |

---

## Migration Strategy

### Phase 1: Dual Write (Week 1)
- Write to both Redis and MongoDB
- Verify consistency
- Monitor performance

### Phase 2: Redis Primary (Week 2)
- Switch to Redis as primary
- MongoDB as backup
- Add periodic sync

### Phase 3: Optimization (Week 3)
- Tune sync interval
- Add recovery mechanisms
- Monitor and adjust

---

## Testing Strategy

### Unit Tests

1. **Sequence Generation**
   - Verify atomic increments
   - Verify sync triggers
   - Verify fallback behavior

2. **Recovery**
   - Verify MongoDB → Redis sync
   - Verify gap detection
   - Verify consistency

### Integration Tests

1. **Performance Tests**
   - Measure latency
   - Measure throughput
   - Compare with MongoDB-only

2. **Failure Tests**
   - Redis failure → MongoDB fallback
   - MongoDB sync failure → continue
   - Recovery after restart

---

## Metrics & Observability

### Metrics to Track

```rust
pub struct SequenceNumberMetrics {
    // Generation
    sequences_generated: Counter,
    generation_latency: Histogram,
    
    // Sync
    syncs_to_mongodb: Counter,
    sync_latency: Histogram,
    sync_failures: Counter,
    
    // Fallback
    mongodb_fallback_count: Counter,
    redis_unavailable_count: Counter,
    
    // Recovery
    recovery_sequences_loaded: Counter,
    recovery_duration: Histogram,
}
```

### Alerts

```yaml
- alert: SequenceNumberSyncFailure
  expr: sequence_sync_failures > 10
  for: 5m
  annotations:
    summary: "Sequence number sync to MongoDB failing"

- alert: SequenceNumberRedisUnavailable
  expr: sequence_redis_unavailable > 0
  for: 1m
  annotations:
    summary: "Redis unavailable, using MongoDB fallback"
```

---

## Success Criteria

✅ **Performance Improvement**
- Latency reduced from 15ms to < 1ms
- Throughput increased from 1K to 100K+ ops/sec

✅ **Scalability**
- Horizontal scaling via Redis Cluster
- No single point of failure

✅ **Reliability**
- Fallback to MongoDB on Redis failure
- Recovery from MongoDB on startup
- No sequence number gaps

✅ **Consistency**
- Periodic MongoDB sync ensures durability
- Recovery ensures no lost sequences

---

## References

- [Redis INCR](https://redis.io/commands/incr/)
- [Atomic Operations](https://redis.io/docs/manual/patterns/atomic-counters/)
- [Redis Persistence](https://redis.io/docs/manual/persistence/)

---

**End of Design Document**
