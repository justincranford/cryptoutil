// Copyright (c) 2025 Justin Cranford

package lint_text

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

func TestLint_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Lint(logger, map[string][]string{})

	require.NoError(t, err, "Lint should succeed with no files")
}

func TestLint_ValidUTF8Files(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create valid UTF-8 files.
	validFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(validFile, []byte("Hello, World!"), 0o600)
	require.NoError(t, err)

	validGoFile := filepath.Join(tmpDir, "main.go")
	err = os.WriteFile(validGoFile, []byte("package main\n\nfunc main() {}\n"), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"txt": {validFile},
		"go":  {validGoFile},
	}

	err = Lint(logger, filesByExtension)
	require.NoError(t, err, "Lint should succeed with valid UTF-8 files")
}

func TestLint_FileWithBOM(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create file with UTF-8 BOM.
	bomFile := filepath.Join(tmpDir, "bom.txt")

	bomContent := append([]byte{0xEF, 0xBB, 0xBF}, []byte("Hello with BOM")...)
	err := os.WriteFile(bomFile, bomContent, 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"txt": {bomFile},
	}

	err = Lint(logger, filesByExtension)
	require.Error(t, err, "Lint should fail with BOM file")
	require.Contains(t, err.Error(), "lint-text failed", "Error should indicate lint-text failure")
}

func TestFilterTextFilesViaLint(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(goFile, []byte("package main"), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"go": {goFile},
	}

	err = Lint(logger, filesByExtension)
	require.NoError(t, err, "Lint should succeed with valid go files")
}
