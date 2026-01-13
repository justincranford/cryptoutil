// Copyright (c) 2025 ZREV Enterprises LLC. All rights reserved.
// Use of this source code is governed by the MIT License.

// Package pbkdf2 provides FIPS 140-2/140-3 compliant password hashing using PBKDF2-HMAC-SHA256.
package pbkdf2

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/sha3"
)

const (
	// Iterations600k is the OWASP 2025 recommended minimum for PBKDF2-HMAC-SHA256.
	Iterations600k = 600000
	
	// SaltLength32 is 256 bits of salt.
	SaltLength32 = 32
	
	// KeyLength32 is 256 bits of derived key.
	KeyLength32 = 32
)

// HashPassword generates a FIPS-compliant PBKDF2-HMAC-SHA256 hash of the password.
// Returns PHC format: $pbkdf2-sha256$600000$<base64-salt>$<base64-hash>
func HashPassword(password string) (string, error) {
	return HashPasswordWithIterations(password, Iterations600k)
}

// HashPasswordWithIterations allows customizing iteration count for testing.
func HashPasswordWithIterations(password string, iterations int) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}
	
	if iterations < 210000 {
		return "", fmt.Errorf("iterations must be at least 210000 (OWASP 2023 minimum), got %d", iterations)
	}
	
	// Generate cryptographically secure random salt.
	salt := make([]byte, SaltLength32)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}
	
	// Derive key using PBKDF2-HMAC-SHA256.
	hash := pbkdf2.Key([]byte(password), salt, iterations, KeyLength32, sha3.New256)
	
	// Encode in PHC format.
	saltB64 := base64.RawStdEncoding.EncodeToString(salt)
	hashB64 := base64.RawStdEncoding.EncodeToString(hash)
	
	return fmt.Sprintf("$pbkdf2-sha256$%d$%s$%s", iterations, saltB64, hashB64), nil
}

// VerifyPassword verifies a password against a PBKDF2 hash.
// Returns true if password matches, false otherwise.
func VerifyPassword(password, storedHash string) (bool, error) {
	if password == "" {
		return false, fmt.Errorf("password cannot be empty")
	}
	
	if storedHash == "" {
		return false, fmt.Errorf("stored hash cannot be empty")
	}
	
	// Parse PHC format: $pbkdf2-sha256$iterations$salt$hash
	parts := strings.Split(storedHash, "$")
	if len(parts) != 5 {
		return false, fmt.Errorf("invalid hash format: expected 5 parts, got %d", len(parts))
	}
	
	if parts[0] != "" {
		return false, fmt.Errorf("invalid hash format: expected empty first part")
	}
	
	if parts[1] != "pbkdf2-sha256" {
		return false, fmt.Errorf("invalid hash algorithm: expected pbkdf2-sha256, got %s", parts[1])
	}
	
	var iterations int
	if _, err := fmt.Sscanf(parts[2], "%d", &iterations); err != nil {
		return false, fmt.Errorf("invalid iterations: %w", err)
	}
	
	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false, fmt.Errorf("invalid salt encoding: %w", err)
	}
	
	storedHashBytes, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("invalid hash encoding: %w", err)
	}
	
	// Derive key with same parameters.
	computedHash := pbkdf2.Key([]byte(password), salt, iterations, len(storedHashBytes), sha3.New256)
	
	// Constant-time comparison to prevent timing attacks.
	if subtle.ConstantTimeCompare(computedHash, storedHashBytes) == 1 {
		return true, nil
	}
	
	return false, nil
}

// DetectHashType returns the hash algorithm type from the hash string.
// Supports: "bcrypt", "pbkdf2", "unknown"
func DetectHashType(hash string) string {
	if strings.HasPrefix(hash, "$2a$") || strings.HasPrefix(hash, "$2b$") || strings.HasPrefix(hash, "$2y$") {
		return "bcrypt"
	}
	
	if strings.HasPrefix(hash, "$pbkdf2-sha256$") {
		return "pbkdf2"
	}
	
	return "unknown"
}
