// Copyright (c) 2025 Justin Cranford
//

package apis

import (
	"bytes"
	"context"
	json "encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	cryptoutilAppsJoseJaDomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockElasticJWKRepository is a mock implementation of ElasticJWKRepository.
type MockElasticJWKRepository struct {
	mock.Mock
}

func (m *MockElasticJWKRepository) Create(ctx context.Context, jwk *cryptoutilAppsJoseJaDomain.ElasticJWK) error {
	args := m.Called(ctx, jwk)

	return args.Error(0) //nolint:wrapcheck // Mock returns test-controlled error
}

func (m *MockElasticJWKRepository) Get(ctx context.Context, tenantID googleUuid.UUID, kid string) (*cryptoutilAppsJoseJaDomain.ElasticJWK, error) {
	args := m.Called(ctx, tenantID, kid)
	if args.Get(0) == nil {
		return nil, args.Error(1) //nolint:wrapcheck // Mock returns test-controlled error
	}

	return args.Get(0).(*cryptoutilAppsJoseJaDomain.ElasticJWK), args.Error(1) //nolint:errcheck,wrapcheck // Mock type assertion and error controlled by test
}

func (m *MockElasticJWKRepository) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilAppsJoseJaDomain.ElasticJWK, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1) //nolint:wrapcheck // Mock returns test-controlled error
	}

	return args.Get(0).(*cryptoutilAppsJoseJaDomain.ElasticJWK), args.Error(1) //nolint:errcheck,wrapcheck // Mock type assertion and error controlled by test
}

func (m *MockElasticJWKRepository) List(ctx context.Context, tenantID googleUuid.UUID, offset, limit int) ([]*cryptoutilAppsJoseJaDomain.ElasticJWK, int64, error) {
	args := m.Called(ctx, tenantID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2) //nolint:errcheck,wrapcheck // Mock type assertion and error controlled by test
	}

	return args.Get(0).([]*cryptoutilAppsJoseJaDomain.ElasticJWK), args.Get(1).(int64), args.Error(2) //nolint:errcheck,wrapcheck // Mock type assertion and error controlled by test
}

func (m *MockElasticJWKRepository) Update(ctx context.Context, jwk *cryptoutilAppsJoseJaDomain.ElasticJWK) error {
	args := m.Called(ctx, jwk)

	return args.Error(0) //nolint:wrapcheck // Mock returns test-controlled error
}

func (m *MockElasticJWKRepository) Delete(ctx context.Context, id googleUuid.UUID) error {
	args := m.Called(ctx, id)

	return args.Error(0) //nolint:wrapcheck // Mock returns test-controlled error
}

func (m *MockElasticJWKRepository) IncrementMaterialCount(ctx context.Context, id googleUuid.UUID) error {
	args := m.Called(ctx, id)

	return args.Error(0) //nolint:wrapcheck // Mock returns test-controlled error
}

func (m *MockElasticJWKRepository) DecrementMaterialCount(ctx context.Context, id googleUuid.UUID) error {
	args := m.Called(ctx, id)

	return args.Error(0) //nolint:wrapcheck // Mock returns test-controlled error
}

// MockMaterialJWKRepository is a mock implementation of MaterialJWKRepository.
type MockMaterialJWKRepository struct {
	mock.Mock
}

func (m *MockMaterialJWKRepository) Create(ctx context.Context, jwk *cryptoutilAppsJoseJaDomain.MaterialJWK) error {
	args := m.Called(ctx, jwk)

	return args.Error(0) //nolint:wrapcheck // Mock returns test-controlled error
}

func (m *MockMaterialJWKRepository) GetByMaterialKID(ctx context.Context, materialKID string) (*cryptoutilAppsJoseJaDomain.MaterialJWK, error) {
	args := m.Called(ctx, materialKID)
	if args.Get(0) == nil {
		return nil, args.Error(1) //nolint:wrapcheck // Mock returns test-controlled error
	}

	return args.Get(0).(*cryptoutilAppsJoseJaDomain.MaterialJWK), args.Error(1) //nolint:errcheck,wrapcheck // Mock type assertion and error controlled by test
}

func (m *MockMaterialJWKRepository) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilAppsJoseJaDomain.MaterialJWK, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1) //nolint:wrapcheck // Mock returns test-controlled error
	}

	return args.Get(0).(*cryptoutilAppsJoseJaDomain.MaterialJWK), args.Error(1) //nolint:errcheck,wrapcheck // Mock type assertion and error controlled by test
}

func (m *MockMaterialJWKRepository) GetActiveMaterial(ctx context.Context, elasticJWKID googleUuid.UUID) (*cryptoutilAppsJoseJaDomain.MaterialJWK, error) {
	args := m.Called(ctx, elasticJWKID)
	if args.Get(0) == nil {
		return nil, args.Error(1) //nolint:wrapcheck // Mock returns test-controlled error
	}

	return args.Get(0).(*cryptoutilAppsJoseJaDomain.MaterialJWK), args.Error(1) //nolint:errcheck,wrapcheck // Mock type assertion and error controlled by test
}

func (m *MockMaterialJWKRepository) ListByElasticJWK(ctx context.Context, elasticJWKID googleUuid.UUID, offset, limit int) ([]*cryptoutilAppsJoseJaDomain.MaterialJWK, int64, error) {
	args := m.Called(ctx, elasticJWKID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2) //nolint:errcheck,wrapcheck // Mock type assertion and error controlled by test
	}

	return args.Get(0).([]*cryptoutilAppsJoseJaDomain.MaterialJWK), args.Get(1).(int64), args.Error(2) //nolint:errcheck,wrapcheck // Mock type assertion and error controlled by test
}

func (m *MockMaterialJWKRepository) RotateMaterial(ctx context.Context, elasticJWKID googleUuid.UUID, newMaterial *cryptoutilAppsJoseJaDomain.MaterialJWK) error {
	args := m.Called(ctx, elasticJWKID, newMaterial)

	return args.Error(0) //nolint:wrapcheck // Mock returns test-controlled error
}

func (m *MockMaterialJWKRepository) RetireMaterial(ctx context.Context, id googleUuid.UUID) error {
	args := m.Called(ctx, id)

	return args.Error(0) //nolint:wrapcheck // Mock returns test-controlled error
}

func (m *MockMaterialJWKRepository) Delete(ctx context.Context, id googleUuid.UUID) error {
	args := m.Called(ctx, id)

	return args.Error(0) //nolint:wrapcheck // Mock returns test-controlled error
}

func (m *MockMaterialJWKRepository) CountMaterials(ctx context.Context, elasticJWKID googleUuid.UUID) (int64, error) {
	args := m.Called(ctx, elasticJWKID)

	return args.Get(0).(int64), args.Error(1) //nolint:errcheck,wrapcheck // Mock type assertion and error controlled by test
}

// MockAuditConfigRepository is a mock implementation of AuditConfigRepository.
type MockAuditConfigRepository struct {
	mock.Mock
}

func (m *MockAuditConfigRepository) Get(ctx context.Context, tenantID googleUuid.UUID, operation string) (*cryptoutilAppsJoseJaDomain.AuditConfig, error) {
	args := m.Called(ctx, tenantID, operation)
	if args.Get(0) == nil {
		return nil, args.Error(1) //nolint:wrapcheck // Mock returns test-controlled error
	}

	return args.Get(0).(*cryptoutilAppsJoseJaDomain.AuditConfig), args.Error(1) //nolint:errcheck,wrapcheck // Mock type assertion and error controlled by test
}

func (m *MockAuditConfigRepository) GetAllForTenant(ctx context.Context, tenantID googleUuid.UUID) ([]*cryptoutilAppsJoseJaDomain.AuditConfig, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1) //nolint:wrapcheck // Mock returns test-controlled error
	}

	return args.Get(0).([]*cryptoutilAppsJoseJaDomain.AuditConfig), args.Error(1) //nolint:errcheck,wrapcheck // Mock type assertion and error controlled by test
}

func (m *MockAuditConfigRepository) Upsert(ctx context.Context, config *cryptoutilAppsJoseJaDomain.AuditConfig) error {
	args := m.Called(ctx, config)

	return args.Error(0) //nolint:wrapcheck // Mock returns test-controlled error
}

func (m *MockAuditConfigRepository) Delete(ctx context.Context, tenantID googleUuid.UUID, operation string) error {
	args := m.Called(ctx, tenantID, operation)

	return args.Error(0) //nolint:wrapcheck // Mock returns test-controlled error
}

func (m *MockAuditConfigRepository) ShouldAudit(ctx context.Context, tenantID googleUuid.UUID, operation string) (bool, error) {
	args := m.Called(ctx, tenantID, operation)

	return args.Bool(0), args.Error(1) //nolint:wrapcheck // Mock returns test-controlled error
}

// MockAuditLogRepository is a mock implementation of AuditLogRepository.
type MockAuditLogRepository struct {
	mock.Mock
}

func (m *MockAuditLogRepository) Create(ctx context.Context, entry *cryptoutilAppsJoseJaDomain.AuditLogEntry) error {
	args := m.Called(ctx, entry)

	return args.Error(0) //nolint:wrapcheck // Mock returns test-controlled error
}

func (m *MockAuditLogRepository) List(ctx context.Context, tenantID googleUuid.UUID, offset, limit int) ([]*cryptoutilAppsJoseJaDomain.AuditLogEntry, int64, error) {
	args := m.Called(ctx, tenantID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2) //nolint:errcheck,wrapcheck // Mock type assertion and error controlled by test
	}

	return args.Get(0).([]*cryptoutilAppsJoseJaDomain.AuditLogEntry), args.Get(1).(int64), args.Error(2) //nolint:errcheck,wrapcheck // Mock type assertion and error controlled by test
}

func (m *MockAuditLogRepository) ListByElasticJWK(ctx context.Context, elasticJWKID googleUuid.UUID, offset, limit int) ([]*cryptoutilAppsJoseJaDomain.AuditLogEntry, int64, error) {
	args := m.Called(ctx, elasticJWKID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2) //nolint:errcheck,wrapcheck // Mock type assertion and error controlled by test
	}

	return args.Get(0).([]*cryptoutilAppsJoseJaDomain.AuditLogEntry), args.Get(1).(int64), args.Error(2) //nolint:errcheck,wrapcheck // Mock type assertion and error controlled by test
}

func (m *MockAuditLogRepository) ListByOperation(ctx context.Context, tenantID googleUuid.UUID, operation string, offset, limit int) ([]*cryptoutilAppsJoseJaDomain.AuditLogEntry, int64, error) {
	args := m.Called(ctx, tenantID, operation, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2) //nolint:errcheck,wrapcheck // Mock type assertion and error controlled by test
	}

	return args.Get(0).([]*cryptoutilAppsJoseJaDomain.AuditLogEntry), args.Get(1).(int64), args.Error(2) //nolint:errcheck,wrapcheck // Mock type assertion and error controlled by test
}

func (m *MockAuditLogRepository) GetByRequestID(ctx context.Context, requestID string) (*cryptoutilAppsJoseJaDomain.AuditLogEntry, error) {
	args := m.Called(ctx, requestID)
	if args.Get(0) == nil {
		return nil, args.Error(1) //nolint:wrapcheck // Mock returns test-controlled error
	}

	return args.Get(0).(*cryptoutilAppsJoseJaDomain.AuditLogEntry), args.Error(1) //nolint:errcheck,wrapcheck // Mock type assertion and error controlled by test
}

func (m *MockAuditLogRepository) DeleteOlderThan(ctx context.Context, tenantID googleUuid.UUID, days int) (int64, error) {
	args := m.Called(ctx, tenantID, days)

	return args.Get(0).(int64), args.Error(1) //nolint:errcheck,wrapcheck // Mock type assertion and error controlled by test
}

// Test constants to satisfy goconst linter.
const (
	testElasticKID     = "test-elastic-kid"
	testIncrementError = "test-increment-error"
	testCreateJWKBody  = `{"kid":"test","algorithm":"RSA/2048","use":"sig","max_materials":5}`
)

// setupTestHandler creates a test handler with mocks.
func setupTestHandler() (*JWKHandler, *MockElasticJWKRepository, *MockMaterialJWKRepository, *MockAuditConfigRepository, *MockAuditLogRepository) {
	elasticRepo := new(MockElasticJWKRepository)
	materialRepo := new(MockMaterialJWKRepository)
	auditConfigRepo := new(MockAuditConfigRepository)
	auditLogRepo := new(MockAuditLogRepository)

	// Create minimal JWKGenService for testing.
	jwkGenService := &cryptoutilSharedCryptoJose.JWKGenService{}

	// Create minimal BarrierService for testing.
	barrierService := &cryptoutilAppsTemplateServiceServerBarrier.Service{}

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
	materialRepo.On("ListByElasticJWK", mock.Anything, elasticID, 0, defaultLimit).Return(materials, int64(1), nil)

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
	materialRepo.On("ListByElasticJWK", mock.Anything, elasticJWK.ID, 0, defaultLimit).Return(nil, int64(0), errors.New("list failed"))

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
	resp, err := app.Test(req)
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
	resp, err := app.Test(req)
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
	resp, err := app.Test(req)
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
	resp, err := app.Test(req)
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
	resp, err := app.Test(req)
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
	resp, err := app.Test(req)
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
	resp, err := app.Test(req)
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
	resp, err := app.Test(req)
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
	resp, err := app.Test(req)
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
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	require.NoError(t, resp.Body.Close())

	elasticRepo.AssertExpectations(t)
}

// ==================== Additional Coverage Tests ====================

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
	materialRepo.On("ListByElasticJWK", mock.Anything, elasticID, 0, defaultLimit).Return(materials, int64(1), nil)

	app.Get("/elastic-jwks/:kid/materials", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}, handler.HandleListMaterialJWKs())

	req := httptest.NewRequest(fiber.MethodGet, fmt.Sprintf("/elastic-jwks/%s/materials", kid), nil)
	resp, err := app.Test(req)
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
	resp, err := app.Test(req)
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
	resp, err := app.Test(req)
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
	resp, err := app.Test(req)
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

	resp, err := app.Test(req)
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
	resp, err := app.Test(req)
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
	resp, err := app.Test(req)
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
	resp, err := app.Test(req)
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
	resp, err := app.Test(req)
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
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}
