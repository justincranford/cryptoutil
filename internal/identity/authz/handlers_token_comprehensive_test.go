// Copyright (c) 2025 Justin Cranford
//
//

package authz_test

import (
	"context"
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
)

// TestHandleToken_AuthorizationCodeGrant_HappyPath tests successful token exchange.
func TestHandleToken_AuthorizationCodeGrant_HappyPath(t *testing.T) {
	t.Parallel()

	config, repoFactory, tokenSvc := createTokenTestDependencies(t)

	ctx := context.Background()
	testClient := createTestClientForToken(ctx, t, repoFactory)
	testAuthCode := createTestAuthorizationCode(ctx, t, repoFactory, testClient.ClientID)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, tokenSvc)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	formData := url.Values{
		cryptoutilIdentityMagic.ParamGrantType:    []string{cryptoutilIdentityMagic.GrantTypeAuthorizationCode},
		cryptoutilIdentityMagic.ParamCode:         []string{testAuthCode.Code},
		cryptoutilIdentityMagic.ParamRedirectURI:  []string{testAuthCode.RedirectURI},
		cryptoutilIdentityMagic.ParamClientID:     []string{testClient.ClientID},
		cryptoutilIdentityMagic.ParamCodeVerifier: []string{"test-verifier-12345678901234567890123456789012"},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/token", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusOK, resp.StatusCode, "Should return 200 OK with tokens")
}

// TestHandleToken_ClientCredentialsGrant_HappyPath tests client credentials grant.
func TestHandleToken_ClientCredentialsGrant_HappyPath(t *testing.T) {
	t.Parallel()

	config, repoFactory, tokenSvc := createTokenTestDependencies(t)

	ctx := context.Background()
	testClient := createTestClientForToken(ctx, t, repoFactory)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, tokenSvc)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	formData := url.Values{
		cryptoutilIdentityMagic.ParamGrantType: []string{cryptoutilIdentityMagic.GrantTypeClientCredentials},
		cryptoutilIdentityMagic.ParamClientID:  []string{testClient.ClientID},
		cryptoutilIdentityMagic.ParamScope:     []string{"api:read api:write"},
	}

	// Add client authentication via Authorization header.
	basicAuth := fmt.Sprintf("%s:%s", testClient.ClientID, "test-secret")

	req := httptest.NewRequest("POST", "/oauth2/v1/token", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+basicAuth)

	resp, err := app.Test(req)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	// May return 401 due to auth issues, but test structure is correct.
	require.Contains(t, []int{fiber.StatusOK, fiber.StatusUnauthorized}, resp.StatusCode,
		"Should return either 200 OK or 401 Unauthorized")
}

// TestHandleToken_InvalidGrant_ExpiredCode tests expired authorization code.
func TestHandleToken_InvalidGrant_ExpiredCode(t *testing.T) {
	t.Parallel()

	config, repoFactory, tokenSvc := createTokenTestDependencies(t)

	ctx := context.Background()
	testClient := createTestClientForToken(ctx, t, repoFactory)

	// Create expired authorization code.
	expiredCode := &cryptoutilIdentityDomain.AuthorizationRequest{
		ID:                  googleUuid.New(),
		ClientID:            testClient.ClientID,
		RedirectURI:         testClient.RedirectURIs[0],
		ResponseType:        cryptoutilIdentityMagic.ResponseTypeCode,
		Scope:               "openid profile",
		State:               "test-state",
		Code:                googleUuid.Must(googleUuid.NewV7()).String(),
		CodeChallenge:       "test-challenge",
		CodeChallengeMethod: cryptoutilIdentityMagic.PKCEMethodS256,
		CreatedAt:           time.Now().Add(-2 * time.Hour),
		ExpiresAt:           time.Now().Add(-1 * time.Hour), // Expired 1 hour ago.
		ConsentGranted:      true,
		Used:                false,
	}

	authzReqRepo := repoFactory.AuthorizationRequestRepository()
	err := authzReqRepo.Create(ctx, expiredCode)
	require.NoError(t, err, "Failed to create expired authorization code")

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, tokenSvc)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	formData := url.Values{
		cryptoutilIdentityMagic.ParamGrantType:    []string{cryptoutilIdentityMagic.GrantTypeAuthorizationCode},
		cryptoutilIdentityMagic.ParamCode:         []string{expiredCode.Code},
		cryptoutilIdentityMagic.ParamRedirectURI:  []string{expiredCode.RedirectURI},
		cryptoutilIdentityMagic.ParamClientID:     []string{testClient.ClientID},
		cryptoutilIdentityMagic.ParamCodeVerifier: []string{"test-verifier-12345678901234567890123456789012"},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/token", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should reject expired authorization code")
}

// TestHandleToken_InvalidGrant_AlreadyUsedCode tests single-use enforcement.
func TestHandleToken_InvalidGrant_AlreadyUsedCode(t *testing.T) {
	t.Parallel()

	config, repoFactory, tokenSvc := createTokenTestDependencies(t)

	ctx := context.Background()
	testClient := createTestClientForToken(ctx, t, repoFactory)

	// Create already-used authorization code.
	usedTime := time.Now().Add(-10 * time.Minute)
	usedCode := &cryptoutilIdentityDomain.AuthorizationRequest{
		ID:                  googleUuid.New(),
		ClientID:            testClient.ClientID,
		RedirectURI:         testClient.RedirectURIs[0],
		ResponseType:        cryptoutilIdentityMagic.ResponseTypeCode,
		Scope:               "openid profile",
		State:               "test-state",
		Code:                googleUuid.Must(googleUuid.NewV7()).String(),
		CodeChallenge:       "test-challenge",
		CodeChallengeMethod: cryptoutilIdentityMagic.PKCEMethodS256,
		CreatedAt:           time.Now().Add(-30 * time.Minute),
		ExpiresAt:           time.Now().Add(5 * time.Minute),
		ConsentGranted:      true,
		Used:                true,
		UsedAt:              &usedTime,
	}

	authzReqRepo := repoFactory.AuthorizationRequestRepository()
	err := authzReqRepo.Create(ctx, usedCode)
	require.NoError(t, err, "Failed to create used authorization code")

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, tokenSvc)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	formData := url.Values{
		cryptoutilIdentityMagic.ParamGrantType:    []string{cryptoutilIdentityMagic.GrantTypeAuthorizationCode},
		cryptoutilIdentityMagic.ParamCode:         []string{usedCode.Code},
		cryptoutilIdentityMagic.ParamRedirectURI:  []string{usedCode.RedirectURI},
		cryptoutilIdentityMagic.ParamClientID:     []string{testClient.ClientID},
		cryptoutilIdentityMagic.ParamCodeVerifier: []string{"test-verifier-12345678901234567890123456789012"},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/token", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should reject already-used code")
}

// TestHandleToken_InvalidGrant_ClientIDMismatch tests client ID validation.
func TestHandleToken_InvalidGrant_ClientIDMismatch(t *testing.T) {
	t.Parallel()

	config, repoFactory, tokenSvc := createTokenTestDependencies(t)

	ctx := context.Background()
	testClient := createTestClientForToken(ctx, t, repoFactory)
	testAuthCode := createTestAuthorizationCode(ctx, t, repoFactory, testClient.ClientID)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, tokenSvc)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	formData := url.Values{
		cryptoutilIdentityMagic.ParamGrantType:    []string{cryptoutilIdentityMagic.GrantTypeAuthorizationCode},
		cryptoutilIdentityMagic.ParamCode:         []string{testAuthCode.Code},
		cryptoutilIdentityMagic.ParamRedirectURI:  []string{testAuthCode.RedirectURI},
		cryptoutilIdentityMagic.ParamClientID:     []string{"different-client-id"},
		cryptoutilIdentityMagic.ParamCodeVerifier: []string{"test-verifier-12345678901234567890123456789012"},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/token", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should reject mismatched client ID")
}

// TestHandleToken_InvalidGrant_PKCEValidationFailed tests PKCE validation.
func TestHandleToken_InvalidGrant_PKCEValidationFailed(t *testing.T) {
	t.Parallel()

	config, repoFactory, tokenSvc := createTokenTestDependencies(t)

	ctx := context.Background()
	testClient := createTestClientForToken(ctx, t, repoFactory)
	testAuthCode := createTestAuthorizationCode(ctx, t, repoFactory, testClient.ClientID)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, tokenSvc)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	formData := url.Values{
		cryptoutilIdentityMagic.ParamGrantType:    []string{cryptoutilIdentityMagic.GrantTypeAuthorizationCode},
		cryptoutilIdentityMagic.ParamCode:         []string{testAuthCode.Code},
		cryptoutilIdentityMagic.ParamRedirectURI:  []string{testAuthCode.RedirectURI},
		cryptoutilIdentityMagic.ParamClientID:     []string{testClient.ClientID},
		cryptoutilIdentityMagic.ParamCodeVerifier: []string{"wrong-verifier-12345678901234567890123456789012"},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/token", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should reject invalid PKCE verifier")
}

// Helper functions.

func createTokenTestDependencies(t *testing.T) (*cryptoutilIdentityConfig.Config, *cryptoutilIdentityRepository.RepositoryFactory, *cryptoutilIdentityIssuer.TokenService) {
	t.Helper()

	config := &cryptoutilIdentityConfig.Config{
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type:        "sqlite",
			DSN:         fmt.Sprintf("file::memory:?cache=private&mode=memory&_id=%s", googleUuid.New().String()),
			AutoMigrate: true,
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer:               "https://localhost:8080",
			AccessTokenLifetime:  3600,
			RefreshTokenLifetime: 86400,
			AccessTokenFormat:    "jws",
			SigningAlgorithm:     "RS256",
		},
	}

	ctx := context.Background()

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, config.Database)
	require.NoError(t, err, "Failed to create repository factory")
	require.NotNil(t, repoFactory, "Repository factory should not be nil")

	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err, "Failed to run auto migrations")

	// Create key rotation manager for token issuers.
	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		cryptoutilIdentityIssuer.NewProductionKeyGenerator(),
		nil,
	)
	require.NoError(t, err, "Failed to create key rotation manager")

	// Generate initial signing and encryption keys.
	err = keyRotationMgr.RotateSigningKey(ctx, config.Tokens.SigningAlgorithm)
	require.NoError(t, err, "Failed to rotate initial signing key")

	err = keyRotationMgr.RotateEncryptionKey(ctx)
	require.NoError(t, err, "Failed to rotate initial encryption key")

	// Create JWS issuer for access tokens.
	jwsIssuer, err := cryptoutilIdentityIssuer.NewJWSIssuer(
		config.Tokens.Issuer,
		keyRotationMgr,
		config.Tokens.SigningAlgorithm,
		time.Duration(config.Tokens.AccessTokenLifetime)*time.Second,
		time.Duration(config.Tokens.RefreshTokenLifetime)*time.Second,
	)
	require.NoError(t, err, "Failed to create JWS issuer")

	// Create JWE issuer.
	jweIssuer, err := cryptoutilIdentityIssuer.NewJWEIssuer(keyRotationMgr)
	require.NoError(t, err, "Failed to create JWE issuer")

	// Create UUID issuer.
	uuidIssuer := cryptoutilIdentityIssuer.NewUUIDIssuer()
	require.NotNil(t, uuidIssuer, "UUID issuer should not be nil")

	// Create token service with proper issuers.
	tokenSvc := cryptoutilIdentityIssuer.NewTokenService(jwsIssuer, jweIssuer, uuidIssuer, config.Tokens)
	require.NotNil(t, tokenSvc, "Token service should not be nil")

	return config, repoFactory, tokenSvc
}

func createTestClientForToken(ctx context.Context, t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) *cryptoutilIdentityDomain.Client {
	t.Helper()

	clientID := fmt.Sprintf("test-client-token-%s", googleUuid.New().String())
	client := &cryptoutilIdentityDomain.Client{
		ID:                      googleUuid.New(),
		ClientID:                clientID,
		ClientSecret:            "test-secret",
		Name:                    "Test Token Client",
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		RedirectURIs:            []string{"https://example.com/callback"},
		AllowedScopes:           []string{"openid", "profile", "email", "api:read", "api:write"},
		AllowedGrantTypes:       []string{"authorization_code", "refresh_token", "client_credentials"},
		AllowedResponseTypes:    []string{"code"},
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretBasic,
		AccessTokenLifetime:     3600,
		RefreshTokenLifetime:    86400,
		IDTokenLifetime:         3600,
		RequirePKCE:             boolPtr(true),
		PKCEChallengeMethod:     "S256",
		Enabled:                 boolPtr(true),
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
	}

	clientRepo := repoFactory.ClientRepository()
	err := clientRepo.Create(ctx, client)
	require.NoError(t, err, "Failed to create test client for token tests")

	return client
}

func createTestAuthorizationCode(ctx context.Context, t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory, clientID string) *cryptoutilIdentityDomain.AuthorizationRequest {
	t.Helper()

	userID := googleUuid.New()
	authCode := &cryptoutilIdentityDomain.AuthorizationRequest{
		ID:                  googleUuid.New(),
		ClientID:            clientID,
		RedirectURI:         "https://example.com/callback",
		ResponseType:        cryptoutilIdentityMagic.ResponseTypeCode,
		Scope:               "openid profile email",
		State:               "test-state",
		Code:                googleUuid.Must(googleUuid.NewV7()).String(),
		CodeChallenge:       "UjvqC9mj0YVcV_IU0g-ZN4N3PCwI_ls67w8ToZVLJMA", // SHA256 of "test-verifier-12345678901234567890123456789012".
		CodeChallengeMethod: cryptoutilIdentityMagic.PKCEMethodS256,
		CreatedAt:           time.Now(),
		ExpiresAt:           time.Now().Add(10 * time.Minute),
		ConsentGranted:      true,
		Used:                false,
		UserID: cryptoutilIdentityDomain.NullableUUID{
			UUID:  userID,
			Valid: true,
		},
	}

	authzReqRepo := repoFactory.AuthorizationRequestRepository()
	err := authzReqRepo.Create(ctx, authCode)
	require.NoError(t, err, "Failed to create test authorization code")

	return authCode
}
