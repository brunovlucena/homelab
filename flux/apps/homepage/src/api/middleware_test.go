package main

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-contrib/cors"
	gzipMiddleware "github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Test setup for middleware tests
func setupMiddlewareTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add middleware
	router.Use(requestLogger())
	router.Use(errorHandler())

	// Add test routes
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	router.GET("/error", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "test error"})
	})

	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	return router
}

// =============================================================================
// 📝 REQUEST LOGGER TESTS
// =============================================================================

func TestRequestLogger(t *testing.T) {
	router := setupMiddlewareTestRouter()

	t.Run("Successful Request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("User-Agent", "test-agent")
		req.Header.Set("X-Forwarded-For", "192.168.1.1")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "test", response["message"])
	})

	t.Run("Request with Custom Headers", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("POST Request with Body", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/test", strings.NewReader(`{"test": "data"}`))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		// Router returns 404 for unregistered routes, 405 would require explicit method handling
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// =============================================================================
// ❌ ERROR HANDLER TESTS
// =============================================================================

func TestErrorHandler(t *testing.T) {
	router := setupMiddlewareTestRouter()

	t.Run("Panic Recovery", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/panic", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "error")
	})

	t.Run("Normal Error Response", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/error", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "test error", response["error"])
	})
}

// =============================================================================
// 🔒 CORS MIDDLEWARE TESTS
// =============================================================================

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	t.Run("Preflight Request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "GET")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type")

		router.ServeHTTP(w, req)

		// CORS preflight typically returns 204 No Content
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNoContent, "Expected 200 or 204, got %d", w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
	})

	t.Run("Actual Request with CORS Headers", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	})
}

// =============================================================================
// 🗜️ GZIP MIDDLEWARE TESTS
// =============================================================================

func TestGzipMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add gzip middleware
	router.Use(gzipMiddleware.Gzip(gzipMiddleware.DefaultCompression))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "This is a test message that should be compressed",
			"data":    strings.Repeat("test data ", 100),
		})
	})

	t.Run("Request with Accept-Encoding: gzip", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))
	})

	t.Run("Request without Accept-Encoding", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "", w.Header().Get("Content-Encoding"))
	})
}

// =============================================================================
// 🏃 BENCHMARK TESTS
// =============================================================================

func BenchmarkRequestLogger(b *testing.B) {
	router := setupMiddlewareTestRouter()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
	}
}

func BenchmarkErrorHandler(b *testing.B) {
	router := setupMiddlewareTestRouter()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/error", nil)
		router.ServeHTTP(w, req)
	}
}

func BenchmarkCORSMiddleware(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		router.ServeHTTP(w, req)
	}
}

func BenchmarkGzipMiddleware(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(gzipMiddleware.Gzip(gzipMiddleware.DefaultCompression))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "This is a test message that should be compressed",
			"data":    strings.Repeat("test data ", 100),
		})
	})

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		router.ServeHTTP(w, req)
	}
}

// =============================================================================
// 🔧 INTEGRATION TESTS
// =============================================================================

func TestMiddlewareIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add all middleware in the same order as main application
	router.Use(requestLogger())
	router.Use(errorHandler())
	router.Use(gzipMiddleware.Gzip(gzipMiddleware.DefaultCompression))
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	t.Run("All Middleware Working Together", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("User-Agent", "test-agent")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Check CORS headers
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))

		// Check gzip headers
		assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))

		// Decompress gzipped response
		var bodyReader io.Reader = w.Body
		if w.Header().Get("Content-Encoding") == "gzip" {
			gzReader, err := gzip.NewReader(w.Body)
			assert.NoError(t, err)
			defer func(gzReader *gzip.Reader) {
				if err := gzReader.Close(); err != nil {
					t.Logf("Error closing gzip reader: %v", err)
				}
			}(gzReader)
			bodyReader = gzReader
		}

		bodyBytes, err := io.ReadAll(bodyReader)
		assert.NoError(t, err)

		var response map[string]interface{}
		err = json.Unmarshal(bodyBytes, &response)
		assert.NoError(t, err)
		assert.Equal(t, "test", response["message"])
	})
}

// =============================================================================
// 🧪 EDGE CASE TESTS
// =============================================================================

func TestMiddlewareEdgeCases(t *testing.T) {
	router := setupMiddlewareTestRouter()

	t.Run("Very Long User Agent", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("User-Agent", strings.Repeat("very-long-user-agent-", 100))

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Empty Headers", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Special Characters in Headers", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("User-Agent", "test-agent-🚀-emoji")
		req.Header.Set("X-Custom-Header", "special-chars: !@#$%^&*()")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
