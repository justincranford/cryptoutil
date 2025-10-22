// Package magic provides commonly used magic numbers and values as named constants.
// This file contains miscellaneous constants.
package magic

import (
	"time"
)

// Miscellaneous constants.
const (
	// AnswerToLifeUniverseEverything - Answer to life, the universe, and everything.
	AnswerToLifeUniverseEverything = 42

	// CountDefaultPageSize - Default page size for pagination.
	CountDefaultPageSize = 25

	// EmptyString - Empty string constant.
	EmptyString = ""

	// StringUTCFormat - UTC time format string.
	StringUTCFormat = "2006-01-02T15:04:05Z"
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
	// StringUUIDRegexPattern - UUID regex pattern for validation.
	StringUUIDRegexPattern = `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`

	// FilePermissionsDefault - Default file permissions for created files.
	FilePermissionsDefault = 0o600
	// FilePermOwnerReadWriteOnly - Owner read/write only (0o600).
	FilePermOwnerReadWriteOnly = 0o600
	// FilePermOwnerReadWriteGroupRead - Owner read/write, group/other read (0o644).
	FilePermOwnerReadWriteGroupRead = 0o644
	// FilePermAllReadWrite - All users read/write (0o666).
	FilePermAllReadWrite = 0o666
	// FilePermOwnerReadWriteExecuteGroupReadExecute - Owner read/write/execute, group read/execute (0o750).
	FilePermOwnerReadWriteExecuteGroupReadExecute = 0o750
	// FilePermOwnerReadWriteExecuteGroupOtherReadExecute - Owner read/write/execute, group/other read/execute (0o755).
	FilePermOwnerReadWriteExecuteGroupOtherReadExecute = 0o755

	// FingerprintLeeway - Leeway for fingerprint matching in unseal operations.
	FingerprintLeeway = 1
	// StatusHealthy - Healthy status string.
	StatusHealthy = "HEALTHY"
	// StatusUnhealthy - Unhealthy status string.
	StatusUnhealthy = "UNHEALTHY"
	// MaxInt64 - Maximum int64 value (= 2^63-1 = 9,223,372,036,854,775,807).
	MaxInt64 = int64(^uint64(0) >> 1)

	// MaxLifetimeValues - Max int64 as uint64.
	MaxLifetimeValues = uint64(MaxInt64)
	// MaxLifetimeDuration - Max int64 as nanoseconds (= 292.47 years).
	MaxLifetimeDuration = time.Duration(MaxInt64)
	// BoolDefaultHelp - Default help flag value.
	BoolDefaultHelp = false
	// BoolDefaultDevMode - Default dev mode flag value.
	BoolDefaultDevMode = false
	// BoolDefaultDryRun - Default dry run flag value.
	BoolDefaultDryRun = false
	// BoolCSRFTokenCookieSecure - Default CSRF token cookie secure flag.
	BoolCSRFTokenCookieSecure = true
	// BoolCSRFTokenCookieHTTPOnly - Default CSRF token cookie HTTPOnly flag.
	BoolCSRFTokenCookieHTTPOnly = false
	// BoolCSRFTokenCookieSessionOnly - Default CSRF token cookie session only flag.
	BoolCSRFTokenCookieSessionOnly = true
	// BoolCSRFTokenSingleUseToken - Default CSRF token single use flag.
	BoolCSRFTokenSingleUseToken = false
)

const (
	// CountDevModeRandomBytesLength - Random bytes length for dev mode.
	CountDevModeRandomBytesLength = 32
)
