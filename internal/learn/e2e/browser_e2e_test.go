// Copyright (c) 2025 Justin Cranford
//
//

package e2e_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestE2E_BrowserHealth tests the health endpoint via /browser/** paths.
func TestE2E_BrowserHealth(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, baseURL+"/browser/api/v1/health", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
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

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	// Register Alice and Bob.
	alice := registerUserBrowser(t, client, baseURL, "alice_browser", "alicepass123")
	bob := registerUserBrowser(t, client, baseURL, "bob_browser", "bobpass123")

	// Alice sends encrypted message to Bob.
	plaintext := "Browser encrypted message for Bob"
	messageID := sendMessageBrowser(t, client, baseURL, plaintext, alice.Token, bob.ID)
	require.NotEmpty(t, messageID)

	// Bob receives messages - server decrypts and returns plaintext.
	messages := receiveMessagesBrowser(t, client, baseURL, bob.Token)
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

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	// Register three users.
	alice := registerUserBrowser(t, client, baseURL, "alice_multi_browser", "alicepass123")
	bob := registerUserBrowser(t, client, baseURL, "bob_multi_browser", "bobpass123")
	charlie := registerUserBrowser(t, client, baseURL, "charlie_multi_browser", "charliepass123")

	// Alice sends message to both Bob and Charlie.
	plaintext := "Browser multi-receiver test"
	messageID := sendMessageBrowser(t, client, baseURL, plaintext, alice.Token, bob.ID, charlie.ID)
	require.NotEmpty(t, messageID)

	// Bob receives message.
	bobMessages := receiveMessagesBrowser(t, client, baseURL, bob.Token)
	require.Len(t, bobMessages, 1)

	// Charlie receives message.
	charlieMessages := receiveMessagesBrowser(t, client, baseURL, charlie.Token)
	require.Len(t, charlieMessages, 1)

	// Verify both received the same plaintext (server decrypted for each).
	bobDecrypted, ok := bobMessages[0]["encrypted_content"].(string)
	require.True(t, ok)
	require.Equal(t, plaintext, bobDecrypted)

	charlieDecrypted, ok := charlieMessages[0]["encrypted_content"].(string)
	require.True(t, ok)
	require.Equal(t, plaintext, charlieDecrypted)
}

// TestE2E_BrowserMessageDeletion tests message deletion via /browser/** paths.
// Phase 5a: Cascade delete removes recipient JWKs from messages_recipient_jwks table.
func TestE2E_BrowserMessageDeletion(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	// Register Alice and Bob.
	alice := registerUserBrowser(t, client, baseURL, "alice_delete_browser", "alicepass123")
	bob := registerUserBrowser(t, client, baseURL, "bob_delete_browser", "bobpass123")

	// Alice sends message to Bob.
	plaintext := "Message to be deleted via browser"
	messageID := sendMessageBrowser(t, client, baseURL, plaintext, alice.Token, bob.ID)
	require.NotEmpty(t, messageID)

	// Bob receives message.
	messages := receiveMessagesBrowser(t, client, baseURL, bob.Token)
	require.Len(t, messages, 1)

	// Alice deletes message (sender can delete).
	deleteMessageBrowser(t, client, baseURL, messageID, alice.Token)

	// Bob receives messages again - should be empty.
	messagesAfterDelete := receiveMessagesBrowser(t, client, baseURL, bob.Token)
	require.Len(t, messagesAfterDelete, 0)
}
