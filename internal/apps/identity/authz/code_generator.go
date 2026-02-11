// Copyright (c) 2025 Justin Cranford
//
//

package authz

import (
	crand "crypto/rand"
	"encoding/base64"
	"fmt"

	cryptoutilIdentityMagic "cryptoutil/internal/apps/identity/magic"
)

// GenerateAuthorizationCode generates a cryptographically secure random authorization code.
func GenerateAuthorizationCode() (string, error) {
	bytes := make([]byte, cryptoutilIdentityMagic.DefaultAuthCodeLength)

	_, err := crand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate authorization code: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
