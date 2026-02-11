// Copyright (c) 2025 Justin Cranford
//

package integration

import (
	"testing"

	cryptoutilAppsCipherImClient "cryptoutil/internal/apps/cipher/im/client"
	cryptoutilAppsTemplateServiceTestingE2e "cryptoutil/internal/apps/template/service/testing/e2e"

	"github.com/stretchr/testify/require"
)

// TestE2E_FullEncryptionFlow tests the complete encryption workflow via /service/** paths.
// Phase 5a: Server-side decryption using JWE Compact Serialization.
func TestE2E_FullEncryptionFlow(t *testing.T) {
	t.Parallel()

	// Register users.
	user1 := cryptoutilAppsTemplateServiceTestingE2e.RegisterTestUserService(t, sharedHTTPClient, publicBaseURL)
	user2 := cryptoutilAppsTemplateServiceTestingE2e.RegisterTestUserService(t, sharedHTTPClient, publicBaseURL)

	// user1 sends encrypted message to user2.
	plaintext := "Hello " + user2.Username + ", this is a secret message from " + user1.Username + "!"

	messageID, err := cryptoutilAppsCipherImClient.SendMessage(sharedHTTPClient, publicBaseURL, plaintext, user1.Token, user2.ID)
	require.NoError(t, err)
	require.NotEmpty(t, messageID, "message ID should not be empty")

	// user2 receives messages.
	messages, err := cryptoutilAppsCipherImClient.ReceiveMessagesService(sharedHTTPClient, publicBaseURL, user2.Token)
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

// TestE2E_MultiReceiverEncryption tests sending to multiple recipients.
// Phase 5a: Each recipient gets their own JWK copy in messages_recipient_jwks table.
func TestE2E_MultiReceiverEncryption(t *testing.T) {
	t.Parallel()

	// Register users.
	user1 := cryptoutilAppsTemplateServiceTestingE2e.RegisterTestUserService(t, sharedHTTPClient, publicBaseURL)
	user2 := cryptoutilAppsTemplateServiceTestingE2e.RegisterTestUserService(t, sharedHTTPClient, publicBaseURL)
	user3 := cryptoutilAppsTemplateServiceTestingE2e.RegisterTestUserService(t, sharedHTTPClient, publicBaseURL)

	// user1 sends message to both user2 and user3.
	plaintext := "Hello to both of you!"

	messageID, err := cryptoutilAppsCipherImClient.SendMessage(sharedHTTPClient, publicBaseURL, plaintext, user1.Token, user2.ID, user3.ID)
	require.NoError(t, err)
	require.NotEmpty(t, messageID, "message ID should not be empty")

	// user2 receives messages.
	user2Messages, err := cryptoutilAppsCipherImClient.ReceiveMessagesService(sharedHTTPClient, publicBaseURL, user2.Token)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(user2Messages), 1, "user2 should have at least 1 message")

	// Verify user2 can decrypt.
	user2Msg := user2Messages[0]
	user2Decrypted, ok := user2Msg["encrypted_content"].(string)
	require.True(t, ok)
	require.Equal(t, plaintext, user2Decrypted)

	// user3 receives messages.
	user3Messages, err := cryptoutilAppsCipherImClient.ReceiveMessagesService(sharedHTTPClient, publicBaseURL, user3.Token)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(user3Messages), 1, "user3 should have at least 1 message")

	// Verify user3 can decrypt.
	user3Msg := user3Messages[0]
	user3Decrypted, ok := user3Msg["encrypted_content"].(string)
	require.True(t, ok)
	require.Equal(t, plaintext, user3Decrypted)
}

// TestE2E_MessageDeletion tests deleting messages.
// Phase 5a: CASCADE DELETE recipient JWKs when message deleted.
func TestE2E_MessageDeletion(t *testing.T) {
	t.Parallel()

	// Register users.
	user1 := cryptoutilAppsTemplateServiceTestingE2e.RegisterTestUserService(t, sharedHTTPClient, publicBaseURL)
	user2 := cryptoutilAppsTemplateServiceTestingE2e.RegisterTestUserService(t, sharedHTTPClient, publicBaseURL)

	// user1 sends message to user2.
	plaintext := "This message will be deleted!"

	messageID, err := cryptoutilAppsCipherImClient.SendMessage(sharedHTTPClient, publicBaseURL, plaintext, user1.Token, user2.ID)
	require.NoError(t, err)
	require.NotEmpty(t, messageID, "message ID should not be empty")

	// user2 receives messages.
	messages, err := cryptoutilAppsCipherImClient.ReceiveMessagesService(sharedHTTPClient, publicBaseURL, user2.Token)
	require.NoError(t, err)
	require.Len(t, messages, 1)

	// user1 deletes the message.
	err = cryptoutilAppsCipherImClient.DeleteMessageService(sharedHTTPClient, publicBaseURL, messageID, user1.Token)
	require.NoError(t, err)

	// user2 receives messages again (should be empty).
	messages, err = cryptoutilAppsCipherImClient.ReceiveMessagesService(sharedHTTPClient, publicBaseURL, user2.Token)
	require.NoError(t, err)
	require.Len(t, messages, 0, "user2 should have no messages after deletion")
}
