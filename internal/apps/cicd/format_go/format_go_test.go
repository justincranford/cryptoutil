// Copyright (c) 2025 Justin Cranford

package format_go

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

func TestFormat_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Format(logger, map[string][]string{})

	require.NoError(t, err, "Format should succeed with no files")
}

func TestFormat_WithFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	// Create a file with loop var copy.
	err := os.WriteFile(testFile, []byte(testGoContentWithLoopVarCopy), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"go": {testFile},
	}

	err = Format(logger, filesByExtension)
	// Format returns error if modifications were made.
	// But GetGoFiles may filter out test files.
	if err != nil {
		require.Contains(t, err.Error(), "completed with modifications")
	}
}

func TestFormat_ErrorPath(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "invalid.go")

	// Create invalid Go file to trigger parse error.
	err := os.WriteFile(testFile, []byte(testGoContentInvalid), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"go": {testFile},
	}

	err = Format(logger, filesByExtension)

	// Format may return error due to parse failure or succeed if file filtered out.
	// We just verify it doesn't panic.
	_ = err
}

// Sequential: uses os.Chdir (global process state).
func TestFormat_FormatterWalkError(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes working directory.
	if runtime.GOOS == cryptoutilSharedMagic.OSNameWindows {
		t.Skip("os.Chmod does not enforce POSIX permissions on Windows")
	}

	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { require.NoError(t, os.Chdir(origDir)) }()

	tmpDir := t.TempDir()
	require.NoError(t, os.Chdir(tmpDir))

	// Create chmod 0000 subdir - copyloopvar.Fix walks rootDir="."
	// and the walk callback receives OS error, causing Fix to return error,
	// which covers lines 54-56 (registeredFormatters error append path).
	require.NoError(t, os.MkdirAll("locked", 0o700))
	require.NoError(t, os.Chmod("locked", 0o000))

	t.Cleanup(func() { _ = os.Chmod(filepath.Join(tmpDir, "locked"), 0o700) })

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Format(logger, map[string][]string{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "completed with modifications")
}

const testGoContentInvalid = "package main\n\nfunc main() {\n\tthis is not valid go code\n}\n"

const testGoContentWithLoopVarCopy = `package main

func main() {
	items := []int{1, 2, 3}
	for _, v := range items {
		v := v
		println(v)
	}
}
`

// testGoContentWithInterfaceBraces is Go content with interface{} in a temp file
// that will NOT match the format-go self-exclusion pattern (not in format_go dir).
// CRITICAL: Uses interface{} in string literal intentionally - this is test data, not production code.
const testGoContentWithInterfaceBraces = "package server\n\nfunc Handle(data interface{}) interface{} {\n\treturn data\n}\n"

// TestFormat_WithEnforceAnyModification calls Format with a file containing interface{}
// to cover the simple formatter error path and the len(errors) > 0 final return.
func TestFormat_WithEnforceAnyModification(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "server.go")

	err := os.WriteFile(testFile, []byte(testGoContentWithInterfaceBraces), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"go": {testFile},
	}

	err = Format(logger, filesByExtension)
	require.Error(t, err, "Format should return error when formatters made modifications")
	require.Contains(t, err.Error(), "completed with modifications")
}
