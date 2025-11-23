// Copyright (c) 2025 Justin Cranford
//
//

package idp

import (
	crand "crypto/rand"
	"encoding/base64"
	"fmt"

	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// generateRandomString generates a cryptographically secure random string of specified length.
// Uses crypto/rand for security and base64 URL encoding for safe string representation.
func generateRandomString(length int) string {
	// Calculate required bytes for desired string length after base64 encoding.
	// Base64 encoding expands 3 bytes to 4 characters, so we need length*3/4 bytes.
	// Add extra bytes to ensure we have enough after encoding and trimming.
	numBytes := (length*cryptoutilIdentityMagic.Base64ExpansionNumerator)/cryptoutilIdentityMagic.Base64ExpansionDenominator + 1

	bytes := make([]byte, numBytes)

	if _, err := crand.Read(bytes); err != nil {
		// Critical error: crypto/rand failure is a security issue.
		panic(fmt.Sprintf("Failed to generate random string: %v", err))
	}

	// Use URL-safe base64 encoding and trim to exact length.
	return base64.URLEncoding.EncodeToString(bytes)[:length]
}
