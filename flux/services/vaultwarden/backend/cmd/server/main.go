package main

import (
	"log"
	"os"

	"github.com/brunolucena/homelab/vaultwarden/internal/api"
	"github.com/brunolucena/homelab/vaultwarden/internal/auth"
	"github.com/brunolucena/homelab/vaultwarden/internal/vault"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize Vault client
	vaultAddr := getEnv("VAULT_ADDR", "http://vault.vault-system.svc.cluster.local:8200")
	vaultToken := getEnv("VAULT_TOKEN", "")

	vaultClient, err := vault.NewClient(vaultAddr, vaultToken)
	if err != nil {
		log.Fatalf("Failed to initialize Vault client: %v", err)
	}

	// Initialize auth service
	jwtSecret := getEnv("JWT_SECRET", "change-me-in-production")
	authService := auth.NewService(jwtSecret)

	// Initialize API handlers
	handlers := api.NewHandlers(vaultClient, authService)

	// Setup router
	router := gin.Default()

	// CORS middleware
	router.Use(corsMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Bitwarden-compatible endpoints
	v1 := router.Group("/api")
	{
		// Identity endpoints
		identity := v1.Group("/identity")
		{
			identity.POST("/connect/token", handlers.Login)
		}

		// Protected routes
		protected := v1.Group("/")
		protected.Use(authMiddleware(authService))
		{
			// Ciphers (password entries)
			protected.GET("/ciphers", handlers.ListCiphers)
			protected.POST("/ciphers", handlers.CreateCipher)
			protected.GET("/ciphers/:id", handlers.GetCipher)
			protected.PUT("/ciphers/:id", handlers.UpdateCipher)
			protected.DELETE("/ciphers/:id", handlers.DeleteCipher)

			// Profile
			protected.GET("/profile", handlers.GetProfile)
			protected.PUT("/profile", handlers.UpdateProfile)
		}
	}

	// Start server
	port := getEnv("PORT", "8080")
	log.Printf("Starting server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
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

func authMiddleware(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(401, gin.H{"error": "Missing authorization header"})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix if present
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		userID, err := authService.ValidateToken(token)
		if err != nil {
			c.JSON(401, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}
