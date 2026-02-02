# 🔄 BACKEND-004: Async Job Processing with Worker Pools

**Priority**: P0 | **Status**: ✅ Implemented  | **Story Points**: 8  
**Linear URL**: https://linear.app/bvlucena/issue/BVL-217/backend-004-async-job-processing-with-worker-pools

---

## 📋 User Story

**As a** Backend Developer  
**I want to** process job creation requests asynchronously using worker pools  
**So that** HTTP requests return immediately while jobs are created in the background, improving API responsiveness

---

## 🎯 Acceptance Criteria

### ✅ Worker Pool Management
- [ ] Create configurable worker pool (default: 5 workers)
- [ ] Initialize workers on service startup
- [ ] Gracefully shutdown workers on service termination
- [ ] Handle worker panics without crashing service
- [ ] Support worker pool scaling based on queue depth

### ✅ Job Queue Management
- [ ] Buffered channel for job requests (default: 100 capacity)
- [ ] FIFO queue processing
- [ ] Block new requests when queue is full
- [ ] Track queue depth metrics
- [ ] Support priority-based queuing (future)

### ✅ Async Job Creation
- [ ] Accept job creation requests from event handler
- [ ] Return immediate response with job name
- [ ] Process job creation in background worker
- [ ] Handle job creation failures in worker
- [ ] Emit metrics for async operations

### ✅ Result Tracking
- [ ] Store job creation results in memory
- [ ] Support result retrieval by correlation ID
- [ ] Implement result TTL (default: 1 hour)
- [ ] Clean up old results periodically
- [ ] Provide stats endpoint for monitoring

### ✅ Error Handling
- [ ] Handle queue full scenarios
- [ ] Handle worker crashes
- [ ] Handle job creation failures
- [ ] Retry failed operations (configurable)
- [ ] Log all errors with context

---

## 🔧 Technical Implementation

### File: `internal/handler/async_job_creator.go`

```go
// Async Job Creator Interface
type AsyncJobCreator interface {
    CreateJobAsync(ctx context.Context, buildRequest *builds.BuildRequest) (string, error)
    GetStats() map[string]interface{}
    Shutdown(ctx context.Context) error
}

// Implementation
type AsyncJobCreatorImpl struct {
    jobManager  JobManager
    obs         *observability.Observability
    workerCount int
    queueSize   int
    
    // Worker pool channels
    jobQueue    chan *asyncJobRequest
    resultMap   sync.Map // correlation_id -> result
    shutdownCh  chan struct{}
    wg          sync.WaitGroup
}

// Async job request structure
type asyncJobRequest struct {
    ctx          context.Context
    buildRequest *builds.BuildRequest
    resultChan   chan *asyncJobResult
}

type asyncJobResult struct {
    jobName string
    err     error
}

// Create Job Asynchronously
func (a *AsyncJobCreatorImpl) CreateJobAsync(ctx context.Context, buildRequest *builds.BuildRequest) (string, error) {
    ctx, span := a.obs.StartSpan(ctx, "create_job_async")
    defer span.End()
    
    // Generate job name immediately
    jobName := a.jobManager.GenerateJobName(buildRequest.ThirdPartyID, buildRequest.ParserID)
    
    // Create result channel
    resultChan := make(chan *asyncJobResult, 1)
    
    // Create async request
    req := &asyncJobRequest{
        ctx:          ctx,
        buildRequest: buildRequest,
        resultChan:   resultChan,
    }
    
    // Try to enqueue job
    select {
    case a.jobQueue <- req:
        a.obs.Info(ctx, "Job creation request queued",
            "job_name", jobName,
            "queue_depth", len(a.jobQueue))
        
        // Return immediately with job name
        // Actual creation happens in background
        return jobName, nil
        
    default:
        // Queue is full
        a.obs.Error(ctx, fmt.Errorf("job queue full"), "Job queue is full, rejecting request",
            "queue_size", a.queueSize,
            "job_name", jobName)
        return "", fmt.Errorf("job queue full, please retry later")
    }
}

// Worker goroutine
func (a *AsyncJobCreatorImpl) worker(workerID int) {
    defer a.wg.Done()
    
    a.obs.Info(context.Background(), "Worker started", "worker_id", workerID)
    
    for {
        select {
        case req := <-a.jobQueue:
            // Process job creation
            a.processJobCreation(workerID, req)
            
        case <-a.shutdownCh:
            a.obs.Info(context.Background(), "Worker shutting down", "worker_id", workerID)
            return
        }
    }
}

func (a *AsyncJobCreatorImpl) processJobCreation(workerID int, req *asyncJobRequest) {
    ctx := req.ctx
    
    ctx, span := a.obs.StartSpan(ctx, "process_job_creation_worker")
    defer span.End()
    
    a.obs.Info(ctx, "Worker processing job creation",
        "worker_id", workerID,
        "third_party_id", req.buildRequest.ThirdPartyID,
        "parser_id", req.buildRequest.ParserID)
    
    // Create job using job manager
    jobName := a.jobManager.GenerateJobName(req.buildRequest.ThirdPartyID, req.buildRequest.ParserID)
    
    job, err := a.jobManager.CreateJob(ctx, jobName, req.buildRequest)
    
    result := &asyncJobResult{
        jobName: jobName,
        err:     err,
    }
    
    if err != nil {
        a.obs.Error(ctx, err, "Worker failed to create job",
            "worker_id", workerID,
            "job_name", jobName)
    } else {
        a.obs.Info(ctx, "Worker successfully created job",
            "worker_id", workerID,
            "job_name", job.Name,
            "job_uid", string(job.UID))
    }
    
    // Send result back
    select {
    case req.resultChan <- result:
    default:
        // Result channel closed or full, ignore
    }
}

// Get Statistics
func (a *AsyncJobCreatorImpl) GetStats() map[string]interface{} {
    return map[string]interface{}{
        "worker_count":   a.workerCount,
        "queue_size":     len(a.jobQueue),
        "max_queue_size": a.queueSize,
    }
}
```

---

## 📊 Worker Pool Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    HTTP Request Handler                      │
│  (CloudEvent processing, immediate response)                 │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ↓
              ┌──────────────────────┐
              │   Job Queue          │
              │  (Buffered Channel)  │
              │   Capacity: 100      │
              └──────────┬───────────┘
                         │
         ┌───────────────┼───────────────┐
         ↓               ↓               ↓
    ┌────────┐      ┌────────┐      ┌────────┐
    │Worker 1│      │Worker 2│ ...  │Worker N│
    └────┬───┘      └────┬───┘      └────┬───┘
         │               │               │
         └───────────────┼───────────────┘
                         ↓
              ┌──────────────────────┐
              │   Job Manager        │
              │  (Kubernetes API)    │
              └──────────────────────┘
```

---

## 🧪 Testing Scenarios

### 1. Normal Async Processing
```bash
# Send 10 build requests rapidly
for i in {1..10}; do
  curl -X POST http://localhost:8080/events \
    -H "Ce-Type: network.notifi.lambda.build.start" \
    -H "Ce-Source: network.notifi.customer-$i" \
    -d "{\"parser_id\":\"parser-$i\"}" &
done
```

**Expected**:
- All requests return immediately with 200 OK
- Jobs queued for processing
- Workers process jobs in background
- Check stats: `curl http://localhost:8080/async-jobs/stats`

### 2. Queue Full Scenario
```bash
# Fill queue beyond capacity
for i in {1..150}; do
  curl -X POST http://localhost:8080/events \
    -H "Ce-Type: network.notifi.lambda.build.start" \
    -d "{\"parser_id\":\"parser-$i\"}" &
done
```

**Expected**:
- First 100 requests queued successfully
- Remaining 50 requests return error: "job queue full"
- No service crash or degradation

### 3. Worker Failure Handling
```bash
# Send request that causes worker panic
curl -X POST http://localhost:8080/events \
  -H "Ce-Type: network.notifi.lambda.build.start" \
  -d "{\"parser_id\":null}"
```

**Expected**:
- Worker catches panic
- Error logged
- Worker continues processing other jobs
- Service remains healthy

### 4. Graceful Shutdown
```bash
# Send SIGTERM to service
kubectl delete pod knative-lambda-builder-xxx -n knative-lambda

# Monitor graceful shutdown
kubectl logs -f knative-lambda-builder-xxx -n knative-lambda
```

**Expected**:
- Service stops accepting new requests
- Workers finish processing current jobs
- Graceful shutdown within 30s
- No job loss or corruption

---

## 📈 Performance Requirements

- **Queue Throughput**: 100 jobs/second
- **Worker Latency**: < 5s per job creation
- **Queue Response Time**: < 100ms (immediate return)
- **Memory Overhead**: < 50MB for queue + workers
- **Shutdown Time**: < 30s for graceful shutdown

---

## 🔍 Monitoring & Alerts

### Metrics
- `async_job_queue_depth` - Current queue depth
- `async_job_queue_full_total` - Queue full rejections
- `async_job_processing_duration_seconds` - Worker processing time
- `async_worker_errors_total` - Worker error count
- `async_worker_panics_total` - Worker panic count

### Alerts
- **Queue Near Full**: Alert if queue > 80% capacity for 5 minutes
- **Worker Errors**: Alert if error rate > 10% over 5 minutes
- **Worker Panics**: Alert on any worker panic
- **Slow Processing**: Alert if p95 worker latency > 10s

### Stats Endpoint
```bash
curl http://localhost:8080/async-jobs/stats
```

```json
{
  "worker_count": 5,
  "queue_size": 23,
  "max_queue_size": 100,
  "results_count": 150,
  "is_shutdown": false
}
```

---

## 🏗️ Code References

**Main Files**:
- `internal/handler/async_job_creator.go` - Worker pool implementation
- `internal/handler/event_handler.go` - Integration with event processing
- `internal/config/build.go` - Worker pool configuration

**Configuration**:
```go
type BuildConfig struct {
    WorkerCount int  // default: 5
    QueueSize   int  // default: 100
}
```

---

## 📚 Related Documentation

- [BACKEND-003: Kubernetes Job Lifecycle](BACKEND-003-kubernetes-job-lifecycle.md)
- [BACKEND-005: Rate Limiting and Resilience](BACKEND-005-rate-limiting-resilience.md)
- Go Concurrency Patterns: https://go.dev/blog/pipelines

---

**Last Updated**: October 29, 2025  
**Owner**: Backend Team  
**Status**: Production Ready

