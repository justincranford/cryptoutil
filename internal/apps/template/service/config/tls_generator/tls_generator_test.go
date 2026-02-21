// Copyright (c) 2025 Justin Cranford

package tls_generator

import (
	"crypto/elliptic"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
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

	// Generate server key pair.
	serverKeyPair, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(elliptic.P384())
	require.NoError(t, err)

	// Use Intermediate CA as issuing CA (its PrivateKey is valid after CreateCASubjects).
	issuingCA := caSubjects[0]

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

	// Verify certificate chain: server cert + intermediate CA cert + root CA cert = 3 total.
	require.Len(t, tlsCert.Certificate, 3)

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
