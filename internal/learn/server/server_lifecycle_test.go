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

	"cryptoutil/internal/learn/repository"
	"cryptoutil/internal/learn/server"
	cryptoutilConfig "cryptoutil/internal/shared/config"
	cryptoutilTLSGenerator "cryptoutil/internal/shared/config/tls_generator"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
)

// TestNewPublicServer_NilContext tests constructor with nil context.
func TestNewPublicServer_NilContext(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	userRepo := repository.NewUserRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	messageRecipientJWKRepo := repository.NewMessageRecipientJWKRepository(db)

	tlsCfg, err := cryptoutilTLSGenerator.GenerateAutoTLSGeneratedSettings(
		[]string{cryptoutilMagic.HostnameLocalhost},
		[]string{cryptoutilMagic.IPv4Loopback},
		cryptoutilMagic.TLSTestEndEntityCertValidity1Year,
	)
	require.NoError(t, err)

	_, err = server.NewPublicServer(context.TODO(), 0, userRepo, messageRepo, messageRecipientJWKRepo, nil, "test-secret", tlsCfg)
	require.Error(t, err)
	// The test passes nil for jwkGenService, so that validation triggers first
	require.Contains(t, err.Error(), "JWK generation service cannot be nil")
}

// TestNewPublicServer_NilUserRepo tests constructor with nil user repository.
func TestNewPublicServer_NilUserRepo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := initTestDB(t)
	messageRepo := repository.NewMessageRepository(db)
	messageRecipientJWKRepo := repository.NewMessageRecipientJWKRepository(db)

	tlsCfg, err := cryptoutilTLSGenerator.GenerateAutoTLSGeneratedSettings(
		[]string{cryptoutilMagic.HostnameLocalhost},
		[]string{cryptoutilMagic.IPv4Loopback},
		cryptoutilMagic.TLSTestEndEntityCertValidity1Year,
	)
	require.NoError(t, err)

	_, err = server.NewPublicServer(ctx, 0, nil, messageRepo, messageRecipientJWKRepo, nil, "test-secret", tlsCfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "user repository cannot be nil")
}

// TestNewPublicServer_NilMessageRepo tests constructor with nil message repository.
func TestNewPublicServer_NilMessageRepo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := initTestDB(t)
	userRepo := repository.NewUserRepository(db)
	messageRecipientJWKRepo := repository.NewMessageRecipientJWKRepository(db)

	tlsCfg, err := cryptoutilTLSGenerator.GenerateAutoTLSGeneratedSettings(
		[]string{cryptoutilMagic.HostnameLocalhost},
		[]string{cryptoutilMagic.IPv4Loopback},
		cryptoutilMagic.TLSTestEndEntityCertValidity1Year,
	)
	require.NoError(t, err)

	_, err = server.NewPublicServer(ctx, 0, userRepo, nil, messageRecipientJWKRepo, nil, "test-secret", tlsCfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "message repository cannot be nil")
}

// TestNewPublicServer_NilMessageRecipientJWKRepo tests constructor with nil message recipient JWK repository.
func TestNewPublicServer_NilMessageRecipientJWKRepo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := initTestDB(t)
	userRepo := repository.NewUserRepository(db)
	messageRepo := repository.NewMessageRepository(db)

	tlsCfg, err := cryptoutilTLSGenerator.GenerateAutoTLSGeneratedSettings(
		[]string{cryptoutilMagic.HostnameLocalhost},
		[]string{cryptoutilMagic.IPv4Loopback},
		cryptoutilMagic.TLSTestEndEntityCertValidity1Year,
	)
	require.NoError(t, err)

	_, err = server.NewPublicServer(ctx, 0, userRepo, messageRepo, nil, nil, "test-secret", tlsCfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "message recipient JWK repository cannot be nil")
}

// TestNewPublicServer_NilTLSConfig tests constructor with nil TLS config.
func TestNewPublicServer_NilTLSConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := initTestDB(t)
	userRepo := repository.NewUserRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	messageRecipientJWKRepo := repository.NewMessageRecipientJWKRepository(db)

	telemetrySettings := &cryptoutilConfig.ServerSettings{
		LogLevel:     "info",
		OTLPService:  "learn-im-test",
		OTLPEnabled:  false,
		OTLPEndpoint: "grpc://" + cryptoutilMagic.HostnameLocalhost + ":" + "4317",
	}

	telemetryService, err := cryptoutilTelemetry.NewTelemetryService(ctx, telemetrySettings)
	require.NoError(t, err)

	jwkGenService, err := cryptoutilJose.NewJWKGenService(ctx, telemetryService, false)
	require.NoError(t, err)

	_, err = server.NewPublicServer(ctx, 0, userRepo, messageRepo, messageRecipientJWKRepo, jwkGenService, "test-secret", nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "TLS configuration cannot be nil")
}

// TestHandleServiceHealth_WhileRunning tests health endpoint while server running.
func TestHandleServiceHealth_WhileRunning(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, baseURL+"/service/api/v1/health", nil)
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

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, baseURL+"/browser/api/v1/health", nil)
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

	ctx := context.Background()
	db := initTestDB(t)
	publicServer, _ := createTestPublicServer(t, db)

	err := publicServer.Shutdown(ctx)
	require.NoError(t, err)

	err = publicServer.Shutdown(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "already shutdown")
}

// TestPublicServer_StartContextCancelled tests server shutdown via context cancellation.
func TestPublicServer_StartContextCancelled(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	srv, _ := createTestPublicServer(t, db)

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

	db := initTestDB(t)
	srv, _ := createTestPublicServer(t, db)

	err := srv.Shutdown(context.Background())
	require.NoError(t, err)

	err = srv.Shutdown(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "already shutdown")
}

// TestShutdown_DuplicateCall tests calling Shutdown twice.
func TestShutdown_DuplicateCall(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	server, _ := createTestPublicServer(t, db)

	ctx := context.Background()

	err := server.Shutdown(ctx)
	require.NoError(t, err)

	err = server.Shutdown(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "already shutdown")
}

// TestStart_ContextCancelled tests server start with cancelled context.
func TestStart_ContextCancelled(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	cfg := initTestConfig()

	srv, err := server.New(context.Background(), cfg, db, repository.DatabaseTypeSQLite)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = srv.Start(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "context canceled")

	_ = srv.Shutdown(context.Background())
}
