// Copyright (c) 2025 Justin Cranford

package all_enforce_utf8

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"cryptoutil/internal/cmd/cicd/common"
	cryptoutilMagic "cryptoutil/internal/common/magic"
)

// Enforce enforces UTF-8 encoding without BOM for all text files.
func Enforce(logger *common.Logger, allFiles []string) error {
	logger.Log("Enforcing file encoding (UTF-8 without BOM)")

	finalFiles := filterTextFiles(allFiles)

	if len(finalFiles) == 0 {
		logger.Log("UTF-8 enforcement completed (no files)")

		return nil
	}

	logger.Log(fmt.Sprintf("Found %d files to check for UTF-8 encoding", len(finalFiles)))

	encodingViolations := checkFilesEncoding(finalFiles)

	if len(encodingViolations) > 0 {
		fmt.Fprintln(os.Stderr, "\n❌ Found file encoding violations:")

		for _, violation := range encodingViolations {
			fmt.Fprintf(os.Stderr, "  - %s\n", violation)
		}

		fmt.Fprintln(os.Stderr, "\nPlease fix the encoding issues above. Use UTF-8 without BOM for all text files.")
		fmt.Fprintln(os.Stderr, "PowerShell example: $utf8NoBom = New-Object System.Text.UTF8Encoding $false; [System.IO.File]::WriteAllText('file.txt', 'content', $utf8NoBom)")

		return fmt.Errorf("file encoding violations found: %d files have incorrect encoding", len(encodingViolations))
	} else {
		fmt.Fprintln(os.Stderr, "\n✅ All files have correct UTF-8 encoding without BOM")
	}

	logger.Log("UTF-8 enforcement completed")

	return nil
}

func filterTextFiles(allFiles []string) []string {
	var finalFiles []string

	for _, filePath := range allFiles {
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

	return finalFiles
}

func checkFilesEncoding(finalFiles []string) []string {
	var encodingViolations []string

	var violationsMutex sync.Mutex

	var wg sync.WaitGroup

	fileChan := make(chan string, len(finalFiles))
	resultChan := make(chan []string, len(finalFiles))

	for range cryptoutilMagic.Utf8EnforceWorkerPoolSize {
		wg.Go(func() {
			for filePath := range fileChan {
				if issues := checkFileEncoding(filePath); len(issues) > 0 {
					var violations []string
					for _, issue := range issues {
						violations = append(violations, fmt.Sprintf("%s: %s", filePath, issue))
					}

					resultChan <- violations
				} else {
					resultChan <- nil // Send nil for files with no issues
				}
			}
		})
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

func checkFileEncoding(filePath string) []string {
	file, err := os.Open(filePath)
	if err != nil {
		return []string{fmt.Sprintf("Error reading file: %v", err)}
	}

	defer func() { _ = file.Close() }() //nolint:errcheck // Cleanup in error detection context

	// Read only the first 4 bytes (maximum needed for BOM detection)
	buffer := make([]byte, 4)

	n, err := file.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return []string{fmt.Sprintf("Error reading file: %v", err)}
	}

	var issues []string

	// Check for UTF-32 LE BOM (FF FE 00 00) - check longest first
	if n >= 4 && buffer[0] == 0xFF && buffer[1] == 0xFE && buffer[2] == 0x00 && buffer[3] == 0x00 {
		issues = append(issues, "contains UTF-32 LE BOM (should be UTF-8 without BOM)")

		return issues // Return immediately when BOM is found
	}

	// Check for UTF-32 BE BOM (00 00 FE FF)
	if n >= 4 && buffer[0] == 0x00 && buffer[1] == 0x00 && buffer[2] == 0xFE && buffer[3] == 0xFF {
		issues = append(issues, "contains UTF-32 BE BOM (should be UTF-8 without BOM)")

		return issues // Return immediately when BOM is found
	}

	// Check for UTF-16 LE BOM (FF FE)
	if n >= 2 && buffer[0] == 0xFF && buffer[1] == 0xFE {
		issues = append(issues, "contains UTF-16 LE BOM (should be UTF-8 without BOM)")

		return issues // Return immediately when BOM is found
	}

	// Check for UTF-16 BE BOM (FE FF)
	if n >= 2 && buffer[0] == 0xFE && buffer[1] == 0xFF {
		issues = append(issues, "contains UTF-16 BE BOM (should be UTF-8 without BOM)")

		return issues // Return immediately when BOM is found
	}

	// Check for UTF-8 BOM (EF BB BF)
	if n >= 3 && buffer[0] == 0xEF && buffer[1] == 0xBB && buffer[2] == 0xBF {
		issues = append(issues, "contains UTF-8 BOM (should be UTF-8 without BOM)")

		return issues // Return immediately when BOM is found
	}

	return issues
}
