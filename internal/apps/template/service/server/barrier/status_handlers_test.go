// Copyright (c) 2025 Justin Cranford
//
//

package barrier

import (
	"bytes"
	"context"
	json "encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// errMockDBFailure is a standard error for mock database failures.
var errMockDBFailure = errors.New("mock database failure")

// mockStatusTransaction implements Transaction interface for testing error scenarios.
type mockStatusTransaction struct {
	ctx                         context.Context
	rootKey                     *RootKey
	intermediateKey             *IntermediateKey
	contentKey                  *ContentKey
	getRootKeyLatestErr         error
	getRootKeyErr               error
	addRootKeyErr               error
	getIntermediateKeyLatestErr error
	getIntermediateKeyErr       error
	addIntermediateKeyErr       error
	getContentKeyErr            error
	addContentKeyErr            error
}

func (m *mockStatusTransaction) Context() context.Context {
	return m.ctx
}

func (m *mockStatusTransaction) GetRootKeyLatest() (*RootKey, error) {
	if m.getRootKeyLatestErr != nil {
		return nil, m.getRootKeyLatestErr
	}

	return m.rootKey, nil
}

func (m *mockStatusTransaction) GetRootKey(_ *googleUuid.UUID) (*RootKey, error) {
	if m.getRootKeyErr != nil {
		return nil, m.getRootKeyErr
	}

	return m.rootKey, nil
}

func (m *mockStatusTransaction) AddRootKey(_ *RootKey) error {
	return m.addRootKeyErr
}

func (m *mockStatusTransaction) GetIntermediateKeyLatest() (*IntermediateKey, error) {
	if m.getIntermediateKeyLatestErr != nil {
		return nil, m.getIntermediateKeyLatestErr
	}

	return m.intermediateKey, nil
}

func (m *mockStatusTransaction) GetIntermediateKey(_ *googleUuid.UUID) (*IntermediateKey, error) {
	if m.getIntermediateKeyErr != nil {
		return nil, m.getIntermediateKeyErr
	}

	return m.intermediateKey, nil
}

func (m *mockStatusTransaction) AddIntermediateKey(_ *IntermediateKey) error {
	return m.addIntermediateKeyErr
}

func (m *mockStatusTransaction) GetContentKey(_ *googleUuid.UUID) (*ContentKey, error) {
	if m.getContentKeyErr != nil {
		return nil, m.getContentKeyErr
	}

	return m.contentKey, nil
}

func (m *mockStatusTransaction) AddContentKey(_ *ContentKey) error {
	return m.addContentKeyErr
}

// mockStatusRepository implements Repository interface for testing error scenarios.
type mockStatusRepository struct {
	tx             *mockStatusTransaction
	withTxErr      error
	shouldCallTxFn bool
	shutdownCalled bool
}

func (m *mockStatusRepository) WithTransaction(ctx context.Context, fn func(tx Transaction) error) error {
	if m.withTxErr != nil {
		return m.withTxErr
	}

	if m.shouldCallTxFn && m.tx != nil {
		m.tx.ctx = ctx

		return fn(m.tx)
	}

	return nil
}

func (m *mockStatusRepository) Shutdown() {
	m.shutdownCalled = true
}

// newMockStatusRepository creates a mockStatusRepository with a mockStatusTransaction for testing.
func newMockStatusRepository() *mockStatusRepository {
	return &mockStatusRepository{
		tx:             &mockStatusTransaction{},
		shouldCallTxFn: true,
	}
}

// TestHandleGetBarrierKeysStatus_Success tests successful retrieval of barrier keys status.
func TestHandleGetBarrierKeysStatus_Success(t *testing.T) {
	t.Parallel()

	// Setup test environment (creates root + intermediate keys automatically).
	app, rotationService, _ := setupRotationTestEnvironment(t)

	// Create status service from rotation service's repository.
	statusService, err := NewStatusService(rotationService.repository)
	require.NoError(t, err)

	// Register status routes.
	RegisterStatusRoutes(app, statusService)

	// Make HTTP request with increased timeout (SQLite GORM can have slow queries).
	req := httptest.NewRequest("GET", "/admin/api/v1/barrier/keys/status", nil)
	resp, err := app.Test(req, 5000) // 5-second timeout for SQLite contention
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Verify response body.
	var statusResp KeysStatusResponse

	err = json.NewDecoder(resp.Body).Decode(&statusResp)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	// Verify root key exists.
	require.NotNil(t, statusResp.RootKey)
	require.NotEmpty(t, statusResp.RootKey.UUID)
	require.Greater(t, statusResp.RootKey.CreatedAt, int64(0))

	// Verify intermediate key exists.
	require.NotNil(t, statusResp.IntermediateKey)
	require.NotEmpty(t, statusResp.IntermediateKey.UUID)
	require.Greater(t, statusResp.IntermediateKey.CreatedAt, int64(0))
}

// TestNewStatusService_NilRepository tests NewStatusService with nil repository.
func TestNewStatusService_NilRepository(t *testing.T) {
	t.Parallel()

	statusService, err := NewStatusService(nil)
	require.Nil(t, statusService)
	require.Error(t, err)
	require.Contains(t, err.Error(), "repository must be non-nil")
}

// TestRegisterStatusRoutes_Integration tests full HTTP integration.
func TestRegisterStatusRoutes_Integration(t *testing.T) {
	t.Parallel()

	// Setup test environment.
	_, rotationService, _ := setupRotationTestEnvironment(t)

	statusService, err := NewStatusService(rotationService.repository)
	require.NoError(t, err)

	// Create fiber app and register routes.
	app := fiber.New()
	RegisterStatusRoutes(app, statusService)

	// Verify route is registered (GET request succeeds with increased timeout for SQLite).
	req := httptest.NewRequest("GET", "/admin/api/v1/barrier/keys/status", nil)
	resp, err := app.Test(req, 5000) // 5-second timeout for SQLite contention
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.NoError(t, resp.Body.Close())

	// Verify wrong HTTP method returns 405 Method Not Allowed.
	req = httptest.NewRequest("POST", "/admin/api/v1/barrier/keys/status", bytes.NewReader([]byte(`{}`)))
	resp, err = app.Test(req, 5000) // 5-second timeout for SQLite contention
	require.NoError(t, err)
	require.Equal(t, fiber.StatusMethodNotAllowed, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

// TestGetBarrierKeysStatus_RootKeyError tests GetBarrierKeysStatus when GetRootKeyLatest fails.
func TestGetBarrierKeysStatus_RootKeyError(t *testing.T) {
	t.Parallel()

	// Create mock repository that returns error on GetRootKeyLatest.
	mockRepo := newMockStatusRepository()
	mockRepo.tx.getRootKeyLatestErr = errMockDBFailure

	statusService, err := NewStatusService(mockRepo)
	require.NoError(t, err)

	ctx := context.Background()
	status, err := statusService.GetBarrierKeysStatus(ctx)

	require.Nil(t, status)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get latest root key")
}

// TestGetBarrierKeysStatus_IntermediateKeyError tests GetBarrierKeysStatus when GetIntermediateKeyLatest fails.
func TestGetBarrierKeysStatus_IntermediateKeyError(t *testing.T) {
	t.Parallel()

	// Create mock repository that returns error on GetIntermediateKeyLatest.
	mockRepo := newMockStatusRepository()
	mockRepo.tx.getIntermediateKeyLatestErr = errMockDBFailure
	// Root key succeeds with valid data.
	rootUUID, _ := googleUuid.NewV7()
	mockRepo.tx.rootKey = &RootKey{
		UUID:      rootUUID,
		Encrypted: "encrypted-root-key",
		CreatedAt: 1234567890,
		UpdatedAt: 1234567890,
	}

	statusService, err := NewStatusService(mockRepo)
	require.NoError(t, err)

	ctx := context.Background()
	status, err := statusService.GetBarrierKeysStatus(ctx)

	require.Nil(t, status)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get latest intermediate key")
}

// TestHandleGetBarrierKeysStatus_ServiceError tests HandleGetBarrierKeysStatus when service returns error.
func TestHandleGetBarrierKeysStatus_ServiceError(t *testing.T) {
	t.Parallel()

	// Create mock repository that returns error.
	mockRepo := newMockStatusRepository()
	mockRepo.tx.getRootKeyLatestErr = errMockDBFailure

	statusService, err := NewStatusService(mockRepo)
	require.NoError(t, err)

	// Create fiber app and register routes.
	app := fiber.New()
	RegisterStatusRoutes(app, statusService)

	// Make HTTP request.
	req := httptest.NewRequest("GET", "/admin/api/v1/barrier/keys/status", nil)
	resp, err := app.Test(req, 5000)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	// Verify error response body.
	var errorResp map[string]string

	err = json.NewDecoder(resp.Body).Decode(&errorResp)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Contains(t, errorResp["error"], "Failed to retrieve barrier keys status")
}

// TestGetBarrierKeysStatus_NoKeys tests GetBarrierKeysStatus when no keys exist (nil responses).
func TestGetBarrierKeysStatus_NoKeys(t *testing.T) {
	t.Parallel()

	// Create mock repository that returns nil keys (no error).
	mockRepo := newMockStatusRepository()
	// Both rootKey and intermediateKey are nil by default.

	statusService, err := NewStatusService(mockRepo)
	require.NoError(t, err)

	ctx := context.Background()
	status, err := statusService.GetBarrierKeysStatus(ctx)

	require.NoError(t, err)
	require.NotNil(t, status)
	require.Nil(t, status.RootKey)
	require.Nil(t, status.IntermediateKey)
}

// TestGetBarrierKeysStatus_TransactionError tests GetBarrierKeysStatus when WithTransaction fails.
func TestGetBarrierKeysStatus_TransactionError(t *testing.T) {
	t.Parallel()

	// Create mock repository that returns error on WithTransaction.
	mockRepo := newMockStatusRepository()
	mockRepo.withTxErr = errMockDBFailure
	mockRepo.shouldCallTxFn = false

	statusService, err := NewStatusService(mockRepo)
	require.NoError(t, err)

	ctx := context.Background()
	status, err := statusService.GetBarrierKeysStatus(ctx)

	require.Nil(t, status)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get barrier keys status")
}
