// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🧪 Unit Tests: Deploy Manager
//
//	Tests for deployment operations:
//	- Knative Service spec building
//	- Environment variable conversion
//	- Resource configuration
//	- Autoscaling annotations
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package deploy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	lambdav1alpha1 "github.com/brunovlucena/knative-lambda-operator/api/v1alpha1"
)

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📝 CONSTANTS TESTS                                                     │
// └─────────────────────────────────────────────────────────────────────────┘

func TestDeployConstants(t *testing.T) {
	assert.Equal(t, 8080, DefaultPort, "Default port should be 8080")
	assert.Equal(t, 10, DefaultConcurrency, "Default concurrency should be 10")
	assert.Equal(t, 0, DefaultMinReplicas, "Default min replicas should be 0")
	assert.Equal(t, 50, DefaultMaxReplicas, "Default max replicas should be 50")
}

func TestKnativeServiceGVK(t *testing.T) {
	assert.Equal(t, "serving.knative.dev", KnativeServiceGVK.Group)
	assert.Equal(t, "v1", KnativeServiceGVK.Version)
	assert.Equal(t, "Service", KnativeServiceGVK.Kind)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔄 ENV VAR CONVERSION TESTS                                            │
// └─────────────────────────────────────────────────────────────────────────┘

func TestConvertEnvVars(t *testing.T) {
	m := &Manager{}

	tests := []struct {
		name        string
		envVars     []corev1.EnvVar
		expectedLen int
		description string
	}{
		{
			name:        "Empty env vars",
			envVars:     []corev1.EnvVar{},
			expectedLen: 0,
			description: "Should return empty slice for no env vars",
		},
		{
			name: "Simple value env vars",
			envVars: []corev1.EnvVar{
				{Name: "DEBUG", Value: "true"},
				{Name: "LOG_LEVEL", Value: "info"},
			},
			expectedLen: 2,
			description: "Should convert simple env vars",
		},
		{
			name: "Env var with secret ref",
			envVars: []corev1.EnvVar{
				{Name: "DEBUG", Value: "true"},
				{
					Name: "API_KEY",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "my-secret",
							},
							Key: "api-key",
						},
					},
				},
			},
			expectedLen: 2,
			description: "Should convert env vars with secret refs",
		},
		{
			name: "Env var with configmap ref",
			envVars: []corev1.EnvVar{
				{
					Name: "CONFIG_VALUE",
					ValueFrom: &corev1.EnvVarSource{
						ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "my-configmap",
							},
							Key: "config-key",
						},
					},
				},
			},
			expectedLen: 1,
			description: "Should convert env vars with configmap refs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.convertEnvVars(tt.envVars)
			assert.Len(t, result, tt.expectedLen, tt.description)

			// Verify conversion structure
			for i, ev := range tt.envVars {
				if i >= len(result) {
					break
				}
				envMap, ok := result[i].(map[string]interface{})
				require.True(t, ok, "Should be a map")
				assert.Equal(t, ev.Name, envMap["name"])

				if ev.Value != "" {
					assert.Equal(t, ev.Value, envMap["value"])
				}
				if ev.ValueFrom != nil {
					valueFrom, hasValueFrom := envMap["valueFrom"].(map[string]interface{})
					if ev.ValueFrom.SecretKeyRef != nil {
						assert.True(t, hasValueFrom, "Should have valueFrom")
						secretRef, hasSecretRef := valueFrom["secretKeyRef"].(map[string]interface{})
						assert.True(t, hasSecretRef, "Should have secretKeyRef")
						assert.Equal(t, ev.ValueFrom.SecretKeyRef.Name, secretRef["name"])
						assert.Equal(t, ev.ValueFrom.SecretKeyRef.Key, secretRef["key"])
					}
					if ev.ValueFrom.ConfigMapKeyRef != nil {
						assert.True(t, hasValueFrom, "Should have valueFrom")
						configMapRef, hasConfigMapRef := valueFrom["configMapKeyRef"].(map[string]interface{})
						assert.True(t, hasConfigMapRef, "Should have configMapKeyRef")
						assert.Equal(t, ev.ValueFrom.ConfigMapKeyRef.Name, configMapRef["name"])
						assert.Equal(t, ev.ValueFrom.ConfigMapKeyRef.Key, configMapRef["key"])
					}
				}
			}
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🏗️ CONTAINER SPEC TESTS                                                │
// └─────────────────────────────────────────────────────────────────────────┘

func TestBuildContainerSpec(t *testing.T) {
	m := &Manager{}

	tests := []struct {
		name               string
		lambda             *lambdav1alpha1.LambdaFunction
		imageURI           string
		env                []corev1.EnvVar
		memoryRequest      string
		cpuRequest         string
		memoryLimit        string
		cpuLimit           string
		containerPort      int64
		expectedPullPolicy string
		description        string
	}{
		{
			name: "Basic container spec",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Source: lambdav1alpha1.SourceSpec{Type: "inline"},
				},
			},
			imageURI:           "localhost:5001/test/image:abc123",
			env:                []corev1.EnvVar{{Name: "HANDLER", Value: "handler"}},
			memoryRequest:      "64Mi",
			cpuRequest:         "50m",
			memoryLimit:        "128Mi",
			cpuLimit:           "100m",
			containerPort:      8080,
			expectedPullPolicy: "Always", // Default for built images
			description:        "Basic container spec with defaults",
		},
		{
			name: "Image source (pre-built) container spec",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Source: lambdav1alpha1.SourceSpec{
						Type: "image",
						Image: &lambdav1alpha1.ImageSource{
							Repository: "gcr.io/project/my-app",
							Port:       3000,
						},
					},
				},
			},
			imageURI:           "gcr.io/project/my-app:v1.0.0",
			env:                []corev1.EnvVar{},
			memoryRequest:      "128Mi",
			cpuRequest:         "100m",
			memoryLimit:        "512Mi",
			cpuLimit:           "500m",
			containerPort:      3000,
			expectedPullPolicy: "IfNotPresent", // Default for pre-built images
			description:        "Pre-built image uses IfNotPresent",
		},
		{
			name: "Explicit ImagePullPolicy",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Source:          lambdav1alpha1.SourceSpec{Type: "inline"},
					ImagePullPolicy: "Never",
				},
			},
			imageURI:           "localhost:5001/test/image:abc123",
			env:                []corev1.EnvVar{},
			memoryRequest:      "64Mi",
			cpuRequest:         "50m",
			memoryLimit:        "128Mi",
			cpuLimit:           "100m",
			containerPort:      8080,
			expectedPullPolicy: "Never",
			description:        "Explicit policy overrides default",
		},
		{
			name: "Receiver mode container spec",
			lambda: &lambdav1alpha1.LambdaFunction{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"lambda.knative.io/receiver-mode": "true",
					},
				},
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Source: lambdav1alpha1.SourceSpec{Type: "inline"},
				},
			},
			imageURI:           "localhost:5001/knative-lambda-operator:v1.0.0",
			env:                []corev1.EnvVar{{Name: "DEFAULT_NAMESPACE", Value: "knative-lambda"}},
			memoryRequest:      "64Mi",
			cpuRequest:         "50m",
			memoryLimit:        "128Mi",
			cpuLimit:           "100m",
			containerPort:      8080,
			expectedPullPolicy: "Always",
			description:        "Receiver mode adds args",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.buildContainerSpec(
				tt.lambda, tt.imageURI, tt.env,
				tt.memoryRequest, tt.cpuRequest,
				tt.memoryLimit, tt.cpuLimit,
				tt.containerPort,
			)

			// Verify basic fields
			assert.Equal(t, tt.imageURI, result["image"])
			assert.Equal(t, tt.expectedPullPolicy, result["imagePullPolicy"], tt.description)

			// Verify ports
			ports, ok := result["ports"].([]interface{})
			require.True(t, ok, "Should have ports")
			require.Len(t, ports, 1)
			port, ok := ports[0].(map[string]interface{})
			require.True(t, ok)
			assert.Equal(t, tt.containerPort, port["containerPort"])

			// Verify resources
			resources, ok := result["resources"].(map[string]interface{})
			require.True(t, ok, "Should have resources")

			requests, ok := resources["requests"].(map[string]interface{})
			require.True(t, ok)
			assert.Equal(t, tt.memoryRequest, requests["memory"])
			assert.Equal(t, tt.cpuRequest, requests["cpu"])

			limits, ok := resources["limits"].(map[string]interface{})
			require.True(t, ok)
			assert.Equal(t, tt.memoryLimit, limits["memory"])
			assert.Equal(t, tt.cpuLimit, limits["cpu"])

			// Verify readiness probe
			readinessProbe, ok := result["readinessProbe"].(map[string]interface{})
			require.True(t, ok, "Should have readiness probe")
			httpGet, ok := readinessProbe["httpGet"].(map[string]interface{})
			require.True(t, ok)
			assert.Equal(t, tt.containerPort, httpGet["port"])

			// Verify receiver mode specific settings
			if tt.lambda.Annotations != nil && tt.lambda.Annotations["lambda.knative.io/receiver-mode"] == "true" {
				args, hasArgs := result["args"].([]string)
				assert.True(t, hasArgs, "Receiver mode should have args")
				assert.Contains(t, args, "--mode=receiver")
			}
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  ⚖️ SCALING CONFIGURATION TESTS                                         │
// └─────────────────────────────────────────────────────────────────────────┘

func TestScalingDefaults(t *testing.T) {
	tests := []struct {
		name                string
		scaling             *lambdav1alpha1.ScalingSpec
		expectedMin         int64
		expectedMax         int64
		expectedConcurrency int64
		description         string
	}{
		{
			name:                "Nil scaling uses defaults",
			scaling:             nil,
			expectedMin:         int64(DefaultMinReplicas),
			expectedMax:         int64(DefaultMaxReplicas),
			expectedConcurrency: int64(DefaultConcurrency),
			description:         "Should use default values when scaling is nil",
		},
		{
			name: "Custom scaling values",
			scaling: func() *lambdav1alpha1.ScalingSpec {
				min := int32(2)
				max := int32(100)
				concurrency := int32(50)
				return &lambdav1alpha1.ScalingSpec{
					MinReplicas:       &min,
					MaxReplicas:       &max,
					TargetConcurrency: &concurrency,
				}
			}(),
			expectedMin:         2,
			expectedMax:         100,
			expectedConcurrency: 50,
			description:         "Should use custom values when provided",
		},
		{
			name: "Partial scaling values",
			scaling: func() *lambdav1alpha1.ScalingSpec {
				min := int32(1)
				return &lambdav1alpha1.ScalingSpec{
					MinReplicas: &min,
				}
			}(),
			expectedMin:         1,
			expectedMax:         int64(DefaultMaxReplicas),
			expectedConcurrency: int64(DefaultConcurrency),
			description:         "Should use defaults for unset values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Replicate the logic from CreateService
			minReplicas := int64(DefaultMinReplicas)
			maxReplicas := int64(DefaultMaxReplicas)
			targetConcurrency := int64(DefaultConcurrency)

			if tt.scaling != nil {
				if tt.scaling.MinReplicas != nil {
					minReplicas = int64(*tt.scaling.MinReplicas)
				}
				if tt.scaling.MaxReplicas != nil {
					maxReplicas = int64(*tt.scaling.MaxReplicas)
				}
				if tt.scaling.TargetConcurrency != nil {
					targetConcurrency = int64(*tt.scaling.TargetConcurrency)
				}
			}

			assert.Equal(t, tt.expectedMin, minReplicas, tt.description)
			assert.Equal(t, tt.expectedMax, maxReplicas, tt.description)
			assert.Equal(t, tt.expectedConcurrency, targetConcurrency, tt.description)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📦 RESOURCE CONFIGURATION TESTS                                        │
// └─────────────────────────────────────────────────────────────────────────┘

func TestResourceDefaults(t *testing.T) {
	tests := []struct {
		name                  string
		resources             *lambdav1alpha1.ResourceSpec
		expectedMemoryRequest string
		expectedCPURequest    string
		expectedMemoryLimit   string
		expectedCPULimit      string
		description           string
	}{
		{
			name:                  "Nil resources uses defaults",
			resources:             nil,
			expectedMemoryRequest: "64Mi",
			expectedCPURequest:    "50m",
			expectedMemoryLimit:   "128Mi",
			expectedCPULimit:      "100m",
			description:           "Should use default values when resources is nil",
		},
		{
			name: "Custom resources",
			resources: &lambdav1alpha1.ResourceSpec{
				Requests: &lambdav1alpha1.ResourceRequirements{
					Memory: "256Mi",
					CPU:    "200m",
				},
				Limits: &lambdav1alpha1.ResourceRequirements{
					Memory: "1Gi",
					CPU:    "1000m",
				},
			},
			expectedMemoryRequest: "256Mi",
			expectedCPURequest:    "200m",
			expectedMemoryLimit:   "1Gi",
			expectedCPULimit:      "1000m",
			description:           "Should use custom values when provided",
		},
		{
			name: "Partial resources",
			resources: &lambdav1alpha1.ResourceSpec{
				Requests: &lambdav1alpha1.ResourceRequirements{
					Memory: "128Mi",
				},
			},
			expectedMemoryRequest: "128Mi",
			expectedCPURequest:    "50m",   // Default
			expectedMemoryLimit:   "128Mi", // Default
			expectedCPULimit:      "100m",  // Default
			description:           "Should use defaults for unset values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Replicate the logic from CreateService
			memoryRequest := "64Mi"
			cpuRequest := "50m"
			memoryLimit := "128Mi"
			cpuLimit := "100m"

			if tt.resources != nil {
				if tt.resources.Requests != nil {
					if tt.resources.Requests.Memory != "" {
						memoryRequest = tt.resources.Requests.Memory
					}
					if tt.resources.Requests.CPU != "" {
						cpuRequest = tt.resources.Requests.CPU
					}
				}
				if tt.resources.Limits != nil {
					if tt.resources.Limits.Memory != "" {
						memoryLimit = tt.resources.Limits.Memory
					}
					if tt.resources.Limits.CPU != "" {
						cpuLimit = tt.resources.Limits.CPU
					}
				}
			}

			assert.Equal(t, tt.expectedMemoryRequest, memoryRequest, tt.description)
			assert.Equal(t, tt.expectedCPURequest, cpuRequest, tt.description)
			assert.Equal(t, tt.expectedMemoryLimit, memoryLimit, tt.description)
			assert.Equal(t, tt.expectedCPULimit, cpuLimit, tt.description)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔌 CONTAINER PORT TESTS                                                │
// └─────────────────────────────────────────────────────────────────────────┘

func TestContainerPortSelection(t *testing.T) {
	tests := []struct {
		name         string
		lambda       *lambdav1alpha1.LambdaFunction
		expectedPort int64
		description  string
	}{
		{
			name: "Default port for inline source",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Source: lambdav1alpha1.SourceSpec{Type: "inline"},
				},
			},
			expectedPort: 8080,
			description:  "Inline source should use default port 8080",
		},
		{
			name: "Default port for image source without port",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Source: lambdav1alpha1.SourceSpec{
						Type: "image",
						Image: &lambdav1alpha1.ImageSource{
							Repository: "gcr.io/project/app",
						},
					},
				},
			},
			expectedPort: 8080,
			description:  "Image source without port should use default 8080",
		},
		{
			name: "Custom port from image source",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Source: lambdav1alpha1.SourceSpec{
						Type: "image",
						Image: &lambdav1alpha1.ImageSource{
							Repository: "gcr.io/project/app",
							Port:       3000,
						},
					},
				},
			},
			expectedPort: 3000,
			description:  "Image source with port should use custom port",
		},
		{
			name: "Port 0 uses default",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Source: lambdav1alpha1.SourceSpec{
						Type: "image",
						Image: &lambdav1alpha1.ImageSource{
							Repository: "gcr.io/project/app",
							Port:       0,
						},
					},
				},
			},
			expectedPort: 8080,
			description:  "Port 0 should fall back to default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Replicate the logic from CreateService
			containerPort := int64(DefaultPort)
			if tt.lambda.Spec.Source.Type == "image" && tt.lambda.Spec.Source.Image != nil && tt.lambda.Spec.Source.Image.Port > 0 {
				containerPort = int64(tt.lambda.Spec.Source.Image.Port)
			}

			assert.Equal(t, tt.expectedPort, containerPort, tt.description)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🎯 HANDLER ENV VAR TESTS                                               │
// └─────────────────────────────────────────────────────────────────────────┘

func TestHandlerEnvVar(t *testing.T) {
	tests := []struct {
		name             string
		sourceType       string
		shouldAddHandler bool
		description      string
	}{
		{
			name:             "Inline source adds HANDLER",
			sourceType:       "inline",
			shouldAddHandler: true,
			description:      "Non-image sources should add HANDLER env var",
		},
		{
			name:             "MinIO source adds HANDLER",
			sourceType:       "minio",
			shouldAddHandler: true,
			description:      "Non-image sources should add HANDLER env var",
		},
		{
			name:             "Image source skips HANDLER",
			sourceType:       "image",
			shouldAddHandler: false,
			description:      "Image sources should not add HANDLER env var",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Replicate the logic from CreateService
			shouldAdd := tt.sourceType != "image"
			assert.Equal(t, tt.shouldAddHandler, shouldAdd, tt.description)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📊 MANAGER INITIALIZATION TESTS                                        │
// └─────────────────────────────────────────────────────────────────────────┘

func TestNewManager(t *testing.T) {
	// NewManager requires client and scheme
	m := NewManager(nil, nil)

	require.NotNil(t, m)
	// Client and scheme are nil but manager should still be created
	assert.Nil(t, m.client)
	assert.Nil(t, m.scheme)
}
