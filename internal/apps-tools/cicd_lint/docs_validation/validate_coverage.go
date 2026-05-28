// Copyright (c) 2025-2026 Justin Cranford.
package docs_validation

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const toAppendixCoverageMatchGroups = 3

// PropagationEntry describes a single required handbook chunk from the manifest.
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
	OrphanedChunks     []string // @to-appendix in ENG-HANDBOOK.md but missing from manifest
	CompositionIssues  []string // structural issues in @to-appendix marker declarations
	ManifestChunks     int
	ArchitectureChunks int
}

// sourceBlockRegex matches <!-- @from-eng-handbook as="CHUNK_ID" --> on a single line.
var sourceBlockRegex = regexp.MustCompile(`<!--\s+@from-eng-handbook\s+as="([^"]+)"\s+-->`)

// toAppendixMarkerRegex matches <!-- @to-appendix as="CHUNK_ID" appendixes="..." -->.
var toAppendixMarkerRegex = regexp.MustCompile(`<!--\s+@to-appendix\s+as="([^"]+)"\s+appendixes="([^"]+)"\s+-->`)

// unstableChunkIDRegex catches likely section-number-driven IDs (e.g., section-13-4-rules).
// A single numeric token (e.g., rfc-2119-keywords) is allowed; two adjacent numeric
// tokens joined by hyphen indicate section-number coupling.
var unstableChunkIDRegex = regexp.MustCompile(`[0-9]+-[0-9]+`)

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

// ExtractPropagateChunks returns all chunk IDs declared in ENG-HANDBOOK.md @to-appendix markers.
func ExtractPropagateChunks(readFile func(string) ([]byte, error)) ([]string, error) {
	data, err := readFile("docs/ENG-HANDBOOK.md")
	if err != nil {
		return nil, fmt.Errorf("failed to read docs/ENG-HANDBOOK.md: %w", err)
	}

	seen := make(map[string]bool)

	var chunks []string

	var insideCatalog bool

	for _, line := range strings.Split(string(data), "\n") {
		trimmedLine := strings.TrimSpace(line)
		// Skip content inside @file-catalog blocks: it is embedded file content, not handbook markers.
		if strings.HasPrefix(trimmedLine, cryptoutilSharedMagic.CICDFileCatalogStartMarker) &&
			!strings.Contains(trimmedLine, cryptoutilSharedMagic.CICDFileCatalogEndMarker) {
			insideCatalog = true

			continue
		}

		if trimmedLine == cryptoutilSharedMagic.CICDFileCatalogEndMarker ||
			trimmedLine == cryptoutilSharedMagic.CICDFileCatalogPairEndMarker {
			insideCatalog = false

			continue
		}

		if insideCatalog {
			continue
		}

		match := toAppendixMarkerRegex.FindStringSubmatch(line)
		if len(match) == toAppendixCoverageMatchGroups {
			chunkID := match[1]
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

func extractToAppendixMappings(readFile func(string) ([]byte, error)) (map[string][]string, map[string]int, []string, error) {
	data, err := readFile("docs/ENG-HANDBOOK.md")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to read docs/ENG-HANDBOOK.md: %w", err)
	}

	chunkTargets := make(map[string][]string)
	chunkLineNumbers := make(map[string]int)

	var issues []string

	lines := strings.Split(string(data), "\n")

	var insideCatalog bool

	for idx, line := range lines {
		lineNumber := idx + 1

		trimmedLine := strings.TrimSpace(line)
		// Skip content inside @file-catalog blocks: it is embedded file content, not handbook markers.
		if strings.HasPrefix(trimmedLine, cryptoutilSharedMagic.CICDFileCatalogStartMarker) &&
			!strings.Contains(trimmedLine, cryptoutilSharedMagic.CICDFileCatalogEndMarker) {
			insideCatalog = true

			continue
		}

		if trimmedLine == cryptoutilSharedMagic.CICDFileCatalogEndMarker ||
			trimmedLine == cryptoutilSharedMagic.CICDFileCatalogPairEndMarker {
			insideCatalog = false

			continue
		}

		if insideCatalog {
			continue
		}

		toAppendixMatch := toAppendixMarkerRegex.FindStringSubmatch(line)
		if len(toAppendixMatch) == 3 { //nolint:mnd // full match plus two capture groups
			chunkID := toAppendixMatch[1]
			if unstableChunkIDRegex.MatchString(chunkID) {
				issues = append(issues, fmt.Sprintf("unstable semantic chunk id %q at line %d uses section-like numerics", chunkID, lineNumber))
			}

			appendixesCSV := toAppendixMatch[2]

			targets := strings.Split(appendixesCSV, ", ")
			if len(targets) == 0 {
				issues = append(issues, fmt.Sprintf("to-appendix chunk %q at line %d must declare at least one appendix target", chunkID, lineNumber))

				continue
			}

			for _, target := range targets {
				if strings.TrimSpace(target) == "" {
					issues = append(issues, fmt.Sprintf("to-appendix chunk %q at line %d contains an empty appendix target", chunkID, lineNumber))
				}
			}

			if chunkIDRegex.MatchString(chunkID) {
				chunkTargets[chunkID] = targets
				chunkLineNumbers[chunkID] = lineNumber
			}
		}
	}

	return chunkTargets, chunkLineNumbers, issues, nil
}

// ValidateCoverage runs the full coverage validation:
//  1. Loads the required-propagations manifest.
//  2. Extracts @to-appendix chunk IDs from ENG-HANDBOOK.md.
//  3. Scans instruction/agent files for @from-eng-handbook chunk IDs.
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

	chunkTargets, chunkLineNumbers, additionalIssues, err := extractToAppendixMappings(readFile)
	if err != nil {
		return nil, err
	}

	// Build a set of chunk IDs declared in the manifest.
	manifestSet := make(map[string]bool, len(manifest.RequiredPropagations))
	for _, entry := range manifest.RequiredPropagations {
		manifestSet[entry.ChunkID] = true
	}

	result := &CoverageResult{
		ManifestChunks:     len(manifest.RequiredPropagations),
		ArchitectureChunks: len(archChunks),
	}

	// Check each manifest entry.
	for _, entry := range manifest.RequiredPropagations {
		for _, target := range entry.RequiredTargets {
			handbookTargets := chunkTargets[entry.ChunkID]
			if len(handbookTargets) > 0 {
				foundTarget := false

				for _, handbookTarget := range handbookTargets {
					if target == handbookTarget {
						foundTarget = true

						break
					}
				}

				if !foundTarget {
					result.CompositionIssues = append(result.CompositionIssues,
						fmt.Sprintf("manifest target %q for chunk %q is not listed in @to-appendix at line %d", target, entry.ChunkID, chunkLineNumbers[entry.ChunkID]),
					)
				}
			}

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
					Description: fmt.Sprintf("@from-eng-handbook block for chunk %q not found in %s", entry.ChunkID, target),
				})
			}
		}
	}

	// Find orphaned: @to-appendix in ENG-HANDBOOK.md but not in manifest.
	for _, id := range archChunks {
		if !manifestSet[id] {
			result.OrphanedChunks = append(result.OrphanedChunks, id)
		}
	}

	// Include parse-time issues.
	result.CompositionIssues = append(result.CompositionIssues, additionalIssues...)

	sort.Slice(result.Violations, func(i, j int) bool {
		li := chunkLineNumbers[result.Violations[i].ChunkID]
		lj := chunkLineNumbers[result.Violations[j].ChunkID]

		if li != lj {
			return li < lj
		}

		if result.Violations[i].ChunkID != result.Violations[j].ChunkID {
			return result.Violations[i].ChunkID < result.Violations[j].ChunkID
		}

		return result.Violations[i].File < result.Violations[j].File
	})

	sort.Strings(result.OrphanedChunks)

	return result, nil
}
