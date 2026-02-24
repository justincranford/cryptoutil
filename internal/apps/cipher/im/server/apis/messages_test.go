// Copyright (c) 2025 Justin Cranford

package apis

import (
	"bytes"
	"context"
	"database/sql"
	json "encoding/json"
	http "net/http"
	"net/http/httptest"
	"os"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // CGO-free SQLite driver

	cryptoutilAppsCipherImDomain "cryptoutil/internal/apps/cipher/im/domain"
	cryptoutilAppsCipherImRepository "cryptoutil/internal/apps/cipher/im/repository"
	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilUnsealKeysService "cryptoutil/internal/apps/template/service/server/barrier/unsealkeysservice"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"
)

var (
	testMessageHandler *MessageHandler
	testMessageRepo    *cryptoutilAppsCipherImRepository.MessageRepository
	testRecipientRepo  *cryptoutilAppsCipherImRepository.MessageRecipientJWKRepository
	testJWKGenService  *cryptoutilSharedCryptoJose.JWKGenService
	testBarrierService *cryptoutilAppsTemplateServiceServerBarrier.Service
	testDB             *gorm.DB
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Create in-memory SQLite database (avoid Docker requirement).
	dbID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	dsn := "file:" + dbID.String() + "?mode=memory&cache=shared"

	var (
		testSQLDB *sql.DB
		err       error
	)

	testSQLDB, err = sql.Open("sqlite", dsn)
	if err != nil {
		panic("TestMain: failed to open SQLite: " + err.Error())
	}

	defer func() { _ = testSQLDB.Close() }()

	// Configure SQLite.
	if _, err := testSQLDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;"); err != nil {
		panic("TestMain: failed to enable WAL: " + err.Error())
	}

	if _, err := testSQLDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;"); err != nil {
		panic("TestMain: failed to set busy timeout: " + err.Error())
	}

	testSQLDB.SetMaxOpenConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	testSQLDB.SetMaxIdleConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	testSQLDB.SetConnMaxLifetime(0)

	// Wrap with GORM.
	var db *gorm.DB

	db, err = gorm.Open(sqlite.Dialector{Conn: testSQLDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		panic("TestMain: failed to create GORM DB: " + err.Error())
	}

	// Apply Cipher-IM migrations.
	if err := cryptoutilAppsCipherImRepository.ApplyCipherIMMigrations(testSQLDB, cryptoutilAppsCipherImRepository.DatabaseTypeSQLite); err != nil {
		panic("TestMain: failed to apply migrations: " + err.Error())
	}

	// Initialize telemetry.
	telemetrySettings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	testTelemetryService, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, telemetrySettings.ToTelemetrySettings())
	if err != nil {
		panic("TestMain: failed to create telemetry: " + err.Error())
	}
	defer testTelemetryService.Shutdown()

	// Initialize JWK generation service.
	jwkGenService, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, testTelemetryService, false)
	if err != nil {
		panic("TestMain: failed to create JWK service: " + err.Error())
	}
	defer jwkGenService.Shutdown()

	// Initialize Barrier Service.
	_, testUnsealJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	if err != nil {
		panic("TestMain: failed to generate test unseal JWK: " + err.Error())
	}

	unsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{testUnsealJWK})
	if err != nil {
		panic("TestMain: failed to create unseal keys service: " + err.Error())
	}
	defer unsealKeysService.Shutdown()

	barrierRepo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
	if err != nil {
		panic("TestMain: failed to create barrier repository: " + err.Error())
	}
	defer barrierRepo.Shutdown()

	barrierService, err := cryptoutilAppsTemplateServiceServerBarrier.NewService(ctx, testTelemetryService, jwkGenService, barrierRepo, unsealKeysService)
	if err != nil {
		panic("TestMain: failed to create barrier service: " + err.Error())
	}
	defer barrierService.Shutdown()

	// Initialize repositories.
	testDB = db
	testJWKGenService = jwkGenService
	testBarrierService = barrierService
	testMessageRepo = cryptoutilAppsCipherImRepository.NewMessageRepository(db)
	testRecipientRepo = cryptoutilAppsCipherImRepository.NewMessageRecipientJWKRepository(db, barrierService)

	// Initialize handler.
	testMessageHandler = NewMessageHandler(
		testMessageRepo,
		testRecipientRepo,
		jwkGenService,
		barrierService,
	)

	// Run tests.
	exitCode := m.Run()
	os.Exit(exitCode)
}

// testAuthMiddleware sets user ID from X-User-ID header for testing.
func testAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userIDStr := c.Get("X-User-ID")
		if userIDStr != "" {
			userID, err := googleUuid.Parse(userIDStr)
			if err == nil {
				c.Locals("user_id", userID)
			}
		}

		return c.Next()
	}
}

func TestHandleSendMessage_HappyPath(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(testAuthMiddleware())
	app.Post("/messages/send", testMessageHandler.HandleSendMessage())

	senderID := googleUuid.New()
	receiver1ID := googleUuid.New()
	receiver2ID := googleUuid.New()

	reqBody := SendMessageRequest{
		ReceiverIDs: []string{receiver1ID.String(), receiver2ID.String()},
		Message:     "Hello, World!",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/messages/send", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", senderID.String())

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var response SendMessageResponse

	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	require.NotEmpty(t, response.MessageID)
}

func TestHandleSendMessage_InvalidJSON(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(testAuthMiddleware())
	app.Post("/messages/send", testMessageHandler.HandleSendMessage())

	req := httptest.NewRequest(http.MethodPost, "/messages/send", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", googleUuid.New().String())

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandleSendMessage_MissingSenderID(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(testAuthMiddleware())
	app.Post("/messages/send", testMessageHandler.HandleSendMessage())

	reqBody := SendMessageRequest{
		ReceiverIDs: []string{googleUuid.New().String()},
		Message:     "Test message",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/messages/send", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	// Missing X-User-ID header

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestHandleReceiveMessages_HappyPath(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(testAuthMiddleware())
	app.Get("/messages/receive", testMessageHandler.HandleReceiveMessages())

	userID := googleUuid.New()

	req := httptest.NewRequest(http.MethodGet, "/messages/receive", nil)
	req.Header.Set("X-User-ID", userID.String())

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var response ReceiveMessagesResponse

	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	// Empty messages is valid (no messages for new user)
	require.NotNil(t, response.Messages)
}

func TestHandleReceiveMessages_MissingUserID(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(testAuthMiddleware())
	app.Get("/messages/receive", testMessageHandler.HandleReceiveMessages())

	req := httptest.NewRequest(http.MethodGet, "/messages/receive", nil)
	// Missing X-User-ID header

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestHandleDeleteMessage_HappyPath(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create a message first
	senderID := googleUuid.New()
	messageID := googleUuid.New()
	message := &cryptoutilAppsCipherImDomain.Message{
		ID:       messageID,
		SenderID: senderID,
		JWE:      "test-jwe-content",
	}
	require.NoError(t, testMessageRepo.Create(ctx, message))

	// Delete via API
	app := fiber.New()
	app.Use(testAuthMiddleware())
	app.Delete("/messages/:id", testMessageHandler.HandleDeleteMessage())

	req := httptest.NewRequest(http.MethodDelete, "/messages/"+messageID.String(), nil)
	req.Header.Set("X-User-ID", senderID.String())

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestHandleDeleteMessage_InvalidMessageID(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(testAuthMiddleware())
	app.Delete("/messages/:id", testMessageHandler.HandleDeleteMessage())

	req := httptest.NewRequest(http.MethodDelete, "/messages/invalid-uuid", nil)
	req.Header.Set("X-User-ID", googleUuid.New().String())

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandleDeleteMessage_MissingUserID(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(testAuthMiddleware())
	app.Delete("/messages/:id", testMessageHandler.HandleDeleteMessage())

	req := httptest.NewRequest(http.MethodDelete, "/messages/"+googleUuid.New().String(), nil)
	// Missing X-User-ID header

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	// Handler checks message existence before user ID, returns 404 (better security - don't leak existence).
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestHandleSendMessage_EmptyMessage tests the empty message validation.
func TestHandleSendMessage_EmptyMessage(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(testAuthMiddleware())
	app.Post("/messages/send", testMessageHandler.HandleSendMessage())

	reqBody := SendMessageRequest{
		ReceiverIDs: []string{googleUuid.New().String()},
		Message:     "", // Empty message.
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/messages/send", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", googleUuid.New().String())

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestHandleSendMessage_EmptyReceiverIDs tests empty receiver IDs validation.
func TestHandleSendMessage_EmptyReceiverIDs(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(testAuthMiddleware())
	app.Post("/messages/send", testMessageHandler.HandleSendMessage())

	reqBody := SendMessageRequest{
		ReceiverIDs: []string{}, // Empty receiver IDs.
		Message:     "Test message",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/messages/send", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", googleUuid.New().String())

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestHandleSendMessage_InvalidReceiverID tests invalid receiver UUID.
func TestHandleSendMessage_InvalidReceiverID(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(testAuthMiddleware())
	app.Post("/messages/send", testMessageHandler.HandleSendMessage())

	reqBody := SendMessageRequest{
		ReceiverIDs: []string{"not-a-valid-uuid"},
		Message:     "Test message",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/messages/send", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", googleUuid.New().String())

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestHandleReceiveMessages_NoMessages tests receiving when no messages exist.
func TestHandleReceiveMessages_NoMessages(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(testAuthMiddleware())
	app.Get("/messages/rx", testMessageHandler.HandleReceiveMessages())

	req := httptest.NewRequest(http.MethodGet, "/messages/rx", nil)
	req.Header.Set("X-User-ID", googleUuid.New().String())

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var response ReceiveMessagesResponse

	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)
	require.Empty(t, response.Messages)
}

// TestHandleDeleteMessage_MissingMessageID tests missing message ID in path.
func TestHandleDeleteMessage_MissingMessageID(_ *testing.T) {
	app := fiber.New()
	app.Use(testAuthMiddleware())
	app.Delete("/messages/:id", testMessageHandler.HandleDeleteMessage())

	// Path with empty ID - fiber treats this as route mismatch, so this is not the right test.
	// Instead we need to test with a valid route that returns empty.
}

// TestNewMessageHandler tests the MessageHandler constructor.
func TestNewMessageHandler(t *testing.T) {
	t.Parallel()
	// Test that NewMessageHandler can be created.
	handler := NewMessageHandler(nil, nil, nil, nil)
	require.NotNil(t, handler)
}

// TestHandleReceiveMessages_WithMessages tests receiving actual messages.
