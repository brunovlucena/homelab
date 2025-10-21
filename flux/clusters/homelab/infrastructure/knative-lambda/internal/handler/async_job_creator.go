// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	⚡ ASYNC JOB CREATOR - Parallel job creation with worker pool
//
//	🎯 Purpose: Handle parallel Kaniko job creation to improve throughput
//	💡 Features: Worker pool, async processing, result tracking, error handling
//
//	🏛️ ARCHITECTURE:
//	🎯 Worker Pool - Multiple goroutines for parallel job creation
//	📊 Result Tracking - Track job creation results and status
//	🔄 Async Processing - Non-blocking job creation with callbacks
//	🛡️ Error Handling - Comprehensive error handling and retry logic
//
//	🔄 PROCESSING FLOW:
//	1. Job creation request → Add to work queue → Worker picks up → Create job → Return result
//	2. Multiple workers can process jobs simultaneously
//	3. Results are tracked and can be retrieved asynchronously
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package handler

import (
	"context"
	"fmt"
	"sync"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	"knative-lambda-new/internal/observability"
	"knative-lambda-new/pkg/builds"
)

// 🎯 JobCreationRequest - "Request for job creation"
type JobCreationRequest struct {
	BuildRequest *builds.BuildRequest
	JobName      string
	CorrelationID string
	CreatedAt    time.Time
}

// 🎯 JobCreationResult - "Result of job creation attempt"
type JobCreationResult struct {
	Request     *JobCreationRequest
	Job         *batchv1.Job
	Error       error
	Attempts    int
	Duration    time.Duration
	CompletedAt time.Time
}

// 🎯 AsyncJobCreator - "Parallel job creation with worker pool"
// Implements AsyncJobCreatorInterface
type AsyncJobCreator struct {
	// 🔧 Configuration
	workerCount    int
	maxRetries     int
	retryDelay     time.Duration
	maxQueueSize   int

	// 🎯 Worker Pool
	workQueue      chan *JobCreationRequest
	resultQueue    chan *JobCreationResult
	workers        []*jobWorker
	workerWg       sync.WaitGroup

	// 📊 State Management
	results        map[string]*JobCreationResult
	resultsMutex   sync.RWMutex
	shutdownChan   chan struct{}
	isShutdown     bool
	shutdownMutex  sync.RWMutex

	// 🔧 Dependencies
	jobManager     JobManager
	obs            *observability.Observability
}

// 🎯 jobWorker - "Individual worker for job creation"
type jobWorker struct {
	id             int
	creator        *AsyncJobCreator
	ctx            context.Context
	cancel         context.CancelFunc
}

// 🏗️ NewAsyncJobCreator - "Create new async job creator"
func NewAsyncJobCreator(config AsyncJobCreatorConfig) *AsyncJobCreator {
	creator := &AsyncJobCreator{
		workerCount:   config.WorkerCount,
		maxRetries:    config.MaxRetries,
		retryDelay:    config.RetryDelay,
		maxQueueSize:  config.MaxQueueSize,
		workQueue:     make(chan *JobCreationRequest, config.MaxQueueSize),
		resultQueue:   make(chan *JobCreationResult, config.MaxQueueSize),
		results:       make(map[string]*JobCreationResult),
		shutdownChan:  make(chan struct{}),
		jobManager:    config.JobManager,
		obs:           config.Observability,
	}

	// Start workers
	creator.startWorkers()

	// Start result processor
	go creator.processResults()

	return creator
}

// 🎯 AsyncJobCreatorConfig - "Configuration for async job creator"
type AsyncJobCreatorConfig struct {
	WorkerCount   int
	MaxRetries    int
	RetryDelay    time.Duration
	MaxQueueSize  int
	JobManager    JobManager
	Observability *observability.Observability
}

// ⚡ CreateJobAsync - "Create job asynchronously"
func (c *AsyncJobCreator) CreateJobAsync(ctx context.Context, buildRequest *builds.BuildRequest) (string, error) {
	c.shutdownMutex.RLock()
	if c.isShutdown {
		c.shutdownMutex.RUnlock()
		return "", fmt.Errorf("async job creator is shutdown")
	}
	c.shutdownMutex.RUnlock()

	// Generate job name
	jobName := c.jobManager.GenerateJobName(buildRequest.ThirdPartyID, buildRequest.ParserID)
	correlationID := buildRequest.CorrelationID

	// Create request
	request := &JobCreationRequest{
		BuildRequest:  buildRequest,
		JobName:       jobName,
		CorrelationID: correlationID,
		CreatedAt:     time.Now(),
	}

	// Add to work queue (non-blocking)
	select {
	case c.workQueue <- request:
		c.obs.Info(ctx, "Job creation request queued successfully",
			"job_name", jobName,
			"third_party_id", buildRequest.ThirdPartyID,
			"parser_id", buildRequest.ParserID,
			"correlation_id", correlationID,
			"queue_size", len(c.workQueue))
		return jobName, nil
	default:
		return "", fmt.Errorf("work queue is full, cannot queue job creation request")
	}
}

// 📊 GetJobCreationResult - "Get result of job creation"
func (c *AsyncJobCreator) GetJobCreationResult(correlationID string) (*JobCreationResult, bool) {
	c.resultsMutex.RLock()
	defer c.resultsMutex.RUnlock()

	result, exists := c.results[correlationID]
	return result, exists
}

// 📊 WaitForJobCreation - "Wait for job creation to complete"
func (c *AsyncJobCreator) WaitForJobCreation(ctx context.Context, correlationID string) (*JobCreationResult, error) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			if result, exists := c.GetJobCreationResult(correlationID); exists {
				return result, nil
			}
		}
	}
}

// 🔄 startWorkers - "Start worker pool"
func (c *AsyncJobCreator) startWorkers() {
	for i := 0; i < c.workerCount; i++ {
		workerCtx, cancel := context.WithCancel(context.Background())
		worker := &jobWorker{
			id:      i,
			creator: c,
			ctx:     workerCtx,
			cancel:  cancel,
		}

		c.workers = append(c.workers, worker)
		c.workerWg.Add(1)

		go func(w *jobWorker) {
			defer c.workerWg.Done()
			w.run()
		}(worker)
	}

	c.obs.Info(context.Background(), "Started async job creator workers",
		"worker_count", c.workerCount,
		"max_queue_size", c.maxQueueSize)
}

// 🔄 run - "Worker main loop"
func (w *jobWorker) run() {
	for {
		select {
		case <-w.ctx.Done():
			return
		case request := <-w.creator.workQueue:
			w.processJobCreation(request)
		}
	}
}

// 🔄 processJobCreation - "Process individual job creation"
func (w *jobWorker) processJobCreation(request *JobCreationRequest) {
	startTime := time.Now()
	ctx := context.Background()

	w.creator.obs.Info(ctx, "Worker processing job creation",
		"worker_id", w.id,
		"job_name", request.JobName,
		"third_party_id", request.BuildRequest.ThirdPartyID,
		"parser_id", request.BuildRequest.ParserID,
		"correlation_id", request.CorrelationID)

	// Try to create the job with retry logic
	var job *batchv1.Job
	var err error
	var attempts int

	for attempt := 0; attempt < w.creator.maxRetries; attempt++ {
		attempts = attempt + 1
		
		job, err = w.creator.jobManager.CreateJob(ctx, request.JobName, request.BuildRequest)
		if err == nil {
			break // Success, exit retry loop
		}

		// Check if this is a conflict error (job already exists)
		if w.isConflictError(err) {
			w.creator.obs.Info(ctx, "Job already exists, checking if it's the same job",
				"worker_id", w.id,
				"job_name", request.JobName,
				"attempt", attempts,
				"max_retries", w.creator.maxRetries,
				"third_party_id", request.BuildRequest.ThirdPartyID,
				"parser_id", request.BuildRequest.ParserID)

			// Check if the existing job is for the same parser
			existingJob, findErr := w.creator.jobManager.FindExistingJob(ctx, request.BuildRequest.ThirdPartyID, request.BuildRequest.ParserID)
			if findErr == nil && existingJob != nil {
				w.creator.obs.Info(ctx, "Found existing job for same parser",
					"worker_id", w.id,
					"job_name", existingJob.Name,
					"third_party_id", request.BuildRequest.ThirdPartyID,
					"parser_id", request.BuildRequest.ParserID)

				// Create success result with existing job
				result := &JobCreationResult{
					Request:     request,
					Job:         existingJob,
					Error:       nil,
					Attempts:    attempts,
					Duration:    time.Since(startTime),
					CompletedAt: time.Now(),
				}

				w.creator.resultQueue <- result
				return
			}

			// If we can't find the existing job, wait and retry
			if attempt < w.creator.maxRetries-1 {
				retryDelay := w.creator.retryDelay * time.Duration(1<<attempt) // Exponential backoff
				w.creator.obs.Info(ctx, "Waiting before retry",
					"worker_id", w.id,
					"job_name", request.JobName,
					"attempt", attempts,
					"retry_delay", retryDelay,
					"third_party_id", request.BuildRequest.ThirdPartyID,
					"parser_id", request.BuildRequest.ParserID)

				select {
				case <-w.ctx.Done():
					return
				case <-time.After(retryDelay):
					continue
				}
			}
		}

		// For non-conflict errors, don't retry
		break
	}

	// Create result
	result := &JobCreationResult{
		Request:     request,
		Job:         job,
		Error:       err,
		Attempts:    attempts,
		Duration:    time.Since(startTime),
		CompletedAt: time.Now(),
	}

	// Send result to result queue
	w.creator.resultQueue <- result

	if err != nil {
		w.creator.obs.Error(ctx, err, "Failed to create job after retries",
			"worker_id", w.id,
			"job_name", request.JobName,
			"max_retries", w.creator.maxRetries,
			"third_party_id", request.BuildRequest.ThirdPartyID,
			"parser_id", request.BuildRequest.ParserID,
			"attempts", attempts,
			"duration", result.Duration)
	} else {
		w.creator.obs.Info(ctx, "Successfully created job",
			"worker_id", w.id,
			"job_name", job.Name,
			"job_uid", string(job.UID),
			"third_party_id", request.BuildRequest.ThirdPartyID,
			"parser_id", request.BuildRequest.ParserID,
			"attempts", attempts,
			"duration", result.Duration)
	}
}

// 🔄 processResults - "Process results from workers"
func (c *AsyncJobCreator) processResults() {
	for {
		select {
		case <-c.shutdownChan:
			return
		case result := <-c.resultQueue:
			c.resultsMutex.Lock()
			c.results[result.Request.CorrelationID] = result
			c.resultsMutex.Unlock()

			c.obs.Info(context.Background(), "Job creation result processed",
				"correlation_id", result.Request.CorrelationID,
				"job_name", result.Request.JobName,
				"success", result.Error == nil,
				"attempts", result.Attempts,
				"duration", result.Duration)
		}
	}
}

// 🔍 isConflictError - "Check if error is a conflict error"
func (w *jobWorker) isConflictError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return w.creator.isConflictErrorString(errStr)
}

// 🔍 isConflictErrorString - "Check if error string indicates conflict"
func (c *AsyncJobCreator) isConflictErrorString(errStr string) bool {
	return len(errStr) > 0 && (len(errStr) > 12 && errStr[:12] == "already exists" || len(errStr) > 12 && errStr[:12] == "AlreadyExists")
}

// 🛑 Shutdown - "Gracefully shutdown async job creator"
func (c *AsyncJobCreator) Shutdown(ctx context.Context) error {
	c.shutdownMutex.Lock()
	if c.isShutdown {
		c.shutdownMutex.Unlock()
		return nil
	}
	c.isShutdown = true
	c.shutdownMutex.Unlock()

	// Signal shutdown
	close(c.shutdownChan)

	// Cancel all workers
	for _, worker := range c.workers {
		worker.cancel()
	}

	// Wait for workers to finish
	done := make(chan struct{})
	go func() {
		c.workerWg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		c.obs.Info(ctx, "Async job creator shutdown completed",
			"worker_count", c.workerCount)
		return nil
	}
}

// 📊 GetStats - "Get async job creator statistics"
func (c *AsyncJobCreator) GetStats() map[string]interface{} {
	c.resultsMutex.RLock()
	defer c.resultsMutex.RUnlock()

	stats := map[string]interface{}{
		"worker_count":    c.workerCount,
		"queue_size":      len(c.workQueue),
		"max_queue_size":  c.maxQueueSize,
		"results_count":   len(c.results),
		"is_shutdown":     c.isShutdown,
	}

	return stats
}
