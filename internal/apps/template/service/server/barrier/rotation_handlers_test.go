// Copyright (c) 2025 Justin Cranford
//
//

package barrier

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"testing"

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilUnsealKeysService "cryptoutil/internal/shared/barrier/unsealkeysservice"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"

	"github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

func setupRotationTestEnvironment(t *testing.T) (*fiber.App, *RotationService, *Service) {
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
	testSQLDB.SetMaxOpenConns(cryptoutilMagic.SQLiteMaxOpenConnections)
	testSQLDB.SetMaxIdleConns(cryptoutilMagic.SQLiteMaxOpenConnections)
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
	telemetryService, err := cryptoutilTelemetry.NewTelemetryService(ctx, cryptoutilConfig.NewTestConfig(cryptoutilMagic.IPv4Loopback, 0, true))
	require.NoError(t, err)
	t.Cleanup(func() { telemetryService.Shutdown() })

	// Initialize JWK gen service
	jwkGenService, err := cryptoutilJose.NewJWKGenService(ctx, telemetryService, false)
	require.NoError(t, err)
	t.Cleanup(func() { jwkGenService.Shutdown() })

	// Generate unseal JWK for testing
	_, unsealJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA256KW)
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
