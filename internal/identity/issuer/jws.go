package issuer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	googleUuid "github.com/google/uuid"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// JWSIssuer issues JWS (signed) JWT tokens using versioned signing keys.
type JWSIssuer struct {
	issuer           string
	keyRotationMgr   *KeyRotationManager
	defaultAlgorithm string
	accessTokenTTL   time.Duration
	idTokenTTL       time.Duration
	legacySigningKey any    // Deprecated: for backward compatibility.
	legacySigningAlg string // Deprecated: for backward compatibility.
}

// NewJWSIssuer creates a new JWS issuer with the specified key rotation manager.
func NewJWSIssuer(
	issuer string,
	keyRotationMgr *KeyRotationManager,
	defaultAlgorithm string,
	accessTokenTTL time.Duration,
	idTokenTTL time.Duration,
) (*JWSIssuer, error) {
	// Validate issuer.
	if issuer == "" {
		return nil, cryptoutilIdentityAppErr.ErrInvalidConfiguration
	}

	// Validate signing algorithm.
	if defaultAlgorithm == "" {
		return nil, cryptoutilIdentityAppErr.ErrInvalidConfiguration
	}

	// Key rotation manager is optional for backward compatibility.
	// If nil, must use NewJWSIssuerLegacy instead.
	if keyRotationMgr == nil {
		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrInvalidConfiguration,
			fmt.Errorf("key rotation manager is required; use NewJWSIssuerLegacy for backward compatibility"),
		)
	}

	return &JWSIssuer{
		issuer:           issuer,
		keyRotationMgr:   keyRotationMgr,
		defaultAlgorithm: defaultAlgorithm,
		accessTokenTTL:   accessTokenTTL,
		idTokenTTL:       idTokenTTL,
	}, nil
}

// NewJWSIssuerLegacy creates a new JWS issuer with a single signing key (deprecated).
func NewJWSIssuerLegacy(
	issuer string,
	signingKey any,
	signingAlg string,
	accessTokenTTL time.Duration,
	idTokenTTL time.Duration,
) (*JWSIssuer, error) {
	// Validate issuer.
	if issuer == "" {
		return nil, cryptoutilIdentityAppErr.ErrInvalidConfiguration
	}

	// Validate signing algorithm.
	if signingAlg == "" {
		return nil, cryptoutilIdentityAppErr.ErrInvalidConfiguration
	}

	// Validate signing key.
	if signingKey == nil {
		return nil, cryptoutilIdentityAppErr.ErrInvalidConfiguration
	}

	return &JWSIssuer{
		issuer:           issuer,
		legacySigningKey: signingKey,
		legacySigningAlg: signingAlg,
		accessTokenTTL:   accessTokenTTL,
		idTokenTTL:       idTokenTTL,
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
		return "", cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrTokenIssuanceFailed,
			fmt.Errorf("missing required claim: %s", cryptoutilIdentityMagic.ClaimSub),
		)
	}

	if _, ok := claims[cryptoutilIdentityMagic.ClaimAud]; !ok {
		return "", cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrTokenIssuanceFailed,
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
		return nil, cryptoutilIdentityAppErr.ErrInvalidToken
	}

	// Decode claims.
	claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrTokenValidationFailed,
			fmt.Errorf("failed to decode claims: %w", err),
		)
	}

	// Parse claims JSON.
	var claims map[string]any
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrTokenValidationFailed,
			fmt.Errorf("failed to parse claims: %w", err),
		)
	}

	// Validate expiration.
	if exp, ok := claims[cryptoutilIdentityMagic.ClaimExp].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, cryptoutilIdentityAppErr.ErrTokenExpired
		}
	}

	return claims, nil
}

// buildJWS builds a JWS token using the active signing key.
func (i *JWSIssuer) buildJWS(claims map[string]any) (string, error) {
	var signingAlg string

	var keyID string

	// Get active signing key (or use legacy key).
	if i.keyRotationMgr != nil {
		activeKey, err := i.keyRotationMgr.GetActiveSigningKey()
		if err != nil {
			return "", cryptoutilIdentityAppErr.WrapError(
				cryptoutilIdentityAppErr.ErrTokenIssuanceFailed,
				fmt.Errorf("failed to get active signing key: %w", err),
			)
		}

		signingAlg = activeKey.Algorithm
		keyID = activeKey.KeyID
	} else {
		// Legacy mode: use single signing key.
		signingAlg = i.legacySigningAlg
		keyID = "" // No key ID in legacy mode.
	}

	// Create JWS header.
	header := map[string]any{
		"alg": signingAlg,
		"typ": "JWT",
	}

	// Add key ID if available.
	if keyID != "" {
		header["kid"] = keyID
	}

	// Encode header.
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrTokenIssuanceFailed,
			fmt.Errorf("failed to marshal header: %w", err),
		)
	}

	headerEncoded := base64.RawURLEncoding.EncodeToString(headerBytes)

	// Encode claims.
	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return "", cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrTokenIssuanceFailed,
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
