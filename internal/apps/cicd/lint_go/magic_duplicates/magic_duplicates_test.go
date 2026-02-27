// Copyright (c) 2025 Justin Cranford

package magic_duplicates

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"

	"github.com/stretchr/testify/require"
)

// writeMagicFile creates a file inside dir with the given content.
func writeMagicFile(t *testing.T, dir, name, content string) {
	t.Helper()

	err := os.WriteFile(filepath.Join(dir, name), []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)
}

func TestCheckMagicDuplicatesInDir_NoDuplicates(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeMagicFile(t, dir, "magic_strings.go", `package magic

const (
ProtocolHTTPS = "https"
SchemeHTTP    = "http"
)
`)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-dup-test")
	err := CheckMagicDuplicatesInDir(logger, dir)
	require.NoError(t, err)
}

func TestCheckMagicDuplicatesInDir_WithDuplicates(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeMagicFile(t, dir, "magic_net.go", `package magic

const (
ProtocolHTTPS = "https"
SchemeHTTPS   = "https"
)
`)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-dup-test")
	err := CheckMagicDuplicatesInDir(logger, dir)
	// magic-duplicates is informational: violations are logged but do not return an error.
	require.NoError(t, err)
}

func TestCheckMagicDuplicatesInDir_TrivialInts_NotDuplicate(t *testing.T) {
	t.Parallel()

	// Trivial integers (0,1,2,3,4,-1) are still flagged as duplicates if they
	// share the same value — isMagicTrivialLiteral only suppresses usage scanning,
	// not the duplicate-definition check.
	dir := t.TempDir()
	writeMagicFile(t, dir, "magic_sizes.go", `package magic

const (
DefaultSizeA = 5
DefaultSizeB = 5
)
`)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-dup-test")
	err := CheckMagicDuplicatesInDir(logger, dir)
	// magic-duplicates is informational: violations are logged but do not return an error.
	require.NoError(t, err)
}

func TestCheckMagicDuplicatesInDir_InvalidDir(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-dup-test")
	err := CheckMagicDuplicatesInDir(logger, "/nonexistent/path/that/does/not/exist")
	require.Error(t, err)
}

func TestCheckMagicDuplicatesInDir_SingleConstant(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeMagicFile(t, dir, "magic_only.go", `package magic

const Alone = "solo"
`)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-dup-test")
	err := CheckMagicDuplicatesInDir(logger, dir)
	require.NoError(t, err)
}

func TestCheckMagicDuplicatesInDir_EmptyPackage(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeMagicFile(t, dir, "magic.go", `package magic
`)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-dup-test")
	err := CheckMagicDuplicatesInDir(logger, dir)
	require.NoError(t, err)
}

func TestCheckMagicDuplicatesInDir_MultiFile_Duplicates(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeMagicFile(t, dir, "magic_a.go", `package magic

const AlgoRSA = "RSA"
`)
	writeMagicFile(t, dir, "magic_b.go", `package magic

const AlgorithmRSA = "RSA"
`)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-dup-test")
	err := CheckMagicDuplicatesInDir(logger, dir)
	// magic-duplicates is informational: violations are logged but do not return an error.
	require.NoError(t, err)
}

func TestCheck_UsesMagicDefaultDir(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	// Check() calls CheckMagicDuplicatesInDir with MagicDefaultDir="internal/shared/magic".
	// When run from the package test directory, that relative path does not exist,
	// so Check() returns an error. This exercises the Check() code path.
	err := Check(logger)
	require.Error(t, err, "Check() should fail when MagicDefaultDir does not exist relative to CWD")
	require.Contains(t, err.Error(), "failed to parse magic package")
}

func TestCheckMagicDuplicatesInDir_MultipleDuplicateGroups(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeMagicFile(t, dir, "magic_multi.go", `package magic

const (
	ProtoHTTPS  = "https"
	SchemeHTTPS = "https"
	SizeA       = 42
	SizeB       = 42
	AlgoRSA     = "RSA"
	AlgoRSA2    = "RSA"
)
`)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-dup-test")
	err := CheckMagicDuplicatesInDir(logger, dir)
	// magic-duplicates is informational: violations are logged but do not return an error.
	require.NoError(t, err)
}

// writeGoFile creates a .go file inside a subdirectory of dir.
func writeGoFile(t *testing.T, dir, subPkg, name, content string) {
	t.Helper()

	pkgDir := filepath.Join(dir, subPkg)
	require.NoError(t, os.MkdirAll(pkgDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))

	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, name), []byte(content), cryptoutilSharedMagic.CacheFilePermissions))
}

func TestCheckCrossFileDuplicatesInDir_NoDuplicates(t *testing.T) {
	t.Parallel()

	magicDir := t.TempDir()
	rootDir := t.TempDir()

	writeGoFile(t, rootDir, "pkg/a", "a.go", `package a

const AlgoRSA = "RSA"
`)
	writeGoFile(t, rootDir, "pkg/b", "b.go", `package b

const ProtoHTTPS = "https"
`)

	logger := cryptoutilCmdCicdCommon.NewLogger("cross-dup-test")
	err := CheckCrossFileDuplicatesInDir(logger, magicDir, rootDir)
	require.NoError(t, err)
}

func TestCheckCrossFileDuplicatesInDir_FindsDuplicates(t *testing.T) {
	t.Parallel()

	magicDir := t.TempDir()
	rootDir := t.TempDir()

	// Same value "https" declared in two different packages.
	writeGoFile(t, rootDir, "pkg/a", "a.go", `package a

const ProtoHTTPS = "https"
`)
	writeGoFile(t, rootDir, "pkg/b", "b.go", `package b

const SchemeHTTPS = "https"
`)

	logger := cryptoutilCmdCicdCommon.NewLogger("cross-dup-test")
	// magic-cross-duplicates is informational: violations are logged but do not return an error.
	err := CheckCrossFileDuplicatesInDir(logger, magicDir, rootDir)
	require.NoError(t, err)
}

func TestCheckCrossFileDuplicatesInDir_SameFileTwice_NotCrossDuplicate(t *testing.T) {
	t.Parallel()

	magicDir := t.TempDir()
	rootDir := t.TempDir()

	// Same value in one file does not count as a cross-file duplicate.
	writeGoFile(t, rootDir, "pkg/a", "a.go", `package a

const (
ProtoHTTPS  = "https"
SchemeHTTPS = "https"
)
`)

	logger := cryptoutilCmdCicdCommon.NewLogger("cross-dup-test")
	err := CheckCrossFileDuplicatesInDir(logger, magicDir, rootDir)
	require.NoError(t, err)
}

func TestCheckCrossFileDuplicatesInDir_SkipsMagicDir(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	magicDir := filepath.Join(rootDir, "shared", "magic")

	require.NoError(t, os.MkdirAll(magicDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))

	// Constant in magic dir should not be reported.
	require.NoError(t, os.WriteFile(filepath.Join(magicDir, "magic_net.go"), []byte(`package magic

const ProtoHTTPS = "https"
`), cryptoutilSharedMagic.CacheFilePermissions))

	// Same value in a non-magic file.
	writeGoFile(t, rootDir, "pkg/a", "a.go", `package a

const SchemeHTTPS = "https"
`)

	logger := cryptoutilCmdCicdCommon.NewLogger("cross-dup-test")
	// Only one non-magic file has the value, so no cross-file duplicate.
	err := CheckCrossFileDuplicatesInDir(logger, magicDir, rootDir)
	require.NoError(t, err)
}

func TestCheckCrossFileDuplicatesInDir_AbsMagicDirError(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test replaces package-level injectable function.
	// Force magicDuplicatesAbsFn to error on the magicDir call.
	callCount := 0
	origFn := magicDuplicatesAbsFn

	t.Cleanup(func() { magicDuplicatesAbsFn = origFn })

	magicDuplicatesAbsFn = func(path string) (string, error) {
		callCount++
		if callCount == 1 {
			return "", os.ErrInvalid
		}

		return origFn(path)
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("cross-dup-test")
	err := CheckCrossFileDuplicatesInDir(logger, "/some/magic", "/some/root")
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot resolve magic dir")
}

func TestCheckCrossFileDuplicatesInDir_AbsRootDirError(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test replaces package-level injectable function.
	callCount := 0
	origFn := magicDuplicatesAbsFn

	t.Cleanup(func() { magicDuplicatesAbsFn = origFn })

	magicDuplicatesAbsFn = func(path string) (string, error) {
		callCount++
		if callCount == 2 {
			return "", os.ErrInvalid
		}

		return origFn(path)
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("cross-dup-test")
	err := CheckCrossFileDuplicatesInDir(logger, "/some/magic", "/some/root")
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot resolve root dir")
}

func TestCheckCrossFileDuplicatesInDir_WalkError(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test replaces package-level injectable function.
	origFn := magicDuplicatesWalkFn

	t.Cleanup(func() { magicDuplicatesWalkFn = origFn })

	magicDuplicatesWalkFn = func(root string, fn filepath.WalkFunc) error {
		return os.ErrPermission
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("cross-dup-test")
	err := CheckCrossFileDuplicatesInDir(logger, t.TempDir(), t.TempDir())
	require.Error(t, err)
	require.Contains(t, err.Error(), "directory walk failed")
}
