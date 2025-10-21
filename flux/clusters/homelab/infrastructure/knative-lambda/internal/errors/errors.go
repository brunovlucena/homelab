// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🚨 UNIFIED ERROR HANDLING SYSTEM - Comprehensive error management
//
//	🎯 Purpose: Centralized error handling with types, constants, and utilities
//	💡 Features: Error types, constants, utilities, observability integration
//
//	📋 COMPONENTS:
//	🎯 Error Types - Structured error types with rich context
//	📝 Error Constants - Centralized error message constants
//	🔧 Error Utilities - Common error handling patterns
//	📊 Error Metrics - Error classification and monitoring
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package errors

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🎯 ERROR TYPES - "Categorized error types with rich context"          │
// └─────────────────────────────────────────────────────────────────────────┘

// 🎯 ValidationError - "Input validation errors"
type ValidationError struct {
	Field   string      `json:"field"`
	Value   interface{} `json:"value"`
	Reason  string      `json:"reason"`
	Context string      `json:"context,omitempty"`
	Time    time.Time   `json:"time"`
}

func (e *ValidationError) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("validation failed for field '%s' with value '%v' in context '%s': %s", e.Field, e.Value, e.Context, e.Reason)
	}
	return fmt.Sprintf("validation failed for field '%s' with value '%v': %s", e.Field, e.Value, e.Reason)
}

func (e *ValidationError) Is(target error) bool {
	_, ok := target.(*ValidationError)
	return ok
}

// 🎯 ConfigurationError - "Configuration and setup errors"
type ConfigurationError struct {
	Component string      `json:"component"`
	Setting   string      `json:"setting"`
	Value     interface{} `json:"value,omitempty"`
	Reason    string      `json:"reason"`
	Time      time.Time   `json:"time"`
}

func (e *ConfigurationError) Error() string {
	if e.Value != nil {
		return fmt.Sprintf("configuration error in component '%s' for setting '%s' with value '%v': %s", e.Component, e.Setting, e.Value, e.Reason)
	}
	return fmt.Sprintf("configuration error in component '%s' for setting '%s': %s", e.Component, e.Setting, e.Reason)
}

func (e *ConfigurationError) Is(target error) bool {
	_, ok := target.(*ConfigurationError)
	return ok
}

// 🎯 ConnectionError - "Network and connection errors"
type ConnectionError struct {
	Service   string        `json:"service"`
	Endpoint  string        `json:"endpoint,omitempty"`
	Operation string        `json:"operation"`
	Timeout   time.Duration `json:"timeout,omitempty"`
	Retries   int           `json:"retries,omitempty"`
	Reason    string        `json:"reason"`
	Time      time.Time     `json:"time"`
}

func (e *ConnectionError) Error() string {
	if e.Endpoint != "" {
		return fmt.Sprintf("connection failed to %s at %s during %s: %s", e.Service, e.Endpoint, e.Operation, e.Reason)
	}
	return fmt.Sprintf("connection failed to %s during %s: %s", e.Service, e.Operation, e.Reason)
}

func (e *ConnectionError) Is(target error) bool {
	_, ok := target.(*ConnectionError)
	return ok
}

// 🎯 TimeoutError - "Operation timeout errors"
type TimeoutError struct {
	Operation string        `json:"operation"`
	Timeout   time.Duration `json:"timeout"`
	Context   string        `json:"context,omitempty"`
	Time      time.Time     `json:"time"`
}

func (e *TimeoutError) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("operation '%s' in context '%s' timed out after %v", e.Operation, e.Context, e.Timeout)
	}
	return fmt.Sprintf("operation '%s' timed out after %v", e.Operation, e.Timeout)
}

func (e *TimeoutError) Is(target error) bool {
	_, ok := target.(*TimeoutError)
	return ok
}

// 🎯 NotFoundError - "Resource not found errors"
type NotFoundError struct {
	Resource   string    `json:"resource"`
	Identifier string    `json:"identifier"`
	Context    string    `json:"context,omitempty"`
	Time       time.Time `json:"time"`
}

func (e *NotFoundError) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("%s not found: %s in context '%s'", e.Resource, e.Identifier, e.Context)
	}
	return fmt.Sprintf("%s not found: %s", e.Resource, e.Identifier)
}

func (e *NotFoundError) Is(target error) bool {
	_, ok := target.(*NotFoundError)
	return ok
}

// 🎯 PermissionError - "Authorization and permission errors"
type PermissionError struct {
	Operation string    `json:"operation"`
	Resource  string    `json:"resource"`
	Subject   string    `json:"subject,omitempty"`
	Context   string    `json:"context,omitempty"`
	Time      time.Time `json:"time"`
}

func (e *PermissionError) Error() string {
	if e.Subject != "" && e.Context != "" {
		return fmt.Sprintf("permission denied: subject '%s' cannot %s %s in context '%s'", e.Subject, e.Operation, e.Resource, e.Context)
	}
	if e.Subject != "" {
		return fmt.Sprintf("permission denied: subject '%s' cannot %s %s", e.Subject, e.Operation, e.Resource)
	}
	if e.Context != "" {
		return fmt.Sprintf("permission denied: cannot %s %s in context '%s'", e.Operation, e.Resource, e.Context)
	}
	return fmt.Sprintf("permission denied: cannot %s %s", e.Operation, e.Resource)
}

func (e *PermissionError) Is(target error) bool {
	_, ok := target.(*PermissionError)
	return ok
}

// 🎯 SystemError - "System and infrastructure errors"
type SystemError struct {
	Component string      `json:"component"`
	Operation string      `json:"operation"`
	Code      string      `json:"code,omitempty"`
	Details   interface{} `json:"details,omitempty"`
	Time      time.Time   `json:"time"`
}

func (e *SystemError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("system error in component '%s' during operation '%s' (code: %s)", e.Component, e.Operation, e.Code)
	}
	return fmt.Sprintf("system error in component '%s' during operation '%s'", e.Component, e.Operation)
}

func (e *SystemError) Is(target error) bool {
	_, ok := target.(*SystemError)
	return ok
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📝 ERROR CONSTANTS - "Centralized error message constants"            │
// └─────────────────────────────────────────────────────────────────────────┘

// Configuration Errors
const (
	ErrConfigValidationFailed = "configuration validation failed"
	ErrConfigRequired         = "configuration required"
	ErrConfigInvalid          = "configuration invalid"

	// Port and network configuration errors
	ErrInvalidPort           = "invalid port: %d (must be between %d-%d)"
	ErrInvalidMetricsPort    = "invalid metrics port: %d (must be between %d-%d)"
	ErrInvalidNamespace      = "invalid namespace: %s (must be between 1-%d characters)"
	ErrInvalidServiceAccount = "invalid service account: %s (must be between 1-%d characters)"
	ErrInvalidSampleRate     = "invalid sample rate: %f (must be between 0-1)"

	// AWS configuration errors
	ErrAWSRegionRequired = "AWS region is required"
	ErrAWSConfigFailed   = "failed to load AWS config"

	// Kubernetes configuration errors
	ErrKubeconfigNotFound  = "kubeconfig not found: KUBECONFIG or HOME environment variable not set"
	ErrK8sConfigFailed     = "failed to create Kubernetes config"
	ErrK8sClientFailed     = "failed to create Kubernetes client"
	ErrDynamicClientFailed = "failed to create dynamic client"

	// Environment and setup errors
	ErrNamespaceNotConfigured   = "namespace not configured"
	ErrServiceNameNotConfigured = "service name not configured"

	// Sidecar configuration errors
	ErrKanikoNamespaceRequired  = "KANIKO_NAMESPACE is required"
	ErrKanikoPodNameRequired    = "KANIKO_POD_NAME is required"
	ErrBuildJobNameRequired     = "BUILD_JOB_NAME is required"
	ErrImageURIRequired         = "IMAGE_URI is required"
	ErrCorrelationIDRequired    = "CORRELATION_ID is required"
	ErrKnativeBrokerURLRequired = "KNATIVE_BROKER_URL is required"

	// Duration parsing errors
	ErrInvalidMonitorInterval = "invalid MONITOR_INTERVAL: %w"
	ErrInvalidBuildTimeout    = "invalid BUILD_TIMEOUT: %w"

	// Boolean parsing errors
	ErrInvalidTLSEnabled     = "invalid TLS_ENABLED: %w"
	ErrInvalidMetricsEnabled = "invalid METRICS_ENABLED: %w"

	// Integer parsing errors
	ErrInvalidRunAsUser  = "invalid RUN_AS_USER: %w"
	ErrInvalidRunAsGroup = "invalid RUN_AS_GROUP: %w"

	// TLS configuration errors
	ErrTLSCertAndKeyRequired = "TLS cert path and key path are required when TLS is enabled"
)

// Validation Errors
const (
	// Event validation errors
	ErrMissingCloudEventHeaders        = "missing required CloudEvent headers"
	ErrUnsupportedEventType            = "unsupported event type: %s"
	ErrInvalidEventBody                = "invalid event body"
	ErrInvalidBuildCompletionEventBody = "invalid build completion event body"
	ErrEventValidationFailed           = "event validation failed"
	ErrFailedToParseEventData          = "failed to parse event data"

	// Data validation errors
	ErrThirdPartyIDRequired = "ThirdPartyID is required in event data"
	ErrParserIDRequired     = "ParserID is required in event data"
	ErrEventDataCannotBeNil = "event data cannot be nil"

	// ID validation errors
	ErrIDCannotBeEmpty            = "ID cannot be empty"
	ErrIDTooLong                  = "ID too long (max 100 characters)"
	ErrIDContainsDangerousContent = "ID contains potentially dangerous content"
	ErrIDContainsInvalidChars     = "ID contains invalid characters"

	// Resource validation errors
	ErrResourceNameRequired    = "resource name is required"
	ErrResourceKindRequired    = "resource kind is required"
	ErrResourceVersionRequired = "resource version is required"
	ErrInvalidJobSpec          = "invalid job spec: no containers defined"
	ErrParserStreamNil         = "parser stream is nil - S3 object may be empty or corrupted"
	ErrFactoryReturnedNil      = "factory returned nil"

	// Resource quantity validation errors
	ErrCPUQuantityEmpty        = "CPU quantity is empty"
	ErrCPUQuantityNegative     = "CPU quantity cannot be negative"
	ErrCPUQuantityTooLarge     = "CPU quantity is too large (max 100 cores)"
	ErrMemoryQuantityEmpty     = "memory quantity is empty"
	ErrMemoryQuantityNegative  = "memory quantity cannot be negative"
	ErrMemoryQuantityTooLarge  = "memory quantity is too large (max 1TB)"
	ErrContainerTypeEmpty      = "container type is empty"
	ErrResourceRequirementsNil = "resource requirements cannot be nil"

	// Field validation errors
	ErrThirdPartyIDEmpty  = "ThirdPartyID is required"
	ErrParserIDEmpty      = "ParserID is required"
	ErrSourceBucketEmpty  = "SourceBucket is required"
	ErrSourceKeyEmpty     = "SourceKey is required"
	ErrCorrelationIDEmpty = "CorrelationID is required"

	// Format validation errors
	ErrInvalidNamespaceFormat      = "invalid namespace: %w"
	ErrInvalidServiceAccountFormat = "invalid service account: %w"
	ErrInvalidAWSRegionFormat      = "invalid AWS region: %w"
	ErrInvalidAWSAccountIDFormat   = "invalid AWS account ID: %w"
	ErrInvalidECRRegistryFormat    = "invalid ECR registry: %w"
	ErrThirdPartyIDTooLong         = "third party ID too long (max 63 characters)"

	// Event structure validation errors
	ErrBuildRequestNil    = "build request is nil"
	ErrCloudEventNil      = "cloud event is nil"
	ErrEventIDEmpty       = "event ID is empty"
	ErrEventTypeEmpty     = "event type is empty"
	ErrEventSourceEmpty   = "event source is empty"
	ErrEventTimeEmpty     = "event time is empty"
	ErrEventDataEmpty     = "event data is empty"
	ErrInvalidEventSource = "invalid event source format: %s"

	// AWS validation errors
	ErrAWSRegionEmpty            = "AWS region is empty"
	ErrAWSAccountIDEmpty         = "AWS account ID is empty"
	ErrECRRegistryEmpty          = "ECR registry is empty"
	ErrInvalidAWSRegionLength    = "AWS region length must be between %d and %d characters: %s"
	ErrInvalidAWSAccountIDLength = "AWS account ID must be exactly %d digits: %s"
	ErrInvalidECRRegistrySuffix  = "ECR registry must end with '.amazonaws.com': %s"

	// S3 validation errors
	ErrInvalidS3BucketLength = "S3 bucket name length must be between %d and %d characters: %s"
	ErrInvalidS3BucketChars  = "S3 bucket name contains invalid characters: %s"
	ErrS3KeyTooLong          = "S3 key too long (max %d characters): %s"
	ErrInvalidS3KeyFormat    = "S3 key must start with 'global/parser/': %s"

	// Job validation errors
	ErrJobNameEmpty  = "job name is empty"
	ErrImageURIEmpty = "image URI is empty"

	// Naming validation errors
	ErrNameEmpty                   = "name is empty"
	ErrNameTooLong                 = "name is too long (max 63 characters)"
	ErrNameInvalidStart            = "name must start with alphanumeric character"
	ErrNameInvalidEnd              = "name must end with alphanumeric character"
	ErrNameInvalidCharacters       = "name contains invalid characters (only alphanumeric and hyphens allowed)"
	ErrLabelKeyEmpty               = "label key is empty"
	ErrLabelKeyTooLong             = "label key is too long (max 253 characters)"
	ErrLabelKeyInvalidStart        = "label key must start with alphanumeric character"
	ErrLabelKeyInvalidEnd          = "label key must end with alphanumeric character"
	ErrLabelKeyInvalidCharacters   = "label key contains invalid characters (only alphanumeric, hyphens, and dots allowed)"
	ErrLabelValueEmpty             = "label value is empty"
	ErrLabelValueTooLong           = "label value is too long (max 63 characters)"
	ErrLabelValueInvalidStart      = "label value must start with alphanumeric character"
	ErrLabelValueInvalidEnd        = "label value must end with alphanumeric character"
	ErrLabelValueInvalidCharacters = "label value contains invalid characters (only alphanumeric and hyphens allowed)"
)

// Kubernetes Errors
const (
	// Kubernetes client errors
	ErrK8sConnectionFailed = "kubernetes connection failed"

	// Job management errors
	ErrFailedToCreateJob  = "failed to create build job"
	ErrFailedToSubmitJob  = "failed to submit build job"
	ErrJobAlreadyExists   = "job already exists but failed to get details"
	ErrFailedToDeleteJob  = "failed to delete failed job %s"
	ErrFailedToListJobs   = "failed to list jobs with selector %s"
	ErrFailedToGetService = "failed to get service %s"

	// Resource management errors
	ErrFailedToCreateResource  = "failed to create critical resource %s"
	ErrFailedToApplyResource   = "failed to apply %s resource '%s'"
	ErrFailedToDetermineGVR    = "failed to determine GroupVersionResource for %s"
	ErrFailedToMarshalResource = "failed to marshal resource object"

	// Resource-specific errors
	ErrPermissionDenied    = "permission denied - check RBAC permissions for %s in namespace %s"
	ErrInvalidResourceSpec = "invalid resource specification for %s '%s'"
	ErrServerTimeout       = "server timeout - Kubernetes API server is overloaded for %s"
	ErrServiceUnavailable  = "service unavailable - Kubernetes API server is down for %s"

	// Dynamic client errors
	ErrDynamicClientNotInitialized           = "dynamic client is not initialized, cannot apply resource"
	ErrDynamicClientNotInitializedForService = "dynamic client is not initialized, cannot create Knative service"

	// Resource parsing errors
	ErrFailedToParseCPULimit    = "failed to parse CPU limit '%s'"
	ErrFailedToParseMemoryLimit = "failed to parse Memory limit '%s'"

	// Job status errors
	ErrJobIsNil          = "job object is nil"
	ErrJobNamespaceEmpty = "job namespace is empty"
)

// AWS Errors
const (
	// S3 errors
	ErrS3ObjectNotFound       = "S3 object not found"
	ErrS3BuildContextNotFound = "S3 build context not found: s3://%s/%s"
	ErrS3FailedToCheckObject  = "failed to check object existence"
	ErrS3FailedToGetObject    = "failed to get object info"
	ErrS3FailedToDownload     = "failed to download parser code"
	ErrS3FailedToUpload       = "failed to upload object"
	ErrS3ListBucketsFailed    = "S3 list buckets failed"

	// ECR errors
	ErrECRFailedToCreateRepo  = "failed to create ECR repository"
	ErrECRDescribeReposFailed = "ECR describe repositories failed"
)

// Build Errors
const (
	// Build process errors
	ErrFailedToCreateBuildContext = "failed to create build context"
	ErrFailedToCreateBuildArchive = "failed to create build context archive"
	ErrFailedToUploadBuildContext = "failed to upload build context"

	// Parser file errors
	ErrParserFileNotFound      = "parser file not found: s3://%s/%s - The parser ID '%s' does not exist in the S3 bucket. Please ensure the parser file has been uploaded to the correct location"
	ErrFailedToCheckParserFile = "failed to check if parser file exists"
	ErrParserFileSizeExceeded  = "security violation: parser file size (%d bytes) exceeds maximum allowed size (%d bytes)"
	ErrParserCodeSizeMismatch  = "parser code size mismatch: expected %d bytes, got %d bytes"

	// Template processing errors
	ErrFailedToReadTemplate    = "failed to read template %s"
	ErrFailedToParseTemplate   = "failed to parse template %s"
	ErrFailedToExecuteTemplate = "failed to execute template %s"

	// Archive creation errors
	ErrFailedToWriteTarHeader       = "failed to write tar header for %s"
	ErrFailedToWriteTarContent      = "failed to write tar content for %s"
	ErrFailedToWriteParserTarHeader = "failed to write tar header for parser code"
	ErrFailedToStreamParserCode     = "failed to stream parser code to tar"

	// Service creation errors
	ErrFailedToCreateLambdaService  = "failed to create lambda service for existing job"
	ErrFailedToCreateKnativeService = "failed to create Knative service"

	// Container errors
	ErrKanikoContainerNotFound = "kaniko container not found"

	// CloudEvents errors
	ErrFailedToSetEventData            = "failed to set event data"
	ErrFailedToCreateCloudEventsClient = "failed to create CloudEvents client"
	ErrFailedToSendEventToBroker       = "failed to send event to broker"
	ErrBrokerURLNotConfigured          = "broker URL is not configured"
)

// HTTP Errors
const (
	// HTTP request errors
	ErrHTTPRequestFailed = "HTTP request failed"
	ErrHTTPTimeout       = "HTTP request timeout"
	ErrHTTPConflict      = "HTTP conflict"
	ErrHTTPServerError   = "HTTP server error"

	// HTTP response errors
	ErrHTTPResponseInvalid  = "HTTP response invalid"
	ErrHTTPResponseEmpty    = "HTTP response empty"
	ErrHTTPResponseTooLarge = "HTTP response too large"

	// HTTP Error Response Types
	ErrHTTPBadRequest          = "bad_request"
	ErrHTTPUnauthorized        = "unauthorized"
	ErrHTTPForbidden           = "forbidden"
	ErrHTTPNotFound            = "not_found"
	ErrHTTPMethodNotAllowed    = "method_not_allowed"
	ErrHTTPRequestTimeout      = "request_timeout"
	ErrHTTPRequestTooLarge     = "request_entity_too_large"
	ErrHTTPTooManyRequests     = "too_many_requests"
	ErrHTTPInternalServerError = "internal_server_error"
	ErrHTTPServiceUnavailable  = "service_unavailable"
	ErrHTTPGatewayTimeout      = "gateway_timeout"

	// HTTP Error Response Messages
	ErrHTTPBadRequestMessage          = "The request could not be processed due to invalid syntax"
	ErrHTTPUnauthorizedMessage        = "Authentication is required to access this resource"
	ErrHTTPForbiddenMessage           = "Access to this resource is forbidden"
	ErrHTTPNotFoundMessage            = "The requested resource was not found"
	ErrHTTPMethodNotAllowedMessage    = "The HTTP method is not allowed for this resource"
	ErrHTTPRequestTimeoutMessage      = "The request timed out"
	ErrHTTPRequestTooLargeMessage     = "The request entity is too large"
	ErrHTTPTooManyRequestsMessage     = "Too many requests, please try again later"
	ErrHTTPInternalServerErrorMessage = "An internal server error occurred"
	ErrHTTPServiceUnavailableMessage  = "The service is temporarily unavailable"
	ErrHTTPGatewayTimeoutMessage      = "The gateway timed out"
)

// Security Errors
const (
	// Security violation errors
	ErrSecurityViolationInvalidThirdPartyID = "security violation: invalid third party ID in build completion event"
	ErrSecurityViolationInvalidParserID     = "security violation: invalid parser ID in build completion event"
	ErrSecurityViolationUntrustedImageURI   = "security violation: untrusted image URI in build completion event. Expected '%s', got '%s'"
	ErrSecurityForbidden                    = "access forbidden"
)

// Observability Errors
const (
	// Tracing errors
	ErrFailedToCreateOTLPExporter = "failed to create OTLP exporter"

	// Metrics errors
	ErrMetricsCollectionFailed = "metrics collection failed"
	ErrMetricsExportFailed     = "metrics export failed"

	// Logging errors
	ErrLoggingFailed   = "logging failed"
	ErrLogExportFailed = "log export failed"
)

// Retry Errors
const (
	// Retry mechanism errors
	ErrOperationFailedAfterRetries = "operation %s failed after %d retries"
	ErrRetryTimeout                = "retry timeout"
	ErrMaxRetriesExceeded          = "max retries exceeded"
	ErrRetryBackoffFailed          = "retry backoff failed"

	// Rate limiter errors
	ErrRateLimiterCreation  = "failed to create rate limiter"
	ErrRateLimiterCheck     = "rate limit check failed"
	ErrRateLimiterRejection = "rate limit exceeded"
	ErrRateLimiterClose     = "failed to close rate limiter"
)

// Component and operation constants for error creation
const (
	// Component names
	Kubernetes = "kubernetes"
	AWS        = "AWS"
	S3         = "S3"
	ECR        = "ECR"

	// Operation types
	Kubeconfig    = "kubeconfig"
	Config        = "config"
	Client        = "client"
	DynamicClient = "dynamic_client"
	HealthCheck   = "health_check"
	Connection    = "connection"
	Validation    = "validation"
	Operation     = "operation"
)

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔧 ERROR CONSTRUCTORS - "Factory functions for creating error types"  │
// └─────────────────────────────────────────────────────────────────────────┘

// 🔧 NewValidationError - "Create validation error with context"
func NewValidationError(field string, value interface{}, reason string) error {
	return &ValidationError{
		Field:  field,
		Value:  value,
		Reason: reason,
		Time:   time.Now(),
	}
}

// 🔧 NewConfigurationError - "Create configuration error"
func NewConfigurationError(component, setting, reason string) error {
	return &ConfigurationError{
		Component: component,
		Setting:   setting,
		Reason:    reason,
		Time:      time.Now(),
	}
}

// 🔧 NewConfigurationErrorWithValue - "Create configuration error with value"
func NewConfigurationErrorWithValue(component, setting string, value interface{}, reason string) error {
	return &ConfigurationError{
		Component: component,
		Setting:   setting,
		Value:     value,
		Reason:    reason,
		Time:      time.Now(),
	}
}

// 🔧 NewSystemError - "Create system error"
func NewSystemError(component, operation string) error {
	return &SystemError{
		Component: component,
		Operation: operation,
		Time:      time.Now(),
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔍 ERROR TYPE CHECKERS - "Type checking and classification helpers"   │
// └─────────────────────────────────────────────────────────────────────────┘

// 🔍 IsNotFoundError - "Check if error is a not found error"
func IsNotFoundError(err error) bool {
	return isErrorType[*NotFoundError](err)
}

// 🔍 isErrorType - "Generic error type checker"
func isErrorType[T any](err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(T)
	return ok
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔄 ERROR WRAPPING - "Context-aware error wrapping"                    │
// └─────────────────────────────────────────────────────────────────────────┘

// 🔄 WrapWithContext - "Wrap error with additional context"
func WrapWithContext(err error, context string) error {
	if err == nil {
		return nil
	}

	// Try to preserve the original error type while adding context
	switch e := err.(type) {
	case *ValidationError:
		return &ValidationError{
			Field:   e.Field,
			Value:   e.Value,
			Reason:  e.Reason,
			Context: context,
			Time:    e.Time,
		}
	case *ConfigurationError:
		return &ConfigurationError{
			Component: e.Component,
			Setting:   e.Setting,
			Value:     e.Value,
			Reason:    e.Reason,
			Time:      e.Time,
		}
	case *ConnectionError:
		return &ConnectionError{
			Service:   e.Service,
			Endpoint:  e.Endpoint,
			Operation: e.Operation,
			Timeout:   e.Timeout,
			Retries:   e.Retries,
			Reason:    e.Reason,
			Time:      e.Time,
		}
	case *TimeoutError:
		return &TimeoutError{
			Operation: e.Operation,
			Timeout:   e.Timeout,
			Context:   context,
			Time:      e.Time,
		}
	case *NotFoundError:
		return &NotFoundError{
			Resource:   e.Resource,
			Identifier: e.Identifier,
			Context:    context,
			Time:       e.Time,
		}
	case *PermissionError:
		return &PermissionError{
			Operation: e.Operation,
			Resource:  e.Resource,
			Subject:   e.Subject,
			Context:   context,
			Time:      e.Time,
		}
	case *SystemError:
		return &SystemError{
			Component: e.Component,
			Operation: e.Operation,
			Code:      e.Code,
			Details:   e.Details,
			Time:      e.Time,
		}
	default:
		return fmt.Errorf("%s: %w", context, err)
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🚨 ERROR HANDLING UTILITIES - "Reusable error handling patterns"      │
// └─────────────────────────────────────────────────────────────────────────┘

// Observability interface for logging - minimal interface for error handling
type Observability interface {
	Error(ctx context.Context, err error, message string, fields ...interface{})
	Info(ctx context.Context, message string, fields ...interface{})
}

// Helper functions for error handling
func extractLabel(labels []string, key string) string {
	for i := 0; i < len(labels)-1; i += 2 {
		if labels[i] == key {
			return labels[i+1]
		}
	}
	return "unknown"
}

func classifyError(err error) (category, severity string) {
	errStr := err.Error()

	// Classify by error type
	switch {
	case strings.Contains(errStr, "validation"):
		category = "validation"
		severity = "warning"
	case strings.Contains(errStr, "connection"):
		category = "connection"
		severity = "error"
	case strings.Contains(errStr, "timeout"):
		category = "timeout"
		severity = "error"
	case strings.Contains(errStr, "permission"):
		category = "security"
		severity = "critical"
	case strings.Contains(errStr, "not found"):
		category = "not_found"
		severity = "warning"
	default:
		category = "unknown"
		severity = "error"
	}

	return category, severity
}

// 🎯 HandleError - "Enhanced error handling with metrics and context"
func HandleError(ctx context.Context, obs Observability, err error, message string, labels ...string) error {
	if err == nil {
		return nil
	}

	// Extract context from labels
	component := extractLabel(labels, "component")
	operation := extractLabel(labels, "operation")
	context := extractLabel(labels, "context")

	// Determine error category and severity
	category, severity := classifyError(err)

	// Record error metrics
	recordErrorMetrics(category, severity, context, component, operation)

	// Log error with structured logging
	obs.Error(ctx, err, message,
		"category", category,
		"severity", severity,
		"component", component,
		"operation", operation,
		"context", context,
	)

	return err
}

// 🎯 HandleFatalError - "Handle fatal errors with metrics and termination"
func HandleFatalError(ctx context.Context, obs Observability, err error, message string, labels ...string) {
	if err == nil {
		return
	}

	// Extract context from labels
	component := extractLabel(labels, "component")
	operation := extractLabel(labels, "operation")
	context := extractLabel(labels, "context")

	// Determine error category and severity
	category, severity := classifyError(err)

	// Record error metrics
	recordErrorMetrics(category, severity, context, component, operation)

	// Log fatal error
	obs.Error(ctx, err, message,
		"category", category,
		"severity", severity,
		"component", component,
		"operation", operation,
		"context", context,
		"fatal", true,
	)

	// Exit with error code
	os.Exit(1)
}

// 🚨 ValidateRequired - "Validate required fields"
func ValidateRequired(fieldName string, value interface{}) error {
	if value == nil {
		return fmt.Errorf("%s: %s", ErrConfigRequired, fieldName)
	}

	switch v := value.(type) {
	case string:
		if v == "" {
			return fmt.Errorf("%s: %s", ErrConfigRequired, fieldName)
		}
	case []string:
		if len(v) == 0 {
			return fmt.Errorf("%s: %s", ErrConfigRequired, fieldName)
		}
	}

	return nil
}

// 🚨 WrapError - "Wrap error with context"
func WrapError(err error, message string, fields ...interface{}) error {
	if err == nil {
		return nil
	}

	if len(fields) > 0 {
		fieldStr := ""
		for i := 0; i < len(fields); i += 2 {
			if i+1 < len(fields) {
				if fieldStr != "" {
					fieldStr += ", "
				}
				fieldStr += fmt.Sprintf("%v=%v", fields[i], fields[i+1])
			}
		}
		return fmt.Errorf("%s (%s): %w", message, fieldStr, err)
	}

	return fmt.Errorf("%s: %w", message, err)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔧 INITIALIZATION UTILITIES - "Common initialization patterns"         │
// └─────────────────────────────────────────────────────────────────────────┘

// 🔧 InitializeComponent - "Initialize component with error handling"
func InitializeComponent(ctx context.Context, obs Observability, name string, initFunc func() error) error {
	obs.Info(ctx, "Initializing component", "component", name)

	err := initFunc()
	if err != nil {
		return HandleError(ctx, obs, err, ErrConfigInvalid,
			"component", name, "error_type", "initialization")
	}

	obs.Info(ctx, "Component initialized successfully", "component", name)
	return nil
}

// 🔧 InitializeService - "Initialize service with observability"
func InitializeService(ctx context.Context, obs Observability, serviceName string, initFunc func() error) {
	obs.Info(ctx, "Starting service initialization", "service", serviceName)

	err := initFunc()
	if err != nil {
		HandleFatalError(ctx, obs, err, ErrConfigInvalid,
			"service", serviceName, "error_type", "initialization")
	}

	obs.Info(ctx, "Service initialized successfully", "service", serviceName)
}

// recordErrorMetrics records error metrics if a recorder is available
func recordErrorMetrics(category, severity, context, component, operation string) {
	// This function is kept for future use when metrics integration is implemented
}
