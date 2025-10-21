// Package constants provides magic values and other useful constants.
// This package centralizes commonly used constants to avoid magic numbers
// and improve code maintainability across the cryptoutil codebase.
package constants

// File permission constants for common use cases.
// These constants help avoid magic numbers in file operations.
const (
	// PermOwnerReadWriteOnly - Owner read/write only (0o600).
	PermOwnerReadWriteOnly = 0o600
	// PermOwnerReadWriteGroupRead - Owner read/write, group/other read (0o644).
	PermOwnerReadWriteGroupRead = 0o644
	// PermAllReadWrite - All users read/write (0o666).
	PermAllReadWrite = 0o666
	// PermOwnerReadWriteExecuteGroupReadExecute - Owner read/write/execute, group read/execute (0o750).
	PermOwnerReadWriteExecuteGroupReadExecute = 0o750
	// PermOwnerReadWriteExecuteGroupOtherReadExecute - Owner read/write/execute, group/other read/execute (0o755).
	PermOwnerReadWriteExecuteGroupOtherReadExecute = 0o755
)
