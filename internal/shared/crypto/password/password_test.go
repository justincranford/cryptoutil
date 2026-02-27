// Copyright (c) 2025 ZREV Enterprises LLC. All rights reserved.
// Use of this source code is governed by the MIT License.

package password

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
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

func TestVerifyPassword_Bcrypt_Legacy(t *testing.T) {
	t.Parallel()

	const testPassword = "LegacyPassword123!"

	// Generate legacy bcrypt hash (simulating existing database record).
	bcryptHash, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)
	require.NoError(t, err)

	tests := []struct {
		name          string
		password      string
		expectMatch   bool
		expectUpgrade bool
	}{
		{
			name:          "correct legacy password",
			password:      testPassword,
			expectMatch:   true,
			expectUpgrade: true, // bcrypt always needs upgrade
		},
		{
			name:          "incorrect legacy password",
			password:      "WrongPassword",
			expectMatch:   false,
			expectUpgrade: true, // bcrypt always needs upgrade
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			match, needsUpgrade, err := VerifyPassword(tt.password, string(bcryptHash))
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

func TestMigrationWorkflow(t *testing.T) {
	t.Parallel()

	const userPassword = "UserPassword123!"

	// Step 1: User has legacy bcrypt password.
	bcryptHash, err := bcrypt.GenerateFromPassword([]byte(userPassword), bcrypt.DefaultCost)
	require.NoError(t, err)

	// Step 2: User logs in - verify with bcrypt (legacy).
	match, needsUpgrade, err := VerifyPassword(userPassword, string(bcryptHash))
	require.NoError(t, err)
	require.True(t, match, "legacy password should match")
	require.True(t, needsUpgrade, "bcrypt should always need upgrade")

	// Step 3: Upgrade to PBKDF2 (opportunistic migration).
	var newHash string
	if needsUpgrade {
		newHash, err = HashPassword(userPassword)
		require.NoError(t, err)
	}

	// Step 4: Verify with new PBKDF2 hash.
	match, needsUpgrade, err = VerifyPassword(userPassword, newHash)
	require.NoError(t, err)
	require.True(t, match, "upgraded password should match")
	require.False(t, needsUpgrade, "PBKDF2 should not need upgrade")

	// Step 5: Wrong password should still fail.
	match, _, err = VerifyPassword("WrongPassword", newHash)
	require.NoError(t, err)
	require.False(t, match, "wrong password should not match")
}
