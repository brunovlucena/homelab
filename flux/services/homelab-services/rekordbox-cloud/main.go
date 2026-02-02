package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "rekordbox-cloud-server",
		})
	})

	// API endpoints
	api := r.Group("/api/v1")
	{
		api.GET("/library", getLibrary)
		api.POST("/library/sync", syncLibrary)
		api.GET("/tracks/:id", getTrack)
		api.POST("/tracks/analyze", analyzeTrack)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Rekordbox Cloud server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func getLibrary(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"library": []gin.H{},
		"total":   0,
	})
}

func syncLibrary(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "syncing",
	})
}

func getTrack(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"id":   c.Param("id"),
		"name": "Track",
	})
}

func analyzeTrack(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "analyzing",
	})
}

