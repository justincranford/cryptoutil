// Copyright (c) 2025 Justin Cranford

package tls_generator

import (
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"testing"
	"time"

	cryptoutilSharedCryptoCertificate "cryptoutil/internal/shared/crypto/certificate"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// Package-level test fixtures (generated once in TestMain).
var (
	testCAKeyPairs      []*cryptoutilSharedCryptoKeygen.KeyPair
	testCASubjects      []*cryptoutilSharedCryptoCertificate.Subject
	testIssuingCAKey    any
	testServerKeyPair   *cryptoutilSharedCryptoKeygen.KeyPair
	testServerSubject   *cryptoutilSharedCryptoCertificate.Subject
	testServerCertPEM   []byte
	testServerKeyPEM    []byte
	testECKeyPair       *cryptoutilSharedCryptoKeygen.KeyPair
	testECServerSubject *cryptoutilSharedCryptoCertificate.Subject
	testECServerCertPEM []byte
	testECServerKeyPEM  []byte
)

// TestMain runs once before all tests to generate shared test fixtures.
// This significantly improves test performance by generating certificates once instead of per-test.
func TestMain(m *testing.M) {
	var err error

	// Generate test duration (1 year validity).
	duration := time.Duration(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year) * 24 * time.Hour

	// Generate 2-tier CA hierarchy (Root + Intermediate).
	testCAKeyPairs = make([]*cryptoutilSharedCryptoKeygen.KeyPair, 2)

	for i := range testCAKeyPairs {
		testCAKeyPairs[i], err = cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(elliptic.P384())
		if err != nil {
			panic("failed to generate CA key pair: " + err.Error())
		}
	}

	testCASubjects, err = cryptoutilSharedCryptoCertificate.CreateCASubjects(testCAKeyPairs, "Test CA", duration)
	if err != nil {
		panic("failed to create CA subjects: " + err.Error())
	}

	// Save issuing CA private key (CreateCASubjects clears it).
	testIssuingCAKey = testCAKeyPairs[len(testCAKeyPairs)-1].Private

	// Generate ECDSA P-384 server key pair and certificate.
	testServerKeyPair, err = cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(elliptic.P384())
	if err != nil {
		panic("failed to generate server key pair: " + err.Error())
	}

	// Restore issuing CA private key for signing.
	issuingCA := testCASubjects[len(testCASubjects)-1]
	issuingCA.KeyMaterial.PrivateKey = testIssuingCAKey

	testServerSubject, err = cryptoutilSharedCryptoCertificate.CreateEndEntitySubject(
		issuingCA,
		testServerKeyPair,
		"Test Server",
		duration,
		[]string{"localhost", "test-server"},
		nil,
		nil,
		nil,
		x509.KeyUsageDigitalSignature|x509.KeyUsageKeyEncipherment,
		[]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	)
	if err != nil {
		panic("failed to create server subject: " + err.Error())
	}

	// Serialize server certificate chain to PEM.
	for _, cert := range testServerSubject.KeyMaterial.CertificateChain {
		testServerCertPEM = append(testServerCertPEM, pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		})...)
	}

	// Serialize server private key to PKCS8 PEM.
	keyBytes, err := x509.MarshalPKCS8PrivateKey(testServerSubject.KeyMaterial.PrivateKey)
	if err != nil {
		panic("failed to marshal server private key: " + err.Error())
	}

	testServerKeyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: keyBytes,
	})

	// Generate EC P-256 server key pair and certificate (for EC-specific tests).
	testECKeyPair, err = cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(elliptic.P256())
	if err != nil {
		panic("failed to generate EC key pair: " + err.Error())
	}

	// Restore issuing CA private key again.
	issuingCA.KeyMaterial.PrivateKey = testIssuingCAKey

	testECServerSubject, err = cryptoutilSharedCryptoCertificate.CreateEndEntitySubject(
		issuingCA,
		testECKeyPair,
		"Test EC Server",
		duration,
		[]string{"localhost"},
		nil,
		nil,
		nil,
		x509.KeyUsageDigitalSignature|x509.KeyUsageKeyEncipherment,
		[]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	)
	if err != nil {
		panic("failed to create EC server subject: " + err.Error())
	}

	// Serialize EC server certificate chain to PEM.
	for _, cert := range testECServerSubject.KeyMaterial.CertificateChain {
		testECServerCertPEM = append(testECServerCertPEM, pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		})...)
	}

	// Serialize EC server private key to PKCS8 PEM.
	ecKeyBytes, err := x509.MarshalPKCS8PrivateKey(testECServerSubject.KeyMaterial.PrivateKey)
	if err != nil {
		panic("failed to marshal EC server private key: " + err.Error())
	}

	testECServerKeyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: ecKeyBytes,
	})

	// Run all tests.
	os.Exit(m.Run())
}

// TestGenerateTLSMaterial_NilConfig verifies that nil config returns error.
func TestGenerateTLSMaterial_NilConfig(t *testing.T) {
	t.Parallel()

	material, err := GenerateTLSMaterial(nil)
	require.Error(t, err)
	require.Nil(t, material)
	require.Contains(t, err.Error(), "TLS config cannot be nil")
}

// TestGenerateTLSMaterial_UnknownMode verifies that unknown TLS mode returns error.
func TestGenerateTLSMaterial_UnknownMode(t *testing.T) {
	t.Parallel()

	// Empty settings should return an informative error about missing material.
	cfg := &TLSGeneratedSettings{}

	material, err := GenerateTLSMaterial(cfg)
	require.Error(t, err)
	require.Nil(t, material)
	require.Contains(t, err.Error(), "no TLS certificate material provided")
}

// TestGenerateTLSMaterialStatic_HappyPath tests static mode with valid certificate chain.
func TestGenerateTLSMaterialStatic_HappyPath(t *testing.T) {
	t.Parallel()

	// Generate 3-tier CA hierarchy for testing.
	duration := time.Duration(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year) * 24 * time.Hour

	// Create 2 CA key pairs (Root + Intermediate).
	caKeyPairs := make([]*cryptoutilSharedCryptoKeygen.KeyPair, 2)

	var err error

	for i := range caKeyPairs {
		caKeyPairs[i], err = cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(elliptic.P384())
		require.NoError(t, err)
	}

	caSubjects, err := cryptoutilSharedCryptoCertificate.CreateCASubjects(caKeyPairs, "Test CA", duration)
	require.NoError(t, err)

	// Save issuing CA private key before CreateCASubjects clears it.
	issuingCAPrivateKey := caKeyPairs[len(caKeyPairs)-1].Private

	// Generate server key pair.
	serverKeyPair, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(elliptic.P384())
	require.NoError(t, err)

	// Restore issuing CA private key.
	issuingCA := caSubjects[len(caSubjects)-1]
	issuingCA.KeyMaterial.PrivateKey = issuingCAPrivateKey

	serverSubject, err := cryptoutilSharedCryptoCertificate.CreateEndEntitySubject(
		issuingCA,
		serverKeyPair,
		"Test Server",
		duration,
		[]string{"localhost", "test-server"},
		nil,
		nil,
		nil,
		x509.KeyUsageDigitalSignature|x509.KeyUsageKeyEncipherment,
		[]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	)
	require.NoError(t, err)

	// Serialize certificate chain to PEM.
	var certPEM []byte

	for _, cert := range serverSubject.KeyMaterial.CertificateChain {
		certPEM = append(certPEM, pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		})...)
	}

	// Serialize private key to PEM (PKCS8).
	keyBytes, err := x509.MarshalPKCS8PrivateKey(serverSubject.KeyMaterial.PrivateKey)
	require.NoError(t, err)

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: keyBytes,
	})

	// Test static mode.
	cfg := &TLSGeneratedSettings{
		StaticCertPEM: certPEM,
		StaticKeyPEM:  keyPEM,
	}

	material, err := GenerateTLSMaterial(cfg)
	require.NoError(t, err)
	require.NotNil(t, material)
	require.NotNil(t, material.Config)
	require.NotNil(t, material.RootCAPool)
	require.NotNil(t, material.IntermediateCAPool)

	// Verify TLS 1.3.
	require.Equal(t, uint16(tls.VersionTLS13), material.Config.MinVersion)

	// Verify certificate was loaded.
	require.Len(t, material.Config.Certificates, 1)

	tlsCert := material.Config.Certificates[0]
	require.NotNil(t, tlsCert.Leaf)
	require.Equal(t, "Test Server", tlsCert.Leaf.Subject.CommonName)

	// Verify certificate chain includes server + intermediate CA (root CA not included as per TLS best practice).	// The chain includes server cert + intermediate CA cert = 2 total.	require.Len(t, tlsCert.Certificate, 3)

	// Verify DNS names.
	require.Contains(t, tlsCert.Leaf.DNSNames, "localhost")
	require.Contains(t, tlsCert.Leaf.DNSNames, "test-server")

	// Verify certificate pools populated.
	// (x509.CertPool doesn't expose subjects method for verification, but we validated chain parsing above)

	// Verify TLS config is usable.
	require.Len(t, material.Config.Certificates, 1)
	require.Equal(t, uint16(tls.VersionTLS13), material.Config.MinVersion)
}

// TestGenerateTLSMaterialStatic_MissingCertPEM verifies error when StaticCertPEM is missing.
func TestGenerateTLSMaterialStatic_MissingCertPEM(t *testing.T) {
	t.Parallel()

	cfg := &TLSGeneratedSettings{
		StaticKeyPEM: []byte("dummy-key"),
	}

	material, err := GenerateTLSMaterial(cfg)
	require.Error(t, err)
	require.Nil(t, material)
	require.Contains(t, err.Error(), "static mode requires StaticCertPEM")
}

// TestGenerateTLSMaterialStatic_MissingKeyPEM verifies error when StaticKeyPEM is missing.
func TestGenerateTLSMaterialStatic_MissingKeyPEM(t *testing.T) {
	t.Parallel()

	cfg := &TLSGeneratedSettings{
		StaticCertPEM: []byte("dummy-cert"),
	}

	material, err := GenerateTLSMaterial(cfg)
	require.Error(t, err)
	require.Nil(t, material)
	require.Contains(t, err.Error(), "static mode requires StaticKeyPEM")
}

// TestGenerateTLSMaterialStatic_InvalidCertPEM verifies error when certificate PEM is malformed.
func TestGenerateTLSMaterialStatic_InvalidCertPEM(t *testing.T) {
	t.Parallel()

	cfg := &TLSGeneratedSettings{
		StaticCertPEM: []byte("invalid-pem-data"),
		StaticKeyPEM:  []byte("invalid-key-data"),
	}

	material, err := GenerateTLSMaterial(cfg)
	require.Error(t, err)
	require.Nil(t, material)
	require.Contains(t, err.Error(), "failed to parse static TLS certificate")
}

// TestGenerateTLSMaterialMixed_HappyPath tests mixed mode with CA + auto server cert.
func TestGenerateTLSMaterialMixed_HappyPath(t *testing.T) {
	t.Parallel()

	// Generate CA for testing.
	duration := time.Duration(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year) * 24 * time.Hour

	caKeyPair, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(elliptic.P384())
	require.NoError(t, err)

	caSubjects, err := cryptoutilSharedCryptoCertificate.CreateCASubjects([]*cryptoutilSharedCryptoKeygen.KeyPair{caKeyPair}, "Test CA", duration)
	require.NoError(t, err)

	caCert := caSubjects[0].KeyMaterial.CertificateChain[0]

	// Serialize CA certificate to PEM.
	caCertPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCert.Raw,
	})

	// Serialize CA private key to PEM.
	caKeyBytes, err := x509.MarshalPKCS8PrivateKey(caKeyPair.Private)
	require.NoError(t, err)

	caKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: caKeyBytes,
	})

	// Test mixed mode: providing only CA material should instruct caller to pre-generate server cert.
	cfg := &TLSGeneratedSettings{
		MixedCACertPEM: caCertPEM,
		MixedCAKeyPEM:  caKeyPEM,
	}

	material, err := GenerateTLSMaterial(cfg)
	require.Error(t, err)
	require.Nil(t, material)

	// Use helper to generate server cert signed by CA and then generate TLS material.
	mixedCfg, err := GenerateServerCertFromCA(caCertPEM, caKeyPEM, []string{"localhost", "mixed-test"}, []string{"127.0.0.1", "::1"}, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
	require.NoError(t, err)
	require.NotNil(t, mixedCfg)

	material, err = GenerateTLSMaterial(mixedCfg)
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
	require.Equal(t, "Server Certificate", tlsCert.Leaf.Subject.CommonName)

	// Verify DNS names.
	require.Contains(t, tlsCert.Leaf.DNSNames, "localhost")
	require.Contains(t, tlsCert.Leaf.DNSNames, "mixed-test")

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

	// Verify signed by CA.
	require.Equal(t, caCert.Subject.String(), tlsCert.Leaf.Issuer.String())

	// Verify TLS config is usable.
	require.Len(t, material.Config.Certificates, 1)
	require.Equal(t, uint16(tls.VersionTLS13), material.Config.MinVersion)
}

// TestGenerateTLSMaterialMixed_MissingCACertPEM verifies error when MixedCACertPEM is missing.
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
	rsaKey, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	rsaPrivKey, ok := rsaKey.Private.(*rsa.PrivateKey)
	require.True(t, ok, "Expected RSA private key")

	// Create a self-signed RSA CA certificate manually.
	notBefore := time.Now().UTC()
	notAfter := notBefore.Add(time.Duration(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year) * 24 * time.Hour)

	// Generate a random serial number (CA/Browser Forum compliant).
	serialNumber, err := crand.Int(crand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
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

	caCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCertDER})

	// Marshal RSA key as PKCS#1 (RSA PRIVATE KEY format) - this is the code path we're testing.
	caKeyBytes := x509.MarshalPKCS1PrivateKey(rsaPrivKey)
	caKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: caKeyBytes})

	// Call GenerateServerCertFromCA - will fail at server certificate creation stage
	// (ECDSA signature algorithm mismatch), but RSA PRIVATE KEY parsing should succeed.
	_, err = GenerateServerCertFromCA(caCertPEM, caKeyPEM, []string{"localhost"}, []string{"127.0.0.1"}, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)

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

	_, err := GenerateServerCertFromCA(invalidCACertPEM, validCAKeyPEM, []string{"localhost"}, []string{"127.0.0.1"}, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode CA certificate PEM")
}

// TestGenerateServerCertFromCA_MalformedCACert verifies error when CA cert data is malformed.
func TestGenerateServerCertFromCA_MalformedCACert(t *testing.T) {
	t.Parallel()

	// Use CERTIFICATE block with invalid data.
	malformedCACertPEM := []byte("-----BEGIN CERTIFICATE-----\nbm90YXZhbGlkY2VydA==\n-----END CERTIFICATE-----\n")
	validCAKeyPEM := testServerKeyPEM

	_, err := GenerateServerCertFromCA(malformedCACertPEM, validCAKeyPEM, []string{"localhost"}, []string{"127.0.0.1"}, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
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

	_, err = GenerateServerCertFromCA(caCertPEM, invalidCAKeyPEM, []string{"localhost"}, []string{"127.0.0.1"}, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
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

	_, err = GenerateServerCertFromCA(caCertPEM, unsupportedKeyPEM, []string{"localhost"}, []string{"127.0.0.1"}, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
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

	_, err = GenerateServerCertFromCA(caCertPEM, malformedKeyPEM, []string{"localhost"}, []string{"127.0.0.1"}, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
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
	mixedCfg, err := GenerateServerCertFromCA(caCertPEM, caKeyPEM, []string{"localhost"}, []string{"127.0.0.1"}, 0)
	require.NoError(t, err)
	require.NotNil(t, mixedCfg)

	// Parse the server certificate and verify validity period.
	certBlock, _ := pem.Decode(mixedCfg.StaticCertPEM)
	require.NotNil(t, certBlock)

	serverCert, err := x509.ParseCertificate(certBlock.Bytes)
	require.NoError(t, err)

	// Verify validity period is approximately 365 days.
	validityDuration := serverCert.NotAfter.Sub(serverCert.NotBefore)
	expectedDuration := time.Duration(365) * 24 * time.Hour
	tolerance := 24 * time.Hour

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
