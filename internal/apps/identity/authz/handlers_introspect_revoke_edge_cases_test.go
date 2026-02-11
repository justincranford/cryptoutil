// Copyright (c) 2025 Justin Cranford

package authz_test

import (
	"context"
	json "encoding/json"
	"fmt"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity/authz"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityMagic "cryptoutil/internal/apps/identity/magic"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

// TestHandleIntrospect_InvalidTokenFormat validates error for malformed token parameter.
func TestHandleIntrospect_InvalidTokenFormat(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testID := googleUuid.Must(googleUuid.NewV7()).String()

	// Clear migration state to ensure fresh database for this test.
	cryptoutilIdentityRepository.ResetMigrationStateForTesting()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type:        "sqlite",
		DSN:         fmt.Sprintf("file:test_%s.db?mode=memory&cache=shared", testID),
		AutoMigrate: true,
	}

	cfg := &cryptoutilIdentityConfig.Config{
		Database: dbConfig,
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			AccessTokenLifetime: 3600,
		},
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err, "Failed to create repository factory")

	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err, "Failed to run migrations")

	svc := cryptoutilIdentityAuthz.NewService(cfg, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	svc.RegisterRoutes(app)

	formBody := url.Values{
		cryptoutilIdentityMagic.ParamToken: []string{""},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/introspect", strings.NewReader(formBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should return 400 for empty token")
}

// TestHandleRevoke_InvalidTokenFormat validates error for malformed token parameter.
func TestHandleRevoke_InvalidTokenFormat(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testID := googleUuid.Must(googleUuid.NewV7()).String()

	// Clear migration state to ensure fresh database for this test.
	cryptoutilIdentityRepository.ResetMigrationStateForTesting()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type:        "sqlite",
		DSN:         fmt.Sprintf("file:test_%s.db?mode=memory&cache=shared", testID),
		AutoMigrate: true,
	}

	cfg := &cryptoutilIdentityConfig.Config{
		Database: dbConfig,
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			AccessTokenLifetime: 3600,
		},
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err, "Failed to create repository factory")

	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err, "Failed to run migrations")

	svc := cryptoutilIdentityAuthz.NewService(cfg, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	svc.RegisterRoutes(app)

	formBody := url.Values{
		cryptoutilIdentityMagic.ParamToken: []string{""},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/revoke", strings.NewReader(formBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should return 400 for empty token")
}

// TestHandleIntrospect_UnknownToken validates inactive response for non-existent token.
func TestHandleIntrospect_UnknownToken(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testID := googleUuid.Must(googleUuid.NewV7()).String()

	// Clear migration state to ensure fresh database for this test.
	cryptoutilIdentityRepository.ResetMigrationStateForTesting()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type:        "sqlite",
		DSN:         fmt.Sprintf("file:test_%s.db?mode=memory&cache=shared", testID),
		AutoMigrate: true,
	}

	cfg := &cryptoutilIdentityConfig.Config{
		Database: dbConfig,
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			AccessTokenLifetime: 3600,
		},
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err, "Failed to create repository factory")

	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err, "Failed to run migrations")

	svc := cryptoutilIdentityAuthz.NewService(cfg, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	svc.RegisterRoutes(app)

	unknownToken := googleUuid.Must(googleUuid.NewV7()).String()

	formBody := url.Values{
		cryptoutilIdentityMagic.ParamToken: []string{unknownToken},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/introspect", strings.NewReader(formBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusOK, resp.StatusCode, "Should return 200 OK")

	var body map[string]any

	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err, "Response body should be valid JSON")

	active, ok := body["active"].(bool)
	require.True(t, ok, "Response should contain active field")
	require.False(t, active, "Unknown token should be inactive")
}

// TestHandleRevoke_AlreadyRevokedToken validates success for already revoked token.
func TestHandleRevoke_AlreadyRevokedToken(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testID := googleUuid.Must(googleUuid.NewV7()).String()

	// Clear migration state to ensure fresh database for this test.
	cryptoutilIdentityRepository.ResetMigrationStateForTesting()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type:        "sqlite",
		DSN:         fmt.Sprintf("file:test_%s.db?mode=memory&cache=shared", testID),
		AutoMigrate: true,
	}

	cfg := &cryptoutilIdentityConfig.Config{
		Database: dbConfig,
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			AccessTokenLifetime: 3600,
		},
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err, "Failed to create repository factory")

	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err, "Failed to run migrations")

	userUUID := googleUuid.Must(googleUuid.NewV7())
	testUser := &cryptoutilIdentityDomain.User{
		ID:           userUUID,
		Sub:          fmt.Sprintf("user-%s", userUUID),
		PasswordHash: "dummy-hash",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	userRepo := repoFactory.UserRepository()
	err = userRepo.Create(ctx, testUser)
	require.NoError(t, err, "Failed to create test user")

	clientUUID := googleUuid.Must(googleUuid.NewV7())
	testClient := &cryptoutilIdentityDomain.Client{
		ID:                   clientUUID,
		ClientID:             fmt.Sprintf("test-client-%s", clientUUID),
		ClientType:           cryptoutilIdentityDomain.ClientTypeConfidential,
		Name:                 "Test Client",
		AllowedGrantTypes:    []string{cryptoutilIdentityMagic.GrantTypeRefreshToken},
		AllowedScopes:        []string{"test-scope"},
		RedirectURIs:         []string{"https://example.com/callback"},
		RequirePKCE:          boolPtr(false),
		AccessTokenLifetime:  3600,
		RefreshTokenLifetime: 86400,
		IDTokenLifetime:      3600,
		Enabled:              boolPtr(true),
		CreatedAt:            time.Now().UTC(),
		UpdatedAt:            time.Now().UTC(),
	}

	clientRepo := repoFactory.ClientRepository()
	err = clientRepo.Create(ctx, testClient)
	require.NoError(t, err, "Failed to create test client")

	tokenValue := googleUuid.Must(googleUuid.NewV7()).String()
	now := time.Now().UTC()
	revokedAt := now.Add(-1 * time.Hour)

	revokedToken := &cryptoutilIdentityDomain.Token{
		ID:          googleUuid.Must(googleUuid.NewV7()),
		TokenType:   "refresh",
		TokenFormat: "uuid",
		TokenValue:  tokenValue,
		ClientID:    clientUUID,
		UserID: cryptoutilIdentityDomain.NullableUUID{
			UUID:  userUUID,
			Valid: true,
		},
		ExpiresAt: now.Add(24 * time.Hour),
		IssuedAt:  now,
		RevokedAt: &revokedAt,
	}

	tokenRepo := repoFactory.TokenRepository()
	err = tokenRepo.Create(ctx, revokedToken)
	require.NoError(t, err, "Failed to create revoked token")

	svc := cryptoutilIdentityAuthz.NewService(cfg, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	svc.RegisterRoutes(app)

	formBody := url.Values{
		cryptoutilIdentityMagic.ParamToken: []string{tokenValue},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/revoke", strings.NewReader(formBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusOK, resp.StatusCode, "Should return 200 OK for already revoked token")
}
