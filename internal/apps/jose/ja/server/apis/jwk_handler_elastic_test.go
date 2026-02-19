// Copyright (c) 2025 Justin Cranford
//

package apis

import (
	"bytes"
	json "encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	cryptoutilAppsJoseJaDomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewJWKHandler(t *testing.T) {
	t.Parallel()

	elasticRepo := new(MockElasticJWKRepository)
	materialRepo := new(MockMaterialJWKRepository)
	auditConfigRepo := new(MockAuditConfigRepository)
	auditLogRepo := new(MockAuditLogRepository)
	jwkGenService := &cryptoutilSharedCryptoJose.JWKGenService{}
	barrierService := &cryptoutilAppsTemplateServiceServerBarrier.Service{}

	handler := NewJWKHandler(
		elasticRepo,
		materialRepo,
		auditConfigRepo,
		auditLogRepo,
		jwkGenService,
		barrierService,
	)

	require.NotNil(t, handler)
	require.Equal(t, elasticRepo, handler.elasticJWKRepo)
	require.Equal(t, materialRepo, handler.materialJWKRepo)
	require.Equal(t, auditConfigRepo, handler.auditConfigRepo)
	require.Equal(t, auditLogRepo, handler.auditLogRepo)
	require.Equal(t, jwkGenService, handler.jwkGenService)
	require.Equal(t, barrierService, handler.barrierService)
}

func TestHandleCreateElasticJWK_Success(t *testing.T) {
	t.Parallel()

	handler, elasticRepo, _, _, _ := setupTestHandler()
	app := setupFiberApp()

	tenantID := googleUuid.New()

	// Mock repository.
	elasticRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.ElasticJWK")).
		Return(nil)

	// Setup route with middleware.
	app.Post("/jwk", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleCreateElasticJWK())

	// Prepare request.
	reqBody := CreateElasticJWKRequest{
		Algorithm:    "RSA/2048",
		Use:          "sig",
		MaxMaterials: 10,
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/jwk", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Send request.
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusCreated, resp.StatusCode)

	// Parse response.
	var response ElasticJWKResponse

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response))
	require.NotEmpty(t, response.KID)
	require.Equal(t, tenantID.String(), response.TenantID)
	require.Equal(t, "RSA", response.KeyType)
	require.Equal(t, "RSA/2048", response.Algorithm)
	require.Equal(t, "sig", response.Use)
	require.Equal(t, 10, response.MaxMaterials)

	elasticRepo.AssertExpectations(t)
}

func TestHandleCreateElasticJWK_MissingTenantContext(t *testing.T) {
	t.Parallel()

	handler, _, _, _, _ := setupTestHandler()
	app := setupFiberApp()

	// No tenant/realm context set.
	app.Post("/jwk", handler.HandleCreateElasticJWK())

	reqBody := CreateElasticJWKRequest{
		Algorithm:    "RSA/2048",
		Use:          "sig",
		MaxMaterials: 10,
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/jwk", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestHandleCreateElasticJWK_InvalidAlgorithm(t *testing.T) {
	t.Parallel()

	handler, _, _, _, _ := setupTestHandler()
	app := setupFiberApp()

	tenantID := googleUuid.New()

	app.Post("/jwk", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleCreateElasticJWK())

	reqBody := CreateElasticJWKRequest{
		Algorithm:    "INVALID",
		Use:          "sig",
		MaxMaterials: 10,
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/jwk", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandleCreateElasticJWK_RepositoryError(t *testing.T) {
	t.Parallel()

	handler, elasticRepo, _, _, _ := setupTestHandler()
	app := setupFiberApp()

	tenantID := googleUuid.New()

	// Mock repository error.
	elasticRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.ElasticJWK")).
		Return(errors.New("database error"))

	app.Post("/jwk", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleCreateElasticJWK())

	reqBody := CreateElasticJWKRequest{
		Algorithm:    "RSA/2048",
		Use:          "sig",
		MaxMaterials: 10,
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/jwk", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	elasticRepo.AssertExpectations(t)
}

func TestHandleGetElasticJWK_Success(t *testing.T) {
	t.Parallel()

	handler, elasticRepo, _, _, _ := setupTestHandler()
	app := setupFiberApp()

	tenantID := googleUuid.New()
	kid := googleUuid.New()

	// Mock repository.
	expectedJWK := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:                   kid,
		TenantID:             tenantID,
		KID:                  kid.String(),
		KeyType:              "RSA",
		Algorithm:            "RSA/2048",
		Use:                  "sig",
		MaxMaterials:         10,
		CurrentMaterialCount: 1,
		CreatedAt:            time.Now().UTC(),
	}
	elasticRepo.On("Get", mock.Anything, tenantID, kid.String()).
		Return(expectedJWK, nil)

	app.Get("/jwk/:kid", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleGetElasticJWK())

	req := httptest.NewRequest("GET", "/jwk/"+kid.String(), nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response ElasticJWKResponse

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response))
	require.Equal(t, kid.String(), response.KID)
	require.Equal(t, tenantID.String(), response.TenantID)

	elasticRepo.AssertExpectations(t)
}

func TestHandleGetElasticJWK_NotFound(t *testing.T) {
	t.Parallel()

	handler, elasticRepo, _, _, _ := setupTestHandler()
	app := setupFiberApp()

	tenantID := googleUuid.New()
	kid := googleUuid.New()

	// Mock repository not found.
	elasticRepo.On("Get", mock.Anything, tenantID, kid.String()).
		Return(nil, errors.New("not found"))

	app.Get("/jwk/:kid", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleGetElasticJWK())

	req := httptest.NewRequest("GET", "/jwk/"+kid.String(), nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	elasticRepo.AssertExpectations(t)
}

func TestHandleListElasticJWKs_Success(t *testing.T) {
	t.Parallel()

	handler, elasticRepo, _, _, _ := setupTestHandler()
	app := setupFiberApp()

	tenantID := googleUuid.New()

	// Mock repository.
	expectedJWKs := []*cryptoutilAppsJoseJaDomain.ElasticJWK{
		{
			ID:                   googleUuid.New(),
			TenantID:             tenantID,
			KID:                  "kid1",
			KeyType:              "RSA",
			Algorithm:            "RSA/2048",
			Use:                  "sig",
			MaxMaterials:         10,
			CurrentMaterialCount: 1,
			CreatedAt:            time.Now().UTC(),
		},
		{
			ID:                   googleUuid.New(),
			TenantID:             tenantID,
			KID:                  "kid2",
			KeyType:              "EC",
			Algorithm:            "EC/P256",
			Use:                  "enc",
			MaxMaterials:         5,
			CurrentMaterialCount: 0,
			CreatedAt:            time.Now().UTC(),
		},
	}
	elasticRepo.On("List", mock.Anything, tenantID, 0, 100).
		Return(expectedJWKs, int64(2), nil)

	app.Get("/jwks", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleListElasticJWKs())

	req := httptest.NewRequest("GET", "/jwks", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response ListResponse

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response))
	require.Equal(t, int64(2), response.Total)

	elasticRepo.AssertExpectations(t)
}

func TestHandleDeleteElasticJWK_Success(t *testing.T) {
	t.Parallel()

	handler, elasticRepo, _, _, _ := setupTestHandler()
	app := setupFiberApp()

	tenantID := googleUuid.New()
	kid := googleUuid.New()

	// Mock repository.
	existingJWK := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:       kid,
		TenantID: tenantID,
		KID:      kid.String(),
	}
	elasticRepo.On("Get", mock.Anything, tenantID, kid.String()).
		Return(existingJWK, nil)
	elasticRepo.On("Delete", mock.Anything, kid).
		Return(nil)

	app.Delete("/jwk/:kid", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleDeleteElasticJWK())

	req := httptest.NewRequest("DELETE", "/jwk/"+kid.String(), nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNoContent, resp.StatusCode)

	elasticRepo.AssertExpectations(t)
}

// ==================== Material JWK Handler Tests ====================

func TestHandleCreateMaterialJWK_Success(t *testing.T) {
	t.Parallel()

	handler, elasticRepo, materialRepo, _, _ := setupTestHandler()
	app := setupFiberApp()

	tenantID := googleUuid.New()
	kid := testElasticKID

	elasticJWK := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:                   googleUuid.New(),
		TenantID:             tenantID,
		KID:                  kid,
		MaxMaterials:         5,
		CurrentMaterialCount: 2,
	}

	elasticRepo.On("Get", mock.Anything, tenantID, kid).Return(elasticJWK, nil)
	materialRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.MaterialJWK")).Return(nil)
	elasticRepo.On("IncrementMaterialCount", mock.Anything, elasticJWK.ID).Return(nil)

	app.Post("/elastic-jwks/:kid/materials", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleCreateMaterialJWK())

	req := httptest.NewRequest(fiber.MethodPost, fmt.Sprintf("/elastic-jwks/%s/materials", kid), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusCreated, resp.StatusCode)
	require.NoError(t, resp.Body.Close())

	elasticRepo.AssertExpectations(t)
	materialRepo.AssertExpectations(t)
}

func TestHandleCreateMaterialJWK_MaxMaterialsReached(t *testing.T) {
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

	app.Post("/elastic-jwks/:kid/materials", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleCreateMaterialJWK())

	req := httptest.NewRequest(fiber.MethodPost, fmt.Sprintf("/elastic-jwks/%s/materials", kid), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusConflict, resp.StatusCode)
	require.NoError(t, resp.Body.Close())

	elasticRepo.AssertExpectations(t)
}

func TestHandleListMaterialJWKs_Success(t *testing.T) {
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

	materials := []*cryptoutilAppsJoseJaDomain.MaterialJWK{
		{
			ID:           googleUuid.New(),
			ElasticJWKID: elasticID,
			MaterialKID:  "material-1",
			Active:       true,
			CreatedAt:    time.Now().UTC(),
		},
	}

	elasticRepo.On("Get", mock.Anything, tenantID, kid).Return(elasticJWK, nil)
	materialRepo.On("ListByElasticJWK", mock.Anything, elasticID, 0, cryptoutilSharedMagic.DefaultAPIListLimit).Return(materials, int64(1), nil)

	app.Get("/elastic-jwks/:kid/materials", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleListMaterialJWKs())

	req := httptest.NewRequest(fiber.MethodGet, fmt.Sprintf("/elastic-jwks/%s/materials", kid), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.NoError(t, resp.Body.Close())

	elasticRepo.AssertExpectations(t)
	materialRepo.AssertExpectations(t)
}
