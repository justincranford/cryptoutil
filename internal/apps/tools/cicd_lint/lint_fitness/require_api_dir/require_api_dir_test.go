// Copyright (c) 2025 Justin Cranford

package require_api_dir

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// findProjectRoot finds the project root by looking for go.mod.
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

// TestCheck_DirectCall verifies that Check() delegates to CheckInDir correctly from the real workspace.
func TestCheck_DirectCall(t *testing.T) {
	t.Parallel()

	_, err := findProjectRoot()
	if err != nil {
		t.Skip("Skipping integration test - cannot find project root (no go.mod)")
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	// Check() calls CheckInDir(logger, ".") which uses the current working directory.
	// Tests run from the package directory, so "." resolves to the project subdir,
	// not the project root. This verifies the delegation works without errors from the project.
	err = CheckInDir(logger, ".")
	// The "." dir is internal/apps/cicd/lint_fitness/require_api_dir, which has no api/ subdir.
	// So CheckInDir will return an error - we just verify Check() delegates correctly.
	_ = err
	// Now call Check() and verify it produces the same result.
	err2 := Check(logger)
	require.Equal(t, err == nil, err2 == nil)
}

func TestCheckInDir_AllValid(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create all registered services with required files.
	for _, svc := range knownAPIServices {
		apiName := svc.Product + "-" + svc.Service
		svcAPIDir := filepath.Join(tmpDir, "api", apiName)
		require.NoError(t, os.MkdirAll(svcAPIDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

		for _, f := range requiredFiles {
			require.NoError(t, os.WriteFile(filepath.Join(svcAPIDir, f), []byte("package x\n"), cryptoutilSharedMagic.CacheFilePermissions))
		}
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_MissingAPIDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create api/ root dir but no service subdirs.
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "api"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "directory missing")
}

func TestCheckInDir_MissingGenerateGo(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create all services with all required files, then omit generate.go for first service.
	firstSvc := knownAPIServices[0]
	firstAPIName := firstSvc.Product + "-" + firstSvc.Service

	for _, svc := range knownAPIServices {
		apiName := svc.Product + "-" + svc.Service
		svcAPIDir := filepath.Join(tmpDir, "api", apiName)
		require.NoError(t, os.MkdirAll(svcAPIDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

		// Skip generate.go for the first service to isolate the violation.
		if apiName != firstAPIName {
			for _, f := range requiredFiles {
				require.NoError(t, os.WriteFile(filepath.Join(svcAPIDir, f), []byte("package x\n"), cryptoutilSharedMagic.CacheFilePermissions))
			}
		}
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("api/%s: missing required file generate.go", firstAPIName))
}

func TestCheckInDir_MissingAPIRootDir(t *testing.T) {
	t.Parallel()

	// tmpDir has no api/ subdirectory at all.
	tmpDir := t.TempDir()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "api/ directory not found")
}
