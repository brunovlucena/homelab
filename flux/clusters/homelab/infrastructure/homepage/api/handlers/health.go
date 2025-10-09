package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheck returns the health status of the API
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "bruno-site-api",
		"version": "1.0.0",
	})
}
