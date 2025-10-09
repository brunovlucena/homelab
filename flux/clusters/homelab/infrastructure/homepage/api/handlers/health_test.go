package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestHealthCheck tests the HealthCheck handler
func TestHealthCheck(t *testing.T) {
	tests := []struct {
		name             string
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "success - returns healthy status",
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "healthy", response["status"])
				assert.Equal(t, "bruno-site-api", response["service"])
				assert.Equal(t, "1.0.0", response["version"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)

			HealthCheck(c)

			assert.Equal(t, http.StatusOK, w.Code)
			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}

// TestHealthCheckFormat tests the response format
func TestHealthCheckFormat(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)

	HealthCheck(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify all required fields exist
	assert.Contains(t, response, "status")
	assert.Contains(t, response, "service")
	assert.Contains(t, response, "version")
}

// TestHealthCheckIdempotency tests that health check is idempotent
func TestHealthCheckIdempotency(t *testing.T) {
	// Call health check multiple times
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)

		HealthCheck(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "healthy", response["status"])
	}
}
