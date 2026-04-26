// Copyright (c) 2025 Justin Cranford
//
//

package builder

import (
	"context"
	"fmt"
	"time"

	fiber "github.com/gofiber/fiber/v2"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// JWTAuthMode specifies the authentication mode for the service.
type JWTAuthMode string

const (
	// JWTAuthModeSession uses session-based auth (no JWT validation).
	JWTAuthModeSession JWTAuthMode = "session"
	// JWTAuthModeRequired requires JWT authentication on all protected routes.
	JWTAuthModeRequired JWTAuthMode = "required"
	// JWTAuthModeOptional allows JWT authentication but doesn't require it.
	JWTAuthModeOptional JWTAuthMode = "optional"
)

// JWTContextKey is the context key for validated JWT claims.
type JWTContextKey struct{}

// JWTClaims represents parsed JWT claims.
type JWTClaims struct {
	Subject   string
	Issuer    string
	Audience  []string
	ExpiresAt time.Time
	IssuedAt  time.Time
	NotBefore time.Time
	Scopes    []string
	TenantID  string
	RealmID   string
	Extra     map[string]any
}

// JWTAuthConfig configures JWT authentication.
type JWTAuthConfig struct {
	// Mode specifies when JWT authentication is performed.
	Mode JWTAuthMode

	// JWKSURL is the endpoint for fetching public keys.
	JWKSURL string

	// CacheTTL is how long to cache JWKS keys (default: 5 minutes).
	CacheTTL time.Duration

	// RequiredIssuer validates the 'iss' claim.
	RequiredIssuer string

	// RequiredAudience validates the 'aud' claim.
	RequiredAudience string

	// AllowedAlgorithms restricts signing algorithms (e.g., RS256, ES256).
	AllowedAlgorithms []string

	// SkipPaths are paths that bypass JWT validation.
	SkipPaths []string

	// ErrorDetailLevel controls error verbosity: "minimal", "standard", "verbose".
	ErrorDetailLevel string
}

// JWTValidator defines the interface for JWT validation.
type JWTValidator interface {
	// ValidateToken validates a JWT token string and returns claims.
	ValidateToken(ctx context.Context, tokenString string) (*JWTClaims, error)

	// JWTMiddleware returns a Fiber middleware handler for JWT validation.
	JWTMiddleware() fiber.Handler
}

// NewDefaultJWTAuthConfig creates a default JWT auth configuration.
func NewDefaultJWTAuthConfig() *JWTAuthConfig {
	return &JWTAuthConfig{
		Mode:              JWTAuthModeSession,
		CacheTTL:          cryptoutilSharedMagic.JWKSCacheTTL,
		AllowedAlgorithms: []string{cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilSharedMagic.JoseAlgRS384, cryptoutilSharedMagic.JoseAlgRS512, cryptoutilSharedMagic.JoseAlgES256, cryptoutilSharedMagic.JoseAlgES384, cryptoutilSharedMagic.JoseAlgES512},
		ErrorDetailLevel:  "minimal",
	}
}

// NewKMSJWTAuthConfig creates a JWT auth configuration suitable for KMS-style services.
func NewKMSJWTAuthConfig(jwksURL, issuer, audience string) *JWTAuthConfig {
	return &JWTAuthConfig{
		Mode:              JWTAuthModeRequired,
		JWKSURL:           jwksURL,
		CacheTTL:          cryptoutilSharedMagic.JWKSCacheTTL,
		RequiredIssuer:    issuer,
		RequiredAudience:  audience,
		AllowedAlgorithms: []string{cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilSharedMagic.JoseAlgRS384, cryptoutilSharedMagic.JoseAlgRS512, cryptoutilSharedMagic.JoseAlgES256, cryptoutilSharedMagic.JoseAlgES384, cryptoutilSharedMagic.JoseAlgES512},
		ErrorDetailLevel:  "minimal",
	}
}

// HasScope returns true if the claims contain the specified scope.
func (c *JWTClaims) HasScope(scope string) bool {
	for _, s := range c.Scopes {
		if s == scope {
			return true
		}
	}

	return false
}

// HasAnyScope returns true if the claims contain any of the specified scopes.
func (c *JWTClaims) HasAnyScope(scopes ...string) bool {
	for _, scope := range scopes {
		if c.HasScope(scope) {
			return true
		}
	}

	return false
}

// HasAllScopes returns true if the claims contain all of the specified scopes.
func (c *JWTClaims) HasAllScopes(scopes ...string) bool {
	for _, scope := range scopes {
		if !c.HasScope(scope) {
			return false
		}
	}

	return true
}

// GetJWTClaims extracts JWT claims from request context.
func GetJWTClaims(ctx context.Context) *JWTClaims {
	if claims, ok := ctx.Value(JWTContextKey{}).(*JWTClaims); ok {
		return claims
	}

	return nil
}

// Validate validates the JWT auth configuration.
func (c *JWTAuthConfig) Validate() error {
	if c.Mode == JWTAuthModeSession {
		return nil
	}

	if c.JWKSURL == "" {
		return fmt.Errorf("JWKS URL is required when JWT auth mode is %s", c.Mode)
	}

	if len(c.AllowedAlgorithms) == 0 {
		return fmt.Errorf("at least one allowed algorithm must be specified")
	}

	return nil
}

// IsEnabled returns true if JWT authentication is enabled.
func (c *JWTAuthConfig) IsEnabled() bool {
	return c.Mode != JWTAuthModeSession
}

// IsRequired returns true if JWT authentication is required.
func (c *JWTAuthConfig) IsRequired() bool {
	return c.Mode == JWTAuthModeRequired
}

// ShouldSkipPath returns true if the path should skip JWT validation.
func (c *JWTAuthConfig) ShouldSkipPath(path string) bool {
	for _, skip := range c.SkipPaths {
		if path == skip {
			return true
		}
	}

	return false
}
