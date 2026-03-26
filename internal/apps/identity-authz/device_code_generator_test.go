// Copyright (c) 2025 Justin Cranford
//
//

package authz

import (
	"encoding/base64"
	"regexp"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// TestGenerateDeviceCode validates device code generation.
func TestGenerateDeviceCode(t *testing.T) {
	t.Parallel()

	// Generate multiple device codes and check properties.
	codes := make(map[string]bool)

	for i := 0; i < cryptoutilSharedMagic.JoseJAMaxMaterials; i++ {
		code, err := GenerateDeviceCode()
		require.NoError(t, err, "Failed to generate device code")
		require.NotEmpty(t, code, "Device code should not be empty")

		// Check uniqueness.
		require.False(t, codes[code], "Duplicate device code generated: %s", code)
		codes[code] = true

		// Check base64url encoding validity.
		decoded, err := base64.RawURLEncoding.DecodeString(code)
		require.NoError(t, err, "Device code should be valid base64url: %s", code)
		require.Len(t, decoded, cryptoutilSharedMagic.DefaultDeviceCodeLength, "Decoded device code should be 32 bytes")

		// Check length (base64url encoding produces ~43 characters for 32 bytes).
		require.GreaterOrEqual(t, len(code), 40, "Device code should be at least 40 characters")
		require.LessOrEqual(t, len(code), cryptoutilSharedMagic.IMMaxUsernameLength, "Device code should be at most 50 characters")
	}
}

// TestGenerateDeviceCode_EntropyCheck validates sufficient entropy.
func TestGenerateDeviceCode_EntropyCheck(t *testing.T) {
	t.Parallel()

	const sampleSize = 1000

	codes := make(map[string]bool, sampleSize)

	for i := 0; i < sampleSize; i++ {
		code, err := GenerateDeviceCode()
		require.NoError(t, err)

		codes[code] = true
	}

	// Expect 100% uniqueness (no collisions in 1000 samples for 256-bit entropy).
	require.Len(t, codes, sampleSize, "All device codes should be unique")
}

// TestGenerateUserCode validates user code generation.
func TestGenerateUserCode(t *testing.T) {
	t.Parallel()

	// Generate multiple user codes and check properties.
	codes := make(map[string]bool)

	for i := 0; i < cryptoutilSharedMagic.JoseJAMaxMaterials; i++ {
		code, err := GenerateUserCode()
		require.NoError(t, err, "Failed to generate user code")
		require.NotEmpty(t, code, "User code should not be empty")

		// Check uniqueness.
		require.False(t, codes[code], "Duplicate user code generated: %s", code)
		codes[code] = true

		// Check format: XXXX-YYYY (4 chars - 4 chars).
		require.Len(t, code, 9, "User code should be 9 characters (XXXX-YYYY)")
		require.Equal(t, "-", string(code[4]), "User code should have hyphen at position 4")

		// Check character set (uppercase alphanumeric, no 0/O/I/1/L).
		validPattern := regexp.MustCompile(`^[ABCDEFGHJKLMNPQRSTUVWXYZ23456789]{4}-[ABCDEFGHJKLMNPQRSTUVWXYZ23456789]{4}$`)
		require.True(t, validPattern.MatchString(code), "User code should match pattern XXXX-YYYY with valid charset: %s", code)

		// Ensure no ambiguous characters.
		require.NotContains(t, code, "0", "User code should not contain '0'")
		require.NotContains(t, code, "O", "User code should not contain 'O'")
		require.NotContains(t, code, "I", "User code should not contain 'I'")
		require.NotContains(t, code, "1", "User code should not contain '1'")
		require.NotContains(t, code, "L", "User code should not contain 'L'")
	}
}

// TestGenerateUserCode_EntropyCheck validates sufficient entropy.
func TestGenerateUserCode_EntropyCheck(t *testing.T) {
	t.Parallel()

	const sampleSize = 1000

	codes := make(map[string]bool, sampleSize)

	for i := 0; i < sampleSize; i++ {
		code, err := GenerateUserCode()
		require.NoError(t, err)

		codes[code] = true
	}

	// Expect high uniqueness (charset=32, length=8 -> 32^8 = ~1.2 trillion combinations).
	// For 1000 samples, expect near 100% uniqueness (collisions extremely rare).
	uniquenessRate := float64(len(codes)) / float64(sampleSize)
	require.GreaterOrEqual(t, uniquenessRate, 0.99, "User codes should have >99%% uniqueness in 1000 samples")
}

// TestGenerateUserCode_Format validates consistent formatting.
func TestGenerateUserCode_Format(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		runs int
	}{
		{"single generation", 1},
		{"multiple generations", cryptoutilSharedMagic.IMMaxUsernameLength},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			for i := 0; i < tc.runs; i++ {
				code, err := GenerateUserCode()
				require.NoError(t, err)

				// Validate format components.
				require.Len(t, code, 9, "User code length should be 9")
				require.Equal(t, "-", string(code[4]), "Hyphen position incorrect")

				// Extract segments.
				segment1 := code[:4]
				segment2 := code[cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries:]

				require.Len(t, segment1, 4, "First segment should be 4 characters")
				require.Len(t, segment2, 4, "Second segment should be 4 characters")

				// Each segment should contain only valid charset.
				for _, ch := range segment1 {
					require.Contains(t, cryptoutilSharedMagic.RecoveryCodeCharset, string(ch), "Invalid character in segment 1: %c", ch)
				}

				for _, ch := range segment2 {
					require.Contains(t, cryptoutilSharedMagic.RecoveryCodeCharset, string(ch), "Invalid character in segment 2: %c", ch)
				}
			}
		})
	}
}

// TestGenerateUserCode_NoAmbiguousCharacters validates exclusion of ambiguous characters.
func TestGenerateUserCode_NoAmbiguousCharacters(t *testing.T) {
	t.Parallel()

	const sampleSize = 500

	ambiguousChars := []string{"0", "O", "I", "1", "L"}

	for i := 0; i < sampleSize; i++ {
		code, err := GenerateUserCode()
		require.NoError(t, err)

		for _, ambiguous := range ambiguousChars {
			require.NotContains(t, code, ambiguous, "User code contains ambiguous character '%s': %s", ambiguous, code)
		}
	}
}
