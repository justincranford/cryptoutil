// Copyright (c) 2025 Justin Cranford
//
//

// Package server provides reusable server infrastructure for cryptoutil services.
package server

import (
	"context"
	"crypto/x509"

	"gorm.io/gorm"

	cryptoutilAppsFrameworkServiceServerBarrier "cryptoutil/internal/apps/framework/service/server/barrier"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

// ServiceServer defines the contract that all cryptoutil services must satisfy.
// All 10 services (sm-im, sm-kms, jose-ja, pki-ca, skeleton-template, identity-authz/idp/rp/rs/spa)
// must implement this interface for compile-time framework conformance.
//
// Usage: var _ ServiceServer = (*XxxServer)(nil).
type ServiceServer interface {
	// Start begins serving public and admin HTTPS endpoints.
	// Blocks until context is cancelled or an unrecoverable error occurs.
	Start(ctx context.Context) error

	// Shutdown gracefully shuts down all servers and closes database connections.
	Shutdown(ctx context.Context) error

	// DB returns the GORM database connection (primarily for tests).
	DB() *gorm.DB

	// App returns the underlying Application wrapper (primarily for tests).
	App() *Application

	// PublicPort returns the actual port the public server is listening on.
	// Useful when configured with port 0 for dynamic allocation.
	PublicPort() int

	// AdminPort returns the actual port the admin server is listening on.
	// Useful when configured with port 0 for dynamic allocation.
	AdminPort() int

	// SetReady marks the server as ready (enables /admin/api/v1/readyz to return 200 OK).
	SetReady(ready bool)

	// PublicBaseURL returns the base URL for the public server (e.g. https://127.0.0.1:8080).
	PublicBaseURL() string

	// AdminBaseURL returns the base URL for the admin server (e.g. https://127.0.0.1:9090).
	AdminBaseURL() string

	// PublicServerActualPort returns the actual port the public server is listening on.
	// Alias for PublicPort() — both return the same value.
	PublicServerActualPort() int

	// AdminServerActualPort returns the actual port the admin server is listening on.
	// Alias for AdminPort() — both return the same value.
	AdminServerActualPort() int

	// TLSRootCAPool returns the root CA certificate pool for the server's TLS chain.
	// Used by test infrastructure to configure secure HTTP clients without InsecureSkipVerify.
	// Returns nil when the server has not yet started or uses a non-PublicServerBase implementation.
	TLSRootCAPool() *x509.CertPool

	// AdminTLSRootCAPool returns the root CA certificate pool for the admin server's TLS chain.
	// Used by test infrastructure to configure secure HTTP clients for admin endpoints without InsecureSkipVerify.
	// Returns nil when the admin server has not yet started.
	AdminTLSRootCAPool() *x509.CertPool

	// JWKGen returns the JWK generation service used by this server.
	// Used by integration tests to verify JWK generation is correctly wired.
	JWKGen() *cryptoutilSharedCryptoJose.JWKGenService

	// Telemetry returns the telemetry service used by this server.
	// Used by integration tests to verify OpenTelemetry is correctly wired.
	Telemetry() *cryptoutilSharedTelemetry.TelemetryService

	// Barrier returns the barrier (encryption-at-rest) service used by this server.
	// Used by integration tests to verify the barrier service is correctly wired.
	Barrier() *cryptoutilAppsFrameworkServiceServerBarrier.Service
}
