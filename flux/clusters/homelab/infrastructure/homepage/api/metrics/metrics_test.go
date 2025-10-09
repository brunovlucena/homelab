package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// 🧪 TEST HELPERS
// =============================================================================

// getCounterValue retrieves the current value of a counter metric
func getCounterValue(t *testing.T, collector prometheus.Collector) float64 {
	metricChan := make(chan prometheus.Metric, 10)
	collector.Collect(metricChan)
	close(metricChan)

	var total float64
	for metric := range metricChan {
		var metricDto dto.Metric
		if err := metric.Write(&metricDto); err != nil {
			t.Fatalf("Failed to write metric: %v", err)
		}

		if metricDto.Counter != nil {
			total += *metricDto.Counter.Value
		}
	}

	return total
}

// =============================================================================
// 📊 PROJECTS METRICS TESTS
// =============================================================================

func TestProjectsLoadErrorMetric(t *testing.T) {
	// Record initial value
	initialValue := getCounterValue(t, ProjectsLoadErrors)

	// Record some errors
	RecordProjectsLoadError("database_unavailable")
	RecordProjectsLoadError("query_error")
	RecordProjectsLoadError("not_found")

	// Verify counter increased
	finalValue := getCounterValue(t, ProjectsLoadErrors)
	assert.Greater(t, finalValue, initialValue, "Projects load errors counter should increase")
}

func TestProjectsLoadSuccessMetric(t *testing.T) {
	// Record initial value
	initialValue := getCounterValue(t, ProjectsLoadSuccess)

	// Record success
	RecordProjectsLoadSuccess()
	RecordProjectsLoadSuccess()

	// Verify counter increased
	finalValue := getCounterValue(t, ProjectsLoadSuccess)
	assert.Greater(t, finalValue, initialValue, "Projects load success counter should increase")
}

func TestProjectsLoadDurationMetric(t *testing.T) {
	// Record some durations
	ProjectsLoadDuration.Observe(0.001) // 1ms
	ProjectsLoadDuration.Observe(0.010) // 10ms
	ProjectsLoadDuration.Observe(0.100) // 100ms

	// Histogram should have recorded observations
	// We can't easily check the exact values, but we can verify it doesn't panic
	assert.NotNil(t, ProjectsLoadDuration, "Projects load duration metric should be initialized")
}

// =============================================================================
// 💼 EXPERIENCE METRICS TESTS
// =============================================================================

func TestExperienceLoadErrorMetric(t *testing.T) {
	// Record initial value
	initialValue := getCounterValue(t, ExperienceLoadErrors)

	// Record some errors
	RecordExperienceLoadError("database_unavailable")
	RecordExperienceLoadError("database_query_error")

	// Verify counter increased
	finalValue := getCounterValue(t, ExperienceLoadErrors)
	assert.Greater(t, finalValue, initialValue, "Experience load errors counter should increase")
}

func TestExperienceLoadSuccessMetric(t *testing.T) {
	// Record initial value
	initialValue := getCounterValue(t, ExperienceLoadSuccess)

	// Record success
	RecordExperienceLoadSuccess()

	// Verify counter increased
	finalValue := getCounterValue(t, ExperienceLoadSuccess)
	assert.Greater(t, finalValue, initialValue, "Experience load success counter should increase")
}

func TestExperienceLoadDurationMetric(t *testing.T) {
	// Record some durations
	ExperienceLoadDuration.Observe(0.005) // 5ms
	ExperienceLoadDuration.Observe(0.050) // 50ms

	// Histogram should have recorded observations
	assert.NotNil(t, ExperienceLoadDuration, "Experience load duration metric should be initialized")
}

// =============================================================================
// 💾 DATABASE METRICS TESTS
// =============================================================================

func TestDatabaseConnectionErrorMetric(t *testing.T) {
	// Record initial value
	initialValue := getCounterValue(t, DatabaseConnectionErrors)

	// Record connection error
	RecordDatabaseConnectionError()

	// Verify counter increased
	finalValue := getCounterValue(t, DatabaseConnectionErrors)
	assert.Greater(t, finalValue, initialValue, "Database connection errors counter should increase")
}

func TestDatabaseQueryErrorMetric(t *testing.T) {
	// Record initial value
	initialValue := getCounterValue(t, DatabaseQueryErrors)

	// Record query errors
	RecordDatabaseError("select", "projects")
	RecordDatabaseError("insert", "experiences")

	// Verify counter increased
	finalValue := getCounterValue(t, DatabaseQueryErrors)
	assert.Greater(t, finalValue, initialValue, "Database query errors counter should increase")
}

// =============================================================================
// 🔴 REDIS METRICS TESTS
// =============================================================================

func TestRedisOperationErrorMetric(t *testing.T) {
	// Record initial value
	initialValue := getCounterValue(t, RedisOperationErrors)

	// Record Redis errors
	RecordRedisError("get")
	RecordRedisError("set")

	// Verify counter increased
	finalValue := getCounterValue(t, RedisOperationErrors)
	assert.Greater(t, finalValue, initialValue, "Redis operation errors counter should increase")
}

// =============================================================================
// 📦 MINIO METRICS TESTS
// =============================================================================

func TestMinIOOperationErrorMetric(t *testing.T) {
	// Record initial value
	initialValue := getCounterValue(t, MinIOOperationErrors)

	// Record MinIO errors
	RecordMinIOError("get_object")
	RecordMinIOError("put_object")

	// Verify counter increased
	finalValue := getCounterValue(t, MinIOOperationErrors)
	assert.Greater(t, finalValue, initialValue, "MinIO operation errors counter should increase")
}

// =============================================================================
// 🤖 AGENT-SRE METRICS TESTS
// =============================================================================

func TestAgentSREErrorMetric(t *testing.T) {
	// Record initial value
	initialValue := getCounterValue(t, AgentSRERequestErrors)

	// Record Agent-SRE errors
	RecordAgentSREError("/chat", "timeout")
	RecordAgentSREError("/mcp/chat", "connection_refused")

	// Verify counter increased
	finalValue := getCounterValue(t, AgentSRERequestErrors)
	assert.Greater(t, finalValue, initialValue, "Agent-SRE request errors counter should increase")
}

func TestAgentSREDurationMetric(t *testing.T) {
	// Record some durations
	AgentSRERequestDuration.WithLabelValues("/chat").Observe(1.5)
	AgentSRERequestDuration.WithLabelValues("/mcp/chat").Observe(2.5)

	// Histogram should have recorded observations
	assert.NotNil(t, AgentSRERequestDuration, "Agent-SRE request duration metric should be initialized")
}

// =============================================================================
// 🔧 METRIC HELPER FUNCTION TESTS
// =============================================================================

func TestAllHelperFunctionsDoNotPanic(t *testing.T) {
	t.Run("RecordProjectsLoadError", func(t *testing.T) {
		assert.NotPanics(t, func() {
			RecordProjectsLoadError("test_error")
		})
	})

	t.Run("RecordProjectsLoadSuccess", func(t *testing.T) {
		assert.NotPanics(t, func() {
			RecordProjectsLoadSuccess()
		})
	})

	t.Run("RecordExperienceLoadError", func(t *testing.T) {
		assert.NotPanics(t, func() {
			RecordExperienceLoadError("test_error")
		})
	})

	t.Run("RecordExperienceLoadSuccess", func(t *testing.T) {
		assert.NotPanics(t, func() {
			RecordExperienceLoadSuccess()
		})
	})

	t.Run("RecordDatabaseError", func(t *testing.T) {
		assert.NotPanics(t, func() {
			RecordDatabaseError("select", "test_table")
		})
	})

	t.Run("RecordDatabaseConnectionError", func(t *testing.T) {
		assert.NotPanics(t, func() {
			RecordDatabaseConnectionError()
		})
	})

	t.Run("RecordRedisError", func(t *testing.T) {
		assert.NotPanics(t, func() {
			RecordRedisError("test_operation")
		})
	})

	t.Run("RecordMinIOError", func(t *testing.T) {
		assert.NotPanics(t, func() {
			RecordMinIOError("test_operation")
		})
	})

	t.Run("RecordAgentSREError", func(t *testing.T) {
		assert.NotPanics(t, func() {
			RecordAgentSREError("/test", "test_error")
		})
	})
}

// =============================================================================
// 📊 METRIC INITIALIZATION TESTS
// =============================================================================

func TestAllMetricsAreInitialized(t *testing.T) {
	t.Run("HTTP Metrics", func(t *testing.T) {
		assert.NotNil(t, HTTPRequestsTotal, "HTTPRequestsTotal should be initialized")
		assert.NotNil(t, HTTPRequestDuration, "HTTPRequestDuration should be initialized")
	})

	t.Run("Projects Metrics", func(t *testing.T) {
		assert.NotNil(t, ProjectsLoadErrors, "ProjectsLoadErrors should be initialized")
		assert.NotNil(t, ProjectsLoadSuccess, "ProjectsLoadSuccess should be initialized")
		assert.NotNil(t, ProjectsLoadDuration, "ProjectsLoadDuration should be initialized")
	})

	t.Run("Experience Metrics", func(t *testing.T) {
		assert.NotNil(t, ExperienceLoadErrors, "ExperienceLoadErrors should be initialized")
		assert.NotNil(t, ExperienceLoadSuccess, "ExperienceLoadSuccess should be initialized")
		assert.NotNil(t, ExperienceLoadDuration, "ExperienceLoadDuration should be initialized")
	})

	t.Run("Database Metrics", func(t *testing.T) {
		assert.NotNil(t, DatabaseConnectionErrors, "DatabaseConnectionErrors should be initialized")
		assert.NotNil(t, DatabaseQueryErrors, "DatabaseQueryErrors should be initialized")
	})

	t.Run("Redis Metrics", func(t *testing.T) {
		assert.NotNil(t, RedisOperationErrors, "RedisOperationErrors should be initialized")
	})

	t.Run("MinIO Metrics", func(t *testing.T) {
		assert.NotNil(t, MinIOOperationErrors, "MinIOOperationErrors should be initialized")
	})

	t.Run("Agent-SRE Metrics", func(t *testing.T) {
		assert.NotNil(t, AgentSRERequestErrors, "AgentSRERequestErrors should be initialized")
		assert.NotNil(t, AgentSRERequestDuration, "AgentSRERequestDuration should be initialized")
	})
}

// =============================================================================
// 🏃 BENCHMARK TESTS
// =============================================================================

func BenchmarkRecordProjectsLoadError(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RecordProjectsLoadError("benchmark_error")
	}
}

func BenchmarkRecordProjectsLoadSuccess(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RecordProjectsLoadSuccess()
	}
}

func BenchmarkProjectsLoadDurationObserve(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ProjectsLoadDuration.Observe(0.001)
	}
}

func BenchmarkRecordExperienceLoadError(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RecordExperienceLoadError("benchmark_error")
	}
}

func BenchmarkAllMetricOperations(b *testing.B) {
	b.Run("Projects", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			RecordProjectsLoadSuccess()
			RecordProjectsLoadError("test")
			ProjectsLoadDuration.Observe(0.001)
		}
	})

	b.Run("Experience", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			RecordExperienceLoadSuccess()
			RecordExperienceLoadError("test")
			ExperienceLoadDuration.Observe(0.001)
		}
	})

	b.Run("Database", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			RecordDatabaseError("select", "table")
			RecordDatabaseConnectionError()
		}
	})
}

