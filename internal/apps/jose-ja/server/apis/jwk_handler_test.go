// Copyright (c) 2025 Justin Cranford
//

package apis

import (
	"context"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilAppsFrameworkServiceServerBarrier "cryptoutil/internal/apps/framework/service/server/barrier"
	cryptoutilAppsJoseJaModel "cryptoutil/internal/apps/jose-ja/server/model"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockElasticJWKRepository is a mock implementation of ElasticJWKRepository.
type MockElasticJWKRepository struct {
	mock.Mock
}

func (m *MockElasticJWKRepository) Create(ctx context.Context, jwk *cryptoutilAppsJoseJaModel.ElasticJWK) error {
	args := m.Called(ctx, jwk)

	return args.Error(0) //nolint:wrapcheck // Mock returns test-controlled error
}

func (m *MockElasticJWKRepository) Get(ctx context.Context, tenantID googleUuid.UUID, kid string) (*cryptoutilAppsJoseJaModel.ElasticJWK, error) {
	args := m.Called(ctx, tenantID, kid)
	if args.Get(0) == nil {
		return nil, args.Error(1) //nolint:wrapcheck // Mock returns test-controlled error
	}

	result, ok := args.Get(0).(*cryptoutilAppsJoseJaModel.ElasticJWK)
	if !ok {
		panic("MockElasticJWKRepository.Get: expected *cryptoutilAppsJoseJaModel.ElasticJWK")
	}

	return result, args.Error(1) //nolint:wrapcheck // Mock returns test-controlled error
}

func (m *MockElasticJWKRepository) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilAppsJoseJaModel.ElasticJWK, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1) //nolint:wrapcheck // Mock returns test-controlled error
	}

	result, ok := args.Get(0).(*cryptoutilAppsJoseJaModel.ElasticJWK)
	if !ok {
		panic("MockElasticJWKRepository.GetByID: expected *cryptoutilAppsJoseJaModel.ElasticJWK")
	}

	return result, args.Error(1) //nolint:wrapcheck // Mock returns test-controlled error
}

func (m *MockElasticJWKRepository) List(ctx context.Context, tenantID googleUuid.UUID, offset, limit int) ([]*cryptoutilAppsJoseJaModel.ElasticJWK, int64, error) {
	args := m.Called(ctx, tenantID, offset, limit)
	if args.Get(0) == nil {
		count, ok := args.Get(1).(int64)
		if !ok {
			panic("MockElasticJWKRepository.List: expected int64 for count")
		}

		return nil, count, args.Error(2) //nolint:wrapcheck // Mock returns test-controlled error
	}

	results, ok := args.Get(0).([]*cryptoutilAppsJoseJaModel.ElasticJWK)
	if !ok {
		panic("MockElasticJWKRepository.List: expected []*cryptoutilAppsJoseJaModel.ElasticJWK")
	}

	count, ok := args.Get(1).(int64)
	if !ok {
		panic("MockElasticJWKRepository.List: expected int64 for count")
	}

	return results, count, args.Error(2) //nolint:wrapcheck // Mock returns test-controlled error
}

func (m *MockElasticJWKRepository) Update(ctx context.Context, jwk *cryptoutilAppsJoseJaModel.ElasticJWK) error {
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

func (m *MockMaterialJWKRepository) Create(ctx context.Context, jwk *cryptoutilAppsJoseJaModel.MaterialJWK) error {
	args := m.Called(ctx, jwk)

	return args.Error(0) //nolint:wrapcheck // Mock returns test-controlled error
}

func (m *MockMaterialJWKRepository) GetByMaterialKID(ctx context.Context, materialKID string) (*cryptoutilAppsJoseJaModel.MaterialJWK, error) {
	args := m.Called(ctx, materialKID)
	if args.Get(0) == nil {
		return nil, args.Error(1) //nolint:wrapcheck // Mock returns test-controlled error
	}

	result, ok := args.Get(0).(*cryptoutilAppsJoseJaModel.MaterialJWK)
	if !ok {
		panic("MockMaterialJWKRepository.GetByMaterialKID: expected *cryptoutilAppsJoseJaModel.MaterialJWK")
	}

	return result, args.Error(1) //nolint:wrapcheck // Mock returns test-controlled error
}

func (m *MockMaterialJWKRepository) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilAppsJoseJaModel.MaterialJWK, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1) //nolint:wrapcheck // Mock returns test-controlled error
	}

	result, ok := args.Get(0).(*cryptoutilAppsJoseJaModel.MaterialJWK)
	if !ok {
		panic("MockMaterialJWKRepository.GetByID: expected *cryptoutilAppsJoseJaModel.MaterialJWK")
	}

	return result, args.Error(1) //nolint:wrapcheck // Mock returns test-controlled error
}

func (m *MockMaterialJWKRepository) GetActiveMaterial(ctx context.Context, elasticJWKID googleUuid.UUID) (*cryptoutilAppsJoseJaModel.MaterialJWK, error) {
	args := m.Called(ctx, elasticJWKID)
	if args.Get(0) == nil {
		return nil, args.Error(1) //nolint:wrapcheck // Mock returns test-controlled error
	}

	result, ok := args.Get(0).(*cryptoutilAppsJoseJaModel.MaterialJWK)
	if !ok {
		panic("MockMaterialJWKRepository.GetActiveMaterial: expected *cryptoutilAppsJoseJaModel.MaterialJWK")
	}

	return result, args.Error(1) //nolint:wrapcheck // Mock returns test-controlled error
}

func (m *MockMaterialJWKRepository) ListByElasticJWK(ctx context.Context, elasticJWKID googleUuid.UUID, offset, limit int) ([]*cryptoutilAppsJoseJaModel.MaterialJWK, int64, error) {
	args := m.Called(ctx, elasticJWKID, offset, limit)
	if args.Get(0) == nil {
		count, ok := args.Get(1).(int64)
		if !ok {
			panic("MockMaterialJWKRepository.ListByElasticJWK: expected int64 for count")
		}

		return nil, count, args.Error(2) //nolint:wrapcheck // Mock returns test-controlled error
	}

	results, ok := args.Get(0).([]*cryptoutilAppsJoseJaModel.MaterialJWK)
	if !ok {
		panic("MockMaterialJWKRepository.ListByElasticJWK: expected []*cryptoutilAppsJoseJaModel.MaterialJWK")
	}

	count, ok := args.Get(1).(int64)
	if !ok {
		panic("MockMaterialJWKRepository.ListByElasticJWK: expected int64 for count")
	}

	return results, count, args.Error(2) //nolint:wrapcheck // Mock returns test-controlled error
}

func (m *MockMaterialJWKRepository) RotateMaterial(ctx context.Context, elasticJWKID googleUuid.UUID, newMaterial *cryptoutilAppsJoseJaModel.MaterialJWK) error {
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

	count, ok := args.Get(0).(int64)
	if !ok {
		panic("MockMaterialJWKRepository.CountMaterials: expected int64")
	}

	return count, args.Error(1) //nolint:wrapcheck // Mock returns test-controlled error
}

// MockAuditConfigRepository is a mock implementation of AuditConfigRepository.
type MockAuditConfigRepository struct {
	mock.Mock
}

func (m *MockAuditConfigRepository) Get(ctx context.Context, tenantID googleUuid.UUID, operation string) (*cryptoutilAppsJoseJaModel.AuditConfig, error) {
	args := m.Called(ctx, tenantID, operation)
	if args.Get(0) == nil {
		return nil, args.Error(1) //nolint:wrapcheck // Mock returns test-controlled error
	}

	result, ok := args.Get(0).(*cryptoutilAppsJoseJaModel.AuditConfig)
	if !ok {
		panic("MockAuditConfigRepository.Get: expected *cryptoutilAppsJoseJaModel.AuditConfig")
	}

	return result, args.Error(1) //nolint:wrapcheck // Mock returns test-controlled error
}

func (m *MockAuditConfigRepository) GetAllForTenant(ctx context.Context, tenantID googleUuid.UUID) ([]*cryptoutilAppsJoseJaModel.AuditConfig, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1) //nolint:wrapcheck // Mock returns test-controlled error
	}

	results, ok := args.Get(0).([]*cryptoutilAppsJoseJaModel.AuditConfig)
	if !ok {
		panic("MockAuditConfigRepository.GetAllForTenant: expected []*cryptoutilAppsJoseJaModel.AuditConfig")
	}

	return results, args.Error(1) //nolint:wrapcheck // Mock returns test-controlled error
}

func (m *MockAuditConfigRepository) Upsert(ctx context.Context, config *cryptoutilAppsJoseJaModel.AuditConfig) error {
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

func (m *MockAuditLogRepository) Create(ctx context.Context, entry *cryptoutilAppsJoseJaModel.AuditLogEntry) error {
	args := m.Called(ctx, entry)

	return args.Error(0) //nolint:wrapcheck // Mock returns test-controlled error
}

func (m *MockAuditLogRepository) List(ctx context.Context, tenantID googleUuid.UUID, offset, limit int) ([]*cryptoutilAppsJoseJaModel.AuditLogEntry, int64, error) {
	args := m.Called(ctx, tenantID, offset, limit)
	if args.Get(0) == nil {
		count, ok := args.Get(1).(int64)
		if !ok {
			panic("MockAuditLogRepository.List: expected int64 for count")
		}

		return nil, count, args.Error(2) //nolint:wrapcheck // Mock returns test-controlled error
	}

	results, ok := args.Get(0).([]*cryptoutilAppsJoseJaModel.AuditLogEntry)
	if !ok {
		panic("MockAuditLogRepository.List: expected []*cryptoutilAppsJoseJaModel.AuditLogEntry")
	}

	count, ok := args.Get(1).(int64)
	if !ok {
		panic("MockAuditLogRepository.List: expected int64 for count")
	}

	return results, count, args.Error(2) //nolint:wrapcheck // Mock returns test-controlled error
}

func (m *MockAuditLogRepository) ListByElasticJWK(ctx context.Context, elasticJWKID googleUuid.UUID, offset, limit int) ([]*cryptoutilAppsJoseJaModel.AuditLogEntry, int64, error) {
	args := m.Called(ctx, elasticJWKID, offset, limit)
	if args.Get(0) == nil {
		count, ok := args.Get(1).(int64)
		if !ok {
			panic("MockAuditLogRepository.ListByElasticJWK: expected int64 for count")
		}

		return nil, count, args.Error(2) //nolint:wrapcheck // Mock returns test-controlled error
	}

	results, ok := args.Get(0).([]*cryptoutilAppsJoseJaModel.AuditLogEntry)
	if !ok {
		panic("MockAuditLogRepository.ListByElasticJWK: expected []*cryptoutilAppsJoseJaModel.AuditLogEntry")
	}

	count, ok := args.Get(1).(int64)
	if !ok {
		panic("MockAuditLogRepository.ListByElasticJWK: expected int64 for count")
	}

	return results, count, args.Error(2) //nolint:wrapcheck // Mock returns test-controlled error
}

func (m *MockAuditLogRepository) ListByOperation(ctx context.Context, tenantID googleUuid.UUID, operation string, offset, limit int) ([]*cryptoutilAppsJoseJaModel.AuditLogEntry, int64, error) {
	args := m.Called(ctx, tenantID, operation, offset, limit)
	if args.Get(0) == nil {
		count, ok := args.Get(1).(int64)
		if !ok {
			panic("MockAuditLogRepository.ListByOperation: expected int64 for count")
		}

		return nil, count, args.Error(2) //nolint:wrapcheck // Mock returns test-controlled error
	}

	results, ok := args.Get(0).([]*cryptoutilAppsJoseJaModel.AuditLogEntry)
	if !ok {
		panic("MockAuditLogRepository.ListByOperation: expected []*cryptoutilAppsJoseJaModel.AuditLogEntry")
	}

	count, ok := args.Get(1).(int64)
	if !ok {
		panic("MockAuditLogRepository.ListByOperation: expected int64 for count")
	}

	return results, count, args.Error(2) //nolint:wrapcheck // Mock returns test-controlled error
}

func (m *MockAuditLogRepository) GetByRequestID(ctx context.Context, requestID string) (*cryptoutilAppsJoseJaModel.AuditLogEntry, error) {
	args := m.Called(ctx, requestID)
	if args.Get(0) == nil {
		return nil, args.Error(1) //nolint:wrapcheck // Mock returns test-controlled error
	}

	result, ok := args.Get(0).(*cryptoutilAppsJoseJaModel.AuditLogEntry)
	if !ok {
		panic("MockAuditLogRepository.GetByRequestID: expected *cryptoutilAppsJoseJaModel.AuditLogEntry")
	}

	return result, args.Error(1) //nolint:wrapcheck // Mock returns test-controlled error
}

func (m *MockAuditLogRepository) DeleteOlderThan(ctx context.Context, tenantID googleUuid.UUID, days int) (int64, error) {
	args := m.Called(ctx, tenantID, days)

	count, ok := args.Get(0).(int64)
	if !ok {
		panic("MockAuditLogRepository.DeleteOlderThan: expected int64")
	}

	return count, args.Error(1) //nolint:wrapcheck // Mock returns test-controlled error
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
	barrierService := &cryptoutilAppsFrameworkServiceServerBarrier.Service{}

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
				cryptoutilSharedMagic.StringError: err.Error(),
			})
		},
	})

	return app
}
