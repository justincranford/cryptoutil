// Copyright (c) 2025-2026 Justin Cranford.
//
//

package magic

import "time"

// Test server timeout constants.
const (
	// TestReadTimeout is the default read timeout for test HTTP servers.
	TestReadTimeout = 30 * time.Second
	// TestWriteTimeout is the default write timeout for test HTTP servers.
	TestWriteTimeout = 30 * time.Second
	// TestIdleTimeout is the default idle timeout for test HTTP servers.
	TestIdleTimeout = 120 * time.Second
	// TestRefreshTokenLifetime is the default refresh token lifetime for tests.
	TestRefreshTokenLifetime = 24 * time.Hour
	// TestServerWaitTickerInterval is the interval for checking server readiness.
	TestServerWaitTickerInterval = 100 * time.Millisecond
	// DatabasePropagationDelay is the time to wait for database updates to propagate.
	DatabasePropagationDelay = 50 * time.Millisecond
)

// Fiber test timeout constants.
const (
	// FiberTestTimeoutMs is the timeout in milliseconds for Fiber app.Test() calls in parallel tests.
	// Increased from default 1000ms (1s) to handle concurrent test execution under load.
	FiberTestTimeoutMs = 30000 // 30 seconds in milliseconds
)

// Base64 encoding constants.
const (
	// Base64ExpansionNumerator is the numerator for base64 expansion ratio (3 bytes → 4 chars).
	Base64ExpansionNumerator = 3
	// Base64ExpansionDenominator is the denominator for base64 expansion ratio (3 bytes → 4 chars).
	Base64ExpansionDenominator = 4
)

// Client secret constants.
const (
	// ClientSecretLength is the length of generated client secrets (32 bytes = 256 bits).
	ClientSecretLength = 32
	// HashPrefixLength is the number of characters to log from password hashes for audit trails.
	HashPrefixLength = 12
)

// PKCE testing constants.
const (
	// S256_METHOD is a subtest label for S256 PKCE flows.
	S256_METHOD = "s256"
	// PLAIN_METHOD is a subtest label for plain PKCE flows.
	PLAIN_METHOD = "plain"
	// DEFAULT_METHOD_S256 is a subtest label for default-method PKCE flows.
	DEFAULT_METHOD_S256 = "default_s256"
	// INVALID_METHOD is a synthetic unsupported method for negative-path tests.
	INVALID_METHOD = "invalid_method"

	// TEST_CODE_VERIFIER is a deterministic verifier used by PKCE tests.
	TEST_CODE_VERIFIER = "test-code-verifier"
	// S256_CODE_CHALLENGE is a deterministic verifier seed used for S256 challenge generation.
	S256_CODE_CHALLENGE = "s256-code-challenge"
	// TEST_CODE_CHALLENGE is a deterministic challenge value used by validation tests.
	TEST_CODE_CHALLENGE = "test-code-challenge"
	// WRONG_VERIFIER is a mismatched verifier for negative validation tests.
	WRONG_VERIFIER = "wrong-verifier"
	// SHORT_VERIFIER is a subtest label for short-verifier coverage.
	SHORT_VERIFIER = "short_verifier"
	// LONG_VERIFIER is a subtest label for long-verifier coverage.
	LONG_VERIFIER = "long_verifier"
	// VERY_LONG_VERIFIER is a long verifier input for boundary checks.
	VERY_LONG_VERIFIER = "this-is-a-very-long-pkce-verifier-used-for-boundary-validation-cases"
	// EMPTY_VERIFIER is a subtest label for empty-verifier coverage.
	EMPTY_VERIFIER = "empty_verifier"
	// EMPTY_VERIFIER_WITH_NON_EMPTY_CHALLENGE is a negative-case subtest label.
	EMPTY_VERIFIER_WITH_NON_EMPTY_CHALLENGE = "empty_verifier_with_non_empty_challenge"
	// NON_EMPTY_VERIFIER_WITH_EMPTY_CHALLENGE is a negative-case subtest label.
	NON_EMPTY_VERIFIER_WITH_EMPTY_CHALLENGE = "non_empty_verifier_with_empty_challenge"
	// NON_EMPTY is a generic non-empty string for validation edge cases.
	NON_EMPTY = "non-empty"
)
