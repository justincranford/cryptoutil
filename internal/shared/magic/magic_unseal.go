// Copyright (c) 2025 Justin Cranford
//
//

package magic

import "time"

// Cryptographic algorithm and key constants.
// This file contains all crypto-related magic values used throughout the application.

const (
	// DefaultUnsealModeSysInfo - Sysinfo unseal mode.
	DefaultUnsealModeSysInfo = "sysinfo"
	// DefaultSysInfoAllTimeout - 10 seconds duration, rate limit maximum.
	DefaultSysInfoAllTimeout = 10 * time.Second
	// DefaultSysInfoCPUTimeout - Default system timeout duration.
	DefaultSysInfoCPUTimeout = 10 * time.Second
	// DefaultSysInfoMemoryTimeout is the default system memory timeout duration.
	DefaultSysInfoMemoryTimeout = 5 * time.Second
	// DefaultSysInfoHostTimeout is the default system host timeout duration.
	DefaultSysInfoHostTimeout = 5 * time.Second

	// DefaultMaxUnsealFiles - Maximum number of files allowed.
	DefaultMaxUnsealFiles = 10
	// DefaultMaxBytesPerUnsealFile - Maximum bytes per file allowed.
	DefaultMaxBytesPerUnsealFile = 10 << 20 // 10MB

	// MaxUnsealSharedSecrets - Maximum number of shared secrets allowed.
	MaxUnsealSharedSecrets = 256
	// MinSharedSecretLength - Minimum shared secret length in bytes.
	MinSharedSecretLength = 32
	// MaxSharedSecretLength - Maximum shared secret length in bytes.
	MaxSharedSecretLength = 64

	// DerivedKeySizeBytes - Derived key size in bytes.
	DerivedKeySizeBytes = 32

	// UUIDBytesLength - UUID byte length (16 bytes for standard UUID).
	UUIDBytesLength = 16

	// RandomKeySizeBytes - Random bytes length for dev mode.
	RandomKeySizeBytes = 32
)

// DefaultUnsealFiles - Default unseal files slice.
var DefaultUnsealFiles = []string{}

// Unseal JWK derivation constants for deterministic key ID generation.
var (
	// FixedContextForDerivedKid - Derive context for key identifier JWKs.
	FixedContextForDerivedKid = []byte("fixed context for derive unseal JWKs key identifier v1")

	// FixedIKMForDerivedKid - Fixed derive bytes for key identifier secret.
	FixedIKMForDerivedKid = []byte("fixed IKM for derive bytes for key identifier secret v1")
	// FixedSaltForDerivedKid - Fixed derive bytes for key identifier salt.
	FixedSaltForDerivedKid = []byte("fixed IKM for derive bytes for key identifier salt v1")

	// FixedContextForDerivedSecret - Derive context for key material JWKs.
	FixedContextForDerivedSecret = []byte("fixed context for derive unseal JWKs key material v1")

	// FixedIKMForDerivedSecret - Fixed derive bytes for key material secret.
	FixedIKMForDerivedSecret = []byte("fixed IKM for derive bytes for key material secret v1")
	// FixedSaltForDerivedSecret - Fixed derive bytes for key material salt.
	FixedSaltForDerivedSecret = []byte("fixed IKM for derive bytes for key material salt v1")
)
