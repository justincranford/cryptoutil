// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"context"
	"crypto"
	crand "crypto/rand"
	"crypto/x509"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilCACrypto "cryptoutil/internal/apps/sm-kms/server/pki/crypto"
	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const historicalAuthorityTestID = "historical-test-authority"

func TestRestoreHistoricalCA_ValidationAndRestore(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	certificate, privateKey, err := createSelfSignedCA(provider, crand.Reader)
	require.NoError(t, err)
	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)

	validAuthority := &cryptoutilCAStorage.StoredAuthority{
		ID:                        historicalAuthorityTestID,
		CertificateDER:            certificate.Raw,
		PrivateKeyEncrypted:       privateKeyDER,
		PrivateKeyEnvelopeVersion: cryptoutilSharedMagic.PKIPrivateKeyEnvelopeVersionV1,
	}
	decrypt := func(_ context.Context, ciphertext []byte) ([]byte, error) {
		return append([]byte(nil), ciphertext...), nil
	}

	tests := []struct {
		name      string
		ctx       context.Context
		authority func(*testing.T) *cryptoutilCAStorage.StoredAuthority
		decrypt   contentCipherFn
		wantErr   string
	}{
		{
			name:      "restore valid historical authority",
			ctx:       context.Background(),
			authority: func(*testing.T) *cryptoutilCAStorage.StoredAuthority { return validAuthority },
			decrypt:   decrypt,
		},
		{
			name:      pkiTestNilContext,
			authority: func(*testing.T) *cryptoutilCAStorage.StoredAuthority { return validAuthority },
			decrypt:   decrypt,
			wantErr:   pkiTestContextIsRequired,
		},
		{
			name: pkiTestNilAuthority,
			ctx:  context.Background(),
			authority: func(*testing.T) *cryptoutilCAStorage.StoredAuthority {
				return nil
			},
			decrypt: decrypt,
			wantErr: pkiTestNonEmptyIDRequired,
		},
		{
			name: "empty authority ID",
			ctx:  context.Background(),
			authority: func(*testing.T) *cryptoutilCAStorage.StoredAuthority {
				copy := *validAuthority
				copy.ID = ""

				return &copy
			},
			decrypt: decrypt,
			wantErr: pkiTestNonEmptyIDRequired,
		},
		{
			name: "untrimmed authority ID",
			ctx:  context.Background(),
			authority: func(*testing.T) *cryptoutilCAStorage.StoredAuthority {
				copy := *validAuthority
				copy.ID = " " + historicalAuthorityTestID

				return &copy
			},
			decrypt: decrypt,
			wantErr: pkiTestNonEmptyIDRequired,
		},
		{
			name: pkiTestNilDecryptFunction,
			ctx:  context.Background(),
			authority: func(*testing.T) *cryptoutilCAStorage.StoredAuthority {
				return validAuthority
			},
			wantErr: pkiTestAuthorityKeyDecryptionRequired,
		},
		{
			name: "invalid certificate DER",
			ctx:  context.Background(),
			authority: func(*testing.T) *cryptoutilCAStorage.StoredAuthority {
				copy := *validAuthority
				copy.CertificateDER = []byte("invalid certificate")

				return &copy
			},
			decrypt: decrypt,
			wantErr: "failed to parse historical PKI authority certificate",
		},
		{
			name: pkiTestCertificateIsNotCA,
			ctx:  context.Background(),
			authority: func(t *testing.T) *cryptoutilCAStorage.StoredAuthority {
				return historicalAuthorityWithTemplate(t, certificate, privateKey, func(template *x509.Certificate) {
					template.IsCA = false
					template.BasicConstraintsValid = true
					template.KeyUsage = 0
					template.MaxPathLen = -1
					template.MaxPathLenZero = false
				})
			},
			decrypt: decrypt,
			wantErr: pkiTestMustBeCAWithValidBasicConstraints,
		},
		{
			name: "certificate lacks signing key usage",
			ctx:  context.Background(),
			authority: func(t *testing.T) *cryptoutilCAStorage.StoredAuthority {
				return historicalAuthorityWithTemplate(t, certificate, privateKey, func(template *x509.Certificate) {
					template.KeyUsage = x509.KeyUsageDigitalSignature
				})
			},
			decrypt: decrypt,
			wantErr: "must permit certificate and CRL signing",
		},
		{
			name: "certificate is not self-signed",
			ctx:  context.Background(),
			authority: func(t *testing.T) *cryptoutilCAStorage.StoredAuthority {
				return historicalAuthorityWithNonSelfSignedCertificate(t, validAuthority, certificate, privateKey)
			},
			decrypt: decrypt,
			wantErr: pkiTestNotSelfSigned,
		},
		{
			name: pkiTestCertificateNotYetValid,
			ctx:  context.Background(),
			authority: func(t *testing.T) *cryptoutilCAStorage.StoredAuthority {
				return historicalAuthorityWithTemplate(t, certificate, privateKey, func(template *x509.Certificate) {
					template.NotBefore = time.Now().UTC().Add(time.Hour)
				})
			},
			decrypt: decrypt,
			wantErr: "not yet valid",
		},
		{
			name: "decryption failure",
			ctx:  context.Background(),
			authority: func(*testing.T) *cryptoutilCAStorage.StoredAuthority {
				return validAuthority
			},
			decrypt: func(context.Context, []byte) ([]byte, error) {
				return nil, errors.New("injected historical decrypt failure")
			},
			wantErr: "injected historical decrypt failure",
		},
		{
			name: "invalid private key DER",
			ctx:  context.Background(),
			authority: func(*testing.T) *cryptoutilCAStorage.StoredAuthority {
				copy := *validAuthority
				copy.PrivateKeyEncrypted = []byte("invalid private key")

				return &copy
			},
			decrypt: decrypt,
			wantErr: "failed to parse historical PKI authority private key",
		},
		{
			name: "certificate and private key mismatch",
			ctx:  context.Background(),
			authority: func(t *testing.T) *cryptoutilCAStorage.StoredAuthority {
				otherCertificate, _, createErr := createSelfSignedCA(provider, crand.Reader)
				require.NoError(t, createErr)

				copy := *validAuthority
				copy.CertificateDER = otherCertificate.Raw

				return &copy
			},
			decrypt: decrypt,
			wantErr: pkiTestCertificateAndPrivateKeyMismatch,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			authority := tc.authority(t)

			restoredCertificate, restoredKey, restoreErr := restoreHistoricalCA(tc.ctx, authority, tc.decrypt)
			if tc.wantErr != "" {
				require.ErrorContains(t, restoreErr, tc.wantErr)
				require.Nil(t, restoredCertificate)
				require.Nil(t, restoredKey)

				return
			}

			require.NoError(t, restoreErr)
			require.Equal(t, certificate.Raw, restoredCertificate.Raw)
			require.NotNil(t, restoredKey)
		})
	}
}

func historicalAuthorityWithTemplate(
	t *testing.T,
	certificate *x509.Certificate,
	privateKey any,
	mutate func(*x509.Certificate),
) *cryptoutilCAStorage.StoredAuthority {
	t.Helper()

	template := *certificate
	template.NotBefore = time.Now().UTC().Add(-time.Hour)
	template.NotAfter = time.Now().UTC().Add(24 * time.Hour)
	mutate(&template)

	certificateDER, err := x509.CreateCertificate(crand.Reader, &template, &template, certificate.PublicKey, privateKey)
	require.NoError(t, err)

	keyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)

	return &cryptoutilCAStorage.StoredAuthority{
		ID:                        historicalAuthorityTestID,
		CertificateDER:            certificateDER,
		PrivateKeyEncrypted:       keyDER,
		PrivateKeyEnvelopeVersion: cryptoutilSharedMagic.PKIPrivateKeyEnvelopeVersionV1,
	}
}

func historicalAuthorityWithNonSelfSignedCertificate(
	t *testing.T,
	authority *cryptoutilCAStorage.StoredAuthority,
	certificate *x509.Certificate,
	privateKey any,
) *cryptoutilCAStorage.StoredAuthority {
	t.Helper()

	template := *certificate
	template.NotBefore = time.Now().UTC().Add(-time.Hour)
	template.NotAfter = time.Now().UTC().Add(24 * time.Hour)
	_, otherPrivateKey, err := createSelfSignedCA(cryptoutilCACrypto.NewSoftwareProvider(), crand.Reader)
	require.NoError(t, err)

	otherSigner, ok := otherPrivateKey.(crypto.Signer)
	require.True(t, ok)

	parent := *certificate
	parent.PublicKey = otherSigner.Public()
	certificateDER, err := x509.CreateCertificate(crand.Reader, &template, &parent, certificate.PublicKey, otherPrivateKey)
	require.NoError(t, err)

	copy := *authority
	copy.CertificateDER = certificateDER

	return &copy
}
