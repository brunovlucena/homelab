// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🛡️ SECURITY - Security validation and protection mechanisms
//
//	🎯 Purpose: Validate inputs and protect against security vulnerabilities
//	💡 Features: Input validation, security scanning, access control
//
//	🏛️ ARCHITECTURE:
//	🔍 Input Validation - Validate all incoming data for security compliance
//	🚫 Security Scanning - Scan for vulnerabilities and malicious content
//	🔐 Access Control - Control access to resources and operations
//	📊 Security Monitoring - Monitor security events and violations
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package security

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"knative-lambda-new/internal/observability"

	"go.opentelemetry.io/otel/attribute"
)

// SecurityValidator provides security validation functionality with comprehensive observability
type SecurityValidator struct {
	obs *observability.Observability
}

// SecurityValidationResult represents the result of a security validation
type SecurityValidationResult struct {
	Valid    bool
	Error    error
	Warnings []string
	Details  map[string]interface{}
}

// SecurityEventType represents different types of security events
type SecurityEventType string

const (
	// Validation Events
	EventTypeInputValidation     SecurityEventType = "input_validation"
	EventTypeImageValidation     SecurityEventType = "image_validation"
	EventTypeNamespaceValidation SecurityEventType = "namespace_validation"
	EventTypeEventDataValidation SecurityEventType = "event_data_validation"
	EventTypeIDValidation        SecurityEventType = "id_validation"

	// Threat Detection Events
	EventTypeSQLInjectionDetected     SecurityEventType = "sql_injection_detected"
	EventTypeXSSDetected              SecurityEventType = "xss_detected"
	EventTypePathTraversalDetected    SecurityEventType = "path_traversal_detected"
	EventTypeMaliciousContentDetected SecurityEventType = "malicious_content_detected"

	// Access Control Events
	EventTypeAccessDenied   SecurityEventType = "access_denied"
	EventTypeAccessGranted  SecurityEventType = "access_granted"
	EventTypeAuthentication SecurityEventType = "authentication"
	EventTypeAuthorization  SecurityEventType = "authorization"

	// Security Violation Events
	EventTypeSecurityViolation SecurityEventType = "security_violation"
	EventTypeRateLimitExceeded SecurityEventType = "rate_limit_exceeded"
	EventTypeQuotaExceeded     SecurityEventType = "quota_exceeded"
)

// SecurityThreatLevel represents the severity of a security threat
type SecurityThreatLevel string

const (
	ThreatLevelLow      SecurityThreatLevel = "low"
	ThreatLevelMedium   SecurityThreatLevel = "medium"
	ThreatLevelHigh     SecurityThreatLevel = "high"
	ThreatLevelCritical SecurityThreatLevel = "critical"
)

// NewSecurityValidator creates a new security validator with observability
func NewSecurityValidator(obs *observability.Observability) *SecurityValidator {
	return &SecurityValidator{
		obs: obs,
	}
}

// ValidateInput validates input for security compliance with comprehensive observability
func (s *SecurityValidator) ValidateInput(ctx context.Context, input string) *SecurityValidationResult {
	ctx, span := s.obs.StartSpan(ctx, "security.validate_input")
	defer span.End()

	start := time.Now()
	result := &SecurityValidationResult{
		Valid:   true,
		Details: make(map[string]interface{}),
	}

	// Add input metadata to span
	span.SetAttributes(
		attribute.Int("security.input_length", len(input)),
		attribute.String("security.input_type", "string"),
	)

	// Record validation attempt
	s.obs.Info(ctx, "Starting input validation",
		"input_length", len(input),
		"validation_type", "security_input")

	if input == "" {
		result.Valid = false
		result.Error = fmt.Errorf("input cannot be empty")
		s.recordValidationFailure(ctx, EventTypeInputValidation, "empty_input", ThreatLevelLow, result.Error)
		return result
	}

	// Check for potential injection attacks
	// Check XSS first as it's more specific (looks for <script> tags, not just "script" word)
	if containsXSS(input) {
		result.Valid = false
		result.Error = fmt.Errorf("input contains potential XSS attack")
		s.recordSecurityThreat(ctx, EventTypeXSSDetected, "xss_attack", ThreatLevelHigh, input)
		return result
	}

	if containsSQLInjection(input) {
		result.Valid = false
		result.Error = fmt.Errorf("input contains potential SQL injection")
		s.recordSecurityThreat(ctx, EventTypeSQLInjectionDetected, "sql_injection", ThreatLevelHigh, input)
		return result
	}

	if containsPathTraversal(input) {
		result.Valid = false
		result.Error = fmt.Errorf("input contains potential path traversal")
		s.recordSecurityThreat(ctx, EventTypePathTraversalDetected, "path_traversal", ThreatLevelMedium, input)
		return result
	}

	// Record successful validation
	duration := time.Since(start)
	s.recordValidationSuccess(ctx, EventTypeInputValidation, "input_validation", duration, result.Details)

	return result
}

// ValidateImageName validates Docker image names with observability
func (s *SecurityValidator) ValidateImageName(ctx context.Context, imageName string) *SecurityValidationResult {
	ctx, span := s.obs.StartSpan(ctx, "security.validate_image_name")
	defer span.End()

	start := time.Now()
	result := &SecurityValidationResult{
		Valid:   true,
		Details: make(map[string]interface{}),
	}

	span.SetAttributes(
		attribute.String("security.image_name", imageName),
		attribute.String("security.validation_type", "image_name"),
	)

	s.obs.Info(ctx, "Starting image name validation",
		"image_name", imageName,
		"validation_type", "image_name")

	if imageName == "" {
		result.Valid = false
		result.Error = fmt.Errorf("image name cannot be empty")
		s.recordValidationFailure(ctx, EventTypeImageValidation, "empty_image_name", ThreatLevelLow, result.Error)
		return result
	}

	// Basic image name validation
	imageRegex := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*[a-zA-Z0-9](:[a-zA-Z0-9._-]*)?$`)
	if !imageRegex.MatchString(imageName) {
		result.Valid = false
		result.Error = fmt.Errorf("invalid image name format: %s", imageName)
		s.recordValidationFailure(ctx, EventTypeImageValidation, "invalid_image_format", ThreatLevelMedium, result.Error)
		return result
	}

	// Check for potentially malicious image names
	if containsMaliciousImagePatterns(imageName) {
		result.Valid = false
		result.Error = fmt.Errorf("image name contains potentially malicious patterns: %s", imageName)
		s.recordSecurityThreat(ctx, EventTypeMaliciousContentDetected, "malicious_image_name", ThreatLevelHigh, imageName)
		return result
	}

	duration := time.Since(start)
	s.recordValidationSuccess(ctx, EventTypeImageValidation, "image_name_validation", duration, result.Details)

	return result
}

// ValidateNamespace validates Kubernetes namespace names with observability
func (s *SecurityValidator) ValidateNamespace(ctx context.Context, namespace string) *SecurityValidationResult {
	ctx, span := s.obs.StartSpan(ctx, "security.validate_namespace")
	defer span.End()

	start := time.Now()
	result := &SecurityValidationResult{
		Valid:   true,
		Details: make(map[string]interface{}),
	}

	span.SetAttributes(
		attribute.String("security.namespace", namespace),
		attribute.String("security.validation_type", "namespace"),
	)

	s.obs.Info(ctx, "Starting namespace validation",
		"namespace", namespace,
		"validation_type", "namespace")

	if namespace == "" {
		result.Valid = false
		result.Error = fmt.Errorf("namespace cannot be empty")
		s.recordValidationFailure(ctx, EventTypeNamespaceValidation, "empty_namespace", ThreatLevelLow, result.Error)
		return result
	}

	// Kubernetes namespace validation
	namespaceRegex := regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)
	if !namespaceRegex.MatchString(namespace) {
		result.Valid = false
		result.Error = fmt.Errorf("invalid namespace format: %s", namespace)
		s.recordValidationFailure(ctx, EventTypeNamespaceValidation, "invalid_namespace_format", ThreatLevelMedium, result.Error)
		return result
	}

	// Check for reserved namespaces
	if isReservedNamespace(namespace) {
		result.Warnings = append(result.Warnings, fmt.Sprintf("namespace '%s' is reserved and may cause conflicts", namespace))
	}

	duration := time.Since(start)
	s.recordValidationSuccess(ctx, EventTypeNamespaceValidation, "namespace_validation", duration, result.Details)

	return result
}

// ValidateEventData validates CloudEvent data with observability
func (s *SecurityValidator) ValidateEventData(ctx context.Context, data interface{}) *SecurityValidationResult {
	ctx, span := s.obs.StartSpan(ctx, "security.validate_event_data")
	defer span.End()

	start := time.Now()
	result := &SecurityValidationResult{
		Valid:   true,
		Details: make(map[string]interface{}),
	}

	span.SetAttributes(
		attribute.String("security.data_type", fmt.Sprintf("%T", data)),
		attribute.String("security.validation_type", "event_data"),
	)

	s.obs.Info(ctx, "Starting event data validation",
		"data_type", fmt.Sprintf("%T", data),
		"validation_type", "event_data")

	if data == nil {
		result.Valid = false
		result.Error = fmt.Errorf("event data cannot be nil")
		s.recordValidationFailure(ctx, EventTypeEventDataValidation, "nil_event_data", ThreatLevelLow, result.Error)
		return result
	}

	// Convert data to string for security scanning if possible
	if dataStr, ok := data.(string); ok {
		if containsMaliciousContent(dataStr) {
			result.Valid = false
			result.Error = fmt.Errorf("event data contains potentially malicious content")
			s.recordSecurityThreat(ctx, EventTypeMaliciousContentDetected, "malicious_event_data", ThreatLevelHigh, dataStr)
			return result
		}
	}

	duration := time.Since(start)
	s.recordValidationSuccess(ctx, EventTypeEventDataValidation, "event_data_validation", duration, result.Details)

	return result
}

// ValidateID validates an ID field with observability
func (s *SecurityValidator) ValidateID(ctx context.Context, id string) *SecurityValidationResult {
	ctx, span := s.obs.StartSpan(ctx, "security.validate_id")
	defer span.End()

	start := time.Now()
	result := &SecurityValidationResult{
		Valid:   true,
		Details: make(map[string]interface{}),
	}

	span.SetAttributes(
		attribute.Int("security.id_length", len(id)),
		attribute.String("security.validation_type", "id"),
	)

	s.obs.Info(ctx, "Starting ID validation",
		"id_length", len(id),
		"validation_type", "id")

	if id == "" {
		result.Valid = false
		result.Error = fmt.Errorf("ID cannot be empty")
		s.recordValidationFailure(ctx, EventTypeIDValidation, "empty_id", ThreatLevelLow, result.Error)
		return result
	}

	// Basic ID validation - alphanumeric and hyphens only
	idRegex := regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
	if !idRegex.MatchString(id) {
		result.Valid = false
		result.Error = fmt.Errorf("invalid ID format: %s", id)
		s.recordValidationFailure(ctx, EventTypeIDValidation, "invalid_id_format", ThreatLevelMedium, result.Error)
		return result
	}

	// Check for potentially malicious ID patterns
	if containsMaliciousIDPatterns(id) {
		result.Valid = false
		result.Error = fmt.Errorf("ID contains potentially malicious patterns: %s", id)
		s.recordSecurityThreat(ctx, EventTypeMaliciousContentDetected, "malicious_id", ThreatLevelHigh, id)
		return result
	}

	duration := time.Since(start)
	s.recordValidationSuccess(ctx, EventTypeIDValidation, "id_validation", duration, result.Details)

	return result
}

// recordValidationSuccess records successful validation metrics and traces
func (s *SecurityValidator) recordValidationSuccess(ctx context.Context, eventType SecurityEventType, operation string, duration time.Duration, details map[string]interface{}) {
	// Record security event
	s.obs.RecordSecurityEvent(ctx, string(eventType), map[string]interface{}{
		"operation": operation,
		"status":    "success",
		"duration":  duration.String(),
		"details":   details,
	})

	// Record metrics using labels that match errorTotal metric definition
	// errorTotal expects: component, error_type, severity, knative_service_name
	s.obs.RecordMetric("counter", "security_validations_total", 1, map[string]string{
		"component":  "security",
		"error_type": operation,
		"severity":   "info",
		"event_type": string(eventType),
	})

	s.obs.RecordMetric("histogram", "security_validation_duration_seconds", duration.Seconds(), map[string]string{
		"component":  "security",
		"error_type": operation,
		"severity":   "info",
		"event_type": string(eventType),
	})

	s.obs.Info(ctx, "Security validation successful",
		"event_type", eventType,
		"operation", operation,
		"duration", duration.String())
}

// recordValidationFailure records failed validation metrics and traces
func (s *SecurityValidator) recordValidationFailure(ctx context.Context, eventType SecurityEventType, operation string, threatLevel SecurityThreatLevel, err error) {
	// Record security event
	s.obs.RecordSecurityEvent(ctx, string(eventType), map[string]interface{}{
		"operation":    operation,
		"status":       "failure",
		"threat_level": string(threatLevel),
		"error":        err.Error(),
	})

	// Record metrics using labels that match errorTotal metric definition
	// errorTotal expects: component, error_type, severity, knative_service_name
	s.obs.RecordMetric("counter", "security_validations_total", 1, map[string]string{
		"component":  "security",
		"error_type": operation,
		"severity":   string(threatLevel),
		"event_type": string(eventType),
	})

	s.obs.RecordMetric("counter", "security_validation_failures_total", 1, map[string]string{
		"component":  "security",
		"error_type": operation,
		"severity":   string(threatLevel),
		"event_type": string(eventType),
	})

	s.obs.Error(ctx, err, "Security validation failed",
		"event_type", eventType,
		"operation", operation,
		"threat_level", threatLevel)
}

// recordSecurityThreat records security threat detection metrics and traces
func (s *SecurityValidator) recordSecurityThreat(ctx context.Context, eventType SecurityEventType, operation string, threatLevel SecurityThreatLevel, content string) {
	// Record security event
	s.obs.RecordSecurityEvent(ctx, string(eventType), map[string]interface{}{
		"operation":      operation,
		"status":         "threat_detected",
		"threat_level":   string(threatLevel),
		"content_length": len(content),
	})

	// Record metrics using labels that match errorTotal metric definition
	// errorTotal expects: component, error_type, severity, knative_service_name
	s.obs.RecordMetric("counter", "security_threats_detected_total", 1, map[string]string{
		"component":  "security",
		"error_type": operation,
		"severity":   string(threatLevel),
		"event_type": string(eventType),
	})

	s.obs.RecordMetric("counter", "security_threats_by_level_total", 1, map[string]string{
		"component":    "security",
		"error_type":   "threat_detection",
		"severity":     string(threatLevel),
		"threat_level": string(threatLevel),
	})

	// Record high-priority alert for critical threats
	if threatLevel == ThreatLevelCritical {
		s.obs.RecordMetric("counter", "security_critical_threats_total", 1, map[string]string{
			"component":  "security",
			"error_type": operation,
			"severity":   "critical",
			"event_type": string(eventType),
		})
	}

	s.obs.Error(ctx, fmt.Errorf("security threat detected: %s", operation), "Security threat detected",
		"event_type", eventType,
		"operation", operation,
		"threat_level", threatLevel,
		"content_length", len(content))
}

// containsSQLInjection checks for potential SQL injection
func containsSQLInjection(input string) bool {
	sqlPatterns := []string{
		`(?i)(union|select|insert|update|delete|drop|create|alter)`,
		`(?i)(or|and)\s+\d+\s*=\s*\d+`,
		`(?i)(or|and)\s+['"]\w+['"]\s*=\s*['"]\w+['"]`,
		`(?i)(exec|execute|script)`,
	}

	for _, pattern := range sqlPatterns {
		matched, _ := regexp.MatchString(pattern, input)
		if matched {
			return true
		}
	}

	return false
}

// containsXSS checks for potential XSS attacks
func containsXSS(input string) bool {
	xssPatterns := []string{
		`(?i)<script[^>]*>.*?</script>`,
		`(?i)javascript:`,
		`(?i)on\w+\s*=`,
		`(?i)<iframe[^>]*>`,
		`(?i)<object[^>]*>`,
		`(?i)<embed[^>]*>`,
	}

	for _, pattern := range xssPatterns {
		matched, _ := regexp.MatchString(pattern, input)
		if matched {
			return true
		}
	}

	return false
}

// containsPathTraversal checks for potential path traversal attacks
func containsPathTraversal(input string) bool {
	// Check for literal path traversal patterns
	traversalPatterns := []string{
		"../",
		"..\\",
		"%2e%2e%2f",
		"%2e%2e%5c",
		"..%2f",
		"..%5c",
	}

	lowerInput := strings.ToLower(input)
	for _, pattern := range traversalPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}

	return false
}

// containsMaliciousImagePatterns checks for potentially malicious image name patterns
func containsMaliciousImagePatterns(imageName string) bool {
	maliciousPatterns := []string{
		`(?i)(eval|exec|system|shell)`,
		`(?i)(root|admin|privileged)`,
		`(?i)(backdoor|trojan|virus|malware)`,
		`(?i)(\.\.|%2e%2e)`, // Path traversal in image names
	}

	for _, pattern := range maliciousPatterns {
		matched, _ := regexp.MatchString(pattern, imageName)
		if matched {
			return true
		}
	}

	return false
}

// containsMaliciousContent checks for potentially malicious content
func containsMaliciousContent(content string) bool {
	maliciousPatterns := []string{
		`(?i)(eval|exec|system|shell)`,
		`(?i)(backdoor|trojan|virus|malware)`,
		`(?i)(rootkit|keylogger|spyware)`,
		`(?i)(<script|javascript:|vbscript:)`,
	}

	for _, pattern := range maliciousPatterns {
		matched, _ := regexp.MatchString(pattern, content)
		if matched {
			return true
		}
	}

	return false
}

// containsMaliciousIDPatterns checks for potentially malicious ID patterns
func containsMaliciousIDPatterns(id string) bool {
	maliciousPatterns := []string{
		`(?i)(admin|root|system)`,
		`(?i)(test|debug|dev)`,
		`(?i)(\.\.|%2e%2e)`,         // Path traversal
		`(?i)(union|select|insert)`, // SQL injection
	}

	for _, pattern := range maliciousPatterns {
		matched, _ := regexp.MatchString(pattern, id)
		if matched {
			return true
		}
	}

	return false
}

// isReservedNamespace checks if a namespace is reserved
func isReservedNamespace(namespace string) bool {
	reservedNamespaces := []string{
		"kube-system",
		"kube-public",
		"kube-node-lease",
		"default",
		"kube-flannel",
		"ingress-nginx",
		"cert-manager",
		"monitoring",
		"logging",
	}

	for _, reserved := range reservedNamespaces {
		if namespace == reserved {
			return true
		}
	}

	return false
}
