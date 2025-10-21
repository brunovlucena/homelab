package config

import (
	"os"
	"testing"

	"knative-lambda-new/internal/constants"

	"github.com/stretchr/testify/assert"
)

func TestObservabilityConfig_ExemplarsFromConstants(t *testing.T) {
	// Test case: Verify that exemplars configuration uses constants
	t.Run("UseConstantsForDefaults", func(t *testing.T) {
		// Create a new config builder and load observability config
		builder := NewConfigBuilder()
		builder.loadObservabilityFromEnvironment()

		// Verify that constants are used for defaults
		assert.Equal(t, constants.ExemplarsEnabledDefault, builder.config.Observability.ExemplarsEnabled)
		assert.Equal(t, constants.ExemplarsMaxPerMetricDefault, builder.config.Observability.ExemplarsMaxPerMetric)
		assert.Equal(t, constants.ExemplarsSampleRateDefault, builder.config.Observability.ExemplarsSampleRate)
		assert.Equal(t, constants.ExemplarsTraceIDLabelDefault, builder.config.Observability.ExemplarsTraceIDLabel)
		assert.Equal(t, constants.ExemplarsSpanIDLabelDefault, builder.config.Observability.ExemplarsSpanIDLabel)
		assert.Equal(t, constants.ExemplarsIncludeLabelsDefault, builder.config.Observability.ExemplarsIncludeLabels)
	})
}

func TestObservabilityConfig_ObservabilityFromConstants(t *testing.T) {
	// Test case: Verify that observability configuration uses constants
	t.Run("UseConstantsForDefaults", func(t *testing.T) {
		// Create a new config builder and load observability config
		builder := NewConfigBuilder()
		builder.loadObservabilityFromEnvironment()

		// Verify that constants are used for defaults
		assert.Equal(t, constants.TracingEnabledDefault, builder.config.Observability.TracingEnabled)
		assert.Equal(t, constants.SampleRateDefault, builder.config.Observability.SampleRate)
		assert.Equal(t, constants.MetricsEnabledDefault, builder.config.Observability.MetricsEnabled)
	})
}

func TestObservabilityConfig_ExemplarsFromEnvironment(t *testing.T) {
	// Test case: Load exemplars configuration from environment variables
	t.Run("LoadExemplarsFromEnvironment", func(t *testing.T) {
		// Set exemplars environment variables
		os.Setenv("EXEMPLARS_ENABLED", "false")
		os.Setenv("EXEMPLARS_MAX_PER_METRIC", "15")
		os.Setenv("EXEMPLARS_SAMPLE_RATE", "0.25")
		os.Setenv("EXEMPLARS_TRACE_ID_LABEL", "custom_trace_id")
		os.Setenv("EXEMPLARS_SPAN_ID_LABEL", "custom_span_id")
		os.Setenv("EXEMPLARS_INCLUDE_LABELS", "custom_label1,custom_label2")

		// Clean up environment variables after test
		defer func() {
			os.Unsetenv("EXEMPLARS_ENABLED")
			os.Unsetenv("EXEMPLARS_MAX_PER_METRIC")
			os.Unsetenv("EXEMPLARS_SAMPLE_RATE")
			os.Unsetenv("EXEMPLARS_TRACE_ID_LABEL")
			os.Unsetenv("EXEMPLARS_SPAN_ID_LABEL")
			os.Unsetenv("EXEMPLARS_INCLUDE_LABELS")
		}()

		// Create a new config builder and load observability config
		builder := NewConfigBuilder()
		builder.loadObservabilityFromEnvironment()

		// Verify that environment variables were loaded correctly
		assert.False(t, builder.config.Observability.ExemplarsEnabled)
		assert.Equal(t, 15, builder.config.Observability.ExemplarsMaxPerMetric)
		assert.Equal(t, 0.25, builder.config.Observability.ExemplarsSampleRate)
		assert.Equal(t, "custom_trace_id", builder.config.Observability.ExemplarsTraceIDLabel)
		assert.Equal(t, "custom_span_id", builder.config.Observability.ExemplarsSpanIDLabel)
		assert.Equal(t, "custom_label1,custom_label2", builder.config.Observability.ExemplarsIncludeLabels)
	})
}

func TestObservabilityConfig_NewObservabilityConfigUsesConstants(t *testing.T) {
	// Test case: Verify that NewObservabilityConfig uses constants
	t.Run("NewObservabilityConfigUsesConstants", func(t *testing.T) {
		// Create a new observability config
		config := NewObservabilityConfig()

		// Verify that constants are used
		assert.Equal(t, constants.TracingEnabledDefault, config.TracingEnabled)
		assert.Equal(t, constants.SampleRateDefault, config.SampleRate)
		assert.Equal(t, constants.MetricsEnabledDefault, config.MetricsEnabled)
		assert.Equal(t, constants.ExemplarsEnabledDefault, config.ExemplarsEnabled)
		assert.Equal(t, constants.ExemplarsMaxPerMetricDefault, config.ExemplarsMaxPerMetric)
		assert.Equal(t, constants.ExemplarsSampleRateDefault, config.ExemplarsSampleRate)
		assert.Equal(t, constants.ExemplarsTraceIDLabelDefault, config.ExemplarsTraceIDLabel)
		assert.Equal(t, constants.ExemplarsSpanIDLabelDefault, config.ExemplarsSpanIDLabel)
		assert.Equal(t, constants.ExemplarsIncludeLabelsDefault, config.ExemplarsIncludeLabels)
	})
}

// Test that should fail to verify test framework is working
func TestObservabilityConfig_ShouldFail(t *testing.T) {
	t.Run("ThisTestShouldFail", func(t *testing.T) {
		// This test should fail to verify our test framework works
		// Uncomment the line below to make it fail
		// t.Errorf("This test is intentionally failing to verify test framework")

		// For now, let's make it pass but add a comment explaining
		t.Log("This test is designed to fail when uncommented to verify test framework")
	})
}
