package events

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"

	lambdav1alpha1 "github.com/brunovlucena/knative-lambda-operator/api/v1alpha1"
	"github.com/brunovlucena/knative-lambda-operator/internal/metrics"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//  🌐 CLOUDEVENTS TYPE CONSTANTS
//
//  Format: io.knative.lambda.<category>.<entity>.<action>
//  Categories: command, lifecycle, invoke, response, notification
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

const (
	// Event type prefix
	EventTypePrefix = "io.knative.lambda"

	// ┌─────────────────────────────────────────────────────────────────────┐
	// │  📤 COMMAND EVENTS - Requests for actions (present tense)          │
	// └─────────────────────────────────────────────────────────────────────┘

	// Build commands
	EventTypeCommandBuildStart  = "io.knative.lambda.command.build.start"
	EventTypeCommandBuildCancel = "io.knative.lambda.command.build.cancel"
	EventTypeCommandBuildRetry  = "io.knative.lambda.command.build.retry"

	// Service commands
	EventTypeCommandServiceCreate = "io.knative.lambda.command.service.create"
	EventTypeCommandServiceUpdate = "io.knative.lambda.command.service.update"
	EventTypeCommandServiceDelete = "io.knative.lambda.command.service.delete"

	// Function commands
	EventTypeCommandFunctionDeploy   = "io.knative.lambda.command.function.deploy"
	EventTypeCommandFunctionRollback = "io.knative.lambda.command.function.rollback"

	// ┌─────────────────────────────────────────────────────────────────────┐
	// │  📊 LIFECYCLE EVENTS - State changes (past tense)                  │
	// └─────────────────────────────────────────────────────────────────────┘

	// Function lifecycle
	EventTypeLifecycleFunctionCreated  = "io.knative.lambda.lifecycle.function.created"
	EventTypeLifecycleFunctionUpdated  = "io.knative.lambda.lifecycle.function.updated"
	EventTypeLifecycleFunctionDeleted  = "io.knative.lambda.lifecycle.function.deleted"
	EventTypeLifecycleFunctionReady    = "io.knative.lambda.lifecycle.function.ready"
	EventTypeLifecycleFunctionDegraded = "io.knative.lambda.lifecycle.function.degraded"

	// Build lifecycle
	EventTypeLifecycleBuildStarted     = "io.knative.lambda.lifecycle.build.started"
	EventTypeLifecycleBuildProgressing = "io.knative.lambda.lifecycle.build.progressing"
	EventTypeLifecycleBuildCompleted   = "io.knative.lambda.lifecycle.build.completed"
	EventTypeLifecycleBuildFailed      = "io.knative.lambda.lifecycle.build.failed"
	EventTypeLifecycleBuildTimeout     = "io.knative.lambda.lifecycle.build.timeout"
	EventTypeLifecycleBuildCancelled   = "io.knative.lambda.lifecycle.build.cancelled"

	// Service lifecycle
	EventTypeLifecycleServiceCreated = "io.knative.lambda.lifecycle.service.created"
	EventTypeLifecycleServiceUpdated = "io.knative.lambda.lifecycle.service.updated"
	EventTypeLifecycleServiceDeleted = "io.knative.lambda.lifecycle.service.deleted"
	EventTypeLifecycleServiceReady   = "io.knative.lambda.lifecycle.service.ready"
	EventTypeLifecycleServiceScaled  = "io.knative.lambda.lifecycle.service.scaled"

	// ┌─────────────────────────────────────────────────────────────────────┐
	// │  🚀 INVOKE EVENTS - Trigger lambda execution                       │
	// └─────────────────────────────────────────────────────────────────────┘

	EventTypeInvokeSync      = "io.knative.lambda.invoke.sync"
	EventTypeInvokeAsync     = "io.knative.lambda.invoke.async"
	EventTypeInvokeScheduled = "io.knative.lambda.invoke.scheduled"
	EventTypeInvokeRetry     = "io.knative.lambda.invoke.retry"

	// ┌─────────────────────────────────────────────────────────────────────┐
	// │  📨 RESPONSE EVENTS - Lambda execution results                     │
	// └─────────────────────────────────────────────────────────────────────┘

	EventTypeResponseSuccess = "io.knative.lambda.response.success"
	EventTypeResponseError   = "io.knative.lambda.response.error"
	EventTypeResponseTimeout = "io.knative.lambda.response.timeout"

	// ┌─────────────────────────────────────────────────────────────────────┐
	// │  🔔 NOTIFICATION EVENTS - Alerts and audit                         │
	// └─────────────────────────────────────────────────────────────────────┘

	EventTypeNotificationAlertCritical = "io.knative.lambda.notification.alert.critical"
	EventTypeNotificationAlertWarning  = "io.knative.lambda.notification.alert.warning"
	EventTypeNotificationAlertInfo     = "io.knative.lambda.notification.alert.info"
	EventTypeNotificationAuditAccess   = "io.knative.lambda.notification.audit.access"
	EventTypeNotificationAuditChange   = "io.knative.lambda.notification.audit.change"

	// ┌─────────────────────────────────────────────────────────────────────┐
	// │  🔧 LEGACY ALIASES - For backward compatibility                    │
	// └─────────────────────────────────────────────────────────────────────┘

	// Deprecated: Use EventTypeLifecycleFunctionCreated
	EventTypeFunctionCreated = EventTypeLifecycleFunctionCreated
	// Deprecated: Use EventTypeLifecycleFunctionUpdated
	EventTypeFunctionUpdated = EventTypeLifecycleFunctionUpdated
	// Deprecated: Use EventTypeLifecycleFunctionDeleted
	EventTypeFunctionDeleted = EventTypeLifecycleFunctionDeleted
	// Deprecated: Use EventTypeLifecycleBuildStarted
	EventTypeBuildStarted = EventTypeLifecycleBuildStarted
	// Deprecated: Use EventTypeLifecycleBuildCompleted
	EventTypeBuildCompleted = EventTypeLifecycleBuildCompleted
	// Deprecated: Use EventTypeLifecycleBuildFailed
	EventTypeBuildFailed = EventTypeLifecycleBuildFailed
	// Deprecated: Use EventTypeLifecycleBuildTimeout
	EventTypeBuildTimeout = EventTypeLifecycleBuildTimeout
	// Deprecated: Use EventTypeLifecycleBuildCancelled
	EventTypeBuildCancelled = EventTypeLifecycleBuildCancelled
	// Deprecated: Use EventTypeLifecycleBuildCancelled
	EventTypeBuildStopped = EventTypeLifecycleBuildCancelled
	// Deprecated: Use EventTypeLifecycleServiceCreated
	EventTypeServiceCreated = EventTypeLifecycleServiceCreated
	// Deprecated: Use EventTypeLifecycleServiceUpdated
	EventTypeServiceUpdated = EventTypeLifecycleServiceUpdated
	// Deprecated: Use EventTypeLifecycleServiceDeleted
	EventTypeServiceDeleted = EventTypeLifecycleServiceDeleted
	// Deprecated: Use EventTypeLifecycleFunctionUpdated
	EventTypeStatusUpdated = EventTypeLifecycleFunctionUpdated
	// Deprecated: Use EventTypeNotificationAlertInfo
	EventTypeHealthCheck = EventTypeNotificationAlertInfo
	// Deprecated: Use EventTypeInvokeAsync
	EventTypeParserStarted = EventTypeInvokeAsync
	// Deprecated: Use EventTypeResponseSuccess
	EventTypeParserCompleted = EventTypeResponseSuccess
	// Deprecated: Use EventTypeResponseError
	EventTypeParserFailed = EventTypeResponseError

	// Default CloudEvents source format
	DefaultSourceFormat = "io.knative.lambda/operator"

	// Default broker URL
	DefaultBrokerURL = "http://knative-lambda-broker-broker-ingress.knative-lambda.svc.cluster.local"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//  📦 EVENT DATA STRUCTURES
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// Config holds event manager configuration
type Config struct {
	BrokerURL string
	Source    string
	Enabled   bool
}

// Manager handles CloudEvents emission
type Manager struct {
	config     Config
	httpClient *http.Client
}

// NewManager creates a new event manager
func NewManager(config Config) *Manager {
	if config.BrokerURL == "" {
		config.BrokerURL = DefaultBrokerURL
	}
	if config.Source == "" {
		config.Source = DefaultSourceFormat
	}

	return &Manager{
		config: config,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📊 FUNCTION EVENT DATA                                                │
// └─────────────────────────────────────────────────────────────────────────┘

// FunctionEventData represents data for function lifecycle events
type FunctionEventData struct {
	Name               string          `json:"name"`
	Namespace          string          `json:"namespace"`
	Runtime            *RuntimeData    `json:"runtime,omitempty"`
	Phase              string          `json:"phase,omitempty"`
	Conditions         []ConditionData `json:"conditions,omitempty"`
	Generation         int64           `json:"generation,omitempty"`
	ObservedGeneration int64           `json:"observedGeneration,omitempty"`
}

// RuntimeData represents runtime information
type RuntimeData struct {
	Language string `json:"language,omitempty"`
	Version  string `json:"version,omitempty"`
	Handler  string `json:"handler,omitempty"`
}

// ConditionData represents a condition
type ConditionData struct {
	Type               string `json:"type"`
	Status             string `json:"status"`
	Reason             string `json:"reason,omitempty"`
	Message            string `json:"message,omitempty"`
	LastTransitionTime string `json:"lastTransitionTime,omitempty"`
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔨 BUILD EVENT DATA                                                   │
// └─────────────────────────────────────────────────────────────────────────┘

// BuildEventData represents data for build lifecycle events
type BuildEventData struct {
	Name        string     `json:"name"`
	Namespace   string     `json:"namespace"`
	JobName     string     `json:"jobName,omitempty"`
	BuildID     string     `json:"buildId,omitempty"`
	ImageURI    string     `json:"imageUri,omitempty"`
	Digest      string     `json:"digest,omitempty"`
	StartedAt   string     `json:"startedAt,omitempty"`
	CompletedAt string     `json:"completedAt,omitempty"`
	Duration    string     `json:"duration,omitempty"`
	Phase       string     `json:"phase,omitempty"`
	Error       *ErrorData `json:"error,omitempty"`
	LogsURL     string     `json:"logsUrl,omitempty"`
}

// ErrorData represents error information
type ErrorData struct {
	Type      string `json:"type,omitempty"`
	Message   string `json:"message"`
	Code      string `json:"code,omitempty"`
	Retryable bool   `json:"retryable,omitempty"`
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🚀 SERVICE EVENT DATA                                                 │
// └─────────────────────────────────────────────────────────────────────────┘

// ServiceEventData represents data for service lifecycle events
type ServiceEventData struct {
	Name                string        `json:"name"`
	Namespace           string        `json:"namespace"`
	ServiceName         string        `json:"serviceName,omitempty"`
	URL                 string        `json:"url,omitempty"`
	LatestRevision      string        `json:"latestRevision,omitempty"`
	LatestReadyRevision string        `json:"latestReadyRevision,omitempty"`
	Ready               bool          `json:"ready,omitempty"`
	Replicas            *ReplicasData `json:"replicas,omitempty"`
	Traffic             []TrafficData `json:"traffic,omitempty"`
}

// ReplicasData represents replica information
type ReplicasData struct {
	Desired   int32 `json:"desired,omitempty"`
	Ready     int32 `json:"ready,omitempty"`
	Available int32 `json:"available,omitempty"`
}

// TrafficData represents traffic split information
type TrafficData struct {
	RevisionName   string `json:"revisionName,omitempty"`
	Percent        int64  `json:"percent,omitempty"`
	LatestRevision bool   `json:"latestRevision,omitempty"`
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📨 INVOKE/RESPONSE EVENT DATA                                         │
// └─────────────────────────────────────────────────────────────────────────┘

// InvokeEventData represents data for invoke events
type InvokeEventData struct {
	FunctionName  string                 `json:"functionName"`
	Namespace     string                 `json:"namespace"`
	InvocationID  string                 `json:"invocationId"`
	CorrelationID string                 `json:"correlationId,omitempty"`
	Payload       map[string]interface{} `json:"payload,omitempty"`
	Metadata      *InvokeMetadata        `json:"metadata,omitempty"`
}

// InvokeMetadata represents invocation metadata
type InvokeMetadata struct {
	TraceID    string `json:"traceId,omitempty"`
	SpanID     string `json:"spanId,omitempty"`
	RetryCount int    `json:"retryCount,omitempty"`
	DeadlineAt string `json:"deadlineAt,omitempty"`
}

// ResponseEventData represents data for response events
type ResponseEventData struct {
	FunctionName  string           `json:"functionName"`
	Namespace     string           `json:"namespace"`
	InvocationID  string           `json:"invocationId"`
	CorrelationID string           `json:"correlationId,omitempty"`
	Result        *ResultData      `json:"result,omitempty"`
	Error         *ErrorData       `json:"error,omitempty"`
	Metrics       *ResponseMetrics `json:"metrics,omitempty"`
	DLQ           *DLQData         `json:"dlq,omitempty"`
}

// ResultData represents function execution result
type ResultData struct {
	StatusCode int                    `json:"statusCode"`
	Body       map[string]interface{} `json:"body,omitempty"`
}

// ResponseMetrics represents response metrics
type ResponseMetrics struct {
	DurationMs   int64 `json:"durationMs"`
	ColdStart    bool  `json:"coldStart,omitempty"`
	MemoryUsedMb int64 `json:"memoryUsedMb,omitempty"`
}

// DLQData represents DLQ routing information
type DLQData struct {
	Routed    bool   `json:"routed"`
	QueueName string `json:"queueName,omitempty"`
	RoutedAt  string `json:"routedAt,omitempty"`
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔔 NOTIFICATION EVENT DATA                                            │
// └─────────────────────────────────────────────────────────────────────────┘

// AlertEventData represents data for alert events
type AlertEventData struct {
	AlertName   string            `json:"alertName"`
	Severity    string            `json:"severity"`
	Summary     string            `json:"summary"`
	Description string            `json:"description,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	StartsAt    string            `json:"startsAt,omitempty"`
	EndsAt      string            `json:"endsAt,omitempty"`
}

// AuditEventData represents data for audit events
type AuditEventData struct {
	Action       string            `json:"action"`
	Resource     string            `json:"resource"`
	ResourceName string            `json:"resourceName"`
	Namespace    string            `json:"namespace"`
	User         string            `json:"user,omitempty"`
	Reason       string            `json:"reason,omitempty"`
	Details      map[string]string `json:"details,omitempty"`
	Timestamp    string            `json:"timestamp"`
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//  📤 EMIT FUNCTIONS - LIFECYCLE EVENTS
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// EmitFunctionCreated emits a lifecycle.function.created event
func (m *Manager) EmitFunctionCreated(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction) error {
	data := m.buildFunctionEventData(lambda)
	return m.emit(ctx, EventTypeLifecycleFunctionCreated, lambda, data)
}

// EmitFunctionUpdated emits a lifecycle.function.updated event
func (m *Manager) EmitFunctionUpdated(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction) error {
	data := m.buildFunctionEventData(lambda)
	return m.emit(ctx, EventTypeLifecycleFunctionUpdated, lambda, data)
}

// EmitFunctionDeleted emits a lifecycle.function.deleted event
func (m *Manager) EmitFunctionDeleted(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction) error {
	data := &FunctionEventData{
		Name:      lambda.Name,
		Namespace: lambda.Namespace,
	}
	return m.emit(ctx, EventTypeLifecycleFunctionDeleted, lambda, data)
}

// EmitFunctionReady emits a lifecycle.function.ready event
func (m *Manager) EmitFunctionReady(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction) error {
	data := m.buildFunctionEventData(lambda)
	return m.emit(ctx, EventTypeLifecycleFunctionReady, lambda, data)
}

// EmitFunctionDegraded emits a lifecycle.function.degraded event
func (m *Manager) EmitFunctionDegraded(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction, reason string) error {
	data := m.buildFunctionEventData(lambda)
	// Add degraded condition
	data.Conditions = append(data.Conditions, ConditionData{
		Type:               "Degraded",
		Status:             "True",
		Reason:             reason,
		LastTransitionTime: time.Now().Format(time.RFC3339),
	})
	return m.emit(ctx, EventTypeLifecycleFunctionDegraded, lambda, data)
}

// EmitBuildStarted emits a lifecycle.build.started event
func (m *Manager) EmitBuildStarted(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction, jobName string) error {
	data := &BuildEventData{
		Name:      lambda.Name,
		Namespace: lambda.Namespace,
		JobName:   jobName,
		BuildID:   fmt.Sprintf("build-%s-%d", lambda.Name, time.Now().Unix()),
		StartedAt: time.Now().Format(time.RFC3339),
		Phase:     "Started",
	}
	return m.emit(ctx, EventTypeLifecycleBuildStarted, lambda, data)
}

// EmitBuildProgressing emits a lifecycle.build.progressing event
func (m *Manager) EmitBuildProgressing(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction) error {
	data := m.buildBuildEventData(lambda)
	data.Phase = "Progressing"
	return m.emit(ctx, EventTypeLifecycleBuildProgressing, lambda, data)
}

// EmitBuildCompleted emits a lifecycle.build.completed event
func (m *Manager) EmitBuildCompleted(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction, imageURI string) error {
	data := m.buildBuildEventData(lambda)
	data.ImageURI = imageURI
	data.CompletedAt = time.Now().Format(time.RFC3339)
	data.Phase = "Completed"

	if lambda.Status.BuildStatus != nil && lambda.Status.BuildStatus.StartedAt != nil {
		dur := time.Since(lambda.Status.BuildStatus.StartedAt.Time)
		data.Duration = dur.String()
	}

	return m.emit(ctx, EventTypeLifecycleBuildCompleted, lambda, data)
}

// EmitBuildFailed emits a lifecycle.build.failed event
func (m *Manager) EmitBuildFailed(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction, errMsg string) error {
	data := m.buildBuildEventData(lambda)
	data.CompletedAt = time.Now().Format(time.RFC3339)
	data.Phase = "Failed"
	data.Error = &ErrorData{
		Type:      "BuildError",
		Message:   errMsg,
		Code:      "BUILD_FAILED",
		Retryable: true,
	}

	if lambda.Status.BuildStatus != nil && lambda.Status.BuildStatus.StartedAt != nil {
		dur := time.Since(lambda.Status.BuildStatus.StartedAt.Time)
		data.Duration = dur.String()
	}

	return m.emit(ctx, EventTypeLifecycleBuildFailed, lambda, data)
}

// EmitBuildTimeout emits a lifecycle.build.timeout event
func (m *Manager) EmitBuildTimeout(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction) error {
	data := m.buildBuildEventData(lambda)
	data.CompletedAt = time.Now().Format(time.RFC3339)
	data.Phase = "Timeout"
	data.Error = &ErrorData{
		Type:      "TimeoutError",
		Message:   "Build timeout exceeded",
		Code:      "BUILD_TIMEOUT",
		Retryable: true,
	}

	return m.emit(ctx, EventTypeLifecycleBuildTimeout, lambda, data)
}

// EmitBuildCancelled emits a lifecycle.build.cancelled event
func (m *Manager) EmitBuildCancelled(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction, reason string) error {
	data := m.buildBuildEventData(lambda)
	data.CompletedAt = time.Now().Format(time.RFC3339)
	data.Phase = "Cancelled"
	data.Error = &ErrorData{
		Type:      "CancellationError",
		Message:   reason,
		Code:      "BUILD_CANCELLED",
		Retryable: false,
	}

	return m.emit(ctx, EventTypeLifecycleBuildCancelled, lambda, data)
}

// EmitServiceCreated emits a lifecycle.service.created event
func (m *Manager) EmitServiceCreated(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction, serviceName string) error {
	data := &ServiceEventData{
		Name:        lambda.Name,
		Namespace:   lambda.Namespace,
		ServiceName: serviceName,
	}
	return m.emit(ctx, EventTypeLifecycleServiceCreated, lambda, data)
}

// EmitServiceUpdated emits a lifecycle.service.updated event
func (m *Manager) EmitServiceUpdated(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction) error {
	data := m.buildServiceEventData(lambda)
	return m.emit(ctx, EventTypeLifecycleServiceUpdated, lambda, data)
}

// EmitServiceDeleted emits a lifecycle.service.deleted event
func (m *Manager) EmitServiceDeleted(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction) error {
	data := &ServiceEventData{
		Name:      lambda.Name,
		Namespace: lambda.Namespace,
	}
	if lambda.Status.ServiceStatus != nil {
		data.ServiceName = lambda.Status.ServiceStatus.ServiceName
	}
	return m.emit(ctx, EventTypeLifecycleServiceDeleted, lambda, data)
}

// EmitServiceReady emits a lifecycle.service.ready event
func (m *Manager) EmitServiceReady(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction) error {
	data := m.buildServiceEventData(lambda)
	data.Ready = true
	return m.emit(ctx, EventTypeLifecycleServiceReady, lambda, data)
}

// EmitServiceScaled emits a lifecycle.service.scaled event
func (m *Manager) EmitServiceScaled(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction, replicas int32) error {
	data := m.buildServiceEventData(lambda)
	data.Replicas = &ReplicasData{
		Desired: replicas,
	}
	return m.emit(ctx, EventTypeLifecycleServiceScaled, lambda, data)
}

// EmitStatusUpdated emits a status update event (deprecated - use specific events)
func (m *Manager) EmitStatusUpdated(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction) error {
	return m.EmitFunctionUpdated(ctx, lambda)
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//  📨 EMIT FUNCTIONS - INVOKE/RESPONSE EVENTS
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// EmitInvokeAsync emits an invoke.async event
func (m *Manager) EmitInvokeAsync(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction, invocationID, correlationID string, payload map[string]interface{}) error {
	data := &InvokeEventData{
		FunctionName:  lambda.Name,
		Namespace:     lambda.Namespace,
		InvocationID:  invocationID,
		CorrelationID: correlationID,
		Payload:       payload,
		Metadata: &InvokeMetadata{
			RetryCount: 0,
			DeadlineAt: time.Now().Add(5 * time.Minute).Format(time.RFC3339),
		},
	}
	return m.emit(ctx, EventTypeInvokeAsync, lambda, data)
}

// EmitResponseSuccess emits a response.success event
func (m *Manager) EmitResponseSuccess(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction, invocationID, correlationID string, result map[string]interface{}, durationMs int64) error {
	data := &ResponseEventData{
		FunctionName:  lambda.Name,
		Namespace:     lambda.Namespace,
		InvocationID:  invocationID,
		CorrelationID: correlationID,
		Result: &ResultData{
			StatusCode: 200,
			Body:       result,
		},
		Metrics: &ResponseMetrics{
			DurationMs: durationMs,
		},
	}
	return m.emit(ctx, EventTypeResponseSuccess, lambda, data)
}

// EmitResponseError emits a response.error event
func (m *Manager) EmitResponseError(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction, invocationID, correlationID, errMsg string, retryable bool) error {
	data := &ResponseEventData{
		FunctionName:  lambda.Name,
		Namespace:     lambda.Namespace,
		InvocationID:  invocationID,
		CorrelationID: correlationID,
		Error: &ErrorData{
			Type:      "RuntimeError",
			Message:   errMsg,
			Code:      "RUNTIME_ERROR",
			Retryable: retryable,
		},
	}

	if retryable {
		data.DLQ = &DLQData{
			Routed:    true,
			QueueName: fmt.Sprintf("%s-dlq", lambda.Name),
			RoutedAt:  time.Now().Format(time.RFC3339),
		}
	}

	return m.emit(ctx, EventTypeResponseError, lambda, data)
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//  🔔 EMIT FUNCTIONS - NOTIFICATION EVENTS
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// EmitAlert emits an alert notification event
func (m *Manager) EmitAlert(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction, severity, alertName, summary, description string, labels map[string]string) error {
	var eventType string
	switch severity {
	case "critical":
		eventType = EventTypeNotificationAlertCritical
	case "warning":
		eventType = EventTypeNotificationAlertWarning
	default:
		eventType = EventTypeNotificationAlertInfo
	}

	data := &AlertEventData{
		AlertName:   alertName,
		Severity:    severity,
		Summary:     summary,
		Description: description,
		Labels:      labels,
		StartsAt:    time.Now().Format(time.RFC3339),
	}

	return m.emit(ctx, eventType, lambda, data)
}

// EmitAuditChange emits an audit.change notification event
func (m *Manager) EmitAuditChange(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction, action, reason string, details map[string]string) error {
	data := &AuditEventData{
		Action:       action,
		Resource:     "LambdaFunction",
		ResourceName: lambda.Name,
		Namespace:    lambda.Namespace,
		Reason:       reason,
		Details:      details,
		Timestamp:    time.Now().Format(time.RFC3339),
	}

	return m.emit(ctx, EventTypeNotificationAuditChange, lambda, data)
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//  🔧 HELPER FUNCTIONS
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// buildFunctionEventData creates FunctionEventData from a LambdaFunction
func (m *Manager) buildFunctionEventData(lambda *lambdav1alpha1.LambdaFunction) *FunctionEventData {
	data := &FunctionEventData{
		Name:               lambda.Name,
		Namespace:          lambda.Namespace,
		Phase:              string(lambda.Status.Phase),
		Generation:         lambda.Generation,
		ObservedGeneration: lambda.Status.ObservedGeneration,
	}

	if lambda.Spec.Runtime.Language != "" {
		data.Runtime = &RuntimeData{
			Language: lambda.Spec.Runtime.Language,
			Version:  lambda.Spec.Runtime.Version,
			Handler:  lambda.Spec.Runtime.Handler,
		}
	}

	// Convert conditions
	for _, cond := range lambda.Status.Conditions {
		data.Conditions = append(data.Conditions, ConditionData{
			Type:               string(cond.Type),
			Status:             string(cond.Status),
			Reason:             cond.Reason,
			Message:            cond.Message,
			LastTransitionTime: cond.LastTransitionTime.Format(time.RFC3339),
		})
	}

	return data
}

// buildBuildEventData creates BuildEventData from a LambdaFunction
func (m *Manager) buildBuildEventData(lambda *lambdav1alpha1.LambdaFunction) *BuildEventData {
	data := &BuildEventData{
		Name:      lambda.Name,
		Namespace: lambda.Namespace,
	}

	if lambda.Status.BuildStatus != nil {
		data.JobName = lambda.Status.BuildStatus.JobName
		data.ImageURI = lambda.Status.BuildStatus.ImageURI
		if lambda.Status.BuildStatus.StartedAt != nil {
			data.StartedAt = lambda.Status.BuildStatus.StartedAt.Format(time.RFC3339)
		}
	}

	return data
}

// buildServiceEventData creates ServiceEventData from a LambdaFunction
func (m *Manager) buildServiceEventData(lambda *lambdav1alpha1.LambdaFunction) *ServiceEventData {
	data := &ServiceEventData{
		Name:      lambda.Name,
		Namespace: lambda.Namespace,
	}

	if lambda.Status.ServiceStatus != nil {
		data.ServiceName = lambda.Status.ServiceStatus.ServiceName
		data.URL = lambda.Status.ServiceStatus.URL
		data.Ready = lambda.Status.ServiceStatus.Ready
		data.LatestRevision = lambda.Status.ServiceStatus.LatestRevision
	}

	return data
}

// getSource returns the CloudEvents source for a lambda
func (m *Manager) getSource(lambda *lambdav1alpha1.LambdaFunction) string {
	return fmt.Sprintf("%s/%s/%s", m.config.Source, lambda.Namespace, lambda.Name)
}

// emit sends a CloudEvent to the broker
func (m *Manager) emit(ctx context.Context, eventType string, lambda *lambdav1alpha1.LambdaFunction, data interface{}) error {
	// Record metrics for build and service events
	m.recordEventMetrics(eventType)

	if !m.config.Enabled {
		return nil
	}

	// Create CloudEvent
	event := cloudevents.NewEvent()
	event.SetID(uuid.New().String())
	event.SetType(eventType)
	event.SetSource(m.getSource(lambda))
	event.SetSubject(fmt.Sprintf("%s/%s", lambda.Namespace, lambda.Name))
	event.SetTime(time.Now())

	// Add extension attributes for correlation
	event.SetExtension("correlationid", uuid.New().String())

	if err := event.SetData(cloudevents.ApplicationJSON, data); err != nil {
		return fmt.Errorf("failed to set event data: %w", err)
	}

	// Serialize event
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Send to broker
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.config.BrokerURL, bytes.NewReader(eventBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/cloudevents+json")
	req.Header.Set("Ce-Id", event.ID())
	req.Header.Set("Ce-Type", event.Type())
	req.Header.Set("Ce-Source", event.Source())
	req.Header.Set("Ce-Specversion", "1.0")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("broker returned status %d", resp.StatusCode)
	}

	return nil
}

// recordEventMetrics records Prometheus metrics for build and service lifecycle events
func (m *Manager) recordEventMetrics(eventType string) {
	// Build events - match exact event types
	if strings.HasPrefix(eventType, "io.knative.lambda.lifecycle.build.") {
		var status string
		switch eventType {
		case EventTypeLifecycleBuildStarted:
			status = "start"
		case EventTypeLifecycleBuildCompleted:
			status = "complete"
		case EventTypeLifecycleBuildFailed:
			status = "failed"
		case EventTypeLifecycleBuildTimeout:
			status = "timeout"
		case EventTypeLifecycleBuildCancelled:
			status = "cancel"
		case EventTypeLifecycleBuildProgressing:
			status = "progressing"
		default:
			// Handle deprecated aliases
			if eventType == EventTypeBuildStarted || eventType == EventTypeBuildCompleted ||
				eventType == EventTypeBuildFailed || eventType == EventTypeBuildTimeout ||
				eventType == EventTypeBuildCancelled || eventType == EventTypeBuildStopped {
				// Map deprecated types to their current equivalents
				if eventType == EventTypeBuildStarted {
					status = "start"
				} else if eventType == EventTypeBuildCompleted {
					status = "complete"
				} else if eventType == EventTypeBuildFailed {
					status = "failed"
				} else if eventType == EventTypeBuildTimeout {
					status = "timeout"
				} else if eventType == EventTypeBuildCancelled || eventType == EventTypeBuildStopped {
					status = "cancel"
				}
			} else {
				return // Unknown build event type
			}
		}
		metrics.BuildEventsTotal.WithLabelValues(status).Inc()
		return
	}

	// Service events - match exact event types
	if strings.HasPrefix(eventType, "io.knative.lambda.lifecycle.service.") {
		var status string
		switch eventType {
		case EventTypeLifecycleServiceCreated:
			status = "create"
		case EventTypeLifecycleServiceUpdated:
			status = "update"
		case EventTypeLifecycleServiceDeleted:
			status = "delete"
		case EventTypeLifecycleServiceReady:
			status = "ready"
		case EventTypeLifecycleServiceScaled:
			status = "scaled"
		default:
			// Handle deprecated aliases
			if eventType == EventTypeServiceCreated || eventType == EventTypeServiceUpdated ||
				eventType == EventTypeServiceDeleted {
				if eventType == EventTypeServiceCreated {
					status = "create"
				} else if eventType == EventTypeServiceUpdated {
					status = "update"
				} else if eventType == EventTypeServiceDeleted {
					status = "delete"
				}
			} else {
				return // Unknown service event type
			}
		}
		metrics.ServiceEventsTotal.WithLabelValues(status).Inc()
		return
	}

	// Parser events - track invoke.async (parser started), response.success (parser completed), response.error (parser failed)
	if eventType == EventTypeInvokeAsync || eventType == EventTypeParserStarted {
		metrics.ParserEventsTotal.WithLabelValues("start").Inc()
		return
	}
	if eventType == EventTypeResponseSuccess || eventType == EventTypeParserCompleted {
		metrics.ParserEventsTotal.WithLabelValues("complete").Inc()
		return
	}
	if eventType == EventTypeResponseError || eventType == EventTypeParserFailed {
		metrics.ParserEventsTotal.WithLabelValues("failed").Inc()
		return
	}
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//  📋 EVENT TYPE HELPERS
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// IsCommandEvent returns true if the event type is a command event
func IsCommandEvent(eventType string) bool {
	return len(eventType) > len(EventTypePrefix)+9 && eventType[len(EventTypePrefix)+1:len(EventTypePrefix)+8] == "command"
}

// IsLifecycleEvent returns true if the event type is a lifecycle event
func IsLifecycleEvent(eventType string) bool {
	return len(eventType) > len(EventTypePrefix)+11 && eventType[len(EventTypePrefix)+1:len(EventTypePrefix)+10] == "lifecycle"
}

// IsInvokeEvent returns true if the event type is an invoke event
func IsInvokeEvent(eventType string) bool {
	return len(eventType) > len(EventTypePrefix)+8 && eventType[len(EventTypePrefix)+1:len(EventTypePrefix)+7] == "invoke"
}

// IsResponseEvent returns true if the event type is a response event
func IsResponseEvent(eventType string) bool {
	return len(eventType) > len(EventTypePrefix)+10 && eventType[len(EventTypePrefix)+1:len(EventTypePrefix)+9] == "response"
}

// IsNotificationEvent returns true if the event type is a notification event
func IsNotificationEvent(eventType string) bool {
	return len(eventType) > len(EventTypePrefix)+14 && eventType[len(EventTypePrefix)+1:len(EventTypePrefix)+13] == "notification"
}
