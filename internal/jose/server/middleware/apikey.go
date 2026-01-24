// Copyright (c) 2025 Justin Cranford

// Package middleware provides HTTP middleware for the JOSE Authority Server.
package middleware

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"

	fiber "github.com/gofiber/fiber/v2"
)

// APIKeyConfig configures API key authentication middleware.
type APIKeyConfig struct {
	// HeaderName is the header containing the API key (default: X-API-Key).
	HeaderName string

	// QueryParam is the optional query parameter containing the API key.
	QueryParam string

	// ValidKeys is a map of API key to client name/metadata.
	ValidKeys map[string]string

	// KeyValidator is a function to validate API keys dynamically.
	// If provided, takes precedence over ValidKeys.
	KeyValidator func(ctx context.Context, apiKey string) (clientName string, valid bool, err error)

	// Skipper returns true to skip authentication for specific routes.
	Skipper func(c *fiber.Ctx) bool

	// ErrorDetailLevel controls error verbosity ("minimal", "basic", "detailed").
	ErrorDetailLevel string
}

// DefaultAPIKeyHeader is the default header for API keys.
const DefaultAPIKeyHeader = "X-API-Key"

// apiKeyMaskMinLength is the minimum length for masking API keys.
const apiKeyMaskMinLength = 8

// DefaultAPIKeyConfig returns default API key configuration.
func DefaultAPIKeyConfig() *APIKeyConfig {
	return &APIKeyConfig{
		HeaderName:       DefaultAPIKeyHeader,
		QueryParam:       "",
		ValidKeys:        make(map[string]string),
		KeyValidator:     nil,
		Skipper:          nil,
		ErrorDetailLevel: "basic",
	}
}

// APIKeyMiddleware provides API key authentication.
type APIKeyMiddleware struct {
	config *APIKeyConfig
}

// NewAPIKeyMiddleware creates a new API key middleware.
func NewAPIKeyMiddleware(config *APIKeyConfig) *APIKeyMiddleware {
	if config == nil {
		config = DefaultAPIKeyConfig()
	}

	if config.HeaderName == "" {
		config.HeaderName = DefaultAPIKeyHeader
	}

	return &APIKeyMiddleware{
		config: config,
	}
}

// APIKeyContextKey is the context key for API key client info.
type APIKeyContextKey struct{}

// APIKeyInfo contains authenticated API key information.
type APIKeyInfo struct {
	// ClientName is the name/identifier of the authenticated client.
	ClientName string

	// APIKey is the API key used (partially masked for logging).
	APIKeyMasked string
}

// Handler returns the Fiber middleware handler.
func (m *APIKeyMiddleware) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if we should skip authentication.
		if m.config.Skipper != nil && m.config.Skipper(c) {
			return c.Next()
		}

		// Try to get API key from header.
		apiKey := c.Get(m.config.HeaderName)

		// Try query parameter as fallback.
		if apiKey == "" && m.config.QueryParam != "" {
			apiKey = c.Query(m.config.QueryParam)
		}

		// No API key provided.
		if apiKey == "" {
			return m.errorResponse(c, fiber.StatusUnauthorized, "unauthorized", "API key required")
		}

		// Validate API key.
		clientName, valid, err := m.validateKey(c.Context(), apiKey)
		if err != nil {
			return m.errorResponse(c, fiber.StatusInternalServerError, "internal_error", "failed to validate API key")
		}

		if !valid {
			return m.errorResponse(c, fiber.StatusUnauthorized, "unauthorized", "invalid API key")
		}

		// Store client info in context.
		c.Locals(APIKeyContextKey{}, &APIKeyInfo{
			ClientName:   clientName,
			APIKeyMasked: maskAPIKey(apiKey),
		})

		return c.Next()
	}
}

// validateKey validates the API key using dynamic validator or static map.
func (m *APIKeyMiddleware) validateKey(ctx context.Context, apiKey string) (string, bool, error) {
	// Try dynamic validator first.
	if m.config.KeyValidator != nil {
		return m.config.KeyValidator(ctx, apiKey)
	}

	// Fall back to static key map.
	if clientName, ok := m.config.ValidKeys[apiKey]; ok {
		return clientName, true, nil
	}

	return "", false, nil
}

// errorResponse sends a JSON error response.
func (m *APIKeyMiddleware) errorResponse(c *fiber.Ctx, status int, errorCode, message string) error {
	if err := c.Status(status).JSON(fiber.Map{
		"error":   errorCode,
		"message": message,
	}); err != nil {
		return fmt.Errorf("failed to send error response: %w", err)
	}

	return nil
}

// maskAPIKey masks an API key for safe logging.
func maskAPIKey(apiKey string) string {
	if len(apiKey) <= apiKeyMaskMinLength {
		return "****"
	}

	return apiKey[:4] + "****" + apiKey[len(apiKey)-4:]
}

// GetAPIKeyInfo retrieves API key info from context.
func GetAPIKeyInfo(c *fiber.Ctx) *APIKeyInfo {
	if info, ok := c.Locals(APIKeyContextKey{}).(*APIKeyInfo); ok {
		return info
	}

	return nil
}

// RequireAPIKey creates middleware that requires a valid API key.
func RequireAPIKey(validKeys map[string]string) fiber.Handler {
	config := DefaultAPIKeyConfig()
	config.ValidKeys = validKeys

	mw := NewAPIKeyMiddleware(config)

	return mw.Handler()
}

// RequireAPIKeyWithConfig creates middleware with custom configuration.
func RequireAPIKeyWithConfig(config *APIKeyConfig) fiber.Handler {
	mw := NewAPIKeyMiddleware(config)

	return mw.Handler()
}

// SecureCompare performs a constant-time comparison to prevent timing attacks.
func SecureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// APIKeyStore is an interface for integrating with dynamic key storage.
// It validates an API key against a key store.
type APIKeyStore interface {
	GetClientByAPIKey(ctx context.Context, apiKey string) (clientName string, found bool, err error)
}

// NewAPIKeyValidatorFromStore creates a key validator from an API key store.
func NewAPIKeyValidatorFromStore(store APIKeyStore) func(ctx context.Context, apiKey string) (string, bool, error) {
	return func(ctx context.Context, apiKey string) (string, bool, error) {
		if store == nil {
			return "", false, errors.New("API key store not configured")
		}

		return store.GetClientByAPIKey(ctx, apiKey)
	}
}
