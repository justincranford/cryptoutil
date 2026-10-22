// Copyright (c) 2025-2026 Justin Cranford.
package handler

import (
	"context"
	"crypto/x509"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilCACrypto "cryptoutil/internal/apps/sm-kms/server/pki/crypto"
	cryptoutilCAProfileSubject "cryptoutil/internal/apps/sm-kms/server/pki/profile/subject"
	cryptoutilCAServiceIssuer "cryptoutil/internal/apps/sm-kms/server/pki/service/issuer"
	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestStoreIssuedCertificate_IssuerAuthorityIdentity(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)
	testSetup.Issuer.GetCAConfig().AuthorityID = "stable-test-authority"
	keyPair, err := testSetup.Provider.GenerateKeyPair(cryptoutilCACrypto.KeySpec{
		Type:       cryptoutilSharedMagic.ECDSA,
		ECDSACurve: cryptoutilSharedMagic.P256,
	})
	require.NoError(t, err)

	issued, _, err := testSetup.Issuer.Issue(&cryptoutilCAServiceIssuer.CertificateRequest{
		SubjectRequest:   &cryptoutilCAProfileSubject.Request{CommonName: "stored.example.test"},
		PublicKey:        keyPair.PublicKey,
		ValidityDuration: time.Hour,
	})
	require.NoError(t, err)

	tests := []struct {
		name            string
		issuer          *cryptoutilCAServiceIssuer.Issuer
		mutate          func(*cryptoutilCAServiceIssuer.IssuedCertificate)
		wantAuthorityID string
		wantErr         string
	}{
		{
			name:    "missing issuer",
			wantErr: "issuing authority identity",
		},
		{
			name:            "configured issuer",
			issuer:          testSetup.Issuer,
			wantAuthorityID: testSetup.Issuer.GetCAConfig().AuthorityID,
		},
		{
			name:   "issuer distinguished name mismatch",
			issuer: testSetup.Issuer,
			mutate: func(certificate *cryptoutilCAServiceIssuer.IssuedCertificate) {
				cloned := *certificate.Certificate
				cloned.RawIssuer = []byte("CN=unexpected")
				certificate.Certificate = &cloned
			},
			wantErr: "issuer does not match",
		},
		{
			name:   testESTAuthorityKeyIdentifierMismatch,
			issuer: testSetup.Issuer,
			mutate: func(certificate *cryptoutilCAServiceIssuer.IssuedCertificate) {
				cloned := *certificate.Certificate
				cloned.AuthorityKeyId = []byte("unexpected")
				certificate.Certificate = &cloned
			},
			wantErr: "authority key identifier does not match",
		},
		{
			name:   "invalid issuer signature",
			issuer: testSetup.Issuer,
			mutate: func(certificate *cryptoutilCAServiceIssuer.IssuedCertificate) {
				cloned := *certificate.Certificate
				cloned.Signature = []byte("unexpected")
				certificate.Certificate = &cloned
			},
			wantErr: "signature does not match",
		},
		{
			name:   "serial number mismatch",
			issuer: testSetup.Issuer,
			mutate: func(certificate *cryptoutilCAServiceIssuer.IssuedCertificate) {
				certificate.SerialNumber = "1234"
			},
			wantErr: "serial number does not match",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			handler := &Handler{
				issuer:  tc.issuer,
				storage: cryptoutilCAStorage.NewMemoryStore(),
			}

			issuedCertificate := *issued

			issuedCertificate.Certificate = func() *x509.Certificate {
				cloned := *issued.Certificate

				return &cloned
			}()
			if tc.mutate != nil {
				tc.mutate(&issuedCertificate)
			}

			stored, err := handler.storeIssuedCertificate(
				context.Background(),
				&issuedCertificate,
				cryptoutilSharedMagic.PKIProfileIDTLSServer,
				googleUuid.Must(googleUuid.NewV7()),
			)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				require.Nil(t, stored)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantAuthorityID, stored.IssuerAuthorityID)
		})
	}
}
