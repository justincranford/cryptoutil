// Copyright (c) 2025 Justin Cranford
//
//

/*
Package digests provides low-level cryptographic primitives for key derivation and hashing operations.

# Overview

This package contains pure cryptographic implementations without business logic:
  - PBKDF2-HMAC-SHA256: Password-based key derivation (for low-entropy secrets)
  - HKDF: HMAC-based Key Derivation Function (for high-entropy secrets)
  - SHA-2: Direct hashing utilities (SHA-224, SHA-256, SHA-384, SHA-512)

These primitives are consumed by the hash package for higher-level secret management APIs.

# PBKDF2 (Password-Based Key Derivation Function 2)

Used for deriving cryptographic keys from low-entropy secrets (passwords, PINs). Applies a pseudorandom function (HMAC-SHA256) with a salt and high iteration count to increase computational cost against brute-force attacks.

**Functions:**

  - PBKDF2WithParams(params PBKDF2Params, password []byte) ([]byte, error)
  - VerifySecret(storedHash string, providedPassword []byte, params PBKDF2Params) (bool, error)

**Format:**

	{version}$pbkdf2-sha256${iterations}${base64_salt}${base64_dk}

**Example:**

	params := PBKDF2Params{
	    Version:           "1",
	    HashName:          "pbkdf2-sha256",
	    Iterations:        600000,
	    SaltBytes:         32,
	    DerivedKeyLength:  32,
	}
	dk, err := PBKDF2WithParams(params, []byte("my_password"))
	// dk = 32-byte derived key

# HKDF (HMAC-based Key Derivation Function)

Used for deriving cryptographic keys from high-entropy secrets (API keys, tokens). More efficient than PBKDF2 due to lower computational overhead.

**Functions:**

  - HKDF(digestName string, secret, salt, info []byte, outputLength int) ([]byte, error)
  - HKDFwithSHA256(secret, salt, info []byte, outputLength int) ([]byte, error)
  - HKDFwithSHA384(secret, salt, info []byte, outputLength int) ([]byte, error)
  - HKDFwithSHA512(secret, salt, info []byte, outputLength int) ([]byte, error)
  - HKDFwithSHA224(secret, salt, info []byte, outputLength int) ([]byte, error)

**Supported Digest Algorithms:**

  - "sha512": SHA-512 (64-byte hash output)
  - "sha384": SHA-384 (48-byte hash output)
  - "sha256": SHA-256 (32-byte hash output, default)
  - "sha224": SHA-224 (28-byte hash output)

**Example:**

	secret := []byte("sk_live_1234567890abcdef")
	salt := make([]byte, 32) // Generate random salt
	info := []byte("api_key_derivation_v1")
	dk, err := HKDF("sha256", secret, salt, info, 32)
	// dk = 32-byte derived key

# SHA-2 Direct Hashing

Provides direct SHA-2 hashing without key derivation. Use for non-secret data hashing.

**Functions:**

  - SHA512(data []byte) []byte
  - SHA384(data []byte) []byte
  - SHA256(data []byte) []byte
  - SHA224(data []byte) []byte

**Example:**

	data := []byte("hash_me")
	hash := SHA256(data)
	// hash = 32-byte SHA-256 digest

# PBKDF2Params Structure

The PBKDF2Params struct encapsulates all parameters for PBKDF2 key derivation:

	type PBKDF2Params struct {
	    Version          string // Parameter set version (e.g., "1")
	    HashName         string // Algorithm name (e.g., "pbkdf2-sha256")
	    Iterations       int    // Iteration count (e.g., 600000)
	    SaltBytes        int    // Salt length in bytes (e.g., 32)
	    DerivedKeyLength int    // Output key length in bytes (e.g., 32)
	}

# Security Considerations

**PBKDF2:**
  - Use high iteration count (≥600,000 for 2023 OWASP recommendations)
  - Use random salt (32 bytes minimum)
  - Use FIPS-approved hash function (SHA-256, SHA-384, SHA-512)

**HKDF:**
  - Use random salt for non-deterministic key derivation
  - Use fixed info for deterministic key derivation (context binding)
  - Use FIPS-approved hash function (SHA-256, SHA-384, SHA-512)

**SHA-2:**
  - SHA-256 minimum for cryptographic applications
  - SHA-512 for higher security requirements
  - Never use SHA-1 or MD5 (deprecated, not FIPS-approved)

# FIPS 140-3 Compliance

All algorithms in this package are FIPS 140-3 approved:
  - PBKDF2-HMAC-SHA256: Approved key derivation function
  - HKDF: Approved key derivation function (SP 800-56C Rev 2)
  - SHA-2 family: Approved hash functions (FIPS 180-4)

# Package Dependencies

This package depends only on Go standard library and golang.org/x/crypto:
  - crypto/sha256, crypto/sha512: SHA-2 implementations
  - crypto/hmac: HMAC construction
  - crypto/subtle: Constant-time comparison
  - golang.org/x/crypto/hkdf: HKDF implementation
  - golang.org/x/crypto/pbkdf2: PBKDF2 implementation
*/
package digests
