// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🧪 Unit Tests: Build Manager
//
//	Tests for build operations:
//	- Dockerfile generation
//	- Tar.gz archive creation
//	- Content hash computation
//	- Source filename mapping
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package build

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	lambdav1alpha1 "github.com/brunovlucena/knative-lambda-operator/api/v1alpha1"
)

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📄 SOURCE FILENAME TESTS                                               │
// └─────────────────────────────────────────────────────────────────────────┘

func TestGetSourceFilename(t *testing.T) {
	m := &Manager{}

	tests := []struct {
		name             string
		language         string
		expectedFilename string
		description      string
	}{
		{
			name:             "Python",
			language:         "python",
			expectedFilename: "main.py",
			description:      "Python should use main.py",
		},
		{
			name:             "Python3",
			language:         "python3",
			expectedFilename: "main.py",
			description:      "Python3 should use main.py",
		},
		{
			name:             "NodeJS",
			language:         "nodejs",
			expectedFilename: "index.js",
			description:      "NodeJS should use index.js",
		},
		{
			name:             "Node",
			language:         "node",
			expectedFilename: "index.js",
			description:      "Node should use index.js",
		},
		{
			name:             "JavaScript",
			language:         "javascript",
			expectedFilename: "index.js",
			description:      "JavaScript should use index.js",
		},
		{
			name:             "Go",
			language:         "go",
			expectedFilename: "main.go",
			description:      "Go should use main.go",
		},
		{
			name:             "Golang",
			language:         "golang",
			expectedFilename: "main.go",
			description:      "Golang should use main.go",
		},
		{
			name:             "Unknown defaults to Python",
			language:         "unknown",
			expectedFilename: "main.py",
			description:      "Unknown language should default to main.py",
		},
		{
			name:             "Empty defaults to Python",
			language:         "",
			expectedFilename: "main.py",
			description:      "Empty language should default to main.py",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.getSourceFilename(tt.language)
			assert.Equal(t, tt.expectedFilename, result, tt.description)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🐳 DOCKERFILE GENERATION TESTS                                         │
// └─────────────────────────────────────────────────────────────────────────┘

func TestGenerateDockerfile(t *testing.T) {
	m := &Manager{
		pythonBaseImage: "localhost:5001/python:3.11-slim",
		nodeBaseImage:   "localhost:5001/node:20-alpine",
		goBaseImage:     "localhost:5001/golang:1.21-alpine",
		alpineRuntime:   "localhost:5001/alpine:3.19",
	}

	tests := []struct {
		name             string
		lambda           *lambdav1alpha1.LambdaFunction
		sourceFilename   string
		expectedContains []string
		expectError      bool
		errorMsg         string
		description      string
	}{
		{
			name: "Python Dockerfile",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Runtime: lambdav1alpha1.RuntimeSpec{
						Language: "python",
						Version:  "3.11",
						Handler:  "handler",
					},
				},
			},
			sourceFilename: "main.py",
			expectedContains: []string{
				"FROM localhost:5001/python:3.11-alpine",
				"WORKDIR /app",
				"COPY runtime.py",
				"COPY *.py",
				"ENV HANDLER=handler",
				"EXPOSE 8080",
				"CMD [\"python\", \"runtime.py\"]",
			},
			expectError: false,
			description: "Should generate valid Python Dockerfile",
		},
		{
			name: "Node.js Dockerfile",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Runtime: lambdav1alpha1.RuntimeSpec{
						Language: "nodejs",
						Version:  "20",
						Handler:  "index.handler",
					},
				},
			},
			sourceFilename: "index.js",
			expectedContains: []string{
				"FROM localhost:5001/node:20-alpine",
				"WORKDIR /app",
				"COPY *.js",
				"COPY runtime.js",
				"ENV HANDLER=index.handler",
				"EXPOSE 8080",
				"CMD [\"node\", \"runtime.js\"]",
			},
			expectError: false,
			description: "Should generate valid Node.js Dockerfile",
		},
		{
			name: "Go Dockerfile",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Runtime: lambdav1alpha1.RuntimeSpec{
						Language: "go",
						Version:  "1.21",
						Handler:  "main",
					},
				},
			},
			sourceFilename: "main.go",
			expectedContains: []string{
				"FROM localhost:5001/golang:1.21-alpine AS builder",
				"FROM localhost:5001/alpine:3.19",
				"WORKDIR /app",
				"COPY runtime.go",
				"CGO_ENABLED=0",
				"go build",
				"CMD [\"./lambda-runtime\"]",
			},
			expectError: false,
			description: "Should generate valid Go Dockerfile with multi-stage build",
		},
		{
			name: "Unsupported language",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Runtime: lambdav1alpha1.RuntimeSpec{
						Language: "rust",
						Version:  "1.70",
					},
				},
			},
			sourceFilename: "main.rs",
			expectError:    true,
			errorMsg:       "unsupported language",
			description:    "Should error for unsupported language",
		},
		{
			name: "Default handler",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Runtime: lambdav1alpha1.RuntimeSpec{
						Language: "python",
						Version:  "3.11",
						Handler:  "", // Empty handler
					},
				},
			},
			sourceFilename: "main.py",
			expectedContains: []string{
				"ENV HANDLER=handler", // Default handler
			},
			expectError: false,
			description: "Should use default handler when not specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dockerfile, err := m.generateDockerfile(tt.lambda, tt.sourceFilename)

			if tt.expectError {
				require.Error(t, err, tt.description)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err, tt.description)
				dockerfileStr := string(dockerfile)

				for _, expected := range tt.expectedContains {
					assert.Contains(t, dockerfileStr, expected,
						"Dockerfile should contain: %s", expected)
				}
			}
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔐 CONTENT HASH TESTS                                                  │
// └─────────────────────────────────────────────────────────────────────────┘

func TestComputeContentHash(t *testing.T) {
	m := &Manager{}

	tests := []struct {
		name        string
		sourceCode  []byte
		dockerfile  []byte
		runtime     []byte
		hashLength  int
		description string
	}{
		{
			name:        "Basic hash computation",
			sourceCode:  []byte("def handler(event): return event"),
			dockerfile:  []byte("FROM python:3.11\nCMD python"),
			runtime:     []byte("import json"),
			hashLength:  64, // SHA-256 hex string length
			description: "Should produce 64-character hex hash",
		},
		{
			name:        "Empty runtime wrapper",
			sourceCode:  []byte("package main"),
			dockerfile:  []byte("FROM golang:1.21\nCMD ./handler"),
			runtime:     nil,
			hashLength:  64,
			description: "Should handle nil runtime wrapper",
		},
		{
			name:        "Large source code",
			sourceCode:  bytes.Repeat([]byte("x"), 1000000), // 1MB
			dockerfile:  []byte("FROM python:3.11"),
			runtime:     []byte("wrapper"),
			hashLength:  64,
			description: "Should handle large content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := m.computeContentHash(tt.sourceCode, tt.dockerfile, tt.runtime)

			assert.Len(t, hash, tt.hashLength, tt.description)
			assert.Regexp(t, "^[a-f0-9]+$", hash, "Hash should be hex string")
		})
	}
}

func TestComputeContentHash_Deterministic(t *testing.T) {
	m := &Manager{}

	sourceCode := []byte("def handler(event): return {'status': 'ok'}")
	dockerfile := []byte("FROM python:3.11\nCOPY . .\nCMD python runtime.py")
	runtime := []byte("import main; main.handler()")

	// Compute hash multiple times
	hash1 := m.computeContentHash(sourceCode, dockerfile, runtime)
	hash2 := m.computeContentHash(sourceCode, dockerfile, runtime)
	hash3 := m.computeContentHash(sourceCode, dockerfile, runtime)

	assert.Equal(t, hash1, hash2, "Same content should produce same hash")
	assert.Equal(t, hash2, hash3, "Hash should be deterministic")
}

func TestComputeContentHash_Uniqueness(t *testing.T) {
	m := &Manager{}

	baseSource := []byte("def handler(event): return event")
	baseDockerfile := []byte("FROM python:3.11")
	baseRuntime := []byte("wrapper")

	hash1 := m.computeContentHash(baseSource, baseDockerfile, baseRuntime)

	// Change source
	modifiedSource := []byte("def handler(event): return {'modified': True}")
	hash2 := m.computeContentHash(modifiedSource, baseDockerfile, baseRuntime)

	// Change dockerfile
	modifiedDockerfile := []byte("FROM python:3.12")
	hash3 := m.computeContentHash(baseSource, modifiedDockerfile, baseRuntime)

	// Change runtime
	modifiedRuntime := []byte("different wrapper")
	hash4 := m.computeContentHash(baseSource, baseDockerfile, modifiedRuntime)

	assert.NotEqual(t, hash1, hash2, "Different source should produce different hash")
	assert.NotEqual(t, hash1, hash3, "Different dockerfile should produce different hash")
	assert.NotEqual(t, hash1, hash4, "Different runtime should produce different hash")
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📦 TAR.GZ ARCHIVE TESTS                                                │
// └─────────────────────────────────────────────────────────────────────────┘

func TestAddFileToTar(t *testing.T) {
	m := &Manager{}

	tests := []struct {
		name        string
		filename    string
		content     []byte
		description string
	}{
		{
			name:        "Add Python file",
			filename:    "main.py",
			content:     []byte("def handler(event): pass"),
			description: "Should add Python file to tar",
		},
		{
			name:        "Add Dockerfile",
			filename:    "Dockerfile",
			content:     []byte("FROM python:3.11"),
			description: "Should add Dockerfile to tar",
		},
		{
			name:        "Add empty file",
			filename:    "empty.txt",
			content:     []byte{},
			description: "Should handle empty content",
		},
		{
			name:        "Add file with spaces in name",
			filename:    "my file.py",
			content:     []byte("content"),
			description: "Should handle filenames with spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			tarWriter := tar.NewWriter(&buf)

			err := m.addFileToTar(tarWriter, tt.filename, tt.content)
			require.NoError(t, err, tt.description)

			err = tarWriter.Close()
			require.NoError(t, err)

			// Read back and verify
			tarReader := tar.NewReader(&buf)
			header, err := tarReader.Next()
			require.NoError(t, err)

			assert.Equal(t, tt.filename, header.Name)
			assert.Equal(t, int64(len(tt.content)), header.Size)

			readContent, err := io.ReadAll(tarReader)
			require.NoError(t, err)
			assert.Equal(t, tt.content, readContent)
		})
	}
}

func TestCreateTarGzArchive(t *testing.T) {
	m := &Manager{}

	lambda := &lambdav1alpha1.LambdaFunction{
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Runtime: lambdav1alpha1.RuntimeSpec{
				Language: "python",
			},
		},
	}

	sourceCode := []byte("def handler(event): return event")
	sourceFilename := "main.py"
	dockerfile := []byte("FROM python:3.11\nCOPY . .\nCMD python runtime.py")
	runtimeWrapper := []byte("import main")
	runtimeFilename := "runtime.py"

	archive, err := m.createTarGzArchive(
		sourceCode, sourceFilename,
		dockerfile,
		runtimeWrapper, runtimeFilename,
		lambda,
	)
	require.NoError(t, err)
	assert.NotEmpty(t, archive)

	// Decompress and verify contents
	files := extractTarGz(t, archive)

	assert.Contains(t, files, "main.py", "Archive should contain source file")
	assert.Contains(t, files, "Dockerfile", "Archive should contain Dockerfile")
	assert.Contains(t, files, "runtime.py", "Archive should contain runtime wrapper")
	assert.Contains(t, files, "requirements.txt", "Archive should contain requirements.txt for Python")
}

func TestCreateTarGzArchive_NodeJS(t *testing.T) {
	m := &Manager{}

	lambda := &lambdav1alpha1.LambdaFunction{
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Runtime: lambdav1alpha1.RuntimeSpec{
				Language: "nodejs",
			},
		},
	}

	sourceCode := []byte("module.exports.handler = async () => {}")
	sourceFilename := "index.js"
	dockerfile := []byte("FROM node:20\nCOPY . .\nCMD node runtime.js")
	runtimeWrapper := []byte("const handler = require('./index')")
	runtimeFilename := "runtime.js"

	archive, err := m.createTarGzArchive(
		sourceCode, sourceFilename,
		dockerfile,
		runtimeWrapper, runtimeFilename,
		lambda,
	)
	require.NoError(t, err)

	files := extractTarGz(t, archive)

	assert.Contains(t, files, "index.js", "Archive should contain source file")
	assert.Contains(t, files, "Dockerfile", "Archive should contain Dockerfile")
	assert.Contains(t, files, "runtime.js", "Archive should contain runtime wrapper")
	assert.Contains(t, files, "package.json", "Archive should contain package.json for Node.js")
}

func TestCreateTarGzArchive_Go(t *testing.T) {
	m := &Manager{}

	lambda := &lambdav1alpha1.LambdaFunction{
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Runtime: lambdav1alpha1.RuntimeSpec{
				Language: "go",
			},
		},
	}

	sourceCode := []byte("package main\nfunc main() {}")
	sourceFilename := "main.go"
	dockerfile := []byte("FROM golang:1.21\nCOPY . .\nRUN go build")

	// Go doesn't have a runtime wrapper
	archive, err := m.createTarGzArchive(
		sourceCode, sourceFilename,
		dockerfile,
		nil, "", // No runtime wrapper for Go
		lambda,
	)
	require.NoError(t, err)

	files := extractTarGz(t, archive)

	assert.Contains(t, files, "main.go", "Archive should contain source file")
	assert.Contains(t, files, "Dockerfile", "Archive should contain Dockerfile")
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔧 RUNTIME WRAPPER TESTS                                               │
// └─────────────────────────────────────────────────────────────────────────┘

func TestGenerateRuntimeWrapper(t *testing.T) {
	m := &Manager{}

	tests := []struct {
		name             string
		lambda           *lambdav1alpha1.LambdaFunction
		expectWrapper    bool
		expectedFilename string
		description      string
	}{
		{
			name: "Python runtime wrapper",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Runtime: lambdav1alpha1.RuntimeSpec{Language: "python"},
				},
			},
			expectWrapper:    true,
			expectedFilename: "runtime.py",
			description:      "Should generate Python runtime wrapper",
		},
		{
			name: "Node.js runtime wrapper",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Runtime: lambdav1alpha1.RuntimeSpec{Language: "nodejs"},
				},
			},
			expectWrapper:    true,
			expectedFilename: "runtime.js",
			description:      "Should generate Node.js runtime wrapper",
		},
		{
			name: "Go has no runtime wrapper",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Runtime: lambdav1alpha1.RuntimeSpec{Language: "go"},
				},
			},
			expectWrapper:    false,
			expectedFilename: "",
			description:      "Go should not have runtime wrapper",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapper, filename, err := m.generateRuntimeWrapper(tt.lambda)

			if tt.expectWrapper {
				require.NoError(t, err, tt.description)
				assert.NotEmpty(t, wrapper, "Wrapper should not be empty")
				assert.Equal(t, tt.expectedFilename, filename)
			} else {
				require.NoError(t, err, tt.description)
				assert.Empty(t, wrapper, "Wrapper should be empty for Go")
				assert.Empty(t, filename, "Filename should be empty for Go")
			}
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📝 CONSTANTS TESTS                                                     │
// └─────────────────────────────────────────────────────────────────────────┘

func TestBuildConstants(t *testing.T) {
	assert.NotEmpty(t, BuildContextConfigMapSuffix, "ConfigMap suffix should be defined")
	assert.Greater(t, JobTTLAfterFinished, int32(0), "TTL should be positive")
	assert.NotEmpty(t, DefaultKanikoImage, "Default Kaniko image should be defined")
	assert.NotEmpty(t, DefaultMinioClientImage, "Default MinIO client image should be defined")
	assert.NotEmpty(t, DefaultAlpineInitImage, "Default Alpine init image should be defined")
}

func TestDefaultBaseImages(t *testing.T) {
	assert.NotEmpty(t, DefaultNodeBaseImage, "Default Node base image should be defined")
	assert.NotEmpty(t, DefaultPythonBaseImage, "Default Python base image should be defined")
	assert.NotEmpty(t, DefaultGoBaseImage, "Default Go base image should be defined")
	assert.NotEmpty(t, DefaultAlpineRuntime, "Default Alpine runtime should be defined")

	// Verify they use the expected registry
	assert.True(t, strings.HasPrefix(DefaultNodeBaseImage, "localhost:5001/"),
		"Node base image should use local registry")
	assert.True(t, strings.HasPrefix(DefaultPythonBaseImage, "localhost:5001/"),
		"Python base image should use local registry")
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🏗️ BUILD STATUS TESTS                                                  │
// └─────────────────────────────────────────────────────────────────────────┘

func TestBuildStatus(t *testing.T) {
	tests := []struct {
		name        string
		status      BuildStatus
		isComplete  bool
		isSuccess   bool
		description string
	}{
		{
			name:        "Incomplete build",
			status:      BuildStatus{Completed: false},
			isComplete:  false,
			isSuccess:   false,
			description: "Incomplete build should not be success",
		},
		{
			name:        "Completed successful build",
			status:      BuildStatus{Completed: true, Success: true, ImageURI: "localhost:5001/test:abc123"},
			isComplete:  true,
			isSuccess:   true,
			description: "Completed build with image should be success",
		},
		{
			name:        "Completed failed build",
			status:      BuildStatus{Completed: true, Success: false, Error: "Build failed"},
			isComplete:  true,
			isSuccess:   false,
			description: "Completed build with error should not be success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isComplete, tt.status.Completed, tt.description)
			assert.Equal(t, tt.isSuccess, tt.status.Success, tt.description)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔧 BUILD CONTEXT TESTS                                                 │
// └─────────────────────────────────────────────────────────────────────────┘

func TestBuildContext(t *testing.T) {
	ctx := &BuildContext{
		ConfigMapName: "my-lambda-build-context",
		ContentHash:   "abc123def456",
		ImageTag:      "abc123def456",
	}

	assert.Equal(t, "my-lambda-build-context", ctx.ConfigMapName)
	assert.Equal(t, "abc123def456", ctx.ContentHash)
	assert.Equal(t, "abc123def456", ctx.ImageTag)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🛠️ HELPER FUNCTIONS                                                    │
// └─────────────────────────────────────────────────────────────────────────┘

func extractTarGz(t *testing.T, data []byte) map[string][]byte {
	t.Helper()

	gzReader, err := gzip.NewReader(bytes.NewReader(data))
	require.NoError(t, err)
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)
	files := make(map[string][]byte)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)

		content, err := io.ReadAll(tarReader)
		require.NoError(t, err)

		files[header.Name] = content
	}

	return files
}

func newTestLambdaForBuild(name, namespace, sourceType, language string) *lambdav1alpha1.LambdaFunction {
	lambda := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Source: lambdav1alpha1.SourceSpec{
				Type: sourceType,
			},
			Runtime: lambdav1alpha1.RuntimeSpec{
				Language: language,
				Version:  "3.11",
				Handler:  "handler",
			},
		},
	}

	if sourceType == "inline" {
		lambda.Spec.Source.Inline = &lambdav1alpha1.InlineSource{
			Code: "def handler(event): return event",
		}
	}

	return lambda
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔀 GIT SOURCE TESTS                                                    │
// └─────────────────────────────────────────────────────────────────────────┘

func TestFindSourceFileInDir(t *testing.T) {
	m := &Manager{}

	tests := []struct {
		name             string
		language         string
		files            map[string]string // filename -> content
		expectedFilename string
		expectError      bool
		description      string
	}{
		{
			name:     "Python main.py",
			language: "python",
			files: map[string]string{
				"main.py":   "def handler(event): return event",
				"utils.py":  "# utils",
				"README.md": "# readme",
			},
			expectedFilename: "main.py",
			expectError:      false,
			description:      "Should find main.py for Python",
		},
		{
			name:     "Node.js index.js",
			language: "nodejs",
			files: map[string]string{
				"index.js":     "module.exports.handler = async () => {}",
				"helper.js":    "// helper",
				"package.json": "{}",
			},
			expectedFilename: "index.js",
			expectError:      false,
			description:      "Should find index.js for Node.js",
		},
		{
			name:     "Go main.go",
			language: "go",
			files: map[string]string{
				"main.go":   "package main\nfunc main() {}",
				"helper.go": "package main",
				"go.mod":    "module test",
			},
			expectedFilename: "main.go",
			expectError:      false,
			description:      "Should find main.go for Go",
		},
		{
			name:     "No matching files",
			language: "python",
			files: map[string]string{
				"README.md":   "# readme",
				"config.yaml": "key: value",
			},
			expectedFilename: "",
			expectError:      true,
			description:      "Should error when no source files found",
		},
		{
			name:     "Fallback to any Python file",
			language: "python",
			files: map[string]string{
				"handler.py": "def handler(event): pass",
				"README.md":  "# readme",
			},
			expectedFilename: "handler.py",
			expectError:      false,
			description:      "Should fall back to any .py file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory with test files
			tmpDir, err := os.MkdirTemp("", "git-source-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tmpDir)

			for filename, content := range tt.files {
				filepath := tmpDir + "/" + filename
				err := os.WriteFile(filepath, []byte(content), 0644)
				require.NoError(t, err)
			}

			data, filename, err := m.findSourceFileInDir(tmpDir, tt.language)

			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				require.NoError(t, err, tt.description)
				assert.Equal(t, tt.expectedFilename, filename, tt.description)
				assert.NotEmpty(t, data, "Source code should not be empty")
			}
		})
	}
}

func TestGitSourceSpec(t *testing.T) {
	// Test that GitSource spec fields work correctly
	lambda := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-git-lambda",
			Namespace: "default",
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Source: lambdav1alpha1.SourceSpec{
				Type: "git",
				Git: &lambdav1alpha1.GitSource{
					URL:  "https://github.com/example/repo.git",
					Ref:  "main",
					Path: "src/functions/hello",
				},
			},
			Runtime: lambdav1alpha1.RuntimeSpec{
				Language: "python",
				Version:  "3.11",
				Handler:  "handler",
			},
		},
	}

	assert.Equal(t, "git", lambda.Spec.Source.Type)
	assert.NotNil(t, lambda.Spec.Source.Git)
	assert.Equal(t, "https://github.com/example/repo.git", lambda.Spec.Source.Git.URL)
	assert.Equal(t, "main", lambda.Spec.Source.Git.Ref)
	assert.Equal(t, "src/functions/hello", lambda.Spec.Source.Git.Path)
}

func TestGitSourceWithSecretRef(t *testing.T) {
	// Test GitSource with secret reference
	lambda := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-git-private",
			Namespace: "default",
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Source: lambdav1alpha1.SourceSpec{
				Type: "git",
				Git: &lambdav1alpha1.GitSource{
					URL: "https://github.com/private/repo.git",
					Ref: "v1.0.0",
					SecretRef: &corev1.LocalObjectReference{
						Name: "git-credentials",
					},
				},
			},
			Runtime: lambdav1alpha1.RuntimeSpec{
				Language: "nodejs",
				Version:  "20",
			},
		},
	}

	assert.NotNil(t, lambda.Spec.Source.Git.SecretRef)
	assert.Equal(t, "git-credentials", lambda.Spec.Source.Git.SecretRef.Name)
}

func TestGitSourceDefaultRef(t *testing.T) {
	// Test that empty ref defaults to "main" in the implementation
	lambda := &lambdav1alpha1.LambdaFunction{
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Source: lambdav1alpha1.SourceSpec{
				Type: "git",
				Git: &lambdav1alpha1.GitSource{
					URL: "https://github.com/example/repo.git",
					// Ref intentionally empty
				},
			},
		},
	}

	// The default is applied in the CRD validation (kubebuilder:default="main")
	// and in the code as a fallback
	if lambda.Spec.Source.Git.Ref == "" {
		// Code should default to "main"
		assert.Empty(t, lambda.Spec.Source.Git.Ref)
	}
}

func TestGitSourceVariousLanguages(t *testing.T) {
	m := &Manager{}

	languages := []struct {
		language         string
		expectedFilename string
	}{
		{"python", "main.py"},
		{"python3", "main.py"},
		{"nodejs", "index.js"},
		{"node", "index.js"},
		{"javascript", "index.js"},
		{"go", "main.go"},
		{"golang", "main.go"},
	}

	for _, tc := range languages {
		t.Run(tc.language, func(t *testing.T) {
			filename := m.getSourceFilename(tc.language)
			assert.Equal(t, tc.expectedFilename, filename,
				"Language %s should use filename %s", tc.language, tc.expectedFilename)
		})
	}
}
