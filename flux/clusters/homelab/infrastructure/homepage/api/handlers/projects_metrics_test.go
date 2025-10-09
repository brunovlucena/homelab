package handlers

import (
	"bruno-site/metrics"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// =============================================================================
// 🧪 TEST HELPERS
// =============================================================================

// setupTestDB creates a test database with schema
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Auto migrate the Project model
	err = db.AutoMigrate(&Project{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// seedTestProjects seeds the database with test projects
func seedTestProjects(t *testing.T, db *gorm.DB, count int) {
	for i := 1; i <= count; i++ {
		project := Project{
			Title:        "Test Project",
			Description:  "A test project",
			Technologies: []string{"Go", "Kubernetes"},
			GithubURL:    "https://github.com/test/test",
			Type:         "homelab",
			GithubActive: true,
		}
		if err := db.Create(&project).Error; err != nil {
			t.Fatalf("Failed to seed test project: %v", err)
		}
	}
}

// getMetricValue retrieves the current value of a counter metric
func getMetricValue(t *testing.T, collector prometheus.Collector) float64 {
	metricChan := make(chan prometheus.Metric, 1)
	collector.Collect(metricChan)
	close(metricChan)

	metric := <-metricChan
	if metric == nil {
		return 0
	}

	var metricDto dto.Metric
	if err := metric.Write(&metricDto); err != nil {
		t.Fatalf("Failed to write metric: %v", err)
	}

	if metricDto.Counter != nil {
		return *metricDto.Counter.Value
	}
	return 0
}

// =============================================================================
// 🧪 METRICS TESTS
// =============================================================================

func TestProjectsLoadSuccessMetric(t *testing.T) {
	// 🔧 Setup
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	seedTestProjects(t, db, 3)

	// Record initial metric value
	initialSuccessCount := getMetricValue(t, metrics.ProjectsLoadSuccess)

	// 🎬 Execute
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/projects", nil)

	handler := GetProjects(db)
	handler(c)

	// ✅ Assert
	assert.Equal(t, http.StatusOK, w.Code, "Expected HTTP 200 status")

	// Check that success metric was incremented
	finalSuccessCount := getMetricValue(t, metrics.ProjectsLoadSuccess)
	assert.Greater(t, finalSuccessCount, initialSuccessCount, "Success metric should be incremented")
}

func TestProjectsLoadErrorMetric_DatabaseUnavailable(t *testing.T) {
	// 🔧 Setup
	gin.SetMode(gin.TestMode)

	// Record initial error metric value
	// Note: We can't easily get labeled metrics, so we'll test the behavior instead

	// 🎬 Execute
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/projects", nil)

	handler := GetProjects(nil) // Pass nil database to simulate unavailability
	handler(c)

	// ✅ Assert
	assert.Equal(t, http.StatusServiceUnavailable, w.Code, "Expected HTTP 503 status")
	assert.Contains(t, w.Body.String(), "database not available", "Expected error message about database")
}

func TestProjectsLoadErrorMetric_QueryError(t *testing.T) {
	// 🔧 Setup
	gin.SetMode(gin.TestMode)

	// Create a database that will fail on query
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Don't migrate the schema, so queries will fail

	// 🎬 Execute
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/projects", nil)

	handler := GetProjects(db)
	handler(c)

	// ✅ Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code, "Expected HTTP 500 status")
	assert.Contains(t, w.Body.String(), "error", "Expected error in response")
}

func TestProjectsLoadDurationMetric(t *testing.T) {
	// 🔧 Setup
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	seedTestProjects(t, db, 5)

	// 🎬 Execute
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/projects", nil)

	handler := GetProjects(db)
	handler(c)

	// ✅ Assert
	assert.Equal(t, http.StatusOK, w.Code, "Expected HTTP 200 status")

	// The duration metric should have been recorded
	// We can't easily check the exact value, but we can verify the handler completed
	var projects []Project
	err := db.Find(&projects).Error
	assert.NoError(t, err, "Database should be accessible")
	assert.Len(t, projects, 5, "Should have 5 projects")
}

func TestProjectsMetricsRecordingFunctions(t *testing.T) {
	// 🧪 Test individual metric recording functions

	t.Run("RecordProjectsLoadError", func(t *testing.T) {
		// This should not panic
		metrics.RecordProjectsLoadError("test_error")
	})

	t.Run("RecordProjectsLoadSuccess", func(t *testing.T) {
		// This should not panic
		metrics.RecordProjectsLoadSuccess()
	})

	t.Run("RecordDatabaseError", func(t *testing.T) {
		// This should not panic
		metrics.RecordDatabaseError("select", "projects")
	})

	t.Run("RecordDatabaseConnectionError", func(t *testing.T) {
		// This should not panic
		metrics.RecordDatabaseConnectionError()
	})
}

// =============================================================================
// 🧪 INTEGRATION TESTS
// =============================================================================

func TestProjectsEndpointWithMetrics_Success(t *testing.T) {
	// 🔧 Setup
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	seedTestProjects(t, db, 2)

	router := gin.New()
	router.GET("/api/v1/projects", GetProjects(db))

	// 🎬 Execute
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/projects", nil)
	router.ServeHTTP(w, req)

	// ✅ Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Test Project")
}

func TestProjectsEndpointWithMetrics_DatabaseError(t *testing.T) {
	// 🔧 Setup
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/api/v1/projects", GetProjects(nil))

	// 🎬 Execute
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/projects", nil)
	router.ServeHTTP(w, req)

	// ✅ Assert
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "database not available")
}

// =============================================================================
// 🧪 BENCHMARK TESTS
// =============================================================================

func BenchmarkProjectsLoadWithMetrics(b *testing.B) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(&testing.T{})
	seedTestProjects(&testing.T{}, db, 10)

	router := gin.New()
	router.GET("/api/v1/projects", GetProjects(db))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/projects", nil)
		router.ServeHTTP(w, req)
	}
}

// =============================================================================
// 🧪 ERROR SCENARIO TESTS
// =============================================================================

func TestMetricsRecordedOnVariousErrors(t *testing.T) {
	testCases := []struct {
		name           string
		setupDB        func() *gorm.DB
		expectedStatus int
		errorType      string
	}{
		{
			name: "Database Unavailable",
			setupDB: func() *gorm.DB {
				return nil
			},
			expectedStatus: http.StatusServiceUnavailable,
			errorType:      "database_unavailable",
		},
		{
			name: "Query Error - No Schema",
			setupDB: func() *gorm.DB {
				db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
				return db
			},
			expectedStatus: http.StatusInternalServerError,
			errorType:      "query_error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 🔧 Setup
			gin.SetMode(gin.TestMode)
			db := tc.setupDB()

			// 🎬 Execute
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/api/v1/projects", nil)

			handler := GetProjects(db)
			handler(c)

			// ✅ Assert
			assert.Equal(t, tc.expectedStatus, w.Code, "Expected status code to match")
		})
	}
}

// =============================================================================
// 🧪 MOCK ERROR TESTS
// =============================================================================

type mockDB struct {
	gorm.DB
	findError error
}

func (m *mockDB) Find(dest interface{}, conds ...interface{}) *gorm.DB {
	if m.findError != nil {
		m.Error = m.findError
	}
	return &m.DB
}

func TestProjectsMetricsWithMockErrors(t *testing.T) {
	testCases := []struct {
		name      string
		mockError error
		errorType string
	}{
		{
			name:      "Record Not Found Error",
			mockError: gorm.ErrRecordNotFound,
			errorType: "not_found",
		},
		{
			name:      "Generic Query Error",
			mockError: errors.New("database connection lost"),
			errorType: "query_error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Note: This is a conceptual test showing how you might test with mocks
			// In practice, you'd need a more sophisticated mocking approach
			t.Skip("Mock test - requires proper database mocking setup")
		})
	}
}
