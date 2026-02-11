// Copyright (c) 2025 Justin Cranford
//
//

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

// TestIntrospectionRevocationFlow validates the complete flow:
// 1. Token is active (introspect returns active:true)
// 2. Token is revoked (revoke returns 200 OK)
// 3. Token is inactive (introspect returns active:false)
//
// This is the CRITICAL P3.2.5 integration test for introspection revocation check.
func TestIntrospectionRevocationFlow(t *testing.T) {
	t.Parallel()

	config := createRevocationFlowTestConfig(t)
	repoFactory := createRevocationFlowTestRepoFactory(t)
	testClient := createRevocationFlowTestClient(t, repoFactory)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	// Create a fresh token for this test.
	testToken := createRevocationFlowTestToken(t, repoFactory, testClient.ID, false)

	// Step 1: Verify token is ACTIVE before revocation.
	t.Log("Step 1: Verifying token is active before revocation...")

	introspectResult := introspectTokenForRevocationFlow(t, app, testToken.TokenValue)

	active, ok := introspectResult["active"].(bool)
	require.True(t, ok, "Response should have active field")
	require.True(t, active, "Token should be ACTIVE before revocation")

	clientID, ok := introspectResult["client_id"].(string)
	require.True(t, ok, "Response should have client_id field")
	require.Equal(t, testClient.ID.String(), clientID, "Token should belong to test client")

	t.Log("✅ Step 1 PASSED: Token is active")

	// Step 2: REVOKE the token.
	t.Log("Step 2: Revoking token...")

	revokeFormBody := url.Values{
		cryptoutilIdentityMagic.ParamToken: []string{testToken.TokenValue},
	}

	revokeReq := httptest.NewRequest("POST", "/oauth2/v1/revoke", strings.NewReader(revokeFormBody.Encode()))
	revokeReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	revokeResp, err := app.Test(revokeReq)
	require.NoError(t, err, "Revocation request should succeed")

	defer func() { _ = revokeResp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusOK, revokeResp.StatusCode, "Revocation should return 200 OK per RFC 7009")

	t.Log("✅ Step 2 PASSED: Token revoked successfully")

	// Step 3: Verify token is INACTIVE after revocation.
	t.Log("Step 3: Verifying token is inactive after revocation...")

	postRevokeResult := introspectTokenForRevocationFlow(t, app, testToken.TokenValue)

	postRevokeActive, ok := postRevokeResult["active"].(bool)
	require.True(t, ok, "Post-revoke response should have active field")
	require.False(t, postRevokeActive, "Token should be INACTIVE after revocation")

	t.Log("✅ Step 3 PASSED: Token is inactive after revocation")
	t.Log("✅ FULL FLOW PASSED: Introspection correctly validates revocation status")
}

// TestIntrospectionRefreshTokenRevocation validates revocation of refresh tokens.
func TestIntrospectionRefreshTokenRevocation(t *testing.T) {
	t.Parallel()

	config := createRevocationFlowTestConfig(t)
	repoFactory := createRevocationFlowTestRepoFactory(t)
	testClient := createRevocationFlowTestClient(t, repoFactory)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	// Create a refresh token.
	ctx := context.Background()
	tokenRepo := repoFactory.TokenRepository()

	tokenID := googleUuid.Must(googleUuid.NewV7())
	refreshToken := &cryptoutilIdentityDomain.Token{
		ID:         tokenID,
		TokenValue: fmt.Sprintf("refresh-token-%s", tokenID.String()),
		TokenType:  cryptoutilIdentityDomain.TokenTypeRefresh,
		ClientID:   testClient.ID,
		Scopes:     []string{"openid", "offline_access"},
		ExpiresAt:  time.Now().UTC().Add(7 * 24 * time.Hour), // 7 days.
		IssuedAt:   time.Now().UTC(),
	}

	err := tokenRepo.Create(ctx, refreshToken)
	require.NoError(t, err, "Failed to create refresh token")

	// Step 1: Verify refresh token is active.
	result := introspectTokenForRevocationFlow(t, app, refreshToken.TokenValue)
	active, _ := result["active"].(bool) //nolint:errcheck // Test assertion
	require.True(t, active, "Refresh token should be active before revocation")

	// Step 2: Revoke with token_type_hint=refresh_token.
	revokeFormBody := url.Values{
		cryptoutilIdentityMagic.ParamToken:         []string{refreshToken.TokenValue},
		cryptoutilIdentityMagic.ParamTokenTypeHint: []string{cryptoutilIdentityMagic.TokenTypeRefreshToken},
	}

	revokeReq := httptest.NewRequest("POST", "/oauth2/v1/revoke", strings.NewReader(revokeFormBody.Encode()))
	revokeReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	revokeResp, err := app.Test(revokeReq)
	require.NoError(t, err)

	defer func() { _ = revokeResp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusOK, revokeResp.StatusCode)

	// Step 3: Verify refresh token is inactive.
	postRevokeResult := introspectTokenForRevocationFlow(t, app, refreshToken.TokenValue)
	postRevokeActive, _ := postRevokeResult["active"].(bool) //nolint:errcheck // Test assertion
	require.False(t, postRevokeActive, "Refresh token should be inactive after revocation")
}

// TestIntrospectionTokenTypeHintMismatch validates that token_type_hint mismatch is handled.
func TestIntrospectionTokenTypeHintMismatch(t *testing.T) {
	t.Parallel()

	config := createRevocationFlowTestConfig(t)
	repoFactory := createRevocationFlowTestRepoFactory(t)
	testClient := createRevocationFlowTestClient(t, repoFactory)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	// Create an access token.
	accessToken := createRevocationFlowTestToken(t, repoFactory, testClient.ID, false)

	// Try to revoke access token with refresh_token hint - should fail.
	revokeFormBody := url.Values{
		cryptoutilIdentityMagic.ParamToken:         []string{accessToken.TokenValue},
		cryptoutilIdentityMagic.ParamTokenTypeHint: []string{cryptoutilIdentityMagic.TokenTypeRefreshToken},
	}

	revokeReq := httptest.NewRequest("POST", "/oauth2/v1/revoke", strings.NewReader(revokeFormBody.Encode()))
	revokeReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	revokeResp, err := app.Test(revokeReq)
	require.NoError(t, err)

	defer func() { _ = revokeResp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusBadRequest, revokeResp.StatusCode, "Should return 400 for token type mismatch")
}

// TestIntrospectionMultipleRevocationsIdempotent validates that revoking an already-revoked token is idempotent.
func TestIntrospectionMultipleRevocationsIdempotent(t *testing.T) {
	t.Parallel()

	config := createRevocationFlowTestConfig(t)
	repoFactory := createRevocationFlowTestRepoFactory(t)
	testClient := createRevocationFlowTestClient(t, repoFactory)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	// Create a token.
	testToken := createRevocationFlowTestToken(t, repoFactory, testClient.ID, false)

	// Revoke the token multiple times - all should succeed (idempotent per RFC 7009).
	for i := 0; i < 3; i++ {
		revokeFormBody := url.Values{
			cryptoutilIdentityMagic.ParamToken: []string{testToken.TokenValue},
		}

		revokeReq := httptest.NewRequest("POST", "/oauth2/v1/revoke", strings.NewReader(revokeFormBody.Encode()))
		revokeReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		revokeResp, err := app.Test(revokeReq)
		require.NoError(t, err, "Revocation request %d should succeed", i+1)

		defer func() { _ = revokeResp.Body.Close() }() //nolint:errcheck // Test cleanup

		require.Equal(t, fiber.StatusOK, revokeResp.StatusCode, "Revocation %d should return 200 OK", i+1)
	}

	// Verify token is still inactive.
	result := introspectTokenForRevocationFlow(t, app, testToken.TokenValue)
	active, _ := result["active"].(bool) //nolint:errcheck // Test assertion
	require.False(t, active, "Token should remain inactive after multiple revocations")
}

// TestIntrospectionNonExistentToken validates introspection of non-existent tokens.
func TestIntrospectionNonExistentToken(t *testing.T) {
	t.Parallel()

	config := createRevocationFlowTestConfig(t)
	repoFactory := createRevocationFlowTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	// Introspect a non-existent token.
	result := introspectTokenForRevocationFlow(t, app, "non-existent-token-xyz123")

	active, ok := result["active"].(bool)
	require.True(t, ok, "Response should have active field")
	require.False(t, active, "Non-existent token should return active:false")
}

// introspectTokenForRevocationFlow performs token introspection for revocation flow tests.
func introspectTokenForRevocationFlow(t *testing.T, app *fiber.App, tokenValue string) map[string]any {
	t.Helper()

	formBody := url.Values{
		cryptoutilIdentityMagic.ParamToken: []string{tokenValue},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/introspect", strings.NewReader(formBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	require.NoError(t, err, "Introspection request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusOK, resp.StatusCode, "Introspection should return 200 OK")

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Should decode introspection response")

	return result
}

// createRevocationFlowTestConfig creates a test configuration.
func createRevocationFlowTestConfig(t *testing.T) *cryptoutilIdentityConfig.Config {
	t.Helper()

	testID := googleUuid.Must(googleUuid.NewV7()).String()

	return &cryptoutilIdentityConfig.Config{
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type: "sqlite",
			DSN:  fmt.Sprintf("file:revocation_flow_test_%s.db?mode=memory&cache=shared", testID),
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer: "https://localhost:8080",
		},
	}
}

// createRevocationFlowTestRepoFactory creates a repository factory.
func createRevocationFlowTestRepoFactory(t *testing.T) *cryptoutilIdentityRepository.RepositoryFactory {
	t.Helper()

	cfg := createRevocationFlowTestConfig(t)
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

	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err, "Failed to run auto migrations")

	return repoFactory
}

// createRevocationFlowTestClient creates a test client.
func createRevocationFlowTestClient(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) *cryptoutilIdentityDomain.Client {
	t.Helper()

	ctx := context.Background()
	clientRepo := repoFactory.ClientRepository()

	clientUUID, err := googleUuid.NewV7()
	require.NoError(t, err, "Failed to generate client UUID")

	testClient := &cryptoutilIdentityDomain.Client{
		ID:                      clientUUID,
		ClientID:                fmt.Sprintf("revoke-test-client-%s", clientUUID.String()),
		Name:                    "Revocation Flow Test Client",
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		AllowedGrantTypes:       []string{cryptoutilIdentityMagic.GrantTypeClientCredentials},
		AllowedScopes:           []string{"openid", "profile", "email", "offline_access"},
		RedirectURIs:            []string{"https://example.com/callback"},
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretBasic,
	}

	err = clientRepo.Create(ctx, testClient)
	require.NoError(t, err, "Failed to create test client")

	return testClient
}

// createRevocationFlowTestToken creates a test access token.
func createRevocationFlowTestToken(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory, clientID googleUuid.UUID, revoked bool) *cryptoutilIdentityDomain.Token {
	t.Helper()

	ctx := context.Background()
	tokenRepo := repoFactory.TokenRepository()

	tokenID := googleUuid.Must(googleUuid.NewV7())

	testToken := &cryptoutilIdentityDomain.Token{
		ID:         tokenID,
		TokenValue: fmt.Sprintf("revoke-flow-token-%s", tokenID.String()),
		TokenType:  cryptoutilIdentityDomain.TokenTypeAccess,
		ClientID:   clientID,
		Scopes:     []string{"openid", "profile"},
		ExpiresAt:  time.Now().UTC().Add(1 * time.Hour),
		IssuedAt:   time.Now().UTC(),
	}

	if revoked {
		now := time.Now().UTC()
		testToken.RevokedAt = &now
	}

	err := tokenRepo.Create(ctx, testToken)
	require.NoError(t, err, "Failed to create test token")

	return testToken
}
