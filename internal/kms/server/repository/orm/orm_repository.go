// Copyright (c) 2025 Justin Cranford
//
//

package orm

import (
	"context"
	"fmt"

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSQLRepository "cryptoutil/internal/kms/server/repository/sqlrepository"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"

	"gorm.io/gorm"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver for database/sql
	_ "modernc.org/sqlite"             // SQLite driver for database/sql
)

// OrmRepository provides GORM-based database operations for the KMS server.
type OrmRepository struct {
	telemetryService *cryptoutilTelemetry.TelemetryService
	sqlRepository    *cryptoutilSQLRepository.SQLRepository
	jwkGenService    *cryptoutilJose.JWKGenService
	verboseMode      bool
	gormDB           *gorm.DB
}

// NewOrmRepository creates a new OrmRepository with the provided dependencies.
func NewOrmRepository(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, sqlRepository *cryptoutilSQLRepository.SQLRepository, jwkGenService *cryptoutilJose.JWKGenService, settings *cryptoutilConfig.ServiceTemplateServerSettings) (*OrmRepository, error) {
	if ctx == nil {
		return nil, fmt.Errorf("ctx must be non-nil")
	} else if telemetryService == nil {
		return nil, fmt.Errorf("telemetryService must be non-nil")
	} else if sqlRepository == nil {
		return nil, fmt.Errorf("sqlRepository must be non-nil")
	} else if jwkGenService == nil {
		return nil, fmt.Errorf("jwkGenService must be non-nil")
	}

	gormDB, err := cryptoutilSQLRepository.CreateGormDB(sqlRepository)
	if err != nil {
		return nil, fmt.Errorf("failed to connect with gormDB: %w", err)
	}

	return &OrmRepository{telemetryService: telemetryService, sqlRepository: sqlRepository, jwkGenService: jwkGenService, gormDB: gormDB, verboseMode: settings.VerboseMode}, nil
}

// Shutdown releases resources held by the OrmRepository.
func (r *OrmRepository) Shutdown() {
	// no-op
}
