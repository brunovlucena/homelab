package handler

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"knative-lambda-new/internal/errors"
	testhelpers "knative-lambda-new/internal/testing"
	"knative-lambda-new/pkg/builds"
)

func TestParseBuildRequest_ValidData(t *testing.T) {
	// Create a valid CloudEvent with proper BuildEventData
	eventData := builds.BuildEventData{
		ThirdPartyID: "test-third-party",
		ParserID:     "test-parser",
		ContextID:    "test-context",
		Parameters: map[string]interface{}{
			"buildType": "container",
			"runtime":   "nodejs22",
			"blockId":   "test-block",
		},
	}

	event := cloudevents.NewEvent()
	event.SetID("test-event-id")
	event.SetSource("test-source")
	event.SetType(builds.EventTypeBuildStart)
	event.SetTime(time.Now())
	event.SetData(cloudevents.ApplicationJSON, eventData)

	// Create a properly initialized handler with observability and config
	obs := testhelpers.CreateTestObservability(t)
	cfg := testhelpers.CreateTestConfig(t)
	handler := &EventHandlerImpl{
		obs:    obs,
		config: cfg,
	}

	// Test the parsing
	buildRequest, err := handler.ParseBuildRequest(context.Background(), &event)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, buildRequest)
	assert.Equal(t, "test-third-party", buildRequest.ThirdPartyID)
	assert.Equal(t, "test-parser", buildRequest.ParserID)
	assert.Equal(t, "container", buildRequest.BuildType)
	assert.Equal(t, "nodejs22", buildRequest.Runtime)
	assert.Equal(t, "test-block", buildRequest.BlockID)
}

func TestParseBuildRequest_MissingRequiredFields(t *testing.T) {
	// Test case 1: Missing ThirdPartyID
	eventData := builds.BuildEventData{
		ParserID: "test-parser",
		Parameters: map[string]interface{}{
			"buildType": "container",
		},
	}

	event := cloudevents.NewEvent()
	event.SetID("test-event-id")
	event.SetSource("test-source")
	event.SetType(builds.EventTypeBuildStart)
	event.SetTime(time.Now())
	event.SetData(cloudevents.ApplicationJSON, eventData)

	handler := &EventHandlerImpl{}

	buildRequest, err := handler.ParseBuildRequest(context.Background(), &event)

	assert.Error(t, err)
	assert.Nil(t, buildRequest)

	// Check if it's a validation error
	var validationErr *errors.ValidationError
	assert.ErrorAs(t, err, &validationErr)
	assert.Contains(t, validationErr.Error(), "validation failed for field 'build_request_data'")
	assert.Contains(t, validationErr.Error(), "thirdPartyID is required")
}

func TestParseBuildRequest_InvalidJSON(t *testing.T) {
	// Test case 2: Invalid JSON structure
	invalidData := map[string]interface{}{
		"invalid_field": "invalid_value",
		"another_field": 123,
	}

	event := cloudevents.NewEvent()
	event.SetID("test-event-id")
	event.SetSource("test-source")
	event.SetType(builds.EventTypeBuildStart)
	event.SetTime(time.Now())
	event.SetData(cloudevents.ApplicationJSON, invalidData)

	handler := &EventHandlerImpl{}

	buildRequest, err := handler.ParseBuildRequest(context.Background(), &event)

	assert.Error(t, err)
	assert.Nil(t, buildRequest)

	// Check if it's a validation error
	var validationErr *errors.ValidationError
	assert.ErrorAs(t, err, &validationErr)
	assert.Contains(t, validationErr.Error(), "validation failed for field 'build_request_data'")
}

func TestBuildEventData_Validation(t *testing.T) {
	tests := []struct {
		name           string
		eventData      builds.BuildEventData
		expectedErrors []string
	}{
		{
			name: "Valid data",
			eventData: builds.BuildEventData{
				ThirdPartyID: "valid-third-party",
				ParserID:     "valid-parser",
			},
			expectedErrors: []string{},
		},
		{
			name: "Missing ThirdPartyID",
			eventData: builds.BuildEventData{
				ParserID: "valid-parser",
			},
			expectedErrors: []string{"thirdPartyID is required"},
		},
		{
			name: "Missing ParserID",
			eventData: builds.BuildEventData{
				ThirdPartyID: "valid-third-party",
			},
			expectedErrors: []string{"parserID is required"},
		},
		{
			name:      "Both fields missing",
			eventData: builds.BuildEventData{},
			expectedErrors: []string{
				"thirdPartyID is required",
				"parserID is required",
			},
		},
		{
			name: "ThirdPartyID too long",
			eventData: builds.BuildEventData{
				ThirdPartyID: string(make([]byte, 101)), // 101 characters
				ParserID:     "valid-parser",
			},
			expectedErrors: []string{"thirdPartyID must be 100 characters or less"},
		},
		{
			name: "ParserID too long",
			eventData: builds.BuildEventData{
				ThirdPartyID: "valid-third-party",
				ParserID:     string(make([]byte, 101)), // 101 characters
			},
			expectedErrors: []string{"parserID must be 100 characters or less"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := tt.eventData.Validate()
			assert.ElementsMatch(t, tt.expectedErrors, errors)

			isValid := tt.eventData.IsValid()
			assert.Equal(t, len(tt.expectedErrors) == 0, isValid)
		})
	}
}

func TestBuildEventData_GetParameterAsString(t *testing.T) {
	eventData := builds.BuildEventData{
		ThirdPartyID: "test-third-party",
		ParserID:     "test-parser",
		Parameters: map[string]interface{}{
			"string_param": "string_value",
			"int_param":    123,
			"bool_param":   true,
		},
	}

	// Test valid string parameter
	value, ok := eventData.GetParameterAsString("string_param")
	assert.True(t, ok)
	assert.Equal(t, "string_value", value)

	// Test non-string parameter
	value, ok = eventData.GetParameterAsString("int_param")
	assert.False(t, ok)
	assert.Equal(t, "", value)

	// Test non-existent parameter
	value, ok = eventData.GetParameterAsString("non_existent")
	assert.False(t, ok)
	assert.Equal(t, "", value)

	// Test with nil parameters
	eventData.Parameters = nil
	value, ok = eventData.GetParameterAsString("string_param")
	assert.False(t, ok)
	assert.Equal(t, "", value)
}

func TestBuildEventData_GetParameterAsInt(t *testing.T) {
	eventData := builds.BuildEventData{
		ThirdPartyID: "test-third-party",
		ParserID:     "test-parser",
		Parameters: map[string]interface{}{
			"int_param":    123,
			"float_param":  456.0,
			"string_int":   "789",
			"string_param": "not_a_number",
			"bool_param":   true,
		},
	}

	// Test valid int parameter
	value, ok := eventData.GetParameterAsInt("int_param")
	assert.True(t, ok)
	assert.Equal(t, 123, value)

	// Test float parameter (should be converted to int)
	value, ok = eventData.GetParameterAsInt("float_param")
	assert.True(t, ok)
	assert.Equal(t, 456, value)

	// Test string that can be converted to int
	value, ok = eventData.GetParameterAsInt("string_int")
	assert.True(t, ok)
	assert.Equal(t, 789, value)

	// Test string that cannot be converted to int
	value, ok = eventData.GetParameterAsInt("string_param")
	assert.False(t, ok)
	assert.Equal(t, 0, value)

	// Test non-existent parameter
	value, ok = eventData.GetParameterAsInt("non_existent")
	assert.False(t, ok)
	assert.Equal(t, 0, value)

	// Test with nil parameters
	eventData.Parameters = nil
	value, ok = eventData.GetParameterAsInt("int_param")
	assert.False(t, ok)
	assert.Equal(t, 0, value)
}

// Helper function to create a CloudEvent for testing
func createTestCloudEvent(eventType string, data interface{}) cloudevents.Event {
	event := cloudevents.NewEvent()
	event.SetID("test-event-id")
	event.SetSource("test-source")
	event.SetType(eventType)
	event.SetTime(time.Now())
	event.SetData(cloudevents.ApplicationJSON, data)
	return event
}

// Example of how to test the actual error message format
func TestValidationError_MessageFormat(t *testing.T) {
	// This test demonstrates the exact error message format that would be returned
	validationErr := errors.NewValidationError("build_request_data", nil, "failed to parse build request data: invalid JSON")

	expectedMessage := "validation failed for field 'build_request_data' with value '<nil>': failed to parse build request data: invalid JSON"
	assert.Equal(t, expectedMessage, validationErr.Error())

	// Test the JSON structure
	jsonData, err := json.Marshal(validationErr)
	require.NoError(t, err)

	var parsed map[string]interface{}
	err = json.Unmarshal(jsonData, &parsed)
	require.NoError(t, err)

	assert.Equal(t, "build_request_data", parsed["field"])
	assert.Nil(t, parsed["value"])
	assert.Equal(t, "failed to parse build request data: invalid JSON", parsed["reason"])
	assert.NotNil(t, parsed["time"])
}
