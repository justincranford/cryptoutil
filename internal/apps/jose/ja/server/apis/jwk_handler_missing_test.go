// Copyright (c) 2025 Justin Cranford
//

package apis

import (
	"errors"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandleCreateMaterialJWK_ElasticJWKNotFound(t *testing.T) {
	t.Parallel()

	handler, elasticRepo, _, _, _ := setupTestHandler()

	app := setupFiberApp()

	tenantID := googleUuid.New()

	elasticRepo.On("Get", mock.Anything, tenantID, "nonexistent").Return(nil, errors.New("not found"))

	app.Post("/elastic-jwks/:kid/materials", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleCreateMaterialJWK())

	req := httptest.NewRequest(fiber.MethodPost, "/elastic-jwks/nonexistent/materials", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	require.NoError(t, resp.Body.Close())

	elasticRepo.AssertExpectations(t)
}

func TestHandleListMaterialJWKs_MissingContext(t *testing.T) {
	t.Parallel()

	handler, _, _, _, _ := setupTestHandler()

	app := setupFiberApp()

	// Test missing context.
	app.Get("/elastic-jwks/:kid/materials", handler.HandleListMaterialJWKs())

	req := httptest.NewRequest(fiber.MethodGet, "/elastic-jwks/test-kid/materials", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestHandleListMaterialJWKs_InvalidTenantFormat(t *testing.T) {
	t.Parallel()

	handler, _, _, _, _ := setupTestHandler()

	app := setupFiberApp()

	app.Get("/elastic-jwks/:kid/materials", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", "not-a-uuid")
		c.Locals("realm_id", googleUuid.New())

		return c.Next()
	}, handler.HandleListMaterialJWKs())

	req := httptest.NewRequest(fiber.MethodGet, "/elastic-jwks/test-kid/materials", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestHandleListMaterialJWKs_ElasticJWKNotFound(t *testing.T) {
	t.Parallel()

	handler, elasticRepo, _, _, _ := setupTestHandler()

	app := setupFiberApp()

	tenantID := googleUuid.New()

	elasticRepo.On("Get", mock.Anything, tenantID, "nonexistent").Return(nil, errors.New("not found"))

	app.Get("/elastic-jwks/:kid/materials", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleListMaterialJWKs())

	req := httptest.NewRequest(fiber.MethodGet, "/elastic-jwks/nonexistent/materials", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	require.NoError(t, resp.Body.Close())

	elasticRepo.AssertExpectations(t)
}

func TestHandleGetActiveMaterialJWK_MissingContext(t *testing.T) {
	t.Parallel()

	handler, _, _, _, _ := setupTestHandler()

	app := setupFiberApp()

	// Test missing context.
	app.Get("/elastic-jwks/:kid/materials/active", handler.HandleGetActiveMaterialJWK())

	req := httptest.NewRequest(fiber.MethodGet, "/elastic-jwks/test-kid/materials/active", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestHandleGetActiveMaterialJWK_InvalidTenantFormat(t *testing.T) {
	t.Parallel()

	handler, _, _, _, _ := setupTestHandler()

	app := setupFiberApp()

	app.Get("/elastic-jwks/:kid/materials/active", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", "not-a-uuid")
		c.Locals("realm_id", googleUuid.New())

		return c.Next()
	}, handler.HandleGetActiveMaterialJWK())

	req := httptest.NewRequest(fiber.MethodGet, "/elastic-jwks/test-kid/materials/active", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestHandleGetActiveMaterialJWK_ElasticJWKNotFound(t *testing.T) {
	t.Parallel()

	handler, elasticRepo, _, _, _ := setupTestHandler()

	app := setupFiberApp()

	tenantID := googleUuid.New()

	elasticRepo.On("Get", mock.Anything, tenantID, "nonexistent").Return(nil, errors.New("not found"))

	app.Get("/elastic-jwks/:kid/materials/active", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleGetActiveMaterialJWK())

	req := httptest.NewRequest(fiber.MethodGet, "/elastic-jwks/nonexistent/materials/active", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	require.NoError(t, resp.Body.Close())

	elasticRepo.AssertExpectations(t)
}

func TestHandleRotateMaterialJWK_MissingContext(t *testing.T) {
	t.Parallel()

	handler, _, _, _, _ := setupTestHandler()

	app := setupFiberApp()

	// Test missing context.
	app.Post("/elastic-jwks/:kid/materials/rotate", handler.HandleRotateMaterialJWK())

	req := httptest.NewRequest(fiber.MethodPost, "/elastic-jwks/test-kid/materials/rotate", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestHandleRotateMaterialJWK_InvalidTenantFormat(t *testing.T) {
	t.Parallel()

	handler, _, _, _, _ := setupTestHandler()

	app := setupFiberApp()

	app.Post("/elastic-jwks/:kid/materials/rotate", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", "not-a-uuid")
		c.Locals("realm_id", googleUuid.New())

		return c.Next()
	}, handler.HandleRotateMaterialJWK())

	req := httptest.NewRequest(fiber.MethodPost, "/elastic-jwks/test-kid/materials/rotate", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestHandleRotateMaterialJWK_ElasticJWKNotFound(t *testing.T) {
	t.Parallel()

	handler, elasticRepo, _, _, _ := setupTestHandler()

	app := setupFiberApp()

	tenantID := googleUuid.New()

	elasticRepo.On("Get", mock.Anything, tenantID, "nonexistent").Return(nil, errors.New("not found"))

	app.Post("/elastic-jwks/:kid/materials/rotate", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleRotateMaterialJWK())

	req := httptest.NewRequest(fiber.MethodPost, "/elastic-jwks/nonexistent/materials/rotate", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	require.NoError(t, resp.Body.Close())

	elasticRepo.AssertExpectations(t)
}

// ==================== Additional Coverage Tests ====================
