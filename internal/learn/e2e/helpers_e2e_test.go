// Copyright (c) 2025 Justin Cranford
//
//

package e2e_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // CGO-free SQLite driver

	"cryptoutil/internal/learn/repository"
	"cryptoutil/internal/learn/server"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilRandom "cryptoutil/internal/shared/util/random"
)

// Wait for server to bind to port.
const (
	maxWaitAttempts = 50
	waitInterval    = 100 * time.Millisecond
	testPort        = 0 // Use port 0 for dynamic allocation (prevents port conflicts in tests).
)

var testJWTSecret = googleUuid.Must(googleUuid.NewV7()).String()

// testUser represents a user with their authentication token for testing.
type testUser struct {
	ID       googleUuid.UUID // UUIDv7
	Username string
	Password string
	Token    string // JWT authentication token.
}

// initTestDB creates an in-memory SQLite database with schema.
func initTestDB() (*gorm.DB, error) {
	ctx := context.Background()

	// Create unique in-memory database per test to avoid table conflicts.
	dbID, err := googleUuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate UUIDv7: %w", err)
	}

	dsn := "file:" + dbID.String() + "?mode=memory&cache=private"

	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite: %w", err)
	}

	// Configure SQLite for concurrent operations.
	_, err = sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
	if err != nil {
		return nil, fmt.Errorf("failed to enable WAL: %w", err)
	}

	_, err = sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;")
	if err != nil {
		return nil, fmt.Errorf("failed to set busy_timeout: %w", err)
	}

	sqlDB.SetMaxOpenConns(cryptoutilMagic.SQLiteMaxOpenConnections)
	sqlDB.SetMaxIdleConns(cryptoutilMagic.SQLiteMaxOpenConnections)
	sqlDB.SetConnMaxLifetime(0) // In-memory: keep connections alive.

	// Wrap with GORM using sqlite Dialector.
	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open GORM DB: %w", err)
	}

	// Run migrations using embedded migration files.
	err = repository.ApplyMigrations(sqlDB, repository.DatabaseTypeSQLite)
	if err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	return db, nil
}

// createTestPublicServer creates a PublicServer for testing using shared resources from TestMain.
func createTestPublicServer(db *gorm.DB) (*server.PublicServer, string, error) {
	ctx := context.Background()

	userRepo := repository.NewUserRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	messageRecipientJWKRepo := repository.NewMessageRecipientJWKRepository(db)

	// Create server using shared JWK generation service and TLS configuration.
	publicServer, err := server.NewPublicServer(ctx, testPort, userRepo, messageRepo, messageRecipientJWKRepo, sharedJWKGenService, testJWTSecret, sharedTLSConfig)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create PublicServer: %w", err)
	}

	// Start server in background.
	errChan := make(chan error, 1)

	go func() {
		if startErr := publicServer.Start(ctx); startErr != nil {
			errChan <- startErr
		}
	}()

	actualPort := 0
	for i := 0; i < maxWaitAttempts; i++ {
		actualPort = publicServer.ActualPort()
		if actualPort > 0 {
			break
		}

		time.Sleep(waitInterval)
	}

	if actualPort <= 0 {
		return nil, "", fmt.Errorf("server did not bind to port")
	}

	baseURL := "https://" + cryptoutilMagic.IPv4Loopback + ":" + strconv.Itoa(actualPort)

	return publicServer, baseURL, nil
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

// registerServiceUser registers a user and returns the user with private key.
func registerServiceUser(t *testing.T, client *http.Client, baseURL, username, password string) *testUser {
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

	// Login to get JWT token.
	token := loginUser(t, client, baseURL, username, password)

	return &testUser{
		ID:       userID,
		Username: username,
		Password: password,
		Token:    token,
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

// receiveMessagesService retrieves messages for the specified receiver.
func receiveMessagesService(t *testing.T, client *http.Client, baseURL, token string) []map[string]any {
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

// deleteMessageService deletes a message via /service/api/v1/messages/:id.
func deleteMessageService(t *testing.T, client *http.Client, baseURL, messageID, token string) {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, baseURL+"/service/api/v1/messages/"+messageID, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusNoContent, resp.StatusCode)
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

	// Login to get JWT token.
	token := loginUserBrowser(t, client, baseURL, username, password)

	return &testUser{
		ID:       userID,
		Username: username,
		Password: password,
		Token:    token,
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

func registerTestUserService(t *testing.T, client *http.Client, baseURL string) *testUser {
	username, err := cryptoutilRandom.GenerateUsernameSimple()
	require.NoError(t, err, "failed to generate username")

	password, err := cryptoutilRandom.GeneratePasswordSimple()
	require.NoError(t, err, "failed to generate password")

	return registerServiceUser(t, client, baseURL, username, password)
}

func registerTestUserBrowser(t *testing.T, client *http.Client, baseURL string) *testUser {
	username, err := cryptoutilRandom.GenerateUsernameSimple()
	require.NoError(t, err, "failed to generate username")

	password, err := cryptoutilRandom.GeneratePasswordSimple()
	require.NoError(t, err, "failed to generate password")

	return registerUserBrowser(t, client, baseURL, username, password)
}
