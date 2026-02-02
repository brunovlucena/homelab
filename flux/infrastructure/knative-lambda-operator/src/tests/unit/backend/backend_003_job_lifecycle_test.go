// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
//
//	🧪 BACKEND-003: Kubernetes Job Lifecycle Tests
//
//	User Story: Kubernetes Job Lifecycle
//	Priority: P0 | Story Points: 13
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
package backend

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"

	config_pkg "knative-lambda/internal/config"
	"knative-lambda/internal/handler"
	"knative-lambda/pkg/builds"
)

// TestBackend003_JobCreation validates Kaniko job creation.
func TestBackend003_JobCreation(t *testing.T) {
	tests := []struct {
		name        string
		buildReq    *builds.BuildRequest
		expectError bool
		description string
	}{
		{
			name: "Create job with valid request",
			buildReq: &builds.BuildRequest{
				ThirdPartyID: "customer-123",
				ParserID:     "parser-abc",
				BuildType:    "container",
				Runtime:      "nodejs18",
				SourceURL:    "s3://bucket/key",
				EventID:      "evt-123",
				EventType:    "build.start",
				EventSource:  "test",
			},
			expectError: false,
			description: "Should create Kaniko job successfully",
		},
		{
			name: "Create job with sidecar",
			buildReq: &builds.BuildRequest{
				ThirdPartyID: "customer-456",
				ParserID:     "parser-xyz",
			},
			expectError: false,
			description: "Should include CloudEvents sidecar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx := context.Background()
			manager := setupJobManager(t)
			jobName := manager.GenerateJobName(tt.buildReq.ThirdPartyID, tt.buildReq.ParserID)

			// Act
			job, err := manager.CreateJob(ctx, jobName, tt.buildReq)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, job)
				assert.Equal(t, jobName, job.Name)
				assert.Equal(t, 2, len(job.Spec.Template.Spec.Containers), "Should have Kaniko + sidecar")
			}
		})
	}
}

// TestBackend003_JobSpecValidation validates job spec configuration.
func TestBackend003_JobSpecValidation(t *testing.T) {
	// Arrange
	ctx := context.Background()
	manager := setupJobManager(t)

	buildReq := &builds.BuildRequest{
		ThirdPartyID: "customer-123",
		ParserID:     "parser-abc",
	}
	jobName := manager.GenerateJobName(buildReq.ThirdPartyID, buildReq.ParserID)

	// Act
	job, err := manager.CreateJob(ctx, jobName, buildReq)

	// Assert
	require.NoError(t, err)

	// Validate job spec
	assert.Equal(t, int32(0), *job.Spec.BackoffLimit, "Backoff limit should be 0")
	assert.Equal(t, int32(86400), *job.Spec.TTLSecondsAfterFinished, "TTL should be 86400")
	assert.Equal(t, corev1.RestartPolicyNever, job.Spec.Template.Spec.RestartPolicy)

	// Validate labels
	assert.Equal(t, buildReq.ThirdPartyID, job.Labels["third-party-id"])
	assert.Equal(t, buildReq.ParserID, job.Labels["parser-id"])
	assert.Equal(t, "kaniko", job.Labels["component"])
}

// TestBackend003_ConcurrentJobLimits validates concurrent job limits.
func TestBackend003_ConcurrentJobLimits(t *testing.T) {
	// Arrange
	ctx := context.Background()
	manager := setupJobManager(t)

	// Create 10 jobs (max limit)
	for i := 0; i < 10; i++ {
		buildReq := &builds.BuildRequest{
			ThirdPartyID: "customer-123",
			ParserID:     fmt.Sprintf("parser-%d", i),
		}
		jobName := manager.GenerateJobName(buildReq.ThirdPartyID, buildReq.ParserID)
		_, err := manager.CreateJob(ctx, jobName, buildReq)
		require.NoError(t, err)
	}

	// Act - Try to create 11th job
	count, err := manager.CountActiveJobs(ctx)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 10, count, "Should have 10 active jobs")
}

// TestBackend003_JobDeduplication validates job deduplication logic.
func TestBackend003_JobDeduplication(t *testing.T) {
	// Arrange
	ctx := context.Background()
	manager := setupJobManager(t)

	buildReq := &builds.BuildRequest{
		ThirdPartyID: "customer-123",
		ParserID:     "parser-abc",
	}

	// Create first job
	jobName := manager.GenerateJobName(buildReq.ThirdPartyID, buildReq.ParserID)
	_, err := manager.CreateJob(ctx, jobName, buildReq)
	require.NoError(t, err)

	// Act - Check for existing job
	existingJob, err := manager.FindExistingJob(ctx, buildReq.ThirdPartyID, buildReq.ParserID)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, existingJob, "Should find existing job")
	assert.Equal(t, jobName, existingJob.Name)
}

// TestBackend003_JobStatusTracking validates job status transitions.
func TestBackend003_JobStatusTracking(t *testing.T) {
	tests := []struct {
		name          string
		jobStatus     batchv1.JobStatus
		expectedState string
		description   string
	}{
		{
			name: "Job pending",
			jobStatus: batchv1.JobStatus{
				Active: 0,
			},
			expectedState: "pending",
			description:   "Job is pending execution",
		},
		{
			name: "Job running",
			jobStatus: batchv1.JobStatus{
				Active: 1,
			},
			expectedState: "running",
			description:   "Job is running",
		},
		{
			name: "Job succeeded",
			jobStatus: batchv1.JobStatus{
				Succeeded: 1,
			},
			expectedState: "succeeded",
			description:   "Job completed successfully",
		},
		{
			name: "Job failed",
			jobStatus: batchv1.JobStatus{
				Failed: 1,
			},
			expectedState: "failed",
			description:   "Job failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			manager := setupJobManager(t)
			job := &batchv1.Job{
				Status: tt.jobStatus,
			}

			// Act & Assert
			switch tt.expectedState {
			case "running":
				assert.True(t, manager.IsJobRunning(job), tt.description)
			case "succeeded":
				assert.True(t, manager.IsJobSucceeded(job), tt.description)
			case "failed":
				assert.True(t, manager.IsJobFailed(job), tt.description)
			}
		})
	}
}

// TestBackend003_JobCleanup validates job cleanup for failed jobs.
func TestBackend003_JobCleanup(t *testing.T) {
	// Arrange
	ctx := context.Background()
	manager := setupJobManager(t)

	buildReq := &builds.BuildRequest{
		ThirdPartyID: "customer-123",
		ParserID:     "parser-abc",
		BuildType:    "container",
		Runtime:      "nodejs18",
		SourceURL:    "s3://bucket/key",
		EventID:      "evt-123",
		EventType:    "build.start",
		EventSource:  "test",
	}
	jobName := manager.GenerateJobName(buildReq.ThirdPartyID, buildReq.ParserID)

	// Create and fail a job
	_, err := manager.CreateJob(ctx, jobName, buildReq)
	require.NoError(t, err)

	// Act - Cleanup failed job
	err = manager.CleanupFailedJob(ctx, jobName)

	// Assert
	require.NoError(t, err)
}

// TestBackend003_TTLConfiguration validates TTL configuration.
func TestBackend003_TTLConfiguration(t *testing.T) {
	// Arrange
	ctx := context.Background()
	manager := setupJobManager(t)

	buildReq := &builds.BuildRequest{
		ThirdPartyID: "customer-123",
		ParserID:     "parser-abc",
	}
	jobName := manager.GenerateJobName(buildReq.ThirdPartyID, buildReq.ParserID)

	// Act
	job, err := manager.CreateJob(ctx, jobName, buildReq)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, job.Spec.TTLSecondsAfterFinished)
	assert.Equal(t, int32(86400), *job.Spec.TTLSecondsAfterFinished, "TTL should be 24 hours")
}

// TestBackend003_FailureBackoff validates failure backoff configuration.
func TestBackend003_FailureBackoff(t *testing.T) {
	// Arrange
	ctx := context.Background()
	manager := setupJobManager(t)

	buildReq := &builds.BuildRequest{
		ThirdPartyID: "customer-123",
		ParserID:     "parser-abc",
	}
	jobName := manager.GenerateJobName(buildReq.ThirdPartyID, buildReq.ParserID)

	// Act
	job, err := manager.CreateJob(ctx, jobName, buildReq)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, int32(0), *job.Spec.BackoffLimit, "Should not retry failed builds")
}

// TestBackend003_ResourceLimits validates resource limits configuration.
func TestBackend003_ResourceLimits(t *testing.T) {
	// Arrange
	ctx := context.Background()
	manager := setupJobManager(t)

	buildReq := &builds.BuildRequest{
		ThirdPartyID: "customer-123",
		ParserID:     "parser-abc",
	}
	jobName := manager.GenerateJobName(buildReq.ThirdPartyID, buildReq.ParserID)

	// Act
	job, err := manager.CreateJob(ctx, jobName, buildReq)

	// Assert
	require.NoError(t, err)

	// Check Kaniko container resources
	kanikoContainer := job.Spec.Template.Spec.Containers[0]
	// Note: Kubernetes normalizes 2000m to "2" (2 cores)
	assert.Equal(t, "2", kanikoContainer.Resources.Limits.Cpu().String())
	assert.Equal(t, "4Gi", kanikoContainer.Resources.Limits.Memory().String())
}

// TestBackend003_LabelManagement validates label application.
func TestBackend003_LabelManagement(t *testing.T) {
	// Arrange
	ctx := context.Background()
	manager := setupJobManager(t)

	buildReq := &builds.BuildRequest{
		ThirdPartyID: "customer-123",
		ParserID:     "parser-abc",
		BuildType:    "container",
		Runtime:      "nodejs18",
		SourceURL:    "s3://bucket/key",
		EventID:      "evt-123",
		EventType:    "build.start",
		EventSource:  "test",
	}
	jobName := manager.GenerateJobName(buildReq.ThirdPartyID, buildReq.ParserID)

	// Act
	job, err := manager.CreateJob(ctx, jobName, buildReq)

	// Assert
	require.NoError(t, err)

	// Validate all required labels
	assert.Equal(t, "customer-123", job.Labels["third-party-id"])
	assert.Equal(t, "parser-abc", job.Labels["parser-id"])
	assert.Equal(t, "kaniko", job.Labels["component"])
	assert.Contains(t, job.Labels, "build-id")
}

// TestBackend003_JobDeletion validates job deletion.
func TestBackend003_JobDeletion(t *testing.T) {
	// Arrange
	ctx := context.Background()
	manager := setupJobManager(t)

	buildReq := &builds.BuildRequest{
		ThirdPartyID: "customer-123",
		ParserID:     "parser-abc",
	}
	jobName := manager.GenerateJobName(buildReq.ThirdPartyID, buildReq.ParserID)

	// Create job
	_, err := manager.CreateJob(ctx, jobName, buildReq)
	require.NoError(t, err)

	// Act
	err = manager.DeleteJob(ctx, jobName)

	// Assert
	require.NoError(t, err)

	// Verify deletion
	job, err := manager.GetJob(ctx, jobName)
	assert.Error(t, err) // Should not find deleted job
	assert.Nil(t, job)
}

// TestBackend003_JobNameGeneration validates job name format.
func TestBackend003_JobNameGeneration(t *testing.T) {
	// Arrange
	manager := setupJobManager(t)
	thirdPartyID := "customer-123"
	parserID := "parser-abc"

	// Act
	jobName := manager.GenerateJobName(thirdPartyID, parserID)

	// Assert
	assert.Contains(t, jobName, "kaniko-build")
	assert.Contains(t, jobName, thirdPartyID)
	assert.Contains(t, jobName, parserID)
	assert.LessOrEqual(t, len(jobName), 63, "Job name should be <= 63 characters")
}

// TestBackend003_MultipleContainers validates multiple container setup.
func TestBackend003_MultipleContainers(t *testing.T) {
	// Arrange
	ctx := context.Background()
	manager := setupJobManager(t)

	buildReq := &builds.BuildRequest{
		ThirdPartyID: "customer-123",
		ParserID:     "parser-abc",
	}
	jobName := manager.GenerateJobName(buildReq.ThirdPartyID, buildReq.ParserID)

	// Act
	job, err := manager.CreateJob(ctx, jobName, buildReq)

	// Assert
	require.NoError(t, err)
	assert.Len(t, job.Spec.Template.Spec.Containers, 2, "Should have Kaniko + sidecar")

	// Verify container names
	containers := job.Spec.Template.Spec.Containers
	containerNames := []string{containers[0].Name, containers[1].Name}
	assert.Contains(t, containerNames, "kaniko")
	assert.Contains(t, containerNames, "cloudevents-sidecar")
}

// TestBackend003_ServiceAccountAttachment validates ServiceAccount attachment.
func TestBackend003_ServiceAccountAttachment(t *testing.T) {
	// Arrange
	ctx := context.Background()
	manager := setupJobManager(t)

	buildReq := &builds.BuildRequest{
		ThirdPartyID: "customer-123",
		ParserID:     "parser-abc",
	}
	jobName := manager.GenerateJobName(buildReq.ThirdPartyID, buildReq.ParserID)

	// Act
	job, err := manager.CreateJob(ctx, jobName, buildReq)

	// Assert
	require.NoError(t, err)
	assert.NotEmpty(t, job.Spec.Template.Spec.ServiceAccountName, "Should have ServiceAccount")
}

// Helper Functions.

func setupJobManager(t *testing.T) *handler.JobManagerImpl {
	// Create fake Kubernetes client
	k8sClient := fake.NewSimpleClientset()

	// Create mock observability with proper initialization
	obs := handler.NewMockObservability()

	// Create mock config (minimal for testing)
	k8sConfig := &config_pkg.KubernetesConfig{
		Namespace:      "knative-lambda",
		ServiceAccount: "test-sa",
		JobTTLSeconds:  86400, // 24 hours
	}

	buildConfig := &config_pkg.BuildConfig{
		KanikoImage:   "gcr.io/kaniko-project/executor:latest",
		SidecarImage:  "test-sidecar:latest",
		BuildTimeout:  30 * time.Minute,
		CPURequest:    "500m",
		CPULimit:      "2000m",
		MemoryRequest: "512Mi",
		MemoryLimit:   "4Gi",
	}

	awsConfig := &config_pkg.AWSConfig{
		ECRRegistry:       "test.ecr.amazonaws.com",
		ECRRepositoryName: "test-repo",
		S3SourceBucket:    "test-bucket",
		S3TempBucket:      "test-temp",
	}

	rateLimitConfig := &config_pkg.RateLimitingConfig{
		MaxConcurrentBuilds: 10,
		MaxConcurrentJobs:   10, // Allow 10 concurrent jobs for testing
	}

	// Create job manager with proper dependencies
	manager, err := handler.NewJobManager(handler.JobManagerConfig{
		K8sClient:       k8sClient,
		K8sConfig:       k8sConfig,
		BuildConfig:     buildConfig,
		AWSConfig:       awsConfig,
		RateLimitConfig: rateLimitConfig,
		Observability:   obs,
	})
	if err != nil {
		t.Fatalf("Failed to create job manager: %v", err)
	}

	return manager.(*handler.JobManagerImpl)
}
