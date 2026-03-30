// Copyright (c) 2025 Justin Cranford

package docs_validation

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestComputeFileCoverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		content       string
		wantTotal     int
		wantCovered   int
		wantHasSource bool
	}{
		{
			name:          "no source blocks",
			content:       "line 1\nline 2\nline 3",
			wantTotal:     3,
			wantCovered:   0,
			wantHasSource: false,
		},
		{
			name:          "single source block",
			content:       "<!-- @source from=\"docs/ARCHITECTURE.md\" as=\"chunk\" -->\ncontent\n<!-- @/source -->",
			wantTotal:     3,
			wantCovered:   3,
			wantHasSource: true,
		},
		{
			name:          "multiple source blocks",
			content:       "<!-- @source from=\"docs/ARCHITECTURE.md\" as=\"a\" -->\n<!-- @/source -->\n<!-- @source from=\"docs/ARCHITECTURE.md\" as=\"b\" -->\n<!-- @/source -->",
			wantTotal:     4,
			wantCovered:   4,
			wantHasSource: true,
		},
		{
			name:          "empty file",
			content:       "",
			wantTotal:     1,
			wantCovered:   0,
			wantHasSource: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fc := computeFileCoverage("test.md", tc.content)

			require.Equal(t, tc.wantTotal, fc.TotalLines)
			require.Equal(t, tc.wantCovered, fc.CoveredLines)
			require.Equal(t, tc.wantHasSource, fc.HasSource)
		})
	}
}

func TestComputeCoverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		files          map[string]string
		wantTotalFiles int
		wantCovered    int
		wantZeroCount  int
	}{
		{
			name: "mixed coverage",
			files: map[string]string{
				cryptoutilSharedMagic.CICDCopilotInstructionsFile:       "plain text only\nno source blocks",
				".github/instructions/01.instructions.md": "before\n<!-- @source from=\"docs/ARCHITECTURE.md\" as=\"x\" -->\ncontent\n<!-- @/source -->\nafter",
			},
			wantTotalFiles: 2,
			wantCovered:    1,
			wantZeroCount:  1,
		},
		{
			name: "all covered",
			files: map[string]string{
				".github/instructions/01.instructions.md": "<!-- @source from=\"docs/ARCHITECTURE.md\" as=\"x\" -->\nline\n<!-- @/source -->",
			},
			wantTotalFiles: 1,
			wantCovered:    1,
			wantZeroCount:  0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create a temp directory structure.
			rootDir := t.TempDir()

			for relPath, content := range tc.files {
				fullPath := rootDir + "/" + relPath
				dir := fullPath[:strings.LastIndex(fullPath, "/")]
				require.NoError(t, os.MkdirAll(dir, 0o700))
				require.NoError(t, os.WriteFile(fullPath, []byte(content), cryptoutilSharedMagic.FilePermissionsDefault))
			}

			readFile := rootedReadFile(rootDir)
			result, err := ComputeCoverage(rootDir, readFile)

			require.NoError(t, err)
			require.Equal(t, tc.wantTotalFiles, result.TotalFiles)
			require.Equal(t, tc.wantCovered, result.CoveredFiles)
			require.Len(t, result.ZeroCoverageFiles, tc.wantZeroCount)
		})
	}
}

func TestFormatCoverageResults(t *testing.T) {
	t.Parallel()

	result := &PropagationCoverageResult{
		TotalFiles:        3,
		CoveredFiles:      2,
		ZeroCoverageFiles: []string{".github/agents/test.agent.md"},
		TotalLines:        4,
		CoveredLines:      1,
	}

	report := FormatCoverageResults(result)

	require.Contains(t, report, "FILE COVERAGE: 2/3 files have @source blocks (67%)")
	require.Contains(t, report, "LINE COVERAGE: 1/4 lines inside @source blocks (25%)")
	require.Contains(t, report, "ZERO COVERAGE FILES (1)")
	require.Contains(t, report, ".github/agents/test.agent.md")
	require.Contains(t, report, "67% file coverage")
	require.Contains(t, report, "25% line coverage")
}

func TestFormatCoverageResults_NoCoverage(t *testing.T) {
	t.Parallel()

	result := &PropagationCoverageResult{
		TotalFiles:        2,
		CoveredFiles:      0,
		ZeroCoverageFiles: []string{"a.md", "b.md"},
		TotalLines:        4,
		CoveredLines:      0,
	}

	report := FormatCoverageResults(result)

	require.Contains(t, report, "FILE COVERAGE: 0/2 files have @source blocks (0%)")
	require.Contains(t, report, "LINE COVERAGE: 0/4 lines inside @source blocks (0%)")
	require.Contains(t, report, "ZERO COVERAGE FILES (2)")
}

func TestFormatCoverageResults_EmptyInput(t *testing.T) {
	t.Parallel()

	result := &PropagationCoverageResult{}
	report := FormatCoverageResults(result)

	require.Contains(t, report, "0/0 files")
	require.Contains(t, report, "0% file coverage")
}

func TestPercentage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		covered  int
		total    int
		expected float64
	}{
		{name: "zero total", covered: 0, total: 0, expected: 0},
		{name: "full coverage", covered: 1, total: 1, expected: cryptoutilSharedMagic.PercentageBasis100},
		{name: "half coverage", covered: 1, total: 2, expected: cryptoutilSharedMagic.PercentageBasis100 / 2},
		{name: "zero covered", covered: 0, total: 1, expected: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := percentage(tc.covered, tc.total)
			require.InDelta(t, tc.expected, result, cryptoutilSharedMagic.Tolerance1Percent)
		})
	}
}

// Sequential: modifies package-level findProjectRootFn seam.
func TestPropagationCoverageCommand_RootError(t *testing.T) {
	orig := findProjectRootFn

	t.Cleanup(func() { findProjectRootFn = orig })

	findProjectRootFn = func() (string, error) {
		return "", fmt.Errorf("injected root error")
	}

	var stdout, stderr bytes.Buffer

	exitCode := PropagationCoverageCommand(&stdout, &stderr)

	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr.String(), "injected root error")
}

// Sequential: modifies package-level findProjectRootFn seam.
func TestPropagationCoverageCommand_Integration(t *testing.T) {
	orig := findProjectRootFn

	t.Cleanup(func() { findProjectRootFn = orig })

	rootDir := t.TempDir()

	// Create instruction file with @source block.
	instrDir := rootDir + "/.github/instructions"
	require.NoError(t, os.MkdirAll(instrDir, 0o700))
	require.NoError(t, os.WriteFile(instrDir+"/01.instructions.md", []byte("before\n<!-- @source from=\"docs/ARCHITECTURE.md\" as=\"chunk\" -->\nline\n<!-- @/source -->\nafter"), cryptoutilSharedMagic.FilePermissionsDefault))

	findProjectRootFn = func() (string, error) {
		return rootDir, nil
	}

	var stdout, stderr bytes.Buffer

	exitCode := PropagationCoverageCommand(&stdout, &stderr)

	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout.String(), "Propagation Coverage Report")
	require.Contains(t, stdout.String(), "1/1 files")
}
