// Copyright (c) 2025 Justin Cranford

package test_presence

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
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

func TestCheckInDir_AllValid(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	pkgDir := filepath.Join(tmpDir, "internal", "apps", "myproduct")
	require.NoError(t, os.MkdirAll(pkgDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "logic.go"), []byte("package myproduct\nfunc Hello() {}"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "logic_test.go"), []byte("package myproduct"), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_MissingTestFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	pkgDir := filepath.Join(tmpDir, "internal", "apps", "myproduct")
	require.NoError(t, os.MkdirAll(pkgDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "logic.go"), []byte("package myproduct\nfunc Hello() {}"), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no _test.go files")
}

func TestCheckInDir_ExcludedMagicDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	magicDir := filepath.Join(tmpDir, "internal", "apps", "shared", "magic")
	require.NoError(t, os.MkdirAll(magicDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(magicDir, "constants.go"), []byte("package magic\nfunc X() {}"), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_ExcludedUnifiedDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	unifiedDir := filepath.Join(tmpDir, "internal", "apps", "jose", "ja", "unified")
	require.NoError(t, os.MkdirAll(unifiedDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(unifiedDir, "server.go"), []byte("package unified\nfunc X() {}"), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_ExcludedArchivedDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	archivedDir := filepath.Join(tmpDir, "internal", "apps", "pki", "_ca-archived", "server")
	require.NoError(t, os.MkdirAll(archivedDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(archivedDir, "server.go"), []byte("package server\nfunc X() {}"), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_ExcludedTestingDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testingDir := filepath.Join(tmpDir, "internal", "apps", "template", "service", "testing", "helpers")
	require.NoError(t, os.MkdirAll(testingDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(testingDir, "helpers.go"), []byte("package helpers\nfunc X() {}"), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_ExcludedCmdMainDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cmdMainDir := filepath.Join(tmpDir, "internal", "apps", "identity", "cmd", "main", "authz")
	require.NoError(t, os.MkdirAll(cmdMainDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(cmdMainDir, "authz.go"), []byte("package authz\nfunc X() {}"), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_GeneratedFilesIgnored(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	pkgDir := filepath.Join(tmpDir, "internal", "apps", "myproduct")
	require.NoError(t, os.MkdirAll(pkgDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "server.gen.go"), []byte("package myproduct\nfunc X() {}"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "models_gen.go"), []byte("package myproduct\nfunc X() {}"), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_EmbedOnlyPackageSkipped(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	pkgDir := filepath.Join(tmpDir, "internal", "apps", "myproduct", "repository")
	require.NoError(t, os.MkdirAll(pkgDir, 0o755))

	// File with only embed.FS variable declaration (no func declarations).
	content := "package repository\nimport \"embed\"\n//go:embed migrations/*.sql\nvar MigrationsFS embed.FS\n"
	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "migrations.go"), []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_NoAppsDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "internal/apps directory not found")
}

func TestCheckDirForGoFiles_EmptyDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	hasLogic, hasTests := checkDirForGoFiles(tmpDir)
	require.False(t, hasLogic)
	require.False(t, hasTests)
}

func TestCheckDirForGoFiles_NonexistentDir(t *testing.T) {
	t.Parallel()

	hasLogic, hasTests := checkDirForGoFiles("/nonexistent")
	require.False(t, hasLogic)
	require.False(t, hasTests)
}

func TestCheckDirForGoFiles_OnlyTestFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "logic_test.go"), []byte("package pkg"), cryptoutilSharedMagic.CacheFilePermissions))

	hasLogic, hasTests := checkDirForGoFiles(tmpDir)
	require.False(t, hasLogic)
	require.True(t, hasTests)
}

func TestFileHasLogic_WithFunc(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	f := filepath.Join(tmpDir, "logic.go")
	require.NoError(t, os.WriteFile(f, []byte("package pkg\nfunc Hello() {}"), cryptoutilSharedMagic.CacheFilePermissions))

	require.True(t, fileHasLogic(f))
}

func TestFileHasLogic_WithoutFunc(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	f := filepath.Join(tmpDir, "vars.go")
	require.NoError(t, os.WriteFile(f, []byte("package pkg\nvar X = 5"), cryptoutilSharedMagic.CacheFilePermissions))

	require.False(t, fileHasLogic(f))
}

func TestFileHasLogic_NonexistentFile(t *testing.T) {
	t.Parallel()

	require.True(t, fileHasLogic("/nonexistent/file.go"))
}

func TestCheckInDir_MultiplePackages(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Package with logic and tests.
	pkg1 := filepath.Join(tmpDir, "internal", "apps", "prod1")
	require.NoError(t, os.MkdirAll(pkg1, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(pkg1, "code.go"), []byte("package prod1\nfunc X() {}"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(pkg1, "code_test.go"), []byte("package prod1"), cryptoutilSharedMagic.CacheFilePermissions))

	// Package with logic but missing tests.
	pkg2 := filepath.Join(tmpDir, "internal", "apps", "prod2")
	require.NoError(t, os.MkdirAll(pkg2, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(pkg2, "code.go"), []byte("package prod2\nfunc Y() {}"), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "prod2")
}

func TestCheck_FromProjectRoot(t *testing.T) {
	root, err := findProjectRoot()
	if err != nil {
		t.Skip("Skipping - cannot find project root")
	}

	origDir, wdErr := os.Getwd()
	require.NoError(t, wdErr)

	require.NoError(t, os.Chdir(root))

	t.Cleanup(func() {
		require.NoError(t, os.Chdir(origDir))
	})

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger)
	require.NoError(t, err)
}

func TestCheckInDir_WalkPermissionError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	appsDir := filepath.Join(tmpDir, "internal", "apps")
	badDir := filepath.Join(appsDir, "badpackage")
	require.NoError(t, os.MkdirAll(badDir, 0o755))
	require.NoError(t, os.Chmod(badDir, 0o000))

	t.Cleanup(func() {
		_ = os.Chmod(badDir, 0o755)
	})

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to walk internal/apps")
}

// Seam tests below: NOT parallel because they modify package-level seam variables.

func saveRestoreSeams(t *testing.T) {
	t.Helper()

	origRel := relFunc

	t.Cleanup(func() {
		relFunc = origRel
	})
}

func TestSeam_CheckInDir_RelError(t *testing.T) {
	saveRestoreSeams(t)

	relFunc = func(_, _ string) (string, error) {
		return "", fmt.Errorf("injected rel error")
	}

	tmpDir := t.TempDir()
	appsDir := filepath.Join(tmpDir, "internal", "apps")
	pkgDir := filepath.Join(appsDir, "myservice")
	require.NoError(t, os.MkdirAll(pkgDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "handler.go"), []byte("package myservice\nfunc Handle() {}\n"), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "test presence violations")
}
