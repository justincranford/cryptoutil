// Copyright (c) 2025 Justin Cranford
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

// Sequential: modifies package-level seam variables (jwkKeySet, etc.).
func TestEdDSASetError(t *testing.T) {
	saveRestoreSeams(t)

	kid := googleUuid.Must(googleUuid.NewV7())
	edAlg := AlgEdDSA

	_, edPriv, err := ed25519.GenerateKey(crand.Reader)
	require.NoError(t, err)

	edKey := &cryptoutilSharedCryptoKeygen.KeyPair{Private: edPriv, Public: edPriv.Public()}

	// JWS: Set error on 1st Set (kty OKP).
	t.Run("JWS_EdDSA_set1", func(t *testing.T) {
		resetSeams()

		jwkKeySet = countingSet(1)
		_, _, _, _, _, err := CreateJWSJWKFromKey(&kid, &edAlg, edKey)
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
		resetSeams()

		jwkKeySet = countingSet(1)
		_, _, _, _, _, err := CreateJWEJWKFromKey(&kid, &encAlg, &ecdhAlg, ecdhKey)
		require.Error(t, err, "expected Set error for JWE ECDH kty Set")
		require.True(t, errors.Is(err, errInjected), "error should wrap errInjected, got: %v", err)
	})

	// JWK: Set error on 1st Set (kty OKP).
	genAlg := cryptoutilOpenapiModel.OKPEd25519

	t.Run("JWK_EdDSA_set1", func(t *testing.T) {
		resetSeams()

		jwkKeySet = countingSet(1)
		_, _, _, _, _, err := CreateJWKFromKey(&kid, &genAlg, edKey)
		require.Error(t, err, "expected Set error for JWK EdDSA kty Set")
		require.True(t, errors.Is(err, errInjected), "error should wrap errInjected, got: %v", err)
	})
}

// Sequential: modifies package-level seam variables (generateRSAKeyPair, generateECDSAKeyPair, etc.).
func TestValidateOrGenerate_KeygenErrors(t *testing.T) {
	saveRestoreSeams(t)

	t.Run("RSA_keygen_error", func(t *testing.T) {
		resetSeams()

		generateRSAKeyPair = func(_ int) (*cryptoutilSharedCryptoKeygen.KeyPair, error) {
			return nil, errInjected
		}
		_, err := validateOrGenerateRSAJWK(nil, cryptoutilSharedMagic.RSAKeySize2048)
		require.Error(t, err)
		require.True(t, errors.Is(err, errInjected))
	})

	t.Run("ECDSA_keygen_error", func(t *testing.T) {
		resetSeams()

		generateECDSAKeyPair = func(_ elliptic.Curve) (*cryptoutilSharedCryptoKeygen.KeyPair, error) {
			return nil, errInjected
		}
		_, err := validateOrGenerateEcdsaJWK(nil, elliptic.P256())
		require.Error(t, err)
		require.True(t, errors.Is(err, errInjected))
	})

	t.Run("EDDSA_keygen_error", func(t *testing.T) {
		resetSeams()

		generateEDDSAKeyPair = func(_ string) (*cryptoutilSharedCryptoKeygen.KeyPair, error) {
			return nil, errInjected
		}
		_, err := validateOrGenerateEddsaJWK(nil, cryptoutilSharedMagic.EdCurveEd25519)
		require.Error(t, err)
		require.True(t, errors.Is(err, errInjected))
	})

	t.Run("HMAC_keygen_error", func(t *testing.T) {
		resetSeams()

		generateHMACKey = func(_ int) (cryptoutilSharedCryptoKeygen.SecretKey, error) {
			return nil, errInjected
		}
		_, err := validateOrGenerateHMACJWK(nil, cryptoutilSharedMagic.HMACKeySize256)
		require.Error(t, err)
		require.True(t, errors.Is(err, errInjected))
	})

	t.Run("AES_keygen_error", func(t *testing.T) {
		resetSeams()

		generateAESKey = func(_ int) (cryptoutilSharedCryptoKeygen.SecretKey, error) {
			return nil, errInjected
		}
		_, err := validateOrGenerateAESJWK(nil, cryptoutilSharedMagic.AESKeySize256)
		require.Error(t, err)
		require.True(t, errors.Is(err, errInjected))
	})
}
