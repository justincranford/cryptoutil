//go:build !integration

package server

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilTemplateServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// TestPublicServerBase_StartContextCancellation tests Start when context is canceled before server runs.
// Target: public_server_base.go:151-156 (context cancellation error path).
func TestPublicServerBase_StartContextCancellation(t *testing.T) {
	t.Parallel()

	config := &PublicServerConfig{
		BindAddress: "127.0.0.1",
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

// TestPublicServerBase_ShutdownTwice tests Shutdown called twice.
// Target: public_server_base.go:170-172 (double shutdown error path).
func TestPublicServerBase_ShutdownTwice(t *testing.T) {
	t.Parallel()

	config := &PublicServerConfig{
		BindAddress: "127.0.0.1",
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
// NOTE: Cannot easily trigger JWKGen init error in practice (requires telemetry/pool failure).
// This test documents the code path exists but is difficult to cover.
func TestNewServiceTemplate_JWKGenInitError(t *testing.T) {
	t.Skip("JWKGen init error difficult to trigger - requires telemetry/pool failure")
}

// TestNewServiceTemplate_OptionError tests NewServiceTemplate when option application fails.
// Target: service_template.go:99 (option apply error).
func TestNewServiceTemplate_OptionError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Provide minimal valid config so telemetry initialization passes.
	config := &cryptoutilConfig.ServiceTemplateServerSettings{
		OTLPService:  "test-service",
		LogLevel:     "INFO",
		OTLPEndpoint: "http://localhost:4318", // Valid OTLP HTTP endpoint.
	}

	// Create unique in-memory database per test (using modernc.org/sqlite).
	dbID, err := googleUuid.NewV7()
	require.NoError(t, err)

	dsn := "file:" + dbID.String() + "?mode=memory&cache=private"

	sqlDB, err := sql.Open("sqlite", dsn)
	require.NoError(t, err)

	defer func() { _ = sqlDB.Close() }()

	// Configure SQLite for concurrent operations.
	_, err = sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
	require.NoError(t, err)

	_, err = sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;")
	require.NoError(t, err)

	sqlDB.SetMaxOpenConns(cryptoutilMagic.SQLiteMaxOpenConnections)
	sqlDB.SetMaxIdleConns(cryptoutilMagic.SQLiteMaxOpenConnections)
	sqlDB.SetConnMaxLifetime(0)

	// Wrap with GORM.
	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	// Create option that always fails.
	failingOption := func(st *ServiceTemplate) error {
		return fmt.Errorf("intentional option failure")
	}

	_, err = NewServiceTemplate(ctx, config, db, cryptoutilTemplateServerRepository.DatabaseTypeSQLite, failingOption)
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
		barrier:   nil,
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
	return 8080
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
	return 9090
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
