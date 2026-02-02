// Copyright (c) 2025 Justin Cranford
//
//

// Package server provides the KMS server infrastructure using ServerBuilder.
package server

import (
	"context"
	"fmt"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServerBuilder "cryptoutil/internal/apps/template/service/server/builder"
)

// KMSBuilderAdapterSettings contains KMS-specific configuration for the builder adapter.
type KMSBuilderAdapterSettings struct {
	// JWKSURL is the URL for fetching JWT public keys for authentication.
	JWKSURL string

	// JWTIssuer is the expected issuer claim in JWTs.
	JWTIssuer string

	// JWTAudience is the expected audience claim in JWTs.
	JWTAudience string
}

// Validate validates the KMS-specific settings.
func (s *KMSBuilderAdapterSettings) Validate() error {
	// All fields are optional - KMS can operate without JWT authentication.
	return nil
}

// KMSBuilderAdapter configures ServerBuilder for KMS-specific requirements.
// KMS differs from template services in these ways:
// - Single-tenant (no multi-tenancy)
// - No barrier service (uses separate encryption)
// - Domain-only migrations (no template migrations)
// - JWT authentication via external JWKS (optional)
// - OpenAPI strict server pattern.
type KMSBuilderAdapter struct {
	ctx         context.Context
	settings    *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings
	kmsSettings *KMSBuilderAdapterSettings
}

// NewKMSBuilderAdapter creates a new adapter for KMS ServerBuilder configuration.
func NewKMSBuilderAdapter(
	ctx context.Context,
	settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings,
	kmsSettings *KMSBuilderAdapterSettings,
) (*KMSBuilderAdapter, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}

	if settings == nil {
		return nil, fmt.Errorf("settings cannot be nil")
	}

	// Use empty KMS settings if none provided.
	if kmsSettings == nil {
		kmsSettings = &KMSBuilderAdapterSettings{}
	}

	if err := kmsSettings.Validate(); err != nil {
		return nil, fmt.Errorf("invalid KMS settings: %w", err)
	}

	return &KMSBuilderAdapter{
		ctx:         ctx,
		settings:    settings,
		kmsSettings: kmsSettings,
	}, nil
}

// ConfigureBuilder configures a ServerBuilder with KMS-specific settings.
// Returns the configured builder ready for Build().
func (a *KMSBuilderAdapter) ConfigureBuilder() *cryptoutilAppsTemplateServiceServerBuilder.ServerBuilder {
	builder := cryptoutilAppsTemplateServiceServerBuilder.NewServerBuilder(a.ctx, a.settings)

	// Configure barrier as disabled (KMS uses separate encryption mechanism).
	builder.WithBarrierConfig(cryptoutilAppsTemplateServiceServerBuilder.NewDisabledBarrierConfig())

	// Configure domain-only migrations (KMS doesn't use template migrations).
	// NOTE: KMS has its own migration system in repository/sqlrepository, so we disable migrations entirely.
	builder.WithMigrationConfig(cryptoutilAppsTemplateServiceServerBuilder.NewDisabledMigrationConfig())

	// Configure JWT authentication if JWKS URL is provided.
	if a.kmsSettings.JWKSURL != "" {
		jwtConfig := cryptoutilAppsTemplateServiceServerBuilder.NewKMSJWTAuthConfig(
			a.kmsSettings.JWKSURL,
			a.kmsSettings.JWTIssuer,
			a.kmsSettings.JWTAudience,
		)
		builder.WithJWTAuth(jwtConfig)
	} else {
		// No JWT authentication - KMS uses its own middleware.
		builder.WithJWTAuth(cryptoutilAppsTemplateServiceServerBuilder.NewDefaultJWTAuthConfig())
	}

	// Configure OpenAPI strict server registration.
	strictConfig := cryptoutilAppsTemplateServiceServerBuilder.NewDefaultStrictServerConfig().
		WithBrowserBasePath(a.settings.PublicBrowserAPIContextPath).
		WithServiceBasePath(a.settings.PublicServiceAPIContextPath)

	builder.WithStrictServer(strictConfig)

	return builder
}

// Context returns the adapter's context.
func (a *KMSBuilderAdapter) Context() context.Context {
	return a.ctx
}

// Settings returns the adapter's settings.
func (a *KMSBuilderAdapter) Settings() *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings {
	return a.settings
}

// KMSSettings returns the adapter's KMS-specific settings.
func (a *KMSBuilderAdapter) KMSSettings() *KMSBuilderAdapterSettings {
	return a.kmsSettings
}
