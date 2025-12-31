// Copyright (c) 2025 Justin Cranford

package crypto

import (
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilRandom "cryptoutil/internal/shared/util/random"
)

func TestHashPassword(t *testing.T) {
	t.Parallel()

	password, err := cryptoutilRandom.GeneratePasswordSimple()
	require.NoError(t, err)

	// Use fast version for testing (1,000 iterations vs 600,000).
	hash, err := HashPasswordForTest(password)
	require.NoError(t, err)
	require.NotNil(t, hash)
	require.Equal(t, 64, len(hash), "hash should be 64 bytes (32 salt + 32 hash)")
}

func TestHashPassword_DifferentSalts(t *testing.T) {
	t.Parallel()

	password := "SamePassword" // pragma: allowlist secret - Test vector for password hashing

	// Use fast version for testing (1,000 iterations vs 600,000).
	hash1, err := HashPasswordForTest(password)
	require.NoError(t, err)

	hash2, err := HashPasswordForTest(password)
	require.NoError(t, err)

	// Hashes should differ due to different random salts.
	require.False(t, string(hash1) == string(hash2), "hashes should differ with different salts")
}

func TestVerifyPassword_Success(t *testing.T) {
	t.Parallel()

	password, err := cryptoutilRandom.GeneratePasswordSimple()
	require.NoError(t, err)

	// Use fast version for testing (1,000 iterations vs 600,000).
	hash, err := HashPasswordForTest(password)
	require.NoError(t, err)

	match, err := VerifyPasswordForTest(password, hash)
	require.NoError(t, err)
	require.True(t, match, "password should verify successfully")
}

func TestVerifyPassword_WrongPassword(t *testing.T) {
	t.Parallel()

	password := "CorrectPassword"    // pragma: allowlist secret - Test vector for password verification
	wrongPassword := "WrongPassword" // pragma: allowlist secret - Test vector for password verification

	// Use fast version for testing (1,000 iterations vs 600,000).
	hash, err := HashPasswordForTest(password)
	require.NoError(t, err)

	match, err := VerifyPasswordForTest(wrongPassword, hash)
	require.NoError(t, err)
	require.False(t, match, "wrong password should not verify")
}

func TestVerifyPassword_InvalidHashLength(t *testing.T) {
	t.Parallel()

	password, err := cryptoutilRandom.GeneratePasswordSimple()
	require.NoError(t, err)

	invalidHash := []byte{1, 2, 3} // Too short.

	// Use fast version for testing (1,000 iterations vs 600,000).
	match, err := VerifyPasswordForTest(password, invalidHash)
	require.Error(t, err)
	require.False(t, match)
	require.Contains(t, err.Error(), "invalid stored hash length")
}

func TestHashPassword_EmptyPassword(t *testing.T) {
	t.Parallel()

	// Use fast version for testing (1,000 iterations vs 600,000).
	hash, err := HashPasswordForTest("")
	require.NoError(t, err)
	require.NotNil(t, hash)

	match, err := VerifyPasswordForTest("", hash)
	require.NoError(t, err)
	require.True(t, match, "empty password should verify")
}

func TestCompareHashes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    []byte
		b    []byte
		want bool
	}{
		{
			name: "equal hashes",
			a:    []byte{1, 2, 3, 4},
			b:    []byte{1, 2, 3, 4},
			want: true,
		},
		{
			name: "different hashes",
			a:    []byte{1, 2, 3, 4},
			b:    []byte{4, 3, 2, 1},
			want: false,
		},
		{
			name: "different lengths",
			a:    []byte{1, 2, 3},
			b:    []byte{1, 2, 3, 4},
			want: false,
		},
		{
			name: "empty hashes",
			a:    []byte{},
			b:    []byte{},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := compareHashes(tt.a, tt.b)
			require.Equal(t, tt.want, got)
		})
	}
}
