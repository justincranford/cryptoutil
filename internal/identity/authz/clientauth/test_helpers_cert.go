// Copyright (c) 2025 Justin Cranford

package clientauth

import (
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const testCertValidityHours = 24 // Test certificates valid for 24 hours

// Helper functions for certificate generation in tests.

func createTestCAForAuth(t *testing.T) (*x509.Certificate, *ecdsa.PrivateKey) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err, "CA key generation should succeed")

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "Test CA",
		},
		NotBefore:             time.Now().Add(-testCertValidityHours * time.Hour),
		NotAfter:              time.Now().Add(testCertValidityHours * time.Hour),
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

func createTestClientCertForAuth(t *testing.T, caCert *x509.Certificate, caKey *ecdsa.PrivateKey) *x509.Certificate {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err, "Client key generation should succeed")

	template := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			CommonName: "Test Client",
		},
		NotBefore: time.Now().Add(-1 * time.Hour),
		NotAfter:  time.Now().Add(testCertValidityHours * time.Hour),
		KeyUsage:  x509.KeyUsageDigitalSignature,
	}

	certDER, err := x509.CreateCertificate(crand.Reader, template, caCert, &key.PublicKey, caKey)
	require.NoError(t, err, "Client certificate creation should succeed")

	cert, err := x509.ParseCertificate(certDER)
	require.NoError(t, err, "Client certificate parsing should succeed")

	return cert
}
