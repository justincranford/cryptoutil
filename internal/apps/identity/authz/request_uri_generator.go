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

// GenerateRequestURI generates a cryptographically random request_uri per RFC 9126.
// Format: urn:ietf:params:oauth:request_uri:<base64url-encoded-random-bytes>
// The request_uri serves as an opaque reference to a pushed authorization request,
// providing request integrity and confidentiality by keeping authorization parameters
// server-side rather than exposing them in browser URLs.
func GenerateRequestURI() (string, error) {
	randomBytes := make([]byte, cryptoutilIdentityMagic.DefaultRequestURILength)
	if _, err := crand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes for request_uri: %w", err)
	}

	// Base64url encode without padding (RFC 4648 Section 5).
	encoded := base64.RawURLEncoding.EncodeToString(randomBytes)

	return cryptoutilIdentityMagic.RequestURIPrefix + encoded, nil
}
