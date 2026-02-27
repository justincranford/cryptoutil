// Copyright (c) 2025 Justin Cranford

// Package utf8 enforces UTF-8 encoding without BOM for text files.
package utf8

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

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

	fmt.Fprintln(os.Stderr, "\n✅ All files have correct UTF-8 encoding without BOM")

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

	file, err := os.Open(filePath)
	if err != nil {
		return []string{fmt.Sprintf("failed to open file: %v", err)}
	}
	defer file.Close() //nolint:errcheck // Defer close is best-effort.

	// Read first 3 bytes to check for BOM.
	header := make([]byte, 3)

	n, err := file.Read(header)
	if err != nil && !errors.Is(err, io.EOF) {
		return []string{fmt.Sprintf("failed to read file: %v", err)}
	}

	// Check for UTF-8 BOM (EF BB BF).
	// #nosec G602 -- bounds explicitly checked: n >= 3 ensures header[0], header[1], header[2] are valid.
	if n >= 3 && header[0] == 0xEF && header[1] == 0xBB && header[2] == 0xBF {
		issues = append(issues, "file has UTF-8 BOM marker")
	}

	return issues
}

// flattenFileMap converts a map of extension -> files to a flat slice of all files.
func flattenFileMap(filesByExtension map[string][]string) []string {
	var allFiles []string

	for _, files := range filesByExtension {
		allFiles = append(allFiles, files...)
	}

	return allFiles
}
