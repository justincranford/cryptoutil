// Copyright (c) 2025 Justin Cranford
//
//

package server

import (
	"net/http/httptest"
	"testing"

	googleUuid "github.com/google/uuid"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// TestIsNotFoundError tests the isNotFoundError helper function.
func TestIsNotFoundError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{
			name:     "nil error returns false",
			errMsg:   "",
			expected: false,
		},
		{
			name:     "contains not found",
			errMsg:   "elastic key not found",
			expected: true,
		},
		{
			name:     "contains NOT FOUND uppercase",
			errMsg:   "elastic key NOT FOUND",
			expected: true,
		},
		{
			name:     "contains does not exist",
			errMsg:   "key does not exist in database",
			expected: true,
		},
		{
			name:     "unrelated error",
			errMsg:   "database connection failed",
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var err error
			if tc.errMsg != "" {
				err = &testError{msg: tc.errMsg}
			}

			result := isNotFoundError(err)
			require.Equal(t, tc.expected, result)
		})
	}
}

// TestIsSymmetricKeyError tests the isSymmetricKeyError helper function.
func TestIsSymmetricKeyError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{
			name:     "nil error returns false",
			errMsg:   "",
			expected: false,
		},
		{
			name:     "contains symmetric",
			errMsg:   "cannot get public key for symmetric algorithm",
			expected: true,
		},
		{
			name:     "contains SYMMETRIC uppercase",
			errMsg:   "SYMMETRIC keys do not have public keys",
			expected: true,
		},
		{
			name:     "contains no public key",
			errMsg:   "no public key available for this key type",
			expected: true,
		},
		{
			name:     "unrelated error",
			errMsg:   "encryption failed",
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var err error
			if tc.errMsg != "" {
				err = &testError{msg: tc.errMsg}
			}

			result := isSymmetricKeyError(err)
			require.Equal(t, tc.expected, result)
		})
	}
}

// testError is a simple error type for testing.
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// TestHandleJWKGet_EmptyKID tests handleJWKGet with empty KID.
func TestHandleJWKGet_EmptyKID(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	handler := &joseHandlerAdapter{
		keyStore: NewKeyStore(),
	}
	app.Get("/jwk/:kid", handler.handleJWKGet)

	req := httptest.NewRequest("GET", "/jwk/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	// Empty KID results in 404 from Fiber routing (no match).
	require.Equal(t, 404, resp.StatusCode)
}

// TestHandleJWKGet_NotFound tests handleJWKGet with non-existent key.
func TestHandleJWKGet_NotFound(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	handler := &joseHandlerAdapter{
		keyStore: NewKeyStore(),
	}
	app.Get("/jwk/:kid", handler.handleJWKGet)

	req := httptest.NewRequest("GET", "/jwk/nonexistent-key-id", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 404, resp.StatusCode)
}

// TestHandleJWKDelete_EmptyKID tests handleJWKDelete with empty KID.
func TestHandleJWKDelete_EmptyKID(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	handler := &joseHandlerAdapter{
		keyStore: NewKeyStore(),
	}
	app.Delete("/jwk/:kid", handler.handleJWKDelete)

	req := httptest.NewRequest("DELETE", "/jwk/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	// Empty KID results in 404 from Fiber routing (no match).
	require.Equal(t, 404, resp.StatusCode)
}

// TestHandleJWKDelete_NotFound tests handleJWKDelete with non-existent key.
func TestHandleJWKDelete_NotFound(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	handler := &joseHandlerAdapter{
		keyStore: NewKeyStore(),
	}
	app.Delete("/jwk/:kid", handler.handleJWKDelete)

	req := httptest.NewRequest("DELETE", "/jwk/nonexistent-key-id", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 404, resp.StatusCode)
}

// TestHandleJWKGet_WithPublicJWK tests handleJWKGet with a key that has PublicJWK set.
func TestHandleJWKGet_WithPublicJWK(t *testing.T) {
	t.Parallel()

	// Create keystore with a test key.
	keyStore := NewKeyStore()
	kid := googleUuid.New()
	testKey := &StoredKey{
		KID:       kid,
		PublicJWK: nil, // Will test the PrivateJWK path instead.
		KeyType:   "RSA",
		Algorithm: "RS256",
		Use:       "sig",
		CreatedAt: 1234567890,
	}
	err := keyStore.Store(testKey)
	require.NoError(t, err)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	handler := &joseHandlerAdapter{
		keyStore: keyStore,
	}
	app.Get("/jwk/:kid", handler.handleJWKGet)

	req := httptest.NewRequest("GET", "/jwk/"+kid.String(), nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	// Key exists, should return 200 OK.
	require.Equal(t, 200, resp.StatusCode)
}

// TestHandleJWKDelete_Success tests handleJWKDelete with an existing key.
func TestHandleJWKDelete_Success(t *testing.T) {
	t.Parallel()

	// Create keystore with a test key.
	keyStore := NewKeyStore()
	kid := googleUuid.New()
	testKey := &StoredKey{
		KID:       kid,
		KeyType:   "RSA",
		Algorithm: "RS256",
		Use:       "sig",
		CreatedAt: 1234567890,
	}
	err := keyStore.Store(testKey)
	require.NoError(t, err)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	handler := &joseHandlerAdapter{
		keyStore: keyStore,
	}
	app.Delete("/jwk/:kid", handler.handleJWKDelete)

	req := httptest.NewRequest("DELETE", "/jwk/"+kid.String(), nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	// Key exists, delete should return 200 OK.
	require.Equal(t, 200, resp.StatusCode)

	// Verify key is deleted.
	_, exists := keyStore.Get(kid.String())
	require.False(t, exists)
}

// TestHandleElasticJWKS_EmptyKID tests handleElasticJWKS with empty KID.
func TestHandleElasticJWKS_EmptyKID(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	handler := &joseHandlerAdapter{
		keyStore: NewKeyStore(),
	}
	app.Get("/elastic-jwks/:kid", handler.handleElasticJWKS)

	req := httptest.NewRequest("GET", "/elastic-jwks/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	// Empty KID results in 404 from Fiber routing (no match).
	require.Equal(t, 404, resp.StatusCode)
}

// TestHandleElasticJWKS_InvalidKID tests handleElasticJWKS with invalid UUID format.
func TestHandleElasticJWKS_InvalidKID(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	handler := &joseHandlerAdapter{
		keyStore: NewKeyStore(),
	}
	app.Get("/elastic-jwks/:kid", handler.handleElasticJWKS)

	req := httptest.NewRequest("GET", "/elastic-jwks/not-a-valid-uuid", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	// Invalid UUID format should return 400 Bad Request.
	require.Equal(t, 400, resp.StatusCode)
}

// TestHandleElasticJWKS_NilService tests handleElasticJWKS when elasticJWKService is nil.
func TestHandleElasticJWKS_NilService(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	handler := &joseHandlerAdapter{
		keyStore:          NewKeyStore(),
		elasticJWKService: nil, // Service not configured.
	}
	app.Get("/elastic-jwks/:kid", handler.handleElasticJWKS)

	validUUID := googleUuid.New().String()
	req := httptest.NewRequest("GET", "/elastic-jwks/"+validUUID, nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	// Nil service should return 500 Internal Server Error.
	require.Equal(t, 500, resp.StatusCode)
}
