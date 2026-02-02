package build

import (
	"os"
	"time"
)

// Config holds build configuration loaded from environment
type Config struct {
	// Registry settings
	// DefaultRegistry is used by Kaniko to push images (uses k8s service DNS)
	DefaultRegistry string
	// PullRegistry is used by kubelet to pull images (uses containerd mirror, e.g. localhost:5001)
	PullRegistry string
	KanikoImage  string

	// Helper images
	MinIOClientImage string
	GCSClientImage   string
	GitClientImage   string

	// Base images for runtimes (used by Kaniko in Dockerfile FROM)
	// These use k8s service DNS since Kaniko runs inside the pod
	NodeBaseImage      string
	PythonBaseImage    string
	GoBaseImage        string
	AlpineRuntimeImage string // Used in Dockerfile FROM for Go runtime

	// Init container image (pulled by kubelet via containerd mirror)
	AlpineInitImage string // Used for inline source init containers

	// MinIO secret key names (for credential mapping)
	MinIOAccessKeyField string
	MinIOSecretKeyField string

	// AWS/S3 secret key names
	AWSAccessKeyField string
	AWSSecretKeyField string

	// Default timeout
	DefaultTimeout time.Duration
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		DefaultRegistry:  getEnv("BUILD_DEFAULT_REGISTRY", "localhost:5001"),
		PullRegistry:     getEnv("BUILD_PULL_REGISTRY", "localhost:5001"),
		KanikoImage:      getEnv("BUILD_KANIKO_IMAGE", "localhost:5001/kaniko-executor:v1.19.2"),
		MinIOClientImage: getEnv("BUILD_MINIO_CLIENT_IMAGE", "localhost:5001/mc:RELEASE.2025-08-13T08-35-41Z-cpuv1"),
		GCSClientImage:   getEnv("BUILD_GCS_CLIENT_IMAGE", "gcr.io/google.com/cloudsdktool/cloud-sdk:slim"),
		GitClientImage:   getEnv("BUILD_GIT_CLIENT_IMAGE", "alpine/git:2.43.0"),
		NodeBaseImage:    getEnv("BUILD_NODE_BASE_IMAGE", "localhost:5001/node"),
		PythonBaseImage:  getEnv("BUILD_PYTHON_BASE_IMAGE", "localhost:5001/python"),
		GoBaseImage:      getEnv("BUILD_GO_BASE_IMAGE", "localhost:5001/golang"),
		// AlpineRuntimeImage: Used by Kaniko in Dockerfile FROM (can use k8s service DNS)
		AlpineRuntimeImage: getEnv("BUILD_ALPINE_RUNTIME_IMAGE", getEnv("BUILD_ALPINE_BASE_IMAGE", "localhost:5001/alpine:3.19")),
		// AlpineInitImage: Used for init containers (must use containerd mirror)
		AlpineInitImage:     getEnv("BUILD_ALPINE_INIT_IMAGE", getEnv("BUILD_ALPINE_BASE_IMAGE", "localhost:5001/alpine:3.19")),
		MinIOAccessKeyField: getEnv("MINIO_ACCESS_KEY_FIELD", "access-key"),
		MinIOSecretKeyField: getEnv("MINIO_SECRET_KEY_FIELD", "secret-key"),
		AWSAccessKeyField:   getEnv("AWS_ACCESS_KEY_FIELD", "AWS_ACCESS_KEY_ID"),
		AWSSecretKeyField:   getEnv("AWS_SECRET_KEY_FIELD", "AWS_SECRET_ACCESS_KEY"),
		DefaultTimeout:      parseDuration(getEnv("BUILD_DEFAULT_TIMEOUT", "30m")),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 30 * time.Minute
	}
	return d
}
