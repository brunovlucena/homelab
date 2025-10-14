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
// 🏥 Always returns 200 OK for Kubernetes probes - API is healthy if it can respond
// Dependencies are reported but don't affect overall health status
func HealthCheck(c *gin.Context) {
	status := "healthy"
	statusCode := http.StatusOK
	dependencies := make(map[string]interface{})
	hasDegradedDependencies := false

	// 🤖 Check Jamie service health
	if jamieChecker != nil {
		jamieStart := time.Now()
		if err := jamieChecker.CheckHealth(); err != nil {
			hasDegradedDependencies = true
			dependencies["jamie"] = map[string]interface{}{
				"status":       "degraded",
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

	// API is healthy if it can respond, even if dependencies are down
	// Mark as "degraded" if dependencies are unhealthy, but still return 200 OK
	if hasDegradedDependencies {
		status = "degraded"
	}

	c.JSON(statusCode, gin.H{
		"status":       status,
		"service":      "bruno-site-api",
		"version":      "1.0.0",
		"dependencies": dependencies,
	})
}
