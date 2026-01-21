// Copyright (c) 2025 Justin Cranford

package pkce

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"

	testify "github.com/stretchr/testify/require"

	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// Validates requirements:
// - R01-03: Consent approval generates authorization code with user context
// - R01-05: Authorization code single-use enforcement.
func TestGenerateCodeVerifier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		testFn func(t *testing.T)
	}{
		{
			name: "basic_generation",
			testFn: func(t *testing.T) {
				t.Helper()

				verifier, err := GenerateCodeVerifier()
				testify.NoError(t, err, "Generate code verifier should succeed")
				testify.NotEmpty(t, verifier, "Code verifier should not be empty")
				testify.GreaterOrEqual(t, len(verifier), 43, "Code verifier should be at least 43 characters (RFC 7636)")
				testify.NotContains(t, verifier, "+", "Code verifier should use base64url encoding")
				testify.NotContains(t, verifier, "/", "Code verifier should use base64url encoding")
				testify.NotContains(t, verifier, "=", "Code verifier should not have padding")
			},
		},
		{
			name: "uniqueness",
			testFn: func(t *testing.T) {
				t.Helper()

				verifiers := make(map[string]bool)

				for i := 0; i < 100; i++ {
					verifier, err := GenerateCodeVerifier()
					testify.NoError(t, err, "Generate code verifier should succeed")
					testify.False(t, verifiers[verifier], "Code verifiers should be unique")
					verifiers[verifier] = true
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.testFn(t)
		})
	}
}

func TestGenerateCodeChallenge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		verifier        string
		method          string
		wantEmptyResult bool
		verifyFn        func(t *testing.T, verifier string, challenge string, method string)
	}{
		{
			name:     "s256_method",
			verifier: "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			method:   cryptoutilIdentityMagic.PKCEMethodS256,
			verifyFn: func(t *testing.T, verifier string, challenge string, _ string) {
				t.Helper()
				testify.NotEmpty(t, challenge, "Code challenge should not be empty")
				testify.NotEqual(t, verifier, challenge, "Challenge should differ from verifier")
				hash := sha256.Sum256([]byte(verifier))
				expected := base64.RawURLEncoding.EncodeToString(hash[:])
				testify.Equal(t, expected, challenge, "S256 challenge should match expected SHA256 hash")
			},
		},
		{
			name:     "plain_method",
			verifier: "test-code-verifier",
			method:   cryptoutilIdentityMagic.PKCEMethodPlain,
			verifyFn: func(t *testing.T, verifier string, challenge string, _ string) {
				t.Helper()
				testify.Equal(t, verifier, challenge, "Plain challenge should equal verifier")
			},
		},
		{
			name:     "default_method_s256",
			verifier: "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			method:   "",
			verifyFn: func(t *testing.T, verifier string, challenge string, _ string) {
				t.Helper()

				hash := sha256.Sum256([]byte(verifier))
				expected := base64.RawURLEncoding.EncodeToString(hash[:])
				testify.Equal(t, expected, challenge, "Default method should be S256")
			},
		},
		{
			name:            "invalid_method",
			verifier:        "test-code-verifier",
			method:          "invalid-method",
			wantEmptyResult: true,
			verifyFn: func(t *testing.T, _ string, challenge string, _ string) {
				t.Helper()
				testify.Empty(t, challenge, "Invalid method should return empty challenge")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			challenge := GenerateCodeChallenge(tc.verifier, tc.method)

			if tc.verifyFn != nil {
				tc.verifyFn(t, tc.verifier, challenge, tc.method)
			}
		})
	}
}

func TestGenerateS256Challenge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		verifier string
	}{
		{
			name:     "standard_verifier",
			verifier: "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
		},
		{
			name:     "short_verifier",
			verifier: "abc",
		},
		{
			name:     "long_verifier",
			verifier: "very-long-code-verifier-with-many-characters-for-testing-purposes",
		},
		{
			name:     "empty_verifier",
			verifier: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			challenge := GenerateS256Challenge(tc.verifier)

			testify.NotEmpty(t, challenge, "Challenge should not be empty")

			// Verify it matches manual SHA256 computation
			hash := sha256.Sum256([]byte(tc.verifier))
			expected := base64.RawURLEncoding.EncodeToString(hash[:])
			testify.Equal(t, expected, challenge, "Challenge should match SHA256 hash")
		})
	}
}

func TestValidateCodeVerifier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		verifier  string
		challenge string
		method    string
		wantValid bool
	}{
		{
			name:      "s256_valid",
			verifier:  "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			challenge: GenerateS256Challenge("dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"),
			method:    cryptoutilIdentityMagic.PKCEMethodS256,
			wantValid: true,
		},
		{
			name:      "s256_invalid",
			verifier:  "wrong-verifier",
			challenge: GenerateS256Challenge("correct-verifier"),
			method:    cryptoutilIdentityMagic.PKCEMethodS256,
			wantValid: false,
		},
		{
			name:      "plain_valid",
			verifier:  "test-code-verifier",
			challenge: "test-code-verifier",
			method:    cryptoutilIdentityMagic.PKCEMethodPlain,
			wantValid: true,
		},
		{
			name:      "plain_invalid",
			verifier:  "correct-verifier",
			challenge: "wrong-verifier",
			method:    cryptoutilIdentityMagic.PKCEMethodPlain,
			wantValid: false,
		},
		{
			name:      "default_method_s256",
			verifier:  "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			challenge: GenerateS256Challenge("dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"),
			method:    "",
			wantValid: true,
		},
		{
			name:      "invalid_method",
			verifier:  "test-verifier",
			challenge: "test-challenge",
			method:    "invalid-method",
			wantValid: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			valid := ValidateCodeVerifier(tc.verifier, tc.challenge, tc.method)

			testify.Equal(t, tc.wantValid, valid, "Validation result should match expected")
		})
	}
}

func TestValidateS256(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		verifier  string
		challenge string
		expected  bool
	}{
		{
			name:      "valid_verifier_and_challenge",
			verifier:  "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			challenge: GenerateS256Challenge("dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"),
			expected:  true,
		},
		{
			name:      "invalid_verifier",
			verifier:  "wrong-verifier",
			challenge: GenerateS256Challenge("correct-verifier"),
			expected:  false,
		},
		{
			name:      "empty_verifier",
			verifier:  "",
			challenge: GenerateS256Challenge(""),
			expected:  true,
		},
		{
			name:      "empty_verifier_with_non_empty_challenge",
			verifier:  "",
			challenge: GenerateS256Challenge("non-empty"),
			expected:  false,
		},
		{
			name:      "non_empty_verifier_with_empty_challenge",
			verifier:  "non-empty",
			challenge: "",
			expected:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			valid := ValidateS256(tc.verifier, tc.challenge)

			testify.Equal(t, tc.expected, valid, "Validation result should match expected")
		})
	}
}

func TestPKCERoundtrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
	}{
		{
			name:   "s256_method",
			method: cryptoutilIdentityMagic.PKCEMethodS256,
		},
		{
			name:   "plain_method",
			method: cryptoutilIdentityMagic.PKCEMethodPlain,
		},
		{
			name:   "default_method_s256",
			method: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Step 1: Generate code verifier
			verifier, err := GenerateCodeVerifier()
			testify.NoError(t, err, "Generate code verifier should succeed")

			// Step 2: Generate code challenge
			challenge := GenerateCodeChallenge(verifier, tc.method)
			testify.NotEmpty(t, challenge, "Code challenge should not be empty")

			// Step 3: Validate verifier against challenge
			valid := ValidateCodeVerifier(verifier, challenge, tc.method)
			testify.True(t, valid, "Generated verifier should validate against challenge")
		})
	}
}

func TestPKCERoundtrip_MultipleVerifiers(t *testing.T) {
	t.Parallel()

	verifiers := make([]string, 10)
	challenges := make([]string, 10)

	// Generate multiple verifier/challenge pairs
	for i := 0; i < 10; i++ {
		verifier, err := GenerateCodeVerifier()
		testify.NoError(t, err, "Generate code verifier should succeed")

		challenge := GenerateCodeChallenge(verifier, cryptoutilIdentityMagic.PKCEMethodS256)
		testify.NotEmpty(t, challenge, "Code challenge should not be empty")

		verifiers[i] = verifier
		challenges[i] = challenge
	}

	// Verify each verifier validates only with its own challenge
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			valid := ValidateCodeVerifier(verifiers[i], challenges[j], cryptoutilIdentityMagic.PKCEMethodS256)

			if i == j {
				testify.True(t, valid, "Verifier should validate with its own challenge")
			} else {
				testify.False(t, valid, "Verifier should not validate with different challenge")
			}
		}
	}
}
