// Copyright (c) 2025 Justin Cranford

package magic_constant_location_test

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessMagicConstantLocation "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/magic_constant_location"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLogger() *cryptoutilCmdCicdCommon.Logger {
	return cryptoutilCmdCicdCommon.NewLogger("test")
}

func mkdir(t *testing.T, path string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(path, cryptoutilSharedMagic.DirPermissions))
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	dir := filepath.Dir(path)
	mkdir(t, dir)

	require.NoError(t, os.WriteFile(path, []byte(content), cryptoutilSharedMagic.FilePermissionsDefault))
}

func TestFindViolationsInDir_EmptyDir_NoViolations(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	violations, err := lintFitnessMagicConstantLocation.FindViolationsInDir(tmpDir)

	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestFindViolationsInDir_SuspiciousPortConst_Violation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	writeFile(t, filepath.Join(tmpDir, "pkg", "server.go"), `package pkg

const defaultPort = 8080
`)

	violations, err := lintFitnessMagicConstantLocation.FindViolationsInDir(tmpDir)

	require.NoError(t, err)
	require.Len(t, violations, 1)
	assert.Contains(t, violations[0], "defaultPort")
	assert.Contains(t, violations[0], "suspicious port range")
}

func TestFindViolationsInDir_SafeIntConst_NoViolation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	writeFile(t, filepath.Join(tmpDir, "pkg", "util.go"), `package pkg

const maxRetries = 3
`)

	violations, err := lintFitnessMagicConstantLocation.FindViolationsInDir(tmpDir)

	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestFindViolationsInDir_MagicPackageExcluded(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	writeFile(t, filepath.Join(tmpDir, "internal", "shared", "magic", "magic_ports.go"), `package magic

const DefaultPort = 8080
`)

	violations, err := lintFitnessMagicConstantLocation.FindViolationsInDir(tmpDir)

	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestFindViolationsInDir_TestFileExcluded(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	writeFile(t, filepath.Join(tmpDir, "pkg", "server_test.go"), `package pkg

const testPort = 8080
`)

	violations, err := lintFitnessMagicConstantLocation.FindViolationsInDir(tmpDir)

	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestFindViolationsInDir_GeneratedFileExcluded(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	writeFile(t, filepath.Join(tmpDir, "pkg", "server.gen.go"), `package pkg

const generatedPort = 8080
`)

	violations, err := lintFitnessMagicConstantLocation.FindViolationsInDir(tmpDir)

	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestFindViolationsInDir_ConstBlock_Violation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	writeFile(t, filepath.Join(tmpDir, "pkg", "config.go"), `package pkg

const (
	adminPort = 9090
	publicPort = 8080
)
`)

	violations, err := lintFitnessMagicConstantLocation.FindViolationsInDir(tmpDir)

	require.NoError(t, err)
	require.Len(t, violations, 2)
}

func TestFindViolationsInDir_BelowRange_NoViolation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	writeFile(t, filepath.Join(tmpDir, "pkg", "limit.go"), `package pkg

const maxItems = 999
`)

	violations, err := lintFitnessMagicConstantLocation.FindViolationsInDir(tmpDir)

	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestFindViolationsInDir_AboveRange_NoViolation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	writeFile(t, filepath.Join(tmpDir, "pkg", "big.go"), `package pkg

const bigNumber = 65536
`)

	violations, err := lintFitnessMagicConstantLocation.FindViolationsInDir(tmpDir)

	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestCheckInDir_Informational_NoError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	writeFile(t, filepath.Join(tmpDir, "pkg", "server.go"), `package pkg

const port = 8080
`)

	// CheckInDir is informational — always returns nil even with violations.
	err := lintFitnessMagicConstantLocation.CheckInDir(newTestLogger(), tmpDir)

	require.NoError(t, err)
}

func TestCheckInDir_Clean_NoError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	err := lintFitnessMagicConstantLocation.CheckInDir(newTestLogger(), tmpDir)

	require.NoError(t, err)
}

func TestFindViolationsInDir_Integration(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)

	violations, err := lintFitnessMagicConstantLocation.FindViolationsInDir(root)

	require.NoError(t, err)
	// Informational: violations expected but not blocking.
	t.Logf("Found %d suspicious const(s) outside internal/shared/magic/ (informational)", len(violations))
}

func findProjectRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	require.NoError(t, err, "failed to get working directory")

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Skip("skipping integration test: cannot find project root (no go.mod)")
		}

		dir = parent
	}
}
