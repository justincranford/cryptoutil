// Copyright (c) 2025 Justin Cranford
//
//

package barrier

import (
	"bytes"
	"context"
	"database/sql"
	json "encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http/httptest"
	"testing"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilUnsealKeysService "cryptoutil/internal/apps/template/service/server/barrier/unsealkeysservice"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

// errMockRotationDBFailure is a standard error for mock database failures in rotation tests.
var errMockRotationDBFailure = errors.New("mock rotation database failure")

// mockRotationTransaction implements Transaction interface for testing error scenarios.
type mockRotationTransaction struct {
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

func (m *mockRotationTransaction) Context() context.Context {
	return m.ctx
}

func (m *mockRotationTransaction) GetRootKeyLatest() (*RootKey, error) {
	if m.getRootKeyLatestErr != nil {
		return nil, m.getRootKeyLatestErr
	}

	return m.rootKey, nil
}

func (m *mockRotationTransaction) GetRootKey(_ *googleUuid.UUID) (*RootKey, error) {
	if m.getRootKeyErr != nil {
		return nil, m.getRootKeyErr
	}

	return m.rootKey, nil
}

func (m *mockRotationTransaction) AddRootKey(_ *RootKey) error {
	return m.addRootKeyErr
}

func (m *mockRotationTransaction) GetIntermediateKeyLatest() (*IntermediateKey, error) {
	if m.getIntermediateKeyLatestErr != nil {
		return nil, m.getIntermediateKeyLatestErr
	}

	return m.intermediateKey, nil
}

func (m *mockRotationTransaction) GetIntermediateKey(_ *googleUuid.UUID) (*IntermediateKey, error) {
	if m.getIntermediateKeyErr != nil {
		return nil, m.getIntermediateKeyErr
	}

	return m.intermediateKey, nil
}

func (m *mockRotationTransaction) AddIntermediateKey(_ *IntermediateKey) error {
	return m.addIntermediateKeyErr
}

func (m *mockRotationTransaction) GetContentKey(_ *googleUuid.UUID) (*ContentKey, error) {
	if m.getContentKeyErr != nil {
		return nil, m.getContentKeyErr
	}

	return m.contentKey, nil
}

func (m *mockRotationTransaction) AddContentKey(_ *ContentKey) error {
	return m.addContentKeyErr
}

// mockRotationRepository implements Repository interface for testing error scenarios.
type mockRotationRepository struct {
	tx             *mockRotationTransaction
	withTxErr      error
	shouldCallTxFn bool
	shutdownCalled bool
}

func (m *mockRotationRepository) WithTransaction(ctx context.Context, fn func(tx Transaction) error) error {
	if m.withTxErr != nil {
		return m.withTxErr
	}

	if m.shouldCallTxFn && m.tx != nil {
		m.tx.ctx = ctx

		return fn(m.tx)
	}

	return nil
}

func (m *mockRotationRepository) Shutdown() {
	m.shutdownCalled = true
}

// newMockRotationRepository creates a mockRotationRepository with a mockRotationTransaction for testing.
func newMockRotationRepository() *mockRotationRepository {
	return &mockRotationRepository{
		tx:             &mockRotationTransaction{},
		shouldCallTxFn: true,
	}
}

func setupRotationTestEnvironment(t *testing.T) (*fiber.App, *RotationService, *Service) {
	t.Helper()

	ctx := context.Background()

	// Create in-memory database
	dbID, _ := googleUuid.NewV7()
	dsn := "file:" + dbID.String() + "?mode=memory&cache=shared"

	testSQLDB, err := sql.Open("sqlite", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = testSQLDB.Close() })

	// Configure SQLite
	_, err = testSQLDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
	require.NoError(t, err)
	_, err = testSQLDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;")
	require.NoError(t, err)
	testSQLDB.SetMaxOpenConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	testSQLDB.SetMaxIdleConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	testSQLDB.SetConnMaxLifetime(0)

	// Wrap with GORM
	testDB, err := gorm.Open(sqlite.Dialector{Conn: testSQLDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	// Create barrier tables
	err = testDB.AutoMigrate(&RootKey{}, &IntermediateKey{}, &ContentKey{})
	require.NoError(t, err)

	// Initialize telemetry
	telemetryService, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true).ToTelemetrySettings())
	require.NoError(t, err)
	t.Cleanup(func() { telemetryService.Shutdown() })

	// Initialize JWK gen service
	jwkGenService, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetryService, false)
	require.NoError(t, err)
	t.Cleanup(func() { jwkGenService.Shutdown() })

	// Generate unseal JWK for testing
	_, unsealJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	unsealService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	t.Cleanup(func() { unsealService.Shutdown() })

	// Create barrier repository
	barrierRepo, err := NewGormRepository(testDB)
	require.NoError(t, err)
	t.Cleanup(func() { barrierRepo.Shutdown() })

	// Create barrier service (initializes root and intermediate keys)
	barrierService, err := NewService(ctx, telemetryService, jwkGenService, barrierRepo, unsealService)
	require.NoError(t, err)
	t.Cleanup(func() { barrierService.Shutdown() })

	// Create rotation service
	rotationService, err := NewRotationService(jwkGenService, barrierRepo, unsealService)
	require.NoError(t, err)

	// Create Fiber app for HTTP testing
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Register rotation routes
	RegisterRotationRoutes(app, rotationService)

	return app, rotationService, barrierService
}

func TestRotateRootKey_Success(t *testing.T) {
	app, rotationService, barrierService := setupRotationTestEnvironment(t)

	// Encrypt data before rotation
	clearData := []byte("sensitive data before root rotation")
	encryptedDataBefore, err := barrierService.EncryptContentWithContext(context.Background(), clearData)
	require.NoError(t, err)

	// Make rotation request
	reqBody := map[string]string{
		"reason": "scheduled quarterly rotation",
	}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/admin/api/v1/barrier/rotate/root", bytes.NewReader(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Parse response
	respBody, _ := io.ReadAll(resp.Body)

	var rotateResp RotateRootKeyResponse

	err = json.Unmarshal(respBody, &rotateResp)
	require.NoError(t, err)

	// Verify response
	require.NotEmpty(t, rotateResp.OldKeyUUID)
	require.NotEmpty(t, rotateResp.NewKeyUUID)
	require.NotEqual(t, rotateResp.OldKeyUUID, rotateResp.NewKeyUUID)
	require.Equal(t, "scheduled quarterly rotation", rotateResp.Reason)
	require.Greater(t, rotateResp.RotatedAt, int64(0))

	// Verify old ciphertext still decryptable (elastic rotation)
	decrypted, err := barrierService.DecryptContentWithContext(context.Background(), encryptedDataBefore)
	require.NoError(t, err)
	require.Equal(t, clearData, decrypted)

	// Encrypt new data after rotation
	newClearData := []byte("sensitive data after root rotation")
	encryptedDataAfter, err := barrierService.EncryptContentWithContext(context.Background(), newClearData)
	require.NoError(t, err)

	// Verify new ciphertext decryptable
	decryptedAfter, err := barrierService.DecryptContentWithContext(context.Background(), encryptedDataAfter)
	require.NoError(t, err)
	require.Equal(t, newClearData, decryptedAfter)

	// Verify rotation service was called
	require.NotNil(t, rotationService)
}

func TestRotateIntermediateKey_Success(t *testing.T) {
	app, _, barrierService := setupRotationTestEnvironment(t)

	// Encrypt data before rotation
	clearData := []byte("sensitive data before intermediate rotation")
	encryptedDataBefore, err := barrierService.EncryptContentWithContext(context.Background(), clearData)
	require.NoError(t, err)

	// Make rotation request
	reqBody := map[string]string{
		"reason": "security incident response",
	}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/admin/api/v1/barrier/rotate/intermediate", bytes.NewReader(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Parse response
	respBody, _ := io.ReadAll(resp.Body)

	var rotateResp RotateIntermediateKeyResponse

	err = json.Unmarshal(respBody, &rotateResp)
	require.NoError(t, err)

	// Verify response
	require.NotEmpty(t, rotateResp.OldKeyUUID)
	require.NotEmpty(t, rotateResp.NewKeyUUID)
	require.NotEqual(t, rotateResp.OldKeyUUID, rotateResp.NewKeyUUID)
	require.Equal(t, "security incident response", rotateResp.Reason)
	require.Greater(t, rotateResp.RotatedAt, int64(0))

	// Verify old ciphertext still decryptable (elastic rotation)
	decrypted, err := barrierService.DecryptContentWithContext(context.Background(), encryptedDataBefore)
	require.NoError(t, err)
	require.Equal(t, clearData, decrypted)

	// Encrypt new data after rotation
	newClearData := []byte("sensitive data after intermediate rotation")
	encryptedDataAfter, err := barrierService.EncryptContentWithContext(context.Background(), newClearData)
	require.NoError(t, err)

	// Verify new ciphertext decryptable
	decryptedAfter, err := barrierService.DecryptContentWithContext(context.Background(), encryptedDataAfter)
	require.NoError(t, err)
	require.Equal(t, newClearData, decryptedAfter)
}

func TestRotateContentKey_Success(t *testing.T) {
	app, _, barrierService := setupRotationTestEnvironment(t)

	// Encrypt data before rotation
	clearData := []byte("sensitive data before content rotation")
	encryptedDataBefore, err := barrierService.EncryptContentWithContext(context.Background(), clearData)
	require.NoError(t, err)

	// Make rotation request
	reqBody := map[string]string{
		"reason": "annual security audit requirement",
	}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/admin/api/v1/barrier/rotate/content", bytes.NewReader(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Parse response
	respBody, _ := io.ReadAll(resp.Body)

	var rotateResp RotateContentKeyResponse

	err = json.Unmarshal(respBody, &rotateResp)
	require.NoError(t, err)

	// Verify response
	require.NotEmpty(t, rotateResp.NewKeyUUID)
	require.Equal(t, "annual security audit requirement", rotateResp.Reason)
	require.Greater(t, rotateResp.RotatedAt, int64(0))

	// Verify old ciphertext still decryptable (elastic rotation)
	decrypted, err := barrierService.DecryptContentWithContext(context.Background(), encryptedDataBefore)
	require.NoError(t, err)
	require.Equal(t, clearData, decrypted)

	// Encrypt new data after rotation
	newClearData := []byte("sensitive data after content rotation")
	encryptedDataAfter, err := barrierService.EncryptContentWithContext(context.Background(), newClearData)
	require.NoError(t, err)

	// Verify new ciphertext decryptable
	decryptedAfter, err := barrierService.DecryptContentWithContext(context.Background(), encryptedDataAfter)
	require.NoError(t, err)
	require.Equal(t, newClearData, decryptedAfter)
}

func TestRotateKey_MissingReason(t *testing.T) {
	app, _, _ := setupRotationTestEnvironment(t)

	// Test all three endpoints
	endpoints := []string{
		"/admin/api/v1/barrier/rotate/root",
		"/admin/api/v1/barrier/rotate/intermediate",
		"/admin/api/v1/barrier/rotate/content",
	}

	for _, endpoint := range endpoints {
		t.Run(fmt.Sprintf("endpoint=%s", endpoint), func(t *testing.T) {
			// Empty request body (missing reason field)
			req := httptest.NewRequest("POST", endpoint, bytes.NewReader([]byte("{}")))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

			// Parse error response
			respBody, _ := io.ReadAll(resp.Body)

			var errResp map[string]string

			err = json.Unmarshal(respBody, &errResp)
			require.NoError(t, err)
			require.Equal(t, "validation_error", errResp["error"])
			require.Contains(t, errResp["message"], "at least 10 characters")
		})
	}
}

func TestRotateKey_ShortReason(t *testing.T) {
	app, _, _ := setupRotationTestEnvironment(t)

	// Test all three endpoints
	endpoints := []string{
		"/admin/api/v1/barrier/rotate/root",
		"/admin/api/v1/barrier/rotate/intermediate",
		"/admin/api/v1/barrier/rotate/content",
	}

	for _, endpoint := range endpoints {
		t.Run(fmt.Sprintf("endpoint=%s", endpoint), func(t *testing.T) {
			// Reason too short (less than 10 characters)
			reqBody := map[string]string{
				"reason": "short",
			}
			reqJSON, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("POST", endpoint, bytes.NewReader(reqJSON))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

			// Parse error response
			respBody, _ := io.ReadAll(resp.Body)

			var errResp map[string]string

			err = json.Unmarshal(respBody, &errResp)
			require.NoError(t, err)
			require.Equal(t, "validation_error", errResp["error"])
			require.Contains(t, errResp["message"], "at least 10 characters")
		})
	}
}

// TestRegisterRotationRoutes_NilAdminServer tests that RegisterRotationRoutes panics with nil adminServer.
func TestRegisterRotationRoutes_NilAdminServer(t *testing.T) {
	t.Parallel()

	// Create a mock rotation service.
	rotationService := &RotationService{}

	// Should panic with nil adminServer.
	require.PanicsWithValue(t, "adminServer must be non-nil", func() {
		RegisterRotationRoutes(nil, rotationService)
	})
}

// TestRegisterRotationRoutes_NilRotationService tests that RegisterRotationRoutes panics with nil rotationService.
func TestRegisterRotationRoutes_NilRotationService(t *testing.T) {
	t.Parallel()

	// Create a Fiber app.
	app := fiber.New()

	// Should panic with nil rotationService.
	require.PanicsWithValue(t, "rotationService must be non-nil", func() {
		RegisterRotationRoutes(app, nil)
	})
}

// TestRotateKey_TooLongReason tests that rotation requests fail with too long reason.
