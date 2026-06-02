// Copyright (c) 2025-2026 Justin Cranford.
// Package magic provides shared constants for the CA package.
package magic

import "time"

// BackdateBuffer allows slight backdating to handle clock skew.
const BackdateBuffer = 1 * time.Minute

// HexBase is the base for hex encoding serial numbers.
const HexBase = 16

// SerialNumberLength is 20 bytes (160 bits) per CA/Browser Forum requirements.
const SerialNumberLength = 20

// DefaultPageLimit is the default number of items per page for pagination.
const DefaultPageLimit = 20

// BitsPerByte is the number of bits per byte (for key size calculations).
const BitsPerByte = 8

// CA_ITEMS is the pki-ca table name for CAItem persistence model.
const CA_ITEMS = "ca_items"

// PKI CA hierarchy and status labels.
const (
	ROOT         = "root"
	INTERMEDIATE = "intermediate"
	ISSUING      = "issuing"
	ACTIVE       = "active"
	PENDING      = "pending"
	SUSPENDED    = "suspended"
	REVOKED      = "revoked"
	EXPIRED      = "expired"
)

// PKI CA test labels and curve aliases.
const (
	TEST_CA_NAME = "test-ca"
	TEST_CA_CN   = "Test CA"
	P256         = "P-256"
	P384         = "P-384"
	P521         = "P-521"
)

// File permission constants.
const (
	DirPermissions     = 0o755
	FilePermissions    = 0o644
	KeyFilePermissions = 0o600
)
