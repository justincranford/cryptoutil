// Package magic provides commonly used magic numbers and values as named constants.
// This file contains server-related constants.
package magic

import "time"

// Server configuration constants.
const (
	// ServerMaxRequestBodySize - Maximum request body size for test server (1MB).
	ServerMaxRequestBodySize = 1 << 20
	// ServerIdleTimeout - Idle timeout for test server connections (30 seconds).
	ServerIdleTimeout = 30 * time.Second
	// ServerReadHeaderTimeout - Header read timeout for test server (10 seconds).
	ServerReadHeaderTimeout = 10 * time.Second
	// ServerMaxHeaderBytes - Maximum header bytes for test server (1MB).
	ServerMaxHeaderBytes = 1 << 20
)
