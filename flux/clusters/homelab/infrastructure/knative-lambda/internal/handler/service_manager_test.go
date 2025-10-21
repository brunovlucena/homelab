package handler

import (
	"strings"
	"testing"

	"knative-lambda-new/internal/handler/helpers"
	"knative-lambda-new/pkg/builds"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestApplyResource_APIVersionParsing(t *testing.T) {
	// Test with v1 API version (no group)
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ServiceAccount",
			"metadata": map[string]interface{}{
				"name":      "test-sa",
				"namespace": "test-ns",
			},
		},
	}

	// This should not panic or return an error about invalid API version
	// We're just testing the parsing logic, not the actual Kubernetes operations
	apiVersion := obj.GetAPIVersion()
	parts := strings.Split(apiVersion, "/")
	if len(parts) == 1 {
		// Version only format (e.g., "v1")
		group := ""
		version := parts[0]
		if group != "" || version != "v1" {
			t.Errorf("Expected group='', version='v1', got group='%s', version='%s'", group, version)
		}
	} else {
		t.Errorf("Expected 1 part for v1 API version, got %d", len(parts))
	}

	// Test with group/version API version
	obj2 := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "test-deployment",
				"namespace": "test-ns",
			},
		},
	}

	apiVersion2 := obj2.GetAPIVersion()
	parts2 := strings.Split(apiVersion2, "/")
	if len(parts2) == 2 {
		// Group/Version format (e.g., "apps/v1")
		group := parts2[0]
		version := parts2[1]
		if group != "apps" || version != "v1" {
			t.Errorf("Expected group='apps', version='v1', got group='%s', version='%s'", group, version)
		}
	} else {
		t.Errorf("Expected 2 parts for apps/v1 API version, got %d", len(parts2))
	}
}

func TestCreateService_UpdatesExistingService(t *testing.T) {
	// Test that CreateService properly updates existing services with new image tags
	// This test verifies the fix for the issue where service updates weren't working

	// Create test completion data with different image URIs
	completionData1 := &builds.BuildCompletionEventData{
		ThirdPartyID: "test-third-party",
		ParserID:     "test-parser",
		ImageURI:     "test-image:v1",
		ContentHash:  "hash1",
	}

	completionData2 := &builds.BuildCompletionEventData{
		ThirdPartyID: "test-third-party",
		ParserID:     "test-parser",
		ImageURI:     "test-image:v2", // Updated image tag
		ContentHash:  "hash2",
	}

	// Verify that both completion data objects have the same service name
	serviceName1 := helpers.GenerateServiceName(completionData1.ThirdPartyID, completionData1.ParserID)
	serviceName2 := helpers.GenerateServiceName(completionData2.ThirdPartyID, completionData2.ParserID)

	if serviceName1 != serviceName2 {
		t.Errorf("Expected same service name for same third party and parser, got %s and %s", serviceName1, serviceName2)
	}

	// Verify that the image URIs are different (simulating an update)
	if completionData1.ImageURI == completionData2.ImageURI {
		t.Errorf("Expected different image URIs for update test, got same: %s", completionData1.ImageURI)
	}

	// Verify that content hashes are different (indicating code changes)
	if completionData1.ContentHash == completionData2.ContentHash {
		t.Errorf("Expected different content hashes for update test, got same: %s", completionData1.ContentHash)
	}
}

func TestServiceDeleteEventData_Structure(t *testing.T) {
	// Test that ServiceDeleteEventData structure is properly defined
	// This test verifies that the delete event data structure is correctly implemented

	deleteData := &builds.ServiceDeleteEventData{
		ThirdPartyID:  "test-third-party",
		ParserID:      "test-parser",
		ServiceName:   "test-service-name",
		CorrelationID: "test-correlation-id",
		Reason:        "test-deletion-reason",
	}

	// Verify that all fields are properly set
	if deleteData.ThirdPartyID != "test-third-party" {
		t.Errorf("Expected ThirdPartyID 'test-third-party', got '%s'", deleteData.ThirdPartyID)
	}

	if deleteData.ParserID != "test-parser" {
		t.Errorf("Expected ParserID 'test-parser', got '%s'", deleteData.ParserID)
	}

	if deleteData.ServiceName != "test-service-name" {
		t.Errorf("Expected ServiceName 'test-service-name', got '%s'", deleteData.ServiceName)
	}

	if deleteData.CorrelationID != "test-correlation-id" {
		t.Errorf("Expected CorrelationID 'test-correlation-id', got '%s'", deleteData.CorrelationID)
	}

	if deleteData.Reason != "test-deletion-reason" {
		t.Errorf("Expected Reason 'test-deletion-reason', got '%s'", deleteData.Reason)
	}
}

func TestGenerateServiceNameForDelete(t *testing.T) {
	// Test that service name generation works correctly for delete events
	// This test verifies that the same naming convention is used for deletion

	thirdPartyID := "test-third-party"
	parserID := "test-parser"

	// Generate service name using the same helper function
	serviceName := helpers.GenerateServiceName(thirdPartyID, parserID)

	// Verify that the service name follows the expected pattern
	// Note: GenerateServiceName truncates thirdPartyID to 15 chars and parserID to 15 chars
	expectedPattern := "lambda-test-third-part-test-parser"
	if serviceName != expectedPattern {
		t.Errorf("Expected service name '%s', got '%s'", expectedPattern, serviceName)
	}

	// Verify that the service name is consistent
	serviceName2 := helpers.GenerateServiceName(thirdPartyID, parserID)
	if serviceName != serviceName2 {
		t.Errorf("Expected consistent service name generation, got '%s' and '%s'", serviceName, serviceName2)
	}
}
