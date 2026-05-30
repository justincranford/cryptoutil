// Copyright (c) 2025-2026 Justin Cranford.
package apps_product_template

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

// copyManifest copies the real product MANIFEST.yaml into a synthetic root directory.
func copyManifest(t *testing.T, realRoot, tmpDir string) {
	t.Helper()

	srcPath := filepath.Join(realRoot, "api", "cryptosuite-registry", "templates", "internal", "apps", cryptoutilSharedMagic.CICDTemplateExpansionKeyProduct, "MANIFEST.yaml")
	destDir := filepath.Join(tmpDir, "api", "cryptosuite-registry", "templates", "internal", "apps", cryptoutilSharedMagic.CICDTemplateExpansionKeyProduct)

	require.NoError(t, os.MkdirAll(destDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	data, err := os.ReadFile(srcPath)
	require.NoError(t, err)

	root, err := os.OpenRoot(destDir)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, root.Close())
	}()

	require.NoError(t, root.WriteFile("MANIFEST.yaml", data, cryptoutilSharedMagic.CacheFilePermissions))
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
	require.Contains(t, err.Error(), "failed to read product MANIFEST.yaml")
}

// TestCheckInDir_InvalidManifest exercises the YAML parse error path.
func TestCheckInDir_InvalidManifest(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	manifestDir := filepath.Join(tmpDir, "api", "cryptosuite-registry", "templates", "internal", "apps", cryptoutilSharedMagic.CICDTemplateExpansionKeyProduct)

	require.NoError(t, os.MkdirAll(manifestDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(manifestDir, "MANIFEST.yaml"), []byte(":\tinvalid::yaml{"), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse product MANIFEST.yaml")
}

// TestCheckInDir_NoAppsDir exercises the "internal/apps not found" error path.
func TestCheckInDir_NoAppsDir(t *testing.T) {
	t.Parallel()

	root, err := findProjectRoot()
	if err != nil {
		t.Skip("cannot find project root")
	}

	tmpDir := t.TempDir()
	copyManifest(t, root, tmpDir)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "internal/apps directory not found")
}

// TestCheckInDir_MissingProductFile exercises the missing root-file violation path.
func TestCheckInDir_MissingProductFile(t *testing.T) {
	t.Parallel()

	root, err := findProjectRoot()
	if err != nil {
		t.Skip("cannot find project root")
	}

	tmpDir := t.TempDir()
	copyManifest(t, root, tmpDir)

	// Create product dirs but omit all root files.
	for _, product := range cryptoutilFitnessRegistry.AllProducts() {
		require.NoError(t, os.MkdirAll(
			filepath.Join(tmpDir, "internal", "apps", product.ID),
			cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute,
		))
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing required root file")
}

// TestCheckInDir_AllValid verifies no violations when all required product files are present.
func TestCheckInDir_AllValid(t *testing.T) {
	t.Parallel()

	root, err := findProjectRoot()
	if err != nil {
		t.Skip("cannot find project root")
	}

	tmpDir := t.TempDir()
	copyManifest(t, root, tmpDir)

	for _, product := range cryptoutilFitnessRegistry.AllProducts() {
		productDir := filepath.Join(tmpDir, "internal", "apps", product.ID)
		require.NoError(t, os.MkdirAll(productDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.WriteFile(filepath.Join(productDir, product.ID+".go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(productDir, product.ID+"_test.go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}
