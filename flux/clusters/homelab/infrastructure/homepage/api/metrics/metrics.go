package metrics

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// =============================================================================
// 📊 OPENTELEMETRY METRICS
// =============================================================================

var (
	meter metric.Meter

	// 🎯 HTTP Request metrics
	HTTPRequestsTotal   metric.Int64Counter
	HTTPRequestDuration metric.Float64Histogram
	HTTPActiveRequests  metric.Int64UpDownCounter // Track concurrent requests
	HTTPRequestSize     metric.Int64Histogram     // Track request payload sizes
	HTTPResponseSize    metric.Int64Histogram     // Track response payload sizes

	// 📦 Projects API metrics
	ProjectsLoadErrors   metric.Int64Counter
	ProjectsLoadSuccess  metric.Int64Counter
	ProjectsLoadDuration metric.Float64Histogram
	ProjectsCount        metric.Int64Gauge // Total projects in database

	// 💼 Experience API metrics
	ExperienceLoadErrors   metric.Int64Counter
	ExperienceLoadSuccess  metric.Int64Counter
	ExperienceLoadDuration metric.Float64Histogram

	// 🛠️ Skills API metrics
	SkillsLoadErrors   metric.Int64Counter
	SkillsLoadSuccess  metric.Int64Counter
	SkillsLoadDuration metric.Float64Histogram

	// 💾 Database metrics
	DatabaseConnectionErrors metric.Int64Counter
	DatabaseQueryErrors      metric.Int64Counter
	DatabaseQueryDuration    metric.Float64Histogram // Track query performance
	DatabaseActiveConns      metric.Int64Gauge       // Track active connections

	// 🔴 Redis metrics
	RedisOperationErrors   metric.Int64Counter
	RedisOperationDuration metric.Float64Histogram
	RedisCacheHits         metric.Int64Counter
	RedisCacheMisses       metric.Int64Counter

	// 📦 MinIO metrics
	MinIOOperationErrors   metric.Int64Counter
	MinIOOperationDuration metric.Float64Histogram
	MinIOUploadSize        metric.Int64Histogram
	MinIODownloadSize      metric.Int64Histogram

	// 🤖 Agent-SRE proxy metrics
	AgentSRERequestErrors   metric.Int64Counter
	AgentSRERequestDuration metric.Float64Histogram

	// 🤖 Jamie proxy metrics
	JamieRequestErrors   metric.Int64Counter
	JamieRequestDuration metric.Float64Histogram

	// 🤖 Agent Bruno proxy metrics
	AgentBrunoRequestErrors   metric.Int64Counter
	AgentBrunoRequestDuration metric.Float64Histogram
)

// InitMetrics initializes all OpenTelemetry metrics
func InitMetrics() error {
	meter = otel.Meter("bruno-site")

	var err error

	// ============================================================================
	// 🎯 HTTP Request metrics
	// ============================================================================
	HTTPRequestsTotal, err = meter.Int64Counter(
		"http.requests.total",
		metric.WithDescription("Total number of HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return err
	}

	HTTPRequestDuration, err = meter.Float64Histogram(
		"http.request.duration",
		metric.WithDescription("HTTP request duration"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	// ============================================================================
	// 📦 Projects API metrics
	// ============================================================================
	ProjectsLoadErrors, err = meter.Int64Counter(
		"bruno.site.projects.load.errors",
		metric.WithDescription("Total number of errors when loading projects from database"),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		return err
	}

	ProjectsLoadSuccess, err = meter.Int64Counter(
		"bruno.site.projects.load.success",
		metric.WithDescription("Total number of successful project loads from database"),
		metric.WithUnit("{load}"),
	)
	if err != nil {
		return err
	}

	ProjectsLoadDuration, err = meter.Float64Histogram(
		"bruno.site.projects.load.duration",
		metric.WithDescription("Time taken to load projects from database"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	// ============================================================================
	// 💼 Experience API metrics
	// ============================================================================
	ExperienceLoadErrors, err = meter.Int64Counter(
		"bruno.site.experience.load.errors",
		metric.WithDescription("Total number of errors when loading experience data from database"),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		return err
	}

	ExperienceLoadSuccess, err = meter.Int64Counter(
		"bruno.site.experience.load.success",
		metric.WithDescription("Total number of successful experience data loads from database"),
		metric.WithUnit("{load}"),
	)
	if err != nil {
		return err
	}

	ExperienceLoadDuration, err = meter.Float64Histogram(
		"bruno.site.experience.load.duration",
		metric.WithDescription("Time taken to load experience data from database"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	// ============================================================================
	// 💾 Database metrics
	// ============================================================================
	DatabaseConnectionErrors, err = meter.Int64Counter(
		"bruno.site.database.connection.errors",
		metric.WithDescription("Total number of database connection errors"),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		return err
	}

	DatabaseQueryErrors, err = meter.Int64Counter(
		"bruno.site.database.query.errors",
		metric.WithDescription("Total number of database query errors"),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		return err
	}

	// ============================================================================
	// 🔴 Redis metrics
	// ============================================================================
	RedisOperationErrors, err = meter.Int64Counter(
		"bruno.site.redis.operation.errors",
		metric.WithDescription("Total number of Redis operation errors"),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		return err
	}

	// ============================================================================
	// 📦 MinIO metrics
	// ============================================================================
	MinIOOperationErrors, err = meter.Int64Counter(
		"bruno.site.minio.operation.errors",
		metric.WithDescription("Total number of MinIO operation errors"),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		return err
	}

	// ============================================================================
	// 🤖 Agent-SRE proxy metrics
	// ============================================================================
	AgentSRERequestErrors, err = meter.Int64Counter(
		"bruno.site.agent_sre.request.errors",
		metric.WithDescription("Total number of Agent-SRE proxy request errors"),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		return err
	}

	AgentSRERequestDuration, err = meter.Float64Histogram(
		"bruno.site.agent_sre.request.duration",
		metric.WithDescription("Agent-SRE proxy request duration"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	log.Println("✅ OpenTelemetry metrics initialized")
	return nil
}

// =============================================================================
// 📊 METRIC HELPER FUNCTIONS
// =============================================================================

// RecordProjectsLoadError records a project load error with error type
func RecordProjectsLoadError(errorType string) {
	ProjectsLoadErrors.Add(context.Background(), 1,
		metric.WithAttributes(attribute.String("error_type", errorType)))
}

// RecordProjectsLoadSuccess records a successful project load
func RecordProjectsLoadSuccess() {
	ProjectsLoadSuccess.Add(context.Background(), 1)
}

// RecordDatabaseError records a database error
func RecordDatabaseError(operation, table string) {
	DatabaseQueryErrors.Add(context.Background(), 1,
		metric.WithAttributes(
			attribute.String("operation", operation),
			attribute.String("table", table)))
}

// RecordDatabaseConnectionError records a database connection error
func RecordDatabaseConnectionError() {
	DatabaseConnectionErrors.Add(context.Background(), 1)
}

// RecordRedisError records a Redis operation error
func RecordRedisError(operation string) {
	RedisOperationErrors.Add(context.Background(), 1,
		metric.WithAttributes(attribute.String("operation", operation)))
}

// RecordMinIOError records a MinIO operation error
func RecordMinIOError(operation string) {
	MinIOOperationErrors.Add(context.Background(), 1,
		metric.WithAttributes(attribute.String("operation", operation)))
}

// RecordAgentSREError records an Agent-SRE proxy error
func RecordAgentSREError(endpoint, errorType string) {
	AgentSRERequestErrors.Add(context.Background(), 1,
		metric.WithAttributes(
			attribute.String("endpoint", endpoint),
			attribute.String("error_type", errorType)))
}

// RecordExperienceLoadError records an experience load error with error type
func RecordExperienceLoadError(errorType string) {
	ExperienceLoadErrors.Add(context.Background(), 1,
		metric.WithAttributes(attribute.String("error_type", errorType)))
}

// RecordExperienceLoadSuccess records a successful experience load
func RecordExperienceLoadSuccess() {
	ExperienceLoadSuccess.Add(context.Background(), 1)
}
