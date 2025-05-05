package sqlrepository

import (
	"context"
	"fmt"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
)

func RequireNewForTest(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, dbType SupportedDBType) *SqlRepository {
	var sqlRepository *SqlRepository
	var err error
	switch dbType {
	case DBTypeSQLite:
		sqlRepository, err = NewSqlRepository(ctx, telemetryService, DBTypeSQLite, ":memory:", ContainerModeDisabled)
	case DBTypePostgres:
		sqlRepository, err = NewSqlRepository(ctx, telemetryService, DBTypePostgres, "", ContainerModeRequired)
	default:
		err = fmt.Errorf("unsupported dbType %s", dbType)
	}
	cryptoutilAppErr.RequireNoError(err, "failed to initialize SQL provider")
	return sqlRepository
}
