// Copyright (c) 2025 Justin Cranford

// Package lint_golangci provides validation for golangci-lint configuration files.
// Validates that .golangci.yml uses v2 schema and detects deprecated v1 options.
package lint_golangci

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/cmd/cicd/common"
)

const lineSeparatorLength = 80

// deprecatedV1Options maps deprecated v1 options to their v2 replacements.
var deprecatedV1Options = map[string]string{
	"wsl:":                "wsl_v5: (renamed in v2)",
	"deadcode:":           "removed in v2 (use unused linter)",
	"structcheck:":        "removed in v2 (use unused linter)",
	"varcheck:":           "removed in v2 (use unused linter)",
	"interfacer:":         "removed in v2 (use revive)",
	"maligned:":           "removed in v2 (use govet fieldalignment)",
	"scopelint:":          "removed in v2 (use exportloopref)",
	"golint:":             "removed in v2 (use revive)",
	"force-err-cuddling:": "removed in v2 (always enabled in wsl_v5)",
	"ignore-words:":       "removed from misspell in v2 (use locale)",
	"ignoreSigs:":         "removed from wrapcheck in v2 (use ignorePackageGlobs)",
	"ignoreComments:":     "removed from goconst in v2 (always enabled)",
}

// deprecatedLinters lists linters removed in golangci-lint v2.
var deprecatedLinters = []string{
	"deadcode",
	"structcheck",
	"varcheck",
	"interfacer",
	"maligned",
	"scopelint",
	"golint",
	"ifshort",
	"nosnakecase",
	"exhaustivestruct",
}

// ConfigViolation represents a configuration validation issue.
type ConfigViolation struct {
	File       string
	Line       int
	Content    string
	Reason     string
	Severity   string
	Suggestion string
}

// LintGolangCIConfig validates all golangci-lint configuration files for v2 compatibility.
// Returns an error if any config file uses deprecated v1 options.
func LintGolangCIConfig(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error {
	logger.Log("Running golangci-lint v2 config validation...")

	configFiles := FindGolangCIConfigFiles(filesByExtension)
	if len(configFiles) == 0 {
		logger.Log("No golangci-lint config files found")

		return nil
	}

	logger.Log(fmt.Sprintf("Found %d golangci-lint config files to validate", len(configFiles)))

	var violations []ConfigViolation

	for _, file := range configFiles {
		fileViolations, err := checkGolangCIConfig(file)
		if err != nil {
			logger.Log(fmt.Sprintf("Warning: failed to check %s: %v", file, err))

			continue
		}

		violations = append(violations, fileViolations...)
	}

	if len(violations) > 0 {
		printConfigViolations(violations)

		return fmt.Errorf("lint-golangci-config failed: %d v2 compatibility violations found", len(violations))
	}

	logger.Log("lint-golangci-config passed: no v2 compatibility violations")

	return nil
}

// checkGolangCIConfig validates a single golangci-lint config file.
func checkGolangCIConfig(filePath string) ([]ConfigViolation, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	defer func() { _ = file.Close() }()

	var violations []ConfigViolation

	scanner := bufio.NewScanner(file)
	lineNum := 0
	inLintersSettings := false
	inLintersList := false

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Track if we're in linters-settings section.
		if strings.HasPrefix(trimmed, "linters-settings:") {
			inLintersSettings = true
			inLintersList = false

			continue
		}

		// Track if we're in linters.enable section.
		if strings.HasPrefix(trimmed, "linters:") {
			inLintersSettings = false
			inLintersList = true

			continue
		}

		// Exit sections when we hit a new top-level key.
		if len(line) > 0 && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			if !strings.HasPrefix(trimmed, "#") && strings.Contains(trimmed, ":") {
				inLintersSettings = false
				inLintersList = false
			}
		}

		// Check for deprecated v1 options in linters-settings.
		if inLintersSettings {
			for deprecated, replacement := range deprecatedV1Options {
				if strings.Contains(trimmed, deprecated) {
					violations = append(violations, ConfigViolation{
						File:       filePath,
						Line:       lineNum,
						Content:    trimmed,
						Reason:     fmt.Sprintf("Deprecated v1 option: %s", deprecated),
						Severity:   "ERROR",
						Suggestion: replacement,
					})
				}
			}
		}

		// Check for deprecated linters in linters.enable.
		if inLintersList {
			for _, deprecated := range deprecatedLinters {
				// Match linter name in various formats: "- linter" or "enable: [linter, ...]".
				pattern := regexp.MustCompile(`(?:^-\s*` + deprecated + `\s*$|enable:\s*\[.*\b` + deprecated + `\b)`)
				if pattern.MatchString(trimmed) {
					violations = append(violations, ConfigViolation{
						File:       filePath,
						Line:       lineNum,
						Content:    trimmed,
						Reason:     fmt.Sprintf("Deprecated linter: %s", deprecated),
						Severity:   "ERROR",
						Suggestion: fmt.Sprintf("%s was removed in golangci-lint v2", deprecated),
					})
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return violations, nil
}

// printConfigViolations outputs all detected configuration violations.
func printConfigViolations(violations []ConfigViolation) {
	fmt.Println()
	fmt.Println("GOLANGCI-LINT V2 COMPATIBILITY VIOLATIONS")
	fmt.Println(strings.Repeat("=", lineSeparatorLength))

	for _, v := range violations {
		fmt.Printf("\nFile: %s\n", v.File)
		fmt.Printf("Line: %d\n", v.Line)
		fmt.Printf("Content: %s\n", v.Content)
		fmt.Printf("Reason: %s\n", v.Reason)
		fmt.Printf("Severity: %s\n", v.Severity)
		fmt.Printf("Suggestion: %s\n", v.Suggestion)
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", lineSeparatorLength))
	fmt.Println("Fix: Update your .golangci.yml to v2 schema:")
	fmt.Println("   1. Rename 'wsl:' to 'wsl_v5:'")
	fmt.Println("   2. Remove deprecated linters (deadcode, structcheck, varcheck, etc.)")
	fmt.Println("   3. Remove deprecated options (force-err-cuddling, ignore-words, etc.)")
	fmt.Println("   4. See: https://golangci-lint.run/product/roadmap/#v2")
	fmt.Println()
}

// FindGolangCIConfigFiles finds all golangci-lint config files in the project.
func FindGolangCIConfigFiles(filesByExtension map[string][]string) []string {
	var configFiles []string

	// Check for yml files.
	for _, file := range filesByExtension["yml"] {
		filename := filepath.Base(file)
		if filename == ".golangci.yml" || filename == "golangci.yml" {
			configFiles = append(configFiles, file)
		}
	}

	// Check for yaml files.
	for _, file := range filesByExtension["yaml"] {
		filename := filepath.Base(file)
		if filename == ".golangci.yaml" || filename == "golangci.yaml" {
			configFiles = append(configFiles, file)
		}
	}

	// Check for toml files.
	for _, file := range filesByExtension["toml"] {
		filename := filepath.Base(file)
		if filename == ".golangci.toml" || filename == "golangci.toml" {
			configFiles = append(configFiles, file)
		}
	}

	return configFiles
}
