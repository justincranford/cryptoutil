// Copyright (c) 2025 Justin Cranford

package digests

import (
	crand "crypto/rand"
	sha256 "crypto/sha256"
	sha512 "crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"hash"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"golang.org/x/crypto/pbkdf2"
)

// Injectable var for testing - allows error path coverage for crypto/rand.Read.
var digestsRandReadFn = crand.Read

// PBKDF2Params defines parameters for PBKDF2-HMAC hashing.
type PBKDF2Params struct {
	// Version identifier (e.g., "1", "2", "3") for versioned hash format.
	Version string

	// HashName is the algorithm identifier (e.g., "pbkdf2-sha256").
	HashName string

	// Iterations is the number of PBKDF2 iterations (OWASP: 600,000+ for SHA-256).
	Iterations int

	// SaltLength is the salt size in bytes (OWASP: 32+ bytes = 256 bits).
	SaltLength int

	// KeyLength is the derived key length in bytes (32 bytes = 256 bits).
	KeyLength int

	// HashFunc returns the hash function for PBKDF2 (e.g., sha256.New).
	HashFunc func() hash.Hash

	// Pepper is the version-specific secret pepper value.
	// MANDATORY per OWASP Password Storage Cheat Sheet: https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#peppering
	// Pattern: PBKDF2(password||pepper, salt, iterations, keyLength)
	// Storage: Docker/Kubernetes secrets (NEVER in DB/source code)
	// Rotation: Requires version bump + re-hash all records (lazy migration)
	Pepper string
}

// PBKDF2WithParams returns a formatted PBKDF2 hash string using specified parameter set.
// Format: {version}$hashname$iter$base64(salt)$base64(dk).
//
// CRITICAL: OWASP MANDATORY requirement - pepper MUST be concatenated with secret before PBKDF2.
// Reference: https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#peppering
// Pattern: PBKDF2(password||pepper, salt, iterations, keyLength).
func PBKDF2WithParams(secret string, params *PBKDF2Params) (string, error) {
	if secret == "" {
		return "", errors.New("secret is empty")
	} else if params == nil {
		return "", errors.New("parameter set is nil")
	}

	salt := make([]byte, params.SaltLength)
	if _, err := digestsRandReadFn(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// CRITICAL: Concatenate secret||pepper before PBKDF2 (OWASP requirement).
	// Pepper MUST be configured in params.Pepper (loaded from Docker/K8s secrets).
	// Empty pepper is allowed (backward compatibility with non-peppered hashes).
	pepperedSecret := secret + params.Pepper

	dk := pbkdf2.Key([]byte(pepperedSecret), salt, params.Iterations, params.KeyLength, params.HashFunc)

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
//  2. Versioned PBKDF2 format: {version}$pbkdf2-sha384$iter$base64(salt)$base64(dk)
//  3. Versioned PBKDF2 format: {version}$pbkdf2-sha512$iter$base64(salt)$base64(dk)
//
// Legacy format (no version prefix) is NOT supported - all hashes must be versioned.
//
// CRITICAL: This function does NOT support pepper - use VerifySecretWithParams for peppered hashes.
// For backward compatibility with non-peppered hashes only.
func VerifySecret(stored, provided string) (bool, error) {
	return VerifySecretWithParams(stored, provided, &PBKDF2Params{Pepper: ""})
}

// VerifySecretWithParams verifies a stored hash against a provided secret using specified parameter set.
// CRITICAL: OWASP MANDATORY requirement - pepper MUST be concatenated with secret before PBKDF2.
// Reference: https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#peppering
// Pattern: PBKDF2(password||pepper, salt, iterations, keyLength).
func VerifySecretWithParams(stored, provided string, params *PBKDF2Params) (bool, error) {
	version, hashname, iter, salt, expectedDK, err := parsePbkdf2Params(stored)
	if err != nil {
		return false, err
	}

	// Determine hash function based on hashname
	var hashFunc func() hash.Hash

	switch hashname {
	case cryptoutilSharedMagic.PBKDF2DefaultHashName:
		hashFunc = sha256.New
	case cryptoutilSharedMagic.PBKDF2SHA384HashName:
		hashFunc = sha512.New384
	case cryptoutilSharedMagic.PBKDF2SHA512HashName:
		hashFunc = sha512.New
	default:
		return false, fmt.Errorf("unsupported hash algorithm: %s (supported: %s, %s, %s)", hashname, cryptoutilSharedMagic.PBKDF2DefaultHashName, cryptoutilSharedMagic.PBKDF2SHA384HashName, cryptoutilSharedMagic.PBKDF2SHA512HashName)
	}

	// CRITICAL: Concatenate provided||pepper before PBKDF2 (OWASP requirement).
	// Pepper MUST be configured in params.Pepper (loaded from Docker/K8s secrets).
	// Empty pepper is allowed (backward compatibility with non-peppered hashes).
	pepperedSecret := provided + params.Pepper

	derived := pbkdf2.Key([]byte(pepperedSecret), salt, iter, len(expectedDK), hashFunc)

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

func parsePbkdf2Params(stored string) (string, string, int, []byte, []byte, error) {
	if stored == "" {
		return "", "", 0, nil, nil, errors.New("stored hash empty")
	}

	// ONLY support versioned format: {version}$hashname$iter$salt$dk
	if !strings.HasPrefix(stored, "{") {
		return "", "", 0, nil, nil, errors.New("unsupported hash format: must use versioned format {version}$hashname$iter$salt$dk")
	}

	parts := strings.Split(stored, "$")
	if len(parts) != cryptoutilSharedMagic.PBKDF2VersionedFormatParts {
		return "", "", 0, nil, nil, fmt.Errorf("invalid versioned hash format (expected %d parts, got %d)", cryptoutilSharedMagic.PBKDF2VersionedFormatParts, len(parts))
	}

	// Extract version from {1} format
	versionPart := parts[0]
	if !strings.HasPrefix(versionPart, "{") || !strings.HasSuffix(versionPart, "}") {
		return "", "", 0, nil, nil, fmt.Errorf("invalid version format: must be {version}")
	}

	version := versionPart[1 : len(versionPart)-1]
	hashname := parts[1]

	var iter int
	if _, err := fmt.Sscanf(parts[2], "%d", &iter); err != nil || iter <= 0 {
		return "", "", 0, nil, nil, fmt.Errorf("invalid iterations: %w", err)
	}

	saltB64 := parts[3]
	dkB64 := parts[4]

	salt, err := base64.RawStdEncoding.DecodeString(saltB64)
	if err != nil {
		return "", "", 0, nil, nil, fmt.Errorf("invalid salt encoding: %w", err)
	}

	expectedDK, err := base64.RawStdEncoding.DecodeString(dkB64)
	if err != nil {
		return "", "", 0, nil, nil, fmt.Errorf("invalid dk encoding: %w", err)
	}

	return version, hashname, iter, salt, expectedDK, nil
}
