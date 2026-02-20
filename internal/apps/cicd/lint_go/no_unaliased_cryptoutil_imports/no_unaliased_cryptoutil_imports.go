// Copyright (c) 2025 Justin Cranford

// Package no_unaliased_cryptoutil_imports verifies that all cryptoutil imports use aliases.
package no_unaliased_cryptoutil_imports

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

const (
	excludeDirVendor     = "vendor"
	excludeDirGit        = ".git"
	importBlockEndMarker = ")" // End of import block marker.
)

// Check validates all cryptoutil imports use aliases from .golangci.yml.
// Returns error if any unaliased cryptoutil imports are found.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	logger.Log("Checking for unaliased cryptoutil imports")

	violations, err := FindUnaliasedCryptoutilImports()
	if err != nil {
		return fmt.Errorf("failed to check cryptoutil imports: %w", err)
	}

	if len(violations) > 0 {
		PrintCryptoutilImportViolations(violations)

		return fmt.Errorf("found %d unaliased cryptoutil imports", len(violations))
	}

	logger.Log("✅ All cryptoutil imports use aliases")

	return nil
}

func FindUnaliasedCryptoutilImports() ([]string, error) {
	var violations []string

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor directories.
		if info.IsDir() && (info.Name() == excludeDirVendor || info.Name() == excludeDirGit) {
			return filepath.SkipDir
		}

		// Only check .go files.
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			fileViolations, err := CheckGoFileForUnaliasedCryptoutilImports(path)
			if err != nil {
				return fmt.Errorf("error checking %s: %w", path, err)
			}

			violations = append(violations, fileViolations...)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking directory tree: %w", err)
	}

	return violations, nil
}

func CheckGoFileForUnaliasedCryptoutilImports(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", path, err)
	}

	defer func() { _ = file.Close() }()

	var violations []string

	scanner := bufio.NewScanner(file)
	lineNum := 0
	inImportBlock := false

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Track import block.
		if trimmed == "import (" {
			inImportBlock = true

			continue
		}

		if trimmed == importBlockEndMarker && inImportBlock {
			inImportBlock = false

			continue
		}

		// Check for unaliased cryptoutil imports.
		// Pattern: starts with optional whitespace, then "cryptoutil/".
		// If it has an alias, it would be: alias "cryptoutil/..."
		// If it doesn't have an alias, it would be: "cryptoutil/..."
		if inImportBlock || strings.HasPrefix(trimmed, "import ") {
			// Extract the import line.
			importLine := trimmed

			if strings.HasPrefix(trimmed, "import ") {
				importLine = strings.TrimPrefix(trimmed, "import ")
			}

			importLine = strings.TrimSpace(importLine)

			// Check if it starts with "cryptoutil/" (unaliased).
			if strings.HasPrefix(importLine, `"cryptoutil/`) {
				violations = append(violations, fmt.Sprintf("%s:%d: unaliased cryptoutil import detected (must use importas alias)", path, lineNum))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading %s: %w", path, err)
	}

	return violations, nil
}

func PrintCryptoutilImportViolations(violations []string) {
	fmt.Fprintln(os.Stderr, "❌ Unaliased cryptoutil imports found:")
	fmt.Fprintln(os.Stderr)

	for _, v := range violations {
		fmt.Fprintf(os.Stderr, "  - %s\n", v)
	}

	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Fix:")
	fmt.Fprintln(os.Stderr, "  1. All cryptoutil imports MUST use aliases defined in .golangci.yml")
	fmt.Fprintln(os.Stderr, "  2. Run: golangci-lint run --fix")
	fmt.Fprintln(os.Stderr, "  3. If alias is missing, add it to .golangci.yml importas section")
}
