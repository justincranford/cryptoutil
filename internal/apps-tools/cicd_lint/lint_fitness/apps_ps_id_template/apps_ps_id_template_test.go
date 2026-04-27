// Copyright (c) 2025 Justin Cranford

package apps_ps_id_template

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilFitnessRegistry "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// findProjectRoot traverses up from the current directory to locate go.mod.
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

// copyManifest copies the real PS-ID MANIFEST.yaml into a synthetic root directory.
func copyManifest(t *testing.T, realRoot, tmpDir string) {
	t.Helper()

	const templateDirName = cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID

	srcPath := filepath.Join(realRoot, "api", "cryptosuite-registry", "templates", "internal", "apps", templateDirName, "MANIFEST.yaml")
	destDir := filepath.Join(tmpDir, "api", "cryptosuite-registry", "templates", "internal", "apps", templateDirName)

	require.NoError(t, os.MkdirAll(destDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	data, err := os.ReadFile(srcPath)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(destDir, "MANIFEST.yaml"), data, cryptoutilSharedMagic.CacheFilePermissions))
}

// createFullPSIDRoot creates all required files for all PS-IDs according to the manifest +
// per-PS-ID exclusion maps (matching the production exclusions).
func createFullPSIDRoot(t *testing.T, realRoot, tmpDir string) {
	t.Helper()

	copyManifest(t, realRoot, tmpDir)

	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		psDir := filepath.Join(tmpDir, "internal", "apps", ps.PSID)
		serverDir := filepath.Join(psDir, "server")

		require.NoError(t, os.MkdirAll(serverDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

		// Required root files (respecting production exclusions).
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+".go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+"_usage.go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))

		if !knownRootFileExclusions["__SERVICE___cli_test.go"][ps.PSID] {
			require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+"_cli_test.go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
		}

		// Required server file: server.go (no exclusions).
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, "server.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))

		// testmain_test.go: create only for non-excluded PS-IDs.
		if !knownServerFileExclusions["testmain_test.go"][ps.PSID] {
			require.NoError(t, os.WriteFile(filepath.Join(serverDir, "testmain_test.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
		}
		// swagger.go, swagger_test.go, lifecycle_test, port_conflict_test: all excluded, skip.
	}
}

// TestCheck_RealWorkspace verifies the linter passes against the actual workspace.
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

// TestCheckInDir_NoManifest exercises the "manifest not found" error path.
func TestCheckInDir_NoManifest(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, t.TempDir())
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read PS-ID MANIFEST.yaml")
}

// TestCheckInDir_InvalidManifest exercises the YAML parse error path.
func TestCheckInDir_InvalidManifest(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	manifestDir := filepath.Join(tmpDir, "api", "cryptosuite-registry", "templates", "internal", "apps", cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID)

	require.NoError(t, os.MkdirAll(manifestDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(manifestDir, "MANIFEST.yaml"), []byte(":\tinvalid::yaml{"), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse PS-ID MANIFEST.yaml")
}

// TestCheckInDir_NoAppsDir exercises the "internal/apps not found" error path.
func TestCheckInDir_NoAppsDir(t *testing.T) {
	t.Parallel()

	root, err := findProjectRoot()
	if err != nil {
		t.Skip("cannot find project root")
	}

	// Use a tmpDir that has the MANIFEST (borrowed from real root) but no internal/apps/.
	tmpDir := t.TempDir()
	copyManifest(t, root, tmpDir)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "internal/apps directory not found")
}

// TestCheckInDir_WithExclusions_AllPass verifies the linter passes when all non-excluded
// PS-IDs have their required files.
func TestCheckInDir_WithExclusions_AllPass(t *testing.T) {
	t.Parallel()

	root, err := findProjectRoot()
	if err != nil {
		t.Skip("cannot find project root")
	}

	tmpDir := t.TempDir()
	createFullPSIDRoot(t, root, tmpDir)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

// TestCheckInDir_NoExclusions_MissingRootFile exercises the root-file violation path.
func TestCheckInDir_NoExclusions_MissingRootFile(t *testing.T) {
	t.Parallel()

	root, err := findProjectRoot()
	if err != nil {
		t.Skip("cannot find project root")
	}

	tmpDir := t.TempDir()
	copyManifest(t, root, tmpDir)

	// Create server dirs for all PS-IDs but omit all root files.
	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		require.NoError(t, os.MkdirAll(
			filepath.Join(tmpDir, "internal", "apps", ps.PSID, "server"),
			cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute,
		))
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = ExportedCheckInDirNoExclusions(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing required root file")
}

// TestCheckInDir_NoExclusions_MissingRequiredDir exercises the required-dir violation path.
func TestCheckInDir_NoExclusions_MissingRequiredDir(t *testing.T) {
	t.Parallel()

	root, err := findProjectRoot()
	if err != nil {
		t.Skip("cannot find project root")
	}

	tmpDir := t.TempDir()
	copyManifest(t, root, tmpDir)

	// Create PS-ID dirs with root files but no server/ dir.
	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		psDir := filepath.Join(tmpDir, "internal", "apps", ps.PSID)
		require.NoError(t, os.MkdirAll(psDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+".go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+"_usage.go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+"_cli_test.go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = ExportedCheckInDirNoExclusions(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing required directory")
}

// TestCheckInDir_NoExclusions_MissingServerFile exercises the server-file violation path.
func TestCheckInDir_NoExclusions_MissingServerFile(t *testing.T) {
	t.Parallel()

	root, err := findProjectRoot()
	if err != nil {
		t.Skip("cannot find project root")
	}

	tmpDir := t.TempDir()
	copyManifest(t, root, tmpDir)

	// Create full root files and server/ dir but no server files.
	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		psDir := filepath.Join(tmpDir, "internal", "apps", ps.PSID)
		serverDir := filepath.Join(psDir, "server")
		require.NoError(t, os.MkdirAll(serverDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+".go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+"_usage.go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+"_cli_test.go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
		// No server files created.
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = ExportedCheckInDirNoExclusions(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing required server file")
}

// TestCheckInDir_NoExclusions_AllValid verifies no violations when all required files present.
func TestCheckInDir_NoExclusions_AllValid(t *testing.T) {
	t.Parallel()

	root, err := findProjectRoot()
	if err != nil {
		t.Skip("cannot find project root")
	}

	tmpDir := t.TempDir()
	copyManifest(t, root, tmpDir)

	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		psDir := filepath.Join(tmpDir, "internal", "apps", ps.PSID)
		serverDir := filepath.Join(psDir, "server")
		require.NoError(t, os.MkdirAll(serverDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

		// All required root files.
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+".go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+"_usage.go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+"_cli_test.go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))

		// All required server files.
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, "server.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, "swagger.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, "swagger_test.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, "testmain_test.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, ps.Service+"_lifecycle_test.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, ps.Service+"_port_conflict_test.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = ExportedCheckInDirNoExclusions(logger, tmpDir)
	require.NoError(t, err)
}
