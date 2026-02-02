package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test setup helpers
func setupTestRouterForHandlers() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Setup routes
	api := router.Group("/api/v1")
	{
		api.GET("/projects", getProjects)
		api.GET("/projects/:id", getProject)
		api.POST("/projects", createProject)
		api.PUT("/projects/:id", updateProject)
		api.DELETE("/projects/:id", deleteProject)

		api.GET("/skills", getSkills)
		api.GET("/skills/:id", getSkill)
		api.POST("/skills", createSkill)
		api.PUT("/skills/:id", updateSkill)
		api.DELETE("/skills/:id", deleteSkill)

		api.GET("/experiences", getExperiences)
		api.GET("/experiences/:id", getExperience)
		api.POST("/experiences", createExperience)
		api.PUT("/experiences/:id", updateExperience)
		api.DELETE("/experiences/:id", deleteExperience)

		api.GET("/content", getContent)
		api.GET("/content/:type", getContentByType)
		api.POST("/content", createContent)
		api.PUT("/content/:id", updateContent)
		api.DELETE("/content/:id", deleteContent)

		api.GET("/about", getAbout)
		api.PUT("/about", updateAbout)

		api.GET("/contact", getContact)
		api.PUT("/contact", updateContact)

		api.POST("/analytics/track", handleAnalyticsTrack)
	}

	return router
}

// =============================================================================
// 🎯 PROJECT TESTS
// =============================================================================

func TestGetProjects(t *testing.T) {
	router := setupTestRouterForHandlers()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/projects", nil)
	router.ServeHTTP(w, req)

	// Since we don't have a real DB in unit tests, this will return an error
	// In a real scenario, you'd mock the database
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Failed to fetch projects", response["error"])
}

func TestGetProject(t *testing.T) {
	router := setupTestRouterForHandlers()

	t.Run("Valid ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/projects/1", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Invalid ID - Empty", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/projects/", nil)
		router.ServeHTTP(w, req)

		// Gin redirects trailing slashes, so we accept either 301 or 404
		assert.True(t, w.Code == http.StatusMovedPermanently || w.Code == http.StatusNotFound, "Expected 301 or 404, got %d", w.Code)
	})

	t.Run("Invalid ID - Non-numeric", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/projects/abc", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Invalid project ID format", response["error"])
	})

	t.Run("Invalid ID - Zero", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/projects/0", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Invalid project ID format", response["error"])
	})
}

func TestCreateProject(t *testing.T) {
	router := setupTestRouterForHandlers()

	t.Run("Valid Project", func(t *testing.T) {
		project := Project{
			Title:        "Test Project",
			Description:  "A test project description",
			Type:         "web",
			Technologies: []string{"Go", "React"},
			Active:       true,
		}

		jsonData, _ := json.Marshal(project)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/projects", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/projects", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Invalid request body", response["error"])
	})

	t.Run("Empty Title", func(t *testing.T) {
		project := Project{
			Title:       "",
			Description: "A test project description",
			Type:        "web",
		}

		jsonData, _ := json.Marshal(project)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/projects", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Invalid title length", response["error"])
	})

	t.Run("Title Too Long", func(t *testing.T) {
		longTitle := make([]byte, 256)
		for i := range longTitle {
			longTitle[i] = 'a'
		}

		project := Project{
			Title:       string(longTitle),
			Description: "A test project description",
			Type:        "web",
		}

		jsonData, _ := json.Marshal(project)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/projects", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Invalid title length", response["error"])
	})

	t.Run("Empty Description", func(t *testing.T) {
		project := Project{
			Title:       "Test Project",
			Description: "",
			Type:        "web",
		}

		jsonData, _ := json.Marshal(project)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/projects", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Invalid description length", response["error"])
	})

	t.Run("Description Too Long", func(t *testing.T) {
		longDescription := make([]byte, 5001)
		for i := range longDescription {
			longDescription[i] = 'a'
		}

		project := Project{
			Title:       "Test Project",
			Description: string(longDescription),
			Type:        "web",
		}

		jsonData, _ := json.Marshal(project)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/projects", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Invalid description length", response["error"])
	})

	t.Run("Empty Type", func(t *testing.T) {
		project := Project{
			Title:       "Test Project",
			Description: "A test project description",
			Type:        "",
		}

		jsonData, _ := json.Marshal(project)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/projects", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Invalid type length", response["error"])
	})

	t.Run("Type Too Long", func(t *testing.T) {
		longType := make([]byte, 101)
		for i := range longType {
			longType[i] = 'a'
		}

		project := Project{
			Title:       "Test Project",
			Description: "A test project description",
			Type:        string(longType),
		}

		jsonData, _ := json.Marshal(project)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/projects", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Invalid type length", response["error"])
	})
}

func TestUpdateProject(t *testing.T) {
	router := setupTestRouterForHandlers()

	t.Run("Valid Update", func(t *testing.T) {
		project := Project{
			Title:        "Updated Project",
			Description:  "An updated test project description",
			Type:         "web",
			Technologies: []string{"Go", "React", "TypeScript"},
			Active:       true,
		}

		jsonData, _ := json.Marshal(project)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/v1/projects/1", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		project := Project{
			Title:       "Updated Project",
			Description: "An updated test project description",
			Type:        "web",
		}

		jsonData, _ := json.Marshal(project)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/v1/projects/abc", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Invalid project ID format", response["error"])
	})
}

func TestDeleteProject(t *testing.T) {
	router := setupTestRouterForHandlers()

	t.Run("Valid ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/api/v1/projects/1", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/api/v1/projects/abc", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Invalid project ID format", response["error"])
	})
}

// =============================================================================
// 🛠️ SKILL TESTS
// =============================================================================

func TestGetSkills(t *testing.T) {
	router := setupTestRouterForHandlers()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/skills", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreateSkill(t *testing.T) {
	router := setupTestRouterForHandlers()

	t.Run("Valid Skill", func(t *testing.T) {
		skill := Skill{
			Name:        "Go",
			Category:    "Backend",
			Proficiency: 5,
			Icon:        "go-icon",
			Order:       1,
		}

		jsonData, _ := json.Marshal(skill)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/skills", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Empty Name", func(t *testing.T) {
		skill := Skill{
			Name:        "",
			Category:    "Backend",
			Proficiency: 5,
		}

		jsonData, _ := json.Marshal(skill)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/skills", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Invalid skill name length", response["error"])
	})

	t.Run("Name Too Long", func(t *testing.T) {
		longName := make([]byte, 101)
		for i := range longName {
			longName[i] = 'a'
		}

		skill := Skill{
			Name:        string(longName),
			Category:    "Backend",
			Proficiency: 5,
		}

		jsonData, _ := json.Marshal(skill)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/skills", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Invalid skill name length", response["error"])
	})

	t.Run("Empty Category", func(t *testing.T) {
		skill := Skill{
			Name:        "Go",
			Category:    "",
			Proficiency: 5,
		}

		jsonData, _ := json.Marshal(skill)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/skills", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Invalid category length", response["error"])
	})

	t.Run("Invalid Proficiency - Too Low", func(t *testing.T) {
		skill := Skill{
			Name:        "Go",
			Category:    "Backend",
			Proficiency: 0,
		}

		jsonData, _ := json.Marshal(skill)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/skills", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Invalid proficiency level (1-5)", response["error"])
	})

	t.Run("Invalid Proficiency - Too High", func(t *testing.T) {
		skill := Skill{
			Name:        "Go",
			Category:    "Backend",
			Proficiency: 6,
		}

		jsonData, _ := json.Marshal(skill)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/skills", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Invalid proficiency level (1-5)", response["error"])
	})

	t.Run("Icon Too Long", func(t *testing.T) {
		longIcon := make([]byte, 51)
		for i := range longIcon {
			longIcon[i] = 'a'
		}

		skill := Skill{
			Name:        "Go",
			Category:    "Backend",
			Proficiency: 5,
			Icon:        string(longIcon),
		}

		jsonData, _ := json.Marshal(skill)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/skills", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Invalid icon length", response["error"])
	})
}

// =============================================================================
// 💼 EXPERIENCE TESTS
// =============================================================================

func TestGetExperiences(t *testing.T) {
	router := setupTestRouterForHandlers()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/experiences", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreateExperience(t *testing.T) {
	router := setupTestRouterForHandlers()

	t.Run("Valid Experience", func(t *testing.T) {
		endDate := "2023-12"
		experience := Experience{
			Title:        "Software Engineer",
			Company:      "Test Company",
			StartDate:    "2020-01",
			EndDate:      &endDate,
			Current:      false,
			Description:  "Worked on various projects",
			Technologies: []string{"Go", "React", "PostgreSQL"},
			Order:        1,
			Active:       true,
		}

		jsonData, _ := json.Marshal(experience)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/experiences", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/experiences", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Invalid request body", response["error"])
	})
}

// =============================================================================
// 📄 CONTENT TESTS
// =============================================================================

func TestGetContent(t *testing.T) {
	router := setupTestRouterForHandlers()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/content", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetContentByType(t *testing.T) {
	router := setupTestRouterForHandlers()

	t.Run("Valid Type", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/content/about", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestCreateContent(t *testing.T) {
	router := setupTestRouterForHandlers()

	t.Run("Valid Content", func(t *testing.T) {
		content := Content{
			Type:  "about",
			Value: "This is about content",
		}

		jsonData, _ := json.Marshal(content)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/content", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/content", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Invalid request body", response["error"])
	})
}

// =============================================================================
// 👤 ABOUT TESTS
// =============================================================================

func TestGetAbout(t *testing.T) {
	router := setupTestRouterForHandlers()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/about", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateAbout(t *testing.T) {
	router := setupTestRouterForHandlers()

	t.Run("Valid About Data", func(t *testing.T) {
		aboutData := AboutData{
			Description: "Updated about description",
		}

		jsonData, _ := json.Marshal(aboutData)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/v1/about", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/v1/about", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Invalid request body", response["error"])
	})
}

// =============================================================================
// 📞 CONTACT TESTS
// =============================================================================

func TestGetContact(t *testing.T) {
	router := setupTestRouterForHandlers()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/contact", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateContact(t *testing.T) {
	router := setupTestRouterForHandlers()

	t.Run("Valid Contact Data", func(t *testing.T) {
		contactData := ContactData{
			Email:        "test@example.com",
			Location:     "Test Location",
			LinkedIn:     "https://linkedin.com/test",
			GitHub:       "https://github.com/test",
			Availability: "Available",
		}

		jsonData, _ := json.Marshal(contactData)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/v1/contact", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/v1/contact", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Invalid request body", response["error"])
	})
}

// =============================================================================
// 📊 ANALYTICS TESTS
// =============================================================================

func TestHandleAnalyticsTrack(t *testing.T) {
	router := setupTestRouterForHandlers()

	t.Run("Valid Track Data", func(t *testing.T) {
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
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/analytics/track", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// =============================================================================
// 🛠️ UTILITY FUNCTION TESTS
// =============================================================================

func TestIsValidURL(t *testing.T) {
	t.Run("Valid HTTP URL", func(t *testing.T) {
		assert.True(t, isValidURL("http://example.com"))
	})

	t.Run("Valid HTTPS URL", func(t *testing.T) {
		assert.True(t, isValidURL("https://example.com"))
	})

	t.Run("Valid URL with Path", func(t *testing.T) {
		assert.True(t, isValidURL("https://example.com/path/to/resource"))
	})

	t.Run("Valid URL with Query Parameters", func(t *testing.T) {
		assert.True(t, isValidURL("https://example.com?param=value"))
	})

	t.Run("Invalid URL - Empty String", func(t *testing.T) {
		assert.False(t, isValidURL(""))
	})

	t.Run("Invalid URL - No Scheme", func(t *testing.T) {
		assert.False(t, isValidURL("example.com"))
	})

	t.Run("Invalid URL - Invalid Scheme", func(t *testing.T) {
		assert.False(t, isValidURL("ftp://example.com"))
	})

	t.Run("Invalid URL - No Host", func(t *testing.T) {
		assert.False(t, isValidURL("http://"))
	})

	t.Run("Invalid URL - Malformed", func(t *testing.T) {
		assert.False(t, isValidURL("not-a-url"))
	})
}

// =============================================================================
// 🏃 BENCHMARK TESTS
// =============================================================================

func BenchmarkGetProjects(b *testing.B) {
	router := setupTestRouterForHandlers()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/projects", nil)
		router.ServeHTTP(w, req)
	}
}

func BenchmarkGetProject(b *testing.B) {
	router := setupTestRouterForHandlers()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/projects/1", nil)
		router.ServeHTTP(w, req)
	}
}

func BenchmarkCreateProject(b *testing.B) {
	router := setupTestRouterForHandlers()
	project := Project{
		Title:        "Benchmark Project",
		Description:  "A benchmark test project",
		Type:         "web",
		Technologies: []string{"Go", "React"},
		Active:       true,
	}
	jsonData, _ := json.Marshal(project)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/projects", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
	}
}

func BenchmarkIsValidURL(b *testing.B) {
	url := "https://example.com/path/to/resource?param=value"

	for i := 0; i < b.N; i++ {
		isValidURL(url)
	}
}
