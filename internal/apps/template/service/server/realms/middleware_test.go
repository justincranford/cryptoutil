// Copyright (c) 2025 Justin Cranford
//

package realms

import (
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const testJWTSecret = "test-jwt-secret-key-for-testing-12345"

// createTestToken generates a JWT token for testing.
func createTestToken(t *testing.T, userID string, username string, secret string, expiration time.Time) string {
	t.Helper()

	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiration),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	return tokenString
}

func TestJWTMiddleware_MissingAuthorizationHeader(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(JWTMiddleware(testJWTSecret))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)
}

func TestJWTMiddleware_InvalidAuthorizationFormat(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(JWTMiddleware(testJWTSecret))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz") // Basic auth instead of Bearer

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)
}

func TestJWTMiddleware_InvalidToken(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(JWTMiddleware(testJWTSecret))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-string")

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)
}

func TestJWTMiddleware_ExpiredToken(t *testing.T) {
	t.Parallel()

	userID := googleUuid.New()
	expiredToken := createTestToken(t, userID.String(), "testuser", testJWTSecret, time.Now().Add(-1*time.Hour))

	app := fiber.New()
	app.Use(JWTMiddleware(testJWTSecret))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)
}

func TestJWTMiddleware_WrongSecret(t *testing.T) {
	t.Parallel()

	userID := googleUuid.New()
	// Token signed with different secret.
	tokenWithWrongSecret := createTestToken(t, userID.String(), "testuser", "wrong-secret", time.Now().Add(1*time.Hour))

	app := fiber.New()
	app.Use(JWTMiddleware(testJWTSecret))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenWithWrongSecret)

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)
}

func TestJWTMiddleware_InvalidUserID(t *testing.T) {
	t.Parallel()

	// Token with invalid UUID format for user_id.
	invalidUserIDToken := createTestToken(t, "not-a-valid-uuid", "testuser", testJWTSecret, time.Now().Add(1*time.Hour))

	app := fiber.New()
	app.Use(JWTMiddleware(testJWTSecret))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+invalidUserIDToken)

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)
}

func TestJWTMiddleware_ValidToken_Success(t *testing.T) {
	t.Parallel()

	userID := googleUuid.New()
	validToken := createTestToken(t, userID.String(), "testuser", testJWTSecret, time.Now().Add(1*time.Hour))

	var capturedUserID googleUuid.UUID

	app := fiber.New()
	app.Use(JWTMiddleware(testJWTSecret))
	app.Get("/protected", func(c *fiber.Ctx) error {
		// Capture the user ID from context.
		capturedUserID = c.Locals(ContextKeyUserID).(googleUuid.UUID) //nolint:errcheck // Test assertion

		return c.SendString("success")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.Equal(t, userID, capturedUserID)

	err = resp.Body.Close()
	require.NoError(t, err)
}

func TestJWTMiddleware_UnsupportedSigningMethod(t *testing.T) {
	t.Parallel()

	// Create token with RS256 (RSA) signing method instead of HS256 (HMAC).
	// The middleware should reject this as it only supports HMAC.
	userID := googleUuid.New()
	claims := &Claims{
		UserID:   userID.String(),
		Username: "testuser",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Create a token with "none" signing method (should be rejected).
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	app := fiber.New()
	app.Use(JWTMiddleware(testJWTSecret))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)
}
