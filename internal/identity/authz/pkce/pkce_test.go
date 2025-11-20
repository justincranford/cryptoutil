// Copyright (c) 2025 Justin Cranford

package pkce

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"

	testify "github.com/stretchr/testify/require"

	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// TestGenerateCodeVerifier tests PKCE code verifier generation.
func TestGenerateCodeVerifier(t *testing.T) {
	t.Parallel()

	verifier, err := GenerateCodeVerifier()

	testify.NoError(t, err, "Generate code verifier should succeed")
	testify.NotEmpty(t, verifier, "Code verifier should not be empty")

	// RFC 7636: code verifier should be at least 43 characters
	testify.GreaterOrEqual(t, len(verifier), 43, "Code verifier should be at least 43 characters")

	// Verify base64url encoding (should not contain + or / or =)
	testify.NotContains(t, verifier, "+", "Code verifier should use base64url encoding")
	testify.NotContains(t, verifier, "/", "Code verifier should use base64url encoding")
	testify.NotContains(t, verifier, "=", "Code verifier should not have padding")
}

// TestGenerateCodeVerifier_Uniqueness tests that generated verifiers are unique.
func TestGenerateCodeVerifier_Uniqueness(t *testing.T) {
	t.Parallel()

	verifiers := make(map[string]bool)

	// Generate multiple verifiers and ensure uniqueness
	for i := 0; i < 100; i++ {
		verifier, err := GenerateCodeVerifier()
		testify.NoError(t, err, "Generate code verifier should succeed")
		testify.False(t, verifiers[verifier], "Code verifiers should be unique")

		verifiers[verifier] = true
	}
}

// TestGenerateCodeChallenge_S256 tests S256 code challenge generation.
func TestGenerateCodeChallenge_S256(t *testing.T) {
	t.Parallel()

	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"

	challenge := GenerateCodeChallenge(verifier, cryptoutilIdentityMagic.PKCEMethodS256)

	testify.NotEmpty(t, challenge, "Code challenge should not be empty")
	testify.NotEqual(t, verifier, challenge, "Challenge should differ from verifier")

	// Verify expected S256 challenge (SHA256 hash)
	hash := sha256.Sum256([]byte(verifier))
	expected := base64.RawURLEncoding.EncodeToString(hash[:])
	testify.Equal(t, expected, challenge, "S256 challenge should match expected SHA256 hash")
}

// TestGenerateCodeChallenge_Plain tests plain code challenge generation.
func TestGenerateCodeChallenge_Plain(t *testing.T) {
	t.Parallel()

	verifier := "test-code-verifier"

	challenge := GenerateCodeChallenge(verifier, cryptoutilIdentityMagic.PKCEMethodPlain)

	testify.Equal(t, verifier, challenge, "Plain challenge should equal verifier")
}

// TestGenerateCodeChallenge_DefaultMethod tests default method (S256).
func TestGenerateCodeChallenge_DefaultMethod(t *testing.T) {
	t.Parallel()

	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"

	// Empty method should default to S256
	challenge := GenerateCodeChallenge(verifier, "")

	hash := sha256.Sum256([]byte(verifier))
	expected := base64.RawURLEncoding.EncodeToString(hash[:])
	testify.Equal(t, expected, challenge, "Default method should be S256")
}

// TestGenerateCodeChallenge_InvalidMethod tests handling of invalid method.
func TestGenerateCodeChallenge_InvalidMethod(t *testing.T) {
	t.Parallel()

	verifier := "test-code-verifier"

	challenge := GenerateCodeChallenge(verifier, "invalid-method")

	testify.Empty(t, challenge, "Invalid method should return empty challenge")
}

// TestGenerateS256Challenge tests S256 challenge generation directly.
func TestGenerateS256Challenge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		verifier string
	}{
		{
			name:     "Standard verifier",
			verifier: "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
		},
		{
			name:     "Short verifier",
			verifier: "abc",
		},
		{
			name:     "Long verifier",
			verifier: "very-long-code-verifier-with-many-characters-for-testing-purposes",
		},
		{
			name:     "Empty verifier",
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

// TestValidateCodeVerifier_S256_Valid tests successful S256 validation.
func TestValidateCodeVerifier_S256_Valid(t *testing.T) {
	t.Parallel()

	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	challenge := GenerateS256Challenge(verifier)

	valid := ValidateCodeVerifier(verifier, challenge, cryptoutilIdentityMagic.PKCEMethodS256)

	testify.True(t, valid, "Valid verifier should pass validation")
}

// TestValidateCodeVerifier_S256_Invalid tests failed S256 validation.
func TestValidateCodeVerifier_S256_Invalid(t *testing.T) {
	t.Parallel()

	verifier := "correct-verifier"
	wrongVerifier := "wrong-verifier"
	challenge := GenerateS256Challenge(verifier)

	valid := ValidateCodeVerifier(wrongVerifier, challenge, cryptoutilIdentityMagic.PKCEMethodS256)

	testify.False(t, valid, "Invalid verifier should fail validation")
}

// TestValidateCodeVerifier_Plain_Valid tests successful plain validation.
func TestValidateCodeVerifier_Plain_Valid(t *testing.T) {
	t.Parallel()

	verifier := "test-code-verifier"
	challenge := verifier // Plain method uses verifier as challenge

	valid := ValidateCodeVerifier(verifier, challenge, cryptoutilIdentityMagic.PKCEMethodPlain)

	testify.True(t, valid, "Valid plain verifier should pass validation")
}

// TestValidateCodeVerifier_Plain_Invalid tests failed plain validation.
func TestValidateCodeVerifier_Plain_Invalid(t *testing.T) {
	t.Parallel()

	verifier := "correct-verifier"
	challenge := "wrong-verifier"

	valid := ValidateCodeVerifier(verifier, challenge, cryptoutilIdentityMagic.PKCEMethodPlain)

	testify.False(t, valid, "Invalid plain verifier should fail validation")
}

// TestValidateCodeVerifier_DefaultMethod tests default method (S256) validation.
func TestValidateCodeVerifier_DefaultMethod(t *testing.T) {
	t.Parallel()

	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	challenge := GenerateS256Challenge(verifier)

	// Empty method should default to S256
	valid := ValidateCodeVerifier(verifier, challenge, "")

	testify.True(t, valid, "Default method should be S256")
}

// TestValidateCodeVerifier_InvalidMethod tests handling of invalid method.
func TestValidateCodeVerifier_InvalidMethod(t *testing.T) {
	t.Parallel()

	verifier := "test-verifier"
	challenge := "test-challenge"

	valid := ValidateCodeVerifier(verifier, challenge, "invalid-method")

	testify.False(t, valid, "Invalid method should fail validation")
}

// TestValidateS256 tests S256 validation directly.
func TestValidateS256(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		verifier  string
		challenge string
		expected  bool
	}{
		{
			name:      "Valid verifier and challenge",
			verifier:  "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			challenge: GenerateS256Challenge("dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"),
			expected:  true,
		},
		{
			name:      "Invalid verifier",
			verifier:  "wrong-verifier",
			challenge: GenerateS256Challenge("correct-verifier"),
			expected:  false,
		},
		{
			name:      "Empty verifier",
			verifier:  "",
			challenge: GenerateS256Challenge(""),
			expected:  true,
		},
		{
			name:      "Empty verifier with non-empty challenge",
			verifier:  "",
			challenge: GenerateS256Challenge("non-empty"),
			expected:  false,
		},
		{
			name:      "Non-empty verifier with empty challenge",
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

// TestPKCERoundtrip tests full PKCE flow: generate verifier → generate challenge → validate.
func TestPKCERoundtrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
	}{
		{
			name:   "S256 method",
			method: cryptoutilIdentityMagic.PKCEMethodS256,
		},
		{
			name:   "Plain method",
			method: cryptoutilIdentityMagic.PKCEMethodPlain,
		},
		{
			name:   "Default method (S256)",
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

// TestPKCERoundtrip_MultipleVerifiers tests PKCE with multiple verifiers.
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
