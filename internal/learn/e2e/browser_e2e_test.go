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

// TestE2E_BrowserHealth tests the browser health endpoint.
func TestE2E_BrowserHealth(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, baseURL+"/browser/api/v1/health", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	// NOTE: CORS middleware not yet implemented (Phase 8+).
	// Phase 8+ will add CORS middleware and verify Access-Control-Allow-Origin header.
	// corsOrigin := resp.Header.Get("Access-Control-Allow-Origin")
	// require.NotEmpty(t, corsOrigin, "browser path should include CORS headers")
}

// TestE2E_BrowserFullEncryptionFlow tests the complete encryption workflow via /browser/** paths.
func TestE2E_BrowserFullEncryptionFlow(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	// Register two users via browser path.
	alice := registerUserBrowser(t, client, baseURL, "alice_browser", "alicepass123")
	bob := registerUserBrowser(t, client, baseURL, "bob_browser", "bobpass123")

	require.NotEqual(t, alice.ID, bob.ID)
	require.NotEmpty(t, alice.Token)
	require.NotEmpty(t, bob.Token)

	// Alice sends encrypted message to Bob.
	plaintext := "Browser E2E test message"
	messageID := sendMessageBrowser(t, client, baseURL, plaintext, alice.Token, bob.ID)
	require.NotEmpty(t, messageID)

	// Bob receives and decrypts message.
	messages := receiveMessagesBrowser(t, client, baseURL, bob.Token)
	require.Len(t, messages, 1)

	// Extract encryption data from message.
	msg := messages[0]
	ciphertextHex, ok := msg["encrypted_content"].(string)
	require.True(t, ok, "encrypted_content should be string")

	nonceHex, ok := msg["nonce"].(string)
	require.True(t, ok, "nonce should be string")

	ephemeralPubKeyHex, ok := msg["sender_pub_key"].(string)
	require.True(t, ok, "sender_pub_key should be string")

	ciphertext, err := hex.DecodeString(ciphertextHex)
	require.NoError(t, err)

	nonce, err := hex.DecodeString(nonceHex)
	require.NoError(t, err)

	ephemeralPubKey, err := hex.DecodeString(ephemeralPubKeyHex)
	require.NoError(t, err)

	// Parse Bob's private key.
	bobPrivateKey, err := cryptoutilCrypto.ParseECDHPrivateKey(bob.PrivateKey)
	require.NoError(t, err)

	// Decrypt and verify.
	decrypted, err := cryptoutilCrypto.DecryptMessage(ciphertext, nonce, ephemeralPubKey, bobPrivateKey)
	require.NoError(t, err)
	require.Equal(t, plaintext, string(decrypted))
}

// TestE2E_BrowserMultiReceiverEncryption tests message encryption for multiple receivers via /browser/** paths.
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

	// Verify both can decrypt the same message.
	for _, recipientData := range []struct {
		name       string
		messages   []map[string]any
		privateKey []byte
	}{
		{"Bob", bobMessages, bob.PrivateKey},
		{"Charlie", charlieMessages, charlie.PrivateKey},
	} {
		msg := recipientData.messages[0]

		ciphertextHex, ok := msg["encrypted_content"].(string)
		require.True(t, ok)

		ciphertext, err := hex.DecodeString(ciphertextHex)
		require.NoError(t, err)

		nonceHex, ok := msg["nonce"].(string)
		require.True(t, ok)

		nonce, err := hex.DecodeString(nonceHex)
		require.NoError(t, err)

		ephemeralPubKeyHex, ok := msg["sender_pub_key"].(string)
		require.True(t, ok)

		ephemeralPubKey, err := hex.DecodeString(ephemeralPubKeyHex)
		require.NoError(t, err)

		// Parse recipient's private key.
		privateKey, err := cryptoutilCrypto.ParseECDHPrivateKey(recipientData.privateKey)
		require.NoError(t, err, "%s private key should parse", recipientData.name)

		decrypted, err := cryptoutilCrypto.DecryptMessage(ciphertext, nonce, ephemeralPubKey, privateKey)
		require.NoError(t, err, "%s should decrypt successfully", recipientData.name)
		require.Equal(t, plaintext, string(decrypted), "%s should see correct plaintext", recipientData.name)
	}
}

// TestE2E_BrowserMessageDeletion tests message deletion via /browser/** paths.
func TestE2E_BrowserMessageDeletion(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	// Register two users.
	alice := registerUserBrowser(t, client, baseURL, "alice_delete_browser", "alicepass123")
	bob := registerUserBrowser(t, client, baseURL, "bob_delete_browser", "bobpass123")

	// Alice sends message to Bob.
	plaintext := "Browser delete test message"
	messageID := sendMessageBrowser(t, client, baseURL, plaintext, alice.Token, bob.ID)
	require.NotEmpty(t, messageID)

	// Bob receives message.
	messages := receiveMessagesBrowser(t, client, baseURL, bob.Token)
	require.Len(t, messages, 1)

	msgIDFromResponse, ok := messages[0]["message_id"].(string)
	require.True(t, ok, "message_id should be string")

	// Alice (sender) deletes message.
	deleteMessageBrowser(t, client, baseURL, msgIDFromResponse, alice.Token)

	// Verify message deleted (Bob should no longer see it).
	remainingMessages := receiveMessagesBrowser(t, client, baseURL, bob.Token)
	require.Len(t, remainingMessages, 0, "message should be deleted")
}
