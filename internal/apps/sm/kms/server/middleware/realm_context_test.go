// Copyright (c) 2025 Justin Cranford
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package middleware

import (
	"context"
	"io"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRealmContextMiddleware_FromJWT(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	userID := googleUuid.New()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	// Middleware to set up JWT claims before RealmContextMiddleware
	app.Use(func(c *fiber.Ctx) error {
		claims := &JWTClaims{
			Subject: userID.String(),
			Scopes:  []string{"kms:read", "kms:write"},
			Custom: map[string]any{
				"tenant_id": tenantID.String(),
				"realm_id":  realmID.String(),
			},
		}
		ctx := context.WithValue(c.UserContext(), JWTContextKey{}, claims)
		c.SetUserContext(ctx)

		return c.Next()
	})

	app.Use(RealmContextMiddleware())

	app.Get("/test", func(c *fiber.Ctx) error {
		realmCtx := GetRealmContext(c.UserContext())
		if realmCtx == nil {
			return c.Status(500).SendString("no realm context")
		}

		return c.JSON(fiber.Map{
			"tenant_id": realmCtx.TenantID.String(),
			"realm_id":  realmCtx.RealmID.String(),
			"user_id":   realmCtx.UserID.String(),
			"source":    realmCtx.Source,
			"scopes":    realmCtx.Scopes,
		})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, 200, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Contains(t, string(body), tenantID.String())
	require.Contains(t, string(body), realmID.String())
	require.Contains(t, string(body), userID.String())
	require.Contains(t, string(body), `"source":"jwt"`)
}

func TestRealmContextMiddleware_FromOIDC(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	// Middleware to set up OIDC claims
	app.Use(func(c *fiber.Ctx) error {
		claims := &OIDCClaims{
			TenantID: tenantID.String(),
		}
		ctx := context.WithValue(c.UserContext(), OIDCClaimsContextKey{}, claims)
		c.SetUserContext(ctx)

		return c.Next()
	})

	app.Use(RealmContextMiddleware())

	app.Get("/test", func(c *fiber.Ctx) error {
		realmCtx := GetRealmContext(c.UserContext())
		if realmCtx == nil {
			return c.Status(500).SendString("no realm context")
		}

		return c.JSON(fiber.Map{
			"tenant_id": realmCtx.TenantID.String(),
			"source":    realmCtx.Source,
		})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, 200, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Contains(t, string(body), tenantID.String())
	require.Contains(t, string(body), `"source":"oidc"`)
}

func TestRealmContextMiddleware_FromHeader(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	// Middleware to set up tenant from header (TenantMiddleware would do this)
	app.Use(func(c *fiber.Ctx) error {
		ctx := context.WithValue(c.UserContext(), TenantContextKey{}, tenantID.String())
		c.SetUserContext(ctx)

		return c.Next()
	})

	app.Use(RealmContextMiddleware())

	app.Get("/test", func(c *fiber.Ctx) error {
		realmCtx := GetRealmContext(c.UserContext())
		if realmCtx == nil {
			return c.Status(500).SendString("no realm context")
		}

		return c.JSON(fiber.Map{
			"tenant_id": realmCtx.TenantID.String(),
			"source":    realmCtx.Source,
		})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, 200, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Contains(t, string(body), tenantID.String())
	require.Contains(t, string(body), `"source":"header"`)
}

func TestRequireRealmMiddleware_NoTenant(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	app.Use(RealmContextMiddleware())
	app.Use(RequireRealmMiddleware())

	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, 401, resp.StatusCode)
}

func TestRequireRealmMiddleware_WithTenant(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	// Set up JWT claims with tenant
	app.Use(func(c *fiber.Ctx) error {
		claims := &JWTClaims{
			Custom: map[string]any{
				"tenant_id": tenantID.String(),
			},
		}
		ctx := context.WithValue(c.UserContext(), JWTContextKey{}, claims)
		c.SetUserContext(ctx)

		return c.Next()
	})

	app.Use(RealmContextMiddleware())
	app.Use(RequireRealmMiddleware())

	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, 200, resp.StatusCode)
}

func TestRealmContextMiddleware_FromJWT_WithUserAndClient(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	userID := googleUuid.New()
	clientID := googleUuid.New()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	// Middleware to set JWT claims with user_id and client_id in custom claims.
	app.Use(func(c *fiber.Ctx) error {
		claims := &JWTClaims{
			Scopes: []string{"kms:read"},
			Custom: map[string]any{
				"tenant_id": tenantID.String(),
				"realm_id":  realmID.String(),
				"user_id":   userID.String(),
				"client_id": clientID.String(),
			},
		}
		ctx := context.WithValue(c.UserContext(), JWTContextKey{}, claims)
		c.SetUserContext(ctx)

		return c.Next()
	})

	app.Use(RealmContextMiddleware())

	app.Get("/test", func(c *fiber.Ctx) error {
		realmCtx := GetRealmContext(c.UserContext())
		if realmCtx == nil {
			return c.Status(500).SendString("no realm context")
		}

		return c.JSON(fiber.Map{
			"tenant_id": realmCtx.TenantID.String(),
			"realm_id":  realmCtx.RealmID.String(),
			"user_id":   realmCtx.UserID.String(),
			"client_id": realmCtx.ClientID.String(),
			"source":    realmCtx.Source,
		})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, 200, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Contains(t, string(body), userID.String())
	require.Contains(t, string(body), clientID.String())
	require.Contains(t, string(body), tenantID.String())
	require.Contains(t, string(body), realmID.String())
}

func TestRealmContextMiddleware_FromOIDC_WithTenantIDs(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	// Middleware to set OIDC claims with TenantIDs array (not TenantID string).
	app.Use(func(c *fiber.Ctx) error {
		claims := &OIDCClaims{
			TenantIDs: []string{tenantID.String()},
		}
		ctx := context.WithValue(c.UserContext(), OIDCClaimsContextKey{}, claims)
		c.SetUserContext(ctx)

		return c.Next()
	})

	app.Use(RealmContextMiddleware())

	app.Get("/test", func(c *fiber.Ctx) error {
		realmCtx := GetRealmContext(c.UserContext())
		if realmCtx == nil {
			return c.Status(500).SendString("no realm context")
		}

		return c.JSON(fiber.Map{
			"tenant_id": realmCtx.TenantID.String(),
			"source":    realmCtx.Source,
		})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, 200, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Contains(t, string(body), tenantID.String())
	require.Contains(t, string(body), `"source":"oidc"`)
}

func TestGetRealmContext_NilContext(t *testing.T) {
	t.Parallel()

	result := GetRealmContext(nil) //nolint:staticcheck // Testing nil context handling explicitly
	require.Nil(t, result)
}

func TestGetRealmContext_NoValue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	result := GetRealmContext(ctx)
	require.Nil(t, result)
}
