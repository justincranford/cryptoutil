// Copyright (c) 2025 Justin Cranford
//
//

package e2e_test

import (
	"encoding/json"
	"net/http"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilE2E "cryptoutil/internal/template/testing/e2e"
)

// sendMessage sends a message to one or more receivers.
func sendMessage(t *testing.T, client *http.Client, baseURL, message, token string, receiverIDs ...googleUuid.UUID) string {
	t.Helper()

	receiverIDStrs := make([]string, len(receiverIDs))
	for i, id := range receiverIDs {
		receiverIDStrs[i] = id.String()
	}

	reqBody := map[string]any{
		"message":      message,
		"receiver_ids": receiverIDStrs,
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	resp := cryptoutilE2E.SendAuthenticatedRequest(t, client, http.MethodPut, baseURL+"/service/api/v1/messages/tx", token, reqJSON)

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var respBody map[string]string

	cryptoutilE2E.DecodeJSONResponse(t, resp, &respBody)

	return respBody["message_id"]
}

// receiveMessagesService retrieves messages for the specified receiver.
func receiveMessagesService(t *testing.T, client *http.Client, baseURL, token string) []map[string]any {
	t.Helper()

	resp := cryptoutilE2E.SendAuthenticatedRequest(t, client, http.MethodGet, baseURL+"/service/api/v1/messages/rx", token, nil)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody map[string][]map[string]any

	cryptoutilE2E.DecodeJSONResponse(t, resp, &respBody)

	return respBody["messages"]
}

// deleteMessageService deletes a message via /service/api/v1/messages/:id.
func deleteMessageService(t *testing.T, client *http.Client, baseURL, messageID, token string) {
	t.Helper()

	resp := cryptoutilE2E.SendAuthenticatedRequest(t, client, http.MethodDelete, baseURL+"/service/api/v1/messages/"+messageID, token, nil)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

// sendMessageBrowser sends a message to one or more receivers via /browser/api/v1/messages/tx.
func sendMessageBrowser(t *testing.T, client *http.Client, baseURL, message, token string, receiverIDs ...googleUuid.UUID) string {
	t.Helper()

	receiverIDStrs := make([]string, len(receiverIDs))
	for i, id := range receiverIDs {
		receiverIDStrs[i] = id.String()
	}

	reqBody := map[string]any{
		"message":      message,
		"receiver_ids": receiverIDStrs,
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	resp := cryptoutilE2E.SendAuthenticatedRequest(t, client, http.MethodPut, baseURL+"/browser/api/v1/messages/tx", token, reqJSON)

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var respBody map[string]string

	cryptoutilE2E.DecodeJSONResponse(t, resp, &respBody)

	return respBody["message_id"]
}

// receiveMessagesBrowser retrieves messages for the specified receiver via /browser/api/v1/messages/rx.
func receiveMessagesBrowser(t *testing.T, client *http.Client, baseURL, token string) []map[string]any {
	t.Helper()

	resp := cryptoutilE2E.SendAuthenticatedRequest(t, client, http.MethodGet, baseURL+"/browser/api/v1/messages/rx", token, nil)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody map[string][]map[string]any

	cryptoutilE2E.DecodeJSONResponse(t, resp, &respBody)

	return respBody["messages"]
}

// deleteMessageBrowser deletes a message via /browser/api/v1/messages/:id.
func deleteMessageBrowser(t *testing.T, client *http.Client, baseURL, messageID, token string) {
	t.Helper()

	resp := cryptoutilE2E.SendAuthenticatedRequest(t, client, http.MethodDelete, baseURL+"/browser/api/v1/messages/"+messageID, token, nil)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func registerTestUserService(t *testing.T, client *http.Client, baseURL string) *cryptoutilE2E.TestUser {
	t.Helper()

	return cryptoutilE2E.RegisterTestUserService(t, client, baseURL)
}

func registerTestUserBrowser(t *testing.T, client *http.Client, baseURL string) *cryptoutilE2E.TestUser {
	t.Helper()

	return cryptoutilE2E.RegisterTestUserBrowser(t, client, baseURL)
}
