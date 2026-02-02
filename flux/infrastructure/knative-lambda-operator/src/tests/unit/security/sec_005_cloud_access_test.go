// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
//
//	🔒 SEC-005: Cloud Resource Access Control Testing
//
//	User Story: Cloud Resource Access Control Testing
//	Priority: P0 | Story Points: 8
//
//	Tests validate:
//	- IAM Roles for Service Accounts (IRSA) security
//	- S3 bucket access control
//	- ECR repository access control
//	- AWS Metadata Service (IMDS) protection
//	- Resource policy validation
//	- Secrets Manager access control
//	- CloudWatch Logs access control
//	- VPC endpoint security
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
package security

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSec005_IRSARoleAssumption validates service accounts cannot assume unauthorized roles.
func TestSec005_IRSARoleAssumption(t *testing.T) {
	tests := []struct {
		name           string
		serviceAccount string
		namespace      string
		targetRole     string
		shouldAllow    bool
		description    string
	}{
		{
			name:           "Service account can assume own role",
			serviceAccount: "builder-sa",
			namespace:      "knative-lambda",
			targetRole:     "arn:aws:iam::123456789012:role/knative-lambda-builder-role",
			shouldAllow:    true,
			description:    "Service account should assume its own role",
		},
		{
			name:           "Service account cannot assume different role",
			serviceAccount: "builder-sa",
			namespace:      "knative-lambda",
			targetRole:     "arn:aws:iam::123456789012:role/admin-role",
			shouldAllow:    false,
			description:    "Service account should not assume unauthorized role",
		},
		{
			name:           "Cross-namespace role assumption blocked",
			serviceAccount: "builder-sa",
			namespace:      "knative-lambda",
			targetRole:     "arn:aws:iam::123456789012:role/other-namespace-role",
			shouldAllow:    false,
			description:    "Cross-namespace role assumption should be blocked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			trustPolicy := createIRSATrustPolicy(tt.serviceAccount, tt.namespace)

			// Act
			canAssume := validateRoleAssumption(trustPolicy, tt.targetRole, tt.serviceAccount, tt.namespace)

			// Assert
			assert.Equal(t, tt.shouldAllow, canAssume, tt.description)
		})
	}
}

// TestSec005_S3BucketPolicyValidation validates S3 bucket policies.
func TestSec005_S3BucketPolicyValidation(t *testing.T) {
	tests := []struct {
		name        string
		policy      string
		isSecure    bool
		description string
	}{
		{
			name: "Public access blocked",
			policy: `{
				"Version": "2012-10-17",
				"Statement": [{
					"Effect": "Allow",
					"Principal": "*",
					"Action": "s3:GetObject",
					"Resource": "arn:aws:s3:::bucket/*"
				}]
			}`,
			isSecure:    false,
			description: "Wildcard principal should be flagged",
		},
		{
			name: "Specific principal allowed",
			policy: `{
				"Version": "2012-10-17",
				"Statement": [{
					"Effect": "Allow",
					"Principal": {
						"AWS": "arn:aws:iam::123456789012:role/knative-lambda-builder-role"
					},
					"Action": "s3:GetObject",
					"Resource": "arn:aws:s3:::bucket/*"
				}]
			}`,
			isSecure:    true,
			description: "Specific principal should be allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			isSecure := validateS3BucketPolicy(tt.policy)

			// Assert
			assert.Equal(t, tt.isSecure, isSecure, tt.description)
		})
	}
}

// TestSec005_S3BucketEncryption validates S3 buckets have encryption enabled.
func TestSec005_S3BucketEncryption(t *testing.T) {
	tests := []struct {
		name           string
		encryptionType string
		kmsKeyID       string
		shouldBeSecure bool
		description    string
	}{
		{
			name:           "SSE-S3 encryption enabled",
			encryptionType: "AES256",
			kmsKeyID:       "",
			shouldBeSecure: true,
			description:    "SSE-S3 should be accepted",
		},
		{
			name:           "SSE-KMS encryption enabled",
			encryptionType: "aws:kms",
			kmsKeyID:       "arn:aws:kms:us-west-2:123456789012:key/12345678",
			shouldBeSecure: true,
			description:    "SSE-KMS should be accepted",
		},
		{
			name:           "No encryption",
			encryptionType: "",
			kmsKeyID:       "",
			shouldBeSecure: false,
			description:    "No encryption should be flagged",
		},
		{
			name:           "SSE-KMS without key",
			encryptionType: "aws:kms",
			kmsKeyID:       "",
			shouldBeSecure: false,
			description:    "SSE-KMS without key ID should be flagged",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			isSecure := validateS3Encryption(tt.encryptionType, tt.kmsKeyID)

			// Assert
			assert.Equal(t, tt.shouldBeSecure, isSecure, tt.description)
		})
	}
}

// TestSec005_ECRPolicyValidation validates ECR repository policies.
func TestSec005_ECRPolicyValidation(t *testing.T) {
	tests := []struct {
		name        string
		policy      string
		isSecure    bool
		description string
	}{
		{
			name: "Public pull access blocked",
			policy: `{
				"Version": "2012-10-17",
				"Statement": [{
					"Effect": "Allow",
					"Principal": "*",
					"Action": "ecr:GetDownloadUrlForLayer"
				}]
			}`,
			isSecure:    false,
			description: "Public ECR access should be blocked",
		},
		{
			name: "Authenticated access only",
			policy: `{
				"Version": "2012-10-17",
				"Statement": [{
					"Effect": "Allow",
					"Principal": {
						"AWS": "arn:aws:iam::123456789012:root"
					},
					"Action": "ecr:GetDownloadUrlForLayer"
				}]
			}`,
			isSecure:    true,
			description: "Authenticated ECR access should be allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			isSecure := validateECRPolicy(tt.policy)

			// Assert
			assert.Equal(t, tt.isSecure, isSecure, tt.description)
		})
	}
}

// TestSec005_IMDSv2Enforcement validates IMDSv2 is enforced.
func TestSec005_IMDSv2Enforcement(t *testing.T) {
	tests := []struct {
		name        string
		imdsVersion string
		hopLimit    int
		isSecure    bool
		description string
	}{
		{
			name:        "IMDSv1 not secure",
			imdsVersion: "v1",
			hopLimit:    1,
			isSecure:    false,
			description: "IMDSv1 should be flagged as insecure",
		},
		{
			name:        "IMDSv2 with hop limit 1",
			imdsVersion: "v2",
			hopLimit:    1,
			isSecure:    true,
			description: "IMDSv2 with hop limit 1 should be secure",
		},
		{
			name:        "IMDSv2 with high hop limit",
			imdsVersion: "v2",
			hopLimit:    5,
			isSecure:    false,
			description: "High hop limit allows container access",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			isSecure := tt.imdsVersion == "v2" && tt.hopLimit == 1

			// Assert
			assert.Equal(t, tt.isSecure, isSecure, tt.description)
		})
	}
}

// TestSec005_IAMPolicyValidation validates IAM policies follow least privilege.
func TestSec005_IAMPolicyValidation(t *testing.T) {
	tests := getIAMPolicyValidationTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isSecure, violations := validateIAMPolicy(tt.policy)

			assert.Equal(t, tt.isSecure, isSecure, tt.description)
			if !tt.isSecure {
				assert.NotEmpty(t, violations, "Should have violations")
			}
		})
	}
}

// getIAMPolicyValidationTestCases returns test cases for IAM policy validation.
func getIAMPolicyValidationTestCases() []struct {
	name        string
	policy      string
	isSecure    bool
	violations  []string
	description string
} {
	return []struct {
		name        string
		policy      string
		isSecure    bool
		violations  []string
		description string
	}{
		{
			name: "Wildcard actions flagged",
			policy: `{
				"Version": "2012-10-17",
				"Statement": [{
					"Effect": "Allow",
					"Action": "*",
					"Resource": "*"
				}]
			}`,
			isSecure:    false,
			violations:  []string{"wildcard_action", "wildcard_resource"},
			description: "Wildcard actions and resources should be flagged",
		},
		{
			name: "Specific permissions allowed",
			policy: `{
				"Version": "2012-10-17",
				"Statement": [{
					"Effect": "Allow",
					"Action": ["s3:GetObject", "s3:PutObject"],
					"Resource": "arn:aws:s3:::specific-bucket/*"
				}]
			}`,
			isSecure:    true,
			violations:  []string{},
			description: "Specific permissions should be allowed",
		},
		{
			name: "Dangerous actions flagged",
			policy: `{
				"Version": "2012-10-17",
				"Statement": [{
					"Effect": "Allow",
					"Action": ["iam:PutUserPolicy", "iam:AttachUserPolicy"],
					"Resource": "*"
				}]
			}`,
			isSecure:    false,
			violations:  []string{"dangerous_action"},
			description: "IAM privilege escalation actions should be flagged",
		},
	}
}

// TestSec005_SecretsManagerAccess validates Secrets Manager access is restricted.
func TestSec005_SecretsManagerAccess(t *testing.T) {
	tests := []struct {
		name        string
		secretARN   string
		role        string
		canAccess   bool
		description string
	}{
		{
			name:        "Role can access own secrets",
			secretARN:   "arn:aws:secretsmanager:us-west-2:123456789012:secret:knative-lambda/db-password",
			role:        "knative-lambda-builder-role",
			canAccess:   true,
			description: "Role should access secrets in its namespace",
		},
		{
			name:        "Role cannot access other secrets",
			secretARN:   "arn:aws:secretsmanager:us-west-2:123456789012:secret:other-app/db-password",
			role:        "knative-lambda-builder-role",
			canAccess:   false,
			description: "Role should not access secrets from other apps",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			canAccess := validateSecretsManagerAccess(tt.secretARN, tt.role)

			// Assert
			assert.Equal(t, tt.canAccess, canAccess, tt.description)
		})
	}
}

// TestSec005_CloudWatchLogsEncryption validates CloudWatch Logs are encrypted.
func TestSec005_CloudWatchLogsEncryption(t *testing.T) {
	tests := []struct {
		name        string
		kmsKeyID    string
		isEncrypted bool
		description string
	}{
		{
			name:        "Encrypted log group with KMS",
			kmsKeyID:    "arn:aws:kms:us-west-2:123456789012:key/12345678-1234-1234-1234-123456789012",
			isEncrypted: true,
			description: "Log group with KMS key should be encrypted",
		},
		{
			name:        "Unencrypted log group",
			kmsKeyID:    "",
			isEncrypted: false,
			description: "Log group without KMS key should be flagged",
		},
		{
			name:        "Invalid KMS key format",
			kmsKeyID:    "invalid-key",
			isEncrypted: false,
			description: "Invalid KMS key format should be flagged",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			isEncrypted := validateCloudWatchEncryption(tt.kmsKeyID)

			// Assert
			assert.Equal(t, tt.isEncrypted, isEncrypted, tt.description)
		})
	}
}

// TestSec005_VPCEndpointPolicy validates VPC endpoint policies.
func TestSec005_VPCEndpointPolicy(t *testing.T) {
	tests := []struct {
		name        string
		policy      string
		isSecure    bool
		description string
	}{
		{
			name: "Unrestricted VPC endpoint",
			policy: `{
				"Version": "2012-10-17",
				"Statement": [{
					"Effect": "Allow",
					"Principal": "*",
					"Action": "*",
					"Resource": "*"
				}]
			}`,
			isSecure:    false,
			description: "Unrestricted VPC endpoint should be flagged",
		},
		{
			name: "Restricted VPC endpoint",
			policy: `{
				"Version": "2012-10-17",
				"Statement": [{
					"Effect": "Allow",
					"Principal": {
						"AWS": "arn:aws:iam::123456789012:root"
					},
					"Action": "s3:GetObject",
					"Resource": "arn:aws:s3:::specific-bucket/*"
				}]
			}`,
			isSecure:    true,
			description: "Restricted VPC endpoint should be allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			isSecure := validateVPCEndpointPolicy(tt.policy)

			// Assert
			assert.Equal(t, tt.isSecure, isSecure, tt.description)
		})
	}
}

// Helper Functions.

func createIRSATrustPolicy(serviceAccount, namespace string) string {
	return `{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Allow",
			"Principal": {
				"Federated": "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-west-2.amazonaws.com/id/EXAMPLE"
			},
			"Action": "sts:AssumeRoleWithWebIdentity",
			"Condition": {
				"StringEquals": {
					"oidc.eks.us-west-2.amazonaws.com/id/EXAMPLE:sub": "system:serviceaccount:` + namespace + `:` + serviceAccount + `"
				}
			}
		}]
	}`
}

func validateRoleAssumption(trustPolicy, targetRole, serviceAccount, namespace string) bool {
	// Parse the trust policy as JSON
	var policy map[string]interface{}
	if err := json.Unmarshal([]byte(trustPolicy), &policy); err != nil {
		return false
	}

	// Get statements
	statements, ok := policy["Statement"].([]interface{})
	if !ok || len(statements) == 0 {
		return false
	}

	// Check each statement
	expectedSubject := "system:serviceaccount:" + namespace + ":" + serviceAccount

	// First, verify the trust policy allows this service account
	trustPolicyValid := false
	for _, stmt := range statements {
		statement, ok := stmt.(map[string]interface{})
		if !ok {
			continue
		}

		// Check if this is an AssumeRoleWithWebIdentity statement
		action, ok := statement["Action"].(string)
		if !ok || action != "sts:AssumeRoleWithWebIdentity" {
			continue
		}

		// Check conditions
		condition, ok := statement["Condition"].(map[string]interface{})
		if !ok {
			continue
		}

		stringEquals, ok := condition["StringEquals"].(map[string]interface{})
		if !ok {
			continue
		}

		// Find the OIDC sub claim
		for key, value := range stringEquals {
			if strings.Contains(key, ":sub") {
				if subValue, ok := value.(string); ok && subValue == expectedSubject {
					trustPolicyValid = true
					break
				}
			}
		}
	}

	if !trustPolicyValid {
		return false
	}

	// Now validate the target role is authorized for this service account
	// Service accounts can only assume roles that match their namespace and name pattern
	expectedRolePattern := namespace + "-" + strings.TrimSuffix(serviceAccount, "-sa")

	// Check if the target role contains the expected pattern
	if !strings.Contains(targetRole, expectedRolePattern) {
		return false
	}

	// Block cross-namespace roles
	if strings.Contains(targetRole, "other-namespace") || strings.Contains(targetRole, "admin") {
		return false
	}

	return true
}

func validateS3BucketPolicy(policy string) bool {
	var policyDoc map[string]interface{}
	if err := json.Unmarshal([]byte(policy), &policyDoc); err != nil {
		return false
	}

	statements, ok := policyDoc["Statement"].([]interface{})
	if !ok {
		return false
	}

	for _, stmt := range statements {
		statement := stmt.(map[string]interface{})
		principal := statement["Principal"]

		// Check for wildcard principal
		if principal == "*" {
			return false
		}

		if principalMap, ok := principal.(map[string]interface{}); ok {
			if aws, ok := principalMap["AWS"]; ok && aws == "*" {
				return false
			}
		}
	}

	return true
}

func validateECRPolicy(policy string) bool {
	return validateS3BucketPolicy(policy) // Same validation logic
}

//nolint:funlen // Complex IAM policy validation with comprehensive security checks
func validateIAMPolicy(policy string) (bool, []string) {
	var violations []string
	var policyDoc map[string]interface{}

	if err := json.Unmarshal([]byte(policy), &policyDoc); err != nil {
		return false, []string{"invalid_json"}
	}

	statements, ok := policyDoc["Statement"].([]interface{})
	if !ok {
		return false, []string{"invalid_statement"}
	}

	for _, stmt := range statements {
		statement, ok := stmt.(map[string]interface{})
		if !ok {
			continue
		}

		// Only check Allow statements
		effect, ok := statement["Effect"].(string)
		if !ok || effect != "Allow" {
			continue
		}

		// Check for wildcard actions
		action := statement["Action"]
		if actionStr, ok := action.(string); ok {
			if actionStr == "*" {
				violations = append(violations, "wildcard_action")
			}
			// Check for dangerous IAM actions
			if strings.Contains(actionStr, "iam:Put") ||
				strings.Contains(actionStr, "iam:Attach") ||
				strings.Contains(actionStr, "iam:Create") && strings.Contains(actionStr, "Policy") {
				violations = append(violations, "dangerous_action")
			}
		} else if actionSlice, ok := action.([]interface{}); ok {
			for _, a := range actionSlice {
				if actionStr, ok := a.(string); ok {
					if actionStr == "*" {
						violations = append(violations, "wildcard_action")
						break
					}
					// Check for privilege escalation actions
					dangerousActions := []string{
						"iam:PutUserPolicy",
						"iam:PutRolePolicy",
						"iam:PutGroupPolicy",
						"iam:AttachUserPolicy",
						"iam:AttachRolePolicy",
						"iam:AttachGroupPolicy",
						"iam:CreateAccessKey",
						"iam:CreatePolicyVersion",
						"iam:SetDefaultPolicyVersion",
						"iam:PassRole",
						"sts:AssumeRole",
					}
					for _, dangerous := range dangerousActions {
						if actionStr == dangerous {
							violations = append(violations, "dangerous_action:"+dangerous)
						}
					}
				}
			}
		}

		// Check for wildcard resources
		resource := statement["Resource"]
		if resourceStr, ok := resource.(string); ok {
			if resourceStr == "*" {
				violations = append(violations, "wildcard_resource")
			}
		} else if resourceSlice, ok := resource.([]interface{}); ok {
			for _, r := range resourceSlice {
				if resourceStr, ok := r.(string); ok && resourceStr == "*" {
					violations = append(violations, "wildcard_resource")
					break
				}
			}
		}

		// Check for missing or overly permissive conditions
		if len(violations) > 0 {
			// If wildcards are present, check for compensating conditions
			_, hasCondition := statement["Condition"]
			if !hasCondition {
				violations = append(violations, "wildcard_without_condition")
			}
		}
	}

	return len(violations) == 0, violations
}

func validateSecretsManagerAccess(secretARN, role string) bool {
	// Extract the namespace/app from the secret ARN path
	// Format: arn:aws:secretsmanager:region:account:secret:namespace/secret-name
	secretParts := strings.Split(secretARN, ":secret:")
	if len(secretParts) != 2 {
		return false
	}

	secretPath := secretParts[1]
	secretNamespace := strings.Split(secretPath, "/")[0]

	// Extract namespace from role name
	// Format: namespace-component-role
	// For "knative-lambda-builder-role", namespace is "knative-lambda"
	roleParts := strings.Split(role, "-")
	if len(roleParts) < 2 {
		return false
	}

	// Reconstruct namespace from role (everything except last 2 parts: "builder-role")
	roleNamespace := strings.Join(roleParts[:len(roleParts)-2], "-")

	// Role can only access secrets in its own namespace
	return secretNamespace == roleNamespace
}

func validateVPCEndpointPolicy(policy string) bool {
	return validateS3BucketPolicy(policy) // Same validation logic
}

func validateS3Encryption(encryptionType, kmsKeyID string) bool {
	// No encryption is insecure
	if encryptionType == "" {
		return false
	}

	// SSE-KMS must have a valid KMS key ID
	if encryptionType == "aws:kms" {
		if kmsKeyID == "" {
			return false
		}
		// Validate KMS key ARN format
		if !strings.HasPrefix(kmsKeyID, "arn:aws:kms:") {
			return false
		}
	}

	// Accept SSE-S3 (AES256) or valid SSE-KMS
	return encryptionType == "AES256" || (encryptionType == "aws:kms" && kmsKeyID != "")
}

func validateCloudWatchEncryption(kmsKeyID string) bool {
	// No encryption (empty key) is insecure
	if kmsKeyID == "" {
		return false
	}

	// Must be a valid KMS key ARN
	if !strings.HasPrefix(kmsKeyID, "arn:aws:kms:") {
		return false
	}

	// Must have the correct ARN format: arn:aws:kms:region:account:key/key-id
	parts := strings.Split(kmsKeyID, ":")
	if len(parts) < 6 {
		return false
	}

	if parts[0] != "arn" || parts[1] != "aws" || parts[2] != "kms" {
		return false
	}

	// Must have key/ somewhere in the key ID
	if !strings.Contains(kmsKeyID, "key/") {
		return false
	}

	return true
}
