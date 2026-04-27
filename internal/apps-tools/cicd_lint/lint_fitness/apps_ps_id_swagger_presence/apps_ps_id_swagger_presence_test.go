// Copyright (c) 2025 Justin Cranford

package apps_ps_id_swagger_presence

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

// createExcludedAndNonExcludedPSIDs populates the synthetic root for all PS-IDs.
// Excluded PS-IDs get the PS-ID dir only. Non-excluded PS-IDs get full swagger files.
func createAllPSIDsWithSwaggerFiles(t *testing.T, tmpDir string) {
	t.Helper()

	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		serverDir := filepath.Join(tmpDir, "internal", "apps", ps.PSID, "server")
		require.NoError(t, os.MkdirAll(serverDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

		if !knownExclusions[ps.PSID] {
			require.NoError(t, os.WriteFile(filepath.Join(serverDir, "swagger.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
			require.NoError(t, os.WriteFile(filepath.Join(serverDir, "swagger_test.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
		}
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

func TestCheckInDir_AllExcluded_PassesWithoutSwagger(t *testing.T) {
	t.Parallel()

	// When all non-excluded PS-IDs have swagger files (or there are none), expect success.
	tmpDir := t.TempDir()
	createAllPSIDsWithSwaggerFiles(t, tmpDir)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_NoExclusions_MissingSwagger(t *testing.T) {
	t.Parallel()

	// With no exclusions, all PS-IDs are checked. Create the apps dir but no swagger files.
	tmpDir := t.TempDir()

	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		require.NoError(t, os.MkdirAll(
			filepath.Join(tmpDir, "internal", "apps", ps.PSID, "server"),
			cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute,
		))
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := ExportedCheckInDirNoExclusions(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "swagger presence violations")
}

func TestCheckInDir_NoExclusions_AllValid(t *testing.T) {
	t.Parallel()

	// With no exclusions, all PS-IDs need swagger files.
	tmpDir := t.TempDir()

	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		serverDir := filepath.Join(tmpDir, "internal", "apps", ps.PSID, "server")
		require.NoError(t, os.MkdirAll(serverDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, "swagger.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, "swagger_test.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := ExportedCheckInDirNoExclusions(logger, tmpDir)
	require.NoError(t, err)
}

func TestCheckPSIDSwaggerFiles_MissingPSIDDir(t *testing.T) {
	t.Parallel()

	errs := ExportedCheckPSIDSwaggerFiles("/nonexistent/path", "my-psid")
	require.Len(t, errs, 1)
	require.Contains(t, errs[0], "PS-ID directory missing")
	require.Contains(t, errs[0], "my-psid")
}

// TestCheckPSIDSwaggerFiles_MissingSwaggerGo tests checkPSIDSwaggerFiles directly for missing swagger.go.
func TestCheckPSIDSwaggerFiles_MissingSwaggerGo(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	serverDir := filepath.Join(tmpDir, "server")
	require.NoError(t, os.MkdirAll(serverDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	// Only create swagger_test.go; swagger.go is absent.
	require.NoError(t, os.WriteFile(filepath.Join(serverDir, "swagger_test.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))

	errs := ExportedCheckPSIDSwaggerFiles(tmpDir, "test-psid")
	require.Len(t, errs, 1)
	require.Contains(t, errs[0], "swagger.go")
}

// TestCheckPSIDSwaggerFiles_MissingSwaggerTestGo tests checkPSIDSwaggerFiles directly for missing swagger_test.go.
func TestCheckPSIDSwaggerFiles_MissingSwaggerTestGo(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	serverDir := filepath.Join(tmpDir, "server")
	require.NoError(t, os.MkdirAll(serverDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	// Only create swagger.go; swagger_test.go is absent.
	require.NoError(t, os.WriteFile(filepath.Join(serverDir, "swagger.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))

	errs := ExportedCheckPSIDSwaggerFiles(tmpDir, "test-psid")
	require.Len(t, errs, 1)
	require.Contains(t, errs[0], "swagger_test.go")
}

// TestCheckPSIDSwaggerFiles_BothMissing tests checkPSIDSwaggerFiles when both files are absent.
func TestCheckPSIDSwaggerFiles_BothMissing(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	serverDir := filepath.Join(tmpDir, "server")
	require.NoError(t, os.MkdirAll(serverDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	// Neither swagger.go nor swagger_test.go.

	errs := ExportedCheckPSIDSwaggerFiles(tmpDir, "test-psid")
	require.Len(t, errs, 2)
}

// TestCheckPSIDSwaggerFiles_AllPresent tests checkPSIDSwaggerFiles when both files exist.
func TestCheckPSIDSwaggerFiles_AllPresent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	serverDir := filepath.Join(tmpDir, "server")
	require.NoError(t, os.MkdirAll(serverDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(serverDir, "swagger.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(serverDir, "swagger_test.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))

	errs := ExportedCheckPSIDSwaggerFiles(tmpDir, "test-psid")
	require.Empty(t, errs)
}

func TestCheckInDir_ExcludedPSIDPassesWithoutSwagger(t *testing.T) {
	t.Parallel()

	// Verify excluded PS-IDs don't trigger an error when swagger is absent.
	var excludedPSID string

	for psid := range knownExclusions {
		excludedPSID = psid

		break
	}

	if excludedPSID == "" {
		t.Skip("no excluded PS-IDs in knownExclusions")
	}

	tmpDir := t.TempDir()

	// Create only the excluded PS-ID directory (no swagger files).
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "internal", "apps", excludedPSID, "server"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	// Create all other PS-IDs fully (including swagger if non-excluded).
	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		if ps.PSID == excludedPSID {
			continue
		}

		serverDir := filepath.Join(tmpDir, "internal", "apps", ps.PSID, "server")
		require.NoError(t, os.MkdirAll(serverDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

		if !knownExclusions[ps.PSID] {
			require.NoError(t, os.WriteFile(filepath.Join(serverDir, "swagger.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
			require.NoError(t, os.WriteFile(filepath.Join(serverDir, "swagger_test.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
		}
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}
