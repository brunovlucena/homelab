// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🎯 HANDLER INTERFACES - Service Component Contracts
//
//	🎯 Purpose: Define contracts for all service components and handlers
//	💡 Features: Interface segregation, dependency injection, loose coupling
//
//	🏛️ ARCHITECTURE:
//	🎯 HTTP Handlers - Server operations and request routing
//	🔧 Job Management - Kubernetes job lifecycle operations
//	📥 Event Processing - CloudEvent validation and processing
//	🏥 Health Management - Health checks and monitoring
//	📦 Build Context - Build context creation and validation
//	🔗 Component Container - Dependency injection container
//
//	🔧 INTERFACE CATEGORIES:
//	🌐 HTTP Operations - Server management and request handling
//	⚙️ Job Operations - Kubernetes job lifecycle management
//	📨 Event Operations - CloudEvent processing and validation
//	🏥 Health Operations - Health monitoring and status checks
//	📦 Build Operations - Build context management
//	🔗 Container Operations - Dependency injection management
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package handler

import (
	"context"
	"io"
	"net/http"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"knative-lambda-new/internal/config"
	"knative-lambda-new/internal/observability"
	"knative-lambda-new/pkg/builds"
)

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🌐 HTTP HANDLER INTERFACES - "HTTP server and request operations"      │
// └─────────────────────────────────────────────────────────────────────────┘

// HTTPHandler manages HTTP server operations and routing
// This interface provides the core HTTP server functionality including
// server startup and route registration for all HTTP endpoints
type HTTPHandler interface {
	// StartServer starts the HTTP server and begins listening for requests
	// This method should handle graceful shutdown and error propagation
	StartServer(ctx context.Context) error

	// RegisterRoutes registers all HTTP routes with the provided router
	// This method sets up all endpoints including health checks, build operations,
	// and CloudEvent processing endpoints
	RegisterRoutes(router http.Handler) http.Handler
}

// CloudEventHandler handles CloudEvent-specific operations
// This interface provides the contract for processing incoming CloudEvents
// and managing the event processing pipeline
type CloudEventHandler interface {
	// HandleCloudEvent handles incoming CloudEvents from HTTP requests
	// This method should parse, validate, and process CloudEvents according
	// to the CloudEvents specification
	HandleCloudEvent(w http.ResponseWriter, r *http.Request)
}

// HealthHandler handles health check operations

// BuildHandler handles build-related HTTP operations
// This interface provides the contract for build management operations
// including listing, retrieving, and canceling builds
type BuildHandler interface {
	// HandleListBuilds handles build listing requests
	// This should return a list of all builds with their current status
	HandleListBuilds(w http.ResponseWriter, r *http.Request)

	// HandleGetBuild handles individual build requests
	// This should return detailed information about a specific build
	HandleGetBuild(w http.ResponseWriter, r *http.Request)

	// HandleCancelBuild handles build cancellation requests
	// This should gracefully cancel a running build and clean up resources
	HandleCancelBuild(w http.ResponseWriter, r *http.Request)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  ⚙️ JOB MANAGEMENT INTERFACES - "Kubernetes job lifecycle operations"   │
// └─────────────────────────────────────────────────────────────────────────┘

// JobCreator handles job creation operations
// This interface provides the contract for creating new Kubernetes jobs
// for building Lambda functions
type JobCreator interface {
	// CreateJob creates a new Kubernetes job for building a Lambda function
	// This method should create a job with the appropriate configuration,
	// environment variables, and resource limits
	CreateJob(ctx context.Context, jobName string, buildRequest *builds.BuildRequest) (*batchv1.Job, error)

	// GenerateJobName generates a unique job name based on third party ID and parser ID
	// This ensures job names are unique across the cluster and follow naming conventions
	GenerateJobName(thirdPartyID, parserID string) string
}

// JobFinder handles job discovery operations
// This interface provides the contract for finding and retrieving existing jobs
type JobFinder interface {
	// FindExistingJob finds an existing job by third party ID and parser ID
	// This method should search for jobs that match the given criteria
	FindExistingJob(ctx context.Context, thirdPartyID, parserID string) (*batchv1.Job, error)

	// GetJob retrieves a job by its name
	// This method should return the job if it exists, or an error if not found
	GetJob(ctx context.Context, jobName string) (*batchv1.Job, error)
}

// JobManager handles job lifecycle operations
// This interface combines job creation and discovery operations with
// lifecycle management including deletion and cleanup
type JobManager interface {
	JobCreator
	JobFinder
	JobStatusChecker

	// DeleteJob deletes a job by name
	// This method should gracefully delete the job and clean up associated resources
	DeleteJob(ctx context.Context, jobName string) error

	// CleanupFailedJob cleans up a failed job and its associated resources
	// This method should handle cleanup of failed jobs including logs, volumes, and other resources
	CleanupFailedJob(ctx context.Context, jobName string) error

	// HasFailedJobs checks if there are any failed jobs in the namespace
	// This method should return true if any job has failed, preventing new job creation
	HasFailedJobs(ctx context.Context) (bool, error)

	// CountActiveJobs counts the number of active jobs in the namespace
	// This method is used to enforce concurrent job limits
	CountActiveJobs(ctx context.Context) (int, error)
}

// JobStatusChecker handles job status checking operations
// This interface provides the contract for checking the status of Kubernetes jobs
type JobStatusChecker interface {
	// IsJobRunning checks if a job is currently running
	// This method should check the job's status to determine if it's actively running
	IsJobRunning(job *batchv1.Job) bool

	// IsJobFailed checks if a job has failed
	// This method should check the job's status to determine if it has failed
	IsJobFailed(job *batchv1.Job) bool

	// IsJobSucceeded checks if a job has succeeded
	// This method should check the job's status to determine if it has completed successfully
	IsJobSucceeded(job *batchv1.Job) bool
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🏥 HEALTH MANAGEMENT INTERFACES - "Health monitoring and status checks" │
// └─────────────────────────────────────────────────────────────────────────┘

// ServiceCreator handles service creation operations
// This interface provides the contract for creating and updating Knative services
// and associated Kubernetes resources
type ServiceCreator interface {
	// CreateService creates or updates a Knative service for a completed build
	// This method should create or update all necessary Kubernetes resources including
	// the Knative service, service account, config maps, and monitoring resources
	CreateService(ctx context.Context, serviceName string, completionData *builds.BuildCompletionEventData) error

	// GenerateServiceName generates a unique service name based on third party ID and parser ID
	// This ensures service names are unique and follow naming conventions
	GenerateServiceName(thirdPartyID, parserID string) string
}

// ServiceChecker handles service verification operations
// This interface provides the contract for checking service existence and status
type ServiceChecker interface {
	// CheckServiceExists checks if a service exists in the cluster
	// This method should verify that the specified service exists and is accessible
	CheckServiceExists(ctx context.Context, serviceName string) (bool, error)
}

// ServiceManager handles service lifecycle operations
// This interface combines service creation and checking operations with
// resource management for Knative services
type ServiceManager interface {
	ServiceCreator
	ServiceChecker

	// CreateServiceAccountResource creates a service account resource for the service
	// This method should create a Kubernetes service account with appropriate permissions
	CreateServiceAccountResource(serviceName string, completionData *builds.BuildCompletionEventData) *unstructured.Unstructured

	// CreateConfigMapResource creates a config map resource for the service
	// This method should create a Kubernetes config map with service configuration
	CreateConfigMapResource(serviceName string, completionData *builds.BuildCompletionEventData) *unstructured.Unstructured

	// CreateKnativeServiceResource creates a Knative service resource
	// This method should create the main Knative service resource with appropriate configuration
	CreateKnativeServiceResource(serviceName string, completionData *builds.BuildCompletionEventData) *unstructured.Unstructured

	// CreateTriggerResource creates a trigger resource for the service
	// This method should create a Knative trigger for handling incoming events
	CreateTriggerResource(serviceName string, completionData *builds.BuildCompletionEventData) *unstructured.Unstructured

	// ApplyResource applies a Kubernetes resource to the cluster
	// This method should handle the creation or update of Kubernetes resources
	ApplyResource(ctx context.Context, obj *unstructured.Unstructured) error

	// DeleteService deletes a Knative service and all associated resources
	// This method should remove the service, trigger, service account, and config map
	DeleteService(ctx context.Context, serviceName string) error
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📦 BUILD CONTEXT INTERFACES - "Build context creation and validation"  │
// └─────────────────────────────────────────────────────────────────────────┘

// BuildContextCreator handles build context creation operations
// This interface provides the contract for creating build contexts for Lambda functions
type BuildContextCreator interface {
	// CreateBuildContext creates a build context for a Lambda function
	// This method should create the complete build context including source code,
	// dependencies, and configuration files, and return the S3 key where it was uploaded
	CreateBuildContext(ctx context.Context, buildRequest *builds.BuildRequest) (string, error)

	// CreateBuildContextArchive creates a build context archive
	// This method should create a compressed archive of the build context
	// including parser code and all necessary files
	CreateBuildContextArchive(ctx context.Context, out io.Writer, buildRequest *builds.BuildRequest, parserFiles map[string][]byte) error
}

// BuildContextValidator handles build context validation operations
// This interface provides the contract for validating build requests and contexts
type BuildContextValidator interface {
	// ValidateBuildRequest validates a build request
	// This method should validate that the build request contains all required
	// information and meets business rules
	ValidateBuildRequest(buildRequest *builds.BuildRequest) error
}

// BuildContextManager handles build context lifecycle operations
// This interface combines build context creation and validation operations
type BuildContextManager interface {
	BuildContextCreator
	BuildContextValidator
}

// EventHandler handles the main event processing logic
// This interface provides the contract for processing CloudEvents and managing the build pipeline
type EventHandler interface {
	// ProcessCloudEvent processes an incoming CloudEvent and returns a response
	// This method should handle the complete event processing pipeline including
	// validation, parsing, build context creation, and job management
	ProcessCloudEvent(ctx context.Context, event *cloudevents.Event) (*builds.HandlerResponse, error)

	// ValidateEvent validates a CloudEvent according to the CloudEvents specification
	ValidateEvent(ctx context.Context, event *cloudevents.Event) error

	// ParseBuildRequest parses a build request from a CloudEvent
	ParseBuildRequest(ctx context.Context, event *cloudevents.Event) (*builds.BuildRequest, error)

	// IsSupportedEventType checks if an event type is supported by the service
	IsSupportedEventType(eventType string) bool
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  ⚡ ASYNC JOB CREATOR INTERFACE - "Parallel job creation operations"     │
// └─────────────────────────────────────────────────────────────────────────┘

// AsyncJobCreatorInterface handles parallel job creation operations
// This interface provides the contract for creating Kubernetes jobs asynchronously
// with a worker pool pattern for improved throughput
type AsyncJobCreatorInterface interface {
	// CreateJobAsync creates a job asynchronously and returns immediately
	// This method queues the job creation request and returns the job name
	CreateJobAsync(ctx context.Context, buildRequest *builds.BuildRequest) (string, error)

	// GetJobCreationResult retrieves the result of a job creation attempt
	// This method returns the result if available, or false if not yet completed
	GetJobCreationResult(correlationID string) (*JobCreationResult, bool)

	// WaitForJobCreation waits for a job creation to complete
	// This method blocks until the job creation is complete or context is cancelled
	WaitForJobCreation(ctx context.Context, correlationID string) (*JobCreationResult, error)

	// GetStats returns statistics about the async job creator
	// This method provides metrics about worker count, queue size, and results
	GetStats() map[string]interface{}

	// Shutdown gracefully shuts down the async job creator
	// This method stops all workers and cleans up resources
	Shutdown(ctx context.Context) error
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔗 COMPONENT CONTAINER INTERFACE - "Dependency injection container"    │
// └─────────────────────────────────────────────────────────────────────────┘

// ComponentContainer manages all service components and their dependencies
// This interface provides the contract for the dependency injection container
// that manages all service components and their lifecycle
type ComponentContainer interface {
	// HTTP and Request Handling Components
	GetHTTPHandler() HTTPHandler
	GetCloudEventHandler() CloudEventHandler

	// Job Management Components
	GetJobManager() JobManager
	GetAsyncJobCreator() AsyncJobCreatorInterface

	// Event Processing Components
	GetEventHandler() EventHandler

	// Service Management Components
	GetServiceManager() ServiceManager

	// Build Context Components
	GetBuildContextManager() BuildContextManager

	// Core Service Components
	GetConfig() *config.Config
	GetObservability() *observability.Observability

	// Lifecycle Management
	Shutdown(ctx context.Context) error
}
