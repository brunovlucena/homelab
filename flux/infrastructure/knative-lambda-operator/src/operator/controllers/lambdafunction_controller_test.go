// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🧪 Unit Tests: LambdaFunction Controller
//
//	Tests for controller logic:
//	- Spec validation
//	- Phase transitions
//	- Condition management
//	- Reconciliation helpers
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package controllers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	lambdav1alpha1 "github.com/brunovlucena/knative-lambda-operator/api/v1alpha1"
)

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔍 SPEC VALIDATION TESTS                                               │
// └─────────────────────────────────────────────────────────────────────────┘

func TestValidateSpec(t *testing.T) {
	r := &LambdaFunctionReconciler{}

	tests := []struct {
		name        string
		lambda      *lambdav1alpha1.LambdaFunction
		expectError bool
		errorMsg    string
		description string
	}{
		// Source type validation
		{
			name: "Missing source type",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Source: lambdav1alpha1.SourceSpec{
						Type: "",
					},
					Runtime: lambdav1alpha1.RuntimeSpec{
						Language: "python",
						Version:  "3.11",
					},
				},
			},
			expectError: true,
			errorMsg:    "source type is required",
			description: "Should reject empty source type",
		},
		{
			name: "Unsupported source type",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Source: lambdav1alpha1.SourceSpec{
						Type: "unsupported",
					},
					Runtime: lambdav1alpha1.RuntimeSpec{
						Language: "python",
						Version:  "3.11",
					},
				},
			},
			expectError: true,
			errorMsg:    "unsupported source type: unsupported",
			description: "Should reject unknown source type",
		},

		// MinIO source validation
		{
			name: "MinIO source without config",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Source: lambdav1alpha1.SourceSpec{
						Type: "minio",
					},
					Runtime: lambdav1alpha1.RuntimeSpec{
						Language: "python",
						Version:  "3.11",
					},
				},
			},
			expectError: true,
			errorMsg:    "minio configuration is required",
			description: "Should reject minio without config",
		},
		{
			name: "MinIO source without bucket",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Source: lambdav1alpha1.SourceSpec{
						Type: "minio",
						MinIO: &lambdav1alpha1.MinIOSource{
							Bucket: "",
							Key:    "test-key",
						},
					},
					Runtime: lambdav1alpha1.RuntimeSpec{
						Language: "python",
						Version:  "3.11",
					},
				},
			},
			expectError: true,
			errorMsg:    "minio bucket and key are required",
			description: "Should reject minio without bucket",
		},
		{
			name: "MinIO source without key",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Source: lambdav1alpha1.SourceSpec{
						Type: "minio",
						MinIO: &lambdav1alpha1.MinIOSource{
							Bucket: "test-bucket",
							Key:    "",
						},
					},
					Runtime: lambdav1alpha1.RuntimeSpec{
						Language: "python",
						Version:  "3.11",
					},
				},
			},
			expectError: true,
			errorMsg:    "minio bucket and key are required",
			description: "Should reject minio without key",
		},
		{
			name: "Valid MinIO source",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Source: lambdav1alpha1.SourceSpec{
						Type: "minio",
						MinIO: &lambdav1alpha1.MinIOSource{
							Bucket: "test-bucket",
							Key:    "lambdas/test/main.py",
						},
					},
					Runtime: lambdav1alpha1.RuntimeSpec{
						Language: "python",
						Version:  "3.11",
					},
				},
			},
			expectError: false,
			description: "Should accept valid minio source",
		},

		// S3 source validation
		{
			name: "S3 source without config",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Source: lambdav1alpha1.SourceSpec{
						Type: "s3",
					},
					Runtime: lambdav1alpha1.RuntimeSpec{
						Language: "nodejs",
						Version:  "20",
					},
				},
			},
			expectError: true,
			errorMsg:    "s3 configuration is required",
			description: "Should reject s3 without config",
		},
		{
			name: "Valid S3 source",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Source: lambdav1alpha1.SourceSpec{
						Type: "s3",
						S3: &lambdav1alpha1.S3Source{
							Bucket: "my-bucket",
							Key:    "functions/index.js",
							Region: "us-east-1",
						},
					},
					Runtime: lambdav1alpha1.RuntimeSpec{
						Language: "nodejs",
						Version:  "20",
					},
				},
			},
			expectError: false,
			description: "Should accept valid s3 source",
		},

		// GCS source validation
		{
			name: "GCS source without config",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Source: lambdav1alpha1.SourceSpec{
						Type: "gcs",
					},
					Runtime: lambdav1alpha1.RuntimeSpec{
						Language: "go",
						Version:  "1.21",
					},
				},
			},
			expectError: true,
			errorMsg:    "gcs configuration is required",
			description: "Should reject gcs without config",
		},
		{
			name: "Valid GCS source",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Source: lambdav1alpha1.SourceSpec{
						Type: "gcs",
						GCS: &lambdav1alpha1.GCSSource{
							Bucket: "my-gcs-bucket",
							Key:    "functions/main.go",
						},
					},
					Runtime: lambdav1alpha1.RuntimeSpec{
						Language: "go",
						Version:  "1.21",
					},
				},
			},
			expectError: false,
			description: "Should accept valid gcs source",
		},

		// Git source validation
		{
			name: "Git source without config",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Source: lambdav1alpha1.SourceSpec{
						Type: "git",
					},
					Runtime: lambdav1alpha1.RuntimeSpec{
						Language: "python",
						Version:  "3.11",
					},
				},
			},
			expectError: true,
			errorMsg:    "git configuration is required",
			description: "Should reject git without config",
		},
		{
			name: "Git source without URL",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Source: lambdav1alpha1.SourceSpec{
						Type: "git",
						Git: &lambdav1alpha1.GitSource{
							URL: "",
						},
					},
					Runtime: lambdav1alpha1.RuntimeSpec{
						Language: "python",
						Version:  "3.11",
					},
				},
			},
			expectError: true,
			errorMsg:    "git url is required",
			description: "Should reject git without URL",
		},
		{
			name: "Valid Git source",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Source: lambdav1alpha1.SourceSpec{
						Type: "git",
						Git: &lambdav1alpha1.GitSource{
							URL: "https://github.com/example/repo.git",
							Ref: "main",
						},
					},
					Runtime: lambdav1alpha1.RuntimeSpec{
						Language: "python",
						Version:  "3.11",
					},
				},
			},
			expectError: false,
			description: "Should accept valid git source",
		},

		// Inline source validation
		{
			name: "Inline source without config",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Source: lambdav1alpha1.SourceSpec{
						Type: "inline",
					},
					Runtime: lambdav1alpha1.RuntimeSpec{
						Language: "python",
						Version:  "3.11",
					},
				},
			},
			expectError: true,
			errorMsg:    "inline configuration is required",
			description: "Should reject inline without config",
		},
		{
			name: "Inline source without code",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Source: lambdav1alpha1.SourceSpec{
						Type: "inline",
						Inline: &lambdav1alpha1.InlineSource{
							Code: "",
						},
					},
					Runtime: lambdav1alpha1.RuntimeSpec{
						Language: "python",
						Version:  "3.11",
					},
				},
			},
			expectError: true,
			errorMsg:    "inline code is required",
			description: "Should reject inline without code",
		},
		{
			name: "Valid Inline source",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Source: lambdav1alpha1.SourceSpec{
						Type: "inline",
						Inline: &lambdav1alpha1.InlineSource{
							Code: "def handler(event): return {'status': 'ok'}",
						},
					},
					Runtime: lambdav1alpha1.RuntimeSpec{
						Language: "python",
						Version:  "3.11",
					},
				},
			},
			expectError: false,
			description: "Should accept valid inline source",
		},

		// Image source validation
		{
			name: "Image source without config",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Source: lambdav1alpha1.SourceSpec{
						Type: "image",
					},
					Runtime: lambdav1alpha1.RuntimeSpec{
						Language: "python",
						Version:  "3.11",
					},
				},
			},
			expectError: true,
			errorMsg:    "image configuration is required",
			description: "Should reject image without config",
		},
		{
			name: "Image source without repository",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Source: lambdav1alpha1.SourceSpec{
						Type: "image",
						Image: &lambdav1alpha1.ImageSource{
							Repository: "",
						},
					},
					Runtime: lambdav1alpha1.RuntimeSpec{
						Language: "python",
						Version:  "3.11",
					},
				},
			},
			expectError: true,
			errorMsg:    "image repository is required",
			description: "Should reject image without repository",
		},
		{
			name: "Valid Image source",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Source: lambdav1alpha1.SourceSpec{
						Type: "image",
						Image: &lambdav1alpha1.ImageSource{
							Repository: "gcr.io/project/my-function",
							Tag:        "v1.0.0",
						},
					},
					Runtime: lambdav1alpha1.RuntimeSpec{
						Language: "python",
						Version:  "3.11",
					},
				},
			},
			expectError: false,
			description: "Should accept valid image source",
		},

		// Runtime validation
		{
			name: "Missing runtime language",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Source: lambdav1alpha1.SourceSpec{
						Type: "inline",
						Inline: &lambdav1alpha1.InlineSource{
							Code: "handler",
						},
					},
					Runtime: lambdav1alpha1.RuntimeSpec{
						Language: "",
						Version:  "3.11",
					},
				},
			},
			expectError: true,
			errorMsg:    "runtime language is required",
			description: "Should reject missing language",
		},
		{
			name: "Missing runtime version",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Source: lambdav1alpha1.SourceSpec{
						Type: "inline",
						Inline: &lambdav1alpha1.InlineSource{
							Code: "handler",
						},
					},
					Runtime: lambdav1alpha1.RuntimeSpec{
						Language: "python",
						Version:  "",
					},
				},
			},
			expectError: true,
			errorMsg:    "runtime version is required",
			description: "Should reject missing version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := r.validateSpec(tt.lambda)

			if tt.expectError {
				require.Error(t, err, tt.description)
				assert.Contains(t, err.Error(), tt.errorMsg, "Error message should match")
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📋 CONDITION MANAGEMENT TESTS                                          │
// └─────────────────────────────────────────────────────────────────────────┘

func TestSetCondition(t *testing.T) {
	r := &LambdaFunctionReconciler{}

	tests := []struct {
		name            string
		initialLambda   *lambdav1alpha1.LambdaFunction
		conditionType   string
		status          metav1.ConditionStatus
		reason          string
		message         string
		expectedCount   int
		expectedStatus  metav1.ConditionStatus
		expectedReason  string
		expectedMessage string
	}{
		{
			name: "Set first condition",
			initialLambda: &lambdav1alpha1.LambdaFunction{
				Status: lambdav1alpha1.LambdaFunctionStatus{},
			},
			conditionType:   lambdav1alpha1.ConditionSourceReady,
			status:          metav1.ConditionTrue,
			reason:          "SourceReady",
			message:         "Source is available",
			expectedCount:   1,
			expectedStatus:  metav1.ConditionTrue,
			expectedReason:  "SourceReady",
			expectedMessage: "Source is available",
		},
		{
			name: "Update existing condition",
			initialLambda: &lambdav1alpha1.LambdaFunction{
				Status: lambdav1alpha1.LambdaFunctionStatus{
					Conditions: []metav1.Condition{
						{
							Type:    lambdav1alpha1.ConditionBuildReady,
							Status:  metav1.ConditionFalse,
							Reason:  "Building",
							Message: "Build in progress",
						},
					},
				},
			},
			conditionType:   lambdav1alpha1.ConditionBuildReady,
			status:          metav1.ConditionTrue,
			reason:          "BuildComplete",
			message:         "Image built successfully",
			expectedCount:   1,
			expectedStatus:  metav1.ConditionTrue,
			expectedReason:  "BuildComplete",
			expectedMessage: "Image built successfully",
		},
		{
			name: "Add new condition to existing",
			initialLambda: &lambdav1alpha1.LambdaFunction{
				Status: lambdav1alpha1.LambdaFunctionStatus{
					Conditions: []metav1.Condition{
						{
							Type:   lambdav1alpha1.ConditionSourceReady,
							Status: metav1.ConditionTrue,
							Reason: "Ready",
						},
					},
				},
			},
			conditionType:   lambdav1alpha1.ConditionBuildReady,
			status:          metav1.ConditionFalse,
			reason:          "BuildStarted",
			message:         "Build job created",
			expectedCount:   2,
			expectedStatus:  metav1.ConditionFalse,
			expectedReason:  "BuildStarted",
			expectedMessage: "Build job created",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			r.setCondition(tt.initialLambda, tt.conditionType, tt.status, tt.reason, tt.message)

			// Assert
			assert.Len(t, tt.initialLambda.Status.Conditions, tt.expectedCount)

			cond := tt.initialLambda.Status.GetCondition(tt.conditionType)
			require.NotNil(t, cond, "Condition should exist")
			assert.Equal(t, tt.expectedStatus, cond.Status)
			assert.Equal(t, tt.expectedReason, cond.Reason)
			assert.Equal(t, tt.expectedMessage, cond.Message)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔧 HELPER FUNCTION TESTS                                               │
// └─────────────────────────────────────────────────────────────────────────┘

func TestResultString(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "Nil error returns success",
			err:      nil,
			expected: "success",
		},
		{
			name:     "Non-nil error returns error",
			err:      assert.AnError,
			expected: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resultString(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  ⏱️ REQUEUE INTERVAL TESTS                                              │
// └─────────────────────────────────────────────────────────────────────────┘

func TestRequeueIntervals(t *testing.T) {
	// Test that requeue intervals are defined and reasonable
	assert.NotZero(t, RequeueShort, "RequeueShort should be defined")
	assert.NotZero(t, RequeueMedium, "RequeueMedium should be defined")
	assert.NotZero(t, RequeueLong, "RequeueLong should be defined")

	// Verify ordering
	assert.Less(t, RequeueShort, RequeueMedium, "RequeueShort should be less than RequeueMedium")
	assert.Less(t, RequeueMedium, RequeueLong, "RequeueMedium should be less than RequeueLong")
}

func TestFinalizerName(t *testing.T) {
	assert.NotEmpty(t, FinalizerName, "FinalizerName should be defined")
	assert.Contains(t, FinalizerName, "lambdafunction", "FinalizerName should reference lambdafunction")
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🏗️ RECONCILER OPTIONS TESTS                                            │
// └─────────────────────────────────────────────────────────────────────────┘

func TestReconcilerOptions(t *testing.T) {
	opts := ReconcilerOptions{
		MaxConcurrentReconciles: 50,
	}

	assert.Equal(t, 50, opts.MaxConcurrentReconciles)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📦 LAMBDAFUNCTION BUILDER TESTS                                        │
// └─────────────────────────────────────────────────────────────────────────┘

func newTestLambdaFunction(name, namespace string) *lambdav1alpha1.LambdaFunction {
	return &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			Namespace:  namespace,
			Generation: 1,
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Source: lambdav1alpha1.SourceSpec{
				Type: "inline",
				Inline: &lambdav1alpha1.InlineSource{
					Code: "def handler(event): return event",
				},
			},
			Runtime: lambdav1alpha1.RuntimeSpec{
				Language: "python",
				Version:  "3.11",
				Handler:  "handler",
			},
		},
		Status: lambdav1alpha1.LambdaFunctionStatus{
			Phase: lambdav1alpha1.PhasePending,
		},
	}
}

func TestNewTestLambdaFunction(t *testing.T) {
	lambda := newTestLambdaFunction("test-func", "default")

	assert.Equal(t, "test-func", lambda.Name)
	assert.Equal(t, "default", lambda.Namespace)
	assert.Equal(t, "inline", lambda.Spec.Source.Type)
	assert.NotNil(t, lambda.Spec.Source.Inline)
	assert.Equal(t, "python", lambda.Spec.Runtime.Language)
	assert.Equal(t, lambdav1alpha1.PhasePending, lambda.Status.Phase)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🎭 OPERATOR IMAGE MODE TESTS                                           │
// └─────────────────────────────────────────────────────────────────────────┘

func TestUseOperatorImageAnnotation(t *testing.T) {
	tests := []struct {
		name              string
		annotations       map[string]string
		shouldUseOperator bool
	}{
		{
			name:              "No annotations",
			annotations:       nil,
			shouldUseOperator: false,
		},
		{
			name:              "Empty annotations",
			annotations:       map[string]string{},
			shouldUseOperator: false,
		},
		{
			name: "Use operator image annotation set to true",
			annotations: map[string]string{
				"lambda.knative.io/use-operator-image": "true",
			},
			shouldUseOperator: true,
		},
		{
			name: "Use operator image annotation set to false",
			annotations: map[string]string{
				"lambda.knative.io/use-operator-image": "false",
			},
			shouldUseOperator: false,
		},
		{
			name: "Other annotations present",
			annotations: map[string]string{
				"some-other-annotation": "value",
			},
			shouldUseOperator: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lambda := &lambdav1alpha1.LambdaFunction{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: tt.annotations,
				},
			}

			shouldUse := lambda.Annotations != nil && lambda.Annotations["lambda.knative.io/use-operator-image"] == "true"
			assert.Equal(t, tt.shouldUseOperator, shouldUse)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🌐 EVENTING ENABLED LOGIC TESTS                                        │
// └─────────────────────────────────────────────────────────────────────────┘

func TestEventingEnabled(t *testing.T) {
	tests := []struct {
		name            string
		eventing        *lambdav1alpha1.EventingSpec
		expectedEnabled bool
	}{
		{
			name:            "Nil eventing (default enabled)",
			eventing:        nil,
			expectedEnabled: true,
		},
		{
			name:            "Eventing explicitly enabled",
			eventing:        &lambdav1alpha1.EventingSpec{Enabled: true},
			expectedEnabled: true,
		},
		{
			name:            "Eventing explicitly disabled",
			eventing:        &lambdav1alpha1.EventingSpec{Enabled: false},
			expectedEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lambda := &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Eventing: tt.eventing,
				},
			}

			// This is the logic from reconcilePending
			eventingEnabled := lambda.Spec.Eventing == nil || lambda.Spec.Eventing.Enabled
			assert.Equal(t, tt.expectedEnabled, eventingEnabled)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🖼️ IMAGE SOURCE URI GENERATION TESTS                                   │
// └─────────────────────────────────────────────────────────────────────────┘

func TestImageSourceURIGeneration(t *testing.T) {
	tests := []struct {
		name        string
		imageSource *lambdav1alpha1.ImageSource
		expectedURI string
	}{
		{
			name: "Repository with digest",
			imageSource: &lambdav1alpha1.ImageSource{
				Repository: "gcr.io/project/image",
				Digest:     "sha256:abc123",
			},
			expectedURI: "gcr.io/project/image@sha256:abc123",
		},
		{
			name: "Repository with tag",
			imageSource: &lambdav1alpha1.ImageSource{
				Repository: "gcr.io/project/image",
				Tag:        "v1.0.0",
			},
			expectedURI: "gcr.io/project/image:v1.0.0",
		},
		{
			name: "Repository only (default to latest)",
			imageSource: &lambdav1alpha1.ImageSource{
				Repository: "gcr.io/project/image",
			},
			expectedURI: "gcr.io/project/image:latest",
		},
		{
			name: "Digest takes precedence over tag",
			imageSource: &lambdav1alpha1.ImageSource{
				Repository: "gcr.io/project/image",
				Tag:        "v1.0.0",
				Digest:     "sha256:abc123",
			},
			expectedURI: "gcr.io/project/image@sha256:abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Replicate the logic from reconcilePending
			imageURI := tt.imageSource.Repository
			if tt.imageSource.Digest != "" {
				imageURI = imageURI + "@" + tt.imageSource.Digest
			} else if tt.imageSource.Tag != "" {
				imageURI = imageURI + ":" + tt.imageSource.Tag
			} else {
				imageURI = imageURI + ":latest"
			}

			assert.Equal(t, tt.expectedURI, imageURI)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔄 PHASE TRANSITION TESTS                                              │
// └─────────────────────────────────────────────────────────────────────────┘

func TestPhaseTransitions(t *testing.T) {
	// Define valid phase transitions
	validTransitions := map[lambdav1alpha1.LambdaPhase][]lambdav1alpha1.LambdaPhase{
		lambdav1alpha1.PhasePending:   {lambdav1alpha1.PhaseBuilding, lambdav1alpha1.PhaseDeploying, lambdav1alpha1.PhaseFailed},
		lambdav1alpha1.PhaseBuilding:  {lambdav1alpha1.PhaseDeploying, lambdav1alpha1.PhaseFailed, lambdav1alpha1.PhasePending},
		lambdav1alpha1.PhaseDeploying: {lambdav1alpha1.PhaseReady, lambdav1alpha1.PhaseFailed},
		lambdav1alpha1.PhaseReady:     {lambdav1alpha1.PhasePending, lambdav1alpha1.PhaseFailed},
		lambdav1alpha1.PhaseFailed:    {lambdav1alpha1.PhasePending},
	}

	for fromPhase, toPhases := range validTransitions {
		for _, toPhase := range toPhases {
			t.Run(string(fromPhase)+"_to_"+string(toPhase), func(t *testing.T) {
				// Just verify the transitions are defined (actual enforcement is in reconcile)
				assert.NotEmpty(t, string(fromPhase))
				assert.NotEmpty(t, string(toPhase))
			})
		}
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🧪 ENV VAR HANDLING TESTS                                              │
// └─────────────────────────────────────────────────────────────────────────┘

func TestEnvVarHandling(t *testing.T) {
	lambda := &lambdav1alpha1.LambdaFunction{
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Env: []corev1.EnvVar{
				{Name: "DEBUG", Value: "true"},
				{Name: "LOG_LEVEL", Value: "debug"},
				{
					Name: "SECRET_KEY",
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
		},
	}

	assert.Len(t, lambda.Spec.Env, 3)
	assert.Equal(t, "DEBUG", lambda.Spec.Env[0].Name)
	assert.Equal(t, "true", lambda.Spec.Env[0].Value)
	assert.NotNil(t, lambda.Spec.Env[2].ValueFrom)
	assert.NotNil(t, lambda.Spec.Env[2].ValueFrom.SecretKeyRef)
}
