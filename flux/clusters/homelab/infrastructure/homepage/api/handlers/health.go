package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// DependencyChecker interface for checking service dependencies
type DependencyChecker interface {
	CheckHealth() error
}

var (
	// Global dependency checkers - will be set by router
	jamieChecker DependencyChecker
)

// SetJamieChecker sets the Jamie dependency checker
func SetJamieChecker(checker DependencyChecker) {
	jamieChecker = checker
}

// HealthCheck returns the health status of the API and all dependencies
func HealthCheck(c *gin.Context) {
	status := "healthy"
	statusCode := http.StatusOK
	dependencies := make(map[string]interface{})

	// 🤖 Check Jamie service health
	if jamieChecker != nil {
		jamieStart := time.Now()
		if err := jamieChecker.CheckHealth(); err != nil {
			status = "unhealthy"
			statusCode = http.StatusServiceUnavailable
			dependencies["jamie"] = map[string]interface{}{
				"status":       "unhealthy",
				"error":        err.Error(),
				"responseTime": time.Since(jamieStart).Milliseconds(),
			}
		} else {
			dependencies["jamie"] = map[string]interface{}{
				"status":       "healthy",
				"responseTime": time.Since(jamieStart).Milliseconds(),
			}
		}
	}

	c.JSON(statusCode, gin.H{
		"status":       status,
		"service":      "bruno-site-api",
		"version":      "1.0.0",
		"dependencies": dependencies,
	})
}
