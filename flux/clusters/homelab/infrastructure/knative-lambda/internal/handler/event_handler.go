// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🎯 EVENT HANDLER - Composed handler with focused components
//
//	🎯 Purpose: Orchestrate focused components for CloudEvent processing
//	💡 Features: Component composition, dependency injection, loose coupling
//
//	🏛️ ARCHITECTURE:
//	🎯 Component Orchestration - Coordinate focused components
//	🔗 Dependency Injection - Inject dependencies into components
//	🔄 Event Flow - Route events through appropriate components
//	📊 Response Handling - Aggregate responses from components
//
//	🔄 EVENT FLOW:
//	1. CloudEvent (build.start) → Parse request → Create S3 context → Create Kaniko job
//	2. CloudEvent (build.complete) → Parse completion → Create Knative service + trigger
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	batchv1 "k8s.io/api/batch/v1"

	"knative-lambda-new/internal/config"
	"knative-lambda-new/internal/constants"
	"knative-lambda-new/internal/errors"
	"knative-lambda-new/internal/observability"
	"knative-lambda-new/pkg/builds"
)

// 🎯 EventHandlerImpl - "Composed handler with focused components"
type EventHandlerImpl struct {
	// 🎯 Dependency Injection Container - Centralized component management
	container ComponentContainer

	// 🔧 Shared Dependencies
	config *config.Config
	obs    *observability.Observability
}

// 🎯 EventHandlerConfig - "Configuration for creating event handler"
type EventHandlerConfig struct {
	Container ComponentContainer
}

// 🏗️ NewEventHandler - "Create new event handler with composed components"
func NewEventHandler(config EventHandlerConfig) (*EventHandlerImpl, error) {
	if config.Container == nil {
		return nil, errors.NewConfigurationError("event_handler", "container", "container cannot be nil")
	}

	// Validate all components are initialized
	if err := config.Container.(*ComponentContainerImpl).ValidateComponents(); err != nil {
		return nil, errors.NewConfigurationError("event_handler", "container", fmt.Sprintf("container validation failed: %v", err))
	}

	return &EventHandlerImpl{
		container: config.Container,
		config:    config.Container.GetConfig(),
		obs:       config.Container.GetObservability(),
	}, nil
}

// 📥 ProcessCloudEvent - "Process CloudEvent with comprehensive observability"
func (h *EventHandlerImpl) ProcessCloudEvent(ctx context.Context, event *cloudevents.Event) (*builds.HandlerResponse, error) {
	// Create metrics recorder for this request
	metricsRec := observability.NewMetricsRecorder(h.obs)

	// Start distributed tracing span
	ctx, span := h.obs.StartSpanWithAttributes(ctx, "process_cloud_event", map[string]string{
		"event.type":   event.Type(),
		"event.source": event.Source(),
		"event.id":     event.ID(),
	})
	defer span.End()

	// Validate event using internal validation
	if err := h.ValidateEvent(ctx, event); err != nil {
		metricsRec.RecordError(ctx, "event_handler", "validation_error", "error")
		return nil, err
	}

	// Record build request metric for start events
	if h.isBuildStartEvent(event) {
		if buildRequest, err := h.ParseBuildRequest(ctx, event); err == nil {
			metricsRec.RecordBuildRequest(ctx, buildRequest.ThirdPartyID, buildRequest.ParserID, "received")
		}
	}

	// Process event with comprehensive tracing
	response, err := h.processEventWithTracing(ctx, event, metricsRec)
	if err != nil {
		metricsRec.RecordError(ctx, "event_handler", "processing_error", "error")
		return nil, err
	}

	return response, nil
}

// processBuildStartEvent processes build start events
func (h *EventHandlerImpl) processBuildStartEvent(ctx context.Context, event *cloudevents.Event) (*builds.HandlerResponse, error) {
	// Create a comprehensive span for the entire build start process
	ctx, span := h.obs.StartSpanWithAttributes(ctx, "process_build_start_event", map[string]string{
		"event.type":    event.Type(),
		"event.source":  event.Source(),
		"event.id":      event.ID(),
		"event.subject": event.Subject(),
	})
	defer span.End()

	// Parse build request with tracing using internal parsing
	ctx, parseSpan := h.obs.StartSpan(ctx, "parse_build_request")
	buildRequest, err := h.ParseBuildRequest(ctx, event)
	if err != nil {
		parseSpan.End()
		h.obs.Error(ctx, err, "Failed to parse build request data")
		return nil, errors.NewValidationError("build_request_data", nil, fmt.Sprintf("failed to parse build request data: %v", err))
	}
	parseSpan.End()

	// Add build request details to the main span
	span.SetAttributes(
		attribute.String("build.third_party_id", buildRequest.ThirdPartyID),
		attribute.String("build.parser_id", buildRequest.ParserID),
		attribute.String("build.correlation_id", buildRequest.CorrelationID),
	)

	h.obs.Info(ctx, "Processing build start event",
		"third_party_id", buildRequest.ThirdPartyID,
		"parser_id", buildRequest.ParserID,
		"correlation_id", buildRequest.CorrelationID)

	// Create build context with tracing
	ctx, contextSpan := h.obs.StartSpan(ctx, "create_build_context")
	buildContextKey, err := h.createBuildContext(ctx, buildRequest)
	if err != nil {
		contextSpan.End()
		h.obs.Error(ctx, err, "Failed to create build context",
			"third_party_id", buildRequest.ThirdPartyID,
			"parser_id", buildRequest.ParserID,
			"correlation_id", buildRequest.CorrelationID)
		return nil, err
	}
	contextSpan.SetAttributes(attribute.String("build.context_key", buildContextKey))
	contextSpan.End()

	// Create Kaniko job asynchronously using worker pool
	ctx, jobSpan := h.obs.StartSpan(ctx, "create_kaniko_job_async")
	h.obs.Info(ctx, "Creating Kaniko job asynchronously",
		"third_party_id", buildRequest.ThirdPartyID,
		"parser_id", buildRequest.ParserID,
		"correlation_id", buildRequest.CorrelationID,
		"build_context_key", buildContextKey)

	// Get async job creator from container
	asyncJobCreator := h.container.GetAsyncJobCreator()
	if asyncJobCreator == nil {
		jobSpan.End()
		h.obs.Error(ctx, fmt.Errorf("async job creator not available"), "Async job creator not available")
		return nil, fmt.Errorf("async job creator not available")
	}

	// Queue job creation asynchronously
	jobName, err := asyncJobCreator.CreateJobAsync(ctx, buildRequest)
	if err != nil {
		jobSpan.End()
		h.obs.Error(ctx, err, "Failed to queue job creation",
			"third_party_id", buildRequest.ThirdPartyID,
			"parser_id", buildRequest.ParserID,
			"correlation_id", buildRequest.CorrelationID,
			"error_details", err.Error())
		return nil, err
	}

	// Return immediate response - job creation is now handled asynchronously
	response := &builds.HandlerResponse{
		Status:        "queued",
		Message:       "Build job creation queued successfully",
		JobName:       jobName,
		CorrelationID: buildRequest.CorrelationID,
	}

	jobSpan.SetAttributes(
		attribute.String("job.name", jobName),
		attribute.String("job.status", "queued"),
		attribute.String("processing.mode", "async"),
	)
	jobSpan.End()

	h.obs.Info(ctx, "Build start event processed successfully",
		"third_party_id", buildRequest.ThirdPartyID,
		"parser_id", buildRequest.ParserID,
		"correlation_id", buildRequest.CorrelationID,
		"response_status", response.Status,
		"response_message", response.Message,
		"job_name", response.JobName)

	return response, nil
}

// processBuildCompleteEvent processes build completion events
func (h *EventHandlerImpl) processBuildCompleteEvent(ctx context.Context, event *cloudevents.Event) (*builds.HandlerResponse, error) {
	ctx, span := h.obs.StartSpan(ctx, "process_build_complete_event")
	defer span.End()

	// Parse completion data using focused event processor
	var completionData builds.BuildCompletionEventData
	if err := event.DataAs(&completionData); err != nil {
		h.obs.Error(ctx, err, "Failed to parse build completion event data")
		return nil, errors.NewValidationError("build_completion_event_data", nil, fmt.Sprintf("failed to parse build completion event data: %v", err))
	}

	h.obs.Info(ctx, "Build completion event details",
		"third_party_id", completionData.ThirdPartyID,
		"parser_id", completionData.ParserID,
		"status", completionData.Status,
		"image_uri", completionData.ImageURI,
		"job_name", completionData.JobName)

	// 🎯 IDEMPOTENT HANDLING: Create lambda and triggers ONLY if Docker image exists (build successful)
	// This ensures lambda and triggers are created only when there's a valid image to deploy

	// Handle failed builds by cleaning up the failed job and NOT creating lambda/triggers
	if completionData.Status != "success" {
		h.obs.Info(ctx, "Build was not successful, cleaning up failed job and skipping lambda/trigger creation",
			"status", completionData.Status,
			"error", completionData.ErrorMessage,
			"job_name", completionData.JobName,
			"image_uri", completionData.ImageURI)

		// Record K8s job failure metric
		if h.obs.GetMetrics() != nil {
			metricsRec := observability.NewMetricsRecorder(h.obs)
			metricsRec.RecordK8sJobFailure(ctx, "kaniko", "build_failed")
		}

		// Clean up the failed job to prevent resource accumulation
		jobManager := h.container.GetJobManager()
		if err := jobManager.CleanupFailedJob(ctx, completionData.JobName); err != nil {
			h.obs.Error(ctx, err, "Failed to cleanup failed job",
				"job_name", completionData.JobName,
				"third_party_id", completionData.ThirdPartyID,
				"parser_id", completionData.ParserID,
				"correlation_id", completionData.CorrelationID)
			// Don't return error here as the main issue is handled
		} else {
			h.obs.Info(ctx, "Successfully cleaned up failed job",
				"job_name", completionData.JobName,
				"third_party_id", completionData.ThirdPartyID,
				"parser_id", completionData.ParserID,
				"correlation_id", completionData.CorrelationID)
		}

		// Return response indicating build failed, job cleaned up, and no lambda/triggers created
		return &builds.HandlerResponse{
			Status:        "failed",
			Message:       fmt.Sprintf("Build %s, job cleaned up, no lambda/triggers created (no Docker image)", completionData.Status),
			JobName:       completionData.JobName,
			CorrelationID: completionData.CorrelationID,
		}, nil
	}

	// Record K8s job success metric
	if h.obs.GetMetrics() != nil {
		metricsRec := observability.NewMetricsRecorder(h.obs)
		metricsRec.RecordK8sJobSuccess(ctx, "kaniko")
	}

	// 🚀 CREATE LAMBDA/TRIGGERS: Only create if build was successful (Docker image exists)
	// This ensures lambda and triggers are created only when there's a valid image to deploy
	return h.createServiceAndTrigger(ctx, &completionData)
}

// createServiceAndTrigger creates Knative service and trigger for successful builds
func (h *EventHandlerImpl) createServiceAndTrigger(ctx context.Context, completionData *builds.BuildCompletionEventData) (*builds.HandlerResponse, error) {
	ctx, span := h.obs.StartSpan(ctx, "create_service_and_trigger")
	defer span.End()

	// Validate completion data
	if completionData == nil {
		h.obs.Error(ctx, fmt.Errorf("completion data is nil"), "Completion data cannot be nil")
		return nil, errors.NewConfigurationError("event_handler", "completion_data", "completion data cannot be nil")
	}

	serviceManager := h.container.GetServiceManager()

	// Generate service name
	serviceName := serviceManager.GenerateServiceName(completionData.ThirdPartyID, completionData.ParserID)

	h.obs.Info(ctx, "Creating Knative service and trigger (build successful - Docker image exists)",
		"service_name", serviceName,
		"third_party_id", completionData.ThirdPartyID,
		"parser_id", completionData.ParserID,
		"image_uri", completionData.ImageURI,
		"build_status", completionData.Status)

	// 🚀 PARALLEL CREATION: Create Knative service and trigger simultaneously
	// This improves performance by creating both resources in parallel instead of sequentially
	// 🎯 IDEMPOTENT: This creates lambda and triggers for successful builds only
	if err := serviceManager.CreateService(ctx, serviceName, completionData); err != nil {
		h.obs.Error(ctx, err, "Failed to create Knative service",
			"service_name", serviceName,
			"build_status", completionData.Status)
		return nil, err
	}

	h.obs.Info(ctx, "Successfully created Knative service and trigger in parallel",
		"service_name", serviceName,
		"third_party_id", completionData.ThirdPartyID,
		"parser_id", completionData.ParserID,
		"build_status", completionData.Status)

	// Return success status for successful builds
	status := "service_created"
	message := "Knative service and trigger created successfully in parallel"

	return &builds.HandlerResponse{
		Status:        status,
		Message:       message,
		JobName:       completionData.JobName,
		CorrelationID: completionData.CorrelationID,
	}, nil
}

// createBuildContext creates build context using focused component
func (h *EventHandlerImpl) createBuildContext(ctx context.Context, buildRequest *builds.BuildRequest) (string, error) {
	return h.container.GetBuildContextManager().CreateBuildContext(ctx, buildRequest)
}

// createOrUpdateJob creates or updates job using focused components
func (h *EventHandlerImpl) createOrUpdateJob(ctx context.Context, buildRequest *builds.BuildRequest, buildContextKey string) (*builds.HandlerResponse, error) {
	ctx, span := h.obs.StartSpanWithAttributes(ctx, "create_or_update_job", map[string]string{
		"third_party_id":    buildRequest.ThirdPartyID,
		"parser_id":         buildRequest.ParserID,
		"correlation_id":    buildRequest.CorrelationID,
		"build_context_key": buildContextKey,
	})
	defer span.End()

	jobManager := h.container.GetJobManager()

	h.obs.Info(ctx, "Checking for existing job",
		"third_party_id", buildRequest.ThirdPartyID,
		"parser_id", buildRequest.ParserID,
		"correlation_id", buildRequest.CorrelationID)

	// Find existing job using focused job finder
	existingJob, err := jobManager.FindExistingJob(ctx, buildRequest.ThirdPartyID, buildRequest.ParserID)
	if err != nil {
		h.obs.Error(ctx, err, "Failed to find existing job",
			"third_party_id", buildRequest.ThirdPartyID,
			"parser_id", buildRequest.ParserID,
			"correlation_id", buildRequest.CorrelationID,
			"error_details", err.Error())
		return nil, err
	}

	if existingJob != nil {
		h.obs.Info(ctx, "Found existing job, checking if it's still active",
			"job_name", existingJob.Name,
			"job_uid", string(existingJob.UID),
			"third_party_id", buildRequest.ThirdPartyID,
			"parser_id", buildRequest.ParserID,
			"correlation_id", buildRequest.CorrelationID,
			"job_status_active", existingJob.Status.Active,
			"job_status_failed", existingJob.Status.Failed,
			"job_status_succeeded", existingJob.Status.Succeeded)

		// Check if the existing job is still active (only truly running jobs)
		// Use the job manager's IsJobRunning method for consistent status checking
		jobManager := h.container.GetJobManager()
		if jobManager.IsJobRunning(existingJob) {
			h.obs.Info(ctx, "Existing job is still active, skipping new job creation to prevent duplicates",
				"job_name", existingJob.Name,
				"third_party_id", buildRequest.ThirdPartyID,
				"parser_id", buildRequest.ParserID,
				"correlation_id", buildRequest.CorrelationID,
				"job_status_active", existingJob.Status.Active,
				"job_status_failed", existingJob.Status.Failed,
				"job_status_succeeded", existingJob.Status.Succeeded)

			return &builds.HandlerResponse{
				Status:  "skipped",
				Message: "Build already in progress",
				JobName: existingJob.Name,
			}, nil
		}

		// Check if the existing job failed recently to implement failure backoff
		if existingJob.Status.Failed > 0 {
			// Check if the job failed within the failure backoff period
			jobAge := time.Since(existingJob.CreationTimestamp.Time)
			failureBackoffPeriod := constants.JobFailureBackoffPeriodDefault

			if jobAge < failureBackoffPeriod {
				h.obs.Info(ctx, "Existing job failed recently, implementing failure backoff",
					"job_name", existingJob.Name,
					"job_age", jobAge,
					"failure_backoff_period", failureBackoffPeriod,
					"failed_count", existingJob.Status.Failed,
					"third_party_id", buildRequest.ThirdPartyID,
					"parser_id", buildRequest.ParserID,
					"correlation_id", buildRequest.CorrelationID)

				// Clean up the failed job
				if err := jobManager.CleanupFailedJob(ctx, existingJob.Name); err != nil {
					h.obs.Error(ctx, err, "Failed to cleanup failed job during backoff",
						"job_name", existingJob.Name,
						"third_party_id", buildRequest.ThirdPartyID,
						"parser_id", buildRequest.ParserID,
						"correlation_id", buildRequest.CorrelationID)
				}

				return &builds.HandlerResponse{
					Status:  "backoff",
					Message: fmt.Sprintf("Build failed recently, backoff period active (job age: %v)", jobAge),
					JobName: existingJob.Name,
				}, nil
			}
		}

		// Only delete and recreate if the job is not active and not in backoff period
		h.obs.Info(ctx, "Existing job is not active, deleting and recreating",
			"job_name", existingJob.Name,
			"third_party_id", buildRequest.ThirdPartyID,
			"parser_id", buildRequest.ParserID,
			"correlation_id", buildRequest.CorrelationID)

		// Delete existing job
		if err := jobManager.DeleteJob(ctx, existingJob.Name); err != nil {
			h.obs.Error(ctx, err, "Failed to delete existing job",
				"job_name", existingJob.Name,
				"third_party_id", buildRequest.ThirdPartyID,
				"parser_id", buildRequest.ParserID,
				"correlation_id", buildRequest.CorrelationID,
				"error_details", err.Error())
			return nil, err
		}

		h.obs.Info(ctx, "Successfully deleted existing job, waiting for deletion to complete",
			"job_name", existingJob.Name,
			"third_party_id", buildRequest.ThirdPartyID,
			"parser_id", buildRequest.ParserID,
			"correlation_id", buildRequest.CorrelationID)

		// Wait for the job to be fully deleted before creating a new one
		if err := h.waitForJobDeletion(ctx, existingJob.Name); err != nil {
			h.obs.Error(ctx, err, "Failed to wait for job deletion",
				"job_name", existingJob.Name,
				"third_party_id", buildRequest.ThirdPartyID,
				"parser_id", buildRequest.ParserID,
				"correlation_id", buildRequest.CorrelationID,
				"error_details", err.Error())
			return nil, err
		}

		h.obs.Info(ctx, "Job deletion completed, proceeding with new job creation",
			"job_name", existingJob.Name,
			"third_party_id", buildRequest.ThirdPartyID,
			"parser_id", buildRequest.ParserID,
			"correlation_id", buildRequest.CorrelationID)
	} else {
		h.obs.Info(ctx, "No existing job found, will create new job",
			"third_party_id", buildRequest.ThirdPartyID,
			"parser_id", buildRequest.ParserID,
			"correlation_id", buildRequest.CorrelationID)
	}

	h.obs.Info(ctx, "Creating new job",
		"third_party_id", buildRequest.ThirdPartyID,
		"parser_id", buildRequest.ParserID,
		"correlation_id", buildRequest.CorrelationID)

	return h.createNewJob(ctx, buildRequest)
}

// createNewJob creates a new job using focused components
func (h *EventHandlerImpl) createNewJob(ctx context.Context, buildRequest *builds.BuildRequest) (*builds.HandlerResponse, error) {
	ctx, span := h.obs.StartSpanWithAttributes(ctx, "create_new_job", map[string]string{
		"third_party_id": buildRequest.ThirdPartyID,
		"parser_id":      buildRequest.ParserID,
		"correlation_id": buildRequest.CorrelationID,
	})
	defer span.End()

	jobManager := h.container.GetJobManager()

	h.obs.Info(ctx, "Generating job name",
		"third_party_id", buildRequest.ThirdPartyID,
		"parser_id", buildRequest.ParserID,
		"correlation_id", buildRequest.CorrelationID)

	// Generate job name using focused job creator
	jobName := jobManager.GenerateJobName(buildRequest.ThirdPartyID, buildRequest.ParserID)

	h.obs.Info(ctx, "Generated job name",
		"job_name", jobName,
		"third_party_id", buildRequest.ThirdPartyID,
		"parser_id", buildRequest.ParserID,
		"correlation_id", buildRequest.CorrelationID)

	// Create job using focused job creator
	h.obs.Info(ctx, "Creating job in Kubernetes",
		"job_name", jobName,
		"third_party_id", buildRequest.ThirdPartyID,
		"parser_id", buildRequest.ParserID,
		"correlation_id", buildRequest.CorrelationID)

	// Try to create the job with retry logic to handle race conditions
	var job *batchv1.Job
	var err error
	maxRetries := 3
	retryDelay := 100 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		job, err = jobManager.CreateJob(ctx, jobName, buildRequest)
		if err == nil {
			break // Success, exit retry loop
		}

		// Check if this is a conflict error (job already exists)
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "AlreadyExists") {
			h.obs.Info(ctx, "Job already exists, checking if it's the same job",
				"job_name", jobName,
				"attempt", attempt+1,
				"max_retries", maxRetries,
				"third_party_id", buildRequest.ThirdPartyID,
				"parser_id", buildRequest.ParserID,
				"correlation_id", buildRequest.CorrelationID)

			// Check if the existing job is for the same parser
			existingJob, findErr := jobManager.FindExistingJob(ctx, buildRequest.ThirdPartyID, buildRequest.ParserID)
			if findErr == nil && existingJob != nil {
				h.obs.Info(ctx, "Found existing job for same parser, returning success",
					"job_name", existingJob.Name,
					"third_party_id", buildRequest.ThirdPartyID,
					"parser_id", buildRequest.ParserID,
					"correlation_id", buildRequest.CorrelationID)

				return &builds.HandlerResponse{
					Status:  "started",
					Message: "Build already started by another request",
					JobName: existingJob.Name,
				}, nil
			}

			// If we can't find the existing job, wait and retry
			if attempt < maxRetries-1 {
				h.obs.Info(ctx, "Waiting before retry",
					"job_name", jobName,
					"attempt", attempt+1,
					"retry_delay", retryDelay,
					"third_party_id", buildRequest.ThirdPartyID,
					"parser_id", buildRequest.ParserID,
					"correlation_id", buildRequest.CorrelationID)

				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(retryDelay):
					retryDelay *= 2 // Exponential backoff
				}
				continue
			}
		}

		// For non-conflict errors, don't retry
		break
	}

	if err != nil {
		h.obs.Error(ctx, err, "Failed to create job in Kubernetes after retries",
			"job_name", jobName,
			"max_retries", maxRetries,
			"third_party_id", buildRequest.ThirdPartyID,
			"parser_id", buildRequest.ParserID,
			"correlation_id", buildRequest.CorrelationID,
			"error_details", err.Error())
		return nil, err
	}

	h.obs.Info(ctx, "Successfully created job in Kubernetes",
		"job_name", job.Name,
		"job_uid", string(job.UID),
		"third_party_id", buildRequest.ThirdPartyID,
		"parser_id", buildRequest.ParserID,
		"correlation_id", buildRequest.CorrelationID,
		"job_creation_timestamp", job.CreationTimestamp.String())

	return &builds.HandlerResponse{
		Status:  "started",
		Message: "Build started successfully",
		JobName: job.Name,
	}, nil
}

// waitForJobDeletion waits for a job to be fully deleted from Kubernetes
func (h *EventHandlerImpl) waitForJobDeletion(ctx context.Context, jobName string) error {
	ctx, span := h.obs.StartSpan(ctx, "wait_for_job_deletion")
	defer span.End()

	h.obs.Info(ctx, "Waiting for job deletion to complete", "job_name", jobName)

	// Get the job manager to check job status
	jobManager := h.container.GetJobManager()

	// Wait for the job to be deleted using configured timeouts
	maxWaitTime := h.config.Kubernetes.JobDeletionWaitTimeout
	checkInterval := h.config.Kubernetes.JobDeletionCheckInterval
	elapsed := time.Duration(0)

	for elapsed < maxWaitTime {
		// Check if the job still exists
		job, err := jobManager.GetJob(ctx, jobName)
		if err != nil {
			// If we get a "not found" error, the job has been deleted
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "NotFound") {
				h.obs.Info(ctx, "Job deletion confirmed", "job_name", jobName, "elapsed_time", elapsed.String())
				return nil
			}
			// For other errors, log but continue waiting
			h.obs.Info(ctx, "Error checking job status, continuing to wait",
				"job_name", jobName,
				"error", err.Error(),
				"elapsed_time", elapsed.String())
		} else if job != nil {
			// Job still exists, check if it's being deleted
			if job.DeletionTimestamp != nil {
				h.obs.Info(ctx, "Job is being deleted, continuing to wait",
					"job_name", jobName,
					"deletion_timestamp", job.DeletionTimestamp.String(),
					"elapsed_time", elapsed.String())
			} else {
				h.obs.Info(ctx, "Job still exists and not being deleted",
					"job_name", jobName,
					"elapsed_time", elapsed.String())
			}
		}

		// Wait before checking again
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(checkInterval):
			elapsed += checkInterval
		}
	}

	// If we get here, the job deletion timed out
	return fmt.Errorf("job deletion timed out after %v", maxWaitTime)
}

// isBuildStartEvent checks if event is a build start event
func (h *EventHandlerImpl) isBuildStartEvent(event *cloudevents.Event) bool {
	return event.Type() == builds.EventTypeBuildStart
}

// isJobStartEvent checks if event is a job start event
func (h *EventHandlerImpl) isJobStartEvent(event *cloudevents.Event) bool {
	return event.Type() == builds.EventTypeJobStart
}

// isBuildCompleteEvent checks if event is a build completion event
func (h *EventHandlerImpl) isBuildCompleteEvent(event *cloudevents.Event) bool {
	return event.Type() == builds.EventTypeBuildComplete || event.Type() == builds.EventTypeBuildFailed
}

// isParserStartEvent checks if event is a parser start event
func (h *EventHandlerImpl) isParserStartEvent(event *cloudevents.Event) bool {
	return event.Type() == builds.EventTypeParserStart
}

// isParserCompleteEvent checks if event is a parser completion event
func (h *EventHandlerImpl) isParserCompleteEvent(event *cloudevents.Event) bool {
	return event.Type() == builds.EventTypeParserComplete || event.Type() == builds.EventTypeParserFailed
}

// isServiceDeleteEvent checks if event is a service deletion event
func (h *EventHandlerImpl) isServiceDeleteEvent(event *cloudevents.Event) bool {
	return event.Type() == builds.EventTypeServiceDelete
}

// processParserStartEvent handles parser start events
func (h *EventHandlerImpl) processParserStartEvent(ctx context.Context, event *cloudevents.Event) (*builds.HandlerResponse, error) {
	ctx, span := h.obs.StartSpan(ctx, "process_parser_start_event")
	defer span.End()

	// For parser start events, we extract basic information from the event
	var eventData builds.BuildEventData
	if err := event.DataAs(&eventData); err != nil {
		h.obs.Error(ctx, err, "Failed to parse parser start event data")
		return nil, errors.NewValidationError("parser_start_event_data", nil, fmt.Sprintf("failed to parse parser start event data: %v", err))
	}

	h.obs.Info(ctx, "Parser start event received",
		"third_party_id", eventData.ThirdPartyID,
		"parser_id", eventData.ParserID)

	// For parser start events, we typically just acknowledge receipt
	// The actual processing is handled by the created Knative service
	return &builds.HandlerResponse{
		Status:        "acknowledged",
		Message:       "Parser start event acknowledged",
		CorrelationID: eventData.ContextID,
	}, nil
}

// processParserCompleteEvent handles parser completion events
func (h *EventHandlerImpl) processParserCompleteEvent(ctx context.Context, event *cloudevents.Event) (*builds.HandlerResponse, error) {
	ctx, span := h.obs.StartSpan(ctx, "process_parser_complete_event")
	defer span.End()

	// Parse completion data using the same structure as build completion
	var completionData builds.BuildCompletionEventData
	if err := event.DataAs(&completionData); err != nil {
		h.obs.Error(ctx, err, "Failed to parse parser completion event data")
		return nil, errors.NewValidationError("parser_completion_event_data", nil, fmt.Sprintf("failed to parse parser completion event data: %v", err))
	}

	h.obs.Info(ctx, "Parser completion event details",
		"third_party_id", completionData.ThirdPartyID,
		"parser_id", completionData.ParserID,
		"status", completionData.Status)

	// Handle parser completion based on status
	if completionData.Status == "success" {
		h.obs.Info(ctx, "Parser completed successfully",
			"third_party_id", completionData.ThirdPartyID,
			"parser_id", completionData.ParserID)
	} else {
		h.obs.Info(ctx, "Parser failed",
			"status", completionData.Status,
			"error", completionData.ErrorMessage)
	}

	return &builds.HandlerResponse{
		Status:        "processed",
		Message:       fmt.Sprintf("Parser %s processed", completionData.Status),
		CorrelationID: completionData.CorrelationID,
	}, nil
}

// processJobStartEvent handles job start events from RabbitMQ kaniko-jobs queue
func (h *EventHandlerImpl) processJobStartEvent(ctx context.Context, event *cloudevents.Event) (*builds.HandlerResponse, error) {
	ctx, span := h.obs.StartSpanWithAttributes(ctx, "process_job_start_event", map[string]string{
		"event.type":    event.Type(),
		"event.source":  event.Source(),
		"event.id":      event.ID(),
		"event.subject": event.Subject(),
	})
	defer span.End()

	// Parse job start event data
	var jobStartData builds.JobStartEventData
	if err := event.DataAs(&jobStartData); err != nil {
		h.obs.Error(ctx, err, "Failed to parse job start event data")
		return nil, errors.NewValidationError("job_start_event_data", nil, fmt.Sprintf("failed to parse job start event data: %v", err))
	}

	// Add job start details to the main span
	span.SetAttributes(
		attribute.String("job.third_party_id", jobStartData.ThirdPartyID),
		attribute.String("job.parser_id", jobStartData.ParserID),
		attribute.String("job.correlation_id", jobStartData.CorrelationID),
		attribute.String("job.job_name", jobStartData.JobName),
	)

	h.obs.Info(ctx, "Processing job start event from kaniko-jobs queue",
		"third_party_id", jobStartData.ThirdPartyID,
		"parser_id", jobStartData.ParserID,
		"correlation_id", jobStartData.CorrelationID,
		"job_name", jobStartData.JobName,
		"priority", jobStartData.Priority)

	// Convert JobStartEventData to BuildRequest for async job creation
	buildRequest := &builds.BuildRequest{
		ThirdPartyID:  jobStartData.ThirdPartyID,
		ParserID:      jobStartData.ParserID,
		CorrelationID: jobStartData.CorrelationID,
		Metadata:      jobStartData.Parameters, // Use Parameters as Metadata
	}

	// Create build context with tracing
	ctx, contextSpan := h.obs.StartSpan(ctx, "create_build_context")
	buildContextKey, err := h.createBuildContext(ctx, buildRequest)
	if err != nil {
		contextSpan.End()
		h.obs.Error(ctx, err, "Failed to create build context for job start event",
			"third_party_id", jobStartData.ThirdPartyID,
			"parser_id", jobStartData.ParserID,
			"correlation_id", jobStartData.CorrelationID)
		return nil, err
	}
	contextSpan.SetAttributes(attribute.String("build.context_key", buildContextKey))
	contextSpan.End()

	// Get async job creator from container
	asyncJobCreator := h.container.GetAsyncJobCreator()
	if asyncJobCreator == nil {
		h.obs.Error(ctx, fmt.Errorf("async job creator not available"), "Async job creator not available")
		return nil, fmt.Errorf("async job creator not available")
	}

	// Queue job creation asynchronously
	jobName, err := asyncJobCreator.CreateJobAsync(ctx, buildRequest)
	if err != nil {
		h.obs.Error(ctx, err, "Failed to queue job creation for job start event",
			"third_party_id", jobStartData.ThirdPartyID,
			"parser_id", jobStartData.ParserID,
			"correlation_id", jobStartData.CorrelationID,
			"error_details", err.Error())
		return nil, err
	}

	// Return immediate response - job creation is now handled asynchronously
	response := &builds.HandlerResponse{
		Status:        "queued",
		Message:       "Job creation queued successfully from kaniko-jobs queue",
		JobName:       jobName,
		CorrelationID: jobStartData.CorrelationID,
	}

	h.obs.Info(ctx, "Job start event processed successfully",
		"third_party_id", jobStartData.ThirdPartyID,
		"parser_id", jobStartData.ParserID,
		"correlation_id", jobStartData.CorrelationID,
		"response_status", response.Status,
		"response_message", response.Message,
		"job_name", response.JobName)

	return response, nil
}

// processServiceDeleteEvent handles service deletion events
func (h *EventHandlerImpl) processServiceDeleteEvent(ctx context.Context, event *cloudevents.Event) (*builds.HandlerResponse, error) {
	ctx, span := h.obs.StartSpan(ctx, "process_service_delete_event")
	defer span.End()

	// Parse deletion data
	var deleteData builds.ServiceDeleteEventData
	if err := event.DataAs(&deleteData); err != nil {
		h.obs.Error(ctx, err, "Failed to parse service deletion event data")
		return nil, errors.NewValidationError("service_delete_event_data", nil, fmt.Sprintf("failed to parse service deletion event data: %v", err))
	}

	h.obs.Info(ctx, "Service deletion event details",
		"third_party_id", deleteData.ThirdPartyID,
		"parser_id", deleteData.ParserID,
		"service_name", deleteData.ServiceName,
		"reason", deleteData.Reason)

	serviceManager := h.container.GetServiceManager()

	// Generate service name if not provided
	serviceName := deleteData.ServiceName
	if serviceName == "" {
		serviceName = serviceManager.GenerateServiceName(deleteData.ThirdPartyID, deleteData.ParserID)
		h.obs.Info(ctx, "Generated service name for deletion", "service_name", serviceName)
	}

	h.obs.Info(ctx, "Deleting Knative service",
		"service_name", serviceName,
		"third_party_id", deleteData.ThirdPartyID,
		"parser_id", deleteData.ParserID)

	// Delete the service and all associated resources
	if err := serviceManager.DeleteService(ctx, serviceName); err != nil {
		h.obs.Error(ctx, err, "Failed to delete Knative service", "service_name", serviceName)
		return nil, err
	}

	h.obs.Info(ctx, "Successfully deleted Knative service",
		"service_name", serviceName,
		"third_party_id", deleteData.ThirdPartyID,
		"parser_id", deleteData.ParserID)

	return &builds.HandlerResponse{
		Status:        "deleted",
		Message:       "Knative service and associated resources deleted successfully",
		CorrelationID: deleteData.CorrelationID,
	}, nil
}

// processEventWithTracing processes the CloudEvent with comprehensive tracing
func (h *EventHandlerImpl) processEventWithTracing(ctx context.Context, event *cloudevents.Event, metricsRec *observability.MetricsRecorder) (*builds.HandlerResponse, error) {
	ctx, span := h.obs.StartSpanWithAttributes(ctx, "process_event", map[string]string{
		"event.type":   event.Type(),
		"event.source": event.Source(),
		"event.id":     event.ID(),
	})
	defer span.End()

	// Record event processing start
	start := time.Now()

	// Process event based on type
	var response *builds.HandlerResponse
	var err error

	if h.isBuildCompleteEvent(event) {
		response, err = h.processBuildCompleteEvent(ctx, event)
	} else if h.isBuildStartEvent(event) {
		response, err = h.processBuildStartEvent(ctx, event)
	} else if h.isJobStartEvent(event) {
		response, err = h.processJobStartEvent(ctx, event)
	} else if h.isParserCompleteEvent(event) {
		response, err = h.processParserCompleteEvent(ctx, event)
	} else if h.isParserStartEvent(event) {
		response, err = h.processParserStartEvent(ctx, event)
	} else if h.isServiceDeleteEvent(event) {
		response, err = h.processServiceDeleteEvent(ctx, event)
	} else {
		err = errors.NewValidationError("event_type", event.Type(), "unsupported event type")
	}

	// Record processing duration
	duration := time.Since(start)
	span.SetAttributes(attribute.Float64("processing.duration_seconds", duration.Seconds()))

	// Record metrics based on event type
	if h.isBuildStartEvent(event) {
		if buildRequest, parseErr := h.ParseBuildRequest(ctx, event); parseErr == nil {
			if err != nil {
				metricsRec.RecordBuildFailure(ctx, buildRequest.ThirdPartyID, buildRequest.ParserID, "processing_error")
			} else {
				metricsRec.RecordBuildSuccess(ctx, buildRequest.ThirdPartyID, buildRequest.ParserID)
			}
			metricsRec.RecordBuildDuration(ctx, buildRequest.ThirdPartyID, buildRequest.ParserID, duration)
		}
	} else if h.isParserStartEvent(event) {
		var eventData builds.BuildEventData
		if parseErr := event.DataAs(&eventData); parseErr == nil {
			if err != nil {
				metricsRec.RecordBuildFailure(ctx, eventData.ThirdPartyID, eventData.ParserID, "parser_processing_error")
			} else {
				metricsRec.RecordBuildSuccess(ctx, eventData.ThirdPartyID, eventData.ParserID)
			}
			metricsRec.RecordBuildDuration(ctx, eventData.ThirdPartyID, eventData.ParserID, duration)
		}
	}

	return response, err
}

// Shutdown gracefully shuts down all components
func (h *EventHandlerImpl) Shutdown(ctx context.Context) error {
	return h.container.Shutdown(ctx)
}

// =============================================================================
// EVENT PROCESSOR METHODS (CONSOLIDATED FROM EventProcessor)
// =============================================================================

// ValidateEvent validates a CloudEvent according to the CloudEvents specification
func (h *EventHandlerImpl) ValidateEvent(ctx context.Context, event *cloudevents.Event) error {
	_, span := h.obs.StartSpan(ctx, "validate_event")
	defer span.End()

	// Validate required headers
	eventType := event.Type()
	eventSource := event.Source()
	eventID := event.ID()

	if eventType == "" || eventSource == "" || eventID == "" {
		return errors.NewValidationError("cloud_event_headers", nil, "missing required CloudEvent headers")
	}

	// Check if event type is supported
	if !h.IsSupportedEventType(eventType) {
		return errors.NewValidationError("event_type", eventType, "unsupported event type")
	}

	// Parse event body based on event type
	if strings.Contains(eventType, "build.complete") {
		var eventData builds.BuildCompletionEventData
		if err := event.DataAs(&eventData); err != nil {
			return errors.NewValidationError("build_completion_event_body", nil, fmt.Sprintf("invalid build completion event body: %v", err))
		}
	} else if strings.Contains(eventType, "service.delete") {
		var eventData builds.ServiceDeleteEventData
		if err := event.DataAs(&eventData); err != nil {
			return errors.NewValidationError("service_delete_event_body", nil, fmt.Sprintf("invalid service delete event body: %v", err))
		}
	} else {
		var eventData builds.BuildEventData
		if err := event.DataAs(&eventData); err != nil {
			return errors.NewValidationError("event_body", nil, fmt.Sprintf("invalid event body: %v", err))
		}
	}

	return nil
}

// ParseBuildRequest parses a CloudEvent data into a BuildRequest
func (h *EventHandlerImpl) ParseBuildRequest(ctx context.Context, event *cloudevents.Event) (*builds.BuildRequest, error) {
	ctx, span := h.obs.StartSpan(ctx, "parse_build_request")
	defer span.End()

	// Log the raw event data for debugging
	h.obs.Info(ctx, "Parsing build request from CloudEvent",
		"event_type", event.Type(),
		"event_source", event.Source(),
		"event_id", event.ID(),
		"event_time", event.Time(),
		"data_content_type", event.DataContentType())

	// Log the raw event data if available
	if event.Data() != nil {
		h.obs.Info(ctx, "Raw event data",
			"data_type", fmt.Sprintf("%T", event.Data()),
			"data_length", len(fmt.Sprintf("%v", event.Data())))

		// Try to get the raw JSON data
		if rawData, err := json.Marshal(event.Data()); err == nil {
			h.obs.Info(ctx, "Raw JSON data",
				"json_data", string(rawData))
		}
	}

	var eventData builds.BuildEventData

	// Get the raw data and handle it properly
	rawData := event.Data()
	// Show preview of raw data (first 200 bytes)
	previewLength := 200
	if len(rawData) < previewLength {
		previewLength = len(rawData)
	}
	h.obs.Info(ctx, "Raw event data",
		"data_length", len(rawData),
		"data_preview", string(rawData[:previewLength]))

	// Set the eventData fields from thirdPartyId and parserId from the source and subject, where they are in the
	// form of network.notifi.<thirdPartyId> and network.notifi.<parserId>. Don't include the "network.notifi." prefix
	eventData.ThirdPartyID = strings.TrimPrefix(event.Source(), "network.notifi.")
	eventData.ParserID = event.Subject()

	// The raw data is now directly the event data (not a complete CloudEvent)
	// Unmarshal it directly into BuildEventData
	if err := json.Unmarshal(rawData, &eventData); err != nil {
		h.obs.Error(ctx, err, "Failed to unmarshal event data into BuildEventData",
			"event_type", event.Type(),
			"event_source", event.Source(),
			"error_details", err.Error())
		return nil, errors.NewValidationError("build_request_data", nil, fmt.Sprintf("failed to parse build request data: %v", err))
	}

	// Debug: Log the actual parsed data
	h.obs.Info(ctx, "DEBUG: Parsed event data",
		"third_party_id", eventData.ThirdPartyID,
		"parser_id", eventData.ParserID)

	// Log the parsed event data for debugging
	h.obs.Info(ctx, "Successfully parsed BuildEventData",
		"third_party_id", eventData.ThirdPartyID,
		"parser_id", eventData.ParserID)

	correlationID := h.getCorrelationID(ctx, event)

	// Generate source bucket from service configuration
	sourceBucket := h.config.AWS.GetS3SourceBucket()

	// Extract block ID from event data for parallel processing
	blockID := ""

	// Generate source key for parser code location
	sourceKey := fmt.Sprintf("global/parser/%s", eventData.ParserID)

	// Create default build config from service configuration
	buildConfig := builds.BuildConfig{
		DockerfilePath: "Dockerfile",
		DockerContext:  ".",
		CPULimit:       h.config.Build.GetCPULimit(),
		MemoryLimit:    h.config.Build.GetMemoryLimit(),
		TimeoutSeconds: int(h.config.Build.GetBuildTimeout().Seconds()),
		Environment:    make(map[string]string),
		BuildSteps:     []builds.BuildStep{},
		CacheEnabled:   true,
	}

	// Debug: Log the actual parsed data
	h.obs.Info(ctx, "DEBUG: Parsed build event data",
		"third_party_id", eventData.ThirdPartyID,
		"parser_id", eventData.ParserID)

	// Extract build type and runtime from event data or use defaults
	buildType := "container"                                        // Default to container builds
	runtime := h.config.Lambda.DefaultRuntime                       // Get runtime from configuration
	sourceURL := fmt.Sprintf("s3://%s/%s", sourceBucket, sourceKey) // Generate source URL from S3 location

	// Override defaults if provided in event data
	if buildTypeStr, ok := eventData.GetParameterAsString("buildType"); ok {
		buildType = buildTypeStr
		h.obs.Info(ctx, "Using build type from parameters", "build_type", buildType)
	}
	if runtimeStr, ok := eventData.GetParameterAsString("runtime"); ok {
		runtime = runtimeStr
		h.obs.Info(ctx, "Using runtime from parameters", "runtime", runtime)
	}

	// Create build request with all necessary information
	buildRequest := &builds.BuildRequest{
		ThirdPartyID:  eventData.ThirdPartyID,
		ParserID:      eventData.ParserID,
		BuildType:     buildType,
		Runtime:       runtime,
		SourceURL:     sourceURL,
		SourceBucket:  sourceBucket,
		SourceKey:     sourceKey,
		BlockID:       blockID,
		BuildConfig:   buildConfig,
		CorrelationID: correlationID,
		CreatedAt:     time.Now(),
	}

	h.obs.Info(ctx, "Successfully created BuildRequest",
		"third_party_id", buildRequest.ThirdPartyID,
		"parser_id", buildRequest.ParserID,
		"build_type", buildRequest.BuildType,
		"runtime", buildRequest.Runtime,
		"source_url", buildRequest.SourceURL,
		"correlation_id", buildRequest.CorrelationID)

	return buildRequest, nil
}

// IsSupportedEventType checks if an event type is supported by the service
func (h *EventHandlerImpl) IsSupportedEventType(eventType string) bool {
	supportedTypes := []string{
		// Build Events
		builds.EventTypeBuildStart,
		builds.EventTypeBuildComplete,
		builds.EventTypeBuildFailed,
		builds.EventTypeBuildTimeout,
		builds.EventTypeBuildCancel,

		// Parser Events
		builds.EventTypeParserStart,
		builds.EventTypeParserComplete,
		builds.EventTypeParserFailed,

		// Service Management Events
		builds.EventTypeServiceDelete,
	}

	for _, supportedType := range supportedTypes {
		if eventType == supportedType {
			return true
		}
	}
	return false
}

// getCorrelationID gets or generates correlation ID
func (h *EventHandlerImpl) getCorrelationID(ctx context.Context, event *cloudevents.Event) string {
	// Try to get correlation ID from context first
	if correlationID, ok := ctx.Value(observability.CorrelationIDKey).(string); ok && correlationID != "" {
		return correlationID
	}

	// Try to get from event extensions
	if correlationID, ok := event.Extensions()["correlationid"].(string); ok && correlationID != "" {
		return correlationID
	}

	// Generate new correlation ID
	return uuid.New().String()
}
