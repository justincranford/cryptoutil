// Copyright (c) 2025 Justin Cranford

package hash

// HashLowEntropyNonDeterministic hashes a low-entropy secret (e.g., password, PIN) using a random salt.
// Each invocation produces a different hash for the same input (non-deterministic).
//
// This function uses PBKDF2-HMAC-SHA256 with cryptographically random salt generation.
// Suitable for authentication credentials where entropy cannot be guaranteed high.
//
// For high-entropy secrets (API keys, tokens), use HashHighEntropyNonDeterministic.
// For deterministic hashing, use HashLowEntropyDeterministic or HashHighEntropyDeterministic.
//
// FIPS mode is ALWAYS enabled - no configurable algorithm selection.
func HashLowEntropyNonDeterministic(secret string) (string, error) {
	return HashSecretPBKDF2(secret)
}
