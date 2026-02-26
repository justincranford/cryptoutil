package middleware

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	http "net/http"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

func TestJWTMiddleware_HeaderValidation(t *testing.T) {
	t.Parallel()

	validator := newTestJWTValidator(t, errorDetailLevelStd)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "non-bearer scheme",
			authHeader:     "Basic dXNlcjpwYXNz",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "empty bearer token",
			authHeader:     cryptoutilSharedMagic.AuthorizationBearerPrefix,
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

func TestPublicKeyFromJWK_UnsupportedType(t *testing.T) {
	t.Parallel()

	// Create a symmetric (oct) JWK key which is not a supported public key type.
	symKey, err := joseJwk.Import([]byte("test-symmetric-key-32-bytes-long!"))
	require.NoError(t, err)

	pubKey, err := PublicKeyFromJWK(symKey)
	require.Error(t, err)
	require.Nil(t, pubKey)
	require.Contains(t, err.Error(), "unsupported key type")
}
