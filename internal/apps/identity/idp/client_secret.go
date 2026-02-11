// Copyright (c) 2025 Justin Cranford
//
//

package idp

import (
	crand "crypto/rand"
	"encoding/base64"
	"fmt"

	cryptoutilIdentityClientAuth "cryptoutil/internal/apps/identity/authz/clientauth"
)

const (
	// Client secret length (256 bits / 8 bits per byte = 32 bytes).
	clientSecretLength = 32
)

// GenerateClientSecret generates a cryptographically secure client secret.
// Returns (plaintextSecret, hashedSecret, error).
// The plaintext secret should be shown to the user ONCE and then discarded.
// The hashed secret should be stored in the database.
func GenerateClientSecret() (string, string, error) {
	// Generate 32 bytes of cryptographic randomness (256 bits).
	secretBytes := make([]byte, clientSecretLength)
	if _, err := crand.Read(secretBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate client secret: %w", err)
	}

	// Encode as base64 for string representation.
	plaintextSecret := base64.StdEncoding.EncodeToString(secretBytes)

	// Hash the secret using PBKDF2-HMAC-SHA256 (FIPS 140-3 approved).
	hashedSecret, err := cryptoutilIdentityClientAuth.HashLowEntropyNonDeterministic(plaintextSecret)
	if err != nil {
		return "", "", fmt.Errorf("failed to hash client secret: %w", err)
	}

	return plaintextSecret, hashedSecret, nil
}
