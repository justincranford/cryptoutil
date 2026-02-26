// Copyright (c) 2025 Justin Cranford
//
//

package listener_test

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"testing"

	"gorm.io/gorm"

	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilAppsTemplateServiceServerListener "cryptoutil/internal/apps/template/service/server/listener"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilAppsTemplateServiceServerTestutil "cryptoutil/internal/apps/template/service/server/testutil"

	"github.com/stretchr/testify/require"
)

// Test constants for OTLP configuration.
const (
	testLogLevel     = "info"
	testOTLPService  = "test-service"
	testOTLPEndpoint = "http://localhost:4318"
)

// ===========================
// StartApplicationListener Tests
// ===========================

// TestStartApplicationListener_NilContext verifies nil context is rejected.
func TestStartApplicationListener_NilContext(t *testing.T) {
	t.Parallel()

	cfg := &cryptoutilAppsTemplateServiceServerListener.ApplicationConfig{
		ServiceTemplateServerSettings: cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(),
		DB:                            &gorm.DB{},
		DBType:                        cryptoutilAppsTemplateServiceServerRepository.DatabaseTypeSQLite,
		PublicServerFactory:           mockPublicServerFactory,
	}

	listener, err := cryptoutilAppsTemplateServiceServerListener.StartApplicationListener(nil, cfg) //nolint:staticcheck // Testing nil context.

	require.Error(t, err)
	require.Contains(t, err.Error(), "context cannot be nil")
	require.Nil(t, listener)
}

// TestStartApplicationListener_NilConfig verifies nil config is rejected.
func TestStartApplicationListener_NilConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	listener, err := cryptoutilAppsTemplateServiceServerListener.StartApplicationListener(ctx, nil)

	require.Error(t, err)
	require.Contains(t, err.Error(), "config cannot be nil")
	require.Nil(t, listener)
}

// TestStartApplicationListener_NilServiceTemplateServerSettings verifies nil settings rejected.
func TestStartApplicationListener_NilServiceTemplateServerSettings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := &cryptoutilAppsTemplateServiceServerListener.ApplicationConfig{
		ServiceTemplateServerSettings: nil,
		DB:                            &gorm.DB{},
		DBType:                        cryptoutilAppsTemplateServiceServerRepository.DatabaseTypeSQLite,
		PublicServerFactory:           mockPublicServerFactory,
	}

	listener, err := cryptoutilAppsTemplateServiceServerListener.StartApplicationListener(ctx, cfg)

	require.Error(t, err)
	require.Contains(t, err.Error(), "ServiceTemplateServerSettings cannot be nil")
	require.Nil(t, listener)
}

// TestStartApplicationListener_NilDB verifies nil database is rejected.
func TestStartApplicationListener_NilDB(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := &cryptoutilAppsTemplateServiceServerListener.ApplicationConfig{
		ServiceTemplateServerSettings: cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(),
		DB:                            nil,
		DBType:                        cryptoutilAppsTemplateServiceServerRepository.DatabaseTypeSQLite,
		PublicServerFactory:           mockPublicServerFactory,
	}

	listener, err := cryptoutilAppsTemplateServiceServerListener.StartApplicationListener(ctx, cfg)

	require.Error(t, err)
	require.Contains(t, err.Error(), "DB cannot be nil")
	require.Nil(t, listener)
}

// TestStartApplicationListener_NilPublicServerFactory verifies factory is required.
func TestStartApplicationListener_NilPublicServerFactory(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := &cryptoutilAppsTemplateServiceServerListener.ApplicationConfig{
		ServiceTemplateServerSettings: cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(),
		DB:                            &gorm.DB{},
		DBType:                        cryptoutilAppsTemplateServiceServerRepository.DatabaseTypeSQLite,
		PublicServerFactory:           nil,
	}

	listener, err := cryptoutilAppsTemplateServiceServerListener.StartApplicationListener(ctx, cfg)

	require.Error(t, err)
	require.Contains(t, err.Error(), "PublicServerFactory cannot be nil")
	require.Nil(t, listener)
}

// TestStartApplicationListener_ReturnsNotImplementedError verifies implementation in progress.
func TestStartApplicationListener_ReturnsNotImplementedError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Use valid settings with all required fields to pass template initialization.
	settings := cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings()
	settings.LogLevel = testLogLevel         // Valid log level.
	settings.OTLPService = testOTLPService   // Ensure service name is set.
	settings.OTLPEndpoint = testOTLPEndpoint // Valid OTLP endpoint.

	cfg := &cryptoutilAppsTemplateServiceServerListener.ApplicationConfig{
		ServiceTemplateServerSettings: settings,
		DB:                            &gorm.DB{},
		DBType:                        cryptoutilAppsTemplateServiceServerRepository.DatabaseTypeSQLite,
		PublicServerFactory:           mockPublicServerFactory,
	}

	listener, err := cryptoutilAppsTemplateServiceServerListener.StartApplicationListener(ctx, cfg)

	require.Error(t, err)
	require.Contains(t, err.Error(), "implementation in progress")
	require.NotNil(t, listener) // Returns partial listener even with error.
}

// ===========================
// ApplicationListener Method Tests
// ===========================

// TestApplicationListener_ActualPublicPort verifies port accessor returns expected value.
func TestApplicationListener_ActualPublicPort(t *testing.T) {
	t.Parallel()

	listener := &cryptoutilAppsTemplateServiceServerListener.ApplicationListener{}
	port := listener.ActualPublicPort()

	require.Equal(t, uint16(0), port) // Default value.
}

// TestApplicationListener_ActualPrivatePort verifies port accessor returns expected value.
func TestApplicationListener_ActualPrivatePort(t *testing.T) {
	t.Parallel()

	listener := &cryptoutilAppsTemplateServiceServerListener.ApplicationListener{}
	port := listener.ActualPrivatePort()

	require.Equal(t, uint16(0), port) // Default value.
}

// TestApplicationListener_Config verifies config accessor returns nil by default.
func TestApplicationListener_Config(t *testing.T) {
	t.Parallel()

	listener := &cryptoutilAppsTemplateServiceServerListener.ApplicationListener{}

	cfg := listener.Config()

	require.Nil(t, cfg) // Default nil value (config not set in this test).
}

// TestApplicationListener_Shutdown_NoApp verifies shutdown with nil app is safe.
func TestApplicationListener_Shutdown_NoApp(t *testing.T) {
	t.Parallel()

	listener := &cryptoutilAppsTemplateServiceServerListener.ApplicationListener{}

	// Shutdown should not panic with nil app.
	require.NotPanics(t, func() {
		listener.Shutdown()
	})
}

// TestApplicationListener_Shutdown_WithApp verifies shutdown properly calls app.Shutdown.
// This test covers the l.app != nil branch in Shutdown().
func TestApplicationListener_Shutdown_WithApp(t *testing.T) {
	t.Parallel()

	// Create a mock public server.
	mockPublicServer := &mockPublicServerImpl{}
	// Create a mock admin server.
	mockAdminServer := &mockAdminServerImpl{}

	// Create Application with mock servers.
	app, err := cryptoutilAppsTemplateServiceServer.NewApplication(context.Background(), mockPublicServer, mockAdminServer)
	require.NoError(t, err)

	listener := &cryptoutilAppsTemplateServiceServerListener.ApplicationListener{}
	listener.SetApplicationForTesting(app)

	// Shutdown should call app.Shutdown and not panic.
	require.NotPanics(t, func() {
		listener.Shutdown()
	})
}

// TestApplicationListener_Shutdown_WithShutdownFunc tests shutdown when shutdownFunc is set.
// Uses StartApplicationListener which sets up the shutdownFunc internally.
func TestApplicationListener_Shutdown_WithShutdownFunc(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Use valid settings with all required fields to pass template initialization.
	settings := cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings()
	settings.LogLevel = testLogLevel
	settings.OTLPService = testOTLPService
	settings.OTLPEndpoint = testOTLPEndpoint

	cfg := &cryptoutilAppsTemplateServiceServerListener.ApplicationConfig{
		ServiceTemplateServerSettings: settings,
		DB:                            &gorm.DB{},
		DBType:                        cryptoutilAppsTemplateServiceServerRepository.DatabaseTypeSQLite,
		PublicServerFactory:           mockPublicServerFactory,
	}

	listener, err := cryptoutilAppsTemplateServiceServerListener.StartApplicationListener(ctx, cfg)
	require.Error(t, err) // Expected: "implementation in progress".
	require.NotNil(t, listener)

	// Shutdown should call the internal shutdownFunc (covers shutdownFunc != nil branch).
	require.NotPanics(t, func() {
		listener.Shutdown()
	})

	// Second shutdown call should also be safe (idempotent).
	require.NotPanics(t, func() {
		listener.Shutdown()
	})
}

// ===========================
// Health Check Function Tests
// ===========================

// TestSendLivenessCheck_InvalidSettings verifies error handling with invalid settings.
func TestSendLivenessCheck_InvalidSettings(t *testing.T) {
	t.Parallel()

	// Create settings with invalid private base URL (will fail to connect).
	settings := cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings()
	settings.BindPrivatePort = 1 // Invalid port that won't have listener.

	_, err := cryptoutilAppsTemplateServiceServerListener.SendLivenessCheck(settings)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get liveness check")
}

// TestSendReadinessCheck_InvalidSettings verifies error handling with invalid settings.
func TestSendReadinessCheck_InvalidSettings(t *testing.T) {
	t.Parallel()

	// Create settings with invalid private base URL (will fail to connect).
	settings := cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings()
	settings.BindPrivatePort = 1 // Invalid port that won't have listener.

	_, err := cryptoutilAppsTemplateServiceServerListener.SendReadinessCheck(settings)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get readiness check")
}

// TestSendShutdownRequest_InvalidSettings verifies error handling with invalid settings.
func TestSendShutdownRequest_InvalidSettings(t *testing.T) {
	t.Parallel()

	// Create settings with invalid private base URL (will fail to connect).
	settings := cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings()
	settings.BindPrivatePort = 1 // Invalid port that won't have listener.

	err := cryptoutilAppsTemplateServiceServerListener.SendShutdownRequest(settings)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to send shutdown request")
}

// ===========================
// Mock Helpers
// ===========================

// mockPublicServerFactory is a test factory that returns nil server and error.
func mockPublicServerFactory(
	_ context.Context,
	_ *cryptoutilAppsTemplateServiceServerListener.ApplicationConfig,
	_ *cryptoutilAppsTemplateServiceServer.ServiceTemplate,
) (cryptoutilAppsTemplateServiceServer.IPublicServer, error) {
	return nil, nil // Minimal mock for validation tests.
}

// mockPublicServerImpl implements IPublicServer for testing.
type mockPublicServerImpl struct{}

func (m *mockPublicServerImpl) Start(_ context.Context) error    { return nil }
func (m *mockPublicServerImpl) Shutdown(_ context.Context) error { return nil }
func (m *mockPublicServerImpl) ActualPort() int                  { return cryptoutilSharedMagic.DemoServerPort }
func (m *mockPublicServerImpl) PublicBaseURL() string            { return "https://127.0.0.1:8080" }

// mockAdminServerImpl implements IAdminServer for testing.
type mockAdminServerImpl struct{}

func (m *mockAdminServerImpl) Start(_ context.Context) error    { return nil }
func (m *mockAdminServerImpl) Shutdown(_ context.Context) error { return nil }
func (m *mockAdminServerImpl) ActualPort() int                  { return cryptoutilSharedMagic.JoseJAAdminPort }
func (m *mockAdminServerImpl) SetReady(_ bool)                  {}
func (m *mockAdminServerImpl) AdminBaseURL() string             { return "https://127.0.0.1:9090" }
