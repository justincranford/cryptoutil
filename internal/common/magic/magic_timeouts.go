// Package magic provides commonly used magic numbers and values as named constants.
// This file contains timeout and duration constants.
package magic

import "time"

// Timeouts and durations.
const (
	// Timeout1Second - 1 second duration, common timeout unit.
	Timeout1Second = 1 * time.Second //nolint:stylecheck // established API name
	// Timeout1Minute - 1 minute duration, server shutdown timeout.
	Timeout1Minute = 1 * time.Minute //nolint:stylecheck // established API name
	// Timeout10Seconds - 10 seconds duration, rate limit maximum.
	Timeout10Seconds = 10 * time.Second //nolint:stylecheck // established API name
	// Timeout5Seconds - 5 seconds duration for memory and host ID operations.
	Timeout5Seconds = 5 * time.Second //nolint:stylecheck // established API name
	// Timeout100Milliseconds - 100 milliseconds duration for brief backoff operations.
	Timeout100Milliseconds = 100 * time.Millisecond //nolint:stylecheck // established API name
	// Timeout5Minutes - 5 minutes duration for Docker Compose operations.
	Timeout5Minutes = 5 * time.Minute //nolint:stylecheck // established API name
	// Timeout30Seconds - 30 seconds duration for various operations.
	Timeout30Seconds = 30 * time.Second //nolint:stylecheck // established API name
	// Timeout15Seconds - 15 seconds duration for Docker Compose initialization.
	Timeout15Seconds = 15 * time.Second //nolint:stylecheck // established API name
	// Timeout2Seconds - 2 seconds duration for service retry intervals.
	Timeout2Seconds = 2 * time.Second //nolint:stylecheck // established API name
	// Timeout5Milliseconds - 5 milliseconds duration for test delays.
	Timeout5Milliseconds = 5 * time.Millisecond //nolint:stylecheck // established API name
	// TimeoutGitHubAPIDelay - Delay between GitHub API calls to avoid rate limits.
	TimeoutGitHubAPIDelay = 200 * time.Millisecond
	// TimeoutGitHubAPITimeout - Timeout for GitHub API requests.
	TimeoutGitHubAPITimeout = 10 * time.Second
	// TimeoutCSRFTokenMaxAge - CSRF token maximum age (1 hour).
	TimeoutCSRFTokenMaxAge = 1 * time.Hour
	// TimeoutLogs - Timeout for logs operations.
	TimeoutLogs = 500 * time.Millisecond
	// TimeoutMetrics - Timeout for metrics operations.
	TimeoutMetrics = 2000 * time.Millisecond
	// TimeoutTraces - Timeout for traces operations.
	TimeoutTraces = 1000 * time.Millisecond
	// TimeoutForceFlush - Timeout for force flush on shutdown.
	TimeoutForceFlush = 3 * time.Second
	// TimeoutClientShutdownRequest - Client shutdown request timeout.
	TimeoutClientShutdownRequest = 5 * time.Second
	// TimeoutClientLivenessRequest - Client liveness request timeout.
	TimeoutClientLivenessRequest = 3 * time.Second
	// TimeoutClientReadinessRequest - Client readiness request timeout.
	TimeoutClientReadinessRequest = 5 * time.Second
	// TimeoutClientLivenessStart - Client liveness start timeout.
	TimeoutClientLivenessStart = 200 * time.Millisecond
	// TimeoutHealthCheck - Health check timeout.
	TimeoutHealthCheck = 5 * time.Second
	// TimeoutTestServerReady - Test server ready timeout.
	TimeoutTestServerReady = 30 * time.Second
	// TimeoutTestServerReadyRetryDelay - Test server ready retry delay.
	TimeoutTestServerReadyRetryDelay = 500 * time.Millisecond
	// TimeoutPoolMaintenanceInterval - Ticker interval for periodic pool maintenance checks.
	TimeoutPoolMaintenanceInterval = 500 * time.Millisecond
	// TimeoutDays1 - 1 day duration.
	TimeoutDays1 = 24 * time.Hour
	// TimeoutDays30 - 30 days duration.
	TimeoutDays30 = 30 * TimeoutDays1
	// TimeoutDays365 - 365 days duration.
	TimeoutDays365 = 365 * TimeoutDays1
	// TimeoutHTTPHealthRequest - HTTP health request timeout.
	TimeoutHTTPHealthRequest = 5 * time.Second
)
