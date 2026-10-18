// Copyright (c) 2025-2026 Justin Cranford.
package issuer

import (
	"context"
	"encoding/base64"
	"testing"

	testify "github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	failedToCreateAESCipherError        = "failed to create AES cipher"
	failedToGetActiveEncryptionKeyError = "failed to get active encryption key"
)

// TestEncryptDecryptToken_RoundTrip tests encrypt->decrypt round trip.
func TestEncryptDecryptToken_RoundTrip(t *testing.T) {
	t.Parallel()

	gen := NewProductionKeyGenerator()

	mgr, err := NewKeyRotationManager(DefaultKeyRotationPolicy(), gen, nil)
	testify.NoError(t, err)

	err = mgr.RotateEncryptionKey(context.Background())
	testify.NoError(t, err)

	jweIssuer := &JWEIssuer{
		keyRotationMgr: mgr,
	}

	ctx := context.Background()

	plaintext := "test-plaintext-token-data"

	encrypted, err := jweIssuer.EncryptToken(ctx, plaintext)
	testify.NoError(t, err)
	testify.NotEmpty(t, encrypted)

	decrypted, err := jweIssuer.DecryptToken(ctx, encrypted)
	testify.NoError(t, err)
	testify.Equal(t, plaintext, decrypted)
}

// TestEncryptToken_NoEncryptionKey tests encryption when no key is available.
func TestEncryptToken_NoEncryptionKey(t *testing.T) {
	t.Parallel()

	jweIssuer := &JWEIssuer{}

	ctx := context.Background()

	encrypted, err := jweIssuer.EncryptToken(ctx, "test-plaintext")
	testify.Error(t, err)
	testify.Empty(t, encrypted)
}

// TestDecryptToken_InvalidFormat tests decryption with invalid input formats.
func TestDecryptToken_InvalidFormat(t *testing.T) {
	t.Parallel()

	gen := NewProductionKeyGenerator()

	mgr, err := NewKeyRotationManager(DefaultKeyRotationPolicy(), gen, nil)
	testify.NoError(t, err)

	jweIssuer := &JWEIssuer{
		keyRotationMgr:      mgr,
		legacyEncryptionKey: []byte("this-is-not-a-valid-aes-key"),
	}

	ctx := context.Background()

	tests := []struct {
		name    string
		token   string
		wantErr string
	}{
		{
			name:    "invalid base64 input",
			token:   "===invalid===",
			wantErr: "failed to decode base64",
		},
		{
			name:    "ciphertext too short",
			token:   base64.RawURLEncoding.EncodeToString([]byte("ab")),
			wantErr: failedToCreateAESCipherError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			decrypted, err := jweIssuer.DecryptToken(ctx, tc.token)
			testify.Error(t, err)
			testify.Empty(t, decrypted)
			testify.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestJWEIssuer_KeyAndCipherErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	generator := NewProductionKeyGenerator()
	manager, err := NewKeyRotationManager(DefaultKeyRotationPolicy(), generator, nil)
	testify.NoError(t, err)

	tests := []struct {
		name    string
		issuer  *JWEIssuer
		token   string
		encrypt bool
		wantErr string
	}{
		{
			name:    "encrypt without active key",
			issuer:  &JWEIssuer{keyRotationMgr: manager},
			encrypt: true,
			wantErr: failedToGetActiveEncryptionKeyError,
		},
		{
			name:    "decrypt with unknown key id",
			issuer:  &JWEIssuer{keyRotationMgr: manager},
			token:   base64.RawURLEncoding.EncodeToString([]byte{0, 7, 'm', 'i', 's', 's', 'i', 'n', 'g', 0}),
			wantErr: "failed to get encryption key",
		},
		{
			name:    "decrypt without active key",
			issuer:  &JWEIssuer{keyRotationMgr: manager},
			token:   base64.RawURLEncoding.EncodeToString([]byte{0, 0, 1}),
			wantErr: failedToGetActiveEncryptionKeyError,
		},
		{
			name:    "encrypt with invalid legacy key",
			issuer:  &JWEIssuer{legacyEncryptionKey: []byte{}},
			encrypt: true,
			wantErr: failedToCreateAESCipherError,
		},
		{
			name:    "decrypt with invalid legacy key",
			issuer:  &JWEIssuer{legacyEncryptionKey: []byte{}},
			token:   base64.RawURLEncoding.EncodeToString([]byte("ciphertext")),
			wantErr: failedToCreateAESCipherError,
		},
		{
			name:    "decrypt truncated ciphertext",
			issuer:  &JWEIssuer{legacyEncryptionKey: make([]byte, cryptoutilSharedMagic.AES256KeySize)},
			token:   base64.RawURLEncoding.EncodeToString([]byte("short")),
			wantErr: cryptoutilSharedMagic.ErrorInvalidToken,
		},
		{
			name:    "decrypt invalid GCM tag",
			issuer:  &JWEIssuer{legacyEncryptionKey: make([]byte, cryptoutilSharedMagic.AES256KeySize)},
			token:   base64.RawURLEncoding.EncodeToString(make([]byte, 32)),
			wantErr: "failed to decrypt token",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var operationErr error
			if tc.encrypt {
				_, operationErr = tc.issuer.EncryptToken(ctx, "plaintext")
			} else {
				_, operationErr = tc.issuer.DecryptToken(ctx, tc.token)
			}

			testify.ErrorContains(t, operationErr, tc.wantErr)
		})
	}
}
