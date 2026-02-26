// Copyright (c) 2025 Justin Cranford
//
//

package hash

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestHashHighEntropyNonDeterministic tests the high entropy random hash provider.
func TestHashHighEntropyNonDeterministic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		secret      string
		expectError bool
	}{
		{name: "valid_high_entropy_secret", secret: "api-key-1234567890abcdefghijklmnopqrstuvwxyz", expectError: false},
		{name: "empty_secret", secret: "", expectError: true},
		{name: "long_secret", secret: strings.Repeat("x", cryptoutilSharedMagic.DefaultLogsBatchSize), expectError: false},
		{name: "unicode_secret", secret: "秘密🔐密钥", expectError: false},
		{name: "special_characters", secret: "!@#$%^&*()_+-={}[]|\\:;\"'<>,.?/", expectError: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			hash, err := HashHighEntropyNonDeterministic(tt.secret)

			if tt.expectError {
				require.Error(t, err, "expected error for %s", tt.name)
				require.Empty(t, hash, "hash should be empty on error")
			} else {
				require.NoError(t, err, "should hash successfully")
				require.NotEmpty(t, hash, "hash should not be empty")
				require.True(t, strings.HasPrefix(hash, "hkdf-sha256$"), "hash should have correct prefix")

				// Verify hash.
				valid, err := VerifySecretHKDFRandom(hash, tt.secret)
				require.NoError(t, err, "should verify hash")
				require.True(t, valid, "hash should verify correctly")
			}
		})
	}
}

// TestHashSecretHKDFRandom tests HKDF random salt hashing.
func TestHashSecretHKDFRandom(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		secret      string
		expectError bool
	}{
		{name: "valid_secret", secret: "high-entropy-secret-1234567890", expectError: false},
		{name: "empty_secret", secret: "", expectError: true},
		{name: "long_secret", secret: strings.Repeat("a", cryptoutilSharedMagic.MaxUnsealSharedSecrets), expectError: false},
		{name: "unicode_secret", secret: "日本語パスワード", expectError: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			hash, err := HashSecretHKDFRandom(tt.secret)

			if tt.expectError {
				require.Error(t, err, "expected error for %s", tt.name)
				require.Empty(t, hash, "hash should be empty on error")
			} else {
				require.NoError(t, err, "should hash successfully")
				require.NotEmpty(t, hash, "hash should not be empty")

				// Parse format: hkdf-sha256$salt$dk.
				const expectedParts = 3

				parts := strings.Split(hash, "$")
				require.Len(t, parts, expectedParts, "hash should have exactly %d parts", expectedParts)
				require.Equal(t, cryptoutilSharedMagic.HKDFHashName, parts[0], "hash algorithm should be hkdf-sha256")
			}
		})
	}
}

// TestHashSecretHKDFRandom_Uniqueness tests that each hash is unique (non-deterministic).
func TestHashSecretHKDFRandom_Uniqueness(t *testing.T) {
	t.Parallel()

	secret := "same-secret-for-all-hashes"
	iterations := cryptoutilSharedMagic.JoseJADefaultMaxMaterials

	hashes := make(map[string]bool)

	for range iterations {
		hash, err := HashSecretHKDFRandom(secret)
		require.NoError(t, err, "should hash successfully")

		// Each hash should be unique due to random salt.
		require.False(t, hashes[hash], "hash should be unique")
		hashes[hash] = true
	}

	require.Len(t, hashes, iterations, "should have %d unique hashes", iterations)
}

// TestVerifySecretHKDFRandom tests HKDF hash verification.
func TestVerifySecretHKDFRandom(t *testing.T) {
	t.Parallel()

	secret := "test-secret-for-verification"
	hash, err := HashSecretHKDFRandom(secret)
	require.NoError(t, err, "should hash successfully")

	tests := []struct {
		name          string
		stored        string
		provided      string
		expectError   bool
		expectedValid bool
	}{
		{name: "valid_hash_matches", stored: hash, provided: secret, expectError: false, expectedValid: true},
		{name: "valid_hash_wrong_password", stored: hash, provided: "wrong-secret", expectError: false, expectedValid: false},
		{name: "empty_stored_hash", stored: "", provided: secret, expectError: true, expectedValid: false},
		{name: "empty_provided_secret", stored: hash, provided: "", expectError: true, expectedValid: false},
		{name: "invalid_hash_format", stored: "not-a-valid-hash", expectError: true, expectedValid: false},
		{name: "invalid_salt_encoding", stored: "hkdf-sha256$invalid!!!$abc123", expectError: true, expectedValid: false},
		{name: "invalid_dk_encoding", stored: "hkdf-sha256$abc123$invalid!!!", expectError: true, expectedValid: false},
		{name: "wrong_algorithm", stored: "unknown-algo$abc123$def456", expectError: true, expectedValid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			valid, err := VerifySecretHKDFRandom(tt.stored, tt.provided)

			if tt.expectError {
				require.Error(t, err, "expected error for %s", tt.name)
			} else {
				require.NoError(t, err, "should not error for %s", tt.name)
				require.Equal(t, tt.expectedValid, valid, "validity should match expected")
			}
		})
	}
}

// TestHashHighEntropyNonDeterministic_CrossVerification tests that different hash instances verify correctly.
func TestHashHighEntropyNonDeterministic_CrossVerification(t *testing.T) {
	t.Parallel()

	secret := "cross-verification-test"

	// Generate multiple hashes.
	hash1, err := HashHighEntropyNonDeterministic(secret)
	require.NoError(t, err, "should hash successfully")

	hash2, err := HashHighEntropyNonDeterministic(secret)
	require.NoError(t, err, "should hash successfully")

	// Hashes should be different (random salt).
	require.NotEqual(t, hash1, hash2, "hashes should be different due to random salt")

	// Both should verify correctly.
	valid1, err := VerifySecretHKDFRandom(hash1, secret)
	require.NoError(t, err, "should verify hash1")
	require.True(t, valid1, "hash1 should verify")

	valid2, err := VerifySecretHKDFRandom(hash2, secret)
	require.NoError(t, err, "should verify hash2")
	require.True(t, valid2, "hash2 should verify")

	// Wrong password should not verify.
	invalid1, err := VerifySecretHKDFRandom(hash1, "wrong-secret")
	require.NoError(t, err, "should not error on wrong password")
	require.False(t, invalid1, "hash1 should not verify with wrong password")

	invalid2, err := VerifySecretHKDFRandom(hash2, "wrong-secret")
	require.NoError(t, err, "should not error on wrong password")
	require.False(t, invalid2, "hash2 should not verify with wrong password")
}
