// Copyright (c) 2025-2026 Justin Cranford.
// Package docs_validation provides catalog-file linters for Appendix E of ENG-HANDBOOK.md.
//
// Two linters are provided:
//
//  1. CheckCatalogFiles: verifies that each @file-catalog / @file-catalog-pair entry in
//     ENG-HANDBOOK.md contains content that exactly matches the corresponding file(s) on disk.
//
//  2. CheckCatalogPropagation: verifies that every @appendix-propagate chunk whose target file
//     has a catalog entry contains a matching @source block inside that catalog entry's body.
package docs_validation

import (
	"fmt"
	"regexp"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// FileCatalogEntry represents a @file-catalog single-file entry extracted from ENG-HANDBOOK.md.
type FileCatalogEntry struct {
	Path       string // repo-relative path, e.g., "CLAUDE.md"
	Content    string // verbatim file content stored in the catalog
	LineNumber int    // 1-based line number of the opening marker
}

// FileCatalogPairEntry represents a @file-catalog-pair entry (shared body, two frontmatters).
type FileCatalogPairEntry struct {
	CopilotPath        string // e.g., ".github/agents/beast-mode.agent.md"
	ClaudePath         string // e.g., ".claude/agents/beast-mode.md"
	CopilotFrontmatter string // complete frontmatter text including --- delimiters
	ClaudeFrontmatter  string // complete frontmatter text including --- delimiters
	Body               string // shared body (everything after the frontmatter closing ---)
	LineNumber         int    // 1-based line number of the opening marker
}

// CatalogViolation is a single failed check in catalog validation.
type CatalogViolation struct {
	File   string // file path(s) involved
	Field  string // short category label
	Detail string // human-readable description
}

// CatalogFilesResult holds the result of the catalog-files check.
type CatalogFilesResult struct {
	Violations []CatalogViolation
	Checked    int
}

// CatalogPropagationResult holds the result of the catalog-propagation check.
type CatalogPropagationResult struct {
	Violations []CatalogViolation
	Checked    int
}

// fileCatalogRegex matches <!-- @file-catalog path="PATH" --> opening markers.
// Group 1: path value.
var fileCatalogRegex = regexp.MustCompile(`^<!-- @file-catalog path="([^"]+)" -->$`)

// fileCatalogPairRegex matches <!-- @file-catalog-pair copilot="..." claude="..." --> opening markers.
// Group 1: copilot path; Group 2: claude path.
var fileCatalogPairRegex = regexp.MustCompile(`^<!-- @file-catalog-pair copilot="([^"]+)" claude="([^"]+)" -->$`)

// extractFileCatalogEntries parses all @file-catalog single-file blocks from handbook content.
func extractFileCatalogEntries(content string) []FileCatalogEntry {
	var entries []FileCatalogEntry

	lines := strings.Split(content, "\n")

	var current *FileCatalogEntry

	contentLines := make([]string, 0, len(lines))

	for i, line := range lines {
		lineNum := i + 1

		if current == nil {
			if match := fileCatalogRegex.FindStringSubmatch(line); len(match) == cryptoutilSharedMagic.FileCatalogSingleMatchGroups {
				current = &FileCatalogEntry{
					Path:       match[1],
					LineNumber: lineNum,
				}
				contentLines = nil

				continue
			}

			continue
		}

		if strings.TrimSpace(line) == cryptoutilSharedMagic.CICDFileCatalogEndMarker {
			// Add trailing newline; no stripping — content starts at the first line after the marker.
			raw := strings.Join(contentLines, "\n") + "\n"

			current.Content = raw
			entries = append(entries, *current)
			current = nil
			contentLines = nil

			continue
		}

		contentLines = append(contentLines, line)
	}

	return entries
}

// pairScanState tracks which sub-section of a @file-catalog-pair we are currently parsing.
type pairScanState int

const (
	pairScanStateNone      pairScanState = iota // between pair markers or not in a pair
	pairScanStateCopilotFM                      // inside <!-- @copilot-frontmatter:start --> block
	pairScanStateClaudeFM                       // inside <!-- @claude-frontmatter:start --> block
	pairScanStateBody                           // inside <!-- @file-body:start --> block
)

// extractFileCatalogPairEntries parses all @file-catalog-pair blocks from handbook content.
func extractFileCatalogPairEntries(content string) []FileCatalogPairEntry {
	var entries []FileCatalogPairEntry

	lines := strings.Split(content, "\n")

	var current *FileCatalogPairEntry

	state := pairScanStateNone

	var sectionLines []string

	for i, line := range lines {
		lineNum := i + 1

		trimmed := strings.TrimSpace(line)

		// Open a new pair entry.
		if current == nil {
			if match := fileCatalogPairRegex.FindStringSubmatch(line); len(match) == cryptoutilSharedMagic.FileCatalogPairMatchGroups {
				current = &FileCatalogPairEntry{
					CopilotPath: match[1],
					ClaudePath:  match[2],
					LineNumber:  lineNum,
				}
				state = pairScanStateNone
				sectionLines = nil
			}

			continue
		}

		// Close the pair entry.
		if trimmed == cryptoutilSharedMagic.CICDFileCatalogPairEndMarker {
			entries = append(entries, *current)
			current = nil
			state = pairScanStateNone
			sectionLines = nil

			continue
		}

		// Sub-section transitions.
		switch trimmed {
		case cryptoutilSharedMagic.CICDFileCatalogCopilotFMStartMarker:
			state = pairScanStateCopilotFM
			sectionLines = nil

			continue

		case cryptoutilSharedMagic.CICDFileCatalogCopilotFMEndMarker:
			current.CopilotFrontmatter = extractPairSection(sectionLines)
			state = pairScanStateNone
			sectionLines = nil

			continue

		case cryptoutilSharedMagic.CICDFileCatalogClaudeFMStartMarker:
			state = pairScanStateClaudeFM
			sectionLines = nil

			continue

		case cryptoutilSharedMagic.CICDFileCatalogClaudeFMEndMarker:
			current.ClaudeFrontmatter = extractPairSection(sectionLines)
			state = pairScanStateNone
			sectionLines = nil

			continue

		case cryptoutilSharedMagic.CICDFileCatalogBodyStartMarker:
			state = pairScanStateBody
			sectionLines = nil

			continue

		case cryptoutilSharedMagic.CICDFileCatalogBodyEndMarker:
			current.Body = extractPairSection(sectionLines)
			state = pairScanStateNone
			sectionLines = nil

			continue
		}

		// Accumulate lines inside a sub-section.
		if state != pairScanStateNone {
			sectionLines = append(sectionLines, line)
		}
	}

	return entries
}

// extractPairSection joins lines and adds a trailing newline.
// No leading-newline stripping — body sections may legitimately start with a blank line
// (e.g., the blank line between the YAML closing --- and the first heading).
func extractPairSection(lines []string) string {
	return strings.Join(lines, "\n") + "\n"
}

// reconstructSingleFile returns the expected on-disk content for a single-file catalog entry.
// Content is already the complete verbatim file content; we just normalize line endings.
func reconstructSingleFile(entry FileCatalogEntry) string {
	return strings.ReplaceAll(entry.Content, "\r\n", "\n")
}

// reconstructPairFile reconstructs file content from frontmatter + body.
// The frontmatter includes the --- delimiters; body is everything after the closing ---.
func reconstructPairFile(frontmatter, body string) string {
	// Both sections were stored with their own trailing newline.
	// Ensure the join produces: "---\nYAML\n---\nBODY"
	fm := strings.ReplaceAll(frontmatter, "\r\n", "\n")
	bd := strings.ReplaceAll(body, "\r\n", "\n")

	return fm + bd
}

// CheckCatalogFiles verifies that every @file-catalog and @file-catalog-pair entry in
// ENG-HANDBOOK.md matches the corresponding file(s) on disk exactly (LF-normalized).
func CheckCatalogFiles(rootDir string, readFileFn func(string) ([]byte, error)) (*CatalogFilesResult, error) {
	result := &CatalogFilesResult{}

	handbookContent, err := readFileFn("docs/ENG-HANDBOOK.md")
	if err != nil {
		return nil, fmt.Errorf("failed to read docs/ENG-HANDBOOK.md: %w", err)
	}

	normalizedHandbook := strings.ReplaceAll(string(handbookContent), "\r\n", "\n")

	singles := extractFileCatalogEntries(normalizedHandbook)
	pairs := extractFileCatalogPairEntries(normalizedHandbook)

	// Check single-file entries.
	for _, entry := range singles {
		result.Checked++

		diskBytes, readErr := readFileFn(entry.Path)
		if readErr != nil {
			result.Violations = append(result.Violations, CatalogViolation{
				File:   entry.Path,
				Field:  "disk-read",
				Detail: fmt.Sprintf("cannot read file from disk: %v", readErr),
			})

			continue
		}

		diskContent := strings.ReplaceAll(string(diskBytes), "\r\n", "\n")
		catalogContent := reconstructSingleFile(entry)

		if diskContent != catalogContent {
			result.Violations = append(result.Violations, CatalogViolation{
				File:   entry.Path,
				Field:  "content-mismatch",
				Detail: fmt.Sprintf("catalog content at handbook line %d does not match file on disk", entry.LineNumber),
			})
		}
	}

	// Check pair entries.
	for _, pair := range pairs {
		result.Checked += catalogPairFilesCount // 2 files per pair

		// Copilot file.
		copilotExpected := reconstructPairFile(pair.CopilotFrontmatter, pair.Body)

		diskCopilot, readCopilotErr := readFileFn(pair.CopilotPath)
		if readCopilotErr != nil {
			result.Violations = append(result.Violations, CatalogViolation{
				File:   pair.CopilotPath,
				Field:  "disk-read",
				Detail: fmt.Sprintf("cannot read Copilot file from disk: %v", readCopilotErr),
			})
		} else {
			diskCopilotStr := strings.ReplaceAll(string(diskCopilot), "\r\n", "\n")
			if diskCopilotStr != copilotExpected {
				result.Violations = append(result.Violations, CatalogViolation{
					File:   pair.CopilotPath,
					Field:  "content-mismatch",
					Detail: fmt.Sprintf("catalog Copilot content at handbook line %d does not match file on disk", pair.LineNumber),
				})
			}
		}

		// Claude file.
		claudeExpected := reconstructPairFile(pair.ClaudeFrontmatter, pair.Body)

		diskClaude, readClaudeErr := readFileFn(pair.ClaudePath)
		if readClaudeErr != nil {
			result.Violations = append(result.Violations, CatalogViolation{
				File:   pair.ClaudePath,
				Field:  "disk-read",
				Detail: fmt.Sprintf("cannot read Claude file from disk: %v", readClaudeErr),
			})
		} else {
			diskClaudeStr := strings.ReplaceAll(string(diskClaude), "\r\n", "\n")
			if diskClaudeStr != claudeExpected {
				result.Violations = append(result.Violations, CatalogViolation{
					File:   pair.ClaudePath,
					Field:  "content-mismatch",
					Detail: fmt.Sprintf("catalog Claude content at handbook line %d does not match file on disk", pair.LineNumber),
				})
			}
		}
	}

	return result, nil
}

// catalogPairFilesCount is the number of files represented by each @file-catalog-pair entry.
const catalogPairFilesCount = 2

// CheckCatalogPropagation verifies that every @appendix-propagate chunk whose target file
// has a @file-catalog or @file-catalog-pair entry in ENG-HANDBOOK.md also has a matching
// @source block inside that catalog entry's body with identical content.
func CheckCatalogPropagation(rootDir string, readFileFn func(string) ([]byte, error)) (*CatalogPropagationResult, error) {
	result := &CatalogPropagationResult{}

	handbookContent, err := readFileFn("docs/ENG-HANDBOOK.md")
	if err != nil {
		return nil, fmt.Errorf("failed to read docs/ENG-HANDBOOK.md: %w", err)
	}

	normalizedHandbook := strings.ReplaceAll(string(handbookContent), "\r\n", "\n")

	// Extract all @appendix-propagate blocks.
	propagateBlocks := extractPropagateBlocks(normalizedHandbook)

	// Build a map from file path to catalog body content.
	// For single-file entries the "body" is the full content.
	// For pair entries the body is the shared body section.
	catalogBodies := buildCatalogBodies(normalizedHandbook)

	// For each appendix-propagate block, split the comma-separated target files list and
	// check each file that appears in the catalog.
	for _, pb := range propagateBlocks {
		// Only check blocks that are appendix-propagate (not plain @propagate).
		if pb.AppendixID == "" {
			continue
		}

		// TargetFile may be a comma-separated list of file paths (e.g., multiple files share a chunk).
		rawTargets := strings.Split(pb.TargetFile, ",")

		for _, rawTarget := range rawTargets {
			targetFile := strings.TrimSpace(rawTarget)
			if targetFile == "" {
				continue
			}

			body, hasCatalog := catalogBodies[targetFile]
			if !hasCatalog {
				// Target file not in catalog — not our concern for this linter.
				continue
			}

			result.Checked++

			// Extract @source blocks from the catalog body.
			sourceBlocks := extractSourceBlocks(body)

			// Find a matching @source block by chunk ID.
			var found *SourceBlock

			for idx := range sourceBlocks {
				if sourceBlocks[idx].ChunkID == pb.ChunkID {
					sb := sourceBlocks[idx]
					found = &sb

					break
				}
			}

			if found == nil {
				result.Violations = append(result.Violations, CatalogViolation{
					File:   targetFile,
					Field:  "source-missing",
					Detail: fmt.Sprintf("chunk %q: no @source block found in catalog body for file %q", pb.ChunkID, targetFile),
				})

				continue
			}

			if found.Content != pb.Content {
				result.Violations = append(result.Violations, CatalogViolation{
					File:   targetFile,
					Field:  "source-mismatch",
					Detail: fmt.Sprintf("chunk %q: catalog @source content differs from @appendix-propagate content in file %q", pb.ChunkID, targetFile),
				})
			}
		}
	}

	return result, nil
}

// buildCatalogBodies builds a map from file path to the searchable body content
// (for single-file entries: full content; for pairs: shared body only).
func buildCatalogBodies(handbookContent string) map[string]string {
	bodies := make(map[string]string)

	for _, entry := range extractFileCatalogEntries(handbookContent) {
		bodies[entry.Path] = entry.Content
	}

	for _, pair := range extractFileCatalogPairEntries(handbookContent) {
		// Pair chunks are in the shared body.
		bodies[pair.CopilotPath] = pair.Body
		bodies[pair.ClaudePath] = pair.Body
	}

	return bodies
}
