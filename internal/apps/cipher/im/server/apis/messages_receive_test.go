// Copyright (c) 2025 Justin Cranford

package apis

import (
	"bytes"
	"context"
	json "encoding/json"
	http "net/http"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite" // CGO-free SQLite driver

	cryptoutilAppsCipherImDomain "cryptoutil/internal/apps/cipher/im/domain"
)

func TestHandleReceiveMessages_WithMessages(t *testing.T) {
	t.Parallel()
	// Create sender and receiver.
	senderID := googleUuid.New()
	receiverID := googleUuid.New()

	// Send a message first using the API.
	sendApp := fiber.New()
	sendApp.Use(testAuthMiddleware())
	sendApp.Post("/messages/send", testMessageHandler.HandleSendMessage())

	sendReqBody := SendMessageRequest{
		ReceiverIDs: []string{receiverID.String()},
		Message:     "Test message for receive test",
	}
	sendBodyBytes, _ := json.Marshal(sendReqBody)

	sendReq := httptest.NewRequest(http.MethodPost, "/messages/send", bytes.NewReader(sendBodyBytes))
	sendReq.Header.Set("Content-Type", "application/json")
	sendReq.Header.Set("X-User-ID", senderID.String())

	sendResp, err := sendApp.Test(sendReq)
	require.NoError(t, err)

	defer func() { _ = sendResp.Body.Close() }()

	require.Equal(t, http.StatusCreated, sendResp.StatusCode)

	// Now receive messages as the recipient.
	receiveApp := fiber.New()
	receiveApp.Use(testAuthMiddleware())
	receiveApp.Get("/messages/receive", testMessageHandler.HandleReceiveMessages())

	receiveReq := httptest.NewRequest(http.MethodGet, "/messages/receive", nil)
	receiveReq.Header.Set("X-User-ID", receiverID.String())

	receiveResp, err := receiveApp.Test(receiveReq)
	require.NoError(t, err)

	defer func() { _ = receiveResp.Body.Close() }()

	require.Equal(t, http.StatusOK, receiveResp.StatusCode)

	var response ReceiveMessagesResponse

	err = json.NewDecoder(receiveResp.Body).Decode(&response)
	require.NoError(t, err)

	// Should have at least one message.
	require.NotEmpty(t, response.Messages)
}

// TestHandleReceiveMessages_CorruptedJWK tests behavior with corrupted JWK data.
func TestHandleReceiveMessages_CorruptedJWK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create a message with corrupted encrypted JWK.
	senderID := googleUuid.New()
	receiverID := googleUuid.New()
	messageID := googleUuid.New()

	message := &cryptoutilAppsCipherImDomain.Message{
		ID:       messageID,
		SenderID: senderID,
		JWE:      "corrupted-jwe-content-not-valid",
	}
	require.NoError(t, testMessageRepo.Create(ctx, message))

	// Create corrupted recipient JWK record.
	recipientJWK := &cryptoutilAppsCipherImDomain.MessageRecipientJWK{
		ID:           googleUuid.New(),
		MessageID:    messageID,
		RecipientID:  receiverID,
		EncryptedJWK: "corrupted-encrypted-jwk-data",
	}
	require.NoError(t, testRecipientRepo.Create(ctx, recipientJWK))

	// Try to receive messages - should handle error gracefully.
	app := fiber.New()
	app.Use(testAuthMiddleware())
	app.Get("/messages/receive", testMessageHandler.HandleReceiveMessages())

	req := httptest.NewRequest(http.MethodGet, "/messages/receive", nil)
	req.Header.Set("X-User-ID", receiverID.String())

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	// Should return 200 OK (corrupted messages are skipped, not errors).
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var response ReceiveMessagesResponse

	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	// Message should be skipped due to decryption failure.
	require.Empty(t, response.Messages)
}

// TestHandleReceiveMessages_NoJWKForRecipient tests behavior when no JWK exists.
func TestHandleReceiveMessages_NoJWKForRecipient(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create a message WITHOUT recipient JWK (simulates missing encryption key).
	senderID := googleUuid.New()
	receiverID := googleUuid.New()
	messageID := googleUuid.New()

	message := &cryptoutilAppsCipherImDomain.Message{
		ID:       messageID,
		SenderID: senderID,
		JWE:      "some-jwe-content",
	}
	require.NoError(t, testMessageRepo.Create(ctx, message))

	// Note: NOT creating recipient JWK record.

	// Try to receive messages - should handle missing JWK gracefully.
	app := fiber.New()
	app.Use(testAuthMiddleware())
	app.Get("/messages/receive", testMessageHandler.HandleReceiveMessages())

	req := httptest.NewRequest(http.MethodGet, "/messages/receive", nil)
	req.Header.Set("X-User-ID", receiverID.String())

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	// Should return 200 OK (messages without JWK are skipped).
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var response ReceiveMessagesResponse

	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	// Message should be skipped due to missing JWK.
	require.Empty(t, response.Messages)
}

// TestHandleDeleteMessage_ForbiddenNotOwner tests deleting message by non-owner.
func TestHandleDeleteMessage_ForbiddenNotOwner(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create a message as one user.
	senderID := googleUuid.New()
	otherUserID := googleUuid.New()
	messageID := googleUuid.New()

	message := &cryptoutilAppsCipherImDomain.Message{
		ID:       messageID,
		SenderID: senderID,
		JWE:      "test-jwe-content",
	}
	require.NoError(t, testMessageRepo.Create(ctx, message))

	// Try to delete as different user.
	app := fiber.New()
	app.Use(testAuthMiddleware())
	app.Delete("/messages/:id", testMessageHandler.HandleDeleteMessage())

	req := httptest.NewRequest(http.MethodDelete, "/messages/"+messageID.String(), nil)
	req.Header.Set("X-User-ID", otherUserID.String())

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

// TestHandleDeleteMessage_NotFound tests deleting non-existent message.
func TestHandleDeleteMessage_NotFound(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(testAuthMiddleware())
	app.Delete("/messages/:id", testMessageHandler.HandleDeleteMessage())

	nonExistentID := googleUuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/messages/"+nonExistentID.String(), nil)
	req.Header.Set("X-User-ID", googleUuid.New().String())

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestHandleSendMessage_GenerateJWKFailure would require mocking the JWK service.
// This is a defensive code path that's hard to test without mocks.

// TestHandleSendMessage_EncryptMessageFailure would require mocking the encryption.
// This is a defensive code path that's hard to test without mocks.

// TestHandleDeleteMessage_ExistingMessageNoAuth tests that HandleDeleteMessage returns
// 401 when a message exists but no user authentication context is set.
// This covers the !ok branch for userID extraction after a successful FindByID.
func TestHandleDeleteMessage_ExistingMessageNoAuth(t *testing.T) {
        t.Parallel()

        ctx := context.Background()

        // Create a message so FindByID succeeds.
        senderID := googleUuid.New()
        messageID := googleUuid.New()
        message := &cryptoutilAppsCipherImDomain.Message{
                ID:       messageID,
                SenderID: senderID,
                JWE:      "test-jwe-no-auth",
        }
        require.NoError(t, testMessageRepo.Create(ctx, message))

        // Build handler with NO auth middleware (so user_id local is never set).
        app := fiber.New()
        // Deliberately omit testAuthMiddleware so c.Locals(ContextKeyUserID) stays nil.
        app.Delete("/messages/:id", testMessageHandler.HandleDeleteMessage())

        req := httptest.NewRequest(http.MethodDelete, "/messages/"+messageID.String(), nil)
        // No X-User-ID header, no middleware to set the local.

        resp, err := app.Test(req, -1)
        require.NoError(t, err)

        defer func() { _ = resp.Body.Close() }()

        // Should return 401 because message exists but user context is missing.
        require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
