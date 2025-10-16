package metrics

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// =============================================================================
// 📊 PROMETHEUS METRICS
// =============================================================================

const (
	namespace = "homepage"
)

var (
	// 🎯 HTTP Request metrics
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPActiveRequests  *prometheus.GaugeVec
	HTTPRequestSize     *prometheus.HistogramVec
	HTTPResponseSize    *prometheus.HistogramVec

	// 📦 Projects API metrics
	ProjectsLoadErrors   *prometheus.CounterVec
	ProjectsLoadSuccess  prometheus.Counter
	ProjectsLoadDuration prometheus.Histogram
	ProjectsCount        prometheus.Gauge

	// 💼 Experience API metrics
	ExperienceLoadErrors   *prometheus.CounterVec
	ExperienceLoadSuccess  prometheus.Counter
	ExperienceLoadDuration prometheus.Histogram

	// 🛠️ Skills API metrics
	SkillsLoadErrors   *prometheus.CounterVec
	SkillsLoadSuccess  prometheus.Counter
	SkillsLoadDuration prometheus.Histogram

	// 💾 Database metrics
	DatabaseConnectionErrors prometheus.Counter
	DatabaseQueryErrors      *prometheus.CounterVec
	DatabaseQueryDuration    prometheus.Histogram
	DatabaseActiveConns      prometheus.Gauge

	// 🔴 Redis metrics
	RedisOperationErrors   *prometheus.CounterVec
	RedisOperationDuration prometheus.Histogram
	RedisCacheHits         prometheus.Counter
	RedisCacheMisses       prometheus.Counter

	// 📦 MinIO metrics
	MinIOOperationErrors   *prometheus.CounterVec
	MinIOOperationDuration prometheus.Histogram
	MinIOUploadSize        prometheus.Histogram
	MinIODownloadSize      prometheus.Histogram

	// 🤖 Agent-SRE proxy metrics
	AgentSRERequestErrors   *prometheus.CounterVec
	AgentSRERequestDuration prometheus.Histogram

	// 🤖 Agent Bruno proxy metrics (Homepage chatbot and knowledge assistant)
	AgentBrunoRequestErrors   *prometheus.CounterVec
	AgentBrunoRequestDuration prometheus.Histogram
)

// InitMetrics initializes all Prometheus metrics
func InitMetrics() error {
	// ============================================================================
	// 🎯 HTTP Request metrics
	// ============================================================================
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "path", "status_code"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path", "status_code"},
	)

	HTTPActiveRequests = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "http_active_requests",
			Help:      "Number of active HTTP requests",
		},
		[]string{"method", "path"},
	)

	HTTPRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_size_bytes",
			Help:      "HTTP request size in bytes",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path"},
	)

	HTTPResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_response_size_bytes",
			Help:      "HTTP response size in bytes",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path"},
	)

	// ============================================================================
	// 📦 Projects API metrics
	// ============================================================================
	ProjectsLoadErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "projects_load_errors_total",
			Help:      "Total number of errors when loading projects from database",
		},
		[]string{"error_type"},
	)

	ProjectsLoadSuccess = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "projects_load_success_total",
			Help:      "Total number of successful project loads from database",
		},
	)

	ProjectsLoadDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "projects_load_duration_seconds",
			Help:      "Time taken to load projects from database",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
	)

	ProjectsCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "projects_count",
			Help:      "Total number of projects in database",
		},
	)

	// ============================================================================
	// 💼 Experience API metrics
	// ============================================================================
	ExperienceLoadErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "experience_load_errors_total",
			Help:      "Total number of errors when loading experience data from database",
		},
		[]string{"error_type"},
	)

	ExperienceLoadSuccess = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "experience_load_success_total",
			Help:      "Total number of successful experience data loads from database",
		},
	)

	ExperienceLoadDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "experience_load_duration_seconds",
			Help:      "Time taken to load experience data from database",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
	)

	// ============================================================================
	// 🛠️ Skills API metrics
	// ============================================================================
	SkillsLoadErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "skills_load_errors_total",
			Help:      "Total number of errors when loading skills from database",
		},
		[]string{"error_type"},
	)

	SkillsLoadSuccess = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "skills_load_success_total",
			Help:      "Total number of successful skills loads from database",
		},
	)

	SkillsLoadDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "skills_load_duration_seconds",
			Help:      "Time taken to load skills from database",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
	)

	// ============================================================================
	// 💾 Database metrics
	// ============================================================================
	DatabaseConnectionErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "database_connection_errors_total",
			Help:      "Total number of database connection errors",
		},
	)

	DatabaseQueryErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "database_query_errors_total",
			Help:      "Total number of database query errors",
		},
		[]string{"operation", "table"},
	)

	DatabaseQueryDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "database_query_duration_seconds",
			Help:      "Database query duration in seconds",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
	)

	DatabaseActiveConns = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "database_active_connections",
			Help:      "Number of active database connections",
		},
	)

	// ============================================================================
	// 🔴 Redis metrics
	// ============================================================================
	RedisOperationErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "redis_operation_errors_total",
			Help:      "Total number of Redis operation errors",
		},
		[]string{"operation"},
	)

	RedisOperationDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "redis_operation_duration_seconds",
			Help:      "Redis operation duration in seconds",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
	)

	RedisCacheHits = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "redis_cache_hits_total",
			Help:      "Total number of Redis cache hits",
		},
	)

	RedisCacheMisses = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "redis_cache_misses_total",
			Help:      "Total number of Redis cache misses",
		},
	)

	// ============================================================================
	// 📦 MinIO metrics
	// ============================================================================
	MinIOOperationErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "minio_operation_errors_total",
			Help:      "Total number of MinIO operation errors",
		},
		[]string{"operation"},
	)

	MinIOOperationDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "minio_operation_duration_seconds",
			Help:      "MinIO operation duration in seconds",
			Buckets:   []float64{0.01, 0.05, 0.1, 0.5, 1, 2.5, 5, 10},
		},
	)

	MinIOUploadSize = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "minio_upload_size_bytes",
			Help:      "MinIO upload size in bytes",
			Buckets:   prometheus.ExponentialBuckets(1024, 10, 8),
		},
	)

	MinIODownloadSize = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "minio_download_size_bytes",
			Help:      "MinIO download size in bytes",
			Buckets:   prometheus.ExponentialBuckets(1024, 10, 8),
		},
	)

	// ============================================================================
	// 🤖 Agent-SRE proxy metrics
	// ============================================================================
	AgentSRERequestErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "agent_sre_request_errors_total",
			Help:      "Total number of Agent-SRE proxy request errors",
		},
		[]string{"endpoint", "error_type"},
	)

	AgentSRERequestDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "agent_sre_request_duration_seconds",
			Help:      "Agent-SRE proxy request duration in seconds",
			Buckets:   []float64{0.1, 0.5, 1, 2.5, 5, 10, 30},
		},
	)

	// ============================================================================
	// 🤖 Agent Bruno proxy metrics
	// ============================================================================
	AgentBrunoRequestErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "agent_bruno_request_errors_total",
			Help:      "Total number of Agent Bruno proxy request errors",
		},
		[]string{"endpoint", "error_type"},
	)

	AgentBrunoRequestDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "agent_bruno_request_duration_seconds",
			Help:      "Agent Bruno proxy request duration in seconds",
			Buckets:   []float64{0.1, 0.5, 1, 2.5, 5, 10, 30},
		},
	)

	log.Println("✅ Prometheus metrics initialized")
	return nil
}

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
