package main

import (
	"context"
	"log"
	"os"
	"time"

	"homepage/config"
	"homepage/database"
	"homepage/handlers"
	"homepage/metrics"
	"homepage/router"
	"homepage/storage"

	"github.com/gin-gonic/gin"
)

// 🌐 Bruno Site API
// Production-ready API server + OpenTelemetry → Alloy → Logfire
func main() {
	ctx := context.Background()

	log.Println("═══════════════════════════════════════════════════════════")
	log.Println("🚀 Homepage API - Initialization Starting")
	log.Println("═══════════════════════════════════════════════════════════")

	// 📊 Initialize OpenTelemetry → Alloy → Logfire
	log.Println("📊 Step 1/7: Initializing OpenTelemetry...")
	shutdown, err := InitOTel(ctx)
	if err != nil {
		log.Printf("❌ ERROR: OpenTelemetry initialization failed: %v", err)
		log.Println("⚠️  WARNING: Continuing without OpenTelemetry tracing")
	} else {
		log.Println("✅ OpenTelemetry tracing configured successfully")

		// Set up Prometheus metrics handler
		promHandler := PrometheusHandler()
		handlers.PrometheusHandlerFunc = func(c *gin.Context) {
			promHandler.ServeHTTP(c.Writer, c.Request)
		}

		defer func() {
			log.Println("🔄 Shutting down OpenTelemetry...")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := shutdown(ctx); err != nil {
				log.Printf("❌ ERROR: OpenTelemetry shutdown failed: %v", err)
			} else {
				log.Println("✅ OpenTelemetry shutdown completed")
			}
		}()
	}

	// 📊 Initialize metrics
	log.Println("📊 Step 2/7: Initializing metrics...")
	if err := metrics.InitMetrics(); err != nil {
		log.Fatalf("❌ FATAL ERROR: Failed to initialize metrics: %v", err)
	}
	log.Println("✅ Metrics initialized: projects, experiences, database, redis, minio")

	// 📊 Initialize frontend metrics
	log.Println("📊 Step 2b/7: Initializing frontend metrics...")
	if err := handlers.InitFrontendMetrics(); err != nil {
		log.Fatalf("❌ FATAL ERROR: Failed to initialize frontend metrics: %v", err)
	}
	log.Println("✅ Frontend metrics initialized: page views, API calls, errors, Web Vitals")

	// 🔧 Load configuration
	log.Println("🔧 Step 3/7: Loading configuration...")
	cfg := config.Load()
	log.Printf("✅ Configuration loaded (CORS: %s)", cfg.CORSOrigin)

	// 🗄️ Initialize database
	log.Println("🗄️  Step 4/7: Initializing PostgreSQL database...")
	db, err := database.Initialize(cfg.DatabaseURL)
	if err != nil {
		log.Printf("❌ ERROR: Database initialization failed: %v", err)
		log.Println("⚠️  WARNING: API will run in degraded mode without database")
		log.Println("⚠️  WARNING: All database-dependent endpoints will return errors")
	} else {
		log.Println("✅ PostgreSQL database connected successfully")
	}

	// 🔴 Initialize Redis
	log.Println("🔴 Step 5/7: Initializing Redis cache...")
	redis, err := database.InitializeRedis(cfg.RedisURL)
	if err != nil {
		log.Printf("❌ ERROR: Redis initialization failed: %v", err)
		log.Println("⚠️  WARNING: API will run without caching capabilities")
	} else {
		log.Println("✅ Redis cache connected successfully")
	}

	// 📦 Initialize MinIO client
	log.Println("📦 Step 6/7: Initializing MinIO object storage...")
	minioClient, err := storage.NewMinIOClient(cfg.MinIO)
	if err != nil {
		log.Printf("❌ ERROR: MinIO initialization failed: %v", err)
		log.Println("⚠️  WARNING: Asset serving will not be available")
	} else {
		log.Printf("✅ MinIO object storage connected (endpoint: %s)", cfg.MinIO.Endpoint)
	}

	// 🚀 Initialize router
	log.Println("🚀 Step 7/7: Initializing HTTP router and middleware...")
	r := router.SetupRouter(cfg, db, redis, minioClient)
	log.Println("✅ HTTP router configured with CORS, compression, and tracing")

	// 🎵 Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("═══════════════════════════════════════════════════════════")
	log.Println("🌐 Homepage API - Ready to Accept Connections")
	log.Println("═══════════════════════════════════════════════════════════")
	log.Printf("📍 Server Address: 0.0.0.0:%s", port)
	log.Printf("🔗 Health Check: http://localhost:%s/health", port)
	log.Printf("📊 Metrics: http://localhost:%s/metrics", port)
	log.Printf("🎨 Frontend Origin: %s", cfg.CORSOrigin)

	// Log service integrations
	log.Println("───────────────────────────────────────────────────────────")
	log.Println("🔌 Service Integrations:")
	if db != nil {
		log.Println("  ✅ PostgreSQL: Connected")
	} else {
		log.Println("  ❌ PostgreSQL: Disconnected")
	}
	if redis != nil {
		log.Println("  ✅ Redis: Connected")
	} else {
		log.Println("  ❌ Redis: Disconnected")
	}
	if minioClient != nil {
		log.Println("  ✅ MinIO: Connected")
	} else {
		log.Println("  ❌ MinIO: Disconnected")
	}
	log.Println("═══════════════════════════════════════════════════════════")

	log.Printf("🎧 Listening on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("❌ FATAL ERROR: Failed to start HTTP server: %v", err)
	}
}
