// Copyright (c) 2025 Justin Cranford

package docs_validation

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// PropagationEntry describes a single required @propagate chunk from the manifest.
type PropagationEntry struct {
	ChunkID         string   `yaml:"chunk_id"`
	SourceFile      string   `yaml:"source_file"`
	RequiredTargets []string `yaml:"required_targets"`
}

// PropagationsManifest is the root type for docs/required-propagations.yaml.
type PropagationsManifest struct {
	RequiredPropagations []PropagationEntry `yaml:"required_propagations"`
}

// CoverageViolation describes a single coverage validation failure.
type CoverageViolation struct {
	ChunkID     string
	File        string
	Description string
}

// CoverageResult holds the result of validate-coverage.
type CoverageResult struct {
	Violations         []CoverageViolation
	OrphanedChunks     []string // @propagate in ENG-HANDBOOK.md but missing from manifest
	ManifestChunks     int
	ArchitectureChunks int
}

// sourceBlockRegex matches <!-- @source from="..." as="CHUNK_ID" --> on a single line.
var sourceBlockRegex = regexp.MustCompile(`<!--\s+@source\s+from="[^"]+"\s+as="([^"]+)"\s+-->`)

// propagateMarkerRegex matches <!-- @propagate to="..." as="CHUNK_ID" --> (single-line form).
var propagateMarkerRegex = regexp.MustCompile(`<!--\s+@propagate\s+to="([^"]+)"\s+as="([^"]+)"\s+-->`)

// chunkIDRegex validates that a captured chunk ID matches the grammar: [a-z][a-z0-9-]*.
// This filters out false positives from code-block grammar examples in ENG-HANDBOOK.md.
var chunkIDRegex = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// LoadPropagationsManifest reads and parses the required-propagations YAML manifest.
func LoadPropagationsManifest(readFile func(string) ([]byte, error)) (*PropagationsManifest, error) {
	data, err := readFile(cryptoutilSharedMagic.CICDRequiredPropagationsManifest)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", cryptoutilSharedMagic.CICDRequiredPropagationsManifest, err)
	}

	var manifest PropagationsManifest
	if err = yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", cryptoutilSharedMagic.CICDRequiredPropagationsManifest, err)
	}

	return &manifest, nil
}

// ExtractPropagateChunks returns all chunk IDs declared in ENG-HANDBOOK.md @propagate markers.
func ExtractPropagateChunks(readFile func(string) ([]byte, error)) ([]string, error) {
	data, err := readFile("docs/ENG-HANDBOOK.md")
	if err != nil {
		return nil, fmt.Errorf("failed to read docs/ENG-HANDBOOK.md: %w", err)
	}

	seen := make(map[string]bool)

	var chunks []string

	for _, line := range strings.Split(string(data), "\n") {
		match := propagateMarkerRegex.FindStringSubmatch(line)
		if len(match) >= cryptoutilSharedMagic.PropagateMarkerMatchGroups {
			chunkID := match[2] // group 2 = as="CHUNK_ID"
			if !chunkIDRegex.MatchString(chunkID) {
				continue // skip grammar examples and other non-conforming captures
			}

			if !seen[chunkID] {
				seen[chunkID] = true
				chunks = append(chunks, chunkID)
			}
		}
	}

	sort.Strings(chunks)

	return chunks, nil
}

// ExtractSourceChunks scans all instruction/agent files and returns a map of
// chunk_id → sorted list of files that contain <!-- @source ... as="chunk_id" -->.
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

	// Sort the file lists for deterministic output.
	for k := range result {
		sort.Strings(result[k])
	}

	return result, nil
}

// extractSourceChunksFromContent scans a single file's content for @source blocks.
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

// ValidateCoverage runs the full coverage validation:
//  1. Loads the required-propagations manifest.
//  2. Extracts @propagate chunk IDs from ENG-HANDBOOK.md.
//  3. Scans instruction/agent files for @source chunk IDs.
//  4. Returns violations and orphaned chunks.
func ValidateCoverage(rootDir string, readFile func(string) ([]byte, error)) (*CoverageResult, error) {
	return validateCoverage(rootDir, readFile, ExtractSourceChunks)
}

func validateCoverage(rootDir string, readFile func(string) ([]byte, error), extractFn func(string, func(string) ([]byte, error)) (map[string][]string, error)) (*CoverageResult, error) {
	manifest, err := LoadPropagationsManifest(readFile)
	if err != nil {
		return nil, err
	}

	archChunks, err := ExtractPropagateChunks(readFile)
	if err != nil {
		return nil, err
	}

	sourceChunks, err := extractFn(rootDir, readFile)
	if err != nil {
		return nil, err
	}

	// Build a set of chunk IDs declared in the manifest.
	manifestSet := make(map[string]bool, len(manifest.RequiredPropagations))
	for _, entry := range manifest.RequiredPropagations {
		manifestSet[entry.ChunkID] = true
	}

	// Build a set of chunk IDs from ENG-HANDBOOK.md.
	archSet := make(map[string]bool, len(archChunks))
	for _, id := range archChunks {
		archSet[id] = true
	}

	result := &CoverageResult{
		ManifestChunks:     len(manifest.RequiredPropagations),
		ArchitectureChunks: len(archChunks),
	}

	// Check each manifest entry.
	for _, entry := range manifest.RequiredPropagations {
		for _, target := range entry.RequiredTargets {
			filesWithChunk := sourceChunks[entry.ChunkID]

			found := false

			for _, f := range filesWithChunk {
				if f == target {
					found = true

					break
				}
			}

			if !found {
				result.Violations = append(result.Violations, CoverageViolation{
					ChunkID:     entry.ChunkID,
					File:        target,
					Description: fmt.Sprintf("@source block for chunk %q not found in %s", entry.ChunkID, target),
				})
			}
		}
	}

	// Find orphaned: @propagate in ENG-HANDBOOK.md but not in manifest.
	for _, id := range archChunks {
		if !manifestSet[id] {
			result.OrphanedChunks = append(result.OrphanedChunks, id)
		}
	}

	sort.Slice(result.Violations, func(i, j int) bool {
		if result.Violations[i].ChunkID != result.Violations[j].ChunkID {
			return result.Violations[i].ChunkID < result.Violations[j].ChunkID
		}

		return result.Violations[i].File < result.Violations[j].File
	})

	sort.Strings(result.OrphanedChunks)

	return result, nil
}

// FormatCoverageValidationResults formats the validate-coverage results.
func FormatCoverageValidationResults(result *CoverageResult) string {
	var sb strings.Builder

	sb.WriteString("=== Propagation Coverage Validation ===\n\n")
	fmt.Fprintf(&sb, "Manifest chunks:      %d\n", result.ManifestChunks)
	fmt.Fprintf(&sb, "Architecture chunks:  %d\n", result.ArchitectureChunks)

	if len(result.OrphanedChunks) > 0 {
		fmt.Fprintf(&sb, "\nORPHANED CHUNKS (%d) — in ENG-HANDBOOK.md but missing from manifest:\n", len(result.OrphanedChunks))

		for _, id := range result.OrphanedChunks {
			fmt.Fprintf(&sb, "  - %s\n", id)
		}
	}

	if len(result.Violations) > 0 {
		fmt.Fprintf(&sb, "\nMISSING @SOURCE BLOCKS (%d):\n", len(result.Violations))

		for _, v := range result.Violations {
			fmt.Fprintf(&sb, "  - chunk=%s file=%s\n", v.ChunkID, v.File)
		}

		sb.WriteString("\nCoverage validation FAILED. Add missing @source blocks or update the manifest.\n")
	} else if len(result.OrphanedChunks) > 0 {
		sb.WriteString("\nCoverage validation FAILED. Add orphaned chunks to docs/required-propagations.yaml.\n")
	} else {
		sb.WriteString("\nAll required @propagate chunks are covered by @source blocks.\n")
	}

	return sb.String()
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

	if len(result.Violations) > 0 || len(result.OrphanedChunks) > 0 {
		return 1
	}

	return 0
}
