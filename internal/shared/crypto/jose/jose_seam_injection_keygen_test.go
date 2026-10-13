// Copyright (c) 2025-2026 Justin Cranford.
//
//

package crypto

import (
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/elliptic"
	crand "crypto/rand"
	"errors"
	"testing"

	cryptoutilOpenapiModel "cryptoutil/api/sm-kms/models"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const (
	testKeygenHMACLabel = "HMAC"
	testKeygenAESLabel  = "AES"
)

func TestEdDSASetError(t *testing.T) {
	t.Parallel()

	kid := googleUuid.Must(googleUuid.NewV7())
	edAlg := AlgEdDSA

	_, edPriv, err := ed25519.GenerateKey(crand.Reader)
	require.NoError(t, err)

	edKey := &cryptoutilSharedCryptoKeygen.KeyPair{Private: edPriv, Public: edPriv.Public()}

	// JWS: Set error on 1st Set (kty OKP).
	t.Run("JWS_EdDSA_set1", func(t *testing.T) {
		t.Parallel()

		operations := newJOSEOperations()
		operations.setJWK = countingSet(1)
		_, _, _, _, _, err := createJWSJWKFromKeyWithOperations(&kid, &edAlg, edKey, operations)
		require.Error(t, err, "expected Set error for JWS EdDSA kty Set")
		require.True(t, errors.Is(err, errInjected), "error should wrap errInjected, got: %v", err)
	})

	// JWE: Set error on 1st Set (kty OKP) — use ECDH-ES with ecdh key.
	ecdhAlg := AlgECDHES
	encAlg := EncA256GCM
	ecdhPrivKey, err := ecdh.P521().GenerateKey(crand.Reader)
	require.NoError(t, err)

	ecdhKey := &cryptoutilSharedCryptoKeygen.KeyPair{Private: ecdhPrivKey, Public: ecdhPrivKey.PublicKey()}

	t.Run("JWE_ECDH_set1", func(t *testing.T) {
		t.Parallel()

		operations := newJOSEOperations()
		operations.setJWK = countingSet(1)
		_, _, _, _, _, err := createJWEJWKFromKeyWithOperations(&kid, &encAlg, &ecdhAlg, ecdhKey, operations)
		require.Error(t, err, "expected Set error for JWE ECDH kty Set")
		require.True(t, errors.Is(err, errInjected), "error should wrap errInjected, got: %v", err)
	})

	// JWK: Set error on 1st Set (kty OKP).
	genAlg := cryptoutilOpenapiModel.OKPEd25519

	t.Run("JWK_EdDSA_set1", func(t *testing.T) {
		t.Parallel()

		operations := newJOSEOperations()
		operations.setJWK = countingSet(1)
		_, _, _, _, _, err := createJWKFromKeyWithOperations(&kid, &genAlg, edKey, operations)
		require.Error(t, err, "expected Set error for JWK EdDSA kty Set")
		require.True(t, errors.Is(err, errInjected), "error should wrap errInjected, got: %v", err)
	})
}

func TestValidateOrGenerate_KeygenErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{
			name: cryptoutilSharedMagic.KeyTypeRSA,
			run: func(t *testing.T) {
				key, err := validateOrGenerateRSAJWK(nil, cryptoutilSharedMagic.RSAKeySize2048)
				require.NoError(t, err)
				require.NotNil(t, key)

				_, err = validateOrGenerateRSAJWKWithGenerator(nil, cryptoutilSharedMagic.RSAKeySize2048, func(_ int) (*cryptoutilSharedCryptoKeygen.KeyPair, error) {
					return nil, errInjected
				})
				require.ErrorIs(t, err, errInjected)
			},
		},
		{
			name: cryptoutilSharedMagic.ECDSA,
			run: func(t *testing.T) {
				key, err := validateOrGenerateEcdsaJWK(nil, elliptic.P256())
				require.NoError(t, err)
				require.NotNil(t, key)

				_, err = validateOrGenerateEcdsaJWKWithGenerator(nil, elliptic.P256(), func(_ elliptic.Curve) (*cryptoutilSharedCryptoKeygen.KeyPair, error) {
					return nil, errInjected
				})
				require.ErrorIs(t, err, errInjected)
			},
		},
		{
			name: cryptoutilSharedMagic.JoseAlgEdDSA,
			run: func(t *testing.T) {
				key, err := validateOrGenerateEddsaJWK(nil, cryptoutilSharedMagic.EdCurveEd25519)
				require.NoError(t, err)
				require.NotNil(t, key)

				_, err = validateOrGenerateEddsaJWKWithGenerator(nil, cryptoutilSharedMagic.EdCurveEd25519, func(_ string) (*cryptoutilSharedCryptoKeygen.KeyPair, error) {
					return nil, errInjected
				})
				require.ErrorIs(t, err, errInjected)
			},
		},
		{
			name: testKeygenHMACLabel,
			run: func(t *testing.T) {
				key, err := validateOrGenerateHMACJWK(nil, cryptoutilSharedMagic.HMACKeySize256)
				require.NoError(t, err)
				require.NotEmpty(t, key)

				_, err = validateOrGenerateHMACJWKWithGenerator(nil, cryptoutilSharedMagic.HMACKeySize256, func(_ int) (cryptoutilSharedCryptoKeygen.SecretKey, error) {
					return nil, errInjected
				})
				require.ErrorIs(t, err, errInjected)
			},
		},
		{
			name: testKeygenAESLabel,
			run: func(t *testing.T) {
				key, err := validateOrGenerateAESJWK(nil, cryptoutilSharedMagic.AESKeySize256)
				require.NoError(t, err)
				require.NotEmpty(t, key)

				_, err = validateOrGenerateAESJWKWithGenerator(nil, cryptoutilSharedMagic.AESKeySize256, func(_ int) (cryptoutilSharedCryptoKeygen.SecretKey, error) {
					return nil, errInjected
				})
				require.ErrorIs(t, err, errInjected)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.run(t)
		})
	}
}
