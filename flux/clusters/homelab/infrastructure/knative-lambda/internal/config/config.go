// Package config provides centralized configuration management for the Knative Lambda service.
//
// The package handles environment variable loading, validation, defaults, and structured
// configuration for all service components including HTTP server, Kubernetes, AWS services,
// RabbitMQ, observability, security, and build process settings.
//
// Configuration is loaded from environment variables following 12-factor app principles
// with secure defaults and comprehensive validation.
package config

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"k8s.io/client-go/rest"

	"knative-lambda-new/internal/constants"
	"knative-lambda-new/internal/errors"
)

// Config represents the complete configuration for the Knative Lambda service.
//
// The configuration is organized into logical sections for different service components:
// - HTTP: Server settings, ports, timeouts
// - Kubernetes: Namespace, client configuration, RBAC
// - AWS: ECR registry, S3 buckets, IAM settings
// - RabbitMQ: Connection strings, exchanges, queues
// - Observability: Metrics, tracing, logging
// - Security: Authentication, authorization, validation
// - Build: Kaniko settings, timeouts, resource limits
// - Lambda: Function-specific settings
// - LambdaServices: Configuration for dynamically created lambda services
// - Knative: Eventing configuration
// - MetricsPusher: Configuration for metrics-pusher sidecar
// - Features: Feature flags and toggles
// - Performance: Performance tuning parameters
type Config struct {
	// Environment specifies the deployment environment (dev, prd, local)
	Environment string `envconfig:"ENVIRONMENT" default:"dev" validate:"required,oneof=dev prd local"`

	// Component configurations
	HTTP       *HTTPConfig       `json:"http"`
	Kubernetes *KubernetesConfig `json:"kubernetes"`
	AWS        *AWSConfig        `json:"aws"`

	Observability  *ObservabilityConfig  `json:"observability"`
	Build          *BuildConfig          `json:"build"`
	Lambda         *LambdaConfig         `json:"lambda"`
	LambdaServices *LambdaServicesConfig `json:"lambda_services"`
	Knative        *KnativeConfig        `json:"knative"`
	Security       *SecurityConfig       `json:"security"`
	RateLimiting   *RateLimitingConfig   `json:"rate_limiting"`
	Notifi         *NotifiConfig         `json:"notifi"`
	MetricsPusher  *MetricsPusherConfig  `json:"metrics_pusher"`
}

// ReloadFromEnvironment reloads the configuration from environment variables at runtime.
// This allows for dynamic configuration updates without service restart.
func (c *Config) ReloadFromEnvironment() error {
	builder := NewConfigBuilder().WithEnvironment(c.Environment)
	builder.LoadFromEnvironment().Validate()
	if builder.err != nil {
		return builder.err
	}
	// Overwrite all fields with new configuration
	*c = *builder.config
	return nil
}

// LoadConfig loads and validates the service configuration using the ConfigBuilder pattern.
// Returns a fully validated Config struct or an error if validation fails.
func LoadConfig() (*Config, error) {
	cfg, err := NewConfigBuilder().
		WithEnvironment(getEnv("ENVIRONMENT", constants.EnvironmentDev)).
		LoadFromEnvironment().
		Validate().
		Build()

	if err != nil {
		return nil, errors.NewConfigurationError("config", "validation", fmt.Sprintf("configuration validation failed: %v", err))
	}

	return cfg, nil
}

// Validate performs comprehensive validation of all configuration components.
// Returns an error if any component fails validation with detailed error context.
func (c *Config) Validate() error {
	// Validate component configurations
	if err := c.HTTP.Validate(); err != nil {
		return errors.NewConfigurationError("http", "validation", fmt.Sprintf("HTTP config validation failed: %v", err))
	}

	if err := c.Kubernetes.Validate(); err != nil {
		return errors.NewConfigurationError("kubernetes", "validation", fmt.Sprintf("Kubernetes config validation failed: %v", err))
	}

	if err := c.AWS.Validate(); err != nil {
		return errors.NewConfigurationError("aws", "validation", fmt.Sprintf("AWS config validation failed: %v", err))
	}

	if err := c.Observability.Validate(); err != nil {
		return errors.NewConfigurationError("observability", "validation", fmt.Sprintf("Observability config validation failed: %v", err))
	}

	if err := c.Build.Validate(); err != nil {
		return errors.NewConfigurationError("build", "validation", fmt.Sprintf("Build config validation failed: %v", err))
	}

	if err := c.Lambda.Validate(); err != nil {
		return errors.NewConfigurationError("lambda", "validation", fmt.Sprintf("Lambda config validation failed: %v", err))
	}

	if err := c.Knative.Validate(); err != nil {
		return errors.NewConfigurationError("knative", "validation", fmt.Sprintf("Knative config validation failed: %v", err))
	}

	if err := c.Security.Validate(); err != nil {
		return errors.NewConfigurationError("security", "validation", fmt.Sprintf("Security config validation failed: %v", err))
	}

	if err := c.RateLimiting.Validate(); err != nil {
		return errors.NewConfigurationError("rate_limiting", "validation", fmt.Sprintf("Rate limiting config validation failed: %v", err))
	}

	if err := c.Notifi.Validate(); err != nil {
		return errors.NewConfigurationError("notifi", "validation", fmt.Sprintf("Notifi config validation failed: %v", err))
	}

	return nil
}

// GetKubernetesConfig returns the Kubernetes client configuration.
func (c *Config) GetKubernetesConfig() (*rest.Config, error) {
	return c.Kubernetes.GetKubernetesConfig()
}

// GetAWSConfig returns the AWS SDK configuration for the given context.
func (c *Config) GetAWSConfig(ctx context.Context) (aws.Config, error) {
	return c.AWS.GetAWSConfig(ctx)
}
