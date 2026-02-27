// Copyright (c) 2025 Justin Cranford

package product_wiring

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
	cmdDir := filepath.Join(tmpDir, "cmd")

	// Create product entry points.
	for _, product := range knownProducts {
		dir := filepath.Join(cmdDir, product)
		require.NoError(t, os.MkdirAll(dir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), cryptoutilSharedMagic.CacheFilePermissions))
	}

	// Create service entry points.
	for _, pair := range knownServices {
		dir := filepath.Join(cmdDir, pair.product+"-"+pair.service)
		require.NoError(t, os.MkdirAll(dir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), cryptoutilSharedMagic.CacheFilePermissions))
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_MissingProductEntry(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cmdDir := filepath.Join(tmpDir, "cmd")

	// Create all service entries but skip one product entry.
	for i, product := range knownProducts {
		if i == 0 {
			continue // Skip first product.
		}

		dir := filepath.Join(cmdDir, product)
		require.NoError(t, os.MkdirAll(dir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), cryptoutilSharedMagic.CacheFilePermissions))
	}

	for _, pair := range knownServices {
		dir := filepath.Join(cmdDir, pair.product+"-"+pair.service)
		require.NoError(t, os.MkdirAll(dir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), cryptoutilSharedMagic.CacheFilePermissions))
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing product entry point")
}

func TestCheckInDir_MissingServiceEntry(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cmdDir := filepath.Join(tmpDir, "cmd")

	for _, product := range knownProducts {
		dir := filepath.Join(cmdDir, product)
		require.NoError(t, os.MkdirAll(dir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), cryptoutilSharedMagic.CacheFilePermissions))
	}

	// Create all service entries but skip one.
	for i, pair := range knownServices {
		if i == 0 {
			continue // Skip first service.
		}

		dir := filepath.Join(cmdDir, pair.product+"-"+pair.service)
		require.NoError(t, os.MkdirAll(dir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), cryptoutilSharedMagic.CacheFilePermissions))
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing service entry point")
}

func TestCheckInDir_NoCmdDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cmd directory not found")
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
