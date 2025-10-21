// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🔒 SECURITY CONFIGURATION - Security and validation configuration
//
//	🎯 Purpose: Security settings, input validation, and security headers
//	💡 Features: Security headers, input validation
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package config

// 🔒 SecurityConfig - "Security configuration"
type SecurityConfig struct {
	// Validation Configuration
	ValidateInput bool // Default: true

	// Security Configuration
	SecurityHeaders map[string]string `envconfig:"SECURITY_HEADERS"`

	// Feature Flags
	SecurityEnabled bool `envconfig:"SECURITY_ENABLED" default:"true"`
	DebugMode       bool `envconfig:"DEBUG_MODE" default:"false"`
	DryRun          bool `envconfig:"DRY_RUN" default:"false"`
}

// 🔧 NewSecurityConfig - "Create security configuration with defaults"
func NewSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		ValidateInput:   true,
		SecurityHeaders: make(map[string]string),
		SecurityEnabled: true,
		DebugMode:       false,
		DryRun:          false,
	}
}

// 🔧 Validate - "Validate security configuration"
func (c *SecurityConfig) Validate() error {
	// Security headers validation can be added here if needed
	return nil
}

// 🔧 IsValidationEnabled - "Check if input validation is enabled"
func (c *SecurityConfig) IsValidationEnabled() bool {
	return c.ValidateInput
}

// 🔧 IsSecurityEnabled - "Check if security features are enabled"
func (c *SecurityConfig) IsSecurityEnabled() bool {
	return c.SecurityEnabled
}

// 🔧 IsDebugMode - "Check if debug mode is enabled"
func (c *SecurityConfig) IsDebugMode() bool {
	return c.DebugMode
}

// 🔧 IsDryRun - "Check if dry run mode is enabled"
func (c *SecurityConfig) IsDryRun() bool {
	return c.DryRun
}

// 🔧 GetSecurityHeaders - "Get security headers"
func (c *SecurityConfig) GetSecurityHeaders() map[string]string {
	return c.SecurityHeaders
}
