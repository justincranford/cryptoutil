// Copyright (c) 2025 Justin Cranford
//
//

package authz_test

import (
	"context"
	json "encoding/json"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/identity/authz"
	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
)

// TestHandleOAuthMetadata validates OAuth 2.1 Authorization Server Metadata endpoint.
func TestHandleOAuthMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		issuer         string
		expectedStatus int
		validateBody   func(t *testing.T, body map[string]any)
	}{
		{
			name:           "returns_valid_oauth_metadata",
			issuer:         "https://auth.example.com",
			expectedStatus: fiber.StatusOK,
			validateBody: func(t *testing.T, body map[string]any) {
				t.Helper()

				// Verify issuer.
				require.Equal(t, "https://auth.example.com", body["issuer"])

				// Verify endpoints.
				require.Equal(t, "https://auth.example.com/oauth2/v1/authorize", body["authorization_endpoint"])
				require.Equal(t, "https://auth.example.com/oauth2/v1/token", body["token_endpoint"])
				require.Equal(t, "https://auth.example.com/oauth2/v1/introspect", body["introspection_endpoint"])
				require.Equal(t, "https://auth.example.com/oauth2/v1/revoke", body["revocation_endpoint"])
				require.Equal(t, "https://auth.example.com/oauth2/v1/jwks", body["jwks_uri"])

				// Verify grant types.
				grantTypes, ok := body["grant_types_supported"].([]any)
				require.True(t, ok, "grant_types_supported should be an array")
				require.Contains(t, grantTypes, "authorization_code")
				require.Contains(t, grantTypes, "refresh_token")
				require.Contains(t, grantTypes, "client_credentials")

				// Verify response types.
				responseTypes, ok := body["response_types_supported"].([]any)
				require.True(t, ok, "response_types_supported should be an array")
				require.Contains(t, responseTypes, "code")

				// Verify PKCE support.
				codeChallenges, ok := body["code_challenge_methods_supported"].([]any)
				require.True(t, ok, "code_challenge_methods_supported should be an array")
				require.Contains(t, codeChallenges, "S256")

				// Verify token auth methods.
				authMethods, ok := body["token_endpoint_auth_methods_supported"].([]any)
				require.True(t, ok, "token_endpoint_auth_methods_supported should be an array")
				require.Contains(t, authMethods, "client_secret_post")
				require.Contains(t, authMethods, "client_secret_basic")
			},
		},
		{
			name:           "returns_valid_oauth_metadata_with_different_issuer",
			issuer:         "https://id.company.io",
			expectedStatus: fiber.StatusOK,
			validateBody: func(t *testing.T, body map[string]any) {
				t.Helper()

				require.Equal(t, "https://id.company.io", body["issuer"])
				require.Equal(t, "https://id.company.io/oauth2/v1/authorize", body["authorization_endpoint"])
				require.Equal(t, "https://id.company.io/oauth2/v1/token", body["token_endpoint"])
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create service with test config.
			config := createDiscoveryTestConfig(t, tc.issuer)
			repoFactory := createDiscoveryTestRepoFactory(t, config)

			svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
			require.NotNil(t, svc, "Service should not be nil")

			app := fiber.New()
			svc.RegisterRoutes(app)

			// Create test request.
			req := httptest.NewRequest(fiber.MethodGet, "/.well-known/oauth-authorization-server", nil)

			// Execute request.
			resp, err := app.Test(req)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

			// Validate status.
			require.Equal(t, tc.expectedStatus, resp.StatusCode)

			// Parse and validate body.
			var body map[string]any

			err = json.NewDecoder(resp.Body).Decode(&body)
			require.NoError(t, err)

			tc.validateBody(t, body)
		})
	}
}

// TestHandleOIDCDiscovery validates OpenID Connect Discovery endpoint.
func TestHandleOIDCDiscovery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		issuer         string
		expectedStatus int
		validateBody   func(t *testing.T, body map[string]any)
	}{
		{
			name:           "returns_valid_oidc_discovery",
			issuer:         "https://oidc.example.com",
			expectedStatus: fiber.StatusOK,
			validateBody: func(t *testing.T, body map[string]any) {
				t.Helper()

				// Verify issuer.
				require.Equal(t, "https://oidc.example.com", body["issuer"])

				// Verify OIDC-specific endpoints.
				require.Equal(t, "https://oidc.example.com/oauth2/v1/authorize", body["authorization_endpoint"])
				require.Equal(t, "https://oidc.example.com/oauth2/v1/token", body["token_endpoint"])
				require.Equal(t, "https://oidc.example.com/oauth2/v1/userinfo", body["userinfo_endpoint"])
				require.Equal(t, "https://oidc.example.com/oauth2/v1/jwks", body["jwks_uri"])

				// Verify subject types.
				subjectTypes, ok := body["subject_types_supported"].([]any)
				require.True(t, ok, "subject_types_supported should be an array")
				require.Contains(t, subjectTypes, "public")

				// Verify ID token signing algorithms.
				signingAlgs, ok := body["id_token_signing_alg_values_supported"].([]any)
				require.True(t, ok, "id_token_signing_alg_values_supported should be an array")
				require.Contains(t, signingAlgs, "RS256")
				require.Contains(t, signingAlgs, "ES256")

				// Verify OIDC scopes.
				scopes, ok := body["scopes_supported"].([]any)
				require.True(t, ok, "scopes_supported should be an array")
				require.Contains(t, scopes, "openid")
				require.Contains(t, scopes, "profile")
				require.Contains(t, scopes, "email")
				require.Contains(t, scopes, "address")
				require.Contains(t, scopes, "phone")

				// Verify OIDC claims.
				claims, ok := body["claims_supported"].([]any)
				require.True(t, ok, "claims_supported should be an array")
				require.Contains(t, claims, "sub")
				require.Contains(t, claims, "iss")
				require.Contains(t, claims, "name")
				require.Contains(t, claims, "email")
				require.Contains(t, claims, "email_verified")
			},
		},
		{
			name:           "returns_correct_response_types_for_oidc",
			issuer:         "https://oidc.test.io",
			expectedStatus: fiber.StatusOK,
			validateBody: func(t *testing.T, body map[string]any) {
				t.Helper()

				// OIDC Discovery includes additional response types.
				responseTypes, ok := body["response_types_supported"].([]any)
				require.True(t, ok, "response_types_supported should be an array")
				require.Contains(t, responseTypes, "code")
				require.Contains(t, responseTypes, "id_token")
				require.Contains(t, responseTypes, "token id_token")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create service with test config.
			config := createDiscoveryTestConfig(t, tc.issuer)
			repoFactory := createDiscoveryTestRepoFactory(t, config)

			svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
			require.NotNil(t, svc, "Service should not be nil")

			app := fiber.New()
			svc.RegisterRoutes(app)

			// Create test request.
			req := httptest.NewRequest(fiber.MethodGet, "/.well-known/openid-configuration", nil)

			// Execute request.
			resp, err := app.Test(req)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

			// Validate status.
			require.Equal(t, tc.expectedStatus, resp.StatusCode)

			// Parse and validate body.
			var body map[string]any

			err = json.NewDecoder(resp.Body).Decode(&body)
			require.NoError(t, err)

			tc.validateBody(t, body)
		})
	}
}

// TestHandleJWKS validates JSON Web Key Set endpoint.
func TestHandleJWKS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		expectedStatus int
		validateBody   func(t *testing.T, body map[string]any)
	}{
		{
			name:           "returns_valid_jwks",
			expectedStatus: fiber.StatusOK,
			validateBody: func(t *testing.T, body map[string]any) {
				t.Helper()

				// Verify keys array exists.
				keys, ok := body["keys"]
				require.True(t, ok, "JWKS should have 'keys' field")
				require.NotNil(t, keys, "keys should not be nil")

				// Keys should be an array (may be empty if no token service configured).
				keysArray, ok := keys.([]any)
				require.True(t, ok, "keys should be an array")

				// If keys exist, validate structure.
				if len(keysArray) > 0 {
					firstKey, keyOk := keysArray[0].(map[string]any)
					require.True(t, keyOk, "key should be an object")

					// Check required JWK fields.
					require.Contains(t, firstKey, "kty", "key should have 'kty' field")
					require.Contains(t, firstKey, "kid", "key should have 'kid' field")
					require.Contains(t, firstKey, "use", "key should have 'use' field")
					require.Equal(t, "sig", firstKey["use"], "key use should be 'sig'")
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create service with test config.
			config := createDiscoveryTestConfig(t, "https://jwks.example.com")
			repoFactory := createDiscoveryTestRepoFactory(t, config)

			svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
			require.NotNil(t, svc, "Service should not be nil")

			app := fiber.New()
			svc.RegisterRoutes(app)

			// Create test request.
			req := httptest.NewRequest(fiber.MethodGet, "/oauth2/v1/jwks", nil)

			// Execute request.
			resp, err := app.Test(req)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

			// Validate status.
			require.Equal(t, tc.expectedStatus, resp.StatusCode)

			// Parse and validate body.
			var body map[string]any

			err = json.NewDecoder(resp.Body).Decode(&body)
			require.NoError(t, err)

			tc.validateBody(t, body)
		})
	}
}

// createDiscoveryTestConfig creates a test Config with specified issuer.
func createDiscoveryTestConfig(t *testing.T, issuer string) *cryptoutilIdentityConfig.Config {
	t.Helper()

	return &cryptoutilIdentityConfig.Config{
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type: "sqlite",
			DSN:  "file::memory:?cache=private",
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer: issuer,
		},
	}
}

// createDiscoveryTestRepoFactory creates a test RepositoryFactory.
func createDiscoveryTestRepoFactory(t *testing.T, cfg *cryptoutilIdentityConfig.Config) *cryptoutilIdentityRepository.RepositoryFactory {
	t.Helper()

	ctx := context.Background()

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, cfg.Database)
	require.NoError(t, err, "Failed to create repository factory")
	require.NotNil(t, repoFactory, "Repository factory should not be nil")

	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err, "Failed to run auto migrations")

	return repoFactory
}
