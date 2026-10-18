// Copyright (c) 2025-2026 Justin Cranford.
//
//

package magic

import "time"

const (
	// FiberAppIDRequestAttribute - Fiber app ID request attribute key.
	FiberAppIDRequestAttribute = "fiberAppID"

	// StringError - Error string constant.
	StringError        = "error"
	StringMessage      = "message"
	StringUnauthorized = "unauthorized"
	StringWaiting      = "waiting"
	// StringStatus - Status string constant.
	StringStatus = "status"
	// RequestCreationFailureMessage - Request creation failure message.
	RequestCreationFailureMessage = "Failed to create request"
	// StringStatusOK - OK status string.
	StringStatusOK = "ok"
	// StringStatusDegraded - Degraded status string.
	StringStatusDegraded = "degraded"
	// StringStatusShuttingDown - Shutting-down status string.
	StringStatusShuttingDown = "shutting down"

	// WaitBeforeShutdownDuration - Client liveness start timeout.
	WaitBeforeShutdownDuration = 200 * time.Millisecond

	// ClientLivenessRequestTimeout - Client liveness request timeout.
	ClientLivenessRequestTimeout = 3 * time.Second
	// ClientReadinessRequestTimeout - Client readiness request timeout.
	ClientReadinessRequestTimeout = 5 * time.Second
	// ClientShutdownRequestTimeout - Client shutdown request timeout.
	ClientShutdownRequestTimeout = 5 * time.Second

	// DefaultServiceHealthEndpoint is the service health endpoint.
	DefaultServiceHealthEndpoint = "/service/api/v1/health"
	// DefaultBrowserHealthEndpoint is the browser health endpoint.
	DefaultBrowserHealthEndpoint = "/browser/api/v1/health"
	// DefaultAdminLivezEndpoint is the admin liveness endpoint.
	DefaultAdminLivezEndpoint = DefaultPrivateAdminAPIContextPath + PrivateAdminLivezRequestPath

	// DatabaseHealthCheckTimeout - Database health check timeout.
	DatabaseHealthCheckTimeout = 5 * time.Second
	// OtelCollectorHealthCheckTimeout - Otel collector health check timeout.
	OtelCollectorHealthCheckTimeout = 5 * time.Second
)

// StringProviderInternal - Internal provider string.
var StringProviderInternal = "Internal"

// HTTP authorization constants.
const (
	// HTTPAuthorizationBearerPrefix - HTTP Authorization Bearer scheme prefix (with trailing space).
	HTTPAuthorizationBearerPrefix = "Bearer "

	// DefaultAPIListLimit - Default pagination limit for list API endpoints.
	DefaultAPIListLimit = 100

	// AuthorizationCheckTimeout - Timeout for authorization service checks from relying party.
	AuthorizationCheckTimeout = 5 * time.Second
)
