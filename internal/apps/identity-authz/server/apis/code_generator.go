// Copyright (c) 2025-2026 Justin Cranford.
//
//

package apis

import (
	crand "crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// GenerateAuthorizationCode generates a cryptographically secure random authorization code.
func GenerateAuthorizationCode() (string, error) {
	bytes := make([]byte, cryptoutilSharedMagic.DefaultAuthCodeLength)

	_, err := io.ReadFull(crand.Reader, bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate authorization code: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
