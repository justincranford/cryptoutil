// Copyright (c) 2025 Justin Cranford
//
//

package e2e_test_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestE2E_BrowserHealth tests the health endpoint via /browser/** paths.
func TestE2E_BrowserHealth(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, baseURL+"/browser/api/v1/health", nil)
	require.NoError(t, err)

	resp, err := sharedHTTPClient.Do(req)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestE2E_BrowserFullEncryptionFlow tests full encryption flow via /browser/** paths.
// Phase 5a: Server-side decryption using JWE Compact Serialization.
func TestE2E_BrowserFullEncryptionFlow(t *testing.T) {
	t.Parallel()

	// Register users.
	user1 := registerTestUserBrowser(t, sharedHTTPClient, baseURL)
	user2 := registerTestUserBrowser(t, sharedHTTPClient, baseURL)

	// user1 sends encrypted message to user2.
	plaintext := "Browser encrypted message for user2"
	messageID := sendMessageBrowser(t, sharedHTTPClient, baseURL, plaintext, user1.Token, user2.ID)
	require.NotEmpty(t, messageID)

	// user2 receives messages - server decrypts and returns plaintext.
	messages := receiveMessagesBrowser(t, sharedHTTPClient, baseURL, user2.Token)
	require.Len(t, messages, 1)

	receivedMsg := messages[0]

	// Phase 5a: encrypted_content contains plaintext (server already decrypted).
	decryptedContent, ok := receivedMsg["encrypted_content"].(string)
	require.True(t, ok, "encrypted_content should be string")
	require.Equal(t, plaintext, decryptedContent)
}

// TestE2E_BrowserMultiReceiverEncryption tests message encryption for multiple receivers via /browser/** paths.
// Phase 5a: Each recipient gets their own JWK copy for decryption.
func TestE2E_BrowserMultiReceiverEncryption(t *testing.T) {
	t.Parallel()

	// Register three users.
	user1 := registerTestUserBrowser(t, sharedHTTPClient, baseURL)
	user2 := registerTestUserBrowser(t, sharedHTTPClient, baseURL)
	user3 := registerTestUserBrowser(t, sharedHTTPClient, baseURL)

	// user1 sends message to both user2 and user3.
	plaintext := "Browser multi-receiver test"
	messageID := sendMessageBrowser(t, sharedHTTPClient, baseURL, plaintext, user1.Token, user2.ID, user3.ID)
	require.NotEmpty(t, messageID)

	// user2 receives message.
	user2Messages := receiveMessagesBrowser(t, sharedHTTPClient, baseURL, user2.Token)
	require.Len(t, user2Messages, 1)

	// user3 receives message.
	user3Messages := receiveMessagesBrowser(t, sharedHTTPClient, baseURL, user3.Token)
	require.Len(t, user3Messages, 1)
	// Verify both received the same plaintext (server decrypted for each).
	user2Decrypted, ok := user2Messages[0]["encrypted_content"].(string)
	require.True(t, ok)
	require.Equal(t, plaintext, user2Decrypted)

	user3Decrypted, ok := user3Messages[0]["encrypted_content"].(string)
	require.True(t, ok)
	require.Equal(t, plaintext, user3Decrypted)
}

// TestE2E_BrowserMessageDeletion tests message deletion via /browser/** paths.
// Phase 5a: Cascade delete removes recipient JWKs from messages_recipient_jwks table.
func TestE2E_BrowserMessageDeletion(t *testing.T) {
	t.Parallel()

	// Register users.
	user1 := registerTestUserBrowser(t, sharedHTTPClient, baseURL)
	user2 := registerTestUserBrowser(t, sharedHTTPClient, baseURL)

	// user1 sends message to user2.
	plaintext := "Message to be deleted via browser"
	messageID := sendMessageBrowser(t, sharedHTTPClient, baseURL, plaintext, user1.Token, user2.ID)
	require.NotEmpty(t, messageID)

	// user2 receives message.
	messages := receiveMessagesBrowser(t, sharedHTTPClient, baseURL, user2.Token)
	require.Len(t, messages, 1)

	// user1 deletes message (sender can delete).
	deleteMessageBrowser(t, sharedHTTPClient, baseURL, messageID, user1.Token)
	// user2 receives messages again - should be empty.
	messagesAfterDelete := receiveMessagesBrowser(t, sharedHTTPClient, baseURL, user2.Token)
	require.Len(t, messagesAfterDelete, 0)
}
