// Copyright (c) 2025 Justin Cranford
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
