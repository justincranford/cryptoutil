// Copyright (c) 2025-2026 Justin Cranford.
package pkce

import (
	sha256 "crypto/sha256"
	"encoding/base64"
	"testing"

	testify "github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
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
				testify.GreaterOrEqual(t, len(verifier), cryptoutilSharedMagic.DefaultCodeChallengeLength, "Code verifier should be at least 43 characters (RFC 7636)")
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

				for i := 0; i < cryptoutilSharedMagic.JoseJAMaxMaterials; i++ {
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
			name:     cryptoutilSharedMagic.S256_METHOD,
			verifier: cryptoutilSharedMagic.S256_CODE_CHALLENGE,
			method:   cryptoutilSharedMagic.PKCEMethodS256,
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
			name:     cryptoutilSharedMagic.PLAIN_METHOD,
			verifier: cryptoutilSharedMagic.TEST_CODE_VERIFIER,
			method:   cryptoutilSharedMagic.PKCEMethodPlain,
			verifyFn: func(t *testing.T, verifier string, challenge string, _ string) {
				t.Helper()
				testify.Equal(t, verifier, challenge, "Plain challenge should equal verifier")
			},
		},
		{
			name:     cryptoutilSharedMagic.DEFAULT_METHOD_S256,
			verifier: cryptoutilSharedMagic.S256_CODE_CHALLENGE,
			method:   "",
			verifyFn: func(t *testing.T, verifier string, challenge string, _ string) {
				t.Helper()

				hash := sha256.Sum256([]byte(verifier))
				expected := base64.RawURLEncoding.EncodeToString(hash[:])
				testify.Equal(t, expected, challenge, "Default method should be S256")
			},
		},
		{
			name:            cryptoutilSharedMagic.INVALID_METHOD,
			verifier:        cryptoutilSharedMagic.TEST_CODE_VERIFIER,
			method:          cryptoutilSharedMagic.INVALID_METHOD,
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
			name:     cryptoutilSharedMagic.S256_METHOD,
			verifier: cryptoutilSharedMagic.S256_CODE_CHALLENGE,
		},
		{
			name:     cryptoutilSharedMagic.SHORT_VERIFIER,
			verifier: "abc",
		},
		{
			name:     cryptoutilSharedMagic.LONG_VERIFIER,
			verifier: cryptoutilSharedMagic.VERY_LONG_VERIFIER,
		},
		{
			name:     cryptoutilSharedMagic.EMPTY_VERIFIER,
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
			name:      cryptoutilSharedMagic.S256_METHOD,
			verifier:  cryptoutilSharedMagic.S256_CODE_CHALLENGE,
			challenge: GenerateS256Challenge(cryptoutilSharedMagic.S256_CODE_CHALLENGE),
			method:    cryptoutilSharedMagic.PKCEMethodS256,
			wantValid: true,
		},
		{
			name:      cryptoutilSharedMagic.S256_METHOD,
			verifier:  cryptoutilSharedMagic.WRONG_VERIFIER,
			challenge: GenerateS256Challenge(cryptoutilSharedMagic.S256_CODE_CHALLENGE),
			method:    cryptoutilSharedMagic.PKCEMethodS256,
			wantValid: false,
		},
		{
			name:      cryptoutilSharedMagic.PLAIN_METHOD,
			verifier:  cryptoutilSharedMagic.TEST_CODE_VERIFIER,
			challenge: cryptoutilSharedMagic.TEST_CODE_VERIFIER,
			method:    cryptoutilSharedMagic.PKCEMethodPlain,
			wantValid: true,
		},
		{
			name:      cryptoutilSharedMagic.PLAIN_METHOD,
			verifier:  cryptoutilSharedMagic.TEST_CODE_VERIFIER,
			challenge: cryptoutilSharedMagic.WRONG_VERIFIER,
			method:    cryptoutilSharedMagic.PKCEMethodPlain,
			wantValid: false,
		},
		{
			name:      cryptoutilSharedMagic.DEFAULT_METHOD_S256,
			verifier:  cryptoutilSharedMagic.S256_CODE_CHALLENGE,
			challenge: GenerateS256Challenge(cryptoutilSharedMagic.S256_CODE_CHALLENGE),
			method:    "",
			wantValid: true,
		},
		{
			name:      cryptoutilSharedMagic.INVALID_METHOD,
			verifier:  cryptoutilSharedMagic.TEST_CODE_VERIFIER,
			challenge: cryptoutilSharedMagic.TEST_CODE_CHALLENGE,
			method:    cryptoutilSharedMagic.INVALID_METHOD,
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
			name:      cryptoutilSharedMagic.S256_METHOD,
			verifier:  cryptoutilSharedMagic.S256_CODE_CHALLENGE,
			challenge: GenerateS256Challenge(cryptoutilSharedMagic.S256_CODE_CHALLENGE),
			expected:  true,
		},
		{
			name:      cryptoutilSharedMagic.S256_METHOD,
			verifier:  cryptoutilSharedMagic.WRONG_VERIFIER,
			challenge: GenerateS256Challenge(cryptoutilSharedMagic.S256_CODE_CHALLENGE),
			expected:  false,
		},
		{
			name:      cryptoutilSharedMagic.EMPTY_VERIFIER,
			verifier:  "",
			challenge: GenerateS256Challenge(""),
			expected:  true,
		},
		{
			name:      cryptoutilSharedMagic.EMPTY_VERIFIER_WITH_NON_EMPTY_CHALLENGE,
			verifier:  "",
			challenge: GenerateS256Challenge(cryptoutilSharedMagic.NON_EMPTY),
			expected:  false,
		},
		{
			name:      cryptoutilSharedMagic.NON_EMPTY_VERIFIER_WITH_EMPTY_CHALLENGE,
			verifier:  cryptoutilSharedMagic.NON_EMPTY,
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
			name:   cryptoutilSharedMagic.S256_METHOD,
			method: cryptoutilSharedMagic.PKCEMethodS256,
		},
		{
			name:   cryptoutilSharedMagic.PLAIN_METHOD,
			method: cryptoutilSharedMagic.PKCEMethodPlain,
		},
		{
			name:   cryptoutilSharedMagic.DEFAULT_METHOD_S256,
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

	verifiers := make([]string, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	challenges := make([]string, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)

	// Generate multiple verifier/challenge pairs
	for i := 0; i < cryptoutilSharedMagic.JoseJADefaultMaxMaterials; i++ {
		verifier, err := GenerateCodeVerifier()
		testify.NoError(t, err, "Generate code verifier should succeed")

		challenge := GenerateCodeChallenge(verifier, cryptoutilSharedMagic.PKCEMethodS256)
		testify.NotEmpty(t, challenge, "Code challenge should not be empty")

		verifiers[i] = verifier
		challenges[i] = challenge
	}

	// Verify each verifier validates only with its own challenge
	for i := 0; i < cryptoutilSharedMagic.JoseJADefaultMaxMaterials; i++ {
		for j := 0; j < cryptoutilSharedMagic.JoseJADefaultMaxMaterials; j++ {
			valid := ValidateCodeVerifier(verifiers[i], challenges[j], cryptoutilSharedMagic.PKCEMethodS256)

			if i == j {
				testify.True(t, valid, "Verifier should validate with its own challenge")
			} else {
				testify.False(t, valid, "Verifier should not validate with different challenge")
			}
		}
	}
}
