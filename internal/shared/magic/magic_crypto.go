// Copyright (c) 2025 Justin Cranford
//
//

package magic

import (
	"crypto/sha256"
	"crypto/sha512"
	"hash"
	"time"
)

// DefaultPoolConfig defines the worker and pool size configuration for key generation pools.
type DefaultPoolConfig struct {
	NumWorkers uint32
	MaxSize    uint32
}

// Cryptographic algorithm and key constants.
// This file contains all crypto-related magic values used throughout the application.

const (
	// EdDSA curve names.
	EdCurveEd448   = "Ed448"
	EdCurveEd25519 = "Ed25519"

	// Elliptic curve names.
	ECCurveP256 = "P256"
	ECCurveP384 = "P384"
	ECCurveP521 = "P521"

	// SHA2 digest algorithm names.
	SHA512 = "SHA512"
	SHA384 = "SHA384"
	SHA256 = "SHA256"
	SHA224 = "SHA224"

	// RSA key sizes in bits.
	RSAKeySize2048 = 2048
	RSAKeySize3072 = 3072
	RSAKeySize4096 = 4096

	// Symmetric key sizes in bits.
	SymmetricKeySize128 = 128
	SymmetricKeySize192 = 192
	SymmetricKeySize256 = 256
	SymmetricKeySize384 = 384
	SymmetricKeySize512 = 512

	// JWE key sizes in bits.
	JWEA128KeySize = SymmetricKeySize128
	JWEA192KeySize = SymmetricKeySize192
	JWEA256KeySize = SymmetricKeySize256
	JWEA384KeySize = SymmetricKeySize384
	JWEA512KeySize = SymmetricKeySize512

	// Secret generation byte lengths.
	SecretGenerationDefaultByteLength = 32

	// AES key sizes in bits.
	AESKeySize128 = SymmetricKeySize128
	AESKeySize192 = SymmetricKeySize192
	AESKeySize256 = SymmetricKeySize256

	// AES HMAC-SHA2 key sizes in bits.
	AESHSKeySize256 = SymmetricKeySize256
	AESHSKeySize384 = SymmetricKeySize384
	AESHSKeySize512 = SymmetricKeySize512

	// HMAC key sizes in bits.
	HMACKeySize256 = SymmetricKeySize256
	HMACKeySize384 = SymmetricKeySize384
	HMACKeySize512 = SymmetricKeySize512

	// Minimum HMAC key size in bits.
	MinHMACKeySize = SymmetricKeySize256

	// HKDF test constants.
	HKDFSHA224OutputLength = 28
	HKDFSHA256OutputLength = 32
	HKDFSHA384OutputLength = 48
	HKDFSHA512OutputLength = 64
	HKDFMaxMultiplier      = 255
	HKDFSHA224MaxLength    = 7140  // 255 * 28
	HKDFSHA256MaxLength    = 8160  // 255 * 32
	HKDFSHA384MaxLength    = 12240 // 255 * 48
	HKDFSHA512MaxLength    = 16320 // 255 * 64

	// JWK generation pool sizes (min, max) by algorithm type.

	// MaxPoolLifetimeValuesInt64 - Maximum int64 value (= 2^63-1 = 9,223,372,036,854,775,807).
	MaxPoolLifetimeValuesInt64 = int64(^uint64(0) >> 1)

	// MaxPoolLifetimeValues - Max int64 as uint64.
	MaxPoolLifetimeValues = uint64(MaxPoolLifetimeValuesInt64)
	// MaxPoolLifetimeDuration - Max int64 as nanoseconds (= 292.47 years).
	MaxPoolLifetimeDuration = time.Duration(MaxPoolLifetimeValuesInt64)

	// PoolMaintenanceInterval - Ticker interval for periodic pool maintenance checks.
	PoolMaintenanceInterval = 500 * time.Millisecond

	// TestPoolMaxSize - Maximum pool size for test configurations.
	TestPoolMaxSize = 3

	// PBKDF2 configuration - FIPS 140-3 approved password hashing.
	PBKDF2Prefix            = "pbkdf2-sha256" // Hash format prefix for PBKDF2-HMAC-SHA256 hashes.
	PBKDF2DerivedKeyLength  = 32              // Derived key length in bytes (32 = 256 bits).
	PBKDF2DefaultHashName   = "pbkdf2-sha256" // Algorithm name for PBKDF2 SHA-256 (default).
	PBKDF2SHA384HashName    = "pbkdf2-sha384" // Algorithm name for PBKDF2 SHA-384.
	PBKDF2SHA512HashName    = "pbkdf2-sha512" // Algorithm name for PBKDF2 SHA-512.
	PBKDF2DefaultAlgorithm  = "SHA-256"       // Default PRF algorithm for PBKDF2.
	PBKDF2DefaultSaltBytes  = 32              // Salt length in bytes (32 = 256 bits).
	PBKDF2DefaultHashBytes  = 32              // Derived key length in bytes (32 = 256 bits for SHA-256).
	PBKDF2SHA384HashBytes   = 48              // Derived key length in bytes (48 = 384 bits for SHA-384).
	PBKDF2SHA512HashBytes   = 64              // Derived key length in bytes (64 = 512 bits for SHA-512).
	PBKDF2MinIterations     = 210000          // OWASP minimum iterations for PBKDF2-HMAC-SHA256 (2023).

	// PBKDF2 iteration counts - OWASP/NIST historical standards.
	PBKDF2DefaultIterations = 600000 // Version 1 (2023): OWASP current recommendation.
	PBKDF2V2Iterations      = 310000 // Version 2 (2021): NIST SP 800-63B Rev. 3 recommendation.
	PBKDF2V3Iterations      = 1000   // Version 3 (2017): Legacy/migration support (NIST 2017 minimum).

	// PBKDF2 hash format constants.
	PBKDF2VersionedFormatParts = 5 // Number of parts in versioned hash format: {version}$hashname$iter$salt$dk.

	// HKDF hash name constants - for hash format strings.
	HKDFHashName          = "hkdf-sha256"           // HKDF-SHA256 with random salt (non-deterministic).
	HKDFFixedLowHashName  = "hkdf-sha256-fixed"     // HKDF-SHA256 with fixed info (deterministic, low-entropy).
	HKDFFixedHighHashName = "hkdf-sha256-fixed-high" // HKDF-SHA256 with fixed info (deterministic, high-entropy).

	// HKDF hash format constants.
	HKDFDelimiter = "$" // Delimiter for HKDF hash format parts.
)

// HKDF deterministic hashing constants (fixed info parameters for determinism).
var (
	HKDFFixedInfoLowEntropy  = []byte("cryptoutil-hkdf-low-entropy-v1")  // Fixed info for low-entropy deterministic hashing.
	HKDFFixedInfoHighEntropy = []byte("cryptoutil-hkdf-high-entropy-v1") // Fixed info for high-entropy deterministic hashing.
)

// PBKDF2HashFunction returns the hash function for the given algorithm name.
func PBKDF2HashFunction(algorithm string) func() hash.Hash {
	switch algorithm {
	case "SHA-256", "sha256", "SHA256":
		return sha256.New
	case "SHA-512", "sha512", "SHA512":
		return sha512.New
	case "SHA-384", "sha384", "SHA384":
		return sha512.New384
	default:
		return sha256.New
	}
}

var (
	DefaultPoolConfigRSA4096     = DefaultPoolConfig{NumWorkers: 9, MaxSize: 9}  //nolint:mnd
	DefaultPoolConfigRSA3072     = DefaultPoolConfig{NumWorkers: 6, MaxSize: 6}  //nolint:mnd
	DefaultPoolConfigRSA2048     = DefaultPoolConfig{NumWorkers: 3, MaxSize: 3}  //nolint:mnd
	DefaultPoolConfigECDSAP521   = DefaultPoolConfig{NumWorkers: 3, MaxSize: 9}  //nolint:mnd
	DefaultPoolConfigECDSAP384   = DefaultPoolConfig{NumWorkers: 2, MaxSize: 6}  //nolint:mnd
	DefaultPoolConfigECDSAP256   = DefaultPoolConfig{NumWorkers: 1, MaxSize: 3}  //nolint:mnd
	DefaultPoolConfigECDHP521    = DefaultPoolConfig{NumWorkers: 3, MaxSize: 9}  //nolint:mnd
	DefaultPoolConfigECDHP384    = DefaultPoolConfig{NumWorkers: 2, MaxSize: 6}  //nolint:mnd
	DefaultPoolConfigECDHP256    = DefaultPoolConfig{NumWorkers: 1, MaxSize: 3}  //nolint:mnd
	DefaultPoolConfigED25519     = DefaultPoolConfig{NumWorkers: 1, MaxSize: 2}  //nolint:mnd
	DefaultPoolConfigAES256      = DefaultPoolConfig{NumWorkers: 3, MaxSize: 9}  //nolint:mnd
	DefaultPoolConfigAES192      = DefaultPoolConfig{NumWorkers: 2, MaxSize: 6}  //nolint:mnd
	DefaultPoolConfigAES128      = DefaultPoolConfig{NumWorkers: 1, MaxSize: 3}  //nolint:mnd
	DefaultPoolConfigAES256HS512 = DefaultPoolConfig{NumWorkers: 3, MaxSize: 6}  //nolint:mnd
	DefaultPoolConfigAES192HS384 = DefaultPoolConfig{NumWorkers: 2, MaxSize: 4}  //nolint:mnd
	DefaultPoolConfigAES128HS256 = DefaultPoolConfig{NumWorkers: 1, MaxSize: 2}  //nolint:mnd
	DefaultPoolConfigHMAC512     = DefaultPoolConfig{NumWorkers: 3, MaxSize: 9}  //nolint:mnd
	DefaultPoolConfigHMAC384     = DefaultPoolConfig{NumWorkers: 2, MaxSize: 6}  //nolint:mnd
	DefaultPoolConfigHMAC256     = DefaultPoolConfig{NumWorkers: 1, MaxSize: 3}  //nolint:mnd
	DefaultPoolConfigUUIDv7      = DefaultPoolConfig{NumWorkers: 2, MaxSize: 20} //nolint:mnd
)
