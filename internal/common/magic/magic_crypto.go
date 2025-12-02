// Copyright (c) 2025 Justin Cranford
//
//

package magic

import "time"

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

	// PBKDF2 configuration (Session 5 Q12: SHA-256, 600K iterations, 32-byte salt).
	PBKDF2Prefix            = "pbkdf2-sha256" // Hash format prefix for PBKDF2-HMAC-SHA256 hashes.
	PBKDF2DerivedKeyLength  = 32              // Derived key length in bytes (32 = 256 bits).
	PBKDF2DefaultHashName   = "pbkdf2"        // Algorithm name for PBKDF2.
	PBKDF2DefaultAlgorithm  = "SHA-256"       // Default PRF algorithm for PBKDF2.
	PBKDF2DefaultSaltBytes  = 32              // Salt length in bytes (32 = 256 bits, Session 5 Q12).
	PBKDF2DefaultHashBytes  = 32              // Derived key length in bytes (32 = 256 bits).
	PBKDF2MinIterations     = 210000          // OWASP minimum iterations for PBKDF2-HMAC-SHA256 (2023).
	PBKDF2DefaultIterations = 600000          // Default iteration count (Session 5 Q12: 600K).
)

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
