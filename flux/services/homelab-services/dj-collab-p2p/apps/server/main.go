package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoClient *mongo.Client
	upgrader    = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins in development
		},
	}
)

func main() {
	// Connect to MongoDB
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	mongoClient = client
	defer client.Disconnect(context.Background())

	// Setup Gin router
	r := gin.Default()

	// CORS middleware
	r.Use(corsMiddleware())

	// API routes
	api := r.Group("/api/v1")
	{
		api.GET("/health", healthCheck)
		api.POST("/sessions", createSession)
		api.GET("/sessions/:id", getSession)
		api.POST("/sessions/:id/join", joinSession)
		api.GET("/sessions/:id/ws", handleWebSocket)
		api.POST("/sessions/:id/state", updateState)
		api.GET("/users/:id", getUser)
		api.POST("/users", createUser)
		api.GET("/leaderboard", getLeaderboard)
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "ok",
		"service": "dj-collab-p2p-server",
	})
}
