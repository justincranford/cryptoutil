// Copyright (c) 2025 Justin Cranford
//
//

package authz_test

import (
	"context"
	"net/http/httptest"
	"net/url"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity/authz"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestHandleAuthorizeGET_MissingClientID(t *testing.T) {
	t.Parallel()

	config := createAuthorizeTestConfig(t)
	repoFactory := createAuthorizeTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	query := url.Values{
		cryptoutilSharedMagic.ParamResponseType: []string{cryptoutilSharedMagic.ResponseTypeCode},
		cryptoutilSharedMagic.ParamRedirectURI:  []string{"https://example.com/callback"},
		cryptoutilSharedMagic.ParamScope:        []string{"openid profile"},
		cryptoutilSharedMagic.ParamState:        []string{"test-state"},
	}

	req := httptest.NewRequest("GET", "/oauth2/v1/authorize?"+query.Encode(), nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() {
		_ = resp.Body.Close() //nolint:errcheck // Test cleanup
	}()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should return 400 Bad Request")
}

func TestHandleAuthorizeGET_MissingResponseType(t *testing.T) {
	t.Parallel()

	config := createAuthorizeTestConfig(t)
	repoFactory := createAuthorizeTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	query := url.Values{
		cryptoutilSharedMagic.ParamClientID:    []string{"test-client"},
		cryptoutilSharedMagic.ParamRedirectURI: []string{"https://example.com/callback"},
		cryptoutilSharedMagic.ParamScope:       []string{"openid profile"},
		cryptoutilSharedMagic.ParamState:       []string{"test-state"},
	}

	req := httptest.NewRequest("GET", "/oauth2/v1/authorize?"+query.Encode(), nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() {
		_ = resp.Body.Close() //nolint:errcheck // Test cleanup
	}()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should return 400 Bad Request")
}

func TestHandleAuthorizeGET_MissingRedirectURI(t *testing.T) {
	t.Parallel()

	config := createAuthorizeTestConfig(t)
	repoFactory := createAuthorizeTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	query := url.Values{
		cryptoutilSharedMagic.ParamClientID:     []string{"test-client"},
		cryptoutilSharedMagic.ParamResponseType: []string{cryptoutilSharedMagic.ResponseTypeCode},
		cryptoutilSharedMagic.ParamScope:        []string{"openid profile"},
		cryptoutilSharedMagic.ParamState:        []string{"test-state"},
	}

	req := httptest.NewRequest("GET", "/oauth2/v1/authorize?"+query.Encode(), nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should return 400 Bad Request")
}

func TestHandleAuthorizePOST_MissingClientID(t *testing.T) {
	t.Parallel()

	config := createAuthorizeTestConfig(t)
	repoFactory := createAuthorizeTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	reqBody := url.Values{}

	req := httptest.NewRequest("POST", "/oauth2/v1/authorize", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	for key, values := range reqBody {
		for _, value := range values {
			req.Form.Set(key, value)
		}
	}

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should return 400 Bad Request")
}

func createAuthorizeTestConfig(t *testing.T) *cryptoutilIdentityConfig.Config {
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

func createAuthorizeTestRepoFactory(t *testing.T) *cryptoutilIdentityRepository.RepositoryFactory {
	t.Helper()

	cfg := createAuthorizeTestConfig(t)
	ctx := context.Background()

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, cfg.Database)
	require.NoError(t, err, "Failed to create repository factory")
	require.NotNil(t, repoFactory, "Repository factory should not be nil")

	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err, "Failed to run auto migrations")

	return repoFactory
}
