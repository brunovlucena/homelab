package monitor

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"knative-lambda-new/internal/constants"
	"knative-lambda-new/pkg/builds"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// BuildMonitor monitors Kaniko container and publishes events to Knative broker
type BuildMonitor struct {
	k8sClient kubernetes.Interface
	config    BuildMonitorConfig
	logger    *slog.Logger
}

// BuildMonitorConfig configuration for build monitor
type BuildMonitorConfig struct {
	KanikoNamespace     string
	KanikoPodName       string
	KanikoContainerName string
	PollInterval        time.Duration
	BuildTimeout        time.Duration
	JobName             string
	ImageURI            string
	ThirdPartyID        string
	ParserID            string
	ContentHash         string // New: content hash for unique image tagging
	CorrelationID       string
	BrokerURL           string
}

// Using shared BuildCompletionEventData struct instead of local BuildData
// This ensures compatibility with the event handler

// NewBuildMonitor creates a new build monitor that sends events to Knative broker
func NewBuildMonitor(config BuildMonitorConfig) (*BuildMonitor, error) {
	logger := slog.Default()

	// Create Kubernetes client
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf(constants.ErrK8sConfigFailed+": %w", err)
	}

	k8sClient, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrK8sClientFailed+": %w", err)
	}

	return &BuildMonitor{
		k8sClient: k8sClient,
		config:    config,
		logger:    logger,
	}, nil
}

// MonitorBuild monitors the Kaniko build process
func (bm *BuildMonitor) MonitorBuild(ctx context.Context) error {
	bm.logger.Info("Starting build monitoring",
		"namespace", bm.config.KanikoNamespace,
		"pod", bm.config.KanikoPodName,
		"container", bm.config.KanikoContainerName,
		"job_name", bm.config.JobName)

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, bm.config.BuildTimeout)
	defer cancel()

	// Monitor build with timeout
	ticker := time.NewTicker(bm.config.PollInterval)
	defer ticker.Stop()

	startTime := time.Now()

	for {
		select {
		case <-timeoutCtx.Done():
			bm.logger.Error("Build timed out", "duration", time.Since(startTime))
			return bm.publishTimeoutEvent(ctx, startTime)

		case <-ticker.C:
			// First, check job status to detect job-level failures
			jobFailed, jobErr := bm.checkJobFailure(timeoutCtx)
			if jobErr != nil {
				bm.logger.Error("Failed to check job status", "error", jobErr)
			} else if jobFailed {
				bm.logger.Error("Job failed with BackoffLimitExceeded", "job_name", bm.config.JobName)
				return bm.publishJobFailureEvent(ctx, startTime)
			}

			// Check container status
			status, err := bm.getContainerStatus(timeoutCtx)
			if err != nil {
				bm.logger.Error("Failed to get container status", "error", err)
				continue
			}

			// Check if build is complete
			if status.State.Running != nil {
				bm.logger.Debug("Build still running")
				continue
			}

			// Build completed - check if success or failure
			if status.State.Terminated != nil {
				return bm.handleBuildCompletion(ctx, status.State.Terminated, startTime)
			}

			// Container is waiting - check for errors
			if status.State.Waiting != nil {
				bm.logger.Warn("Container waiting", "reason", status.State.Waiting.Reason)
				continue
			}
		}
	}
}

// getContainerStatus gets the status of the Kaniko container
func (bm *BuildMonitor) getContainerStatus(ctx context.Context) (v1.ContainerStatus, error) {
	pod, err := bm.k8sClient.CoreV1().Pods(bm.config.KanikoNamespace).Get(
		ctx,
		bm.config.KanikoPodName,
		metav1.GetOptions{},
	)
	if err != nil {
		return v1.ContainerStatus{}, fmt.Errorf(constants.ErrFailedToGetService, "pod", err)
	}

	// Find the Kaniko container
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Name == bm.config.KanikoContainerName {
			return containerStatus, nil
		}
	}

	return v1.ContainerStatus{}, fmt.Errorf(constants.ErrKanikoContainerNotFound)
}

// checkJobFailure checks if the job has failed with BackoffLimitExceeded
func (bm *BuildMonitor) checkJobFailure(ctx context.Context) (bool, error) {
	job, err := bm.k8sClient.BatchV1().Jobs(bm.config.KanikoNamespace).Get(
		ctx,
		bm.config.JobName,
		metav1.GetOptions{},
	)
	if err != nil {
		return false, fmt.Errorf("failed to get job: %w", err)
	}

	// Check if job has failed
	if job.Status.Failed > 0 {
		bm.logger.Info("Job has failed pods",
			"job_name", bm.config.JobName,
			"failed_pods", job.Status.Failed,
			"active_pods", job.Status.Active,
			"succeeded_pods", job.Status.Succeeded)

		// Check if job has reached backoff limit
		if job.Status.Failed >= *job.Spec.BackoffLimit {
			return true, nil
		}
	}

	return false, nil
}

// publishJobFailureEvent publishes a job failure event
func (bm *BuildMonitor) publishJobFailureEvent(ctx context.Context, startTime time.Time) error {
	buildData := builds.BuildCompletionEventData{
		ThirdPartyID:  bm.config.ThirdPartyID,
		ParserID:      bm.config.ParserID,
		ContentHash:   bm.config.ContentHash, // New: include content hash
		JobName:       bm.config.JobName,
		ImageURI:      bm.config.ImageURI,
		Status:        "failed",
		ErrorMessage:  "Job failed with BackoffLimitExceeded",
		ExitCode:      1,
		Duration:      time.Since(startTime),
		CorrelationID: bm.config.CorrelationID,
	}

	bm.logger.Error("Job failed with BackoffLimitExceeded",
		"job_name", bm.config.JobName,
		"duration", time.Since(startTime))

	return bm.publishBuildEvent(ctx, builds.EventTypeBuildFailed, buildData)
}

// handleBuildCompletion handles build completion (success or failure)
func (bm *BuildMonitor) handleBuildCompletion(ctx context.Context, terminated *v1.ContainerStateTerminated, startTime time.Time) error {
	duration := time.Since(startTime)

	buildData := builds.BuildCompletionEventData{
		ThirdPartyID:  bm.config.ThirdPartyID,
		ParserID:      bm.config.ParserID,
		ContentHash:   bm.config.ContentHash, // New: include content hash
		JobName:       bm.config.JobName,
		ImageURI:      bm.config.ImageURI,
		ExitCode:      int(terminated.ExitCode),
		Duration:      duration,
		CorrelationID: bm.config.CorrelationID,
	}

	if terminated.ExitCode == 0 {
		// Build succeeded
		buildData.Status = "success"
		bm.logger.Info("Build completed successfully",
			"duration", duration,
			"exit_code", terminated.ExitCode)

		return bm.publishBuildEvent(ctx, builds.EventTypeBuildComplete, buildData)
	} else {
		// Build failed
		buildData.Status = "failed"
		buildData.ErrorMessage = terminated.Message
		if buildData.ErrorMessage == "" {
			buildData.ErrorMessage = terminated.Reason
		}

		bm.logger.Error("Build failed",
			"duration", duration,
			"exit_code", terminated.ExitCode,
			"error", buildData.ErrorMessage)

		return bm.publishBuildEvent(ctx, builds.EventTypeBuildFailed, buildData)
	}
}

// publishTimeoutEvent publishes a timeout event
func (bm *BuildMonitor) publishTimeoutEvent(ctx context.Context, startTime time.Time) error {
	buildData := builds.BuildCompletionEventData{
		ThirdPartyID:  bm.config.ThirdPartyID,
		ParserID:      bm.config.ParserID,
		ContentHash:   bm.config.ContentHash, // New: include content hash
		JobName:       bm.config.JobName,
		ImageURI:      bm.config.ImageURI,
		Status:        "timeout",
		ErrorMessage:  fmt.Sprintf("Build timed out after %s", bm.config.BuildTimeout),
		Duration:      time.Since(startTime),
		CorrelationID: bm.config.CorrelationID,
	}

	return bm.publishBuildEvent(ctx, builds.EventTypeBuildTimeout, buildData)
}

// publishBuildEvent publishes a build event directly to the Knative broker
func (bm *BuildMonitor) publishBuildEvent(ctx context.Context, eventType string, buildData builds.BuildCompletionEventData) error {
	// Create a standard CloudEvent using the CloudEvents Go SDK
	event := cloudevents.NewEvent()
	event.SetSpecVersion(cloudevents.VersionV1)
	event.SetType(eventType)
	event.SetSource(builds.SourceBuilder)
	event.SetID(uuid.New().String())
	event.SetSubject(fmt.Sprintf("build/%s/%s", buildData.ThirdPartyID, buildData.ParserID))
	event.SetTime(time.Now())

	// Set the data as JSON
	if err := event.SetData(cloudevents.ApplicationJSON, buildData); err != nil {
		return fmt.Errorf(constants.ErrFailedToSetEventData+": %w", err)
	}

	// Add extensions
	event.SetExtension("correlationid", buildData.CorrelationID)
	event.SetExtension("thirdpartyid", buildData.ThirdPartyID)
	event.SetExtension("parserid", buildData.ParserID)
	event.SetExtension("jobname", buildData.JobName)

	// Create CloudEvents HTTP client
	client, err := cloudevents.NewClientHTTP()
	if err != nil {
		return fmt.Errorf(constants.ErrFailedToCreateCloudEventsClient+": %w", err)
	}

	// Send event to Knative broker
	brokerURL := bm.config.BrokerURL
	if brokerURL == "" {
		return fmt.Errorf(constants.ErrBrokerURLNotConfigured)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Create target context for the broker
	ceCtx := cloudevents.ContextWithTarget(ctx, brokerURL)

	// Send the event
	result := client.Send(ceCtx, event)
	if cloudevents.IsUndelivered(result) {
		return fmt.Errorf(constants.ErrFailedToSendEventToBroker+": %w", result)
	}

	bm.logger.Info("Published build event to Knative broker",
		"event_id", event.ID(),
		"event_type", eventType,
		"status", buildData.Status,
		"broker_url", brokerURL)

	return nil
}

// Close closes the build monitor
func (bm *BuildMonitor) Close() error {
	// No resources to close when using HTTP client
	return nil
}
