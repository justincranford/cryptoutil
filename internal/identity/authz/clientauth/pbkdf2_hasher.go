// Copyright (c) 2025 Justin Cranford
//
//

package clientauth

import (
	crand "crypto/rand"
	sha256 "crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/pbkdf2"

	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// PBKDF2Hasher implements SecretHasher using FIPS 140-3 approved PBKDF2-HMAC-SHA256.
type PBKDF2Hasher struct {
	iterations int
	saltLength int
	keyLength  int
}

// NewPBKDF2Hasher creates a new PBKDF2 hasher with FIPS-approved parameters.
func NewPBKDF2Hasher() *PBKDF2Hasher {
	return &PBKDF2Hasher{
		iterations: cryptoutilIdentityMagic.PBKDF2Iterations,
		saltLength: cryptoutilIdentityMagic.PBKDF2SaltLength,
		keyLength:  cryptoutilIdentityMagic.PBKDF2KeyLength,
	}
}

// HashLowEntropyNonDeterministic hashes a plaintext client secret using PBKDF2-HMAC-SHA256.
// Format: $pbkdf2-sha256$iterations$base64(salt)$base64(hash).
func (h *PBKDF2Hasher) HashLowEntropyNonDeterministic(plaintext string) (string, error) {
	// Generate cryptographically secure random salt.
	salt := make([]byte, h.saltLength)
	if _, err := crand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive key using PBKDF2-HMAC-SHA256.
	hash := pbkdf2.Key([]byte(plaintext), salt, h.iterations, h.keyLength, sha256.New)

	// Encode as: $pbkdf2-sha256$iterations$salt$hash.
	encoded := fmt.Sprintf("$%s$%d$%s$%s",
		cryptoutilSharedMagic.PBKDF2DefaultHashName,
		h.iterations,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return encoded, nil
}

// CompareSecret compares a hashed secret with a plaintext secret using constant-time comparison.
func (h *PBKDF2Hasher) CompareSecret(hashed, plaintext string) error {
	// Parse stored hash format: $pbkdf2-sha256$iterations$salt$hash.
	parts := strings.Split(hashed, "$")
	if len(parts) != 5 || parts[0] != "" || parts[1] != cryptoutilSharedMagic.PBKDF2DefaultHashName {
		return fmt.Errorf("invalid hash format")
	}

	var iterations int
	if _, err := fmt.Sscanf(parts[2], "%d", &iterations); err != nil {
		return fmt.Errorf("invalid iterations: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return fmt.Errorf("invalid salt encoding: %w", err)
	}

	storedHash, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return fmt.Errorf("invalid hash encoding: %w", err)
	}

	// Derive key from plaintext using stored salt and iterations.
	derivedHash := pbkdf2.Key([]byte(plaintext), salt, iterations, len(storedHash), sha256.New)

	// Constant-time comparison to prevent timing attacks.
	if subtle.ConstantTimeCompare(derivedHash, storedHash) != 1 {
		return fmt.Errorf("client secret mismatch")
	}

	return nil
}
