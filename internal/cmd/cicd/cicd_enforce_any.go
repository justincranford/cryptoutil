// Package cicd provides CI/CD quality control checks for the cryptoutil project.
//
// This file contains the go-enforce-any command implementation.
package cicd

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	cryptoutilMagic "cryptoutil/internal/common/magic"
)

// goEnforceAny enforces custom Go source code fixes across all Go files.
// It applies automated fixes like replacing any with any.
// This command modifies files in place and exits with code 1 if any files were modified.
func goEnforceAny(allFiles []string) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[PERF] goEnforceAny started at %s\n", start.Format(time.RFC3339Nano))

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

		end := time.Now()
		fmt.Fprintf(os.Stderr, "[PERF] goEnforceAny: duration=%v start=%s end=%s (no Go files)\n",
			end.Sub(start), start.Format(time.RFC3339Nano), end.Format(time.RFC3339Nano))

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

	end := time.Now()
	fmt.Fprintf(os.Stderr, "[PERF] goEnforceAny: duration=%v start=%s end=%s files=%d modified=%d replacements=%d\n",
		end.Sub(start), start.Format(time.RFC3339Nano), end.Format(time.RFC3339Nano), len(goFiles), filesModified, totalReplacements)
}

// processGoFile applies custom Go source code fixes to a single file.
// Currently replaces any with any.
// Returns the number of replacements made and any error encountered.
func processGoFile(filePath string) (int, error) {
	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read file: %w", err)
	}

	originalContent := string(content)

	// IMPORTANT: DO NOT CHANGE: Replace any with any
	// Use a regex to match any as a whole word, not part of other identifiers
	// Construct the pattern to avoid self-replacement in this source file
	interfacePattern := `interface\{\}`
	re := regexp.MustCompile(interfacePattern)
	modifiedContent := re.ReplaceAllString(originalContent, "any")

	// Count actual replacements by counting any in original content
	replacements := strings.Count(originalContent, "any")

	// Only write if there were changes
	if replacements > 0 {
		err = os.WriteFile(filePath, []byte(modifiedContent), cryptoutilMagic.FilePermissionsDefault)
		if err != nil {
			return 0, fmt.Errorf("failed to write file: %w", err)
		}
	}

	return replacements, nil
}
