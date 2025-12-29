// Copyright (c) 2025 Justin Cranford
//

package server_test

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

	cryptoutilDomain "cryptoutil/internal/learn/domain"
	"cryptoutil/internal/learn/repository"
	"cryptoutil/internal/learn/server"
	cryptoutilConfig "cryptoutil/internal/shared/config"
	cryptoutilTLSGenerator "cryptoutil/internal/shared/config/tls_generator"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
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

	// Run migrations using embedded migration files.
	err = repository.ApplyMigrations(sqlDB, repository.DatabaseTypeSQLite)
	require.NoError(t, err)

	return db
}

// initTestConfig creates a properly configured AppConfig for testing.
func initTestConfig() *server.AppConfig {
	cfg := server.DefaultAppConfig()
	cfg.BindPublicPort = 0                     // Dynamic port allocation for tests
	cfg.BindPrivatePort = 0                    // Dynamic port allocation for tests
	cfg.OTLPService = "learn-im-test"          // Required for telemetry initialization
	cfg.LogLevel = "info"                      // Required for logger initialization
	cfg.OTLPEndpoint = "grpc://localhost:4317" // Required for OTLP endpoint validation
	cfg.OTLPEnabled = false                    // Disable actual OTLP export in tests

	return cfg
}

// createTestPublicServer creates a PublicServer for testing.
func createTestPublicServer(t *testing.T, db *gorm.DB) (*server.PublicServer, string) {
	t.Helper()

	ctx := context.Background()

	userRepo := repository.NewUserRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	messageRecipientJWKRepo := repository.NewMessageRecipientJWKRepository(db)

	// Initialize telemetry for JWKGenService (minimal config for tests).
	telemetrySettings := &cryptoutilConfig.ServerSettings{
		LogLevel:     "info",
		OTLPService:  "learn-im-test",
		OTLPEnabled:  false, // Tests use in-process telemetry only.
		OTLPEndpoint: "grpc://localhost:4317",
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
		[]string{"localhost"},
		[]string{cryptoutilMagic.IPv4Loopback},
		cryptoutilMagic.TLSTestEndEntityCertValidity1Year,
	)
	require.NoError(t, err)

	// Generate random JWT secret for tests (QUIZME Q8 requirement - no hardcoded secrets).
	jwtSecretID, err := googleUuid.NewV7()
	require.NoError(t, err)

	jwtSecret := jwtSecretID.String()

	publicServer, err := server.NewPublicServer(ctx, testPort, userRepo, messageRepo, messageRecipientJWKRepo, jwkGenService, jwtSecret, tlsCfg)
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
		Timeout: cryptoutilMagic.LearnDefaultTimeout, // Increased for concurrent test execution.
	}
}

// testUserWithToken represents a test user with authentication token.
type testUserWithToken struct {
	User  *cryptoutilDomain.User
	Token string
}

// registerAndLoginTestUser registers a user and logs in to get JWT token.
func registerAndLoginTestUser(t *testing.T, client *http.Client, baseURL string) *testUserWithToken {
	t.Helper()

	// Generate random username and password (QUIZME Q8 requirement - no hardcoded passwords).
	usernameID, err := googleUuid.NewV7()
	require.NoError(t, err)

	username := "user_" + usernameID.String()[:8]

	passwordID, err := googleUuid.NewV7()
	require.NoError(t, err)

	password := passwordID.String()

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

	return &cryptoutilDomain.User{
		ID:       userID,
		Username: username,
	}
}
