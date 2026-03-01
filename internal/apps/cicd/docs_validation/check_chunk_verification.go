// Copyright (c) 2025 Justin Cranford

package docs_validation

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ChunkMapping defines a mapping from an ARCHITECTURE.md section to an instruction file.
// MarkerText is a substring that MUST appear in the destination file to confirm propagation.
type ChunkMapping struct {
	ArchSection string // e.g., "12.4.11"
	Description string // e.g., "Validation Pipeline Architecture"
	DestFile    string // relative path from project root
	MarkerText  string // text to search for in the destination file
}

// ChunkVerificationResult holds the result of verifying a single chunk.
type ChunkVerificationResult struct {
	Mapping ChunkMapping
	Found   bool
	Error   error
}

// chunkMappings defines the hardcoded mapping from ARCHITECTURE.md sections to instruction files.
// Each entry specifies a section, its destination file, and a marker text to verify propagation.
func chunkMappings() []ChunkMapping {
	return []ChunkMapping{
		{
			ArchSection: "12.4.11",
			Description: "Validation Pipeline Architecture",
			DestFile:    ".github/instructions/04-01.deployment.instructions.md",
			MarkerText:  "Section 12.4.11",
		},
		{
			ArchSection: "12.5",
			Description: "Config File Architecture",
			DestFile:    ".github/instructions/02-01.architecture.instructions.md",
			MarkerText:  "Config File Architecture",
		},
		{
			ArchSection: "12.5",
			Description: "Config File Architecture (data-infra)",
			DestFile:    ".github/instructions/03-04.data-infrastructure.instructions.md",
			MarkerText:  "Config File Architecture",
		},
		{
			ArchSection: "12.6",
			Description: "Secrets Management in Deployments",
			DestFile:    ".github/instructions/04-01.deployment.instructions.md",
			MarkerText:  "Section 12.6",
		},
		{
			ArchSection: "12.6",
			Description: "Secrets Management in Deployments (security)",
			DestFile:    ".github/instructions/02-05.security.instructions.md",
			MarkerText:  "Section 12.6",
		},
		{
			ArchSection: "6.10",
			Description: "Secrets Detection Strategy",
			DestFile:    ".github/instructions/02-05.security.instructions.md",
			MarkerText:  "Secrets Detection Strategy",
		},
		{
			ArchSection: "11.2.5",
			Description: "Mutation Testing Scope",
			DestFile:    ".github/instructions/03-02.testing.instructions.md",
			MarkerText:  "cmd/cicd/",
		},
		{
			ArchSection: "12.7",
			Description: "Documentation Propagation Strategy",
			DestFile:    ".github/copilot-instructions.md",
			MarkerText:  "Documentation Propagation",
		},
		{
			ArchSection: "12.8",
			Description: "Validator Error Aggregation Pattern",
			DestFile:    ".github/instructions/03-01.coding.instructions.md",
			MarkerText:  "Validator Error Aggregation",
		},
		{
			ArchSection: "9.10",
			Description: "CICD Command Architecture",
			DestFile:    ".github/instructions/04-01.deployment.instructions.md",
			MarkerText:  "cicd-command-naming",
		},
		{
			ArchSection: "2.5",
			Description: "Mandatory Review Passes (beast-mode)",
			DestFile:    ".github/instructions/01-02.beast-mode.instructions.md",
			MarkerText:  "mandatory-review-passes",
		},
		{
			ArchSection: "2.5",
			Description: "Mandatory Review Passes (evidence-based)",
			DestFile:    ".github/instructions/06-01.evidence-based.instructions.md",
			MarkerText:  "mandatory-review-passes",
		},
	}
}

// VerifyChunks checks that all mapped chunks exist in their destination instruction files.
// It reads files using the provided readFile function (for testability).
// Returns results for each mapping and whether all passed.
func VerifyChunks(mappings []ChunkMapping, readFile func(string) ([]byte, error)) ([]ChunkVerificationResult, bool) {
	results := make([]ChunkVerificationResult, 0, len(mappings))
	allPassed := true

	for _, mapping := range mappings {
		result := ChunkVerificationResult{Mapping: mapping}

		content, err := readFile(mapping.DestFile)
		if err != nil {
			result.Error = fmt.Errorf("failed to read %s: %w", mapping.DestFile, err)
			result.Found = false
			allPassed = false

			results = append(results, result)

			continue
		}

		result.Found = strings.Contains(string(content), mapping.MarkerText)
		if !result.Found {
			allPassed = false
		}

		results = append(results, result)
	}

	return results, allPassed
}

// FormatVerificationResults formats verification results as a human-readable report.
func FormatVerificationResults(results []ChunkVerificationResult, allPassed bool) string {
	var sb strings.Builder

	sb.WriteString("=== Chunk Verification Report ===\n\n")

	passCount := 0
	failCount := 0

	for _, r := range results {
		switch {
		case r.Error != nil:
			failCount++

			sb.WriteString(fmt.Sprintf("FAIL [%s] %s\n", r.Mapping.ArchSection, r.Mapping.Description))
			sb.WriteString(fmt.Sprintf("     Error: %s\n", r.Error))
		case !r.Found:
			failCount++

			sb.WriteString(fmt.Sprintf("FAIL [%s] %s\n", r.Mapping.ArchSection, r.Mapping.Description))
			sb.WriteString(fmt.Sprintf("     Missing marker %q in %s\n", r.Mapping.MarkerText, r.Mapping.DestFile))
		default:
			passCount++

			sb.WriteString(fmt.Sprintf("PASS [%s] %s\n", r.Mapping.ArchSection, r.Mapping.Description))
		}
	}

	sb.WriteString(fmt.Sprintf("\n=== Summary: %d PASS, %d FAIL ===\n", passCount, failCount))

	if allPassed {
		sb.WriteString("All chunks verified successfully.\n")
	} else {
		sb.WriteString("Chunk verification FAILED. Fix missing chunks.\n")
	}

	return sb.String()
}

// rootedReadFile creates a file reader that prepends rootDir to relative paths.
func rootedReadFile(rootDir string) func(string) ([]byte, error) {
	return func(path string) ([]byte, error) {
		return os.ReadFile(filepath.Join(rootDir, path))
	}
}

// findProjectRoot walks up from cwd to find the directory containing go.mod.
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found in any parent directory")
		}

		dir = parent
	}
}

// CheckChunkVerification is the entry point for the check-chunk-verification subcommand.
// Returns exit code: 0 for all chunks verified, 1 for any missing.
func CheckChunkVerification(stdout, stderr io.Writer) int {
	rootDir, err := findProjectRoot()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error: %s\n", err)

		return 1
	}

	return checkChunkVerificationWithRoot(rootDir, stdout)
}

// checkChunkVerificationWithRoot verifies chunks using a specified root directory.
func checkChunkVerificationWithRoot(rootDir string, stdout io.Writer) int {
	mappings := chunkMappings()
	results, allPassed := VerifyChunks(mappings, rootedReadFile(rootDir))
	report := FormatVerificationResults(results, allPassed)

	_, _ = fmt.Fprint(stdout, report)

	if !allPassed {
		return 1
	}

	return 0
}
