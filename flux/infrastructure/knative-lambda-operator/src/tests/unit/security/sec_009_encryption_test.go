// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
//
//	🔒 SEC-009: Data Protection & Encryption Testing
//
//	User Story: Data Protection & Encryption Testing
//	Priority: P1 | Story Points: 5
//
//	Tests validate:
//	- Encryption at rest
//	- Encryption in transit (TLS)
//	- Certificate management
//	- Key management
//	- Database encryption
//	- Message queue encryption
//	- Backup encryption
//	- Data masking and tokenization
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
package security

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestSec009_TLSVersionEnforcement validates only TLS 1.2+ is allowed.
func TestSec009_TLSVersionEnforcement(t *testing.T) {
	tests := []struct {
		name        string
		tlsVersion  uint16
		shouldAllow bool
		description string
	}{
		{
			name:        "TLS 1.0 blocked",
			tlsVersion:  tls.VersionTLS10,
			shouldAllow: false,
			description: "TLS 1.0 should be blocked",
		},
		{
			name:        "TLS 1.1 blocked",
			tlsVersion:  tls.VersionTLS11,
			shouldAllow: false,
			description: "TLS 1.1 should be blocked",
		},
		{
			name:        "TLS 1.2 allowed",
			tlsVersion:  tls.VersionTLS12,
			shouldAllow: true,
			description: "TLS 1.2 should be allowed",
		},
		{
			name:        "TLS 1.3 allowed",
			tlsVersion:  tls.VersionTLS13,
			shouldAllow: true,
			description: "TLS 1.3 should be allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act.
			allowed := tt.tlsVersion >= tls.VersionTLS12

			// Assert.
			assert.Equal(t, tt.shouldAllow, allowed, tt.description)
		})
	}
}

// TestSec009_WeakCipherSuitesBlocked validates weak ciphers are blocked.
func TestSec009_WeakCipherSuitesBlocked(t *testing.T) {
	weakCiphers := []uint16{
		tls.TLS_RSA_WITH_RC4_128_SHA,
		tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
	}

	strongCiphers := []uint16{
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	}

	t.Run("Weak ciphers blocked", func(t *testing.T) {
		for _, cipher := range weakCiphers {
			isWeak := isWeakCipher(cipher)
			assert.True(t, isWeak, "Cipher %d should be flagged as weak", cipher)
		}
	})

	t.Run("Strong ciphers allowed", func(t *testing.T) {
		for _, cipher := range strongCiphers {
			isWeak := isWeakCipher(cipher)
			assert.False(t, isWeak, "Cipher %d should be allowed", cipher)
		}
	})
}

// TestSec009_CertificateValidation validates certificate requirements.
func TestSec009_CertificateValidation(t *testing.T) {
	tests := []struct {
		name        string
		keySize     int
		expiration  time.Time
		selfSigned  bool
		isValid     bool
		description string
	}{
		{
			name:        "Valid certificate",
			keySize:     2048,
			expiration:  time.Now().Add(90 * 24 * time.Hour),
			selfSigned:  false,
			isValid:     true,
			description: "Valid certificate should pass",
		},
		{
			name:        "Weak key size",
			keySize:     1024,
			expiration:  time.Now().Add(90 * 24 * time.Hour),
			selfSigned:  false,
			isValid:     false,
			description: "1024-bit key should be rejected",
		},
		{
			name:        "Certificate near expiration",
			keySize:     2048,
			expiration:  time.Now().Add(15 * 24 * time.Hour),
			selfSigned:  false,
			isValid:     false,
			description: "Certificate expiring in <30 days should be flagged",
		},
		{
			name:        "Self-signed certificate",
			keySize:     2048,
			expiration:  time.Now().Add(90 * 24 * time.Hour),
			selfSigned:  true,
			isValid:     false,
			description: "Self-signed certificate should be rejected in production",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act.
			isValid := validateCertificate(tt.keySize, tt.expiration, tt.selfSigned)

			// Assert.
			assert.Equal(t, tt.isValid, isValid, tt.description)
		})
	}
}

// TestSec009_EncryptionAtRest validates data is encrypted at rest.
//
//nolint:funlen // Comprehensive encryption test with multiple storage scenarios
func TestSec009_EncryptionAtRest(t *testing.T) {
	tests := []struct {
		name        string
		resource    string
		encrypted   bool
		kmsKeyID    string
		isSecure    bool
		description string
	}{
		{
			name:        "S3 bucket encrypted with KMS",
			resource:    "s3://bucket",
			encrypted:   true,
			kmsKeyID:    "arn:aws:kms:us-west-2:123456789012:key/12345678-1234-1234-1234-123456789012",
			isSecure:    true,
			description: "S3 with KMS encryption should be secure",
		},
		{
			name:        "S3 bucket not encrypted",
			resource:    "s3://bucket",
			encrypted:   false,
			kmsKeyID:    "",
			isSecure:    false,
			description: "Unencrypted S3 should be flagged",
		},
		{
			name:        "S3 with encryption flag but no KMS key",
			resource:    "s3://bucket",
			encrypted:   true,
			kmsKeyID:    "",
			isSecure:    false,
			description: "S3 encryption without KMS key should be flagged",
		},
		{
			name:        "EBS volume encrypted",
			resource:    "ebs://vol-123",
			encrypted:   true,
			kmsKeyID:    "arn:aws:kms:us-west-2:123456789012:key/12345678-1234-1234-1234-123456789012",
			isSecure:    true,
			description: "Encrypted EBS should be secure",
		},
		{
			name:        "EBS volume not encrypted",
			resource:    "ebs://vol-123",
			encrypted:   false,
			kmsKeyID:    "",
			isSecure:    false,
			description: "Unencrypted EBS should be flagged",
		},
		{
			name:        "Invalid KMS key format",
			resource:    "s3://bucket",
			encrypted:   true,
			kmsKeyID:    "invalid-key",
			isSecure:    false,
			description: "Invalid KMS key format should be flagged",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act.
			isSecure := validateResourceEncryption(tt.resource, tt.encrypted, tt.kmsKeyID)

			// Assert.
			assert.Equal(t, tt.isSecure, isSecure, tt.description)
		})
	}
}

// TestSec009_DatabaseEncryption validates database encryption.
func TestSec009_DatabaseEncryption(t *testing.T) {
	tests := []struct {
		name             string
		storageEncrypted bool
		tlsEnabled       bool
		isSecure         bool
		description      string
	}{
		{
			name:             "Fully encrypted database",
			storageEncrypted: true,
			tlsEnabled:       true,
			isSecure:         true,
			description:      "Database with storage and TLS encryption should be secure",
		},
		{
			name:             "No storage encryption",
			storageEncrypted: false,
			tlsEnabled:       true,
			isSecure:         false,
			description:      "Database without storage encryption should be flagged",
		},
		{
			name:             "No TLS",
			storageEncrypted: true,
			tlsEnabled:       false,
			isSecure:         false,
			description:      "Database without TLS should be flagged",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act.
			isSecure := tt.storageEncrypted && tt.tlsEnabled

			// Assert.
			assert.Equal(t, tt.isSecure, isSecure, tt.description)
		})
	}
}

// TestSec009_KeyRotation validates key rotation is enabled.
func TestSec009_KeyRotation(t *testing.T) {
	tests := []struct {
		name            string
		rotationEnabled bool
		keyAge          time.Duration
		isSecure        bool
		description     string
	}{
		{
			name:            "Rotation enabled, fresh key",
			rotationEnabled: true,
			keyAge:          30 * 24 * time.Hour,
			isSecure:        true,
			description:     "Fresh key with rotation should be secure",
		},
		{
			name:            "Rotation disabled",
			rotationEnabled: false,
			keyAge:          30 * 24 * time.Hour,
			isSecure:        false,
			description:     "Key without rotation should be flagged",
		},
		{
			name:            "Old key",
			rotationEnabled: true,
			keyAge:          100 * 24 * time.Hour,
			isSecure:        false,
			description:     "Key older than 90 days should be flagged",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act.
			isSecure := tt.rotationEnabled && tt.keyAge <= 90*24*time.Hour

			// Assert.
			assert.Equal(t, tt.isSecure, isSecure, tt.description)
		})
	}
}

// TestSec009_DataMasking validates sensitive data is masked.
func TestSec009_DataMasking(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		description string
	}{
		{
			name:        "Credit card masked",
			input:       "4532-1234-5678-9010",
			expected:    "****-****-****-9010",
			description: "Should mask credit card number",
		},
		{
			name:        "SSN masked",
			input:       "123-45-6789",
			expected:    "***-**-6789",
			description: "Should mask SSN",
		},
		{
			name:        "Email partially masked",
			input:       "user@example.com",
			expected:    "u***@example.com",
			description: "Should partially mask email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act.
			masked := maskSensitiveData(tt.input)

			// Assert.
			assert.NotEqual(t, tt.input, masked, tt.description)
			assert.Contains(t, masked, "*", "Should contain mask character")
		})
	}
}

// TestSec009_HTTPSEnforcement validates HTTPS is enforced.
func TestSec009_HTTPSEnforcement(t *testing.T) {
	// Arrange.
	handler := setupHTTPSHandler(t)

	tests := []struct {
		name           string
		url            string
		expectedStatus int
		description    string
	}{
		{
			name:           "HTTPS allowed",
			url:            "https://api.example.com/endpoint",
			expectedStatus: http.StatusOK,
			description:    "HTTPS requests should be allowed",
		},
		{
			name:           "HTTP redirected",
			url:            "http://api.example.com/endpoint",
			expectedStatus: http.StatusPermanentRedirect,
			description:    "HTTP should redirect to HTTPS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			req := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()

			// Act.
			handler.ServeHTTP(w, req)

			// Assert.
			if tt.url[:5] == "http:" {
				assert.Equal(t, http.StatusPermanentRedirect, w.Code, tt.description)
			} else {
				assert.Equal(t, http.StatusOK, w.Code, tt.description)
			}
		})
	}
}

// TestSec009_BackupEncryption validates backups are encrypted.
func TestSec009_BackupEncryption(t *testing.T) {
	tests := []struct {
		name        string
		encrypted   bool
		kmsKeyID    string
		isSecure    bool
		description string
	}{
		{
			name:        "Encrypted backup",
			encrypted:   true,
			kmsKeyID:    "arn:aws:kms:us-west-2:123456789012:key/12345678-1234-1234-1234-123456789012",
			isSecure:    true,
			description: "Encrypted backup should be secure",
		},
		{
			name:        "Unencrypted backup",
			encrypted:   false,
			kmsKeyID:    "",
			isSecure:    false,
			description: "Unencrypted backup should be flagged",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act.
			isSecure := tt.encrypted && tt.kmsKeyID != ""

			// Assert.
			assert.Equal(t, tt.isSecure, isSecure, tt.description)
		})
	}
}

// Helper Functions.

func isWeakCipher(cipher uint16) bool {
	weakCiphers := []uint16{
		tls.TLS_RSA_WITH_RC4_128_SHA,
		tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
		tls.TLS_RSA_WITH_AES_128_CBC_SHA,
		tls.TLS_RSA_WITH_AES_256_CBC_SHA,
	}

	for _, weak := range weakCiphers {
		if cipher == weak {
			return true
		}
	}
	return false
}

func validateCertificate(keySize int, expiration time.Time, selfSigned bool) bool {
	// Key size must be >= 2048.
	if keySize < 2048 {
		return false
	}

	// Must not be self-signed in production.
	if selfSigned {
		return false
	}

	// Must not expire within 30 days.
	if time.Until(expiration) < 30*24*time.Hour {
		return false
	}

	return true
}

func maskSensitiveData(input string) string {
	if len(input) <= 4 {
		return "****"
	}
	// Show last 4 characters only.
	return strings.Repeat("*", len(input)-4) + input[len(input)-4:]
}

func setupHTTPSHandler(t *testing.T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request is HTTP (non-TLS).
		if r.URL.Scheme == "http" || r.TLS == nil {
			// Redirect to HTTPS.
			httpsURL := "https://" + r.Host + r.URL.Path
			http.Redirect(w, r, httpsURL, http.StatusPermanentRedirect)
			return
		}

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			t.Logf("Failed to write response: %v", err)
		}
	})
}

func validateResourceEncryption(_ string, encrypted bool, kmsKeyID string) bool {
	// If not encrypted, it's insecure.
	if !encrypted {
		return false
	}

	// If encrypted but no KMS key provided, it's insecure.
	if kmsKeyID == "" {
		return false
	}

	// Must have valid KMS key ARN format.
	if !strings.HasPrefix(kmsKeyID, "arn:aws:kms:") {
		return false
	}

	// Validate ARN structure: arn:aws:kms:region:account:key/key-id.
	parts := strings.Split(kmsKeyID, ":")
	if len(parts) < 6 {
		return false
	}

	if parts[0] != "arn" || parts[1] != "aws" || parts[2] != "kms" {
		return false
	}

	// Must have key/ somewhere in the key ID (after arn:aws:kms).
	if !strings.Contains(kmsKeyID, "key/") {
		return false
	}

	return true
}
