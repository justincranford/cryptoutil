// Copyright (c) 2025 Justin Cranford

package tls_generator

import (
	crand "crypto/rand"
	rsa "crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestGenerateTLSMaterialStatic_ChainWithInvalidCert(t *testing.T) {
	t.Parallel()

	// Use shared test fixtures from TestMain to get a valid cert/key pair.
	// Then append an invalid certificate block to the PEM chain.
	invalidCertPEM := append([]byte{}, testServerCertPEM...)
	invalidCertPEM = append(invalidCertPEM, []byte("-----BEGIN CERTIFICATE-----\nTGhpcyBpcyBub3QgYSB2YWxpZCBjZXJ0aWZpY2F0ZQ==\n-----END CERTIFICATE-----\n")...)

	cfg := &TLSGeneratedSettings{
		StaticCertPEM: invalidCertPEM,
		StaticKeyPEM:  testServerKeyPEM,
	}

	material, err := GenerateTLSMaterial(cfg)
	require.Error(t, err)
	require.Nil(t, material)
	require.Contains(t, err.Error(), "failed to parse certificate")
}

// TestGenerateServerCertFromCA_RSAPrivateKeyFormat verifies RSA PRIVATE KEY format can be parsed.
// Note: Full RSA CA signing is not supported by the underlying certificate library (CreateEndEntitySubject uses ECDSA).
// This test verifies the key parsing path works correctly.
func TestGenerateServerCertFromCA_RSAPrivateKeyFormat(t *testing.T) {
	t.Parallel()

	// Generate RSA key pair for CA.
	rsaKey, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	rsaPrivKey, ok := rsaKey.Private.(*rsa.PrivateKey)
	require.True(t, ok, "Expected RSA private key")

	// Create a self-signed RSA CA certificate manually.
	notBefore := time.Now().UTC()
	notAfter := notBefore.Add(time.Duration(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year) * cryptoutilSharedMagic.HoursPerDay * time.Hour)

	// Generate a random serial number (CA/Browser Forum compliant).
	serialNumber, err := crand.Int(crand.Reader, new(big.Int).Lsh(big.NewInt(1), cryptoutilSharedMagic.TLSSelfSignedCertSerialNumberBits))
	require.NoError(t, err)

	caTemplate := &x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{CommonName: "RSA Test CA"},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            1,
	}

	// Self-sign the CA certificate.
	caCertDER, err := x509.CreateCertificate(crand.Reader, caTemplate, caTemplate, &rsaPrivKey.PublicKey, rsaPrivKey)
	require.NoError(t, err)

	caCertPEM := pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeCertificate, Bytes: caCertDER})

	// Marshal RSA key as PKCS#1 (RSA PRIVATE KEY format) - this is the code path we're testing.
	caKeyBytes := x509.MarshalPKCS1PrivateKey(rsaPrivKey)
	caKeyPEM := pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeRSAPrivateKey, Bytes: caKeyBytes})

	// Call GenerateServerCertFromCA - will fail at server certificate creation stage
	// (ECDSA signature algorithm mismatch), but RSA PRIVATE KEY parsing should succeed.
	_, err = GenerateServerCertFromCA(caCertPEM, caKeyPEM, []string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault}, []string{cryptoutilSharedMagic.IPv4Loopback}, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)

	// The error should be about signature algorithm mismatch, NOT about RSA key parsing.
	// This confirms the RSA PRIVATE KEY format was successfully parsed.
	require.Error(t, err)
	require.Contains(t, err.Error(), "SignatureAlgorithm")
	require.NotContains(t, err.Error(), "failed to parse CA private key")
}

// TestGenerateServerCertFromCA_InvalidCACertPEM verifies error when CA cert PEM is invalid.
func TestGenerateServerCertFromCA_InvalidCACertPEM(t *testing.T) {
	t.Parallel()

	// Use invalid/non-CERTIFICATE PEM block.
	invalidCACertPEM := []byte("-----BEGIN INVALID-----\nYmFkZGF0YQ==\n-----END INVALID-----\n")
	validCAKeyPEM := testServerKeyPEM // Just need something non-empty.

	_, err := GenerateServerCertFromCA(invalidCACertPEM, validCAKeyPEM, []string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault}, []string{cryptoutilSharedMagic.IPv4Loopback}, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode CA certificate PEM")
}

// TestGenerateServerCertFromCA_MalformedCACert verifies error when CA cert data is malformed.
func TestGenerateServerCertFromCA_MalformedCACert(t *testing.T) {
	t.Parallel()

	// Use CERTIFICATE block with invalid data.
	malformedCACertPEM := []byte("-----BEGIN CERTIFICATE-----\nbm90YXZhbGlkY2VydA==\n-----END CERTIFICATE-----\n")
	validCAKeyPEM := testServerKeyPEM

	_, err := GenerateServerCertFromCA(malformedCACertPEM, validCAKeyPEM, []string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault}, []string{cryptoutilSharedMagic.IPv4Loopback}, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse CA certificate")
}

// TestGenerateServerCertFromCA_InvalidCAKeyPEM verifies error when CA key PEM cannot be decoded.
func TestGenerateServerCertFromCA_InvalidCAKeyPEM(t *testing.T) {
	t.Parallel()

	// Generate a valid CA cert.
	caCertPEM, _, err := GenerateTestCA()
	require.NoError(t, err)

	// Use non-PEM data for key.
	invalidCAKeyPEM := []byte("this is not valid PEM data")

	_, err = GenerateServerCertFromCA(caCertPEM, invalidCAKeyPEM, []string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault}, []string{cryptoutilSharedMagic.IPv4Loopback}, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode CA private key PEM")
}

// TestGenerateServerCertFromCA_UnsupportedKeyType verifies error when CA key type is unsupported.
func TestGenerateServerCertFromCA_UnsupportedKeyType(t *testing.T) {
	t.Parallel()

	// Generate a valid CA cert.
	caCertPEM, _, err := GenerateTestCA()
	require.NoError(t, err)

	// Use unsupported key type (e.g., "PUBLIC KEY" instead of private key).
	unsupportedKeyPEM := []byte("-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEZe3P\n-----END PUBLIC KEY-----\n")

	_, err = GenerateServerCertFromCA(caCertPEM, unsupportedKeyPEM, []string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault}, []string{cryptoutilSharedMagic.IPv4Loopback}, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported CA private key type")
}

// TestGenerateServerCertFromCA_MalformedPrivateKey verifies error when private key data is malformed.
func TestGenerateServerCertFromCA_MalformedPrivateKey(t *testing.T) {
	t.Parallel()

	// Generate a valid CA cert.
	caCertPEM, _, err := GenerateTestCA()
	require.NoError(t, err)

	// Use PRIVATE KEY block with invalid data.
	malformedKeyPEM := []byte("-----BEGIN PRIVATE KEY-----\nbm90YXZhbGlka2V5\n-----END PRIVATE KEY-----\n")

	_, err = GenerateServerCertFromCA(caCertPEM, malformedKeyPEM, []string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault}, []string{cryptoutilSharedMagic.IPv4Loopback}, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse CA private key")
}

// TestGenerateServerCertFromCA_DefaultValidity verifies default validity when validityDays is 0.
func TestGenerateServerCertFromCA_DefaultValidity(t *testing.T) {
	t.Parallel()

	// Generate a valid CA.
	caCertPEM, caKeyPEM, err := GenerateTestCA()
	require.NoError(t, err)

	// Generate server cert with validityDays=0 (should default to 365).
	mixedCfg, err := GenerateServerCertFromCA(caCertPEM, caKeyPEM, []string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault}, []string{cryptoutilSharedMagic.IPv4Loopback}, 0)
	require.NoError(t, err)
	require.NotNil(t, mixedCfg)

	// Parse the server certificate and verify validity period.
	certBlock, _ := pem.Decode(mixedCfg.StaticCertPEM)
	require.NotNil(t, certBlock)

	serverCert, err := x509.ParseCertificate(certBlock.Bytes)
	require.NoError(t, err)

	// Verify validity period is approximately 365 days.
	validityDuration := serverCert.NotAfter.Sub(serverCert.NotBefore)
	expectedDuration := time.Duration(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year) * cryptoutilSharedMagic.HoursPerDay * time.Hour
	tolerance := cryptoutilSharedMagic.HoursPerDay * time.Hour

	require.InDelta(t, expectedDuration.Seconds(), validityDuration.Seconds(), tolerance.Seconds())
}

// TestGenerateAutoTLSGeneratedSettings_EmptyDNSAndIPs verifies auto mode works with empty DNS and IPs.
func TestGenerateAutoTLSGeneratedSettings_EmptyDNSAndIPs(t *testing.T) {
	t.Parallel()

	cfg, err := GenerateAutoTLSGeneratedSettings(nil, nil, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
	require.NoError(t, err)
	material, err := GenerateTLSMaterial(cfg)
	require.NoError(t, err)
	require.NotNil(t, material)

	// Verify certificate was generated with no DNS names or IPs.
	require.Len(t, material.Config.Certificates, 1)

	tlsCert := material.Config.Certificates[0]
	require.NotNil(t, tlsCert.Leaf)
	require.Empty(t, tlsCert.Leaf.DNSNames)
	require.Empty(t, tlsCert.Leaf.IPAddresses)
}
