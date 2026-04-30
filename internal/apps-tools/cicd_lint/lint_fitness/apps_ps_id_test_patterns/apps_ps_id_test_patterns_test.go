// Copyright (c) 2025-2026 Justin Cranford.
package apps_ps_id_test_patterns

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilFitnessRegistry "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// emptyExclusions is a convenience for tests that want no exclusions.
var emptyExclusions = map[string]bool{}

// findProjectRoot finds the project root by looking for go.mod.
func findProjectRoot() (string, error) {
	dir, _ := os.Getwd()

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}

		dir = parent
	}
}

// createAllPSIDsWithTestFiles creates all three test pattern files for every PS-ID.
func createAllPSIDsWithTestFiles(t *testing.T, tmpDir string) {
	t.Helper()

	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		serverDir := filepath.Join(tmpDir, "internal", "apps", ps.PSID, "server")
		require.NoError(t, os.MkdirAll(serverDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, "testmain_test.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, ps.Service+"_lifecycle_test.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, ps.Service+"_port_conflict_test.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
	}
}

func TestCheck_RealWorkspace(t *testing.T) {
	t.Parallel()

	root, err := findProjectRoot()
	if err != nil {
		t.Skip("Skipping integration test - cannot find project root (no go.mod)")
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = CheckInDir(logger, root)
	require.NoError(t, err)
}

// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheck_Integration(t *testing.T) {
	root, err := findProjectRoot()
	if err != nil {
		t.Skip("Skipping integration test - cannot find project root (no go.mod)")
	}

	origDir, getErr := os.Getwd()
	require.NoError(t, getErr)

	require.NoError(t, os.Chdir(root))

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger)
	require.NoError(t, err)
}

func TestCheckInDir_NoAppsDir(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, t.TempDir())
	require.Error(t, err)
	require.Contains(t, err.Error(), "internal/apps directory not found")
}

// TestCheckInDir_NoExclusions_AllValid verifies all PS-IDs pass when all three test files present.
func TestCheckInDir_NoExclusions_AllValid(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createAllPSIDsWithTestFiles(t, tmpDir)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := ExportedCheckInDirWithExclusions(logger, tmpDir, emptyExclusions, emptyExclusions, emptyExclusions)
	require.NoError(t, err)
}

// TestCheckInDir_NoExclusions_MissingTestMain verifies error on missing testmain_test.go.
func TestCheckInDir_NoExclusions_MissingTestMain(t *testing.T) {
	t.Parallel()

	services := cryptoutilFitnessRegistry.AllProductServices()
	if len(services) == 0 {
		t.Skip("no product services in registry")
	}

	target := services[0]

	tmpDir := t.TempDir()
	createAllPSIDsWithTestFiles(t, tmpDir)

	require.NoError(t, os.Remove(filepath.Join(tmpDir, "internal", "apps", target.PSID, "server", "testmain_test.go")))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := ExportedCheckInDirWithExclusions(logger, tmpDir, emptyExclusions, emptyExclusions, emptyExclusions)
	require.Error(t, err)
	require.Contains(t, err.Error(), "testmain_test.go")
}

// TestCheckInDir_NoExclusions_MissingLifecycle verifies error on missing *_lifecycle_test.go.
func TestCheckInDir_NoExclusions_MissingLifecycle(t *testing.T) {
	t.Parallel()

	services := cryptoutilFitnessRegistry.AllProductServices()
	if len(services) == 0 {
		t.Skip("no product services in registry")
	}

	target := services[0]

	tmpDir := t.TempDir()
	createAllPSIDsWithTestFiles(t, tmpDir)

	require.NoError(t, os.Remove(filepath.Join(tmpDir, "internal", "apps", target.PSID, "server", target.Service+"_lifecycle_test.go")))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := ExportedCheckInDirWithExclusions(logger, tmpDir, emptyExclusions, emptyExclusions, emptyExclusions)
	require.Error(t, err)
	require.Contains(t, err.Error(), "_lifecycle_test.go")
}

// TestCheckInDir_NoExclusions_MissingPortConflict verifies error on missing *_port_conflict_test.go.
func TestCheckInDir_NoExclusions_MissingPortConflict(t *testing.T) {
	t.Parallel()

	services := cryptoutilFitnessRegistry.AllProductServices()
	if len(services) == 0 {
		t.Skip("no product services in registry")
	}

	target := services[0]

	tmpDir := t.TempDir()
	createAllPSIDsWithTestFiles(t, tmpDir)

	require.NoError(t, os.Remove(filepath.Join(tmpDir, "internal", "apps", target.PSID, "server", target.Service+"_port_conflict_test.go")))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := ExportedCheckInDirWithExclusions(logger, tmpDir, emptyExclusions, emptyExclusions, emptyExclusions)
	require.Error(t, err)
	require.Contains(t, err.Error(), "_port_conflict_test.go")
}

// TestCheckInDir_NoExclusions_MissingPSIDDir verifies error on missing PS-ID directory.
func TestCheckInDir_NoExclusions_MissingPSIDDir(t *testing.T) {
	t.Parallel()

	services := cryptoutilFitnessRegistry.AllProductServices()
	if len(services) == 0 {
		t.Skip("no product services in registry")
	}

	target := services[0]

	tmpDir := t.TempDir()
	createAllPSIDsWithTestFiles(t, tmpDir)

	require.NoError(t, os.RemoveAll(filepath.Join(tmpDir, "internal", "apps", target.PSID)))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := ExportedCheckInDirWithExclusions(logger, tmpDir, emptyExclusions, emptyExclusions, emptyExclusions)
	require.Error(t, err)
	require.Contains(t, err.Error(), "PS-ID directory missing")
}

// TestFileExists verifies fileExists handles both existing and absent paths.
func TestFileExists(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	existing := filepath.Join(tmpDir, "exists.go")
	require.NoError(t, os.WriteFile(existing, []byte("package x\n"), cryptoutilSharedMagic.CacheFilePermissions))

	require.True(t, ExportedFileExists(existing))
	require.False(t, ExportedFileExists(filepath.Join(tmpDir, "nope.go")))
	require.False(t, ExportedFileExists(tmpDir)) // directory — not a file
}

// TestGlobExists verifies globExists finds matching files.
func TestGlobExists(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "foo_lifecycle_test.go"), []byte("package x\n"), cryptoutilSharedMagic.CacheFilePermissions))

	require.True(t, ExportedGlobExists(tmpDir, "*_lifecycle_test.go"))
	require.False(t, ExportedGlobExists(tmpDir, "*_port_conflict_test.go"))
	require.False(t, ExportedGlobExists(filepath.Join(tmpDir, "nope"), "*_lifecycle_test.go")) // nonexistent dir
}
