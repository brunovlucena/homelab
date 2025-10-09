package main

import (
	"log"
	"os"

	"bruno-site/config"
	"bruno-site/database"
	"bruno-site/router"
	"bruno-site/storage"
)

// 🌐 Bruno Site API
// Production-ready API server with MinIO integration
func main() {
	// 🔧 Load configuration
	cfg := config.Load()

	// 🗄️ Initialize database
	db, err := database.Initialize(cfg.DatabaseURL)
	if err != nil {
		log.Printf("⚠️  Warning: Database initialization failed: %v", err)
		log.Println("Continuing without database...")
	}

	// 🔴 Initialize Redis
	redis, err := database.InitializeRedis(cfg.RedisURL)
	if err != nil {
		log.Printf("⚠️  Warning: Redis initialization failed: %v", err)
		log.Println("Continuing without Redis...")
	}

	// 📦 Initialize MinIO client
	minioClient, err := storage.NewMinIOClient(cfg.MinIO)
	if err != nil {
		log.Printf("⚠️  Warning: MinIO initialization failed: %v", err)
		log.Println("Continuing without MinIO...")
	}

	// 🚀 Initialize router
	r := router.SetupRouter(cfg, db, redis, minioClient)

	// 🎵 Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🌐 Bruno Site API starting on port %s", port)
	log.Printf("🎨 Frontend API URL: %s", cfg.CORSOrigin)
	if minioClient != nil {
		log.Printf("📦 MinIO integration enabled")
	}

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
