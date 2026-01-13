// Copyright (c) 2025 ZREV Enterprises LLC. All rights reserved.
// Use of this source code is governed by the MIT License.

// Package password provides dual-mode password hashing supporting both legacy bcrypt
// (verification only) and FIPS-compliant PBKDF2-HMAC-SHA256 (generation + verification).
package password

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"

	cryptoutilPBKDF2 "cryptoutil/internal/shared/crypto/pbkdf2"
)

// HashPassword generates a FIPS-compliant PBKDF2-HMAC-SHA256 hash.
// Always use this for new passwords.
func HashPassword(password string) (string, error) {
	return cryptoutilPBKDF2.HashPassword(password)
}

// VerifyPassword verifies a password against either bcrypt (legacy) or PBKDF2 (new) hash.
// Automatically detects hash type and uses appropriate verification method.
//
// Returns: (match bool, needsUpgrade bool, error)
//   - match: true if password matches hash
//   - needsUpgrade: true if hash is bcrypt (should be upgraded to PBKDF2 on next change)
//   - error: non-nil if verification fails
func VerifyPassword(password, storedHash string) (bool, bool, error) {
	if password == "" {
		return false, false, fmt.Errorf("password cannot be empty")
	}
	
	if storedHash == "" {
		return false, false, fmt.Errorf("stored hash cannot be empty")
	}
	
	hashType := cryptoutilPBKDF2.DetectHashType(storedHash)
	
	switch hashType {
	case "bcrypt":
		// Legacy bcrypt - verify only, DO NOT generate new bcrypt hashes.
		err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
		if err != nil {
			if err == bcrypt.ErrMismatchedHashAndPassword {
				return false, true, nil // Password doesn't match, but still needs upgrade.
			}
			return false, true, fmt.Errorf("bcrypt verification failed: %w", err)
		}
		return true, true, nil // Match, needs upgrade to PBKDF2.
		
	case "pbkdf2":
		// Modern FIPS-compliant PBKDF2.
		match, err := cryptoutilPBKDF2.VerifyPassword(password, storedHash)
		if err != nil {
			return false, false, fmt.Errorf("pbkdf2 verification failed: %w", err)
		}
		return match, false, nil // No upgrade needed.
		
	default:
		return false, false, fmt.Errorf("unknown hash type: %s", hashType)
	}
}

// DetectHashType returns the hash algorithm type from the hash string.
// Supports: "bcrypt", "pbkdf2", "unknown"
func DetectHashType(hash string) string {
	return cryptoutilPBKDF2.DetectHashType(hash)
}
