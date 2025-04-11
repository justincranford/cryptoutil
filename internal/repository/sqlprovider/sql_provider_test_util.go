package sqlprovider

import (
	"context"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"
	"fmt"
)

func NewSqlProviderForTest(ctx context.Context, telemetryService *cryptoutilTelemetry.Service, dbType SupportedDBType) (*SqlProvider, error) {
	switch dbType {
	case DBTypeSQLite:
		return NewSqlProvider(ctx, telemetryService, DBTypeSQLite, ":memory:", ContainerModeDisabled)
	case DBTypePostgres:
		return NewSqlProvider(ctx, telemetryService, DBTypePostgres, "", ContainerModeRequired)
	default:
		return nil, fmt.Errorf("unsupported dbType %s", dbType)
	}
}
