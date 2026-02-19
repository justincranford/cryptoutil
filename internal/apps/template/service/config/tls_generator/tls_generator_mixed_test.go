// Copyright (c) 2025 Justin Cranford

package tls_generator

import (
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"net"
	"testing"
	"time"

	cryptoutilSharedCryptoCertificate "cryptoutil/internal/shared/crypto/certificate"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestGenerateTLSMaterialMixed_MissingCACertPEM(t *testing.T) {
	t.Parallel()

	// Expect GenerateServerCertFromCA to fail when CA cert PEM is missing.
	_, err := GenerateServerCertFromCA(nil, []byte("dummy-key"), []string{"localhost"}, []string{"127.0.0.1"}, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
	require.Error(t, err)
	require.Contains(t, err.Error(), "CA certificate PEM")
}

// TestGenerateTLSMaterialMixed_MissingCAKeyPEM verifies error when MixedCAKeyPEM is missing.
func TestGenerateTLSMaterialMixed_MissingCAKeyPEM(t *testing.T) {
	t.Parallel()

	// Generate a valid CA cert without providing the private key to test missing CA key handling.
	caKeyPair, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(elliptic.P384())
	require.NoError(t, err)

	caSubjects, err := cryptoutilSharedCryptoCertificate.CreateCASubjects([]*cryptoutilSharedCryptoKeygen.KeyPair{caKeyPair}, "Test CA", time.Duration(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)*24*time.Hour)
	require.NoError(t, err)

	caCert := caSubjects[0].KeyMaterial.CertificateChain[0]
	caCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCert.Raw})

	_, err = GenerateServerCertFromCA(caCertPEM, nil, []string{"localhost"}, []string{"127.0.0.1"}, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
	require.Error(t, err)
	require.Contains(t, err.Error(), "CA private key PEM")
}

// TestGenerateTLSMaterialMixed_InvalidIPAddress verifies error when IP address is invalid.
func TestGenerateTLSMaterialMixed_InvalidIPAddress(t *testing.T) {
	t.Parallel()

	// Generate valid CA for test setup.
	duration := time.Duration(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year) * 24 * time.Hour

	caKeyPair, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(elliptic.P384())
	require.NoError(t, err)

	caSubjects, err := cryptoutilSharedCryptoCertificate.CreateCASubjects([]*cryptoutilSharedCryptoKeygen.KeyPair{caKeyPair}, "Test CA", duration)
	require.NoError(t, err)

	caCert := caSubjects[0].KeyMaterial.CertificateChain[0]

	caCertPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCert.Raw,
	})

	caKeyBytes, err := x509.MarshalPKCS8PrivateKey(caKeyPair.Private)
	require.NoError(t, err)

	caKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: caKeyBytes,
	})

	// Test with invalid IP address - GenerateServerCertFromCA should detect invalid IP.
	_, err = GenerateServerCertFromCA(caCertPEM, caKeyPEM, []string{"localhost"}, []string{"invalid-ip"}, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid IP address")
}

func TestGenerateTLSMaterialMixed_ECPrivateKey(t *testing.T) {
	t.Parallel()

	// Generate EC CA for testing EC PRIVATE KEY format.
	duration := time.Duration(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year) * 24 * time.Hour

	caKeyPair, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(elliptic.P384())
	require.NoError(t, err)

	caSubjects, err := cryptoutilSharedCryptoCertificate.CreateCASubjects([]*cryptoutilSharedCryptoKeygen.KeyPair{caKeyPair}, "EC Test CA", duration)
	require.NoError(t, err)

	caCert := caSubjects[0].KeyMaterial.CertificateChain[0]

	caCertPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCert.Raw,
	})

	// Marshal EC key as SEC1 (EC PRIVATE KEY format).
	caKeyBytes, marshalErr := x509.MarshalECPrivateKey(caKeyPair.Private.(*ecdsa.PrivateKey)) //nolint:errcheck // Error checked via require.NoError on next line.
	require.NoError(t, marshalErr)

	caKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: caKeyBytes,
	})

	// Generate server cert signed by EC CA and verify TLS material can be built.
	mixedCfg, err := GenerateServerCertFromCA(caCertPEM, caKeyPEM, []string{"localhost", "ec-test"}, []string{"127.0.0.1", "::1"}, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
	require.NoError(t, err)
	material, err := GenerateTLSMaterial(mixedCfg)
	require.NoError(t, err)
	require.NotNil(t, material)
	require.NotNil(t, material.Config)
	require.Len(t, material.Config.Certificates, 1)
}

// TestGenerateTLSMaterialAuto_HappyPath tests auto mode with full CA hierarchy generation.
func TestGenerateTLSMaterialAuto_HappyPath(t *testing.T) {
	t.Parallel()

	// Generate auto-mode TLSGeneratedSettings using helper and verify resulting TLS material.
	cfg, err := GenerateAutoTLSGeneratedSettings([]string{"localhost", "auto-test"}, []string{"127.0.0.1", "::1"}, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
	require.NoError(t, err)
	material, err := GenerateTLSMaterial(cfg)
	require.NoError(t, err)
	require.NotNil(t, material)
	require.NotNil(t, material.Config)
	require.NotNil(t, material.RootCAPool)
	require.NotNil(t, material.IntermediateCAPool)

	// Verify TLS 1.3.
	require.Equal(t, uint16(tls.VersionTLS13), material.Config.MinVersion)

	// Verify certificate was generated.
	require.Len(t, material.Config.Certificates, 1)

	tlsCert := material.Config.Certificates[0]
	require.NotNil(t, tlsCert.Leaf)
	require.Equal(t, "Auto-Generated Server Certificate", tlsCert.Leaf.Subject.CommonName)

	// Verify DNS names.
	require.Contains(t, tlsCert.Leaf.DNSNames, "localhost")
	require.Contains(t, tlsCert.Leaf.DNSNames, "auto-test")

	// Verify IP addresses.
	require.Len(t, tlsCert.Leaf.IPAddresses, 2)

	// Parse expected IPs.
	expectedIP1 := parseIP(t, "127.0.0.1")
	expectedIP2 := parseIP(t, "::1")

	// Check if IPs match (handling IPv4-mapped IPv6).
	foundIP1 := false
	foundIP2 := false

	for _, ip := range tlsCert.Leaf.IPAddresses {
		if ip.Equal(expectedIP1) {
			foundIP1 = true
		}

		if ip.Equal(expectedIP2) {
			foundIP2 = true
		}
	}

	require.True(t, foundIP1, "expected IP 127.0.0.1 not found in certificate")
	require.True(t, foundIP2, "expected IP ::1 not found in certificate")

	// Verify 2-tier CA hierarchy for auto mode (server + intermediate CA = 2 certs).
	// Root CA is not included in the chain as per TLS best practice (client should have root CA).
	require.Len(t, tlsCert.Certificate, 2)

	// Verify issuing CA is intermediate (not root).
	serverCert := tlsCert.Leaf
	require.NotEqual(t, serverCert.Issuer.String(), serverCert.Subject.String())

	// Verify TLS config is usable.
	require.Len(t, material.Config.Certificates, 1)
	require.Equal(t, uint16(tls.VersionTLS13), material.Config.MinVersion)
}

// TestGenerateTLSMaterialAuto_DefaultValidity verifies that default validity is applied when not specified.
func TestGenerateTLSMaterialAuto_DefaultValidity(t *testing.T) {
	t.Parallel()

	cfg, err := GenerateAutoTLSGeneratedSettings([]string{"localhost"}, []string{"127.0.0.1"}, 0) // validityDays=0 -> default
	require.NoError(t, err)
	material, err := GenerateTLSMaterial(cfg)
	require.NoError(t, err)
	require.NotNil(t, material)
	require.NotNil(t, material.Config)

	// Verify certificate was generated with default validity.
	require.Len(t, material.Config.Certificates, 1)

	tlsCert := material.Config.Certificates[0]
	require.NotNil(t, tlsCert.Leaf)

	// Verify validity period is approximately 365 days (allow 1 day tolerance for time.Now() drift).
	validityDuration := tlsCert.Leaf.NotAfter.Sub(tlsCert.Leaf.NotBefore)
	expectedDuration := time.Duration(365) * 24 * time.Hour
	tolerance := 24 * time.Hour

	require.InDelta(t, expectedDuration.Seconds(), validityDuration.Seconds(), tolerance.Seconds())
}

// TestGenerateTLSMaterialAuto_EmptyDNSNames verifies that auto mode works with empty DNS names.
func TestGenerateTLSMaterialAuto_EmptyDNSNames(t *testing.T) {
	t.Parallel()

	cfg, err := GenerateAutoTLSGeneratedSettings(nil, []string{"127.0.0.1"}, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
	require.NoError(t, err)
	material, err := GenerateTLSMaterial(cfg)
	require.NoError(t, err)
	require.NotNil(t, material)

	// Verify certificate was generated with no DNS names.
	require.Len(t, material.Config.Certificates, 1)

	tlsCert := material.Config.Certificates[0]
	require.NotNil(t, tlsCert.Leaf)
	require.Empty(t, tlsCert.Leaf.DNSNames)
}

// TestGenerateTLSMaterialAuto_InvalidIPAddress verifies error when IP address is invalid.
func TestGenerateTLSMaterialAuto_InvalidIPAddress(t *testing.T) {
	t.Parallel()

	_, err := GenerateAutoTLSGeneratedSettings([]string{"localhost"}, []string{"not-an-ip"}, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid IP address")
}

// Helper: parseIP parses IP address string (used for test assertions).
func parseIP(t *testing.T, ipStr string) (ip net.IP) {
	t.Helper()

	ip = net.ParseIP(ipStr)
	require.NotNil(t, ip, "failed to parse IP address: %s", ipStr)

	return ip
}

// TestGenerateTestCA_HappyPath verifies GenerateTestCA generates valid CA certificate and key.
func TestGenerateTestCA_HappyPath(t *testing.T) {
	t.Parallel()

	caCertPEM, caKeyPEM, err := GenerateTestCA()
	require.NoError(t, err)
	require.NotEmpty(t, caCertPEM)
	require.NotEmpty(t, caKeyPEM)

	// Verify CA certificate is valid PEM.
	certBlock, _ := pem.Decode(caCertPEM)
	require.NotNil(t, certBlock, "CA cert should be valid PEM")
	require.Equal(t, "CERTIFICATE", certBlock.Type)

	// Parse and verify certificate properties.
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	require.NoError(t, err)
	require.True(t, cert.IsCA, "certificate should be a CA")
	require.True(t, cert.BasicConstraintsValid)

	// Verify CA key is valid PEM.
	keyBlock, _ := pem.Decode(caKeyPEM)
	require.NotNil(t, keyBlock, "CA key should be valid PEM")
	require.Equal(t, "PRIVATE KEY", keyBlock.Type)

	// Parse and verify key is ECDSA P-384.
	key, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	require.NoError(t, err)

	ecKey, ok := key.(*ecdsa.PrivateKey)
	require.True(t, ok, "key should be ECDSA")
	require.Equal(t, elliptic.P384().Params().Name, ecKey.Curve.Params().Name)
}

// TestGenerateTestCA_UsableForSigning verifies CA can sign server certificates.
func TestGenerateTestCA_UsableForSigning(t *testing.T) {
	t.Parallel()

	// Generate test CA.
	caCertPEM, caKeyPEM, err := GenerateTestCA()
	require.NoError(t, err)

	// Use CA to generate a server certificate.
	tlsSettings, err := GenerateServerCertFromCA(
		caCertPEM,
		caKeyPEM,
		[]string{"localhost"},
		[]string{"127.0.0.1"},
		cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year,
	)
	require.NoError(t, err)
	require.NotEmpty(t, tlsSettings.StaticCertPEM)
	require.NotEmpty(t, tlsSettings.StaticKeyPEM)

	// Verify server certificate is valid.
	certBlock, _ := pem.Decode(tlsSettings.StaticCertPEM)
	require.NotNil(t, certBlock)

	serverCert, err := x509.ParseCertificate(certBlock.Bytes)
	require.NoError(t, err)
	require.False(t, serverCert.IsCA, "server cert should not be CA")
	require.Contains(t, serverCert.DNSNames, "localhost")
}

// TestGenerateTestCA_UniquePerCall verifies each call generates unique CA.
func TestGenerateTestCA_UniquePerCall(t *testing.T) {
	t.Parallel()

	// Generate two CAs.
	caCert1, caKey1, err := GenerateTestCA()
	require.NoError(t, err)

	caCert2, caKey2, err := GenerateTestCA()
	require.NoError(t, err)

	// Verify they are different.
	require.NotEqual(t, caCert1, caCert2, "CA certificates should be unique")
	require.NotEqual(t, caKey1, caKey2, "CA keys should be unique")
}

// TestGenerateTLSMaterialStatic_ChainWithInvalidCert verifies error when certificate chain contains invalid cert.
