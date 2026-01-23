// Copyright (c) 2025 Justin Cranford
//
//

package application

import (
	"context"
	"database/sql"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilTemplateBusinessLogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilTemplateRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilTemplateService "cryptoutil/internal/apps/template/service/server/service"
	cryptoutilContainer "cryptoutil/internal/shared/container"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// ApplicationCore extends ApplicationBasic with database infrastructure.
// Handles automatic database provisioning (SQLite in-memory, PostgreSQL testcontainer, or external DB).
type ApplicationCore struct {
	Basic               *ApplicationBasic
	DB                  *gorm.DB
	ShutdownDBContainer func()
	Settings            *cryptoutilConfig.ServiceTemplateServerSettings
}

// StartApplicationCore initializes core application infrastructure including database.
// Automatically provisions database based on settings.DatabaseURL and settings.DatabaseContainer:
// - SQLite in-memory: DatabaseURL="file::memory:?cache=shared"
// - PostgreSQL testcontainer: DatabaseURL empty + DatabaseContainer="required"/"preferred"
// - External DB: DatabaseURL with postgres:// or file:// scheme.
func StartApplicationCore(ctx context.Context, settings *cryptoutilConfig.ServiceTemplateServerSettings) (*ApplicationCore, error) {
	if ctx == nil {
		return nil, fmt.Errorf("ctx cannot be nil")
	} else if settings == nil {
		return nil, fmt.Errorf("settings cannot be nil")
	}

	// Start basic infrastructure.
	basic, err := StartApplicationBasic(ctx, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to start basic application: %w", err)
	}

	core := &ApplicationCore{
		Basic:               basic,
		ShutdownDBContainer: func() {}, // No-op by default.
		Settings:            settings,
	}

	// Provision database based on DatabaseURL and DatabaseContainer settings.
	db, shutdownContainer, err := provisionDatabase(ctx, basic, settings)
	if err != nil {
		basic.TelemetryService.Slogger.Error("failed to provision database", "error", err)
		core.Shutdown()

		return nil, fmt.Errorf("failed to provision database: %w", err)
	}

	core.DB = db
	core.ShutdownDBContainer = shutdownContainer

	return core, nil
}

// ApplicationCoreWithServices extends ApplicationCore with initialized business services.
// This struct encapsulates ALL service bootstrap logic previously scattered in ServerBuilder.
type ApplicationCoreWithServices struct {
	Core *ApplicationCore

	// Repositories.
	BarrierRepository   *cryptoutilBarrier.GormBarrierRepository
	RealmRepository     cryptoutilTemplateRepository.TenantRealmRepository
	TenantRepository    cryptoutilTemplateRepository.TenantRepository
	UserRepository      cryptoutilTemplateRepository.UserRepository
	JoinRequestRepository cryptoutilTemplateRepository.TenantJoinRequestRepository

	// Services.
	BarrierService       *cryptoutilBarrier.BarrierService
	RealmService         cryptoutilTemplateService.RealmService
	SessionManager       *cryptoutilTemplateBusinessLogic.SessionManagerService
	RegistrationService  *cryptoutilTemplateBusinessLogic.TenantRegistrationService
	RotationService      *cryptoutilBarrier.RotationService
	StatusService        *cryptoutilBarrier.StatusService
}

// StartApplicationCoreWithServices initializes core application infrastructure AND all business services.
// This is the proper place for service bootstrap logic (not in ServerBuilder).
// Phase W: Moved from ServerBuilder.Build() to encapsulate bootstrap logic in application layer.
func StartApplicationCoreWithServices(ctx context.Context, settings *cryptoutilConfig.ServiceTemplateServerSettings) (*ApplicationCoreWithServices, error) {
	// Start core infrastructure (telemetry, JWK gen, unseal, database).
	core, err := StartApplicationCore(ctx, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to start application core: %w", err)
	}

	services := &ApplicationCoreWithServices{
		Core: core,
	}

	// Create barrier repository and service.
	barrierRepo, err := cryptoutilBarrier.NewGormBarrierRepository(core.DB)
	if err != nil {
		core.Shutdown()

		return nil, fmt.Errorf("failed to create barrier repository: %w", err)
	}

	services.BarrierRepository = barrierRepo

	barrierService, err := cryptoutilBarrier.NewBarrierService(
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
	realmRepo := cryptoutilTemplateRepository.NewTenantRealmRepository(core.DB)
	services.RealmRepository = realmRepo

	realmService := cryptoutilTemplateService.NewRealmService(realmRepo)
	services.RealmService = realmService

	// Create session manager service.
	sessionManager, err := cryptoutilTemplateBusinessLogic.NewSessionManagerService(
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
	tenantRepo := cryptoutilTemplateRepository.NewTenantRepository(core.DB)
	services.TenantRepository = tenantRepo

	userRepo := cryptoutilTemplateRepository.NewUserRepository(core.DB)
	services.UserRepository = userRepo

	joinRequestRepo := cryptoutilTemplateRepository.NewTenantJoinRequestRepository(core.DB)
	services.JoinRequestRepository = joinRequestRepo

	registrationService := cryptoutilTemplateBusinessLogic.NewTenantRegistrationService(
		core.DB,
		tenantRepo,
		userRepo,
		joinRequestRepo,
	)
	services.RegistrationService = registrationService

	// Create barrier rotation service.
	rotationService, err := cryptoutilBarrier.NewRotationService(
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
	statusService, err := cryptoutilBarrier.NewStatusService(barrierRepo)
	if err != nil {
		core.Shutdown()

		return nil, fmt.Errorf("failed to create status service: %w", err)
	}

	services.StatusService = statusService

	return services, nil
}

// Shutdown gracefully shuts down all core services (LIFO order).
func (a *ApplicationCore) Shutdown() {
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


// provisionDatabase handles all database provisioning scenarios:
// 1. Internal managed SQLite instance (file::memory:?cache=shared)
// 2. Internal managed PostgreSQL testcontainer (DatabaseContainer=required/preferred)
// 3. External DB connection (postgres:// or file:// scheme).
func provisionDatabase(ctx context.Context, basic *ApplicationBasic, settings *cryptoutilConfig.ServiceTemplateServerSettings) (*gorm.DB, func(), error) {
	databaseURL := settings.DatabaseURL
	containerMode := settings.DatabaseContainer

	shutdownContainer := func() {} // No-op by default.

	// Determine database type from URL.
	var isSQLite bool

	var isPostgres bool

	if databaseURL == "" || databaseURL == cryptoutilMagic.SQLiteInMemoryDSN || databaseURL == cryptoutilMagic.SQLiteMemoryPlaceholder {
		isSQLite = true
		databaseURL = cryptoutilMagic.SQLiteInMemoryDSN // Normalize SQLite in-memory URL.
	} else if len(databaseURL) >= 9 && databaseURL[:9] == "postgres:" {
		isPostgres = true
	} else if len(databaseURL) >= 7 && databaseURL[:7] == "file://" {
		isSQLite = true
	} else {
		return nil, nil, fmt.Errorf("unsupported database URL scheme: %s", databaseURL)
	}

	// Handle PostgreSQL testcontainer provisioning.
	if isPostgres && containerMode != "" && containerMode != "disabled" {
		basic.TelemetryService.Slogger.Debug("attempting to start PostgreSQL testcontainer", "containerMode", containerMode)

		containerURL, cleanup, err := cryptoutilContainer.StartPostgres(
			ctx,
			basic.TelemetryService,
			"test_db",
			"test_user",
			"test_password",
		)
		if err == nil {
			basic.TelemetryService.Slogger.Info("successfully started PostgreSQL testcontainer", "containerURL", containerURL)
			databaseURL = containerURL
			shutdownContainer = cleanup
		} else if containerMode == "required" {
			basic.TelemetryService.Slogger.Error("failed to start required PostgreSQL testcontainer", "error", err)

			return nil, nil, fmt.Errorf("failed to start required PostgreSQL testcontainer: %w", err)
		} else {
			basic.TelemetryService.Slogger.Warn("failed to start preferred PostgreSQL testcontainer, falling back to external DB", "error", err)
		}
	}

	// Open database connection.
	var db *gorm.DB

	var err error

	if isSQLite {
		basic.TelemetryService.Slogger.Debug("opening SQLite database", "url", databaseURL)
		db, err = openSQLite(ctx, databaseURL, settings.VerboseMode)
	} else if isPostgres {
		basic.TelemetryService.Slogger.Debug("opening PostgreSQL database", "url", maskPassword(databaseURL))
		db, err = openPostgreSQL(ctx, databaseURL, settings.VerboseMode)
	} else {
		return nil, shutdownContainer, fmt.Errorf("unsupported database type")
	}

	if err != nil {
		shutdownContainer()

		return nil, nil, fmt.Errorf("failed to open database: %w", err)
	}

	basic.TelemetryService.Slogger.Info("database connection established successfully")

	return db, shutdownContainer, nil
}

// openSQLite opens a SQLite database connection with GORM and configures WAL mode.
func openSQLite(ctx context.Context, databaseURL string, debugMode bool) (*gorm.DB, error) {
	// Open database connection using database/sql.
	sqlDB, err := sql.Open("sqlite", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// Configure SQLite for concurrent operations.
	// Note: Skip WAL mode for in-memory databases as it's not supported.
	isInMemory := databaseURL == ":memory:" || databaseURL == "file::memory:?cache=shared" ||
		(len(databaseURL) >= 7 && databaseURL[:7] == "file:/:" && (len(databaseURL) < 9 || databaseURL[7:9] == ":m"))

	if !isInMemory {
		if _, err := sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;"); err != nil {
			_ = sqlDB.Close()

			return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
		}
	}

	if _, err := sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;"); err != nil {
		_ = sqlDB.Close()

		return nil, fmt.Errorf("failed to set busy timeout: %w", err)
	}

	// Wrap with GORM.
	dialector := sqlite.Dialector{Conn: sqlDB}

	gormConfig := &gorm.Config{SkipDefaultTransaction: true}
	if debugMode {
		gormConfig.Logger = gormConfig.Logger.LogMode(cryptoutilMagic.GormLogModeInfo)
	}

	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		_ = sqlDB.Close()

		return nil, fmt.Errorf("failed to initialize GORM: %w", err)
	}

	// Configure connection pool for GORM transactions.
	sqlDB, err = db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(cryptoutilMagic.SQLiteMaxOpenConnections)
	sqlDB.SetMaxIdleConns(cryptoutilMagic.SQLiteMaxOpenConnections)
	sqlDB.SetConnMaxLifetime(0) // In-memory: never close connections.

	return db, nil
}

// openPostgreSQL opens a PostgreSQL database connection with GORM.
func openPostgreSQL(_ context.Context, databaseURL string, debugMode bool) (*gorm.DB, error) {
	gormConfig := &gorm.Config{SkipDefaultTransaction: true}
	if debugMode {
		gormConfig.Logger = gormConfig.Logger.LogMode(cryptoutilMagic.GormLogModeInfo)
	}

	db, err := gorm.Open(postgres.Open(databaseURL), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL database: %w", err)
	}

	// Configure connection pool.
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(cryptoutilMagic.PostgreSQLMaxOpenConns)
	sqlDB.SetMaxIdleConns(cryptoutilMagic.PostgreSQLMaxIdleConns)

	return db, nil
}

// maskPassword masks the password in a PostgreSQL connection string for logging.
func maskPassword(dsn string) string {
	// Simple masking: replace password with "***".
	// Format: postgres://user:password@host:port/db
	// This is a naive implementation; production code should use url.Parse.
	start := 0

	for i := 0; i < len(dsn); i++ {
		if dsn[i] == ':' && i > 0 && dsn[i-1] == '/' {
			start = i + 1

			break
		}
	}

	if start == 0 {
		return dsn
	}

	end := start

	for i := start; i < len(dsn); i++ {
		if dsn[i] == '@' {
			end = i

			break
		}
	}

	if end == start {
		return dsn
	}

	return dsn[:start] + "***" + dsn[end:]
}
