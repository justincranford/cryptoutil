package middleware

import (
	crand "crypto/rand"
	rsa "crypto/rsa"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	json "encoding/json"
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

// testJWKSServer creates a mock JWKS server and returns signed JWT tokens.
type testJWKSServer struct {
	server     *httptest.Server
	privateKey *rsa.PrivateKey
	keyID      string
	keySet     joseJwk.Set
}

func newTestJWKSServer(t *testing.T) *testJWKSServer {
	t.Helper()

	const rsaKeyBits = 2048

	privateKey, err := rsa.GenerateKey(crand.Reader, rsaKeyBits)
	require.NoError(t, err)

	jwkKey, err := joseJwk.Import(privateKey.Public())
	require.NoError(t, err)

	keyID := "test-key-1"

	require.NoError(t, jwkKey.Set(joseJwk.KeyIDKey, keyID))
	require.NoError(t, jwkKey.Set(joseJwk.AlgorithmKey, joseJwa.RS256()))
	require.NoError(t, jwkKey.Set(joseJwk.KeyUsageKey, cryptoutilSharedMagic.JoseKeyUseSig))

	keySet := joseJwk.NewSet()

	require.NoError(t, keySet.AddKey(jwkKey))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		buf, marshalErr := json.Marshal(keySet)
		require.NoError(t, marshalErr)

		_, writeErr := w.Write(buf)
		require.NoError(t, writeErr)
	}))

	t.Cleanup(srv.Close)

	return &testJWKSServer{
		server:     srv,
		privateKey: privateKey,
		keyID:      keyID,
		keySet:     keySet,
	}
}

func (s *testJWKSServer) signToken(t *testing.T, claims map[string]any) string {
	t.Helper()

	privJWK, err := joseJwk.Import(s.privateKey)
	require.NoError(t, err)
	require.NoError(t, privJWK.Set(joseJwk.KeyIDKey, s.keyID))
	require.NoError(t, privJWK.Set(joseJwk.AlgorithmKey, joseJwa.RS256()))

	builder := jwt.NewBuilder()

	for k, v := range claims {
		builder = builder.Claim(k, v)
	}

	token, err := builder.Build()
	require.NoError(t, err)

	signed, err := jwt.Sign(token, jwt.WithKey(joseJwa.RS256(), privJWK))
	require.NoError(t, err)

	return string(signed)
}

func TestValidateToken_WithMockJWKS(t *testing.T) {
	t.Parallel()

	jwksServer := newTestJWKSServer(t)
	now := time.Now().UTC()

	tests := []struct {
		name    string
		claims  map[string]any
		config  JWTValidatorConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid token with all claims",
			claims: map[string]any{
				cryptoutilSharedMagic.ClaimSub:   "user-123",
				cryptoutilSharedMagic.ClaimIss:   "https://auth.example.com",
				cryptoutilSharedMagic.ClaimAud:   []string{"https://api.example.com"},
				cryptoutilSharedMagic.ClaimExp:   now.Add(1 * time.Hour).Unix(),
				cryptoutilSharedMagic.ClaimIat:   now.Unix(),
				cryptoutilSharedMagic.ClaimNbf:   now.Add(-1 * time.Minute).Unix(),
				cryptoutilSharedMagic.ClaimJti:   "token-1",
				cryptoutilSharedMagic.ClaimScope: "read write",
			},
			config: JWTValidatorConfig{
				JWKSURL:          jwksServer.server.URL,
				RequiredIssuer:   "https://auth.example.com",
				RequiredAudience: "https://api.example.com",
			},
			wantErr: false,
		},
		{
			name: "valid token minimal claims",
			claims: map[string]any{
				cryptoutilSharedMagic.ClaimSub: "user-456",
				cryptoutilSharedMagic.ClaimExp: now.Add(1 * time.Hour).Unix(),
				cryptoutilSharedMagic.ClaimIat: now.Unix(),
			},
			config: JWTValidatorConfig{
				JWKSURL: jwksServer.server.URL,
			},
			wantErr: false,
		},
		{
			name: "expired token",
			claims: map[string]any{
				cryptoutilSharedMagic.ClaimSub: "user-expired",
				cryptoutilSharedMagic.ClaimExp: now.Add(-1 * time.Hour).Unix(),
				cryptoutilSharedMagic.ClaimIat: now.Add(-2 * time.Hour).Unix(),
			},
			config: JWTValidatorConfig{
				JWKSURL: jwksServer.server.URL,
			},
			wantErr: true,
			errMsg:  "token validation failed",
		},
		{
			name: "wrong issuer",
			claims: map[string]any{
				cryptoutilSharedMagic.ClaimSub: "user-789",
				cryptoutilSharedMagic.ClaimIss: "https://wrong-issuer.com",
				cryptoutilSharedMagic.ClaimExp: now.Add(1 * time.Hour).Unix(),
				cryptoutilSharedMagic.ClaimIat: now.Unix(),
			},
			config: JWTValidatorConfig{
				JWKSURL:        jwksServer.server.URL,
				RequiredIssuer: "https://auth.example.com",
			},
			wantErr: true,
			errMsg:  "token validation failed",
		},
		{
			name: "wrong audience",
			claims: map[string]any{
				cryptoutilSharedMagic.ClaimSub: "user-aud",
				cryptoutilSharedMagic.ClaimAud: []string{"https://wrong-api.com"},
				cryptoutilSharedMagic.ClaimExp: now.Add(1 * time.Hour).Unix(),
				cryptoutilSharedMagic.ClaimIat: now.Unix(),
			},
			config: JWTValidatorConfig{
				JWKSURL:          jwksServer.server.URL,
				RequiredAudience: "https://api.example.com",
			},
			wantErr: true,
			errMsg:  "token validation failed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			validator, err := NewJWTValidator(tc.config)
			require.NoError(t, err)

			tokenString := jwksServer.signToken(t, tc.claims)
			claims, err := validator.ValidateToken(t.Context(), tokenString)

			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, claims)
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, claims)
			}
		})
	}
}

func TestValidateToken_InvalidJWKSURL(t *testing.T) {
	t.Parallel()

	validator, err := NewJWTValidator(JWTValidatorConfig{
		JWKSURL: "http://127.0.0.1:1/nonexistent",
	})
	require.NoError(t, err)

	claims, err := validator.ValidateToken(t.Context(), "some-token")
	require.Error(t, err)
	require.Nil(t, claims)
	require.Contains(t, err.Error(), "failed to get JWKS")
}

func TestValidateToken_InvalidTokenString(t *testing.T) {
	t.Parallel()

	jwksServer := newTestJWKSServer(t)

	validator, err := NewJWTValidator(JWTValidatorConfig{
		JWKSURL: jwksServer.server.URL,
	})
	require.NoError(t, err)

	claims, err := validator.ValidateToken(t.Context(), "not-a-jwt-token")
	require.Error(t, err)
	require.Nil(t, claims)
	require.Contains(t, err.Error(), "token validation failed")
}

func TestJWTMiddleware_FullFlow(t *testing.T) {
	t.Parallel()

	jwksServer := newTestJWKSServer(t)
	now := time.Now().UTC()

	validator, err := NewJWTValidator(JWTValidatorConfig{
		JWKSURL: jwksServer.server.URL,
	})
	require.NoError(t, err)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name: "valid token succeeds",
			authHeader: cryptoutilSharedMagic.AuthorizationBearerPrefix + jwksServer.signToken(t, map[string]any{
				cryptoutilSharedMagic.ClaimSub:   "user-123",
				cryptoutilSharedMagic.ClaimExp:   now.Add(1 * time.Hour).Unix(),
				cryptoutilSharedMagic.ClaimIat:   now.Unix(),
				cryptoutilSharedMagic.ClaimScope: cryptoutilSharedMagic.ScopeRead,
			}),
			expectedStatus: http.StatusOK,
		},
		{
			name: "expired token fails",
			authHeader: cryptoutilSharedMagic.AuthorizationBearerPrefix + jwksServer.signToken(t, map[string]any{
				cryptoutilSharedMagic.ClaimSub: "user-expired",
				cryptoutilSharedMagic.ClaimExp: now.Add(-1 * time.Hour).Unix(),
				cryptoutilSharedMagic.ClaimIat: now.Add(-2 * time.Hour).Unix(),
			}),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid token string",
			authHeader:     "Bearer not-a-valid-jwt",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Get("/test", validator.JWTMiddleware(), func(c *fiber.Ctx) error {
				return c.SendStatus(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", tc.authHeader)

			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

func TestValidateToken_WithRevocationCheck(t *testing.T) {
	t.Parallel()

	jwksServer := newTestJWKSServer(t)

	introspectionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)

		_, writeErr := w.Write([]byte(`{"active":true}`))
		require.NoError(t, writeErr)
	}))
	t.Cleanup(introspectionServer.Close)

	now := time.Now().UTC()

	validator, err := NewJWTValidator(JWTValidatorConfig{
		JWKSURL:             jwksServer.server.URL,
		RevocationCheckMode: RevocationCheckEveryRequest,
		IntrospectionURL:    introspectionServer.URL,
	})
	require.NoError(t, err)

	tokenString := jwksServer.signToken(t, map[string]any{
		cryptoutilSharedMagic.ClaimSub:   "user-revcheck",
		cryptoutilSharedMagic.ClaimExp:   now.Add(1 * time.Hour).Unix(),
		cryptoutilSharedMagic.ClaimIat:   now.Unix(),
		cryptoutilSharedMagic.ClaimScope: "read write",
	})

	claims, err := validator.ValidateToken(t.Context(), tokenString)
	require.NoError(t, err)
	require.NotNil(t, claims)
	require.Equal(t, "user-revcheck", claims.Subject)
}

func TestValidateToken_RevokedDuringCheck(t *testing.T) {
	t.Parallel()

	jwksServer := newTestJWKSServer(t)

	introspectionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)

		_, writeErr := w.Write([]byte(`{"active":false}`))
		require.NoError(t, writeErr)
	}))
	t.Cleanup(introspectionServer.Close)

	now := time.Now().UTC()

	validator, err := NewJWTValidator(JWTValidatorConfig{
		JWKSURL:             jwksServer.server.URL,
		RevocationCheckMode: RevocationCheckEveryRequest,
		IntrospectionURL:    introspectionServer.URL,
	})
	require.NoError(t, err)

	tokenString := jwksServer.signToken(t, map[string]any{
		cryptoutilSharedMagic.ClaimSub: "revoked-user",
		cryptoutilSharedMagic.ClaimExp: now.Add(1 * time.Hour).Unix(),
		cryptoutilSharedMagic.ClaimIat: now.Unix(),
	})

	claims, err := validator.ValidateToken(t.Context(), tokenString)
	require.Error(t, err)
	require.Nil(t, claims)
	require.Contains(t, err.Error(), "revoked")
}

func TestValidateToken_WithAllowedAlgorithms(t *testing.T) {
	t.Parallel()

	jwksServer := newTestJWKSServer(t)
	now := time.Now().UTC()

	validator, err := NewJWTValidator(JWTValidatorConfig{
		JWKSURL:           jwksServer.server.URL,
		AllowedAlgorithms: DefaultAllowedAlgorithms(),
	})
	require.NoError(t, err)

	tokenString := jwksServer.signToken(t, map[string]any{
		cryptoutilSharedMagic.ClaimSub: "user-alg",
		cryptoutilSharedMagic.ClaimExp: now.Add(1 * time.Hour).Unix(),
		cryptoutilSharedMagic.ClaimIat: now.Unix(),
	})

	claims, err := validator.ValidateToken(t.Context(), tokenString)
	require.NoError(t, err)
	require.NotNil(t, claims)
}
