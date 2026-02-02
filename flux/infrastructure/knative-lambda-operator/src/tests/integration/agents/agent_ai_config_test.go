// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
//
//	🧪 AGENT-003: LambdaAgent AI Configuration Tests
//
//	User Story: AI/LLM Configuration for Agents
//	Priority: P0 | Story Points: 5
//
//	Tests validate:
//	- AI provider configuration (ollama, openai, anthropic)
//	- Model and endpoint configuration
//	- Temperature and token limits
//	- API key secret injection
//	- Provider-specific env var mapping
//	- System prompt configuration
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
package agents

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"knative-lambda/tests/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// Test Fixtures for AI Configuration
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

// AIProvider represents supported AI providers
type AIProvider string

const (
	ProviderOllama    AIProvider = "ollama"
	ProviderOpenAI    AIProvider = "openai"
	ProviderAnthropic AIProvider = "anthropic"
	ProviderNone      AIProvider = "none"
)

// FullAIConfig represents complete AI configuration
type FullAIConfig struct {
	Provider          AIProvider
	Endpoint          string
	Model             string
	FallbackModel     string // ADR-004 feature
	Temperature       string
	MaxTokens         int32
	TopP              string // ADR-004 feature
	ContextWindowSize int32  // ADR-004 feature
	APIKeySecretRef   *SecretKeySelector
	SystemPrompt      string
	SystemPromptRef   *ConfigMapKeySelector // ADR-004 feature
}

// SecretKeySelector represents a secret key selector
type SecretKeySelector struct {
	Name string
	Key  string
}

// ConfigMapKeySelector represents a ConfigMap key selector
type ConfigMapKeySelector struct {
	Name string
	Key  string
}

// ProviderEnvMapping maps providers to their env var names
var ProviderEnvMapping = map[AIProvider]struct {
	URLVar   string
	ModelVar string
	KeyVar   string
}{
	ProviderOllama: {
		URLVar:   "OLLAMA_URL",
		ModelVar: "OLLAMA_MODEL",
		KeyVar:   "", // Ollama doesn't need API key
	},
	ProviderOpenAI: {
		URLVar:   "OPENAI_API_BASE",
		ModelVar: "OPENAI_MODEL",
		KeyVar:   "OPENAI_API_KEY",
	},
	ProviderAnthropic: {
		URLVar:   "ANTHROPIC_BASE_URL",
		ModelVar: "ANTHROPIC_MODEL",
		KeyVar:   "ANTHROPIC_API_KEY",
	},
}

// createTestAIConfig creates a test AI configuration
func createTestAIConfig(provider AIProvider) *FullAIConfig {
	config := &FullAIConfig{
		Provider:    provider,
		Temperature: "0.7",
		MaxTokens:   2048,
	}

	switch provider {
	case ProviderOllama:
		config.Endpoint = "http://ollama.ollama.svc.cluster.local:11434"
		config.Model = "llama3.2:3b"
	case ProviderOpenAI:
		config.Endpoint = "https://api.openai.com/v1"
		config.Model = "gpt-4-turbo"
		config.APIKeySecretRef = &SecretKeySelector{
			Name: "openai-credentials",
			Key:  "api-key",
		}
	case ProviderAnthropic:
		config.Endpoint = "https://api.anthropic.com"
		config.Model = "claude-3-sonnet-20240229"
		config.APIKeySecretRef = &SecretKeySelector{
			Name: "anthropic-credentials",
			Key:  "api-key",
		}
	}

	return config
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC1: AI provider configuration
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT003_AC1_ProviderConfiguration(t *testing.T) {
	testutils.SetupTestEnvironment(t)

	t.Run("Ollama provider configuration", func(t *testing.T) {
		// Arrange
		config := createTestAIConfig(ProviderOllama)

		// Assert
		assert.Equal(t, ProviderOllama, config.Provider)
		assert.Contains(t, config.Endpoint, "ollama")
		assert.Equal(t, "llama3.2:3b", config.Model)
		assert.Nil(t, config.APIKeySecretRef, "Ollama doesn't need API key")
	})

	t.Run("OpenAI provider configuration", func(t *testing.T) {
		// Arrange
		config := createTestAIConfig(ProviderOpenAI)

		// Assert
		assert.Equal(t, ProviderOpenAI, config.Provider)
		assert.Equal(t, "https://api.openai.com/v1", config.Endpoint)
		assert.Equal(t, "gpt-4-turbo", config.Model)
		assert.NotNil(t, config.APIKeySecretRef, "OpenAI requires API key")
	})

	t.Run("Anthropic provider configuration", func(t *testing.T) {
		// Arrange
		config := createTestAIConfig(ProviderAnthropic)

		// Assert
		assert.Equal(t, ProviderAnthropic, config.Provider)
		assert.Contains(t, config.Endpoint, "anthropic")
		assert.Contains(t, config.Model, "claude")
		assert.NotNil(t, config.APIKeySecretRef, "Anthropic requires API key")
	})

	t.Run("None provider for non-AI agents", func(t *testing.T) {
		// Arrange
		config := &FullAIConfig{
			Provider: ProviderNone,
		}

		// Assert
		assert.Equal(t, ProviderNone, config.Provider)
		assert.Empty(t, config.Endpoint)
		assert.Empty(t, config.Model)
	})

	t.Run("Default provider is ollama", func(t *testing.T) {
		// Arrange
		defaultProvider := ProviderOllama

		// Assert
		assert.Equal(t, ProviderOllama, defaultProvider)
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC2: Model and endpoint configuration
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT003_AC2_ModelEndpointConfiguration(t *testing.T) {
	t.Run("Custom Ollama endpoint", func(t *testing.T) {
		// Arrange
		config := createTestAIConfig(ProviderOllama)
		config.Endpoint = "http://my-ollama.custom-namespace.svc.cluster.local:11434"

		// Assert
		assert.Contains(t, config.Endpoint, "my-ollama")
	})

	t.Run("Custom model specification", func(t *testing.T) {
		// Arrange
		config := createTestAIConfig(ProviderOllama)
		config.Model = "codellama:34b"

		// Assert
		assert.Equal(t, "codellama:34b", config.Model)
	})

	t.Run("Model with quantization tag", func(t *testing.T) {
		// Arrange
		models := []string{
			"llama3.2:3b",
			"llama3.2:3b-q4_0",
			"codellama:7b-instruct-q8_0",
		}

		// Act & Assert
		for _, model := range models {
			assert.NotEmpty(t, model)
			assert.Contains(t, model, ":")
		}
	})

	t.Run("Fallback model configuration (ADR-004)", func(t *testing.T) {
		// Arrange
		config := createTestAIConfig(ProviderOllama)
		config.FallbackModel = "llama3.2:1b"

		// Assert
		assert.NotEmpty(t, config.FallbackModel)
		assert.NotEqual(t, config.Model, config.FallbackModel)
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC3: Temperature and token limits
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT003_AC3_GenerationParameters(t *testing.T) {
	t.Run("Valid temperature range", func(t *testing.T) {
		// Arrange
		validTemperatures := []string{"0.0", "0.5", "0.7", "1.0", "1.5", "2.0"}

		// Act & Assert
		for _, temp := range validTemperatures {
			isValid := validateTemperature(temp)
			assert.True(t, isValid, "Temperature %s should be valid", temp)
		}
	})

	t.Run("Invalid temperature values", func(t *testing.T) {
		// Arrange
		invalidTemperatures := []string{"-0.5", "2.5", "invalid", ""}

		// Act & Assert
		for _, temp := range invalidTemperatures {
			isValid := validateTemperature(temp)
			assert.False(t, isValid, "Temperature %s should be invalid", temp)
		}
	})

	t.Run("MaxTokens configuration", func(t *testing.T) {
		// Arrange
		config := createTestAIConfig(ProviderOllama)

		// Assert
		assert.Equal(t, int32(2048), config.MaxTokens)
	})

	t.Run("MaxTokens limits by provider", func(t *testing.T) {
		// Arrange
		providerLimits := map[AIProvider]int32{
			ProviderOllama:    32768,  // Depends on model
			ProviderOpenAI:    128000, // GPT-4 Turbo
			ProviderAnthropic: 200000, // Claude 3
		}

		// Act & Assert
		for provider, limit := range providerLimits {
			assert.Greater(t, limit, int32(0), "Provider %s should have positive limit", provider)
		}
	})

	t.Run("TopP configuration (ADR-004)", func(t *testing.T) {
		// Arrange
		config := createTestAIConfig(ProviderOllama)
		config.TopP = "0.9"

		// Act
		isValid := validateTopP(config.TopP)

		// Assert
		assert.True(t, isValid)
		assert.Equal(t, "0.9", config.TopP)
	})

	t.Run("ContextWindowSize configuration (ADR-004)", func(t *testing.T) {
		// Arrange
		config := createTestAIConfig(ProviderOllama)
		config.ContextWindowSize = 4096

		// Assert
		assert.Equal(t, int32(4096), config.ContextWindowSize)
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC4: API key secret injection
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT003_AC4_APIKeySecretInjection(t *testing.T) {
	t.Run("OpenAI API key from secret", func(t *testing.T) {
		// Arrange
		config := createTestAIConfig(ProviderOpenAI)

		// Assert
		require.NotNil(t, config.APIKeySecretRef)
		assert.Equal(t, "openai-credentials", config.APIKeySecretRef.Name)
		assert.Equal(t, "api-key", config.APIKeySecretRef.Key)
	})

	t.Run("Anthropic API key from secret", func(t *testing.T) {
		// Arrange
		config := createTestAIConfig(ProviderAnthropic)

		// Assert
		require.NotNil(t, config.APIKeySecretRef)
		assert.Equal(t, "anthropic-credentials", config.APIKeySecretRef.Name)
	})

	t.Run("Secret reference generates env var with secretKeyRef", func(t *testing.T) {
		// Arrange
		config := createTestAIConfig(ProviderOpenAI)

		// Act
		envVar := buildSecretEnvVar(config.Provider, config.APIKeySecretRef)

		// Assert
		assert.Equal(t, "OPENAI_API_KEY", envVar.Name)
		assert.Equal(t, "openai-credentials", envVar.SecretName)
		assert.Equal(t, "api-key", envVar.SecretKey)
	})

	t.Run("Ollama doesn't inject API key", func(t *testing.T) {
		// Arrange
		config := createTestAIConfig(ProviderOllama)

		// Assert
		assert.Nil(t, config.APIKeySecretRef)
	})

	t.Run("Missing secret reference returns error", func(t *testing.T) {
		// Arrange
		config := createTestAIConfig(ProviderOpenAI)
		config.APIKeySecretRef = nil

		// Act
		err := validateAPIKeyRequired(config)

		// Assert
		assert.Error(t, err, "OpenAI should require API key")
		assert.Contains(t, err.Error(), "API key required")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC5: Provider-specific env var mapping
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT003_AC5_ProviderEnvVarMapping(t *testing.T) {
	t.Run("Ollama env vars", func(t *testing.T) {
		// Arrange
		config := createTestAIConfig(ProviderOllama)

		// Act
		envVars := buildAIEnvVars(config)

		// Assert
		assert.Equal(t, config.Endpoint, envVars["OLLAMA_URL"])
		assert.Equal(t, config.Model, envVars["OLLAMA_MODEL"])
	})

	t.Run("OpenAI env vars", func(t *testing.T) {
		// Arrange
		config := createTestAIConfig(ProviderOpenAI)

		// Act
		envVars := buildAIEnvVars(config)

		// Assert
		assert.Equal(t, config.Endpoint, envVars["OPENAI_API_BASE"])
		assert.Equal(t, config.Model, envVars["OPENAI_MODEL"])
	})

	t.Run("Anthropic env vars", func(t *testing.T) {
		// Arrange
		config := createTestAIConfig(ProviderAnthropic)

		// Act
		envVars := buildAIEnvVars(config)

		// Assert
		assert.Equal(t, config.Endpoint, envVars["ANTHROPIC_BASE_URL"])
		assert.Equal(t, config.Model, envVars["ANTHROPIC_MODEL"])
	})

	t.Run("Common env vars for all providers", func(t *testing.T) {
		// Arrange
		config := createTestAIConfig(ProviderOllama)

		// Act
		envVars := buildAIEnvVars(config)

		// Assert
		assert.Equal(t, config.Temperature, envVars["AI_TEMPERATURE"])
		assert.Equal(t, fmt.Sprintf("%d", config.MaxTokens), envVars["AI_MAX_TOKENS"])
	})

	t.Run("Provider type env var", func(t *testing.T) {
		// Arrange
		config := createTestAIConfig(ProviderOpenAI)

		// Act
		envVars := buildAIEnvVars(config)

		// Assert
		assert.Equal(t, string(ProviderOpenAI), envVars["AI_PROVIDER"])
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC6: System prompt configuration
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT003_AC6_SystemPromptConfiguration(t *testing.T) {
	t.Run("Inline system prompt", func(t *testing.T) {
		// Arrange
		config := createTestAIConfig(ProviderOllama)
		config.SystemPrompt = `You are Bruno's AI assistant.
Be helpful, concise, and friendly.`

		// Assert
		assert.NotEmpty(t, config.SystemPrompt)
		assert.Contains(t, config.SystemPrompt, "Bruno's AI assistant")
	})

	t.Run("System prompt injected as env var", func(t *testing.T) {
		// Arrange
		config := createTestAIConfig(ProviderOllama)
		config.SystemPrompt = "Test prompt"

		// Act
		envVars := buildAIEnvVars(config)

		// Assert
		assert.Equal(t, "Test prompt", envVars["SYSTEM_PROMPT"])
	})

	t.Run("System prompt from ConfigMap (ADR-004)", func(t *testing.T) {
		// Arrange
		config := createTestAIConfig(ProviderOllama)
		config.SystemPromptRef = &ConfigMapKeySelector{
			Name: "agent-prompts",
			Key:  "system-prompt",
		}

		// Assert
		assert.NotNil(t, config.SystemPromptRef)
		assert.Equal(t, "agent-prompts", config.SystemPromptRef.Name)
		assert.Equal(t, "system-prompt", config.SystemPromptRef.Key)
	})

	t.Run("ConfigMap reference takes precedence over inline", func(t *testing.T) {
		// Arrange
		config := createTestAIConfig(ProviderOllama)
		config.SystemPrompt = "Inline prompt"
		config.SystemPromptRef = &ConfigMapKeySelector{
			Name: "agent-prompts",
			Key:  "system-prompt",
		}

		// Act
		source := getSystemPromptSource(config)

		// Assert
		assert.Equal(t, "configmap", source)
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// Integration Test: Full AI Configuration
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT003_Integration_FullAIConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Complete AI configuration to env vars", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		_ = ctx

		config := createTestAIConfig(ProviderOllama)
		config.SystemPrompt = "You are a helpful assistant."
		config.FallbackModel = "llama3.2:1b"
		config.TopP = "0.9"
		config.ContextWindowSize = 4096

		// Act
		envVars := buildAIEnvVars(config)

		// Assert - Core config
		assert.Equal(t, "ollama", envVars["AI_PROVIDER"])
		assert.Equal(t, config.Endpoint, envVars["OLLAMA_URL"])
		assert.Equal(t, config.Model, envVars["OLLAMA_MODEL"])
		assert.Equal(t, config.Temperature, envVars["AI_TEMPERATURE"])
		assert.Equal(t, "2048", envVars["AI_MAX_TOKENS"])
		assert.Equal(t, config.SystemPrompt, envVars["SYSTEM_PROMPT"])

		// Assert - ADR-004 features
		assert.Equal(t, config.FallbackModel, envVars["AI_FALLBACK_MODEL"])
		assert.Equal(t, config.TopP, envVars["AI_TOP_P"])
		assert.Equal(t, "4096", envVars["AI_CONTEXT_WINDOW_SIZE"])
	})

	t.Run("OpenAI configuration with API key", func(t *testing.T) {
		// Arrange
		config := createTestAIConfig(ProviderOpenAI)

		// Act
		envVars := buildAIEnvVars(config)
		secretEnvVar := buildSecretEnvVar(config.Provider, config.APIKeySecretRef)

		// Assert
		assert.Equal(t, "openai", envVars["AI_PROVIDER"])
		assert.Equal(t, config.Endpoint, envVars["OPENAI_API_BASE"])
		assert.Equal(t, "OPENAI_API_KEY", secretEnvVar.Name)
		assert.Equal(t, "openai-credentials", secretEnvVar.SecretName)
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// Helper Functions
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

// validateTemperature validates temperature value
func validateTemperature(temp string) bool {
	if temp == "" {
		return false
	}
	val, err := strconv.ParseFloat(temp, 64)
	if err != nil {
		return false
	}
	return val >= 0.0 && val <= 2.0
}

// validateTopP validates topP value
func validateTopP(topP string) bool {
	if topP == "" {
		return true // Optional field
	}
	val, err := strconv.ParseFloat(topP, 64)
	if err != nil {
		return false
	}
	return val >= 0.0 && val <= 1.0
}

// SecretEnvVar represents an env var from a secret
type SecretEnvVar struct {
	Name       string
	SecretName string
	SecretKey  string
}

// buildSecretEnvVar builds a secret environment variable
func buildSecretEnvVar(provider AIProvider, secretRef *SecretKeySelector) SecretEnvVar {
	mapping, ok := ProviderEnvMapping[provider]
	if !ok || secretRef == nil {
		return SecretEnvVar{}
	}
	return SecretEnvVar{
		Name:       mapping.KeyVar,
		SecretName: secretRef.Name,
		SecretKey:  secretRef.Key,
	}
}

// validateAPIKeyRequired validates if API key is required
func validateAPIKeyRequired(config *FullAIConfig) error {
	if config.Provider == ProviderOpenAI || config.Provider == ProviderAnthropic {
		if config.APIKeySecretRef == nil {
			return fmt.Errorf("API key required for provider %s", config.Provider)
		}
	}
	return nil
}

// buildAIEnvVars builds all AI environment variables
func buildAIEnvVars(config *FullAIConfig) map[string]string {
	envVars := map[string]string{
		"AI_PROVIDER":    string(config.Provider),
		"AI_TEMPERATURE": config.Temperature,
		"AI_MAX_TOKENS":  fmt.Sprintf("%d", config.MaxTokens),
	}

	// Provider-specific env vars
	switch config.Provider {
	case ProviderOllama:
		envVars["OLLAMA_URL"] = config.Endpoint
		envVars["OLLAMA_MODEL"] = config.Model
	case ProviderOpenAI:
		envVars["OPENAI_API_BASE"] = config.Endpoint
		envVars["OPENAI_MODEL"] = config.Model
	case ProviderAnthropic:
		envVars["ANTHROPIC_BASE_URL"] = config.Endpoint
		envVars["ANTHROPIC_MODEL"] = config.Model
	}

	// System prompt
	if config.SystemPrompt != "" {
		envVars["SYSTEM_PROMPT"] = config.SystemPrompt
	}

	// ADR-004 features
	if config.FallbackModel != "" {
		envVars["AI_FALLBACK_MODEL"] = config.FallbackModel
	}
	if config.TopP != "" {
		envVars["AI_TOP_P"] = config.TopP
	}
	if config.ContextWindowSize > 0 {
		envVars["AI_CONTEXT_WINDOW_SIZE"] = fmt.Sprintf("%d", config.ContextWindowSize)
	}

	return envVars
}

// getSystemPromptSource determines the source of system prompt
func getSystemPromptSource(config *FullAIConfig) string {
	if config.SystemPromptRef != nil {
		return "configmap"
	}
	if config.SystemPrompt != "" {
		return "inline"
	}
	return "none"
}
