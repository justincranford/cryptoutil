// Copyright (c) 2025 Justin Cranford
//

package apis

import (
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	cryptoutilAppsJoseJaDomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandleGetActiveMaterialJWK_Success(t *testing.T) {
	t.Parallel()

	handler, elasticRepo, materialRepo, _, _ := setupTestHandler()
	app := setupFiberApp()

	tenantID := googleUuid.New()
	kid := testElasticKID
	elasticID := googleUuid.New()

	elasticJWK := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:       elasticID,
		TenantID: tenantID,
		KID:      kid,
	}

	activeMaterial := &cryptoutilAppsJoseJaDomain.MaterialJWK{
		ID:           googleUuid.New(),
		ElasticJWKID: elasticID,
		MaterialKID:  "active-material",
		Active:       true,
		CreatedAt:    time.Now().UTC(),
	}

	elasticRepo.On("Get", mock.Anything, tenantID, kid).Return(elasticJWK, nil)
	materialRepo.On("GetActiveMaterial", mock.Anything, elasticID).Return(activeMaterial, nil)

	app.Get("/elastic-jwks/:kid/materials/active", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleGetActiveMaterialJWK())

	req := httptest.NewRequest(fiber.MethodGet, fmt.Sprintf("/elastic-jwks/%s/materials/active", kid), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.NoError(t, resp.Body.Close())

	elasticRepo.AssertExpectations(t)
	materialRepo.AssertExpectations(t)
}

func TestHandleRotateMaterialJWK_Success(t *testing.T) {
	t.Parallel()

	handler, elasticRepo, materialRepo, _, _ := setupTestHandler()
	app := setupFiberApp()

	tenantID := googleUuid.New()
	kid := testElasticKID
	elasticID := googleUuid.New()

	elasticJWK := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:                   elasticID,
		TenantID:             tenantID,
		KID:                  kid,
		MaxMaterials:         5,
		CurrentMaterialCount: 3,
	}

	elasticRepo.On("Get", mock.Anything, tenantID, kid).Return(elasticJWK, nil)
	materialRepo.On("RotateMaterial", mock.Anything, elasticID, mock.AnythingOfType("*domain.MaterialJWK")).Return(nil)
	elasticRepo.On("IncrementMaterialCount", mock.Anything, elasticID).Return(nil)

	app.Post("/elastic-jwks/:kid/materials/rotate", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleRotateMaterialJWK())

	req := httptest.NewRequest(fiber.MethodPost, fmt.Sprintf("/elastic-jwks/%s/materials/rotate", kid), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusCreated, resp.StatusCode)
	require.NoError(t, resp.Body.Close())

	elasticRepo.AssertExpectations(t)
	materialRepo.AssertExpectations(t)
}

func TestHandleRotateMaterialJWK_MaxMaterialsReached(t *testing.T) {
	t.Parallel()

	handler, elasticRepo, _, _, _ := setupTestHandler()
	app := setupFiberApp()

	tenantID := googleUuid.New()
	kid := testElasticKID

	elasticJWK := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:                   googleUuid.New(),
		TenantID:             tenantID,
		KID:                  kid,
		MaxMaterials:         5,
		CurrentMaterialCount: 5,
	}

	elasticRepo.On("Get", mock.Anything, tenantID, kid).Return(elasticJWK, nil)

	app.Post("/elastic-jwks/:kid/materials/rotate", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleRotateMaterialJWK())

	req := httptest.NewRequest(fiber.MethodPost, fmt.Sprintf("/elastic-jwks/%s/materials/rotate", kid), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusConflict, resp.StatusCode)
	require.NoError(t, resp.Body.Close())

	elasticRepo.AssertExpectations(t)
}

// ==================== Crypto Operation Handler Tests ====================

func TestHandleGetJWKS_Success(t *testing.T) {
	t.Parallel()

	handler, _, _, _, _ := setupTestHandler()
	app := setupFiberApp()

	app.Get("/.well-known/jwks.json", handler.HandleGetJWKS())

	req := httptest.NewRequest(fiber.MethodGet, "/.well-known/jwks.json", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestHandleSign_NotImplemented(t *testing.T) {
	t.Parallel()

	handler, _, _, _, _ := setupTestHandler()
	app := setupFiberApp()

	app.Post("/sign", handler.HandleSign())

	req := httptest.NewRequest(fiber.MethodPost, "/sign", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNotImplemented, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestHandleVerify_NotImplemented(t *testing.T) {
	t.Parallel()

	handler, _, _, _, _ := setupTestHandler()
	app := setupFiberApp()

	app.Post("/verify", handler.HandleVerify())

	req := httptest.NewRequest(fiber.MethodPost, "/verify", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNotImplemented, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestHandleEncrypt_NotImplemented(t *testing.T) {
	t.Parallel()

	handler, _, _, _, _ := setupTestHandler()
	app := setupFiberApp()

	app.Post("/encrypt", handler.HandleEncrypt())

	req := httptest.NewRequest(fiber.MethodPost, "/encrypt", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNotImplemented, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestHandleDecrypt_NotImplemented(t *testing.T) {
	t.Parallel()

	handler, _, _, _, _ := setupTestHandler()
	app := setupFiberApp()

	app.Post("/decrypt", handler.HandleDecrypt())

	req := httptest.NewRequest(fiber.MethodPost, "/decrypt", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNotImplemented, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

// ==================== Additional Error Path Tests ====================

func TestHandleDeleteElasticJWK_MissingKID(t *testing.T) {
	t.Parallel()

	handler, _, _, _, _ := setupTestHandler()
	app := setupFiberApp()

	tenantID := googleUuid.New()

	// Route without :kid param path - simulates empty kid.
	app.Delete("/elastic-jwks/", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleDeleteElasticJWK())

	req := httptest.NewRequest(fiber.MethodDelete, "/elastic-jwks/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestHandleDeleteElasticJWK_RepositoryDeleteError(t *testing.T) {
	t.Parallel()

	handler, elasticRepo, _, _, _ := setupTestHandler()
	app := setupFiberApp()

	tenantID := googleUuid.New()
	kid := "test-delete-error"

	elasticJWK := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:       googleUuid.New(),
		TenantID: tenantID,
		KID:      kid,
	}

	elasticRepo.On("Get", mock.Anything, tenantID, kid).Return(elasticJWK, nil)
	elasticRepo.On("Delete", mock.Anything, elasticJWK.ID).Return(errors.New("delete failed"))

	app.Delete("/elastic-jwks/:kid", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleDeleteElasticJWK())

	req := httptest.NewRequest(fiber.MethodDelete, fmt.Sprintf("/elastic-jwks/%s", kid), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	require.NoError(t, resp.Body.Close())

	elasticRepo.AssertExpectations(t)
}

func TestHandleCreateMaterialJWK_CreateRepositoryError(t *testing.T) {
	t.Parallel()

	handler, elasticRepo, materialRepo, _, _ := setupTestHandler()
	app := setupFiberApp()

	tenantID := googleUuid.New()
	kid := "test-create-error"

	elasticJWK := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:                   googleUuid.New(),
		TenantID:             tenantID,
		KID:                  kid,
		MaxMaterials:         5,
		CurrentMaterialCount: 2,
	}

	elasticRepo.On("Get", mock.Anything, tenantID, kid).Return(elasticJWK, nil)
	materialRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.MaterialJWK")).Return(errors.New("create failed"))

	app.Post("/elastic-jwks/:kid/materials", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleCreateMaterialJWK())

	req := httptest.NewRequest(fiber.MethodPost, fmt.Sprintf("/elastic-jwks/%s/materials", kid), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	require.NoError(t, resp.Body.Close())

	elasticRepo.AssertExpectations(t)
	materialRepo.AssertExpectations(t)
}

func TestHandleCreateMaterialJWK_IncrementCountError(t *testing.T) {
	t.Parallel()

	handler, elasticRepo, materialRepo, _, _ := setupTestHandler()
	app := setupFiberApp()

	tenantID := googleUuid.New()
	kid := testIncrementError

	elasticJWK := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:                   googleUuid.New(),
		TenantID:             tenantID,
		KID:                  kid,
		MaxMaterials:         5,
		CurrentMaterialCount: 2,
	}

	elasticRepo.On("Get", mock.Anything, tenantID, kid).Return(elasticJWK, nil)
	materialRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.MaterialJWK")).Return(nil)
	elasticRepo.On("IncrementMaterialCount", mock.Anything, elasticJWK.ID).Return(errors.New("increment failed"))

	app.Post("/elastic-jwks/:kid/materials", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleCreateMaterialJWK())

	req := httptest.NewRequest(fiber.MethodPost, fmt.Sprintf("/elastic-jwks/%s/materials", kid), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	require.NoError(t, resp.Body.Close())

	elasticRepo.AssertExpectations(t)
	materialRepo.AssertExpectations(t)
}

func TestHandleListMaterialJWKs_RepositoryError(t *testing.T) {
	t.Parallel()

	handler, elasticRepo, materialRepo, _, _ := setupTestHandler()
	app := setupFiberApp()

	tenantID := googleUuid.New()
	kid := "test-list-error"

	elasticJWK := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:       googleUuid.New(),
		TenantID: tenantID,
		KID:      kid,
	}

	elasticRepo.On("Get", mock.Anything, tenantID, kid).Return(elasticJWK, nil)
	materialRepo.On("ListByElasticJWK", mock.Anything, elasticJWK.ID, 0, cryptoutilSharedMagic.DefaultAPIListLimit).Return(nil, int64(0), errors.New("list failed"))

	app.Get("/elastic-jwks/:kid/materials", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleListMaterialJWKs())

	req := httptest.NewRequest(fiber.MethodGet, fmt.Sprintf("/elastic-jwks/%s/materials", kid), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	require.NoError(t, resp.Body.Close())

	elasticRepo.AssertExpectations(t)
	materialRepo.AssertExpectations(t)
}

func TestHandleGetActiveMaterialJWK_NoActiveMaterial(t *testing.T) {
	t.Parallel()

	handler, elasticRepo, materialRepo, _, _ := setupTestHandler()
	app := setupFiberApp()

	tenantID := googleUuid.New()
	kid := "test-no-active"

	elasticJWK := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:       googleUuid.New(),
		TenantID: tenantID,
		KID:      kid,
	}

	elasticRepo.On("Get", mock.Anything, tenantID, kid).Return(elasticJWK, nil)
	materialRepo.On("GetActiveMaterial", mock.Anything, elasticJWK.ID).Return(nil, errors.New("no active material"))

	app.Get("/elastic-jwks/:kid/materials/active", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleGetActiveMaterialJWK())

	req := httptest.NewRequest(fiber.MethodGet, fmt.Sprintf("/elastic-jwks/%s/materials/active", kid), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	require.NoError(t, resp.Body.Close())

	elasticRepo.AssertExpectations(t)
	materialRepo.AssertExpectations(t)
}

func TestHandleRotateMaterialJWK_RotateRepositoryError(t *testing.T) {
	t.Parallel()

	handler, elasticRepo, materialRepo, _, _ := setupTestHandler()
	app := setupFiberApp()

	tenantID := googleUuid.New()
	kid := "test-rotate-error"

	elasticJWK := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:                   googleUuid.New(),
		TenantID:             tenantID,
		KID:                  kid,
		MaxMaterials:         5,
		CurrentMaterialCount: 2,
	}

	elasticRepo.On("Get", mock.Anything, tenantID, kid).Return(elasticJWK, nil)
	materialRepo.On("RotateMaterial", mock.Anything, elasticJWK.ID, mock.AnythingOfType("*domain.MaterialJWK")).Return(errors.New("rotate failed"))

	app.Post("/elastic-jwks/:kid/materials/rotate", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleRotateMaterialJWK())

	req := httptest.NewRequest(fiber.MethodPost, fmt.Sprintf("/elastic-jwks/%s/materials/rotate", kid), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	require.NoError(t, resp.Body.Close())

	elasticRepo.AssertExpectations(t)
	materialRepo.AssertExpectations(t)
}
