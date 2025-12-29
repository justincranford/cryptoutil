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

	hash, err := HashPassword(password)
	require.NoError(t, err)
	require.NotNil(t, hash)
	require.Equal(t, 64, len(hash), "hash should be 64 bytes (32 salt + 32 hash)")
}

func TestHashPassword_DifferentSalts(t *testing.T) {
	t.Parallel()

	password := "SamePassword"

	hash1, err := HashPassword(password)
	require.NoError(t, err)

	hash2, err := HashPassword(password)
	require.NoError(t, err)

	// Hashes should differ due to different random salts.
	require.False(t, string(hash1) == string(hash2), "hashes should differ with different salts")
}

func TestVerifyPassword_Success(t *testing.T) {
	t.Parallel()

	password, err := cryptoutilRandom.GeneratePasswordSimple()
	require.NoError(t, err)

	hash, err := HashPassword(password)
	require.NoError(t, err)

	match, err := VerifyPassword(password, hash)
	require.NoError(t, err)
	require.True(t, match, "password should verify successfully")
}

func TestVerifyPassword_WrongPassword(t *testing.T) {
	t.Parallel()

	password := "CorrectPassword"
	wrongPassword := "WrongPassword"

	hash, err := HashPassword(password)
	require.NoError(t, err)

	match, err := VerifyPassword(wrongPassword, hash)
	require.NoError(t, err)
	require.False(t, match, "wrong password should not verify")
}

func TestVerifyPassword_InvalidHashLength(t *testing.T) {
	t.Parallel()

	password, err := cryptoutilRandom.GeneratePasswordSimple()
	require.NoError(t, err)
	invalidHash := []byte{1, 2, 3} // Too short.

	match, err := VerifyPassword(password, invalidHash)
	require.Error(t, err)
	require.False(t, match)
	require.Contains(t, err.Error(), "invalid stored hash length")
}

func TestHashPassword_EmptyPassword(t *testing.T) {
	t.Parallel()

	hash, err := HashPassword("")
	require.NoError(t, err)
	require.NotNil(t, hash)

	match, err := VerifyPassword("", hash)
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
