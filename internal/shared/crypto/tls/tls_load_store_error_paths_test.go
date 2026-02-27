// Copyright (c) 2025 Justin Cranford
//
//

package tls

import (
	"crypto/tls"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"encoding/pem"
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadCertificatePKCS12(t *testing.T) {
	t.Parallel()

	_, err := LoadCertificatePKCS12("cert.p12", "password")

	require.Error(t, err)
	require.Contains(t, err.Error(), "not yet implemented")
}

func TestLoadCertificatePEM_ReadCertError(t *testing.T) {
	t.Parallel()

	_, err := LoadCertificatePEM("/nonexistent/path/cert.pem", "")

	require.Error(t, err)
}

func TestLoadCertificatePEM_EmptyCertList(t *testing.T) {
	t.Parallel()

	// Write a PEM file with a non-CERTIFICATE block.
	tmpFile, err := os.CreateTemp(t.TempDir(), "test*.pem")
	require.NoError(t, err)

	_ = pem.Encode(tmpFile, &pem.Block{Type: cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey, Bytes: []byte("fakekey")})
	require.NoError(t, tmpFile.Close())

	_, err = LoadCertificatePEM(tmpFile.Name(), "")

	require.Error(t, err)
	require.Contains(t, err.Error(), "no certificates found")
}

func TestLoadCertificatePEM_ParseCertError(t *testing.T) {
	t.Parallel()

	// Write a CERTIFICATE PEM block with invalid DER bytes.
	tmpFile, err := os.CreateTemp(t.TempDir(), "test*.pem")
	require.NoError(t, err)

	_ = pem.Encode(tmpFile, &pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeCertificate, Bytes: []byte("notvalidder")})
	require.NoError(t, tmpFile.Close())

	_, err = LoadCertificatePEM(tmpFile.Name(), "")

	require.Error(t, err)
}

func TestLoadCertificatePEM_ReadKeyError(t *testing.T) {
	t.Parallel()

	// Create a real certificate PEM file.
	chain, err := CreateCAChain(DefaultCAChainOptions("test.local"))
	require.NoError(t, err)

	subject, err := chain.CreateEndEntity(ServerEndEntityOptions("server.test.local", []string{"server.test.local", cryptoutilSharedMagic.DefaultOTLPHostnameDefault}, []net.IP{net.ParseIP(cryptoutilSharedMagic.IPv4Loopback)}))
	require.NoError(t, err)

	opts := DefaultStorageOptions(t.TempDir())
	stored, err := StoreCertificate(subject, opts)
	require.NoError(t, err)

	// Now try to load with a nonexistent key path.
	_, err = LoadCertificatePEM(stored.CertificatePath, "/nonexistent/key.pem")

	require.Error(t, err)
}

func TestLoadCertificatePEM_DecodeKeyError(t *testing.T) {
	t.Parallel()

	// Create a real certificate PEM file.
	chain, err := CreateCAChain(DefaultCAChainOptions("test.local"))
	require.NoError(t, err)

	subject, err := chain.CreateEndEntity(ServerEndEntityOptions("server.test.local", []string{"server.test.local", cryptoutilSharedMagic.DefaultOTLPHostnameDefault}, []net.IP{net.ParseIP(cryptoutilSharedMagic.IPv4Loopback)}))
	require.NoError(t, err)

	opts := DefaultStorageOptions(t.TempDir())
	stored, err := StoreCertificate(subject, opts)
	require.NoError(t, err)

	// Write an invalid (non-PEM) key file.
	keyFile := stored.CertificatePath + ".key"
	require.NoError(t, os.WriteFile(keyFile, []byte("not valid PEM content"), cryptoutilSharedMagic.CacheFilePermissions)) //nolint:gosec // Test file.

	_, err = LoadCertificatePEM(stored.CertificatePath, keyFile)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode private key PEM")
}

func TestLoadCertificatePEM_ParseKeyError(t *testing.T) {
	t.Parallel()

	// Create a real certificate PEM file.
	chain, err := CreateCAChain(DefaultCAChainOptions("test.local"))
	require.NoError(t, err)

	subject, err := chain.CreateEndEntity(ServerEndEntityOptions("server.test.local", []string{"server.test.local", cryptoutilSharedMagic.DefaultOTLPHostnameDefault}, []net.IP{net.ParseIP(cryptoutilSharedMagic.IPv4Loopback)}))
	require.NoError(t, err)

	opts := DefaultStorageOptions(t.TempDir())
	stored, err := StoreCertificate(subject, opts)
	require.NoError(t, err)

	// Write a key PEM file with valid PEM block but invalid PKCS8 bytes.
	keyFile := stored.CertificatePath + ".key"
	keyPEMBytes := pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey, Bytes: []byte("notvalidpkcs8")})
	require.NoError(t, os.WriteFile(keyFile, keyPEMBytes, cryptoutilSharedMagic.CacheFilePermissions)) //nolint:gosec // Test file.

	_, err = LoadCertificatePEM(stored.CertificatePath, keyFile)

	require.Error(t, err)
}

func TestNewClientForTest(t *testing.T) {
	t.Parallel()

	client := NewClientForTest()

	require.NotNil(t, client)
}

func TestCreateEndEntity_NilIssuingCA(t *testing.T) {
	t.Parallel()

	chain := &CAChain{IssuingCA: nil}

	_, err := chain.CreateEndEntity(&EndEntityOptions{SubjectName: "test.local"})

	require.Error(t, err)
	require.Contains(t, err.Error(), "no issuing CA available")
}

func TestNewServerConfig_ClientAuthWithNilClientCAs(t *testing.T) {
	t.Parallel()

	subject := testSubjectHelper(t)

	config, err := NewServerConfig(&ServerConfigOptions{
		Subject:    subject,
		ClientAuth: tls.RequireAndVerifyClientCert,
		ClientCAs:  nil, // Should default to rootCAsPool.
	})

	require.NoError(t, err)
	require.NotNil(t, config)
	require.NotNil(t, config.TLSConfig.ClientCAs, "ClientCAs should be set from rootCAsPool")
}

func TestNewServerConfig_WithCipherSuites(t *testing.T) {
	t.Parallel()

	subject := testSubjectHelper(t)

	config, err := NewServerConfig(&ServerConfigOptions{
		Subject:      subject,
		CipherSuites: []uint16{tls.TLS_AES_128_GCM_SHA256, tls.TLS_AES_256_GCM_SHA384},
	})

	require.NoError(t, err)
	require.NotNil(t, config)
	require.Equal(t, []uint16{tls.TLS_AES_128_GCM_SHA256, tls.TLS_AES_256_GCM_SHA384}, config.TLSConfig.CipherSuites)
}

func TestStoreCertificate_UnsupportedFormat(t *testing.T) {
	t.Parallel()

	subject := testSubjectHelper(t)
	opts := DefaultStorageOptions(t.TempDir())
	opts.Format = StorageFormat("unsupported")

	_, err := StoreCertificate(subject, opts)

	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported storage format")
}

func TestLoadCertificatePEM_CACertificate(t *testing.T) {
	t.Parallel()

	chain, err := CreateCAChain(DefaultCAChainOptions("test-ca.local"))
	require.NoError(t, err)

	// Store the root CA certificate.
	opts := DefaultStorageOptions(t.TempDir())
	stored, err := StoreCertificate(chain.RootCA, opts)
	require.NoError(t, err)

	// Load it back - should hit the IsCA branch and set MaxPathLen.
	subject, err := LoadCertificatePEM(stored.CertificatePath, stored.PrivateKeyPath)
	require.NoError(t, err)
	require.NotNil(t, subject)
	require.True(t, subject.KeyMaterial.CertificateChain[0].IsCA)
}
