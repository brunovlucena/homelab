package metrics

import (
	"database/sql"
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// =============================================================================
// 📊 DATABASE PROMETHEUS METRICS
// =============================================================================

var (
	// DbQueriesTotal tracks total number of database queries
	DbQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table"},
	)

	// DbQueryDuration tracks database query duration
	DbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"operation", "table"},
	)

	// DbQueryErrors tracks database query errors
	DbQueryErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_query_errors_total",
			Help: "Total number of database query errors",
		},
		[]string{"operation", "table"},
	)

	// =============================================================================
	// 🔌 CONNECTION POOL METRICS (SRE-003)
	// =============================================================================

	// DbConnectionWaitDuration tracks how long requests wait for a connection
	DbConnectionWaitDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "db_connection_wait_duration_seconds",
			Help:    "Time spent waiting for a database connection from the pool",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
	)

	// DbConnectionErrorsTotal tracks connection-level errors (not query errors)
	DbConnectionErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_connection_errors_total",
			Help: "Total number of database connection errors",
		},
		[]string{"error_type"}, // "timeout", "refused", "pool_exhausted", "other"
	)

	// DbPoolMaxOpen shows the configured max open connections
	DbPoolMaxOpen = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_pool_max_open_connections",
			Help: "Maximum number of open connections to the database (configured)",
		},
	)

	// DbPoolMaxIdle shows the configured max idle connections
	DbPoolMaxIdle = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_pool_max_idle_connections",
			Help: "Maximum number of idle connections in the pool (configured)",
		},
	)

	// DbPoolOpenConnections shows current open connections (from db.Stats())
	DbPoolOpenConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_pool_open_connections",
			Help: "Current number of open connections (in use + idle)",
		},
	)

	// DbPoolInUseConnections shows connections currently in use (from db.Stats())
	DbPoolInUseConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_pool_in_use_connections",
			Help: "Number of connections currently in use",
		},
	)

	// DbPoolIdleConnections shows idle connections (from db.Stats())
	DbPoolIdleConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_pool_idle_connections",
			Help: "Number of idle connections",
		},
	)

	// DbPoolWaitCount tracks total number of times a connection was waited for
	DbPoolWaitCount = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "db_pool_wait_count_total",
			Help: "Total number of connections waited for",
		},
	)

	// DbPoolWaitDuration tracks total time spent waiting for connections
	DbPoolWaitDuration = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "db_pool_wait_duration_seconds_total",
			Help: "Total time blocked waiting for a new connection",
		},
	)

	// DbPoolMaxIdleClosed tracks connections closed due to SetMaxIdleConns
	DbPoolMaxIdleClosed = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "db_pool_max_idle_closed_total",
			Help: "Total number of connections closed due to SetMaxIdleConns",
		},
	)

	// DbPoolMaxLifetimeClosed tracks connections closed due to SetConnMaxLifetime
	DbPoolMaxLifetimeClosed = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "db_pool_max_lifetime_closed_total",
			Help: "Total number of connections closed due to SetConnMaxLifetime",
		},
	)

	// DbPoolMaxIdleTimeClosed tracks connections closed due to SetConnMaxIdleTime
	DbPoolMaxIdleTimeClosed = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "db_pool_max_idle_time_closed_total",
			Help: "Total number of connections closed due to SetConnMaxIdleTime",
		},
	)

	// DbPoolUtilization tracks pool utilization as a percentage
	DbPoolUtilization = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_pool_utilization_ratio",
			Help: "Connection pool utilization (in_use / max_open)",
		},
	)

	// Track previous stats for delta calculations
	prevWaitCount         int64
	prevWaitDuration      time.Duration
	prevMaxIdleClosed     int64
	prevMaxLifetimeClosed int64
	prevMaxIdleTimeClosed int64
)

// RecordDatabaseMetrics records database operation metrics
// This function should be called after every database operation
// Parameters:
//   - operation: The type of SQL operation (SELECT, INSERT, UPDATE, DELETE, UPSERT)
//   - table: The database table being operated on
//   - duration: How long the query took
//   - success: Whether the query succeeded
func RecordDatabaseMetrics(operation, table string, duration time.Duration, success bool) {
	DbQueriesTotal.WithLabelValues(operation, table).Inc()
	DbQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())

	if !success {
		DbQueryErrors.WithLabelValues(operation, table).Inc()
	}
}

// RecordConnectionWait records time spent waiting for a connection
func RecordConnectionWait(duration time.Duration) {
	DbConnectionWaitDuration.Observe(duration.Seconds())
}

// RecordConnectionError records a connection error by type
func RecordConnectionError(errorType string) {
	DbConnectionErrorsTotal.WithLabelValues(errorType).Inc()
}

// SetPoolConfig records the configured pool settings
func SetPoolConfig(maxOpen, maxIdle int) {
	DbPoolMaxOpen.Set(float64(maxOpen))
	DbPoolMaxIdle.Set(float64(maxIdle))
}

// UpdatePoolStats updates connection pool metrics from db.Stats()
// This should be called periodically (e.g., every 15 seconds)
func UpdatePoolStats(db *sql.DB, maxOpen int) {
	if db == nil {
		return
	}

	stats := db.Stats()

	// Update gauge metrics directly
	DbPoolOpenConnections.Set(float64(stats.OpenConnections))
	DbPoolInUseConnections.Set(float64(stats.InUse))
	DbPoolIdleConnections.Set(float64(stats.Idle))

	// Calculate utilization
	if maxOpen > 0 {
		utilization := float64(stats.InUse) / float64(maxOpen)
		DbPoolUtilization.Set(utilization)
	}

	// Update counters with deltas (stats are cumulative)
	if stats.WaitCount > prevWaitCount {
		DbPoolWaitCount.Add(float64(stats.WaitCount - prevWaitCount))
		prevWaitCount = stats.WaitCount
	}

	if stats.WaitDuration > prevWaitDuration {
		DbPoolWaitDuration.Add((stats.WaitDuration - prevWaitDuration).Seconds())
		prevWaitDuration = stats.WaitDuration
	}

	if stats.MaxIdleClosed > prevMaxIdleClosed {
		DbPoolMaxIdleClosed.Add(float64(stats.MaxIdleClosed - prevMaxIdleClosed))
		prevMaxIdleClosed = stats.MaxIdleClosed
	}

	if stats.MaxLifetimeClosed > prevMaxLifetimeClosed {
		DbPoolMaxLifetimeClosed.Add(float64(stats.MaxLifetimeClosed - prevMaxLifetimeClosed))
		prevMaxLifetimeClosed = stats.MaxLifetimeClosed
	}

	if stats.MaxIdleTimeClosed > prevMaxIdleTimeClosed {
		DbPoolMaxIdleTimeClosed.Add(float64(stats.MaxIdleTimeClosed - prevMaxIdleTimeClosed))
		prevMaxIdleTimeClosed = stats.MaxIdleTimeClosed
	}

	log.Printf("📊 DB Pool Stats: open=%d, in_use=%d, idle=%d, wait_count=%d, utilization=%.2f%%",
		stats.OpenConnections, stats.InUse, stats.Idle, stats.WaitCount,
		float64(stats.InUse)/float64(maxOpen)*100)
}

// StartPoolStatsCollector starts a goroutine that periodically collects pool stats
func StartPoolStatsCollector(db *sql.DB, maxOpen int, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Collect initial stats
		UpdatePoolStats(db, maxOpen)

		for range ticker.C {
			UpdatePoolStats(db, maxOpen)
		}
	}()
	log.Printf("📊 Started connection pool stats collector (interval: %v)", interval)
}
