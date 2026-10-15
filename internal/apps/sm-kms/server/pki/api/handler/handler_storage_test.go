// Copyright (c) 2025-2026 Justin Cranford.
package handler

import (
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilCAServiceIssuer "cryptoutil/internal/apps/sm-kms/server/pki/service/issuer"
	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const testStoredCertificateSerialNumber = "1234"

func TestStoreIssuedCertificate_IssuerAuthorityIdentity(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)
	testSetup.Issuer.GetCAConfig().AuthorityID = "stable-test-authority"
	serialNumber, ok := new(big.Int).SetString(testStoredCertificateSerialNumber, cryptoutilSharedMagic.HexBase)
	require.True(t, ok)

	issued := &cryptoutilCAServiceIssuer.IssuedCertificate{
		Certificate: &x509.Certificate{
			SerialNumber: serialNumber,
			Subject:      pkix.Name{CommonName: "stored.example.test"},
			Issuer:       pkix.Name{CommonName: "test-ca"},
			NotBefore:    time.Now().UTC(),
			NotAfter:     time.Now().UTC().Add(time.Hour),
		},
		CertificatePEM: []byte("issued certificate"),
		SerialNumber:   testStoredCertificateSerialNumber,
	}

	tests := []struct {
		name            string
		issuer          *cryptoutilCAServiceIssuer.Issuer
		wantAuthorityID string
		wantErr         bool
	}{
		{
			name:    "missing issuer",
			wantErr: true,
		},
		{
			name:            "configured issuer",
			issuer:          testSetup.Issuer,
			wantAuthorityID: testSetup.Issuer.GetCAConfig().AuthorityID,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			handler := &Handler{
				issuer:  tc.issuer,
				storage: cryptoutilCAStorage.NewMemoryStore(),
			}

			stored, err := handler.storeIssuedCertificate(
				context.Background(),
				issued,
				cryptoutilSharedMagic.PKIProfileIDTLSServer,
				googleUuid.Nil,
			)
			if tc.wantErr {
				require.ErrorContains(t, err, "issuing authority identity")
				require.Nil(t, stored)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantAuthorityID, stored.IssuerAuthorityID)
		})
	}
}
