// Copyright (c) 2025 Justin Cranford

package configs_empty_dir

import (
	"errors"
	"io/fs"
	"os"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// Sequential: modifies package-level configsEmptyDirStatFn seam.
func TestFindViolationsInDir_StatError(t *testing.T) {
	orig := configsEmptyDirStatFn

	defer func() { configsEmptyDirStatFn = orig }()

	configsEmptyDirStatFn = func(_ string) (os.FileInfo, error) {
		return nil, errors.New("injected stat error")
	}

	violations, err := FindViolationsInDir(".")
	require.Error(t, err)
	require.Nil(t, violations)
	require.Contains(t, err.Error(), "injected stat error")
}

// Sequential: modifies package-level configsEmptyDirWalkFn seam.
func TestFindViolationsInDir_WalkError(t *testing.T) {
	origStat := configsEmptyDirStatFn
	origWalk := configsEmptyDirWalkFn

	defer func() {
		configsEmptyDirStatFn = origStat
		configsEmptyDirWalkFn = origWalk
	}()

	configsEmptyDirStatFn = func(_ string) (os.FileInfo, error) {
		return os.Stat(t.TempDir())
	}

	configsEmptyDirWalkFn = func(_ string, _ fs.WalkDirFunc) error {
		return errors.New("injected walk error")
	}

	violations, err := FindViolationsInDir(".")
	require.Error(t, err)
	require.Nil(t, violations)
	require.Contains(t, err.Error(), "injected walk error")
}

// Sequential: modifies package-level configsEmptyDirWalkFn seam — walks callback with error.
func TestFindViolationsInDir_WalkCallbackError(t *testing.T) {
	origStat := configsEmptyDirStatFn
	origWalk := configsEmptyDirWalkFn

	defer func() {
		configsEmptyDirStatFn = origStat
		configsEmptyDirWalkFn = origWalk
	}()

	configsEmptyDirStatFn = func(_ string) (os.FileInfo, error) {
		return os.Stat(t.TempDir())
	}

	injectedErr := errors.New("injected walkdir callback error")

	configsEmptyDirWalkFn = func(root string, fn fs.WalkDirFunc) error {
		// Call the callback with a walkErr to trigger the error path.
		return fn(root, nil, injectedErr)
	}

	violations, err := FindViolationsInDir(".")
	require.Error(t, err)
	require.Nil(t, violations)
	require.Contains(t, err.Error(), "injected walkdir callback error")
}

// Sequential: modifies package-level configsEmptyDirReadDirFn seam.
func TestFindViolationsInDir_ReadDirError(t *testing.T) {
	origStat := configsEmptyDirStatFn
	origRead := configsEmptyDirReadDirFn

	defer func() {
		configsEmptyDirStatFn = origStat
		configsEmptyDirReadDirFn = origRead
	}()

	// Create real configs dir so stat and walk succeed.
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(root+"/configs", cryptoutilSharedMagic.CICDTempDirPermissions))

	// Allow stat to succeed against real configs dir.
	configsEmptyDirStatFn = func(_ string) (os.FileInfo, error) {
		return os.Stat(root + "/configs")
	}

	configsEmptyDirReadDirFn = func(_ string) ([]os.DirEntry, error) {
		return nil, errors.New("injected readdir error")
	}

	violations, err := FindViolationsInDir(root)
	require.Error(t, err)
	require.Nil(t, violations)
	require.Contains(t, err.Error(), "injected readdir error")
}
