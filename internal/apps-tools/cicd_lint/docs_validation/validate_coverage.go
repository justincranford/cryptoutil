// Copyright (c) 2025-2026 Justin Cranford.
package docs_validation

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const sectionToAppendixMatchLen = 3

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
	CompositionIssues  []string // appendix-propagate missing a matching section-to-appendix mapping
	ManifestChunks     int
	ArchitectureChunks int
}

type sectionContribution struct {
	AppendixID string
	ChunkID    string
	LineNumber int
}

type appendixPropagation struct {
	AppendixID string
	ChunkID    string
	TargetFile string
	LineNumber int
}

type directPropagation struct {
	ChunkID    string
	TargetFile string
	LineNumber int
}

// sourceBlockRegex matches <!-- @source from="..." as="CHUNK_ID" --> on a single line.
var sourceBlockRegex = regexp.MustCompile(`<!--\s+@source\s+from="[^"]+"\s+as="([^"]+)"\s+-->`)

// propagateMarkerRegex matches <!-- @propagate to="..." as="CHUNK_ID" --> (single-line form).
var propagateMarkerRegex = regexp.MustCompile(`<!--\s+@propagate\s+to="([^"]+)"\s+as="([^"]+)"\s+-->`)

// appendixPropagateMarkerRegex matches <!-- @appendix-propagate from="..." to="..." as="CHUNK_ID" -->.
var appendixPropagateMarkerRegex = regexp.MustCompile(`<!--\s+@appendix-propagate\s+from="([^"]+)"\s+to="([^"]+)"\s+as="([^"]+)"\s+-->`)

// sectionToAppendixMarkerRegex matches <!-- @section-to-appendix to="..." as="CHUNK_ID" -->.
var sectionToAppendixMarkerRegex = regexp.MustCompile(`<!--\s+@section-to-appendix\s+to="([^"]+)"\s+as="([^"]+)"\s+-->`)

// appendixWhyMarkerRegex matches <!-- @appendix-why from="..." why-this-exists="..." -->.
var appendixWhyMarkerRegex = regexp.MustCompile(`<!--\s+@appendix-why\s+from="([^"]+)"\s+why-this-exists="([^"]+)"\s+-->`)

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

		appendixMatch := appendixPropagateMarkerRegex.FindStringSubmatch(line)
		if len(appendixMatch) == 4 { //nolint:mnd // full match plus three capture groups
			chunkID := appendixMatch[3]
			if !chunkIDRegex.MatchString(chunkID) {
				continue
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

func extractSectionAppendixMappings(readFile func(string) ([]byte, error)) (
	map[string]map[string]bool,
	map[string]map[string]bool,
	[]sectionContribution,
	[]appendixPropagation,
	[]directPropagation,
	[]string,
	error,
) {
	data, err := readFile("docs/ENG-HANDBOOK.md")
	if err != nil {
		return nil, nil, nil, nil, nil, nil, fmt.Errorf("failed to read docs/ENG-HANDBOOK.md: %w", err)
	}

	sectionMappings := make(map[string]map[string]bool)
	appendixMappings := make(map[string]map[string]bool)

	var sectionEdges []sectionContribution

	var appendixEdges []appendixPropagation

	var directEdges []directPropagation

	var issues []string

	lines := strings.Split(string(data), "\n")

	for idx, line := range lines {
		lineNumber := idx + 1

		sectionMatch := sectionToAppendixMarkerRegex.FindStringSubmatch(line)
		if len(sectionMatch) == sectionToAppendixMatchLen {
			appendixID := sectionMatch[1]

			chunkID := sectionMatch[2]
			if unstableChunkIDRegex.MatchString(chunkID) {
				issues = append(issues, fmt.Sprintf("unstable semantic chunk id %q at line %d uses section-like numerics", chunkID, lineNumber))
			}

			if chunkIDRegex.MatchString(chunkID) {
				if sectionMappings[chunkID] == nil {
					sectionMappings[chunkID] = make(map[string]bool)
				}

				sectionMappings[chunkID][appendixID] = true
				sectionEdges = append(sectionEdges, sectionContribution{
					AppendixID: appendixID,
					ChunkID:    chunkID,
					LineNumber: lineNumber,
				})
			}
		}

		directMatch := propagateMarkerRegex.FindStringSubmatch(line)
		if len(directMatch) >= cryptoutilSharedMagic.PropagateMarkerMatchGroups {
			targetCSV := directMatch[1]
			chunkID := directMatch[2]

			if chunkIDRegex.MatchString(chunkID) {
				targets := strings.Split(targetCSV, ", ")
				for _, target := range targets {
					directEdges = append(directEdges, directPropagation{
						ChunkID:    chunkID,
						TargetFile: target,
						LineNumber: lineNumber,
					})
				}
			}
		}

		appendixMatch := appendixPropagateMarkerRegex.FindStringSubmatch(line)
		if len(appendixMatch) == 4 { //nolint:mnd // full match plus three capture groups
			appendixID := appendixMatch[1]
			targetCSV := appendixMatch[2]

			chunkID := appendixMatch[3]
			if unstableChunkIDRegex.MatchString(chunkID) {
				issues = append(issues, fmt.Sprintf("unstable semantic chunk id %q at line %d uses section-like numerics", chunkID, lineNumber))
			}

			hasWhy := false

			if idx > 0 {
				prev := strings.TrimSpace(lines[idx-1])

				whyMatch := appendixWhyMarkerRegex.FindStringSubmatch(prev)
				if len(whyMatch) == 3 && whyMatch[1] == appendixID && len(strings.TrimSpace(whyMatch[2])) >= cryptoutilSharedMagic.IdentityDefaultMaxIdleConns {
					hasWhy = true
				}
			}

			if !hasWhy {
				issues = append(issues, fmt.Sprintf("appendix-propagate block at line %d for appendix %q must have adjacent @appendix-why note with why-this-exists text", lineNumber, appendixID))
			}

			if chunkIDRegex.MatchString(chunkID) {
				if appendixMappings[chunkID] == nil {
					appendixMappings[chunkID] = make(map[string]bool)
				}

				appendixMappings[chunkID][appendixID] = true

				targets := strings.Split(targetCSV, ", ")
				for _, target := range targets {
					appendixEdges = append(appendixEdges, appendixPropagation{
						AppendixID: appendixID,
						ChunkID:    chunkID,
						TargetFile: target,
						LineNumber: lineNumber,
					})
				}
			}
		}
	}

	return sectionMappings, appendixMappings, sectionEdges, appendixEdges, directEdges, issues, nil
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

	sectionMappings, appendixMappings, sectionEdges, appendixEdges, directEdges, additionalIssues, err := extractSectionAppendixMappings(readFile)
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

	// Item 1: reject orphan appendix blocks with no downstream appendix-propagate targets.
	appendixHasDownstream := make(map[string]bool)

	for _, edge := range appendixEdges {
		if edge.TargetFile != "" {
			appendixHasDownstream[edge.AppendixID] = true
		}
	}

	seenAppendix := make(map[string]bool)
	for _, edge := range sectionEdges {
		if seenAppendix[edge.AppendixID] {
			continue
		}

		seenAppendix[edge.AppendixID] = true
		if !appendixHasDownstream[edge.AppendixID] {
			result.CompositionIssues = append(result.CompositionIssues,
				fmt.Sprintf("orphan appendix %q has semantic contributions but no appendix-propagate downstream target", edge.AppendixID),
			)
		}
	}

	// Item 2: reject semantic contribution blocks that do not feed any appendix block.
	for _, edge := range sectionEdges {
		if appendixMappings[edge.ChunkID] == nil || !appendixMappings[edge.ChunkID][edge.AppendixID] {
			result.CompositionIssues = append(result.CompositionIssues,
				fmt.Sprintf("section-to-appendix chunk %q to %q at line %d does not feed any appendix-propagate block", edge.ChunkID, edge.AppendixID, edge.LineNumber),
			)
		}
	}

	// Reverse mapping check: appendix-propagate must map back to semantic contribution.
	for _, edge := range appendixEdges {
		if sectionMappings[edge.ChunkID] == nil || !sectionMappings[edge.ChunkID][edge.AppendixID] {
			result.CompositionIssues = append(result.CompositionIssues,
				fmt.Sprintf("appendix-propagate chunk %q from %q at line %d has no matching @section-to-appendix mapping", edge.ChunkID, edge.AppendixID, edge.LineNumber),
			)
		}
	}

	// Item 3: reject direct downstream propagation for semantic chunks.
	for _, edge := range directEdges {
		if sectionMappings[edge.ChunkID] != nil {
			result.CompositionIssues = append(result.CompositionIssues,
				fmt.Sprintf("direct @propagate for semantic chunk %q to %s at line %d is not allowed; use appendix-propagate", edge.ChunkID, edge.TargetFile, edge.LineNumber),
			)
		}
	}

	// Item 4 + 5: include parse-time issues (unstable IDs + missing why-this-exists notes).
	result.CompositionIssues = append(result.CompositionIssues, additionalIssues...)

	sort.Slice(result.Violations, func(i, j int) bool {
		if result.Violations[i].ChunkID != result.Violations[j].ChunkID {
			return result.Violations[i].ChunkID < result.Violations[j].ChunkID
		}

		return result.Violations[i].File < result.Violations[j].File
	})

	sort.Strings(result.OrphanedChunks)

	return result, nil
}
