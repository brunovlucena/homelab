// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🔍 NAMING HELPERS - Kubernetes Resource Naming Utilities
//
//	🎯 Purpose: Generate valid Kubernetes resource names and sanitize labels
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package helpers

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"
)

// GenerateJobName - Generate unique job name based on parser and timestamp
func GenerateJobName(thirdPartyID, parserID string) string {
	// Include timestamp to ensure uniqueness and prevent race conditions
	timestamp := time.Now().UnixNano()
	// Create a hash of the parser ID and timestamp to ensure uniqueness
	hashInput := fmt.Sprintf("%s-%s-%d", thirdPartyID, parserID, timestamp)
	hash := sha256.Sum256([]byte(hashInput))
	hashStr := fmt.Sprintf("%x", hash)[:8]
	return generateName("build", []string{thirdPartyID, parserID, hashStr}, 63)
}

// GenerateServiceName - Generate valid Kubernetes service name
func GenerateServiceName(thirdPartyID, parserID string) string {
	shortThirdPartyID := truncate(thirdPartyID, 15)
	shortParserID := truncate(parserID, 15)
	return generateName("lambda", []string{shortThirdPartyID, shortParserID}, 55)
}

// GenerateFunctionName - Generate function name for Lambda functions
func GenerateFunctionName(thirdPartyID, parserID string) string {
	return generateName("lambda", []string{thirdPartyID, parserID}, 63)
}

// SanitizeLabelValue - Sanitize value for Kubernetes label
func SanitizeLabelValue(value string) string {
	if value == "" {
		return "empty"
	}

	sanitized := sanitizeName(value, 63)
	sanitized = strings.Trim(sanitized, "-")

	if sanitized == "" {
		return "default"
	}

	return sanitized
}

// generateName - Generate and sanitize a name with given prefix and parts
func generateName(prefix string, parts []string, maxLength int) string {
	name := fmt.Sprintf("%s-%s", prefix, strings.Join(parts, "-"))
	return sanitizeName(name, maxLength)
}

// sanitizeName - Sanitize string for Kubernetes name requirements
func sanitizeName(name string, maxLength int) string {
	// Convert to lowercase and replace invalid characters
	sanitized := strings.ToLower(name)
	sanitized = strings.ReplaceAll(sanitized, "_", "-")
	sanitized = strings.ReplaceAll(sanitized, ".", "-")

	// Keep only alphanumeric and hyphens
	var result strings.Builder
	for _, char := range sanitized {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
			result.WriteRune(char)
		}
	}
	sanitized = result.String()

	// Truncate and trim
	if len(sanitized) > maxLength {
		sanitized = sanitized[:maxLength]
	}

	return strings.TrimRight(sanitized, "-_.")
}

// truncate - Truncate string to specified length
func truncate(input string, maxLength int) string {
	if len(input) > maxLength {
		return input[:maxLength]
	}
	return input
}
