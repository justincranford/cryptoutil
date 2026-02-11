// Copyright (c) 2025 Justin Cranford
//
//

package magic

import (
	sha256 "crypto/sha256"
	sha512 "crypto/sha512"
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
	// EdCurveEd448 is the Ed448 curve name.
	EdCurveEd448 = "Ed448"
	// EdCurveEd25519 is the Ed25519 curve name.
	EdCurveEd25519 = "Ed25519"

	// ECCurveP256 is the P-256 elliptic curve name.
	ECCurveP256 = "P256"
	// ECCurveP384 is the P-384 elliptic curve name.
	ECCurveP384 = "P384"
	// ECCurveP521 is the P-521 elliptic curve name.
	ECCurveP521 = "P521"

	// SHA512 is the SHA-512 digest algorithm name.
	SHA512 = "SHA512"
	// SHA384 is the SHA-384 digest algorithm name.
	SHA384 = "SHA384"
	// SHA256 is the SHA-256 digest algorithm name.
	SHA256 = "SHA256"
	// SHA224 is the SHA-224 digest algorithm name.
	SHA224 = "SHA224"

	// RSAKeySize2048 is the RSA 2048-bit key size constant.
	RSAKeySize2048 = 2048
	// RSAKeySize3072 is the RSA 3072-bit key size constant.
	RSAKeySize3072 = 3072
	// RSAKeySize4096 is the RSA 4096-bit key size constant.
	RSAKeySize4096 = 4096

	// SymmetricKeySize128 is the 128-bit symmetric key size constant.
	SymmetricKeySize128 = 128
	// SymmetricKeySize192 is the 192-bit symmetric key size constant.
	SymmetricKeySize192 = 192
	// SymmetricKeySize256 is the 256-bit symmetric key size constant.
	SymmetricKeySize256 = 256
	// SymmetricKeySize384 is the 384-bit symmetric key size constant.
	SymmetricKeySize384 = 384
	// SymmetricKeySize512 is the 512-bit symmetric key size constant.
	SymmetricKeySize512 = 512

	// JWEA128KeySize is the JWE 128-bit key size constant.
	JWEA128KeySize = SymmetricKeySize128
	// JWEA192KeySize is the JWE 192-bit key size constant.
	JWEA192KeySize = SymmetricKeySize192
	// JWEA256KeySize is the JWE 256-bit key size constant.
	JWEA256KeySize = SymmetricKeySize256
	// JWEA384KeySize is the JWE 384-bit key size constant.
	JWEA384KeySize = SymmetricKeySize384
	// JWEA512KeySize is the JWE 512-bit key size constant.
	JWEA512KeySize = SymmetricKeySize512

	// SecretGenerationDefaultByteLength is the default secret generation byte length.
	SecretGenerationDefaultByteLength = 32

	// AESKeySize128 is the AES 128-bit key size constant.
	AESKeySize128 = SymmetricKeySize128
	// AESKeySize192 is the AES 192-bit key size constant.
	AESKeySize192 = SymmetricKeySize192
	// AESKeySize256 is the AES 256-bit key size constant.
	AESKeySize256 = SymmetricKeySize256

	// AESHSKeySize256 is the AES HMAC-SHA2 256-bit key size constant.
	AESHSKeySize256 = SymmetricKeySize256
	// AESHSKeySize384 is the AES HMAC-SHA2 384-bit key size constant.
	AESHSKeySize384 = SymmetricKeySize384
	// AESHSKeySize512 is the AES HMAC-SHA2 512-bit key size constant.
	AESHSKeySize512 = SymmetricKeySize512

	// HMACKeySize256 is the HMAC 256-bit key size constant.
	HMACKeySize256 = SymmetricKeySize256
	// HMACKeySize384 is the HMAC 384-bit key size constant.
	HMACKeySize384 = SymmetricKeySize384
	// HMACKeySize512 is the HMAC 512-bit key size constant.
	HMACKeySize512 = SymmetricKeySize512

	// MinHMACKeySize is the minimum HMAC key size in bits.
	MinHMACKeySize = SymmetricKeySize256

	// HKDFSHA224OutputLength is the HKDF-SHA224 output length constant.
	HKDFSHA224OutputLength = 28
	// HKDFSHA256OutputLength is the HKDF-SHA256 output length constant.
	HKDFSHA256OutputLength = 32
	// HKDFSHA384OutputLength is the HKDF-SHA384 output length constant.
	HKDFSHA384OutputLength = 48
	// HKDFSHA512OutputLength is the HKDF-SHA512 output length constant.
	HKDFSHA512OutputLength = 64
	// HKDFMaxMultiplier is the HKDF maximum multiplier constant.
	HKDFMaxMultiplier = 255
	// HKDFSHA224MaxLength is the HKDF-SHA224 maximum output length.
	HKDFSHA224MaxLength = 7140 // 255 * 28
	// HKDFSHA256MaxLength is the HKDF-SHA256 maximum output length.
	HKDFSHA256MaxLength = 8160 // 255 * 32
	// HKDFSHA384MaxLength is the HKDF-SHA384 maximum output length.
	HKDFSHA384MaxLength = 12240 // 255 * 48
	// HKDFSHA512MaxLength is the HKDF-SHA512 maximum output length.
	HKDFSHA512MaxLength = 16320 // 255 * 64

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

	// PBKDF2Prefix is the hash format prefix for PBKDF2-HMAC-SHA256 hashes.
	PBKDF2Prefix = "pbkdf2-sha256"
	// PBKDF2DerivedKeyLength is the derived key length in bytes (32 = 256 bits).
	PBKDF2DerivedKeyLength = 32
	// PBKDF2DefaultHashName is the algorithm name for PBKDF2 SHA-256 (default).
	PBKDF2DefaultHashName = "pbkdf2-sha256"
	// PBKDF2SHA384HashName is the algorithm name for PBKDF2 SHA-384.
	PBKDF2SHA384HashName = "pbkdf2-sha384"
	// PBKDF2SHA512HashName is the algorithm name for PBKDF2 SHA-512.
	PBKDF2SHA512HashName = "pbkdf2-sha512"
	// PBKDF2DefaultAlgorithm is the default PRF algorithm for PBKDF2.
	PBKDF2DefaultAlgorithm = "SHA-256"
	// PBKDF2DefaultSaltBytes is the salt length in bytes (32 = 256 bits).
	PBKDF2DefaultSaltBytes = 32
	// PBKDF2DefaultHashBytes is the derived key length in bytes (32 = 256 bits for SHA-256).
	PBKDF2DefaultHashBytes = 32
	// PBKDF2SHA384HashBytes is the derived key length in bytes (48 = 384 bits for SHA-384).
	PBKDF2SHA384HashBytes = 48
	// PBKDF2SHA512HashBytes is the derived key length in bytes (64 = 512 bits for SHA-512).
	PBKDF2SHA512HashBytes = 64
	// PBKDF2MinIterations is the OWASP minimum iterations for PBKDF2-HMAC-SHA256 (2023).
	PBKDF2MinIterations = 210000

	// PBKDF2DefaultIterations is the PBKDF2 iteration count for Version 1 (2023).
	PBKDF2DefaultIterations = 600000
	// PBKDF2V2Iterations is the Version 2 (2021): NIST SP 800-63B Rev. 3 recommendation.
	PBKDF2V2Iterations = 310000
	// PBKDF2V3Iterations is the Version 3 (2017): Legacy/migration support (NIST 2017 minimum).
	PBKDF2V3Iterations = 1000

	// PBKDF2VersionedFormatParts is the number of parts in versioned hash format.
	PBKDF2VersionedFormatParts = 5

	// HKDFHashName is the HKDF-SHA256 hash name constant.
	HKDFHashName = "hkdf-sha256"
	// HKDFFixedLowHashName is the HKDF-SHA256 with fixed info (deterministic, low-entropy).
	HKDFFixedLowHashName = "hkdf-sha256-fixed"
	// HKDFFixedHighHashName is the HKDF-SHA256 with fixed info (deterministic, high-entropy).
	HKDFFixedHighHashName = "hkdf-sha256-fixed-high"

	// HKDFDelimiter is the delimiter for HKDF hash format parts.
	HKDFDelimiter = "$"
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

// Default pool configurations for key generation pools.
var (
	// DefaultPoolConfigRSA4096 is the pool configuration for RSA 4096-bit key generation.
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
