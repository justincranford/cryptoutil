package tls

import (
	"context"
	"crypto/x509"
	"errors"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilSharedCryptoAsn1 "cryptoutil/internal/shared/crypto/asn1"
	cryptoutilSharedCryptoCertificate "cryptoutil/internal/shared/crypto/certificate"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedPool "cryptoutil/internal/shared/pool"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

func TestNewGenerator_KeyPoolCreationError(t *testing.T) {
	t.Parallel()

	_, err := newGenerator(
		context.Background(),
		&cryptoutilSharedTelemetry.TelemetryService{},
		func(
			_ *cryptoutilSharedPool.ValueGenPoolConfig[*cryptoutilSharedCryptoKeygen.KeyPair],
			_ error,
		) (*cryptoutilSharedPool.ValueGenPool[*cryptoutilSharedCryptoKeygen.KeyPair], error) {
			return nil, errors.New("injected key pool creation failure")
		},
	)

	require.ErrorContains(t, err, "failed to create ECDSA P-384 key pool")
	require.ErrorContains(t, err, "injected key pool creation failure")
}

func TestCertificateChainEncodingErrors(t *testing.T) {
	t.Parallel()

	fixture := newPublishValidationFixture(t)
	subject := certificateSubjectForTest(fixture, "test-ca")
	tests := []struct {
		name    string
		run     func(*Generator, string) error
		wantErr string
	}{
		{
			name: "keystore certificate",
			run: func(gen *Generator, root string) error {
				return gen.writeKeystore(root, "keystore", subject)
			},
			wantErr: testTLSPEMEncodeCert,
		},
		{
			name: "truststore certificate",
			run: func(gen *Generator, root string) error {
				return gen.writeTruststore(root, "truststore", subject)
			},
			wantErr: testTLSPEMEncodeCert,
		},
		{
			name: "TLS config certificate",
			run: func(gen *Generator, root string) error {
				return gen.writeTLSConfigYAML(root, subject)
			},
			wantErr: "pem encode issuing CA cert chain",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gen := &Generator{
				mkdirAllFn: func(string, os.FileMode) error { return nil },
				writeFileFn: func(string, []byte, os.FileMode) error {
					return nil
				},
				encodePKCS12Fn: func(any, *x509.Certificate, []*x509.Certificate) ([]byte, error) {
					return []byte("pkcs12"), nil
				},
				encodeTrustPKCS12Fn: func([]*x509.Certificate) ([]byte, error) {
					return []byte("truststore"), nil
				},
				encodeCertChainFn: func([]*x509.Certificate) ([]byte, error) {
					return nil, errors.New("injected certificate chain encoding failure")
				},
			}

			err := tc.run(gen, t.TempDir())
			require.ErrorContains(t, err, tc.wantErr)
			require.ErrorContains(t, err, "injected certificate chain encoding failure")
		})
	}
}

func TestGeneratePSIDCerts_AdminCACertificateEncodingError(t *testing.T) {
	t.Parallel()

	fixture := newPublishValidationFixture(t)
	psID := "test-psid"
	adminSuffix := PKIInitAppInstanceSuffixes()[0]
	adminIssuingName := "private-https-mutual-issuing-ca-" + psID + "-" + adminSuffix
	adminIssuingEncodes := 0

	gen := &Generator{
		getKeyPairFn: func() *cryptoutilSharedCryptoKeygen.KeyPair {
			return nil
		},
		mkdirAllFn: func(string, os.FileMode) error { return nil },
		writeFileFn: func(string, []byte, os.FileMode) error {
			return nil
		},
		createCAFn: func(_ *cryptoutilSharedCryptoCertificate.Subject, _ any, name string, _ *cryptoutilSharedCryptoKeygen.KeyPair, _ time.Duration, _ int) (*cryptoutilSharedCryptoCertificate.Subject, error) {
			return certificateSubjectForTest(fixture, name), nil
		},
		createLeafFn: func(_ *cryptoutilSharedCryptoCertificate.Subject, _ *cryptoutilSharedCryptoKeygen.KeyPair, name string, _ time.Duration, _ []string, _ []net.IP, _ []string, _ x509.KeyUsage, _ []x509.ExtKeyUsage) (*cryptoutilSharedCryptoCertificate.Subject, error) {
			return certificateSubjectForTest(fixture, name), nil
		},
		encodePKCS12Fn: func(any, *x509.Certificate, []*x509.Certificate) ([]byte, error) {
			return []byte("pkcs12"), nil
		},
		encodeTrustPKCS12Fn: func([]*x509.Certificate) ([]byte, error) {
			return []byte("truststore"), nil
		},
		getRealmsForPSIDFn: func(string) ([]string, error) {
			return []string{"test-realm"}, nil
		},
		encodeCertChainFn: func(certs []*x509.Certificate) ([]byte, error) {
			if certs[0].Subject.CommonName == adminIssuingName {
				adminIssuingEncodes++
				if adminIssuingEncodes == 3 {
					return nil, errors.New("injected admin CA PEM encoding failure")
				}
			}

			return cryptoutilSharedCryptoAsn1.PEMEncodeCertChain(certs)
		},
	}
	shared := &sharedCAs{
		globalServerIssuing:   certificateSubjectForTest(fixture, "global-server"),
		grafanaClientIssuing:  certificateSubjectForTest(fixture, "grafana-client"),
		otelClientIssuing:     certificateSubjectForTest(fixture, "otel-client"),
		postgresServerIssuing: certificateSubjectForTest(fixture, "postgres-server"),
		postgresClientIssuing: certificateSubjectForTest(fixture, "postgres-client"),
	}

	_, err := gen.generatePSIDCerts(filepath.Join(t.TempDir(), "certs"), psID, shared)

	require.ErrorContains(t, err, "cat6 pem encode admin CA cert chain")
	require.ErrorContains(t, err, "injected admin CA PEM encoding failure")
}

func certificateSubjectForTest(fixture publishValidationFixture, name string) *cryptoutilSharedCryptoCertificate.Subject {
	certificate := *fixture.certificate
	certificate.Subject.CommonName = name

	return &cryptoutilSharedCryptoCertificate.Subject{
		KeyMaterial: cryptoutilSharedCryptoCertificate.KeyMaterial{
			PrivateKey:       fixture.privateKey,
			CertificateChain: []*x509.Certificate{&certificate},
		},
	}
}
