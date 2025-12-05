// Copyright (c) 2025 Justin Cranford

// Package magic provides shared constants for the CA package.
package magic

import "time"

// backdateBuffer allows slight backdating to handle clock skew.
const BackdateBuffer = 1 * time.Minute

// hexBase is the base for hex encoding serial numbers.
const HexBase = 16

// serialNumberLength is 20 bytes (160 bits) per CA/Browser Forum requirements.
const SerialNumberLength = 20

// DefaultPageLimit is the default number of items per page for pagination.
const DefaultPageLimit = 20

// File permission constants.
const (
	DirPermissions     = 0o755
	FilePermissions    = 0o644
	KeyFilePermissions = 0o600
)