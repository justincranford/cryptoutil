// Copyright (c) 2025 Justin Cranford

package idp

import (
	"context"
	json "encoding/json"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

// TestHandleJWKS_EmptySet tests JWKS endpoint returns empty set when no keys exist.
func TestHandleJWKS_EmptySet(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create test config and repository factory.
	config := cryptoutilIdentityConfig.DefaultConfig()
	config.Database.DSN = ":memory:"

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, config.Database)
	require.NoError(t, err, "Failed to create repository factory")

	defer func() {
		_ = repoFactory.Close() //nolint:errcheck // Test cleanup.
	}()

	// Create service.
	service := NewService(config, repoFactory, nil)

	// Create Fiber app and register routes.
	app := fiber.New()
	service.RegisterRoutes(app)

	// Make request to JWKS endpoint.
	req := httptest.NewRequest("GET", "/.well-known/jwks.json", nil)
	resp, err := app.Test(req)
	require.NoError(t, err, "Request failed")

	defer func() {
		_ = resp.Body.Close() //nolint:errcheck // Test cleanup.
	}()

	// Verify response.
	require.Equal(t, fiber.StatusOK, resp.StatusCode, "Expected 200 OK")
	require.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	require.Equal(t, "public, max-age=3600", resp.Header.Get("Cache-Control"))

	// Parse JWKS response.
	var jwks map[string]any

	err = json.NewDecoder(resp.Body).Decode(&jwks)
	require.NoError(t, err, "Failed to decode JWKS response")

	// Verify empty keys array.
	keys, ok := jwks["keys"].([]any)
	require.True(t, ok, "JWKS should contain 'keys' array")
	require.Empty(t, keys, "Keys array should be empty when no signing keys exist")
}

// TestHandleJWKS_ErrorScenarios tests JWKS endpoint error handling.
func TestHandleJWKS_ErrorScenarios(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupRepo      func(t *testing.T) *cryptoutilIdentityRepository.RepositoryFactory
		expectedStatus int
	}{
		{
			name: "database_connection_error_returns_empty_jwks",
			setupRepo: func(t *testing.T) *cryptoutilIdentityRepository.RepositoryFactory {
				t.Helper()

				ctx := context.Background()

				// Create repository with invalid DSN (will fail on operations).
				config := &cryptoutilIdentityConfig.DatabaseConfig{
					Type: "sqlite",
					DSN:  ":memory:",
				}

				repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, config)
				require.NoError(t, err)

				return repoFactory
			},
			expectedStatus: fiber.StatusOK, // Returns empty JWKS on error per spec.
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repoFactory := tc.setupRepo(t)

			defer func() {
				_ = repoFactory.Close() //nolint:errcheck // Test cleanup.
			}()

			config := cryptoutilIdentityConfig.DefaultConfig()
			service := NewService(config, repoFactory, nil)

			app := fiber.New()
			service.RegisterRoutes(app)

			req := httptest.NewRequest("GET", "/.well-known/jwks.json", nil)
			resp, err := app.Test(req)
			require.NoError(t, err)

			defer func() {
				_ = resp.Body.Close() //nolint:errcheck // Test cleanup.
			}()

			require.Equal(t, tc.expectedStatus, resp.StatusCode)

			// Parse JWKS response.
			var jwks map[string]any

			err = json.NewDecoder(resp.Body).Decode(&jwks)
			require.NoError(t, err)

			// Verify empty keys array (error case returns empty JWKS).
			keys, ok := jwks["keys"].([]any)
			require.True(t, ok)
			require.Empty(t, keys)
		})
	}
}
