package validation

import (
	"strings"
	"testing"
)

func TestValidateHandler(t *testing.T) {
	tests := []struct {
		name        string
		handler     string
		wantErr     bool
		errorCode   string
		description string
	}{
		// Valid cases
		{
			name:        "valid simple handler",
			handler:     "main.handler",
			wantErr:     false,
			description: "Standard Python/Node handler format",
		},
		{
			name:        "valid with underscore",
			handler:     "index.process_event",
			wantErr:     false,
			description: "Handler with underscore in function name",
		},
		{
			name:        "valid with numbers",
			handler:     "app2.handler3",
			wantErr:     false,
			description: "Handler with numbers",
		},
		{
			name:        "empty handler uses default",
			handler:     "",
			wantErr:     false,
			description: "Empty handler is OK, will use default",
		},

		// BLUE-002: Template Injection Attacks
		{
			name:        "template injection via semicolon",
			handler:     `main.handler"; import os; print(os.environ); x="`,
			wantErr:     true,
			errorCode:   "HANDLER_INJECTION",
			description: "BLUE-002: Semicolon injection attempt",
		},
		{
			name:        "template injection via quotes",
			handler:     `main.handler"}}{{.}}{{`,
			wantErr:     true,
			errorCode:   "HANDLER_INJECTION", // Braces detected as injection chars
			description: "BLUE-002: Go template escape attempt",
		},
		{
			name:        "template injection via backticks",
			handler:     "main.handler`id`",
			wantErr:     true,
			errorCode:   "HANDLER_INJECTION", // Backticks detected as injection chars
			description: "BLUE-002: Shell command injection via backticks",
		},
		{
			name:        "command injection via shell",
			handler:     "main.handler; rm -rf /",
			wantErr:     true,
			errorCode:   "HANDLER_INJECTION",
			description: "Command injection attempt",
		},
		{
			name:        "command injection via pipe",
			handler:     "main.handler | cat /etc/passwd",
			wantErr:     true,
			errorCode:   "HANDLER_INJECTION",
			description: "Pipe-based command injection",
		},
		{
			name:        "command injection via dollar",
			handler:     "main.handler$(id)",
			wantErr:     true,
			errorCode:   "HANDLER_INJECTION",
			description: "Shell substitution injection",
		},
		{
			name:        "invalid format - no dot",
			handler:     "mainhandler",
			wantErr:     true,
			errorCode:   "HANDLER_INVALID_FORMAT",
			description: "Handler must have module.function format",
		},
		{
			name:        "invalid format - multiple dots",
			handler:     "my.module.handler",
			wantErr:     true,
			errorCode:   "HANDLER_INVALID_FORMAT",
			description: "Only one dot allowed",
		},
		{
			name:        "invalid format - starts with number",
			handler:     "1main.handler",
			wantErr:     true,
			errorCode:   "HANDLER_INVALID_FORMAT",
			description: "Cannot start with number",
		},
		{
			name:        "too long handler",
			handler:     strings.Repeat("a", 50) + "." + strings.Repeat("b", 51),
			wantErr:     true,
			errorCode:   "HANDLER_TOO_LONG",
			description: "Handler exceeds max length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHandler(tt.handler)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if vErr, ok := err.(*ValidationError); ok {
					if vErr.Code != tt.errorCode {
						t.Errorf("ValidateHandler() error code = %v, want %v", vErr.Code, tt.errorCode)
					}
				}
			}
		})
	}
}

func TestValidateGitURL(t *testing.T) {
	tests := []struct {
		name        string
		gitURL      string
		wantErr     bool
		errorCode   string
		description string
	}{
		// Valid cases
		{
			name:        "valid https github",
			gitURL:      "https://github.com/user/repo.git",
			wantErr:     false,
			description: "Standard GitHub HTTPS URL",
		},
		{
			name:        "valid https gitlab",
			gitURL:      "https://gitlab.com/user/repo.git",
			wantErr:     false,
			description: "Standard GitLab HTTPS URL",
		},
		{
			name:        "valid ssh github",
			gitURL:      "git@github.com:user/repo.git",
			wantErr:     true, // git@ URLs don't parse as standard URLs
			description: "SSH URLs need special handling",
		},
		{
			name:        "valid git protocol",
			gitURL:      "git://github.com/user/repo.git",
			wantErr:     false,
			description: "Git protocol URL",
		},
		{
			name:        "valid internal http",
			gitURL:      "http://gitea.gitea.svc.cluster.local/user/repo.git",
			wantErr:     false,
			description: "Internal Kubernetes service HTTP is allowed",
		},

		// BLUE-001: SSRF Attacks
		// Note: For HTTP URLs, scheme check happens first (rejects non-HTTPS for external hosts)
		// This is defense-in-depth: even if scheme check passed, host check would block
		{
			name:        "SSRF AWS metadata",
			gitURL:      "http://169.254.169.254/latest/meta-data/",
			wantErr:     true,
			errorCode:   "URL_INSECURE_SCHEME", // HTTP to external rejected first
			description: "BLUE-001: AWS metadata endpoint SSRF",
		},
		{
			name:        "SSRF GCP metadata",
			gitURL:      "http://metadata.google.internal/computeMetadata/v1/",
			wantErr:     true,
			errorCode:   "URL_INSECURE_SCHEME", // HTTP to external rejected first
			description: "BLUE-001: GCP metadata endpoint SSRF",
		},
		{
			name:        "SSRF Kubernetes API",
			gitURL:      "https://kubernetes.default.svc/api/v1/secrets",
			wantErr:     true,
			errorCode:   "URL_BLOCKED_HOST", // HTTPS passes, then host blocked
			description: "BLUE-001: Kubernetes API SSRF",
		},
		{
			name:        "SSRF localhost",
			gitURL:      "http://localhost:8080/",
			wantErr:     true,
			errorCode:   "URL_INSECURE_SCHEME", // HTTP to external rejected first
			description: "BLUE-001: Localhost SSRF",
		},
		{
			name:        "SSRF loopback IP",
			gitURL:      "http://127.0.0.1:8080/",
			wantErr:     true,
			errorCode:   "URL_INSECURE_SCHEME", // HTTP to external rejected first
			description: "BLUE-001: Loopback IP SSRF",
		},
		{
			name:        "SSRF private IP 10.x",
			gitURL:      "http://10.0.0.1:8080/",
			wantErr:     true,
			errorCode:   "URL_INSECURE_SCHEME", // HTTP to external rejected first
			description: "BLUE-001: Private 10.x network SSRF",
		},
		{
			name:        "SSRF private IP 172.x",
			gitURL:      "http://172.16.0.1:8080/",
			wantErr:     true,
			errorCode:   "URL_INSECURE_SCHEME", // HTTP to external rejected first
			description: "BLUE-001: Private 172.x network SSRF",
		},
		{
			name:        "SSRF private IP 192.168.x",
			gitURL:      "http://192.168.1.1:8080/",
			wantErr:     true,
			errorCode:   "URL_INSECURE_SCHEME", // HTTP to external rejected first
			description: "BLUE-001: Private 192.168.x network SSRF",
		},
		{
			name:        "SSRF private IP via HTTPS",
			gitURL:      "https://10.0.0.1:8080/",
			wantErr:     true,
			errorCode:   "URL_BLOCKED_IP", // HTTPS passes, IP range blocked
			description: "BLUE-001: Private IP blocked even with HTTPS",
		},
		{
			name:        "insecure HTTP external",
			gitURL:      "http://github.com/user/repo.git",
			wantErr:     true,
			errorCode:   "URL_INSECURE_SCHEME",
			description: "External HTTP not allowed (must use HTTPS)",
		},
		{
			name:        "invalid scheme file",
			gitURL:      "file:///etc/passwd",
			wantErr:     true,
			errorCode:   "URL_INVALID_SCHEME",
			description: "File scheme not allowed",
		},
		{
			name:        "command injection in URL",
			gitURL:      "https://github.com/user/repo.git; rm -rf /",
			wantErr:     true,
			errorCode:   "URL_INJECTION",
			description: "Command injection in URL",
		},
		{
			name:        "empty URL",
			gitURL:      "",
			wantErr:     true,
			errorCode:   "URL_REQUIRED",
			description: "URL is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGitURL(tt.gitURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGitURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errorCode != "" {
				if vErr, ok := err.(*ValidationError); ok {
					if vErr.Code != tt.errorCode {
						t.Errorf("ValidateGitURL() error code = %v, want %v", vErr.Code, tt.errorCode)
					}
				}
			}
		})
	}
}

func TestValidateGitRef(t *testing.T) {
	tests := []struct {
		name        string
		ref         string
		wantErr     bool
		errorCode   string
		description string
	}{
		// Valid cases
		{
			name:        "valid branch main",
			ref:         "main",
			wantErr:     false,
			description: "Standard main branch",
		},
		{
			name:        "valid branch with slash",
			ref:         "feature/my-feature",
			wantErr:     false,
			description: "Feature branch with slash",
		},
		{
			name:        "valid tag",
			ref:         "v1.0.0",
			wantErr:     false,
			description: "Semantic version tag",
		},
		{
			name:        "valid commit hash",
			ref:         "abc123def456",
			wantErr:     false,
			description: "Short commit hash",
		},
		{
			name:        "empty uses default",
			ref:         "",
			wantErr:     false,
			description: "Empty ref uses default (main)",
		},

		// Command Injection Attacks
		{
			name:        "command injection via semicolon",
			ref:         "main; rm -rf /",
			wantErr:     true,
			errorCode:   "REF_INJECTION",
			description: "Semicolon command injection",
		},
		{
			name:        "command injection via pipe",
			ref:         "main | cat /etc/passwd",
			wantErr:     true,
			errorCode:   "REF_INJECTION",
			description: "Pipe command injection",
		},
		{
			name:        "command injection via backtick",
			ref:         "main`id`",
			wantErr:     true,
			errorCode:   "REF_INJECTION",
			description: "Backtick command injection",
		},
		{
			name:        "path traversal",
			ref:         "../../../etc/passwd",
			wantErr:     true,
			errorCode:   "REF_INVALID_FORMAT", // Dots in pattern trigger format error first
			description: "Path traversal attempt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGitRef(tt.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGitRef() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errorCode != "" {
				if vErr, ok := err.(*ValidationError); ok {
					if vErr.Code != tt.errorCode {
						t.Errorf("ValidateGitRef() error code = %v, want %v", vErr.Code, tt.errorCode)
					}
				}
			}
		})
	}
}

func TestValidateGitPath(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		wantErr     bool
		errorCode   string
		description string
	}{
		// Valid cases
		{
			name:        "valid simple path",
			path:        "src",
			wantErr:     false,
			description: "Simple subdirectory",
		},
		{
			name:        "valid nested path",
			path:        "src/lambda/function",
			wantErr:     false,
			description: "Nested subdirectory",
		},
		{
			name:        "valid file path",
			path:        "src/main.py",
			wantErr:     false,
			description: "Path to specific file",
		},
		{
			name:        "empty uses root",
			path:        "",
			wantErr:     false,
			description: "Empty path means repo root",
		},

		// BLUE-005: Path Traversal Attacks
		{
			name:        "path traversal double dot",
			path:        "../../../etc/passwd",
			wantErr:     true,
			errorCode:   "PATH_TRAVERSAL",
			description: "BLUE-005: Classic path traversal",
		},
		{
			name:        "path traversal in middle",
			path:        "src/../../../etc/passwd",
			wantErr:     true,
			errorCode:   "PATH_TRAVERSAL",
			description: "BLUE-005: Path traversal mid-path",
		},
		{
			name:        "absolute path escape",
			path:        "/etc/passwd",
			wantErr:     true,
			errorCode:   "PATH_ESCAPE",
			description: "Absolute path attempt",
		},
		{
			name:        "command injection",
			path:        "src; rm -rf /",
			wantErr:     true,
			errorCode:   "PATH_INJECTION",
			description: "Command injection in path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGitPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGitPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errorCode != "" {
				if vErr, ok := err.(*ValidationError); ok {
					if vErr.Code != tt.errorCode {
						t.Errorf("ValidateGitPath() error code = %v, want %v", vErr.Code, tt.errorCode)
					}
				}
			}
		})
	}
}

func TestValidateMinIOEndpoint(t *testing.T) {
	tests := []struct {
		name      string
		endpoint  string
		wantErr   bool
		errorCode string
	}{
		// Valid cases
		{"valid internal endpoint", "minio.minio.svc.cluster.local:9000", false, ""},
		{"valid hostname only", "minio.minio.svc.cluster.local", false, ""},
		{"empty uses default", "", false, ""},

		// Invalid cases
		{"SSRF localhost", "localhost:9000", true, "ENDPOINT_BLOCKED_HOST"},
		{"SSRF 127.0.0.1", "127.0.0.1:9000", true, "ENDPOINT_BLOCKED_HOST"},
		{"command injection", "minio.svc; rm -rf /", true, "ENDPOINT_INJECTION"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMinIOEndpoint(tt.endpoint)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMinIOEndpoint() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateBucketName(t *testing.T) {
	tests := []struct {
		name      string
		bucket    string
		wantErr   bool
		errorCode string
	}{
		// Valid cases
		{"valid simple", "mybucket", false, ""},
		{"valid with dots", "my.bucket.name", false, ""},
		{"valid with hyphens", "my-bucket-name", false, ""},

		// Invalid cases
		{"empty bucket", "", true, "BUCKET_REQUIRED"},
		{"too short", "ab", true, "BUCKET_INVALID_FORMAT"},
		{"uppercase", "MyBucket", true, "BUCKET_INVALID_FORMAT"},
		{"command injection", "bucket; rm -rf /", true, "BUCKET_INJECTION"},
		{"special chars", "bucket$name", true, "BUCKET_INJECTION"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBucketName(tt.bucket)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBucketName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateObjectKey(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		wantErr   bool
		errorCode string
	}{
		// Valid cases
		{"valid simple", "lambda/main.py", false, ""},
		{"valid with special chars", "lambda-function/main.py", false, ""},

		// Invalid cases
		{"empty key", "", true, "KEY_REQUIRED"},
		{"path traversal", "lambda/../../../etc/passwd", true, "KEY_PATH_TRAVERSAL"},
		{"command injection", "lambda; rm -rf /", true, "KEY_INJECTION"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateObjectKey(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateObjectKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeHandler(t *testing.T) {
	tests := []struct {
		name     string
		handler  string
		expected string
	}{
		{"valid handler returned as-is", "main.handler", "main.handler"},
		{"empty returns default", "", "main.handler"},
		{"injection returns default", `main.handler"; os.system("id")`, "main.handler"},
		{"invalid format returns default", "justhandler", "main.handler"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeHandler(tt.handler)
			if result != tt.expected {
				t.Errorf("SanitizeHandler() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidateSecurePath(t *testing.T) {
	tests := []struct {
		name     string
		basePath string
		reqPath  string
		wantErr  bool
	}{
		{"valid subpath", "/tmp/repo", "src/main.py", false},
		{"valid nested", "/tmp/repo", "src/lambda/handler.py", false},
		{"empty path OK", "/tmp/repo", "", false},
		{"path traversal escape", "/tmp/repo", "../../../etc/passwd", true},
		{"absolute path escape", "/tmp/repo", "/etc/passwd", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSecurePath(tt.basePath, tt.reqPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSecurePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Benchmark tests for performance-critical validation functions
func BenchmarkValidateHandler(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ValidateHandler("main.handler")
	}
}

func BenchmarkValidateGitURL(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ValidateGitURL("https://github.com/user/repo.git")
	}
}
