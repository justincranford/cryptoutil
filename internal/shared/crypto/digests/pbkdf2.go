// Copyright (c) 2025 Justin Cranford

package digests

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	cryptoutilMagic "cryptoutil/internal/shared/magic"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2"
)

// HashSecretPBKDF2 returns a formatted PBKDF2 hash string: pbkdf2$iter$base64(salt)$base64(dk).
func HashSecretPBKDF2(secret string) (string, error) {
	if secret == "" {
		return "", errors.New("secret is empty")
	}

	salt := make([]byte, cryptoutilMagic.PBKDF2DefaultSaltBytes)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	dk := pbkdf2.Key([]byte(secret), salt, cryptoutilMagic.PBKDF2DefaultIterations, cryptoutilMagic.PBKDF2DerivedKeyLength, sha256.New)

	return fmt.Sprintf("%s$%d$%s$%s",
		cryptoutilMagic.PBKDF2Prefix,
		cryptoutilMagic.PBKDF2DefaultIterations,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(dk)), nil
}

// VerifySecret verifies a stored hash against a provided secret. It supports PBKDF2 formatted hashes and falls back to bcrypt for legacy entries.
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

	parts := strings.Split(stored, "$")
	if len(parts) != 4 || parts[0] != cryptoutilMagic.PBKDF2DefaultHashName {
		return false, fmt.Errorf("unsupported hash format")
	}

	// parts: ["pbkdf2", iterations, saltB64, dkB64]
	iter := 0
	if _, err := fmt.Sscanf(parts[1], "%d", &iter); err != nil || iter <= 0 {
		return false, fmt.Errorf("invalid iterations")
	}

	saltB64 := parts[2]
	dkB64 := parts[3]

	salt, err := base64.RawStdEncoding.DecodeString(saltB64)
	if err != nil {
		return false, fmt.Errorf("invalid salt encoding: %w", err)
	}

	expectedDK, err := base64.RawStdEncoding.DecodeString(dkB64)
	if err != nil {
		return false, fmt.Errorf("invalid dk encoding: %w", err)
	}

	derived := pbkdf2.Key([]byte(provided), salt, iter, len(expectedDK), sha256.New)

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

	return equal, nil
}
