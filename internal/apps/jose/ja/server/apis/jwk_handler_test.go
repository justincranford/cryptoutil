// Copyright (c) 2025 Justin Cranford
//

package apis

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	joseJADomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"

	"github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockElasticJWKRepository is a mock implementation of ElasticJWKRepository.
type MockElasticJWKRepository struct {
	mock.Mock
}

func (m *MockElasticJWKRepository) Create(ctx context.Context, jwk *joseJADomain.ElasticJWK) error {
	args := m.Called(ctx, jwk)
	return args.Error(0)
}

func (m *MockElasticJWKRepository) Get(ctx context.Context, tenantID, realmID googleUuid.UUID, kid string) (*joseJADomain.ElasticJWK, error) {
	args := m.Called(ctx, tenantID, realmID, kid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*joseJADomain.ElasticJWK), args.Error(1)
}

func (m *MockElasticJWKRepository) GetByID(ctx context.Context, id googleUuid.UUID) (*joseJADomain.ElasticJWK, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*joseJADomain.ElasticJWK), args.Error(1)
}

func (m *MockElasticJWKRepository) List(ctx context.Context, tenantID, realmID googleUuid.UUID, offset, limit int) ([]*joseJADomain.ElasticJWK, int64, error) {
	args := m.Called(ctx, tenantID, realmID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*joseJADomain.ElasticJWK), args.Get(1).(int64), args.Error(2)
}

func (m *MockElasticJWKRepository) Update(ctx context.Context, jwk *joseJADomain.ElasticJWK) error {
	args := m.Called(ctx, jwk)
	return args.Error(0)
}

func (m *MockElasticJWKRepository) Delete(ctx context.Context, id googleUuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockElasticJWKRepository) IncrementMaterialCount(ctx context.Context, id googleUuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockElasticJWKRepository) DecrementMaterialCount(ctx context.Context, id googleUuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockMaterialJWKRepository is a mock implementation of MaterialJWKRepository.
type MockMaterialJWKRepository struct {
	mock.Mock
}

func (m *MockMaterialJWKRepository) Create(ctx context.Context, jwk *joseJADomain.MaterialJWK) error {
	args := m.Called(ctx, jwk)
	return args.Error(0)
}

func (m *MockMaterialJWKRepository) GetByMaterialKID(ctx context.Context, materialKID string) (*joseJADomain.MaterialJWK, error) {
	args := m.Called(ctx, materialKID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*joseJADomain.MaterialJWK), args.Error(1)
}

func (m *MockMaterialJWKRepository) GetByID(ctx context.Context, id googleUuid.UUID) (*joseJADomain.MaterialJWK, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*joseJADomain.MaterialJWK), args.Error(1)
}

func (m *MockMaterialJWKRepository) GetActiveMaterial(ctx context.Context, elasticJWKID googleUuid.UUID) (*joseJADomain.MaterialJWK, error) {
	args := m.Called(ctx, elasticJWKID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*joseJADomain.MaterialJWK), args.Error(1)
}

func (m *MockMaterialJWKRepository) ListByElasticJWK(ctx context.Context, elasticJWKID googleUuid.UUID, offset, limit int) ([]*joseJADomain.MaterialJWK, int64, error) {
	args := m.Called(ctx, elasticJWKID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*joseJADomain.MaterialJWK), args.Get(1).(int64), args.Error(2)
}

func (m *MockMaterialJWKRepository) RotateMaterial(ctx context.Context, elasticJWKID googleUuid.UUID, newMaterial *joseJADomain.MaterialJWK) error {
	args := m.Called(ctx, elasticJWKID, newMaterial)
	return args.Error(0)
}

func (m *MockMaterialJWKRepository) RetireMaterial(ctx context.Context, id googleUuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMaterialJWKRepository) Delete(ctx context.Context, id googleUuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMaterialJWKRepository) CountMaterials(ctx context.Context, elasticJWKID googleUuid.UUID) (int64, error) {
	args := m.Called(ctx, elasticJWKID)
	return args.Get(0).(int64), args.Error(1)
}

// MockAuditConfigRepository is a mock implementation of AuditConfigRepository.
type MockAuditConfigRepository struct {
	mock.Mock
}

func (m *MockAuditConfigRepository) Get(ctx context.Context, tenantID googleUuid.UUID, operation string) (*joseJADomain.AuditConfig, error) {
	args := m.Called(ctx, tenantID, operation)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*joseJADomain.AuditConfig), args.Error(1)
}

func (m *MockAuditConfigRepository) GetAllForTenant(ctx context.Context, tenantID googleUuid.UUID) ([]*joseJADomain.AuditConfig, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*joseJADomain.AuditConfig), args.Error(1)
}

func (m *MockAuditConfigRepository) Upsert(ctx context.Context, config *joseJADomain.AuditConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockAuditConfigRepository) Delete(ctx context.Context, tenantID googleUuid.UUID, operation string) error {
	args := m.Called(ctx, tenantID, operation)
	return args.Error(0)
}

func (m *MockAuditConfigRepository) ShouldAudit(ctx context.Context, tenantID googleUuid.UUID, operation string) (bool, error) {
	args := m.Called(ctx, tenantID, operation)
	return args.Bool(0), args.Error(1)
}

// MockAuditLogRepository is a mock implementation of AuditLogRepository.
type MockAuditLogRepository struct {
	mock.Mock
}

func (m *MockAuditLogRepository) Create(ctx context.Context, entry *joseJADomain.AuditLogEntry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

func (m *MockAuditLogRepository) List(ctx context.Context, tenantID googleUuid.UUID, offset, limit int) ([]*joseJADomain.AuditLogEntry, int64, error) {
	args := m.Called(ctx, tenantID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*joseJADomain.AuditLogEntry), args.Get(1).(int64), args.Error(2)
}

func (m *MockAuditLogRepository) ListByElasticJWK(ctx context.Context, elasticJWKID googleUuid.UUID, offset, limit int) ([]*joseJADomain.AuditLogEntry, int64, error) {
	args := m.Called(ctx, elasticJWKID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*joseJADomain.AuditLogEntry), args.Get(1).(int64), args.Error(2)
}

func (m *MockAuditLogRepository) ListByOperation(ctx context.Context, tenantID googleUuid.UUID, operation string, offset, limit int) ([]*joseJADomain.AuditLogEntry, int64, error) {
	args := m.Called(ctx, tenantID, operation, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*joseJADomain.AuditLogEntry), args.Get(1).(int64), args.Error(2)
}

func (m *MockAuditLogRepository) GetByRequestID(ctx context.Context, requestID string) (*joseJADomain.AuditLogEntry, error) {
	args := m.Called(ctx, requestID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*joseJADomain.AuditLogEntry), args.Error(1)
}

func (m *MockAuditLogRepository) DeleteOlderThan(ctx context.Context, tenantID googleUuid.UUID, days int) (int64, error) {
	args := m.Called(ctx, tenantID, days)
	return args.Get(0).(int64), args.Error(1)
}

// setupTestHandler creates a test handler with mocks.
func setupTestHandler() (*JWKHandler, *MockElasticJWKRepository, *MockMaterialJWKRepository, *MockAuditConfigRepository, *MockAuditLogRepository) {
	elasticRepo := new(MockElasticJWKRepository)
	materialRepo := new(MockMaterialJWKRepository)
	auditConfigRepo := new(MockAuditConfigRepository)
	auditLogRepo := new(MockAuditLogRepository)

	// Create minimal JWKGenService for testing.
	jwkGenService := &cryptoutilJose.JWKGenService{}

	// Create minimal BarrierService for testing.
	barrierService := &cryptoutilBarrier.BarrierService{}

	handler := NewJWKHandler(
		elasticRepo,
		materialRepo,
		auditConfigRepo,
		auditLogRepo,
		jwkGenService,
		barrierService,
	)

	return handler, elasticRepo, materialRepo, auditConfigRepo, auditLogRepo
}

// setupFiberApp creates a Fiber app for testing.
func setupFiberApp() *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})
	return app
}

func TestNewJWKHandler(t *testing.T) {
	t.Parallel()

	elasticRepo := new(MockElasticJWKRepository)
	materialRepo := new(MockMaterialJWKRepository)
	auditConfigRepo := new(MockAuditConfigRepository)
	auditLogRepo := new(MockAuditLogRepository)
	jwkGenService := &cryptoutilJose.JWKGenService{}
	barrierService := &cryptoutilBarrier.BarrierService{}

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
	realmID := googleUuid.New()

	// Mock repository.
	elasticRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.ElasticJWK")).
		Return(nil)

	// Setup route with middleware.
	app.Post("/jwk", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)
		c.Locals("realm_id", realmID)
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
	require.Equal(t, realmID.String(), response.RealmID)
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
	realmID := googleUuid.New()

	app.Post("/jwk", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)
		c.Locals("realm_id", realmID)
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
	realmID := googleUuid.New()

	// Mock repository error.
	elasticRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.ElasticJWK")).
		Return(errors.New("database error"))

	app.Post("/jwk", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)
		c.Locals("realm_id", realmID)
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
	realmID := googleUuid.New()
	kid := googleUuid.New()

	// Mock repository.
	expectedJWK := &joseJADomain.ElasticJWK{
		ID:                   kid,
		TenantID:             tenantID,
		RealmID:              realmID,
		KID:                  kid.String(),
		KeyType:              "RSA",
		Algorithm:            "RSA/2048",
		Use:                  "sig",
		MaxMaterials:         10,
		CurrentMaterialCount: 1,
		CreatedAt:            time.Now(),
	}
	elasticRepo.On("Get", mock.Anything, tenantID, realmID, kid.String()).
		Return(expectedJWK, nil)

	app.Get("/jwk/:kid", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)
		c.Locals("realm_id", realmID)
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
	require.Equal(t, realmID.String(), response.RealmID)

	elasticRepo.AssertExpectations(t)
}

func TestHandleGetElasticJWK_NotFound(t *testing.T) {
	t.Parallel()

	handler, elasticRepo, _, _, _ := setupTestHandler()
	app := setupFiberApp()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	kid := googleUuid.New()

	// Mock repository not found.
	elasticRepo.On("Get", mock.Anything, tenantID, realmID, kid.String()).
		Return(nil, errors.New("not found"))

	app.Get("/jwk/:kid", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)
		c.Locals("realm_id", realmID)
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
	realmID := googleUuid.New()

	// Mock repository.
	expectedJWKs := []*joseJADomain.ElasticJWK{
		{
			ID:                   googleUuid.New(),
			TenantID:             tenantID,
			RealmID:              realmID,
			KID:                  "kid1",
			KeyType:              "RSA",
			Algorithm:            "RSA/2048",
			Use:                  "sig",
			MaxMaterials:         10,
			CurrentMaterialCount: 1,
			CreatedAt:            time.Now(),
		},
		{
			ID:                   googleUuid.New(),
			TenantID:             tenantID,
			RealmID:              realmID,
			KID:                  "kid2",
			KeyType:              "EC",
			Algorithm:            "EC/P256",
			Use:                  "enc",
			MaxMaterials:         5,
			CurrentMaterialCount: 0,
			CreatedAt:            time.Now(),
		},
	}
	elasticRepo.On("List", mock.Anything, tenantID, realmID, 0, 100).
		Return(expectedJWKs, int64(2), nil)

	app.Get("/jwks", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)
		c.Locals("realm_id", realmID)
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
	realmID := googleUuid.New()
	kid := googleUuid.New()

	// Mock repository.
	existingJWK := &joseJADomain.ElasticJWK{
		ID:       kid,
		TenantID: tenantID,
		RealmID:  realmID,
		KID:      kid.String(),
	}
	elasticRepo.On("Get", mock.Anything, tenantID, realmID, kid.String()).
		Return(existingJWK, nil)
	elasticRepo.On("Delete", mock.Anything, kid).
		Return(nil)

	app.Delete("/jwk/:kid", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)
		c.Locals("realm_id", realmID)
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
	realmID := googleUuid.New()
	kid := "test-elastic-kid"

	elasticJWK := &joseJADomain.ElasticJWK{
		ID:                   googleUuid.New(),
		TenantID:             tenantID,
		RealmID:              realmID,
		KID:                  kid,
		MaxMaterials:         5,
		CurrentMaterialCount: 2,
	}

	elasticRepo.On("Get", mock.Anything, tenantID, realmID, kid).Return(elasticJWK, nil)
	materialRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.MaterialJWK")).Return(nil)
	elasticRepo.On("IncrementMaterialCount", mock.Anything, elasticJWK.ID).Return(nil)

	app.Post("/elastic-jwks/:kid/materials", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)
		c.Locals("realm_id", realmID)
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
	realmID := googleUuid.New()
	kid := "test-elastic-kid"

	elasticJWK := &joseJADomain.ElasticJWK{
		ID:                   googleUuid.New(),
		TenantID:             tenantID,
		RealmID:              realmID,
		KID:                  kid,
		MaxMaterials:         5,
		CurrentMaterialCount: 5,
	}

	elasticRepo.On("Get", mock.Anything, tenantID, realmID, kid).Return(elasticJWK, nil)

	app.Post("/elastic-jwks/:kid/materials", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)
		c.Locals("realm_id", realmID)
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
	realmID := googleUuid.New()
	kid := "test-elastic-kid"
	elasticID := googleUuid.New()

	elasticJWK := &joseJADomain.ElasticJWK{
		ID:       elasticID,
		TenantID: tenantID,
		RealmID:  realmID,
		KID:      kid,
	}

	materials := []*joseJADomain.MaterialJWK{
		{
			ID:           googleUuid.New(),
			ElasticJWKID: elasticID,
			MaterialKID:  "material-1",
			Active:       true,
			CreatedAt:    time.Now(),
		},
	}

	elasticRepo.On("Get", mock.Anything, tenantID, realmID, kid).Return(elasticJWK, nil)
	materialRepo.On("ListByElasticJWK", mock.Anything, elasticID, 0, defaultLimit).Return(materials, int64(1), nil)

	app.Get("/elastic-jwks/:kid/materials", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)
		c.Locals("realm_id", realmID)
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

func TestHandleGetActiveMaterialJWK_Success(t *testing.T) {
	t.Parallel()

	handler, elasticRepo, materialRepo, _, _ := setupTestHandler()
	app := setupFiberApp()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	kid := "test-elastic-kid"
	elasticID := googleUuid.New()

	elasticJWK := &joseJADomain.ElasticJWK{
		ID:       elasticID,
		TenantID: tenantID,
		RealmID:  realmID,
		KID:      kid,
	}

	activeMaterial := &joseJADomain.MaterialJWK{
		ID:           googleUuid.New(),
		ElasticJWKID: elasticID,
		MaterialKID:  "active-material",
		Active:       true,
		CreatedAt:    time.Now(),
	}

	elasticRepo.On("Get", mock.Anything, tenantID, realmID, kid).Return(elasticJWK, nil)
	materialRepo.On("GetActiveMaterial", mock.Anything, elasticID).Return(activeMaterial, nil)

	app.Get("/elastic-jwks/:kid/materials/active", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)
		c.Locals("realm_id", realmID)
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
	realmID := googleUuid.New()
	kid := "test-elastic-kid"
	elasticID := googleUuid.New()

	elasticJWK := &joseJADomain.ElasticJWK{
		ID:                   elasticID,
		TenantID:             tenantID,
		RealmID:              realmID,
		KID:                  kid,
		MaxMaterials:         5,
		CurrentMaterialCount: 3,
	}

	elasticRepo.On("Get", mock.Anything, tenantID, realmID, kid).Return(elasticJWK, nil)
	materialRepo.On("RotateMaterial", mock.Anything, elasticID, mock.AnythingOfType("*domain.MaterialJWK")).Return(nil)
	elasticRepo.On("IncrementMaterialCount", mock.Anything, elasticID).Return(nil)

	app.Post("/elastic-jwks/:kid/materials/rotate", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)
		c.Locals("realm_id", realmID)
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
	realmID := googleUuid.New()
	kid := "test-elastic-kid"

	elasticJWK := &joseJADomain.ElasticJWK{
		ID:                   googleUuid.New(),
		TenantID:             tenantID,
		RealmID:              realmID,
		KID:                  kid,
		MaxMaterials:         5,
		CurrentMaterialCount: 5,
	}

	elasticRepo.On("Get", mock.Anything, tenantID, realmID, kid).Return(elasticJWK, nil)

	app.Post("/elastic-jwks/:kid/materials/rotate", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)
		c.Locals("realm_id", realmID)
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
