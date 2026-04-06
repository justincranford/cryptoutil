// Copyright (c) 2025 Justin Cranford

package compose_db_naming_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessComposeDBNaming "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/compose_db_naming"
	lintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func newTestLogger() *cryptoutilCmdCicdCommon.Logger {
	return cryptoutilCmdCicdCommon.NewLogger("test")
}

func findProjectRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	require.NoError(t, err, "failed to get working directory")

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Skip("skipping integration test: cannot find project root (no go.mod)")
		}

		dir = parent
	}
}

// writeComposeYML writes a compose.yml with the given content under deployments/{psID}/.
func writeComposeYML(t *testing.T, tmpDir, psID string, content string) {
	t.Helper()

	deployDir := filepath.Join(tmpDir, "deployments", psID)
	require.NoError(t, os.MkdirAll(deployDir, cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.WriteFile(filepath.Join(deployDir, "compose.yml"), []byte(content), cryptoutilSharedMagic.FilePermissions))
}

// correctDBCompose generates a minimal compose.yml without a per-PS-ID DB service.
// After Framework v8, per-PS-ID postgres is provided by the shared-postgres tier.
func correctDBCompose(psID string) string {
	return "services:\n  " + psID + "-app-sqlite-1: {}\n"
}

// setupAllComposeFiles creates correct compose files for all 10 PS.
func setupAllComposeFiles(t *testing.T, tmpDir string) {
	t.Helper()

	for _, ps := range lintFitnessRegistry.AllProductServices() {
		writeComposeYML(t, tmpDir, ps.PSID, correctDBCompose(ps.PSID))
	}
}

func TestCheck_DelegatesToCheckInDir(t *testing.T) {
	// Not parallel: changes process working directory.
	root := findProjectRoot(t)

	orig, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(root))

	defer func() { _ = os.Chdir(orig) }()

	err = lintFitnessComposeDBNaming.Check(newTestLogger())
	require.NoError(t, err, "Check() should pass on project root (delegates to CheckInDir)")
}

func TestCheck_RealWorkspace(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)

	err := lintFitnessComposeDBNaming.CheckInDir(newTestLogger(), root)
	require.NoError(t, err)
}

func TestCheckInDir_AllCorrect(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupAllComposeFiles(t, tmpDir)

	err := lintFitnessComposeDBNaming.CheckInDir(newTestLogger(), tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_MissingComposeFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupAllComposeFiles(t, tmpDir)

	require.NoError(t, os.Remove(filepath.Join(tmpDir, "deployments", cryptoutilSharedMagic.OTLPServiceSMIM, "compose.yml")))

	err := lintFitnessComposeDBNaming.CheckInDir(newTestLogger(), tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), cryptoutilSharedMagic.OTLPServiceSMIM)
}

func TestCheckInDir_DBServicePresent_Fails(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupAllComposeFiles(t, tmpDir)

	// Write sm-im compose WITH the legacy per-PS-ID DB service — must be rejected.
	writeComposeYML(t, tmpDir, cryptoutilSharedMagic.OTLPServiceSMIM, `services:
  sm-im-app-sqlite-1: {}
  sm-im-db-postgres-1:
    container_name: sm-im-postgres
    hostname: sm-im-postgres
`)

	err := lintFitnessComposeDBNaming.CheckInDir(newTestLogger(), tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), cryptoutilSharedMagic.OTLPServiceSMIM)
	assert.Contains(t, err.Error(), "sm-im-db-postgres-1")
	assert.Contains(t, err.Error(), "shared-postgres tier")
}

func TestCheckInDir_InvalidYAML(t *testing.T) {
	t.Parallel()

	psID := cryptoutilSharedMagic.OTLPServiceSMIM

	tmpDir := t.TempDir()
	setupAllComposeFiles(t, tmpDir)

	// Overwrite sm-im compose.yml with invalid YAML.
	deployDir := filepath.Join(tmpDir, "deployments", psID)
	require.NoError(t, os.WriteFile(filepath.Join(deployDir, "compose.yml"), []byte("services: [\ninvalid yaml"), cryptoutilSharedMagic.FilePermissions))

	err := lintFitnessComposeDBNaming.CheckInDir(newTestLogger(), tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot parse")
}
