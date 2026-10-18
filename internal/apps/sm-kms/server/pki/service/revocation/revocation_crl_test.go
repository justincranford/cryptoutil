// Copyright (c) 2025-2026 Justin Cranford.
package revocation

import (
	"context"
	"crypto"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"io"
	"math/big"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"

	cryptoutilCACrypto "cryptoutil/internal/apps/sm-kms/server/pki/crypto"
)

// TestCRLService_GetRevokedCertificates tests retrieving revoked certificates.
func TestCRLService_GetRevokedCertificates(t *testing.T) {
	t.Parallel()

	caCert, caKey := createTestCA(t)
	provider := cryptoutilCACrypto.NewSoftwareProvider()

	config := &CRLConfig{
		Issuer:     caCert,
		PrivateKey: caKey,
		Provider:   provider,
		Validity:   cryptoutilSharedMagic.HoursPerDay * time.Hour,
	}

	svc, err := NewCRLService(config)
	require.NoError(t, err)

	// Initially empty.
	revoked := svc.GetRevokedCertificates()
	require.Len(t, revoked, 0)

	// Revoke a certificate.
	serial1 := big.NewInt(12345)
	err = svc.Revoke(serial1, ReasonKeyCompromise)
	require.NoError(t, err)

	// Should have 1 revoked certificate.
	revoked = svc.GetRevokedCertificates()
	require.Len(t, revoked, 1)
	require.Equal(t, serial1, revoked[0].SerialNumber)
	require.Equal(t, ReasonKeyCompromise, revoked[0].Reason)

	// Revoke another certificate.
	serial2 := big.NewInt(67890)
	err = svc.Revoke(serial2, ReasonSuperseded)
	require.NoError(t, err)

	// Should have 2 revoked certificates.
	revoked = svc.GetRevokedCertificates()
	require.Len(t, revoked, 2)
}

func TestCRLService_UsesDurableIssuerScopedNumbers(t *testing.T) {
	t.Parallel()

	caCert, caKey := createTestCA(t)

	var (
		allocatedIssuerKeyID string
		allocationCount      int64
	)

	service, err := NewCRLService(&CRLConfig{
		Issuer:     caCert,
		PrivateKey: caKey,
		Provider:   cryptoutilCACrypto.NewSoftwareProvider(),
		Validity:   time.Hour,
		AllocateCRLNumber: func(ctx context.Context, issuerKeyID string) (*big.Int, error) {
			require.Equal(t, "allocator-context", ctx.Value(testContextKey{}))

			allocatedIssuerKeyID = issuerKeyID
			allocationCount++

			return big.NewInt(40 + allocationCount), nil
		},
	})
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), testContextKey{}, "allocator-context")

	for expected := int64(41); expected <= 42; expected++ {
		crlDER, generateErr := service.GenerateCRLWithRevokedCertificatesContext(ctx, nil)
		require.NoError(t, generateErr)

		crl, parseErr := x509.ParseRevocationList(crlDER)
		require.NoError(t, parseErr)
		require.Equal(t, big.NewInt(expected), crl.Number)
	}

	require.Equal(t, hex.EncodeToString(caCert.SubjectKeyId), allocatedIssuerKeyID)
	require.Equal(t, int64(2), allocationCount)
}

func TestCRLService_DurableNumberAllocationFailureIsReturned(t *testing.T) {
	t.Parallel()

	caCert, caKey := createTestCA(t)
	service, err := NewCRLService(&CRLConfig{
		Issuer:     caCert,
		PrivateKey: caKey,
		Provider:   cryptoutilCACrypto.NewSoftwareProvider(),
		Validity:   time.Hour,
		AllocateCRLNumber: func(context.Context, string) (*big.Int, error) {
			return nil, errors.New("sequence store unavailable")
		},
	})
	require.NoError(t, err)

	_, err = service.GenerateCRLContext(context.Background())
	require.ErrorContains(t, err, "sequence store unavailable")
}

func TestCRLService_ConsumesNumberWhenGenerationFails(t *testing.T) {
	t.Parallel()

	caCert, caKey := createTestCA(t)

	var allocatedNumber int64

	service, err := NewCRLService(&CRLConfig{
		Issuer:     caCert,
		PrivateKey: caKey,
		Provider:   &failFirstSignatureAlgorithmProvider{Provider: cryptoutilCACrypto.NewSoftwareProvider()},
		Validity:   time.Hour,
		AllocateCRLNumber: func(context.Context, string) (*big.Int, error) {
			allocatedNumber++

			return big.NewInt(allocatedNumber), nil
		},
	})
	require.NoError(t, err)

	_, err = service.GenerateCRLContext(context.Background())
	require.ErrorContains(t, err, "injected signature algorithm failure")

	crlDER, err := service.GenerateCRLContext(context.Background())
	require.NoError(t, err)

	crl, err := x509.ParseRevocationList(crlDER)
	require.NoError(t, err)
	require.Equal(t, big.NewInt(2), crl.Number)
	require.Equal(t, int64(2), allocatedNumber)
}

func TestCRLService_GenerationValidationErrors(t *testing.T) {
	t.Parallel()

	caCert, caKey := createTestCA(t)

	tests := []struct {
		name      string
		configure func(*CRLConfig)
		generate  func(*CRLService) error
		wantErr   string
	}{
		{
			name: "nil context",
			generate: func(service *CRLService) error {
				_, err := service.GenerateCRLWithRevokedCertificatesContext(nilCRLTestContext(), nil)

				return err
			},
			wantErr: testCRLContextIsRequired,
		},
		{
			name: "durable numbering requires issuer key identifier",
			configure: func(config *CRLConfig) {
				issuer := *config.Issuer
				issuer.SubjectKeyId = nil
				config.Issuer = &issuer
				config.AllocateCRLNumber = func(context.Context, string) (*big.Int, error) {
					return big.NewInt(1), nil
				}
			},
			generate: func(service *CRLService) error {
				_, err := service.GenerateCRLContext(context.Background())

				return err
			},
			wantErr: "issuer subject key identifier is required",
		},
		{
			name: "nil durable number",
			configure: func(config *CRLConfig) {
				config.AllocateCRLNumber = func(context.Context, string) (*big.Int, error) {
					return nil, nil
				}
			},
			generate: func(service *CRLService) error {
				_, err := service.GenerateCRLContext(context.Background())

				return err
			},
			wantErr: testCRLInvalidSequenceNumber,
		},
		{
			name: "nonpositive durable number",
			configure: func(config *CRLConfig) {
				config.AllocateCRLNumber = func(context.Context, string) (*big.Int, error) {
					return big.NewInt(0), nil
				}
			},
			generate: func(service *CRLService) error {
				_, err := service.GenerateCRLContext(context.Background())

				return err
			},
			wantErr: testCRLInvalidSequenceNumber,
		},
		{
			name: "CRL signer failure",
			configure: func(config *CRLConfig) {
				config.PrivateKey = failingCRLSigner{publicKey: config.Issuer.PublicKey}
			},
			generate: func(service *CRLService) error {
				_, err := service.GenerateCRLContext(context.Background())

				return err
			},
			wantErr: "failed to create CRL",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config := &CRLConfig{
				Issuer:     caCert,
				PrivateKey: caKey,
				Provider:   cryptoutilCACrypto.NewSoftwareProvider(),
				Validity:   time.Hour,
			}
			if tc.configure != nil {
				tc.configure(config)
			}

			service, err := NewCRLService(config)
			require.NoError(t, err)
			require.ErrorContains(t, tc.generate(service), tc.wantErr)
		})
	}
}

func TestCRLService_PEMWrappersPropagateGenerationErrors(t *testing.T) {
	t.Parallel()

	caCert, caKey := createTestCA(t)
	tests := []struct {
		name    string
		call    func(*CRLService) error
		wantErr string
	}{
		{
			name:    "basic PEM wrapper",
			wantErr: testCRLFailedToGetSignature,
			call: func(service *CRLService) error {
				_, err := service.GenerateCRLPEM()

				return err
			},
		},
		{
			name:    "context PEM wrapper",
			wantErr: testCRLContextIsRequired,
			call: func(service *CRLService) error {
				_, err := service.GenerateCRLPEMContext(nilCRLTestContext())

				return err
			},
		},
		{
			name:    "revoked certificate PEM wrapper",
			wantErr: testCRLFailedToGetSignature,
			call: func(service *CRLService) error {
				_, err := service.GenerateCRLPEMWithRevokedCertificates(nil)

				return err
			},
		},
		{
			name:    "context revoked certificate PEM wrapper",
			wantErr: testCRLContextIsRequired,
			call: func(service *CRLService) error {
				_, err := service.GenerateCRLPEMWithRevokedCertificatesContext(nilCRLTestContext(), nil)

				return err
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			service, err := NewCRLService(&CRLConfig{
				Issuer:     caCert,
				PrivateKey: caKey,
				Provider:   &failFirstSignatureAlgorithmProvider{Provider: cryptoutilCACrypto.NewSoftwareProvider()},
				Validity:   time.Hour,
			})
			require.NoError(t, err)

			err = tc.call(service)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func nilCRLTestContext() context.Context {
	return nil
}

type failingCRLSigner struct {
	publicKey crypto.PublicKey
}

func (s failingCRLSigner) Public() crypto.PublicKey {
	return s.publicKey
}

func (failingCRLSigner) Sign(io.Reader, []byte, crypto.SignerOpts) ([]byte, error) {
	return nil, errors.New("injected CRL signing failure")
}

type failFirstSignatureAlgorithmProvider struct {
	cryptoutilCACrypto.Provider
	calls int
}

func (p *failFirstSignatureAlgorithmProvider) GetSignatureAlgorithm(publicKey crypto.PublicKey) (x509.SignatureAlgorithm, error) {
	p.calls++
	if p.calls == 1 {
		return 0, errors.New("injected signature algorithm failure")
	}

	return p.Provider.GetSignatureAlgorithm(publicKey)
}

type testContextKey struct{}
