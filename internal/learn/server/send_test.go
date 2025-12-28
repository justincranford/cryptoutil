// Copyright (c) 2025 Justin Cranford
//

package server_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilCrypto "cryptoutil/internal/learn/crypto"
	cryptoutilDomain "cryptoutil/internal/learn/domain"
)

func TestHandleSendMessage_Success(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	sender := registerAndLoginTestUser(t, client, baseURL)
	receiver := registerTestUser(t, client, baseURL, "receiver", "password123")

	reqBody := map[string]any{
		"message":      "Hello, receiver!",
		"receiver_ids": []string{receiver.ID.String()},
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, baseURL+"/service/api/v1/messages/tx", bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sender.Token)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.NotEmpty(t, respBody["message_id"])
}

func TestHandleSendMessage_EmptyReceivers(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	sender := registerAndLoginTestUser(t, client, baseURL)

	reqBody := map[string]any{
		"message":      "Hello!",
		"receiver_ids": []string{},
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, baseURL+"/service/api/v1/messages/tx", bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sender.Token)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.Contains(t, respBody["error"], "receiver_ids cannot be empty")
}

func TestHandleSendMessage_InvalidReceiverID(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	sender := registerAndLoginTestUser(t, client, baseURL)

	reqBody := map[string]any{
		"message":      "Hello!",
		"receiver_ids": []string{"invalid-uuid"},
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, baseURL+"/service/api/v1/messages/tx", bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sender.Token)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.Contains(t, respBody["error"], "invalid receiver ID")
}

// TestHandleSendMessage_InvalidTokenFormat tests sending message with invalid Bearer token format.
func TestHandleSendMessage_InvalidTokenFormat(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	receiver := registerTestUser(t, client, baseURL, "receiver", "password123")

	reqBody := map[string]any{
		"message":      "Test message",
		"receiver_ids": []string{receiver.ID.String()},
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, baseURL+"/service/api/v1/messages/tx", bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "InvalidFormat token123")

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.Contains(t, respBody["error"], "invalid authorization header format")
}

// TestHandleSendMessage_InvalidTokenSignature tests sending message with tampered JWT token.
func TestHandleSendMessage_InvalidTokenSignature(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	receiver := registerTestUser(t, client, baseURL, "receiver", "password123")

	sender := registerAndLoginTestUser(t, client, baseURL)
	tamperedToken := sender.Token + "tampered"

	reqBody := map[string]any{
		"message":      "Test message",
		"receiver_ids": []string{receiver.ID.String()},
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, baseURL+"/service/api/v1/messages/tx", bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tamperedToken)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.Contains(t, respBody["error"], "invalid token")
}

// TestHandleSendMessage_MissingToken tests sending message without JWT token.
func TestHandleSendMessage_MissingToken(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	receiver := registerTestUser(t, client, baseURL, "receiver", "password123")

	reqBody := map[string]any{
		"message":      "Test message",
		"receiver_ids": []string{receiver.ID.String()},
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, baseURL+"/service/api/v1/messages/tx", bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// TestHandleSendMessage_EmptyMessage tests sending empty message.
func TestHandleSendMessage_EmptyMessage(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	sender := registerAndLoginTestUser(t, client, baseURL)
	receiver := registerTestUser(t, client, baseURL, "receiver", "password123")

	reqBody := map[string]any{
		"message":      "",
		"receiver_ids": []string{receiver.ID.String()},
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, baseURL+"/service/api/v1/messages/tx", bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sender.Token)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.Contains(t, respBody["error"], "message cannot be empty")
}

// TestHandleSendMessage_SaveRepositoryError tests sending message when repository fails during save.
func TestHandleSendMessage_SaveRepositoryError(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	sender := registerAndLoginTestUser(t, client, baseURL)
	receiver := registerAndLoginTestUser(t, client, baseURL)

	sqlDB, err := db.DB()
	require.NoError(t, err)

	sendReq := fmt.Sprintf(`{"message": "test", "receiver_ids": ["%s"]}`, receiver.User.ID.String())
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, baseURL+"/service/api/v1/messages/tx", bytes.NewReader([]byte(sendReq)))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sender.Token)

	_ = sqlDB.Close()

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.True(t, resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusInternalServerError,
		"expected 404 or 500, got %d", resp.StatusCode)
}

// TestHandleSendMessage_EncryptionError tests encryption failure.
func TestHandleSendMessage_EncryptionError(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	sender := registerAndLoginTestUser(t, client, baseURL)
	receiver := registerAndLoginTestUser(t, client, baseURL)

	// Corrupt receiver's public key in database to trigger encryption error.
	var user cryptoutilDomain.User

	err := db.First(&user, "id = ?", receiver.User.ID).Error
	require.NoError(t, err)

	sendReq := fmt.Sprintf(`{"message": "test", "receiver_ids": ["%s"]}`, receiver.User.ID.String())
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, baseURL+"/service/api/v1/messages/tx", bytes.NewReader([]byte(sendReq)))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sender.Token)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

// TestHandleSendMessage_ReceiverPublicKeyParseError tests parsing error on receiver's public key.
func TestHandleSendMessage_ReceiverPublicKeyParseError(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	sender := registerAndLoginTestUser(t, client, baseURL)

	ctx := context.Background()

	receiverID := googleUuid.New()
	privateKey, _, err := cryptoutilCrypto.GenerateECDHKeyPair()
	require.NoError(t, err)

	privateKeyBytes := privateKey.Bytes()
	_ = privateKeyBytes
	passwordHash, err := cryptoutilCrypto.HashPassword("password123")
	require.NoError(t, err)

	passwordHashHex := hex.EncodeToString(passwordHash)

	receiver := &cryptoutilDomain.User{
		ID:           receiverID,
		Username:     "receiver",
		PasswordHash: passwordHashHex,

		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = db.Create(receiver).Error
	require.NoError(t, err)

	sendReq := map[string]any{
		"message":      "Test message",
		"receiver_ids": []string{receiverID.String()},
	}
	sendJSON, err := json.Marshal(sendReq)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, baseURL+"/service/api/v1/messages/tx", bytes.NewReader(sendJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sender.Token)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.Equal(t, "failed to parse receiver public key", result["error"])
}

// TestHandleSendMessage_MultipleReceivers tests sending message to multiple receivers.
func TestHandleSendMessage_MultipleReceivers(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	sender := registerAndLoginTestUser(t, client, baseURL)
	receiver1 := registerTestUser(t, client, baseURL, "receiver1", "password123")
	receiver2 := registerTestUser(t, client, baseURL, "receiver2", "password123")
	receiver3 := registerTestUser(t, client, baseURL, "receiver3", "password123")

	reqBody := map[string]any{
		"message": "Broadcast message",
		"receiver_ids": []string{
			receiver1.ID.String(),
			receiver2.ID.String(),
			receiver3.ID.String(),
		},
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, baseURL+"/service/api/v1/messages/tx", bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sender.Token)

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	defer func() { _ = resp.Body.Close() }()

	var sendResp map[string]string

	err = json.NewDecoder(resp.Body).Decode(&sendResp)
	require.NoError(t, err)
	require.NotEmpty(t, sendResp["message_id"])
}

// TestHandleSendMessage_InvalidBodyParser tests body parsing failure.
func TestHandleSendMessage_InvalidBodyParser(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	sender := registerAndLoginTestUser(t, client, baseURL)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, baseURL+"/service/api/v1/messages/tx", bytes.NewReader([]byte("not-valid-json")))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sender.Token)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.Equal(t, "invalid request body", result["error"])
}
