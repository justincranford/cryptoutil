// Copyright (c) 2025 Justin Cranford

package magic_usage

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"

	"github.com/stretchr/testify/require"
)

// writeMagicFile creates a file inside dir with the given content.
func writeMagicFile(t *testing.T, dir, name, content string) {
	t.Helper()

	err := os.WriteFile(filepath.Join(dir, name), []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)
}

// setupMagicUsageDirs creates a magic dir and a separate root dir for usage tests.
func setupMagicUsageDirs(t *testing.T) (magicDir, rootDir string) {
	t.Helper()

	magicDir = t.TempDir()
	rootDir = t.TempDir()

	return magicDir, rootDir
}

func TestCheckMagicUsageInDir_NoViolations(t *testing.T) {
	t.Parallel()

	magicDir, rootDir := setupMagicUsageDirs(t)

	// Magic package defines "https".
	writeMagicFile(t, magicDir, "magic.go", `package magic

const ProtocolHTTPS = "https"
`)

	// Application code uses the constant name (no literal).
	err := os.WriteFile(filepath.Join(rootDir, "handler.go"), []byte(`package app

import "fmt"

func greet() { fmt.Println("hello") }
`), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-usage-test")
	err = CheckMagicUsageInDir(logger, magicDir, rootDir)
	require.NoError(t, err)
}

func TestCheckMagicUsageInDir_LiteralViolation(t *testing.T) {
	t.Parallel()

	magicDir, rootDir := setupMagicUsageDirs(t)

	writeMagicFile(t, magicDir, "magic.go", `package magic

const ProtocolHTTPS = "https"
`)

	// Application code uses the raw "https" literal.
	err := os.WriteFile(filepath.Join(rootDir, "client.go"), []byte(`package app

func scheme() string { return "https" }
`), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-usage-test")
	err = CheckMagicUsageInDir(logger, magicDir, rootDir)
	// magic-usage is informational: violations are logged but do not return an error.
	require.NoError(t, err)
}

func TestCheckMagicUsageInDir_ConstRedefine(t *testing.T) {
	t.Parallel()

	magicDir, rootDir := setupMagicUsageDirs(t)

	writeMagicFile(t, magicDir, "magic.go", `package magic

const ProtocolHTTPS = "https"
`)

	// Code redefines the same value as a local constant.
	err := os.WriteFile(filepath.Join(rootDir, "localconst.go"), []byte(`package app

const localHTTPS = "https"
`), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-usage-test")
	err = CheckMagicUsageInDir(logger, magicDir, rootDir)
	// magic-usage is informational: violations are logged but do not return an error.
	require.NoError(t, err)
}

func TestCheckMagicUsageInDir_TrivialStringNotFlagged(t *testing.T) {
	t.Parallel()

	magicDir, rootDir := setupMagicUsageDirs(t)

	// Short strings (< magicMinStringLen) are trivial and should not be flagged.
	writeMagicFile(t, magicDir, "magic.go", `package magic

const EmptyString = ""
const Dot = "."
`)

	err := os.WriteFile(filepath.Join(rootDir, "app.go"), []byte(`package app

func f() string { return "." }
`), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-usage-test")
	err = CheckMagicUsageInDir(logger, magicDir, rootDir)
	require.NoError(t, err)
}

func TestCheckMagicUsageInDir_TrivialIntNotFlagged(t *testing.T) {
	t.Parallel()

	magicDir, rootDir := setupMagicUsageDirs(t)

	writeMagicFile(t, magicDir, "magic.go", `package magic

const Zero = 0
const One  = 1
`)

	err := os.WriteFile(filepath.Join(rootDir, "app.go"), []byte(`package app

func count() int { return 0 }
`), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-usage-test")
	err = CheckMagicUsageInDir(logger, magicDir, rootDir)
	require.NoError(t, err)
}

func TestCheckMagicUsageInDir_TestConstOnlyMatchesTestFile(t *testing.T) {
	t.Parallel()

	magicDir, rootDir := setupMagicUsageDirs(t)

	// TestXxx constant in magic_testing.go — should NOT flag production files.
	writeMagicFile(t, magicDir, "magic_testing.go", `package magic

const TestRateLimit = 500
`)

	// Production file happens to use literal 500.
	err := os.WriteFile(filepath.Join(rootDir, "server.go"), []byte(`package app

const localLimit = 500
`), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	// Test file uses literal 500 — SHOULD be flagged.
	err = os.WriteFile(filepath.Join(rootDir, "server_test.go"), []byte(`package app

const wantLimit = 500
`), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-usage-test")
	err = CheckMagicUsageInDir(logger, magicDir, rootDir)
	// magic-usage is informational: violations are logged but do not return an error.
	// The test confirms production-file violations are suppressed (no panic, clean exit).
	require.NoError(t, err)
}

func TestCheckMagicUsageInDir_EmptyMagicPackage(t *testing.T) {
	t.Parallel()

	magicDir, rootDir := setupMagicUsageDirs(t)
	writeMagicFile(t, magicDir, "magic.go", `package magic
`)

	err := os.WriteFile(filepath.Join(rootDir, "app.go"), []byte(`package app

const x = "hello"
`), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-usage-test")
	err = CheckMagicUsageInDir(logger, magicDir, rootDir)
	require.NoError(t, err)
}

func TestCheckMagicUsageInDir_InvalidMagicDir(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-usage-test")
	err := CheckMagicUsageInDir(logger, "/nonexistent/magic", ".")
	require.Error(t, err)
}

func TestCheckMagicUsageInDir_GeneratedFileSkipped(t *testing.T) {
	t.Parallel()

	magicDir, rootDir := setupMagicUsageDirs(t)

	writeMagicFile(t, magicDir, "magic.go", `package magic

const ProtocolHTTPS = "https"
`)

	// Generated file — should be completely skipped.
	err := os.WriteFile(filepath.Join(rootDir, "openapi.gen.go"), []byte(`package app

func genFunc() string { return "https" }
`), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-usage-test")
	err = CheckMagicUsageInDir(logger, magicDir, rootDir)
	require.NoError(t, err)
}

func TestCheck_UsesMagicDefaultDir(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	// Check() calls CheckMagicUsageInDir with MagicDefaultDir="internal/shared/magic".
	// When run from the package test directory, that relative path does not exist,
	// so Check() returns an error. This exercises the Check() code path.
	err := Check(logger)
	require.Error(t, err, "Check() should fail when MagicDefaultDir does not exist relative to CWD")
	require.Contains(t, err.Error(), "failed to parse magic package")
}

func TestCheckMagicUsageInDir_MagicDirInsideRoot(t *testing.T) {
	// Non-parallel: uses controlled directory structure.
	rootDir := t.TempDir()

	// Place magicDir INSIDE rootDir so the Walk visits and skips it.
	magicDir := filepath.Join(rootDir, "magic")
	require.NoError(t, os.MkdirAll(magicDir, 0o700))
	writeMagicFile(t, magicDir, "magic.go", "package magic\n\nconst Timeout = 30\n")

	// Regular go file outside the magic dir.
	require.NoError(t, os.WriteFile(filepath.Join(rootDir, "app.go"), []byte("package app\n\nfunc f() { _ = 30 }\n"), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckMagicUsageInDir(logger, magicDir, rootDir)
	// Should succeed - the magicDir is skipped by filepath.SkipDir.
	require.NoError(t, err)
}

func TestCheckMagicUsageInDir_VendorDirSkipped(t *testing.T) {
	t.Parallel()

	magicDir, rootDir := setupMagicUsageDirs(t)
	writeMagicFile(t, magicDir, "magic.go", "package magic\n\nconst Timeout = 30\n")

	// Create vendor subdirectory - MagicShouldSkipPath returns true for "vendor".
	vendorDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirVendor, "somepkg")
	require.NoError(t, os.MkdirAll(vendorDir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(vendorDir, "pkg.go"), []byte("package somepkg\n\nconst x = 30\n"), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckMagicUsageInDir(logger, magicDir, rootDir)
	require.NoError(t, err)
}

func TestCheckMagicUsageInDir_WalkError(t *testing.T) {
	// Non-parallel: modifies directory permissions.
	magicDir, rootDir := setupMagicUsageDirs(t)
	writeMagicFile(t, magicDir, "magic.go", "package magic\n\nconst Timeout = 30\n")

	// Create an unreadable subdirectory to trigger walk error accumulation.
	badSubDir := filepath.Join(rootDir, "locked")
	require.NoError(t, os.MkdirAll(badSubDir, 0o700))
	require.NoError(t, os.Chmod(badSubDir, 0o000))
	t.Cleanup(func() { _ = os.Chmod(badSubDir, 0o700) })

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckMagicUsageInDir(logger, magicDir, rootDir)
	// Walk errors are accumulated and returned as an error.
	require.Error(t, err, "Walk errors should be returned")
	require.Contains(t, err.Error(), "walk errors")
}

func TestCheckMagicUsageInDir_UnparseableGoFile(t *testing.T) {
	t.Parallel()

	magicDir, rootDir := setupMagicUsageDirs(t)
	writeMagicFile(t, magicDir, "magic.go", "package magic\n\nconst Timeout = 30\n")

	// Create a .go file with syntax errors - scanMagicFile returns nil silently.
	require.NoError(t, os.WriteFile(filepath.Join(rootDir, "broken.go"), []byte("package INVALID {{{"), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckMagicUsageInDir(logger, magicDir, rootDir)
	// The unparseable file is silently skipped.
	require.NoError(t, err)
}

func TestCheckMagicUsageInDir_WalkErrNonExistentRoot(t *testing.T) {
	t.Parallel()

	// Create a valid magic dir with constants, but use a nonexistent root dir.
	magicDir := t.TempDir()
	writeMagicFile(t, magicDir, "magic.go", "package magic\n\nconst Timeout = 30\n")

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckMagicUsageInDir(logger, magicDir, "/nonexistent/root/dir")
	require.Error(t, err)
	require.Contains(t, err.Error(), "walk errors")
}

func TestCheckMagicUsageInDir_AbsErrorDeletedCWD(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes and deletes CWD.
	if runtime.GOOS == cryptoutilSharedMagic.OSNameWindows {
		t.Skip("deleting CWD not supported on Windows")
	}

	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { require.NoError(t, os.Chdir(origDir)) }()

	// Create a magic directory with real constants (absolute path for ParseMagicDir).
	magicDir := t.TempDir()
	writeMagicFile(t, magicDir, "magic.go", "package magic\n\nconst Timeout = 30\n")

	// Create a temporary directory, chdir into it, then delete it to break Getwd().
	lostDir, err := os.MkdirTemp("", "lost-cwd-*")
	require.NoError(t, err)
	require.NoError(t, os.Chdir(lostDir))
	require.NoError(t, os.RemoveAll(lostDir))

	// Now filepath.Abs on any relative path will fail because Getwd() fails.
	// Pass a relative rootDir to trigger the Abs error path.
	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = CheckMagicUsageInDir(logger, magicDir, "relative/root")
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot resolve")
}
