// Copyright (c) 2025 Justin Cranford

// Package lint_compose provides linting for Docker Compose files.
// Validates that admin ports (9090) are NEVER exposed to the host.
package lint_compose

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

// adminPortPattern matches port mappings that expose admin port 9090 to the host.
// Examples that SHOULD match (violations):
//
//	"9090:9090"    - Direct mapping
//	"19090:9090"   - Different host port
//	"9090-9099:9090" - Port range starting with 9090
//	"9080-9089:9090" - Port range mapping to 9090
//
// Examples that should NOT match (valid):
//
//	"8080:8080"    - Public port, not admin
//	"# 9090:9090"  - Commented out
var adminPortPattern = regexp.MustCompile(`^\s*-\s*["']?\d+:9090["']?\s*(?:#.*)?$`)

// portRangeToAdmin matches port ranges that map to admin port 9090.
var portRangeToAdmin = regexp.MustCompile(`^\s*-\s*["']?\d+-\d+:9090["']?\s*(?:#.*)?$`)

// Violation represents a compose file security violation.
type Violation struct {
	File    string
	Line    int
	Content string
	Reason  string
}

// Lint checks all Docker Compose files for admin port exposure violations.
// Returns an error if any compose file exposes port 9090 to the host.
func Lint(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error {
	logger.Log("Running Docker Compose lint (admin port exposure check)...")

	// Find all compose files.
	composeFiles := findComposeFiles(filesByExtension)
	if len(composeFiles) == 0 {
		logger.Log("No compose files found")

		return nil
	}

	logger.Log(fmt.Sprintf("Found %d compose files to check", len(composeFiles)))

	var violations []Violation

	for _, file := range composeFiles {
		fileViolations, err := checkComposeFile(file)
		if err != nil {
			logger.Log(fmt.Sprintf("Warning: failed to check %s: %v", file, err))

			continue
		}

		violations = append(violations, fileViolations...)
	}

	if len(violations) > 0 {
		printViolations(violations)

		return fmt.Errorf("lint-compose failed: %d admin port exposure violations found", len(violations))
	}

	logger.Log("✅ lint-compose passed: no admin port exposure violations")

	return nil
}

// findComposeFiles returns all Docker Compose files from the file map.
func findComposeFiles(filesByExtension map[string][]string) []string {
	var composeFiles []string

	// Check yml and yaml files for compose files.
	// NOTE: filesByExtension keys are WITHOUT dots (e.g., "yml" not ".yml").
	for _, ext := range []string{"yml", "yaml"} {
		files, ok := filesByExtension[ext]
		if !ok {
			continue
		}

		for _, file := range files {
			base := filepath.Base(file)
			// Match compose*.yml, docker-compose*.yml patterns.
			if strings.HasPrefix(base, "compose") ||
				strings.HasPrefix(base, "docker-compose") {
				composeFiles = append(composeFiles, file)
			}
		}
	}

	return composeFiles
}

// checkComposeFile checks a single compose file for admin port exposure.
func checkComposeFile(filePath string) ([]Violation, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	defer func() { _ = file.Close() }()

	var violations []Violation

	scanner := bufio.NewScanner(file)
	lineNum := 0
	inPortsSection := false

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Track if we're in a ports section.
		if strings.Contains(line, "ports:") {
			inPortsSection = true

			continue
		}

		// Exit ports section when we hit a new key (not indented list item).
		if inPortsSection && !strings.HasPrefix(strings.TrimSpace(line), "-") &&
			len(strings.TrimSpace(line)) > 0 &&
			!strings.HasPrefix(strings.TrimSpace(line), "#") {
			inPortsSection = false
		}

		// Skip if not in ports section.
		if !inPortsSection {
			continue
		}

		// Skip comment-only lines.
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Check for admin port exposure.
		if adminPortPattern.MatchString(line) || portRangeToAdmin.MatchString(line) {
			violations = append(violations, Violation{
				File:    filePath,
				Line:    lineNum,
				Content: strings.TrimSpace(line),
				Reason:  "Admin port 9090 MUST NOT be exposed to host (security violation per 02-03.https-ports.instructions.md)",
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return violations, nil
}

const (
	// lineSeparatorLength defines the length of line separators in output.
	lineSeparatorLength = 60
)

// printViolations outputs all detected violations.
func printViolations(violations []Violation) {
	fmt.Println()
	fmt.Println("❌ SECURITY VIOLATIONS: Admin port 9090 exposed to host")
	fmt.Println(strings.Repeat("=", lineSeparatorLength))

	for _, v := range violations {
		fmt.Printf("\n📁 File: %s\n", v.File)
		fmt.Printf("📍 Line: %d\n", v.Line)
		fmt.Printf("📝 Content: %s\n", v.Content)
		fmt.Printf("⚠️  Reason: %s\n", v.Reason)
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", lineSeparatorLength))
	fmt.Println("💡 Fix: Remove port mapping or use internal-only networking")
	fmt.Println("   Admin APIs should only be accessible from within containers")
	fmt.Println("   (health checks use 127.0.0.1:9090 inside container)")
	fmt.Println()
}
