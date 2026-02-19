// Copyright (c) 2025 Justin Cranford
//

package apis

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	cryptoutilAppsJoseJaDomain "cryptoutil/internal/apps/jose/ja/domain"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMapAlgorithmToKeyType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		algorithm string
		expected  string
	}{
		{"RSA 2048", "RSA/2048", cryptoutilAppsJoseJaDomain.KeyTypeRSA},
		{"RSA 3072", "RSA/3072", cryptoutilAppsJoseJaDomain.KeyTypeRSA},
		{"RSA 4096", "RSA/4096", cryptoutilAppsJoseJaDomain.KeyTypeRSA},
		{"EC P256", "EC/P256", cryptoutilAppsJoseJaDomain.KeyTypeEC},
		{"EC P384", "EC/P384", cryptoutilAppsJoseJaDomain.KeyTypeEC},
		{"EC P521", "EC/P521", cryptoutilAppsJoseJaDomain.KeyTypeEC},
		{"OKP Ed25519", "OKP/Ed25519", cryptoutilAppsJoseJaDomain.KeyTypeOKP},
		{"OKP Ed448", "OKP/Ed448", cryptoutilAppsJoseJaDomain.KeyTypeOKP},
		{"oct 256", "oct/256", cryptoutilAppsJoseJaDomain.KeyTypeOct},
		{"oct 384", "oct/384", cryptoutilAppsJoseJaDomain.KeyTypeOct},
		{"oct 512", "oct/512", cryptoutilAppsJoseJaDomain.KeyTypeOct},
		{"unknown", "unknown", ""},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := mapAlgorithmToKeyType(tt.algorithm)
			require.Equal(t, tt.expected, result)
		})
	}
}

// Tests for missing context error paths.

func TestHandleGetElasticJWK_MissingContext(t *testing.T) {
	t.Parallel()

	handler, _, _, _, _ := setupTestHandler()

	app := setupFiberApp()

	// Test missing both tenant and realm.
	app.Get("/elastic-jwks/:kid", handler.HandleGetElasticJWK())

	req := httptest.NewRequest(fiber.MethodGet, "/elastic-jwks/test-kid", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestHandleGetElasticJWK_InvalidTenantFormat(t *testing.T) {
	t.Parallel()

	handler, _, _, _, _ := setupTestHandler()

	app := setupFiberApp()

	// Set invalid tenant_id format (string instead of UUID).
	app.Get("/elastic-jwks/:kid", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", "not-a-uuid")
		c.Locals("realm_id", googleUuid.New())

		return c.Next()
	}, handler.HandleGetElasticJWK())

	req := httptest.NewRequest(fiber.MethodGet, "/elastic-jwks/test-kid", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestHandleListElasticJWKs_MissingContext(t *testing.T) {
	t.Parallel()

	handler, _, _, _, _ := setupTestHandler()

	app := setupFiberApp()

	// Test missing context.
	app.Get("/elastic-jwks", handler.HandleListElasticJWKs())

	req := httptest.NewRequest(fiber.MethodGet, "/elastic-jwks", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestHandleListElasticJWKs_InvalidTenantFormat(t *testing.T) {
	t.Parallel()

	handler, _, _, _, _ := setupTestHandler()

	app := setupFiberApp()

	app.Get("/elastic-jwks", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", "not-a-uuid")
		c.Locals("realm_id", googleUuid.New())

		return c.Next()
	}, handler.HandleListElasticJWKs())

	req := httptest.NewRequest(fiber.MethodGet, "/elastic-jwks", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestHandleListElasticJWKs_RepositoryError(t *testing.T) {
	t.Parallel()

	handler, elasticRepo, _, _, _ := setupTestHandler()

	app := setupFiberApp()

	tenantID := googleUuid.New()

	elasticRepo.On("List", mock.Anything, tenantID, 0, 100).Return(nil, int64(0), errors.New("list failed"))

	app.Get("/elastic-jwks", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleListElasticJWKs())

	req := httptest.NewRequest(fiber.MethodGet, "/elastic-jwks", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	require.NoError(t, resp.Body.Close())

	elasticRepo.AssertExpectations(t)
}

func TestHandleDeleteElasticJWK_MissingContext(t *testing.T) {
	t.Parallel()

	handler, _, _, _, _ := setupTestHandler()

	app := setupFiberApp()

	// Test missing context (after KID validation passes).
	app.Delete("/elastic-jwks/:kid", handler.HandleDeleteElasticJWK())

	req := httptest.NewRequest(fiber.MethodDelete, "/elastic-jwks/test-kid", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestHandleDeleteElasticJWK_InvalidTenantFormat(t *testing.T) {
	t.Parallel()

	handler, _, _, _, _ := setupTestHandler()

	app := setupFiberApp()

	app.Delete("/elastic-jwks/:kid", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", "not-a-uuid")
		c.Locals("realm_id", googleUuid.New())

		return c.Next()
	}, handler.HandleDeleteElasticJWK())

	req := httptest.NewRequest(fiber.MethodDelete, "/elastic-jwks/test-kid", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestHandleCreateElasticJWK_MissingContext(t *testing.T) {
	t.Parallel()

	handler, _, _, _, _ := setupTestHandler()

	app := setupFiberApp()

	// Test missing context.
	app.Post("/elastic-jwks", handler.HandleCreateElasticJWK())

	reqBody := testCreateJWKBody
	req := httptest.NewRequest(fiber.MethodPost, "/elastic-jwks", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestHandleCreateElasticJWK_InvalidTenantFormat(t *testing.T) {
	t.Parallel()

	handler, _, _, _, _ := setupTestHandler()

	app := setupFiberApp()

	app.Post("/elastic-jwks", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", "not-a-uuid")
		c.Locals("realm_id", googleUuid.New())

		return c.Next()
	}, handler.HandleCreateElasticJWK())

	reqBody := testCreateJWKBody
	req := httptest.NewRequest(fiber.MethodPost, "/elastic-jwks", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestHandleCreateMaterialJWK_MissingContext(t *testing.T) {
	t.Parallel()

	handler, _, _, _, _ := setupTestHandler()

	app := setupFiberApp()

	// Test missing context.
	app.Post("/elastic-jwks/:kid/materials", handler.HandleCreateMaterialJWK())

	req := httptest.NewRequest(fiber.MethodPost, "/elastic-jwks/test-kid/materials", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestHandleCreateMaterialJWK_InvalidTenantFormat(t *testing.T) {
	t.Parallel()

	handler, _, _, _, _ := setupTestHandler()

	app := setupFiberApp()

	app.Post("/elastic-jwks/:kid/materials", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", "not-a-uuid")
		c.Locals("realm_id", googleUuid.New())

		return c.Next()
	}, handler.HandleCreateMaterialJWK())

	req := httptest.NewRequest(fiber.MethodPost, "/elastic-jwks/test-kid/materials", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}
