package metrics

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// TestMain initializes metrics before running tests
func TestMain(m *testing.M) {
	// Initialize metrics for tests
	if err := InitMetrics(); err != nil {
		panic("Failed to initialize metrics: " + err.Error())
	}

	// Run tests
	code := m.Run()

	// Exit with the test result code
	os.Exit(code)
}

// =============================================================================
// 📊 PROJECTS METRICS TESTS
// =============================================================================

func TestProjectsLoadErrorMetric(t *testing.T) {
	// Verify metric recording doesn't panic
	assert.NotPanics(t, func() {
		RecordProjectsLoadError("database_unavailable")
		RecordProjectsLoadError("query_error")
		RecordProjectsLoadError("not_found")
	}, "Recording projects load errors should not panic")
}

func TestProjectsLoadSuccessMetric(t *testing.T) {
	// Verify metric recording doesn't panic
	assert.NotPanics(t, func() {
		RecordProjectsLoadSuccess()
		RecordProjectsLoadSuccess()
	}, "Recording projects load success should not panic")
}

func TestProjectsLoadDurationMetric(t *testing.T) {
	// Record some durations using OpenTelemetry API
	assert.NotPanics(t, func() {
		ProjectsLoadDuration.Record(context.Background(), 0.001) // 1ms
		ProjectsLoadDuration.Record(context.Background(), 0.010) // 10ms
		ProjectsLoadDuration.Record(context.Background(), 0.100) // 100ms
	}, "Recording projects load duration should not panic")

	// Verify histogram is initialized
	assert.NotNil(t, ProjectsLoadDuration, "Projects load duration metric should be initialized")
}

// =============================================================================
// 💼 EXPERIENCE METRICS TESTS
// =============================================================================

func TestExperienceLoadErrorMetric(t *testing.T) {
	// Verify metric recording doesn't panic
	assert.NotPanics(t, func() {
		RecordExperienceLoadError("database_unavailable")
		RecordExperienceLoadError("database_query_error")
	}, "Recording experience load errors should not panic")
}

func TestExperienceLoadSuccessMetric(t *testing.T) {
	// Verify metric recording doesn't panic
	assert.NotPanics(t, func() {
		RecordExperienceLoadSuccess()
	}, "Recording experience load success should not panic")
}

func TestExperienceLoadDurationMetric(t *testing.T) {
	// Record some durations using OpenTelemetry API
	assert.NotPanics(t, func() {
		ExperienceLoadDuration.Record(context.Background(), 0.005) // 5ms
		ExperienceLoadDuration.Record(context.Background(), 0.050) // 50ms
	}, "Recording experience load duration should not panic")

	// Verify histogram is initialized
	assert.NotNil(t, ExperienceLoadDuration, "Experience load duration metric should be initialized")
}

// =============================================================================
// 💾 DATABASE METRICS TESTS
// =============================================================================

func TestDatabaseConnectionErrorMetric(t *testing.T) {
	// Verify metric recording doesn't panic
	assert.NotPanics(t, func() {
		RecordDatabaseConnectionError()
	}, "Recording database connection errors should not panic")
}

func TestDatabaseQueryErrorMetric(t *testing.T) {
	// Verify metric recording doesn't panic
	assert.NotPanics(t, func() {
		RecordDatabaseError("select", "projects")
		RecordDatabaseError("insert", "experiences")
	}, "Recording database query errors should not panic")
}

// =============================================================================
// 🔴 REDIS METRICS TESTS
// =============================================================================

func TestRedisOperationErrorMetric(t *testing.T) {
	// Verify metric recording doesn't panic
	assert.NotPanics(t, func() {
		RecordRedisError("get")
		RecordRedisError("set")
	}, "Recording Redis operation errors should not panic")
}

// =============================================================================
// 📦 MINIO METRICS TESTS
// =============================================================================

func TestMinIOOperationErrorMetric(t *testing.T) {
	// Verify metric recording doesn't panic
	assert.NotPanics(t, func() {
		RecordMinIOError("get_object")
		RecordMinIOError("put_object")
	}, "Recording MinIO operation errors should not panic")
}

// =============================================================================
// 🤖 AGENT-SRE METRICS TESTS
// =============================================================================

func TestAgentSREErrorMetric(t *testing.T) {
	// Verify metric recording doesn't panic
	assert.NotPanics(t, func() {
		RecordAgentSREError("/chat", "timeout")
		RecordAgentSREError("/mcp/chat", "connection_refused")
	}, "Recording Agent-SRE request errors should not panic")
}

func TestAgentSREDurationMetric(t *testing.T) {
	// Record some durations using OpenTelemetry API with attributes
	assert.NotPanics(t, func() {
		AgentSRERequestDuration.Record(context.Background(), 1.5,
			metric.WithAttributes(attribute.String("endpoint", "/chat")))
		AgentSRERequestDuration.Record(context.Background(), 2.5,
			metric.WithAttributes(attribute.String("endpoint", "/mcp/chat")))
	}, "Recording Agent-SRE request duration should not panic")

	// Verify histogram is initialized
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

func BenchmarkProjectsLoadDurationRecord(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		ProjectsLoadDuration.Record(ctx, 0.001)
	}
}

func BenchmarkRecordExperienceLoadError(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RecordExperienceLoadError("benchmark_error")
	}
}

func BenchmarkAllMetricOperations(b *testing.B) {
	ctx := context.Background()

	b.Run("Projects", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			RecordProjectsLoadSuccess()
			RecordProjectsLoadError("test")
			ProjectsLoadDuration.Record(ctx, 0.001)
		}
	})

	b.Run("Experience", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			RecordExperienceLoadSuccess()
			RecordExperienceLoadError("test")
			ExperienceLoadDuration.Record(ctx, 0.001)
		}
	})

	b.Run("Database", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			RecordDatabaseError("select", "table")
			RecordDatabaseConnectionError()
		}
	})
}
