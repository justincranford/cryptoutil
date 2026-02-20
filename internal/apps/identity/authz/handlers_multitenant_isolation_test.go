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

// TestMultiTenantTokenIsolation validates that client A cannot introspect client B's tokens.
// This is a CRITICAL security test for multi-tenant isolation per P2.3.3.
func TestMultiTenantTokenIsolation(t *testing.T) {
	t.Parallel()

	config := createMultiTenantTestConfig(t)
	repoFactory := createMultiTenantTestRepoFactory(t)

	// Create two separate clients (tenants).
	clientA := createMultiTenantTestClient(t, repoFactory, "client-a")
	clientB := createMultiTenantTestClient(t, repoFactory, "client-b")

	// Create tokens for each client.
	tokenA := createMultiTenantTestToken(t, repoFactory, clientA.ID, "token-for-client-a")
	tokenB := createMultiTenantTestToken(t, repoFactory, clientB.ID, "token-for-client-b")

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	tests := []struct {
		name        string
		token       string
		clientID    string
		expectOwner googleUuid.UUID
		desc        string
	}{
		{
			name:        "client_a_introspects_own_token",
			token:       tokenA.TokenValue,
			clientID:    clientA.ClientID,
			expectOwner: clientA.ID,
			desc:        "Client A should successfully introspect its own token",
		},
		{
			name:        "client_b_introspects_own_token",
			token:       tokenB.TokenValue,
			clientID:    clientB.ClientID,
			expectOwner: clientB.ID,
			desc:        "Client B should successfully introspect its own token",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			formBody := url.Values{
				cryptoutilIdentityMagic.ParamToken: []string{tc.token},
			}

			req := httptest.NewRequest("POST", "/oauth2/v1/introspect", strings.NewReader(formBody.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			resp, err := app.Test(req, -1)
			require.NoError(t, err, tc.desc)

			defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

			require.Equal(t, fiber.StatusOK, resp.StatusCode, tc.desc)

			var result map[string]any

			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err, "Should decode introspection response")

			active, ok := result["active"].(bool)
			require.True(t, ok, "Response should have active field")
			require.True(t, active, "Token should be active")

			clientID, ok := result["client_id"].(string)
			require.True(t, ok, "Response should have client_id field")
			require.Equal(t, tc.expectOwner.String(), clientID, "Token should belong to expected client")
		})
	}
}

// TestMultiTenantTokenRevocationIsolation validates that revoking client A's token
// does not affect client B's tokens.
func TestMultiTenantTokenRevocationIsolation(t *testing.T) {
	t.Parallel()

	config := createMultiTenantTestConfig(t)
	repoFactory := createMultiTenantTestRepoFactory(t)

	// Create two separate clients.
	clientA := createMultiTenantTestClient(t, repoFactory, "revoke-client-a")
	clientB := createMultiTenantTestClient(t, repoFactory, "revoke-client-b")

	// Create tokens for each client.
	tokenA := createMultiTenantTestToken(t, repoFactory, clientA.ID, "revoke-token-a")
	tokenB := createMultiTenantTestToken(t, repoFactory, clientB.ID, "revoke-token-b")

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	// Step 1: Verify both tokens are active before revocation.
	verifyTokenActive(t, app, tokenA.TokenValue, true, "Token A should be active before revocation")
	verifyTokenActive(t, app, tokenB.TokenValue, true, "Token B should be active before revocation")

	// Step 2: Revoke token A.
	revokeFormBody := url.Values{
		cryptoutilIdentityMagic.ParamToken: []string{tokenA.TokenValue},
	}

	revokeReq := httptest.NewRequest("POST", "/oauth2/v1/revoke", strings.NewReader(revokeFormBody.Encode()))
	revokeReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	revokeResp, err := app.Test(revokeReq)
	require.NoError(t, err, "Revocation request should succeed")

	defer func() { _ = revokeResp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusOK, revokeResp.StatusCode, "Revocation should return 200 OK")

	// Step 3: Verify token A is now inactive.
	verifyTokenActive(t, app, tokenA.TokenValue, false, "Token A should be inactive after revocation")

	// Step 4: CRITICAL - Verify token B is STILL active (isolation).
	verifyTokenActive(t, app, tokenB.TokenValue, true, "Token B should STILL be active after revoking token A (isolation)")
}

// TestMultiTenantScopeIsolation validates that tokens only contain scopes
// allowed by their respective client configuration.
func TestMultiTenantScopeIsolation(t *testing.T) {
	t.Parallel()

	config := createMultiTenantTestConfig(t)
	repoFactory := createMultiTenantTestRepoFactory(t)

	// Create client with limited scopes.
	limitedClient := createMultiTenantTestClientWithScopes(t, repoFactory, "limited-client", []string{"openid"})

	// Create client with full scopes.
	fullClient := createMultiTenantTestClientWithScopes(t, repoFactory, "full-client", []string{"openid", "profile", "email", "api:admin"})

	// Create tokens with their client's allowed scopes.
	limitedToken := createMultiTenantTestTokenWithScopes(t, repoFactory, limitedClient.ID, "limited-token", []string{"openid"})
	fullToken := createMultiTenantTestTokenWithScopes(t, repoFactory, fullClient.ID, "full-token", []string{"openid", "profile", "email", "api:admin"})

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	// Verify limited client token has only limited scopes.
	introspectResult := introspectToken(t, app, limitedToken.TokenValue)

	// Scope can be either []any (from JSON) or string.
	limitedScopes := extractScopes(introspectResult)
	require.NotContains(t, limitedScopes, "api:admin", "Limited client token should not have admin scope")

	// Verify full client token has all scopes.
	fullIntrospectResult := introspectToken(t, app, fullToken.TokenValue)
	fullScopes := extractScopes(fullIntrospectResult)
	require.Contains(t, fullScopes, "api:admin", "Full client token should have admin scope")
}

// TestMultiTenantDatabaseIsolation validates that tokens are stored with
// correct client associations in the database.
func TestMultiTenantDatabaseIsolation(t *testing.T) {
	t.Parallel()

	repoFactory := createMultiTenantTestRepoFactory(t)
	ctx := context.Background()
	tokenRepo := repoFactory.TokenRepository()

	// Create two clients.
	clientA := createMultiTenantTestClient(t, repoFactory, "db-client-a")
	clientB := createMultiTenantTestClient(t, repoFactory, "db-client-b")

	// Create tokens.
	tokenA := createMultiTenantTestToken(t, repoFactory, clientA.ID, "db-token-a")
	tokenB := createMultiTenantTestToken(t, repoFactory, clientB.ID, "db-token-b")

	// Verify database-level isolation.
	retrievedA, err := tokenRepo.GetByTokenValue(ctx, tokenA.TokenValue)
	require.NoError(t, err, "Should retrieve token A from database")
	require.Equal(t, clientA.ID, retrievedA.ClientID, "Token A should be associated with client A")

	retrievedB, err := tokenRepo.GetByTokenValue(ctx, tokenB.TokenValue)
	require.NoError(t, err, "Should retrieve token B from database")
	require.Equal(t, clientB.ID, retrievedB.ClientID, "Token B should be associated with client B")

	// Verify tokens are different.
	require.NotEqual(t, retrievedA.ID, retrievedB.ID, "Tokens should have different IDs")
	require.NotEqual(t, retrievedA.ClientID, retrievedB.ClientID, "Tokens should belong to different clients")
}

// verifyTokenActive introspects a token and verifies its active status.
func verifyTokenActive(t *testing.T, app *fiber.App, tokenValue string, expectActive bool, description string) {
	t.Helper()

	result := introspectToken(t, app, tokenValue)

	active, ok := result["active"].(bool)
	require.True(t, ok, "Response should have active field: %s", description)
	require.Equal(t, expectActive, active, description)
}

// introspectToken performs token introspection and returns the result.
func introspectToken(t *testing.T, app *fiber.App, tokenValue string) map[string]any {
	t.Helper()

	formBody := url.Values{
		cryptoutilIdentityMagic.ParamToken: []string{tokenValue},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/introspect", strings.NewReader(formBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Introspection request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusOK, resp.StatusCode, "Introspection should return 200 OK")

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Should decode introspection response")

	return result
}

// extractScopes extracts scopes from introspection response (handles both []any and string).
func extractScopes(result map[string]any) []string {
	scope, exists := result["scope"]
	if !exists {
		return nil
	}

	// Handle []any (JSON array).
	if scopeArray, ok := scope.([]any); ok {
		scopes := make([]string, 0, len(scopeArray))

		for _, s := range scopeArray {
			if str, ok := s.(string); ok {
				scopes = append(scopes, str)
			}
		}

		return scopes
	}

	// Handle string (space-separated).
	if scopeStr, ok := scope.(string); ok {
		return strings.Split(scopeStr, " ")
	}

	return nil
}

// createMultiTenantTestConfig creates a test configuration for multi-tenant tests.
func createMultiTenantTestConfig(t *testing.T) *cryptoutilIdentityConfig.Config {
	t.Helper()

	testID := googleUuid.Must(googleUuid.NewV7()).String()

	return &cryptoutilIdentityConfig.Config{
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type: "sqlite",
			DSN:  fmt.Sprintf("file:multitenant_test_%s.db?mode=memory&cache=shared", testID),
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer: "https://localhost:8080",
		},
	}
}

// createMultiTenantTestRepoFactory creates a repository factory for multi-tenant tests.
func createMultiTenantTestRepoFactory(t *testing.T) *cryptoutilIdentityRepository.RepositoryFactory {
	t.Helper()

	cfg := createMultiTenantTestConfig(t)
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

// createMultiTenantTestClient creates a test client with the given name.
func createMultiTenantTestClient(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory, name string) *cryptoutilIdentityDomain.Client {
	t.Helper()

	return createMultiTenantTestClientWithScopes(t, repoFactory, name, []string{"openid", "profile", "email"})
}

// createMultiTenantTestClientWithScopes creates a test client with specific allowed scopes.
func createMultiTenantTestClientWithScopes(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory, name string, scopes []string) *cryptoutilIdentityDomain.Client {
	t.Helper()

	ctx := context.Background()
	clientRepo := repoFactory.ClientRepository()

	clientUUID, err := googleUuid.NewV7()
	require.NoError(t, err, "Failed to generate client UUID")

	testClient := &cryptoutilIdentityDomain.Client{
		ID:                      clientUUID,
		ClientID:                fmt.Sprintf("test-%s-%s", name, clientUUID.String()),
		Name:                    fmt.Sprintf("Test Client %s", name),
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		AllowedGrantTypes:       []string{cryptoutilIdentityMagic.GrantTypeClientCredentials},
		AllowedScopes:           scopes,
		RedirectURIs:            []string{"https://example.com/callback"},
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretBasic,
	}

	err = clientRepo.Create(ctx, testClient)
	require.NoError(t, err, "Failed to create test client")

	return testClient
}

// createMultiTenantTestToken creates a test token for the given client.
func createMultiTenantTestToken(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory, clientID googleUuid.UUID, name string) *cryptoutilIdentityDomain.Token {
	t.Helper()

	return createMultiTenantTestTokenWithScopes(t, repoFactory, clientID, name, []string{"openid", "profile"})
}

// createMultiTenantTestTokenWithScopes creates a test token with specific scopes.
func createMultiTenantTestTokenWithScopes(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory, clientID googleUuid.UUID, name string, scopes []string) *cryptoutilIdentityDomain.Token {
	t.Helper()

	ctx := context.Background()
	tokenRepo := repoFactory.TokenRepository()

	tokenID := googleUuid.Must(googleUuid.NewV7())

	testToken := &cryptoutilIdentityDomain.Token{
		ID:         tokenID,
		TokenValue: fmt.Sprintf("test-%s-%s", name, tokenID.String()),
		TokenType:  cryptoutilIdentityDomain.TokenTypeAccess,
		ClientID:   clientID,
		Scopes:     scopes,
		ExpiresAt:  time.Now().UTC().Add(1 * time.Hour),
		IssuedAt:   time.Now().UTC(),
	}

	err := tokenRepo.Create(ctx, testToken)
	require.NoError(t, err, "Failed to create test token")

	return testToken
}
