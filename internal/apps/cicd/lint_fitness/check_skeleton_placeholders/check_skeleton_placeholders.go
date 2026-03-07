// Copyright (c) 2025 Justin Cranford

// Package check_skeleton_placeholders detects unreplaced skeleton placeholder strings
// in Go source files outside the canonical skeleton-template directories.
// When a developer copies the skeleton-template to create a new service, all
// occurrences of 'skeleton', 'Skeleton', and 'SKELETON' must be renamed.
package check_skeleton_placeholders

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Violation represents a file that contains an unreplaced skeleton placeholder string.
type Violation struct {
	File    string
	Line    int
	Word    string
	Content string
}

// skeletonWords are the placeholder words to detect (case-sensitive variants).
// These words in non-skeleton source files indicate an unreplaced template placeholder.
var skeletonWords = []string{cryptoutilSharedMagic.SkeletonProductName, cryptoutilSharedMagic.SkeletonProductNameTitleCase, cryptoutilSharedMagic.SkeletonProductNameUpperCase}

// excludedDirPrefixes are directory prefixes that are legitimately allowed to contain skeleton references.
// Paths are relative to the scan rootDir and use forward slashes.
var excludedDirPrefixes = []string{
	"internal/apps/skeleton/",
	"cmd/skeleton-template/",
	"cmd/skeleton/",
	"internal/shared/magic/",
	"internal/apps/cryptoutil/",
	"internal/apps/cicd/",
	"internal/apps/template/",
}

// excludedDirNames are directory names (single component) that skip entire subtrees.
var excludedDirNames = []string{cryptoutilSharedMagic.CICDExcludeDirVendor, cryptoutilSharedMagic.CICDExcludeDirGit, "test-output", "node_modules"}

// Test seams: replaceable in tests to exercise unreachable OS-level error paths.
// See ARCHITECTURE.md Section 10.2.4 (Test Seam Injection Pattern).
var (
	filepathAbs = filepath.Abs
	filepathRel = filepath.Rel
)

// Check scans non-skeleton Go source files in the current directory for unreplaced placeholders.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir scans non-skeleton Go source files under rootDir for unreplaced skeleton placeholders.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking for unreplaced skeleton placeholders in non-skeleton Go source files...")

	violations, err := FindViolations(rootDir)
	if err != nil {
		return fmt.Errorf("failed to scan for skeleton placeholders: %w", err)
	}

	if len(violations) > 0 {
		printViolations(violations)

		return fmt.Errorf("[ValidateSkeleton] found %d unreplaced skeleton placeholder(s) in non-skeleton source files | Fix: rename 'skeleton' to your service name | See: ARCHITECTURE.md Section 5.1", len(violations))
	}

	logger.Log("\u2705 Skeleton placeholder check passed \u2014 no unreplaced placeholders found")

	return nil
}

// FindViolations returns all unreplaced skeleton placeholder violations under rootDir.
func FindViolations(rootDir string) ([]Violation, error) {
	// Resolve rootDir to an absolute path for reliable relative-path calculation.
	absRoot, err := filepathAbs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve rootDir: %w", err)
	}

	var violations []Violation

	walkErr := filepath.WalkDir(absRoot, func(path string, d os.DirEntry, walkEntryErr error) error {
		if walkEntryErr != nil {
			return walkEntryErr
		}

		if d.IsDir() {
			for _, name := range excludedDirNames {
				if d.Name() == name {
					return filepath.SkipDir
				}
			}

			return nil
		}

		// Only check .go source files (not test files).
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Compute path relative to rootDir for consistent prefix matching.
		relPath, relErr := filepathRel(absRoot, path)
		if relErr != nil {
			return fmt.Errorf("failed to compute relative path for %s: %w", path, relErr)
		}

		relPath = filepath.ToSlash(relPath)

		// Skip files in directories that legitimately reference skeleton.
		for _, prefix := range excludedDirPrefixes {
			if strings.HasPrefix(relPath, prefix) {
				return nil
			}
		}

		fileViolations, scanErr := scanFile(path)
		if scanErr != nil {
			return fmt.Errorf("failed to read %s: %w", path, scanErr)
		}

		violations = append(violations, fileViolations...)

		return nil
	})
	if walkErr != nil {
		return violations, fmt.Errorf("failed to walk directory tree: %w", walkErr)
	}

	return violations, nil
}

// scanFile returns skeleton placeholder violations found in the given file.
func scanFile(path string) ([]Violation, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	defer func() { _ = f.Close() }()

	var violations []Violation

	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		for _, word := range skeletonWords {
			if strings.Contains(line, word) {
				violations = append(violations, Violation{
					File:    path,
					Line:    lineNum,
					Word:    word,
					Content: strings.TrimSpace(line),
				})

				break // Report each line once even if multiple words match.
			}
		}
	}

	if scanErr := scanner.Err(); scanErr != nil {
		return violations, fmt.Errorf("failed to scan file: %w", scanErr)
	}

	return violations, nil
}

// printViolations prints all skeleton placeholder violations to stderr.
func printViolations(violations []Violation) {
	fmt.Fprintf(os.Stderr, "\n\u274c Found %d unreplaced skeleton placeholder(s) in non-skeleton source files:\n", len(violations))

	for _, v := range violations {
		fmt.Fprintf(os.Stderr, "  [ValidateSkeleton] %s:%d: unreplaced placeholder %q | Fix: rename to your service name | See: ARCHITECTURE.md Section 5.1\n",
			v.File, v.Line, v.Word)
		fmt.Fprintf(os.Stderr, "    > %s\n", v.Content)
	}

	fmt.Fprintln(os.Stderr)
}
