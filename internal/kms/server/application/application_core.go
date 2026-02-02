// Copyright (c) 2025 Justin Cranford
//
//

package application

import (
	"context"
	"fmt"
	"io/fs"
	"strings"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServerApplication "cryptoutil/internal/apps/template/service/server/application"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilKmsServerBusinesslogic "cryptoutil/internal/kms/server/businesslogic"
	cryptoutilKmsServerDemo "cryptoutil/internal/kms/server/demo"
	cryptoutilKmsServerRepository "cryptoutil/internal/kms/server/repository"
	cryptoutilOrmRepository "cryptoutil/internal/kms/server/repository/orm"
	cryptoutilBarrierService "cryptoutil/internal/shared/barrier"

	"gorm.io/gorm"
)

// ServerApplicationCore provides core server application components including database, ORM, barrier, and business logic services.
type ServerApplicationCore struct {
	ServerApplicationBasic *ServerApplicationBasic
	TemplateCore           *cryptoutilAppsTemplateServerApplication.Core
	DB                     *gorm.DB
	OrmRepository          *cryptoutilOrmRepository.OrmRepository
	BarrierService         *cryptoutilBarrierService.BarrierService
	BusinessLogicService   *cryptoutilKmsServerBusinesslogic.BusinessLogicService
	Settings               *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings
}

// StartServerApplicationCore initializes and starts a core server application with all essential services.
func StartServerApplicationCore(ctx context.Context, settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) (*ServerApplicationCore, error) {
	serverApplicationBasic, err := StartServerApplicationBasic(ctx, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to start basic server application: %w", err)
	}

	jwkGenService := serverApplicationBasic.JWKGenService

	serverApplicationCore := &ServerApplicationCore{}
	serverApplicationCore.ServerApplicationBasic = serverApplicationBasic
	serverApplicationCore.Settings = settings

	// Use template's StartCore to provision database (GORM directly, no SQLRepository).
	templateCore, err := cryptoutilAppsTemplateServerApplication.StartCore(ctx, settings)
	if err != nil {
		serverApplicationBasic.TelemetryService.Slogger.Error("failed to start template core (database)", "error", err)
		serverApplicationCore.Shutdown()

		return nil, fmt.Errorf("failed to start template core: %w", err)
	}

	serverApplicationCore.TemplateCore = templateCore
	serverApplicationCore.DB = templateCore.DB

	// CRITICAL: Apply migrations BEFORE initializing services (BarrierService needs barrier_root_keys table).
	// Get underlying sql.DB for migrations.
	sqlDB, err := templateCore.DB.DB()
	if err != nil {
		serverApplicationBasic.TelemetryService.Slogger.Error("failed to get sql.DB from GORM", "error", err)
		serverApplicationCore.Shutdown()

		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Determine database type from URL.
	databaseType := determineDatabaseType(settings.DatabaseURL)
	serverApplicationBasic.TelemetryService.Slogger.Debug("applying migrations", "databaseType", databaseType, "databaseURL", settings.DatabaseURL)

	// Create merged migrations (template 1001-1004 + KMS domain 2001+).
	mergedFS := &mergedMigrations{
		templateFS:   cryptoutilAppsTemplateServiceServerRepository.MigrationsFS,
		templatePath: "migrations",
		domainFS:     cryptoutilKmsServerRepository.MigrationsFS,
		domainPath:   "migrations",
	}

	// Apply merged migrations.
	if err := cryptoutilAppsTemplateServiceServerRepository.ApplyMigrationsFromFS(sqlDB, mergedFS, "", databaseType); err != nil {
		serverApplicationBasic.TelemetryService.Slogger.Error("failed to apply migrations", "error", err)
		serverApplicationCore.Shutdown()

		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	serverApplicationBasic.TelemetryService.Slogger.Debug("migrations applied successfully")

	// Use NewOrmRepositoryFromGORM (GORM directly, no SQLRepository wrapper).
	ormRepository, err := cryptoutilOrmRepository.NewOrmRepositoryFromGORM(ctx, serverApplicationBasic.TelemetryService, templateCore.DB, jwkGenService, settings.VerboseMode)
	if err != nil {
		serverApplicationBasic.TelemetryService.Slogger.Error("failed to create ORM repository", "error", err)
		serverApplicationCore.Shutdown()

		return nil, fmt.Errorf("failed to create ORM repository: %w", err)
	}

	serverApplicationCore.OrmRepository = ormRepository

	barrierService, err := cryptoutilBarrierService.NewService(ctx, serverApplicationBasic.TelemetryService, jwkGenService, ormRepository, serverApplicationBasic.UnsealKeysService)
	if err != nil {
		serverApplicationBasic.TelemetryService.Slogger.Error("failed to initialize barrier service", "error", err)
		serverApplicationCore.Shutdown()

		return nil, fmt.Errorf("failed to create barrier service: %w", err)
	}

	serverApplicationCore.BarrierService = barrierService

	businessLogicService, err := cryptoutilKmsServerBusinesslogic.NewBusinessLogicService(ctx, serverApplicationBasic.TelemetryService, jwkGenService, ormRepository, barrierService)
	if err != nil {
		serverApplicationBasic.TelemetryService.Slogger.Error("failed to initialize business logic service", "error", err)
		serverApplicationCore.Shutdown()

		return nil, fmt.Errorf("failed to initialize business logic service: %w", err)
	}

	serverApplicationCore.BusinessLogicService = businessLogicService

	// Seed or reset demo data if demo mode is enabled.
	if settings.DemoMode {
		serverApplicationBasic.TelemetryService.Slogger.Info("Demo mode enabled, seeding demo data")

		err = cryptoutilKmsServerDemo.SeedDemoData(ctx, serverApplicationBasic.TelemetryService, businessLogicService)
		if err != nil {
			serverApplicationBasic.TelemetryService.Slogger.Error("failed to seed demo data", "error", err)
			serverApplicationCore.Shutdown()

			return nil, fmt.Errorf("failed to seed demo data: %w", err)
		}
	} else if settings.ResetDemoMode {
		serverApplicationBasic.TelemetryService.Slogger.Info("Reset demo mode enabled, resetting demo data")

		err = cryptoutilKmsServerDemo.ResetDemoData(ctx, serverApplicationBasic.TelemetryService, businessLogicService)
		if err != nil {
			serverApplicationBasic.TelemetryService.Slogger.Error("failed to reset demo data", "error", err)
			serverApplicationCore.Shutdown()

			return nil, fmt.Errorf("failed to reset demo data: %w", err)
		}
	}

	return serverApplicationCore, nil
}

// Shutdown returns a shutdown function that gracefully stops all core application services.
func (c *ServerApplicationCore) Shutdown() func() {
	return func() {
		if c.ServerApplicationBasic != nil && c.ServerApplicationBasic.TelemetryService != nil {
			c.ServerApplicationBasic.TelemetryService.Slogger.Debug("stopping server core")
		}

		if c.BarrierService != nil {
			c.BarrierService.Shutdown()
		}

		if c.OrmRepository != nil {
			c.OrmRepository.Shutdown()
		}

		// Shutdown template core (database container cleanup).
		if c.TemplateCore != nil {
			c.TemplateCore.Shutdown()
		}

		if c.ServerApplicationBasic != nil {
			c.ServerApplicationBasic.Shutdown()
		}
	}
}

// determineDatabaseType returns "sqlite" or "postgres" based on the database URL.
func determineDatabaseType(databaseURL string) string {
	if databaseURL == "" ||
		databaseURL == "file::memory:?cache=shared" ||
		databaseURL == ":memory:" ||
		strings.HasPrefix(databaseURL, "file:") ||
		strings.HasPrefix(databaseURL, "sqlite://") {
		return "sqlite"
	}

	return "postgres"
}

// mergedMigrations combines template and domain migrations into a single filesystem view.
type mergedMigrations struct {
	templateFS   fs.FS
	templatePath string
	domainFS     fs.FS
	domainPath   string
}

const (
	currentDir = "."
	pathSep    = "/"
)

// Open implements fs.FS interface.
func (m *mergedMigrations) Open(name string) (fs.File, error) {
	// Try domain migrations first (they have higher version numbers).
	if m.domainFS != nil {
		fullPath := m.domainPath
		if name != currentDir && name != "" {
			fullPath = m.domainPath + pathSep + name
		}

		if f, err := m.domainFS.Open(fullPath); err == nil {
			return f, nil
		}
	}

	// Fall back to template migrations.
	fullPath := m.templatePath
	if name != currentDir && name != "" {
		fullPath = m.templatePath + pathSep + name
	}

	f, err := m.templateFS.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open migration file %s: %w", name, err)
	}

	return f, nil
}

// ReadDir implements fs.ReadDirFS interface.
func (m *mergedMigrations) ReadDir(name string) ([]fs.DirEntry, error) {
	var entries []fs.DirEntry

	// Read template migrations.
	templatePath := m.templatePath
	if name != currentDir && name != "" {
		templatePath = m.templatePath + pathSep + name
	}

	if templateEntries, err := fs.ReadDir(m.templateFS, templatePath); err == nil {
		entries = append(entries, templateEntries...)
	}

	// Read domain migrations.
	if m.domainFS != nil {
		domainPath := m.domainPath
		if name != currentDir && name != "" {
			domainPath = m.domainPath + pathSep + name
		}

		if domainEntries, err := fs.ReadDir(m.domainFS, domainPath); err == nil {
			entries = append(entries, domainEntries...)
		}
	}

	return entries, nil
}

// ReadFile implements fs.ReadFileFS interface.
func (m *mergedMigrations) ReadFile(name string) ([]byte, error) {
	// Try domain migrations first.
	if m.domainFS != nil {
		fullPath := m.domainPath
		if name != currentDir && name != "" {
			fullPath = m.domainPath + pathSep + name
		}

		if data, err := fs.ReadFile(m.domainFS, fullPath); err == nil {
			return data, nil
		}
	}

	// Fall back to template migrations.
	fullPath := m.templatePath
	if name != currentDir && name != "" {
		fullPath = m.templatePath + pathSep + name
	}

	return fs.ReadFile(m.templateFS, fullPath)
}

// Stat implements fs.StatFS interface.
func (m *mergedMigrations) Stat(name string) (fs.FileInfo, error) {
	// Try domain migrations first.
	if m.domainFS != nil {
		fullPath := m.domainPath
		if name != currentDir && name != "" {
			fullPath = m.domainPath + pathSep + name
		}

		if info, err := fs.Stat(m.domainFS, fullPath); err == nil {
			return info, nil
		}
	}

	// Fall back to template migrations.
	fullPath := m.templatePath
	if name != currentDir && name != "" {
		fullPath = m.templatePath + pathSep + name
	}

	return fs.Stat(m.templateFS, fullPath)
}
