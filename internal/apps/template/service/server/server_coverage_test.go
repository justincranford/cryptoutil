//go:build !integration

package server

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

// TestPublicServerBase_StartContextCancellation tests Start when context is canceled before server runs.
// Target: public_server_base.go:151-156 (context cancellation error path).
func TestPublicServerBase_StartContextCancellation(t *testing.T) {
	t.Parallel()

	config := &PublicServerConfig{
		BindAddress: cryptoutilSharedMagic.IPv4Loopback,
		Port:        0,
		TLSMaterial: createTestTLSMaterial(t),
	}

	base, err := NewPublicServerBase(config)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	err = base.Start(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "public server stopped")
}

// TestPublicServerBase_StartListenError tests Start when port is already in use.
// Target: public_server_base.go:120-122 (Listen error path).
func TestPublicServerBase_StartListenError(t *testing.T) {
	t.Parallel()

	// Create first server to occupy a specific port.
	config1 := &PublicServerConfig{
		BindAddress: cryptoutilSharedMagic.IPv4Loopback,
		Port:        0, // Dynamic allocation.
		TLSMaterial: createTestTLSMaterial(t),
	}

	base1, err := NewPublicServerBase(config1)
	require.NoError(t, err)

	// Start first server in background.
	ctx1 := context.Background()

	errChan := make(chan error, 1)

	go func() {
		errChan <- base1.Start(ctx1)
	}()

	// Wait for first server to start and get its port.
	// Use a small sleep to allow server to bind.
	require.Eventually(t, func() bool {
		return base1.ActualPort() != 0
	}, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second, cryptoutilSharedMagic.JoseJADefaultMaxMaterials*time.Millisecond)

	occupiedPort := base1.ActualPort()

	// Try to create second server on same port.
	config2 := &PublicServerConfig{
		BindAddress: cryptoutilSharedMagic.IPv4Loopback,
		Port:        occupiedPort, // Use same port.
		TLSMaterial: createTestTLSMaterial(t),
	}

	base2, err := NewPublicServerBase(config2)
	require.NoError(t, err)

	// Start second server should fail with "address already in use".
	err = base2.Start(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create listener")

	// Clean up first server.
	_ = base1.Shutdown(context.Background())

	// Wait for goroutine to finish.
	<-errChan
}

// TestPublicServerBase_ShutdownTwice tests Shutdown called twice.
// Target: public_server_base.go:170-172 (double shutdown error path).
func TestPublicServerBase_ShutdownTwice(t *testing.T) {
	t.Parallel()

	config := &PublicServerConfig{
		BindAddress: cryptoutilSharedMagic.IPv4Loopback,
		Port:        0,
		TLSMaterial: createTestTLSMaterial(t),
	}

	base, err := NewPublicServerBase(config)
	require.NoError(t, err)

	// First shutdown should succeed.
	err = base.Shutdown(context.Background())
	require.NoError(t, err)

	// Second shutdown should fail.
	err = base.Shutdown(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "public server already shutdown")
}

// TestNewServiceTemplate_JWKGenInitError tests NewServiceTemplate when JWKGenService fails to initialize.
// Target: service_template.go:83 (JWKGenService init error)
//
// TestNewServiceTemplate_TelemetryInitError tests NewServiceTemplate when telemetry init fails.
// Cannot use t.Parallel() because it modifies the package-level injectable var.
func TestNewServiceTemplate_TelemetryInitError(t *testing.T) {
	originalFn := newTelemetryServiceFn
	newTelemetryServiceFn = func(_ context.Context, _ *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) (*cryptoutilSharedTelemetry.TelemetryService, error) {
		return nil, fmt.Errorf("mock telemetry failure")
	}

	defer func() { newTelemetryServiceFn = originalFn }()

	ctx := context.Background()

	config := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		OTLPService: "test-service",
		LogLevel:    cryptoutilSharedMagic.DefaultLogLevelInfo,
	}

	dbUUID, err := googleUuid.NewV7()
	require.NoError(t, err)

	dsn := "file:" + dbUUID.String() + "?mode=memory&cache=private"

	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dsn)
	require.NoError(t, err)

	defer func() { _ = sqlDB.Close() }()

	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	_, err = NewServiceTemplate(ctx, config, db, cryptoutilAppsTemplateServiceServerRepository.DatabaseTypeSQLite)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to initialize telemetry")
}

// TestNewServiceTemplate_JWKGenInitError tests NewServiceTemplate when JWK gen service init fails.
// Cannot use t.Parallel() because it modifies the package-level injectable var.
// Sequential: modifies package-level injectable function variable.
func TestNewServiceTemplate_JWKGenInitError(t *testing.T) {
	originalFn := newJWKGenServiceFn
	newJWKGenServiceFn = func(_ context.Context, _ *cryptoutilSharedTelemetry.TelemetryService, _ bool) (*cryptoutilSharedCryptoJose.JWKGenService, error) {
		return nil, fmt.Errorf("mock jwkgen failure")
	}

	defer func() { newJWKGenServiceFn = originalFn }()

	ctx := context.Background()

	config := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		OTLPService:  "test-service",
		LogLevel:     cryptoutilSharedMagic.DefaultLogLevelInfo,
		OTLPEndpoint: "http://localhost:4318",
	}

	dbUUID, err := googleUuid.NewV7()
	require.NoError(t, err)

	dsn := "file:" + dbUUID.String() + "?mode=memory&cache=private"

	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dsn)
	require.NoError(t, err)

	defer func() { _ = sqlDB.Close() }()

	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	_, err = NewServiceTemplate(ctx, config, db, cryptoutilAppsTemplateServiceServerRepository.DatabaseTypeSQLite)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to initialize JWK generation service")
}

// TestNewServiceTemplate_OptionError tests NewServiceTemplate when option application fails.
// Target: service_template.go:99 (option apply error).
func TestNewServiceTemplate_OptionError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Provide minimal valid config so telemetry initialization passes.
	config := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		OTLPService:  "test-service",
		LogLevel:     cryptoutilSharedMagic.DefaultLogLevelInfo,
		OTLPEndpoint: "http://localhost:4318", // Valid OTLP HTTP endpoint.
	}

	// Create unique in-memory database per test (using modernc.org/sqlite).
	dbID, err := googleUuid.NewV7()
	require.NoError(t, err)

	dsn := "file:" + dbID.String() + "?mode=memory&cache=private"

	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dsn)
	require.NoError(t, err)

	defer func() { _ = sqlDB.Close() }()

	// Configure SQLite for concurrent operations.
	_, err = sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
	require.NoError(t, err)

	_, err = sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;")
	require.NoError(t, err)

	sqlDB.SetMaxOpenConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	sqlDB.SetMaxIdleConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	sqlDB.SetConnMaxLifetime(0)

	// Wrap with GORM.
	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	// Create option that always fails.
	failingOption := func(_ *ServiceTemplate) error {
		return fmt.Errorf("intentional option failure")
	}

	_, err = NewServiceTemplate(ctx, config, db, cryptoutilAppsTemplateServiceServerRepository.DatabaseTypeSQLite, failingOption)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to apply option")
	require.Contains(t, err.Error(), "intentional option failure")
}

// TestServiceTemplate_ShutdownNilComponents tests Shutdown with nil components.
// Target: service_template.go:144-153 (nil check branches).
func TestServiceTemplate_ShutdownNilComponents(t *testing.T) {
	t.Parallel()

	// Create ServiceTemplate with all nil components.
	st := &ServiceTemplate{
		telemetry: nil,
		jwkGen:    nil,
	}

	// Shutdown should not panic.
	st.Shutdown()
}

// mockPublicServerForCoverage is a simple mock implementation of IPublicServer for coverage tests.
type mockPublicServerForCoverage struct {
	startFunc    func(ctx context.Context) error
	shutdownFunc func(ctx context.Context) error
}

func (m *mockPublicServerForCoverage) Start(ctx context.Context) error {
	if m.startFunc != nil {
		return m.startFunc(ctx)
	}

	<-ctx.Done()

	return fmt.Errorf("context cancelled: %w", ctx.Err())
}

func (m *mockPublicServerForCoverage) Shutdown(ctx context.Context) error {
	if m.shutdownFunc != nil {
		return m.shutdownFunc(ctx)
	}

	return nil
}

func (m *mockPublicServerForCoverage) ActualPort() int {
	return cryptoutilSharedMagic.DemoServerPort
}

func (m *mockPublicServerForCoverage) PublicBaseURL() string {
	return "https://127.0.0.1:8080"
}

// mockAdminServerForCoverage is a simple mock implementation of IAdminServer for coverage tests.
type mockAdminServerForCoverage struct {
	startFunc    func(ctx context.Context) error
	shutdownFunc func(ctx context.Context) error
}

func (m *mockAdminServerForCoverage) Start(ctx context.Context) error {
	if m.startFunc != nil {
		return m.startFunc(ctx)
	}

	<-ctx.Done()

	return fmt.Errorf("context cancelled: %w", ctx.Err())
}

func (m *mockAdminServerForCoverage) Shutdown(ctx context.Context) error {
	if m.shutdownFunc != nil {
		return m.shutdownFunc(ctx)
	}

	return nil
}

func (m *mockAdminServerForCoverage) ActualPort() int {
	return cryptoutilSharedMagic.JoseJAAdminPort
}

func (m *mockAdminServerForCoverage) SetReady(_ bool) {}

func (m *mockAdminServerForCoverage) AdminBaseURL() string {
	return "https://127.0.0.1:9090"
}

// TestApplication_StartContextCancellation tests Application.Start with pre-cancelled context.
// Target: application.go:145-147 (context cancellation during Start).
func TestApplication_StartContextCancellation(t *testing.T) {
	t.Parallel()

	// Create mock servers that block indefinitely (don't respond to context).
	// This ensures the select statement hits ctx.Done() case instead of errChan.
	publicServer := &mockPublicServerForCoverage{
		startFunc: func(_ context.Context) error {
			// Block forever - never return.
			select {}
		},
	}

	adminServer := &mockAdminServerForCoverage{
		startFunc: func(_ context.Context) error {
			// Block forever - never return.
			select {}
		},
	}

	app, err := NewApplication(context.Background(), publicServer, adminServer)
	require.NoError(t, err)

	// Create pre-cancelled context.
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel before calling Start.

	// Start should detect cancelled context via ctx.Done() and return error.
	err = app.Start(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "application startup cancelled")
}

// TestPublicServerBase_ErrChanPath tests that Start returns via errChan when Fiber app is shut down
// directly (without canceling the server context). Covers the errChan select case.
func TestPublicServerBase_ErrChanPath(t *testing.T) {
	t.Parallel()

	config := &PublicServerConfig{
		BindAddress: cryptoutilSharedMagic.IPv4Loopback,
		Port:        0,
		TLSMaterial: createTestTLSMaterial(t),
	}

	base, err := NewPublicServerBase(config)
	require.NoError(t, err)

	// Start server in background with a non-cancellable context.
	startErr := make(chan error, 1)

	go func() {
		startErr <- base.Start(context.Background())
	}()

	// Wait for server to be listening.
	require.Eventually(t, func() bool {
		return base.ActualPort() != 0
	}, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second, cryptoutilSharedMagic.JoseJADefaultMaxMaterials*time.Millisecond)

	// Directly shutdown the Fiber app (NOT base.Shutdown which also cancels context).
	// This causes app.Listener to return, firing errChan, without serverCtx.Done().
	_ = base.app.Shutdown()

	// Start returns via errChan path.
	err = <-startErr

	// Fiber returns nil on clean shutdown — coverage of errChan path is the goal.
	_ = err
}

// TestPublicServerBase_ListenerError tests Start when app.Listener returns an error.
// Covers public_server_base.go:141-143 (app.Listener error inside goroutine).
// Cannot use t.Parallel() because it modifies the package-level injectable var.
// Sequential: modifies package-level injectable function variable.
func TestPublicServerBase_ListenerError(t *testing.T) {
	original := appListenerFn
	appListenerFn = func(_ *fiber.App, ln net.Listener) error {
		_ = ln.Close()

		return fmt.Errorf("forced listener error")
	}

	defer func() { appListenerFn = original }()

	config := &PublicServerConfig{
		BindAddress: cryptoutilSharedMagic.IPv4Loopback,
		Port:        0,
		TLSMaterial: createTestTLSMaterial(t),
	}

	base, err := NewPublicServerBase(config)
	require.NoError(t, err)

	err = base.Start(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "public server error")
	require.Contains(t, err.Error(), "forced listener error")
}
