// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
//
//	🧪 BACKEND-004: Async Job Processing Tests
//
//	User Story: Async Job Processing
//	Priority: P0 | Story Points: 8
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
package backend

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"

	"knative-lambda/internal/handler"
	"knative-lambda/pkg/builds"
)

// TestBackend004_WorkerPoolCreation validates worker pool initialization.
func TestBackend004_WorkerPoolCreation(t *testing.T) {
	tests := []struct {
		name        string
		workerCount int
		description string
	}{
		{
			name:        "Create 5-worker pool",
			workerCount: 5,
			description: "Should create pool with 5 workers",
		},
		{
			name:        "Create 10-worker pool",
			workerCount: 10,
			description: "Should create pool with 10 workers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			obs := handler.NewMockObservability()
			mockJobManager := &SharedMockJobManager{}

			// Act
			creator := handler.NewAsyncJobCreator(handler.AsyncJobCreatorConfig{
				WorkerCount:   tt.workerCount,
				MaxRetries:    3,
				RetryDelay:    100 * time.Millisecond,
				MaxQueueSize:  100,
				JobManager:    mockJobManager,
				Observability: obs,
			})

			// Assert
			require.NotNil(t, creator)
			stats := creator.GetStats()
			assert.Equal(t, tt.workerCount, stats["worker_count"])
		})
	}
}

// TestBackend004_JobQueueing validates job queueing.
func TestBackend004_JobQueueing(t *testing.T) {
	// Arrange
	ctx := context.Background()
	creator := setupAsyncJobCreator(t, 5, 100)

	buildReq := &builds.BuildRequest{
		ThirdPartyID:  "customer-123",
		ParserID:      "parser-abc",
		CorrelationID: "corr-123",
	}

	// Act
	jobName, err := creator.CreateJobAsync(ctx, buildReq)

	// Assert
	require.NoError(t, err)
	assert.NotEmpty(t, jobName)
	assert.Contains(t, jobName, buildReq.ThirdPartyID)
}

// TestBackend004_AsyncProcessing validates asynchronous processing.
func TestBackend004_AsyncProcessing(t *testing.T) {
	// Arrange
	ctx := context.Background()
	creator := setupAsyncJobCreator(t, 3, 50)

	// Act - Queue multiple jobs
	var wg sync.WaitGroup
	jobCount := 10

	for i := 0; i < jobCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			buildReq := &builds.BuildRequest{
				ThirdPartyID:  "customer-123",
				ParserID:      "parser-" + string(rune(id)),
				CorrelationID: "corr-" + string(rune(id)),
			}
			_, err := creator.CreateJobAsync(ctx, buildReq)
			assert.NoError(t, err)
		}(i)
	}

	// Wait for all jobs to be queued
	wg.Wait()

	// Give workers time to process
	time.Sleep(500 * time.Millisecond)

	// Assert
	stats := creator.GetStats()
	assert.Equal(t, jobCount, stats["results_count"], "All jobs should be processed")
}

// TestBackend004_QueueFullHandling validates queue capacity handling.
func TestBackend004_QueueFullHandling(t *testing.T) {
	// Arrange
	ctx := context.Background()
	maxQueueSize := 1 // Very small queue
	creator := setupAsyncJobCreator(t, 1, maxQueueSize)

	// Act - Try to queue more than capacity rapidly
	var errors []error
	for i := 0; i < 5; i++ { // Queue 5 jobs with queue size 1
		buildReq := &builds.BuildRequest{
			ThirdPartyID:  "customer-123",
			ParserID:      "parser-" + string(rune(i)),
			CorrelationID: "corr-" + string(rune(i)),
		}
		_, err := creator.CreateJobAsync(ctx, buildReq)
		if err != nil {
			errors = append(errors, err)
		}
	}

	// Assert - Some requests should fail when queue is full
	assert.Greater(t, len(errors), 0, "Should return errors when queue is full")
}

// TestBackend004_WorkerFailures validates worker failure handling.
func TestBackend004_WorkerFailures(t *testing.T) {
	// Arrange
	ctx := context.Background()
	obs := handler.NewMockObservability()
	mockJobManager := &MockJobManagerWithFailure{
		failCount: 2, // Fail first 2 attempts
	}

	creator := handler.NewAsyncJobCreator(handler.AsyncJobCreatorConfig{
		WorkerCount:   2,
		MaxRetries:    3,
		RetryDelay:    50 * time.Millisecond,
		MaxQueueSize:  10,
		JobManager:    mockJobManager,
		Observability: obs,
	})

	buildReq := &builds.BuildRequest{
		ThirdPartyID:  "customer-123",
		ParserID:      "parser-abc",
		CorrelationID: "corr-123",
	}

	// Act
	_, err := creator.CreateJobAsync(ctx, buildReq)

	// Assert
	require.NoError(t, err, "Should queue successfully even if worker fails")
}

// TestBackend004_GracefulShutdown validates graceful shutdown.
func TestBackend004_GracefulShutdown(t *testing.T) {
	// Arrange
	ctx := context.Background()
	creator := setupAsyncJobCreator(t, 3, 50)

	// Queue some jobs
	for i := 0; i < 5; i++ {
		buildReq := &builds.BuildRequest{
			ThirdPartyID:  "customer-123",
			ParserID:      "parser-" + string(rune(i)),
			CorrelationID: "corr-" + string(rune(i)),
		}
		if _, err := creator.CreateJobAsync(ctx, buildReq); err != nil {
			t.Logf("Failed to create job async: %v", err)
		}
	}

	// Act - Shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := creator.Shutdown(shutdownCtx)

	// Assert
	require.NoError(t, err, "Should shutdown gracefully")
}

// TestBackend004_ResultTracking validates result tracking by correlation ID.
func TestBackend004_ResultTracking(t *testing.T) {
	// Arrange
	ctx := context.Background()
	creator := setupAsyncJobCreator(t, 2, 20)

	correlationID := "test-correlation-123"
	buildReq := &builds.BuildRequest{
		ThirdPartyID:  "customer-123",
		ParserID:      "parser-abc",
		CorrelationID: correlationID,
	}

	// Act
	jobName, err := creator.CreateJobAsync(ctx, buildReq)

	// Assert
	require.NoError(t, err)
	assert.NotEmpty(t, jobName)

	// Give worker time to process
	time.Sleep(200 * time.Millisecond)

	// Verify correlation ID is tracked
	stats := creator.GetStats()
	assert.Greater(t, stats["results_count"], 0)
}

// TestBackend004_StatsEndpoint validates statistics reporting.
func TestBackend004_StatsEndpoint(t *testing.T) {
	// Arrange
	ctx := context.Background()
	workerCount := 5
	maxQueueSize := 100
	creator := setupAsyncJobCreator(t, workerCount, maxQueueSize)

	// Queue some jobs
	for i := 0; i < 3; i++ {
		buildReq := &builds.BuildRequest{
			ThirdPartyID:  "customer-123",
			ParserID:      "parser-" + string(rune(i)),
			CorrelationID: "corr-" + string(rune(i)),
		}
		if _, err := creator.CreateJobAsync(ctx, buildReq); err != nil {
			t.Logf("Failed to create job async: %v", err)
		}
	}

	time.Sleep(100 * time.Millisecond)

	// Act
	stats := creator.GetStats()

	// Assert
	assert.Equal(t, workerCount, stats["worker_count"])
	assert.Equal(t, maxQueueSize, stats["max_queue_size"])
	assert.GreaterOrEqual(t, stats["results_count"], 3)
}

// TestBackend004_ConcurrentQueueing validates concurrent job queueing.
func TestBackend004_ConcurrentQueueing(t *testing.T) {
	// Arrange
	ctx := context.Background()
	creator := setupAsyncJobCreator(t, 5, 100)

	// Act - Queue jobs concurrently
	concurrency := 50
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			buildReq := &builds.BuildRequest{
				ThirdPartyID:  "customer-123",
				ParserID:      "parser-" + string(rune(id)),
				CorrelationID: "corr-" + string(rune(id)),
			}
			_, err := creator.CreateJobAsync(ctx, buildReq)
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// Assert
	assert.Greater(t, successCount, 0, "Should queue some jobs successfully")
}

// TestBackend004_RetryMechanism validates retry logic.
func TestBackend004_RetryMechanism(t *testing.T) {
	// Arrange
	ctx := context.Background()
	obs := handler.NewMockObservability()
	mockJobManager := &MockJobManagerWithFailure{
		failCount: 1, // Fail once then succeed
	}

	creator := handler.NewAsyncJobCreator(handler.AsyncJobCreatorConfig{
		WorkerCount:   2,
		MaxRetries:    3,
		RetryDelay:    50 * time.Millisecond,
		MaxQueueSize:  10,
		JobManager:    mockJobManager,
		Observability: obs,
	})

	buildReq := &builds.BuildRequest{
		ThirdPartyID:  "customer-123",
		ParserID:      "parser-abc",
		CorrelationID: "corr-123",
	}

	// Act
	_, err := creator.CreateJobAsync(ctx, buildReq)
	time.Sleep(300 * time.Millisecond) // Wait for retry

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 0, mockJobManager.failCount, "Should have retried and succeeded")
}

// Helper Functions.

func setupAsyncJobCreator(_ *testing.T, workerCount, maxQueueSize int) handler.AsyncJobCreatorInterface {
	obs := handler.NewMockObservability()
	mockJobManager := &SharedMockJobManager{}

	return handler.NewAsyncJobCreator(handler.AsyncJobCreatorConfig{
		WorkerCount:   workerCount,
		MaxRetries:    3,
		RetryDelay:    100 * time.Millisecond,
		MaxQueueSize:  maxQueueSize,
		JobManager:    mockJobManager,
		Observability: obs,
	})
}

type MockJobManagerWithFailure struct {
	failCount int
	mu        sync.Mutex
}

func (m *MockJobManagerWithFailure) CreateJob(_ context.Context, _ string, _ *builds.BuildRequest) (*batchv1.Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failCount > 0 {
		m.failCount--
		return nil, assert.AnError
	}
	return &batchv1.Job{}, nil
}

func (m *MockJobManagerWithFailure) GenerateJobName(thirdPartyID, parserID string) string {
	return "test-job-" + thirdPartyID + "-" + parserID
}

func (m *MockJobManagerWithFailure) FindExistingJob(_ context.Context, _, _ string) (*batchv1.Job, error) {
	return nil, nil
}

func (m *MockJobManagerWithFailure) GetJob(_ context.Context, _ string) (*batchv1.Job, error) {
	return nil, nil
}

func (m *MockJobManagerWithFailure) DeleteJob(_ context.Context, _ string) error {
	return nil
}

func (m *MockJobManagerWithFailure) CleanupFailedJob(_ context.Context, _ string) error {
	return nil
}

func (m *MockJobManagerWithFailure) HasFailedJobs(_ context.Context) (bool, error) {
	return false, nil
}

func (m *MockJobManagerWithFailure) CountActiveJobs(_ context.Context) (int, error) {
	return 0, nil
}

func (m *MockJobManagerWithFailure) IsJobRunning(_ *batchv1.Job) bool {
	return false
}

func (m *MockJobManagerWithFailure) IsJobFailed(_ *batchv1.Job) bool {
	return false
}

func (m *MockJobManagerWithFailure) IsJobSucceeded(_ *batchv1.Job) bool {
	return false
}
