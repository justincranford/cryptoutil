// Copyright (c) 2025 Justin Cranford
//
//

package tls

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"crypto/elliptic"
	"crypto/tls"
	"crypto/x509"
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

	subject, err := chain.CreateEndEntity(ServerEndEntityOptions("server.test.local", []string{"server.test.local", cryptoutilSharedMagic.DefaultOTLPHostnameDefault}, []net.IP{net.ParseIP(cryptoutilSharedMagic.IPv4Loopback)}))
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
	label := strings.Repeat("a", cryptoutilSharedMagic.MinSerialNumberBits)
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

	_, err = chain.CreateEndEntity(ServerEndEntityOptions("server.test.local", []string{"server.test.local", cryptoutilSharedMagic.DefaultOTLPHostnameDefault}, []net.IP{net.ParseIP(cryptoutilSharedMagic.IPv4Loopback)}))

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

	_, err = chain.CreateEndEntity(ServerEndEntityOptions("server.test.local", []string{"server.test.local", cryptoutilSharedMagic.DefaultOTLPHostnameDefault}, []net.IP{net.ParseIP(cryptoutilSharedMagic.IPv4Loopback)}))

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
