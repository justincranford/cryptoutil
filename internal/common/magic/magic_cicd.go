// Package magic provides magic values and constants for the cryptoutil project.
//
// This file contains constants related to CI/CD operations.
package magic

const (
	// UI constants.
	SeparatorLength = 50

	// Minimum number of regex match groups for action parsing.
	MinActionMatchGroups = 3

	// Cache file permissions (owner read/write only).
	CacheFilePermissions = 0o600

	// Dependency check mode names.
	ModeNameDirect = "direct"
	ModeNameAll    = "all"
)
