package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		wantErr bool
	}{
		{
			name: "valid configuration",
			envVars: map[string]string{
				"ENVIRONMENT": "test",
				"NAMESPACE":   "test-namespace",
			},
			wantErr: false,
		},
		{
			name:    "missing required environment",
			envVars: map[string]string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			defer func() {
				for k := range tt.envVars {
					os.Unsetenv(k)
				}
			}()

			config, err := LoadConfig()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, config)
			assert.Equal(t, tt.envVars["ENVIRONMENT"], config.Environment)
			if tt.envVars["NAMESPACE"] != "" {
				assert.Equal(t, tt.envVars["NAMESPACE"], config.Kubernetes.Namespace)
			}
		})
	}
}

func TestConfig_Environment(t *testing.T) {
	os.Setenv("ENVIRONMENT", "test-env")
	os.Setenv("NAMESPACE", "test-namespace")
	defer func() {
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("NAMESPACE")
	}()

	config, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "test-env", config.Environment)
}

func TestConfig_Namespace(t *testing.T) {
	os.Setenv("ENVIRONMENT", "test")
	os.Setenv("NAMESPACE", "test-namespace")
	defer func() {
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("NAMESPACE")
	}()

	config, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "test-namespace", config.Kubernetes.Namespace)
}

func TestConfig_HTTP(t *testing.T) {
	os.Setenv("ENVIRONMENT", "test")
	os.Setenv("NAMESPACE", "test-namespace")
	os.Setenv("HTTP_PORT", "8080")
	defer func() {
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("NAMESPACE")
		os.Unsetenv("HTTP_PORT")
	}()

	config, err := LoadConfig()
	require.NoError(t, err)

	assert.NotNil(t, config.HTTP)
	assert.Equal(t, "8080", config.HTTP.Port)
}

func TestConfig_AWS(t *testing.T) {
	os.Setenv("ENVIRONMENT", "test")
	os.Setenv("NAMESPACE", "test-namespace")
	os.Setenv("AWS_REGION", "us-west-2")
	os.Setenv("AWS_ACCOUNT_ID", "123456789012")
	os.Setenv("ECR_REGISTRY", "123456789012.dkr.ecr.us-west-2.amazonaws.com")
	defer func() {
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("NAMESPACE")
		os.Unsetenv("AWS_REGION")
		os.Unsetenv("AWS_ACCOUNT_ID")
		os.Unsetenv("ECR_REGISTRY")
	}()

	config, err := LoadConfig()
	require.NoError(t, err)

	assert.NotNil(t, config.AWS)
	assert.Equal(t, "us-west-2", config.AWS.GetRegion())
}

func TestConfig_Kubernetes(t *testing.T) {
	os.Setenv("ENVIRONMENT", "test")
	os.Setenv("NAMESPACE", "test-namespace")
	defer func() {
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("NAMESPACE")
	}()

	config, err := LoadConfig()
	require.NoError(t, err)

	assert.NotNil(t, config.Kubernetes)
	assert.Equal(t, "test-namespace", config.Kubernetes.Namespace)
}

func TestConfig_Observability(t *testing.T) {
	os.Setenv("ENVIRONMENT", "test")
	os.Setenv("NAMESPACE", "test-namespace")
	os.Setenv("LOG_LEVEL", "debug")
	defer func() {
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("NAMESPACE")
		os.Unsetenv("LOG_LEVEL")
	}()

	config, err := LoadConfig()
	require.NoError(t, err)

	assert.NotNil(t, config.Observability)
	assert.Equal(t, "debug", config.Observability.LogLevel)
}

func TestConfig_Validate(t *testing.T) {
	os.Setenv("ENVIRONMENT", "test")
	os.Setenv("NAMESPACE", "test-namespace")
	defer func() {
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("NAMESPACE")
	}()

	config, err := LoadConfig()
	require.NoError(t, err)

	err = config.Validate()
	assert.NoError(t, err)
}

func TestConfig_ReloadFromEnvironment(t *testing.T) {
	os.Setenv("ENVIRONMENT", "test")
	os.Setenv("NAMESPACE", "test-namespace")
	defer func() {
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("NAMESPACE")
	}()

	config, err := LoadConfig()
	require.NoError(t, err)

	originalEnv := config.Environment

	// Change environment variable
	os.Setenv("ENVIRONMENT", "new-env")

	err = config.ReloadFromEnvironment()
	assert.NoError(t, err)
	assert.Equal(t, "new-env", config.Environment)
	assert.NotEqual(t, originalEnv, config.Environment)
}
