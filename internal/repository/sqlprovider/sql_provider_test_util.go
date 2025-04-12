package sqlprovider

import (
	"context"
	"fmt"

	cryptoutilAppErr "cryptoutil/internal/apperr"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"
)

func RequireNewForTest(ctx context.Context, telemetryService *cryptoutilTelemetry.Service, dbType SupportedDBType) *SqlProvider {
	var sqlProvider *SqlProvider
	var err error
	switch dbType {
	case DBTypeSQLite:
		sqlProvider, err = NewSqlProvider(ctx, telemetryService, DBTypeSQLite, ":memory:", ContainerModeDisabled)
	case DBTypePostgres:
		sqlProvider, err = NewSqlProvider(ctx, telemetryService, DBTypePostgres, "", ContainerModeRequired)
	default:
		err = fmt.Errorf("unsupported dbType %s", dbType)
	}
	cryptoutilAppErr.RequireNoError(err, "failed to initialize SQL provider")
	return sqlProvider
}
