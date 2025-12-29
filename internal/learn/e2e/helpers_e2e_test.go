// Copyright (c) 2025 Justin Cranford
//
//

package e2e_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // CGO-free SQLite driver

	"cryptoutil/internal/learn/repository"
	"cryptoutil/internal/learn/server"
	cryptoutilConfig "cryptoutil/internal/shared/config"
	cryptoutilTLSGenerator "cryptoutil/internal/shared/config/tls_generator"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
)

// testUser represents a user with their authentication token for testing.
type testUser struct {
	ID       googleUuid.UUID // UUIDv7
	Username string
	Token    string // JWT authentication token.
}

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

	// Run migrations using embedded migration files.
	err = repository.ApplyMigrations(sqlDB, repository.DatabaseTypeSQLite)
	require.NoError(t, err)

	return db
}

// createTestPublicServer creates a PublicServer for testing.
func createTestPublicServer(t *testing.T, db *gorm.DB) (*server.PublicServer, string) {
	t.Helper()

	ctx := context.Background()

	userRepo := repository.NewUserRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	messageRecipientJWKRepo := repository.NewMessageRecipientJWKRepository(db)

	// Initialize telemetry for JWKGenService (minimal config for e2e tests).
	telemetrySettings := &cryptoutilConfig.ServerSettings{
		LogLevel:     "info",
		OTLPService:  "learn-im-e2e",
		OTLPEnabled:  false, // E2E tests use in-process telemetry only.
		OTLPEndpoint: "grpc://" + cryptoutilMagic.HostnameLocalhost + ":" + "4317",
	}

	telemetryService, err := cryptoutilTelemetry.NewTelemetryService(ctx, telemetrySettings)
	require.NoError(t, err)

	// Initialize JWK Generation Service for message encryption.
	jwkGenService, err := cryptoutilJose.NewJWKGenService(ctx, telemetryService, false)
	require.NoError(t, err)

	// Use port 0 for dynamic allocation (prevents port conflicts in tests).
	const testPort = 0

	// TLS config with localhost subject.
	tlsCfg, err := cryptoutilTLSGenerator.GenerateAutoTLSGeneratedSettings(
		[]string{cryptoutilMagic.HostnameLocalhost},
		[]string{cryptoutilMagic.IPv4Loopback},
		cryptoutilMagic.TLSTestEndEntityCertValidity1Year,
	)
	require.NoError(t, err)

	const testJWTSecret = "learn-im-test-secret-e2e"

	publicServer, err := server.NewPublicServer(ctx, testPort, userRepo, messageRepo, messageRecipientJWKRepo, jwkGenService, testJWTSecret, tlsCfg)
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
		Timeout: cryptoutilMagic.LearnDefaultTimeout, // Increased for concurrent test execution (matches server tests).
	}
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

	// Login to get JWT token.
	token := loginUser(t, client, baseURL, username, password)

	return &testUser{
		ID:       userID,
		Username: username,
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

// deleteMessage deletes a message via /service/api/v1/messages/:id.
func deleteMessage(t *testing.T, client *http.Client, baseURL, messageID, token string) {
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
