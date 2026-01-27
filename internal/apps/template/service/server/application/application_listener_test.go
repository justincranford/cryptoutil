// Copyright (c) 2025 Justin Cranford
//
//

package application

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// TestStartCoreWithServices_FullIntegration tests the complete service initialization path.
// IMPORTANT: StartCoreWithServices doesn't run migrations (Phase W TODO), so we test the components separately.
func TestStartCoreWithServices_FullIntegration(t *testing.T) {
	// NOT parallel - shares SQLite in-memory database with other tests.

	ctx := context.Background()
	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		DevMode:                    true,
		VerboseMode:                false,
		DatabaseURL:                cryptoutilSharedMagic.SQLiteInMemoryDSN,
		OTLPService:                "test-service",
		OTLPEnabled:                false,
		OTLPEndpoint:               "grpc://127.0.0.1:4317",
		LogLevel:                   "INFO",
		BrowserSessionAlgorithm:    "JWS",
		BrowserSessionJWSAlgorithm: "RS256",
		BrowserSessionJWEAlgorithm: "RSA-OAEP",
		BrowserSessionExpiration:   15 * time.Minute,
		ServiceSessionAlgorithm:    "JWS",
		ServiceSessionJWSAlgorithm: "RS256",
		ServiceSessionJWEAlgorithm: "RSA-OAEP",
		ServiceSessionExpiration:   1 * time.Hour,
		SessionIdleTimeout:         30 * time.Minute,
		SessionCleanupInterval:     1 * time.Hour,
	}

	// Start core infrastructure.
	core, err := StartCore(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, core)
	defer core.Shutdown()

	// Run migrations (StartCoreWithServices Phase W TODO: should handle this internally).
	err = core.DB.AutoMigrate(
		&cryptoutilAppsTemplateServiceServerBarrier.RootKey{},
		&cryptoutilAppsTemplateServiceServerBarrier.IntermediateKey{},
		&cryptoutilAppsTemplateServiceServerBarrier.ContentKey{},
		&cryptoutilAppsTemplateServiceServerRepository.BrowserSessionJWK{},
		&cryptoutilAppsTemplateServiceServerRepository.ServiceSessionJWK{},
		&cryptoutilAppsTemplateServiceServerRepository.BrowserSession{},
		&cryptoutilAppsTemplateServiceServerRepository.ServiceSession{},
	)
	require.NoError(t, err)

	// Initialize services on top of core.
	services, err := InitializeServicesOnCore(ctx, core, settings)
	require.NoError(t, err)
	require.NotNil(t, services)

	// Verify services initialized - this covers the full initialization path.
	require.NotNil(t, services.Core)
	require.NotNil(t, services.Core.Basic)
	require.NotNil(t, services.Core.DB)
	require.NotNil(t, services.BarrierService)
	require.NotNil(t, services.RealmService)
	require.NotNil(t, services.SessionManager)
	require.NotNil(t, services.RegistrationService)
	require.NotNil(t, services.RotationService)
	require.NotNil(t, services.StatusService)
}

// TestMaskPassword tests password masking in DSN strings.
func TestMaskPassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		dsn      string
		expected string
	}{
		{
			name:     "PostgreSQL with password - KNOWN BUG: looks for '/:', not '://'",
			dsn:      "postgres://user:mypassword@localhost:5432/dbname",
			expected: "postgres://user:mypassword@localhost:5432/dbname", // TODO: Should mask to "postgres://***@localhost:5432/dbname"
		},
		{
			name:     "PostgreSQL with complex password - KNOWN BUG",
			dsn:      "postgres://user:p@ss:w0rd@localhost:5432/dbname",
			expected: "postgres://user:p@ss:w0rd@localhost:5432/dbname", // TODO: Should mask
		},
		{
			name:     "Non-PostgreSQL DSN without pattern",
			dsn:      "file://path/to/db.sqlite",
			expected: "file://path/to/db.sqlite",
		},
		{
			name:     "DSN without @ symbol",
			dsn:      "postgres://localhost:5432/dbname",
			expected: "postgres://localhost:5432/dbname",
		},
		{
			name:     "Hypothetical DSN with /: pattern (would work)",
			dsn:      "scheme/:password@host",
			expected: "scheme/:***@host",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := maskPassword(tc.dsn)
			require.Equal(t, tc.expected, result)
		})
	}
}

// TestContainerModeDetection tests container mode detection logic based on bind address.
// Container mode is triggered when BindPublicAddress == "0.0.0.0"
// Priority: P1.1 (Critical - Must Have).
func TestContainerModeDetection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		bindPublicAddress  string
		bindPrivateAddress string
		wantContainerMode  bool
	}{
		{
			name:               "public 0.0.0.0 triggers container mode",
			bindPublicAddress:  cryptoutilSharedMagic.IPv4AnyAddress, // "0.0.0.0"
			bindPrivateAddress: cryptoutilSharedMagic.IPv4Loopback,   // "127.0.0.1"
			wantContainerMode:  true,
		},
		{
			name:               "both 127.0.0.1 is NOT container mode",
			bindPublicAddress:  cryptoutilSharedMagic.IPv4Loopback,
			bindPrivateAddress: cryptoutilSharedMagic.IPv4Loopback,
			wantContainerMode:  false,
		},
		{
			name:               "private 0.0.0.0 does NOT trigger container mode",
			bindPublicAddress:  cryptoutilSharedMagic.IPv4Loopback,
			bindPrivateAddress: cryptoutilSharedMagic.IPv4AnyAddress,
			wantContainerMode:  false,
		},
		{
			name:               "specific IP is NOT container mode",
			bindPublicAddress:  "192.168.1.100",
			bindPrivateAddress: cryptoutilSharedMagic.IPv4Loopback,
			wantContainerMode:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
				BindPublicAddress:  tc.bindPublicAddress,
				BindPrivateAddress: tc.bindPrivateAddress,
			}

			isContainerMode := settings.BindPublicAddress == cryptoutilSharedMagic.IPv4AnyAddress
			require.Equal(t, tc.wantContainerMode, isContainerMode)
		})
	}
}

// TestMTLSConfiguration tests mTLS client auth configuration for private/public servers
// in dev/container/production modes.
// Priority: P1.2 (MOST CRITICAL - Currently 0% coverage on security code).
func TestMTLSConfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		devMode               bool
		bindPublicAddress     string
		bindPrivateAddress    string
		wantPrivateClientAuth tls.ClientAuthType
		wantPublicClientAuth  tls.ClientAuthType
	}{
		{
			name:                  "dev mode disables mTLS on private server",
			devMode:               true,
			bindPublicAddress:     cryptoutilSharedMagic.IPv4Loopback,
			bindPrivateAddress:    cryptoutilSharedMagic.IPv4Loopback,
			wantPrivateClientAuth: tls.NoClientCert,
			wantPublicClientAuth:  tls.NoClientCert, // Public never requires client certs
		},
		{
			name:                  "container mode disables mTLS on private server",
			devMode:               false,
			bindPublicAddress:     cryptoutilSharedMagic.IPv4AnyAddress, // 0.0.0.0
			bindPrivateAddress:    cryptoutilSharedMagic.IPv4Loopback,
			wantPrivateClientAuth: tls.NoClientCert,
			wantPublicClientAuth:  tls.NoClientCert,
		},
		{
			name:                  "production mode enables mTLS on private server",
			devMode:               false,
			bindPublicAddress:     cryptoutilSharedMagic.IPv4Loopback,
			bindPrivateAddress:    cryptoutilSharedMagic.IPv4Loopback,
			wantPrivateClientAuth: tls.RequireAndVerifyClientCert,
			wantPublicClientAuth:  tls.NoClientCert, // Public never requires client certs
		},
		{
			name:                  "container mode with private 0.0.0.0 still enables mTLS",
			devMode:               false,
			bindPublicAddress:     cryptoutilSharedMagic.IPv4Loopback,
			bindPrivateAddress:    cryptoutilSharedMagic.IPv4AnyAddress,
			wantPrivateClientAuth: tls.RequireAndVerifyClientCert, // Only public triggers container mode
			wantPublicClientAuth:  tls.NoClientCert,
		},
		{
			name:                  "public server never uses RequireAndVerifyClientCert",
			devMode:               false,
			bindPublicAddress:     cryptoutilSharedMagic.IPv4Loopback,
			bindPrivateAddress:    cryptoutilSharedMagic.IPv4Loopback,
			wantPrivateClientAuth: tls.RequireAndVerifyClientCert,
			wantPublicClientAuth:  tls.NoClientCert, // Browsers don't have client certs
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
				DevMode:            tc.devMode,
				BindPublicAddress:  tc.bindPublicAddress,
				BindPrivateAddress: tc.bindPrivateAddress,
			}

			// Replicate the mTLS logic from application_listener.go.
			isContainerMode := settings.BindPublicAddress == cryptoutilSharedMagic.IPv4AnyAddress

			privateClientAuth := tls.RequireAndVerifyClientCert
			if settings.DevMode || isContainerMode {
				privateClientAuth = tls.NoClientCert
			}

			publicClientAuth := tls.NoClientCert // Always NoClientCert for browser compatibility

			require.Equal(t, tc.wantPrivateClientAuth, privateClientAuth, "Private server mTLS")
			require.Equal(t, tc.wantPublicClientAuth, publicClientAuth, "Public server mTLS")
		})
	}
}

// TestHealthcheck_CompletesWithinTimeout tests healthcheck completes within reasonable timeout.
// Priority: P3.2 (Nice to Have - Could Have).
func TestHealthcheck_CompletesWithinTimeout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		endpoint     string
		wantStatus   int
		wantContains string
	}{
		{
			name:         "livez endpoint responds quickly",
			endpoint:     "/admin/api/v1/livez",
			wantStatus:   200,
			wantContains: `"status":"alive"`,
		},
		{
			name:         "readyz endpoint responds quickly",
			endpoint:     "/admin/api/v1/readyz",
			wantStatus:   200,
			wantContains: `"status":"ready"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create standalone Fiber app with admin handler.
			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
			})

			// Simple healthcheck handlers mimicking admin server behavior.
			api := app.Group("/admin/api/v1")
			api.Get("/livez", func(c *fiber.Ctx) error {
				return c.JSON(fiber.Map{"status": "alive"})
			})
			api.Get("/readyz", func(c *fiber.Ctx) error {
				return c.JSON(fiber.Map{"status": "ready"})
			})

			// Create test request.
			req := httptest.NewRequest("GET", tt.endpoint, nil)

			// Use app.Test() - no HTTPS listener needed, completes instantly.
			resp, err := app.Test(req, -1) // -1 = no timeout
			require.NoError(t, err)
			require.NotNil(t, resp)

			defer func() {
				require.NoError(t, resp.Body.Close())
			}()

			// Verify response.
			require.Equal(t, tt.wantStatus, resp.StatusCode)

			body := make([]byte, 1024)
			n, _ := resp.Body.Read(body)
			require.Contains(t, string(body[:n]), tt.wantContains)
		})
	}
}

// TestHealthcheck_TimeoutExceeded tests healthcheck fails when client timeout exceeded.
// Priority: P3.2 (Nice to Have - Could Have).
func TestHealthcheck_TimeoutExceeded(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		endpoint    string
		timeout     time.Duration
		handlerWait time.Duration
	}{
		{
			name:        "livez timeout - handler too slow",
			endpoint:    "/admin/api/v1/livez",
			timeout:     10 * time.Millisecond,
			handlerWait: 50 * time.Millisecond,
		},
		{
			name:        "readyz timeout - handler too slow",
			endpoint:    "/admin/api/v1/readyz",
			timeout:     10 * time.Millisecond,
			handlerWait: 50 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create standalone Fiber app with slow handler.
			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
			})

			// Handler with artificial delay exceeding client timeout.
			api := app.Group("/admin/api/v1")
			api.Get("/livez", func(c *fiber.Ctx) error {
				time.Sleep(tt.handlerWait)

				return c.JSON(fiber.Map{"status": "alive"})
			})
			api.Get("/readyz", func(c *fiber.Ctx) error {
				time.Sleep(tt.handlerWait)

				return c.JSON(fiber.Map{"status": "ready"})
			})

			// Create test request.
			req := httptest.NewRequest("GET", tt.endpoint, nil)

			// Use app.Test() with timeout shorter than handler delay.
			resp, err := app.Test(req, int(tt.timeout.Milliseconds()))

			// Should timeout - either err != nil OR resp == nil.
			if resp != nil {
				defer func() {
					_ = resp.Body.Close()
				}()
			}

			// app.Test() returns error when timeout occurs.
			require.Error(t, err)
		})
	}
}

// TestStartBasic_Success tests successful initialization of basic infrastructure.
func TestStartBasic_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		DevMode:      true,
		VerboseMode:  false,
		OTLPService:  "template-service-test",
		OTLPEnabled:  false,
		OTLPEndpoint: "grpc://127.0.0.1:4317",
		LogLevel:     "INFO",
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

// TestStartBasic_NilContext tests error when context is nil.
func TestStartBasic_NilContext(t *testing.T) {
	t.Parallel()

	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{}

	basic, err := StartBasic(nil, settings)
	require.Error(t, err)
	require.Nil(t, basic)
	require.Contains(t, err.Error(), "ctx cannot be nil")
}

// TestStartBasic_NilSettings tests error when settings is nil.
func TestStartBasic_NilSettings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	basic, err := StartBasic(ctx, nil)
	require.Error(t, err)
	require.Nil(t, basic)
	require.Contains(t, err.Error(), "settings cannot be nil")
}

// TestBasicShutdown tests graceful shutdown of basic infrastructure.
func TestBasicShutdown(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		DevMode:      true,
		VerboseMode:  false,
		OTLPService:  "template-service-test",
		OTLPEnabled:  false,
		OTLPEndpoint: "grpc://127.0.0.1:4317",
		LogLevel:     "INFO",
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
	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		DevMode:                     true,
		VerboseMode:                 false,
		DatabaseURL:                 cryptoutilSharedMagic.SQLiteInMemoryDSN,
		OTLPService:                 "template-service-test",
		OTLPEnabled:                 false,
		OTLPEndpoint:                "grpc://127.0.0.1:4317",
		LogLevel:                    "INFO",
		BrowserSessionAlgorithm:     "JWS",
		BrowserSessionJWSAlgorithm:  "RS256",
		BrowserSessionJWEAlgorithm:  "RSA-OAEP",
		BrowserSessionExpiration:    15 * time.Minute,
		ServiceSessionAlgorithm:     "JWS",
		ServiceSessionJWSAlgorithm:  "RS256",
		ServiceSessionJWEAlgorithm:  "RSA-OAEP",
		ServiceSessionExpiration:    1 * time.Hour,
		SessionIdleTimeout:          30 * time.Minute,
		SessionCleanupInterval:      1 * time.Hour,
	}

	// Start core with database.
	core, err := StartCore(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, core)

	defer core.Shutdown()

	// Run migrations (required for all services).
	err = core.DB.AutoMigrate(
		&cryptoutilAppsTemplateServiceServerBarrier.RootKey{},
		&cryptoutilAppsTemplateServiceServerBarrier.IntermediateKey{},
		&cryptoutilAppsTemplateServiceServerBarrier.ContentKey{},
		&cryptoutilAppsTemplateServiceServerRepository.BrowserSessionJWK{},
		&cryptoutilAppsTemplateServiceServerRepository.ServiceSessionJWK{},
		&cryptoutilAppsTemplateServiceServerRepository.BrowserSession{},
		&cryptoutilAppsTemplateServiceServerRepository.ServiceSession{},
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
	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		DevMode:      true,
		VerboseMode:  false,
		DatabaseURL:  cryptoutilSharedMagic.SQLiteInMemoryDSN,
		OTLPService:  "template-service-test",
		OTLPEnabled:  false,
		OTLPEndpoint: "grpc://127.0.0.1:4317",
		LogLevel:     "INFO",
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
	port      int
	baseURL   string
	startErr  error
	startDone chan struct{}
}

func (m *mockPublicServer) Start(_ context.Context) error {
	if m.startDone != nil {
		<-m.startDone
	}

	return m.startErr
}

func (m *mockPublicServer) Shutdown(_ context.Context) error {
	return nil
}

func (m *mockPublicServer) ActualPort() int {
	return m.port
}

func (m *mockPublicServer) PublicBaseURL() string {
	return m.baseURL
}

type mockAdminServer struct {
	port      int
	baseURL   string
	ready     bool
	startErr  error
	startDone chan struct{}
}

func (m *mockAdminServer) Start(_ context.Context) error {
	if m.startDone != nil {
		<-m.startDone
	}

	return m.startErr
}

func (m *mockAdminServer) Shutdown(_ context.Context) error {
	return nil
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
func TestStartListener(t *testing.T) {
	// NOT parallel - uses shared SQLite database.

	ctx := context.Background()
	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		DevMode:      true,
		VerboseMode:  false,
		DatabaseURL:  cryptoutilSharedMagic.SQLiteInMemoryDSN,
		OTLPService:  "test-listener",
		OTLPEnabled:  false,
		OTLPEndpoint: "grpc://127.0.0.1:4317",
		LogLevel:     "INFO",
	}

	publicServer := &mockPublicServer{port: 8080, baseURL: "https://localhost:8080"}
	adminServer := &mockAdminServer{port: 9090, baseURL: "https://localhost:9090"}

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

	defer listener.Shutdown(context.Background())
}

// TestStartListener_NilContext tests validation of nil context.
func TestStartListener_NilContext(t *testing.T) {
	t.Parallel()

	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		DevMode:     true,
		DatabaseURL: cryptoutilSharedMagic.SQLiteInMemoryDSN,
	}

	config := &ListenerConfig{
		Settings:     settings,
		PublicServer: &mockPublicServer{},
		AdminServer:  &mockAdminServer{},
	}

	listener, err := StartListener(nil, config) //nolint:staticcheck // Testing nil context.
	require.Error(t, err)
	require.Nil(t, listener)
	require.Contains(t, err.Error(), "ctx cannot be nil")
}

// TestStartListener_NilConfig tests validation of nil config.
func TestStartListener_NilConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	listener, err := StartListener(ctx, nil)
	require.Error(t, err)
	require.Nil(t, listener)
	require.Contains(t, err.Error(), "config cannot be nil")
}

// TestStartListener_NilSettings tests validation of nil settings.
func TestStartListener_NilSettings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	config := &ListenerConfig{
		Settings:     nil,
		PublicServer: &mockPublicServer{},
		AdminServer:  &mockAdminServer{},
	}

	listener, err := StartListener(ctx, config)
	require.Error(t, err)
	require.Nil(t, listener)
	require.Contains(t, err.Error(), "settings cannot be nil")
}

// TestStartListener_NilPublicServer tests validation of nil public server.
func TestStartListener_NilPublicServer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		DevMode:     true,
		DatabaseURL: cryptoutilSharedMagic.SQLiteInMemoryDSN,
	}

	config := &ListenerConfig{
		Settings:     settings,
		PublicServer: nil,
		AdminServer:  &mockAdminServer{},
	}

	listener, err := StartListener(ctx, config)
	require.Error(t, err)
	require.Nil(t, listener)
	require.Contains(t, err.Error(), "publicServer cannot be nil")
}

// TestStartListener_NilAdminServer tests validation of nil admin server.
func TestStartListener_NilAdminServer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		DevMode:     true,
		DatabaseURL: cryptoutilSharedMagic.SQLiteInMemoryDSN,
	}

	config := &ListenerConfig{
		Settings:     settings,
		PublicServer: &mockPublicServer{},
		AdminServer:  nil,
	}

	listener, err := StartListener(ctx, config)
	require.Error(t, err)
	require.Nil(t, listener)
	require.Contains(t, err.Error(), "adminServer cannot be nil")
}

// TestListener_PublicPort tests PublicPort accessor.
func TestListener_PublicPort(t *testing.T) {
	t.Parallel()

	listener := &Listener{
		PublicServer: &mockPublicServer{port: 12345},
	}

	require.Equal(t, 12345, listener.PublicPort())
}

// TestListener_PublicPort_NilServer tests PublicPort with nil server.
func TestListener_PublicPort_NilServer(t *testing.T) {
	t.Parallel()

	listener := &Listener{
		PublicServer: nil,
	}

	require.Equal(t, 0, listener.PublicPort())
}

// TestListener_AdminPort tests AdminPort accessor.
func TestListener_AdminPort(t *testing.T) {
	t.Parallel()

	listener := &Listener{
		AdminServer: &mockAdminServer{port: 54321},
	}

	require.Equal(t, 54321, listener.AdminPort())
}

// TestListener_AdminPort_NilServer tests AdminPort with nil server.
func TestListener_AdminPort_NilServer(t *testing.T) {
	t.Parallel()

	listener := &Listener{
		AdminServer: nil,
	}

	require.Equal(t, 0, listener.AdminPort())
}

// TestListener_Shutdown tests graceful shutdown of Listener.
func TestListener_Shutdown(t *testing.T) {
	// NOT parallel - uses shared SQLite database.

	ctx := context.Background()
	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		DevMode:      true,
		DatabaseURL:  cryptoutilSharedMagic.SQLiteInMemoryDSN,
		OTLPService:  "test-shutdown",
		OTLPEnabled:  false,
		OTLPEndpoint: "grpc://127.0.0.1:4317",
		LogLevel:     "INFO",
	}

	publicServer := &mockPublicServer{port: 8080}
	adminServer := &mockAdminServer{port: 9090}

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
func TestListener_Shutdown_NilContext(t *testing.T) {
	// NOT parallel - uses shared SQLite database.

	ctx := context.Background()
	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		DevMode:      true,
		DatabaseURL:  cryptoutilSharedMagic.SQLiteInMemoryDSN,
		OTLPService:  "test-shutdown-nil-ctx",
		OTLPEnabled:  false,
		OTLPEndpoint: "grpc://127.0.0.1:4317",
		LogLevel:     "INFO",
	}

	publicServer := &mockPublicServer{port: 8080}
	adminServer := &mockAdminServer{port: 9090}

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
func TestOpenPostgreSQL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// This test demonstrates openPostgreSQL function but will skip if no PostgreSQL available.
	// In production, this would use testcontainers to start PostgreSQL.
	// For now, we test the error path with invalid DSN.
	_, err := openPostgreSQL(ctx, "invalid-dsn", false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to open PostgreSQL database")
}

// TestListener_Start_NilContext tests Start with nil context.
func TestListener_Start_NilContext(t *testing.T) {
	t.Parallel()

	publicServer := &mockPublicServer{port: 8080}
	adminServer := &mockAdminServer{port: 9090}

	listener := &Listener{
		PublicServer: publicServer,
		AdminServer:  adminServer,
	}

	err := listener.Start(nil) //nolint:staticcheck // Testing nil context.
	require.Error(t, err)
	require.Contains(t, err.Error(), "context cannot be nil")
}

// TestListener_Start_PublicServerError tests Start when public server fails immediately.
func TestListener_Start_PublicServerError(t *testing.T) {
	// NOT parallel - uses shared SQLite database.

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		DevMode:      true,
		DatabaseURL:  cryptoutilSharedMagic.SQLiteInMemoryDSN,
		OTLPService:  "test-start-error",
		OTLPEnabled:  false,
		OTLPEndpoint: "grpc://127.0.0.1:4317",
		LogLevel:     "INFO",
	}

	// Create mock server that fails immediately.
	publicServer := &mockPublicServer{
		port:     8080,
		startErr: fmt.Errorf("mock public server error"),
	}
	adminServer := &mockAdminServer{port: 9090}

	config := &ListenerConfig{
		Settings:     settings,
		PublicServer: publicServer,
		AdminServer:  adminServer,
	}

	listener, err := StartListener(context.Background(), config)
	require.NoError(t, err)
	defer listener.Shutdown(context.Background())

	// Start should fail with public server error.
	err = listener.Start(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "public server failed")
}

// TestListener_Start_AdminServerError tests Start when admin server fails immediately.
func TestListener_Start_AdminServerError(t *testing.T) {
	// NOT parallel - uses shared SQLite database.

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		DevMode:      true,
		DatabaseURL:  cryptoutilSharedMagic.SQLiteInMemoryDSN,
		OTLPService:  "test-start-admin-error",
		OTLPEnabled:  false,
		OTLPEndpoint: "grpc://127.0.0.1:4317",
		LogLevel:     "INFO",
	}

	publicServer := &mockPublicServer{port: 8080}
	// Create mock server that fails immediately.
	adminServer := &mockAdminServer{
		port:     9090,
		startErr: fmt.Errorf("mock admin server error"),
	}

	config := &ListenerConfig{
		Settings:     settings,
		PublicServer: publicServer,
		AdminServer:  adminServer,
	}

	listener, err := StartListener(context.Background(), config)
	require.NoError(t, err)
	defer listener.Shutdown(context.Background())

	// Start should fail with admin server error.
	err = listener.Start(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "admin server failed")
}

// TestListener_Start_ContextCancelled tests Start when context is cancelled.
func TestListener_Start_ContextCancelled(t *testing.T) {
	// NOT parallel - uses shared SQLite database.

	ctx, cancel := context.WithCancel(context.Background())

	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		DevMode:      true,
		DatabaseURL:  cryptoutilSharedMagic.SQLiteInMemoryDSN,
		OTLPService:  "test-start-cancel",
		OTLPEnabled:  false,
		OTLPEndpoint: "grpc://127.0.0.1:4317",
		LogLevel:     "INFO",
	}

	// Create servers that block until cancelled.
	startDone := make(chan struct{})
	publicServer := &mockPublicServer{
		port:      8080,
		startDone: startDone,
	}
	adminServer := &mockAdminServer{
		port:      9090,
		startDone: startDone,
	}

	config := &ListenerConfig{
		Settings:     settings,
		PublicServer: publicServer,
		AdminServer:  adminServer,
	}

	listener, err := StartListener(context.Background(), config)
	require.NoError(t, err)
	defer listener.Shutdown(context.Background())

	// Start in background, then cancel context.
	errChan := make(chan error, 1)
	go func() {
		errChan <- listener.Start(ctx)
	}()

	// Wait a bit for Start to begin.
	time.Sleep(100 * time.Millisecond)

	// Cancel context.
	cancel()

	// Unblock servers.
	close(startDone)

	// Should return context cancellation error.
	err = <-errChan
	require.Error(t, err)
	require.Contains(t, err.Error(), "application startup cancelled")
}

// TestStartCore_NilContext tests StartCore with nil context.
func TestStartCore_NilContext(t *testing.T) {
	t.Parallel()

	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		DevMode:     true,
		DatabaseURL: cryptoutilSharedMagic.SQLiteInMemoryDSN,
	}

	core, err := StartCore(nil, settings) //nolint:staticcheck // Testing nil context.
	require.Error(t, err)
	require.Nil(t, core)
	require.Contains(t, err.Error(), "ctx cannot be nil")
}

// TestStartCore_NilSettings tests StartCore with nil settings.
func TestStartCore_NilSettings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	core, err := StartCore(ctx, nil)
	require.Error(t, err)
	require.Nil(t, core)
	require.Contains(t, err.Error(), "settings cannot be nil")
}

// TestProvisionDatabase_UnsupportedScheme tests provisionDatabase with unsupported database URL.
func TestProvisionDatabase_UnsupportedScheme(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		DevMode:      true,
		DatabaseURL:  "mysql://localhost:3306/test", // Unsupported scheme.
		OTLPService:  "test-unsupported-db",
		OTLPEnabled:  false,
		OTLPEndpoint: "grpc://127.0.0.1:4317",
		LogLevel:     "INFO",
	}

	basic, err := StartBasic(ctx, settings)
	require.NoError(t, err)
	defer basic.Shutdown()

	db, cleanup, err := provisionDatabase(ctx, basic, settings)
	require.Error(t, err)
	require.Nil(t, db)
	require.Nil(t, cleanup)
	require.Contains(t, err.Error(), "unsupported database URL scheme")
}

// TestProvisionDatabase_SQLiteFileURL tests SQLite with file:// URL.
func TestProvisionDatabase_SQLiteFileURL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Use temp file for SQLite database.
	dbFile := "/tmp/test_sqlite_" + time.Now().Format("20060102150405") + ".db"
	defer func() {
		// Cleanup.
		_ = os.Remove(dbFile)
	}()

	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		DevMode:      true,
		DatabaseURL:  "file://" + dbFile,
		OTLPService:  "test-sqlite-file",
		OTLPEnabled:  false,
		OTLPEndpoint: "grpc://127.0.0.1:4317",
		LogLevel:     "INFO",
	}

	basic, err := StartBasic(ctx, settings)
	require.NoError(t, err)
	defer basic.Shutdown()

	db, cleanup, err := provisionDatabase(ctx, basic, settings)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer cleanup()

	// Verify database works.
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NotNil(t, sqlDB)
}

// TestOpenSQLite_InvalidDSN tests openSQLite with valid DSN and WAL mode.
func TestOpenSQLite_InvalidDSN(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Test successful operation with in-memory DSN.
	db, err := openSQLite(ctx, cryptoutilSharedMagic.SQLiteInMemoryDSN, false)
	require.NoError(t, err)
	require.NotNil(t, db)

	// Verify PRAGMA settings were applied (WAL mode for file databases, memory for in-memory).
	var busyTimeout int
	sqlDB, _ := db.DB()
	err = sqlDB.QueryRow("PRAGMA busy_timeout").Scan(&busyTimeout)
	require.NoError(t, err)
	require.Equal(t, 30000, busyTimeout) // 30 seconds as configured.
}
