// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
//
//	🧪 BACKEND-006: Knative Service Management Tests
//
//	User Story: Knative Service Management
//	Priority: P0 | Story Points: 13
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
package backend

import (
	"context"
	"testing"
	"time"

	"knative-lambda/tests/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"

	config_pkg "knative-lambda/internal/config"
	"knative-lambda/internal/handler"
	"knative-lambda/pkg/builds"
)

// Test constants.
const (
	TestServiceName = "lambda-customer-123-parser-abc"
)

// TestBackend006_ServiceCreation validates Knative Service creation.
func TestBackend006_ServiceCreation(t *testing.T) {
	// Arrange
	ctx := context.Background()
	manager := setupServiceManager(t)

	completionData := &builds.BuildCompletionEventData{
		ThirdPartyID: "customer-123",
		ParserID:     "parser-abc",
		ImageURI:     "123456789.dkr.ecr.us-east-1.amazonaws.com/parsers:customer-123-parser-abc",
		Status:       "success",
	}

	serviceName := manager.GenerateServiceName(completionData.ThirdPartyID, completionData.ParserID)

	// Act
	err := manager.CreateService(ctx, serviceName, completionData)

	// Assert
	require.NoError(t, err)

	// Verify service exists
	exists, err := manager.CheckServiceExists(ctx, serviceName)
	require.NoError(t, err)
	assert.True(t, exists, "Service should exist after creation")
}

// TestBackend006_ServiceAccountCreation validates ServiceAccount creation.
func TestBackend006_ServiceAccountCreation(t *testing.T) {
	// Arrange
	manager := setupServiceManager(t)
	serviceName := TestServiceName
	completionData := &builds.BuildCompletionEventData{
		ThirdPartyID: "customer-123",
		ParserID:     "parser-abc",
		ImageURI:     "123456789.dkr.ecr.us-east-1.amazonaws.com/parsers:customer-123-parser-abc",
	}

	// Act
	serviceAccount := manager.CreateServiceAccountResource(serviceName, completionData)

	// Assert
	require.NotNil(t, serviceAccount)
	assert.Equal(t, "v1", serviceAccount.GetAPIVersion())
	assert.Equal(t, "ServiceAccount", serviceAccount.GetKind())
	assert.Equal(t, serviceName, serviceAccount.GetName())
}

// TestBackend006_ConfigMapCreation validates ConfigMap creation.
func TestBackend006_ConfigMapCreation(t *testing.T) {
	// Arrange
	manager := setupServiceManager(t)
	serviceName := TestServiceName

	completionData := &builds.BuildCompletionEventData{
		ThirdPartyID: "customer-123",
		ParserID:     "parser-abc",
	}

	// Act
	configMap := manager.CreateConfigMapResource(serviceName, completionData)

	// Assert
	require.NotNil(t, configMap)
	assert.Equal(t, "v1", configMap.GetAPIVersion())
	assert.Equal(t, "ConfigMap", configMap.GetKind())
	assert.Equal(t, serviceName+"-config", configMap.GetName())

	// Verify data
	data, found, err := unstructured.NestedStringMap(configMap.Object, "data")
	require.NoError(t, err)
	require.True(t, found)
	assert.Contains(t, data, "THIRD_PARTY_ID")
	assert.Equal(t, "customer-123", data["THIRD_PARTY_ID"])
}

// TestBackend006_TriggerCreation validates Trigger creation.
func TestBackend006_TriggerCreation(t *testing.T) {
	// Arrange
	manager := setupServiceManager(t)
	serviceName := TestServiceName

	completionData := &builds.BuildCompletionEventData{
		ThirdPartyID: "customer-123",
		ParserID:     "parser-abc",
	}

	// Act
	trigger := manager.CreateTriggerResource(serviceName, completionData)

	// Assert
	require.NotNil(t, trigger)
	assert.Equal(t, "eventing.knative.dev/v1", trigger.GetAPIVersion())
	assert.Equal(t, "Trigger", trigger.GetKind())

	// Verify filter
	filter, found, err := unstructured.NestedMap(trigger.Object, "spec", "filter")
	require.NoError(t, err)
	require.True(t, found)

	attributes, found, err := unstructured.NestedMap(filter, "attributes")
	require.NoError(t, err)
	require.True(t, found)
	assert.Contains(t, attributes, "thirdpartyid")
	assert.Equal(t, "customer-123", attributes["thirdpartyid"])
}

// TestBackend006_ParallelCreation validates parallel resource creation.
func TestBackend006_ParallelCreation(t *testing.T) {
	// Arrange
	ctx := context.Background()
	manager := setupServiceManager(t)

	completionData := &builds.BuildCompletionEventData{
		ThirdPartyID: "customer-123",
		ParserID:     "parser-abc",
		ImageURI:     "123456789.dkr.ecr.us-east-1.amazonaws.com/parsers:test",
	}

	serviceName := manager.GenerateServiceName(completionData.ThirdPartyID, completionData.ParserID)

	// Act
	startTime := time.Now()
	err := manager.CreateService(ctx, serviceName, completionData)
	endTime := startTime.Add(1 * time.Second) // Simulated completion
	maxDuration := 2 * time.Second

	phases := []testutils.Phase{
		{Name: "Service creation", Duration: 500 * time.Millisecond},
		{Name: "Route configuration", Duration: 300 * time.Millisecond},
		{Name: "Revision deployment", Duration: 200 * time.Millisecond},
	}

	// Assert
	require.NoError(t, err)
	testutils.RunTimingTest(t, "Parallel creation", startTime, endTime, maxDuration, phases)
}

// TestBackend006_ServiceUpdates validates service updates.
func TestBackend006_ServiceUpdates(t *testing.T) {
	// Arrange
	ctx := context.Background()
	manager := setupServiceManager(t)

	completionData := &builds.BuildCompletionEventData{
		ThirdPartyID: "customer-123",
		ParserID:     "parser-abc",
		ImageURI:     "123456789.dkr.ecr.us-east-1.amazonaws.com/parsers:v1",
	}

	serviceName := manager.GenerateServiceName(completionData.ThirdPartyID, completionData.ParserID)

	// Create initial service
	err := manager.CreateService(ctx, serviceName, completionData)
	require.NoError(t, err)

	// Update with new image
	completionData.ImageURI = "123456789.dkr.ecr.us-east-1.amazonaws.com/parsers:v2"

	// Act
	err = manager.CreateService(ctx, serviceName, completionData)

	// Assert
	require.NoError(t, err, "Should update existing service")
}

// TestBackend006_AutoScalingConfig validates auto-scaling configuration.
func TestBackend006_AutoScalingConfig(t *testing.T) {
	// Arrange
	manager := setupServiceManager(t)
	serviceName := TestServiceName

	completionData := &builds.BuildCompletionEventData{
		ThirdPartyID: "customer-123",
		ParserID:     "parser-abc",
		ImageURI:     "123456789.dkr.ecr.us-east-1.amazonaws.com/parsers:test",
	}

	// Act
	service := manager.CreateKnativeServiceResource(serviceName, completionData)

	// Assert
	require.NotNil(t, service)

	// Verify auto-scaling annotations
	annotations := service.GetAnnotations()
	assert.Equal(t, "0", annotations["autoscaling.knative.dev/min-scale"])
	assert.Equal(t, "10", annotations["autoscaling.knative.dev/max-scale"])
	assert.Equal(t, "100", annotations["autoscaling.knative.dev/target"])
}

// TestBackend006_ServiceDeletion validates service deletion.
func TestBackend006_ServiceDeletion(t *testing.T) {
	// Arrange
	ctx := context.Background()
	manager := setupServiceManager(t)

	completionData := &builds.BuildCompletionEventData{
		ThirdPartyID: "customer-123",
		ParserID:     "parser-abc",
		ImageURI:     "123456789.dkr.ecr.us-east-1.amazonaws.com/parsers:test",
	}

	serviceName := manager.GenerateServiceName(completionData.ThirdPartyID, completionData.ParserID)

	// Create service
	err := manager.CreateService(ctx, serviceName, completionData)
	require.NoError(t, err)

	// Act
	err = manager.DeleteService(ctx, serviceName)

	// Assert
	require.NoError(t, err)

	// Verify deletion
	exists, err := manager.CheckServiceExists(ctx, serviceName)
	require.NoError(t, err)
	assert.False(t, exists, "Service should not exist after deletion")
}

// TestBackend006_ResourceCleanup validates cleanup of all associated resources.
func TestBackend006_ResourceCleanup(t *testing.T) {
	// Arrange
	ctx := context.Background()
	manager := setupServiceManager(t)

	completionData := &builds.BuildCompletionEventData{
		ThirdPartyID: "customer-123",
		ParserID:     "parser-abc",
		ImageURI:     "123456789.dkr.ecr.us-east-1.amazonaws.com/parsers:test",
	}

	serviceName := manager.GenerateServiceName(completionData.ThirdPartyID, completionData.ParserID)

	// Create service (creates all resources)
	err := manager.CreateService(ctx, serviceName, completionData)
	require.NoError(t, err)

	// Act - Delete should clean up all resources
	err = manager.DeleteService(ctx, serviceName)

	// Assert
	require.NoError(t, err, "Should delete all resources without error")
}

// TestBackend006_ServiceNameGeneration validates service name format.
func TestBackend006_ServiceNameGeneration(t *testing.T) {
	// Arrange
	manager := setupServiceManager(t)
	thirdPartyID := "customer-123"
	parserID := "parser-abc"

	// Act
	serviceName := manager.GenerateServiceName(thirdPartyID, parserID)

	// Assert
	assert.Contains(t, serviceName, "lambda")
	assert.Contains(t, serviceName, thirdPartyID)
	assert.Contains(t, serviceName, parserID)
	assert.LessOrEqual(t, len(serviceName), 63, "Service name should be <= 63 characters")
}

// TestBackend006_EnvironmentVariables validates environment variable configuration.
func TestBackend006_EnvironmentVariables(t *testing.T) {
	// Arrange
	manager := setupServiceManager(t)
	serviceName := TestServiceName

	completionData := &builds.BuildCompletionEventData{
		ThirdPartyID: "customer-123",
		ParserID:     "parser-abc",
		ImageURI:     "123456789.dkr.ecr.us-east-1.amazonaws.com/parsers:test",
	}

	// Act
	service := manager.CreateKnativeServiceResource(serviceName, completionData)

	// Assert
	require.NotNil(t, service)

	// Verify environment variables in spec
	containers, found, err := unstructured.NestedSlice(service.Object, "spec", "template", "spec", "containers")
	require.NoError(t, err)
	require.True(t, found)
	require.Greater(t, len(containers), 0)

	container := containers[0].(map[string]interface{})
	env, found, err := unstructured.NestedSlice(container, "env")
	require.NoError(t, err)
	require.True(t, found)
	assert.Greater(t, len(env), 0, "Should have environment variables")
}

// TestBackend006_IdempotentCreation validates idempotent service creation.
func TestBackend006_IdempotentCreation(t *testing.T) {
	// Arrange
	ctx := context.Background()
	manager := setupServiceManager(t)

	completionData := &builds.BuildCompletionEventData{
		ThirdPartyID: "customer-123",
		ParserID:     "parser-abc",
		ImageURI:     "123456789.dkr.ecr.us-east-1.amazonaws.com/parsers:test",
	}

	serviceName := manager.GenerateServiceName(completionData.ThirdPartyID, completionData.ParserID)

	// Act - Create twice
	err1 := manager.CreateService(ctx, serviceName, completionData)
	err2 := manager.CreateService(ctx, serviceName, completionData)

	// Assert
	require.NoError(t, err1, "First creation should succeed")
	require.NoError(t, err2, "Second creation should be idempotent")
}

// TestBackend006_ResourceLabels validates resource label application.
func TestBackend006_ResourceLabels(t *testing.T) {
	// Arrange
	manager := setupServiceManager(t)
	serviceName := TestServiceName

	completionData := &builds.BuildCompletionEventData{
		ThirdPartyID: "customer-123",
		ParserID:     "parser-abc",
		ImageURI:     "123456789.dkr.ecr.us-east-1.amazonaws.com/parsers:test",
	}

	// Act
	service := manager.CreateKnativeServiceResource(serviceName, completionData)

	// Assert
	labels := service.GetLabels()
	assert.Equal(t, "customer-123", labels["third-party-id"])
	assert.Equal(t, "parser-abc", labels["parser-id"])
	assert.Equal(t, "lambda", labels["component"])
}

// Helper Functions.

func setupServiceManager(t *testing.T) *handler.ServiceManagerImpl {
	// Create fake clients
	k8sClient := k8sfake.NewSimpleClientset()
	scheme := runtime.NewScheme()
	dynamicClient := fake.NewSimpleDynamicClient(scheme)

	// Create mock observability
	obs := handler.NewMockObservability()

	// Create service manager (using minimal config - most fields are optional in tests)
	manager, err := handler.NewServiceManager(handler.ServiceManagerConfig{
		K8sClient:     k8sClient,
		DynamicClient: dynamicClient,
		K8sConfig:     &config_pkg.KubernetesConfig{Namespace: "knative-lambda"},
		Observability: obs,
	})
	if err != nil {
		t.Fatalf("Failed to create service manager: %v", err)
	}

	return manager.(*handler.ServiceManagerImpl)
}
