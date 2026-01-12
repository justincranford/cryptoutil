// Copyright (c) 2025 Justin Cranford
//

package server_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cryptoutil/internal/apps/cipher/im/repository"
	"cryptoutil/internal/apps/cipher/im/server"
	cryptoutilTemplateServiceTesting "cryptoutil/internal/apps/template/service/testing/httpservertests"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// TestNewPublicServer_NilContext tests constructor with nil context.
func TestNewPublicServer_NilContext(t *testing.T) {
	t.Parallel()

	// Use shared resources from TestMain - no need to create db/repos/tlsCfg in every test.
	cleanTestDB(t)

	userRepo := repository.NewUserRepository(testDB)
	messageRepo := repository.NewMessageRepository(testDB)
	messageRecipientJWKRepo := repository.NewMessageRecipientJWKRepository(testDB, nil)

	_, err := server.NewPublicServer(context.Background(), cryptoutilMagic.IPv4Loopback, 0, userRepo, messageRepo, messageRecipientJWKRepo, nil, nil, nil, testTLSCfg)
	require.Error(t, err)
	// The test passes nil for jwkGenService, so that validation triggers first
	require.Contains(t, err.Error(), "JWK generation service cannot be nil")
}

// TestNewPublicServer_NilUserRepo tests constructor with nil user repository.
func TestNewPublicServer_NilUserRepo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cleanTestDB(t)

	messageRepo := repository.NewMessageRepository(testDB)
	messageRecipientJWKRepo := repository.NewMessageRecipientJWKRepository(testDB, nil)

	_, err := server.NewPublicServer(ctx, cryptoutilMagic.IPv4Loopback, 0, nil, messageRepo, messageRecipientJWKRepo, nil, nil, nil, testTLSCfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "user repository cannot be nil")
}

// TestNewPublicServer_NilMessageRepo tests constructor with nil message repository.
func TestNewPublicServer_NilMessageRepo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cleanTestDB(t)

	userRepo := repository.NewUserRepository(testDB)
	messageRecipientJWKRepo := repository.NewMessageRecipientJWKRepository(testDB, nil)

	_, err := server.NewPublicServer(ctx, cryptoutilMagic.IPv4Loopback, 0, userRepo, nil, messageRecipientJWKRepo, nil, nil, nil, testTLSCfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "message repository cannot be nil")
}

// TestNewPublicServer_NilMessageRecipientJWKRepo tests constructor with nil message recipient JWK repository.
func TestNewPublicServer_NilMessageRecipientJWKRepo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cleanTestDB(t)

	userRepo := repository.NewUserRepository(testDB)
	messageRepo := repository.NewMessageRepository(testDB)

	_, err := server.NewPublicServer(ctx, cryptoutilMagic.IPv4Loopback, 0, userRepo, messageRepo, nil, nil, nil, nil, testTLSCfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "message recipient JWK repository cannot be nil")
}

// TestNewPublicServer_NilTLSConfig tests constructor with nil TLS config.
func TestNewPublicServer_NilTLSConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cleanTestDB(t)

	userRepo := repository.NewUserRepository(testDB)
	messageRepo := repository.NewMessageRepository(testDB)
	messageRecipientJWKRepo := repository.NewMessageRecipientJWKRepository(testDB, nil)

	// Get session manager from testCipherIMServer to pass validation.
	sessionManager := testCipherIMServer.SessionManager()

	_, err := server.NewPublicServer(ctx, cryptoutilMagic.IPv4Loopback, 0, userRepo, messageRepo, messageRecipientJWKRepo, testJWKGenService, nil, sessionManager, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "TLS configuration cannot be nil")
}

// TestHandleServiceHealth_WhileRunning tests health endpoint while server running.
func TestHandleServiceHealth_WhileRunning(t *testing.T) {
	t.Parallel()

	_, svcBaseURL := createTestPublicServer(t, testDB)
	client := createHTTPClient(t)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, svcBaseURL+"/service/api/v1/health", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.Equal(t, "healthy", respBody["status"])
}

// TestHandleBrowserHealth_WhileRunning tests browser health endpoint.
func TestHandleBrowserHealth_WhileRunning(t *testing.T) {
	t.Parallel()

	_, svcBaseURL := createTestPublicServer(t, testDB)
	client := createHTTPClient(t)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, svcBaseURL+"/browser/api/v1/health", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.Equal(t, "healthy", respBody["status"])
}

// TestShutdown_MultipleCalls tests calling Shutdown multiple times.
func TestShutdown_MultipleCalls(t *testing.T) {
	t.Parallel()

	createServer := func(t *testing.T) cryptoutilTemplateServiceTesting.HTTPServer {
		t.Helper()
		publicServer, _ := createTestPublicServer(t, testDB)

		return publicServer
	}

	cryptoutilTemplateServiceTesting.TestShutdownDoubleCall(t, createServer)
}

// TestPublicServer_StartContextCancelled tests server shutdown via context cancellation.
func TestPublicServer_StartContextCancelled(t *testing.T) {
	t.Parallel()

	srv, _ := createTestPublicServer(t, testDB)

	ctx, cancel := context.WithCancel(context.Background())

	errChan := make(chan error, 1)

	go func() {
		errChan <- srv.Start(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	cancel()

	err := <-errChan
	require.Error(t, err)
	require.Contains(t, err.Error(), "context canceled")
}

// TestPublicServer_DoubleShutdown tests calling Shutdown twice.
func TestPublicServer_DoubleShutdown(t *testing.T) {
	t.Parallel()

	createServer := func(t *testing.T) cryptoutilTemplateServiceTesting.HTTPServer {
		t.Helper()
		srv, _ := createTestPublicServer(t, testDB)

		return srv
	}

	cryptoutilTemplateServiceTesting.TestShutdownDoubleCall(t, createServer)
}

// TestShutdown_DuplicateCall tests calling Shutdown twice.
func TestShutdown_DuplicateCall(t *testing.T) {
	t.Parallel()

	createServer := func(t *testing.T) cryptoutilTemplateServiceTesting.HTTPServer {
		t.Helper()
		server, _ := createTestPublicServer(t, testDB)

		return server
	}

	cryptoutilTemplateServiceTesting.TestShutdownDoubleCall(t, createServer)
}

// TestStart_ContextCancelled tests server start with cancelled context.
func TestStart_ContextCancelled(t *testing.T) {
	t.Parallel()

	cleanTestDB(t)

	cfg := initTestConfig()

	srv, err := server.NewFromConfig(context.Background(), cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = srv.Start(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "context canceled")

	_ = srv.Shutdown(context.Background())
}
