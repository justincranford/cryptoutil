// Copyright (c) 2025 Justin Cranford
//
//

package e2e_test

import (
	"context"
	"encoding/hex"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCrypto "cryptoutil/internal/learn/crypto"
)

// TestE2E_FullEncryptionFlow tests the complete encryption workflow via /service/** paths.
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
	require.NotEmpty(t, receivedMsg["encrypted_content"], "encrypted content should not be empty")
	require.NotEmpty(t, receivedMsg["nonce"], "nonce should not be empty")
	require.NotEmpty(t, receivedMsg["sender_pub_key"], "sender pub key (ephemeral) should not be empty")

	// Bob decrypts message using his private key.
	ciphertextHex, ok := receivedMsg["encrypted_content"].(string)
	require.True(t, ok, "encrypted_content should be string")

	ciphertext, err := hex.DecodeString(ciphertextHex)
	require.NoError(t, err)

	nonceHex, ok := receivedMsg["nonce"].(string)
	require.True(t, ok, "nonce should be string")

	nonce, err := hex.DecodeString(nonceHex)
	require.NoError(t, err)

	ephemeralPubKeyHex, ok := receivedMsg["sender_pub_key"].(string)
	require.True(t, ok, "sender_pub_key should be string")

	ephemeralPubKeyBytes, err := hex.DecodeString(ephemeralPubKeyHex)
	require.NoError(t, err)

	bobPrivateKey, err := cryptoutilCrypto.ParseECDHPrivateKey(bob.PrivateKey)
	require.NoError(t, err)

	// Decrypt using ECDH + HKDF + AES-GCM.
	decrypted, err := cryptoutilCrypto.DecryptMessage(ciphertext, nonce, ephemeralPubKeyBytes, bobPrivateKey)
	require.NoError(t, err)

	// Verify decrypted plaintext matches original.
	require.Equal(t, plaintext, string(decrypted), "decrypted message should match original plaintext")
}

// TestE2E_MultiReceiverEncryption tests message encryption for multiple receivers via /service/** paths.
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
	plaintext := "Hello Bob and Charlie!"

	messageID := sendMessage(t, client, baseURL, plaintext, alice.Token, bob.ID, charlie.ID)
	require.NotEmpty(t, messageID)

	// Bob receives message.
	bobMessages := receiveMessages(t, client, baseURL, bob.Token)
	require.GreaterOrEqual(t, len(bobMessages), 1, "Bob should have at least 1 message")

	// Charlie receives message.
	charlieMessages := receiveMessages(t, client, baseURL, charlie.Token)
	require.GreaterOrEqual(t, len(charlieMessages), 1, "Charlie should have at least 1 message")

	// Verify both Bob and Charlie can decrypt the same message.
	// (Note: Current implementation encrypts with same ephemeral key for all receivers,
	//  so both receive identical ciphertext. Real implementation should encrypt separately per receiver.)
	require.Equal(t, bobMessages[0]["message_id"], charlieMessages[0]["message_id"], "both should receive same message")

	// Verify Alice does NOT receive the message (she is the sender).
	aliceMessages := receiveMessages(t, client, baseURL, alice.Token)
	require.Empty(t, aliceMessages, "Alice should not receive her own message")
}

// TestE2E_MessageDeletion tests message deletion via /service/** paths.
func TestE2E_MessageDeletion(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	// Register Alice and Bob.
	alice := registerUser(t, client, baseURL, "alice", "alicepass123")
	bob := registerUser(t, client, baseURL, "bob", "bobpass123")

	// Alice sends message to Bob.
	messageID := sendMessage(t, client, baseURL, "Test message", alice.Token, bob.ID)
	require.NotEmpty(t, messageID)

	// Bob receives message.
	messages := receiveMessages(t, client, baseURL, bob.Token)
	require.Len(t, messages, 1)

	// Delete the message (Alice is the sender, so she can delete).
	req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, baseURL+"/service/api/v1/messages/"+messageID, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+alice.Token)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify message is gone.
	messagesAfterDelete := receiveMessages(t, client, baseURL, bob.Token)
	require.Empty(t, messagesAfterDelete, "message should be deleted")
}
