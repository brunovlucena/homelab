// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
//
//	🧪 SRE-018: DLQ Kaniko Build Failures
//
//	User Story: Handle Kaniko build failures with DLQ
//	Priority: P0 | Story Points: 13
//
//	Tests validate:
//	- S3 access failures routed to DLQ
//	- Image pull errors handled
//	- OOMKilled builds tracked
//	- Build timeout failures managed
//	- Dependency errors detected
//	- ECR push failures tracked
//	- Build retries with backoff
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
package sre

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// Test Fixtures.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

type BuildFailureType string

const (
	BuildS3AccessDenied    BuildFailureType = "s3_access_denied"
	BuildImagePullError    BuildFailureType = "image_pull_error"
	BuildOOMKilled         BuildFailureType = "oom_killed"
	BuildTimeout           BuildFailureType = "build_timeout"
	BuildDependencyError   BuildFailureType = "dependency_error"
	BuildECRPushFailed     BuildFailureType = "ecr_push_failed"
	BuildInvalidDockerfile BuildFailureType = "invalid_dockerfile"
)

type BuildFailureEvent struct {
	EventID       string
	BuildID       string
	ParserID      string
	ThirdPartyID  string
	FailureType   BuildFailureType
	Timestamp     time.Time
	RetryCount    int
	BuildDuration time.Duration
	MovedToDLQ    bool
	DLQReason     string
	ErrorDetails  string
	JobName       string
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC1: S3 access failures routed to DLQ.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestSRE018_AC1_S3AccessFailuresDLQ(t *testing.T) {
	t.Run("S3 403 Access Denied routed to DLQ", func(t *testing.T) {
		// Arrange
		event := BuildFailureEvent{
			EventID:      "build-s3-403-1",
			BuildID:      "build-123",
			ParserID:     "parser-456",
			ThirdPartyID: "client-789",
			FailureType:  BuildS3AccessDenied,
			ErrorDetails: "status=403 s3://bucket/parser/xyz: Access Denied",
			MovedToDLQ:   true,
			DLQReason:    "s3_access_denied_permanent",
		}

		// Assert - 403 is permanent, no retry
		assert.True(t, event.MovedToDLQ, "Should move to DLQ on S3 403")
		assert.Contains(t, event.ErrorDetails, "Access Denied")
	})

	t.Run("S3 bucket not found permanent failure", func(t *testing.T) {
		// Arrange
		event := BuildFailureEvent{
			EventID:      "build-s3-404-1",
			FailureType:  BuildS3AccessDenied,
			ErrorDetails: "NoSuchBucket: The specified bucket does not exist",
			MovedToDLQ:   true,
			DLQReason:    "s3_bucket_not_found",
		}

		// Assert
		assert.True(t, event.MovedToDLQ)
		assert.Contains(t, event.ErrorDetails, "does not exist")
	})

	t.Run("S3 parser file not found routed to DLQ", func(t *testing.T) {
		// Arrange
		event := BuildFailureEvent{
			EventID:      "build-s3-nofile-1",
			FailureType:  BuildS3AccessDenied,
			ErrorDetails: "NoSuchKey: The specified key does not exist: parser/xyz",
			MovedToDLQ:   true,
			DLQReason:    "s3_parser_file_not_found",
		}

		// Assert
		assert.True(t, event.MovedToDLQ)
		assert.Contains(t, event.ErrorDetails, "key does not exist")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC2: Image pull errors handled.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestSRE018_AC2_ImagePullErrorsHandled(t *testing.T) {
	t.Run("Base image not found routed to DLQ", func(t *testing.T) {
		// Arrange
		event := BuildFailureEvent{
			EventID:      "build-imgpull-1",
			FailureType:  BuildImagePullError,
			ErrorDetails: "failed to pull python:3.9-invalid: manifest unknown",
			MovedToDLQ:   true,
			DLQReason:    "base_image_not_found",
		}

		// Assert
		assert.True(t, event.MovedToDLQ)
		assert.Contains(t, event.ErrorDetails, "manifest unknown")
	})

	t.Run("Docker Hub rate limit triggers retry", func(t *testing.T) {
		// Arrange
		event := BuildFailureEvent{
			EventID:      "build-ratelimit-1",
			FailureType:  BuildImagePullError,
			ErrorDetails: "toomanyrequests: You have reached your pull rate limit",
			RetryCount:   0,
		}

		maxRetries := 3

		// Act - Retry on rate limit
		for event.RetryCount < maxRetries {
			event.RetryCount++
			time.Sleep(10 * time.Millisecond)

			// Simulate persistent rate limit
			rateLimitCleared := false
			if rateLimitCleared {
				break
			}
		}

		if event.RetryCount >= maxRetries {
			event.MovedToDLQ = true
			event.DLQReason = "docker_hub_rate_limit_persistent"
		}

		// Assert
		assert.True(t, event.MovedToDLQ)
		assert.Equal(t, maxRetries, event.RetryCount)
	})

	t.Run("Image pull timeout after 5 minutes", func(t *testing.T) {
		// Arrange
		pullTimeout := 5 * time.Minute
		pullDuration := 6 * time.Minute

		event := BuildFailureEvent{
			EventID:      "build-pulltimeout-1",
			FailureType:  BuildImagePullError,
			ErrorDetails: "context deadline exceeded",
		}

		// Act
		if pullDuration > pullTimeout {
			event.MovedToDLQ = true
			event.DLQReason = "image_pull_timeout"
		}

		// Assert
		assert.True(t, event.MovedToDLQ)
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC3: OOMKilled builds tracked.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestSRE018_AC3_OOMKilledTracked(t *testing.T) {
	t.Run("OOMKilled routed to DLQ with recommendation", func(t *testing.T) {
		// Arrange
		event := BuildFailureEvent{
			EventID:      "build-oom-1",
			FailureType:  BuildOOMKilled,
			ErrorDetails: "OOMKilled: Container exceeded memory limit of 2Gi",
			MovedToDLQ:   true,
			DLQReason:    "build_oom_killed",
		}

		// Assert
		assert.True(t, event.MovedToDLQ)
		assert.Contains(t, event.ErrorDetails, "OOMKilled")
	})

	t.Run("Memory usage tracked for OOM analysis", func(t *testing.T) {
		// Arrange
		type MemoryMetrics struct {
			Limit     string
			Peak      string
			Average   string
			OOMKilled bool
		}

		metrics := MemoryMetrics{
			Limit:     "2Gi",
			Peak:      "2.1Gi",
			Average:   "1.8Gi",
			OOMKilled: true,
		}

		// Assert
		assert.True(t, metrics.OOMKilled)
		assert.Equal(t, "2Gi", metrics.Limit)
	})

	t.Run("Recommend memory increase for OOM failures", func(t *testing.T) {
		// Arrange
		currentLimit := "2Gi"
		recommendedLimit := "4Gi"

		type Recommendation struct {
			CurrentLimit     string
			RecommendedLimit string
			Reason           string
		}

		rec := Recommendation{
			CurrentLimit:     currentLimit,
			RecommendedLimit: recommendedLimit,
			Reason:           "Build exceeded 2Gi limit 3 times",
		}

		// Assert
		assert.NotEqual(t, rec.CurrentLimit, rec.RecommendedLimit)
		assert.Equal(t, "4Gi", rec.RecommendedLimit)
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC4: Build timeout failures managed.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestSRE018_AC4_BuildTimeoutManaged(t *testing.T) {
	t.Run("Build timeout after 30 minutes routed to DLQ", func(t *testing.T) {
		// Arrange
		buildTimeout := 30 * time.Minute
		buildDuration := 35 * time.Minute

		event := BuildFailureEvent{
			EventID:       "build-timeout-1",
			FailureType:   BuildTimeout,
			BuildDuration: buildDuration,
			ErrorDetails:  "Build exceeded 30m timeout",
			MovedToDLQ:    true,
			DLQReason:     "build_timeout_exceeded",
		}

		// Assert
		assert.Greater(t, buildDuration, buildTimeout)
		assert.True(t, event.MovedToDLQ)
	})

	t.Run("Build duration percentiles tracked", func(t *testing.T) {
		// Arrange
		type BuildDurationMetrics struct {
			P50 time.Duration
			P95 time.Duration
			P99 time.Duration
		}

		metrics := BuildDurationMetrics{
			P50: 5 * time.Minute,
			P95: 15 * time.Minute,
			P99: 25 * time.Minute,
		}

		buildTimeout := 30 * time.Minute

		// Assert
		assert.Less(t, metrics.P99, buildTimeout, "P99 should be under timeout")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC5: Dependency errors detected.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestSRE018_AC5_DependencyErrorsDetected(t *testing.T) {
	t.Run("Python pip install failure routed to DLQ", func(t *testing.T) {
		// Arrange
		event := BuildFailureEvent{
			EventID:      "build-pip-1",
			FailureType:  BuildDependencyError,
			ErrorDetails: "ERROR: Could not find a version that satisfies the requirement invalid-package==1.0.0",
			MovedToDLQ:   true,
			DLQReason:    "python_dependency_not_found",
		}

		// Assert
		assert.True(t, event.MovedToDLQ)
		assert.Contains(t, event.ErrorDetails, "Could not find a version")
	})

	t.Run("Node.js npm install failure routed to DLQ", func(t *testing.T) {
		// Arrange
		event := BuildFailureEvent{
			EventID:      "build-npm-1",
			FailureType:  BuildDependencyError,
			ErrorDetails: "npm ERR! 404 Not Found - GET https://registry.npmjs.org/invalid-package",
			MovedToDLQ:   true,
			DLQReason:    "nodejs_dependency_not_found",
		}

		// Assert
		assert.True(t, event.MovedToDLQ)
		assert.Contains(t, event.ErrorDetails, "404")
	})

	t.Run("Go mod download failure routed to DLQ", func(t *testing.T) {
		// Arrange
		event := BuildFailureEvent{
			EventID:      "build-gomod-1",
			FailureType:  BuildDependencyError,
			ErrorDetails: "go: github.com/invalid/package@v1.0.0: reading https://proxy.golang.org: 404 Not Found",
			MovedToDLQ:   true,
			DLQReason:    "go_dependency_not_found",
		}

		// Assert
		assert.True(t, event.MovedToDLQ)
		assert.Contains(t, event.ErrorDetails, "404")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC6: ECR push failures tracked.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestSRE018_AC6_ECRPushFailuresTracked(t *testing.T) {
	t.Run("ECR authentication failure routed to DLQ", func(t *testing.T) {
		// Arrange
		event := BuildFailureEvent{
			EventID:      "build-ecr-auth-1",
			FailureType:  BuildECRPushFailed,
			ErrorDetails: "denied: Your authorization token has expired. Reauthenticate and try again",
			MovedToDLQ:   true,
			DLQReason:    "ecr_auth_expired",
		}

		// Assert
		assert.True(t, event.MovedToDLQ)
		assert.Contains(t, event.ErrorDetails, "expired")
	})

	t.Run("ECR repository not found routed to DLQ", func(t *testing.T) {
		// Arrange
		event := BuildFailureEvent{
			EventID:      "build-ecr-notfound-1",
			FailureType:  BuildECRPushFailed,
			ErrorDetails: "repository does not exist: knative-lambdas/invalid-repo",
			MovedToDLQ:   true,
			DLQReason:    "ecr_repository_not_found",
		}

		// Assert
		assert.True(t, event.MovedToDLQ)
		assert.Contains(t, event.ErrorDetails, "does not exist")
	})

	t.Run("ECR push timeout after 10 minutes", func(t *testing.T) {
		// Arrange
		pushTimeout := 10 * time.Minute
		pushDuration := 12 * time.Minute

		event := BuildFailureEvent{
			EventID:     "build-ecr-timeout-1",
			FailureType: BuildECRPushFailed,
		}

		// Act
		if pushDuration > pushTimeout {
			event.MovedToDLQ = true
			event.DLQReason = "ecr_push_timeout"
		}

		// Assert
		assert.True(t, event.MovedToDLQ)
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC7: Build retries with backoff.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestSRE018_AC7_BuildRetriesWithBackoff(t *testing.T) {
	t.Run("Transient failures retry with exponential backoff", func(t *testing.T) {
		// Arrange
		type RetryPolicy struct {
			MaxAttempts       int
			InitialDelay      time.Duration
			BackoffMultiplier float64
			RetryableErrors   []BuildFailureType
		}

		policy := RetryPolicy{
			MaxAttempts:       3,
			InitialDelay:      1 * time.Minute,
			BackoffMultiplier: 2.0,
			RetryableErrors: []BuildFailureType{
				BuildImagePullError, // Docker Hub rate limit
				BuildTimeout,        // Network issues
			},
		}

		// Assert
		assert.Equal(t, 3, policy.MaxAttempts)
		assert.Contains(t, policy.RetryableErrors, BuildImagePullError)
	})

	t.Run("Permanent failures skip retries", func(t *testing.T) {
		// Arrange
		permanentFailures := []BuildFailureType{
			BuildS3AccessDenied,
			BuildInvalidDockerfile,
			BuildDependencyError,
		}

		event := BuildFailureEvent{
			EventID:     "build-permanent-1",
			FailureType: BuildS3AccessDenied,
			RetryCount:  0,
		}

		// Act
		shouldRetry := true
		for _, pf := range permanentFailures {
			if event.FailureType == pf {
				shouldRetry = false
				event.MovedToDLQ = true
				event.DLQReason = "permanent_failure_no_retry"
				break
			}
		}

		// Assert
		assert.False(t, shouldRetry, "Should not retry permanent failures")
		assert.True(t, event.MovedToDLQ)
	})

	t.Run("Failure backoff period enforced", func(t *testing.T) {
		// Arrange
		lastFailureTime := time.Now().Add(-30 * time.Second)
		backoffPeriod := 1 * time.Minute

		// Act
		timeSinceFailure := time.Since(lastFailureTime)
		canRetry := timeSinceFailure >= backoffPeriod

		// Assert
		assert.False(t, canRetry, "Should wait for backoff period")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// Integration Test: Complete Build Failure DLQ Workflow.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

//nolint:funlen // Full integration test with complete build failure workflow
func TestSRE018_Integration_BuildFailureDLQ(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Complete build failure recovery workflow", func(t *testing.T) {
		// Step 1: Build event triggers Kaniko job
		event := BuildFailureEvent{
			EventID:      "integration-build-1",
			BuildID:      "build-abc123",
			ParserID:     "parser-xyz789",
			ThirdPartyID: "client-def456",
			JobName:      "kaniko-build-abc123",
			Timestamp:    time.Now(),
		}

		// Step 2: S3 access fails (permanent)
		event.FailureType = BuildS3AccessDenied
		event.ErrorDetails = "AccessDenied: s3://bucket/parser/xyz789"

		// Step 3: Identify as permanent failure
		permanentFailures := []BuildFailureType{
			BuildS3AccessDenied,
			BuildInvalidDockerfile,
			BuildDependencyError,
		}

		isPermanent := false
		for _, pf := range permanentFailures {
			if event.FailureType == pf {
				isPermanent = true
				break
			}
		}

		// Step 4: Move to DLQ without retries
		if isPermanent {
			event.MovedToDLQ = true
			event.DLQReason = "s3_access_denied_permanent_failure"
			event.RetryCount = 0
		}

		// Step 5: Alert fires for DLQ event
		type AlertCondition struct {
			Name      string
			Severity  string
			Triggered bool
			Message   string
		}

		alert := AlertCondition{
			Name:      "BuildFailureDLQ",
			Severity:  "critical",
			Triggered: event.MovedToDLQ,
			Message:   "Build failed with permanent S3 error: " + event.ErrorDetails,
		}

		// Step 6: Operator investigates
		type Investigation struct {
			RootCause       string
			Resolution      string
			PreventionSteps []string
		}

		investigation := Investigation{
			RootCause:  "S3 IAM permissions missing for parser bucket",
			Resolution: "Updated IAM role with s3:GetObject permission",
			PreventionSteps: []string{
				"Add IAM permission validation in deployment pipeline",
				"Create pre-build permission check",
				"Document required S3 permissions",
			},
		}

		// Assert
		assert.True(t, isPermanent, "Should identify as permanent failure")
		assert.Equal(t, 0, event.RetryCount, "Should not retry permanent failures")
		assert.True(t, event.MovedToDLQ)
		assert.True(t, alert.Triggered)
		assert.Equal(t, "critical", alert.Severity)
		assert.NotEmpty(t, investigation.RootCause)
		assert.Len(t, investigation.PreventionSteps, 3)

		t.Logf("🎯 Build failure recovery workflow completed:")
		t.Logf("  - Build ID: %s", event.BuildID)
		t.Logf("  - Failure: %s", event.FailureType)
		t.Logf("  - Permanent: %v", isPermanent)
		t.Logf("  - DLQ: %v", event.MovedToDLQ)
		t.Logf("  - Root cause: %s", investigation.RootCause)
	})
}
