// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

package apis

import (
	"bytes"
	"context"
	"database/sql"
	json "encoding/json"
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // CGO-free SQLite driver

	cryptoutilAppsCipherImDomain "cryptoutil/internal/apps/cipher/im/domain"
	cryptoutilAppsCipherImRepository "cryptoutil/internal/apps/cipher/im/repository"
	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"
)

// createMixedHandler creates a handler with working messageRepo (testDB) but broken recipientJWKRepo (closed DB).
func createMixedHandler(t *testing.T) *MessageHandler {
	t.Helper()

	ctx := context.Background()

	dbID, err := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	require.NoError(t, err)

	dsn := "file:" + dbID.String() + "?mode=memory&cache=shared"

	tempSQLDB, err := sql.Open("sqlite", dsn)
	require.NoError(t, err)

	_, err = tempSQLDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
	require.NoError(t, err)

	_, err = tempSQLDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;")
	require.NoError(t, err)

	tempSQLDB.SetMaxOpenConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	tempSQLDB.SetMaxIdleConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	tempSQLDB.SetConnMaxLifetime(0)

	tempDB, err := gorm.Open(sqlite.Dialector{Conn: tempSQLDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	err = cryptoutilAppsCipherImRepository.ApplyCipherIMMigrations(tempSQLDB, cryptoutilAppsCipherImRepository.DatabaseTypeSQLite)
	require.NoError(t, err)

	err = tempSQLDB.Close()
	require.NoError(t, err)

	brokenRecipientRepo := cryptoutilAppsCipherImRepository.NewMessageRecipientJWKRepository(tempDB, testBarrierService)

	return NewMessageHandler(testMessageRepo, brokenRecipientRepo, testJWKGenService, testBarrierService)
}

// createTriggerDB creates a new in-memory SQLite DB with migrations applied and a trigger installed.
func createTriggerDB(t *testing.T, triggerSQL string) (*gorm.DB, *sql.DB) {
	t.Helper()

	ctx := context.Background()

	dbID, err := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	require.NoError(t, err)

	dsn := "file:" + dbID.String() + "?mode=memory&cache=shared"

	triggerSQLDB, err := sql.Open("sqlite", dsn)
	require.NoError(t, err)

	_, err = triggerSQLDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
	require.NoError(t, err)

	_, err = triggerSQLDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;")
	require.NoError(t, err)

	triggerSQLDB.SetMaxOpenConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	triggerSQLDB.SetMaxIdleConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	triggerSQLDB.SetConnMaxLifetime(0)

	triggerDB, err := gorm.Open(sqlite.Dialector{Conn: triggerSQLDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	err = cryptoutilAppsCipherImRepository.ApplyCipherIMMigrations(triggerSQLDB, cryptoutilAppsCipherImRepository.DatabaseTypeSQLite)
	require.NoError(t, err)

	_, err = triggerSQLDB.ExecContext(ctx, triggerSQL)
	require.NoError(t, err)

	return triggerDB, triggerSQLDB
}

// createShutdownJWKHandler creates a handler with a shutdown JWKGenService to trigger key generation errors.
func createShutdownJWKHandler(t *testing.T) *MessageHandler {
	t.Helper()

	ctx := context.Background()

	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	telemetry, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, settings.ToTelemetrySettings())
	require.NoError(t, err)

	shutdownJWKGen, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetry, false)
	require.NoError(t, err)

	shutdownJWKGen.Shutdown()

	return NewMessageHandler(testMessageRepo, testRecipientRepo, shutdownJWKGen, testBarrierService)
}

// TestHandleSendMessage_JWKGenShutdownError tests the error path when JWK generation fails (block 1: L105).
func TestHandleSendMessage_JWKGenShutdownError(t *testing.T) {
	t.Parallel()

	handler := createShutdownJWKHandler(t)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(testAuthMiddleware())
	app.Post("/messages/send", handler.HandleSendMessage())

	reqBody := SendMessageRequest{
		ReceiverIDs: []string{googleUuid.New().String()},
		Message:     "test message for jwk gen error",
	}

	bodyBytes, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/messages/send", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", googleUuid.New().String())

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

// TestHandleSendMessage_RecipientJWKCreateError tests the error path when recipient JWK creation fails (block 5: L173).
func TestHandleSendMessage_RecipientJWKCreateError(t *testing.T) {
	t.Parallel()

	handler := createMixedHandler(t)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(testAuthMiddleware())
	app.Post("/messages/send", handler.HandleSendMessage())

	reqBody := SendMessageRequest{
		ReceiverIDs: []string{googleUuid.New().String()},
		Message:     "test message for recipient create error",
	}

	bodyBytes, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/messages/send", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", googleUuid.New().String())

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

// TestHandleReceiveMessages_MarkAsReadTriggerError tests the MarkAsRead error path using a SQLite trigger (block 6: L212).
func TestHandleReceiveMessages_MarkAsReadTriggerError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	triggerDB, triggerSQLDB := createTriggerDB(t, "CREATE TRIGGER prevent_update BEFORE UPDATE ON messages BEGIN SELECT RAISE(ABORT, 'update not allowed'); END;")

	defer func() { _ = triggerSQLDB.Close() }()

	senderID := googleUuid.New()
	receiverID := googleUuid.New()
	messageID := googleUuid.New()

	message := &cryptoutilAppsCipherImDomain.Message{
		ID:        messageID,
		SenderID:  senderID,
		JWE:       "test-jwe-mark-as-read-trigger",
		CreatedAt: time.Now().UTC(),
	}

	msgRepo := cryptoutilAppsCipherImRepository.NewMessageRepository(triggerDB)
	require.NoError(t, msgRepo.Create(ctx, message))

	recipientJWK := &cryptoutilAppsCipherImDomain.MessageRecipientJWK{
		ID:           googleUuid.New(),
		MessageID:    messageID,
		RecipientID:  receiverID,
		EncryptedJWK: "trigger-test-encrypted-jwk",
	}

	jwkRepo := cryptoutilAppsCipherImRepository.NewMessageRecipientJWKRepository(triggerDB, testBarrierService)
	require.NoError(t, jwkRepo.Create(ctx, recipientJWK))

	handler := NewMessageHandler(msgRepo, jwkRepo, testJWKGenService, testBarrierService)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(testAuthMiddleware())
	app.Get("/messages/rx", handler.HandleReceiveMessages())

	req := httptest.NewRequest(http.MethodGet, "/messages/rx", nil)
	req.Header.Set("X-User-ID", receiverID.String())

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandleReceiveMessages_JWKDecryptionVariants tests ParseKey and DecryptBytes error paths (blocks 8+9: L240, L249).
func TestHandleReceiveMessages_JWKDecryptionVariants(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Generate a valid JWK for the "wrong key" test case.
	_, _, _, validJWKBytes, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgDir)
	require.NoError(t, err)

	tests := []struct {
		name    string
		jwkData []byte // Data to barrier-encrypt and store as EncryptedJWK.
		jwe     string // Message JWE content.
	}{
		{
			name:    "barrier-encrypted non-JWK causes ParseKey error",
			jwkData: []byte("this-is-not-valid-jwk-json"),
			jwe:     "irrelevant-jwe-data",
		},
		{
			name:    "valid JWK but invalid JWE causes decrypt error",
			jwkData: validJWKBytes,
			jwe:     "not-a-valid-jwe-compact-serialization",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			receiverID := googleUuid.New()
			messageID := googleUuid.New()

			// Barrier-encrypt the JWK data.
			encryptedJWKBytes, err := testBarrierService.EncryptContentWithContext(ctx, tc.jwkData)
			require.NoError(t, err)

			// Insert message directly into test DB.
			message := &cryptoutilAppsCipherImDomain.Message{
				ID:        messageID,
				SenderID:  googleUuid.New(),
				JWE:       tc.jwe,
				CreatedAt: time.Now().UTC(),
			}
			require.NoError(t, testMessageRepo.Create(ctx, message))

			// Insert recipient JWK with barrier-encrypted data.
			recipientJWK := &cryptoutilAppsCipherImDomain.MessageRecipientJWK{
				ID:           googleUuid.New(),
				MessageID:    messageID,
				RecipientID:  receiverID,
				EncryptedJWK: string(encryptedJWKBytes),
			}
			require.NoError(t, testRecipientRepo.Create(ctx, recipientJWK))

			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Use(testAuthMiddleware())
			app.Get("/messages/rx", testMessageHandler.HandleReceiveMessages())

			req := httptest.NewRequest(http.MethodGet, "/messages/rx", nil)
			req.Header.Set("X-User-ID", receiverID.String())

			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, http.StatusOK, resp.StatusCode)

			var response ReceiveMessagesResponse

			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)
			require.Empty(t, response.Messages)
		})
	}
}

// TestHandleDeleteMessage_EmptyMessageID tests the empty message ID path (block 10: L272).
func TestHandleDeleteMessage_EmptyMessageID(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(testAuthMiddleware())

	// Register handler on a route WITHOUT :id param so c.Params("id") returns "".
	app.Delete("/messages/delete", testMessageHandler.HandleDeleteMessage())

	req := httptest.NewRequest(http.MethodDelete, "/messages/delete", nil)
	req.Header.Set("X-User-ID", googleUuid.New().String())

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestHandleDeleteMessage_RecipientJWKDeleteFailed tests the DeleteByMessageID error path (block 11: L314).
func TestHandleDeleteMessage_RecipientJWKDeleteFailed(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	senderID := googleUuid.New()
	messageID := googleUuid.New()

	message := &cryptoutilAppsCipherImDomain.Message{
		ID:        messageID,
		SenderID:  senderID,
		JWE:       "test-jwe-recipient-delete-error",
		CreatedAt: time.Now().UTC(),
	}
	require.NoError(t, testMessageRepo.Create(ctx, message))

	handler := createMixedHandler(t)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(testAuthMiddleware())
	app.Delete("/messages/:id", handler.HandleDeleteMessage())

	req := httptest.NewRequest(http.MethodDelete, "/messages/"+messageID.String(), nil)
	req.Header.Set("X-User-ID", senderID.String())

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

// TestHandleDeleteMessage_MessageDeleteTriggerError tests the Delete error path using a SQLite trigger (block 12: L322).
func TestHandleDeleteMessage_MessageDeleteTriggerError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	triggerDB, triggerSQLDB := createTriggerDB(t,
		"CREATE TRIGGER prevent_msg_delete BEFORE DELETE ON messages BEGIN SELECT RAISE(ABORT, 'delete not allowed'); END;")

	defer func() { _ = triggerSQLDB.Close() }()

	senderID := googleUuid.New()
	messageID := googleUuid.New()

	message := &cryptoutilAppsCipherImDomain.Message{
		ID:        messageID,
		SenderID:  senderID,
		JWE:       "test-jwe-delete-trigger",
		CreatedAt: time.Now().UTC(),
	}

	msgRepo := cryptoutilAppsCipherImRepository.NewMessageRepository(triggerDB)
	require.NoError(t, msgRepo.Create(ctx, message))

	jwkRepo := cryptoutilAppsCipherImRepository.NewMessageRecipientJWKRepository(triggerDB, testBarrierService)

	handler := NewMessageHandler(msgRepo, jwkRepo, testJWKGenService, testBarrierService)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(testAuthMiddleware())
	app.Delete("/messages/:id", handler.HandleDeleteMessage())

	req := httptest.NewRequest(http.MethodDelete, "/messages/"+messageID.String(), nil)
	req.Header.Set("X-User-ID", senderID.String())

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
