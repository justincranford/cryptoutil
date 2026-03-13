// Copyright (c) 2025 Justin Cranford
//

package apis

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	cryptoutilAppsJoseJaModel "cryptoutil/internal/apps/jose/ja/model"

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
		{"RSA 2048", cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaModel.KeyTypeRSA},
		{"RSA 3072", cryptoutilSharedMagic.JoseKeyTypeRSA3072, cryptoutilAppsJoseJaModel.KeyTypeRSA},
		{"RSA 4096", cryptoutilSharedMagic.JoseKeyTypeRSA4096, cryptoutilAppsJoseJaModel.KeyTypeRSA},
		{"EC P256", cryptoutilSharedMagic.JoseKeyTypeECP256, cryptoutilAppsJoseJaModel.KeyTypeEC},
		{"EC P384", cryptoutilSharedMagic.JoseKeyTypeECP384, cryptoutilAppsJoseJaModel.KeyTypeEC},
		{"EC P521", cryptoutilSharedMagic.JoseKeyTypeECP521, cryptoutilAppsJoseJaModel.KeyTypeEC},
		{"OKP Ed25519", cryptoutilSharedMagic.JoseKeyTypeOKPEd25519, cryptoutilAppsJoseJaModel.KeyTypeOKP},
		{"OKP Ed448", "OKP/Ed448", cryptoutilAppsJoseJaModel.KeyTypeOKP},
		{"oct 256", cryptoutilSharedMagic.JoseKeyTypeOct256, cryptoutilAppsJoseJaModel.KeyTypeOct},
		{"oct 384", cryptoutilSharedMagic.JoseKeyTypeOct384, cryptoutilAppsJoseJaModel.KeyTypeOct},
		{"oct 512", cryptoutilSharedMagic.JoseKeyTypeOct512, cryptoutilAppsJoseJaModel.KeyTypeOct},
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
	resp, err := app.Test(req, -1)
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
	resp, err := app.Test(req, -1)
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
	resp, err := app.Test(req, -1)
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
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestHandleListElasticJWKs_RepositoryError(t *testing.T) {
	t.Parallel()

	handler, elasticRepo, _, _, _ := setupTestHandler()

	app := setupFiberApp()

	tenantID := googleUuid.New()

	elasticRepo.On("List", mock.Anything, tenantID, 0, cryptoutilSharedMagic.JoseJAMaxMaterials).Return(nil, int64(0), errors.New("list failed"))

	app.Get("/elastic-jwks", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleListElasticJWKs())

	req := httptest.NewRequest(fiber.MethodGet, "/elastic-jwks", nil)
	resp, err := app.Test(req, -1)
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
	resp, err := app.Test(req, -1)
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
	resp, err := app.Test(req, -1)
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
	resp, err := app.Test(req, -1)
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
	resp, err := app.Test(req, -1)
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
	resp, err := app.Test(req, -1)
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
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}
