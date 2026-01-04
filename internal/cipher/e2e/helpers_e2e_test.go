// Copyright (c) 2025 Justin Cranford
//
//

package e2e_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"cryptoutil/internal/cipher/repository"
	"cryptoutil/internal/cipher/server"
	"cryptoutil/internal/cipher/server/config"
	cryptoutilConfig "cryptoutil/internal/shared/config"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilE2E "cryptoutil/internal/template/testing/e2e"
)

// testPort uses port 0 for dynamic allocation (prevents port conflicts in tests).
const testPort = 0

var testJWTSecret = googleUuid.Must(googleUuid.NewV7()).String() // TODO Use random secret in DB, protected at rest with barrier layer encryption

// initTestDB creates an in-memory SQLite database with schema using template helper.
func initTestDB() (*gorm.DB, error) {
	ctx := context.Background()

	applyMigrations := func(sqlDB *sql.DB) error {
		return repository.ApplyMigrations(sqlDB, repository.DatabaseTypeSQLite)
	}

	return cryptoutilE2E.InitTestDB(ctx, applyMigrations)
}

// createTestPublicServer creates a PublicServer for testing using shared resources from TestMain.
func createTestPublicServer(db *gorm.DB) (*server.PublicServer, string, error) {
	ctx := context.Background()

	userRepo := repository.NewUserRepository(db)
	messageRepo := repository.NewMessageRepository(db)

	// Note: This old helper creates PublicServer without barrier service.
	// For full server testing, use createTestCipherIMServer() instead.
	publicServer, err := server.NewPublicServer(ctx, testPort, userRepo, messageRepo, nil, sharedJWKGenService, nil, testJWTSecret, sharedTLSConfig)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create PublicServer: %w", err)
	}

	// Start server in background using template helper.
	_ = cryptoutilE2E.StartServerAsync(ctx, publicServer)

	// Wait for server to bind to port using template helper.
	waitParams := cryptoutilE2E.DefaultServerWaitParams()

	baseURL, err := cryptoutilE2E.WaitForServerPort(publicServer, waitParams)
	if err != nil {
		return nil, "", err
	}

	return publicServer, baseURL, nil
}

// createTestCipherIMServer creates a full CipherIMServer for testing using shared resources from TestMain.
// Returns the server instance, public URL, and admin URL.
func createTestCipherIMServer(db *gorm.DB) (*server.CipherIMServer, string, string, error) {
	ctx := context.Background()

	// Create AppConfig with test settings.
	cfg := &config.AppConfig{
		ServerSettings: cryptoutilConfig.ServerSettings{
			BindPublicProtocol:    cryptoutilMagic.ProtocolHTTPS,
			BindPublicAddress:     cryptoutilMagic.IPv4Loopback,
			BindPublicPort:        0, // Dynamic allocation
			BindPrivateProtocol:   cryptoutilMagic.ProtocolHTTPS,
			BindPrivateAddress:    cryptoutilMagic.IPv4Loopback,
			BindPrivatePort:       0, // Dynamic allocation
			TLSPublicDNSNames:     []string{cryptoutilMagic.HostnameLocalhost},
			TLSPublicIPAddresses:  []string{cryptoutilMagic.IPv4Loopback},
			TLSPrivateDNSNames:    []string{cryptoutilMagic.HostnameLocalhost},
			TLSPrivateIPAddresses: []string{cryptoutilMagic.IPv4Loopback},
			CORSAllowedOrigins:    []string{},
			OTLPService:           "cipher-im-e2e-test",
			OTLPEndpoint:          "grpc://localhost:4317",
			LogLevel:              "error",
		},
		JWTSecret: testJWTSecret,
	}

	// Create full server.
	cipherServer, err := server.New(ctx, cfg, db, repository.DatabaseTypeSQLite)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to create cipher server: %w", err)
	}

	// Start server in background.
	errChan := make(chan error, 1)

	go func() {
		if startErr := cipherServer.Start(ctx); startErr != nil {
			errChan <- startErr
		}
	}()

	// Wait for both servers to bind to ports.
	const (
		maxWaitAttempts = 50
		waitInterval    = 100 * time.Millisecond
	)

	var publicPort int
	var adminPort int

	for i := 0; i < maxWaitAttempts; i++ {
		publicPort = cipherServer.PublicPort()

		adminPortValue, _ := cipherServer.AdminPort()
		adminPort = adminPortValue

		if publicPort > 0 && adminPort > 0 {
			break
		}

		select {
		case err := <-errChan:
			return nil, "", "", fmt.Errorf("server startup error: %w", err)
		case <-time.After(waitInterval):
		}
	}

	if publicPort == 0 {
		return nil, "", "", fmt.Errorf("createTestCipherIMServer: public server did not bind to port")
	}

	if adminPort == 0 {
		return nil, "", "", fmt.Errorf("createTestCipherIMServer: admin server did not bind to port")
	}

	publicURL := fmt.Sprintf("https://%s:%d", cryptoutilMagic.IPv4Loopback, publicPort)
	adminURL := fmt.Sprintf("https://%s:%d", cryptoutilMagic.IPv4Loopback, adminPort)

	return cipherServer, publicURL, adminURL, nil
}

// loginUser logs in a user and returns JWT token (delegates to template helper).
func loginUser(t *testing.T, client *http.Client, baseURL, username, password string) string {
	t.Helper()

	return cryptoutilE2E.LoginUser(t, client, baseURL, "/service/api/v1/users/login", username, password)
}

// registerServiceUser registers a user and returns the user with private key (delegates to template helper).
func registerServiceUser(t *testing.T, client *http.Client, baseURL, username, password string) *cryptoutilE2E.TestUser {
	t.Helper()

	return cryptoutilE2E.RegisterServiceUser(t, client, baseURL, username, password)
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

	resp := cryptoutilE2E.SendAuthenticatedRequest(t, client, http.MethodPut, baseURL+"/service/api/v1/messages/tx", token, reqJSON)

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var respBody map[string]string

	cryptoutilE2E.DecodeJSONResponse(t, resp, &respBody)

	return respBody["message_id"]
}

// receiveMessagesService retrieves messages for the specified receiver.
func receiveMessagesService(t *testing.T, client *http.Client, baseURL, token string) []map[string]any {
	t.Helper()

	resp := cryptoutilE2E.SendAuthenticatedRequest(t, client, http.MethodGet, baseURL+"/service/api/v1/messages/rx", token, nil)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody map[string][]map[string]any

	cryptoutilE2E.DecodeJSONResponse(t, resp, &respBody)

	return respBody["messages"]
}

// deleteMessageService deletes a message via /service/api/v1/messages/:id.
func deleteMessageService(t *testing.T, client *http.Client, baseURL, messageID, token string) {
	t.Helper()

	resp := cryptoutilE2E.SendAuthenticatedRequest(t, client, http.MethodDelete, baseURL+"/service/api/v1/messages/"+messageID, token, nil)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

// registerUserBrowser registers a user via /browser/api/v1/users/register (delegates to template helper).
func registerUserBrowser(t *testing.T, client *http.Client, baseURL, username, password string) *cryptoutilE2E.TestUser {
	t.Helper()

	return cryptoutilE2E.RegisterBrowserUser(t, client, baseURL, username, password)
}

// loginUserBrowser logs in a user via /browser/api/v1/users/login and returns JWT token (delegates to template helper).
func loginUserBrowser(t *testing.T, client *http.Client, baseURL, username, password string) string {
	t.Helper()

	return cryptoutilE2E.LoginUser(t, client, baseURL, "/browser/api/v1/users/login", username, password)
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

	resp := cryptoutilE2E.SendAuthenticatedRequest(t, client, http.MethodPut, baseURL+"/browser/api/v1/messages/tx", token, reqJSON)

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var respBody map[string]string

	cryptoutilE2E.DecodeJSONResponse(t, resp, &respBody)

	return respBody["message_id"]
}

// receiveMessagesBrowser retrieves messages for the specified receiver via /browser/api/v1/messages/rx.
func receiveMessagesBrowser(t *testing.T, client *http.Client, baseURL, token string) []map[string]any {
	t.Helper()

	resp := cryptoutilE2E.SendAuthenticatedRequest(t, client, http.MethodGet, baseURL+"/browser/api/v1/messages/rx", token, nil)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody map[string][]map[string]any

	cryptoutilE2E.DecodeJSONResponse(t, resp, &respBody)

	return respBody["messages"]
}

// deleteMessageBrowser deletes a message via /browser/api/v1/messages/:id.
func deleteMessageBrowser(t *testing.T, client *http.Client, baseURL, messageID, token string) {
	t.Helper()

	resp := cryptoutilE2E.SendAuthenticatedRequest(t, client, http.MethodDelete, baseURL+"/browser/api/v1/messages/"+messageID, token, nil)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func registerTestUserService(t *testing.T, client *http.Client, baseURL string) *cryptoutilE2E.TestUser {
	t.Helper()

	return cryptoutilE2E.RegisterTestUserService(t, client, baseURL)
}

func registerTestUserBrowser(t *testing.T, client *http.Client, baseURL string) *cryptoutilE2E.TestUser {
	t.Helper()

	return cryptoutilE2E.RegisterTestUserBrowser(t, client, baseURL)
}
