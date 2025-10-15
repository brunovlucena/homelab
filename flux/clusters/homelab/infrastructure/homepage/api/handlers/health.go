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
	agentBrunoChecker DependencyChecker
)

// SetAgentBrunoChecker sets the Agent Bruno dependency checker
func SetAgentBrunoChecker(checker DependencyChecker) {
	agentBrunoChecker = checker
}

// HealthCheck returns the health status of the API and all dependencies
// 🏥 Always returns 200 OK for Kubernetes probes - API is healthy if it can respond
// Dependencies are reported but don't affect overall health status
func HealthCheck(c *gin.Context) {
	status := "healthy"
	statusCode := http.StatusOK
	dependencies := make(map[string]interface{})
	hasDegradedDependencies := false

	// 🤖 Check Agent Bruno service health
	if agentBrunoChecker != nil {
		agentBrunoStart := time.Now()
		if err := agentBrunoChecker.CheckHealth(); err != nil {
			hasDegradedDependencies = true
			dependencies["agent-bruno"] = map[string]interface{}{
				"status":       "degraded",
				"error":        err.Error(),
				"responseTime": time.Since(agentBrunoStart).Milliseconds(),
			}
		} else {
			dependencies["agent-bruno"] = map[string]interface{}{
				"status":       "healthy",
				"responseTime": time.Since(agentBrunoStart).Milliseconds(),
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
