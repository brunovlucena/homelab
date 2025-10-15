package handlers

import (
	"github.com/gin-gonic/gin"
)

// PrometheusHandlerFunc is set by main.go to expose OpenTelemetry metrics
var PrometheusHandlerFunc func(*gin.Context)

// =============================================================================
// 📊 METRICS HANDLER
// =============================================================================

// MetricsHandler exposes Prometheus metrics endpoint
// This handler is set by main.go after OpenTelemetry initialization
func MetricsHandler(c *gin.Context) {
	if PrometheusHandlerFunc != nil {
		PrometheusHandlerFunc(c)
	} else {
		c.String(503, "Prometheus exporter not initialized")
	}
}

