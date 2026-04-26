// Copyright (c) 2025 Justin Cranford
//
//

package orm

import (
	"context"
	"fmt"

	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	cryptoutilFrameworkOrm "cryptoutil/internal/apps-framework/service/server/repository/orm"

	"gorm.io/gorm"
)

// Re-export framework ORM types as aliases so existing sm-kms code requires no import changes.
type (
	OrmRepository   = cryptoutilFrameworkOrm.OrmRepository
	OrmTransaction  = cryptoutilFrameworkOrm.OrmTransaction
	TransactionMode = cryptoutilFrameworkOrm.TransactionMode
)

// Re-export framework transaction mode constants.
var (
	AutoCommit = cryptoutilFrameworkOrm.AutoCommit
	ReadWrite  = cryptoutilFrameworkOrm.ReadWrite
	ReadOnly   = cryptoutilFrameworkOrm.ReadOnly
)

// NewOrmRepository creates a new OrmRepository. Delegates to the framework implementation.
func NewOrmRepository(ctx context.Context, telemetryService *cryptoutilSharedTelemetry.TelemetryService, gormDB *gorm.DB, jwkGenService *cryptoutilSharedCryptoJose.JWKGenService, verboseMode bool) (*OrmRepository, error) {
	repo, err := cryptoutilFrameworkOrm.NewOrmRepository(ctx, telemetryService, gormDB, jwkGenService, verboseMode)
	if err != nil {
		return nil, fmt.Errorf("failed to create ORM repository: %w", err)
	}

	return repo, nil
}
