// Copyright (c) 2025 Justin Cranford
//

// Package server implements the pki-ca HTTPS server using the service template.
package server

import (
	"context"
	"crypto/x509"
	"fmt"

	"gorm.io/gorm"

	cryptoutilAppsCaServerConfig "cryptoutil/internal/apps/pki/ca/server/config"
	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilAppsTemplateServiceServerBuilder "cryptoutil/internal/apps/template/service/server/builder"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

// PKICAServer wraps the service template Application for the pki-ca service.
type PKICAServer struct {
	app       *cryptoutilAppsTemplateServiceServer.Application
	db        *gorm.DB
	resources *cryptoutilAppsTemplateServiceServerBuilder.ServiceResources
}

// NewFromConfig creates a new PKI CA server from configuration.
func NewFromConfig(ctx context.Context, cfg *cryptoutilAppsCaServerConfig.PKICAServerSettings) (*PKICAServer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("failed to create pki-ca server: %w", fmt.Errorf("context is nil"))
	}

	if cfg == nil {
		return nil, fmt.Errorf("failed to create pki-ca server: %w", fmt.Errorf("config is nil"))
	}

	resources, err := cryptoutilAppsTemplateServiceServerBuilder.Build(ctx, cfg.ServiceTemplateServerSettings, &cryptoutilAppsTemplateServiceServerBuilder.DomainConfig{})
	if err != nil {
		return nil, fmt.Errorf("failed to build pki-ca server: %w", err)
	}

	return &PKICAServer{app: resources.Application, db: resources.DB, resources: resources}, nil
}

// Start starts the pki-ca server (blocking).
func (s *PKICAServer) Start(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("failed to start pki-ca server: %w", fmt.Errorf("context is nil"))
	}

	if err := s.app.Start(ctx); err != nil {
		return fmt.Errorf("failed to start pki-ca server: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the pki-ca server.
func (s *PKICAServer) Shutdown(ctx context.Context) error {
	if err := s.app.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown pki-ca server: %w", err)
	}

	return nil
}

// PublicPort returns the actual port the public server is listening on (for tests).
// Useful when configured with port 0 for dynamic allocation.
func (s *PKICAServer) PublicPort() int {
	return s.app.PublicPort()
}

// AdminPort returns the actual port the admin server is listening on (for tests).
// Useful when configured with port 0 for dynamic allocation.
func (s *PKICAServer) AdminPort() int {
	return s.app.AdminPort()
}

// SetReady marks the server as ready (enables /admin/v1/readyz to return 200 OK).
func (s *PKICAServer) SetReady(ready bool) {
	s.app.SetReady(ready)
}

// PublicBaseURL returns the public server base URL (for tests).
func (s *PKICAServer) PublicBaseURL() string {
	return s.app.PublicBaseURL()
}

// AdminBaseURL returns the admin server base URL (for tests).
func (s *PKICAServer) AdminBaseURL() string {
	return s.app.AdminBaseURL()
}

// PublicServerActualPort returns the actual port the public server is listening on.
// Useful when configured with port 0 for dynamic allocation.
func (s *PKICAServer) PublicServerActualPort() int {
	return s.app.PublicPort()
}

// AdminServerActualPort returns the actual port the admin server is listening on.
// Useful when configured with port 0 for dynamic allocation.
func (s *PKICAServer) AdminServerActualPort() int {
	return s.app.AdminPort()
}

// DB returns the GORM database instance for domain operations.
func (s *PKICAServer) DB() *gorm.DB {
	return s.db
}

// App returns the application wrapper (for tests).
func (s *PKICAServer) App() *cryptoutilAppsTemplateServiceServer.Application {
	return s.app
}

// TLSRootCAPool returns the root CA pool for test client TLS configuration (public server).
func (s *PKICAServer) TLSRootCAPool() *x509.CertPool {
	return s.app.TLSRootCAPool()
}

// AdminTLSRootCAPool returns the admin TLS root CA pool for test client TLS configuration.
func (s *PKICAServer) AdminTLSRootCAPool() *x509.CertPool {
	return s.app.AdminTLSRootCAPool()
}

// JWKGen returns the JWK generation service used by this server.
func (s *PKICAServer) JWKGen() *cryptoutilSharedCryptoJose.JWKGenService {
	if s.resources != nil {
		return s.resources.JWKGenService
	}

	return nil
}

// Telemetry returns the telemetry service used by this server.
func (s *PKICAServer) Telemetry() *cryptoutilSharedTelemetry.TelemetryService {
	if s.resources != nil {
		return s.resources.TelemetryService
	}

	return nil
}

// Barrier returns the barrier (encryption-at-rest) service used by this server.
func (s *PKICAServer) Barrier() *cryptoutilAppsTemplateServiceServerBarrier.Service {
	if s.resources != nil {
		return s.resources.BarrierService
	}

	return nil
}

// Compile-time assertion: PKICAServer must implement ServiceServer.
var _ cryptoutilAppsTemplateServiceServer.ServiceServer = (*PKICAServer)(nil)
