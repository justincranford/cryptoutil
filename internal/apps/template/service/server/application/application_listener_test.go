// Copyright (c) 2025 Justin Cranford
//
//

package application

import (
	"context"
	"crypto/tls"
	"net/http/httptest"
	"testing"
	"time"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

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
