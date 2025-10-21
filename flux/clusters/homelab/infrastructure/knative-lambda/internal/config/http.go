// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🌐 HTTP CONFIGURATION - HTTP server and API configuration
//
//	🎯 Purpose: HTTP server settings, timeouts, middleware configuration
//	💡 Features: Port configuration, timeout settings, request limits
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package config

import (
	"strconv"
	"time"

	"knative-lambda-new/internal/constants"
	"knative-lambda-new/internal/errors"
)

// 🌐 HTTPConfig - "HTTP server and API configuration"
type HTTPConfig struct {
	// Server Configuration
	Port        int `envconfig:"PORT" default:"8080" validate:"required,min=1024,max=65535"`
	MetricsPort int `envconfig:"METRICS_PORT" default:"8080" validate:"required,min=1024,max=65535"`

	// Timeout Configuration
	Timeout    time.Duration `envconfig:"TIMEOUT" default:"30s" validate:"required,min=1s"`
	APITimeout time.Duration `envconfig:"API_TIMEOUT" default:"400ms" validate:"required,min=50ms,max=1s"`

	// Request Configuration
	MaxRequestSize int64 `envconfig:"MAX_REQUEST_SIZE" default:"10485760"` // 10MB

	// List Configuration
	DefaultListLimit int `envconfig:"DEFAULT_LIST_LIMIT" default:"50" validate:"required,min=1,max=1000"`
	MaxListLimit     int `envconfig:"MAX_LIST_LIMIT" default:"100" validate:"required,min=1,max=1000"`

	// Security Configuration

	ValidateInput bool // Default: true
}

// 🔧 NewHTTPConfig - "Create HTTP configuration with defaults"
func NewHTTPConfig() *HTTPConfig {
	return &HTTPConfig{
		Port:             constants.PortDefault,
		MetricsPort:      constants.MetricsPortDefault,
		Timeout:          constants.RequestTimeoutDefault,
		APITimeout:       constants.APITimeoutDefault,
		MaxRequestSize:   constants.MaxRequestSizeDefault,
		DefaultListLimit: constants.DefaultListLimit,
		MaxListLimit:     constants.MaxListLimit,
		ValidateInput:    true,
	}
}

// 🔧 Validate - "Validate HTTP configuration"
func (c *HTTPConfig) Validate() error {
	if !constants.IsValidPort(c.Port) {
		return errors.NewValidationError("port", c.Port, constants.ErrPortRange1024To65535)
	}

	if !constants.IsValidPort(c.MetricsPort) {
		return errors.NewValidationError("metrics_port", c.MetricsPort, constants.ErrMetricsPortRange1024To65535)
	}

	if c.Timeout < time.Second {
		return errors.NewValidationError("timeout", c.Timeout, constants.ErrTimeoutMin1Second)
	}

	if c.APITimeout < 50*time.Millisecond || c.APITimeout > time.Second {
		return errors.NewValidationError("api_timeout", c.APITimeout, constants.ErrAPITimeoutRange50msTo1s)
	}

	if c.MaxRequestSize <= 0 {
		return errors.NewValidationError("max_request_size", c.MaxRequestSize, constants.ErrMaxRequestSizePositive)
	}

	if c.DefaultListLimit < 1 || c.DefaultListLimit > 1000 {
		return errors.NewValidationError("default_list_limit", c.DefaultListLimit, constants.ErrDefaultListLimitRange1To1000)
	}

	if c.MaxListLimit < 1 || c.MaxListLimit > 1000 {
		return errors.NewValidationError("max_list_limit", c.MaxListLimit, constants.ErrMaxListLimitRange1To1000)
	}

	if c.DefaultListLimit > c.MaxListLimit {
		return errors.NewValidationError("default_list_limit", c.DefaultListLimit, constants.ErrDefaultListLimitGreaterThanMax)
	}

	return nil
}

// 🔧 GetServerAddress - "Get HTTP server address"
func (c *HTTPConfig) GetServerAddress() string {
	return ":" + strconv.Itoa(c.Port)
}

// 🔧 GetMetricsAddress - "Get metrics server address"
func (c *HTTPConfig) GetMetricsAddress() string {
	return ":" + strconv.Itoa(c.MetricsPort)
}
