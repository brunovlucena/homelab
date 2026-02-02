package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	// 🌐 Web framework and middleware
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"

	// 🔧 Environment and database
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	// 📊 Prometheus monitoring
	"github.com/prometheus/client_golang/prometheus/promhttp"

	// 🗄️ Redis caching
	"github.com/redis/go-redis/v9"

	// 🤖 LLM services

	"bruno-api/services"
)

// =============================================================================
// 📋 GLOBAL VARIABLES
// =============================================================================

var (
	db          *sql.DB
	redisClient *redis.Client
	llmService  *services.LLMService
)

// =============================================================================
// 🚀 MAIN APPLICATION
// =============================================================================

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize database connection
	if err := initDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Initialize Redis connection
	if err := initRedis(); err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("Error closing Redis: %v", err)
		}
	}()

	// Initialize LLM service
	initLLMService()

	// Initialize metrics
	initMetrics()

	// Initialize OpenTelemetry (if enabled)
	initTracing()

	// Setup Gin router
	router := setupRouter()

	// Get port from environment
	port := getEnv("PORT", "8080")

	// Start server
	log.Printf("🚀 Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// =============================================================================
// 🔧 INITIALIZATION FUNCTIONS
// =============================================================================

func initDatabase() error {
	// Construct connection string from individual environment variables
	host := getEnv("DATABASE_HOST", "localhost")
	port := getEnv("DATABASE_PORT", "5432")
	user := getEnv("DATABASE_USER", "postgres")
	password := getEnv("PGPASSWORD", "secure-password")
	dbname := getEnv("DATABASE_NAME", "bruno_site")

	// URL-encode the password to handle special characters
	encodedPassword := url.QueryEscape(password)
	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", user, encodedPassword, host, port, dbname)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Retry connection with exponential backoff
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		if err := db.Ping(); err != nil {
			log.Printf("⏳ Database connection attempt %d/%d failed: %v", i+1, maxRetries, err)
			if i < maxRetries-1 {
				time.Sleep(time.Duration(i+1) * 2 * time.Second)
			}
		} else {
			log.Println("✅ Database connected successfully")
			return nil
		}
	}

	return fmt.Errorf("failed to connect to database after %d attempts", maxRetries)
}

func initRedis() error {
	redisURL := getEnv("REDIS_URL", "redis://localhost:6379")

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return err
	}

	redisClient = redis.NewClient(opts)

	// Retry connection with exponential backoff
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := redisClient.Ping(ctx).Err(); err != nil {
			cancel()
			log.Printf("⏳ Redis connection attempt %d/%d failed: %v", i+1, maxRetries, err)
			if i < maxRetries-1 {
				time.Sleep(time.Duration(i+1) * 2 * time.Second)
			}
		} else {
			cancel()
			log.Println("✅ Redis connected successfully")
			return nil
		}
	}

	return fmt.Errorf("failed to connect to Redis after %d attempts", maxRetries)
}

func initLLMService() {
	llmService = services.NewLLMService(db)

	// Test LLM service health
	if err := llmService.HealthCheck(); err != nil {
		log.Printf("⚠️ LLM service health check failed: %v", err)
		log.Println("💡 Make sure Ollama is running and the model is available")
	} else {
		log.Println("🤖 LLM service initialized and healthy")
	}
}

func initMetrics() {
	// Set application version metrics
	SetApplicationVersion(
		getEnv("APP_VERSION", "1.0.0"),
		getEnv("BUILD_DATE", "unknown"),
		getEnv("GIT_COMMIT", "unknown"),
	)

	// Start uptime tracking
	go func() {
		startTime := time.Now()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			UpdateUptime(time.Since(startTime))
		}
	}()

	log.Println("📊 Prometheus metrics initialized")
}

func initTracing() {
	// OpenTelemetry initialization (currently disabled)
	// This can be enabled when needed for distributed tracing
	log.Println("ℹ️  OpenTelemetry tracing disabled")
}

// =============================================================================
// 🌐 ROUTER SETUP
// =============================================================================

func setupRouter() *gin.Engine {
	// Set Gin mode
	if getEnv("GIN_MODE", "release") == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Add middleware in order
	router.Use(requestIDMiddleware())       // Request ID first for tracing
	router.Use(prometheusMiddleware())      // Prometheus metrics
	router.Use(requestLogger())             // Logging with request ID
	router.Use(errorHandler())              // Error recovery
	router.Use(securityHeadersMiddleware()) // Security headers

	// Gzip compression - exclude metrics endpoint for Prometheus compatibility
	router.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedPaths([]string{"/metrics"})))

	// CORS configuration - restrict to allowed origins
	allowedOrigins := strings.Split(getEnv("CORS_ORIGINS", "https://lucena.cloud,https://www.lucena.cloud,http://localhost:5173,http://localhost:3000"), ",")
	router.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID", "X-RateLimit-Remaining"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check endpoint (no rate limiting)
	router.GET("/health", healthCheck)

	// Prometheus metrics endpoint (no authentication or rate limiting required)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Register API routes for both /api and /api/v1 prefixes
	for _, prefix := range []string{"/api", "/api/v1"} {
		registerAPIRoutes(router.Group(prefix))
	}

	return router
}

// registerAPIRoutes registers all API routes to a router group
// This eliminates duplication between /api and /api/v1 routes
func registerAPIRoutes(api *gin.RouterGroup) {
	// Apply rate limiting to all API routes
	api.Use(rateLimitMiddleware(apiRateLimiter))

	// Projects
	api.GET("/projects", getProjects)
	api.GET("/projects/:id", getProject)
	api.POST("/projects", createProject)
	api.PUT("/projects/:id", updateProject)
	api.DELETE("/projects/:id", deleteProject)

	// Skills
	api.GET("/skills", getSkills)
	api.GET("/skills/:id", getSkill)
	api.POST("/skills", createSkill)
	api.PUT("/skills/:id", updateSkill)
	api.DELETE("/skills/:id", deleteSkill)

	// Experiences
	api.GET("/experiences", getExperiences)
	api.GET("/experiences/:id", getExperience)
	api.POST("/experiences", createExperience)
	api.PUT("/experiences/:id", updateExperience)
	api.DELETE("/experiences/:id", deleteExperience)

	// Content
	api.GET("/content", getContent)
	api.GET("/content/:type", getContentByType)
	api.POST("/content", createContent)
	api.PUT("/content/:id", updateContent)
	api.DELETE("/content/:id", deleteContent)

	// Site Config (dynamic titles, subtitles, etc)
	api.GET("/config", getSiteConfig)
	api.PUT("/config", updateSiteConfig)

	// About
	api.GET("/about", getAbout)
	api.PUT("/about", updateAbout)

	// Contact
	api.GET("/contact", getContact)
	api.PUT("/contact", updateContact)

	// 🤖 AI Chat endpoint (with stricter rate limiting)
	chatGroup := api.Group("/chat")
	chatGroup.Use(rateLimitMiddleware(chatRateLimiter))
	chatGroup.POST("", handleChat)
	chatGroup.GET("/health", handleChatHealth)

	// 📊 Analytics endpoint
	api.POST("/analytics/track", handleAnalyticsTrack)
}

// =============================================================================
// 🛠️ UTILITY FUNCTIONS
// =============================================================================

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Helper function to truncate strings for logging
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "bruno-api",
	})
}

// =============================================================================
// 🤖 CHAT HANDLERS
// =============================================================================

func handleChat(c *gin.Context) {
	startTime := time.Now()
	requestID := fmt.Sprintf("chat_handler_%d", startTime.UnixNano())

	log.Printf("🤖 [%s] Chat request received", requestID)
	log.Printf("   📍 Remote IP: %s", c.ClientIP())
	log.Printf("   📍 User Agent: %s", c.GetHeader("User-Agent"))
	log.Printf("   📍 Content-Type: %s", c.GetHeader("Content-Type"))
	log.Printf("   📍 Content-Length: %s", c.GetHeader("Content-Length"))

	var request services.ChatRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("❌ [%s] JSON binding failed: %v", requestID, err)
		log.Printf("   📄 Request body: %s", c.Request.Body)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ [%s] JSON binding successful", requestID)
	log.Printf("   📝 Message: %s", truncateString(request.Message, 100))
	log.Printf("   📝 Context: %s", truncateString(request.Context, 50))

	// Validate message is not empty
	if strings.TrimSpace(request.Message) == "" {
		log.Printf("❌ [%s] Empty message received", requestID)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Message cannot be empty",
		})
		return
	}

	log.Printf("🔄 [%s] Processing chat request...", requestID)

	// Process chat request
	response, err := llmService.ProcessChat(request)
	duration := time.Since(startTime)

	if err != nil {
		log.Printf("❌ [%s] Chat processing error: %v", requestID, err)
		log.Printf("   🔍 Error type: %T", err)
		log.Printf("   🔍 Full error details: %+v", err)

		// Record failed chat session metrics
		RecordChatSession(false)
		RecordLLMMetrics("agent-bruno", duration, false, 0)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to process chat request",
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ [%s] Chat request completed successfully in %v", requestID, duration)
	log.Printf("   📤 Response length: %d chars", len(response.Response))
	log.Printf("   🎯 Model used: %s", response.Model)

	// Record successful chat session metrics
	RecordChatSession(true)
	RecordLLMMetrics(response.Model, duration, true, len(response.Response))

	c.JSON(http.StatusOK, response)
}

func handleChatHealth(c *gin.Context) {
	if err := llmService.HealthCheck(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "unhealthy",
			"error":     err.Error(),
			"timestamp": time.Now().UTC(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"provider":  "agent-bruno",
		"model":     "llama3.2:3b",
		"timestamp": time.Now().UTC(),
	})
}
