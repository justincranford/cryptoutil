// Copyright (c) 2025 ZREV Enterprises LLC. All rights reserved.
// Use of this source code is governed by the MIT License.

package password

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	t.Parallel()

	hash, err := HashPassword("TestPassword123!")
	require.NoError(t, err)
	require.NotEmpty(t, hash)
	require.Contains(t, hash, "$pbkdf2-sha256$")
}

func TestVerifyPassword_PBKDF2(t *testing.T) {
	t.Parallel()

	const testPassword = "TestPassword123!"

	// Generate PBKDF2 hash.
	hash, err := HashPassword(testPassword)
	require.NoError(t, err)

	tests := []struct {
		name          string
		password      string
		expectMatch   bool
		expectUpgrade bool
	}{
		{
			name:          "correct password",
			password:      testPassword,
			expectMatch:   true,
			expectUpgrade: false, // PBKDF2 doesn't need upgrade
		},
		{
			name:          "incorrect password",
			password:      "WrongPassword",
			expectMatch:   false,
			expectUpgrade: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			match, needsUpgrade, err := VerifyPassword(tt.password, hash)
			require.NoError(t, err)
			require.Equal(t, tt.expectMatch, match)
			require.Equal(t, tt.expectUpgrade, needsUpgrade)
		})
	}
}

func TestVerifyPassword_EmptyInputs(t *testing.T) {
	t.Parallel()

	hash, err := HashPassword("TestPassword123!")
	require.NoError(t, err)

	tests := []struct {
		name        string
		password    string
		storedHash  string
		expectError bool
	}{
		{
			name:        "empty password",
			password:    "",
			storedHash:  hash,
			expectError: true,
		},
		{
			name:        "empty hash",
			password:    "TestPassword123!",
			storedHash:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			match, needsUpgrade, err := VerifyPassword(tt.password, tt.storedHash)
			require.Error(t, err)
			require.False(t, match)
			require.False(t, needsUpgrade)
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
			name:         "non-fips140-hash 2a",
			hash:         "$2a$12$abcdefghijklmnopqrstuv",
			expectedType: "unknown",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			hashType := DetectHashType(tt.hash)
			require.Equal(t, tt.expectedType, hashType)
		})
	}
}

func TestVerifyPassword_BcryptAndUnknownRejected(t *testing.T) {
	t.Parallel()

	// Verify that legacy bcrypt hashes and unknown formats are rejected.
	tests := []struct {
		name string
		hash string
	}{
		{name: "bcrypt 2a prefix", hash: "$2a$12$abcdefghijklmnopqrstuv"},
		{name: "bcrypt 2b prefix", hash: "$2b$12$abcdefghijklmnopqrstuv"},
		{name: "plain text", hash: "not-a-real-hash"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			match, needsUpgrade, err := VerifyPassword("anypassword", tt.hash)
			require.Error(t, err)
			require.False(t, match)
			require.False(t, needsUpgrade)
		})
	}
}
