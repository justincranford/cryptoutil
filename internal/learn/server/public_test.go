// Copyright (c) 2025 Justin Cranford
//
//

package server_test

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

// testUserWithToken represents a test user with authentication token.
type testUserWithToken struct {
	User  *cryptoutilDomain.User
	Token string
}

// registerAndLoginTestUser registers a user and logs in to get JWT token.
func registerAndLoginTestUser(t *testing.T, client *http.Client, baseURL, username, password string) *testUserWithToken {
	t.Helper()

	// Register user.
	user := registerTestUser(t, client, baseURL, username, password)

	// Login to get token.
	loginReqBody := map[string]string{
		"username": username,
		"password": password,
	}
	loginReqJSON, err := json.Marshal(loginReqBody)
	require.NoError(t, err)

	loginReq, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/service/api/v1/users/login", bytes.NewReader(loginReqJSON))
	require.NoError(t, err)
	loginReq.Header.Set("Content-Type", "application/json")

	loginResp, err := client.Do(loginReq)
	require.NoError(t, err)

	defer func() { _ = loginResp.Body.Close() }()

	require.Equal(t, http.StatusOK, loginResp.StatusCode)

	var loginRespBody map[string]string

	err = json.NewDecoder(loginResp.Body).Decode(&loginRespBody)
	require.NoError(t, err)

	return &testUserWithToken{
		User:  user,
		Token: loginRespBody["token"],
	}
}

// registerTestUser is a helper that registers a user and returns the user domain object.
func registerTestUser(t *testing.T, client *http.Client, baseURL, username, password string) *cryptoutilDomain.User {
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

	pubKey, err := cryptoutilCrypto.ParseECDHPublicKey(publicKeyBytes)
	require.NoError(t, err)

	return &cryptoutilDomain.User{
		ID:        userID,
		Username:  username,
		PublicKey: pubKey.Bytes(),
	}
}

func TestHandleRegisterUser_Success(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	// Prepare request.
	reqBody := map[string]string{
		"username": "testuser",
		"password": "password123",
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	// Send POST /service/api/v1/users/register.
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/service/api/v1/users/register", bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	// Verify response.
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)

	require.NotEmpty(t, respBody["user_id"])
	require.NotEmpty(t, respBody["public_key"])

	// Verify public key is valid hex.
	publicKeyBytes, err := hex.DecodeString(respBody["public_key"])
	require.NoError(t, err)

	const expectedPublicKeyLength = 65 // X9.62 uncompressed format.
	require.Len(t, publicKeyBytes, expectedPublicKeyLength)
}

func TestHandleRegisterUser_UsernameTooShort(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	// Username too short (< 3 characters).
	reqBody := map[string]string{
		"username": "ab",
		"password": "password123",
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/service/api/v1/users/register", bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.Contains(t, respBody["error"], "username must be 3-50 characters")
}

func TestHandleRegisterUser_PasswordTooShort(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	// Password too short (< 8 characters).
	reqBody := map[string]string{
		"username": "testuser",
		"password": "pass",
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/service/api/v1/users/register", bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.Contains(t, respBody["error"], "password must be at least 8 characters")
}

func TestHandleRegisterUser_DuplicateUsername(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	userRepo := repository.NewUserRepository(db)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	// Create existing user.
	passwordHash, err := cryptoutilCrypto.HashPassword("password123")
	require.NoError(t, err)

	// Generate key pair for test user.
	privateKey, publicKeyBytes, err := cryptoutilCrypto.GenerateECDHKeyPair()
	require.NoError(t, err)

	existingUser := &cryptoutilDomain.User{
		ID:           googleUuid.New(),
		Username:     "existinguser",
		PasswordHash: hex.EncodeToString(passwordHash),
		PublicKey:    publicKeyBytes,
		PrivateKey:   privateKey.Bytes(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err = userRepo.Create(context.Background(), existingUser)
	require.NoError(t, err)

	// Attempt to register with same username.
	reqBody := map[string]string{
		"username": "existinguser",
		"password": "password123",
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/service/api/v1/users/register", bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusConflict, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.Contains(t, respBody["error"], "username already exists")
}

func TestHandleLoginUser_Success(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	userRepo := repository.NewUserRepository(db)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	// Create user.
	passwordHash, err := cryptoutilCrypto.HashPassword("password123")
	require.NoError(t, err)

	// Generate key pair for test user.
	privateKey, publicKeyBytes, err := cryptoutilCrypto.GenerateECDHKeyPair()
	require.NoError(t, err)

	user := &cryptoutilDomain.User{
		ID:           googleUuid.New(),
		Username:     "loginuser",
		PasswordHash: hex.EncodeToString(passwordHash),
		PublicKey:    publicKeyBytes,
		PrivateKey:   privateKey.Bytes(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err = userRepo.Create(context.Background(), user)
	require.NoError(t, err)

	// Login with correct credentials.
	reqBody := map[string]string{
		"username": "loginuser",
		"password": "password123",
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

	require.NotEmpty(t, respBody["token"], "JWT token should be returned")
	require.NotEmpty(t, respBody["expires_at"])
}

func TestHandleLoginUser_WrongPassword(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	userRepo := repository.NewUserRepository(db)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	// Create user.
	passwordHash, err := cryptoutilCrypto.HashPassword("password123")
	require.NoError(t, err)

	// Generate key pair for test user.
	privateKey, publicKeyBytes, err := cryptoutilCrypto.GenerateECDHKeyPair()
	require.NoError(t, err)

	user := &cryptoutilDomain.User{
		ID:           googleUuid.New(),
		Username:     "wrongpassuser",
		PasswordHash: hex.EncodeToString(passwordHash),
		PublicKey:    publicKeyBytes,
		PrivateKey:   privateKey.Bytes(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err = userRepo.Create(context.Background(), user)
	require.NoError(t, err)

	// Login with wrong password.
	reqBody := map[string]string{
		"username": "wrongpassuser",
		"password": "wrongpassword",
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/service/api/v1/users/login", bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.Contains(t, respBody["error"], "invalid credentials")
}

func TestHandleLoginUser_UserNotFound(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	// Login with non-existent user.
	reqBody := map[string]string{
		"username": "nonexistent",
		"password": "password123",
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/service/api/v1/users/login", bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.Contains(t, respBody["error"], "invalid credentials")
}

func TestHandleSendMessage_Success(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	// Create sender and receiver users with authentication.
	sender := registerAndLoginTestUser(t, client, baseURL, "sender", "password123")
	receiver := registerTestUser(t, client, baseURL, "receiver", "password123")

	// Send message.
	reqBody := map[string]any{
		"message":      "Hello, receiver!",
		"receiver_ids": []string{receiver.ID.String()},
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, baseURL+"/service/api/v1/messages/tx", bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sender.Token)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.NotEmpty(t, respBody["message_id"])
}

func TestHandleSendMessage_EmptyReceivers(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	// Create sender with authentication.
	sender := registerAndLoginTestUser(t, client, baseURL, "sender", "password123")

	// Send message with empty receivers.
	reqBody := map[string]any{
		"message":      "Hello!",
		"receiver_ids": []string{},
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, baseURL+"/service/api/v1/messages/tx", bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sender.Token)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.Contains(t, respBody["error"], "receiver_ids cannot be empty")
}

func TestHandleSendMessage_InvalidReceiverID(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	// Create sender with authentication.
	sender := registerAndLoginTestUser(t, client, baseURL, "sender", "password123")

	// Send message with invalid receiver ID.
	reqBody := map[string]any{
		"message":      "Hello!",
		"receiver_ids": []string{"invalid-uuid"},
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, baseURL+"/service/api/v1/messages/tx", bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sender.Token)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.Contains(t, respBody["error"], "invalid receiver ID")
}

func TestHandleReceiveMessages_Empty(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	// Create receiver with authentication.
	receiver := registerAndLoginTestUser(t, client, baseURL, "receiver", "password123")

	// Retrieve messages without sending any.
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, baseURL+"/service/api/v1/messages/rx", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+receiver.Token)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody map[string][]any

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.Empty(t, respBody["messages"])
}

func TestHandleDeleteMessage_Success(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	// Create receiver and send message.
	receiver := registerTestUser(t, client, baseURL, "receiver", "password123")

	reqBody := map[string]any{
		"message":      "Test message",
		"receiver_ids": []string{receiver.ID.String()},
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, baseURL+"/service/api/v1/messages/tx", bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	var sendResp map[string]string

	err = json.NewDecoder(resp.Body).Decode(&sendResp)
	require.NoError(t, err)

	_ = resp.Body.Close()

	messageID := sendResp["message_id"]

	// Get message from database to verify it exists.
	var message cryptoutilDomain.Message

	err = db.Where("id = ?", messageID).First(&message).Error
	require.NoError(t, err)

	// Delete the message.
	req, err = http.NewRequestWithContext(context.Background(), http.MethodDelete, baseURL+"/service/api/v1/messages/"+messageID, nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestHandleDeleteMessage_InvalidID(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	// Create sender with authentication.
	sender := registerAndLoginTestUser(t, client, baseURL, "sender", "password123")

	// Delete with invalid ID.
	req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, baseURL+"/service/api/v1/messages/invalid-id", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+sender.Token)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.Contains(t, respBody["error"], "invalid message ID")
}

func TestHandleDeleteMessage_NotFound(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	// Create sender with authentication.
	sender := registerAndLoginTestUser(t, client, baseURL, "sender", "password123")

	// Delete non-existent message.
	messageID := googleUuid.New()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, baseURL+"/service/api/v1/messages/"+messageID.String(), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+sender.Token)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusNotFound, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.Contains(t, respBody["error"], "message not found")
}
