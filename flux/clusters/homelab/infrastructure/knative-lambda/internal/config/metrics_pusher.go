// Package config provides centralized configuration management for the Knative Lambda service.
//
// This file contains the MetricsPusher configuration for lambda services.
package config

import (
	"fmt"
	"strconv"
	"time"

	"knative-lambda-new/internal/constants"
	"knative-lambda-new/internal/errors"
)

// MetricsPusherConfig represents the configuration for the metrics-pusher sidecar
type MetricsPusherConfig struct {
	// Enabled determines if metrics-pusher is enabled for lambda services
	Enabled bool `envconfig:"METRICS_PUSHER_ENABLED" default:"true"`

	// Image configuration
	ImageRegistry   string `envconfig:"METRICS_PUSHER_IMAGE_REGISTRY" default:"339954290315.dkr.ecr.us-west-2.amazonaws.com"`
	ImageRepository string `envconfig:"METRICS_PUSHER_IMAGE_REPOSITORY" default:"knative-lambdas/knative-lambda-metrics-pusher"`
	ImageTag        string `envconfig:"METRICS_PUSHER_IMAGE_TAG" default:"latest"`
	ImagePullPolicy string `envconfig:"METRICS_PUSHER_IMAGE_PULL_POLICY" default:"Always"`

	// Remote write configuration
	RemoteWriteURL string `envconfig:"METRICS_PUSHER_REMOTE_WRITE_URL" default:"http://prometheus-kube-prometheus-prometheus.prometheus:9090/api/v1/write"`

	// Timing configuration
	PushInterval time.Duration `envconfig:"METRICS_PUSHER_PUSH_INTERVAL" default:"30s"`
	Timeout      time.Duration `envconfig:"METRICS_PUSHER_TIMEOUT" default:"10s"`

	// Logging configuration
	LogLevel  string `envconfig:"METRICS_PUSHER_LOG_LEVEL" default:"info"`
	LogFormat string `envconfig:"METRICS_PUSHER_LOG_FORMAT" default:"json"`

	// Metrics configuration
	QueueProxyMetricsPort string `envconfig:"METRICS_PUSHER_QUEUE_PROXY_PORT" default:"9091"`
	QueueProxyMetricsPath string `envconfig:"METRICS_PUSHER_QUEUE_PROXY_PATH" default:"/metrics"`
	BuilderMetricsPort    string `envconfig:"METRICS_PUSHER_BUILDER_PORT" default:"8081"`
	BuilderMetricsPath    string `envconfig:"METRICS_PUSHER_BUILDER_PATH" default:"/metrics"`

	// Failure tolerance configuration
	FailureTolerance int `envconfig:"METRICS_PUSHER_FAILURE_TOLERANCE" default:"5"`

	// Resource configuration
	ResourceCPURequest    string `envconfig:"METRICS_PUSHER_CPU_REQUEST" default:"50m"`
	ResourceMemoryRequest string `envconfig:"METRICS_PUSHER_MEMORY_REQUEST" default:"64Mi"`
	ResourceCPULimit      string `envconfig:"METRICS_PUSHER_CPU_LIMIT" default:"100m"`
	ResourceMemoryLimit   string `envconfig:"METRICS_PUSHER_MEMORY_LIMIT" default:"128Mi"`
}

// NewMetricsPusherConfig creates a new MetricsPusherConfig with default values
func NewMetricsPusherConfig() *MetricsPusherConfig {
	return &MetricsPusherConfig{
		Enabled:               constants.MetricsPusherEnabledDefault,
		ImageRegistry:         constants.MetricsPusherImageRegistryDefault,
		ImageRepository:       constants.MetricsPusherImageRepositoryDefault,
		ImageTag:              constants.MetricsPusherImageTagDefault,
		ImagePullPolicy:       constants.MetricsPusherImagePullPolicyDefault,
		RemoteWriteURL:        constants.MetricsPusherRemoteWriteURLDefault,
		PushInterval:          constants.MetricsPusherPushIntervalDefault,
		Timeout:               constants.MetricsPusherTimeoutDefault,
		LogLevel:              constants.MetricsPusherLogLevelDefault,
		LogFormat:             constants.MetricsPusherLogFormatDefault,
		QueueProxyMetricsPort: constants.MetricsPusherQueueProxyPortDefault,
		QueueProxyMetricsPath: constants.MetricsPusherQueueProxyPathDefault,
		BuilderMetricsPort:    constants.MetricsPusherBuilderMetricsPortDefault,
		BuilderMetricsPath:    constants.MetricsPusherBuilderPathDefault,
		FailureTolerance:      constants.MetricsPusherFailureToleranceDefault,
		ResourceCPURequest:    constants.MetricsPusherCPURequestDefault,
		ResourceMemoryRequest: constants.MetricsPusherMemoryRequestDefault,
		ResourceCPULimit:      constants.MetricsPusherCPULimitDefault,
		ResourceMemoryLimit:   constants.MetricsPusherMemoryLimitDefault,
	}
}

// GetImageURL returns the full image URL for the metrics-pusher
func (c *MetricsPusherConfig) GetImageURL() string {
	return fmt.Sprintf("%s/%s:%s", c.ImageRegistry, c.ImageRepository, c.ImageTag)
}

// Validate validates the MetricsPusherConfig
func (c *MetricsPusherConfig) Validate() error {
	if c.RemoteWriteURL == "" {
		return errors.NewConfigurationError("metrics_pusher", "remote_write_url", "remote write URL is required")
	}

	if c.PushInterval <= 0 {
		return errors.NewConfigurationError("metrics_pusher", "push_interval", "push interval must be positive")
	}

	if c.Timeout <= 0 {
		return errors.NewConfigurationError("metrics_pusher", "timeout", "timeout must be positive")
	}

	if c.FailureTolerance < 0 {
		return errors.NewConfigurationError("metrics_pusher", "failure_tolerance", "failure tolerance must be non-negative")
	}

	// Validate resource requests and limits
	if err := validateResourceString("cpu_request", c.ResourceCPURequest); err != nil {
		return err
	}
	if err := validateResourceString("memory_request", c.ResourceMemoryRequest); err != nil {
		return err
	}
	if err := validateResourceString("cpu_limit", c.ResourceCPULimit); err != nil {
		return err
	}
	if err := validateResourceString("memory_limit", c.ResourceMemoryLimit); err != nil {
		return err
	}

	return nil
}

// validateResourceString validates a Kubernetes resource string
func validateResourceString(fieldName, value string) error {
	if value == "" {
		return errors.NewConfigurationError("metrics_pusher", fieldName, fmt.Sprintf("%s cannot be empty", fieldName))
	}
	return nil
}

// GetFailureToleranceString returns the failure tolerance as a string
func (c *MetricsPusherConfig) GetFailureToleranceString() string {
	return strconv.Itoa(c.FailureTolerance)
}

// GetPushIntervalString returns the push interval as a string
func (c *MetricsPusherConfig) GetPushIntervalString() string {
	return c.PushInterval.String()
}

// GetTimeoutString returns the timeout as a string
func (c *MetricsPusherConfig) GetTimeoutString() string {
	return c.Timeout.String()
}
