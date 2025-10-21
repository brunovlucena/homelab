// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	📊 OBSERVABILITY - Comprehensive metrics, logging, and tracing system
//
//	🎯 Purpose: Centralized observability with structured logging, metrics, and tracing
//	💡 Features: Structured logging, Prometheus metrics, OpenTelemetry tracing, system monitoring
//
//	🏛️ ARCHITECTURE:
//	📝 Structured Logging - JSON logging with correlation IDs and trace context
//	📊 Metrics Collection - Prometheus metrics for monitoring and alerting
//	🔍 Distributed Tracing - OpenTelemetry tracing for request flows
//	💻 System Monitoring - Runtime performance and resource monitoring
//	🎯 Health Monitoring - Health check endpoints and status
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package observability

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// CorrelationIDKey is the context key for correlation IDs
type correlationIDKey struct{}

var CorrelationIDKey = correlationIDKey{}

// Observability provides centralized observability functionality
type Observability struct {
	logger  *logrus.Logger
	tracer  trace.Tracer
	service string
	version string
	env     string

	// Metrics
	metrics *Metrics

	// System Metrics Collector
	systemMetricsCollector *SystemMetricsCollector

	// Exemplar Recorder
	exemplarRecorder *ExemplarRecorder

	// Exemplars Configuration
	exemplarsConfig ExemplarsConfig
}

// Config holds observability configuration
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	LogLevel       string
	MetricsEnabled bool
	TracingEnabled bool
	OTLPEndpoint   string
	SampleRate     float64
	Exemplars      ExemplarsConfig
}

// Metrics holds all Prometheus metrics
type Metrics struct {
	// CloudEvents Metrics
	CloudEventsTotal       *prometheus.CounterVec
	CloudEventDuration     *prometheus.HistogramVec
	CloudEventSize         *prometheus.HistogramVec
	CloudEventResponseSize *prometheus.HistogramVec

	// Business Logic Metrics
	buildRequestsTotal   *prometheus.CounterVec
	buildRequestDuration *prometheus.HistogramVec
	buildSuccessTotal    *prometheus.CounterVec
	buildFailureTotal    *prometheus.CounterVec
	buildQueueSize       *prometheus.GaugeVec
	buildQueueDuration   *prometheus.HistogramVec

	// Kubernetes Metrics
	k8sJobCreationTotal    *prometheus.CounterVec
	k8sJobCreationDuration *prometheus.HistogramVec
	k8sJobSuccessTotal     *prometheus.CounterVec
	k8sJobFailureTotal     *prometheus.CounterVec
	k8sJobDuration         *prometheus.HistogramVec

	// AWS Metrics
	awsS3UploadTotal    *prometheus.CounterVec
	awsS3UploadDuration *prometheus.HistogramVec
	awsS3UploadSize     *prometheus.HistogramVec
	awsECRPushTotal     *prometheus.CounterVec
	awsECRPushDuration  *prometheus.HistogramVec

	// System Metrics
	systemMemoryUsage *prometheus.GaugeVec
	systemCPUUsage    *prometheus.GaugeVec
	systemGoroutines  *prometheus.GaugeVec
	systemHeapAlloc   *prometheus.GaugeVec

	// Error Metrics
	errorTotal *prometheus.CounterVec
	errorRate  *prometheus.GaugeVec

	// Registry for all metrics
	registry *prometheus.Registry
}

// MetricsRecorder provides easy-to-use methods for recording business metrics
type MetricsRecorder struct {
	obs *Observability
}

// SystemMetricsCollector collects and records system performance metrics
type SystemMetricsCollector struct {
	obs           *Observability
	metricsRec    *MetricsRecorder
	collectTicker *time.Ticker
	stopChan      chan struct{}
}

// New creates a new observability instance
func New(config Config) (*Observability, error) {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	})

	// Set log level
	level, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Initialize tracer
	var tracer trace.Tracer
	if config.TracingEnabled {
		tracer, err = initializeTracer(config)
		if err != nil {
			logger.WithError(err).Warn("Failed to initialize tracer, using noop tracer")
			tracer = noop.NewTracerProvider().Tracer(config.ServiceName)
		}
	} else {
		tracer = noop.NewTracerProvider().Tracer(config.ServiceName)
	}

	// Initialize metrics
	var metrics *Metrics
	if config.MetricsEnabled {
		metrics = initializeMetrics(config.ServiceName, config.ServiceVersion, config.Environment)
	}

	// Set default exemplars config if not provided
	if config.Exemplars.TraceIDLabel == "" {
		config.Exemplars = DefaultExemplarsConfig()
	}

	// Create the observability instance first
	obs := &Observability{
		logger:          logger,
		tracer:          tracer,
		service:         config.ServiceName,
		version:         config.ServiceVersion,
		env:             config.Environment,
		metrics:         metrics,
		exemplarsConfig: config.Exemplars,
	}

	// Initialize system metrics collector if metrics are enabled
	if config.MetricsEnabled {
		obs.systemMetricsCollector = NewSystemMetricsCollector(obs)
		// Note: We'll start the collector externally to allow for proper context management
	}

	// Initialize exemplar recorder
	obs.exemplarRecorder = NewExemplarRecorder(obs)

	return obs, nil
}

// initializeTracer sets up OpenTelemetry tracing
func initializeTracer(config Config) (trace.Tracer, error) {
	// Create OTLP exporter
	exporter, err := otlptracegrpc.New(
		context.Background(),
		otlptracegrpc.WithEndpoint(config.OTLPEndpoint),
		otlptracegrpc.WithInsecure(), // For local development
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create resource with service information
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
			semconv.DeploymentEnvironment(config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(config.SampleRate)),
	)

	// Set global trace provider
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Tracer(config.ServiceName), nil
}

// initializeMetrics creates and registers all Prometheus metrics
func initializeMetrics(serviceName, serviceVersion, environment string) *Metrics {
	registry := prometheus.NewRegistry()

	// Common labels for all metrics
	commonLabels := prometheus.Labels{
		"service": serviceName,
		"version": serviceVersion,
		"env":     environment,
	}

	metrics := &Metrics{
		// CloudEvents Metrics
		CloudEventsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "cloudevents_total",
				Help:        "Total number of CloudEvents processed",
				ConstLabels: commonLabels,
			},
			[]string{"method", "endpoint", "status_code", "handler", "knative_service_name"},
		),

		CloudEventDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        "cloudevent_duration_seconds",
				Help:        "CloudEvent processing duration in seconds",
				Buckets:     prometheus.DefBuckets,
				ConstLabels: commonLabels,
			},
			[]string{"method", "endpoint", "handler", "knative_service_name"},
		),

		CloudEventSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        "cloudevent_size_bytes",
				Help:        "CloudEvent size in bytes",
				Buckets:     prometheus.ExponentialBuckets(100, 10, 8),
				ConstLabels: commonLabels,
			},
			[]string{"method", "endpoint", "knative_service_name"},
		),

		CloudEventResponseSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        "cloudevent_response_size_bytes",
				Help:        "CloudEvent response size in bytes",
				Buckets:     prometheus.ExponentialBuckets(100, 10, 8),
				ConstLabels: commonLabels,
			},
			[]string{"method", "endpoint", "status_code", "knative_service_name"},
		),

		// Business Logic Metrics
		buildRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "build_requests_total",
				Help:        "Total number of build requests",
				ConstLabels: commonLabels,
			},
			[]string{"status", "knative_service_name"},
		),

		buildRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        "build_request_duration_seconds",
				Help:        "Build request duration in seconds",
				Buckets:     prometheus.ExponentialBuckets(1, 2, 10),
				ConstLabels: commonLabels,
			},
			[]string{"knative_service_name"},
		),

		buildSuccessTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "build_success_total",
				Help:        "Total number of successful builds",
				ConstLabels: commonLabels,
			},
			[]string{"knative_service_name"},
		),

		buildFailureTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "build_failure_total",
				Help:        "Total number of failed builds",
				ConstLabels: commonLabels,
			},
			[]string{"error_type", "knative_service_name"},
		),

		buildQueueSize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name:        "build_queue_size",
				Help:        "Current number of builds in queue",
				ConstLabels: commonLabels,
			},
			[]string{"priority", "knative_service_name"},
		),

		buildQueueDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        "build_queue_duration_seconds",
				Help:        "Time builds spend in queue",
				Buckets:     prometheus.DefBuckets,
				ConstLabels: commonLabels,
			},
			[]string{"priority", "knative_service_name"},
		),

		// Kubernetes Metrics
		k8sJobCreationTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "k8s_job_creation_total",
				Help:        "Total number of Kubernetes jobs created",
				ConstLabels: commonLabels,
			},
			[]string{"job_type", "status", "knative_service_name"},
		),

		k8sJobCreationDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        "k8s_job_creation_duration_seconds",
				Help:        "Kubernetes job creation duration in seconds",
				Buckets:     prometheus.DefBuckets,
				ConstLabels: commonLabels,
			},
			[]string{"job_type", "knative_service_name"},
		),

		k8sJobSuccessTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "k8s_job_success_total",
				Help:        "Total number of successful Kubernetes jobs",
				ConstLabels: commonLabels,
			},
			[]string{"job_type", "knative_service_name"},
		),

		k8sJobFailureTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "k8s_job_failure_total",
				Help:        "Total number of failed Kubernetes jobs",
				ConstLabels: commonLabels,
			},
			[]string{"job_type", "error_type", "knative_service_name"},
		),

		k8sJobDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        "k8s_job_duration_seconds",
				Help:        "Kubernetes job duration in seconds",
				Buckets:     prometheus.ExponentialBuckets(1, 2, 10),
				ConstLabels: commonLabels,
			},
			[]string{"job_type", "status", "knative_service_name"},
		),

		// AWS Metrics
		awsS3UploadTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "aws_s3_upload_total",
				Help:        "Total number of S3 uploads",
				ConstLabels: commonLabels,
			},
			[]string{"bucket", "status", "knative_service_name"},
		),

		awsS3UploadDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        "aws_s3_upload_duration_seconds",
				Help:        "S3 upload duration in seconds",
				Buckets:     prometheus.DefBuckets,
				ConstLabels: commonLabels,
			},
			[]string{"bucket", "knative_service_name"},
		),

		awsS3UploadSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        "aws_s3_upload_size_bytes",
				Help:        "S3 upload size in bytes",
				Buckets:     prometheus.ExponentialBuckets(1024, 2, 10),
				ConstLabels: commonLabels,
			},
			[]string{"bucket", "knative_service_name"},
		),

		awsECRPushTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "aws_ecr_push_total",
				Help:        "Total number of ECR image pushes",
				ConstLabels: commonLabels,
			},
			[]string{"repository", "status", "knative_service_name"},
		),

		awsECRPushDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        "aws_ecr_push_duration_seconds",
				Help:        "ECR push duration in seconds",
				Buckets:     prometheus.ExponentialBuckets(1, 2, 10),
				ConstLabels: commonLabels,
			},
			[]string{"repository", "knative_service_name"},
		),

		// System Metrics
		systemMemoryUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name:        "system_memory_usage_bytes",
				Help:        "Current memory usage in bytes",
				ConstLabels: commonLabels,
			},
			[]string{"type", "knative_service_name"},
		),

		systemCPUUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name:        "system_cpu_usage_percent",
				Help:        "Current CPU usage percentage",
				ConstLabels: commonLabels,
			},
			[]string{"type", "knative_service_name"},
		),

		systemGoroutines: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name:        "system_goroutines",
				Help:        "Current number of goroutines",
				ConstLabels: commonLabels,
			},
			[]string{"knative_service_name"},
		),

		systemHeapAlloc: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name:        "system_heap_alloc_bytes",
				Help:        "Current heap allocation in bytes",
				ConstLabels: commonLabels,
			},
			[]string{"knative_service_name"},
		),

		// Error Metrics
		errorTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "error_total",
				Help:        "Total number of errors",
				ConstLabels: commonLabels,
			},
			[]string{"component", "error_type", "severity", "knative_service_name"},
		),

		errorRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name:        "error_rate",
				Help:        "Error rate per minute",
				ConstLabels: commonLabels,
			},
			[]string{"component", "knative_service_name"},
		),

		registry: registry,
	}

	// Register all metrics
	registry.MustRegister(
		metrics.CloudEventsTotal,
		metrics.CloudEventDuration,
		metrics.CloudEventSize,
		metrics.CloudEventResponseSize,
		metrics.buildRequestsTotal,
		metrics.buildRequestDuration,
		metrics.buildSuccessTotal,
		metrics.buildFailureTotal,
		metrics.buildQueueSize,
		metrics.buildQueueDuration,
		metrics.k8sJobCreationTotal,
		metrics.k8sJobCreationDuration,
		metrics.k8sJobSuccessTotal,
		metrics.k8sJobFailureTotal,
		metrics.k8sJobDuration,
		metrics.awsS3UploadTotal,
		metrics.awsS3UploadDuration,
		metrics.awsS3UploadSize,
		metrics.awsECRPushTotal,
		metrics.awsECRPushDuration,
		metrics.systemMemoryUsage,
		metrics.systemCPUUsage,
		metrics.systemGoroutines,
		metrics.systemHeapAlloc,
		metrics.errorTotal,
		metrics.errorRate,
	)

	return metrics
}

// Info logs an info message with structured fields
func (o *Observability) Info(ctx context.Context, message string, fields ...interface{}) {
	entry := o.logger.WithContext(ctx)
	if correlationID := ctx.Value(CorrelationIDKey); correlationID != nil {
		entry = entry.WithField("correlation_id", correlationID)
	}

	// Add trace context if available
	if span := trace.SpanFromContext(ctx); span != nil {
		entry = entry.WithFields(logrus.Fields{
			"trace_id": span.SpanContext().TraceID().String(),
			"span_id":  span.SpanContext().SpanID().String(),
		})
	}

	// Convert fields to logrus fields
	logFields := make(logrus.Fields)
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			logFields[fmt.Sprintf("%v", fields[i])] = fields[i+1]
		}
	}

	entry.WithFields(logFields).Info(message)
}

// Error logs an error message with structured fields
func (o *Observability) Error(ctx context.Context, err error, message string, fields ...interface{}) {
	entry := o.logger.WithContext(ctx).WithError(err)
	if correlationID := ctx.Value(CorrelationIDKey); correlationID != nil {
		entry = entry.WithField("correlation_id", correlationID)
	}

	// Add trace context if available
	if span := trace.SpanFromContext(ctx); span != nil {
		entry = entry.WithFields(logrus.Fields{
			"trace_id": span.SpanContext().TraceID().String(),
			"span_id":  span.SpanContext().SpanID().String(),
		})
	}

	// Convert fields to logrus fields
	logFields := make(logrus.Fields)
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			logFields[fmt.Sprintf("%v", fields[i])] = fields[i+1]
		}
	}

	entry.WithFields(logFields).Error(message)

	// Record error metric
	if o.metrics != nil {
		o.metrics.errorTotal.WithLabelValues("general", "unknown", "error", o.GetServiceName()).Inc()
	}
}

// StartSpan starts a new trace span
func (o *Observability) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return o.tracer.Start(ctx, name, opts...)
}

// StartSpanWithAttributes starts a new trace span with attributes
func (o *Observability) StartSpanWithAttributes(ctx context.Context, name string, attrs map[string]string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	spanOpts := append(opts, trace.WithAttributes(
		attribute.String("service.name", o.service),
		attribute.String("service.version", o.version),
		attribute.String("environment", o.env),
	))

	// Add custom attributes
	for k, v := range attrs {
		spanOpts = append(spanOpts, trace.WithAttributes(attribute.String(k, v)))
	}

	return o.tracer.Start(ctx, name, spanOpts...)
}

// RecordMetric records a custom metric
func (o *Observability) RecordMetric(metricType string, name string, value float64, labels map[string]string) {
	if o.metrics == nil {
		return
	}

	// Convert labels map to slice for Prometheus
	labelValues := make([]string, 0, len(labels))
	for _, v := range labels {
		labelValues = append(labelValues, v)
	}

	// This is a simplified implementation - in a real system you'd have specific metric types
	// For now, we'll use the error counter as an example
	if metricType == "counter" && len(labelValues) > 0 {
		o.metrics.errorTotal.WithLabelValues(labelValues...).Add(value)
	}
}

// GetMetricsHandler returns the Prometheus metrics HTTP handler
func (o *Observability) GetMetricsHandler() http.Handler {
	if o.metrics == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Metrics not enabled"))
		})
	}
	return promhttp.HandlerFor(o.metrics.registry, promhttp.HandlerOpts{})
}

// GetServiceName returns the service name
func (o *Observability) GetServiceName() string {
	return o.service
}

// Shutdown gracefully shuts down the observability system
func (o *Observability) Shutdown(ctx context.Context) error {
	// Create a new context with timeout for shutdown operations
	// This prevents issues with canceled contexts during shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Stop system metrics collector if it exists
	if o.systemMetricsCollector != nil {
		o.systemMetricsCollector.Stop()
	}

	// Flush any pending traces
	if tp, ok := otel.GetTracerProvider().(*sdktrace.TracerProvider); ok {
		if err := tp.Shutdown(shutdownCtx); err != nil {
			// Don't return error for context canceled during shutdown
			if shutdownCtx.Err() == context.Canceled || shutdownCtx.Err() == context.DeadlineExceeded {
				o.logger.Info("Tracer provider shutdown completed (context timeout)")
				return nil
			}
			return fmt.Errorf("failed to shutdown tracer provider: %w", err)
		}
	}
	return nil
}

// RecordSecurityEvent records a security event
func (o *Observability) RecordSecurityEvent(ctx context.Context, eventType string, details map[string]interface{}) {
	o.Info(ctx, "Security event recorded", "event_type", eventType, "details", details)

	// Record security metric
	if o.metrics != nil {
		o.metrics.errorTotal.WithLabelValues("security", eventType, "info", o.GetServiceName()).Inc()
	}
}

// GetMetrics returns the metrics instance for direct access
func (o *Observability) GetMetrics() *Metrics {
	return o.metrics
}

// StartSystemMetricsCollection starts the system metrics collector with the specified interval
func (o *Observability) StartSystemMetricsCollection(ctx context.Context, interval time.Duration) {
	if o.systemMetricsCollector != nil {
		o.systemMetricsCollector.Start(ctx, interval)
	}
}

// StopSystemMetricsCollection stops the system metrics collector
func (o *Observability) StopSystemMetricsCollection() {
	if o.systemMetricsCollector != nil {
		o.systemMetricsCollector.Stop()
	}
}

// GetSystemMetricsCollector returns the system metrics collector instance
func (o *Observability) GetSystemMetricsCollector() *SystemMetricsCollector {
	return o.systemMetricsCollector
}

// GetExemplarRecorder returns the exemplar recorder instance
func (o *Observability) GetExemplarRecorder() *ExemplarRecorder {
	return o.exemplarRecorder
}

// GetExemplarsConfig returns the exemplars configuration
func (o *Observability) GetExemplarsConfig() ExemplarsConfig {
	return o.exemplarsConfig
}

// WithCorrelationID adds a correlation ID to the context
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, CorrelationIDKey, correlationID)
}

// GetCorrelationID retrieves the correlation ID from the context
func GetCorrelationID(ctx context.Context) string {
	if correlationID := ctx.Value(CorrelationIDKey); correlationID != nil {
		return correlationID.(string)
	}
	return ""
}

// =============================================================================
// METRICS RECORDER METHODS
// =============================================================================

// NewMetricsRecorder creates a new metrics recorder
func NewMetricsRecorder(obs *Observability) *MetricsRecorder {
	return &MetricsRecorder{obs: obs}
}

// RecordBuildRequest records a build request metric
func (mr *MetricsRecorder) RecordBuildRequest(ctx context.Context, thirdPartyID, parserID, status string) {
	if mr.obs.GetMetrics() == nil {
		return
	}

	mr.obs.GetMetrics().buildRequestsTotal.WithLabelValues(status, mr.obs.GetServiceName()).Inc()

	// Add span attributes
	if span := trace.SpanFromContext(ctx); span != nil {
		span.SetAttributes(
			attribute.String("build.third_party_id", thirdPartyID),
			attribute.String("build.parser_id", parserID),
			attribute.String("build.status", status),
		)
	}

	// Record exemplar information
	if traceID, spanID := getTraceContext(ctx); traceID != "" {
		mr.obs.Info(ctx, "Build request exemplar recorded",
			"metric_name", "build_requests_total",
			"third_party_id", thirdPartyID,
			"parser_id", parserID,
			"status", status,
			"trace_id", traceID,
			"span_id", spanID,
		)
	}
}

// RecordBuildDuration records build duration metric
func (mr *MetricsRecorder) RecordBuildDuration(ctx context.Context, thirdPartyID, parserID string, duration time.Duration) {
	if mr.obs.GetMetrics() == nil {
		return
	}

	durationSeconds := duration.Seconds()
	mr.obs.GetMetrics().buildRequestDuration.WithLabelValues(mr.obs.GetServiceName()).Observe(durationSeconds)

	// Add span attributes
	if span := trace.SpanFromContext(ctx); span != nil {
		span.SetAttributes(
			attribute.String("build.third_party_id", thirdPartyID),
			attribute.String("build.parser_id", parserID),
			attribute.Float64("build.duration_seconds", durationSeconds),
		)
	}

	// Record exemplar information
	if traceID, spanID := getTraceContext(ctx); traceID != "" {
		mr.obs.Info(ctx, "Build duration exemplar recorded",
			"metric_name", "build_request_duration",
			"third_party_id", thirdPartyID,
			"parser_id", parserID,
			"duration_seconds", durationSeconds,
			"trace_id", traceID,
			"span_id", spanID,
		)
	}
}

// RecordBuildSuccess records a successful build
func (mr *MetricsRecorder) RecordBuildSuccess(ctx context.Context, thirdPartyID, parserID string) {
	if mr.obs.GetMetrics() == nil {
		return
	}

	mr.obs.GetMetrics().buildSuccessTotal.WithLabelValues(mr.obs.GetServiceName()).Inc()

	// Add span attributes
	if span := trace.SpanFromContext(ctx); span != nil {
		span.SetAttributes(
			attribute.String("build.third_party_id", thirdPartyID),
			attribute.String("build.parser_id", parserID),
			attribute.String("build.result", "success"),
		)
	}
}

// RecordBuildFailure records a failed build
func (mr *MetricsRecorder) RecordBuildFailure(ctx context.Context, thirdPartyID, parserID, errorType string) {
	if mr.obs.GetMetrics() == nil {
		return
	}

	mr.obs.GetMetrics().buildFailureTotal.WithLabelValues(errorType, mr.obs.GetServiceName()).Inc()

	// Add span attributes
	if span := trace.SpanFromContext(ctx); span != nil {
		span.SetAttributes(
			attribute.String("build.third_party_id", thirdPartyID),
			attribute.String("build.parser_id", parserID),
			attribute.String("build.result", "failure"),
			attribute.String("build.error_type", errorType),
		)
	}
}

// RecordK8sJobCreation records a Kubernetes job creation
func (mr *MetricsRecorder) RecordK8sJobCreation(ctx context.Context, jobType, status string) {
	if mr.obs.GetMetrics() == nil {
		return
	}

	mr.obs.GetMetrics().k8sJobCreationTotal.WithLabelValues(jobType, status, mr.obs.GetServiceName()).Inc()

	// Add span attributes
	if span := trace.SpanFromContext(ctx); span != nil {
		span.SetAttributes(
			attribute.String("k8s.job_type", jobType),
			attribute.String("k8s.job_status", status),
		)
	}

	// Record exemplar information
	if traceID, spanID := getTraceContext(ctx); traceID != "" {
		mr.obs.Info(ctx, "K8s job creation exemplar recorded",
			"metric_name", "k8s_job_creation_total",
			"job_type", jobType,
			"status", status,
			"trace_id", traceID,
			"span_id", spanID,
		)
	}
}

// RecordK8sJobSuccess records a successful Kubernetes job
func (mr *MetricsRecorder) RecordK8sJobSuccess(ctx context.Context, jobType string) {
	if mr.obs.GetMetrics() == nil {
		return
	}

	mr.obs.GetMetrics().k8sJobSuccessTotal.WithLabelValues(jobType, mr.obs.GetServiceName()).Inc()

	// Add span attributes
	if span := trace.SpanFromContext(ctx); span != nil {
		span.SetAttributes(
			attribute.String("k8s.job_type", jobType),
			attribute.String("k8s.job_result", "success"),
		)
	}
}

// RecordK8sJobFailure records a failed Kubernetes job
func (mr *MetricsRecorder) RecordK8sJobFailure(ctx context.Context, jobType, errorType string) {
	if mr.obs.GetMetrics() == nil {
		return
	}

	mr.obs.GetMetrics().k8sJobFailureTotal.WithLabelValues(jobType, errorType, mr.obs.GetServiceName()).Inc()

	// Add span attributes
	if span := trace.SpanFromContext(ctx); span != nil {
		span.SetAttributes(
			attribute.String("k8s.job_type", jobType),
			attribute.String("k8s.job_result", "failure"),
			attribute.String("k8s.error_type", errorType),
		)
	}
}

// RecordS3Upload records an S3 upload operation
func (mr *MetricsRecorder) RecordS3Upload(ctx context.Context, bucket, status string) {
	if mr.obs.GetMetrics() == nil {
		return
	}

	mr.obs.GetMetrics().awsS3UploadTotal.WithLabelValues(bucket, status, mr.obs.GetServiceName()).Inc()

	// Add span attributes
	if span := trace.SpanFromContext(ctx); span != nil {
		span.SetAttributes(
			attribute.String("aws.s3.bucket", bucket),
			attribute.String("aws.s3.status", status),
		)
	}
}

// RecordS3UploadDuration records S3 upload duration
func (mr *MetricsRecorder) RecordS3UploadDuration(ctx context.Context, bucket string, duration time.Duration) {
	if mr.obs.GetMetrics() == nil {
		return
	}

	mr.obs.GetMetrics().awsS3UploadDuration.WithLabelValues(bucket, mr.obs.GetServiceName()).Observe(duration.Seconds())

	// Add span attributes
	if span := trace.SpanFromContext(ctx); span != nil {
		span.SetAttributes(
			attribute.String("aws.s3.bucket", bucket),
			attribute.Float64("aws.s3.upload_duration_seconds", duration.Seconds()),
		)
	}
}

// RecordECRPush records an ECR push operation
func (mr *MetricsRecorder) RecordECRPush(ctx context.Context, repository, status string) {
	if mr.obs.GetMetrics() == nil {
		return
	}

	mr.obs.GetMetrics().awsECRPushTotal.WithLabelValues(repository, status, mr.obs.GetServiceName()).Inc()

	// Add span attributes
	if span := trace.SpanFromContext(ctx); span != nil {
		span.SetAttributes(
			attribute.String("aws.ecr.repository", repository),
			attribute.String("aws.ecr.status", status),
		)
	}
}

// RecordECRPushDuration records ECR push duration
func (mr *MetricsRecorder) RecordECRPushDuration(ctx context.Context, repository string, duration time.Duration) {
	if mr.obs.GetMetrics() == nil {
		return
	}

	mr.obs.GetMetrics().awsECRPushDuration.WithLabelValues(repository, mr.obs.GetServiceName()).Observe(duration.Seconds())

	// Add span attributes
	if span := trace.SpanFromContext(ctx); span != nil {
		span.SetAttributes(
			attribute.String("aws.ecr.repository", repository),
			attribute.Float64("aws.ecr.push_duration_seconds", duration.Seconds()),
		)
	}
}

// RecordS3UploadSize records S3 upload size
func (mr *MetricsRecorder) RecordS3UploadSize(ctx context.Context, bucket string, size int64) {
	if mr.obs.GetMetrics() == nil {
		return
	}

	mr.obs.GetMetrics().awsS3UploadSize.WithLabelValues(bucket, mr.obs.GetServiceName()).Observe(float64(size))

	// Add span attributes
	if span := trace.SpanFromContext(ctx); span != nil {
		span.SetAttributes(
			attribute.String("aws.s3.bucket", bucket),
			attribute.Int64("aws.s3.upload_size_bytes", size),
		)
	}
}

// RecordError records an error metric
func (mr *MetricsRecorder) RecordError(ctx context.Context, component, errorType, severity string) {
	if mr.obs.GetMetrics() == nil {
		return
	}

	mr.obs.GetMetrics().errorTotal.WithLabelValues(component, errorType, severity, mr.obs.GetServiceName()).Inc()

	// Add span attributes
	if span := trace.SpanFromContext(ctx); span != nil {
		span.SetAttributes(
			attribute.String("error.component", component),
			attribute.String("error.type", errorType),
			attribute.String("error.severity", severity),
		)
	}

	// Record exemplar information
	if traceID, spanID := getTraceContext(ctx); traceID != "" {
		mr.obs.Info(ctx, "Error exemplar recorded",
			"metric_name", "error_total",
			"component", component,
			"error_type", errorType,
			"severity", severity,
			"trace_id", traceID,
			"span_id", spanID,
		)
	}
}

// RecordSystemMetrics records current system metrics
func (mr *MetricsRecorder) RecordSystemMetrics() {
	if mr.obs.GetMetrics() == nil {
		return
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Memory metrics
	mr.obs.GetMetrics().systemMemoryUsage.WithLabelValues("alloc", mr.obs.GetServiceName()).Set(float64(m.Alloc))
	mr.obs.GetMetrics().systemMemoryUsage.WithLabelValues("total_alloc", mr.obs.GetServiceName()).Set(float64(m.TotalAlloc))
	mr.obs.GetMetrics().systemMemoryUsage.WithLabelValues("sys", mr.obs.GetServiceName()).Set(float64(m.Sys))
	mr.obs.GetMetrics().systemMemoryUsage.WithLabelValues("heap_alloc", mr.obs.GetServiceName()).Set(float64(m.HeapAlloc))
	mr.obs.GetMetrics().systemMemoryUsage.WithLabelValues("heap_sys", mr.obs.GetServiceName()).Set(float64(m.HeapSys))

	// Goroutine count
	mr.obs.GetMetrics().systemGoroutines.WithLabelValues(mr.obs.GetServiceName()).Set(float64(runtime.NumGoroutine()))

	// Heap allocation
	mr.obs.GetMetrics().systemHeapAlloc.WithLabelValues(mr.obs.GetServiceName()).Set(float64(m.HeapAlloc))
}

// =============================================================================
// SYSTEM METRICS COLLECTOR METHODS
// =============================================================================

// NewSystemMetricsCollector creates a new system metrics collector
func NewSystemMetricsCollector(obs *Observability) *SystemMetricsCollector {
	return &SystemMetricsCollector{
		obs:        obs,
		metricsRec: NewMetricsRecorder(obs),
		stopChan:   make(chan struct{}),
	}
}

// Start begins collecting system metrics at the specified interval
func (smc *SystemMetricsCollector) Start(ctx context.Context, interval time.Duration) {
	smc.collectTicker = time.NewTicker(interval)

	// Collect initial metrics
	smc.collectSystemMetrics(ctx)

	go func() {
		for {
			select {
			case <-smc.collectTicker.C:
				smc.collectSystemMetrics(ctx)
			case <-smc.stopChan:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop stops the system metrics collection
func (smc *SystemMetricsCollector) Stop() {
	if smc.collectTicker != nil {
		smc.collectTicker.Stop()
	}
	close(smc.stopChan)
}

// collectSystemMetrics collects all system metrics
func (smc *SystemMetricsCollector) collectSystemMetrics(ctx context.Context) {
	// Create a span for metrics collection
	spanCtx, span := smc.obs.StartSpan(ctx, "system_metrics_collection")
	defer span.End()

	// Collect memory metrics
	smc.collectMemoryMetrics(spanCtx)

	// Collect CPU metrics
	smc.collectCPUMetrics(spanCtx)

	// Collect goroutine metrics
	smc.collectGoroutineMetrics(spanCtx)

	// Collect GC metrics
	smc.collectGCMetrics(spanCtx)

	// Record overall system metrics
	smc.metricsRec.RecordSystemMetrics()
}

// collectMemoryMetrics collects memory usage metrics
func (smc *SystemMetricsCollector) collectMemoryMetrics(ctx context.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Record memory metrics
	if smc.obs.GetMetrics() != nil {
		smc.obs.GetMetrics().systemMemoryUsage.WithLabelValues("alloc", smc.obs.GetServiceName()).Set(float64(m.Alloc))
		smc.obs.GetMetrics().systemMemoryUsage.WithLabelValues("total_alloc", smc.obs.GetServiceName()).Set(float64(m.TotalAlloc))
		smc.obs.GetMetrics().systemMemoryUsage.WithLabelValues("sys", smc.obs.GetServiceName()).Set(float64(m.Sys))
		smc.obs.GetMetrics().systemMemoryUsage.WithLabelValues("heap_alloc", smc.obs.GetServiceName()).Set(float64(m.HeapAlloc))
		smc.obs.GetMetrics().systemMemoryUsage.WithLabelValues("heap_sys", smc.obs.GetServiceName()).Set(float64(m.HeapSys))
		smc.obs.GetMetrics().systemMemoryUsage.WithLabelValues("heap_idle", smc.obs.GetServiceName()).Set(float64(m.HeapIdle))
		smc.obs.GetMetrics().systemMemoryUsage.WithLabelValues("heap_inuse", smc.obs.GetServiceName()).Set(float64(m.HeapInuse))
		smc.obs.GetMetrics().systemMemoryUsage.WithLabelValues("heap_released", smc.obs.GetServiceName()).Set(float64(m.HeapReleased))
		smc.obs.GetMetrics().systemMemoryUsage.WithLabelValues("heap_objects", smc.obs.GetServiceName()).Set(float64(m.HeapObjects))
		smc.obs.GetMetrics().systemMemoryUsage.WithLabelValues("stack_inuse", smc.obs.GetServiceName()).Set(float64(m.StackInuse))
		smc.obs.GetMetrics().systemMemoryUsage.WithLabelValues("stack_sys", smc.obs.GetServiceName()).Set(float64(m.StackSys))
		smc.obs.GetMetrics().systemMemoryUsage.WithLabelValues("mspan_inuse", smc.obs.GetServiceName()).Set(float64(m.MSpanInuse))
		smc.obs.GetMetrics().systemMemoryUsage.WithLabelValues("mspan_sys", smc.obs.GetServiceName()).Set(float64(m.MSpanSys))
		smc.obs.GetMetrics().systemMemoryUsage.WithLabelValues("mcache_inuse", smc.obs.GetServiceName()).Set(float64(m.MCacheInuse))
		smc.obs.GetMetrics().systemMemoryUsage.WithLabelValues("mcache_sys", smc.obs.GetServiceName()).Set(float64(m.MCacheSys))
		smc.obs.GetMetrics().systemMemoryUsage.WithLabelValues("buck_hash_sys", smc.obs.GetServiceName()).Set(float64(m.BuckHashSys))
		smc.obs.GetMetrics().systemMemoryUsage.WithLabelValues("gc_sys", smc.obs.GetServiceName()).Set(float64(m.GCSys))
		smc.obs.GetMetrics().systemMemoryUsage.WithLabelValues("other_sys", smc.obs.GetServiceName()).Set(float64(m.OtherSys))
	}

	// Add span attributes
	if span := trace.SpanFromContext(ctx); span != nil {
		span.SetAttributes(
			attribute.Int64("memory.alloc_bytes", int64(m.Alloc)),
			attribute.Int64("memory.total_alloc_bytes", int64(m.TotalAlloc)),
			attribute.Int64("memory.sys_bytes", int64(m.Sys)),
			attribute.Int64("memory.heap_alloc_bytes", int64(m.HeapAlloc)),
			attribute.Int64("memory.heap_sys_bytes", int64(m.HeapSys)),
			attribute.Int64("memory.heap_objects", int64(m.HeapObjects)),
		)
	}

	// Log memory usage if it's high
	if m.HeapAlloc > 100*1024*1024 { // 100MB threshold
		smc.obs.Info(ctx, "High memory usage detected",
			"heap_alloc_mb", m.HeapAlloc/1024/1024,
			"heap_sys_mb", m.HeapSys/1024/1024,
			"heap_objects", m.HeapObjects,
		)
	}
}

// collectCPUMetrics collects CPU usage metrics
func (smc *SystemMetricsCollector) collectCPUMetrics(ctx context.Context) {
	// Get CPU count
	numCPU := runtime.NumCPU()

	// Get current CPU usage (simplified - in a real implementation you'd use more sophisticated CPU monitoring)
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Calculate CPU usage based on GC time (simplified metric)
	gcCPUPercent := float64(m.PauseTotalNs) / float64(time.Now().UnixNano()) * 100

	if smc.obs.GetMetrics() != nil {
		smc.obs.GetMetrics().systemCPUUsage.WithLabelValues("gc", smc.obs.GetServiceName()).Set(gcCPUPercent)
		smc.obs.GetMetrics().systemCPUUsage.WithLabelValues("cores", smc.obs.GetServiceName()).Set(float64(numCPU))
	}

	// Add span attributes
	if span := trace.SpanFromContext(ctx); span != nil {
		span.SetAttributes(
			attribute.Int("cpu.cores", numCPU),
			attribute.Float64("cpu.gc_percent", gcCPUPercent),
		)
	}
}

// collectGoroutineMetrics collects goroutine metrics
func (smc *SystemMetricsCollector) collectGoroutineMetrics(ctx context.Context) {
	numGoroutines := runtime.NumGoroutine()

	if smc.obs.GetMetrics() != nil {
		smc.obs.GetMetrics().systemGoroutines.WithLabelValues(smc.obs.GetServiceName()).Set(float64(numGoroutines))
	}

	// Add span attributes
	if span := trace.SpanFromContext(ctx); span != nil {
		span.SetAttributes(
			attribute.Int("goroutines.count", numGoroutines),
		)
	}

	// Log if goroutine count is high
	if numGoroutines > 1000 {
		smc.obs.Info(ctx, "High goroutine count detected",
			"goroutine_count", numGoroutines,
		)
	}
}

// collectGCMetrics collects garbage collection metrics
func (smc *SystemMetricsCollector) collectGCMetrics(ctx context.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Calculate GC metrics
	gcCycles := m.NumGC
	gcPauseTotal := m.PauseTotalNs
	gcPauseNs := m.PauseNs[(gcCycles+255)%256] // Last GC pause

	// Add span attributes
	if span := trace.SpanFromContext(ctx); span != nil {
		span.SetAttributes(
			attribute.Int64("gc.cycles", int64(gcCycles)),
			attribute.Int64("gc.pause_total_ns", int64(gcPauseTotal)),
			attribute.Int64("gc.last_pause_ns", int64(gcPauseNs)),
		)
	}

	// Log if GC is taking too long
	if gcPauseNs > uint64(10*time.Millisecond.Nanoseconds()) {
		smc.obs.Info(ctx, "Long GC pause detected",
			"gc_pause_ms", float64(gcPauseNs)/float64(time.Millisecond.Nanoseconds()),
			"gc_cycles", gcCycles,
		)
	}
}

// GetMemoryStats returns current memory statistics
func (smc *SystemMetricsCollector) GetMemoryStats() runtime.MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m
}

// GetGoroutineCount returns current goroutine count
func (smc *SystemMetricsCollector) GetGoroutineCount() int {
	return runtime.NumGoroutine()
}

// GetCPUCount returns the number of CPU cores
func (smc *SystemMetricsCollector) GetCPUCount() int {
	return runtime.NumCPU()
}

// CollectMetricsOnDemand collects metrics immediately (useful for testing or manual collection)
func (smc *SystemMetricsCollector) CollectMetricsOnDemand(ctx context.Context) {
	smc.collectSystemMetrics(ctx)
}
