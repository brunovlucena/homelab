package handler

import (
	"context"
	"testing"
	"time"

	"knative-lambda-new/internal/observability"
	"knative-lambda-new/pkg/builds"

	batchv1 "k8s.io/api/batch/v1"
)

// MockJobManager implements JobManager for testing
type MockJobManager struct {
	createJobCalled bool
	createJobError  error
	createJobResult *builds.HandlerResponse
}

func (m *MockJobManager) CreateJob(ctx context.Context, jobName string, buildRequest *builds.BuildRequest) (*batchv1.Job, error) {
	m.createJobCalled = true
	if m.createJobError != nil {
		return nil, m.createJobError
	}
	return &batchv1.Job{}, nil
}

func (m *MockJobManager) GenerateJobName(thirdPartyID, parserID string) string {
	return "test-job-" + thirdPartyID + "-" + parserID
}

func (m *MockJobManager) FindExistingJob(ctx context.Context, thirdPartyID, parserID string) (*batchv1.Job, error) {
	return nil, nil
}

func (m *MockJobManager) GetJob(ctx context.Context, jobName string) (*batchv1.Job, error) {
	return nil, nil
}

func (m *MockJobManager) DeleteJob(ctx context.Context, jobName string) error {
	return nil
}

func (m *MockJobManager) CleanupFailedJob(ctx context.Context, jobName string) error {
	return nil
}

func (m *MockJobManager) HasFailedJobs(ctx context.Context) (bool, error) {
	return false, nil
}

func (m *MockJobManager) CountActiveJobs(ctx context.Context) (int, error) {
	return 0, nil
}

func (m *MockJobManager) IsJobRunning(job *batchv1.Job) bool {
	return false
}

func (m *MockJobManager) IsJobFailed(job *batchv1.Job) bool {
	return false
}

func (m *MockJobManager) IsJobSucceeded(job *batchv1.Job) bool {
	return false
}

func TestAsyncJobCreator(t *testing.T) {
	// Create mock observability
	obs, err := observability.New(observability.Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		LogLevel:       "info",
		MetricsEnabled: false,
		TracingEnabled: false,
	})
	if err != nil {
		t.Fatalf("Failed to create observability: %v", err)
	}

	// Create mock job manager
	mockJobManager := &MockJobManager{}

	// Create async job creator
	creator := NewAsyncJobCreator(AsyncJobCreatorConfig{
		WorkerCount:   2,
		MaxRetries:    2,
		RetryDelay:    10 * time.Millisecond,
		MaxQueueSize:  10,
		JobManager:    mockJobManager,
		Observability: obs,
	})

	// Test basic functionality
	t.Run("CreateJobAsync", func(t *testing.T) {
		buildRequest := &builds.BuildRequest{
			ThirdPartyID:  "test-third-party",
			ParserID:      "test-parser",
			CorrelationID: "test-correlation",
		}

		jobName, err := creator.CreateJobAsync(context.Background(), buildRequest)
		if err != nil {
			t.Fatalf("Failed to create job async: %v", err)
		}

		expectedJobName := "test-job-test-third-party-test-parser"
		if jobName != expectedJobName {
			t.Errorf("Expected job name %s, got %s", expectedJobName, jobName)
		}

		// Wait a bit for the job to be processed
		time.Sleep(100 * time.Millisecond)

		// Check if job creation was called
		if !mockJobManager.createJobCalled {
			t.Error("Job creation was not called")
		}
	})

	// Test statistics
	t.Run("GetStats", func(t *testing.T) {
		stats := creator.GetStats()

		if stats["worker_count"] != 2 {
			t.Errorf("Expected worker count 2, got %v", stats["worker_count"])
		}

		if stats["max_queue_size"] != 10 {
			t.Errorf("Expected max queue size 10, got %v", stats["max_queue_size"])
		}
	})

	// Test shutdown
	t.Run("Shutdown", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := creator.Shutdown(ctx)
		if err != nil {
			t.Fatalf("Failed to shutdown: %v", err)
		}
	})
}
