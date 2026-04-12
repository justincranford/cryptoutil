// Copyright (c) 2025 Justin Cranford

package template_drift

import (
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
)

func TestCheckDockerfile(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-template-dockerfile")
	err := checkDockerfileInDir(logger, "../../../../../..", instantiate)
	require.NoError(t, err)
}

func TestCheckCompose(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-template-compose")
	err := checkComposeInDir(logger, "../../../../../..", instantiate)
	require.NoError(t, err)
}

func TestCheckConfigCommon(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-template-config-common")
	err := checkConfigCommonInDir(logger, "../../../../../..", instantiate)
	require.NoError(t, err)
}

func TestCheckConfigSQLite(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-template-config-sqlite")
	err := checkConfigSQLiteInDir(logger, "../../../../../..", instantiate)
	require.NoError(t, err)
}

func TestCheckConfigPostgreSQL(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-template-config-postgresql")
	err := checkConfigPostgreSQLInDir(logger, "../../../../../..", instantiate)
	require.NoError(t, err)
}

func TestCheckStandaloneConfig(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-template-standalone-config")
	err := checkStandaloneConfigInDir(logger, "../../../../../..", instantiate)
	require.NoError(t, err)
}
