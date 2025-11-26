// Copyright (c) 2025 Justin Cranford
//
//

package clientauth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// PBKDF2 parameters (FIPS 140-3 approved).
	pbkdf2Iterations = 600000 // OWASP recommendation for PBKDF2-HMAC-SHA256
	saltLength       = 32     // 256 bits
	keyLength        = 32     // 256 bits
)

// HashSecret hashes a client secret using PBKDF2-HMAC-SHA256 (FIPS 140-3 approved).
// Returns a base64-encoded string: "salt:hash".
func HashSecret(secret string) (string, error) {
	// Generate random salt.
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive key using PBKDF2-HMAC-SHA256.
	hash := pbkdf2.Key([]byte(secret), salt, pbkdf2Iterations, keyLength, sha256.New)

	// Encode as "salt:hash" in base64.
	encoded := base64.StdEncoding.EncodeToString(salt) + ":" + base64.StdEncoding.EncodeToString(hash)

	return encoded, nil
}

// CompareSecret compares a hashed secret with a plain secret using constant-time comparison.
// hashed format: "salt:hash" (base64-encoded).
func CompareSecret(hashed, plain string) (bool, error) {
	// Parse hashed secret.
	parts := splitHashedSecret(hashed)
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid hashed secret format (expected 'salt:hash')")
	}

	// Decode salt.
	salt, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return false, fmt.Errorf("failed to decode salt: %w", err)
	}

	// Decode stored hash.
	storedHash, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return false, fmt.Errorf("failed to decode hash: %w", err)
	}

	// Derive hash from plain secret.
	derivedHash := pbkdf2.Key([]byte(plain), salt, pbkdf2Iterations, keyLength, sha256.New)

	// Constant-time comparison.
	return subtle.ConstantTimeCompare(storedHash, derivedHash) == 1, nil
}

// splitHashedSecret splits "salt:hash" into [salt, hash].
func splitHashedSecret(hashed string) []string {
	result := make([]string, 0, 2)

	for i := 0; i < len(hashed); i++ {
		if hashed[i] == ':' {
			result = append(result, hashed[:i], hashed[i+1:])

			break
		}
	}

	return result
}
