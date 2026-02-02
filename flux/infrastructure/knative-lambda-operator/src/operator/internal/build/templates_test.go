// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🧪 Unit Tests: Runtime Templates
//
//	Tests for runtime template processing:
//	- Parser context construction
//	- contextId extraction
//	- parameters extraction
//	- Edge cases and fallbacks
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package build

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📄 RUNTIME TEMPLATE TESTS - Parser Context Construction                │
// └─────────────────────────────────────────────────────────────────────────┘

func TestNodeJSRuntimeTemplate_ParserContextConstruction(t *testing.T) {
	template, err := GetRuntimeTemplate("nodejs")
	require.NoError(t, err, "Should load nodejs runtime template")
	require.NotEmpty(t, template, "Template should not be empty")

	// Test 1: Verify parser context construction code exists
	t.Run("contains parser context construction", func(t *testing.T) {
		assert.Contains(t, template, "parserContext", "Template should construct parserContext object")
		assert.Contains(t, template, "contextId", "Template should extract contextId")
		assert.Contains(t, template, "parameters", "Template should extract parameters")
	})

	// Test 2: Verify contextId extraction with fallback
	t.Run("contextId extraction with fallback", func(t *testing.T) {
		// Should have fallback to 'unknown-context'
		assert.Contains(t, template, "eventData.contextId || 'unknown-context'",
			"Template should have contextId fallback to 'unknown-context'")
	})

	// Test 3: Verify parameters extraction with fallbacks
	t.Run("parameters extraction with fallbacks", func(t *testing.T) {
		// Should check both eventData.parameters and eventData.parameter
		assert.Contains(t, template, "eventData.parameters || eventData.parameter || {}",
			"Template should check both 'parameters' and 'parameter' fields with empty object fallback")
	})

	// Test 4: Verify parser context is passed to handler
	t.Run("parser context passed to handler", func(t *testing.T) {
		assert.Contains(t, template, "handler(parserContext)",
			"Template should pass parserContext to handler, not event.data directly")
	})

	// Test 5: Verify logging for debugging
	t.Run("contains debug logging", func(t *testing.T) {
		assert.Contains(t, template, "Parser context prepared",
			"Template should log parser context preparation")
		assert.Contains(t, template, "hasParameters",
			"Template should log whether parameters exist")
		assert.Contains(t, template, "parametersKeys",
			"Template should log parameter keys for debugging")
	})
}

func TestNodeJSRuntimeTemplate_EventDataHandling(t *testing.T) {
	template, err := GetRuntimeTemplate("nodejs")
	require.NoError(t, err, "Should load nodejs runtime template")

	t.Run("handles missing event.data", func(t *testing.T) {
		// Should handle case where event.data is null/undefined
		assert.Contains(t, template, "event.data || {}",
			"Template should handle missing event.data with empty object fallback")
	})

	t.Run("constructs eventData variable", func(t *testing.T) {
		// Should create eventData variable for cleaner code
		assert.Contains(t, template, "const eventData = event.data || {}",
			"Template should create eventData variable")
	})
}

func TestNodeJSRuntimeTemplate_BackwardCompatibility(t *testing.T) {
	template, err := GetRuntimeTemplate("nodejs")
	require.NoError(t, err, "Should load nodejs runtime template")

	t.Run("supports both 'parameters' and 'parameter' field names", func(t *testing.T) {
		// Should support both singular and plural forms for backward compatibility
		hasParametersCheck := strings.Contains(template, "eventData.parameters")
		hasParameterCheck := strings.Contains(template, "eventData.parameter")

		assert.True(t, hasParametersCheck || hasParameterCheck,
			"Template should check for either 'parameters' or 'parameter' field")
	})
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📄 TEMPLATE CONTENT VALIDATION TESTS                                    │
// └─────────────────────────────────────────────────────────────────────────┘

func TestNodeJSRuntimeTemplate_RequiredComponents(t *testing.T) {
	template, err := GetRuntimeTemplate("nodejs")
	require.NoError(t, err, "Should load nodejs runtime template")

	requiredComponents := []struct {
		name        string
		content     string
		description string
	}{
		{
			name:        "HTTP.toEvent",
			content:     "HTTP.toEvent",
			description: "Should parse CloudEvent from HTTP request",
		},
		{
			name:        "handler function",
			content:     "handler",
			description: "Should call handler function",
		},
		{
			name:        "error handling",
			content:     "catch",
			description: "Should have error handling",
		},
		{
			name:        "correlation ID",
			content:     "correlationId",
			description: "Should extract correlation ID",
		},
		{
			name:        "logging",
			content:     "logInfo",
			description: "Should have logging functions",
		},
	}

	for _, tc := range requiredComponents {
		t.Run(tc.name, func(t *testing.T) {
			assert.Contains(t, template, tc.content,
				"Template should contain %s: %s", tc.name, tc.description)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📄 EDGE CASE TESTS                                                     │
// └─────────────────────────────────────────────────────────────────────────┘

func TestNodeJSRuntimeTemplate_EdgeCases(t *testing.T) {
	template, err := GetRuntimeTemplate("nodejs")
	require.NoError(t, err, "Should load nodejs runtime template")

	t.Run("handles null/undefined event.data", func(t *testing.T) {
		// Should not crash when event.data is null
		assert.Contains(t, template, "event.data || {}",
			"Template should handle null event.data gracefully")
	})

	t.Run("handles missing contextId", func(t *testing.T) {
		// Should use fallback when contextId is missing
		assert.Contains(t, template, "'unknown-context'",
			"Template should use 'unknown-context' fallback when contextId is missing")
	})

	t.Run("handles missing parameters", func(t *testing.T) {
		// Should use empty object when parameters are missing
		assert.Contains(t, template, "|| {}",
			"Template should use empty object fallback when parameters are missing")
	})

	t.Run("handles empty parameters object", func(t *testing.T) {
		// Should handle empty parameters object without errors
		assert.Contains(t, template, "Object.keys(parserContext.parameters || {})",
			"Template should safely handle empty parameters object")
	})
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📄 TEMPLATE INTEGRATION TESTS                                          │
// └─────────────────────────────────────────────────────────────────────────┘

func TestNodeJSRuntimeTemplate_Integration(t *testing.T) {
	template, err := GetRuntimeTemplate("nodejs")
	require.NoError(t, err, "Should load nodejs runtime template")

	t.Run("complete parser context flow", func(t *testing.T) {
		// Verify the complete flow from event.data to handler call
		flowSteps := []string{
			"const eventData = event.data || {}",
			"const parserContext = {",
			"contextId:",
			"parameters:",
			"handler(parserContext)",
		}

		for _, step := range flowSteps {
			assert.Contains(t, template, step,
				"Template should contain step: %s", step)
		}
	})

	t.Run("logging integration", func(t *testing.T) {
		// Verify logging is integrated with context preparation
		assert.Contains(t, template, "Parser context prepared",
			"Template should log context preparation")

		// Verify logging includes context details
		assert.Contains(t, template, "contextId: parserContext.contextId",
			"Template should log contextId in debug output")
		assert.Contains(t, template, "hasParameters: !!parserContext.parameters",
			"Template should log hasParameters flag")
	})
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📄 REGRESSION TESTS - Ensure fix is present                            │
// └─────────────────────────────────────────────────────────────────────────┘

func TestNodeJSRuntimeTemplate_RegressionTests(t *testing.T) {
	template, err := GetRuntimeTemplate("nodejs")
	require.NoError(t, err, "Should load nodejs runtime template")

	t.Run("does not pass event.data directly to handler", func(t *testing.T) {
		// This was the bug - handler was called with event.data directly
		// Now it should use parserContext instead
		hasDirectCall := strings.Contains(template, "handler(event.data")
		assert.False(t, hasDirectCall,
			"Template should NOT pass event.data directly to handler (this was the bug)")
	})

	t.Run("constructs parser context before handler call", func(t *testing.T) {
		// Verify parserContext is constructed before handler is called
		contextIndex := strings.Index(template, "const parserContext =")
		handlerIndex := strings.Index(template, "handler(parserContext)")

		assert.Greater(t, handlerIndex, contextIndex,
			"parserContext should be constructed before handler call")
	})

	t.Run("uses parserContext variable name", func(t *testing.T) {
		// Verify the variable is named parserContext (not safeEventData or other names)
		assert.Contains(t, template, "parserContext",
			"Template should use 'parserContext' variable name for clarity")
	})
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📄 TEMPLATE LOADING TESTS                                              │
// └─────────────────────────────────────────────────────────────────────────┘

func TestGetRuntimeTemplate_AllLanguages(t *testing.T) {
	languages := []struct {
		name     string
		language string
	}{
		{"Node.js", "nodejs"},
		{"Node", "node"},
		{"JavaScript", "javascript"},
		{"JS", "js"},
		{"Python", "python"},
		{"Go", "go"},
	}

	for _, lang := range languages {
		t.Run(lang.name, func(t *testing.T) {
			template, err := GetRuntimeTemplate(lang.language)
			require.NoError(t, err, "Should load template for %s", lang.name)
			assert.NotEmpty(t, template, "Template should not be empty for %s", lang.name)
		})
	}
}

func TestGetRuntimeTemplate_DefaultFallback(t *testing.T) {
	// Test that unknown language defaults to nodejs
	template, err := GetRuntimeTemplate("unknown-language")
	require.NoError(t, err, "Should not error on unknown language")
	assert.NotEmpty(t, template, "Should return nodejs template as default")

	// Verify it's actually the nodejs template
	assert.Contains(t, template, "node", "Default template should be nodejs")
}
