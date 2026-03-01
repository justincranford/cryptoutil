// Copyright (c) 2025 Justin Cranford

package docs_validation

import (
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// PropagateBlock represents a @propagate block extracted from ARCHITECTURE.md.
type PropagateBlock struct {
	TargetFile string // e.g., ".github/instructions/02-06.authn.instructions.md"
	ChunkID    string // e.g., "key-principles"
	Content    string // verbatim body text between markers
	LineNumber int    // 1-based line number of the @propagate marker
}

// SourceBlock represents a @source block extracted from an instruction/agent file.
type SourceBlock struct {
	SourceFile string // e.g., ".github/instructions/02-06.authn.instructions.md"
	ChunkID    string // e.g., "key-principles"
	Content    string // verbatim body text between markers
	LineNumber int    // 1-based line number of the @source marker
}

// ChunkResult represents the validation result for a single chunk.
type ChunkResult struct {
	ChunkID        string
	PropagateBlock PropagateBlock
	SourceBlock    *SourceBlock // nil if missing.
	Status         ChunkStatus
}

// ChunkStatus is the validation status of a chunk.
type ChunkStatus int

const (
	// ChunkStatusMatch means propagate and source content are identical.
	ChunkStatusMatch ChunkStatus = iota
	// ChunkStatusMismatch means content differs between propagate and source.
	ChunkStatusMismatch
	// ChunkStatusMissing means the @source block was not found in the target file.
	ChunkStatusMissing
	// ChunkStatusFileNotFound means the target file does not exist.
	ChunkStatusFileNotFound
)

// ChunkValidationResult holds the overall chunk validation results.
type ChunkValidationResult struct {
	Results    []ChunkResult
	Matched    int
	Mismatched int
	Missing    int
	FileErrors int
}

// propagateRegex matches the @propagate opening marker.
var propagateRegex = regexp.MustCompile(`^<!-- @propagate to="([^"]+)" as="([^"]+)" -->$`)

// sourceRegex matches the @source opening marker.
var sourceRegex = regexp.MustCompile(`<!-- @source from="([^"]+)" as="([^"]+)" -->`)

// extractPropagateBlocks extracts all @propagate blocks from ARCHITECTURE.md content,
// skipping markers inside code fences but preserving code fences within propagated content.
func extractPropagateBlocks(content string) []PropagateBlock {
	var blocks []PropagateBlock

	lines := strings.Split(content, "\n")
	inCodeFence := false

	var current *PropagateBlock

	var contentLines []string

	for i, line := range lines {
		lineNum := i + 1

		// Track code fences only outside propagate blocks.
		// Propagated content may contain its own code fences.
		if current == nil && strings.HasPrefix(line, "```") {
			inCodeFence = !inCodeFence

			continue
		}

		if inCodeFence {
			continue
		}

		// Check for @propagate start (only when not already in a block).
		if current == nil {
			if match := propagateRegex.FindStringSubmatch(line); len(match) == cryptoutilSharedMagic.PropagateMarkerMatchGroups {
				current = &PropagateBlock{
					TargetFile: match[1],
					ChunkID:    match[2],
					LineNumber: lineNum,
				}
				contentLines = nil

				continue
			}

			continue
		}

		// Check for @/propagate end.
		if strings.Contains(line, "<!-- @/propagate -->") {
			current.Content = strings.Join(contentLines, "\n") + "\n"
			blocks = append(blocks, *current)
			current = nil
			contentLines = nil

			continue
		}

		// Accumulate content inside propagate block.
		contentLines = append(contentLines, line)
	}

	return blocks
}

// extractSourceBlocks extracts all @source blocks from a file's content.
func extractSourceBlocks(content string) []SourceBlock {
	var blocks []SourceBlock

	lines := strings.Split(content, "\n")

	var current *SourceBlock

	var contentLines []string

	for i, line := range lines {
		lineNum := i + 1

		// Check for @source start.
		if match := sourceRegex.FindStringSubmatch(line); len(match) == cryptoutilSharedMagic.PropagateMarkerMatchGroups {
			current = &SourceBlock{
				ChunkID:    match[2],
				LineNumber: lineNum,
			}
			contentLines = nil

			continue
		}

		// Check for @/source end.
		if strings.Contains(line, "<!-- @/source -->") && current != nil {
			current.Content = strings.Join(contentLines, "\n") + "\n"
			blocks = append(blocks, *current)
			current = nil
			contentLines = nil

			continue
		}

		// Accumulate content.
		if current != nil {
			contentLines = append(contentLines, line)
		}
	}

	return blocks
}

// ValidateChunks compares all @propagate blocks against their @source counterparts.
func ValidateChunks(rootDir string, readFile func(string) ([]byte, error)) (*ChunkValidationResult, error) {
	archContent, err := readFile("docs/ARCHITECTURE.md")
	if err != nil {
		return nil, fmt.Errorf("failed to read docs/ARCHITECTURE.md: %w", err)
	}

	propagateBlocks := extractPropagateBlocks(string(archContent))

	result := &ChunkValidationResult{}

	// Group source blocks by file to avoid reading files multiple times.
	fileSourceBlocks := make(map[string][]SourceBlock)
	failedFiles := make(map[string]bool)

	for _, pb := range propagateBlocks {
		if _, loaded := fileSourceBlocks[pb.TargetFile]; !loaded && !failedFiles[pb.TargetFile] {
			targetContent, readErr := readFile(pb.TargetFile)
			if readErr != nil {
				failedFiles[pb.TargetFile] = true

				result.Results = append(result.Results, ChunkResult{
					ChunkID:        pb.ChunkID,
					PropagateBlock: pb,
					Status:         ChunkStatusFileNotFound,
				})
				result.FileErrors++

				continue
			}

			sourceBlocks := extractSourceBlocks(string(targetContent))
			for idx := range sourceBlocks {
				sourceBlocks[idx].SourceFile = pb.TargetFile
			}

			fileSourceBlocks[pb.TargetFile] = sourceBlocks
		}

		// Check if file had a read error.
		if failedFiles[pb.TargetFile] {
			result.Results = append(result.Results, ChunkResult{
				ChunkID:        pb.ChunkID,
				PropagateBlock: pb,
				Status:         ChunkStatusFileNotFound,
			})
			result.FileErrors++

			continue
		}

		// Find matching @source block.
		sourceBlocks := fileSourceBlocks[pb.TargetFile]

		var matched *SourceBlock

		for idx := range sourceBlocks {
			if sourceBlocks[idx].ChunkID == pb.ChunkID {
				sb := sourceBlocks[idx]
				matched = &sb

				break
			}
		}

		if matched == nil {
			result.Results = append(result.Results, ChunkResult{
				ChunkID:        pb.ChunkID,
				PropagateBlock: pb,
				Status:         ChunkStatusMissing,
			})
			result.Missing++

			continue
		}

		if pb.Content == matched.Content {
			result.Results = append(result.Results, ChunkResult{
				ChunkID:        pb.ChunkID,
				PropagateBlock: pb,
				SourceBlock:    matched,
				Status:         ChunkStatusMatch,
			})
			result.Matched++
		} else {
			result.Results = append(result.Results, ChunkResult{
				ChunkID:        pb.ChunkID,
				PropagateBlock: pb,
				SourceBlock:    matched,
				Status:         ChunkStatusMismatch,
			})
			result.Mismatched++
		}
	}

	return result, nil
}

// FormatChunkResults formats chunk validation results for display.
func FormatChunkResults(result *ChunkValidationResult) string {
	var sb strings.Builder

	sb.WriteString("=== ARCHITECTURE.md Chunk Verification ===\n\n")

	// Group issues by type.
	var mismatches, missing, fileErrors []ChunkResult

	for _, cr := range result.Results {
		switch cr.Status {
		case ChunkStatusMatch:
			// Matched chunks need no special handling.
		case ChunkStatusMismatch:
			mismatches = append(mismatches, cr)
		case ChunkStatusMissing:
			missing = append(missing, cr)
		case ChunkStatusFileNotFound:
			fileErrors = append(fileErrors, cr)
		}
	}

	if len(fileErrors) > 0 {
		sb.WriteString(fmt.Sprintf("FILE NOT FOUND (%d):\n", len(fileErrors)))

		sort.Slice(fileErrors, func(i, j int) bool { return fileErrors[i].ChunkID < fileErrors[j].ChunkID })

		for _, cr := range fileErrors {
			sb.WriteString(fmt.Sprintf("  FAIL [%s] target=%s (line %d)\n", cr.ChunkID, cr.PropagateBlock.TargetFile, cr.PropagateBlock.LineNumber))
		}

		sb.WriteString("\n")
	}

	if len(missing) > 0 {
		sb.WriteString(fmt.Sprintf("MISSING @source BLOCKS (%d):\n", len(missing)))

		sort.Slice(missing, func(i, j int) bool { return missing[i].ChunkID < missing[j].ChunkID })

		for _, cr := range missing {
			sb.WriteString(fmt.Sprintf("  FAIL [%s] not found in %s (propagate at line %d)\n", cr.ChunkID, cr.PropagateBlock.TargetFile, cr.PropagateBlock.LineNumber))
		}

		sb.WriteString("\n")
	}

	if len(mismatches) > 0 {
		sb.WriteString(fmt.Sprintf("CONTENT MISMATCHES (%d):\n", len(mismatches)))

		sort.Slice(mismatches, func(i, j int) bool { return mismatches[i].ChunkID < mismatches[j].ChunkID })

		for _, cr := range mismatches {
			sb.WriteString(fmt.Sprintf("  STALE [%s] in %s (propagate line %d, source line %d)\n",
				cr.ChunkID, cr.PropagateBlock.TargetFile, cr.PropagateBlock.LineNumber, cr.SourceBlock.LineNumber))
		}

		sb.WriteString("\n")
	}

	total := len(result.Results)
	sb.WriteString(fmt.Sprintf("=== Summary: %d chunks, %d matched, %d mismatched, %d missing, %d file errors ===\n",
		total, result.Matched, result.Mismatched, result.Missing, result.FileErrors))

	if result.Mismatched == 0 && result.Missing == 0 && result.FileErrors == 0 {
		sb.WriteString("All propagated chunks are in sync.\n")
	} else {
		sb.WriteString("Chunk verification FAILED. Fix stale or missing propagation.\n")
	}

	return sb.String()
}

// ValidateChunksCommand is the CLI entry point for validate-chunks.
// Returns exit code: 0 if all chunks match, 1 if any issues found.
func ValidateChunksCommand(stdout, stderr io.Writer) int {
	rootDir, err := findProjectRoot()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error: %s\n", err)

		return 1
	}

	return validateChunksWithRoot(rootDir, stdout, stderr)
}

// validateChunksWithRoot validates chunks using a specified root directory.
func validateChunksWithRoot(rootDir string, stdout, stderr io.Writer) int {
	result, err := ValidateChunks(rootDir, rootedReadFile(rootDir))
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error: %s\n", err)

		return 1
	}

	report := FormatChunkResults(result)
	_, _ = fmt.Fprint(stdout, report)

	if result.Mismatched > 0 || result.Missing > 0 || result.FileErrors > 0 {
		return 1
	}

	return 0
}
