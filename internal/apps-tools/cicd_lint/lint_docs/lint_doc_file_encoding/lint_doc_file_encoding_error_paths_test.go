// Copyright (c) 2025-2026 Justin Cranford.
package lint_doc_file_encoding

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestCheck(t *testing.T) {
	t.Parallel()

	err := Check(cryptoutilCmdCicdCommon.NewLogger("lint-doc-file-encoding-test"))
	require.NoError(t, err)

	err = checkWithRootFinder(
		cryptoutilCmdCicdCommon.NewLogger("lint-doc-file-encoding-test"),
		func() (string, error) { return "", errors.New("injected root failure") },
	)
	require.ErrorContains(t, err, "find project root")
	require.ErrorContains(t, err, "injected root failure")
}

func TestCheckInDirWithOperations_CollectAndReadFailures(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("lint-doc-file-encoding-test")
	rootDir := t.TempDir()
	operations := productionFileOperations()
	operations.readFile = func(path string) ([]byte, error) {
		if path == filepath.Join(rootDir, cryptoutilSharedMagic.CICDDocsHandbookPath) {
			return nil, errors.New("injected read failure")
		}

		return []byte("valid\n"), nil
	}

	err := checkInDirWithOperations(logger, rootDir, operations)
	require.ErrorContains(t, err, "injected read failure")

	operations = productionFileOperations()
	operations.stat = func(string) (os.FileInfo, error) {
		return nil, errors.New("injected stat failure")
	}

	err = checkInDirWithOperations(logger, rootDir, operations)
	require.ErrorContains(t, err, "collect policy files")
	require.ErrorContains(t, err, "injected stat failure")
}

func TestCollectPolicyFilesWithOperations_MissingAndWalkFailures(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	filesFromDefaultOperations, err := collectPolicyFiles(rootDir)
	require.NoError(t, err)
	require.Len(t, filesFromDefaultOperations, 5)

	files, err := collectPolicyFilesWithOperations(rootDir, productionFileOperations())
	require.NoError(t, err)
	require.Len(t, files, 5)

	injectedWalkErr := errors.New("injected walk failure")
	operations := productionFileOperations()
	operations.walkDir = func(_ string, walkFn fs.WalkDirFunc) error {
		return walkFn("ignored", nil, injectedWalkErr)
	}

	require.NoError(t, os.MkdirAll(
		filepath.Join(rootDir, policyDirectories[0]),
		cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute,
	))
	_, err = collectPolicyFilesWithOperations(rootDir, operations)
	require.ErrorContains(t, err, "walk")
	require.ErrorContains(t, err, "injected walk failure")

	operations.walkDir = func(string, fs.WalkDirFunc) error {
		return injectedWalkErr
	}
	_, err = collectPolicyFilesWithOperations(rootDir, operations)
	require.ErrorContains(t, err, "injected walk failure")
}

func TestCheckFileWithReadFile_ErrorAndInvalidEncoding(t *testing.T) {
	t.Parallel()

	_, err := checkFileWithReadFile("missing.md", func(string) ([]byte, error) {
		return nil, errors.New("injected read failure")
	})
	require.ErrorContains(t, err, "read file")

	issues, err := checkFileWithReadFile("invalid.md", func(string) ([]byte, error) {
		return []byte{0xFF}, nil
	})
	require.NoError(t, err)
	require.Equal(t, []string{"contains invalid UTF-8 byte sequences"}, issues)

	issues, err = checkFileWithReadFile("multiple.md", func(string) ([]byte, error) {
		return []byte{0xEF, 0xBB, 0xBF, 'a', '\r', '\n', 0x00}, nil
	})
	require.NoError(t, err)
	require.Len(t, issues, 3)
	require.Contains(t, issues[0], "UTF-8 BOM")
	require.Contains(t, issues[1], "NUL byte")
	require.Contains(t, issues[2], "CRLF")
}

func TestFindProjectRootWithGetwd(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(rootDir, "go.mod"), []byte("module example\n"), cryptoutilSharedMagic.CacheFilePermissions))

	getwd := func() (string, error) { return rootDir, nil }
	foundRoot, err := findProjectRootWithGetwd(getwd)
	require.NoError(t, err)
	require.Equal(t, rootDir, foundRoot)

	getwd = func() (string, error) { return "", errors.New("injected getwd failure") }
	_, err = findProjectRootWithGetwd(getwd)
	require.ErrorContains(t, err, "failed to get working directory")

	noModuleDir := t.TempDir()
	_, err = findProjectRootWithOperations(
		func() (string, error) { return noModuleDir, nil },
		func(string) (os.FileInfo, error) { return nil, fs.ErrNotExist },
	)
	require.ErrorContains(t, err, "go.mod not found")
}

func TestCollectPolicyFilesWithOperations_StatErrorIncludesDirectory(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	injectedStatErr := fmt.Errorf("stat unavailable")
	operations := productionFileOperations()
	operations.stat = func(path string) (os.FileInfo, error) {
		if path == filepath.Join(rootDir, policyDirectories[0]) {
			return nil, injectedStatErr
		}

		return os.Stat(path)
	}

	_, err := collectPolicyFilesWithOperations(rootDir, operations)
	require.ErrorContains(t, err, "stat")
	require.ErrorIs(t, err, injectedStatErr)
}
