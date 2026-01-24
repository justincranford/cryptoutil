// Copyright (c) 2025 Justin Cranford
//
//

package pkce

import (
	sha256 "crypto/sha256"
	"encoding/base64"

	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// ValidateCodeVerifier validates a PKCE code verifier against a stored code challenge.
func ValidateCodeVerifier(codeVerifier, codeChallenge, method string) bool {
	if method == "" {
		method = cryptoutilIdentityMagic.PKCEMethodS256
	}

	switch method {
	case cryptoutilIdentityMagic.PKCEMethodPlain:
		return codeVerifier == codeChallenge
	case cryptoutilIdentityMagic.PKCEMethodS256:
		return ValidateS256(codeVerifier, codeChallenge)
	default:
		return false
	}
}

// ValidateS256 validates S256 PKCE code verifier.
func ValidateS256(codeVerifier, codeChallenge string) bool {
	hash := sha256.Sum256([]byte(codeVerifier))
	computed := base64.RawURLEncoding.EncodeToString(hash[:])

	return computed == codeChallenge
}
