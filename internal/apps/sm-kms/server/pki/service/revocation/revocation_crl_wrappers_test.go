// Copyright (c) 2025-2026 Justin Cranford.
package revocation

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	cryptoutilCACrypto "cryptoutil/internal/apps/sm-kms/server/pki/crypto"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestCRLService_GenerateCRLWrappers(t *testing.T) {
	t.Parallel()

	caCert, caKey := createTestCA(t)
	service, err := NewCRLService(&CRLConfig{
		Issuer:     caCert,
		PrivateKey: caKey,
		Provider:   cryptoutilCACrypto.NewSoftwareProvider(),
		Validity:   cryptoutilSharedMagic.HoursPerDay * time.Hour,
	})
	require.NoError(t, err)

	revokedCertificates := []RevokedCertificate{
		{
			SerialNumber:   big.NewInt(700),
			RevocationTime: time.Now().UTC(),
			Reason:         ReasonKeyCompromise,
		},
	}

	tests := []struct {
		name        string
		call        func(*CRLService) ([]byte, error)
		wantPEM     bool
		wantRevoked int
	}{
		{
			name: "DER from supplied revoked certificates",
			call: func(service *CRLService) ([]byte, error) {
				return service.GenerateCRLWithRevokedCertificates(revokedCertificates)
			},
			wantRevoked: 1,
		},
		{
			name: "PEM with context",
			call: func(service *CRLService) ([]byte, error) {
				return service.GenerateCRLPEMContext(context.Background())
			},
			wantPEM: true,
		},
		{
			name: "PEM from supplied revoked certificates",
			call: func(service *CRLService) ([]byte, error) {
				return service.GenerateCRLPEMWithRevokedCertificates(revokedCertificates)
			},
			wantPEM:     true,
			wantRevoked: 1,
		},
		{
			name: "PEM from supplied revoked certificates with context",
			call: func(service *CRLService) ([]byte, error) {
				return service.GenerateCRLPEMWithRevokedCertificatesContext(context.Background(), revokedCertificates)
			},
			wantPEM:     true,
			wantRevoked: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			crlBytes, err := tc.call(service)
			require.NoError(t, err)
			require.NotEmpty(t, crlBytes)

			if tc.wantPEM {
				block, rest := pem.Decode(crlBytes)
				require.NotNil(t, block)
				require.Empty(t, rest)
				require.Equal(t, "X509 CRL", block.Type)
				crlBytes = block.Bytes
			}

			crl, err := x509.ParseRevocationList(crlBytes)
			require.NoError(t, err)
			require.Len(t, crl.RevokedCertificateEntries, tc.wantRevoked)
		})
	}
}
