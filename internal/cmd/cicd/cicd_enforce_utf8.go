package cicd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// File patterns for encoding checks (include patterns).
var enforceUtf8FileIncludePatterns = []string{ //nolint:unused
	// Source code files
	"*.go",    // Go source files
	"*.java",  // Java source files
	"*.sh",    // Shell scripts
	"*.py",    // Python scripts and utilities
	"*.ps1",   // PowerShell scripts
	"*.psm1",  // PowerShell module files
	"*.psd1",  // PowerShell data files
	"*.bat",   // Windows batch files
	"*.cmd",   // Windows command files
	"*.c",     // C source files
	"*.cpp",   // C++ source files
	"*.h",     // C/C++ header files
	"*.php",   // PHP files
	"*.rb",    // Ruby files
	"*.rs",    // Rust source files
	"*.js",    // JavaScript files
	"*.ts",    // TypeScript files
	"*.tsx",   // TypeScript React files
	"*.vue",   // Vue.js files
	"*.kt",    // Kotlin source files
	"*.kts",   // Kotlin script files
	"*.swift", // Swift source files
	// Database and configuration files
	"*.sql",        // SQL files
	"*.xml",        // XML configuration and data files
	"*.yml",        // YAML files
	"*.yaml",       // YAML files
	"*.json",       // JSON files
	"*.toml",       // TOML configuration files
	"*.tmpl",       // Template files
	"*.properties", // Properties files
	"*.ini",        // INI files
	"*.cfg",        // Configuration files
	"*.conf",       // Configuration files
	"*.config",     // Configuration files
	"config",       // Generic config files
	".env",         // Environment variable files
	// Build files
	"Dockerfile", // Dockerfiles
	"Makefile",   // Makefiles
	"*.mk",       // Makefiles
	"*.cmake",    // CMake files
	"*.gradle",   // Gradle build files
	// Data files
	"*.csv",    // CSV data files
	"*.pem",    // PEM files
	"*.secret", // Secret files
	// Documentation and markup files
	"*.html",     // HTML files
	"*.css",      // CSS files
	"*.md",       // Markdown files
	"*.txt",      // Text files
	"*.asciidoc", // AsciiDoc files
	"*.adoc",     // AsciiDoc files
}

// Exclusion patterns for file processing (exclude patterns).
var enforceUtf8FileExcludePatterns = []string{ //nolint:unused
	`_gen\.go$`,     // Generated files
	`\.pb\.go$`,     // Protocol buffer files
	`vendor/`,       // Vendored dependencies
	`api/client`,    // Generated API client
	`api/model`,     // Generated API models
	`api/server`,    // Generated API server
	`.git/`,         // Git directory
	`node_modules/`, // Node.js dependencies
	// NOTE: cicd_checks.go and cicd_checks_test.go are intentionally NOT excluded from all-enforce-utf8
	// as these files should validate their own UTF-8 encoding compliance
}

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

		for _, pattern := range enforceUtf8FileIncludePatterns {
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

		for _, pattern := range enforceUtf8FileExcludePatterns {
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

	fmt.Fprintf(os.Stderr, "Found %d files to check for UTF-8 encoding\n", len(finalFiles))

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
