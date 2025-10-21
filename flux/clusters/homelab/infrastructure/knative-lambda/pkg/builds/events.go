// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🌐 CLOUDEVENTS - CloudEvent structures and publisher interfaces
//
//	🎯 Purpose: CloudEvent handling, publishing, and event management
//	💡 Features: Event publishing, builder pattern, publisher stats
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package builds

import (
	"context"
	"time"
)

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🌐 CLOUDEVENT STRUCTURE - "CloudEvent data model"                     │
// └─────────────────────────────────────────────────────────────────────────┘

// 🌐 CloudEvent - "CloudEvent structure for event processing"
type CloudEvent struct {
	SpecVersion     string            `json:"specversion"`
	Type            string            `json:"type"`
	Source          string            `json:"source"`
	ID              string            `json:"id"`
	Time            *time.Time        `json:"time,omitempty"`
	DataContentType string            `json:"datacontenttype,omitempty"`
	Subject         string            `json:"subject,omitempty"`
	Data            interface{}       `json:"data,omitempty"`
	Extensions      map[string]string `json:"-"`
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📤 PUBLISH OPTIONS - "Event publishing configuration"                 │
// └─────────────────────────────────────────────────────────────────────────┘

// 📤 PublishOptions - "Options for publishing events"
type PublishOptions struct {
	Exchange   string
	RoutingKey string
	Retry      bool
	RetryCount int
	RetryDelay time.Duration
	Timeout    time.Duration
	Headers    map[string]string
	Priority   int
	Persistent bool
}

// 📤 PublishResult - "Result of a publish operation"
type PublishResult struct {
	EventID   string
	Success   bool
	Error     error
	Timestamp time.Time
	Duration  time.Duration
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📡 PUBLISHER INTERFACE - "Event publisher contract"                   │
// └─────────────────────────────────────────────────────────────────────────┘

// 📡 Publisher - "Interface for publishing CloudEvents"
type Publisher interface {
	// 📤 Publish a single CloudEvent
	Publish(ctx context.Context, event *CloudEvent, opts *PublishOptions) (*PublishResult, error)

	// 📤 Publish multiple CloudEvents in batch
	PublishBatch(ctx context.Context, events []*CloudEvent, opts *PublishOptions) ([]*PublishResult, error)

	// 🔧 Publisher lifecycle management
	Connect(ctx context.Context) error
	Close(ctx context.Context) error
	IsConnected() bool

	// 📊 Publisher health and metrics
	Health(ctx context.Context) error
	Stats() PublisherStats
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📊 PUBLISHER STATS - "Publisher performance metrics"                  │
// └─────────────────────────────────────────────────────────────────────────┘

// 📊 PublisherStats - "Publisher statistics and metrics"
type PublisherStats struct {
	EventsPublished    int64
	EventsFailed       int64
	ConnectionUptime   time.Duration
	LastPublishTime    time.Time
	LastError          error
	CurrentConnections int
}
