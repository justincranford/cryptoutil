// Package cicd provides CI/CD quality control checks for the cryptoutil project.
//
// This file contains the go-enforce-any command implementation.
package cicd

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	cryptoutilMagic "cryptoutil/internal/common/magic"
)

// goEnforceAnyFileExcludePatterns defines files that should be excluded from the go-enforce-any command
// to prevent self-modification of the enforce-any hook implementation and related files.
var goEnforceAnyFileExcludePatterns = []string{
	`internal[/\\]cmd[/\\]cicd[/\\]cicd_enforce_any\.go$`,          // Exclude this file itself to prevent self-modification
	`internal[/\\]cmd[/\\]cicd[/\\]cicd_enforce_any_test\.go$`,     // Exclude test file to preserve deliberate test patterns
	`internal[/\\]cmd[/\\]cicd[/\\]file_patterns_enforce_any\.go$`, // Exclude pattern definitions to prevent self-modification
	`api/client`,    // Generated API client
	`api/model`,     // Generated API models
	`api/server`,    // Generated API server
	`_gen\.go$`,     // Generated files
	`\.pb\.go$`,     // Protocol buffer files
	`vendor/`,       // Vendored dependencies
	`.git/`,         // Git directory
	`node_modules/`, // Node.js dependencies
}

// goEnforceAny enforces custom Go source code fixes across all Go files.
// It applies automated fixes like replacing interface{} with any.
// Files matching goEnforceAnyFileExcludePatterns are skipped to prevent self-modification.
// This command modifies files in place and exits with code 1 if any files were modified.
func goEnforceAny(logger *LogUtil, allFiles []string) {
	fmt.Fprintln(os.Stderr, "Running go-enforce-any - Custom Go source code fixes...")

	// Find all .go files
	var goFiles []string

	for _, path := range allFiles {
		if strings.HasSuffix(path, ".go") {
			// Check if file should be excluded
			excluded := false

			for _, pattern := range goEnforceAnyFileExcludePatterns {
				matched, err := regexp.MatchString(pattern, path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error matching pattern %s: %v\n", pattern, err)

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
		fmt.Fprintln(os.Stderr, "No Go files found to process")

		logger.Log("goEnforceAny completed (no Go files)")

		return
	}

	fmt.Fprintf(os.Stderr, "Found %d Go files to process\n", len(goFiles))

	// Process each file
	filesModified := 0
	totalReplacements := 0

	for _, filePath := range goFiles {
		replacements, err := processGoFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", filePath, err)

			continue
		}

		if replacements > 0 {
			filesModified++
			totalReplacements += replacements
			fmt.Fprintf(os.Stderr, "Modified %s: %d replacements\n", filePath, replacements)
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
		os.Exit(1) // Exit with error to indicate files were modified
	} else {
		fmt.Fprintln(os.Stderr, "\n✅ All Go files are already properly formatted")
	}

	logger.Log("goEnforceAny completed")
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
		err = os.WriteFile(filePath, []byte(modifiedContent), cryptoutilMagic.FilePermissionsDefault)
		if err != nil {
			return 0, fmt.Errorf("failed to write file: %w", err)
		}
	}

	return replacements, nil
}
