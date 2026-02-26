// Copyright (c) 2025 Justin Cranford

package cicd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// PropagationRef represents a reference from an instruction/agent file to ARCHITECTURE.md.
type PropagationRef struct {
	SourceFile string // e.g., ".github/instructions/02-01.architecture.instructions.md"
	LineNumber int    // 1-based line number in SourceFile
	Anchor     string // e.g., "941-otel-collector-processor-constraints"
	RawRef     string // e.g., "See [ARCHITECTURE.md Section 9.4.1 ...](...)"
}

// LevelCoverage tracks section coverage at a specific heading level.
type LevelCoverage struct {
	Total      int // Total sections at this level in ARCHITECTURE.md.
	Referenced int // Sections at this level referenced by instruction/agent files.
}

// PropagationResult holds validation results.
type PropagationResult struct {
	ValidRefs    []PropagationRef
	BrokenRefs   []PropagationRef
	OrphanedKeys []string // ARCHITECTURE.md anchors with zero references
	TotalAnchors int
	HighImpact   LevelCoverage // ## sections.
	MediumImpact LevelCoverage // ### sections.
	LowImpact    LevelCoverage // #### sections.
}

// anchorRegex matches ARCHITECTURE.md#anchor-fragment in markdown links.
var anchorRegex = regexp.MustCompile(`ARCHITECTURE\.md#([a-z0-9_-]+)\)`)

// headerToAnchor converts a Markdown header text to a GitHub-flavored anchor.
// Rules: lowercase, spaces to hyphens, remove non-alphanumeric except hyphens/underscores,
// strip trailing hyphens. Consecutive hyphens are preserved (GitHub behavior).
func headerToAnchor(header string) string {
	// Strip leading # and spaces.
	text := strings.TrimLeft(header, "# ")

	// Remove emoji (common in ARCHITECTURE.md section headers).
	var sb strings.Builder

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' || r == '-' || r == '_' || r == '.' || r == '/' || r == '(' || r == ')' || r == '&' || r == '\'' {
			sb.WriteRune(r)
		}
	}

	text = sb.String()

	// Lowercase.
	text = strings.ToLower(text)

	// Replace spaces with hyphens.
	text = strings.ReplaceAll(text, " ", "-")

	// Remove characters that aren't alphanumeric, hyphens, or underscores.
	cleaned := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}

		return -1
	}, text)

	// NOTE: Do NOT collapse consecutive hyphens — GitHub Flavored Markdown preserves them.

	// Strip leading and trailing hyphens.
	cleaned = strings.Trim(cleaned, "-")

	return cleaned
}

// extractAnchorsFromArchitecture reads ARCHITECTURE.md and returns a set of valid anchors.
func extractAnchorsFromArchitecture(content string) map[string]bool {
	anchors := make(map[string]bool)

	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "#") {
			anchor := headerToAnchor(line)
			if anchor != "" {
				anchors[anchor] = true
			}
		}
	}

	return anchors
}

// extractRefsFromFile reads a file and extracts all ARCHITECTURE.md anchor references.
func extractRefsFromFile(relPath, content string) []PropagationRef {
	var refs []PropagationRef

	lines := strings.Split(content, "\n")

	for i, line := range lines {
		matches := anchorRegex.FindAllStringSubmatch(line, -1)

		for _, match := range matches {
			if len(match) >= 2 {
				refs = append(refs, PropagationRef{
					SourceFile: relPath,
					LineNumber: i + 1,
					Anchor:     match[1],
					RawRef:     truncateRef(line),
				})
			}
		}
	}

	return refs
}

// truncateRef truncates a line for display.
func truncateRef(line string) string {
	const maxRefLength = 120

	line = strings.TrimSpace(line)

	if len(line) > maxRefLength {
		return line[:maxRefLength] + "..."
	}

	return line
}

// ValidatePropagation performs the full validation.
func ValidatePropagation(rootDir string, readFile func(string) ([]byte, error)) (*PropagationResult, error) {
	// Read ARCHITECTURE.md.
	archContent, err := readFile("docs/ARCHITECTURE.md")
	if err != nil {
		return nil, fmt.Errorf("failed to read docs/ARCHITECTURE.md: %w", err)
	}

	anchors := extractAnchorsFromArchitecture(string(archContent))

	// Scan instruction and agent files.
	var allRefs []PropagationRef

	scanDirs := []struct {
		dir     string
		pattern string
	}{
		{dir: ".github/instructions", pattern: "*.instructions.md"},
		{dir: ".github/agents", pattern: "*.agent.md"},
	}

	// Also scan copilot-instructions.md directly.
	copilotContent, err := readFile(".github/copilot-instructions.md")
	if err == nil {
		allRefs = append(allRefs, extractRefsFromFile(".github/copilot-instructions.md", string(copilotContent))...)
	}

	for _, sd := range scanDirs {
		dirPath := filepath.Join(rootDir, sd.dir)

		entries, dirErr := os.ReadDir(dirPath)
		if dirErr != nil {
			continue // Directory may not exist.
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			matched, matchErr := filepath.Match(sd.pattern, entry.Name())
			if matchErr != nil || !matched {
				continue
			}

			relPath := filepath.Join(sd.dir, entry.Name())

			content, readErr := readFile(relPath)
			if readErr != nil {
				continue
			}

			allRefs = append(allRefs, extractRefsFromFile(relPath, string(content))...)
		}
	}

	// Classify refs as valid or broken.
	result := &PropagationResult{
		TotalAnchors: len(anchors),
	}

	referencedAnchors := make(map[string]bool)

	for _, ref := range allRefs {
		if anchors[ref.Anchor] {
			result.ValidRefs = append(result.ValidRefs, ref)
			referencedAnchors[ref.Anchor] = true
		} else {
			result.BrokenRefs = append(result.BrokenRefs, ref)
		}
	}

	// Find orphaned anchors and compute per-level coverage statistics.
	archLines := strings.Split(string(archContent), "\n")

	for _, line := range archLines {
		// Determine heading level.
		switch {
		case strings.HasPrefix(line, "#### "):
			anchor := headerToAnchor(line)
			if anchor != "" {
				result.LowImpact.Total++

				if referencedAnchors[anchor] {
					result.LowImpact.Referenced++
				}
			}
		case strings.HasPrefix(line, "### "):
			anchor := headerToAnchor(line)
			if anchor != "" {
				result.MediumImpact.Total++

				if referencedAnchors[anchor] {
					result.MediumImpact.Referenced++
				} else {
					result.OrphanedKeys = append(result.OrphanedKeys, anchor)
				}
			}
		case strings.HasPrefix(line, "## "):
			anchor := headerToAnchor(line)
			if anchor != "" {
				result.HighImpact.Total++

				if referencedAnchors[anchor] {
					result.HighImpact.Referenced++
				} else {
					result.OrphanedKeys = append(result.OrphanedKeys, anchor)
				}
			}
		}
	}

	sort.Strings(result.OrphanedKeys)

	return result, nil
}

// formatLevelCoverage formats a single line of level coverage output.
func formatLevelCoverage(label string, lc LevelCoverage) string {
	if lc.Total == 0 {
		return fmt.Sprintf("%s: 0/0 (N/A)\n", label)
	}

	const percentMultiplier = 100

	pct := percentMultiplier * lc.Referenced / lc.Total

	return fmt.Sprintf("%s: %d/%d (%d%%)\n", label, lc.Referenced, lc.Total, pct)
}

// FormatPropagationResults formats validation results for display.
func FormatPropagationResults(result *PropagationResult) string {
	var sb strings.Builder

	sb.WriteString("=== ARCHITECTURE.md Propagation Validation ===\n\n")

	// Broken refs.
	if len(result.BrokenRefs) > 0 {
		sb.WriteString(fmt.Sprintf("BROKEN REFERENCES (%d):\n", len(result.BrokenRefs)))

		for _, ref := range result.BrokenRefs {
			sb.WriteString(fmt.Sprintf("  FAIL %s:%d -> #%s\n", ref.SourceFile, ref.LineNumber, ref.Anchor))
		}

		sb.WriteString("\n")
	}

	// Orphaned sections.
	if len(result.OrphanedKeys) > 0 {
		sb.WriteString(fmt.Sprintf("ORPHANED SECTIONS (%d of %d, ## and ### level):\n", len(result.OrphanedKeys), result.TotalAnchors))

		for _, anchor := range result.OrphanedKeys {
			sb.WriteString(fmt.Sprintf("  WARN #%s\n", anchor))
		}

		sb.WriteString("\n")
	}

	// Summary.
	referencedCount := len(result.ValidRefs)
	brokenCount := len(result.BrokenRefs)
	orphanedCount := len(result.OrphanedKeys)

	// Section coverage by impact level.
	sb.WriteString("SECTION COVERAGE:\n")
	sb.WriteString(formatLevelCoverage("  High   (##  )", result.HighImpact))
	sb.WriteString(formatLevelCoverage("  Medium (### )", result.MediumImpact))
	sb.WriteString(formatLevelCoverage("  Low    (####)", result.LowImpact))

	combinedTotal := result.HighImpact.Total + result.MediumImpact.Total
	combinedReferenced := result.HighImpact.Referenced + result.MediumImpact.Referenced

	sb.WriteString(formatLevelCoverage("  Combined ##/###", LevelCoverage{Total: combinedTotal, Referenced: combinedReferenced}))
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("=== Summary: %d valid refs, %d broken refs, %d orphaned sections ===\n", referencedCount, brokenCount, orphanedCount))

	if brokenCount == 0 {
		sb.WriteString("All references resolve to valid ARCHITECTURE.md sections.\n")
	} else {
		sb.WriteString("Propagation validation FAILED. Fix broken references.\n")
	}

	return sb.String()
}

// ValidatePropagationCommand is the CLI entry point for validate-propagation.
// Returns exit code: 0 if no broken refs, 1 if broken refs found.
func ValidatePropagationCommand(stdout, stderr io.Writer) int {
	rootDir, err := findProjectRoot()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error: %s\n", err)

		return 1
	}

	return validatePropagationWithRoot(rootDir, stdout, stderr)
}

// validatePropagationWithRoot validates propagation using a specified root directory.
func validatePropagationWithRoot(rootDir string, stdout, stderr io.Writer) int {
	result, err := ValidatePropagation(rootDir, rootedReadFile(rootDir))
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error: %s\n", err)

		return 1
	}

	report := FormatPropagationResults(result)
	_, _ = fmt.Fprint(stdout, report)

	if len(result.BrokenRefs) > 0 {
		return 1
	}

	return 0
}
