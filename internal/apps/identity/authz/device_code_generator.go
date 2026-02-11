// Copyright (c) 2025 Justin Cranford
//
//

package authz

import (
	crand "crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"

	cryptoutilIdentityMagic "cryptoutil/internal/apps/identity/magic"
)

// GenerateDeviceCode generates a cryptographically secure device code (RFC 8628 Section 3.2).
// Returns a base64url-encoded random value suitable for use as a device_code parameter.
// The device code should have sufficient entropy to prevent brute-force attacks.
func GenerateDeviceCode() (string, error) {
	bytes := make([]byte, cryptoutilIdentityMagic.DefaultDeviceCodeLength)

	if _, err := crand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate device code: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// GenerateUserCode generates a human-readable user code (RFC 8628 Section 6.1).
// Returns a code formatted as "XXXX-YYYY" where X and Y are uppercase alphanumeric
// characters excluding ambiguous characters (0, O, I, 1) for easier manual entry.
//
// Example: "WDJB-MJHT", "A2F3-KJ98".
//
// Format rationale:
// - 8 characters total (4-4 with hyphen separator).
// - Charset excludes ambiguous characters: 0/O, I/1, L/1.
// - ~34 bits of entropy (32^8 ≈ 1.2 trillion combinations).
// - Easy to read aloud and type on mobile devices.
func GenerateUserCode() (string, error) {
	// Exclude ambiguous characters: 0, O, I, 1, L.
	const (
		charset = "ABCDEFGHJKMNPQRSTUVWXYZ23456789" // Removed L from charset
		length  = 8
	)

	code := make([]byte, length)

	for i := range code {
		num, err := crand.Int(crand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("failed to generate user code: %w", err)
		}

		code[i] = charset[num.Int64()]
	}

	// Format as XXXX-YYYY for readability.
	return fmt.Sprintf("%s-%s", string(code[:4]), string(code[4:])), nil
}
