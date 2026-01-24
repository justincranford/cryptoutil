// Copyright (c) 2025 Justin Cranford
//
//

package authz_test

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/identity/authz"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// TestGenerateAuthorizationCode_Success validates successful code generation.
func TestGenerateAuthorizationCode_Success(t *testing.T) {
	t.Parallel()

	code, err := cryptoutilIdentityAuthz.GenerateAuthorizationCode()
	require.NoError(t, err, "Code generation should succeed")
	require.NotEmpty(t, code, "Code should not be empty")

	// Validate base64url encoding.
	decoded, err := base64.RawURLEncoding.DecodeString(code)
	require.NoError(t, err, "Code should be valid base64url")
	require.Len(t, decoded, cryptoutilIdentityMagic.DefaultAuthCodeLength, "Decoded code should match expected length")
}

// TestGenerateAuthorizationCode_Uniqueness validates that generated codes are unique.
func TestGenerateAuthorizationCode_Uniqueness(t *testing.T) {
	t.Parallel()

	const iterations = 100

	codes := make(map[string]bool, iterations)

	for range iterations {
		code, err := cryptoutilIdentityAuthz.GenerateAuthorizationCode()
		require.NoError(t, err, "Code generation should succeed")
		require.NotContains(t, codes, code, "Code should be unique")
		codes[code] = true
	}

	require.Len(t, codes, iterations, "All codes should be unique")
}

// TestGenerateAuthorizationCode_Length validates code length consistency.
func TestGenerateAuthorizationCode_Length(t *testing.T) {
	t.Parallel()

	const iterations = 10

	for range iterations {
		code, err := cryptoutilIdentityAuthz.GenerateAuthorizationCode()
		require.NoError(t, err, "Code generation should succeed")

		decoded, err := base64.RawURLEncoding.DecodeString(code)
		require.NoError(t, err, "Code should decode successfully")
		require.Len(t, decoded, cryptoutilIdentityMagic.DefaultAuthCodeLength, "Decoded length should be consistent")
	}
}

// TestGenerateAuthorizationCode_Format validates base64url encoding format.
func TestGenerateAuthorizationCode_Format(t *testing.T) {
	t.Parallel()

	code, err := cryptoutilIdentityAuthz.GenerateAuthorizationCode()
	require.NoError(t, err, "Code generation should succeed")

	// base64url should not contain + / or = characters.
	require.NotContains(t, code, "+", "Code should not contain + character")
	require.NotContains(t, code, "/", "Code should not contain / character")
	require.NotContains(t, code, "=", "Code should not contain = padding")
}
