// Copyright (c) 2025 Justin Cranford

// Package mfa provides multi-factor authentication functionality.
package mfa

import (
	crand "crypto/rand"
	"fmt"

	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// GenerateRecoveryCode generates a cryptographically random recovery code.
// Format: XXXX-XXXX-XXXX-XXXX (4 groups of 4 chars).
// Uses charset that excludes ambiguous characters (0/O, 1/I/L).
func GenerateRecoveryCode() (string, error) {
	const (
		groupSize  = 4
		groupCount = 4
		totalChars = groupSize * groupCount
	)

	randomBytes := make([]byte, totalChars)
	if _, err := crand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	charset := cryptoutilIdentityMagic.RecoveryCodeCharset
	code := make([]byte, totalChars)

	for i := range totalChars {
		code[i] = charset[int(randomBytes[i])%len(charset)]
	}

	// Format with hyphens: XXXX-XXXX-XXXX-XXXX.
	formatted := fmt.Sprintf("%s-%s-%s-%s",
		code[0:4],
		code[4:8],
		code[8:12],
		code[12:16])

	return formatted, nil
}

// GenerateRecoveryCodes generates a batch of unique recovery codes.
// Returns error if duplicate code detected (extremely unlikely with 256-bit entropy).
func GenerateRecoveryCodes(count int) ([]string, error) {
	codes := make([]string, count)
	seen := make(map[string]bool, count)

	for i := range count {
		for {
			code, err := GenerateRecoveryCode()
			if err != nil {
				return nil, err
			}

			if !seen[code] {
				codes[i] = code
				seen[code] = true

				break
			}
		}
	}

	return codes, nil
}
