// Copyright (c) 2025-2026 Justin Cranford.
// Package utf8 enforces UTF-8 encoding without BOM for text files.
package utf8

import (
	"fmt"
	"os"
	"sync"
	"unicode/utf8"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var nullByte byte

// Check enforces UTF-8 encoding without BOM for all text files.
func Check(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error {
	logger.Log("Enforcing file encoding (UTF-8 without BOM)")

	// Flatten all files from the map into a single slice.
	allFiles := flattenFileMap(filesByExtension)
	finalFiles := FilterTextFiles(allFiles)

	if len(finalFiles) == 0 {
		logger.Log("UTF-8 enforcement completed (no files)")

		return nil
	}

	logger.Log(fmt.Sprintf("Found %d text files out of %d total files to check for UTF-8 encoding", len(finalFiles), len(allFiles)))

	encodingViolations := checkFilesEncoding(finalFiles)

	if len(encodingViolations) > 0 {
		fmt.Fprintln(os.Stderr, "\n❌ Found file encoding violations:")

		for _, violation := range encodingViolations {
			fmt.Fprintf(os.Stderr, "  - %s\n", violation)
		}

		fmt.Fprintln(os.Stderr, "\nPlease fix the encoding issues above. Use UTF-8 without BOM for all text files.")
		fmt.Fprintln(os.Stderr, "PowerShell example: $utf8NoBom = New-Object System.Text.UTF8Encoding $false; [System.IO.File]::WriteAllText('file.txt', 'content', $utf8NoBom)")

		return fmt.Errorf("file encoding violations found: %d files have incorrect encoding", len(encodingViolations))
	}

	logger.LogWithPrefix("utf8", "✅ All files have correct UTF-8 encoding without BOM")

	logger.Log("UTF-8 enforcement completed")

	return nil
}

func FilterTextFiles(allFiles []string) []string {
	// Apply command-specific filtering (self-exclusion and generated files).
	// Directory-level exclusions already applied by ListAllFiles.
	return cryptoutilCmdCicdCommon.FilterFilesForCommand(allFiles, "lint-text")
}

func checkFilesEncoding(finalFiles []string) []string {
	var encodingViolations []string

	var violationsMutex sync.Mutex

	var wg sync.WaitGroup

	fileChan := make(chan string, len(finalFiles))
	resultChan := make(chan []string, len(finalFiles))

	for range cryptoutilSharedMagic.Utf8EnforceWorkerPoolSize {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for filePath := range fileChan {
				if issues := CheckFileEncoding(filePath); len(issues) > 0 {
					var violations []string
					for _, issue := range issues {
						violations = append(violations, fmt.Sprintf("%s: %s", filePath, issue))
					}

					resultChan <- violations
				} else {
					resultChan <- nil // Send nil for files with no issues.
				}
			}
		}()
	}

	go func() {
		defer close(fileChan)

		for _, filePath := range finalFiles {
			fileChan <- filePath
		}
	}()

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for violations := range resultChan {
		if violations != nil {
			violationsMutex.Lock()

			encodingViolations = append(encodingViolations, violations...)

			violationsMutex.Unlock()
		}
	}

	return encodingViolations
}

func CheckFileEncoding(filePath string) []string {
	var issues []string

	data, err := os.ReadFile(filePath)
	if err != nil {
		return []string{fmt.Sprintf("failed to open file: %v", err)}
	}

	if len(data) == 0 {
		return issues
	}

	// Check for UTF-16 BOM markers before UTF-8 validity checks.
	if len(data) >= 2 {
		if (data[0] == 0xFF && data[1] == 0xFE) || (data[0] == 0xFE && data[1] == 0xFF) {
			issues = append(issues, "file has UTF-16 BOM marker (UTF-16 is prohibited)")
		}
	}

	// Check for UTF-8 BOM (EF BB BF).
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		issues = append(issues, "file has UTF-8 BOM marker")
	}

	if !utf8.Valid(data) {
		issues = append(issues, "file is not valid UTF-8")
	}

	for i, b := range data {
		if b == nullByte {
			issues = append(issues, fmt.Sprintf("file contains NUL byte at offset %d (likely UTF-16 or binary data)", i))

			break
		}
	}

	if containsCRLF(data) {
		issues = append(issues, "file contains CRLF line endings (LF-only policy)")
	}

	return issues
}

func containsCRLF(data []byte) bool {
	for i := 0; i < len(data)-1; i++ {
		if data[i] == '\r' && data[i+1] == '\n' {
			return true
		}
	}

	return false
}

// flattenFileMap converts a map of extension -> files to a flat slice of all files.
func flattenFileMap(filesByExtension map[string][]string) []string {
	var allFiles []string

	for _, files := range filesByExtension {
		allFiles = append(allFiles, files...)
	}

	return allFiles
}
