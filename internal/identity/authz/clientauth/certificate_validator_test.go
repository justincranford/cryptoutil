// Copyright (c) 2025 Justin Cranford
//
//

package clientauth_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cryptoutil/internal/identity/authz/clientauth"
)

// TestCACertificateValidator_ValidCertificate validates successful certificate validation.
func TestCACertificateValidator_ValidCertificate(t *testing.T) {
	t.Parallel()

	// Create CA and client certificates.
	caCert, caKey := createTestCA(t)
	clientCert := createTestClientCert(t, caCert, caKey)

	// Create validator with CA trust.
	caPool := x509.NewCertPool()
	caPool.AddCert(caCert)

	validator := clientauth.NewCACertificateValidator(caPool, nil)
	validator.SetValidationOptions(false, false) // Disable subject/fingerprint validation.

	// Validate client certificate.
	err := validator.ValidateCertificate(clientCert, [][]byte{clientCert.Raw})
	require.NoError(t, err, "Valid certificate should pass validation")
}

// TestCACertificateValidator_ExpiredCertificate validates expired certificate rejection.
func TestCACertificateValidator_ExpiredCertificate(t *testing.T) {
	t.Parallel()

	// Create CA and expired client certificate.
	caCert, caKey := createTestCA(t)
	clientCert := createExpiredClientCert(t, caCert, caKey)

	// Create validator with CA trust.
	caPool := x509.NewCertPool()
	caPool.AddCert(caCert)

	validator := clientauth.NewCACertificateValidator(caPool, nil)

	// Validate expired client certificate.
	err := validator.ValidateCertificate(clientCert, [][]byte{clientCert.Raw})
	require.Error(t, err, "Expired certificate should fail validation")
	require.Contains(t, err.Error(), "expired", "Error should indicate expiration")
}

// TestCACertificateValidator_UntrustedCA validates untrusted CA rejection.
func TestCACertificateValidator_UntrustedCA(t *testing.T) {
	t.Parallel()

	// Create two separate CAs.
	trustedCA, _ := createTestCA(t)
	untrustedCA, untrustedKey := createTestCA(t)

	// Create client cert signed by untrusted CA.
	clientCert := createTestClientCert(t, untrustedCA, untrustedKey)

	// Create validator with only trusted CA.
	caPool := x509.NewCertPool()
	caPool.AddCert(trustedCA)

	validator := clientauth.NewCACertificateValidator(caPool, nil)

	// Validate client certificate signed by untrusted CA.
	err := validator.ValidateCertificate(clientCert, [][]byte{clientCert.Raw})
	require.Error(t, err, "Certificate from untrusted CA should fail validation")
}

// TestCACertificateValidator_NilCertificate validates nil certificate handling.
func TestCACertificateValidator_NilCertificate(t *testing.T) {
	t.Parallel()

	caPool := x509.NewCertPool()
	validator := clientauth.NewCACertificateValidator(caPool, nil)

	err := validator.ValidateCertificate(nil, nil)
	require.Error(t, err, "Nil certificate should fail validation")
}

// createTestCA creates a test Certificate Authority.
func createTestCA(t *testing.T) (*x509.Certificate, *ecdsa.PrivateKey) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err, "CA key generation should succeed")

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "Test CA",
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err, "CA certificate creation should succeed")

	cert, err := x509.ParseCertificate(certDER)
	require.NoError(t, err, "CA certificate parsing should succeed")

	return cert, key
}

// createTestClientCert creates a test client certificate.
func createTestClientCert(t *testing.T, caCert *x509.Certificate, caKey *ecdsa.PrivateKey) *x509.Certificate {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err, "Client key generation should succeed")

	template := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			CommonName: "Test Client",
		},
		NotBefore: time.Now().Add(-1 * time.Hour),
		NotAfter:  time.Now().Add(24 * time.Hour),
		KeyUsage:  x509.KeyUsageDigitalSignature,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, caCert, &key.PublicKey, caKey)
	require.NoError(t, err, "Client certificate creation should succeed")

	cert, err := x509.ParseCertificate(certDER)
	require.NoError(t, err, "Client certificate parsing should succeed")

	return cert
}

// createExpiredClientCert creates an expired test client certificate.
func createExpiredClientCert(t *testing.T, caCert *x509.Certificate, caKey *ecdsa.PrivateKey) *x509.Certificate {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err, "Client key generation should succeed")

	template := &x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject: pkix.Name{
			CommonName: "Expired Client",
		},
		NotBefore:   time.Now().Add(-48 * time.Hour),
		NotAfter:    time.Now().Add(-24 * time.Hour), // Expired.
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, caCert, &key.PublicKey, caKey)
	require.NoError(t, err, "Expired certificate creation should succeed")

	cert, err := x509.ParseCertificate(certDER)
	require.NoError(t, err, "Expired certificate parsing should succeed")

	return cert
}
