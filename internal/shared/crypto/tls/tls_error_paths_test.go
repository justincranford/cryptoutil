// Copyright (c) 2025 Justin Cranford
//
//

package tls

import (
	"crypto/elliptic"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	cryptoutilSharedCryptoCertificate "cryptoutil/internal/shared/crypto/certificate"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"

	"github.com/stretchr/testify/require"
)

// testSubjectHelper creates a real certificate Subject for use in storage and config tests.
func testSubjectHelper(t *testing.T) *cryptoutilSharedCryptoCertificate.Subject {
	t.Helper()

	chain, err := CreateCAChain(DefaultCAChainOptions("test.local"))
	require.NoError(t, err)

	subject, err := chain.CreateEndEntity(ServerEndEntityOptions("server.test.local", []string{"server.test.local", "localhost"}, []net.IP{net.ParseIP("127.0.0.1")}))
	require.NoError(t, err)

	return subject
}

func TestValidateFQDN_TooLong(t *testing.T) {
	t.Parallel()

	// Build a name longer than 253 characters.
	name := strings.Repeat("a", 254)

	err := ValidateFQDN(name)

	require.Error(t, err)
	require.Contains(t, err.Error(), "too long")
}

func TestValidateFQDN_LabelTooLong(t *testing.T) {
	t.Parallel()

	// Build a label longer than 63 characters.
	label := strings.Repeat("a", 64)
	name := label + ".example.com"

	err := ValidateFQDN(name)

	require.Error(t, err)
	require.Contains(t, err.Error(), "label too long")
}

func TestCreateCAChain_KeyGenError(t *testing.T) {
	injectedErr := errors.New("injected keygen error")
	orig := chainGenerateECDSAKeyPairFn

	chainGenerateECDSAKeyPairFn = func(_ elliptic.Curve) (*cryptoutilSharedCryptoKeygen.KeyPair, error) {
		return nil, injectedErr
	}

	defer func() { chainGenerateECDSAKeyPairFn = orig }()

	_, err := CreateCAChain(DefaultCAChainOptions("test.local"))

	require.ErrorIs(t, err, injectedErr)
}

func TestCreateCAChain_CreateCASubjectsError(t *testing.T) {
	injectedErr := errors.New("injected create CA subjects error")
	orig := chainCreateCASubjectsFn

	chainCreateCASubjectsFn = func(_ []*cryptoutilSharedCryptoKeygen.KeyPair, _ string, _ time.Duration) ([]*cryptoutilSharedCryptoCertificate.Subject, error) {
		return nil, injectedErr
	}

	defer func() { chainCreateCASubjectsFn = orig }()

	_, err := CreateCAChain(DefaultCAChainOptions("test.local"))

	require.ErrorIs(t, err, injectedErr)
}

func TestCreateEndEntity_EmptySubjectName(t *testing.T) {
	t.Parallel()

	chain, err := CreateCAChain(DefaultCAChainOptions("test.local"))
	require.NoError(t, err)

	_, err = chain.CreateEndEntity(&EndEntityOptions{SubjectName: ""})

	require.Error(t, err)
	require.Contains(t, err.Error(), "subject name cannot be empty")
}

func TestCreateEndEntity_KeyGenError(t *testing.T) {
	// Create chain before injecting failure.
	chain, err := CreateCAChain(DefaultCAChainOptions("test.local"))
	require.NoError(t, err)

	injectedErr := errors.New("injected keygen error for end entity")
	orig := chainGenerateECDSAKeyPairFn

	chainGenerateECDSAKeyPairFn = func(_ elliptic.Curve) (*cryptoutilSharedCryptoKeygen.KeyPair, error) {
		return nil, injectedErr
	}

	defer func() { chainGenerateECDSAKeyPairFn = orig }()

	_, err = chain.CreateEndEntity(ServerEndEntityOptions("server.test.local", []string{"server.test.local", "localhost"}, []net.IP{net.ParseIP("127.0.0.1")}))

	require.ErrorIs(t, err, injectedErr)
}

func TestCreateEndEntity_CreateSubjectError(t *testing.T) {
	// Create chain before injecting failure.
	chain, err := CreateCAChain(DefaultCAChainOptions("test.local"))
	require.NoError(t, err)

	injectedErr := errors.New("injected create end entity error")
	orig := chainCreateEndEntitySubjectFn

	chainCreateEndEntitySubjectFn = func(
		_ *cryptoutilSharedCryptoCertificate.Subject,
		_ *cryptoutilSharedCryptoKeygen.KeyPair,
		_ string,
		_ time.Duration,
		_ []string,
		_ []net.IP,
		_ []string,
		_ []*url.URL,
		_ x509.KeyUsage,
		_ []x509.ExtKeyUsage,
	) (*cryptoutilSharedCryptoCertificate.Subject, error) {
		return nil, injectedErr
	}

	defer func() { chainCreateEndEntitySubjectFn = orig }()

	_, err = chain.CreateEndEntity(ServerEndEntityOptions("server.test.local", []string{"server.test.local", "localhost"}, []net.IP{net.ParseIP("127.0.0.1")}))

	require.ErrorIs(t, err, injectedErr)
}

func TestNewServerConfig_NilOpts(t *testing.T) {
	t.Parallel()

	_, err := NewServerConfig(nil)

	require.Error(t, err)
	require.Contains(t, err.Error(), "options cannot be nil")
}

func TestNewServerConfig_NilSubject(t *testing.T) {
	t.Parallel()

	_, err := NewServerConfig(&ServerConfigOptions{Subject: nil})

	require.Error(t, err)
	require.Contains(t, err.Error(), "subject cannot be nil")
}

func TestNewServerConfig_BuildTLSCertError(t *testing.T) {
	injectedErr := errors.New("injected build TLS cert error")
	orig := configBuildTLSCertificateFn

	configBuildTLSCertificateFn = func(_ *cryptoutilSharedCryptoCertificate.Subject) (*tls.Certificate, *x509.CertPool, *x509.CertPool, error) {
		return nil, nil, nil, injectedErr
	}

	defer func() { configBuildTLSCertificateFn = orig }()

	_, err := NewServerConfig(&ServerConfigOptions{Subject: testSubjectHelper(t)})

	require.ErrorIs(t, err, injectedErr)
}

func TestNewClientConfig_NilOpts(t *testing.T) {
	t.Parallel()

	_, err := NewClientConfig(nil)

	require.Error(t, err)
	require.Contains(t, err.Error(), "options cannot be nil")
}

func TestNewClientConfig_BuildTLSCertError(t *testing.T) {
	// Create subject before injecting failure.
	subject := testSubjectHelper(t)

	injectedErr := errors.New("injected build client TLS cert error")
	orig := configBuildTLSCertificateFn

	configBuildTLSCertificateFn = func(_ *cryptoutilSharedCryptoCertificate.Subject) (*tls.Certificate, *x509.CertPool, *x509.CertPool, error) {
		return nil, nil, nil, injectedErr
	}

	defer func() { configBuildTLSCertificateFn = orig }()

	_, err := NewClientConfig(&ClientConfigOptions{ClientSubject: subject})

	require.ErrorIs(t, err, injectedErr)
}

func TestStoreCertificate_MkdirAllError(t *testing.T) {
	subject := testSubjectHelper(t)

	injectedErr := errors.New("injected MkdirAll error")
	orig := storageMkdirAllFn

	storageMkdirAllFn = func(_ string, _ os.FileMode) error { return injectedErr }

	defer func() { storageMkdirAllFn = orig }()

	opts := DefaultStorageOptions(t.TempDir())

	_, err := StoreCertificate(subject, opts)

	require.ErrorIs(t, err, injectedErr)
}

func TestStoreCertificate_FormatPKCS12(t *testing.T) {
	t.Parallel()

	subject := testSubjectHelper(t)
	opts := DefaultStorageOptions(t.TempDir())
	opts.Format = FormatPKCS12

	_, err := StoreCertificate(subject, opts)

	require.Error(t, err)
	require.Contains(t, err.Error(), "not yet implemented")
}

func TestStorePEM_WriteFileCertError(t *testing.T) {
	subject := testSubjectHelper(t)

	injectedErr := errors.New("injected write cert error")
	orig := storageWriteFileFn

	storageWriteFileFn = func(_ string, _ []byte, _ os.FileMode) error { return injectedErr }

	defer func() { storageWriteFileFn = orig }()

	opts := DefaultStorageOptions(t.TempDir())

	_, err := StoreCertificate(subject, opts)

	require.ErrorIs(t, err, injectedErr)
}

func TestStorePEM_MarshalPKCS8Error(t *testing.T) {
	subject := testSubjectHelper(t)

	injectedErr := errors.New("injected marshal PKCS8 error")
	orig := storageMarshalPKCS8Fn

	storageMarshalPKCS8Fn = func(_ any) ([]byte, error) { return nil, injectedErr }

	defer func() { storageMarshalPKCS8Fn = orig }()

	opts := DefaultStorageOptions(t.TempDir())
	opts.IncludePrivateKey = true

	_, err := StoreCertificate(subject, opts)

	require.ErrorIs(t, err, injectedErr)
}

func TestStorePEM_WriteFileKeyError(t *testing.T) {
	subject := testSubjectHelper(t)

	injectedErr := errors.New("injected write key error")
	orig := storageWriteFileFn

	certWritten := false

	storageWriteFileFn = func(path string, data []byte, mode os.FileMode) error {
		if !certWritten {
			certWritten = true

			// Let the cert write succeed.
			return os.WriteFile(path, data, mode)
		}

		return injectedErr
	}

	defer func() { storageWriteFileFn = orig }()

	opts := DefaultStorageOptions(t.TempDir())
	opts.IncludePrivateKey = true

	_, err := StoreCertificate(subject, opts)

	require.ErrorIs(t, err, injectedErr)
}

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

	_ = pem.Encode(tmpFile, &pem.Block{Type: "PRIVATE KEY", Bytes: []byte("fakekey")})
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

	_ = pem.Encode(tmpFile, &pem.Block{Type: "CERTIFICATE", Bytes: []byte("notvalidder")})
	require.NoError(t, tmpFile.Close())

	_, err = LoadCertificatePEM(tmpFile.Name(), "")

	require.Error(t, err)
}

func TestLoadCertificatePEM_ReadKeyError(t *testing.T) {
	t.Parallel()

	// Create a real certificate PEM file.
	chain, err := CreateCAChain(DefaultCAChainOptions("test.local"))
	require.NoError(t, err)

	subject, err := chain.CreateEndEntity(ServerEndEntityOptions("server.test.local", []string{"server.test.local", "localhost"}, []net.IP{net.ParseIP("127.0.0.1")}))
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

	subject, err := chain.CreateEndEntity(ServerEndEntityOptions("server.test.local", []string{"server.test.local", "localhost"}, []net.IP{net.ParseIP("127.0.0.1")}))
	require.NoError(t, err)

	opts := DefaultStorageOptions(t.TempDir())
	stored, err := StoreCertificate(subject, opts)
	require.NoError(t, err)

	// Write an invalid (non-PEM) key file.
	keyFile := stored.CertificatePath + ".key"
	require.NoError(t, os.WriteFile(keyFile, []byte("not valid PEM content"), 0o600)) //nolint:gosec // Test file.

	_, err = LoadCertificatePEM(stored.CertificatePath, keyFile)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode private key PEM")
}

func TestLoadCertificatePEM_ParseKeyError(t *testing.T) {
	t.Parallel()

	// Create a real certificate PEM file.
	chain, err := CreateCAChain(DefaultCAChainOptions("test.local"))
	require.NoError(t, err)

	subject, err := chain.CreateEndEntity(ServerEndEntityOptions("server.test.local", []string{"server.test.local", "localhost"}, []net.IP{net.ParseIP("127.0.0.1")}))
	require.NoError(t, err)

	opts := DefaultStorageOptions(t.TempDir())
	stored, err := StoreCertificate(subject, opts)
	require.NoError(t, err)

	// Write a key PEM file with valid PEM block but invalid PKCS8 bytes.
	keyFile := stored.CertificatePath + ".key"
	keyPEMBytes := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte("notvalidpkcs8")})
	require.NoError(t, os.WriteFile(keyFile, keyPEMBytes, 0o600)) //nolint:gosec // Test file.

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
