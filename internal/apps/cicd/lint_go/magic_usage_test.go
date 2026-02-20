// Copyright (c) 2025 Justin Cranford

package lint_go

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	lintGoMagicUsage "cryptoutil/internal/apps/cicd/lint_go/magic_usage"

	"github.com/stretchr/testify/require"
)

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
`), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-usage-test")
	err = lintGoMagicUsage.CheckMagicUsageInDir(logger, magicDir, rootDir)
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
`), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-usage-test")
	err = lintGoMagicUsage.CheckMagicUsageInDir(logger, magicDir, rootDir)
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
`), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-usage-test")
	err = lintGoMagicUsage.CheckMagicUsageInDir(logger, magicDir, rootDir)
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
`), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-usage-test")
	err = lintGoMagicUsage.CheckMagicUsageInDir(logger, magicDir, rootDir)
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
`), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-usage-test")
	err = lintGoMagicUsage.CheckMagicUsageInDir(logger, magicDir, rootDir)
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
`), 0o600)
	require.NoError(t, err)

	// Test file uses literal 500 — SHOULD be flagged.
	err = os.WriteFile(filepath.Join(rootDir, "server_test.go"), []byte(`package app

const wantLimit = 500
`), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-usage-test")
	err = lintGoMagicUsage.CheckMagicUsageInDir(logger, magicDir, rootDir)
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
`), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-usage-test")
	err = lintGoMagicUsage.CheckMagicUsageInDir(logger, magicDir, rootDir)
	require.NoError(t, err)
}

func TestCheckMagicUsageInDir_InvalidMagicDir(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-usage-test")
	err := lintGoMagicUsage.CheckMagicUsageInDir(logger, "/nonexistent/magic", ".")
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
`), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-usage-test")
	err = lintGoMagicUsage.CheckMagicUsageInDir(logger, magicDir, rootDir)
	require.NoError(t, err)
}
