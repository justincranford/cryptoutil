package middleware

import (
	"context"
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	crand "crypto/rand"
	rsa "crypto/rsa"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"errors"
	"fmt"
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
	"github.com/stretchr/testify/require"
)

func newTestJWTValidator(t *testing.T, detailLevel string) *JWTValidator {
	t.Helper()

	v, err := NewJWTValidator(JWTValidatorConfig{
		JWKSURL:           "https://localhost/.well-known/jwks.json",
		ErrorDetailLevel:  detailLevel,
		AllowedAlgorithms: DefaultAllowedAlgorithms(),
	})
	require.NoError(t, err)

	return v
}

func TestExtractClaims(t *testing.T) {
	t.Parallel()

	validator := newTestJWTValidator(t, errorDetailLevelMin)

	tests := []struct {
		name       string
		buildToken func() jwt.Token
		checkFn    func(t *testing.T, claims *JWTClaims)
	}{
		{
			name: "token with standard claims",
			buildToken: func() jwt.Token {
				tok, err := jwt.NewBuilder().
					Subject("user-123").
					Issuer("https://auth.example.com").
					Audience([]string{"api.example.com"}).
					Expiration(time.Now().UTC().Add(time.Hour)).
					IssuedAt(time.Now().UTC()).
					NotBefore(time.Now().UTC().Add(-time.Minute)).
					JwtID("jti-abc").
					Build()
				require.NoError(t, err)

				return tok
			},
			checkFn: func(t *testing.T, claims *JWTClaims) {
				t.Helper()
				require.Equal(t, "user-123", claims.Subject)
				require.Equal(t, "https://auth.example.com", claims.Issuer)
				require.Contains(t, claims.Audience, "api.example.com")
				require.False(t, claims.ExpiresAt.IsZero())
				require.False(t, claims.IssuedAt.IsZero())
				require.False(t, claims.NotBefore.IsZero())
				require.Equal(t, "jti-abc", claims.JTI)
			},
		},
		{
			name: "token with OIDC claims",
			buildToken: func() jwt.Token {
				tok, err := jwt.NewBuilder().
					Subject("user-oidc").
					Build()
				require.NoError(t, err)
				require.NoError(t, tok.Set(cryptoutilSharedMagic.ClaimName, "John Doe"))
				require.NoError(t, tok.Set(cryptoutilSharedMagic.ClaimPreferredUsername, "johndoe"))
				require.NoError(t, tok.Set(cryptoutilSharedMagic.ClaimEmail, "john@example.com"))
				require.NoError(t, tok.Set(cryptoutilSharedMagic.ClaimEmailVerified, true))

				return tok
			},
			checkFn: func(t *testing.T, claims *JWTClaims) {
				t.Helper()
				require.Equal(t, "John Doe", claims.Name)
				require.Equal(t, "johndoe", claims.PreferredUsername)
				require.Equal(t, "john@example.com", claims.Email)
				require.True(t, claims.EmailVerified)
			},
		},
		{
			name: "token with scope claim",
			buildToken: func() jwt.Token {
				tok, err := jwt.NewBuilder().
					Subject("user-scoped").
					Build()
				require.NoError(t, err)
				require.NoError(t, tok.Set(cryptoutilSharedMagic.ClaimScope, "kms:read kms:write kms:admin"))

				return tok
			},
			checkFn: func(t *testing.T, claims *JWTClaims) {
				t.Helper()
				require.Equal(t, "kms:read kms:write kms:admin", claims.Scope)
				require.Equal(t, []string{"kms:read", "kms:write", "kms:admin"}, claims.Scopes)
			},
		},
		{
			name: "empty token has no claims",
			buildToken: func() jwt.Token {
				tok, err := jwt.NewBuilder().Build()
				require.NoError(t, err)

				return tok
			},
			checkFn: func(t *testing.T, claims *JWTClaims) {
				t.Helper()
				require.Empty(t, claims.Subject)
				require.Empty(t, claims.Issuer)
				require.Nil(t, claims.Audience)
				require.True(t, claims.ExpiresAt.IsZero())
				require.Empty(t, claims.Scope)
				require.Nil(t, claims.Scopes)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			token := tc.buildToken()
			claims := validator.extractClaims(token)
			require.NotNil(t, claims)
			require.NotNil(t, claims.Custom)
			tc.checkFn(t, claims)
		})
	}
}

func TestHandleValidationError(t *testing.T) {
	t.Parallel()

	validator := newTestJWTValidator(t, errorDetailLevelStd)

	tests := []struct {
		name         string
		err          error
		expectedCode string
	}{
		{name: "expired error", err: errors.New("token expired"), expectedCode: "token_expired"},
		{name: "revoked error", err: errors.New("token revoked"), expectedCode: "token_revoked"},
		{name: "issuer error", err: errors.New("invalid issuer"), expectedCode: "invalid_issuer"},
		{name: "audience error", err: errors.New("invalid audience"), expectedCode: "invalid_audience"},
		{name: "signature error", err: errors.New("invalid signature"), expectedCode: "invalid_signature"},
		{name: "generic error", err: errors.New("something went wrong"), expectedCode: cryptoutilSharedMagic.ErrorInvalidToken},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Get("/test", func(c *fiber.Ctx) error {
				return validator.handleValidationError(c, tc.err)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})
	}
}

func TestUnauthorizedAndForbiddenError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		detailLevel  string
		callFn       string
		expectedCode int
	}{
		{name: "unauthorized minimal", detailLevel: errorDetailLevelMin, callFn: "unauthorized", expectedCode: http.StatusUnauthorized},
		{name: "unauthorized standard", detailLevel: errorDetailLevelStd, callFn: "unauthorized", expectedCode: http.StatusUnauthorized},
		{name: "forbidden minimal", detailLevel: errorDetailLevelMin, callFn: "forbidden", expectedCode: http.StatusForbidden},
		{name: "forbidden standard", detailLevel: errorDetailLevelStd, callFn: "forbidden", expectedCode: http.StatusForbidden},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			validator := newTestJWTValidator(t, tc.detailLevel)

			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Get("/test", func(c *fiber.Ctx) error {
				if tc.callFn == "unauthorized" {
					return validator.unauthorizedError(c, "test_error", "Test message")
				}

				return validator.forbiddenError(c, "test_error", "Test message")
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, tc.expectedCode, resp.StatusCode)
		})
	}
}

func TestRequireScopeMiddleware_JWTRevocation(t *testing.T) {
	t.Parallel()

	validator := newTestJWTValidator(t, errorDetailLevelStd)

	tests := []struct {
		name           string
		claims         *JWTClaims
		setClaims      bool
		expectedStatus int
	}{
		{
			name:           "has all required scopes",
			claims:         &JWTClaims{Scopes: []string{"kms:read", "kms:write"}},
			setClaims:      true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing required scope",
			claims:         &JWTClaims{Scopes: []string{"kms:read"}},
			setClaims:      true,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "no claims in context",
			setClaims:      false,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{DisableStartupMessage: true})

			app.Use(func(c *fiber.Ctx) error {
				if tc.setClaims {
					ctx := context.WithValue(c.UserContext(), JWTContextKey{}, tc.claims)
					c.SetUserContext(ctx)
				}

				return c.Next()
			})
			app.Get("/test", RequireScopeMiddleware(validator, "kms:read", "kms:write"), func(c *fiber.Ctx) error {
				return c.SendStatus(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

func TestRequireAnyScopeMiddleware_JWTRevocation(t *testing.T) {
	t.Parallel()

	validator := newTestJWTValidator(t, errorDetailLevelStd)

	tests := []struct {
		name           string
		claims         *JWTClaims
		setClaims      bool
		expectedStatus int
	}{
		{
			name:           "has one of required scopes",
			claims:         &JWTClaims{Scopes: []string{"kms:read"}},
			setClaims:      true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing all required scopes",
			claims:         &JWTClaims{Scopes: []string{"other:scope"}},
			setClaims:      true,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "no claims in context",
			setClaims:      false,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{DisableStartupMessage: true})

			app.Use(func(c *fiber.Ctx) error {
				if tc.setClaims {
					ctx := context.WithValue(c.UserContext(), JWTContextKey{}, tc.claims)
					c.SetUserContext(ctx)
				}

				return c.Next()
			})
			app.Get("/test", RequireAnyScopeMiddleware(validator, "kms:read", "kms:write"), func(c *fiber.Ctx) error {
				return c.SendStatus(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

func TestPublicKeyFromJWK(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setupFn func(t *testing.T) joseJwk.Key
		wantErr bool
		keyType string
	}{
		{
			name: "RSA public key",
			setupFn: func(t *testing.T) joseJwk.Key {
				t.Helper()

				rsaKey, err := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)
				require.NoError(t, err)
				jwkKey, err := joseJwk.Import(&rsaKey.PublicKey)
				require.NoError(t, err)

				return jwkKey
			},
			wantErr: false,
			keyType: "*rsa.PublicKey",
		},
		{
			name: "ECDSA public key",
			setupFn: func(t *testing.T) joseJwk.Key {
				t.Helper()

				ecKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
				require.NoError(t, err)
				jwkKey, err := joseJwk.Import(&ecKey.PublicKey)
				require.NoError(t, err)

				return jwkKey
			},
			wantErr: false,
			keyType: "*ecdsa.PublicKey",
		},
		{
			name: "Ed25519 public key",
			setupFn: func(t *testing.T) joseJwk.Key {
				t.Helper()

				pubKey, _, err := ed25519.GenerateKey(crand.Reader)
				require.NoError(t, err)
				jwkKey, err := joseJwk.Import(pubKey)
				require.NoError(t, err)

				return jwkKey
			},
			wantErr: false,
			keyType: "ed25519.PublicKey",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			jwkKey := tc.setupFn(t)
			pubKey, err := PublicKeyFromJWK(jwkKey)

			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, pubKey)
			} else {
				require.NoError(t, err)
				require.NotNil(t, pubKey)
				require.Equal(t, tc.keyType, fmt.Sprintf("%T", pubKey))
			}
		})
	}
}
