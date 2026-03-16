// Copyright (c) 2025 Justin Cranford

package mfa_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityMfa "cryptoutil/internal/apps/identity/mfa"
)

const recoveryCodePattern = `^[ABCDEFGHJKLMNPQRSTUVWXYZ23456789]{4}-[ABCDEFGHJKLMNPQRSTUVWXYZ23456789]{4}-[ABCDEFGHJKLMNPQRSTUVWXYZ23456789]{4}-[ABCDEFGHJKLMNPQRSTUVWXYZ23456789]{4}$`

func TestGenerateRecoveryCode_Format(t *testing.T) {
	t.Parallel()

	code, err := cryptoutilIdentityMfa.GenerateRecoveryCode()
	require.NoError(t, err)

	// Verify format: XXXX-XXXX-XXXX-XXXX (4 groups of 4 chars, separated by hyphens).
	matched, err := regexp.MatchString(recoveryCodePattern, code)
	require.NoError(t, err)
	require.True(t, matched, "code %q does not match expected format", code)
}

func TestGenerateRecoveryCode_Length(t *testing.T) {
	t.Parallel()

	code, err := cryptoutilIdentityMfa.GenerateRecoveryCode()
	require.NoError(t, err)

	// Expected length: 16 chars + 3 hyphens = 19 characters.
	require.Len(t, code, 19, "recovery code should be 19 characters (16 chars + 3 hyphens)")
}

func TestGenerateRecoveryCode_Uniqueness(t *testing.T) {
	t.Parallel()

	const sampleSize = 1000

	seen := make(map[string]bool, sampleSize)

	for range sampleSize {
		code, err := cryptoutilIdentityMfa.GenerateRecoveryCode()
		require.NoError(t, err)
		require.False(t, seen[code], "duplicate code detected: %q", code)
		seen[code] = true
	}

	require.Len(t, seen, sampleSize, "should generate %d unique codes", sampleSize)
}

func TestGenerateRecoveryCodes_Batch(t *testing.T) {
	t.Parallel()

	const count = 10

	codes, err := cryptoutilIdentityMfa.GenerateRecoveryCodes(count)
	require.NoError(t, err)
	require.Len(t, codes, count, "should generate %d codes", count)

	// Verify all codes are unique.
	seen := make(map[string]bool, count)
	// Compile regex once for better performance (avoids SA6000 staticcheck warning).
	re := regexp.MustCompile(recoveryCodePattern)

	for _, code := range codes {
		require.False(t, seen[code], "duplicate code detected in batch: %q", code)
		seen[code] = true

		// Verify each code matches format.
		require.True(t, re.MatchString(code), "code %q does not match expected format", code)
	}
}
