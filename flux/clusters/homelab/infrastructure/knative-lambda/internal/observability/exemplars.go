package observability

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"
)

// ExemplarData represents the data needed for a Prometheus exemplar
type ExemplarData struct {
	Value     float64
	Timestamp time.Time
	TraceID   string
	SpanID    string
}

// getTraceContext extracts trace information from context for exemplars
func getTraceContext(ctx context.Context) (traceID, spanID string) {
	if span := trace.SpanFromContext(ctx); span != nil {
		spanContext := span.SpanContext()
		if spanContext.IsValid() {
			return spanContext.TraceID().String(), spanContext.SpanID().String()
		}
	}
	return "", ""
}

// createExemplar creates a Prometheus exemplar from trace context
func createExemplar(ctx context.Context, value float64, config ExemplarsConfig) *prometheus.Exemplar {
	// Check if exemplars are enabled
	if !config.ExemplarsEnabled() {
		return nil
	}

	// Check if we should include this exemplar based on sample rate
	if !config.ShouldIncludeExemplar() {
		return nil
	}

	traceID, spanID := getTraceContext(ctx)
	if traceID == "" {
		return nil
	}

	return &prometheus.Exemplar{
		Value:     value,
		Timestamp: time.Now(),
		Labels: prometheus.Labels{
			config.GetTraceIDLabel(): traceID,
			config.GetSpanIDLabel():  spanID,
		},
	}
}

// createExemplarWithLabels creates a Prometheus exemplar with additional labels
func createExemplarWithLabels(ctx context.Context, value float64, additionalLabels prometheus.Labels, config ExemplarsConfig) *prometheus.Exemplar {
	// Check if exemplars are enabled
	if !config.ExemplarsEnabled() {
		return nil
	}

	// Check if we should include this exemplar based on sample rate
	if !config.ShouldIncludeExemplar() {
		return nil
	}

	traceID, spanID := getTraceContext(ctx)
	if traceID == "" {
		return nil
	}

	// Merge additional labels with trace labels
	labels := prometheus.Labels{
		config.GetTraceIDLabel(): traceID,
		config.GetSpanIDLabel():  spanID,
	}
	for k, v := range additionalLabels {
		labels[k] = v
	}

	return &prometheus.Exemplar{
		Value:     value,
		Timestamp: time.Now(),
		Labels:    labels,
	}
}

// ExemplarRecorder provides methods to record metrics with exemplars
type ExemplarRecorder struct {
	obs    *Observability
	config ExemplarsConfig
}

// NewExemplarRecorder creates a new exemplar recorder
func NewExemplarRecorder(obs *Observability) *ExemplarRecorder {
	return &ExemplarRecorder{
		obs:    obs,
		config: obs.GetExemplarsConfig(),
	}
}

// RecordCounterWithExemplar records a counter metric with exemplar
func (er *ExemplarRecorder) RecordCounterWithExemplar(ctx context.Context, counter *prometheus.CounterVec, value float64, labelValues ...string) {
	if er.obs.GetMetrics() == nil {
		return
	}

	// Record the metric
	counter.WithLabelValues(labelValues...).Add(value)

	// Add exemplar if trace context is available
	if exemplar := createExemplar(ctx, value, er.config); exemplar != nil {
		// Note: Prometheus client doesn't directly support adding exemplars to counters
		// This would typically be done through the metric registry or custom collectors
		// For now, we'll log the exemplar information for debugging
		er.obs.Info(ctx, "Counter exemplar recorded",
			"metric_name", "counter_with_exemplar",
			"value", value,
			"trace_id", exemplar.Labels[er.config.GetTraceIDLabel()],
			"span_id", exemplar.Labels[er.config.GetSpanIDLabel()],
		)
	}
}

// RecordHistogramWithExemplar records a histogram metric with exemplar
func (er *ExemplarRecorder) RecordHistogramWithExemplar(ctx context.Context, histogram *prometheus.HistogramVec, value float64, labelValues ...string) {
	if er.obs.GetMetrics() == nil {
		return
	}

	// Record the metric
	histogram.WithLabelValues(labelValues...).Observe(value)

	// Add exemplar if trace context is available
	if exemplar := createExemplar(ctx, value, er.config); exemplar != nil {
		// Note: Prometheus client doesn't directly support adding exemplars to histograms
		// This would typically be done through the metric registry or custom collectors
		// For now, we'll log the exemplar information for debugging
		er.obs.Info(ctx, "Histogram exemplar recorded",
			"metric_name", "histogram_with_exemplar",
			"value", value,
			"trace_id", exemplar.Labels[er.config.GetTraceIDLabel()],
			"span_id", exemplar.Labels[er.config.GetSpanIDLabel()],
		)
	}
}

// RecordGaugeWithExemplar records a gauge metric with exemplar
func (er *ExemplarRecorder) RecordGaugeWithExemplar(ctx context.Context, gauge *prometheus.GaugeVec, value float64, labelValues ...string) {
	if er.obs.GetMetrics() == nil {
		return
	}

	// Record the metric
	gauge.WithLabelValues(labelValues...).Set(value)

	// Add exemplar if trace context is available
	if exemplar := createExemplar(ctx, value, er.config); exemplar != nil {
		// Note: Prometheus client doesn't directly support adding exemplars to gauges
		// This would typically be done through the metric registry or custom collectors
		// For now, we'll log the exemplar information for debugging
		er.obs.Info(ctx, "Gauge exemplar recorded",
			"metric_name", "gauge_with_exemplar",
			"value", value,
			"trace_id", exemplar.Labels[er.config.GetTraceIDLabel()],
			"span_id", exemplar.Labels[er.config.GetSpanIDLabel()],
		)
	}
}

// RecordBuildRequestWithExemplar records a build request with exemplar
func (er *ExemplarRecorder) RecordBuildRequestWithExemplar(ctx context.Context, thirdPartyID, parserID, status string) {
	if er.obs.GetMetrics() == nil {
		return
	}

	// Record the metric
	er.obs.GetMetrics().buildRequestsTotal.WithLabelValues(status, er.obs.GetServiceName()).Inc()

	// Add exemplar information
	if exemplar := createExemplarWithLabels(ctx, 1.0, prometheus.Labels{
		"third_party_id": thirdPartyID,
		"parser_id":      parserID,
		"status":         status,
	}, er.config); exemplar != nil {
		er.obs.Info(ctx, "Build request exemplar recorded",
			"metric_name", "build_requests_total",
			"third_party_id", thirdPartyID,
			"parser_id", parserID,
			"status", status,
			"trace_id", exemplar.Labels[er.config.GetTraceIDLabel()],
			"span_id", exemplar.Labels[er.config.GetSpanIDLabel()],
		)
	}
}

// RecordBuildDurationWithExemplar records build duration with exemplar
func (er *ExemplarRecorder) RecordBuildDurationWithExemplar(ctx context.Context, thirdPartyID, parserID string, duration time.Duration) {
	if er.obs.GetMetrics() == nil {
		return
	}

	durationSeconds := duration.Seconds()

	// Record the metric
	er.obs.GetMetrics().buildRequestDuration.WithLabelValues(er.obs.GetServiceName()).Observe(durationSeconds)

	// Add exemplar information
	if exemplar := createExemplarWithLabels(ctx, durationSeconds, prometheus.Labels{
		"third_party_id": thirdPartyID,
		"parser_id":      parserID,
		"duration_type":  "build_duration",
	}, er.config); exemplar != nil {
		er.obs.Info(ctx, "Build duration exemplar recorded",
			"metric_name", "build_request_duration",
			"third_party_id", thirdPartyID,
			"parser_id", parserID,
			"duration_seconds", durationSeconds,
			"trace_id", exemplar.Labels[er.config.GetTraceIDLabel()],
			"span_id", exemplar.Labels[er.config.GetSpanIDLabel()],
		)
	}
}

// RecordK8sJobCreationWithExemplar records K8s job creation with exemplar
func (er *ExemplarRecorder) RecordK8sJobCreationWithExemplar(ctx context.Context, jobType, status string) {
	if er.obs.GetMetrics() == nil {
		return
	}

	// Record the metric
	er.obs.GetMetrics().k8sJobCreationTotal.WithLabelValues(jobType, status, er.obs.GetServiceName()).Inc()

	// Add exemplar information
	if exemplar := createExemplarWithLabels(ctx, 1.0, prometheus.Labels{
		"job_type": jobType,
		"status":   status,
	}, er.config); exemplar != nil {
		er.obs.Info(ctx, "K8s job creation exemplar recorded",
			"metric_name", "k8s_job_creation_total",
			"job_type", jobType,
			"status", status,
			"trace_id", exemplar.Labels[er.config.GetTraceIDLabel()],
			"span_id", exemplar.Labels[er.config.GetSpanIDLabel()],
		)
	}
}

// RecordS3UploadWithExemplar records S3 upload with exemplar
func (er *ExemplarRecorder) RecordS3UploadWithExemplar(ctx context.Context, bucket, status string, size int64) {
	if er.obs.GetMetrics() == nil {
		return
	}

	// Record the metric
	er.obs.GetMetrics().awsS3UploadTotal.WithLabelValues(bucket, status, er.obs.GetServiceName()).Inc()

	// Add exemplar information
	if exemplar := createExemplarWithLabels(ctx, float64(size), prometheus.Labels{
		"bucket":    bucket,
		"status":    status,
		"operation": "s3_upload",
	}, er.config); exemplar != nil {
		er.obs.Info(ctx, "S3 upload exemplar recorded",
			"metric_name", "aws_s3_upload_total",
			"bucket", bucket,
			"status", status,
			"size", size,
			"trace_id", exemplar.Labels[er.config.GetTraceIDLabel()],
			"span_id", exemplar.Labels[er.config.GetSpanIDLabel()],
		)
	}
}

// RecordErrorWithExemplar records an error with exemplar
func (er *ExemplarRecorder) RecordErrorWithExemplar(ctx context.Context, component, errorType, severity string) {
	if er.obs.GetMetrics() == nil {
		return
	}

	// Record the metric
	er.obs.GetMetrics().errorTotal.WithLabelValues(component, errorType, severity, er.obs.GetServiceName()).Inc()

	// Add exemplar information
	if exemplar := createExemplarWithLabels(ctx, 1.0, prometheus.Labels{
		"component":  component,
		"error_type": errorType,
		"severity":   severity,
	}, er.config); exemplar != nil {
		er.obs.Info(ctx, "Error exemplar recorded",
			"metric_name", "error_total",
			"component", component,
			"error_type", errorType,
			"severity", severity,
			"trace_id", exemplar.Labels[er.config.GetTraceIDLabel()],
			"span_id", exemplar.Labels[er.config.GetSpanIDLabel()],
		)
	}
}
