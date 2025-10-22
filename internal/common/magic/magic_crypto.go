package magic

// Cryptographic algorithm and key constants.
// This file contains all crypto-related magic values used throughout the application.

const (
	// CountExpectedSysInfos - Expected number of system info items.
	CountExpectedSysInfos = 13
	// CountMaxSharedSecrets - Maximum number of shared secrets allowed.
	CountMaxSharedSecrets = 256
	// CountMinSharedSecretLength - Minimum shared secret length in bytes.
	CountMinSharedSecretLength = 32
	// CountMaxSharedSecretLength - Maximum shared secret length in bytes.
	CountMaxSharedSecretLength = 64
	// CountDerivedKeySizeBytes - Derived key size in bytes.
	CountDerivedKeySizeBytes = 32

	// EdDSA curve names.
	EdCurveEd448   = "Ed448"
	EdCurveEd25519 = "Ed25519"

	// Elliptic curve names.
	ECCurveP256 = "P256"
	ECCurveP384 = "P384"
	ECCurveP521 = "P521"

	// RSA key sizes in bits.
	RSAKeySize2048 = 2048
	RSAKeySize3072 = 3072
	RSAKeySize4096 = 4096

	// AES key sizes in bits.
	AESKeySize128 = 128
	AESKeySize192 = 192
	AESKeySize256 = 256

	// AES HMAC-SHA2 key sizes in bits.
	AESHSKeySize256 = 256
	AESHSKeySize384 = 384
	AESHSKeySize512 = 512

	// HMAC key sizes in bits.
	HMACKeySize256 = 256
	HMACKeySize384 = 384
	HMACKeySize512 = 512

	// Minimum HMAC key size in bits.
	MinHMACKeySize = 256

	// SHA digest algorithm names.
	SHADigestSHA512 = "SHA512"
	SHADigestSHA384 = "SHA384"
	SHADigestSHA256 = "SHA256"
	SHADigestSHA224 = "SHA224"

	// Bits to bytes conversion factor.
	BitsToBytes = 8

	// Serial number bit sizes for cryptographic range.
	MinSerialNumberBits = 64
	MaxSerialNumberBits = 159

	// HKDF test constants.
	HKDFSHA256OutputLength = 32
	HKDFSHA384OutputLength = 48
	HKDFSHA512OutputLength = 64
	HKDFSHA224OutputLength = 28
	HKDFMaxMultiplier      = 255
	HKDFSHA256MaxLength    = 8160  // 255 * 32
	HKDFSHA384MaxLength    = 12240 // 255 * 48
	HKDFSHA512MaxLength    = 16320 // 255 * 64
	HKDFSHA224MaxLength    = 7140  // 255 * 28

	// JWE key sizes in bits.
	JWEA256KeySize   = 256
	JWEA192KeySize   = 192
	JWEA128KeySize   = 128
	JWEA512KeySize   = 512
	JWEA384KeySize   = 384
	JWEKEA256KeySize = 256
	JWEKEA192KeySize = 192
	JWEKEA128KeySize = 128

	// JWK generation pool sizes (min, max) by algorithm type.
	JWKGenPoolRSA4096Min   = 9
	JWKGenPoolRSA4096Max   = 9
	JWKGenPoolRSA3072Min   = 6
	JWKGenPoolRSA3072Max   = 6
	JWKGenPoolRSA2048Min   = 3
	JWKGenPoolRSA2048Max   = 3
	JWKGenPoolECDSAP521Min = 3
	JWKGenPoolECDSAP521Max = 9
	JWKGenPoolECDSAP384Min = 2
	JWKGenPoolECDSAP384Max = 6
	JWKGenPoolECDSAP256Min = 1
	JWKGenPoolECDSAP256Max = 3
	JWKGenPoolECDHP521Min  = 3
	JWKGenPoolECDHP521Max  = 9
	JWKGenPoolECDHP384Min  = 2
	JWKGenPoolECDHP384Max  = 6
	JWKGenPoolECDHP256Min  = 1
	JWKGenPoolECDHP256Max  = 3
	JWKGenPoolED25519Min   = 1
	JWKGenPoolED25519Max   = 2
	JWKGenPoolAES256Min    = 3
	JWKGenPoolAES256Max    = 9
	JWKGenPoolAES192Min    = 2
	JWKGenPoolAES192Max    = 6
	JWKGenPoolAES128Min    = 1
	JWKGenPoolAES128Max    = 3
	JWKGenPoolHMAC512Min   = 3
	JWKGenPoolHMAC512Max   = 9
	JWKGenPoolHMAC384Min   = 2
	JWKGenPoolHMAC384Max   = 6
	JWKGenPoolHMAC256Min   = 1
	JWKGenPoolHMAC256Max   = 3
	JWKGenPoolUUIDv7Min    = 2
	JWKGenPoolUUIDv7Max    = 20
)
