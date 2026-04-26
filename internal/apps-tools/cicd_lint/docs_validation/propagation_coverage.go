// Copyright (c) 2025 Justin Cranford

package docs_validation

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// PropagationCoverageResult holds coverage metrics for @source block propagation.
type PropagationCoverageResult struct {
	TotalFiles        int
	CoveredFiles      int
	ZeroCoverageFiles []string
	TotalLines        int
	CoveredLines      int
}

// FileCoverage describes the source block coverage of a single instruction/agent file.
type FileCoverage struct {
	RelPath      string
	TotalLines   int
	CoveredLines int
	HasSource    bool
}

// ComputeCoverage analyzes instruction and agent files for @source block coverage.
func ComputeCoverage(rootDir string, readFile func(string) ([]byte, error)) (*PropagationCoverageResult, error) {
	scanDirs := []struct {
		dir     string
		pattern string
	}{
		{dir: cryptoutilSharedMagic.CICDGithubInstructionsDir, pattern: cryptoutilSharedMagic.CICDInstructionsPattern},
		{dir: cryptoutilSharedMagic.CICDGithubAgentsDir, pattern: cryptoutilSharedMagic.CICDAgentsPattern},
		{dir: cryptoutilSharedMagic.CICDClaudeAgentsDir, pattern: cryptoutilSharedMagic.CICDClaudeAgentsPattern},
	}

	var fileCoverages []FileCoverage

	// Scan copilot-instructions.md directly.
	copilotContent, err := readFile(cryptoutilSharedMagic.CICDCopilotInstructionsFile)
	if err == nil {
		fileCoverages = append(fileCoverages, computeFileCoverage(cryptoutilSharedMagic.CICDCopilotInstructionsFile, string(copilotContent)))
	}

	for _, sd := range scanDirs {
		dirPath := filepath.Join(rootDir, sd.dir)

		entries, dirErr := os.ReadDir(dirPath)
		if dirErr != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			matched, matchErr := filepath.Match(sd.pattern, entry.Name())
			if matchErr != nil || !matched {
				continue
			}

			relPath := filepath.ToSlash(filepath.Join(sd.dir, entry.Name()))

			content, readErr := readFile(relPath)
			if readErr != nil {
				continue
			}

			fileCoverages = append(fileCoverages, computeFileCoverage(relPath, string(content)))
		}
	}

	result := &PropagationCoverageResult{}

	for _, fc := range fileCoverages {
		result.TotalFiles++
		result.TotalLines += fc.TotalLines
		result.CoveredLines += fc.CoveredLines

		if fc.HasSource {
			result.CoveredFiles++
		} else {
			result.ZeroCoverageFiles = append(result.ZeroCoverageFiles, fc.RelPath)
		}
	}

	sort.Strings(result.ZeroCoverageFiles)

	return result, nil
}

// computeFileCoverage computes line-level coverage for a single file.
func computeFileCoverage(relPath, content string) FileCoverage {
	lines := strings.Split(content, "\n")
	fc := FileCoverage{
		RelPath:    relPath,
		TotalLines: len(lines),
	}

	inSource := false

	for _, line := range lines {
		switch {
		case strings.Contains(line, "<!-- @source ") && strings.Contains(line, " as=\""):
			inSource = true
			fc.HasSource = true
			fc.CoveredLines++ // The @source marker itself counts as covered.
		case strings.Contains(line, "<!-- @/source -->"):
			inSource = false
			fc.CoveredLines++ // The close marker counts as covered.
		case inSource:
			fc.CoveredLines++
		}
	}

	return fc
}

// FormatCoverageResults formats the coverage report.
func FormatCoverageResults(result *PropagationCoverageResult) string {
	var sb strings.Builder

	sb.WriteString("=== Propagation Coverage Report ===\n\n")

	// File coverage.
	filePct := percentage(result.CoveredFiles, result.TotalFiles)
	fmt.Fprintf(&sb, "FILE COVERAGE: %d/%d files have @source blocks (%.0f%%)\n", result.CoveredFiles, result.TotalFiles, filePct)

	// Line coverage.
	linePct := percentage(result.CoveredLines, result.TotalLines)
	fmt.Fprintf(&sb, "LINE COVERAGE: %d/%d lines inside @source blocks (%.0f%%)\n", result.CoveredLines, result.TotalLines, linePct)

	// Zero coverage files.
	if len(result.ZeroCoverageFiles) > 0 {
		fmt.Fprintf(&sb, "\nZERO COVERAGE FILES (%d):\n", len(result.ZeroCoverageFiles))

		for _, f := range result.ZeroCoverageFiles {
			fmt.Fprintf(&sb, "  - %s\n", f)
		}
	}

	fmt.Fprintf(&sb, "\n=== Summary: %d files, %.0f%% file coverage, %.0f%% line coverage ===\n", result.TotalFiles, filePct, linePct)

	return sb.String()
}

// percentage computes a safe percentage (returns 0 when total is 0).
func percentage(covered, total int) float64 {
	if total == 0 {
		return 0
	}

	return float64(covered) / float64(total) * cryptoutilSharedMagic.PercentageBasis100
}

// PropagationCoverageCommand is the entry point for the propagation-coverage sub-command.
func PropagationCoverageCommand(stdout, stderr io.Writer) int {
	return propagationCoverageCommand(stdout, stderr, findProjectRoot)
}

func propagationCoverageCommand(stdout, stderr io.Writer, rootFn func() (string, error)) int {
	rootDir, err := rootFn()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error: %s\n", err)

		return 1
	}

	return propagationCoverageWithRoot(rootDir, stdout, stderr)
}

// propagationCoverageWithRoot computes coverage using a specified root directory.
func propagationCoverageWithRoot(rootDir string, stdout, _ io.Writer) int {
	result, err := ComputeCoverage(rootDir, rootedReadFile(rootDir))
	if err != nil {
		_, _ = fmt.Fprintf(stdout, "Error: %s\n", err)

		return 1
	}

	report := FormatCoverageResults(result)
	_, _ = fmt.Fprint(stdout, report)

	return 0
}
