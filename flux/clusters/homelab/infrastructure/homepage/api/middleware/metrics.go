package middleware

import (
	"strconv"
	"time"

	"github.com/brunovlucena/homelab/homepage-api/metrics"

	"github.com/gin-gonic/gin"
)

// HTTPMetricsMiddleware records HTTP request metrics for Prometheus
// Records: latency, traffic, errors for Golden Signals monitoring
func HTTPMetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		path := c.Request.URL.Path
		method := c.Request.Method

		// Record HTTP request count with labels
		metrics.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()

		// Record HTTP request duration
		metrics.HTTPRequestDuration.WithLabelValues(method, path, status).Observe(duration)
	}
}
