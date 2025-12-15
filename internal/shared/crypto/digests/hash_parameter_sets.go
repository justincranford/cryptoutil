// Copyright (c) 2025 Justin Cranford

package digests

import (
	"crypto/sha256"
	"hash"

	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// PBKDF2ParameterSet defines parameters for PBKDF2-HMAC hashing.
type PBKDF2ParameterSet struct {
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
}

// DefaultPBKDF2ParameterSet returns the default PBKDF2-HMAC-SHA256 parameter set (version "1").
//
// Parameters:
// - 600,000 iterations (OWASP 2023 recommendation for PBKDF2-HMAC-SHA256)
// - 32-byte salt (256 bits)
// - 32-byte key (256 bits)
// - SHA-256 hash function.
func DefaultPBKDF2ParameterSet() PBKDF2ParameterSet {
	return PBKDF2ParameterSet{
		Version:    "1",
		HashName:   cryptoutilMagic.PBKDF2DefaultHashName,
		Iterations: cryptoutilMagic.PBKDF2DefaultIterations,
		SaltLength: cryptoutilMagic.PBKDF2DefaultSaltBytes,
		KeyLength:  cryptoutilMagic.PBKDF2DerivedKeyLength,
		HashFunc:   sha256.New,
	}
}

// PBKDF2ParameterSetV1 returns version "1" parameter set (same as default).
func PBKDF2ParameterSetV1() PBKDF2ParameterSet {
	return DefaultPBKDF2ParameterSet()
}

// PBKDF2ParameterSetV2 returns version "2" parameter set with increased iterations.
//
// Parameters:
// - 1,000,000 iterations (future-proof against hardware improvements)
// - 32-byte salt (256 bits)
// - 32-byte key (256 bits)
// - SHA-256 hash function.
func PBKDF2ParameterSetV2() PBKDF2ParameterSet {
	return PBKDF2ParameterSet{
		Version:    "2",
		HashName:   cryptoutilMagic.PBKDF2DefaultHashName,
		Iterations: cryptoutilMagic.PBKDF2V2Iterations,
		SaltLength: cryptoutilMagic.PBKDF2DefaultSaltBytes,
		KeyLength:  cryptoutilMagic.PBKDF2DerivedKeyLength,
		HashFunc:   sha256.New,
	}
}

// PBKDF2ParameterSetV3 returns version "3" parameter set with maximum security.
//
// Parameters:
// - 2,000,000 iterations (defense against future GPU/ASIC attacks)
// - 32-byte salt (256 bits)
// - 32-byte key (256 bits)
// - SHA-256 hash function.
func PBKDF2ParameterSetV3() PBKDF2ParameterSet {
	return PBKDF2ParameterSet{
		Version:    "3",
		HashName:   cryptoutilMagic.PBKDF2DefaultHashName,
		Iterations: cryptoutilMagic.PBKDF2V3Iterations,
		SaltLength: cryptoutilMagic.PBKDF2DefaultSaltBytes,
		KeyLength:  cryptoutilMagic.PBKDF2DerivedKeyLength,
		HashFunc:   sha256.New,
	}
}
