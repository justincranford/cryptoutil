// Copyright (c) 2025-2026 Justin Cranford.

// Package database implements explicit, suite-routed lifecycle database operations.
package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"

	cryptoutilAppsFrameworkServiceServerRepository "cryptoutil/internal/apps-framework/service/server/repository"
	authz "cryptoutil/internal/apps/identity-authz/server/repository"
	idp "cryptoutil/internal/apps/identity-idp/server/repository"
	rp "cryptoutil/internal/apps/identity-rp/server/repository"
	rs "cryptoutil/internal/apps/identity-rs/server/repository"
	spa "cryptoutil/internal/apps/identity-spa/server/repository"
	skeleton "cryptoutil/internal/apps/skeleton-template/server/repository"
	cryptoutilAppsSmKmsRepository "cryptoutil/internal/apps/sm-kms/server/repository"
)

// Migrations returns the same merged embedded FS used by each service builder.
// It never accepts operator SQL or a filesystem override.
func Migrations(psid string) (fs.FS, error) {
	switch psid {
	case psidAuthz:
		return authz.GetMergedMigrationsFS(), nil
	case psidIDP:
		return idp.GetMergedMigrationsFS(), nil
	case psidRP:
		return rp.GetMergedMigrationsFS(), nil
	case psidRS:
		return rs.GetMergedMigrationsFS(), nil
	case psidSPA:
		return spa.GetMergedMigrationsFS(), nil
	case psidSkeleton:
		return skeleton.GetMergedMigrationsFS(), nil
	case psidKMS:
		return cryptoutilAppsFrameworkServiceServerRepository.NewMergedMigrationsFS(cryptoutilAppsSmKmsRepository.MigrationsFS), nil
	default:
		return nil, errors.New("unsupported database service")
	}
}

// Apply uses the existing production framework migration runner.
func Apply(ctx context.Context, db *sql.DB, psid, backend string) error {
	if ctx == nil || db == nil {
		return errors.New("database migration requires context and database")
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("database migration context: %w", err)
	}

	files, err := Migrations(psid)
	if err != nil {
		return err
	}

	if backend != SQLite && backend != Postgres {
		return errors.New("unsupported database backend")
	}

	dbType := cryptoutilAppsFrameworkServiceServerRepository.DatabaseTypeSQLite
	if backend == Postgres {
		dbType = cryptoutilAppsFrameworkServiceServerRepository.DatabaseTypePostgreSQL
	}

	if err := cryptoutilAppsFrameworkServiceServerRepository.NewMigrationRunner(files, "migrations").ApplyContext(ctx, db, dbType); err != nil {
		if ctx.Err() != nil {
			return contextFailure(ctx)
		}

		return errors.New("embedded database migration rejected")
	}

	return contextFailure(ctx)
}
