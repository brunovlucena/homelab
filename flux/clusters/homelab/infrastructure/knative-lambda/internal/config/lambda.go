// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🚀 LAMBDA CONFIGURATION - Lambda function configuration
//
//	🎯 Purpose: Lambda function settings, runtime configuration, resource limits
//	💡 Features: Runtime settings, handler configuration, resource limits
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package config

import (
	"knative-lambda-new/internal/constants"
	"knative-lambda-new/internal/errors"
	"strconv"
)

// 🚀 LambdaConfig - "Lambda function configuration"
type LambdaConfig struct {
	// Runtime Configuration
	DefaultRuntime string `envconfig:"DEFAULT_RUNTIME" validate:"required"`
	DefaultHandler string `envconfig:"DEFAULT_HANDLER" validate:"required"`
	DefaultTrigger string `envconfig:"DEFAULT_TRIGGER" validate:"required"`

	// Resource Configuration
	FunctionMemoryLimit   string `envconfig:"FUNCTION_MEMORY_LIMIT" validate:"required"`
	FunctionCPULimit      string `envconfig:"FUNCTION_CPU_LIMIT" validate:"required"`
	FunctionMemoryRequest string `envconfig:"FUNCTION_MEMORY_REQUEST" validate:"required"`
	FunctionCPURequest    string `envconfig:"FUNCTION_CPU_REQUEST" validate:"required"`
	FunctionMemoryLimitMi string `envconfig:"FUNCTION_MEMORY_LIMIT_MI" validate:"required"`
	FunctionCPULimitM     string `envconfig:"FUNCTION_CPU_LIMIT_M" validate:"required"`
}

// 🚀 LambdaServicesConfig - "Configuration for dynamically created lambda services"
// 🎯 Purpose: Autoscaling and resource configuration for lambda services
type LambdaServicesConfig struct {
	// 🔧 AUTOSCALING: Configuration for lambda service autoscaling
	MinScale               string `envconfig:"LAMBDA_SERVICES_MIN_SCALE" validate:"required"`
	MaxScale               string `envconfig:"LAMBDA_SERVICES_MAX_SCALE" validate:"required"`
	TargetConcurrency      string `envconfig:"LAMBDA_SERVICES_TARGET_CONCURRENCY" validate:"required"`
	TargetUtilization      string `envconfig:"LAMBDA_SERVICES_TARGET_UTILIZATION" validate:"required"`
	Target                 string `envconfig:"LAMBDA_SERVICES_TARGET" validate:"required"`
	ContainerConcurrency   string `envconfig:"LAMBDA_SERVICES_CONTAINER_CONCURRENCY" validate:"required"`
	ScaleToZeroGracePeriod string `envconfig:"LAMBDA_SERVICES_SCALE_TO_ZERO_GRACE_PERIOD" validate:"required"`
	ScaleDownDelay         string `envconfig:"LAMBDA_SERVICES_SCALE_DOWN_DELAY" validate:"required"`
	StableWindow           string `envconfig:"LAMBDA_SERVICES_STABLE_WINDOW" validate:"required"`
	// 🚀 PANIC MODE: For rapid scaling during traffic spikes
	PanicWindowPercentage    string `envconfig:"LAMBDA_SERVICES_PANIC_WINDOW_PERCENTAGE" validate:"required"`
	PanicThresholdPercentage string `envconfig:"LAMBDA_SERVICES_PANIC_THRESHOLD_PERCENTAGE" validate:"required"`

	// 📦 RESOURCE CONFIGURATION - Specific to lambda services
	ResourceMemoryRequest string `envconfig:"LAMBDA_SERVICES_RESOURCE_MEMORY_REQUEST" validate:"required"`
	ResourceCPURequest    string `envconfig:"LAMBDA_SERVICES_RESOURCE_CPU_REQUEST" validate:"required"`
	ResourceMemoryLimit   string `envconfig:"LAMBDA_SERVICES_RESOURCE_MEMORY_LIMIT" validate:"required"`
	ResourceCPULimit      string `envconfig:"LAMBDA_SERVICES_RESOURCE_CPU_LIMIT" validate:"required"`

	// ⏰ TIMEOUT CONFIGURATION - Specific to lambda services
	TimeoutResponse string `envconfig:"LAMBDA_SERVICES_TIMEOUT_RESPONSE" validate:"required"`
	TimeoutIdle     string `envconfig:"LAMBDA_SERVICES_TIMEOUT_IDLE" validate:"required"`
}

// 🔧 NewLambdaConfig - "Create Lambda configuration with defaults"
func NewLambdaConfig() *LambdaConfig {
	return &LambdaConfig{}
}

// 🔧 NewLambdaServicesConfig - "Create Lambda services configuration with defaults"
func NewLambdaServicesConfig() *LambdaServicesConfig {
	return &LambdaServicesConfig{}
}

// 🔧 Validate - "Validate Lambda configuration"
func (c *LambdaConfig) Validate() error {
	if c.DefaultRuntime == "" {
		return errors.NewConfigurationError("lambda", "default_runtime", "default runtime cannot be empty")
	}
	if c.DefaultHandler == "" {
		return errors.NewConfigurationError("lambda", "default_handler", "default handler cannot be empty")
	}
	if c.DefaultTrigger == "" {
		return errors.NewConfigurationError("lambda", "default_trigger", "default trigger cannot be empty")
	}
	if c.FunctionMemoryLimit == "" {
		return errors.NewConfigurationError("lambda", "function_memory_limit", "function memory limit cannot be empty")
	}
	if c.FunctionCPULimit == "" {
		return errors.NewConfigurationError("lambda", "function_cpu_limit", "function CPU limit cannot be empty")
	}
	if c.FunctionMemoryRequest == "" {
		return errors.NewConfigurationError("lambda", "function_memory_request", "function memory request cannot be empty")
	}
	if c.FunctionCPURequest == "" {
		return errors.NewConfigurationError("lambda", "function_cpu_request", "function CPU request cannot be empty")
	}
	if c.FunctionMemoryLimitMi == "" {
		return errors.NewConfigurationError("lambda", "function_memory_limit_mi", "function memory limit mi cannot be empty")
	}
	if c.FunctionCPULimitM == "" {
		return errors.NewConfigurationError("lambda", "function_cpu_limit_m", "function CPU limit m cannot be empty")
	}
	return nil
}

// 🔧 getIntValue - "Generic helper to safely convert string to int with fallback"
//
// 📋 WHAT THIS DOES:
//
//	Safely converts a string value to integer with fallback.
//	Handles nil config, empty strings, and conversion errors.
//
// 🎯 WHY WE NEED THIS:
//   - DRY principle: Single function for all string-to-int conversions
//   - Handles all edge cases consistently
//   - Provides safe fallback values
func (c *LambdaServicesConfig) getIntValue(value string, fallback int) int {
	if c == nil {
		return fallback
	}

	if value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

// 🔧 GetContainerConcurrencyInt - "Get container concurrency as integer"
func (c *LambdaServicesConfig) GetContainerConcurrencyInt() int {
	// Convert string constant to int for fallback
	fallback, _ := strconv.Atoi(constants.LambdaServicesContainerConcurrencyDefault)
	return c.getIntValue(c.ContainerConcurrency, fallback)
}

// 🔧 Validate - "Validate Lambda services configuration"
func (c *LambdaServicesConfig) Validate() error {
	if c.MinScale == "" {
		return errors.NewConfigurationError("lambda_services", "min_scale", "min scale cannot be empty")
	}
	if c.MaxScale == "" {
		return errors.NewConfigurationError("lambda_services", "max_scale", "max scale cannot be empty")
	}
	if c.TargetConcurrency == "" {
		return errors.NewConfigurationError("lambda_services", "target_concurrency", "target concurrency cannot be empty")
	}
	if c.TargetUtilization == "" {
		return errors.NewConfigurationError("lambda_services", "target_utilization", "target utilization cannot be empty")
	}
	if c.Target == "" {
		return errors.NewConfigurationError("lambda_services", "target", "target cannot be empty")
	}
	if c.ContainerConcurrency == "" {
		return errors.NewConfigurationError("lambda_services", "container_concurrency", "container concurrency cannot be empty")
	}
	if c.ScaleToZeroGracePeriod == "" {
		return errors.NewConfigurationError("lambda_services", "scale_to_zero_grace_period", "scale to zero grace period cannot be empty")
	}
	if c.ScaleDownDelay == "" {
		return errors.NewConfigurationError("lambda_services", "scale_down_delay", "scale down delay cannot be empty")
	}
	if c.StableWindow == "" {
		return errors.NewConfigurationError("lambda_services", "stable_window", "stable window cannot be empty")
	}
	if c.ResourceMemoryRequest == "" {
		return errors.NewConfigurationError("lambda_services", "resource_memory_request", "resource memory request cannot be empty")
	}
	if c.ResourceCPURequest == "" {
		return errors.NewConfigurationError("lambda_services", "resource_cpu_request", "resource CPU request cannot be empty")
	}
	if c.ResourceMemoryLimit == "" {
		return errors.NewConfigurationError("lambda_services", "resource_memory_limit", "resource memory limit cannot be empty")
	}
	if c.ResourceCPULimit == "" {
		return errors.NewConfigurationError("lambda_services", "resource_cpu_limit", "resource CPU limit cannot be empty")
	}
	if c.TimeoutResponse == "" {
		return errors.NewConfigurationError("lambda_services", "timeout_response", "timeout response cannot be empty")
	}
	if c.TimeoutIdle == "" {
		return errors.NewConfigurationError("lambda_services", "timeout_idle", "timeout idle cannot be empty")
	}
	return nil
}
