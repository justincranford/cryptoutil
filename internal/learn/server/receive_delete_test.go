// Copyright (c) 2025 Justin Cranford
//

package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilLearnDomain "cryptoutil/internal/learn/domain"
	cryptoutilRandom "cryptoutil/internal/shared/util/random"
)

func TestHandleReceiveMessages_Empty(t *testing.T) {
	// NOTE: t.Parallel() removed due to cleanTestDB() data isolation issue
	// See: learn_test_isolation_issue.txt
	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	receiver := registerAndLoginTestUser(t, client, baseURL)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, baseURL+"/service/api/v1/messages/rx", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+receiver.Token)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody map[string][]any

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.Empty(t, respBody["messages"])
}

func TestHandleDeleteMessage_Success(t *testing.T) {
	// NOTE: t.Parallel() removed due to cleanTestDB() data isolation issue
	// See: learn_test_isolation_issue.txt

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	sender := registerAndLoginTestUser(t, client, baseURL)
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
	req.Header.Set("Authorization", "Bearer "+sender.Token)

	resp, err := client.Do(req)
	require.NoError(t, err)

	var sendResp map[string]string

	err = json.NewDecoder(resp.Body).Decode(&sendResp)
	require.NoError(t, err)

	_ = resp.Body.Close()

	messageID := sendResp["message_id"]

	// Get message from database to verify it exists.
	var message cryptoutilLearnDomain.Message

	err = db.Where("id = ?", messageID).First(&message).Error
	require.NoError(t, err)

	req, err = http.NewRequestWithContext(context.Background(), http.MethodDelete, baseURL+"/service/api/v1/messages/"+messageID, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+sender.Token)

	resp, err = client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestHandleDeleteMessage_InvalidID(t *testing.T) {
	// NOTE: t.Parallel() removed due to cleanTestDB() data isolation issue
	// See: learn_test_isolation_issue.txt

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	sender := registerAndLoginTestUser(t, client, baseURL)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, baseURL+"/service/api/v1/messages/invalid-id", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+sender.Token)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.Contains(t, respBody["error"], "invalid message ID")
}

func TestHandleDeleteMessage_NotFound(t *testing.T) {
	// NOTE: t.Parallel() removed due to cleanTestDB() data isolation issue
	// See: learn_test_isolation_issue.txt

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	sender := registerAndLoginTestUser(t, client, baseURL)

	messageID := googleUuid.New()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, baseURL+"/service/api/v1/messages/"+messageID.String(), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+sender.Token)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusNotFound, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.Contains(t, respBody["error"], "message not found")
}

// TestHandleReceiveMessages_MissingToken tests receiving messages without JWT token.
func TestHandleReceiveMessages_MissingToken(t *testing.T) {
	// NOTE: t.Parallel() removed due to cleanTestDB() data isolation issue
	// See: learn_test_isolation_issue.txt

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, baseURL+"/service/api/v1/messages/rx", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// TestHandleDeleteMessage_MissingToken tests deleting message without JWT token.
func TestHandleDeleteMessage_MissingToken(t *testing.T) {
	// NOTE: t.Parallel() removed due to cleanTestDB() data isolation issue
	// See: learn_test_isolation_issue.txt

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	messageID := googleUuid.New()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, baseURL+"/service/api/v1/messages/"+messageID.String(), nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// TestHandleDeleteMessage_NotOwner tests deleting message user doesn't own.
func TestHandleDeleteMessage_NotOwner(t *testing.T) {
	// NOTE: t.Parallel() removed due to cleanTestDB() data isolation issue
	// See: learn_test_isolation_issue.txt
	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	sender := registerAndLoginTestUser(t, client, baseURL)
	receiver := registerAndLoginTestUser(t, client, baseURL)

	reqBody := map[string]any{
		"message":      "Test message",
		"receiver_ids": []string{receiver.User.ID.String()},
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, baseURL+"/service/api/v1/messages/tx", bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sender.Token)

	resp, err := client.Do(req)
	require.NoError(t, err)

	var sendResp map[string]string

	err = json.NewDecoder(resp.Body).Decode(&sendResp)
	require.NoError(t, err)

	_ = resp.Body.Close()

	messageID := sendResp["message_id"]

	req, err = http.NewRequestWithContext(context.Background(), http.MethodDelete, baseURL+"/service/api/v1/messages/"+messageID, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+receiver.Token)

	resp, err = client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusForbidden, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.Contains(t, respBody["error"], "only the sender can delete")
}

// TestHandleDeleteMessage_RepositoryError tests deleting message when repository fails.
// DISABLED: Cannot close shared database in TestMain pattern.
// This test needs to use a mock repository instead of closing the real database.
// Closing shared database breaks all parallel tests that run after this test.
/*
func TestHandleDeleteMessage_RepositoryError(t *testing.T) {
	// NOTE: t.Parallel() removed due to cleanTestDB() data isolation issue
	// See: learn_test_isolation_issue.txt

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	sender := registerAndLoginTestUser(t, client, baseURL)
	receiver := registerAndLoginTestUser(t, client, baseURL)

	sendReq := fmt.Sprintf(`{"message": "test", "receiver_ids": ["%s"]}`, receiver.User.ID.String())
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, baseURL+"/service/api/v1/messages/tx", bytes.NewReader([]byte(sendReq)))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sender.Token)

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var sendResp map[string]string

	err = json.NewDecoder(resp.Body).Decode(&sendResp)
	require.NoError(t, err)

	_ = resp.Body.Close()

	messageID := sendResp["message_id"]

	sqlDB, err := db.DB()
	require.NoError(t, err)

	_ = sqlDB.Close()

	req, err = http.NewRequestWithContext(context.Background(), http.MethodDelete, baseURL+"/service/api/v1/messages/"+messageID, http.NoBody)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+sender.Token)

	resp, err = client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}
*/

// TestHandleReceiveMessages_EmptyInbox tests receiving when no messages exist.
func TestHandleReceiveMessages_EmptyInbox(t *testing.T) {
	// NOTE: t.Parallel() removed due to cleanTestDB() data isolation issue
	// See: learn_test_isolation_issue.txt

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	receiver := registerAndLoginTestUser(t, client, baseURL)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, baseURL+"/service/api/v1/messages/rx", http.NoBody)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+receiver.Token)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody map[string][]any

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.Empty(t, respBody["messages"])
}

// TestHandleDeleteMessage_InvalidMessageID tests delete with invalid UUID.
func TestHandleDeleteMessage_InvalidMessageID(t *testing.T) {
	// NOTE: t.Parallel() removed due to cleanTestDB() data isolation issue
	// See: learn_test_isolation_issue.txt

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	sender := registerAndLoginTestUser(t, client, baseURL)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, baseURL+"/service/api/v1/messages/invalid-uuid", http.NoBody)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+sender.Token)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestHandleReceiveMessages_WithMessages tests successfully retrieving messages.
func TestHandleReceiveMessages_WithMessages(t *testing.T) {
	// NOTE: t.Parallel() removed due to cleanTestDB() data isolation issue
	// See: learn_test_isolation_issue.txt
	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	sender := registerAndLoginTestUser(t, client, baseURL)
	receiver := registerAndLoginTestUser(t, client, baseURL)

	reqBody := map[string]any{
		"message":      "Hello receiver",
		"receiver_ids": []string{receiver.User.ID.String()},
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

	req, err = http.NewRequestWithContext(context.Background(), http.MethodGet, baseURL+"/service/api/v1/messages/rx", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+receiver.Token)

	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	defer func() { _ = resp.Body.Close() }()

	var rxResp map[string][]map[string]any

	err = json.NewDecoder(resp.Body).Decode(&rxResp)
	require.NoError(t, err)
	require.Len(t, rxResp["messages"], 1)

	msg := rxResp["messages"][0]
	require.NotEmpty(t, msg["message_id"], "message_id should not be empty")
	require.NotEmpty(t, msg["sender_pub_key"], "sender_pub_key should not be empty")
	require.NotEmpty(t, msg["encrypted_content"], "encrypted_content should not be empty")
	// NOTE: nonce not checked - JWE Compact embeds nonce in format, not extractable separately
	require.NotEmpty(t, msg["created_at"], "created_at should not be empty")
}

// TestHandleReceiveMessages_MultipleMessages tests receiving multiple messages.
func TestHandleReceiveMessages_MultipleMessages(t *testing.T) {
	// NOTE: t.Parallel() removed due to cleanTestDB() data isolation issue
	// See: learn_test_isolation_issue.txt
	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	sender := registerAndLoginTestUser(t, client, baseURL)
	receiver := registerAndLoginTestUser(t, client, baseURL)

	for i := 1; i <= 3; i++ {
		reqBody := map[string]any{
			"message":      fmt.Sprintf("Message %d", i),
			"receiver_ids": []string{receiver.User.ID.String()},
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
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, baseURL+"/service/api/v1/messages/rx", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+receiver.Token)

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	defer func() { _ = resp.Body.Close() }()

	var rxResp map[string][]map[string]any

	err = json.NewDecoder(resp.Body).Decode(&rxResp)
	require.NoError(t, err)
	require.Len(t, rxResp["messages"], 3)

	for _, msg := range rxResp["messages"] {
		require.NotEmpty(t, msg["message_id"])
		require.NotEmpty(t, msg["sender_pub_key"])
		require.NotEmpty(t, msg["encrypted_content"])
		// NOTE: nonce not checked - JWE Compact embeds nonce in format, not extractable separately
		require.NotEmpty(t, msg["created_at"])
	}
}

// TestHandleDeleteMessage_EmptyID tests delete message with empty message ID.
func TestHandleDeleteMessage_EmptyID(t *testing.T) {
	// NOTE: t.Parallel() removed due to cleanTestDB() data isolation issue
	// See: learn_test_isolation_issue.txt

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	user := registerAndLoginTestUser(t, client, baseURL)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, baseURL+"/service/api/v1/messages/", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+user.Token)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestHandleReceiveMessages_MessageReceiverNotFound tests when receiver entry not found in message.
func TestHandleReceiveMessages_MessageReceiverNotFound(t *testing.T) {
	// NOTE: t.Parallel() removed due to cleanTestDB() data isolation issue
	// See: learn_test_isolation_issue.txt
	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	sender := registerAndLoginTestUser(t, client, baseURL)
	receiver1 := registerAndLoginTestUser(t, client, baseURL)

	sendReqBody := map[string]any{
		"receiver_ids": []string{receiver1.User.ID.String()},
		"message":      "Test message",
	}
	sendReqJSON, err := json.Marshal(sendReqBody)
	require.NoError(t, err)

	sendReq, err := http.NewRequestWithContext(context.Background(), http.MethodPut, baseURL+"/service/api/v1/messages/tx", bytes.NewReader(sendReqJSON))
	require.NoError(t, err)
	sendReq.Header.Set("Content-Type", "application/json")
	sendReq.Header.Set("Authorization", "Bearer "+sender.Token)

	sendResp, err := client.Do(sendReq)
	require.NoError(t, err)

	defer func() { _ = sendResp.Body.Close() }()

	require.Equal(t, http.StatusCreated, sendResp.StatusCode)

	recvReq, err := http.NewRequestWithContext(context.Background(), http.MethodGet, baseURL+"/service/api/v1/messages/rx", nil)
	require.NoError(t, err)
	recvReq.Header.Set("Authorization", "Bearer "+receiver1.Token)

	recvResp, err := client.Do(recvReq)
	require.NoError(t, err)

	defer func() { _ = recvResp.Body.Close() }()

	require.Equal(t, http.StatusOK, recvResp.StatusCode)

	var recvResult map[string]any

	err = json.NewDecoder(recvResp.Body).Decode(&recvResult)
	require.NoError(t, err)

	messages, ok := recvResult["messages"].([]any)
	require.True(t, ok)
	require.Len(t, messages, 1)
}

// TestHandleReceiveMessages_RepositoryError tests receiving messages when repository fails.
// DISABLED: Cannot close shared database in TestMain pattern.
// This test needs to use a mock repository instead of closing the real database.
// Closing shared database breaks all parallel tests that run after this test.
/*
func TestHandleReceiveMessages_RepositoryError(t *testing.T) {
	// NOTE: t.Parallel() removed due to cleanTestDB() data isolation issue
	// See: learn_test_isolation_issue.txt

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	receiver := registerAndLoginTestUser(t, client, baseURL)

	sqlDB, err := db.DB()
	require.NoError(t, err)

	_ = sqlDB.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, baseURL+"/service/api/v1/messages/rx", http.NoBody)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+receiver.Token)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
*/
