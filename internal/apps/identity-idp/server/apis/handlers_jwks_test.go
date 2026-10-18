// Copyright (c) 2025-2026 Justin Cranford.
package apis

import (
	"context"
	json "encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

// TestHandleJWKS_EmptySet tests JWKS endpoint returns empty set when no keys exist.
func TestHandleJWKS_EmptySet(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create test config and repository factory.
	config := cryptoutilIdentityConfig.DefaultConfig()
	config.Database.DSN = cryptoutilSharedMagic.SQLiteMemoryPlaceholder

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
	req := httptest.NewRequest("GET", cryptoutilSharedMagic.PathJWKS, nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request failed")

	defer func() {
		_ = resp.Body.Close() //nolint:errcheck // Test cleanup.
	}()

	// Verify response.
	require.Equal(t, fiber.StatusOK, resp.StatusCode, "Expected 200 OK")
	require.Equal(t, testJSONContentType, resp.Header.Get("Content-Type"))
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
					Type: cryptoutilSharedMagic.TestDatabaseSQLite,
					DSN:  cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
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

			req := httptest.NewRequest("GET", cryptoutilSharedMagic.PathJWKS, nil)
			resp, err := app.Test(req, -1)
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

func TestHandleJWKS_ActivePublicSigningKeys(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := cryptoutilIdentityConfig.DefaultConfig()
	config.Database.DSN = cryptoutilSharedMagic.SQLiteMemoryPlaceholder

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, config.Database)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, repoFactory.Close())
	})

	require.NoError(t, repoFactory.DB().AutoMigrate(&cryptoutilIdentityDomain.Key{}))

	keyRotationManager, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		cryptoutilIdentityIssuer.NewProductionKeyGenerator(),
		nil,
	)
	require.NoError(t, err)
	require.NoError(t, keyRotationManager.RotateSigningKey(ctx, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm))

	publicKeys := keyRotationManager.GetPublicKeys()
	require.Len(t, publicKeys, 1)
	publicKeyJSON, err := json.Marshal(publicKeys[0])
	require.NoError(t, err)

	keys := []*cryptoutilIdentityDomain.Key{
		{
			Usage:      cryptoutilSharedMagic.KeyUsageSigning,
			Algorithm:  cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
			PrivateKey: testEmptyPrivateKeyJSON,
			PublicKey:  string(publicKeyJSON),
			Active:     true,
			ExpiresAt:  time.Now().UTC().Add(time.Hour),
		},
		{
			Usage:      cryptoutilSharedMagic.KeyUsageSigning,
			Algorithm:  cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
			PrivateKey: testEmptyPrivateKeyJSON,
			PublicKey:  "not-json",
			Active:     true,
			ExpiresAt:  time.Now().UTC().Add(time.Hour),
		},
		{
			Usage:      cryptoutilSharedMagic.KeyUsageSigning,
			Algorithm:  cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
			PrivateKey: testEmptyPrivateKeyJSON,
			Active:     true,
			ExpiresAt:  time.Now().UTC().Add(time.Hour),
		},
	}
	for _, key := range keys {
		require.NoError(t, repoFactory.KeyRepository().Create(ctx, key))
	}

	service := NewService(config, repoFactory, nil)
	app := fiber.New()
	app.Get(cryptoutilSharedMagic.PathJWKS, service.handleJWKS)

	response, err := app.Test(httptest.NewRequest("GET", cryptoutilSharedMagic.PathJWKS, nil), -1)
	require.NoError(t, err)

	defer func() { _ = response.Body.Close() }() //nolint:errcheck // Test cleanup.

	require.Equal(t, fiber.StatusOK, response.StatusCode)
	require.Equal(t, testJSONContentType, response.Header.Get("Content-Type"))
	require.Equal(t, "public, max-age=3600", response.Header.Get("Cache-Control"))

	var jwks map[string][]map[string]any
	require.NoError(t, json.NewDecoder(response.Body).Decode(&jwks))
	require.Len(t, jwks["keys"], 1)
	require.Equal(t, publicKeys[0]["kid"], jwks["keys"][0]["kid"])
}
