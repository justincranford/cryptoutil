// Copyright (c) 2025 Justin Cranford

package hash

import (
	"crypto/sha256"
	"crypto/sha512"

	cryptoutilDigests "cryptoutil/internal/shared/crypto/digests"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// HashSecretPBKDF2 returns a formatted PBKDF2 hash string using default parameter set (version "1").
// Format: {1}$pbkdf2-sha256$iter$base64(salt)$base64(dk).
func HashSecretPBKDF2(secret string) (string, error) {
	return cryptoutilDigests.PBKDF2WithParams(secret, DefaultPBKDF2ParameterSet())
}

// DefaultPBKDF2ParameterSet returns the default PBKDF2-HMAC-SHA256 parameter set (version "1").
//
// Parameters:
// - 600,000 iterations (OWASP 2023 recommendation for PBKDF2-HMAC-SHA256)
// - 32-byte salt (256 bits)
// - 32-byte key (256 bits)
// - SHA-256 hash function.
func DefaultPBKDF2ParameterSet() *cryptoutilDigests.PBKDF2Params {
	return &cryptoutilDigests.PBKDF2Params{
		Version:    "1",
		HashName:   cryptoutilMagic.PBKDF2DefaultHashName,
		Iterations: cryptoutilMagic.PBKDF2DefaultIterations,
		SaltLength: cryptoutilMagic.PBKDF2DefaultSaltBytes,
		KeyLength:  cryptoutilMagic.PBKDF2DerivedKeyLength,
		HashFunc:   sha256.New,
	}
}

// PBKDF2ParameterSetV1 returns version "1" parameter set (same as default).
func PBKDF2ParameterSetV1() *cryptoutilDigests.PBKDF2Params {
	return DefaultPBKDF2ParameterSet()
}

// PBKDF2ParameterSetV2 returns version "2" parameter set (OWASP 2021 standard).
//
// Parameters:
// - 310,000 iterations (NIST SP 800-63B Rev. 3 recommendation, 2021)
// - 32-byte salt (256 bits)
// - 32-byte key (256 bits)
// - SHA-256 hash function.
func PBKDF2ParameterSetV2() *cryptoutilDigests.PBKDF2Params {
	return &cryptoutilDigests.PBKDF2Params{
		Version:    "2",
		HashName:   cryptoutilMagic.PBKDF2DefaultHashName,
		Iterations: cryptoutilMagic.PBKDF2V2Iterations,
		SaltLength: cryptoutilMagic.PBKDF2DefaultSaltBytes,
		KeyLength:  cryptoutilMagic.PBKDF2DerivedKeyLength,
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
func PBKDF2ParameterSetV3() *cryptoutilDigests.PBKDF2Params {
	return &cryptoutilDigests.PBKDF2Params{
		Version:    "3",
		HashName:   cryptoutilMagic.PBKDF2DefaultHashName,
		Iterations: cryptoutilMagic.PBKDF2V3Iterations,
		SaltLength: cryptoutilMagic.PBKDF2DefaultSaltBytes,
		KeyLength:  cryptoutilMagic.PBKDF2DerivedKeyLength,
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
func PBKDF2SHA384ParameterSetV1() *cryptoutilDigests.PBKDF2Params {
	return &cryptoutilDigests.PBKDF2Params{
		Version:    "1",
		HashName:   cryptoutilMagic.PBKDF2SHA384HashName,
		Iterations: cryptoutilMagic.PBKDF2DefaultIterations,
		SaltLength: cryptoutilMagic.PBKDF2DefaultSaltBytes,
		KeyLength:  cryptoutilMagic.PBKDF2SHA384HashBytes,
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
func PBKDF2SHA384ParameterSetV2() *cryptoutilDigests.PBKDF2Params {
	return &cryptoutilDigests.PBKDF2Params{
		Version:    "2",
		HashName:   cryptoutilMagic.PBKDF2SHA384HashName,
		Iterations: cryptoutilMagic.PBKDF2V2Iterations,
		SaltLength: cryptoutilMagic.PBKDF2DefaultSaltBytes,
		KeyLength:  cryptoutilMagic.PBKDF2SHA384HashBytes,
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
func PBKDF2SHA384ParameterSetV3() *cryptoutilDigests.PBKDF2Params {
	return &cryptoutilDigests.PBKDF2Params{
		Version:    "3",
		HashName:   cryptoutilMagic.PBKDF2SHA384HashName,
		Iterations: cryptoutilMagic.PBKDF2V3Iterations,
		SaltLength: cryptoutilMagic.PBKDF2DefaultSaltBytes,
		KeyLength:  cryptoutilMagic.PBKDF2SHA384HashBytes,
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
func PBKDF2SHA512ParameterSetV1() *cryptoutilDigests.PBKDF2Params {
	return &cryptoutilDigests.PBKDF2Params{
		Version:    "1",
		HashName:   cryptoutilMagic.PBKDF2SHA512HashName,
		Iterations: cryptoutilMagic.PBKDF2DefaultIterations,
		SaltLength: cryptoutilMagic.PBKDF2DefaultSaltBytes,
		KeyLength:  cryptoutilMagic.PBKDF2SHA512HashBytes,
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
func PBKDF2SHA512ParameterSetV2() *cryptoutilDigests.PBKDF2Params {
	return &cryptoutilDigests.PBKDF2Params{
		Version:    "2",
		HashName:   cryptoutilMagic.PBKDF2SHA512HashName,
		Iterations: cryptoutilMagic.PBKDF2V2Iterations,
		SaltLength: cryptoutilMagic.PBKDF2DefaultSaltBytes,
		KeyLength:  cryptoutilMagic.PBKDF2SHA512HashBytes,
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
func PBKDF2SHA512ParameterSetV3() *cryptoutilDigests.PBKDF2Params {
	return &cryptoutilDigests.PBKDF2Params{
		Version:    "3",
		HashName:   cryptoutilMagic.PBKDF2SHA512HashName,
		Iterations: cryptoutilMagic.PBKDF2V3Iterations,
		SaltLength: cryptoutilMagic.PBKDF2DefaultSaltBytes,
		KeyLength:  cryptoutilMagic.PBKDF2SHA512HashBytes,
		HashFunc:   sha512.New,
	}
}
