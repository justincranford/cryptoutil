package middleware

import (
	"context"
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestServiceAuth_ClientCredentials(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		config         ServiceAuthConfig
		authHeader     string
		expectedStatus int
	}{
		{
			name: "no client credentials config",
			config: ServiceAuthConfig{
				AllowedMethods: []AuthMethod{AuthMethodClientCredentials},
			},
			authHeader:     "Bearer some-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "not fully implemented",
			config: ServiceAuthConfig{
				AllowedMethods: []AuthMethod{AuthMethodClientCredentials},
				ClientCredentialsConfig: &ClientCredentialsConfig{
					TokenEndpoint: "https://auth.example.com/token",
				},
			},
			authHeader:     "Bearer some-token",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mw, err := NewServiceAuthMiddleware(tc.config)
			require.NoError(t, err)

			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Use(mw.Middleware())
			app.Get("/test", func(c *fiber.Ctx) error {
				return c.SendStatus(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}

			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

func TestServiceAuth_Helpers(t *testing.T) {
	t.Parallel()

	t.Run("GetServiceAuthInfo from empty context", func(t *testing.T) {
		t.Parallel()

		info := GetServiceAuthInfo(context.Background())
		require.Nil(t, info)
	})

	t.Run("GetServiceAuthInfo with value", func(t *testing.T) {
		t.Parallel()

		expected := &ServiceAuthInfo{
			Method:      AuthMethodAPIKey,
			ServiceName: "test-svc",
		}

		ctx := context.WithValue(context.Background(), ServiceAuthContextKey{}, expected)
		info := GetServiceAuthInfo(ctx)
		require.NotNil(t, info)
		require.Equal(t, expected.ServiceName, info.ServiceName)
	})

	t.Run("RequireServiceAuth delegates to Middleware", func(t *testing.T) {
		t.Parallel()

		mw, err := NewServiceAuthMiddleware(ServiceAuthConfig{
			AllowedMethods: []AuthMethod{AuthMethodAPIKey},
			APIKeyConfig: &APIKeyConfig{
				ValidKeys: map[string]string{"key1": "svc1"},
			},
		})
		require.NoError(t, err)

		handler := RequireServiceAuth(mw)
		require.NotNil(t, handler)
	})

	t.Run("ConfigureTLSForMTLS require cert", func(t *testing.T) {
		t.Parallel()

		config := ConfigureTLSForMTLS(true)
		require.NotNil(t, config)
	})

	t.Run("ConfigureTLSForMTLS optional cert", func(t *testing.T) {
		t.Parallel()

		config := ConfigureTLSForMTLS(false)
		require.NotNil(t, config)
	})
}

func TestServiceAuth_MTLSNilConfig(t *testing.T) {
	t.Parallel()

	// mTLS with nil config returns unauthorized.
	mw, err := NewServiceAuthMiddleware(ServiceAuthConfig{
		AllowedMethods: []AuthMethod{AuthMethodMTLS},
	})
	require.NoError(t, err)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(mw.Middleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestServiceAuth_APIKeyNilKeys(t *testing.T) {
	t.Parallel()

	mw, err := NewServiceAuthMiddleware(ServiceAuthConfig{
		AllowedMethods: []AuthMethod{AuthMethodAPIKey},
		APIKeyConfig:   &APIKeyConfig{},
	})
	require.NoError(t, err)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(mw.Middleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "some-key")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestServiceAuth_APIKeyNilConfig(t *testing.T) {
	t.Parallel()

	mw, err := NewServiceAuthMiddleware(ServiceAuthConfig{
		AllowedMethods: []AuthMethod{AuthMethodAPIKey},
	})
	require.NoError(t, err)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(mw.Middleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "some-key")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestTryAuthenticate_UnsupportedMethod(t *testing.T) {
	t.Parallel()

	mw, err := NewServiceAuthMiddleware(ServiceAuthConfig{
		AllowedMethods: []AuthMethod{"unsupported"},
	})
	require.NoError(t, err)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(mw.Middleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestServiceAuth_ClientCredentialsWithJWT(t *testing.T) {
	t.Parallel()

	jwksServer := newTestJWKSServer(t)
	now := time.Now().UTC()

	mw, err := NewServiceAuthMiddleware(ServiceAuthConfig{
		AllowedMethods:          []AuthMethod{AuthMethodClientCredentials, AuthMethodJWT},
		JWTConfig:               &JWTValidatorConfig{JWKSURL: jwksServer.server.URL},
		ClientCredentialsConfig: &ClientCredentialsConfig{},
	})
	require.NoError(t, err)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(mw.Middleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	tokenString := jwksServer.signToken(t, map[string]any{
		"sub":                "client-creds-service",
		"exp":                now.Add(1 * time.Hour).Unix(),
		"iat":                now.Unix(),
		"preferred_username": "my-service",
		"scope":              "read",
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}
