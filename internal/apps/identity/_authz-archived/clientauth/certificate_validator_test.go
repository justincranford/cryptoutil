// Copyright (c) 2025 Justin Cranford
//
//

package clientauth_test

import (
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityClientAuth "cryptoutil/internal/apps/identity/authz/clientauth"
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

	validator := cryptoutilIdentityClientAuth.NewCACertificateValidator(caPool, nil)
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

	validator := cryptoutilIdentityClientAuth.NewCACertificateValidator(caPool, nil)

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

	validator := cryptoutilIdentityClientAuth.NewCACertificateValidator(caPool, nil)

	// Validate client certificate signed by untrusted CA.
	err := validator.ValidateCertificate(clientCert, [][]byte{clientCert.Raw})
	require.Error(t, err, "Certificate from untrusted CA should fail validation")
}

// TestCACertificateValidator_NilCertificate validates nil certificate handling.
func TestCACertificateValidator_NilCertificate(t *testing.T) {
	t.Parallel()

	caPool := x509.NewCertPool()
	validator := cryptoutilIdentityClientAuth.NewCACertificateValidator(caPool, nil)

	err := validator.ValidateCertificate(nil, nil)
	require.Error(t, err, "Nil certificate should fail validation")
}

// createTestCA creates a test Certificate Authority.
func createTestCA(t *testing.T) (*x509.Certificate, *ecdsa.PrivateKey) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err, "CA key generation should succeed")

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "Test CA",
		},
		NotBefore:             time.Now().UTC().Add(-1 * time.Hour),
		NotAfter:              time.Now().UTC().Add(cryptoutilSharedMagic.HoursPerDay * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certDER, err := x509.CreateCertificate(crand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err, "CA certificate creation should succeed")

	cert, err := x509.ParseCertificate(certDER)
	require.NoError(t, err, "CA certificate parsing should succeed")

	return cert, key
}

// createTestClientCert creates a test client certificate.
func createTestClientCert(t *testing.T, caCert *x509.Certificate, caKey *ecdsa.PrivateKey) *x509.Certificate {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err, "Client key generation should succeed")

	template := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			CommonName: "Test Client",
		},
		NotBefore: time.Now().UTC().Add(-1 * time.Hour),
		NotAfter:  time.Now().UTC().Add(cryptoutilSharedMagic.HoursPerDay * time.Hour),
		KeyUsage:  x509.KeyUsageDigitalSignature,
	}

	certDER, err := x509.CreateCertificate(crand.Reader, template, caCert, &key.PublicKey, caKey)
	require.NoError(t, err, "Client certificate creation should succeed")

	cert, err := x509.ParseCertificate(certDER)
	require.NoError(t, err, "Client certificate parsing should succeed")

	return cert
}

// createExpiredClientCert creates an expired test client certificate.
func createExpiredClientCert(t *testing.T, caCert *x509.Certificate, caKey *ecdsa.PrivateKey) *x509.Certificate {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err, "Client key generation should succeed")

	template := &x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject: pkix.Name{
			CommonName: "Expired Client",
		},
		NotBefore:   time.Now().UTC().Add(-cryptoutilSharedMagic.HMACSHA384KeySize * time.Hour),
		NotAfter:    time.Now().UTC().Add(-cryptoutilSharedMagic.HoursPerDay * time.Hour), // Expired.
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	certDER, err := x509.CreateCertificate(crand.Reader, template, caCert, &key.PublicKey, caKey)
	require.NoError(t, err, "Expired certificate creation should succeed")

	cert, err := x509.ParseCertificate(certDER)
	require.NoError(t, err, "Expired certificate parsing should succeed")

	return cert
}

// TestCACertificateValidator_IsRevoked_Deprecated tests the deprecated IsRevoked method.
func TestCACertificateValidator_IsRevoked_Deprecated(t *testing.T) {
	t.Parallel()

	validator := cryptoutilIdentityClientAuth.NewCACertificateValidator(nil, nil)

	// Deprecated method always returns false.
	isRevoked := validator.IsRevoked(big.NewInt(cryptoutilSharedMagic.AnswerToLifeUniverseEverything))
	require.False(t, isRevoked, "deprecated IsRevoked should always return false")
}

// TestSelfSignedCertificateValidator_ValidateCertificate tests self-signed certificate validation.
func TestSelfSignedCertificateValidator_ValidateCertificate(t *testing.T) {
	t.Parallel()

	// Create test self-signed certificate.
	privKey, err := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	certTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "Test Self-Signed Client",
			Organization: []string{"Test Org"},
		},
		NotBefore:             time.Now().UTC().Add(-1 * time.Hour),
		NotAfter:              time.Now().UTC().Add(cryptoutilSharedMagic.HoursPerDay * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(crand.Reader, certTemplate, certTemplate, &privKey.PublicKey, privKey)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(certDER)
	require.NoError(t, err)

	// Create expired certificate.
	expiredTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			CommonName:   "Expired Client",
			Organization: []string{"Test Org"},
		},
		NotBefore:             time.Now().UTC().Add(-cryptoutilSharedMagic.HMACSHA384KeySize * time.Hour),
		NotAfter:              time.Now().UTC().Add(-cryptoutilSharedMagic.HoursPerDay * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	expiredCertDER, err := x509.CreateCertificate(crand.Reader, expiredTemplate, expiredTemplate, &privKey.PublicKey, privKey)
	require.NoError(t, err)

	expiredCert, err := x509.ParseCertificate(expiredCertDER)
	require.NoError(t, err)

	// Create not-yet-valid certificate.
	futureTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject: pkix.Name{
			CommonName:   "Future Client",
			Organization: []string{"Test Org"},
		},
		NotBefore:             time.Now().UTC().Add(cryptoutilSharedMagic.HoursPerDay * time.Hour),
		NotAfter:              time.Now().UTC().Add(cryptoutilSharedMagic.HMACSHA384KeySize * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	futureCertDER, err := x509.CreateCertificate(crand.Reader, futureTemplate, futureTemplate, &privKey.PublicKey, privKey)
	require.NoError(t, err)

	futureCert, err := x509.ParseCertificate(futureCertDER)
	require.NoError(t, err)

	tests := []struct {
		name        string
		pinnedCerts map[string]*x509.Certificate
		clientCert  *x509.Certificate
		wantErr     bool
		errContains string
	}{
		{
			name: "valid pinned certificate",
			pinnedCerts: map[string]*x509.Certificate{
				"test-client": cert,
			},
			clientCert: cert,
			wantErr:    false,
		},
		{
			name:        "nil certificate",
			pinnedCerts: map[string]*x509.Certificate{},
			clientCert:  nil,
			wantErr:     true,
			errContains: "Invalid client authentication",
		},
		{
			name: "expired certificate",
			pinnedCerts: map[string]*x509.Certificate{
				"expired-client": expiredCert,
			},
			clientCert:  expiredCert,
			wantErr:     true,
			errContains: "certificate has expired",
		},
		{
			name: "not yet valid certificate",
			pinnedCerts: map[string]*x509.Certificate{
				"future-client": futureCert,
			},
			clientCert:  futureCert,
			wantErr:     true,
			errContains: "certificate is not yet valid",
		},
		{
			name:        "certificate not pinned",
			pinnedCerts: map[string]*x509.Certificate{},
			clientCert:  cert,
			wantErr:     true,
			errContains: "certificate not found in pinned certificates",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			validator := cryptoutilIdentityClientAuth.NewSelfSignedCertificateValidator(tc.pinnedCerts)
			err := validator.ValidateCertificate(tc.clientCert, nil)

			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestSelfSignedCertificateValidator_IsRevoked tests IsRevoked for self-signed certs.
func TestSelfSignedCertificateValidator_IsRevoked(t *testing.T) {
	t.Parallel()

	validator := cryptoutilIdentityClientAuth.NewSelfSignedCertificateValidator(nil)

	// IsRevoked always returns false for self-signed certificates.
	isRevoked := validator.IsRevoked(big.NewInt(cryptoutilSharedMagic.AnswerToLifeUniverseEverything))
	require.False(t, isRevoked, "self-signed validator IsRevoked should always return false")
}

// TestCertificateParser_ParsePEMCertificate tests PEM certificate parsing.
func TestCertificateParser_ParsePEMCertificate(t *testing.T) {
	t.Parallel()

	// Create test certificate.
	privKey, err := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "Test Certificate",
		},
		NotBefore: time.Now().UTC().Add(-1 * time.Hour),
		NotAfter:  time.Now().UTC().Add(cryptoutilSharedMagic.HoursPerDay * time.Hour),
	}

	certDER, err := x509.CreateCertificate(crand.Reader, template, template, &privKey.PublicKey, privKey)
	require.NoError(t, err)

	// Encode as PEM.
	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  cryptoutilSharedMagic.StringPEMTypeCertificate,
		Bytes: certDER,
	})

	tests := []struct {
		name        string
		pemData     []byte
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid PEM certificate",
			pemData: pemData,
			wantErr: false,
		},
		{
			name:        "invalid PEM data",
			pemData:     []byte("not a PEM certificate"),
			wantErr:     true,
			errContains: "failed to decode PEM certificate",
		},
		{
			name: "invalid certificate data",
			pemData: pem.EncodeToMemory(&pem.Block{
				Type:  cryptoutilSharedMagic.StringPEMTypeCertificate,
				Bytes: []byte("invalid certificate bytes"),
			}),
			wantErr:     true,
			errContains: "failed to parse certificate",
		},
	}

	parser := &cryptoutilIdentityClientAuth.CertificateParser{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cert, err := parser.ParsePEMCertificate(tc.pemData)

			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)
				require.Nil(t, cert)
			} else {
				require.NoError(t, err)
				require.NotNil(t, cert)
				require.Equal(t, template.Subject.CommonName, cert.Subject.CommonName)
			}
		})
	}
}
