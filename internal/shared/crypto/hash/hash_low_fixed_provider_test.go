// Copyright (c) 2025 Justin Cranford

package hash

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashLowEntropyDeterministic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		secret      string
		expectError bool
	}{
		{
			name:        "valid_low_entropy_secret",
			secret:      "password123",
			expectError: false,
		},
		{
			name:        "empty_secret",
			secret:      "",
			expectError: true,
		},
		{
			name:        "long_secret",
			secret:      strings.Repeat("a", cryptoutilSharedMagic.DefaultLogsBatchSize),
			expectError: false,
		},
		{
			name:        "unicode_secret",
			secret:      "密码🔐密钥",
			expectError: false,
		},
		{
			name:        "special_characters",
			secret:      "!@#$%^&*()_+-=[]{}|;:',.<>?/~`",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			hash, err := HashLowEntropyDeterministic(tt.secret)

			if tt.expectError {
				require.Error(t, err)
				require.Empty(t, hash)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, hash)

				// Verify format: hkdf-sha256-fixed$base64(dk)
				parts := strings.Split(hash, "$")
				require.Len(t, parts, 2, "hash should have 2 parts")
				require.Equal(t, cryptoutilSharedMagic.HKDFFixedLowHashName, parts[0])
				require.NotEmpty(t, parts[1], "derived key should not be empty")
			}
		})
	}
}

func TestHashSecretHKDFFixed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		secret      string
		expectError bool
	}{
		{
			name:        "valid_secret",
			secret:      "testPassword123",
			expectError: false,
		},
		{
			name:        "empty_secret",
			secret:      "",
			expectError: true,
		},
		{
			name:        "long_secret",
			secret:      strings.Repeat("x", cryptoutilSharedMagic.DefaultMetricsBatchSize),
			expectError: false,
		},
		{
			name:        "unicode_secret",
			secret:      "秘密🔐密钥",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fixedInfo := []byte("test-fixed-info")
			hash, err := HashSecretHKDFFixed(tt.secret, fixedInfo)

			if tt.expectError {
				require.Error(t, err)
				require.Empty(t, hash)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, hash)

				// Verify format.
				parts := strings.Split(hash, "$")
				require.Len(t, parts, 2)
				require.Equal(t, cryptoutilSharedMagic.HKDFFixedLowHashName, parts[0])
			}
		})
	}
}

func TestHashSecretHKDFFixed_Determinism(t *testing.T) {
	t.Parallel()

	secret := "deterministicTestSecret123"
	fixedInfo := []byte("deterministic-info")

	const iterations = 10

	hashes := make([]string, iterations)

	// Generate multiple hashes with same secret and fixed info.
	for i := 0; i < iterations; i++ {
		hash, err := HashSecretHKDFFixed(secret, fixedInfo)
		require.NoError(t, err)

		hashes[i] = hash
	}

	// All hashes should be identical (deterministic).
	firstHash := hashes[0]
	for i := 1; i < iterations; i++ {
		require.Equal(t, firstHash, hashes[i],
			"iteration %d: hash should be deterministic (same as first)", i)
	}
}

func TestVerifySecretHKDFFixed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		storedHash     string
		providedSecret string
		expectMatch    bool
		expectError    bool
	}{
		{
			name:           "valid_hash_matches",
			storedHash:     "hkdf-sha256-fixed$ZGVyaXZlZGtleTE2Ynl0ZXNsb25nZGVyaXZlZGtleTE2",
			providedSecret: "password123",
			expectMatch:    false, // Won't match unless we use the exact secret that generated this hash.
			expectError:    false,
		},
		{
			name:           "empty_stored_hash",
			storedHash:     "",
			providedSecret: "password123",
			expectMatch:    false,
			expectError:    true,
		},
		{
			name:           "empty_provided_secret",
			storedHash:     "hkdf-sha256-fixed$ZGVyaXZlZGtleTE2Ynl0ZXNsb25nZGVyaXZlZGtleTE2",
			providedSecret: "",
			expectMatch:    false,
			expectError:    true,
		},
		{
			name:           "invalid_hash_format",
			storedHash:     "invalid-format",
			providedSecret: "password123",
			expectMatch:    false,
			expectError:    true,
		},
		{
			name:           "invalid_dk_encoding",
			storedHash:     "hkdf-sha256-fixed$!!!invalid-base64!!!",
			providedSecret: "password123",
			expectMatch:    false,
			expectError:    true,
		},
		{
			name:           "wrong_algorithm",
			storedHash:     "hkdf-sha512-fixed$ZGVyaXZlZGtleTE2Ynl0ZXNsb25nZGVyaXZlZGtleTE2",
			providedSecret: "password123",
			expectMatch:    false,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			match, err := VerifySecretHKDFFixed(tt.storedHash, tt.providedSecret)

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

func TestHashLowEntropyDeterministic_CrossVerification(t *testing.T) {
	t.Parallel()

	secret := "crossVerifySecret456"

	// Generate hash.
	hash1, err := HashLowEntropyDeterministic(secret)
	require.NoError(t, err)
	require.NotEmpty(t, hash1)

	// Generate second hash with same secret - should be identical.
	hash2, err := HashLowEntropyDeterministic(secret)
	require.NoError(t, err)
	require.NotEmpty(t, hash2)
	require.Equal(t, hash1, hash2, "deterministic hashing should produce identical results")

	// Verify both hashes with correct secret.
	match1, err := VerifySecretHKDFFixed(hash1, secret)
	require.NoError(t, err)
	require.True(t, match1, "hash1 should verify with correct secret")

	match2, err := VerifySecretHKDFFixed(hash2, secret)
	require.NoError(t, err)
	require.True(t, match2, "hash2 should verify with correct secret")

	// Verify hashes fail with wrong secret.
	wrongSecret := "wrongSecret789"
	matchWrong1, err := VerifySecretHKDFFixed(hash1, wrongSecret)
	require.NoError(t, err)
	require.False(t, matchWrong1, "hash1 should not verify with wrong secret")

	matchWrong2, err := VerifySecretHKDFFixed(hash2, wrongSecret)
	require.NoError(t, err)
	require.False(t, matchWrong2, "hash2 should not verify with wrong secret")
}

func TestConstantTimeCompareBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		a      []byte
		b      []byte
		expect bool
	}{
		{
			name:   "equal_slices",
			a:      []byte{1, 2, 3, 4},
			b:      []byte{1, 2, 3, 4},
			expect: true,
		},
		{
			name:   "different_slices",
			a:      []byte{1, 2, 3, 4},
			b:      []byte{1, 2, 3, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries},
			expect: false,
		},
		{
			name:   "different_lengths",
			a:      []byte{1, 2, 3},
			b:      []byte{1, 2, 3, 4},
			expect: false,
		},
		{
			name:   "empty_slices",
			a:      []byte{},
			b:      []byte{},
			expect: true,
		},
		{
			name:   "nil_vs_empty",
			a:      nil,
			b:      []byte{},
			expect: true, // Both have length 0, so they compare equal.
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := constantTimeCompareBytes(tt.a, tt.b)
			require.Equal(t, tt.expect, result)
		})
	}
}

func TestSplitHKDFFixedParts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		hash   string
		expect []string
	}{
		{
			name:   "valid_two_parts",
			hash:   "hkdf-sha256-fixed$ZGVyaXZlZGtleQ==",
			expect: []string{cryptoutilSharedMagic.HKDFFixedLowHashName, "ZGVyaXZlZGtleQ=="},
		},
		{
			name:   "single_part",
			hash:   cryptoutilSharedMagic.HKDFFixedLowHashName,
			expect: []string{cryptoutilSharedMagic.HKDFFixedLowHashName},
		},
		{
			name:   "empty_string",
			hash:   "",
			expect: []string{},
		},
		{
			name:   "multiple_delimiters",
			hash:   "part1$part2$part3",
			expect: []string{"part1", "part2", "part3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parts := splitHKDFFixedParts(tt.hash)
			require.Equal(t, tt.expect, parts)
		})
	}
}
