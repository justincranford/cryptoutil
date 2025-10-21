package magic

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

	// Test algorithm and provider constants.
	TestAlgorithmRSA = "RSA"
	TestProviderGO   = "GO"

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
)
