// Copyright (c) 2025 Iwan van der Kleijn
// SPDX-License-Identifier: MIT

package mfa_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityMFA "cryptoutil/internal/identity/mfa"
)

func TestGenerateRecoveryCode_Format(t *testing.T) {
	t.Parallel()

	code, err := cryptoutilIdentityMFA.GenerateRecoveryCode()
	require.NoError(t, err)

	// Verify format: XXXX-XXXX-XXXX-XXXX (4 groups of 4 chars, separated by hyphens).
	pattern := `^[ABCDEFGHJKLMNPQRSTUVWXYZ23456789]{4}-[ABCDEFGHJKLMNPQRSTUVWXYZ23456789]{4}-[ABCDEFGHJKLMNPQRSTUVWXYZ23456789]{4}-[ABCDEFGHJKLMNPQRSTUVWXYZ23456789]{4}$`
	matched, err := regexp.MatchString(pattern, code)
	require.NoError(t, err)
	require.True(t, matched, "code %q does not match expected format", code)
}

func TestGenerateRecoveryCode_Length(t *testing.T) {
	t.Parallel()

	code, err := cryptoutilIdentityMFA.GenerateRecoveryCode()
	require.NoError(t, err)

	// Expected length: 16 chars + 3 hyphens = 19 characters.
	require.Len(t, code, 19, "recovery code should be 19 characters (16 chars + 3 hyphens)")
}

func TestGenerateRecoveryCode_Uniqueness(t *testing.T) {
	t.Parallel()

	const sampleSize = 1000
	seen := make(map[string]bool, sampleSize)

	for range sampleSize {
		code, err := cryptoutilIdentityMFA.GenerateRecoveryCode()
		require.NoError(t, err)
		require.False(t, seen[code], "duplicate code detected: %q", code)
		seen[code] = true
	}

	require.Len(t, seen, sampleSize, "should generate %d unique codes", sampleSize)
}

func TestGenerateRecoveryCodes_Batch(t *testing.T) {
	t.Parallel()

	const count = 10
	codes, err := cryptoutilIdentityMFA.GenerateRecoveryCodes(count)
	require.NoError(t, err)
	require.Len(t, codes, count, "should generate %d codes", count)

	// Verify all codes are unique.
	seen := make(map[string]bool, count)
	for _, code := range codes {
		require.False(t, seen[code], "duplicate code detected in batch: %q", code)
		seen[code] = true

		// Verify each code matches format.
		pattern := `^[ABCDEFGHJKLMNPQRSTUVWXYZ23456789]{4}-[ABCDEFGHJKLMNPQRSTUVWXYZ23456789]{4}-[ABCDEFGHJKLMNPQRSTUVWXYZ23456789]{4}-[ABCDEFGHJKLMNPQRSTUVWXYZ23456789]{4}$`
		matched, err := regexp.MatchString(pattern, code)
		require.NoError(t, err)
		require.True(t, matched, "code %q does not match expected format", code)
	}
}
