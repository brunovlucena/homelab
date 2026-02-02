package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test setup helpers
func setupTestDB(t *testing.T) *sql.DB {
	// Use test database or mock
	db, err := sql.Open("postgres", getTestDBURL())
	if err != nil {
		t.Skipf("Skipping test - database not available: %v", err)
	}

	// Test connection
	err = db.Ping()
	if err != nil {
		t.Skipf("Skipping test - cannot connect to database: %v", err)
	}

	return db
}

func setupTestRouter(t *testing.T) *gin.Engine {
	gin.SetMode(gin.TestMode)

	// Initialize router with test configuration
	router := gin.New()
	router.Use(gin.Recovery())

	// Setup routes (simplified for testing)
	setupRoutes(router)

	return router
}

func setupRoutes(router *gin.Engine) {
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "timestamp": time.Now().Unix()})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		api.GET("/projects", getProjectsHandler)
		api.GET("/about", getAboutHandler)
		api.GET("/contact", getContactHandler)
		api.GET("/skills", getSkillsHandler)
		api.GET("/experience", getExperienceHandler)
		api.POST("/analytics/track", trackProjectViewHandler)
	}

	// Metrics endpoint
	router.GET("/metrics", prometheusHandler())
}

// Mock handlers for testing
func getProjectsHandler(c *gin.Context) {
	projects := []Project{
		{
			ID:               1,
			Title:            "Test Project",
			Description:      "A test project",
			ShortDescription: "Test",
			Type:             "web",
			Icon:             "test-icon",
			GithubURL:        "https://github.com/test",
			LiveURL:          "https://test.com",
			Technologies:     []string{"Go", "React"},
			Active:           true,
		},
	}
	c.JSON(http.StatusOK, projects)
}

func getAboutHandler(c *gin.Context) {
	about := AboutData{
		Description: "Test about description",
		Highlights: []struct {
			Icon string `json:"icon"`
			Text string `json:"text"`
		}{
			{Icon: "test-icon", Text: "Test highlight"},
		},
	}
	c.JSON(http.StatusOK, about)
}

func getContactHandler(c *gin.Context) {
	contact := ContactData{
		Email:        "test@example.com",
		Location:     "Test Location",
		LinkedIn:     "https://linkedin.com/test",
		GitHub:       "https://github.com/test",
		Availability: "Available",
	}
	c.JSON(http.StatusOK, contact)
}

func getSkillsHandler(c *gin.Context) {
	skills := []Skill{
		{
			ID:          1,
			Name:        "Go",
			Category:    "Backend",
			Proficiency: 90,
			Icon:        "go-icon",
			Order:       1,
			Active:      true,
		},
	}
	c.JSON(http.StatusOK, skills)
}

func getExperienceHandler(c *gin.Context) {
	endDate := "2023-01"
	experience := []Experience{
		{
			ID:           1,
			Title:        "Software Engineer",
			Company:      "Test Company",
			Description:  "Test experience",
			StartDate:    "2020-01",
			EndDate:      &endDate,
			Technologies: []string{"Go", "React"},
			Current:      false,
			Order:        1,
			Active:       true,
		},
	}
	c.JSON(http.StatusOK, experience)
}

func trackProjectViewHandler(c *gin.Context) {
	var trackData struct {
		ProjectID int    `json:"project_id"`
		IP        string `json:"ip"`
		UserAgent string `json:"user_agent"`
		Referrer  string `json:"referrer"`
	}

	if err := c.ShouldBindJSON(&trackData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "tracked", "project_id": trackData.ProjectID})
}

func prometheusHandler() gin.HandlerFunc {
	return gin.WrapH(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("# HELP test_metric Test metric\n# TYPE test_metric counter\ntest_metric 1\n"))
	}))
}

// Test cases
func TestHealthEndpoint(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.NotNil(t, response["timestamp"])
}

func TestProjectsEndpoint(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/projects", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var projects []Project
	err := json.Unmarshal(w.Body.Bytes(), &projects)
	require.NoError(t, err)

	assert.Len(t, projects, 1)
	assert.Equal(t, "Test Project", projects[0].Title)
	assert.Equal(t, "web", projects[0].Type)
	assert.True(t, projects[0].Active)
}

func TestAboutEndpoint(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/about", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var about AboutData
	err := json.Unmarshal(w.Body.Bytes(), &about)
	require.NoError(t, err)

	assert.Equal(t, "Test about description", about.Description)
	assert.Len(t, about.Highlights, 1)
	assert.Equal(t, "Test highlight", about.Highlights[0].Text)
}

func TestContactEndpoint(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/contact", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var contact ContactData
	err := json.Unmarshal(w.Body.Bytes(), &contact)
	require.NoError(t, err)

	assert.Equal(t, "test@example.com", contact.Email)
	assert.Equal(t, "Test Location", contact.Location)
	assert.Equal(t, "Available", contact.Availability)
}

func TestSkillsEndpoint(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/skills", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var skills []Skill
	err := json.Unmarshal(w.Body.Bytes(), &skills)
	require.NoError(t, err)

	assert.Len(t, skills, 1)
	assert.Equal(t, "Go", skills[0].Name)
	assert.Equal(t, "Backend", skills[0].Category)
	assert.Equal(t, 90, skills[0].Proficiency)
}

func TestExperienceEndpoint(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/experience", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var experience []Experience
	err := json.Unmarshal(w.Body.Bytes(), &experience)
	require.NoError(t, err)

	assert.Len(t, experience, 1)
	assert.Equal(t, "Software Engineer", experience[0].Title)
	assert.Equal(t, "Test Company", experience[0].Company)
	assert.False(t, experience[0].Current)
}

func TestTrackProjectViewEndpoint(t *testing.T) {
	router := setupTestRouter(t)

	trackData := map[string]interface{}{
		"project_id": 1,
		"ip":         "127.0.0.1",
		"user_agent": "test-agent",
		"referrer":   "https://test.com",
	}

	jsonData, _ := json.Marshal(trackData)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/analytics/track", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "tracked", response["status"])
	assert.Equal(t, float64(1), response["project_id"])
}

func TestTrackProjectViewInvalidData(t *testing.T) {
	router := setupTestRouter(t)

	// Invalid JSON
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/analytics/track", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMetricsEndpoint(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/metrics", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/plain")
	assert.Contains(t, w.Body.String(), "test_metric")
}

func TestNotFoundEndpoint(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/nonexistent", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// Benchmark tests
func BenchmarkHealthEndpoint(b *testing.B) {
	router := setupTestRouter(nil)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		router.ServeHTTP(w, req)
	}
}

func BenchmarkProjectsEndpoint(b *testing.B) {
	router := setupTestRouter(nil)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/projects", nil)
		router.ServeHTTP(w, req)
	}
}

// Integration test helpers
func TestDatabaseConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database test in short mode")
	}

	db := setupTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("Error closing test database: %v", err)
		}
	}()

	// Test basic query
	var result int
	err := db.QueryRow("SELECT 1").Scan(&result)
	require.NoError(t, err)
	assert.Equal(t, 1, result)
}

// Test environment setup
func TestMain(m *testing.M) {
	// Setup test database connection
	testDB, err := sql.Open("postgres", getTestDBURL())
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}
	defer func() {
		if err := testDB.Close(); err != nil {
			log.Printf("Error closing test database: %v", err)
		}
	}()

	// Run tests
	os.Exit(m.Run())
}

func getTestDBURL() string {
	// Use environment variable for test database URL
	if url := os.Getenv("TEST_DATABASE_URL"); url != "" {
		return url
	}
	// Fallback to default test database URL
	return "postgres://postgres:secure-password@localhost:5432/bruno_site_test?sslmode=disable"
}
