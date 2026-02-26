// Copyright (c) 2025 ZREV Enterprises LLC. All rights reserved.
// Use of this source code is governed by the MIT License.

package pbkdf2

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		password    string
		expectError bool
	}{
		{
			name:        "valid password",
			password:    "SecurePassword123!",
			expectError: false,
		},
		{
			name:        "empty password",
			password:    "",
			expectError: true,
		},
		{
			name:        "long password",
			password:    strings.Repeat("a", cryptoutilSharedMagic.JoseJADefaultListLimit),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			hash, err := HashPassword(tt.password)

			if tt.expectError {
				require.Error(t, err)
				require.Empty(t, hash)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, hash)
				require.True(t, strings.HasPrefix(hash, "$pbkdf2-sha256$"))

				// Verify hash format: $pbkdf2-sha256$iterations$salt$hash
				parts := strings.Split(hash, "$")
				require.Len(t, parts, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)
				require.Equal(t, "", parts[0])
				require.Equal(t, cryptoutilSharedMagic.PBKDF2Prefix, parts[1])
				require.Equal(t, "600000", parts[2])
				require.NotEmpty(t, parts[3]) // salt
				require.NotEmpty(t, parts[4]) // hash
			}
		})
	}
}

func TestVerifyPassword(t *testing.T) {
	t.Parallel()

	const testPassword = "TestPassword123!"

	// Generate a hash for testing.
	hash, err := HashPassword(testPassword)
	require.NoError(t, err)

	tests := []struct {
		name        string
		password    string
		storedHash  string
		expectMatch bool
		expectError bool
	}{
		{
			name:        "correct password",
			password:    testPassword,
			storedHash:  hash,
			expectMatch: true,
			expectError: false,
		},
		{
			name:        "incorrect password",
			password:    "WrongPassword",
			storedHash:  hash,
			expectMatch: false,
			expectError: false,
		},
		{
			name:        "empty password",
			password:    "",
			storedHash:  hash,
			expectMatch: false,
			expectError: true,
		},
		{
			name:        "empty hash",
			password:    testPassword,
			storedHash:  "",
			expectMatch: false,
			expectError: true,
		},
		{
			name:        "invalid hash format",
			password:    testPassword,
			storedHash:  "invalid-hash",
			expectMatch: false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			match, err := VerifyPassword(tt.password, tt.storedHash)

			if tt.expectError {
				require.Error(t, err)
				require.False(t, match)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectMatch, match)
			}
		})
	}
}

func TestDetectHashType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		hash         string
		expectedType string
	}{
		{
			name:         "bcrypt 2a",
			hash:         "$2a$12$abcdefghijklmnopqrstuv",
			expectedType: "bcrypt",
		},
		{
			name:         "bcrypt 2b",
			hash:         "$2b$12$abcdefghijklmnopqrstuv",
			expectedType: "bcrypt",
		},
		{
			name:         "bcrypt 2y",
			hash:         "$2y$12$abcdefghijklmnopqrstuv",
			expectedType: "bcrypt",
		},
		{
			name:         "pbkdf2",
			hash:         "$pbkdf2-sha256$600000$salt$hash",
			expectedType: "pbkdf2",
		},
		{
			name:         "unknown",
			hash:         "plain-text-password",
			expectedType: "unknown",
		},
		{
			name:         "empty",
			hash:         "",
			expectedType: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			hashType := DetectHashType(tt.hash)
			require.Equal(t, tt.expectedType, hashType)
		})
	}
}

func TestHashPasswordWithIterations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		password    string
		iterations  int
		expectError bool
	}{
		{
			name:        "minimum iterations (210000)",
			password:    "Password123!",
			iterations:  cryptoutilSharedMagic.PBKDF2MinIterations,
			expectError: false,
		},
		{
			name:        "recommended iterations (600000)",
			password:    "Password123!",
			iterations:  cryptoutilSharedMagic.IMPBKDF2Iterations,
			expectError: false,
		},
		{
			name:        "too few iterations",
			password:    "Password123!",
			iterations:  cryptoutilSharedMagic.PBKDF2Iterations,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			hash, err := HashPasswordWithIterations(tt.password, tt.iterations)

			if tt.expectError {
				require.Error(t, err)
				require.Empty(t, hash)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, hash)

				// Verify password works with generated hash.
				match, err := VerifyPassword(tt.password, hash)
				require.NoError(t, err)
				require.True(t, match)
			}
		})
	}
}

func TestDifferentPasswordsDifferentHashes(t *testing.T) {
	t.Parallel()

	hash1, err := HashPassword("Password1")
	require.NoError(t, err)

	hash2, err := HashPassword("Password2")
	require.NoError(t, err)

	require.NotEqual(t, hash1, hash2, "different passwords should produce different hashes")
}

func TestSamePasswordDifferentHashes(t *testing.T) {
	t.Parallel()

	const password = "TestPassword123!"

	hash1, err := HashPassword(password)
	require.NoError(t, err)

	hash2, err := HashPassword(password)
	require.NoError(t, err)

	require.NotEqual(t, hash1, hash2, "same password should produce different hashes due to random salt")

	// But both should verify successfully.
	match1, err := VerifyPassword(password, hash1)
	require.NoError(t, err)
	require.True(t, match1)

	match2, err := VerifyPassword(password, hash2)
	require.NoError(t, err)
	require.True(t, match2)
}
