// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
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

func TestValidateOrGenerateJWEAESJWK_Generate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		enc           *joseJwa.ContentEncryptionAlgorithm
		alg           *joseJwa.KeyEncryptionAlgorithm
		keyBitsLength int
		allowedEncs   []*joseJwa.ContentEncryptionAlgorithm
		expectError   bool
	}{
		{
			name:          "A256KW with A256GCM",
			enc:           &EncA256GCM,
			alg:           &AlgA256KW,
			keyBitsLength: 256,
			allowedEncs:   []*joseJwa.ContentEncryptionAlgorithm{&EncA256GCM},
			expectError:   false,
		},
		{
			name:          "A192KW with A192GCM",
			enc:           &EncA192GCM,
			alg:           &AlgA192KW,
			keyBitsLength: 192,
			allowedEncs:   []*joseJwa.ContentEncryptionAlgorithm{&EncA192GCM},
			expectError:   false,
		},
		{
			name:          "A128KW with A128GCM",
			enc:           &EncA128GCM,
			alg:           &AlgA128KW,
			keyBitsLength: 128,
			allowedEncs:   []*joseJwa.ContentEncryptionAlgorithm{&EncA128GCM},
			expectError:   false,
		},
		{
			name:          "dir with A256GCM",
			enc:           &EncA256GCM,
			alg:           &AlgDir,
			keyBitsLength: 256,
			allowedEncs:   []*joseJwa.ContentEncryptionAlgorithm{&EncA256GCM},
			expectError:   false,
		},
		{
			name:          "dir with A256CBC-HS512",
			enc:           &EncA256CBCHS512,
			alg:           &AlgDir,
			keyBitsLength: 512,
			allowedEncs:   []*joseJwa.ContentEncryptionAlgorithm{&EncA256CBCHS512},
			expectError:   false,
		},
		{
			name:          "disallowed enc",
			enc:           &EncA128GCM,
			alg:           &AlgA256KW,
			keyBitsLength: 256,
			allowedEncs:   []*joseJwa.ContentEncryptionAlgorithm{&EncA256GCM},
			expectError:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			keyBytes, err := validateOrGenerateJWEAESJWK(nil, tc.enc, tc.alg, tc.keyBitsLength, tc.allowedEncs...)

			if tc.expectError {
				require.Error(t, err)
				require.Nil(t, keyBytes)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, keyBytes)
			require.Equal(t, tc.keyBitsLength/8, len(keyBytes))
		})
	}
}

func TestValidateOrGenerateJWEAESJWK_Validate(t *testing.T) {
	t.Parallel()

	validKey256, err := cryptoutilSharedCryptoKeygen.GenerateAESKey(256)
	require.NoError(t, err)

	validKey128, err := cryptoutilSharedCryptoKeygen.GenerateAESKey(128)
	require.NoError(t, err)

	tests := []struct {
		name        string
		key         cryptoutilSharedCryptoKeygen.Key
		enc         *joseJwa.ContentEncryptionAlgorithm
		alg         *joseJwa.KeyEncryptionAlgorithm
		keyBitsLen  int
		allowedEncs []*joseJwa.ContentEncryptionAlgorithm
		expectError bool
	}{
		{
			name:        "valid 256-bit key",
			key:         validKey256,
			enc:         &EncA256GCM,
			alg:         &AlgA256KW,
			keyBitsLen:  256,
			allowedEncs: []*joseJwa.ContentEncryptionAlgorithm{&EncA256GCM},
			expectError: false,
		},
		{
			name:        "wrong key length",
			key:         validKey128,
			enc:         &EncA256GCM,
			alg:         &AlgA256KW,
			keyBitsLen:  256,
			allowedEncs: []*joseJwa.ContentEncryptionAlgorithm{&EncA256GCM},
			expectError: true,
		},
		{
			name:        "nil key bytes",
			key:         cryptoutilSharedCryptoKeygen.SecretKey(nil),
			enc:         &EncA256GCM,
			alg:         &AlgA256KW,
			keyBitsLen:  256,
			allowedEncs: []*joseJwa.ContentEncryptionAlgorithm{&EncA256GCM},
			expectError: true,
		},
		{
			name:        "wrong key type",
			key:         &cryptoutilSharedCryptoKeygen.KeyPair{},
			enc:         &EncA256GCM,
			alg:         &AlgA256KW,
			keyBitsLen:  256,
			allowedEncs: []*joseJwa.ContentEncryptionAlgorithm{&EncA256GCM},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			keyBytes, err := validateOrGenerateJWEAESJWK(tc.key, tc.enc, tc.alg, tc.keyBitsLen, tc.allowedEncs...)

			if tc.expectError {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, keyBytes)
		})
	}
}

func TestValidateOrGenerateJWEAESJWK_InvalidEncWithDir(t *testing.T) {
	t.Parallel()

	// AlgDir with invalid enc (not GCM or CBC-HS) should trigger default case.
	invalidEnc := joseJwa.NewContentEncryptionAlgorithm("A256XYZ")
	alg := joseJwa.DIRECT()
	allowedEncs := []*joseJwa.ContentEncryptionAlgorithm{&invalidEnc}

	keyBytes, err := validateOrGenerateJWEAESJWK(nil, &invalidEnc, &alg, 256, allowedEncs...)
	require.Error(t, err)
	require.Nil(t, keyBytes)
	require.Contains(t, err.Error(), "valid JWE JWK alg")
	require.Contains(t, err.Error(), "but invalid enc")
}

func TestValidateOrGenerateJWEAESJWK_WrongKeyType(t *testing.T) {
	t.Parallel()

	// Pass KeyPair (RSA) instead of SecretKey ([]byte).
	rsaKey, err := rsa.GenerateKey(crand.Reader, 2048)
	require.NoError(t, err)

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: rsaKey,
		Public:  &rsaKey.PublicKey,
	}

	result, err := validateOrGenerateJWEAESJWK(keyPair, &EncA256GCM, &AlgA256KW, 256, &EncA256GCM)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "unsupported key type *keygen.KeyPair")
}

func TestValidateOrGenerateJWEAESJWK_DisallowedEnc(t *testing.T) {
	t.Parallel()

	// Test enc not in allowedEncs list.
	secretKey, err := cryptoutilSharedCryptoKeygen.GenerateAESKey(256)
	require.NoError(t, err)

	// Use A128GCM but only allow A256GCM.
	result, err := validateOrGenerateJWEAESJWK(secretKey, &EncA128GCM, &AlgA256KW, 256, &EncA256GCM)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "enc A128GCM not allowed")
}

func TestValidateOrGenerateJWEAESJWK_NilKey(t *testing.T) {
	t.Parallel()

	// Test nil key bytes.
	var secretKey cryptoutilSharedCryptoKeygen.SecretKey

	result, err := validateOrGenerateJWEAESJWK(secretKey, &EncA256GCM, &AlgA256KW, 256, &EncA256GCM)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "invalid nil key bytes")
}

func TestValidateOrGenerateJWEAESJWK_WrongLength(t *testing.T) {
	t.Parallel()

	// Test wrong length key (128 bits instead of 256 bits).
	secretKey, err := cryptoutilSharedCryptoKeygen.GenerateAESKey(128)
	require.NoError(t, err)

	result, err := validateOrGenerateJWEAESJWK(secretKey, &EncA256GCM, &AlgA256KW, 256, &EncA256GCM)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "invalid key length")
}

func TestValidateOrGenerateJWERSAJWK_Generate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		enc           *joseJwa.ContentEncryptionAlgorithm
		alg           *joseJwa.KeyEncryptionAlgorithm
		keyBitsLength int
		allowedEncs   []*joseJwa.ContentEncryptionAlgorithm
		expectError   bool
	}{
		{
			name:          "RSA-OAEP with A256GCM",
			enc:           &EncA256GCM,
			alg:           &AlgRSAOAEP,
			keyBitsLength: 2048,
			allowedEncs:   []*joseJwa.ContentEncryptionAlgorithm{&EncA256GCM},
			expectError:   false,
		},
		{
			name:          "RSA-OAEP-256 with A192GCM",
			enc:           &EncA192GCM,
			alg:           &AlgRSAOAEP256,
			keyBitsLength: 3072,
			allowedEncs:   []*joseJwa.ContentEncryptionAlgorithm{&EncA192GCM},
			expectError:   false,
		},
		{
			name:          "disallowed enc",
			enc:           &EncA128GCM,
			alg:           &AlgRSAOAEP,
			keyBitsLength: 2048,
			allowedEncs:   []*joseJwa.ContentEncryptionAlgorithm{&EncA256GCM},
			expectError:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			keyPair, err := validateOrGenerateJWERSAJWK(nil, tc.enc, tc.alg, tc.keyBitsLength, tc.allowedEncs...)

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

func TestValidateOrGenerateJWEEcdhJWK_ValidateExistingKey(t *testing.T) {
	t.Parallel()

	// Generate valid ECDH P-256 key pair.
	validKeyPair, err := cryptoutilSharedCryptoKeygen.GenerateECDHKeyPair(ecdh.P256())
	require.NoError(t, err)

	// Test validation with valid key.
	validated, err := validateOrGenerateJWEEcdhJWK(validKeyPair, &EncA256GCM, &AlgECDHES, ecdh.P256(), &EncA256GCM)
	require.NoError(t, err)
	require.Equal(t, validKeyPair, validated)
}

func TestValidateOrGenerateJWEEcdhJWK_WrongKeyType(t *testing.T) {
	t.Parallel()

	// Use symmetric key (wrong type for ECDH).
	symmetricKey := cryptoutilSharedCryptoKeygen.SecretKey(make([]byte, 32))

	validated, err := validateOrGenerateJWEEcdhJWK(symmetricKey, &EncA256GCM, &AlgECDHES, ecdh.P256(), &EncA256GCM)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "unsupported key type")
}

func TestValidateOrGenerateJWERSAJWK_ValidateExistingKey(t *testing.T) {
	t.Parallel()

	// Generate valid RSA key pair.
	validKeyPair, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	// Test validation with valid key.
	validated, err := validateOrGenerateJWERSAJWK(validKeyPair, &EncA256GCM, &AlgRSAOAEP, 2048, &EncA256GCM)
	require.NoError(t, err)
	require.Equal(t, validKeyPair, validated)
}

func TestValidateOrGenerateJWERSAJWK_WrongKeyType(t *testing.T) {
	t.Parallel()

	// Use symmetric key (wrong type for RSA).
	symmetricKey := cryptoutilSharedCryptoKeygen.SecretKey(make([]byte, 32))

	validated, err := validateOrGenerateJWERSAJWK(symmetricKey, &EncA256GCM, &AlgRSAOAEP, 2048, &EncA256GCM)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "unsupported key type")
}

func TestValidateOrGenerateJWERSAJWK_NilPrivateKey(t *testing.T) {
	t.Parallel()

	// Generate RSA key then set private to nil.
	rsaKey, err := rsa.GenerateKey(crand.Reader, 2048)
	require.NoError(t, err)

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: (*rsa.PrivateKey)(nil), // Typed nil to pass type check
		Public:  &rsaKey.PublicKey,
	}

	result, err := validateOrGenerateJWERSAJWK(keyPair, &EncA256GCM, &AlgRSAOAEP, 2048, &EncA256GCM)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "invalid nil RSA private key")
}

func TestValidateOrGenerateJWERSAJWK_NilPublicKey(t *testing.T) {
	t.Parallel()

	// Generate RSA private key, create KeyPair with typed nil public.
	rsaKey, err := rsa.GenerateKey(crand.Reader, 2048)
	require.NoError(t, err)

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: rsaKey,
		Public:  (*rsa.PublicKey)(nil), // Typed nil to pass type check
	}

	result, err := validateOrGenerateJWERSAJWK(keyPair, &EncA256GCM, &AlgRSAOAEP, 2048, &EncA256GCM)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "invalid nil RSA public key")
}

func TestValidateOrGenerateJWERSAJWK_WrongPrivateKeyType(t *testing.T) {
	t.Parallel()

	// Create KeyPair with ECDSA private key instead of RSA.
	ecdsaKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: ecdsaKey,
		Public:  &ecdsaKey.PublicKey,
	}

	result, err := validateOrGenerateJWERSAJWK(keyPair, &EncA256GCM, &AlgRSAOAEP, 2048, &EncA256GCM)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "unsupported key type *ecdsa.PrivateKey")
}

func TestValidateOrGenerateJWERSAJWK_WrongPublicKeyType(t *testing.T) {
	t.Parallel()

	// Generate RSA private key, pair with ECDSA public key (invalid).
	rsaKey, err := rsa.GenerateKey(crand.Reader, 2048)
	require.NoError(t, err)

	ecdsaKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: rsaKey,
		Public:  &ecdsaKey.PublicKey,
	}

	result, err := validateOrGenerateJWERSAJWK(keyPair, &EncA256GCM, &AlgRSAOAEP, 2048, &EncA256GCM)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "unsupported key type *ecdsa.PublicKey")
}

func TestValidateOrGenerateJWERSAJWK_DisallowedEnc(t *testing.T) {
	t.Parallel()

	// Test enc not in allowedEncs list.
	rsaKey, err := rsa.GenerateKey(crand.Reader, 2048)
	require.NoError(t, err)

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: rsaKey,
		Public:  &rsaKey.PublicKey,
	}

	// Use A128GCM but only allow A256GCM.
	result, err := validateOrGenerateJWERSAJWK(keyPair, &EncA128GCM, &AlgRSAOAEP, 2048, &EncA256GCM)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "enc A128GCM not allowed")
}

func TestValidateOrGenerateJWEEcdhJWK_WrongPrivateKeyType(t *testing.T) {
	t.Parallel()

	// Generate RSA key instead of ECDH.
	rsaKey, err := rsa.GenerateKey(crand.Reader, 2048)
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
	key := make(cryptoutilSharedCryptoKeygen.SecretKey, 32)
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
	keyPair, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(2048)
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
	key := make(cryptoutilSharedCryptoKeygen.SecretKey, 32)
	_, _ = crand.Read(key)

	_, _, _, _, _, err := CreateJWEJWKFromKey(nil, &enc, &alg, key)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWE JWK headers")
}

func TestCreateJWEJWKFromKey_NilAlg(t *testing.T) {
	t.Parallel()

	kid := googleUuid.New()
	enc := joseJwa.A256GCM()
	key := make(cryptoutilSharedCryptoKeygen.SecretKey, 32)
	_, _ = crand.Read(key)

	_, _, _, _, _, err := CreateJWEJWKFromKey(&kid, &enc, nil, key)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWE JWK headers")
}

func TestCreateJWEJWKFromKey_NilEnc(t *testing.T) {
	t.Parallel()

	kid := googleUuid.New()
	alg := joseJwa.DIRECT()
	key := make(cryptoutilSharedCryptoKeygen.SecretKey, 32)
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
	secretKey, err := cryptoutilSharedCryptoKeygen.GenerateAESKey(256)
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
	secretKey, err := cryptoutilSharedCryptoKeygen.GenerateAESKey(256)
	require.NoError(t, err)

	_, _, _, _, _, err = CreateJWEJWKFromKey(&kid, &enc, &invalidAlg, secretKey)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported JWE JWK alg")
}

// TestCreateJWEJWKFromKey_RSA_AllAlgorithms tests RSA with all key encryption algorithms.
func TestCreateJWEJWKFromKey_RSA_AllAlgorithms(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		enc     joseJwa.ContentEncryptionAlgorithm
		alg     joseJwa.KeyEncryptionAlgorithm
		keySize int
	}{
		{"RSA_OAEP_256_A256GCM", joseJwa.A256GCM(), joseJwa.RSA_OAEP_256(), 2048},
		{"RSA_OAEP_384_A256GCM", joseJwa.A256GCM(), joseJwa.RSA_OAEP_384(), 3072},
		{"RSA_OAEP_512_A256GCM", joseJwa.A256GCM(), joseJwa.RSA_OAEP_512(), 4096},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			kid := googleUuid.New()
			keyPair, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(tt.keySize)
			require.NoError(t, err)

			resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWEJWKFromKey(&kid, &tt.enc, &tt.alg, keyPair)
			require.NoError(t, err)
			require.Equal(t, &kid, resultKid)
			require.NotNil(t, nonPublicJWK)
			require.NotNil(t, publicJWK)
			require.NotEmpty(t, nonPublicBytes)
			require.NotEmpty(t, publicBytes)

			require.Equal(t, KtyRSA, nonPublicJWK.KeyType())
		})
	}
}

// TestCreateJWEJWKFromKey_ECDH_AllCurves tests ECDH with all curves.
func TestCreateJWEJWKFromKey_ECDH_AllCurves(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		enc   joseJwa.ContentEncryptionAlgorithm
		alg   joseJwa.KeyEncryptionAlgorithm
		curve ecdh.Curve
	}{
		{"ECDH_ES_P256", joseJwa.A256GCM(), joseJwa.ECDH_ES(), ecdh.P256()},
		{"ECDH_ES_P384", joseJwa.A256GCM(), joseJwa.ECDH_ES(), ecdh.P384()},
		{"ECDH_ES_P521", joseJwa.A256GCM(), joseJwa.ECDH_ES(), ecdh.P521()},
		{"ECDH_ES_A256KW_P256", joseJwa.A256GCM(), joseJwa.ECDH_ES_A256KW(), ecdh.P256()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			kid := googleUuid.New()
			keyPair, err := cryptoutilSharedCryptoKeygen.GenerateECDHKeyPair(tt.curve)
			require.NoError(t, err)

			resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWEJWKFromKey(&kid, &tt.enc, &tt.alg, keyPair)
			require.NoError(t, err)
			require.Equal(t, &kid, resultKid)
			require.NotNil(t, nonPublicJWK)
			require.NotNil(t, publicJWK)
			require.NotEmpty(t, nonPublicBytes)
			require.NotEmpty(t, publicBytes)

			require.Equal(t, KtyEC, nonPublicJWK.KeyType())
		})
	}
}

// TestCreateJWEJWKFromKey_AES_AllSizes tests AES secret key with all sizes.
func TestCreateJWEJWKFromKey_AES_AllSizes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		enc     joseJwa.ContentEncryptionAlgorithm
		alg     joseJwa.KeyEncryptionAlgorithm
		keySize int
	}{
		{"A128KW_A128GCM", joseJwa.A128GCM(), joseJwa.A128KW(), 128},
		{"A192KW_A192GCM", joseJwa.A192GCM(), joseJwa.A192KW(), 192},
		{"A256KW_A256GCM", joseJwa.A256GCM(), joseJwa.A256KW(), 256},
		{"dir_A256GCM", joseJwa.A256GCM(), joseJwa.DIRECT(), 256},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			kid := googleUuid.New()
			secretKey, err := cryptoutilSharedCryptoKeygen.GenerateAESKey(tt.keySize)
			require.NoError(t, err)

			resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWEJWKFromKey(&kid, &tt.enc, &tt.alg, secretKey)
			require.NoError(t, err)
			require.Equal(t, &kid, resultKid)
			require.NotNil(t, nonPublicJWK)
			require.Nil(t, publicJWK)
			require.NotEmpty(t, nonPublicBytes)
			require.Empty(t, publicBytes)

			require.Equal(t, KtyOCT, nonPublicJWK.KeyType())
		})
	}
}

// TestCreateJWEJWKFromKey_ErrorCases tests comprehensive error handling.
func TestCreateJWEJWKFromKey_ErrorCases(t *testing.T) {
	t.Parallel()

	t.Run("NilKid", func(t *testing.T) {
		t.Parallel()

		enc := joseJwa.A256GCM()
		alg := joseJwa.A256KW()
		key, err := cryptoutilSharedCryptoKeygen.GenerateAESKey(256)
		require.NoError(t, err)

		resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWEJWKFromKey(nil, &enc, &alg, key)
		require.Error(t, err)
		require.Nil(t, resultKid)
		require.Nil(t, nonPublicJWK)
		require.Nil(t, publicJWK)
		require.Empty(t, nonPublicBytes)
		require.Empty(t, publicBytes)
		require.Contains(t, err.Error(), "JWE JWK kid must be valid")
	})

	t.Run("NilEnc", func(t *testing.T) {
		t.Parallel()

		kid := googleUuid.New()
		alg := joseJwa.A256KW()
		key, err := cryptoutilSharedCryptoKeygen.GenerateAESKey(256)
		require.NoError(t, err)

		resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWEJWKFromKey(&kid, nil, &alg, key)
		require.Error(t, err)
		require.Nil(t, resultKid)
		require.Nil(t, nonPublicJWK)
		require.Nil(t, publicJWK)
		require.Empty(t, nonPublicBytes)
		require.Empty(t, publicBytes)
		require.Contains(t, err.Error(), "JWE JWK enc must be non-nil")
	})

	t.Run("NilAlg", func(t *testing.T) {
		t.Parallel()

		kid := googleUuid.New()
		enc := joseJwa.A256GCM()
		key, err := cryptoutilSharedCryptoKeygen.GenerateAESKey(256)
		require.NoError(t, err)

		resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWEJWKFromKey(&kid, &enc, nil, key)
		require.Error(t, err)
		require.Nil(t, resultKid)
		require.Nil(t, nonPublicJWK)
		require.Nil(t, publicJWK)
		require.Empty(t, nonPublicBytes)
		require.Empty(t, publicBytes)
		require.Contains(t, err.Error(), "JWE JWK alg must be non-nil")
	})

	t.Run("NilKey", func(t *testing.T) {
		t.Parallel()

		kid := googleUuid.New()
		enc := joseJwa.A256GCM()
		alg := joseJwa.A256KW()

		resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWEJWKFromKey(&kid, &enc, &alg, nil)
		require.Error(t, err)
		require.Nil(t, resultKid)
		require.Nil(t, nonPublicJWK)
		require.Nil(t, publicJWK)
		require.Empty(t, nonPublicBytes)
		require.Empty(t, publicBytes)
		require.Contains(t, err.Error(), "JWE JWK key must be non-nil")
	})
}
