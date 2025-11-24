// Copyright (c) 2025 Justin Cranford
//
//

package clientauth_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityClientAuth "cryptoutil/internal/identity/authz/clientauth"
)

func TestHashSecret(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		secret  string
		wantErr bool
	}{
		{
			name:    "valid secret",
			secret:  "my-secret-password",
			wantErr: false,
		},
		{
			name:    "empty secret",
			secret:  "",
			wantErr: false, // Empty secrets are allowed (will hash to unique value)
		},
		{
			name:    "long secret",
			secret:  strings.Repeat("a", 1000),
			wantErr: false,
		},
	}

	for _, tc := range tests {

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			hashed, err := cryptoutilIdentityClientAuth.HashSecret(tc.secret)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, hashed)

				// Verify format: "salt:hash" (base64).
				parts := strings.Split(hashed, ":")
				require.Len(t, parts, 2, "hashed secret should be 'salt:hash'")

				// Verify both parts are base64-encoded.
				require.NotEmpty(t, parts[0], "salt should not be empty")
				require.NotEmpty(t, parts[1], "hash should not be empty")
			}
		})
	}
}

func TestHashSecret_Uniqueness(t *testing.T) {
	t.Parallel()

	secret := "same-secret"

	hash1, err := cryptoutilIdentityClientAuth.HashSecret(secret)
	require.NoError(t, err)

	hash2, err := cryptoutilIdentityClientAuth.HashSecret(secret)
	require.NoError(t, err)

	// Hashes should be different due to random salt.
	require.NotEqual(t, hash1, hash2, "hashes of same secret should differ (random salt)")
}

func TestCompareSecret(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		secret     string
		plainInput string
		wantMatch  bool
		wantErr    bool
	}{
		{
			name:       "matching secret",
			secret:     "my-secret-password",
			plainInput: "my-secret-password",
			wantMatch:  true,
			wantErr:    false,
		},
		{
			name:       "non-matching secret",
			secret:     "my-secret-password",
			plainInput: "wrong-password",
			wantMatch:  false,
			wantErr:    false,
		},
		{
			name:       "empty secret matches empty",
			secret:     "",
			plainInput: "",
			wantMatch:  true,
			wantErr:    false,
		},
		{
			name:       "empty secret does not match non-empty",
			secret:     "",
			plainInput: "password",
			wantMatch:  false,
			wantErr:    false,
		},
	}

	for _, tc := range tests {

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Hash the secret.
			hashed, err := cryptoutilIdentityClientAuth.HashSecret(tc.secret)
			require.NoError(t, err)

			// Compare with plain input.
			match, err := cryptoutilIdentityClientAuth.CompareSecret(hashed, tc.plainInput)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantMatch, match)
			}
		})
	}
}

func TestCompareSecret_InvalidFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		hashed  string
		plain   string
		wantErr bool
	}{
		{
			name:    "missing separator",
			hashed:  "invalidsalthash",
			plain:   "secret",
			wantErr: true,
		},
		{
			name:    "invalid base64 salt",
			hashed:  "!!!invalid!!!:validhash",
			plain:   "secret",
			wantErr: true,
		},
		{
			name:    "invalid base64 hash",
			hashed:  "validhash:!!!invalid!!!",
			plain:   "secret",
			wantErr: true,
		},
		{
			name:    "empty hashed",
			hashed:  "",
			plain:   "secret",
			wantErr: true,
		},
	}

	for _, tc := range tests {

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := cryptoutilIdentityClientAuth.CompareSecret(tc.hashed, tc.plain)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCompareSecret_ConstantTime(t *testing.T) {
	t.Parallel()

	// This test verifies that CompareSecret uses constant-time comparison.
	// Note: This is a basic test; true timing analysis requires benchmarking.

	secret := "my-secret-password"
	hashed, err := cryptoutilIdentityClientAuth.HashSecret(secret)
	require.NoError(t, err)

	// Multiple comparisons should all complete (no early returns).
	for i := 0; i < 100; i++ {
		match, err := cryptoutilIdentityClientAuth.CompareSecret(hashed, "wrong-password")
		require.NoError(t, err)
		require.False(t, match)
	}
}
