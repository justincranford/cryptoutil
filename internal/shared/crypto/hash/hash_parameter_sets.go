// Copyright (c) 2025 Justin Cranford

package hash

import (
	sha256 "crypto/sha256"
	sha512 "crypto/sha512"
	"fmt"

	cryptoutilSharedCryptoDigests "cryptoutil/internal/shared/crypto/digests"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	hashPBKDF2WithParamsFn       = cryptoutilSharedCryptoDigests.PBKDF2WithParams
	hashVerifySecretWithParamsFn = cryptoutilSharedCryptoDigests.VerifySecretWithParams
)

// HashSecretPBKDF2 returns a formatted PBKDF2 hash string using default parameter set (version "1").
// Format: {1}$pbkdf2-sha256$iter$base64(salt)$base64(dk).
//
// CRITICAL: This function does NOT load pepper from Docker secrets.
// For production use with OWASP-compliant peppered hashing:
//  1. Load pepper using ConfigurePeppers(registry, pepperConfigs)
//  2. Get parameter set: params := registry.GetDefaultParameterSet()
//  3. Hash with pepper: HashSecretPBKDF2WithParams(secret, params)
func HashSecretPBKDF2(secret string) (string, error) {
	hash, err := hashPBKDF2WithParamsFn(secret, DefaultPBKDF2ParameterSet())
	if err != nil {
		return "", fmt.Errorf("failed to generate PBKDF2 hash: %w", err)
	}

	return hash, nil
}

// HashSecretPBKDF2WithParams returns a formatted PBKDF2 hash string using specified parameter set.
// Format: {version}$pbkdf2-sha256$iter$base64(salt)$base64(dk).
//
// CRITICAL: OWASP MANDATORY requirement - params MUST include pepper loaded from Docker/K8s secrets.
// Pattern: PBKDF2(password||pepper, salt, iterations, keyLength)
// Reference: https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#peppering
func HashSecretPBKDF2WithParams(secret string, params *cryptoutilSharedCryptoDigests.PBKDF2Params) (string, error) {
	hash, err := hashPBKDF2WithParamsFn(secret, params)
	if err != nil {
		return "", fmt.Errorf("failed to generate PBKDF2 hash: %w", err)
	}

	return hash, nil
}

// VerifySecretPBKDF2WithParams verifies a stored hash against a provided secret using specified parameter set.
// CRITICAL: params MUST include pepper loaded from Docker/K8s secrets (same pepper used during hashing).
// Pattern: PBKDF2(password||pepper, salt, iterations, keyLength).
func VerifySecretPBKDF2WithParams(stored, provided string, params *cryptoutilSharedCryptoDigests.PBKDF2Params) (bool, error) {
	valid, err := hashVerifySecretWithParamsFn(stored, provided, params)
	if err != nil {
		return false, fmt.Errorf("failed to verify secret: %w", err)
	}

	return valid, nil
}

// DefaultPBKDF2ParameterSet returns the default PBKDF2-HMAC-SHA256 parameter set (version "1").
//
// Parameters:
// - 600,000 iterations (OWASP 2023 recommendation for PBKDF2-HMAC-SHA256)
// - 32-byte salt (256 bits)
// - 32-byte key (256 bits)
// - SHA-256 hash function.
func DefaultPBKDF2ParameterSet() *cryptoutilSharedCryptoDigests.PBKDF2Params {
	return &cryptoutilSharedCryptoDigests.PBKDF2Params{
		Version:    "1",
		HashName:   cryptoutilSharedMagic.PBKDF2DefaultHashName,
		Iterations: cryptoutilSharedMagic.PBKDF2DefaultIterations,
		SaltLength: cryptoutilSharedMagic.PBKDF2DefaultSaltBytes,
		KeyLength:  cryptoutilSharedMagic.PBKDF2DerivedKeyLength,
		HashFunc:   sha256.New,
	}
}

// PBKDF2ParameterSetV1 returns version "1" parameter set (same as default).
func PBKDF2ParameterSetV1() *cryptoutilSharedCryptoDigests.PBKDF2Params {
	return DefaultPBKDF2ParameterSet()
}

// PBKDF2ParameterSetV2 returns version "2" parameter set (OWASP 2021 standard).
//
// Parameters:
// - 310,000 iterations (NIST SP 800-63B Rev. 3 recommendation, 2021)
// - 32-byte salt (256 bits)
// - 32-byte key (256 bits)
// - SHA-256 hash function.
func PBKDF2ParameterSetV2() *cryptoutilSharedCryptoDigests.PBKDF2Params {
	return &cryptoutilSharedCryptoDigests.PBKDF2Params{
		Version:    "2",
		HashName:   cryptoutilSharedMagic.PBKDF2DefaultHashName,
		Iterations: cryptoutilSharedMagic.PBKDF2V2Iterations,
		SaltLength: cryptoutilSharedMagic.PBKDF2DefaultSaltBytes,
		KeyLength:  cryptoutilSharedMagic.PBKDF2DerivedKeyLength,
		HashFunc:   sha256.New,
	}
}

// PBKDF2ParameterSetV3 returns version "3" parameter set (OWASP 2017 legacy).
//
// Parameters:
// - 1,000 iterations (NIST 2017 minimum, for legacy password migration support)
// - 32-byte salt (256 bits)
// - 32-byte key (256 bits)
// - SHA-256 hash function.
//
// Note: This low iteration count is ONLY for migrating legacy passwords from systems
// that used weak hashing (e.g., old databases). New passwords MUST use V1 (600k) or V2 (310k).
func PBKDF2ParameterSetV3() *cryptoutilSharedCryptoDigests.PBKDF2Params {
	return &cryptoutilSharedCryptoDigests.PBKDF2Params{
		Version:    "3",
		HashName:   cryptoutilSharedMagic.PBKDF2DefaultHashName,
		Iterations: cryptoutilSharedMagic.PBKDF2V3Iterations,
		SaltLength: cryptoutilSharedMagic.PBKDF2DefaultSaltBytes,
		KeyLength:  cryptoutilSharedMagic.PBKDF2DerivedKeyLength,
		HashFunc:   sha256.New,
	}
}

// PBKDF2SHA384ParameterSetV1 returns SHA-384 version "1" parameter set (OWASP 2023).
//
// Parameters:
// - 600,000 iterations (OWASP 2023 recommendation)
// - 32-byte salt (256 bits)
// - 48-byte key (384 bits for SHA-384 output)
// - SHA-384 hash function.
func PBKDF2SHA384ParameterSetV1() *cryptoutilSharedCryptoDigests.PBKDF2Params {
	return &cryptoutilSharedCryptoDigests.PBKDF2Params{
		Version:    "1",
		HashName:   cryptoutilSharedMagic.PBKDF2SHA384HashName,
		Iterations: cryptoutilSharedMagic.PBKDF2DefaultIterations,
		SaltLength: cryptoutilSharedMagic.PBKDF2DefaultSaltBytes,
		KeyLength:  cryptoutilSharedMagic.PBKDF2SHA384HashBytes,
		HashFunc:   sha512.New384,
	}
}

// PBKDF2SHA384ParameterSetV2 returns SHA-384 version "2" parameter set (OWASP 2021).
//
// Parameters:
// - 310,000 iterations (NIST SP 800-63B Rev. 3 recommendation, 2021)
// - 32-byte salt (256 bits)
// - 48-byte key (384 bits for SHA-384 output)
// - SHA-384 hash function.
func PBKDF2SHA384ParameterSetV2() *cryptoutilSharedCryptoDigests.PBKDF2Params {
	return &cryptoutilSharedCryptoDigests.PBKDF2Params{
		Version:    "2",
		HashName:   cryptoutilSharedMagic.PBKDF2SHA384HashName,
		Iterations: cryptoutilSharedMagic.PBKDF2V2Iterations,
		SaltLength: cryptoutilSharedMagic.PBKDF2DefaultSaltBytes,
		KeyLength:  cryptoutilSharedMagic.PBKDF2SHA384HashBytes,
		HashFunc:   sha512.New384,
	}
}

// PBKDF2SHA384ParameterSetV3 returns SHA-384 version "3" parameter set (OWASP 2017 legacy).
//
// Parameters:
// - 1,000 iterations (NIST 2017 minimum, for legacy password migration)
// - 32-byte salt (256 bits)
// - 48-byte key (384 bits for SHA-384 output)
// - SHA-384 hash function.
func PBKDF2SHA384ParameterSetV3() *cryptoutilSharedCryptoDigests.PBKDF2Params {
	return &cryptoutilSharedCryptoDigests.PBKDF2Params{
		Version:    "3",
		HashName:   cryptoutilSharedMagic.PBKDF2SHA384HashName,
		Iterations: cryptoutilSharedMagic.PBKDF2V3Iterations,
		SaltLength: cryptoutilSharedMagic.PBKDF2DefaultSaltBytes,
		KeyLength:  cryptoutilSharedMagic.PBKDF2SHA384HashBytes,
		HashFunc:   sha512.New384,
	}
}

// PBKDF2SHA512ParameterSetV1 returns SHA-512 version "1" parameter set (OWASP 2023).
//
// Parameters:
// - 600,000 iterations (OWASP 2023 recommendation)
// - 32-byte salt (256 bits)
// - 64-byte key (512 bits for SHA-512 output)
// - SHA-512 hash function.
func PBKDF2SHA512ParameterSetV1() *cryptoutilSharedCryptoDigests.PBKDF2Params {
	return &cryptoutilSharedCryptoDigests.PBKDF2Params{
		Version:    "1",
		HashName:   cryptoutilSharedMagic.PBKDF2SHA512HashName,
		Iterations: cryptoutilSharedMagic.PBKDF2DefaultIterations,
		SaltLength: cryptoutilSharedMagic.PBKDF2DefaultSaltBytes,
		KeyLength:  cryptoutilSharedMagic.PBKDF2SHA512HashBytes,
		HashFunc:   sha512.New,
	}
}

// PBKDF2SHA512ParameterSetV2 returns SHA-512 version "2" parameter set (OWASP 2021).
//
// Parameters:
// - 310,000 iterations (NIST SP 800-63B Rev. 3 recommendation, 2021)
// - 32-byte salt (256 bits)
// - 64-byte key (512 bits for SHA-512 output)
// - SHA-512 hash function.
func PBKDF2SHA512ParameterSetV2() *cryptoutilSharedCryptoDigests.PBKDF2Params {
	return &cryptoutilSharedCryptoDigests.PBKDF2Params{
		Version:    "2",
		HashName:   cryptoutilSharedMagic.PBKDF2SHA512HashName,
		Iterations: cryptoutilSharedMagic.PBKDF2V2Iterations,
		SaltLength: cryptoutilSharedMagic.PBKDF2DefaultSaltBytes,
		KeyLength:  cryptoutilSharedMagic.PBKDF2SHA512HashBytes,
		HashFunc:   sha512.New,
	}
}

// PBKDF2SHA512ParameterSetV3 returns SHA-512 version "3" parameter set (OWASP 2017 legacy).
//
// Parameters:
// - 1,000 iterations (NIST 2017 minimum, for legacy password migration)
// - 32-byte salt (256 bits)
// - 64-byte key (512 bits for SHA-512 output)
// - SHA-512 hash function.
func PBKDF2SHA512ParameterSetV3() *cryptoutilSharedCryptoDigests.PBKDF2Params {
	return &cryptoutilSharedCryptoDigests.PBKDF2Params{
		Version:    "3",
		HashName:   cryptoutilSharedMagic.PBKDF2SHA512HashName,
		Iterations: cryptoutilSharedMagic.PBKDF2V3Iterations,
		SaltLength: cryptoutilSharedMagic.PBKDF2DefaultSaltBytes,
		KeyLength:  cryptoutilSharedMagic.PBKDF2SHA512HashBytes,
		HashFunc:   sha512.New,
	}
}
