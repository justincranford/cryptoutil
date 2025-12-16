// Copyright (c) 2025 Justin Cranford

package hash

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashHighEntropyDeterministic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		secret      string
		expectError bool
	}{
		{
			name:        "valid_high_entropy_secret",
			secret:      "sk_live_51A2b3C4d5E6f7G8h9I0J1K2L3M4N5O6",
			expectError: false,
		},
		{
			name:        "empty_secret",
			secret:      "",
			expectError: true,
		},
		{
			name:        "long_secret",
			secret:      strings.Repeat("a", 2048),
			expectError: false,
		},
		{
			name:        "unicode_secret",
			secret:      "令牌🔑密钥",
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

			hash, err := HashHighEntropyDeterministic(tt.secret)

			if tt.expectError {
				require.Error(t, err)
				require.Empty(t, hash)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, hash)

				// Verify format: hkdf-sha256-fixed-high$base64(dk)
				parts := strings.Split(hash, "$")
				require.Len(t, parts, 2, "hash should have 2 parts")
				require.Equal(t, "hkdf-sha256-fixed-high", parts[0])
				require.NotEmpty(t, parts[1], "derived key should not be empty")
			}
		})
	}
}

func TestHashSecretHKDFFixedHigh(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		secret      string
		expectError bool
	}{
		{
			name:        "valid_secret",
			secret:      "sk_test_4eC39HqLyjWDarjtT1zdp7dc",
			expectError: false,
		},
		{
			name:        "empty_secret",
			secret:      "",
			expectError: true,
		},
		{
			name:        "long_secret",
			secret:      strings.Repeat("x", 4096),
			expectError: false,
		},
		{
			name:        "unicode_secret",
			secret:      "令牌🔑密钥",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fixedInfo := []byte("test-fixed-info-high")
			hash, err := HashSecretHKDFFixedHigh(tt.secret, fixedInfo)

			if tt.expectError {
				require.Error(t, err)
				require.Empty(t, hash)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, hash)

				// Verify format.
				parts := strings.Split(hash, "$")
				require.Len(t, parts, 2)
				require.Equal(t, "hkdf-sha256-fixed-high", parts[0])
			}
		})
	}
}

func TestHashSecretHKDFFixedHigh_Determinism(t *testing.T) {
	t.Parallel()

	secret := "sk_live_deterministicTokenABCDEF123456"
	fixedInfo := []byte("deterministic-info-high")

	const iterations = 10
	hashes := make([]string, iterations)

	// Generate multiple hashes with same secret and fixed info.
	for i := 0; i < iterations; i++ {
		hash, err := HashSecretHKDFFixedHigh(secret, fixedInfo)
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

func TestVerifySecretHKDFFixedHigh(t *testing.T) {
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
			storedHash:     "hkdf-sha256-fixed-high$ZGVyaXZlZGtleTE2Ynl0ZXNsb25nZGVyaXZlZGtleTE2",
			providedSecret: "sk_live_51A2b3C4d5E6f7G8h9I0J1K2L3M4N5O6",
			expectMatch:    false, // Won't match unless we use the exact secret that generated this hash.
			expectError:    false,
		},
		{
			name:           "empty_stored_hash",
			storedHash:     "",
			providedSecret: "sk_live_51A2b3C4d5E6f7G8h9I0J1K2L3M4N5O6",
			expectMatch:    false,
			expectError:    true,
		},
		{
			name:           "empty_provided_secret",
			storedHash:     "hkdf-sha256-fixed-high$ZGVyaXZlZGtleTE2Ynl0ZXNsb25nZGVyaXZlZGtleTE2",
			providedSecret: "",
			expectMatch:    false,
			expectError:    true,
		},
		{
			name:           "invalid_hash_format",
			storedHash:     "invalid-format",
			providedSecret: "sk_live_51A2b3C4d5E6f7G8h9I0J1K2L3M4N5O6",
			expectMatch:    false,
			expectError:    true,
		},
		{
			name:           "invalid_dk_encoding",
			storedHash:     "hkdf-sha256-fixed-high$!!!invalid-base64!!!",
			providedSecret: "sk_live_51A2b3C4d5E6f7G8h9I0J1K2L3M4N5O6",
			expectMatch:    false,
			expectError:    true,
		},
		{
			name:           "wrong_algorithm",
			storedHash:     "hkdf-sha512-fixed-high$ZGVyaXZlZGtleTE2Ynl0ZXNsb25nZGVyaXZlZGtleTE2",
			providedSecret: "sk_live_51A2b3C4d5E6f7G8h9I0J1K2L3M4N5O6",
			expectMatch:    false,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			match, err := VerifySecretHKDFFixedHigh(tt.storedHash, tt.providedSecret)

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

func TestHashHighEntropyDeterministic_CrossVerification(t *testing.T) {
	t.Parallel()

	secret := "sk_live_crossVerifyToken123456789ABCDEF"

	// Generate hash.
	hash1, err := HashHighEntropyDeterministic(secret)
	require.NoError(t, err)
	require.NotEmpty(t, hash1)

	// Generate second hash with same secret - should be identical.
	hash2, err := HashHighEntropyDeterministic(secret)
	require.NoError(t, err)
	require.NotEmpty(t, hash2)
	require.Equal(t, hash1, hash2, "deterministic hashing should produce identical results")

	// Verify both hashes with correct secret.
	match1, err := VerifySecretHKDFFixedHigh(hash1, secret)
	require.NoError(t, err)
	require.True(t, match1, "hash1 should verify with correct secret")

	match2, err := VerifySecretHKDFFixedHigh(hash2, secret)
	require.NoError(t, err)
	require.True(t, match2, "hash2 should verify with correct secret")

	// Verify hashes fail with wrong secret.
	wrongSecret := "sk_live_wrongToken999999999XYZABC"
	matchWrong1, err := VerifySecretHKDFFixedHigh(hash1, wrongSecret)
	require.NoError(t, err)
	require.False(t, matchWrong1, "hash1 should not verify with wrong secret")

	matchWrong2, err := VerifySecretHKDFFixedHigh(hash2, wrongSecret)
	require.NoError(t, err)
	require.False(t, matchWrong2, "hash2 should not verify with wrong secret")
}

func TestConstantTimeCompareBytesHigh(t *testing.T) {
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
			b:      []byte{1, 2, 3, 5},
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

			result := constantTimeCompareBytesHigh(tt.a, tt.b)
			require.Equal(t, tt.expect, result)
		})
	}
}

func TestSplitHKDFFixedHighParts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		hash   string
		expect []string
	}{
		{
			name:   "valid_two_parts",
			hash:   "hkdf-sha256-fixed-high$ZGVyaXZlZGtleQ==",
			expect: []string{"hkdf-sha256-fixed-high", "ZGVyaXZlZGtleQ=="},
		},
		{
			name:   "single_part",
			hash:   "hkdf-sha256-fixed-high",
			expect: []string{"hkdf-sha256-fixed-high"},
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

			parts := splitHKDFFixedHighParts(tt.hash)
			require.Equal(t, tt.expect, parts)
		})
	}
}

func TestHashHighEntropyDeterministic_VsLowEntropy(t *testing.T) {
	t.Parallel()

	// Same secret should produce different hashes with low-entropy vs high-entropy fixed info.
	secret := "testSecret123"

	lowEntropyHash, err := HashLowEntropyDeterministic(secret)
	require.NoError(t, err)
	require.NotEmpty(t, lowEntropyHash)
	require.Contains(t, lowEntropyHash, "hkdf-sha256-fixed$")

	highEntropyHash, err := HashHighEntropyDeterministic(secret)
	require.NoError(t, err)
	require.NotEmpty(t, highEntropyHash)
	require.Contains(t, highEntropyHash, "hkdf-sha256-fixed-high$")

	// Hashes should be different due to different fixed info parameters.
	require.NotEqual(t, lowEntropyHash, highEntropyHash,
		"low-entropy and high-entropy hashes should differ due to different fixed info")
}
