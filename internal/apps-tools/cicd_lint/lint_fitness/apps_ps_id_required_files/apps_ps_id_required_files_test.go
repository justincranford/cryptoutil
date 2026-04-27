// Copyright (c) 2025 Justin Cranford

package apps_ps_id_required_files

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilFitnessRegistry "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

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

// createAllPSIDs creates all required files for every PS-ID in the synthetic root.
func createAllPSIDs(t *testing.T, tmpDir string) {
	t.Helper()

	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		serviceDir := filepath.Join(tmpDir, "internal", "apps", ps.PSID)
		require.NoError(t, os.MkdirAll(serviceDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.WriteFile(filepath.Join(serviceDir, ps.Service+".go"), []byte("package x\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(serviceDir, ps.Service+"_usage.go"), []byte("package x\n"), cryptoutilSharedMagic.CacheFilePermissions))
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

func TestCheck_FromProjectRoot(t *testing.T) {
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

func TestCheckInDir_AllValid(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createAllPSIDs(t, tmpDir)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_NoAppsDir(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, t.TempDir())
	require.Error(t, err)
	require.Contains(t, err.Error(), "internal/apps directory not found")
}

func TestCheckInDir_MissingPSIDDir(t *testing.T) {
	t.Parallel()

	services := cryptoutilFitnessRegistry.AllProductServices()
	if len(services) == 0 {
		t.Skip("no product services in registry")
	}

	target := services[0]

	tmpDir := t.TempDir()

	// Create all PS-IDs except the first.
	for _, ps := range services[1:] {
		serviceDir := filepath.Join(tmpDir, "internal", "apps", ps.PSID)
		require.NoError(t, os.MkdirAll(serviceDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.WriteFile(filepath.Join(serviceDir, ps.Service+".go"), []byte("package x\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(serviceDir, ps.Service+"_usage.go"), []byte("package x\n"), cryptoutilSharedMagic.CacheFilePermissions))
	}

	// Create only the apps dir (not the target PS-ID dir).
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "internal", "apps"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "PS-ID directory missing")
	require.Contains(t, err.Error(), target.PSID)
}

func TestCheckInDir_MissingEntryFile(t *testing.T) {
	t.Parallel()

	services := cryptoutilFitnessRegistry.AllProductServices()
	if len(services) == 0 {
		t.Skip("no product services in registry")
	}

	target := services[0]

	tmpDir := t.TempDir()
	createAllPSIDs(t, tmpDir)

	// Remove entry file for the first PS-ID.
	entryFile := filepath.Join(tmpDir, "internal", "apps", target.PSID, target.Service+".go")
	require.NoError(t, os.Remove(entryFile))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing required file")
	require.Contains(t, err.Error(), target.Service+".go")
}

func TestCheckInDir_MissingUsageFile(t *testing.T) {
	t.Parallel()

	services := cryptoutilFitnessRegistry.AllProductServices()
	if len(services) == 0 {
		t.Skip("no product services in registry")
	}

	target := services[0]

	tmpDir := t.TempDir()
	createAllPSIDs(t, tmpDir)

	// Remove usage file for the first PS-ID.
	usageFile := filepath.Join(tmpDir, "internal", "apps", target.PSID, target.Service+"_usage.go")
	require.NoError(t, os.Remove(usageFile))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing required file")
	require.Contains(t, err.Error(), target.Service+"_usage.go")
}

func TestCheckInDir_MultipleViolations(t *testing.T) {
	t.Parallel()

	services := cryptoutilFitnessRegistry.AllProductServices()
	if len(services) < 2 {
		t.Skip("need at least 2 product services for multi-violation test")
	}

	tmpDir := t.TempDir()
	createAllPSIDs(t, tmpDir)

	// Remove entry file from two PS-IDs.
	for _, ps := range services[:2] {
		entryFile := filepath.Join(tmpDir, "internal", "apps", ps.PSID, ps.Service+".go")
		require.NoError(t, os.Remove(entryFile))
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)

	// Both violations should appear in the error.
	require.Contains(t, err.Error(), services[0].Service+".go")
	require.Contains(t, err.Error(), services[1].Service+".go")
}
