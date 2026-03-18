// Copyright (c) 2025 Justin Cranford

package tls_generator

import (
	"crypto/elliptic"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"net/url"
	"testing"
	"time"

	cryptoutilSharedCryptoCertificate "cryptoutil/internal/shared/crypto/certificate"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

var errInjectTLSGen = errors.New("injected test error")

func TestGenerateAutoTLSGeneratedSettings_CAKeyGenError(t *testing.T) {
	orig := generateECDSAKeyPairFn
	generateECDSAKeyPairFn = func(_ elliptic.Curve) (*cryptoutilSharedCryptoKeygen.KeyPair, error) {
		return nil, errInjectTLSGen
	}

	defer func() { generateECDSAKeyPairFn = orig }()

	_, err := GenerateAutoTLSGeneratedSettings([]string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault}, nil, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
	require.ErrorContains(t, err, "failed to generate CA key pair")
}

func TestGenerateAutoTLSGeneratedSettings_CASubjectsError(t *testing.T) {
	orig := createCASubjectsFn
	createCASubjectsFn = func(_ []*cryptoutilSharedCryptoKeygen.KeyPair, _ string, _ time.Duration) ([]*cryptoutilSharedCryptoCertificate.Subject, error) {
		return nil, errInjectTLSGen
	}

	defer func() { createCASubjectsFn = orig }()

	_, err := GenerateAutoTLSGeneratedSettings([]string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault}, nil, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
	require.ErrorContains(t, err, "failed to create CA subjects")
}

func TestGenerateAutoTLSGeneratedSettings_ServerKeyGenError(t *testing.T) {
	call := 0
	orig := generateECDSAKeyPairFn
	generateECDSAKeyPairFn = func(c elliptic.Curve) (*cryptoutilSharedCryptoKeygen.KeyPair, error) {
		call++
		if call > 2 {
			return nil, errInjectTLSGen
		}

		return orig(c)
	}

	defer func() { generateECDSAKeyPairFn = orig }()

	_, err := GenerateAutoTLSGeneratedSettings([]string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault}, nil, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
	require.ErrorContains(t, err, "failed to generate server key pair")
}

func TestGenerateAutoTLSGeneratedSettings_CreateEndEntityError(t *testing.T) {
	orig := createEndEntitySubjectFn
	createEndEntitySubjectFn = func(_ *cryptoutilSharedCryptoCertificate.Subject, _ *cryptoutilSharedCryptoKeygen.KeyPair, _ string, _ time.Duration, _ []string, _ []net.IP, _ []string, _ []*url.URL, _ x509.KeyUsage, _ []x509.ExtKeyUsage) (*cryptoutilSharedCryptoCertificate.Subject, error) {
		return nil, errInjectTLSGen
	}

	defer func() { createEndEntitySubjectFn = orig }()

	_, err := GenerateAutoTLSGeneratedSettings([]string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault}, nil, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
	require.ErrorContains(t, err, "failed to create server certificate")
}

func TestGenerateAutoTLSGeneratedSettings_BuildTLSError(t *testing.T) {
	orig := buildTLSCertificateFn
	buildTLSCertificateFn = func(_ *cryptoutilSharedCryptoCertificate.Subject) (*tls.Certificate, *x509.CertPool, *x509.CertPool, error) {
		return nil, nil, nil, errInjectTLSGen
	}

	defer func() { buildTLSCertificateFn = orig }()

	_, err := GenerateAutoTLSGeneratedSettings([]string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault}, nil, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
	require.ErrorContains(t, err, "failed to build TLS certificate")
}

func TestGenerateAutoTLSGeneratedSettings_MarshalKeyError(t *testing.T) {
	orig := marshalPKCS8PrivateKeyFn
	marshalPKCS8PrivateKeyFn = func(_ any) ([]byte, error) { return nil, errInjectTLSGen }

	defer func() { marshalPKCS8PrivateKeyFn = orig }()

	_, err := GenerateAutoTLSGeneratedSettings([]string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault}, nil, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
	require.ErrorContains(t, err, "failed to marshal server private key")
}

func TestGenerateTestCA_KeyGenError(t *testing.T) {
	orig := generateECDSAKeyPairFn
	generateECDSAKeyPairFn = func(_ elliptic.Curve) (*cryptoutilSharedCryptoKeygen.KeyPair, error) {
		return nil, errInjectTLSGen
	}

	defer func() { generateECDSAKeyPairFn = orig }()

	_, _, err := GenerateTestCA()
	require.ErrorContains(t, err, "failed to generate CA key pair")
}

func TestGenerateTestCA_CASubjectsError(t *testing.T) {
	orig := createCASubjectsFn
	createCASubjectsFn = func(_ []*cryptoutilSharedCryptoKeygen.KeyPair, _ string, _ time.Duration) ([]*cryptoutilSharedCryptoCertificate.Subject, error) {
		return nil, errInjectTLSGen
	}

	defer func() { createCASubjectsFn = orig }()

	_, _, err := GenerateTestCA()
	require.ErrorContains(t, err, "failed to create CA subjects")
}

func TestGenerateTestCA_MarshalKeyError(t *testing.T) {
	orig := marshalPKCS8PrivateKeyFn
	marshalPKCS8PrivateKeyFn = func(_ any) ([]byte, error) { return nil, errInjectTLSGen }

	defer func() { marshalPKCS8PrivateKeyFn = orig }()

	_, _, err := GenerateTestCA()
	require.ErrorContains(t, err, "failed to marshal CA private key")
}

func TestGenerateServerCertFromCA_ServerKeyGenError(t *testing.T) {
	caCertPEM, caKeyPEM, err := GenerateTestCA()
	require.NoError(t, err)

	orig := generateECDSAKeyPairFn
	generateECDSAKeyPairFn = func(_ elliptic.Curve) (*cryptoutilSharedCryptoKeygen.KeyPair, error) {
		return nil, errInjectTLSGen
	}

	defer func() { generateECDSAKeyPairFn = orig }()

	_, err = GenerateServerCertFromCA(caCertPEM, caKeyPEM, []string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault}, []string{cryptoutilSharedMagic.IPv4Loopback}, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
	require.ErrorContains(t, err, "failed to generate server key pair")
}

func TestGenerateServerCertFromCA_BuildTLSError(t *testing.T) {
	caCertPEM, caKeyPEM, err := GenerateTestCA()
	require.NoError(t, err)

	orig := buildTLSCertificateFn
	buildTLSCertificateFn = func(_ *cryptoutilSharedCryptoCertificate.Subject) (*tls.Certificate, *x509.CertPool, *x509.CertPool, error) {
		return nil, nil, nil, errInjectTLSGen
	}

	defer func() { buildTLSCertificateFn = orig }()

	_, err = GenerateServerCertFromCA(caCertPEM, caKeyPEM, []string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault}, []string{cryptoutilSharedMagic.IPv4Loopback}, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
	require.ErrorContains(t, err, "failed to build TLS certificate")
}

func TestGenerateServerCertFromCA_MarshalKeyError(t *testing.T) {
	caCertPEM, caKeyPEM, err := GenerateTestCA()
	require.NoError(t, err)

	orig := marshalPKCS8PrivateKeyFn
	marshalPKCS8PrivateKeyFn = func(_ any) ([]byte, error) { return nil, errInjectTLSGen }

	defer func() { marshalPKCS8PrivateKeyFn = orig }()

	_, err = GenerateServerCertFromCA(caCertPEM, caKeyPEM, []string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault}, []string{cryptoutilSharedMagic.IPv4Loopback}, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year)
	require.ErrorContains(t, err, "failed to marshal server private key")
}
