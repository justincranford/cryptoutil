// Copyright (c) 2025-2026 Justin Cranford.
package apps_product_no_service_dirs

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

// createCleanProductDirs creates product dirs with no service subdirs.
func createCleanProductDirs(t *testing.T, tmpDir string) {
	t.Helper()

	for _, product := range cryptoutilFitnessRegistry.AllProducts() {
		productDir := filepath.Join(tmpDir, "internal", "apps", product.ID)
		require.NoError(t, os.MkdirAll(productDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		// Create only a non-service subdir as a shared package placeholder.
		require.NoError(t, os.MkdirAll(filepath.Join(productDir, "shared"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
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

func TestCheckInDir_CleanProductDirs(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createCleanProductDirs(t, tmpDir)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

// TestCheckInDir_NoExclusions_ServiceDirFound verifies that a service-named subdir is reported
// when not in exclusions.
func TestCheckInDir_NoExclusions_ServiceDirFound(t *testing.T) {
	t.Parallel()

	// Find the first product+service pair to use as a violation target.
	services := cryptoutilFitnessRegistry.AllProductServices()
	if len(services) == 0 {
		t.Skip("no product services in registry")
	}

	target := services[0]

	tmpDir := t.TempDir()
	createCleanProductDirs(t, tmpDir)

	// Create a service-named subdir inside the product dir.
	serviceSubdir := filepath.Join(tmpDir, "internal", "apps", target.Product, target.Service)
	require.NoError(t, os.MkdirAll(serviceSubdir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := ExportedCheckInDirWithExclusions(logger, tmpDir, emptyExclusions)
	require.Error(t, err)
	require.Contains(t, err.Error(), "service-named subdir found in product directory")
	require.Contains(t, err.Error(), target.Service)
}

// TestCheckInDir_ExclusionSuppressesViolation verifies that an excluded service dir passes.
func TestCheckInDir_ExclusionSuppressesViolation(t *testing.T) {
	t.Parallel()

	services := cryptoutilFitnessRegistry.AllProductServices()
	if len(services) == 0 {
		t.Skip("no product services in registry")
	}

	target := services[0]

	tmpDir := t.TempDir()
	createCleanProductDirs(t, tmpDir)

	// Create a service-named subdir.
	serviceSubdir := filepath.Join(tmpDir, "internal", "apps", target.Product, target.Service)
	require.NoError(t, os.MkdirAll(serviceSubdir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	// Exclude it.
	exclusions := map[string]bool{target.Product + "/" + target.Service: true}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := ExportedCheckInDirWithExclusions(logger, tmpDir, exclusions)
	require.NoError(t, err)
}

// TestCheckInDir_MissingProductDirIsNotViolation verifies a missing product dir doesn't error here.
func TestCheckInDir_MissingProductDirIsNotViolation(t *testing.T) {
	t.Parallel()

	products := cryptoutilFitnessRegistry.AllProducts()
	if len(products) == 0 {
		t.Skip("no products in registry")
	}

	tmpDir := t.TempDir()

	// Create apps dir but NOT the product dir.
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "internal", "apps"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

// TestBuildProductServiceMap verifies the map is correctly built from the registry.
func TestBuildProductServiceMap(t *testing.T) {
	t.Parallel()

	m := ExportedBuildProductServiceMap()

	require.NotEmpty(t, m)

	// Every ProductService from the registry must be in the map.
	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		require.True(t, m[ps.Product][ps.Service], "expected %s/%s in map", ps.Product, ps.Service)
	}
}

// TestCheckProductDir_ReadError tests the error path when reading the directory fails.
func TestCheckProductDir_ReadError(t *testing.T) {
	t.Parallel()

	// Create a directory and immediately make it unreadable by creating a file at the path
	// we pass as productDir but pointing to a non-directory.
	tmpDir := t.TempDir()
	fileAsDir := filepath.Join(tmpDir, "notadir")
	require.NoError(t, os.WriteFile(fileAsDir, []byte("x"), cryptoutilSharedMagic.CacheFilePermissions))

	errs := ExportedCheckProductDir(fileAsDir, "testprod", map[string]bool{"svc": true}, emptyExclusions)
	require.NotEmpty(t, errs)
	require.Contains(t, errs[0], "cannot read product directory")
}
