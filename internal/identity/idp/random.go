// Copyright (c) 2025 Justin Cranford
//
//

package idp

import (
	googleUuid "github.com/google/uuid"
)

// generateRandomString generates a cryptographically secure random string.
// Uses UUIDv7 for time-ordered uniqueness to prevent UNIQUE constraint violations
// in parallel test execution.
func generateRandomString(_ int) string {
	// Use UUIDv7 for time-ordered uniqueness (prevents race conditions).
	// UUIDv7 provides 128-bit randomness + timestamp ordering.
	// Length parameter ignored - UUID format is fixed 36 characters.
	return googleUuid.Must(googleUuid.NewV7()).String()
}
