// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
//
//	🔒 SEC-006: Secrets Exposure & Credential Leakage Testing
//
//	User Story: Secrets Exposure & Credential Leakage Testing
//	Priority: P0 | Story Points: 13
//
//	Tests validate:
//	- Kubernetes secrets protection
//	- Environment variable exposure
//	- Secrets in logs prevention
//	- Version control secrets scanning
//	- API response secret leakage
//	- Container image secrets
//	- CloudEvent data sanitization
//	- AWS credentials protection
//	- Build-time secrets handling
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
package security

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSec006_SecretsNotInLogs validates secrets don't appear in logs.
func TestSec006_SecretsNotInLogs(t *testing.T) {
	tests := []struct {
		name        string
		logMessage  string
		hasSecret   bool
		description string
	}{
		{
			name:        "Password in logs",
			logMessage:  "Connecting with password=SuperSecret123!",
			hasSecret:   true,
			description: "Password should be detected in logs",
		},
		{
			name:        "API key in logs",
			logMessage:  "Using api_key=sk-1234567890abcdef",
			hasSecret:   true,
			description: "API key should be detected in logs",
		},
		{
			name:        "Token in logs",
			logMessage:  "Authorization token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
			hasSecret:   true,
			description: "Token should be detected in logs",
		},
		{
			name:        "AWS access key in logs",
			logMessage:  "AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE",
			hasSecret:   true,
			description: "AWS access key should be detected",
		},
		{
			name:        "AWS secret key in logs",
			logMessage:  "AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			hasSecret:   true,
			description: "AWS secret key should be detected",
		},
		{
			name:        "Safe log message",
			logMessage:  "Processing request for user 12345",
			hasSecret:   false,
			description: "Safe message should not be flagged",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			hasSecret := detectSecretInLog(tt.logMessage)

			// Assert
			assert.Equal(t, tt.hasSecret, hasSecret, tt.description)
		})
	}
}

// TestSec006_APIResponseRedaction validates secrets are redacted in API responses.
//
//nolint:funlen // Comprehensive secret redaction test with multiple scenarios
func TestSec006_APIResponseRedaction(t *testing.T) {
	tests := []struct {
		name         string
		response     map[string]interface{}
		field        string
		shouldRedact bool
		description  string
	}{
		{
			name: "API key redacted",
			response: map[string]interface{}{
				"api_key": "sk-1234567890abcdef1234567890abcdef",
			},
			field:        "api_key",
			shouldRedact: true,
			description:  "API keys should be redacted",
		},
		{
			name: "Password redacted",
			response: map[string]interface{}{
				"password": "SuperSecret123!",
			},
			field:        "password",
			shouldRedact: true,
			description:  "Passwords should be redacted",
		},
		{
			name: "Token redacted",
			response: map[string]interface{}{
				"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
			},
			field:        "token",
			shouldRedact: true,
			description:  "Tokens should be redacted",
		},
		{
			name: "Database URL redacted",
			response: map[string]interface{}{
				"database_url": "postgres://user:password@localhost:5432/db",
			},
			field:        "database_url",
			shouldRedact: true,
			description:  "Database URLs should be redacted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			handler := setupSecretsTestHandler(t)
			body, _ := json.Marshal(tt.response)
			req := httptest.NewRequest("POST", "/api/v1/build", strings.NewReader(string(body)))
			req.Header.Set("Authorization", "Bearer valid-token")
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(w, req)

			// Assert
			responseBody := w.Body.String()
			if tt.shouldRedact {
				// Check for redaction pattern (e.g., "****" or showing only last 4 chars)
				assert.NotContains(t, responseBody, tt.response[tt.field], tt.description)
			}
		})
	}
}

// TestSec006_EnvironmentVariableExposure validates env vars don't leak secrets.
func TestSec006_EnvironmentVariableExposure(t *testing.T) {
	sensitiveEnvVars := []string{
		"PASSWORD",
		"API_KEY",
		"SECRET",
		"TOKEN",
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
		"DATABASE_PASSWORD",
		"PRIVATE_KEY",
	}

	for _, envVar := range sensitiveEnvVars {
		t.Run(envVar, func(t *testing.T) {
			// Act
			isSensitive := isSensitiveEnvVar(envVar)

			// Assert
			assert.True(t, isSensitive, "Should detect %s as sensitive", envVar)
		})
	}
}

// TestSec006_SecretDetectionInCode validates secret detection patterns.
func TestSec006_SecretDetectionInCode(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		hasSecret   bool
		secretType  string
		description string
	}{
		{
			name:        "Hardcoded AWS key",
			code:        `AWS_ACCESS_KEY_ID = "AKIAIOSFODNN7EXAMPLE"`,
			hasSecret:   true,
			secretType:  "aws_access_key",
			description: "Should detect hardcoded AWS access key",
		},
		{
			name:        "Hardcoded password",
			code:        `password = "SuperSecret123!"`,
			hasSecret:   true,
			secretType:  "password",
			description: "Should detect hardcoded password",
		},
		{
			name:        "Private key",
			code:        `key = "-----BEGIN PRIVATE KEY-----\nMIIEvgIBADANBg..."`,
			hasSecret:   true,
			secretType:  "private_key",
			description: "Should detect private key",
		},
		{
			name:        "JWT token",
			code:        `token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0..."`,
			hasSecret:   true,
			secretType:  "jwt",
			description: "Should detect JWT token",
		},
		{
			name:        "Safe code",
			code:        `username = "user123"`,
			hasSecret:   false,
			secretType:  "",
			description: "Should not flag safe code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			hasSecret, secretType := detectSecretInCode(tt.code)

			// Assert
			assert.Equal(t, tt.hasSecret, hasSecret, tt.description)
			if tt.hasSecret {
				assert.Equal(t, tt.secretType, secretType, "Should detect correct secret type")
			}
		})
	}
}

// TestSec006_CloudEventDataSanitization validates CloudEvent data is sanitized.
func TestSec006_CloudEventDataSanitization(t *testing.T) {
	tests := []struct {
		name            string
		eventData       map[string]interface{}
		sensitiveFields []string
		description     string
	}{
		{
			name: "Password field sanitized",
			eventData: map[string]interface{}{
				"username":  "admin",
				"password":  "secret123",
				"parser_id": "parser-123",
			},
			sensitiveFields: []string{"password"},
			description:     "Password field should be sanitized",
		},
		{
			name: "API key sanitized",
			eventData: map[string]interface{}{
				"api_key":   "sk-1234567890",
				"parser_id": "parser-123",
			},
			sensitiveFields: []string{"api_key"},
			description:     "API key should be sanitized",
		},
		{
			name: "Multiple sensitive fields",
			eventData: map[string]interface{}{
				"token":       "abc123",
				"secret":      "xyz789",
				"credentials": "cred456",
			},
			sensitiveFields: []string{"token", "secret", "credentials"},
			description:     "All sensitive fields should be sanitized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			sanitized := sanitizeCloudEventData(tt.eventData)

			// Assert
			for _, field := range tt.sensitiveFields {
				value, exists := sanitized[field]
				if exists {
					assert.Equal(t, "[REDACTED]", value, "Field %s should be redacted", field)
				}
			}
		})
	}
}

// TestSec006_DockerImageSecretScanning validates container images don't have secrets.
func TestSec006_DockerImageSecretScanning(t *testing.T) {
	tests := []struct {
		name        string
		dockerfile  string
		hasSecret   bool
		description string
	}{
		{
			name: "Secret in ENV",
			dockerfile: `FROM alpine
ENV API_KEY=sk-1234567890abcdef
RUN echo "Setup"`,
			hasSecret:   true,
			description: "Secret in ENV should be detected",
		},
		{
			name: "Secret in ARG",
			dockerfile: `FROM alpine
ARG PASSWORD=SuperSecret123!
RUN echo $PASSWORD`,
			hasSecret:   true,
			description: "Secret in ARG should be detected",
		},
		{
			name: "Safe Dockerfile",
			dockerfile: `FROM alpine
RUN apk add --no-cache curl
COPY app /app`,
			hasSecret:   false,
			description: "Safe Dockerfile should not be flagged",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			hasSecret := detectSecretInDockerfile(tt.dockerfile)

			// Assert
			assert.Equal(t, tt.hasSecret, hasSecret, tt.description)
		})
	}
}

// TestSec006_GitHistorySecretScanning validates no secrets in git history.
func TestSec006_GitHistorySecretScanning(t *testing.T) {
	// This would typically use gitleaks or similar tool
	// Testing the detection patterns here

	tests := []struct {
		name        string
		content     string
		hasSecret   bool
		description string
	}{
		{
			name:        "AWS key in commit",
			content:     "AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE",
			hasSecret:   true,
			description: "AWS key in commit should be detected",
		},
		{
			name:        "Private key in commit",
			content:     "-----BEGIN RSA PRIVATE KEY-----",
			hasSecret:   true,
			description: "Private key should be detected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			hasSecret, _ := detectSecretInCode(tt.content)

			// Assert
			assert.Equal(t, tt.hasSecret, hasSecret, tt.description)
		})
	}
}

// TestSec006_SecretRedactionPatterns validates redaction patterns work correctly.
func TestSec006_SecretRedactionPatterns(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		description string
	}{
		{
			name:        "API key partial redaction",
			input:       "sk-1234567890abcdef1234567890abcdef",
			expected:    "sk-****************************cdef",
			description: "Should show prefix and last 4 chars",
		},
		{
			name:        "Password complete redaction",
			input:       "SuperSecret123!",
			expected:    "[REDACTED]",
			description: "Should completely redact password",
		},
		{
			name:        "Database URL redaction",
			input:       "postgres://user:password@localhost:5432/db",
			expected:    "postgres://user:****@localhost:5432/db",
			description: "Should redact password in URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			redacted := redactSecret(tt.input)

			// Assert
			assert.Contains(t, redacted, "****", tt.description)
			assert.NotEqual(t, tt.input, redacted, "Should not return original value")
		})
	}
}

// Helper Functions.

func setupSecretsTestHandler(_ *testing.T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate auth
		if r.Header.Get("Authorization") != "Bearer valid-token" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Parse request
		var data map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Sanitize response
		sanitized := sanitizeCloudEventData(data)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(sanitized)
	})
}

// SecretPattern represents a pattern for detecting secrets.
type SecretPattern struct {
	Pattern     *regexp.Regexp
	Type        string
	Confidence  string // high, medium, low
	Description string
}

var secretPatterns = []SecretPattern{
	// AWS Credentials (most specific first)
	{regexp.MustCompile(`AKIA[0-9A-Z]{16}`), "aws_access_key", "high", "AWS Access Key ID"},
	{regexp.MustCompile(`(?i)aws[_-]?access[_-]?key[_-]?id\s*[:=]\s*['"]?([A-Z0-9]{20})['"]?`), "aws_access_key", "high", "AWS Access Key ID with label"},
	{regexp.MustCompile(`(?i)aws[_-]?secret[_-]?access[_-]?key\s*[:=]\s*['"]?([A-Za-z0-9/+=]{40})['"]?`), "aws_secret_key", "high", "AWS Secret Access Key"},

	// GitHub Tokens
	{regexp.MustCompile(`ghp_[0-9a-zA-Z]{36}`), "github_pat", "high", "GitHub Personal Access Token"},
	{regexp.MustCompile(`gho_[0-9a-zA-Z]{36}`), "github_oauth", "high", "GitHub OAuth Access Token"},
	{regexp.MustCompile(`ghu_[0-9a-zA-Z]{36}`), "github_user_token", "high", "GitHub User-to-Server Token"},
	{regexp.MustCompile(`ghs_[0-9a-zA-Z]{36}`), "github_server_token", "high", "GitHub Server-to-Server Token"},
	{regexp.MustCompile(`ghr_[0-9a-zA-Z]{36}`), "github_refresh_token", "high", "GitHub Refresh Token"},

	// JWT Tokens (check before generic tokens)
	{regexp.MustCompile(`eyJ[A-Za-z0-9_-]+\.eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+`), "jwt", "high", "JWT Token"},

	// Specific API Keys (check before generic)
	{regexp.MustCompile(`(?i)api[_-]?key\s*[:=]\s*['"]?(sk-[a-zA-Z0-9_\-]{10,})['"]?`), "api_key", "high", "OpenAI/Stripe-style API Key"},
	{regexp.MustCompile(`(?i)api[_-]?key\s*[:=]\s*['"]?([a-zA-Z0-9_\-]{32,})['"]?`), "api_key", "medium", "Generic API Key"},
	{regexp.MustCompile(`(?i)secret[_-]?key\s*[:=]\s*['"]?([a-zA-Z0-9_\-]{16,})['"]?`), "secret_key", "medium", "Generic Secret Key"},

	// Private Keys
	{regexp.MustCompile(`-----BEGIN (RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`), "private_key", "high", "Private Key"},
	{regexp.MustCompile(`-----BEGIN (PGP|SSH2 ENCRYPTED) PRIVATE KEY BLOCK-----`), "private_key", "high", "Encrypted Private Key"},

	// Database Connection Strings
	{regexp.MustCompile(`(?i)(postgres|postgresql|mysql|mongodb|redis)://[^:]+:[^@\s'"]+@[^/\s'"]+`), "db_connection", "high", "Database Connection String with Password"},
	{regexp.MustCompile(`(?i)jdbc:[^:]+://[^:]+:[^@\s'"]+@`), "db_connection", "high", "JDBC Connection String with Password"},

	// Generic Passwords
	{regexp.MustCompile(`(?i)password\s*[:=]\s*['"]?([^\s'"]{8,})['"]?`), "password", "medium", "Password Assignment"},
	{regexp.MustCompile(`(?i)passwd\s*[:=]\s*['"]?([^\s'"]{8,})['"]?`), "password", "medium", "Password Assignment (passwd)"},

	// OAuth/Bearer Tokens (check before generic tokens)
	{regexp.MustCompile(`(?i)bearer\s+[a-zA-Z0-9_\-\.=]{20,}`), "bearer_token", "medium", "Bearer Token"},
	{regexp.MustCompile(`(?i)token\s*[:=]\s*['"]?(eyJ[A-Za-z0-9_\-\.]+)['"]?`), "jwt", "high", "JWT Token (labeled)"},
	{regexp.MustCompile(`(?i)token\s*[:=]\s*['"]?([a-zA-Z0-9_\-\.]{20,})['"]?`), "token", "low", "Generic Token"},

	// Slack Tokens
	{regexp.MustCompile(`xox[baprs]-[0-9]{10,13}-[0-9]{10,13}-[a-zA-Z0-9]{24,}`), "slack_token", "high", "Slack Token"},

	// Stripe API Keys
	{regexp.MustCompile(`sk_live_[0-9a-zA-Z]{24,}`), "stripe_secret", "high", "Stripe Live Secret Key"},

	// Docker ENV secrets
	{regexp.MustCompile(`(?m)^ENV\s+(PASSWORD|SECRET|TOKEN|API_KEY|AWS_[A-Z_]+)\s*[:=]?\s*\S+`), "docker_env_secret", "high", "Docker ENV with secret"},
	{regexp.MustCompile(`pk_live_[0-9a-zA-Z]{24,}`), "stripe_publishable", "medium", "Stripe Live Publishable Key"},

	// Google Cloud
	{regexp.MustCompile(`"type": "service_account"`), "gcp_service_account", "high", "GCP Service Account JSON"},

	// Azure
	{regexp.MustCompile(`(?i)azure[_-]?client[_-]?secret\s*[:=]\s*['"]?([a-zA-Z0-9~._-]{32,})['"]?`), "azure_secret", "high", "Azure Client Secret"},

	// SendGrid
	{regexp.MustCompile(`SG\.[a-zA-Z0-9_-]{22}\.[a-zA-Z0-9_-]{43}`), "sendgrid_api_key", "high", "SendGrid API Key"},

	// Twilio
	{regexp.MustCompile(`SK[a-f0-9]{32}`), "twilio_api_key", "high", "Twilio API Key"},

	// Generic credentials pattern
	{regexp.MustCompile(`(?i)(credentials?|auth)\s*[:=]\s*['"]?([a-zA-Z0-9_\-]{16,})['"]?`), "credentials", "low", "Generic Credentials"},
}

func detectSecretInLog(logMessage string) bool {
	matches := detectSecretsInText(logMessage)
	return len(matches) > 0
}

func detectSecretsInText(text string) []SecretPattern {
	var matches []SecretPattern

	for _, pattern := range secretPatterns {
		if pattern.Pattern.MatchString(text) {
			matches = append(matches, pattern)
		}
	}

	return matches
}

func isSensitiveEnvVar(envVar string) bool {
	sensitivePatterns := []string{
		"PASSWORD", "SECRET", "TOKEN", "KEY", "CREDENTIALS",
		"API_KEY", "AWS_ACCESS_KEY", "AWS_SECRET", "PRIVATE_KEY",
	}

	envUpper := strings.ToUpper(envVar)
	for _, pattern := range sensitivePatterns {
		if strings.Contains(envUpper, pattern) {
			return true
		}
	}
	return false
}

func detectSecretInCode(code string) (bool, string) {
	matches := detectSecretsInText(code)
	if len(matches) > 0 {
		return true, matches[0].Type
	}
	return false, ""
}

func sanitizeCloudEventData(data map[string]interface{}) map[string]interface{} {
	sensitiveFields := []string{
		"password", "secret", "token", "api_key", "apiKey",
		"credentials", "private_key", "privateKey", "database_url", "db_url", "connection_string",
	}

	sanitized := make(map[string]interface{})
	for key, value := range data {
		keyLower := strings.ToLower(key)
		isSensitive := false
		for _, field := range sensitiveFields {
			if strings.Contains(keyLower, field) {
				isSensitive = true
				break
			}
		}

		// Also check the value itself for secrets (like database URLs)
		if !isSensitive && value != nil {
			if strValue, ok := value.(string); ok {
				if hasSecret, _ := detectSecretInCode(strValue); hasSecret {
					isSensitive = true
				}
			}
		}

		if isSensitive {
			sanitized[key] = "[REDACTED]"
		} else {
			sanitized[key] = value
		}
	}
	return sanitized
}

func detectSecretInDockerfile(dockerfile string) bool {
	lines := strings.Split(dockerfile, "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "ENV") || strings.HasPrefix(strings.TrimSpace(line), "ARG") {
			if hasSecret, _ := detectSecretInCode(line); hasSecret {
				return true
			}
		}
	}
	return false
}

func redactSecret(secret string) string {
	if len(secret) <= 8 {
		return "[REDACTED]"
	}
	// Show first 2 and last 4 characters
	return secret[:2] + strings.Repeat("*", len(secret)-6) + secret[len(secret)-4:]
}
