// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"crypto/ecdh"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"testing"

	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/stretchr/testify/require"
)

func TestValidateOrGenerateJWEEcdhJWK_WrongPrivateKeyType(t *testing.T) {
	t.Parallel()

	// Generate RSA key instead of ECDH.
	rsaKey, err := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: rsaKey,
		Public:  &rsaKey.PublicKey,
	}

	result, err := validateOrGenerateJWEEcdhJWK(keyPair, &EncA256GCM, &AlgECDHES, ecdh.P256(), &EncA256GCM)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "unsupported key type *rsa.PrivateKey")
}

func TestValidateOrGenerateJWEEcdhJWK_WrongPublicKeyType(t *testing.T) {
	t.Parallel()

	// Generate ECDH private + ECDSA public (mismatched types).
	ecdhKey, err := ecdh.P256().GenerateKey(crand.Reader)
	require.NoError(t, err)

	ecdsaKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: ecdhKey,
		Public:  &ecdsaKey.PublicKey,
	}

	result, err := validateOrGenerateJWEEcdhJWK(keyPair, &EncA256GCM, &AlgECDHES, ecdh.P256(), &EncA256GCM)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "unsupported key type *ecdsa.PublicKey")
}

func TestValidateOrGenerateJWEEcdhJWK_DisallowedEnc(t *testing.T) {
	t.Parallel()

	// Test enc not in allowedEncs list.
	ecdhKey, err := ecdh.P256().GenerateKey(crand.Reader)
	require.NoError(t, err)

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: ecdhKey,
		Public:  ecdhKey.PublicKey(),
	}

	// Use A128GCM but only allow A256GCM.
	result, err := validateOrGenerateJWEEcdhJWK(keyPair, &EncA128GCM, &AlgECDHES, ecdh.P256(), &EncA256GCM)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "enc A128GCM not allowed")
}

func TestValidateOrGenerateJWEEcdhJWK_NilPrivateKey(t *testing.T) {
	t.Parallel()

	// Generate ECDH public key, create KeyPair with typed nil private.
	ecdhPriv, err := ecdh.P256().GenerateKey(crand.Reader)
	require.NoError(t, err)

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: (*ecdh.PrivateKey)(nil), // Typed nil to pass type check
		Public:  ecdhPriv.PublicKey(),
	}

	result, err := validateOrGenerateJWEEcdhJWK(keyPair, &EncA256GCM, &AlgECDHES, ecdh.P256(), &EncA256GCM)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "unsupported nil ECDH private key")
}

func TestValidateOrGenerateJWEEcdhJWK_NilPublicKey(t *testing.T) {
	t.Parallel()

	// Generate ECDH private key, create KeyPair with typed nil public.
	ecdhPriv, err := ecdh.P256().GenerateKey(crand.Reader)
	require.NoError(t, err)

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: ecdhPriv,
		Public:  (*ecdh.PublicKey)(nil), // Typed nil to pass type check
	}

	result, err := validateOrGenerateJWEEcdhJWK(keyPair, &EncA256GCM, &AlgECDHES, ecdh.P256(), &EncA256GCM)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "unsupported nil ECDH public key")
}

func TestValidateOrGenerateJWEEcdhJWK_Generate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		enc         *joseJwa.ContentEncryptionAlgorithm
		alg         *joseJwa.KeyEncryptionAlgorithm
		allowedEncs []*joseJwa.ContentEncryptionAlgorithm
		expectError bool
	}{
		{
			name:        "ECDH-ES with A256GCM P256",
			enc:         &EncA256GCM,
			alg:         &AlgECDHES,
			allowedEncs: []*joseJwa.ContentEncryptionAlgorithm{&EncA256GCM},
			expectError: false,
		},
		{
			name:        "ECDH-ES+A256KW with A192GCM P384",
			enc:         &EncA192GCM,
			alg:         &AlgECDHESA256KW,
			allowedEncs: []*joseJwa.ContentEncryptionAlgorithm{&EncA192GCM},
			expectError: false,
		},
		{
			name:        "disallowed enc",
			enc:         &EncA128GCM,
			alg:         &AlgECDHES,
			allowedEncs: []*joseJwa.ContentEncryptionAlgorithm{&EncA256GCM},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Use P256 curve for ECDH key generation.
			keyPair, err := validateOrGenerateJWEEcdhJWK(nil, tc.enc, tc.alg, ecdh.P256(), tc.allowedEncs...)

			if tc.expectError {
				require.Error(t, err)
				require.Nil(t, keyPair)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, keyPair)
			require.NotNil(t, keyPair.Private)
			require.NotNil(t, keyPair.Public)
		})
	}
}

func TestCreateJWEJWKFromKey_AESSecretKey(t *testing.T) {
	t.Parallel()

	// AES secret key has no public key component.
	kid := googleUuid.New()
	enc := joseJwa.A256GCM()
	alg := joseJwa.A256GCMKW()
	key := make(cryptoutilSharedCryptoKeygen.SecretKey, cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes)
	_, _ = crand.Read(key)

	_, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWEJWKFromKey(&kid, &enc, &alg, key)
	require.NoError(t, err)
	require.NotNil(t, nonPublicJWK)
	require.Nil(t, publicJWK) // AES should have no public key
	require.Empty(t, publicBytes)
	require.NotEmpty(t, nonPublicBytes)
}

func TestCreateJWEJWKFromKey_ECDHKeyPair(t *testing.T) {
	t.Parallel()

	// ECDH key pair has public key component.
	kid := googleUuid.New()
	enc := joseJwa.A256GCM()
	alg := joseJwa.ECDH_ES_A256KW()
	keyPair, err := cryptoutilSharedCryptoKeygen.GenerateECDHKeyPair(ecdh.P256())
	require.NoError(t, err)

	_, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWEJWKFromKey(&kid, &enc, &alg, keyPair)
	require.NoError(t, err)
	require.NotNil(t, nonPublicJWK)
	require.NotNil(t, publicJWK) // ECDH should have public key
	require.NotEmpty(t, publicBytes)
	require.NotEmpty(t, nonPublicBytes)
}

func TestCreateJWEJWKFromKey_RSAKeyPair(t *testing.T) {
	t.Parallel()

	// RSA key pair has public key component.
	kid := googleUuid.New()
	enc := joseJwa.A256GCM()
	alg := joseJwa.RSA_OAEP_256()
	keyPair, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	_, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWEJWKFromKey(&kid, &enc, &alg, keyPair)
	require.NoError(t, err)
	require.NotNil(t, nonPublicJWK)
	require.NotNil(t, publicJWK) // RSA should have public key
	require.NotEmpty(t, publicBytes)
	require.NotEmpty(t, nonPublicBytes)

	// Verify KeyType is RSA
	require.Equal(t, joseJwa.RSA().String(), nonPublicJWK.KeyType().String())
}

func TestCreateJWEJWKFromKey_UnexpectedKeyPairPrivateType(t *testing.T) {
	t.Parallel()

	// Use KeyPair with unexpected private key type (string).
	kid := googleUuid.New()
	enc := joseJwa.A256GCM()
	alg := joseJwa.A256GCMKW()
	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: "not-a-private-key",
		Public:  "not-a-public-key",
	}

	_, _, _, _, _, err := CreateJWEJWKFromKey(&kid, &enc, &alg, keyPair)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported key type *keygen.KeyPair")
}

func TestCreateJWEJWKFromKey_NilKid(t *testing.T) {
	t.Parallel()

	enc := joseJwa.A256GCM()
	alg := joseJwa.DIRECT()
	key := make(cryptoutilSharedCryptoKeygen.SecretKey, cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes)
	_, _ = crand.Read(key)

	_, _, _, _, _, err := CreateJWEJWKFromKey(nil, &enc, &alg, key)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWE JWK headers")
}

func TestCreateJWEJWKFromKey_NilAlg(t *testing.T) {
	t.Parallel()

	kid := googleUuid.New()
	enc := joseJwa.A256GCM()
	key := make(cryptoutilSharedCryptoKeygen.SecretKey, cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes)
	_, _ = crand.Read(key)

	_, _, _, _, _, err := CreateJWEJWKFromKey(&kid, &enc, nil, key)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWE JWK headers")
}

func TestCreateJWEJWKFromKey_NilEnc(t *testing.T) {
	t.Parallel()

	kid := googleUuid.New()
	alg := joseJwa.DIRECT()
	key := make(cryptoutilSharedCryptoKeygen.SecretKey, cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes)
	_, _ = crand.Read(key)

	_, _, _, _, _, err := CreateJWEJWKFromKey(&kid, nil, &alg, key)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWE JWK headers")
}

func TestCreateJWEJWKFromKey_NilKey(t *testing.T) {
	t.Parallel()

	kid := googleUuid.New()
	enc := joseJwa.A256GCM()
	alg := joseJwa.DIRECT()

	_, _, _, _, _, err := CreateJWEJWKFromKey(&kid, &enc, &alg, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWE JWK headers")
}

// TestCreateJWEJWKFromKey_EmptySecretKey tests validation error for empty AES key.
func TestCreateJWEJWKFromKey_EmptySecretKey(t *testing.T) {
	t.Parallel()

	kid := googleUuid.New()
	enc := joseJwa.A256GCM()
	alg := joseJwa.DIRECT()
	emptyKey := cryptoutilSharedCryptoKeygen.SecretKey("")

	_, _, _, _, _, err := CreateJWEJWKFromKey(&kid, &enc, &alg, emptyKey)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWE JWK headers")
}

// TestCreateJWEJWKFromKey_NilPrivateKeyPair tests validation error for nil KeyPair.
func TestCreateJWEJWKFromKey_NilPrivateKeyPair(t *testing.T) {
	t.Parallel()

	kid := googleUuid.New()
	enc := joseJwa.A256GCM()
	alg := joseJwa.ECDH_ES()
	invalidKeyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: nil,
		Public:  nil,
	}

	_, _, _, _, _, err := CreateJWEJWKFromKey(&kid, &enc, &alg, invalidKeyPair)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWE JWK headers")
}

func TestCreateJWEJWKFromKey_InvalidEnc(t *testing.T) {
	t.Parallel()

	kid := googleUuid.New()
	// Create ContentEncryptionAlgorithm not in EncToBitsLength switch.
	invalidEnc := joseJwa.NewContentEncryptionAlgorithm("INVALID_ENC")
	alg := joseJwa.A256KW()
	secretKey, err := cryptoutilSharedCryptoKeygen.GenerateAESKey(cryptoutilSharedMagic.MaxUnsealSharedSecrets)
	require.NoError(t, err)

	_, _, _, _, _, err = CreateJWEJWKFromKey(&kid, &invalidEnc, &alg, secretKey)
	require.Error(t, err)
	require.Contains(t, err.Error(), "JWE JWK length error")
}

func TestCreateJWEJWKFromKey_UnsupportedAlg(t *testing.T) {
	t.Parallel()

	kid := googleUuid.New()
	enc := joseJwa.A256GCM()
	// Create unsupported KeyEncryptionAlgorithm.
	invalidAlg := joseJwa.NewKeyEncryptionAlgorithm("UNSUPPORTED_ALG")
	secretKey, err := cryptoutilSharedCryptoKeygen.GenerateAESKey(cryptoutilSharedMagic.MaxUnsealSharedSecrets)
	require.NoError(t, err)

	_, _, _, _, _, err = CreateJWEJWKFromKey(&kid, &enc, &invalidAlg, secretKey)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported JWE JWK alg")
}

// TestCreateJWEJWKFromKey_RSA_AllAlgorithms tests RSA with all key encryption algorithms.
