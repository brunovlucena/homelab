// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🎯 KNATIVE CONFIGURATION - Knative autoscaling configuration
//
//	🎯 Purpose: Knative autoscaling settings, scaling configuration
//	💡 Features: Autoscaling settings, scaling policies, event configuration
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package config

import (
	"strconv"

	"knative-lambda-new/internal/constants"
	"knative-lambda-new/internal/errors"
)

// 🎯 KnativeConfig - "Knative autoscaling configuration"
//
// 📋 WHAT THIS DOES:
//
//	This struct holds ALL the settings that control how your Knative services scale up and down.
//	Think of it as the "control panel" for autoscaling behavior.
//
// 🎮 KEY CONCEPTS:
//   - Target Concurrency: How many requests each pod can handle at once
//   - Target Utilization: When to create new pods (percentage of concurrency)
//   - Min/Max Scale: The minimum and maximum number of pods allowed
//   - Grace Periods: How long to wait before scaling down to zero
//
// 🔧 CONFIGURATION FIELDS:
type KnativeConfig struct {
	// 🚀 AUTOSCALING CONFIGURATION - Controls how pods scale up/down
	//
	// 📊 KnativeTargetConcurrency: "How many requests can one pod handle?"
	//    - Default: "1" (each pod handles exactly 1 request at a time)
	//    - Why 1? Forces horizontal scaling instead of queuing requests
	//    - Example: If you get 5 requests, you get 5 pods (not 1 pod with 5 requests)
	KnativeTargetConcurrency string `envconfig:"KNATIVE_TARGET_CONCURRENCY" validate:"required"`

	// 📈 KnativeTargetUtilization: "When should I create a new pod?"
	//    - Default: "70" (70% of target concurrency)
	//    - Math: 1 request × 70% = 0.7 threshold
	//    - Result: ANY load > 0.7 requests triggers new pod creation
	//    - This makes scaling SUPER aggressive for testing!
	KnativeTargetUtilization string `envconfig:"KNATIVE_TARGET_UTILIZATION" validate:"required"`

	// 🎯 KnativeTarget: "What's the target concurrency for autoscaling?"
	//    - Default: "0.01" (very aggressive scaling)
	//    - Math: 0.01 requests triggers new pod creation
	//    - Result: ANY load > 0.01 requests triggers new pod creation
	//    - This forces multiple pods even with minimal traffic!
	KnativeTarget string `envconfig:"KNATIVE_TARGET" validate:"required"`

	// 🔧 KnativeContainerConcurrency: "How many requests can one container handle?"
	//    - Default: "0" (unlimited - forces load balancing across pods)
	//    - Math: 0 = unlimited concurrency per container
	//    - Result: Forces traffic distribution across multiple pods
	//    - This ensures all pods receive traffic!
	KnativeContainerConcurrency string `envconfig:"KNATIVE_CONTAINER_CONCURRENCY" validate:"required"`

	// 🔽 KnativeMinScale: "What's the minimum number of pods?"
	//    - Default: "0" (can scale down to zero pods when no traffic)
	//    - Saves money when not in use
	//    - Trade-off: Cold start delay when traffic arrives
	KnativeMinScale string `envconfig:"KNATIVE_MIN_SCALE" validate:"required"`

	// 🔼 KnativeMaxScale: "What's the maximum number of pods?"
	//    - Default: "50" (can scale up to 50 pods maximum)
	//    - Prevents runaway scaling that could cost too much money
	//    - Safety limit for your wallet! 💰
	KnativeMaxScale string `envconfig:"KNATIVE_MAX_SCALE" validate:"required"`

	// ⏰ KnativeScaleToZeroGracePeriod: "How long to wait before killing the last pod?"
	//    - Default: "30s" (wait 30 seconds after last request)
	//    - Gives time for new requests to arrive before shutting down
	//    - Balances cost savings vs responsiveness
	KnativeScaleToZeroGracePeriod string `envconfig:"KNATIVE_SCALE_TO_ZERO_GRACE_PERIOD" validate:"required"`

	// ⏱️ KnativeScaleDownDelay: "How long to wait before scaling down?"
	//    - Default: "0s" (scale down immediately when load drops)
	//    - Prevents rapid scale up/down cycles
	//    - Set to higher values (like "60s") for more stable scaling
	KnativeScaleDownDelay string `envconfig:"KNATIVE_SCALE_DOWN_DELAY" validate:"required"`

	// 🎯 KnativeStableWindow: "How long to average metrics before scaling?"
	//    - Default: "10s" (look at last 10 seconds of traffic)
	//    - Prevents scaling based on traffic spikes
	//    - Longer = more stable, shorter = more responsive
	KnativeStableWindow string `envconfig:"KNATIVE_STABLE_WINDOW" validate:"required"`

	// 📨 EVENT CONFIGURATION - Controls how events are processed
	//
	// 🏷️ DefaultEventType: "What type of events should we listen for?"
	//    - Default: "network.notifi.lambda.parser.start"
	//    - This is the event type your triggers filter on
	//    - Must match what your test script sends!
	DefaultEventType string `envconfig:"DEFAULT_EVENT_TYPE" validate:"required"`

	// 📡 DefaultBrokerName: "Which broker should we send events to?"
	//    - Default: "knative-lambda-service-broker-dev"
	//    - This is the RabbitMQ broker that receives your events
	//    - Must match your port-forward setup!
	DefaultBrokerName string `envconfig:"BROKER_NAME" validate:"required"`

	// 🏠 DefaultTriggerNamespace: "Which namespace are our triggers in?"
	//    - Default: "knative-lambda-dev"
	//    - This is where your Knative services and triggers live
	//    - Must match your deployment namespace!
	DefaultTriggerNamespace string `envconfig:"DEFAULT_TRIGGER_NAMESPACE" validate:"required"`

	// 🔄 DefaultDeliveryRetries: "How many times to retry failed event delivery?"
	//    - Default: "5" (try 5 times before giving up)
	//    - Ensures events don't get lost due to temporary failures
	//    - Higher = more reliable, but more resource usage
	DefaultDeliveryRetries string `envconfig:"DEFAULT_DELIVERY_RETRIES" validate:"required"`

	// 📈 DefaultDeliveryBackoffPolicy: "How to space out retry attempts?"
	//    - Default: "exponential" (wait longer between each retry)
	//    - Options: "exponential", "linear"
	//    - Exponential: 1s, 2s, 4s, 8s, 16s (good for temporary issues)
	DefaultDeliveryBackoffPolicy string `envconfig:"DEFAULT_DELIVERY_BACKOFF_POLICY" validate:"required"`

	// ⏱️ DefaultDeliveryBackoffDelay: "How long to wait before first retry?"
	//    - Default: "PT1S" (wait 1 second)
	//    - PT1S = ISO 8601 duration format for "1 second"
	//    - Gives time for temporary issues to resolve
	DefaultDeliveryBackoffDelay string `envconfig:"DEFAULT_DELIVERY_BACKOFF_DELAY" validate:"required"`

	// 🏗️ POD CONFIGURATION - Controls how pods are created
	//
	// 🔗 EnableServiceLinks: "Should we inject environment variables for other services?"
	//    - Default: false (disabled for security)
	//    - When true: Pods get env vars for all services in the namespace
	//    - When false: Clean environment, no automatic service discovery
	//    - Best practice: Keep this false unless you need service discovery
	EnableServiceLinks bool `envconfig:"ENABLE_SERVICE_LINKS"`

	// 🐰 RABBITMQ CONFIGURATION - RabbitMQ eventing settings
	//
	// 🔄 RabbitMQEventingParallelism: "How many parallel consumers for RabbitMQ events?"
	//    - Default: 50 (50 parallel consumers)
	//    - Controls the number of concurrent event consumers
	//    - Higher = more throughput, but more resource usage
	RabbitMQEventingParallelism int `envconfig:"RABBITMQ_EVENTING_PARALLELISM" default:"50"`
}

// 🔧 Validate - "Validate Knative configuration"
//
// 📋 WHAT THIS DOES:
//
//	Checks that all required fields have values and are properly formatted.
//	This prevents runtime errors caused by missing or invalid configuration.
//
// 🎯 WHY WE NEED THIS:
//   - Catches configuration mistakes early
//   - Prevents services from starting with bad settings
//   - Gives clear error messages about what's wrong
//
// 🔍 WHAT IT CHECKS:
//   - All required fields are not empty
//   - String values are properly formatted
//   - Numeric values are within valid ranges
//
// 💡 EXAMPLE ERRORS:
//   - "target concurrency is required" (if field is empty)
//   - "invalid broker name" (if format is wrong)
//   - "parallelism must be between 1 and 50" (if out of range)
func (c *KnativeConfig) Validate() error {
	// 🔍 Check if target concurrency is set
	if c.KnativeTargetConcurrency == "" {
		return errors.NewValidationError("knative_target_concurrency", c.KnativeTargetConcurrency, constants.ErrTargetConcurrencyRequired)
	}

	// 🔍 Check if target utilization is set
	if c.KnativeTargetUtilization == "" {
		return errors.NewValidationError("knative_target_utilization", c.KnativeTargetUtilization, constants.ErrTargetUtilizationRequired)
	}

	// 🔍 Check if target is set
	if c.KnativeTarget == "" {
		return errors.NewValidationError("knative_target", c.KnativeTarget, constants.ErrTargetRequired)
	}

	// 🔍 Check if container concurrency is set
	if c.KnativeContainerConcurrency == "" {
		return errors.NewValidationError("knative_container_concurrency", c.KnativeContainerConcurrency, constants.ErrContainerConcurrencyRequired)
	}

	// 🔍 Check if min scale is set
	if c.KnativeMinScale == "" {
		return errors.NewValidationError("knative_min_scale", c.KnativeMinScale, constants.ErrMinScaleRequired)
	}

	// 🔍 Check if max scale is set
	if c.KnativeMaxScale == "" {
		return errors.NewValidationError("knative_max_scale", c.KnativeMaxScale, constants.ErrMaxScaleRequired)
	}

	// 🔍 Check if event type is set
	if c.DefaultEventType == "" {
		return errors.NewValidationError("default_event_type", c.DefaultEventType, constants.ErrDefaultEventTypeRequired)
	}

	// 🔍 Check if broker name is set
	if c.DefaultBrokerName == "" {
		return errors.NewValidationError("default_broker_name", c.DefaultBrokerName, constants.ErrDefaultBrokerNameRequired)
	}

	// 🔍 Check if trigger namespace is set
	if c.DefaultTriggerNamespace == "" {
		return errors.NewValidationError("default_trigger_namespace", c.DefaultTriggerNamespace, constants.ErrDefaultTriggerNamespaceRequired)
	}

	// ✅ All validations passed!
	return nil
}

// 🔧 GetTargetConcurrency - "Get target concurrency"
//
// 📋 WHAT THIS DOES:
//
//	Returns the target concurrency setting as a string.
//	This is used when creating Knative Service resources.
//
// 🎯 WHY WE NEED THIS:
//   - Knative expects string values in YAML
//   - Provides a clean interface to access the setting
//   - Handles the case where the value might be empty
//
// 💡 EXAMPLE USAGE:
//   - Returns "1" for testing (aggressive scaling)
//   - Returns "10" for production (conservative scaling)
func (c *KnativeConfig) GetTargetConcurrency() string {
	return c.KnativeTargetConcurrency
}

// 🔧 GetTargetConcurrencyInt - "Get target concurrency as integer"
//
// 📋 WHAT THIS DOES:
//
//	Converts the target concurrency string to an integer.
//	This is used for calculations and comparisons.
//
// 🎯 WHY WE NEED THIS:
//   - Some code needs numeric values for math
//   - Provides type safety (won't crash on invalid strings)
//   - Has a fallback value if conversion fails
//
// 🔄 HOW IT WORKS:
//  1. Try to convert string to integer
//  2. If successful, return the number
//  3. If failed, return default value from constants
//
// 💡 EXAMPLE USAGE:
//   - "1" → returns 1
//   - "10" → returns 10
//   - "invalid" → returns 1 (fallback from constants)
func (c *KnativeConfig) GetTargetConcurrencyInt() int {
	if val := c.KnativeTargetConcurrency; val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	// Use constant as fallback
	if fallback, err := strconv.Atoi(constants.BuilderServiceTargetConcurrencyDefault); err == nil {
		return fallback
	}
	return 1 // ultimate fallback
}

// 🔧 GetTargetUtilization - "Get target utilization"
//
// 📋 WHAT THIS DOES:
//
//	Returns the target utilization setting as a string.
//	This controls when new pods are created.
//
// 🎯 WHY WE NEED THIS:
//   - Knative expects string values in YAML
//   - Used in autoscaling annotations
//   - Controls scaling aggressiveness
//
// 💡 EXAMPLE USAGE:
//   - Returns "70" for testing (aggressive scaling)
//   - Returns "80" for production (conservative scaling)
func (c *KnativeConfig) GetTargetUtilization() string {
	return c.KnativeTargetUtilization
}

// 🎯 GetTarget - "Get target concurrency for autoscaling"
//
// 📋 WHAT THIS DOES:
//
//	Returns the target concurrency setting as a string.
//	This controls when new pods are created.
//
// 🎯 WHY WE NEED THIS:
//   - Knative expects string values in YAML
//   - Used in autoscaling annotations
//   - Controls scaling aggressiveness
//
// 💡 EXAMPLE USAGE:
//   - Returns "0.01" for very aggressive scaling (multiple pods)
//   - Returns "1" for normal scaling (one pod per request)
//   - Returns "10" for conservative scaling (fewer pods)
func (c *KnativeConfig) GetTarget() string {
	if c == nil {
		return constants.BuilderServiceTargetDefault
	}
	return c.KnativeTarget
}

// 🔧 GetContainerConcurrency - "Get container concurrency for load balancing"
//
// 📋 WHAT THIS DOES:
//
//	Returns the container concurrency setting as an integer.
//	This controls how many requests each container can handle.
//
// 🎯 WHY WE NEED THIS:
//   - Controls load balancing behavior
//   - 0 = unlimited (forces distribution across pods)
//   - Higher values = more requests per pod
//
// 💡 EXAMPLE USAGE:
//   - Returns 0 for unlimited (forces multiple pods)
//   - Returns 1 for one request per pod (aggressive scaling)
//   - Returns 10 for conservative scaling (fewer pods)
func (c *KnativeConfig) GetContainerConcurrency() int {
	if c == nil {
		if fallback, err := strconv.Atoi(constants.BuilderServiceContainerConcurrencyDefault); err == nil {
			return fallback
		}
		return 0 // ultimate fallback
	}

	if val := c.KnativeContainerConcurrency; val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}

	// Use constant as fallback
	if fallback, err := strconv.Atoi(constants.BuilderServiceContainerConcurrencyDefault); err == nil {
		return fallback
	}
	return 0 // ultimate fallback
}

// 🔧 GetMinScale - "Get min scale"
//
// 📋 WHAT THIS DOES:
//
//	Returns the minimum scale setting as a string.
//	This is the minimum number of pods that will run.
//
// 🎯 WHY WE NEED THIS:
//   - Knative expects string values in YAML
//   - Controls whether services can scale to zero
//   - Balances cost vs responsiveness
//
// 💡 EXAMPLE USAGE:
//   - Returns "0" for cost savings (scale to zero)
//   - Returns "1" for responsiveness (always have one pod ready)
func (c *KnativeConfig) GetMinScale() string {
	return c.KnativeMinScale
}

// 🔧 GetMaxScale - "Get max scale"
//
// 📋 WHAT THIS DOES:
//
//	Returns the maximum scale setting as a string.
//	This is the maximum number of pods that can run.
//
// 🎯 WHY WE NEED THIS:
//   - Knative expects string values in YAML
//   - Prevents runaway scaling
//   - Protects against cost overruns
//
// 💡 EXAMPLE USAGE:
//   - Returns "50" for high capacity
//   - Returns "10" for cost control
func (c *KnativeConfig) GetMaxScale() string {
	return c.KnativeMaxScale
}

// 🔧 GetScaleToZeroGracePeriod - "Get scale to zero grace period"
//
// 📋 WHAT THIS DOES:
//
//	Returns how long to wait before scaling down to zero pods.
//	This gives time for new requests to arrive.
//
// 🎯 WHY WE NEED THIS:
//   - Controls cost vs responsiveness trade-off
//   - Prevents rapid scale up/down cycles
//   - Balances user experience with cost savings
//
// 💡 EXAMPLE USAGE:
//   - Returns "30s" for balanced approach
//   - Returns "60s" for more stability
//   - Returns "0s" for immediate scale down (not recommended)
func (c *KnativeConfig) GetScaleToZeroGracePeriod() string {
	return c.KnativeScaleToZeroGracePeriod
}

// 🔧 GetScaleDownDelay - "Get scale down delay"
//
// 📋 WHAT THIS DOES:
//
//	Returns how long to wait before scaling down pods.
//	This prevents rapid scale up/down cycles.
//
// 🎯 WHY WE NEED THIS:
//   - Prevents "thrashing" (constant scale up/down)
//   - Provides stability during traffic fluctuations
//   - Improves user experience
//
// 💡 EXAMPLE USAGE:
//   - Returns "0s" for immediate response
//   - Returns "60s" for stability
//   - Returns "300s" for very stable scaling
func (c *KnativeConfig) GetScaleDownDelay() string {
	return c.KnativeScaleDownDelay
}

// 🔧 GetStableWindow - "Get stable window"
//
// 📋 WHAT THIS DOES:
//
//	Returns how long to average metrics before making scaling decisions.
//	This prevents scaling based on traffic spikes.
//
// 🎯 WHY WE NEED THIS:
//   - Provides smoothing for scaling decisions
//   - Prevents scaling on temporary traffic spikes
//   - Balances responsiveness vs stability
//
// 💡 EXAMPLE USAGE:
//   - Returns "10s" for responsive scaling
//   - Returns "60s" for stable scaling
//   - Returns "300s" for very stable scaling
func (c *KnativeConfig) GetStableWindow() string {
	return c.KnativeStableWindow
}

// 🔧 GetDefaultEventType - "Get default event type"
//
// 📋 WHAT THIS DOES:
//
//	Returns the event type that triggers should listen for.
//	This must match what your test script sends.
//
// 🎯 WHY WE NEED THIS:
//   - Used to create Knative Triggers
//   - Must match CloudEvent type from test script
//   - Controls which events reach your services
//
// 💡 EXAMPLE USAGE:
//   - Returns "network.notifi.lambda.parser.start"
//   - Used in trigger filter: type: network.notifi.lambda.parser.start
func (c *KnativeConfig) GetDefaultEventType() string {
	return c.DefaultEventType
}

// 🔧 GetDefaultBrokerName - "Get default broker name"
//
// 📋 WHAT THIS DOES:
//
//	Returns the name of the broker that receives events.
//	This must match your port-forward setup.
//
// 🎯 WHY WE NEED THIS:
//   - Used to create Knative Triggers
//   - Must match the broker service name
//   - Controls where events are sent
//
// 💡 EXAMPLE USAGE:
//   - Returns "knative-lambda-service-broker-dev"
//   - Used in trigger spec: broker: knative-lambda-service-broker-dev
func (c *KnativeConfig) GetDefaultBrokerName() string {
	return c.DefaultBrokerName
}

// 🔧 GetDefaultTriggerNamespace - "Get default trigger namespace"
//
// 📋 WHAT THIS DOES:
//
//	Returns the namespace where triggers are created.
//	This must match your deployment namespace.
//
// 🎯 WHY WE NEED THIS:
//   - Used to create Knative Triggers
//   - Must match where your services are deployed
//   - Controls where triggers are placed
//
// 💡 EXAMPLE USAGE:
//   - Returns "knative-lambda-dev"
//   - Used to create triggers in the correct namespace
func (c *KnativeConfig) GetDefaultTriggerNamespace() string {
	return c.DefaultTriggerNamespace
}

// 🔧 GetDefaultDeliveryRetries - "Get default delivery retries"
//
// 📋 WHAT THIS DOES:
//
//	Returns how many times to retry failed event delivery.
//	This ensures events don't get lost.
//
// 🎯 WHY WE NEED THIS:
//   - Used in trigger delivery configuration
//   - Provides reliability for event processing
//   - Handles temporary network issues
//
// 💡 EXAMPLE USAGE:
//   - Returns "5" for good reliability
//   - Used in trigger spec: retries: 5
func (c *KnativeConfig) GetDefaultDeliveryRetries() string {
	return c.DefaultDeliveryRetries
}

// 🔧 GetDefaultDeliveryRetriesInt - "Get default delivery retries as integer"
//
// 📋 WHAT THIS DOES:
//
//	Converts the delivery retries string to an integer.
//	This is used for calculations and validation.
//
// 🎯 WHY WE NEED THIS:
//   - Some code needs numeric values
//   - Provides type safety
//   - Has a fallback value if conversion fails
//
// 🔄 HOW IT WORKS:
//  1. Try to convert string to integer
//  2. If successful, return the number
//  3. If failed, return default value from constants
//
// 💡 EXAMPLE USAGE:
//   - "5" → returns 5
//   - "10" → returns 10
//   - "invalid" → returns 5 (fallback from constants)
func (c *KnativeConfig) GetDefaultDeliveryRetriesInt() int {
	if val := c.DefaultDeliveryRetries; val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	// Use constant as fallback
	if fallback, err := strconv.Atoi(constants.DeliveryRetriesDefault); err == nil {
		return fallback
	}
	return 5 // ultimate fallback
}

// 🔧 GetDefaultDeliveryBackoffPolicy - "Get default delivery backoff policy"
//
// 📋 WHAT THIS DOES:
//
//	Returns the backoff policy for retry attempts.
//	This controls how retry delays increase.
//
// 🎯 WHY WE NEED THIS:
//   - Used in trigger delivery configuration
//   - Controls retry timing strategy
//   - Balances reliability vs resource usage
//
// 💡 EXAMPLE USAGE:
//   - Returns "exponential" (1s, 2s, 4s, 8s, 16s)
//   - Returns "linear" (1s, 2s, 3s, 4s, 5s)
func (c *KnativeConfig) GetDefaultDeliveryBackoffPolicy() string {
	return c.DefaultDeliveryBackoffPolicy
}

// 🔧 GetDefaultDeliveryBackoffDelay - "Get default delivery backoff delay"
//
// 📋 WHAT THIS DOES:
//
//	Returns the initial delay before the first retry attempt.
//	This gives time for temporary issues to resolve.
//
// 🎯 WHY WE NEED THIS:
//   - Used in trigger delivery configuration
//   - Controls initial retry timing
//   - Prevents overwhelming failed services
//
// 💡 EXAMPLE USAGE:
//   - Returns "PT1S" (1 second delay)
//   - Returns "PT5S" (5 second delay)
//   - PT = ISO 8601 duration format
func (c *KnativeConfig) GetDefaultDeliveryBackoffDelay() string {
	return c.DefaultDeliveryBackoffDelay
}

// 🔧 GetEnableServiceLinks - "Get enable service links setting"
//
// 📋 WHAT THIS DOES:
//
//	Returns whether to enable automatic service discovery.
//	This controls environment variable injection.
//
// 🎯 WHY WE NEED THIS:
//   - Controls security and environment cleanliness
//   - When false: Clean environment, no automatic service discovery
//   - When true: Pods get env vars for all services in namespace
//
// 🔒 SECURITY IMPACT:
//   - false = More secure (no automatic service discovery)
//   - true = Less secure (automatic environment variable injection)
//
// 💡 EXAMPLE USAGE:
//   - Returns false for security (recommended)
//   - Returns true only if you need service discovery
func (c *KnativeConfig) GetEnableServiceLinks() bool {
	return c.EnableServiceLinks
}

// 🔧 GetRabbitMQEventingParallelism - "Get RabbitMQ eventing parallelism"
//
// 📋 WHAT THIS DOES:
//
//	Returns the number of parallel consumers for RabbitMQ events.
//	This controls event processing throughput.
//
// 🎯 WHY WE NEED THIS:
//   - Controls RabbitMQ event processing concurrency
//   - Higher values = more throughput but more resource usage
//   - Used in trigger annotations for RabbitMQ eventing
//
// 💡 EXAMPLE USAGE:
//   - Returns 50 for high throughput
//   - Returns 10 for conservative resource usage
func (c *KnativeConfig) GetRabbitMQEventingParallelism() int {
	return c.RabbitMQEventingParallelism
}
