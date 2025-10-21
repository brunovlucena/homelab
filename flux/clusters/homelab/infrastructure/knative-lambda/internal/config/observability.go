// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	📊 OBSERVABILITY CONFIGURATION - Logging, metrics, and tracing configuration
//
//	🎯 Purpose: Observability settings, logging levels, metrics collection, tracing
//	💡 Features: Log levels, metrics configuration, tracing settings, sampling
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package config

import (
	"knative-lambda-new/internal/constants"
	"knative-lambda-new/internal/errors"
	"knative-lambda-new/internal/observability"
	"strings"
)

// 📊 ObservabilityConfig - "Observability configuration"
type ObservabilityConfig struct {
	// Service Information
	ServiceName    string `envconfig:"SERVICE_NAME" default:"knative-lambda-new"`
	ServiceVersion string `envconfig:"SERVICE_VERSION" default:"1.0.0"`

	// Logging Configuration
	LogLevel string `envconfig:"LOG_LEVEL" default:"info" validate:"required,oneof=debug info warn error"`

	// Tracing Configuration
	OTLPEndpoint   string  `envconfig:"OTEL_EXPORTER_OTLP_ENDPOINT" default:"tempo-distributor.tempo.svc.cluster.local:4317"`
	TracingEnabled bool    `envconfig:"TRACING_ENABLED" default:"true"`
	SampleRate     float64 `envconfig:"SAMPLE_RATE" default:"1.0" validate:"min=0,max=1"`

	// Metrics Configuration
	MetricsEnabled bool `envconfig:"METRICS_ENABLED" default:"true"`

	// Exemplars Configuration
	ExemplarsEnabled       bool    `envconfig:"EXEMPLARS_ENABLED" default:"true"`
	ExemplarsMaxPerMetric  int     `envconfig:"EXEMPLARS_MAX_PER_METRIC" default:"10"`
	ExemplarsSampleRate    float64 `envconfig:"EXEMPLARS_SAMPLE_RATE" default:"0.1" validate:"min=0,max=1"`
	ExemplarsTraceIDLabel  string  `envconfig:"EXEMPLARS_TRACE_ID_LABEL" default:"trace_id"`
	ExemplarsSpanIDLabel   string  `envconfig:"EXEMPLARS_SPAN_ID_LABEL" default:"span_id"`
	ExemplarsIncludeLabels string  `envconfig:"EXEMPLARS_INCLUDE_LABELS" default:"third_party_id,parser_id,job_type,status,component,error_type"`
}

// 🔧 NewObservabilityConfig - "Create observability configuration with defaults"
func NewObservabilityConfig() *ObservabilityConfig {
	return &ObservabilityConfig{
		ServiceName:            constants.ServiceNameDefault,
		ServiceVersion:         constants.ServiceVersionDefault,
		LogLevel:               constants.LogLevelInfo,
		OTLPEndpoint:           constants.OTLPEndpointDefault,
		TracingEnabled:         constants.TracingEnabledDefault,
		SampleRate:             constants.SampleRateDefault,
		MetricsEnabled:         constants.MetricsEnabledDefault,
		ExemplarsEnabled:       constants.ExemplarsEnabledDefault,
		ExemplarsMaxPerMetric:  constants.ExemplarsMaxPerMetricDefault,
		ExemplarsSampleRate:    constants.ExemplarsSampleRateDefault,
		ExemplarsTraceIDLabel:  constants.ExemplarsTraceIDLabelDefault,
		ExemplarsSpanIDLabel:   constants.ExemplarsSpanIDLabelDefault,
		ExemplarsIncludeLabels: constants.ExemplarsIncludeLabelsDefault,
	}
}

// 🔧 Validate - "Validate observability configuration"
func (c *ObservabilityConfig) Validate() error {
	if !constants.IsValidLogLevel(c.LogLevel) {
		return errors.NewValidationError("log_level", c.LogLevel, constants.ErrLogLevelValid)
	}

	if c.SampleRate < 0 || c.SampleRate > 1 {
		return errors.NewValidationError("sample_rate", c.SampleRate, constants.ErrSampleRateRange0To1)
	}

	if c.ServiceName == "" {
		return errors.NewValidationError("service_name", c.ServiceName, constants.ErrServiceNameRequired)
	}

	if c.ServiceVersion == "" {
		return errors.NewValidationError("service_version", c.ServiceVersion, constants.ErrServiceVersionRequired)
	}

	if c.ExemplarsSampleRate < 0 || c.ExemplarsSampleRate > 1 {
		return errors.NewValidationError("exemplars_sample_rate", c.ExemplarsSampleRate, constants.ErrExemplarsSampleRateRange0To1)
	}

	if c.ExemplarsMaxPerMetric <= 0 {
		return errors.NewValidationError("exemplars_max_per_metric", c.ExemplarsMaxPerMetric, constants.ErrExemplarsMaxPerMetricPositive)
	}

	return nil
}

// 🔧 ToExemplarsConfig - "Convert to exemplars configuration"
func (c *ObservabilityConfig) ToExemplarsConfig() observability.ExemplarsConfig {
	includeLabels := strings.Split(c.ExemplarsIncludeLabels, ",")
	return observability.ExemplarsConfig{
		Enabled:               c.ExemplarsEnabled,
		MaxExemplarsPerMetric: c.ExemplarsMaxPerMetric,
		SampleRate:            c.ExemplarsSampleRate,
		IncludeLabels:         includeLabels,
		TraceIDLabel:          c.ExemplarsTraceIDLabel,
		SpanIDLabel:           c.ExemplarsSpanIDLabel,
	}
}

// 🔧 GetServiceName - "Get service name"
func (c *ObservabilityConfig) GetServiceName() string {
	return c.ServiceName
}

// 🔧 GetServiceVersion - "Get service version"
func (c *ObservabilityConfig) GetServiceVersion() string {
	return c.ServiceVersion
}

// 🔧 GetLogLevel - "Get log level"
func (c *ObservabilityConfig) GetLogLevel() string {
	return c.LogLevel
}

// 🔧 GetOTLPEndpoint - "Get OTLP endpoint"
func (c *ObservabilityConfig) GetOTLPEndpoint() string {
	return c.OTLPEndpoint
}

// 🔧 IsTracingEnabled - "Check if tracing is enabled"
func (c *ObservabilityConfig) IsTracingEnabled() bool {
	return c.TracingEnabled
}

// 🔧 IsMetricsEnabled - "Check if metrics are enabled"
func (c *ObservabilityConfig) IsMetricsEnabled() bool {
	return c.MetricsEnabled
}

// 🔧 GetSampleRate - "Get sample rate"
func (c *ObservabilityConfig) GetSampleRate() float64 {
	return c.SampleRate
}
