// Copyright (c) 2025-2026 Justin Cranford.
package leftover_coverage_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	leftoverCoverage "cryptoutil/internal/apps-tools/cicd_lint/lint_go/leftover_coverage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestCheck_NoViolations(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger(t.Name())
	err := leftoverCoverage.Check(logger)

	require.NoError(t, err)
}

func TestCheckInDir_Clean(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, dir, "some_handler_test.go", "")
	writeFile(t, dir, "some_handler.go", "")

	logger := cryptoutilCmdCicdCommon.NewLogger(t.Name())
	err := leftoverCoverage.CheckInDir(logger, dir, filepath.Walk)

	require.NoError(t, err)
}

func TestCheckInDir_ExcludedDirectoriesDoNotContaminateScan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		dirName string
	}{
		{name: cryptoutilSharedMagic.CICDExcludeDirTestOutput, dirName: cryptoutilSharedMagic.CICDExcludeDirTestOutput},
		{name: cryptoutilSharedMagic.CICDExcludeDirVendor, dirName: cryptoutilSharedMagic.CICDExcludeDirVendor},
		{name: "git", dirName: cryptoutilSharedMagic.CICDExcludeDirGit},
		{name: "Python virtual environment", dirName: cryptoutilSharedMagic.CICDExcludeDirVenv},
		{name: "Node.js dependencies", dirName: cryptoutilSharedMagic.CICDExcludeDirNodeModules},
		{name: "workflow reports", dirName: cryptoutilSharedMagic.CICDExcludeDirWorkflowReports},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootDir := t.TempDir()
			repositoryDir := filepath.Join(rootDir, "repository")
			excludedDir := filepath.Join(rootDir, "nested", cryptoutilSharedMagic.CICDExcludeDirWorktree, tc.dirName, "nested")

			require.NoError(t, os.MkdirAll(repositoryDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))
			require.NoError(t, os.MkdirAll(excludedDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))
			writeFile(t, repositoryDir, "handler_coverage_test.go", "")
			writeFile(t, excludedDir, "ignored_coverage_test.go", "")

			err := leftoverCoverage.CheckInDir(
				cryptoutilCmdCicdCommon.NewLogger(t.Name()),
				rootDir,
				filepath.Walk,
			)

			require.ErrorContains(t, err, "1 banned test file name(s) found")
			require.Contains(t, err.Error(), filepath.Join("repository", "handler_coverage_test.go"))
		})
	}
}

func TestCheckInDir_BannedNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
	}{
		{name: "coverage suffix", filename: "handler_coverage_test.go"},
		{name: "coverage2 suffix", filename: "handler_coverage2_test.go"},
		{name: "comprehensive suffix", filename: "handler_comprehensive_test.go"},
		{name: "gaps suffix", filename: "handler_gaps_test.go"},
		{name: "coverage_gaps suffix", filename: "handler_coverage_gaps_test.go"},
		{name: "highcov suffix", filename: "handler_highcov_test.go"},
		{name: "extra suffix", filename: "handler_extra_test.go"},
		{name: "additional suffix", filename: "handler_additional_test.go"},
		{name: "edge_cases suffix", filename: "handler_edge_cases_test.go"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			writeFile(t, dir, tc.filename, "")

			logger := cryptoutilCmdCicdCommon.NewLogger(t.Name())
			err := leftoverCoverage.CheckInDir(logger, dir, filepath.Walk)

			require.Error(t, err)
			require.Contains(t, err.Error(), tc.filename)
		})
	}
}

func TestCheckInDir_RepositoryScanExcludedDirectories(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		excludedDir          string
		includeRealViolation bool
	}{
		{name: "nested git metadata", excludedDir: cryptoutilSharedMagic.CICDExcludeDirGit},
		{name: "nested vendored dependencies", excludedDir: cryptoutilSharedMagic.CICDExcludeDirVendor},
		{name: "nested generated test output", excludedDir: cryptoutilSharedMagic.CICDExcludeDirTestOutput},
		{
			name:                 "real violation remains detected",
			excludedDir:          cryptoutilSharedMagic.CICDExcludeDirTestOutput,
			includeRealViolation: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			excludedNestedDir := filepath.Join(dir, "nested", tc.excludedDir, "deeper")
			require.NoError(t, os.MkdirAll(excludedNestedDir, cryptoutilSharedMagic.CICDTempDirPermissions))

			excludedFilename := "generated_coverage_test.go"
			writeFile(t, excludedNestedDir, excludedFilename, "")

			realFilename := "handler_gaps_test.go"

			if tc.includeRealViolation {
				realDir := filepath.Join(dir, "src", "handler")
				require.NoError(t, os.MkdirAll(realDir, cryptoutilSharedMagic.CICDTempDirPermissions))
				writeFile(t, realDir, realFilename, "")
			}

			err := leftoverCoverage.CheckInDir(
				cryptoutilCmdCicdCommon.NewLogger(t.Name()),
				dir,
				filepath.Walk,
			)

			if tc.includeRealViolation {
				require.Error(t, err)
				require.ErrorContains(t, err, realFilename)
				require.NotContains(t, err.Error(), excludedFilename)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestCheckInDir_ExceptionMatchesParentDir(t *testing.T) {
	t.Parallel()

	// cicd_coverage/cicd_coverage_test.go is allowed because stem == parent dir name.
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "cicd_coverage")
	require.NoError(t, os.MkdirAll(pkgDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))
	writeFile(t, pkgDir, "cicd_coverage_test.go", "")

	logger := cryptoutilCmdCicdCommon.NewLogger(t.Name())
	err := leftoverCoverage.CheckInDir(logger, dir, filepath.Walk)

	require.NoError(t, err)
}

func TestCheckInDir_WalkError(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger(t.Name())
	errWalk := func(_ string, _ filepath.WalkFunc) error {
		return os.ErrPermission
	}

	err := leftoverCoverage.CheckInDir(logger, ".", errWalk)

	require.Error(t, err)
	require.Contains(t, err.Error(), "directory walk failed")
}

func TestCheckInDir_WalkCallbackError(t *testing.T) {
	t.Parallel()

	errWalk := func(_ string, walkFn filepath.WalkFunc) error {
		return walkFn("injected_test.go", nil, os.ErrPermission)
	}

	err := leftoverCoverage.CheckInDir(
		cryptoutilCmdCicdCommon.NewLogger(t.Name()),
		".",
		errWalk,
	)

	require.ErrorContains(t, err, "directory walk failed")
	require.ErrorIs(t, err, os.ErrPermission)
}

// writeFile creates a file with given name and content in dir.
func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()

	err := os.WriteFile(filepath.Join(dir, name), []byte(content), cryptoutilSharedMagic.FilePermissionsDefault)
	require.NoError(t, err)
}
