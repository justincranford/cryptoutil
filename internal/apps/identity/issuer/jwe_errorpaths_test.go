// Copyright (c) 2025 Justin Cranford

package issuer

import (
	"context"
	"encoding/base64"
	"testing"

	testify "github.com/stretchr/testify/require"
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
		legacyEncryptionKey: []byte("this-is-not-a-valid-aes-keyecho Done"),
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
			wantErr: "failed to create AES cipher",
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
