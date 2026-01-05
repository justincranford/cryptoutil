// Copyright (c) 2025 Justin Cranford
//
//

package e2e_test

import (
	"net/http"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cipherClient "cryptoutil/internal/cipher/client"
)

// sendMessage sends a message to one or more receivers.
func sendMessage(t *testing.T, sharedHTTPClient *http.Client, baseURL, message, token string, receiverIDs ...googleUuid.UUID) string {
	t.Helper()

	messageID, err := cipherClient.SendMessage(sharedHTTPClient, baseURL, message, token, receiverIDs...)
	require.NoError(t, err)
	return messageID
}

// receiveMessagesService retrieves messages for the specified receiver.
func receiveMessagesService(t *testing.T, sharedHTTPClient *http.Client, baseURL, token string) []map[string]any {
	t.Helper()

	messages, err := cipherClient.ReceiveMessagesService(sharedHTTPClient, baseURL, token)
	require.NoError(t, err)
	return messages
}

// deleteMessageService deletes a message via /service/api/v1/messages/:id.
func deleteMessageService(t *testing.T, sharedHTTPClient *http.Client, baseURL, messageID, token string) {
	t.Helper()

	err := cipherClient.DeleteMessageService(sharedHTTPClient, baseURL, messageID, token)
	require.NoError(t, err)
}

// sendMessageBrowser sends a message to one or more receivers via /browser/api/v1/messages/tx.
func sendMessageBrowser(t *testing.T, sharedHTTPClient *http.Client, baseURL, message, token string, receiverIDs ...googleUuid.UUID) string {
	t.Helper()

	messageID, err := cipherClient.SendMessageBrowser(sharedHTTPClient, baseURL, message, token, receiverIDs...)
	require.NoError(t, err)
	return messageID
}
