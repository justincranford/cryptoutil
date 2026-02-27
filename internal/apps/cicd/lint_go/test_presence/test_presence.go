// Copyright (c) 2025 Justin Cranford

// Package test_presence validates that Go packages have corresponding test files.
package test_presence

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

// excludedDirs are directories that do not require test files.
var excludedDirs = map[string]bool{
	"magic":           true, // Constants only, no executable logic.
	"migrations":      true, // SQL files, not Go packages.
	"_testdata":       true, // Test fixture data.
	"testdata":        true, // Test fixture data.
	"testutil":        true, // Test utility packages (tested transitively).
	"testutils":       true, // Test utility packages (tested transitively).
	"fixtures":        true, // Test fixture data.
	"unified":         true, // Unified server wiring (tested via integration tests).
	"compose":         true, // Docker Compose helpers (tested via E2E).
	"e2e_helpers":     true, // E2E test helpers (tested transitively).
	"httpservertests": true, // HTTP server test helpers (tested transitively).
}

// excludedPathSegments are path components that indicate test infrastructure directories.
var excludedPathSegments = []string{
	"/testing/",  // Test helper package trees.
	"/cmd/main/", // Legacy main sub-packages (tested via binary).
}

// excludedPrefixes are directory name prefixes to skip.
var excludedPrefixes = []string{
	"_", // Archived or ignored directories.
}

// Check validates test presence from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates test presence under rootDir/internal/apps/.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	var errors []string

	appsDir := filepath.Join(rootDir, "internal", "apps")
	if _, statErr := os.Stat(appsDir); os.IsNotExist(statErr) {
		return fmt.Errorf("internal/apps directory not found at %s", appsDir)
	}

	walkErr := filepath.Walk(appsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return nil
		}

		dirName := info.Name()

		// Skip excluded directories.
		if excludedDirs[dirName] {
			return filepath.SkipDir
		}

		for _, prefix := range excludedPrefixes {
			if strings.HasPrefix(dirName, prefix) {
				return filepath.SkipDir
			}
		}

		// Skip excluded path segments.
		normalizedPath := filepath.ToSlash(path) + "/"
		for _, segment := range excludedPathSegments {
			if strings.Contains(normalizedPath, segment) {
				return filepath.SkipDir
			}
		}

		// Check if this directory has Go source files with executable logic.
		hasLogicFiles, hasTestFiles := checkDirForGoFiles(path)

		if hasLogicFiles && !hasTestFiles {
			relPath, relErr := filepath.Rel(rootDir, path)
			if relErr != nil {
				relPath = path
			}

			errors = append(errors, fmt.Sprintf("%s: package has .go source files but no _test.go files", relPath))
		}

		return nil
	})
	if walkErr != nil {
		return fmt.Errorf("failed to walk internal/apps: %w", walkErr)
	}

	if len(errors) > 0 {
		return fmt.Errorf("test presence violations:\n%s", strings.Join(errors, "\n"))
	}

	logger.Log("test-presence: all packages have test files")

	return nil
}

// checkDirForGoFiles checks if a directory contains Go source files with executable logic and test files.
// Files containing only variable declarations (e.g., embed.FS) are not counted as requiring tests.
func checkDirForGoFiles(dir string) (hasLogicFiles, hasTestFiles bool) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, false
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		if !strings.HasSuffix(name, ".go") {
			continue
		}

		// Skip generated files.
		if strings.HasSuffix(name, ".gen.go") || strings.HasSuffix(name, "_gen.go") {
			continue
		}

		if strings.HasSuffix(name, "_test.go") {
			hasTestFiles = true

			continue
		}

		// Check if file contains function declarations (executable logic).
		filePath := filepath.Join(dir, name)
		if fileHasLogic(filePath) {
			hasLogicFiles = true
		}
	}

	return hasLogicFiles, hasTestFiles
}

// fileHasLogic checks if a Go source file contains function or method declarations.
// Files with only package declarations, imports, types, constants, and variables are not considered to have logic.
func fileHasLogic(filePath string) bool {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return true // Assume logic if unable to read.
	}

	for _, line := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "func ") {
			return true
		}
	}

	return false
}
