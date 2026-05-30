// Copyright (c) 2025-2026 Justin Cranford.
//
//

// Package server provides the KMS server using the template's ServerBuilder.
package server

import (
	"context"
	"crypto/x509"
	"fmt"
	"sync/atomic"

	"gorm.io/gorm"

	cryptoutilKmsServer "cryptoutil/api/sm-kms/server"
	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilAppsFrameworkServiceServer "cryptoutil/internal/apps-framework/service/server"
	cryptoutilAppsFrameworkServiceServerBarrier "cryptoutil/internal/apps-framework/service/server/barrier"
	cryptoutilAppsFrameworkServiceServerBuilder "cryptoutil/internal/apps-framework/service/server/builder"
	cryptoutilAppsFrameworkServiceServerMiddleware "cryptoutil/internal/apps-framework/service/server/middleware"
	cryptoutilKmsServerBusinesslogic "cryptoutil/internal/apps/sm-kms/server/businesslogic"
	cryptoutilAppsSmKmsServerHandler "cryptoutil/internal/apps/sm-kms/server/handler"
	cryptoutilAppsSmKmsRepository "cryptoutil/internal/apps/sm-kms/server/repository"
	cryptoutilOrmRepository "cryptoutil/internal/apps/sm-kms/server/repository/orm"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
)

// KMSServer wraps the template's ServerBuilder infrastructure with KMS-specific services.
type KMSServer struct {
	settings  *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings
	resources *cryptoutilAppsFrameworkServiceServerBuilder.ServiceResources
	ready     atomic.Bool
}

// NewKMSServerFromConfig creates a new KMS server using the template's ServerBuilder.
// All KMS-specific services (OrmRepository, BusinessLogicService) are created inside the RouteRegistration callback using builder-provided resources.
func NewKMSServerFromConfig(
	ctx context.Context,
	settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings,
) (*KMSServer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}

	if settings == nil {
		return nil, fmt.Errorf("settings cannot be nil")
	}

	resources, err := cryptoutilAppsFrameworkServiceServerBuilder.Build(ctx, settings, &cryptoutilAppsFrameworkServiceServerBuilder.DomainConfig{
		MigrationsFS:   cryptoutilAppsSmKmsRepository.MigrationsFS,
		MigrationsPath: "migrations",
		RouteRegistration: func(publicServerBase *cryptoutilAppsFrameworkServiceServer.PublicServerBase, res *cryptoutilAppsFrameworkServiceServerBuilder.ServiceResources) error {
			ormRepo, err := cryptoutilOrmRepository.NewOrmRepository(ctx, res.TelemetryService, res.DB, res.JWKGenService, settings.VerboseMode)
			if err != nil {
				return fmt.Errorf("failed to create orm repository: %w", err)
			}

			bizLogicService, err := cryptoutilKmsServerBusinesslogic.NewBusinessLogicService(ctx, res.TelemetryService, res.JWKGenService, ormRepo, res.BarrierService)
			if err != nil {
				return fmt.Errorf("failed to create business logic service: %w", err)
			}

			return registerKMSRoutes(publicServerBase.App(), bizLogicService, settings, res)
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build KMS server: %w", err)
	}

	return &KMSServer{
		settings:  settings,
		resources: resources,
	}, nil
}

// tenantContextBridgeMiddleware copies tenant/realm IDs from Fiber locals
// (set by template SessionMiddleware) into the Go context as a RealmContext,
// so KMS business logic can continue to use GetRealmContext(ctx).
func tenantContextBridgeMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tid, ok := c.Locals(cryptoutilAppsFrameworkServiceServerMiddleware.ContextKeyTenantID).(googleUuid.UUID)
		if ok && tid != googleUuid.Nil {
			rid, _ := c.Locals(cryptoutilAppsFrameworkServiceServerMiddleware.ContextKeyRealmID).(googleUuid.UUID)
			rc := &cryptoutilAppsFrameworkServiceServerMiddleware.RealmContext{
				TenantID: tid,
				RealmID:  rid,
				Source:   "session",
			}
			c.SetUserContext(context.WithValue(c.UserContext(), cryptoutilAppsFrameworkServiceServerMiddleware.RealmContextKey{}, rc))
		}

		return c.Next()
	}
}

// registerKMSRoutes registers KMS-specific routes on the public Fiber app.
func registerKMSRoutes(
	app *fiber.App,
	bizLogicService *cryptoutilKmsServerBusinesslogic.BusinessLogicService,
	settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings,
	res *cryptoutilAppsFrameworkServiceServerBuilder.ServiceResources,
) error {
	// Create the OpenAPI strict server handler.
	openapiStrictServer := cryptoutilAppsSmKmsServerHandler.NewOpenapiStrictServer(bizLogicService)
	openapiStrictHandler := cryptoutilKmsServer.NewStrictHandler(openapiStrictServer, nil)

	// Create route groups so middleware is scoped to each prefix.
	// Using app.Use() would apply middleware globally to all routes on the app.
	// Using group.Use() scopes middleware only to routes registered on that group.
	browserGroup := app.Group(settings.PublicBrowserAPIContextPath)
	serviceGroup := app.Group(settings.PublicServiceAPIContextPath)

	bridgeMW := fiber.Handler(tenantContextBridgeMiddleware())

	if res != nil && res.SessionManager != nil {
		browserSessionMW := cryptoutilAppsFrameworkServiceServerMiddleware.BrowserSessionMiddleware(res.SessionManager)
		serviceSessionMW := cryptoutilAppsFrameworkServiceServerMiddleware.ServiceSessionMiddleware(res.SessionManager)

		browserGroup.Use(browserSessionMW, bridgeMW)
		serviceGroup.Use(serviceSessionMW, bridgeMW)
	}

	// Register handlers on browser and service groups with empty BaseURL
	// (the group already carries the base path prefix).
	groupOpts := cryptoutilKmsServer.FiberServerOptions{}
	cryptoutilKmsServer.RegisterHandlersWithOptions(browserGroup, openapiStrictHandler, groupOpts)
	cryptoutilKmsServer.RegisterHandlersWithOptions(serviceGroup, openapiStrictHandler, groupOpts)

	// Compatibility routes for consolidated sm-kms and sm-kms APIs.
	if res != nil && res.DB != nil && res.BarrierService != nil {
		elasticRepo := cryptoutilAppsSmKmsRepository.NewElasticJWKRepository(res.DB)
		materialRepo := cryptoutilAppsSmKmsRepository.NewMaterialJWKRepository(res.DB)
		messageRepo := cryptoutilAppsSmKmsRepository.NewMessageRepository(res.DB)
		messageRecipientRepo := cryptoutilAppsSmKmsRepository.NewMessageRecipientJWKRepository(res.DB, res.BarrierService)

		jwkCompatHandler := cryptoutilAppsSmKmsServerHandler.NewJWKCompatHandler(elasticRepo, materialRepo)
		jwksCompatHandler := cryptoutilAppsSmKmsServerHandler.NewJWKSCompatHandler(elasticRepo, materialRepo)
		messageCompatHandler := cryptoutilAppsSmKmsServerHandler.NewMessageHandler(messageRepo, messageRecipientRepo)

		registerCompatRoutes := func(group fiber.Router) {
			group.Get("/jwks", jwksCompatHandler.HandleGetJWKS())
			group.Get("/elastic-keys/:elasticKeyID/material-keys/active", jwkCompatHandler.HandleGetActiveMaterialJWK())
			group.Post("/elastic-keys/:elasticKeyID/rotate", jwkCompatHandler.HandleRotateMaterialJWK())

			group.Get("/messages", messageCompatHandler.HandleListMessages())
			group.Get("/messages/:messageID", messageCompatHandler.HandleGetMessage())
			group.Delete("/messages/:messageID", messageCompatHandler.HandleDeleteMessage())
			group.Post("/messages/send", messageCompatHandler.HandleSendMessage())
			group.Get("/messages/receive", messageCompatHandler.HandleReceiveMessages())
		}

		registerCompatRoutes(browserGroup)
		registerCompatRoutes(serviceGroup)
	}

	return nil
}

// Start starts the KMS server.
func (s *KMSServer) Start(ctx context.Context) error {
	if s.resources == nil || s.resources.Application == nil {
		return fmt.Errorf("server not initialized")
	}

	s.ready.Store(true)
	s.resources.Application.SetReady(true)

	if err := s.resources.Application.Start(ctx); err != nil {
		return fmt.Errorf("failed to start KMS server: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the KMS server.
func (s *KMSServer) Shutdown(ctx context.Context) error {
	s.ready.Store(false)

	// Shutdown server infrastructure.
	if s.resources != nil {
		if s.resources.Application != nil {
			_ = s.resources.Application.Shutdown(ctx)
		}

		if s.resources.ShutdownCore != nil {
			s.resources.ShutdownCore()
		}

		if s.resources.ShutdownContainer != nil {
			s.resources.ShutdownContainer()
		}
	}

	return nil
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
func (s *KMSServer) Resources() *cryptoutilAppsFrameworkServiceServerBuilder.ServiceResources {
	return s.resources
}

// Settings returns the server settings.
func (s *KMSServer) Settings() *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings {
	return s.settings
}

// SetReady marks the server as ready (enables /admin/api/v1/readyz to return 200 OK).
func (s *KMSServer) SetReady(ready bool) {
	if s.resources != nil && s.resources.Application != nil {
		s.resources.Application.SetReady(ready)
	}

	s.ready.Store(ready)
}

// DB returns the GORM database connection (for tests).
func (s *KMSServer) DB() *gorm.DB {
	if s.resources != nil {
		return s.resources.DB
	}

	return nil
}

// App returns the application wrapper (for tests).
func (s *KMSServer) App() *cryptoutilAppsFrameworkServiceServer.Application {
	if s.resources != nil {
		return s.resources.Application
	}

	return nil
}

// TLSRootCAPool returns the root CA pool for test client TLS configuration (public server).
func (s *KMSServer) TLSRootCAPool() *x509.CertPool {
	if s.resources != nil && s.resources.Application != nil {
		return s.resources.Application.TLSRootCAPool()
	}

	return nil
}

// AdminTLSRootCAPool returns the admin TLS root CA pool for test client TLS configuration.
func (s *KMSServer) AdminTLSRootCAPool() *x509.CertPool {
	if s.resources != nil && s.resources.Application != nil {
		return s.resources.Application.AdminTLSRootCAPool()
	}

	return nil
}

// JWKGen returns the JWK generation service used by this server.
func (s *KMSServer) JWKGen() *cryptoutilSharedCryptoJose.JWKGenService {
	if s.resources != nil {
		return s.resources.JWKGenService
	}

	return nil
}

// Telemetry returns the telemetry service used by this server.
func (s *KMSServer) Telemetry() *cryptoutilSharedTelemetry.TelemetryService {
	if s.resources != nil {
		return s.resources.TelemetryService
	}

	return nil
}

// Barrier returns the barrier (encryption-at-rest) service used by this server.
func (s *KMSServer) Barrier() *cryptoutilAppsFrameworkServiceServerBarrier.Service {
	if s.resources != nil {
		return s.resources.BarrierService
	}

	return nil
}

// Compile-time assertion: KMSServer must implement ServiceServer.
var _ cryptoutilAppsFrameworkServiceServer.ServiceServer = (*KMSServer)(nil)
