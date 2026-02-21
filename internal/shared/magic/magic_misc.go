// Copyright (c) 2025 Justin Cranford
//
//

package magic

// Miscellaneous constants.
const (
	// AnswerToLifeUniverseEverything - Answer to life, the universe, and everything.
	AnswerToLifeUniverseEverything = 42

	// EmptyString - Empty string constant.
	EmptyString = ""

	// Bits to bytes conversion factor.
	BitsToBytes = 8

	// DefaultProfile - Default profile name. Empty means no profile, use explicit configuration.
	DefaultProfile = EmptyString

	// StringUTCFormat - UTC time format string.
	StringUTCFormat = "2006-01-02T15:04:05Z"
	// StringUUIDRegexPattern - UUID regex pattern for validation.
	StringUUIDRegexPattern = `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`
	// UUIDStringLength - Length of UUID string with hyphens (8-4-4-4-12 = 36).
	UUIDStringLength = 36

	// FilePermissionsDefault - Default file permissions for created files.
	FilePermissionsDefault = 0o600
	// FilePermOwnerReadOnlyGroupOtherReadOnly - Owner read-only, group/other read-only (0o444).
	FilePermOwnerReadOnlyGroupOtherReadOnly = 0o444
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

	// DefaultHelp - Default help flag value.
	DefaultHelp = false
	// DefaultDevMode - Default dev mode flag value.
	DefaultDevMode = false
	// DefaultDemoMode - Default demo mode flag value.
	DefaultDemoMode = false
	// DefaultResetDemoMode - Default reset-demo mode flag value.
	DefaultResetDemoMode = false
	// DefaultDryRun - Default dry run flag value.
	DefaultDryRun = false

	// ServiceVersion - Current service version.
	ServiceVersion = "1.0.0"
)

// DefaultConfigFiles - Default config files slice.
var DefaultConfigFiles = []string{}
