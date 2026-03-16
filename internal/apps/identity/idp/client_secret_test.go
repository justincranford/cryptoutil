// Copyright (c) 2025 Justin Cranford
//
//

package idp

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"encoding/base64"
	"testing"

	cryptoutilIdentityClientAuth "cryptoutil/internal/apps/identity/authz/clientauth"

	testify "github.com/stretchr/testify/require"
)

func TestGenerateClientSecret(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{"generate and verify secret"},
		{"generate unique secrets"},
		{"secret length validation"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			plaintext, hashed, err := GenerateClientSecret()

			testify.NoError(t, err)
			testify.NotEmpty(t, plaintext)
			testify.NotEmpty(t, hashed)

			// Verify the plaintext is base64 encoded.
			decoded, err := base64.StdEncoding.DecodeString(plaintext)
			testify.NoError(t, err)
			testify.Len(t, decoded, clientSecretLength)

			// Verify the hashed secret can validate the plaintext.
			valid, err := cryptoutilIdentityClientAuth.CompareSecret(hashed, plaintext)
			testify.NoError(t, err)
			testify.True(t, valid)

			// Verify wrong plaintext fails validation.
			wrongValid, err := cryptoutilIdentityClientAuth.CompareSecret(hashed, "wrong_secret")
			testify.NoError(t, err)
			testify.False(t, wrongValid)
		})
	}
}

func TestGenerateClientSecretUniqueness(t *testing.T) {
	t.Parallel()

	// Generate multiple secrets and verify they're all unique.
	secrets := make(map[string]bool)

	for i := 0; i < cryptoutilSharedMagic.JoseJADefaultMaxMaterials; i++ {
		plaintext, hashed, err := GenerateClientSecret()
		testify.NoError(t, err)

		// Verify plaintext is unique.
		testify.False(t, secrets[plaintext], "duplicate plaintext generated")
		secrets[plaintext] = true

		// Verify hashed is unique.
		testify.False(t, secrets[hashed], "duplicate hashed secret generated")
		secrets[hashed] = true
	}
}
