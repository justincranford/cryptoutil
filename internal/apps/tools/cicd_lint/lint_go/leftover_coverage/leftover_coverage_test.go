// Copyright (c) 2025 Justin Cranford

package leftover_coverage_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	leftoverCoverage "cryptoutil/internal/apps/tools/cicd_lint/lint_go/leftover_coverage"
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

func TestCheckInDir_ExceptionMatchesParentDir(t *testing.T) {
	t.Parallel()

	// cicd_coverage/cicd_coverage_test.go is allowed because stem == parent dir name.
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "cicd_coverage")
	require.NoError(t, os.MkdirAll(pkgDir, 0o750))
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

// writeFile creates a file with given name and content in dir.
func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()

	err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600)
	require.NoError(t, err)
}
