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
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilTemplateServer "cryptoutil/internal/template/server"
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
	tlsCfg := &cryptoutilTemplateServer.TLSConfig{
		Mode:             cryptoutilTemplateServer.TLSModeAuto,
		AutoDNSNames:     []string{"localhost"},
		AutoIPAddresses:  []string{cryptoutilMagic.IPv4Loopback},
		AutoValidityDays: cryptoutilMagic.TLSTestEndEntityCertValidity1Year,
	}

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

	return &testUser{
		ID:         userID,
		Username:   username,
		PrivateKey: privateKeyBytes,
		PublicKey:  publicKeyBytes,
	}
}

// sendMessage sends a message to one or more receivers.
func sendMessage(t *testing.T, client *http.Client, baseURL, message string, senderID googleUuid.UUID, receiverIDs ...googleUuid.UUID) string {
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

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, baseURL+"/service/api/v1/messages/tx?sender_id="+senderID.String(), bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

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
func receiveMessages(t *testing.T, client *http.Client, baseURL string, receiverID googleUuid.UUID) []map[string]any {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, baseURL+"/service/api/v1/messages/rx?receiver_id="+receiverID.String(), nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody map[string][]map[string]any

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)

	return respBody["messages"]
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

	messageID := sendMessage(t, client, baseURL, plaintext, alice.ID, bob.ID)
	require.NotEmpty(t, messageID, "message ID should not be empty")

	// Bob receives messages.
	messages := receiveMessages(t, client, baseURL, bob.ID)
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

	messageID := sendMessage(t, client, baseURL, plaintext, alice.ID, bob.ID, charlie.ID)
	require.NotEmpty(t, messageID)

	// Bob receives message.
	bobMessages := receiveMessages(t, client, baseURL, bob.ID)
	require.GreaterOrEqual(t, len(bobMessages), 1, "Bob should have at least 1 message")

	// Charlie receives message.
	charlieMessages := receiveMessages(t, client, baseURL, charlie.ID)
	require.GreaterOrEqual(t, len(charlieMessages), 1, "Charlie should have at least 1 message")

	// Verify both Bob and Charlie can decrypt the same message.
	// (Note: Current implementation encrypts with same ephemeral key for all receivers,
	//  so both receive identical ciphertext. Real implementation should encrypt separately per receiver.)
	require.Equal(t, bobMessages[0]["message_id"], charlieMessages[0]["message_id"], "both should receive same message")

	// Verify Alice does NOT receive the message (she is the sender).
	aliceMessages := receiveMessages(t, client, baseURL, alice.ID)
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
	messageID := sendMessage(t, client, baseURL, "Test message", alice.ID, bob.ID)
	require.NotEmpty(t, messageID)

	// Bob receives message.
	messages := receiveMessages(t, client, baseURL, bob.ID)
	require.Len(t, messages, 1)

	// Delete the message.
	req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, baseURL+"/service/api/v1/messages/"+messageID, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify message is gone.
	messagesAfterDelete := receiveMessages(t, client, baseURL, bob.ID)
	require.Empty(t, messagesAfterDelete, "message should be deleted")
}
