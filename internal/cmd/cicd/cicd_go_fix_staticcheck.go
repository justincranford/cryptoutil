// Copyright (c) 2025 Justin Cranford
//
//

package cicd

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode"
)

// Error string pattern matchers.
var (
	// Pattern to match error creation with capitalized first word.
	// Matches: fmt.Errorf("Uppercase..."), errors.New("Uppercase...")
	errorStringPattern = regexp.MustCompile(`(fmt\.Errorf|errors\.New)\s*\(\s*"([A-Z][^"]*)"`)

	// Common acronyms that should remain uppercase.
	// These are excluded from lowercase conversion.
	commonAcronyms = map[string]bool{
		"HTTP":  true,
		"HTTPS": true,
		"URL":   true,
		"URI":   true,
		"API":   true,
		"JSON":   true,
		"XML":   true,
		"HTML":  true,
		"SQL":   true,
		"DB":    true,
		"ID":    true,
		"UUID":  true,
		"JWT":   true,
		"TLS":   true,
		"SSL":   true,
		"TCP":   true,
		"UDP":   true,
		"IP":    true,
		"DNS":   true,
		"CPU":   true,
		"RAM":   true,
		"OS":    true,
		"UI":    true,
		"CLI":   true,
		"RSA":   true,
		"AES":   true,
		"EC":    true,
		"ECDSA": true,
		"HMAC":  true,
		"SHA":   true,
		"PEM":   true,
		"DER":   true,
		"PKCS":  true,
		"JWK":   true,
		"JWE":   true,
		"JWS":   true,
		"JWA":   true,
		"OTLP":  true,
		"CORS":  true,
		"CSRF":  true,
		"WAL":   true,
		"ORM":   true,
		"GORM":  true,
		"AWS":   true,
		"GCP":   true,
		"CICD":  true,
		"CI":    true,
		"CD":    true,
		"PR":    true,
		"EOF":   true,
		"OK":    true,
	}
)

// goFixAll runs all auto-fix commands in sequence.
// This is a convenience command that orchestrates all go-fix-* commands.
func goFixAll(logger *LogUtil, files []string) error {
	logger.Log("Starting go-fix-all: running all auto-fix commands")

	commands := []struct {
		name string
		fn   func(*LogUtil, []string) error
	}{
		{"go-fix-staticcheck-error-strings", goFixStaticcheckErrorStrings},
		// Add more auto-fix commands here as they are implemented:
		// {"go-fix-copyloopvar", goFixCopyLoopVar},
		// {"go-fix-thelper", goFixTHelper},
	}

	var errors []error
	successCount := 0

	for _, cmd := range commands {
		logger.Log(fmt.Sprintf("Running: %s", cmd.name))

		err := cmd.fn(logger, files)
		if err != nil {
			errors = append(errors, fmt.Errorf("%s failed: %w", cmd.name, err))
		} else {
			successCount++
		}
	}

	if len(errors) > 0 {
		logger.Log(fmt.Sprintf("go-fix-all completed: %d succeeded, %d failed", successCount, len(errors)))

		// Join all errors
		var errMsgs []string
		for _, err := range errors {
			errMsgs = append(errMsgs, err.Error())
		}

		return fmt.Errorf("go-fix-all failures:\n%s", strings.Join(errMsgs, "\n"))
	}

	logger.Log(fmt.Sprintf("go-fix-all completed successfully: %d commands executed", successCount))

	return nil
}

// goFixStaticcheckErrorStrings fixes error strings that start with uppercase letters.
// According to Go style guide, error strings should not start with capital letters
// (unless beginning with proper nouns or acronyms).
func goFixStaticcheckErrorStrings(logger *LogUtil, files []string) error {
	logger.Log("Starting staticcheck error string fixes")

	goFiles := filterGoFiles(files)
	if len(goFiles) == 0 {
		logger.Log("No Go files to process")

		return nil
	}

	logger.Log(fmt.Sprintf("Processing %d Go files", len(goFiles)))

	totalFixed := 0

	for _, file := range goFiles {
		fixed, err := fixErrorStringsInFile(file)
		if err != nil {
			return fmt.Errorf("failed to fix error strings in %s: %w", file, err)
		}

		totalFixed += fixed
	}

	if totalFixed > 0 {
		logger.Log(fmt.Sprintf("Fixed %d error strings", totalFixed))

		return fmt.Errorf("fixed %d error strings - please review changes", totalFixed)
	}

	logger.Log("No error strings needed fixing")

	return nil
}

// fixErrorStringsInFile processes a single Go file and fixes error strings.
func fixErrorStringsInFile(filePath string) (int, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read file: %w", err)
	}

	originalContent := string(content)
	fixedContent := originalContent
	fixCount := 0

	// Find all matches and process them
	matches := errorStringPattern.FindAllStringSubmatch(originalContent, -1)
	if len(matches) == 0 {
		return 0, nil
	}

	// Process each match
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		fullMatch := match[0]       // e.g., `fmt.Errorf("Missing openapi spec"`
		errorMsg := match[2]        // e.g., `Missing openapi spec`
		
		// Check if the first word is an acronym
		firstWord := strings.Fields(errorMsg)[0]
		if commonAcronyms[firstWord] {
			// Skip - this is an acronym that should remain uppercase
			continue
		}

		// Lowercase the first character
		fixedMsg := lowercaseFirst(errorMsg)
		if fixedMsg == errorMsg {
			// No change needed
			continue
		}

		// Replace in the full match
		fixedMatch := strings.Replace(fullMatch, `"`+errorMsg+`"`, `"`+fixedMsg+`"`, 1)
		fixedContent = strings.Replace(fixedContent, fullMatch, fixedMatch, 1)
		fixCount++
	}

	// Only write if changes were made
	if fixCount > 0 {
		err = os.WriteFile(filePath, []byte(fixedContent), 0o644)
		if err != nil {
			return 0, fmt.Errorf("failed to write file: %w", err)
		}
	}

	return fixCount, nil
}

// lowercaseFirst converts the first character of a string to lowercase.
func lowercaseFirst(s string) string {
	if s == "" {
		return s
	}

	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])

	return string(runes)
}

// filterGoFiles filters a list of files to only include Go source files.
func filterGoFiles(files []string) []string {
	var goFiles []string

	for _, file := range files {
		// Skip non-Go files
		if !strings.HasSuffix(file, ".go") {
			continue
		}

		// Skip generated files
		if strings.HasSuffix(file, "_gen.go") || strings.HasSuffix(file, ".pb.go") {
			continue
		}

		// Normalize path separators for consistent checking
		normalizedPath := strings.ReplaceAll(file, "\\", "/")

		// Skip vendor directory
		if strings.Contains(normalizedPath, "/vendor/") || strings.HasPrefix(normalizedPath, "vendor/") {
			continue
		}

		// Skip API generated code
		if strings.Contains(normalizedPath, "/api/client/") ||
			strings.Contains(normalizedPath, "/api/model/") ||
			strings.Contains(normalizedPath, "/api/server/") ||
			strings.HasPrefix(normalizedPath, "api/client/") ||
			strings.HasPrefix(normalizedPath, "api/model/") ||
			strings.HasPrefix(normalizedPath, "api/server/") {
			continue
		}

		goFiles = append(goFiles, file)
	}

	return goFiles
}
