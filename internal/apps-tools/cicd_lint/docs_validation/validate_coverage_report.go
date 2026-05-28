// Copyright (c) 2025-2026 Justin Cranford.
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

// extractSourceChunksFromContent scans a single file's content for @from-eng-handbook blocks.
func extractSourceChunksFromContent(relPath, content string, result map[string][]string) {
	for _, line := range strings.Split(content, "\n") {
		match := sourceBlockRegex.FindStringSubmatch(line)
		if len(match) >= 2 { //nolint:mnd // match[1] requires at least 2 elements (full + group)
			chunkID := match[1]
			if !chunkIDRegex.MatchString(chunkID) {
				continue // skip grammar examples and other non-conforming captures
			}

			result[chunkID] = append(result[chunkID], relPath)
		}
	}
}

// FormatCoverageValidationResults formats the validate-coverage results.
func FormatCoverageValidationResults(result *CoverageResult) string {
	var sb strings.Builder

	sb.WriteString("=== Propagation Coverage Validation ===\n\n")
	fmt.Fprintf(&sb, "Manifest chunks:      %d\n", result.ManifestChunks)
	fmt.Fprintf(&sb, "Architecture chunks:  %d\n", result.ArchitectureChunks)

	if len(result.OrphanedChunks) > 0 {
		fmt.Fprintf(&sb, "\nORPHANED CHUNKS (%d) - in ENG-HANDBOOK.md but missing from manifest:\n", len(result.OrphanedChunks))

		for _, id := range result.OrphanedChunks {
			fmt.Fprintf(&sb, "  - %s\n", id)
		}
	}

	if len(result.CompositionIssues) > 0 {
		fmt.Fprintf(&sb, "\nSECTION/APPENDIX COMPOSITION ISSUES (%d, review-order):\n", len(result.CompositionIssues))

		for _, issue := range result.CompositionIssues {
			fmt.Fprintf(&sb, "  - %s\n", issue)
		}
	}

	if len(result.Violations) > 0 {
		fmt.Fprintf(&sb, "\nMISSING @from-eng-handbook BLOCKS (%d):\n", len(result.Violations))

		for _, v := range result.Violations {
			fmt.Fprintf(&sb, "  - chunk=%s file=%s\n", v.ChunkID, v.File)
		}

		sb.WriteString("\nCoverage validation FAILED. Add missing @from-eng-handbook blocks or update the manifest.\n")
	} else if len(result.OrphanedChunks) > 0 {
		sb.WriteString("\nCoverage validation FAILED. Add orphaned chunks to docs/required-propagations.yaml.\n")
	} else if len(result.CompositionIssues) > 0 {
		sb.WriteString("\nCoverage validation FAILED. Fix @to-appendix marker composition issues.\n")
	} else {
		sb.WriteString("\nAll required @to-appendix chunks are covered by @from-eng-handbook blocks.\n")
	}

	return sb.String()
}

// ExtractSourceChunks scans all instruction/agent files and returns a map of
// chunk_id → sorted list of files that contain <!-- @from-eng-handbook as="chunk_id" -->.
func ExtractSourceChunks(rootDir string, readFile func(string) ([]byte, error)) (map[string][]string, error) {
	scanDirs := []struct {
		dir     string
		pattern string
	}{
		{dir: cryptoutilSharedMagic.CICDGithubInstructionsDir, pattern: cryptoutilSharedMagic.CICDInstructionsPattern},
		{dir: cryptoutilSharedMagic.CICDGithubAgentsDir, pattern: cryptoutilSharedMagic.CICDAgentsPattern},
		{dir: cryptoutilSharedMagic.CICDClaudeAgentsDir, pattern: cryptoutilSharedMagic.CICDClaudeAgentsPattern},
	}

	result := make(map[string][]string)

	// Scan copilot-instructions.md directly.
	if content, err := readFile(cryptoutilSharedMagic.CICDCopilotInstructionsFile); err == nil {
		extractSourceChunksFromContent(cryptoutilSharedMagic.CICDCopilotInstructionsFile, string(content), result)
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

			extractSourceChunksFromContent(relPath, string(content), result)
		}
	}

	// Scan skill pairs recursively: .github/skills/*/SKILL.md and .claude/skills/*/SKILL.md.
	skillDirs := []string{
		cryptoutilSharedMagic.CICDGithubSkillsDir,
		cryptoutilSharedMagic.CICDClaudeSkillsDir,
	}

	for _, skillsDir := range skillDirs {
		skillsDirPath := filepath.Join(rootDir, skillsDir)

		skillEntries, skillsErr := os.ReadDir(skillsDirPath)
		if skillsErr != nil {
			continue
		}

		for _, skillEntry := range skillEntries {
			if !skillEntry.IsDir() {
				continue
			}

			relPath := filepath.ToSlash(filepath.Join(skillsDir, skillEntry.Name(), cryptoutilSharedMagic.CICDSkillFileName))

			content, readErr := readFile(relPath)
			if readErr != nil {
				continue
			}

			extractSourceChunksFromContent(relPath, string(content), result)
		}
	}

	// Sort the file lists for deterministic output.
	for k := range result {
		sort.Strings(result[k])
	}

	return result, nil
}

// ValidateCoverageCommand is the CLI entry point for validate-coverage.
// Returns exit code: 0 if all chunks covered, 1 on any failure.
func ValidateCoverageCommand(stdout, stderr io.Writer) int {
	return validateCoverageCommand(stdout, stderr, findProjectRoot)
}

func validateCoverageCommand(stdout, stderr io.Writer, rootFn func() (string, error)) int {
	rootDir, err := rootFn()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error: %s\n", err)

		return 1
	}

	return validateCoverageWithRoot(rootDir, stdout, stderr)
}

// validateCoverageWithRoot runs validate-coverage using a specified root directory.
func validateCoverageWithRoot(rootDir string, stdout, stderr io.Writer) int {
	result, err := ValidateCoverage(rootDir, rootedReadFile(rootDir))
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error: %s\n", err)

		return 1
	}

	report := FormatCoverageValidationResults(result)
	_, _ = fmt.Fprint(stdout, report)

	if len(result.Violations) > 0 || len(result.OrphanedChunks) > 0 || len(result.CompositionIssues) > 0 {
		return 1
	}

	return 0
}
