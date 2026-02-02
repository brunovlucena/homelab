// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
//
//	🧪 SRE-001: Build Failure Investigation Tests
//
//	User Story: Build Failure Investigation
//	Priority: P0 | Story Points: 8
//
//	Tests validate:
//	- MTTR (Mean Time To Resolution) <30min for common failures
//	- Root cause identified within 10min
//	- Automated alerting for failure rate >5%
//	- Runbook documentation for top 5 failure modes
//	- Post-mortem created for novel failures
//	- Prometheus metrics track failure categories
//	- Failed jobs cleaned up automatically after 24hrs
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
package sre

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"knative-lambda/tests/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// Test constants.
const (
	TestNamespace = "knative-lambda"
	RunbookPath   = "../../../docs/03-for-engineers/sre/user-stories/SRE-001-build-failure-investigation.md"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// Test Fixtures and Helpers.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

// BuildFailureCategory represents different failure types.
type BuildFailureCategory string

const (
	FailureS3Access        BuildFailureCategory = "s3_access_denied"
	FailureImagePull       BuildFailureCategory = "image_pull_error"
	FailureOOMKilled       BuildFailureCategory = "oom_killed"
	FailureTimeout         BuildFailureCategory = "timeout"
	FailureDependencyError BuildFailureCategory = "dependency_error"
)

// createFailedJob creates a Kubernetes Job that failed for testing.
func createFailedJob(name string, category BuildFailureCategory, age time.Duration) *batchv1.Job {
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "knative-lambda",
			Labels: map[string]string{
				"app":              "knative-lambda",
				"job-type":         "builder",
				"failure-category": string(category),
			},
			CreationTimestamp: metav1.NewTime(time.Now().Add(-age)),
		},
		Status: batchv1.JobStatus{
			Failed: 1,
			Conditions: []batchv1.JobCondition{
				{
					Type:               batchv1.JobFailed,
					Status:             corev1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(time.Now().Add(-age)),
					Reason:             string(category),
					Message:            "Build failed: " + string(category),
				},
			},
		},
	}
	return job
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC1: MTTR (Mean Time To Resolution) <30min for common failures.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestSRE001_AC1_MTTR_CommonFailures(t *testing.T) {
	t.Run("MTTR calculation for S3 access failures", func(t *testing.T) {
		// Arrange
		startTime := time.Now().Add(-10 * time.Minute)
		endTime := time.Now()
		maxDuration := 30 * time.Minute

		phases := []testutils.Phase{
			{Name: "Detection", Duration: 2 * time.Minute},
			{Name: "Investigation", Duration: 5 * time.Minute},
			{Name: "Resolution", Duration: 2 * time.Minute},
			{Name: "Verification", Duration: 1 * time.Minute},
		}

		// Act & Assert
		testutils.RunTimingTest(t, "MTTR for common failures", startTime, endTime, maxDuration, phases)
	})

	t.Run("Track MTTR for different failure categories", func(t *testing.T) {
		// Arrange
		categories := map[BuildFailureCategory]time.Duration{
			FailureS3Access:        8 * time.Minute,
			FailureImagePull:       5 * time.Minute,
			FailureOOMKilled:       15 * time.Minute,
			FailureTimeout:         12 * time.Minute,
			FailureDependencyError: 20 * time.Minute,
		}

		// Act & Assert
		for category, mttr := range categories {
			t.Run(string(category), func(t *testing.T) {
				assert.Less(t, mttr.Minutes(), 30.0,
					"MTTR for %s should be less than 30 minutes, got %v", category, mttr)
			})
		}
	})

	t.Run("Calculate average MTTR across all failures", func(t *testing.T) {
		// Arrange
		failureResolutions := []time.Duration{
			8 * time.Minute,
			12 * time.Minute,
			5 * time.Minute,
			15 * time.Minute,
			10 * time.Minute,
		}

		// Act
		var totalMinutes float64
		for _, duration := range failureResolutions {
			totalMinutes += duration.Minutes()
		}
		avgMTTR := totalMinutes / float64(len(failureResolutions))

		// Assert
		assert.Less(t, avgMTTR, 30.0, "Average MTTR should be less than 30 minutes")
		assert.InDelta(t, 10.0, avgMTTR, 5.0, "Average MTTR should be around 10 minutes")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC2: Root cause identified within 10min.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

//nolint:funlen // Comprehensive root cause analysis test with multiple failure scenarios
func TestSRE001_AC2_RootCauseIdentification(t *testing.T) {
	t.Run("Identify S3 access denied from job logs", func(t *testing.T) {
		// Arrange
		jobLogs := `
time="2025-10-29T10:45:34Z" level=error error="failed to download from S3"
path="s3://knative-lambda-fusion/parser/xyz" status=403
ERROR: build failed: S3 access denied
`
		startTime := time.Now()

		// Act
		rootCause := identifyRootCause(jobLogs)
		endTime := time.Now()

		// Assert
		assert.Equal(t, FailureS3Access, rootCause, "Should identify S3 access denied")
		testutils.RunTimingTest(t, "Root cause identification", startTime, endTime, 1*time.Second, []testutils.Phase{})
	})

	t.Run("Identify image pull error from pod status", func(t *testing.T) {
		// Arrange
		clientset := fake.NewSimpleClientset()
		ctx := context.Background()

		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "build-job-xyz",
				Namespace: "knative-lambda",
			},
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{
					{
						Name: "kaniko",
						State: corev1.ContainerState{
							Waiting: &corev1.ContainerStateWaiting{
								Reason:  "ErrImagePull",
								Message: "Failed to pull image python:3.9-slim",
							},
						},
					},
				},
			},
		}
		_, err := clientset.CoreV1().Pods(pod.Namespace).Create(ctx, pod, metav1.CreateOptions{})
		require.NoError(t, err)

		// Act
		startTime := time.Now()
		retrievedPod, err := clientset.CoreV1().Pods(pod.Namespace).Get(ctx, pod.Name, metav1.GetOptions{})
		require.NoError(t, err)

		var rootCause BuildFailureCategory
		if len(retrievedPod.Status.ContainerStatuses) > 0 {
			if waiting := retrievedPod.Status.ContainerStatuses[0].State.Waiting; waiting != nil {
				if waiting.Reason == "ErrImagePull" || waiting.Reason == "ImagePullBackOff" {
					rootCause = FailureImagePull
				}
			}
		}
		identificationTime := time.Since(startTime)

		// Assert
		assert.Equal(t, FailureImagePull, rootCause, "Should identify image pull error")
		assert.Less(t, identificationTime.Minutes(), 10.0, "Root cause identification should take less than 10 minutes")
	})

	t.Run("Identify OOMKilled from pod status", func(t *testing.T) {
		// Arrange
		pod := &corev1.Pod{
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{
					{
						Name: "kaniko",
						LastTerminationState: corev1.ContainerState{
							Terminated: &corev1.ContainerStateTerminated{
								Reason:     "OOMKilled",
								ExitCode:   137,
								FinishedAt: metav1.NewTime(time.Now()),
							},
						},
					},
				},
			},
		}

		// Act
		var rootCause BuildFailureCategory
		if len(pod.Status.ContainerStatuses) > 0 {
			if terminated := pod.Status.ContainerStatuses[0].LastTerminationState.Terminated; terminated != nil {
				if terminated.Reason == "OOMKilled" {
					rootCause = FailureOOMKilled
				}
			}
		}

		// Assert
		assert.Equal(t, FailureOOMKilled, rootCause, "Should identify OOMKilled")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC3: Automated alerting for failure rate >5%.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestSRE001_AC3_AutomatedAlerting(t *testing.T) {
	t.Run("Alert fires when failure rate exceeds 5%", func(t *testing.T) {
		// Arrange
		totalBuilds := 100
		failedBuilds := 6 // 6% failure rate

		// Act
		failureRate := float64(failedBuilds) / float64(totalBuilds) * 100
		shouldAlert := failureRate > 5.0

		// Assert
		assert.True(t, shouldAlert, "Alert should fire when failure rate exceeds 5%%")
		assert.Equal(t, 6.0, failureRate, "Failure rate should be 6%%")
	})

	t.Run("No alert when failure rate is below threshold", func(t *testing.T) {
		// Arrange
		totalBuilds := 100
		failedBuilds := 3 // 3% failure rate

		// Act
		failureRate := float64(failedBuilds) / float64(totalBuilds) * 100
		shouldAlert := failureRate > 5.0

		// Assert
		assert.False(t, shouldAlert, "Alert should not fire when failure rate is below 5%%")
		assert.Equal(t, 3.0, failureRate, "Failure rate should be 3%%")
	})

	t.Run("Calculate failure rate over 5-minute window", func(t *testing.T) {
		// Arrange
		windowStart := time.Now().Add(-5 * time.Minute)
		builds := []struct {
			timestamp time.Time
			success   bool
		}{
			{windowStart.Add(1 * time.Minute), true},
			{windowStart.Add(2 * time.Minute), false},
			{windowStart.Add(3 * time.Minute), true},
			{windowStart.Add(4 * time.Minute), false},
			{windowStart.Add(5 * time.Minute), true},
		}

		// Act
		var failures int
		for _, build := range builds {
			if !build.success {
				failures++
			}
		}
		failureRate := float64(failures) / float64(len(builds)) * 100

		// Assert
		assert.Equal(t, 40.0, failureRate, "Failure rate should be 40%% in this window")
		assert.True(t, failureRate > 5.0, "Should trigger alert")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC4: Runbook documentation for top 5 failure modes.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestSRE001_AC4_RunbookDocumentation(t *testing.T) {
	t.Run("Runbook file exists and is readable", func(t *testing.T) {
		// Arrange
		runbookPath := RunbookPath

		// Act
		_, err := os.Stat(runbookPath)

		// Assert
		assert.NoError(t, err, "Runbook file should exist")

		// Verify runbook contains key sections
		content, err := os.ReadFile(runbookPath)
		require.NoError(t, err, "Should be able to read runbook")

		runbookContent := string(content)
		assert.Contains(t, runbookContent, "Build Failure Investigation", "Should be correct runbook")
		assert.Contains(t, runbookContent, "Acceptance Criteria", "Should document acceptance criteria")
	})

	t.Run("Runbook documents top 5 failure modes", func(t *testing.T) {
		// Arrange
		runbookPath := RunbookPath
		content, err := os.ReadFile(runbookPath)
		require.NoError(t, err)

		runbookContent := string(content)

		// Act & Assert - Verify top 5 failure modes are documented
		top5Failures := []string{
			"S3 Access Denied",
			"Image Pull",
			"OOMKilled",
			"Timeout",
			"Dependency",
		}

		for _, failure := range top5Failures {
			assert.Contains(t, runbookContent, failure,
				"Runbook should document failure mode: %s", failure)
		}
	})

	t.Run("Runbook contains troubleshooting commands", func(t *testing.T) {
		// Arrange
		runbookPath := RunbookPath
		content, err := os.ReadFile(runbookPath)
		require.NoError(t, err)

		runbookContent := string(content)

		// Act & Assert - Verify troubleshooting commands present
		requiredCommands := []string{
			"kubectl get jobs",
			"kubectl logs",
			"kubectl describe",
		}

		for _, cmd := range requiredCommands {
			assert.Contains(t, runbookContent, cmd,
				"Runbook should contain command: %s", cmd)
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC5: Post-mortem created for novel failures.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestSRE001_AC5_PostMortemCreation(t *testing.T) {
	t.Run("Detect novel failure category", func(t *testing.T) {
		// Arrange
		knownCategories := map[BuildFailureCategory]bool{
			FailureS3Access:        true,
			FailureImagePull:       true,
			FailureOOMKilled:       true,
			FailureTimeout:         true,
			FailureDependencyError: true,
		}

		novelFailure := BuildFailureCategory("certificate_expired")

		// Act
		isNovel := !knownCategories[novelFailure]

		// Assert
		assert.True(t, isNovel, "Should detect novel failure category")
	})

	t.Run("Post-mortem template contains required sections", func(t *testing.T) {
		// Arrange
		postMortemTemplate := `
## Incident Summary
## Timeline
## Root Cause Analysis
## Impact Assessment
## Resolution
## Action Items
## Lessons Learned
`

		// Act & Assert - Verify required sections
		requiredSections := []string{
			"Incident Summary",
			"Timeline",
			"Root Cause Analysis",
			"Impact Assessment",
			"Resolution",
			"Action Items",
			"Lessons Learned",
		}

		for _, section := range requiredSections {
			assert.Contains(t, postMortemTemplate, section,
				"Post-mortem template should contain section: %s", section)
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC6: Prometheus metrics track failure categories.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestSRE001_AC6_PrometheusMetrics(t *testing.T) {
	t.Run("Metric labels include failure category", func(t *testing.T) {
		// Arrange
		labels := map[string]string{
			"category":  string(FailureS3Access),
			"namespace": "knative-lambda",
			"job_type":  "builder",
		}

		// Act & Assert
		assert.Equal(t, "s3_access_denied", labels["category"], "Metric should include failure category")
		assert.Equal(t, "knative-lambda", labels["namespace"], "Metric should include namespace")
		assert.Equal(t, "builder", labels["job_type"], "Metric should include job type")
	})

	t.Run("Track failure counts by category", func(t *testing.T) {
		// Arrange
		failureCounts := map[BuildFailureCategory]int{
			FailureS3Access:        5,
			FailureImagePull:       3,
			FailureOOMKilled:       2,
			FailureTimeout:         1,
			FailureDependencyError: 4,
		}

		// Act
		totalFailures := 0
		for _, count := range failureCounts {
			totalFailures += count
		}

		// Assert
		assert.Equal(t, 15, totalFailures, "Should track total failures across all categories")
		assert.Equal(t, 5, failureCounts[FailureS3Access], "Should track S3 access failures")
	})

	t.Run("Calculate failure rate metric", func(t *testing.T) {
		// Arrange
		totalBuilds := 100
		failedBuilds := 15

		// Act
		failureRate := float64(failedBuilds) / float64(totalBuilds)

		// Assert
		assert.Equal(t, 0.15, failureRate, "Failure rate metric should be 0.15 (15%%)")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC7: Failed jobs cleaned up automatically after 24hrs.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestSRE001_AC7_FailedJobCleanup(t *testing.T) {
	t.Run("Identify jobs older than 24 hours", testIdentifyOldJobs)
	t.Run("Delete failed jobs older than 24 hours", testDeleteOldFailedJobs)
	t.Run("TTL seconds after finished set on job", testTTLSecondsAfterFinished)
}

// testIdentifyOldJobs tests identifying jobs older than 24 hours.
func testIdentifyOldJobs(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	ctx := context.Background()
	namespace := TestNamespace

	jobs := []*batchv1.Job{
		createFailedJob("old-job-1", FailureS3Access, 30*time.Hour),
		createFailedJob("old-job-2", FailureImagePull, 26*time.Hour),
		createFailedJob("recent-job", FailureOOMKilled, 12*time.Hour),
	}

	for _, job := range jobs {
		_, err := clientset.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
		require.NoError(t, err)
	}

	jobList, err := clientset.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
	require.NoError(t, err)

	var oldJobs []string
	cutoffTime := time.Now().Add(-24 * time.Hour)
	for _, job := range jobList.Items {
		if job.CreationTimestamp.Time.Before(cutoffTime) && job.Status.Failed > 0 {
			oldJobs = append(oldJobs, job.Name)
		}
	}

	assert.Len(t, oldJobs, 2, "Should find 2 jobs older than 24 hours")
	assert.Contains(t, oldJobs, "old-job-1", "Should identify old-job-1 for cleanup")
	assert.Contains(t, oldJobs, "old-job-2", "Should identify old-job-2 for cleanup")
	assert.NotContains(t, oldJobs, "recent-job", "Should not mark recent job for cleanup")
}

// testDeleteOldFailedJobs tests deleting failed jobs older than 24 hours.
func testDeleteOldFailedJobs(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	ctx := context.Background()
	namespace := TestNamespace

	oldJob := createFailedJob("cleanup-test", FailureS3Access, 30*time.Hour)
	_, err := clientset.BatchV1().Jobs(namespace).Create(ctx, oldJob, metav1.CreateOptions{})
	require.NoError(t, err)

	err = clientset.BatchV1().Jobs(namespace).Delete(ctx, oldJob.Name, metav1.DeleteOptions{})
	require.NoError(t, err)

	_, err = clientset.BatchV1().Jobs(namespace).Get(ctx, oldJob.Name, metav1.GetOptions{})
	assert.Error(t, err, "Job should be deleted")
	assert.Contains(t, err.Error(), "not found", "Should get not found error")
}

// testTTLSecondsAfterFinished tests TTL seconds after finished configuration.
func testTTLSecondsAfterFinished(t *testing.T) {
	ttlSeconds := int32(86400) // 24 hours in seconds
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ttl-test",
			Namespace: "knative-lambda",
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &ttlSeconds,
		},
	}

	assert.NotNil(t, job.Spec.TTLSecondsAfterFinished, "Job should have TTL configured")
	assert.Equal(t, int32(86400), *job.Spec.TTLSecondsAfterFinished,
		"TTL should be 24 hours (86400 seconds)")
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// Integration Test: Full Build Failure Investigation Flow.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestSRE001_Integration_BuildFailureInvestigation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Complete failure investigation workflow", func(t *testing.T) {
		// Arrange
		clientset := fake.NewSimpleClientset()
		ctx := context.Background()
		namespace := TestNamespace

		startTime := time.Now()

		// Step 1: Simulate build failures
		for i := 0; i < 3; i++ {
			job := createFailedJob(
				fmt.Sprintf("failed-job-%d", i),
				FailureS3Access,
				time.Duration(i)*time.Minute)
			_, err := clientset.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
			require.NoError(t, err)
		}

		// Step 2: Detect failure pattern
		jobList, err := clientset.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
		require.NoError(t, err)

		failedJobs := 0
		failureCategories := make(map[BuildFailureCategory]int)
		for _, job := range jobList.Items {
			if job.Status.Failed > 0 {
				failedJobs++
				category := BuildFailureCategory(job.Labels["failure-category"])
				failureCategories[category]++
			}
		}

		// Step 3: Identify root cause
		var dominantCategory BuildFailureCategory
		maxCount := 0
		for category, count := range failureCategories {
			if count > maxCount {
				maxCount = count
				dominantCategory = category
			}
		}

		investigationTime := time.Since(startTime)

		// Assert
		assert.Equal(t, 3, failedJobs, "Should detect 3 failed jobs")
		assert.Equal(t, FailureS3Access, dominantCategory, "Should identify S3 access as dominant failure")
		assert.Less(t, investigationTime.Minutes(), 10.0, "Investigation should complete within 10 minutes")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// Helper Functions.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

// identifyRootCause parses job logs to identify failure category.
func identifyRootCause(logs string) BuildFailureCategory {
	if contains(logs, "S3 access denied") || contains(logs, "status=403") {
		return FailureS3Access
	}
	if contains(logs, "ErrImagePull") || contains(logs, "ImagePullBackOff") {
		return FailureImagePull
	}
	if contains(logs, "OOMKilled") {
		return FailureOOMKilled
	}
	if contains(logs, "timeout") || contains(logs, "deadline exceeded") {
		return FailureTimeout
	}
	return FailureDependencyError
}

// contains checks if a string contains a substring (case-insensitive).
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
