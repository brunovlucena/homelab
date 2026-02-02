// Package observability provides OpenTelemetry instrumentation for the knative-lambda operator.
// It includes distributed tracing (Tempo), metrics with exemplars (Prometheus), and structured logging (Loki).
//
// Features:
// - W3C Trace Context propagation (traceparent header)
// - OTLP exporter to Tempo endpoint
// - Configurable trace sampling
// - K8s resource attributes (pod name, namespace, etc.)
// - Span creation for key operations (reconcile, build, deploy, CloudEvents)
package observability

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	// ServiceName is the default service name for the operator
	ServiceName = "knative-lambda-operator"
	// ServiceNamespace is the default namespace
	ServiceNamespace = "knative-lambda"

	// Span names for key operations
	SpanNameReconcile         = "reconcile"
	SpanNameReconcilePhase    = "reconcile.phase"
	SpanNameBuildContext      = "build.create_context"
	SpanNameBuildJob          = "build.create_job"
	SpanNameBuildStatus       = "build.get_status"
	SpanNameDeployService     = "deploy.create_service"
	SpanNameDeployStatus      = "deploy.get_status"
	SpanNameEventingReconcile = "eventing.reconcile"
	SpanNameCloudEventReceive = "cloudevents.receive"
	SpanNameCloudEventProcess = "cloudevents.process"
)

// Config holds configuration for the observability stack
type Config struct {
	// ServiceName is the name of the service
	ServiceName string
	// ServiceNamespace is the Kubernetes namespace
	ServiceNamespace string
	// ServiceVersion is the version of the service
	ServiceVersion string
	// Environment is the deployment environment (production, staging, etc.)
	Environment string
	// OTLPEndpoint is the OTLP collector endpoint (default: alloy.observability.svc:4317)
	OTLPEndpoint string
	// TracingSamplingRate is the sampling rate for traces (0.0 - 1.0)
	// Default: 1.0 (100%) for development, configure lower for production
	TracingSamplingRate float64
	// MetricsEnabled enables OTEL metrics export
	MetricsEnabled bool
	// TracingEnabled enables distributed tracing
	TracingEnabled bool

	// K8s resource attributes
	// PodName is the name of the current pod (from POD_NAME env var)
	PodName string
	// PodNamespace is the namespace of the current pod (from POD_NAMESPACE env var)
	PodNamespace string
	// NodeName is the name of the node (from NODE_NAME env var)
	NodeName string
}

// DefaultConfig returns a default observability configuration
// K8s resource attributes are read from environment variables injected via fieldRef
func DefaultConfig() Config {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "tempo.tempo.svc:4317"
	}

	// Parse sampling rate from env, default to 1.0 (100%) for development
	samplingRate := 1.0
	if rateStr := os.Getenv("OTEL_TRACES_SAMPLER_ARG"); rateStr != "" {
		if parsed, err := parseFloat(rateStr); err == nil {
			samplingRate = parsed
		}
	}

	// Check if tracing is enabled (default: true)
	tracingEnabled := os.Getenv("OTEL_TRACING_ENABLED") != "false"

	return Config{
		ServiceName:         getEnvOrDefault("OTEL_SERVICE_NAME", ServiceName),
		ServiceNamespace:    getEnvOrDefault("OTEL_SERVICE_NAMESPACE", ServiceNamespace),
		ServiceVersion:      os.Getenv("VERSION"),
		Environment:         getEnvOrDefault("ENVIRONMENT", "production"),
		OTLPEndpoint:        endpoint,
		TracingSamplingRate: samplingRate,
		MetricsEnabled:      true,
		TracingEnabled:      tracingEnabled,

		// K8s resource attributes from downward API
		PodName:      os.Getenv("POD_NAME"),
		PodNamespace: os.Getenv("POD_NAMESPACE"),
		NodeName:     os.Getenv("NODE_NAME"),
	}
}

// getEnvOrDefault returns the environment variable value or a default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseFloat parses a string to float64
func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

// Provider manages the observability stack
type Provider struct {
	config         Config
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *metric.MeterProvider
	Tracer         trace.Tracer
}

// NewProvider creates a new observability provider
func NewProvider(cfg Config) (*Provider, error) {
	p := &Provider{
		config: cfg,
	}

	// Build resource attributes
	attrs := []attribute.KeyValue{
		// Service attributes
		semconv.ServiceName(cfg.ServiceName),
		semconv.ServiceNamespace(cfg.ServiceNamespace),
		semconv.ServiceVersion(cfg.ServiceVersion),
		semconv.DeploymentEnvironment(cfg.Environment),
		attribute.String("faas.runtime", "go"),
	}

	// Add K8s resource attributes if available
	if cfg.PodName != "" {
		attrs = append(attrs, semconv.K8SPodName(cfg.PodName))
	}
	if cfg.PodNamespace != "" {
		attrs = append(attrs, semconv.K8SNamespaceName(cfg.PodNamespace))
	}
	if cfg.NodeName != "" {
		attrs = append(attrs, semconv.K8SNodeName(cfg.NodeName))
	}

	// Create resource with service and K8s information
	res, err := resource.New(context.Background(),
		resource.WithAttributes(attrs...),
		resource.WithHost(),
		resource.WithProcess(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Setup tracing if enabled
	if cfg.TracingEnabled {
		if err := p.setupTracing(res); err != nil {
			return nil, fmt.Errorf("failed to setup tracing: %w", err)
		}
	}

	// Setup metrics if enabled
	if cfg.MetricsEnabled {
		if err := p.setupMetrics(res); err != nil {
			return nil, fmt.Errorf("failed to setup metrics: %w", err)
		}
	}

	// Set up W3C trace context propagation (traceparent header)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return p, nil
}

// setupTracing configures the OTEL tracing pipeline
func (p *Provider) setupTracing(res *resource.Resource) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create OTLP trace exporter
	conn, err := grpc.DialContext(ctx, p.config.OTLPEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to OTLP endpoint: %w", err)
	}

	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Create sampler based on config
	var sampler sdktrace.Sampler
	if p.config.TracingSamplingRate >= 1.0 {
		sampler = sdktrace.AlwaysSample()
	} else if p.config.TracingSamplingRate <= 0.0 {
		sampler = sdktrace.NeverSample()
	} else {
		sampler = sdktrace.TraceIDRatioBased(p.config.TracingSamplingRate)
	}

	// Create tracer provider
	p.tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(traceExporter,
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithMaxExportBatchSize(512),
		),
		sdktrace.WithSampler(sdktrace.ParentBased(sampler)),
	)

	otel.SetTracerProvider(p.tracerProvider)
	p.Tracer = p.tracerProvider.Tracer(p.config.ServiceName)

	return nil
}

// setupMetrics configures the OTEL metrics pipeline
func (p *Provider) setupMetrics(res *resource.Resource) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create OTLP metrics exporter
	metricExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(p.config.OTLPEndpoint),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return fmt.Errorf("failed to create metric exporter: %w", err)
	}

	// Create meter provider
	p.meterProvider = metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			metric.WithInterval(15*time.Second),
		)),
	)

	otel.SetMeterProvider(p.meterProvider)

	return nil
}

// Shutdown gracefully shuts down the observability stack
func (p *Provider) Shutdown(ctx context.Context) error {
	var errs []error

	if p.tracerProvider != nil {
		if err := p.tracerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("tracer shutdown: %w", err))
		}
	}

	if p.meterProvider != nil {
		if err := p.meterProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("meter shutdown: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}
	return nil
}

// StartSpan starts a new span with common lambda attributes
func (p *Provider) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if p.Tracer == nil {
		return ctx, trace.SpanFromContext(ctx)
	}
	return p.Tracer.Start(ctx, name, opts...)
}

// StartReconcileSpan starts a span for the main reconcile operation
func (p *Provider) StartReconcileSpan(ctx context.Context, functionName, namespace string) (context.Context, trace.Span) {
	return p.StartSpan(ctx, SpanNameReconcile,
		trace.WithAttributes(
			attribute.String("lambda.function", functionName),
			attribute.String("lambda.namespace", namespace),
			attribute.String("operation.type", "reconcile"),
		),
		trace.WithSpanKind(trace.SpanKindInternal),
	)
}

// StartReconcilePhaseSpan starts a span for a specific reconcile phase
func (p *Provider) StartReconcilePhaseSpan(ctx context.Context, functionName, namespace, phase string) (context.Context, trace.Span) {
	return p.StartSpan(ctx, SpanNameReconcilePhase,
		trace.WithAttributes(
			attribute.String("lambda.function", functionName),
			attribute.String("lambda.namespace", namespace),
			attribute.String("lambda.phase", phase),
			attribute.String("operation.type", "reconcile.phase"),
		),
		trace.WithSpanKind(trace.SpanKindInternal),
	)
}

// StartBuildContextSpan starts a span for build context creation
func (p *Provider) StartBuildContextSpan(ctx context.Context, functionName, namespace, runtime string) (context.Context, trace.Span) {
	return p.StartSpan(ctx, SpanNameBuildContext,
		trace.WithAttributes(
			attribute.String("lambda.function", functionName),
			attribute.String("lambda.namespace", namespace),
			attribute.String("lambda.runtime", runtime),
			attribute.String("operation.type", "build.context"),
		),
		trace.WithSpanKind(trace.SpanKindInternal),
	)
}

// StartBuildJobSpan starts a span for build job creation
func (p *Provider) StartBuildJobSpan(ctx context.Context, functionName, namespace, jobName string) (context.Context, trace.Span) {
	return p.StartSpan(ctx, SpanNameBuildJob,
		trace.WithAttributes(
			attribute.String("lambda.function", functionName),
			attribute.String("lambda.namespace", namespace),
			attribute.String("build.job_name", jobName),
			attribute.String("operation.type", "build.job"),
		),
		trace.WithSpanKind(trace.SpanKindInternal),
	)
}

// StartBuildStatusSpan starts a span for checking build status
func (p *Provider) StartBuildStatusSpan(ctx context.Context, namespace, jobName string) (context.Context, trace.Span) {
	return p.StartSpan(ctx, SpanNameBuildStatus,
		trace.WithAttributes(
			attribute.String("lambda.namespace", namespace),
			attribute.String("build.job_name", jobName),
			attribute.String("operation.type", "build.status"),
		),
		trace.WithSpanKind(trace.SpanKindInternal),
	)
}

// StartDeployServiceSpan starts a span for Knative service deployment
func (p *Provider) StartDeployServiceSpan(ctx context.Context, functionName, namespace string) (context.Context, trace.Span) {
	return p.StartSpan(ctx, SpanNameDeployService,
		trace.WithAttributes(
			attribute.String("lambda.function", functionName),
			attribute.String("lambda.namespace", namespace),
			attribute.String("operation.type", "deploy.service"),
		),
		trace.WithSpanKind(trace.SpanKindInternal),
	)
}

// StartDeployStatusSpan starts a span for checking deployment status
func (p *Provider) StartDeployStatusSpan(ctx context.Context, functionName, namespace string) (context.Context, trace.Span) {
	return p.StartSpan(ctx, SpanNameDeployStatus,
		trace.WithAttributes(
			attribute.String("lambda.function", functionName),
			attribute.String("lambda.namespace", namespace),
			attribute.String("operation.type", "deploy.status"),
		),
		trace.WithSpanKind(trace.SpanKindInternal),
	)
}

// StartEventingReconcileSpan starts a span for eventing infrastructure reconciliation
func (p *Provider) StartEventingReconcileSpan(ctx context.Context, functionName, namespace string) (context.Context, trace.Span) {
	return p.StartSpan(ctx, SpanNameEventingReconcile,
		trace.WithAttributes(
			attribute.String("lambda.function", functionName),
			attribute.String("lambda.namespace", namespace),
			attribute.String("operation.type", "eventing.reconcile"),
		),
		trace.WithSpanKind(trace.SpanKindInternal),
	)
}

// StartCloudEventReceiveSpan starts a span for CloudEvent reception
func (p *Provider) StartCloudEventReceiveSpan(ctx context.Context, eventType, eventSource, eventID string) (context.Context, trace.Span) {
	return p.StartSpan(ctx, SpanNameCloudEventReceive,
		trace.WithAttributes(
			attribute.String("cloudevents.type", eventType),
			attribute.String("cloudevents.source", eventSource),
			attribute.String("cloudevents.id", eventID),
			attribute.String("operation.type", "cloudevents.receive"),
		),
		trace.WithSpanKind(trace.SpanKindConsumer),
	)
}

// StartCloudEventProcessSpan starts a span for CloudEvent processing
func (p *Provider) StartCloudEventProcessSpan(ctx context.Context, eventType, functionName, namespace string) (context.Context, trace.Span) {
	return p.StartSpan(ctx, SpanNameCloudEventProcess,
		trace.WithAttributes(
			attribute.String("cloudevents.type", eventType),
			attribute.String("lambda.function", functionName),
			attribute.String("lambda.namespace", namespace),
			attribute.String("operation.type", "cloudevents.process"),
		),
		trace.WithSpanKind(trace.SpanKindInternal),
	)
}

// RecordError records an error on a span
func RecordError(span trace.Span, err error, msg string) {
	if span == nil || err == nil {
		return
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, msg)
}

// SetSpanOK marks a span as successful
func SetSpanOK(span trace.Span) {
	if span == nil {
		return
	}
	span.SetStatus(codes.Ok, "")
}

// AddSpanEvent adds an event to the span
func AddSpanEvent(span trace.Span, name string, attrs ...attribute.KeyValue) {
	if span == nil {
		return
	}
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// Deprecated span methods for backward compatibility

// RecordBuildSpan records a build operation span (deprecated: use StartBuildContextSpan)
func (p *Provider) RecordBuildSpan(ctx context.Context, functionName, namespace, runtime string) (context.Context, trace.Span) {
	return p.StartBuildContextSpan(ctx, functionName, namespace, runtime)
}

// RecordDeploySpan records a deploy operation span (deprecated: use StartDeployServiceSpan)
func (p *Provider) RecordDeploySpan(ctx context.Context, functionName, namespace string) (context.Context, trace.Span) {
	return p.StartDeployServiceSpan(ctx, functionName, namespace)
}

// RecordReconcileSpan records a reconciliation span (deprecated: use StartReconcilePhaseSpan)
func (p *Provider) RecordReconcileSpan(ctx context.Context, functionName, namespace, phase string) (context.Context, trace.Span) {
	return p.StartReconcilePhaseSpan(ctx, functionName, namespace, phase)
}

// SpanFromContext returns the current span from context
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// ContextWithSpan returns a context with the given span
func ContextWithSpan(ctx context.Context, span trace.Span) context.Context {
	return trace.ContextWithSpan(ctx, span)
}

// GetTraceID returns the trace ID from context
func GetTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// GetSpanID returns the span ID from context
func GetSpanID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}

// Global provider instance for easy access
var globalProvider *Provider

// SetGlobalProvider sets the global observability provider
func SetGlobalProvider(p *Provider) {
	globalProvider = p
}

// GetGlobalProvider returns the global observability provider
func GetGlobalProvider() *Provider {
	return globalProvider
}

// Tracer returns the global tracer, or a no-op tracer if not initialized
func Tracer() trace.Tracer {
	if globalProvider != nil && globalProvider.Tracer != nil {
		return globalProvider.Tracer
	}
	return otel.Tracer(ServiceName)
}

// StartSpanFromContext starts a span using the global tracer
func StartSpanFromContext(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return Tracer().Start(ctx, name, opts...)
}

// WithLambdaAttributes returns span options with lambda function attributes
func WithLambdaAttributes(functionName, namespace, runtime string) trace.SpanStartOption {
	return trace.WithAttributes(
		attribute.String("lambda.function", functionName),
		attribute.String("lambda.namespace", namespace),
		attribute.String("lambda.runtime", runtime),
	)
}

// WithBuildAttributes returns span options with build-related attributes
func WithBuildAttributes(jobName, imageURI string) trace.SpanStartOption {
	return trace.WithAttributes(
		attribute.String("build.job_name", jobName),
		attribute.String("build.image_uri", imageURI),
	)
}

// WithCloudEventAttributes returns span options with CloudEvent attributes
func WithCloudEventAttributes(eventType, eventSource, eventID string) trace.SpanStartOption {
	return trace.WithAttributes(
		attribute.String("cloudevents.type", eventType),
		attribute.String("cloudevents.source", eventSource),
		attribute.String("cloudevents.id", eventID),
	)
}
