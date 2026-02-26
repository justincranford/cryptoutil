// Copyright (c) 2025 Justin Cranford

package hash

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"strings"
	"testing"

	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"

	"github.com/stretchr/testify/require"
)

type highEntropyTest struct {
	name        string
	input       string
	expectError bool
}

// randomHighEntropyValue generates a random high-entropy value for testing (simulates API keys, tokens).
func randomHighEntropyValue(t *testing.T, length int) string {
	t.Helper()

	value, err := cryptoutilSharedUtilRandom.GenerateString(length)
	require.NoError(t, err)

	return value
}

func highEntropyTests(t *testing.T) []highEntropyTest {
	t.Helper()

	randomValue1, err := cryptoutilSharedUtilRandom.GenerateString(cryptoutilSharedMagic.RealmMinTokenLengthBytes)
	require.NoError(t, err)

	randomValue2, err := cryptoutilSharedUtilRandom.GenerateString(cryptoutilSharedMagic.MaxUnsealSharedSecrets)
	require.NoError(t, err)

	return []highEntropyTest{
		{
			name:        "empty",
			input:       "",
			expectError: true,
		},
		{
			name:        "short",
			input:       randomValue1,
			expectError: false,
		},
		{
			name:        "long",
			input:       randomValue2,
			expectError: false,
		},
	}
}

func TestHashHighEntropyDeterministic(t *testing.T) {
	t.Parallel()

	for _, tt := range highEntropyTests(t) {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			hash, err := HashHighEntropyDeterministic(tt.input)

			if tt.expectError {
				require.Error(t, err)
				require.Empty(t, hash)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, hash)

				// Verify format: hkdf-sha256-fixed-high$base64(dk)
				parts := strings.Split(hash, "$")
				require.Len(t, parts, 2, "hash should have 2 parts")
				require.Equal(t, cryptoutilSharedMagic.HKDFFixedHighHashName, parts[0])
				require.NotEmpty(t, parts[1], "derived key should not be empty")
			}
		})
	}
}

func TestHashSecretHKDFFixedHigh(t *testing.T) {
	t.Parallel()

	for _, tt := range highEntropyTests(t) {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fixedInfo := []byte("test-fixed-info-high")
			hash, err := HashSecretHKDFFixedHigh(tt.input, fixedInfo)

			if tt.expectError {
				require.Error(t, err)
				require.Empty(t, hash)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, hash)

				// Verify format.
				parts := strings.Split(hash, "$")
				require.Len(t, parts, 2)
				require.Equal(t, cryptoutilSharedMagic.HKDFFixedHighHashName, parts[0])
			}
		})
	}
}

func TestHashSecretHKDFFixedHigh_Determinism(t *testing.T) {
	t.Parallel()

	// Generate high-entropy test value (simulates API key/token)
	value, err := cryptoutilSharedUtilRandom.GenerateString(cryptoutilSharedMagic.UUIDStringLength)
	require.NoError(t, err)

	fixedInfo := []byte("deterministic-info-high")

	const iterations = 10

	hashes := make([]string, iterations)

	// Generate multiple hashes with same value and fixed info.
	for i := 0; i < iterations; i++ {
		hash, err := HashSecretHKDFFixedHigh(value, fixedInfo)
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
		input          string
		expectedOutput string
		expectMatch    bool
		expectError    bool
	}{
		{
			name:           "valid_hash_matches",
			expectedOutput: "hkdf-sha256-fixed-high$ZGVyaXZlZGtleTE2Ynl0ZXNsb25nZGVyaXZlZGtleTE2",
			input:          randomHighEntropyValue(t, cryptoutilSharedMagic.UUIDStringLength),
			expectMatch:    false, // Won't match unless we use the exact secret that generated this hash.
			expectError:    false,
		},
		{
			name:           "empty_stored_hash",
			expectedOutput: "",
			input:          randomHighEntropyValue(t, cryptoutilSharedMagic.UUIDStringLength),
			expectMatch:    false,
			expectError:    true,
		},
		{
			name:           "empty_provided_secret",
			expectedOutput: "hkdf-sha256-fixed-high$ZGVyaXZlZGtleTE2Ynl0ZXNsb25nZGVyaXZlZGtleTE2", // pragma: allowlist secret
			input:          "",
			expectMatch:    false,
			expectError:    true,
		},
		{
			name:           "invalid_hash_format",
			expectedOutput: "invalid-format",
			input:          randomHighEntropyValue(t, cryptoutilSharedMagic.UUIDStringLength),
			expectMatch:    false,
			expectError:    true,
		},
		{
			name:           "invalid_dk_encoding",
			expectedOutput: "hkdf-sha256-fixed-high$!!!invalid-base64!!!",
			input:          randomHighEntropyValue(t, cryptoutilSharedMagic.UUIDStringLength),
			expectMatch:    false,
			expectError:    true,
		},
		{
			name:           "wrong_algorithm",
			expectedOutput: "hkdf-sha512-fixed-high$ZGVyaXZlZGtleTE2Ynl0ZXNsb25nZGVyaXZlZGtleTE2",
			input:          randomHighEntropyValue(t, cryptoutilSharedMagic.UUIDStringLength),
			expectMatch:    false,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			match, err := VerifySecretHKDFFixedHigh(tt.expectedOutput, tt.input)

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

	// Generate high-entropy test value (simulates API key/token)
	secret, err := cryptoutilSharedUtilRandom.GenerateString(cryptoutilSharedMagic.UUIDStringLength)
	require.NoError(t, err)

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
	wrongSecret, err := cryptoutilSharedUtilRandom.GenerateString(cryptoutilSharedMagic.UUIDStringLength)
	require.NoError(t, err)

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
			expect: []string{cryptoutilSharedMagic.HKDFFixedHighHashName, "ZGVyaXZlZGtleQ=="},
		},
		{
			name:   "single_part",
			hash:   cryptoutilSharedMagic.HKDFFixedHighHashName,
			expect: []string{cryptoutilSharedMagic.HKDFFixedHighHashName},
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
