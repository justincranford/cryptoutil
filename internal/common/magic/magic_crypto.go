package magic

import "time"

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
	JWKGenPoolRSA4096NumWorkers   = 9
	JWKGenPoolRSA4096PoolSize     = 9
	JWKGenPoolRSA3072NumWorkers   = 6
	JWKGenPoolRSA3072PoolSize     = 6
	JWKGenPoolRSA2048NumWorkers   = 3
	JWKGenPoolRSA2048PoolSize     = 3
	JWKGenPoolECDSAP521NumWorkers = 3
	JWKGenPoolECDSAP521PoolSize   = 9
	JWKGenPoolECDSAP384NumWorkers = 2
	JWKGenPoolECDSAP384PoolSize   = 6
	JWKGenPoolECDSAP256NumWorkers = 1
	JWKGenPoolECDSAP256PoolSize   = 3
	JWKGenPoolECDHP521NumWorkers  = 3
	JWKGenPoolECDHP521PoolSize    = 9
	JWKGenPoolECDHP384NumWorkers  = 2
	JWKGenPoolECDHP384PoolSize    = 6
	JWKGenPoolECDHP256NumWorkers  = 1
	JWKGenPoolECDHP256PoolSize    = 3
	JWKGenPoolED25519NumWorkers   = 1
	JWKGenPoolED25519PoolSize     = 2
	JWKGenPoolAES256NumWorkers    = 3
	JWKGenPoolAES256PoolSize      = 9
	JWKGenPoolAES192NumWorkers    = 2
	JWKGenPoolAES192PoolSize      = 6
	JWKGenPoolAES128NumWorkers    = 1
	JWKGenPoolAES128PoolSize      = 3
	JWKGenPoolHMAC512NumWorkers   = 3
	JWKGenPoolHMAC512PoolSize     = 9
	JWKGenPoolHMAC384NumWorkers   = 2
	JWKGenPoolHMAC384PoolSize     = 6
	JWKGenPoolHMAC256NumWorkers   = 1
	JWKGenPoolHMAC256PoolSize     = 3
	JWKGenPoolUUIDv7NumWorkers    = 2
	JWKGenPoolUUIDv7PoolSize      = 20

	// MaxPoolLifetimeValuesInt64 - Maximum int64 value (= 2^63-1 = 9,223,372,036,854,775,807).
	MaxPoolLifetimeValuesInt64 = int64(^uint64(0) >> 1)

	// MaxPoolLifetimeValues - Max int64 as uint64.
	MaxPoolLifetimeValues = uint64(MaxPoolLifetimeValuesInt64)
	// MaxPoolLifetimeDuration - Max int64 as nanoseconds (= 292.47 years).
	MaxPoolLifetimeDuration = time.Duration(MaxPoolLifetimeValuesInt64)

	// PoolMaintenanceInterval - Ticker interval for periodic pool maintenance checks.
	PoolMaintenanceInterval = 500 * time.Millisecond
)
