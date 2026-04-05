// Copyright (c) 2025 Justin Cranford

// Package lint_architecture_links validates that all ENG-HANDBOOK.md section anchors
// referenced in instruction, agent, and skill files resolve to real headings in
// docs/ENG-HANDBOOK.md. Prevents broken cross-references that lead to dead links.
// Implements the agent self-containment requirement from ENG-HANDBOOK.md §2.1.1.
package lint_architecture_links

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// architectureAnchorRegex matches ENG-HANDBOOK.md#anchor-name patterns in markdown.
// Includes underscores because GitHub's anchor algorithm preserves them (e.g., format_go).
var architectureAnchorRegex = regexp.MustCompile(`ENG-HANDBOOK\.md#([a-z0-9_][a-z0-9_-]*)`)

// headingRegex matches H1–H4 markdown headings.
var headingRegex = regexp.MustCompile(`^(#{1,4}) (.+)`)

// targetRelDirs lists directories (relative to project root) containing files to validate.
var targetRelDirs = []string{
	cryptoutilSharedMagic.CICDGithubInstructionsDir,
	cryptoutilSharedMagic.CICDGithubAgentsDir,
	cryptoutilSharedMagic.CICDGithubSkillsDir,
}

// excludedRelPaths lists path components (relative to targetRelDirs) to skip.
// Scaffold templates use placeholder anchors like #xy-anchor that are never real.
var excludedSkillDirNames = map[string]struct{}{
	"instruction-scaffold": {},
	"skill-scaffold":       {},
	"agent-scaffold":       {},
}

// architectureMdRelPath is the path to ENG-HANDBOOK.md relative to the project root.
const architectureMdRelPath = "docs/ENG-HANDBOOK.md"

// Check runs the lint-architecture-links check, automatically finding the project root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return checkWithFS(logger, os.Getwd, filepath.WalkDir, os.ReadFile)
}

// checkWithFS runs the check using injected filesystem operations for testability.
func checkWithFS(
	logger *cryptoutilCmdCicdCommon.Logger,
	getwdFn func() (string, error),
	walkFn func(root string, fn fs.WalkDirFunc) error,
	readFileFn func(name string) ([]byte, error),
) error {
	root, err := findProjectRoot(getwdFn)
	if err != nil {
		return fmt.Errorf("lint-architecture-links: %w", err)
	}

	archPath := filepath.Join(root, filepath.FromSlash(architectureMdRelPath))

	archContent, err := readFileFn(archPath)
	if err != nil {
		return fmt.Errorf("lint-architecture-links: failed to read ENG-HANDBOOK.md: %w", err)
	}

	anchors := extractAnchors(string(archContent))
	logger.Log(fmt.Sprintf("Extracted %d anchors from ENG-HANDBOOK.md", len(anchors)))

	var violations []string

	for _, relDir := range targetRelDirs {
		absDir := filepath.Join(root, filepath.FromSlash(relDir))

		dirViolations, walkErr := checkDir(logger, absDir, relDir, anchors, walkFn, readFileFn)
		if walkErr != nil {
			return fmt.Errorf("lint-architecture-links: %w", walkErr)
		}

		violations = append(violations, dirViolations...)
	}

	if len(violations) > 0 {
		for _, v := range violations {
			logger.Log(fmt.Sprintf("  broken anchor: %s", v))
		}

		return fmt.Errorf("lint-architecture-links: %d broken ENG-HANDBOOK.md anchor(s) found", len(violations))
	}

	logger.Log("lint-architecture-links: all ENG-HANDBOOK.md anchors are valid")

	return nil
}

// checkDir walks a directory and validates all ENG-HANDBOOK.md anchor references in .md files.
func checkDir(
	logger *cryptoutilCmdCicdCommon.Logger,
	absDir, relDir string,
	anchors map[string]struct{},
	walkFn func(root string, fn fs.WalkDirFunc) error,
	readFileFn func(name string) ([]byte, error),
) ([]string, error) {
	var violations []string

	err := walkFn(absDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("failed to access %s: %w", path, walkErr)
		}

		// Skip excluded scaffold directories (they contain placeholder anchors).
		if d.IsDir() {
			if _, excluded := excludedSkillDirNames[d.Name()]; excluded {
				return fs.SkipDir
			}

			return nil
		}

		if !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}

		content, readErr := readFileFn(path)
		if readErr != nil {
			return fmt.Errorf("failed to read %s: %w", path, readErr)
		}

		refs := architectureAnchorRegex.FindAllStringSubmatch(string(content), -1)
		for _, ref := range refs {
			anchor := ref[1]
			if _, ok := anchors[anchor]; !ok {
				rel := strings.TrimPrefix(path, absDir)
				rel = strings.TrimPrefix(filepath.Join(relDir, rel), string(filepath.Separator))
				violations = append(violations, fmt.Sprintf("%s: #%s", rel, anchor))

				logger.Log(fmt.Sprintf("  broken anchor in %s: #%s", rel, anchor))
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk %s: %w", relDir, err)
	}

	return violations, nil
}

// extractAnchors parses ENG-HANDBOOK.md content and returns all heading anchors as a set.
// The anchor format follows GitHub's heading-to-anchor conversion algorithm.
func extractAnchors(content string) map[string]struct{} {
	anchors := make(map[string]struct{})

	for _, line := range strings.Split(content, "\n") {
		matches := headingRegex.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		anchor := headingToAnchor(matches[2])
		if anchor != "" {
			anchors[anchor] = struct{}{}
		}
	}

	return anchors
}

// headingToAnchor converts a markdown heading text to a GitHub-compatible anchor.
// Algorithm: lowercase, keep only alphanumeric + hyphen, replace spaces with hyphens.
func headingToAnchor(text string) string {
	var sb strings.Builder

	for _, r := range strings.ToLower(text) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '_':
			sb.WriteRune(r)
		case r == ' ':
			sb.WriteRune('-')
		}
	}

	return sb.String()
}

// findProjectRoot walks up from cwd to find the directory containing go.mod.
func findProjectRoot(getwdFn func() (string, error)) (string, error) {
	dir, err := getwdFn()
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
