// Package magic provides commonly used magic numbers and values as named constants.
// This file contains miscellaneous constants.
package magic

import (
	"math"
	"time"
)

// Miscellaneous constants.
const (
	// AnswerToLifeUniverseEverything - Answer to life, the universe, and everything.
	AnswerToLifeUniverseEverything = 42
	// FingerprintLeeway - Leeway for fingerprint matching in unseal operations.
	FingerprintLeeway = 1
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
	// StatusHealthy - Healthy status string.
	StatusHealthy = "HEALTHY"
	// StatusUnhealthy - Unhealthy status string.
	StatusUnhealthy = "UNHEALTHY"
	// ModeNameDirect - Direct dependency mode name.
	ModeNameDirect = "direct"
	// ModeNameAll - All dependencies mode name.
	ModeNameAll = "all"
	// LogLevelAll - All log level for comprehensive logging.
	LogLevelAll = "ALL"
	// EmptyString - Empty string constant.
	EmptyString = ""
	// MaxInt64 - Maximum int64 value (= 2^63-1 = 9,223,372,036,854,775,807).
	MaxInt64 = int64(^uint64(0) >> 1)
	// MaxLifetimeValues - Max int64 as uint64.
	MaxLifetimeValues = uint64(MaxInt64)
	// MaxLifetimeDuration - Max int64 as nanoseconds (= 292.47 years).
	MaxLifetimeDuration = time.Duration(MaxInt64)
	// LogLevelAllValue - Lowest possible level (enable everything).
	LogLevelAllValue = math.MinInt
	// LogLevelTrace - Trace log level.
	LogLevelTrace = -8
	// LogLevelConfig - Config log level.
	LogLevelConfig = -2
	// LogLevelWarn - Warning log level.
	LogLevelWarn = 4
	// LogLevelMax - Maximum log level.
	LogLevelMax = math.MaxInt
	// BoolDefaultHelp - Default help flag value.
	BoolDefaultHelp = false
	// BoolDefaultVerboseMode - Default verbose mode flag value.
	BoolDefaultVerboseMode = false
	// BoolDefaultDevMode - Default dev mode flag value.
	BoolDefaultDevMode = false
	// BoolDefaultDryRun - Default dry run flag value.
	BoolDefaultDryRun = false
	// BoolDefaultOTLP - Default OTLP flag value.
	BoolDefaultOTLP = false
	// BoolDefaultOTLPConsole - Default OTLP console flag value.
	BoolDefaultOTLPConsole = false
	// BoolCSRFTokenCookieSecure - Default CSRF token cookie secure flag.
	BoolCSRFTokenCookieSecure = true
	// BoolCSRFTokenCookieHTTPOnly - Default CSRF token cookie HTTPOnly flag.
	BoolCSRFTokenCookieHTTPOnly = false
	// BoolCSRFTokenCookieSessionOnly - Default CSRF token cookie session only flag.
	BoolCSRFTokenCookieSessionOnly = true
	// BoolCSRFTokenSingleUseToken - Default CSRF token single use flag.
	BoolCSRFTokenSingleUseToken = false
)
