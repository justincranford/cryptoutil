package magic

import "time"

const (
	// FiberAppIDRequestAttribute - Fiber app ID request attribute key.
	FiberAppIDRequestAttribute = "fiberAppID"

	// StringError - Error string constant.
	StringError = "error"
	// StringStatus - Status string constant.
	StringStatus = "status"
	// StringStatusOK - OK status string.
	StringStatusOK = "ok"
	// StringStatusDegraded - Degraded status string.
	StringStatusDegraded = "degraded"

	// StringProviderInternal - Internal provider string.
	StringProviderInternal = "Internal"

	// TimeoutClientLivenessStart - Client liveness start timeout.
	TimeoutClientLivenessStart = 200 * time.Millisecond

	// TimeoutClientLivenessRequest - Client liveness request timeout.
	TimeoutClientLivenessRequest = 3 * time.Second
	// TimeoutClientReadinessRequest - Client readiness request timeout.
	TimeoutClientReadinessRequest = 5 * time.Second
	// TimeoutClientShutdownRequest - Client shutdown request timeout.
	TimeoutClientShutdownRequest = 5 * time.Second

	// TimeoutHealthCheck - Health check timeout.
	TimeoutHealthCheck = 5 * time.Second
)
