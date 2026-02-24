// Copyright (c) 2025 Justin Cranford
//
//

package authz_test

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity/authz"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestHandleIntrospect_TokenTypeHint tests introspect with token_type_hint parameter.
func TestHandleIntrospect_TokenTypeHint(t *testing.T) {
	t.Parallel()

	config := createIntrospectTestConfig(t)
	repoFactory := createIntrospectTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc)

	app := fiber.New()
	svc.RegisterRoutes(app)

	reqBody := url.Values{
		cryptoutilSharedMagic.ParamToken:         []string{"non-existent-token"},
		cryptoutilSharedMagic.ParamTokenTypeHint: []string{"access_token"},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/introspect", strings.NewReader(reqBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	// Non-existent token returns active: false (with or without token_type_hint).
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
}

// TestHandleRevoke_InvalidToken tests revoke with non-existent token (should succeed per spec).
func TestHandleRevoke_InvalidToken(t *testing.T) {
	t.Parallel()

	config := createIntrospectTestConfig(t)
	repoFactory := createIntrospectTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc)

	app := fiber.New()
	svc.RegisterRoutes(app)

	reqBody := url.Values{
		cryptoutilSharedMagic.ParamToken: []string{"non-existent-token"},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/revoke", strings.NewReader(reqBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	// Revoke endpoint always returns 200 OK per OAuth 2.1 spec (even for invalid tokens).
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
}
