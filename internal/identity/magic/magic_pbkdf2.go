// Copyright (c) 2025 Justin Cranford
//
//

package magic

const (
	// PBKDF2Iterations is the iteration count for PBKDF2-HMAC-SHA256 (FIPS 140-3 approved).
	// NIST SP 800-132 recommends minimum 10,000 iterations for password-based key derivation.
	// Using 100,000 iterations for enhanced security against brute-force attacks.
	PBKDF2Iterations = 100000

	// PBKDF2SaltLength is the salt length in bytes for PBKDF2.
	// NIST SP 800-132 recommends minimum 128 bits (16 bytes) of salt.
	PBKDF2SaltLength = 16

	// PBKDF2KeyLength is the derived key length in bytes.
	// Using 32 bytes (256 bits) to match SHA-256 output size.
	PBKDF2KeyLength = 32
)
