// Package metrics provides Prometheus metrics for the knative-lambda operator.
// These metrics are critical for monitoring operator health at scale (10M+ lambdas).
// Supports exemplars for linking metrics to traces (Prometheus → Tempo).
package metrics

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	namespace = "knative_lambda"
	subsystem = "operator"
)

var (
	// ReconcileTotal counts total reconciliations by result
	ReconcileTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "reconcile_total",
			Help:      "Total number of reconciliations by phase and result",
		},
		[]string{"phase", "result"},
	)

	// ReconcileDuration measures reconcile latency
	ReconcileDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "reconcile_duration_seconds",
			Help:      "Duration of reconciliations in seconds",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30, 60},
		},
		[]string{"phase"},
	)

	// LambdaFunctionsTotal tracks total lambda count by phase
	LambdaFunctionsTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "lambdafunctions_total",
			Help:      "Current number of LambdaFunction resources by phase",
		},
		[]string{"namespace", "phase"},
	)

	// BuildJobsActive tracks active build jobs
	BuildJobsActive = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "build_jobs_active",
			Help:      "Number of active Kaniko build jobs",
		},
		[]string{"namespace"},
	)

	// BuildDuration measures build time
	BuildDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "build_duration_seconds",
			Help:      "Duration of image builds in seconds",
			Buckets:   []float64{10, 30, 60, 120, 300, 600, 900, 1200, 1800},
		},
		[]string{"runtime", "result"},
	)

	// EventingResourcesTotal tracks eventing resources
	EventingResourcesTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "eventing_resources_total",
			Help:      "Number of eventing resources (brokers, triggers) by type",
		},
		[]string{"namespace", "resource_type"},
	)

	// APIServerRequestsTotal counts API server interactions
	APIServerRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "apiserver_requests_total",
			Help:      "Total API server requests by verb and resource",
		},
		[]string{"verb", "resource", "result"},
	)

	// WorkQueueDepth tracks work queue depth
	WorkQueueDepth = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "workqueue_depth",
			Help:      "Current depth of the work queue",
		},
	)

	// WorkQueueLatency tracks time items spend in queue
	WorkQueueLatency = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "workqueue_latency_seconds",
			Help:      "Time items spend in the work queue",
			Buckets:   []float64{.001, .01, .1, 1, 10, 60, 300},
		},
	)

	// ErrorsTotal counts errors by type
	ErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "errors_total",
			Help:      "Total errors by type and component",
		},
		[]string{"component", "error_type"},
	)

	// BuildEventsTotal counts build lifecycle events
	BuildEventsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "build_events_total",
			Help:      "Total number of build lifecycle events by status",
		},
		[]string{"status"},
	)

	// ServiceEventsTotal counts service lifecycle events
	ServiceEventsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "service_events_total",
			Help:      "Total number of service lifecycle events by status",
		},
		[]string{"status"},
	)

	// ParserEventsTotal counts parser/invoke events
	ParserEventsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "parser_events_total",
			Help:      "Total number of parser/invoke events by status",
		},
		[]string{"status"},
	)
)

// Register registers all metrics with the controller-runtime metrics registry
func Register() {
	metrics.Registry.MustRegister(
		ReconcileTotal,
		ReconcileDuration,
		LambdaFunctionsTotal,
		BuildJobsActive,
		BuildDuration,
		EventingResourcesTotal,
		APIServerRequestsTotal,
		WorkQueueDepth,
		WorkQueueLatency,
		ErrorsTotal,
		BuildEventsTotal,
		ServiceEventsTotal,
		ParserEventsTotal,
	)
}

// ReconcilerMetrics wraps metrics for use in the reconciler
// Supports exemplars for linking metrics to traces
type ReconcilerMetrics struct {
	// ExemplarsEnabled enables trace exemplars on histogram metrics
	ExemplarsEnabled bool
}

// NewReconcilerMetrics creates a new ReconcilerMetrics
func NewReconcilerMetrics() *ReconcilerMetrics {
	return &ReconcilerMetrics{
		ExemplarsEnabled: true,
	}
}

// NewReconcilerMetricsWithExemplars creates ReconcilerMetrics with exemplar support
func NewReconcilerMetricsWithExemplars(enabled bool) *ReconcilerMetrics {
	return &ReconcilerMetrics{
		ExemplarsEnabled: enabled,
	}
}

// RecordReconcile records a reconciliation
func (m *ReconcilerMetrics) RecordReconcile(phase, result string, durationSeconds float64) {
	ReconcileTotal.WithLabelValues(phase, result).Inc()
	ReconcileDuration.WithLabelValues(phase).Observe(durationSeconds)
}

// RecordReconcileWithContext records a reconciliation with trace exemplar
func (m *ReconcilerMetrics) RecordReconcileWithContext(ctx context.Context, phase, result string, durationSeconds float64) {
	ReconcileTotal.WithLabelValues(phase, result).Inc()

	if m.ExemplarsEnabled {
		exemplar := m.extractExemplar(ctx)
		if exemplar != nil {
			// Use ExemplarObserver for histogram with exemplar
			ReconcileDuration.WithLabelValues(phase).(prometheus.ExemplarObserver).ObserveWithExemplar(durationSeconds, exemplar)
			return
		}
	}
	ReconcileDuration.WithLabelValues(phase).Observe(durationSeconds)
}

// RecordBuild records a build completion
func (m *ReconcilerMetrics) RecordBuild(runtime, result string, durationSeconds float64) {
	BuildDuration.WithLabelValues(runtime, result).Observe(durationSeconds)
}

// RecordBuildWithContext records a build completion with trace exemplar
func (m *ReconcilerMetrics) RecordBuildWithContext(ctx context.Context, runtime, result string, durationSeconds float64) {
	if m.ExemplarsEnabled {
		exemplar := m.extractExemplar(ctx)
		if exemplar != nil {
			BuildDuration.WithLabelValues(runtime, result).(prometheus.ExemplarObserver).ObserveWithExemplar(durationSeconds, exemplar)
			return
		}
	}
	BuildDuration.WithLabelValues(runtime, result).Observe(durationSeconds)
}

// RecordError records an error
func (m *ReconcilerMetrics) RecordError(component, errorType string) {
	ErrorsTotal.WithLabelValues(component, errorType).Inc()
}

// RecordErrorWithContext records an error with trace exemplar
func (m *ReconcilerMetrics) RecordErrorWithContext(ctx context.Context, component, errorType string) {
	if m.ExemplarsEnabled {
		exemplar := m.extractExemplar(ctx)
		if exemplar != nil {
			ErrorsTotal.WithLabelValues(component, errorType).(prometheus.ExemplarAdder).AddWithExemplar(1, exemplar)
			return
		}
	}
	ErrorsTotal.WithLabelValues(component, errorType).Inc()
}

// SetLambdaCount sets the lambda count for a namespace/phase
func (m *ReconcilerMetrics) SetLambdaCount(namespace, phase string, count float64) {
	LambdaFunctionsTotal.WithLabelValues(namespace, phase).Set(count)
}

// SetActiveBuildJobs sets the active build job count
func (m *ReconcilerMetrics) SetActiveBuildJobs(namespace string, count float64) {
	BuildJobsActive.WithLabelValues(namespace).Set(count)
}

// extractExemplar extracts trace_id and span_id from context for exemplars
func (m *ReconcilerMetrics) extractExemplar(ctx context.Context) prometheus.Labels {
	span := trace.SpanFromContext(ctx)
	if span == nil || !span.SpanContext().IsValid() {
		return nil
	}

	return prometheus.Labels{
		"trace_id": span.SpanContext().TraceID().String(),
		"span_id":  span.SpanContext().SpanID().String(),
	}
}

// RED Metrics (Rate, Errors, Duration) for Lambda Functions
var (
	// FunctionInvocationsTotal counts function invocations (Rate)
	FunctionInvocationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "function",
			Name:      "invocations_total",
			Help:      "Total number of function invocations by status",
		},
		[]string{"function", "namespace", "status"},
	)

	// FunctionDuration measures function execution duration (Duration)
	FunctionDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "function",
			Name:      "duration_seconds",
			Help:      "Duration of function invocations in seconds with exemplars",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30},
		},
		[]string{"function", "namespace"},
	)

	// FunctionErrorsTotal counts function errors by type (Errors)
	FunctionErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "function",
			Name:      "errors_total",
			Help:      "Total number of function errors by type",
		},
		[]string{"function", "namespace", "error_type"},
	)

	// FunctionColdStartsTotal counts cold starts
	FunctionColdStartsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "function",
			Name:      "cold_starts_total",
			Help:      "Total number of function cold starts",
		},
		[]string{"function", "namespace"},
	)
)

func init() {
	// Register RED metrics for functions
	metrics.Registry.MustRegister(
		FunctionInvocationsTotal,
		FunctionDuration,
		FunctionErrorsTotal,
		FunctionColdStartsTotal,
	)
}
