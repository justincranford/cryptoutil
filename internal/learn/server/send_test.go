// Copyright (c) 2025 Justin Cranford
//

package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilRandom "cryptoutil/internal/shared/util/random"
)

func TestHandleSendMessage_Success(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	sender := registerAndLoginTestUser(t, client, baseURL)

	receiverUsername, err := cryptoutilRandom.GenerateUsernameSimple()
	require.NoError(t, err)

	receiverPassword, err := cryptoutilRandom.GeneratePasswordSimple()
	require.NoError(t, err)

	receiver := registerTestUser(t, client, baseURL, receiverUsername, receiverPassword)

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
	require.Contains(t, respBody["error"], "invalid recipient ID")
}

// TestHandleSendMessage_InvalidTokenFormat tests sending message with invalid Bearer token format.
func TestHandleSendMessage_InvalidTokenFormat(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	receiverUsername, err := cryptoutilRandom.GenerateUsernameSimple()
	require.NoError(t, err)

	receiverPassword, err := cryptoutilRandom.GeneratePasswordSimple()
	require.NoError(t, err)

	receiver := registerTestUser(t, client, baseURL, receiverUsername, receiverPassword)

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

	receiverUsername, err := cryptoutilRandom.GenerateUsernameSimple()
	require.NoError(t, err)

	receiverPassword, err := cryptoutilRandom.GeneratePasswordSimple()
	require.NoError(t, err)

	receiver := registerTestUser(t, client, baseURL, receiverUsername, receiverPassword)

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

	receiverPassword, err := cryptoutilRandom.GeneratePasswordSimple()
	require.NoError(t, err)
	receiver := registerTestUser(t, client, baseURL, "receiver", receiverPassword)

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

	receiverUsername, err := cryptoutilRandom.GenerateUsernameSimple()
	require.NoError(t, err)

	receiverPassword, err := cryptoutilRandom.GeneratePasswordSimple()
	require.NoError(t, err)

	receiver := registerTestUser(t, client, baseURL, receiverUsername, receiverPassword)

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
// DISABLED: Cannot close shared database in TestMain pattern - breaks parallel tests.
// TODO: Rewrite using mock repository instead of closing shared database.
/*
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
*/

// TestHandleSendMessage_EncryptionError tests encryption failure.
func TestHandleSendMessage_EncryptionError(t *testing.T) {
	t.Parallel()

	// TODO: Implement proper encryption error testing when mocking infrastructure available.
	// Currently difficult to trigger encryption errors without mocking JWKGenService.
	// The encryption happens in cryptoutilJose.EncryptBytes() which is hard to corrupt
	// without proper dependency injection for testing.
	t.Skip("Encryption error testing requires mocking infrastructure - deferred to Phase 10")
}

// TestHandleSendMessage_MultipleReceivers tests sending message to multiple receivers.
func TestHandleSendMessage_MultipleReceivers(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	sender := registerAndLoginTestUser(t, client, baseURL)

	receiver1Username, err := cryptoutilRandom.GenerateUsernameSimple()
	require.NoError(t, err)

	receiver1Password, err := cryptoutilRandom.GeneratePasswordSimple()
	require.NoError(t, err)

	receiver1 := registerTestUser(t, client, baseURL, receiver1Username, receiver1Password)

	receiver2Username, err := cryptoutilRandom.GenerateUsernameSimple()
	require.NoError(t, err)

	receiver2Password, err := cryptoutilRandom.GeneratePasswordSimple()
	require.NoError(t, err)

	receiver2 := registerTestUser(t, client, baseURL, receiver2Username, receiver2Password)

	receiver3Username, err := cryptoutilRandom.GenerateUsernameSimple()
	require.NoError(t, err)

	receiver3Password, err := cryptoutilRandom.GeneratePasswordSimple()
	require.NoError(t, err)

	receiver3 := registerTestUser(t, client, baseURL, receiver3Username, receiver3Password)

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
