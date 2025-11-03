// Package magic provides commonly used magic numbers and values as named constants.
// This file contains test-specific constants.
package magic

import "time"

// TestCleartext - Standard test cleartext for encryption/decryption tests.
var TestCleartext = "Hello, World!"

// Test data constants.
const (
	// StatusHealthy - Healthy status string.
	StatusHealthy = "HEALTHY"
	// StatusUnhealthy - Unhealthy status string.
	StatusUnhealthy = "UNHEALTHY"
	// TestStatusPass - Test status indicating success.
	TestStatusPass = "PASS"
	// TestStatusFail - Test status indicating failure.
	TestStatusFail = "FAIL"
	// TestStatusSkip - Test status indicating skipped.
	TestStatusSkip = "SKIP"
	// TestStatusEmojiPass - Emoji for passed test status.
	TestStatusEmojiPass = "✅"
	// TestStatusEmojiFail - Emoji for failed test status.
	TestStatusEmojiFail = "❌"
	// TestStatusEmojiSkip - Emoji for skipped test status.
	TestStatusEmojiSkip = "⏭️"
)

const (
	// TestTimeoutDockerComposeInit - Timeout for Docker Compose services to initialize.
	TestTimeoutDockerHealth = 10 * time.Second //nolint:stylecheck // established API name
	// TestTimeoutDockerHealth - Timeout for Docker health checks.
	TestTimeoutCryptoutilReady = 10 * time.Second //nolint:stylecheck // Cryptoutil needs time to unseal - reduced for fast fail
	// TestTimeoutCryptoutilReady - Timeout for Cryptoutil readiness checks.
	TestTimeoutTestExecution = 60 * time.Second //nolint:stylecheck // Overall test timeout - reduced for fast fail
	// TestTimeoutDockerComposeInit - Timeout for Docker Compose services to initialize.
	TestTimeoutDockerComposeInit = 5 * time.Second //nolint:stylecheck // Time to wait for Docker Compose services to initialize after startup
	// TestTimeoutServiceRetry - Timeout for service retry intervals.
	TestTimeoutServiceRetry = 500 * time.Millisecond //nolint:stylecheck // Check more frequently

	// TimeoutTestServerReady - Test server ready timeout.
	TimeoutTestServerReady = 30 * time.Second
	// TimeoutTestServerReadyRetryDelay - Test server ready retry delay.
	TimeoutTestServerReadyRetryDelay = 500 * time.Millisecond

	// TestDefaultServerShutdownTimeout - Default server shutdown timeout duration.
	TestDefaultServerShutdownTimeout = 1 * time.Minute //nolint:stylecheck // established API name

	// TestTimeoutHTTPRetryInterval - Default HTTP retry interval duration.
	TestTimeoutHTTPRetryInterval = 1 * time.Second //nolint:stylecheck // established API name
	// TestTimeoutHTTPClient - Default HTTP client timeout duration.
	TestTimeoutHTTPClient = 10 * time.Second //nolint:stylecheck // established API name

	// TimeoutHTTPHealthRequest - HTTP health request timeout.
	TimeoutHTTPHealthRequest = 5 * time.Second

	// TimeoutGitHubAPIDelay - Delay between GitHub API calls to avoid rate limits.
	TimeoutGitHubAPIDelay = 200 * time.Millisecond
	// TimeoutGitHubAPITimeout - Timeout for GitHub API requests.
	TimeoutGitHubAPITimeout = 10 * time.Second
	// TimeoutGitHubAPICacheTTL - TTL for GitHub API response cache (1 hour).
	TimeoutGitHubAPICacheTTL = 1 * time.Hour

	// TestSleepCancelChanContext - 5 milliseconds duration for test delays.
	TestSleepCancelChanContext = 5 * time.Millisecond //nolint:stylecheck // established API name

	// TestTLSClientRetryWait - 100 milliseconds duration for brief backoff operations.
	TestTLSClientRetryWait = 100 * time.Millisecond //nolint:stylecheck // established API name
)

// Test server timeout constants.
const (
	// TestTLSServerStartupDelay - TLS server startup delay for tests.
	TestTLSServerStartupDelay = 500 * time.Millisecond
	// TestTLSServerWriteTimeout - TLS server write timeout for tests.
	TestTLSServerWriteTimeout = 500 * time.Millisecond
	// TestTLSServerReadTimeout - TLS server read timeout for tests.
	TestTLSServerReadTimeout = 500 * time.Millisecond
	// TestTLSRetryBaseDelay - Base delay for TLS retry operations in tests.
	TestTLSRetryBaseDelay = 10 * time.Millisecond
	// TestTLSMaxRetries - Maximum retry attempts for TLS operations in tests.
	TestTLSMaxRetries = 3

	// TestHTTPServerStartupDelay - HTTP server startup delay for tests.
	TestHTTPServerStartupDelay = 500 * time.Millisecond
	// TestHTTPServerWriteTimeout - HTTP server write timeout for tests.
	TestHTTPServerWriteTimeout = 500 * time.Millisecond
	// TestHTTPServerReadTimeout - HTTP server read timeout for tests.
	TestHTTPServerReadTimeout = 500 * time.Millisecond
	// TestHTTPRetryBaseDelay - Base delay for HTTP retry operations in tests.
	TestHTTPRetryBaseDelay = 10 * time.Millisecond
	// TestHTTPMaxRetries - Maximum retry attempts for HTTP operations in tests.
	TestHTTPMaxRetries = 3
)

// Test duration constants.
const (
	// TestNegativeDuration - Negative duration for testing invalid inputs.
	TestNegativeDuration = -1
	// TestHourDuration - One hour duration for testing.
	TestHourDuration = time.Hour
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

	// TestReportWidth - Standard width for test summary reports.
	TestReportWidth = 80
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

// Test cryptographic algorithm constants for E2E tests.
const (
	// TestJwkJwsAlgorithm - JWS algorithm for elastic key signing in tests.
	TestJwkJwsAlgorithm = "RS256"
)

var (
	// TestJwkJweAlgorithm - JWE algorithm for elastic key encryption in tests.
	TestJwkJweAlgorithm = "A256GCM/A256KW"
	// TestElasticKeyImportAllowed - Default import allowed setting for test elastic keys.
	TestElasticKeyImportAllowed = false
	// TestElasticKeyVersioningAllowed - Default versioning allowed setting for test elastic keys.
	TestElasticKeyVersioningAllowed = true
)
