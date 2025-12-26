// Copyright (c) 2025 Justin Cranford

package tls_generator

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"net"
	"os"
	"testing"
	"time"

	cryptoutilConfig "cryptoutil/internal/shared/config"
	cryptoutilCertificate "cryptoutil/internal/shared/crypto/certificate"
	cryptoutilKeyGen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// Package-level test fixtures (generated once in TestMain).
var (
	testCAKeyPairs      []*cryptoutilKeyGen.KeyPair
	testCASubjects      []*cryptoutilCertificate.Subject
	testIssuingCAKey    any
	testServerKeyPair   *cryptoutilKeyGen.KeyPair
	testServerSubject   *cryptoutilCertificate.Subject
	testServerCertPEM   []byte
	testServerKeyPEM    []byte
	testECKeyPair       *cryptoutilKeyGen.KeyPair
	testECServerSubject *cryptoutilCertificate.Subject
	testECServerCertPEM []byte
	testECServerKeyPEM  []byte
)

// TestMain runs once before all tests to generate shared test fixtures.
// This significantly improves test performance by generating certificates once instead of per-test.
func TestMain(m *testing.M) {
	var err error

	// Generate test duration (1 year validity).
	duration := time.Duration(cryptoutilMagic.TLSTestEndEntityCertValidity1Year) * 24 * time.Hour

	// Generate 2-tier CA hierarchy (Root + Intermediate).
	testCAKeyPairs = make([]*cryptoutilKeyGen.KeyPair, 2)

	for i := range testCAKeyPairs {
		testCAKeyPairs[i], err = cryptoutilKeyGen.GenerateECDSAKeyPair(elliptic.P384())
		if err != nil {
			panic("failed to generate CA key pair: " + err.Error())
		}
	}

	testCASubjects, err = cryptoutilCertificate.CreateCASubjects(testCAKeyPairs, "Test CA", duration)
	if err != nil {
		panic("failed to create CA subjects: " + err.Error())
	}

	// Save issuing CA private key (CreateCASubjects clears it).
	testIssuingCAKey = testCAKeyPairs[len(testCAKeyPairs)-1].Private

	// Generate ECDSA P-384 server key pair and certificate.
	testServerKeyPair, err = cryptoutilKeyGen.GenerateECDSAKeyPair(elliptic.P384())
	if err != nil {
		panic("failed to generate server key pair: " + err.Error())
	}

	// Restore issuing CA private key for signing.
	issuingCA := testCASubjects[len(testCASubjects)-1]
	issuingCA.KeyMaterial.PrivateKey = testIssuingCAKey

	testServerSubject, err = cryptoutilCertificate.CreateEndEntitySubject(
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
	testECKeyPair, err = cryptoutilKeyGen.GenerateECDSAKeyPair(elliptic.P256())
	if err != nil {
		panic("failed to generate EC key pair: " + err.Error())
	}

	// Restore issuing CA private key again.
	issuingCA.KeyMaterial.PrivateKey = testIssuingCAKey

	testECServerSubject, err = cryptoutilCertificate.CreateEndEntitySubject(
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

	cfg := &TLSGeneratedSettings{
		Mode: cryptoutilConfig.TLSMode("invalid-mode"),
	}

	material, err := GenerateTLSMaterial(cfg)
	require.Error(t, err)
	require.Nil(t, material)
	require.Contains(t, err.Error(), "unknown TLS mode")
	require.Contains(t, err.Error(), "invalid-mode")
}

// TestGenerateTLSMaterialStatic_HappyPath tests static mode with valid certificate chain.
func TestGenerateTLSMaterialStatic_HappyPath(t *testing.T) {
	t.Parallel()

	// Generate 3-tier CA hierarchy for testing.
	duration := time.Duration(cryptoutilMagic.TLSTestEndEntityCertValidity1Year) * 24 * time.Hour

	// Create 2 CA key pairs (Root + Intermediate).
	caKeyPairs := make([]*cryptoutilKeyGen.KeyPair, 2)

	var err error

	for i := range caKeyPairs {
		caKeyPairs[i], err = cryptoutilKeyGen.GenerateECDSAKeyPair(elliptic.P384())
		require.NoError(t, err)
	}

	caSubjects, err := cryptoutilCertificate.CreateCASubjects(caKeyPairs, "Test CA", duration)
	require.NoError(t, err)

	// Save issuing CA private key before CreateCASubjects clears it.
	issuingCAPrivateKey := caKeyPairs[len(caKeyPairs)-1].Private

	// Generate server key pair.
	serverKeyPair, err := cryptoutilKeyGen.GenerateECDSAKeyPair(elliptic.P384())
	require.NoError(t, err)

	// Restore issuing CA private key.
	issuingCA := caSubjects[len(caSubjects)-1]
	issuingCA.KeyMaterial.PrivateKey = issuingCAPrivateKey

	serverSubject, err := cryptoutilCertificate.CreateEndEntitySubject(
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
		Mode:          cryptoutilConfig.TLSModeStatic,
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
		Mode:         cryptoutilConfig.TLSModeStatic,
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
		Mode:          cryptoutilConfig.TLSModeStatic,
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
		Mode:          cryptoutilConfig.TLSModeStatic,
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
	duration := time.Duration(cryptoutilMagic.TLSTestEndEntityCertValidity1Year) * 24 * time.Hour

	caKeyPair, err := cryptoutilKeyGen.GenerateECDSAKeyPair(elliptic.P384())
	require.NoError(t, err)

	caSubjects, err := cryptoutilCertificate.CreateCASubjects([]*cryptoutilKeyGen.KeyPair{caKeyPair}, "Test CA", duration)
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

	// Test mixed mode.
	cfg := &TLSGeneratedSettings{
		Mode:             cryptoutilConfig.TLSModeMixed,
		MixedCACertPEM:   caCertPEM,
		MixedCAKeyPEM:    caKeyPEM,
		AutoDNSNames:     []string{"localhost", "mixed-test"},
		AutoIPAddresses:  []string{"127.0.0.1", "::1"},
		AutoValidityDays: cryptoutilMagic.TLSTestEndEntityCertValidity1Year,
	}

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

	cfg := &TLSGeneratedSettings{
		Mode:          cryptoutilConfig.TLSModeMixed,
		MixedCAKeyPEM: []byte("dummy-key"),
	}

	material, err := GenerateTLSMaterial(cfg)
	require.Error(t, err)
	require.Nil(t, material)
	require.Contains(t, err.Error(), "mixed mode requires MixedCACertPEM")
}

// TestGenerateTLSMaterialMixed_MissingCAKeyPEM verifies error when MixedCAKeyPEM is missing.
func TestGenerateTLSMaterialMixed_MissingCAKeyPEM(t *testing.T) {
	t.Parallel()

	cfg := &TLSGeneratedSettings{
		Mode:           cryptoutilConfig.TLSModeMixed,
		MixedCACertPEM: []byte("dummy-cert"),
	}

	material, err := GenerateTLSMaterial(cfg)
	require.Error(t, err)
	require.Nil(t, material)
	require.Contains(t, err.Error(), "mixed mode requires MixedCAKeyPEM")
}

// TestGenerateTLSMaterialMixed_InvalidIPAddress verifies error when IP address is invalid.
func TestGenerateTLSMaterialMixed_InvalidIPAddress(t *testing.T) {
	t.Parallel()

	// Generate valid CA for test setup.
	duration := time.Duration(cryptoutilMagic.TLSTestEndEntityCertValidity1Year) * 24 * time.Hour

	caKeyPair, err := cryptoutilKeyGen.GenerateECDSAKeyPair(elliptic.P384())
	require.NoError(t, err)

	caSubjects, err := cryptoutilCertificate.CreateCASubjects([]*cryptoutilKeyGen.KeyPair{caKeyPair}, "Test CA", duration)
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

	// Test with invalid IP address.
	cfg := &TLSGeneratedSettings{
		Mode:            cryptoutilConfig.TLSModeMixed,
		MixedCACertPEM:  caCertPEM,
		MixedCAKeyPEM:   caKeyPEM,
		AutoIPAddresses: []string{"invalid-ip"},
	}

	material, err := GenerateTLSMaterial(cfg)
	require.Error(t, err)
	require.Nil(t, material)
	require.Contains(t, err.Error(), "invalid IP address")
}

func TestGenerateTLSMaterialMixed_ECPrivateKey(t *testing.T) {
	t.Parallel()

	// Generate EC CA for testing EC PRIVATE KEY format.
	duration := time.Duration(cryptoutilMagic.TLSTestEndEntityCertValidity1Year) * 24 * time.Hour

	caKeyPair, err := cryptoutilKeyGen.GenerateECDSAKeyPair(elliptic.P384())
	require.NoError(t, err)

	caSubjects, err := cryptoutilCertificate.CreateCASubjects([]*cryptoutilKeyGen.KeyPair{caKeyPair}, "EC Test CA", duration)
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

	cfg := &TLSGeneratedSettings{
		Mode:             cryptoutilConfig.TLSModeMixed,
		MixedCACertPEM:   caCertPEM,
		MixedCAKeyPEM:    caKeyPEM,
		AutoDNSNames:     []string{"localhost", "ec-test"},
		AutoIPAddresses:  []string{"127.0.0.1", "::1"},
		AutoValidityDays: cryptoutilMagic.TLSTestEndEntityCertValidity1Year,
	}

	material, err := GenerateTLSMaterial(cfg)
	require.NoError(t, err)
	require.NotNil(t, material)
	require.NotNil(t, material.Config)
	require.Len(t, material.Config.Certificates, 1)
}

// TestGenerateTLSMaterialAuto_HappyPath tests auto mode with full CA hierarchy generation.
func TestGenerateTLSMaterialAuto_HappyPath(t *testing.T) {
	t.Parallel()

	cfg := &TLSGeneratedSettings{
		Mode:             cryptoutilConfig.TLSModeAuto,
		AutoDNSNames:     []string{"localhost", "auto-test"},
		AutoIPAddresses:  []string{"127.0.0.1", "::1"},
		AutoValidityDays: cryptoutilMagic.TLSTestEndEntityCertValidity1Year,
	}

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

	cfg := &TLSGeneratedSettings{
		Mode:            cryptoutilConfig.TLSModeAuto,
		AutoDNSNames:    []string{"localhost"},
		AutoIPAddresses: []string{"127.0.0.1"},
		// AutoValidityDays is 0 (not set) - should default to 365.
	}

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

	cfg := &TLSGeneratedSettings{
		Mode:             cryptoutilConfig.TLSModeAuto,
		AutoIPAddresses:  []string{"127.0.0.1"},
		AutoValidityDays: cryptoutilMagic.TLSTestEndEntityCertValidity1Year,
	}

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

	cfg := &TLSGeneratedSettings{
		Mode:            cryptoutilConfig.TLSModeAuto,
		AutoDNSNames:    []string{"localhost"},
		AutoIPAddresses: []string{"not-an-ip"},
	}

	material, err := GenerateTLSMaterial(cfg)
	require.Error(t, err)
	require.Nil(t, material)
	require.Contains(t, err.Error(), "invalid IP address")
}

// Helper: parseIP parses IP address string (used for test assertions).
func parseIP(t *testing.T, ipStr string) (ip net.IP) {
	t.Helper()

	ip = net.ParseIP(ipStr)
	require.NotNil(t, ip, "failed to parse IP address: %s", ipStr)

	return ip
}
