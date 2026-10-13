// Copyright (c) 2025-2026 Justin Cranford.
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

const (
	testSuccessfulGeneration     = "successful generation"
	testCASubjectCreationError   = "CA subject creation error"
	testServerKeyGenerationError = "server key generation error"
	testTLSCertificateBuildError = "TLS certificate build error"
	testPrivateKeyEncodingError  = "private key encoding error"
)

var errInjectTLSGen = errors.New("injected test error")

func TestGenerateAutoTLSGeneratedSettings_Operations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		configure func(*tlsGenerationOperations)
		wantErr   string
	}{
		{name: testSuccessfulGeneration},
		{
			name: "CA key generation error",
			configure: func(operations *tlsGenerationOperations) {
				operations.generateECDSAKeyPair = func(elliptic.Curve) (*cryptoutilSharedCryptoKeygen.KeyPair, error) {
					return nil, errInjectTLSGen
				}
			},
			wantErr: errorFailedToGenerateCAKeyPair,
		},
		{
			name: testCASubjectCreationError,
			configure: func(operations *tlsGenerationOperations) {
				operations.createCASubjects = func([]*cryptoutilSharedCryptoKeygen.KeyPair, string, time.Duration) ([]*cryptoutilSharedCryptoCertificate.Subject, error) {
					return nil, errInjectTLSGen
				}
			},
			wantErr: errorFailedToCreateCASubjects,
		},
		{
			name: testServerKeyGenerationError,
			configure: func(operations *tlsGenerationOperations) {
				generateKeyPair := operations.generateECDSAKeyPair

				var calls int

				operations.generateECDSAKeyPair = func(curve elliptic.Curve) (*cryptoutilSharedCryptoKeygen.KeyPair, error) {
					calls++
					if calls == cryptoutilSharedMagic.DefaultTLSAutoCAChainTiers {
						return nil, errInjectTLSGen
					}

					return generateKeyPair(curve)
				}
			},
			wantErr: errorFailedToGenerateServerKeyPair,
		},
		{
			name: "server subject creation error",
			configure: func(operations *tlsGenerationOperations) {
				operations.createEndEntity = func(*cryptoutilSharedCryptoCertificate.Subject, *cryptoutilSharedCryptoKeygen.KeyPair, string, time.Duration, []string, []net.IP, []string, []*url.URL, x509.KeyUsage, []x509.ExtKeyUsage) (*cryptoutilSharedCryptoCertificate.Subject, error) {
					return nil, errInjectTLSGen
				}
			},
			wantErr: "failed to create server certificate",
		},
		{
			name: testTLSCertificateBuildError,
			configure: func(operations *tlsGenerationOperations) {
				operations.buildTLSCertificate = func(*cryptoutilSharedCryptoCertificate.Subject) (*tls.Certificate, *x509.CertPool, *x509.CertPool, error) {
					return nil, nil, nil, errInjectTLSGen
				}
			},
			wantErr: errorFailedToBuildTLSCertificate,
		},
		{
			name: testPrivateKeyEncodingError,
			configure: func(operations *tlsGenerationOperations) {
				operations.pemEncodeKey = func(any) ([]byte, error) { return nil, errInjectTLSGen }
			},
			wantErr: errorFailedToPEMEncodeServerPrivateKey,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			operations := defaultTLSGenerationOperations()
			if tc.configure != nil {
				tc.configure(&operations)
			}

			generated, err := generateAutoTLSGeneratedSettingsWithOperations(
				[]string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault},
				nil,
				cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year,
				operations,
			)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.NotEmpty(t, generated.StaticCertPEM)
			require.NotEmpty(t, generated.StaticKeyPEM)
		})
	}
}

func TestGenerateTestCA_Operations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		configure func(*tlsGenerationOperations)
		wantErr   string
	}{
		{name: testSuccessfulGeneration},
		{
			name: "key generation error",
			configure: func(operations *tlsGenerationOperations) {
				operations.generateECDSAKeyPair = func(elliptic.Curve) (*cryptoutilSharedCryptoKeygen.KeyPair, error) {
					return nil, errInjectTLSGen
				}
			},
			wantErr: errorFailedToGenerateCAKeyPair,
		},
		{
			name: testCASubjectCreationError,
			configure: func(operations *tlsGenerationOperations) {
				operations.createCASubjects = func([]*cryptoutilSharedCryptoKeygen.KeyPair, string, time.Duration) ([]*cryptoutilSharedCryptoCertificate.Subject, error) {
					return nil, errInjectTLSGen
				}
			},
			wantErr: errorFailedToCreateCASubjects,
		},
		{
			name: testPrivateKeyEncodingError,
			configure: func(operations *tlsGenerationOperations) {
				operations.pemEncodeKey = func(any) ([]byte, error) { return nil, errInjectTLSGen }
			},
			wantErr: "failed to PEM-encode CA private key",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			operations := defaultTLSGenerationOperations()
			if tc.configure != nil {
				tc.configure(&operations)
			}

			certPEM, keyPEM, err := generateTestCAWithOperations(operations)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.NotEmpty(t, certPEM)
			require.NotEmpty(t, keyPEM)
		})
	}
}

func TestGenerateServerCertFromCA_Operations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		configure func(*tlsGenerationOperations)
		wantErr   string
	}{
		{name: testSuccessfulGeneration},
		{
			name: testServerKeyGenerationError,
			configure: func(operations *tlsGenerationOperations) {
				operations.generateECDSAKeyPair = func(elliptic.Curve) (*cryptoutilSharedCryptoKeygen.KeyPair, error) {
					return nil, errInjectTLSGen
				}
			},
			wantErr: errorFailedToGenerateServerKeyPair,
		},
		{
			name: testTLSCertificateBuildError,
			configure: func(operations *tlsGenerationOperations) {
				operations.buildTLSCertificate = func(*cryptoutilSharedCryptoCertificate.Subject) (*tls.Certificate, *x509.CertPool, *x509.CertPool, error) {
					return nil, nil, nil, errInjectTLSGen
				}
			},
			wantErr: errorFailedToBuildTLSCertificate,
		},
		{
			name: testPrivateKeyEncodingError,
			configure: func(operations *tlsGenerationOperations) {
				operations.pemEncodeKey = func(any) ([]byte, error) { return nil, errInjectTLSGen }
			},
			wantErr: errorFailedToPEMEncodeServerPrivateKey,
		},
	}

	caCertPEM, caKeyPEM, err := GenerateTestCA()
	require.NoError(t, err)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			operations := defaultTLSGenerationOperations()
			if tc.configure != nil {
				tc.configure(&operations)
			}

			generated, err := generateServerCertFromCAWithOperations(
				caCertPEM,
				caKeyPEM,
				[]string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault},
				[]string{cryptoutilSharedMagic.IPv4Loopback},
				cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year,
				operations,
			)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				require.Nil(t, generated)

				return
			}

			require.NoError(t, err)
			require.NotEmpty(t, generated.StaticCertPEM)
			require.NotEmpty(t, generated.StaticKeyPEM)
		})
	}
}
