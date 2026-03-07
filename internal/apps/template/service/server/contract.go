// Copyright (c) 2025 Justin Cranford
//
//

// Package server provides reusable server infrastructure for cryptoutil services.
package server

import (
	"context"

	"gorm.io/gorm"
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
}
