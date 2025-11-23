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
)

// Base64 encoding constants.
const (
	// Base64ExpansionNumerator is the numerator for base64 expansion ratio (3 bytes → 4 chars).
	Base64ExpansionNumerator = 3
	// Base64ExpansionDenominator is the denominator for base64 expansion ratio (3 bytes → 4 chars).
	Base64ExpansionDenominator = 4
)
