// Copyright (c) 2025 Justin Cranford

/*
Package hash provides high-level cryptographic hashing APIs for secret management with semantic naming and parameter versioning.

# Architecture

This package serves as the business logic layer on top of the low-level cryptographic primitives in the digests package. It provides:
  - Semantic API naming (LowEntropy vs HighEntropy, Deterministic vs NonDeterministic)
  - Parameter versioning for future algorithm upgrades (V1, V2, V3)
  - Consistent hash format specifications across all providers
  - Registry-based parameter set management

# Provider Selection Guide

Choose the appropriate provider based on entropy level and determinism requirements:

**Low-Entropy Secrets** (passwords, PINs, passphrases):
  - HashLowEntropyNonDeterministic(): Uses PBKDF2-HMAC-SHA256 with random salt
  - HashLowEntropyDeterministic(): Uses HKDF-SHA256 with fixed info (for symmetric keys)

**High-Entropy Secrets** (API keys, tokens, cryptographic keys):
  - HashHighEntropyNonDeterministic(): Uses HKDF-SHA256 with random salt
  - HashHighEntropyDeterministic(): Uses HKDF-SHA256 with fixed info (for symmetric keys)

# Hash Format Specifications

**PBKDF2 Format** (Low-Entropy Non-Deterministic):

	{version}$pbkdf2-sha256${iterations}${base64_salt}${base64_dk}

Example:

	{1}$pbkdf2-sha256$600000$Y29uc3RhbnRfc2FsdC8xMjM0NTY3ODkwMTIzNDU2Nzg5MDEyMzQ1Njc4OTA=$dmVyaWZpZWRfa2V5LzEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5MA==

**HKDF Format** (High-Entropy Non-Deterministic):

	hkdf-sha256${base64_salt}${base64_dk}

Example:

	hkdf-sha256$Y29uc3RhbnRfc2FsdC8xMjM0NTY3ODkwMTIzNDU2Nzg5MDEyMzQ1Njc4OTA=$dmVyaWZpZWRfa2V5LzEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5MA==

**HKDF-Fixed-Low Format** (Low-Entropy Deterministic):

	hkdf-sha256-fixed${base64_dk}

Example:

	hkdf-sha256-fixed$dmVyaWZpZWRfa2V5LzEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5MA==

**HKDF-Fixed-High Format** (High-Entropy Deterministic):

	hkdf-sha256-fixed-high${base64_dk}

Example:

	hkdf-sha256-fixed-high$dmVyaWZpZWRfa2V5LzEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5MA==

# Parameter Versioning

PBKDF2 uses versioned parameter sets to support future algorithm upgrades:
  - V1: 600,000 iterations (default, OWASP 2023 recommended minimum)
  - V2: 1,000,000 iterations (enhanced security)
  - V3: 2,000,000 iterations (maximum security)

Access parameter sets via ParameterSetRegistry:

	registry := NewParameterSetRegistry()
	v1Params := registry.GetParameterSet("1")      // Returns V1 params
	v2Params := registry.GetParameterSet("2")      // Returns V2 params
	defaultParams := registry.GetDefaultParameterSet() // Returns V1 (default)

# Usage Examples

**Low-Entropy Password Hashing** (non-deterministic):

	// Hash a password with PBKDF2
	hash, err := HashLowEntropyNonDeterministic("my_password", DefaultPBKDF2ParameterSet())
	if err != nil {
	    log.Fatal(err)
	}
	// hash = "{1}$pbkdf2-sha256$600000$..."

	// Verify password against stored hash
	match, err := VerifyLowEntropyNonDeterministic("my_password", hash)
	if err != nil {
	    log.Fatal(err)
	}
	// match = true

**High-Entropy API Key Hashing** (non-deterministic):

	// Hash API key with HKDF
	hash, err := HashHighEntropyNonDeterministic("sk_live_1234567890abcdef")
	if err != nil {
	    log.Fatal(err)
	}
	// hash = "hkdf-sha256$randomsalt$..."

	// Verify API key against stored hash
	match, err := VerifyHighEntropyNonDeterministic("sk_live_1234567890abcdef", hash)
	if err != nil {
	    log.Fatal(err)
	}
	// match = true

**Deterministic Hashing** (for symmetric key derivation):

	// Hash secret deterministically with fixed info
	fixedInfo := []byte("app_context_v1")
	hash, err := HashLowEntropyDeterministic("my_password", fixedInfo)
	if err != nil {
	    log.Fatal(err)
	}
	// hash = "hkdf-sha256-fixed$..." (always same for same inputs)

# Security Considerations

  - Low-entropy secrets (passwords): Use PBKDF2 with high iteration count
  - High-entropy secrets (API keys): Use HKDF (lower overhead, still secure)
  - Deterministic hashing: Only use for symmetric key derivation, NOT for password storage
  - Random salt: ALWAYS use for password storage (non-deterministic hashing)
  - Fixed info: Use for deterministic key derivation (e.g., deriving encryption keys)

# Package Dependencies

This package depends on:
  - internal/shared/crypto/digests: Low-level PBKDF2, HKDF, SHA2 primitives
  - internal/shared/magic: Magic constants for hash names and delimiters
*/
package hash
