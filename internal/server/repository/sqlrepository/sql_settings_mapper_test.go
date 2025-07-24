package sqlrepository

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_DevMode(t *testing.T) {
	dbType, url, err := mapDbTypeAndUrl(testTelemetryService, true, "ignored-url")
	require.NoError(t, err, "expected no error for dev mode")
	require.Equal(t, DBTypeSQLite, dbType, "expected SQLite in dev mode")
	require.Equal(t, ":memory:", url, "expected SQLite in-memory")
}

func Test_DatabaseUrl_PostgresSQL(t *testing.T) {
	dbType, url, err := mapDbTypeAndUrl(testTelemetryService, false, "postgres://user:pass@localhost/db")
	require.NoError(t, err, "expected no error for Postgres URL")
	require.Equal(t, DBTypePostgres, dbType, "expected Postgres dbType")
	require.Equal(t, "postgres://user:pass@localhost/db", url, "expected Postgres URL")
}

func Test_DatabaseUrl_Unsupported(t *testing.T) {
	dbType, url, err := mapDbTypeAndUrl(testTelemetryService, false, "mysql://user:pass@localhost/db")
	require.Error(t, err, "expected error for unsupported DB URL")
	require.Equal(t, "", url, "expected empty URL for unsupported DB type")
	require.Equal(t, SupportedDBType(""), dbType, "expected empty dbType for unsupported URL")
}

func Test_DatabaseUrl_Empty(t *testing.T) {
	dbType, url, err := mapDbTypeAndUrl(testTelemetryService, false, "")
	require.Error(t, err, "expected error for empty database URL")
	require.Equal(t, SupportedDBType(""), dbType, "expected empty dbType for empty URL")
	require.Equal(t, "", url, "expected empty URL for empty input")
}

func Test_DatabaseContainerMode_Disabled(t *testing.T) {
	mode, err := mapContainerMode(testTelemetryService, string(ContainerModeDisabled))
	require.NoError(t, err, "expected no error for disabled container mode")
	require.Equal(t, ContainerModeDisabled, mode, "expected mode to match disabled")
}

func Test_DatabaseContainerMode_Preferred(t *testing.T) {
	mode, err := mapContainerMode(testTelemetryService, string(ContainerModePreferred))
	require.NoError(t, err, "expected no error for preferred container mode")
	require.Equal(t, ContainerModePreferred, mode, "expected mode to match preferred")
}

func Test_DatabaseContainerMode_Required(t *testing.T) {
	mode, err := mapContainerMode(testTelemetryService, string(ContainerModeRequired))
	require.NoError(t, err, "expected no error for required container mode")
	require.Equal(t, ContainerModeRequired, mode, "expected mode to match required")
}

func Test_DatabaseContainerMode_Invalid(t *testing.T) {
	mode, err := mapContainerMode(testTelemetryService, "invalid-mode")
	require.Error(t, err, "expected error for unsupported container mode")
	require.Equal(t, ContainerMode(""), mode, "expected empty mode for invalid input")
}

func Test_DatabaseContainerMode_Empty(t *testing.T) {
	mode, err := mapContainerMode(testTelemetryService, "")
	require.Error(t, err, "expected error for empty container mode")
	require.Equal(t, ContainerMode(""), mode, "expected empty mode for empty input")
}
