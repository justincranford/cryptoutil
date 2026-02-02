// Copyright (c) 2025 Justin Cranford
//
//

// Package server provides the KMS server using the template's ServerBuilder.
package server

import (
	"context"
	"fmt"
	"sync/atomic"

	cryptoutilOpenapiServer "cryptoutil/api/server"
	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilAppsTemplateServiceServerBuilder "cryptoutil/internal/apps/template/service/server/builder"
	cryptoutilServerApplication "cryptoutil/internal/kms/server/application"
	cryptoutilKmsServerHandler "cryptoutil/internal/kms/server/handler"

	fiber "github.com/gofiber/fiber/v2"
)

// KMSServer wraps the template's ServerBuilder infrastructure with KMS-specific services.
type KMSServer struct {
	ctx       context.Context
	settings  *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings
	resources *cryptoutilAppsTemplateServiceServerBuilder.ServiceResources
	kmsCore   *cryptoutilServerApplication.ServerApplicationCore
	ready     atomic.Bool
}

// NewKMSServer creates a new KMS server using the template's ServerBuilder.
func NewKMSServer(
	ctx context.Context,
	settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings,
	kmsSettings *KMSBuilderAdapterSettings,
) (*KMSServer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}

	if settings == nil {
		return nil, fmt.Errorf("settings cannot be nil")
	}

	// Create the adapter to configure ServerBuilder for KMS.
	adapter, err := NewKMSBuilderAdapter(ctx, settings, kmsSettings)
	if err != nil {
		return nil, fmt.Errorf("failed to create KMS builder adapter: %w", err)
	}

	// Initialize KMS-specific services BEFORE building the server.
	// KMS has its own database setup (SQLRepository) and barrier (shared/barrier).
	kmsCore, err := cryptoutilServerApplication.StartServerApplicationCore(ctx, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to start KMS application core: %w", err)
	}

	// Configure and build the server using the adapter.
	builder := adapter.ConfigureBuilder()

	// Register KMS-specific routes.
	builder.WithPublicRouteRegistration(func(
		publicServerBase *cryptoutilAppsTemplateServiceServer.PublicServerBase,
		resources *cryptoutilAppsTemplateServiceServerBuilder.ServiceResources,
	) error {
		return registerKMSRoutes(publicServerBase.App(), kmsCore, settings)
	})

	// Build the server infrastructure (TLS, servers, middleware, health endpoints).
	resources, err := builder.Build()
	if err != nil {
		kmsCore.Shutdown()

		return nil, fmt.Errorf("failed to build KMS server: %w", err)
	}

	server := &KMSServer{
		ctx:       ctx,
		settings:  settings,
		resources: resources,
		kmsCore:   kmsCore,
	}

	return server, nil
}

// registerKMSRoutes registers KMS-specific routes on the public Fiber app.
func registerKMSRoutes(
	app *fiber.App,
	kmsCore *cryptoutilServerApplication.ServerApplicationCore,
	settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings,
) error {
	// Create the OpenAPI strict server handler.
	openapiStrictServer := cryptoutilKmsServerHandler.NewOpenapiStrictServer(kmsCore.BusinessLogicService)
	openapiStrictHandler := cryptoutilOpenapiServer.NewStrictHandler(openapiStrictServer, nil)

	// Configure browser API options.
	publicBrowserFiberServerOptions := cryptoutilOpenapiServer.FiberServerOptions{
		BaseURL: settings.PublicBrowserAPIContextPath,
	}

	// Configure service API options.
	publicServiceFiberServerOptions := cryptoutilOpenapiServer.FiberServerOptions{
		BaseURL: settings.PublicServiceAPIContextPath,
	}

	// Register handlers on both browser and service paths.
	cryptoutilOpenapiServer.RegisterHandlersWithOptions(app, openapiStrictHandler, publicBrowserFiberServerOptions)
	cryptoutilOpenapiServer.RegisterHandlersWithOptions(app, openapiStrictHandler, publicServiceFiberServerOptions)

	return nil
}

// Start starts the KMS server.
func (s *KMSServer) Start() error {
	if s.resources == nil || s.resources.Application == nil {
		return fmt.Errorf("server not initialized")
	}

	s.ready.Store(true)
	s.resources.Application.SetReady(true)

	if err := s.resources.Application.Start(s.ctx); err != nil {
		return fmt.Errorf("failed to start KMS server: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the KMS server.
func (s *KMSServer) Shutdown() {
	s.ready.Store(false)

	// Shutdown KMS-specific services.
	if s.kmsCore != nil {
		s.kmsCore.Shutdown()
	}

	// Shutdown server infrastructure.
	if s.resources != nil {
		if s.resources.Application != nil {
			_ = s.resources.Application.Shutdown(s.ctx)
		}

		if s.resources.ShutdownCore != nil {
			s.resources.ShutdownCore()
		}

		if s.resources.ShutdownContainer != nil {
			s.resources.ShutdownContainer()
		}
	}
}

// IsReady returns whether the server is ready to serve requests.
func (s *KMSServer) IsReady() bool {
	return s.ready.Load()
}

// PublicPort returns the actual public port the server is listening on.
func (s *KMSServer) PublicPort() int {
	if s.resources != nil && s.resources.Application != nil {
		return s.resources.Application.PublicPort()
	}

	return 0
}

// AdminPort returns the actual admin port the server is listening on.
func (s *KMSServer) AdminPort() int {
	if s.resources != nil && s.resources.Application != nil {
		return s.resources.Application.AdminPort()
	}

	return 0
}

// PublicBaseURL returns the base URL for the public server.
func (s *KMSServer) PublicBaseURL() string {
	if s.resources != nil && s.resources.Application != nil {
		return s.resources.Application.PublicBaseURL()
	}

	return ""
}

// AdminBaseURL returns the base URL for the admin server.
func (s *KMSServer) AdminBaseURL() string {
	if s.resources != nil && s.resources.Application != nil {
		return s.resources.Application.AdminBaseURL()
	}

	return ""
}

// Resources returns the service resources from ServerBuilder.
func (s *KMSServer) Resources() *cryptoutilAppsTemplateServiceServerBuilder.ServiceResources {
	return s.resources
}

// KMSCore returns the KMS application core.
func (s *KMSServer) KMSCore() *cryptoutilServerApplication.ServerApplicationCore {
	return s.kmsCore
}

// Settings returns the server settings.
func (s *KMSServer) Settings() *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings {
	return s.settings
}
