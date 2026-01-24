// Copyright (c) 2025 Justin Cranford
//
//

package middleware

import (
	"context"
	"crypto"
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	rsa "crypto/rsa"
	json "encoding/json"
	"errors"
	"fmt"
	http "net/http"
	"strings"
	"sync"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
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
func (v *JWTValidator) performRevocationCheck(ctx context.Context, tokenString string, claims *JWTClaims) error {
	// Determine if revocation check is needed.
	shouldCheck := v.shouldPerformRevocationCheck(claims)
	if !shouldCheck {
		return nil
	}

	// Perform introspection check.
	active, err := v.checkRevocation(ctx, tokenString)
	if err != nil {
		return fmt.Errorf("revocation check failed: %w", err)
	}

	if !active {
		return errors.New("token has been revoked")
	}

	return nil
}

// shouldPerformRevocationCheck determines if revocation check is needed.
func (v *JWTValidator) shouldPerformRevocationCheck(claims *JWTClaims) bool {
	// Check if introspection URL is configured.
	if v.config.IntrospectionURL == "" {
		return false
	}

	// Backwards compatibility: check old boolean flag.
	if v.config.RevocationCheckEnabled {
		return true
	}

	// Check based on configured mode.
	switch v.config.RevocationCheckMode {
	case RevocationCheckEveryRequest:
		return true
	case RevocationCheckSensitiveOnly:
		return v.hasSensitiveScope(claims)
	case RevocationCheckInterval:
		// NOTE: Interval-based caching with token JTI will be implemented when RevocationCheckInterval mode is actively used.
		// For now, treat as every-request for security.
		return true
	case RevocationCheckDisabled, "":
		return false
	default:
		return false
	}
}

// hasSensitiveScope checks if claims contain any sensitive scopes.
func (v *JWTValidator) hasSensitiveScope(claims *JWTClaims) bool {
	if len(v.config.SensitiveScopes) == 0 {
		// Default sensitive scopes if not configured.
		defaultSensitiveScopes := []string{"admin", "write", "delete", "kms:admin", "kms:write"}

		for _, scope := range defaultSensitiveScopes {
			if claims.HasScope(scope) {
				return true
			}
		}

		return false
	}

	// Check configured sensitive scopes.
	for _, scope := range v.config.SensitiveScopes {
		if claims.HasScope(scope) {
			return true
		}
	}

	return false
}

// getJWKS retrieves JWKS from cache or fetches from URL.
func (v *JWTValidator) getJWKS(ctx context.Context) (joseJwk.Set, error) {
	v.cache.RLock()

	if v.cache.keySet != nil && time.Since(v.cache.lastUpdate) < v.cache.ttl {
		keySet := v.cache.keySet
		v.cache.RUnlock()

		return keySet, nil
	}

	v.cache.RUnlock()

	// Fetch fresh JWKS.
	return v.refreshJWKS(ctx)
}

// refreshJWKS fetches fresh JWKS from the configured URL.
func (v *JWTValidator) refreshJWKS(ctx context.Context) (joseJwk.Set, error) {
	v.cache.Lock()
	defer v.cache.Unlock()

	// Double-check after acquiring write lock.
	if v.cache.keySet != nil && time.Since(v.cache.lastUpdate) < v.cache.ttl {
		return v.cache.keySet, nil
	}

	// Fetch JWKS.
	keySet, err := joseJwk.Fetch(ctx, v.config.JWKSURL, joseJwk.WithHTTPClient(v.httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS from %s: %w", v.config.JWKSURL, err)
	}

	v.cache.keySet = keySet
	v.cache.lastUpdate = time.Now()

	return keySet, nil
}

// checkRevocation verifies token is not revoked via introspection.
func (v *JWTValidator) checkRevocation(ctx context.Context, tokenString string) (bool, error) {
	// Build introspection request.
	reqBody := "token=" + tokenString

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.config.IntrospectionURL, strings.NewReader(reqBody))
	if err != nil {
		return false, fmt.Errorf("failed to create introspection request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Add client authentication.
	if v.config.IntrospectionClientID != "" && v.config.IntrospectionClientSecret != "" {
		req.SetBasicAuth(v.config.IntrospectionClientID, v.config.IntrospectionClientSecret)
	}

	// Execute request.
	resp, err := v.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("introspection request failed: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("introspection returned status %d", resp.StatusCode)
	}

	// Parse response.
	var result struct {
		Active bool `json:"active"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("failed to parse introspection response: %w", err)
	}

	return result.Active, nil
}

// extractClaims converts JWT token to JWTClaims struct.
func (v *JWTValidator) extractClaims(token jwt.Token) *JWTClaims {
	claims := &JWTClaims{
		Custom: make(map[string]any),
	}

	// Extract standard claims (all return value, ok pattern in jwx v3).
	if sub, ok := token.Subject(); ok {
		claims.Subject = sub
	}

	if iss, ok := token.Issuer(); ok {
		claims.Issuer = iss
	}

	if aud, ok := token.Audience(); ok {
		claims.Audience = aud
	}

	if exp, ok := token.Expiration(); ok {
		claims.ExpiresAt = exp
	}

	if iat, ok := token.IssuedAt(); ok {
		claims.IssuedAt = iat
	}

	if nbf, ok := token.NotBefore(); ok {
		claims.NotBefore = nbf
	}

	if jti, ok := token.JwtID(); ok {
		claims.JTI = jti
	}

	// Extract OIDC claims.
	var name string
	if err := token.Get("name", &name); err == nil {
		claims.Name = name
	}

	var username string
	if err := token.Get("preferred_username", &username); err == nil {
		claims.PreferredUsername = username
	}

	var email string
	if err := token.Get("email", &email); err == nil {
		claims.Email = email
	}

	var emailVerified bool
	if err := token.Get("email_verified", &emailVerified); err == nil {
		claims.EmailVerified = emailVerified
	}

	// Extract scope claim.
	var scope string
	if err := token.Get("scope", &scope); err == nil {
		claims.Scope = scope
		claims.Scopes = strings.Fields(claims.Scope)
	}

	return claims
}

// isAlgorithmAllowed checks if algorithm is in allowed list.
func (v *JWTValidator) isAlgorithmAllowed(alg string) bool {
	for _, allowed := range v.config.AllowedAlgorithms {
		if alg == allowed {
			return true
		}
	}

	return false
}

// unauthorizedError returns a 401 error response.
func (v *JWTValidator) unauthorizedError(c *fiber.Ctx, errorCode, message string) error {
	response := fiber.Map{"error": errorCode}

	if v.config.ErrorDetailLevel != errorDetailLevelMin {
		response["message"] = message
	}

	if err := c.Status(fiber.StatusUnauthorized).JSON(response); err != nil {
		return fmt.Errorf("failed to send unauthorized response: %w", err)
	}

	return nil
}

// forbiddenError returns a 403 error response.
func (v *JWTValidator) forbiddenError(c *fiber.Ctx, errorCode, message string) error {
	response := fiber.Map{"error": errorCode}

	if v.config.ErrorDetailLevel != errorDetailLevelMin {
		response["message"] = message
	}

	if err := c.Status(fiber.StatusForbidden).JSON(response); err != nil {
		return fmt.Errorf("failed to send forbidden response: %w", err)
	}

	return nil
}

// handleValidationError converts validation error to appropriate HTTP response.
func (v *JWTValidator) handleValidationError(c *fiber.Ctx, err error) error {
	errMsg := err.Error()

	// Check for specific error types.
	switch {
	case strings.Contains(errMsg, "expired"):
		return v.unauthorizedError(c, "token_expired", "Token has expired")
	case strings.Contains(errMsg, "revoked"):
		return v.unauthorizedError(c, "token_revoked", "Token has been revoked")
	case strings.Contains(errMsg, "issuer"):
		return v.unauthorizedError(c, "invalid_issuer", "Token issuer is invalid")
	case strings.Contains(errMsg, "audience"):
		return v.unauthorizedError(c, "invalid_audience", "Token audience is invalid")
	case strings.Contains(errMsg, "signature"):
		return v.unauthorizedError(c, "invalid_signature", "Token signature is invalid")
	default:
		return v.unauthorizedError(c, "invalid_token", "Token validation failed")
	}
}

// GetJWTClaims extracts JWT claims from request context.
func GetJWTClaims(ctx context.Context) *JWTClaims {
	if claims, ok := ctx.Value(JWTContextKey{}).(*JWTClaims); ok {
		return claims
	}

	return nil
}

// HasScope checks if the JWT claims contain a specific scope.
func (c *JWTClaims) HasScope(scope string) bool {
	for _, s := range c.Scopes {
		if s == scope {
			return true
		}
	}

	return false
}

// HasAnyScope checks if the JWT claims contain any of the specified scopes.
func (c *JWTClaims) HasAnyScope(scopes ...string) bool {
	for _, scope := range scopes {
		if c.HasScope(scope) {
			return true
		}
	}

	return false
}

// HasAllScopes checks if the JWT claims contain all specified scopes.
func (c *JWTClaims) HasAllScopes(scopes ...string) bool {
	for _, scope := range scopes {
		if !c.HasScope(scope) {
			return false
		}
	}

	return true
}

// RequireScopeMiddleware returns middleware that requires specific scopes.
func RequireScopeMiddleware(validator *JWTValidator, requiredScopes ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := GetJWTClaims(c.UserContext())
		if claims == nil {
			return validator.unauthorizedError(c, "missing_claims", "JWT claims not found in context")
		}

		if !claims.HasAllScopes(requiredScopes...) {
			return validator.forbiddenError(c, "insufficient_scope", "Missing required scopes")
		}

		return c.Next()
	}
}

// RequireAnyScopeMiddleware returns middleware that requires any of the specified scopes.
func RequireAnyScopeMiddleware(validator *JWTValidator, requiredScopes ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := GetJWTClaims(c.UserContext())
		if claims == nil {
			return validator.unauthorizedError(c, "missing_claims", "JWT claims not found in context")
		}

		if !claims.HasAnyScope(requiredScopes...) {
			return validator.forbiddenError(c, "insufficient_scope", "Missing required scopes")
		}

		return c.Next()
	}
}

// DefaultAllowedAlgorithms returns FIPS-approved algorithms.
func DefaultAllowedAlgorithms() []string {
	return []string{
		joseJwa.RS256().String(),
		joseJwa.RS384().String(),
		joseJwa.RS512().String(),
		joseJwa.ES256().String(),
		joseJwa.ES384().String(),
		joseJwa.ES512().String(),
		joseJwa.PS256().String(),
		joseJwa.PS384().String(),
		joseJwa.PS512().String(),
		joseJwa.EdDSA().String(),
	}
}

// PublicKeyFromJWK extracts public key from JWK.
func PublicKeyFromJWK(key joseJwk.Key) (crypto.PublicKey, error) {
	var pubKey any
	if err := joseJwk.Export(key, &pubKey); err != nil {
		return nil, fmt.Errorf("failed to extract public key: %w", err)
	}

	// Verify it's a supported public key type.
	switch pk := pubKey.(type) {
	case *rsa.PublicKey:
		return pk, nil
	case *ecdsa.PublicKey:
		return pk, nil
	case ed25519.PublicKey:
		return pk, nil
	default:
		return nil, fmt.Errorf("unsupported key type: %T", pubKey)
	}
}
