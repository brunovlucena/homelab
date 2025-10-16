package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
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

// =============================================================================
// 📊 FRONTEND METRICS COLLECTION
// =============================================================================

var (
	frontendMeter metric.Meter

	// Frontend metrics instruments
	frontendPageViews       metric.Int64Counter
	frontendAPIRequests     metric.Int64Counter
	frontendAPIErrors       metric.Int64Counter
	frontendAPIDuration     metric.Float64Histogram
	frontendErrors          metric.Int64Counter
	frontendInteractions    metric.Int64Counter
	frontendWebVitals       metric.Float64Histogram
	frontendSessionDuration metric.Float64Histogram
	frontendResourceLoad    metric.Float64Histogram
)

// InitFrontendMetrics initializes OpenTelemetry metrics for frontend
func InitFrontendMetrics() error {
	frontendMeter = otel.Meter("homepage-frontend")

	var err error

	// Page views
	frontendPageViews, err = frontendMeter.Int64Counter(
		"frontend_page_views_total",
		metric.WithDescription("Total number of page views"),
		metric.WithUnit("{view}"),
	)
	if err != nil {
		return err
	}

	// API requests
	frontendAPIRequests, err = frontendMeter.Int64Counter(
		"frontend_api_requests_total",
		metric.WithDescription("Total number of API requests from frontend"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return err
	}

	// API errors
	frontendAPIErrors, err = frontendMeter.Int64Counter(
		"frontend_api_errors_total",
		metric.WithDescription("Total number of API errors from frontend"),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		return err
	}

	// API duration
	frontendAPIDuration, err = frontendMeter.Float64Histogram(
		"frontend_api_request_duration_seconds",
		metric.WithDescription("API request duration from frontend"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	// Errors
	frontendErrors, err = frontendMeter.Int64Counter(
		"frontend_errors_total",
		metric.WithDescription("Total number of frontend errors"),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		return err
	}

	// User interactions
	frontendInteractions, err = frontendMeter.Int64Counter(
		"frontend_user_interactions_total",
		metric.WithDescription("Total number of user interactions"),
		metric.WithUnit("{interaction}"),
	)
	if err != nil {
		return err
	}

	// Web Vitals
	frontendWebVitals, err = frontendMeter.Float64Histogram(
		"frontend_web_vitals",
		metric.WithDescription("Web Vitals metrics (LCP, FID, CLS, FCP, TTFB)"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return err
	}

	// Session duration
	frontendSessionDuration, err = frontendMeter.Float64Histogram(
		"frontend_session_duration_seconds",
		metric.WithDescription("User session duration"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	// Resource load duration
	frontendResourceLoad, err = frontendMeter.Float64Histogram(
		"frontend_resource_load_duration_seconds",
		metric.WithDescription("Resource load duration"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	log.Println("✅ Frontend OpenTelemetry metrics initialized")
	return nil
}

// =============================================================================
// 📊 FRONTEND METRICS REQUEST TYPES
// =============================================================================

// MetricData represents a single metric from the frontend
type MetricData struct {
	Name      string            `json:"name" binding:"required"`
	Value     float64           `json:"value" binding:"required"`
	Labels    map[string]string `json:"labels,omitempty"`
	Timestamp int64             `json:"timestamp,omitempty"`
}

// FrontendMetricsRequest represents the request payload for frontend metrics
type FrontendMetricsRequest struct {
	Metrics []MetricData `json:"metrics" binding:"required"`
}

// =============================================================================
// 📊 FRONTEND METRICS HANDLER
// =============================================================================

// FrontendMetricsHandler receives and processes metrics from the frontend
func FrontendMetricsHandler(c *gin.Context) {
	var req FrontendMetricsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	ctx := context.Background()

	// Process each metric
	for _, metricData := range req.Metrics {
		// Convert labels map to attributes
		attrs := make([]attribute.KeyValue, 0, len(metricData.Labels))
		for key, value := range metricData.Labels {
			attrs = append(attrs, attribute.String(key, value))
		}

		// Record metric based on name
		switch metricData.Name {
		case "frontend_page_views_total":
			frontendPageViews.Add(ctx, int64(metricData.Value), metric.WithAttributes(attrs...))

		case "frontend_api_requests_total":
			frontendAPIRequests.Add(ctx, int64(metricData.Value), metric.WithAttributes(attrs...))

		case "frontend_api_errors_total":
			frontendAPIErrors.Add(ctx, int64(metricData.Value), metric.WithAttributes(attrs...))

		case "frontend_api_request_duration_seconds":
			frontendAPIDuration.Record(ctx, metricData.Value, metric.WithAttributes(attrs...))

		case "frontend_errors_total":
			frontendErrors.Add(ctx, int64(metricData.Value), metric.WithAttributes(attrs...))

		case "frontend_user_interactions_total":
			frontendInteractions.Add(ctx, int64(metricData.Value), metric.WithAttributes(attrs...))

		case "frontend_session_duration_seconds":
			frontendSessionDuration.Record(ctx, metricData.Value, metric.WithAttributes(attrs...))

		case "frontend_resource_load_duration_seconds":
			frontendResourceLoad.Record(ctx, metricData.Value, metric.WithAttributes(attrs...))

		default:
			// For Web Vitals and other histogram metrics
			if len(metricData.Name) > 20 && metricData.Name[:20] == "frontend_web_vitals_" {
				frontendWebVitals.Record(ctx, metricData.Value, metric.WithAttributes(attrs...))
			} else {
				// Log unknown metrics in development
				if gin.Mode() == gin.DebugMode {
					log.Printf("⚠️  Unknown frontend metric: %s", metricData.Name)
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":           "success",
		"metrics_received": len(req.Metrics),
	})
}
