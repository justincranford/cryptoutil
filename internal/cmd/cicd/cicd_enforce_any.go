// Copyright (c) 2025 Justin Cranford
//
//

package cicd

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	cryptoutilMagic "cryptoutil/internal/common/magic"
	cryptoutilFiles "cryptoutil/internal/common/util/files"
	"cryptoutil/internal/cmd/cicd/common"
)

// goEnforceAny enforces custom Go source code fixes across all Go files.
// It applies automated fixes like replacing interface{} with any.
// Files matching goEnforceAnyFileExcludePatterns are skipped to prevent self-modification.
// Returns an error if any files were modified (to indicate changes were made).
func goEnforceAny(logger *common.Logger, allFiles []string) error {
	logger.Log("Enforcing 'any' instead of 'interface{}' in Go files...")

	// Find all .go files
	var goFiles []string

	for _, path := range allFiles {
		if strings.HasSuffix(path, ".go") {
			// Check if file should be excluded
			excluded := false

			for _, pattern := range cryptoutilMagic.GoEnforceAnyFileExcludePatterns {
				matched, err := regexp.MatchString(pattern, path)
				if err != nil {
					logger.Log(fmt.Sprintf("Error matching pattern %s: %v", pattern, err))

					continue
				}

				if matched {
					excluded = true

					break
				}
			}

			if !excluded {
				goFiles = append(goFiles, path)
			}
		}
	}

	if len(goFiles) == 0 {
		logger.Log("goEnforceAny completed (no Go files)")

		return nil
	}

	logger.Log(fmt.Sprintf("Found %d Go files to process", len(goFiles)))

	// Process each file
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

	// Summary
	fmt.Fprintf(os.Stderr, "\n=== GOFUMPTER SUMMARY ===\n")
	fmt.Fprintf(os.Stderr, "Files processed: %d\n", len(goFiles))
	fmt.Fprintf(os.Stderr, "Files modified: %d\n", filesModified)
	fmt.Fprintf(os.Stderr, "Total replacements: %d\n", totalReplacements)

	if filesModified > 0 {
		fmt.Fprintln(os.Stderr, "\n✅ Successfully applied custom Go source code fixes")
		fmt.Fprintln(os.Stderr, "Please review and commit the changes")

		return fmt.Errorf("modified %d files with %d total replacements", filesModified, totalReplacements)
	} else {
		fmt.Fprintln(os.Stderr, "\n✅ All Go files are already properly formatted")
	}

	logger.Log("goEnforceAny completed")

	return nil
}

// processGoFile applies custom Go source code fixes to a single file.
// Currently replaces interface{} with any.
// This function is protected from self-modification by goEnforceAnyFileExcludePatterns.
// Returns the number of replacements made and any error encountered.
func processGoFile(filePath string) (int, error) {
	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read file: %w", err)
	}

	originalContent := string(content)

	// IMPORTANT: Replace interface{} with any
	// This regex matches the literal string "interface{}" in Go source code
	// The goEnforceAnyFileExcludePatterns above prevent this file from being processed
	// to avoid self-modification of the enforce-any hook implementation
	interfacePattern := `interface\{\}`
	re := regexp.MustCompile(interfacePattern)
	modifiedContent := re.ReplaceAllString(originalContent, "any")

	// Count actual replacements by counting interface{} in original content
	replacements := strings.Count(originalContent, "interface{}")

	// Only write if there were changes
	if replacements > 0 {
		err = cryptoutilFiles.WriteFile(filePath, modifiedContent, cryptoutilMagic.FilePermissionsDefault)
		if err != nil {
			return 0, fmt.Errorf("failed to write file: %w", err)
		}
	}

	return replacements, nil
}
