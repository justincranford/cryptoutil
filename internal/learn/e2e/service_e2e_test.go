// Copyright (c) 2025 Justin Cranford
//

package e2e_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestE2E_FullEncryptionFlow tests the complete encryption workflow via /service/** paths.
// Phase 5a: Server-side decryption using JWE Compact Serialization.
func TestE2E_FullEncryptionFlow(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	// Register Alice and Bob.
	alice := registerUser(t, client, baseURL, "alice", "alicepass123")
	bob := registerUser(t, client, baseURL, "bob", "bobpass123")

	// Alice sends encrypted message to Bob.
	plaintext := "Hello Bob, this is a secret message from Alice!"

	messageID := sendMessage(t, client, baseURL, plaintext, alice.Token, bob.ID)
	require.NotEmpty(t, messageID, "message ID should not be empty")

	// Bob receives messages.
	messages := receiveMessages(t, client, baseURL, bob.Token)
	require.Len(t, messages, 1, "Bob should have 1 message")

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

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	// Register Alice, Bob, and Charlie.
	alice := registerUser(t, client, baseURL, "alice", "alicepass123")
	bob := registerUser(t, client, baseURL, "bob", "bobpass123")
	charlie := registerUser(t, client, baseURL, "charlie", "charliepass123")

	// Alice sends message to both Bob and Charlie.
	plaintext := "Hello to both of you!"

	messageID := sendMessage(t, client, baseURL, plaintext, alice.Token, bob.ID, charlie.ID)
	require.NotEmpty(t, messageID, "message ID should not be empty")

	// Bob receives messages.
	bobMessages := receiveMessages(t, client, baseURL, bob.Token)
	require.GreaterOrEqual(t, len(bobMessages), 1, "Bob should have at least 1 message")

	// Verify Bob can decrypt.
	bobMsg := bobMessages[0]
	bobDecrypted, ok := bobMsg["encrypted_content"].(string)
	require.True(t, ok)
	require.Equal(t, plaintext, bobDecrypted)

	// Charlie receives messages.
	charlieMessages := receiveMessages(t, client, baseURL, charlie.Token)
	require.GreaterOrEqual(t, len(charlieMessages), 1, "Charlie should have at least 1 message")

	// Verify Charlie can decrypt.
	charlieMsg := charlieMessages[0]
	charlieDecrypted, ok := charlieMsg["encrypted_content"].(string)
	require.True(t, ok)
	require.Equal(t, plaintext, charlieDecrypted)
}

// TestE2E_MessageDeletion tests deleting messages.
// Phase 5a: CASCADE DELETE recipient JWKs when message deleted.
func TestE2E_MessageDeletion(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	// Register Alice and Bob.
	alice := registerUser(t, client, baseURL, "alice", "alicepass123")
	bob := registerUser(t, client, baseURL, "bob", "bobpass123")

	// Alice sends message to Bob.
	plaintext := "This message will be deleted!"

	messageID := sendMessage(t, client, baseURL, plaintext, alice.Token, bob.ID)
	require.NotEmpty(t, messageID, "message ID should not be empty")

	// Bob receives messages.
	messages := receiveMessages(t, client, baseURL, bob.Token)
	require.Len(t, messages, 1)

	// Alice deletes the message.
	deleteMessage(t, client, baseURL, messageID, alice.Token)

	// Bob receives messages again (should be empty).
	messages = receiveMessages(t, client, baseURL, bob.Token)
	require.Len(t, messages, 0, "Bob should have no messages after deletion")
}
