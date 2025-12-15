// Copyright (c) 2025 Justin Cranford

package digests

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"hash"
	"strings"

	cryptoutilMagic "cryptoutil/internal/shared/magic"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2"
)

// HashSecretPBKDF2 returns a formatted PBKDF2 hash string using default parameter set (version "1").
// Format: {1}$pbkdf2-sha256$iter$base64(salt)$base64(dk).
func HashSecretPBKDF2(secret string) (string, error) {
	return HashSecretPBKDF2WithParams(secret, DefaultPBKDF2ParameterSet())
}

// HashSecretPBKDF2WithParams returns a formatted PBKDF2 hash string using specified parameter set.
// Format: {version}$hashname$iter$base64(salt)$base64(dk).
func HashSecretPBKDF2WithParams(secret string, params PBKDF2ParameterSet) (string, error) {
	if secret == "" {
		return "", errors.New("secret is empty")
	}

	salt := make([]byte, params.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	dk := pbkdf2.Key([]byte(secret), salt, params.Iterations, params.KeyLength, params.HashFunc)

	return fmt.Sprintf("{%s}$%s$%d$%s$%s",
		params.Version,
		params.HashName,
		params.Iterations,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(dk)), nil
}

// VerifySecret verifies a stored hash against a provided secret.
// Supports:
//  1. Versioned PBKDF2 format: {version}$pbkdf2-sha256$iter$base64(salt)$base64(dk)
//  2. Legacy PBKDF2 format: pbkdf2-sha256$iter$base64(salt)$base64(dk)
//  3. Legacy bcrypt format: $2a$..., $2b$..., $2y$...
func VerifySecret(stored, provided string) (bool, error) {
	if stored == "" {
		return false, errors.New("stored hash empty")
	}

	// Legacy bcrypt support: hashes start with $2a$, $2b$, $2y$
	if strings.HasPrefix(stored, "$2a$") || strings.HasPrefix(stored, "$2b$") || strings.HasPrefix(stored, "$2y$") {
		err := bcrypt.CompareHashAndPassword([]byte(stored), []byte(provided))
		if err == nil {
			return true, nil
		}

		// bcrypt.ErrMismatchedHashAndPassword means wrong password = not an error, just false
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil //nolint:nilerr // Wrong password is not an error condition
		}

		// Any other bcrypt error (invalid hash format, etc) is a real error
		return false, fmt.Errorf("bcrypt hash verification failed: %w", err)
	}

	// Handle versioned format: {version}$hashname$iter$salt$dk
	var (
		version, hashname string
		iter              int
		saltB64, dkB64    string
	)

	if strings.HasPrefix(stored, "{") {
		// Versioned format: {1}$pbkdf2-sha256$600000$salt$dk
		parts := strings.Split(stored, "$")
		if len(parts) != cryptoutilMagic.PBKDF2VersionedFormatParts {
			return false, fmt.Errorf("invalid versioned hash format (expected %d parts)", cryptoutilMagic.PBKDF2VersionedFormatParts)
		}

		// Extract version from {1} format
		versionPart := parts[0]
		if !strings.HasPrefix(versionPart, "{") || !strings.HasSuffix(versionPart, "}") {
			return false, fmt.Errorf("invalid version format")
		}

		version = versionPart[1 : len(versionPart)-1]

		hashname = parts[1]
		if _, err := fmt.Sscanf(parts[2], "%d", &iter); err != nil || iter <= 0 {
			return false, fmt.Errorf("invalid iterations")
		}

		saltB64 = parts[3]
		dkB64 = parts[4]
	} else {
		// Legacy format: pbkdf2-sha256$600000$salt$dk (no version prefix)
		parts := strings.Split(stored, "$")
		if len(parts) != cryptoutilMagic.PBKDF2LegacyFormatParts {
			return false, fmt.Errorf("invalid legacy hash format (expected %d parts)", cryptoutilMagic.PBKDF2LegacyFormatParts)
		}

		hashname = parts[0]
		if hashname != cryptoutilMagic.PBKDF2DefaultHashName {
			return false, fmt.Errorf("unsupported hash format")
		}

		if _, err := fmt.Sscanf(parts[1], "%d", &iter); err != nil || iter <= 0 {
			return false, fmt.Errorf("invalid iterations")
		}

		saltB64 = parts[2]
		dkB64 = parts[3]
		version = "legacy"
	}

	salt, err := base64.RawStdEncoding.DecodeString(saltB64)
	if err != nil {
		return false, fmt.Errorf("invalid salt encoding: %w", err)
	}

	expectedDK, err := base64.RawStdEncoding.DecodeString(dkB64)
	if err != nil {
		return false, fmt.Errorf("invalid dk encoding: %w", err)
	}

	// Determine hash function based on hashname (currently only SHA-256 supported)
	var hashFunc func() hash.Hash
	if hashname == cryptoutilMagic.PBKDF2DefaultHashName {
		hashFunc = sha256.New
	} else {
		return false, fmt.Errorf("unsupported hash algorithm: %s", hashname)
	}

	derived := pbkdf2.Key([]byte(provided), salt, iter, len(expectedDK), hashFunc)

	if len(derived) != len(expectedDK) {
		return false, nil
	}

	// Constant-time compare
	equal := true

	for i := 0; i < len(derived); i++ {
		if derived[i] != expectedDK[i] {
			equal = false
		}
	}

	_ = version // Version extracted but not yet used for parameter set lookup

	return equal, nil
}
