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
	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilIdentityMagic "cryptoutil/internal/apps/identity/magic"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

// TestHandleRefreshTokenGrant_Success validates successful refresh token exchange.
func TestHandleRefreshTokenGrant_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testID := googleUuid.Must(googleUuid.NewV7()).String()

	// Clear migration state to ensure fresh database for this test.
	cryptoutilIdentityRepository.ResetMigrationStateForTesting()

	cfg := &cryptoutilIdentityConfig.Config{
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type: "sqlite",
			DSN:  fmt.Sprintf("file:test_%s.db?mode=memory&cache=shared", testID),
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer:              "https://localhost:8080",
			SigningAlgorithm:    "RS256",
			AccessTokenFormat:   "jws",
			AccessTokenLifetime: 3600,
		},
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, cfg.Database)
	require.NoError(t, err, "Failed to create repository factory")

	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err, "Failed to run migrations")

	// Create key rotation manager for token issuers.
	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		cryptoutilIdentityIssuer.NewProductionKeyGenerator(),
		nil,
	)
	require.NoError(t, err, "Failed to create key rotation manager")

	// Generate initial signing key.
	err = keyRotationMgr.RotateSigningKey(ctx, cfg.Tokens.SigningAlgorithm)
	require.NoError(t, err, "Failed to rotate initial signing key")

	// Create JWS issuer for access tokens.
	jwsIssuer, err := cryptoutilIdentityIssuer.NewJWSIssuer(
		cfg.Tokens.Issuer,
		keyRotationMgr,
		cfg.Tokens.SigningAlgorithm,
		time.Duration(cfg.Tokens.AccessTokenLifetime)*time.Second,
		time.Duration(cfg.Tokens.AccessTokenLifetime)*time.Second,
	)
	require.NoError(t, err, "Failed to create JWS issuer")

	uuidIssuer := cryptoutilIdentityIssuer.NewUUIDIssuer()
	tokenSvc := cryptoutilIdentityIssuer.NewTokenService(jwsIssuer, nil, uuidIssuer, cfg.Tokens)

	// Create test client for foreign key constraint
	clientUUID := googleUuid.Must(googleUuid.NewV7())

	testClient := &cryptoutilIdentityDomain.Client{
		ID:                      clientUUID,
		ClientID:                fmt.Sprintf("test-client-%s", clientUUID.String()),
		Name:                    "Test Client Refresh",
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		AllowedGrantTypes:       []string{cryptoutilIdentityMagic.GrantTypeRefreshToken},
		AllowedScopes:           []string{"openid"},
		RedirectURIs:            []string{"https://example.com/callback"},
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
		AccessTokenLifetime:     3600,
		RefreshTokenLifetime:    86400,
		IDTokenLifetime:         3600,
	}

	clientRepo := repoFactory.ClientRepository()
	err = clientRepo.Create(ctx, testClient)
	require.NoError(t, err, "Failed to create test client")

	// Create valid refresh token in database
	userUUID := googleUuid.Must(googleUuid.NewV7())
	tokenValue := googleUuid.NewString()

	// User must exist for foreign key constraint
	testUser := &cryptoutilIdentityDomain.User{
		ID:           userUUID,
		Sub:          fmt.Sprintf("user-%s", userUUID.String()),
		PasswordHash: "dummy-hash",
	}

	userRepo := repoFactory.UserRepository()
	err = userRepo.Create(ctx, testUser)
	require.NoError(t, err, "Failed to create test user")

	refreshToken := &cryptoutilIdentityDomain.Token{
		ID:          googleUuid.Must(googleUuid.NewV7()),
		TokenType:   cryptoutilIdentityDomain.TokenTypeRefresh,
		TokenFormat: cryptoutilIdentityDomain.TokenFormatUUID,
		TokenValue:  tokenValue,
		ClientID:    clientUUID,
		UserID:      cryptoutilIdentityDomain.NullableUUID{UUID: userUUID, Valid: true},
		ExpiresAt:   time.Now().UTC().Add(24 * time.Hour),
		IssuedAt:    time.Now().UTC(),
	}

	tokenRepo := repoFactory.TokenRepository()
	err = tokenRepo.Create(ctx, refreshToken)
	require.NoError(t, err, "Failed to create refresh token")

	svc := cryptoutilIdentityAuthz.NewService(cfg, repoFactory, tokenSvc)
	require.NotNil(t, svc, "Service should not be nil")

	err = svc.Start(context.Background())
	require.NoError(t, err, "Service start should succeed")

	defer func() {
		err := svc.Stop(context.Background())
		require.NoError(t, err, "Service stop should succeed")
	}()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	svc.RegisterRoutes(app)

	formBody := url.Values{
		cryptoutilIdentityMagic.ParamGrantType:    []string{cryptoutilIdentityMagic.GrantTypeRefreshToken},
		cryptoutilIdentityMagic.ParamRefreshToken: []string{tokenValue},
		cryptoutilIdentityMagic.ParamClientID:     []string{testClient.ClientID},
		cryptoutilIdentityMagic.ParamScope:        []string{"openid"},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/token", strings.NewReader(formBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusOK, resp.StatusCode, "Refresh token grant should succeed")

	var body map[string]any

	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err, "Response body should be valid JSON")

	accessToken, ok := body["access_token"].(string)
	require.True(t, ok, "Response should contain access_token")
	require.NotEmpty(t, accessToken, "Access token should not be empty")

	tokenType, ok := body["token_type"].(string)
	require.True(t, ok, "Response should contain token_type")
	require.Equal(t, "Bearer", tokenType, "Token type should be Bearer")
}

// TestHandleRefreshTokenGrant_MissingRefreshTokenParam validates error for missing refresh_token.
func TestHandleRefreshTokenGrant_MissingRefreshTokenParam(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testID := googleUuid.Must(googleUuid.NewV7()).String()

	// Clear migration state to ensure fresh database for this test.
	cryptoutilIdentityRepository.ResetMigrationStateForTesting()

	cfg := &cryptoutilIdentityConfig.Config{
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type: "sqlite",
			DSN:  fmt.Sprintf("file:test_%s.db?mode=memory&cache=shared", testID),
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			AccessTokenLifetime: 3600,
		},
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, cfg.Database)
	require.NoError(t, err, "Failed to create repository factory")

	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err, "Failed to run migrations")

	svc := cryptoutilIdentityAuthz.NewService(cfg, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	err = svc.Start(context.Background())
	require.NoError(t, err, "Service start should succeed")

	defer func() {
		err := svc.Stop(context.Background())
		require.NoError(t, err, "Service stop should succeed")
	}()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	svc.RegisterRoutes(app)

	formBody := url.Values{
		cryptoutilIdentityMagic.ParamGrantType: []string{cryptoutilIdentityMagic.GrantTypeRefreshToken},
		cryptoutilIdentityMagic.ParamClientID:  []string{"test-client"},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/token", strings.NewReader(formBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should return 400 for missing refresh_token parameter")
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should return 400 for missing refresh_token")

	var body map[string]any

	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err, "Response body should be valid JSON")

	errorCode, ok := body["error"].(string)
	require.True(t, ok, "Response should contain error field")
	require.Equal(t, cryptoutilIdentityMagic.ErrorInvalidRequest, errorCode, "Error should be invalid_request")
}

// TestHandleRefreshTokenGrant_InvalidRefreshToken validates error for non-existent refresh token.
func TestHandleRefreshTokenGrant_InvalidRefreshToken(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testID := googleUuid.Must(googleUuid.NewV7()).String()

	// Clear migration state to ensure fresh database for this test.
	cryptoutilIdentityRepository.ResetMigrationStateForTesting()

	cfg := &cryptoutilIdentityConfig.Config{
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type: "sqlite",
			DSN:  fmt.Sprintf("file:test_%s.db?mode=memory&cache=shared", testID),
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			AccessTokenLifetime: 3600,
		},
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, cfg.Database)
	require.NoError(t, err, "Failed to create repository factory")

	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err, "Failed to run migrations")

	svc := cryptoutilIdentityAuthz.NewService(cfg, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	err = svc.Start(context.Background())
	require.NoError(t, err, "Service start should succeed")

	defer func() {
		err := svc.Stop(context.Background())
		require.NoError(t, err, "Service stop should succeed")
	}()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	svc.RegisterRoutes(app)

	formBody := url.Values{
		cryptoutilIdentityMagic.ParamGrantType:    []string{cryptoutilIdentityMagic.GrantTypeRefreshToken},
		cryptoutilIdentityMagic.ParamRefreshToken: []string{"invalid-refresh-token"},
		cryptoutilIdentityMagic.ParamClientID:     []string{"test-client"},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/token", strings.NewReader(formBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	// Token lookup fails (404 Not Found) because token doesn't exist
	require.Contains(t, []int{fiber.StatusBadRequest, fiber.StatusNotFound}, resp.StatusCode, "Should return 400 or 404 for invalid refresh token")

	var body map[string]any

	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err, "Response body should be valid JSON")

	errorCode, ok := body["error"].(string)
	require.True(t, ok, "Response should contain error field")
	require.Equal(t, cryptoutilIdentityMagic.ErrorInvalidGrant, errorCode, "Error should be invalid_grant")
}

// TestHandleRefreshTokenGrant_RevokedToken validates error for revoked refresh token.
func TestHandleRefreshTokenGrant_RevokedToken(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testID := googleUuid.Must(googleUuid.NewV7()).String()

	// Clear migration state to ensure fresh database for this test.
	cryptoutilIdentityRepository.ResetMigrationStateForTesting()

	cfg := &cryptoutilIdentityConfig.Config{
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type: "sqlite",
			DSN:  fmt.Sprintf("file:test_%s.db?mode=memory&cache=shared", testID),
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			AccessTokenLifetime: 3600,
		},
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, cfg.Database)
	require.NoError(t, err, "Failed to create repository factory")

	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err, "Failed to run migrations")

	// Create test client for foreign key constraint
	clientUUID := googleUuid.Must(googleUuid.NewV7())

	testClient := &cryptoutilIdentityDomain.Client{
		ID:                      clientUUID,
		ClientID:                fmt.Sprintf("test-client-%s", clientUUID.String()),
		Name:                    "Test Client Revoked",
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		AllowedGrantTypes:       []string{cryptoutilIdentityMagic.GrantTypeRefreshToken},
		AllowedScopes:           []string{"openid"},
		RedirectURIs:            []string{"https://example.com/callback"},
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
		AccessTokenLifetime:     3600,
		RefreshTokenLifetime:    86400,
		IDTokenLifetime:         3600,
	}

	clientRepo := repoFactory.ClientRepository()
	err = clientRepo.Create(ctx, testClient)
	require.NoError(t, err, "Failed to create test client")

	// Create revoked refresh token
	userUUID := googleUuid.Must(googleUuid.NewV7())
	tokenValue := googleUuid.NewString()
	now := time.Now().UTC()

	// User must exist for foreign key constraint
	testUser := &cryptoutilIdentityDomain.User{
		ID:           userUUID,
		Sub:          fmt.Sprintf("user-%s", userUUID.String()),
		PasswordHash: "dummy-hash",
	}

	userRepo := repoFactory.UserRepository()
	err = userRepo.Create(ctx, testUser)
	require.NoError(t, err, "Failed to create test user")

	refreshToken := &cryptoutilIdentityDomain.Token{
		ID:          googleUuid.Must(googleUuid.NewV7()),
		TokenType:   cryptoutilIdentityDomain.TokenTypeRefresh,
		TokenFormat: cryptoutilIdentityDomain.TokenFormatUUID,
		TokenValue:  tokenValue,
		ClientID:    clientUUID,
		UserID:      cryptoutilIdentityDomain.NullableUUID{UUID: userUUID, Valid: true},
		ExpiresAt:   time.Now().UTC().Add(24 * time.Hour),
		RevokedAt:   &now,
		IssuedAt:    time.Now().UTC(),
	}

	tokenRepo := repoFactory.TokenRepository()
	err = tokenRepo.Create(ctx, refreshToken)
	require.NoError(t, err, "Failed to create revoked refresh token")

	svc := cryptoutilIdentityAuthz.NewService(cfg, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	err = svc.Start(context.Background())
	require.NoError(t, err, "Service start should succeed")

	defer func() {
		err := svc.Stop(context.Background())
		require.NoError(t, err, "Service stop should succeed")
	}()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	svc.RegisterRoutes(app)

	formBody := url.Values{
		cryptoutilIdentityMagic.ParamGrantType:    []string{cryptoutilIdentityMagic.GrantTypeRefreshToken},
		cryptoutilIdentityMagic.ParamRefreshToken: []string{tokenValue},
		cryptoutilIdentityMagic.ParamClientID:     []string{"test-client"},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/token", strings.NewReader(formBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should return 400 for revoked refresh token")

	var body map[string]any

	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err, "Response body should be valid JSON")

	errorCode, ok := body["error"].(string)
	require.True(t, ok, "Response should contain error field")
	require.Equal(t, cryptoutilIdentityMagic.ErrorInvalidGrant, errorCode, "Error should be invalid_grant")
}
