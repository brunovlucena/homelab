package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brunovlucena/homelab/homepage-api/cdn"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestNewCloudflareHandler tests CloudflareHandler initialization
func TestNewCloudflareHandler(t *testing.T) {
	cloudflareCDN := cdn.NewCloudflareCDN("test-zone", "test-token", "example.com", true, 3600)
	handler := NewCloudflareHandler(cloudflareCDN)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.cdn)
}

// TestGetCDNStatus tests the GetCDNStatus handler
func TestGetCDNStatus(t *testing.T) {
	tests := []struct {
		name             string
		cdnEnabled       bool
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:       "success - CDN disabled",
			cdnEnabled: false,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, false, response["enabled"])
				assert.Equal(t, "Cloudflare CDN is disabled", response["message"])
			},
		},
		{
			name:       "success - CDN enabled without credentials",
			cdnEnabled: true,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, true, response["enabled"])
				assert.Equal(t, false, response["healthy"])
				assert.Contains(t, response["message"], "unhealthy")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cloudflareCDN := cdn.NewCloudflareCDN("", "", "example.com", tt.cdnEnabled, 3600)
			handler := NewCloudflareHandler(cloudflareCDN)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/cloudflare/status", nil)

			handler.GetCDNStatus(c)

			assert.Equal(t, http.StatusOK, w.Code)
			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}

// TestGetAssetURL tests the GetAssetURL handler
func TestGetAssetURL(t *testing.T) {
	tests := []struct {
		name               string
		queryParams        map[string]string
		expectedStatusCode int
		validateResponse   func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "success - returns asset URL without optimization",
			queryParams: map[string]string{
				"path": "images/logo.png",
			},
			expectedStatusCode: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, true, response["success"])
				assert.Contains(t, response["url"], "logo.png")
				assert.Equal(t, "images/logo.png", response["path"])
			},
		},
		{
			name: "success - returns optimized image URL",
			queryParams: map[string]string{
				"path":   "images/photo.jpg",
				"width":  "300",
				"height": "200",
				"format": "webp",
			},
			expectedStatusCode: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, true, response["success"])
				assert.Contains(t, response["url"], "width=300")
				assert.Contains(t, response["url"], "height=200")
				assert.Contains(t, response["url"], "format=webp")
			},
		},
		{
			name:               "error - missing path parameter",
			queryParams:        map[string]string{},
			expectedStatusCode: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, false, response["success"])
				assert.Equal(t, "Asset path is required", response["message"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cloudflareCDN := cdn.NewCloudflareCDN("test-zone", "test-token", "example.com", true, 3600)
			handler := NewCloudflareHandler(cloudflareCDN)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Build query string
			url := "/api/v1/cloudflare/asset?"
			for key, value := range tt.queryParams {
				url += key + "=" + value + "&"
			}

			c.Request = httptest.NewRequest(http.MethodGet, url, nil)

			// Set query params
			for key, value := range tt.queryParams {
				c.Request.URL.RawQuery += key + "=" + value + "&"
			}

			handler.GetAssetURL(c)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}

// TestGetCacheHeaders tests the GetCacheHeaders handler
func TestGetCacheHeaders(t *testing.T) {
	tests := []struct {
		name             string
		cdnEnabled       bool
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:       "success - returns cache headers with CDN enabled",
			cdnEnabled: true,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, true, response["success"])
				assert.NotNil(t, response["headers"])

				headers := response["headers"].(map[string]interface{})
				assert.Contains(t, headers, "Cache-Control")
			},
		},
		{
			name:       "success - returns cache headers with CDN disabled",
			cdnEnabled: false,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, true, response["success"])
				assert.NotNil(t, response["headers"])

				headers := response["headers"].(map[string]interface{})
				assert.Contains(t, headers, "Cache-Control")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cloudflareCDN := cdn.NewCloudflareCDN("test-zone", "test-token", "example.com", tt.cdnEnabled, 7200)
			handler := NewCloudflareHandler(cloudflareCDN)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/cloudflare/headers", nil)

			handler.GetCacheHeaders(c)

			assert.Equal(t, http.StatusOK, w.Code)
			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}

// TestPurgeCache tests the PurgeCache handler
func TestPurgeCache(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        interface{}
		cdnEnabled         bool
		expectedStatusCode int
		validateResponse   func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "error - CDN disabled",
			requestBody: PurgeCacheRequest{
				All: true,
			},
			cdnEnabled:         false,
			expectedStatusCode: http.StatusServiceUnavailable,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, false, response["success"])
				assert.Equal(t, "Cloudflare CDN is not enabled", response["message"])
			},
		},
		{
			name:               "error - invalid request format",
			requestBody:        "invalid json",
			cdnEnabled:         true,
			expectedStatusCode: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, false, response["success"])
				assert.Equal(t, "Invalid request format", response["message"])
			},
		},
		{
			name:               "error - no purge target specified",
			requestBody:        PurgeCacheRequest{},
			cdnEnabled:         true,
			expectedStatusCode: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, false, response["success"])
				assert.Equal(t, "No purge target specified", response["message"])
			},
		},
		{
			name: "error - purge all with no credentials (should fail)",
			requestBody: PurgeCacheRequest{
				All: true,
			},
			cdnEnabled:         true,
			expectedStatusCode: http.StatusInternalServerError,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, false, response["success"])
				assert.Contains(t, response["message"], "Failed to purge cache")
			},
		},
		{
			name: "error - purge specific files with no credentials",
			requestBody: PurgeCacheRequest{
				Files: []string{"https://example.com/image1.jpg", "https://example.com/image2.jpg"},
			},
			cdnEnabled:         true,
			expectedStatusCode: http.StatusInternalServerError,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, false, response["success"])
				assert.Contains(t, response["message"], "Failed to purge cache")
			},
		},
		{
			name: "error - purge by tags with no credentials",
			requestBody: PurgeCacheRequest{
				Tags: []string{"images", "css"},
			},
			cdnEnabled:         true,
			expectedStatusCode: http.StatusInternalServerError,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, false, response["success"])
				assert.Contains(t, response["message"], "Failed to purge cache")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create CDN without credentials so purge will fail (for testing error cases)
			cloudflareCDN := cdn.NewCloudflareCDN("", "", "example.com", tt.cdnEnabled, 3600)
			handler := NewCloudflareHandler(cloudflareCDN)

			bodyBytes, _ := json.Marshal(tt.requestBody)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/cloudflare/purge", bytes.NewBuffer(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.PurgeCache(c)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}

// TestPurgeCacheRequestJSON tests JSON marshaling for PurgeCacheRequest
func TestPurgeCacheRequestJSON(t *testing.T) {
	request := PurgeCacheRequest{
		All:   true,
		Files: []string{"https://example.com/file1.jpg"},
		Tags:  []string{"images"},
	}

	jsonData, err := json.Marshal(request)
	assert.NoError(t, err)

	var decoded PurgeCacheRequest
	err = json.Unmarshal(jsonData, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, request.All, decoded.All)
	assert.Equal(t, request.Files, decoded.Files)
	assert.Equal(t, request.Tags, decoded.Tags)
}

// TestPurgeCacheResponseJSON tests JSON marshaling for PurgeCacheResponse
func TestPurgeCacheResponseJSON(t *testing.T) {
	response := PurgeCacheResponse{
		Success: true,
		Message: "Cache purged successfully",
		Files:   5,
	}

	jsonData, err := json.Marshal(response)
	assert.NoError(t, err)

	var decoded PurgeCacheResponse
	err = json.Unmarshal(jsonData, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, response.Success, decoded.Success)
	assert.Equal(t, response.Message, decoded.Message)
	assert.Equal(t, response.Files, decoded.Files)
}
