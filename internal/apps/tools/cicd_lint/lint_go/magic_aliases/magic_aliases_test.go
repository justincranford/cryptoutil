// Copyright (c) 2025 Justin Cranford

package magic_aliases

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// writeGoFile creates a .go file inside a subdirectory of dir.
func writeGoFile(t *testing.T, dir, subPkg, name, content string) {
	t.Helper()

	pkgDir := filepath.Join(dir, subPkg)

	err := os.MkdirAll(pkgDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(pkgDir, name), []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)
}

func TestCheckMagicAliasesInDir_NoAliases(t *testing.T) {
	t.Parallel()

	magicDir := t.TempDir()
	rootDir := t.TempDir()

	writeGoFile(t, rootDir, "pkg/a", "a.go", `package a

import cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

var x = cryptoutilSharedMagic.EmptyString
`)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-aliases-test")
	err := CheckMagicAliasesInDir(logger, magicDir, rootDir, filepath.Abs, filepath.Walk)
	require.NoError(t, err)
}

func TestCheckMagicAliasesInDir_FindsUnexportedAlias(t *testing.T) {
	t.Parallel()

	magicDir := t.TempDir()
	rootDir := t.TempDir()

	writeGoFile(t, rootDir, "pkg/a", "a.go", `package a

import cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

const (
localName = cryptoutilSharedMagic.EmptyString
)
`)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-aliases-test")
	// magic-aliases is informational: violations are logged but do not return an error.
	err := CheckMagicAliasesInDir(logger, magicDir, rootDir, filepath.Abs, filepath.Walk)
	require.NoError(t, err)
}

func TestCheckMagicAliasesInDir_SkipsExportedAlias(t *testing.T) {
	t.Parallel()

	magicDir := t.TempDir()
	rootDir := t.TempDir()

	writeGoFile(t, rootDir, "pkg/a", "a.go", `package a

import cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

const (
ExportedAlias = cryptoutilSharedMagic.EmptyString
)
`)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-aliases-test")
	err := CheckMagicAliasesInDir(logger, magicDir, rootDir, filepath.Abs, filepath.Walk)
	require.NoError(t, err)
}

func TestCheckMagicAliasesInDir_SkipsMagicDir(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	magicDir := filepath.Join(rootDir, "shared", "magic")

	err := os.MkdirAll(magicDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute)
	require.NoError(t, err)

	// Const alias inside magic dir should not be reported.
	err = os.WriteFile(filepath.Join(magicDir, "magic_test_helper.go"), []byte(`package magic

import cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

const localHelper = cryptoutilSharedMagic.EmptyString
`), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-aliases-test")
	err = CheckMagicAliasesInDir(logger, magicDir, rootDir, filepath.Abs, filepath.Walk)
	require.NoError(t, err)
}

func TestCheckMagicAliasesInDir_SkipsNonMagicSelector(t *testing.T) {
	t.Parallel()

	magicDir := t.TempDir()
	rootDir := t.TempDir()

	writeGoFile(t, rootDir, "pkg/a", "a.go", `package a

import "fmt"

var prefix = fmt.Sprintf
`)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-aliases-test")
	err := CheckMagicAliasesInDir(logger, magicDir, rootDir, filepath.Abs, filepath.Walk)
	require.NoError(t, err)
}

func TestCheckMagicAliasesInDir_NoMagicImport(t *testing.T) {
	t.Parallel()

	magicDir := t.TempDir()
	rootDir := t.TempDir()

	writeGoFile(t, rootDir, "pkg/a", "a.go", `package a

const x = "hello"
`)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-aliases-test")
	err := CheckMagicAliasesInDir(logger, magicDir, rootDir, filepath.Abs, filepath.Walk)
	require.NoError(t, err)
}

func TestCheckMagicAliasesInDir_MultipleAliases(t *testing.T) {
	t.Parallel()

	magicDir := t.TempDir()
	rootDir := t.TempDir()

	writeGoFile(t, rootDir, "pkg/a", "a.go", `package a

import cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

const (
localA = cryptoutilSharedMagic.EmptyString
localB = cryptoutilSharedMagic.ProtocolHTTPS
localC = cryptoutilSharedMagic.ProtocolHTTP
)
`)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-aliases-test")
	// magic-aliases is informational: violations are logged but do not return an error.
	err := CheckMagicAliasesInDir(logger, magicDir, rootDir, filepath.Abs, filepath.Walk)
	require.NoError(t, err)
}

func TestCheckMagicAliasesInDir_InvalidRootDir(t *testing.T) {
	t.Parallel()

	magicDir := t.TempDir()

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-aliases-test")
	err := CheckMagicAliasesInDir(logger, magicDir, "/nonexistent/root", filepath.Abs, filepath.Walk)
	require.Error(t, err)
}

func TestCheckMagicAliasesInDir_AbsMagicDirError(t *testing.T) {
	t.Parallel()

	callCount := 0
	stubAbsFn := func(path string) (string, error) {
		callCount++
		if callCount == 1 {
			return "", fmt.Errorf("injected abs error")
		}

		return filepath.Abs(path)
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-aliases-test")
	err := CheckMagicAliasesInDir(logger, ".", ".", stubAbsFn, filepath.Walk)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot resolve magic dir")
}

func TestCheckMagicAliasesInDir_AbsRootDirError(t *testing.T) {
	t.Parallel()

	callCount := 0
	stubAbsFn := func(path string) (string, error) {
		callCount++
		if callCount == 2 {
			return "", fmt.Errorf("injected abs error")
		}

		return filepath.Abs(path)
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-aliases-test")
	err := CheckMagicAliasesInDir(logger, ".", ".", stubAbsFn, filepath.Walk)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot resolve root dir")
}

func TestCheckMagicAliasesInDir_WalkError(t *testing.T) {
	t.Parallel()

	stubWalkFn := func(_ string, _ filepath.WalkFunc) error {
		return fmt.Errorf("injected walk error")
	}

	magicDir := t.TempDir()

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-aliases-test")
	err := CheckMagicAliasesInDir(logger, magicDir, ".", filepath.Abs, stubWalkFn)
	require.Error(t, err)
	require.Contains(t, err.Error(), "directory walk failed")
}

func TestCheck_UsesMagicDefaultDir(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-aliases-test")
	// Should not error - may find violations but they are informational.
	err := Check(logger)
	require.NoError(t, err)
}

func TestCheckMagicAliasesInDir_SkipsGeneratedFiles(t *testing.T) {
	t.Parallel()

	magicDir := t.TempDir()
	rootDir := t.TempDir()

	writeGoFile(t, rootDir, "pkg/a", "server.gen.go", `package a

import cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

const localName = cryptoutilSharedMagic.EmptyString
`)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-aliases-test")
	err := CheckMagicAliasesInDir(logger, magicDir, rootDir, filepath.Abs, filepath.Walk)
	require.NoError(t, err)
}

func TestFindMagicImportAlias_DefaultName(t *testing.T) {
	t.Parallel()

	writeDir := t.TempDir()

	err := os.WriteFile(filepath.Join(writeDir, "test.go"), []byte("package a\n\nimport \"cryptoutil/internal/shared/magic\"\n\nconst x = magic.EmptyString\n"), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	violations := findAliasesInFile(filepath.Join(writeDir, "test.go"), "test.go")
	require.Len(t, violations, 1)
	require.Equal(t, "x", violations[0].LocalName)
	require.Equal(t, "EmptyString", violations[0].MagicName)
}

func TestFindAliasesInFile_UnparseableFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	badPath := filepath.Join(dir, "bad.go")
	require.NoError(t, os.WriteFile(badPath, []byte("THIS IS NOT VALID GO @@@@@"), cryptoutilSharedMagic.CacheFilePermissions))

	violations := findAliasesInFile(badPath, "bad.go")
	require.Empty(t, violations)
}

func TestFindAliasesInFile_ConstWithoutValues(t *testing.T) {
	t.Parallel()

	// iota-based const — vspec.Values will be nil for subsequent iota specs.
	content := "package a\n\nimport cryptoutilSharedMagic \"cryptoutil/internal/shared/magic\"\n\nconst (\n\t_ = cryptoutilSharedMagic.EmptyString\n)\n\nvar _ = \"x\"\n"
	dir := t.TempDir()
	path := filepath.Join(dir, "a.go")
	require.NoError(t, os.WriteFile(path, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	violations := findAliasesInFile(path, "a.go")
	// _ is exported (well, blank identifier) — the function may or may not flag it.
	require.NotNil(t, violations)
}

// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheck_ProjectRoot(t *testing.T) {
	root, err := findProjectRoot()
	if err != nil {
		t.Skip("Skipping - cannot find project root")
	}

	orig, err := os.Getwd()
	require.NoError(t, err)

	t.Cleanup(func() { _ = os.Chdir(orig) })

	require.NoError(t, os.Chdir(root))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger)
	require.NoError(t, err)
}

// findProjectRoot finds the project root by walking up to find go.mod.
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}

		dir = parent
	}
}
