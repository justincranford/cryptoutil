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

	"github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilCrypto "cryptoutil/internal/crypto"
	"cryptoutil/internal/identity/authz"
	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityIssuer "cryptoutil/internal/identity/issuer"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
)

// TestAuthenticateClient_BasicAuthSuccess validates HTTP Basic authentication success.
func TestAuthenticateClient_BasicAuthSuccess(t *testing.T) {
	t.Parallel()

	config := createClientAuthFlowTestConfig(t)
	repoFactory := createClientAuthFlowTestRepoFactory(t)

	testClient := createClientAuthFlowTestClient(t, repoFactory, cryptoutilIdentityDomain.ClientAuthMethodSecretBasic)

	// Create token service using legacy JWS issuer (no key rotation manager needed for testing)
	jwsIssuer := createClientAuthFlowTestJWSIssuerLegacy(t, config)
	tokenSvc := createClientAuthFlowTestTokenServiceLegacy(t, jwsIssuer, config)

	svc := authz.NewService(config, repoFactory, tokenSvc)
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

	resp, err := app.Test(req)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	// Log response details if not 200
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
	repoFactory := createClientAuthFlowTestRepoFactory(t)

	testClient := createClientAuthFlowTestClient(t, repoFactory, cryptoutilIdentityDomain.ClientAuthMethodSecretPost)

	// Create token service using legacy JWS issuer (no key rotation manager needed for testing)
	jwsIssuer := createClientAuthFlowTestJWSIssuerLegacy(t, config)
	tokenSvc := createClientAuthFlowTestTokenServiceLegacy(t, jwsIssuer, config)

	svc := authz.NewService(config, repoFactory, tokenSvc)
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

	resp, err := app.Test(req)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusOK, resp.StatusCode, "POST auth should succeed")
}

// TestAuthenticateClient_NoCredentialsFailure validates missing credentials error.
func TestAuthenticateClient_NoCredentialsFailure(t *testing.T) {
	t.Parallel()

	config := createClientAuthFlowTestConfig(t)
	repoFactory := createClientAuthFlowTestRepoFactory(t)

	svc := authz.NewService(config, repoFactory, nil)
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

	resp, err := app.Test(req)
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

func createClientAuthFlowTestRepoFactory(t *testing.T) *cryptoutilIdentityRepository.RepositoryFactory {
	t.Helper()

	cfg := createClientAuthFlowTestConfig(t)
	ctx := context.Background()

	// Clear migration state to ensure fresh database for this test.
	cryptoutilIdentityRepository.ResetMigrationStateForTesting()

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
	hashedSecret, err := cryptoutilCrypto.HashSecretPBKDF2("test-secret")
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

func createClientAuthFlowTestJWSIssuerLegacy(t *testing.T, config *cryptoutilIdentityConfig.Config) *cryptoutilIdentityIssuer.JWSIssuer {
	t.Helper()

	// Use legacy JWS issuer with simple signing key (no key rotation manager)
	signingKey := []byte("test-signing-key-32-bytes-long!!") // 32 bytes for HS256
	signingAlg := "HS256"                                    // HMAC SHA-256 for testing

	jwsIssuer, err := cryptoutilIdentityIssuer.NewJWSIssuerLegacy(
		config.Tokens.Issuer,
		signingKey,
		signingAlg,
		time.Duration(config.Tokens.AccessTokenLifetime)*time.Second,
		time.Duration(config.Tokens.AccessTokenLifetime)*time.Second,
	)
	require.NoError(t, err, "Failed to create legacy JWS issuer")
	require.NotNil(t, jwsIssuer, "JWS issuer should not be nil")

	return jwsIssuer
}

func createClientAuthFlowTestTokenServiceLegacy(
	t *testing.T,
	jwsIssuer *cryptoutilIdentityIssuer.JWSIssuer,
	config *cryptoutilIdentityConfig.Config,
) *cryptoutilIdentityIssuer.TokenService {
	t.Helper()

	// Create UUID issuer (no JWE needed for testing)
	uuidIssuer := cryptoutilIdentityIssuer.NewUUIDIssuer()
	require.NotNil(t, uuidIssuer, "UUID issuer should not be nil")

	// Create token service with nil JWE issuer (not needed for client_credentials grant)
	tokenSvc := cryptoutilIdentityIssuer.NewTokenService(jwsIssuer, nil, uuidIssuer, config.Tokens)
	require.NotNil(t, tokenSvc, "Token service should not be nil")

	return tokenSvc
}
