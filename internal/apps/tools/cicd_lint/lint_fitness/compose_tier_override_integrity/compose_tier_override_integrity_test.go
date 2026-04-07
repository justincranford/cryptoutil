// Copyright (c) 2025 Justin Cranford

package compose_tier_override_integrity_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessComposeTierOverrideIntegrity "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/compose_tier_override_integrity"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func testLogger() *cryptoutilCmdCicdCommon.Logger {
	return cryptoutilCmdCicdCommon.NewLogger("test")
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.WriteFile(path, []byte(content), cryptoutilSharedMagic.FilePermissions))
}

func setupTierFixture(t *testing.T, root string) {
	t.Helper()

	composeTemplate := `services:
  app: {}
secrets:
  postgres-url.secret: {file: ./secrets/postgres-url.secret}
  postgres-username.secret: {file: ./secrets/postgres-username.secret}
  postgres-password.secret: {file: ./secrets/postgres-password.secret}
  postgres-database.secret: {file: ./secrets/postgres-database.secret}
`

	for _, tier := range []string{cryptoutilSharedMagic.DefaultOTLPServiceDefault, "sm"} {
		writeFile(t, filepath.Join(root, "deployments", tier, "compose.yml"), composeTemplate)
		writeFile(t, filepath.Join(root, "deployments", tier, "secrets", "postgres-username.secret"), tier+"_user")
		writeFile(t, filepath.Join(root, "deployments", tier, "secrets", "postgres-password.secret"), tier+"_pass")
		writeFile(t, filepath.Join(root, "deployments", tier, "secrets", "postgres-database.secret"), tier+"_db")
		writeFile(t, filepath.Join(root, "deployments", tier, "secrets", "postgres-url.secret"), "postgres://"+tier+"_user:"+tier+"_pass@postgres-leader:5432/"+tier+"_db?sslmode=disable")
	}
}

func TestCheckInDir_Valid(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	setupTierFixture(t, root)

	err := lintFitnessComposeTierOverrideIntegrity.CheckInDir(testLogger(), root)
	require.NoError(t, err)
}

func TestCheckInDir_ForbiddenBuilderService(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	setupTierFixture(t, root)

	writeFile(t, filepath.Join(root, "deployments", cryptoutilSharedMagic.DefaultOTLPServiceDefault, "compose.yml"),
		fmt.Sprintf("services:\n  %s: {}\nsecrets:\n  postgres-url.secret: {file: ./secrets/postgres-url.secret}\n  postgres-username.secret: {file: ./secrets/postgres-username.secret}\n  postgres-password.secret: {file: ./secrets/postgres-password.secret}\n  postgres-database.secret: {file: ./secrets/postgres-database.secret}\n", cryptoutilSharedMagic.DockerJobBuilderCryptoutil))

	err := lintFitnessComposeTierOverrideIntegrity.CheckInDir(testLogger(), root)
	require.Error(t, err)
	require.Contains(t, err.Error(), cryptoutilSharedMagic.DockerJobBuilderCryptoutil)
}

func TestCheckInDir_MissingPostgresSecretOverride(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	setupTierFixture(t, root)

	writeFile(t, filepath.Join(root, "deployments", "sm", "compose.yml"), `services:
  app: {}
secrets:
  postgres-url.secret: {file: ./secrets/postgres-url.secret}
  postgres-username.secret: {file: ./secrets/postgres-username.secret}
`)

	err := lintFitnessComposeTierOverrideIntegrity.CheckInDir(testLogger(), root)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing required postgres secret override")
}

func TestCheckInDir_PostgresURLNotSynced(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	setupTierFixture(t, root)

	writeFile(t, filepath.Join(root, "deployments", "sm", "secrets", "postgres-url.secret"), "postgres://wrong:wrong@postgres-leader:5432/sm_db")

	err := lintFitnessComposeTierOverrideIntegrity.CheckInDir(testLogger(), root)
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not match")
}
