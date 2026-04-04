// Copyright (c) 2025 Justin Cranford
//
//

package application

import (
	"context"
	"database/sql"
	"fmt"

	"gorm.io/gorm"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps/framework/service/config"
	cryptoutilAppsFrameworkServiceServerBarrier "cryptoutil/internal/apps/framework/service/server/barrier"
	cryptoutilUnsealKeysService "cryptoutil/internal/apps/framework/service/server/barrier/unsealkeysservice"
	cryptoutilAppsFrameworkServiceServerBusinesslogic "cryptoutil/internal/apps/framework/service/server/businesslogic"
	cryptoutilAppsFrameworkServiceServerRepository "cryptoutil/internal/apps/framework/service/server/repository"
	cryptoutilAppsFrameworkServiceServerService "cryptoutil/internal/apps/framework/service/server/service"
	cryptoutilSharedContainer "cryptoutil/internal/shared/container"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

// Core extends Basic with database infrastructure.
// Handles automatic database provisioning (SQLite in-memory, PostgreSQL testcontainer, or external DB).
type Core struct {
	Basic               *Basic
	DB                  *gorm.DB
	ShutdownDBContainer func()
	Settings            *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings
}

// StartCoreWithServices initializes core application infrastructure AND all business services.
// This is the proper place for service bootstrap logic (not in ServerBuilder).
// Phase W: Moved from ServerBuilder.Build() to encapsulate bootstrap logic in application layer.
func StartCoreWithServices(ctx context.Context, settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings) (*CoreWithServices, error) {
	// Start core infrastructure (telemetry, JWK gen, unseal, database).
	core, err := StartCore(ctx, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to start application core: %w", err)
	}

	// Initialize services on top of core infrastructure.
	return InitializeServicesOnCore(ctx, core, settings)
}

// StartCore initializes core application infrastructure including database.
// Automatically provisions database based on settings.DatabaseURL and settings.DatabaseContainer:
// - SQLite in-memory: DatabaseURL="file::memory:?cache=shared"
// - PostgreSQL testcontainer: DatabaseURL empty + DatabaseContainer="required"/"preferred"
// - External DB: DatabaseURL with postgres:// or file:// scheme.
func StartCore(ctx context.Context, settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings) (*Core, error) {
	return startCoreInternal(ctx, settings,
		cryptoutilSharedCryptoJose.NewJWKGenService,
		cryptoutilSharedContainer.StartPostgres,
		sql.Open,
		func(dialector gorm.Dialector, config *gorm.Config) (*gorm.DB, error) {
			return gorm.Open(dialector, config)
		},
	)
}

func startCoreInternal(
	ctx context.Context,
	settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings,
	newJWKGenServiceFn func(ctx context.Context, telemetryService *cryptoutilSharedTelemetry.TelemetryService, devMode bool) (*cryptoutilSharedCryptoJose.JWKGenService, error),
	startPostgresFn func(ctx context.Context, telemetryService *cryptoutilSharedTelemetry.TelemetryService, dbName, username, password string) (string, func(), error),
	sqlOpenFn func(driverName, dataSourceName string) (*sql.DB, error),
	gormOpenSQLiteFn func(dialector gorm.Dialector, config *gorm.Config) (*gorm.DB, error),
) (*Core, error) {
	if ctx == nil {
		return nil, fmt.Errorf("ctx cannot be nil")
	} else if settings == nil {
		return nil, fmt.Errorf("settings cannot be nil")
	}

	// Start basic infrastructure.
	basic, err := startBasicInternal(ctx, settings, newJWKGenServiceFn)
	if err != nil {
		return nil, fmt.Errorf("failed to start basic application: %w", err)
	}

	core := &Core{
		Basic:               basic,
		ShutdownDBContainer: func() {}, // No-op by default.
		Settings:            settings,
	}

	// Provision database based on DatabaseURL and DatabaseContainer settings.
	db, shutdownContainer, err := provisionDatabaseInternal(ctx, basic, settings, startPostgresFn, sqlOpenFn, gormOpenSQLiteFn)
	if err != nil {
		basic.TelemetryService.Slogger.Error("failed to provision database", cryptoutilSharedMagic.StringError, err)
		core.Shutdown()

		return nil, fmt.Errorf("failed to provision database: %w", err)
	}

	core.DB = db
	core.ShutdownDBContainer = shutdownContainer

	return core, nil
}

// CoreWithServices extends Core with initialized business services.
// This struct encapsulates ALL service bootstrap logic previously scattered in ServerBuilder.
type CoreWithServices struct {
	Core *Core

	// Repositories.
	Repository            *cryptoutilAppsFrameworkServiceServerBarrier.GormRepository
	RealmRepository       cryptoutilAppsFrameworkServiceServerRepository.TenantRealmRepository
	TenantRepository      cryptoutilAppsFrameworkServiceServerRepository.TenantRepository
	UserRepository        cryptoutilAppsFrameworkServiceServerRepository.UserRepository
	JoinRequestRepository cryptoutilAppsFrameworkServiceServerRepository.TenantJoinRequestRepository

	// Services.
	BarrierService      *cryptoutilAppsFrameworkServiceServerBarrier.Service
	RealmService        cryptoutilAppsFrameworkServiceServerService.RealmService
	SessionManager      *cryptoutilAppsFrameworkServiceServerBusinesslogic.SessionManagerService
	RegistrationService *cryptoutilAppsFrameworkServiceServerBusinesslogic.TenantRegistrationService
	RotationService     *cryptoutilAppsFrameworkServiceServerBarrier.RotationService
	StatusService       *cryptoutilAppsFrameworkServiceServerBarrier.StatusService
}

// InitializeServicesOnCore initializes all business logic services on an existing Core.
// CRITICAL: This must be called AFTER migrations have been applied (BarrierService needs barrier_root_keys table).
func InitializeServicesOnCore(
	ctx context.Context,
	core *Core,
	settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings,
) (*CoreWithServices, error) {
	return initializeServicesOnCoreInternal(ctx, core, settings,
		cryptoutilAppsFrameworkServiceServerBarrier.NewGormRepository,
		cryptoutilAppsFrameworkServiceServerBusinesslogic.NewSessionManagerService,
		cryptoutilAppsFrameworkServiceServerBarrier.NewRotationService,
		cryptoutilAppsFrameworkServiceServerBarrier.NewStatusService,
	)
}

func initializeServicesOnCoreInternal(
	ctx context.Context,
	core *Core,
	settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings,
	newBarrierGormRepositoryFn func(db *gorm.DB) (*cryptoutilAppsFrameworkServiceServerBarrier.GormRepository, error),
	newSessionManagerServiceFn func(ctx context.Context, db *gorm.DB, telemetryService *cryptoutilSharedTelemetry.TelemetryService, jwkGenService *cryptoutilSharedCryptoJose.JWKGenService, barrierService *cryptoutilAppsFrameworkServiceServerBarrier.Service, settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings) (*cryptoutilAppsFrameworkServiceServerBusinesslogic.SessionManagerService, error),
	newRotationServiceFn func(jwkGenService *cryptoutilSharedCryptoJose.JWKGenService, repository cryptoutilAppsFrameworkServiceServerBarrier.Repository, unsealKeysService cryptoutilUnsealKeysService.UnsealKeysService) (*cryptoutilAppsFrameworkServiceServerBarrier.RotationService, error),
	newStatusServiceFn func(repository cryptoutilAppsFrameworkServiceServerBarrier.Repository) (*cryptoutilAppsFrameworkServiceServerBarrier.StatusService, error),
) (*CoreWithServices, error) {
	if core == nil {
		return nil, fmt.Errorf("core is nil")
	}

	services := &CoreWithServices{
		Core: core,
	}

	// Create barrier repository and service.
	barrierRepo, err := newBarrierGormRepositoryFn(core.DB)
	if err != nil {
		core.Shutdown()

		return nil, fmt.Errorf("failed to create barrier repository: %w", err)
	}

	services.Repository = barrierRepo

	barrierService, err := cryptoutilAppsFrameworkServiceServerBarrier.NewService(
		ctx,
		core.Basic.TelemetryService,
		core.Basic.JWKGenService,
		barrierRepo,
		core.Basic.UnsealKeysService,
	)
	if err != nil {
		core.Shutdown()

		return nil, fmt.Errorf("failed to create barrier service: %w", err)
	}

	services.BarrierService = barrierService

	// Create realm repository and service.
	realmRepo := cryptoutilAppsFrameworkServiceServerRepository.NewTenantRealmRepository(core.DB)
	services.RealmRepository = realmRepo

	realmService := cryptoutilAppsFrameworkServiceServerService.NewRealmService(realmRepo)
	services.RealmService = realmService

	// Create session manager service.
	sessionManager, err := newSessionManagerServiceFn(
		ctx,
		core.DB,
		core.Basic.TelemetryService,
		core.Basic.JWKGenService,
		barrierService,
		settings,
	)
	if err != nil {
		core.Shutdown()

		return nil, fmt.Errorf("failed to create session manager service: %w", err)
	}

	services.SessionManager = sessionManager

	// Create tenant registration service and dependencies.
	tenantRepo := cryptoutilAppsFrameworkServiceServerRepository.NewTenantRepository(core.DB)
	services.TenantRepository = tenantRepo

	userRepo := cryptoutilAppsFrameworkServiceServerRepository.NewUserRepository(core.DB)
	services.UserRepository = userRepo

	joinRequestRepo := cryptoutilAppsFrameworkServiceServerRepository.NewTenantJoinRequestRepository(core.DB)
	services.JoinRequestRepository = joinRequestRepo

	registrationService := cryptoutilAppsFrameworkServiceServerBusinesslogic.NewTenantRegistrationService(
		core.DB,
		tenantRepo,
		userRepo,
		joinRequestRepo,
	)
	services.RegistrationService = registrationService

	// Create barrier rotation service.
	rotationService, err := newRotationServiceFn(
		core.Basic.JWKGenService,
		barrierRepo,
		core.Basic.UnsealKeysService,
	)
	if err != nil {
		core.Shutdown()

		return nil, fmt.Errorf("failed to create rotation service: %w", err)
	}

	services.RotationService = rotationService

	// Create barrier status service.
	statusService, err := newStatusServiceFn(barrierRepo)
	if err != nil {
		core.Shutdown()

		return nil, fmt.Errorf("failed to create status service: %w", err)
	}

	services.StatusService = statusService

	return services, nil
}

// Shutdown gracefully shuts down all core services (LIFO order).
func (a *Core) Shutdown() {
	if a.Basic != nil && a.Basic.TelemetryService != nil {
		a.Basic.TelemetryService.Slogger.Debug("stopping application core")
	}

	// Shutdown database container (if any).
	if a.ShutdownDBContainer != nil {
		a.ShutdownDBContainer()
	}

	// Close database connection.
	if a.DB != nil {
		sqlDB, err := a.DB.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	}

	// Shutdown basic infrastructure.
	if a.Basic != nil {
		a.Basic.Shutdown()
	}
}
