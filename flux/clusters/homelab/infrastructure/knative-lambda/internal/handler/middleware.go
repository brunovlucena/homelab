// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🔗 MIDDLEWARE - HTTP middleware for observability and security
//
//	🎯 Purpose: HTTP middleware for distributed tracing, metrics, and security
//	💡 Features: Request tracing, metrics collection, correlation IDs, security headers
//
//	🏛️ ARCHITECTURE:
//	🔍 Distributed Tracing - OpenTelemetry span creation and propagation
//	📊 Metrics Collection - Prometheus metrics for HTTP requests
//	🆔 Correlation IDs - Request correlation for log aggregation
//	🛡️ Security Headers - Security and CORS headers
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"

	"knative-lambda-new/internal/observability"
	"knative-lambda-new/internal/resilience"
)

// Middleware represents HTTP middleware functions
type Middleware func(http.Handler) http.Handler

// ObservabilityMiddleware creates middleware for distributed tracing and metrics
func ObservabilityMiddleware(obs *observability.Observability) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Extract trace context from headers
			ctx := r.Context()
			ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(r.Header))

			// Generate correlation ID if not present
			correlationID := r.Header.Get("X-Correlation-ID")
			if correlationID == "" {
				correlationID = uuid.New().String()
			}
			ctx = observability.WithCorrelationID(ctx, correlationID)

			// Create span for the request
			spanName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
			spanCtx, span := obs.StartSpanWithAttributes(ctx, spanName, map[string]string{
				"http.method":     r.Method,
				"http.url":        r.URL.String(),
				"http.user_agent": r.UserAgent(),
				"http.remote_ip":  getClientIP(r),
				"correlation_id":  correlationID,
			})
			defer span.End()

			// Add correlation ID to response headers
			w.Header().Set("X-Correlation-ID", correlationID)

			// Create response writer wrapper for metrics
			responseWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Record request size
			requestSize := r.ContentLength
			if requestSize > 0 && obs.GetMetrics() != nil {
				obs.GetMetrics().CloudEventSize.WithLabelValues(r.Method, r.URL.Path, obs.GetServiceName()).Observe(float64(requestSize))
			}

			// Process request
			next.ServeHTTP(responseWriter, r.WithContext(spanCtx))

			// Record metrics
			duration := time.Since(start).Seconds()
			if obs.GetMetrics() != nil {
				statusCode := strconv.Itoa(responseWriter.statusCode)
				obs.GetMetrics().CloudEventsTotal.WithLabelValues(r.Method, r.URL.Path, statusCode, "http", obs.GetServiceName()).Inc()
				obs.GetMetrics().CloudEventDuration.WithLabelValues(r.Method, r.URL.Path, "http", obs.GetServiceName()).Observe(duration)
				obs.GetMetrics().CloudEventResponseSize.WithLabelValues(r.Method, r.URL.Path, statusCode, obs.GetServiceName()).Observe(float64(responseWriter.size))
			}

			// Record span attributes
			span.SetAttributes(
				attribute.Int("http.status_code", responseWriter.statusCode),
				attribute.Int64("http.request.size", requestSize),
				attribute.Int64("http.response.size", int64(responseWriter.size)),
				attribute.Float64("http.duration", duration),
			)

			// Log request details
			obs.Info(spanCtx, "HTTP request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status_code", responseWriter.statusCode,
				"duration_seconds", duration,
				"request_size", requestSize,
				"response_size", responseWriter.size,
				"user_agent", r.UserAgent(),
				"remote_ip", getClientIP(r),
			)
		})
	}
}

// SecurityMiddleware creates middleware for security headers
func SecurityMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Security headers
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Content-Security-Policy", "default-src 'self'")

			// CORS headers for API endpoints
			if strings.HasPrefix(r.URL.Path, "/api/") {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Correlation-ID")
				w.Header().Set("Access-Control-Max-Age", "86400")

				// Handle preflight requests
				if r.Method == "OPTIONS" {
					w.WriteHeader(http.StatusOK)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitMiddleware creates middleware for rate limiting
func RateLimitMiddleware(rateLimiter *resilience.RateLimiter) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip rate limiting for metrics and health check endpoints
			if r.URL.Path == "/metrics" || r.URL.Path == "/health" {
				next.ServeHTTP(w, r)
				return
			}

			// Check rate limit for other requests
			if !rateLimiter.Allow() {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error": "rate limit exceeded", "message": "Too many requests"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// MetricsMiddleware creates middleware for metrics endpoint
func MetricsMiddleware(obs *observability.Observability) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/metrics" {
				obs.GetMetricsHandler().ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code and size
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(data)
	rw.size += size
	return size, err
}

// getClientIP extracts the real client IP address
func getClientIP(r *http.Request) string {
	// Check for forwarded headers
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if commaIndex := strings.Index(ip, ","); commaIndex != -1 {
			return strings.TrimSpace(ip[:commaIndex])
		}
		return strings.TrimSpace(ip)
	}

	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}

	if ip := r.Header.Get("X-Client-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}

	// Fall back to remote address
	if r.RemoteAddr != "" {
		// Remove port if present
		if colonIndex := strings.LastIndex(r.RemoteAddr, ":"); colonIndex != -1 {
			return r.RemoteAddr[:colonIndex]
		}
		return r.RemoteAddr
	}

	return "unknown"
}

// ChainMiddleware chains multiple middleware functions
func ChainMiddleware(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// CreateDefaultMiddlewareChain creates the default middleware chain
func CreateDefaultMiddlewareChain(obs *observability.Observability, rateLimiter *resilience.RateLimiter) Middleware {
	return ChainMiddleware(
		MetricsMiddleware(obs),
		SecurityMiddleware(),
		RateLimitMiddleware(rateLimiter),
		ObservabilityMiddleware(obs),
	)
}
