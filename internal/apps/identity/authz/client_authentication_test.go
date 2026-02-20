// Copyright (c) 2025 Justin Cranford
//
//

package authz_test

import (
	"context"
	"encoding/base64"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity/authz"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

func TestClientAuthentication_BasicAuth_InvalidFormat(t *testing.T) {
	t.Parallel()

	config := createClientAuthTestConfig(t)
	repoFactory := createClientAuthTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	reqBody := url.Values{
		"grant_type": []string{"client_credentials"},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/token", strings.NewReader(reqBody.Encode()))
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
		"grant_type": []string{"client_credentials"},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/token", strings.NewReader(reqBody.Encode()))
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
		"grant_type":    []string{"client_credentials"},
		"client_secret": []string{"secret"},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/token", strings.NewReader(reqBody.Encode()))
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
		"grant_type":    []string{"client_credentials"},
		"client_id":     []string{"invalid-client-id"},
		"client_secret": []string{"secret"},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/token", strings.NewReader(reqBody.Encode()))
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
			Type: "sqlite",
			DSN:  "file::memory:?cache=private",
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer: "https://localhost:8080",
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
