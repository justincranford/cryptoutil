// Copyright (c) 2025 Justin Cranford
//
//

// Package pkce provides Proof Key for Code Exchange (PKCE) implementation for OAuth 2.0.
package pkce

import (
	crand "crypto/rand"
	sha256 "crypto/sha256"
	"encoding/base64"
	"fmt"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	codeVerifierLength = 43 // Minimum length per RFC 7636.
)

// GenerateCodeVerifier generates a cryptographically random PKCE code verifier.
func GenerateCodeVerifier() (string, error) {
	bytes := make([]byte, codeVerifierLength)

	_, err := crand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate code verifier: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// GenerateCodeChallenge generates a PKCE code challenge from a code verifier.
func GenerateCodeChallenge(codeVerifier, method string) string {
	if method == "" {
		method = cryptoutilSharedMagic.PKCEMethodS256
	}

	switch method {
	case cryptoutilSharedMagic.PKCEMethodPlain:
		return codeVerifier
	case cryptoutilSharedMagic.PKCEMethodS256:
		return GenerateS256Challenge(codeVerifier)
	default:
		return ""
	}
}

// GenerateS256Challenge generates S256 PKCE code challenge.
func GenerateS256Challenge(codeVerifier string) string {
	hash := sha256.Sum256([]byte(codeVerifier))

	return base64.RawURLEncoding.EncodeToString(hash[:])
}
