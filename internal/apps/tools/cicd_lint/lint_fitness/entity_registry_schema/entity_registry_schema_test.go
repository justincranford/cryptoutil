// Copyright (c) 2025 Justin Cranford

package entity_registry_schema_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessEntityRegistrySchema "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/entity_registry_schema"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

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

func TestCheckInDir_RealProject(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)
	logger := cryptoutilCmdCicdCommon.NewLogger("entity-registry-schema-test")

	err := lintFitnessEntityRegistrySchema.CheckInDir(logger, root)

	require.NoError(t, err, "real registry.yaml must be valid")
}

func TestCheckInDir_MissingRegistry(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmdCicdCommon.NewLogger("entity-registry-schema-test")

	err := lintFitnessEntityRegistrySchema.CheckInDir(logger, tmpDir)

	require.Error(t, err, "missing registry.yaml must produce an error")
	require.Contains(t, err.Error(), "entity registry schema violations")
}

func TestCheckInDir_InvalidYAML(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	registryDir := filepath.Join(tmpDir, "api", "cryptosuite-registry")
	require.NoError(t, os.MkdirAll(registryDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	badYAML := "suites: [\n  invalid yaml {{\n"
	require.NoError(t, os.WriteFile(filepath.Join(registryDir, "registry.yaml"), []byte(badYAML), cryptoutilSharedMagic.FilePermissionsDefault))

	logger := cryptoutilCmdCicdCommon.NewLogger("entity-registry-schema-test")

	err := lintFitnessEntityRegistrySchema.CheckInDir(logger, tmpDir)

	require.Error(t, err, "invalid YAML must produce an error")
	require.Contains(t, err.Error(), "entity registry schema violations")
}

func TestCheckInDir_EmptySuites(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	registryDir := filepath.Join(tmpDir, "api", "cryptosuite-registry")
	require.NoError(t, os.MkdirAll(registryDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	yaml := "suites: []\nproducts: []\nproduct_services: []\n"
	require.NoError(t, os.WriteFile(filepath.Join(registryDir, "registry.yaml"), []byte(yaml), cryptoutilSharedMagic.FilePermissionsDefault))

	logger := cryptoutilCmdCicdCommon.NewLogger("entity-registry-schema-test")

	err := lintFitnessEntityRegistrySchema.CheckInDir(logger, tmpDir)

	require.Error(t, err, "empty suites must produce an error")
	require.Contains(t, err.Error(), "entity registry schema violations")
}

// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheck_RealProject(t *testing.T) {
	// Check() calls CheckInDir(logger, ".") — only valid when CWD is project root.
	root := findProjectRoot(t)

	origDir, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(root))

	t.Cleanup(func() { _ = os.Chdir(origDir) })

	logger := cryptoutilCmdCicdCommon.NewLogger("entity-registry-schema-test")

	err = lintFitnessEntityRegistrySchema.Check(logger)

	require.NoError(t, err, "Check() from project root must pass")
}
