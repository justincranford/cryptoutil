// Copyright (c) 2025 Justin Cranford
//
//

// Package listener provides high-level application lifecycle management.
//
// This package encapsulates the complete service startup pattern used across
// all cryptoutil services (sm-im, jose-ja, identity-*, sm-kms, pki-ca, skeleton-template).
//
// The ApplicationListener provides a unified interface for:
// - Starting full service with telemetry, database, barrier, public/admin servers
// - Health checks (liveness, readiness)
// - Graceful shutdown
//
// Usage pattern (TestMain):
//
//	listener, err := StartApplicationListener(ctx, cfg, db, dbType, handlers)
//	if err != nil {
//	    panic(err)
//	}
//	defer listener.Shutdown()
//	os.Exit(m.Run())
package listener

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"gorm.io/gorm"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilNetwork "cryptoutil/internal/shared/util/network"
)

// ApplicationListener encapsulates complete service lifecycle (telemetry, DB, servers, shutdown).
//
// Provides unified interface matching sm-kms pattern:
// - StartApplicationListener: Complete service initialization
// - SendLivenessCheck: Lightweight health check
// - SendReadinessCheck: Heavyweight dependencies check
// - SendShutdownRequest: Graceful shutdown via API
// - Shutdown: Direct shutdown for cleanup.
type ApplicationListener struct {
	app               *cryptoutilAppsTemplateServiceServer.Application
	config            *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings
	shutdownFunc      func()
	actualPublicPort  uint16
	actualPrivatePort uint16

	// TLS configuration (for client requests to server).
	PublicTLSServer  *TLSServerConfig
	PrivateTLSServer *TLSServerConfig
}

// TLSServerConfig holds TLS certificate and CA pools for server verification.
// Matches sm-kms pattern for compatibility.
type TLSServerConfig struct {
	Certificate         *tls.Certificate
	RootCAsPool         *x509.CertPool
	IntermediateCAsPool *x509.CertPool
	Config              *tls.Config
}

// HandlerRegistration is a function that registers routes on IPublicServer.
// Allows product-specific business logic injection without hardcoding.
//
// Example sm-im usage:
//
//	func RegisterSmIMHandlers(server cryptoutilTemplateServer.IPublicServer, userRepo, messageRepo) error {
//	    app := server.(*PublicServer).App() // Type assertion to access fiber.App
//	    app.Post("/api/v1/messages", handleSendMessage)
//	    return nil
//	}
type HandlerRegistration func(server cryptoutilAppsTemplateServiceServer.IPublicServer) error

// PublicServerFactory creates a product-specific public server.
//
// Each service (sm-im, jose-ja, identity-*, sm-kms, pki-ca, skeleton-template) provides its own factory
// that knows how to construct the service's unique public server with appropriate:
// - Repositories
// - Business logic handlers
// - OpenAPI specifications
// - Authentication/authorization middleware.
//
// Example sm-im factory:
//
//	func NewPublicServerFromConfig(
//	    ctx context.Context,
//	    cfg *ApplicationConfig,
//	    template *cryptoutilTemplateServer.ServiceTemplate,
//	) (cryptoutilTemplateServer.IPublicServer, error) {
//	    // Create repositories
//	    userRepo := repository.NewUserRepository(cfg.DB)
//	    messageRepo := repository.NewMessageRepository(cfg.DB)
//
//	    // Create public server
//	    return NewPublicServer(ctx, cfg, template, userRepo, messageRepo)
//	}
type PublicServerFactory func(
	ctx context.Context,
	cfg *ApplicationConfig,
	template *cryptoutilAppsTemplateServiceServer.ServiceTemplate,
) (cryptoutilAppsTemplateServiceServer.IPublicServer, error)

// ApplicationConfig encapsulates all product-service specific configuration.
//
// This is the injection point for product-specific:
// - ServiceTemplateServerSettings (bind addresses, ports, TLS, OTLP)
// - Database connection and type
// - Public server factory (product-specific server creation)
// - Handler registration (OpenAPI, business logic routes)
// - Optional barrier service configuration.
type ApplicationConfig struct {
	// ServiceTemplateServerSettings contains common settings (bind addresses, TLS, OTLP, etc.).
	ServiceTemplateServerSettings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings

	// Database connection (managed externally, e.g., test-container or production pool).
	DB *gorm.DB

	// DatabaseType determines migration strategy and SQL dialect.
	DBType cryptoutilAppsTemplateServiceServerRepository.DatabaseType

	// PublicServerFactory creates the product-specific public server.
	// REQUIRED: Each service must provide its own factory function.
	PublicServerFactory PublicServerFactory

	// PublicHandlers registers additional product-specific routes on public server (optional).
	// Called after server creation but before Start().
	// Most services embed routes in their server constructor, making this optional.
	PublicHandlers HandlerRegistration

	// AdminHandlers registers product-specific admin routes (optional).
	// Examples: barrier rotation, custom diagnostics.
	AdminHandlers HandlerRegistration
}

// StartApplicationListener creates and starts a full service application.
//
// This is the primary entry point for all cryptoutil services (sm-im, jose-ja, identity-*, sm-kms, pki-ca, skeleton-template).
//
// Initialization sequence:
// 1. Create ServiceTemplate (telemetry, JWK gen, optional barrier)
// 2. Create public server (bind address, TLS, handlers)
// 3. Create admin server (always 127.0.0.1:9090, health checks)
// 4. Create Application (manages both servers)
// 5. Start servers in background goroutines
// 6. Return listener for health checks and shutdown.
//
// Parameters:
// - ctx: Context for initialization (must not be nil)
// - cfg: Product-service configuration (ServiceTemplateServerSettings, DB, handlers)
//
// Returns:
// - *ApplicationListener: Running service ready for requests
// - error: Non-nil if initialization or startup fails.
//
// Example sm-im usage:
//
//	cfg := &ApplicationConfig{
//	    ServiceTemplateServerSettings: cryptoutilConfig.NewTestConfig("127.0.0.1", 0, true),
//	    DB: testDB,
//	    DBType: cryptoutilTemplateServerRepository.DatabaseTypeSQLite,
//	    PublicHandlers: func(srv cryptoutilTemplateServer.IPublicServer) error {
//	        // Register sm-im routes
//	        return nil
//	    },
//	}
//	listener, err := StartApplicationListener(ctx, cfg)
func StartApplicationListener(ctx context.Context, cfg *ApplicationConfig) (*ApplicationListener, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	} else if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	} else if cfg.ServiceTemplateServerSettings == nil {
		return nil, fmt.Errorf("config.ServiceTemplateServerSettings cannot be nil")
	} else if cfg.DB == nil {
		return nil, fmt.Errorf("config.DB cannot be nil")
	} else if cfg.PublicServerFactory == nil {
		return nil, fmt.Errorf("config.PublicServerFactory cannot be nil (each service must provide factory)")
	}

	// Create ServiceTemplate with shared infrastructure.
	// This initializes telemetry, JWK generation, and optionally barrier service.
	template, err := cryptoutilAppsTemplateServiceServer.NewServiceTemplate(ctx, cfg.ServiceTemplateServerSettings, cfg.DB, cfg.DBType)
	if err != nil {
		return nil, fmt.Errorf("failed to create service template: %w", err)
	}

	// TODO: Create public server (product-specific implementation will inject handlers).
	// For now, return error indicating implementation needed.
	// Each product service (sm-im, jose-ja, etc.) will need to provide:
	// - Public server constructor (NewPublicServer)
	// - Handler registration via cfg.PublicHandlers
	//
	// Example pattern (to be implemented per-service):
	// publicServer, err := NewPublicServer(ctx, cfg, template)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to create public server: %w", err)
	// }
	//
	// cfg.PublicHandlers(publicServer) // Inject product-specific routes

	// TODO: Create admin server (reusable across all services).
	// adminServer, err := cryptoutilTemplateServerListener.NewAdminHTTPServer(ctx, cfg.ServiceTemplateServerSettings, adminTLSCfg)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to create admin server: %w", err)
	// }
	//
	// if cfg.AdminHandlers != nil {
	//     cfg.AdminHandlers(adminServer) // Inject optional admin routes (barrier rotation, etc.)
	// }

	// TODO: Create Application and start servers.
	// app, err := cryptoutilTemplateServer.NewApplication(ctx, publicServer, adminServer)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to create application: %w", err)
	// }

	// Placeholder shutdown function (will include telemetry, JWK gen, servers).
	shutdownFunc := func() {
		template.Shutdown()
	}

	return &ApplicationListener{
		app:          nil, // TODO: Populate after Application creation
		config:       cfg.ServiceTemplateServerSettings,
		shutdownFunc: shutdownFunc,
		// TODO: Extract actual ports from started servers
		actualPublicPort:  0,
		actualPrivatePort: 0,
	}, fmt.Errorf("StartApplicationListener: implementation in progress (product-specific server creation needed)")
}

// SendLivenessCheck performs lightweight health check (process alive?).
//
// This endpoint should respond quickly (<100ms) without heavyweight operations.
// Used by Kubernetes liveness probes to detect deadlocks or hangs.
//
// Failure action: Restart container.
//
// Returns:
// - []byte: Response body (usually "OK" or JSON status)
// - error: Non-nil if request fails or times out.
func SendLivenessCheck(settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.ClientLivenessRequestTimeout)
	defer cancel()

	_, _, result, err := cryptoutilSharedUtilNetwork.HTTPGetLivez(ctx, settings.PrivateBaseURL(), settings.PrivateAdminAPIContextPath, 0, nil, settings.DevMode)
	if err != nil {
		return nil, fmt.Errorf("failed to get liveness check: %w", err)
	}

	return result, nil
}

// SendReadinessCheck performs heavyweight health check (dependencies healthy?).
//
// This endpoint checks database connectivity, dependent services, critical resources.
// Used by Kubernetes readiness probes to control load balancer traffic.
//
// Failure action: Remove from load balancer (do NOT restart).
//
// Returns:
// - []byte: Response body with dependency status details
// - error: Non-nil if request fails or dependencies unhealthy.
func SendReadinessCheck(settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.ClientReadinessRequestTimeout)
	defer cancel()

	_, _, result, err := cryptoutilSharedUtilNetwork.HTTPGetReadyz(ctx, settings.PrivateBaseURL(), settings.PrivateAdminAPIContextPath, 0, nil, settings.DevMode)
	if err != nil {
		return nil, fmt.Errorf("failed to get readiness check: %w", err)
	}

	return result, nil
}

// SendShutdownRequest triggers graceful shutdown via admin API.
//
// This is the recommended way to stop a service (vs. SIGTERM/SIGKILL).
// Allows service to:
// - Stop accepting new requests
// - Drain in-flight requests (up to 30s)
// - Close database connections
// - Flush telemetry buffers
// - Release resources cleanly.
//
// Returns:
// - error: Non-nil if shutdown request fails to send (service may already be down).
func SendShutdownRequest(settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) error {
	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.ClientShutdownRequestTimeout)
	defer cancel()

	_, _, _, err := cryptoutilSharedUtilNetwork.HTTPPostShutdown(ctx, settings.PrivateBaseURL(), settings.PrivateAdminAPIContextPath, 0, nil, settings.DevMode)
	if err != nil {
		return fmt.Errorf("failed to send shutdown request: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the service.
//
// Shutdown sequence:
// 1. Stop accepting new requests (admin server first, then public)
// 2. Wait for in-flight requests to complete (up to 30s)
// 3. Shutdown services (barrier, JWK gen, telemetry)
// 4. Close database connections
//
// This method is idempotent (safe to call multiple times).
func (l *ApplicationListener) Shutdown() {
	if l.shutdownFunc != nil {
		l.shutdownFunc()
	}

	if l.app != nil {
		ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultShutdownTimeout)
		defer cancel()

		_ = l.app.Shutdown(ctx) // Best-effort shutdown, ignore errors during cleanup.
	}
}

// ActualPublicPort returns the actual public server port after dynamic allocation.
//
// Useful for testing with port 0 (OS assigns random available port).
// Returns 0 if server not yet started.
func (l *ApplicationListener) ActualPublicPort() uint16 {
	return l.actualPublicPort
}

// ActualPrivatePort returns the actual admin server port (should always be 9090 in production).
//
// Returns dynamic port for tests using port 0.
// Returns 0 if server not yet started.
func (l *ApplicationListener) ActualPrivatePort() uint16 {
	return l.actualPrivatePort
}

// Config returns the server settings used to configure this listener.
// Useful for health checks and client connections.
func (l *ApplicationListener) Config() *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings {
	return l.config
}

// SetApplicationForTesting sets the application field for testing purposes.
// This method enables testing the shutdown path when app != nil.
// ONLY USE IN TESTS - not for production code.
func (l *ApplicationListener) SetApplicationForTesting(app *cryptoutilAppsTemplateServiceServer.Application) {
	l.app = app
}
