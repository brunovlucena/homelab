// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🔧 BUILD CONFIG - Build configuration models and validation
//
//	🎯 Purpose: Data structures for build configuration and parameters
//	💡 Features: Docker config, Lambda config, resource limits, validation
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package builds

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔧 BUILD CONFIG MODEL - "Build configuration"                         │
// └─────────────────────────────────────────────────────────────────────────┘

// 🔧 BuildConfig - "Build configuration and parameters"
type BuildConfig struct {
	// Docker build configuration
	DockerfilePath string            `json:"dockerfile_path,omitempty"`
	DockerContext  string            `json:"docker_context,omitempty"`
	BuildArgs      map[string]string `json:"build_args,omitempty"`

	// Resource limits
	CPULimit       string `json:"cpu_limit,omitempty"`
	MemoryLimit    string `json:"memory_limit,omitempty"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty"`

	// Lambda configuration
	LambdaHandler string `json:"lambda_handler,omitempty"`
	LambdaTimeout int    `json:"lambda_timeout,omitempty"`
	LambdaMemory  int    `json:"lambda_memory,omitempty"`

	// Environment variables
	Environment map[string]string `json:"environment,omitempty"`

	// Build steps
	PreBuildSteps  []BuildStep `json:"pre_build_steps,omitempty"`
	BuildSteps     []BuildStep `json:"build_steps,omitempty"`
	PostBuildSteps []BuildStep `json:"post_build_steps,omitempty"`

	// Caching configuration
	CacheEnabled bool     `json:"cache_enabled,omitempty"`
	CacheKeys    []string `json:"cache_keys,omitempty"`

	// Security configuration
	SecurityScanEnabled bool     `json:"security_scan_enabled,omitempty"`
	AllowedRegistries   []string `json:"allowed_registries,omitempty"`
}
