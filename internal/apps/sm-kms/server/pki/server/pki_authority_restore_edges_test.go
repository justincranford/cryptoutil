// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"context"
	"crypto"
	crand "crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCACrypto "cryptoutil/internal/apps/sm-kms/server/pki/crypto"
	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

type pkiAuthorityGetFailureStore struct {
	cryptoutilCAStorage.AuthorityStore
	err error
}

func (s pkiAuthorityGetFailureStore) Get(context.Context, string) (*cryptoutilCAStorage.StoredAuthority, error) {
	return nil, s.err
}

type pkiAuthorityUnencodableSigner struct {
	crypto.Signer
}

func TestGetOrCreateCAWithConfigValidationAndLookupErrors(t *testing.T) {
	t.Parallel()

	validConfig := DefaultPKIAuthorityConfig()
	encrypt := func(context.Context, []byte) ([]byte, error) { return nil, nil }
	decrypt := func(context.Context, []byte) ([]byte, error) { return nil, nil }
	failure := errors.New("injected authority lookup failure")
	validStore := pkiAuthorityGetFailureStore{err: failure}

	tests := []struct {
		name      string
		ctx       context.Context
		provider  cryptoutilCACrypto.Provider
		store     cryptoutilCAStorage.AuthorityStore
		encrypt   contentCipherFn
		decrypt   contentCipherFn
		configure func(*PKIAuthorityConfig)
		wantErr   string
	}{
		{name: pkiTestNilContext, provider: cryptoutilCACrypto.NewSoftwareProvider(), store: validStore, encrypt: encrypt, decrypt: decrypt, wantErr: pkiTestContextIsRequired},
		{name: pkiTestNilProvider, ctx: context.Background(), store: validStore, encrypt: encrypt, decrypt: decrypt, wantErr: cryptoutilSharedMagic.PKIErrorCryptoProviderRequired},
		{name: "nil store", ctx: context.Background(), provider: cryptoutilCACrypto.NewSoftwareProvider(), encrypt: encrypt, decrypt: decrypt, wantErr: "authority store is required"},
		{name: "nil encryption function", ctx: context.Background(), provider: cryptoutilCACrypto.NewSoftwareProvider(), store: validStore, decrypt: decrypt, wantErr: pkiTestEncryptionAndDecryptionRequired},
		{name: "nil decryption function", ctx: context.Background(), provider: cryptoutilCACrypto.NewSoftwareProvider(), store: validStore, encrypt: encrypt, wantErr: pkiTestEncryptionAndDecryptionRequired},
		{
			name:     "invalid authority configuration",
			ctx:      context.Background(),
			provider: cryptoutilCACrypto.NewSoftwareProvider(),
			store:    validStore,
			encrypt:  encrypt,
			decrypt:  decrypt,
			configure: func(config *PKIAuthorityConfig) {
				config.ID = ""
			},
			wantErr: "invalid PKI authority configuration",
		},
		{
			name:     pkiTestAuthorityLookupFailure,
			ctx:      context.Background(),
			provider: cryptoutilCACrypto.NewSoftwareProvider(),
			store:    validStore,
			encrypt:  encrypt,
			decrypt:  decrypt,
			wantErr:  "failed to load persisted PKI authority",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config := validConfig
			if tc.configure != nil {
				tc.configure(&config)
			}

			_, _, err := getOrCreateCAWithConfig(tc.ctx, tc.provider, crand.Reader, tc.store, tc.encrypt, tc.decrypt, config)
			require.ErrorContains(t, err, tc.wantErr)

			if tc.name == pkiTestAuthorityLookupFailure {
				require.ErrorIs(t, err, failure)
			}
		})
	}
}

func TestGetOrCreateCAWithConfigRejectsUnencodablePrivateKey(t *testing.T) {
	t.Parallel()

	baseProvider := cryptoutilCACrypto.NewSoftwareProvider()
	provider := pkiAuthorityTestProvider{
		Provider: baseProvider,
		generateKeyPair: func(spec cryptoutilCACrypto.KeySpec) (*cryptoutilCACrypto.KeyPair, error) {
			pair, err := baseProvider.GenerateKeyPair(spec)
			if err != nil {
				return nil, err
			}

			signer, ok := pair.PrivateKey.(crypto.Signer)
			require.True(t, ok)

			pair.PrivateKey = pkiAuthorityUnencodableSigner{Signer: signer}

			return pair, nil
		},
	}
	store := pkiAuthorityGetFailureStore{err: cryptoutilCAStorage.ErrAuthorityNotFound}
	encrypt := func(context.Context, []byte) ([]byte, error) { return nil, nil }
	decrypt := func(context.Context, []byte) ([]byte, error) { return nil, nil }

	_, _, err := getOrCreateCAWithConfig(
		context.Background(),
		provider,
		crand.Reader,
		store,
		encrypt,
		decrypt,
		DefaultPKIAuthorityConfig(),
	)
	require.ErrorContains(t, err, "failed to encode PKI authority private key")
}

func TestRestoreCAWithConfigFailures(t *testing.T) {
	t.Parallel()

	config := DefaultPKIAuthorityConfig()
	certificate, privateKey, err := createSelfSignedCAWithConfig(cryptoutilCACrypto.NewSoftwareProvider(), crand.Reader, config)
	require.NoError(t, err)
	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)

	validAuthority := &cryptoutilCAStorage.StoredAuthority{
		ID:                        config.ID,
		CertificateDER:            certificate.Raw,
		PrivateKeyEncrypted:       privateKeyDER,
		PrivateKeyEnvelopeVersion: cryptoutilSharedMagic.PKIPrivateKeyEnvelopeVersionV1,
	}
	copyDecrypt := func(_ context.Context, ciphertext []byte) ([]byte, error) {
		return append([]byte(nil), ciphertext...), nil
	}
	decryptFailure := errors.New("injected persisted-key decrypt failure")

	tests := []struct {
		name      string
		authority func() *cryptoutilCAStorage.StoredAuthority
		decrypt   contentCipherFn
		wantErr   string
	}{
		{
			name: pkiTestNilAuthority,
			authority: func() *cryptoutilCAStorage.StoredAuthority {
				return nil
			},
			wantErr: "persisted PKI authority is required",
		},
		{
			name: "authority ID mismatch",
			authority: func() *cryptoutilCAStorage.StoredAuthority {
				authority := *validAuthority
				authority.ID = "different-authority"

				return &authority
			},
			decrypt: copyDecrypt,
			wantErr: "does not match configured ID",
		},
		{
			name: "malformed certificate",
			authority: func() *cryptoutilCAStorage.StoredAuthority {
				authority := *validAuthority
				authority.CertificateDER = []byte("invalid certificate")

				return &authority
			},
			decrypt: copyDecrypt,
			wantErr: "failed to parse persisted PKI authority certificate",
		},
		{
			name:      pkiTestNilDecryptFunction,
			authority: func() *cryptoutilCAStorage.StoredAuthority { return validAuthority },
			wantErr:   pkiTestAuthorityKeyDecryptionRequired,
		},
		{
			name:      pkiTestDecryptFailure,
			authority: func() *cryptoutilCAStorage.StoredAuthority { return validAuthority },
			decrypt: func(context.Context, []byte) ([]byte, error) {
				return nil, decryptFailure
			},
			wantErr: "failed to decrypt persisted PKI authority private key",
		},
		{
			name: "malformed private key",
			authority: func() *cryptoutilCAStorage.StoredAuthority {
				authority := *validAuthority
				authority.PrivateKeyEncrypted = []byte("invalid private key")

				return &authority
			},
			decrypt: copyDecrypt,
			wantErr: "failed to parse persisted PKI authority private key",
		},
		{
			name: "invalid envelope version",
			authority: func() *cryptoutilCAStorage.StoredAuthority {
				authority := *validAuthority
				authority.PrivateKeyEnvelopeVersion++

				return &authority
			},
			decrypt: copyDecrypt,
			wantErr: "unsupported PKI authority private key envelope version",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			authority := tc.authority()
			_, _, restoreErr := restoreCAWithConfig(context.Background(), authority, tc.decrypt, config)
			require.ErrorContains(t, restoreErr, tc.wantErr)

			if tc.name == "decrypt failure" {
				require.ErrorIs(t, restoreErr, decryptFailure)
			}
		})
	}

	t.Run("valid authority", func(t *testing.T) {
		t.Parallel()

		restoredCertificate, restoredKey, restoreErr := restoreCAWithConfig(context.Background(), validAuthority, copyDecrypt, config)
		require.NoError(t, restoreErr)
		require.Equal(t, certificate.Raw, restoredCertificate.Raw)
		require.IsType(t, privateKey, restoredKey)
	})

	t.Run("valid PEM private key input remains rejected", func(t *testing.T) {
		t.Parallel()

		authority := *validAuthority
		authority.PrivateKeyEncrypted = pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey, Bytes: privateKeyDER})
		_, _, restoreErr := restoreCAWithConfig(
			context.Background(),
			&authority,
			copyDecrypt,
			config,
		)
		require.ErrorContains(t, restoreErr, "failed to parse persisted PKI authority private key")
	})
}

var _ crypto.Signer = pkiAuthorityInvalidPublicSigner{}
