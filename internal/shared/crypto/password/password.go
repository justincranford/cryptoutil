// Copyright (c) 2025 ZREV Enterprises LLC. All rights reserved.
// Use of this source code is governed by the MIT License.

// Package password provides FIPS-compliant PBKDF2-HMAC-SHA256 password hashing.
package password

import (
	"fmt"

	cryptoutilSharedCryptoPbkdf2 "cryptoutil/internal/shared/crypto/pbkdf2"
)

// HashPassword generates a FIPS-compliant PBKDF2-HMAC-SHA256 hash.
// Always use this for new passwords.
func HashPassword(password string) (string, error) {
	return hashPasswordInternal(password, cryptoutilSharedCryptoPbkdf2.HashPassword)
}

func hashPasswordInternal(password string, passwordHashFn func(string) (string, error)) (string, error) {
	hash, err := passwordHashFn(password)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return hash, nil
}

// VerifyPassword verifies a password against a PBKDF2-HMAC-SHA256 hash.
//
// Returns: (match bool, needsUpgrade bool, error)
//   - match: true if password matches hash
//   - needsUpgrade: always false (no legacy hash types remain)
//   - error: non-nil if verification fails
func VerifyPassword(password, storedHash string) (bool, bool, error) {
	if password == "" {
		return false, false, fmt.Errorf("password cannot be empty")
	}

	if storedHash == "" {
		return false, false, fmt.Errorf("stored hash cannot be empty")
	}

	hashType := cryptoutilSharedCryptoPbkdf2.DetectHashType(storedHash)

	switch hashType {
	case "pbkdf2":
		return verifyPasswordInternal(password, storedHash, cryptoutilSharedCryptoPbkdf2.VerifyPassword)

	default:
		return false, false, fmt.Errorf("unknown hash type: %s", hashType)
	}
}

// DetectHashType returns the hash algorithm type from the hash string.
// Supports: "pbkdf2", "unknown".
func DetectHashType(hash string) string {
	return cryptoutilSharedCryptoPbkdf2.DetectHashType(hash)
}

func verifyPasswordInternal(password, storedHash string, passwordVerifyFn func(string, string) (bool, error)) (bool, bool, error) {
	match, err := passwordVerifyFn(password, storedHash)
	if err != nil {
		return false, false, fmt.Errorf("pbkdf2 verification failed: %w", err)
	}

	return match, false, nil // No upgrade needed.
}
