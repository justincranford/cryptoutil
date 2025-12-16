package clientauth

import (
	cryptoutilDigests "cryptoutil/internal/shared/crypto/digests"
	cryptoutilHash "cryptoutil/internal/shared/crypto/hash"
)

// HashLowEntropyNonDeterministic hashes a client secret using PBKDF2-HMAC-SHA256 (FIPS 140-3 approved).
// Returns a versioned PBKDF2 hash string.
func HashLowEntropyNonDeterministic(secret string) (string, error) {
	return cryptoutilHash.HashLowEntropyNonDeterministic(secret)
}

// CompareSecret compares a hashed secret with a plain secret using constant-time comparison.
// hashed format: versioned PBKDF2 format.
func CompareSecret(hashed, plain string) (bool, error) {
	return cryptoutilDigests.VerifySecret(hashed, plain)
}
