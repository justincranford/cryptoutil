// Package magic provides commonly used magic numbers and values as named constants.
// This file contains string constants.
package magic

// String constants.
const (
	// StringUTCFormat - UTC time format string.
	StringUTCFormat = "2006-01-02T15:04:05Z"
	// StringLivezPath - Livez endpoint path.
	StringLivezPath = "/livez"
	// StringReadyzPath - Readyz endpoint path.
	StringReadyzPath = "/readyz"
	// StringShutdownPath - Shutdown endpoint path.
	StringShutdownPath = "/shutdown"
	// StringError - Error string constant.
	StringError = "error"
	// StringStatus - Status string constant.
	StringStatus = "status"
	// StringProtocolHTTPS - HTTPS protocol string.
	StringProtocolHTTPS = "https"
	// StringStatusOK - OK status string.
	StringStatusOK = "ok"
	// StringStatusDegraded - Degraded status string.
	StringStatusDegraded = "degraded"
	// StringProviderInternal - Internal provider string.
	StringProviderInternal = "Internal"
	// StringUUIDRegexPattern - UUID regex pattern for validation.
	StringUUIDRegexPattern = `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`
)
