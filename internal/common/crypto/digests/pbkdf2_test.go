// Copyright (c) 2025 Justin Cranford

package crypto_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCryptoDigests "cryptoutil/internal/common/crypto/digests"
)

func TestHashSecretPBKDF2_HappyPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		secret string
	}{
		{"ShortPassword", "password123"},
		{"LongPassword", "this_is_a_very_long_password_with_many_characters_1234567890"},
		{"SpecialChars", "p@ssw0rd!#$%"},
		{"Unicode", "пароль密码"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			hash, err := cryptoutilCryptoDigests.HashSecretPBKDF2(tc.secret)
			require.NoError(t, err)
			require.NotEmpty(t, hash)
			require.True(t, strings.HasPrefix(hash, "pbkdf2-sha256$"))

			// Verify format: pbkdf2-sha256$iter$salt$dk (4 parts)
			parts := strings.Split(hash, "$")
			require.Len(t, parts, 4, "Expected format: pbkdf2-sha256$iter$salt$dk")
			require.Equal(t, "pbkdf2-sha256", parts[0])
			require.NotEmpty(t, parts[1]) // iterations
			require.NotEmpty(t, parts[2]) // salt
			require.NotEmpty(t, parts[3]) // derived key
		})
	}
}

func TestHashSecretPBKDF2_SadPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		secret      string
		expectError bool
	}{
		{"EmptySecret", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			hash, err := cryptoutilCryptoDigests.HashSecretPBKDF2(tc.secret)

			if tc.expectError {
				require.Error(t, err)
				require.Empty(t, hash)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, hash)
			}
		})
	}
}

func TestHashSecretPBKDF2_Uniqueness(t *testing.T) {
	t.Parallel()

	// Same secret should produce different hashes (due to random salt)
	secret := "test_password"

	hash1, err := cryptoutilCryptoDigests.HashSecretPBKDF2(secret)
	require.NoError(t, err)

	hash2, err := cryptoutilCryptoDigests.HashSecretPBKDF2(secret)
	require.NoError(t, err)

	require.NotEqual(t, hash1, hash2, "Same secret should produce different hashes (random salt)")
}

func TestVerifySecret_PBKDF2_HappyPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		secret string
	}{
		{"ShortPassword", "password123"},
		{"LongPassword", "this_is_a_very_long_password_with_many_characters_1234567890"},
		{"SpecialChars", "p@ssw0rd!#$%"},
		{"Unicode", "пароль密码"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Hash the secret
			hash, err := cryptoutilCryptoDigests.HashSecretPBKDF2(tc.secret)
			require.NoError(t, err)

			// Verify correct secret
			ok, err := cryptoutilCryptoDigests.VerifySecret(hash, tc.secret)
			require.NoError(t, err)
			require.True(t, ok, "Correct secret should verify")

			// Verify wrong secret
			ok, err = cryptoutilCryptoDigests.VerifySecret(hash, tc.secret+"wrong")
			require.NoError(t, err)
			require.False(t, ok, "Wrong secret should not verify")
		})
	}
}

func TestVerifySecret_SadPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		stored      string
		provided    string
		expectError bool
		shouldMatch bool
	}{
		{"EmptyStored", "", "password", true, false},
		{"UnsupportedFormat", "md5$abc123", "password", true, false},
		{"InvalidIterations", "pbkdf2-sha256$abc$salt$dk", "password", true, false},
		{"InvalidSaltEncoding", "pbkdf2-sha256$10000$!!!invalid!!!$dk", "password", true, false},
		{"InvalidDKEncoding", "pbkdf2-sha256$10000$c2FsdA$!!!invalid!!!", "password", true, false},
		{"TooFewParts", "pbkdf2-sha256$10000$salt", "password", true, false},
		{"TooManyParts", "pbkdf2-sha256$10000$salt$dk$extra", "password", true, false},
		{"ZeroIterations", "pbkdf2-sha256$0$c2FsdA$ZGs", "password", true, false},
		{"NegativeIterations", "pbkdf2-sha256$-1$c2FsdA$ZGs", "password", true, false},
		{"EmptyProvided", "pbkdf2-sha256$10000$c2FsdA$ZGs", "", false, false},
		{"WrongAlgorithm", "pbkdf2$10000$c2FsdA$ZGs", "password", true, false},
		{"DifferentLengthDK", "pbkdf2-sha256$10000$c2FsdA$c2hvcnQ", "password", false, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ok, err := cryptoutilCryptoDigests.VerifySecret(tc.stored, tc.provided)

			if tc.expectError {
				require.Error(t, err)
				require.False(t, ok)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.shouldMatch, ok)
			}
		})
	}
}

func TestVerifySecret_BCryptLegacy(t *testing.T) {
	t.Parallel()

	// Skip bcrypt legacy tests - bcrypt is NOT FIPS 140-3 approved
	// Legacy bcrypt support deprecated in favor of PBKDF2-HMAC-SHA256
	t.Skip("bcrypt is NOT FIPS 140-3 approved - legacy support will be removed")
}

func TestVerifySecret_RoundTrip(t *testing.T) {
	t.Parallel()

	secrets := []string{
		"password123",
		"this_is_a_very_long_password_with_many_characters_1234567890",
		"p@ssw0rd!#$%",
		"пароль密码",
	}

	for _, secret := range secrets {
		t.Run("RoundTrip_"+secret, func(t *testing.T) {
			t.Parallel()

			// Hash
			hash, err := cryptoutilCryptoDigests.HashSecretPBKDF2(secret)
			require.NoError(t, err)

			// Verify correct
			ok, err := cryptoutilCryptoDigests.VerifySecret(hash, secret)
			require.NoError(t, err)
			require.True(t, ok)

			// Verify wrong
			ok, err = cryptoutilCryptoDigests.VerifySecret(hash, secret+"wrong")
			require.NoError(t, err)
			require.False(t, ok)
		})
	}
}
