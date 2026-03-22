// Copyright (c) 2025 Justin Cranford

package admin_bind_address

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"

	"github.com/stretchr/testify/require"
)

func newTestLogger() *cryptoutilCmdCicdCommon.Logger {
	return cryptoutilCmdCicdCommon.NewLogger("test")
}

func writeGoFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	require.NoError(t, os.MkdirAll(dir, cryptoutilSharedMagic.DirPermissions))
	p := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(p, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	return p
}

func TestCheckInDir_CleanConfig_Passes(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	writeGoFile(t, tmp, "config.go", "package config\n\nvar defaults = struct{ BindPrivateAddress string }{BindPrivateAddress: \"127.0.0.1\"}\n")
	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_AdminBindZeroZero_Fails(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	writeGoFile(t, tmp, "config.go", "package config\n\nvar bad = struct{ BindPrivateAddress string }{BindPrivateAddress: \"0.0.0.0\"}\n")
	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "admin bind address")
}

func TestCheckInDir_CommentLine_Skipped(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	// Comment mentioning the bad pattern should not be flagged.
	writeGoFile(t, tmp, "doc.go", "package doc\n\n// BindPrivateAddress: \"0.0.0.0\" is NOT recommended, use 127.0.0.1.\nfunc Good() {}\n")
	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_NonGoFile_Ignored(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	p := filepath.Join(tmp, "config.yaml")
	require.NoError(t, os.WriteFile(p, []byte("BindPrivateAddress: \"0.0.0.0\"\n"), cryptoutilSharedMagic.CacheFilePermissions))

	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_TestFile_Skipped(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	// Test files may contain "0.0.0.0" as test fixture strings; they are excluded.
	writeGoFile(t, tmp, "config_test.go", "package config_test\n\nvar bad = struct{ BindPrivateAddress string }{BindPrivateAddress: \"0.0.0.0\"}\n")
	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_VendorDir_Skipped(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	vendorDir := filepath.Join(tmp, cryptoutilSharedMagic.CICDExcludeDirVendor)
	writeGoFile(t, vendorDir, "dep.go", "package dep\n\nvar c = struct{ BindPrivateAddress string }{BindPrivateAddress: \"0.0.0.0\"}\n")

	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_EmptyDir_Passes(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestScanForAdminBindViolations_Clean(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	p := writeGoFile(t, tmp, "config.go", "package config\nvar x = \"BindPrivateAddress 127.0.0.1\"\n")
	violations, err := scanForAdminBindViolations(p, tmp)
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestScanForAdminBindViolations_WithViolation(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	p := writeGoFile(t, tmp, "bad.go", "package bad\n\nvar cfg = struct{ BindPrivateAddress string }{BindPrivateAddress: \"0.0.0.0\"}\n")
	violations, err := scanForAdminBindViolations(p, tmp)
	require.NoError(t, err)
	require.NotEmpty(t, violations)
}

func TestScanForAdminBindViolations_NonexistentFile_Error(t *testing.T) {
	t.Parallel()

	_, err := scanForAdminBindViolations("/nonexistent/file.go", "/tmp")
	require.Error(t, err)
}

func TestCheckInDir_GitDir_Skipped(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	gitDir := filepath.Join(tmp, cryptoutilSharedMagic.CICDExcludeDirGit)
	writeGoFile(t, gitDir, "hook.go", "package git\nvar x = struct{ BindPrivateAddress string }{BindPrivateAddress: \"0.0.0.0\"}\n")

	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

// Sequential: modifies package-level adminBindWalkFn seam.
func TestCheckInDir_WalkError(t *testing.T) {
	orig := adminBindWalkFn

	t.Cleanup(func() { adminBindWalkFn = orig })

	adminBindWalkFn = func(_ string, _ filepath.WalkFunc) error {
		return fmt.Errorf("injected walk error")
	}

	err := CheckInDir(newTestLogger(), t.TempDir())
	require.Error(t, err)
	require.Contains(t, err.Error(), "filesystem walk failed")
}

// Sequential: modifies package-level adminBindWalkFn seam.
func TestCheckInDir_WalkCallbackError(t *testing.T) {
	orig := adminBindWalkFn

	t.Cleanup(func() { adminBindWalkFn = orig })

	adminBindWalkFn = func(_ string, fn filepath.WalkFunc) error {
		return fn("bad/path", nil, fmt.Errorf("injected callback error"))
	}

	err := CheckInDir(newTestLogger(), t.TempDir())
	require.Error(t, err)
	require.Contains(t, err.Error(), "filesystem walk failed")
}

// Sequential: modifies package-level adminBindOpenFn seam.
func TestCheckInDir_ScanOpenError(t *testing.T) {
	orig := adminBindOpenFn

	t.Cleanup(func() { adminBindOpenFn = orig })

	adminBindOpenFn = func(_ string) (*os.File, error) {
		return nil, fmt.Errorf("injected open error")
	}

	tmp := t.TempDir()
	writeGoFile(t, tmp, "main.go", "package main\n")

	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "filesystem walk failed")
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

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

// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheck_Integration(t *testing.T) {
	root, err := findProjectRoot()
	if err != nil {
		t.Skip("Skipping - cannot find project root")
	}

	origDir, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(root))

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-admin-bind-address")

	err = Check(logger)
	require.NoError(t, err)
}
