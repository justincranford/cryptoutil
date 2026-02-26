// Copyright (c) 2025 Justin Cranford
//

package apis

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"

	cryptoutilAppsJoseJaDomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/mock"
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
				cryptoutilSharedMagic.StringError: err.Error(),
			})
		},
	})

	return app
}
