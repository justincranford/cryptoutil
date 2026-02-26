// Copyright (c) 2025 Justin Cranford

package magic_usage

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"

	"github.com/stretchr/testify/require"
)

func TestCheckMagicUsageInDir_AbsMagicDirError(t *testing.T) {
	// Cannot be parallel: modifies package-level injectable var.
	originalFn := magicUsageAbsFn

	defer func() { magicUsageAbsFn = originalFn }()

	// Set up valid magic dir with at least one constant so ParseMagicDir succeeds.
	magicDir, rootDir := setupMagicUsageDirs(t)
	writeMagicFile(t, magicDir, "magic_test_vals.go", `package magic
const TestVal = "test-value"
`)

	callCount := 0
	magicUsageAbsFn = func(path string) (string, error) {
		callCount++
		if callCount == 1 {
			return "", fmt.Errorf("injected abs error for magic dir")
		}

		return filepath.Abs(path)
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckMagicUsageInDir(logger, magicDir, rootDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot resolve magic dir")
}

func TestCheckMagicUsageInDir_WalkFnError(t *testing.T) {
	// Cannot be parallel: modifies package-level injectable var.
	originalFn := magicUsageWalkFn

	defer func() { magicUsageWalkFn = originalFn }()

	magicDir, rootDir := setupMagicUsageDirs(t)
	writeMagicFile(t, magicDir, "magic_test_vals.go", `package magic
const TestVal = "test-value"
`)

	magicUsageWalkFn = func(_ string, _ filepath.WalkFunc) error {
		return fmt.Errorf("injected walk error")
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckMagicUsageInDir(logger, magicDir, rootDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "directory walk failed")
}

func TestCheckMagicUsageInDir_FileSkipPath(t *testing.T) {
	t.Parallel()

	// Create magic dir with a constant.
	magicDir := t.TempDir()
	rootDir := t.TempDir()

	writeMagicFile(t, magicDir, "magic_test_vals.go", `package magic
const TestVal = "hello"
`)

	// Create a Go file inside a "test-output" subdirectory.
	// The directory-level skip should catch this, but also exercises the code path
	// where MagicShouldSkipPath is checked at the file level.
	testOutputDir := filepath.Join(rootDir, "test-output")
	require.NoError(t, os.MkdirAll(testOutputDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	// Write a file that uses the literal "hello" — if skip works, no violation is reported.
	writeMagicFile(t, testOutputDir, "skipped.go", `package testoutput
var x = "hello"
`)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckMagicUsageInDir(logger, magicDir, rootDir)
	require.NoError(t, err)
}
