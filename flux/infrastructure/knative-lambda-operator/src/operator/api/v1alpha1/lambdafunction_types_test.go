// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🧪 Unit Tests: LambdaFunction Types
//
//	Tests for API type helpers:
//	- SetCondition / GetCondition
//	- Status management
//	- Phase constants
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📋 STATUS CONDITION TESTS                                              │
// └─────────────────────────────────────────────────────────────────────────┘

func TestLambdaFunctionStatus_SetCondition(t *testing.T) {
	tests := []struct {
		name              string
		initialConditions []metav1.Condition
		newCondition      metav1.Condition
		expectedCount     int
		description       string
	}{
		{
			name:              "Add first condition",
			initialConditions: nil,
			newCondition: metav1.Condition{
				Type:    ConditionSourceReady,
				Status:  metav1.ConditionTrue,
				Reason:  "SourceReady",
				Message: "Source is ready",
			},
			expectedCount: 1,
			description:   "Should add condition to empty list",
		},
		{
			name: "Update existing condition",
			initialConditions: []metav1.Condition{
				{
					Type:    ConditionSourceReady,
					Status:  metav1.ConditionFalse,
					Reason:  "SourceNotReady",
					Message: "Source is not ready",
				},
			},
			newCondition: metav1.Condition{
				Type:    ConditionSourceReady,
				Status:  metav1.ConditionTrue,
				Reason:  "SourceReady",
				Message: "Source is now ready",
			},
			expectedCount: 1,
			description:   "Should update existing condition",
		},
		{
			name: "Add different condition type",
			initialConditions: []metav1.Condition{
				{
					Type:    ConditionSourceReady,
					Status:  metav1.ConditionTrue,
					Reason:  "SourceReady",
					Message: "Source is ready",
				},
			},
			newCondition: metav1.Condition{
				Type:    ConditionBuildReady,
				Status:  metav1.ConditionFalse,
				Reason:  "Building",
				Message: "Build in progress",
			},
			expectedCount: 2,
			description:   "Should add new condition type",
		},
		{
			name: "Multiple conditions with update",
			initialConditions: []metav1.Condition{
				{Type: ConditionSourceReady, Status: metav1.ConditionTrue, Reason: "Ready", Message: "Ready"},
				{Type: ConditionBuildReady, Status: metav1.ConditionFalse, Reason: "Building", Message: "Building"},
				{Type: ConditionDeployReady, Status: metav1.ConditionFalse, Reason: "Pending", Message: "Pending"},
			},
			newCondition: metav1.Condition{
				Type:    ConditionBuildReady,
				Status:  metav1.ConditionTrue,
				Reason:  "BuildComplete",
				Message: "Build completed successfully",
			},
			expectedCount: 3,
			description:   "Should update middle condition in list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			status := &LambdaFunctionStatus{
				Conditions: tt.initialConditions,
			}

			// Act
			status.SetCondition(tt.newCondition)

			// Assert
			assert.Len(t, status.Conditions, tt.expectedCount, tt.description)

			// Verify the condition was set correctly
			found := status.GetCondition(tt.newCondition.Type)
			require.NotNil(t, found, "Condition should exist after SetCondition")
			assert.Equal(t, tt.newCondition.Status, found.Status, "Condition status should match")
			assert.Equal(t, tt.newCondition.Reason, found.Reason, "Condition reason should match")
			assert.Equal(t, tt.newCondition.Message, found.Message, "Condition message should match")
		})
	}
}

func TestLambdaFunctionStatus_GetCondition(t *testing.T) {
	tests := []struct {
		name           string
		conditions     []metav1.Condition
		conditionType  string
		expectNil      bool
		expectedStatus metav1.ConditionStatus
		description    string
	}{
		{
			name:          "Get condition from empty list",
			conditions:    nil,
			conditionType: ConditionSourceReady,
			expectNil:     true,
			description:   "Should return nil for empty conditions",
		},
		{
			name: "Get existing condition",
			conditions: []metav1.Condition{
				{
					Type:   ConditionSourceReady,
					Status: metav1.ConditionTrue,
					Reason: "Ready",
				},
			},
			conditionType:  ConditionSourceReady,
			expectNil:      false,
			expectedStatus: metav1.ConditionTrue,
			description:    "Should return existing condition",
		},
		{
			name: "Get non-existing condition type",
			conditions: []metav1.Condition{
				{
					Type:   ConditionSourceReady,
					Status: metav1.ConditionTrue,
					Reason: "Ready",
				},
			},
			conditionType: ConditionBuildReady,
			expectNil:     true,
			description:   "Should return nil for non-existing type",
		},
		{
			name: "Get condition from multiple",
			conditions: []metav1.Condition{
				{Type: ConditionSourceReady, Status: metav1.ConditionTrue, Reason: "Ready"},
				{Type: ConditionBuildReady, Status: metav1.ConditionFalse, Reason: "Building"},
				{Type: ConditionDeployReady, Status: metav1.ConditionUnknown, Reason: "Unknown"},
			},
			conditionType:  ConditionBuildReady,
			expectNil:      false,
			expectedStatus: metav1.ConditionFalse,
			description:    "Should return correct condition from list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			status := &LambdaFunctionStatus{
				Conditions: tt.conditions,
			}

			// Act
			result := status.GetCondition(tt.conditionType)

			// Assert
			if tt.expectNil {
				assert.Nil(t, result, tt.description)
			} else {
				require.NotNil(t, result, tt.description)
				assert.Equal(t, tt.expectedStatus, result.Status, "Status should match")
			}
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📊 PHASE CONSTANTS TESTS                                               │
// └─────────────────────────────────────────────────────────────────────────┘

func TestLambdaPhase_Constants(t *testing.T) {
	// Verify all phase constants are defined correctly
	phases := map[LambdaPhase]string{
		PhasePending:   "Pending",
		PhaseBuilding:  "Building",
		PhaseDeploying: "Deploying",
		PhaseReady:     "Ready",
		PhaseFailed:    "Failed",
		PhaseDeleting:  "Deleting",
	}

	for phase, expected := range phases {
		t.Run(string(phase), func(t *testing.T) {
			assert.Equal(t, expected, string(phase), "Phase constant should have correct value")
		})
	}
}

func TestCondition_Constants(t *testing.T) {
	// Verify all condition type constants
	conditions := []string{
		ConditionSourceReady,
		ConditionBuildReady,
		ConditionEventingReady,
		ConditionDeployReady,
		ConditionServiceReady,
	}

	for _, cond := range conditions {
		t.Run(cond, func(t *testing.T) {
			assert.NotEmpty(t, cond, "Condition constant should not be empty")
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔧 SOURCE TYPE TESTS                                                   │
// └─────────────────────────────────────────────────────────────────────────┘

func TestSourceSpec_Validation(t *testing.T) {
	tests := []struct {
		name        string
		source      SourceSpec
		expectValid bool
		description string
	}{
		{
			name: "Valid MinIO source",
			source: SourceSpec{
				Type: "minio",
				MinIO: &MinIOSource{
					Bucket: "test-bucket",
					Key:    "test-key",
				},
			},
			expectValid: true,
			description: "MinIO source with bucket and key should be valid",
		},
		{
			name: "Valid S3 source",
			source: SourceSpec{
				Type: "s3",
				S3: &S3Source{
					Bucket: "test-bucket",
					Key:    "test-key",
				},
			},
			expectValid: true,
			description: "S3 source with bucket and key should be valid",
		},
		{
			name: "Valid Git source",
			source: SourceSpec{
				Type: "git",
				Git: &GitSource{
					URL: "https://github.com/example/repo",
				},
			},
			expectValid: true,
			description: "Git source with URL should be valid",
		},
		{
			name: "Valid Inline source",
			source: SourceSpec{
				Type: "inline",
				Inline: &InlineSource{
					Code: "def handler(event): return event",
				},
			},
			expectValid: true,
			description: "Inline source with code should be valid",
		},
		{
			name: "Valid Image source",
			source: SourceSpec{
				Type: "image",
				Image: &ImageSource{
					Repository: "gcr.io/project/image",
					Tag:        "latest",
				},
			},
			expectValid: true,
			description: "Image source with repository should be valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Assert non-nil source configuration based on type
			switch tt.source.Type {
			case "minio":
				if tt.expectValid {
					assert.NotNil(t, tt.source.MinIO, tt.description)
					assert.NotEmpty(t, tt.source.MinIO.Bucket)
					assert.NotEmpty(t, tt.source.MinIO.Key)
				}
			case "s3":
				if tt.expectValid {
					assert.NotNil(t, tt.source.S3, tt.description)
					assert.NotEmpty(t, tt.source.S3.Bucket)
					assert.NotEmpty(t, tt.source.S3.Key)
				}
			case "git":
				if tt.expectValid {
					assert.NotNil(t, tt.source.Git, tt.description)
					assert.NotEmpty(t, tt.source.Git.URL)
				}
			case "inline":
				if tt.expectValid {
					assert.NotNil(t, tt.source.Inline, tt.description)
					assert.NotEmpty(t, tt.source.Inline.Code)
				}
			case "image":
				if tt.expectValid {
					assert.NotNil(t, tt.source.Image, tt.description)
					assert.NotEmpty(t, tt.source.Image.Repository)
				}
			}
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🏗️ BUILD STATUS TESTS                                                  │
// └─────────────────────────────────────────────────────────────────────────┘

func TestBuildStatusInfo_Lifecycle(t *testing.T) {
	now := metav1.Now()

	tests := []struct {
		name        string
		buildStatus BuildStatusInfo
		isComplete  bool
		isSuccess   bool
		description string
	}{
		{
			name: "Build in progress",
			buildStatus: BuildStatusInfo{
				JobName:   "test-build-123",
				StartedAt: &now,
				Attempt:   1,
			},
			isComplete:  false,
			isSuccess:   false,
			description: "Build without CompletedAt is in progress",
		},
		{
			name: "Build completed successfully",
			buildStatus: BuildStatusInfo{
				JobName:     "test-build-123",
				ImageURI:    "localhost:5001/test/image:abc123",
				StartedAt:   &now,
				CompletedAt: &now,
				Attempt:     1,
			},
			isComplete:  true,
			isSuccess:   true,
			description: "Build with ImageURI and CompletedAt is successful",
		},
		{
			name: "Build failed",
			buildStatus: BuildStatusInfo{
				JobName:     "test-build-123",
				StartedAt:   &now,
				CompletedAt: &now,
				Error:       "Build failed: dockerfile parse error",
				Attempt:     2,
			},
			isComplete:  true,
			isSuccess:   false,
			description: "Build with Error is failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Assert completion status
			isComplete := tt.buildStatus.CompletedAt != nil
			assert.Equal(t, tt.isComplete, isComplete, "Completion status: %s", tt.description)

			// Assert success status
			isSuccess := tt.buildStatus.CompletedAt != nil && tt.buildStatus.Error == "" && tt.buildStatus.ImageURI != ""
			assert.Equal(t, tt.isSuccess, isSuccess, "Success status: %s", tt.description)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🌐 SERVICE STATUS TESTS                                                │
// └─────────────────────────────────────────────────────────────────────────┘

func TestServiceStatusInfo_Fields(t *testing.T) {
	serviceStatus := ServiceStatusInfo{
		ServiceName:    "test-lambda",
		URL:            "http://test-lambda.default.svc.cluster.local",
		Ready:          true,
		Replicas:       3,
		LatestRevision: "test-lambda-00001",
	}

	assert.Equal(t, "test-lambda", serviceStatus.ServiceName)
	assert.Equal(t, "http://test-lambda.default.svc.cluster.local", serviceStatus.URL)
	assert.True(t, serviceStatus.Ready)
	assert.Equal(t, int32(3), serviceStatus.Replicas)
	assert.Equal(t, "test-lambda-00001", serviceStatus.LatestRevision)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📦 LAMBDAFUNCTION OBJECT TESTS                                         │
// └─────────────────────────────────────────────────────────────────────────┘

func TestLambdaFunction_Initialization(t *testing.T) {
	lambda := &LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-function",
			Namespace: "default",
		},
		Spec: LambdaFunctionSpec{
			Source: SourceSpec{
				Type: "inline",
				Inline: &InlineSource{
					Code: "def handler(event): return {'status': 'ok'}",
				},
			},
			Runtime: RuntimeSpec{
				Language: "python",
				Version:  "3.11",
				Handler:  "handler",
			},
		},
	}

	assert.Equal(t, "test-function", lambda.Name)
	assert.Equal(t, "default", lambda.Namespace)
	assert.Equal(t, "inline", lambda.Spec.Source.Type)
	assert.Equal(t, "python", lambda.Spec.Runtime.Language)
	assert.Equal(t, "3.11", lambda.Spec.Runtime.Version)
}

func TestLambdaFunction_WithScaling(t *testing.T) {
	minReplicas := int32(1)
	maxReplicas := int32(100)
	targetConcurrency := int32(50)

	lambda := &LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "scaled-function",
			Namespace: "production",
		},
		Spec: LambdaFunctionSpec{
			Source: SourceSpec{
				Type: "inline",
				Inline: &InlineSource{
					Code: "module.exports.handler = async () => ({status: 'ok'})",
				},
			},
			Runtime: RuntimeSpec{
				Language: "nodejs",
				Version:  "20",
			},
			Scaling: &ScalingSpec{
				MinReplicas:       &minReplicas,
				MaxReplicas:       &maxReplicas,
				TargetConcurrency: &targetConcurrency,
			},
		},
	}

	require.NotNil(t, lambda.Spec.Scaling)
	assert.Equal(t, int32(1), *lambda.Spec.Scaling.MinReplicas)
	assert.Equal(t, int32(100), *lambda.Spec.Scaling.MaxReplicas)
	assert.Equal(t, int32(50), *lambda.Spec.Scaling.TargetConcurrency)
}

func TestLambdaFunction_WithResources(t *testing.T) {
	lambda := &LambdaFunction{
		Spec: LambdaFunctionSpec{
			Source: SourceSpec{
				Type: "inline",
				Inline: &InlineSource{
					Code: "package main",
				},
			},
			Runtime: RuntimeSpec{
				Language: "go",
				Version:  "1.21",
			},
			Resources: &ResourceSpec{
				Requests: &ResourceRequirements{
					Memory: "128Mi",
					CPU:    "100m",
				},
				Limits: &ResourceRequirements{
					Memory: "256Mi",
					CPU:    "500m",
				},
			},
		},
	}

	require.NotNil(t, lambda.Spec.Resources)
	require.NotNil(t, lambda.Spec.Resources.Requests)
	require.NotNil(t, lambda.Spec.Resources.Limits)
	assert.Equal(t, "128Mi", lambda.Spec.Resources.Requests.Memory)
	assert.Equal(t, "100m", lambda.Spec.Resources.Requests.CPU)
	assert.Equal(t, "256Mi", lambda.Spec.Resources.Limits.Memory)
	assert.Equal(t, "500m", lambda.Spec.Resources.Limits.CPU)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🎭 EVENTING SPEC TESTS                                                 │
// └─────────────────────────────────────────────────────────────────────────┘

func TestEventingSpec_Defaults(t *testing.T) {
	// Test that eventing is enabled by default (nil means enabled)
	lambda := &LambdaFunction{
		Spec: LambdaFunctionSpec{
			Source: SourceSpec{
				Type: "inline",
				Inline: &InlineSource{
					Code: "handler",
				},
			},
			Runtime: RuntimeSpec{
				Language: "python",
				Version:  "3.11",
			},
		},
	}

	// When Eventing is nil, it should be considered enabled
	eventingEnabled := lambda.Spec.Eventing == nil || lambda.Spec.Eventing.Enabled
	assert.True(t, eventingEnabled, "Eventing should be enabled by default when nil")
}

func TestEventingSpec_ExplicitDisabled(t *testing.T) {
	lambda := &LambdaFunction{
		Spec: LambdaFunctionSpec{
			Source: SourceSpec{
				Type: "inline",
				Inline: &InlineSource{
					Code: "handler",
				},
			},
			Runtime: RuntimeSpec{
				Language: "python",
				Version:  "3.11",
			},
			Eventing: &EventingSpec{
				Enabled: false,
			},
		},
	}

	eventingEnabled := lambda.Spec.Eventing == nil || lambda.Spec.Eventing.Enabled
	assert.False(t, eventingEnabled, "Eventing should be disabled when explicitly set to false")
}

func TestDLQSpec_Defaults(t *testing.T) {
	dlq := &DLQSpec{
		Enabled:          true,
		ExchangeName:     "lambda-dlq-exchange",
		QueueName:        "lambda-dlq-queue",
		RoutingKeyPrefix: "lambda.dlq",
		RetryMaxAttempts: 5,
		MessageTTL:       604800000, // 7 days
		MaxLength:        50000,
		OverflowPolicy:   "reject-publish",
	}

	assert.True(t, dlq.Enabled)
	assert.Equal(t, "lambda-dlq-exchange", dlq.ExchangeName)
	assert.Equal(t, "lambda-dlq-queue", dlq.QueueName)
	assert.Equal(t, 5, dlq.RetryMaxAttempts)
	assert.Equal(t, 604800000, dlq.MessageTTL)
	assert.Equal(t, 50000, dlq.MaxLength)
	assert.Equal(t, "reject-publish", dlq.OverflowPolicy)
}
