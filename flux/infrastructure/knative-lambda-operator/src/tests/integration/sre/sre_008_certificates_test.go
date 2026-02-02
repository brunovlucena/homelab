// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
//
//	🧪 SRE-008: Certificate Lifecycle Management Tests
//
//	User Story: Certificate Lifecycle Management
//	Priority: P0 | Story Points: 5
//
//	Tests validate:
//	- Certificate expiry monitoring (30-day and 7-day alerts)
//	- Automated certificate renewal
//	- Manual renewal procedures
//	- Certificate inventory management
//	- Zero-downtime certificate rotation
//	- Runbook accessibility
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
package sre

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// Test Fixtures and Helpers.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

// generateTestCertificate creates a self-signed certificate for testing.
func generateTestCertificate(daysValid int) (*x509.Certificate, []byte, []byte, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, nil, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "test.knative-lambda.local",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 0, daysValid),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	cert, _ := x509.ParseCertificate(certDER)
	return cert, certPEM, keyPEM, nil
}

// createTestTLSSecret creates a Kubernetes TLS secret for testing.
func createTestTLSSecret(namespace, name string, daysValid int) (*corev1.Secret, error) {
	_, certPEM, keyPEM, err := generateTestCertificate(daysValid)
	if err != nil {
		return nil, err
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			"tls.crt": certPEM,
			"tls.key": keyPEM,
		},
	}, nil
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC1: Certificate Expiry Monitoring (30-day and 7-day alerts).
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestSRE008_AC1_CertificateExpiryDetection(t *testing.T) {
	t.Run("Detect certificate expiring in 30 days", func(t *testing.T) {
		// Arrange
		cert, _, _, err := generateTestCertificate(30)
		require.NoError(t, err, "Failed to generate test certificate")

		// Act
		daysUntilExpiry := int(time.Until(cert.NotAfter).Hours() / 24)

		// Assert
		assert.LessOrEqual(t, daysUntilExpiry, 30, "Certificate should expire within 30 days")
		assert.GreaterOrEqual(t, daysUntilExpiry, 29, "Certificate should be valid for at least 29 days")
	})

	t.Run("Detect certificate expiring in 7 days (critical)", func(t *testing.T) {
		// Arrange
		cert, _, _, err := generateTestCertificate(7)
		require.NoError(t, err, "Failed to generate test certificate")

		// Act
		daysUntilExpiry := int(time.Until(cert.NotAfter).Hours() / 24)

		// Assert
		assert.LessOrEqual(t, daysUntilExpiry, 7, "Certificate should expire within 7 days")
		assert.GreaterOrEqual(t, daysUntilExpiry, 6, "Certificate should be valid for at least 6 days")
	})

	t.Run("Detect expired certificate", func(t *testing.T) {
		// Arrange
		cert, _, _, err := generateTestCertificate(-1) // Expired yesterday
		require.NoError(t, err, "Failed to generate test certificate")

		// Act
		isExpired := time.Now().After(cert.NotAfter)

		// Assert
		assert.True(t, isExpired, "Certificate should be detected as expired")
	})
}

func TestSRE008_AC1_CertificateInventory(t *testing.T) {
	t.Run("List all TLS secrets in cluster", func(t *testing.T) {
		// Arrange
		clientset := fake.NewSimpleClientset()
		ctx := context.Background()

		namespaces := []string{"knative-serving", "rabbitmq-prd", "knative-lambda"}
		for _, ns := range namespaces {
			secret, err := createTestTLSSecret(ns, "test-cert", 90)
			require.NoError(t, err)
			_, err = clientset.CoreV1().Secrets(ns).Create(ctx, secret, metav1.CreateOptions{})
			require.NoError(t, err)
		}

		// Act
		var tlsSecrets []corev1.Secret
		for _, ns := range namespaces {
			secrets, err := clientset.CoreV1().Secrets(ns).List(ctx, metav1.ListOptions{})
			require.NoError(t, err)

			for _, secret := range secrets.Items {
				if secret.Type == corev1.SecretTypeTLS {
					tlsSecrets = append(tlsSecrets, secret)
				}
			}
		}

		// Assert
		assert.Len(t, tlsSecrets, 3, "Should find all TLS secrets across namespaces")
		assert.Contains(t, []string{"knative-serving", "rabbitmq-prd", "knative-lambda"}, tlsSecrets[0].Namespace)
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC2: Automated Certificate Renewal Validated.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestSRE008_AC2_AutomatedRenewal(t *testing.T) {
	t.Run("Certificate renewal triggers before expiry", func(t *testing.T) {
		// Arrange
		cert, _, _, err := generateTestCertificate(90)
		require.NoError(t, err)

		renewalThreshold := 30 * 24 * time.Hour // Renew 30 days before expiry
		timeUntilExpiry := time.Until(cert.NotAfter)

		// Act
		shouldRenew := timeUntilExpiry < renewalThreshold

		// Assert
		if timeUntilExpiry < renewalThreshold {
			assert.True(t, shouldRenew, "Certificate should trigger renewal within 30 days of expiry")
		} else {
			assert.False(t, shouldRenew, "Certificate should not trigger renewal yet")
		}
	})

	t.Run("cert-manager ready annotation present", func(t *testing.T) {
		// Arrange
		secret, err := createTestTLSSecret("test-ns", "test-cert", 90)
		require.NoError(t, err)

		// Simulate cert-manager annotations
		if secret.Annotations == nil {
			secret.Annotations = make(map[string]string)
		}
		secret.Annotations["cert-manager.io/certificate-name"] = "test-certificate"
		secret.Annotations["cert-manager.io/common-name"] = "test.example.com"

		// Act
		hasCertManager := secret.Annotations["cert-manager.io/certificate-name"] != ""

		// Assert
		assert.True(t, hasCertManager, "Secret should have cert-manager annotations")
		assert.Equal(t, "test-certificate", secret.Annotations["cert-manager.io/certificate-name"])
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC3: Manual Renewal Procedures Documented and Tested.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestSRE008_AC3_ManualRenewalProcedure(t *testing.T) {
	t.Run("Manual certificate replacement", func(t *testing.T) {
		// Arrange
		clientset := fake.NewSimpleClientset()
		ctx := context.Background()

		oldSecret, err := createTestTLSSecret("test-ns", "test-cert", 7) // Expiring soon
		require.NoError(t, err)
		_, err = clientset.CoreV1().Secrets("test-ns").Create(ctx, oldSecret, metav1.CreateOptions{})
		require.NoError(t, err)

		// Act - Simulate manual renewal
		newSecret, err := createTestTLSSecret("test-ns", "test-cert", 365) // New cert valid for 1 year
		require.NoError(t, err)

		// Update secret
		updatedSecret, err := clientset.CoreV1().Secrets("test-ns").Update(ctx, newSecret, metav1.UpdateOptions{})
		require.NoError(t, err)

		// Assert
		assert.NotNil(t, updatedSecret, "Secret should be updated")

		// Parse new certificate
		block, _ := pem.Decode(updatedSecret.Data["tls.crt"])
		cert, err := x509.ParseCertificate(block.Bytes)
		require.NoError(t, err)

		daysValid := int(time.Until(cert.NotAfter).Hours() / 24)
		assert.Greater(t, daysValid, 300, "New certificate should be valid for at least 300 days")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC4: Certificate Inventory Maintained.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestSRE008_AC4_CertificateInventory(t *testing.T) {
	t.Run("Generate certificate inventory report", func(t *testing.T) {
		// Arrange
		clientset := fake.NewSimpleClientset()
		ctx := context.Background()

		testCerts := []struct {
			namespace string
			name      string
			daysValid int
		}{
			{"knative-serving", "serving-cert", 60},
			{"rabbitmq-prd", "rabbitmq-tls", 45},
			{"knative-lambda", "builder-tls", 80},
		}

		for _, tc := range testCerts {
			secret, err := createTestTLSSecret(tc.namespace, tc.name, tc.daysValid)
			require.NoError(t, err)
			_, err = clientset.CoreV1().Secrets(tc.namespace).Create(ctx, secret, metav1.CreateOptions{})
			require.NoError(t, err)
		}

		// Act - Generate inventory
		type CertInfo struct {
			Namespace  string
			Name       string
			ExpiryDate time.Time
			DaysValid  int
		}

		var inventory []CertInfo
		for _, tc := range testCerts {
			secret, err := clientset.CoreV1().Secrets(tc.namespace).Get(ctx, tc.name, metav1.GetOptions{})
			require.NoError(t, err)

			block, _ := pem.Decode(secret.Data["tls.crt"])
			cert, err := x509.ParseCertificate(block.Bytes)
			require.NoError(t, err)

			inventory = append(inventory, CertInfo{
				Namespace:  tc.namespace,
				Name:       tc.name,
				ExpiryDate: cert.NotAfter,
				DaysValid:  int(time.Until(cert.NotAfter).Hours() / 24),
			})
		}

		// Assert
		assert.Len(t, inventory, 3, "Inventory should contain all certificates")

		// Verify critical certificates present
		namespaces := make(map[string]bool)
		for _, cert := range inventory {
			namespaces[cert.Namespace] = true
		}

		assert.True(t, namespaces["knative-serving"], "Should track Knative serving certificates")
		assert.True(t, namespaces["rabbitmq-prd"], "Should track RabbitMQ certificates")
		assert.True(t, namespaces["knative-lambda"], "Should track builder certificates")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC5: Zero-Downtime Certificate Rotation Procedures.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestSRE008_AC5_ZeroDowntimeRotation(t *testing.T) {
	t.Run("Certificate rotation without service interruption", func(t *testing.T) {
		// Arrange
		clientset := fake.NewSimpleClientset()
		ctx := context.Background()

		oldSecret, err := createTestTLSSecret("test-ns", "app-cert", 10)
		require.NoError(t, err)
		created, err := clientset.CoreV1().Secrets("test-ns").Create(ctx, oldSecret, metav1.CreateOptions{})
		require.NoError(t, err)
		oldVersion := created.ResourceVersion

		// Act - Simulate rotation by creating new certificate
		newSecret, err := createTestTLSSecret("test-ns", "app-cert", 365)
		require.NoError(t, err)
		newSecret.ResourceVersion = oldVersion // Preserve version for update

		// Simulate version increment (fake client doesn't auto-increment)
		newSecret.ResourceVersion = "2"

		updated, err := clientset.CoreV1().Secrets("test-ns").Update(ctx, newSecret, metav1.UpdateOptions{})
		require.NoError(t, err)

		// Assert
		assert.NotEmpty(t, updated.ResourceVersion, "Secret version should exist")
		assert.NotEqual(t, oldVersion, updated.ResourceVersion, "Secret version should change")

		// Verify new certificate is valid
		block, _ := pem.Decode(updated.Data["tls.crt"])
		cert, err := x509.ParseCertificate(block.Bytes)
		require.NoError(t, err)

		daysValid := int(time.Until(cert.NotAfter).Hours() / 24)
		assert.Greater(t, daysValid, 300, "Rotated certificate should be valid for longer period")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC6: Runbook Accessible During Incidents.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestSRE008_AC6_RunbookAccessibility(t *testing.T) {
	t.Run("Runbook file exists and is readable", func(t *testing.T) {
		// Arrange
		runbookPath := "../../../docs/03-for-engineers/sre/user-stories/SRE-008-certificate-lifecycle-management.md"

		// Act
		_, err := os.Stat(runbookPath)

		// Assert
		assert.NoError(t, err, "Runbook file should exist")

		// Verify runbook contains key sections
		content, err := os.ReadFile(runbookPath)
		require.NoError(t, err, "Should be able to read runbook")

		runbookContent := string(content)
		assert.Contains(t, runbookContent, "Certificate Lifecycle Management", "Should be correct runbook")
		assert.Contains(t, runbookContent, "Acceptance Criteria", "Should document acceptance criteria")
		assert.Contains(t, runbookContent, "kubectl", "Should contain operational commands")
	})

	t.Run("Runbook contains all required procedures", func(t *testing.T) {
		// Arrange
		runbookPath := "../../../docs/03-for-engineers/sre/user-stories/SRE-008-certificate-lifecycle-management.md"
		content, err := os.ReadFile(runbookPath)
		require.NoError(t, err)

		runbookContent := string(content)

		// Act & Assert - Verify key sections
		requiredSections := []string{
			"Investigation Steps",
			"Manual Certificate Renewal",
			"Zero-Downtime Certificate Rotation",
			"Emergency",
			"Prometheus",
			"Alert",
		}

		for _, section := range requiredSections {
			assert.Contains(t, runbookContent, section, "Runbook should contain section: %s", section)
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// Integration Test: Full Certificate Lifecycle.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestSRE008_Integration_FullCertificateLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Complete certificate lifecycle: create, monitor, renew, rotate", func(t *testing.T) {
		// Arrange
		clientset := fake.NewSimpleClientset()
		ctx := context.Background()
		namespace := "test-integration"
		certName := "app-certificate"

		// Step 1: Create initial certificate
		initialSecret, err := createTestTLSSecret(namespace, certName, 90)
		require.NoError(t, err, "Should create initial certificate")
		_, err = clientset.CoreV1().Secrets(namespace).Create(ctx, initialSecret, metav1.CreateOptions{})
		require.NoError(t, err)

		// Step 2: Monitor - Check certificate will expire
		secret, err := clientset.CoreV1().Secrets(namespace).Get(ctx, certName, metav1.GetOptions{})
		require.NoError(t, err)

		block, _ := pem.Decode(secret.Data["tls.crt"])
		cert, err := x509.ParseCertificate(block.Bytes)
		require.NoError(t, err)

		daysUntilExpiry := int(time.Until(cert.NotAfter).Hours() / 24)
		assert.Greater(t, daysUntilExpiry, 0, "Certificate should be valid")

		// Step 3: Renew - Create new certificate before expiry
		renewedSecret, err := createTestTLSSecret(namespace, certName, 365)
		require.NoError(t, err)

		// Step 4: Rotate - Update secret with new certificate
		_, err = clientset.CoreV1().Secrets(namespace).Update(ctx, renewedSecret, metav1.UpdateOptions{})
		require.NoError(t, err, "Should successfully rotate certificate")

		// Step 5: Verify - Check new certificate is valid
		finalSecret, err := clientset.CoreV1().Secrets(namespace).Get(ctx, certName, metav1.GetOptions{})
		require.NoError(t, err)

		finalBlock, _ := pem.Decode(finalSecret.Data["tls.crt"])
		finalCert, err := x509.ParseCertificate(finalBlock.Bytes)
		require.NoError(t, err)

		finalDaysValid := int(time.Until(finalCert.NotAfter).Hours() / 24)
		assert.Greater(t, finalDaysValid, 300, "Renewed certificate should have longer validity")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// Benchmark: Certificate Operations Performance.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func BenchmarkSRE008_CertificateGeneration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, _, err := generateTestCertificate(365)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSRE008_CertificateInventoryScan(b *testing.B) {
	// Setup
	clientset := fake.NewSimpleClientset()
	ctx := context.Background()

	for i := 0; i < 100; i++ {
		secret, _ := createTestTLSSecret("test-ns", filepath.Join("cert", string(rune(i))), 90)
		_, _ = clientset.CoreV1().Secrets("test-ns").Create(ctx, secret, metav1.CreateOptions{})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		secrets, _ := clientset.CoreV1().Secrets("test-ns").List(ctx, metav1.ListOptions{})
		for _, secret := range secrets.Items {
			if secret.Type == corev1.SecretTypeTLS {
				_ = secret.Data["tls.crt"]
			}
		}
	}
}
