// Copyright (c) 2025 Justin Cranford

package configs_naming_test

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	lintFitnessConfigsNaming "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/configs_naming"
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

func setupConfigsDir(t *testing.T, tmpDir string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, cryptoutilSharedMagic.CICDConfigsDir), cryptoutilSharedMagic.DirPermissions))
}

func createProductDir(t *testing.T, tmpDir, product string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, cryptoutilSharedMagic.CICDConfigsDir, product), cryptoutilSharedMagic.DirPermissions))
}

func createServiceDir(t *testing.T, tmpDir, product, service string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, cryptoutilSharedMagic.CICDConfigsDir, product, service), cryptoutilSharedMagic.DirPermissions))
}

func writeConfigFile(t *testing.T, tmpDir string, relPath string, content string) {
	t.Helper()

	fp := filepath.Join(tmpDir, relPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(fp), cryptoutilSharedMagic.DirPermissions))
	require.NoError(t, os.WriteFile(fp, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))
}

// TestFindViolationsInDir_EmptyConfigsDir verifies no violations on empty configs/ dir.
func TestFindViolationsInDir_EmptyConfigsDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupConfigsDir(t, tmpDir)

	violations, err := lintFitnessConfigsNaming.FindViolationsInDir(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, violations)
}

// TestFindViolationsInDir_NonExistentConfigsDir verifies error when configs/ dir is missing.
func TestFindViolationsInDir_NonExistentConfigsDir(t *testing.T) {
	t.Parallel()

	violations, err := lintFitnessConfigsNaming.FindViolationsInDir("/nonexistent/path/does/not/exist")
	require.Error(t, err)
	assert.Nil(t, violations)
}

// TestFindViolationsInDir_ValidSuiteDir verifies suite-level dir (cryptoutil/) is allowed.
func TestFindViolationsInDir_ValidSuiteDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	writeConfigFile(t, tmpDir, filepath.Join(cryptoutilSharedMagic.CICDConfigsDir, cryptoutilSharedMagic.DefaultOTLPServiceDefault, "cryptoutil.yml"), "# config\n")

	violations, err := lintFitnessConfigsNaming.FindViolationsInDir(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, violations)
}

// TestFindViolationsInDir_ValidProductDirs verifies all 5 product dirs are allowed.
func TestFindViolationsInDir_ValidProductDirs(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupConfigsDir(t, tmpDir)

	for _, product := range []string{
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.JoseProductName,
		cryptoutilSharedMagic.PKIProductName,
		cryptoutilSharedMagic.SkeletonProductName,
		cryptoutilSharedMagic.SMProductName,
	} {
		createProductDir(t, tmpDir, product)
	}

	violations, err := lintFitnessConfigsNaming.FindViolationsInDir(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, violations)
}

// TestFindViolationsInDir_UnknownTopLevelDir verifies unknown top-level dirs are violations.
func TestFindViolationsInDir_UnknownTopLevelDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupConfigsDir(t, tmpDir)
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, cryptoutilSharedMagic.CICDConfigsDir, "legacy"), cryptoutilSharedMagic.DirPermissions))

	violations, err := lintFitnessConfigsNaming.FindViolationsInDir(tmpDir)
	require.NoError(t, err)
	require.Len(t, violations, 1)
	assert.Contains(t, violations[0], "configs/legacy")
	assert.Contains(t, violations[0], "unknown product or suite directory")
}

// TestFindViolationsInDir_UnknownTopLevelDir_Multiple verifies multiple unknown dirs are all reported.
func TestFindViolationsInDir_UnknownTopLevelDir_Multiple(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupConfigsDir(t, tmpDir)
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, cryptoutilSharedMagic.CICDConfigsDir, "legacy"), cryptoutilSharedMagic.DirPermissions))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, cryptoutilSharedMagic.CICDConfigsDir, "old"), cryptoutilSharedMagic.DirPermissions))

	violations, err := lintFitnessConfigsNaming.FindViolationsInDir(tmpDir)
	require.NoError(t, err)
	require.Len(t, violations, 2)
}

// TestFindViolationsInDir_ValidServiceDirs verifies known service dirs are allowed.
func TestFindViolationsInDir_ValidServiceDirs(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	createServiceDir(t, tmpDir, cryptoutilSharedMagic.SMProductName, cryptoutilSharedMagic.IMServiceName)
	createServiceDir(t, tmpDir, cryptoutilSharedMagic.SMProductName, cryptoutilSharedMagic.KMSServiceName)

	violations, err := lintFitnessConfigsNaming.FindViolationsInDir(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, violations)
}

// TestFindViolationsInDir_UnknownServiceDir_WithPSIDFiles verifies violation for misnamed service dir with PS-ID files.
func TestFindViolationsInDir_UnknownServiceDir_WithPSIDFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	writeConfigFile(t, tmpDir, filepath.Join(cryptoutilSharedMagic.CICDConfigsDir, cryptoutilSharedMagic.SMProductName, "unknown", cryptoutilSharedMagic.SMProductName+"-unknown-sqlite.yml"), "# wrong service\n")

	violations, err := lintFitnessConfigsNaming.FindViolationsInDir(tmpDir)
	require.NoError(t, err)
	require.Len(t, violations, 1)
	assert.Contains(t, violations[0], "configs/"+cryptoutilSharedMagic.SMProductName+"/unknown")
	assert.Contains(t, violations[0], "unknown service directory")
}

// TestFindViolationsInDir_ProductLevelSpecialDir verifies product-level special dirs with non-PS-ID files are allowed.
func TestFindViolationsInDir_ProductLevelSpecialDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	writeConfigFile(t, tmpDir, filepath.Join(cryptoutilSharedMagic.CICDConfigsDir, cryptoutilSharedMagic.IdentityProductName, "policies", "adaptive-auth.yml"), "# policy\n")

	violations, err := lintFitnessConfigsNaming.FindViolationsInDir(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, violations)
}

// TestFindViolationsInDir_ProductLevelFilesAllowed verifies files directly in product dir are allowed.
func TestFindViolationsInDir_ProductLevelFilesAllowed(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	writeConfigFile(t, tmpDir, filepath.Join(cryptoutilSharedMagic.CICDConfigsDir, cryptoutilSharedMagic.SkeletonProductName, "skeleton-server.yml"), "# skeleton\n")

	violations, err := lintFitnessConfigsNaming.FindViolationsInDir(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, violations)
}

// TestFindViolationsInDir_UnknownServiceDir_NoProductFiles verifies product-level unknown dirs without PS-ID files are allowed.
func TestFindViolationsInDir_UnknownServiceDir_NoProductFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, cryptoutilSharedMagic.CICDConfigsDir, cryptoutilSharedMagic.IdentityProductName, "policies"), cryptoutilSharedMagic.DirPermissions))

	violations, err := lintFitnessConfigsNaming.FindViolationsInDir(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, violations)
}

// TestFindViolationsInDir_FilesInConfigsRootIgnored verifies files (not dirs) directly in configs/ are ignored.
func TestFindViolationsInDir_FilesInConfigsRootIgnored(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupConfigsDir(t, tmpDir)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, cryptoutilSharedMagic.CICDConfigsDir, "README.md"), []byte("# readme\n"), cryptoutilSharedMagic.CacheFilePermissions))

	violations, err := lintFitnessConfigsNaming.FindViolationsInDir(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, violations)
}

// TestCheckInDir_ValidStructure verifies CheckInDir passes on a valid structure.
func TestCheckInDir_ValidStructure(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupConfigsDir(t, tmpDir)
	createServiceDir(t, tmpDir, cryptoutilSharedMagic.SMProductName, cryptoutilSharedMagic.KMSServiceName)

	err := lintFitnessConfigsNaming.CheckInDir(newTestLogger(), tmpDir)
	require.NoError(t, err)
}

// TestCheckInDir_InvalidStructure verifies CheckInDir fails on violations.
func TestCheckInDir_InvalidStructure(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupConfigsDir(t, tmpDir)
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, cryptoutilSharedMagic.CICDConfigsDir, "bad"), cryptoutilSharedMagic.DirPermissions))

	err := lintFitnessConfigsNaming.CheckInDir(newTestLogger(), tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "configs/bad")
}

// TestCheckInDir_NonExistentRoot_ReturnsError verifies CheckInDir error on missing dir.
func TestCheckInDir_NonExistentRoot_ReturnsError(t *testing.T) {
	t.Parallel()

	err := lintFitnessConfigsNaming.CheckInDir(newTestLogger(), "/nonexistent/path/does/not/exist")
	require.Error(t, err)
}

// TestCheck_Integration runs the linter against the real workspace.
func TestCheck_Integration(t *testing.T) {
	root := findProjectRoot(t)

	err := lintFitnessConfigsNaming.CheckInDir(newTestLogger(), root)
	require.NoError(t, err)
}

// TestCheck_FromWorkspaceRoot verifies Check() (no rootDir) works from project root.
// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheck_FromWorkspaceRoot(t *testing.T) {
	root := findProjectRoot(t)

	origDir, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(root))

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	err = lintFitnessConfigsNaming.Check(newTestLogger())
	require.NoError(t, err)
}
