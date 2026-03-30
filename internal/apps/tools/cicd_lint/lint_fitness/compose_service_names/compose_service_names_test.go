// Copyright (c) 2025 Justin Cranford

package compose_service_names_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessComposeServiceNames "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/compose_service_names"
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

// writeComposeYML writes a compose.yml with the given services block under deployments/{psID}/.
func writeComposeYML(t *testing.T, tmpDir, psID, servicesBlock string) {
	t.Helper()

	deployDir := filepath.Join(tmpDir, "deployments", psID)
	require.NoError(t, os.MkdirAll(deployDir, cryptoutilSharedMagic.CICDTempDirPermissions))

	content := "services:\n" + servicesBlock
	require.NoError(t, os.WriteFile(filepath.Join(deployDir, "compose.yml"), []byte(content), cryptoutilSharedMagic.FilePermissions))
}

// correctServicesBlock generates the 4 required service entries for a PS-ID.
func correctServicesBlock(psID string) string {
	return fmt.Sprintf("  %s-app-sqlite-1: {}\n  %s-app-postgres-1: {}\n  %s-app-postgres-2: {}\n  %s-db-postgres-1: {}\n",
		psID, psID, psID, psID)
}

// setupAllComposeFiles creates correct compose files for all 10 PS.
func setupAllComposeFiles(t *testing.T, tmpDir string) {
	t.Helper()

	for _, ps := range lintFitnessRegistry.AllProductServices() {
		writeComposeYML(t, tmpDir, ps.PSID, correctServicesBlock(ps.PSID))
	}
}

func TestCheck_DelegatesToCheckInDir(t *testing.T) {
	// Not parallel: changes process working directory.
	root := findProjectRoot(t)

	orig, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(root))

	defer func() { _ = os.Chdir(orig) }()

	err = lintFitnessComposeServiceNames.Check(newTestLogger())
	require.NoError(t, err, "Check() should pass on project root (delegates to CheckInDir)")
}

func TestCheck_RealWorkspace(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)

	err := lintFitnessComposeServiceNames.CheckInDir(newTestLogger(), root)
	require.NoError(t, err)
}

func TestCheckInDir_AllCorrect(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupAllComposeFiles(t, tmpDir)

	err := lintFitnessComposeServiceNames.CheckInDir(newTestLogger(), tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_MissingComposeFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupAllComposeFiles(t, tmpDir)

	require.NoError(t, os.Remove(filepath.Join(tmpDir, "deployments", cryptoutilSharedMagic.OTLPServiceSMIM, "compose.yml")))

	err := lintFitnessComposeServiceNames.CheckInDir(newTestLogger(), tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), cryptoutilSharedMagic.OTLPServiceSMIM)
}

func TestCheckInDir_MissingRequiredService(t *testing.T) {
	t.Parallel()

	psID := cryptoutilSharedMagic.OTLPServiceSMIM

	missingServiceTests := []struct {
		name           string
		missingService string
	}{
		{"missing sqlite service", cryptoutilSharedMagic.IME2ESQLiteContainer},
		{"missing postgres-1 service", cryptoutilSharedMagic.IME2EPostgreSQL1Container},
		{"missing postgres-2 service", cryptoutilSharedMagic.IME2EPostgreSQL2Container},
		{"missing db service", psID + "-db-postgres-1"},
	}

	for _, tc := range missingServiceTests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			setupAllComposeFiles(t, tmpDir)

			// Overwrite sm-im compose without the missing service.
			allServices := []string{
				psID + "-app-sqlite-1",
				psID + "-app-postgres-1",
				psID + "-app-postgres-2",
				psID + "-db-postgres-1",
			}
			services := ""

			for _, svc := range allServices {
				if svc != tc.missingService {
					services += "  " + svc + ": {}\n"
				}
			}

			writeComposeYML(t, tmpDir, psID, services)

			err := lintFitnessComposeServiceNames.CheckInDir(newTestLogger(), tmpDir)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.missingService)
		})
	}
}
