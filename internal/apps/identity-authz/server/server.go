// Copyright (c) 2025 Justin Cranford

// Package server implements the identity-authz HTTPS server using the service template.
package server

import (
	"context"
	"crypto/x509"
	"fmt"

	"gorm.io/gorm"

	cryptoutilAppsFrameworkServiceServer "cryptoutil/internal/apps/framework/service/server"
	cryptoutilAppsFrameworkServiceServerBarrier "cryptoutil/internal/apps/framework/service/server/barrier"
	cryptoutilAppsFrameworkServiceServerBuilder "cryptoutil/internal/apps/framework/service/server/builder"
	cryptoutilAppsFrameworkServiceServerBusinesslogic "cryptoutil/internal/apps/framework/service/server/businesslogic"
	cryptoutilAppsFrameworkServiceServerRepository "cryptoutil/internal/apps/framework/service/server/repository"
	cryptoutilAppsFrameworkServiceServerService "cryptoutil/internal/apps/framework/service/server/service"
	cryptoutilAppsIdentityAuthzServerConfig "cryptoutil/internal/apps/identity-authz/server/config"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

// AuthzServer represents the identity-authz service application.
// This is an OAuth 2.1 Authorization Server with OIDC Discovery support.
type AuthzServer struct {
	app *cryptoutilAppsFrameworkServiceServer.Application
	db  *gorm.DB

	// Authz configuration.
	cfg *cryptoutilAppsIdentityAuthzServerConfig.IdentityAuthzServerSettings

	// Template services.
	telemetryService      *cryptoutilSharedTelemetry.TelemetryService
	jwkGenService         *cryptoutilSharedCryptoJose.JWKGenService
	barrierService        *cryptoutilAppsFrameworkServiceServerBarrier.Service
	sessionManagerService *cryptoutilAppsFrameworkServiceServerBusinesslogic.SessionManagerService
	realmService          cryptoutilAppsFrameworkServiceServerService.RealmService

	// Template repositories.
	realmRepo cryptoutilAppsFrameworkServiceServerRepository.TenantRealmRepository

	// Shutdown functions.
	shutdownCore      func()
	shutdownContainer func()
}

// NewFromConfig creates a new identity-authz server from IdentityAuthzServerSettings.
// Uses service-framework builder for infrastructure initialization.
func NewFromConfig(ctx context.Context, cfg *cryptoutilAppsIdentityAuthzServerConfig.IdentityAuthzServerSettings) (*AuthzServer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	} else if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	resources, err := cryptoutilAppsFrameworkServiceServerBuilder.Build(ctx, cfg.ServiceFrameworkServerSettings, &cryptoutilAppsFrameworkServiceServerBuilder.DomainConfig{
		RouteRegistration: func(base *cryptoutilAppsFrameworkServiceServer.PublicServerBase, _ *cryptoutilAppsFrameworkServiceServerBuilder.ServiceResources) error {
			// Create public server with authz handlers.
			publicServer := NewPublicServer(base, cfg)

			// Register all routes (standard + authz-specific).
			if err := publicServer.registerRoutes(); err != nil {
				return fmt.Errorf("failed to register public routes: %w", err)
			}

			return nil
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build identity-authz service: %w", err)
	}

	// Create identity-authz server wrapper.
	server := &AuthzServer{
		app:                   resources.Application,
		db:                    resources.DB,
		cfg:                   cfg,
		telemetryService:      resources.TelemetryService,
		jwkGenService:         resources.JWKGenService,
		barrierService:        resources.BarrierService,
		sessionManagerService: resources.SessionManager,
		realmService:          resources.RealmService,
		realmRepo:             resources.RealmRepository,
		shutdownCore:          resources.ShutdownCore,
		shutdownContainer:     resources.ShutdownContainer,
	}

	return server, nil
}

// Start begins serving both public and admin HTTPS endpoints.
// Blocks until context is cancelled or an unrecoverable error occurs.
func (s *AuthzServer) Start(ctx context.Context) error {
	if err := s.app.Start(ctx); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down all servers and closes database connections.
func (s *AuthzServer) Shutdown(ctx context.Context) error {
	if err := s.app.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown application: %w", err)
	}

	// Invoke shutdown callbacks.
	if s.shutdownCore != nil {
		s.shutdownCore()
	}

	if s.shutdownContainer != nil {
		s.shutdownContainer()
	}

	return nil
}

// Config returns the server configuration (for tests).
func (s *AuthzServer) Config() *cryptoutilAppsIdentityAuthzServerConfig.IdentityAuthzServerSettings {
	return s.cfg
}

// DB returns the GORM database connection (for tests).
func (s *AuthzServer) DB() *gorm.DB {
	return s.db
}

// App returns the application wrapper (for tests).
func (s *AuthzServer) App() *cryptoutilAppsFrameworkServiceServer.Application {
	return s.app
}

// JWKGen returns the JWK generation service (for tests).
func (s *AuthzServer) JWKGen() *cryptoutilSharedCryptoJose.JWKGenService {
	return s.jwkGenService
}

// Telemetry returns the telemetry service (for tests).
func (s *AuthzServer) Telemetry() *cryptoutilSharedTelemetry.TelemetryService {
	return s.telemetryService
}

// Barrier returns the barrier service (for tests).
func (s *AuthzServer) Barrier() *cryptoutilAppsFrameworkServiceServerBarrier.Service {
	return s.barrierService
}

// PublicPort returns the actual port the public server is listening on (for tests).
// Useful when configured with port 0 for dynamic allocation.
func (s *AuthzServer) PublicPort() int {
	return s.app.PublicPort()
}

// AdminPort returns the actual port the admin server is listening on (for tests).
// Useful when configured with port 0 for dynamic allocation.
func (s *AuthzServer) AdminPort() int {
	return s.app.AdminPort()
}

// SetReady marks the server as ready (enables /admin/api/v1/readyz to return 200 OK).
func (s *AuthzServer) SetReady(ready bool) {
	s.app.SetReady(ready)
}

// PublicBaseURL returns the public server base URL (for tests).
func (s *AuthzServer) PublicBaseURL() string {
	return s.app.PublicBaseURL()
}

// AdminBaseURL returns the admin server base URL (for tests).
func (s *AuthzServer) AdminBaseURL() string {
	return s.app.AdminBaseURL()
}

// PublicServerActualPort returns the actual port the public server is listening on.
// Alias for PublicPort() — both return the same value.
func (s *AuthzServer) PublicServerActualPort() int {
	return s.app.PublicPort()
}

// AdminServerActualPort returns the actual port the admin server is listening on.
// Alias for AdminPort() — both return the same value.
func (s *AuthzServer) AdminServerActualPort() int {
	return s.app.AdminPort()
}

// TLSRootCAPool returns the root CA pool for test client TLS configuration (public server).
func (s *AuthzServer) TLSRootCAPool() *x509.CertPool {
	return s.app.TLSRootCAPool()
}

// AdminTLSRootCAPool returns the admin TLS root CA pool for test client TLS configuration.
func (s *AuthzServer) AdminTLSRootCAPool() *x509.CertPool {
	return s.app.AdminTLSRootCAPool()
}

// Compile-time assertion: AuthzServer must implement ServiceServer.
var _ cryptoutilAppsFrameworkServiceServer.ServiceServer = (*AuthzServer)(nil)
