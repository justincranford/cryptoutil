// Copyright (c) 2025 Justin Cranford
//
//

package clientauth

import (
	"fmt"

	cryptoutilDigests "cryptoutil/internal/shared/crypto/digests"
	cryptoutilHash "cryptoutil/internal/shared/crypto/hash"
)

// HashLowEntropyNonDeterministic hashes a client secret using PBKDF2-HMAC-SHA256 (FIPS 140-3 approved).
// Returns a versioned PBKDF2 hash string.
func HashLowEntropyNonDeterministic(secret string) (string, error) {
	hash, err := cryptoutilHash.HashLowEntropyNonDeterministic(secret)
	if err != nil {
		return "", fmt.Errorf("failed to hash secret: %w", err)
	}

	return hash, nil
}

// CompareSecret compares a hashed secret with a plain secret using constant-time comparison.
// hashed format: versioned PBKDF2 format.
func CompareSecret(hashed, plain string) (bool, error) {
	match, err := cryptoutilDigests.VerifySecret(hashed, plain)
	if err != nil {
		return false, fmt.Errorf("failed to verify secret: %w", err)
	}

	return match, nil
}
