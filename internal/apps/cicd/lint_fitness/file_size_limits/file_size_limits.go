// Copyright (c) 2025 Justin Cranford

// Package file_size_limits enforces ARCHITECTURE.md Section 11.2.6 file size limits.
// Soft limit: 300 lines (warning), Hard limit: 500 lines (error).
// Excludes: *_gen.go, *_test.go, internal/shared/magic/ (constants only).
package file_size_limits

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	// softLimit is the line count threshold for a warning.
	softLimit = 300
	// hardLimit is the line count threshold for an error (violations block CI).
	hardLimit = 500
)

// Violation records a file that exceeds a size threshold.
type Violation struct {
	File    string
	Lines   int
	IsError bool // true = hard limit exceeded, false = soft limit warning
}

// Check enforces file size limits from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir enforces file size limits under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking file size limits...")

	projectRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return fmt.Errorf("failed to resolve root dir: %w", err)
	}

	var violations []Violation

	walkErr := filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			name := info.Name()
			if name == cryptoutilSharedMagic.CICDExcludeDirGit || name == cryptoutilSharedMagic.CICDExcludeDirVendor || name == "test-output" || name == "api" {
				return filepath.SkipDir
			}

			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		if shouldExclude(path) {
			return nil
		}

		lines, countErr := countLines(path)
		if countErr != nil {
			return countErr
		}

		if lines > hardLimit {
			violations = append(violations, Violation{File: path, Lines: lines, IsError: true})
		} else if lines > softLimit {
			violations = append(violations, Violation{File: path, Lines: lines, IsError: false})
		}

		return nil
	})
	if walkErr != nil {
		return fmt.Errorf("filesystem walk failed: %w", walkErr)
	}

	hasErrors := reportViolations(violations, projectRoot)

	if hasErrors {
		errorCount := 0

		for _, v := range violations {
			if v.IsError {
				errorCount++
			}
		}

		return fmt.Errorf("found %d files exceeding hard limit (%d lines)", errorCount, hardLimit)
	}

	logger.Log("File size limits check passed")

	return nil
}

// shouldExclude returns true if the file should be excluded from size checks.
func shouldExclude(path string) bool {
	base := filepath.Base(path)

	// Exclude generated files (both _gen.go and .gen.go patterns).
	if strings.HasSuffix(base, "_gen.go") || strings.HasSuffix(base, ".gen.go") {
		return true
	}

	// Exclude test files.
	if strings.HasSuffix(base, "_test.go") {
		return true
	}

	// Exclude magic constants package (constants only, no executable logic).
	if strings.Contains(filepath.ToSlash(path), "internal/shared/magic/") {
		return true
	}

	return false
}

// countLines counts the number of lines in a file.
func countLines(filePath string) (int, error) {
	f, err := os.Open(filePath) //nolint:gosec // filePath from filepath.Walk, controlled
	if err != nil {
		return 0, fmt.Errorf("open %s: %w", filePath, err)
	}

	defer func() { _ = f.Close() }()

	count := 0
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		count++
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("scanning file %s: %w", filePath, err)
	}

	return count, nil
}

// reportViolations prints violation information and returns true if any hard limit violations found.
func reportViolations(violations []Violation, projectRoot string) bool {
	hasErrors := false

	for _, v := range violations {
		rel, _ := filepath.Rel(projectRoot, v.File)

		if v.IsError {
			fmt.Fprintf(os.Stderr, "ERROR: %s (%d lines, exceeds hard limit of %d)\n", rel, v.Lines, hardLimit)

			hasErrors = true
		} else {
			fmt.Fprintf(os.Stderr, "WARN:  %s (%d lines, exceeds soft limit of %d)\n", rel, v.Lines, softLimit)
		}
	}

	return hasErrors
}
