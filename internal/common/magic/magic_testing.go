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

const (
	// TestTimeoutDockerComposeInit - Timeout for Docker Compose services to initialize.
	TestTimeoutDockerHealth = 30 * time.Second //nolint:stylecheck // established API name
	// TestTimeoutDockerHealth - Timeout for Docker health checks.
	TestTimeoutCryptoutilReady = 30 * time.Second //nolint:stylecheck // Cryptoutil needs time to unseal - reduced for fast fail
	// TestTimeoutCryptoutilReady - Timeout for Cryptoutil readiness checks.
	TestTimeoutTestExecution = 30 * time.Second //nolint:stylecheck // Overall test timeout - reduced for fast fail
	// TestTimeoutDockerComposeInit - Timeout for Docker Compose services to initialize.
	TestTimeoutDockerComposeInit = 15 * time.Second //nolint:stylecheck // Time to wait for Docker Compose services to initialize after startup
	// TestTimeoutServiceRetry - Timeout for service retry intervals.
	TestTimeoutServiceRetry = 2 * time.Second //nolint:stylecheck // Check more frequently

	// TimeoutTestServerReady - Test server ready timeout.
	TimeoutTestServerReady = 30 * time.Second
	// TimeoutTestServerReadyRetryDelay - Test server ready retry delay.
	TimeoutTestServerReadyRetryDelay = 500 * time.Millisecond

	// TestDefaultServerShutdownTimeout - Default server shutdown timeout duration.
	TestDefaultServerShutdownTimeout = 1 * time.Minute //nolint:stylecheck // established API name

	// TestDefaultHTTPRetryInterval - Default HTTP retry interval duration.
	TestDefaultHTTPRetryInterval = 1 * time.Second //nolint:stylecheck // established API name
	// TestDefaultHTTPClientTimeout - Default HTTP client timeout duration.
	TestDefaultHTTPClientTimeout = 10 * time.Second //nolint:stylecheck // established API name

	// TimeoutHTTPHealthRequest - HTTP health request timeout.
	TimeoutHTTPHealthRequest = 5 * time.Second

	// TimeoutGitHubAPIDelay - Delay between GitHub API calls to avoid rate limits.
	TimeoutGitHubAPIDelay = 200 * time.Millisecond
	// TimeoutGitHubAPITimeout - Timeout for GitHub API requests.
	TimeoutGitHubAPITimeout = 10 * time.Second

	// TestSleepCancelChanContext - 5 milliseconds duration for test delays.
	TestSleepCancelChanContext = 5 * time.Millisecond //nolint:stylecheck // established API name

	// TestTLSClientRetryWait - 100 milliseconds duration for brief backoff operations.
	TestTLSClientRetryWait = 100 * time.Millisecond //nolint:stylecheck // established API name
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
