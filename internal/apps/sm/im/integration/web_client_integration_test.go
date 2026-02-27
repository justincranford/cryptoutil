//go:build integration
// +build integration

// Copyright (c) 2025 Justin Cranford
//

package integration

import (
	"testing"

	cryptoutilAppsSmImClient "cryptoutil/internal/apps/sm/im/client"
	cryptoutilAppsTemplateServiceTestingE2eHelpers "cryptoutil/internal/apps/template/service/testing/e2e_helpers"

	"github.com/stretchr/testify/require"
)

// TestE2E_BrowserFullEncryptionFlow tests the complete encryption workflow via /browser/** paths.
// Phase 5a: Server-side decryption using JWE Compact Serialization for browser clients.
func TestE2E_BrowserFullEncryptionFlow(t *testing.T) {
	t.Parallel()

	// Register users via browser endpoints.
	user1 := cryptoutilAppsTemplateServiceTestingE2eHelpers.RegisterTestUserBrowser(t, sharedHTTPClient, publicBaseURL)
	user2 := cryptoutilAppsTemplateServiceTestingE2eHelpers.RegisterTestUserBrowser(t, sharedHTTPClient, publicBaseURL)

	// user1 sends encrypted message to user2 via browser endpoint.
	plaintext := "Hello " + user2.Username + ", this is a browser message from " + user1.Username + "!"

	messageID, err := cryptoutilAppsSmImClient.SendMessageBrowser(sharedHTTPClient, publicBaseURL, plaintext, user1.Token, user2.ID)
	require.NoError(t, err)
	require.NotEmpty(t, messageID, "message ID should not be empty")

	// user2 receives messages via browser endpoint.
	messages, err := cryptoutilAppsSmImClient.ReceiveMessagesBrowser(sharedHTTPClient, publicBaseURL, user2.Token)
	require.NoError(t, err)
	require.Len(t, messages, 1, "%s should have 1 message", user2.Username)

	receivedMsg := messages[0]
	require.NotEmpty(t, receivedMsg["message_id"], "message ID should not be empty")

	// Phase 5a: Server decrypts using JWE Compact, returns plaintext.
	// encrypted_content field contains decrypted plaintext (not ciphertext).
	decryptedContent, ok := receivedMsg["encrypted_content"].(string)
	require.True(t, ok, "encrypted_content should be string")

	// Verify decrypted message matches original plaintext.
	require.Equal(t, plaintext, decryptedContent, "decrypted message should match original")
}

// TestE2E_BrowserMultiReceiverEncryption tests sending to multiple recipients via browser paths.
// Phase 5a: Each recipient gets their own JWK copy in messages_recipient_jwks table.
func TestE2E_BrowserMultiReceiverEncryption(t *testing.T) {
	t.Parallel()

	// Register users via browser endpoints.
	user1 := cryptoutilAppsTemplateServiceTestingE2eHelpers.RegisterTestUserBrowser(t, sharedHTTPClient, publicBaseURL)
	user2 := cryptoutilAppsTemplateServiceTestingE2eHelpers.RegisterTestUserBrowser(t, sharedHTTPClient, publicBaseURL)
	user3 := cryptoutilAppsTemplateServiceTestingE2eHelpers.RegisterTestUserBrowser(t, sharedHTTPClient, publicBaseURL)

	// user1 sends to both user2 and user3 via browser endpoint.
	plaintext := "Group message from " + user1.Username + " to multiple recipients!"
	messageID, err := cryptoutilAppsSmImClient.SendMessageBrowser(sharedHTTPClient, publicBaseURL, plaintext, user1.Token, user2.ID, user3.ID)
	require.NoError(t, err)
	require.NotEmpty(t, messageID, "message ID should not be empty")

	// user2 receives message.
	messages2, err := cryptoutilAppsSmImClient.ReceiveMessagesBrowser(sharedHTTPClient, publicBaseURL, user2.Token)
	require.NoError(t, err)
	require.Len(t, messages2, 1, "%s should have 1 message", user2.Username)

	decrypted2, ok := messages2[0]["encrypted_content"].(string)
	require.True(t, ok, "encrypted_content should be string")
	require.Equal(t, plaintext, decrypted2, "user2 message should match original")

	// user3 receives same message.
	messages3, err := cryptoutilAppsSmImClient.ReceiveMessagesBrowser(sharedHTTPClient, publicBaseURL, user3.Token)
	require.NoError(t, err)
	require.Len(t, messages3, 1, "%s should have 1 message", user3.Username)

	decrypted3, ok := messages3[0]["encrypted_content"].(string)
	require.True(t, ok, "encrypted_content should be string")
	require.Equal(t, plaintext, decrypted3, "user3 message should match original")
}

// TestE2E_BrowserMessageDeletion tests deleting messages via browser paths.
const testMessageDeletion = "This message will be deleted!"

func TestE2E_BrowserMessageDeletion(t *testing.T) {
	t.Parallel()

	// Register users via browser endpoints.
	user1 := cryptoutilAppsTemplateServiceTestingE2eHelpers.RegisterTestUserBrowser(t, sharedHTTPClient, publicBaseURL)
	user2 := cryptoutilAppsTemplateServiceTestingE2eHelpers.RegisterTestUserBrowser(t, sharedHTTPClient, publicBaseURL)

	// user1 sends message to user2 via browser endpoint.
	plaintext := testMessageDeletion
	messageID, err := cryptoutilAppsSmImClient.SendMessageBrowser(sharedHTTPClient, publicBaseURL, plaintext, user1.Token, user2.ID)
	require.NoError(t, err)
	require.NotEmpty(t, messageID, "message ID should not be empty")

	// user2 receives message.
	messagesBefore, err := cryptoutilAppsSmImClient.ReceiveMessagesBrowser(sharedHTTPClient, publicBaseURL, user2.Token)
	require.NoError(t, err)
	require.Len(t, messagesBefore, 1, "%s should have 1 message before deletion", user2.Username)

	// user1 (sender) deletes message via browser endpoint.
	err = cryptoutilAppsSmImClient.DeleteMessageBrowser(sharedHTTPClient, publicBaseURL, messageID, user1.Token)
	require.NoError(t, err)

	// user2 confirms message deleted.
	messagesAfter, err := cryptoutilAppsSmImClient.ReceiveMessagesBrowser(sharedHTTPClient, publicBaseURL, user2.Token)
	require.NoError(t, err)
	require.Len(t, messagesAfter, 0, "%s should have 0 messages after deletion", user2.Username)
}
