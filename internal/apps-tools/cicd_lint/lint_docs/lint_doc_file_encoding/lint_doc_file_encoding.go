// Copyright (c) 2025-2026 Justin Cranford.
// Package lint_doc_file_encoding enforces UTF-8 (without BOM) and LF-only
// line endings for documentation policy artifacts.
package lint_doc_file_encoding

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
)

const (
	handbookPath = "docs/ENG-HANDBOOK.md"
)

var nullByte byte

var policyGlobPatterns = []string{
	".github/instructions/*.instructions.md",
	".github/agents/*.agent.md",
	".claude/agents/*.md",
	".github/skills/*/SKILL.md",
	".claude/skills/*/SKILL.md",
}

// Check enforces UTF-8 without BOM and LF-only line endings for handbook,
// instructions, agents, and skills.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	rootDir, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("find project root: %w", err)
	}

	return CheckInDir(logger, rootDir)
}

// CheckInDir enforces UTF-8 without BOM and LF-only line endings for handbook,
// instructions, agents, and skills in rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking handbook/instruction/agent/skill file encoding and line endings...")

	files, err := collectPolicyFiles(rootDir)
	if err != nil {
		return fmt.Errorf("collect policy files: %w", err)
	}

	var violations []string

	for _, filePath := range files {
		issues, err := checkFile(filePath)
		if err != nil {
			violations = append(violations, fmt.Sprintf("%s: %v", filePath, err))

			continue
		}

		for _, issue := range issues {
			violations = append(violations, fmt.Sprintf("%s: %s", filePath, issue))
		}
	}

	if len(violations) > 0 {
		sort.Strings(violations)

		return fmt.Errorf("doc encoding policy violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("lint-doc-file-encoding: all policy artifacts are UTF-8 without BOM and LF-only")

	return nil
}

func collectPolicyFiles(rootDir string) ([]string, error) {
	files := []string{filepath.Join(rootDir, handbookPath)}

	for _, pattern := range policyGlobPatterns {
		fullPattern := filepath.Join(rootDir, pattern)

		matches, err := filepath.Glob(fullPattern)
		if err != nil {
			return nil, fmt.Errorf("glob %q: %w", pattern, err)
		}

		files = append(files, matches...)
	}

	seen := make(map[string]struct{}, len(files))

	unique := make([]string, 0, len(files))
	for _, f := range files {
		clean := filepath.Clean(f)
		if _, ok := seen[clean]; ok {
			continue
		}

		seen[clean] = struct{}{}
		unique = append(unique, clean)
	}

	sort.Strings(unique)

	return unique, nil
}

func checkFile(filePath string) ([]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var issues []string

	if len(data) >= 2 {
		if (data[0] == 0xFF && data[1] == 0xFE) || (data[0] == 0xFE && data[1] == 0xFF) {
			issues = append(issues, "contains UTF-16 BOM (UTF-16 permanently banned)")
		}
	}

	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		issues = append(issues, "contains UTF-8 BOM")
	}

	if !utf8.Valid(data) {
		issues = append(issues, "contains invalid UTF-8 byte sequences")
	}

	for i, b := range data {
		if b == nullByte {
			issues = append(issues, fmt.Sprintf("contains NUL byte at offset %d (likely UTF-16)", i))

			break
		}
	}

	for i := 0; i < len(data)-1; i++ {
		if data[i] == '\r' && data[i+1] == '\n' {
			issues = append(issues, "contains CRLF line endings (CRLF permanently banned)")

			break
		}
	}

	return issues, nil
}

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
