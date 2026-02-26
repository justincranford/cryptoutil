package middleware

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	crand "crypto/rand"
	rsa "crypto/rsa"
	json "encoding/json"
	"errors"
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
	"github.com/stretchr/testify/require"
)

func TestNewServiceAuthMiddleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  ServiceAuthConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid API key config",
			config: ServiceAuthConfig{
				AllowedMethods: []AuthMethod{AuthMethodAPIKey},
				APIKeyConfig: &APIKeyConfig{
					ValidKeys: map[string]string{"key1": "svc1"},
				},
			},
			wantErr: false,
		},
		{
			name:    "empty allowed methods",
			config:  ServiceAuthConfig{},
			wantErr: true,
			errMsg:  "at least one auth method must be allowed",
		},
		{
			name: "valid JWT config",
			config: ServiceAuthConfig{
				AllowedMethods: []AuthMethod{AuthMethodJWT},
				JWTConfig: &JWTValidatorConfig{
					JWKSURL: "https://auth.example.com/.well-known/jwks.json",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid JWT config",
			config: ServiceAuthConfig{
				AllowedMethods: []AuthMethod{AuthMethodJWT},
				JWTConfig: &JWTValidatorConfig{
					JWKSURL: "",
				},
			},
			wantErr: true,
			errMsg:  "failed to create JWT validator",
		},
		{
			name: "mTLS config",
			config: ServiceAuthConfig{
				AllowedMethods: []AuthMethod{AuthMethodMTLS},
				MTLSConfig: &MTLSConfig{
					RequireClientCert: true,
				},
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mw, err := NewServiceAuthMiddleware(tc.config)

			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, mw)
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, mw)
			}
		})
	}
}

func TestServiceAuth_APIKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		headerName     string
		headerValue    string
		validKeys      map[string]string
		expectedStatus int
	}{
		{
			name:           "valid API key",
			headerName:     "X-API-Key",
			headerValue:    "valid-key-123",
			validKeys:      map[string]string{"valid-key-123": "test-service"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid API key",
			headerName:     "X-API-Key",
			headerValue:    "wrong-key",
			validKeys:      map[string]string{"valid-key-123": "test-service"},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "missing API key",
			headerName:     "X-API-Key",
			headerValue:    "",
			validKeys:      map[string]string{"valid-key-123": "test-service"},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mw, err := NewServiceAuthMiddleware(ServiceAuthConfig{
				AllowedMethods: []AuthMethod{AuthMethodAPIKey},
				APIKeyConfig: &APIKeyConfig{
					ValidKeys: tc.validKeys,
				},
			})
			require.NoError(t, err)

			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Use(mw.Middleware())
			app.Get("/test", func(c *fiber.Ctx) error {
				return c.SendStatus(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tc.headerValue != "" {
				req.Header.Set(tc.headerName, tc.headerValue)
			}

			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

func TestServiceAuth_APIKeyValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		apiKey         string
		validator      func(ctx context.Context, apiKey string) (string, bool, error)
		expectedStatus int
	}{
		{
			name:   "valid via validator",
			apiKey: "dynamic-key",
			validator: func(_ context.Context, _ string) (string, bool, error) {
				return "dynamic-svc", true, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "invalid via validator",
			apiKey: "bad-key",
			validator: func(_ context.Context, _ string) (string, bool, error) {
				return "", false, nil
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:   "validator error",
			apiKey: "error-key",
			validator: func(_ context.Context, _ string) (string, bool, error) {
				return "", false, errors.New("db error")
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mw, err := NewServiceAuthMiddleware(ServiceAuthConfig{
				AllowedMethods: []AuthMethod{AuthMethodAPIKey},
				APIKeyConfig: &APIKeyConfig{
					KeyValidator: tc.validator,
				},
			})
			require.NoError(t, err)

			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Use(mw.Middleware())
			app.Get("/test", func(c *fiber.Ctx) error {
				return c.SendStatus(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("X-API-Key", tc.apiKey)

			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

func TestServiceAuth_JWT(t *testing.T) {
	t.Parallel()

	const rsaKeyBits = 2048

	privateKey, err := rsa.GenerateKey(crand.Reader, rsaKeyBits)
	require.NoError(t, err)

	pubJWK, err := joseJwk.Import(privateKey.Public())
	require.NoError(t, err)

	keyID := "sa-test-key-1"

	require.NoError(t, pubJWK.Set(joseJwk.KeyIDKey, keyID))
	require.NoError(t, pubJWK.Set(joseJwk.AlgorithmKey, joseJwa.RS256()))
	require.NoError(t, pubJWK.Set(joseJwk.KeyUsageKey, cryptoutilSharedMagic.JoseKeyUseSig))

	keySet := joseJwk.NewSet()

	require.NoError(t, keySet.AddKey(pubJWK))

	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		buf, marshalErr := json.Marshal(keySet)
		require.NoError(t, marshalErr)

		_, writeErr := w.Write(buf)
		require.NoError(t, writeErr)
	}))
	t.Cleanup(jwksServer.Close)

	signToken := func(t *testing.T, sub string, exp time.Time) string {
		t.Helper()

		privJWK, importErr := joseJwk.Import(privateKey)
		require.NoError(t, importErr)
		require.NoError(t, privJWK.Set(joseJwk.KeyIDKey, keyID))
		require.NoError(t, privJWK.Set(joseJwk.AlgorithmKey, joseJwa.RS256()))

		now := time.Now().UTC()

		token, buildErr := jwt.NewBuilder().
			Claim(cryptoutilSharedMagic.ClaimSub, sub).
			Claim(cryptoutilSharedMagic.ClaimIat, now.Unix()).
			Claim(cryptoutilSharedMagic.ClaimExp, exp.Unix()).
			Claim(cryptoutilSharedMagic.ClaimScope, cryptoutilSharedMagic.ScopeRead).
			Build()
		require.NoError(t, buildErr)

		signed, signErr := jwt.Sign(token, jwt.WithKey(joseJwa.RS256(), privJWK))
		require.NoError(t, signErr)

		return string(signed)
	}

	mw, err := NewServiceAuthMiddleware(ServiceAuthConfig{
		AllowedMethods: []AuthMethod{AuthMethodJWT},
		JWTConfig: &JWTValidatorConfig{
			JWKSURL: jwksServer.URL,
		},
	})
	require.NoError(t, err)

	now := time.Now().UTC()

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "valid JWT token",
			authHeader:     cryptoutilSharedMagic.AuthorizationBearerPrefix + signToken(t, "svc-user", now.Add(1*time.Hour)),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "expired JWT token",
			authHeader:     cryptoutilSharedMagic.AuthorizationBearerPrefix + signToken(t, "svc-user", now.Add(-1*time.Hour)),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "no bearer prefix",
			authHeader:     "Basic dXNlcjpwYXNz",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "missing header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

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
