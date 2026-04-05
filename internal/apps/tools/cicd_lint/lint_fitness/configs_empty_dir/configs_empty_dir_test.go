// Copyright (c) 2025 Justin Cranford

package configs_empty_dir_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessConfigsEmptyDir "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/configs_empty_dir"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func newTestLogger() *cryptoutilCmdCicdCommon.Logger {
	return cryptoutilCmdCicdCommon.NewLogger("configs-empty-dir-test")
}

// mkdir creates a directory path relative to the test root.
func mkdir(t *testing.T, path string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(path, cryptoutilSharedMagic.CICDTempDirPermissions))
}

// touch creates an empty file at path, creating parent directories as needed.
func touch(t *testing.T, path string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(path), cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.WriteFile(path, []byte{}, cryptoutilSharedMagic.CICDOutputFilePermissions))
}

// findProjectRoot walks up from the test working directory looking for go.mod.
func findProjectRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	require.NoError(t, err)

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Skip("go.mod not found; skipping workspace test")

			return ""
		}

		dir = parent
	}
}

func TestFindViolationsInDir_NoConfigsDir_ReturnsError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	violations, err := lintFitnessConfigsEmptyDir.FindViolationsInDir(root, os.Stat, filepath.WalkDir, os.ReadDir)
	require.Error(t, err)
	require.Nil(t, violations)
}

func TestFindViolationsInDir_EmptyConfigsRoot_ReturnsRootAsViolation(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mkdir(t, filepath.Join(root, cryptoutilSharedMagic.CICDConfigsDir))

	violations, err := lintFitnessConfigsEmptyDir.FindViolationsInDir(root, os.Stat, filepath.WalkDir, os.ReadDir)
	require.NoError(t, err)
	// The configs/ dir itself is empty (0 children) → violation
	require.Len(t, violations, 1)
}

func TestFindViolationsInDir_ConfigsDirWithFile_NoViolations(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	touch(t, filepath.Join(root, cryptoutilSharedMagic.CICDConfigsDir, "some.yml"))

	violations, err := lintFitnessConfigsEmptyDir.FindViolationsInDir(root, os.Stat, filepath.WalkDir, os.ReadDir)
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestFindViolationsInDir_EmptySubdir_IsViolation(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	// configs/ has a file (not empty), but configs/empty-sub/ has no children
	touch(t, filepath.Join(root, cryptoutilSharedMagic.CICDConfigsDir, "some.yml"))
	mkdir(t, filepath.Join(root, cryptoutilSharedMagic.CICDConfigsDir, "empty-sub"))

	violations, err := lintFitnessConfigsEmptyDir.FindViolationsInDir(root, os.Stat, filepath.WalkDir, os.ReadDir)
	require.NoError(t, err)
	require.Len(t, violations, 1)
	require.Contains(t, violations[0], "empty-sub")
}

func TestFindViolationsInDir_GitkeepDir_NoViolation(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	// configs/ has a .gitkeep child → not empty
	touch(t, filepath.Join(root, cryptoutilSharedMagic.CICDConfigsDir, ".gitkeep"))

	violations, err := lintFitnessConfigsEmptyDir.FindViolationsInDir(root, os.Stat, filepath.WalkDir, os.ReadDir)
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestFindViolationsInDir_SubdirWithGitkeep_NoViolation(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	// configs/ has a subdir with .gitkeep → subdir not empty
	touch(t, filepath.Join(root, cryptoutilSharedMagic.CICDConfigsDir, "some.yml"))
	touch(t, filepath.Join(root, cryptoutilSharedMagic.CICDConfigsDir, "subdir", ".gitkeep"))

	violations, err := lintFitnessConfigsEmptyDir.FindViolationsInDir(root, os.Stat, filepath.WalkDir, os.ReadDir)
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestFindViolationsInDir_MultipleEmptySubdirs_AllViolations(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	configsDir := filepath.Join(root, cryptoutilSharedMagic.CICDConfigsDir)
	// configs/ root has a file (not a violation)
	touch(t, filepath.Join(configsDir, "root.yml"))
	// Two empty subdirs (violations)
	mkdir(t, filepath.Join(configsDir, "empty-a"))
	mkdir(t, filepath.Join(configsDir, "empty-b"))

	violations, err := lintFitnessConfigsEmptyDir.FindViolationsInDir(root, os.Stat, filepath.WalkDir, os.ReadDir)
	require.NoError(t, err)
	require.Len(t, violations, 2)
}

func TestFindViolationsInDir_DirWithOnlySubdirs_IsNotViolation(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	configsDir := filepath.Join(root, cryptoutilSharedMagic.CICDConfigsDir)
	// configs/ itself only has subdirs, but one of those subdirs has a file
	touch(t, filepath.Join(configsDir, "ps-id", "config.yml"))

	violations, err := lintFitnessConfigsEmptyDir.FindViolationsInDir(root, os.Stat, filepath.WalkDir, os.ReadDir)
	require.NoError(t, err)
	// configs/ has child "ps-id/" → not empty
	// configs/ps-id/ has child "config.yml" → not empty
	require.Empty(t, violations)
}

func TestCheckInDir_ValidStructure(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	touch(t, filepath.Join(root, cryptoutilSharedMagic.CICDConfigsDir, "ps-id", "ps-id-main.yml"))

	logger := newTestLogger()
	err := lintFitnessConfigsEmptyDir.CheckInDir(logger, root, os.Stat, filepath.WalkDir, os.ReadDir)
	require.NoError(t, err)
}

func TestCheckInDir_EmptyDir_ReturnsError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	touch(t, filepath.Join(root, cryptoutilSharedMagic.CICDConfigsDir, "root.yml"))
	mkdir(t, filepath.Join(root, cryptoutilSharedMagic.CICDConfigsDir, "empty-dir"))

	logger := newTestLogger()
	err := lintFitnessConfigsEmptyDir.CheckInDir(logger, root, os.Stat, filepath.WalkDir, os.ReadDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty directories in configs/")
}

func TestCheckInDir_NoConfigsDir_ReturnsError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	logger := newTestLogger()
	err := lintFitnessConfigsEmptyDir.CheckInDir(logger, root, os.Stat, filepath.WalkDir, os.ReadDir)
	require.Error(t, err)
}

func TestFindViolationsInDir_StatError(t *testing.T) {
	t.Parallel()

	violations, err := lintFitnessConfigsEmptyDir.FindViolationsInDir(
		".",
		func(_ string) (os.FileInfo, error) { return nil, errors.New("injected stat error") },
		filepath.WalkDir,
		os.ReadDir,
	)
	require.Error(t, err)
	require.Nil(t, violations)
	require.Contains(t, err.Error(), "injected stat error")
}

func TestFindViolationsInDir_WalkError(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	violations, err := lintFitnessConfigsEmptyDir.FindViolationsInDir(
		tmp,
		func(_ string) (os.FileInfo, error) { return os.Stat(tmp) },
		func(_ string, _ fs.WalkDirFunc) error { return errors.New("injected walk error") },
		os.ReadDir,
	)
	require.Error(t, err)
	require.Nil(t, violations)
	require.Contains(t, err.Error(), "injected walk error")
}

func TestFindViolationsInDir_WalkCallbackError(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	violations, err := lintFitnessConfigsEmptyDir.FindViolationsInDir(
		tmp,
		func(_ string) (os.FileInfo, error) { return os.Stat(tmp) },
		func(root string, fn fs.WalkDirFunc) error {
			return fn(root, nil, errors.New("injected walkdir callback error"))
		},
		os.ReadDir,
	)
	require.Error(t, err)
	require.Nil(t, violations)
	require.Contains(t, err.Error(), "injected walkdir callback error")
}

func TestFindViolationsInDir_ReadDirError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, cryptoutilSharedMagic.CICDConfigsDir), cryptoutilSharedMagic.CICDTempDirPermissions))

	violations, err := lintFitnessConfigsEmptyDir.FindViolationsInDir(
		root,
		os.Stat,
		filepath.WalkDir,
		func(_ string) ([]os.DirEntry, error) { return nil, errors.New("injected readdir error") },
	)
	require.Error(t, err)
	require.Nil(t, violations)
	require.Contains(t, err.Error(), "injected readdir error")
}

func TestCheck_Integration(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)

	logger := newTestLogger()
	origDir, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(root))

	defer func() { require.NoError(t, os.Chdir(origDir)) }()

	err = lintFitnessConfigsEmptyDir.Check(logger)
	require.NoError(t, err)
}
