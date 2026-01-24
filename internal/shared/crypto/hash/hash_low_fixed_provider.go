// Copyright (c) 2025 Justin Cranford

package hash

import (
	"encoding/base64"
	"errors"
	"fmt"

	cryptoutilSharedCryptoDigests "cryptoutil/internal/shared/crypto/digests"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// HashLowEntropyDeterministic hashes a low-entropy secret (e.g., password, PIN) using HKDF-SHA256 with a fixed info parameter.
// This produces deterministic output (same secret → same hash every time).
// Use this when you need consistent hashing for low-entropy secrets (e.g., password lookup tables, caching).
//
// Use cases:
//   - Password lookup tables where determinism is required
//   - Caching password hashes for performance
//   - Migration scenarios where deterministic output is needed
//
// Security considerations:
//   - No random salt means identical secrets produce identical hashes (vulnerable to rainbow tables)
//   - Use only when determinism is absolutely required
//   - For better security, prefer HashLowEntropyNonDeterministic (PBKDF2 with random salt)
//
// Format: "hkdf-sha256-fixed$base64(dk)"
// Returns: HKDF-based hash string in the format above, or error if secret is empty.
func HashLowEntropyDeterministic(secret string) (string, error) {
	return HashSecretHKDFFixed(secret, cryptoutilSharedMagic.HKDFFixedInfoLowEntropy)
}

// HashSecretHKDFFixed performs HKDF-SHA256 key derivation with a fixed info parameter (deterministic).
// This function is the underlying implementation for deterministic HKDF-based hashing.
//
// Parameters:
//   - secret: The secret to hash (must not be empty)
//   - fixedInfo: Fixed info parameter for HKDF (used instead of random salt for determinism)
//
// Format: "hkdf-sha256-fixed$base64(dk)"
//   - hkdf-sha256-fixed: Algorithm identifier
//   - base64(dk): Base64-encoded derived key (32 bytes)
//
// Returns: Hash string in the format above, or error if secret is empty or HKDF fails.
func HashSecretHKDFFixed(secret string, fixedInfo []byte) (string, error) {
	if secret == "" {
		return "", errors.New("secret cannot be empty")
	}

	algorithm := cryptoutilSharedMagic.HKDFFixedLowHashName

	const dkLength = cryptoutilSharedMagic.PBKDF2DerivedKeyLength // 32 bytes

	// Use HKDF with no salt (nil), fixed info parameter for deterministic output.
	secretBytes := []byte(secret)

	dk, err := cryptoutilSharedCryptoDigests.HKDF(cryptoutilSharedMagic.SHA256, secretBytes, nil, fixedInfo, dkLength)
	if err != nil {
		return "", fmt.Errorf("HKDF key derivation failed: %w", err)
	}

	// Format: hkdf-sha256-fixed$base64(dk)
	dkBase64 := base64.StdEncoding.EncodeToString(dk)

	return fmt.Sprintf("%s%s%s", algorithm, cryptoutilSharedMagic.HKDFDelimiter, dkBase64), nil
}

// VerifySecretHKDFFixed verifies a secret against a stored HKDF-fixed hash.
// Uses constant-time comparison to prevent timing attacks.
//
// Parameters:
//   - storedHash: The stored hash string (format: "hkdf-sha256-fixed$base64(dk)")
//   - providedSecret: The secret to verify
//
// Returns: true if the secret matches the stored hash, false otherwise, or error if inputs are invalid.
func VerifySecretHKDFFixed(storedHash, providedSecret string) (bool, error) {
	if storedHash == "" {
		return false, errors.New("stored hash cannot be empty")
	}

	if providedSecret == "" {
		return false, errors.New("provided secret cannot be empty")
	}

	// Parse stored hash: hkdf-sha256-fixed$base64(dk)
	parts := splitHKDFFixedParts(storedHash)

	const expectedParts = 2
	if len(parts) != expectedParts {
		return false, fmt.Errorf("invalid hash format: expected %d parts, got %d", expectedParts, len(parts))
	}

	algorithm := parts[0]
	if algorithm != "hkdf-sha256-fixed" {
		return false, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}

	storedDKBase64 := parts[1]

	storedDK, err := base64.StdEncoding.DecodeString(storedDKBase64)
	if err != nil {
		return false, fmt.Errorf("failed to decode stored derived key: %w", err)
	}

	// Re-derive the key using the same fixed info parameter.
	providedHash, err := HashSecretHKDFFixed(providedSecret, cryptoutilSharedMagic.HKDFFixedInfoLowEntropy)
	if err != nil {
		return false, fmt.Errorf("failed to hash provided secret: %w", err)
	}

	providedParts := splitHKDFFixedParts(providedHash)
	if len(providedParts) != expectedParts {
		return false, errors.New("invalid provided hash format")
	}

	providedDKBase64 := providedParts[1]

	providedDK, err := base64.StdEncoding.DecodeString(providedDKBase64)
	if err != nil {
		return false, fmt.Errorf("failed to decode provided derived key: %w", err)
	}

	// Constant-time comparison to prevent timing attacks.
	return constantTimeCompareBytes(storedDK, providedDK), nil
}

// splitHKDFFixedParts splits an HKDF-fixed hash string into its components.
// Expected format: "hkdf-sha256-fixed$base64(dk)" → ["hkdf-sha256-fixed", "base64(dk)"].
func splitHKDFFixedParts(hash string) []string {
	parts := make([]string, 0, 2)

	delimiter := cryptoutilSharedMagic.HKDFDelimiter[0] // Use magic constant delimiter.

	start := 0

	for i := 0; i < len(hash); i++ {
		if hash[i] == delimiter {
			parts = append(parts, hash[start:i])
			start = i + 1
		}
	}
	// Add the last part.
	if start < len(hash) {
		parts = append(parts, hash[start:])
	}

	return parts
}

// constantTimeCompareBytes performs a constant-time comparison of two byte slices.
// This prevents timing attacks by ensuring comparison time is independent of where the mismatch occurs.
//
// Returns: true if slices are equal, false otherwise.
func constantTimeCompareBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	result := byte(0)
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}

	return result == 0
}
