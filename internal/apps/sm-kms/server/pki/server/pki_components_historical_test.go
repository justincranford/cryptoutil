// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"context"
	crand "crypto/rand"
	"crypto/x509"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCACrypto "cryptoutil/internal/apps/sm-kms/server/pki/crypto"
	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const pkiComponentHistoricalAuthorityID = "historical-pki-authority"

func TestNewPKIComponentsPersistedAuthorityHandling(t *testing.T) {
	t.Parallel()

	historicalCert, historicalKey, err := createSelfSignedCA(cryptoutilCACrypto.NewSoftwareProvider(), crand.Reader)
	require.NoError(t, err)
	historicalKeyDER, err := x509.MarshalPKCS8PrivateKey(historicalKey)
	require.NoError(t, err)

	historicalAuthority := &cryptoutilCAStorage.StoredAuthority{
		ID:                        pkiComponentHistoricalAuthorityID,
		CertificateDER:            historicalCert.Raw,
		PrivateKeyEncrypted:       historicalKeyDER,
		PrivateKeyEnvelopeVersion: cryptoutilSharedMagic.PKIPrivateKeyEnvelopeVersionV1,
	}
	failure := errors.New(pkiComponentInjectedFailure)

	tests := []struct {
		name           string
		listFn         func(context.Context) ([]*cryptoutilCAStorage.StoredAuthority, error)
		decryptFn      contentCipherFn
		wantErr        string
		wantHistorical bool
	}{
		{
			name: "persisted authority list failure",
			listFn: func(context.Context) ([]*cryptoutilCAStorage.StoredAuthority, error) {
				return nil, failure
			},
			wantErr: pkiTestFailedListPersistedPKIAuthorities,
		},
		{
			name: "nil authority entry",
			listFn: func(context.Context) ([]*cryptoutilCAStorage.StoredAuthority, error) {
				return []*cryptoutilCAStorage.StoredAuthority{nil}, nil
			},
			wantErr: pkiTestAuthorityNilEntry,
		},
		{
			name: "invalid historical authority",
			listFn: func(context.Context) ([]*cryptoutilCAStorage.StoredAuthority, error) {
				return []*cryptoutilCAStorage.StoredAuthority{{
					ID:                        pkiComponentHistoricalAuthorityID,
					PrivateKeyEnvelopeVersion: cryptoutilSharedMagic.PKIPrivateKeyEnvelopeVersionV1 + 1,
				}}, nil
			},
			decryptFn: func(context.Context, []byte) ([]byte, error) {
				return nil, errors.New("unexpected historical decrypt")
			},
			wantErr: "failed to restore historical PKI authority",
		},
		{
			name: "restore historical authority and skip active authority",
			listFn: func(context.Context) ([]*cryptoutilCAStorage.StoredAuthority, error) {
				return []*cryptoutilCAStorage.StoredAuthority{
					{ID: pkiComponentStableAuthorityID},
					historicalAuthority,
				}, nil
			},
			decryptFn: func(_ context.Context, ciphertext []byte) ([]byte, error) {
				return append([]byte(nil), ciphertext...), nil
			},
			wantHistorical: true,
		},
		{
			name: "duplicate historical authority endpoint",
			listFn: func(context.Context) ([]*cryptoutilCAStorage.StoredAuthority, error) {
				return []*cryptoutilCAStorage.StoredAuthority{historicalAuthority, historicalAuthority}, nil
			},
			decryptFn: func(_ context.Context, ciphertext []byte) ([]byte, error) {
				return append([]byte(nil), ciphertext...), nil
			},
			wantErr: "duplicate public revocation endpoint authority ID",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			factories := pkiComponentFactoriesForTest()
			factories.authorityID = pkiComponentStableAuthorityID
			factories.listAuthoritiesFn = tc.listFn
			factories.decryptAuthorityKeyFn = tc.decryptFn

			components, buildErr := newPKIComponents(factories)
			if tc.wantErr != "" {
				require.ErrorContains(t, buildErr, tc.wantErr)
				require.Nil(t, components)

				return
			}

			require.NoError(t, buildErr)
			require.NotNil(t, components)

			if tc.wantHistorical {
				require.Contains(t, components.publicRevocationEndpoints, pkiComponentHistoricalAuthorityID)
			}
		})
	}
}
