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
