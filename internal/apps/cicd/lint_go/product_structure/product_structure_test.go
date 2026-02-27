// Copyright (c) 2025 Justin Cranford

package product_structure

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

	for _, product := range knownProducts {
		productDir := filepath.Join(tmpDir, "internal", "apps", product)
		require.NoError(t, os.MkdirAll(productDir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(productDir, product+".go"), []byte("package "+product), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(productDir, product+"_test.go"), []byte("package "+product), cryptoutilSharedMagic.CacheFilePermissions))
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_MissingEntryFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	for _, product := range knownProducts {
		productDir := filepath.Join(tmpDir, "internal", "apps", product)
		require.NoError(t, os.MkdirAll(productDir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(productDir, product+"_test.go"), []byte("package "+product), cryptoutilSharedMagic.CacheFilePermissions))
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing entry file")
}

func TestCheckInDir_MissingTestFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	for _, product := range knownProducts {
		productDir := filepath.Join(tmpDir, "internal", "apps", product)
		require.NoError(t, os.MkdirAll(productDir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(productDir, product+".go"), []byte("package "+product), cryptoutilSharedMagic.CacheFilePermissions))
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing test file")
}

func TestCheckInDir_MissingProductDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "internal", "apps"), 0o755))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "product directory missing")
}

func TestCheckInDir_NoAppsDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "internal/apps directory not found")
}

// Sequential: uses os.Chdir (global process state).
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
