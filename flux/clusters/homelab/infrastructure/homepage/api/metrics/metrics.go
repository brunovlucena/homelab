package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// =============================================================================
// 📊 PROMETHEUS METRICS
// =============================================================================

var (
	// 🎯 HTTP Request metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "handler", "code"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "handler"},
	)

	// 📦 Projects API metrics
	ProjectsLoadErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bruno_site_projects_load_errors_total",
			Help: "Total number of errors when loading projects from database",
		},
		[]string{"error_type"},
	)

	ProjectsLoadSuccess = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "bruno_site_projects_load_success_total",
			Help: "Total number of successful project loads from database",
		},
	)

	ProjectsLoadDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "bruno_site_projects_load_duration_seconds",
			Help:    "Time taken to load projects from database",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 2.0},
		},
	)

	// 💼 Experience API metrics
	ExperienceLoadErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bruno_site_experience_load_errors_total",
			Help: "Total number of errors when loading experience data from database",
		},
		[]string{"error_type"},
	)

	ExperienceLoadSuccess = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "bruno_site_experience_load_success_total",
			Help: "Total number of successful experience data loads from database",
		},
	)

	ExperienceLoadDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "bruno_site_experience_load_duration_seconds",
			Help:    "Time taken to load experience data from database",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 2.0},
		},
	)

	// 💾 Database metrics
	DatabaseConnectionErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "bruno_site_database_connection_errors_total",
			Help: "Total number of database connection errors",
		},
	)

	DatabaseQueryErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bruno_site_database_query_errors_total",
			Help: "Total number of database query errors",
		},
		[]string{"operation", "table"},
	)

	// 🔴 Redis metrics
	RedisOperationErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bruno_site_redis_operation_errors_total",
			Help: "Total number of Redis operation errors",
		},
		[]string{"operation"},
	)

	// 📦 MinIO metrics
	MinIOOperationErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bruno_site_minio_operation_errors_total",
			Help: "Total number of MinIO operation errors",
		},
		[]string{"operation"},
	)

	// 🤖 Agent-SRE proxy metrics
	AgentSRERequestErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bruno_site_agent_sre_request_errors_total",
			Help: "Total number of Agent-SRE proxy request errors",
		},
		[]string{"endpoint", "error_type"},
	)

	AgentSRERequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "bruno_site_agent_sre_request_duration_seconds",
			Help:    "Agent-SRE proxy request duration in seconds",
			Buckets: []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0, 30.0},
		},
		[]string{"endpoint"},
	)
)

// =============================================================================
// 📊 METRIC HELPER FUNCTIONS
// =============================================================================

// RecordProjectsLoadError records a project load error with error type
func RecordProjectsLoadError(errorType string) {
	ProjectsLoadErrors.WithLabelValues(errorType).Inc()
}

// RecordProjectsLoadSuccess records a successful project load
func RecordProjectsLoadSuccess() {
	ProjectsLoadSuccess.Inc()
}

// RecordDatabaseError records a database error
func RecordDatabaseError(operation, table string) {
	DatabaseQueryErrors.WithLabelValues(operation, table).Inc()
}

// RecordDatabaseConnectionError records a database connection error
func RecordDatabaseConnectionError() {
	DatabaseConnectionErrors.Inc()
}

// RecordRedisError records a Redis operation error
func RecordRedisError(operation string) {
	RedisOperationErrors.WithLabelValues(operation).Inc()
}

// RecordMinIOError records a MinIO operation error
func RecordMinIOError(operation string) {
	MinIOOperationErrors.WithLabelValues(operation).Inc()
}

// RecordAgentSREError records an Agent-SRE proxy error
func RecordAgentSREError(endpoint, errorType string) {
	AgentSRERequestErrors.WithLabelValues(endpoint, errorType).Inc()
}

// RecordExperienceLoadError records an experience load error with error type
func RecordExperienceLoadError(errorType string) {
	ExperienceLoadErrors.WithLabelValues(errorType).Inc()
}

// RecordExperienceLoadSuccess records a successful experience load
func RecordExperienceLoadSuccess() {
	ExperienceLoadSuccess.Inc()
}
