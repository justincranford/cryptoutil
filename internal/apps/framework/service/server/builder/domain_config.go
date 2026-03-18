// Copyright (c) 2025 Justin Cranford
//

// Package builder provides a simplified API for constructing service infrastructure.
package builder

import (
	"context"
	"fmt"
	"io/fs"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps/framework/service/config"
	cryptoutilAppsFrameworkServiceServer "cryptoutil/internal/apps/framework/service/server"
)

// DomainConfig contains domain-specific configuration for a service.
// Pass to Build() to construct service infrastructure in <=10 lines per service.
type DomainConfig struct {
	// MigrationsFS contains domain-specific migrations (nil = no domain migrations).
	MigrationsFS fs.FS

	// MigrationsPath is the path within MigrationsFS (e.g., "migrations").
	// Required when MigrationsFS is non-nil.
	MigrationsPath string

	// RouteRegistration is the domain-specific route setup callback.
	// Receives initialized PublicServerBase and ServiceResources.
	// nil is valid (skeleton services with no domain routes).
	RouteRegistration func(*cryptoutilAppsFrameworkServiceServer.PublicServerBase, *ServiceResources) error
}

// Build constructs a complete service infrastructure using the domain config.
// Replaces the NewServerBuilder+WithDomainMigrations+WithPublicRouteRegistration chain.
// Returns an error if ctx or settings is nil, or if Build() fails.
func Build(
	ctx context.Context,
	settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings,
	domain *DomainConfig,
) (*ServiceResources, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}

	if settings == nil {
		return nil, fmt.Errorf("settings cannot be nil")
	}

	b := NewServerBuilder(ctx, settings)

	if domain != nil {
		if domain.MigrationsFS != nil {
			b.WithDomainMigrations(domain.MigrationsFS, domain.MigrationsPath)
		}

		if domain.RouteRegistration != nil {
			b.WithPublicRouteRegistration(domain.RouteRegistration)
		}
	}

	resources, err := b.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build service: %w", err)
	}

	return resources, nil
}
