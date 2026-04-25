// Copyright (c) 2025 Justin Cranford

package duplicate_ps_id_test_func_names_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessDuplicatePSIDTestFuncNames "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/duplicate_ps_id_test_func_names"

	"github.com/stretchr/testify/require"
)

// writeTestFile creates a Go test file with the given function names under dir.
func writeTestFile(t *testing.T, dir, psID, filename string, funcNames []string) {
	t.Helper()

	serverDir := filepath.Join(dir, "internal", "apps", psID, "server")
	require.NoError(t, os.MkdirAll(serverDir, 0o755))

	var sb strings.Builder
	sb.WriteString("package server_test\nimport \"testing\"\n")

	for _, fn := range funcNames {
		sb.WriteString("func " + fn + "(t *testing.T) { t.Parallel() }\n")
	}

	require.NoError(t, os.WriteFile(filepath.Join(serverDir, filename), []byte(sb.String()), 0o600))
}

func TestFindDuplicates_NoDuplicates_EmptyDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	results, err := lintFitnessDuplicatePSIDTestFuncNames.FindDuplicates(dir)

	require.NoError(t, err)
	require.Empty(t, results)
}

func TestFindDuplicates_BelowThreshold_NotReported(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Same function in only 2 PS-IDs — below threshold of 3.
	writeTestFile(t, dir, "sm-kms", "server_test.go", []string{"TestNewFromConfig_NilContext"})
	writeTestFile(t, dir, "jose-ja", "server_test.go", []string{"TestNewFromConfig_NilContext"})

	results, err := lintFitnessDuplicatePSIDTestFuncNames.FindDuplicates(dir)

	require.NoError(t, err)
	require.Empty(t, results)
}

func TestFindDuplicates_AtThreshold_Reported(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Same function in exactly 3 PS-IDs — at threshold, should be reported.
	writeTestFile(t, dir, "sm-kms", "server_test.go", []string{"TestNewFromConfig_NilContext"})
	writeTestFile(t, dir, "jose-ja", "server_test.go", []string{"TestNewFromConfig_NilContext"})
	writeTestFile(t, dir, "pki-ca", "server_test.go", []string{"TestNewFromConfig_NilContext"})

	results, err := lintFitnessDuplicatePSIDTestFuncNames.FindDuplicates(dir)

	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, "TestNewFromConfig_NilContext", results[0].FuncName)
	require.Len(t, results[0].PSIDs, 3)
}

func TestFindDuplicates_RankedWorstFirst(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// FuncA in 4 PS-IDs, FuncB in 3 PS-IDs — FuncA should come first.
	for _, psID := range []string{"sm-kms", "jose-ja", "pki-ca", "sm-im"} {
		writeTestFile(t, dir, psID, "server_test.go", []string{"TestFuncA", "TestFuncB"})
	}

	// FuncB only in 3 of the 4.
	writeTestFile(t, dir, "skeleton-template", "server_test.go", []string{"TestFuncA"})

	results, err := lintFitnessDuplicatePSIDTestFuncNames.FindDuplicates(dir)

	require.NoError(t, err)
	require.True(t, len(results) >= 2)
	// First result should have the higher occurrence count.
	require.GreaterOrEqual(t, len(results[0].PSIDs), len(results[1].PSIDs))
}

func TestFindDuplicates_IntegrationTestsExcluded(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Integration test files should not be considered — function in them should not count.
	for _, psID := range []string{"sm-kms", "jose-ja", "pki-ca"} {
		serverDir := filepath.Join(dir, "internal", "apps", psID, "server")
		require.NoError(t, os.MkdirAll(serverDir, 0o755))

		content := "package server_test\nimport \"testing\"\nfunc TestSomething(t *testing.T) { t.Parallel() }\n"
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, "server_integration_test.go"), []byte(content), 0o600))
	}

	results, err := lintFitnessDuplicatePSIDTestFuncNames.FindDuplicates(dir)

	require.NoError(t, err)
	require.Empty(t, results)
}

func TestFindDuplicates_FrameworkExcluded(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	fwDir := filepath.Join(dir, "internal", "apps", "framework", "service", "server")
	require.NoError(t, os.MkdirAll(fwDir, 0o755))

	content := "package server_test\nimport \"testing\"\nfunc TestNewFromConfig_NilContext(t *testing.T) {}\n"
	require.NoError(t, os.WriteFile(filepath.Join(fwDir, "server_test.go"), []byte(content), 0o600))

	// Framework alone should not trigger duplication.
	results, err := lintFitnessDuplicatePSIDTestFuncNames.FindDuplicates(dir)

	require.NoError(t, err)
	require.Empty(t, results)
}

func TestFindDuplicates_NestedSubPackagesExcluded(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Files under server/handler/ (nested) should not be considered — only server/ direct.
	for _, psID := range []string{"sm-kms", "jose-ja", "pki-ca"} {
		handlerDir := filepath.Join(dir, "internal", "apps", psID, "server", "handler")
		require.NoError(t, os.MkdirAll(handlerDir, 0o755))

		content := "package handler_test\nimport \"testing\"\nfunc TestHandlerFn(t *testing.T) {}\n"
		require.NoError(t, os.WriteFile(filepath.Join(handlerDir, "handler_test.go"), []byte(content), 0o600))
	}

	results, err := lintFitnessDuplicatePSIDTestFuncNames.FindDuplicates(dir)

	require.NoError(t, err)
	require.Empty(t, results)
}

func TestCheckInDir_ReturnsNil_WhenDuplicatesFound(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	for _, psID := range []string{"sm-kms", "jose-ja", "pki-ca"} {
		writeTestFile(t, dir, psID, "server_test.go", []string{"TestNewFromConfig_NilContext"})
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test-duplicate-ps-id-test-func-names")
	err := lintFitnessDuplicatePSIDTestFuncNames.CheckInDir(logger, dir)

	// Informational linter — logs violations but does not fail.
	require.NoError(t, err)
}

func TestCheckInDir_ReturnsNil_WhenClean(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logger := cryptoutilCmdCicdCommon.NewLogger("test-duplicate-ps-id-test-func-names")
	err := lintFitnessDuplicatePSIDTestFuncNames.CheckInDir(logger, dir)

	require.NoError(t, err)
}
