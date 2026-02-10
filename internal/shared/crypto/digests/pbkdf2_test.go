// Copyright (c) 2025 Justin Cranford

package digests

import (
	sha256 "crypto/sha256"
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// Test constants for password strings.
var (
	testPassword = googleUuid.Must(googleUuid.NewV7()).String()
)

func FastPBKDF2ParameterSet() *PBKDF2Params {
	return &PBKDF2Params{
		Version:    "1",
		HashName:   cryptoutilSharedMagic.PBKDF2DefaultHashName,
		Iterations: cryptoutilSharedMagic.PBKDF2DefaultIterations,
		SaltLength: cryptoutilSharedMagic.PBKDF2DefaultSaltBytes,
		KeyLength:  cryptoutilSharedMagic.PBKDF2DerivedKeyLength,
		HashFunc:   sha256.New,
	}
}

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

			hash, err := PBKDF2WithParams(tc.secret, FastPBKDF2ParameterSet())

			if tc.wantErr {
				require.Error(t, err)
				require.Empty(t, hash)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, hash)
				require.True(t, strings.HasPrefix(hash, "{1}$"+cryptoutilSharedMagic.PBKDF2DefaultHashName+"$"))

				parts := strings.Split(hash, "$")
				require.Len(t, parts, 5)
				require.Equal(t, "{1}", parts[0])
				require.Equal(t, cryptoutilSharedMagic.PBKDF2DefaultHashName, parts[1])
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
				hash, _ := PBKDF2WithParams(secret, FastPBKDF2ParameterSet())

				return hash, secret
			},
			wantOK:  true,
			wantErr: false,
		},
		{
			name: "valid PBKDF2 hash does not match wrong password",
			setup: func() (string, string) {
				hash, _ := PBKDF2WithParams("correct-password", FastPBKDF2ParameterSet())

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
			name: "invalid hash format - non-versioned",
			setup: func() (string, string) {
				return "invalid$format", testPassword
			},
			wantOK:   false,
			wantErr:  true,
			errMatch: "unsupported hash format",
		},
		{
			name: "invalid iterations in hash - old format rejected",
			setup: func() (string, string) {
				return cryptoutilSharedMagic.PBKDF2DefaultHashName + "$invalid$salt$dk", testPassword
			},
			wantOK:   false,
			wantErr:  true,
			errMatch: "unsupported hash format",
		},
		{
			name: "zero iterations in hash - old format rejected",
			setup: func() (string, string) {
				return cryptoutilSharedMagic.PBKDF2DefaultHashName + "$0$salt$dk", testPassword
			},
			wantOK:   false,
			wantErr:  true,
			errMatch: "unsupported hash format",
		},
		{
			name: "invalid salt encoding - old format rejected",
			setup: func() (string, string) {
				return cryptoutilSharedMagic.PBKDF2DefaultHashName + "$1000$!!!invalid!!!$dk", testPassword
			},
			wantOK:   false,
			wantErr:  true,
			errMatch: "unsupported hash format",
		},
		{
			name: "invalid dk encoding - old format rejected",
			setup: func() (string, string) {
				return cryptoutilSharedMagic.PBKDF2DefaultHashName + "$1000$dGVzdA$!!!invalid!!!", testPassword
			},
			wantOK:   false,
			wantErr:  true,
			errMatch: "unsupported hash format",
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

// TestVerifySecret_LegacyBcrypt removed - bcrypt is BANNED (NOT FIPS 140-3 approved).
// Legacy password migration should use PBKDF2 with lower parameter sets (V3=2017 with 1000 iterations),
// not banned cryptographic algorithms like bcrypt.

func TestHashSecretPBKDF2_Uniqueness(t *testing.T) {
	t.Parallel()

	// Same secret should produce different hashes (due to random salt).
	secret := "same-password"
	hash1, err1 := PBKDF2WithParams(secret, FastPBKDF2ParameterSet())
	hash2, err2 := PBKDF2WithParams(secret, FastPBKDF2ParameterSet())

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
