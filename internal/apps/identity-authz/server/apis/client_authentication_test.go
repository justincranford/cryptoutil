// Copyright (c) 2025-2026 Justin Cranford.
//
//

package apis_test

import (
	"context"
	"encoding/base64"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity-authz/server/apis"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

func TestClientAuthentication_FallbackErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		clientID         string
		authHeader       string
		clientAuthMethod cryptoutilIdentityDomain.ClientAuthMethod
	}{
		{
			name:       "basic_auth_without_client_secret_separator",
			authHeader: "Basic " + base64.StdEncoding.EncodeToString([]byte("malformed")),
		},
		{
			name:     "unknown_client_id_without_secret",
			clientID: "unregistered-client",
		},
		{
			name:             "basic_client_id_without_secret",
			clientAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretBasic,
		},
		{
			name:             "mtls_client_without_tls_connection",
			clientAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodTLSClientAuth,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config := createClientAuthFlowTestConfig(t)
			repoFactory := createClientAuthFlowTestRepoFactory(t, config)

			clientID := tc.clientID
			if tc.clientAuthMethod != "" {
				client := createClientAuthFlowTestClient(t, repoFactory, tc.clientAuthMethod)
				clientID = client.ClientID
			}

			service := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
			app := fiber.New()
			service.RegisterRoutes(app)

			form := url.Values{
				cryptoutilSharedMagic.ParamGrantType: []string{cryptoutilSharedMagic.GrantTypeClientCredentials},
			}
			if clientID != "" {
				form.Set(cryptoutilSharedMagic.ParamClientID, clientID)
			}

			request := httptest.NewRequest(testPost, testOauth2V1Token, strings.NewReader(form.Encode()))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			if tc.authHeader != "" {
				request.Header.Set("Authorization", tc.authHeader)
			}

			response, err := app.Test(request, -1)

			require.NoError(t, err)
			defer func() {
				require.NoError(t, response.Body.Close())
			}()

			require.Equal(t, fiber.StatusUnauthorized, response.StatusCode)
		})
	}
}

func TestClientAuthentication_BasicAuth_InvalidFormat(t *testing.T) {
	t.Parallel()

	config := createClientAuthTestConfig(t)
	repoFactory := createClientAuthTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	reqBody := url.Values{
		cryptoutilSharedMagic.ParamGrantType: []string{cryptoutilSharedMagic.GrantTypeClientCredentials},
	}

	req := httptest.NewRequest(testPost, testOauth2V1Token, strings.NewReader(reqBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic invalid-base64")

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() {
		_ = resp.Body.Close() //nolint:errcheck // Test cleanup
	}()

	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode, "Should return 401 Unauthorized")
}

func TestClientAuthentication_BasicAuth_InvalidClientID(t *testing.T) {
	t.Parallel()

	config := createClientAuthTestConfig(t)
	repoFactory := createClientAuthTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	// Encode invalid-client-id:secret.
	credentials := "invalid-client-id:secret"
	encoded := base64.StdEncoding.EncodeToString([]byte(credentials))

	reqBody := url.Values{
		cryptoutilSharedMagic.ParamGrantType: []string{cryptoutilSharedMagic.GrantTypeClientCredentials},
	}

	req := httptest.NewRequest(testPost, testOauth2V1Token, strings.NewReader(reqBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+encoded)

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() {
		_ = resp.Body.Close() //nolint:errcheck // Test cleanup
	}()

	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode, "Should return 401 Unauthorized")
}

func TestClientAuthentication_PostAuth_MissingClientID(t *testing.T) {
	t.Parallel()

	config := createClientAuthTestConfig(t)
	repoFactory := createClientAuthTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	reqBody := url.Values{
		cryptoutilSharedMagic.ParamGrantType:    []string{cryptoutilSharedMagic.GrantTypeClientCredentials},
		cryptoutilSharedMagic.ParamClientSecret: []string{testSecretValue2BB80D53},
	}

	req := httptest.NewRequest(testPost, testOauth2V1Token, strings.NewReader(reqBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() {
		_ = resp.Body.Close() //nolint:errcheck // Test cleanup
	}()

	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode, "Should return 401 Unauthorized")
}

func TestClientAuthentication_PostAuth_InvalidClientID(t *testing.T) {
	t.Parallel()

	config := createClientAuthTestConfig(t)
	repoFactory := createClientAuthTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	reqBody := url.Values{
		cryptoutilSharedMagic.ParamGrantType:    []string{cryptoutilSharedMagic.GrantTypeClientCredentials},
		cryptoutilSharedMagic.ClaimClientID:     []string{"invalid-client-id"},
		cryptoutilSharedMagic.ParamClientSecret: []string{testSecretValue2BB80D53},
	}

	req := httptest.NewRequest(testPost, testOauth2V1Token, strings.NewReader(reqBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() {
		_ = resp.Body.Close() //nolint:errcheck // Test cleanup
	}()

	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode, "Should return 401 Unauthorized")
}

func createClientAuthTestConfig(t *testing.T) *cryptoutilIdentityConfig.Config {
	t.Helper()

	return &cryptoutilIdentityConfig.Config{
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type: cryptoutilSharedMagic.TestDatabaseSQLite,
			DSN:  testFileMemoryCachePrivate,
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer: testHTTPSLocalhost8080,
		},
	}
}

func createClientAuthTestRepoFactory(t *testing.T) *cryptoutilIdentityRepository.RepositoryFactory {
	t.Helper()

	cfg := createClientAuthTestConfig(t)
	ctx := context.Background()

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, cfg.Database)
	require.NoError(t, err, "Failed to create repository factory")
	require.NotNil(t, repoFactory, "Repository factory should not be nil")

	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err, "Failed to run auto migrations")

	return repoFactory
}
