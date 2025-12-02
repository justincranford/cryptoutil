// Copyright (c) 2025 Justin Cranford

package userauth

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Validates requirements:
// - R04-01: Client secrets hashed with PBKDF2-HMAC-SHA256.
func TestHashToken_Success(t *testing.T) {
	t.Parallel()

	plaintext := "test-otp-123456"
	hash, err := HashToken(plaintext)

	require.NoError(t, err)
	require.NotEmpty(t, hash)
	require.NotEqual(t, plaintext, hash, "Hash must differ from plaintext")
	require.True(t, strings.HasPrefix(hash, "pbkdf2-sha256$"), "PBKDF2 hash must have pbkdf2-sha256$ prefix")
}

func TestHashToken_EmptyToken(t *testing.T) {
	t.Parallel()

	hash, err := HashToken("")

	require.ErrorIs(t, err, ErrInvalidToken)
	require.Empty(t, hash)
}

func TestHashToken_DifferentHashesForSameToken(t *testing.T) {
	t.Parallel()

	plaintext := "test-token-collision"
	hash1, err1 := HashToken(plaintext)
	hash2, err2 := HashToken(plaintext)

	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NotEqual(t, hash1, hash2, "PBKDF2 must produce different hashes due to random salt")
}

func TestHashToken_PBKDF2Format(t *testing.T) {
	t.Parallel()

	plaintext := "format-check-token"
	hash, err := HashToken(plaintext)

	require.NoError(t, err)
	require.True(t, strings.HasPrefix(hash, "pbkdf2-sha256$"), "Hash must use PBKDF2-SHA256 format")
	require.Contains(t, hash, "$", "Hash must contain iteration separator")
}

func TestVerifyToken_Success(t *testing.T) {
	t.Parallel()

	plaintext := "verify-me-token"
	hash, err := HashToken(plaintext)
	require.NoError(t, err)

	err = VerifyToken(plaintext, hash)
	require.NoError(t, err, "Verification must succeed for correct plaintext")
}

func TestVerifyToken_Mismatch(t *testing.T) {
	t.Parallel()

	plaintext := "correct-token"
	wrongPlaintext := "wrong-token"
	hash, err := HashToken(plaintext)
	require.NoError(t, err)

	err = VerifyToken(wrongPlaintext, hash)
	require.ErrorIs(t, err, ErrTokenMismatch)
}

func TestVerifyToken_EmptyPlaintext(t *testing.T) {
	t.Parallel()

	hash, err := HashToken("valid-token")
	require.NoError(t, err)

	err = VerifyToken("", hash)
	require.ErrorIs(t, err, ErrInvalidToken)
}

func TestVerifyToken_EmptyHash(t *testing.T) {
	t.Parallel()

	err := VerifyToken("some-token", "")
	require.ErrorIs(t, err, ErrTokenMismatch)
}

func TestVerifyToken_MalformedHash(t *testing.T) {
	t.Parallel()

	err := VerifyToken("some-token", "invalid-bcrypt-hash")
	require.ErrorIs(t, err, ErrTokenMismatch)
}

func TestHashAndVerify_RoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		plaintext string
	}{
		{"Short OTP", "123456"},
		{"Long magic link token", "abcd1234-efgh5678-ijkl9012-mnop3456"},
		{"Special characters", "token!@#$%^&*()_+-={}[]|\\:;\"'<>,.?/~`"},
		{"Unicode token", "🔐🔑🗝️tokenΩΨΦ"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			hash, err := HashToken(tc.plaintext)
			require.NoError(t, err)

			err = VerifyToken(tc.plaintext, hash)
			require.NoError(t, err, "Round-trip hash/verify must succeed")

			// Verify wrong token fails.
			err = VerifyToken(tc.plaintext+"wrong", hash)
			require.ErrorIs(t, err, ErrTokenMismatch)
		})
	}
}
