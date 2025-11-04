package cicd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	cryptoutilMagic "cryptoutil/internal/common/magic"
)

// allEnforceUtf8 enforces UTF-8 encoding without BOM for all text files.
// It filters files based on include/exclude patterns and checks each file for proper encoding.
// Any violations cause the function to print human-friendly messages and exit with a non-zero status.
func allEnforceUtf8(logger *LogUtil, allFiles []string) {
	fmt.Fprintln(os.Stderr, "Enforcing file encoding (UTF-8 without BOM)...")

	// Filter files from allFiles based on include/exclude patterns
	var finalFiles []string

	for _, filePath := range allFiles {
		// Check if matches any include pattern
		included := false

		for _, pattern := range cryptoutilMagic.EnforceUtf8FileIncludePatterns {
			if pattern == "" {
				continue
			}

			// Handle different pattern types
			if strings.HasPrefix(pattern, "*.") {
				// Extension pattern like "*.go"
				ext := strings.TrimPrefix(pattern, "*")
				if strings.HasSuffix(filePath, ext) {
					included = true

					break
				}
			} else {
				// Exact filename match like "Dockerfile"
				if filepath.Base(filePath) == pattern {
					included = true

					break
				}
			}
		}

		if !included {
			continue
		}

		// Check exclude patterns
		excluded := false

		for _, pattern := range cryptoutilMagic.EnforceUtf8FileExcludePatterns {
			matched, err := regexp.MatchString(pattern, filePath)
			if err != nil {
				continue
			}

			if matched {
				excluded = true

				break
			}
		}

		if !excluded {
			finalFiles = append(finalFiles, filePath)
		}
	}

	if len(finalFiles) == 0 {
		fmt.Fprintln(os.Stderr, "No files found to check")

		logger.Log("allEnforceUtf8 completed (no files)")

		return
	}

	logger.Log(fmt.Sprintf("Found %d files to check for UTF-8 encoding", len(finalFiles)))

	// Check each file
	var encodingViolations []string

	for _, filePath := range finalFiles {
		if issues := checkFileEncoding(filePath); len(issues) > 0 {
			for _, issue := range issues {
				encodingViolations = append(encodingViolations, fmt.Sprintf("%s: %s", filePath, issue))
			}
		}
	}

	if len(encodingViolations) > 0 {
		fmt.Fprintln(os.Stderr, "\n❌ Found file encoding violations:")

		for _, violation := range encodingViolations {
			fmt.Fprintf(os.Stderr, "  - %s\n", violation)
		}

		fmt.Fprintln(os.Stderr, "\nPlease fix the encoding issues above. Use UTF-8 without BOM for all text files.")
		fmt.Fprintln(os.Stderr, "PowerShell example: $utf8NoBom = New-Object System.Text.UTF8Encoding $false; [System.IO.File]::WriteAllText('file.txt', 'content', $utf8NoBom)")
		os.Exit(1) // Fail the build
	} else {
		fmt.Fprintln(os.Stderr, "\n✅ All files have correct UTF-8 encoding without BOM")
	}

	logger.Log("allEnforceUtf8 completed")
}

// checkFileEncoding checks a single file for proper UTF-8 encoding without BOM.
// It returns a slice of issues found, empty if the file is properly encoded.
func checkFileEncoding(filePath string) []string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return []string{fmt.Sprintf("Error reading file: %v", err)}
	}

	var issues []string

	// Check for UTF-32 LE BOM (FF FE 00 00) - check longest first
	if len(content) >= 4 && content[0] == 0xFF && content[1] == 0xFE && content[2] == 0x00 && content[3] == 0x00 {
		issues = append(issues, "contains UTF-32 LE BOM (should be UTF-8 without BOM)")

		return issues // Return immediately when BOM is found
	}

	// Check for UTF-32 BE BOM (00 00 FE FF)
	if len(content) >= 4 && content[0] == 0x00 && content[1] == 0x00 && content[2] == 0xFE && content[3] == 0xFF {
		issues = append(issues, "contains UTF-32 BE BOM (should be UTF-8 without BOM)")

		return issues // Return immediately when BOM is found
	}

	// Check for UTF-16 LE BOM (FF FE)
	if len(content) >= 2 && content[0] == 0xFF && content[1] == 0xFE {
		issues = append(issues, "contains UTF-16 LE BOM (should be UTF-8 without BOM)")

		return issues // Return immediately when BOM is found
	}

	// Check for UTF-16 BE BOM (FE FF)
	if len(content) >= 2 && content[0] == 0xFE && content[1] == 0xFF {
		issues = append(issues, "contains UTF-16 BE BOM (should be UTF-8 without BOM)")

		return issues // Return immediately when BOM is found
	}

	// Check for UTF-8 BOM (EF BB BF)
	if len(content) >= 3 && content[0] == 0xEF && content[1] == 0xBB && content[2] == 0xBF {
		issues = append(issues, "contains UTF-8 BOM (should be UTF-8 without BOM)")

		return issues // Return immediately when BOM is found
	}

	return issues
}
