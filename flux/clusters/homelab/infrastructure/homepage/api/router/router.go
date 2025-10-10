package router

import (
	"bruno-site/cache"
	"bruno-site/cdn"
	"bruno-site/config"
	"bruno-site/handlers"
	"bruno-site/storage"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"gorm.io/gorm"
)

// SetupRouter sets up the Gin router with all routes and middleware
func SetupRouter(cfg *config.Config, db *gorm.DB, redis *redis.Client, minioClient *storage.MinIOClient) *gin.Engine {
	r := gin.Default()

	// Initialize cache manager (for future use)
	_ = cache.NewCacheManager(redis)

	// Initialize Cloudflare CDN
	cloudflareCDN := cdn.NewCloudflareCDN(
		cfg.Cloudflare.ZoneID,
		cfg.Cloudflare.APIToken,
		cfg.Cloudflare.Domain,
		cfg.Cloudflare.Enabled,
		cfg.Cloudflare.CacheTTL,
	)

	// Initialize Cloudflare handler
	cloudflareHandler := handlers.NewCloudflareHandler(cloudflareCDN)

	// Initialize Agent-SRE handler
	agentSREHandler := handlers.NewAgentSREHandler(handlers.AgentSREConfig{
		ServiceURL: cfg.AgentSREURL, // Get from config, fallback to default
	})

	// 🤖 Initialize Jamie handler (AI-powered SRE assistant)
	jamieHandler := handlers.NewJamieHandler(handlers.JamieConfig{
		ServiceURL: cfg.JamieURL, // Get from config, fallback to default
	})

	// 🏥 Register Jamie as a dependency for health checks
	handlers.SetJamieChecker(jamieHandler)

	// 📊 OpenTelemetry middleware for automatic tracing
	r.Use(otelgin.Middleware("bruno-site"))

	// Compression middleware (Golden Rule #6: Payload Compression)
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.CORSOrigin, "*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
	}))

	// Health check
	r.GET("/health", handlers.HealthCheck)

	// API routes
	api := r.Group("/api/v1")
	{
		// Projects routes
		projects := api.Group("/projects")
		{
			projects.GET("", handlers.GetProjects(db))
			projects.GET("/:id", handlers.GetProject(db))
			projects.POST("", handlers.CreateProject(db))
			projects.PUT("/:id", handlers.UpdateProject(db))
			projects.DELETE("/:id", handlers.DeleteProject(db))
		}

		// Skills routes
		skills := api.Group("/skills")
		{
			skills.GET("", handlers.GetSkills(db))
			skills.GET("/:id", handlers.GetSkill(db))
			skills.POST("", handlers.CreateSkill(db))
			skills.PUT("/:id", handlers.UpdateSkill(db))
			skills.DELETE("/:id", handlers.DeleteSkill(db))
		}

		// Experiences routes
		experiences := api.Group("/experiences")
		{
			experiences.GET("", handlers.GetExperiences(db))
			experiences.GET("/:id", handlers.GetExperience(db))
			experiences.POST("", handlers.CreateExperience(db))
			experiences.PUT("/:id", handlers.UpdateExperience(db))
			experiences.DELETE("/:id", handlers.DeleteExperience(db))
		}

		// Content routes
		content := api.Group("/content")
		{
			content.GET("", handlers.GetContent(db))
			content.GET("/:type", handlers.GetContentByKey(db))
			content.POST("", handlers.CreateContent(db))
			content.PUT("/:id", handlers.UpdateContent(db))
			content.DELETE("/:id", handlers.DeleteContent(db))
		}

		// About routes
		api.GET("/about", handlers.GetAbout(db))
		api.PUT("/about", handlers.UpdateAbout(db))

		// Contact routes
		api.GET("/contact", handlers.GetContact(db))
		api.PUT("/contact", handlers.UpdateContact(db))

		// Assets proxy route - proxy images from MinIO
		if minioClient != nil {
			api.GET("/assets/*path", handlers.ProxyAsset(minioClient))
		}

		// Cloudflare CDN routes
		cloudflare := api.Group("/cloudflare")
		{
			cloudflare.GET("/status", cloudflareHandler.GetCDNStatus)
			cloudflare.GET("/asset", cloudflareHandler.GetAssetURL)
			cloudflare.GET("/headers", cloudflareHandler.GetCacheHeaders)
			cloudflare.POST("/purge", cloudflareHandler.PurgeCache)
		}

		// Agent-SRE proxy routes
		agentSRE := api.Group("/agent-sre")
		{
			// Health and status endpoints
			agentSRE.GET("/health", agentSREHandler.Health)
			agentSRE.GET("/ready", agentSREHandler.Ready)
			agentSRE.GET("/status", agentSREHandler.Status)

			// Chat endpoints
			agentSRE.POST("/chat", agentSREHandler.Chat)
			agentSRE.POST("/mcp/chat", agentSREHandler.MCPChat)

			// Log analysis endpoints
			agentSRE.POST("/analyze-logs", agentSREHandler.AnalyzeLogs)
			agentSRE.POST("/mcp/analyze-logs", agentSREHandler.MCPAnalyzeLogs)
		}

		// 🤖 Jamie (AI-powered SRE assistant) proxy routes
		jamie := api.Group("/jamie")
		{
			// Health and status endpoints
			jamie.GET("/health", jamieHandler.Health)
			jamie.GET("/ready", jamieHandler.Ready)

			// 💬 Main chatbot endpoint for Homepage
			jamie.POST("/chat", jamieHandler.Chat)

			// 📊 SRE operations via Jamie
			jamie.POST("/golden-signals", jamieHandler.CheckGoldenSignals)
			jamie.POST("/prometheus/query", jamieHandler.QueryPrometheus)
			jamie.POST("/pod-logs", jamieHandler.GetPodLogs)
			jamie.POST("/analyze-logs", jamieHandler.AnalyzeLogs)
		}
	}

	return r
}
