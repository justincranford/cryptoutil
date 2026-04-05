// Copyright (c) 2025 Justin Cranford

package subcommand_completeness

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
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

// setupValidServiceDir creates a single service directory with a Go file that calls RouteService.
func setupValidServiceDir(t *testing.T, tmpDir, psID string) {
	t.Helper()

	serviceDir := filepath.Join(tmpDir, "internal", "apps", psID)
	require.NoError(t, os.MkdirAll(serviceDir, cryptoutilSharedMagic.CICDTempDirPermissions))

	content := "package " + psID + "\n\nimport \"cryptoutil/internal/apps/framework/service/cli\"\n\nfunc Handler() { cli.RouteService() }\n"
	require.NoError(t, os.WriteFile(filepath.Join(serviceDir, psID+".go"), []byte(content), cryptoutilSharedMagic.FilePermissions))
}

// setupAllValidServices creates valid service directories for all registry PS-IDs.
func setupAllValidServices(t *testing.T, tmpDir string) {
	t.Helper()

	for _, ps := range lintFitnessRegistry.AllProductServices() {
		setupValidServiceDir(t, tmpDir, ps.PSID)
	}
}

func TestCheck_DelegatesToCheckInDir(t *testing.T) {
	// Not parallel: changes process working directory.
	root := findProjectRoot(t)

	orig, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(root))

	defer func() { _ = os.Chdir(orig) }()

	err = Check(newTestLogger())
	require.NoError(t, err, "Check() should pass on real workspace")
}

func TestCheckInDir_RealWorkspace(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)

	err := CheckInDir(newTestLogger(), root, os.ReadDir, os.ReadFile)
	require.NoError(t, err, "all 10 registry services should use RouteService")
}

func TestCheckInDir_AllServicesValid(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupAllValidServices(t, tmpDir)

	err := CheckInDir(newTestLogger(), tmpDir, os.ReadDir, os.ReadFile)
	require.NoError(t, err)
}

func TestCheckInDir_MissingRouteService(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Set up all services first with valid RouteService calls.
	setupAllValidServices(t, tmpDir)

	// Overwrite the sm-kms production file with one that does NOT call RouteService.
	smKmsDir := filepath.Join(tmpDir, "internal", "apps", cryptoutilSharedMagic.OTLPServiceSMKMS)
	content := "package kms\n\nfunc Handler() {}\n"
	require.NoError(t, os.WriteFile(filepath.Join(smKmsDir, cryptoutilSharedMagic.OTLPServiceSMKMS+".go"), []byte(content), cryptoutilSharedMagic.FilePermissions))

	err := CheckInDir(newTestLogger(), tmpDir, os.ReadDir, os.ReadFile)
	require.Error(t, err)
	assert.Contains(t, err.Error(), cryptoutilSharedMagic.OTLPServiceSMKMS)
	assert.Contains(t, err.Error(), "RouteService")
}

func TestCheckInDir_MissingServiceDirectory(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Only create 9 of 10 services, leaving sm-kms missing.
	for _, ps := range lintFitnessRegistry.AllProductServices() {
		if ps.PSID == cryptoutilSharedMagic.OTLPServiceSMKMS {
			continue
		}

		setupValidServiceDir(t, tmpDir, ps.PSID)
	}

	err := CheckInDir(newTestLogger(), tmpDir, os.ReadDir, os.ReadFile)
	require.Error(t, err)
	assert.Contains(t, err.Error(), cryptoutilSharedMagic.OTLPServiceSMKMS)
	assert.Contains(t, err.Error(), "cannot read service directory")
}

func TestCheckInDir_OnlyTestFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Set up all services except sm-im which gets only _test.go files (no RouteService in production code).
	setupAllValidServices(t, tmpDir)

	smImDir := filepath.Join(tmpDir, "internal", "apps", cryptoutilSharedMagic.OTLPServiceSMIM)

	// Remove the valid Go file.
	require.NoError(t, os.Remove(filepath.Join(smImDir, cryptoutilSharedMagic.OTLPServiceSMIM+".go")))

	// Replace with only a test file that has RouteService (should NOT satisfy the check).
	testContent := "package im_test\n\nfunc TestRouteService() { _ = \"RouteService\" }\n"
	require.NoError(t, os.WriteFile(filepath.Join(smImDir, "im_test.go"), []byte(testContent), cryptoutilSharedMagic.FilePermissions))

	err := CheckInDir(newTestLogger(), tmpDir, os.ReadDir, os.ReadFile)
	require.Error(t, err)
	assert.Contains(t, err.Error(), cryptoutilSharedMagic.OTLPServiceSMIM)
}

func TestCheckInDir_MultipleViolations(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Only create service dirs WITHOUT RouteService for sm-kms and sm-im.
	for _, ps := range lintFitnessRegistry.AllProductServices() {
		serviceDir := filepath.Join(tmpDir, "internal", "apps", ps.PSID)
		require.NoError(t, os.MkdirAll(serviceDir, cryptoutilSharedMagic.CICDTempDirPermissions))

		var content string
		if ps.PSID == cryptoutilSharedMagic.OTLPServiceSMKMS || ps.PSID == cryptoutilSharedMagic.OTLPServiceSMIM {
			content = "package " + ps.PSID + "\n\nfunc Handler() {}\n"
		} else {
			content = "package " + ps.PSID + "\n\nfunc Handler() { _ = \"RouteService\" }\n"
		}

		require.NoError(t, os.WriteFile(filepath.Join(serviceDir, ps.PSID+".go"), []byte(content), cryptoutilSharedMagic.FilePermissions))
	}

	err := CheckInDir(newTestLogger(), tmpDir, os.ReadDir, os.ReadFile)
	require.Error(t, err)
	assert.Contains(t, err.Error(), cryptoutilSharedMagic.OTLPServiceSMKMS)
	assert.Contains(t, err.Error(), cryptoutilSharedMagic.OTLPServiceSMIM)
}

func TestCheckInDir_ReadFileError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupAllValidServices(t, tmpDir)

	smKmsFile := filepath.Join(tmpDir, "internal", "apps", cryptoutilSharedMagic.OTLPServiceSMKMS, cryptoutilSharedMagic.OTLPServiceSMKMS+".go")

	err := CheckInDir(newTestLogger(), tmpDir, os.ReadDir, func(path string) ([]byte, error) {
		if path == smKmsFile {
			return nil, os.ErrPermission
		}

		return os.ReadFile(path)
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), cryptoutilSharedMagic.OTLPServiceSMKMS)
	assert.Contains(t, err.Error(), "cannot read")
}

func TestCheckInDir_ReadDirError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupAllValidServices(t, tmpDir)

	err := CheckInDir(newTestLogger(), tmpDir, func(path string) ([]os.DirEntry, error) {
		if filepath.Base(path) == cryptoutilSharedMagic.OTLPServiceSMKMS {
			return nil, os.ErrPermission
		}

		return os.ReadDir(path)
	}, os.ReadFile)
	require.Error(t, err)
	assert.Contains(t, err.Error(), cryptoutilSharedMagic.OTLPServiceSMKMS)
}
