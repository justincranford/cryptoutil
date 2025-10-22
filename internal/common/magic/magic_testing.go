// Package magic provides commonly used magic numbers and values as named constants.
// This file contains test-specific constants.
package magic

import "time"

// Test data constants.
const (
	// TestCleartext - Standard test cleartext for encryption/decryption tests.
	TestCleartext = "Hello, World!"

	// StatusHealthy - Healthy status string.
	StatusHealthy = "HEALTHY"
	// StatusUnhealthy - Unhealthy status string.
	StatusUnhealthy = "UNHEALTHY"
)

// Test settings constants.
const (
	// TestDefaultLogLevelAll - All log level for comprehensive logging.
	TestDefaultLogLevelAll = "ALL"

	// TestDefaultDevMode - Default dev mode flag for tests.
	TestDefaultDevMode = true

	// TestDefaultRateLimitBrowserIP - Default browser IP rate limit.
	TestDefaultRateLimitBrowserIP uint16 = 1000
	// TestDefaultRateLimitServiceIP - Default service IP rate limit.
	TestDefaultRateLimitServiceIP uint16 = 500

	// TestDefaultServerIdleTimeout - Idle timeout for test server connections (30 seconds).
	TestDefaultServerIdleTimeout = 30 * time.Second
	// TestDefaultServerReadHeaderTimeout - Header read timeout for test server (10 seconds).
	TestDefaultServerReadHeaderTimeout = 10 * time.Second
	// TestDefaultServerMaxHeaderBytes - Maximum header bytes for test server (1MB).
	TestDefaultServerMaxHeaderBytes = 1 << 20
)

// Test data for mock system information.
const (
	// MockRuntimeGoArch - Mock architecture for testing.
	MockRuntimeGoArch = "amd64"
	// MockRuntimeGoOS - Mock operating system for testing.
	MockRuntimeGoOS = "linux"
	// MockCPUVendorID - Mock CPU vendor ID for testing.
	MockCPUVendorID = "GenuineIntel"
	// MockCPUFamily - Mock CPU family identifier for testing.
	MockCPUFamily = "6"
	// MockCPUModel - Mock CPU model identifier for testing.
	MockCPUModel = "0"
	// MockCPUModelName - Mock CPU model name for testing.
	MockCPUModelName = "Intel(R) Core(TM) i7-8550U"
	// MockHostname - Mock hostname for testing.
	MockHostname = "mock-hostname"
	// MockHostID - Mock host ID for testing.
	MockHostID = "mock-host-id"
	// MockUserID - Mock user ID for testing.
	MockUserID = "mock-user-id-1000"
	// MockGroupID - Mock group ID for testing.
	MockGroupID = "mock-group-id-1000"
	// MockUsername - Mock username for testing.
	MockUsername = "mock-username"
	// MockCPUCount - Mock number of CPUs for testing.
	MockCPUCount = 4
	// MockRAMMB - Mock RAM size in MB for testing (8GB).
	MockRAMMB = 8192
)
