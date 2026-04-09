// Copyright (c) 2025 Justin Cranford
//
//

package application

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps/framework/service/config"
	cryptoutilAppsFrameworkServiceServerBarrier "cryptoutil/internal/apps/framework/service/server/barrier"
	cryptoutilAppsFrameworkServiceServerRepository "cryptoutil/internal/apps/framework/service/server/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestStartBasic_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
		DevMode:      true,
		VerboseMode:  false,
		OTLPService:  "template-service-test",
		OTLPEnabled:  false,
		OTLPEndpoint: cryptoutilSharedMagic.DefaultOTLPEndpointDefault,
		LogLevel:     cryptoutilSharedMagic.DefaultLogLevelInfo,
	}

	basic, err := StartBasic(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, basic)

	defer basic.Shutdown()

	// Verify services initialized.
	require.NotNil(t, basic.TelemetryService)
	require.NotNil(t, basic.UnsealKeysService)
	require.NotNil(t, basic.JWKGenService)
	require.Equal(t, settings, basic.Settings)
}

func TestStartBasic_NilValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ctx      context.Context
		settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings
		wantErr  string
	}{
		{name: "nil context", ctx: nil, settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{}, wantErr: "ctx cannot be nil"},
		{name: "nil settings", ctx: context.Background(), settings: nil, wantErr: "settings cannot be nil"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			basic, err := StartBasic(tc.ctx, tc.settings) //nolint:staticcheck // Testing nil context error handling.
			require.Error(t, err)
			require.Nil(t, basic)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

// TestBasicShutdown tests graceful shutdown of basic infrastructure.
func TestBasicShutdown(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
		DevMode:      true,
		VerboseMode:  false,
		OTLPService:  "template-service-test",
		OTLPEnabled:  false,
		OTLPEndpoint: cryptoutilSharedMagic.DefaultOTLPEndpointDefault,
		LogLevel:     cryptoutilSharedMagic.DefaultLogLevelInfo,
	}

	basic, err := StartBasic(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, basic)

	// Shutdown should not panic.
	require.NotPanics(t, func() {
		basic.Shutdown()
	})
}

// TestInitializeServicesOnCore_Success tests successful service initialization.
func TestInitializeServicesOnCore_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Use unique temporary file database to avoid shared state pollution.
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Convert to proper file URI (file:///abs/path on all platforms).
	slashPath := filepath.ToSlash(dbPath)
	if !strings.HasPrefix(slashPath, "/") {
		slashPath = "/" + slashPath
	}

	dbName := fmt.Sprintf("file://%s?mode=rwc&cache=shared", slashPath)

	settings := &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
		DevMode:                    true,
		VerboseMode:                false,
		DatabaseURL:                dbName,
		OTLPService:                "template-service-test",
		OTLPEnabled:                false,
		OTLPEndpoint:               cryptoutilSharedMagic.DefaultOTLPEndpointDefault,
		LogLevel:                   cryptoutilSharedMagic.DefaultLogLevelInfo,
		BrowserSessionAlgorithm:    cryptoutilSharedMagic.DefaultServiceSessionAlgorithm,
		BrowserSessionJWSAlgorithm: cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		BrowserSessionJWEAlgorithm: cryptoutilSharedMagic.JoseAlgRSAOAEP,
		BrowserSessionExpiration:   15 * time.Minute,
		ServiceSessionAlgorithm:    cryptoutilSharedMagic.DefaultServiceSessionAlgorithm,
		ServiceSessionJWSAlgorithm: cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		ServiceSessionJWEAlgorithm: cryptoutilSharedMagic.JoseAlgRSAOAEP,
		ServiceSessionExpiration:   1 * time.Hour,
		SessionIdleTimeout:         cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days * time.Minute,
		SessionCleanupInterval:     1 * time.Hour,
	}

	// Start core with database.
	core, err := StartCore(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, core)

	defer core.Shutdown()

	// Run migrations (required for all services).
	err = core.DB.AutoMigrate(
		&cryptoutilAppsFrameworkServiceServerBarrier.RootKey{},
		&cryptoutilAppsFrameworkServiceServerBarrier.IntermediateKey{},
		&cryptoutilAppsFrameworkServiceServerBarrier.ContentKey{},
		&cryptoutilAppsFrameworkServiceServerRepository.BrowserSessionJWK{},
		&cryptoutilAppsFrameworkServiceServerRepository.ServiceSessionJWK{},
		&cryptoutilAppsFrameworkServiceServerRepository.BrowserSession{},
		&cryptoutilAppsFrameworkServiceServerRepository.ServiceSession{},
	)
	require.NoError(t, err)

	// Initialize services on core.
	services, err := InitializeServicesOnCore(ctx, core, settings)
	require.NoError(t, err)
	require.NotNil(t, services)

	// Verify all services initialized.
	require.NotNil(t, services.Repository)
	require.NotNil(t, services.BarrierService)
	require.NotNil(t, services.RealmRepository)
	require.NotNil(t, services.RealmService)
	require.NotNil(t, services.SessionManager)
	require.NotNil(t, services.TenantRepository)
	require.NotNil(t, services.UserRepository)
	require.NotNil(t, services.JoinRequestRepository)
	require.NotNil(t, services.RegistrationService)
	require.NotNil(t, services.RotationService)
	require.NotNil(t, services.StatusService)
}

// TestCoreShutdown tests graceful shutdown of core infrastructure.
func TestCoreShutdown(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
		DevMode:      true,
		VerboseMode:  false,
		DatabaseURL:  cryptoutilSharedMagic.SQLiteInMemoryDSN,
		OTLPService:  "template-service-test",
		OTLPEnabled:  false,
		OTLPEndpoint: cryptoutilSharedMagic.DefaultOTLPEndpointDefault,
		LogLevel:     cryptoutilSharedMagic.DefaultLogLevelInfo,
	}

	core, err := StartCore(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, core)

	// Shutdown should not panic.
	require.NotPanics(t, func() {
		core.Shutdown()
	})
}

// Mock servers for Listener testing.
type mockPublicServer struct {
	port        int
	baseURL     string
	startErr    error
	shutdownErr error
	startDone   chan struct{}
}

func (m *mockPublicServer) Start(_ context.Context) error {
	if m.startDone != nil {
		<-m.startDone
	}

	return m.startErr
}

func (m *mockPublicServer) Shutdown(_ context.Context) error {
	return m.shutdownErr
}

func (m *mockPublicServer) ActualPort() int {
	return m.port
}

func (m *mockPublicServer) PublicBaseURL() string {
	return m.baseURL
}

type mockAdminServer struct {
	port        int
	baseURL     string
	ready       bool
	startErr    error
	shutdownErr error
	startDone   chan struct{}
}

func (m *mockAdminServer) Start(_ context.Context) error {
	if m.startDone != nil {
		<-m.startDone
	}

	return m.startErr
}

func (m *mockAdminServer) Shutdown(_ context.Context) error {
	return m.shutdownErr
}

func (m *mockAdminServer) ActualPort() int {
	return m.port
}

func (m *mockAdminServer) SetReady(ready bool) {
	m.ready = ready
}

func (m *mockAdminServer) AdminBaseURL() string {
	return m.baseURL
}

// TestStartListener tests creation and initialization of Listener.
// Sequential: uses shared SQLite in-memory database.
func TestStartListener(t *testing.T) {
	// NOT parallel - uses shared SQLite database.
	ctx := context.Background()
	settings := &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
		DevMode:      true,
		VerboseMode:  false,
		DatabaseURL:  cryptoutilSharedMagic.SQLiteInMemoryDSN,
		OTLPService:  "test-listener",
		OTLPEnabled:  false,
		OTLPEndpoint: cryptoutilSharedMagic.DefaultOTLPEndpointDefault,
		LogLevel:     cryptoutilSharedMagic.DefaultLogLevelInfo,
	}

	publicServer := &mockPublicServer{port: cryptoutilSharedMagic.TestServerPort, baseURL: "https://localhost:8080"}
	adminServer := &mockAdminServer{port: cryptoutilSharedMagic.JoseJAAdminPort, baseURL: "https://localhost:9090"}

	config := &ListenerConfig{
		Settings:     settings,
		PublicServer: publicServer,
		AdminServer:  adminServer,
	}

	listener, err := StartListener(ctx, config)
	require.NoError(t, err)
	require.NotNil(t, listener)
	require.NotNil(t, listener.Core)
	require.Equal(t, publicServer, listener.PublicServer)
	require.Equal(t, adminServer, listener.AdminServer)
	require.Equal(t, settings, listener.Settings)

	defer func() { _ = listener.Shutdown(context.Background()) }()
}

func TestStartListener_NilValidation(t *testing.T) {
	t.Parallel()

	defaultSettings := &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
		DevMode:     true,
		DatabaseURL: cryptoutilSharedMagic.SQLiteInMemoryDSN,
	}

	tests := []struct {
		name    string
		ctx     context.Context
		config  *ListenerConfig
		wantErr string
	}{
		{
			name:    "nil context",
			ctx:     nil,
			config:  &ListenerConfig{Settings: defaultSettings, PublicServer: &mockPublicServer{}, AdminServer: &mockAdminServer{}},
			wantErr: "ctx cannot be nil",
		},
		{
			name:    "nil config",
			ctx:     context.Background(),
			config:  nil,
			wantErr: "config cannot be nil",
		},
		{
			name:    "nil settings",
			ctx:     context.Background(),
			config:  &ListenerConfig{Settings: nil, PublicServer: &mockPublicServer{}, AdminServer: &mockAdminServer{}},
			wantErr: "settings cannot be nil",
		},
		{
			name:    "nil public server",
			ctx:     context.Background(),
			config:  &ListenerConfig{Settings: defaultSettings, PublicServer: nil, AdminServer: &mockAdminServer{}},
			wantErr: "publicServer cannot be nil",
		},
		{
			name:    "nil admin server",
			ctx:     context.Background(),
			config:  &ListenerConfig{Settings: defaultSettings, PublicServer: &mockPublicServer{}, AdminServer: nil},
			wantErr: "adminServer cannot be nil",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			listener, err := StartListener(tc.ctx, tc.config) //nolint:staticcheck // Testing nil context error handling.
			require.Error(t, err)
			require.Nil(t, listener)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestListener_PortAccessors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		listener *Listener
		accessor func(*Listener) int
		want     int
	}{
		{name: "public port", listener: &Listener{PublicServer: &mockPublicServer{port: 12345}}, accessor: (*Listener).PublicPort, want: 12345},
		{name: "public port nil server", listener: &Listener{PublicServer: nil}, accessor: (*Listener).PublicPort, want: 0},
		{name: "admin port", listener: &Listener{AdminServer: &mockAdminServer{port: 54321}}, accessor: (*Listener).AdminPort, want: 54321},
		{name: "admin port nil server", listener: &Listener{AdminServer: nil}, accessor: (*Listener).AdminPort, want: 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, tc.accessor(tc.listener))
		})
	}
}

// TestListener_Shutdown tests graceful shutdown of Listener.
// Sequential: uses shared SQLite in-memory database.
func TestListener_Shutdown(t *testing.T) {
	// NOT parallel - uses shared SQLite database.
	ctx := context.Background()
	settings := &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
		DevMode:      true,
		DatabaseURL:  cryptoutilSharedMagic.SQLiteInMemoryDSN,
		OTLPService:  "test-shutdown",
		OTLPEnabled:  false,
		OTLPEndpoint: cryptoutilSharedMagic.DefaultOTLPEndpointDefault,
		LogLevel:     cryptoutilSharedMagic.DefaultLogLevelInfo,
	}

	publicServer := &mockPublicServer{port: cryptoutilSharedMagic.TestServerPort}
	adminServer := &mockAdminServer{port: cryptoutilSharedMagic.JoseJAAdminPort}

	config := &ListenerConfig{
		Settings:     settings,
		PublicServer: publicServer,
		AdminServer:  adminServer,
	}

	listener, err := StartListener(ctx, config)
	require.NoError(t, err)

	err = listener.Shutdown(ctx)
	require.NoError(t, err)
	require.False(t, adminServer.ready) // Should set ready=false.
}

// TestListener_Shutdown_NilContext tests Shutdown with nil context.
// Sequential: uses shared SQLite in-memory database.
func TestListener_Shutdown_NilContext(t *testing.T) {
	// NOT parallel - uses shared SQLite database.
	ctx := context.Background()
	settings := &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
		DevMode:      true,
		DatabaseURL:  cryptoutilSharedMagic.SQLiteInMemoryDSN,
		OTLPService:  "test-shutdown-nil-ctx",
		OTLPEnabled:  false,
		OTLPEndpoint: cryptoutilSharedMagic.DefaultOTLPEndpointDefault,
		LogLevel:     cryptoutilSharedMagic.DefaultLogLevelInfo,
	}

	publicServer := &mockPublicServer{port: cryptoutilSharedMagic.TestServerPort}
	adminServer := &mockAdminServer{port: cryptoutilSharedMagic.JoseJAAdminPort}

	config := &ListenerConfig{
		Settings:     settings,
		PublicServer: publicServer,
		AdminServer:  adminServer,
	}

	listener, err := StartListener(ctx, config)
	require.NoError(t, err)

	// Shutdown with nil context should use Background.
	err = listener.Shutdown(nil) //nolint:staticcheck // Testing nil context.
	require.NoError(t, err)
}

// TestOpenPostgreSQL tests PostgreSQL database connection.
// NOTE: Requires PostgreSQL running. Uses environment variable DATABASE_URL_POSTGRES if available.
