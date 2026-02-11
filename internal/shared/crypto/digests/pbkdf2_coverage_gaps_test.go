// Copyright (c) 2025 Justin Cranford

package digests

import (
	sha256 "crypto/sha256"
	"hash"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestPBKDF2WithParams_ErrorPaths tests error handling for PBKDF2WithParams.
func TestPBKDF2WithParams_ErrorPaths(t *testing.T) {
	t.Parallel()

	validParams := &PBKDF2Params{
		Version:    "1",
		HashName:   "pbkdf2-sha256",
		Iterations: 600000,
		SaltLength: 32,
		KeyLength:  32,
		HashFunc:   func() hash.Hash { return sha256.New() }, // Required for PBKDF2
	}

	tests := []struct {
		name          string
		secret        string
		params        *PBKDF2Params
		expectedError bool
		errorContains string
	}{
		{
			name:          "empty_secret",
			secret:        "",
			params:        validParams,
			expectedError: true,
			errorContains: "secret is empty",
		},
		{
			name:          "nil_params",
			secret:        "test_secret",
			params:        nil,
			expectedError: true,
			errorContains: "parameter set is nil",
		},
		{
			name:          "valid_short_secret",
			secret:        "abc",
			params:        validParams,
			expectedError: false,
		},
		{
			name:          "valid_long_secret",
			secret:        strings.Repeat("long", 100), // 400 chars
			params:        validParams,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			hash, err := PBKDF2WithParams(tt.secret, tt.params)

			if tt.expectedError {
				require.Error(t, err, "Expected error for %s", tt.name)

				if tt.errorContains != "" {
					require.Contains(t, err.Error(), tt.errorContains, "Error message should contain expected substring")
				}

				require.Empty(t, hash, "Hash should be empty when error occurs")
			} else {
				require.NoError(t, err, "Expected no error for %s", tt.name)
				require.NotEmpty(t, hash, "Hash should not be empty")
				require.Contains(t, hash, "{1}$pbkdf2-sha256$", "Hash should have correct format")
			}
		})
	}
}

// TestVerifySecret_ErrorPaths tests error handling for VerifySecret.
func TestVerifySecret_ErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		stored        string
		provided      string
		expectedValid bool
		expectedError bool
		errorContains string
	}{
		{
			name:          "empty_stored_hash",
			stored:        "",
			provided:      "test_secret",
			expectedValid: false,
			expectedError: true,
			errorContains: "stored hash empty",
		},
		{
			name:          "non_versioned_format",
			stored:        "pbkdf2-sha256$600000$salt$dk",
			provided:      "test_secret",
			expectedValid: false,
			expectedError: true,
			errorContains: "unsupported hash format",
		},
		{
			name:          "invalid_parts_count",
			stored:        "{1}$pbkdf2-sha256$600000",
			provided:      "test_secret",
			expectedValid: false,
			expectedError: true,
			errorContains: "invalid versioned hash format",
		},
		{
			name:          "malformed_version",
			stored:        "1$pbkdf2-sha256$600000$c2FsdA$ZGVyaXZlZA",
			provided:      "test_secret",
			expectedValid: false,
			expectedError: true,
			errorContains: "unsupported hash format",
		},
		{
			name:          "invalid_iterations",
			stored:        "{1}$pbkdf2-sha256$abc$c2FsdA$ZGVyaXZlZA",
			provided:      "test_secret",
			expectedValid: false,
			expectedError: true,
			errorContains: "invalid iterations",
		},
		{
			name:          "negative_iterations",
			stored:        "{1}$pbkdf2-sha256$-1000$c2FsdA$ZGVyaXZlZA",
			provided:      "test_secret",
			expectedValid: false,
			expectedError: true,
			errorContains: "invalid iterations",
		},
		{
			name:          "malformed_base64_salt",
			stored:        "{1}$pbkdf2-sha256$600000$!!!invalid!!!$ZGVyaXZlZA",
			provided:      "test_secret",
			expectedValid: false,
			expectedError: true,
			errorContains: "invalid salt encoding",
		},
		{
			name:          "malformed_base64_dk",
			stored:        "{1}$pbkdf2-sha256$600000$c2FsdA$!!!invalid!!!",
			provided:      "test_secret",
			expectedValid: false,
			expectedError: true,
			errorContains: "invalid dk encoding",
		},
		{
			name:          "unsupported_hash_algorithm",
			stored:        "{1}$pbkdf2-md5$600000$c2FsdA$ZGVyaXZlZA",
			provided:      "test_secret",
			expectedValid: false,
			expectedError: true,
			errorContains: "unsupported hash algorithm",
		},
		{
			name:          "sha384_valid_format",
			stored:        "{1}$pbkdf2-sha384$600000$VGVzdFNhbHQxMjM0NTY3ODkwMTIzNDU2Nzg5MDEy$dGVzdF9kZXJpdmVkX2tleV80OF9ieXRlc19sb25nX2Zvcl9zaGEzODQ",
			provided:      "test_secret",
			expectedValid: false, // Will fail hash comparison (fake hash)
			expectedError: false,
		},
		{
			name:          "sha512_valid_format",
			stored:        "{1}$pbkdf2-sha512$600000$VGVzdFNhbHQxMjM0NTY3ODkwMTIzNDU2Nzg5MDEy$dGVzdF9kZXJpdmVkX2tleV82NF9ieXRlc19sb25nX2Zvcl9zaGE1MTJfaGFzaF90ZXN0",
			provided:      "test_secret",
			expectedValid: false, // Will fail hash comparison (fake hash)
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			valid, err := VerifySecret(tt.stored, tt.provided)

			if tt.expectedError {
				require.Error(t, err, "Expected error for %s", tt.name)

				if tt.errorContains != "" {
					require.Contains(t, err.Error(), tt.errorContains, "Error message should contain expected substring")
				}

				require.False(t, valid, "Valid should be false when error occurs")
			} else {
				require.NoError(t, err, "Expected no error for %s", tt.name)
				require.Equal(t, tt.expectedValid, valid, "Valid flag should match expected")
			}
		})
	}
}

// TestParsePbkdf2Params_ErrorPaths tests error handling for parsePbkdf2Params (indirectly via VerifySecret).
// This test is redundant with TestVerifySecret_ErrorPaths but provides explicit coverage.
func TestParsePbkdf2Params_CoverageCheck(t *testing.T) {
	t.Parallel()

	// These cases exercise parsePbkdf2Params branches not directly tested elsewhere
	tests := []struct {
		name          string
		stored        string
		errorContains string
	}{
		{
			name:          "version_without_closing_brace",
			stored:        "{1$pbkdf2-sha256$600000$c2FsdA$ZGVyaXZlZA",
			errorContains: "invalid version format",
		},
		{
			name:          "version_without_opening_brace",
			stored:        "1}$pbkdf2-sha256$600000$c2FsdA$ZGVyaXZlZA",
			errorContains: "unsupported hash format",
		},
		{
			name:          "zero_iterations",
			stored:        "{1}$pbkdf2-sha256$0$c2FsdA$ZGVyaXZlZA",
			errorContains: "invalid iterations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			valid, err := VerifySecret(tt.stored, "any_secret")
			require.Error(t, err, "Expected error for %s", tt.name)
			require.Contains(t, err.Error(), tt.errorContains, "Error message should contain expected substring")
			require.False(t, valid, "Valid should be false when error occurs")
		})
	}
}
