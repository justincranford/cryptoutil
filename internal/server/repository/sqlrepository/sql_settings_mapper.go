package sqlrepository

import (
	"fmt"
	"strings"

	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
)

func mapDBTypeAndURL(telemetryService *cryptoutilTelemetry.TelemetryService, devMode bool, databaseURL string) (SupportedDBType, string, error) {
	if devMode {
		telemetryService.Slogger.Debug("running in dev mode, using in-memory SQLite database")
		return DBTypeSQLite, ":memory:", nil
	} else if strings.HasPrefix(databaseURL, "postgres://") {
		telemetryService.Slogger.Debug("running in production mode, using PostgreSQL database")
		return DBTypePostgres, databaseURL, nil
	}

	return "", "", fmt.Errorf("unsupported database URL format: %s", databaseURL)
}

func mapContainerMode(telemetryService *cryptoutilTelemetry.TelemetryService, containerMode string) (ContainerMode, error) {
	switch containerMode {
	case string(ContainerModeDisabled):
		telemetryService.Slogger.Debug("container mode is disabled, using provided database URL")
		return ContainerModeDisabled, nil
	case string(ContainerModePreferred):
		telemetryService.Slogger.Debug("container mode is preferred, trying to start a container")
		return ContainerModePreferred, nil
	case string(ContainerModeRequired):
		telemetryService.Slogger.Debug("container mode is required, trying to start a container")
		return ContainerModeRequired, nil
	default:
		return "", fmt.Errorf("unsupported container mode: %s", containerMode)
	}
}
