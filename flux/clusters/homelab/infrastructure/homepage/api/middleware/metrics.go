package middleware

import (
	"bruno-site/metrics"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
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
		status := c.Writer.Status()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Record HTTP request count with labels
		metrics.HTTPRequestsTotal.Add(c.Request.Context(), 1,
			metric.WithAttributes(
				attribute.String("application", "bruno-site"),
				attribute.String("method", method),
				attribute.String("path", path),
				attribute.String("status_code", strconv.Itoa(status)),
			))

		// Record HTTP request duration
		metrics.HTTPRequestDuration.Record(c.Request.Context(), duration,
			metric.WithAttributes(
				attribute.String("application", "bruno-site"),
				attribute.String("method", method),
				attribute.String("path", path),
				attribute.String("status_code", strconv.Itoa(status)),
			))
	}
}

