// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	📝 COMMON TYPES - Common types and constants for Knative Lambda Service
//
//	🎯 Purpose: Shared types, constants, and enumerations
//	💡 Features: Build status, common constants, shared interfaces
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package builds

import (
	"strconv"
	"time"
)

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📊 BUILD STATUS ENUMERATION - "Build status constants"                │
// └─────────────────────────────────────────────────────────────────────────┘

// 📊 BuildStatus - "Build status enumeration"
type BuildStatus string

const (
	BuildStatusPending   BuildStatus = "pending"
	BuildStatusRunning   BuildStatus = "running"
	BuildStatusCompleted BuildStatus = "completed"
	BuildStatusFailed    BuildStatus = "failed"
	BuildStatusCancelled BuildStatus = "cancelled"
	BuildStatusTimeout   BuildStatus = "timeout"
)

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔧 BUILD STEP - "Individual build step configuration"                  │
// └─────────────────────────────────────────────────────────────────────────┘

// 🔧 BuildStep - "Individual build step"
type BuildStep struct {
	Name        string            `json:"name"`
	Command     string            `json:"command"`
	Args        []string          `json:"args,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	WorkingDir  string            `json:"working_dir,omitempty"`
	Timeout     int               `json:"timeout,omitempty"`
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🏗️ BUILD JOB - "Kubernetes job representation"                        │
// └─────────────────────────────────────────────────────────────────────────┘

// 🏗️ BuildJob - "Kubernetes job representation"
type BuildJob struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Namespace    string            `json:"namespace"`
	Status       string            `json:"status"`
	CreatedAt    time.Time         `json:"created_at"`
	StartedAt    *time.Time        `json:"started_at,omitempty"`
	CompletedAt  *time.Time        `json:"completed_at,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
	Annotations  map[string]string `json:"annotations,omitempty"`
	BuildRequest *BuildRequest     `json:"build_request,omitempty"`
	ErrorMessage string            `json:"error_message,omitempty"`
	RetryCount   int               `json:"retry_count"`
	LastRetryAt  *time.Time        `json:"last_retry_at,omitempty"`
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🌐 BUILD EVENT DATA - "CloudEvent data structure for build requests"   │
// └─────────────────────────────────────────────────────────────────────────┘

// 🌐 BuildEventData - "CloudEvent data structure for build requests"
type BuildEventData struct {
	ThirdPartyID string                 `json:"third_party_id"`
	ParserID     string                 `json:"parser_id"`
	ContentHash  string                 `json:"content_hash,omitempty"` // New: content hash for unique image tagging
	ContextID    string                 `json:"context_id,omitempty"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
}

// Validate checks if the BuildEventData is valid and returns validation errors
func (b *BuildEventData) Validate() []string {
	var errors []string

	if b.ThirdPartyID == "" {
		errors = append(errors, "third_party_id is required")
	}

	if b.ParserID == "" {
		errors = append(errors, "parser_id is required")
	}

	// Validate ThirdPartyID format (alphanumeric and hyphens only, max 100 chars)
	if b.ThirdPartyID != "" && len(b.ThirdPartyID) > 100 {
		errors = append(errors, "third_party_id must be 100 characters or less")
	}

	// Validate ParserID format (alphanumeric and hyphens only, max 100 chars)
	if b.ParserID != "" && len(b.ParserID) > 100 {
		errors = append(errors, "parser_id must be 100 characters or less")
	}

	return errors
}

// IsValid returns true if the BuildEventData is valid
func (b *BuildEventData) IsValid() bool {
	return len(b.Validate()) == 0
}

// GetParameterAsString safely extracts a string parameter
func (b *BuildEventData) GetParameterAsString(key string) (string, bool) {
	if b.Parameters == nil {
		return "", false
	}

	if val, ok := b.Parameters[key]; ok {
		if str, ok := val.(string); ok {
			return str, true
		}
	}
	return "", false
}

// GetParameterAsInt safely extracts an int parameter
func (b *BuildEventData) GetParameterAsInt(key string) (int, bool) {
	if b.Parameters == nil {
		return 0, false
	}

	if val, ok := b.Parameters[key]; ok {
		switch v := val.(type) {
		case int:
			return v, true
		case float64:
			return int(v), true
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i, true
			}
		}
	}
	return 0, false
}

// 🌐 BuildCompletionEventData - "CloudEvent data structure for build completion events"
type BuildCompletionEventData struct {
	ThirdPartyID  string        `json:"third_party_id"`
	ParserID      string        `json:"parser_id"`
	ContentHash   string        `json:"content_hash,omitempty"` // New: content hash for unique image tagging
	JobName       string        `json:"job_name"`
	ImageURI      string        `json:"image_uri,omitempty"`
	Status        string        `json:"status"`
	ErrorMessage  string        `json:"error_message,omitempty"`
	ExitCode      int           `json:"exit_code,omitempty"`
	Duration      time.Duration `json:"duration"`
	CorrelationID string        `json:"correlation_id"`
}

// 🌐 JobStartEventData - "CloudEvent data structure for job start events"
type JobStartEventData struct {
	ThirdPartyID  string                 `json:"third_party_id"`
	ParserID      string                 `json:"parser_id"`
	CorrelationID string                 `json:"correlation_id"`
	JobName       string                 `json:"job_name,omitempty"` // Optional: if not provided, will be generated
	Parameters    map[string]interface{} `json:"parameters,omitempty"`
	Priority      int                    `json:"priority,omitempty"` // Optional: job priority (1-10, 1=highest)
}

// 🌐 ServiceDeleteEventData - "CloudEvent data structure for service deletion events"
type ServiceDeleteEventData struct {
	ThirdPartyID  string `json:"third_party_id"`
	ParserID      string `json:"parser_id"`
	ServiceName   string `json:"service_name,omitempty"` // Optional: if not provided, will be generated
	CorrelationID string `json:"correlation_id"`
	Reason        string `json:"reason,omitempty"` // Optional: reason for deletion
}

// 🌐 HandlerResponse - "HTTP response structure"
type HandlerResponse struct {
	Status        string    `json:"status"`
	Message       string    `json:"message"`
	JobName       string    `json:"job_name,omitempty"`
	CorrelationID string    `json:"correlation_id,omitempty"`
	Error         string    `json:"error,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🎯 EVENT TYPES - "CloudEvent type constants"                          │
// └─────────────────────────────────────────────────────────────────────────┘

// Common event types and constants for CloudEvents
const (
	// 🚀 Build Events
	EventTypeBuildStart    = "network.notifi.lambda.build.start"
	EventTypeBuildComplete = "network.notifi.lambda.build.complete"
	EventTypeBuildFailed   = "network.notifi.lambda.build.failed"
	EventTypeBuildTimeout  = "network.notifi.lambda.build.timeout"
	EventTypeBuildCancel   = "network.notifi.lambda.build.cancel"
	EventTypeJobStart      = "network.notifi.lambda.job.start"

	// 🔄 Parser Events
	EventTypeParserStart    = "network.notifi.lambda.parser.start"
	EventTypeParserComplete = "network.notifi.lambda.parser.complete"
	EventTypeParserFailed   = "network.notifi.lambda.parser.failed"

	// 🚀 Rebuild Events
	EventTypeRebuildStart = "network.notifi.lambda.rebuild.start"

	// 📊 Status Events
	EventTypeStatusUpdate = "network.notifi.lambda.status.update"
	EventTypeHealthCheck  = "network.notifi.lambda.health.check"

	// 🔧 Management Events
	EventTypeServiceCreate = "network.notifi.lambda.service.create"
	EventTypeServiceUpdate = "network.notifi.lambda.service.update"
	EventTypeServiceDelete = "network.notifi.lambda.service.delete"
)

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🌐 EVENT SOURCES - "CloudEvent source constants"                      │
// └─────────────────────────────────────────────────────────────────────────┘

// 🌐 Event Sources
const (
	SourceParsers    = "network.notifi.parsers"
	SourceBuilder    = "network.notifi.builder"
	SourceMonitoring = "network.notifi.monitoring"
	SourceAPI        = "network.notifi.api"
	SourceTest       = "network.notifi.test"
)

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📊 CLOUDEVENT CONSTANTS - "CloudEvent specification constants"        │
// └─────────────────────────────────────────────────────────────────────────┘

// 📊 CloudEvent Specification Version
const (
	CloudEventSpecVersion = "1.0"
	ContentTypeJSON       = "application/json"
	ContentTypeCloudEvent = "application/cloudevents+json"
)
