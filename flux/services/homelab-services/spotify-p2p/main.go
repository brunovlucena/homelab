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
			"service": "spotify-p2p-server",
		})
	})

	// API endpoints
	api := r.Group("/api/v1")
	{
		api.GET("/library", getLibrary)
		api.POST("/library/discover", discoverLibrary)
		api.GET("/stations", getStations)
		api.POST("/stations", createStation)
		api.GET("/peers", getPeers)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Spotify P2P server starting on port %s", port)
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

func discoverLibrary(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "discovering",
	})
}

func getStations(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"stations": []gin.H{},
	})
}

func createStation(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"id":     "station-1",
		"status": "created",
	})
}

func getPeers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"peers": []gin.H{},
	})
}

