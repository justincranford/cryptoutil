// Copyright (c) 2025 Justin Cranford
//
//

package jose

import (
	"crypto/ecdh"
	"testing"

	cryptoutilKeyGen "cryptoutil/internal/common/crypto/keygen"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/stretchr/testify/require"
)

func TestValidateOrGenerateJWEAESJWK_Generate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		enc            *joseJwa.ContentEncryptionAlgorithm
		alg            *joseJwa.KeyEncryptionAlgorithm
		keyBitsLength  int
		allowedEncs    []*joseJwa.ContentEncryptionAlgorithm
		expectError    bool
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
