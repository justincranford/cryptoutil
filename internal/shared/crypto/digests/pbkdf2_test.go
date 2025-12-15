// Copyright (c) 2025 Justin Cranford

package digests

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Test constants for password strings.
const (
	testPassword = "password"
)

func TestHashSecretPBKDF2(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		secret  string
		wantErr bool
	}{
		{
			name:    "valid secret",
			secret:  "test-password-123",
			wantErr: false,
		},
		{
			name:    "empty secret",
			secret:  "",
			wantErr: true,
		},
		{
			name:    "long secret",
			secret:  strings.Repeat("a", 1000),
			wantErr: false,
		},
		{
			name:    "special characters",
			secret:  "!@#$%^&*()_+-=[]{}|;':\",./<>?",
			wantErr: false,
		},
		{
			name:    "unicode secret",
			secret:  "密码测试🔐",
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			hash, err := HashSecretPBKDF2(tc.secret)

			if tc.wantErr {
				require.Error(t, err)
				require.Empty(t, hash)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, hash)
				require.True(t, strings.HasPrefix(hash, "{1}$pbkdf2-sha256$"))

				parts := strings.Split(hash, "$")
				require.Len(t, parts, 5)
				require.Equal(t, "{1}", parts[0])
				require.Equal(t, "pbkdf2-sha256", parts[1])
			}
		})
	}
}

func TestHashSecret(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		secret  string
		wantErr bool
	}{
		{
			name:    "valid secret",
			secret:  "test-password",
			wantErr: false,
		},
		{
			name:    "empty secret",
			secret:  "",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			hash, err := HashLowEntropyNonDeterministic(tc.secret)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, hash)
				require.True(t, strings.HasPrefix(hash, "{1}$pbkdf2-sha256$"))
			}
		})
	}
}

func TestVerifySecret(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func() (string, string)
		wantOK   bool
		wantErr  bool
		errMatch string
	}{
		{
			name: "valid PBKDF2 hash matches",
			setup: func() (string, string) {
				secret := "correct-password"
				hash, _ := HashSecretPBKDF2(secret)

				return hash, secret
			},
			wantOK:  true,
			wantErr: false,
		},
		{
			name: "valid PBKDF2 hash does not match wrong password",
			setup: func() (string, string) {
				hash, _ := HashSecretPBKDF2("correct-password")

				return hash, "wrong-password"
			},
			wantOK:  false,
			wantErr: false,
		},
		{
			name: "empty stored hash",
			setup: func() (string, string) {
				return "", "any-password"
			},
			wantOK:   false,
			wantErr:  true,
			errMatch: "stored hash empty",
		},
		{
			name: "invalid hash format",
			setup: func() (string, string) {
				return "invalid$format", testPassword
			},
			wantOK:   false,
			wantErr:  true,
			errMatch: "invalid legacy hash format",
		},
		{
			name: "invalid iterations in hash",
			setup: func() (string, string) {
				return "pbkdf2-sha256$invalid$salt$dk", testPassword
			},
			wantOK:   false,
			wantErr:  true,
			errMatch: "invalid iterations",
		},
		{
			name: "zero iterations in hash",
			setup: func() (string, string) {
				return "pbkdf2-sha256$0$salt$dk", testPassword
			},
			wantOK:   false,
			wantErr:  true,
			errMatch: "invalid iterations",
		},
		{
			name: "invalid salt encoding",
			setup: func() (string, string) {
				return "pbkdf2-sha256$1000$!!!invalid!!!$dk", testPassword
			},
			wantOK:   false,
			wantErr:  true,
			errMatch: "invalid salt encoding",
		},
		{
			name: "invalid dk encoding",
			setup: func() (string, string) {
				return "pbkdf2-sha256$1000$dGVzdA$!!!invalid!!!", testPassword
			},
			wantOK:   false,
			wantErr:  true,
			errMatch: "invalid dk encoding",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			stored, provided := tc.setup()
			ok, err := VerifySecret(stored, provided)

			if tc.wantErr {
				require.Error(t, err)

				if tc.errMatch != "" {
					require.Contains(t, err.Error(), tc.errMatch)
				}
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tc.wantOK, ok)
		})
	}
}

func TestVerifySecret_LegacyBcrypt(t *testing.T) {
	t.Parallel()

	// Test legacy bcrypt hash verification (migration support).
	// Using pre-computed bcrypt hash for known passwords.
	bcryptHashes := []struct {
		name     string
		hash     string
		prefix   string
		password string
	}{
		{
			name: "bcrypt 2a prefix",
			// Pre-computed bcrypt hash for "test" with cost 10.
			hash:     "$2a$10$2u64ymwxrqdCLciYcWvnhu/yVIMZJvEGqY6.K4Nor56xHIAXib22q",
			prefix:   "$2a$",
			password: "test",
		},
	}

	for _, tc := range bcryptHashes {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Verify the hash has the expected prefix.
			require.True(t, strings.HasPrefix(tc.hash, tc.prefix))

			// Verify correct password matches.
			ok, err := VerifySecret(tc.hash, tc.password)
			require.NoError(t, err)
			require.True(t, ok, "password should match hash")

			// Wrong password should not match.
			ok, err = VerifySecret(tc.hash, "wrong-password")
			require.NoError(t, err)
			require.False(t, ok, "wrong password should not match")
		})
	}
}

func TestHashSecretPBKDF2_Uniqueness(t *testing.T) {
	t.Parallel()

	// Same secret should produce different hashes (due to random salt).
	secret := "same-password"
	hash1, err1 := HashSecretPBKDF2(secret)
	hash2, err2 := HashSecretPBKDF2(secret)

	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NotEqual(t, hash1, hash2, "hashes should be unique due to random salt")

	// Both hashes should still verify against the same secret.
	ok1, err := VerifySecret(hash1, secret)
	require.NoError(t, err)
	require.True(t, ok1)

	ok2, err := VerifySecret(hash2, secret)
	require.NoError(t, err)
	require.True(t, ok2)
}
