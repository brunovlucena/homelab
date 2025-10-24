// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🏗️ BUILD CONFIGURATION - Build process and resource configuration
//
//	🎯 Purpose: Build process settings, Kaniko configuration, resource limits
//	💡 Features: Kaniko settings, timeouts, resource limits, image configuration
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package config

import (
	"time"

	"knative-lambda-new/internal/constants"
	"knative-lambda-new/internal/errors"
)

// 🏗️ BuildConfig - "Build process configuration"
type BuildConfig struct {
	// Kaniko Configuration
	KanikoImage  string `envconfig:"KANIKO_IMAGE" default:"339954290315.dkr.ecr.us-west-2.amazonaws.com/knative-lambda-kaniko-executor:v1.24.0" validate:"required"`
	SidecarImage string `envconfig:"SIDECAR_IMAGE" validate:"required"`

	// Timeout Configuration
	BuildTimeout time.Duration `envconfig:"BUILD_TIMEOUT" default:"30m" validate:"required,min=5m"`

	// Resource Configuration
	CPURequest    string `envconfig:"CPU_REQUEST" default:"500m" validate:"required"`
	CPULimit      string `envconfig:"CPU_LIMIT" default:"2000m" validate:"required"`
	MemoryRequest string `envconfig:"MEMORY_REQUEST" default:"512Mi" validate:"required"`
	MemoryLimit   string `envconfig:"MEMORY_LIMIT" default:"2Gi" validate:"required"`

	// Size Limits
	MaxParserSize int64 `envconfig:"MAX_PARSER_SIZE" default:"52428800"` // 50MB
}

// 🔧 NewBuildConfig - "Create build configuration with defaults"
func NewBuildConfig() *BuildConfig {
	return &BuildConfig{
		KanikoImage:   constants.KanikoImageDefault,
		SidecarImage:  "",
		BuildTimeout:  constants.BuildTimeoutDefault,
		CPURequest:    constants.CPURequestDefault,
		CPULimit:      constants.CPULimitDefault,
		MemoryRequest: constants.MemoryRequestDefault,
		MemoryLimit:   constants.MemoryLimitDefault,
		MaxParserSize: constants.K8sMaxParserSizeDefault,
	}
}

// 🔧 Validate - "Validate build configuration"
func (c *BuildConfig) Validate() error {
	if c.KanikoImage == "" {
		return errors.NewValidationError("kaniko_image", c.KanikoImage, constants.ErrKanikoImageRequired)
	}

	if c.SidecarImage == "" {
		return errors.NewValidationError("sidecar_image", c.SidecarImage, constants.ErrSidecarImageRequired)
	}

	if c.BuildTimeout < 5*time.Minute {
		return errors.NewValidationError("build_timeout", c.BuildTimeout, constants.ErrBuildTimeoutMin5Minutes)
	}

	if c.CPURequest == "" {
		return errors.NewValidationError("cpu_request", c.CPURequest, constants.ErrCPURequestRequired)
	}

	if c.CPULimit == "" {
		return errors.NewValidationError("cpu_limit", c.CPULimit, constants.ErrCPULimitRequired)
	}

	if c.MemoryRequest == "" {
		return errors.NewValidationError("memory_request", c.MemoryRequest, constants.ErrMemoryRequestRequired)
	}

	if c.MemoryLimit == "" {
		return errors.NewValidationError("memory_limit", c.MemoryLimit, constants.ErrMemoryLimitRequired)
	}

	if c.MaxParserSize <= 0 {
		return errors.NewValidationError("max_parser_size", c.MaxParserSize, constants.ErrMaxParserSizePositive)
	}

	return nil
}

// 🔧 GetKanikoImage - "Get Kaniko image"
func (c *BuildConfig) GetKanikoImage() string {
	return c.KanikoImage
}

// 🔧 GetSidecarImage - "Get sidecar image"
func (c *BuildConfig) GetSidecarImage() string {
	return c.SidecarImage
}

// 🔧 GetBuildTimeout - "Get build timeout"
func (c *BuildConfig) GetBuildTimeout() time.Duration {
	return c.BuildTimeout
}

// 🔧 GetCPURequest - "Get CPU request"
func (c *BuildConfig) GetCPURequest() string {
	return c.CPURequest
}

// 🔧 GetCPULimit - "Get CPU limit"
func (c *BuildConfig) GetCPULimit() string {
	return c.CPULimit
}

// 🔧 GetMemoryRequest - "Get memory request"
func (c *BuildConfig) GetMemoryRequest() string {
	return c.MemoryRequest
}

// 🔧 GetMemoryLimit - "Get memory limit"
func (c *BuildConfig) GetMemoryLimit() string {
	return c.MemoryLimit
}

// 🔧 GetMaxParserSize - "Get maximum parser size"
func (c *BuildConfig) GetMaxParserSize() int64 {
	return c.MaxParserSize
}
