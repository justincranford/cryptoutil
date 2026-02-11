// Copyright (c) 2025 Justin Cranford
//
//

package idp_test

import (
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityIdp "cryptoutil/internal/apps/identity/idp"
)

// TestRegisterMiddleware_NilConfig validates middleware registration with nil config.
func TestRegisterMiddleware_NilConfig(t *testing.T) {
	t.Parallel()

	// Create service with nil config (uses default CORS origins).
	service := cryptoutilIdentityIdp.NewService(nil, nil, nil)
	require.NotNil(t, service)

	app := fiber.New()

	// Should not panic with nil config - uses default CORS wildcard.
	service.RegisterMiddleware(app)

	require.NotNil(t, app)
}

// TestRegisterMiddleware_EmptyCORSOrigins validates middleware registration with empty CORS origins.
func TestRegisterMiddleware_EmptyCORSOrigins(t *testing.T) {
	t.Parallel()

	config := &cryptoutilIdentityConfig.Config{
		Security: &cryptoutilIdentityConfig.SecurityConfig{
			CORSAllowedOrigins: []string{}, // Empty slice, not nil
		},
	}

	service := cryptoutilIdentityIdp.NewService(config, nil, nil)
	require.NotNil(t, service)

	app := fiber.New()

	// Should not panic with empty CORS origins - uses default wildcard.
	service.RegisterMiddleware(app)

	require.NotNil(t, app)
}

// TestRegisterMiddleware_ValidCORSOrigins validates middleware registration with configured CORS origins.
func TestRegisterMiddleware_ValidCORSOrigins(t *testing.T) {
	t.Parallel()

	config := &cryptoutilIdentityConfig.Config{
		Security: &cryptoutilIdentityConfig.SecurityConfig{
			CORSAllowedOrigins: []string{"https://example.com", "https://app.example.com"},
		},
	}

	service := cryptoutilIdentityIdp.NewService(config, nil, nil)
	require.NotNil(t, service)

	app := fiber.New()

	// Should not panic with configured CORS origins.
	service.RegisterMiddleware(app)

	require.NotNil(t, app)
}

// TestRegisterMiddleware_RateLimitConfig validates middleware registration with rate limiting disabled.
func TestRegisterMiddleware_RateLimitDisabled(t *testing.T) {
	t.Parallel()

	config := &cryptoutilIdentityConfig.Config{
		Security: &cryptoutilIdentityConfig.SecurityConfig{
			RateLimitEnabled: false,
		},
	}

	service := cryptoutilIdentityIdp.NewService(config, nil, nil)
	require.NotNil(t, service)

	app := fiber.New()

	// Should not panic with rate limiting disabled.
	service.RegisterMiddleware(app)

	require.NotNil(t, app)
}

// TestRegisterMiddleware_RateLimitEnabled validates middleware registration with rate limiting enabled.
func TestRegisterMiddleware_RateLimitEnabled(t *testing.T) {
	t.Parallel()

	config := &cryptoutilIdentityConfig.Config{
		Security: &cryptoutilIdentityConfig.SecurityConfig{
			RateLimitEnabled:  true,
			RateLimitRequests: 100,
		},
	}

	service := cryptoutilIdentityIdp.NewService(config, nil, nil)
	require.NotNil(t, service)

	app := fiber.New()

	// Should not panic with rate limiting enabled.
	service.RegisterMiddleware(app)

	require.NotNil(t, app)
}
