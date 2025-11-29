// Copyright (c) 2025 Justin Cranford

package magic

// Key usage types.
const (
	KeyUsageSigning    = "signing"    // Signing keys for JWT/JWS.
	KeyUsageEncryption = "encryption" // Encryption keys for JWE.
)

// JWK key types (RFC 7518 Section 6).
const (
	KeyTypeRSA = "RSA" // RSA key type.
	KeyTypeEC  = "EC"  // Elliptic Curve key type.
	KeyTypeOct = "oct" // Symmetric (octet sequence) key type.
)

// JWS signing algorithms (RFC 7518).
const (
	AlgorithmRS256 = "RS256" // RSA PKCS#1 v1.5 with SHA-256.
	AlgorithmRS384 = "RS384" // RSA PKCS#1 v1.5 with SHA-384.
	AlgorithmRS512 = "RS512" // RSA PKCS#1 v1.5 with SHA-512.
	AlgorithmES256 = "ES256" // ECDSA with P-256 and SHA-256.
	AlgorithmES384 = "ES384" // ECDSA with P-384 and SHA-384.
	AlgorithmES512 = "ES512" // ECDSA with P-521 and SHA-512.
	AlgorithmHS256 = "HS256" // HMAC with SHA-256.
	AlgorithmHS384 = "HS384" // HMAC with SHA-384.
	AlgorithmHS512 = "HS512" // HMAC with SHA-512.
)
