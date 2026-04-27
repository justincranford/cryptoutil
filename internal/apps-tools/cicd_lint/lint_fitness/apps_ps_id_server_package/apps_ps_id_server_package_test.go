// Copyright (c) 2025 Justin Cranford

package apps_ps_id_server_package

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

// createAllPSIDsWithServerFiles creates server/server.go and (where applicable) server/public_server.go.
func createAllPSIDsWithServerFiles(t *testing.T, tmpDir string) {
	t.Helper()

	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		serverDir := filepath.Join(tmpDir, "internal", "apps", ps.PSID, "server")
		require.NoError(t, os.MkdirAll(serverDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, "server.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))

		if !knownExclusionsPublicServer[ps.PSID] {
			require.NoError(t, os.WriteFile(filepath.Join(serverDir, "public_server.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
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

func TestCheckInDir_AllValid(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createAllPSIDsWithServerFiles(t, tmpDir)

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
		serverDir := filepath.Join(tmpDir, "internal", "apps", ps.PSID, "server")
		require.NoError(t, os.MkdirAll(serverDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, "server.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))

		if !knownExclusionsPublicServer[ps.PSID] {
			require.NoError(t, os.WriteFile(filepath.Join(serverDir, "public_server.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
		}
	}

	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "internal", "apps"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "PS-ID directory missing")
	require.Contains(t, err.Error(), target.PSID)
}

func TestCheckInDir_MissingServerGo(t *testing.T) {
	t.Parallel()

	services := cryptoutilFitnessRegistry.AllProductServices()
	if len(services) == 0 {
		t.Skip("no product services in registry")
	}

	target := services[0]

	tmpDir := t.TempDir()
	createAllPSIDsWithServerFiles(t, tmpDir)

	// Remove server.go for the first PS-ID.
	require.NoError(t, os.Remove(filepath.Join(tmpDir, "internal", "apps", target.PSID, "server", "server.go")))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing required file")
	require.Contains(t, err.Error(), "server/server.go")
}

func TestCheckInDir_MissingPublicServerGo(t *testing.T) {
	t.Parallel()

	// Find a PS-ID that is NOT excluded from public_server.go check.
	var target *cryptoutilFitnessRegistry.ProductService

	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		if !knownExclusionsPublicServer[ps.PSID] {
			target = &ps

			break
		}
	}

	if target == nil {
		t.Skip("no non-excluded PS-IDs found")
	}

	tmpDir := t.TempDir()
	createAllPSIDsWithServerFiles(t, tmpDir)

	// Remove public_server.go for the target PS-ID.
	require.NoError(t, os.Remove(filepath.Join(tmpDir, "internal", "apps", target.PSID, "server", "public_server.go")))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing required file")
	require.Contains(t, err.Error(), "server/public_server.go")
}

func TestCheckInDir_ExcludedPSIDPassesWithoutPublicServer(t *testing.T) {
	t.Parallel()

	// Verify that an excluded PS-ID passes even when public_server.go is absent.
	var excludedPSID string

	for psid := range knownExclusionsPublicServer {
		excludedPSID = psid

		break
	}

	if excludedPSID == "" {
		t.Skip("no excluded PS-IDs in knownExclusionsPublicServer")
	}

	tmpDir := t.TempDir()
	createAllPSIDsWithServerFiles(t, tmpDir)

	// Verify excluded PS-ID doesn't have public_server.go (it shouldn't since createAllPSIDsWithServerFiles skips it).
	publicServerPath := filepath.Join(tmpDir, "internal", "apps", excludedPSID, "server", "public_server.go")
	_, statErr := os.Stat(publicServerPath)
	require.True(t, os.IsNotExist(statErr), "excluded PS-ID should not have public_server.go in test fixture")

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}
