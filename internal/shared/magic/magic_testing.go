// Copyright (c) 2025 Justin Cranford
//
//

package magic

import "time"

// TestCleartext - Standard test cleartext for encryption/decryption tests.
var TestCleartext = "Hello, World!"

// Demo CLI constants (Session 3 Q11-15, Session 5 Q1, Q15).
const (
	// DefaultDemoTimeout - Overall timeout for demo execution.
	DefaultDemoTimeout = 5 * time.Minute
	// DefaultDemoRetryCount - Number of retries for failed demo operations.
	DefaultDemoRetryCount = 3
	// DefaultDemoRetryDelay - Delay between retry attempts.
	DefaultDemoRetryDelay = 1 * time.Second
	// DefaultServerStartupDelay - Delay to allow server to start.
	DefaultServerStartupDelay = 500 * time.Millisecond
	// DefaultHealthCheckInterval - Interval between health check polls.
	DefaultHealthCheckInterval = 1 * time.Second
	// DefaultSpinnerInterval - Interval for spinner animation frames.
	DefaultSpinnerInterval = 100 * time.Millisecond
)

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

// Test execution probability constants - Control the probability of running specific test cases.
// Used for table-driven tests with multiple algorithm variants to reduce test execution time.
// Example: TestProbAlways = 1.0 (100%), TestProbHalf = 0.5 (50%), TestProbThird = 0.333 (33.3%).
const (
	// TestProbAlways - Execute test always (100% probability).
	TestProbAlways = 1.0
	// TestProbHalf - Execute test 50% of the time.
	TestProbHalf = 0.5
	// TestProbThird - Execute test 33.3% of the time.
	TestProbThird = 0.333
	// TestProbQuarter - Execute test 25% of the time.
	TestProbQuarter = 0.25
	// TestProbTenth - Execute test 10% of the time.
	TestProbTenth = 0.1
	// TestProbNever - Never execute test (0% probability), effectively skip.
	TestProbNever = 0.0
)

const (
	// TestTimeoutDockerHealth - Timeout for Docker health checks. Allows for:
	// - postgres: start_period=5s + (interval=5s * retries=5) = up to 30s
	// - cryptoutil: start_period=10s + (interval=5s * retries=5) = up to 35s
	// - dependency chain: healthcheck-secrets → builder → postgres → otel healthcheck sidecar → cryptoutil instances
	// - otel healthcheck sidecar: 10s sleep + (15 retries * 2s) = up to 40s
	// - GitHub Actions runner overhead: shared CPU, network latency, cold starts
	// Total: 5 minutes allows margin for GitHub Actions environment (was 3 min, insufficient).
	TestTimeoutDockerHealth = 300 * time.Second //nolint:stylecheck // established API name
	// TestTimeoutCryptoutilReady - Timeout for Cryptoutil readiness checks. Cryptoutil needs time to unseal.
	TestTimeoutCryptoutilReady = 30 * time.Second //nolint:stylecheck // Cryptoutil needs time to unseal
	// TestTimeoutTestExecution - Overall test execution timeout.
	TestTimeoutTestExecution = 5 * time.Minute //nolint:stylecheck // Overall test timeout
	// TestTimeoutDockerComposeInit - Timeout for Docker Compose services to initialize after startup.
	TestTimeoutDockerComposeInit = 10 * time.Second //nolint:stylecheck // Time to wait for Docker Compose services to initialize after startup
	// TestTimeoutServiceRetry - Timeout for service retry intervals. Check frequently.
	TestTimeoutServiceRetry = 2 * time.Second //nolint:stylecheck // Check every 2 seconds

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

	// GitHubRateLimitRemainingThreshold - Threshold for remaining API calls before adding extra delay.
	GitHubRateLimitRemainingThreshold = 10
	// GitHubRateLimitExtraDelay - Extra delay when remaining API calls are below threshold.
	GitHubRateLimitExtraDelay = 2 * time.Second

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

	// TestRSAKeySize - RSA key size for test certificate generation.
	TestRSAKeySize = 2048
	// TestRandomStringLength32 - Standard random string length for tokens.
	TestRandomStringLength32 = 32
	// TestRandomStringLength16 - Short random string length for session IDs.
	TestRandomStringLength16 = 16
	// TestRandomStringLength64 - Long random string length for ID tokens.
	TestRandomStringLength64 = 64
	// TestTokenExpirationSeconds - Standard token expiration time in seconds.
	TestTokenExpirationSeconds = 3600
	// TestServiceStartupDelaySeconds - Delay to allow services to start.
	TestServiceStartupDelaySeconds = 3
	// TestHTTPHealthTimeoutSeconds - HTTP health check timeout.
	TestHTTPHealthTimeoutSeconds = 5
	// TestAuthZServerPort - Mock authorization server port for standalone E2E tests.
	// NOTE: Used by internal/apps/identity/test/e2e/mock_services.go for in-process mock servers.
	// These are NOT deployment ports. Ideally should use port 0 for dynamic allocation.
	TestAuthZServerPort = 8080
	// TestIDPServerPort - Mock identity provider server port for standalone E2E tests.
	TestIDPServerPort = 8081
	// TestResourceServerPort - Mock resource server port for standalone E2E tests.
	TestResourceServerPort = 8082
	// TestSPARPServerPort - Mock SPA relying party server port for standalone E2E tests.
	TestSPARPServerPort = 8083
)

var (
	// TestJwkJweAlgorithm - JWE algorithm for elastic key encryption in tests.
	TestJwkJweAlgorithm = "A256GCM/A256KW"
	// TestElasticKeyImportAllowed - Default import allowed setting for test elastic keys.
	TestElasticKeyImportAllowed = false
	// TestElasticKeyVersioningAllowed - Default versioning allowed setting for test elastic keys.
	TestElasticKeyVersioningAllowed = true
)
