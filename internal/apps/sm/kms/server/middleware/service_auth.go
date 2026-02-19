// Copyright (c) 2025 Justin Cranford
//
//

package middleware

import (
	"context"
	"crypto/subtle"
	"crypto/tls"
	"errors"
	"fmt"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
)

// AuthMethod represents the authentication method used.
type AuthMethod string

const (
	// AuthMethodJWT uses JWT bearer tokens.
	AuthMethodJWT AuthMethod = "jwt"

	// AuthMethodMTLS uses mutual TLS certificate authentication.
	AuthMethodMTLS AuthMethod = "mtls"

	// AuthMethodAPIKey uses API key authentication via header.
	AuthMethodAPIKey AuthMethod = "api-key"

	// AuthMethodClientCredentials uses OAuth2 client credentials.
	AuthMethodClientCredentials AuthMethod = "client-credentials"
)

// ServiceAuthConfig configures service-to-service authentication.
type ServiceAuthConfig struct {
	// AllowedMethods specifies which auth methods are allowed.
	// Order matters - first matching method wins.
	AllowedMethods []AuthMethod

	// JWTConfig for JWT authentication.
	JWTConfig *JWTValidatorConfig

	// MTLSConfig for mTLS authentication.
	MTLSConfig *MTLSConfig

	// APIKeyConfig for API key authentication.
	APIKeyConfig *APIKeyConfig

	// ClientCredentialsConfig for OAuth2 client credentials.
	ClientCredentialsConfig *ClientCredentialsConfig

	// ErrorDetailLevel controls error verbosity.
	ErrorDetailLevel string
}

// MTLSConfig configures mTLS authentication.
type MTLSConfig struct {
	// RequireClientCert requires client certificate.
	RequireClientCert bool

	// AllowedCNs restricts allowed Common Names.
	AllowedCNs []string

	// AllowedOUs restricts allowed Organizational Units.
	AllowedOUs []string

	// AllowedDNSSANs restricts allowed DNS Subject Alternative Names.
	AllowedDNSSANs []string
}

// APIKeyConfig configures API key authentication.
type APIKeyConfig struct {
	// HeaderName is the header containing the API key (default: X-API-Key).
	HeaderName string

	// ValidKeys is a map of API key to service name/metadata.
	ValidKeys map[string]string

	// KeyValidator is a function to validate API keys dynamically.
	// If provided, takes precedence over ValidKeys.
	KeyValidator func(ctx context.Context, apiKey string) (serviceName string, valid bool, err error)
}

// ClientCredentialsConfig configures OAuth2 client credentials authentication.
type ClientCredentialsConfig struct {
	// TokenEndpoint is the OAuth2 token endpoint URL.
	TokenEndpoint string

	// IntrospectionEndpoint is the OAuth2 introspection endpoint URL.
	IntrospectionEndpoint string

	// ValidateClientID validates the client ID from introspection response.
	ValidateClientID func(clientID string) bool
}

// ServiceAuthContextKey is the context key for service auth info.
type ServiceAuthContextKey struct{}

// ServiceAuthInfo contains authenticated service information.
type ServiceAuthInfo struct {
	// Method is the authentication method used.
	Method AuthMethod

	// ServiceName identifies the authenticated service.
	ServiceName string

	// ClientID for OAuth2-based auth methods.
	ClientID string

	// Subject for JWT-based auth.
	Subject string

	// CertificateCN for mTLS auth.
	CertificateCN string

	// Scopes from JWT or introspection.
	Scopes []string

	// Metadata contains additional auth-specific data.
	Metadata map[string]any
}

// ServiceAuthMiddleware provides configurable service-to-service authentication.
type ServiceAuthMiddleware struct {
	config       ServiceAuthConfig
	jwtValidator *JWTValidator
}

// NewServiceAuthMiddleware creates a new service auth middleware.
func NewServiceAuthMiddleware(config ServiceAuthConfig) (*ServiceAuthMiddleware, error) {
	if len(config.AllowedMethods) == 0 {
		return nil, errors.New("at least one auth method must be allowed")
	}

	if config.ErrorDetailLevel == "" {
		config.ErrorDetailLevel = errorDetailLevelMin
	}

	var jwtValidator *JWTValidator

	// Initialize JWT validator if JWT is allowed.
	for _, method := range config.AllowedMethods {
		if method == AuthMethodJWT && config.JWTConfig != nil {
			var err error

			jwtValidator, err = NewJWTValidator(*config.JWTConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to create JWT validator: %w", err)
			}

			break
		}
	}

	return &ServiceAuthMiddleware{
		config:       config,
		jwtValidator: jwtValidator,
	}, nil
}

// Middleware returns the Fiber middleware handler.
func (m *ServiceAuthMiddleware) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Try each allowed method in order.
		for _, method := range m.config.AllowedMethods {
			authInfo, err := m.tryAuthenticate(c, method)
			if err != nil {
				continue // Try next method.
			}

			if authInfo != nil {
				// Authentication successful - store info in context.
				ctx := context.WithValue(c.UserContext(), ServiceAuthContextKey{}, authInfo)
				c.SetUserContext(ctx)

				return c.Next()
			}
		}

		// No authentication method succeeded.
		return m.unauthorizedError(c, "authentication_required", "No valid authentication provided")
	}
}

// tryAuthenticate attempts authentication with the specified method.
func (m *ServiceAuthMiddleware) tryAuthenticate(c *fiber.Ctx, method AuthMethod) (*ServiceAuthInfo, error) {
	switch method {
	case AuthMethodJWT:
		return m.authenticateJWT(c)
	case AuthMethodMTLS:
		return m.authenticateMTLS(c)
	case AuthMethodAPIKey:
		return m.authenticateAPIKey(c)
	case AuthMethodClientCredentials:
		return m.authenticateClientCredentials(c)
	default:
		return nil, fmt.Errorf("unsupported auth method: %s", method)
	}
}

// authenticateJWT validates JWT bearer token.
func (m *ServiceAuthMiddleware) authenticateJWT(c *fiber.Ctx) (*ServiceAuthInfo, error) {
	if m.jwtValidator == nil {
		return nil, errors.New("JWT validator not configured")
	}

	// Extract token from Authorization header.
	authHeader := c.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, cryptoutilSharedMagic.HTTPAuthorizationBearerPrefix) {
		return nil, errors.New("no bearer token")
	}

	tokenString := strings.TrimPrefix(authHeader, cryptoutilSharedMagic.HTTPAuthorizationBearerPrefix)
	if tokenString == "" {
		return nil, errors.New("empty bearer token")
	}

	claims, err := m.jwtValidator.ValidateToken(c.Context(), tokenString)
	if err != nil {
		return nil, fmt.Errorf("JWT validation failed: %w", err)
	}

	return &ServiceAuthInfo{
		Method:      AuthMethodJWT,
		ServiceName: claims.PreferredUsername,
		Subject:     claims.Subject,
		Scopes:      claims.Scopes,
		Metadata:    claims.Custom,
	}, nil
}

// authenticateMTLS validates client certificate.
func (m *ServiceAuthMiddleware) authenticateMTLS(c *fiber.Ctx) (*ServiceAuthInfo, error) {
	if m.config.MTLSConfig == nil {
		return nil, errors.New("mTLS config not provided")
	}

	// Get TLS connection state.
	tlsState := c.Context().TLSConnectionState()
	if tlsState == nil {
		return nil, errors.New("no TLS connection")
	}

	// Check for peer certificates.
	if len(tlsState.PeerCertificates) == 0 {
		if m.config.MTLSConfig.RequireClientCert {
			return nil, errors.New("client certificate required")
		}

		return nil, errors.New("no client certificate")
	}

	// Get client certificate.
	clientCert := tlsState.PeerCertificates[0]

	// Validate CN if configured.
	if len(m.config.MTLSConfig.AllowedCNs) > 0 {
		if !m.isAllowedValue(clientCert.Subject.CommonName, m.config.MTLSConfig.AllowedCNs) {
			return nil, fmt.Errorf("CN %s not allowed", clientCert.Subject.CommonName)
		}
	}

	// Validate OU if configured.
	if len(m.config.MTLSConfig.AllowedOUs) > 0 {
		allowed := false

		for _, ou := range clientCert.Subject.OrganizationalUnit {
			if m.isAllowedValue(ou, m.config.MTLSConfig.AllowedOUs) {
				allowed = true

				break
			}
		}

		if !allowed {
			return nil, fmt.Errorf("OU not allowed")
		}
	}

	// Validate DNS SANs if configured.
	if len(m.config.MTLSConfig.AllowedDNSSANs) > 0 {
		allowed := false

		for _, san := range clientCert.DNSNames {
			if m.isAllowedValue(san, m.config.MTLSConfig.AllowedDNSSANs) {
				allowed = true

				break
			}
		}

		if !allowed {
			return nil, fmt.Errorf("DNS SAN not allowed")
		}
	}

	return &ServiceAuthInfo{
		Method:        AuthMethodMTLS,
		ServiceName:   clientCert.Subject.CommonName,
		CertificateCN: clientCert.Subject.CommonName,
		Metadata: map[string]any{
			"serial_number": clientCert.SerialNumber.String(),
			"not_after":     clientCert.NotAfter,
			"issuer":        clientCert.Issuer.CommonName,
		},
	}, nil
}

// authenticateAPIKey validates API key from header.
func (m *ServiceAuthMiddleware) authenticateAPIKey(c *fiber.Ctx) (*ServiceAuthInfo, error) {
	if m.config.APIKeyConfig == nil {
		return nil, errors.New("API key config not provided")
	}

	headerName := m.config.APIKeyConfig.HeaderName
	if headerName == "" {
		headerName = "X-API-Key"
	}

	apiKey := c.Get(headerName)
	if apiKey == "" {
		return nil, errors.New("no API key provided")
	}

	// Use dynamic validator if provided.
	if m.config.APIKeyConfig.KeyValidator != nil {
		serviceName, valid, err := m.config.APIKeyConfig.KeyValidator(c.Context(), apiKey)
		if err != nil {
			return nil, fmt.Errorf("API key validation failed: %w", err)
		}

		if !valid {
			return nil, errors.New("invalid API key")
		}

		return &ServiceAuthInfo{
			Method:      AuthMethodAPIKey,
			ServiceName: serviceName,
		}, nil
	}

	// Use static key map.
	if m.config.APIKeyConfig.ValidKeys == nil {
		return nil, errors.New("no valid keys configured")
	}

	// Constant-time comparison for all keys.
	for key, serviceName := range m.config.APIKeyConfig.ValidKeys {
		if subtle.ConstantTimeCompare([]byte(apiKey), []byte(key)) == 1 {
			return &ServiceAuthInfo{
				Method:      AuthMethodAPIKey,
				ServiceName: serviceName,
			}, nil
		}
	}

	return nil, errors.New("invalid API key")
}

// authenticateClientCredentials validates OAuth2 client credentials via introspection.
func (m *ServiceAuthMiddleware) authenticateClientCredentials(c *fiber.Ctx) (*ServiceAuthInfo, error) {
	if m.config.ClientCredentialsConfig == nil {
		return nil, errors.New("client credentials config not provided")
	}

	// For client credentials, we expect a bearer token from the token endpoint.
	// This token should be validated via introspection.
	authHeader := c.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, cryptoutilSharedMagic.HTTPAuthorizationBearerPrefix) {
		return nil, errors.New("no bearer token for client credentials")
	}

	// Token validation would use introspection endpoint.
	// For now, delegate to JWT validator if configured.
	if m.jwtValidator != nil {
		return m.authenticateJWT(c)
	}

	return nil, errors.New("client credentials authentication not fully implemented")
}

// isAllowedValue checks if a value is in the allowed list.
func (m *ServiceAuthMiddleware) isAllowedValue(value string, allowed []string) bool {
	for _, a := range allowed {
		if value == a {
			return true
		}
	}

	return false
}

// unauthorizedError returns 401 error response.
func (m *ServiceAuthMiddleware) unauthorizedError(c *fiber.Ctx, errorCode, message string) error {
	response := fiber.Map{"error": errorCode}

	if m.config.ErrorDetailLevel != errorDetailLevelMin {
		response["message"] = message
	}

	if err := c.Status(fiber.StatusUnauthorized).JSON(response); err != nil {
		return fmt.Errorf("failed to send unauthorized response: %w", err)
	}

	return nil
}

// GetServiceAuthInfo extracts service auth info from request context.
func GetServiceAuthInfo(ctx context.Context) *ServiceAuthInfo {
	if info, ok := ctx.Value(ServiceAuthContextKey{}).(*ServiceAuthInfo); ok {
		return info
	}

	return nil
}

// RequireServiceAuth middleware ensures service is authenticated.
func RequireServiceAuth(authMiddleware *ServiceAuthMiddleware) fiber.Handler {
	return authMiddleware.Middleware()
}

// ConfigureTLSForMTLS returns TLS config for mTLS client verification.
func ConfigureTLSForMTLS(requireClientCert bool) *tls.Config {
	clientAuth := tls.VerifyClientCertIfGiven
	if requireClientCert {
		clientAuth = tls.RequireAndVerifyClientCert
	}

	return &tls.Config{
		ClientAuth: clientAuth,
		MinVersion: tls.VersionTLS13,
	}
}
