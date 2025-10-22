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

	// WaitBeforeShutdownDuration - Client liveness start timeout.
	WaitBeforeShutdownDuration = 200 * time.Millisecond

	// ClientLivenessRequestTimeout - Client liveness request timeout.
	ClientLivenessRequestTimeout = 3 * time.Second
	// ClientReadinessRequestTimeout - Client readiness request timeout.
	ClientReadinessRequestTimeout = 5 * time.Second
	// ClientShutdownRequestTimeout - Client shutdown request timeout.
	ClientShutdownRequestTimeout = 5 * time.Second

	// DatabaseHealthCheckTimeout - Database health check timeout.
	DatabaseHealthCheckTimeout = 5 * time.Second
	// OtelCollectorHealthCheckTimeout - Otel collector health check timeout.
	OtelCollectorHealthCheckTimeout = 5 * time.Second
)
