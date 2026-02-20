// Copyright (c) 2025 Justin Cranford
//

package apis

import (
	"bytes"
	json "encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	cryptoutilAppsJoseJaDomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandleListMaterialJWKs_WithRetiredMaterial(t *testing.T) {
	t.Parallel()

	handler, elasticRepo, materialRepo, _, _ := setupTestHandler()
	app := setupFiberApp()

	tenantID := googleUuid.New()
	kid := "test-retired-material"
	elasticID := googleUuid.New()

	elasticJWK := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:       elasticID,
		TenantID: tenantID,
		KID:      kid,
	}

	retiredAt := time.Now().UTC().Add(-24 * time.Hour)
	materials := []*cryptoutilAppsJoseJaDomain.MaterialJWK{
		{
			ID:           googleUuid.New(),
			ElasticJWKID: elasticID,
			MaterialKID:  "material-retired",
			Active:       false,
			RetiredAt:    &retiredAt,
			CreatedAt:    time.Now().UTC().Add(-48 * time.Hour),
		},
	}

	elasticRepo.On("Get", mock.Anything, tenantID, kid).Return(elasticJWK, nil)
	materialRepo.On("ListByElasticJWK", mock.Anything, elasticID, 0, cryptoutilSharedMagic.DefaultAPIListLimit).Return(materials, int64(1), nil)

	app.Get("/elastic-jwks/:kid/materials", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleListMaterialJWKs())

	req := httptest.NewRequest(fiber.MethodGet, fmt.Sprintf("/elastic-jwks/%s/materials", kid), nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response ListResponse

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response))
	require.Equal(t, int64(1), response.Total)

	elasticRepo.AssertExpectations(t)
	materialRepo.AssertExpectations(t)
}

func TestHandleRotateMaterialJWK_IncrementCountError(t *testing.T) {
	t.Parallel()

	handler, elasticRepo, materialRepo, _, _ := setupTestHandler()
	app := setupFiberApp()

	tenantID := googleUuid.New()
	kid := testIncrementError
	elasticID := googleUuid.New()

	elasticJWK := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:                   elasticID,
		TenantID:             tenantID,
		KID:                  kid,
		MaxMaterials:         5,
		CurrentMaterialCount: 2,
	}

	elasticRepo.On("Get", mock.Anything, tenantID, kid).Return(elasticJWK, nil)
	materialRepo.On("RotateMaterial", mock.Anything, elasticID, mock.AnythingOfType("*domain.MaterialJWK")).Return(nil)
	elasticRepo.On("IncrementMaterialCount", mock.Anything, elasticID).Return(errors.New("increment failed"))

	app.Post("/elastic-jwks/:kid/materials/rotate", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleRotateMaterialJWK())

	req := httptest.NewRequest(fiber.MethodPost, fmt.Sprintf("/elastic-jwks/%s/materials/rotate", kid), nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	require.NoError(t, resp.Body.Close())

	elasticRepo.AssertExpectations(t)
	materialRepo.AssertExpectations(t)
}

func TestHandleGetElasticJWK_MissingKID(t *testing.T) {
	t.Parallel()

	handler, _, _, _, _ := setupTestHandler()
	app := setupFiberApp()

	tenantID := googleUuid.New()

	// Route without :kid param - simulates empty kid.
	app.Get("/elastic-jwks/", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleGetElasticJWK())

	req := httptest.NewRequest(fiber.MethodGet, "/elastic-jwks/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestHandleCreateElasticJWK_InvalidBody(t *testing.T) {
	t.Parallel()

	handler, _, _, _, _ := setupTestHandler()
	app := setupFiberApp()

	tenantID := googleUuid.New()

	app.Post("/elastic-jwks", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleCreateElasticJWK())

	// Invalid JSON body.
	req := httptest.NewRequest(fiber.MethodPost, "/elastic-jwks", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestHandleCreateElasticJWK_DefaultMaxMaterials(t *testing.T) {
	t.Parallel()

	handler, elasticRepo, _, _, _ := setupTestHandler()
	app := setupFiberApp()

	tenantID := googleUuid.New()

	// Mock repository.
	elasticRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.ElasticJWK")).
		Return(nil)

	app.Post("/elastic-jwks", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleCreateElasticJWK())

	// Request with max_materials = 0 (should default to 10).
	reqBody := CreateElasticJWKRequest{
		Algorithm:    "RSA/2048",
		Use:          "sig",
		MaxMaterials: 0,
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(fiber.MethodPost, "/elastic-jwks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusCreated, resp.StatusCode)

	var response ElasticJWKResponse

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response))
	require.Equal(t, 10, response.MaxMaterials)

	elasticRepo.AssertExpectations(t)
}

func TestHandleDeleteElasticJWK_NotFound(t *testing.T) {
	t.Parallel()

	handler, elasticRepo, _, _, _ := setupTestHandler()
	app := setupFiberApp()

	tenantID := googleUuid.New()

	elasticRepo.On("Get", mock.Anything, tenantID, "nonexistent").Return(nil, errors.New("not found"))

	app.Delete("/elastic-jwks/:kid", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleDeleteElasticJWK())

	req := httptest.NewRequest(fiber.MethodDelete, "/elastic-jwks/nonexistent", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	require.NoError(t, resp.Body.Close())

	elasticRepo.AssertExpectations(t)
}

func TestHandleCreateMaterialJWK_MissingKID(t *testing.T) {
	t.Parallel()

	handler, _, _, _, _ := setupTestHandler()
	app := setupFiberApp()

	tenantID := googleUuid.New()

	// Route without :kid param - simulates empty kid.
	app.Post("/elastic-jwks//materials", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleCreateMaterialJWK())

	req := httptest.NewRequest(fiber.MethodPost, "/elastic-jwks//materials", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestHandleListMaterialJWKs_MissingKID(t *testing.T) {
	t.Parallel()

	handler, _, _, _, _ := setupTestHandler()
	app := setupFiberApp()

	tenantID := googleUuid.New()

	// Route without :kid param - simulates empty kid.
	app.Get("/elastic-jwks//materials", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleListMaterialJWKs())

	req := httptest.NewRequest(fiber.MethodGet, "/elastic-jwks//materials", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestHandleGetActiveMaterialJWK_MissingKID(t *testing.T) {
	t.Parallel()

	handler, _, _, _, _ := setupTestHandler()
	app := setupFiberApp()

	tenantID := googleUuid.New()

	// Route without :kid param - simulates empty kid.
	app.Get("/elastic-jwks//materials/active", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleGetActiveMaterialJWK())

	req := httptest.NewRequest(fiber.MethodGet, "/elastic-jwks//materials/active", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestHandleRotateMaterialJWK_MissingKID(t *testing.T) {
	t.Parallel()

	handler, _, _, _, _ := setupTestHandler()
	app := setupFiberApp()

	tenantID := googleUuid.New()

	// Route without :kid param - simulates empty kid.
	app.Post("/elastic-jwks//materials/rotate", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleRotateMaterialJWK())

	req := httptest.NewRequest(fiber.MethodPost, "/elastic-jwks//materials/rotate", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}
