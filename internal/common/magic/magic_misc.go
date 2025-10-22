// Package magic provides commonly used magic numbers and values as named constants.
// This file contains miscellaneous constants.
package magic

// Miscellaneous constants.
const (
	// AnswerToLifeUniverseEverything - Answer to life, the universe, and everything.
	AnswerToLifeUniverseEverything = 42

	// EmptyString - Empty string constant.
	EmptyString = ""

	// StringUTCFormat - UTC time format string.
	StringUTCFormat = "2006-01-02T15:04:05Z"
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

	// DefaultBoolHelp - Default help flag value.
	DefaultBoolHelp = false
	// DefaultBoolDevMode - Default dev mode flag value.
	DefaultBoolDevMode = false
	// DefaultBoolDryRun - Default dry run flag value.
	DefaultBoolDryRun = false
)
