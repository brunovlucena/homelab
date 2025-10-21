package handler

import (
	"testing"
)

func TestParseSourceURL(t *testing.T) {
	tests := []struct {
		name       string
		sourceURL  string
		wantBucket string
		wantKey    string
		wantErr    bool
	}{
		{
			name:       "valid S3 URL - development",
			sourceURL:  "s3://knative-lambda-dev-fusion-modules-tmp/global/parser/0197ad6c10b973b2b854a0e652155b7e",
			wantBucket: "knative-lambda-dev-fusion-modules-tmp",
			wantKey:    "global/parser/0197ad6c10b973b2b854a0e652155b7e",
			wantErr:    false,
		},
		{
			name:      "invalid S3 URL - missing protocol",
			sourceURL: "notifi-uw2-dev-fusion-modules/global/parser/0197ad6c10b973b2b854a0e652155b7e",
			wantErr:   true,
		},
		{
			name:      "invalid S3 URL - missing bucket",
			sourceURL: "s3:///global/parser/0197ad6c10b973b2b854a0e652155b7e",
			wantErr:   true,
		},
		{
			name:      "invalid S3 URL - empty",
			sourceURL: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a minimal BuildContextManagerImpl for testing
			manager := &BuildContextManagerImpl{}

			bucket, key, err := manager.parseSourceURL(tt.sourceURL)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseSourceURL() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("parseSourceURL() unexpected error: %v", err)
				return
			}

			if bucket != tt.wantBucket {
				t.Errorf("parseSourceURL() bucket = %v, want %v", bucket, tt.wantBucket)
			}

			if key != tt.wantKey {
				t.Errorf("parseSourceURL() key = %v, want %v", key, tt.wantKey)
			}
		})
	}
}

func TestGetParserFileName(t *testing.T) {
	tests := []struct {
		name     string
		runtime  string
		expected string
	}{
		{
			name:     "nodejs22 runtime",
			runtime:  "nodejs22",
			expected: "index.js",
		},
		{
			name:     "nodejs22.x runtime",
			runtime:  "nodejs22.x",
			expected: "index.js",
		},

		// {
		// 	name:     "python3.9 runtime",
		// 	runtime:  "python3.9",
		// 	expected: "lambda_function.py",
		// },
		// {
		// 	name:     "go1.x runtime",
		// 	runtime:  "go1.x",
		// 	expected: "main.go",
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := &BuildContextManagerImpl{}
			result := manager.getParserFileName(tt.runtime)
			if result != tt.expected {
				t.Errorf("getParserFileName(%s) = %s, want %s", tt.runtime, result, tt.expected)
			}
		})
	}
}

// ES module conversion tests removed - no longer needed since we execute parser files directly
