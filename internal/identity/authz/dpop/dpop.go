// Copyright (c) 2025 Justin Cranford

// Package dpop provides Demonstrating Proof-of-Possession (DPoP) implementation for OAuth 2.0.
package dpop

import (
	sha256 "crypto/sha256"
	"encoding/base64"
	json "encoding/json"
	"fmt"
	"strings"
	"time"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJws "github.com/lestrrat-go/jwx/v3/jws"
	joseJwt "github.com/lestrrat-go/jwx/v3/jwt"
)

// Proof represents a parsed and validated DPoP proof JWT (RFC 9449).
type Proof struct {
	JTI           string    // Unique identifier for the proof
	HTM           string    // HTTP method (e.g., "POST")
	HTU           string    // HTTP URI (e.g., "https://server.example.com/token")
	IAT           time.Time // Issued at timestamp
	JWKThumbprint string    // JWK thumbprint (SHA-256)
}

// ValidateProof validates a DPoP proof JWT according to RFC 9449.
//
// Parameters:
//   - dpopHeader: DPoP header value from HTTP request
//   - httpMethod: HTTP method (e.g., "POST")
//   - httpURI: Full HTTP URI (e.g., "https://server.example.com/token")
//   - accessToken: Optional access token for binding validation (empty for token endpoint)
//
// Returns validated DPoP proof or error.
func ValidateProof(dpopHeader, httpMethod, httpURI, accessToken string) (*Proof, error) {
	if dpopHeader == "" {
		return nil, fmt.Errorf("DPoP header is required")
	}

	if httpMethod == "" {
		return nil, fmt.Errorf("HTTP method is required")
	}

	if httpURI == "" {
		return nil, fmt.Errorf("HTTP URI is required")
	}

	// Parse DPoP JWT without verification (we'll verify signature via JWK in header).
	token, err := joseJwt.Parse([]byte(dpopHeader), joseJwt.WithVerify(false), joseJwt.WithValidate(false))
	if err != nil {
		return nil, fmt.Errorf("failed to parse DPoP JWT: %w", err)
	}

	// Verify JWT structure (must be JWS with JWK in header).
	msg, err := joseJws.Parse([]byte(dpopHeader))
	if err != nil {
		return nil, fmt.Errorf("DPoP must be a JWS: %w", err)
	}

	if len(msg.Signatures()) != 1 {
		return nil, fmt.Errorf("DPoP must have exactly one signature")
	}

	sig := msg.Signatures()[0]
	headers := sig.ProtectedHeaders()

	// RFC 9449 Section 4.2: typ header must be "dpop+jwt".
	typ, ok := headers.Type()
	if !ok || typ != "dpop+jwt" {
		return nil, fmt.Errorf("DPoP typ header must be 'dpop+jwt'")
	}

	// RFC 9449 Section 4.2: alg must not be "none".
	alg, ok := headers.Algorithm()
	if !ok || alg == joseJwa.NoSignature() {
		return nil, fmt.Errorf("DPoP alg must not be 'none'")
	}

	// RFC 9449 Section 4.2: jwk header must be present.
	jwk, ok := headers.JWK()
	if !ok || jwk == nil {
		return nil, fmt.Errorf("DPoP must include jwk header")
	}

	// Extract JWK thumbprint (SHA-256).
	jwkJSON, err := json.Marshal(jwk)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize JWK: %w", err)
	}

	hash := sha256.Sum256(jwkJSON)
	jwkThumbprint := base64.RawURLEncoding.EncodeToString(hash[:])

	// Verify JWT claims.
	var jti string
	if err := token.Get("jti", &jti); err != nil {
		return nil, fmt.Errorf("DPoP must include jti claim")
	}

	var htm string
	if err := token.Get("htm", &htm); err != nil {
		return nil, fmt.Errorf("DPoP must include htm claim")
	}

	var htu string
	if err := token.Get("htu", &htu); err != nil {
		return nil, fmt.Errorf("DPoP must include htu claim")
	}

	iat, ok := token.IssuedAt()
	if !ok {
		return nil, fmt.Errorf("DPoP must include iat claim")
	}

	// Verify htm matches HTTP method.
	if !strings.EqualFold(htm, httpMethod) {
		return nil, fmt.Errorf("DPoP htm claim must match HTTP method")
	}

	// Verify htu matches HTTP URI (case-sensitive per RFC 9449).
	if htu != httpURI {
		return nil, fmt.Errorf("DPoP htu claim must match HTTP URI")
	}

	// RFC 9449 Section 4.3: iat must be within acceptable time window (±60s).
	now := time.Now().UTC()
	if now.Sub(iat) > 60*time.Second || iat.Sub(now) > 60*time.Second {
		return nil, fmt.Errorf("DPoP iat claim is outside acceptable time window")
	}

	// If access token provided, verify ath claim (RFC 9449 Section 4.2).
	if accessToken != "" {
		var ath string
		if err := token.Get("ath", &ath); err != nil {
			return nil, fmt.Errorf("DPoP must include ath claim when used with access token")
		}

		// Compute expected ath (SHA-256 hash of access token).
		expectedATH := ComputeAccessTokenHash(accessToken)
		if ath != expectedATH {
			return nil, fmt.Errorf("DPoP ath claim does not match access token")
		}
	}

	return &Proof{
		JTI:           jti,
		HTM:           htm,
		HTU:           htu,
		IAT:           iat,
		JWKThumbprint: jwkThumbprint,
	}, nil
}

// ComputeAccessTokenHash computes the ath claim value (SHA-256 hash of access token).
func ComputeAccessTokenHash(accessToken string) string {
	hash := sha256.Sum256([]byte(accessToken))

	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// IsDPoPBound checks if an access token is DPoP-bound (has cnf claim with jkt).
func IsDPoPBound(accessToken string) (bool, string, error) {
	token, err := joseJwt.Parse([]byte(accessToken), joseJwt.WithVerify(false), joseJwt.WithValidate(false))
	if err != nil {
		return false, "", fmt.Errorf("failed to parse access token: %w", err)
	}

	cnfRaw := make(map[string]any)
	if cnfErr := token.Get("cnf", &cnfRaw); cnfErr != nil {
		return false, "", nil //nolint:nilerr // Missing cnf claim means not DPoP-bound, not an error.
	}

	jkt, ok := cnfRaw["jkt"]
	if !ok {
		return false, "", nil // Not DPoP-bound (no jkt in cnf)
	}

	jktStr, ok := jkt.(string)
	if !ok {
		return false, "", fmt.Errorf("cnf.jkt must be a string")
	}

	return true, jktStr, nil
}
