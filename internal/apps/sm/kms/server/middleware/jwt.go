// Copyright (c) 2025 Justin Cranford
//
//

package middleware

import (
	"context"
	"errors"
	"fmt"
	http "net/http"
	"strings"
	"sync"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

// RevocationCheckMode controls when revocation checks are performed.
type RevocationCheckMode string

const (
	// RevocationCheckEveryRequest checks every request (most secure, highest latency).
	RevocationCheckEveryRequest RevocationCheckMode = "every-request"

	// RevocationCheckSensitiveOnly checks only for sensitive operations (write, admin).
	RevocationCheckSensitiveOnly RevocationCheckMode = "sensitive-only"

	// RevocationCheckInterval checks at a configurable interval (cached result).
	RevocationCheckInterval RevocationCheckMode = "interval"

	// RevocationCheckDisabled disables revocation checks entirely.
	RevocationCheckDisabled RevocationCheckMode = "disabled"
)

// JWTValidatorConfig configures JWT validation middleware.
type JWTValidatorConfig struct {
	// JWKS endpoint URL for fetching public keys.
	JWKSURL string

	// CacheTTL is how long to cache JWKS keys (default: 5 minutes).
	CacheTTL time.Duration

	// RequiredIssuer validates the 'iss' claim.
	RequiredIssuer string

	// RequiredAudience validates the 'aud' claim.
	RequiredAudience string

	// AllowedAlgorithms restricts signing algorithms.
	AllowedAlgorithms []string

	// RevocationCheckEnabled enables introspection-based revocation checks.
	// Deprecated: Use RevocationCheckMode instead.
	RevocationCheckEnabled bool

	// RevocationCheckMode controls when to perform revocation checks.
	// Options: every-request, sensitive-only, interval, disabled.
	RevocationCheckMode RevocationCheckMode

	// RevocationCheckInterval is the interval between revocation checks.
	// Used when RevocationCheckMode is "interval". Default: 5 minutes.
	RevocationCheckInterval time.Duration

	// SensitiveScopes defines scopes that trigger revocation checks
	// when RevocationCheckMode is "sensitive-only".
	SensitiveScopes []string

	// IntrospectionURL for checking token revocation.
	IntrospectionURL string

	// IntrospectionClientID for authenticating introspection requests.
	IntrospectionClientID string

	// IntrospectionClientSecret for authenticating introspection requests.
	IntrospectionClientSecret string

	// ErrorDetailLevel controls how much info to return in errors.
	// Values: "minimal" (prod), "standard", "verbose" (dev).
	ErrorDetailLevel string
}

// JWTValidator validates JWT tokens.
type JWTValidator struct {
	config     JWTValidatorConfig
	cache      *jwksCache
	httpClient *http.Client
}

// jwksCache caches JWKS keys with TTL.
type jwksCache struct {
	sync.RWMutex
	keySet     joseJwk.Set
	lastUpdate time.Time
	ttl        time.Duration
}

// JWTContextKey is the context key for validated JWT claims.
type JWTContextKey struct{}

// JWTClaims represents validated JWT claims.
type JWTClaims struct {
	Subject   string    `json:"sub"`
	Issuer    string    `json:"iss"`
	Audience  []string  `json:"aud"`
	ExpiresAt time.Time `json:"exp"`
	IssuedAt  time.Time `json:"iat"`
	NotBefore time.Time `json:"nbf"`
	JTI       string    `json:"jti"`

	// OIDC standard claims.
	Name              string `json:"name,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	Email             string `json:"email,omitempty"`
	EmailVerified     bool   `json:"email_verified,omitempty"`

	// Scope claim.
	Scope  string   `json:"scope,omitempty"`
	Scopes []string `json:"-"` // Parsed from scope claim.

	// Custom claims map for extension.
	Custom map[string]any `json:"-"`
}

// JWT validation constants.
const (
	defaultJWKSCacheTTL   = 5 * time.Minute
	defaultHTTPTimeout    = 10 * time.Second
	errorDetailLevelMin   = "minimal"
	errorDetailLevelStd   = "standard"
	errorDetailLevelDebug = "verbose"
)

// NewJWTValidator creates a new JWT validator.
func NewJWTValidator(config JWTValidatorConfig) (*JWTValidator, error) {
	if config.JWKSURL == "" {
		return nil, errors.New("JWKS URL is required")
	}

	if config.CacheTTL == 0 {
		config.CacheTTL = defaultJWKSCacheTTL
	}

	if config.ErrorDetailLevel == "" {
		config.ErrorDetailLevel = errorDetailLevelMin
	}

	return &JWTValidator{
		config: config,
		cache: &jwksCache{
			ttl: config.CacheTTL,
		},
		httpClient: &http.Client{
			Timeout: defaultHTTPTimeout,
		},
	}, nil
}

// JWTMiddleware returns Fiber middleware for JWT validation.
func (v *JWTValidator) JWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract token from Authorization header.
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return v.unauthorizedError(c, "missing_token", "Authorization header is required")
		}

		// Check Bearer prefix.
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return v.unauthorizedError(c, "invalid_token", "Authorization header must use Bearer scheme")
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			return v.unauthorizedError(c, "invalid_token", "Bearer token is empty")
		}

		// Validate token.
		claims, err := v.ValidateToken(c.Context(), tokenString)
		if err != nil {
			return v.handleValidationError(c, err)
		}

		// Store claims in context.
		ctx := context.WithValue(c.UserContext(), JWTContextKey{}, claims)
		c.SetUserContext(ctx)

		return c.Next()
	}
}

// ValidateToken validates a JWT token and returns claims.
func (v *JWTValidator) ValidateToken(ctx context.Context, tokenString string) (*JWTClaims, error) {
	// Get JWKS.
	keySet, err := v.getJWKS(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get JWKS: %w", err)
	}

	// Build validation options.
	options := []jwt.ParseOption{
		jwt.WithKeySet(keySet),
		jwt.WithValidate(true),
	}

	// Add issuer validation if configured.
	if v.config.RequiredIssuer != "" {
		options = append(options, jwt.WithIssuer(v.config.RequiredIssuer))
	}

	// Add audience validation if configured.
	if v.config.RequiredAudience != "" {
		options = append(options, jwt.WithAudience(v.config.RequiredAudience))
	}

	// Parse and validate token.
	token, err := jwt.ParseString(tokenString, options...)
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	// Verify algorithm if restricted.
	if len(v.config.AllowedAlgorithms) > 0 {
		// Note: Algorithm verification happens during ParseString with KeySet.
		// This is a secondary check for explicit algorithm allowlist.
		var alg string
		if err := token.Get("alg", &alg); err == nil && !v.isAlgorithmAllowed(alg) {
			return nil, fmt.Errorf("algorithm %s is not allowed", alg)
		}
	}

	// Extract claims first (needed for sensitive scope check).
	claims := v.extractClaims(token)

	// Check revocation based on configured mode.
	if err := v.performRevocationCheck(ctx, tokenString, claims); err != nil {
		return nil, err
	}

	return claims, nil
}

// performRevocationCheck checks token revocation based on configured mode.
