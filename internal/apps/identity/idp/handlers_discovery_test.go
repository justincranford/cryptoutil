// Copyright (c) 2025 Justin Cranford
//
//

package idp

import (
	"context"
	json "encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestDiscoveryHandler_ReturnsValidOIDCMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		requestProtocol  string
		requestHost      string
		expectedIssuer   string
		expectedEndpoint string
	}{
		{
			name:             "HTTPS request",
			requestProtocol:  cryptoutilSharedMagic.ProtocolHTTPS,
			requestHost:      "identity.example.com",
			expectedIssuer:   "https://identity.example.com/oidc/v1/",
			expectedEndpoint: "https://identity.example.com/authz/v1/authorize",
		},
		{
			name:             "HTTP request (development)",
			requestProtocol:  cryptoutilSharedMagic.ProtocolHTTP,
			requestHost:      "localhost:8080",
			expectedIssuer:   "http://localhost:8080/oidc/v1/",
			expectedEndpoint: "http://localhost:8080/authz/v1/authorize",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create test service with minimal config.
			config := &cryptoutilIdentityConfig.Config{
				Database: &cryptoutilIdentityConfig.DatabaseConfig{
					Type: cryptoutilSharedMagic.TestDatabaseSQLite,
					DSN:  cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
				},
				Tokens: &cryptoutilIdentityConfig.TokenConfig{
					AccessTokenLifetime: cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
				},
			}

			repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(context.Background(), config.Database)
			require.NoError(t, err)

			defer func() {
				_ = repoFactory.Close() //nolint:errcheck // Test cleanup
			}()

			tokenSvc := cryptoutilIdentityIssuer.NewTokenService(nil, nil, nil, config.Tokens)
			testService := NewService(config, repoFactory, tokenSvc)

			// Create test Fiber app.
			app := fiber.New()
			testService.RegisterRoutes(app)

			// Make request to discovery endpoint.
			req := httptest.NewRequest("GET", "https://"+tc.requestHost+cryptoutilSharedMagic.PathDiscovery, nil)

			// Set protocol via X-Forwarded-Proto header (standard proxy header).
			req.Header.Set("X-Forwarded-Proto", tc.requestProtocol)

			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			require.Equal(t, fiber.StatusOK, resp.StatusCode)

			defer func() {
				_ = resp.Body.Close() //nolint:errcheck // Test cleanup
			}()

			// Parse response body as DiscoveryMetadata.
			bodyBytes, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var metadata DiscoveryMetadata

			err = json.Unmarshal(bodyBytes, &metadata)
			require.NoError(t, err)

			// Verify required OIDC fields.
			require.Contains(t, metadata.Issuer, tc.expectedIssuer, "Issuer should match expected prefix")
			require.Equal(t, tc.expectedEndpoint, metadata.AuthorizationEndpoint)
			require.NotEmpty(t, metadata.TokenEndpoint)
			require.NotEmpty(t, metadata.UserInfoEndpoint)
			require.NotEmpty(t, metadata.JWKSUri)

			// Verify supported scopes include OIDC standard scopes.
			require.Contains(t, metadata.ScopesSupported, cryptoutilSharedMagic.ScopeOpenID)
			require.Contains(t, metadata.ScopesSupported, cryptoutilSharedMagic.ClaimProfile)
			require.Contains(t, metadata.ScopesSupported, cryptoutilSharedMagic.ClaimEmail)
			require.Contains(t, metadata.ScopesSupported, cryptoutilSharedMagic.ScopeOfflineAccess)

			// Verify response types (OAuth 2.1 authorization code only).
			require.Contains(t, metadata.ResponseTypesSupported, cryptoutilSharedMagic.ResponseTypeCode)

			// Verify grant types (OAuth 2.1).
			require.Contains(t, metadata.GrantTypesSupported, cryptoutilSharedMagic.GrantTypeAuthorizationCode)
			require.Contains(t, metadata.GrantTypesSupported, cryptoutilSharedMagic.GrantTypeRefreshToken)
			require.Contains(t, metadata.GrantTypesSupported, cryptoutilSharedMagic.GrantTypeClientCredentials)

			// Verify subject types (OIDC).
			require.Contains(t, metadata.SubjectTypesSupported, cryptoutilSharedMagic.SubjectTypePublic)

			// Verify signing algorithms (FIPS 140-3 approved).
			require.Contains(t, metadata.IDTokenSigningAlgValuesSupported, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm)
			require.Contains(t, metadata.IDTokenSigningAlgValuesSupported, cryptoutilSharedMagic.JoseAlgES256)
			require.Contains(t, metadata.IDTokenSigningAlgValuesSupported, cryptoutilSharedMagic.JoseAlgEdDSA)

			// Verify token endpoint auth methods (OAuth 2.1).
			require.Contains(t, metadata.TokenEndpointAuthMethodsSupported, cryptoutilSharedMagic.ClientAuthMethodSecretBasic)
			require.Contains(t, metadata.TokenEndpointAuthMethodsSupported, cryptoutilSharedMagic.ClientAuthMethodPrivateKeyJWT)
			require.Contains(t, metadata.TokenEndpointAuthMethodsSupported, cryptoutilSharedMagic.ClientAuthMethodTLSClientAuth)

			// Verify PKCE support (OAuth 2.1 required).
			require.Contains(t, metadata.CodeChallengeMethodsSupported, cryptoutilSharedMagic.PKCEMethodS256)

			// Verify claims supported (OIDC standard claims).
			require.Contains(t, metadata.ClaimsSupported, cryptoutilSharedMagic.ClaimSub)
			require.Contains(t, metadata.ClaimsSupported, cryptoutilSharedMagic.ClaimEmail)
			require.Contains(t, metadata.ClaimsSupported, cryptoutilSharedMagic.ClaimEmailVerified)
			require.Contains(t, metadata.ClaimsSupported, cryptoutilSharedMagic.ClaimName)

			// Verify revocation endpoint (OAuth 2.1).
			require.NotEmpty(t, metadata.RevocationEndpoint)

			// Verify introspection endpoint (OAuth 2.1).
			require.NotEmpty(t, metadata.IntrospectionEndpoint)
		})
	}
}

func TestDiscoveryHandler_HTTPSByDefault(t *testing.T) {
	t.Parallel()

	config := &cryptoutilIdentityConfig.Config{
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type: cryptoutilSharedMagic.TestDatabaseSQLite,
			DSN:  cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			AccessTokenLifetime: cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
		},
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(context.Background(), config.Database)
	require.NoError(t, err)

	defer func() {
		_ = repoFactory.Close() //nolint:errcheck // Test cleanup
	}()

	tokenSvc := cryptoutilIdentityIssuer.NewTokenService(nil, nil, nil, config.Tokens)
	testService := NewService(config, repoFactory, tokenSvc)

	app := fiber.New()
	testService.RegisterRoutes(app)

	req := httptest.NewRequest("GET", "https://identity.example.com/.well-known/openid-configuration", nil)

	// Explicitly set X-Forwarded-Proto to test HTTPS default behavior.
	req.Header.Set("X-Forwarded-Proto", cryptoutilSharedMagic.ProtocolHTTPS)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	defer func() {
		_ = resp.Body.Close() //nolint:errcheck // Test cleanup
	}()

	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var metadata DiscoveryMetadata

	err = json.Unmarshal(bodyBytes, &metadata)
	require.NoError(t, err)

	require.Contains(t, metadata.Issuer, "https://", "Should default to HTTPS")
	require.Contains(t, metadata.AuthorizationEndpoint, "https://", "Should default to HTTPS")
}

func TestDiscoveryHandler_AllRequiredOIDCFieldsPresent(t *testing.T) {
	t.Parallel()

	config := &cryptoutilIdentityConfig.Config{
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type: cryptoutilSharedMagic.TestDatabaseSQLite,
			DSN:  cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			AccessTokenLifetime: cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
		},
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(context.Background(), config.Database)
	require.NoError(t, err)

	defer func() {
		_ = repoFactory.Close() //nolint:errcheck // Test cleanup
	}()

	tokenSvc := cryptoutilIdentityIssuer.NewTokenService(nil, nil, nil, config.Tokens)
	testService := NewService(config, repoFactory, tokenSvc)

	app := fiber.New()
	testService.RegisterRoutes(app)

	req := httptest.NewRequest("GET", "https://identity.example.com/.well-known/openid-configuration", nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() {
		_ = resp.Body.Close() //nolint:errcheck // Test cleanup
	}()

	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var metadata DiscoveryMetadata

	err = json.Unmarshal(bodyBytes, &metadata)
	require.NoError(t, err)

	// Verify all REQUIRED OIDC discovery fields per https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderMetadata.
	require.NotEmpty(t, metadata.Issuer, "issuer REQUIRED")
	require.NotEmpty(t, metadata.AuthorizationEndpoint, "authorization_endpoint REQUIRED")
	require.NotEmpty(t, metadata.TokenEndpoint, "token_endpoint REQUIRED")
	require.NotEmpty(t, metadata.UserInfoEndpoint, "userinfo_endpoint REQUIRED")
	require.NotEmpty(t, metadata.JWKSUri, "jwks_uri REQUIRED")
	require.NotEmpty(t, metadata.ScopesSupported, "scopes_supported RECOMMENDED")
	require.NotEmpty(t, metadata.ResponseTypesSupported, "response_types_supported REQUIRED")
	require.NotEmpty(t, metadata.SubjectTypesSupported, "subject_types_supported REQUIRED")
	require.NotEmpty(t, metadata.IDTokenSigningAlgValuesSupported, "id_token_signing_alg_values_supported REQUIRED")
}
