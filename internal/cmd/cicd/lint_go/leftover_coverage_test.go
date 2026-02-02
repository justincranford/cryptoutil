// Copyright (c) 2025 Justin Cranford

package lint_go

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/cmd/cicd/common"

	"github.com/stretchr/testify/require"
)

func TestMatchesCoveragePattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{name: "matches *.out", filename: "coverage.out", want: true},
		{name: "matches *.cov", filename: "profile.cov", want: true},
		{name: "matches *.prof", filename: "cpu.prof", want: true},
		{name: "matches *coverage*.html", filename: "mycoverage_report.html", want: true},
		{name: "matches *coverage*.txt", filename: "test_coverage_data.txt", want: true},
		{name: "does not match .go", filename: "main.go", want: false},
		{name: "does not match .md", filename: "README.md", want: false},
		{name: "does not match random txt", filename: "notes.txt", want: false},
		{name: "does not match yaml", filename: "config.yaml", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := matchesCoveragePattern(tc.filename)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestCheckLeftoverCoverage_NoFiles(t *testing.T) {
	// NOTE: Cannot use t.Parallel() because os.Chdir() affects global state.

	// Create a temporary directory with no coverage files.
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	defer func() { _ = os.Chdir(origDir) }()

	// Create some non-coverage files.
	err = os.WriteFile("main.go", []byte("package main"), 0o600)
	require.NoError(t, err)

	err = os.WriteFile("README.md", []byte("# Test"), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("leftover-coverage-test")
	err = checkLeftoverCoverage(logger)
	require.NoError(t, err)
}

func TestCheckLeftoverCoverage_WithCoverageFiles(t *testing.T) {
	// NOTE: Cannot use t.Parallel() because os.Chdir() affects global state.

	// Create a temporary directory with coverage files.
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	defer func() { _ = os.Chdir(origDir) }()

	// Create coverage files that should be deleted.
	coverageFiles := []string{
		"coverage.out",
		"profile.cov",
		"cpu.prof",
		filepath.Join("subdir", "test_coverage.txt"),
	}

	for _, cf := range coverageFiles {
		dir := filepath.Dir(cf)
		if dir != "." {
			err = os.MkdirAll(dir, 0o700)
			require.NoError(t, err)
		}

		err = os.WriteFile(cf, []byte("coverage data"), 0o600)
		require.NoError(t, err)
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("leftover-coverage-test")
	err = checkLeftoverCoverage(logger)

	// Should return error because files were found and deleted.
	require.Error(t, err)
	require.Contains(t, err.Error(), "found and deleted")

	// Verify files were actually deleted.
	for _, cf := range coverageFiles {
		_, err = os.Stat(cf)
		require.True(t, os.IsNotExist(err), "File should have been deleted: %s", cf)
	}
}
