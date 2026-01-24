// Copyright (c) 2025 Justin Cranford

package authz_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/identity/authz"
	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityIssuer "cryptoutil/internal/identity/issuer"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
	cryptoutilSharedCryptoHash "cryptoutil/internal/shared/crypto/hash"
)

// TestAuthenticateClient_BasicAuthSuccess validates HTTP Basic authentication success.
func TestAuthenticateClient_BasicAuthSuccess(t *testing.T) {
	t.Parallel()

	config := createClientAuthFlowTestConfig(t)
	repoFactory := createClientAuthFlowTestRepoFactory(t, config)

	testClient := createClientAuthFlowTestClient(t, repoFactory, cryptoutilIdentityDomain.ClientAuthMethodSecretBasic)

	// Create token service using ProductionKeyGenerator with RS256 signing.
	tokenSvc := createClientAuthFlowTestTokenService(t, config)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, tokenSvc)
	require.NotNil(t, svc, "Service should not be nil")

	err := svc.Start(context.Background())
	require.NoError(t, err, "Service start should succeed")

	defer func() {
		err := svc.Stop(context.Background())
		require.NoError(t, err, "Service stop should succeed")
	}()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			t.Logf("Fiber error handler called: %v", err)
			t.Logf("Fiber error type: %T", err)
			t.Logf("Fiber error details: %+v", err)

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})
	svc.RegisterRoutes(app)

	formBody := url.Values{
		cryptoutilIdentityMagic.ParamGrantType: []string{cryptoutilIdentityMagic.GrantTypeClientCredentials},
	}

	basicAuth := base64.StdEncoding.EncodeToString([]byte(testClient.ClientID + ":test-secret"))

	req := httptest.NewRequest("POST", "/oauth2/v1/token", strings.NewReader(formBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+basicAuth)

	// Use 30 second timeout for parallel test execution under load.
	resp, err := app.Test(req, cryptoutilIdentityMagic.FiberTestTimeoutMs)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	// Log response details if not 200.
	if resp.StatusCode != fiber.StatusOK {
		bodyBytes := make([]byte, 1024)
		n, _ := resp.Body.Read(bodyBytes) //nolint:errcheck // Test logging
		t.Logf("Response status: %d, body: %s", resp.StatusCode, string(bodyBytes[:n]))
	}

	require.Equal(t, fiber.StatusOK, resp.StatusCode, "Basic auth should succeed")
}

// TestAuthenticateClient_PostAuthSuccess validates POST body authentication success.
func TestAuthenticateClient_PostAuthSuccess(t *testing.T) {
	t.Parallel()

	config := createClientAuthFlowTestConfig(t)
	repoFactory := createClientAuthFlowTestRepoFactory(t, config)

	testClient := createClientAuthFlowTestClient(t, repoFactory, cryptoutilIdentityDomain.ClientAuthMethodSecretPost)

	// Create token service using ProductionKeyGenerator with RS256 signing.
	tokenSvc := createClientAuthFlowTestTokenService(t, config)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, tokenSvc)
	require.NotNil(t, svc, "Service should not be nil")

	err := svc.Start(context.Background())
	require.NoError(t, err, "Service start should succeed")

	defer func() {
		err := svc.Stop(context.Background())
		require.NoError(t, err, "Service stop should succeed")
	}()

	app := fiber.New()
	svc.RegisterRoutes(app)

	formBody := url.Values{
		cryptoutilIdentityMagic.ParamGrantType:    []string{cryptoutilIdentityMagic.GrantTypeClientCredentials},
		cryptoutilIdentityMagic.ParamClientID:     []string{testClient.ClientID},
		cryptoutilIdentityMagic.ParamClientSecret: []string{"test-secret"},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/token", strings.NewReader(formBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Use 30 second timeout for parallel test execution under load.
	resp, err := app.Test(req, cryptoutilIdentityMagic.FiberTestTimeoutMs)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusOK, resp.StatusCode, "POST auth should succeed")
}

// TestAuthenticateClient_NoCredentialsFailure validates missing credentials error.
func TestAuthenticateClient_NoCredentialsFailure(t *testing.T) {
	t.Parallel()

	config := createClientAuthFlowTestConfig(t)
	repoFactory := createClientAuthFlowTestRepoFactory(t, config)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	err := svc.Start(context.Background())
	require.NoError(t, err, "Service start should succeed")

	defer func() {
		err := svc.Stop(context.Background())
		require.NoError(t, err, "Service stop should succeed")
	}()

	app := fiber.New()
	svc.RegisterRoutes(app)

	formBody := url.Values{
		cryptoutilIdentityMagic.ParamGrantType: []string{cryptoutilIdentityMagic.GrantTypeClientCredentials},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/token", strings.NewReader(formBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Use 30 second timeout for parallel test execution under load.
	resp, err := app.Test(req, cryptoutilIdentityMagic.FiberTestTimeoutMs)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode, "Should return 401 for missing credentials")
}

func createClientAuthFlowTestConfig(t *testing.T) *cryptoutilIdentityConfig.Config {
	t.Helper()

	testID := googleUuid.Must(googleUuid.NewV7()).String()

	return &cryptoutilIdentityConfig.Config{
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type: "sqlite",
			DSN:  fmt.Sprintf("file:test_%s.db?mode=memory&cache=shared", testID),
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer:              "https://localhost:8080",
			SigningAlgorithm:    "RS256",
			AccessTokenFormat:   "jws", // Use JWS format for testing (legacy issuer)
			AccessTokenLifetime: 3600,
		},
	}
}

func createClientAuthFlowTestRepoFactory(t *testing.T, cfg *cryptoutilIdentityConfig.Config) *cryptoutilIdentityRepository.RepositoryFactory {
	t.Helper()

	ctx := context.Background()

	// Each test uses a unique DSN (via googleUuid.NewV7() in createClientAuthFlowTestConfig),
	// so no migration state reset is needed. The in-memory database check in Migrate()
	// already skips caching for in-memory databases.

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type:        cfg.Database.Type,
		DSN:         cfg.Database.DSN,
		AutoMigrate: true,
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err, "Failed to create repository factory")
	require.NotNil(t, repoFactory, "Repository factory should not be nil")

	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err, "Failed to run auto migrations")

	return repoFactory
}

func createClientAuthFlowTestClient(
	t *testing.T,
	repoFactory *cryptoutilIdentityRepository.RepositoryFactory,
	authMethod cryptoutilIdentityDomain.ClientAuthMethod,
) *cryptoutilIdentityDomain.Client {
	t.Helper()

	ctx := context.Background()
	clientRepo := repoFactory.ClientRepository()

	clientUUID, err := googleUuid.NewV7()
	require.NoError(t, err, "Failed to generate client UUID")

	// Generate proper PBKDF2 hash for "test-secret" using cryptoutilCrypto.HashSecretPBKDF2
	// Format: $pbkdf2-sha256$iterations$base64(salt)$base64(hash)
	hashedSecret, err := cryptoutilSharedCryptoHash.HashSecretPBKDF2("test-secret")
	require.NoError(t, err, "Failed to hash client secret")

	testClient := &cryptoutilIdentityDomain.Client{
		ID:                      clientUUID,
		ClientID:                fmt.Sprintf("test-client-%s", clientUUID.String()),
		Name:                    "Test Client",
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		AllowedGrantTypes:       []string{cryptoutilIdentityMagic.GrantTypeClientCredentials},
		AllowedScopes:           []string{"openid", "profile", "email"},
		RedirectURIs:            []string{"https://example.com/callback"},
		TokenEndpointAuthMethod: authMethod,
		ClientSecret:            hashedSecret,
		AccessTokenLifetime:     3600,
		RefreshTokenLifetime:    86400,
		IDTokenLifetime:         3600,
	}

	err = clientRepo.Create(ctx, testClient)
	require.NoError(t, err, "Failed to create test client")

	return testClient
}

func createClientAuthFlowTestTokenService(t *testing.T, config *cryptoutilIdentityConfig.Config) *cryptoutilIdentityIssuer.TokenService {
	t.Helper()

	ctx := context.Background()

	// Create key rotation manager for token issuers.
	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		cryptoutilIdentityIssuer.NewProductionKeyGenerator(),
		nil,
	)
	require.NoError(t, err, "Failed to create key rotation manager")

	// Generate initial signing key.
	err = keyRotationMgr.RotateSigningKey(ctx, config.Tokens.SigningAlgorithm)
	require.NoError(t, err, "Failed to rotate initial signing key")

	// Create JWS issuer for access tokens.
	jwsIssuer, err := cryptoutilIdentityIssuer.NewJWSIssuer(
		config.Tokens.Issuer,
		keyRotationMgr,
		config.Tokens.SigningAlgorithm,
		time.Duration(config.Tokens.AccessTokenLifetime)*time.Second,
		time.Duration(config.Tokens.AccessTokenLifetime)*time.Second,
	)
	require.NoError(t, err, "Failed to create JWS issuer")

	uuidIssuer := cryptoutilIdentityIssuer.NewUUIDIssuer()

	return cryptoutilIdentityIssuer.NewTokenService(jwsIssuer, nil, uuidIssuer, config.Tokens)
}
