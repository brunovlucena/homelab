// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	📦 BUILD REQUEST - Build request models and validation
//
//	🎯 Purpose: Data structures for build requests from CloudEvents
//	💡 Features: Request validation, ID generation, image tagging
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package builds

import (
	"time"
)

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🏗️ BUILD REQUEST MODEL - "CloudEvent build request"                   │
// └─────────────────────────────────────────────────────────────────────────┘

// 🏗️ BuildRequest - "Build request from CloudEvent"
type BuildRequest struct {
	// Event metadata
	EventID     string `json:"event_id" validate:"required"`
	EventType   string `json:"event_type" validate:"required"`
	EventSource string `json:"event_source" validate:"required"`

	// Build identifiers
	ThirdPartyID string `json:"third_party_id" validate:"required,min=1,max=100"`
	ParserID     string `json:"parser_id" validate:"required,min=1,max=100"`
	ContentHash  string `json:"content_hash,omitempty"` // New: content hash for unique image tagging
	BlockID      string `json:"block_id,omitempty"` // For parallel processing of different blocks

	// Build configuration
	BuildType string `json:"build_type" validate:"required,oneof=container lambda"`
	Runtime   string `json:"runtime" validate:"required"`
	SourceURL string `json:"source_url" validate:"required,url"`

	// S3 source configuration
	SourceBucket string `json:"source_bucket,omitempty"`
	SourceKey    string `json:"source_key,omitempty"`

	// Optional parameters
	BuildTimeout int               `json:"build_timeout,omitempty" validate:"min=60,max=3600"`
	Environment  map[string]string `json:"environment,omitempty"`
	BuildArgs    map[string]string `json:"build_args,omitempty"`
	Tags         []string          `json:"tags,omitempty"`

	// Metadata
	Timestamp     time.Time              `json:"timestamp"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`

	// Build configuration
	BuildConfig BuildConfig `json:"build_config,omitempty"`

	// Creation timestamp
	CreatedAt time.Time `json:"created_at"`
}
