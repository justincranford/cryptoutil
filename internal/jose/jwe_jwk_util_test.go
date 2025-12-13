// Copyright (c) 2025 Justin Cranford
//
//

package jose

import (
	"crypto/ecdh"
	crand "crypto/rand"
	"crypto/rsa"
	"testing"

	cryptoutilKeyGen "cryptoutil/internal/common/crypto/keygen"

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

	validKey256, err := cryptoutilKeyGen.GenerateAESKey(256)
	require.NoError(t, err)

	validKey128, err := cryptoutilKeyGen.GenerateAESKey(128)
	require.NoError(t, err)

	tests := []struct {
		name        string
		key         cryptoutilKeyGen.Key
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
			key:         cryptoutilKeyGen.SecretKey(nil),
			enc:         &EncA256GCM,
			alg:         &AlgA256KW,
			keyBitsLen:  256,
			allowedEncs: []*joseJwa.ContentEncryptionAlgorithm{&EncA256GCM},
			expectError: true,
		},
		{
			name:        "wrong key type",
			key:         &cryptoutilKeyGen.KeyPair{},
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
	validKeyPair, err := cryptoutilKeyGen.GenerateECDHKeyPair(ecdh.P256())
	require.NoError(t, err)

	// Test validation with valid key.
	validated, err := validateOrGenerateJWEEcdhJWK(validKeyPair, &EncA256GCM, &AlgECDHES, ecdh.P256(), &EncA256GCM)
	require.NoError(t, err)
	require.Equal(t, validKeyPair, validated)
}

func TestValidateOrGenerateJWEEcdhJWK_WrongKeyType(t *testing.T) {
	t.Parallel()

	// Use symmetric key (wrong type for ECDH).
	symmetricKey := cryptoutilKeyGen.SecretKey(make([]byte, 32))

	validated, err := validateOrGenerateJWEEcdhJWK(symmetricKey, &EncA256GCM, &AlgECDHES, ecdh.P256(), &EncA256GCM)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "unsupported key type")
}

func TestValidateOrGenerateJWERSAJWK_ValidateExistingKey(t *testing.T) {
	t.Parallel()

	// Generate valid RSA key pair.
	validKeyPair, err := cryptoutilKeyGen.GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	// Test validation with valid key.
	validated, err := validateOrGenerateJWERSAJWK(validKeyPair, &EncA256GCM, &AlgRSAOAEP, 2048, &EncA256GCM)
	require.NoError(t, err)
	require.Equal(t, validKeyPair, validated)
}

func TestValidateOrGenerateJWERSAJWK_WrongKeyType(t *testing.T) {
	t.Parallel()

	// Use symmetric key (wrong type for RSA).
	symmetricKey := cryptoutilKeyGen.SecretKey(make([]byte, 32))

	validated, err := validateOrGenerateJWERSAJWK(symmetricKey, &EncA256GCM, &AlgRSAOAEP, 2048, &EncA256GCM)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "unsupported key type")
}

func TestValidateOrGenerateJWERSAJWK_NilPrivateKey(t *testing.T) {
	t.Parallel()

	// KeyPair with nil private key.
	keyPair := &cryptoutilKeyGen.KeyPair{
		Private: nil,
		Public:  &rsa.PublicKey{},
	}

	result, err := validateOrGenerateJWERSAJWK(keyPair, &EncA256GCM, &AlgRSAOAEP, 2048, &EncA256GCM)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "unsupported key type")
}

func TestValidateOrGenerateJWERSAJWK_NilPublicKey(t *testing.T) {
	t.Parallel()

	// Generate RSA private key, create KeyPair with nil public.
	rsaKey, err := rsa.GenerateKey(crand.Reader, 2048)
	require.NoError(t, err)

	keyPair := &cryptoutilKeyGen.KeyPair{
		Private: rsaKey,
		Public:  nil,
	}

	result, err := validateOrGenerateJWERSAJWK(keyPair, &EncA256GCM, &AlgRSAOAEP, 2048, &EncA256GCM)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "unsupported key type")
}

func TestValidateOrGenerateJWEEcdhJWK_NilPrivateKey(t *testing.T) {
	t.Parallel()

	// KeyPair with nil private key.
	keyPair := &cryptoutilKeyGen.KeyPair{
		Private: nil,
		Public:  &ecdh.PublicKey{},
	}

	result, err := validateOrGenerateJWEEcdhJWK(keyPair, &EncA256GCM, &AlgECDHES, ecdh.P256(), &EncA256GCM)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "unsupported key type")
}

func TestValidateOrGenerateJWEEcdhJWK_NilPublicKey(t *testing.T) {
	t.Parallel()

	// Generate ECDH private key, create KeyPair with nil public.
	ecdhPriv, err := ecdh.P256().GenerateKey(crand.Reader)
	require.NoError(t, err)

	keyPair := &cryptoutilKeyGen.KeyPair{
		Private: ecdhPriv,
		Public:  nil,
	}

	result, err := validateOrGenerateJWEEcdhJWK(keyPair, &EncA256GCM, &AlgECDHES, ecdh.P256(), &EncA256GCM)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "unsupported key type")
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
	key := make(cryptoutilKeyGen.SecretKey, 32)
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
	keyPair, err := cryptoutilKeyGen.GenerateECDHKeyPair(ecdh.P256())
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
	keyPair, err := cryptoutilKeyGen.GenerateRSAKeyPair(2048)
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



func TestCreateJWEJWKFromKey_NilKid(t *testing.T) {
	t.Parallel()

	enc := joseJwa.A256GCM()
	alg := joseJwa.DIRECT()
	key := make(cryptoutilKeyGen.SecretKey, 32)
	_, _ = crand.Read(key)

	_, _, _, _, _, err := CreateJWEJWKFromKey(nil, &enc, &alg, key)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWE JWK headers")
}

func TestCreateJWEJWKFromKey_NilAlg(t *testing.T) {
	t.Parallel()

	kid := googleUuid.New()
	enc := joseJwa.A256GCM()
	key := make(cryptoutilKeyGen.SecretKey, 32)
	_, _ = crand.Read(key)

	_, _, _, _, _, err := CreateJWEJWKFromKey(&kid, &enc, nil, key)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWE JWK headers")
}

func TestCreateJWEJWKFromKey_NilEnc(t *testing.T) {
	t.Parallel()

	kid := googleUuid.New()
	alg := joseJwa.DIRECT()
	key := make(cryptoutilKeyGen.SecretKey, 32)
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
	emptyKey := cryptoutilKeyGen.SecretKey("")

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
	invalidKeyPair := &cryptoutilKeyGen.KeyPair{
		Private: nil,
		Public:  nil,
	}

	_, _, _, _, _, err := CreateJWEJWKFromKey(&kid, &enc, &alg, invalidKeyPair)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWE JWK headers")
}
