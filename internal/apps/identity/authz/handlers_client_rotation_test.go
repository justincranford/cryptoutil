// Copyright (c) 2025 Justin Cranford
//
//

package authz

import (
	"context"
	json "encoding/json"
	"fmt"
	http "net/http"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityClientAuth "cryptoutil/internal/apps/identity/authz/clientauth"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedCryptoHash "cryptoutil/internal/shared/crypto/hash"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestClientSecretRotation_EndToEnd validates complete rotation flow.
func TestClientSecretRotation_EndToEnd(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Setup database and service.
	dbCfg := &cryptoutilIdentityConfig.DatabaseConfig{
		Type:        "sqlite",
		DSN:         ":memory:",
		AutoMigrate: true,
	}
	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbCfg)
	require.NoError(t, err)

	// Run database migrations.
	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err)

	appCfg := &cryptoutilIdentityConfig.Config{
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			AccessTokenLifetime:  3600,
			RefreshTokenLifetime: 86400,
			IDTokenLifetime:      3600,
		},
	}
	tokenSvc := cryptoutilIdentityIssuer.NewTokenService(nil, nil, nil, appCfg.Tokens)
	service := NewService(appCfg, repoFactory, tokenSvc)

	app := fiber.New()
	service.RegisterRoutes(app)

	// CRITICAL: Generate plaintext secret FIRST, then hash it.
	// Client.Create() will use the provided hash to create ClientSecretVersion.
	// Create test client with known secret.
	originalSecret := "original-secret-" + googleUuid.Must(googleUuid.NewV7()).String()
	hashedSecret, err := cryptoutilSharedCryptoHash.HashLowEntropyNonDeterministic(originalSecret)
	require.NoError(t, err)

	client := &cryptoutilIdentityDomain.Client{
		ClientID:                "test-client-" + googleUuid.Must(googleUuid.NewV7()).String(),
		ClientSecret:            hashedSecret,
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		Name:                    "Test Client Rotation",
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretBasic,
	}

	// Create client - this will create ClientSecretVersion (version 1) using provided hash.
	err = repoFactory.ClientRepository().Create(ctx, client)
	require.NoError(t, err)

	// Fetch created client to get updated state after ClientSecretVersion creation.
	createdClient, err := repoFactory.ClientRepository().GetByClientID(ctx, client.ClientID)
	require.NoError(t, err)

	// Verify original secret works before rotation.
	match, err := cryptoutilIdentityClientAuth.CompareSecret(createdClient.ClientSecret, originalSecret)
	require.NoError(t, err)
	require.True(t, match)

	// Rotate the secret via HTTP endpoint.
	// Note: Use client.ClientID (string) instead of client.ID (UUID) for URL parameter.
	reqURL := fmt.Sprintf("/oauth2/v1/clients/%s/rotate-secret", client.ClientID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, nil)
	require.NoError(t, err)
	req.SetBasicAuth(client.ClientID, originalSecret)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() {
		//nolint:errcheck // Test cleanup - error intentionally ignored
		_ = resp.Body.Close()
	}()

	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Parse new secret from response.
	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	newSecret, ok := result["client_secret"].(string)
	require.True(t, ok)
	require.NotEmpty(t, newSecret)

	// Wait briefly for database update to propagate.
	time.Sleep(cryptoutilSharedMagic.DatabasePropagationDelay)

	// Retrieve updated client from database.
	updatedClient, err := repoFactory.ClientRepository().GetByID(ctx, client.ID)
	require.NoError(t, err)

	// Verify old secret NO LONGER works.
	match, err = cryptoutilIdentityClientAuth.CompareSecret(updatedClient.ClientSecret, originalSecret)
	require.NoError(t, err)
	require.False(t, match, "Old secret should not work after rotation")

	// Verify new secret DOES work.
	match, err = cryptoutilIdentityClientAuth.CompareSecret(updatedClient.ClientSecret, newSecret)
	require.NoError(t, err)
	require.True(t, match, "New secret should work after rotation")
}

// TestClientSecretRotation_InvalidClientID validates error when authentication missing.
func TestClientSecretRotation_InvalidClientID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Setup database and service.
	dbCfg := &cryptoutilIdentityConfig.DatabaseConfig{
		Type:        "sqlite",
		DSN:         ":memory:",
		AutoMigrate: true,
	}
	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbCfg)
	require.NoError(t, err)

	// Run database migrations.
	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err)

	appCfg := &cryptoutilIdentityConfig.Config{
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			AccessTokenLifetime:  3600,
			RefreshTokenLifetime: 86400,
			IDTokenLifetime:      3600,
		},
	}
	tokenSvc := cryptoutilIdentityIssuer.NewTokenService(nil, nil, nil, appCfg.Tokens)
	service := NewService(appCfg, repoFactory, tokenSvc)

	app := fiber.New()
	service.RegisterRoutes(app)

	// Attempt to rotate without authentication.
	reqURL := "/oauth2/v1/clients/some-client-id/rotate-secret"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, nil)
	require.NoError(t, err)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() {
		//nolint:errcheck // Test cleanup - error intentionally ignored
		_ = resp.Body.Close()
	}()

	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.Equal(t, cryptoutilSharedMagic.ErrorInvalidClient, result["error"])
	require.Contains(t, result["error_description"], "Client authentication failed")
}

// TestClientSecretRotation_ClientNotFound validates error handling for non-existent client.
func TestClientSecretRotation_ClientNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Setup service.
	dbCfg := &cryptoutilIdentityConfig.DatabaseConfig{
		Type:        "sqlite",
		DSN:         ":memory:",
		AutoMigrate: true,
	}
	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbCfg)
	require.NoError(t, err)

	// Run database migrations.
	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err)

	appCfg := &cryptoutilIdentityConfig.Config{
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			AccessTokenLifetime:  3600,
			RefreshTokenLifetime: 86400,
			IDTokenLifetime:      3600,
		},
	}
	tokenSvc := cryptoutilIdentityIssuer.NewTokenService(nil, nil, nil, appCfg.Tokens)
	service := NewService(appCfg, repoFactory, tokenSvc)

	app := fiber.New()
	service.RegisterRoutes(app)

	// Create auth client for authentication.
	authSecret := "auth-secret-" + googleUuid.Must(googleUuid.NewV7()).String()
	authHashed, err := cryptoutilSharedCryptoHash.HashLowEntropyNonDeterministic(authSecret)
	require.NoError(t, err)

	authClient := &cryptoutilIdentityDomain.Client{
		ClientID:                "auth-client-" + googleUuid.Must(googleUuid.NewV7()).String(),
		ClientSecret:            authHashed,
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		Name:                    "Auth Client",
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretBasic,
	}

	err = repoFactory.ClientRepository().Create(ctx, authClient)
	require.NoError(t, err)

	// Attempt to rotate non-existent client.
	nonExistentID := googleUuid.Must(googleUuid.NewV7())
	reqURL := fmt.Sprintf("/oauth2/v1/clients/%s/rotate-secret", nonExistentID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, nil)
	require.NoError(t, err)
	req.SetBasicAuth(authClient.ClientID, authSecret)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() {
		//nolint:errcheck // Test cleanup - error intentionally ignored
		_ = resp.Body.Close()
	}()

	// Should fail with forbidden (client can only rotate own secret).
	require.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.Equal(t, cryptoutilSharedMagic.ErrorAccessDenied, result["error"])
}
