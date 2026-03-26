// Copyright (c) 2025 Justin Cranford
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
	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps/framework/service/config"
	cryptoutilAppsFrameworkServiceServer "cryptoutil/internal/apps/framework/service/server"
	cryptoutilAppsFrameworkServiceServerBarrier "cryptoutil/internal/apps/framework/service/server/barrier"
	cryptoutilAppsFrameworkServiceServerBuilder "cryptoutil/internal/apps/framework/service/server/builder"
	cryptoutilAppsFrameworkServiceServerMiddleware "cryptoutil/internal/apps/framework/service/server/middleware"
	cryptoutilKmsServerBusinesslogic "cryptoutil/internal/apps/sm-kms/server/businesslogic"
	cryptoutilKmsServerHandler "cryptoutil/internal/apps/sm-kms/server/handler"
	cryptoutilKmsServerMiddleware "cryptoutil/internal/apps/sm-kms/server/middleware"
	cryptoutilAppsSmKmsServerRepository "cryptoutil/internal/apps/sm-kms/server/repository"
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

// NewKMSServer creates a new KMS server using the template's ServerBuilder.
// All KMS-specific services (OrmRepository, BusinessLogicService) are created inside the RouteRegistration callback using builder-provided resources.
func NewKMSServer(
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
		MigrationsFS:   cryptoutilAppsSmKmsServerRepository.MigrationsFS,
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
			rc := &cryptoutilKmsServerMiddleware.RealmContext{
				TenantID: tid,
				RealmID:  rid,
				Source:   "session",
			}
			c.SetUserContext(context.WithValue(c.UserContext(), cryptoutilKmsServerMiddleware.RealmContextKey{}, rc))
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
	openapiStrictServer := cryptoutilKmsServerHandler.NewOpenapiStrictServer(bizLogicService)
	openapiStrictHandler := cryptoutilKmsServer.NewStrictHandler(openapiStrictServer, nil)

	// Build middleware chains: session auth + tenant context bridge.
	// When res is nil (unit tests), routes are registered without auth middleware.
	bridgeMW := cryptoutilKmsServer.MiddlewareFunc(tenantContextBridgeMiddleware())

	var (
		browserMiddlewares []cryptoutilKmsServer.MiddlewareFunc
		serviceMiddlewares []cryptoutilKmsServer.MiddlewareFunc
	)

	if res != nil && res.SessionManager != nil {
		browserSessionMW := cryptoutilAppsFrameworkServiceServerMiddleware.BrowserSessionMiddleware(res.SessionManager)
		serviceSessionMW := cryptoutilAppsFrameworkServiceServerMiddleware.ServiceSessionMiddleware(res.SessionManager)
		browserMiddlewares = []cryptoutilKmsServer.MiddlewareFunc{cryptoutilKmsServer.MiddlewareFunc(browserSessionMW), bridgeMW}
		serviceMiddlewares = []cryptoutilKmsServer.MiddlewareFunc{cryptoutilKmsServer.MiddlewareFunc(serviceSessionMW), bridgeMW}
	}

	// Configure browser API options.
	publicBrowserFiberServerOptions := cryptoutilKmsServer.FiberServerOptions{
		BaseURL:     settings.PublicBrowserAPIContextPath,
		Middlewares: browserMiddlewares,
	}

	// Configure service API options.
	publicServiceFiberServerOptions := cryptoutilKmsServer.FiberServerOptions{
		BaseURL:     settings.PublicServiceAPIContextPath,
		Middlewares: serviceMiddlewares,
	}

	// Register handlers on both browser and service paths.
	cryptoutilKmsServer.RegisterHandlersWithOptions(app, openapiStrictHandler, publicBrowserFiberServerOptions)
	cryptoutilKmsServer.RegisterHandlersWithOptions(app, openapiStrictHandler, publicServiceFiberServerOptions)

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

// PublicServerActualPort returns the actual port the public server is listening on.
// Alias for PublicPort() — both return the same value.
func (s *KMSServer) PublicServerActualPort() int {
	return s.PublicPort()
}

// AdminServerActualPort returns the actual port the admin server is listening on.
// Alias for AdminPort() — both return the same value.
func (s *KMSServer) AdminServerActualPort() int {
	return s.AdminPort()
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
