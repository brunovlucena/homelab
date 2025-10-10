package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// =============================================================================
// 📊 METRICS HANDLER
// =============================================================================

// MetricsHandler exposes Prometheus metrics endpoint
// Note: This is for Prometheus scraping compatibility
// Actual metrics are recorded via OpenTelemetry and exported to Alloy
func MetricsHandler(c *gin.Context) {
	// Use Prometheus HTTP handler to expose metrics
	handler := promhttp.Handler()
	handler.ServeHTTP(c.Writer, c.Request)
}

