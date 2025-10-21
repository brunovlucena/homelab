// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	⏱️ RATE LIMITING - Rate limiting configuration and management
//
//	🎯 Purpose: Configure rate limiting for different operation types
//	💡 Features: Multi-level rate limiting, burst control, memory monitoring
//
//	🏛️ ARCHITECTURE:
//	🔧 Rate Limiting Config - Configuration for different operation types
//	📊 Burst Control - Control burst sizes for different operations
//	💾 Memory Monitoring - Monitor memory usage and apply limits
//	🔄 Cleanup Management - Manage cleanup intervals and TTLs
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package config

import (
	"fmt"
	"time"

	"knative-lambda-new/internal/constants"
	"knative-lambda-new/internal/errors"
)

// RateLimitingConfig holds rate limiting and performance configuration
type RateLimitingConfig struct {
	// Rate Limiting Configuration
	Enabled                    bool `envconfig:"RATE_LIMITING_ENABLED" default:"true"`
	BuildContextRequestsPerMin int  `envconfig:"BUILD_CONTEXT_REQUESTS_PER_MIN" default:"5"`
	BuildContextBurstSize      int  `envconfig:"BUILD_CONTEXT_BURST_SIZE" default:"2"`
	K8sJobRequestsPerMin       int  `envconfig:"K8S_JOB_REQUESTS_PER_MIN" default:"10"`
	K8sJobBurstSize            int  `envconfig:"K8S_JOB_BURST_SIZE" default:"3"`
	ClientRequestsPerMin       int  `envconfig:"CLIENT_REQUESTS_PER_MIN" default:"5"`
	ClientBurstSize            int  `envconfig:"CLIENT_BURST_SIZE" default:"2"`
	S3UploadRequestsPerMin     int  `envconfig:"S3_UPLOAD_REQUESTS_PER_MIN" default:"50"`
	S3UploadBurstSize          int  `envconfig:"S3_UPLOAD_BURST_SIZE" default:"10"`

	// Memory Management
	MaxMemoryUsagePercent float64       `envconfig:"MAX_MEMORY_USAGE_PERCENT" default:"80.0"`
	MemoryCheckInterval   time.Duration `envconfig:"MEMORY_CHECK_INTERVAL" default:"30s"`

	// Cleanup Configuration
	CleanupInterval time.Duration `envconfig:"CLEANUP_INTERVAL" default:"5m"`
	ClientTTL       time.Duration `envconfig:"CLIENT_TTL" default:"1h"`

	// Performance Configuration
	MaxConcurrentBuilds int           `envconfig:"MAX_CONCURRENT_BUILDS" default:"10"`
	MaxConcurrentJobs   int           `envconfig:"MAX_CONCURRENT_JOBS" default:"5"`
	BuildTimeout        time.Duration `envconfig:"BUILD_TIMEOUT" default:"30m"`
	JobTimeout          time.Duration `envconfig:"JOB_TIMEOUT" default:"1h"`
	RequestTimeout      time.Duration `envconfig:"REQUEST_TIMEOUT" default:"5m"`
}

// NewRateLimitingConfig creates a new rate limiting and performance configuration with defaults
func NewRateLimitingConfig() *RateLimitingConfig {
	return &RateLimitingConfig{
		Enabled:                    true,
		BuildContextRequestsPerMin: constants.BuildContextRequestsPerMinDefault,
		BuildContextBurstSize:      constants.BuildContextBurstSizeDefault,
		K8sJobRequestsPerMin:       constants.K8sJobRequestsPerMinDefault,
		K8sJobBurstSize:            constants.K8sJobBurstSizeDefault,
		ClientRequestsPerMin:       constants.ClientRequestsPerMinDefault,
		ClientBurstSize:            constants.ClientBurstSizeDefault,
		S3UploadRequestsPerMin:     constants.S3UploadRequestsPerMinDefault,
		S3UploadBurstSize:          constants.S3UploadBurstSizeDefault,
		MaxMemoryUsagePercent:      constants.MaxMemoryUsagePercentDefault,
		MemoryCheckInterval:        constants.MemoryCheckIntervalDefault,
		CleanupInterval:            constants.CleanupIntervalDefault,
		ClientTTL:                  constants.ClientTTLDefault,
		MaxConcurrentBuilds:        constants.MaxConcurrentBuildsDefault,
		MaxConcurrentJobs:          constants.MaxConcurrentJobsDefault,
		BuildTimeout:               constants.RateLimitBuildTimeoutDefault,
		JobTimeout:                 constants.RateLimitJobTimeoutDefault,
		RequestTimeout:             constants.RateLimitRequestTimeoutDefault,
	}
}

// Validate validates the rate limiting configuration
func (c *RateLimitingConfig) Validate() error {
	if !c.Enabled {
		return nil // Skip validation if rate limiting is disabled
	}

	// Validate rate limiting configuration
	if c.BuildContextRequestsPerMin <= 0 {
		return errors.NewValidationError("build_context_requests_per_min", c.BuildContextRequestsPerMin, constants.ErrRateLimitPositive)
	}

	if c.BuildContextBurstSize <= 0 {
		return errors.NewValidationError("build_context_burst_size", c.BuildContextBurstSize, constants.ErrRateLimitPositive)
	}

	if c.K8sJobRequestsPerMin <= 0 {
		return errors.NewValidationError("k8s_job_requests_per_min", c.K8sJobRequestsPerMin, constants.ErrRateLimitPositive)
	}

	if c.K8sJobBurstSize <= 0 {
		return errors.NewValidationError("k8s_job_burst_size", c.K8sJobBurstSize, constants.ErrRateLimitPositive)
	}

	if c.ClientRequestsPerMin <= 0 {
		return errors.NewValidationError("client_requests_per_min", c.ClientRequestsPerMin, constants.ErrRateLimitPositive)
	}

	if c.ClientBurstSize <= 0 {
		return errors.NewValidationError("client_burst_size", c.ClientBurstSize, constants.ErrRateLimitPositive)
	}

	if c.S3UploadRequestsPerMin <= 0 {
		return errors.NewValidationError("s3_upload_requests_per_min", c.S3UploadRequestsPerMin, constants.ErrRateLimitPositive)
	}

	if c.S3UploadBurstSize <= 0 {
		return errors.NewValidationError("s3_upload_burst_size", c.S3UploadBurstSize, constants.ErrRateLimitPositive)
	}

	// Validate memory configuration
	if c.MaxMemoryUsagePercent <= 0 || c.MaxMemoryUsagePercent > 100 {
		return errors.NewValidationError("max_memory_usage_percent", c.MaxMemoryUsagePercent, constants.ErrMaxMemoryUsagePercentRange0To100)
	}

	if c.MemoryCheckInterval <= 0 {
		return errors.NewValidationError("memory_check_interval", c.MemoryCheckInterval, constants.ErrMemoryCheckIntervalPositive)
	}

	// Validate cleanup configuration
	if c.CleanupInterval <= 0 {
		return errors.NewValidationError("cleanup_interval", c.CleanupInterval, constants.ErrCleanupIntervalPositive)
	}

	if c.ClientTTL <= 0 {
		return errors.NewValidationError("client_ttl", c.ClientTTL, constants.ErrClientTTLPositive)
	}

	// Validate performance configuration
	if c.MaxConcurrentBuilds <= 0 {
		return errors.NewValidationError("max_concurrent_builds", c.MaxConcurrentBuilds, constants.ErrMaxConcurrentBuildsPositive)
	}

	if c.MaxConcurrentJobs <= 0 {
		return errors.NewValidationError("max_concurrent_jobs", c.MaxConcurrentJobs, constants.ErrMaxConcurrentJobsPositive)
	}

	if c.BuildTimeout <= 0 {
		return errors.NewValidationError("build_timeout", c.BuildTimeout, constants.ErrBuildTimeoutPositive)
	}

	if c.JobTimeout <= 0 {
		return errors.NewValidationError("job_timeout", c.JobTimeout, constants.ErrJobTimeoutPositive)
	}

	if c.RequestTimeout <= 0 {
		return errors.NewValidationError("request_timeout", c.RequestTimeout, constants.ErrRequestTimeoutPositive)
	}

	return nil
}

// ToMultiLevelRateLimiterConfig converts the rate limiting config to a map for the multi-level rate limiter
func (c *RateLimitingConfig) ToMultiLevelRateLimiterConfig() (map[string]interface{}, error) {
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("invalid rate limiting configuration: %w", err)
	}

	return map[string]interface{}{
		"BuildContextRequestsPerMin": c.BuildContextRequestsPerMin,
		"BuildContextBurstSize":      c.BuildContextBurstSize,
		"K8sJobRequestsPerMin":       c.K8sJobRequestsPerMin,
		"K8sJobBurstSize":            c.K8sJobBurstSize,
		"ClientRequestsPerMin":       c.ClientRequestsPerMin,
		"ClientBurstSize":            c.ClientBurstSize,
		"S3UploadRequestsPerMin":     c.S3UploadRequestsPerMin,
		"S3UploadBurstSize":          c.S3UploadBurstSize,
		"MaxMemoryUsagePercent":      c.MaxMemoryUsagePercent,
		"MemoryCheckInterval":        c.MemoryCheckInterval,
		"CleanupInterval":            c.CleanupInterval,
		"ClientTTL":                  c.ClientTTL,
		"MaxConcurrentBuilds":        c.MaxConcurrentBuilds,
		"MaxConcurrentJobs":          c.MaxConcurrentJobs,
		"BuildTimeout":               c.BuildTimeout,
		"JobTimeout":                 c.JobTimeout,
		"RequestTimeout":             c.RequestTimeout,
	}, nil
}
