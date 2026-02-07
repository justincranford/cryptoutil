// Copyright (c) 2025 Justin Cranford

package format_go

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilFiles "cryptoutil/internal/shared/util/files"
)

// enforceAny enforces custom Go source code fixes across all Go files.
// It applies automated fixes like replacing `interface{}` with any.
//
// CRITICAL SELF-MODIFICATION PREVENTION:
// This file and its tests MUST use exclusion patterns to avoid self-modification.
// The exclusion pattern "format_go" in GetGoFiles() prevents this file from being processed.
// Test files MUST use `interface{}` in test data, NOT any, to avoid test failures.
//
// Files matching exclusion patterns are skipped to prevent self-modification.
// Returns an error if any files were modified (to indicate changes were made).
func enforceAny(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error {
	logger.Log("Enforcing 'any' instead of 'interface{}' in Go files...")

	// Get only Go files from the map.
	goFiles := filterGoFiles(filesByExtension)

	if len(goFiles) == 0 {
		logger.Log("Any enforcement completed (no Go files)")

		return nil
	}

	logger.Log(fmt.Sprintf("Found %d Go files to process", len(goFiles)))

	// Process each file.
	filesModified := 0
	totalReplacements := 0

	for _, filePath := range goFiles {
		replacements, err := processGoFile(filePath)
		if err != nil {
			logger.Log(fmt.Sprintf("Error processing %s: %v", filePath, err))

			continue
		}

		if replacements > 0 {
			filesModified++
			totalReplacements += replacements
			logger.Log(fmt.Sprintf("Modified %s: %d replacements", filePath, replacements))
		}
	}

	// Summary.
	fmt.Fprintf(os.Stderr, "\n=== GO FORMAT SUMMARY ===\n")
	fmt.Fprintf(os.Stderr, "Files processed: %d\n", len(goFiles))
	fmt.Fprintf(os.Stderr, "Files modified: %d\n", filesModified)
	fmt.Fprintf(os.Stderr, "Total replacements: %d\n", totalReplacements)

	if filesModified > 0 {
		fmt.Fprintln(os.Stderr, "\n✅ Successfully applied custom Go source code fixes")
		fmt.Fprintln(os.Stderr, "Please review and commit the changes")

		return fmt.Errorf("modified %d files with %d total replacements", filesModified, totalReplacements)
	}

	fmt.Fprintln(os.Stderr, "\n✅ All Go files are already properly formatted")

	logger.Log("Any enforcement completed")

	return nil
}

// processGoFile applies custom Go source code fixes to a single file.
// Currently replaces `interface{}` with any.
// This function is protected from self-modification by exclusion patterns.
// Returns the number of replacements made and any error encountered.
func processGoFile(filePath string) (int, error) {
	// DEFENSIVE CHECK: Never process format_go package source files
	// Only check actual source directory, not tmpDir test files
	// Use absolute path comparison to handle both Windows and Unix paths
	absPath, pathErr := filepath.Abs(filePath)
	if pathErr == nil {
		// Check if this is the actual format_go source directory
		if strings.Contains(absPath, filepath.Join("internal", "cmd", "cicd", "format_go")) &&
			!strings.Contains(absPath, filepath.Join("R:", "temp")) && // Not tmpDir
			!strings.Contains(absPath, filepath.Join("C:", "temp")) { // Not tmpDir
			return 0, nil // Skip self-modification silently
		}
	}

	// Read the file.
	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read file: %w", err)
	}

	originalContent := string(content)

	// CRITICAL: Replace `interface{}` with any
	//
	// SELF-MODIFICATION PROTECTION:
	// 1. This file is EXCLUDED by GetGoFiles("format-go") exclusion pattern
	// 2. LLM agents (Copilot, Grok) MUST NOT modify comments in this function
	// 3. Test data MUST use `interface{}` (NOT any) to verify replacement works
	// 4. Narrow-focus LLM refactoring loses exclusion context → ALWAYS preserve these comments
	//
	// This regex matches the literal string `interface{}` in Go source code.
	// The exclusion patterns prevent this file from being processed
	// to avoid self-modification of the enforce-any hook implementation.
	interfacePattern := `interface\{\}`
	re := regexp.MustCompile(interfacePattern)
	modifiedContent := re.ReplaceAllString(originalContent, "any")

	// Count actual replacements (occurrences of `interface{}` in original).
	replacements := strings.Count(originalContent, "interface{}")

	// Only write if there were changes.
	if replacements > 0 {
		err = cryptoutilSharedUtilFiles.WriteFile(filePath, modifiedContent, cryptoutilSharedMagic.FilePermissionsDefault)
		if err != nil {
			return 0, fmt.Errorf("failed to write file: %w", err)
		}
	}

	return replacements, nil
}

// filterGoFiles extracts Go files from the file map and applies exclusion patterns.
func filterGoFiles(filesByExtension map[string][]string) []string {
	// Apply command-specific filtering (self-exclusion and generated files).
	// Directory-level exclusions already applied by ListAllFiles.
	return cryptoutilCmdCicdCommon.GetGoFiles(filesByExtension, "format-go")
}
