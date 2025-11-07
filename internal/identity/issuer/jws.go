package issuer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	googleUuid "github.com/google/uuid"

	cryptoutilIdentityApperr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// JWSIssuer issues JWS (signed) JWT tokens using cryptoutil crypto primitives.
type JWSIssuer struct {
	issuer         string
	signingKey     any
	signingAlg     string
	accessTokenTTL time.Duration
	idTokenTTL     time.Duration
}

// NewJWSIssuer creates a new JWS issuer with the specified signing key.
func NewJWSIssuer(
	issuer string,
	signingKey any,
	signingAlg string,
	accessTokenTTL time.Duration,
	idTokenTTL time.Duration,
) (*JWSIssuer, error) {
	// Validate issuer.
	if issuer == "" {
		return nil, cryptoutilIdentityApperr.ErrInvalidConfiguration
	}

	// Validate signing algorithm.
	if signingAlg == "" {
		return nil, cryptoutilIdentityApperr.ErrInvalidConfiguration
	}

	// Validate signing key.
	if signingKey == nil {
		return nil, cryptoutilIdentityApperr.ErrInvalidConfiguration
	}

	return &JWSIssuer{
		issuer:         issuer,
		signingKey:     signingKey,
		signingAlg:     signingAlg,
		accessTokenTTL: accessTokenTTL,
		idTokenTTL:     idTokenTTL,
	}, nil
}

// IssueAccessToken issues a JWS access token with the specified claims.
func (i *JWSIssuer) IssueAccessToken(ctx context.Context, claims map[string]any) (string, error) {
	// Create token claims.
	tokenClaims := make(map[string]any)
	tokenClaims[cryptoutilIdentityMagic.ClaimIss] = i.issuer
	tokenClaims[cryptoutilIdentityMagic.ClaimIat] = time.Now().Unix()
	tokenClaims[cryptoutilIdentityMagic.ClaimExp] = time.Now().Add(i.accessTokenTTL).Unix()
	tokenClaims[cryptoutilIdentityMagic.ClaimJti] = googleUuid.NewString()

	// Add standard claims.
	if sub, ok := claims[cryptoutilIdentityMagic.ClaimSub].(string); ok {
		tokenClaims[cryptoutilIdentityMagic.ClaimSub] = sub
	}

	if aud, ok := claims[cryptoutilIdentityMagic.ClaimAud]; ok {
		tokenClaims[cryptoutilIdentityMagic.ClaimAud] = aud
	}

	if scope, ok := claims[cryptoutilIdentityMagic.ParamScope].(string); ok {
		tokenClaims[cryptoutilIdentityMagic.ParamScope] = scope
	}

	// Add custom claims.
	for key, value := range claims {
		if !isStandardClaim(key) {
			tokenClaims[key] = value
		}
	}

	// Build JWS token (simplified format: base64(header).base64(claims).signature).
	return i.buildJWS(tokenClaims)
}

// IssueIDToken issues a JWS ID token with OIDC claims.
func (i *JWSIssuer) IssueIDToken(ctx context.Context, claims map[string]any) (string, error) {
	// Validate required OIDC claims.
	if _, ok := claims[cryptoutilIdentityMagic.ClaimSub].(string); !ok {
		return "", cryptoutilIdentityApperr.WrapError(
			cryptoutilIdentityApperr.ErrTokenIssuanceFailed,
			fmt.Errorf("missing required claim: %s", cryptoutilIdentityMagic.ClaimSub),
		)
	}

	if _, ok := claims[cryptoutilIdentityMagic.ClaimAud]; !ok {
		return "", cryptoutilIdentityApperr.WrapError(
			cryptoutilIdentityApperr.ErrTokenIssuanceFailed,
			fmt.Errorf("missing required claim: %s", cryptoutilIdentityMagic.ClaimAud),
		)
	}

	// Create token claims.
	tokenClaims := make(map[string]any)
	tokenClaims[cryptoutilIdentityMagic.ClaimIss] = i.issuer
	tokenClaims[cryptoutilIdentityMagic.ClaimIat] = time.Now().Unix()
	tokenClaims[cryptoutilIdentityMagic.ClaimExp] = time.Now().Add(i.idTokenTTL).Unix()
	tokenClaims[cryptoutilIdentityMagic.ClaimJti] = googleUuid.NewString()

	// Add all claims (including OIDC profile/email/address/phone claims).
	for key, value := range claims {
		if !isStandardClaim(key) || key == cryptoutilIdentityMagic.ClaimSub || key == cryptoutilIdentityMagic.ClaimAud {
			tokenClaims[key] = value
		}
	}

	// Build JWS token.
	return i.buildJWS(tokenClaims)
}

// ValidateToken validates a JWS token and returns its claims (stub for now).
func (i *JWSIssuer) ValidateToken(ctx context.Context, tokenString string) (map[string]any, error) {
	// Parse JWT parts (header.claims.signature).
	parts := strings.Split(tokenString, ".")
	if len(parts) != cryptoutilIdentityMagic.JWSPartCount {
		return nil, cryptoutilIdentityApperr.ErrInvalidToken
	}

	// Decode claims.
	claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, cryptoutilIdentityApperr.WrapError(
			cryptoutilIdentityApperr.ErrTokenValidationFailed,
			fmt.Errorf("failed to decode claims: %w", err),
		)
	}

	// Parse claims JSON.
	var claims map[string]any
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return nil, cryptoutilIdentityApperr.WrapError(
			cryptoutilIdentityApperr.ErrTokenValidationFailed,
			fmt.Errorf("failed to parse claims: %w", err),
		)
	}

	// Validate expiration.
	if exp, ok := claims[cryptoutilIdentityMagic.ClaimExp].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, cryptoutilIdentityApperr.ErrTokenExpired
		}
	}

	return claims, nil
}

// buildJWS builds a JWS token (simplified for MVP without signature verification).
func (i *JWSIssuer) buildJWS(claims map[string]any) (string, error) {
	// Create JWS header.
	header := map[string]any{
		"alg": i.signingAlg,
		"typ": "JWT",
	}

	// Encode header.
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", cryptoutilIdentityApperr.WrapError(
			cryptoutilIdentityApperr.ErrTokenIssuanceFailed,
			fmt.Errorf("failed to marshal header: %w", err),
		)
	}

	headerEncoded := base64.RawURLEncoding.EncodeToString(headerBytes)

	// Encode claims.
	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return "", cryptoutilIdentityApperr.WrapError(
			cryptoutilIdentityApperr.ErrTokenIssuanceFailed,
			fmt.Errorf("failed to marshal claims: %w", err),
		)
	}

	claimsEncoded := base64.RawURLEncoding.EncodeToString(claimsBytes)

	// Create signing input.
	signingInput := headerEncoded + "." + claimsEncoded

	// For MVP: use a stub signature (will integrate cryptoutil signing later).
	signature := base64.RawURLEncoding.EncodeToString([]byte("stub-signature"))

	return signingInput + "." + signature, nil
}

// isStandardClaim checks if a claim is a standard JWT/OIDC claim.
func isStandardClaim(claim string) bool {
	standardClaims := []string{
		cryptoutilIdentityMagic.ClaimIss,
		cryptoutilIdentityMagic.ClaimSub,
		cryptoutilIdentityMagic.ClaimAud,
		cryptoutilIdentityMagic.ClaimExp,
		cryptoutilIdentityMagic.ClaimNbf,
		cryptoutilIdentityMagic.ClaimIat,
		cryptoutilIdentityMagic.ClaimJti,
	}

	for _, std := range standardClaims {
		if claim == std {
			return true
		}
	}

	return false
}
