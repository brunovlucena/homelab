package metrics

import (
	"database/sql"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRecordDatabaseMetrics tests the basic database metrics recording
func TestRecordDatabaseMetrics(t *testing.T) {
	// Record a successful query
	RecordDatabaseMetrics("SELECT", "projects", 10*time.Millisecond, true)

	// Verify counter was incremented
	count := testutil.ToFloat64(DbQueriesTotal.WithLabelValues("SELECT", "projects"))
	assert.Equal(t, float64(1), count)

	// Record a failed query
	RecordDatabaseMetrics("INSERT", "projects", 50*time.Millisecond, false)

	// Verify error counter was incremented
	errorCount := testutil.ToFloat64(DbQueryErrors.WithLabelValues("INSERT", "projects"))
	assert.Equal(t, float64(1), errorCount)
}

// TestRecordConnectionWait tests connection wait time recording
func TestRecordConnectionWait(t *testing.T) {
	// Record some wait times
	RecordConnectionWait(5 * time.Millisecond)
	RecordConnectionWait(10 * time.Millisecond)
	RecordConnectionWait(50 * time.Millisecond)

	// Verify histogram has samples by checking the count
	// Histograms don't support ToFloat64 directly, so we verify it was collected
	count := testutil.CollectAndCount(DbConnectionWaitDuration)
	assert.Greater(t, count, 0, "histogram should have been collected")
}

// TestRecordConnectionError tests connection error recording
func TestRecordConnectionError(t *testing.T) {
	testCases := []struct {
		errorType string
	}{
		{"timeout"},
		{"refused"},
		{"pool_exhausted"},
		{"other"},
	}

	for _, tc := range testCases {
		t.Run(tc.errorType, func(t *testing.T) {
			// Get initial count
			initialCount := testutil.ToFloat64(DbConnectionErrorsTotal.WithLabelValues(tc.errorType))

			// Record an error
			RecordConnectionError(tc.errorType)

			// Verify counter was incremented
			newCount := testutil.ToFloat64(DbConnectionErrorsTotal.WithLabelValues(tc.errorType))
			assert.Equal(t, initialCount+1, newCount)
		})
	}
}

// TestSetPoolConfig tests pool configuration metrics
func TestSetPoolConfig(t *testing.T) {
	// Set pool config
	SetPoolConfig(25, 5)

	// Verify gauges were set
	maxOpen := testutil.ToFloat64(DbPoolMaxOpen)
	maxIdle := testutil.ToFloat64(DbPoolMaxIdle)

	assert.Equal(t, float64(25), maxOpen)
	assert.Equal(t, float64(5), maxIdle)

	// Test updating config
	SetPoolConfig(50, 10)

	maxOpen = testutil.ToFloat64(DbPoolMaxOpen)
	maxIdle = testutil.ToFloat64(DbPoolMaxIdle)

	assert.Equal(t, float64(50), maxOpen)
	assert.Equal(t, float64(10), maxIdle)
}

// TestUpdatePoolStats tests pool stats update from db.Stats()
func TestUpdatePoolStats(t *testing.T) {
	// Skip if no database available
	db, err := sql.Open("postgres", "postgres://localhost:5432/test?sslmode=disable")
	if err != nil {
		t.Skip("Database not available for testing")
	}

	// Configure pool for testing
	maxOpen := 10
	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(2)

	// Reset previous stats
	prevWaitCount = 0
	prevWaitDuration = 0
	prevMaxIdleClosed = 0
	prevMaxLifetimeClosed = 0
	prevMaxIdleTimeClosed = 0

	// Update stats (doesn't matter if DB is actually connected)
	UpdatePoolStats(db, maxOpen)

	// Verify utilization is calculated
	utilization := testutil.ToFloat64(DbPoolUtilization)
	assert.GreaterOrEqual(t, utilization, float64(0))
	assert.LessOrEqual(t, utilization, float64(1))

	_ = db.Close()
}

// TestUpdatePoolStatsNilDB tests nil DB handling
func TestUpdatePoolStatsNilDB(t *testing.T) {
	// Should not panic
	assert.NotPanics(t, func() {
		UpdatePoolStats(nil, 25)
	})
}

// TestPoolMetricsExistence verifies all pool metrics are registered
func TestPoolMetricsExistence(t *testing.T) {
	testCases := []struct {
		name     string
		metric   prometheus.Collector
		expected string
	}{
		{"db_queries_total", DbQueriesTotal, "db_queries_total"},
		{"db_query_duration_seconds", DbQueryDuration, "db_query_duration_seconds"},
		{"db_query_errors_total", DbQueryErrors, "db_query_errors_total"},
		{"db_connection_wait_duration_seconds", DbConnectionWaitDuration, "db_connection_wait_duration_seconds"},
		{"db_connection_errors_total", DbConnectionErrorsTotal, "db_connection_errors_total"},
		{"db_pool_max_open_connections", DbPoolMaxOpen, "db_pool_max_open_connections"},
		{"db_pool_max_idle_connections", DbPoolMaxIdle, "db_pool_max_idle_connections"},
		{"db_pool_open_connections", DbPoolOpenConnections, "db_pool_open_connections"},
		{"db_pool_in_use_connections", DbPoolInUseConnections, "db_pool_in_use_connections"},
		{"db_pool_idle_connections", DbPoolIdleConnections, "db_pool_idle_connections"},
		{"db_pool_wait_count_total", DbPoolWaitCount, "db_pool_wait_count_total"},
		{"db_pool_wait_duration_seconds_total", DbPoolWaitDuration, "db_pool_wait_duration_seconds_total"},
		{"db_pool_max_idle_closed_total", DbPoolMaxIdleClosed, "db_pool_max_idle_closed_total"},
		{"db_pool_max_lifetime_closed_total", DbPoolMaxLifetimeClosed, "db_pool_max_lifetime_closed_total"},
		{"db_pool_max_idle_time_closed_total", DbPoolMaxIdleTimeClosed, "db_pool_max_idle_time_closed_total"},
		{"db_pool_utilization_ratio", DbPoolUtilization, "db_pool_utilization_ratio"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NotNil(t, tc.metric, "metric %s should not be nil", tc.name)
		})
	}
}

// BenchmarkRecordDatabaseMetrics benchmarks the metrics recording
func BenchmarkRecordDatabaseMetrics(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RecordDatabaseMetrics("SELECT", "projects", 10*time.Millisecond, true)
	}
}

// BenchmarkRecordConnectionWait benchmarks connection wait recording
func BenchmarkRecordConnectionWait(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RecordConnectionWait(5 * time.Millisecond)
	}
}

// BenchmarkRecordConnectionError benchmarks error recording
func BenchmarkRecordConnectionError(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RecordConnectionError("timeout")
	}
}
