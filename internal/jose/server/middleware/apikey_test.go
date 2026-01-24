// Copyright (c) 2025 Justin Cranford

package middleware

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyMiddleware_NoKey(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	config := &APIKeyConfig{
		ValidKeys: map[string]string{"test-key": "test-client"},
	}
	app.Use(RequireAPIKeyWithConfig(config))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestAPIKeyMiddleware_ValidKey(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	validKeys := map[string]string{"test-key-12345": "test-client"}
	config := &APIKeyConfig{
		ValidKeys: validKeys,
	}
	app.Use(RequireAPIKeyWithConfig(config))
	app.Get("/test", func(c *fiber.Ctx) error {
		info := GetAPIKeyInfo(c)
		require.NotNil(t, info)
		require.Equal(t, "test-client", info.ClientName)

		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "test-key-12345")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestAPIKeyMiddleware_InvalidKey(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	config := &APIKeyConfig{
		ValidKeys: map[string]string{"valid-key": "test-client"},
	}
	app.Use(RequireAPIKeyWithConfig(config))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "invalid-key")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestAPIKeyMiddleware_CustomHeader(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	config := &APIKeyConfig{
		HeaderName: "Authorization",
		ValidKeys:  map[string]string{"bearer-token": "test-client"},
	}
	app.Use(RequireAPIKeyWithConfig(config))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "bearer-token")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestAPIKeyMiddleware_QueryParam(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	config := &APIKeyConfig{
		QueryParam: "api_key",
		ValidKeys:  map[string]string{"query-key-12345": "test-client"},
	}
	app.Use(RequireAPIKeyWithConfig(config))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test?api_key=query-key-12345", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestAPIKeyMiddleware_Skipper(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	config := &APIKeyConfig{
		ValidKeys: map[string]string{"test-key": "test-client"},
		Skipper: func(c *fiber.Ctx) bool {
			return c.Path() == "/public"
		},
	}
	app.Use(RequireAPIKeyWithConfig(config))
	app.Get("/public", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	app.Get("/private", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	// Public route should be accessible without API key.
	req := httptest.NewRequest("GET", "/public", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.NoError(t, resp.Body.Close())

	// Private route should require API key.
	req = httptest.NewRequest("GET", "/private", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestAPIKeyMiddleware_DynamicValidator(t *testing.T) {
	t.Parallel()

	validatorCalled := false

	app := fiber.New()
	config := &APIKeyConfig{
		KeyValidator: func(_ context.Context, apiKey string) (string, bool, error) {
			validatorCalled = true

			if apiKey == "dynamic-key" {
				return "dynamic-client", true, nil
			}

			return "", false, nil
		},
	}
	app.Use(RequireAPIKeyWithConfig(config))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "dynamic-key")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
	require.True(t, validatorCalled)
}

func TestAPIKeyMiddleware_ValidatorError(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	config := &APIKeyConfig{
		KeyValidator: func(_ context.Context, _ string) (string, bool, error) {
			return "", false, errors.New("database error")
		},
	}
	app.Use(RequireAPIKeyWithConfig(config))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "any-key")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestDefaultAPIKeyConfig(t *testing.T) {
	t.Parallel()

	config := DefaultAPIKeyConfig()
	require.Equal(t, "X-API-Key", config.HeaderName)
	require.Empty(t, config.QueryParam)
	require.NotNil(t, config.ValidKeys)
	require.Nil(t, config.KeyValidator)
	require.Nil(t, config.Skipper)
	require.Equal(t, "basic", config.ErrorDetailLevel)
}

func TestNewAPIKeyMiddleware_NilConfig(t *testing.T) {
	t.Parallel()

	mw := NewAPIKeyMiddleware(nil)
	require.NotNil(t, mw)
	require.NotNil(t, mw.config)
	require.Equal(t, "X-API-Key", mw.config.HeaderName)
}

func TestMaskAPIKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		apiKey   string
		expected string
	}{
		{
			name:     "short_key",
			apiKey:   "short",
			expected: "****",
		},
		{
			name:     "exactly_8_chars",
			apiKey:   "12345678",
			expected: "****",
		},
		{
			name:     "long_key",
			apiKey:   "abcd1234efgh5678",
			expected: "abcd****5678",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := maskAPIKey(tc.apiKey)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestSecureCompare(t *testing.T) {
	t.Parallel()

	require.True(t, SecureCompare("test", "test"))
	require.False(t, SecureCompare("test", "other"))
	require.False(t, SecureCompare("test", "tes"))
	require.True(t, SecureCompare("", ""))
}

func TestGetAPIKeyInfo_NoInfo(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		info := GetAPIKeyInfo(c)
		require.Nil(t, info)

		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestRequireAPIKey_Helper(t *testing.T) {
	t.Parallel()

	validKeys := map[string]string{"helper-key-1234": "helper-client"}
	handler := RequireAPIKey(validKeys)
	require.NotNil(t, handler)

	app := fiber.New()
	app.Use(handler)
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "helper-key-1234")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestNewAPIKeyValidatorFromStore(t *testing.T) {
	t.Parallel()

	t.Run("nil_store", func(t *testing.T) {
		t.Parallel()

		validator := NewAPIKeyValidatorFromStore(nil)
		_, _, err := validator(context.Background(), "any-key")
		require.Error(t, err)
	})

	t.Run("valid_store", func(t *testing.T) {
		t.Parallel()

		store := &mockAPIKeyStore{
			keys: map[string]string{"store-key": "store-client"},
		}
		validator := NewAPIKeyValidatorFromStore(store)
		clientName, valid, err := validator(context.Background(), "store-key")
		require.NoError(t, err)
		require.True(t, valid)
		require.Equal(t, "store-client", clientName)
	})
}

// mockAPIKeyStore is a test implementation of APIKeyStore.
type mockAPIKeyStore struct {
	keys map[string]string
}

func (m *mockAPIKeyStore) GetClientByAPIKey(_ context.Context, apiKey string) (string, bool, error) {
	if clientName, ok := m.keys[apiKey]; ok {
		return clientName, true, nil
	}

	return "", false, nil
}
