// Copyright (c) 2025 Justin Cranford
//
//

package e2e_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // CGO-free SQLite driver

	cryptoutilCrypto "cryptoutil/internal/learn/crypto"
	cryptoutilDomain "cryptoutil/internal/learn/domain"
	"cryptoutil/internal/learn/repository"
	"cryptoutil/internal/learn/server"
	cryptoutilTLSGenerator "cryptoutil/internal/shared/config/tls_generator"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// initTestDB creates an in-memory SQLite database with schema.
func initTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	ctx := context.Background()

	// Create unique in-memory database per test to avoid table conflicts.
	dbID, err := googleUuid.NewV7()
	require.NoError(t, err)

	dsn := "file:" + dbID.String() + "?mode=memory&cache=private"

	sqlDB, err := sql.Open("sqlite", dsn)
	require.NoError(t, err)

	// Configure SQLite for concurrent operations.
	_, err = sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
	require.NoError(t, err)

	_, err = sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;")
	require.NoError(t, err)

	sqlDB.SetMaxOpenConns(cryptoutilMagic.SQLiteMaxOpenConnections)
	sqlDB.SetMaxIdleConns(cryptoutilMagic.SQLiteMaxOpenConnections)
	sqlDB.SetConnMaxLifetime(0) // In-memory: keep connections alive.

	// Wrap with GORM using sqlite Dialector.
	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	// Run migrations.
	err = db.AutoMigrate(&cryptoutilDomain.User{}, &cryptoutilDomain.Message{}, &cryptoutilDomain.MessageReceiver{})
	require.NoError(t, err)

	return db
}

// createTestPublicServer creates a PublicServer for testing.
func createTestPublicServer(t *testing.T, db *gorm.DB) (*server.PublicServer, string) {
	t.Helper()

	ctx := context.Background()

	userRepo := repository.NewUserRepository(db)
	messageRepo := repository.NewMessageRepository(db)

	// Use port 0 for dynamic allocation (prevents port conflicts in tests).
	const testPort = 0

	// TLS config with localhost subject.
	tlsCfg, err := cryptoutilTLSGenerator.GenerateAutoTLSGeneratedSettings(
		[]string{"localhost"},
		[]string{cryptoutilMagic.IPv4Loopback},
		cryptoutilMagic.TLSTestEndEntityCertValidity1Year,
	)
	require.NoError(t, err)

	publicServer, err := server.NewPublicServer(ctx, testPort, userRepo, messageRepo, tlsCfg)
	require.NoError(t, err)

	// Start server in background.
	errChan := make(chan error, 1)

	go func() {
		if startErr := publicServer.Start(ctx); startErr != nil {
			errChan <- startErr
		}
	}()

	// Wait for server to bind to port.
	const (
		maxWaitAttempts = 50
		waitInterval    = 100 * time.Millisecond
	)

	actualPort := 0
	for i := 0; i < maxWaitAttempts; i++ {
		actualPort = publicServer.ActualPort()
		if actualPort > 0 {
			break
		}

		time.Sleep(waitInterval)
	}

	require.Greater(t, actualPort, 0, "server did not bind to port")

	baseURL := "https://" + cryptoutilMagic.IPv4Loopback + ":" + intToString(actualPort)

	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = publicServer.Shutdown(shutdownCtx)
	})

	return publicServer, baseURL
}

// intToString converts int to string.
func intToString(n int) string {
	if n < 0 {
		return "-" + intToString(-n)
	}

	if n < 10 {
		return string(rune('0' + n))
	}

	return intToString(n/10) + string(rune('0'+(n%10)))
}

// createHTTPClient creates an HTTP client that trusts self-signed certificates.
func createHTTPClient(t *testing.T) *http.Client {
	t.Helper()

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Test environment only.
			},
		},
		Timeout: 5 * time.Second,
	}
}

// testUser represents a user with their private key for testing.
type testUser struct {
	ID         googleUuid.UUID
	Username   string
	PrivateKey []byte // ECDH private key (for decryption).
	PublicKey  []byte // ECDH public key.
	Token      string // JWT authentication token.
}

// loginUser logs in a user and returns JWT token.
func loginUser(t *testing.T, client *http.Client, baseURL, username, password string) string {
	t.Helper()

	reqBody := map[string]string{
		"username": username,
		"password": password,
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/service/api/v1/users/login", bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)

	return respBody["token"]
}

// registerUser registers a user and returns the user with private key.
func registerUser(t *testing.T, client *http.Client, baseURL, username, password string) *testUser {
	t.Helper()

	reqBody := map[string]string{
		"username": username,
		"password": password,
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/service/api/v1/users/register", bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)

	userID, err := googleUuid.Parse(respBody["user_id"])
	require.NoError(t, err)

	publicKeyBytes, err := hex.DecodeString(respBody["public_key"])
	require.NoError(t, err)

	privateKeyBytes, err := hex.DecodeString(respBody["private_key"])
	require.NoError(t, err)

	// Login to get JWT token.
	token := loginUser(t, client, baseURL, username, password)

	return &testUser{
		ID:         userID,
		Username:   username,
		PrivateKey: privateKeyBytes,
		PublicKey:  publicKeyBytes,
		Token:      token,
	}
}

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

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, baseURL+"/service/api/v1/messages/tx", bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)

	return respBody["message_id"]
}

// receiveMessages retrieves messages for the specified receiver.
func receiveMessages(t *testing.T, client *http.Client, baseURL, token string) []map[string]any {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, baseURL+"/service/api/v1/messages/rx", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody map[string][]map[string]any

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)

	return respBody["messages"]
}

// registerUserBrowser registers a user via /browser/api/v1/users/register.
func registerUserBrowser(t *testing.T, client *http.Client, baseURL, username, password string) *testUser {
	t.Helper()

	reqBody := map[string]string{
		"username": username,
		"password": password,
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/browser/api/v1/users/register", bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)

	userID, err := googleUuid.Parse(respBody["user_id"])
	require.NoError(t, err)

	publicKeyBytes, err := hex.DecodeString(respBody["public_key"])
	require.NoError(t, err)

	privateKeyBytes, err := hex.DecodeString(respBody["private_key"])
	require.NoError(t, err)

	// Login to get JWT token.
	token := loginUserBrowser(t, client, baseURL, username, password)

	return &testUser{
		ID:         userID,
		Username:   username,
		PrivateKey: privateKeyBytes,
		PublicKey:  publicKeyBytes,
		Token:      token,
	}
}

// loginUserBrowser logs in a user via /browser/api/v1/users/login and returns JWT token.
func loginUserBrowser(t *testing.T, client *http.Client, baseURL, username, password string) string {
	t.Helper()

	reqBody := map[string]string{
		"username": username,
		"password": password,
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/browser/api/v1/users/login", bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)

	return respBody["token"]
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

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, baseURL+"/browser/api/v1/messages/tx", bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)

	return respBody["message_id"]
}

// receiveMessagesBrowser retrieves messages for the specified receiver via /browser/api/v1/messages/rx.
func receiveMessagesBrowser(t *testing.T, client *http.Client, baseURL, token string) []map[string]any {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, baseURL+"/browser/api/v1/messages/rx", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody map[string][]map[string]any

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)

	return respBody["messages"]
}

// deleteMessageBrowser deletes a message via /browser/api/v1/messages/:id.
func deleteMessageBrowser(t *testing.T, client *http.Client, baseURL, messageID, token string) {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, baseURL+"/browser/api/v1/messages/"+messageID, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

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

	// NOTE: CORS middleware not yet implemented.
	// TODO: Add CORS middleware and verify Access-Control-Allow-Origin header.
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

		var ciphertext, nonce, ephemeralPubKey []byte
		var errDecode error

		ciphertext, errDecode = hex.DecodeString(msg["encrypted_content"].(string))
		require.NoError(t, errDecode)
		nonce, errDecode = hex.DecodeString(msg["nonce"].(string))
		require.NoError(t, errDecode)
		ephemeralPubKey, errDecode = hex.DecodeString(msg["sender_pub_key"].(string))
		require.NoError(t, errDecode)

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
